# PLAN 2026-08-06 — bugs_open/204: build path resolves sections by stored identity

Picked up under the standing bug-backlog brief (see
`HANDOFF_2026-08-05_next_bug_pickup.md`). Standing four run 2026-08-06 morning:
who-owns names only the FILING session's commits (loancalculator lane, which is
*blocked by* this bug, not fixing it); `git log` on
`plan_sections_action.go` shows nothing since 6a7ab87a8 (a predicate swap, not
the lookup); live-transcript grep found four sessions mentioning the file — all
on OTHER lines (867–875 = 201's empty-input return; 247, 989 = other bugs) or
quoting the bug file itself; `site_work_items` has no open item on this target.
**Unowned. Taken.**

## Re-verified 2026-08-06

Census re-run live: **57/57 unresolvable on loancalculator.co.uk, 87 across 6
sites** (was 86/5 at filing — loanandmortgagecalculator.co.uk gained one).
Chassis v1.0.1256. Lookup still `components[sectionName]` at `:918`, keyed by
name/function only (`loadComponentSchemas`, `:1144`).

## Decisions and their reasons

1. **Resolve by stored identity FIRST, name/function second — exactly 182's
   decided semantics, transplanted to the build path.** The re-render path
   (`rerender_page_sections_action.go:234-291`) already does id-first with an
   observe-only log when both routes resolve and disagree ("id wins where the
   name map used to"). Same ordering here, same log shape, so the two call
   sites of the same judgement cannot drift — drift between these two paths is
   the documented cause of this bug (016b §9; 182's fix edited this very file
   and left the sibling heuristic).

2. **The "no pageID" comment at `:1139` is about the workflow, not about
   identity being unavailable.** `pages` has UNIQUE `(site_id, name)`
   (`pages_site_id_name_key`) and the action already holds both. One join
   query resolves the page's `page_components` slot→component_id map. The
   rerender path resolves the page by the same two keys (`:149-155`).

3. **A pinned component whose template fails the guard defers LOUDLY and does
   NOT fall back to the name map** — same reasoning as 182's rerender comment:
   the page names that exact component as its own; rendering a
   coincidentally name-matched one instead is the same silent substitution one
   level down. The deferred item carries a `Missing` entry whose Reason names
   the component id and says "repair the component, do not create a new one",
   so the `needs_section_data` item it files is actionable rather than the
   empty-suffix junk the canary produced. It does NOT file
   `needs_new_component` (the component exists). One broken component must
   not block the other 56 sections — defer the section, not the run (unlike
   rerender, where a broken pin is fatal, because rerender overwrites good
   HTML in place; planning does not).

4. **A slot whose id is absent from the by-id result (retired component /
   is_active=false / no such row) falls through to the name path** — mirrors
   `loadContentComponentsByID`'s own contract ("the caller's fallback still
   gets a chance").

5. **Duplicate slot_names are LEGITIMATE** (LANDMINES.md: 11 pages use e.g.
   `generic-text-block` 2–3×). Same slot name + same component_id → map it;
   same name + DIFFERENT ids → drop that name from the map with a warn (falls
   back to today's behaviour) rather than picking one arbitrarily.

6. **A query error loading the map fails the action** — silent degradation
   here would plan junk (`needs_new_component` ×2 per section, measured) on a
   decomposed site; a loud transient failure costs a retry. Absence of rows
   (initial build, page not yet created) is normal and yields an empty map.

7. **`page_name` is Optional in the input spec** — empty pageName skips the
   map load entirely (no page identity to read), preserving today's behaviour.

## Edits

- `platform/orchestration/actions/plan_sections_action.go`:
  - new `loadPageSlotComponentIDs(ctx, db, siteID, pageName, logger)` —
    the one query + duplicate-slot handling;
  - in `PlanSectionsAction`: load the map + `loadComponentSchemasByID`
    (reused, NOT a third resolver) after `loadComponentSchemas`; insert
    **Path 0** (stored-identity) ahead of Path 1 in the section loop.
- new `platform/orchestration/actions/plan_sections_slot_identity_test.go`
  (sqlmock, the `rerender_page_sections_resolve_test.go` shape):
  positional slot resolves by id; id wins over name (observe-only);
  invalid pinned template defers loudly with named reason and files NO
  needs_new_component; no stored rows → name path unchanged; conflicting
  duplicate slot ids → name fallback; map query error → action errors.

## Verification (from the bug file §How to verify)

- Census: unresolvable-by-either-route = 0 for loancalculator (by-id route now
  exists for all 57).
- Post-roll: pod-grep a positive AND negative control, then re-fire the
  `voiceh-canary` content_rewrite and assert prose changed + zero
  `needs_new_component`/junk `needs_section_data` filed.
- The 12 `lock_type='permanent'` tool rows untouched (fix only changes
  resolution; the writer's lock handling is unchanged).

## Council round 1 (REVISE, corr d3e232b8) — corrections and answers, 2026-08-06

- **CORRECTED: "the council already reviewed the id-wins flip for 182" was an
  overclaim** (guardian seat's precedent check). 182 shipped `Council-Submitted`
  (corr `80fbbe7d`) and only its `fix_plan` artifact exists — **no council_report
  ever landed**, and `bugs_closed/182` itself says "verdict pending at close".
  The precedent is PRODUCTION precedent (live since a43be1e70, 2026-08-03),
  not council precedent. Logged in WRONG_CALLS.
- **CORRECTED: "sole consumer" was wrong** (guardian). TWO live workflow steps
  use the action: `page-build-handler.plan_sections` AND
  `page-content-writer.plan_sections` (bugs_open/087's fallback plan builder).
  Both map the same three keys (site_id / sections / page_name) the same way,
  so the change behaves identically for both.
- Caller inventory (bug_historian): `loadComponentSchemas` has exactly TWO
  callers fleet-wide — plan_sections:878 and rerender:232, both now id-first.
  `loadSingleComponentSchema`'s one caller is the selector (Path 2), whose
  component identity comes from its own scored candidate's function, not from
  a slot name — not the same judgement.
- Flip blast radius measured for the BUILD path (bug_historian): **23 stored
  sections across 9 sites** where slot_name name/function-resolves to a
  component OTHER than the pinned component_id (query in NOTES; rerender's own
  measure was 13 with a narrower predicate). Each fires the observe-only log
  when a build actually plans it.
- **Open architecture question, recorded as asked by the architecture seat:**
  should the tri-state resolution (id-hit / id-dropped-loud-defer /
  id-absent-fallthrough) be ONE shared helper called by both
  plan_sections and rerender_page_sections, rather than two structurally
  identical inline blocks? Not done in this fix: rerender's branch is welded to
  its fatal-list/carry semantics and plan_sections' to defer/work-item
  semantics — the shared part is the *decision*, the divergent part is the
  *consequence*. A shared decision-only helper is worth a look; routing to the
  next council round per the seat's note rather than growing this fix.
- loadStoredSections (rerender) was checked before writing
  `loadPageSlotComponentIDs` (reuse_agent): it reads full stored rows
  (content_data, rendered_html, position) by page_id for the render; the build
  path has no page_id and must not load page content to plan. Different key,
  different shape — the shared piece is already shared
  (loadComponentSchemasByID / componentInfoFromRaw).
