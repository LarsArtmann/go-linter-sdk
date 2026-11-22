# Domain Language

> Ubiquitous-language glossary for `go-linter-sdk`. These terms shape how every
> file in this repo should be read and written. When code and this file
> disagree, the **code** is the source of truth — fix this file.

## Bounded context

This SDK has one bounded context: **lint rule execution**. A linter is a
collection of rules that examine a directory and emit findings. The SDK
provides the scaffolding (types, registry, execution, error attribution); the
consumer provides the rules. There is no converter layer — rules emit
`finding.Finding` directly.

The SDK depends on [`go-finding`](https://github.com/larsartmann/go-finding)
for the output types (`Finding`, `Report`, `Severity`, `Detector`, plus the
per-finding `Confidence` / `FixStrategy` / `GroupID` vocabulary — see below)
and nothing else. This is deliberate: any tool in the ecosystem can adopt the
SDK without coupling to a specific runner or framework.

## Terms

### Rule

A single lint check. The atomic unit of a linter. A rule declares its identity
(ID, name, description, category, severity) and a `Check` function that examines
a directory and returns zero or more findings.

Every rule has **dual identity** with distinct stability contracts:

- **ID** — the stable identifier. Never changes once published. Used for
  registry deduplication, suppression matching, filter config, and the
  `finding.RuleName` field. Think `"G001"`, `"C001"`.
- **Name** — the display name. Human-readable, can change across versions
  without breaking suppression or filtering. Think `"missing transaction
commit"`.

Rules are the **only** place domain knowledge about "what is wrong with this
code" lives. Everything else in the SDK is plumbing. A linter is a set of
rules; the SDK runs them.

- **Lives at:** `rule.go` — `type Rule interface`
- **Key contract:** `ID()` returns the stable identifier; `Check(ctx, dir)`
  emits findings directly, no intermediate type.

### RuleFunc

The common-case adapter that satisfies `Rule` by combining a `RuleMeta` header
with a `Run` closure. Most rules are written as a `RuleFunc` struct literal,
not by manually implementing the `Rule` interface.

A `RuleFunc` is **data-shaped**: the metadata reads like a struct, and the
`Check` logic is a function. This makes rules easy to scan in a rules.go
file.

- **Lives at:** `rule.go` — `type RuleFunc`
- **Error wrapping:** if `Run` returns an error, `Check` wraps it into a
  `RuleError` carrying the rule's ID — the single chokepoint.
- **Helpers:** `RuleFunc.NewFinding(message, pos)` returns a pre-stamped
  `finding.Builder` carrying the rule's ID, tool name, severity, and category
  from `RuleMeta`; `OptIn(rf)` wraps a `RuleFunc` as disabled-by-default.

### RuleMeta

The declarative identity of a rule: its ID, name, description, category, and
severity. Supplied as a struct literal so a rule reads like a data record, not
a method set.

`ID` is **required** — it is the stable identifier that never changes once
published. The registry panics on empty IDs at registration.

Fields use abbreviated names (`Cat`, `Sev`) intentionally — they appear in
every rule definition and the verbosity cost compounds.

- **Lives at:** `rule.go` — `type RuleMeta`
- **Validation:** all four identity fields (ID, Name, Description, Category)
  are checked at registration via `validateRuleIdentity`, which works through
  the `Rule` interface for any implementation. `RuleMeta.Validate()` provides
  the same check for use during rule construction before the registry.
  `Register` panics if any field is empty.

### Category

A taxonomy string classifying _what kind of issue_ a rule detects. Used for
filtering, grouping, and reporting. The type is `type Category string` — an
**open** type. Eleven values are provided as recommendations (`design`,
`structure`, `error-handling`, `correctness`, `naming`, `documentation`,
`best-practice`, `style`, `performance`, `security`, `configuration`), but
consumers can define their own for
domain-specific taxonomies (e.g., `Category("api")`, `Category("boilerplate")`).

A category is orthogonal to severity: `CategorySecurity` + `SeverityWarning`
and `CategoryStyle` + `SeverityError` are both valid combinations.

- **Lives at:** `rule.go` — `type Category` (const block for recommended values)
- **Maps to** `finding.Category` at the finding boundary.

### Registry

A thread-safe collection of rules that drives standalone execution
(`Run`), BuildFlow integration (`DetectorFromRegistry`), and pipeline
integration (`DetectorsFromRegistry`). A linter registers all its rules
(typically in `init()` or a constructor) and the registry handles the rest.

`Register` panics on duplicate rule IDs or any empty identity field (ID, Name,
Description, Category) — these are programming errors that should surface at
startup, not silently at runtime.

- **Lives at:** `registry.go` — `type Registry`
- **Thread safety:** `sync.RWMutex`; `Register` writes, `All`/`Run` read.
- **Beyond registration and `Run`:** `Get`/`Has` lookup, `Deregister` for
  runtime removal (snapshot semantics mid-run), `NewRegistry(WithToolName(...))`
  to stamp tool identity onto findings and the report, `FilterRules` (for
  `[]RuleFunc` slices) and `SelectRules` (for `[]Rule`, OptIn-safe) for
  `--enable`/`--disable` set logic, and `ExitCodeFromReport` /
  `ExitCodeFromFindings` / `ExitCodeByConfidence` to map a report or finding
  slice to process exit codes.

### RuleError

An error wrapper that attributes a failure to a specific rule by ID. When a
rule's `Check` returns an error, it is wrapped exactly once into a
`*RuleError`. Without this, a `Registry.Run` over many rules could only report
"something failed" — the caller could not tell _which_ rule.

The companion `ErrRuleFailed` sentinel lets callers check "did any rule fail?"
without caring which. Use `errors.Is(err, ErrRuleFailed)` for the boolean
check; use `errors.AsType[*RuleError](err)` (Go 1.26+) docguard:allow (when the API was introduced, not a pin) to recover the rule ID.

- **Lives at:** `errors.go` — `type RuleError` (var `ErrRuleFailed`)
- **No double-wrapping:** `wrapRuleError` (`registry.go`) passes through if
  the error is already a `*RuleError`, so `RuleFunc.Check`'s wrap is never
  re-wrapped by `Registry.Run`, `DetectorFromRegistry`, or
  `DetectorsFromRegistry`.

### RuleErrors

A helper function that extracts every `*RuleError` from a (possibly joined)
error tree. Essential when `Registry.Run` is called with `ContinueOnError()`,
which joins all rule failures into a single error — callers need to enumerate
the individual failures by rule ID.

Handles both `errors.Join` results (`Unwrap() []error`) and single-chain
wrapping (`fmt.Errorf("...: %w", inner)`), traversing the full tree.

- **Lives at:** `errors.go` — `func RuleErrors(err error) []*RuleError`

### ErrMissingFields

The sentinel error for rule identity validation failures. When a rule's ID,
Name, Description, or Category is empty, `Register` wraps the failure in
`ErrMissingFields` (via `validateIdentityFields`). Consumers can use
`errors.Is(err, ErrMissingFields)` to distinguish "bad rule definition" from
"rule execution failure" (`ErrRuleFailed`).

- **Lives at:** `rule.go` — `var ErrMissingFields`

### DetectorFromRegistry / DetectorsFromRegistry

The two integration paths from `Registry` to the `go-finding` ecosystem:

- **DetectorFromRegistry** — returns a single `finding.Detector` that runs
  every rule sequentially. Used for BuildFlow DAG integration.
- **DetectorsFromRegistry** — returns `[]finding.Detector` (one per rule) for
  `go-finding/pipeline` integration with per-rule parallelism, timeouts, and
  error isolation.

Both read the working directory from the context and wrap rule execution
errors into `*RuleError`.

- **Lives at:** `registry.go`

### CoverageReporter

An optional `Rule` extension for rules that can measure their own scan
coverage: `Coverage(ctx, dir)` returns how many files the rule actually
examined and which nested module roots it deliberately skipped. The registry
and both detector adapters fold this evidence into the report
(`Summary.FilesScanned` / `Summary.SkippedModules`) through go-finding's
`DetectResult` channel.

Coverage is **evidence, not a filter** — it changes nothing about which
findings a rule emits. Its reason to exist: without it every report renders
`filesScanned: 0`, and a fleet dashboard cannot tell "clean tree" from
"scanned nothing". Implement it on any rule that walks a directory tree.

- **Lives at:** `rule.go` — `type CoverageReporter`
- **Called once per rule per run, after `Check`** — return evidence for the
  most recent scan.
- **Known gap:** `OptIn` wraps the `RuleFunc` value, so an opt-in rule cannot
  carry coverage through the wrapper (relevant to TODO_LIST #29 fleet
  adoption; opt-in rules stay coverage-silent).

### RuleSet

An immutable, order-preserving rule collection with O(1) ID lookup — the
mutex-free alternative to [Registry](#registry) for consumers that assemble
a fixed rule set once and never mutate it. Same validation contract as
`Register` (panics on duplicate IDs or empty identity fields); the input
slice is copied.

RuleSet holds data, not execution: `Check(ctx, dir)` runs the mutex-free
findings-only path (Relativize-normalized, fail-fast), while
`ToRegistry(opts...)` converts to a mutable registry when a full
`*finding.Report` with coverage evidence is needed. Execution logic lives in
exactly one place — the registry — never duplicated here.

- **Lives at:** `ruleset.go` — `type RuleSet` (`NewRuleSet`, `Get`, `Has`,
  `Rules`, `Len`, `ToRegistry`, `Check`)
- **Measured:** `Get` ~7–10 ns/op flat to 1000 rules (≈ half of Registry's
  map+mutex lookup); `Check` 0 allocs end-to-end.

### Filter

A severity/category gate over finding slices (`Apply`/`Keep`), distinct from
`SelectRules`/`FilterRules` — those choose WHICH RULES run, the Filter
chooses WHICH FINDINGS survive. The zero value keeps everything; criteria
compose (`MinSeverity`, `Categories` as a set-literal map for O(1)
per-finding membership). `Apply` never mutates its input.

- **Lives at:** `filter.go` — `type Filter`

### SelectRules

The `--enable`/`--disable` set logic over a `[]Rule` slice — the OptIn-safe
sibling of `FilterRules` (which takes `[]RuleFunc`). `OptIn` returns an
unexported wrapper type, so opt-in rules cannot ride a `[]RuleFunc`;
`SelectRules` accepts them and lets the caller gate registration off
`IsEnabledByDefault()` instead of comparing rule IDs. Disable wins over
enable; order is preserved.

- **Lives at:** `registry.go` — `func SelectRules(all []Rule, enable,
  disable map[string]bool) []Rule`
- **Named over the plan's `FilterRulesRules`** (rejected: reads like a
  typo/stutter) — decision recorded in CHANGELOG and docs/decisions/.

## Borrowed vocabulary (go-finding)

The SDK's output side is `go-finding`'s vocabulary. Rules emit these types
directly; the SDK never translates them. Definitions below mirror the pinned
source in the module cache (version: see `go.mod`). The vocabulary keeps
growing across releases: recent releases add typed fix edits
(`Finding.Edits` / `TextEdit`) and
`Report.SkippedModules` (nested `go.mod` boundaries the tool did not analyze).

- **Finding** — one detected issue: identity (`ID`, `Rule`, `ToolName`),
  location (`Position`), `Message`, `Severity`, plus optional `Category`,
  `Confidence`, `FixStrategy`, `GroupID`, `Edits` (typed `[]TextEdit` fix list,
  authoritative for applying a fix; never reconstruct it from the display pair),
  and suppression metadata. Built via
  `finding.NewBuilder(...)` (validating) rather than struct literals.
- **Report** — the aggregated result of a run: `ToolInfo`, deduplicated
  findings, and metadata such as `FilesScanned` and `SkippedModules`.
  `Registry.Run` returns one; `ExitCodeFromReport` consumes one.
  Accumulation semantic: `AddDetectResult` SUMS per-rule counts, so
  `Summary.FilesScanned` is N·F (N rules × F files each walked), not a
  unique-file count — deliberate evidence-of-work, pinned by cmdguard/lint's
  and go-humanize-linter's tests.
- **Severity** — impact of a finding: `info` < `warning` < `error` <
  `critical`. A rule declares its default severity in `RuleMeta.Sev`; it can
  still emit per-finding severities that differ.
- **Confidence** — how sure the rule is that this specific finding is real:
  `ConfidenceNone` (0.0), `ConfidenceLow` (0.25), `ConfidenceMedium` (0.5),
  `ConfidenceHigh` (0.75), `ConfidenceFull` (1.0). Per-finding, via
  `Builder.WithConfidence` — the SDK deliberately has no rule-level default.
- **FixStrategy** — what remediation exists for a finding: `none`,
  `suggest` (human-readable hint only), `direct` (mechanically applicable),
  `ai` (reserved; pipeline triage currently treats it as `suggest`).
  `HasFix()` asks "any fix?"; `IsAutoFixable()` asks "can the machine apply
  it?" (strict subset: `direct` only). Per-finding, via
  `Builder.WithFixStrategy`.
- **GroupID** — optional branded string linking findings that belong to one
  logical issue (e.g. a clone group: N findings, one root cause).
  `Report.GroupFindings()` collapses them for triage. Rules set it via
  `Builder.WithGroupID`; findings without one are ungrouped.
- **Detector** — the ecosystem's execution interface (`Detect(ctx)`,
  named). The SDK's two adapters (`DetectorFromRegistry`,
  `DetectorsFromRegistry`) exist purely to present rules as detectors.

**Fix outcomes and rollback are NOT SDK concepts.** Applying `direct` fixes,
recording per-finding fix outcomes, and scoped rollback belong to the
`go-finding/pipeline` module's FixEngine — a separate module this SDK does
not import. The SDK's contract ends at declaring `FixStrategy`; consumers
that want detect → fix → verify loops adopt the pipeline themselves via
`DetectorsFromRegistry`.

## Relationships

```
Rule (interface)
 ├── ID() → stable identifier (never changes)
 ├── Name() → display name (mutable)
 ├── RuleFunc (adapter) ── RuleMeta (identity) + Run (closure)
 │                        ├── ID (required)
 │                        ├── Category (open string type, shape-validated)
 │                        ├── ToolName (stamped onto NewFinding output)
 │                        ├── NewFinding() → pre-stamped finding.Builder
 │                        └── finding.Severity (from go-finding)
 ├── OptIn(RuleFunc) → disabled-by-default rule
 └── Check() → []finding.Finding (output, from go-finding)

CoverageReporter (optional Rule extension)
 └── Coverage() → (filesScanned, skippedModules) → report coverage evidence

Registry ──holds──→ []Rule
 ├── Run() → *finding.Report              (standalone CLI path)
 │     └── findings pass through finding.Relativize (portable IDs)
 │     └── per-rule DetectResult feeds report.AddDetectResult
 ├── DetectorFromRegistry() → finding.Detector     (BuildFlow DAG path)
 └── DetectorsFromRegistry() → []finding.Detector  (pipeline path, per-rule)
       └── both adapters ALSO satisfy finding.DetectorV2 (coverage channel)

RuleError ──wraps──→ rule failure (carries RuleID)
ErrRuleFailed ──sentinel matched by──→ errors.Is
ErrMissingFields ──sentinel──→ empty identity fields at registration
ErrInvalidCategory ──sentinel──→ malformed Category shape
```

The output vocabulary (`Finding`, `Report`, `Detector`, `DetectResult`,
`Relativize`) is `go-finding`'s, borrowed verbatim — the SDK never
translates it.

## What is NOT in this domain

- **No Violation / Issue type.** Rules emit `finding.Finding` directly. This
  is the core design decision — it eliminates the converter layer that
  `branching-flow` (1,871 LOC) and `erraudit` (1,214 LOC) each
  maintain.
- **No CLI binary.** The SDK is library-only today. A `cmd/` wrapper is a
  ROADMAP item, not a domain concept.
- **No configuration parsing.** Rules are registered programmatically. There
  is no YAML/TOML config layer in this domain.
