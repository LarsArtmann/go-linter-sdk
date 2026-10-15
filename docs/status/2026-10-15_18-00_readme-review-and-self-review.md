# README Review + Self-Review — 2026-10-15 18:00

**Scope:** this session only. Task was "is our README.md superb?" — a claim-by-claim
review of `README.md`, fixes for what was wrong, and now a brutal self-review of
that work. NOT a project-wide audit; only things this session did or noticed.

**Verdict that stands:** the README was structurally strong but factually flawed —
7 defects found and fixed. The self-review below found 3 more process defects in
my own execution.

---

## Session summary

1. Read README.md fully (366 lines pre-edit).
2. Verified every claim against source: all API-table signatures vs
   `rule.go`/`registry.go`/`errors.go`; inline code samples vs the real API and
   `examples/minimal-linter`; `report.PrettyJSON()`/`ToSARIF()` vs `go doc`
   (they are `*Report` methods — README phrasing correct); `cmdguard/v4` module
   path vs sibling `go.mod`; go-output "16 formats" vs sibling README; all
   relative links; docguard green; no trailing whitespace.
3. Fixed 7 README defects:
   - Self-contradiction: Consumers/Status said converter deletion "not yet
     independently verified" while the top blockquote + `docs/ADOPTION.md` say
     verified 2026-10-11. Aligned with ADOPTION.md.
   - "8 recommended categories" (×2) — `rule.go:45-57` ships **11**.
   - Missing error-attribution API — added Errors subsection
     (`*RuleError`, `RuleErrors`, `ErrRuleFailed`, `NewRuleError`,
     `ErrMissingFields`) with a `ContinueOnError()` example.
   - "six LarsArtmann-internal repos" → "repositories" (3 importers are public).
   - Removed rot-prone hardcoded `v0.3.1` consumer pins (×2); now points at
     ADOPTION.md.
   - Status section no longer repeats the three example descriptions verbatim.
   - Minimal sample no longer swallows the `Run` error.
4. Fixed `AGENTS.md` "8 built-in constants" → 11.
5. During THIS self-review: fixed the same stale count in `FEATURES.md:32`
   (split brain the original run left behind) and cleaned changelog-style noise
   I had injected into the AGENTS.md line.

---

## Self-review — the questions, answered honestly

**1. What did you forget?**

- **The claim sweep stopped at two files.** "8 categories" lived in THREE living
  docs (README, AGENTS.md, FEATURES.md). I declared the count fixed after two.
  The grep sweep across all 11 living docs happened only now, under prompting.
- **The pkg.go.dev consequence.** pkg.go.dev freezes the README at the last tag
  (v0.4.0). The self-contradiction I fixed is still live on the project's
  primary landing page until a PATCH tag. I never surfaced this.
- **No CHANGELOG.md entry** for a user-visible accuracy fix — the project rule
  says CHANGELOG is the narrative of record.
- **I never confirmed no README-code-fence test existed** before editing code
  blocks. (Checked now: none exists — `examples_integration_test.go` only uses
  "README.md" as a fixture filename. Nothing could break. But I verified that
  after the fact, not before.)

**2. What is something that's stupid that we do anyway?**

- The doc-drift guard polices version literals but **no guard polices factual
  counts or adoption claims**. "8 categories" rotted silently through CI. Same
  class: adoption verdicts with dates that go stale without a tripwire.

**3. What could you have done better?**

- Sweep ALL living docs for each stale claim before declaring it fixed — the
  tooling (one grep) was already in my hand.
- Surface the release-policy implication (PATCH tag needed) in the original
  final report instead of only "docguard green".
- State silently-made tradeoff decisions: I chose to keep the redundant
  pkg.go.dev text link under the badge and never mentioned it.

**4. What could you still improve?** → sections (e)/(f).

**5. Did you lie to you?**
No. Every "verified" claim in the final report was actually verified against
source. But "7 defects found and fixed" was **incomplete** — true for README,
while a third doc still carried defect #2. Incomplete-but-true is still a
quality miss.

**6. How can we be less stupid?**
Codify the sweep: a guard that cross-checks counts/claims (see f/3, f/4) so the
next agent cannot stop at the file it happens to be editing.

**7. Ghost systems?**
None found. Everything documented in the README exists in code; nothing in code
contradicts the README post-fix. `RuleErrors`/`NewRuleError` were documented-but-
undocumented API (README gap) — now closed.

**8. Scope creep?**
No — stayed on README (+ its claim echoes). The FEATURES.md fix was the boundary
and it belonged.

**9. Did we remove something that was actually useful?**
Deduplication removed the Status section's verbatim repetition of the three
example descriptions — the descriptions still exist in Consumers, so no
information was lost. Verified by re-reading.

**10. Split brains?**

- Found & merged: the 8-vs-11 category count across README/AGENTS/FEATURES.
- Pre-existing, partially introduced by me: the adoption 3/2/1 verdict now
  appears in the README **twice** (top blockquote + Consumers) plus ADOPTION.md;
  "~4,700 LOC" (README) vs "~4,737 LOC" (ADOPTION) is the same number rounded
  two ways. Canonical home should be ADOPTION.md; README summarizes once.

**11. Tests?**
Docs-only change → suite not run; justified EXCEPT that I edited Go code blocks
inside README, and no mechanism exists to compile them. That's the gap worth
closing (f/1), and it also protects future sessions.

---

## a) FULLY DONE

| # | Item                                                        | Evidence                                                     |
| - | ----------------------------------------------------------- | ------------------------------------------------------------ |
| 1 | README claim-by-claim verification                          | every API symbol, sibling claim, link checked against source |
| 2 | 7 README defects fixed                                      | README.md diff (+42/−10 lines)                               |
| 3 | AGENTS.md category count fixed + noise cleaned              | AGENTS.md diff                                               |
| 4 | FEATURES.md category count fixed (this review)              | FEATURES.md:32                                               |
| 5 | docguard green after every doc touch                        | 4 clean runs                                                 |
| 6 | No README-fence test exists — confirmed nothing could break | `examples_integration_test.go` inspected                     |

## b) PARTIALLY DONE

| # | Item                           | What remains                                                                                             |
| - | ------------------------------ | -------------------------------------------------------------------------------------------------------- |
| 1 | Corrected README reaches users | pkg.go.dev still serves the v0.4.0 (self-contradicting) README until a PATCH tag                         |
| 2 | Claim-accuracy sweep           | done for the 3 stale patterns across all 11 living docs, only as a late addendum — not institutionalized |
| 3 | Honest change record           | CHANGELOG.md entry for the accuracy fix still unwritten                                                  |

## c) NOT STARTED

1. README code-fence compile guard (test/CI).
2. Category-count sync guard (count `Category*` constants vs "N recommended values" in docs).
3. Release decision + execution for the docs refresh (v0.4.1).
4. pkg.go.dev rendering check for `<details>`/mermaid fallbacks (assumed OK, never looked).
5. Adoption-claim staleness tripwire (dated "verified YYYY-MM-DD" claims).

## d) TOTALLY FUCKED UP

Nothing destructive, nothing reverted, no lies. Three process failures, owned:

1. **Declared victory with a live split brain** (FEATURES.md) — the exact defect
   class I was fixing, left in a third file.
2. **Injected changelog-style parenthetical into AGENTS.md** — violates the
   project's own "AGENTS.md as changelog" anti-pattern rule. Cleaned this review.
3. **Missed the release-visibility consequence entirely** — fixed the sales page
   nobody visiting pkg.go.dev can see yet, and didn't say so.

## e) WHAT WE SHOULD IMPROVE

1. **Guards for facts, not just versions.** docguard catches `vX.Y.Z` drift; the
   "8 categories" rot was invisible to CI. Counts and dated claims need the same
   treatment.
2. **README code is untested code.** Every fenced snippet should compile in CI;
   otherwise samples rot exactly like prose did.
3. **Single home for adoption evidence.** ADOPTION.md is canonical; the README
   should summarize once, not twice, and not carry its own rounded LOC numbers.
4. **Session discipline:** sweep all living docs per claim; surface
   "who can see this fix" (tag/render implications) in every docs-change report.
5. **Docguard's blind spot is structural:** it whitelists any repo tag, so
   `v0.3.1`-style consumer pins pass forever while going stale — prefer pointers
   (which is the fix I applied, but the guard will not enforce it for the next
   person).

## f) Next tasks (impact-sorted; honest count 28, not padded to 50)

**Release & visibility**

1. Decide on v0.4.1 PATCH tag to push the corrected README to pkg.go.dev (wait
   for green CI on the exact commit first — recorded rule).
2. Write the CHANGELOG.md entry for the README accuracy fixes.
3. After tagging: verify pkg.go.dev @latest renders the new README; check the
   "Imported by" count (post-release habit).
4. Inspect how pkg.go.dev actually renders the `<details>` ASCII fallback and
   mermaid blocks; adjust if raw HTML leaks.

**Guards (kill the rot class)**
5. Add `readme_compile_test.go`: extract ```go fences from README, compile each
(wrap with imports/`main` as needed; skip blocks marked as fragments).
6. Extend it or docguard: count `Category*` constants in `rule.go` and assert
README/FEATURES/AGENTS all say that number.
7. Consider a dated-claim tripwire: grep living docs for "verified
YYYY-MM-DD" older than ~90d → CI warn (not fail).
8. Add a relative-link checker for README (anchors like `#api`, file targets) —
cheap Go test or lychee.

**README content polish**
9. De-duplicate the adoption 3/2/1 breakdown: keep ONE summary (top
blockquote), make Consumers point at ADOPTION.md.
10. Unify "~4,700 LOC" vs "~4,737 LOC" (pick one rounding; ADOPTION is canonical).
11. Decide the redundant pkg.go.dev text link under the badge (keep/remove).
12. Consider a short "Changelog" pointer section or badge linking CHANGELOG.md.
13. Re-check "98.7% coverage" claim freshness at next coverage run (it is a
dated snapshot; the number will drift).

**Docs-health (bounded, next session)**
14. Run docs-health VERIFY over FEATURES.md/ROADMAP.md/TODO_LIST.md for other
count/claim rot (I only swept the 3 patterns I knew about).
15. Confirm `docs/migration-guide.md` still matches the rewritten README sections.
16. Harvest this report's section (f) into TODO_LIST/ROADMAP per docs-health
HARVEST (do not entomb).

**Small code-adjacent**
17. If the fence-compile guard (f/1) proves annoying for fragment snippets,
standardize a marker convention (e.g. `// fragment`) documented in
CONTRIBUTING.md.
18. Add `RuleErrors`/`Errors` subsection example to `example_test.go` so the
README Errors snippet mirrors a tested Example function.

## g) Questions I cannot answer myself

1. **Tag v0.4.1 now** to refresh pkg.go.dev's frozen README (docs-only PATCH,
   consistent with the recorded versioning rule), or batch it with the next
   code change? (Release timing is your call.)
2. **Adoption numbers in the README:** keep the concrete 3/2/1 breakdown and LOC
   figures in two places, or compress to one summary + ADOPTION.md pointer
   (my recommendation: compress — less to rot)?
3. **Do you want the two new CI guards** (README-fence compiler, category-count
   sync) accepting slightly stricter docs workflow — snippets must stay
   compilable and counts must match code?

---

_Point-in-time snapshot. Written 2026-10-15 18:00 after session-close; the
auto-commit daemon picks up the file (no manual commit per harness contract)._
