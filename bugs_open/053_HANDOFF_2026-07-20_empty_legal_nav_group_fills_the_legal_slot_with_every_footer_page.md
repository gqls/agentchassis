# 053 — an empty `legal` nav group fills the footer's legal slot with every footer page

**Filed:** 2026-07-20 · **Branch:** `085_debug_and_feature_loops`
**Status:** CODE LIVE in prod **v1.0.1146** (candidate 1, `85d39f9b9`) — but **stays OPEN**: live
sites still serve the old footer because the output is a cached chrome artefact and nothing has
re-rendered them yet. Close only when a re-rendered site's `.footer-legal` is verified empty against
the live page.
**Council gate:** `SUBMISSION_CORR=550b9727-730b-44f9-8d37-5c56c2ce6615`, 3 rounds, all REVISE —
but **substantively approved** (see *Council review* below). R1 found a real gap (fixed, `309f519fc`);
R2 = 11/12 approve, R3 = only guardian (low). Stopped at R3 on owner ruling (2026-07-22): no
`Council-Reviewed:` trailer (earned by APPROVED only), the remaining objections are low
"independently-confirm-a-claim" nits from rotating seats and the fix is already live and correct.
**Pod-verified 2026-07-21:** `strings /app/agent-chassis | grep -c siteHasAnyNavItems` = 4 on
`agent-chassis-...-xrkv6` (image `v1.0.1146`); positive control `getNavItemsFromPagesFallback` = 6.
Commit 11:43:55 UTC < image start 12:15:20 UTC, so the build carries it.
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

## Fix applied — 2026-07-21 (candidate 1, commit `85d39f9b9`)

Chose **candidate 1**: it is the smallest correct change and it fixes every group type at once,
where candidate 2 papers over only `legal`. Candidate 3 was **deliberately deferred** — see below.

`GetNavItems` (`nav_tables.go`) no longer runs the pages fallback on any zero-row nav-table
result. It now consults a new gate, `siteHasAnyNavItems`, and falls back **only when the site has
no `site_nav_items` rows in any group**:

```go
items := getNavItemsFromTables(...)
if len(items) > 0 { return items }
if siteHasAnyNavItems(ctx, db, siteID, logger) {
    return []NavItem{}          // nav-table site, truthful empty answer for this group
}
return getNavItemsFromPagesFallback(...)   // pre-nav-table site, or tables absent
```

`siteHasAnyNavItems` runs `SELECT EXISTS(SELECT 1 FROM site_nav_items WHERE site_id = $1)`. If the
table itself does not exist (older deployments), the query errors on `does not exist` and the gate
returns `false`, preserving the pre-nav-table fallback exactly as before.

**Tested** — `nav_tables_fallback_test.go`, five sqlmock/observer cases mapping to the verify list:
1. nav-table site, 0 legal items → **0** links, and the fallback query is asserted *not* to run;
2. nav-table site *with* legal items → those items, gate never consulted (regression guard);
3. pre-nav-table site (0 rows) → pages fallback runs;
4. *(added round 2)* anomaly guard — a `primary`-including request that comes back empty on a
   nav-table site logs exactly one **Warn**; a lone `legal` empty logs **none** (loud vs quiet);
5. tables absent (`does not exist`) → pages fallback runs.
All pass (round 1 ran against `git archive HEAD` + these two files while an unrelated untracked WIP
file in `discovery_checks` broke the shared tree; that cleared by round 2 and the suite runs in-tree).

**Candidate 3 (deployedOnly for chrome nav) NOT applied, on purpose.** With candidate 1 in place the
gaswholesalers legal-slot 404 (`/fuel-pricing-framework.html`) disappears anyway — that site has 18
nav items and 0 legal, so its legal fallback no longer runs at all. Candidate 3's remaining value is
the header/quick-links path and genuine pre-nav-table sites, but its *correct* predicate is
`deployed_at IS NULL`, **not** the `build_status = 'deployed'` that the current `deployedOnly` flag
emits — 34 fleet pages are `needs_rebuild` and serve 200 fine, so flipping the flag on today would
**drop valid links**. That predicate fix is `/bugs_open/052`; candidate 3 should ride it, not this.

**Corrected fleet figures (re-grounded 2026-07-21).** The counts in *Fleet exposure* have moved since
filing — nav rows are live state. Today, sites with **active legal nav rows** are
`leopardessconsulting.co.uk` (2), `finetuning.uk` (2), `ai-agent-orchestration.com` (2) and
`idea.uk` (1) — not "leopardess (6) and finetuning (1)". robot-hands still has **15** nav items and
**0** legal, and its legal fallback still returns **14** footer pages today (reproduced live). The
*mechanism* is unchanged; only the per-site tallies drifted. Every real live site now has
`primary` nav rows, so no real site is a pre-nav-table site — the fallback branch survives only for
newly-created sites before `PopulateNavTablesAction` runs (verify list #3: the branch is not dead,
just not exercised by any current live domain).

## Council review — 3 rounds, all REVISE, substantively approved; stopped by owner ruling

`SUBMISSION_CORR=550b9727…`. **Round 1 earned a real fix; rounds 2–3 were single low-severity
"independently confirm a claim" nitpicks from rotating seats.** The gate treats *any* objection —
even `low` — as REVISE, so a sound change with one nit never reaches a unanimous APPROVED. Owner
ruling 2026-07-22: **stop at round 3, apply no `Council-Reviewed:` trailer** (the trailer is earned
by an APPROVED verdict only — [[council-reviewed-trailer-discipline]]), because the fix is already
live and correct and every substantive concern is resolved.

- **Round 2 → REVISE, 11/12 approve.** The fail-loud guard flipped `bug_historian`, `guardian`,
  `reuse_agent` and `prior_art_librarian` to *approve*. Sole objection: `edit-quality` (low) — the
  *sketch* used `containsGroupType` without evidence it pre-exists; "pending that check, verdict
  would move to approve". (It does pre-exist, `nav_tables.go:346`; sketch-evidence gap, not a code
  defect. The runbook trap: reviewers judge the sketch, not the file.)
- **Round 3 → REVISE.** Added the `containsGroupType` provenance; `edit-quality` cleared. Sole
  objection: `guardian` (low) — wanted the blast-radius claim *independently* confirmed, specifically
  that `GetNavItems` is a render-time leaf and not reached from work-item dispatch / agent-spawning.
  **Answered definitively:** all seven callers are render/site-assembly actions
  (`RenderSiteComponentsAction`, `RenderFooter`, `BuildRenderContextAction`, `RerenderSitePagesAction`,
  `LoadSiteForRebuildAction`, `SyncPagesToDBAction`, `buildRenderContextFromDB`); a grep of
  `coordinator*/dispatch*/agentbase/internal/` finds **zero** callers. Blast radius is closed.

### Round 1 in detail — REVISE, and the fail-loud guard it earned (`309f519fc`)

Verdict **REVISE**, decided by `bug_historian`; **6 of 10 seats approved** (compliance, render_guardian,
debug_historian, constitution, mission, edit-quality-modulo-one-nit). The objections were worth the run:

- **`bug_historian` (medium, the decider) + `guardian` (medium) + `edit-quality` (low)** — candidate 1
  *respected* the truthful empty answer but returned empty for **every** group with only a `Debug`
  log. That is a **fresh instance of the very silent-empty-render class this bug is about** (016b §9):
  if a sync/write bug ever dropped a site's `primary` nav, `GetNavItems` would serve an empty menu
  silently, indistinguishable from the legitimate empty-`legal` case. **Fix:** the empty-return path
  now logs **`Warn`** when the request *includes the `primary` group* (never legitimately empty on a
  nav-table site) and stays `Debug` for a lone `legal`/`utility`/`content` group. Near-zero cost — no
  live site has empty primary nav today, so it never fires until a real anomaly. A `site_work_item`
  was the heavier option; the `Warn` is the near-zero-cost first step `bug_historian` itself proposed.
- **`guardian` (blast radius)** — "the fix generalises to all callers; prove it." Answered from the
  codebase, not asserted: every `GetNavItems` caller passes either the `primary` group or a lone
  `[legal]` (enumerated), so only the legal caller is legitimately empty on a nav-table site — and a
  `primary`-including caller returning empty is exactly the anomaly now made loud.
- **`reuse_agent` (low) + `prior_art_librarian` (b)** — "show you checked for an existing helper."
  Grepped `platform/ internal/ pkg/`: no pre-existing "site has any nav rows" helper (the other
  `site_nav_items` queries are rename-tool / create-tool / orphan-pages), and exactly one definition
  of `siteHasAnyNavItems`. Genuinely new, not a dormant duplicate.
- **`prior_art_librarian` (HIGH)** — flagged the risks note "confirmed present in production v1.0.1146"
  as contradicting edit-2's `add` operation (the `bugs_closed/031` false-already-deployed shape).
  **Both are true and not contradictory:** this thread commits per task and the fleet builds from
  HEAD, so the *new* function was committed and shipped in v1.0.1146 **before** this advisory review.
  The pod-grep is real (count 4, positive control 6), not a copied trail. Clarified in the round-2
  rationale rather than repeated as a bare claim.

Resubmitted round 2 on the same correlation; the code is already committed and live, so each verdict
is recorded for the trail, not as a ship gate (see *Council review* summary above for R2/R3).

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
