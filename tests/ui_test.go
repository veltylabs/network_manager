package tests

import (
	"testing"

	"github.com/veltylabs/network_manager/ui"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/router"
)

type uiCaller struct{ calls []string }

func (c *uiCaller) Call(op string, args model.Encodable, into model.Decodable, done func(error)) {
	c.calls = append(c.calls, op)
	if done != nil {
		done(nil)
	}
}

func (c *uiCaller) Dispatch(op string, args model.Encodable) {}

// Every screen builds — including the read-only ones (plan, connections,
// history), whose records have no form widgets — and none calls the server
// before it is activated.
func TestUI_EveryScreenBuilds(t *testing.T) {
	c := &uiCaller{}
	for _, build := range []struct {
		id string
		fn func(router.Caller, model.IDGenerator, string) (platformd.UIModule, error)
	}{
		{ui.ID, ui.Browser},
		{ui.PlanID, ui.PlanBrowser},
		{ui.ConnectionsID, ui.ConnectionsBrowser},
		{ui.HistoryID, ui.HistoryBrowser},
	} {
		m, err := build.fn(c, &mockIDGen{}, tenantA)
		if err != nil {
			t.Fatalf("%s: %v", build.id, err)
		}
		if m.ModelName() != build.id {
			t.Errorf("ModelName = %q, want %q", m.ModelName(), build.id)
		}
	}
	if len(c.calls) != 0 {
		t.Errorf("calls before activation: %v", c.calls)
	}
}
