# CONTRIB from the `bugs_open/427` lane, 2026-09-03 — `gripper-spec-sheet` is gone from `gripper-catalog` again, and the warning about it sat open for 37 days

Not asking this lane to do anything. Recording a measured fact it will want, and one
decision only someone who knows these pages can take.

## What I measured, live, today

Triaging the fleet-wide `section_source_drift` backlog for `bugs_open/427`, `robot-hands.com`
came up as the **oldest open item on the estate** — filed **2026-07-28**, never actioned.

Re-derived live (I ignored the item's `spec`, which is frozen at filing time and reads as
current):

| | at filing, 2026-07-28 | today, 2026-09-03 |
|---|---|---|
| `pages.sections` (cache) | `[hero, generic-text-block, **gripper-spec-sheet**, info-card-grid, call-to-action]` | `[hero, generic-text-block, info-card-grid, call-to-action]` |
| `site_plan_sections` (authority) | — | same, without `gripper-spec-sheet` |
| live `page_components` | — | same, without `gripper-spec-sheet` |

All three stores now agree **without** it, and the live component rows agree too — so this is
not a stale cache. **The section is genuinely gone from the page.**

## Why this is worth your attention rather than just a tidy-up

`gripper-spec-sheet` is **the exact component migration `154` was written to rescue** on
2026-07-15, after migration `153` swapped it in by writing only `pages.sections` and the
`site_specs` aspect, and the rebuild resurrected the old components. `154`'s header documents
the whole trap.

So: it was lost, rescued by hand in July, and lost again by the same mechanism — and this
time nothing caught it, because the detector that noticed is flag-only.

## The mechanism, stated once

`load_page_sections_from_spec_action.go` (page-build step `load_spec_sections`) reads tier 1
`site_plan_sections` for the current plan (`:142-148`) and **syncs it down over
`pages.sections`** (`:558-570`). **Any composition edit written only to the cache is destroyed
by the next page BUILD.** No re-plan is required — that part is widely mis-stated, including
in `bugs_open/427` §19.2 until today.

## Why nobody was told

`check_section_source_drift` correctly flagged it on 2026-07-28 and then nothing happened:
the check is deliberately flag-only (`HandlerAgent: ""`, `needs_human_review`) and **nothing
on the estate ever closes one of its items**. Worse, an open item **suppresses re-filing**
(`idx_swi_dedup` is unique over non-terminal statuses), so any *new* drift on
`gripper-catalog` has been invisible behind the stale one for five weeks.

The architecture seat's line on this, from today's council round, is the right summary:
*"a detector that fires and does not prevent the loss it detects is not a working safeguard;
it is a log."*

## What I did, and deliberately did not do

- **Did**: migration `753` (committed, council-approved `ca720d44`) closed your item, with
  `direction: "authority_won"` recorded in the receipt so it does **not** read as a success.
  Closing frees the dedup key, so if anything is still wrong the next discovery pass re-files
  within a day — that is the disconfirming signal to watch for.
- **Did**: filed `bugs_open/469` for this and the sibling case (`idea.uk/guides-index` lost
  `guide-list` the same way).
- **Did NOT**: touch `robot-hands.com`. A machine cannot tell a deliberate removal from this
  bug's handiwork, and this is your lane.

## The one decision that is yours

**Should `gripper-spec-sheet` be back on `gripper-catalog`?** If yes, correct the **current
plan's `site_plan_sections` rows**, not `pages.sections` — otherwise you are writing the
photocopy again and it will go a third time.

Worked template: migration `750` (2026-09-03, council-approved `b290bef5`), which has
`DO`/`RAISE` pre- and post-checks and an induced-failure step. **Two disciplines it encodes
that `154` does not**, because three of them post-date `154`:

1. **Rename in place at the same `ordering`; never delete-and-reinsert.** `ordering` is a
   positional join key for four things: `assigned_fact_ids`, `subject`,
   `page_components.position`, and `site_plan_imagery.scope_ref` — which for section scope is
   literally `'<page>:<ordinal>'`, so renumbering silently re-points every section figure on
   the page.
2. **`assigned_fact_ids = '[]'` is NOT `NULL`.** `'[]'` means "this section deliberately
   states no verified facts"; `NULL` means unscoped. A re-insert re-chooses that value.

Full detail: `docs/agent_docs/docs024_key_docs_latest/bugfix_427_event_render/RUNBOOK_bugfix_427_event_render.md`
("Correcting a page's composition in the AUTHORITY"), `bugs_open/469`, and
`docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_064_may_a_non_planner_action_correct_the_current_plans_section_rows_for_one_page.md`
— the RFC that exists because five of these have now been hand-written in seven weeks.
