# PLAN 2026-08-16 — validate accepts what the planner's menu offered (bugs_open/282)

**Lane:** `bugfix_282_validate_accepts_menu`, opened 2026-08-16.
**Bug:** `bugs_open/282_HANDOFF_2026-08-15_validate_resolver_drops_planner_placed_tool_sections.md`
**Filed by:** the loancalculator lane, which is blocked on it (its D2 batch — 11
`owned_page_review` + `needs_page:index` — is held until this ships).

## The defect, in plain terms

A planner agent is shown a menu of components it may use. It writes a plan naming
some of them. A later step in the same workflow checks every name against reality
and throws away the ones it cannot recognise.

Those two steps disagreed. The menu was widened in August (migration 407) to
include a site's own **tool** components — the calculators on
loancalculator.co.uk — behind a per-site flag. The checker was never widened, so
it did not recognise a single one of them. The planner proposed twelve
calculators; the checker deleted all twelve; the plan was written without them,
and nothing failed. The only trace was one `Warn` line per deleted section.

## Evidence it is real and still live (2026-08-16, live DB)

Run corr `2f74a975-1a87-40a8-af88-a9bd2ecc1510` (build-site-planner, plan
`dcbae4df`, 2026-08-15 14:25):

| where | what it holds |
|---|---|
| `collected_data.available_components` | 151 rows, **including** 11 finance tool functions — the planner *was* offered them |
| `collected_data.llm_plan.result` | a tool function on **12 pages** (index: `["hero","tool-loan-repayment",…]`) |
| `collected_data.validate_plan` | **none** of them (index: 5 sections) |

`validate_plan` is `validate_site_plan`'s own output field, so the drop is
inside that action — before `write_site_plan` ever sees the plan. Code at HEAD
unchanged: `v3_site_actions.go` resolve pass drops any unresolvable name;
`loadComponentNameResolver` loads only `component_level IN ('section','element')`.

## The design decision: consume the offer's OUTPUT, do not copy its predicate

The bug file's preferred candidate was "mirror the menu's own predicate into the
resolver, factored into ONE shared helper". **Rejected, and this is the main
design call of the lane.** The menu is not Go — it is a SQL string living in
`agent_definitions`, and it has *already* drifted past 407's text: migration
**419** added a `requires-backend` gate to the same query, and guards its own
apply by asserting 407's exact bytes. A Go copy would be a third hand-maintained
copy of a string two migrations already treat as fragile.

So the acceptance surface consumes **the rows the planner was actually shown** —
`available_components`, already in `CollectedData` at validate time. One list,
one source, nothing to keep in step. 016b §9's own verdict for this class:
*"single-sourcing is a guarantee, a lockstep test is a backstop."*

Decisions that follow, and their reasons:

- **UNION, not intersection.** The menu's rows are ADDED to the section/element
  base; nothing is removed. Intersection (also drop names the menu *withheld*)
  would be a tightening on every site on the estate — a real hole for the
  `bugs_open/276` requires-backend class, but a separate decision with its own
  blast radius. Recorded as observed-not-fixed, below.
- **Opt-in, unsafe default OFF** (owner ruling 2026-08-02 §2). The step must name
  the field: `menu_field` in `validate_site_plan`'s step config. Absent key =
  today's behaviour, byte for byte.
- **The shared resolver's signature and query do NOT change.**
  `loadComponentNameResolver` has four callers: validate, and three in
  `apply_gap_plan_action.go` (content-gap-planner, 131 dispatches/30d). 407 and
  PLAN-049 record that the gap-planner's menu is **deliberately** not widened —
  *"gap-planning a NEW tool page is a different authority"*. A widening of the
  shared resolver would have handed it that authority silently. The new arm is a
  method the gap-plan path never calls.
- **Order-safe both ways**, which is mandatory here: config is live on apply, Go
  rides the next roll. Migration applied + Go unrolled = key unread = today.
  Go rolled + migration unapplied = no `menu_field` = today.

## Shape of the change

| file | what |
|---|---|
| `platform/orchestration/actions/component_name_resolver_menu.go` (new) | `addMenu(rows)`, `menuRowsFrom(collected, path)`, `resolvedViaMenu(fn)` + the rationale as a doc comment |
| `platform/orchestration/actions/v3_site_actions.go` | two small hunks: `menuOnly` field on the struct; the `menu_field` read at the call site + the "resolved via the planner's menu" log |
| `docs/agent_docs/sql_for_agents/439_validate_plan_accepts_the_planner_menu.sql` (+ ROLLBACK) | sets `menu_field: "available_components"` on build-site-planner's validate_plan step; guard checks menu_field AND `load_components.output_field` **together** |
| `platform/orchestration/actions/component_name_resolver_menu_test.go` (new) | 9 tests incl. the un-opted-in negative control |

**The new code went in a NEW file on purpose.** `v3_site_actions.go` is edited by
several sessions at once (it was dirty from the 283 lane this morning); a
same-file passenger cannot be prevented by any pathspec, so the smaller the hunk
in that file, the smaller the exposure.

## Verification (why the tests are worth their bytes)

All nine pass — and, per the estate's own rule that a passing test proves nothing
until it can fail, both load-bearing arms were **mutated**:

- `addMenu` made a no-op → `TestValidateSitePlan_MenuFieldKeepsAToolSectionTheBaseWouldDrop`
  fails with `sections = [hero faq], want [hero tool-loan-repayment faq]`;
- the `menu_field` gate removed (menu always read) →
  `TestValidateSitePlan_WithoutMenuFieldTheToolSectionIsStillDropped` fails.

Restored, all nine pass again.

## Observed, deliberately NOT fixed here

1. **Pass B2 restores realised `pages.sections` BEFORE the resolve pass**, so a
   re-plan on a *decomposed* site (positional slot names like `tool-2`,
   `prose-0` — 43 pages fleet-wide) would have those dropped by the same
   resolver. Adjacent to 285/LOCK-008; the menu arm does not address it, because
   positional slots are not component functions.
2. **Intersection semantics** — see above.
3. **TP-004 gates by page ROLE, not by section**: a tool function planned onto a
   content-role page with no locked row is placed by the generic builder without
   review. Bounded today by 407's own placement gate (only tools already on the
   site can appear in the menu at all).
