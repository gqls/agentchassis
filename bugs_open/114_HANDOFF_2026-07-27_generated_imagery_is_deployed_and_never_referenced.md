# 114 — imagery is planned, generated, deployed, and then nothing points a page at it

**Filed:** 2026-07-27 by the brochure_component_library workstream. The owner said
"there is not enough imagery on the site". The site had **21 generated images live and
serving 200**, and referenced three of them.
**Severity:** Medium-High. No data loss and nothing looks broken in the database — the
assets exist, the pages render, every status says complete. The site is simply bare.
**Class:** integrity of a multi-stage pipeline whose last stage is a queued work item
nobody drains.
**Status:** OPEN. The affected site is repaired at the data level; the mechanism is
untouched and 14 of 28 such items fleet-wide are parked the same way.

---

## Symptom, measured

```
$ for k in $(planned image keys); do curl -o /dev/null -w "%{http_code}" .../$k.jpg; done
   21 of 23 planned images -> 200          (the other 2 were transient curl resets; both 200 on retry)
$ grep -oE 'src="[^"]*\.(jpg|png|svg)"' *.html | sort -u
   /assets/images/logo.jpg                 <- the only image on the homepage
```

Six page heroes, fifteen line-illustration icons and a brand illustration, all
generated to the brief's house style (line illustration, one consistent tint), all
deployed, all unreferenced. Meanwhile four carousel cards and one feature block
rendered a **broken-image icon with the alt text showing**, because three components
fall back to `/assets/images/hero.jpg` — a filename that exists on no site in the fleet.

## Cause — three independent links in one chain

### (a) the re-render that would have wired them in was parked

```sql
SELECT status, count(*) FROM site_work_items WHERE spec->>'reason'='image_landed' GROUP BY 1;
   needs_human_review | 14
   complete           | 13
   cancelled          |  1
```

Five of those parked rows are this site's, created 2026-07-20/21:
*"Re-render index after its image asset landed"*, *"… capabilities …"*, *"… about …"*,
*"… multi-agent-review-council …"*. They have sat in `needs_human_review` for a week.
**Half of every "re-render after the image landed" item the platform has ever filed is
in that state**, which means the imagery stage of the pipeline is, fleet-wide, roughly
a coin flip.

### (b) three components guess a filename instead of degrading

`hero-card-carousel`, `people-feature-block` and `image-hover-card-grid` each carried:

```
src="{{if .image}}{{.image}}{{else}}/assets/images/hero.jpg{{end}}"
```

`/assets/images/hero.jpg` is not a convention anywhere — the real one is
`hero-<page>.jpg`. So "no image supplied" rendered as a **visibly broken** image rather
than as no image. The brief for this very site asks that components "degrade
gracefully"; this is the opposite of that, and it is worse than a blank because it
looks like a fault in the site.

### (c) a writer invented a whole path convention and nothing checked

`model-fine-tuning`'s `image-hover-card-grid` carried six card images under
`/images/illustrations/<name>.svg` — plausible filenames under a directory this
platform has never used. All six 404. They were written into `content_data` as free
text, and no stage between writing and serving asked whether they resolved.

### (d) 52 asset rows carry an unrendered template expression as their URL

```sql
SELECT count(*) FROM assets WHERE url LIKE '%input-data%';   -- 52
   robot-hands.com 19 | gamesdesign.co.uk 19 | fundamentallyai.com 12 | leopardessconsulting.co.uk 2
```

The stored URL is literally `/assets/images/input-data.asset-key.jpg` — a template
variable that never rendered, persisted as if it were a path. **This is a separate
defect and it is NOT diagnosed here**; it is recorded because it means the `assets`
table cannot be trusted as the answer to "where is this site's image", which is exactly
what the `image_url_404` discovery check asks it.

## Why nothing caught any of it

`check_image_url_404` exists and is enabled. Its own header says why it could not help:

> *"This is the lightweight DB-only version (per PLAN section 1.3): we don't check
> whether the file is actually in git, only whether an assets row exists. The full
> HTTP/git version would catch deployment failures too, but is deferred until we have
> git-adapter integration on the discovery path."*

So: (b) passes, because an `assets` row with purpose `hero` does exist. (c) passes,
because `/images/illustrations/…` does not match the `/assets/images/…` pattern the
check scans for at all. (a) is not a check's business — it is a queue nobody drains.

The deferred half is the whole half. `scripts/render_audit.py` (committed 2026-07-27)
does it from the other end: it renders the page, asks the browser which images failed,
then re-checks each one over HTTP before reporting.

> **Do not skip that re-check.** Measured the same day: 41 images reported broken
> across 7 pages, of which **35 served 200 on an unhurried retry** — a headless render
> fires every image request at once and our own origins throttle the burst. An audit
> that reports those 35 sends someone regenerating assets that were already there.

## Fix candidates, ordered by what closes the door

1. **Make the missing-asset case unrepresentable in markup**: a component with no image
   renders no `<img>`. *Done* for the three components (`085d`); a guessed filename can
   no longer reach the page.
2. **Drain, or stop filing, `image_landed`**. 14 parked rows are either work the
   platform intends to do (in which case something must do it) or a promise it should
   not make. This is the one that keeps the class closed; the rest is cleanup.
3. **Validate an image path at write time**, where the writer's invented
   `/images/illustrations/…` would have been rejected against the site's own asset
   list — cheaper than detecting it after deploy.
4. **Finish `image_url_404`'s deferred HTTP half**, or route the render audit's image
   findings into the same work-item stream.

## Related, deliberately not merged into this file

- `bugs_open/113` — the palette/contrast defect found in the same measurement pass.
  Different mechanism; they only share a discovery method.
- **`bugs_open/115`** — the three `audit_finding_brief_fidelity` rows this site filed on
  2026-07-24, still at `status='detected'`, one of which reads *"Only 2 of 27 components
  contain images — raising serious doubt that the illustration system is meaningfully
  present"*. They are the only rows of that item type in the whole database and nothing
  consumes them. Filed by a **concurrent thread in this same workstream** working the
  same owner report; it and this file were written within minutes of each other, and it
  was deliberately narrowed to the detection gap once the overlap was noticed. Read them
  together: 114 is why the imagery never arrived, 115 is why nobody was told.

---

> ## ADDENDUM 2026-07-29 — the hero.jpg fallback's mechanism traced, and the three live pages repaired (fundamentallyai)
>
> The "three components fell back to `/assets/images/hero.jpg`" finding is now fully
> explained and repaired on fundamentallyai.com (data-level; the platform gap stays open):
>
> - The literal comes from **`sites.content_data.hero_url`** — a site-wide legacy default
>   written before any hero asset existed. `BuildRenderContext` injects it site-wide; the
>   comment at `plan_sections_action.go:1608` names this exactly.
> - `plan_sections`' per-page hero aliasing defeats it **only on the plan_sections path**.
>   The LLM-free rerender path does not re-resolve fields
>   (`flag_page_image_rebuild_action.go` header), so scoped re-renders faithfully preserve
>   the stale value forever.
> - **`check_placeholder_image_in_use` can never fire for this class**: it keys on "no
>   assets row with purpose='hero'", but a site whose pages merely PREDATE its hero assets
>   has the assets row and still serves the placeholder. The check tests the wrong absence.
> - Measured merge order on the rerender path: injected site-wide `hero_url` >
>   per-page `content_data.hero_url`; only a `site_plan_imagery` page-scope row wins.
> - A dead value can also sit **stored but unrendered** (`content_data.hero_url` on a page
>   whose current template ignores it) — invisible to any crawl of rendered output, waiting
>   for a future template change. Found one on fundamentallyai's calculator page.
> - CSS `background-image: url()` references are invisible to href/img-src censuses, and
>   the gradient overlay makes the failure visually silent. Any imagery audit here must
>   grep `url('…')` in rendered_html too.
>
> Repair applied 2026-07-29 (fundamentallyai only): site default → real homepage hero,
> plan-intended hero on capabilities, LLM-free re-renders; verified on persisted rows and
> served pages. Full evidence: brochure_component_library NOTES, 2026-07-29 entry.

---

## Contribution 2026-07-29 (session "bugsearch 6") — your traced mechanism is live on SIX more sites, named and probed

Not a fix and not my lane. The 07-29 repair note above says **"fundamentallyai only"**,
and this file already says `/assets/images/hero.jpg` is *"a filename that exists on no
site in the fleet"*. Both true. What was not written down is **how many sites are
currently painting it**, so here it is, measured today.

**Ten live sites still carry the legacy site-wide default** `sites.content_data.hero_url`
— the injected value your session traced through `BuildRenderContext`:

```
dartsonline.com · gamesdesign.co.uk · idea.uk · oufe.com · relojistas.com
robot-hands.com · vetcomparison.uk · vonc.com · webdesign.co.uk   -> /assets/images/hero.jpg
fundamentallyai.com                                              -> /assets/images/hero-home.jpg   (your repair)
```

**Six of them are serving a 404 on deployed pages right now**, HTTP-probed
2026-07-29, every one `/assets/images/hero.jpg`:

```
404  gamesdesign.co.uk    404  idea.uk        404  oufe.com
404  relojistas.com       404  vonc.com       404  webdesign.co.uk
```

`idea.uk` is the site taking real money and `oufe.com` is the one with the
contemporaneous legal record, so two of the six are the estate's more exposed sites.
(The other four of the ten either do not reference the default on a deployed page or
resolve it — `robot-hands.com` and `dartsonline.com` did not appear in the deployed-page
reference scan, and `vetcomparison.uk` has no `/assets/images/*` references at all.
Re-derive rather than trusting that parenthesis.)

**Why no sweep will tell you this**, which is the part that connects to your own note
that *"the check tests the wrong absence"*: `image_url_404` has the identical flaw, and
I have just diagnosed it in `bugs_open/128`. It skips any rendered path whose basename
matches an active **asset purpose** for that site — so every one of these six is masked
by its own site's `hero` asset. Measured: **79 of 95 distinct rendered image paths on 13
sites are invisible to that check, 83% of its nominal surface**, and these six 404s are
inside the invisible set. So `check_placeholder_image_in_use` and `image_url_404` fail
the same way for the same reason, which makes it a class rather than two bugs.

Left entirely to this lane to decide whether to repair the other six the way
fundamentallyai was repaired — you have the merge-order finding and the LLM-free rerender
recipe, and I do not want to compete with a session that has that context. Flagging it
because your repair note reads as complete for the site it names, and a reader could
reasonably take "fundamentallyai only" as "the rest are fine".

---

## Contribution 2026-08-13 — a fresh instance, dartsonline.com's `flight-shapes.html`, both class (c)/(d) shapes present at once

From an owner conversation about adding explanatory images to `dartsonline.com`'s guides —
not from this bug's own workstream, contributing rather than opening a parallel account.

**The page is not imageless — its one image is invisible to an `<img>` scan, which is worth
restating because it nearly produced a false "zero images" claim of my own.** The hero is a
CSS `background-image` (`content-hero-flight-shapes.jpg`, confirmed **200**), exactly the
trap this file's ADDENDUM already names ("CSS `background-image: url()` references are
invisible to href/img-src censuses").

**The real finding: an illustration that would answer the owner's actual request — a
comparison diagram of the flight shapes the article discusses — already exists, generated
2026-08-05, and is wired to nothing:**

```
asset_key                                 purpose        url
illustration_flight_shapes_comparison     icon           /assets/images/input-data.asset-key.jpg   (class (d) — the literal template artefact, confirmed live at 200, i.e. really deployed under its broken name)
illustration_flight_shapes_comparison_lg  content_hero   <real signed S3 URL>
```

Neither `asset_key` appears in any page's `rendered_html`, fleet-wide (checked by direct
`ILIKE` scan, 0 rows) — an orphan by this file's own definition, not inferred from silence
elsewhere. `entity_type`/`entity_id` are both null on both rows, so nothing ever linked them
to the page they were generated for. No `site_plan_imagery` row names `flight-shapes` at all
(0 rows) — consistent with class (c): these were generated outside the planner path (matching
naming convention `illustration_<page>_comparison`, not `content_hero_<page>` alone), so the
planner-driven resolvers this file already names had nothing to find.

**Class (d) is not just a stored artefact this time — the broken path is live-serving.**
`input-data.asset-key.jpg` returns **200** (control: a genuinely nonexistent filename on the
same site returns 404), so at some point something actually wrote a file to that literal,
unrendered-template name and it has been sitting there, referenced by nothing, ever since.

**Nothing applied** — no page edit, no asset rewrite. Left for whoever next works this bug or
the imagery pipeline generally: the cheap win is wiring `illustration_flight_shapes_comparison_lg`
onto this one page (it is exactly the image the content needs); the fix that closes the door
for the *next* orphan is still candidate 2 in this file (drain or stop filing the queue that
was supposed to catch this).

> ## ⚠ CORRECTED 2026-08-14 — THE ENTRY ABOVE IS WRONG ABOUT THE MECHANISM. This asset was never an orphan: it was WIRED, and then DESTROYED. It does not belong to this bug at all.
>
> **What caught it:** reading the owning lane's own log before acting on my own finding —
> `dartsonline_traffic/README_where_we_are.md`, entry 2026-08-05, which states that all 8
> built guides were illustrated that day, *"three technical line-diagrams (barrel profiles,
> the board-setup measurements, **the four flight shapes**)"*. My "0 rendered_html references
> fleet-wide" scan was accurate **as measured today** and I read it as "was never wired",
> which the scan cannot distinguish from "was wired and has since been overwritten". A
> present-tense absence is not a history.
>
> **The history, `[MEASURED]` from `page_component_history` on `page_id`
> `73f9e8ad-30f9-4383-9452-5164e6a3ca1a`:**
>
> | when | what |
> |---|---|
> | 2026-08-05 11:23 / 11:28 | the two illustration assets are created |
> | 2026-08-05 11:52:48.095654 | archived article-body `content_data` — **4018 chars, contains `<img` AND `flight-shapes-comparison`** |
> | 2026-08-05 11:52:48.185471 | current live row written — **6207 chars, contains neither**, `source = save_page_sections_overwrite` |
>
> **90 milliseconds apart.** A wholesale body rewrite replaced the prose with a longer,
> genuinely better version and took the embedded figure with it. It has been gone for nine
> days and nothing reported it — the work item completed, the page deployed, the asset kept
> serving 200.
>
> **The 08-05 session PREDICTED this in writing and was ignored by me, not by the platform:**
> *"another thread was actively rewriting several of these exact guide pages' body text while
> I was placing images into them … three of my placements landed in a page moments before it
> was replaced, and were silently lost rather than erroring … if a caption vanishes from a
> guide later, that's most likely why, not a new bug."* That is exactly what happened, and it
> is the fourth known instance of their four.
>
> **Why this matters beyond the correction — the in-body imagery mechanism EXISTS and is
> simply not durable.** The recovered markup is a plain `<figure><img><figcaption>` embedded
> in the `article-body` component's single `content` field (which is `source: "llm"`, so the
> prose and the imagery share one overwritable blob). So "there is no way to put an image
> inside an article body" — which I believed and had told the owner — is **false**. The real
> defect is that in-body imagery has no representation the regenerator can see, so any
> rewrite silently discards it. That is the `save` REPLACES / `rerender` MERGES class already
> recorded for a different payload in `bugs_open/238`, and the no-archive-no-warning shape in
> `bugs_open/226` / `229` — **not this file's "nothing ever points a page at it" class.**
>
> **So: this instance is NOT evidence for bug 114** and should not be counted in its
> population. Left here rather than deleted because the wrong reading is the useful part —
> an orphan census run at one instant cannot tell a never-wired asset from a destroyed
> placement, and both this file's own scans and mine have that blind spot. Distinguish them
> with `page_component_history`, which is what settled it in one query.

---

## CONTRIBUTION 2026-08-16 — ten clean instances in one night, one producer, one site (mortgagecalculator.co.uk), and the consumer the producer's own header names cannot see them

By the mortgagecalculator adoption lane. A one-shot `design-discovery-agent` run (08-14 21:46Z)
had `content_image_missing` file 10 `needs_imagery` GENERATE items (`kind: content_hero`,
`scope: page`, `scope_ref: tool-<x>`, keys `content_hero_tool_*`) for the site's tool pages,
which are listed by a `query.*` consumer. Owner: *"let the hero items run."* They ran, 08-15
19:31–19:43Z: **10/10 `complete`, 10 `active` assets, 10 files serving 200**
(e.g. `/assets/images/content-hero-tool-repayment.jpg`, 115,250 B).

**Nothing references any of them** `[MEASURED 2026-08-16 ~10:15Z]`:
- `assets.entity_type` / `entity_id` **NULL on all 10** — the same null-link shape this file
  records at its `content_hero` census above, so the listing-card join
  (`pageImageProjection`: `entity_type='page' AND entity_id=p.id AND purpose='card'`) can
  never pick them up, and the header's own convergence story ("pass 2 sees each card's origin
  no longer matches and re-derives") has no origin to see.
- Each tool page DID re-render through the image-landed flow (`needs_page` from
  image-build-handler 19:32Z → `page_rerender … assemble` 19:46Z → hero component updated
  19:47:51Z) — and its hero fields resolved to the **site fallback**: `hero_url` and
  `background_image` both `/assets/images/hero.jpg`. `rendered_html` contains
  `content-hero-tool-repayment` **0** times. The page-scoped asset was not mapped into the
  page-scoped hero by that flow.
- The homepage `tool-list` `items[0].image` is still `""` (the frozen-array shape the 08-14
  NOTES on this site trace to `plan_sections_action.go:2076` — no per-array-element source).

So the header comment's line 14 — *"The article page re-renders with its new hero via the
normal image-landed flow"* — is a claim, and on this site it is false: the flow re-rendered
the page and the page still points at the fallback. `[INFERRED, not read]`: the flow maps
`hero_url` from the plan/site hero (`site_plan_imagery` join or site fallback), and a
`content_hero_*`-keyed asset with no plan row and no entity link matches neither.

**Two things this instance adds to the file:**
1. A **controlled** population — same producer, same night, same site, no other lane touching
   imagery — so it is not the "one instant census" blind spot the section above warns about
   (`page_component_history` for `tool-repayment` shows one rerender, no destroyed placement).
2. The **cost of an owner saying "let them run"** on this class is now a number: 10 paid
   generations, 0 pixels changed on any page. Until the entity link is set at generation time
   (or the page hero mapping reads `content_hero_<page>`), `content_image_missing`'s GENERATE
   arm should be treated as spend without effect on any site whose listed pages are not
   articles with cards.

Left in place: assets active, files served, nothing cancelled — they are the fix's test
fixture when someone wires the link. Query to find them:
`SELECT asset_key FROM assets WHERE site_id='62b5978e-4271-4589-8e00-4baebfc0447c' AND asset_key LIKE 'content_hero_tool_%';`

---

## CONTRIBUTION 2026-08-17 from the `bugs_open/284` lane — the fleet census, and it splits three ways

Not a competing fix — a measurement, offered because the owner asked my lane to "add
the missing images" for the 31 open `image_url_404` findings and the census says most of
them are **not missing images at all**. This file's subject (imagery deployed and never
referenced) is the mirror of what I found, so the numbers belong here.

**How the 30 `unbacked_path` findings actually split** (live table, 2026-08-17; the
join asks, per finding, whether the site has an ACTIVE asset under the exact
`asset_key` the page references, and separately whether it has one of that `purpose`):

| referenced basename | findings | asset under that EXACT `asset_key` | any asset of that purpose |
|---|---|---|---|
| `hero` | 10 | **4** | 10 |
| `case-study-*` | 10 | **5** | 0 |
| `favicon` | 4 | 0 | 0 |
| `og-card` | 4 | 0 | 0 |
| `logo` | 2 | **2** | 2 |

So, by what the fix actually is:

- **11 findings — the asset EXISTS under the exact referenced key** (4 `hero`, 2 `logo`,
  5 `case-study-*`). Nothing needs generating; the deploy to the canonical web path
  either never ran or failed. `deploy_image_asset_action.go:383` documents the mapping
  (`asset_key=hero, purpose=hero → assets/images/hero.jpg`), so the page is referencing
  exactly the right path and the artefact is simply not there. **This is this file's
  class, from the other end.**
- **6 findings — `hero`, where the site's heroes are all PAGE-SCOPED.** e.g.
  lendzy.co.uk has 9 active heroes, keyed `hero_home`, `hero_about`, `hero_price_cap`,
  `hero_your_rights`, … and **none** keyed plain `hero`, so nothing deploys to the base
  `assets/images/hero.jpg` that a page still references. The fix is a repoint (or giving
  that page an asset keyed `hero`), not a tenth hero. `bugs_open/168` is adjacent
  (`DeployedWebPath` and underscore keys).
- **5 findings — `case-study-*` with no asset at all** (ai-agent-orchestration.com,
  finetuning.uk). Genuinely missing; generation is the right action.
- **8 findings — `favicon` (4) and `og-card` (4), nothing exists.** Genuinely missing,
  and **owned by `bugs_open/131`** — left alone here deliberately.

**⚠ And a live blocker for anyone about to generate any of them.** The imagery path is
failing today: `needs_imagery` shows **12 failed against 4 complete on 2026-08-17**
(`remortgagecalculator.uk` 10, `loancalculator.co.uk` 2), all with
`step call_asset_deployer failed: … failed to get latest commit/base tree for branch
"master": github API request failed with status: 404`. **The obvious cause is not the
cause:** all four sites involved have an EMPTY `sites.github_repo`, but so do the sites
whose items COMPLETED in the same window — the same two domains appear on both sides of
the same day, so an empty repo config does not discriminate. Not diagnosed further; it
is those lanes' site work, but it means filing fresh imagery items today buys a 3-in-4
chance of another failure row.

**What I have deliberately NOT done:** no imagery items filed, no deploys dispatched, no
rows touched. The owner asked for images to be added; the honest answer is that 17 of 30
need a deploy or a repoint instead, 8 belong to `131`, and the 5 that really do need
generating should wait for the deployer's 404 to be understood. Recorded in
`docs024_key_docs_latest/bugfix_284_flag_only_items_promoted/` (NOTES + README).

### ADDENDUM 2026-08-17 evening — WHY the 5 "genuinely missing" ones are missing: the generator's surface list does not include their page type

Follow-up to the census above, and it closes the question rather than leaving it as "5
need generating". I did **not** file them, and this is why.

`check_content_image_missing` is the framework's producer for content heroes — it is the
check named in `spec.check` on every healthy `needs_imagery` row on these sites. Its
population is a hardcoded surface list:

```go
// check_content_image_missing.go
:131   PageType: "blog-post",
:137   PageType: "tool",
```

and its sweep is `... AND p.page_type = $2` (:223). **The case-study pages are
`page_type = 'content'`** (measured on finetuning.uk + ai-agent-orchestration.com). So
they are outside the generator's population by construction — no pass of that check can
ever emit imagery for them, however many times it runs.

That is the actual defect behind those findings: **the page templates reference
`/assets/images/case-study-*.jpg`, and no producer in the framework generates content
imagery for the page type those references live on.** `image_url_404` is reporting it
accurately; the gap is upstream of the report.

**What I deliberately did NOT do, and would ask this lane to weigh:**

- **No hand-filed `needs_imagery` rows.** The healthy rows carry a prompt composed by the
  framework from the page's own title and description; hand-writing five prompts would be
  a session writing site content (against the 2026-08-06 ruling), and it would paper over
  the surface gap so nobody fixes it.
- **No widening of the surface list by me.** Adding `content` there is a two-line change
  and a **fleet-wide** one: every `page_type='content'` page on every site enters the
  generator's population at once, which is real image-generation spend and belongs to
  whoever owns this lane, with the usual gate. Worth measuring the population first —
  `SELECT count(*) FROM pages WHERE page_type='content'` across the fleet — because that
  count is the size of the first sweep.

**Also settled tonight, since it was recorded above as a blocker: the deployer's 404 is
GONE.** After the v1.0.1307 roll (`revision=a6d1c53c`, digest matched to the pods), the
re-triaged items resumed: one **completed** end-to-end at 18:33 and the only failure since
is `call_asset_deployer … timed out after 3 retries` — a transient, not the structural
`branch "master" 404`. The completion re-used an asset generated pre-roll at 13:35
(`hero_lenders`, `62ffe42d`), which is why `assets` shows nothing created since the roll
and why that zero is not a sign of failure. So generation+deploy works again; the five
above are blocked by the surface list, not by the deployer.

---

## FIX IN PROGRESS 2026-08-22 — the site-wide default was being CORRUPTED BY THE PIPELINE ITSELF, and four claims in this file are corrected

By the `bugfix_114_imagery_wiring` lane (`docs024_key_docs_latest/bugfix_114_imagery_wiring/`),
which is taking the framework half of this bug rather than contributing another instance.
Ownership re-checked first: three lanes cite 114, none owns the fix.

### The bug is still valid, and larger than filed `[MEASURED 2026-08-22]`

| | at filing | today |
|---|---|---|
| sites carrying the legacy `content_data.hero_url` | 10 (07-29 contribution) | **18** |
| `assets` rows with no entity link | — | **518 of 580 as of 2026-08-22**; in the last 14 days only `card` (45) is ever linked |
| content_hero assets vs components wired to one | — | **23 of 94 as of 2026-08-22** fleet-wide (gamesdesign 14/0, finetuning 14/0, leopardess 7/0, fundamentallyai 6/0). ⚠ A census goes STALE BY ADDITION: re-run before quoting — `git log --since=2026-08-22 --diff-filter=A -- platform/orchestration/actions/discovery_checks/` for new producers |
| parked `image_landed` items | 14 parked / 13 complete | **8 parked / 40 complete** — the queue now mostly drains |

**fundamentallyai's 2026-07-29 repair has reverted.** It reads `/assets/images/hero.jpg`
again. Nobody undid it.

### WHY it reverted, and why the spread from 10 to 18 sites — this is the new finding

`StoreAssetAction` wrote `sites.content_data.<purpose>_url` on **every** asset store,
deriving the value from `storage.BuildAssetPaths(purpose, ext)` — from the **purpose
alone**. The deployer commits under the **asset key**
(`deploy_image_asset_action.go:403` → `storage.DeployedAssetPath`). Two derivations for
one artefact, so **every page-scoped hero generation re-stamped the site-wide default
with a path that may exist nowhere**. An operator repair could never hold: the next
generation overwrote it.

The census that makes it unambiguous — `count(DISTINCT value)` per key across all sites:

```
hero_url          18 sites, 1 distinct value   /assets/images/hero.jpg
icon_url          16 sites, 1 distinct value   /assets/images/icon.jpg      404 on all
logo_url          15 sites, 1 distinct value   /assets/images/logo.png
content_hero_url   6 sites, 1 distinct value   /assets/images/content_hero.jpg   404 on all
illustration_url   3 sites, 1 distinct value   /assets/images/illustration.jpg
sprite_sheet_url   1 site,  1 distinct value   /assets/images/sprite_sheet.jpg
```

**One distinct value per key across every site is the signature of a value with no site
and no asset input.** `content_hero.jpg` and `icon.jpg` are filenames the deployer cannot
produce for any input; HTTP-probed 2026-08-22, both 404 on all six sites carrying them.

And the gate that should have stopped it **already existed in config and nothing read
it**: `image-build-handler`'s imagery store step has passed `update_site_brand_assets:
false` since it was written; `grep -rn "update_site_brand_assets" --include=*.go platform/
internal/` returned nothing. Six live steps declare `true`, two declare `false`, and the
two declaring `false` are exactly the page-scoped ones.

**Fixed and committed** (`ebd1ce890`, register **IMG-072**, council corr `3c0560f3`):
the declaration is honoured in both directions, an undeclared caller writes only when
`asset_key == purpose`, and the value is now the deployer's own path. Three mutation
proofs. **Inert until the next chassis roll.** The repair of the existing 18 rows is
deliberately held until the gate is live — applying it first invites the next generation
to undo it, which is precisely what happened to fundamentallyai.

### The 08-15 mortgagecalculator population, re-examined — and two of my own hypotheses refuted

The 08-16 contribution above reports 10 tool heroes with 0 references. Re-measured:
**two of eight are wired** (`tool-equity-release`, `tool-overpayment`), six are not. Same
site, same day, same agent. That contrast is more informative than the zero.

Ruled out, each by measurement rather than argument:
- **a race** — every asset was `active` 972–2650 s before its render, and
  `tool-affordability` rendered **2.2 days** later and still missed;
- **the plan routes** — mcalc's current plan has **no** page-scope hero row for any tool
  page and **no** site-scope hero row at all (only `logo`), so routes 1 and 3 were empty
  for all eight equally;
- **which flow ran** — filed to the diagnosis loop (run corr
  `ea7dfeef-c11d-40c4-b24f-b8f42413b1ae`; verdict UNVERIFIABLE, but narrowing): both
  pages carried `handler_agent='page-build-handler'` and the wrong value was produced by
  that flow's own resolution, not by a later overwrite;
- **a stored value being sticky** — the failing page had carried
  `/assets/images/hero.jpg` since 13:46 that day and the wired ones had no stored
  `hero_url` at all, which looked decisive until the population was widened: **ten pages
  fleet-wide carry that value in `page_component_history` and are wired to a content-hero
  today** (`idea.uk/tool-funding-fit` has 23 such versions). `carryStored` fires only when
  the source resolves nothing, exactly as its comment says.

What survives is narrower: routes 1, 2 and 4 of `ensureAssets` are **all** gated on
`pageName != ""`, route 3 needed a row that site lacks, route 5 is ungated — so the
outcome is exactly "every pageName-gated route skipped". **Not asserted as the cause**:
the 08-15 orchestration rows are purged and the runtime evidence the loop asked for
cannot be recovered.

So the lane fixed what made it unknowable instead of guessing (`736108464`): the
fallback branch now logs that it fired and whether the page-scoped routes were eligible
at all. Falling through is legitimate for a legacy site with no plan imagery and is also
exactly this bug's symptom — the two were indistinguishable after the fact.

### ⚠ FOUR CLAIMS IN THIS FILE ARE STALE — corrected here, dated, not edited away

1. **The ADDENDUM's "the LLM-free rerender path does not re-resolve fields"
   (`flag_page_image_rebuild_action.go` header) is a MISREADING.** The header says that of
   the *terminal assemble leg*. `rerender_page_sections` **does** re-resolve
   (`rerender_page_sections_action.go:20-23`, `:459`).
2. **The ADDENDUM's merge-order claim is CONTRADICTED by source.** It states injected
   site-wide `hero_url` beats per-page `content_data.hero_url`. Fresh `plan.ResolvedData`
   is merged **last** and wins (`rerender_page_sections_action.go:614-620`); the base only
   wins when the fresh data carries no hero.
3. **`plan_sections_action.go:1608` no longer names that comment** — it is now
   `:2424-2432`. Line refs in this file are a year of drift; resolve by symbol.
4. **"Why nothing caught any of it" quotes a header that has since been replaced.**
   `check_image_url_404` was fixed and closed (`bugs_closed/128`, live v1.0.1219) — it
   compares exact deployed paths now and scans `site_components` too. What still stands is
   class (c): paths outside the `/assets/images/` prefix remain invisible to it. Separately
   `check_placeholder_image_in_use` was narrowed 2026-08-12 to the canonical `asset_key`,
   and when it fires it files `needs_hero_image` — i.e. **generates a new site hero rather
   than wiring the page-scoped one that already exists**, which is still this bug.

### Still open, and who has it

- **The entity link** (`assets.entity_type`/`entity_id` never written at generation; the
  only writer is `derive_card_asset_action.go:214-228`, `purpose='card'` hardcoded) and
  **event-driven convergence** — this lane's parts 2 and 3, specified in
  `bugfix_114_imagery_wiring/PLAN_2026-08-22_imagery_wiring.md`, not yet built.
- **⚠ Convergence has been dead for eleven days and nothing said so.** The DERIVE arm that
  writes the entity link needs a later `design-discovery-agent` sweep. `site_discovery_rotation`
  shows that lane's newest `last_selected_at` as **2026-08-11** while all four sibling lanes
  are current (08-21/22). The `site-discovery-staleness` CronJob (`bugs_open/230`) reports
  it **daily** — the design lane simply never appears in "stamps advanced last 24h" — and
  nothing consumes the report. **That belongs to 230's lane**; recorded here because it is
  why a one-shot generation never converges, and it is the argument for making convergence
  event-driven rather than sweep-driven.
- ~~**`check_undeployed_assets.go:289-305`** matches `rendered_html` against the underscored
  purpose while deployed files carry the hyphenated key.~~ **REFUTED 2026-08-22, same day,
  before it was acted on.** In SQL `LIKE`, `_` is a single-character wildcard, so
  `content_hero-%` matches `content-hero-tool-x.jpg`. Both predicates return **31** over
  deployed `page_components` — the check sees them. The function carries a comment saying
  the pattern *"is deliberately left unescaped"* and to read the file header before
  changing it; I quoted the query under that line without reading the header it points at.
  Recorded rather than deleted because the `[UNVERIFIED]` marker is what made it get
  checked instead of filed.
- **Owner decisions, unasked as yet:** widening `check_content_image_missing`'s surface to
  `page_type='content'` (fleet-wide generation spend); and the disposition of the five
  parked rows whose pages resolve no sections.



---

## POINTER 2026-08-24 — the sibling failure, filed as `bugs_open/384`

`dartsonline_traffic` lane. This file's cases are assets that are **not referenceable** — entity
link NULL, or a hero mapping that resolves to the site fallback. **384 is the case where the
asset IS derived, IS linked and IS joinable, and the listing still does not show it**, because
nothing re-renders the listing when a card lands: `derive_card_asset_action.go` has no rebuild
emission of any kind, and `check_orphan_pages` keys on membership rather than card freshness.

Worth knowing here because **this file's own fix created that exposure**: once
`emitContentCardDerive` made cards land promptly at the landing event rather than waiting for a
sweep, the stale listing became the remaining hop rather than a rarity hidden behind slow
derivation. Live instance: 4 of 12 homepage cards on dartsonline.com with the bytes served and
the page stale `[MEASURED 2026-08-24]`.

---

## CONTRIB 2026-09-02 — two live instances of this bug's exact shape, and both are a COMPONENT FIELD problem rather than a queue problem

**From:** the `inline_guide_imagery` lane. **Nothing dispatched at either site.** Found by
calibrating a proposed detector against the live population rather than by hitting the symptom —
which is why these are offered as evidence, not as a new theory.

**What I ran.** For every `scope='section'` imagery row on a CURRENT plan whose `key` joins an
**active** asset, does that page render the asset anywhere? `[MEASURED 2026-09-02]`

| site | scope_ref | asset (active) | rendered on the page? |
|---|---|---|---|
| fundamentallyai.com | `about:2` | `illustration_people_feature` | **NO** |
| idea.uk | `index:1` | `illustration_workshop_concept` | yes |
| vonc.com | `about:2` | `illustration_game_master` | **NO** |
| vonc.com | `index:2` | `illustration_gauntlet_cta` | yes |

**Two of four.** Planned, generated, deployed, active, paid for — and referenced by nothing. That
is this bug's sentence, still true five weeks on, on two sites nobody has looked at.

**The two failures have DIFFERENT causes, and neither is the undrained queue this file opens with.
Both are in the component's own field declaration.**

**1. vonc.com `/about` — the field asks for the wrong thing, and gets the page hero.** Plan
ordinal 2 is `game-master-explanation`, whose `gm_image_url` is sourced **`site_assets.image`**.
That is not a literal key: `imageryplan.imageRoleAliases` maps `image → hero` unconditionally, so
the section resolves the page's own hero. Measured at the stored HTML: positions 2
(`content-block-about`, also `site_assets.image`) and 3 (`game-master-explanation`) BOTH render
`/assets/images/hero-about.jpg` — **the same file twice on one page, directly beneath the hero
that already shows it** — while `illustration_game_master` sits active and unreferenced. This is
the trap IMG-074 documents; what is new is a live instance where it is **defeating a planned
section-scope illustration**, not merely duplicating chrome.

**2. fundamentallyai.com `/about` — the field has NO source at all, so nobody writes it.** Plan
ordinal 2 is `people-feature-block`, whose `image` field declares an empty `source` (its sibling
`image_alt` is `llm`, `required: true`). Read at HEAD, an empty source is reached by **neither**
path: `planSection` adds a field to `llmFields`/`llmFieldSpecs` only `if source == "llm"`
(:2669), so the writer is never told about it; `resolve()` returns `(nil, true)` for an empty
source (:946), so `found && value != nil` fails; and `carryStored` declines an empty source
(:2524). Resolved by nobody, written by nobody, carried by nobody — the section renders no `src`
at all, permanently and silently. `[MEASURED 2026-09-02]` this is **one field on one active
component** fleet-wide, so it is a one-off rather than a class — but it is invisible by
construction, and no existing check names it.

**The fleet population behind case 1, which IS a class** `[MEASURED 2026-09-02]`: of the
image-named fields on active components, **19 fields across 7 components are sourced
`site_assets.image`** — every one of them resolves to its page's own hero. Against 1 sourced
`site_assets.hero` (honest), 1 `site_assets.illustration`, 1 `site_assets.background`, 1
`site_assets.product_screenshot`, 12 `llm`, and the empty one above.

**This meets a trigger that was written down and left waiting.** IMG-074's council round recorded
the `bug_historian` objection that the other components declaring `site_assets.image` remain
exposed, and the architecture seat's response: *"if a THIRD component hits the same alias trap,
that is the point to ask whether the alias map needs an explicit opt-out."* Two components on one
page, plus a planned illustration defeated, is that point.

**What changed that makes this newly worth fixing rather than merely worth noting.** Until
2026-09-01 a section-scope imagery row could not bind to its own section — every section on a page
resolved the first row of its kind — so repointing these fields would have delivered one
illustration to every section. That is fixed and live (**IMG-075**, chassis `v1.0.1351`,
symbol-probed at both replicas). **So the remedy for case 1 is now the same one migration 644
applied to `Illustrated Text Block`: repoint the field from `site_assets.image` to
`site_assets.illustration`, and the planned illustration reaches the section that planned it.**
Both of these pages would bind today — their plan order and live section order agree, which is the
condition IMG-075's guard checks.

⚠ **Two cautions, because neither is my site and I have not run either remedy.** Repointing a field
whose current value is the hero will make sections that were *silently* showing the hero show
nothing until an illustration exists for them — `on_missing: skip_field` on both fields makes that
degrade quietly rather than break, but it is a visible change on live pages and belongs to whoever
owns them. And per IMG-074's own note, several of the eight components declaring
`site_assets.image` may legitimately want the hero; the two named here demonstrably do not,
because a planned illustration for that exact section exists and is active.

The re-runnable query is in `docs024_key_docs_latest/inline_guide_imagery/RUNBOOK_inline_guide_imagery.md`.

---

## CONTRIB 2026-09-02 (editorial_design_uplift) — the population is **189 pages across 21 sites**, and the obvious component-level remedy is WRONG (I applied it and rolled it back)

Arrived here from the opposite end: the owner's second boxingonline review said *"there is not
enough imagery in any of the pages"*. Six `/blog/` articles, each with a generated, deployed,
HTTP-200 `content_hero` asset, and each serving exactly one `<img>` — the logo. That is this
bug's opening sentence, three sites and five weeks later.

### 1. The size of it, which I do not think this file has had

`[MEASURED 2026-09-02]` Pages holding an **active** `content_hero` asset under the
`ContentHeroKey` convention, with **no component on the page that could render it** (no field
sourced `site_assets.hero` or `site_assets.image`):

> **189 pages across 21 sites.**

webdesign.co.uk 65 · loanandmortgagecalculator.co.uk 17 · finetuning.uk 12 · gamesdesign.co.uk 11
· leopardessconsulting.co.uk 7 · robot-hands.com 7 · loancash.co.uk 7 · vonc.com 7 · idea.uk 7 ·
dartsonline.com 7 · **boxingonline.com 6** · lampenkap.com 6 · (tail continues)

```sql
WITH hero_comp AS (SELECT id FROM content_components WHERE is_active AND EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(input_schema->'fields',input_schema->'properties')) f
       WHERE f.value->>'source' IN ('site_assets.hero','site_assets.image')))
SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
  JOIN assets a ON a.site_id=p.site_id
                AND a.asset_key='content_hero_'||replace(p.name,'-','_') AND a.status='active'
 WHERE NOT EXISTS (SELECT 1 FROM page_components pc JOIN hero_comp hc ON hc.id=pc.component_id
                    WHERE pc.page_id=p.id AND pc.build_status IS DISTINCT FROM 'removed')
 GROUP BY 1 ORDER BY 2 DESC;
```

### 2. ⚠ THE PART THAT WILL SAVE THE NEXT READER A ROLLBACK

The obvious fix, staring at six unreferenced images, is *"the article component cannot display an
image — give it one"*. **I did that. It passed two council rounds and eleven reviewer seats. It
was wrong, and I rolled it back 69 minutes after applying it.**

`[MEASURED 2026-09-02]` **292 of the 301 pages carrying `article-body`, across 31 sites, ALSO
carry a `hero` component whose `background_image` reads the same `site_assets.hero` key.** A new
image field on `article-body` sourced from that key renders **the same image twice on 97% of the
population** — hero at the top of the page, the identical file again at the top of the article.
That is exactly the defect the `inline_guide_imagery` lane documented on `vonc.com/about` in the
CONTRIB above, and it is what my change would have industrialised.

The pages that motivated it are the **nine-page minority with no hero component at all**. I
measured the motivating case and generalised it to the population. Migration `686` is in the repo
with a DO-NOT-APPLY header and a ledger row deliberately left in place so `--apply` cannot replay
it; nothing was ever rendered with it (0 of 301 instances acquired the field). Full account:
`WRONG_CALLS.md`, *generalised-a-remedy-from-the-motivating-case-to-the-population*.

### 3. The reframe, which I think is this bug's actual shape

Healthy pages are not missing a capability. They render imagery through the **`hero` component
fed by a page-scope `site_plan_imagery` row** — verified at the artefact:
`agritec.uk/blog/insect-bioconversion.html` serves
`background-image: …url('/assets/images/hero-bsf.jpg')`, has exactly 1 page-scope plan hero row,
and its only `<img>` is the logo. The `inline_guide_imagery` lane measured the same shape from
their side: **330 of 432 active guide/blog pages carry a hero component.**

So the ~189 are pages the **planner composed without a hero and without a page-scope plan row** —
and `ContentHeroKey` exists *precisely* to generate a per-article image for **"a page the planner
gave no hero of its own"** (`imageryplan.go:221-233`). **The system correctly detects the
composition gap, generates an image to fill it, and then has nowhere to put the result.** The
defect is upstream of every render-end fix: *why does the planner compose a blog/guide page with
no hero when 330 of 432 peers have one?*

### 4. Link to `bugs_open/412`, which I think explains the persistence

412 (*"a page cannot gain a hero image without a FULL LLM REBUILD of its copy"*, OPEN, unowned)
says the only route from "no hero image" to "hero image" forces a full copy regeneration, which
sites under copy review will not accept. **If that is right, it is why these 189 stay orphaned
even once someone notices** — the cheap remedy is unavailable and the available one is refused.
Whoever picks up either should read the other; I have not verified 412's mechanism first-hand and
am not asserting it, only pointing at the join.

### 5. ADDENDUM, same day — the composition question, measured at the prose grain

`[MEASURED 2026-09-02]` Across all **462** `blog-post` + `guide` pages fleet-wide, the composition
is furniture around a single slab: `hero` on 362 pages, `call-to-action` on 329, `article-body` on
310. And the decisive one:

> **Maximum prose sections on ANY article page in the estate: 1. Pages carrying more than one: 0.**

⚠ **The row COUNT alone would have overstated the opposite** — `blog-post` averages 2.7 section
rows and `guide` 1.8, which does not read as "one slab" until you ask what the rows *are*. The
extra rows are hero and CTA. Counting rows answers a different question from counting prose
sections, and only the second one bears on this.

So the planner has never composed an article out of parts, anywhere, regardless of what its menus
offer — the `inline_guide_imagery` lane reached the same place from the selection side
(illustration-capable sections are selected, but on `landing` pages, one per page, and **zero** on
`blog-post`/`guide`). Three routes, one destination: **article-shaped output is one prose slab plus
chrome**, and the ~189 orphaned content heroes are what happens when that slab has no hero beside
it. Still a question, not a diagnosis — none of us has read the planner.

**Not dispatched anywhere.** No site touched, no work item filed — both remedies are visible
changes on live pages and belong to their owners, and the composition question belongs to whoever
owns the planner.

---

## CONTRIB 2026-09-02 (mortgagecalculator_couk_adoption lane) — a SECOND population, with a different mechanism and a one-line fix: 54 pages that DO have a hero component and still cannot show their own image

Routed here rather than filed new, for the same reason the two contributions above were: this
bug's opening sentence is the symptom, and it is OPEN and actively worked today.

**This is not the ~189.** The contribution above characterises the population as *"pages the
planner composed without a hero"* — no hero component, nowhere to put the image. Correct for that
set. **There is a second, disjoint set where the page HAS a hero component, the asset exists and
is active, the resolver would find it, and the image still cannot appear** — because the component
declares no image-typed field, and that is the flag the resolver is gated on.

### The mechanism, in the platform's own words

`plan_sections_action.go:2846` gates the whole per-page hero path on
**`sectionHasImageField(fieldsRaw)`** — true only if the component's `input_schema` declares a
field of `type: "image"` or `"image_url"`. Only then is the resolved page hero written into
`resolved_data` under the aliases `hero_url` / `background_image`. The comment directly above the
gate states the failure mode:

> *"resolved_data is merged LAST at render time … this is what lets the per-page hero defeat the
> site-wide hero_url that BuildRenderContext still injects for legacy templates: **without it,
> `{{or .hero_url .background_image}}` picks the site-wide value and every page shows the same
> image**."*

`hero-tool` emits exactly that expression in its template — `url('{{or .hero_url
.background_image}}')` — and declares **no image-typed field at all**. So on every `hero-tool`
page the gate is false, the aliases are never written, and the template falls through to the
site-wide default. Permanently, regardless of what imagery exists for that page.

### The measurement, with the control that makes it mean something

[MEASURED 2026-09-02, live DB]

| component | declares an image-typed field? | instances | carrying a per-page `content-hero-*` image | sites |
|---|---|---|---|---|
| `hero-tool` | **no** | 69 | **0** | 21 |
| `hero` | **yes** (`background_image`: `type: image`, `source: site_assets.hero`, `fallback: /assets/images/hero.jpg`, `on_missing: use_fallback`) | 632 | **72** | 36 |

**0 of 69 against 72 of 632.** The `hero` row is the control: same resolver, same alias keys, same
template expression — the only difference is the schema declaration, and it is the difference
between 0% and 11%. Had `hero-tool` pages been showing per-page heroes, this comes out non-zero.

What the 69 render instead: `hero-home.jpg` ×28 (8 sites), `hero.jpg` ×21 (9 sites), no image at
all ×19 (6 sites), and **one** per-page image — `leopardessconsulting.co.uk`'s
`tool-automation-savings-estimator`, which has `background_image` written **directly into
`content_data`**. That single exception is worth more than the other 68: it proves the **template**
is fine and it is only the **resolver path** that is switched off.

**The size of the loss: 54 of the 69 pages have an active asset at exactly their
`ContentHeroKey`** (`'content_hero_' || replace(page_name,'-','_')`), already generated, already
deployed, and structurally unshowable. Re-runnable:

```sql
SELECT count(*) FROM (
  SELECT DISTINCT p.id, p.name, p.site_id
    FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    JOIN content_components cc ON cc.id = pc.component_id
   WHERE cc.function = 'hero-tool') q
  JOIN assets a ON a.site_id = q.site_id
               AND a.asset_key = 'content_hero_' || replace(q.name,'-','_')
               AND a.status = 'active';
```

### Why this one is cheap, and why it does NOT hit the trap the contribution above warns about

The fix is to declare on `hero-tool` the field `hero` already declares, byte for byte:

```json
"background_image": {"type":"image","source":"site_assets.hero",
                     "fallback":"/assets/images/hero.jpg",
                     "required":false,"on_missing":"use_fallback"}
```

Three properties worth checking before anyone applies it:

1. **It cannot render an image twice.** The warning above — *292 of 301 `article-body` pages also
   carry a hero reading the same `site_assets.hero` key, so giving `article-body` its own image
   field duplicates the image on 97% of them* — turns on `article-body` being a **second** consumer
   on a page that already has a hero. `hero-tool` **is** the hero: it is the page's only
   image-bearing band, and it is already emitting a `background-image` today. This declares the
   field on the component that is already drawing the picture, not on a second one.
2. **It is strictly an improvement or a no-op.** `on_missing: use_fallback` with the same
   `/assets/images/hero.jpg` the site-wide injection supplies today means a page with no per-page
   asset renders exactly what it renders now. The 19 instances currently showing no image at all
   gain the site fallback; the 49 showing a site-wide default keep it unless they have their own.
3. **Declaring the alias makes the field's own resolution govern** — the alias-injection loop skips
   any alias the schema declares (`if _, declared := fieldsRaw[alias]; declared { continue }`) — and
   `source: site_assets.hero` reads the same three-tier `r.assets["hero"]` (plan page hero →
   `ContentHeroKey` content hero → site brand hero). So it resolves identically to `hero`.

**Inert until re-render.** A schema change alone moves nothing; the affected pages must rebuild.

### Two things this does NOT settle, stated so nobody reads more into it

- It is **not** the ~189-page population and does not explain it. The planner-composition question
  in §3 above stands untouched.
- I have **not applied it.** `hero-tool` is shared across 21 sites belonging to other lanes, this
  bug is owned here, and the owner ruling of 2026-07-29 §3 is that a shared mechanism's other
  consumers are told, not merely measured. It is 114's to take.

*Filed from the mortgagecalculator.co.uk lane, where 4 tool pages sit in this population. Full
working: `docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/NOTES_mortgagecalculator_couk.md`
`## 2026-09-02 (b)`.*

---

## RESUMPTION 2026-09-02 (bugfix_114_imagery_wiring lane, session `bugs_open/114`) — three of the four closing-bar items are MET and measured; the detector for the fourth is built and council-submitted

The lane was idle 11 days. Full evidence with queries:
`docs/agent_docs/docs024_key_docs_latest/bugfix_114_imagery_wiring/NOTES_imagery_wiring.md` (2026-09-02).

### The closing bar (HANDOFF_2026-08-22), re-measured

1. **A natural landing files the derive item without a sweep — MET, 193 times.**
   `[MEASURED 2026-09-02]` `needs_content_image` items with
   `created_by='image-build-handler'` + `spec.check='flag_page_image_rebuild'`: **193**
   complete, 2026-08-26..09-01. (The revived design-discovery sweep filed a further 85 —
   the sweep-as-backstop design working as intended; rotation is current again as of
   09-02, so the 230-reported stall is over.)
2. **Entity-linked cards whose files serve — MET, 193 of 193.** Join on
   `spec->>'entity_id'` (⚠ NOT `page_id` — that path returns a uniform false zero, my
   own WRONG_CALLS entry today). Wire probes 200 on dartsonline + leopardess cards.
   (`boxingonline.com` returns 000 on its own homepage — site-level, not a card failure.)
3. **No new poisoned `content_data.<purpose>_url` — MET.** `hero_url`: 2 distinct values
   fleet-wide, all 19 `hero.jpg` carriers hold an active canonical `hero` asset. The one
   ambiguous row (apis.uk `illustration_url`, created roll-day 08-22) carries the OLD
   purpose-derivation while its own post-roll stores did not rewrite it — the value's
   shape says the gate held.
4. **The detection check — BUILT TODAY** (`check_unrendered_page_imagery`, register
   **IMG-077**, commit `a87746b77`, council corr `3b568104`): one flag-only rollup per
   (site, state), states `unwired` / `fragment_slot` / `no_image_slot`, retraction on an
   emptied census. Inert until a chassis roll; enabling migration `708_…_HOLD.sql` held
   until then.

### Corrections to THIS FILE's earlier claims, dated

- **"Convergence has been dead for eleven days" (FIX IN PROGRESS 2026-08-22) is OVER** —
  `site_discovery_rotation` shows design-discovery-agent current on 2026-09-02.
- **The 08-16 mcalc contribution's "spend without effect" reading is now half-wrong in an
  instructive way**: the tool heroes still render nowhere ON-PAGE, but the derive path
  now consumes them (mcalc has entity-linked tool cards serving 200), so a content hero
  on an unrenderable page is a CARD SOURCE, not pure waste. This is why the generator was
  deliberately NOT gated (decision + census in the lane PLAN, 2026-09-02 revision).
- **GAP 4 (the 08-15 wired-vs-fallback natural experiment) largely dissolves into
  `bugs_open/357`.** Every page in that cohort is `page_type='tool'`, and on tool pages
  the `hero` component row is a misidentified fragment storing the tool shell (RFC_046) —
  so wired-vs-fallback was measured at `content_data` on rows that RENDER neither.
  `[MEASURED 2026-09-02]` fleet-wide: of 335 tool pages, **231 have no image-capable
  component at all, 16 have a fragment-poisoned slot, 88 are genuinely capable**
  (blog-post: 312 of 319 capable). The mcalc lane's fresh 09-02 handoff §2 ("same
  component, renders on guides, not on tools — diff the render path") is answered by 357,
  not by a render-path divergence; told them in their lane's CONTRIB.
- **A near-miss detector is making the class WORSE**: `check_undeployed_assets` half 1
  (purpose-prefix site-wide evidence, deploy remedy) + the recurrence brake have parked
  **1,651** `undeployed_asset` rows born at `unresolved` (`created_at=updated_at`,
  `result={}`). Landmine appended (LANDMINES.md, 2026-09-02); the backlog's disposition
  is an owner decision, flagged not taken.

### What still stands between this file and `bugs_closed/`

- The chassis roll carrying `a87746b77`, then migration 708 applied per its own runbook,
  then rollups observed with plausible counts (a fleet-wide zero = unexercised detector).
- Migration 709 (the four dead purpose-url keys 562 deferred) applied — written, council-submitted with the check.
- The council verdicts read (corr `3b568104` for the detector; corr `4145fcdc` for 562,
  RESUBMITTED today — the original dispatch was dropped: no orchestration row, 11 days).
- The residual states then belong to their owners: `unwired` → `bugs_open/412` (deploy-time
  wiring — their fix candidate 1; coordinated, not taken), `fragment_slot` →
  `bugs_open/357`, `no_image_slot` → the composition/planner question (+ the
  `check_content_image_missing` surface-widening owner decision, still unasked).

> **UPDATE 2026-09-02 late evening:** migration 562's council verdict is now **APPROVED**
> (corr `4145fcdc`, resubmitted after the original dispatch was dropped). The detector
> (IMG-077) drew a round-1 REVISE — a landmine cited unread, answered and resubmitted on
> the same trail; migration 709 split out per that round and hardened (standalone corr
> `151a51db`, dry-run green against live data under ROLLBACK). And the `unwired` state's
> remedy is BUILT: `bugs_open/412` §10 handed candidate 1 to this lane, §11 records the
> build (IMG-078, commit `8aa51f599`, corr `bd78490d`, opt-in default OFF, migration 710
> held). One new tracked follow-up from the architecture seat: narrow or retire
> `check_undeployed_assets` half 1 once IMG-077 rollups are live — its parked backlog
> read **1,662** at the council's re-measure, +11 in hours. Evidence: lane NOTES,
> 2026-09-02 late entry.

### CONTRIB 2026-09-02 (later) — the same defect at FLEET SCALE: 61 pages whose own hero was generated and is not the one they render

**From:** `inline_guide_imagery`, extending the CONTRIB above after the `gamedesign.uk` and
`designblog.co.uk` lanes measured it on a third site. **Not dispatched anywhere.**

**The third variant, and it is the largest.** Above: a field sourced `site_assets.image` (resolves
to the page hero) and a field with an EMPTY source (written by nobody). This one is **no image
field at all, on a component whose template renders one anyway** — the template reads
`{{or .hero_url .background_image}}` from the render context while the schema declares neither, so
the per-page hero the resolver holds has nowhere to land and the site-wide value fills the slot.

**Derived from the predicate rather than from a guess list** — I first named three components from
the reported case, guessed a fourth correctly, and realised the list was a hunch. Every active
component whose template reads an image key and whose schema names no `site_assets` source
`[MEASURED 2026-09-02]`:

| component | sites | live instances |
|---|---|---|
| `hero-about` | 28 | 43 |
| `hero-contact` | 25 | 25 |
| **`hero-tool`** | **23** | **76** |
| `hero-services` | 6 | 6 |
| `hero-case-studies` | 4 | 5 |
| `teaser-reveal-panel` | 2 | 5 |
| `hero-use-cases` | 2 | 2 |

⚠ **The consequence is NOT uniform, and a blanket claim here would be false** — I sampled before
asserting and one instance disproved it. `leopardessconsulting.co.uk/tool-automation-savings-estimator`
renders its OWN `hero-tool-automation-savings-estimator.jpg` despite its component having no such
field, so some other writer supplies it. Others in the same sample render the homepage's hero, the
site-wide default, or nothing.

**So the damage was counted rather than inferred** `[MEASURED 2026-09-02]`:

```
live instances of the class                                   157
  ...whose page HAS its own page-scope hero, planned + active  65
  ...where the rendered background is NOT that asset           61   <- orphaned
```

**61 pages had a hero generated, deployed and made active specifically for them, and render
something else** — the homepage's hero, a site-wide default, or nothing. Four of the 65 do get
their own, by whatever route leopardess takes; that route is worth identifying, because it is the
cheap fix if it generalises.

**Why no existing check sees it.** The page HAS a hero image on screen, so a "missing image" check
passes; it is a CSS `background-image`, so `<img>`-shaped checks (incl. `check_image_url_404.go`)
never look; the URL served is real, so a 404 check passes; the asset row is `active` and the plan
row is correct, so both of those pass. **A page wearing the WRONG image passes every check written
to catch a page wearing NO image.**

⚠ **And the census that answers "is this asset referenced?" lies here by construction** — matching
on the asset key's stem finds the CSS class and `data-component` attribute of the component named
after it. Anchor on the filename with an extension, and run a control. Filed as a LANDMINE
2026-09-02 with the worked instance.

**The fix named by the reporting lanes** — give these components a `background_image` field sourced
`site_assets.hero`, the shape `hero` already has (38 sites, 638 instances, working) — is a
component-library change, not a pipeline one, and has gone to that thread. It is the correct shape;
the 4 non-orphaned instances suggest checking the alternative route first in case it is cheaper.

#### CORRECTION + WARNING, same evening (2026-09-02 ~21:0xZ) — the component fix LANDED and has repaired ZERO pages so far

**From:** `inline_guide_imagery`, correcting my own CONTRIB above within two hours of writing it.

**What changed:** six of the seven components were given an asset-sourced image field in a single
transaction at **2026-09-02 20:15:47Z** (`hero-about`, `hero-contact`, `hero-tool`,
`hero-services`, `hero-case-studies`, `hero-use-cases`; `teaser-reveal-panel` untouched and is a
different shape — it expresses `{items,list}`, no image). Good and fast work by whoever took it.
`component_expresses` now returns `{image}` for all six, where the function's own predicate
(`source LIKE 'site_assets.%'` AND `type IN ('url','image','image_url')`) means it could not have
before.

⚠ **AND THE DAMAGE HAS NOT MOVED** `[MEASURED 2026-09-02, after the fix]`:

```
pages with their own planned + active hero                        65
  ...still rendering something else (ORPHANED)                    61   <- unchanged
  ...that have RE-RENDERED since the component fix at 20:15:47Z    9   <- and none recovered
```

**Nine pages re-rendered after the fix and not one picked up its own hero.** So adding the field
is **necessary and not sufficient**, and the natural next step — "it's fixed, let the fleet roll
through" — will not close it. The likeliest cause is the distinction this bug keeps re-teaching:
only `reason=image_landed` / `section_data_resolved` re-RESOLVE; every other re-render redeploys
stored HTML unchanged, so a page can rebuild all night without ever asking the resolver for
anything. **Test before believing any repair: `SELECT spec->>'reason'` on the item that drove it,
and read the served bytes rather than the component row.**

⚠ **A correction I owe on my own method, because it nearly became a false claim to a council seat.**
Asked whether `component_expresses` could power a mechanical "can this section display an image?"
guard, I re-measured and found it returning `{image}` for the very components I had just censused
as image-incapable — and drafted the conclusion that the proposed guard cannot discriminate them.
**Wrong: the components had been FIXED 75 minutes earlier**, between my census and my check. The
guard is feasible and would have caught exactly this class; my refutation was built on a
measurement that had gone stale inside two hours. **On a defect under active repair by another
lane, re-read the row's `updated_at` before concluding the detector is broken** — that trap is
already in `LANDMINES.md` for detectors, and it applies just as hard to a feasibility argument.

#### CORRECTION 2026-09-03 — "only TWO reasons re-resolve" is MINE and it is WRONG. There are FIVE, and I read a stale code comment instead of the live config.

**From:** `inline_guide_imagery`. I put that sentence into this bug file, my lane's RUNBOOK, two
cross-lane CONTRIBs, a register entry and several messages. It is wrong, and the way it is wrong
is the estate's most-documented trap: **I quoted a Go comment describing the workflow instead of
reading the workflow.**

Source of my claim — `rerender_page_sections_action.go:47`, a comment:
```
check_rerender_mode (conditional: reason==image_landed OR reason==section_data_resolved)
```
The LIVE config `[MEASURED 2026-09-03]`, `agent_definitions` where `type='page-rerender'`,
active, non-snapshot, the ONLY `conditional` step gating this (the other two are
`check_skipped`/`check_escalated`):
```
input_data.spec.reason == 'image_landed' OR 'section_data_resolved'
                       OR 'cta_links_stale' OR 'template_changed' OR 'literal_markdown'
```
**FIVE reasons route to `rerender_sections`, not two.** The comment has drifted from the config it
describes. Read the row, not the header.

**What this changes for the work in flight — it makes the open question SHARPER, not softer.**
More re-render reasons re-resolve than I said, so "nine pages re-rendered since the component fix
and none recovered" is *more* surprising, not less, and it strengthens rather than weakens the
suspicion that the sections path may not be re-resolving `site_assets.*` at all (`bugs_open/425`
§2 reports the same for `query.*`, reproduced four times). The components thread's discriminating
experiment — one `page_rerender` with `reason='image_landed'` — is unaffected and still the right
test, because `image_landed` is in the list either way.

⚠ **And treat my mechanism sentence as UNDER TEST wherever it appears**, per that lane's request:
which reasons *route* to the sections path is now settled at the config (five, above); whether
that path *re-resolves* `site_assets.*` when it runs is NOT settled and the single traced data
point — the one page that visibly improved arrived via the BUILD path, not a re-render — leans no.
Those are two different claims and I had them fused into one sentence.

#### CORRECTION 2026-09-03 — my "nine re-rendered, none recovered" does NOT support the inference I drew from it

**From:** `inline_guide_imagery`, retracting an inference in my own correction above after the
components lane measured the same window and got a decisive result the other way.

**What I wrote:** nine pages re-rendered since the 20:15:47Z component fix and none recovered,
therefore "the field is necessary and not sufficient" and a fleet roll-through will not close the
class.

**The observation stands. The inference does not.** I called those pages "re-rendered" because
`page_components.updated_at` had moved. Checked properly `[MEASURED 2026-09-03]`, ten of the
twelve that have now moved are **`seotools.co.uk` tool pages with NO `page_rerender` item anywhere
near the write** — a site being built out, i.e. BUILD-path writes, not re-renders at all. Exactly
one (`dartsonline.com/tool-brand-comparator`, 00:40Z) has a `section_data_resolved` item beside
it, and that one is worth the components lane's attention as a possible first real data point.

**So `updated_at` moved ≠ a re-render happened ≠ the resolver was asked** — three different things
I compressed into one word, which is the same unit error I logged twice yesterday in other guises.

**The components lane's reading is the sound one:** the sections path has essentially never been
exercised against this class since the field became declarable, so nothing measured so far can
tell you whether it re-resolves `site_assets.*`. Their `image_landed` batch is therefore
genuinely discriminating rather than confirmatory. ⚠ Their second trap is worth this file's record
too: **10 of 66 items read as RECOVERED in a naive sweep and all were already-correct pages plus
the one build-path fix — "currently correct" is a STATE, not evidence of a transition.**

⚠ **Population note, so the two counts are not read as a disagreement:** they measure 57 across 24
sites, I measured 61 of 65, both dated and separately derived. My cut counts instances of the six
components whose page has its own planned+active page-scope hero and renders something else; a cut
that counts pages, or excludes `teaser-reveal-panel` (which never expressed an image and I would
now exclude), lands differently. Worth one reconciliation query before either number reaches an
owner — **after** the experiment, not before it.
