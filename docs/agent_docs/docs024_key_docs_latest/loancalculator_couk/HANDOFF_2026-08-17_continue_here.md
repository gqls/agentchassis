# HANDOFF — loancalculator.co.uk · **282 PROVEN, 0/11 → 11/11**; the rebuild wave is queued behind `bugs_open/243` (2026-08-17 ~12:05Z)

> ## ⚠ SUPERSEDED by `HANDOFF_2026-08-17b_continue_here.md` — and it contains ONE WRONG CLAIM, kept here deliberately
>
> **Step 2 of "FIRST STEP" below says the 14 guide builds are "stamp-convergence
> churn, not damage". THAT IS FALSE.** Those 15 `needs_page` items were builds for
> **14 NEW `blog-post` pages at `/blog/<slug>.html`** that this re-fire's plan
> invented while dropping all 14 real `guide` pages from the plan. The gate re-opened
> at 12:10:17Z and **all 14 duplicates built and deployed**; they now serve alongside
> the guides they duplicate. Containment was attempted at 12:12 and refused by the
> permission classifier.
>
> The rest of this file — the 282 proof, the 11/11 measurement, the judge query, the
> toolgolden result, the nav findings — is accurate and still worth reading. The
> homepage rebuild it recommends in step 3 **ran and succeeded**: calculator at
> position 2, locks 12/12, toolgolden exit 0. See NOTES §9 for the correction and
> what caught it, and 17b for current state.

> Supersedes `HANDOFF_2026-08-15_fire_in_flight_continue_here.md`. Its two owed
> steps are both DONE (toolgolden, and the D2 re-fire sequence). Evidence and every
> misstep: NOTES `## 2026-08-17`. Owner prose: README_where_we_are, same date.

```
site      loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis   v1.0.1305, stamp 6a782274b (verified at the binary, positive control run)
          — carries bugs_open/282 AND LOCK-008 rounds 1+2 (7d9b7334a, 57336c127)
plan      9463e31d-ee50-482e-94a9-7e186ef25543  is_current, landed 11:46:22Z
          (baseline for any future compare; dcbae4df is the 0/11 pre-fix one)
locks     12/12 held, untouched through the re-fire
pages     28/29 serving 200. guides-index (/guides/index.html) is the only 404,
          still build_status='planned'. tool-standard-calc stays ARCHIVED.
state     THE LANE'S CENTRAL BLOCKER IS DISCHARGED. Plan-level placement of the
          calculators is CORRECT AND PROVEN. The pages have NOT been rebuilt from
          it — the fleet claim gate is shut (bugs_open/243), due to self-clear
          12:09:53Z.
```

## What was proved, and how (do not re-derive — re-read)

**`bugs_open/282` is not merely live, it is proven on the motivating case.** Same
script, same 12-page scope, same planner, one image later: **0 of 11 → 11 of 11**
locked calculators placed on their own pages. Verified as a JOIN of the locked rows'
`content_components.function` against `site_plan_sections`, per page — not by eye.
Zero `RECOMPOSE_INTENT_NOT_REALISED` rows; locks 12/12. Fire corr
`3584d962-d3de-415b-a468-64afab126534`, 11:43:34Z.

⚠ **The script's own judge query #4 CANNOT show this and must not be used.** It
counts `component_name LIKE 'tool-%'`, which also matches the `tool-cta` and
`tool-list` SECTION components: it returned **26** on the 0/11 baseline and returns
26 after. Use this instead:

```sql
WITH locked AS (
  SELECT p.name AS page, cc.function AS fn FROM page_components pc
  JOIN pages p ON pc.page_id=p.id JOIN content_components cc ON cc.id=pc.component_id
  WHERE p.site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
    AND pc.locked_at IS NOT NULL AND p.status='active')
SELECT l.page, l.fn, EXISTS(SELECT 1 FROM site_plans sp JOIN site_plan_sections sps
  ON sps.plan_id=sp.id WHERE sp.site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
  AND sp.is_current AND sps.page_name=l.page AND sps.component_name=l.fn) AS placed
FROM locked l ORDER BY placed, l.page;
```

**Toolgolden is done: 11/11 on arithmetic.** `--compare` exits 1 and prints "11 of
11 diverged"; that headline is worthless here. 0 of **1,340** shared fields changed
value, and `drove` is identical in every vector (so the pages' own numeric defaults
are untouched — the thing that would silently invalidate the comparison). Every
difference is chrome: 2 ids gone (`nav-links-menu`, `mobile-menu-btn`), 1 arrived
(the FAQ block). **The golden the old handoff named was the wrong file** — see its
corrected §1. Current baseline:
`acceptance/GOLDEN_2026-08-17_post_rebuild_tool_values.json`.

## FIRST STEP — confirm the gate opened, then judge the wave it releases

The 21 items minted by the re-fire (15 `needs_page` = 14 guides + index, 5
`needs_imagery`, 1 `needs_rerender`) were all born `triaged` and none could be
claimed: `claim_work_item` gates on `ai_endpoint_health`, whose `claude` row went
`healthy=f` at 11:09:53Z on a spend-cap 400 and is re-probed only hourly. Full
recipe, including the query that tells you WHEN it re-opens: RUNBOOK §"My work item
won't dispatch". This is `bugs_open/243`, known, three other lanes hit it today.

1. `SELECT healthy, last_checked FROM ai_endpoint_health WHERE name='claude';`
   Expect healthy `t` after 12:09:53Z. If still `f`, read the `error` column — and
   do NOT conclude "no dispatch until 2026-09-01", which is the trap the LANDMINES
   correction of 11:55Z exists to stop.
2. Let the wave drain. The 14 guides are **stamp-convergence churn, not damage**
   (same as 08-15): they were rebuilt at plan `34b1`/`dcbae4df`, the new plan says
   the same sections, so the version-compare re-emits; this build re-stamps at
   `9463e31d` and it ends.
3. **`needs_page:index` is the one to actually watch.** It rebuilds the homepage,
   which carries locked `tool-loan-repayment` at position 6. Precedent says the
   floor holds: index was auto-rebuilt 2026-08-16 08:45 by `build-dispatch-loop` on
   an image WITHOUT the 282 fix and the calculator survived (locks 12/12, and
   toolgolden finds index byte-identical today). This rebuild is better informed —
   the calculator is now IN the plan. After it lands, verify at the artefact:
   `toolgolden.py --compare acceptance/GOLDEN_2026-08-17_post_rebuild_tool_values.json <the 11 urls>`
   (URLs from `pages.url`, never name-derived) and re-check locks 12/12.

## THEN: the 11 tool-page tickets, which are the owner's call

The 11 `owned_page_review` items from 08-15 are still open at `needs_human_review`.
They were held pending 282; **that reason is now discharged** and the plan they
would realise is the correct one. They remain a human gate by design (TP-004: the
generic builder clobbers tool pages), so they are the owner's to release, not a
session's. The reconciler skipped-as-queued this run, so no duplicates were minted.

## OWNER DECISIONS OPEN

- **D-NAV (new, and the most visible one).** The rebuilt header is `Home` + `About`
  + a "Get Started" CTA only. The old hand-built site had a `Tools` dropdown listing
  nine calculators (recorded in the 08-08 golden). This is the framework working as
  designed: `classifyPagesForNav` bars `page_type='tool'` AND any `/tools/` URL from
  the primary nav, expecting **a parent listing page** to represent them — and this
  site has one for Guides but **none for Tools**. Calculators stay reachable via
  in-body cross-links (8 per page), so this is degraded navigation, not orphaning.
  Question: create a `tools-index` page the same shape as `guides-index`? Not done.
- **D-GUIDES.** `guides-index` is the last 404. It is NOT barred from the nav —
  `page_type='section-index'` is explicitly exempted, and `site_nav_items` already
  carries `Guides → /guides/index.html` at primary position 2 (written 11:46). So
  building the page should restore the menu entry. It is the same 282-family path as
  index was; with 282 live it should now build.
- **D-ROADMAP (small).** `tool-credit-roadmap` was given
  `tool-credit-health-check` — another page's calculator, and the second copy of that
  function in the plan. It has no locked row of its own so nothing was displaced, but
  a duplicated calculator across two pages is a content decision. Left alone.
- **D3 (carried from 08-15).** `needs_design`, `needs_composition`, 2
  `needs_brand_head_assets` and 3 imagery items are still owner-parked and
  dedup-block fresh mints; the released chrome/css will not regenerate until they
  free up. Un-parking reverses the owner's own 08-12 decision. **If you un-park
  anything, set `status='triaged'`, never `'detected'`** — the dispatcher cannot see
  `detected` and it fails silently (this cost 3 hours on 08-15).
- **D4 (housekeeping).** 19 `lock_blocked_change` + 6 `content_rewrite` at
  needs_human_review, plus the parked set; mostly superseded by the rebuild. Sweep
  once the wave settles.
- **RE-LOCK judgement (carried).** Whether the 8 released locks (chrome 3 + css
  carriers 4 + homepage prose-0) are re-armed is an owner call — the rebuilt
  chrome/css may not need carriers at all.

## Standing cautions (carried, all still true)

- Query runs BY CORRELATION, never `now()`-interval; planner rows purge in ~2 days.
- No orchestration dispatch within ~300s of a chassis pod (re)start.
- A roll kills in-flight council/orchestration runs — check pods before firing.
- `lock_blocked_change` = "the lock was exercised", never "the copy differed".
- The stale `reconcile_rerender:8d7c…` (canary plan) stays deferred FOREVER.
- Baselines stamp from the **DB clock**. (Local and DB agreed to 2s today; they did
  not on 08-15, when local ran 10h behind.)
- **Verify tool placement at `site_plan_sections`, never `pages.sections`** —
  LOCK-008 merges locked rows into the latter, so it carries positional slots that
  do not tell you what the planner chose.
