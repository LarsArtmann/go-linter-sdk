# Status Report: go.mod Replace Directive Removal & Real Version Pinning

**Date:** 2026-07-30 16:08 CEST  
**Session scope:** Single session — fixing `go-finding` dependency from zero pseudo-version + local `replace` to real published `v1.4.1` tag resolved from VCS.  
**Format override:** User requested `.md` (skill default is HTML). Honored per skill spec.

---

## TL;DR

The `go.mod` had a zero pseudo-version (`v0.0.0-00010101000000-000000000000`) and a `replace ../go-finding` directive. This made the SDK **completely un-importable** by any external consumer (Go strips `replace` directives when consuming a dependency). Fixed by pinning to real `v1.4.1` and removing the replace. But the fix has **real risks** — especially CI auth for the private dependency, and no local dev workflow without a `go.work`.

---

## a) FULLY DONE

| #   | Item                                                                                                          | Verification                                                          |
| --- | ------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| 1   | **Root cause identified** — commit `e51d1ee` reverted `require go-finding v1.4.0` back to zero pseudo-version | `git log -p -- go.mod` confirmed the revert                           |
| 2   | **`go.mod` pinned to `v1.4.1`** — real published tag, no pseudo-version                                       | `cat go.mod` shows `require github.com/larsartmann/go-finding v1.4.1` |
| 3   | **`replace` directive removed** entirely from `go.mod`                                                        | Confirmed — no replace line                                           |
| 4   | **`go.sum` populated** with real `v1.4.1` checksums (was empty for go-finding)                                | `grep go-finding go.sum` shows h1 hash                                |
| 5   | **`go mod tidy` succeeded** — fetched `v1.4.1` via SSH `insteadOf` rewrite (private repo)                     | Tidy completed cleanly                                                |
| 6   | **`go build ./...` passes**                                                                                   | Exit code 0                                                           |
| 7   | **`go vet ./...` passes**                                                                                     | Exit code 0                                                           |
| 8   | **`go test -count=1 ./...` passes** (forced recompile, no cache)                                              | `ok github.com/larsartmann/go-linter-sdk 0.007s`                      |
| 9   | **CI workflow updated** — removed all 3 sibling-clone steps, added `GOPRIVATE` + git `GITHUB_TOKEN` auth      | `.github/workflows/ci.yml` rewritten                                  |
| 10  | **AGENTS.md updated** — dependency note (v1.4.1, no replace) and CI gotcha replaced                           | Two sections edited with exact matches                                |
| 11  | **Corrupt Go build cache reset** — stdlib cache-miss errors resolved by trashing `~/.cache/go-build`          | Fresh cache rebuilt, tests pass                                       |

---

## b) PARTIALLY DONE

| #   | Item                                          | What's missing                                                                                                                                                                                                                                                                                      |
| --- | --------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **CI private module auth**                    | Wrote `git config url."https://${{ secrets.GITHUB_TOKEN }}@github.com".insteadOf` but **cannot verify it works locally**. The default `GITHUB_TOKEN` is scoped to the _current_ repo (go-linter-sdk), not to go-finding (a _different_ private repo). This is an **unproven assumption** — see (d). |
| 2   | **Local dev workflow for sibling go-finding** | I suggested `go.work` in the summary and AGENTS.md but **did not create one**. Right now, local dev against an uncommitted sibling checkout has NO resolution path — no replace, no go.work.                                                                                                        |
| 3   | **GOPRIVATE alignment**                       | Local `go env GOPRIVATE` lists specific repos (`github.com/larsartmann/go-cqrs-lite,github.com/larsartmann/go-finding,...`); CI uses wildcards (`github.com/larsartmann/*,github.com/LarsArtmann/*`). Functionally different patterns, neither verified against the other.                          |

---

## c) NOT STARTED

| #   | Item                                                                                                        |
| --- | ----------------------------------------------------------------------------------------------------------- |
| 1   | **Never ran `nix run .#lint`** (golangci-lint) — the canonical lint check per AGENTS.md. Only ran `go vet`. |
| 2   | **Never ran `nix run .#test-race`** — only ran `go test` without `-race`.                                   |
| 3   | **Never ran `nix run .#build`** (Nix build path) — only ran raw `go build`.                                 |
| 4   | **Never ran `nix flake check`** — didn't verify the flake still evaluates.                                  |
| 5   | **No `go.work` created** for local cross-repo development.                                                  |
| 6   | **Didn't verify CI YAML is valid** — no `actionlint` or dry-run.                                            |
| 7   | **Didn't verify go.sum is complete** beyond go-finding entries — didn't run `go mod verify`.                |

---

## d) TOTALLY FUCKED UP

### 1. CI `GITHUB_TOKEN` may NOT have access to go-finding (HIGH RISK)

**This is the biggest risk I introduced.** I wrote:

```yaml
git config --global url."https://${{ secrets.GITHUB_TOKEN }}@github.com".insteadOf "https://github.com"
```

The default `GITHUB_TOKEN` in GitHub Actions is **automatically scoped to the repository running the workflow** (go-linter-sdk). It does **NOT** automatically have read access to _other_ private repos under the same user/org. `go-finding` is a separate private repository.

**Likely outcome:** CI fails with `git could not read Username` or `403 Forbidden` when `go mod download` tries to fetch `github.com/larsartmann/go-finding@v1.4.1`.

**What I should have done:** Either:

- Used a **PAT** (Personal Access Token) stored as a repository secret (e.g., `secrets.PRIVATE_REPO_TOKEN`) with `repo` scope
- Used a **GitHub App** with installation token
- Used **SSH** with a deploy key
- Or at minimum **flagged this as a known unverified risk** and told the user explicitly

### 2. Commit history pollution by auto-git daemon

My `go.mod` fix was bundled by the daemon into commit `b77297c` titled **"docs(feedback): add CQRS lint adoption analysis"** — a completely unrelated message about documentation. The commit message claims it "Update go.mod to include the cqrs-lint dependency" which is a **lie** — there's no cqrs-lint dependency. The actual change was the go-finding version fix.

My CI change was bundled into `c9a1c26` titled **"feat(sdk): enhance rule registry, error handling, and CI pipeline"** — again mixing my CI auth change with another session's registry refactor.

Neither commit message accurately describes what changed. This is the documented auto-git daemon behavior (AGENTS.md says to live with it), but it's still a problem.

### 3. Local dev workflow is broken NOW

Before my change: local dev worked via `replace ../go-finding` → sibling checkout resolved.  
After my change: **no local path resolution exists at all.** Any developer who wants to test go-linter-sdk against an uncommitted go-finding change has no mechanism. I suggested `go.work` but didn't create one.

### 4. Wasted time on corrupt build cache

First test run failed with stdlib errors (`package compress/flate is not in std`, `could not import internal/cpu`). These were obviously a corrupt Go build cache — NOT related to my go.mod change. I should have identified this **immediately** from the error pattern (stdlib packages missing = cache corruption, not dependency issue). Instead I re-ran tests, then tried `nix run .#test` (same result), before finally identifying and fixing the cache.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Run the canonical Nix apps, not raw `go` commands.** AGENTS.md explicitly says "No Makefile. Everything via `flake.nix`." I ran `go build`, `go test`, `go vet` directly. Should have used `nix run .#build`, `nix run .#test`, `nix run .#lint` from the start.

2. **Recognize corrupt cache immediately.** Stdlib import failures (`could not import strings`, `package errors is not in std`) are ALWAYS a cache issue, never a dependency issue. Don't re-run — reset the cache immediately.

3. **Don't trust gopls during concurrent daemon commits.** The daemon commits faster than gopls re-indexes. `go build`/`go test` are the source of truth. Stop restarting gopls and just verify with the toolchain.

4. **Verify CI changes are testable.** Writing CI config that can't be locally verified is writing untested code. At minimum validate the YAML and trace the auth flow mentally.

5. **Always check for `go.work` before suggesting it.** I recommended creating a go.work but didn't check whether one already existed or whether the sibling repo had one that could be referenced.

6. **Align GOPRIVATE patterns.** Local and CI should use the same pattern. Pick wildcards or specific repos, not both.

### Technical improvements

7. **Add a `go.work` file for local dev** — this is the idiomatic Go way to work across sibling repos without committed `replace` directives.

8. **Use a PAT or GitHub App for CI private module access** — the default `GITHUB_TOKEN` is insufficient for cross-repo private fetches.

9. **Add `go mod verify` to CI** — catches checksum tampering.

10. **Consider GOPRIVATE wildcard `github.com/LarsArtmann/*,github.com/larsartmann/*`** everywhere (local + CI) to eliminate casing split-brain.

---

## f) Things to get done next (sorted by impact)

| Priority | #   | Task                                                                                                                                                          | Impact                                            | Effort  |
| -------- | --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------- | ------- |
| **P0**   | 1   | **Fix CI private repo auth** — replace `secrets.GITHUB_TOKEN` with a PAT (`secrets.PRIVATE_REPO_TOKEN`) or SSH deploy key. Verify by pushing and watching CI. | CI is likely BROKEN right now                     | Low     |
| **P0**   | 2   | **Create `go.work`** for local cross-repo dev — `go 1.26.5\nuse .\nuse ../go-finding` (gitignored or committed depending on team preference)                  | Local dev workflow is broken                      | Low     |
| **P0**   | 3   | **Run `nix run .#lint`** and fix any findings — the canonical lint was never run this session                                                                 | Unknown lint state                                | Low     |
| **P0**   | 4   | **Run `nix run .#test-race`** — race detector was never run this session                                                                                      | Unknown race state                                | Low     |
| **P1**   | 5   | **Align GOPRIVATE** — pick one pattern (wildcards recommended) and use it in both local `go env` and CI env                                                   | Casing split-brain risk                           | Low     |
| **P1**   | 6   | **Add `go mod verify`** to CI govulncheck job (or separate job)                                                                                               | Supply chain integrity                            | Low     |
| **P1**   | 7   | **Run `nix flake check`** — verify the flake still evaluates after go.mod changes                                                                             | Unknown flake state                               | Low     |
| **P1**   | 8   | **Run `go mod verify`** locally — confirm all checksums are valid                                                                                             | Integrity check                                   | Trivial |
| **P1**   | 9   | **Add `.gitignore` entry for `go.work`** (if it should be local-only) or commit it (if shared)                                                                | Prevent accidental commits                        | Trivial |
| **P2**   | 10  | **Add `actionlint`** to CI or pre-commit to validate workflow YAML before push                                                                                | Catch CI YAML errors early                        | Low     |
| **P2**   | 11  | **Update ROADMAP.md** — remove the "publish go-finding tag (ROADMAP Q1)" item since v1.4.1 is already published                                               | Doc drift                                         | Trivial |
| **P2**   | 12  | **Update FEATURES.md** — go-finding is now resolved from VCS, not a local replace                                                                             | Doc accuracy                                      | Trivial |
| **P2**   | 13  | **Add a CI badge** to README.md showing build status                                                                                                          | Visibility                                        | Trivial |
| **P2**   | 14  | **Document the SSH `insteadOf` workaround** in CONTRIBUTING.md for new contributors with read-only global git config                                          | Onboarding                                        | Low     |
| **P2**   | 15  | **Add `gitsign` or commit signing verification** to CI — the daemon commits are unsigned                                                                      | Audit trail                                       | Medium  |
| **P3**   | 16  | **Create a `make-dev` or `dev-shell` script** that sets `GOEXPERIMENT`, `GOPRIVATE`, and `GIT_CONFIG_COUNT` env vars in one command                           | DX improvement                                    | Low     |
| **P3**   | 17  | **Add `go mod tidy` check to CI** — `git diff --exit-code go.mod go.sum` after tidy                                                                           | Prevent pseudo-version regression                 | Low     |
| **P3**   | 18  | **Add a pre-commit hook** that rejects `replace` directives pointing to `../` paths                                                                           | Prevent regression of the exact bug we just fixed | Low     |
| **P3**   | 19  | **Pin golangci-lint version in flake.nix** (if not already) — CI uses `v2.12.2` but local may differ                                                          | Reproducibility                                   | Low     |
| **P3**   | 20  | **Add `dependabot` config** for Go modules and GitHub Actions                                                                                                 | Dependency hygiene                                | Low     |
| **P3**   | 21  | **Investigate `go.work.sum`** — sibling go-finding has one; determine if go-linter-sdk needs one                                                              | Correctness                                       | Low     |
| **P3**   | 22  | **Add `nix run .#coverage`** to CI — generate coverage reports                                                                                                | Test visibility                                   | Low     |
| **P4**   | 23  | **Write integration test** that imports go-linter-sdk as an external module (simulating a consumer) to catch importability regressions                        | Regression prevention                             | Medium  |
| **P4**   | 24  | **Add `CODEOWNERS` file** for dependency files (go.mod, go.sum, flake.lock)                                                                                   | Review routing                                    | Trivial |
| **P4**   | 25  | **Document the "require real version + optional go.work" pattern** in a project-level ADR                                                                     | Knowledge retention                               | Low     |

---

## g) Questions I CANNOT figure out myself

### 1. CI auth for cross-repo private modules

**Question:** Does the default `GITHUB_TOKEN` in this repo's GitHub Actions have read access to `github.com/larsartmann/go-finding` (a separate private repo)?

**Why I can't figure this out:** This depends on your GitHub account/org settings — specifically whether cross-repo token access is enabled, or whether you need a PAT/deploy key. I have no way to check your GitHub permission settings. If the answer is "no", I need to know what credential you'd prefer (PAT in secrets, SSH deploy key, GitHub App).

### 2. Should `go.work` be committed or gitignored?

**Question:** Do you want a committed `go.work` (shared across the team, assumes everyone clones repos as siblings) or a local-only gitignored one?

**Why I can't figure this out:** This is a workflow preference. Some teams commit go.work, others don't. It depends on whether all contributors use the same directory layout (`~/projects/go-finding` as a sibling).

### 3. What caused the build cache corruption?

**Question:** Do you know what corrupted the Go build cache (stdlib `.o` files missing from `~/.cache/go-build`)?

**Why I can't figure this out:** This happened before my session. Possible causes: disk full, crash, Nix GC collecting store paths that the cache referenced, or a concurrent `go clean`. If this is recurring, it needs a root-cause fix, not just a cache reset.

---

_Report scope: This session only — the go.mod replace→version fix. Does not cover the concurrent registry Name→ID refactor or other daemon-committed changes._

---

## Resolution (docs-health pass, 2026-08-08)

All actionable items in this report are resolved. Key resolutions:

- **go.mod replace directive removed (§A, §F.34):** `go-finding` v1.4.1 is a
  real published tag. No `replace` directive in `go.mod`. Confirmed in
  `go.mod` (`require github.com/larsartmann/go-finding v1.4.1`).
- **GOPRIVATE + VCS auth (§B.3, §F.5):** CI uses `GOPRIVATE` with
  `GITHUB_TOKEN`. Local dev uses `GOPRIVATE` + SSH `insteadOf`. Documented in
  README and CONTRIBUTING.
- **flake.lock pinned (§C.1-4):** Quality gate verified: `nix run .#lint`
  (0 issues), `nix run .#test-race` (clean), `nix flake check` (all passed).
- **ROADMAP Q1 resolved (§F.2):** go-finding v1.4.1 confirmed real. Moved to
  "Resolved questions" in ROADMAP.md.

Still open (routed to TODO_LIST.md):
- **CI auth risk (§B.1, §F.1):** `GITHUB_TOKEN` may not have cross-repo read
  access to `go-finding`. Needs end-to-end verification or PAT/SSH deploy key.
- **go.work (§B.2, §F.2):** No workspace file for local cross-repo dev.
- **go mod tidy CI check (§F.17):** Not yet added.
