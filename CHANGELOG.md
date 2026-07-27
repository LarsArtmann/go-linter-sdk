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

### Changed

- `RuleFunc.Check` and `Registry.Run` / `DetectorFromRegistry` now wrap errors
  with `RuleError` at a single chokepoint each, with a guard against
  double-wrapping custom `Rule` implementations.
- `go.mod` points `go-finding` at a local replace directive (sibling checkout);
  consumers need `github.com/larsartmann/go-finding` available on the module
  path until a tagged release is published.
- `registry.go` migrated to Go 1.26's generic `errors.AsType[*RuleError]`
  for the no-double-wrap guard (`88bf523`). `registry_test.go` and the
  `errors.go` doc example still use `errors.As`; completion is tracked in
  `TODO_LIST.md`.

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
