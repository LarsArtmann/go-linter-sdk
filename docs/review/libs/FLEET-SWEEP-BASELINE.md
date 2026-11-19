# Fleet sweep baseline — 3 missed channels × 6 consumers (2026-11-19)

> Plan task M14 (F14.1–F14.7), executed immediately after the v0.5.0 release
> (which ships the channels). Purpose: record the BEFORE state so the AFTER
> (consumers adopting CoverageReporter / Relativize / WithEdits) is
> measurable on the next sweep. Sweep method: `rg -c '<channel>' --glob
> '*.go' .` per consumer module, run at each module root.

## Baseline: all channels zero

| Consumer (module root) | SDK pin before wave | DetectorV2 / AddDetectResult / CoverageReporter | Relativize | WithEdits / TextEdit |
| --- | --- | --- | --- | --- |
| go-structure-linter | v0.4.1 | 0 refs | 0 refs | 0 refs |
| samber-linter | v0.4.1 | 0 refs | 0 refs | 0 refs |
| go-humanize-linter | v0.4.1 | 0 refs | 0 refs | 0 refs |
| cmdguard/lint | v0.4.1 | 0 refs | 0 refs | 0 refs |
| project-meta/linter | v0.4.0 | 0 refs | 0 refs | 0 refs |
| BuildFlow/execution (+tools, root indirect) | v0.4.0 / v0.4.1 | 0 refs | 0 refs | 0 refs |

## Bump wave executed 2026-11-19 (F13.5)

All 8 module locations bumped to `go-linter-sdk v0.5.0` with `go get` +
`go build ./...` green; `go test ./...` gates recorded in the session log.
Note: project-meta/linter and BuildFlow's three modules were still on
v0.4.0/v0.4.1 BEFORE this wave — the 2026-10-15 "all six at v0.4.0" snapshot
had silently gone stale (found and fixed here; ADOPTION refresh F13.6
records the corrected table).

## What changes with v0.5.0 without consumer code

- **Relativize:** consumer reports via `Registry.Run` and both adapters now
  emit root-relative finding IDs automatically — the host-path leak
  (`/home/...` in IDs) is gone fleet-wide with zero consumer edits.
- **filesScanned:** still 0 until a consumer rule implements
  `CoverageReporter` — the channel is shipped; the evidence needs one small
  interface per walking rule.
- **Edits:** unchanged (rules opt in via `WithEdits`, hosts apply; ADR 0002).

## AFTER-target for the next sweep

1. Each walking-rule linter (go-structure-linter, samber-linter,
   go-humanize-linter, cmdguard/lint, project-meta/linter) implements
   `CoverageReporter` on its tree-walking rules → `filesScanned ≠ 0` in
   fleet dashboards.
2. BuildFlow/execution consumes `DetectorV2` from SDK adapters in its DAG
   conversion (it already runs `RunDetect`-shaped hosts).
3. Any rule with a mechanical fix attaches `WithEdits` (ADR 0002 GO items).

## How to verify

```bash
for mod in ~/projects/go-structure-linter ~/projects/samber-linter \
           ~/projects/go-humanize-linter ~/projects/cmdguard/lint \
           ~/projects/project-meta/linter ~/projects/buildflow/execution; do
  echo "== $mod"; (cd $mod && rg -c 'DetectorV2|AddDetectResult|CoverageReporter|Relativize|WithEdits' --glob '*.go' . || echo 0 refs)
done
grep -n 'go-linter-sdk' ~/projects/{go-structure-linter,samber-linter,go-humanize-linter}/go.mod \
  ~/projects/cmdguard/lint/go.mod ~/projects/project-meta/linter/go.mod \
  ~/projects/buildflow/{execution,tools}/go.mod
```

— Snapshot 2026-11-19, orchestrating session. Point-in-time; living state in
docs/ADOPTION.md.
