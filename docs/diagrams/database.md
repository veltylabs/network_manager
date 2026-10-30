# network_manager — Database Diagram

```mermaid
flowchart TD
    S[network_setting<br/>id PK · tenant_id · dhcp_server · dynamic_pool<br/>filter_dns · unregistered · updated_at]
    A[network_apply<br/>id PK · tenant_id · fingerprint · changes · applied_at]
```

- Una fila `network_setting` por inquilino (`SaveSetting` la reemplaza). `unregistered`: `no_address` / `local`.
- `network_apply` es historial de solo añadir: una fila por aplicación exitosa, la más nueva primero.
- Las filas de plan, de conexiones y descubiertas nunca se almacenan: se computan desde la puerta de enlace.
