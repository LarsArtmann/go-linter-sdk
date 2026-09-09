# Migration guide — adopting go-linter-sdk in a LarsArtmann linter

**Audience:** maintainers of `branching-flow`, `erraudit`, `go-structure-linter`,
and future linters in the ecosystem.
**Claim being proven:** the SDK deletes the converter layer. The pilot ports in
this repo quantify it; this guide walks the path end-to-end.

## The one-paragraph version

Your linter today carries three layers: a custom `Rule`-ish interface, a
registry, and — the expensive one — code that converts your native
Violation/Detection type into `finding.Finding`. The SDK absorbs all three.
A rule becomes a `linter.RuleFunc` whose `Run` closure emits `finding.Finding`
directly; there is nothing left to convert.

## Before you start

- Go 1.26+ and `GOEXPERIMENT=jsonv2` (the SDK depends on go-finding, which uses
  `encoding/json/v2`). In a Nix project, copy the flake apps pattern
  (`nix run .#test|lint|build`) from this repo.
- Inventory your rules and, for each, note: rule ID (must be preserved —
  suppressions in the wild reference it), severity, confidence, message
  template, and suggestion text.

## Step 1 — add the dependency

```bash
GOEXPERIMENT=jsonv2 go get github.com/larsartmann/go-linter-sdk@latest
```

## Step 2 — port ONE rule (pilot)

Start with your smallest self-contained rule. Wrap it in a `RuleFunc`:

```go
registry := linter.NewRegistry(linter.WithToolName("your-linter"))

registry.Register(linter.RuleFunc{
    Meta: linter.RuleMeta{
        ID:          "your_stable_rule_id", // PRESERVE the original ID
        Name:        "human display name",
        Description: "what it detects",
        Cat:         linter.CategoryStyle, // or your own Category("domain")
        Sev:         finding.SeverityWarning,
        ToolName:    "your-linter", // or let WithToolName auto-stamp it
    },
    Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
        // your detection logic, unchanged
    },
})
```

Emit findings directly from the detection site:

```go
finding.NewBuilder(ruleID, toolName, message, severity,
    finding.Pos(finding.FilePath(path), line, 0),
).
    WithConfidence(0.85).              // per-finding confidence
    WithFixStrategy(finding.FixStrategySuggest).
    WithSuggestion(suggestion).
    WithMetadata(map[string]string{ /* context */ }).
    MustBuild()
```

Or, inside a `RuleFunc`, use `rule.NewFinding(msg, pos)` — it pre-fills ID,
tool name, severity, and category from `RuleMeta`.

**Compatibility checklist for each ported rule:** same ID · same severity ·
same confidence · same message and suggestion text · same metadata keys. The
pilot ports in `examples/` demonstrate all of these.

## Step 3 — wire execution and exit codes

```go
report, err := registry.Run(ctx, dir, linter.ContinueOnError())
if err != nil {
    for _, ruleErr := range linter.RuleErrors(err) {
        log.Printf("rule %s failed: %v", ruleErr.RuleID, ruleErr.Cause)
    }
}
os.Exit(linter.ExitCodeFromReport(report)) // 0 clean, 1 findings
```

CI variants: `ExitCodeByConfidence(report, finding.ConfidenceHigh)` returns 2
when findings exist but all are below threshold — "triage" vs "must fix".

## Step 4 — what you can delete

Per rule ported, delete: the rule's native Detection/Violation type (if it
exists only for conversion), the `to_findings`/bridge function, and — once ALL
rules are ported — the entire converter package.

**Measured on the pilot ports (2026-09-09):**

| Linter              | Rule deleted                         | Original LOC                                                            | Port LOC                                              |
| ------------------- | ------------------------------------ | ----------------------------------------------------------------------- | ----------------------------------------------------- |
| go-structure-linter | `NoGoModRule`                        | ~40 + LinterRule base + types.Project                                   | 82 (complete standalone CLI)                          |
| branching-flow      | `nakedreturn` (`NAKED_RETURN_GUARD`) | 156 (analyzer + Detection + converter) + shared astutil deps            | 195 (complete standalone CLI, docs included)          |
| erraudit            | `oops_redundant_guard`               | 308 (detection core 206 + go/analysis autofixer 102) + violation bridge | ~350 (complete standalone CLI incl. detection + docs) |

The port LOC numbers include `main.go` wiring that the originals don't count
(registry, report printing, exit codes) — infrastructure the SDK now provides
once, instead of per-linter. At full-migration scale (branching-flow's 1,871
LOC, erraudit's 1,214 LOC of converters) the deletion dominates.

## Step 5 — BuildFlow / pipeline integration (optional)

- `linter.DetectorFromRegistry(registry, toolName)` — one detector for
  BuildFlow's DAG.
- `linter.DetectorsFromRegistry(registry)` — one detector per rule for
  go-finding/pipeline with per-rule parallelism, timeouts, and isolation.

## Common pitfalls

- **Do not invent new rule IDs.** Suppressions (`//tool:ignore ID`) in users'
  code reference the old IDs.
- **Do not run `go mod tidy` with a `go.work` workspace active** — workspace
  mode pulls sibling test dependencies into your `go.mod` and CI's tidy check
  will fail. Use `GOWORK=off go mod tidy`.
- **Coverage gates:** example/integration binaries don't attribute coverage;
  measure library packages only.
