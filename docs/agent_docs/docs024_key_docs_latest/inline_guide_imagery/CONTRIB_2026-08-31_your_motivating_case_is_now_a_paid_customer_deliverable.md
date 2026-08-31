# CONTRIB 2026-08-31 — your motivating case is no longer a dartsonline guide; it is a paid customer deliverable

**To:** whoever picks up `inline_guide_imagery`.
**From:** the session reviewing boxingonline.com (`d2aa5206-73bc-4707-a69c-2702c1eb9152`,
order BR-9AUZ59, **the first paid customer build**), 2026-08-31.

`PLAN_2026-08-14_durable_inline_guide_imagery.md` still reads **"Status: design, nothing
implemented."** Seventeen days on, here is what shipped to a paying customer:

- 4 guide pages (`/guides/tool-*-guide.html`), `article-body` of **4,415 / 4,706 / 5,200 / 6,158
  characters**, `[MEASURED 2026-08-31]`.
- **Zero in-body images on any of them.** Every served page on the site carries exactly one
  `<img>` and it is the logo.
- Zero `illustration` and zero `infographic` rows in `site_plan_imagery` for the site (the whole
  request set was 4 hero + 3 icon + 1 logo).
- Fleet-wide, ever: `infographic` = **1 row on 1 site**; `illustration` = 19 rows on 5 sites.

The owner reviewed the site the evening it was built and raised in-body imagery unprompted,
naming infographics and timeline graphics specifically, and pointing at one of these four guides
as the worst case ("the only interesting thing it says … the whole of the page doesn't say much
more than that").

Your plan's framing still holds and is still the right one — the gap is **durability**, not
capability, because in-body markup lives inside the single LLM `content` field and dies on the
next regeneration. What has changed is only the stakes and the exemplar. Detail and the
fleet-wide imagery census are in
`docs/agent_docs/docs024_key_docs_latest/editorial_design_uplift/CONTRIB_2026-08-31_the_infographic_kind_has_ONE_row_fleet_wide_and_the_first_paid_site_shipped_with_one_image.md`.

Nothing is dispatched at this site by me; `site_delivery_and_editor` owns its pipeline.
