← [Stage 3](PLAN_STAGE_3_generate_then_write_the_rule.md) | Next → [Stage 5](PLAN_STAGE_5_browser_side.md)

> Read AGENTS.md and the index `PLAN.md` (rules, tables, external calls) first. Commit when this stage compiles.

## Stage 4 — readers used by other modules

Signatures of every exported reader stay (other modules and apps call them). Bodies read through `m.store` (`ReadOne`, `ReadAll` with `crud.Query`), never `m.db` directly; a `tenantID` parameter different from the module's tenant answers «not found» (`found=false` / empty, `err=nil`). Keep `GetSetting`, `Desired`, `Plan`, `Connections`, `Discover`, `ListApplies`; `SaveSetting` keeps its checks as the rule of `NetworkSetting`; `Apply(tenant, fingerprint)` and `Import()` are deleted (their logic is the rules and calls of Stage 3; the app creates `network_apply` / `network_import` records). If `grep -rn "\.Apply(\|\.Import()" --include=*.go . | grep -v _test` finds a caller, say so in the PR.
