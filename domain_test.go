// Root test: statusFor is unexported; tests/ can only reach the public API.
package networkmanager

import (
	"testing"
	"webtyp.com/network"
)

func TestDomainErrors(t *testing.T) {
	// Test that checks the Error() string of ErrNotConfigured and ErrNotFound
	if ErrNotConfigured.Error() != "network settings not configured" {
		t.Errorf("ErrNotConfigured string mismatch: %q", ErrNotConfigured.Error())
	}
	if ErrNotFound.Error() != "network setting not found" {
		t.Errorf("ErrNotFound string mismatch: %q", ErrNotFound.Error())
	}

	// Test statusFor
	if status := statusFor(ErrNotConfigured); status != 400 {
		t.Errorf("statusFor(ErrNotConfigured) = %d, want 400", status)
	}
	if status := statusFor(ErrNotFound); status != 404 {
		t.Errorf("statusFor(ErrNotFound) = %d, want 404", status)
	}
	if status := statusFor(network.ErrPlanStale); status != 409 {
		t.Errorf("statusFor(ErrPlanStale) = %d, want 409", status)
	}
	if status := statusFor(network.ErrConflicts); status != 409 {
		t.Errorf("statusFor(ErrConflicts) = %d, want 409", status)
	}
	if status := statusFor(ValidationError{}); status != 400 {
		t.Errorf("statusFor(ValidationError) = %d, want 400", status)
	}
	if status := statusFor(nil); status != 200 {
		t.Errorf("statusFor(nil) = %d, want 200", status)
	}
}
