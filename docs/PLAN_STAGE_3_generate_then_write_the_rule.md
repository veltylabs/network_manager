← [Stage 2](PLAN_STAGE_2_module_go.md) | Next → [Stage 4](PLAN_STAGE_4_readers_used_by_other_module.md)

> Read AGENTS.md and the index `PLAN.md` (rules, tables, external calls) first. Commit when this stage compiles.

## Stage 3 — generate, then write the rules & external calls

1. Delete `ops.go`, `crud.go`, `queries.go`, `model_crud.go`, `model_orm.go` and `definition_crud.go` (those that exist). Run `definitionc`: it writes one `definition_gen.go` (the types, schemas, external call types `NetworkApplyApplyGateway`, `NetworkImportRunImport`, and on `Module`, `ModelName()`, `Definition()`, `MountOperations()`) and the skeletons:
   - `rule_<entity>_<verb>.go` for every action declared `AddMyBusinessRule`.
   - `call_<entity>_<call>.go` for every call in `ExternalCalls`.
2. Add `store.go`: for each table of «Tables, stores and rules» that has a Store, `// Store<Entity> is the store of <entity>.` `func (m *Module) Store<Entity>() crud.Store { return m.store }`.
3. Fill each rule skeleton (the receiver is the record; `ApplyPending` etc. are the constants of Stage 1):

```go
// rule_network_apply_create.go
func (p *NetworkApply) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	cur, _, err := m.Plan(p.TenantId)
	if err != nil {
		return crud.RuleAnswer{}, err
	}
	if string(cur.Fingerprint) != p.Fingerprint {
		return crud.RuleAnswer{}, &model.FieldError{Field: "fingerprint", Err: ErrPlanStale}
	}
	p.Status, p.ActorId = ApplyPending, a.UserID
	return crud.ApplyAndPublish(NetworkApplyApplyGateway{}), nil
}

// rule_network_apply_update.go
func (next *NetworkApply) Update(a crud.CurrentUser, tx crud.Tx, old *NetworkApply, fields []string, m *Module) (crud.RuleAnswer, error) {
	if old.Status != ApplyPending {
		return crud.RuleAnswer{}, &model.FieldError{Field: "status", Err: ErrNotPending}
	}
	return crud.ApplyAndPublish(NetworkApplyApplyGateway{}), nil
}
```

   `NetworkImport.Create` sets `p.Status, p.ActorId = ImportPending, a.UserID` and returns `crud.ApplyAndPublish(NetworkImportRunImport{})`; `NetworkImport.Update` is the same as the apply one with `ImportPending`/`NetworkImportRunImport{}`. `NetworkSetting` rules: the checks of `SaveSetting`. Rejection: `return crud.RuleAnswer{}, err`.
4. Fill the two calls (the receiver `p` is the stored row; both run in the background: `a` is empty, never read it; `Gateway` and `Importer` are required by `New`, never nil):

```go
// call_network_apply_apply_gateway.go
func (p *NetworkApply) ApplyGateway(a crud.CurrentUser, m *Module, key crud.IdempotencyKey) (crud.FieldsToSave, error) {
	d, err := m.Desired(p.TenantId)
	if err != nil {
		return nil, err
	}
	plan, err := m.gateway.Apply(d, network.Fingerprint(p.Fingerprint))
	if err != nil {
		return nil, err
	}
	p.Status, p.Changes, p.AppliedAt = ApplyDone, int64(len(plan.Changes)), time.NowNano()
	return crud.Fields("status", "changes", "applied_at"), nil
}

// call_network_import_run_import.go
func (p *NetworkImport) RunImport(a crud.CurrentUser, m *Module, key crud.IdempotencyKey) (crud.FieldsToSave, error) {
	found, err := m.gateway.Discover()
	if err != nil {
		return nil, err
	}
	res, err := m.importer.ImportHosts(found)
	if err != nil {
		return nil, err
	}
	p.Status, p.Imported, p.Skipped = ImportDone, int64(res.Created), int64(len(res.Skipped))
	return crud.Fields("status", "imported", "skipped"), nil
}
```

Delete every `Op*` constant, `res*`/`Resource*` variable, every `op*` handler, every `view.MountSave`/`view.MountDelete`/`reg.Operation` call and the `webtyp.com/router`, `webtyp.com/rpc` imports that only they used (`grep -rn "reg.Operation\|view.MountSave\|view.MountDelete\|router.Context\|rpc.WriteError" --include=*.go . | grep -v _test` → empty).
