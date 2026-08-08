# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### Core types

- Shared linter scaffolding: `Rule` interface, `Registry`, `RuleFunc` adapter,
  `RuleMeta` declarative identity, `Category` open string type (8 recommended
  values). Rules emit `finding.Finding` directly — no converter layer.
- `RuleError` type (`errors.go`) wrapping rule failures with the offending
  rule's stable ID, plus `ErrRuleFailed` sentinel and `NewRuleError` constructor.
  Callers recover the failing rule via `errors.AsType[*RuleError]` (Go 1.26+) or
  check `errors.Is(err, ErrRuleFailed)`.
- `DetectorFromRegistry(registry, toolName) finding.Detector` — adapts a
  registry to a single `finding.Detector` for BuildFlow DAG integration.
  Reads working dir via `finding.WorkingDirFromContext`.
- `DetectorsFromRegistry(registry) []finding.Detector` — returns one
  `finding.Detector` per registered rule for `go-finding/pipeline` integration
  with per-rule parallelism, timeouts, error isolation, and metrics. Each
  detector is named after the rule's ID.
- `ExitCodeFromReport(report) int` — binary exit code: 0 if clean, 1 if any
  findings. The ecosystem convention.
- `IsEnabledByDefault() bool` added to the `Rule` interface. `RuleFunc`
  returns `true` by default; use `OptIn(rf)` to create a disabled-by-default
  rule that only runs when a consumer explicitly enables it.
- `OptIn(rf RuleFunc) Rule` constructor for opt-in rules (noisy, experimental,
  or domain-specific).

#### Registry API

- `Registry.Get(id string) (Rule, bool)` — lookup by stable ID.
- `Registry.Has(id string) bool` — companion to `Get`.
- `Registry.Deregister(id string) bool` — runtime rule removal (plugin
  scenarios); returns true if found.
- `RuleMeta.Validate() error` — checks ID, Name, Description, and Cat are
  non-empty. Available for use during rule construction before the registry.
- `RunOption` functional-option type + `ContinueOnError()` option for
  `Registry.Run`. Default is fail-fast (unchanged); `ContinueOnError()` runs
  all rules regardless of individual failures, collects partial findings, and
  joins errors via `errors.Join`.

#### Testing & examples

- `TestRegistry_ConcurrentReadWrite` stress test exercising the registry's
  `RWMutex` under `-race` (writers + readers + runners).
- `BenchmarkRegistry_Register`, `BenchmarkRegistry_All`, and
  `BenchmarkRegistry_Run` — the first performance baseline for the registry
  hot paths.
- Testable examples (`ExampleRegistry_Run`, `ExampleDetectorsFromRegistry`,
  `ExampleOptIn`, `ExampleRegistry_Run_continueOnError`) — visible on
  pkg.go.dev, verified by `go test`.
- `examples/minimal-linter/` — a minimal consumer linter proving the full
  Rule -> `finding.Finding` -> `Registry.Run` -> `ExitCodeFromReport` path.
- `examples/no-go-mod/` — pilot port of `go-structure-linter`'s `NoGoModRule`,
  rewritten with `go-linter-sdk`. Validates the converter-deletion claim: a
  rule emits `finding.Finding` directly via `finding.NewBuilder(...)`, with no
  intermediate Violation/Issue type.

#### Infrastructure

- Nix flake (`flake.nix`) mirroring the `go-finding` toolchain: pins Go 1.26,
  sets `GOEXPERIMENT=jsonv2`, exposes `test`, `test-race`, `bench`, `build`,
  `vet`, `lint`, `coverage`, and `clean` apps, and wires treefmt
  (gofumpt, goimports, golines@120, nixfmt).
- BuildFlow configuration (`.buildflow.yml`) and a buildflow-managed
  `.gitignore` block.
- golangci-lint v2 configuration (`.golangci.yml`) with the
  `goexperiment.jsonv2` build tag, matching the ecosystem linter set.
- GitHub Actions CI workflow (`.github/workflows/ci.yml`): test
  (ubuntu-latest + macos-latest), lint, format check, govulncheck, and nix
  flake check. Configures VCS auth for private modules (`GOPRIVATE`);
  all actions pinned to commit SHAs.
- Project metadata: `LICENSE` (MIT), `CONTRIBUTING.md`, `AGENTS.md`,
  `.gitattributes`, durable `reports/.gitkeep`.
- `.editorconfig` enforcing UTF-8/LF, tabs for Go/Makefile, 2-space for
  YAML/JSON/Nix/TOML.

#### Documentation

- Living project docs: `FEATURES.md`, `TODO_LIST.md`, `ROADMAP.md`,
  `CHANGELOG.md` — honest feature inventory, short-term work backlog,
  long-term vision, and change log.
- `docs/DOMAIN_LANGUAGE.md` — ubiquitous-language glossary defining the six
  core concepts (`Rule`, `RuleFunc`, `RuleMeta`, `Category`, `Registry`,
  `RuleError`), their relationships, and what is deliberately NOT in the
  domain.
- README "Building findings with Confidence and FixStrategy" section showing
  the `finding.NewBuilder(...)` fluent API and the Confidence/FixStrategy
  value tables.
- README "Two execution paths" section with a mermaid diagram comparing
  `Registry.Run` (sequential, fail-fast) vs `DetectorsFromRegistry` (parallel,
  per-rule isolation).
- README private-dependency note warning that `go-finding` requires
  `GOPRIVATE=github.com/larsartmann/*` + auth, with CI vs. local dev guidance.

### Changed

#### Breaking changes

- **Breaking:** `Rule` interface now requires `ID() string`. Every rule must
  declare a stable ID that never changes once published — used for registry
  deduplication, suppression matching, filter config, and the
  `finding.RuleName` field on emitted findings. `Name()` remains as the
  mutable display name. Custom `Rule` implementations must add the method;
  `RuleFunc` users must set `RuleMeta.ID`.
- **Breaking:** `RuleMeta` now has a required `ID` field (first field).
  `Register` panics on empty IDs.
- **Breaking:** `RuleError.RuleName` renamed to `RuleError.RuleID`. Callers
  accessing the field directly must update; `errors.AsType` / `errors.Is` users
  are unaffected.
- **Breaking:** `Registry.Register` now deduplicates on `ID()` instead of
  `Name()`, and panics on empty IDs. Two rules may share a display `Name()`
  as long as their `ID()` values differ.
- **Breaking:** `NewRuleError` parameter renamed from `ruleName` to `ruleID`.

#### Non-breaking changes

- `Registry.Register` now panics on any empty identity field (ID, Name,
  Description, Category), not just empty ID. Existing tests updated to provide
  complete `RuleMeta`.
- Pinned `go-finding` dependency to `v1.4.1` (published tag). The local
  `replace ../go-finding` directive has been removed — consumers and CI both
  fetch `v1.4.1` directly via VCS auth.
- `RuleFunc.Check` and `Registry.Run` / `DetectorFromRegistry` /
  `DetectorsFromRegistry` now wrap rule execution errors into `*RuleError` at
  a single chokepoint each, with a guard against double-wrapping custom `Rule`
  implementations (`wrapRuleError` passes through if already a `*RuleError`).
- `errors.AsType` migration completed: `registry.go`, `registry_test.go`, and
  the `errors.go` doc example all use Go 1.26's generic
  `errors.AsType[*RuleError]`. The project is gopls-clean (zero
  `errorsastype` hints).
- Tests moved to black-box `package linter_test`, satisfying `testpackage`
  natively with no suppression. Test factories (`makeRule`, `failingRule`)
  return the concrete `RuleFunc` rather than the `Rule` interface, so `ireturn`
  has nothing to flag either.
- `TestRegistry_Run_NoDoubleWrap` restored to a positive exact-identity
  assertion (`cause == errSentinel` proves a single wrap), stronger than the
  previous negative-only check.
- `.golangci.yml` right-sized from first principles: removed the cargo-culted
  `mnd` numbers/functions and `gosec` excludes (zero findings without them —
  this is a pure library), and trimmed `varnamelen` ignore-names from 30+ to
  the 6 conventional abbreviations that actually appear. Added `examples/`
  path exclusion for `depguard`, `forbidigo`, and `mnd` (example CLI code
  legitimately prints and uses exit codes); `tt` added to `varnamelen`
  ignore-names.
- `DetectorFromRegistry(r *Registry, ...)` parameter renamed to `registry` for
  clearer public-API documentation and to clear a `varnamelen` finding without
  suppression.
- `ireturn` allow-list now includes `github.com/larsartmann/go-linter-sdk.Rule`
  for `OptIn`, which intentionally returns the `Rule` interface.
- `Category` documented as an open `string` type. The 8 built-in constants are
  recommendations; consumers can define custom categories
  (e.g., `Category("api")`).
- Documentation one-home rule established: `FEATURES.md` tracks only
  capabilities that have code today; all not-yet-built capabilities live
  exclusively in `ROADMAP.md`. Removed the duplicated "Planned capabilities"
  section from `FEATURES.md`.
- ROADMAP open questions Q1 (go-finding tag — resolved: v1.4.1 is real), Q4
  (flake.lock bump — moot), and Q5 (white-box vs black-box tests — resolved:
  black-box) closed and moved to "Resolved questions."

### Fixed

- Build and lint failures caused by `encoding/json/v2` and
  `encoding/json/jsontext` (used transitively via `go-finding`) being excluded
  by build constraints without `GOEXPERIMENT=jsonv2`. The flake devShells and
  apps now export it durably.
- `erraudit` findings: bare `return nil, err` paths in `registry.go`
  and `rule.go` replaced with typed, rule-attributed errors.
- BuildFlow failure `function 'outputs' called with unexpected argument 'self'`:
  added `self` to the flake `outputs` destructure pattern. Nix 2.34.8
  enforces strict argument checking on `@`-patterns; the sibling `go-finding`
  flake already declared it.
- Coverage output path mismatch: the `coverage` and `clean` flake apps now
  write/read `reports/coverage.out`, aligning with `AGENTS.md`, the `.gitignore`
  convention, and BuildFlow's `test-coverage` step.
- README "5-line linter" example did not compile: the import block was missing
  `"os"` (needed for `os.Exit`). Added the import; example verified to compile
  against the actual exported API.

### Documentation

- Converted all source line-number references in `DOMAIN_LANGUAGE.md` and
  `FEATURES.md` to durable symbol-name references (`rule.go — type Rule`
  instead of `rule.go:89`). Line numbers rot on every edit; symbol names don't.
- `DOMAIN_LANGUAGE.md` corrected: RuleMeta validation now checks all four
  identity fields (ID, Name, Description, Category), not just empty ID.
- `README.md` example updated to show distinct `ID` and `Name` values,
  demonstrating the dual-identity contract. "5-line linter" headline replaced
  with "A minimal linter" (the required `ID` field made the old count
  misleading). Version requirement updated from `v1.2+` to `v1.4+`.
- `CONTRIBUTING.md` updated: removed stale references to the local `replace`
  directive and sibling-checkout requirement (`go-finding` v1.4.1 is published
  and resolved via VCS auth).
- `DetectorFromRegistry` doc comment now cross-references
  `DetectorsFromRegistry` (when to use which execution path).
- `ROADMAP.md` reconciled with codebase: removed 8+ stale items that were
  already shipped, fixed the false "pseudo-version + replace directive" claim
  in the Publication theme, resolved Q1 (go-finding v1.4.1 confirmed real).
- cqrs-lint feedback document moved to `docs/feedback/processed/` and annotated
  with a processing note pointing to Appendix A (maintainer response).
- `TODO_LIST.md` rebuilt with 5 actionable items harvested from the two most
  recent status reports (`docs/status/2026-08-05_*`).
