package tests

import (
	"testing"

	networkmanager "github.com/veltylabs/network_manager"
	"webtyp.com/model"
	"webtyp.com/view"
)

// recordingCaller answers the plan list with two rows and records every call.
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

// Case 12: the plan view declares Apply and sends the rows' fingerprint.
func TestPlanView_ApplyActionSendsFingerprint(t *testing.T) {
	c := &recordingCaller{}
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
	for i, op := range c.calls {
		if op == want {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("calls = %v, want one to %s", c.calls, want)
	}
	args, ok := c.args[idx].(*networkmanager.ApplyNetworkArgs)
	if !ok || args.Fingerprint != "fp-123" {
		t.Errorf("apply args = %#v, want fingerprint fp-123", c.args[idx])
	}
}
