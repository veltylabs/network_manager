//go:build wasm

package main

import (
	networkmanager "github.com/veltylabs/network_manager"
	"github.com/veltylabs/network_manager/seed"
	"github.com/veltylabs/network_manager/ui"
	. "webtyp.com/dom"
	"webtyp.com/events/mock"
	"webtyp.com/layout/platformd"
	"webtyp.com/network"
	netmem "webtyp.com/network/mem"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
)

// demoTenantID is the only tenant of this in-browser demo.
const demoTenantID = "demo"

// demoUser is the fixed identity the demo shell shows: the demo has no login.
type demoUser struct{}

func (demoUser) UserName() string    { return "Demo" }
func (demoUser) UserAvatar() string  { return "" }
func (demoUser) UserRoles() []string { return []string{"Administrador"} }

// demoInventory stands in for device_manager: two registered hosts.
type demoInventory struct{}

func (demoInventory) Hosts() ([]network.Host, error) {
	return []network.Host{
		{Name: "Recepción (ethernet)", MAC: "00:1A:2B:00:01:10", IP: "192.168.1.10", Access: network.AccessInternet},
		{Name: "Impresora (wifi)", MAC: "00:1A:2B:00:01:20", IP: "192.168.1.20", Access: network.AccessLocal},
	}, nil
}

func (demoInventory) ImportHosts(found []network.Discovered) (network.ImportResult, error) {
	return network.ImportResult{Created: len(found)}, nil
}

func main() {
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err)
	}
	gw := netmem.New()
	gw.Connect(network.Connection{MAC: "00:1A:2B:00:01:10", IP: "192.168.1.10", HostName: "recepcion", Source: network.SourceDHCP})
	gw.Connect(network.Connection{MAC: "00:1A:2B:00:09:99", IP: "192.168.1.99", HostName: "celular", Source: network.SourceDHCP})
	gw.AddUnmanaged(network.Discovered{MAC: "00:1A:2B:00:09:98", IP: "192.168.1.98", Name: "PC antiguo", Internet: true})

	inv := demoInventory{}
	nm, err := networkmanager.New(orm.New(mem.New()), networkmanager.Deps{IDs: ids, Hosts: inv, Importer: inv,
		Gateway: gw, Publisher: &mock.Broker{}, TenantID: demoTenantID})
	if err != nil {
		panic(err)
	}
	if _, err := seed.Load(nm, demoTenantID); err != nil {
		panic(err)
	}

	caller := loopback.WithTenant(demoTenantID, nm)
	var modules []platformd.UIModule
	for _, build := range []func() (platformd.UIModule, error){
		func() (platformd.UIModule, error) { return ui.Browser(caller, ids, demoTenantID) },
		func() (platformd.UIModule, error) { return ui.PlanBrowser(caller, ids, demoTenantID) },
		func() (platformd.UIModule, error) { return ui.ConnectionsBrowser(caller, ids, demoTenantID) },
		func() (platformd.UIModule, error) { return ui.HistoryBrowser(caller, ids, demoTenantID) },
	} {
		m, err := build()
		if err != nil {
			panic(err)
		}
		modules = append(modules, m)
	}

	p := &platformd.Platform{
		AppName:   ui.Label + " — demo",
		User:      demoUser{},
		Modules:   modules,
		DefaultID: ui.PlanID,
	}
	Append("body", p)
	select {}
}
