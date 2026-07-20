# 053 — an empty `legal` nav group fills the footer's legal slot with every footer page

**Filed:** 2026-07-20 · **Branch:** `085_debug_and_feature_loops` · **Status:** OPEN, not started
**Severity:** medium — cosmetically wrong on every page of at least 6 live sites; not a 404 in
itself, but it silently *reintroduces* whatever broken links live in the footer page set.
**Class:** structural — a "no rows" result overloaded to mean two different things.
**Found by:** the bugfix-049 session, while measuring what a chrome re-render would emit before
recommending one. Filed separately from `/bugs_open/049` because it is **fleet-wide and
independent of chrome staleness** — 049's three sites are stale, these sites are current.

---

## Symptom

`robot-hands.com` has post-fix chrome (rendered 2026-07-18) and its footer's legal area is:

```html
<div class="footer-legal">
  <a href="/gripper-catalog.html">Gripper Catalog</a><a href="/matchmatrix.html">MatchMatrix</a>
  <a href="/tools.html">Tools</a><a href="/selection-guide.html">Selection Guide</a>
  <a href="/about.html">About</a><a href="/contact.html">Contact</a>
  <a href="/news/index.html">News</a><a href="/gripper-catalog/index.html">Gripper Catalog</a>
  <a href="/matchmatrix-methodology.html">Matchmatrix Methodology</a><a href="/news.html">News</a>
  <a href="/gripper-selection-guide.html">Selection Guide</a><a href="/how-it-works.html">How It Works</a>
  <a href="/learning-center.html">Learning Center</a><a href="/services.html">Services</a>
</div>
```

**Fourteen links, none of them legal, three of them duplicated** (Gripper Catalog, News,
Selection Guide each appear twice under two URLs). All 14 return 200, so this is not a broken
link — it is the whole footer navigation rendered a second time in the legal position.

## Root cause

`RenderSiteComponentsAction` asks for the legal links with:

```go
// render_site_components_action.go:188
legalNavItems := GetNavItems(ctx, params.DB, siteID, []string{NavGroupLegal}, false, 0, params.Logger)
```

`GetNavItems` (`nav_tables.go:65-75`) returns the nav-table rows **only if there are any**, and
otherwise falls through to the pages table:

```go
items := getNavItemsFromTables(ctx, db, siteID, groupTypes, deployedOnly, maxItems, logger)
if len(items) > 0 {
    return items
}
// No nav table entries for this site — fall back to pages table
return getNavItemsFromPagesFallback(ctx, db, siteID, groupTypes, deployedOnly, maxItems, logger)
```

**A zero-row result is overloaded.** It can mean *"this site predates the nav tables"* (the
backward-compatibility case the fallback exists for) or *"this site has nav tables and
genuinely has no legal pages"* — which is the **correct, expected** answer for most sites. The
fallback cannot tell them apart, so it treats a truthful empty answer as a missing table.

The fallback then runs its footer branch (`isHeaderOnly=false`, `includesLegal=true`):

```sql
(in_footer = true OR LOWER(name) IN ('privacy','terms','cookies','disclaimer'))
AND name NOT IN ('index','404','sitemap') AND status IN ('deployed','active')
```

The `in_footer` disjunct dominates the legal-name disjunct, so **every footer page qualifies**.
The legal-name list is only ever additive; nothing constrains the result *to* legal pages.

## Evidence — mechanism confirmed, alternative refuted

The competing explanation is that the footer template renders `quickLinksItems`
(`primary`+`utility`) into `.footer-legal` — a template bug rather than a `GetNavItems` bug.
It is ruled out by count and by order:

| hypothesis | predicted links | actual in `.footer-legal` |
|---|---|---|
| pages fallback for `legal` | **14** | **14** ✓ |
| `quickLinksItems` (primary+utility) | 15 | ✗ |

The fallback query's 14 rows match the live markup **one-for-one, in the same order**,
including both duplicate pairs. Reproduce:

```sql
-- returns 14, matching the live footer exactly
SELECT COALESCE(p.nav_label,p.title,p.name), COALESCE(p.url,'/'||p.name||'.html')
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='robot-hands.com'
  AND (p.in_footer=true OR LOWER(p.name) IN ('privacy','terms','cookies','disclaimer'))
  AND p.name NOT IN ('index','404','sitemap') AND p.status IN ('deployed','active')
ORDER BY CASE WHEN LOWER(p.name) IN ('privacy','terms','cookies','disclaimer') THEN 1 ELSE 0 END,
         COALESCE(p.nav_order,99),
         CASE p.name WHEN 'services' THEN 1 WHEN 'about' THEN 2 WHEN 'contact' THEN 3
                     WHEN 'privacy' THEN 8 WHEN 'terms' THEN 9 ELSE 5 END;
```

## Fleet exposure

```sql
SELECT s.domain, count(*) FILTER (WHERE ng.group_type='legal' AND ni.status='active') AS legal_items
FROM sites s LEFT JOIN site_nav_items ni ON ni.site_id=s.id
LEFT JOIN site_nav_groups ng ON ni.group_id=ng.id GROUP BY 1 ORDER BY 2, 1;
```

**Only `leopardessconsulting.co.uk` (6) and `finetuning.uk` (1) have legal nav rows.** Every
other site has zero and therefore takes the fallback: robot-hands.com, dartsonline.com,
relojistas.com, vetcomparison.uk, vonc.com (all with current chrome, so all serving this
today), plus ai-agent-orchestration.com and gaswholesalers.com (stale chrome, so they will
start serving it the moment 049's re-render runs).

> **This is why `/bugs_open/049`'s two-directional control is weaker than it reads.** That file
> concludes the 2026-06-10 legal-links fix "works and has simply never run for three sites",
> from the observation that post-fix sites look correct. **leopardess is the only post-fix site
> that actually exercises the fixed path** — it is the only one with legal nav rows. The others
> look fine only because their fallback output happens to contain no 404s.
> `NOTES_cta_link_integrity.md`'s *"robot-hands' chrome emits no legal links at all"* is
> **wrong**: it emits fourteen, none legal. Corrected in 049's addendum.

## Why it matters beyond cosmetics

The fallback calls `GetNavItems` with `deployedOnly=false`, so it emits footer pages
**regardless of build state**. On `gaswholesalers.com` that includes
`/fuel-pricing-framework.html` — `deployed_at IS NULL`, live **404**, and already filed as
049's mechanism 2. So a chrome re-render there would delete two phantom legal links (56 anchor
instances) and **introduce a different broken one on all 28 pages**. The legal slot becomes a
new, unaudited surface for exactly the class 049 is about.

## Fix candidates (none applied)

1. **Distinguish "no rows for this group" from "no nav tables at all."** The fallback exists for
   pre-nav-table sites; gate it on that, not on the per-group row count — e.g. fall back only
   when the site has **no** `site_nav_items` rows in *any* group. Smallest correct change, and it
   fixes every group type at once, not just `legal`. A site with 45 active nav items answering
   "no legal pages" is giving a true answer that must be respected.
2. **Constrain the fallback's legal branch to legal pages.** When `includesLegal` is the *only*
   requested group, the query should match the legal-name list **AND NOT** simply `in_footer`.
   Narrower, but leaves the overloaded-empty-result bug in place for other group types.
3. **Pass `deployedOnly=true` for chrome nav.** Independent of the above and worth doing anyway:
   chrome should never link a page that has never been deployed. Note the predicate has to be
   `deployed_at IS NULL`, not `build_status <> 'deployed'` — see 049's Correction 2 and
   `/bugs_open/052`; 34 fleet pages are `needs_rebuild` and serve 200 perfectly well.

Candidates 1 and 3 are independent and compose. 1 without 3 still leaves undeployed pages
linkable from the header/quick-links path.

## How to verify a fix

1. `robot-hands.com`'s `.footer-legal` renders **zero** links after a chrome re-render (it has
   no legal pages, so the honest answer is empty).
2. `leopardessconsulting.co.uk` still renders its **6** real legal links — the regression guard;
   candidate 1 must not break the nav-tables path.
3. A genuinely pre-nav-table site (no `site_nav_items` rows at all) still gets its fallback nav.
   Confirm one exists before assuming the branch is dead — if none does, say so rather than
   silently deleting a branch nobody can exercise.
4. Grep the pod for the changed predicate, not a string the change merely uses.

## Landmines

- **`site_components.rendered_html` is a rendered artefact.** Editing it is undone by the next
  chrome render. Same landmine as 049 and the travelling-docs runtime-fill templates.
- **Chrome only re-renders on explicit trigger** (049 mechanism 1), so any fix here is inert on
  live sites until something asks for a render. Do not report this fixed on the strength of a
  code change — verify against the live footer.
- The duplicate pairs in the live output (`/news/index.html` **and** `/news.html`) are a
  separate smell: two `pages` rows describing one destination. Not chased here.

## Related

- `/bugs_open/049` — stale chrome + unbuilt-page links. Found while measuring 049; its
  Correction 1 records this, and its per-site re-render table depends on it.
- `/bugs_open/052` — a listing derives from the page set with no build-state filter. Same
  family as fix candidate 3: a derived surface that does not consult build state.
- `/bugs_open/018` — idea.uk chrome renders every link `href=""`; established that the chrome
  renderer fills from a hardcoded vocabulary and never reads `input_schema`. Third instance of
  "chrome is under-modelled".
- `/bugs_open/023` — derived fields recomputed on every render, so authored edits cannot hold.
