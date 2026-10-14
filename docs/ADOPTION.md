# Adoption

Who uses `go-linter-sdk`, how, and whether the SDK kept its core promise:
**rules emit `finding.Finding` directly — no custom Issue/Violation converter
layer.** Verified by source inspection on 2026-10-11 (see "Refresh procedure"
below to re-verify).

Consumers pin `go-linter-sdk` per their own `go.mod`; this doc intentionally
duplicates no version numbers (the doc-drift guard, `scripts/check-doc-versions.sh`,
keeps living docs version-pointered).

## Consumers

| Repository                             | Integration pattern                                                                                                                                                              | Converter layer                                                                     |
| -------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `go-humanize-linter`                   | 11 rules as `linter.RuleFunc` in `linter.Registry`; CLI + golangci plugin                                                                                                        | None — born native                                                                  |
| `cmdguard` (`cmdguard/lint`)           | Rules as `linter.RuleFunc`; BuildFlow provider via `go-finding/toolsdk` `Spec`                                                                                                   | None — born native                                                                  |
| `project-meta` (`project-meta/linter`) | `linter.RuleFunc` rules; `linter.Registry`; CLI exit via `linter.ExitCodeByConfidence`                                                                                           | None — born native                                                                  |
| `BuildFlow` (`tools/providers`)        | Runs sibling registries in-process (`Registry.Run`); deleted its `PrioritizedViolation` hierarchy (~4,737 LOC) and 10 converter files in favor of `finding.Finding`              | Deleted (see `BuildFlow/CHANGELOG.md`)                                              |
| `go-structure-linter` (pilot)          | SDK registry/detectors under a rule-level adapter (`internal/rules/sdk_adapter.go`); local `types.Issue` is a **type alias** of `finding.Finding`, so no data translation exists | Converter struct deleted; adapter kept **by design** (ADR-003)                      |
| `samber-linter`                        | Rules are `go/analysis` analyzers (golangci plugin contract); SDK supplies exit-code policy (`linter.ExitCodeByConfidence`)                                                      | Never had one; `analysis.Diagnostic → finding.Finding` mapping is documented design |

## Verdicts on the core claim

- **Born-native consumers (3):** built directly on `linter.RuleFunc` +
  `finding.Finding`; nothing to delete, nothing found.
- **Migrating consumers (2):** `BuildFlow` physically deleted its violation
  hierarchy and converter package; `go-structure-linter` deleted its custom
  `Issue` struct (now an alias) and per-finding translation.
- **Different-shape consumer (1):** `samber-linter` adopts the SDK's exit-code
  contract while keeping `go/analysis` as its rule interface — documented in
  its planning docs as deliberate.

No inspected repository retains a custom intermediate Issue/Violation struct
that is translated field-by-field into findings. The claim holds.

## What "adopted" means here

A repository counts as a consumer when its `go.mod` requires
`github.com/larsartmann/go-linter-sdk` and its code imports the SDK (registry,
rules, detectors, or exit-code helpers). Version numbers live in each
repository's `go.mod`.

## Metrics

Tracked snapshots; update alongside the refresh procedure.

**2026-10-11:**

- Direct consumers (go.mod `require` + code import): **6** (table above).
- Ecosystem `go-finding` version distribution: of 95 `go.mod` files referencing
  `go-finding` across `~/projects` — 60 on the current release, 31 one minor
  behind, 4 scattered on older minors. The SDK itself tracks the current
  release (see `go.mod`).
- pkg.go.dev "Imported by" for this module: **3** public importers shown
  (`cmdguard/lint`, `go-humanize-linter`, `samber-linter/internal/driver`) —
  the rest of the adoption is private or unindexed; the go.mod scan is the
  more complete signal.

## Refresh procedure

1. For each row above: `grep -rn "go-linter-sdk" <repo> --include="go.mod"`.
2. Re-run the converter check: search the consumer for `type (Issue|Violation)`
   structs and `toFinding`-style translation functions; a finding is only a
   problem if a _data_ converter reappears (rule-level adapters documented as
   deliberate, like `go-structure-linter` ADR-003, are acceptable).
3. Update the table and this header's verification date.
4. Record new consumers via the module proxy "Imported by" count on
   pkg.go.dev plus the sibling-repo scan used in the 2026-10-11 session.
