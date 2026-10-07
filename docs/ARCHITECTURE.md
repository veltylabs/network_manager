# network_manager — architecture

## Domain scope

`network_manager` makes a site's network **follow its device inventory**. The inventory
(`device_manager`) says which devices exist, their network cards (MAC, IP) and how much of the network
each may use. This module turns that into a plan for the router, shows it to an administrator
together with who is connected right now, and applies exactly what the administrator reviewed.

You meet it when an administrator registers a PC or changes a device's access level and then opens
the "Red" screen to review and apply the change, or wants to know who is connected before touching
the router.

It knows **no router brand**: the router arrives as a `webtyp.com/network.Gateway`
(`veltylabs/mikrotik` in production, `network/mem` in tests).

## Entities

- **NetworkSetting** (one per tenant): the site-wide values that change without recompiling —
  name of the router's DHCP server, the dynamic pool, the DNS filter resolver, and what unregistered
  devices get (`no_address` or `local`). Maps 1:1 to `network.Settings`.
- **NetworkApply** (history): one row per successful apply — when, the fingerprint, and how many
  changes. Published as `network_manager.network.applied`.
- Output-only rows (never stored): **PlanItem** (one change, conflict or warning of the current plan),
  **ConnectionRow** (a device online, flagged as registered or not, with its inventory name),
  **DiscoveredRow** (hand-made router configuration, for import).

## Flow

```
device_manager ──HostSource.Hosts()──┐
NetworkSetting ──────────────────────┼─► network.Desired ─► Gateway.Plan ─► PlanItems (operator reviews)
                                     │                                          │ fingerprint
operator: apply_network(fingerprint) ┴─► Gateway.Apply(desired, fingerprint) ◄──┘
                                          ├─ ErrPlanStale → 409 (re-plan)
                                          ├─ ErrConflicts → 409
                                          └─ ok → NetworkApply row + event
```

## Decisions

- **Plan, review, apply** — never apply on save or on a timer. Rationale in
  `webtyp.com/network`'s ARCHITECTURE ("Plan, then apply").
- **Settings live in the database**, so a site changes its DNS filter or DHCP server name without a
  new build.
- **No settings ⇒ no plan.** Until the tenant saves a `NetworkSetting`, `plan_network` and
  `apply_network` answer 400 with `network settings not configured`: the module never guesses router
  object names.
- **Who is connected** comes from the gateway, cross-referenced with the inventory by MAC, so an
  unregistered device stands out before closing the network to unregistered devices.

## Ops

| Op | Action | Resource | Description |
|---|---|---|---|
| `get_network_setting` | `r` | `network_setting` | The tenant's settings (404 if none) |
| `save_network_setting` | `c`/`u` | `network_setting` | Create or replace the tenant's settings |
| `plan_network` | `r` | `network` | Current plan: changes, conflicts, warnings (+ fingerprint on every row) |
| `apply_network` | `u` | `network` | Apply the plan with the given fingerprint |
| `list_connections` | `r` | `network` | Devices online now, flagged registered / unregistered |
| `discover_network` | `r` | `network` | Hand-made router configuration (import preview) |
| `import_network` | `c` | `network` | Create inventory hosts from the discovered configuration |
| `list_network_applies` | `r` | `network` | Apply history |

## Out of scope (later phases)

Centralized access points (CAPsMAN), guest network (VLAN), QR onboarding of phones, and showing
devices on `room_layout`'s floor plan — each extends `webtyp.com/network` first.
