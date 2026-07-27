# 098 — archiving a page removes it from every derivation but not from the deployed site, so its frozen listing keeps advertising a 404

**Found:** 2026-07-26, while closing `bugs_closed/052`. Not a residual of 052's
mechanism — 052's derivation defect is fixed and live. This is the *opposite*
mechanism, and 052's fix cannot reach it by construction.

**Severity:** low today (one live instance, on an orphan URL, no sitemap), but the
class is permanent: nothing in the platform can ever repair the affected page,
because every repair path skips archived rows on purpose.

## The gap in one line

`pages.status = 'archived'` correctly removes a page from re-derivation *and* from
re-rendering — so the last HTML that page ever rendered is frozen, and it **keeps
being served**. Archiving retracts a page from the platform's model of the site; it
does not retract it from the site.

## Live instance (verified 2026-07-26 21:14 UTC)

`robot-hands.com`, page `learning-center-index`:

| field | value |
|---|---|
| `url` | `/learning-center/index.html` |
| `status` | `archived` |
| `deployed_at` | stamped |
| component | `b266955f…`, slot `blog-listing`, 4 articles, `updated_at` **2026-07-03** |

That URL serves **HTTP 200, 18,545 bytes** today. Its frozen listing links to
`/blog/learning-center-article.html`, which serves **HTTP/2 404**:

```
$ curl -sS -o /dev/null -w '%{http_code}' https://robot-hands.com/learning-center/index.html
200
$ curl -sS -o /dev/null -D - https://robot-hands.com/blog/learning-center-article.html
HTTP/2 404
```

Verified against the live HTML, not the DB row — the fetched page contains
`href="/blog/learning-center-article.html"`.

So the visible symptom is **exactly 052's** ("a listing advertises a page that
404s") while the cause is not 052's at all. The live listing on the same site —
`learning-center-hub`, the page that is *not* archived — carries 3 articles and is
correct. Only the archived one is wrong, and only because it stopped being
maintained.

## Why 052's fix cannot repair this

052 fixed the *derivation*: `rebuild_blog_listing_action` and the `queryresolve`
listing path now both exclude archived and never-deployed pages. Correct, live,
verified. But a derivation only rewrites a listing **when the page carrying it is
re-rendered**, and an archived page is never re-rendered — that exclusion is
deliberate and is what makes archiving useful as containment everywhere else.

The two facts compose into a trap:

- archiving a *post* stops it being listed **in future renders**;
- archiving a *listing page* stops those future renders happening at all.

Archive both — which is what a site cleanup does — and the stale listing is sealed
in at whatever it said the moment before. `bugs_closed/052` explicitly relies on
archiving as containment route R6. R6 works on the source row and freezes the
artefact; that qualification was not written down anywhere before this file.

## A second finding, and it is the more transferable one

**`deployed_at IS NOT NULL` does not mean "currently fetchable". It means "a deploy
happened once".** The 404 target above has `deployed_at = 2026-05-10 18:58 UTC`
stamped and has been archived since ~2026-07-17. Nothing clears the stamp.

This matters because `deployed_at IS NOT NULL` is the load-bearing half of the
shared eligibility predicates that 052 introduced fleet-wide
(`queryresolve.ListedPageEligibilitySQL`, `DeployedPageEligibilitySQL`,
`FetchablePageEligibilitySQL`). Those predicates are still right — they pair it with
a `status` filter, which is what actually excludes this row — but the *reason* they
are right is not the reason their doc comments give. Any future consumer that reaches
for `deployed_at IS NOT NULL` **alone** as a fetchability test will admit archived
pages that 404. `DeployedPageEligibilitySQL` is exactly that constant, and its doc
comment currently describes it as a spend guard whose false negatives "self-correct
when `deployed_at` is stamped" — the false *positive* documented here is not mentioned.

## Scale — measured 2026-07-26, live DB

```sql
SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.status='archived' AND p.deployed_at IS NOT NULL GROUP BY 1;
-- leopardessconsulting.co.uk | 10
-- robot-hands.com            |  2
```

12 archived-but-once-deployed pages fleet-wide. Of those, **exactly one** carries a
rendered listing artefact (the `articles` array shape) — the robot-hands row above.
The other 11 are ordinary content pages: they may still serve stale content, but they
do not *advertise* anything, so they are a lesser problem.

Containment on the live instance is limited: `/learning-center/index.html` is **not
linked from the live navigation** (the homepage and `/learning-center.html` both link
to `/learning-center-hub.html` instead), and `robot-hands.com/sitemap.xml` returns
**zero `<loc>` entries**. So it is reachable by direct URL, an external link, or a
search engine's memory — not by a visitor browsing the site. That is why this is filed
rather than escalated.

## Fix candidates — ordered by what makes the bad state unrepresentable

1. **Make the deploy reconciling rather than additive.** If a site deploy emitted the
   full intended file set and removed files not in it, an archived page would vanish
   from the site at the next deploy of that site, and "archived but serving" would
   stop being representable. Biggest blast radius, needs its own survey (what else
   currently relies on files persisting across deploys?), but it is the only candidate
   that closes the door rather than narrowing it.
2. **Clear `deployed_at` when a page is archived**, and add the retraction to whatever
   performs the archiving. This makes the column mean what every consumer already
   assumes it means, and it is the fix that protects the predicate family from the
   next consumer who uses `deployed_at IS NOT NULL` alone. Note it is a *semantic*
   change: any query using `deployed_at` as history rather than current state must be
   found first — `grep` the column, not its callers (the method that found the last
   unmigrated copy in 052).
3. **Retract on archive**: an explicit un-deploy step (delete the file, or serve 410)
   fired when a page is archived. Narrower than 1, but leaves every *already*-archived
   page unfixed unless backfilled.
4. **Detect only**: a sweep that curls the URL of every `status='archived' AND
   deployed_at IS NOT NULL` page and reports the 200s. Cheapest, closes nothing, but
   it would have surfaced this instance without anyone reading the code — and today
   nothing does. 12 URLs; this is a small check.

## How to verify a fix

The failing branch is directly reproducible with no setup, because a live instance
exists:

```sh
curl -sS -o /dev/null -w '%{http_code}\n' https://robot-hands.com/learning-center/index.html
# today: 200. After a fix of class 1 or 3: 404 or 410.
```

For candidate 2, the check is the column's meaning, not one row:

```sql
SELECT count(*) FROM pages WHERE status='archived' AND deployed_at IS NOT NULL;
-- today: 12. After the fix + backfill: 0.
```

Do not accept a green run on a site with no archived pages as evidence — that proves
deployment, not correctness.

## Related

- `bugs_closed/052` — the derivation half, fixed and live. This file is the finding
  that closing it turned up, and 052's exposure section is corrected to point here.
- `bugs_open/081` — a *deployed* but mistyped page has no repair path. Same shape at a
  higher level: the repair machinery's guards each exclude a population deliberately,
  and the intersection of those exclusions is a set nothing owns. 081 is `build_status
  = 'deployed'`; this is `status = 'archived'`.
- `bugs_open/097` — CTA integrity misses in-body card links to unbuilt pages. A
  detector of the class in candidate 4 would need to cover both.
- `bugs_open/071` — the validate gate detects every broken link then discards the
  finding. Relevant to candidate 4: detection may already exist and be thrown away.

## Second surface, found 2026-07-27 on relojistas.com — archiving also leaves the NAV row live

This case says archiving retracts a page from the platform's model but not from the
site. There is a **second thing it does not retract, and it is more visible than the
frozen page itself: the navigation entry.**

Retiring relojistas' `contacto` page (owner ruling: the site carries no contact route)
went `pages.status='archived'`, `in_header=false`, `in_footer=false`, then the file
deleted from `vm-sites` — so 098's frozen-artefact problem was handled deliberately and
`/contacto.html` became a correct **404**.

**Every page then advertised that 404 from its own footer.** After a successful
re-render at 14:29:57 — *five minutes after* the archive — the homepage still carried
three `href="/contacto.html"` links.

The nav is not derived from `pages`. `RerenderSinglePageAction` calls `GetNavItems`
(`nav_tables.go:86`), which reads **`site_nav_items` / `site_nav_groups`** and only falls
back to a `pages`-derived query for sites predating those tables (the fallback
`bugs_closed/053` narrowed). The live filter is `ni.status = 'active'`. Nothing in the
archive path writes that table:

```
group_type | label    | url            | status
utility    | Contacto | /contacto.html | active     <- still active after the page was archived
```

**Why this belongs here rather than in its own file:** it is the same defect one layer
up. 098 is *"the page keeps serving"*; this is *"the nav keeps pointing"*. Same cause —
archiving is modelled as a `pages` state change, while every other surface holding a copy
of the page's existence (the deployed file, the nav table) is untouched. A fix that
reaches only the deployed artefact will still leave the nav dead-linking, and a reader of
this case would reasonably assume otherwise.

**It also inverts this case's severity note.** Filed here as "low today — one live
instance, on an orphan URL, no sitemap". A stranded *nav* entry is the opposite: it sits
in the chrome of **every page on the site** — the most-crawled, most-clicked position
there is. On relojistas that was 18 pages advertising one 404.

**For candidate 4 (a detector):** cheap, and needs no crawl —

```sql
SELECT g.site_id, i.label, i.url
FROM site_nav_items i
JOIN site_nav_groups g ON g.id = i.group_id
LEFT JOIN pages p ON p.id = i.page_id
WHERE i.status = 'active'
  AND (p.id IS NULL OR p.status <> 'active');
```

Anything returned is a nav entry pointing at a page the platform has retracted.
**[UNMEASURED] whether other sites carry stranded nav rows** — I checked only this site;
worth running fleet-wide before designing the fix.

> ## MEASURED 2026-07-27 (triage sweep) — run fleet-wide, and **the query as written is wrong**
>
> Fleet-wide it returns **4 rows, all on `leopardessconsulting.co.uk`, and all four are
> false positives.** No other site carries a stranded nav row.
>
> | label | url | `page_id` | live |
> |---|---|---|---|
> | Agent Complexity Estimator | `/tools/tool-agent-complexity-estimator.html` | **NULL** | 200 |
> | Automation Suitability Scorer | `/tools/process-automation-scorer/index.html` | **NULL** | 200 |
> | Password Entropy | `/tools/password-entropy.html` | **NULL** | 200 |
> | ROI Estimator | `/tools/ai-agent-roi-estimator.html` | **NULL** | 200 |
>
> All four have `page_id IS NULL` — nav rows authored by URL with no FK — and every one
> of their URLs resolves to an `active` page and serves **HTTP 200**. The `LEFT JOIN …
> WHERE p.id IS NULL` arm cannot tell *"points at a retracted page"* from *"has no page
> FK at all"*, and the second is the common case: a nav entry created by URL is normal.
>
> **A detector built on the query above would open 4 work items against healthy nav on
> day one and find nothing real** — the failure mode that `bugs_open/033` and `/071`
> already describe (detection nobody can act on). Fall back to the URL when the FK is
> absent:
>
> ```sql
> SELECT s.domain, i.label, i.url
> FROM site_nav_items i
> JOIN site_nav_groups g ON g.id = i.group_id
> JOIN sites s ON s.id = g.site_id
> LEFT JOIN pages p ON p.id = i.page_id
> LEFT JOIN pages q ON q.site_id = g.site_id AND q.url = i.url
> WHERE i.status = 'active'
>   AND COALESCE(p.status, q.status) IS DISTINCT FROM 'active';
> ```
>
> That returns **0 rows** today — i.e. the relojistas repair holds and nothing else in
> the fleet is stranded. Which means the *nav* half of this case has **no live instance
> left**; only the frozen-artefact half does (below).
>
> **The frozen-artefact half is still live and still reproducible**, re-checked
> 2026-07-27:
>
> ```
> 200  https://robot-hands.com/learning-center/index.html
> ```
>
> The archived-but-once-deployed population has grown 12 → **13** (leopardess 10,
> robot-hands 2, relojistas 1 — the newly archived `contacto`, whose file was deleted
> deliberately, so it correctly 404s and is not a new instance).

Repaired here by setting the row `inactive` (backup `bak_reloj_navitem_contacto_20260727`)
and re-firing every page. **Convention trap:** the 4 pre-existing non-live rows fleet-wide
use `inactive`; `archived` is also excluded by the `= 'active'` filter but would be
invisible to anyone querying by convention. Evidence, and the wrong turn that found it:
`traffic_probe/relojistas_rebuild_running_notes.md`, 2026-07-27.
