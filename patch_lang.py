import json

with open('lang.json', 'r') as f:
    data = json.load(f)

# Update with the exact requested translations if there's any mismatch.
new_keys = {
    "Network configuration: DHCP server, DNS filtering, and domain controls.": "Configuración de red: servidor DHCP, filtrado DNS y controles de dominio.",
    "Pending network configuration changes awaiting router application.": "Cambios pendientes de configuración de red en espera de aplicación en el router.",
    "Active client connections, IP and MAC address leases.": "Conexiones activas de clientes, concesiones de direcciones IP y MAC.",
    "Audit trail of previously applied network changes.": "Historial de auditoría de cambios de red aplicados anteriormente."
}

for k, v in new_keys.items():
    if k in data["keys"]:
        data["keys"][k][0] = v

with open('lang.json', 'w') as f:
    json.dump(data, f, indent=2)
