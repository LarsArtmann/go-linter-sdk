# Status Report — 2026-09-17 17:28 CEST

**Session scope:** Companion-library documentation session (go-output / cmdguard layering question, README + AGENTS.md updates).
**Constraint honored:** No new research; everything below comes from this session's own tool calls and files read (README.md, AGENTS.md, ROADMAP.md, go.mod, go-finding API surface, sibling repo READMEs).

---

## Session narrative

1. **Question answered:** Why doesn't `go-linter-sdk` depend on `~/projects/go-output` or `~/projects/cmdguard`? Answer delivered with evidence: layering (SDK = rules/report/exit-codes library; cmdguard = CLI shell; go-output = renderer), single-direct-dep policy (go.mod: only `go-finding` + indirect `go-error-family`), and go-finding already owns JSON/SARIF serialization (`report.ToSARIF()`, `go-finding/doc.go:57`).
2. **README edit:** New "Companion libraries" section (README.md:51-61) recommending the stack to end users: go-finding (serialization) → go-output (rendering) → cmdguard/v4 (CLI shell). Table aligned to repo's hand-formatted style.
3. **AGENTS.md edit:** Layering decision recorded in Gotchas ("decided 2026-09-17") so future sessions don't re-litigate the missing dependencies.
4. **Commit state (observed at 17:28):** Daemon committed the first-turn edits in `aed16a8` (AGENTS.md +9, README.md +14). The table-alignment fix (5+/5− in README.md) is **uncommitted**. Commit `1a7d7c8` also contains a **flake.nix +1 change that is NOT mine** — not investigated, not touched (respect-existing-changes rule).

---

## Self-critique (asked explicitly: what did I forget / do better / still improve?)

### Forgotten

1. **pkg.go.dev README freeze.** AGENTS.md documents that tagged READMEs are frozen — the new section is invisible on pkg.go.dev until the next tag. I knew this gotcha and failed to flag the consequence (docs-only PATCH release is the known remedy) when editing the README.
2. **CHANGELOG.md.** AGENTS.md calls CHANGELOG "the narrative of record"; I added user-facing docs without checking for an Unreleased section.
3. **Cross-link from docs/migration-guide.md.** Migrating linter authors are the exact audience for the companion-stack recommendation; README-only placement means pilot-porters may never see it.
4. **URL verification.** I sourced the companion links from the siblings' own READMEs (good) but never confirmed the repo paths resolve, nor included `go get` lines (cmdguard's module path is `/v4` — a nuance worth one line).
5. **One-home check.** The docs-health one-home rule says decide the single right file; I placed the recommendation in README by intuition without checking FEATURES/TODO_LIST routing.

### Could have done better

6. First `mcp_qmd_get` call returned a garbage object dump; I recovered instantly with `view`, but should have gone there first for a local file.
7. First README table had misaligned padding; fixed in a second edit. Widths should have been computed once.
8. No markdown sanity check after editing (no formatter is wired — dprint standalone, not in treefmt; hand-format is the documented convention, so a lint pass would have caught the misalignment).
9. The layering decision was only recorded in AGENTS.md after the user asked for the README section; per the aggressive-update protocol it should have been recorded at answer time.

### Still to improve (repo-level, observed this session)

10. ROADMAP Q3 ("Library-only, or eventual CLI?") is still open; my README section implicitly takes the library-only side. The decision should be made explicit and dated.
11. README's minimal example (and likely examples/minimal-linter) ignores the error from `registry.Run` (`report, _ :=`) — bad example hygiene; consumers copy examples verbatim.
12. No runnable example shows report *output* (PrettyJSON/SARIF/go-output table) — the companion story would land harder with one.

---

## a) FULLY DONE

| # | Item | Evidence |
| - | ---- | -------- |
| 1 | Architecture answer: why no go-output/cmdguard dependency | Layering + go.mod single-dep policy, file:line evidence |
| 2 | README "Companion libraries" section | README.md:51-61, committed in `aed16a8` |
| 3 | AGENTS.md layering-decision record | Gotchas section, committed in `aed16a8` |
| 4 | go-finding serialization surface verified before claiming | `report.ToSARIF()` at go-finding/doc.go:57; json.go API |
| 5 | Table alignment fix | Pending diff 5+/5− (uncommitted) |
| 6 | Skill loading discipline (how-to-golang, status-report) | Both loaded before task execution |

## b) PARTIALLY DONE

| # | Item | Gap |
| - | ---- | --- |
| 1 | pkg.go.dev visibility of new section | Content done; frozen until next tag |
| 2 | Companion-stack discoverability | README only; migration guide + examples not wired |
| 3 | Session documentation | This report completes it |
| 4 | CHANGELOG narrative | Daemon carries the commit; no CHANGELOG line yet |
| 5 | Table-alignment fix committed | Done locally, awaiting daemon/manual commit |

## c) NOT STARTED (touched zero times this session)

- All adoption work: pilot ports beyond the three examples; go-structure-linter full migration; first external consumer.
- Any SDK feature work (no code changed — go.mod untouched, 0 Go files modified).
- Docs-only v0.3.2 tag for pkg.go.dev refresh.
- Example showing output rendering.
- ROADMAP Q3 decision.
- The auto-git daemon's end-of-session **push** (policy (b), decided 2026-09-09) — commits `1a7d7c8`, `aed16a8` plus the pending diff are unpushed as of 17:28.

## d) TOTALLY FUCKED UP

**Nothing.** No data loss, no broken build, no reverted work, no dependency damage — zero Go files touched, working tree healthy. Two blemishes (misaligned table, one wasted tool call) sit firmly at warning level, not fuckup level. Flagged for honesty, not alarm: the `flake.nix +1` inside daemon commit `1a7d7c8` is not mine and remains uninvestigated — per the never-revert-unauthored-changes rule it stays untouched until the user claims or explains it.

## e) WHAT WE SHOULD IMPROVE

1. **README-edit reflex:** every README change should trigger the pkg.go.dev-freeze question ("does this need a docs-only patch tag?").
2. **One-home check before adding doc sections** — decide the file deliberately, don't default to README.
3. **Verify external links before committing them** to public docs.
4. **Record decisions at answer time**, not when asked to persist them.
5. **Decide a markdown-formatting policy once:** either wire a checker into BuildFlow/treefmt or keep hand-format and self-check tables after edits.
6. **Example hygiene rule:** examples must not swallow errors; they are copy-paste templates.
7. **Push at session end** — the daemon policy makes this mandatory and it was about to be missed.

## f) Up to 50 things to get done next (brainstorm — ROADMAP fuel, not commitments)

**Immediate session follow-ups**
1. Push `master` at session end (daemon commits, never pushes — policy (b)).
2. Commit/land the pending README table-alignment diff (daemon will pick it up).
3. Get user clarification on the non-mine `flake.nix +1` in `1a7d7c8`.
4. Decide ROADMAP Q3 explicitly (library-only forever vs eventual `cmd/` CLI) and date it.
5. If library-only: rewrite the ROADMAP `cmd/` raw idea as "consumers use cmdguard" guidance.
6. Docs-only v0.3.2 tag to refresh pkg.go.dev's frozen README (green CI on the exact commit first — v0.3.0 lesson).
7. After tagging: `go get` the tag to trigger pkg.go.dev indexing; verify frozen README; check Imported-by count (post-release habit).
8. Add companion-stack note to docs/migration-guide.md.
9. Add a CHANGELOG line for the companion-libs section at the next release cut.
10. Verify the two companion URLs resolve; add `go get github.com/larsartmann/cmdguard/v4` nuance (module path is `/v4`).
11. Check FEATURES.md/TODO_LIST routing: does the recommendation need a one-home elsewhere, or is README the single home? Record the answer.

**Example & docs quality**
12. Fix `report, _ :=` swallowed error in README example and examples/minimal-linter.
13. Add an output example: `report.PrettyJSON()` / `report.ToSARIF()` and/or a go-output table render.
14. Extend the README data-flow diagram with an optional "render" downstream node (or explicitly decide not to).
15. Link the README "Design notes" bullet ("Depends only on go-finding") to the Companion libraries section.
16. Verify the 8 canonical categories (added in `0bb5e15`) are listed with meanings in README or linked from it.
17. Verify `ExitCodeByConfidence` doc coverage beyond the API table (example exists in example_test.go:135).
18. Consider a pipeline-integration example (DetectorsFromRegistry → go-finding/pipeline) if none exists — verify first.
19. Annotate this status report when its items complete (docs-health ANNOTATE).
20. HARVEST this section into TODO_LIST/ROADMAP (docs-health loop-closing rule).
21. Verify the ROADMAP "pre-flip private dependency warning" concern (ROADMAP.md:56) is actually resolved.

**Adoption (the strategic gap — README Status admits no production consumer yet)**
22. Pilot-port a second branching-flow rule (validate converter-deletion claim at scale — ROADMAP raw idea).
23. Pilot-port a second erraudit rule (ROADMAP raw idea).
24. Drive go-structure-linter's full migration (cleanest pattern; aliases `Issue = finding.Finding` already).
25. Delete measured converter LOC (branching-flow 1,871 / erraudit 1,214) and publish the numbers.
26. Land the first external consumer; track pkg.go.dev "Imported by".
27. Decide whether the SDK should grow a finding.Report → cmdguard ExitCoder bridge helper, or leave it to consumers (YAGNI guard: wait for a consumer to ask).
28. Keep the companion table version-agnostic (cmdguard currently pins go-output v0.38.0 — irrelevant here by design; don't let versions leak into our README).

**SDK / repo hygiene**
29. Decide the markdown-formatting policy once (wire checker into BuildFlow vs documented hand-format).
30. Keep `GOWORK=off go mod tidy` discipline — a `go.work` sits at repo root right now (R3 tidy-mystery trap).
31. Confirm CI green on current master before any tag.
32. Keep the coverage gate excluding examples (documented decision — don't "fix" back to `./...`).
33. Verify workflow GOTOOLCHAIN pins still match go.mod 1.26.7 if CI files are ever touched (setup-go lag gotcha).
34. Preserve reports/.gitkeep + .gitignore exception discipline.
35. Keep branch protection OFF while the daemon pushes to master (documented posture; revisit only if daemon policy changes).

**Distribution / ecosystem**
36. Decide if go-linter-sdk needs a website (website-launch pattern exists for siblings) or stays pkg.go.dev-only.
37. Consider a one-page "LarsArtmann linting stack" diagram (SDK → go-finding → cmdguard → go-output) in docs/.
38. After next tag: re-check pkg.go.dev renders and README freeze (repeat of #7 — it's a habit, listed once here).
39. Keep doc one-home rule enforcement: ROADMAP = not-yet-built, FEATURES = has code (split-brain killer from sessions 2-5).
40. When go-finding next bumps (v1.10.0 → v1.11+): repeat the documented gate (build/test/race/lint/vet/flake-check) and tidy with GOWORK=off.

(Stopped at 40: the remaining 10 would be padding — items above already outrank anything invented to fill the quota.)

## g) Questions I cannot figure out myself

1. **ROADMAP Q3, decided?** Should the SDK stay library-only permanently (my README section assumes yes), or is the `cmd/` CLI raw idea still alive? This determines whether item 5 is "rewrite" or "delete the section I just wrote".
2. **Docs-only v0.3.2 now or batch it?** Refresh pkg.go.dev's frozen README immediately as a PATCH, or wait for the next functional change?
3. **Whose is the `flake.nix +1` in `1a7d7c8`?** It is not mine (my session touched only README.md + AGENTS.md). Yours, another session's, or unexpected? I will not touch it either way — but it should not stay mysterious in the log.

---

*Point-in-time snapshot — will go stale. Annotate, don't rewrite.*
