# go-finding (root module) utilization review — go-linter-sdk

- **Reviewed:** 2026-11-19 by fleet audit Group A (sub-agent; orchestrated session)
- **Library:** github.com/larsartmann/go-finding (root module)
- **Sub-modules considered:** root, toolsdk, pipeline, analysis, cmd/go-finding (go-finding repo); consumer is single-module. Only the root module is in the consumer's dependency graph (`go list -m all` shows exactly `github.com/larsartmann/go-finding v1.15.1`); toolsdk gets its own report at [go-finding-toolsdk.md](go-finding-toolsdk.md)
- **Version reviewed:** consumer pin v1.15.1 (go.mod:5); library repo latest local tag `v1.15.1` (`git describe` on root files: `cmd/go-finding/v1.15.1-87-g498498f`), HEAD `498498f` 2026-11-18; point-in-time snapshot
- **Verdict:** ADOPTED — the SDK is built spine-to-spine on go-finding (Rule emits `finding.Finding` directly, no converter layer), but the three v1.14–v1.15 capabilities aimed squarely at SDK-style detectors (coverage evidence, host-portable IDs, machine-readable fixes) are all unadopted
- **Score:** 65/100 (moderate) — 7 of 13 applicable capability rows fully leveraged

## What the library offers

Read from the module-cache copy of the pinned version,
`$(go env GOMODCACHE)/github.com/larsartmann/go-finding@v1.15.1/` (identical to the
`v1.15.1` tag in /home/lars/projects/go-finding). 153 exported top-level declarations
(`rg '^(func|type|var|const) [A-Z]' *.go | grep -v _test | wc -l`).

| Capability | Library evidence (file:line @ v1.15.1) |
| --- | --- |
| `Finding` data model + branded identity types (`ID`, `RuleName`, `ToolName`, `FilePath`, `GroupID`) | `finding.go:14` (struct), `branded_types.go:8-16`, `id.go:34` (`GenerateID`) |
| Fluent Builder: `NewBuilder` + 15 `With*` + `Build`/`MustBuild`/`BuildOrDefault` | `finding_builder.go:16` (NewBuilder), `:29-140` (WithID…WithMetadata), `:143-171` |
| `Template` batch builder factory (stamp tool/category/fix/tags/group once) | `finding_builder.go:178` (type), `:189` (`NewTemplate`), `:244` (`Template.Builder`) |
| `Report` aggregation: `NewReport`, `AddFinding(s)`, `ComputeSummary(At)`, `NewReportFromFindings` | `report.go:97-135`, `:213-226`; `Summary` incl. `FilesScanned`/`SkippedModules` at `report.go:71,74-78` |
| Report queries: `Len`, `All`, `ActiveFindings`, `BySeverity`, `GroupFindings(Sorted)`, `Filter`, `Map`, `FindingsSnapshot` | `report_query.go:162,203,32,50,82,114,170,183,13` |
| Coverage-evidence channel: `DetectResult`, `DetectorV2`, `RunDetect`, `(*Report).AddDetectResult` | `detect_result.go:15,39,53,72` (in the pinned v1.15.1, shipped v1.15.0) |
| Working-dir context contract: `WithWorkingDir` / `WorkingDirFromContext` | `context.go:21,31` |
| Detector adapters: `Detector`, `DetectorFunc`, `NamedDetectorFunc`; `CheckBinary`/`RunCmd` | `detector.go:15,25,35,56,70` |
| Host-portable IDs: `Relativize` (added **in the pinned v1.15.1**, issue #47) | `relativize.go:30` |
| Category taxonomy: 16 standard constants, `IsValid`, `IsStandard`, `ParseCategory`, `Compare`, `CategoryForLinter` | `category.go:9-26,44-49,80`, `category_linter.go:195` |
| Fix representation: `WithEdits`/`TextEdit`/`WithBeforeCode`/`WithAfterCode`, `ApplySimpleFixes`, `HasFix`/`IsAutoFixable` | `finding_builder.go:80,64,71`; `text_edit.go:28`; `simple_fix.go:27`; `finding_methods.go:91,111` |
| Finding groups: `WithGroupID`, `GroupFindings`, `GroupFindingsSorted` | `finding_builder.go:102`; `report_query.go:82,114` |
| Confidence model: `ConfidenceLow/Medium/High/Full`, `Compare`, `Clamp`, `ParseConfidence` | `confidence.go:19-26,50,19` |
| Error taxonomy: `FindingError` + category constructors, `IsFindingError`, `CategoryOf` | `errors.go:44` and passim |
| Interchange: `ToSARIF(WithOpts)`, `FindingsFromSARIF`, `FromJSON`/`ReportFromJSON`, `PrettyJSON` | `sarif_export.go:64,70`; `sarif_import.go`; `json.go:110,139,157` |
| Renderers: `FormatText`, `FormatTextRich`, `FormatMarkdown`, `FormatTable` | `format.go:13,39` |
| Cross-run tooling: `Diff`, `Correlate`, `Combine`/`MergeIter`, LSP export | `diff.go:35`; `correlate.go:58`; `merge.go:23`; `lsp.go` |

## What go-linter-sdk uses today

29 distinct `finding.*` symbols across 212 references (`rg -o 'finding\.[A-Z][A-Za-z]*' --type go | sort -u | wc -l` = 29; per-file counts: registry_test.go 83, rule.go 30, registry.go 22, example_test.go 33, examples/* 42, bench_test.go 2).

| Used API | Consumer evidence (file:line) |
| --- | --- |
| `[]finding.Finding` as the Rule output type | rule.go:102 (`Check(ctx, dir) ([]finding.Finding, error)`) — the SDK's founding decision, documented rule.go:75-78 |
| `finding.NewBuilder` (25 call sites) + `WithConfidence/WithFixStrategy/WithSuggestion/WithMetadata/WithCategory` + `MustBuild` | rule.go:266-276 (`RuleFunc.NewFinding`); examples/minimal-linter/main.go:42-51; examples/naked-return-guard/main.go:163-178; examples/no-go-mod/main.go:56-65; examples/oops-redundant-guard/main.go:351-365; example_test.go:27-30; registry_test.go:997-1043 |
| `finding.NewReport` + `ToolInfo` + `AddFindings` + `ComputeSummary` | registry.go:204-225 (`Registry.Run` aggregation) |
| `Report.Len`, `Report.All`, `Confidence.Compare` | registry.go:307 (`ExitCodeFromReport`), registry.go:363-364 (`ExitCodeByConfidence`) |
| `Report.ActiveFindings` | examples/minimal-linter/main.go:63, examples/naked-return-guard/main.go:73, examples/no-go-mod/main.go:77 (example output loops) |
| `finding.Detector` + `NamedDetectorFunc` | registry.go:242-262 (`DetectorFromRegistry`), registry.go:275-301 (`DetectorsFromRegistry`) |
| `WithWorkingDir` / `WorkingDirFromContext` | registry.go:244, registry.go:285 (detectors read the host's dir); registry_test.go:220-226, 716-742 (contract pinned by test); example_test.go:99-105 |
| `Severity`/`Confidence`/`FixStrategy` constants | rule.go:100,124; all four examples; registry_test.go passim |
| `finding.Pos` / `FilePath` / `RuleName` / `ToolName` | rule.go:266-271; every example's builder call |
| `finding.Category` (string conversion) | rule.go:275 (`builder.WithCategory(finding.Category(r.Meta.Cat))`) |

## Capability inventory & gap analysis

| # | Capability | Status | Evidence (consumer) | Evidence (library) | Impact |
| --- | --- | --- | --- | --- | --- |
| 1 | Finding data model + branded identity | **Fully Leveraged** | rule.go:102; IDs auto-generated via Builder throughout | finding.go:14; id.go:34 | The SDK's whole premise (zero converter layer) rests on this; branded types used end-to-end |
| 2 | Builder API | **Fully Leveraged** | 25 call sites incl. rule.go:266; chained confidence/fix/suggestion/metadata in all examples | finding_builder.go:16-171 | Validated findings at construction; no hand-built Finding literals anywhere |
| 3 | Builder `Template` | Partially Used | rule.go:260-279 — `RuleFunc.NewFinding` reimplements Template stamping (tool+category over NewBuilder) in 15 LOC instead of delegating to `NewTemplate(...).WithCategory(...).Builder(...)` | finding_builder.go:189, 244 | Low: 15 LOC duplication; Template also can't carry a default severity from meta, so the SDK's version is a justified superset — but it should delegate for the shared part |
| 4 | Report construction & summary | **Fully Leveraged** | registry.go:204-225 (NewReport→AddFindings→ComputeSummary); WithToolName flows into ToolInfo (registry.go:29-31, 199-207) | report.go:97-135, 213 | Correct incremental aggregation; `NewReportFromFindings` not needed (Run accumulates per-rule) |
| 5 | Report query surface | **Fully Leveraged** | `Len` registry.go:307,359; `All` registry.go:363; `ActiveFindings` in all examples' output loops | report_query.go:32,162,203 | SDK uses exactly the queries its role needs; BySeverity/Filter/Map are host-facing. Footnote: `ExitCodeFromReport` uses `Len()` (suppressed included) while examples print `ActiveFindings()` — latent only, since `Registry.Run` never emits suppressed findings |
| 6 | Coverage-evidence channel (`DetectorV2`/`DetectResult`/`AddDetectResult`) | **Missed Opportunity** | zero references (`rg 'DetectorV2|DetectResult|FilesScanned|SkippedModules' --type go` → 0 hits); CHANGELOG.md:14-19 records the v1.15.0 bump noticing the channel and consciously not moving | detect_result.go:15,39,53,72; Summary fields report.go:71,74-78; guide docs/guides/detector-coverage.md (unreleased HEAD) | Every SDK-built report renders `filesScanned: 0` — fleet CI cannot tell "clean tree" from "scanned nothing", and BuildFlow dashboards under-count coverage for all linters on this SDK |
| 7 | Working-dir context contract | **Fully Leveraged** | registry.go:244,285; pinned by tests registry_test.go:202-260, 716-742 | context.go:21,31 | Module fan-out / per-directory host runs work transparently (README.md:135) |
| 8 | Detector adapters | **Fully Leveraged** | registry.go:243, 282 — both adapters return `finding.Detector` via `NamedDetectorFunc`, named per tool/rule | detector.go:15-51 | Exact type the BuildFlow/toolsdk lane consumes (see toolsdk report) |
| 9 | Host-portable IDs (`Relativize`) | **Missed Opportunity** | zero references; example rules pass raw `filepath.WalkDir` paths into `Pos` (examples/naked-return-guard/main.go:87-113, 167) — absolute under BuildFlow, so finding IDs embed `/home/...` | relativize.go:30 — added in the pinned v1.15.1 explicitly so "a detector invoked with a relative target (CLI lane) and … an absolute working directory (SDK/orchestrator lane) produce the same ID" | Baselines/dedup built on the CLI lane don't match the BuildFlow lane for the same source fact; IDs leak host paths into shared reports |
| 10 | Category taxonomy & validation | Partially Used | rule.go:36-57 defines the SDK's own 11 constants (10 string-identical to the library's 16); converts at rule.go:275; `validateIdentityFields` (rule.go:166-190) checks only non-empty, never `finding.Category(...).IsValid()` | category.go:9-26, 44-49 (`IsValid` rejects `"Bogus Cat"` typos), 80 | Medium-low: deliberate decoupling (documented rule.go:34-44) but a near-copy of the constant set that can drift, and rule categories skip shape validation the library already ships |
| 11 | Fix representation (`Edits`/`TextEdit`/`ApplySimpleFixes`) | **Missed Opportunity** | every example finding stops at `FixStrategySuggest`+`WithSuggestion` (e.g. examples/oops-redundant-guard/main.go:356-358); the oops port explicitly deleted erraudit's 102-LOC autofixer (examples/oops-redundant-guard/main.go:5-7); `Rule` has no fix surface, `FixStrategyDirect` appears once in a doc comment (rule.go:88) | finding_builder.go:80 (`WithEdits`); text_edit.go:28; simple_fix.go:27 (`ApplySimpleFixes`); finding_methods.go:91,111 | Fleet linters on this SDK cannot express machine-applicable fixes, so go-finding's Direct-fix pipeline is unreachable through the SDK; the flagship pilot port *lost* an autofix capability it originally had |
| 12 | Finding groups (`WithGroupID`/`GroupFindings`) | Partially Used | reachable by chaining on the builder that `NewFinding` returns (rule.go:260), but zero plumbing, docs, or examples; `GroupFindings` unused | finding_builder.go:102; report_query.go:82,114 | Low-medium: group-style linters (clone detection) get no template-level support; hosts can still consume GroupID if rules set it |
| 13 | Confidence model | **Fully Leveraged** | consts + raw 0.85 literals in examples (naked-return-guard/main.go:169); `Compare` powers tiered exits registry.go:364 | confidence.go:19-26, 50 | Both named constants and the open [0,1] continuum used as intended |
| 14 | Error taxonomy (`FindingError` etc.) | Not Applicable | SDK's `RuleError` (errors.go:30) is rule-attribution, a concept go-finding does not model; rule errors come from user closures, not finding-domain failures | errors.go:44 | — |
| 15 | Serialization & interchange (SARIF/JSON) | Not Applicable | — the SDK stops at findings/report/exit codes by explicit layering (AGENTS.md:188-196); `--format` work is ROADMAP'd as CLI-layer (ROADMAP.md:49) | sarif_export.go:64; json.go:110-177 | Output belongs to the cmdguard/go-output shell in the intended stack |
| 16 | Renderers (`FormatText`…) | Not Applicable | examples' one-line `fmt.Printf` loops (examples/minimal-linter/main.go:63-65) are deliberate demo minimalism; the library layer must not render (same layering rule) | format.go:13-39 | — |
| 17 | Cross-run tooling (`Diff`/`Correlate`/`Combine`/LSP) | Not Applicable | the SDK produces one fresh single-tool report per run; baselining, dedup, and IDE export are host/CLI concerns | diff.go:35; correlate.go:58; merge.go:23; lsp.go | — |

Row scores: Fully=1 (rows 1,2,4,5,7,8,13 = 7), Partial=0.5 (rows 3,10,12 = 3), Missed=0 (rows 6,9,11 = 3), N/A excluded (rows 14-17 = 4).
**Score = round(100 × 8.5 / 13) = 65/100 (moderate); 7 of 13 applicable rows fully leveraged.**

## Version currency

- Consumer pins **v1.15.1** — the **latest published root tag** (`git -C /home/lars/projects/go-finding tag -l 'v*' --sort=-creatordate | head -2` → `v1.15.1`, `v1.15.0`). Fully current.
- v1.15.1 itself = v1.15.0 + `Relativize` (+65 LOC, relativize.go) and a GenerateID doc note (issue #47) — i.e. the pin already carries the host-portable-ID capability the SDK hasn't adopted.
- Library HEAD is **110 commits past v1.15.1** (2026-11-18), but the unreleased root-module Go API delta is tiny: branded-type tightening (`IsHashID(id ID)` signature change, SARIF `RuleID` now branded — id.go, sarif_export.go), a `VersionPatch` stamp fix (version.go), plus new docs (`docs/guides/detector-coverage.md`, `docs/ecosystem.md` with the 23-repo consumer-pin sweep), CI scripts, and a new `examples/detectorv2` demo. Nothing in the unreleased root API changes any verdict above; the two type tightenings are compile-compatible for this consumer (it never calls `IsHashID` or reads SARIF RuleID).
- Bump history is disciplined: CHANGELOG.md:12-19 records the v1.15.0→v1.15.1 era bump with gates green, and `scripts/check-doc-versions.sh` keeps docs from duplicating the pin.

## The fleet argument

Owner's concern: "I have MANY projects to work on, and limited time and resources; If all my project re-implement/reintegrate everything I will only fix bugs all day long and none of my projects can benefit from collective improvements."

This repo is the fleet's best-case answer on the axis it was built for: it deleted the Violation→Finding converter class (1,871 LOC in branching-flow, 1,214 in erraudit) by making `finding.Finding` the native rule output, and every fix go-finding lands in the Builder, Report, Detector, or context contract flows into all six direct consumers without per-repo work. The three misses all run the other way — they are places where the *library* solved a fleet problem (coverage evidence: "0 findings is ambiguous"; host-portable IDs: CLI lane vs orchestrator lane; machine-readable fixes) and the SDK has not yet turned the crank to inherit it. Each is a small, one-file change here that would otherwise be re-solved (or left unsolved) in every linter repo — exactly the re-implementation tax the owner is trying to kill.

## Ranked opportunities (impact × ease)

1. **Adopt the coverage channel** — add an optional coverage hook (e.g. `Rule` implementations may also satisfy `interface{ Coverage(ctx, dir) (filesScanned int, skipped []string) }` or a `DetectResult`-returning variant) and have `Registry.Run` feed `report.AddDetectResult` instead of `AddFindings`. Impact: every SDK linter's report becomes checkable (`filesScanned ≠ 0`), unblocking fleet dashboards. Ease: moderate (interface addition + Run change; see go-finding's `examples/detectorv2` for the pattern). — routing: TODO_LIST row "Adopt go-finding DetectorV2 coverage evidence in Registry.Run (fleet audit docs/review/libs/go-finding.md row 6)" / ROADMAP line "Coverage-aware Run (DetectorV2)"
2. **Relativize on the BuildFlow/detector path** — in `DetectorFromRegistry`/`DetectorsFromRegistry`, run findings through `finding.Relativize(findings, dir)` before returning (dir is already the ctx working dir). Impact: stable IDs across CLI and orchestrator lanes, no host paths in reports. Ease: trivial (one call, both adapters). — routing: TODO_LIST row "Relativize finding paths in Detector adapters (go-finding v1.15.1 Relativize)" / none
3. **Give rules a fix surface** — optional `Fixer` interface on rules that returns `[]finding.TextEdit` (library type, text_edit.go:28), stamped via `NewFinding(...).WithEdits(...)` with `FixStrategyDirect`; leave application to hosts/`finding.ApplySimpleFixes`. Impact: restores the auto-fix capability the oops port deleted; opens the go-finding Direct pipeline to the whole fleet of SDK linters. Ease: moderate-large (design decision; pairs with ROADMAP's repair questions). — routing: ROADMAP line "Rule-level fix representation (TextEdit/Direct) bridging to go-finding's fix pipeline"
4. **Validate categories with the library's own shape check** — `validateIdentityFields` (rule.go:166) calls `finding.Category(cat).IsValid()` and warns on non-standard-but-valid values; keep the SDK type, drop nothing. Ease: trivial. — routing: TODO_LIST row "Shape-validate RuleMeta.Cat via finding.Category.IsValid()"
5. **Delegate `NewFinding`'s shared stamping to `finding.NewTemplate`** — rule.go:266-276 becomes a Template with tool+category, keeping the meta-severity argument. Ease: trivial, cosmetic. — routing: none (informational)

## Why (not) adopted / flip trigger

Adopted and load-bearing: the Finding/Builder/Report/Detector spine **is** the SDK — 29 symbols, 212 references, and the six-consumer converter-deletion record (README.md:360-377) prove it. The gaps are new-ish library surface (v1.14–v1.15.1) that arrived after the SDK's shape froze; CHANGELOG.md:14-19 shows the team saw the coverage channel at bump time and deferred it. Flip trigger for each gap is listed under Ranked opportunities; none requires breaking the single-dep policy (all capabilities are in the root module the SDK already pins).

## How to verify

```bash
cd /home/lars/projects/go-linter-sdk
GOWORK=off go list -m all | grep finding
# github.com/larsartmann/go-finding v1.15.1   (only the root module; no toolsdk/pipeline/analysis)

rg -o 'finding\.[A-Z][A-Za-z]*' --type go -N | sed 's/^[^:]*://' | sort -u | wc -l
# 29 distinct symbols; rg -c 'finding\.' --type go →
# registry_test.go:83 example_test.go:33 rule.go:30 registry.go:22 bench_test.go:2
# examples/naked-return-guard:13 examples/oops-redundant-guard:10 examples/no-go-mod:10 examples/minimal-linter:9

rg -n 'GroupFindings|FilesScanned|SkippedModules|ToSARIF|FindingsFromSARIF|Relativize|DetectorV2|DetectResult|AddDetectResult|ApplySimpleFixes|WithEdits|WithGroupID|NewTemplate|NotSuppressed|finding\.Combine|Correlate|FormatText' --type go
# (no matches, exit 1) — rows 6, 9, 11, 12 and the N/A rows are genuinely untouched

git -C /home/lars/projects/go-finding tag --sort=-creatordate | head -6
# cmd/go-finding/v1.15.1, analysis/v1.15.1, pipeline/v1.15.1, v1.15.1, cmd/go-finding/v1.15.0, pipeline/v1.15.0
git -C /home/lars/projects/go-finding log --oneline v1.15.1..HEAD | wc -l   # 110 unreleased commits
ls "$(go env GOMODCACHE)/github.com/larsartmann/go-finding@v1.15.1/detect_result.go" \
   "$(go env GOMODCACHE)/github.com/larsartmann/go-finding@v1.15.1/relativize.go"
# both exist — coverage channel and Relativize ship in the pinned version

# spot-check citations:
sed -n '260,279p' /home/lars/projects/go-linter-sdk/rule.go          # NewFinding
sed -n '204,225p' /home/lars/projects/go-linter-sdk/registry.go     # NewReport/AddFindings/ComputeSummary
sed -n '5,7p' /home/lars/projects/go-linter-sdk/examples/oops-redundant-guard/main.go  # deleted 102-LOC autofixer
sed -n '15,50p' "$(go env GOMODCACHE)/github.com/larsartmann/go-finding@v1.15.1/detect_result.go"
sed -n '1,30p' "$(go env GOMODCACHE)/github.com/larsartmann/go-finding@v1.15.1/relativize.go"
```

## Addendum 2026-11-19 (post-audit, same session): rows 6, 9, 10 remediated

TODO #24/#25/#26 were executed the same day as this audit. The SDK now wires:

- **Row 6 (coverage channel): ADOPTED.** `Registry.Run` feeds
  `report.AddDetectResult` (registry.go, `Run`); rules may implement the new
  optional `linter.CoverageReporter` interface to supply `FilesScanned` /
  `SkippedModules`. Both adapters (`DetectorFromRegistry`,
  `DetectorsFromRegistry`) return detectors that also satisfy
  `finding.DetectorV2`, so `finding.RunDetect` hosts receive coverage evidence.
- **Row 9 (Relativize): ADOPTED — semantics verified from source first**
  (copy-in/copy-out, passthrough for relative/out-of-root paths). All three
  execution paths (`Run`, both adapters) run findings through
  `finding.Relativize` against the target dir, so finding IDs no longer embed
  host paths and match across the CLI and orchestrator lanes.
- **Row 10 (category shape validation): ADOPTED.** `validateIdentityFields`
  now rejects malformed categories via `finding.Category.IsValid()` (new
  sentinel `linter.ErrInvalidCategory`; custom lowercase-hyphenated categories
  still welcome).

Score impact: rows 6 and 9 move Missed → Fully, row 10 Partial → Fully.
Recomputed: 10 of 13 applicable rows fully leveraged → **85/100**. The
remaining gap is row 11 (fix representation — ROADMAP theme 1, plan M15).
Behavior change to note for consumers: finding IDs for absolute-dir runs
change (relativization); that is the intended MINOR-release fix (v0.5.0).

> **Re-review trigger:** re-run this audit when the consumer pins a newer
> release of the reviewed library than the version stamped above, or when a
> follow-up task from the 2026-11-19 Pareto plan lands in this repo
> (footers standardized 2026-11-19, plan task M22/F22.1).
