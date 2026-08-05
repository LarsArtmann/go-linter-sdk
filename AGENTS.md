# AGENTS.md — go-linter-sdk

Shared scaffolding for LarsArtmann Go linters (`Rule` interface, registry, detector adapter, exit codes). Eliminates the per-linter converter layer by codifying the `go-structure-linter` pattern: rules emit `finding.Finding` directly.

## Critical: GOEXPERIMENT=jsonv2 REQUIRED

This package depends on `go-finding`, which uses `encoding/json/v2` and `encoding/json/jsontext` (Go's experimental JSON v2). **Every `go` command must run with `GOEXPERIMENT=jsonv2`** or the build fails with `build constraints exclude all Go files`.

| How                   | Command                                                               |
| --------------------- | --------------------------------------------------------------------- |
| Nix app (recommended) | `nix run .#test`, `nix run .#build`, `nix run .#lint`                 |
| Dev shell             | `nix develop` (sets GOEXPERIMENT automatically), then `go test ./...` |
| Inline                | `GOEXPERIMENT=jsonv2 go test ./...`                                   |
| BuildFlow             | `nix develop` then `buildflow` (or `GOEXPERIMENT=jsonv2 buildflow`)   |

## Build Commands

```bash
nix run .#test          # go test ./... -count=1
nix run .#test-race     # go test ./... -race -count=1
nix run .#build         # go build ./...
nix run .#lint          # golangci-lint run ./...
nix run .#vet           # go vet ./...
nix run .#coverage      # go test ./... -coverprofile=reports/coverage.out
nix flake check         # validate flake
```

No Makefile. No justfile. Everything via `flake.nix`.

## Architecture

Core types in package `linter`:

- **`Rule` interface** — `ID()` / `Name()` / `Description()` / `Category()` / `Severity()` / `IsEnabledByDefault()` / `Check(ctx, dir) ([]Finding, error)`. Rules emit `finding.Finding` directly (no intermediate Violation/Issue type). Dual identity: `ID()` is the stable key (registry dedup, suppression, filtering, finding emission); `Name()` is the mutable display name.
- **`RuleFunc` struct** — adapter combining `RuleMeta` header + `Run` closure to satisfy `Rule`. Most rules use this. Enabled by default.
- **`OptIn(rf)`** — wraps a `RuleFunc` as disabled-by-default. Use for noisy, experimental, or domain-specific rules.
- **`Registry`** — thread-safe (`sync.RWMutex`) collection of rules. `Register` panics on duplicate ID or empty identity fields (ID, Name, Description, Category). `Get(id)` / `Has(id)` for lookup; `Deregister(id)` for runtime removal.
- **`Registry.Run`** — standalone execution. Fail-fast by default; pass `ContinueOnError()` to run all rules regardless of failures, collecting partial findings and joining errors.
- **`RuleMeta.Validate()`** — checks ID, Name, Description, Cat are non-empty. Called by `Register` through the `Rule` interface.
- **`DetectorFromRegistry`** — adapts a registry to a single `finding.Detector` for BuildFlow integration. Reads working dir from context.
- **`DetectorsFromRegistry`** — returns `[]finding.Detector` (one per rule) for `go-finding/pipeline` integration with per-rule parallelism, timeouts, and error isolation. Each detector is named after the rule's ID.
- **`ExitCodeFromReport`** — binary: 0 if clean, 1 if any findings.
- **`Category`** is an open `string` type. The 8 built-in constants are recommendations; consumers can define custom categories (`Category("api")`).
- **`finding.Confidence` and `finding.FixStrategy`** are per-finding, set via `finding.NewBuilder(...).WithConfidence(...).WithFixStrategy(...)`. More expressive than rule-level defaults.

### Error Wrapping Pattern

`Registry.Run`, `RuleFunc.Check`, and `DetectorFromRegistry`/`DetectorsFromRegistry` wrap rule execution errors into `*RuleError` (carrying the rule ID). Use `errors.AsType[*RuleError](err)` to extract which rule failed, or `errors.Is(err, ErrRuleFailed)` for a boolean check. `Registry.Run` avoids double-wrapping: if the error is already a `*RuleError` (from `RuleFunc.Check`), it passes through.

## Dependencies

- **`go-finding` v1.4.1** — resolved from VCS as a real published tag (no local `replace` directive). `go.mod` carries the pinned version in `require`; consumers and CI both fetch `v1.4.1` directly. For local cross-repo dev against an uncommitted sibling checkout, add a temporary `go.work` (workspace) or a `replace ../go-finding` line — neither is committed.
- Go 1.26+ (uses `encoding/json/v2` experiment).

## Gotchas & conventions

- **Flake `outputs` must declare `self`.** The `outputs` lambda uses an
  `inputs@{ ... }` destructure. Nix 2.34.8+ enforces strict argument checking on
  `@`-patterns, so omitting `self` breaks BuildFlow with
  `function 'outputs' called with unexpected argument 'self'`. Keep `self,`
  first in the destructure (the sibling `go-finding` flake is the reference;
  fixed here in `c13366c`).
- **`reports/` must exist for BuildFlow's `test-coverage` step.** It is kept
  durable via `reports/.gitkeep` plus a `.gitignore` exception
  (`!reports/`, `reports/*`, `!reports/.gitkeep`) — don't remove the marker.
  The `coverage` flake app writes `reports/coverage.out`.
- **`.golangci.yml` carries the `goexperiment.jsonv2` build tag.** Without it
  golangci-lint can't compile `go-finding`'s transitive `encoding/json/v2`
  imports and silently falls back to defaults. The `lint` app auto-discovers it.
- **Tests are black-box (`package linter_test`), not white-box.** This was a
  deliberate decision to satisfy `testpackage` natively (no suppression) and
  test the real consumer contract. Test factories (`makeRule`, `failingRule`)
  return the concrete `RuleFunc`, not the `Rule` interface, so `ireturn` has
  nothing to flag either. There are no unexported symbols that tests need to
  reach — if one is added, use an `export_test.go` shim rather than reverting
  to white-box.
- **Error matching: `errors.AsType` for types, `errors.Is` for sentinels.**
  `registry.go` uses `errors.AsType[*RuleError]` (Go 1.26 generic); `errors.Is`
  is correct for the `ErrRuleFailed` sentinel. Do not "migrate" `errors.Is`
  calls — see the `hierarchical-errors` skill for the decision tree.
- **CI fetches `go-finding` via VCS auth, not a sibling clone.** There is no
  `replace` directive. `.github/workflows/ci.yml` sets `GOPRIVATE` and configures
  git to use the default `GITHUB_TOKEN` so the private `v1.4.1` tag resolves
  from GitHub directly. Local dev uses `GOPRIVATE` (global) plus an SSH
  `insteadOf` rewrite; a one-off tidy needs `GIT_CONFIG_COUNT=1
GIT_CONFIG_KEY_0="url.git@github.com:.insteadOf"
GIT_CONFIG_VALUE_0="https://github.com/"` if global git config is read-only.
- **Doc one-home rule: FEATURES = what exists, ROADMAP = what's next.**
  `FEATURES.md` tracks only capabilities that have code today
  (`FULLY_FUNCTIONAL` / `PARTIALLY_FUNCTIONAL` / `BROKEN`). Not-yet-built
  capabilities live exclusively in `ROADMAP.md` until they graduate to
  `TODO_LIST.md`. This kills the FEATURES↔ROADMAP PLANNED split-brain that
  recurred across sessions 2-5.
- **Auto-git daemon manages commits.** The daemon commits automatically with
  generated messages (sometimes low-quality). Do not rewrite history to fix
  them — `CHANGELOG.md` is the narrative of record. Focus on keeping the
  CHANGELOG accurate rather than the `git log` pretty.

## Consumers (Planned)

`go-structure-linter` (pilot), `branching-flow`, `erraudit`. No active consumers yet.
