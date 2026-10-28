package tests

import (
	"testing"

	networkmanager "github.com/veltylabs/network_manager"
	"webtyp.com/model"
	"webtyp.com/view"
)

// recordingCaller responde a la lista del plan con dos filas y registra cada llamada.
type recordingCaller struct {
	calls []string
	args  []model.Encodable
}

func (c *recordingCaller) Call(op string, args model.Encodable, into model.Decodable, done func(error)) {
	c.calls = append(c.calls, op)
	c.args = append(c.args, args)
	if list, ok := into.(*networkmanager.PlanItemList); ok {
		*list = networkmanager.PlanItemList{
			{Id: "1", Category: networkmanager.PlanItemChange, Object: "host A", Fingerprint: "fp-123"},
			{Id: "2", Category: networkmanager.PlanItemChange, Object: "host B", Fingerprint: "fp-123"},
		}
	}
	if done != nil {
		done(nil)
	}
}

func (c *recordingCaller) Dispatch(op string, args model.Encodable) {}

// Case 12: la vista del plan declara Apply y envía la huella digital de las filas.
func TestPlanView_ApplyActionSendsFingerprint(t *testing.T) {
	cc := &recordingCaller{}
	c := view.NewRecords().Caller(cc)
	p := networkmanager.NewPlanView(c)
	a, ok := p.(view.Actioner)
	if !ok || len(a.Actions()) != 1 || a.Actions()[0].Op != networkmanager.OpApplyNetwork {
		t.Fatalf("plan view actions = %+v, want exactly apply_network", a)
	}
	var reloadErr error
	p.Reload(func(err error) { reloadErr = err })
	if reloadErr != nil {
		t.Fatalf("Reload: %v", reloadErr)
	}
	var runErr error
	a.Run(networkmanager.OpApplyNetwork, func(err error) { runErr = err })
	if runErr != nil {
		t.Fatalf("Run: %v", runErr)
	}
	want := networkmanager.ModelName + "." + networkmanager.OpApplyNetwork
	idx := -1
	for i, op := range cc.calls {
		if op == want {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("calls = %v, want one to %s", cc.calls, want)
	}
	args, ok := cc.args[idx].(*networkmanager.ApplyNetworkArgs)
	if !ok || args.Fingerprint != "fp-123" {
		t.Errorf("apply args = %#v, want fingerprint fp-123", cc.args[idx])
	}
}

func TestItemsDeclaredOnFields(t *testing.T) {
	setting := networkmanager.NetworkSetting{
		Id:           "1",
		TenantId:     "t-1",
		DhcpServer:   "LAN",
		FilterDns:    "8.8.8.8",
		Unregistered: "drop",
	}
	settingRow := setting.Item()
	if settingRow.Label != "LAN" || settingRow.Description != "8.8.8.8 · drop" {
		t.Errorf("No coincide la fila de NetworkSetting: %+v", settingRow)
	}

	planItem := networkmanager.PlanItem{
		Id:       "2",
		Category: networkmanager.PlanItemChange,
		Kind:     "update",
		Object:   "host A",
		Name:     "host A",
		Before:   "old",
		After:    "new",
		Reason:   "changed",
	}
	planRow := planItem.Item()
	if planRow.Label != "host A" || planRow.Description != "old · new · changed" {
		t.Errorf("No coincide la fila de PlanItem: %+v", planRow)
	}

	conn := networkmanager.ConnectionRow{
		Mac:        "00:11:22",
		Ip:         "10.0.0.1",
		Name:       "my-pc",
		Registered: true,
	}
	connRow := conn.Item()
	if connRow.Label != "my-pc" || connRow.Description != "10.0.0.1 · true" {
		t.Errorf("No coincide la fila de ConnectionRow: %+v", connRow)
	}

	apply := networkmanager.NetworkApply{
		Id:          "3",
		TenantId:    "t-1",
		Fingerprint: "fp-123",
	}
	applyRow := apply.Item()
	if applyRow.Label != "fp-123" || applyRow.Description != "t-1" {
		t.Errorf("No coincide la fila de NetworkApply: %+v", applyRow)
	}
}
