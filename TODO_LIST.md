# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open work

The backlog is clear. All previously-tracked items were completed on
2026-07-27 (see `CHANGELOG.md` for details):

- ~~Complete the `errors.AsType` migration~~ — done
- ~~Add a concurrent registry test~~ — done (`TestRegistry_ConcurrentReadWrite`)
- ~~Add `Benchmark*` for registry hot paths~~ — done (`Register`/`All`/`Run`)
- ~~Restore the positive double-wrap assertion~~ — done
- ~~Resolve the `testpackage` lint decision~~ — done (moved to black-box
  `package linter_test`)
- ~~Verify a fresh-clone BuildFlow run~~ — done (35/36 on a clean clone)
- ~~Add a GitHub Actions CI workflow~~ — done (`.github/workflows/ci.yml`)
- ~~Right-size `.golangci.yml` from first principles~~ — done (dropped
  cargo-culted `mnd`/`gosec`, trimmed `varnamelen`)
- ~~Verify formatter agreement (`nix fmt` vs `golangci-lint --fix`)~~ — done
  (zero diff)
- ~~Cross-check sibling repos for the missing-`self` flake bug~~ — done (all
  clean)
- ~~Replace wholesale `ireturn` test exclusion~~ — done (factories return
  concrete `RuleFunc`)

Longer-horizon work lives in `ROADMAP.md`. When a ROADMAP idea is refined into
a bounded task, it moves here.

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done (dropped here: LICENSE, .golangci.yml,
    CONTRIBUTING.md, reports/.gitkeep, the `vet` app, .editorconfig, the flake
    `self` arg, and several tests).
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
-->
