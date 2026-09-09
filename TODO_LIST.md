# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status      | Meaning                                                     |
| ----------- | ----------------------------------------------------------- |
| TODO        | Not started. Needs doing.                                   |
| IN_PROGRESS | Actively being worked on.                                   |
| BLOCKED     | Cannot proceed, external dependency or decision needed.     |
| DONE        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open work

### Publication (repo went public 2026-09-08)

| # | Task                                                                                          | Impact | Effort | Status | Evidence                                                                                                                            |
| - | --------------------------------------------------------------------------------------------- | ------ | ------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| 3 | Review `docs/status/`, `docs/planning/`, `docs/feedback/` for sensitive ops detail            | Medium | 30m    | TODO   | `docs/status/2026-09-09_02-10` f.8 — token-provisioning narratives are world-visible since 2026-09-08. Prune or consciously accept. |
| 5 | gitleaks deep scan over full history (broad + entropy patterns)                               | Medium | 15m    | TODO   | `docs/status/2026-09-09_02-10` f.7 — the pre-flip scan used only two hand-rolled regexes.                                           |
| 6 | GitHub metadata: description, topics (go, linter, static-analysis), homepage → pkg.go.dev URL | Low    | 5m     | TODO   | `docs/status/2026-09-09_02-10` f.11/f.19.                                                                                           |

### Consumer adoption (the reason this SDK exists)

| # | Task                                          | Impact   | Effort | Status | Evidence                                                                                                                                                        |
| - | --------------------------------------------- | -------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 7 | Pilot-port a branching-flow rule to examples/ | Critical | 90m    | TODO   | Proves converter-deletion at scale. Sibling repo at `/home/lars/projects/branching-flow` (1,871 LOC of converter code). Pareto plan M19; #1 value-proving task. |
| 8 | Pilot-port an erraudit rule to examples/      | High     | 90m    | TODO   | Same proof for the error-handling domain. Sibling repo at `/home/lars/projects/erraudit` (1,214 LOC). Pareto plan M20.                                          |

### Dependency hygiene

None open. (`go-finding` pinned at `v1.9.2` — bumped 2026-09-09, gate green.)

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done.
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
  - Vague / long-term items belong in ROADMAP.md, not here.
-->
