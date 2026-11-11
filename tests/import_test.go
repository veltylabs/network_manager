package tests

import (
	"testing"
	"webtyp.com/crud"
	networkmanager "github.com/veltylabs/network_manager"
)

func TestImport_CountsStored(t *testing.T) {
	mod, _ := setupMod(t)
	row := &networkmanager.NetworkImport{}
	mod.StoreNetworkImport().Create( row)
	row.RunImport(crud.CurrentUser{}, mod, "k")
	if row.Imported != 3 || row.Skipped != 1 {
		t.Errorf("got %d imported %d skipped, want 3 and 1", row.Imported, row.Skipped)
	}
}
