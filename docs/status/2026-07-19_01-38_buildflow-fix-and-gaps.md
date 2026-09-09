# Status Report — go-linter-sdk

**Date:** 2026-07-19 01:38 CEST
**Scope:** Single-session audit of the buildflow failure fix and its aftermath
**Commit:** `18ccbd5` — "Make the SDK buildable and lint-clean under the ecosystem's jsonv2 experiment"
**Verdict:** The immediate fire is out. The house still has structural gaps.

---

## TL;DR

| Dimension        | State                                                       |
| ---------------- | ----------------------------------------------------------- |
| Buildflow        | 21/30 → **31/32 passing** (0 failures; 1 skipped by config) |
| Root cause fix   | **Durable** via `flake.nix`                                 |
| Coverage dir fix | **Fragile** — works only after `nix develop`                |
| Error design     | **Solid** but lightly tested                                |
| Project hygiene  | **Weak** — many ecosystem docs/files missing                |

---

## (a) FULLY DONE

1. **Diagnosed the root cause correctly.** The buildflow cascade traced back to a single missing environment variable: `GOEXPERIMENT=jsonv2`. `go-finding` uses `encoding/json/v2`, which is gated behind that experiment flag. `go-linter-sdk` had no way to set it.
2. **Created `flake.nix`** mirroring `go-finding`'s structure: `devShells.default` + `devShells.ci` with `GOEXPERIMENT=jsonv2`, eight `apps` (`test`, `test-race`, `bench`, `build`, `vet`, `lint`, `coverage`, `clean`), and `treefmt` config (gofumpt, goimports, golines @120, nixfmt).
3. **Created `flake.lock`** pinning the same nixpkgs / flake-parts / treefmt-nix / systems revisions as `go-finding`, so the two sibling repos agree on toolchain.
4. **Created `.buildflow.yml`** so buildflow uses explicit settings instead of defaults.
5. **Added `.gitignore`** (buildflow-managed block) covering secrets, OS files, editors, logs, buildflow artifacts, reports/, and Go build outputs.
6. **Fixed the two `hierarchical-errors` findings** by introducing `RuleError` (`errors.go`) with `RuleName`/`Cause` fields, `Error()`/`Unwrap()`/`Is()` methods, an `ErrRuleFailed` sentinel, and a `NewRuleError` constructor. `RuleFunc.Check`, `Registry.Run`, and `DetectorFromRegistry` all wrap into it.
7. **Prevented double-wrapping** via `wrapRuleError`, which uses `errors.As` to pass through an error that is already a `*RuleError`.
8. **Added 3 new tests** covering: error wrapping in `Run`, no-double-wrap invariant, and error wrapping in `DetectorFromRegistry`. All 9 tests pass.
9. **Created `AGENTS.md`** documenting the `GOEXPERIMENT=jsonv2` requirement, nix app commands, architecture, error-wrapping pattern, and the sibling-repo replace directive.
10. **Ran the full buildflow** and confirmed 31/32 passing (1 skipped via config = `gitleaks`).
11. **Verified via nix apps** that `nix run .#build` and `nix run .#test` both succeed.
12. **Committed everything** with a detailed commit message explaining root cause, fix, and verification.

---

## (b) PARTIALLY DONE

1. **The `reports/` directory fix.** I created the directory locally and added `mkdir -p reports` to the flake.nix `shellHook` (both `default` and `ci` devShells). **But:** the shellHook only fires inside `nix develop`. A fresh clone running `GOEXPERIMENT=jsonv2 buildflow` directly (no nix shell) will hit the same `open reports/coverage.out: no such file or directory` failure. The fix is not reproducible across environments.
2. **Error wrapping design.** The types exist and three happy/sad paths are tested, but edge cases are not: context-cancellation propagation through `RuleError`, nil-cause behavior, and `errors.Is` against wrapped sentinels other than `ErrRuleFailed` are untested.
3. **AGENTS.md.** Written, but it claims "No Makefile. No justfile." which is trivially true (there never were any) and adds nothing. It also doesn't record the `reports/` gotcha or the `go.mod` pseudo-version situation.

---

## (c) NOT STARTED

1. **`LICENSE` file.** README references MIT and links to `LarsArtmann/template-LICENSE`, but no actual `LICENSE` file exists in the repo. pkg.go.dev and GitHub won't render a license badge that resolves.
2. **`.golangci.yml`.** `go-finding` has a 7.5 KB one. `go-linter-sdk` runs golangci-lint with defaults, which means the lint config drifts from the ecosystem.
3. **`CONTRIBUTING.md`, `AUTHORS`, `CONTEXT.md`.** All present in `go-finding`, all absent here.
4. **`CHANGELOG.md`, `FEATURES.md`, `TODO_LIST.md`, `ROADMAP.md`.** Required by the project's own documentation conventions (see global AGENTS.md "Project Documentation Files" table). None exist.
5. **`docs/DOMAIN_LANGUAGE.md`.** The SDK introduces named concepts (`Rule`, `RuleFunc`, `RuleMeta`, `Registry`, `Category`, `RuleError`) that would benefit from a glossary.
6. **GitHub Actions / CI workflow.** `go-finding` has `.github/`. This repo has none. The buildflow pass is local-only.
7. **`go.work` / multi-module workspace.** The repo relies on a `replace` directive to `../go-finding`, which breaks for anyone who doesn't have both repos checked out as siblings. A `go.work` would formalize the workspace.
8. **Benchmarks.** `Registry` uses `sync.RWMutex`; there are no `Benchmark*` functions. For an SDK that linters call on every run, this matters.
9. **Race-detector test for the registry mutex.** `test-race` runs but there's no test that concurrently calls `Register`/`All`/`Run`.
10. **`.goreleaser.yml`.** `go-finding` has one; this repo has none (though it's a library, so maybe not needed).
11. **pkg.go.dev publication readiness.** The README has a badge but the package isn't actually published (pseudo-version dependency).
12. **`git-town.toml`** or equivalent branch-strategy config (present in `go-finding`).

---

## (d) TOTALLY FUCKED UP

1. **The `reports/` fix is a band-aid, not a cure.** I patted myself on the back for "fixing" test-coverage by `mkdir -p reports` locally + shellHook. In reality, anyone who clones this repo and runs `buildflow` without first entering `nix develop` gets the same failure I "fixed." I should have either (a) committed `reports/.gitkeep` with a gitignore exception, (b) made the `coverage` nix app `mkdir -p` before running, or (c) filed this as a buildflow bug. I did none of those.
2. **I committed pre-existing changes without flagging them in the commit body clearly enough.** The `README.md` table reformatting and the `go.mod` `v1.2.0 → v0.0.0-...000000` pseudo-version switch were sitting in the working tree when I started. I folded them into my commit. The pseudo-version is correct _given_ the replace directive, but if `go-finding` is actually published as `v1.2.0`, consumers without the sibling checkout will break. I did not verify which is the source of truth.
3. **I did not run the full buildflow inside `nix develop`** to confirm the shellHook + env actually produce a clean pass. I ran it with `GOEXPERIMENT=jsonv2 buildflow` from my user shell, which sidesteps the nix shell entirely. The flake.nix is only proven to _evaluate_, not to _provide a working shell_.
4. **I claimed the task was complete before stress-testing the reports/ fix.** The todo list went all-green, I declared victory, and only noticed the fragility when reflecting afterwards.

---

## (e) WHAT WE SHOULD IMPROVE

### Process

1. ~~**Stop declaring victory after the happy path passes.** Run the fix in a clean-room scenario (fresh shell, no inherited state) before marking done.~~ done — clean-room/integration gates became standard (sessions 2–5; BuildFlow is the declared quality gate)
2. ~~**Verify nix flakes by entering the shell, not just `nix run`.** `nix run .#foo` executes the app script; it does not validate that `devShells.default` actually produces a usable environment.~~ done — `nix develop --command buildflow` became the canonical gate (session 2+)
3. ~~**Question inherited working-tree changes before absorbing them.** The `go.mod` pseudo-version and README reformatting were not mine. I should have put them in a separate commit or asked.~~ done — daemon-commit convention documented in `AGENTS.md`; CHANGELOG is the narrative of record
4. ~~**Add a "fresh clone" test to the mental checklist.** Would someone who clones this repo tomorrow be able to build it with one command?~~ done — fresh-clone BuildFlow verified 35/36 (session 5)

### Code

5. ~~**`RuleError.Is` only matches `ErrRuleFailed`.** It does not delegate to the cause for custom sentinels. Consider `return target == ErrRuleFailed || errors.Is(e.Cause, target)` so callers can `errors.Is(err, context.Canceled)` without thinking about Unwrap.~~ **Won't implement — `Unwrap()` already chains the cause; `errors.Is(ruleErr, context.Canceled)` passes via unwrapping (proven by `TestRuleError_Is_Canceled`).**
6. ~~**`wrapRuleError` lives in `registry.go` but is really about `RuleError`.** Consider moving it next to `NewRuleError` in `errors.go` for cohesion.~~ **Won't implement — kept in `registry.go` beside its three call sites; `errors.go` stays focused on the type.**
7. ~~**`RuleFunc.Check` wraps but `Rule` interface implementations written by users do not.** The wrapping is a `RuleFunc` convenience, not a contract. Custom `Rule` impls bypass it. Document this or move wrapping into `Registry.Run` only.~~ done — all rules (custom included) are wrapped at the `Registry.Run`/detector chokepoints; pattern documented in `AGENTS.md`
8. ~~**No `RuleError.Rule()` accessor** to fetch the `Rule` metadata, only the name string. For logging that's fine; for tooling it may be limiting.~~ **Won't implement — the `RuleID` string suffices; no consumer need has surfaced.**

### Infrastructure

9. ~~**Add `.golangci.yml`** aligned with `go-finding` so lint rules don't drift.~~ done at `6838158`
10. ~~**Commit `reports/.gitkeep`** (with `!reports/.gitkeep` in `.gitignore`) so the directory exists on fresh clone.~~ done at `6838158`
11. ~~**Add a CI workflow** that runs `nix flake check` + `nix run .#test-race` + `buildflow` on every push.~~ done — CI since session 5; auth-free and green on `487d254` (2026-09-09)

---

## (f) UP TO 50 THINGS TO DO NEXT

### High-impact, low-effort (do first)

1. ~~Commit `reports/.gitkeep` with gitignore exception; kill the shellHook mkdir hack.~~ done at `6838158`
2. ~~Add a real `LICENSE` file (MIT, copied from `template-LICENSE`).~~ done at `6838158`
3. ~~Add `.golangci.yml` mirroring `go-finding`.~~ done at `6838158`
4. ~~Verify `nix develop` actually opens a shell where `buildflow` passes end-to-end.~~ done — session 2 (`nix develop --command buildflow` 33/34, later 35/36)
5. ~~Add a GitHub Actions workflow: `nix flake check`, `nix run .#test-race`, `buildflow`.~~ done — session 5; auth-free and green on `487d254`
6. ~~Revert or justify the `go.mod` pseudo-version; clarify whether `go-finding` v1.2.0 is real.~~ done — pinned to published tags ever since (`v1.4.1` → `v1.7.0`, all public)
7. ~~Add a `.gitkeep`-style test for the fresh-clone scenario in CI.~~ done — fresh-clone BuildFlow verified manually (session 5); CI runs on every push

### Error-handling polish

8. ~~Broaden `RuleError.Is` to delegate to `errors.Is(e.Cause, target)`.~~ **Won't implement — `Unwrap()` already chains; see (e).5.**
9. ~~Move `wrapRuleError` into `errors.go`.~~ **Won't implement — see (e).6.**
10. ~~Add tests for `errors.Is(ruleErr, context.Canceled)` and `context.DeadlineExceeded`.~~ done — `TestRuleError_Is_Canceled` / `TestRuleError_Is_DeadlineExceeded` (`errors_test.go`)
11. ~~Add a test for nil-cause `RuleError`.~~ done — `TestNewRuleError_NilCause` + `FuzzNewRuleError`
12. ~~Add a test for a custom `Rule` implementation (not `RuleFunc`) to document the wrapping gap.~~ done — wrapping was moved to the shared chokepoints (see (e).7), so the gap no longer exists; chokepoint covered by `TestRegistry_Run_WrapsRuleError` and detector wrap tests
13. ~~Consider wrapping errors in `Registry.Run` only (single chokepoint) instead of both `Check` and `Run`.~~ **Won't implement — multi-chokepoint wrap with the `wrapRuleError` no-double-wrap guard is the documented design (`AGENTS.md`).**

### Testing

14. ~~Add `BenchmarkRegistry_Run` with N rules.~~ done — `BenchmarkRegistry_Run` (+ `Register`, `All`) in `registry_test.go`
15. ~~Add `BenchmarkRegistry_Register` contended.~~ done — see above
16. ~~Add a concurrent `Register`/`All`/`Run` race test.~~ done — `TestRegistry_ConcurrentReadWrite` (session 5)
17. ~~Add a test for `Registry.Run` continuing vs. failing-fast on first rule error (currently fails fast — document or make configurable).~~ done — `ContinueOnError()` option + fail-fast/continue tests (v0.2.0)
18. ~~Add a test for `ExitCodeFromReport` with suppressed findings.~~ done — nil/empty/non-empty cases in `TestExitCodeFromReport`; suppression is downstream of the report, not an SDK concept
19. ~~Add a fuzz test for `RuleError` formatting.~~ done — `FuzzNewRuleError` (7.7M execs, 0 failures)
20. ~~Add table-driven tests for `Category` constants and their mapping to `finding.Category`.~~ **Won't implement — constants are compile-time strings; the mapping is exercised via `NewFinding` tests.**

### Documentation

21. ~~Write `docs/DOMAIN_LANGUAGE.md` (Rule, RuleFunc, RuleMeta, Registry, Category, RuleError, Detector).~~ done — session 6
22. ~~Write `FEATURES.md` (DONE: Rule/Registry/Detector/ExitCode; PLANNED: per-category filters, severity-tiered exit codes).~~ done — session 4 (one-home rule: PLANNED lives in ROADMAP only)
23. ~~Write `TODO_LIST.md`.~~ done at `4691f35`
24. ~~Write `ROADMAP.md`.~~ done at `4691f35`
25. ~~Write `CHANGELOG.md` (v0.1.0: initial scaffolding + error types).~~ done at `6838158`
26. ~~Write `CONTRIBUTING.md` (or copy from `go-finding`).~~ done at `6838158`
27. ~~Add `AUTHORS`.~~ **Won't implement — `LICENSE` carries the copyright; go-finding parity deemed unnecessary.**
28. ~~Update `AGENTS.md` with the `reports/` gotcha and the `go.mod` situation.~~ done — "Gotchas & conventions" section (session 4)
29. ~~Add package-level examples (`ExampleRegistry_Run`, `ExampleRuleFunc`) visible in pkg.go.dev.~~ done — 7 `Example*` functions in `example_test.go`; verified rendering on pkg.go.dev (2026-09-09)
30. ~~Revisit README "Status" section — it says "no active consumers"; confirm still true.~~ done — status callout refreshed; still true (zero production consumers)

### Infrastructure

31. ~~Add `go.work` (or justify the replace-only approach).~~ done — `go.work` created 2026-08-08 (gitignored, not committed)
32. ~~Add `.gitattributes` (go-finding has one).~~ done at `6838158`
33. ~~Add `git-town.toml` if the team uses git-town.~~ **Won't implement — no git-town in this ecosystem's workflow.**
34. Configure Renovate / Dependabot for nix + go deps.
35. Add a `nix run .#update` app for dependency updates.
36. Pin `golangci-lint` version in devShell (currently floats with nixpkgs).
37. ~~Add a `devShells.full` with extra tooling (`art-dupl`, `goreleaser`).~~ **Won't implement — `devShells.default` already has everything the workflow needs.**

### Ecosystem alignment

38. ~~Confirm the `DetectorFromRegistry` adapter signature matches what BuildFlow's `toolsdk.Spec.Detect` actually expects.~~ done — verified; cross-referenced in godoc and exercised by tests + examples
39. ~~Pilot-port one rule from `go-structure-linter` to prove the pattern.~~ done — `examples/no-go-mod` (v0.2.0)
40. Pilot-port one rule from `branching-flow` to validate the converter-deletion claim (1,871 LOC). ← still open — TODO_LIST #7
41. Add a `linter.NewRegistryFromRules(...Rules)` convenience constructor. ← still open — ROADMAP Theme 1
42. ~~Add `Registry.Deregister(name)` for symmetry (even if rarely used).~~ done — `Deregister(id)` shipped in v0.2.0
43. ~~Add `Registry.Has(name bool)` / `Registry.Get(name) (Rule, bool)`.~~ done — `Has(id)` / `Get(id)` shipped in v0.2.0
44. ~~Consider `Registry.Run` returning partial results on error (currently discards).~~ done — `ContinueOnError()` collects partial findings (v0.2.0)

### API hardening

45. ~~Make `Category` a distinct type with validation (currently a raw `string`).~~ **Won't implement as stated — `Category` IS a distinct type (`type Category string`) that is deliberately OPEN; validating custom values would close it (documented in `rule.go` + README).**
46. Add `Category.All()` returning all valid categories. ← still open — ROADMAP Theme 1
47. ~~Add `RuleMeta.Validate()` (non-empty name, valid category, valid severity).~~ done — shipped in v0.2.0 (all four identity fields)
48. Consider a `RuleSet` typed wrapper around `[]Rule` for consumers that don't want a mutex'd registry. ← still open — ROADMAP Theme 1
49. Add `ExitCodeFromFindings([]Finding)` convenience that skips building a Report. ← still open — ROADMAP Theme 1
50. ~~Finalize semver contract: document that the package is pre-v1 and what that means for consumers.~~ done — SemVer commitment in `CHANGELOG.md` header; tagged v0.1.0 → v0.3.0

---

## (g) QUESTIONS I CANNOT ANSWER MYSELF

1. ~~**Is `go-finding v1.2.0` (the pre-existing `go.mod` value I overwrote) actually a published version, or was the pseudo-version already the intended state?**~~ resolved — the published lineage is real and public (`v1.4.1` → `v1.7.0`); no replace directive since 2026-07-30 (ROADMAP Q1, resolved)
2. ~~**Should `go-linter-sdk` get its own `LICENSE` file, or is the README link to `LarsArtmann/template-LICENSE` the canonical pattern for all LarsArtmann repos?**~~ resolved — physical MIT `LICENSE` file created at `6838158`
3. ~~**Is the `reports/` directory an ecosystem convention (BuildFlow always writes there) or just `go-finding`'s local choice?**~~ resolved — treated as per-repo convention: `reports/.gitkeep` + gitignore exception pattern adopted and fresh-clone verified (session 5)

---

_Generated by reflection on session 2026-07-19 ~01:17–01:38 CEST. No external research performed; all claims derive from this session's tool output._
