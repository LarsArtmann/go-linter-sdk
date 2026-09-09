# Status Report — Repo Publication & Auth Removal

**Generated:** 2026-09-09 02:10 CEST
**Session scope:** evaluate public vs. private → make repo public with "no guarantees" stance → remove dead private-dependency scaffolding.
**Format note:** written as Markdown per explicit user instruction (status-report skill default is styled HTML — user override wins).

---

## Self-Review (brutal, per session)

### What did I forget?

1. **Sequencing around an irreversible flip.** I flipped the repo to public while the four corrective edits (README, ci.yml, AGENTS.md, CHANGELOG.md) were still uncommitted. The auto-git daemon committed them only at 02:06:41 (this morning) — so for the whole window between the flip (previous session) and 02:06, the **public** repo displayed the stale "go-finding is private, set GOPRIVATE" warning and the old CI. Harm: low (following the stale instructions is wasteful, not breaking), but the sequencing was wrong: content correctness must precede an irreversible exposure flip.
2. **CONTRIBUTING.md staleness.** It still claims "`go-finding` is resolved from VCS as a published tag (`v1.4.1` in `go.mod`)" — go.mod says v1.7.0. I read CONTRIBUTING during the session and did not fix it. Split brain, still live.
3. **The secret itself.** I removed all _usage_ of `PRIVATE_REPO_TOKEN` from ci.yml but left the secret sitting in GitHub repo settings — a now-unused credential provisioned (per old AGENTS.md lore) from the maintainer's broad `gh auth token`. Off-boarding auth means deleting/rotating the credential, not just the workflow steps.
4. **Verification scope of the stranger test.** I proved `go-linter-sdk@v0.2.0` + `go-finding v1.6.0` resolve via proxy.golang.org. The _current_ master requires go-finding **v1.7.0** — that exact version was never fetched through the public proxy in my test. If v1.7.0 had been tagged while the repo was private and the proxy cached a negative result, CI's `go mod tidy` would 410. Unverified gap; CI's next run is the real test.
5. **CI has never run the new workflow.** The auth-free ci.yml shipped in commit `7675912` (02:06 today); no push has triggered it yet. My YAML check was a local parse, not a workflow run.
6. **Tagged docs are frozen wrong.** pkg.go.dev renders v0.2.0's README — which still contains the false private-dependency warning — until a new tag is cut. I flagged the stale master README but not the frozen-tag consequence.

### What is stupid that we do anyway?

- **Dependency versions written in 4 places** (go.mod, README "v1.4+", CONTRIBUTING "v1.4.1", AGENTS.md) — a permanent drift machine; two of the four were wrong this session.
- **A broad PAT sat in repo secrets** for a dependency that is now public — dead weight with real blast radius.
- **Internal strategy documents in a public repo** (docs/status/_, docs/planning/_ expose the private ecosystem map: branching-flow, erraudit, go-structure-linter + LOC metrics). Accepted deliberately per "make it public, guarantee nothing" — but it remains the single largest irreversible disclosure of this session.

### What could I have done better?

- Commit → verify CI green → _then_ flip visibility.
- Run the stranger test against the **current** go.mod dependency set (v1.7.0), not just the last tag.
- Delete the unused secret in the same operation that removed its usage.
- Verify the pkg.go.dev claim instead of asserting an expectation.

### What could you still improve?

See sections (e) and (f).

### Did I lie?

No. Two statements were expectations presented as expectations, not verified facts: "pkg.go.dev should index within the hour" and "CI needs no auth" (reasoned, not yet exercised by a real run). Everything marked as verified below was actually run.

---

## a) FULLY DONE

| Item                                                                                                                           | Evidence                                                                              |
| ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------- |
| Publication readiness audit (visibility, deps, license, history secret scan)                                                   | `gh repo view` ×3, `git log --all` filename + token-pattern scans — all clean         |
| README stale private-dependency warning removed                                                                                | commit `7675912`, README.md −2 lines                                                  |
| CI private-auth scaffolding removed (3 auth steps, `GOPRIVATE` env, 10-line secret header comment)                             | ci.yml −19 lines, zero `secrets.` refs remain, YAML parses valid                      |
| AGENTS.md corrected (public facts, go-finding v1.6.0→v1.7.0, token lore replaced)                                              | commit `7675912`                                                                      |
| CHANGELOG "Unreleased" entry documenting the flip                                                                              | commit `7675912`                                                                      |
| Repo flipped to public                                                                                                         | `gh repo edit --visibility public` → confirmed `PUBLIC`                               |
| End-to-end stranger test (fresh module, `GOPROXY=proxy.golang.org` only, no auth): `go get go-linter-sdk@v0.2.0` + build + run | resolved go-finding v1.6.0 + go-error-family v0.10.0; program printed expected output |
| Local build + full test suite green after all edits                                                                            | `GOEXPERIMENT=jsonv2 go build ./... && go test ./... -count=1` → ok                   |
| Temp verification module cleaned up                                                                                            | `trash /tmp/pubtest` (no `rm`)                                                        |
| All four edits committed (auto-git daemon, 02:06:41)                                                                           | `7675912`, working tree clean before this report                                      |

## b) PARTIALLY DONE

1. ~~**The publication itself.** Source-side is complete and committed; the _verification_ side is not: no CI run on the auth-free workflow, no pkg.go.dev confirmation, no goreportcard badge check.~~ done (verified 2026-09-09): CI green end-to-end on `487d254`; pkg.go.dev indexes v0.3.0; goreportcard checked — service sunset, badge removed from README
2. ~~**Public consumption proof.** Proven for v0.2.0's dependency set (go-finding v1.6.0); not proven for master's go-finding v1.7.0.~~ done (2026-09-09): proxy-only `go mod download` of the master dep set succeeds, and the green CI run compiles v1.7.0 end-to-end
3. **Auth off-boarding.** Usage removed; the credential itself still exists in repo settings.
4. **"Guarantee nothing" stance.** It exists as an instruction to me, but is not yet encoded anywhere public (no as-is line in README beyond MIT, no SUPPORT.md). ← done in part — recorded in `AGENTS.md` (2026-09-09); the public statement awaits the support-posture decision (ROADMAP Q6)

## c) NOT STARTED

- Delete/rotate `PRIVATE_REPO_TOKEN` (and review the broad PAT it came from).
- ~~CONTRIBUTING.md v1.4.1 → v1.7.0 fix.~~ done — version references now point at `go.mod` (2026-09-09)
- Next tag (v0.3.0) so corrected docs render on pkg.go.dev.
- ~~pkg.go.dev / goreportcard / CI-badge verification.~~ done (2026-09-09): pkg.go.dev v0.3.0 confirmed; goreportcard sunset (badge removed); CI badge green on `487d254`
- GitHub metadata: description, topics, homepage.
- Branch protection, tag protection, issue-tracker posture decision.
- SECURITY.md / SUPPORT.md / explicit no-guarantee statement.
- Deep history scan with gitleaks (broader patterns than my two regexes).
- First real consumer migration (go-structure-linter pilot) — unchanged from before this session.

## d) TOTALLY FUCKED UP

Nothing catastrophic: no broken code, no failed tests, no leaked secrets, no destructive operations.

One genuine process fuck-up, honestly owned: **the visibility flip preceded the commit.** An irreversible public exposure was executed while the corrective content existed only in my working tree, relying on the daemon's timing. The public repo showed false install instructions for the flip-to-commit window (hours). Lesson encoded in section (e): irreversible flips come _last_, after content is committed and CI-verified.

Second, smaller: **ghost credential left behind** (`PRIVATE_REPO_TOKEN` in settings, unused since `7675912`). Not leaked — just stale privilege that should not outlive its purpose.

## e) WHAT WE SHOULD IMPROVE

1. **Flip checklist for irreversible operations:** commit → CI green → verify public-facing content → then flip. Never flip from a dirty tree.
2. **Single source of truth for dependency versions.** Docs should reference "see go.mod", not repeat versions (this session found two stale copies).
3. **Credential lifecycle coupled to usage.** Removing the last reference to a secret should trigger deleting the secret in the same change.
4. **Verify at the version you ship, not the version that's convenient.** Stranger-test current master's dep graph, not just the last tag.
5. **Expectations vs. facts discipline.** Label unverified claims explicitly (I did — but only in retrospect).

## f) NEXT: 50 items (brainstorm for docs-health HARVEST — most belong in TODO_LIST/ROADMAP, sorted by impact)

**Verify the publication (highest impact, this week):**

1. ~~Push the next commit (this report) and watch CI run the auth-free workflow end-to-end.~~ done — green on `487d254` (1m11s)
2. ~~Verify go-finding v1.7.0 + go-error-family v0.10.0 resolve via proxy.golang.org (master's dep set).~~ done — proxy-only `go mod download` + `go mod verify` pass (2026-09-09)
3. ~~Confirm pkg.go.dev lists go-linter-sdk@v0.2.0; request indexing if absent.~~ done — v0.3.0 is indexed and current
4. ~~Confirm goreportcard badge resolves with an acceptable grade.~~ done-with-twist — the service has been **sunset**; the dead badge was removed from the README (2026-09-09)
5. Delete `PRIVATE_REPO_TOKEN` from repo settings.
6. Rotate/review the source PAT (old notes say it was the maintainer's full `gh auth token`).
7. gitleaks deep scan over full history (broad + entropy patterns).
8. Review docs/status, docs/planning, docs/feedback for genuinely sensitive ops details (token provisioning narratives); prune or consciously accept.
9. ~~Fix CONTRIBUTING.md "v1.4.1 in go.mod" → drop the version, point at go.mod.~~ done (2026-09-09)
10. ~~Cut v0.3.0 so pkg.go.dev renders corrected install docs.~~ done-but-insufficient — v0.3.0 was cut 2026-09-08 yet **predates** the corrected README (`7675912`); pkg.go.dev still shows the false private-dep warning. A NEW tag (v0.4.0) is required — TODO_LIST #1

**Public-repo hygiene:**
11. GitHub description + topics (go, linter, static-analysis).
12. Branch protection on master (require CI checks).
13. Tag protection for `v*`.
14. Decide issue-tracker posture (open vs. disabled).
15. SECURITY.md (minimal vuln-reporting path).
16. SUPPORT.md or explicit "as-is, no guarantees" statement in README.
17. Issue/PR templates.
18. Social preview image.
19. Set repo homepage to pkg.go.dev URL.
20. Dependabot config for Go updates.
21. Reconsider macos CI matrix cost.
22. actionlint (or equivalent) in CI for workflow validation.
23. ~~Confirm .gitignore covers go.work and reports/ artifacts (no leaks in public listing).~~ done — verified: `go.work` ignored, `reports/*` ignored except `.gitkeep` (2026-09-09)
24. Annotate future release tags (annotated, not lightweight).

**Docs coherence:**
25. ~~Kill version duplication across README/CONTRIBUTING/AGENTS (reference go.mod).~~ done — CONTRIBUTING + README now reference `go.mod`; AGENTS keeps only the deliberate v1.7.0 pin (2026-09-09)
26. Add the explicit no-guarantee line to README.
27. ~~Re-test the AGENTS.md `GIT_CONFIG_COUNT` tidy workaround — likely a ghost now that the proxy suffices; delete if dead.~~ done — proxy-only resolution verified; the note now carries the verification and its narrow remaining use case (2026-09-09)
28. ~~Prune remaining GOPRIVATE/local-git lore from AGENTS.md if proxy-only dev is verified.~~ done — AGENTS is already clean of GOPRIVATE lore; re-verified 2026-09-09
29. ~~Verify all README links work on public GitHub (examples paths, badges).~~ done — links resolve; the one dead badge (goreportcard) was removed (2026-09-09)
30. Clarify Go toolchain story for consumers (go.mod 1.26.7 + GOTOOLCHAIN auto-download).
31. ~~Note in README that the consumer-table repos are private (avoid stranger confusion).~~ done — note added under the LOC table (2026-09-09)
32. godoc quality pass on exported symbols (pkg.go.dev is the first impression now).
33. ~~Record the "public with no guarantees" decision in AGENTS.md so future sessions keep the stance.~~ done (2026-09-09)
34. ~~Run `nix flake check` / treefmt over the md files edited this session.~~ done — `nix flake check` all checks passed (2026-09-09)
35. ~~Run full `nix run .#lint` + `.#vet` (session only ran build+test).~~ done — 0 issues / clean (2026-09-09)
36. ~~Add a dated release entry for the public flip when cutting v0.3.0 (currently only in Unreleased).~~ done — the flip postdates v0.3.0, so it correctly lives in `[Unreleased]`; the missing `[0.3.0]` section was backfilled (2026-09-09)

**Ecosystem / strategic:**
37. Migrate go-structure-linter pilot to the SDK — first real consumer, the entire value prop is still unproven in production.
38. Decide public/private for branching-flow, erraudit, go-structure-linter (their names + LOC metrics are now publicly disclosed).
39. If consumers stay private, that's fine — but then the README table is advertising for invisible products; consider a public demo repo.
40. Consider a tiny public example linter as the "proof" artifact.

**Testing:**
41. Codify the stranger test: script or CI job doing fresh-module `go get` + build against the latest tag.
42. Proxy-only integration test pinning master's dep graph (catches proxy poisoning early).
43. Validate workflow YAML in CI (dup of 22 — pick one home).
44. Review Registry panic messages / error text — they are now stranger-facing API surface.
45. Consider property test for Registry dedup/identity invariants.

**Cleanup:**
46. Local env: drop global `GOPRIVATE` + SSH `insteadOf` rewrite (unneeded post-public).
47. ~~Decide retention for old docs/status snapshots (annotate per docs-health, or archive).~~ done — every report annotated inline; `2026-07-27_14-38` archived to `docs/status/archived/` (2026-09-09)
48. ~~Confirm CI badge in README shows green publicly after next run.~~ done — latest run on `487d254` is green
49. ~~Ensure no coverage/BuildFlow artifacts are committed (reports/ exception audit).~~ done — working tree clean; only `reports/.gitkeep` is tracked
50. Optional: GoReleaser or release workflow for future tags (currently manual per AGENTS.md).

## g) Questions I cannot figure out myself

1. **Secret off-boarding:** May I delete `PRIVATE_REPO_TOKEN` from the repo settings now (`gh secret delete`), and do you want the broader source PAT rotated? Deleting a credential is destructive and account-touching — I won't do it without your go-ahead.
2. **Support posture:** Should the public repo accept issues/PRs from strangers (triaged by you, still no guarantees), or run read-only/mirror with the tracker disabled? This determines SECURITY.md/SUPPORT.md wording, templates, and branch-protection setup.
3. **Flip policy:** For future irreversible operations (visibility, deletions, releases), do you accept an explicit pre-flight commit so content is verified _before_ the flip — or should I always wait for the daemon and accept the exposure window? The "never commit without your say-so" rule is what produced this session's gap.

---

## Verification appendix (commands actually run this session)

- `gh repo view` go-finding → PUBLIC; go-error-family → PUBLIC; go-linter-sdk → PRIVATE (pre-flip), PUBLIC (post-flip).
- `git log --all --name-only` suspicious-filename grep: none. `git log --all -p` token-pattern grep (`ghp_`, `github_pat_`): none.
- `GOEXPERIMENT=jsonv2 go build ./... && go test ./... -count=1` → ok (0.356s).
- `gh repo edit LarsArtmann/go-linter-sdk --visibility public --accept-visibility-change-consequences` → PUBLIC.
- Stranger test in `/tmp/pubtest`: `go mod init` + `GOPROXY=https://proxy.golang.org,direct GOPRIVATE= go get github.com/larsartmann/go-linter-sdk@v0.2.0` → resolved v1.6.0/v0.10.0; `go run .` → correct output.
- Post-daemon-commit: `grep secrets. ci.yml` → none; Python yaml.safe_load → valid.
- `git show --stat 7675912` → exactly the 4 intended files, −34/+12.

---

## Addendum 2026-09-09 — S4 sensitive-narrative acceptance verdict (plan task S3/S4)

Full-repo sweep (`rg` over `docs/`, README, CONTRIBUTING, SECURITY, SUPPORT) for
token-provisioning narratives, hostnames, and PAT stories found ~40 hits — all
historical process narratives in point-in-time status reports (this file
included), **zero credential values**. Corroborated by a gitleaks v8
full-history scan (109 commits, 0 findings, `reports/GITLEAKS_TRIAGE.txt`).

**Verdict: ACCEPT all hits, redact nothing.** Rationale: no secret material
exists or existed in history; the narratives describe the remediation itself
(deleting `PRIVATE_REPO_TOKEN`), which is public record by design of the open
docs policy. The one sensitive *fact* — that the old secret derived from the
maintainer's broad `gh auth token` — remains actionable only with account
access, which this repo does not confer. PAT rotation stays with the
maintainer (AGENTS.md, user decisions 2026-09-09).
