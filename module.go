package networkmanager

import (
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"
	"webtyp.com/network"
	"webtyp.com/orm"
	"webtyp.com/time"
)

// Deps are the module's ports — never a concrete implementation.
type Deps struct {
	IDs       model.IDGenerator    // required
	Hosts     network.HostSource   // required — the inventory (device_manager in an app)
	Importer  network.HostImporter // required — the inventory, for import_network
	Gateway   network.Gateway      // required — the router (veltylabs/mikrotik in an app)
	Publisher events.Publisher     // optional — nil disables publishing silently
	TenantID  string               // required — this installation's tenant
}

type Module struct {
	db       *orm.DB
	ids      model.IDGenerator
	hosts    network.HostSource
	importer network.HostImporter
	gateway  network.Gateway
	pub      events.Publisher
	tenantID string
}

// New connects the module to an already-connected *orm.DB; the schema is
// assumed to exist — see the migrate subpackage.
func New(db *orm.DB, deps Deps) (*Module, error) {
	switch {
	case deps.IDs == nil:
		return nil, fmt.Err("network_manager: Deps.IDs is required")
	case deps.Hosts == nil:
		return nil, fmt.Err("network_manager: Deps.Hosts is required")
	case deps.Importer == nil:
		return nil, fmt.Err("network_manager: Deps.Importer is required")
	case deps.Gateway == nil:
		return nil, fmt.Err("network_manager: Deps.Gateway is required")
	case deps.TenantID == "":
		return nil, fmt.Err("network_manager: Deps.TenantID is required")
	}
	return &Module{db: db, ids: deps.IDs, hosts: deps.Hosts, importer: deps.Importer,
		gateway: deps.Gateway, pub: deps.Publisher, tenantID: deps.TenantID}, nil
}

// GetSetting returns the tenant's settings; ErrNotFound when none were saved.
func (m *Module) GetSetting(tenantID string) (NetworkSetting, error) {
	var s NetworkSetting
	if _, err := ReadOneNetworkSetting(m.db.Query(&s).Where(NetworkSetting_.TenantId).Eq(tenantID), &s); err != nil {
		if err == orm.ErrNotFound {
			return NetworkSetting{}, ErrNotFound
		}
		return NetworkSetting{}, err
	}
	return s, nil
}

// SaveSetting creates or replaces the tenant's single settings row.
func (m *Module) SaveSetting(s NetworkSetting) (NetworkSetting, error) {
	if _, err := network.ParseUnregistered(s.Unregistered); err != nil {
		return NetworkSetting{}, ValidationError{Err: err}
	}
	s.FilterDns = input.CanonicalIP(s.FilterDns)
	existing, err := m.GetSetting(s.TenantId)
	creating := err == ErrNotFound
	if err != nil && !creating {
		return NetworkSetting{}, err
	}
	action := model.ActionUpdate
	if creating {
		s.Id = m.ids.NewID()
		action = model.ActionCreate
	} else {
		s.Id = existing.Id
	}
	s.UpdatedAt = time.Now()
	if err := s.Validate(action); err != nil {
		return NetworkSetting{}, ValidationError{Err: err}
	}
	if creating {
		err = m.db.Create(&s)
	} else {
		err = m.db.Update(&s, orm.Eq(NetworkSetting_.Id, s.Id), orm.Eq(NetworkSetting_.TenantId, s.TenantId))
	}
	if err != nil {
		return NetworkSetting{}, err
	}
	return s, nil
}

// Desired is what the gateway must enforce for the tenant.
func (m *Module) Desired(tenantID string) (network.Desired, error) {
	s, err := m.GetSetting(tenantID)
	if err != nil {
		if err == ErrNotFound {
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

// Plan returns the gateway's plan for the tenant and who is connected now.
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

// Apply applies exactly the plan whose fingerprint the operator reviewed and
// records it in the history.
func (m *Module) Apply(tenantID, fingerprint string) (NetworkApply, error) {
	d, err := m.Desired(tenantID)
	if err != nil {
		return NetworkApply{}, err
	}
	p, err := m.gateway.Apply(d, network.Fingerprint(fingerprint))
	if err != nil {
		return NetworkApply{}, err
	}
	a := NetworkApply{Id: m.ids.NewID(), TenantId: tenantID, Fingerprint: fingerprint,
		Changes: int64(len(p.Changes)), AppliedAt: time.Now()}
	if err := m.db.Create(&a); err != nil {
		return NetworkApply{}, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicNetworkApplied, Payload: &a})
	}
	return a, nil
}

// Connections lists who is online, flagged against the inventory by MAC.
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

// Discover lists the router's hand-made configuration (import preview).
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

// Import hands the discovered configuration to the inventory.
func (m *Module) Import() (ImportReport, error) {
	found, err := m.gateway.Discover()
	if err != nil {
		return ImportReport{}, err
	}
	res, err := m.importer.ImportHosts(found)
	if err != nil {
		return ImportReport{}, err
	}
	return ImportReport{Created: int64(res.Created), Skipped: int64(len(res.Skipped))}, nil
}

// ListApplies is the apply history, newest first.
func (m *Module) ListApplies(tenantID string) ([]NetworkApply, error) {
	var a NetworkApply
	rows, err := ReadAllNetworkApply(m.db.Query(&a).Where(NetworkApply_.TenantId).Eq(tenantID).OrderBy(NetworkApply_.AppliedAt).Desc())
	if err != nil {
		return nil, err
	}
	out := make([]NetworkApply, len(rows))
	for i, r := range rows {
		out[i] = *r
	}
	return out, nil
}

// planItems flattens a plan into list rows: changes, then conflicts, then
// warnings; every row carries the fingerprint, online marks a connected MAC.
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
