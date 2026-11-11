//go:build wasm

package seed

import (
	"webtyp.com/crud"
	"webtyp.com/view"
	"github.com/veltylabs/network_manager/ui"
)

type fakes struct {
	// mock here
}

func (f *fakes) IDs() string { return "id" }

// Demo retorna la configuración del appkit Demo
func Demo() view.Demo {
	return view.Demo{
		Label:    "Demo Red",
		Timezone: "America/Santiago",
		ActingAs: "u1",
		Modules:  []crud.Module{},
		Screens: func(d view.BrowserDeps) []view.Screen {
			return ui.Screens(d.Caller)
		},
	}
}
