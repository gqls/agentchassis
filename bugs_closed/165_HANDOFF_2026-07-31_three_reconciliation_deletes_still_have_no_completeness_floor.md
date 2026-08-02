# 165 — three more reconciliation deletes have no completeness floor, and one of them is the table that has actually lost content

**Filed 2026-07-31** by the `bugfix_135_prune_floor` lane, **at the explicit
direction of the council gate's `bug_historian` seat** (corr
`14239fa4-552f-4821-abaf-ea15ccee4ea5`, round 2, severity HIGH).

> **STATUS: CLOSED 2026-08-02 by OWNER RULING.** All four call sites of the
> destructive shape are guarded and live on chassis **v1.0.1228**. A and B are
> proven on both branches in production; C is live and mutation-proven offline but
> **structurally un-inducible** — see "Closing the case" at the foot of this file
> for exactly what was and was not demonstrated, and why that was judged enough.
> The one piece of live proof still owed rides on `bugs_closed/092`.

~~Status: **OPEN — site A DONE, B and C outstanding.**~~ Latent — no failure to
point at *today*, but two of the named precedents are failures that already
happened.

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
| B | `populate_nav_tables_action.go:147,150` | `DELETE FROM site_nav_items WHERE site_id = $1` then `site_nav_groups` | medium — regeneratable, but a partial nav is served. ~~NOT yet live, NOT yet induced~~ **DONE: live v1.0.1228, BOTH BRANCHES PROVEN 2026-08-02** |
| C | `site_db_actions.go:1474` | `DELETE FROM link_registry WHERE source_page_id = $1` | medium — regeneratable. ~~NOT yet live~~ **live v1.0.1228; mutation-proven offline, and UN-INDUCIBLE live — the table is empty fleet-wide and its only consumer never runs. Closed on inertness; the induction rides on `bugs_closed/092`** |
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
> | `sync_page_links` (C) | `enforceLinkRegistryFloor` errs → caller `return nil, err` (`site_db_actions.go:461-463`) | no — **but see the caveat, C is changing** |
>
> ⚠ **THE C ROW IS ALREADY STALE — RE-CHECK IT BEFORE RELYING ON IT.** Verified
> against committed `983e4b0a2`. Minutes later (22:50, working tree, **uncommitted**)
> another session was mid-refactor of exactly this: `enforceLinkRegistryFloor` now
> returns `(map[string]interface{}, bool)` rather than an `error`, and
> `site_db_actions.go:461` had not yet caught up (the package did not compile). If
> that lands, C becomes skip-the-prune-and-continue like 135 — which would make the
> shared sentence **TRUE** for C and the count "2 of 4", not 3. The finding itself
> is unaffected: A and B are false either way. Recorded rather than quietly
> corrected because a table that was right when written and wrong when read is how
> a confident claim propagates.
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
| `pageflow-builder` | none | fails the orchestration — **verified by mechanism 2026-08-01** (was `[INFERRED]`) |
| `page-rebuild` | none | fails the orchestration — **verified by mechanism** (was `[INFERRED]`) |
| `site-work-orchestrator` | none | fails the orchestration — **verified by mechanism** (was `[INFERRED]`) |
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

**The three `[INFERRED]` rows are now VERIFIED — by establishing the engine's rule
rather than by three inductions (2026-08-01).** Reading the coordinator, a step
error takes exactly one of three exits, in this order:

1. **loop iteration with `continue_on_error`** → skip to the next iteration, the
   workflow CONTINUES (`coordinator.go:908`, `loop_error_handler.go:71-89`);
2. else **`error_step`** (step-level, then `config.error_step`) → route there
   (`coordinator.go:3350-3363`);
3. else → **`failWorkflow`** (`coordinator.go:3363`).

All three consumers nest `save_sections` inside a loop, so exit (1) was the real
risk — a refusal skipping one page and letting the build report success would have
been the silent outcome the guard exists to prevent. Measured: **`continue_on_error`
is UNSET on all three** (`pageflow-builder.build_pages_loop`,
`page-rebuild.build_pages_loop`, `site-work-orchestrator.build_items_loop`), and
none has an `error_step` on the save step. So they take exit (3) and fail.

This is stronger than the three inductions it replaces: it establishes the
mechanism for *every* consumer including future ones, and it would catch a config
change that three passing inductions would not. **It also revealed a
disproportion** — a refusal on one page fails the whole multi-page build in those
three, which is exactly `bugs_open/173`. That case was filed from site C; the
census and the widened blast radius (four loops, not one) are contributed there.

**Still worth doing:** induce a refusal on `page-build-handler` or
`tool-recreation-handler` — the two that route on error — to confirm the work item
lands on the COMPLETED-reporting path too. Recipe in the lane RUNBOOK.

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
Found while auditing `extractSiteID`'s callers for `bugs_closed/092`, which took me through
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
`bugs_closed/092` under the council-verdict section, filed there because the council's
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
   open until `bugs_closed/092` makes the insert half run. **That is a judgement for
   the owning lane, not for this contribution.**
3. **Read the council verdict** on `c69e935a-7134-45c1-81c3-2f1da7831827` and act
   on a REVISE/REJECTED — the code is already on the shared branch.
4. **The guardian's open question from A applies to B and C too**: does a failed
   step actually stop the consumers, or do they mark the orchestration complete? A
   answered it empirically by inducing. B and C have not.

### Sites B and C — council APPROVED 2026-08-01, and the two rounds it took are the useful part

`c69e935a-7134-45c1-81c3-2f1da7831827`, **round 3 APPROVED** (12 approve, 2
advisory objections, none high). Rounds 1 and 2 both returned REVISE. Code:
`983e4b0a2` (B+C), `d1f6a9426` (round-2 attempt), `f4825c9ca` (round-3 revert +
`bugs_open/173`), `af33969f2` (the contract test). **Still not live, still not
induced** — see "What is still owed" above, unchanged.

**What round 2 got wrong, recorded because it is the transferable bit.** Measuring
site C's blast radius showed its only live caller is nested in
`multipage-website-builder`'s `generate_pages_loop`, which sets no
`continue_on_error` — so a refusal fails the **whole site build** over one page's
links. Round 2 answered that by making the action never error: return
`(detail, bool)` and skip. **Four seats ruled it backwards and they were right.**

> `constitution` (HIGH): *"the plan's own rationale identifies the real cause as
> the workflow's missing `continue_on_error`/`error_step` … then works around that
> gap by making the action silently skip (never error) instead of fixing the
> workflow's error-routing config. This is a fix whose rationale names the
> mechanism it steps around rather than repairs it."*

Plus `bug_historian` (HIGH — a refusal turned into a success-shaped return whose
only signal is a queue), `architecture` (a bespoke soft-skip with no shared
vocabulary, built because the core routing could not carry the load), and the
observation that **a bool is ignorable where an error is not**.

**The shape to remember: I had diagnosed the real cause correctly and then routed
around it in the same breath.** The measurement was right; the conclusion drawn
from it was a workaround wearing the measurement as justification. Round 3
reverted C to the uniform error contract and filed the routing gap at its own
layer as **`bugs_open/173`** (loop error routing has no substep granularity), with
fix candidates and an explicit note on why setting `continue_on_error` on that
loop is *also* not the answer — it is loop-level, so it would turn a page that
genuinely failed to build into a silently skipped one.

**Two corrections to seats, recorded because both cut in my favour and should
still be right.** (1) `bug_historian` cited `bugs_open/033` as showing the
`needs_human_review` queue is unmonitored; 033's own header says its three display
bugs are FIXED and LIVE and the queue is visible — what stays open is the *drain*.
The objection survived without that premise, which is why C was reverted rather
than defended. (2) My own first answer on the `handler_agent` landmine was wrong
in the other direction: `count(handler_agent)` counts empty strings, so I read
"363 of 363 carry a handler" off a column where **265 carry none**. Writing no
handler is the majority pattern for that status, not an anomaly. Logged in
`WRONG_CALLS.md`.

## SITE B IS LIVE AND PROVEN — both branches, 2026-08-02 (owning lane)

**Supersedes item 1 of "What is still owed", and answers item 4 for B.**

### Live on v1.0.1228, pod-verified on both replicas

```
save_refused_incomplete          2   <- site A, live since v1.0.1223 (discriminator)
nav_rebuild_refused_incomplete   2   <- site B
link_sync_refused_incomplete     2   <- site C
ZZZ_NOT_A_REAL_SYMBOL_ZZZ        0   <- the grep is honest
```

**No same-diff negative control was available and that is stated rather than
papered over**: `983e4b0a2` is purely additive — it removes no string literal, so
there is nothing that must read 0. The discriminating pair does the same job: a
stale v1.0.1223 image would show site A's marker and NOT B's or C's.

### The refusal branch — induced on oufe.com, 2026-08-02 10:24 BST

Correlation `323173dd-9fd1-4899-bb61-f835f0516b13`, agent `nav-updater`
(`ensure_site_record` → `refresh_nav_tables`). Baseline: 9 pages in scope, 8 nav
items in 3 groups, 0 orchestrations in 24h, no open nav work item.

Induction: **16 marked synthetic rows** (`label LIKE 'INDUCED-165B-%'`) into the
utility group, taking the stored side to 24 while the write side stayed at 8.

```
FAILED @ refresh_nav_tables

populate_nav_tables: rebuild: REFUSED for site oufe.com — this run re-confirmed
too little of what is stored (prune_floor_ratio=0.50): nav items 33% (8 of 24).
```

- **Nothing was deleted.** All 8 real rows byte-identical to the baseline md5s,
  and all 16 synthetics still present (24 total) — the refusal returns *before*
  the transaction opens, exactly as the file header claims.
- **The right cohort fired, alone.** `pages seen` read 9 of 9 and stayed silent;
  only `nav items` was below floor. That is the cohort that exists specifically
  to catch a classifier collapse the page load cannot see.
- **Durable surface landed**: one `site_work_items` row,
  `nav_rebuild_refused_incomplete:a0d7f1ae-…`, status `needs_human_review`, no
  `handler_agent`.
- Cleanup verified by sweeping for the marker **fleet-wide**, not just on the
  target: `SELECT count(*) FROM site_nav_items WHERE label LIKE 'INDUCED-165B-%'
  OR url LIKE '/induced-165b-%'` → 0. oufe.com back to its 8 baseline rows.

**One datum that differs from site A**: for B the refusal reason IS in
`orchestration_states.error`. On A it was only in the chassis log. Do not
generalise either way — read both.

### The pass branch — proven by a REAL production run, not an induced one

Not induced, because it did not need to be. `site-adoption-agent` orchestration
`dcf88c1c-1fc2-4f48-8064-e5a18725c4a6`, **COMPLETED 2026-08-01 09:04**, on
loancash.co.uk:

```
"completeness_status": "passed",
"completeness_reason": "populate_nav_tables: rebuild: floor cleared for site
  loancash.co.uk (prune_floor_ratio=0.50); pages seen 100% (18 of 18),
  nav items 100% (1 of 0)"
```

Both cohorts reported with their raw numbers on a PASSING run, which is the
"don't present the alarm as the output" property the file was designed for. A
genuine build clearing the guard is **stronger** evidence than a synthetic pass:
it proves the floor is reached and inert on real traffic. Note the cosmetic
oddity `100% (1 of 0)` — the stored=0 first-appearance case, correct by design
(a class appearing for the first time must never be able to refuse a prune), but
the sentence reads strangely and is worth a wording pass if the text is ever
revisited.

### The misleading-refusal-sentence finding is now DEMONSTRATED, not merely reasoned

The oufe refusal above ends with the shared sentence:

> *"NOTHING was deleted; the rows this run did not confirm are retained and a
> later run that sees the whole corpus will prune them."*

The first clause is true. **The second is false for site B** — nothing is
"retained pending a later prune"; the entire rebuild was refused, and the next
healthy run deletes and rebuilds the lot wholesale. An operator reading this is
told to wait for a tidy-up that will never come. Previously filed as reasoning;
it is now a production artefact an operator could actually be shown. Unchanged
in substance: the fix belongs in `prune_floor.go`'s `Reason()` as a
caller-supplied "what happens next" clause, and wants its own council round —
**do not simply delete the sentence, "NOTHING was deleted" is load-bearing.**

### Site C — the induction is IMPOSSIBLE today, not merely undone

Re-measured 2026-08-02: `SELECT count(*), count(DISTINCT source_page_id),
max(created_at) FROM link_registry` → **0, 0, NULL**. Still empty fleet-wide,
all history. Its only live consumer (`multipage-website-builder`) still has no
orchestrations. So **neither branch is reachable**: with `Stored=0` every cohort
is treated as fully confirmed, so the floor cannot refuse, and with the action
never running there is nothing to pass either.

This is a blocked prerequisite, not a skipped step. What C *does* have: the code
is live (pod-grepped above) and its refusal path was proven offline by mutation
(neutering the link floor fails its refusal test, run rather than claimed by the
implementing lane). Option (b) of item 2 above is therefore the honest reading —
**C stays open, blocked on `bugs_closed/092`** making the insert half run. It
carries no risk in the meantime precisely because it is provably inert.

### So: what is still owed on this case

- **C's live induction**, blocked on `bugs_closed/092`. Not actionable from here.
- **The refusal-sentence fix** — now with a live artefact behind it.
- **`page-build-handler` / `tool-recreation-handler`** (the two of six consumers
  that route on error rather than failing) still unmeasured empirically. Content
  is protected in both by construction — the floor returns before the DELETE and
  the work item is written before the error — so this is a *visibility* question,
  not a data-loss one.
- **165 does NOT close on B alone.** Both the `bug_historian` and `architecture`
  seats asked specifically that this case not be closed the moment the
  high-stakes site was done. A and B are done; C is blocked. Closing it needs an
  owner ruling on whether "live + mutation-proven + provably inert" clears the
  bar for C, and that is a decision, not a measurement.

## The refusal-sentence fix: APPROVED, and both objections answered with work (2026-08-02)

Council `22cdef56-da93-42f7-b9f9-71ff82abcdf6` — **APPROVED**, 11 reviewers,
6 abstained, `decided_by: "approved with 1 advisory objection(s) — none
high-severity"`. Code `56365d86b` (carries `Council-Submitted:`; 098 credits it
automatically now the correlation is approved — no amend, forward-only).

Both objections were acted on rather than filed, because both were right.

### editquality, MEDIUM — "presents mutation evidence as proof rather than a best-effort local run"

The objection: the mutation run happened in `platform/orchestration/actions`, a
package other sessions edit concurrently, where **a mutant that breaks the build
reads the same as one correctly caught** and a restore can collide with someone
else's edit. Fair, and the fix is cheap.

**Re-run in an isolated `git archive HEAD` tree** (`go.mod go.sum platform
internal pkg` + the two fixture dirs the package tests read — the first attempt
failed the baseline because `doc_subjects_common_test.go` reads
`docs/agent_docs/sql_for_agents` and the experience-register harvest, which is a
missing-fixture failure, *not* a broken HEAD). Baseline green in isolation, then
each mutant applied to a **fresh** copy and **compiled before being trusted**:

| mutation | compiles | result | predicted |
|---|---|---|---|
| M1 restore the fixed borrowed clause | yes | both aftermath tests FAIL | both |
| M2 empty falls back to a consumer clause | yes | only the empty-case test FAILS | only that one |
| **M3 specificity control** — break the unrelated DISABLED branch | yes | only `TestEvaluatePruneFloorDisabledIsAllowedAndSaysSo` fails; **my two PASS** | my two pass |

M3 is evidence the first run did not have, and it is the one that answers the
objection properly: it shows the two new tests fail on *their own* mutation and
not on any change to `Reason`. Restore verified by grep count; isolated baseline
green afterwards.

### editquality, LOW — "'exactly four call sites' rests on a grep, not verified"

Replaced the grep with a mechanical enumeration: **revert the signature to three
arguments in the isolated tree and let the compiler list the callers.**

```
platform/orchestration/actions/code_symbols_actions.go:350:5:      too many arguments in call to verdict.Reason
platform/orchestration/actions/link_registry_prune_floor.go:169:3: too many arguments in call to verdict.Reason
platform/orchestration/actions/nav_prune_floor.go:210:3:           too many arguments in call to verdict.Reason
platform/orchestration/actions/save_sections_prune_floor.go:213:3: too many arguments in call to verdict.Reason
```

Four, repo-wide, from `go build ./...`. **Scope stated honestly:** `go build`
does not compile `_test.go`, so that enumeration covers production callers; the
nine test call sites are covered by the package test build passing.

### bug_historian — "cannot confirm this is not a resubmission of an already-adjudicated diff"

It is not, and the check is exact. `git log -G "func \(v pruneFloorVerdict\)
Reason\("` over `prune_floor.go` returns **two** commits ever: `10524a03c`
(bugs_closed/135, which created the method) and `56365d86b` (this change). The
three council correlations in that file's history are `14239fa4` (135, created
it), `c69e935a` (sites B+C, which used the rule **unchanged**) and `22cdef56`
(this one). No prior round adjudicated a signature change.

**Use `-G`, not `-S`, and this bit me while answering it.** `git log -S` is
occurrence-COUNT based, so it missed my own commit — the edit added a parameter
while preserving the string `func (v pruneFloorVerdict) Reason(`, leaving the
count unchanged. `-S` reported only 135 and would have supported a confident
"nothing has ever changed this signature". Filed in `LANDMINES.md`.

---

## POST-DEPLOY STATUS, 2026-08-02 (the B+C lane, verified at the running artefact)

The owner rolled the chassis to `v1.0.1228`. State of the four call sites, checked
against both running replicas rather than against git or the tag:

| site | action | code | live | proven |
|---|---|---|---|---|
| — | `index_code_symbols` | `bugs_closed/135` | yes | yes |
| A | `save_page_sections` | `save_sections_prune_floor.go` | yes (`v1.0.1223`) | **both branches** |
| B | `populate_nav_tables` | `nav_prune_floor.go` | **yes (`v1.0.1228`)** | **refusal induced live; pass observed in a genuine build** |
| C | `extract_and_sync_links` | `link_registry_prune_floor.go` | **yes (`v1.0.1228`)** | **not possible — see below** |

Verification for B and C was a pod-grep of both replicas with a positive control, a
pipeline control (site A's symbol, live since `1223`) and a **negative** control (the
multi-line page-scope predicate the fix deleted → 0 on both). Details and the two
mis-spelled first attempts: lane `NOTES` §14, `WRONG_CALLS.md` 2026-08-02.

### One item is committed, approved, and NOT live

`56365d86b` — the refusal sentence carried `index_code_symbols`' aftermath ("a later
run … will prune them"), which is false for A, B and C, all three of which refuse the
**whole** operation. Committed 09:36 UTC; the pods started 08:47 UTC. Pod-grepped:
the corrected clause returns **0** on both replicas, the clause it replaces returns
**1**. It ships on the next roll; nothing needs redoing.

Its one live consequence: the durable work item `bf2e9ad6` (oufe.com, from B's
induction) still ends with the false sentence, because the text is rendered at write
time and stored. Future refusals get the corrected wording automatically; that row
does not. Left in place deliberately — it is the induction's proof artefact, and
editing a work item's `spec` to make an old refusal read like a new one would be
rewriting evidence.

### Why site C cannot be proven, restated so it is not read as undone

`link_registry` is **0 rows, 0 pages, `max(created_at)` NULL, all history,
fleet-wide**, and its only live consumer has no orchestrations. Neither branch of
the floor is reachable: `Stored = 0` can never refuse (an empty cohort reads as
fully confirmed by design — a new class must not be able to block a prune), and an
action that never runs can never pass. Blocked on `bugs_closed/092`, and carrying no
risk in the meantime *because* it is provably inert. The guard arms itself the
moment the insert half starts working.

### THE CLOSING QUESTION IS A DECISION, NOT A MEASUREMENT — and it is unchanged

Every measurement this bug can produce has been produced. What is left is a
judgement the repo's own bar does not settle: **does `live + mutation-proven +
provably-inert` clear "fixed AND live" for site C, when the reason it cannot be
induced is that the table it guards has never held a row?**

- **If yes** — 165 closes now. Three sites are proven, the fourth is inert by
  construction, and the outstanding sentence fix is a wording change with a queued
  roll.
- **If no** — 165 stays open, pinned to `bugs_closed/092`, and closes when
  `link_registry` first holds rows and C can be induced like A and B were.

Recorded here so whichever way it goes, the reasoning is on the file rather than in
a session that has ended.

## Closing the case — 2026-08-02, owner ruling

**Closed on: all four call sites guarded, live on v1.0.1228, three of the four
branch-pairs demonstrated in production, and the fourth provably inert.** The
owner's call was on site C specifically, which is the only part that does not
meet the bar this file set itself.

### What is actually proven, stated so nobody has to re-derive it

| site | table | guarded | live | refusal branch | pass branch |
|---|---|---|---|---|---|
| A | `page_components` | yes | v1.0.1223 | **induced in production** 2026-07-31 (`planned sections 35% (7 of 20)`, 7 rows byte-identical) | **induced** — rebuilt normally once cleared |
| B | `site_nav_items` | yes | v1.0.1228 | **induced in production** 2026-08-02 (`nav items 33% (8 of 24)`, 8 rows byte-identical, 16 synthetics intact) | **a genuine production run** — `site-adoption-agent` on loancash.co.uk, both cohorts 100% |
| C | `link_registry` | yes | v1.0.1228 | **offline only** — mutation-proven; un-inducible live | **un-inducible live** |
| (135) | `code_symbols` | yes | v1.0.1218 | induced | induced |

### Why C was closed without the live induction this file demanded

Not "we ran out of time" — **there is nothing to point at.** `link_registry` holds
**0 rows, 0 pages, `max(created_at)` NULL**, all history, fleet-wide (re-measured
2026-08-02, and the table has no retention job so this is not a window artefact).
Its only live consumer, `extract_and_sync_links`, is reachable from exactly one
agent (`multipage-website-builder`) which has no orchestrations in the retained
window. So **both** branches are unreachable by construction:

- the refusal branch cannot fire, because `Stored=0` is treated as fully confirmed
  (a class appearing for the first time must never be able to refuse a prune), so
  the floor allows every prune until the table has rows;
- the pass branch cannot fire, because the action never runs.

The guard therefore **carries no risk in its un-induced state** — it is inert by
construction, not merely untested — and it **arms itself the moment the insert
half starts working**, which is the state this guard exists for and the state
nobody would remember to add a guard in.

**The live induction is not abandoned, it is transferred.** It becomes reachable
exactly when `bugs_closed/092` makes the writer receive its link constraints and the
insert half runs. Whoever closes `092` should induce C's refusal then: insert
synthetic `link_registry` rows for one source page so `LinksToWrite / LinksStored`
falls below 0.5, run the sync, and confirm the refusal fires with its numbers and
that **no** row was deleted. The recipe transfers from
`RUNBOOK_reconciliation_deletes.md` § R-B2.

### What this case established beyond its own fix

- **The cohorts do not transfer, and this file's own suggested partitions were
  wrong twice** — refuted by measurement for A (998 of 1,009 `(page_id,
  slot_name)` groups hold exactly one row, so every per-slot cohort is 1 stored
  and any legitimate single-section removal scores 0%) and for B (`classifyPagesForNav`
  re-homes pages between groups as a matter of course, so group membership is a
  classifier OUTPUT, not an independently-losable class). **Two distinct ways a
  partition can be wrong — too SMALL, and not STABLE — and both were invisible
  until measured.**
- **The general test for whether a consumer needs a second, different-unit
  cohort:** an **AUTHORED** table ratchets (a truncating writer's short output
  becomes the stored baseline, so the row cohort reads 100% for ever and only an
  independent unit still sees the loss); a **DERIVED** table self-heals (the
  numerator is recomputed from the corpus, so a healthy run repairs it). Ask which
  yours is *before* adding a cohort.
- **A cohort can be uninducible from the data side, and that is a design property
  rather than a gap in the test** — B's `pages seen` compares the loader's rows
  against a count built from the same predicate, every nullable column is
  `COALESCE`d and `name`/`id` are `NOT NULL`, so only a genuine driver fault makes
  them diverge. Worth knowing before someone plans an induction for it.
- **The shared refusal text carried one consumer's aftermath to all four**, false
  for three of them, telling their operators to wait for a tidy-up that never
  comes. Fixed as a required caller-supplied clause (`22cdef56`, APPROVED).
- **A green run proves nothing about any of these guards.** They are inert on
  healthy input by design. That is the whole reason each branch had to be induced,
  and the reason C's honest status is "inert", not "fine".

### Residuals, all owned elsewhere

1. **C's live induction** → rides on `bugs_closed/092`.
2. **`page-build-handler` / `tool-recreation-handler`** — two of six consumers
   route on error rather than failing, so a refusal there is recorded while the
   pipeline reports success. Content is protected in both by construction (the
   floor returns before the DELETE and the work item is written before the error),
   so this is a **visibility** question, not a data-loss one.
3. **A refusal on one page aborts a whole multi-page build** → `bugs_open/173`
   (loop error routing has no substep granularity), filed by the B/C lane with
   this lane's measurements contributed: four loops across four agents, and a
   fleet census showing 9 of 20 live loops set `continue_on_error` but **none**
   wraps a floor-guarded action, so nothing is being swallowed today.

Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_165_reconciliation_deletes/`.

---

> ### CORRECTION 2026-08-02, same day, by this lane — "blocked on `092`" was wrong twice
>
> Everything above says site C's live induction is **blocked on `bugs_open/092`**.
> Two errors in that, both mine, both from repeating an inherited pointer instead
> of checking it:
>
> 1. **`092` was already CLOSED** — `bugs_closed/092`, fixed and live on v1.0.1219,
>    induced-proven **2026-07-31**, i.e. the same day the pointer naming it as open
>    was written. I carried it into this file, the 016b §10 index row and the
>    concept register before checking. Paths corrected in all three.
> 2. **Closing `092` did not and will not unblock C.** `link_registry` was
>    re-measured today and is still 0 rows. So the dependency was never really on
>    `092`'s fix at all.
>
> **The real blocker, now measured rather than deferred:** `extract_and_sync_links`
> is carried by exactly one agent, `multipage-website-builder`, which has **0
> orchestrations in the retained window** while the live build pipeline
> (`build-dispatch-loop` 588, `build-pipeline-trigger` 587, `page-rerender` 22,
> `page-build-handler` 1) does not include it. And **0** orchestrations anywhere
> have ever mentioned `links_extracted`, so the action has not executed at all.
> `092`'s honest `[UNDETERMINED]` — "is the registry empty because the exposure
> fires, or because the agent never runs?" — therefore resolves to **the agent
> never runs**, and the full working is written up at the foot of `bugs_closed/092`.
>
> Bounded honestly: `orchestration_states` is retention-clocked (oldest row
> 2026-07-13), so that is "has not run in ~20 days", not "never". The all-history
> half is `link_registry` itself, which has no retention job and has never held a
> row.
>
> **Why this happened, because it is the more useful finding.** This file said the
> link-registry question was `092`'s territory; `092` said it was `165`'s and was
> "contributed there rather than competed with". **Each deferred to the other, so
> neither owned it, and both then closed.** A deferral names a destination and
> nobody re-checks that the destination accepted it. When you hand an item to
> another case, write it into *that case's* file, not only your own.
>
> **None of this changes C's disposition or this case's closure.** C is guarded,
> live, mutation-proven offline and inert by construction; what changed is *which*
> future event makes it inducible — not `092`'s fix, but a decision about whether
> `multipage-website-builder` is retired or revived. That is an owner call, raised
> at the foot of `bugs_closed/092`.


> **CORRECTION 2026-08-02 (later the same day) — the retention figure above is WRONG, and the conclusion needs a different source.**
>
> Everything above bounds "0 orchestrations" with *"`orchestration_states` is
> retention-clocked (oldest row 2026-07-13), so that is ~20 days"*. **It is ~24
> HOURS.** `COMPLETED` rows are reaped after about a day — measured: 2,504
> COMPLETED rows, oldest **24.7h**; FAILED oldest **25.4h** — and the whole-table
> `min(created_at)` reads 2026-07-13 only because `CANCELLED` (24), `RUNNING` (4)
> and `INITIALIZED` (2) are **not** reaped. A handful of stragglers in statuses the
> census was not about set a floor twenty times too long. Caught by watching a row
> I had quoted at 09:40 (`dcf88c1c…`) vanish by 10:40 while the table grew
> 2,454 → 2,546 and its oldest row never moved.
>
> **So the orchestration evidence never supported "the agent never runs" — only
> "not in the last day".** The conclusion is still correct, but on a different and
> much stronger source: **`site_specs` has no retention job and goes back to
> 2026-02-25** (1,874 rows, 36 sites), and across all of it the only
> `recommended_builder` ever recorded is **`pageflow-builder`** — 1,216 rows, 14
> sites. `multipage-website-builder` was never chosen, not once, in five months.
>
> Right answer, wrong reason, and the reason is what was published. Fleet landmine
> filed: "`orchestration_states` keeps terminal rows ~24 HOURS — and
> `min(created_at)` says 20 days". **Any other claim resting on an
> `orchestration_states` census needs re-bounding per status.**
