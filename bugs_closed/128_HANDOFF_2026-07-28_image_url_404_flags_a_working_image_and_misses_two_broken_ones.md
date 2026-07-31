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

---

## Contribution 2026-07-29 (leopardess lane) — a THIRD blind spot: `src=""` is not a URL

Found while fixing missing images on `leopardessconsulting.co.uk`. Not a new bug —
it is the same check's blindness, one class wider than either defect above, so it is
recorded here rather than filed separately.

**`/blog.html` served six `<img src="" alt="..." loading="lazy">` tags** — one per card
in the blog listing. Browsers render an empty `src` as a broken-image icon, and per the
HTML spec resolve it against the current document, so each card also re-requested the
page itself as an image. Measured before the fix:

```
$ curl -s https://leopardessconsulting.co.uk/blog.html | grep -c '<img src=""'
6
```

**Why this matters for 128 specifically:** every defect and every fix candidate above
reasons about a *path* — is it registered, does it 404, does it match a purpose or a
path. An empty `src` has **no path to reason about**. It is invisible to:

- the DB-only check as built (nothing to match against `assets`);
- the **path-based predicate of fix candidate 1** (the empty string matches no asset row,
  and would have to be special-cased not to be silently skipped);
- **the HTTP half of candidate 4** — probing `""` either resolves to the page URL and
  returns **200**, or is skipped as unfetchable. *An HTTP checker would actively confirm
  this broken image as healthy.* That is worth pinning before the HTTP half is built,
  because it is the one case where adding the outbound request makes the report worse
  rather than merely incomplete.

So the acceptance set for any fix here needs a fourth case alongside the triple above:
**an `<img>` whose `src` is empty must be reported, and must not be probed.** The cheap
predicate is structural, not networked — flag `src=""`/`src="#"`/whitespace-only in
rendered HTML — and it needs no git-adapter integration, so it is not blocked by the
standing objection in `verifier_coverage_test.go:171`.

**Root cause of the leopardess instance, for reference** (fixed 2026-07-29, live):
`rebuild_blog_listing_action.go:218` hardcodes `"image": ""` for every article — there is
no per-article imagery on this platform yet — and the shared `content-listing` /
`category-listing` templates in `content_components` emitted `<img src="{{.image}}">`
**unconditionally**. Fixed by guarding the wrapper with `{{if .image}}` in both templates,
so an article with no image now emits no `<img>` at all. Blast radius measured before
applying: the only other consumer with articles (`robot-hands.com/learning-center-hub`,
3 articles) has a non-empty image on all three, so its output is byte-identical.
The Go half is untouched and still writes `""` — the template is now simply honest
about it. **Real card thumbnails remain unbuilt** (per-card imagery, the leopardess
lane's "Phase I3").

---

## FIXED 2026-07-31 — the predicate compares PATHS now, and the fix was a reuse rather than a new predicate

Session `a4917c55` (lane `docs024_key_docs_latest/bugfix_128_image_url_404/`). Commit
`beff42809`. Council `99dca96a-413a-4bcb-b278-9577f920786d`.

### The whole thing re-measured live before touching anything

Not argued from the 07-29 numbers — re-derived today, because a figure carried forward
unchecked is how a stale premise becomes a bug. All **127** distinct `/assets/images/*`
paths rendered by deployed unlocked page components across **13** live sites, each probed
over HTTP as ground truth (four transient `000`s re-probed and confirmed 200; the
webdesign hero re-probed twice with a cache-buster):

| predicate | reports a WORKING image | reports a BROKEN one | SILENT on a broken one |
|---|---|---|---|
| purpose/prefix skip (as shipped) | **21** | 11 | **6** |
| `storage.DeployedWebPath` (the fix) | **1** | **17** | **0** |

The 07-29 diagnosis is **CONFIRMED in full**, including the six masked live 404s. Its
count of "79 of 95 masked" is not directly comparable — today's population is 127 paths,
because the fleet has moved — but the mechanism and its consequences reproduce exactly.

### What the fix is

`storage.DeployedWebPath(asset_key, purpose)` already existed as the platform's single
source of truth for *"the web path a generated asset is committed to and served from"*.
Five writers resolve through it (`plan_sections_action`, `render_site_components_action`,
`emit_sprite_css_action`, `derive_card_asset_action`, `queryresolve`) and
`deploy_image_asset_action` commits to exactly that path via the shared
`storage.AssetKeyFilename`. The check now resolves through the same helper, which makes it
**the exact inverse of the render-time resolver** — it cannot drift from the writers
without the writers changing first. Still DB-only; the standing objection at
`verifier_coverage_test.go:171` (no outbound HTTP on the completion path) is kept.

### Why the 07-29 refutation of "compare paths" was right, and why this is not that

That session measured a path predicate against `assets.url` / `filename` and refuted it —
correctly. Those columns cannot answer the question: of 267 active asset rows measured
today, `url` is a presigned S3 link on **152**, `filename` is empty on **191**,
`storage_path` on **189**, and of the 115 whose `url` does look like a local path, **47
are the unresolved template literal** `/assets/images/input-data.asset-key.jpg`. The
served path is **derived, not stored**. That distinction is the whole fix, and it is now
in `LANDMINES.md`.

### Each defect, and what happened to it

1. **Defect 1 (the name promises HTTP)** — resolved by making the check able to answer the
   question rather than by renaming it. A finding now means "no active asset of this site
   deploys to this path", which is the 404 claim. **The check is deliberately NOT renamed
   to `image_asset_unregistered`** (candidate 1/2 of the two lists above): the name is
   looked up in `agent_definitions.default_config` for `design-discovery-agent`, so
   renaming the registered check disables it silently until a config migration lands. The
   cheaper half of that candidate — honest wording — is done: summaries now read
   *"Pages reference X but no active asset deploys to that path"*.
2. **Defect 2 (the missed branch)** — this was the real defect and it is fixed. Note the
   07-29 correction stands: the miss happened *before* the branch, at the purpose skip.
3. **Defect 3 (chrome never scanned)** — fixed. `site_components` is scanned alongside
   `page_components`. Measured on the chrome surface (38 paths, 13 sites): **+2 real
   findings, 0 false positives** — `idea.uk`'s `/assets/images/favicon.png` and
   `/assets/images/og-card.png`, both 404 **on every page of the site**, neither
   reportable before. Chrome findings are severity `high`; page findings `medium`.
4. **The leopardess contribution's fourth case (`src=""`)** — fixed, structurally, never
   probed. Three are live right now (ai-agent-orchestration ×2, finetuning ×1). Emitted as
   one item per site under the **existing** `image_url_404` item type, so nothing is added
   to the shared work-item vocabulary; `spec.kind` (`unbacked_path` / `empty_src`) tells
   the shapes apart.

### The trap the masking test warned about, answered by deletion

`check_image_url_404_masking_test.go` warned that unmasking would activate
`knownPurposeMapping`'s routing to `image-build-handler`, *"a dormant fleet-wide
auto-regeneration path"*. **That branch is gone, because it was a duplicate.**
`check_placeholder_image_in_use` already owns exactly that repair: same two paths
(`/assets/images/hero.jpg`, `/assets/images/logo.png`), same purposes, same
`needs_hero_image`/`needs_logo` item types, same handler, same severity, same prompt
recovery, same precondition — and **both checks are enabled on the same agent**. They
differ only in `item_key`, so they would have filed two work items for one repair. Neither
has ever fired: `placeholder_image_in_use:%` returns **0** rows fleet-wide and none of the
13 `image_url_404:%` rows carries a routed type. So the fix **adds findings and no new
autonomous repair** — every emission is flag-only.

### One residual false positive, named rather than hidden

`webdesign.co.uk/assets/images/hero.jpg` — 455KB, serves 200, backed by no asset row. A
file committed by no pipeline is invisible to a DB-only check, and it is arguably a true
finding of a different thing: an untracked file nothing maintains and any repo
reconciliation would delete. **1 of 127.**

### Stated cost of the dedup-key change

The item key now carries the extension (`image_url_404:logo.png`, was
`image_url_404:logo`), because `fundamentallyai.com` serves `logo.jpg` (200) and
`logo.png` (404) — two files, one basename, opposite results — and `idx_swi_dedup` is
UNIQUE on `(site_id, item_key)` for non-terminal statuses, so an extension-blind key
silently drops the second finding (`bugs_open/091`'s failure mode). **Six existing
`detected` rows keep the old key and will not dedup against the new ones.** They are
already unclearable (flag-only, no handler, `bugs_open/083`). A sweep may cancel them:

```
finetuning.uk        image_url_404:case-study-facilities          (still a true 404)
finetuning.uk        image_url_404:case-study-financial-data      (still a true 404)
finetuning.uk        image_url_404:case-study-legal-rag           (still a true 404)
finetuning.uk        image_url_404:case-study-logistics-strategy  (still a true 404)
finetuning.uk        image_url_404:case-study-private-ai          (still a true 404)
fundamentallyai.com  image_url_404:brand-illustration             (200 — the old check's false positive)
```

Three further `detected` rows on `robot-hands.com` (`image_url_404:content-hero-tool-*`)
were the old predicate's **false positives** — all three serve 200 and the new predicate
is silent on them. They can be cancelled outright.

### What this check still does NOT answer, deliberately

An asset row that exists but whose **file was never deployed**. `gaswholesalers.com`'s
`/assets/images/logo.png` is the live case: healthy `logo` asset row, 404 on the wire, and
this check is silent by design because the path resolves as far as the database is
concerned. That is `check_undeployed_assets`' remit (`bugs_open/142`). Conflating the two
is how one check ends up asserting the other's finding, so the division is written into
both the check header and `LANDMINES.md`.

### How to verify once a chassis image ≥ `beff42809` is rolled

The 07-29 acceptance triple is restated because it had already gone stale once. Live and
re-probed 2026-07-31:

1. `vonc.com/assets/images/hero.jpg` — **404, was masked by the purpose skip** → must be
   reported. (Same for dartsonline, gamesdesign, idea.uk, oufe, relojistas.)
2. `finetuning.uk/assets/images/case-study-legal-rag.jpg` — **404, was already correctly
   reported** → must stay reported. Regression guard.
3. `fundamentallyai.com/assets/images/brand-illustration.jpg` — **200, unregistered** →
   must NOT be reported at all now (the old check called it a 404).
4. `idea.uk/assets/images/og-card.png` — **404, chrome-only** → must be reported, severity
   `high`. This is the surface that did not exist before.
5. `ai-agent-orchestration.com` — must carry one `image_url_404:empty-src` item counting
   its two empty `<img src="">` tags.

```sql
-- after the roll, run a design discovery sweep and then:
SELECT s.domain, w.item_key, w.severity, w.spec->>'kind', left(w.summary,80)
  FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE w.item_type='image_url_404' AND w.created_at > '2026-07-31'
 ORDER BY 1,2;
```

Expected shape fleet-wide, from today's measurement: **18 page findings + 2 chrome
findings**, of which 17 + 2 are live 404s and one (webdesign's legacy hero) is the named
false positive.

### Council record — REVISE at round 1, answered; round 2 could not sit

`SUBMISSION_CORR=99dca96a-413a-4bcb-b278-9577f920786d`. Round 1: 2 approve
(`reuse_agent`, `guidelines`), 2 object, 4 abstained, `decided_by: gating objection from
bug_historian`.

- **bug_historian (HIGH):** does adopting `storage.DeployedWebPath` inherit a
  shared-mechanism defect patched at one call site? **Answered by audit, and the audit
  is in the code** (`6d3992213`): it is one landmine entry not three (six `doc_notes`
  rows, one body, footprint-expanded); the drift needs an underscore purpose stored with
  `asset_key == purpose`, and over 267 active rows that set is exactly the two
  brand-head purposes, which the code branches on; the deployer branches identically;
  `deploy_path` has never been passed in any orchestration; and 109 of 127 rendered paths
  matched with all 109 serving 200.
- **editquality (medium ×3, low):** bundling. A defect in the **submission**, not the
  change — all four extra edits are in this file, three are causally forced by the core
  fix, and round 2 grounds each in a quote. Note `reuse_agent` **approved** the very edit
  `editquality` objected to, calling it *"precisely the unify-don't-duplicate move"*.

**Round 2 returned no verdict.** It terminated `complete_invalid`, which looks like a
schema refusal and is not — `collected_data->'__step_error'` reads *"You have reached
your specified API usage limits. You will regain access on 2026-08-01 at 00:00 UTC"*
(`bugs_open/130`'s cap). **So the verdict of record is still REVISE**, both commits carry
`Council-Submitted:` rather than `Council-Reviewed:`, and the round-2 answers above are
**not approved**. Resubmit `submission_128_r2.json` with the same `RESUBMIT_CORR` after
00:00 UTC.

### Status

**FIXED IN CODE, NOT YET LIVE — stays in `/bugs_open/` per the standing bar.** Commits
`beff42809` (fix + 10 tests) and `6d3992213` (the drift audit). Inert until a chassis
image ≥ `6d3992213` is built and rolled; until then the six sites above are still
painting a broken hero and `idea.uk` is still serving a 404 favicon and og-card on every
page. Close it when the acceptance set at the end of the previous section passes against
a rolled pod.

---

## CLOSED 2026-07-31 — LIVE on v1.0.1219, pod-verified on both replicas, and PROVEN against live sites

### The image carries the fix, established with a negative control rather than a roll

A roll is not evidence a fix shipped, so the grep asserts three things in one exec — a
string the change ADDED, the string it REMOVED, and an untouched control:

```
                                              t7dgn   z84n8
no active asset deploys to that path            2       2     ← added by this fix
render with no image source                     1       1     ← added by this fix
site_components scan failed                     1       1     ← the chrome surface
Pages reference unknown image                   0       0     ← REMOVED: old code is gone
Pages reference fallback                        1       1     ← control: sibling check intact
```

The zero is the load-bearing line: new code being present would not, on its own, rule out
the old predicate still being there beside it.

### Behaviour proven on live sites, not inferred from the binary

Fired `design-discovery-agent` at `vonc.com` (`34c0ce39`) and `idea.uk` (`1f23e3e2`); both
COMPLETED. Items filed:

| site | item_key | severity | kind | surface | handler |
|---|---|---|---|---|---|
| idea.uk | `image_url_404:favicon.png` | high | unbacked_path | **chrome** | *(none)* |
| idea.uk | `image_url_404:og-card.png` | high | unbacked_path | **chrome** | *(none)* |
| idea.uk | `image_url_404:hero.jpg` | medium | unbacked_path | page | *(none)* |
| vonc.com | `image_url_404:hero.jpg` | medium | unbacked_path | page | *(none)* |

Against the acceptance set:

1. **`vonc.com/assets/images/hero.jpg` — REPORTED.** This is the bug. It was masked by
   vonc's own six `hero`-purpose assets and no sweep could surface it. ✓
2. **`idea.uk` favicon.png + og-card.png — REPORTED, severity `high`, surface `chrome`.**
   The surface that did not exist before; both 404 on every page of the site. ✓
3. **Every item is flag-only** (`handler_agent` empty). The de-duplication invariant
   against `check_placeholder_image_in_use` holds in production, not just in the test. ✓
4. **Item keys carry the extension.** ✓
5. **The negative half, which is the half that condemned the old check:** 19 rendered page
   paths were scanned across the two sites and **2 were reported** — both genuine live
   404s. The other 17 all serve 200 and the check was silent on every one. No `empty-src`
   item on either site, correctly: the three live empty `<img>` tags are on
   `ai-agent-orchestration.com` and `finetuning.uk`.

### Status: FIXED, LIVE, PROVEN → `bugs_closed/`

Commits `beff42809` (fix + 10 tests), `6d3992213` (the drift audit the council asked
for), `b51e4879d` (docs). Live in `v1.0.1219`.

**One thing is still owed and it is not a defect in the fix.** The council verdict of
record is round 1's **REVISE**; round 2 answered both objections but could not sit,
because the panel hit the API cap until 2026-08-01 00:00 UTC (`bugs_open/130`). Both
commits carry `Council-Submitted:`, which asserts nothing, and `098` credits them
automatically when a verdict lands. **Resubmit `submission_128_r2.json` with
`RESUBMIT_CORR=99dca96a-413a-4bcb-b278-9577f920786d` after the cap lifts** — the lane's
NOTES carries the command. Closing the bug on proven behaviour rather than holding it for
a review that cannot run is the right way round; the review is owed on the *change*, and
it is tracked.

### Left deliberately for someone else

- **Nine stale `detected` rows** under the old extension-less key: five `finetuning.uk`
  `case-study-*` (still true 404s, will re-file under the new key), one
  `fundamentallyai.com` `brand-illustration` and three `robot-hands.com`
  `content-hero-tool-*` — **those four were the old predicate's false positives and can be
  cancelled outright.** Not touched here: `bugs_open/083`'s lane is actively working the
  `detected` population and cancelling rows underneath it would be rude and confusing.
- **The `storage.DeployedWebPath` durable fix.** A *new* purpose containing an underscore,
  stored with `asset_key == purpose`, would still be mis-rendered by the helper. The audit
  proves that set is empty today over all 267 active rows; it cannot prove it for assets
  that do not exist yet. That fix belongs in `platform/storage` with five other consumers
  — architecture-scope, the shape `bugs_closed/124` was vetoed for — so it is named in the
  round-2 submission's risks and in `LANDMINES.md`, not smuggled in here.

---

## Council: APPROVED at round 2 (2026-08-01, once the API cap lifted) — and the advisories discharged

`SUBMISSION_CORR=99dca96a-413a-4bcb-b278-9577f920786d`, run `06573962`. **12 approve,
3 object, 3 abstained. `decided_by: approved with 3 advisory objection(s) — none
high-severity`.** The round-1 gating objection from `bug_historian` was withdrawn in its
own words: *"This round is a direct, evidenced answer to my own prior gating objection…
correctly declining architecture-scope creep, citing bugs_closed/124's precedent against
exactly that."*

An approval is not a reason to stop reading. Three seats asked for things that were cheap
and checkable, so they were checked rather than banked:

**`prior_art_librarian` + `editquality` (medium ×2): "load-bearing absence claims asserted
by the author, not independently verified here."** Correct — a seat cannot run SQL. Re-run
2026-08-01, output verbatim:

```
 placeholder_items_all_history      0     <- the duplicate branch's sibling has NEVER fired
 image_url_404 | 17                       <- every row; zero needs_hero_image, zero needs_logo
 orchestrations_with_deploy_path    0     <- the override that could produce a third spelling
 landmine_rows 6 | distinct_bodies  1     <- ONE entry, footprint-expanded; not three
```

(17, not the 13 quoted earlier: the four this fix filed during acceptance are included.)

**`guardian` (medium, edit 4): "confirm no other discovery check already scans
`site_components` under a different lock/status contract this edit could shadow or
double-count against."** Not checked before submission; checked now, and the answer
inverts the concern. **Fourteen other discovery checks already scan `site_components`** —
`check_dead_controls`, `check_phantom_internal_links`, `check_voice_tells`,
`check_broken_nav_links`, `check_undeployed_assets`, `check_integrity` and eight more. This
check was the **outlier for not scanning it**, which is the defect. And its predicate is
*stricter* than the fleet norm, not looser: the others filter only
`site_id` + `rendered_html IS NOT NULL`, while this one also requires `locked_at IS NULL`.
Double-counting is not possible across checks — each owns its own `item_type` and its own
`item_key` namespace, so an `image_url_404` finding cannot collide with a `dead_control`.

**`guardian` + `editquality` (low, edit 5): "does the item_key format change interact with
two-strike / `recurrenceExpected` semantics?"** `load_work_item_actions.go:1164` gates the
two-strike rule on `item.itemKey != "" && !item.recurrenceExpected`, counting prior rows
at `complete`/`failed` in 7 days and rewriting the summary to *"[unresolved after N
attempts]"*. These items carry **no handler agent**, so they cannot be claimed, cannot be
dispatched and cannot reach a terminal status by any handler run — the counter cannot be
incremented by the platform's own machinery. The key change resets it for the affected
keys, which can only make a **false** "[unresolved after N attempts]" label *less* likely.
Benign in the one direction it can move.

**`bug_historian` (low, edit 1): "the durable fix is deferred as 'its own item' rather
than filed."** Fair, and the failure mode it names is real — saying "its own item" and not
creating one is how a deferral disappears. **Filed as `bugs_open/168`**, with the measured
empty risk set, four fix candidates ordered by what closes the door, and the scope warning
that it is a six-consumer shared helper and therefore architecture-scope.

**`guardian` (low, process):** noted for the record that the change was already live under
the after-the-fact review ruling, and that *"a veto here can only produce a forward-only
follow-up, not a rollback"*. Stated in the submission for exactly that reason.

**Commits are credited by `Council-Submitted:` automatically at `098` report time; this
one carries `Council-Reviewed:` because the approved verdict has now been read.**
