> [!NOTE]
> This filing was drafted by GLM via Crush from an AI-run investigation, not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[2026-11-19T17:xx +01:00]`

**TL;DR:** `TestAnalyzeEnv_GOWORKOffEscapesWorkspaceShadow` fails on master today: the first half of its discrimination proof (`Analyze` under a foreign go.work must ERROR) no longer holds — `Analyze` succeeds where the test demands a workspace-shadow load error, so the test asserts a behavior the loader no longer exhibits. The ask: re-derive the discrimination premise (Go workspace behavior change vs. `Analyze` env handling) and update either the test or the env plumbing.

## Problem

`pkg/sdk/sdk_test.go:257` `TestAnalyzeEnv_GOWORKOffEscapesWorkspaceShadow` fails reproducibly as of 2026-11-19 (twice today, `go test ./pkg/sdk/ -run GOWORKOff -count=1`, `GOWORK=off` not set for the failing half):

```
=== RUN   TestAnalyzeEnv_GOWORKOffEscapesWorkspaceShadow
    sdk_test.go:269: Analyze succeeded under a foreign go.work; want the workspace-shadow load error
--- FAIL: TestAnalyzeEnv_GOWORKOffEscapesWorkspaceShadow (0.19s)
```

The test (sdk_test.go:257-296) builds a temp workspace with `go.work` = `use ./other` and a target module `app` that is NOT in the workspace. Its premise, per the doc comment (sdk_test.go:250-256): loading `app` under the inherited workspace errors (the CV 2026-11-12 workspace-shadow trap), and `AnalyzeEnv(..., []string{"GOWORK=off"})` recovers — proving the env reaches the loader.

The first assertion (`Analyze` must fail, sdk_test.go:268-270) is what broke. The second half (`GOWORK=off` yields the HW-1 finding) still passes.

## What I checked

- Fails identically with the go-linter-sdk v0.6.0 bump stashed and applied (bisect-pinned 2026-10-15 during the v0.4.x wave; re-reproduced today on the current tree) — not caused by dependency changes.
- The failing assertion is the plain `sdk.Analyze(ctx, app)` call; the temp module layout and `go.work` content match the doc comment.
- Suspects worth ruling out, in order: a Go toolchain workspace-behavior change (out-of-workspace module loads silently instead of erroring), or `Analyze` no longer inheriting `GOWORK` from the process env.

## Ask

Re-derive the premise: if Go's workspace semantics changed, the test's first half needs a different shadow trigger; if `Analyze` stopped inheriting env, that is a product bug above the test. Either way the discrimination proof should stay intact — it is what guarantees `AnalyzeEnv` overrides actually reach the loader.

💘 Generated with Crush
