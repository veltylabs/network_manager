package ui

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
	networkmanager "github.com/veltylabs/network_manager"
)

// Screens configura las pantallas del módulo.
func Screens(caller router.Caller) []view.Screen {
	return []view.Screen{
		{
			ID: "network_manager.network_setting",
			Label: "Configuración de red",
			Icon: "network_setting",
			Permission: networkmanager.ResourceNetwork,
			Presenter: view.New(
				view.NewCallerLister(caller, view.Ops{Module: networkmanager.ModelName}, func() model.TableSlice { return &networkmanager.NetworkSettingList{} }),
				&networkmanager.NetworkSetting{},
			),
		},
		{
			ID: "network_manager.network_apply",
			Label: "Aplicar plan",
			Icon: "network_apply",
			Permission: networkmanager.ResourceNetwork,
			Presenter: view.New(
				view.NewCallerLister(caller, view.Ops{Module: networkmanager.ModelName}, func() model.TableSlice { return &networkmanager.NetworkApplyList{} }),
				&networkmanager.NetworkApply{},
			),
		},
	}
}
