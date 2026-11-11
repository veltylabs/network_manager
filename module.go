package networkmanager

import (
	"webtyp.com/events"

	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/network"
	"webtyp.com/orm"
)

// Deps son los puertos del módulo — nunca una implementación concreta.
// Ejemplo: deps := networkmanager.Deps{TenantID: "mi_tenant"}
type Deps struct {
	IDs       model.IDGenerator    // requerido
	Hosts     network.HostSource   // requerido — el inventario (device_manager en una app)
	Importer  network.HostImporter // requerido — el inventario, para import_network
	Gateway   network.Gateway      // requerido — el router (veltylabs/mikrotik en una app)
	Publisher events.Publisher     // opcional — nil deshabilita la publicación silenciosamente
	TenantID  string               // requerido — el inquilino de esta instalación
}

type Module struct {
	db       *orm.DB
	store    *orm.Store
	ids      model.IDGenerator
	hosts    network.HostSource
	importer network.HostImporter
	gateway  network.Gateway
	pub      events.Publisher
	tenantID string
}

// New conecta el módulo a un *orm.DB ya conectado; se asume que el esquema
// existe — ver el subpaquete migrate.
// Ejemplo: mod, err := networkmanager.New(db, deps)
func New(db *orm.DB, deps Deps) (*Module, error) {
	switch {
	case deps.IDs == nil:
		return nil, depErr("IDs")
	case deps.Hosts == nil:
		return nil, depErr("Hosts")
	case deps.Importer == nil:
		return nil, depErr("Importer")
	case deps.Gateway == nil:
		return nil, depErr("Gateway")
	case deps.TenantID == "":
		return nil, depErr("TenantID")
	}
	m := &Module{
		db:       db,
		store:    orm.TenantStore(db, deps.TenantID, deps.IDs),
		ids:      deps.IDs,
		hosts:    deps.Hosts,
		importer: deps.Importer,
		gateway:  deps.Gateway,
		pub:      deps.Publisher,
		tenantID: deps.TenantID,
	}
	return m, nil
}

func depErr(name string) error {
	return fmt.Err("network_manager: Deps."+name, "is required")
}

// GetSetting devuelve la configuración del inquilino, o ErrNotFound si no hay.
// Ejemplo: s, err := m.GetSetting(tenantID)
func (m *Module) GetSetting(tenantID string) (NetworkSetting, error) {
	if tenantID != m.tenantID {
		return NetworkSetting{}, ErrNotFound
	}
	var s NetworkSetting
	_, err := ReadOneNetworkSetting(m.db.Query(&s).Where(NetworkSetting_.TenantId).Eq(tenantID), &s)
	if err != nil {
		return NetworkSetting{}, err
	}
	return s, nil
}

// Desired es lo que la puerta de enlace debe imponer para el inquilino.
// Ejemplo: desired, err := m.Desired(cfg)
func (m *Module) Desired(tenantID string) (network.Desired, error) {
	s, err := m.GetSetting(tenantID)
	if err != nil {
		if model.KindOf(err) == model.ErrorNotFound {
			return network.Desired{}, ErrNotConfigured
		}
		return network.Desired{}, err
	}
	unregistered, err := network.ParseUnregistered(s.Unregistered)
	if err != nil {
		return network.Desired{}, err
	}
	hosts, err := m.hosts.Hosts()
	if err != nil {
		return network.Desired{}, err
	}
	return network.Desired{
		Settings: network.Settings{DHCPServer: s.DhcpServer, DynamicPool: s.DynamicPool,
			FilterDNS: s.FilterDns, Unregistered: unregistered},
		Hosts: hosts,
	}, nil
}

// Plan devuelve el plan de la puerta de enlace para el inquilino y quién está conectado ahora.
// Ejemplo: planItems, err := m.Plan()
func (m *Module) Plan(tenantID string) (network.Plan, []network.Connection, error) {
	d, err := m.Desired(tenantID)
	if err != nil {
		return network.Plan{}, nil, err
	}
	p, err := m.gateway.Plan(d)
	if err != nil {
		return network.Plan{}, nil, err
	}
	conns, err := m.gateway.Connections()
	if err != nil {
		return network.Plan{}, nil, err
	}
	return p, conns, nil
}

// Connections lista quién está en línea, marcado contra el inventario por MAC.
// Ejemplo: connections, err := m.Connections()
func (m *Module) Connections() ([]ConnectionRow, error) {
	conns, err := m.gateway.Connections()
	if err != nil {
		return nil, err
	}
	hosts, err := m.hosts.Hosts()
	if err != nil {
		return nil, err
	}
	rows := make([]ConnectionRow, 0, len(conns))
	for _, c := range conns {
		row := ConnectionRow{Mac: c.MAC, Ip: c.IP, HostName: c.HostName, Source: sourceName(c.Source)}
		for _, h := range hosts {
			if h.MAC == c.MAC {
				row.Registered = true
				row.Name = h.Name
				break
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// Discover lista la configuración manual del router (vista previa de importación).
// Ejemplo: rows, err := m.Discover()
func (m *Module) Discover() ([]DiscoveredRow, error) {
	found, err := m.gateway.Discover()
	if err != nil {
		return nil, err
	}
	rows := make([]DiscoveredRow, len(found))
	for i, f := range found {
		rows[i] = DiscoveredRow{Mac: f.MAC, Ip: f.IP, Name: f.Name, Internet: f.Internet}
	}
	return rows, nil
}

// ListApplies es el historial de aplicaciones, el más reciente primero.
// Ejemplo: history, err := m.ListApplies()
func (m *Module) ListApplies(tenantID string) ([]NetworkApply, error) {
	if tenantID != m.tenantID {
		return nil, nil
	}
	rows, err := ReadAllNetworkApply(m.db.Query(&NetworkApply{}).Where(NetworkApply_.TenantId).Eq(tenantID).OrderBy(NetworkApply_.AppliedAt).Desc())
	if err != nil {
		return nil, err
	}
	out := make([]NetworkApply, len(rows))
	for i, r := range rows {
		out[i] = *r
	}
	return out, nil
}

// planItems aplana un plan en filas de lista: cambios, luego conflictos, luego
// advertencias; cada fila lleva la huella digital, online marca una MAC conectada.
func planItems(p network.Plan, conns []network.Connection) []PlanItem {
	online := func(mac string) bool {
		for _, c := range conns {
			if mac != "" && c.MAC == mac {
				return true
			}
		}
		return false
	}
	var items []PlanItem
	add := func(it PlanItem) {
		it.Id = fmt.Convert(len(items) + 1).String()
		it.Fingerprint = string(p.Fingerprint)
		items = append(items, it)
	}
	for _, c := range p.Changes {
		add(PlanItem{Category: PlanItemChange, Kind: changeKindName(c.Kind), Object: c.Object, Name: c.Host.Name,
			Mac: c.Host.MAC, Before: c.Before, After: c.After, Online: online(c.Host.MAC)})
	}
	for _, c := range p.Conflicts {
		add(PlanItem{Category: PlanItemConflict, Name: c.Host.Name, Mac: c.Host.MAC, Reason: c.Reason, Online: online(c.Host.MAC)})
	}
	for _, w := range p.Warnings {
		add(PlanItem{Category: PlanItemWarning, Name: w.Name, Mac: w.MAC, Reason: w.Reason, Online: online(w.MAC)})
	}
	return items
}

func changeKindName(k network.ChangeKind) string {
	switch k {
	case network.ChangeAdd:
		return ChangeKindAdd
	case network.ChangeUpdate:
		return ChangeKindUpdate
	case network.ChangeRemove:
		return ChangeKindRemove
	case network.ChangeAdopt:
		return ChangeKindAdopt
	}
	return ""
}

func sourceName(s network.Source) string {
	switch s {
	case network.SourceDHCP:
		return SourceDHCP
	case network.SourceARP:
		return SourceARP
	}
	return ""
}
