# 052 — a listing regenerates from the page set and re-advertises pages that were never built

*Found 2026-07-20 by the robot-hands thread while removing a dead tool card. **Proven on one
site; fleet exposure measured and currently low** — see the scale section, which is
deliberately unflattering to this bug's importance.*

**Family:** the *derived-field* family, with `/bugs_open/023` (CTA urls are recomputed
label-blind on every render, so authored edits cannot hold). Same shape, different field:
a listing's contents are **derived**, not authored, so editing them is undone by the next
render. Read 023 first — its addendum has the fuller write-up of the class.

Adjacent but **not** the same as `/bugs_open/015` (a mistyped `page_type` orphans a page
from its machinery). 015 and this share a *symptom* — a live link to a page that was never
built — but 015's cause is the page failing to build, and this one's is the listing not
caring whether it built. Fixing 015 would not fix this.

## What happened

`robot-hands.com`'s homepage tool list advertised five tools. Two of them,
`tool-matchmatrix` and `tool-robot-payload-budget-calculator`, were `build_status='planned'`
and had never been built — both URLs served **404**, and the cards were live on the homepage.

The dead card was removed from `page_components.content_data->'items'` and the page
re-rendered. **The card came straight back**, all five entries restored, at
`updated_at 17:10:23`:

```sql
SELECT jsonb_array_length(content_data->'items'),
       (SELECT string_agg(i->>'name',', ') FROM jsonb_array_elements(content_data->'items') i)
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.name='index' AND ...;
-- 5 | tool-matchmatrix, tool-gripper-payload-calculator, tool-gripper-cycle-time-estimator,
--     tool-grip-force-friction-calculator, tool-robot-payload-budget-calculator
```

The listing is regenerated from the site's tool pages on render. It included the page
because the page **existed** — nothing in that path consults `build_status`.

## Root cause

A listing component's items are derived from the page set with no build-state filter. A page
that is `planned` (never built, no components, URL 404s) is indistinguishable, to the
listing, from one that is `deployed`.

Note the asymmetry that makes this easy to miss: **`needs_rebuild` pages usually still
serve**. All three of robot-hands' other tool pages are `needs_rebuild` and return 200 —
they were deployed once and the flag was set later. So the naive fix ("only list
`deployed`") would wrongly delist working pages. **`planned` is the state that means
never-built**; that is the distinction a fix has to draw.

## Scale — measured, and smaller than it looks

Fleet-wide there are 18 non-deployed `page_type='tool'` pages across 7 sites, but that
number is misleading for the reason above. Filtering to the ones that actually mean
never-built:

| | count |
|---|---|
| `planned`, non-archived, tool pages | **2** (dartsonline `tool-setup-builder`, vetcomparison `tool-compliance-deadline-calculator`) |
| ...of those, confirmed **404** live | **2** |
| ...of those, **linked from any component or nav** | **0** |

```sql
-- the linkage check (returned zero rows)
SELECT s.domain, p.name, cc.name FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
LEFT JOIN content_components cc ON cc.id=pc.component_id
WHERE pc.rendered_html LIKE '%/tools/setup-builder/%'
   OR pc.rendered_html LIKE '%/tools/compliance-deadline-calculator/%';
```

**So: the defect is real and proven, and it is currently biting exactly one site.** The other
two never-built tool pages are unreferenced, so no user can reach them. This is filed for the
mechanism, not for an outage. `[UNMEASURED]` whether non-tool listings (article grids, news
lists, card grids) share the same derivation and therefore the same gap — I only traced the
tool list.

## Containment applied (robot-hands only)

`pages.status='archived'` on the never-built page, which is the house route R6 established
and which the listings demonstrably respect (R6 archived six dead article rows and the hub
then listed exactly the three real guides). See
`docs024_key_docs_latest/robot_hands/SQL_2026-07-20_r4e_archive_unbuilt_tool_page.sql`.

Archiving per site is containment, not a fix: it requires a human to notice a 404 first,
which is the step that failed here for weeks.

## Fix candidates

1. **Filter the listing derivation on build state** — exclude `build_status='planned'`
   specifically, NOT "include only `deployed`" (that would delist the working
   `needs_rebuild` pages, see above). Smallest correct change; fixes the tool list only.
2. **Make it a property of the page set, not each listing.** If several listings derive from
   pages independently, each needs the same predicate and they will drift — the exact class
   `/bugs_open/023` is in. A shared "listable pages" helper is the structural version.
   Check first whether article/news/card listings derive the same way; if they do, prefer
   this.
3. **Emit a work item when a listing would advertise a never-built page**, rather than
   silently dropping it — the drop hides a planning gap (a page planned and never built is
   usually a symptom, cf. `/bugs_open/028`). Pair with 1 or 2; on its own it routes to
   `needs_human_review`, which nothing consumes (`/bugs_open/033`).

Prefer (1) now and (2) if the survey in `[UNMEASURED]` above comes back positive.

## How to verify a fix

- Create a `planned` tool page on a test site, render a page carrying the tool list, and
  assert the card is absent. Then set it `deployed` and assert it appears.
- Regression, and this is the one that matters: assert the three robot-hands
  `needs_rebuild` tool pages **still appear** in the list and still serve 200. A fix that
  quietly delists working pages is worse than the bug.
- Check the rendered page, not `content_data` — the item array is regenerated on render, so
  a DB read taken before the render tells you nothing (this is how the bug was missed the
  first time).

## Related

- `/bugs_open/023` + its addendum — the derived-field class this belongs to.
- `/bugs_open/015` — same symptom, different cause; do not merge.
- `/bugs_open/028` — a page-build no-op reporting `complete` is one way pages end up
  `planned` forever.
- `/bugs_open/033` — why a detection-only fix would rot.
