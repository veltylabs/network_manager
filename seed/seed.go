// Package seed contiene los datos de demostración de network_manager.
package seed

import (
	networkmanager "github.com/veltylabs/network_manager"
	"webtyp.com/fmt"
	"webtyp.com/network"
)

const (
	DHCPServer  = "dhcp1"
	DynamicPool = "pool1"
	FilterDNS   = "1.1.1.3"
)

// Load guarda la configuración de demostración para tenantID a través del módulo, así la fila se valida.
// Ejemplo: setting, err := seed.Load(module, "mi_tenant")
func Load(m *networkmanager.Module, tenantID string) (networkmanager.NetworkSetting, error) {
	s, err := m.SaveSetting(networkmanager.NetworkSetting{TenantId: tenantID, DhcpServer: DHCPServer,
		DynamicPool: DynamicPool, FilterDns: FilterDNS, Unregistered: network.UnregisteredLocalName})
	if err != nil {
		return networkmanager.NetworkSetting{}, fmt.Err("seed: SaveSetting", err)
	}
	return s, nil
}
