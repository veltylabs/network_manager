// Package kinds holds network_manager's closed-option field kinds. They live in
// their own package because ormc resolves a field's kind by compiling and
// running its constructor, which it can only import from a package of its own.
package kinds

import (
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/network"
)

// UnregisteredPolicy is the closed-options radio for NetworkSetting.unregistered;
// the values are webtyp.com/network's names.
func UnregisteredPolicy() input.Input {
	return input.Radio(
		fmt.KeyValue{Key: network.UnregisteredNoAddressName, Value: "No IP"},
		fmt.KeyValue{Key: network.UnregisteredLocalName, Value: "Local network only"},
	)
}
