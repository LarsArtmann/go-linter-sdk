# ADR 0002 — Rule-level fix representation: declare, don't apply

**Status:** Proposed (2026-11-19, plan task M15) — recommendation recorded,
owner decides

**Context:** The 2026-11-19 fleet review found the SDK's biggest go-finding
gap: `FixStrategyDirect` and the `Edits`/`TextEdit` fix representation exist
in the pinned library (v1.15.1), but no SDK rule emits a machine-applicable
fix, and the oops pilot port explicitly deleted erraudit's 102-LOC autofixer
(`examples/oops-redundant-guard/main.go:5-7`). The fleet-wide consequence:
linters built on this SDK cannot reach go-finding's Direct-fix pipeline.
Plan tasks F15.1–F15.5 asked for a design, a sketch, a mapping, and a
go/no-go.

## Facts gathered (F15.1, verified from source)

1. **go-finding root module owns the whole fix DATA model:**
   `finding.TextEdit` (`text_edit.go:28`), `Builder.WithEdits`
   (`finding_builder.go:80`), `Finding.HasFix()` / `IsAutoFixable()`
   (`finding_methods.go:91,111`), and `ApplySimpleFixes`
   (`simple_fix.go:27`, root module — no extra dependency).
2. **Application belongs to hosts/pipeline:** the go-finding/pipeline module
   ships a FixEngine (apply → verify → rollback); this SDK does not import
   it (single-dep policy: root module only).
3. **BuildFlow's repair contract is anti-lie by design**
   (`toolsdk/spec.go:197-237`): a `Repairer` returns only a
   `RepairResult{Description, ModifiedFiles, Facts}`; BuildFlow NEVER trusts
   self-reported fix counts — it re-runs Detect and measures the delta.
   Dry-run rides the context (`toolsdk/dryrun.go:10`, `WithDryRun`).
4. **The SDK's layering rule** (AGENTS.md 2026-09-17): rules (this SDK) →
   report (go-finding) → CLI shell (cmdguard) → rendering (go-output). The
   single-dep policy is `go-finding` root only; importing `toolsdk` in
   library code would break it.

## The design (F15.2 sketch)

**Declare at emission, apply at the boundary. No new SDK interface in v0.5.x.**

```go
// Inside a rule's Run closure — the rule KNOWS the fix:
return []finding.Finding{
    rule.NewFinding("redundant nil guard", pos).
        WithConfidence(finding.ConfidenceHigh).
        WithFixStrategy(finding.FixStrategyDirect). // machine-applicable
        WithEdits([]finding.TextEdit{
            {Start: finding.Pos(file, ifLine, 1), End: finding.Pos(file, nilLine+1, 1),
             NewText: "return oops.Wrap(err)"},
        }).
        MustBuild(),
}, nil
```

- Rules that CAN produce a safe mechanical edit attach it at emission time
  via `WithEdits` + `FixStrategyDirect`. The finding is then authoritative
  for its own fix — the display pair (`BeforeCode`/`AfterCode`) stays
  presentational.
- **Application is host-side**, three lanes, all already available:
  1. **Simple host:** `finding.ApplySimpleFixes(report, dir)` (root module)
     for text substitutions.
  2. **BuildFlow lane:** the linter's `main`/tool package implements
     `toolsdk.Repairer` (≈20 lines: read `WorkingDirFromContext`, honor
     `DryRunFromContext`, apply the report's edits, return a
     `RepairResult`; BuildFlow re-detects and measures) — same shape as the
     README's `Spec.Detect` recipe, host-side, so the SDK's dependency tree
     stays untouched.
  3. **Pipeline lane:** `go-finding/pipeline` FixEngine consumes the same
     findings unchanged.

**Erraudit autofixer mapping (F15.3):** the deleted 102-LOC `oopsfix` maps
1:1 onto this design — its "replace the if-block + `return nil` with a single
`return oops.Wrap(err)`" rewrite is one `TextEdit{Start, End, NewText}` per
finding; its guard-detection stays in the rule; its file-rewriting logic
(atomic write, encoding preservation) is exactly what `ApplySimpleFixes` /
`go-atomic-write` already own, so the port DELETES that third instead of
re-adding it.

## Alternatives rejected

1. **SDK-owned `Fixer` interface + `Registry.Fix(...)`** — rejected: adds a
   second execution path whose only job is file mutation; duplicates
   pipeline's FixEngine; puts IO/atomic-write concerns inside a library that
   stops at findings; breaks the single-dep policy the moment it wants
   crash-durable writes (go-atomic-write).
2. **Re-port the erraudit autofixer into the oops example** — rejected as a
   library move; acceptable later as an EXAMPLE showing `WithEdits`
   declaration (see the go/no-go below).
3. **`toolsdk` import inside the SDK** — rejected: same policy violation as
   cmdguard/go-output (AGENTS.md layering decision).

## Go/no-go (F15.5)

- **NO-GO:** SDK-level fix application (no `Fixer` interface, no `Registry`
  fix path) in v0.5.x.
- **GO (follow-up, TODO-able):** teach declaration in an example — extend
  `examples/oops-redundant-guard` to attach `WithEdits` +
  `FixStrategyDirect` to its findings (no application code, ~15 LOC), and a
  README subsection pointing hosts at `ApplySimpleFixes` and the toolsdk
  `Repairer` recipe. This restores the capability the port lost (declared
  fixes) without breaking layering (application stays host-side).
- The design doc routes the GO items to TODO_LIST at harvest (M27).

**Impact if adopted as proposed:** fleet linters gain fix DECLARATION (the
missing row 11 capability moves Missed → Partially), and every host gets
application for free from the ecosystem it already runs in.
