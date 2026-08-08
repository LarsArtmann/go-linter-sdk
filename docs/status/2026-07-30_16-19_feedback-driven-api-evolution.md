# Status Report: Feedback-Driven API Evolution (Session 7)

> **Date:** 2026-07-30 16:19
> **Trigger:** User asked to deeply review `docs/feedback/new/2026-07-30_cqrs-lint-adoption-analysis.md`
> **Outcome:** Wrote decision appendix, implemented 3 API changes, updated 6 docs, introduced stale line references

---

## What Happened This Session

1. Deeply reviewed the cqrs-lint adoption analysis feedback document
2. Cross-referenced every claim against actual source (`rule.go`, `registry.go`, `errors.go`, `go-finding/*`)
3. Wrote a comprehensive Appendix A to the feedback document with accept/reject decisions on all 6 proposed changes
4. User answered 8 design questions; decisions: reject `any` context, accept `DetectorsFromRegistry`, accept `ID()`, document-don't-add Confidence/AutoFix, Category is already open
5. Implemented three changes in one sweep: `DetectorsFromRegistry`, `ID()` on Rule, documentation improvements
6. All tests pass (test, test-race, lint 0 issues, build, vet)

---

## A) FULLY DONE

### Code changes

| File               | Change                                                                                                                                                                                                                                                                           | Status                      |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------- |
| `rule.go`          | Added `ID() string` to `Rule` interface (required, first method). Added `ID` field to `RuleMeta` (required, first field). Added doc comments for dual identity, Category-is-open, finding-level Confidence/FixStrategy. `RuleFunc.Check` now wraps errors with `r.Meta.ID`.      | ✅ Done, tested, lint clean |
| `registry.go`      | `Register` deduplicates on `ID()` and panics on empty IDs. Added `DetectorsFromRegistry(registry) []finding.Detector`. `wrapRuleError` parameter renamed to `ruleID`. All error attribution uses `rule.ID()`.                                                                    | ✅ Done, tested, lint clean |
| `errors.go`        | `RuleError.RuleName` → `RuleError.RuleID`. `NewRuleError` parameter renamed. Doc comments updated.                                                                                                                                                                               | ✅ Done, tested, lint clean |
| `registry_test.go` | All factories (`makeRule`, `failingRule`) and inline literals updated with `ID`. `RuleName` → `RuleID` in assertions. Added 5 new tests: duplicate-ID-panics, empty-ID-panics, `DetectorsFromRegistry` (count/naming, working-dir-from-context, error wrapping, empty registry). | ✅ Done, all pass           |
| `.golangci.yml`    | Added `github.com/larsartmann/go-linter-sdk.Rule` to `ireturn` allow-list for `OptIn`.                                                                                                                                                                                           | ✅ Done, lint clean         |

### Documentation changes

| File                                                          | What was updated                                                                                                                                               | Status                                        |
| ------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| `docs/feedback/new/2026-07-30_cqrs-lint-adoption-analysis.md` | Added Appendix A: Maintainer Response & Decisions with full PRO/CONTRA for each of 6 proposed changes                                                          | ✅ Done                                       |
| `CHANGELOG.md`                                                | Added `[Unreleased]` entries for all breaking and non-breaking changes                                                                                         | ✅ Done                                       |
| `DOMAIN_LANGUAGE.md`                                          | Updated terms: Rule (dual identity), RuleFunc, RuleMeta (ID required), Category (open type), Registry (dedup on ID), RuleError (RuleID), relationships diagram | ✅ Done (but line refs stale — see section D) |
| `README.md`                                                   | Updated example with `ID` field, added pipeline section, updated API tables, updated design notes                                                              | ✅ Done                                       |
| `FEATURES.md`                                                 | Updated all rows for ID(), DetectorsFromRegistry, Category open type, dedup-on-ID                                                                              | ✅ Done (but line refs stale — see section D) |
| `ROADMAP.md`                                                  | Marked `Registry.RunParallel` as superseded by `DetectorsFromRegistry`, updated `RuleMeta.Validate` as partially done                                          | ✅ Done                                       |
| `AGENTS.md`                                                   | Updated architecture section with ID(), DetectorsFromRegistry, Category open type, Confidence/FixStrategy per-finding, error wrapping uses RuleID              | ✅ Done                                       |

### Verification

```
nix run .#test       → ok (0.003s)
nix run .#test-race  → ok (1.019s)
nix run .#lint       → 0 issues
nix run .#build      → ok
nix run .#vet        → ok
```

---

## B) PARTIALLY DONE

### DOMAIN_LANGUAGE.md line references — STALE

I updated the _content_ of every term entry but did NOT update the `Lives at:` line references. The doc comments I added to `rule.go` (the dual-identity section, the building-findings section) shifted every line number downstream. The following references are wrong:

| Documented       | Actual           | Symbol               |
| ---------------- | ---------------- | -------------------- |
| `rule.go:56`     | `rule.go:89`     | `Rule` interface     |
| `rule.go:68`     | `rule.go:102`    | `RuleFunc` struct    |
| `rule.go:75`     | `rule.go:113`    | `RuleMeta` struct    |
| `rule.go:37`     | `rule.go:43`     | Category const block |
| `registry.go:57` | `registry.go:62` | `wrapRuleError`      |

### FEATURES.md line references — STALE

Same problem. All `Lives at:` references that I updated with new content still carry old line numbers:

| Documented        | Actual            | Symbol                        |
| ----------------- | ----------------- | ----------------------------- |
| `rule.go:56`      | `rule.go:89`      | `Rule` interface              |
| `rule.go:68`      | `rule.go:102`     | `RuleFunc` struct             |
| `rule.go:75`      | `rule.go:113`     | `RuleMeta` struct             |
| `registry.go:92`  | `registry.go:100` | `DetectorFromRegistry`        |
| `registry.go:114` | `registry.go:133` | `DetectorsFromRegistry`       |
| `registry.go:153` | `registry.go:164` | `ExitCodeFromReport`          |
| `registry.go:58`  | `registry.go:63`  | `errors.AsType` call          |
| `registry.go:70`  | `registry.go:75`  | `Registry.Run`                |
| `registry.go:20`  | `registry.go:20`  | `NewRegistry` (still correct) |
| `registry.go:30`  | `registry.go:30`  | `Register` (still correct)    |
| `registry.go:44`  | `registry.go:49`  | `All`                         |

---

## C) NOT STARTED

1. **TODO_LIST.md was NOT updated.** It still says "The backlog is clear" with all old items struck through. The new work items from this session (stale line refs, future improvements from the feedback analysis) should be harvested into TODO_LIST.md.
2. **No correction note at top of original feedback document.** The original sections 1-7 still make claims ("Category constants are too restrictive", "no concept of fixability") that Appendix A explicitly calls out as factually wrong. A reader scanning the document top-to-bottom sees the wrong claims first with no indication they're disputed later.
3. **CONTRIBUTING.md not checked for drift.** It still references "local replace directive until go-finding has a tagged release" and shows the sibling checkout requirement. AGENTS.md was already updated (prior session) to say go-finding v1.4.1 is published. Pre-existing drift, not introduced this session, but I should have noticed it.
4. **README "5-line linter" headline is now misleading.** Adding the required `ID` field makes the RuleMeta literal larger. The claim should be updated or removed.

---

## D) TOTALLY FUCKED UP

### Stale line references — MY BUG

**This is the main thing I got wrong.** I added significant doc comment blocks to `rule.go` (the dual-identity section is ~20 lines, the building-findings section is ~10 lines). This shifted every subsequent line number. I then updated the _prose_ in `DOMAIN_LANGUAGE.md` and `FEATURES.md` but copied the OLD line references without verifying them against the new file.

This is exactly the kind of docs-drift bug that the project's own AGENTS.md warns about: "When code and this file disagree, the code is the source of truth — fix this file." I introduced disagreement.

**Root cause:** I should have re-read the modified files (or at least `grep -n` the symbol locations) AFTER the edits and BEFORE writing doc references. Instead I wrote doc references from memory of the pre-edit line numbers.

**Impact:** Low severity (line references are navigation hints, not functional), but high embarrassment (the project explicitly tracks this as an anti-pattern).

**Fix:** Mechanical — re-`grep` all symbols and update all `Lives at:` references in both files. ~15 references to fix.

### README example misses the teaching moment

The updated example uses `ID: "no-fmt-println"` and `Name: "no-fmt-println"` — identical values. The whole point of dual identity is that they CAN differ. The example should show something like `ID: "no-fmt-println"` and `Name: "no fmt.Println in libraries"` to demonstrate the concept. Right now it looks like a redundant field.

---

## E) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Line-reference verification protocol.** After ANY edit to a `.go` file that shifts line numbers, re-`grep` all symbols referenced in docs and update them. This should be a checklist item before declaring a task complete.
2. **Line references in docs are fragile.** Consider dropping exact line numbers from docs entirely and using symbol names only (`rule.go` near `type Rule interface`). Line numbers rot on every edit; symbol names don't.
3. **Feedback processing protocol.** When processing a feedback document, add a correction note at the TOP pointing to the response appendix, so readers see disputed claims flagged immediately.
4. **"5-line linter" claim should be tested.** Any headline claim in README should be verified against the actual example code after every API change.

### Design improvements identified but NOT implemented

5. **`Registry.Run` failure policy is still fail-fast.** The ROADMAP (Theme 6) asks whether to continue and return partial results. `DetectorsFromRegistry` solves this for the pipeline path (the pipeline has `GracefulDegradation`), but `Registry.Run` still stops on the first rule error. This is now a sharper gap: the pipeline path degrades gracefully, the standalone path doesn't.
6. **No `Registry.Get(id)` / `Registry.Has(id)` helpers.** With `ID()` now the canonical key, consumers will want to look up rules by ID. ROADMAP Theme 1 already lists this.
7. **No `Registry.Deregister(id)`.** Same reasoning.
8. **`RuleMeta.Validate()` is still not implemented.** Empty-ID panics at registration (done this session), but there's no validation for empty Name, empty Description, or empty Category. ROADMAP Theme 1.
9. **No examples in `examples/` directory.** The README promises a pattern but there's no runnable example. ROADMAP Theme 2.
10. **No `ExampleRegistry_Run` / `ExampleDetectorsFromRegistry` testable examples.** ROADMAP Theme 5.
11. **The `ireturn` allow-list approach is growing.** We now allow `finding.Detector` and `Rule`. Each new interface-returning function needs an allow-list entry. Consider whether `ireturn` is pulling its weight or just creating maintenance overhead.

### Architecture observations from the feedback analysis

12. **The SDK now has two execution paths with different capabilities.** `Registry.Run` (sequential, fail-fast, no parallelism) vs `DetectorsFromRegistry` → pipeline (parallel, graceful degradation, timeouts, error isolation). The README and docs should make this tradeoff explicit so consumers choose the right path.
13. **`DetectorFromRegistry` (singular) is now the "simple but limited" path.** Its doc comment should cross-reference `DetectorsFromRegistry` and explain when to use which.
14. **The feedback identified that `Category` was perceived as closed despite being open.** This is a documentation problem, not a type problem. The fix (doc comment) is done, but the perception may recur — consider whether the const block layout itself communicates openness poorly.
15. **The feedback's `Confidence`/`AutoFix` complaint was really about discoverability.** The per-finding `finding.Builder` API exists but isn't surfaced from the SDK's docs. The doc comment on `Rule` now mentions it, but a dedicated "Building Findings" guide would be more discoverable.

---

## F) Up to 50 Things We Should Get Done Next

Ranked roughly by impact-to-effort ratio. Items marked **[FIX]** are bugs introduced this session. Items marked **[DRIFT]** are pre-existing drift noticed this session.

### Immediate fixes (this session's debt)

1. **[FIX]** Fix all stale line references in `DOMAIN_LANGUAGE.md` (5 references wrong)
2. **[FIX]** Fix all stale line references in `FEATURES.md` (~11 references wrong)
3. **[FIX]** Update README example to show different ID and Name values (teaching moment)
4. **[FIX]** Update or remove "5-line linter" headline in README
5. **[FIX]** Add correction note at top of original feedback document pointing to Appendix A
6. **[FIX]** Update `TODO_LIST.md` with new open items from this session

### Short-term improvements

7. Add cross-reference from `DetectorFromRegistry` doc comment to `DetectorsFromRegistry` (when to use which)
8. Add a "Building Findings" section to README showing `.WithConfidence()` and `.WithFixStrategy()`
9. Consider dropping exact line numbers from docs entirely — use symbol names only
10. Add `Registry.Get(id string) (Rule, bool)` helper
11. Add `Registry.Has(id string) bool` helper
12. Add `Registry.Deregister(id string)` helper
13. Implement `RuleMeta.Validate()` (check Name non-empty, Category non-empty, Description non-empty)
14. Add testable examples: `ExampleRegistry_Run`, `ExampleDetectorsFromRegistry`, `ExampleOptIn`
15. Update `CONTRIBUTING.md` — remove stale "replace directive until tagged release" language
16. **[DRIFT]** CONTRIBUTING.md still shows sibling checkout requirement; AGENTS.md says go-finding v1.4.1 is published — reconcile
17. **[DRIFT]** README says "Requires go-finding v1.2+"; AGENTS.md says v1.4.1 — reconcile
18. Add a diagram showing the two execution paths (Registry.Run vs DetectorsFromRegistry → pipeline)
19. Document the `Registry.Run` fail-fast vs pipeline graceful-degradation tradeoff explicitly
20. Move feedback file out of `docs/feedback/new/` since it's been processed

### Consumer adoption (ROADMAP Theme 2)

21. Create `examples/` directory with a minimal consumer linter
22. Pilot-port one rule from `go-structure-linter` to prove the Rule → finding.Finding path
23. Create `cmd/` directory with a CLI binary wrapping the registry
24. Pilot-port a rule from `branching-flow` to validate the converter-deletion claim
25. Write a migration guide for existing linters adopting the SDK

### API maturation (ROADMAP Theme 1)

26. Add a `Filter` type for severity/category-based finding filtering
27. Severity-tiered exit codes (`ExitCodeFromReport` is binary)
28. Add `NewRegistryFromRules(rules []Rule) *Registry` convenience
29. Add a typed `RuleSet` wrapper around `[]Rule` for non-mutex'd consumers
30. Add `ExitCodeFromFindings([]Finding) convenience
31. Decide whether `RuleFunc` should use generics for type-safe rule definitions
32. Make `Registry.Run` failure policy configurable (fail-fast vs continue-and-report)
33. Consider a `Category.All()` or `Category.Known()` helper

### Publication & distribution (ROADMAP Theme 3)

34. Verify `go.mod` no longer needs the `replace ../go-finding` directive (AGENTS.md says v1.4.1 is published)
35. pkg.go.dev publication readiness audit (README's claimed API vs actual exported symbols)
36. Add a `go.work` workspace to formalize the sibling-checkout development pattern
37. Verify `self`-based flake versioning is correct for a library-only package

### Quality hardening (ROADMAP Theme 6)

38. Fuzz `NewRuleError` with a nil cause
39. Test `errors.Is(ruleErr, context.Canceled)` / `context.DeadlineExceeded` propagation
40. Confirm `Registry.Register`'s panic-on-duplicate is the right contract for a library
41. Test that `DetectorsFromRegistry` closures don't capture stale registry state (mutation after call)
42. Benchmark `DetectorsFromRegistry` (allocation per detector)
43. Test `DetectorsFromRegistry` with registry mutation after call (snapshot semantics)

### Documentation depth (ROADMAP Theme 5)

44. Package-level examples visible on pkg.go.dev
45. A diagram of `DetectorsFromRegistry → pipeline.New → pipeline.Run`
46. Update `docs/DOMAIN_LANGUAGE.md` with a term entry for `DetectorsFromRegistry` as a concept
47. Document the `finding.Builder` chain as the canonical way to construct findings

### Ecosystem & CI parity (ROADMAP Theme 4)

48. Flake `checks` derivations for `go test`, `go vet`, and `golangci-lint`
49. Additional nix apps: `watch` (live test re-runs), `tidy`, `deps-update`
50. Renovate / Dependabot for nix + go dependencies

---

## G) Questions I Cannot Answer Myself

### Q1: Should docs use exact line numbers at all?

Line references (`rule.go:89`) rot on every edit — I just proved this by introducing ~16 stale references in one session. Two alternatives:

- **(a)** Drop line numbers, use symbol names only: "Lives at: `rule.go`, `type Rule interface`"
- **(b)** Keep line numbers but add a CI check / pre-commit hook that verifies them

Option (a) is zero-maintenance but slightly less navigable. Option (b) is precise but adds tooling overhead. Which approach do you want? This determines whether I fix the stale references or delete them.

### Q2: Should we move the processed feedback document out of `docs/feedback/new/`?

The `new/` directory implies "unprocessed." The document now has an appendix with decisions — it's processed. Should I move it to `docs/feedback/` (parent directory), or leave it in `new/` with the understanding that the appendix marks it as processed? There's no established convention for this yet.

### Q3: The `go.mod` still has `replace github.com/larsartmann/go-finding => ../go-finding`, but AGENTS.md says "go-finding v1.4.1 — resolved from VCS as a real published tag (no local replace directive)." Is the replace directive stale and should it be dropped, or is it intentionally kept for local development?

I cannot resolve this myself because it depends on whether go-finding v1.4.1 is actually fetchable from the Go proxy, which I can't verify without network access to `proxy.golang.org`. If the tag exists and is fetchable, the replace directive should be dropped from `go.mod` (it confuses external consumers). If it's not yet fetchable, AGENTS.md has drift.

---

## Resolution (docs-health pass, 2026-08-08)

All actionable items in this report are resolved. Key resolutions:

- **Feedback document (Q1):** Moved to `docs/feedback/processed/` with a
  processing note pointing to the appendix.
- **go.mod replace directive (Q3):** Resolved — `replace` directive removed.
  `go-finding` v1.4.1 is a real published tag resolved via VCS auth. AGENTS.md
  is accurate.
- **API additions from feedback:** All shipped — `Get`/`Has`/`Deregister`,
  `RuleMeta.Validate`, `ContinueOnError`, `OptIn`, `IsEnabledByDefault`,
  `FilterRules`, `ExitCodeByConfidence`, `ErrMissingFields`, `RuleErrors`,
  `RuleFunc.NewFinding`, `WithToolName`.
- **README updates:** Dual identity (ID vs Name) documented. Execution-paths
  diagram added. Data-flow diagram added. API table completed.
- **Test coverage:** Integration tests for example binaries, fuzz tests,
  context error propagation tests, concurrent Deregister test — all added.
  Coverage at 97.6%.

No open items remain from this report. Forward-looking ideas are in ROADMAP.md.
