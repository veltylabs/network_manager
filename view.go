

package networkmanager

import (
	"webtyp.com/view"
)

// ModuleBrowser is what a module adds to the browser. Zero on the server.
type ModuleBrowser struct {
	Screens func(d view.BrowserDeps) []view.Screen
}
