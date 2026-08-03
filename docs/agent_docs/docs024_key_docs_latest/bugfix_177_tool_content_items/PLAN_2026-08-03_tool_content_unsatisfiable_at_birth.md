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

1. New file `platform/orchestration/actions/tool_content_item.go`: unexported
   helper `raiseToolContentItem(ctx, params, logger, req)` — resolves declared
   sections (plan tables → spec aspect → pages.sections, read-only), counts
   prose sections ≠ tool function, inserts the item only when > 0, returns a
   disposition string (`raised` | `skipped_no_prose_sections` | insert error).
2. `create_tool_component_action.go`: replace the inline INSERT (:365-380) with
   the helper; put the disposition in the action's return map.
3. `deploy_tool_action.go`: replace the inline INSERT (:474-489) with the
   helper; put the disposition in the return map. Behaviour unchanged here (its
   pages always declare 3 prose sections) — the point is one seam, no drift.
4. Sweep (SQL, live immediately): the 8 `needs_human_review` rows → `wont_fix`
   with the original error preserved in the reason (the 286 triage precedent —
   NOT `complete`, no work happened); clear `depends_on` on `9e9ec430` and
   `18bc832c` so the two real crosslink items can dispatch.

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
