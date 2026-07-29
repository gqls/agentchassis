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

---

## Diagnosis 2026-07-29 — the filed hypothesis is REFUTED, and the real mechanism is worse: the check is blind to 83% of the surface it is named for

Session "bugsearch 6". Code read first, as § Defect 2 asked. Not a fix — a diagnosis,
three fact corrections, and a measured blast radius.

### § Defect 2's hypothesis is wrong, and it matters that it is wrong

The filed `[HYPOTHESIS]` was: *"`hero` is a known purpose and `brand-illustration` is
not, so the two took different branches and only the flag-only branch produced a row."*

**Both branches emit a work item.** `check_image_url_404.go:111-128` emits the flag-only
`image_url_404` item; `:142-156` emits a `needs_hero_image` item routed to
`image-build-handler` — with the **same** `ItemKey` (`image_url_404:hero`, `:154`), which
is why the file's dedup ruling-out was searching the right key. Routing cannot explain
the miss. Had the recognised branch fired, an item would exist.

**The miss happens BEFORE the branch, at `:77-87`:**

```go
if knownPurposes[basename] { continue }   // <- hero.jpg dies here
root := basename
if idx := strings.Index(basename, "-"); idx > 0 { root = basename[:idx] }
if knownPurposes[root] { continue }       // <- hero-anything.jpg dies here
```

`knownPurposes` is built by `loadKnownAssetPurposes` (`:207-230`) as
`SELECT DISTINCT purpose FROM assets WHERE site_id=$1 AND purpose IS NOT NULL AND status='active'`.
**It is a set of PURPOSE NAMES, and it is compared against a rendered FILE PATH's
basename.** Verified on the reporting site:

```
fundamentallyai.com active asset purposes: hero (x4, all S3 URLs), icon (x13), og_card
```

So `knownPurposes["hero"]` is true, and `/assets/images/hero.jpg` is skipped — because
the site owns *a* hero asset somewhere, which says nothing whatever about whether that
path resolves. **The real defect: owning one asset of a purpose makes the check blind to
every rendered path sharing that purpose's name** — and the `root` fallback at `:82-87`
widens it to the whole `hero-*`, `icon-*`, `logo-*` prefix space.

### Blast radius, measured fleet-wide rather than argued

Every distinct `/assets/images/*` basename in deployed unlocked `page_components`,
tagged by whether the purpose-skip masks it:

| | distinct paths | sites |
|---|---|---|
| **masked — the check CANNOT report these** | **79** | **13** |
| checkable | 16 | 4 |

**The check is structurally blind to 79 of 95 paths — 83% of the surface its name
claims.** Then probed all 79 over HTTP: **73 serve 200, and SIX ARE LIVE 404s:**

```
404  gamesdesign.co.uk   /assets/images/hero.jpg
404  idea.uk             /assets/images/hero.jpg
404  oufe.com            /assets/images/hero.jpg
404  relojistas.com      /assets/images/hero.jpg
404  vonc.com            /assets/images/hero.jpg
404  webdesign.co.uk     /assets/images/hero.jpg
```

Six deployed sites painting a broken image, and **no sweep can ever surface them**,
because every one is masked by its own site's `hero` asset. That root cause is already
traced — see the contribution added to `bugs_open/114` today; it is a legacy site-wide
`sites.content_data.hero_url` default, and **10 live sites still carry it**.

### The check WORKS where it can see — which is the fair half of the story

finetuning.uk's five `case-study-*.jpg` references (root `case`, not a purpose) are all
live 404s and were **all five correctly flagged** on 2026-07-26, `item_type
image_url_404`, status `detected`. The check is not broken in general; it is scoped by a
predicate that removes most of its subject.

`ai-agent-orchestration.com` has five identical live 404s (`case-study-*.png`) and no
`image_url_404` item — **that one is NOT this bug.** Its `design` pipeline discovery last
ran **2026-04-08**; `content` and `build` ran 07-24/07-26. Three and a half months of no
design sweep, so the check never ran there. A cadence gap, of the `bugs_open/083` family.

### § Defect 3 — CONFIRMED, by code and by data

`collectImagePathReferences` (`:165-202`) queries `page_components` only.
`site_components` — the chrome that appears on **every** page — is never scanned:

```
favicon.png in page_components: 0     in site_components: 1
```

### Three fact corrections — re-probe before quoting this file

1. **`favicon.png` now serves 200**, not 404: 6,970 bytes, `image/png`, confirmed with a
   cache-buster and `cf-cache-status: BYPASS`. It was repaired between 07-28 and 07-29.
2. **`hero.jpg` is no longer referenced on any fundamentallyai deployed page** — the
   brochure lane repaired it data-level at 09:30 on 07-29 (`69a93dffc`, under `114`). The
   URL still 404s; nothing points at it *on that site* any more.
3. **Therefore § "How to verify a fix" is stale as an acceptance set.** Two of its three
   URLs no longer have the properties it relies on. A replacement triple, live as of
   2026-07-29 and drawn from three different sites so it cannot be repaired out from
   under you by one content fix:
   - `vonc.com/assets/images/hero.jpg` — **404, masked by purpose skip** → must be reported;
   - `finetuning.uk/assets/images/case-study-legal-rag.jpg` — **404, already correctly
     reported** → must stay reported (regression guard);
   - `fundamentallyai.com/assets/images/brand-illustration.jpg` — **200, unregistered** →
     must not be described as a 404.

### Fix candidates, revised against what is now known

1. **Compare the rendered path against asset PATHS, not purposes** (new, and it is the
   actual defect). `assets.url` already holds local paths for some rows — e.g.
   fundamentallyai's `icon` row is literally `/assets/images/input-data.asset-key.jpg` —
   so a path-based predicate is possible today without new plumbing. This is candidate 2
   in the original list, correctly diagnosed: it recovers 79 paths and 6 real 404s.
2. **Rename to `image_asset_unregistered`** (original candidate 1) — still right, and now
   *more* obviously right: the check's finding set is "paths whose basename is not an
   asset purpose", which is neither HTTP nor really registration.
3. **Scan `site_components` too** (original candidate 4).
4. **The HTTP half has a standing objection, so do not treat it as free.**
   `discovery_checks/verifier_coverage_test.go:171` records `image_url_404` as
   *"deliberately NOT a verifier candidate: verification would add an outbound HTTP call
   to the completion path"*. `scripts/render_audit.py` already probes served images
   outside the pipeline; `features_open/026` Phase 3 remains the right venue.

**Status: still OPEN and still unowned.** `who-owns.py` reports the filing workstream
(`brochure_component_library`) as owner because it is active and mentions `128` twelve
times — but that lane's own `HANDOFF_2026-07-28b` lists 128 as *"read, still unowned"*
and *"untouched"*. **A filing workstream is not an owning workstream**; check the named
lane's handoff before believing the tool in either direction.
