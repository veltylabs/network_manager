// Package kinds contiene los tipos de campos de opción cerrada de network_manager. Viven en
// su propio paquete porque ormc resuelve el tipo de un campo compilando y
// ejecutando su constructor, el cual solo puede importar desde su propio paquete.
package kinds

import (
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/network"
)

// UnregisteredPolicy es el radio de opciones cerradas para NetworkSetting.unregistered;
// los valores son los nombres de webtyp.com/network.
// Ejemplo: policy := kinds.UnregisteredPolicy()
func UnregisteredPolicy() input.Input {
	return input.Radio(
		fmt.KeyValue{Key: network.UnregisteredNoAddressName, Value: "No IP"},
		fmt.KeyValue{Key: network.UnregisteredLocalName, Value: "Local network only"},
	)
}
