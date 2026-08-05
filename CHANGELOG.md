# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `Registry.Get(id string) (Rule, bool)` — lookup by stable ID.
- `Registry.Has(id string) bool` — companion to `Get`.
- `Registry.Deregister(id string) bool` — runtime rule removal (plugin
  scenarios); returns true if found.
- `RuleMeta.Validate() error` — checks ID, Name, Description, and Cat are
  non-empty. `Register` now calls this through the `Rule` interface and panics
  on any empty identity field (not just empty ID).
- `RunOption` functional-option type + `ContinueOnError()` option for
  `Registry.Run`. Default is fail-fast (unchanged); `ContinueOnError()` runs
  all rules regardless of individual failures, collects partial findings, and
  joins errors via `errors.Join`.
- Testable examples (`ExampleRegistry_Run`, `ExampleDetectorsFromRegistry`,
  `ExampleOptIn`, `ExampleRegistry_Run_continueOnError`) — visible on
  pkg.go.dev, verified by `go test`.
- `examples/minimal-linter/` — a minimal consumer linter proving the full
  Rule → `finding.Finding` → `Registry.Run` → `ExitCodeFromReport` path.
- `examples/no-go-mod/` — pilot port of `go-structure-linter`'s `NoGoModRule`,
  rewritten with `go-linter-sdk`. Validates the converter-deletion claim: a
  rule emits `finding.Finding` directly via `finding.NewBuilder(...)`, with no
  intermediate Violation/Issue type.
- README "Building findings with Confidence and FixStrategy" section showing
  the `finding.NewBuilder(...)` fluent API and the Confidence/FixStrategy
  value tables.
- README "Two execution paths" section with a mermaid diagram comparing
  `Registry.Run` (sequential, fail-fast) vs `DetectorsFromRegistry` (parallel,
  per-rule isolation).
- `.golangci.yml`: `examples/` path exclusion for `depguard`, `forbidigo`,
  and `mnd` (example CLI code legitimately prints and uses exit codes);
  `tt` added to `varnamelen` ignore-names (standard Go testing convention).

### Changed

- `Registry.Register` now panics on any empty identity field (ID, Name,
  Description, Category), not just empty ID. Existing tests updated to provide
  complete `RuleMeta`.

- `ID() string` added to the `Rule` interface. Every rule must declare a stable
  ID that never changes once published — used for registry deduplication,
  suppression matching, filter config, and the `finding.RuleName` field on
  emitted findings. `Name()` remains as the mutable display name.
- `DetectorsFromRegistry(registry) []finding.Detector` — returns one
  `finding.Detector` per registered rule for `go-finding/pipeline` integration
  with per-rule parallelism, timeouts, error isolation, and metrics. Each
  detector is named after the rule's ID. Unlike `DetectorFromRegistry` (which
  collapses all rules into a single opaque detector), this preserves per-rule
  granularity at the pipeline layer.
- `RuleMeta.ID` field (required). `Register` panics on empty IDs.
- `RuleError.RuleID` field (renamed from `RuleName`) — carries the stable rule
  ID instead of the display name.
- `Category` documented as an open `string` type. The 8 built-in constants are
  recommendations; consumers can define custom categories
  (e.g., `Category("api")`).
- `ireturn` allow-list now includes `github.com/larsartmann/go-linter-sdk.Rule`
  for `OptIn`, which intentionally returns the `Rule` interface.

### Changed

- **Breaking:** `Rule` interface now requires `ID() string`. Custom `Rule`
  implementations must add the method. `RuleFunc` users must set `RuleMeta.ID`.
- **Breaking:** `RuleMeta` now has a required `ID` field (first field).
- **Breaking:** `RuleError.RuleName` renamed to `RuleError.RuleID`. Callers
  accessing the field directly must update; `errors.AsType` / `errors.Is` users
  are unaffected.
- **Breaking:** `Registry.Register` now deduplicates on `ID()` instead of
  `Name()`, and panics on empty IDs. Two rules may share a display `Name()`
  as long as their `ID()` values differ.
- **Breaking:** `NewRuleError` parameter renamed from `ruleName` to `ruleID`.
- `IsEnabledByDefault() bool` added to the `Rule` interface. `RuleFunc` returns
  `true` by default; use `OptIn(rf)` to create a disabled-by-default rule that
  only runs when a consumer explicitly enables it. This closes the interface gap
  with `go-structure-linter`'s existing `IsEnabledByDefault()` method, removing
  a blocker for pilot adoption.
- `OptIn(rf RuleFunc) Rule` constructor for opt-in rules.
- Pinned `go-finding` dependency to `v1.4.1` (published tag). The local
  `replace ../go-finding` directive has been removed — consumers and CI both
  fetch `v1.4.1` directly via VCS auth.

### Added

- Shared linter scaffolding: `Rule`, `Registry`, `Detector` types that eliminate
  converter duplication across LarsArtmann Go linters.
- `RuleError` type (`errors.go`) wrapping rule failures with the offending rule
  name, plus `ErrRuleFailed` sentinel and `NewRuleError` constructor, so callers
  can recover the failing rule via `errors.As`.
- Nix flake (`flake.nix`) mirroring the `go-finding` toolchain: pins Go 1.26,
  sets `GOEXPERIMENT=jsonv2`, exposes `test`, `test-race`, `bench`, `build`,
  `vet`, `lint`, `coverage`, and `clean` apps, and wires treefmt
  (gofumpt, goimports, golines@120, nixfmt).
- BuildFlow configuration (`.buildflow.yml`) and a buildflow-managed
  `.gitignore` block.
- golangci-lint v2 configuration (`.golangci.yml`) with the
  `goexperiment.jsonv2` build tag, matching the ecosystem linter set.
- Project metadata: `LICENSE` (MIT), `CONTRIBUTING.md`, `AGENTS.md`,
  `.gitattributes`, durable `reports/.gitkeep`.
- `.editorconfig` enforcing UTF-8/LF, tabs for Go/Makefile, 2-space for
  YAML/JSON/Nix/TOML (`0ca7a41`).
- Living project docs: `FEATURES.md`, `TODO_LIST.md`, `ROADMAP.md` — honest
  feature inventory, short-term work backlog, and long-term vision.
- `TestRegistry_ConcurrentReadWrite` stress test exercising the registry's
  `RWMutex` under `-race` (writers + readers + runners). Previously `-race`
  passed vacuously because no goroutines touched the registry.
- `BenchmarkRegistry_Register`, `BenchmarkRegistry_All`, and
  `BenchmarkRegistry_Run` — the first performance baseline for the registry
  hot paths.
- GitHub Actions CI workflow (`.github/workflows/ci.yml`): test
  (ubuntu-latest + macos-latest), lint, format check, govulncheck, and nix
  flake check. Configures VCS auth for private modules (`GOPRIVATE`);
  all actions pinned to commit SHAs.
- `docs/DOMAIN_LANGUAGE.md` — ubiquitous-language glossary defining the six
  core concepts (`Rule`, `RuleFunc`, `RuleMeta`, `Category`, `Registry`,
  `RuleError`), their relationships, and what is deliberately NOT in the
  domain.

### Changed

- `RuleFunc.Check` and `Registry.Run` / `DetectorFromRegistry` now wrap errors
  with `RuleError` at a single chokepoint each, with a guard against
  double-wrapping custom `Rule` implementations.
- `go.mod` resolves `go-finding` v1.4.1 via VCS auth (`GOPRIVATE`); no local
  replace directive or sibling checkout needed.
- `errors.AsType` migration completed: `registry.go`, `registry_test.go`, and
  the `errors.go` doc example all use Go 1.26's generic
  `errors.AsType[*RuleError]` (production code since `88bf523`; tests + docs
  finished this session). The project is now gopls-clean (zero `errorsastype`
  hints).
- Tests moved to black-box `package linter_test`, satisfying `testpackage`
  natively with no suppression. Test factories (`makeRule`, `failingRule`)
  return the concrete `RuleFunc` rather than the `Rule` interface, so `ireturn`
  has nothing to flag either.
- `TestRegistry_Run_NoDoubleWrap` restored to a positive exact-identity
  assertion (`cause == errSentinel` proves a single wrap), stronger than the
  previous negative-only check. `errors.Is` is deliberately avoided because it
  cannot distinguish one wrap from two.
- `.golangci.yml` right-sized from first principles: removed the cargo-culted
  `mnd` numbers/functions and `gosec` excludes (zero findings without them —
  this is a pure library), and trimmed `varnamelen` ignore-names from 30+ to
  the 6 conventional abbreviations that actually appear. The wholesale
  `_test.go` exclusions for `testpackage` and `ireturn` were dropped (no longer
  needed after the black-box migration).
- `DetectorFromRegistry(r *Registry, ...)` parameter renamed to `registry` for
  clearer public-API documentation and to clear a `varnamelen` finding without
  suppression.
- Documentation one-home rule established: `FEATURES.md` tracks only
  capabilities that have code today; all not-yet-built capabilities live
  exclusively in `ROADMAP.md`. Removed the duplicated "Planned capabilities"
  section from `FEATURES.md` (all items were already in `ROADMAP.md`).
- ROADMAP open questions Q4 (flake.lock bump — moot) and Q5 (white-box vs
  black-box tests — resolved: black-box) closed and moved to a "Resolved
  questions" section.

### Fixed

- Build and lint failures caused by `encoding/json/v2` and
  `encoding/json/jsontext` (used transitively via `go-finding`) being excluded
  by build constraints without `GOEXPERIMENT=jsonv2`. The flake devShells and
  apps now export it durably.
- `erraudit` findings: bare `return nil, err` paths in `registry.go`
  and `rule.go` replaced with typed, rule-attributed errors.
- BuildFlow failure `function 'outputs' called with unexpected argument 'self'`:
  added `self` to the flake `outputs` destructure pattern (`c13366c`). Nix 2.34.8
  enforces strict argument checking on `@`-patterns; the sibling `go-finding`
  flake already declared it.
- Coverage output path mismatch: the `coverage` and `clean` flake apps now
  write/read `reports/coverage.out`, aligning with `AGENTS.md`, the `.gitignore`
  convention, and BuildFlow's `test-coverage` step. Previously both apps used the
  repo-root `./coverage.out`, contradicting every other reference in the repo.
- README "5-line linter" example did not compile: the import block was missing
  `"os"` (needed for `os.Exit`). Added the import; example verified to compile
  against the actual exported API.
- Updated all living docs and source comments to reflect the
  `hierarchical-errors` → `erraudit` project rename. Historical status reports
  were left as-is (point-in-time records).

### Documentation

- Converted all source line-number references in `DOMAIN_LANGUAGE.md` and
  `FEATURES.md` to durable symbol-name references (`rule.go — type Rule`
  instead of `rule.go:89`). Line numbers rot on every edit; symbol names don't.
- `README.md` example updated to show distinct `ID` and `Name` values,
  demonstrating the dual-identity contract. "5-line linter" headline replaced
  with "A minimal linter" (the required `ID` field made the old count
  misleading). Version requirement updated from `v1.2+` to `v1.4+`.
- `CONTRIBUTING.md` updated: removed stale references to the local `replace`
  directive and sibling-checkout requirement (`go-finding` v1.4.1 is published
  and resolved via VCS auth).
- `DetectorFromRegistry` doc comment now cross-references
  `DetectorsFromRegistry` (when to use which execution path).
- cqrs-lint feedback document moved to `docs/feedback/processed/` and annotated
  with a processing note pointing to Appendix A (maintainer response).
- `TODO_LIST.md` rebuilt with 10 actionable items from the feedback analysis
  (registry helpers, `RuleMeta.Validate`, testable examples, consumer pilot).
