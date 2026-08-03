# PLAN — bugfix 177: tool_content items are unsatisfiable at birth

**Lane opened 2026-08-03** (session "bugfix 182" — renamed intent; 182 was found
owned by a live session at the exit-plan-mode step, so this lane took the next
cold bug instead: 177 at reference-heat 23, symbol-grep clean, filing session
closed last night).

## The diagnosis, sharpened from the bug file

`bugs_open/177` filed the consequence (9 of 9 `tool_content:%` items dead in
`needs_human_review` since 2026-07-14, 0% success ever) and marked the cause
inside the generator `[UNVERIFIED]`. It is now verified, and it is an asymmetry
between two sibling actions:

- `deploy_tool_action.go` (source `tool-deployer`) declares
  `pages.sections = ["hero-tool","tool-guide-intro","<fn>","tool-cta"]`
  (deploy_tool_action.go:343-346) **and then** raises the `tool_content` item.
  Zero of the failed items are from this path.
- `create_tool_component_action.go` (source `tool-generator`) copies the item
  emission (create_tool_component_action.go:343-380) **but not the section
  declaration** — its `pages` INSERT (:288-294) carries no `sections`, so the
  column defaults to `[]`. **All 9 failed items are from this path.**

The handler (`page-build-handler`, a DB workflow) resolves sections via
`load_page_sections_from_spec` → site_plan_sections (current plan) →
`site_specs.site_plan` aspect → `pages.sections` → same-role sibling synthesis
(needs plan membership). A freshly generated tool page is in none of these, so
`plan_sections` gets `[]`, `ready_count == 0`, and the WDS-004/149 no-op
routing parks the item in `needs_human_review` — correctly, given what it was
handed. **The item is unsatisfiable at the moment it is minted.**

Control that proves the mechanism: `tool_guide:%` items from the SAME two files
(the companion guide page, which DOES declare
`sections=["hero","article-body","call-to-action"]`) run 4 complete / 1 review.
Same emitter, same handler, same window; the only difference is the declaration.

## Live urgency (measured 2026-08-03 11:37 BST)

`a5cabea0` (tool-cma-obligation-checker, minted 08-02 — the class GREW after
filing) already has **two triaged `content_rewrite` dependents**
(`9e9ec430`, `18bc832c`). A dependency is released only by `complete`/`verified`
(bugs_closed/176), so the exact shape that stalled the fleet twice on 08-02 is
armed again.

## Decision: fix candidates weighed

Bug file's candidates: (1) don't raise when the page is complete; (2) raise only
when the spec declares prose sections; (3) split the handler no-op outcome —
rejected by the file itself (masks 1/2, and bugs_open/015 shows "no sections"
can be a real defect); (4) sweep.

**Chosen: 1+2 merged, as an emit-side satisfiability guard, plus 4.**
A shared helper in `platform/orchestration/actions` used by BOTH tool paths:
raise the `tool_content` item **iff the page declares at least one buildable
prose section beyond the tool widget itself**, asking the same sources the
handler will ask (read-only mirror of the loader's fallbacks 1–3; synthesis
excluded — it requires plan membership, and a page in the plan has its own
sections). If no prose sections: skip, log at INFO, and **surface the skip in
the action's output map** (the 182 lesson: a silent no-op must be observable).

Why not "declare the sections in the create path" (make the item buildable):
that changes the shape of every future generated tool page — behaviour that has
NEVER been exercised (the one deploy-path page with the full declared shape,
from April, still has 1 slot). The bug file's verify section explicitly warns a
completing `tool_content` item is "different (and possibly also wrong)
behaviour". That is an owner decision, recorded as an open question, not a bug
fix.

Why the guard reads the DB rather than trusting a `sections` parameter: the
edge is real — 33 current-plan section rows exist for tool-named pages (e.g.
`{hero,generic-text-block}`), so a generated tool page that IS in the plan
would be satisfiable and must still get its item. The DB-grounded guard answers
correctly for callers that don't know that.

Not architecture-scope under the 2026-07-29/08-02 rulings: no shared mechanism
changes its guarantee — `needs_content_page`'s contract, the handler, and every
other emitter are untouched; this is two call sites in one package sharing a
guard, the `page_role_upsert.go` precedent (bugs_open/175) for the same file
pair.

## Edits (≤8, for the council submission)

> **REVISED after prior-art research (same day).** The first draft preserved
> the hand-rolled `INSERT … ON CONFLICT DO NOTHING` inside the helper. The
> council has already ruled on that exact shape: `gapPlanWorkItem`
> (apply_gap_plan_action.go:720-762, corr `a5b70424`) — hand-rolled inserts
> reimplement half of `insertWorkItem`'s semantics and predate the shared
> door having the fields callers needed. The revised design routes through
> `insertWorkItem` (load_work_item_actions.go:1220), with
> `recurrenceExpected: true` for the same stated reason as the gap plan: the
> item is an ACTION REQUEST — a completed predecessor means the request
> succeeded, and two-strike must not brand a re-request unresolved
> (bugs_open/024's regression). Two-strike counts only `complete`/`failed`
> (:1247), so the 297 sweep's `wont_fix` rows cannot poison future emits.

1. New file `platform/orchestration/actions/tool_content_item.go`: unexported
   helper `raiseToolContentItem(ctx, params, logger, req)` — resolves declared
   sections (plan tables → spec aspect → pages.sections, read-only), counts
   prose sections ≠ tool function. Zero → `skipped_no_prose_sections`, no
   insert. Otherwise builds the same `workItem` both call sites build today
   (needs_content_page / medium / priority 50 / page-build-handler / triaged /
   item_key `tool_content:<fn>:<site>`) and routes it through `insertWorkItem`
   in a short transaction. Disposition string returned: `raised` |
   `skipped_no_prose_sections` | `deduped_open_item` | `insert_failed` (kept
   non-fatal, today's Warn-and-continue behaviour).
2. `create_tool_component_action.go`: replace the inline INSERT (:365-380) with
   the helper; put the disposition in the action's return map.
3. `deploy_tool_action.go`: replace the inline INSERT (:474-489) with the
   helper; put the disposition in the return map. Behaviour preserved here (its
   pages always declare 3 prose sections, so the guard resolves them and
   raises) — the point is one seam, no drift.
4. Sweep (SQL, live immediately): the 8 `needs_human_review` rows → `wont_fix`
   with the original error preserved in the reason (the 286 triage precedent —
   NOT `complete`, no work happened); clear `depends_on` on `9e9ec430` and
   `18bc832c` so the two real crosslink items can dispatch.
5. Register hygiene, same commit: update TL-003 (documents the divergence as
   current behaviour — now conditional emission), note in TL-009 that its "do
   not queue needs_content_page for a page that doesn't want generic content"
   half has shipped; the shape question itself stays open/aspirational.

Out of scope, named for the close-out: the two companion-guide emits in the
same files (same hand-rolled shape, but their items WORK — 4/5 complete);
the 24+ `needs_page` rows from five other sources with the same no-op error
(different item type, different emitters, possibly legitimate deferrals).

Prior art the fix stands on: TL-009 + `tools/tool_widget_clobber/PLAN` §5
Option 2 already recommend exactly this guard ("If a tool page does not want
generic content, do not queue needs_content_page for it"); bugs_open/033's
OWNER RULING 2026-07-25 ("the queue should not fill" — and needs_content_page
is an uncovered type for the revalidator drain, so nothing auto-closes these);
bugs_closed/015's correction (tool pages' empty `sections` is legitimate —
their content comes from elsewhere).

## Verify

- Unit: guard against a page with `sections=[]` and not in any plan → skip; a
  page with plan sections beyond the tool → raise; deploy-path shape → raise.
- Live (after image roll): create a tool via tool-generator on a test site;
  assert NO `tool_content:%` row appears (the bug file's own verify: the
  correct outcome is that the item is never created); positive control:
  `tool_guide:%` item still appears. Pod-grep a symbol the change ADDED and one
  it REMOVED (the negative control, per the 153 lesson).
- Queue: the two `content_rewrite` items leave `triaged` once deps are cleared.

## Open questions for the owner

- Should generated tool pages get prose around the widget (declare
  `["hero-tool","tool-guide-intro","<fn>","tool-cta"]` in the create path)?
  That would make the class buildable instead of unmintable — a design change,
  deliberately NOT taken here.

## CORRECTION 2026-08-03 ~12:10 — edit 4 NARROWED after approval, before apply

> **CORRECTED:** the approved plan cleared `depends_on` on the two triaged
> `content_rewrite` dependents (the 286 precedent: "the crosslink stands on
> its own merits"). Between the verdict and the apply, the diagnosis run on
> that class COMPLETED and found that the item 286 released on that exact
> reasoning (`93f2a3b7`) **destroyed content when it dispatched** — whole-slot
> regeneration, changed heading, dropped paragraphs (bugs_open/178's
> mechanism; doc_notes 2026-08-03). Releasing two more would repeat a known
> destructive outcome. So the sweep retires the zombies and deliberately
> LEAVES `9e9ec430`/`18bc832c` dep-blocked as a visible interlock; the 154
> lane (owner of 178) releases them with its fix. The guardian seat's
> advisory ("confirm no race with the in-flight diagnosis before applying")
> is what prompted the re-read that caught this — the race was fine; the
> diagnosis's own FINDING was the blocker.
