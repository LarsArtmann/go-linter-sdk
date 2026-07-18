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

### Changed

- `RuleFunc.Check` and `Registry.Run` / `DetectorFromRegistry` now wrap errors
  with `RuleError` at a single chokepoint each, with a guard against
  double-wrapping custom `Rule` implementations.
- `go.mod` points `go-finding` at a local replace directive (sibling checkout);
  consumers need `github.com/larsartmann/go-finding` available on the module
  path until a tagged release is published.

### Fixed

- Build and lint failures caused by `encoding/json/v2` and
  `encoding/json/jsontext` (used transitively via `go-finding`) being excluded
  by build constraints without `GOEXPERIMENT=jsonv2`. The flake devShells and
  apps now export it durably.
- `hierarchical-errors` findings: bare `return nil, err` paths in `registry.go`
  and `rule.go` replaced with typed, rule-attributed errors.
