# 165 — three more reconciliation deletes have no completeness floor, and one of them is the table that has actually lost content

**Filed 2026-07-31** by the `bugfix_135_prune_floor` lane, **at the explicit
direction of the council gate's `bug_historian` seat** (corr
`14239fa4-552f-4821-abaf-ea15ccee4ea5`, round 2, severity HIGH). Status: **OPEN,
unowned.** Latent — no failure to point at *today*, but two of the named
precedents are failures that already happened.

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
| A | `save_page_sections_action.go:532` | `DELETE FROM page_components WHERE page_id = $1 AND <agent-writable>` | **HIGH — real content, lost before** |
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
