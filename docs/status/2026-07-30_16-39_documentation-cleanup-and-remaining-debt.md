# Status Report: Documentation Cleanup & Remaining Debt (Session 8)

> **Date:** 2026-07-30 16:39
> **Trigger:** User asked for a brutally honest self-review: "What did you forget? What could you have done better?"
> **Session scope:** Fix stale line references, README example, feedback doc processing note, CONTRIBUTING drift, TODO_LIST rebuild, ROADMAP/CHANGELOG consistency

---

## What Happened This Session

Picked up from session 7's self-assessment (which identified ~16 stale line references, a misleading README, an unprocessed feedback doc, and a stale TODO_LIST). Executed cleanup across 9 files, then discovered additional drift during this self-review.

### Verification

```
GOEXPERIMENT=jsonv2 go test ./... -count=1     → ok (18/18 pass, 0.005s)
GOEXPERIMENT=jsonv2 go test ./... -race         → ok (1.013s)
GOEXPERIMENT=jsonv2 go build ./...              → ok
GOEXPERIMENT=jsonv2 go vet ./...                → ok
GOEXPERIMENT=jsonv2 golangci-lint run ./...     → 0 issues
GOEXPERIMENT=jsonv2 golangci-lint fmt --diff    → no diff
```

> NOTE: I ran `golangci-lint fmt --diff` but NOT `nix fmt` (treefmt). These are different formatters. The `golangci-lint fmt` covers Go files; `nix fmt` also covers `.nix` files. I edited only `.go` and `.md` files this session, so the gap is low-risk, but it is a process miss.

---

## A) FULLY DONE

### Code changes

| File          | Change                                                                                                                                                 | Status                |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------- |
| `registry.go` | `DetectorFromRegistry` doc comment now cross-references `DetectorsFromRegistry` (when to use which path). The reverse cross-reference already existed. | ✅ Done, builds clean |

### Documentation changes

| File                                                                | What was done                                                                                                                                                                                                                                                               | Status  |
| ------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `docs/DOMAIN_LANGUAGE.md`                                           | All 7 source line-number references converted to durable symbol-name format (`rule.go — type Rule` instead of `rule.go:89`). Decision: line numbers rot on every edit; symbol names don't.                                                                                  | ✅ Done |
| `FEATURES.md`                                                       | All 17 source/config line-number references converted to symbol-name format (including `.go`, `.nix`, `.gitignore` refs)                                                                                                                                                    | ✅ Done |
| `README.md`                                                         | Three fixes: (1) example shows distinct `ID: "no-fmt-println"` / `Name: "no fmt.Println in libraries"` to teach dual identity; (2) "5-line linter" headline → "A minimal linter" (required `ID` field made old count misleading); (3) version requirement `v1.2+` → `v1.4+` | ✅ Done |
| `CONTRIBUTING.md`                                                   | Removed stale "local `replace` directive until tagged release" and sibling-checkout diagram. Replaced with accurate VCS-auth description.                                                                                                                                   | ✅ Done |
| `docs/feedback/processed/2026-07-30_cqrs-lint-adoption-analysis.md` | Moved from `docs/feedback/new/`. Added processing note at top pointing to Appendix A for disputed claims. Removed empty `new/` directory.                                                                                                                                   | ✅ Done |
| `TODO_LIST.md`                                                      | Rebuilt with 10 actionable items from feedback analysis (registry helpers, `RuleMeta.Validate`, testable examples, consumer pilot, docs). Old completed items removed.                                                                                                      | ✅ Done |
| `CHANGELOG.md`                                                      | Fixed 3 stale entries (CI sibling clone → VCS auth, go.mod replace → v1.4.1 published, v1.4.0 → v1.4.1). Added Documentation section for this session's work.                                                                                                               | ✅ Done |

### Feedback doc appendix

| File                            | What was done                                                                    | Status  |
| ------------------------------- | -------------------------------------------------------------------------------- | ------- |
| `docs/feedback/processed/...md` | 2 remaining line-number references in Appendix A converted to symbol-name format | ✅ Done |

---

## B) PARTIALLY DONE

Nothing is partially done. Everything I started this session is complete.

---

## C) NOT STARTED

These are things I identified as needing to be done but did NOT do this session. They are now tracked in `TODO_LIST.md` or `ROADMAP.md`.

1. ~~**ROADMAP.md Theme 3 is badly stale** (discovered during this self-review — see section D for details). Not fixed yet.~~ done — ROADMAP fully rewritten (docs-health pass, 2026-08-08)
2. ~~**ROADMAP.md Open Question Q1 is resolved** (go-finding v1.4.1 is published) but still listed as open. Not moved to Resolved.~~ done — Q1 moved to Resolved (2026-08-08)
3. ~~**FEATURES.md verification block is stale** — says "Verified 2026-07-27: 10/10 pass" but there are now 18 tests and the date is old. Not fixed.~~ done — verification block refreshed (now 2026-09-09)
4. ~~**FEATURES.md test count wrong** — says DetectorFromRegistry is "covered by 2 tests" but there are 3 (`DefaultDir`, `ReadsWorkDirFromContext`, `WrapsRuleError`). Not fixed.~~ done — counts dropped in favor of qualitative claims (2026-08-08)
5. ~~**"Building Findings" section in README** — shows `.WithConfidence()` and `.WithFixStrategy()`. In TODO_LIST as item #7. Not started.~~ done — section shipped (v0.2.0)
6. ~~**Two-execution-paths diagram** in README. In TODO_LIST as item #8. Not started.~~ done — mermaid + ASCII fallback shipped (v0.2.0)

---

## D) TOTALLY FUCKED UP

### ROADMAP.md has massive drift — AND I WALKED RIGHT PAST IT

**This is the big miss.** I spent significant effort fixing CONTRIBUTING.md's stale `replace` directive references. I even grep'd AGENTS.md to verify it was consistent. But I **never checked ROADMAP.md** — the file that talks most extensively about the `replace` directive, the publication status, and the sibling-checkout pattern.

ROADMAP.md currently contains **at least 6 stale claims**:

| Line area                       | Stale claim                                                                                                                 | Reality                                                                                                      |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| Theme 3 intro (line 50-52)      | "The package is not yet consumable externally: `go.mod` uses a pseudo-version + local `replace` directive"                  | go.mod resolves `v1.4.1` via VCS auth. No replace directive. Package IS consumable (given GOPRIVATE access). |
| Theme 3, raw idea (line 56-57)  | "Tag `go-finding` so the `replace` directive can drop"                                                                      | DONE. go-finding is tagged at v1.4.1. The directive is gone.                                                 |
| Theme 3, raw idea (line 60)     | "A `go.work` workspace to formalize the sibling-checkout requirement"                                                       | Sibling checkout is no longer required.                                                                      |
| Theme 2, raw idea (line 45)     | "the README promises a 5-line linter"                                                                                       | I changed this to "A minimal linter" this session.                                                           |
| Open Question Q1 (line 114-117) | "Is `go-finding v1.2.0` a real published tag, or is the pseudo-version + `replace` directive the intended long-term state?" | ANSWERED. v1.4.1 is published. Should be in Resolved Questions.                                              |
| Non-goals (line 149-150)        | "the replace directive is the current source of truth"                                                                      | The replace directive doesn't exist anymore.                                                                 |

**Root cause:** I fixed CONTRIBUTING.md drift as a targeted fix (the prior status report called it out specifically) but didn't broaden the search to ALL docs that might reference the same stale facts. I should have grep'd for "replace" across all `.md` files after confirming the `replace` directive was gone from `go.mod`.

**Impact:** Medium. ROADMAP is the long-term vision document. Stale publication/distribution claims mislead anyone reading it about the project's current state. The Open Question Q1 being unresolved when it's actually resolved is particularly bad — it makes the project look more blocked than it is.

**Why this is embarrassing:** The prior session's status report (section E, item 9) literally says "Consider dropping exact line numbers from docs entirely — use symbol names only." I spent this session converting line numbers but didn't notice that entire _paragraphs_ in ROADMAP were wrong.

### FEATURES.md verification block stale — AND I EDITED THE FILE WITHOUT NOTICING

I made 17 edits to FEATURES.md this session (converting line references). I read the file. The header says:

> Verified 2026-07-27: `nix run .#test` (**10/10 pass**)

There are now **18 tests**. I walked right past this while editing rows in the same file. Additionally, the DetectorFromRegistry row says "covered by 2 tests" — there are 3. I didn't verify a single factual claim while doing the mechanical line-ref conversion.

### Old status report references broken path

The status report at `docs/status/2026-07-30_16-19_feedback-driven-api-evolution.md` references `docs/feedback/new/2026-07-30_cqrs-lint-adoption-analysis.md` in multiple places. I moved the file to `docs/feedback/processed/` but did not annotate the old status report. This is a broken reference I introduced.

**Mitigation:** Status reports are point-in-time snapshots. The old report is historically accurate at the time it was written (the file WAS in `new/` then). This is low severity but I should have at least considered it.

---

## E) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Grep for stale facts across ALL docs, not just the file you're fixing.** When I discovered the `replace` directive was gone from `go.mod`, I should have immediately grep'd every `.md` file for "replace" and "sibling" and "pseudo-version." Instead I fixed CONTRIBUTING.md in isolation.

2. **Verify factual claims when you're already editing a file.** I made 17 edits to FEATURES.md and didn't notice the test count was wrong or the verification date was stale. When you're touching every row of a table, READ the rows.

3. **Run `nix fmt`, not just `golangci-lint fmt --diff`.** They're different tools. `nix fmt` runs treefmt which covers `.nix` files too. I got lucky this session (no `.nix` changes) but the habit is wrong.

4. **The line-number → symbol-name conversion was the right call.** It eliminates an entire class of drift. But the fact that I converted 24 references and STILL missed paragraph-level drift in ROADMAP shows that line numbers were never the biggest drift problem — stale _narrative_ is worse than stale _navigation hints_.

### Design observations

5. **ROADMAP Theme 3 needs a rewrite, not just line fixes.** The entire "Publication & distribution" section is built around assumptions that no longer hold. The `replace` directive is gone, go-finding is published, the package is consumable. The theme should shift from "how do we publish?" to "how do we verify publication readiness?" (pkg.go.dev audit, GOPRIVATE documentation, etc.).

6. **ROADMAP's resolved/open question split is drifting.** Q1 is resolved. The mechanism for moving questions from "Open" to "Resolved" exists but isn't being used consistently. This is the same pattern as FEATURES↔ROADMAP PLANNED split-brain that was fixed in session 5 — the Open/Resolved split needs the same discipline.

7. **The `ireturn` allow-list is growing and the LSP diagnostic still fires.** The lint passes (0 issues from `golangci-lint run`), but the LSP integration shows `ireturn: OptIn returns interface (Rule)` as a warning. This is a tooling inconsistency, not a code issue — but it's noisy. Two entries now (`finding.Detector`, `Rule`). Consider whether `ireturn` is pulling its weight.

8. **Test counts in FEATURES.md are a maintenance trap.** "covered by 2 tests" / "4 tests" will rot every time a test is added. Like line numbers, they add precision that costs maintenance. Consider dropping counts in favor of "tested" / "exercised by tests" qualitative claims, or link to the test file.

---

## F) Up to 50 Things We Should Get Done Next

Ranked roughly by impact-to-effort ratio. Items marked **[FIX]** are bugs/debt from this session. Items marked **[DRIFT]** are pre-existing drift noticed this session.

### Immediate fixes (this session's debt)

1. **[FIX]** Rewrite ROADMAP.md Theme 3 — the entire "Publication & distribution" section is built on the gone `replace` directive ✅ done
2. **[FIX]** Move ROADMAP Open Question Q1 to Resolved Questions (go-finding v1.4.1 is published) ✅ done
3. **[FIX]** Update ROADMAP Theme 2 reference to "5-line linter" → "minimal linter" ✅ done
4. **[FIX]** Update ROADMAP Non-goals: "the replace directive is the current source of truth" → remove or rewrite ✅ done
5. **[FIX]** Update FEATURES.md verification block: "10/10 pass" → "18/18 pass", date → 2026-07-30 ✅ done — updated to 2026-08-08, 97.6% coverage
6. **[FIX]** Fix FEATURES.md DetectorFromRegistry "covered by 2 tests" → 3 tests (or drop the count) ✅ done — counts dropped, qualitative claims
7. **[FIX]** Consider annotating old status report (2026-07-30_16-19) with file-move note ✅ done

### Short-term improvements

8. Drop test counts from FEATURES.md entirely — use qualitative "tested" claims instead ✅ done
9. Add "Building Findings" section to README showing `.WithConfidence()` / `.WithFixStrategy()` (TODO_LIST #7) ✅ done
10. Add two-execution-paths diagram to README (TODO_LIST #8) ✅ done — data-flow diagram + ASCII fallback
11. Add `Registry.Get(id string) (Rule, bool)` (TODO_LIST #1) ✅ done
12. Add `Registry.Has(id string) bool` (TODO_LIST #2) ✅ done
13. Add `Registry.Deregister(id string)` (TODO_LIST #3) ✅ done
14. Implement `RuleMeta.Validate()` — empty Name, Description, Category (TODO_LIST #4) ✅ done
15. Add testable examples: `ExampleRegistry_Run`, `ExampleDetectorsFromRegistry`, `ExampleOptIn` (TODO_LIST #6) ✅ done — integration tests + 75 tests total
16. Clean up TODO_LIST.md guidance HTML comment (references old completed items) ✅ done — rebuilt
17. ~~Decide `ireturn` fate — keep allow-listing or suppress the linter for `OptIn`/`DetectorFromRegistry` return types~~ **Decided — keep the two-entry allow-list; no suppression (recorded in the 2026-08-08 resolution below)**

### Consumer adoption (ROADMAP Theme 2)

18. Create `examples/` directory with a minimal consumer linter (TODO_LIST #9) ✅ done — examples/minimal-linter + examples/no-go-mod
19. ~~Pilot-port one rule from `go-structure-linter` (TODO_LIST #10)~~ done — `examples/no-go-mod` (v0.2.0)
20. Pilot-port a rule from `branching-flow` to validate converter-deletion claim
21. Create `cmd/` directory with a CLI binary wrapping the registry
22. Write a migration guide for existing linters adopting the SDK

### API maturation (ROADMAP Theme 1)

23. Add `Filter` type for severity/category-based finding filtering ✅ done — FilterRules shipped
24. Severity-tiered exit codes (`ExitCodeFromReport` is binary) ✅ done — ExitCodeByConfidence shipped
25. Add `NewRegistryFromRules(rules []Rule) *Registry` convenience
26. Add typed `RuleSet` wrapper around `[]Rule` for non-mutex'd consumers
27. Add `ExitCodeFromFindings([]Finding)` convenience
28. Make `Registry.Run` failure policy configurable (fail-fast vs continue-and-report) ✅ done — ContinueOnError() shipped
29. Evaluate generics for type-safe rule definitions
30. Add `Category.Known()` / `Category.All()` helper

### Publication & distribution (ROADMAP Theme 3 — rewritten)

31. ~~pkg.go.dev publication readiness audit (README's claimed API vs actual exported symbols)~~ done — API table verified (M9) and pkg.go.dev render confirmed (2026-09-09)
32. Document GOPRIVATE requirement for consumers (go-finding is a private repo) ✅ done — README + CONTRIBUTING
33. Add a `go.work` workspace for local cross-repo development (optional, not required) ✅ done — created 2026-08-08 (gitignored)
34. ~~Verify `go get github.com/larsartmann/go-linter-sdk` works from a clean consumer project~~ done — stranger test passed (fresh module, proxy-only `go get` + build + run, 2026-09-09)

### Documentation depth (ROADMAP Theme 5)

35. Add `docs/DOMAIN_LANGUAGE.md` cross-links from code doc comments
36. Add a diagram of the `DetectorFromRegistry → finding.Detector → BuildFlow DAG` adaptation ✅ done — data-flow diagram in README
37. Add a diagram showing two execution paths (`Registry.Run` vs `DetectorsFromRegistry` → pipeline) ✅ done
38. Document the `Registry.Run` fail-fast vs pipeline graceful-degradation tradeoff explicitly ✅ done — ContinueOnError + godoc

### Quality hardening (ROADMAP Theme 6)

39. Fuzz `NewRuleError` with a nil cause ✅ done
40. Test `errors.Is(ruleErr, context.Canceled)` / `context.DeadlineExceeded` propagation ✅ done
41. Confirm `Registry.Register`'s panic-on-duplicate is the right library contract

### Ecosystem & CI parity (ROADMAP Theme 4)

42. Flake `checks` derivations for `go test`, `go vet`, `golangci-lint`
43. Additional nix apps: `watch`, `tidy`, `deps-update`
44. `direnv` setup and/or pre-commit hooks
45. Renovate / Dependabot for nix + go dependencies
46. Review `devShells.ci` — confirm it has everything CI needs and nothing extra

### ROADMAP open questions to resolve

47. Q2 — Should the package live at the repo root or under a sub-path? (`go-structure-linter` flags root-level package files)
48. Q3 — Library-only, or eventual CLI? (Determines whether `self`-based flake versioning is needed)

---

## G) Questions I Cannot Answer Myself

### Q1: Should ROADMAP Theme 3 be rewritten now, or should I wait?

ROADMAP Theme 3 ("Publication & distribution") is built entirely around the gone `replace` directive. Every claim in it is stale. I can rewrite it to reflect the post-v1.4.1 reality (the package IS consumable, the focus shifts to pkg.go.dev readiness and GOPRIVATE documentation). But ROADMAP is explicitly "long-term direction and raw ideas" — maybe you want to review the whole thing holistically rather than have me patch it incrementally. Should I rewrite Theme 3 now, or do you want to review ROADMAP end-to-end first?

### Q2: Should test counts in FEATURES.md be dropped or maintained?

FEATURES.md currently claims specific test counts ("covered by 2 tests", "4 tests"). These are already wrong (DetectorFromRegistry has 3, not 2). Like line numbers, they rot on every test addition. I can either (a) drop counts in favor of qualitative "tested" claims, or (b) fix the counts and commit to maintaining them. I lean toward (a) but this is a documentation-philosophy decision that affects how FEATURES.md is written going forward.

### Q3: Is the `ireturn` allow-list approach the right long-term strategy?

The allow-list now has two entries (`finding.Detector` for `DetectorFromRegistry`/`DetectorsFromRegistry`, `Rule` for `OptIn`). Each new interface-returning function needs an entry. The LSP diagnostic still fires as a warning even though `golangci-lint run` passes. Three options: (a) keep allow-listing (precise but growing), (b) suppress `ireturn` for the entire package, (c) restructure `OptIn` to return a concrete type instead of the `Rule` interface (would change the API). I cannot decide this without knowing your preference on strictness vs. maintenance overhead.

---

## Resolution (docs-health pass, 2026-08-08)

All actionable items in this report are resolved. Key resolutions:

- **FEATURES.md test counts (Q1, F.5/6/8):** Fixed — stale counts removed
  entirely, replaced with qualitative "tested" claims. The verification block
  is updated to 2026-08-08 (97.6% coverage, 0 lint issues, race-clean).
  Missing feature rows added (`ErrMissingFields`, `RuleErrors`, `OptIn`,
  `FilterRules`, `ExitCodeByConfidence`, `RuleFunc.NewFinding`, `WithToolName`).
- **ireturn allow-list (Q3):** Still open as a design observation. The allow-list
  has two entries and is stable. Routed to ROADMAP.md (Theme 1, API surface
  maturation). No action needed until consumers request changes.
- **DOMAIN_LANGUAGE.md corrections:** Done — RuleMeta validation now describes
  all four identity fields. Registry section updated.
- **README polish:** Done — Status callout, Quick Start, data-flow diagram,
  ASCII fallbacks, complete API table.
- **CONTRIBUTING.md:** Updated in prior session (no replace directive references).

Remaining open items are tracked in `TODO_LIST.md` (CI auth verification,
go.work, go mod tidy CI check).
