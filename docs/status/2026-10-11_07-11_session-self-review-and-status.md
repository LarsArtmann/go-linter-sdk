# Status Report — Session Self-Review: Ecosystem Adoption & Doc-Drift

- **Date:** 2026-10-11 07:11 CEST
- **Scope:** The single session that reviewed the project, produced a status
  report, harvested follow-ups, and repaired living-doc drift — everything
  below is grounded in *this session's* run and what it noticed.
- **Format note:** the user explicitly requested a `.md` path, so this is
  Markdown, not the skill-canonical HTML. Explicit instruction wins; not
  propagated back as a new default. (The companion HTML dashboard from the same
  session is `docs/status/2026-10-11_06-56_ecosystem-adoption-and-doc-drift-status.html`.)

---

## Trigger

The session began from pasted `project-dependency-graph` output showing
consumers of `go-finding`, `go-finding/toolsdk`, `go-linter-sdk`, and
`linter-autoconfigure-sdk` across `~/projects`. The ask was "Review + reports +
insights", executed as a full status report plus drift repair.

---

## a) FULLY DONE

| # | Item | Evidence |
| - | ---- | -------- |
| a1 | **Gate verified on the working tree** | `go build`, `go vet`, `go test`, `go test -race` green; `golangci-lint` 0 issues; `GOWORK=off go mod tidy` no drift; library coverage **98.7%** |
| a2 | **Dependency currency established** | Proxy check: `go-finding` root latest is `v1.14.0` = the pinned version; no bump needed |
| a3 | **Consumer data independently confirmed** | Recursive `go.mod` grep + graph output: 6 direct consumers at `v0.3.1` (`go-structure-linter`, `samber-linter`, `go-humanize-linter`, `cmdguard/lint`, `project-meta/linter`, `BuildFlow/execution`), ~60 indirect |
| a4 | **HTML status report written** | `docs/status/2026-10-11_06-56_ecosystem-adoption-and-doc-drift-status.html` (a–g + Top 25 + ecosystem insights) |
| a5 | **Drift repaired in AGENTS.md** | Dep block → `v1.14.0` / Go 1.27; "lockstep" claim corrected; consumers section rewritten with the real 6 |
| a6 | **Drift repaired in README.md** | Status callout, consumer list, "Requires Go 1.27+", status section |
| a7 | **Drift repaired in TODO_LIST.md** | Dependency line refreshed; harvested items #10–#18 added |
| a8 | **Drift repaired in ROADMAP.md** | Theme 2 adoption restated; Q1 `v1.4.1` → `v1.14.0` |
| a9 | **Drift repaired in DOMAIN_LANGUAGE.md** | Header → v1.14.0 and content diff-verified against the module cache (`Finding.Edits`/`TextEdit`, `Report.SkippedModules` added) |
| a10 | **CHANGELOG backfilled** | v1.10.0 → v1.14.0 bump + Go 1.26 → 1.27 toolchain entries added under `[Unreleased]` |
| a11 | **FEATURES + migration-guide + .buildflow refreshed** | Coverage re-verified note; Go 1.27; comment de-drifted |
| a12 | **GOEXPERIMENT obsolescence discovered and corrected** | Empirically proven `encoding/json/v2` is default-on in Go 1.27; AGENTS/FEATURES/migration-guide/report updated |
| a13 | **"self" flake gotcha re-verified** | `nix flake check --no-build` passed without an explicit `self`; gotcha softened to a dated note |

---

## b) PARTIALLY DONE

| # | Item | What works | What remains | Blocker | Effort |
| - | ---- | ---------- | ------------ | ------- | ------ |
| b1 | **Post-change `nix flake check`** | Ran green *before* the flake was edited | A concurrent session then removed `GOEXPERIMENT` from the flake/devShells; I did **not** re-run `nix flake check` after that edit | Concurrent session edits | S |
| b2 | **GOEXPERIMENT cleanup** | Proven obsolete; I fixed AGENTS/FEATURES/migration-guide | Concurrent session owns flake/CI/golangci/examples; the repo may still mix updated and un-updated spots | Shared repo state | S |
| b3 | **Report accuracy after the pivot** | Added a GOEXPERIMENT insight card to the HTML | The HTML was written *before* the GOEXPERIMENT discovery; its sections a–f still under-represent it | None | S |
| b4 | **CHANGELOG depth** | v1.11–v1.14 recorded as one entry | Individual release notes for v1.11.0/v1.12.0/v1.13.0/v1.14.0 not enumerated | Time | M |
| b5 | **Consumer-adoption capture** | 6 consumers documented in AGENTS + README | No durable, refreshable adoption doc (harvested as #11, not built) | None | S |
| b6 | **Verification coverage** | Library tested at 98.7%, race-clean | Benchmarks not re-run; `nix run .#test` (the actual app entrypoint) not invoked, only `go test` directly | None | S |

---

## c) NOT STARTED

| # | Item | Why not started | Still wanted? |
| - | ---- | --------------- | ------------- |
| c1 | **Converter-deletion verification** | Out of scope this session; needs per-consumer inspection | Yes — it is the SDK's core claim |
| c2 | **Doc-drift CI guard** (harvested #12) | Reported, not implemented | Yes |
| c3 | **Next release tag** (harvested #14) | No tag this session | Yes |
| c4 | **Benchmark re-run** (harvested #15) | Forgotten — the exact recurring omission | Yes |
| c5 | **`gh pr list` dependabot sweep** (harvested #16) | Not run | Yes |
| c6 | **Flake `checks` for test/vet/lint** (harvested #17) | Roadmap-only | Yes |
| c7 | **ROADMAP Q2/Q3 resolution** | Out of scope | Deferred |

---

## d) TOTALLY FUCKED UP!

Radical honesty on this session's own failures.

| # | Failure | Severity | Root cause | Mitigation |
| - | ------- | -------- | ---------- | ---------- |
| d1 | **Did not detect the concurrent session for most of the run.** Another session was stripping `GOEXPERIMENT` from ci.yml/flake/golangci/examples while I edited the same repo. I only noticed when `git status` showed files I never touched. | High | I checked `git status` once at the start (saw only `flake.nix`) and never re-checked before editing other files | Re-`git status` before each edit phase; treat unexpected diffs as a signal to re-read the tree |
| d2 | **My first verification was contaminated.** I ran the gate with `GOEXPERIMENT=jsonv2` exported, so "green" did not prove the CI-equivalent path. The fresh-environment check came minutes later. | High | Copied the repo's documented ritual instead of testing the actual build contract | Always run the gate in an environment matching CI (`env -i PATH HOME`) |
| d3 | **Wrote the status report prematurely.** The HTML was finished before the GOEXPERIMENT discovery, so its central narrative missed a same-session structural change. | Medium | Treated "report written" as terminal instead of a living draft until the session ended | Keep the report open until the session's findings stabilize |
| d4 | **Edited files a parallel agent was also editing.** Real clobber risk on `.buildflow.yml` (my edit was overwritten by theirs). | Medium | Same as d1 — no concurrency awareness | Detect and coordinate (or pause) when the tree is mutating under me |
| d5 | **Changed the same fact in many places by hand.** I "fixed" the drift by rewriting numbers in 6+ files — reproducing the very drift mechanism I diagnosed. | Medium | Followed the drift's own shape instead of the durable fix (point at `go.mod`) | Encode "see `go.mod`" everywhere; never re-hardcode versions |
| d6 | **AGENTS coverage number was wrong and I only half-fixed earlier numbers.** It said `./... 83.3% / lib 98.2%`; real values are `44.3% / 98.7%`. | Low | Stale claims trusted until re-measured | Re-measure every number before rewriting a doc that contains it |

---

## e) WHAT WE SHOULD IMPROVE

| # | Pattern | Impact | Concrete fix |
| - | ------- | ------ | ------------ |
| e1 | **No concurrency awareness in a mutating shared repo** | Clobber risk, wasted work, false "done" | Re-run `git status` before every edit phase; if the tree changed, re-read the touched files first |
| e2 | **Verification runs in the wrong environment** | Green locally, red in CI | Standardize on `env -i PATH=$PATH HOME=$HOME` (or `nix run .#*`) for every gate |
| e3 | **Report written once, then the world changed** | Snapshot lies to the next reader | Treat the report as draft until session end; one final reconciliation pass |
| e4 | **Prose duplicates `go.mod`** | The drift class that produced this whole session | Replace every hardcoded version with a pointer; add CI enforcement |
| e5 | **"Forgot benchmarks" recurred** (again) | Dep bumps land without perf baselines | Make the bump checklist a script, not prose |
| e6 | **Recommended a guard but shipped no guard** | The fix stays theoretical | Implement the doc-drift check now, not next time |
| e7 | **No HTML validation** | A malformed tag could break the dashboard silently | Add a trivial well-formedness check (or open it) |
| e8 | **Consumers documented, deletion unverified** | The core value claim stays unproven | Inspect one consumer end-to-end next session |

---

## f) Up to 50 Things We Should Get Done Next

Sorted by impact. Category: Bug / Feature / Quality / Cleanup / Docs / Process.

| # | Task | Impact | Effort | Category |
| - | ---- | ------ | ------ | -------- |
| 1 | Verify whether any consumer deleted its converter package | Critical | M | Feature |
| 2 | Add a CI guard that fails when a living doc's version disagrees with `go.mod` | Critical | M | Quality |
| 3 | Re-run `nix flake check` after the concurrent flake edits | High | S | Quality |
| 4 | Run `nix run .#test`/`#lint` (the real app entrypoints), not just `go test` | High | S | Quality |
| 5 | Reconcile the HTML report with the GOEXPERIMENT discovery end-to-end | High | S | Docs |
| 6 | Cut the next release after green CI (5 bumps + flake fix unshipped) | High | M | Release |
| 7 | Re-run benchmarks; compare to `reports/benchmarks-baseline.txt` | High | S | Quality |
| 8 | Convert remaining prose version references to "see `go.mod`" | High | M | Docs |
| 9 | Confirm the concurrent GOEXPERIMENT cleanup left no mixed state (grep whole tree) | High | S | Cleanup |
| 10 | Build a durable, refreshable consumer-adoption doc | Medium | S | Docs |
| 11 | Enumerate CHANGELOG entries per release (v1.11–v1.14) | Medium | M | Docs |
| 12 | Run `gh pr list`; close superseded dependabot PRs | Medium | S | Cleanup |
| 13 | Add flake `checks` for `go test`, `go vet`, `golangci-lint` | Medium | M | Quality |
| 14 | Turn the bump checklist into an executable script | Medium | S | Process |
| 15 | Grep CI workflows for stale version pins docs sync missed | Medium | S | Cleanup |
| 16 | Reconcile CONTRIBUTING.md with the final dependency story | Medium | S | Docs |
| 17 | Resolve ROADMAP Q2 (root package vs `internal/`) | Medium | M | Design |
| 18 | Resolve ROADMAP Q3 (library-only vs eventual CLI) | Low | S | Design |
| 19 | Validate the HTML report renders (well-formedness check) | Medium | S | Quality |
| 20 | Add an adoption smoke test to the CI stranger-test job | Medium | M | Quality |
| 21 | Record the `toolsdk` independent-versioning reality in AGENTS (done) and re-verify at next bump | Low | S | Docs |
| 22 | Re-verify the `self` flake gotcha on the edited flake | Medium | S | Quality |
| 23 | Confirm `GOEXPERIMENT=none` is never set anywhere (CI/flake/env files) | Medium | S | Cleanup |
| 24 | Re-measure and refresh all coverage numbers in docs | Medium | S | Docs |
| 25 | Add a "concurrency check" step: `git status` before each edit phase | Medium | S | Process |
| 26 | Document the auto-commit daemon's interaction with parallel sessions | Low | S | Docs |
| 27 | Verify pkg.go.dev renders the current README/`@latest` | Low | S | Docs |
| 28 | Check the "Imported by" count on pkg.go.dev for adoption signal | Low | S | Docs |
| 29 | Add a docs-health HARVEST pass that pulls f) into TODO_LIST/ROADMAP | High | S | Process |
| 30 | Confirm the CI tidy check still passes with the current go.sum | Medium | S | Quality |
| 31 | Sweep examples' doc comments consistent with the GOEXPERIMENT removal | Low | S | Cleanup |
| 32 | Ensure `.buildflow.yml` matches the final GOEXPERIMENT posture | Medium | S | Cleanup |
| 33 | Add a benchmark step to CI (or a documented manual gate) | Medium | M | Quality |
| 34 | Archive superseded status reports per docs-health policy | Low | S | Docs |
| 35 | Annotate this session's HTML report where the GOEXPERIMENT pivot supersedes it | Medium | S | Docs |
| 36 | Verify `go mod verify` still green after all doc changes | Low | S | Quality |
| 37 | Add a short "how to verify the gate" recipe to CONTRIBUTING | Medium | S | Docs |
| 38 | Confirm the flake apps run without any exported GOEXPERIMENT | Medium | S | Quality |
| 39 | Decide whether jsonv2 wording should be "standard" vs "default-on experiment" | Low | S | Docs |
| 40 | Track adoption as a metric with a refresh cadence | Low | S | Process |
| 41 | Add a release cadence note (when to tag docs-only vs API changes) | Low | S | Docs |
| 42 | Confirm no living doc still claims "no consumers" (grep) | Medium | S | Cleanup |
| 43 | Confirm no living doc still claims Go 1.26 (grep) | Medium | S | Cleanup |
| 44 | Re-run the full gate one final time at session end and record it | High | S | Quality |
| 45 | Note the parallel-session hazard in the project's AGENTS.md | Medium | S | Docs |
| 46 | Cross-check `docs/migration-guide.md` pitfalls against the workspace-tidy lesson | Low | S | Docs |
| 47 | Add a lint/CI assertion that `go.work`/`go.work.sum` stay gitignored | Low | S | Quality |
| 48 | Re-verify coverage gate (≥90% library-only) still enforced in CI | Medium | S | Quality |
| 49 | Summarize the version-distribution health (60/31/3/1) as a tracked metric | Low | S | Docs |
| 50 | Book a follow-up session dedicated to converter-deletion evidence | High | M | Feature |

---

## g) Questions I CANNOT answer myself

1. **Did any of the six consumers actually delete its converter package — and
   is one pilot rule enough to call a consumer "migrated", or does "done"
   require the full port?** I found the importers but defining the migration
   bar is a product decision only you can make.

2. **Is the concurrent session I detected intentional and still running, or was
   it a one-off?** If it is ongoing, I need to coordinate (or stand down) on
   AGENTS.md/flake/CI rather than edit the same files. I cannot see the other
   session's intent.

3. **For the next release: do you want a docs/hygiene PATCH (v0.3.2) now, or
   hold for real API movement (v0.4.0)?** Your recorded 2026-09-09 rule said no
   v0.4.0 without Go-API movement, but five dependency/Go bumps have since
   landed untagged — the choice of version label is yours.

---

_Reported 2026-10-11 07:11 · grounded in this session only · no new research performed_
