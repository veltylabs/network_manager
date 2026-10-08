package tests

import (
	"testing"

	networkmanager "github.com/veltylabs/network_manager"
	"webtyp.com/form"
	"webtyp.com/network"
	netmem "webtyp.com/network/mem"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
)

// Case 11: every required dependency is checked by New.
func TestNew_RequiredDeps(t *testing.T) {
	inv := &fakeInventory{}
	full := networkmanager.Deps{IDs: &mockIDGen{}, Hosts: inv, Importer: inv, Gateway: netmem.New(), TenantID: tenantA}
	for _, c := range []struct {
		name   string
		mutate func(d *networkmanager.Deps)
	}{
		{"IDs", func(d *networkmanager.Deps) { d.IDs = nil }},
		{"Hosts", func(d *networkmanager.Deps) { d.Hosts = nil }},
		{"Importer", func(d *networkmanager.Deps) { d.Importer = nil }},
		{"Gateway", func(d *networkmanager.Deps) { d.Gateway = nil }},
		{"TenantID", func(d *networkmanager.Deps) { d.TenantID = "" }},
	} {
		d := full
		c.mutate(&d)
		_, err := networkmanager.New(orm.New(mem.New()), d)
		want := "network_manager: Deps." + c.name + " is required"
		if err == nil || err.Error() != want {
			t.Errorf("missing %s: err = %v, want %q", c.name, err, want)
		}
	}
}

// Case 1: settings round-trip; RouterOS names with spaces and '-' validate.
func TestSettings_SaveAndReadBack(t *testing.T) {
	f := setup(t)
	f.configured(t)
	s, err := f.m.GetSetting(tenantA)
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if s.DhcpServer != "dhcp LAN" || s.DynamicPool != "default-dhcp" || s.FilterDns != "1.1.1.3" {
		t.Errorf("read back %+v", s)
	}
	// Saving again replaces the single row.
	s.FilterDns = "1.0.0.3"
	if _, err := f.m.SaveSetting(s); err != nil {
		t.Fatalf("SaveSetting (replace): %v", err)
	}
	if ctx := f.invoke(t, networkmanager.OpListNetworkSettings, `{}`); statusOf(ctx) != 200 ||
		!contains(string(ctx.ResponseBody()), "1.0.0.3") || contains(string(ctx.ResponseBody()), "1.1.1.3") {
		t.Errorf("list_network_settings = %d %s, want exactly the replaced row", statusOf(ctx), ctx.ResponseBody())
	}
}

func TestSettings_UnknownUnregisteredIs400(t *testing.T) {
	f := setup(t)
	ctx := f.invoke(t, networkmanager.OpSaveNetworkSetting,
		`{"dhcp_server":"dhcp LAN","filter_dns":"1.1.1.3","unregistered":"bogus"}`)
	if statusOf(ctx) != 400 {
		t.Errorf("save with unregistered=bogus: status %d, want 400", statusOf(ctx))
	}
}

// Case 2: no settings ⇒ no plan.
func TestPlan_NotConfiguredIs400(t *testing.T) {
	f := setup(t)
	if ctx := f.invoke(t, networkmanager.OpPlanNetwork, `{}`); statusOf(ctx) != 400 {
		t.Errorf("plan_network without settings: status %d, want 400", statusOf(ctx))
	}
}

// Case 3: the plan lists both hosts as adds, one fingerprint, online marked.
func TestPlan_RowsCarryFingerprintAndOnline(t *testing.T) {
	f := setup(t)
	f.configured(t)
	f.inv.hosts = []network.Host{hostPC, hostPrinter}
	f.gw.Connect(network.Connection{MAC: hostPC.MAC, IP: hostPC.IP, Source: network.SourceDHCP})

	ctx := f.invoke(t, networkmanager.OpPlanNetwork, `{}`)
	body := string(ctx.ResponseBody())
	if statusOf(ctx) != 200 {
		t.Fatalf("plan_network: status %d %s", statusOf(ctx), body)
	}
	p, conns, err := f.m.Plan(tenantA)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if p.Fingerprint == "" || !contains(body, string(p.Fingerprint)) {
		t.Errorf("rows do not carry the plan's fingerprint %q: %s", p.Fingerprint, body)
	}
	for _, h := range []network.Host{hostPC, hostPrinter} {
		if !contains(body, h.MAC) {
			t.Errorf("plan rows lack %s: %s", h.MAC, body)
		}
	}
	if !contains(body, `"kind":"add"`) || !contains(body, `"online":true`) {
		t.Errorf("expected add rows and one online row: %s", body)
	}
	if len(conns) != 1 {
		t.Errorf("connections = %d, want 1", len(conns))
	}
}

// Case 4: apply with the reviewed fingerprint records history and publishes.
func TestApply_RecordsHistoryAndEvent(t *testing.T) {
	f := setup(t)
	f.configured(t)
	f.inv.hosts = []network.Host{hostPC}
	p, _, err := f.m.Plan(tenantA)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	ctx := f.invoke(t, networkmanager.OpApplyNetwork, `{"fingerprint":"`+string(p.Fingerprint)+`"}`)
	if statusOf(ctx) != 200 {
		t.Fatalf("apply_network: status %d %s", statusOf(ctx), ctx.ResponseBody())
	}
	applies, err := f.m.ListApplies(tenantA)
	if err != nil || len(applies) != 1 || applies[0].Changes == 0 {
		t.Fatalf("history = %+v (err %v), want one apply with changes", applies, err)
	}
	if len(f.pub.Events) != 1 || f.pub.Events[0].Topic != networkmanager.TopicNetworkApplied {
		t.Errorf("events = %+v, want one %s", f.pub.Events, networkmanager.TopicNetworkApplied)
	}
	after, _, _ := f.m.Plan(tenantA)
	if !after.Empty() {
		t.Errorf("plan after apply has %d changes, want none", len(after.Changes))
	}
}

// Case 5: a router change between plan and apply is refused.
func TestApply_StalePlanIs409(t *testing.T) {
	f := setup(t)
	f.configured(t)
	f.inv.hosts = []network.Host{hostPC}
	p, _, _ := f.m.Plan(tenantA)
	f.gw.AddUnmanaged(network.Discovered{MAC: "00:1A:2B:00:00:09", IP: "10.0.0.19"})
	ctx := f.invoke(t, networkmanager.OpApplyNetwork, `{"fingerprint":"`+string(p.Fingerprint)+`"}`)
	if statusOf(ctx) != 409 {
		t.Errorf("stale apply: status %d, want 409", statusOf(ctx))
	}
	if applies, _ := f.m.ListApplies(tenantA); len(applies) != 0 {
		t.Errorf("a refused apply was recorded: %+v", applies)
	}
}

// Case 6: an IP held by hand-made config is a conflict that blocks apply.
func TestApply_ConflictIs409(t *testing.T) {
	f := setup(t)
	f.configured(t)
	f.inv.hosts = []network.Host{hostPC}
	f.gw.AddUnmanaged(network.Discovered{MAC: "00:1A:2B:00:00:09", IP: hostPC.IP})
	ctx := f.invoke(t, networkmanager.OpPlanNetwork, `{}`)
	if !contains(string(ctx.ResponseBody()), `"category":"conflict"`) {
		t.Fatalf("plan has no conflict row: %s", ctx.ResponseBody())
	}
	p, _, _ := f.m.Plan(tenantA)
	if ctx := f.invoke(t, networkmanager.OpApplyNetwork, `{"fingerprint":"`+string(p.Fingerprint)+`"}`); statusOf(ctx) != 409 {
		t.Errorf("apply with conflicts: status %d, want 409", statusOf(ctx))
	}
}

// Case 7: who is connected, flagged against the inventory.
func TestConnections_RegisteredFlag(t *testing.T) {
	f := setup(t)
	f.inv.hosts = []network.Host{hostPC}
	f.gw.Connect(network.Connection{MAC: hostPC.MAC, IP: hostPC.IP, Source: network.SourceDHCP})
	f.gw.Connect(network.Connection{MAC: "00:1A:2B:00:00:77", IP: "10.0.0.77", HostName: "celular", Source: network.SourceARP})
	rows, err := f.m.Connections()
	if err != nil || len(rows) != 2 {
		t.Fatalf("Connections = %+v (err %v)", rows, err)
	}
	if !rows[0].Registered || rows[0].Name != hostPC.Name || rows[0].Source != networkmanager.SourceDHCP {
		t.Errorf("registered row = %+v", rows[0])
	}
	if rows[1].Registered || rows[1].Source != networkmanager.SourceARP {
		t.Errorf("unregistered row = %+v", rows[1])
	}
}

// Case 8: discover lists hand-made config; import hands it to the inventory.
func TestDiscoverAndImport(t *testing.T) {
	f := setup(t)
	found := network.Discovered{MAC: "00:1A:2B:00:00:08", IP: "10.0.0.18", Name: "PC antiguo", Internet: true}
	f.gw.AddUnmanaged(found)
	f.inv.result = network.ImportResult{Created: 1, Skipped: []network.Skipped{{Reason: "x"}, {Reason: "y"}}}

	rows, err := f.m.Discover()
	if err != nil || len(rows) != 1 || rows[0].Mac != found.MAC || !rows[0].Internet {
		t.Fatalf("Discover = %+v (err %v)", rows, err)
	}
	rep, err := f.m.Import()
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if len(f.inv.imported) != 1 || f.inv.imported[0] != found {
		t.Errorf("importer received %+v, want [%+v]", f.inv.imported, found)
	}
	if rep.Created != 1 || rep.Skipped != 2 {
		t.Errorf("report = %+v, want created 1, skipped 2", rep)
	}
}

// Case 9: tenants never see each other's settings or history.
func TestTenantIsolation(t *testing.T) {
	f := setup(t)
	f.configured(t)
	if _, err := f.m.GetSetting("tenant-B"); err == nil || err.Error() != networkmanager.ErrNotFound.Error() {
		t.Errorf("tenant B read tenant A's setting: err = %v", err)
	}
	f.inv.hosts = []network.Host{hostPC}
	p, _, _ := f.m.Plan(tenantA)
	if _, err := f.m.Apply(tenantA, string(p.Fingerprint)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if applies, _ := f.m.ListApplies("tenant-B"); len(applies) != 0 {
		t.Errorf("tenant B sees tenant A's history: %+v", applies)
	}
}

// Case 10: the settings form shows exactly the editable fields.
func TestSettingsFormWidgets(t *testing.T) {
	fm, err := form.New("nm", &networkmanager.NetworkSetting{}, &mockIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}
	for _, name := range []string{"dhcp_server", "dynamic_pool", "filter_dns", "unregistered"} {
		if fm.Input(name) == nil {
			t.Errorf("form lacks %q", name)
		}
	}
	for _, name := range []string{"id", "tenant_id", "updated_at"} {
		if fm.Input(name) != nil {
			t.Errorf("form has %q, which must not be editable", name)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
