# Feedback: Why cqrs-lint Cannot Adopt go-linter-sdk Today

> **From:** cqrs-lint maintainer (go-cqrs-lite monorepo)
> **Date:** 2026-07-30
> **Context:** Evaluated go-linter-sdk for cqrs-lint adoption during an architecture improvement session. cqrs-lint has 100 rules across 8 categories, a pre-built cross-file analysis context, feature-profile detection, and uses `go-finding/pipeline` for execution.

---

## Executive Summary

go-linter-sdk is the right abstraction for **stateless linters** — linters where each rule independently scans a directory and emits findings. branching-flow, erraudit, and go-structure-linter fit this model: each rule is self-contained, reads files itself, and doesn't share state with other rules.

cqrs-lint does **not** fit this model. It is a **stateful, context-driven linter** where 100 rules share a single pre-built `AnalysisContext` containing parsed ASTs, a cross-referenced CQRS registry, a feature profile, and a line cache. The SDK's `Rule.Check(ctx, dir)` signature has no way to pass this context. Adopting the SDK today would require either a hacky context-key workaround or a fundamentally different rule architecture.

This document explains the impedance mismatches in detail and proposes concrete SDK changes that would make adoption feasible.

---

## 1. The Core Mismatch: Stateless Rules vs. Shared Analysis Context

### What the SDK assumes

```go
// go-linter-sdk/rule.go
type Rule interface {
    Name() string
    Description() string
    Category() Category
    Severity() finding.Severity
    IsEnabledByDefault() bool
    Check(ctx context.Context, dir string) ([]finding.Finding, error)
}
```

`Check` receives a `context.Context` and a `dir string`. The rule is expected to scan the directory itself — read files, parse ASTs, match patterns, and return findings. Each rule is self-contained. `Registry.Run` calls `Check` on each rule sequentially, aggregating results.

### What cqrs-lint needs

```go
// cqrs-lint/pkg/analyzer/types.go
type AnalysisContext struct {
    Fset           *token.FileSet           // shared file set for position mapping
    Packages       []*packages.Package      // pre-loaded Go packages with ASTs
    Registry       *CQRSRegistry            // cross-referenced commands, events, folds, deciders, projections
    ProjectRoot    string
    ModulePath     string
    GoFiles        []*GoFile                // flat list of all parsed Go files
    FeatureProfile FeatureProfile            // auto-detected: store kind, command flow, server mode, tracing, snapshot
    RulesConfig    RulesConfig               // user config overrides
    LoadErrors     []PackageLoadError
    lineCache      sync.Map                  // file → []string for SourceLine()
}
```

This context is built **once** by `BuildContext()` (which calls `packages.Load`, scans all files, populates the registry, runs feature detection) and then **shared** with all 100 detectors. Every detector reads from `ctx.GoFiles`, `ctx.Registry`, `ctx.FeatureProfile`, and `ctx.SourceLine()`.

### The concrete incompatibility

cqrs-lint's detectors look like this:

```go
func NewC001Detector(ctx *analyzer.AnalysisContext) finding.Detector {
    return finding.NamedDetectorFunc("C001-missing-tx-commit",
        func(_ context.Context) ([]finding.Finding, error) {
            for _, gf := range ctx.GoFiles {       // ← needs AnalysisContext
                if gf.IsTest { continue }
                for _, decl := range gf.AST.Decls { // ← needs pre-parsed AST
                    // ... pattern matching ...
                    b := finding.NewBuilder("C001", toolName, msg,
                        finding.SeverityCritical,
                        finding.Pos(...)).
                        WithSnippet(ctx.SourceLine(...)) // ← needs line cache
                    // ...
                }
            }
            return findings, nil
        },
    )
}
```

The `*analyzer.AnalysisContext` is captured in the closure. The detector doesn't receive `dir` — it doesn't need to, because the directory was already walked, packages were loaded, and ASTs were parsed before any detector runs.

The SDK's `Rule.Check(ctx, dir)` signature has no slot for this context. The options are:

1. **Stuff `*AnalysisContext` into `context.Context` via a custom key** — works, but is the "hacky" pattern. Every rule needs to extract it: `actx, ok := ctx.Value(cqrsCtxKey).(*AnalysisContext)`. This is what `finding.WorkingDirFromContext` already does for `dir`, but for a rich struct with 10 fields, it's a code smell. And it makes rules untestable without the context key set.

2. **Have each rule independently rebuild the analysis** — each of 100 rules would call `packages.Load(dir)` independently. That's 100× the package loading, 100× the AST parsing, 100× the file walking. On a medium project, `BuildContext()` takes ~500ms. 100× that is 50 seconds. Non-starter.

3. **Don't adopt the SDK** — which is the current state.

---

## 2. The Execution Model Mismatch: Registry.Run vs. Pipeline

### SDK execution

```go
// go-linter-sdk/registry.go
func (r *Registry) Run(ctx context.Context, dir string) (*finding.Report, error) {
    for _, rule := range r.All() {
        findings, err := rule.Check(ctx, dir)
        // ...
        report.AddFindings(findings)
    }
    return report, nil
}
```

Sequential, single-pass, no retry, no fix application, no iterations, no suppression, no stale detection. This is fine for simple linters.

### cqrs-lint execution

```go
// cqrs-lint/run.go
pipeConfig := pipeline.Config{
    MaxIterations:       5,        // fix → re-detect → fix again (up to 5x)
    ParallelDetectors:   true,     // 100 detectors run concurrently
    GracefulDegradation: true,     // one detector failure doesn't kill the run
    DryRun:              !cfg.Fix,
    Timeout:             5 * time.Minute,
    Processors: []pipeline.FindingTransformer{
        suppression.NewSuppressionFilter(),  // //cqrs-lint:ignore(RULE) comments
    },
    FixProviders: []pipeline.FixProvider{fix.NewCQRSFixProvider()},
}
pipe, _ := pipeline.New(pipeConfig, cfg.Path, detectors...)
result, _ := pipe.Run(ctx)
```

cqrs-lint uses `go-finding/pipeline` — a different layer of the go-finding ecosystem. The pipeline provides:
- **Multi-pass fix iteration** (detect → fix → re-detect, up to 5 times)
- **Parallel detector execution** (100 detectors concurrently, thread-safe)
- **Graceful degradation** (one detector crash doesn't kill the run)
- **Suppression processing** (inline `//cqrs-lint:ignore(RULE)` comments)
- **Fix providers** (domain-specific byte-level edits)
- **Metrics collection** (per-detector timing for `--verbose`)
- **Timeouts** (per-detector and global)

The SDK's `Registry.Run` is a flat sequential loop. It doesn't integrate with the pipeline. `DetectorFromRegistry` adapts the registry to a single `finding.Detector`, but then the pipeline sees ONE detector (the registry), not 100 individual ones — so per-detector timing, per-detector timeouts, and per-detector error isolation are lost.

**This is actually the deeper mismatch:** cqrs-lint doesn't use `Registry.Run` at all — it uses the pipeline directly, with each detector registered individually. The SDK's registry is redundant architecture.

---

## 3. The Identity Mismatch: Rule Metadata vs. Catalog Metadata

### SDK approach

Each `Rule` carries its own metadata via interface methods:
```go
Name() string
Description() string
Category() Category
Severity() finding.Severity
IsEnabledByDefault() bool
```

### cqrs-lint approach

cqrs-lint separates metadata from implementation. The catalog (`AllRules() []RuleInfo`) is a declarative table:

```go
type RuleInfo struct {
    ID          string
    Name        string
    Category    string
    Severity    string
    Confidence  string  // ← not in the SDK
    Description string
    AutoFix     bool    // ← not in the SDK
}
```

And the detector is a separate constructor:
```go
func NewC001Detector(ctx *analyzer.AnalysisContext) finding.Detector
```

**Key differences the SDK doesn't support:**

1. **Confidence field** — cqrs-lint has `Confidence` (low/medium/high) on every finding, separate from `Severity`. The SDK's `Rule` interface has no `Confidence()` method. This is a deliberate design choice in cqrs-lint: a rule can be "critical severity, low confidence" (likely a real bug but the heuristic might fire on false positives).

2. **Rule ID** — cqrs-lint uses `"C001"` as the rule ID and `"missing-tx-commit"` as the rule name. The SDK has only `Name()`. The dual-identity (ID for filtering/health-score, name for display) is a cqrs-lint convention that doesn't map to the SDK.

3. **AutoFix flag** — cqrs-lint marks rules as auto-fixable in the catalog. The SDK has no concept of fixability at the Rule level.

4. **Category taxonomy** — the SDK has `CategoryDesign`, `CategoryStructure`, `CategoryErrorHandling`, `CategoryCorrectness`, `CategoryStyle`, `CategoryPerformance`, `CategorySecurity`, `CategoryConfiguration`. cqrs-lint has `correctness`, `api`, `boilerplate`, `consistency`, `architecture`, `security`, `performance`, `version`. Only 3 overlap. Adopting the SDK's categories would lose the CQRS-specific taxonomy that users see in `cqrs-lint rules` output.

5. **Catalog-detector bidirectional test** — cqrs-lint has a meta-test verifying the catalog and detector registry agree on every rule ID. The SDK's approach (metadata lives in the Rule itself) makes this drift impossible by construction, but also makes the catalog less useful as a separate declarative artifact (cqrs-lint uses it for the `rules` subcommand, SARIF metadata, and the health-score computation).

---

## 4. The Feature Profile Mismatch

cqrs-lint auto-detects the consumer project's architecture via `FeatureProfile`:

```go
type FeatureProfile struct {
    Store       StoreKind       // sqlite, postgres, pebble, memory, turso, custom, none
    CommandFlow CommandFlowKind // read-only, sync, commands
    HasServer   bool
    HasSoftDelete bool
    Tracing     TracingKind     // off, on
    Snapshot    SnapshotKind    // off, on
}
```

Rules consult this profile to suppress themselves:
- S002 (encryption) only fires when `HasServer == true`
- A016 (idempotency) only fires when `CommandFlow == commands`
- B014 (OTel) only fires when `Tracing != off`

This is **cross-rule state**: the feature profile is computed once after scanning all files, then shared with all rules. The SDK has no concept of this — each rule would need to independently re-detect whether the project is a server, whether it uses commands, etc. That's either 100× redundant detection work or a context-key hack.

---

## 5. What Would Need to Change in go-linter-sdk

Here are concrete changes, ordered by impact. I would adopt the SDK if **all** of the top 3 are implemented.

### Change 1: Support context-aware rules (critical)

Add an optional interface that rules can implement to receive a pre-built analysis context:

```go
// ContextAwareRule is an optional interface a Rule can implement to receive
// a pre-built analysis context instead of just a directory path. Registry.Run
// and DetectorFromRegistry type-assert for this and pass the context through.
type ContextAwareRule interface {
    Rule
    CheckContext(ctx context.Context, analysis any) ([]finding.Finding, error)
}
```

The `analysis any` parameter is deliberately untyped — the SDK doesn't know about `*analyzer.AnalysisContext`. Each linter defines its own context type. `Registry.Run` passes it through:

```go
func (r *Registry) Run(ctx context.Context, analysis any) (*finding.Report, error) {
    for _, rule := range r.All() {
        if car, ok := rule.(ContextAwareRule); ok {
            findings, err := car.CheckContext(ctx, analysis)
            // ...
        } else {
            findings, err := rule.Check(ctx, analysis.(string)) // backward compat
            // ...
        }
    }
}
```

This is backward compatible — existing rules that only implement `Check(ctx, dir)` continue to work. New rules that need shared state implement `CheckContext`.

### Change 2: Integrate with go-finding/pipeline (critical)

The SDK's `Registry.Run` is a flat sequential loop. It should instead produce a `[]finding.Detector` that the pipeline can consume with per-detector parallelism, timeouts, and error isolation.

Current `DetectorFromRegistry` collapses all rules into ONE detector — losing per-rule granularity. Instead:

```go
// DetectorsFromRegistry returns one finding.Detector per rule, so the pipeline
// can run them in parallel with per-detector timeouts and error isolation.
func DetectorsFromRegistry(registry *Registry, toolName string, analysis any) []finding.Detector {
    var detectors []finding.Detector
    for _, rule := range registry.All() {
        detectors = append(detectors, ruleToDetector(rule, toolName, analysis))
    }
    return detectors
}
```

This lets cqrs-lint drop its hand-rolled `register.go` (130 lines of manual `NewC001Detector(ctx)`, `NewC002Detector(ctx)`, ...) and use the SDK's registry + `DetectorsFromRegistry` instead.

### Change 3: Add Confidence and AutoFix to RuleMeta (critical)

```go
type RuleMeta struct {
    Name        string
    Description string
    Cat         Category
    Sev         finding.Severity
    Conf        finding.Confidence  // ← new
    AutoFix     bool                // ← new
}
```

And add `Confidence()` and `AutoFixable()` methods to the `Rule` interface:

```go
type Rule interface {
    // ... existing methods ...
    Confidence() finding.Confidence   // ← new
    AutoFixable() bool                // ← new
}
```

This lets cqrs-lint's health-score computation and `--fp-suspects` mode work with the SDK's metadata instead of needing a separate catalog.

### Change 4: Support custom category taxonomies (important)

The SDK's 8 `Category` constants are too restrictive. Allow linters to define their own categories:

```go
// RegisterCategory adds a custom category to the SDK's known set.
// Rules can use custom categories without conflicting with the built-ins.
func RegisterCategory(name string) Category
```

Or simpler: change `Category` from a closed set of constants to an open `string` type, and document the recommended values. cqrs-lint's `api`, `boilerplate`, `consistency`, `architecture`, and `version` categories would then work without mapping to the SDK's taxonomy.

### Change 5: Support dual identity (ID + Name) (nice to have)

cqrs-lint uses `"C001"` as a stable ID (for filtering, health-score, suppression) and `"missing-tx-commit"` as a display name. The SDK has only `Name()`. Add:

```go
type Rule interface {
    // ... existing methods ...
    ID() string  // ← new: stable short identifier (e.g., "C001")
}
```

With a default implementation in `RuleFunc` that derives the ID from `Name` (or uses `Name` directly for linters that don't need dual identity).

### Change 6: Support suppression comments (nice to have)

cqrs-lint has `//cqrs-lint:ignore(RULE)` inline suppression. The SDK has no concept of suppression. A `SuppressionParser` interface or a `Processors` hook in `Registry.Run` would let linters plug in suppression filtering without reimplementing it.

---

## 6. What I Would NOT Change in the SDK

- **`Rule` emits `finding.Finding` directly** — this is the SDK's best design decision. No converter layer. cqrs-lint already follows this pattern. Keep it.
- **`RuleFunc` + `RuleMeta` struct-literal pattern** — rules-as-data is excellent. cqrs-lint would adopt this for new rules.
- **`OptIn` for experimental rules** — good idea, cqrs-lint could use this for its `--fp-suspects` rules.
- **`RuleError` wrapping** — cqrs-lint would benefit from this. Currently detector errors are silently dropped.
- **`ExitCodeFromReport`** — simple, useful, correct.
- **Dependency only on `go-finding`** — keep the SDK minimal. Don't add `golang.org/x/tools` as a dependency. The analysis-context pattern should be the linter's responsibility, not the SDK's.

---

## 7. Summary: Adoption Decision Matrix

| SDK Feature | cqrs-lint Need | Gap |
|---|---|---|
| `Rule` interface | 100 detectors | No `Confidence()`, no `AutoFixable()`, no `ID()` |
| `RuleFunc` + `RuleMeta` | Rule declarations | Missing `Conf` and `AutoFix` fields |
| `Registry.Run` | Execution | No shared-context support; sequential, no pipeline |
| `DetectorFromRegistry` | Pipeline integration | Collapses 100 rules to 1 detector (loses per-rule granularity) |
| `ExitCodeFromReport` | Exit code | Works, but cqrs-lint has more nuanced exit logic |
| `RuleError` wrapping | Error attribution | Good — cqrs-lint should adopt this |
| `Category` constants | Rule taxonomy | Only 3 of 8 overlap; no custom category support |
| `OptIn` | Experimental rules | Good — cqrs-lint could use this |
| `Check(ctx, dir)` | Rule execution | No way to pass `*AnalysisContext` |

### Verdict

**Cannot adopt today.** The SDK is designed for stateless linters where each rule independently scans a directory. cqrs-lint is a stateful linter where 100 rules share a pre-built analysis context. The three critical changes (context-aware rules, pipeline integration, confidence/autofix metadata) would close the gap. Without them, adoption would require hacks (context keys) or performance regressions (100× redundant package loading).

With the three critical changes, cqrs-lint could:
- Drop `pkg/rules/register.go` (130 lines of manual registration)
- Drop `pkg/rules/catalog.go` + `catalog_extra.go` (400 lines of duplicate metadata)
- Drop the custom `toolName` constant (already partially done via `lintutil`)
- Adopt `RuleError` wrapping for better error attribution
- Adopt `OptIn` for experimental rules

That's ~530 LOC eliminated and better error handling, with no behavioral regression. Worth the SDK changes.
