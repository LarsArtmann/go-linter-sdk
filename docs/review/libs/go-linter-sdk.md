# go-linter-sdk self-leverage review — go-linter-sdk

- **Reviewed:** 2026-11-19 by the orchestrating session (self-audit; fleet audit Group D)
- **Library:** github.com/larsartmann/go-linter-sdk — this repository IS the library
- **Sub-modules:** this repo is single-module (root `go.mod` only; `examples/` are
  packages inside the root module, exercised via `exec.Command` integration tests —
  `find . -maxdepth 2 -name go.mod` returns only `./go.mod`)
- **Version reviewed:** `v0.4.1-5-gd8f9368` (2026-11-19, `git describe --tags`);
  latest published tag `v0.4.1`; point-in-time snapshot
- **Verdict:** ADOPTED (trivially — it is the library) — **self-leverage HIGH**:
  tests, godoc examples, and CI are superb; the four `examples/` binaries lag the
  API by one generation and teach the pre-helper pattern
- **Score:** 72/100 (high) — 8 of 16 applicable capability rows fully leveraged

## What the library offers (its own surface, as of HEAD)

The complete public API (`go doc` verifiable):

| Capability | Lives at | Notes |
| --- | --- | --- |
| `Rule` interface (dual identity ID/Name) | `rule.go:95` | ID stable, Name mutable |
| `RuleFunc` adapter | `rule.go:108` | data-shaped: `RuleMeta` + `Run` closure |
| `RuleFunc.NewFinding(msg, pos)` pre-stamped builder | `rule.go:260` | kills identity boilerplate in `Run` |
| `OptIn(rf)` disabled-by-default wrapper | `rule.go:206` | noisy/experimental rules |
| `RuleMeta.Validate` + `ErrMissingFields` | `rule.go:152`, `rule.go:139` | fail-fast identity validation |
| 11 recommended `Category` values (open type) | `rule.go:45-57` | naming, documentation, best-practice included |
| `Registry` CRUD: `Register`/`All`/`Get`/`Has`/`Deregister` | `registry.go:54-147` | thread-safe, snapshot semantics mid-run |
| `NewRegistry(WithToolName(...))` | `registry.go:29-48` | auto-stamps `RuleMeta.ToolName` + report header |
| `Registry.Run` + `ContinueOnError()` | `registry.go:176-232` | fail-fast default; joined-error mode |
| `DetectorFromRegistry` (BuildFlow DAG path) | `registry.go:242` | single opaque detector |
| `DetectorsFromRegistry` (pipeline path) | `registry.go:275` | one detector per rule, named by rule ID |
| `ExitCodeFromReport` (binary) | `registry.go:306` | 0 clean / 1 findings |
| `ExitCodeByConfidence` (tiered) | `registry.go:358` | 0/1/2 — CI triage mode |
| `FilterRules(all, enable, disable)` | `registry.go:326` | standard `--enable`/`--disable` set logic |
| `RuleError` / `NewRuleError` / `RuleErrors` / `ErrRuleFailed` | `errors.go` | rule-ID error attribution, join-tree aware |

## What this repo uses of itself today (self-leverage baseline)

Four demonstration artifacts exercise the library:

1. **Black-box test suite** (`*_test.go`, `package linter_test`, 98.7% library
   coverage, race-clean): exercises every API above, including
   `TestWithToolName_AutoStampsRuleFunc` (`registry_test.go:793`),
   `TestFilterRules_*` (`registry_test.go:894-954`), concurrent
   `Deregister`-mid-run snapshot semantics, and deep `RuleErrors` trees.
2. **Runnable godoc examples** (`example_test.go`): `ExampleRuleFunc_NewFinding`
   (`example_test.go:83`), `ExampleFilterRules` (`example_test.go:116`),
   `ExampleExitCodeByConfidence` (`example_test.go:135`), `ExampleOptIn`
   (`example_test.go:152`), `ExampleRegistry_Run_continueOnError` — visible on
   pkg.go.dev.
3. **Four example binaries** (`examples/`): `minimal-linter`, `no-go-mod`,
   `naked-return-guard` (branching-flow pilot port), `oops-redundant-guard`
   (erraudit pilot port) — run via `examples_integration_test.go` (262 LOC).
4. **Stranger-test CI job** (`FEATURES.md:67`): clean-dir, proxy-only
   `go get @latest` + build + run against the published artifact, exercising
   `Registry.Run`, `ExitCodeByConfidence`, `DetectorFromRegistry`.

The ecosystem leverage is verified elsewhere: six consumers at `v0.4.x`, converter
deletion verified at source level (2026-10-11, `docs/ADOPTION.md`) — the
collective-improvement instrument works.

## Capability inventory & gap analysis (self-leverage lens)

| # | Capability | Status | Evidence (self-usage) | Impact |
| --- | --- | --- | --- | --- |
| 1 | `Rule` + `RuleFunc` | 🟢 Fully | all 4 examples; full test suite | — |
| 2 | `OptIn` | 🟡 Partial | godoc + tests (`example_test.go:152`, `registry_test.go:549`); no `examples/` binary demonstrates an opt-in rule + `--enable` flow | medium |
| 3 | `RuleFunc.NewFinding` | 🟡 Partial | godoc example only (`example_test.go:83`); **0 of 4 examples use it** — all hand-roll `finding.NewBuilder` with repeated rule ID/tool/severity args (`examples/minimal-linter/main.go:42-43`, `examples/naked-return-guard/main.go:163`, `examples/no-go-mod/main.go:56`, `examples/oops-redundant-guard/main.go:351`) | **high** |
| 4 | `RuleMeta.Validate` + `ErrMissingFields` | 🟢 Fully | tests + `Register` panic path | — |
| 5 | Registry CRUD incl. `Deregister` | 🟢 Fully | unit + bench + concurrent snapshot tests | — |
| 6 | `WithToolName` / `NewRegistry(opts)` | 🟡 Partial | tests + godoc + stranger-CI; **all 4 examples call bare `NewRegistry()`** (`examples/minimal-linter/main.go:26`, `examples/no-go-mod/main.go:40`, `examples/naked-return-guard/main.go:50`, `examples/oops-redundant-guard/main.go:56`); two hand-stamp `Meta.ToolName` instead (`examples/naked-return-guard/main.go:59`), two stamp the tool name as a builder literal (`examples/minimal-linter/main.go:43`) while the report header silently defaults to `"linter"` (`registry.go:201`) — identity inconsistency in the flagship teaching example | **high** |
| 7 | `Run` + `ContinueOnError` + `RuleErrors` | 🟢 Fully | tests + godoc example | — |
| 8 | `RuleError` attribution | 🟢 Fully | tests incl. deeply nested trees | — |
| 9 | `DetectorFromRegistry` | 🟢 Fully | godoc example + stranger-test CI job | — |
| 10 | `DetectorsFromRegistry` | 🟡 Partial | tests + godoc example; no `examples/` demo of pipeline wiring (deliberate: a demo would import `go-finding/pipeline` into the root module's `go.mod`, breaking the single-dep policy — godoc example is the right ceiling) | low |
| 11 | `ExitCodeFromReport` | 🟢 Fully | all 4 examples use it | — |
| 12 | `ExitCodeByConfidence` | 🟡 Partial | godoc + tests + stranger-CI; no `examples/` binary shows the triage exit-code flow | medium |
| 13 | `FilterRules` | 🟡 Partial | godoc + tests + README (`README.md:298`); no example CLI shows flag parsing → `FilterRules` → filtered registry | medium |
| 14 | Category taxonomy docs | 🟡 Partial | `docs/DOMAIN_LANGUAGE.md:83` said "Eight values" while `rule.go:45-57` ships eleven — **fixed during this audit** (drift found and corrected on sight); the relationship diagram still predates `NewFinding`/`WithToolName` | low |
| 15 | `examples/` as living consumer docs | 🔴 Missed | examples predate the helper generation: they teach `finding.NewBuilder(...)` + manual tool-name stamping, i.e. the exact boilerplate `NewFinding`/`WithToolName` were built (`CHANGELOG.md:271,287`) to eliminate. Highest-traffic teaching artifact = one generation stale | **high** |
| 16 | Self-verification infra (black-box tests, race, stranger-CI, coverage gate) | 🟢 Fully | `FEATURES.md:67,82`; 98.7% library coverage | — |

No N/A rows: this is a self-audit, every capability is in scope by definition.
Row 10's pipeline-demo absence is policy-blocked, so it stays Partial (godoc
example exists) rather than Missed.

## Score derivation (rubric)

- N = 16 rows, A = 16 applicable (no N/A).
- Fully Leveraged = 1: rows 1, 4, 5, 7, 8, 9, 11, 16 → **8 × 1.0**
- Partially Used = 0.5: rows 2, 3, 6, 10, 12, 13, 14 → **7 × 0.5 = 3.5**
- Missed = 0: row 15 → 0
- score = round(100 × 11.5 / 16) = **72 → high band (70-84)**

## Version currency

Pinned to itself: `v0.4.1` published (pkg.go.dev frozen README at that tag);
working tree 5 commits past it (`d8f9368`, daemon auto-commits). The Go floor is
1.27, locked in lockstep across `go.mod` / `flake.nix` / `.golangci.yml`
(enforced by `scripts/check-doc-versions.sh`).

## The fleet argument

This repository IS the owner's collective-improvement instrument for the linter
family: it deleted the per-linter converter layer (branching-flow 1,871 LOC,
erraudit 1,214 LOC class of duplication) and six repos now share rule
scaffolding. The owner's constraint — "if all my projects re-implement everything
I will only fix bugs all day long" — is answered by this SDK existing; the
remaining self-leverage debt (rows 3, 6, 15) matters precisely because `examples/`
are the adoption funnel: every consumer that copies a stale example re-imports
the boilerplate the SDK was built to remove. Fixing the examples is the cheapest
possible fleet-wide improvement: it propagates correct patterns to every future
consumer for zero API change.

## Ranked opportunities (impact × ease)

1. **Modernize all 4 examples to the current API** — `linter.NewRegistry(linter.WithToolName("..."))`
   + `rule.NewFinding(msg, pos)` instead of hand-rolled `finding.NewBuilder`;
   also fixes the report-header/findings identity mismatch in `minimal-linter`
   and `no-go-mod`. — routing: TODO_LIST row "Modernize examples/ to current API
   (WithToolName + NewFinding), fix tool-name identity mismatch".
2. **Add a fifth example demonstrating the CLI trio** — `OptIn` rule +
   `FilterRules(enable, disable)` from `os.Args` + `ExitCodeByConfidence`
   triage exit. — routing: TODO_LIST row "Add opt-in/filter/triage example CLI".
3. **Keep `docs/DOMAIN_LANGUAGE.md` in lockstep with API growth** — this audit
   found "Eight" vs eleven categories and missing terms (fixed on sight); add a
   doc-sync check to the release checklist. — routing: informational (docs-health
   process owns it; no TODO row needed for a fixed instance).
4. **Pipeline wiring demo** — rejected: would import `go-finding/pipeline` into
   the root module and break the single-dep policy (`AGENTS.md` layering rule).
   The godoc example (`ExampleDetectorsFromRegistry`) is the deliberate ceiling.

## Why the verdict

The library is superbly built for the fleet (single-dep, dual-identity rules,
verified converter elimination) and superbly verified (98.7% black-box coverage,
stranger-CI against the published artifact). Self-leverage loses points in
exactly one place: the teaching surface (`examples/`) is one generation behind
the API it teaches. That is cheap to fix and high-leverage.

## How to verify

```bash
cd /home/lars/projects/go-linter-sdk
git describe --tags            # v0.4.1-5-gd8f9368
grep -c 'NewRegistry()' examples/*/main.go          # 4 — all bare, none pass WithToolName
grep -c 'rule.NewFinding\|\.NewFinding(' examples/*/main.go  # 0 — helper unused in examples
grep -n 'NewFinding\|ExampleFilterRules\|ExampleExitCodeByConfidence\|ExampleOptIn' example_test.go | head  # godoc examples exist
grep -n 'Category\(Naming\|Documentation\|BestPractice\)' rule.go               # the 3 categories DOMAIN_LANGUAGE missed
go test ./...                  # full suite green at review HEAD
```

All commands above were executed during this audit; the counts in the comments
are the observed outputs (grep counts verified 2026-11-19).

> **Re-review trigger:** re-run this audit when the consumer pins a newer
> release of the reviewed library than the version stamped above, or when a
> follow-up task from the 2026-11-19 Pareto plan lands in this repo
> (footers standardized 2026-11-19, plan task M22/F22.1).
