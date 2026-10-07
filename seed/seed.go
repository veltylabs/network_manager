// Package seed holds the demo data of network_manager.
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

// Load saves the demo settings for tenantID through the module, so the row is validated.
func Load(m *networkmanager.Module, tenantID string) (networkmanager.NetworkSetting, error) {
	s, err := m.SaveSetting(networkmanager.NetworkSetting{TenantId: tenantID, DhcpServer: DHCPServer,
		DynamicPool: DynamicPool, FilterDns: FilterDNS, Unregistered: network.UnregisteredLocalName})
	if err != nil {
		return networkmanager.NetworkSetting{}, fmt.Err("seed: SaveSetting", err)
	}
	return s, nil
}
