← [Index](PLAN.md) | Next → [Stage 2](PLAN_STAGE_2_module_go.md)

> Read AGENTS.md and the index `PLAN.md` (rules, tables, external calls) first. Commit when this stage compiles.

## Stage 1 — `definition.go`

- Set `Allow` on each table exactly as in «Tables, stores and rules»; a table not listed there gets no `Allow` (internal).
- Mark `ServerOnly: true`: `tenant_id`, `created_at`, `updated_at`, and every field the server fills (`status`, `actor_id`, counts, `applied_at`).
- Delete every `*Args` / `*Result` definition that only carried a removed operation (`grep -n "ArgsModel\|ResultModel" model.go` → only the ones this plan keeps).
- New table `NetworkImportModel` (name `network_import`): `id`, `tenant_id`, `status`, `imported` (`model.Int()`), `skipped` (`model.Int()`), `actor_id`, `created_at` (all but `id` `ServerOnly`); add it to `migrate/`. `NetworkApplyModel` gains `status` and `actor_id` (`ServerOnly`); `changes` and `applied_at` become `ServerOnly` (the call fills them). Constants in `const.go`: `ApplyPending`, `ApplyDone`, `ImportPending`, `ImportDone` (values `pending` / `done`). No `failure` column: the reason of a failed attempt lives in the queue row. New typed error constants (house pattern, in the errors file): `ErrPlanStale` = "the router changed after you reviewed the plan, so nothing was applied; open the plan again and apply the new one" and `ErrNotPending` = "this item is already finished, so there is nothing to retry".
- Declare `ExternalCalls: []model.ExternalCallDefinition{...}` on:
  - `NetworkApplyModel`: `{Name: "ApplyGateway", Mode: model.RetryInBackground, MaxAttempts: 3}`.
  - `NetworkImportModel`: `{Name: "RunImport", Mode: model.RetryInBackground, MaxAttempts: 3}`.
- `ResourceNetwork` is deleted: README permissions table lists each table and read model as its own resource.
- **Keep every `*Args` model that an exported reader's signature still uses** (`grep -n "^func (m \*Module) [A-Z][A-Za-z]*([^)]*Args" *.go` lists them): mark each `Wire: true`; delete only the `*Args` no remaining signature uses.
- D52 (the entity is its own filter): a filter model (`*Filter`, `*FilterModel`, an `*ArgsModel` that only carried filter values) is not kept: its fields become fields of the read model or table (`OmitEmpty`), and the rule reads them from its receiver; delete the filter type and its `Wire`/`NewFilter` plumbing.
- Add `var Definition = model.ModuleDefinition{Name: "network_manager", Entities: []model.Definition{…every table and read model of this plan…}}` and, on every table and read model, `Rules` as «Tables, stores and rules» says (`AddMyBusinessRule` where a rule is named, `HaveNoBusinessRule` for the other allowed actions; a read model: `Read: model.AddMyBusinessRule`). Delete the hand topics (`Topic*`) and `Err*`/resources no code uses any more.

## Read models (`PlanItem`, `ConnectionRow`, `DiscoveredRow`): exact

A model without a table has NO `DB:` marks and NO primary key: `Allow: model.Read`, `Rules: model.Rules{Read: model.AddMyBusinessRule}`, listed in `Definition.Entities`. `definitionc` writes `rule_<model>_read.go` with the method `Read(a crud.CurrentUser, filter *X, args crud.ReadArgs, m *Module) (model.ModelSlice, error)`. A `Search` skeleton appears only for a model that HAS `DB:` marks (a table): if `definitionc` asks for a primary key or writes `Search`, a `DB: &model.FieldDB{PK: true}` mark is still on that field: remove it (never add one to make `definitionc` compile).

### Correction: a read model carries `Wire: true` (reference: `work_schedule/definition.go:48-60`)

`definitionc` treats a model as a TABLE (and asks for `Search` and a primary key) unless it is declared `Wire: true`. Declare each of `PlanItem`, `ConnectionRow`, `DiscoveredRow` exactly like this example (a published read model of this ecosystem):

```go
var PlanItemModel = model.Definition{
	Name:  "plan_item",
	Wire:  true, // not a table: no DB marks, no primary key, nothing is stored
	Allow: model.Read,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "kind", Type: model.Text(), OmitEmpty: true},
		// … the fields of today's PlanItem, every one OmitEmpty, no DB: mark …
	},
}
```

`definitionc` then writes `rule_plan_item_read.go` with `Read(a, filter *PlanItem, args crud.ReadArgs, m *Module) (model.ModelSlice, error)`. Never add `Virtual: true` or a `DB:` mark to satisfy the generator.
