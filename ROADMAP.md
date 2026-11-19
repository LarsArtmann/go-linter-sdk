# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to `TODO_LIST.md`.

## Themes

### 1. API surface maturation

The core (`Rule` / `Registry` / `Detector` / `RuleError`) is stable and
intentionally minimal. The next tier of capabilities is about ergonomics for
consumers building real linters on top of the SDK.

Shipped from this theme (2026-11-19, v0.5.0 and master): `CoverageReporter`
coverage channel, `DetectorV2` adapters, `Relativize`-normalized IDs,
`ErrInvalidCategory`, the `Filter` finding type (severity/category), the
`NewRegistryFromRules` bulk constructor, `AllCategories` (the `Category.All`
idea), `ExitCodeFromFindings` convenience, the O(1) Get/Has/Deregister
lookup story (O(1) `Get`/`Has` ID index; `Deregister` O(n) by design), the
immutable `RuleSet` (`ruleset.go`), and the fifth example
(`examples/cli-triage`). Declined with rationale: `RuleFunc` generics
(`docs/decisions/2026-11-19_rulefunc-generics-evaluation.md`) and SDK-level
fix application (ADR 0002, `docs/decisions/2026-11-19_rule-fix-representation.md`).

Raw ideas:

- Rule-level fix DECLARATION example (ADR 0002's GO half, TODO_LIST #28):
  `WithEdits` + `FixStrategyDirect` in the oops example, README host-side
  application pointers
- CoverageReporter adoption wave at consumers (TODO_LIST #29)

### 2. Consumer adoption (the reason this SDK exists)

The SDK exists to delete converter code. The core claim is now **verified at
the source level (2026-10-11)**: all six internal consumers import the SDK,
none retains a custom Issue/Violation converter layer — one deleted its
violation hierarchy (~4,737 LOC), three were born native, two keep documented
deliberate adapters. Per-repo evidence and the refresh procedure live in
`docs/ADOPTION.md`.

Raw ideas:

- ~~Full migration of `go-structure-linter` to native `linter.RuleFunc` rules~~
  RETIRED 2026-11-19: `go-structure-linter`'s own ADR-003 makes its rule-level
  adapter permanent — this idea proposed overriding a decision that repo
  already made; the converter-deletion claim it served is already verified at
  source level without it (ADOPTION 2026-10-11)
- Spread the SDK to the ~31 repos still on older `go-finding` releases
- When `go-compliance` core tags its first release (v0.1) and a standalone
  (non-BuildFlow) linter wants fleet visibility, add a short docs note:
  `compliance.Record` with
  `GateFindings{"lint": len(report.Findings)}` derived from `ExitCodeFromReport`
  — docs-only, never a dependency (from the 2026-11-19 libs review,
  `docs/review/libs/go-compliance.md`)
- A `cmd/` directory with a production CLI binary wrapping the registry
  (`--enable`/`--disable` flag parsing, `--format` text/JSON/SARIF) — see the
  resolved Q3 for why this is parked

### 3. Publication & distribution

The repo and both dependencies (`go-finding`, `go-error-family`) are public
(since 2026-09-08): `go get` needs no authentication, CI is auth-free, and
pkg.go.dev indexes the module (v0.3.0) with docs, examples, and the full API
surface. Dependency versions live only in `go.mod` — prose never duplicates
them.

Raw ideas:

- Cut the next tag promptly after public-facing doc fixes — pkg.go.dev freezes
  the tagged README, so a stale tag keeps showing stale instructions (v0.3.0
  still renders the pre-flip "private dependency" warning there)
- `self`-based flake versioning (matching `go-finding`'s
  `version = self.rev or self.dirtyRev or "dev"`) — only if a binary emerges;
  today the SDK is library-only and `self` is unused

### 4. Ecosystem & CI parity

CI runs on GitHub Actions (test, lint, fmt, govulncheck, nix flake check).
Local verification uses the flake apps and BuildFlow. There is room to deepen
the nix integration and add automated dependency management.

Raw ideas:

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
  **Decided 2026-10-11: root stays.** The import path
  `github.com/larsartmann/go-linter-sdk` IS the product for a library SDK;
  moving to `internal/` would break every consumer for zero benefit. The
  `go-structure-linter` file-layout suggestion is accepted as a permanent
  warning, not a defect. (session 3, resolved by the adoption-verification
  session)
- **Q3 — Library-only, or eventual CLI?** **Decided 2026-10-11: library-only
  stands.** The SDK deliberately stops at findings/report/exit codes; the CLI
  shell for consumer linters is `cmdguard`'s job (layering decision recorded in
  AGENTS.md, 2026-09-17). A `cmd/` here would duplicate that and drag a
  cobra-class dependency tree into a single-dep library. The flake's `self`
  argument stays unused and `self`-based versioning stays unneeded. (session
  3, resolved by the adoption-verification session)
- **Q7 — Dedicated demo repository?** A separate `go-linter-sdk-demo` repo
  would prove the stranger experience (fresh `go get`, no LarsArtmann git
  config) even more strongly than the in-CI stranger-test job added
  2026-09-09 — at the cost of a second repo to keep on the latest release.
  **Recommendation: defer.** The CI stranger-test job covers the critical path
  (clean dir + proxy-only resolve + build + run); a demo repo earns its keep
  once there are external users asking for one. (2026-09-09, plan task C5)

## Resolved questions

Decisions made in prior sessions, kept for context. These no longer block work.

- **Q6 — What is the support posture of the public repo?** **Decided
  2026-09-09: open.** Issues and PRs from strangers are accepted,
  best-effort triage, no guarantees. `SUPPORT.md`, `SECURITY.md`, and issue
  templates reflect this. Branch protection is deliberately NOT enabled:
  the auto-commit daemon pushes directly to master, and required-status
  checks would wedge it.

- **Q1 — Is `go-finding` a real published tag?** **Yes.** The dependency is a
  real published tag, confirmed by the `require` entry in `go.mod` (version:
  see `go.mod` — prose does not duplicate it). There is no local `replace`
  directive. Consumers and CI both fetch it from the module proxy.
  (resolved session 8; wording made drift-proof 2026-10-11)
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
