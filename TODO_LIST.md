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

| # | Task                                                                           | Impact   | Effort | Status  | Evidence                                                                                                                                                                                                                       |
| - | ------------------------------------------------------------------------------ | -------- | ------ | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | Cut the next release tag (v0.4.0) so pkg.go.dev stops rendering false docs     | Critical | 20m    | TODO    | pkg.go.dev indexes v0.3.0 (tagged `06cd065`, 2026-09-08) whose frozen README still says "go-finding is a private repository. Set GOPRIVATE=..." — wrong since the public flip. Master README is corrected (`7675912`); only a new tag refreshes pkg.go.dev. Tag only after CI is green on the exact commit (`AGENTS.md` gotcha). |
| 2 | Delete the now-unused `PRIVATE_REPO_TOKEN` GitHub secret                        | High     | 5m     | BLOCKED | Maintainer go-ahead required (destructive, account-touching — `docs/status/2026-09-09_02-10` §g.1). Repo is public; `ci.yml` has had zero `secrets.` references since `7675912`. Ghost credential.                             |
| 3 | Review `docs/status/`, `docs/planning/`, `docs/feedback/` for sensitive ops detail | Medium | 30m    | TODO    | `docs/status/2026-09-09_02-10` f.8 — token-provisioning narratives are world-visible since 2026-09-08. Prune or consciously accept.                                                                                            |
| 5 | gitleaks deep scan over full history (broad + entropy patterns)                | Medium   | 15m    | TODO    | `docs/status/2026-09-09_02-10` f.7 — the pre-flip scan used only two hand-rolled regexes.                                                                                                                                       |
| 6 | GitHub metadata: description, topics (go, linter, static-analysis), homepage → pkg.go.dev URL | Low | 5m | TODO | `docs/status/2026-09-09_02-10` f.11/f.19.                                                                                                                                                                                      |

### Consumer adoption (the reason this SDK exists)

| # | Task                                        | Impact   | Effort | Status | Evidence                                                                                                                                                        |
| - | ------------------------------------------- | -------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 7 | Pilot-port a branching-flow rule to examples/ | Critical | 90m   | TODO   | Proves converter-deletion at scale. Sibling repo at `/home/lars/projects/branching-flow` (1,871 LOC of converter code). Pareto plan M19; #1 value-proving task. |
| 8 | Pilot-port an erraudit rule to examples/      | High     | 90m   | TODO   | Same proof for the error-handling domain. Sibling repo at `/home/lars/projects/erraudit` (1,214 LOC). Pareto plan M20.                                           |

### Dependency hygiene

| #  | Task                                                            | Impact | Effort | Status | Evidence                                                                                                                                   |
| -- | --------------------------------------------------------------- | ------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| 9  | Evaluate `go-finding` v1.9.2 bump (`v1.7.0` pinned in `go.mod`)  | Low    | 20m    | TODO   | v1.8.0 and v1.9.2 tags exist upstream (module cache, 2026-09-09); changelogs unreviewed, possible breaking changes. Review, then bump or pin consciously. |

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done.
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
  - Vague / long-term items belong in ROADMAP.md, not here.
-->
