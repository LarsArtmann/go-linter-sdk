# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to `TODO_LIST.md`.

## Themes

### 1. API surface maturation

The core (`Rule` / `Registry` / `Detector` / `RuleError`) is stable and
intentionally minimal. The next tier of capabilities is about ergonomics for
consumers building real linters on top of the SDK.

Raw ideas:

- A `Filter` type for severity/category-based finding filtering
- Severity-tiered exit codes (`ExitCodeFromReport` is binary today)
- ~~`Registry.RunParallel` — rules are independent; the `Register` mutex does not
  block parallel `Check`~~ — superseded by `DetectorsFromRegistry` (pipeline
  handles per-detector parallelism, done session 7)
- Registry helpers: `NewRegistryFromRules`, `Deregister`, `Has`, `Get`
- ~~`RuleMeta.Validate` and `Category.All` — fail fast on invalid identity at
  registration rather than at first run~~ — partially done: empty-ID panics at
  registration (session 7). Full `Validate` (category non-empty, description
  non-empty) is still a raw idea.
- A typed `RuleSet` wrapper around `[]Rule` for consumers that don't want a
  mutex'd registry
- `ExitCodeFromFindings([]Finding)` convenience that skips building a `Report`
- Evaluate whether `RuleFunc` should use generics for type-safe rule definitions

### 2. Consumer adoption (the reason this SDK exists)

The SDK exists to delete converter code: `branching-flow` (1,871 LOC) and
`erraudit` (1,214 LOC) each maintain a bridge package purely because
their native domain types predate `finding.Finding`. The value proposition is
unproven until a real linter migrates.

Raw ideas:

- Pilot-port one rule from `go-structure-linter` (cleanest existing pattern) to
  prove the `Rule → finding.Finding` path end-to-end
- Pilot-port a rule from `branching-flow` to validate the converter-deletion
  claim at scale
- A `cmd/` directory with a CLI binary wrapping the registry (the README
  promises a 5-line linter; no `cmd/` exists yet)
- An `examples/` directory with a minimal consumer linter

### 3. Publication & distribution

The package is not yet consumable externally: `go.mod` uses a pseudo-version +
local `replace` directive pointing at a sibling `go-finding` checkout, and the
README's pkg.go.dev badge does not resolve to a real published package.

Raw ideas:

- Tag `go-finding` so the `replace` directive can drop and external consumers
  can resolve the dependency
- pkg.go.dev publication readiness (audit the README's claimed API surface
  against the actual exported symbols)
- A `go.work` workspace to formalize the sibling-checkout requirement during
  development
- `self`-based flake versioning (matching `go-finding`'s
  `version = self.rev or self.dirtyRev or "dev"`) — only if a binary emerges;
  today the SDK is library-only and `self` is unused

### 4. Ecosystem & CI parity

All verification is currently local. `go-finding` has CI and richer flake
outputs; this repo should match where it earns its place.

Raw ideas:

- Flake `checks` derivations for `go test`, `go vet`, and `golangci-lint` (only
  treefmt is a check derivation today)
- Additional nix apps: `watch` (live test re-runs), `tidy` (`go mod tidy`),
  `deps-update`
- `direnv` setup and/or pre-commit hooks (`pre-commit-hooks.nix`)
- Renovate / Dependabot for nix + go dependencies
- Review `devShells.ci` — confirm it has everything CI needs and nothing extra
- `meta.position` on apps and `flake-schemas` for richer `nix flake show`
  (the CI workflow itself is tracked in `TODO_LIST.md`)

### 5. Documentation depth

The SDK introduces named concepts that deserve a glossary and runnable examples
before the surface grows.

Raw ideas:

- ~~`docs/DOMAIN_LANGUAGE.md` for `Rule`, `RuleFunc`, `RuleMeta`, `Registry`,
  `Category`, `RuleError`~~ — done (session 6)
- Package-level examples (`ExampleRegistry_Run`, `ExampleRuleFunc`) visible on
  pkg.go.dev
- A diagram of the `DetectorFromRegistry → finding.Detector → BuildFlow DAG`
  adaptation

### 6. Quality hardening

Raw ideas:

- Fuzz `NewRuleError` with a nil cause (does `.Error()` panic on `%v` of nil?)
- Tests for `errors.Is(ruleErr, context.Canceled)` /
  `context.DeadlineExceeded` propagation
- Decide `Registry.Run` failure policy: fail-fast (today) vs. continue and
  return partial results — make it explicit or configurable
- Confirm `Registry.Register`'s panic-on-duplicate is the right contract for a
  library (panics in libraries are sometimes controversial)

## Open questions

Unresolved questions routed from the status reports. These are blockers for
decisions, not tasks — they need an answer before work can proceed.

- **Q1 — Is `go-finding v1.2.0` a real published tag, or is the pseudo-version +
  `replace ../go-finding` directive the intended long-term state?** Determines
  whether to push for a `go-finding` release (Theme 3), add a `go.work`, or just
  document the sibling-checkout constraint harder. (sessions 1 & 2)
- **Q2 — Should the package live at the repo root or under a sub-path?**
  `go-structure-linter` flags root-level package files (`registry.go`,
  `rule.go`, `errors.go`) and suggests `internal/`. For an importable library
  SDK the root import path `github.com/larsartmann/go-linter-sdk` may be
  intentional; moving would change it. (session 3, §g.2)
- **Q3 — Library-only, or eventual CLI?** If the SDK stays library-only, the
  flake's `self` arg is genuinely unused and `self`-based versioning is not
  needed. A future `cmd/` binary would flip both. (session 3, §g.3)

## Resolved questions

Decisions made in prior sessions, kept for context. These no longer block work.

- **Q4 — Was the `flake.lock` bump (`21b7b36`) that exposed the missing-`self`
  bug intentional?** Moot. The fix (declaring `self` in the `outputs`
  destructure) is permanent and verified across all sibling repos. The root
  cause was Nix 2.34.8+ enforcing strict `@`-pattern argument checking; which
  exact nixpkgs revision introduced it no longer matters. (resolved session 3,
  `c13366c`)
- **Q5 — White-box (`package linter`) or black-box (`package linter_test`)
  tests?** **Black-box.** Tests were moved to `package linter_test`, satisfying
  `testpackage` natively with no suppression. An `export_test.go` shim is the
  documented escape hatch if a future test needs an unexported symbol.
  (resolved session 5, `b369bba`)

## Non-goals

Things we are deliberately NOT pursuing and why:

- **A binary release / `.goreleaser.yml`:** this is a library first; revisit only
  if a `cmd/` CLI ships (Theme 2).
- **Tightening `gomoddirectives` beyond `replace-local`:** the package is
  pre-v1; the replace directive is the current source of truth.
- **Dual MIT/Apache licensing:** stay MIT, matching `go-finding`.
- **Splitting into sub-modules:** the SDK is ~300 LOC of cohesive library code;
  premature modularization would add overhead without benefit.
