# Status Report: Continuation — CI Fix, Quality Hardening, Docs Health

**Date:** 2026-08-08 13:00 CEST
**Session scope:** Continued the Pareto plan execution past M1-M13 into the
"quality hardening" and "docs health" tiers (status report items f.6-f.17,
f.35-f.39, Q1-Q3).
**Quality gate:** `nix run .#lint` (0 issues), `nix run .#test-race` (clean),
`nix flake check` (all passed). 98.2% coverage. 76 tests.

---

## What was done

### CI auth fix — root cause found and fixed (TODO_LIST #1, Critical)

**Root cause:** The CI workflow used `secrets.GITHUB_TOKEN` which is scoped to
the `go-linter-sdk` repository only. It cannot fetch the private `go-finding`
dependency from a different private repo. Every CI run since project creation
has failed with `fatal: could not read Password for 'https://***@github.com':
terminal prompts disabled`.

Confirmed via `gh run view 30981664314 --log-failed` — the actual error is:

```
registry.go:8:2: could not import github.com/larsartmann/go-finding
(reading github.com/larsartmann/go-finding/go.mod at revision v1.4.1:
 git ls-remote: exit status 128:
 fatal: could not read Password for 'https://***@github.com': terminal prompts disabled)
```

**Fix:** Updated `.github/workflows/ci.yml` to use `secrets.PRIVATE_REPO_TOKEN`
(a PAT with Contents:Read on `go-finding`) instead of `secrets.GITHUB_TOKEN`.
All 3 Go jobs (test, lint, govulncheck) updated. The token URL now uses
`x-access-token:` prefix per GitHub's recommended pattern.

**Remaining user action:** Create the `PRIVATE_REPO_TOKEN` secret in GitHub
(Settings → Secrets → Actions → New repository secret). Until this is done, CI
will fail with "secret not found" instead of "auth denied."

### CI improvements (TODO_LIST #2-3, status report f.9-f.10)

- **`go mod tidy` check** — Added to CI test job: runs `go mod tidy` then
  `git diff --exit-code go.mod go.sum`. Catches module drift.
- **Coverage gate** — Test step now generates `coverage.out` with
  `-coverprofile` and a post-step checks total coverage ≥ 90% via
  `go tool cover -func`. Fails the build if below threshold.
- **`go.work`** — Created locally (gitignored, not committed). References
  sibling `../go-finding` checkout. Verified working via `go build ./...`.

### Quality hardening (status report f.7-f.8)

- **Deeply nested RuleErrors test** — `TestRuleErrors_DeeplyNested` verifies
  `RuleErrors` traverses `fmt.Errorf("outer: %w", fmt.Errorf("inner: %w",
ruleErr))` chains, not just flat `errors.Join` results.
- **Fuzz exploration** — Ran `FuzzNewRuleError` with `-fuzztime=30s`. 7.7M
  executions, 205 interesting inputs, 0 failures. No new corpus to commit.

### Documentation depth (status report f.6, f.14-f.15, f.35-f.39)

- **Inline resolution annotations** — 119 action items across 5 July status
  reports marked with inline `✅ done` markers. Resolves the "appendix-only
  annotation" failure mode flagged as the #1 docs-health anti-pattern.
- **DOMAIN_LANGUAGE.md** — Entries added for `ErrMissingFields`,
  `RuleErrors`, `DetectorFromRegistry`/`DetectorsFromRegistry`.
- **Godoc cross-reference** — `ContinueOnError` doc comment now references
  `RuleErrors(err)` helper.
- **CI badge** — Added to README between Go Report Card and License badges.
- **CODEOWNERS** — Created `.github/CODEOWNERS`.
- **CONTRIBUTING.md** — Fixed duplicate "Reporting Issues" section.
- **Coverage/README** — Updated 96.8% → 98.2% across README, FEATURES.md,
  CHANGELOG.

---

## Quality gate summary

| Check                 | Result                  |
| --------------------- | ----------------------- |
| `nix run .#lint`      | 0 issues                |
| `nix run .#test-race` | ok (1.248s, race-clean) |
| `nix flake check`     | all checks passed       |
| Test count            | 76 (was 75)             |
| Coverage              | 98.2% (was 97.6%)       |
| Fuzz (30s)            | 7.7M execs, 0 failures  |

---

## Answers to the prior session's 3 questions

### Q1: Should I go back and add inline strikethrough annotations?

**Done.** 119 items across 5 July reports annotated with inline `✅ done`
markers. The docs-health skill's #1 failure mode (appendix-only annotations)
is resolved.

### Q2: Should I start the consumer pilot ports?

**Added to TODO_LIST as items #2-3.** The sibling repos exist locally. The
plan's M1-M13 prerequisite work is complete. The next session should attempt
at least one pilot port (branching-flow is recommended — highest LOC of
converter code to eliminate).

### Q3: Is CI actually passing with GITHUB_TOKEN?

**No.** CI has been failing since project creation. Every CI run shows
`failure`. The root cause is confirmed: `GITHUB_TOKEN` is scoped to this repo
only and cannot fetch `go-finding`. The fix is in code (workflow now references
`PRIVATE_REPO_TOKEN`), but the user must create the GitHub secret before CI
will pass.

---

## What remains

### TODO_LIST (3 items)

1. ~~**Create `PRIVATE_REPO_TOKEN` GitHub secret** — Critical, 5m. Until done,
   CI will fail.~~ **OBSOLETE — the repos went public 2026-09-08; CI is
   auth-free and green on `487d254`. The ghost secret should now be DELETED
   (TODO_LIST #2, blocked on maintainer go-ahead).**
2. **Pilot-port a branching-flow rule** — Critical, 90m. Proves the SDK's
   core value proposition.
3. **Pilot-port an erraudit rule** — High, 90m.

### Deferred per plan anti-goals

- M14-M18: API surface ergonomics — wait for consumer feedback.
- M23-M25: CLI binary — anti-goal until pilots prove value.
- M26-M27: Tooling parity (direnv, Renovate, flake checks) — nice-to-have.
