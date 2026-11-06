package networkmanager

import (
	"sort"
	"strings"

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
	reg.Operation(OpListNetworkSettings, m.opListNetworkSettings).Requires(model.ResourceOf(&NetworkSettingModel), model.Read).Accepts(&view.ListArgs{})
	view.MountSave(reg, OpSaveNetworkSetting, func(s *NetworkSetting) error {
		if s.TenantId == "" {
			s.TenantId = m.tenantID
		}
		saved, err := m.SaveSetting(*s)
		*s = saved
		return err
	}).Requires(model.ResourceOf(&NetworkSettingModel), model.Create|model.Update)
	reg.Operation(OpPlanNetwork, m.opPlanNetwork).Requires(ResourceNetwork, model.Read).Accepts(&view.ListArgs{})
	reg.Operation(OpApplyNetwork, m.opApplyNetwork).Requires(ResourceNetwork, model.Update).Accepts(&ApplyNetworkArgs{})
	reg.Operation(OpListConnections, m.opListConnections).Requires(ResourceNetwork, model.Read).Accepts(&view.ListArgs{})
	reg.Operation(OpDiscoverNetwork, m.opDiscoverNetwork).Requires(ResourceNetwork, model.Read).Accepts(nil)
	reg.Operation(OpImportNetwork, m.opImportNetwork).Requires(ResourceNetwork, model.Create).Accepts(nil)
	reg.Operation(OpListNetworkApplies, m.opListNetworkApplies).Requires(ResourceNetwork, model.Read).Accepts(&view.ListArgs{})
}

var _ router.OperationModule = (*Module)(nil)

// statusFor mapea un error de servicio a su estado HTTP (convención de AGENTS.md).

func (m *Module) opListNetworkSettings(ctx router.Context) {
	var args view.ListArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	s, err := m.GetSetting(m.tenantID)

	isNotFound := model.KindOf(err) == model.ErrorNotFound

	if err != nil && !isNotFound {
		rpc.WriteError(ctx, err)
		return
	}
	list := NetworkSettingList{}
	if err == nil {
		if args.After == "" || s.Id > args.After {
			if args.Search == "" || strings.Contains(strings.ToLower(s.DhcpServer), strings.ToLower(args.Search)) {
				list = append(list, &s)
			}
		}
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}


func (m *Module) opPlanNetwork(ctx router.Context) {
	var args view.ListArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	p, conns, err := m.Plan(m.tenantID)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	items := planItems(p, conns)
	limit := int(view.Cap(args.Limit))
	list := make(PlanItemList, 0, len(items))
	for i := range items {
		it := &items[i]
		if args.After != "" && it.Id <= args.After {
			continue
		}
		if args.Search != "" && !strings.Contains(strings.ToLower(it.Name), strings.ToLower(args.Search)) {
			continue
		}
		list = append(list, it)
		if len(list) >= limit {
			break
		}
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
	var args view.ListArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	rows, err := m.Connections()
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Mac < rows[j].Mac
	})
	limit := int(view.Cap(args.Limit))
	list := make(ConnectionRowList, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		if args.After != "" && r.Mac <= args.After {
			continue
		}
		if args.Search != "" && !strings.Contains(strings.ToLower(r.Name), strings.ToLower(args.Search)) &&
			!strings.Contains(strings.ToLower(r.HostName), strings.ToLower(args.Search)) &&
			!strings.Contains(strings.ToLower(r.Ip), strings.ToLower(args.Search)) &&
			!strings.Contains(strings.ToLower(r.Mac), strings.ToLower(args.Search)) {
			continue
		}
		list = append(list, r)
		if len(list) >= limit {
			break
		}
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
	var args view.ListArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	rows, err := m.listApplies(args.Search, args.After, view.Cap(args.Limit))
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(NetworkApplyList, len(rows))
	for i := range rows {
		list[i] = &rows[i]
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}
