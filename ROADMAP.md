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
- `NewRegistryFromRules(rules []Rule) *Registry` constructor for pre-built slices
- `Category.All()` — return all built-in category values
- A typed `RuleSet` wrapper around `[]Rule` for consumers that don't want a
  mutex'd registry
- `ExitCodeFromFindings([]Finding) int` convenience that skips building a Report
- `map[string]int` index alongside the rules slice for O(1) Get/Has/Deregister
  (today all three are O(n) linear scans; fine for <100 rules, quadratic if a
  consumer calls Has in a hot loop at scale)
- Evaluate whether `RuleFunc` should use generics for type-safe rule definitions

### 2. Consumer adoption (the reason this SDK exists)

The SDK exists to delete converter code: `branching-flow` (1,871 LOC) and
`erraudit` (1,214 LOC) each maintain a bridge package purely because their
native domain types predate `finding.Finding`. The value proposition is
unproven until a real linter migrates beyond a single pilot rule.

Raw ideas:

- Pilot-port a rule from `branching-flow` to validate the converter-deletion
  claim at scale
- Pilot-port a rule from `erraudit`
- Full migration of `go-structure-linter` to `go-linter-sdk` (it already
  aliases `Issue = finding.Finding`)
- A `cmd/` directory with a production CLI binary wrapping the registry
  (`--enable`/`--disable` flag parsing, `--format` text/JSON/SARIF)

### 3. Publication & distribution

`go-finding` v1.4.1 is a real published tag resolved via VCS auth (`GOPRIVATE`).
There is no local `replace` directive. The package is externally resolvable
(modulo private-repo authentication).

Raw ideas:

- Verify whether the pkg.go.dev badge resolves now that `go-finding` v1.4.1 is
  published; if not, audit what blocks publication
- pkg.go.dev publication readiness: audit the README's claimed API surface
  against the actual exported symbols
- A `go.work` workspace to formalize sibling-checkout development against an
  uncommitted `go-finding` (neither workspace nor replace is committed)
- `self`-based flake versioning (matching `go-finding`'s
  `version = self.rev or self.dirtyRev or "dev"`) — only if a binary emerges;
  today the SDK is library-only and `self` is unused

### 4. Ecosystem & CI parity

CI runs on GitHub Actions (test, lint, fmt, govulncheck, nix flake check).
Local verification uses the flake apps and BuildFlow. There is room to deepen
the nix integration and add automated dependency management.

Raw ideas:

- Flake `checks` derivations for `go test`, `go vet`, and `golangci-lint` (only
  treefmt is a check derivation today)
- Additional nix apps: `watch` (live test re-runs), `tidy` (`go mod tidy`),
  `deps-update`
- `direnv` setup and/or pre-commit hooks (`pre-commit-hooks.nix`)
- Renovate / Dependabot for nix + go dependencies
- Review `devShells.ci` — confirm it has everything CI needs and nothing extra
- `meta.position` on apps and `flake-schemas` for richer `nix flake show`

### 5. Quality hardening

Raw ideas:

- Confirm `Registry.Register`'s panic-on-duplicate is the right contract for a
  library (panics in libraries are sometimes controversial; once consumers
  exist, this cannot be reversed without a breaking change)

## Open questions

Unresolved questions routed from the status reports. These are blockers for
decisions, not tasks — they need an answer before work can proceed.

- **Q2 — Should the package live at the repo root or under a sub-path?**
  `go-structure-linter` flags root-level package files (`registry.go`,
  `rule.go`, `errors.go`) and suggests `internal/`. For an importable library
  SDK the root import path `github.com/larsartmann/go-linter-sdk` may be
  intentional; moving would change it. (session 3)
- **Q3 — Library-only, or eventual CLI?** If the SDK stays library-only, the
  flake's `self` arg is genuinely unused and `self`-based versioning is not
  needed. A future `cmd/` binary would flip both. (session 3)

## Resolved questions

Decisions made in prior sessions, kept for context. These no longer block work.

- **Q1 — Is `go-finding` a real published tag?** **Yes.** `go-finding v1.4.1`
  is a real published tag, confirmed in `go.mod` (`require
  github.com/larsartmann/go-finding v1.4.1`). There is no local `replace`
  directive. Consumers and CI both fetch `v1.4.1` directly via VCS auth.
  (resolved session 8)
- **Q4 — Was the `flake.lock` bump that exposed the missing-`self` bug
  intentional?** Moot. The fix (declaring `self` in the `outputs` destructure)
  is permanent and verified across all sibling repos. The root cause was Nix
  2.34.8+ enforcing strict `@`-pattern argument checking. (resolved session 3)
- **Q5 — White-box (`package linter`) or black-box (`package linter_test`)
  tests?** **Black-box.** Tests were moved to `package linter_test`, satisfying
  `testpackage` natively with no suppression. An `export_test.go` shim is the
  documented escape hatch if a future test needs an unexported symbol.
  (resolved session 5)

## Non-goals

Things we are deliberately NOT pursuing and why:

- **A binary release / `.goreleaser.yml`:** this is a library first; revisit only
  if a `cmd/` CLI ships.
- **Dual MIT/Apache licensing:** stay MIT, matching `go-finding`.
- **Splitting into sub-modules:** the SDK is cohesive library code; premature
  modularization would add overhead without benefit.
