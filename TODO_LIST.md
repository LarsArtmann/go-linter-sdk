# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.
>
> Harvested from the status reports in `docs/status/` (2026-07-19 sessions 1 & 2,
> 2026-07-27 session 3), deduplicated, and verified against the code on
> 2026-07-27. Already-done items were dropped (they live in `CHANGELOG.md`).

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## High Impact

| Task                                                                 | Status    | Impact | Effort | Evidence                                                                                  |
| -------------------------------------------------------------------- | --------- | ------ | ------ | ----------------------------------------------------------------------------------------- |
| Complete the `errors.AsType` migration                               | 🔴 `TODO` | High   | 15m    | `registry.go` uses `errors.AsType[*RuleError]`; `registry_test.go:196` + `errors.go:18` doc still use `errors.As` (gopls `errorsastype` hint) |
| Add a concurrent registry test (exercising the `RWMutex` under race) | 🔴 `TODO` | High   | 30m    | `registry.go:14` is thread-safe by construction but `-race` passes vacuously; no concurrent `Register`/`All`/`Run` test exists |
| Add `Benchmark*` for the registry hot paths                          | 🔴 `TODO` | High   | 30m    | No `Benchmark*` in `registry_test.go`; an SDK called on every lint run has no perf baseline |
| Restore the positive double-wrap assertion                           | 🔴 `TODO` | High   | 15m    | `TestRegistry_Run_NoDoubleWrap` only catches double-wrap, not zero-wrap (session 2, §d.4/§e.7) |
| Resolve the `testpackage` lint decision                              | 🔴 `TODO` | High   | 30m    | `.golangci.yml:301` silences `testpackage` via path exclusion; tests are white-box (`package linter`). Decide white-box+document or move to `package linter_test` |
| Verify a fresh-clone BuildFlow run                                   | 🔴 `TODO` | High   | 20m    | `reports/.gitkeep` durability (`.gitignore:64`) is theoretical — never proven by `git clone` + `nix develop --command buildflow` |

## Medium Impact

| Task                                                                   | Status    | Impact | Effort | Evidence                                                                                  |
| ---------------------------------------------------------------------- | --------- | ------ | ------ | ----------------------------------------------------------------------------------------- |
| Add a GitHub Actions CI workflow                                       | 🔴 `TODO` | Med    | 1h     | No `.github/`; all verification (build/test/lint/flake/buildflow) is local-only           |
| Right-size `.golangci.yml` from first principles                       | 🔴 `TODO` | Med    | 1h     | Cargo-culted `mnd` numbers, `varnamelen` names, `gosec` exclusions inherited from `go-finding` (session 2, §b.1/§e.5) |
| Verify formatter agreement (`nix fmt` vs `golangci-lint --fix`)        | 🔴 `TODO` | Med    | 20m    | `.golangci.yml` formatters (`gci`/`goimports`/`gofumpt`/`golines`) overlap treefmt; split-brain risk unverified (session 2, §d.3) |
| Cross-check sibling repos for the missing-`self` flake bug             | 🔴 `TODO` | Med    | 20m    | `go-structure-linter`, `branching-flow`, `hierarchical-errors` may share the latent `outputs @{ ... }` bug fixed here in `c13366c` (session 3, §f.7) |

## Low Impact

| Task                                                                  | Status    | Impact | Effort | Evidence                                                                |
| --------------------------------------------------------------------- | --------- | ------ | ------ | ----------------------------------------------------------------------- |
| Replace wholesale `ireturn` test exclusion with targeted `//nolint`   | 🔴 `TODO` | Low    | 10m    | `.golangci.yml:302` excludes `ireturn` for all `_test.go`; `makeRule`/`failingRule` are the only offenders (session 2, §f.18) |
| Enrich `AGENTS.md` with `.golangci.yml`, `reports/.gitkeep`, and the `self`/Nix-version requirement | 🔴 `TODO` | Low    | 15m    | AGENTS.md omits the lint config, the `.gitkeep` mechanism, and the Nix `@`-pattern strictness gotcha |

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
