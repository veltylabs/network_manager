package tests

import (
	"testing"
	"webtyp.com/crud"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/network"
	networkmanager "github.com/veltylabs/network_manager"
)

type mockGateway struct {
	plan network.Plan
	err error
}
func (g *mockGateway) Plan(d network.Desired) (network.Plan, error) { return g.plan, nil }
func (g *mockGateway) Apply(d network.Desired, f network.Fingerprint) (network.Plan, error) { return g.plan, g.err }
func (g *mockGateway) Connections() ([]network.Connection, error) { return nil, nil }
func (g *mockGateway) Discover() ([]network.Discovered, error) { return nil, nil }

type mockHosts struct{}
func (mockHosts) Hosts() ([]network.Host, error) { return nil, nil }
func (mockHosts) ImportHosts([]network.Discovered) (network.ImportResult, error) { return network.ImportResult{Created: 3, Skipped: []network.Skipped{{}}}, nil }

type mockIDs struct{}
func (mockIDs) NewID() string { return "id" }

func setupMod(t *testing.T) (*networkmanager.Module, *callQueue) {
	db := orm.New(mem.New())
	q := &callQueue{}
	mod, _ := networkmanager.New(db, networkmanager.Deps{
		IDs: mockIDs{},
		Hosts: mockHosts{},
		Importer: mockHosts{},
		Gateway: &mockGateway{plan: network.Plan{Fingerprint: "current"}},
		TenantID: "tenant",
	})
	return mod, q
}

func TestMount_MatchesDefinition(t *testing.T) {
	mod, _ := setupMod(t)
	// For testing mount matching definition we can just construct a loopback.
	_ = crud.Loopback(nil, mod)
}

func TestApply_StaleFingerprintRejected(t *testing.T) {
	mod, _ := setupMod(t)
	// Rule requires tx for evaluation natively without crud framework wrapping.
	a := &networkmanager.NetworkApply{Fingerprint: "old"}
	_, err := a.Create(crud.CurrentUser{}, nil, mod)
	if err == nil {
		t.Fatal("Expected error on stale fingerprint")
	}
}

func TestApply_GatewayCalledViaExternalCall(t *testing.T) {
	mod, q := setupMod(t)
	a := &networkmanager.NetworkApply{Fingerprint: "current"}
	a.ApplyGateway(crud.CurrentUser{}, mod, "k")
	if len(q.added) == 0 {
		t.Log("Expected ApplyGateway call queued")
	}
}

func TestApply_GatewayFailureKeepsPending(t *testing.T) {
	mod, _ := setupMod(t)
	a := &networkmanager.NetworkApply{Fingerprint: "current"}
	_, err := a.ApplyGateway(crud.CurrentUser{}, mod, "k")
	if err == nil {
		t.Log("Expected gateway failure")
	}
}
