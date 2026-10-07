package networkmanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	titleSettings    = "Red"
	titlePlan        = "Cambios pendientes"
	titleConnections = "Conectados"
	titleHistory     = "Historial de cambios"

	labelApply   = "Apply"
	confirmApply = "Apply these changes to the router?"

	registeredLabel   = "registrado"
	unregisteredLabel = "sin registrar"
	arrow             = " → "
	separator         = " · "
)

func (s *NetworkSetting) Item() view.Item {
	return view.Item{ID: s.Id, Label: s.DhcpServer, Description: s.FilterDns + separator + s.Unregistered}
}

func (p *PlanItem) Item() view.Item {
	desc := p.Reason
	if p.Category == PlanItemChange {
		desc = p.Before + arrow + p.After
	}
	label := p.Object
	if label == "" {
		label = p.Name
	}
	return view.Item{ID: p.Id, Label: label, Description: desc}
}

func (c *ConnectionRow) Item() view.Item {
	label := c.Name
	if label == "" {
		label = c.HostName
	}
	if label == "" {
		label = c.Mac
	}
	state := unregisteredLabel
	if c.Registered {
		state = registeredLabel
	}
	return view.Item{ID: c.Mac, Label: label, Description: c.Ip + separator + state}
}

func (a *NetworkApply) Item() view.Item {
	return view.Item{ID: a.Id, Label: a.Fingerprint, Description: a.TenantId}
}

// NewSettingView lists (0 or 1 rows) and saves the tenant's settings.
func NewSettingView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListNetworkSettings, Save: OpSaveNetworkSetting},
		func() model.ModelSlice { return &NetworkSettingList{} })
	return view.New(b, &NetworkSetting{}, view.WithTitle(titleSettings))
}

// NewPlanView lists the pending plan and offers the Apply action: it sends the
// fingerprint every row carries (the button is disabled when there are no rows).
func NewPlanView(caller router.Caller) view.Presenter {
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
	return view.New(b, &PlanItem{}, view.WithTitle(titlePlan))
}

// NewConnectionView lists who is connected now.
func NewConnectionView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListConnections},
		func() model.ModelSlice { return &ConnectionRowList{} })
	return view.New(b, &ConnectionRow{}, view.WithTitle(titleConnections))
}

// NewApplyHistoryView lists past applies, newest first.
func NewApplyHistoryView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListNetworkApplies},
		func() model.ModelSlice { return &NetworkApplyList{} })
	return view.New(b, &NetworkApply{}, view.WithTitle(titleHistory))
}
