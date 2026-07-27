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

- **`Rule` interface** — `Name()` / `Description()` / `Category()` / `Severity()` / `Check(ctx, dir) ([]Finding, error)`. Rules emit `finding.Finding` directly (no intermediate Violation/Issue type).
- **`RuleFunc` struct** — adapter combining `RuleMeta` header + `Run` closure to satisfy `Rule`. Most rules use this.
- **`Registry`** — thread-safe (`sync.RWMutex`) collection of rules. `Register` panics on duplicate names (programming error).
- **`DetectorFromRegistry`** — adapts a registry to `finding.Detector` for BuildFlow integration. Reads working dir from context.
- **`ExitCodeFromReport`** — binary: 0 if clean, 1 if any findings.

### Error Wrapping Pattern

`Registry.Run`, `RuleFunc.Check`, and `DetectorFromRegistry` wrap rule execution errors into `*RuleError` (carrying the rule name). Use `errors.As(err, &ruleErr)` to extract which rule failed, or `errors.Is(err, ErrRuleFailed)` for a boolean check. `Registry.Run` avoids double-wrapping: if the error is already a `*RuleError` (from `RuleFunc.Check`), it passes through.

## Dependencies

- **`go-finding`** via replace directive: `replace github.com/larsartmann/go-finding => ../go-finding`. Both repos must be checked out as siblings under the same parent directory.
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

## Consumers (Planned)

`go-structure-linter` (pilot), `branching-flow`, `hierarchical-errors`. No active consumers yet.
