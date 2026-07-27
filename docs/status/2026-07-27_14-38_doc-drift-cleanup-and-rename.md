# Status Report — go-linter-sdk (Session 6)

**Date:** 2026-07-27 14:38 CEST
**Scope:** Close the doc-drift items left open by session 5; rename
`hierarchical-errors` → `erraudit` across living docs
**Verdict:** Every doc-drift item from session 5's self-review is resolved.
BuildFlow 35/36 (unchanged — the 3 root-package findings are ROADMAP Q2).

---

## TL;DR

| Dimension                                 | State                                                                                       |
| ----------------------------------------- | ------------------------------------------------------------------------------------------- |
| ROADMAP Q5 (testpackage) doc drift        | **FIXED** — Q5 closed as resolved (black-box); Q4 closed as moot                            |
| FEATURES↔ROADMAP PLANNED split-brain      | **FIXED** — "Planned capabilities" section removed from FEATURES; one-home rule established |
| README broken `os.Exit` example           | **FIXED** — `"os"` import added; example verified to compile                                |
| `docs/DOMAIN_LANGUAGE.md`                 | **WRITTEN** — 6 core concepts with relationships diagram                                    |
| `hierarchical-errors` → `erraudit` rename | **DONE** — all living docs + source updated; status reports left as historical              |
| CHANGELOG citations                       | **VERIFIED** — `0ca7a41`, `88bf523`, `c13366c` all exist and match descriptions             |
| `//nolint` audit                          | **PASS** — only 1 directive in code (`registry_test.go:216`), with documented reason        |
| BuildFlow                                 | **35/36** — same baseline; only ROADMAP Q2 root-package findings remain                     |
| Tests                                     | **10/10** pass, race-clean                                                                  |
| Lint                                      | **0 issues**                                                                                |

---

## What was done

### 1. ROADMAP open questions closed (Q4 + Q5)

Session 5 resolved the white-box-vs-black-box test decision in code (chose
black-box) but left ROADMAP Q5 open. This session:

- Moved Q5 to a new "Resolved questions" section with the decision and commit
  citation (`b369bba`).
- Moved Q4 (flake.lock bump) to resolved — it's moot; the `self` fix is
  permanent and verified across all sibling repos.
- Left Q1, Q2, Q3 genuinely open (they need user decisions or external
  unblocks).

### 2. FEATURES↔ROADMAP split-brain eliminated

Removed the entire "Planned capabilities (no code yet)" section from
`FEATURES.md`. All 8 items were already in `ROADMAP.md` themes. Established a
**one-home rule**: FEATURES tracks only what exists in code today; ROADMAP
owns all not-yet-built capabilities until they graduate to TODO_LIST.

Documented the rule in FEATURES.md header + AGENTS.md gotchas.

### 3. README example fixed

Added `"os"` to the import block (was missing — `os.Exit` call wouldn't
compile). Verified the full example compiles against the actual exported API
by building it as a standalone file.

### 4. `docs/DOMAIN_LANGUAGE.md` written

Deferred since session 1. Defines the 6 core concepts (`Rule`, `RuleFunc`,
`RuleMeta`, `Category`, `Registry`, `RuleError`) with:

- Precise definitions grounded in actual code locations
- A relationships diagram (ASCII)
- An explicit "What is NOT in this domain" section (no Violation type, no CLI,
  no config parsing)

Updated FEATURES.md (PLANNED → FULLY_FUNCTIONAL) and ROADMAP Theme 5 (struck
through as done).

### 5. `hierarchical-errors` → `erraudit` rename

User informed me the project was renamed. Updated all references in:

- `rule.go` (package doc comments)
- `README.md` (5 references: table, prose, migration path, consumers)
- `AGENTS.md` (consumer list)
- `ROADMAP.md` (converter-deletion claim)
- `CHANGELOG.md` (findings attribution)
- `docs/DOMAIN_LANGUAGE.md` (converter-deletion claim)

Left untouched:

- `AGENTS.md:74` — references the Crush **skill** `hierarchical-errors` (still
  named that in the skill registry; distinct from the repo)
- `docs/status/*.md` — 4 historical status reports (point-in-time records)

### 6. CHANGELOG + AGENTS.md updated

Added CHANGELOG entries for: DOMAIN_LANGUAGE.md, one-home rule, Q4/Q5
resolution, README fix, rename. Added AGENTS.md gotchas for the one-home rule
and auto-git daemon commit convention.

---

## What remains

Only ROADMAP-level open questions need user input:

- **Q1** — Is `go-finding` going to get a tagged release? (blocks external
  consumers + pkg.go.dev)
- **Q2** — Should package files live at repo root or under `internal/`? (the
  only BuildFlow issue source — 3 findings)
- **Q3** — Library-only, or eventual CLI? (determines whether `self`-based
  flake versioning is needed)

Everything else is either done or tracked as a ROADMAP raw idea.
