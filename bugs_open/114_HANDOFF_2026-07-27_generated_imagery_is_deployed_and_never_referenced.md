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
