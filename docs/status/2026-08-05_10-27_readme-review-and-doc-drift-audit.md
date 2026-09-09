# Status Report: README Review & Doc-Drift Audit

**Date:** 2026-08-05 10:27 CEST
**Session scope:** Evaluate README quality, identify the real problem the project solves, improve the README, audit what was missed.

---

## The Real Problem This Project Solves

Every Go linter in a `finding`-based ecosystem reinvents three layers — rule
contract, registry, and a **converter** from its own issue type to
`finding.Finding`. The converter is the expensive one: thousands of LOC touched
on every finding-schema change, in every linter, forever. `go-linter-sdk`
standardizes the rule/execution contract and **makes the converter unnecessary**
— rules emit `finding.Finding` directly.

**Honest caveat:** this value is **unproven** — zero consumers have migrated
(ROADMAP confirms, README now states it).

---

## a) FULLY DONE (this session)

1. **README opening rewritten.** Replaced the insider run-on ("Shared
   scaffolding for LarsArtmann Go linters…") with a plain-language tagline:
   _"A small Go library for building linters that plug into a finding-based
   ecosystem."_ Followed by a 3-sentence paragraph that names the universal
   problem and the core design decision.

2. **"Why?" section reframed.** Now leads with the **universal** problem (a
   generic 3-row table every finding-based linter can relate to), then uses the
   LarsArtmann LOC table as concrete evidence. Previously the LOC table was the
   first thing a reader saw, with no framing.

3. **Private-dependency honesty added.** Installation section now warns that
   `go-finding` is private, requires `GOPRIVATE=github.com/larsartmann/*` + auth,
   and explains CI vs. local dev setup. Previously `go get` silently failed for
   everyone.

4. **Minimal-linter example made runnable.** Replaced `// ... scan dir ...`
   placeholder with a real `finding.NewBuilder(...)` call. Added a pointer to
   `examples/minimal-linter/` with the verified run command.

5. **Build + test verified.** `GOEXPERIMENT=jsonv2 go build ./...` passes.
   `GOEXPERIMENT=jsonv2 go test ./... -count=1` passes (all tests green).
   `examples/minimal-linter` verified: exit 0 with README, exit 1 + finding
   without.

---

## b) PARTIALLY DONE (this session)

1. **README quality pass.** The four highest-leverage improvements are done, but
   the README was already mid-evolution when I started (it gained
   Confidence/FixStrategy tables, a mermaid diagram, and updated API tables from
   a concurrent/prior edit). I did not audit the **entire** README end-to-end for
   consistency after my edits merged with those additions.

---

## c) NOT STARTED (noticed but not addressed this session)

These are things I **noticed during this session** are broken/stale but did not
fix because they were outside the immediate README task:

### Doc drift (ROADMAP.md is severely stale)

1. ~~**ROADMAP Theme 1** lists "Registry helpers: `NewRegistryFromRules`,
   `Deregister`, `Has`, `Get`" as raw ideas. Reality: `Get`, `Has`,
   `Deregister` are **all implemented** (`registry.go:63-106`), confirmed in
   CHANGELOG `[Unreleased]`. Only `NewRegistryFromRules` remains unbuilt.~~
   done — ROADMAP rewritten; done items removed, `NewRegistryFromRules` kept
   as a raw idea (docs-health pass, 2026-08-08)

2. ~~**ROADMAP Theme 1** says `RuleMeta.Validate` is "partially done: empty-ID
   panics at registration. Full `Validate` (category non-empty, description
   non-empty) is still a raw idea." Reality: `validateRuleIdentity`
   (`rule.go:158-176`) checks **all four** fields (ID, Name, Description,
   Category). `RuleMeta.Validate()` (`rule.go:134-152`) also checks all four.
   This is **fully done**.~~
   done — ROADMAP rewritten; `RuleMeta.Validate` done items removed
   (docs-health pass, 2026-08-08)

3. ~~**ROADMAP Theme 2** says "A `cmd/` directory with a CLI binary wrapping the
   registry (the README promises a 5-line linter; no `cmd/` exists yet)."
   Reality: `examples/minimal-linter/` exists and is a complete runnable CLI.
   Referenced in CHANGELOG. The README no longer promises "5 lines" but the
   ROADMAP item is stale.~~
   done — ROADMAP Theme 2 rewritten; stale "5-line" reference removed,
   `cmd/` kept as a production-CLI raw idea (docs-health pass, 2026-08-08)

4. ~~**ROADMAP Theme 2** says "An `examples/` directory with a minimal consumer
   linter" as a raw idea. Reality: **two** examples exist
   (`examples/minimal-linter/`, `examples/no-go-mod/`).~~
   done — ROADMAP rewritten; examples items removed (docs-health pass, 2026-08-08)

5. ~~**ROADMAP Theme 3** says "The package is not yet consumable externally:
   `go.mod` uses a pseudo-version + local `replace` directive pointing at a
   sibling `go-finding` checkout." Reality: `go.mod` shows
   `require github.com/larsartmann/go-finding v1.4.1` — a **real published tag**,
   **no replace directive**. The package IS externally resolvable (modulo
   private-repo auth). This entire theme is outdated.~~
   done — ROADMAP Theme 3 rewritten entirely; false premise corrected
   (docs-health pass, 2026-08-08)

6. ~~**ROADMAP Theme 6** says "Decide `Registry.Run` failure policy: fail-fast
   (today) vs. continue and return partial results — make it explicit or
   configurable." Reality: `ContinueOnError()` option exists (`registry.go:134`),
   documented, tested, in CHANGELOG. **Decided and shipped.**~~
   done — ROADMAP rewritten; failure-policy item removed (docs-health pass,
   2026-08-08)

7. ~~**ROADMAP Q1 (Open)** asks "Is `go-finding v1.2.0` a real published tag…"
   Reality: `go-finding v1.4.1` is a real published tag (per `go.mod` and
   AGENTS.md). **Resolved.** Should move to "Resolved questions."~~
   done — Q1 moved to Resolved questions; confirmed v1.4.1 is real
   (docs-health pass, 2026-08-08)

8. ~~**ROADMAP Theme 5** says "Package-level examples (`ExampleRegistry_Run`,
   `ExampleRuleFunc`) visible on pkg.go.dev" as a raw idea. Reality: four
   testable examples exist in `example_test.go`
   (`ExampleRegistry_Run`, `ExampleDetectorsFromRegistry`, `ExampleOptIn`,
   `ExampleRegistry_Run_continueOnError`). **Done.**~~
   done — ROADMAP Theme 5 (Documentation depth) removed entirely; all items
   were done (docs-health pass, 2026-08-08)

### Doc drift (DOMAIN_LANGUAGE.md)

9. ~~**DOMAIN_LANGUAGE.md RuleMeta section** says: "Validation: empty ID panics at
   registration. Other metadata (invalid category, empty description) is not yet
   validated. Tracked in ROADMAP.md Theme 1." Reality: **all four identity
   fields** are validated at registration via `validateRuleIdentity`
   (`rule.go:158`). The domain language doc is **factually wrong**.~~
   done — DOMAIN_LANGUAGE.md corrected; RuleMeta validation now accurately
   describes all four identity fields (docs-health pass, 2026-08-08)

### Doc drift (FEATURES.md)

10. ~~**FEATURES.md** README row description: "Sales page: why, install, usage, API
    table, migration path." Doesn't mention the new sections (Confidence/
    FixStrategy tables, mermaid execution-path diagram, examples pointer,
    private-dependency note). Minor but stale.~~
    done — FEATURES.md README row updated with all new sections; CI row also
    fixed (docs-health pass, 2026-08-08)

### Verification gap

11. ~~**I did not run `nix run .#lint`** (golangci-lint). I ran `go build` and
    `go test` only. The project has 14 LSP warnings (pre-existing: `wsl_v5`,
    `err113`, `nlreturn`, `exhaustruct`, `bloop`). I don't know if the lint app
    passes or fails. Per AGENTS.md, `nix run .#lint` is the canonical check.~~
    done — the lint habit was fixed in later sessions; `nix run .#lint` reports 0 issues (re-verified 2026-09-09; the `bloop` hints died in M1)

12. ~~**CHANGELOG.md** `[Unreleased]` section does not mention the README
    improvements made this session.~~
    done — CHANGELOG.md `[Unreleased]` restructured; README improvements now
    documented under Added and Documentation (docs-health pass, 2026-08-08)

---

## d) TOTALLY FUCKED UP

1. ~~**ROADMAP.md is the most stale file in the repo.** At least **8 items** listed
   as "raw ideas," "open questions," or "not yet implemented" are **already done
   in code** and logged in CHANGELOG. This is exactly the "doc drift split-brain"
   the AGENTS.md warns about. The ROADMAP has not been reconciled with the
   CHANGELOG or the code since at least session 7 (when Get/Has/Deregister/
   ContinueOnError/Validate/examples all shipped).~~
   done — ROADMAP.md fully rewritten; all stale items removed, false claims
   corrected, Q1 resolved (docs-health pass, 2026-08-08)

2. ~~**DOMAIN_LANGUAGE.md contains a factual lie** about validation completeness.
   A fresh reader relying on this glossary would believe they need to add
   validation that already exists.~~
   done — DOMAIN_LANGUAGE.md corrected; RuleMeta validation now accurately
   describes all four identity fields (docs-health pass, 2026-08-08)

3. ~~**I didn't catch the doc drift during my README work.** I read ROADMAP and
   DOMAIN_LANGUAGE early in the session (for context) but did not flag the drift
   until the self-review prompt forced me to re-examine. A truly excellent
   session would have caught and fixed these alongside the README.~~
   done — docs-health pass executed this session (2026-08-08)

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Always run `nix run .#lint`, not just `go build`/`go test`.** The AGENTS.md
   documents it as canonical. I skipped it. That's a process failure.

2. **Reconcile docs against code after every feature batch.** The Get/Has/
   Deregister/ContinueOnError/Validate/examples batch shipped without updating
   ROADMAP or DOMAIN_LANGUAGE. The one-home rule (AGENTS.md) exists but isn't
   enforced after code changes.

3. **Run a docs-health pass.** This repo has a `docs-health` skill available. The
   ROADMAP/DOMAIN_LANGUAGE drift is exactly what it's designed to catch.

### README improvements still possible

4. **The "Status" section is buried at the bottom.** For a project with zero
   consumers, the "Early. No linter has been ported yet." caveat should be more
   prominent — perhaps a callout near the top, not line 169.

5. **The README is LarsArtmann-centric.** The "Why?" table and "Consumers"
   section name three specific internal repos. If this SDK is meant to be
   reusable beyond LarsArtmann, the framing should generalize. If it's
   intentionally internal, that's fine — but the pkg.go.dev badge and
   `go get` instructions imply external consumption.

6. **No "How it works" diagram.** The mermaid diagram shows execution paths but
   not the data flow (Rule → finding.Finding → Report → ExitCode). A reader who
   doesn't know `go-finding` can't picture the pipeline.

7. **The migration path is aspirational.** Step 5 says "delete the converter
   package" but no linter has actually done this. The README should either mark
   this as "untested" or remove it until proven.

8. **`examples/no-go-mod/` is not mentioned in the README at all.** It's a pilot
   port of a real `go-structure-linter` rule — arguably more compelling evidence
   than `minimal-linter`.

---

## f) Up to 50 things we should get done next

Ranked by impact:

### Doc-drift cleanup (HIGH — these are factual errors)

> Items 1–12 all resolved in the docs-health pass (2026-08-08). See section c)
> above for the inline `done` markers on each. ROADMAP.md was fully rewritten,
> DOMAIN_LANGUAGE.md and FEATURES.md corrected, CHANGELOG `[Unreleased]`
> restructured, and TODO_LIST.md rebuilt with harvested items.

1. ~~Update ROADMAP Theme 1: mark `Get`/`Has`/`Deregister` as done; remove from
   raw ideas~~ done (docs-health pass, 2026-08-08)
2. ~~Update ROADMAP Theme 1: mark `RuleMeta.Validate` (full) as done~~
   done (docs-health pass, 2026-08-08)
3. ~~Update ROADMAP Theme 2: mark `examples/` directory as done (two examples
   exist)~~ done (docs-health pass, 2026-08-08)
4. ~~Update ROADMAP Theme 2: mark `cmd/` CLI as done (minimal-linter exists) or
   clarify it means a production `cmd/` binary~~ done (docs-health pass, 2026-08-08)
5. ~~Update ROADMAP Theme 3: remove "pseudo-version + replace directive" claim;
   mark go-finding v1.4.1 as a real published tag~~
   done (docs-health pass, 2026-08-08)
6. ~~Update ROADMAP Theme 6: mark "Registry.Run failure policy" as decided and
   shipped (`ContinueOnError`)~~ done (docs-health pass, 2026-08-08)
7. ~~Move ROADMAP Q1 from "Open" to "Resolved" (go-finding v1.4.1 is real)~~
   done (docs-health pass, 2026-08-08)
8. ~~Update ROADMAP Theme 5: mark package-level examples as done (4 exist)~~
   done (docs-health pass, 2026-08-08)
9. ~~Fix DOMAIN_LANGUAGE.md: validation now checks all 4 fields, not just ID~~
   done (docs-health pass, 2026-08-08)
10. ~~Update FEATURES.md README row to mention new sections~~
    done (docs-health pass, 2026-08-08)
11. ~~Add CHANGELOG `[Unreleased]` entry for README improvements~~
    done (docs-health pass, 2026-08-08)
12. ~~Run full `docs-health` skill pass to catch any remaining drift~~
    done (docs-health pass, 2026-08-08)

### README polish (MEDIUM)

13. ~~Add "Status" callout near the top (zero consumers, early stage) — still open~~ done — shipped (M7, 2026-08-08)
14. ~~Mention `examples/no-go-mod/` (the pilot port) in the README~~
    done — already mentioned in README Consumers section (`ab42b88`)
15. ~~Add a data-flow diagram (Rule -> finding.Finding -> Report -> ExitCode) —
    still open (ROADMAP material)~~ done — shipped (M8, 2026-08-08)
16. Clarify whether the SDK is internal-only or aims for external adoption —
    still open (ROADMAP Q2/Q3)
17. ~~Mark the migration path as "untested in production" or remove step 5~~
    done — README Status section now says "no production linter has fully
    migrated yet" (`ab42b88`)
18. ~~Audit the full README end-to-end for consistency after concurrent edits —
    still open~~ done — full audit + symbol verification (M9, 2026-08-08)
19. ~~Verify every README API claim against actual exported symbols (automated
    check) — still open~~ done — `go doc` cross-check (M9)
20. ~~Add a "Quick Start" section for readers who want to skim — still open~~ done — shipped (M7)

### Verification & quality (MEDIUM)

21. ~~Run `nix run .#lint` and triage the 14 LSP warnings~~ done — 0 issues; gate green (re-verified 2026-09-09)
22. ~~Fix the `wsl_v5` warnings in `rule.go` (missing whitespace above if blocks)~~ done — lint reports 0 issues
23. ~~Fix the `err113` warnings in `rule.go` (dynamic errors)~~ done — `ErrMissingFields` sentinel shipped (v0.2.0)
24. ~~Fix the `nlreturn` warnings (return with no blank line before)~~ done — lint reports 0 issues
25. ~~Fix the `exhaustruct` warning in `registry.go` (runConfig missing field)~~ done — resolved in the config refactor
26. ~~Fix the `bloop` warnings in `registry_test.go` (b.N → b.Loop())~~ done — M1 (2026-08-08)
27. ~~Run `nix run .#test-race` to confirm race-clean status~~ done — race-clean (re-verified 2026-09-09)
28. ~~Run `nix flake check` to validate the flake~~ done — all checks passed (re-verified 2026-09-09)

### Consumer adoption (HIGH — the reason this SDK exists)

29. Pilot-port one rule from `branching-flow` to prove the 1,871 LOC converter
    deletion claim
30. Pilot-port one rule from `erraudit` to prove the 1,214 LOC converter
    deletion claim
31. Migrate `go-structure-linter` to use this SDK as its registry (it already
    aliases `Issue = finding.Finding`)
32. Write a migration guide based on the actual pilot ports (replace the
    aspirational one)

### API surface (LOW — stable, intentional minimalism)

33. Implement `NewRegistryFromRules([]Rule) *Registry` constructor (the only
    un-done Theme 1 item)
34. Add a `Filter` type for severity/category-based finding filtering (ROADMAP)
35. ~~Evaluate severity-tiered exit codes (ROADMAP)~~ **Decided against — `ExitCodeFromReport` is binary by design; `ExitCodeByConfidence` covers tiering and severity mapping is left to consumers (README design notes)**
36. Add `ExitCodeFromFindings([]Finding) int` convenience (ROADMAP)
37. Evaluate a typed `RuleSet` wrapper (ROADMAP)
38. ~~Fuzz `NewRuleError` with nil cause (ROADMAP Theme 6)~~ done — `TestNewRuleError_NilCause` + `FuzzNewRuleError` (7.7M execs)
39. ~~Test `errors.Is` propagation for `context.Canceled`/`DeadlineExceeded`~~ done — `errors_test.go`
40. Confirm `Registry.Register` panic-on-duplicate is the right library contract

### Infrastructure (LOW)

41. Add flake `checks` derivations for `go test`, `go vet`, `golangci-lint`
42. Add `watch`, `tidy`, `deps-update` nix apps
43. Set up Renovate/Dependabot for nix + go dependencies
44. Add `direnv` setup and/or pre-commit hooks
45. Review `devShells.ci` for completeness
46. Add `meta.position` on apps and `flake-schemas`
47. ~~Audit pkg.go.dev publication readiness (does the badge resolve now that
    v1.4.1 is real?)~~ done — pkg.go.dev indexes v0.3.0; docs, examples, and API verified rendering (2026-09-09)
48. ~~Add a `go.work` workspace for sibling-repo development~~ done — created 2026-08-08 (gitignored)
49. Consider a production `cmd/` binary (beyond examples)
50. ~~Evaluate `.goreleaser.yml` only if a binary ships~~ **Decided — non-goal while library-only (ROADMAP Non-goals)**

---

## g) Questions I CANNOT figure out myself

### Q1 — Is `go-structure-linter` actually using this SDK yet, or still using its own `type Issue = finding.Finding` alias?

The README says `go-structure-linter` is the "pilot target" and "No active
consumers yet." But `examples/no-go-mod/` is described as "a pilot port of
`go-structure-linter`'s `NoGoModRule`." Has `go-structure-linter` itself adopted
the SDK in its own repo, or is the pilot only living in this repo's `examples/`?
This determines whether the README's "zero consumers" claim is still accurate or
already stale.

### Q2 — Is this SDK intentionally internal-only (LarsArtmann ecosystem), or does it aim for external adoption?

The pkg.go.dev badge, `go get` instructions, and Go Report Card badge imply
external consumption. But the "Why?" section, Consumers section, and entire
value proposition are framed around three specific LarsArtmann repos. If it's
internal-only, the README should say so (and badges can stay for future use). If
it aims external, the framing needs to generalize and the private-dependency
constraint is a significant adoption barrier.

### Q3 — Should the ROADMAP/DOMAIN_LANGUAGE drift be fixed right now (same session), or is that a separate scoped task?

I identified 8+ stale ROADMAP items and 1 factual error in DOMAIN_LANGUAGE. I
could fix them all right now (they're straightforward text edits), but the user
scoped this session to README quality. The drift is directly related to docs
health but is a different surface area. Should I proceed with a docs-health pass
now, or wait?

---

## Session self-assessment

**What went well:**

- Identified the core problem correctly (converter elimination)
- Prioritized the 4 highest-leverage README improvements
- Verified all changes with build + test + example run
- Honest about the "zero consumers" caveat

**What went wrong:**

- Didn't run `nix run .#lint` (canonical check per AGENTS.md)
- Didn't catch doc drift until forced to self-review
- Didn't update CHANGELOG for README changes
- Didn't audit the full README after concurrent edits merged

**Score: 6/10.** The README work was solid but I was blind to the surrounding
doc rot until prompted. A truly excellent engineer would have noticed the ROADMAP
drift when they first read it for context — not after a self-review prompt.

---

## Resolution (2026-08-08)

Full docs-health pass executed. All 12 doc-drift findings in section c) and
section f) items 1-12 resolved:

- **ROADMAP.md** — completely rewritten: removed all stale items, fixed the
  false "pseudo-version + replace directive" Publication theme, resolved Q1
  (go-finding v1.4.1 confirmed real), updated non-goals.
- **DOMAIN_LANGUAGE.md** — RuleMeta validation corrected: now accurately
  describes all four identity fields (ID, Name, Description, Category).
- **FEATURES.md** — README row updated with all new sections; CI row corrected
  (VCS auth, not sibling-clone).
- **CHANGELOG.md** — `[Unreleased]` restructured: consolidated duplicate
  Added/Changed sections into clean, deduplicated sets.
- **TODO_LIST.md** — rebuilt with 5 actionable items harvested from this report
  and the companion TODO-execution report.

Section f) items 13-20 (README polish): item 14 done (already in README),
item 17 done (README Status section); items 13, 15, 18-20 remain open.
Items 21-28 (verification): item 11 still open (lint not run in that session;
current diagnostics show only 2 `bloop` hints). Items 29-50 (consumer adoption,
API surface, infrastructure): all routed to ROADMAP.md as raw ideas.

> **Addendum (2026-09-09):** items 13, 15, 18-20 were closed by the M7-M9
> README work (2026-08-08); items 21-28 are green (0 lint issues, race-clean,
> flake check passed — re-verified 2026-09-09); item 11's lint gap is closed.
> Still open from this report: item 16 (internal-vs-external framing — ROADMAP
> Q2/Q3) and the consumer-adoption items (TODO_LIST #7-8).
Section g) Q1 resolved; Q2/Q3 remain open in ROADMAP.
