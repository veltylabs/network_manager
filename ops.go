package networkmanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/rpc"
	"webtyp.com/view"
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
	reg.Operation(OpListNetworkSettings, m.opListNetworkSettings).Requires(model.ResourceOf(&NetworkSettingModel), model.Read).Accepts(nil)
	view.MountSave(reg, OpSaveNetworkSetting, func(s *NetworkSetting) error {
		if s.TenantId == "" {
			s.TenantId = m.tenantID
		}
		saved, err := m.SaveSetting(*s)
		*s = saved
		return err
	}).Requires(model.ResourceOf(&NetworkSettingModel), model.Create|model.Update)
	reg.Operation(OpPlanNetwork, m.opPlanNetwork).Requires(ResourceNetwork, model.Read).Accepts(nil)
	reg.Operation(OpApplyNetwork, m.opApplyNetwork).Requires(ResourceNetwork, model.Update).Accepts(&ApplyNetworkArgs{})
	reg.Operation(OpListConnections, m.opListConnections).Requires(ResourceNetwork, model.Read).Accepts(nil)
	reg.Operation(OpDiscoverNetwork, m.opDiscoverNetwork).Requires(ResourceNetwork, model.Read).Accepts(nil)
	reg.Operation(OpImportNetwork, m.opImportNetwork).Requires(ResourceNetwork, model.Create).Accepts(nil)
	reg.Operation(OpListNetworkApplies, m.opListNetworkApplies).Requires(ResourceNetwork, model.Read).Accepts(nil)
}

var _ router.OperationModule = (*Module)(nil)

// statusFor mapea un error de servicio a su estado HTTP (convención de AGENTS.md).

func (m *Module) opListNetworkSettings(ctx router.Context) {
	list := NetworkSettingList{}
	s, err := m.GetSetting(m.tenantID)

	isNotFound := model.KindOf(err) == model.ErrorNotFound

	if err != nil && !isNotFound {
		rpc.WriteError(ctx, err)
		return
	}
	if err == nil {
		list = append(list, &s)
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}


func (m *Module) opPlanNetwork(ctx router.Context) {
	p, conns, err := m.Plan(m.tenantID)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	items := planItems(p, conns)
	list := make(PlanItemList, len(items))
	for i := range items {
		list[i] = &items[i]
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opApplyNetwork(ctx router.Context) {
	var args ApplyNetworkArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	if err := args.Validate(model.ActionUpdate); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	applied, err := m.Apply(m.tenantID, args.Fingerprint)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	if err := ctx.Encode(&applied); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opListConnections(ctx router.Context) {
	rows, err := m.Connections()
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(ConnectionRowList, len(rows))
	for i := range rows {
		list[i] = &rows[i]
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opDiscoverNetwork(ctx router.Context) {
	rows, err := m.Discover()
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(DiscoveredRowList, len(rows))
	for i := range rows {
		list[i] = &rows[i]
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opImportNetwork(ctx router.Context) {
	rep, err := m.Import()
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	if err := ctx.Encode(&rep); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opListNetworkApplies(ctx router.Context) {
	applies, err := m.ListApplies(m.tenantID)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(NetworkApplyList, len(applies))
	for i := range applies {
		list[i] = &applies[i]
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}
