package networkmanager

import (
	"webtyp.com/model"
	"webtyp.com/network"
	"webtyp.com/router"
)

const (
	OpListNetworkSettings = "list_network_settings"
	OpSaveNetworkSetting  = "save_network_setting"
	OpPlanNetwork         = "plan_network"
	OpApplyNetwork        = "apply_network"
	OpListConnections     = "list_connections"
	OpDiscoverNetwork     = "discover_network"
	OpImportNetwork       = "import_network"
	OpListNetworkApplies  = "list_network_applies"
)

func (m *Module) ModelName() string { return ModelName }

func (m *Module) MountOperations(reg router.OperationRegistry) {
	reg.Operation(OpListNetworkSettings, m.opListNetworkSettings).Requires(ResourceNetworkSetting, model.Read).Accepts(nil)
	reg.Operation(OpSaveNetworkSetting, m.opSaveNetworkSetting).Requires(ResourceNetworkSetting, model.Create|model.Update).Accepts(&NetworkSetting{})
	reg.Operation(OpPlanNetwork, m.opPlanNetwork).Requires(ResourceNetwork, model.Read).Accepts(nil)
	reg.Operation(OpApplyNetwork, m.opApplyNetwork).Requires(ResourceNetwork, model.Update).Accepts(&ApplyNetworkArgs{})
	reg.Operation(OpListConnections, m.opListConnections).Requires(ResourceNetwork, model.Read).Accepts(nil)
	reg.Operation(OpDiscoverNetwork, m.opDiscoverNetwork).Requires(ResourceNetwork, model.Read).Accepts(nil)
	reg.Operation(OpImportNetwork, m.opImportNetwork).Requires(ResourceNetwork, model.Create).Accepts(nil)
	reg.Operation(OpListNetworkApplies, m.opListNetworkApplies).Requires(ResourceNetwork, model.Read).Accepts(nil)
}

var _ router.OperationModule = (*Module)(nil)

// statusFor maps a service error to its HTTP status (AGENTS.md convention).
func statusFor(err error) int {
	if _, ok := err.(ValidationError); ok {
		return 400
	}
	switch {
	case err == ErrNotConfigured, network.IsInvalid(err):
		return 400
	case err == ErrNotFound:
		return 404
	case err == network.ErrPlanStale, err == network.ErrConflicts:
		return 409
	}
	return 500
}

func (m *Module) opListNetworkSettings(ctx router.Context) {
	list := NetworkSettingList{}
	s, err := m.GetSetting(m.tenantID)
	if err != nil && err != ErrNotFound {
		ctx.WriteStatus(500)
		return
	}
	if err == nil {
		list = append(list, &s)
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opSaveNetworkSetting(ctx router.Context) {
	var s NetworkSetting
	if err := ctx.Decode(&s); err != nil {
		ctx.WriteStatus(400)
		return
	}
	if s.TenantId == "" {
		s.TenantId = m.tenantID
	}
	saved, err := m.SaveSetting(s)
	if err != nil {
		ctx.WriteStatus(statusFor(err))
		return
	}
	if err := ctx.Encode(&saved); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opPlanNetwork(ctx router.Context) {
	p, conns, err := m.Plan(m.tenantID)
	if err != nil {
		ctx.WriteStatus(statusFor(err))
		return
	}
	items := planItems(p, conns)
	list := make(PlanItemList, len(items))
	for i := range items {
		list[i] = &items[i]
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opApplyNetwork(ctx router.Context) {
	var args ApplyNetworkArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	if err := args.Validate(model.ActionUpdate); err != nil {
		ctx.WriteStatus(400)
		return
	}
	applied, err := m.Apply(m.tenantID, args.Fingerprint)
	if err != nil {
		ctx.WriteStatus(statusFor(err))
		return
	}
	if err := ctx.Encode(&applied); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opListConnections(ctx router.Context) {
	rows, err := m.Connections()
	if err != nil {
		ctx.WriteStatus(statusFor(err))
		return
	}
	list := make(ConnectionRowList, len(rows))
	for i := range rows {
		list[i] = &rows[i]
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opDiscoverNetwork(ctx router.Context) {
	rows, err := m.Discover()
	if err != nil {
		ctx.WriteStatus(statusFor(err))
		return
	}
	list := make(DiscoveredRowList, len(rows))
	for i := range rows {
		list[i] = &rows[i]
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opImportNetwork(ctx router.Context) {
	rep, err := m.Import()
	if err != nil {
		ctx.WriteStatus(statusFor(err))
		return
	}
	if err := ctx.Encode(&rep); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opListNetworkApplies(ctx router.Context) {
	applies, err := m.ListApplies(m.tenantID)
	if err != nil {
		ctx.WriteStatus(500)
		return
	}
	list := make(NetworkApplyList, len(applies))
	for i := range applies {
		list[i] = &applies[i]
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}
