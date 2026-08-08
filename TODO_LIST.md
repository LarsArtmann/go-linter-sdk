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
| 1   | Create `PRIVATE_REPO_TOKEN` GitHub secret for CI | Critical | 5m    | `.github/workflows/ci.yml` now references `secrets.PRIVATE_REPO_TOKEN` (PAT with Contents:Read on `go-finding`). The default `GITHUB_TOKEN` is scoped to THIS repo and cannot fetch the private `go-finding` dependency — CI has been failing since creation. Create under Settings → Secrets → Actions → New repository secret. Value: a PAT or fine-grained token with read access to `github.com/larsartmann/go-finding`. |

### Consumer adoption

| # | Task                                              | Impact | Effort | Evidence                                                                                                                                  |
| --- | ------------------------------------------------- | ------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| 2   | Pilot-port a branching-flow rule to examples/     | Critical | 90m   | Proves converter-deletion at scale. Sibling repo at `/home/lars/projects/branching-flow` (1,871 LOC of converter code to eliminate). Identified as the #1 value-proving task in the Pareto plan. |
| 3   | Pilot-port an erraudit rule to examples/          | High   | 90m   | Same proof for the error-handling domain. Sibling repo at `/home/lars/projects/erraudit` (1,214 LOC). |

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
  documented TODOs are already done.
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
  - Vague / long-term items belong in ROADMAP.md, not here.
-->
