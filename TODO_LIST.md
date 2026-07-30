# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open work

### API maturation

| # | Status | Task | Notes |
|---|--------|------|-------|
| 1 | 🔴 `TODO` | Add `Registry.Get(id string) (Rule, bool)` | With `ID()` now the canonical key, consumers need lookup by ID |
| 2 | 🔴 `TODO` | Add `Registry.Has(id string) bool` | Companion to `Get` |
| 3 | 🔴 `TODO` | Add `Registry.Deregister(id string)` | Enable runtime rule removal (plugin scenarios) |
| 4 | 🔴 `TODO` | Implement `RuleMeta.Validate()` | Empty-ID panics at `Register` (done); add validation for empty Name, Description, Category |
| 5 | 🔴 `TODO` | Make `Registry.Run` failure policy configurable | Currently fail-fast; `DetectorsFromRegistry` → pipeline degrades gracefully. Add continue-and-report option |
| 6 | 🔴 `TODO` | Add testable examples | `ExampleRegistry_Run`, `ExampleDetectorsFromRegistry`, `ExampleOptIn` |

### Documentation

| # | Status | Task | Notes |
|---|--------|------|-------|
| 7 | 🔴 `TODO` | Add "Building Findings" section to README | Show `.WithConfidence()` and `.WithFixStrategy()` via `finding.NewBuilder(...)` |
| 8 | 🔴 `TODO` | Add two-execution-paths diagram to README | `Registry.Run` (sequential, fail-fast) vs `DetectorsFromRegistry` → pipeline (parallel, graceful) |

### Consumer adoption

| # | Status | Task | Notes |
|---|--------|------|-------|
| 9 | 🔴 `TODO` | Create `examples/` directory | Minimal consumer linter proving the Rule → `finding.Finding` path |
| 10 | 🔴 `TODO` | Pilot-port one rule from `go-structure-linter` | Validate the converter-deletion claim with real code |

---

Longer-horizon work lives in `ROADMAP.md`. When a ROADMAP idea is refined into
a bounded task, it moves here.

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done (dropped here: LICENSE, .golangci.yml,
    CONTRIBUTING.md, reports/.gitkeep, the `vet` app, .editorconfig, the flake
    `self` arg, and several tests).
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
-->
