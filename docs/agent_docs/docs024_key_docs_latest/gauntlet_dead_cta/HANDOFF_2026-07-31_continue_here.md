# HANDOFF — gauntlet_dead_cta / vonc6 thread, 2026-07-31

**Written because this session's context ran long, not because the work stalled.**
Supersedes `HANDOFF_2026-07-29_continue_here.md` for anything below; that file is
still the reference for the rail (§5), the island landmines (§4) and the H ruling
(§7). Handoffs A–D of 2026-07-30 are separate startable units and are NOT
superseded — B is already TAKEN by another thread.

## 0. State in one paragraph

The visitor-identity fix is **live and proven in production**. The
content-duplication checker and its deterministic handler are **built, seeded,
council-reviewed once, and deliberately not enabled**. The chassis roll they were
waiting for has landed and is verified. Nothing is half-applied. The single next
action is a decision about enabling the check, and it needs one other lane's
agreement first.

## 1. DONE and verified this session

| what | state | proof |
|---|---|---|
| tools-api client identity (`bugs_open/139`) | **LIVE + PROVEN + council APPROVED** | island `v1.0.1207` swapped 08:37Z; see §2 |
| `check_content_duplication` + `remove_duplicate_page_sections` + `deduplicate-sections` | built, seeded, **INERT** | commit `feat(151 cand 3)`; chassis `v1.0.1214` pod-verified, §3 |
| Shared normaliser `datahelpers.NormaliseSectionText` | live | replaced two copies + a promised drift test |
| og:image on the gauntlet page | **FIXED** (was listed as a 404 to fix) | HTTP 200, `image/png`, 1200×630 exactly |
| 4 handoffs A–D + `dedup_census.py` + `provocation_visibility.py` | committed | `f599ed87c` |
| `SUMMARY_2026-07-31` | written | the authority dilemma as a worked example |
| Register: `CQ-018` added, `PUB-002` corrected | committed | httpguard's "called by NOTHING" is now false |

## 2. The identity fix — what makes it PROVEN rather than deployed

Council `e053fac4-eeaf-431e-aa88-817c4107476e` → round 1 REVISE (prior-art
evidence objection), round 2 **APPROVED** with **no code change**. The lesson is
reusable: the objection asked what a landmine note said, so the answer was the
note's full text plus a source line, not a rewrite.

**A real `/round` through Cloudflare stored the hash predicted BEFORE the request
was sent** — `9e464fe9fca925b099a25141f40afad5`, being
`sha256(<this machine's egress address per cdn-cgi/trace>)[:16]`. So the column did
not merely change; it changed to the expected value. `count(DISTINCT
client_ip_hash)` on the island's `gauntlet_rounds` went **1 → 2**.

That settles what `139` marked `[INFERRED]`: **`CF-Connecting-IP` reaches the app
process.** The acceptance check passed from ONE machine legitimately, because the
before-state was itself a constant — the 95 prior rows are the second network.

Still open, deliberately: `139` is left in `/bugs_open/` because `who-owns.py`
names another lane and **the close is theirs to call**; their three probe rows are
left in place because they are now the before-state of the measurement.

**Landmine earned:** when comparing against a stored digest, **read the hashing
function's width first.** I queried the full `sha256('172.18.0.1')`, got 0 rows,
and briefly read that as "the constant is not the docker gateway". `hashIP`
truncates to `sum[:16]` (`round.go:33-37`). Their figure was right; my check was
the wrong shape.

## 3. The duplication checker — everything except the last step

**Design, which is an owner ruling of 2026-07-31 and must not be quietly widened:**

```
IN-REMIT  same page, content-IDENTICAL  -> content_duplication item
                                        -> deduplicate-sections
                                        -> remove_duplicate_page_sections
                                           (delete later rows, keep earliest,
                                            renumber, queue a rerender)
RESIDUE   near-duplicate / cross-page   -> ONE capability_gap, NO handler,
                                           do_not_auto_rewrite: true
```

Rationale is in `SUMMARY_2026-07-31`; the short version is that a similarity score
says nothing about meaning, and `bugs_open/149` C1 is the witnessed cost of letting
an LLM rewrite copy unguarded.

#### Why the line sits exactly there — read this before "improving" the check

The remit boundary is not a limitation of the implementation. It is the whole design,
and it is drawn at **the only place where a repair is provably safe**: two rows on one
page whose normalised text is byte-identical. There, "which one do I keep?" has a
correct answer that needs no judgement — keep the earliest, because the duplicate adds
nothing a reader could want. Every widening below removes that property, so each one
is refused **on the same ground**, not on three different grounds:

| the tempting widening | why it is refused |
|---|---|
| "0.92 similar is obviously the same — dedupe it too" | Now something must decide which of two *different* texts survives, and that is an editorial judgement. `bugs_open/149` C1 is what that costs when an LLM makes it unguarded. A threshold does not make the judgement safe, it hides who is making it. |
| "the same fact on two pages is duplication — dedupe across pages" | Two pages restating one fact is usually **correct** — a landing page and a case study *should* both cite the headline number. Cross-page identity is a content-strategy question, and it has no page-local answer at all. |
| "a unique index on `(page_id, slot_name)` would prevent all of this" | Measured: it **breaks 11 legitimate pages**. 10 of the current 11 groups are real repeated slots with genuinely differing content (§3 re-measurement). The schema cannot express the rule because the rule is about *content*, not *shape*. |
| "the residue item has no handler — wire one up" | The missing handler **is the finding.** `do_not_auto_rewrite: true` is a deliberate statement that this class needs a human, recorded as one `capability_gap` so it is visible rather than silently unhandled. Giving it a handler converts an honest gap into an unsupervised rewriter. |

**The test for any future change: does it preserve "the repair needs no judgement"?**
If a human would have to decide which text survives, it is residue, and residue gets a
`capability_gap` — never a handler. Widening this check is an **owner decision**
(2026-07-31), not a refinement a thread makes while passing through.

**Roll verified.** Chassis `v1.0.1214`, both replicas:
`remove_duplicate_page_sections` **4**, `content_duplication` **6**, nonsense
control **0**. So the action is in the running binary and the chain is complete.

**Council round 1 = REVISE (`da3f2d9b-ae6f-492d-ad3b-748323b66367`), and it caught
two REAL defects** — both now fixed in `269_*.sql` **and applied to the live row**
by UPDATE (the seed is `ON CONFLICT DO NOTHING`, so re-running it would not have):

1. `page_id` was in the item spec only → the queued `page_rerender` would have a
   **NULL `page_id` column**. `create_work_item` takes it as an optional input that
   populates the column (`create_work_item_action.go:247-249`).
2. `item_key_prefix` alone yields `<prefix>_<domain>` — **site-wide**. Two pages
   deduplicated close together would collide on `idx_swi_dedup` and **one rerender
   would be silently lost**. Now sets `item_key_suffix_field: input_data.page_id`,
   which is a hard error when unresolved rather than a fallback to the colliding key.

**Round 2 has NOT been submitted.** That is the first concrete task below.

### The reviewer also re-measured my population claim, and it is now stale

`bug_historian` re-ran the fleet query and got **`total_groups: 11,
content_identical: 0`**. My submission quoted 17 groups / 11 legitimate / 6
content-identical-all-on-vonc-about. That was true when measured and is no longer:
**another session hand-fixed vonc's about page** (`d7f69e1c0`, 6 DELETEs +
renumber), which was the entire content-identical population.

⇒ **The in-remit half currently has nothing to detect fleet-wide.** It is a guard
against recurrence, not a backlog-clearer. Say that plainly in round 2 rather than
requoting the old figure — and re-measure before quoting anything, because that is
the third stale-figure correction in this lane this week.

### Re-measured 2026-07-31 ~10:20Z — confirmed, and it now has a stronger proof

Independently re-derived from the schema rather than by re-running the reviewer's
SQL (agreement between two runs of the same query proves nothing about the query).
**`total_groups 11 / content_identical 0 / legitimately_repeated 10 /
has_null_content 1`.** The query is now in `RUNBOOK` §16 — it was missing, which is
how the figure went stale unnoticed in the first place.

**The 11th group is the interesting one, and it is the landmine shape:**
`finetuning.uk /our-position-on-ai.html`, slot `generic-text-block`, 2 rows, both
with `content_data IS NULL` ⇒ `count(DISTINCT md5(...)) = 0`. A naive identity test
reads 0 distinct values as "all the same" and this is **the one group in the fleet
where a false positive would delete a live row.**

**Both halves already exclude it, for a reason independent of NULL-handling:**
`COALESCE(content_data::text,'{}')` → `NormaliseSectionText` → empty string → caught
by the `len(s.Text) < 80` floor, applied identically in
`check_content_duplication.go:238` (detect) and
`remove_duplicate_page_sections_action.go:144` (repair). Same threshold, same shared
normaliser, both halves.

⇒ **So, precisely: enabling the check today would file ZERO items and delete ZERO
rows.** That is the sharpest available statement of "built but inert" — not merely
"not switched on", but *verified to be a no-op against the current fleet*. Two
consequences, both load-bearing:
- It **strengthens** the case for round 2: the design's safety is no longer an
  argument, it is a measurement over the real population including its one
  adversarial case.
- It **weakens** any argument for enabling it in a hurry. There is no backlog to
  clear, so the only thing the switch buys today is recurrence protection — which is
  worth having, and is worth having *after* the owning lane answers, not before.
  Nothing degrades while it waits.

⚠ **Do not carry the `11 / 0` figure forward without re-running §16.** It was true
at 10:20Z. It was a different number 24 hours earlier, and one hand-fix by any
session moves it again.

## 4. NEXT ACTIONS, in order

1. **Resubmit the checker, round 2** with `RESUBMIT_CORR=da3f2d9b-ae6f-492d-ad3b-748323b66367`.
   Content: the two seed fixes (done, live), the corrected population figure
   (**re-run §16 first — it was 11/0 at 10:20Z**, guard-not-backlog), the
   NULL-group no-op proof from §3 (the strongest single item in the submission: the
   one fleet group that could trigger a wrongful delete is excluded by both halves at
   the same threshold), and a note that the sketch now shows the enqueue the
   rationale promised. Working file:
   `<scratchpad>/council_dedup.json` is gone with the session — rebuild from the
   commit message of `feat(151 cand 3)`, which carries the full reasoning.
2. **DO NOT enable the check yet.** Enabling means adding `content_duplication` to a
   discovery agent's check list. Two reasons to wait: the owner has not asked for it,
   and **the first site it runs against starts deleting rows**. `bugs_open/151` is
   owned by the `brochure_component_library` lane (113 commits/14d) and
   fundamentallyai is theirs — a CONTRIB is already in their directory
   (`CONTRIB_2026-07-31_151_candidate_3_is_built.md`) offering them the checker
   outright. **Get their answer first.**
3. **`bugs_open/154`** — `tool-improver` dies at `load_tool` on `tool-auditor`-raised
   items (`input_data.component_id` resolves to nil, and counterintuitively it is the
   items WHOSE COLUMN IS SET that fail). Unowned, mechanism marked `[INFERRED]`;
   read the dispatch `input_mapping` before acting. Consequence: `bugs_closed/010`'s
   convergence guard has STILL never reached cycle 1.
4. **`bugs_open/150`** — the improvement loop reports "No issues found — site is
   clean" straight after promoting findings, and skips its own closing rerender.
   Unowned. Anyone firing the sweep to clear a backlog is told the opposite of what
   happened.
5. **Handoffs A / C / D** of 2026-07-30 are startable. **B is TAKEN** (lane
   `docs024_key_docs_latest/provocation_pipeline/`) — do not start a second thread.
   That thread also **corrected me**: the provocations archive page is NOT broken;
   my 1,293-char figure came from a probe looking for *today's* headline, which
   correctly is not on that page.

## 5. Facts: the answer is "nothing to fix, and adding is a real choice"

Owner ruled SQL-verifiable counts only. **vonc's existing four all verify exactly**
(8/8 archetypes, 3/3 tools, 2/2 guides, 18/18 pages), each carrying its own query
and tolerance, re-checked by the enabled daily `evidence-freshness` task.

My earlier flag that "3 tools looks low" was **wrong**: the facts define themselves
by `page_type`, and I had counted name prefixes. Adding facts is therefore a
deliberate choice about what vonc may claim — not a repair — and every fact added
widens the surface for `bugs_open/149` C1, which still has no claims gate. Put a
proposal to the owner per site; do not bulk-add to make a checker's fact half work.

## 6. The one thing to carry out of this session

Four — arguably six — of this week's wrong calls were **one shape**: comparing
against a ruler I invented instead of the one the data declares. Counting pages by
name prefix when the fact says `page_type`. Grepping for "Monte Carlo" when the
technique need not name itself. Calling a pipeline idle from two observations four
minutes apart. Building a URL instead of reading `pages.url`. Comparing a full
digest against a truncated column. Reading a low visible-char count as breakage.

Three were caught only because someone re-checked a number already reported, and
one reached the owner as an urgent finding with a recommended action. **Re-ground
every figure before ESCALATING, not before acting on the reply** — urgency is
exactly when the check gets skipped and exactly when being wrong costs most.
