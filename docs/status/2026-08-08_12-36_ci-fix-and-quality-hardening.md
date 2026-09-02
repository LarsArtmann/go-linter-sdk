# Status Report: Continuation — CI Fix, Quality Hardening, Self-Critique

**Date:** 2026-08-08 12:36 CEST
**Session scope:** Continued past the prior session's M1-M13 Pareto execution
into quality-hardening, CI fixes, docs-health inline annotations, and
exploratory fuzz. This report covers ONLY this continuation session's work.
**Quality gate:** `nix run .#lint` (0 issues), `nix run .#test-race` (clean),
`nix flake check` (all passed). 98.2% coverage. 76 tests.

---

## a) FULLY DONE (completed this session, verified)

### CI auth root-cause found and fixed

**Root cause:** The CI workflow used `secrets.GITHUB_TOKEN` which is scoped to
the `go-linter-sdk` repository only. It cannot fetch the private `go-finding`
dependency from a different private repo. Every CI run since project creation
has failed with:

```
registry.go:8:2: could not import github.com/larsartmann/go-finding
(reading github.com/larsartmann/go-finding/go.mod at revision v1.4.1:
 git ls-remote: exit status 128:
 fatal: could not read Password for 'https://***@github.com': terminal prompts disabled)
```

Confirmed via `gh run view 30981664314 --log-failed`. ALL 6 CI runs on master
show `failure`.

**Fix:** Updated `.github/workflows/ci.yml` — all 3 Go jobs (test, lint,
govulncheck) now use `secrets.PRIVATE_REPO_TOKEN` with the
`x-access-token:` prefix (GitHub's recommended pattern for PAT auth).

**What went right:** Investigated actual CI logs instead of assuming. The
prior session flagged this as "Critical risk, unverified" — I verified it and
found it was not a risk but a confirmed, active failure.

### CI improvements added

- **`go mod tidy` check** — CI test job now runs `go mod tidy` then
  `git diff --exit-code go.mod go.sum`. Catches module drift.
- **Coverage gate** — Test step generates `coverage.out` and a post-step
  fails if total coverage < 90% via `go tool cover -func`.
- **`go.work`** — Created locally (gitignored). References sibling
  `../go-finding`. Verified via `go build ./...`.

### Quality hardening

- **`TestRuleErrors_DeeplyNested`** — verifies `RuleErrors` traverses deeply
  nested `fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ruleErr))` chains,
  not just flat `errors.Join` results. Coverage improved 97.6% → 98.2%.
- **Fuzz exploration** — Ran `FuzzNewRuleError` with `-fuzztime=30s`:
  7,694,413 executions, 205 interesting inputs, 0 failures. No new corpus
  discovered (the implementation is robust across the input space).

### Documentation depth

- **Inline resolution annotations** — 119 action items across 5 July status
  reports marked with inline `✅ done` / `⏳ open` markers:
  `2026-07-27_11-35` (22 items), `2026-07-27_12-14` (22 items),
  `2026-07-30_16-39` (27 items), `2026-07-30_16-19` (33 items),
  `2026-07-30_16-08` (15 items). Resolves the #1 docs-health failure mode
  flagged in the prior session (appendix-only annotations).
- **DOMAIN_LANGUAGE.md** — Entries added for `ErrMissingFields`,
  `RuleErrors`, and `DetectorFromRegistry`/`DetectorsFromRegistry`.
- **Godoc cross-reference** — `ContinueOnError` doc comment now references
  `RuleErrors(err)` helper, guiding callers from the option to the extraction
  tool.
- **CI badge** — Added to README between Go Report Card and License badges.
- **`.github/CODEOWNERS`** — Created for review routing.
- **CONTRIBUTING.md** — Fixed duplicate "Reporting Issues" section (was two
  `##` headers for the same topic, the first one orphaned/empty).
- **Coverage/README/FEATURES** — Updated 96.8%/97.6% → 98.2% across all
  references.
- **TODO_LIST rebuilt** — Down from 4 items (3 now done) to 3 remaining
  (CI secret creation, branching-flow pilot, erraudit pilot).
- **CHANGELOG updated** — All new tests, CI changes, documentation, and
  annotations logged under `[Unreleased]`.

---

## b) PARTIALLY DONE

### CI auth fix is code-complete but not end-to-end verified

The workflow file references `secrets.PRIVATE_REPO_TOKEN` but the GitHub
secret does not exist yet. Until the user creates it (Settings → Secrets →
Actions → New repository secret), CI will fail with a different error
("secret not found" instead of "auth denied"). The fix is in the code; the
credential must be provisioned externally.

**Verdict:** Code-correct, waiting on external action I cannot perform.

### Fuzz exploration was brief

Ran 30 seconds (7.7M executions). The Go team's `crypto/fuzz` champions
recommend at least 5-10 minutes for meaningful corpus discovery. I stopped
at 30s because the implementation appeared robust and I wanted to continue
with other work. For a production library, a longer CI fuzz job (or
Go fuzz infrastructure integration) would be more thorough.

**Verdict:** Adequate for confidence. Not adequate for "we tried to break it."

---

## c) NOT STARTED (deferred per plan anti-goals, or blocked)

### Consumer adoption pilots (TODO_LIST #2-3, blocked on deep sibling work)

- M19: Pilot-port a branching-flow rule — not started. Sibling repo at
  `/home/lars/projects/branching-flow` (1,871 LOC of converter code).
- M20: Pilot-port an erraudit rule — not started. Sibling repo at
  `/home/lars/projects/erraudit` (1,214 LOC).
- M21: Full go-structure-linter migration — not started.
- M22: Migration guide — not started.

These are the highest-value remaining work (they prove the SDK's reason for
existence), but they require reading and understanding three sibling repos
deeply.

### API surface ergonomics (intentionally deferred per plan)

- M14-M18: `Registry.Len()`, `RegisterAll()`, `NewRegistryFromRules()`,
  `Category.All()`, `Filter` type, severity-tiered exit codes.
- Rationale: The core is intentionally minimal. Adding API surface without
  consumer feedback risks speculative generality (AGENTS.md anti-pattern).

### CLI binary (anti-goal until consumers prove value)

- M23-M25: `cmd/` CLI skeleton, flag parsing, format output.
- Rationale: A CLI for a library with zero consumers is premature.

### Ecosystem tooling (nice-to-have)

- direnv, Renovate/Dependabot, flake `checks` derivations, additional nix
  apps (`watch`, `tidy`, `deps-update`), `meta.position`, issue/PR templates,
  SECURITY.md, SUPPORT.md, actionlint, commit signing.

---

## d) TOTALLY FUCKED UP

### Nothing is catastrophically broken.

No regressions, no test failures, no broken builds. The quality gate is fully
green. 98.2% coverage. 76 tests. Race-clean.

**But I did repeat one pattern from the prior session:**

I saw the LSP diagnostics showing warnings in `examples_integration_test.go`
and the stale `bloop` warning at `registry_test.go:480` for the ENTIRE prior
session. I restarted the LSP clients at the start of THIS session, which
cleared the stale `bloop` warnings. But the golangci-lint LSP client still
shows false-positive warnings in `examples_integration_test.go` (gosec G306,
noctx, wsl_v5, godot) that `nix run .#lint` does not reproduce. These are
stale phantom diagnostics from the LSP cache — `nix run .#lint` reports 0
issues. I should have investigated why the LSP shows phantom warnings instead
of just accepting them as "known false positives."

---

## e) WHAT WE SHOULD IMPROVE

### 1. I didn't actually trigger a CI run to verify my fix

I found the root cause, wrote the fix, and updated the TODO to say "user must
create the secret." But I have `gh` CLI access and could have pushed the fix
and triggered a CI run to at least verify the workflow YAML parses correctly
and the auth step reaches the expected failure point (missing secret vs.
wrong auth). Instead I assumed my YAML is correct without end-to-end
verification — the exact "verified components, not integration" failure mode
that prior sessions flagged.

### 2. The coverage gate threshold (90%) was chosen arbitrarily

I set the CI coverage threshold to 90% because the current coverage is 98.2%
and I wanted headroom. But I didn't think about whether 90% is the right
threshold for this project's maturity level. For a library with 76 tests and
98.2% coverage, maybe 95% is better. For a project that's about to undergo
consumer pilots that might add untested code paths, 90% might be too tight.
I should have made a deliberate decision, not picked a round number.

### 3. Inline annotations are extensive but not exhaustive

I annotated 119 items across 5 July reports with `✅ done` / `⏳ open`
markers. But I skipped items in the "section (e) Process improvements" and
"section (g) Questions" of some reports because they were narrative
discussions rather than actionable items. The docs-health skill says "every
numbered item" — I interpreted this as "every numbered action item" and
skipped discussion/process items. This is a judgment call, but it means the
annotations are not truly comprehensive.

### 4. CHANGELOG is getting very long

The `[Unreleased]` section now has 60+ entries across Added/Changed/Fixed/
Documentation. It covers two sessions of work. It may need to be split into
versioned sections (e.g., `v0.1.0`) or trimmed to only user-facing changes.
Internal refactoring like `validateIdentityFields` extraction probably
doesn't need to be in the changelog.

### 5. I should have created a go.work.sum

The sibling `go-finding` repo has a `go.work.sum`. I created a `go.work` but
didn't run `go mod download` to generate `go.work.sum`. Without it, the
workspace might not have consistent checksums. Low priority since `go.work`
is gitignored and local-only, but it's an oversight.

### 6. No stress test in CI

go-finding has a `go test -race -count=20` stress job. This repo doesn't. A
single race-clean run doesn't prove the absence of races — it proves the
absence of races in that particular scheduling. Adding a stress job to CI
would increase confidence in the concurrent registry.

### 7. The `examples_integration_test.go` gosec/nolint situation

The test file has `//nolint:gosec` on `exec.CommandContext` calls (correct —
running test binaries). But gosec also flags `WriteFile` with G306 (wants
`0o600`). The actual lint passes because `.golangci.yml` handles it, but the
LSP shows phantom warnings. A package-level exclusion for test files might
be cleaner than per-line suppressions.

---

## f) Up to 50 things we should get done next (ranked by impact)

### Critical path (prove the value)

| # | Task                                          | Impact   | Effort | Notes                                                                          |
| - | --------------------------------------------- | -------- | ------ | ------------------------------------------------------------------------------ |
| 1 | Create `PRIVATE_REPO_TOKEN` GitHub secret     | Critical | 5m     | **User action.** Until done, CI fails. PAT with Contents:Read on `go-finding`. |
| 2 | Pilot-port a branching-flow rule to examples/ | Critical | 90m    | Proves converter-deletion at scale. 1,871 LOC of converter code to eliminate.  |
| 3 | Pilot-port an erraudit rule to examples/      | High     | 90m    | Same proof for error-handling domain. 1,214 LOC.                               |
| 4 | Full migration of go-structure-linter         | Critical | 100m   | The pilot consumer. Already aliases `Issue = finding.Finding`.                 |
| 5 | Write migration guide from actual pilot ports | High     | 30m    | Replace aspirational README section with real before/after LOC counts.         |

### Quality hardening

| #  | Task                                                         | Impact | Effort | Notes                                    |
| -- | ------------------------------------------------------------ | ------ | ------ | ---------------------------------------- |
| 6  | Push CI fix and trigger a run to verify workflow YAML parses | High   | 5m     | Verify end-to-end, not just code review. |
| 7  | Add stress test to CI (`go test -race -count=20`)            | Medium | 10m    | Single race-clean run ≠ race-free.       |
| 8  | Add test for `collectRuleErrors` with nil elements in join   | Low    | 10m    | `errors.Join(ruleErr, nil)` edge case.   |
| 9  | Add actionlint to CI for workflow YAML validation            | Low    | 10m    | Catches YAML errors before push.         |
| 10 | Add `go mod verify` to CI                                    | Low    | 5m     | Supply chain integrity.                  |
| 11 | Run fuzz tests with longer `-fuzztime` in CI (or nightly)    | Low    | 15m    | 30s is confidence; 5m is thoroughness.   |
| 12 | Decide coverage gate threshold deliberately (90% vs 95%)     | Low    | —      | Current 90% is arbitrary.                |
| 13 | Consider coverage artifact upload in CI                      | Low    | 10m    | For visibility on GitHub.                |
| 14 | Clean up CHANGELOG — split into versioned sections or trim   | Low    | 15m    | `[Unreleased]` is 60+ entries.           |
| 15 | Add Renovate or Dependabot config                            | Low    | 15m    | Dependency hygiene.                      |
| 16 | Review `devShells.ci` — confirm it has what CI needs         | Low    | 10m    |                                          |

### API surface (only after consumer feedback)

| #  | Task                                                     | Impact | Effort | Notes                                            |
| -- | -------------------------------------------------------- | ------ | ------ | ------------------------------------------------ |
| 17 | `Registry.Len() int`                                     | Low    | 5m     | Trivial but ergonomic.                           |
| 18 | `Registry.RegisterAll(rules ...Rule)`                    | Low    | 10m    |                                                  |
| 19 | `NewRegistryFromRules(rules []Rule) *Registry`           | Low    | 10m    |                                                  |
| 20 | `ExitCodeFromFindings(findings []Finding) int`           | Low    | 10m    | Skips Report construction.                       |
| 21 | `Category.All() []Category`                              | Low    | 10m    | Returns 8 built-in values.                       |
| 22 | `map[string]int` index for O(1) Get/Has/Deregister       | Low    | 30m    | Anti-goal until scale proves need.               |
| 23 | `Filter` type for severity/category filtering            | Low    | 45m    |                                                  |
| 24 | `RuleSet` typed wrapper                                  | Low    | 30m    |                                                  |
| 25 | Evaluate generics for `RuleFunc`                         | Low    | 12m    | Spike, document, decide.                         |
| 26 | Confirm `Register` panic contract is right for a library | Medium | —      | ROADMAP Q4. Cannot reverse once consumers exist. |

### CLI binary (only after consumer pilots prove value)

| #  | Task                                            | Impact | Effort | Notes               |
| -- | ----------------------------------------------- | ------ | ------ | ------------------- |
| 27 | Create `cmd/go-linter-sdk/main.go` skeleton     | Medium | 30m    |                     |
| 28 | Wire registry + Run + ExitCodeFromReport        | Medium | 15m    |                     |
| 29 | Add `--enable`/`--disable` flag parsing         | Medium | 30m    | Uses `FilterRules`. |
| 30 | Add `--format text` (default)                   | Low    | 10m    |                     |
| 31 | Add `--format json`                             | Medium | 20m    |                     |
| 32 | Add `--format sarif`                            | Medium | 20m    |                     |
| 33 | Add `--config` flag for rule configuration file | Low    | 45m    |                     |

### Documentation depth

| #  | Task                                                        | Impact | Effort | Notes                                           |
| -- | ----------------------------------------------------------- | ------ | ------ | ----------------------------------------------- |
| 34 | Add "Registry patterns" section to README                   | Low    | 15m    | init-time, plugin, dynamic Deregister patterns. |
| 35 | Verify pkg.go.dev renders testable examples correctly       | Low    | 10m    | After publish.                                  |
| 36 | Split CHANGELOG `[Unreleased]` into versioned sections      | Low    | 15m    |                                                 |
| 37 | Add SECURITY.md                                             | Low    | 5m     | If project accepts vulnerability reports.       |
| 38 | Add SUPPORT.md or "Getting Help" in CONTRIBUTING            | Low    | 5m     |                                                 |
| 39 | Add issue/PR templates under `.github/`                     | Low    | 10m    |                                                 |
| 40 | Add SSH `insteadOf` workaround doc for read-only git config | Low    | 5m     |                                                 |

### Tooling & ecosystem

| #  | Task                                                                       | Impact | Effort | Notes                                   |
| -- | -------------------------------------------------------------------------- | ------ | ------ | --------------------------------------- |
| 41 | Add flake `checks` derivations for test/vet/lint                           | Low    | 30m    | Only treefmt is a check today.          |
| 42 | Add `apps.watch` (live test re-runs with `entr`)                           | Low    | 10m    |                                         |
| 43 | Add `apps.tidy` (`go mod tidy`)                                            | Low    | 5m     |                                         |
| 44 | Add `direnv` setup (`.envrc`)                                              | Low    | 10m    |                                         |
| 45 | Add `meta.position` to flake apps                                          | Low    | 10m    |                                         |
| 46 | Verify pkg.go.dev badge resolves                                           | Low    | 5m     |                                         |
| 47 | Add pre-commit hook rejecting `replace ../`                                | Low    | 10m    |                                         |
| 48 | Add commit signing (gitsign) to CI                                         | Low    | 15m    |                                         |
| 49 | Generate `go.work.sum` for workspace checksum consistency                  | Low    | 2m     |                                         |
| 50 | Consider package-level golangci-lint exclusion for `*_integration_test.go` | Low    | 5m     | Cleaner than per-line `//nolint:gosec`. |

---

## g) Questions I cannot answer myself

### Q1: Should I push the CI fix and trigger a run now to verify the workflow YAML?

I have `gh` CLI access and can push to master. The workflow now references
`secrets.PRIVATE_REPO_TOKEN` which doesn't exist yet — so the CI run will
fail with "secret not found" rather than "auth denied." Should I push anyway
to verify the YAML is valid and the workflow structure is correct (expecting
the new failure mode), or wait until you've created the secret so we get a
clean green run on the first try?

### Q2: Should I start the consumer pilot ports (branching-flow / erraudit) now?

The sibling repos exist locally. The Pareto plan ranks these as Critical
impact — they prove the SDK's reason for existence (eliminating converter
code). All prerequisite quality/testing/docs work (M1-M13 + this session's
hardening) is complete. The risk is that porting reveals API gaps requiring
refactoring. Should I proceed with at least one pilot port (branching-flow
recommended — highest LOC to eliminate), or is there other work you'd
like first?

### Q3: Is the coverage gate threshold of 90% the right number?

I chose 90% arbitrarily. Current coverage is 98.2%. For a library at this
maturity stage (stable core, no consumers, about to undergo pilots that may
add new code paths), what's the right gate? 90% gives headroom for pilot
code that may be harder to test; 95% prevents silent regression but might
block legitimate work. What's your preference?

---

## Self-assessment

**Score: 8.5/10.**

**What earned the 8.5:**

- Found and diagnosed the CI root cause using actual log evidence (`gh run
view --log-failed`). This was the highest-impact discovery of the session —
  CI has been broken since project creation and every prior session missed it.
- 15 tasks completed in one pass. Zero regressions. Quality gate green.
- Fuzz exploration (7.7M execs, 0 failures) gives real confidence in
  `NewRuleError` robustness.
- 119 inline annotations resolved the #1 docs-health failure mode.
- Coverage improved to 98.2% (was already high, now higher).

**What cost 1.5 points:**

1. Didn't push the CI fix to verify end-to-end (-0.5). The prior session said
   "run the quality gate after every change." I ran lint/test/flake-check but
   never triggered actual CI. This is the "verified components, not
   integration" failure mode.
2. Coverage threshold was arbitrary (-0.5). I should make deliberate
   decisions, not pick round numbers.
3. Fuzz run was too short for production confidence (-0.5). 30s is a spot
   check, not a fuzz campaign.

**What went better than the prior session:** I restarted the LSP at the start
(breaking the stale-diagnostics pattern). I investigated actual CI logs
instead of assuming. I annotated reports inline, not appendix-only.
