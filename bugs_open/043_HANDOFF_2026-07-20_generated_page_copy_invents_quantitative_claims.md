# 043 — generated page copy invents quantitative claims, and nothing checks them against the data

*Found 2026-07-20 by the robot-hands thread while fixing CTA pairing. **Live fabrication was being
served to the public on three pages of robot-hands.com** and has now been corrected on that site;
the platform defect is unfixed and the class is almost certainly fleet-wide.*

**Family:** the fabrication family — `020` (tool-recreation invents a dataset) and `001` (a re-plan
resurrects fabrication audited out days earlier). **Different mechanism, and that is the point.**
`020` is the tool-recreation handler having no data-dependency contract. This is ordinary
**content generation** writing specific numbers into `page_components.content_data` with no source
and no check. A site with no tools at all is still exposed.

## What was live

`robot-hands.com` publishes gripper specifications. The index holds **five** grippers
(`products`, `category='gripper'`, `site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'`). Three
separate stat blocks were serving figures invented whole:

| page | component | field | published | reality |
|---|---|---|---|---|
| `about` | `content-block-about` | Gripper Models Indexed | **1,200+** | 5 |
| `about` | `content-block-about` | Actuation Technologies | **6** | no actuation field exists in the data |
| `about` | `content-block-about` | MatchMatrix Queries Run | "Tracked & Published" | nothing tracks or publishes this |
| `gripper-detail` | `system-stats` | Gripper Models Indexed | **2,400+** | 5 |
| `gripper-detail` | `system-stats` | Manufacturers Covered | **140+** | 5 |
| `gripper-detail` | `system-stats` | Actuation Technologies | **6** | as above |
| `gripper-detail` | `system-stats` | Scored Specification Parameters | **18** | 4 are compared; 7 spec keys exist |
| `index` | `brief-explanation` | Actuation technologies benchmarked | **6** | as above |

Verified live before the fix:

```bash
curl -s https://robot-hands.com/about.html \
  | grep -oE '1,200\+|>6<|Gripper Models Indexed|Actuation Technologies'
# -> 1,200+ / >6< / both labels
```

Two pages also carried the same claim in prose — *"The catalog spans pneumatic, electric, vacuum,
magnetic, soft-robotic, and adhesive actuation technologies"* — and `gripper-detail`'s stat
descriptions asserted process the site does not perform: *"datasheet sources and last-verified
dates are displayed at model level"*, and scoring *"across 18 parameters including gripping force
(N), stroke (mm), cycle time (ms), IP rating, payload-to-weight ratio, and operating pressure
range"*. `products.specifications` holds seven keys and none of cycle time, payload-to-weight or
operating pressure.

## Three things make this worse than a bad number

**(a) The numbers are internally consistent, so they read as sourced.** 1,200+ models across 6
technologies scored on 18 parameters is a coherent story. Nothing about it looks generated. A
reviewer without the DB open has no signal at all.

**(b) The suffixes prove nobody read the output.** `gripper-detail`'s stat block carried
`stat1_suffix='%'` on "Gripper Models Indexed" and `stat2_suffix='ms'` on "Manufacturers Covered" —
unedited placeholders from a generic template. The page was rendering "2,400+%" and "140+ms" and
had been doing so live. That is a strong tell that this component was populated and deployed with
no human or machine pass over the rendered result.

**(c) A partial cleanup is worse than none.** `index/brief-explanation` had two of its three values
already blanked to an em dash (`—`) — someone recognised those had no source — while the fabricated
`6` was left in place. The block then reads as *checked*, with one surviving invention.

## Root cause

The content-generation prompt asks for a stat block and supplies page context, but **no source
binding**: there is no slot for "these are the figures you may use", and no post-generation check
that a number appearing in `content_data` traces to anything. `020` fix candidate (2) proposed
rewriting its rule 9 so it binds *data* and not just arithmetic — that rule lives in
`tool-recreation-handler`'s prompt and does not cover this path at all.

This is also why `020` fix candidate (4) — a machine-readable no-unsourced-claims site flag —
would not have caught it: that candidate is scoped to blocking *generated content deploys which
introduce records*, and these are scalar claims in a stat component, not records.

## Fix candidates

1. **Bind stat components to a source.** A `system-stats` / stat-block component should require
   each value to carry a provenance field (a query, a table, or an explicit `unverified` marker),
   and render nothing where provenance is absent. Structural, and it makes the next invention
   impossible rather than detectable.
2. **Post-generation numeric audit.** Beside the existing `check_completeness`, add a check that
   extracts numeric claims from generated `content_data` and fails the item to
   `needs_human_review` when a figure cannot be tied to a countable source. Cheap net; catches the
   next variant of this and of `020`.
3. **Extend the fake-data prohibition to every generative prompt that can emit a number**, not just
   the tool prompts — the wording `020` proposes ("never generate, synthesise, seed or hard-code
   example records") needs a scalar sibling: *never state a count, total or coverage figure you
   have not been given*.
4. **Lint the placeholder suffixes.** `%` on a model count and `ms` on a manufacturer count are
   mechanically detectable and would have flagged this block on the day it deployed.

Prefer (1)+(3). (4) is nearly free and is the only one that would have caught this specific
instance at deploy time.

## How to verify a fix

- Generate a stat block for a site whose data is thin, and assert it either renders sourced figures
  or renders nothing — never a plausible invention.
- Site-wide sweep, which is how the second and third blocks here were found:
  ```sql
  SELECT p.name, e.k, e.v
  FROM page_components pc JOIN pages p ON p.id=pc.page_id,
  LATERAL jsonb_each_text(pc.content_data) AS e(k,v)
  WHERE e.k ~ 'stat[_]?[0-9]+_?value' ORDER BY p.name, e.k;
  ```
  Run it per site; every value should be traceable. **Check the rendered page too** — the suffix
  bug is invisible in `content_data` alone.
- Grep rendered pages for the suffix tell: `curl <page> | grep -oE '[0-9,]+\+(%|ms|x)'`.

## Containment applied (robot-hands only)

`docs/agent_docs/docs024_key_docs_latest/robot_hands/SQL_2026-07-20_r4b_fabricated_about_stats.sql`
and `..._r4c_fabricated_stats_sitewide.sql`. All nine stat values now trace to a query recorded in
the file header (5 models / 5 manufacturers / 4 parameters compared / 24 published figures).
Re-render queued as `source='robot-hands-r4-cta-pairing'`.

**Not yet done on robot-hands:** the same six-technology claim survives in **42 further
`content_data` fields** (body prose, `features`, `subheadline`, FAQ `questions`, `cards`) across
that one site. Those are substantive copy, not stat fields, and rewriting them is a content
decision rather than a containment step — left for the owner. That ratio (9 stat fields fixed vs
42 prose fields carrying the same unsupported claim) is worth noting for anyone sizing the
fleet-wide version of this: **the stat blocks are the visible tip.**

## Fleet-wide exposure — UNMEASURED

I checked one site. The `system-stats` component and the generic stat-block shape are not
site-specific, so the class very probably reaches other sites, but **I have not run the sweep
fleet-wide and no figure here should be read as if I had.** The verify query above is the sweep;
running it across all sites is the first thing the fixing thread should do, before choosing
between candidates (1)–(4).

## Related

- `/bugs_open/020` — the tool-recreation fabrication. Same family, different path. Fix both.
- `/bugs_open/001` — re-plan resurrecting audited-out fabrication.
- `/bugs_open/023`, `/bugs_open/033` — the delivery gap. robot-hands had 20
  `cta_names_unknown_destination` items in `needs_human_review` and none of them were ever read;
  a detector for this class would rot the same way unless routing is fixed too.
- `docs/agent_docs/docs024_key_docs_latest/vetcomparison/LEGAL_2026-07-15_*` — what published
  unsourced figures cost the last time.
