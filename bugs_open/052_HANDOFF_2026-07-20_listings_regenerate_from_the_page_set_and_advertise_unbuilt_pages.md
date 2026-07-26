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

> **CORRECTED 2026-07-26 — candidate 1's predicate is wrong, and the survey came back
> positive, so (2) was the right call, not (1).** Candidate 1 (`exclude
> build_status='planned'`) was already superseded by this file's own ADDENDUM below,
> which measured four `needs_rebuild`-and-never-deployed pages that candidate 1 would
> have kept advertising. The shipped predicate is the addendum's, not candidate 1's.
> The `[UNMEASURED]` survey is now **answered — see the STATUS section at the end.**
> One other listing derives from the page set the same way, and candidate 3 (emit a work
> item) remains unbuilt and still gated on `/bugs_open/033`.

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

- `/bugs_closed/023` + its addendum — the derived-field class this belongs to. (Moved to
  `/bugs_closed/` on 2026-07-25; this file said `/bugs_open/023` until 2026-07-26.)
- `/bugs_open/015` — same symptom, different cause; do not merge.
- `/bugs_open/028` — a page-build no-op reporting `complete` is one way pages end up
  `planned` forever.
- `/bugs_open/033` — why a detection-only fix would rot.

---

## ADDENDUM 2026-07-20 (bugfix-049 session) — the predicate should be `deployed_at IS NULL`, not `build_status='planned'`

This file's asymmetry section is right and I relied on it: `needs_rebuild` pages usually still
serve, so "list only `deployed`" would delist working pages. Measured fleet-wide, that holds
exactly — **34 of 34** `needs_rebuild` pages with a non-null `deployed_at` return **200**.

But the conclusion drawn from it, *"`planned` is the state that means never-built"*, is very
nearly right rather than right, and fix candidate 1 (`exclude build_status='planned'`) inherits
the gap:

| population | count | live |
|---|---|---|
| `planned`, never deployed | 18 | 404 |
| **`needs_rebuild`, never deployed** | **4** | **404** |
| `needs_rebuild`, deployed once | 34 | 200 |
| `deployed`, never stamped | 1 | 200 (`idea.uk`, a `/bugs_open/040` shape) |

**Four pages are `needs_rebuild` AND have never been deployed**, and candidate 1 would treat
all four as listable. One of them is `gaswholesalers.com/fuel-pricing-framework.html` —
`bugs_open/049`'s mechanism 2, live 404, linked from **28 live footers**. So the naive
predicate misses the worst confirmed real-world instance of this very class.

The predicate that tracks fetchability is **`deployed_at IS NULL`** (optionally with
`build_status <> 'deployed'` to exclude the single unstamped-but-deployed row). Discriminating
pair, same `build_status`, opposite live outcome:

```
gaswholesalers /fuel-pricing-framework.html  needs_rebuild  deployed_at NULL       -> 404
aao            /tools.html                   needs_rebuild  deployed_at 2026-05-02  -> 200
```

Validated by fetching every row in both populations, not inferred from the flag.
Full evidence in `/bugs_open/049`'s addendum (Correction 2). `[UNMEASURED]`: one exception,
`gamesdesign.co.uk /games/jelly-invaders/index.html`, is never-stamped yet serves 200 — a
different deploy path is the obvious guess and I did not chase it.

If candidate 2 (a shared "listable pages" helper) is built, this is the predicate it should
carry — and `/bugs_open/053` fix candidate 3 wants the same one for chrome nav, which is a
third derivation of the same question.

---

## STATUS 2026-07-26 (bugfix-052 session) — core FIXED & LIVE; residual committed, INERT

Two halves. The half this file was filed for is **fixed and live**. A second derivation,
found by finally running this file's own `[UNMEASURED]` survey, is **fixed in code and
inert until the next image roll** — so the case stays in `/bugs_open/` per the
`bugs_closed/README.md` bar.

### Half 1 — the tool-list derivation: FIXED, and LIVE since v1.0.1146

**This landed on 2026-07-21 and nobody updated this file, so for five days the case read
as untouched while the fix was already in production.** Commit `fe2ba5e52` (v1.0.1146)
added the shared predicate and wired it into the generic listing path:

- `queryresolve.FetchablePageEligibilitySQL` (`queryresolve.go:234`) —
  `AND (p.deployed_at IS NOT NULL OR p.build_status = 'deployed')`. This is the ADDENDUM's
  predicate, not fix candidate 1's; its doc comment cites this bug by number.
- `pageListEligibilitySQL(listedOnly)` (`queryresolve.go:245`) — the extraction point that
  makes "the generic listing path always has a build-state floor" unit-testable.
- Applied by `resolvePagesWhereType` (tool-list, game-list, guide-list, archetype-grid)
  and `resolvePagesUnderSection`.
- Pinned by `queryresolve/page_list_eligibility_test.go`.

**Verified live**, not inferred from the tag — pod `agent-chassis-f4d46c88d-p6wqc`,
image `v1.0.1165`, grepping a string the fix CREATED plus controls:

```
strings /app/agent-chassis | grep -c "deployed_at IS NOT NULL OR p.build_status"   -> 2
positive control:  grep -c "AND p.page_type = $2"                                 -> 2
negative control:  grep -c "zzz_not_a_real_symbol_052"                            -> 0
```

So this is **fix candidate 2**, the structural one — a shared "listable pages" helper —
not candidate 1. The file's own recommendation ("prefer (1) now") was overtaken by its
addendum before any code was written.

### Half 2 — the `[UNMEASURED]` survey: ANSWERED, and it found one more path

> The open question was: *"whether non-tool listings (article grids, news lists, card
> grids) share the same derivation and therefore the same gap — I only traced the tool
> list."*

**Answer: yes, exactly one — and it is worse than the original bug in one respect.**
Surveyed every page-set-derived listing in the tree:

| path | floor carried | verdict |
|---|---|---|
| `resolvePagesWhereType` (tool/game/guide/archetype) | `FetchablePageEligibilitySQL` | fixed v1.0.1146 |
| `resolvePagesUnderSection` | `FetchablePageEligibilitySQL` | fixed v1.0.1146 |
| `queryresolve` `blog_posts` source | `ListedPageEligibilitySQL` | already strict |
| `section_index_for.go` (hub "Browse All") | `status` only | resolves a hub URL, not a listing |
| news listings, directory trackers | n/a | derive from `content_feed_items` / `directory_entities`, **not** `pages` |
| **`rebuild_blog_listing_action.go`** | **none** | **the gap** |

`rebuild_blog_listing_action.go` bypasses `queryresolve` entirely and writes
`content_data.articles` + `rendered_html` directly. Its selection was:

```sql
WHERE p.site_id = $1 AND p.page_type = 'blog-post'
  AND p.build_status IN ('deployed', 'needs_rebuild')
```

Two defects:

1. **It is fix candidate 1's shape, the one the addendum measured as wrong.**
   `needs_rebuild` does not mean "still serves": four fleet pages are `needs_rebuild` and
   never deployed, and they 404.
2. **No `status` filter at all** — so `status='archived'` rows were listed anyway. This is
   the worse half, because archiving is the containment route this file recommends
   ("`pages.status='archived'` … which the listings demonstrably respect"). That claim was
   true of every listing path **except this one**. Containment did not hold here.

It is live machinery, not dormant: registered at `registry.go:686` and wired as a step in
the `rerender-pages` workflow — confirmed in `agent_definitions`, not just the seed file.

### Exposure — real defect, measured **zero** live bite today

Fleet-wide, 11 blog-post rows were mis-listed by the old predicate: 5 that would 404 and 6
`archived`. **All 11 are on `robot-hands.com`** — the same site, and the same R6 archiving
episode, that produced this bug in the first place.

But `findBlogPage()` resolves on only three sites fleet-wide, and none of them is
robot-hands:

| site with a resolvable blog page | posts listed | would 404 | archived-but-listed |
|---|---|---|---|
| `ai-agent-orchestration.com` | 14 | 0 | 0 |
| `finetuning.uk` | 16 | 0 | 0 |
| `leopardessconsulting.co.uk` | 6 | 0 | 0 |

robot-hands has no `blog-index` page and no `content` page named `blog`, so the action
skips it entirely. **The gap was latent: reproducible in code, not reachable by any live
site.** It would have become live the moment robot-hands got a blog page — which its own
archived `learning-center-index` (a `blog-listing` slot still holding 4 items) shows was a
plausible next step.

Recording that honestly matters more than the fix: this is the second time this file has
had to say the mechanism is real and the outage is not.

### The fix committed (INERT until the next image roll)

`platform/orchestration/actions/rebuild_blog_listing_action.go`:

- Blog-post selection lifted into a package-level `blogPostsQuery` const that **splices in
  `queryresolve.ListedPageEligibilitySQL`** — the same constant `queryresolve`'s own
  `blog_posts` source uses, plus `status IN ('active','deployed')`. Sharing the constant
  rather than restating it is the point: two derivations of the same article set that
  carry separate literals are the drift class `/bugs_closed/023` documents.
- `findBlogPage()` — both strategies gained `status IN ('active','deployed')`. They had
  none, so an archived `blog-index` page was a valid listing target, and Strategy 2 would
  additionally have **written** `page_type='blog-index'` onto an archived row.
- The empty-result branch now logs a **Warn** (not Info) when a listing component exists
  but no post is eligible: an empty set leaves the existing listing untouched, so it keeps
  advertising its old contents. Blanking it on a transient empty read would be worse, but
  it is not a no-op and it is exactly the stale-listing state this bug is about.

New test `rebuild_blog_listing_eligibility_test.go`, DB-free, mirroring
`queryresolve/page_list_eligibility_test.go`: asserts the fetchability floor, the status
filter, the shared-constant splice, and the `p` alias contract.

**The assertions were checked against the OLD query text, not just the new one** — all
three fire on it. A test that only passes on the fixed code proves nothing about whether
it would catch the regression.

### Regression + drop-set, measured against the live DB

Kept set is **unchanged** on all three real blog sites — nothing delisted:

```sql
SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.page_type='blog-post' AND p.status IN ('active','deployed')
  AND p.deployed_at IS NOT NULL
  AND jsonb_typeof(p.sections)='array' AND jsonb_array_length(p.sections)>0
  AND s.domain IN ('ai-agent-orchestration.com','finetuning.uk','leopardessconsulting.co.uk')
GROUP BY 1;
-- ai-agent-orchestration.com | 14
-- finetuning.uk              | 16
-- leopardessconsulting.co.uk |  6      (14/16/6 before AND after — 36 of 36 kept)
```

Drop set is exactly the intended six, all archived, all robot-hands:

```
grip-force-friction-calculator-guide  needs_rebuild archived  deployed_at NULL
gripper-cycle-time-estimator-guide    needs_rebuild archived  deployed_at NULL
gripper-payload-calculator-guide      needs_rebuild archived  deployed_at NULL
learning-center-article               needs_rebuild archived  2026-05-10   <- archived on purpose
learning-center-post                  needs_rebuild archived  deployed_at NULL
news-post                             needs_rebuild archived  deployed_at NULL
```

### Close-out checklist — what has to happen before this moves to `/bugs_closed/`

The Go half is inert until an image roll past this commit. Then:

1. Pod-grep a string the change creates, with a positive control:
   `strings /app/agent-chassis | grep -c "p.status IN ('active', 'deployed')"`
2. Re-run the two queries above; kept set must still be 14/16/6, drop set still those six.
3. Ideally exercise the failing branch rather than the happy path: give a scratch site a
   `blog-index` page plus one archived and one never-deployed `blog-post`, run
   `rebuild_blog_listing`, and assert neither appears. A green run on a site with no
   offending rows proves deployment, not correctness.
4. Then move this file to `/bugs_closed/` keeping its number.

### Notes for adjacent cases

- **`/bugs_open/053` candidate 3** deferred itself explicitly onto this bug: *"its correct
  predicate is `deployed_at IS NULL` … that predicate fix is `/bugs_open/052`; candidate 3
  should ride it, not this."* **The predicate now exists and is live** —
  `queryresolve.FetchablePageEligibilitySQL`. 053 can proceed. Deliberately NOT done here:
  `nav_tables.go`'s `deployedOnly` flag still emits `build_status = 'deployed'` in two
  places (`:188`, `:277`), and all six chrome call sites pass `deployedOnly=false`, so
  chrome applies no build-state filter at all today. Flipping that is a fleet-wide chrome
  change with 053's own verification list attached; it belongs to 053, not to this file.
- **Candidate 3 of this file** (emit a work item instead of silently dropping) remains
  unbuilt and still gated on `/bugs_open/033` — nothing consumes `needs_human_review`. The
  new Warn above is the cheap stand-in, not a substitute.
- `discovery_checks.neverDeployedPredicate` (`check_phantom_internal_links.go:366`) is an
  independent, unexported, negative-form copy of the same predicate. It matches today only
  because a NULL `build_status` makes the disjunct falsy either way. Two literals encoding
  one rule is the same drift risk this bug is about; worth consolidating, not urgent.
