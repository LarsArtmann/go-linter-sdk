# Status Report — BuildFlow green on Go 1.27 + `encoding/json/v2` cleanup

- **Date:** 2026-10-11 07:11 (Sunday, CEST)
- **Repo:** `github.com/larsartmann/go-linter-sdk`
- **Session trigger:** a failing `buildflow --fix --build-mode=full --budget 5m --log-level warn --max-time 5m` run (exit 69, 2 failed steps, 4 tools with remaining findings, 7 detect-only tool reports).
- **Session outcome:** BuildFlow exits **0** (`passed with warnings 56/64, 83 findings`), all gate-level (error-severity) findings cleared.
- **HEAD at report time:** `417e45a` (auto-commit daemon); my working-tree edits already committed by the daemon.
- **Format override:** this report is `.md` at the user's explicit request (skill canonical is styled HTML). Honored as an explicit override, not propagated to the skill default.

---

## TL;DR

| Metric | Value |
| --- | --- |
| Gate-blocking failures at session start | 2 failed steps (`nix-build`, `nix-build-verify`) + 3 error-severity findings (`go-structure-linter` AGENTS.md age, `securitymd` ×2) |
| Gate-blocking failures now | **0** — `buildflow` exits 0 |
| Warnings remaining (non-gating) | 2 after-auto-fix (`go-structure-linter` internal/, `go-work-sync`) + 6 detect-only tools |
| Root cause of the hard failure | Go floor bumped to 1.27 but `flake.nix` still pinned `go_1_26`; treefmt's `goimports` shelled out to the older sandbox Go and tried to download a newer toolchain in the network-less `nix flake check` sandbox |
| Bonus discovery | `GOEXPERIMENT=jsonv2` is obsolete on Go 1.27 and was removed repo-wide |

---

## a) FULLY DONE

1. **treefmt `nix flake check` failure fixed.** `flake.nix` `goPkg` `pkgs.go_1_26` → `pkgs.go_1_27`, and treefmt's `goimports` overridden to `pkgs.gotools.override { go = goPkg; }` so the formatter's runtime `go` matches `go.mod`'s floor.
   - Evidence: `nix build .#checks.x86_64-linux.treefmt` → succeeds; `nix flake check --no-build` → `all checks passed!` (only a benign "omitted incompatible systems" warning).
   - Files: `flake.nix`.
2. **`go-version-auto-configure` finding cleared.** The tool flagged "flake.nix pins Go 1.26 but the module floor is 1.27". Re-check (`BUILDFLOW_NO_RESULT_CACHE=1`) → **no findings**.
3. **`securitymd` gate cleared (2 error-severity findings).** Rewrote `SECURITY.md` with a "Supported Versions" table, a "Security Practices" section, and a 48-hour initial-response commitment. Re-check → **0 findings**.
4. **`go-structure-linter` AGENTS.md-age error cleared.** `AGENTS.md` was 23 days old (max 14). Content refreshed (see below); the age error no longer appears — only the separate `internal/` warning remains.
5. **`GOEXPERIMENT=jsonv2` removed repo-wide** (obsolete on Go 1.27):
   - `flake.nix` (both devShells + all 8 apps), `.buildflow.yml`, `.golangci.yml` (dropped the `goexperiment.jsonv2` build tag, `run.go: 1.26.7` → `1.27.0`), `.github/workflows/ci.yml`, `README.md`, `CONTRIBUTING.md` (whole "Why it is mandatory" section removed), `FEATURES.md`, and four `examples/*/main.go` godoc comments.
   - Evidence: `GOWORK=off GOEXPERIMENT= go build -a ./...` succeeds; `GOEXPERIMENT= go list std` lists `encoding/json/v2` and `encoding/json/jsontext`; `nix run .#lint` → `0 issues`.
6. **Verified the "never set `GOEXPERIMENT=none`" hazard.** `GOWORK=off GOEXPERIMENT=none go build ./...` fails with `build constraints exclude all Go files in .../encoding/json/jsontext`. (Important nuance: empty `GOEXPERIMENT=` uses defaults and is safe; `=none` disables the default experiment set and breaks the build.)
7. **`go-work-sync` "stale module reference" diagnosed** (see (d) — it is an expected false positive, documented, not repaired).
8. **`CHANGELOG.md` [Unreleased] updated** with the GOEXPERIMENT removal and the `SECURITY.md` completion.
9. **`AGENTS.md` gotchas extended** with: the three-place Go-version lockstep + the treefmt `goimports` toolchain pin (root cause of the failure), the `go.work`/`go-work-sync` false positive, and the `internal/` deliberate non-fix.
10. **Green re-verification end-to-end:** `buildflow --fix --build-mode=full …` → exit 0; `nix run .#lint` → 0 issues; `GOWORK=off GOEXPERIMENT= go test ./...` → `ok`.

---

## b) PARTIALLY DONE

1. **`go-structure-linter`: 2 findings at start → 1 remaining.** The AGENTS.md age **error** is resolved; the `internal/` suggestion remains as a **warning** and is documented as a deliberate non-fix (public library, root `linter` package is the API). Remaining: decide whether to opt the rule out in `.buildflow.yml` `skip_steps` / suppression, or accept the permanent warning. Effort: S.
2. **`go-work-sync`: 1 warning remains.** Diagnosed as a buildflow limitation (`DetectGoWorkDrift`/`FindAllGoModDirs` only scan under the repo root, so the out-of-repo `use ../go-finding` can never be "found on disk"). Documented; the local `go.work` was intentionally **not** deleted to avoid disrupting a concurrent session. Remaining: decide keep-and-document vs. delete `go.work`. Effort: S.
3. **`AGENTS.md` jsonv2 section — two authors, two framings.** A concurrent session had already rewritten the top section as "the experiment is on by default"; I added a matching gotcha. The result is coherent but written by two agents; it should be read once end-to-end for a single voice / no split-brain. Effort: S.
4. **`CHANGELOG.md` [Unreleased]** — the concurrent session's toolchain entry and my GOEXPERIMENT entry coexist; not merged into one narrative bullet. Effort: S.
5. **Doc-age warnings (`doc-files-age-check`)** — I refreshed `AGENTS.md`/`FEATURES.md`/`SECURITY.md`/`CHANGELOG.md`, but `README.md`, `TODO_LIST.md`, and the full `FEATURES.md` inventory are still reported stale (576h / 720h / 768h). Effort: M.

---

## c) NOT STARTED

1. **`file-size-check` (3 warnings):** `registry_test.go` 1135 lines, `registry.go` 371, `examples/oops-redundant-guard/main.go` 367 (max 350). No splitting attempted.
2. **`art-dupl` (22 warnings):** duplication across `examples_integration_test.go` (lines 116–161). No helper extraction attempted.
3. **`branching-flow` (8 info findings):** strong-ID suggestions for `RuleID`/`Get(id)`/`Has(id)`/`Deregister(id)`. No design decision made.
4. **`govulncheck` (16 warnings):** "package requires newer Go version go1.27" — the nixpkgs govulncheck binary's Go is older than 1.27. Not investigated.
5. **`vulnix` (31 warnings):** nixpkgs-input CVEs (async, binutils, bison, cargo, …). Not triaged; no flake input update forced.
6. **BuildFlow binary freshness:** the running binary was built at `25b07c4`, HEAD in the BuildFlow repo is `fbd7a42` — advisory warns results may not reflect current BuildFlow. Not rebuilt.
7. **Global shell `GOEXPERIMENT=jsonv2`:** still exported by the user's environment, so the buildflow preflight warning persists. Not touched (user-level env; source unknown — not in `crush.json`, `.bashrc`, `.profile`, or `.envrc`).
8. **CI execution verification:** `ci.yml` was edited (env block) but the workflow was not run; `actionlint` not re-run locally.
9. **TODO_LIST.md / ROADMAP.md harvest:** section (f) below is not yet routed into them.
10. **`test-race`:** not run explicitly this session (full-mode buildflow includes it, but I did not isolate it).

---

## d) TOTALLY FUCKED UP

1. **The repo is a shared checkout with a live concurrent session — and I did not detect it until my `multiedit` on `AGENTS.md` was rejected ("file has been modified since it was last read").** Severity: process hazard, not data loss. Immediately before that failure the daemon had committed 4 of my files (`ab55dc2`), then another session rewrote `AGENTS.md` under me. Root cause: two Crush sessions + the auto-commit daemon on one working tree, no locking. Mitigation: I re-read before editing and never forced a clobber; but a less careful pass would have silently overwritten the other session's AGENTS.md. This is the single most dangerous thing observed.
2. **My first `go-work-sync` diagnosis was wrong.** I hypothesised a stale BuildFlow **result cache** and wrote a (draft) AGENTS.md note to that effect. The single-step run with `BUILDFLOW_NO_RESULT_CACHE=1` showed no finding, which "confirmed" the hypothesis — but the next full run showed it again, proving it was a **real** (if benign) finding from a source-code limitation I only understood after reading `tools/gomod/module.go`. Lesson: a cache-bypass single-step run is not sufficient evidence that a full-pipeline finding is cached; two different code paths were involved. The wrong hypothesis was corrected before committing.
3. **`.golangci.yml` still carries four `goexperiment.*` build tags (`arenas`, `goroutineleakprofile`, `runtimesecret`, `simd`) whose validity on Go 1.27 I did not verify.** I removed only `goexperiment.jsonv2` because I proved it graduated; the others may have graduated too, making them dead config. Severity: low (harmless if unknown), but it is the same class of stale-experiment rot I was fixing, left half-done.
4. **Nothing catastrophic is broken in the product code** — but the honest read is that this session was largely **recovery from an incomplete Go 1.27 migration**, and the migration's other surfaces (docs, CI counts, remaining experiment tags) are only partially reconciled.

---

## e) WHAT WE SHOULD IMPROVE

1. **Detect concurrent sessions early.** The buildflow preflight already prints the binary-freshness / shared-checkout hint; I ignored its significance until an edit collision forced the issue. Improvement: at session start, run `git log --oneline -5` and compare HEAD to my last-known HEAD before the first edit; treat rapid auto-commits as a live-session signal.
2. **Never infer "cached finding" from one bypass run.** Improvement: to prove cache-staleness, reproduce the finding in the *same* pipeline path (full run) with the cache disabled, not in a different single-step path.
3. **Keep the Go-version pin single-source.** It currently lives in `go.mod`, `flake.nix` (`goPkg`), and `.golangci.yml` (`run.go`) — three places that drifted (the exact failure this session). Improvement: derive `goPkg` and the lint `go:` from a shared Nix `let` reading `go.mod`, or at minimum add a buildflow/CI check that the three agree.
4. **Sweep experiments coherently.** When removing one graduated `goexperiment.*`, audit the sibling tags in the same pass (see (d)3).
5. **Add a CI regression guard for the GOEXPERIMENT removal.** A step that builds with no `GOEXPERIMENT` set would have caught the flake/treefmt divergence that broke this run.
6. **Reconcile `FEATURES.md` counts.** It still says BuildFlow "passes 35/36"; the pipeline now reports 56/64. Stale claims in a status doc are worse than no claim.
7. **Route section (f) via HARVEST.** Otherwise these 50 items die in a timestamped file.

---

## f) Up to 50 things we should get done next

Ranked by impact (Critical → Low). Effort S <30m, M 30m–2h, L >2h.

| # | Task | Impact | Effort | Category |
| --- | --- | --- | --- | --- |
| 1 | Delete the local, gitignored `go.work` (or intentionally keep it) to stop the `go-work-sync` false positive | High | S | Cleanup |
| 2 | Add a CI step that builds with **no** `GOEXPERIMENT` set (regression guard for the jsonv2 removal) | High | S | Quality |
| 3 | Audit the remaining `.golangci.yml` `goexperiment.*` tags (`arenas`, `goroutineleakprofile`, `runtimesecret`, `simd`) against Go 1.27 | High | S | Cleanup |
| 4 | Run `actionlint` locally and confirm the edited `ci.yml` parses | High | S | Quality |
| 5 | Rebuild + reinstall the BuildFlow binary (`nix build . && nix run .#reinstall`) so dogfooding is not stale | High | M | Quality |
| 6 | Refresh `TODO_LIST.md` by harvesting this report | High | M | Documentation |
| 7 | Split `registry_test.go` (1135 lines) into focused files | High | M | Cleanup |
| 8 | Reduce `registry.go` (371 lines) below the 350-line threshold | Medium | M | Cleanup |
| 9 | Reduce `examples/oops-redundant-guard/main.go` (367 lines) below 350 | Medium | M | Cleanup |
| 10 | Extract duplication in `examples_integration_test.go` (22 art-dupl findings) | Medium | M | Cleanup |
| 11 | Decide on `branching-flow`'s strong-ID suggestions (`RuleID` type) — implement or document as deliberate | Medium | M | Feature |
| 12 | Refresh `README.md` (576h stale) | Medium | M | Documentation |
| 13 | Reconcile "35/36" (FEATURES.md) vs actual "56/64" BuildFlow counts | Medium | S | Documentation |
| 14 | Unset the global `GOEXPERIMENT=jsonv2` in the user's shell to clear the buildflow preflight warning | Medium | S | Quality |
| 15 | Merge the two-author `AGENTS.md` jsonv2 section into one voice | Medium | S | Documentation |
| 16 | Merge the two CHANGELOG entries for the same change into one narrative bullet | Low | S | Documentation |
| 17 | Add an `AGENTS.md` note about the shared-checkout / concurrent-session hazard | Medium | S | Documentation |
| 18 | Verify `FEATURES.md` inventory is complete and current (768h stale) | Medium | M | Documentation |
| 19 | Add a docs/migration-guide pitfall for the Go-floor vs flake-pin drift | Medium | S | Documentation |
| 20 | Add `securitymd` green status to `FEATURES.md` | Low | S | Documentation |
| 21 | Verify CI goes green on the post-session commit | High | S | Quality |
| 22 | Run `nix run .#test-race` explicitly and record the result | Medium | S | Quality |
| 23 | Run `nix run .#coverage` and record the exact library coverage | Low | S | Quality |
| 24 | Consider a single-source Go version (derive `goPkg`/lint `go:` from `go.mod`) | Medium | M | Cleanup |
| 25 | Investigate `govulncheck` reporting "requires go1.27" (toolchain of the nixpkgs binary) | Low | M | Quality |
| 26 | Triage `vulnix` CVEs; decide whether a `nix flake update` is warranted | Medium | M | Security |
| 27 | Confirm `SECURITY.md`'s 48h commitment is acceptable to the maintainer | Medium | S | Documentation |
| 28 | Align `SUPPORT.md` response-time wording with `SECURITY.md` | Low | S | Documentation |
| 29 | Re-run the full buildflow once more after all doc edits and capture a clean log | Medium | S | Quality |
| 30 | Verify the `dprint-format` step's markdown rewrites are idempotent (it rewrote AGENTS/FEATURES) | Low | S | Quality |
| 31 | Add `docs/status` HTML from the concurrent session to the review set | Low | S | Documentation |
| 32 | Review the concurrent session's `docs/status/2026-10-11_06-56_*.html` for overlap | Medium | S | Documentation |
| 33 | Ensure `FEATURES.md` "Tooling & infrastructure" row reflects `go_1_27` | Medium | S | Documentation |
| 34 | Add a test asserting `RuleError` carries the rule ID (regression for error wrapping) | Low | M | Quality |
| 35 | Consider raising the file-size threshold or splitting the largest files systematically | Low | M | Cleanup |
| 36 | Record the treefmt `goimports` failure mode in the failure-triage reference (if fleet-wide) | Medium | M | Quality |
| 37 | File/build a BuildFlow improvement: `go-work-sync` should ignore out-of-root `use` dirs | Medium | M | Feature |
| 38 | Add a `GOEXPERIMENT=` vs `GOEXPERIMENT=none` nuance to `CONTRIBUTING.md` | Low | S | Documentation |
| 39 | Check whether other LarsArtmann repos still pin `go_1_26` with a 1.27 floor | Medium | M | Cleanup |
| 40 | Add an `AGENTS.md` "known BuildFlow false positives" section | Low | S | Documentation |
| 41 | Re-run `buildflow -s go-structure-linter --format finding` to confirm only the internal/ warning remains | Low | S | Quality |
| 42 | Decide `internal/` handling: accept warning vs suppress in `.buildflow.yml` | Medium | S | Cleanup |
| 43 | Verify `go mod tidy` is clean with `GOWORK=off` (post-removal) | Medium | S | Quality |
| 44 | Add the session's evidence commands to `CHANGELOG.md` release notes | Low | S | Documentation |
| 45 | Update `README.md` "Companion libraries" if sibling import paths changed | Low | S | Documentation |
| 46 | Spot-check that no doc still claims `GOEXPERIMENT` is required | High | S | Documentation |
| 47 | Add a `--strict` dry run to see what warning-severity promotion would surface | Low | S | Quality |
| 48 | Consider whether `gofumpt`/`golines` should also be pinned to `goPkg` for parity | Low | S | Cleanup |
| 49 | Archive the pre-session failing log for future triage reference | Low | S | Quality |
| 50 | Set up docs-health HARVEST for (f) into TODO_LIST/ROADMAP | High | M | Documentation |

> **HARVEST handoff:** items 1–50 belong in `TODO_LIST.md` (actionable, bounded) and `ROADMAP.md` (the larger refactors: 7–11, 24, 37). Do not let them die in this file.

---

## g) Questions I cannot answer myself

1. **Is the concurrent Crush session on this repo still active, and should I coordinate with it?** I detected interleaved auto-commits (`109312c`, `6c57873`, `38e603b`, `417e45a`) and a colliding `AGENTS.md` write, but I cannot see the other session's intent. If it is still running, two agents editing the same docs will keep colliding.
2. **Do you want the local `go.work` kept (sibling cross-repo dev, permanent warning) or deleted (clean buildflow)?** The choice is a workflow preference I cannot infer — the buildflow finding is a false positive either way.
3. **Is the 48-hour initial-response commitment I added to `SECURITY.md` acceptable, or should it be a softer best-effort phrasing?** `securitymd` demands *a* commitment; only you know the real support posture.

---

## Evidence appendix (commands + observed results)

```
buildflow --fix --build-mode=full --budget 5m --log-level warn --max-time 5m   → exit 0
   "⚠ BuildFlow passed with warnings 56/64 4.2s (83 findings)"
nix build .#checks.x86_64-linux.treefmt                                        → success (was: exit 1)
nix flake check --no-build                                                     → "all checks passed!"
nix run .#lint                                                                 → "0 issues."
GOWORK=off GOEXPERIMENT= go test ./...                                         → ok
GOWORK=off GOEXPERIMENT= go build -a ./...                                     → success
GOEXPERIMENT= go list std | grep encoding/json                                 → json/v2 + jsontext present
GOWORK=off GOEXPERIMENT=none go build ./...                                    → FAILS (build constraints exclude jsontext)
buildflow -s go-version-auto-configure --format finding                        → no findings
buildflow -s go-work-sync --format finding                                     → 1 warning (stale ../go-finding)
buildflow -s securitymd --format finding                                       → no findings
```

## Files touched this session

`flake.nix`, `.buildflow.yml`, `.golangci.yml`, `.github/workflows/ci.yml`, `SECURITY.md`, `AGENTS.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `README.md`, `FEATURES.md`, `examples/{minimal-linter,naked-return-guard,no-go-mod,oops-redundant-guard}/main.go`.
