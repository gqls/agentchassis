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
