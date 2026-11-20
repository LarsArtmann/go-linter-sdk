# Pareto Plan — Fleet-Review Follow-ups & Roadmap (2026-11-19 07:54)

> **Scope:** all open work for `go-linter-sdk` as of 2026-11-19 07:54:
> TODO_LIST #20–#27 (minted by the 2026-11-19 fleet libs review), ROADMAP
> themes 1–5, and the 40 unrouted [NEW]/[PROC] items from
> `docs/status/2026-11-19_07-40_fleet-libs-review-self-review-and-status.md`.
> **Customer** = the 6 direct consumers + ~60 indirect importers + every
> future adopter (the examples are the adoption funnel).
>
> **Errata / verification (skill step 0):** every ranked task was checked
> against source read this session — `Filter`, `NewRegistryFromRules`,
> `Category.All`, `RuleSet`, `ExitCodeFromFindings`, and the O(1) index do
> NOT exist in `registry.go`/`rule.go` (confirmed by full read); #20–#27 were
> created from the audit 40 minutes before this plan and are open by
> construction; the daemon has committed everything since (tree clean at
> `d36e335`). No stale backlog items found. The three open user questions
> (fleet rollout, #24/#25-vs-#20/#21 priority, release shape) are NOT blocking:
> this plan makes a recommendation for each and flags where a different user
> answer would reorder tiers.

## The Pareto thesis

The fleet review found the SDK superbly built and verified, with exactly two
classes of debt: **report truthfulness** (every consumer's report says
`filesScanned: 0` and embeds host paths in finding IDs — the library ships
the fix, the SDK never wired it) and **teaching surface** (the four examples
teach the pre-helper pattern). Both are cheap. Everything else is
nice-to-have ergonomics or parked strategy.

## 1) The 1% that delivers 51% — "make every consumer's reports TRUE"

| #  | Task                                                                      | Why it is the 1%                                                      |
| -- | ------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| M1 | DetectorV2 coverage evidence (`AddDetectResult`) in `Run` + both adapters | `filesScanned ≠ 0` for the whole fleet; unblocks dashboards/CI checks |
| M2 | `finding.Relativize` in both Detector adapters                            | Stable finding IDs across CLI/BuildFlow lanes; one call each          |
| M3 | `RuleMeta.Cat` shape-validation via `finding.Category.IsValid()`          | Closes the taxonomy near-copy gap; trivial                            |

~2.5 h of work, inherited by 6 direct consumers on the next bump. This is the
smallest set with the largest single observable payoff.

## 2) The 4% that delivers 64% — add: "teach the right pattern"

| #  | Task                                                          | Why                                                                     |
| -- | ------------------------------------------------------------- | ----------------------------------------------------------------------- |
| M4 | Modernize all 4 examples (`WithToolName` + `NewFinding`)      | The adoption funnel stops teaching boilerplate the SDK exists to delete |
| M5 | Fifth example: OptIn + FilterRules + ExitCodeByConfidence CLI | The only capabilities with no runnable demo                             |
| M6 | Sync README quickstart + godoc examples to the modern pattern | pkg.go.dev is the shop window                                           |

## 3) The 20% that delivers 80% — add: "discoverability + trust + the next release"

| #   | Task                                                                                                              | Why                                                                          |
| --- | ----------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| M7  | toolsdk registration recipe in README (#27)                                                                       | 14/23 fleet repos re-derived it by hand                                      |
| M8  | README/CONTRIBUTING/ADOPTION pointers to `docs/review/libs/`                                                      | Strangers cannot find the review corpus today                                |
| M9  | DOMAIN_LANGUAGE relationships diagram update                                                                      | Finish the half-done fix-on-sight (diagram predates NewFinding/WithToolName) |
| M10 | Vision-model verification of the dashboard + re-shoot PNGs                                                        | Close the one PARTIAL gate from the audit                                    |
| M11 | go-error-family mini-audit (the forgotten 8th library)                                                            | go.mod names it; the fleet list didn't                                       |
| M12 | Document the golangci-lint-auto-configure converge workflow (#22)                                                 | Tribal knowledge → written practice                                          |
| M13 | **Release v0.5.0 (MINOR)** after M1–M3: CHANGELOG, full gate, tag-on-green, pkg.go.dev verify, consumer bump wave | Ships the 51% to the fleet                                                   |
| M14 | Consumer fleet sweep: 3 missed channels × 6 consumers + before/after baseline                                     | Converts findings into fleet-wide fixes                                      |

## 4) The other 20% (to reach 100%)

Ergonomics (ROADMAP theme 1), the fix-representation design, tag-currency
hygiene, report hygiene, process extraction, parked strategy reviews.

---

## Medium plan — 27 tasks, 30–100 min, sorted by importance/impact/effort/customer-value

Impact: H/M/L on the fleet. Effort: minutes. Value: customer-visible outcome.

| #   | Tier | Task                                                                                                    | Impact | Effort | Customer value                                        | Depends      | Source           |
| --- | ---- | ------------------------------------------------------------------------------------------------------- | ------ | ------ | ----------------------------------------------------- | ------------ | ---------------- |
| M1  | 1%   | Wire `DetectorV2`/`AddDetectResult` into `Run`, `DetectorFromRegistry`, `DetectorsFromRegistry` + tests | H      | 90     | Reports become checkable (`filesScanned ≠ 0`)         | —            | TODO #24         |
| M2  | 1%   | `Relativize` in both adapters + semantics check + tests                                                 | H      | 45     | Stable finding IDs, no host paths                     | —            | TODO #25         |
| M3  | 1%   | `Category.IsValid()` shape-validation in `validateIdentityFields` + tests                               | M      | 30     | Bad categories surface at registration                | —            | TODO #26         |
| M4  | 4%   | Modernize 4 examples to current API; fix tool-name identity mismatch                                    | H      | 90     | Adoption funnel teaches helpers                       | —            | TODO #20         |
| M5  | 4%   | Fifth example `cli-triage` (OptIn + FilterRules + ExitCodeByConfidence)                                 | M      | 90     | Only undemoed capabilities get a demo                 | M4 (pattern) | TODO #21         |
| M6  | 4%   | README quickstart + `example_test.go` sync to modern pattern                                            | M      | 30     | pkg.go.dev shows the right shape                      | M4, M5       | NEW (status f46) |
| M7  | 20%  | toolsdk registration recipe (copy-pasteable) in README                                                  | M      | 45     | Fleet stops re-deriving BuildFlow wiring              | —            | TODO #27         |
| M8  | 20%  | Discoverability links: README docs-section, CONTRIBUTING, ADOPTION crosslink                            | M      | 30     | Review corpus findable by strangers                   | —            | NEW (f13/f43)    |
| M9  | 20%  | DOMAIN_LANGUAGE relationships diagram update + guard run                                                | L      | 30     | Domain doc stops lying by omission                    | —            | NEW (f14)        |
| M10 | 20%  | Vision-verify dashboard PNGs, re-shoot post-edit renders                                                | M      | 30     | Render gate actually closed                           | —            | NEW (f11/f19)    |
| M11 | 20%  | go-error-family mini-audit report + index row                                                           | M      | 60     | Fleet list complete (8/8 deps audited)                | —            | NEW (f12)        |
| M12 | 20%  | AGENTS tooling section: converge workflow note                                                          | L      | 30     | Practice survives the session                         | —            | TODO #22         |
| M13 | 20%  | Release v0.5.0: CHANGELOG, full gate, tag-on-green, push, pkg.go.dev verify, bump-wave kickoff          | H      | 100    | The 51% ships to 6 consumers                          | M1–M3        | status g3 (rec)  |
| M14 | 20%  | Fleet sweep: 3 missed channels × 6 consumers + baseline evidence file                                   | H      | 100    | Findings become fleet fixes; release evidence         | M13          | NEW (f23/f24)    |
| M15 | Rest | Fix-representation DESIGN (Fixer/TextEdit → toolsdk Repairer + dry-run), ADR draft, go/no-go            | M      | 100    | Unblocks direct-fix pipeline (deleted erraudit fixer) | —            | RMAP + f50       |
| M16 | Rest | `Filter` type (severity/category) — tests first                                                         | M      | 90     | Consumer filtering without hand-rolled slices         | —            | RMAP             |
| M17 | Rest | Convenience trio: `NewRegistryFromRules`, `Category.All`, `ExitCodeFromFindings`                        | M      | 90     | Less boilerplate at 6 call sites                      | —            | RMAP             |
| M18 | Rest | O(1) map index for Get/Has/Deregister + bench re-run + race test                                        | M      | 90     | Removes O(n) scans (real at scale)                    | —            | RMAP             |
| M19 | Rest | Typed `RuleSet` wrapper + tests + docs                                                                  | L      | 60     | Mutex-free consumers                                  | —            | RMAP             |
| M20 | Rest | RuleFunc generics evaluation spike + decision note                                                      | L      | 60     | Type-safe rules or a documented no                    | —            | RMAP             |
| M21 | Rest | Tag-currency check: fetch 6 repos, compare local vs published, annotate reports                         | M      | 45     | "Latest local tag" claims become absolute             | —            | NEW (f18)        |
| M22 | Rest | Report hygiene: re-review-trigger footers, markdown consistency pass                                    | L      | 60     | Corpus stays re-reviewable                            | M21          | NEW (f36/f37)    |
| M23 | Rest | CONTRIBUTING "version reviewed" convention + registry.go toolsdk naming fix                             | L      | 30     | Future writers don't trip the guard                   | —            | NEW (f38/f39)    |
| M24 | Rest | Extract fleet-audit prompt template → `docs/review/libs/FLEET-AUDIT-TEMPLATE.md`                        | M      | 60     | Next repo's audit starts compliant                    | —            | NEW (f22/f34)    |
| M25 | Rest | go-finding HEAD watch note + pkg.go.dev v0.4.1 frozen-README check                                      | L      | 30     | Re-audit trigger armed; docs-only PATCH decision      | —            | NEW (f41/f45)    |
| M26 | Rest | `cmd/` CLI decision review (confirm parked, or reopen with Q3 criteria)                                 | L      | 30     | Strategy honesty                                      | M13          | RMAP (parked)    |
| M27 | Rest | docs-health HARVEST this plan + status report into TODO/ROADMAP; docs-gates verify                      | H      | 45     | Nothing entombed in snapshots                         | all          | skill mandate    |

**Total:** ~22 h medium-grain. Tier 1+2 (M1–M6) ≈ 6.5 h delivers the 64%.

---

## Fine plan — 97 tasks, ≤12 min each, sorted by tier then execution order

Grouped under their medium task; within a group, order IS the execution order.

### Tier 1% (M1–M3) — 14 tasks

| #    | Task                                                            | Min | Done-when                      |
| ---- | --------------------------------------------------------------- | --- | ------------------------------ |
| F1.1 | Read `detect_result.go` + `examples/detectorv2` in module cache | 12  | Wiring pattern understood      |
| F1.2 | Design DetectResult flow in `Run` (per-rule vs aggregate)       | 12  | Decision noted in code comment |
| F1.3 | Implement `AddDetectResult` in `Registry.Run`                   | 12  | Run populates coverage         |
| F1.4 | Implement in `DetectorFromRegistry`                             | 12  | Adapter returns DetectResult   |
| F1.5 | Implement in `DetectorsFromRegistry`                            | 12  | Per-rule detector named        |
| F1.6 | Test: `Run` coverage fields non-zero                            | 12  | Green                          |
| F1.7 | Test: adapters coverage fields                                  | 12  | Green                          |
| F1.8 | Update godoc example + bench                                    | 12  | `go test` green                |
| F2.1 | Verify `Relativize` semantics (copy vs mutate) from source      | 12  | Semantics noted                |
| F2.2 | Wire into `DetectorFromRegistry` + test                         | 12  | IDs relative                   |
| F2.3 | Wire into `DetectorsFromRegistry` + test                        | 12  | IDs relative                   |
| F2.4 | Note behavior change in report addendum                         | 12  | go-finding.md annotated        |
| F3.1 | `finding.Category(cat).IsValid()` in `validateIdentityFields`   | 12  | Warns on unknown shape         |
| F3.2 | Tests: valid custom vs invalid shape                            | 12  | Green                          |

### Tier 4% (M4–M6) — 14 tasks

| #    | Task                                                        | Min | Done-when                            |
| ---- | ----------------------------------------------------------- | --- | ------------------------------------ |
| F4.1 | minimal-linter → `NewRegistry(WithToolName)` + `NewFinding` | 12  | Compiles                             |
| F4.2 | minimal-linter integration verify                           | 12  | `go test` green, identity consistent |
| F4.3 | no-go-mod modernize                                         | 12  | Compiles                             |
| F4.4 | no-go-mod verify                                            | 12  | Green                                |
| F4.5 | naked-return-guard modernize                                | 12  | Compiles                             |
| F4.6 | naked-return-guard verify                                   | 12  | Green                                |
| F4.7 | oops-redundant-guard modernize                              | 12  | Compiles                             |
| F4.8 | oops verify + full suite + race                             | 12  | All green                            |
| F5.1 | Scaffold `examples/cli-triage` (arg parsing)                | 12  | Runs                                 |
| F5.2 | OptIn rule + `FilterRules` wiring                           | 12  | Flags filter rules                   |
| F5.3 | `ExitCodeByConfidence` exit path                            | 12  | 0/1/2 observed                       |
| F5.4 | Integration test + README mention                           | 12  | Green                                |
| F6.1 | README quickstart → modern pattern                          | 12  | Snippets match examples              |
| F6.2 | `example_test.go` freshness check                           | 12  | pkg.go.dev examples current          |

### Tier 20% (M7–M14) — 36 tasks

| #     | Task                                                               | Min  | Done-when                           |
| ----- | ------------------------------------------------------------------ | ---- | ----------------------------------- |
| F7.1  | Draft `toolsdk.Register(Spec{...})` snippet against toolsdk source | 12   | Snippet written                     |
| F7.2  | README "Plug into BuildFlow" section + blank-import note           | 12   | Section reads complete              |
| F7.3  | Standalone-compile the snippet in a temp module                    | 12   | Compiles                            |
| F8.1  | README docs-links section                                          | 12   | Links live                          |
| F8.2  | ADOPTION crosslink to review corpus                                | 12   | Linked                              |
| F8.3  | CONTRIBUTING pointer                                               | 12   | Linked                              |
| F9.1  | Redraw DOMAIN_LANGUAGE relationships diagram                       | 12   | NewFinding/WithToolName/OptIn edges |
| F9.2  | Doc-drift guard + docs-gates run                                   | 12   | Green                               |
| F10.1 | Dispatch vision sub-agent on both PNGs                             | 12   | Layout verdict received             |
| F10.2 | Re-shoot renders post-title/footer edits                           | 12   | PNGs current                        |
| F10.3 | Fix any layout findings                                            | 12   | Verdict clean                       |
| F11.1 | go-error-family capability inventory (module cache)                | 12   | Inventory listed                    |
| F11.2 | Usage map + gap grading                                            | 12   | Rows graded                         |
| F11.3 | Write `go-error-family.md` + index row                             | 12   | Corpus 8/8                          |
| F12.1 | AGENTS tooling section: converge workflow                          | 12   | Written                             |
| F12.2 | Verify the documented commands run                                 | 12   | Verified                            |
| F13.1 | Draft v0.5.0 CHANGELOG entry                                       | 12   | Entry complete                      |
| F13.2 | Full gate: lint, vet, test-race, flake-check                       | 12   | All green                           |
| F13.3 | Commit, wait CI green, tag v0.5.0, push tag                        | 12   | Tag on green commit                 |
| F13.4 | pkg.go.dev verify + `go get @latest` check                         | 12   | Site serves v0.5.0                  |
| F13.5 | Consumer bump wave (6 × bump+build+test)                           | 12×6 | All six on v0.5.0                   |
| F13.6 | ADOPTION snapshot update                                           | 12   | Table current                       |
| F14.1 | Sweep go-structure-linter for 3 channels                           | 12   | Findings noted                      |
| F14.2 | Sweep samber-linter                                                | 12   | Noted                               |
| F14.3 | Sweep go-humanize-linter                                           | 12   | Noted                               |
| F14.4 | Sweep cmdguard/lint                                                | 12   | Noted                               |
| F14.5 | Sweep project-meta/linter                                          | 12   | Noted                               |
| F14.6 | Sweep BuildFlow/execution                                          | 12   | Noted                               |
| F14.7 | Baseline evidence file (before/after #24/#25)                      | 12   | Evidence saved                      |

### Rest — 33 tasks

| #     | Task                                                                | Min | Done-when            |
| ----- | ------------------------------------------------------------------- | --- | -------------------- |
| F15.1 | Read toolsdk `Repairer` + `WithDryRun` contract                     | 12  | Understood           |
| F15.2 | Sketch `Fixer` interface on Rule                                    | 12  | Sketch written       |
| F15.3 | Map erraudit autofixer outputs → `TextEdit`                         | 12  | Mapping table        |
| F15.4 | Design doc + ADR draft                                              | 12  | In docs/decisions/   |
| F15.5 | Go/no-go recommendation recorded                                    | 12  | Owner can decide     |
| F16.1 | `Filter` API + failing tests first                                  | 12  | Red                  |
| F16.2 | Implement + make green                                              | 12  | Green                |
| F16.3 | Godoc + README API table row                                        | 12  | Documented           |
| F17.1 | `NewRegistryFromRules` impl + test                                  | 12  | Green                |
| F17.2 | `Category.All()` impl + test                                        | 12  | Green                |
| F17.3 | `ExitCodeFromFindings` impl + test                                  | 12  | Green                |
| F17.4 | Godoc examples for the trio                                         | 12  | On pkg.go.dev        |
| F17.5 | Bench note (no regression)                                          | 12  | Bench compared       |
| F17.6 | README API table rows                                               | 12  | Documented           |
| F18.1 | Add map index alongside rules slice                                 | 12  | Struct changed       |
| F18.2 | Rewrite Get/Has/Deregister on index                                 | 12  | O(1)                 |
| F18.3 | Bench re-run vs baseline                                            | 12  | Improvement shown    |
| F18.4 | Race test (concurrent register/deregister/run)                      | 12  | Race-clean           |
| F19.1 | `RuleSet` type + methods                                            | 12  | Compiles             |
| F19.2 | Tests                                                               | 12  | Green                |
| F19.3 | Docs                                                                | 12  | Documented           |
| F20.1 | Generics spike sketch                                               | 12  | Sketch compiles      |
| F20.2 | Decision note (adopt/decline)                                       | 12  | Recorded             |
| F21.1 | `git fetch` all 6 library repos                                     | 12  | Current              |
| F21.2 | Compare tags; annotate reports if stale                             | 12  | Claims absolute      |
| F22.1 | Standardize re-review-trigger footers (7 reports)                   | 12  | All carry one        |
| F22.2 | Markdown consistency pass                                           | 12  | Consistent           |
| F22.3 | Fix findings from pass                                              | 12  | Clean                |
| F23.1 | CONTRIBUTING "version reviewed" convention                          | 12  | Written              |
| F23.2 | registry.go "tool-sdk" → name toolsdk module                        | 12  | Comment precise      |
| F24.1 | Extract 3 group prompts into template doc                           | 12  | Template exists      |
| F24.2 | Dry-run template against one library                                | 12  | Template works       |
| F25.1 | go-finding HEAD watch note in AGENTS                                | 12  | Trigger armed        |
| F25.2 | pkg.go.dev v0.4.1 frozen-README staleness check                     | 12  | PATCH decision noted |
| F26.1 | Re-read resolved Q3; confirm `cmd/` parked or draft reopen criteria | 12  | Decision current     |
| F27.1 | HARVEST plan+status NEW items → TODO/ROADMAP                        | 12  | Nothing entombed     |
| F27.2 | docs-gates verify                                                   | 12  | Green                |

**Fine total:** 97 tasks ≈ 19.5 h of clock time.

---

## Execution graph (mermaid)

```mermaid
flowchart LR
    subgraph T1["1% → 51%: report truthfulness"]
        M1[M1 DetectorV2 coverage]
        M2[M2 Relativize IDs]
        M3[M3 Category IsValid]
    end
    subgraph T2["4% → 64%: teaching surface"]
        M4[M4 modernize examples]
        M5[M5 cli-triage example]
        M6[M6 README/godoc sync]
    end
    subgraph T3["20% → 80%: ship + trust"]
        M7[M7 toolsdk recipe]
        M8[M8 discoverability]
        M9[M9 domain diagram]
        M10[M10 vision-verify render]
        M11[M11 go-error-family audit]
        M12[M12 converge docs]
        M13[M13 release v0.5.0]
        M14[M14 consumer sweep]
    end
    subgraph TR["other 20% → 100%"]
        M15[M15 fix-representation design]
        M16[M16 Filter type]
        M17[M17 convenience trio]
        M18[M18 O(1) index]
        M19[M19 RuleSet]
        M20[M20 generics eval]
        M21[M21 tag currency]
        M22[M22 report hygiene]
        M23[M23 guard conventions]
        M24[M24 audit template]
        M25[M25 HEAD watch]
        M26[M26 cmd/ review]
    end
    M1 --> M13
    M2 --> M13
    M3 --> M13
    M4 --> M5 --> M6
    M13 --> M14
    M13 --> M26
    M21 --> M22
    M1 -.-> M15
    M15 -.-> M16
    M27[M27 HARVEST everything] --> DONE((plan closed))
    T1 --> M27
    T2 --> M27
    M14 --> M27
    TR -.-> M27
```

Critical path: **M1 → M13 → M14 → M27** (~5.5 h). Everything else parallelizes.

## Verschlimmbesserung guardrails (explicit)

- M1/M2 change consumer-visible report content → MINOR release (per the
  2026-09-09 slot rules), never slipped into a PATCH.
- M2 (Relativize) MUST be copy-vs-mutate verified from source before wiring —
  F2.1 is a hard precondition, not documentation.
- M4/M6 rewrite examples/README that pkg.go.dev froze at v0.4.1 — the frozen
  README stays stale until the release; that is accepted, not a reason to
  re-tag early (tag-on-green rule).
- M18 touches `Registry` internals under race — race test (F18.4) gates it;
  if the map index complicates `Deregister`-mid-run snapshot semantics, abort
  and keep the O(n) scan (documented) rather than break the contract.
- No history rewrites, no force pushes, daemon commits stay untouched.
- M13 tags only on a green CI run on the exact commit (v0.4.0 lesson).

## Where user answers would reorder this plan

- If you pick **examples-first** (question g2): swap M4–M6 before M1–M3; the
  tiers otherwise hold.
- If you **don't want a release soon** (g3): M13/M14 drop to the Rest tier and
  M1–M3 land unshipped on master.
- If you answer **yes to fleet rollout** (g1): M24 (template) jumps into the
  4% tier and every other repo gets its own plan from this template.

— Plan written 2026-11-19 07:54 by the orchestrating session. Point-in-time
snapshot; the living sources are TODO_LIST.md and ROADMAP.md.
