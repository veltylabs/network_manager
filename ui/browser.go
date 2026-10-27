//go:build browser

package ui

import (
	networkmanager "github.com/veltylabs/network_manager"
	"webtyp.com/lang"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	descSettings    lang.Text = "Configuración de red: servidor DHCP, filtrado DNS y controles de dominio."
	descPlan        lang.Text = "Cambios pendientes de configuración de red en espera de aplicación en el router."
	descConnections lang.Text = "Conexiones activas de clientes, concesiones de direcciones IP y MAC."
	descHistory     lang.Text = "Historial de auditoría de cambios de red aplicados anteriormente."
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
