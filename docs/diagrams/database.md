# network_manager — Database Diagram

```mermaid
flowchart TD
    S[network_setting<br/>id PK · tenant_id · dhcp_server · dynamic_pool<br/>filter_dns · unregistered · updated_at]
    A[network_apply<br/>id PK · tenant_id · fingerprint · changes · applied_at]
```

- One `network_setting` row per tenant (`SaveSetting` replaces it). `unregistered`: `no_address` / `local`.
- `network_apply` is append-only history: one row per successful apply, newest first.
- Plan items, connections and discovered rows are never stored: they are computed from the gateway.
