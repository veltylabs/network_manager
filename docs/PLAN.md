---
PLAN: "refactor!: network_manager on crud tables (D19/D20) — Allow in the model, persistence by the framework, only domain rules"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.
> Never edit `model_orm.go` or `model_crud.go`: a build error there means the hand-written code is wrong.

# Plan: `network_manager` on crud tables

Today seven operations share one resource (`ResourceNetwork`). Applying the plan to the devices and importing hosts call external systems (`Deps.Gateway`, `Deps.Importer`): they become records with an outbox (D25). `PlanItem`, `ConnectionRow`, `DiscoveredRow` are not tables (the migration creates only `NetworkSetting` and `NetworkApply`): remove their `DB:` marks; they become read models.

Reference implementation of the same migration (read it first, copy its shape):
`patient_directory` — `model.go` (`Allow`), `ops.go` (`MountPatient`), `rules.go` (hooks), `readers.go` (signatures kept, bodies on `m.store`).

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go install webtyp.com/ormc/cmd/ormc@latest
go get webtyp.com/model@latest webtyp.com/crud@latest webtyp.com/orm@latest webtyp.com/view@latest webtyp.com/appkit@latest
```
These versions must contain `Definition.Allow` and `model.TableSlice` (model), `crud.Table`/`crud.Tx`/`crud.Transactional` (crud),
`orm.TenantStore`/`orm.GlobalStore` returning `*orm.Store` (orm), `ormc` generating `Mount<Entity>` + `<Entity>Rules`, and `view`
whose `NewCallerLister` takes `func() model.TableSlice`. If `go doc webtyp.com/orm TenantStore` fails or `ormc` does not write
`func Mount`, stop and say so in the PR. Remove every `replace` from `go.mod`; `go mod tidy`; run `ormc`.

## The rule of this plan (D19/D20/D21 of the architecture master)

The framework persists; the module declares and writes only domain rules.

| The module no longer writes | Who does it now |
|---|---|
| id, `tenant_id`, `created_at`/`updated_at`, the transaction loop, `db.Create`/`db.Update`/`db.Delete`, «not found», batch atomicity, idempotent retry by id, paging by `After`, `tenant_id` in every query, schema checks repeated from the model | `crud` + `orm.TenantStore` / `orm.GlobalStore` |
| an operation per verb (`list_*`, `save_*`, `upsert_*`, `delete_*`) and its handler | `Allow` in the table's `model.Definition` + the generated `Mount<Entity>` |
| an action with its own permission (approve, settle, receive stock…) | a **record created** in its own table, or a **read model** with its own name (D18) |

A module writes: `Allow` per table; `ops.go` with one `Mount<Entity>(reg, store, <Entity>Rules{…})` per offered table and one
`reg.Read(…)` per read model; `rules.go` with the hooks (`Create(c, tx, rec)`, `Update(c, tx, old, next, fields)`,
`Delete(c, tx, rec)`, `Search(c, text)` returning AND-only conditions, `Committed(c, action, rec, fields)` for events after commit).
A hook changes only the columns in `fields` on Update; to write another row it calls `tx.Create`/`tx.UpdateFields`.

## Tables, stores and rules

| Table (`<X>Model`) | Store | `Allow` | Hooks (`rules.go`) — where their rules come from today |
|---|---|---|---|
| `NetworkSetting` | `orm.TenantStore` | `model.Create \| model.Read \| model.Update` | checks of `SaveSetting` (one row per tenant) |
| `NetworkApply` | `orm.TenantStore` | `model.Create \| model.Read \| model.Update` | `Create`: `fingerprint` must equal the current plan's (logic of `Apply` before the gateway call), status `pending`, `actor_id` = `c.UserID`; `Committed`: run the gateway part of `Apply`, write status; `Update`: only `["status"]` back to `pending` (retry). **Outbox (D25):** the record is created with status `pending`; `Committed` calls the external system and then writes the outcome with `m.store.Atomic(func(tx crud.Tx) error { return tx.UpdateFields(rec, id, []string{"status", "failure", …}) })` — status constants `StatusPending = "pending"`, `StatusDone = "done"`, `StatusFailed = "failed"`. A retry is an Update of `status` back to `pending` (its `Committed` calls again). Never call the external system inside a `Create`/`Update` hook. |
| `NetworkImport (**new table** `network_import`)` | `orm.TenantStore` | `model.Create \| model.Read \| model.Update` | `Create`: status `pending`, `actor_id`; `Committed`: run `Import()` and store the counts of its `ImportReport` and the status; `Update`: retry as above |

## Read models (D18 — `crud.Registry.Read`, no table, `Allow: model.Read` only)

| Model | Filter | Answered by |
|---|---|---|
| `PlanItem` (existing; no `DB:` marks) | none | `Plan(tenant)` items |
| `ConnectionRow` (existing; no `DB:` marks) | none | `Connections()` |
| `DiscoveredRow` (existing; no `DB:` marks) | none | `Discover()` |

A read model is a `model.Definition` **without** `DB:` marks and with `Allow: model.Read`; `ormc` gives it `Allow()` and no
`Mount`. Its permission resource is its own `ModelName`: list it in the README permissions table (roles must be granted it).

## Operation mapping (every operation of today has exactly one row)

| Today | After |
|---|---|
| `list_network_settings (+save)` | `network_setting_r/_c/_u` |
| `plan_network` | read model `PlanItem` |
| `apply_network` | `network_apply_c` with `fingerprint` |
| `list_connections` | read model `ConnectionRow` |
| `discover_network` | read model `DiscoveredRow` |
| `import_network` | `network_import_c` |
| `list_network_applies` | `network_apply_r` |

## Stage 1 — `model.go`

- Set `Allow` on each table exactly as in «Tables, stores and rules»; a table not listed there gets no `Allow` (internal).
- Mark `ServerOnly: true`: `tenant_id`, `created_at`, `updated_at`, and every field the server fills (`status`, `failure`, `actor_id`, counts).
- Delete every `*Args` / `*Result` definition that only carried a removed operation
  (`grep -n "ArgsModel\|ResultModel" model.go` → only the ones this plan keeps).
- New table `NetworkImportModel` (name `network_import`): `id`, `tenant_id`, `status`, `failure`, `imported` (`model.Int()`), `skipped` (`model.Int()`), `actor_id`, `created_at` (all but `id` `ServerOnly`); add it to `migrate/`. `NetworkApplyModel` gains `failure` if it lacks it.
- `ResourceNetwork` is deleted: README permissions table lists each table and read model as its own resource.
- **Keep every `*Args` model that an exported reader's signature still uses** (`grep -n "^func (m \*Module) [A-Z][A-Za-z]*([^)]*Args" *.go` lists them): mark each `Wire: true`; delete only the `*Args` no remaining signature uses.
- Run `ormc`: `model_crud.go` has one `func Mount<Entity>(` per table with `Allow`.

## Stage 2 — `module.go`

`Deps` keeps its fields; add `IDs model.IDGenerator` (required, `depErr("IDs")` when nil) if it is missing.
`Module` gains `store *orm.Store`, built once in `New`: `m.store = orm.TenantStore(db, deps.TenantID, deps.IDs)`.
Delete the hand-written persistence (create/update/upsert/patch/delete/list/get bodies that only move rows) once its rules moved to
`rules.go`: the row-moving bodies of `SaveSetting`, `Apply`, `Import`, `ListApplies` (the gateway calls move to `Committed`).

## Stage 3 — `ops.go` and `rules.go`

```go
const ModelName = "network_manager"

func (m *Module) ModelName() string { return ModelName }

var _ crud.Module = (*Module)(nil)

func (m *Module) MountOperations(reg crud.Registry) {
	MountNetworkSetting(reg, m.store, NetworkSettingRules{Create: m.checkSetting, Update: m.checkSettingChange})
	MountNetworkApply(reg, m.store, NetworkApplyRules{Create: m.checkApply, Update: m.retryApply, Committed: m.pushToDevices})
	MountNetworkImport(reg, m.store, NetworkImportRules{Create: m.checkImport, Update: m.retryImport, Committed: m.runImport})
	reg.Read(&PlanItem{}, nil, m.readPlan)
	reg.Read(&ConnectionRow{}, nil, m.readConnections)
	reg.Read(&DiscoveredRow{}, nil, m.readDiscovered)
}
```
`rules.go` holds every hook named above. **Outbox (D25):** the record is created with status `pending`; `Committed` calls the external system and then writes the outcome with `m.store.Atomic(func(tx crud.Tx) error { return tx.UpdateFields(rec, id, []string{"status", "failure", …}) })` — status constants `StatusPending = "pending"`, `StatusDone = "done"`, `StatusFailed = "failed"`. A retry is an Update of `status` back to `pending` (its `Committed` calls again). Never call the external system inside a `Create`/`Update` hook.
Delete every `Op*` constant, `res*`/`Resource*` variable that named a removed operation, every `op*` handler, every
`view.MountSave`/`view.MountDelete`/`reg.Operation` call and the `webtyp.com/router`, `webtyp.com/rpc` imports
(`grep -rn "reg.Operation\|view.MountSave\|view.MountDelete\|router.Context\|rpc.WriteError" --include=*.go . | grep -v _test` → empty).

## Stage 4 — readers used by other modules

Signatures of every exported reader stay (other modules and apps call them). Bodies read through `m.store` (`ReadOne`, `ReadAll` with `crud.Query`), never `m.db` directly; a `tenantID` parameter different from the module's tenant answers «not found» (`found=false` / empty, `err=nil`). Keep `GetSetting`, `Desired`, `Plan`, `Connections`, `Discover`, `ListApplies`; `SaveSetting`, `Apply`, `Import` stay as callers (create the record inside `m.store.Atomic`; the external call still happens only after commit).

## Stage 5 — browser side (D15, D16)

- `view.go` (`//go:build browser`): every `view.NewCallerLister(caller, view.Ops{…}, newList)` takes
  `func() model.TableSlice { return &<Entity>List{} }` and `view.Ops` has no `Allow` (the screen reads `<Entity>List.Allow()`).
- `module/browser.go`, `module/server.go` (if present): `appkit.ModuleBrowser`/`appkit.BrowserDeps` → `view.ModuleBrowser`/`view.BrowserDeps`.
- Demo: `seed/demo.go` (`//go:build wasm`) returning `view.Demo{Label, Timezone: "America/Santiago", ActingAs: "u1",
  Modules: []crud.Module{mod}, Migrations: <the module's ddl.Step list, or nil if it has no migrations>,
  Screens: func(d view.BrowserDeps) []view.Screen { return ui.Screens(<what today's demo main passes>) }}`, built with
  `orm.New(mem.New())`, `New(db, Deps{…})` with the same fakes today's demo uses, and `seed.Load` (when the module has one).
  Delete the old demo: any `web/` demo (`test ! -e web`).
- `grep -rn "webtyp.com/layout\|appkit.BrowserDeps\|appkit.ModuleBrowser\|router/loopback" .` → empty, `go.mod` included.

## Stage 6 — tests (`tests/`, through `crud.LoopbackActingAs("u1", mod)`; rows in `docs/TESTS.md`)

Delete tests of removed code. Operation names: `crud.Op(&<Entity>{}, model.<Action>)` qualified with `ModelName + "."`. Required:

| Test | Setup → call → expected |
|---|---|
| `TestApply_StaleFingerprintRejected` | create apply with an old fingerprint → field error on `fingerprint`; gateway not called |
| `TestApply_GatewayCalledAfterCommit` | fake gateway records calls; create apply → status `done`, one call |
| `TestApply_GatewayFailureRecorded` | fake gateway fails → status `failed`, `failure` set; update `["status"]` to pending → called again |
| `TestImport_CountsStored` | fake importer returns 3 imported, 1 skipped → record shows 3 and 1 |
| `TestReadModel_Plan` | setting with hosts → `PlanItem` rows |
| `TestDemo_Valid` (`//go:build wasm`) | `seed.Demo()` → no error, `Validate()` nil, at least one screen |

## Stage 7 — translations and documentation (Spanish, veltylabs)

`langc sync .` (zero pending, zero reuse warnings). `README.md` (≤150 lines, 4 mandatory sections; «Sin/Con»: hand-written
operations vs `Allow` + rules) with the operations table of the mapping above. `docs/ARCHITECTURE.md`: rules vs persistence.
Translate to Spanish every README, `docs/*`, diagram and code comment still in English; exported symbols without a usage example get one.

## Delivery check

`gotest` passes; `langc sync .` clean; every grep of this plan empty. A tree that does not compile or whose tests did not run is a
failed delivery.
