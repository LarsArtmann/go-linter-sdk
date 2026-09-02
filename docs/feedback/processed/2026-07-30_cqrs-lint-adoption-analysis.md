# Feedback: Why cqrs-lint Cannot Adopt go-linter-sdk Today

> **From:** cqrs-lint maintainer (go-cqrs-lite monorepo)
> **Date:** 2026-07-30
> **Context:** Evaluated go-linter-sdk for cqrs-lint adoption during an architecture improvement session. cqrs-lint has 100 rules across 8 categories, a pre-built cross-file analysis context, feature-profile detection, and uses `go-finding/pipeline` for execution.

---

> **Processing note (2026-07-30):** This feedback has been reviewed by the
> go-linter-sdk maintainer. Sections 1-7 are the original cqrs-lint perspective.
> Some claims (e.g., "Category constants are too restrictive", "no concept of
> fixability") are factually incorrect — `Category` is already an open type and
> per-finding `Confidence`/`FixStrategy` already exist. See **Appendix A** at
> the end of this document for the full maintainer response and accept/reject
> decisions on all 6 proposed changes.

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

| SDK Feature             | cqrs-lint Need       | Gap                                                            |
| ----------------------- | -------------------- | -------------------------------------------------------------- |
| `Rule` interface        | 100 detectors        | No `Confidence()`, no `AutoFixable()`, no `ID()`               |
| `RuleFunc` + `RuleMeta` | Rule declarations    | Missing `Conf` and `AutoFix` fields                            |
| `Registry.Run`          | Execution            | No shared-context support; sequential, no pipeline             |
| `DetectorFromRegistry`  | Pipeline integration | Collapses 100 rules to 1 detector (loses per-rule granularity) |
| `ExitCodeFromReport`    | Exit code            | Works, but cqrs-lint has more nuanced exit logic               |
| `RuleError` wrapping    | Error attribution    | Good — cqrs-lint should adopt this                             |
| `Category` constants    | Rule taxonomy        | Only 3 of 8 overlap; no custom category support                |
| `OptIn`                 | Experimental rules   | Good — cqrs-lint could use this                                |
| `Check(ctx, dir)`       | Rule execution       | No way to pass `*AnalysisContext`                              |

### Verdict

**Cannot adopt today.** The SDK is designed for stateless linters where each rule independently scans a directory. cqrs-lint is a stateful linter where 100 rules share a pre-built analysis context. The three critical changes (context-aware rules, pipeline integration, confidence/autofix metadata) would close the gap. Without them, adoption would require hacks (context keys) or performance regressions (100× redundant package loading).

With the three critical changes, cqrs-lint could:

- Drop `pkg/rules/register.go` (130 lines of manual registration)
- Drop `pkg/rules/catalog.go` + `catalog_extra.go` (400 lines of duplicate metadata)
- Drop the custom `toolName` constant (already partially done via `lintutil`)
- Adopt `RuleError` wrapping for better error attribution
- Adopt `OptIn` for experimental rules

That's ~530 LOC eliminated and better error handling, with no behavioral regression. Worth the SDK changes.

---

## Appendix A: Maintainer Response & Decisions

> **Date:** 2026-07-30
> **Status:** Decisions recorded after cross-referencing every claim against the actual SDK source (`rule.go`, `registry.go`, `errors.go`), the `go-finding` library (`finding.go`, `confidence.go`, `finding_builder.go`, `context.go`), and `go-finding/pipeline` (`pipeline.go`, `config.go`, `fix_provider.go`, `stage.go`).

### Framing

Three facts reshape the entire response:

1. **No consumers exist yet** (Q8). The SDK has zero downstream users. This means **breaking changes are free** — there has never been a cheaper time to evolve the API, and there will never be again. Every decision below should be evaluated on architectural merit, not backward compatibility.

2. **The SDK's scope is stateless linters** (Q1). cqrs-lint is a stateful, context-driven linter with a fundamentally different architecture (shared `AnalysisContext`, pipeline-driven, 100 rules with cross-rule state). The SDK should **not** bend its core abstractions to accommodate stateful linters. cqrs-lint is a `go-finding/pipeline` consumer, not a `go-linter-sdk` consumer. This is by design, not a gap.

3. **The goal is ecosystem-wide improvement** (Q7), not cqrs-lint appeasement. Changes are evaluated by whether they benefit _all_ future consumers, not by whether they unblock one specific linter.

### Decision Summary

| # | Proposed Change                                 | Verdict                 | Rationale                                                                                          |
| - | ----------------------------------------------- | ----------------------- | -------------------------------------------------------------------------------------------------- |
| 1 | `CheckContext(ctx, analysis any)`               | **REJECT**              | `any` is a type-safety regression; SDK stays stateless; stateful linters use the pipeline directly |
| 2 | `DetectorsFromRegistry` (one detector per rule) | **ACCEPT**              | Benefits all pipeline consumers; low-risk; backward compatible; the real high-value change         |
| 3 | Rule-level `Confidence()` / `AutoFixable()`     | **DOCUMENT, DON'T ADD** | `finding.Finding` already has both at the per-finding level; gap is documentation, not API         |
| 4 | Custom category taxonomy                        | **ALREADY SOLVED**      | `Category` is already `type Category string`; the constants are recommendations, not enforcement   |
| 5 | Dual `ID()` + `Name()`                          | **ACCEPT**              | Stable ID + mutable display name is a clean contract; do it now while breaking changes are free    |
| 6 | Suppression in `Registry.Run`                   | **REJECT**              | Wrong layer; the pipeline already handles suppression via `FindingTransformer` processors          |

---

### Change 1 — Context-aware rules: REJECT

**The proposal:** Add `ContextAwareRule` with `CheckContext(ctx, analysis any)` so stateful linters can pass a pre-built analysis context.

**Why rejected:**

- **`any` is rejected on principle.** The user's position is unambiguous: "I hate `any`'s." A parameter typed `any` discards compile-time safety and forces every consumer to type-assert. It is no better than the `context.Value` key hack the feedback itself calls a "code smell" — it just moves the smell from a context key to a function parameter. Both approaches erase the type system at the exact boundary where types matter most (the analysis context is the richest, most important data a stateful rule touches).

- **The SDK's scope is stateless linters** (Q1). Adding context-aware rules to the SDK's core `Rule` interface or `Registry.Run` path introduces a stateful execution model that conflicts with the SDK's design intent. The `Registry.Run` sequential loop is deliberately simple because stateless rules don't need shared state.

- **The pipeline already solves this.** cqrs-lint already uses `go-finding/pipeline` directly, with each detector capturing its own `*AnalysisContext` in a closure. This works. It is the correct architecture for a stateful linter. The SDK does not need to replicate it.

**What stateful linters should do instead:** Continue using `go-finding/pipeline` directly with custom detector constructors. The SDK is not the right layer for shared analysis contexts. If a future typed analysis-context abstraction emerges, it belongs in `go-finding` (which already owns `Detector`, `WorkingDirFromContext`, and the pipeline), not in this thin SDK layer.

---

### Change 2 — `DetectorsFromRegistry`: ACCEPT

**The proposal:** Return `[]finding.Detector` (one per rule) instead of a single collapsed detector, so the pipeline can run rules in parallel with per-detector timeouts, error isolation, and metrics.

**Why accepted:**

- **Benefits all pipeline consumers**, not just cqrs-lint. Any linter that wants per-rule parallelism, per-rule timing, or per-rule error isolation in the pipeline needs this. `branching-flow`, `erraudit`, and `go-structure-linter` would all benefit when they adopt the pipeline.

- **Low-risk and backward compatible.** The existing `DetectorFromRegistry` (singular) stays for simple use cases. `DetectorsFromRegistry` (plural) is a new function that returns one `finding.Detector` per rule. Both can coexist.

- **`pipeline.Detector` is already a type alias for `finding.Detector`** (verified: `integration_endtoend_test.go` explicitly proves `pipeline.Detector == finding.Detector`). So the return type plugs directly into `pipeline.New(config, rootDir, detectors...)` with zero adaptation.

- **Removes the granularity collapse.** The current `DetectorFromRegistry` wraps all rules into one detector — the pipeline sees a single opaque blob and cannot attribute timing, errors, or timeouts to individual rules. `DetectorsFromRegistry` preserves per-rule identity at the pipeline layer.

**Design note:** Each emitted detector should use the rule's `Name()` (or future `ID()`) as the `finding.NamedDetectorFunc` name, so pipeline metrics and error reports attribute to the correct rule automatically.

**This is the single highest-value change in the document.** It should be implemented first and independently.

---

### Change 3 — Rule-level Confidence/AutoFix: DOCUMENT, DON'T ADD

**The proposal:** Add `Confidence()` and `AutoFixable()` to the `Rule` interface and `Conf`/`AutoFix` to `RuleMeta`.

**Why not added — the API already exists at the right level:**

The feedback claims "the SDK's Rule interface has no `Confidence()` method" and "the SDK has no concept of fixability at the Rule level." Both statements are technically true but **architecturally misleading**: `go-finding` already provides both concepts at the _finding_ level, which is strictly more expressive:

- **`finding.Confidence`** (`go-finding/confidence.go`): a `float64` type with `ConfidenceLow` (0.25), `ConfidenceMedium` (0.5), `ConfidenceHigh` (0.75) constants. The `finding.Builder` has `.WithConfidence(c Confidence)`.

- **`finding.FixStrategy`** (`go-finding/finding.go`): per-finding field with values `FixStrategyNone`, `FixStrategySuggest`, `FixStrategyDirect`, `FixStrategyAI`. The `finding.Builder` has `.WithFixStrategy(fs FixStrategy)`.

A rule-level `Confidence()` default is strictly **less** expressive than per-finding confidence: a heuristic rule that fires with high confidence on pattern A and low confidence on pattern B cannot be represented by a single rule-level value. The same applies to `AutoFix`: a rule may emit some findings that are auto-fixable and others that are only suggestions.

**The real problem is documentation** (Q4). The SDK's package docs (`rule.go`, package doc comment) and `docs/DOMAIN_LANGUAGE.md` do not explain that:

1. Rules should set `Confidence` and `FixStrategy` on individual findings via `finding.NewBuilder(...)`.
2. `RuleMeta.Sev` is the _default_ severity — individual findings can override it.
3. The same override pattern applies to confidence and fix strategy.

**Action:** Improve documentation with a "Building Findings" section showing the `NewBuilder` chain: `.WithConfidence(...)`, `.WithFixStrategy(...)`. This closes the perceived gap without API surface bloat.

---

### Change 4 — Category taxonomy: ALREADY OPEN (documentation gap)

**The proposal:** Allow custom categories or change `Category` to an open `string`.

**The type is already open.** Verified from source (`rule.go`, `type Category`):

```go
type Category string
```

`Category` is a named string type with recommended constants. It is **not** an enum. It is **not** a closed set. `Category("api")`, `Category("boilerplate")`, `Category("version")` all compile and work today. The 8 `CategoryXxx` constants are convenience values, not a constraint.

**The feedback's claim is factually wrong:** "The SDK's 8 Category constants are too restrictive. ... cqrs-lint's `api`, `boilerplate`, `consistency`, `architecture`, `version` categories would then work without mapping." They already work. Try it:

```go
linter.RuleFunc{
    Meta: linter.RuleMeta{
        Name: "my-rule",
        Cat:  linter.Category("api"),  // compiles, runs, reports as "api"
        ...
    },
}
```

**The real problem is documentation** — the SDK presents the constants as if they're the complete vocabulary without stating "define your own if needed."

#### Q5 Deep Dive: Category — PRO / CONTRA / Alternatives

Despite `Category` already being open, the question of whether to keep the recommended constants, remove them, or add a validation mechanism is worth a thorough analysis:

**Option A: Status quo — open string type + recommended constants** (current)

| Aspect          | Assessment                                                                                                                                                                                                                                                 |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Type safety     | `Category("typo")` compiles — no typo detection                                                                                                                                                                                                            |
| Flexibility     | Full — any string works, domain-specific categories supported                                                                                                                                                                                              |
| Discoverability | Good — constants appear in IDE autocomplete as hints                                                                                                                                                                                                       |
| Enforcement     | None — convention only                                                                                                                                                                                                                                     |
| Verdict         | **Keep.** This is the Go-idiomatic pattern (`http.StatusText`, `mime.TypeByExtension`). The constants document intent without restricting expression. Add a doc comment: _"These are recommended values. Define your own for domain-specific categories."_ |

**Option B: Open string type, no constants**

| Aspect          | Assessment                                                                                                     |
| --------------- | -------------------------------------------------------------------------------------------------------------- |
| Type safety     | Same as A (none)                                                                                               |
| Flexibility     | Full                                                                                                           |
| Discoverability | Worse — no autocomplete hints, consumers must read docs                                                        |
| Enforcement     | None                                                                                                           |
| Verdict         | **Reject.** Removes useful guidance without gaining anything. The constants don't constrain; they communicate. |

**Option C: Registry-level category validation**

```go
registry.RegisterCategory("api")  // whitelist
registry.Register(rule)           // panics if rule.Category() not whitelisted
```

| Aspect          | Assessment                                                                                                                       |
| --------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Type safety     | Runtime validation — typos caught at registration                                                                                |
| Flexibility     | Moderate — requires upfront declaration                                                                                          |
| Discoverability | Moderate — registry knows its categories                                                                                         |
| Enforcement     | Opt-in — only registries that call `RegisterCategory` validate                                                                   |
| Complexity      | Adds API surface (`RegisterCategory`, `KnownCategories`, validation logic)                                                       |
| Verdict         | **Reject for now.** Premature for a ~300 LOC SDK with zero consumers. This is a ROADMAP idea if a real consumer needs it. YAGNI. |

**Option D: Typed category set per-linter (generics)**

```go
type RuleWithCategory[C ~string] interface {
    Rule
    Category() C
}
```

| Aspect          | Assessment                                                                                                                                                                         |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Type safety     | Compile-time — each linter defines its own category enum                                                                                                                           |
| Flexibility     | Moderate — categories are per-linter, not cross-linter                                                                                                                             |
| Discoverability | Excellent — linter's categories are a closed enum                                                                                                                                  |
| Enforcement     | Full — invalid categories don't compile                                                                                                                                            |
| Complexity      | High — generics on the Rule interface, breaks the simple struct-literal pattern                                                                                                    |
| Verdict         | **Reject.** Over-engineered. The whole value of `RuleFunc{Meta: RuleMeta{Cat: ...}}` is that it's a plain struct literal. Generics make rule declaration ugly and linter-specific. |

**Recommendation: Option A (status quo) + documentation.** The type is already correct. Add a doc comment on the `const` block clarifying these are recommendations, not a closed set. This is a 2-line documentation fix, not an API change.

---

### Change 5 — Dual ID() + Name(): ACCEPT

**The proposal:** Add `ID()` to the `Rule` interface for a stable short identifier, keeping `Name()` as a display name.

**Stability contract** (per Q6):

- **`ID()` is the canonical stable key.** Once published, it never changes. Used for: registry deduplication (panic on duplicate `ID()`), `RuleError` attribution, suppression matching (`//tool:ignore(ID)`), filtering (`--enable ID`), and the `finding.RuleName` field on emitted findings.
- **`Name()` is the display name.** Human-readable, can change across versions without breaking suppression, filtering, or health-score computation.

**Why accepted:**

- **Do it now — breaking changes are free** (Q8). Adding a method to `Rule` breaks every implementor, but there are no implementors. After the first consumer ships, this becomes a versioned migration. The window is now.

- **Dual identity is a proven pattern.** golangci-lint (`G103`), eslint (`no-unused-vars`), SARIF (`ruleId` vs `ruleName`), and cqrs-lint (`C001` vs `missing-tx-commit`) all separate stable ID from display name. It is not cqrs-lint-specific; it is a universal linter design pattern.

- **It fixes a latent bug.** Today `Name()` is overloaded: it's both the display name AND the stable key (registry dedup, `RuleError.RuleName`, `finding.RuleName`). If a rule's display name changes across versions, suppression comments and filter configs silently break. Separating ID from Name makes the stability contract explicit and enforceable.

**Design decisions:**

1. **`ID()` goes on the `Rule` interface** (not as an optional type assertion). The stability contract is too important to be optional. Every rule must declare its stable ID.

2. **`RuleMeta` gets an `ID` field.** `RuleFunc.ID()` returns `r.Meta.ID`. The struct-literal pattern is preserved.

3. **`ID()` defaults to `Name()` when empty** for ergonomic migration (a rule with `ID: ""` uses its `Name` as the ID), but the documentation should strongly recommend explicit IDs for any rule that might be referenced in suppression comments or config files.

4. **`Registry.Register` deduplicates on `ID()`, not `Name()`.** Two rules with the same ID is a programming error (panic). Two rules with the same display Name is fine.

5. **`RuleError.RuleName` should carry the `ID()`**, not the display name — error attribution needs the stable key for programmatic matching.

6. **`finding.RuleName` on emitted findings should be the `ID()`** — this is what suppression filters and `--enable`/`--disable` match against.

---

### Change 6 — Suppression in `Registry.Run`: REJECT

**The proposal:** Add suppression comment support to `Registry.Run` or via a `SuppressionParser` interface.

**Why rejected:**

- **Wrong layer.** `Registry.Run` is a flat sequential loop for simple linters. Suppression filtering requires parsing source files for inline comments, which is a processing step that belongs in the pipeline — between detection and reporting. `go-finding/pipeline` already provides this via the `Processors []FindingTransformer` config field and the `suppression` package.

- **Duplication risk.** Adding suppression to `Registry.Run` would create a second suppression implementation parallel to the pipeline's. Consumers using the pipeline (the intended path for anything beyond toy linters) would have suppression applied twice or inconsistently.

- **`Registry.Run` is the simple path; the pipeline is the production path.** Linters that need suppression should use the pipeline, not `Registry.Run`. The SDK's job is to provide `DetectorsFromRegistry` (Change 2) so rules plug into the pipeline where suppression already works.

---

### What to actually implement

Ordered by impact-to-effort ratio:

1. **`DetectorsFromRegistry`** (Change 2) — the single highest-value change. Returns `[]finding.Detector`, one per rule, for pipeline consumption. Low-risk, backward compatible, benefits all consumers.

2. **`ID()` on `Rule` + `ID` on `RuleMeta`** (Change 5) — breaking change, but free while there are no consumers. Migrate `Registry.Register` dedup, `RuleError`, and finding emission to use `ID()` as the canonical key. Default to `Name()` when `ID` is empty.

3. **Documentation: finding-level Confidence and FixStrategy** (Change 3) — add a "Building Findings" section showing `finding.NewBuilder(...).WithConfidence(...).WithFixStrategy(...)`. Explain that `RuleMeta.Sev` is a default, not a constraint.

4. **Documentation: Category is open** (Change 4) — add a doc comment on the `Category` const block: "These are recommended values. Define your own `Category("your-domain")` for domain-specific taxonomies."

5. **Do NOT implement:** `CheckContext` (Change 1), rule-level `Confidence()`/`AutoFixable()` (Change 3 API addition), category validation machinery (Change 4 Option C/D), suppression in `Registry.Run` (Change 6).

### What the feedback got wrong

For the record, to prevent these claims from propagating into future design decisions:

1. **"The SDK's 8 Category constants are too restrictive"** — false. `Category` is `type Category string`. Custom categories compile and work today. The gap is documentation, not types.

2. **"The SDK has no concept of fixability"** — misleading. `finding.FixStrategy` (none/suggest/direct/ai) exists at the finding level with full `Builder` support. The SDK inherits this via its `go-finding` dependency. The gap is documentation, not API.

3. **"Confidence is not in the SDK"** — misleading. `finding.Confidence` (float64 with Low/Medium/High constants) exists at the finding level. Same documentation gap.

4. **"Registry.Run doesn't integrate with the pipeline"** — true but expected. `Registry.Run` is the simple-path execution loop. Pipeline integration is via `DetectorFromRegistry` (today) and `DetectorsFromRegistry` (proposed). The two paths serve different complexity tiers.
