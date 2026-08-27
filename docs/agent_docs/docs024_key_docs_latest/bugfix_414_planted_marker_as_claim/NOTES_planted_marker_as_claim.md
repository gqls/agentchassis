# NOTES — bugs_open/414 (append-only, newest at the bottom)

## 2026-08-27 — session 1

### (a) Ownership: filed and handed on, not being worked

`scripts/who-owns.py 414` said OWNED by `portfolio_positioning` (142 commits/14d). Reading further
is what changed the answer: the bug is a **"Late addition (same evening)"** in that lane's handoff,
whose §1 first task is sitemaps, and the lane's session was idle. It filed, recorded the debt, and
moved on. So: resumed here, contributing INTO the bug file rather than forking an account.

### (b) The bug was still live, and the file's own status line was wrong

Re-measured rather than inherited: 2 + 1 served (control phrase "PRA handbook" = 0), 3 components
carrying it in **both** `content_data` and `rendered_html`, none locked.

Then the finding that changed the shape of the work. Running the file's §Population census over
**every** aspect instead of the one that had been edited: lendzy's **current `strategy`** row
(`96eaff0b`, `domain-strategist`, 2026-08-12) still read *"The acceptance marker 'checked against the
FCA handbook, rule by rule' should appear in the site's written copy…"*. The 08-26 fix had stripped
`content_direction` only. **"Regeneration can no longer re-plant the phrase" was false for ten
days.** `strategy` is read by `build-site-planner` and `webdesign-agent` (measured over live
`agent_definitions`), so the instruction was live in a surface an agent reads.

### (c) My own misstep, in the same query — `_` is a SQL wildcard

I first reported that the **key** `acceptance_marker` had propagated, from a column I had computed
as `data::text ILIKE '%acceptance_marker%'`. That predicate matches the PROSE "acceptance marker"
too. Escaped, the key form exists in exactly one superseded row. What caught it: grepping the pretty
JSON for `acceptance_marker` and getting nothing while the SQL said true — **two instruments
disagreeing, and I had believed the one I wrote on purpose.** Logged in `WRONG_CALLS.md`. It mattered
beyond tidiness: "the key propagated" would have pointed the fix at a key-strip, which cannot touch
prose. The truth — *an agent paraphrased the instruction into another aspect* — is what makes a
text-scanning detector the answer.

### (d) The history says the spec was re-planting it every time

`page_component_history`: **14** archived rows carry the phrase — `about` ×3 slots on 08-11
(including a `differentiators` slot that no longer exists) and the guide's `article-body` across
**4 versions, 08-15 → 08-24**. So the guide re-emitted the phrase on every regeneration while the
spec mandated it. Two consequences: the population was never really 3, and a framework rewrite
against a *clean* spec is evidenced (not assumed) to produce copy without it — the homepage instance
vanished on its own when index regenerated on 08-24.

### (e) The audit item was worse than the bug file said

`052d01b0`'s `current_value` is a **fourth** instance of the claim ("Our guides are checked against
the FCA handbook, rule by rule…"), attributed to `index` — a page that no longer carries it. Its
`suggestion` asked for a "How we verify our guides" methodology section, and its `acceptance_test`
was satisfiable innocuously (name any CONC rule), so a handler could pass the test **while** building
the methodology. Rejected under a guard that aborts unless the row is still `needs_human_review`.

### (f) Two things I nearly built that the measurements killed

1. **A generic claims scan over `site_specs`.** Measured: **21** hits over 522 current rows,
   effectively all false. Fifteen are the estate's OWN honesty instructions ("Never invent a person,
   company, scheme…") matching the never-invents pattern — and the negation guard cannot save them,
   because the match *starts* at "never" and the guard only looks backwards. Worse, `evidence_base`
   rows store each site's `banned_claims` **as data**, quoting the sentences they forbid: a generic
   spec scan convicts every site's own immune system, daily, for ever. `brief-negation-check`'s own
   header records this census as tried and **withdrawn within hours** on 2026-08-19 — I was about to
   re-run it.
2. **A content guard on `WriteSiteSpecAction`.** The plant arrived as a **manual** row
   (`source='manual'`, `created_by='cqls'`) and never passed through that action. A guard there
   covers the agent door only, and the agent door was hop TWO.

### (g) The calibration that decided the whole design

The three served sentences are not one shape but two, and the estate already had a family for each:

| component | shape | caught by |
|---|---|---|
| `content-block-about` "**Everything** on this site **is checked**…" | completeness | new indefinite-subject pattern |
| `article-body` "**Every figure** and every rule reference … **is checked**…" | completeness | window 30 → 60 |
| `hero-about` "…, **checked against the FCA handbook, rule by rule**" | diligence | practice-family P6 |

Over 2,405 live components: the existing completeness pattern fires **0** times fleet-wide at window
30 (it is inert today), **1** at 60, and the new entry **1** — each on exactly the sentence it was
written for. For P6, each half alone is unshippable: idiom alone **22**, verb+`against`+rulebook
alone **13** (including lendzy's own *correct* imperative "Check your loan against the FCA rules");
both together **3**, the planted ones, 0 of the other 2,402. **The conjunction is the design, not the
words.**

### (h) The skeptic pass earned its cost, and I was wrong twice

Ran an adversarial review of my own proposal before writing code. It killed the generic spec scan
(f1), and it caught a **stale justification**: I had argued for `brief-negation-check` over a
discovery check *because* discovery checks ride the improvement sweep, dead 2026-05-02 → 08-25. That
is false — discovery checks ride `site-discovery-rotation-*`, and `unverified_claims` has filed 17
items in 7 days. Same decision, honest reason (shape, not cadence). Logged in `WRONG_CALLS.md`,
because a reviewer told the wrong reason cannot re-derive the right one.

### (i) A test that could never fail, found by writing one that expected PRESENCE

`validate_page_content_fleetwide_claims_test.go`'s helper read
`out["issues"].([]ValidationIssue)`; the action returns `[]map[string]string`. So the helper always
returned nil and every false-positive assertion in that file — green since `bugs_open/104` — was
asserting that an empty list contained nothing. The blocker tests were unaffected (they assert on the
error, which is the real mechanism), which is exactly why nobody noticed. **A test expecting absence
cannot detect an instrument that reports absence.** Fixed the helper; all its tests still pass, so
nothing was hiding behind it.

### (j) The ordering I got wrong in the first plan, and the measurement that inverted it

I had written "pattern first, it protects against the Retry". The skeptic traced the enforcement
points: with the pattern live and the phrase still in `content_data`, a rerender regenerates HTML
carrying it, the **persistence floor refuses the save**, and the OLD `rendered_html` keeps serving.
The item lands `unresolved` and nothing a visitor sees changes — a stranded repair that reads like a
working gate. Repair first; the Go is inert until a roll anyway.

### (k) Mutation proofs run, not assumed

Six guards, each turned red by a deliberate break and green again on restore: the window widening,
the indefinite-subject pattern, P6's rulebook list, P6's order-B alternation, the attestation
exemption, the `evidence_base` exclusion, and the bare-aspect surface fallback. The strip's SQL guard
was induced too (corrupted the expected tail; it aborted with nothing changed) — a guard I had not
watched fire would have been a comment.

### (l) The dry run, and the export that lied twice

`claimscan` over the complete corpus: **3 BANNED** (2 mine, both lendzy; 1 pre-existing on
webdesign.co.uk) and **12 PRACTICE** (3 mine, all lendzy; 9 pre-existing). Zero false positives.
Getting the corpus out took three attempts: a single stream dropped **302 of 2,585** rows with an
"unexpected EOF" *after* printing plausible output; a per-domain loop hit the documented
`kubectl exec -i` stdin-eater and processed one domain. What works is writing inside the pod and
`kubectl cp`-ing the file out — with the row count agreeing three ways. **The row-count control is
the only reason I did not ship on an 88% corpus that scanned clean.**

### (m) State at the end of session 1

`fc588e445` committed (9 files, no passengers); council `f4c144ad` submitted and running
(`review_editquality` at 09:1xZ); HEAD verified — the only failing test in `datahelpers` is
**pre-existing**, from `component_hierarchy_walk.go:397` at commit `bc8167100` (another lane, 08-26)
hand-spelling the tombstone predicate. Spec sources clean fleet-wide (0 rows, any aspect, any site).
Audit item rejected. The two repair items are working through a **fleet-wide `resolve_links` /
`spawn_content_writer` flake** — the `about` item lost attempt 1 to `CHILD_ORCHESTRATION_FAILED` and
is retrying; several other lanes' page builds are failing the same way in the same window, so it is
not the payload.
