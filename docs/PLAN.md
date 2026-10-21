---
PLAN: "refactor!: build destinations — web/browser.go and ui/browser.go with //go:build browser"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 17907113701879890727
PR: https://github.com/veltylabs/network_manager/pull/3
REVIEWED_COMMIT: 0167ebbd2b5e6b0624280fc8019ca64eeee0d836
---

> Read AGENTS.md at the repo root before any change.

# network_manager: build destinations migration (`web/browser.go`, `ui/browser.go`)

Part of [BUILD_DESTINATIONS_MASTER_PLAN.md](https://github.com/webtyp/docs/blob/main/BUILD_DESTINATIONS_MASTER_PLAN.md) (D1, D4, D6, D15):
- The browser entry for the demo page is `web/browser.go` (`//go:build browser`) which compiles into `browser.wasm`.
- `ui/browser.go` carries `//go:build browser`.
- `ui/svg.go` keeps `//go:build !wasm` (D7, read by `sitec` at build time).
- Rename `web/client.go` to `web/browser.go` and replace its build tag `//go:build wasm` with `//go:build browser`.
- Any test that exercises browser UI (e.g. `tests/ui_test.go` calling `ui.Browser`) must carry `//go:build browser`.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Environment

```bash
export GOPRIVATE=webtyp.com/*,github.com/veltylabs/*
export GOFLAGS=-mod=mod
```

A PR that does not compile, or whose tests could not run, is a failed delivery. Never pause to ask permission or report progress: when the Delivery check passes, commit, push and open the PR. Stuck for 30 minutes: open the PR anyway listing what fails.

## Stage 1 — Rename `web/client.go` to `web/browser.go`

- [x] 1. Rename `web/client.go` to `web/browser.go` (`git mv web/client.go web/browser.go`).
- [x] 2. In `web/browser.go`, change the build tag at the top:
   ```go
   //go:build browser
   ```
- [x] 3. In `ui/browser.go`, ensure the file carries:
   ```go
   //go:build browser
   ```
   at the top before the package declaration.

## Stage 2 — Update tests and docs

- [x] 1. If any test file references `ui.Browser` (such as `tests/ui_test.go`), ensure it carries `//go:build browser` so it runs with the browser tag.
- [x] 2. Check if any test or doc mentions `web/client.go` or `client.wasm`; update to `web/browser.go` and `browser.wasm`.
- [x] 3. Update `AGENTS.md` if it mentions `web/client.go` or `client.go tagged wasm`.
- [x] 4. Run `gotest ./...` and ensure all tests pass.

## Deletions and acceptance criteria

- `grep -rn "web/client\.go" --include=*.go .` returns empty.
- `test -f web/browser.go && test ! -f web/client.go` succeeds.
- `gotest ./...` passes with zero failures.

## Stages

| # | Files | Goal |
|---|---|---|
| 1 | `web/browser.go`, `ui/browser.go` | [x] Rename to `web/browser.go` with `//go:build browser`, add tag to `ui/browser.go` |
| 2 | tests, docs | [x] Verify tests pass with `gotest ./...` |
