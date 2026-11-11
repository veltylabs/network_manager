package tests

import (
	"testing"
	"webtyp.com/crud"
	networkmanager "github.com/veltylabs/network_manager"
)

func TestReadModel_Plan(t *testing.T) {
	mod, _ := setupMod(t)
	var p networkmanager.PlanItem
	_, err := p.Read(crud.CurrentUser{UserID: "u1"}, nil, crud.ReadArgs{}, mod)
	if err != nil {
		t.Fatalf("Plan Read failed: %v", err)
	}
}
