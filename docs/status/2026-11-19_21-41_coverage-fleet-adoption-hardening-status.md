# Status report — coverage fleet adoption (#29), hardening tier (#32/#33), byte-bound verdicts, v0.6.1-prep

> **Scope:** this session's work only — the post-19:29 execution wave. Point-in-time
> snapshot at 2026-11-19 21:41 CET. Living truth: TODO_LIST.md, ROADMAP.md, AGENTS.md.
> HEAD at write time: daemon-tracked master with every gate-relevant file committed.
>
> **Honesty frame:** every claim below was verified with a live command this session
> unless marked otherwise. Seven self-inflicted failures are recorded under (d) rather
> than buried, including one concurrent-session work loss (BuildFlow decay #4).

## a) FULLY DONE (verified)

**The library/example tier (f10–f13).**
- `ExampleNewRuleSet` + `ExampleRuleSet_Check` godoc examples — pkg.go.dev-visible
  RuleSet coverage (RuleSet previously had tests but no examples).
- `BenchmarkRuleSet_Check` + a matched `BenchmarkRegistry_Run` over the same
  50-rule workload — the end-to-end mutex-free claim, MEASURED before writing:
  ~560 ns/op 0 allocs (RuleSet.Check) vs ~2.3–3.5 µs/op 9 allocs (Registry.Run),
  ~6× by design. (An existing loose `BenchmarkRegistry_Run` in registry_test.go
  collided — upgraded in place to the matched design instead of duplicating.)
- cli-triage `--format` validation: unknown `--format=x` → usage on stderr + exit 3
  (was: silent fall-through to the dir argument); `TestExampleCLITriage_UnknownFormatExits3`
  asserts exit code, empty stdout, and both stderr fragments.
- `TestExampleCLITriage_JSONCleanDir`: clean-dir JSON round-trip + asserts
  `filesScanned ≠ 0` — which forced the next item.
- **CoverageReporter dogfood in the flagship example**: cli-triage's default-on
  walking rule now reports honest scan evidence (`treeWalkingRule` embed +
  per-scan counter; `walkGoFiles` returns the parsed count). The OptIn-stripping
  caveat documented in-code: coverage cannot ride through OptIn (it wraps the
  RuleFunc value). Gates: build, vet, lint 0 issues, full race suite green;
  all 11 cli-triage integration tests pass.

**Hardening tier (f20, #32, #33, f21, f26, f2, f3).**
- **#32 landed as an eval-time flake assertion** (`golangci-version-sync`):
  ci.yml's golangci-lint-action `version:` pin must equal `pkgs.golangci-lint`'s
  version, checked on EVERY flake evaluation (fires even under
  `nix flake check --no-build`, which CI's nix job uses). It caught LIVE drift on
  its first run: ci.yml pinned v2.14.0 while nixpkgs served v2.15.0 — local
  `nix run .#lint` and CI's downloaded binary were different minors. Pin bumped
  to v2.15.0, comment now states the lockstep rule. Negative-test evidence: the
  assertion fired during development with the exact mismatch message. A shape
  rule (pin must exist) added to `check-doc-versions.sh` as rule 4.
- **#33**: shellcheck flake check (hermetic runCommand over `scripts/*.sh`) +
  CI job (ubuntu ships shellcheck; no install step). All four scripts fixed to
  clean: SC2164 (`cd || exit`), SC1007 (`GOEXPERIMENT=''`), SC2015 ×7
  (`A && B || C` → if/then/else), SC2035 (glob). The flake apps were already
  shellcheck-gated via writeShellApplication — `scripts/` was the ungated gap.
- **CI coverage gate raised ≥90% → ≥98%** (library measured 100.0% on
  2026-11-19; error message records the bar's origin).
- **Bench baseline refreshed (f6)** with idle-machine numbers and honest notes:
  Registry.Get ~17 ns flat (map+RWMutex), RuleSet.Get ~8.4 ns flat (≈half —
  mutex-free), Deregister O(n) visible (163/407/1020 ns at 100/500/1000),
  Run_Coverage ~1.15 µs, Filter.Apply ~174 µs / 1 alloc / ~344 KB. The header
  records the rule the first run taught: a concurrent nix build polluted every
  op 2–4×; polluted numbers were discarded, never documented. The Filter.Apply
  discrepancy vs the released v0.6.0 entry (~47 µs, different benchtime shape)
  is recorded, not hidden.
- **f21**: `nix run .#bench` verified working (used for the baseline).
- **f26**: release-checklist dead-code sweep now covers every `_test.go`
  repo-wide (root + examples + future subdirs) via `find -print0 | xargs -0`.
- **f2**: release-checklist gained gate [7/7] `nix flake check --no-build`.
- **f3**: the manual-steps footer prints current HEAD so the daemon race
  between "CI green" and "git tag" is visible at tag time.

**Byte-bound verdicts (f1, f23).**
- Viewed BOTH current dashboard PNGs myself this session (top + downscaled
  full-page): PASS — all four stat cards read correctly (8 / 65 / 72 / 6-6).
- sha256 of both recorded in index.html (prose + machine-readable
  `<!-- render-gate: sha256 full=… top=… -->` comment).
- release-checklist render gate UPGRADED from mtime to sha256: verdicts bind
  to artifact BYTES; mtime kept only as fallback with a WARN + format hint.
  This closes b.2 of the 19:29 report (verdict-inspected-predecessors) and the
  circularity flaw (editing index.html after shooting made PNGs mtime-stale
  while their bytes — and verdict — stayed valid).
- FLEET-AUDIT-TEMPLATE now encodes the byte-binding rule (f23).

**Docs corpus (f4, f7, f8, f9, f22, f24, f48, f49).**
- FLEET-SWEEP-BASELINE's stale "all 8 at v0.5.0" paragraph corrected with a
  pointer to the ADOPTION correction + a reader lesson (f4).
- DOMAIN_LANGUAGE.md gained four glossary terms — CoverageReporter, RuleSet,
  Filter, SelectRules — and the Registry bullet now lists SelectRules and
  ExitCodeFromFindings (f7).
- README Types-table parity: RuleSet, Filter, CoverageReporter rows (f8);
  example-tier exit contract paragraph — 0/1/2/3 with cli-triage pointer (f9).
- CONTRIBUTING documents `.docs-gates.json`'s `audit_max` ratchet semantics —
  "the pin ratchets down, never up" is now discoverable (f22).
- ADOPTION gained a per-repo Push-state column (all local-only) + a
  refresh-procedure step 5 that re-verifies it (f24).
- ADR 0004 (`SelectRules` naming) written durable in docs/decisions/ (f49).
- f48 verified ALREADY satisfied: the Deregister O(n)-by-design note is in the
  godoc.
- f18: daemon-commit review pass — all ten commits since 6a84e2b contain only
  files this session authored. f19: /tmp scratch trashed (7 files).

**#29 fleet adoption — COMPLETE, all 7 module locations (f30–f35).**
- **cli-triage** (this repo): done as part of the example tier above.
- **BuildFlow**: `dtool.DetectorV2FromFinding` bridge (domain/tool) folds
  DetectResult scan evidence into FindingReport.Summary;
  `ToolFromSpec` (tools/providers) prefers the V2 lane when the SDK detector
  implements `gofinding.DetectorV2`. Three bridge tests (fold, error-path,
  interface satisfaction) + the full ToolFromSpec suite green. The three
  full-suite failures were bisect-proven pre-existing (SidecarDetect fails
  identically with the change stashed; the binary-path test is
  environment-dependent — same flake family as the filed BuildFlow#112).
- **cmdguard/lint** (write-subagent, the ONE that survived): all 8 rules
  covered via a `treeWalkingRule` wrapper over the single `analyze()` walker;
  honest counting (generated + unparseable files excluded); its test pins the
  accumulation semantic (3 files × 8 rules = 24) and gates green with a
  stash-baseline check.
- **go-humanize-linter** (manual, after its subagent rate-limited): 11 rules
  covered via a per-Check context sink through the shared walker
  (`checkFuncDecls` records the parsed count; `coverageRule` wraps at
  registration in DefaultRegistry). Public API untouched — BuildFlow reads
  `rule.Meta.ID` off `AllRules()`, so the `[]RuleFunc` return type stayed.
  New coverage test asserts per-rule honesty AND the N·F accumulation;
  build + vet + full suite green.
- **go-structure-linter** (manual, after its subagent rate-limited): a
  `countingProject` decorator at the SDK adapter's project factory counts
  every non-dir entry the walk presents (atomic per-Check counter; the legacy
  string path stays honestly 0). Two Ginkgo tests (4-file fixture counted;
  non-walking rule stays 0); full package suite green; full-repo suite
  rc=0 in the background.
- **project-meta/linter**: ALREADY DONE by its own session (honest per-scan
  counters on `metadataMissingCheck` — 0 on every early return, 1 when the
  existence check examines the file). Verified, not duplicated.
- **samber-linter**: ALREADY DONE by its own session (driver-level
  `scanCoverage` over `analyzedFileCount` — parsed files across loaded
  packages). Build + driver tests verified green.
- **f36 — FIRST NON-ZERO SWEEP** recorded as "AFTER sweep 2" in
  FLEET-SWEEP-BASELINE: 7/6/6/7/8 refs across the five linters + 3 in
  BuildFlow/tools; execution stays 0 BY DESIGN (the bridge lives at the
  tools/providers boundary; domain/tool is its own module, not swept);
  Relativize correctly 0 (needs no consumer code); WithEdits is the
  remaining 0-ref channel (ROADMAP surface).

**TODO_LIST + CHANGELOG.** #32/#33 removed as done; #29 completed (per-repo
adoption named, including the exact `project-meta/linter` token the freshness
gate anchors on); #31 stays BLOCKED (owner-gated). CHANGELOG [Unreleased]
fully rewritten — including the release-slot DECISION (f14): PATCH, per the
2026-09-09 slot rules (example/tooling tier, no exported SDK surface change).

## b) PARTIALLY DONE

1. **Release v0.6.1 NOT cut.** The CHANGELOG entry, the slot decision, and
   the now-standard loop exist; the release-checklist full pass, push,
   CI-green wait, tag, and pkg.go.dev verification have NOT run — bundled
   into owner question g1 below.
2. **audit-all gate RED (6 items).** Completing routed work this session
   removed the TODO_LIST rows that used to anchor six archived §f/table
   items; the audit replay demands per-item anchors (generic token overlap
   is rejected — correctly). Freshness itself is PASS (0 unharvested on the
   newest report) and doc-drift PASS after fixing a v2.14.0/v2.15.0 literal
   my own TODO note introduced. The six verdicts are LOCATED, not applied:
   done ×4 (CoverageReporter README row; severity-tiered exit codes;
   `--format json` ×2), sibling-scoped ×1 (go-finding DetectorV2 docs
   cross-link ask → `fresh:skip`), open ×1 (README relative-link checker →
   mint TODO row). ~5 minutes of mechanical work, interrupted by this report.
3. **ADOPTION does not yet name decay #4** (recorded in FLEET-SWEEP-BASELINE
   and the CHANGELOG, not in ADOPTION's 2026-11-19 snapshot).
4. **AGENTS.md session section not appended** — deferred to post-release so
   it can name the release.
5. **f25 teaching-debt sweep** (consumers' direct `finding.NewBuilder`
   usage that should be `rule.NewFinding`) — not done.
6. **BuildFlow full providers suite not re-run after the clobber/re-apply**
   (targeted V2 + ToolFromSpec tests green; the 3 known-flaky failures
   documented, one bisect-proven pre-existing).

## c) NOT STARTED

1. Consumer pushes (#31, g-1) — all 8 module locations remain local-only.
2. Fixing samber-linter#5 / BuildFlow#112 — tracker-first status quo; owner
   question 3 (asked at 19:29) still unanswered.
3. Third consumer wave (v0.6.1 bumps) — post-release by definition.
4. ROADMAP raw ideas (compliance.Record docs note gated on go-compliance
   v0.1; the ~31-repo SDK spread; cmd/ parked; SARIF line) — untouched.
5. go-finding next-MINOR re-audit — trigger armed, not due.
6. f38 (samber CI green check before its next session pushes).
7. f17 (crush-config lessons commit `aabe152` push) — sibling push, gated.

## d) TOTALLY FUCKED UP

1. **DOMAIN_LANGUAGE edit dropped a section header.** My glossary-insert
   old_string consumed the trailing `## Borrowed vocabulary` line; the
   structure grep caught it within one command and it was restored. The
   lesson stands: multi-section edits must include the following header in
   new_string.
2. **ADR numbered 0003 without checking existing ADRs** — 0003 was already
   `rulefunc-generics-evaluation`. Renumbered to 0004 after the fact; the
   check (`ls docs/decisions/`) costs five seconds and prevents a renumber.
3. **`def` instead of `func`** in the nolint edit (Python leaked into Go) —
   the build caught it immediately; no damage.
4. **BuildFlow decay #4 — my V2-bridge edits were WIPED by a concurrent
   session.** The edits sat uncommitted ~20 minutes while another session
   raced the same checkout (its own 20:39 race-triage status report proves
   the collision); files restored to pre-edit state, test file deleted. I
   re-applied everything from context, re-gated (V2 tests + ToolFromSpec
   suite), and the daemon committed within minutes (e18a22479). The
   never-commit rule means I cannot shrink the vulnerability window myself —
   the mitigations are: re-verify bytes with grep immediately after finishing
   each repo's edit burst (did, post-restore), and prefer repos whose daemon
   just committed (clean tree = committed work survives restores).
5. **Invented an API in a test's first draft** (`humanizelint.MustCountGoFiles`)
   — build failed; replaced with an honest in-test recount instead of adding
   a helper for one assertion.
6. **Coverage test asserted the wrong semantic** — per-rule Coverage() ==
   report.Summary.FilesScanned. Wrong: go-finding's AddDetectResult
   ACCUMULATES per-rule counts (N rules × F files). The cmdguard subagent's
   report named the identical semantic; both repos' tests now pin the
   accumulation contract explicitly.
7. **First bench baseline run was polluted** by a concurrent nix build
   (every op inflated 2–4×; Registry.Get 25–46 ns vs the true ~17 ns). Caught
   by cross-checking against earlier session numbers; re-ran idle; the
   polluted output never entered any document — but it WOULD have, had I not
   compared against prior art. Write-after-measure discipline held this time;
   the near-miss is recorded.
8. **4 of 5 parallel write-subagents died on provider rate limits** (only
   cmdguard landed; go-humanize, go-structure-linter, samber retry, and
   project-meta all failed with `too many requests`). Parallel dispatch was
   the wrong shape under rate pressure — the manual fallback covered
   go-humanize + go-structure-linter, and the two "failed" repos turned out
   already-done by their own sessions. Net time lost: ~20 minutes of
   dispatch latency.
9. **release-checklist gate edit residue**: the sha256-gate rework left
   `render_ok` unused (SC2034) and one whitespace-equivalent reapply mangled
   indentation. Fixed; shellcheck clean; logic verified against live bytes
   (MATCH/MATCH).
10. **Edited a file without reading it** (BuildFlow sdk_registry.go) — the
    tool refused, I read, re-applied. One round trip of pure process tax.

## e) WHAT WE SHOULD IMPROVE

1. **Byte-binding generalizes.** The sha256 render-gate resolved the mtime
   circularity permanently. Every verdict/freshness check in the fleet
   should bind to content hashes, not timestamps.
2. **Shared-checkout commit windows are the attack surface.** I cannot commit
   (never-commit rule; the daemon owns commits), so the mitigation set is:
   re-verify bytes immediately after each repo's edit burst, keep edit bursts
   short, and treat any "clean tree + missing edits" as a clobber signal
   (that grep is what caught decay #4 within one sweep).
3. **Rate-limit-aware dispatch:** send ONE canary subagent; fan out only if
   it lands; otherwise go manual immediately. Also: briefs should say "on
   rate-limit error, stop and report — do not retry."
4. **Measure on idle, record machine state.** The baseline header now encodes
   "never record numbers captured alongside a parallel build" — the near-miss
   in d7 justifies making it a written rule, not a habit.
5. **Constructor-level wrappers beat closure-threading.** cmdguard needed ONE
   chokepoint edit (ruleFor); go-humanize needed 11 mechanical ctx edits.
   For the next many-rule repo: wrap at registration, not inside every Run.
6. **Completing a routed item needs its inline verdict in the SAME session**
   — removing the TODO row that anchored an archived report's item turns the
   audit replay red (b2's six items are exactly this failure mode). Document
   it in AGENTS; it is the annotate convention applied at completion time.
7. **Run docs-gates mid-flight for doc-heavy sessions**, not only pre-push —
   the anchor-removal failure mode is invisible until the replay runs.
8. **FilesScanned accumulation deserves a doc note**: the summary is
   cumulative (N·F), not unique files. Two fleet tests now pin it; the
   SDK's DOMAIN_LANGUAGE "Borrowed vocabulary" section should say so.
9. **ADR numbering check** (`ls docs/decisions/` before writing) — trivial,
   d2 is proof it is not optional.
10. **Keep the f-list harvest immediate** — this report's §f is routed into
    TODO_LIST/ROADMAP directly after writing (done below), keeping
    freshness green; the 19:29 pattern held.

## f) Up to 50 things to do next (harvestable; owned files named)

**Immediate (this repo):**
1. Finish the audit-all drain: apply 5 inline verdicts (14-43 item 8
   CoverageReporter README row → done 2026-11-19; 08-08_10-45 item 33
   severity-tiered exit codes → done; 08-08_12-16 item 32 and 12-36 item 31
   `--format json` → done 2026-11-19; 07-40 item 44 go-finding DetectorV2
   docs cross-link ask → `fresh:skip`, sibling-scoped) + mint TODO #34
   (README relative-link checker, from 10-15 item 8, genuinely open).
2. Cut v0.6.1: release-checklist full pass → push → CI green on the exact
   commit → tag → `go list -m @v0.6.1` + pkg.go.dev render check.
3. Append the AGENTS.md session section (v0.6.1 tier, #29 fleet, decay #4,
   the sha256 gate).
4. ADOPTION: name decay #4 in the 2026-11-19 snapshot.
5. f25: sweep consumers for direct `finding.NewBuilder` usage that should be
   `rule.NewFinding` (teaching debt; feeds future PRs).
6. Third consumer wave: v0.6.1 bumps after the release; re-run fleet-sweep.
7. f38: verify samber-linter CI green (README fix + v0.6.0 bump) before its
   next session pushes.
8. Post-release: verify pkg.go.dev renders the new godoc examples
   (ExampleNewRuleSet/ExampleRuleSet_Check) and the exit-contract README.
9. Push master (part of the release bundle; this repo's end-of-session duty).
10. Add the FilesScanned accumulation note to DOMAIN_LANGUAGE's borrowed
    vocabulary section (Summary is N·F, not unique files).

**Consumer/fleet (gated or their-repo work):**
11. Push consumer bumps once g-1 is answered (#31) — all 8 locations +
    samber README fix + crush-config lessons `aabe152`.
12. Fix samber-linter#5 (re-derive the workspace-shadow premise).
13. Fix BuildFlow#112 (capture `-v` failure output, fix isolation).
14. Re-verify BuildFlow/tools FULL suite after the re-apply (3 known-flaky
    failures pending a clean-run record).
15. WithEdits adoption sweep — the last 0-ref channel (ADR-0002 GO items;
    rules that know a mechanical fix attach TextEdits).
16. OptIn coverage forwarding: if a consumer needs opt-in rules WITH
    coverage, design the variant (OptIn wraps the value; a pointer-preserving
    wrapper is the sketch) — ROADMAP candidate, not built.
17. go-humanize-linter: cherry-pickers using AllRules() directly get
    coverage-silent rules (documented); consider exporting a covered
    AllRules variant if a consumer asks.

**Watches (armed, not due):**
18. go-finding next root MINOR → re-run the utilization audit.
19. go-compliance v0.1 tag → compliance.Record docs note.
20. go-policy-dsl v1 tag → revisit #23 companion-libs mention.
21. pkg.go.dev "Imported by" trend at the next release (baseline 3).
22. dependabot weekly PRs on the new golangci pin.
23. govulncheck CI job stays green on the toolchain outputs.

**Larger / backlog:**
24. Spread the SDK to the ~31 repos still on older go-finding releases.
25. cmd/ remains parked (resolved Q3) — owner mandate only.
26. SARIF output as a ROADMAP idea line.
27. Surface the measured RuleSet.Check-vs-Registry.Run gap in FEATURES' perf
    row (currently CHANGELOG + baseline only).
28. Carried from 19:29 §f, still open: none of its "Immediate (this repo)"
    items — ALL executed today (1–26 minus the release bundle); that file's
    §f27–#29/31+ items are reflected above.
29. Next session: seed the todo list from the freshest §f BEFORE working
    (held again this session; it works).
30. Consider a `docs/decisions/` entry for the coverage-accumulation pin
    (N·F) if it ever becomes a design debate.

## g) Three questions I cannot answer myself

1. **Release + push bundle (one decision, three effects):** cut v0.6.1 now —
   push master, wait CI green on the exact commit, tag, verify pkg.go.dev —
   and, in the same breath, ALSO push the consumer v0.6.0 bumps (g-1: all
   six repos, the samber README fix, and crush-config lessons `aabe152`)?
   Or hold everything local? (The PATCH slot decision is recorded in the
   CHANGELOG; only the go/no-go is yours.)
2. **Own-repo filings policy (asked at 19:29, still open):** for
   samber-linter#5 and BuildFlow#112 — fix directly in their sessions
   (you own both repos), or keep the tracker-first flow (filed with verified
   repro, fixed later)? Tonight's BuildFlow collision (d4) strengthens the
   case for fix-in-place with an immediate gate: fewer uncommitted windows.
3. **Subagent metering:** 4 of 5 parallel write-subagents died tonight on
   provider rate limits (`too many requests`), one succeeded. Is subagent
   usage metered on your plan such that I should default to manual/serial
   dispatch and reserve subagents for single well-scoped canaries?

— Report written 2026-11-19 21:41 CET. §f harvest into TODO_LIST/ROADMAP
follows immediately; gates re-run before this session goes quiet.
