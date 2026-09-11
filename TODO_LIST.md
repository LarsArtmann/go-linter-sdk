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

<!-- Numbering gap at #1, #2, #4: those items were completed and removed
     (1: v0.3.1 release, 2: secret deletion, 4: dprint decision). Numbers are
     never reused — historical reports reference them. -->

### Publication (repo went public 2026-09-08)

None open. (2026-09-09: gitleaks full-history scan clean + sensitive-docs
sweep accepted with dated verdict; GitHub metadata set — description, topics
`go`/`linter`/`static-analysis`, homepage → pkg.go.dev.)

### Consumer adoption (the reason this SDK exists)

None open. (Both pilot ports landed 2026-09-09: `examples/naked-return-guard`
from `branching-flow`, `examples/oops-redundant-guard` from `erraudit`. Full
migrations are the consumers' work now — see `docs/migration-guide.md`.)

### Dependency hygiene

None open. (`go-finding` pinned at `v1.10.0` — bumped 2026-09-11, gate green.)

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done.
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
  - Vague / long-term items belong in ROADMAP.md, not here.
-->
