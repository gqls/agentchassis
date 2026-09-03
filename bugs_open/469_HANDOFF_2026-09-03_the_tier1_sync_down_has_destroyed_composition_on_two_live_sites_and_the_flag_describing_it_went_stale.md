# 469 — the tier-1 sync-down has already DESTROYED page composition on two live sites, and the detector's own flag went stale so nobody could tell

Filed 2026-09-03 by the `bugs_open/427` lane (session "427") while triaging the
`section_source_drift` backlog. **Status: OPEN, unowned.**

**Severity: MEDIUM.** The damage is done and is not spreading fast (two confirmed
instances in five weeks), but it is *silent*, it destroyed deliberate human work, and
the one mechanism built to warn about it reported the damage and then went quiet in a
way that reads as "resolved".

## 0. First-hand verification, per CLAUDE.md's 2026-07-31 ruling

No `090` run. Substituting first-hand verification, all `[MEASURED 2026-09-03]` against
the live `clients_db` in this session: the three stores and the live `page_components`
rows were queried directly for every page named below, and the reverting code path was
read end-to-end at source (`load_page_sections_from_spec_action.go`) rather than
inferred from behaviour. The mechanism is not a hypothesis — it is documented in
migration `154`'s own header from 2026-07-15, and this file's contribution is the
evidence that it has since **fired again and completed**.

## 1. The mechanism (established, not new)

A page's section list lives in three stores, resolved in priority order by
`load_page_sections_from_spec_action.go` (the page-build step `load_spec_sections`):

1. `site_plan_sections` for the current plan — **authoritative** (`:142-148`)
2. `site_specs.site_plan` aspect
3. `pages.sections` — the materialised cache

and the winner is **synced DOWN over `pages.sections`** (`:558-570`).

So an edit made only to the cache is destroyed by the next page **build**. No re-plan is
required. (A re-plan is in fact *safe*: `reconcilePlanWithRealised`,
`v3_site_actions.go:7701-7724`, snaps a `deployed`/`needs_rebuild` page's realised
`pages.sections` back **onto** the plan proposal first. `bugs_open/427` §19.2 has this
backwards and is corrected there.)

## 2. What is NEW here: the loss has completed, twice

`check_section_source_drift` correctly flagged both of these at the time. Both items
then sat open at `needs_human_review` — the check is deliberately flag-only
(`HandlerAgent: ""`) — until this session closed them.

| site / page | item filed | what the cache held at filing | what all three stores say today |
|---|---|---|---|
| `robot-hands.com` / `gripper-catalog` | 2026-07-28 | `[hero, generic-text-block, **gripper-spec-sheet**, info-card-grid, call-to-action]` | `[hero, generic-text-block, info-card-grid, call-to-action]` |
| `idea.uk` / `guides-index` | 2026-08-04 | `[hero, **guide-list**]` | `[hero, content-listing]` |

`[MEASURED 2026-09-03]` in both cases the live `page_components` rows agree with the
authority too — so this is not a stale cache, it is the **component genuinely gone from
the page**.

**`gripper-spec-sheet` is the exact component migration `154` was written to rescue on
2026-07-15**, after `153` swapped it in by writing only `pages.sections` + the aspect.
It was rescued, and then lost again. That is the finding: the remedy did not hold, or a
later write re-introduced the divergence and the sync-down won the second time.

A third, `oufe.com/contact`, also resolved authority-wards (the cache had lost
`contact-info`; the authority restored it). Recorded for completeness, but the direction
test cannot distinguish "a deliberate removal was destroyed" from "an accidental
omission was repaired", and this one looks like the latter. **Do not count it as a third
loss without looking.**

## 3. Why nobody noticed — the part worth fixing

Three properties compound:

- **The check is flag-only and nothing ever closes an item.** Six were open when this
  session looked, the oldest 37 days.
- **The item's `spec` is frozen at filing time.** `spec.authoritative` and
  `spec.pages_sections` are a snapshot, and they read as current. A reader triaging the
  backlog by reading the items learns nothing about today.
- **An open item SUPPRESSES re-filing.** `idx_swi_dedup` is `UNIQUE (site_id, item_key)`
  over non-terminal statuses, so while a stale item sits open, genuinely new drift on
  that same page cannot be filed. The detector blinds itself on exactly the pages it has
  already flagged.

So the sequence is: drift is detected → flagged → nobody has a handler → the build wins
→ the stores agree again → **the item now describes a state that no longer exists and
looks like it might be resolved** → and any fresh drift on that page is invisible behind
it.

## 4. What this session already did

- Migration `753` closes only items whose stores agree again, re-derived at apply time,
  and records a `direction` (`cache_held` / `authority_won` / `third_list`) in each
  receipt so a close cannot silently ratify a loss. Four closed; three are
  `authority_won`. `apis.uk/index` — a live divergence owned by another lane — was
  excluded by the data and left open.
- Migration `750` corrected `boxingonline.com/tool-fight-calendar`'s authority to match
  its live page, so `bugs_open/427`'s fix stops being one build away from destruction.

Neither addresses **this** bug, which is about the two pages that already lost their
composition and about the blindness that let it happen unremarked.

## 5. What is actually open

1. **Should `gripper-spec-sheet` be back on `robot-hands.com/gripper-catalog`, and
   `guide-list` on `idea.uk/guides-index`?** This needs a *human* who knows what those
   pages are for. A machine cannot tell an intended removal from this bug's completion —
   which is precisely why the revalidator design for this item type must classify such a
   case as `unknown` and never as `resolved`. Owner or the owning lanes
   (`robot_hands_gripper_dossier`, `idea.uk`).
2. **The detector needs a closer.** Any check whose items nothing ever resolves will
   accumulate exactly this failure. `check_section_source_drift` never populates
   `CheckResult.Resolved`.
3. **The framework gap that produced the divergence in the first place** — there is no
   typed way to write a composition correction to all three stores, so every lane
   hand-writes SQL and some write only the cache. Going to architecture review as an RFC
   out of the `427` lane; see `bugs_open/427` and the `bugfix_427_event_render` lane docs.

## 6. Cross-references

- `bugs_open/427` §19–§21 — the mechanism, corrected; the migration `750` precedent.
- `docs/agent_docs/sql_for_agents/153_gripper_detail_page_swap.sql` and
  `154_product_detail_plan_sections_fix.sql` — the original 2026-07-15 case on this very
  site, and its hand-written remedy.
- `docs/agent_docs/sql_for_agents/750_…`, `753_…` — this session's two migrations.
- `bugs_open/443` — the adjacent class: pages born with a layout in the cache and
  nothing in the authority (`create_blog_posts`).
- `LANDMINES.md`, "`pages.sections` is a materialised CACHE" — the standing entry.
