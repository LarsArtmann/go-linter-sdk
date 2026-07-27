# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
  flake check. Handles the `replace ../go-finding` directive by cloning
  `go-finding` as a sibling; all actions pinned to commit SHAs.
- `docs/DOMAIN_LANGUAGE.md` — ubiquitous-language glossary defining the six
  core concepts (`Rule`, `RuleFunc`, `RuleMeta`, `Category`, `Registry`,
  `RuleError`), their relationships, and what is deliberately NOT in the
  domain.

### Changed

- `RuleFunc.Check` and `Registry.Run` / `DetectorFromRegistry` now wrap errors
  with `RuleError` at a single chokepoint each, with a guard against
  double-wrapping custom `Rule` implementations.
- `go.mod` points `go-finding` at a local replace directive (sibling checkout);
  consumers need `github.com/larsartmann/go-finding` available on the module
  path until a tagged release is published.
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
- `hierarchical-errors` findings: bare `return nil, err` paths in `registry.go`
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
