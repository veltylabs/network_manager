//go:build browser

package ui

import (
	networkmanager "github.com/veltylabs/network_manager"
	"webtyp.com/lang"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	descSettings    lang.Text = "Network configuration: DHCP server, DNS filtering, and domain controls."
	descPlan        lang.Text = "Pending network configuration changes awaiting router application."
	descConnections lang.Text = "Active client connections, IP and MAC address leases."
	descHistory     lang.Text = "Audit trail of previously applied network changes."
)

// Screens declara las pantallas que este módulo ofrece al menú de la app.
func Screens(caller router.Caller) []view.Screen {
	return []view.Screen{
		{
			ID:         ID,
			Label:      Label,
			Icon:       ID,
			Permission: networkmanager.ResourceNetworkSetting,
			Presenter:  networkmanager.NewSettingView(caller, view.WithDescription(descSettings)),
		},
		{
			ID:         PlanID,
			Label:      PlanLabel,
			Icon:       ID,
			Permission: networkmanager.ResourceNetwork,
			Presenter:  networkmanager.NewPlanView(caller, view.WithDescription(descPlan)),
		},
		{
			ID:         ConnectionsID,
			Label:      ConnectionsLabel,
			Icon:       ID,
			Permission: networkmanager.ResourceNetwork,
			Presenter:  networkmanager.NewConnectionView(caller, view.WithDescription(descConnections)),
		},
		{
			ID:         HistoryID,
			Label:      HistoryLabel,
			Icon:       ID,
			Permission: networkmanager.ResourceNetwork,
			Presenter:  networkmanager.NewApplyHistoryView(caller, view.WithDescription(descHistory)),
		},
	}
}
