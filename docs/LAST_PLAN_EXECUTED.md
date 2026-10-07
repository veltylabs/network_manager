---
PLAN: "feat: network_manager — review and apply the device inventory to the network gateway"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 5897233414007165870
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> Phase F5a of the network administration master plan (private repo `veltylabs/mjosefa-cms`; you do
> not need it). **Depends on `webtyp.com/network` v0.1.0** (phase F2) **and `webtyp.com/view` with
> actions** (`view.Action`, phase F6a). First line of work:
> `go get webtyp.com/network@latest webtyp.com/view@latest`; if either is missing what this plan uses,
> STOP and report — never add a `replace`, never copy their types. This module does **not** depend on `device_manager` or
> `veltylabs/mikrotik`.

# Plan — `github.com/veltylabs/network_manager`

Read first, in this repo: [AGENTS.md](../AGENTS.md) (the rules of every `veltylabs/modules/*` repo —
whitelist, blacklist, op handler shape, status codes, multi-tenancy, testing; they are binding and
not repeated here) and [docs/ARCHITECTURE.md](ARCHITECTURE.md) (entities, flow, ops table — the spec).

Reference implementation to copy the shape from (public):
`https://github.com/veltylabs/device_manager` — `model.go`, `module.go`, `ops.go`, `view.go`,
`ui/browser.go`, `migrate/migrate.go`, `seed/`, `web/`, `tests/setup_test.go`. Mirror its file layout,
`Deps`/`New` validation, op handler shape and status mapping exactly.

The contract (`webtyp.com/network`): read `plan.go`, `network.go`, `access.go`, `errors.go` and
`mem/gateway.go` in the module cache after `go get`.

## Development rules

- AGENTS.md whitelist/blacklist. `webtyp.com/network` is whitelisted; `webtyp.com/network/mem` is
  allowed **only** in `tests/`, `seed/` and `web/`.
- No stdlib `errors`/`strings`/`strconv`/`fmt` in the root package — `webtyp.com/fmt`. No `map`, no
  `reflect`.
- Every enum literal only in exported constants; string literals forbidden in logic.
- Never hand-edit `model_orm.go`: run `ormc` at the repo root (install with
  `go install webtyp.com/ormc/cmd/ormc@latest` if missing).
- Tests in `tests/` (package `tests`), runner `gotest ./...`. Never export a symbol only for tests.
- Delete the gonew stub (`network_manager.go`: `type NetworkManager struct{}`, `New()`).

## Design gate

**1. Prior art.**
- **Terraform Cloud / Atlantis**: a "plan" is shown to a human in a UI, and "apply" runs only that
  reviewed plan; a run history is kept. Adopted: `plan_network` → review → `apply_network(fingerprint)`
  → `NetworkApply` history.
- **UniFi Network Controller / Omada**: the controller holds site settings in its database and
  provisions devices; it also lists connected clients. Adopted: `NetworkSetting` in the DB,
  `list_connections`.
- **NetBox + a provisioning job**: inventory (source of truth) separate from the job that pushes it.
  Adopted: inventory stays in `device_manager`; this module only orchestrates through the contract.
- Why different: it is brand-agnostic (the gateway is injected) and never applies without a reviewed
  fingerprint.

**2. Novice-name test.** Ops read as sentences: "plan network", "apply network", "list
connections", "discover network", "import network". Entities: `NetworkSetting`, `NetworkApply`,
`PlanItem`, `ConnectionRow`, `DiscoveredRow`.

**3. Complexity ledger.**
```
Concepts the developer must learn   +1 module (settings + plan/apply ops); router brand: 0
Files they must touch to do X       wiring in an app: 1 composition-root block
Lines at the call site              networkmanager.New(db, Deps{…}) + MountOperations — 2
Ways to do the same thing           0 (no other path changes the router from the app)
```

**4. Where it belongs.** A domain module (bounded concern: making the network follow the inventory).
Inventory stays in `device_manager` (one concern per repo); the router brand stays in
`veltylabs/mikrotik`; the boundary types live upstream in `webtyp.com/network`.

**5. What it deletes.** The gonew stub. Operationally: maintaining the router by hand per device.

## Stage 1 — `model.go`

```go
package networkmanager

const ModelName = "network_manager"

const (
	ResourceNetwork        = "network"
	ResourceNetworkSetting = "network_setting"
)

// PlanItem categories — the ONLY place these literals exist.
const (
	PlanItemChange   = "change"
	PlanItemConflict = "conflict"
	PlanItemWarning  = "warning"
)

const (
	TopicNetworkApplied = "network_manager.network.applied"
)

var (
	ErrNotConfigured = fmt.Err("network settings not configured")
	ErrNotFound      = fmt.Err("network setting not found")
)

type ValidationError struct{ Err error } // same as device_manager
```

`model.Definition`s (run `ormc` afterwards). **Exact form** — `ormc` only discovers package-level
**variables** whose name ends in `Model` and whose value is a `model.Definition{…}` composite literal;
it never reads struct types:

```go
var NetworkSettingModel = model.Definition{
	Name: "network_setting",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		// …
	},
}
```

`ormc` generates the struct (`NetworkSetting` — the variable name minus `Model`), its list type and
helpers into `model_orm.go`. Never declare `type NetworkSetting…` or `type …Model struct` by hand
(copy the shape of `device_manager/model.go`'s `var DeviceModel = model.Definition{…}`).


- `NetworkSettingModel` — table `network_setting`:
  `id` (Text, PK, OmitEmpty), `tenant_id` (`model.Text()`, NotNull),
  `dhcp_server` (`input.Text()`, NotNull, `Permitted{Letters, Numbers, Spaces: true, Extra: []rune{'-','_','.'}, Minimum: 1, Maximum: 64}`),
  `dynamic_pool` (`input.Text()`, OmitEmpty, same Permitted with `Minimum: 0`),
  `filter_dns` (`input.IP()`, NotNull),
  `unregistered` (closed-options radio — package-local `unregisteredPolicy()` returning
  `input.Radio(fmt.KeyValue{Key: network.UnregisteredNoAddressName, Value: "No IP"}, fmt.KeyValue{Key: network.UnregisteredLocalName, Value: "Local network only"})`, NotNull),
  `updated_at` (`model.Int()`, OmitEmpty).
  **Field charset check**: RouterOS names contain spaces and `-` (`"dhcp LAN"`, `"default-dhcp"`).
  Stage 6 tests that both validate. If they fail because `input.Text()`'s own charset still applies
  despite the field `Permitted`, STOP and report it in the PR (upstream defect in `webtyp.com/model`
  / `webtyp.com/input`) — do not switch to `model.Text()` to dodge it.
- `NetworkApplyModel` — table `network_apply`, output-only (base kinds, no widgets):
  `id` (PK), `tenant_id`, `fingerprint` (Text), `changes` (Int), `applied_at` (Int).
- Output-only, transport-only (no `DB`):
  `PlanItemModel` (`id` = position `"1"`, `"2"`…, `category`, `kind`, `object`, `name`, `mac`,
  `before`, `after`, `reason`, `online` (Bool: the host's MAC is in `Connections()`), `fingerprint`),
  `ConnectionRowModel` (`mac`, `ip`, `host_name`, `source` (`"dhcp"`/`"arp"` — constants), `registered` (Bool), `name`),
  `DiscoveredRowModel` (`mac`, `ip`, `name`, `internet` (Bool)),
  `ImportReportModel` (`created` Int, `skipped` Int),
  `ApplyNetworkArgsModel` (`fingerprint` Text, NotNull).
  Kind names for `PlanItem.kind`: `"add"`, `"update"`, `"remove"`, `"adopt"` (constants; one
  unexported switch maps `network.ChangeKind` → name).

## Stage 2 — `module.go`

```go
type Deps struct {
	IDs       model.IDGenerator    // required
	Hosts     network.HostSource   // required — the inventory (device_manager in an app)
	Importer  network.HostImporter // required — the inventory, for import_network
	Gateway   network.Gateway      // required — the router (veltylabs/mikrotik in an app)
	Publisher events.Publisher     // optional — nil disables publishing silently
	TenantID  string               // required
}

func New(db *orm.DB, deps Deps) (*Module, error)
```
Each missing required dep → error `network_manager: Deps.<Name> is required` (same wording pattern
as device_manager).

Service methods (all take `tenantID string` first, every query/update conditioned on `tenant_id`):

- `GetSetting(tenantID) (NetworkSetting, error)` — `ErrNotFound` when absent.
- `SaveSetting(s NetworkSetting) (NetworkSetting, error)` — validate (`Validate(action)`), then
  `network.ParseUnregistered(s.Unregistered)` (error → `ValidationError`), canonicalize
  `filter_dns` with `input.CanonicalIP`, upsert by tenant (one row per tenant), publish nothing.
- `Desired(tenantID) (network.Desired, error)` — settings (`ErrNotFound` → `ErrNotConfigured`) +
  `deps.Hosts.Hosts()`.
- `Plan(tenantID) (network.Plan, []network.Connection, error)` — `Desired` → `Gateway.Plan`; also
  returns `Gateway.Connections()` to mark `online`.
- `Apply(tenantID, fingerprint string) (NetworkApply, error)` — `Desired` →
  `Gateway.Apply(desired, network.Fingerprint(fingerprint))`; on success insert a `NetworkApply`
  (`changes = len(plan.Changes)`, `applied_at = time.Now()`) and publish
  `events.Event{Topic: TopicNetworkApplied, Payload: &apply}`.
- `Connections(tenantID) ([]ConnectionRow, error)` — `Gateway.Connections()` cross-referenced by MAC
  with `Hosts.Hosts()` (`registered`, `name`).
- `Discover() ([]DiscoveredRow, error)`, `Import() (ImportReport, error)` —
  `Gateway.Discover()` then `Importer.ImportHosts(found)`.
- `ListApplies(tenantID) ([]NetworkApply, error)` — newest first.

## Stage 3 — `ops.go`

Exactly the ops table of docs/ARCHITECTURE.md, with `.Requires(resource, action)` as listed
(`save_network_setting`: `model.Create|model.Update`; `import_network`: `model.Create`;
`apply_network`: `model.Update`). No-args ops `.Accepts(nil)`.

Status mapping (AGENTS.md convention):
- decode/validation/`ErrNotConfigured`/`network.IsInvalid(err)` → 400
- `ErrNotFound` → 404
- `network.ErrPlanStale`, `network.ErrConflicts` → 409
- anything else (gateway unreachable, DB) → 500

`plan_network` builds the `PlanItemList`: changes first (category `change`), then conflicts, then
warnings, every row carrying the plan's fingerprint; an empty plan returns an empty list.

## Stage 4 — `view.go` (Presenters, `view`+`model`+`router` only)

- `NewSettingView(caller) view.Presenter` — list/save over `get_network_setting`/`save_network_setting`
  (follow `device_manager/view.go`; the list shows the single row).
- `NewPlanView(caller) view.Presenter` — list over `plan_network`
  (`Item()`: `Label` = object, `Description` = `before → after` or the reason), with one action:
  ```go
  Actions: []view.Action{{
      Op:      OpApplyNetwork,
      Label:   labelApply,   // "Apply" — constant
      Confirm: confirmApply, // "Apply these changes to the router?" — constant
      Args: func(records []model.Model) model.Encodable {
          // every row carries the same fingerprint; the button is disabled when there are no rows
          item := records[0].(*PlanItem)
          return &ApplyNetworkArgs{Fingerprint: item.Fingerprint}
      },
  }},
  ```
  (spec of `view.Action`: `webtyp.com/view` SPECS §9 in the module cache).
- `NewConnectionView(caller) view.Presenter` — read-only list over `list_connections`
  (`Label` = name or host_name or MAC; `Description` = `ip · registered/unregistered`).
- `NewApplyHistoryView(caller) view.Presenter` — read-only list over `list_network_applies`.

The Apply button is drawn by the renderer (`crudview`) from that action; this module only declares it.

## Stage 5 — `migrate/`, `ui/`, `seed/`, `web/`

- `migrate/migrate.go`: `Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error` creating
  `NetworkSetting` and `NetworkApply` (shape of device_manager's).
- `ui/`: `ID = "network_manager"`, `Label = "Red"`; `Browser(caller, ids, tenantID)` registers the
  four presenters with `crudview` (copy device_manager's `ui/browser.go`; one `platformd.UIModule`
  per screen is fine: IDs `network_manager`, `network_manager_plan`, `network_manager_connections`,
  `network_manager_history`). Icon in `ui/svg.go` (`//go:build !wasm`), like device_manager.
- `seed/seed.go`: `Load(m *Module, tenantID string)` saves one `NetworkSetting`
  (`dhcp_server="dhcp1"`, `filter_dns="1.1.1.3"`, `unregistered=local`, `dynamic_pool="pool1"`).
- `web/`: demo runnable with `webtyp`, wiring `network/mem` (with two `Connect`ed devices and one
  `AddUnmanaged` entry with `Internet: true`) and a small in-file `HostSource`/`HostImporter` returning
  two hosts.

## Stage 6 — tests (`tests/`)

`tests/setup_test.go`: `orm.New(mem.New())` (storage mem), `network/mem` gateway, a fake inventory
type implementing `network.HostSource` + `network.HostImporter` (hosts settable per test; records
what `ImportHosts` received), `router/mock` registry. Cases (each asserts results):

1. `save_network_setting` then `get_network_setting` round-trip; `"dhcp LAN"` and `"default-dhcp"`
   validate; `unregistered="bogus"` → 400.
2. `plan_network` without settings → 400 (`ErrNotConfigured`).
3. Plan with 2 hosts → rows include 2 `change` rows with `kind="add"` for their MACs; every row has the
   same non-empty fingerprint; a host whose MAC was `Connect`ed has `online=true`.
4. `apply_network` with that fingerprint → 200, a `network_apply` row with `changes>0`, one
   `TopicNetworkApplied` event (use `events/mock` or a recording publisher); `plan_network` again →
   empty list.
5. Stale: plan, then `gateway.AddUnmanaged(...)`, apply old fingerprint → 409, no `network_apply` row.
6. Conflict: `AddUnmanaged` holding a desired host's IP with another MAC → plan has a `conflict` row;
   apply → 409.
7. `list_connections`: a connected registered MAC → `registered=true` with the inventory name; an
   unknown MAC → `registered=false`.
8. `discover_network` lists the unmanaged entry; `import_network` passes it to `ImportHosts` and
   returns the importer's counts.
9. Tenant isolation: tenant B cannot read tenant A's setting or apply history.
10. Widget regression: `form.New` over `NetworkSetting` yields inputs for `dhcp_server`,
    `dynamic_pool`, `filter_dns`, `unregistered` only.
11. `New` with each required dep missing → the documented error.
12. `NewPlanView` over a fake `router.Caller`: `Actions()` has one action `apply_network`; after a
    reload with plan rows, `Run("apply_network", …)` calls `network_manager.apply_network` with the
    rows' fingerprint.

## Stage 7 — docs

`README.md` (replace the stub): purpose, quick start (`New` with `Deps`, `MountOperations`, the four
presenters), ops table (copy from ARCHITECTURE), key files table, link to ARCHITECTURE and
`docs/diagrams/database.md` (create it: Mermaid ERD of `network_setting` and `network_apply`).
Verify ARCHITECTURE against the code.

## Acceptance criteria

- `gotest ./...` green; `GOOS=js GOARCH=wasm go build .` succeeds for the root package.
- `grep -rn "device_manager\|veltylabs/mikrotik" --include=*.go .` → empty.
- `grep -rn "network/mem" --include=*.go . | grep -v "^./tests/\|^./seed/\|^./web/"` → empty.
- `grep -rn "map\[" --include=*.go . | grep -v model_orm.go` → empty.
- `grep -rn "type NetworkManager struct" .` → empty.

| Stage | Files | Done when |
|---|---|---|
| 1 | `model.go`, `model_orm.go` (ormc) | generated |
| 2 | `module.go` | compiles |
| 3 | `ops.go` | 8 ops mounted |
| 4 | `view.go` | 4 presenters |
| 5 | `migrate/`, `ui/`, `seed/`, `web/` | demo runs with `webtyp` |
| 6 | `tests/*.go` | 12 cases green |
| 7 | `README.md`, `docs/diagrams/database.md` | done |
