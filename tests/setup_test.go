package tests

import (
	"testing"

	networkmanager "github.com/veltylabs/network_manager"
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/network"
	netmem "webtyp.com/network/mem"
	"webtyp.com/orm"
	"webtyp.com/router"
	"webtyp.com/router/mock"
	"webtyp.com/storage/mem"
)

const tenantA = "tenant-A"

type mockIDGen struct{ counter int }

func (g *mockIDGen) NewID() string {
	g.counter++
	return "test-id-" + fmt.Convert(g.counter).String()
}

var _ model.IDGenerator = (*mockIDGen)(nil)

type mockPublisher struct{ Events []events.Event }

func (p *mockPublisher) Publish(e events.Event) { p.Events = append(p.Events, e) }

// fakeInventory stands in for device_manager: settable hosts, and it records
// what ImportHosts received.
type fakeInventory struct {
	hosts    []network.Host
	imported []network.Discovered
	result   network.ImportResult
}

func (f *fakeInventory) Hosts() ([]network.Host, error) { return f.hosts, nil }
func (f *fakeInventory) ImportHosts(found []network.Discovered) (network.ImportResult, error) {
	f.imported = found
	return f.result, nil
}

var (
	_ network.HostSource   = (*fakeInventory)(nil)
	_ network.HostImporter = (*fakeInventory)(nil)
)

var (
	hostPC      = network.Host{Name: "PC 1 (wifi)", MAC: "00:1A:2B:00:00:01", IP: "10.0.0.11", Access: network.AccessInternet}
	hostPrinter = network.Host{Name: "Impresora (ethernet)", MAC: "00:1A:2B:00:00:02", IP: "10.0.0.12", Access: network.AccessLocal}
)

type fixture struct {
	m   *networkmanager.Module
	gw  *netmem.Gateway
	inv *fakeInventory
	pub *mockPublisher
}

func setup(t *testing.T) fixture {
	t.Helper()
	f := fixture{gw: netmem.New(), inv: &fakeInventory{}, pub: &mockPublisher{}}
	m, err := networkmanager.New(orm.New(mem.New()), networkmanager.Deps{IDs: &mockIDGen{}, Hosts: f.inv,
		Importer: f.inv, Gateway: f.gw, Publisher: f.pub, TenantID: tenantA})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	f.m = m
	return f
}

// configured saves a valid setting for tenantA.
func (f fixture) configured(t *testing.T) {
	t.Helper()
	if _, err := f.m.SaveSetting(networkmanager.NetworkSetting{TenantId: tenantA, DhcpServer: "dhcp LAN",
		DynamicPool: "default-dhcp", FilterDns: "1.1.1.3", Unregistered: network.UnregisteredLocalName}); err != nil {
		t.Fatalf("SaveSetting: %v", err)
	}
}

// invoke runs one op through the module's registry with every permission granted.
func (f fixture) invoke(t *testing.T, op, body string) *mock.Context {
	t.Helper()
	reg := &mock.Router{}
	f.m.MountOperations(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})
	ctx := &mock.Context{InBody: []byte(body)}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+op, ctx)
	return ctx
}

func statusOf(ctx *mock.Context) int {
	if ctx.Status == 0 {
		return 200
	}
	return ctx.Status
}
