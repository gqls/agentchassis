# 128 — `image_url_404` flagged the image that serves 200 and missed both that 404

**Filed:** 2026-07-28 by the brochure_component_library workstream, after the owner asked
what the design-discovery sweep had missed. The sweep ran cleanly and reported 13 items;
this is about the one it got backwards and the two it did not report at all.
**Severity:** Medium. Nothing is destroyed, but a check named for HTTP status reports
none, and its **silence is read as "no broken images"** — three deployed pages carry a
404 image right now and the sweep that just ran says nothing about them.
**Class:** a check whose name describes a capability it does not have.
**Status:** OPEN, unowned.

---

## Symptom, measured 2026-07-28

Design-discovery sweep on `fundamentallyai.com` (correlation `7070eb38`), COMPLETED,
13 items filed. Exactly one `image_url_404`:

```
item_key: image_url_404:brand-illustration
summary:  Pages reference unknown image /assets/images/brand-illustration.jpg
spec:     {"path": "/assets/images/brand-illustration.jpg", ...}
```

Probed twice, spaced:

| URL | HTTP | asset row | flagged? |
|---|---|---|---|
| `/assets/images/brand-illustration.jpg` | **200** | none | **yes** |
| `/assets/images/hero.jpg` | **404** | none | **no** |
| `/assets/images/favicon.png` | **404** | none | **no** |

**An exact inversion:** the working image was reported, both broken ones were not.

## Defect 1 — the name promises HTTP and the check never makes a request

From the check's own header (`check_image_url_404.go:8-13`):

> *"This is the lightweight DB-only version (per PLAN section 1.3): we don't check
> whether the file is actually in git, only whether an assets row exists. The full
> HTTP/git version would catch deployment failures too, but is deferred until we have
> git-adapter integration on the discovery path."*

So it is an **asset-registry** check wearing an HTTP name. That is not a private naming
quibble — it changes how every consumer reads it:

- a finding reads as *"this URL 404s"*. Here it meant *"this URL has no assets row"*,
  and the URL is fine.
- **its absence reads as "no broken images on this site"**, which is the expensive half.
  Two are broken.

`brand-illustration.jpg` is therefore arguably a *correct* finding of a *different*
thing — an untracked asset — reported under a name that makes it look wrong.

## Defect 2 — `hero.jpg` meets the check's OWN criterion and was still missed

This one is not explained by the naming, and it is the reason this is a bug rather than a
rename.

`/assets/images/hero.jpg` has **no `assets` row** and appears in
`page_components.rendered_html` on **three deployed pages**:

```sql
SELECT p.name, p.build_status, cc.function
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  JOIN content_components cc ON cc.id=pc.component_id
 WHERE s.domain='fundamentallyai.com' AND pc.rendered_html LIKE '%/assets/images/hero.jpg%';
-- tool-model-approach-selector-guide | deployed | hero
-- llm-cost-calculator                | deployed | tool-guide-intro
-- self-correction-leopardessconsulting | deployed | hero
```

That is precisely "rendered HTML references an `/assets/images/*` path with no assets
row" — the check's stated remit — and no item was filed.

**Dedup is ruled out**, which was the obvious alternative explanation: the check keys per
basename (`ItemKey: fmt.Sprintf("image_url_404:%s", basename)`, `:123`/`:154`), and there
is **no** `image_url_404:hero` row on this site at any status. It is not blocked, it was
never created.

**[HYPOTHESIS, not verified — read the code before acting on it]** the header describes
two emission branches: *"the path's basename is matched against known purpose names. If a
known purpose, route to image-build-handler. Otherwise emit a flag-only finding."*
`hero` is a known purpose and `brand-illustration` is not, so the two took different
branches and only the flag-only branch produced a row. If so, **every well-known image
purpose — the ones most likely to be on every page — is exactly the set the check
silently drops.** That would make the miss systematic rather than incidental. Confirm at
`check_image_url_404.go:100-160` before believing it.

## Defect 3 — the chrome surface may not be scanned at all

`favicon.png` is referenced by the shared `head` chrome in `site_components`, not by any
`page_components` row, and it 404s on every page of the site. If the check only scans
`page_components.rendered_html`, an entire surface — the one thing that appears on
**every** page — is invisible to it. Same shape as `bugs_open/098` (a repair path that
excludes a population on purpose, and nothing owns the intersection).

## Fix candidates, ordered by what closes the door

1. **Rename it for what it does** (`image_asset_unregistered`) and file the HTTP check
   separately. Cheapest, and it stops the silence being misread. A check that cannot
   answer the question in its name should not carry the name.
2. **Fix the missed branch** (defect 2) — worth more than 1, because a known-purpose
   image is the common case.
3. **Do the HTTP half.** Its own header defers it on git-adapter integration; but
   `scripts/render_audit.py` already fetches every image a page renders and reports
   non-200, so the capability exists outside the pipeline. `features_open/026` Phase 3 is
   where this belongs rather than a fourth parallel implementation.
4. Scan `site_components` as well as `page_components`.

## How to verify a fix

**Against these three URLs specifically**, because they are a natural positive/negative
set on one live site: `hero.jpg` (404, unregistered, referenced on 3 pages) must be
reported; `favicon.png` (404, unregistered, chrome-only) must be reported; and whatever
`brand-illustration.jpg` (200, unregistered) is reported as must not be called a 404.
A fix that reports all three identically has not distinguished anything.

## What this is NOT

- Not a failure of the sweep as a whole — it ran, completed, and filed 13 items including
  two `needs_imagery` rows that correctly caught the calculator page having no image of
  its own. This is one check inside it.
- Not `bugs_open/114` (imagery generated and never referenced) — that is the opposite
  direction: assets that exist and nothing points at them. This is references with no
  asset.
