# Status Report — 2026-10-15 18:47: v0.4.0 bump wave, pkg.go.dev verification, and the cmdguard guard/dprint split-brain fix

Scope: the session that resumed after the 2026-10-11 release session (`d3be415`). It closed TODO #19, recorded a silent
dependency bump, took three user decisions, and executed the consumer bump wave to v0.4.0 across six repos — including a
root-cause fix for a guard-vs-formatter split brain in cmdguard. One concurrent DeepSeek session remained active in this
repo and in BuildFlow throughout.

Headline state: **go-linter-sdk master at `7ebfa21`, pushed, CI green (both workflows). All six direct consumers now pin
v0.4.0.** Two consumer repos carry unpushed commits (no push authorization). One pre-existing red in samber-linter is
bisect-proven unrelated to the SDK.

---

## a) FULLY DONE

| # | Item | Evidence |
| - | ---- | -------- |
| 1 | **TODO #19 closed: pkg.go.dev renders v0.4.0.** Verified mechanically from the page's raw HTML (no summarizer): breadcrumb `go-linter-sdk@v0.4.0`, `Version: v0.4.0` + `Latest` chip, `Published: Oct 11, 2026`, README rendered from the v0.4.0 tag. No "Request a version" escalation needed. | TODO_LIST.md Release section rewritten to "None open"; 2026-10-11 status report annotated inline (`RESOLVED 2026-10-15`); commit `6dc7ebf` |
| 2 | **Silent post-tag dependency bump given a narrative.** Master's `go-finding` v1.14.0 → v1.15.0 (transitive `go-error-family` v0.11.0 → v0.12.0) had landed in auto-commit `154c20a` with zero documentation anywhere (proven: repo-wide grep + `git log -S`). Recorded in CHANGELOG `[Unreleased]` with upstream content verified from go-finding's own CHANGELOG via GitHub API (source-compatible `DetectorV2` coverage channel; `Detector` unchanged). | CHANGELOG.md; doc guard green |
| 3 | **Three user decisions taken and recorded in AGENTS.md** (2026-10-15 block): (1) version policy = keep `0.x` prefix, MINOR-vs-PATCH slot nuance delegated, v0.4.0 stands; (2) concurrent release races accepted as status quo; (3) consumer bump wave approved and executed. | AGENTS.md "User decisions recorded 2026-10-15" |
| 4 | **Bump classified additive at source, not from CHANGELOGs.** Exported-symbol diff v0.3.1→v0.4.0: three new `Category` constants + tests/benchmarks only; the 8 non-test lines are a `//nolint` rename and a doc comment. Published v0.4.0 go.mod inspected from the module cache (`go 1.27`, `go-finding v1.14.0`). | `git diff v0.3.1..v0.4.0` symbol grep; `@v/v0.4.0.mod` |
| 5 | **All six consumers enumerated with hazard sweep** (full-depth `find`, go.work/vendor/replace checks, dirty-tree checks). Discovered the ecosystem had moved since 2026-10-11: four of six already at v0.4.0 via their own sessions. | Sweep output; ADOPTION.md 2026-10-15 snapshot |
| 6 | **cmdguard/lint bumped v0.3.1 → v0.4.0** with go.work hygiene (renamed during `go get`, restored after), `GOWORK=off` throughout, version grep-verified post-mutation. Baseline recorded first (F11): build green, one pre-existing red. | cmdguard `8fd9f26` (daemon), lint/go.mod |
| 7 | **project-meta/linter bumped v0.3.1 → v0.4.0**, same hygiene. Pre-existing untidy go.mod found at baseline and tidied (aligned transitive go-error-family to v0.12.0 + go.sum hashes) before bumping, so the baseline was measurable. | project-meta `5e527d5` (daemon), linter/go.mod |
| 8 | **Phase-4 verification on both bumped modules:** `go build`, `go test ./...`, `go test -race ./...`, `go mod verify` — all green; failing-test count compared against baseline (cmdguard: same single pre-existing red, no new failures). | Session gate runs |
| 9 | **cmdguard's pre-existing red closed:** `TestReadmeRuleTableInSync` (README rule table drifted from `allRuleDefs()`). Regenerated with the module's own command; committed `2e1c122`. | cmdguard `2e1c122` |
| 10 | **cmdguard guard/dprint split-brain fixed at root cause** (the deep item of the session — see d1 for how it bit): `RuleTableMarkdown()` now emits dprint's exact padded table format (byte-equality proven against dprint's own output), and the drift guard compares trimmed content so dprint's ownership of surrounding blank lines stops breaking the byte-for-byte pin. All three gates — guard test, `dprint check`, `go vet` — green simultaneously, full lint suite green at HEAD. | cmdguard `752a5fc` (daemon-swept), lint/rules_table.go + rules_table_test.go |
| 11 | **samber-linter failure classified by direct bisect, not speculation:** `TestGoWorkMultiModule` (internal/driver, HW-1 fixture) fails **identically at v0.3.1** → pre-existing, not v0.4.0-caused. Restored to v0.4.0, tree clean. | Bisect run in session; ADOPTION.md snapshot |
| 12 | **Wave evidence written where it lives:** ADOPTION.md 2026-10-15 snapshot (wave execution, additive proof, samber bisect verdict, BuildFlow skip rationale, "Imported by: 3" re-verified unchanged from the raw pkg.go.dev HTML); AGENTS.md Consumers table updated to v0.4.0. | ADOPTION.md; AGENTS.md; commit `7ebfa21` |
| 13 | **Everything committed; go-linter-sdk pushed and CI-green at HEAD `7ebfa21`** (CI + Dependency Graph, success). Doc guard green after every doc edit (it validates CHANGELOG version mentions against go.mod — caught nothing stale after the `[Unreleased]` entry). | `gh run list`; `scripts/check-doc-versions.sh` |

## b) PARTIALLY DONE

1. **BuildFlow verification (the wave's weak leg).** All three BuildFlow modules already pin v0.4.0 (grep-verified), but
   local gates were skipped deliberately: its working tree carried an uncommitted local `replace` on go-linter-sdk, a
   dirty root go.mod mid-submodule-bump (2026-10-14 pseudo-versions), and an active concurrent session. Local runs would
   have measured the replace target, not the tag. Verification is delegated to BuildFlow's own CI — which I did not wait
   for. The skip is documented in ADOPTION.md, but "verified" and "documented-why-not-verified" are different claims and
   the ADOPTION table does not visually distinguish them.
2. **ADOPTION refresh procedure executed 3 of 4 steps.** Step 1 (go.mod grep) ✓, step 3 (table/date update) ✓, step 4
   (imported-by + sibling scan) partially ("Imported by" only; no fresh dependency-graph scan for NEW consumers).
   Step 2 — the converter-deletion source re-scan across all six repos — was **not** re-run this session; the
   "converter deletion verified" claim still rests on the 2026-10-11 evidence.
3. **samber-linter: classified but not fixed.** The red is bisect-proven pre-existing, but the actual cause (HW-1 rule
   not firing in the go.work fixture — "no health-washing found") was not diagnosed. Declining to fix was the right
   scope call for a wave session, but the repo's master is red right now and the wave report merely notes it.
4. **Pushes: 1 of 3 repos.** go-linter-sdk pushed (per its standing end-of-session policy). cmdguard (`2e1c122`,
   `752a5fc`) and project-meta (`5e527d5`) commits sit local-only — the global no-push-without-ask rule was honored, but
   it means the wave is not actually landed for CI/consumers of those repos.
5. **CHANGELOG `[Unreleased]` exists with no release vehicle.** The dependency-bump entry is written; whether it ships as
   v0.4.1 (PATCH: dependency + refreshed pkg.go.dev README) or accumulates into the next MINOR is an open decision that
   was never put to the user.
6. **Master README is again ahead of the frozen pkg.go.dev README** (the concurrent session's 11-constants and new
   Errors-section fixes are good but won't render on pkg.go.dev until the next tag). Known freeze mechanic, but worth
   tracking as recurring drift.

## c) NOT STARTED

1. samber-linter `TestGoWorkMultiModule` root-cause diagnosis and fix (or issue filing with source-level diagnosis per
   verify-before-filing).
2. Pushing the cmdguard and project-meta commits.
3. The release decision/cut for the `[Unreleased]` dependency bump.
4. Converter-deletion re-scan (ADOPTION refresh step 2).
5. Harvesting this report's section (f) into TODO_LIST.md / ROADMAP.md (docs-health HARVEST) — deliberately deferred:
   the user's instruction was write-then-wait, and most items below are brainstorm-grade, not commitments.
6. Scanning for NEW direct consumers since 2026-10-11 (dependency-graph sweep of `~/projects`).
7. go-finding version-distribution re-scan (the 60/31/4 metric is a 2026-10-11 snapshot).
8. Whether branching-flow and erraudit (the pilot-rule donors) now import the SDK — never checked this session.
9. Benchmark re-run against the 2026-10-11 baseline after the go-finding v1.15.0 bump (no regression expected from a
   source-compatible bump, but "expected" is not "measured").
10. Recording the "re-run the gate after a formatter mutates the artifact" lesson in the fleet's lessons file.

## d) TOTALLY FUCKED UP

1. **A commit message that was false within minutes of being written (repeated offense pattern).** cmdguard `2e1c122`
   claims "the full lint suite is green" — true when I ran it, false at HEAD, because the pre-commit hook's dprint
   auto-fix rewrote the README after my verification and re-broke the byte-for-byte guard. The 2026-10-11 report already
   carried this exact sin ("claimed v0.4.0 verified while neither was true yet"). I verified the splice, then allowed a
   tool to mutate the verified artifact and asserted anyway. Caught it on the very next command and fixed forward
   (`752a5fc`), but the pattern — verify, then let something else touch the thing, then keep the claim — is now a
   two-session habit, which makes it mine.
2. **I misread my own gate once.** The first cmdguard baseline run echoed `exit: 0` while the suite was FAILING — `tail`
   had swallowed the pipeline status and I reported the pipe's exit code as the suite's. Caught one command later when I
   grepped the failure detail, and all later gates used explicit output matching or PIPESTATUS, but for one command the
   session's own telemetry lied to me because of my sloppy shell.
3. **An unrelated 1,537-line generated artifact shipped inside my fix commit.** The daemon's `752a5fc` bundled my staged
   rules_table changes with `website/src/styles/global.out.css` (tailwind build output from the 2-minute pre-commit
   hook). I had a ~2-minute window between finishing the fix and the commit attempt; the race cost the commit itself
   (fatal: cannot lock ref) and then the sweep attributed css noise to my change. Not data loss, and the daemon
   heuristic-messages are accepted practice here — but a generated artifact got a free ride into history because I
   didn't commit narrowly enough, fast enough.
4. **project-meta's one real bump has no narrative commit.** The daemon's `chore: auto-commit 2 changed file(s)` is the
   entire git record of the project-meta bump; the narrative lives only in this repo's ADOPTION.md. Accepted for this
   fleet, but it means that repo's history cannot explain its own go.mod.
5. **The wave's biggest discovery — "4 of 6 consumers had already moved" — means the 2026-10-11 adoption table was
   already stale by execution time.** Not this session's failure, but this session validated it: version-surface claims
   in adoption docs decay within days when six repos run concurrent sessions. Any consumer statement in ADOPTION.md is
   only as old as its last verification date.

## e) WHAT WE SHOULD IMPROVE

1. **Close the mutate-after-verify hole.** Any time a hook, formatter, or generator rewrites a file after the verifying
   command ran, the gate must be re-run before any claim (chat or commit message) stands. Candidate enforcement: a
   personal rule plus a BuildFlow-side "post-hook gate re-run" idea — worth filing upstream rather than memorizing.
2. **Commit messages are verified claims too.** Same gate as chat claims: don't write "green" in a commit message unless
   the verifying command exited 0 against the exact tree being committed (post-hook).
3. **Never read `$?` after a pipeline** — use `PIPESTATUS`/`set -o pipefail`/explicit grep-of-output for gates. One bad
   reading nearly propagated a false baseline.
4. **Check `git log -3 && git status` in EVERY repo immediately before committing.** Six repos × daemons × concurrent
   sessions = races are the default. The cmdguard lock failure and the css bundle were both foreseeable.
5. **Bisect-first for "is it the bump?"** The samber two-command bisect settled in seconds what speculation would have
   stretched into minutes. Make it the reflexive move for any consumer failure during a version wave.
6. **Execute checklists fully, visibly.** ADOPTION's refresh procedure has numbered steps; I did 1, 3, ¾ of 4 and
   skipped 2 without flagging the skip until writing this report. Skipped steps in a known procedure should be called
   out at execution time, not discovered at write-up.
7. **Narrative at bump time, not audit time.** The v1.15.0 bump sat four days with no record. Any detected
   version-surface change gets its CHANGELOG/doc line in the same session it is noticed.
8. **Record-and-skip should leave a note IN the skipped repo.** BuildFlow's tree got a mention here, but nothing in
   BuildFlow tells the next session why a stranger declined to run its gates.
9. **The question-tool batch at the moment of decision worked well** (3 policy decisions in one round, zero
   back-and-forth). Use it at decision time instead of parking questions in reports.
10. **Update the adoption table format to carry verification *mode*, not just date** (e.g. "grep-only" vs
    "grep+build+test") so partial verification like BuildFlow's is visible without reading prose.

## f) Up to 50 things we should get done next

Brainstorm, sorted roughly by impact — items 1–12 are the real queue; the rest is ROADMAP fuel and needs HARVEST routing.

| #  | Task                                                                                                                                       | Impact | Size |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ---- |
| 1  | Push cmdguard (`2e1c122`, `752a5fc`) and project-meta (`5e527d5`) to origin; confirm their CI goes green                                     | High   | S    |
| 2  | Diagnose samber-linter `TestGoWorkMultiModule` (HW-1 fixture) at source; fix in samber-linter or file with verified diagnosis                | High   | M    |
| 3  | Decide the release vehicle for CHANGELOG `[Unreleased]` (v0.4.1 PATCH now vs accumulate) and cut it per go-release (tag after green CI)      | High   | M    |
| 4  | Run ADOPTION refresh step 2 fully: converter-deletion re-scan across all six consumers, refresh verdicts                                     | High   | M    |
| 5  | Post-release: re-verify pkg.go.dev renders the new tag with the refreshed README (frozen-README habit)                                       | Med    | S    |
| 6  | Harvest this section into TODO_LIST.md / ROADMAP.md (docs-health HARVEST) and annotate this report as items complete                          | Med    | S    |
| 7  | Sweep `~/projects` for NEW direct consumers since 2026-10-11 (dependency-graph scan); update ADOPTION table                                  | Med    | M    |
| 8  | Re-run the go-finding version-distribution scan (60/31/4 snapshot is from 2026-10-11)                                                        | Med    | S    |
| 9  | Check whether branching-flow and erraudit (pilot-rule donors) now import the SDK; if so, add to ADOPTION                                      | Med    | S    |
| 10 | Re-run benchmarks vs the 2026-10-11 baseline after the go-finding v1.15.0 bump; refresh baseline file                                        | Med    | S    |
| 11 | BuildFlow: resolve the local `replace` on go-linter-sdk once its in-flight submodule wave lands (commit or discard, record in that repo)     | Med    | S    |
| 12 | Make ADOPTION verification-mode explicit per row (grep-only vs build+test) so partial verification is visible                                | Med    | S    |
| 13 | Record "re-run gate after formatter mutation" in the fleet lessons file (cross-project failure mode, 2 occurrences)                          | Med    | S    |
| 14 | cmdguard: decide whether `website/src/styles/global.out.css` (generated, now 1,537 lines in history) should be gitignored                    | Med    | S    |
| 15 | cmdguard: AGENTS.md is 359 lines (fleet doc guard wants ≤220) — split detailed content into docs/                                             | Med    | M    |
| 16 | cmdguard: evaluate toolsdk v1.15.0 adoption re: `lint/provider/provider.go:33` missing `SwitchCases` finding                                 | Med    | M    |
| 17 | cmdguard: fix go-humanize-linter warning `main.go:193` (manual pluralization)                                                                | Low    | S    |
| 18 | cmdguard: investigate duplicate go-humanize-linter findings ([lint] + [root]) — module fan-out dedup question for BuildFlow                  | Low    | M    |
| 19 | cmdguard: add dprint, tailwindcss, vulnix to devShells.default (kills the nix-fallback warnings every run)                                   | Low    | S    |
| 20 | cmdguard: rebuild stale `result` symlink (`nix build .# -o result`) to quiet vulnix advisory                                                 | Low    | S    |
| 21 | cmdguard: rebuild/reinstall the stale BuildFlow binary (built at `518739c`, HEAD `d114f28`)                                                  | Low    | S    |
| 22 | cmdguard: migrate or delete legacy `~/.cache/buildflow/buildflow.db`                                                                         | Low    | S    |
| 23 | cmdguard: fix or blocklist the tsc `type-check`/`tsconfig-check` environment failures (nix-copy noise, no real diagnostics)                  | Low    | S    |
| 24 | cmdguard: trim the pre-commit step set (budget 60s, actual ~2min) to shrink daemon race windows                                              | Low    | M    |
| 25 | project-meta: normalize `go 1.27.1` patch-floor to `go 1.27` in go.mod + go.work (fleet rule: major.minor only)                              | Low    | S    |
| 26 | project-meta: give bump `5e527d5` a narrative entry in its CHANGELOG if the repo keeps one                                                   | Low    | S    |
| 27 | cmdguard + project-meta: pin `-race` in their CI for the lint modules (verified locally this session, CI coverage unconfirmed)               | Low    | S    |
| 28 | Verify the go.work rename/restore left zero residue in cmdguard and project-meta (committed vs ignored state)                                | Low    | S    |
| 29 | Watch go-finding sub-module tags (toolsdk v1.15.0 leads root) for SDK-relevant surface; scan on next bump                                     | Low    | S    |
| 30 | go-linter-sdk: review `[Unreleased]` wording for the go-error-family indirect bump accuracy before any release                                | Low    | S    |
| 31 | go-linter-sdk: check this repo's AGENTS.md against the fleet agents-md-size guard (cmdguard's tripped at 359/220)                             | Low    | S    |
| 32 | go-linter-sdk: evaluate `DetectorV2`/coverage-channel adoption surface from go-finding v1.15.0 (ROADMAP candidate)                            | Med    | M    |
| 33 | Track GO-2026-6617 (stdlib HTTP/2 HPACK race, flagged by govulncheck in cmdguard); bump toolchain when the fix lands                          | Med    | S    |
| 34 | go-linter-sdk: confirm the stranger-test CI job exercised v0.4.0 (it resolves `@latest`) — evidence in last green run                         | Low    | S    |
| 35 | Annotate the 2026-10-11 report's remaining section-f items that later sessions complete (only #19 annotated so far)                           | Low    | S    |
| 36 | go-linter-sdk: consider folding "docs-only PATCH refreshes pkg.go.dev README" into the release checklist explicitly                          | Low    | S    |
| 37 | samber-linter: after fixing #2, re-run its full suite and re-certify the consumer in ADOPTION                                                | Med    | S    |
| 38 | go-humanize-linter, go-structure-linter: their bumps were done by their own sessions without cross-verification evidence — spot-check their gates | Low | S    |
| 39 | BuildFlow: wait for its CI on the in-flight submodule wave, then re-issue the wave verification for all three modules                         | Med    | S    |
| 40 | Fleet: standardize "record-and-skip leaves a dated note in the skipped repo" (BuildFlow case)                                                | Low    | S    |
| 41 | go-linter-sdk: add the samber bisect pattern (two-command version bisect) to the migration guide's pitfalls section                          | Low    | S    |
| 42 | Fleet: propose a BuildFlow step that re-runs the last gate after any auto-fix rewrote a file (upstream fix for d1)                            | Med    | M    |
| 43 | go-linter-sdk: ROADMAP entry for consumer-facing DetectorV2 passthrough (coverage evidence flows only if the SDK's detectors implement it)    | Low    | S    |
| 44 | cmdguard: post-fix, add a regression test that RuleTableMarkdown output is dprint-stable (round-trip property test)                           | Low    | S    |
| 45 | project-meta: confirm its daemon-committed go.sum state passes its own CI (unpushed)                                                          | Low    | S    |
| 46 | go-linter-sdk: consider noting the "Imported by: 3" refresh cadence in ADOPTION (manual check each wave)                                      | Low    | S    |
| 47 | Sweep stale sibling dirs (`BuildFlow-pr44`, `BuildFlow.vendor.bak`) — delete or document why they stay                                        | Low    | S    |
| 48 | go-linter-sdk: examples' README claims vs master rule count (frozen-tag drift recurs each release; automate a check?)                         | Low    | M    |
| 49 | Fleet: decide whether ecosystem waves should auto-push consumer commits (ties to g1)                                                          | Med    | S    |
| 50 | Debrief: was skipping BuildFlow's local gates the right call? (Ask g3; record the verdict as a precedent)                                     | Low    | S    |

## g) Questions I cannot figure out myself

1. **Should ecosystem-wave commits to consumer repos be pushed automatically?** Your standing rule (recorded 2026-09-09,
   for THIS repo) is that I push at session end; the global default is never-push-without-ask. cmdguard and
   project-meta bumps are sitting unpushed because I could not tell which rule you intend for *other* repos during a
   wave you approved. Push them now?
2. **Release vehicle for the pending `[Unreleased]` entry:** cut **v0.4.1** now (dependency bump + gets the improved
   README onto pkg.go.dev, which currently freezes the stale "8 constants" version) or let it accumulate into the next
   feature MINOR? My lean is v0.4.1 for the README refresh alone; you own the cadence.
3. **samber-linter ownership:** its master is red (`TestGoWorkMultiModule`) and bisect-proven unrelated to the SDK. Is
   that repo free for me to diagnose-and-fix, or is another session of yours actively working it (as in BuildFlow right
   now) and I should file instead of touch?

---

*Snapshot ends. Point-in-time report — annotate, don't rewrite (docs-health ANNOTATE). Section (f) items 1–12 are the
HARVEST candidates; the tail needs routing rigor before becoming commitments.*
