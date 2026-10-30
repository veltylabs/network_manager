package networkmanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	titleSettings    = "Network"
	titlePlan        = "Pending changes"
	titleConnections = "Connected"
	titleHistory     = "Change history"

	labelApply   = "Apply"
	confirmApply = "Apply these changes to the router?"

)

// NewSettingView lista y guarda las configuraciones del inquilino.
// Ejemplo: view := networkmanager.NewSettingView(caller)
func NewSettingView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListNetworkSettings, Save: OpSaveNetworkSetting},
		func() model.ModelSlice { return &NetworkSettingList{} })
	allOpts := append([]view.Option{view.WithTitle(titleSettings)}, opts...)
	return view.New(b, &NetworkSetting{}, allOpts...)
}

// NewPlanView lista el plan pendiente y ofrece la acción Apply.
// Ejemplo: view := networkmanager.NewPlanView(caller)
func NewPlanView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpPlanNetwork, Actions: []view.Action{{
			Op:      OpApplyNetwork,
			Label:   labelApply,
			Confirm: confirmApply,
			Args: func(records []model.Model) model.Encodable {
				item := records[0].(*PlanItem)
				return &ApplyNetworkArgs{Fingerprint: item.Fingerprint}
			},
		}}},
		func() model.ModelSlice { return &PlanItemList{} })
	allOpts := append([]view.Option{view.WithTitle(titlePlan)}, opts...)
	return view.New(b, &PlanItem{}, allOpts...)
}

// NewConnectionView lista quién está conectado ahora.
// Ejemplo: view := networkmanager.NewConnectionView(caller)
func NewConnectionView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListConnections},
		func() model.ModelSlice { return &ConnectionRowList{} })
	allOpts := append([]view.Option{view.WithTitle(titleConnections)}, opts...)
	return view.New(b, &ConnectionRow{}, allOpts...)
}

// NewApplyHistoryView lista las aplicaciones pasadas, las más recientes primero.
// Ejemplo: view := networkmanager.NewApplyHistoryView(caller)
func NewApplyHistoryView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListNetworkApplies},
		func() model.ModelSlice { return &NetworkApplyList{} })
	allOpts := append([]view.Option{view.WithTitle(titleHistory)}, opts...)
	return view.New(b, &NetworkApply{}, allOpts...)
}
