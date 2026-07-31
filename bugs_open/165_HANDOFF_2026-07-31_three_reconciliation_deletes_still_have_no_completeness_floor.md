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
>    scores 2/6). Applies to B as much as to A — `site_nav_items` carries the same
>    lock columns.

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
| A | `save_page_sections_action.go:532` | `DELETE FROM page_components WHERE page_id = $1 AND <agent-writable>` | **HIGH — real content, lost before.** GUARDED 2026-07-31 (`ecf738002`); open until induced live |
| B | `populate_nav_tables_action.go:147,150` | `DELETE FROM site_nav_items WHERE site_id = $1` then `site_nav_groups` | medium — regeneratable, but a partial nav is served |
| C | `site_db_actions.go:1474` | `DELETE FROM link_registry WHERE source_page_id = $1` | medium — regeneratable |
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

**For site A, the acceptance test has an EXTRA question the council raised and
nobody has answered** (`guardian`, medium, round `a54172b6`): does a failed
`save_page_sections` step actually stop the six consumers, or do they swallow it
and mark the orchestration complete? A refusal the pipeline reports as `complete`
adds a row nobody reads. What is visible from config is not enough to decide —
`page-build-handler` has `error_step: mark_item_failed`, `page-rerender` and
`tool-recreation-handler` have none, and the other three nest the step inside a
loop where a top-level census cannot see it. **The induction below answers it
empirically: induce the refusal, then check whether the orchestration row reports
failure or reports complete.** Do that before closing site A.

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
