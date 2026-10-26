//go:build browser

package tests

import (
	"testing"

	"github.com/veltylabs/network_manager/ui"
	"webtyp.com/model"
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
	screens := ui.Screens(c)
	if len(screens) != 4 {
		t.Fatalf("len(screens) = %d, want 4", len(screens))
	}

	for _, s := range screens {
		if err := s.Validate(); err != nil {
			t.Errorf("screen %q Validate: %v", s.ID, err)
		}
		if s.Permission == "" {
			t.Errorf("screen %q Permission is empty", s.ID)
		}
	}

	if len(c.calls) != 0 {
		t.Errorf("calls before activation: %v", c.calls)
	}
}
