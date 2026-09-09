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
for the output types (`Finding`, `Report`, `Severity`, `Detector`) and nothing
else. This is deliberate: any tool in the ecosystem can adopt the SDK without
coupling to a specific runner or framework.

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
`Check` logic is a function. This makes rules easy to scan in a `rules.go`
file.

- **Lives at:** `rule.go` — `type RuleFunc`
- **Error wrapping:** if `Run` returns an error, `Check` wraps it into a
  `RuleError` carrying the rule's ID — the single chokepoint.

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
**open** type. Eight values are provided as recommendations (`design`,
`structure`, `error-handling`, `correctness`, `style`, `performance`,
`security`, `configuration`), but consumers can define their own for
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

### RuleError

An error wrapper that attributes a failure to a specific rule by ID. When a
rule's `Check` returns an error, it is wrapped exactly once into a
`*RuleError`. Without this, a `Registry.Run` over many rules could only report
"something failed" — the caller could not tell _which_ rule.

The companion `ErrRuleFailed` sentinel lets callers check "did any rule fail?"
without caring which. Use `errors.Is(err, ErrRuleFailed)` for the boolean
check; use `errors.AsType[*RuleError](err)` (Go 1.26+) to recover the rule ID.

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

## Borrowed vocabulary (go-finding v1.7.0)

The SDK's output side is `go-finding`'s vocabulary. Rules emit these types
directly; the SDK never translates them. Definitions below mirror the pinned
v1.7.0 source (module cache), not newer tags.

- **Finding** — one detected issue: identity (`ID`, `Rule`, `ToolName`),
  location (`Position`), `Message`, `Severity`, plus optional `Category`,
  `Confidence`, `FixStrategy`, `GroupID`, and suppression metadata. Built via
  `finding.NewBuilder(...)` (validating) rather than struct literals.
- **Report** — the aggregated result of a run: `ToolInfo` plus deduplicated
  findings. `Registry.Run` returns one; `ExitCodeFromReport` consumes one.
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
 │                        ├── Category (open string type)
 │                        └── finding.Severity (from go-finding)
 └── Check() → []finding.Finding (output, from go-finding)

Registry ──holds──→ []Rule
 ├── Run() → *finding.Report              (standalone CLI path)
 ├── DetectorFromRegistry() → finding.Detector     (BuildFlow DAG path)
 └── DetectorsFromRegistry() → []finding.Detector  (pipeline path, per-rule)

RuleError ──wraps──→ rule failure (carries RuleID)
ErrRuleFailed ──sentinel matched by──→ errors.Is
```

## What is NOT in this domain

- **No Violation / Issue type.** Rules emit `finding.Finding` directly. This
  is the core design decision — it eliminates the converter layer that
  `branching-flow` (1,871 LOC) and `erraudit` (1,214 LOC) each
  maintain.
- **No CLI binary.** The SDK is library-only today. A `cmd/` wrapper is a
  ROADMAP item, not a domain concept.
- **No configuration parsing.** Rules are registered programmatically. There
  is no YAML/TOML config layer in this domain.
