← [Stage 5](PLAN_STAGE_5_browser_side.md) | Next → [Stage 7](PLAN_STAGE_7_translations_and_documentati.md)

> Read AGENTS.md and the index `PLAN.md` (rules, tables, external calls) first. Commit when this stage compiles.

## Stage 6 — tests (`tests/`, through `crud.LoopbackActingAsWithQueue(nil, "u1", &callQueue{}, mod)`; rows in `docs/TESTS.md`)

Delete tests of removed code.

`tests/queue_double_test.go` (copy it as is; it is the only queue the tests use):

```go
package tests

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
)

// callQueue is the test double of crud.CallQueue: it records what crud asks and stores nothing.
type callQueue struct {
	added  []crud.CallRequest
	done   []string
	failed []string // the reasons
}

func (q *callQueue) Add(tx crud.Tx, r crud.CallRequest) (string, error) {
	q.added = append(q.added, r)
	return fmt.Sprintf("call-%d", len(q.added)), nil
}
func (q *callQueue) MarkDone(id string) error { q.done = append(q.done, id); return nil }
func (q *callQueue) MarkFailedAttempt(id string, call model.ExternalCallDefinition, reason string) error {
	q.failed = append(q.failed, reason)
	return nil
}
func (q *callQueue) ConfirmCall(id string) error        { return nil }
func (q *callQueue) RejectCall(id, reason string) error { return nil }
```

The caller of every test that touches a call is `crud.LoopbackActingAsWithQueue(nil, "u1", &callQueue{}, mod)` (keep the pointer in a variable to assert `added`, `done`, `failed`).

Operations are `router.Operation{Module: ModelName, Entity: "<entity>", Action: model.<Action>}`, never a string; fixtures follow `crud` v0.0.14 (`reg.AddStoredEntity`, `crud.CurrentUser`). Add `TestMount_MatchesDefinition` (`crud.Mount(reg, nil, nil, mod)` returns nil) and one test that an allowed write publishes a `*crud.Change` on `crud.Topic(ModelName, "<entity>", model.<Action>)`. Required:

| Test | Setup → call → expected |
|---|---|
| `TestApply_StaleFingerprintRejected` | create apply with an old fingerprint → field error on `fingerprint`; gateway not called |
| `TestApply_GatewayCalledViaExternalCall` | fake gateway records calls; create apply with the current fingerprint → `queue.added` has `ApplyGateway`; read the row and call `row.ApplyGateway(crud.CurrentUser{}, mod, "k")` → one gateway call, the row reads `status` = `done` and `changes` = the plan's |
| `TestApply_GatewayFailureKeepsPending` | fake gateway fails → the call returns an error and the row still reads `pending`; update `["status"]` = `ApplyPending` → `queue.added` has a second `ApplyGateway`; on a `done` row the same update → field error on `status` |
| `TestImport_CountsStored` | fake importer returns 3 created, 1 skipped; create import, call `row.RunImport(crud.CurrentUser{}, mod, "k")` → row `imported` = 3, `skipped` = 1, `status` = `done` |
| `TestReadModel_Plan` | setting with hosts → `PlanItem` rows |
| `TestDemo_Valid` (`//go:build wasm`) | `seed.Demo()` → no error, `Validate()` nil, at least one screen |
