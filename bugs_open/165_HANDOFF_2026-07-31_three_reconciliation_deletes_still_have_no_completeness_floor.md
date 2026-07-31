# 165 — three more reconciliation deletes have no completeness floor, and one of them is the table that has actually lost content

**Filed 2026-07-31** by the `bugfix_135_prune_floor` lane, **at the explicit
direction of the council gate's `bug_historian` seat** (corr
`14239fa4-552f-4821-abaf-ea15ccee4ea5`, round 2, severity HIGH). Status: **OPEN —
site A DONE, B and C outstanding.** Latent — no failure to point at *today*, but
two of the named precedents are failures that already happened.

> **UPDATE 2026-07-31 evening — site A is fixed and committed (`ecf738002`).**
> `save_page_sections_action.go` now carries a measured completeness floor
> (`save_sections_prune_floor.go`, 17 tests), reusing `evaluatePruneFloor`
> unchanged. Council `a54172b6-9756-4abc-a9e0-f173ad4de779`, committed with
> `Council-Submitted:` before the verdict landed. **Not yet closed, and it must
> not be until both branches are induced live** — the floor is inert on healthy
> input by design, so the roll alone proves nothing. Lane docs:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_165_reconciliation_deletes/`.
>
> **Two things the fixing session learnt that change the advice below.**
>
> 1. **The per-`slot_name` cohort this file suggests for A does not work, and the
>    measurement is decisive:** 998 of 1,009 `(page_id, slot_name)` groups hold
>    exactly ONE row. Every per-slot cohort would be 1 stored, so a legitimate
>    single-section removal scores it 0% and refuses — 89 real shrinkages in 4.5
>    months, each blocked. What shipped instead is two page-level cohorts in
>    different units: rows (what the save inserts vs what the DELETE removes) and
>    plan (vs `pages.sections`, which seven *other* actions write and this one
>    never does). The plan cohort is what breaks the **ratchet** — once a writer
>    has cut a page to two rows, the row cohort reads 2/2 = 100% for ever.
> 2. **Exclude actively-locked rows from any plan-side denominator.** Counting a
>    slot a lock makes unwritable refuses a *perfect* rebuild of the pages a human
>    curated (`idea.uk/index.html`: 6 planned, 4 locked, so a flawless rebuild
>    scores 2/6). ~~Applies to B as much as to A — `site_nav_items` carries the same
>    lock columns.~~
>
> > **CORRECTED 2026-07-31 by site B's converter — `site_nav_items` has NO lock
> > columns, and site B does not inherit this at all.** The live schema for
> > `site_nav_items` is exactly `id, site_id, group_id, parent_item_id, label, url,
> > page_id, item_type, position, status, metadata, created_at, updated_at`, and a
> > search for `%lock%`/`%owned%`/`%writable%` across `site_nav_items`,
> > `site_nav_groups` and `page_components` matches **only** `page_components`
> > (`lock_type`, `locked_at`, `locked_by`, `lock_expires_at`). Caught by reading
> > `\d site_nav_items` while designing B's denominator. The same claim was in
> > CTXA-025's register entry and is corrected there too. Worth noting how it got
> > written: it is a plausible inheritance asserted about a sibling table, in an
> > update block whose every *other* figure was measured — which is exactly the
> > asymmetry CLAUDE.md's "mark the UNVERIFIED ones too" exists for.

## Why this exists as its own case

`bugs_closed/135` fixed the identical defect for `code_symbols`: a run that
re-upserts what it saw and then deletes the remainder, with **nothing checking
that it saw the corpus**. The fix deliberately generalised the *rule*
(`platform/orchestration/actions/prune_floor.go`, register **CTXA-025**) and
deliberately did **not** convert the siblings. The seat's objection is that
stopping there is itself the platform's most repeated pattern:

> "This plan builds a rigorous, tested, fail-closed completeness guard for
> `code_symbols` and then leaves the mechanistically identical hazard live and
> generic on the table that has actually lost real content before. This is
> precisely pattern (c) — one call site gets the rigorous fix, the sibling(s)
> stay exploitable — and the index has a title for exactly this shape:
> `021_HANDOFF_2026-07-18_durable_write_guard_covers_one_path_only`, plus the
> transferable pattern 'One call site of a shared judgement gets the rigorous
> fix; the sibling stays heuristic' (016b §9)."

It is right, and the answer is a filed case rather than a wider patch — see
"Why not in 135's commit" below.

## The three sites

From `grep -rn 'DELETE FROM' --include=*.go platform/ internal/ pkg/ services/`
(15 statements; these are the four that delete-what-this-run-did-not-see, one of
which is now guarded):

| # | site | delete | stakes |
|---|---|---|---|
| A | `save_page_sections_action.go:532` | `DELETE FROM page_components WHERE page_id = $1 AND <agent-writable>` | **HIGH — real content, lost before. DONE: guarded, live on v1.0.1223, BOTH BRANCHES INDUCED 2026-07-31** |
| B | `populate_nav_tables_action.go:147,150` | `DELETE FROM site_nav_items WHERE site_id = $1` then `site_nav_groups` | medium — regeneratable, but a partial nav is served. **CODE DONE 2026-07-31 (`983e4b0a2`), NOT yet live, NOT yet induced** |
| C | `site_db_actions.go:1474` | `DELETE FROM link_registry WHERE source_page_id = $1` | medium — regeneratable. **CODE DONE 2026-07-31 (`983e4b0a2`), NOT yet live; cannot be induced — see below** |
| — | `code_symbols_actions.go` | guarded 2026-07-31 (`bugs_closed/135`) | — |

**A is the one that matters.** The seat's own lineage for it: 016b §9 cases 1 and
2 (an interactive tool/game silently deleted by a routine content rebuild; the A\*
pathfinding game destroyed by a DELETE+INSERT rebuild, **recurring independently
on a second site**), plus the case files `001_HANDOFF_replan_clobbers_built_pages`,
`037_HANDOFF_needs_rebuild_pages_are_unprotected_by_the_replan_guard`,
`038_HANDOFF_replan_rebuilds_every_deployed_page_and_regenerates_its_content`, and
`bugs_closed/058_HANDOFF_rebuild_path_does_not_honour_page_component_locks`.

Note A already has *a* guard — `pageComponentAgentWritableSQL` keeps the delete
off human-locked and non-agent-writable rows, and 058 added lock honouring. That
is an **authority** guard ("may I delete this row?"). It is not a **completeness**
guard ("did this run see enough to be deleting anything?"), and a writer that
produced two sections instead of twelve passes the authority guard perfectly.

## The mechanism, restated for these tables

Identical to 135: the run's population comes from something that can silently
return less than the whole (an LLM writer that returned a short section list, a
partial plan read, a nav build over a partial page set). A short-but-nonzero
result raises no error, so the delete removes everything the short run did not
re-write, and the outcome is not a broken row — it is *absence*, which reads as
"there was never anything there".

## Fix shape (do NOT copy 135's cohorts)

The rule is reusable; the **cohorts are not**. `evaluatePruneFloor(floor,
[]pruneCohort{...})` takes `(label, confirmed, stored)` triples and returns a
verdict plus a refusal sentence naming its own remedy. What each site must supply
is its own answer to "what can I lose independently, and what is my signal in a
*different unit*":

- **A** — plausibly one cohort per `slot_name`, plus a whole-page signal (sections
  confirmed vs sections stored). **Measure before choosing**: a page legitimately
  losing half its sections is a real edit, and a floor that fires on it is worse
  than useless.
- **B** — per nav group, plus distinct nav items. A site whose nav genuinely halves
  is a real event; the floor must be resolvable exactly as 135's is.
- **C** — per link kind if there is one, plus distinct targets.

**Fail closed on an unmeasurable floor** (135 does), **report a `*_status` rather
than a bare count** (a `deleted: 0` is ambiguous between "nothing to delete" and
"we refused"), and **give the refusal a durable surface** — for a site-scoped
writer `site_work_items` IS available, unlike 135's repo-wide case.

## Why this was not done in 135's commit

Stated plainly so the next thread does not re-litigate it:

1. **Scope.** CLAUDE.md's platform-seam ruling and `bugs_closed/124`'s REJECTED
   verdict both say a shared mechanism arriving inside a bug patch is
   architecture-scope. Converting three more shared writers in a fix for a fourth
   is precisely that, at three times the blast radius.
2. **Live territory.** On 2026-07-31 `save_page_sections_action.go` was actively
   being edited by another lane (a claims guard wired into it 07-30), and the nav
   tables by the `bugfix_149_nav_membership` lane. Landing a completeness floor
   underneath either would have been a same-file collision on a shared tree.
3. **The cohorts are unmeasured.** 135's were chosen *after* reading the live
   distribution (4,992 rows, five kinds, 592 paths). Guessing three more sets from
   the shape of the SQL is how a guard that fires on legitimate edits gets built,
   and a guard that cries wolf gets deleted by the first person it blocks.

None of that argues the work should not happen — only that it is its own work.

## How to verify, when someone does it

> **SITE A IS DONE — both branches induced in production 2026-07-31 on
> v1.0.1223.** Refusal: plan inflated 7→20, orchestration went **FAILED @
> save_sections**, `planned sections 35% (7 of 20)`, work item written
> `needs_human_review`, and **all 7 rows byte-identical to baseline — nothing
> deleted**. Pass: plan restored, `COMPLETED`, both cohorts 100%, 7 sections
> saved, numbers reported on the successful save. Full account, including four
> dead ends before `save_page_sections` could even be reached, in the lane's
> NOTES §7; the induction recipe is in its RUNBOOK.
>
> **The PLAN cohort is what refused — the rows cohort read 7 of 7 = 100%.** The
> obvious single-cohort design would have waved that run straight through. This
> is the ratchet, demonstrated live rather than argued.
>
> **`guardian`'s question is ANSWERED for one consumer and still open for five.**
> page-rerender has no `error_step` on `save_sections` and the orchestration went
> to FAILED, so the refusal genuinely stops the pipeline. **Unmeasured for
> page-build-handler, pageflow-builder, page-rebuild, site-work-orchestrator and
> tool-recreation-handler** — if any of those swallows a step error and marks the
> run complete, the refusal there is a row nobody reads. Worth one induction each.
>
> **FOLLOW-UP, deliberately not done here: the refusal text lies at this call
> site.** It ends *"the rows this run did not confirm are retained and a later run
> that sees the whole corpus will prune them"* — true for 135, which refuses only
> the prune, and FALSE here, where the whole save is refused, nothing is pruned
> later, and the page simply is not rebuilt until someone acts. An operator could
> reasonably read it as self-healing. The sentence is in `prune_floor.go`'s shared
> `Reason()`, so fixing it (a caller-supplied "what happens next" clause) is a
> signature change to a shared mechanism and belongs on its own merits, not
> riding inside a bug fix.

> **RE-SCOPED 2026-07-31 ~22:45, after B and C landed — the sentence is now false
> for THREE of the four consumers, not one.** Verified by reading each caller, not
> by grep:
>
> | consumer | on refusal | is the sentence true? |
> |---|---|---|
> | `index_code_symbols` (135) | skips the prune, run continues | **yes** — a later healthy run does prune |
> | `save_page_sections` (A) | `return nil, err` → step FAILED (induced live) | no |
> | `populate_nav_tables` (B) | `enforceNavPruneFloor` errs → caller `return nil, err` (`populate_nav_tables_action.go:149-151`) | no |
> | `sync_page_links` (C) | `enforceLinkRegistryFloor` errs → caller `return nil, err` (`site_db_actions.go:461-463`) | no |
>
> So the shared refusal now tells the majority of its callers' operators that the
> situation self-heals when it does not. The originally-stated reason for deferring
> ("it is one consumer out of two") no longer holds; what still holds is that the
> clause lives in the shared `Reason()`, so the fix is a signature change touching
> all four and wants its own council round.
>
> **Shape when someone takes it:** `Reason(op, subject, configKey)` gains a
> caller-supplied "what happens next" clause — 135 keeps today's sentence, A/B/C
> get "nothing was written and the existing <page|nav|link set> still stands; this
> run will not retry it." Do NOT simply delete the sentence: "NOTHING was deleted"
> is the load-bearing half and must survive.
>
> Not done in this session because `prune_floor.go` was another session's live
> territory at the time (they committed it at 22:33 and were still writing at
> 22:41), and a same-file collision on a shared tree is the one thing no hook can
> catch.

## The guardian's question, ANSWERED for all six consumers of site A (2026-07-31)

*"Does a failed `save_page_sections` step actually stop the pipeline, or is it
swallowed and the orchestration marked complete?"* — `guardian`, medium, round
`a54172b6`. Settled by reading every consumer's step config (extracted by text
window, because the step-level census misses the loop-nested ones) plus one live
induction:

| consumer | `error_step` on the save step | what a refusal does |
|---|---|---|
| `page-rerender` | none | **orchestration FAILED — proven live 2026-07-31** |
| `pageflow-builder` | none | fails the orchestration [INFERRED — same engine, same absent `error_step` as page-rerender; not induced] |
| `page-rebuild` | none | fails the orchestration [INFERRED, as above] |
| `site-work-orchestrator` | none | fails the orchestration [INFERRED, as above] |
| `page-build-handler` | `mark_item_failed` → `update_work_item_status` → `complete_error` | work item marked FAILED, and **the orchestration reports COMPLETED** — measured: its one retained run carrying `__step_error` ended `COMPLETED` |
| `tool-recreation-handler` | `complete_error` → `complete_workflow` | completes; no orchestration failure |

**So the guardian was right that two consumers do not fail — and wrong that this
makes the refusal a row nobody reads, for a structural reason worth keeping:**

1. **Content is protected in all six, unconditionally.** The floor returns before
   the DELETE, so nothing is removed whatever the pipeline does with the error.
   Pipeline error-handling changes the *visibility* of a refusal, never the
   *protection*.
2. **The refusal is durable in all six.** The guard writes the `site_work_items`
   row (`save_refused_incomplete`, `needs_human_review`, severity high) **before**
   returning the error, so it survives an orchestration that reports COMPLETED and
   outlives the ~2-day `orchestration_states` retention. That is precisely why the
   durable surface was built rather than relying on the step status, and this is
   the evidence that it was the right call.

**What is still worth doing:** induce a refusal on `page-build-handler` or
`tool-recreation-handler` to confirm the work item lands on the
COMPLETED-reporting path too. The three `[INFERRED]` rows above are the cheapest
remaining gap — one induction each would convert them, and the recipe is in the
lane RUNBOOK.

The same bar 135 was held to: **a green run proves nothing.** The floor is inert
on healthy input by design. Induce the fault (write a page with a deliberately
short section set, or point the nav builder at a partial page list), watch the
refusal fire with its numbers, confirm nothing was deleted, then clear the
induction and confirm a normal run still prunes. 135's live induction is scripted
in `docs/agent_docs/docs024_key_docs_latest/bugfix_135_prune_floor/RUNBOOK_prune_floor.md`
and transfers directly.

## Provenance

- `bug_historian`, HIGH, corr `14239fa4-552f-4821-abaf-ea15ccee4ea5` round 2
  (quoted above), plus its MEDIUM on B and C: *"the same unguarded
  reconciliation-delete mechanism at lower content-loss stakes … but still exposed
  to a truncated/partial run producing a small-but-nonzero delete with no error —
  the exact 'no error, no warning' silent-drop shape this council exists to
  catch."*
- The rule to reuse: `platform/orchestration/actions/prune_floor.go`, register
  **CTXA-025**, closed case `bugs_closed/135`.

---

## Contribution, 2026-07-31 (bugfix_092 lane) — `link_registry` has never held a row, so its delete floor is theoretical today

Evidence only, no direction, and no competing fix — this file's lane owns the mechanism.
Found while auditing `extractSiteID`'s callers for `bugs_open/092`, which took me through
`site_db_actions.go` and its `link_registry` writer.

```sql
SELECT count(*) AS rows, count(DISTINCT source_site_id) AS sites, max(created_at) AS newest
FROM link_registry;
--  0 | 0 | (null)
```

**Zero rows, all-history** (the table has no retention job, so this is not a window
artefact — `newest` is NULL because nothing was ever inserted).

Bearing on your third site: the delete at `site_db_actions.go:1474` is the reconciliation
half of `ExtractAndSyncLinksAction`, and its *insert* half has evidently never run to
completion on any site. So for `link_registry` specifically, a completeness floor guards a
corpus that does not yet exist — which may change how you rank it against
`page_components` (your stated live one) without changing whether the floor is right.

Two things I could NOT determine, marked so nobody inherits them as findings:

1. **Why it is empty.** `ExtractAndSyncLinksAction` returns a *success-shaped*
   `{"links_extracted": N, "persisted": false}` when `site_id` does not resolve — it takes
   `extractSiteID`, never checks the result, and skips persistence at
   `params.DB == nil || siteID == uuid.Nil`. That is a plausible cause. But the action runs
   on exactly one agent (`multipage-website-builder`) and there are **0 of its
   orchestrations in the retained window**, so "the exposure fires" and "the agent never
   runs" are indistinguishable from here. **[UNDETERMINED]** — I did not resolve it and it
   should not be quoted as though I had.
2. Whether that matters to your floor at all, which is your call.

Full audit of the five `extractSiteID` callers (three fail loudly, two do not) is in
`bugs_open/092` under the council-verdict section, filed there because the council's
`bug_historian` seat asked for it against 092's plan.

---

## Contribution, 2026-07-31 (separate lane) — sites B and C are CODED and committed; the per-nav-group cohort this file asked for is REFUTED

Code only, in `983e4b0a2`. **This lane does not own 165** —
`bugfix_165_reconciliation_deletes` does, and did site A including its live
induction. This is a contribution into their case, not a competing fix, and the
close is theirs to make. Council `c69e935a-7134-45c1-81c3-2f1da7831827`,
committed with `Council-Submitted:` before the verdict landed.

**Neither site is closed and neither is live.** The code is on the shared branch;
it is inert until a chassis roll, and — exactly as A was held to — a roll proves
nothing on its own, because both floors are inert on healthy input by design.
What is still owed is in "What is still owed" at the bottom of this section.

### The correction: this file's suggested cohort for B is wrong on the data

> This file says: **"B — per nav group, plus distinct nav items."**

The per-group half does not work, and the reason does not apply to site A.
`classifyPagesForNav` **RE-HOMES pages between groups as a matter of course**, so
group membership is a classifier OUTPUT rather than an independently-losable
class. Measured 2026-07-31: robot-hands.com holds a `tools` nav group (created
that day at 12:27; **no Go code writes that `group_key`** — `grep -rn
site_nav_groups --include=*.go platform/ internal/ pkg/` returns only
`populate_nav_tables_action.go` and `nav_tables.go`, so it arrived by hand SQL)
containing `/tools/gripper-safety-factor-calculator`, and the **current
classifier places that same page in `utility`**. A per-group cohort reads
`tools` 0/1 = 0% and refuses that site's nav rebuild **for ever**.

That is the third time in this case that the proposed partition was wrong and the
measurement said so — A's per-`slot_name`, A's first plan denominator, and now B's
per-group. The rule generalises; **the partition never does**.

### What shipped instead

**B — two cohorts, in genuinely different units:**

- `pages seen` — pages the run LOADED vs pages that exist under the loader's own
  predicate. This is the completeness signal proper ("did this run see the
  corpus?" rather than "did this run write less?"), and it is the one that catches
  the actual defect: `loadPagesForNav` logs a warning and **`continue`s past a row
  it cannot scan** (`:258-261`), so a partial read is silent and success-shaped.
  Measured with `navPageScopeSQL`, the **same constant the loader's `WHERE` clause
  is now built from**, so the count and the load cannot drift apart.
- `nav items` — rows to insert vs rows the DELETE removes (site-wide, the
  delete's exact complement). Catches what the page cohort cannot: a classifier
  collapse that loads every page and then places almost none of them, which is
  precisely `bugs_open/149` A2's shape.

**Measured false-positive rate: 0 of 16 sites.** The membership rule was replayed
in SQL against production and the item count a rebuild would produce **equals** the
stored count on every site with nav (finetuning.uk 25=25, ai-agent-orchestration
24=24, gaswholesalers 23=23, leopardess 23=23, robot-hands 17=17, gamesdesign
12=12, dartsonline 9=9, fundamentallyai 9=9, idea.uk 8=8, vonc 8=8, oufe 8=8,
relojistas 7=7, webdesign 5=5, vetcomparison 4=4, and both loan-calculator sites
1=1).

**NO ratchet cohort for B, and the asymmetry is worth keeping.** Site A needed a
plan cohort because `page_components` is **AUTHORED** — a truncating writer's short
output becomes the stored baseline, so the row cohort reads 2/2 = 100% for ever.
`site_nav_items` is **DERIVED** — recomputed from the page corpus on every rebuild
— so a wrongly-truncated nav is repaired by the next healthy run instead of
becoming the new baseline. **A derived table self-heals; an authored one
ratchets.** That is the test for whether a future consumer needs a second cohort.

**C — one cohort, and the partition DEFERRED on purpose, with the query to run
written into the file header.** This file suggested "per link kind if there is one,
plus distinct targets". There is no distribution to partition on:
`link_registry` holds **zero rows all-history** (independently re-verified, and it
matches the `bugfix_092` contribution above). Guessing a partition from the shape
of the SQL is what this case's own PLAN decision 3 warns against, so C ships the
single unpartitioned cohort — sound at any distribution — plus the measurement to
take once the corpus exists.

**C is inert BY CONSTRUCTION today** (`Stored=0` reads as fully confirmed, so
every prune is allowed) and arms itself the moment the insert half starts working.
That is the state nobody would remember to add a guard in, which is the argument
for adding it now rather than the argument against.

### Also done, because B and C would each have re-spelt it

`pruneFloorDetail` and `emitPruneRefusalWorkItem` moved into `prune_floor.go`.
Site A invented both inline; three spellings of one durable surface is the drift
class this council reviews for. Both are **additive and inert** — reachable by
nothing until a caller names them, no existing consumer's behaviour changes — so
under the owner ruling of 2026-07-29 §1 this is normal gate scope, not
architecture scope. `prune_floor.go`'s header no longer lists B and C as
"candidates, deliberately NOT converted"; all four consumers are named.

### Verification actually performed

18 tests, and **four negative controls run rather than claimed** — the tests were
each watched failing with the guard broken:

| control | result |
|---|---|
| neuter the nav floor (no cohorts) | exactly the 4 refusal tests fail; **no allow test fails** |
| add back a per-group cohort | `TestNavFloorAllowsAPageReHomedBetweenGroups` fails — it is what stops the refuted shape being re-added |
| neuter the link floor | its refusal test fails |
| drift `loadPagesForNav` off `navPageScopeSQL` | `TestLoadPagesForNavUsesTheSharedScopePredicate` fails |

Full `actions` package suite passes against a clean `git archive HEAD` + these
changes only (HEAD `b080cb4ae` at the time).

### What is still owed, and by whom

1. **The roll, then BOTH branches induced for B** — same bar as A: induce the
   refusal (point the nav builder at a partial page list), watch it fire with its
   numbers, confirm nothing was deleted, clear the induction, confirm a normal
   rebuild still prunes. `RUNBOOK_prune_floor.md`'s live induction transfers.
2. **C cannot be induced the same way** and this is stated rather than hidden:
   with `link_registry` empty and `ExtractAndSyncLinksAction` having 0
   orchestrations in the retained window, there is no live run to induce. Options
   are (a) close C on the tests plus the roll and say so plainly, or (b) hold C
   open until `bugs_open/092` makes the insert half run. **That is a judgement for
   the owning lane, not for this contribution.**
3. **Read the council verdict** on `c69e935a-7134-45c1-81c3-2f1da7831827` and act
   on a REVISE/REJECTED — the code is already on the shared branch.
4. **The guardian's open question from A applies to B and C too**: does a failed
   step actually stop the consumers, or do they mark the orchestration complete? A
   answered it empirically by inducing. B and C have not.
