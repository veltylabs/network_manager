package ui

import (
	networkmanager "github.com/veltylabs/network_manager"
	"webtyp.com/layout/crudview"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/svg"
	"webtyp.com/view"
)

// Browser builds the settings screen ("Red").
func Browser(caller router.Caller, ids model.IDGenerator, tenantID string) (platformd.UIModule, error) {
	return screen(ID, Label, networkmanager.NewSettingView(caller), ids)
}

// PlanBrowser builds the pending-changes screen, with the Apply action.
func PlanBrowser(caller router.Caller, ids model.IDGenerator, tenantID string) (platformd.UIModule, error) {
	return screen(PlanID, PlanLabel, networkmanager.NewPlanView(caller), ids)
}

// ConnectionsBrowser builds the who-is-connected screen.
func ConnectionsBrowser(caller router.Caller, ids model.IDGenerator, tenantID string) (platformd.UIModule, error) {
	return screen(ConnectionsID, ConnectionsLabel, networkmanager.NewConnectionView(caller), ids)
}

// HistoryBrowser builds the apply-history screen.
func HistoryBrowser(caller router.Caller, ids model.IDGenerator, tenantID string) (platformd.UIModule, error) {
	return screen(HistoryID, HistoryLabel, networkmanager.NewApplyHistoryView(caller), ids)
}

func screen(id, label string, p view.Presenter, ids model.IDGenerator) (platformd.UIModule, error) {
	v, err := crudview.New(crudview.Config{ParentID: id, Presenter: p, IDs: ids})
	if err != nil {
		return nil, err
	}
	return platformd.NewUIModule(id, label, svg.Icon(ID), v), nil
}
