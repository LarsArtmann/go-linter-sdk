# Fleet Library Audit — Reusable Template

> Extracted 2026-11-19 from the seven-library fleet review plus the
> go-error-family follow-up (plan task M24 / F24.1). Dry-run validated
> against `go-error-family` (the 8th audit was written FROM this template
> shape after the fact). Use this to start the next repo's audit compliant:
> every requirement below is one the first run had to learn or was asked for.

## Inputs (collect before dispatching anything)

1. **Fleet list**: every first-party library the repo depends on OR should
   have considered — direct (`go.mod` require) AND transitive-but-first-party
   (run `go mod why <lib>`; go-error-family was missed exactly because it
   only rode in transitively).
2. **Pin sources**: `go.mod` for the consumer's version; `git -C <lib-repo>
   describe --tags` for the library's latest local tag; note if the local
   checkout is ahead of the published tag.
3. **Read-only module cache** as the audit source of truth:
   `$(go env GOMODCACHE)/github.com/<org>/<lib>@<pin>/` — identical to the
   tag, no network needed.

## Report contract (every report must carry)

- **Header block:** reviewed date + who; library import path; sub-modules
  considered (multi-module repos: enumerate and say why each is in/out);
  version reviewed (consumer pin AND library-side tag); verdict; score (or
  "no score" + reason).
- **Capability inventory table:** one row per exported capability cluster,
  with `file:line` evidence in the LIBRARY at the pinned version (read from
  the module cache, cite paths).
- **"What <consumer> uses today" table:** consumer-side `file:line` for every
  usage; count references with a replayable grep.
- **Per-capability gap grades** (choose one):
  - **Fully Leveraged** — used as intended
  - **Partially Used** — used, with documented shallowness
  - **Missed Opportunity** — capability aimed at this consumer, unused
  - **Misused** — used against the grain (rare; name the friction)
  - **Not Applicable** — with a REASON per row (layering, domain mismatch)
- **Derived score:** Fully=1, Partial=0.5, Missed=0, N/A excluded;
  `score = round(100 × points / applicable rows)`. Show the arithmetic.
- **Ranked opportunities:** impact × ease, each with its routing marker
  (TODO_LIST row / ROADMAP line / none-informational).
- **Flip trigger:** every NOT-ADOPTED verdict names the condition that would
  change it (version milestone, new consumer shape, docs landing).
- **Fleet argument:** one paragraph on the owner's collective-improvement
  constraint — what adopting/not adopting means for the other repos.
- **Verify block:** replayable shell commands that re-derive the headline
  claims (grep counts, `sed -n` citations, tag listing).

## Orchestration (sub-agent fan-out)

- Cap parallel write sub-agents at 3 per provider; more hit rate limits.
- Orchestrator writes the self-audit and the dashboard; sub-agents write
  per-library reports.
- QC: re-verify the FIRST report of each group against source before
  trusting the rest (check one grep, one file:line citation, one score's
  arithmetic).
- Route findings into TODO_LIST/ROADMAP at write time, not in a later
  harvesting pass — the routing marker in each report must point at the live
  tracker row from day one.
- Pre-flight providers before dispatching (this run lost two dispatches to a
  depleted prepay and a provider max_tokens bug).

## Dashboard (`index.html`)

- One stat-card row (adopted count, missed-channel count, average score),
  one verdict table row per library, issue cards for cross-cutting findings,
  a dependency-tree section, a follow-ups section, and a verify appendix.
- Render-gate the page with headless chromium
  (`--headless=new --screenshot`), and vision-verify the PNGs with a
  capable model (record which model; this run's gemini/zai attempts failed —
  see the session log for the workaround).

## Point-in-time discipline

- Stamp every version claim; the reports are snapshots, TODO_LIST and
  ROADMAP are the living truth.
- Re-review trigger per report: "next minor bump after <pin>" or a named
  TODO landing. Footers carry it (F22.1).
- After the audit: `git fetch --tags` each library repo and annotate any
  "latest tag" claim that fetch proves stale (this run: none stale).
