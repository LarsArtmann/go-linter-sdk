# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| TODO         | Not started. Needs doing.                                   |
| IN_PROGRESS  | Actively being worked on.                                   |
| BLOCKED      | Cannot proceed, external dependency or decision needed.     |
| DONE         | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open work

### CI & infrastructure

| # | Task                                              | Impact | Effort | Evidence                                                                                                                                  |
| --- | ------------------------------------------------- | ------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Verify CI can fetch `go-finding` via `GITHUB_TOKEN` | Critical | 30m    | `.github/workflows/ci.yml:52` — uses `secrets.GITHUB_TOKEN` which is scoped to THIS repo. `go-finding` is a DIFFERENT private repo. May silently fail or use a different auth path. Flagged HIGH RISK in report 2026-07-30_16-08 §B.1. Must verify end-to-end or switch to PAT/SSH deploy key. |
| 2   | Add `go.work` for local cross-repo dev            | High   | 15m    | No `go.work` exists. Local development against an uncommitted `go-finding` checkout requires manual `replace` lines. A workspace file (not committed) would formalize this. Flagged in report 2026-07-30_16-08 §B.2. |
| 3   | Add `go mod tidy` check to CI                     | Medium | 10m    | CI doesn't verify that `go.mod` and `go.sum` are tidy. A `git diff --exit-code go.mod go.sum` step would catch drift. Flagged in report 2026-07-30_16-39. |

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done.
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
  - Vague / long-term items belong in ROADMAP.md, not here.
-->
