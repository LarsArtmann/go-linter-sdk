# go-linter-sdk

Shared scaffolding for LarsArtmann Go linters — `Rule` interface, registry, detector adapter, exit codes. Eliminates the per-linter converter layer by codifying the `go-structure-linter` pattern: rules emit `finding.Finding` directly.

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/go-linter-sdk.svg)](https://pkg.go.dev/github.com/larsartmann/go-linter-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/larsartmann/go-linter-sdk)](https://goreportcard.com/report/github.com/larsartmann/go-linter-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**[pkg.go.dev](https://pkg.go.dev/github.com/larsartmann/go-linter-sdk)**

---

## Why?

Three LarsArtmann linters — `branching-flow`, `erraudit`, `go-structure-linter` — each independently reinvented the same three layers:

| Layer                                       | branching-flow | erraudit             | go-structure-linter                    |
| ------------------------------------------- | -------------- | ------------------- | -------------------------------------- |
| Rule interface                              | custom         | custom              | custom                                 |
| Registry                                    | custom         | custom              | custom                                 |
| **Violation → `finding.Finding` converter** | **1,871 LOC**  | **1,214 LOC**       | **0** (`type Issue = finding.Finding`) |

The third row is the killer. branching-flow and erraudit each maintain a substantial bridge package purely because their native domain types (`Violation`, `ErrorViolation`) predate `finding.Finding`. Every new finding field requires touching the converter. Every refactor cascades.

`go-structure-linter` got it rightest by aliasing `Issue = finding.Finding` — no converter at all. **`go-linter-sdk` codifies that pattern.** A rule emits `finding.Finding` directly via `finding.NewBuilder(...)`, so there is no intermediate type to convert. A linter that adopts this SDK ships a `rules.go` file and a `main.go` one-liner — the registry, detection, and exit codes are shared.

---

## Installation

```bash
go get github.com/larsartmann/go-linter-sdk
```

Requires Go 1.26+ and [`go-finding`](https://github.com/larsartmann/go-finding) v1.2+.

---

## Usage

### A 5-line linter

```go
package main

import (
    "context"
    "os"

    "github.com/larsartmann/go-finding"
    "github.com/larsartmann/go-linter-sdk"
)

var registry = linter.NewRegistry()

func init() {
    registry.Register(linter.RuleFunc{
        Meta: linter.RuleMeta{
            Name:        "no-fmt-println",
            Description: "fmt.Println is banned in libraries; use a logger",
            Cat:         linter.CategoryStyle,
            Sev:         finding.SeverityWarning,
        },
        Run: func(ctx context.Context, dir string) ([]finding.Finding, error) {
            // ... scan dir, build findings via finding.NewBuilder(...) ...
            return nil, nil
        },
    })
}

func main() {
    report, _ := registry.Run(context.Background(), ".")
    os.Exit(linter.ExitCodeFromReport(report))
}
```

### Plug into BuildFlow's DAG

```go
detector := linter.DetectorFromRegistry(registry, "my-linter")
// Pass `detector` to toolsdk.Spec{ Detect: detector, ... } or directly to
// BuildFlow's domain.DetectorFromFinding(detector, toolName).
```

The working directory is read via `finding.WorkingDirFromContext(ctx)`, so module fan-out and per-directory runs work transparently.

---

## API

### Types

| Type              | Purpose                                                                                                                                                                           |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Rule` interface  | `Name()` / `Description()` / `Category()` / `Severity()` / `Check(ctx, dir) ([]Finding, error)`                                                                                   |
| `RuleFunc` struct | Adapter: combines a `RuleMeta` header with a `Run` closure to satisfy `Rule`                                                                                                      |
| `RuleMeta` struct | Declarative identity: `Name`, `Description`, `Cat`, `Sev`                                                                                                                         |
| `Category`        | `CategoryDesign` / `CategoryStructure` / `CategoryErrorHandling` / `CategoryCorrectness` / `CategoryStyle` / `CategoryPerformance` / `CategorySecurity` / `CategoryConfiguration` |
| `Registry`        | Holds rules; thread-safe with `sync.RWMutex`                                                                                                                                      |

### Functions

| Function                            | Returns                  | Purpose                                                        |
| ----------------------------------- | ------------------------ | -------------------------------------------------------------- |
| `NewRegistry()`                     | `*Registry`              | Empty registry                                                 |
| `(*Registry).Register(rule)`        | —                        | Add a rule (panics on duplicate name)                          |
| `(*Registry).All()`                 | `[]Rule`                 | Snapshot of registered rules                                   |
| `(*Registry).Run(ctx, dir)`         | `*finding.Report, error` | Run all rules; aggregate findings                              |
| `DetectorFromRegistry(r, toolName)` | `finding.Detector`       | Adapt registry to the canonical ecosystem Detector interface   |
| `ExitCodeFromReport(report)`        | `int`                    | 0 if clean, 1 if findings — the ecosystem exit-code convention |

---

## Design notes

- **Depends only on `go-finding`** (the ecosystem hub) so any consumer — CLI, library, LSP server, golangci-lint plugin — can adopt it without coupling.
- **Rules emit `finding.Finding` directly.** No intermediate Violation/Issue type. No converter layer. This is the core design decision.
- **`Registry.Register` panics on duplicate names** — duplicate rule names silently shadow each other at runtime, which is a programming error. Surface it at startup.
- **`ExitCodeFromReport` is binary** (0 clean / 1 any findings). The ecosystem convention; tools that want severity-tiered exit codes do their own mapping.
- **Working directory via context.** `DetectorFromRegistry` reads `finding.WorkingDirFromContext(ctx)`. This lets BuildFlow's module fan-out target individual modules without per-rule plumbing.

---

## Migration path

Existing linters migrate **incrementally** — one rule at a time:

1. Add `go-linter-sdk` as a dependency
2. Pick one rule; convert its native type to emit `finding.Finding` via `finding.NewBuilder(...)`
3. Wrap it in `linter.RuleFunc{Meta: ..., Run: ...}` and `Register` it
4. Repeat for each rule
5. Once all rules are migrated, delete the converter package (`pkg/finding/` in branching-flow / erraudit)

Each step compiles and runs independently. No big-bang migration.

---

## Consumers

Planned:

- `go-structure-linter` — cleanest existing pattern; pilot target
- `branching-flow` (1,871 LOC of converters to delete)
- `erraudit` (1,214 LOC of converters to delete)
- Future linters

No active consumers yet.

## Status

Early. The `Rule`/`Registry`/`Detector` core is stable. The migration story is proven in `go-structure-linter`'s existing `type Issue = finding.Finding` pattern but no linter has been ported yet.

## License

MIT — see [LarsArtmann/template-LICENSE](https://github.com/LarsArtmann/template-LICENSE).
