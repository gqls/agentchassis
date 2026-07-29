# HANDOFF — webdesign.co.uk, continue here

**Written 2026-07-29 ~07:30 UTC.** Supersedes `HANDOFF_2026-07-27b_continue_here.md`
as the cold-start document. That file is still worth reading for the Phase 2 brief
and the D10–D12 rulings, but **every open question in it is now closed** and two of
its warnings were wrong — corrected in place there, and summarised below.

Companion reading, in this order: this file → `PLAN_2026-07-25` §D13 (the exposure
ruling) → `SUMMARY_2026-07-28_what_the_news_feed_taught_us.md` (the read-aloud
version) → `NOTES_webdesign_couk.md` tail (evidence and every misstep).

---

## TL;DR — state on one screen

| thing | state |
|---|---|
| **Cloudflare analytics** | ✅ **LIVE on all 99 pages.** Token in the head schema's `fallback`. |
| **Tool pages** | ✅ All 63 cross-linked to their guide + siblings. Were 100% dead ends. |
| **Home page** | ✅ **Zero broken links** (17 checked). Duplicate sections resolved. |
| **Favicon** | ✅ Live, derived from the site's own orphaned `logo.jpg`. |
| **Fleet 404 pages** | ✅ 20 sites. **But `bugs_open/132` is still OPEN** — they aren't served. |
| **News feed** | ⚠️ **Works, 25 on-topic items.** One query needs a 4th pass. Page NOT built. |
| **Buying design** | ⚠️ **UNBLOCKED** — D13 decided. Nothing written yet. |

---

## 1. DO THIS FIRST — read the news titles, do not trust the counts

The 01:51 tick runs ~twice a day. **Two queries changed and one is untested.**

```sql
SELECT cs.name, to_char(cfi.source_published_at,'MM-DD'), left(cfi.source_title,70)
  FROM content_feed_items cfi
  JOIN content_sources cs ON cs.id=cfi.source_id
  JOIN sites s ON s.id=cfi.site_id
 WHERE s.domain='webdesign.co.uk' ORDER BY cs.name, cfi.source_published_at DESC;
```

**`AI in design` is on its second query and has not been tested.** It refetches on
the next tick. First attempt (`AI generative design workflow designers`) returned
2 of 9 on topic — antibiotics design, energetic-materials design, 2D→3D CAD, a
lighting conference. Now `AI web design UX interface designers`.

> **The rule this thread built, and its refinement — this is the transferable bit.**
> Queries built from **sector-neutral vocabulary** return other sectors' content
> ("industry", "report", "acquisition", "merger", "builder", "tools"). Queries built
> from **domain nouns** return ours (`CSS`, `browser`, `WCAG`, `typeface`).
> **But a domain noun only discriminates if it is UNIQUE to the domain** — "design"
> is shared with every engineering field, which is exactly how the AI query failed.
> Four rounds of retuning produced that sentence; do not spend a fifth rediscovering it.

**Known flaw, recorded not fixed:** dedup keys on `source_url`, so **one story
covered by five outlets passes as five items** — 5 of the 9 design-industry items
are the same Coca-Cola rebrand. Story-level clustering is a platform change.

**Then build the news page.** `/news/index.html` exists as `build_status='planned'`.
~~The nav row will be **recreated automatically** once the page deploys — `refresh_nav_tables`
rebuilds `site_nav_items` from DEPLOYED pages, and it deleted the News row on 07-27.
There is no ordering trap left (see §5).~~

> **CORRECTED 2026-07-29 (session 4) — the automatic-nav claim was FALSE, twice over.**
> (1) Nothing rebuilds the nav on a deploy: `refresh_nav_tables` runs only inside a
> `nav-updater` run, which needs a `nav_drift` work item (discovery files them but
> they can sit at `detected`); session 4 filed one by hand. (2) When it DID run,
> the News row still did not appear: `classifyPagesForNav` skips child-prefixed
> URLs and `isSectionIndexType` omitted `news-index`, so a news page at the
> canonical `/news/index.html` could NEVER enter the nav — watched live in the
> nav-updater pod log 07:52:09Z. Filed + fixed as `bugs_open/141` (one word in
> `v3_site_actions.go` + tests, council corr `e0a52a70`). What caught it: waiting
> for the promised nav row and reading the pod log when it did not come. The page
> build itself DID work automatically once a `needs_page` item existed (SQL_p18;
> the missing link was that nothing watches `pages` for planned rows).

## 2. What is DONE and verified live

- **Analytics.** All 99 pages. The token lives in the **`input_schema` `fallback`**
  of the `webdesign-couk-head` component (`SQL_p13`) — **NOT** in
  `site_components.content_data`, which the chrome renderer never reads. `SQL_p7`'s
  design could not have worked; two verify blocks passed over it because both
  asserted the WRITE and neither exercised the READ.
- **63 tool pages cross-linked** (`SQL_p11` + `gqls/sites`). Copy lifted verbatim
  from the site's own index cards — nothing written, nothing invented.
- **Home page.** Two identical "What's here" sections → "Browse by category"
  (6 anchored cards) + "Start here" (specific tools). Six `id`s added to
  `tools/index.html`'s existing category headings — the categories already existed.
- **Favicon.** Derived from `assets/images/logo.jpg`, which the site already owned
  and referenced nowhere. Size-specific artwork: full mark at 32/48/64, prompt glyph
  at 16 (the full mark is illegible that small — verified by rendering).

## 3. The one rule that has caught this workstream FOUR times

> **A green status is not evidence. Verify the artefact, on the path a user takes.**

Every instance, so the next thread recognises the shape rather than re-learning it:

1. **Deploy green, nothing shipped** — `bugs_open/120`: on a merge commit `HEAD~1`
   is the first parent, so the deploy diff shows only the *other* side. **It is
   always the pusher who loses.** ⇒ `git pull --rebase` on `gqls/sites`, never `git pull`.
2. **Work item `complete`, chrome unchanged** — the beacon. `render_site_components`
   returns `"rendered": true` from an idempotence gate that did not write.
3. **File deployed and correct, change inert** — `bugs_open/132`: `404.html` exists
   and returns 200 by name, and a missing path never reaches it.
4. **`page_rerender` `complete`, `rendered_html` untouched** — assemble republishes
   **stored** HTML and does not re-render a section from `content_data`.
   ⇒ **Always write both fields** (`SQL_p10`, `p11`, `p17`).

## 4. Open bugs this site owns or filed

- **`bugs_open/132` (mine, unowned) — B2 sites serve a raw JSON error blob instead of
  a 404, leaking `objectKey`.** The 20 pages exist; **nothing routes to them.** Fix is
  a few lines in the Cloudflare edge worker, whose source is **in neither repo** —
  snippet is written into the bug. Candidate 4 ("generate a 404.html per site") is
  flagged there as a trap: it is done, and it fixed nothing.
- **`bugs_open/120` (mine, unowned)** — the merge-commit deploy skip, above.
- **`bugs_open/116` (mine, unowned)** — the three link checks have **never run on any
  site**. This is why the original 404 crisis went unnoticed, and why it would again.
- **`bugs_closed/127` (mine, fixed by the bugs-sweep thread within hours)** — news
  search was a plain web search; `search_type` was forced, logged, then dropped
  because the provider interface had no parameter for it. My production verification
  is recorded in that file.

## 5. Two warnings from the old handoff that were WRONG

Both corrected in place in `HANDOFF_2026-07-27b`, repeated here because they changed
what the next thread should do:

- **"Re-rendering chrome early puts a 404 in the header of all 98 pages" — GONE.**
  Superseded twice over: `applyNavVisibility` drops never-deployed targets, *and*
  `refresh_nav_tables` deletes and rebuilds `site_nav_items` from deployed pages,
  which is what actually fired. **I asserted the mechanism that did not run** — the
  conclusion held, the explanation was wrong, and the `[UNVERIFIED]` marker I had put
  on it is the only reason that was cheap to correct.
- **"American spellings on 23 of 98 pages" — the count was fair, the METHOD was
  unsafe.** The same letters are JavaScript identifiers (`optimizedSize` contains
  `optimized`), four live slugs carry the American form, and `meter` in *entropy
  meter* is correct British English. **Owner has since ruled: no rewrite for
  Americanisms alone.** Do not resurrect this from the old numbers.

## 6. Owner rulings that decide real work

- **D13 (2026-07-28) — exposure is settled: our OWN NAMED failures and fixes, at an
  asymmetric ratio.** *"Only claim our fixes once or twice but list the errors
  truthfully as many times as we like."* Full ruling + consequences in
  `PLAN_2026-07-25` §D13. Two things added there that are **not** the owner's words
  and need his eye: (a) a claimed fix needs evidence it **worked**, not that it
  exists — *"almost comparable to a human and sometimes more"* has **no measurement
  behind it**, so find one or soften it; (b) **RAIL: publish CLOSED failures only**,
  or ones with no exploitation value — `bugs_open/132` is an open information
  disclosure and must not be a worked example while it is open.
- **The designers STAY** (07-28), as the traffic engine. Not worth a rewrite for
  spelling alone, but genuine value-adds are welcome.
- **Empty feeds are acceptable if genuinely empty; fewer on-topic articles beat more
  off-topic.** This is why `web platform standards` was repurposed rather than padded.
- **Standing rail: never publish comparative rankings of named agencies.** This
  retired the original `UK web design` query, which was returning almost nothing else.

## 7. Next, in priority order

1. **Read the next feed tick's titles** (§1). If `AI in design` is still under ~3
   on-topic, **repurpose it rather than tune a fifth time** — the seam may be too
   thin for a 30-day window.
2. **Build `/news/index.html`**, then let the nav row reappear by itself.
3. **Write the buying-design section** — unblocked by D13. `PLAN_2026-07-27b_buying_design.md`
   is the live plan. Start with the accessibility-duty tool (a reframe of tools that
   already exist, not a build).
4. **Ordering by popularity** now has data arriving — but leave it until content is
   rewritten, per the owner's own sequencing: instrument → improve → measure → order.

## 8. Landmines specific to this site

- **`vertical_keywords` must move in LOCKSTEP with `content_sources.name`.** They
  exist to collide with the editorial source names so `seed_content_sources` no-ops.
  Rename one without the other and a **sixth** source appears carrying the bare
  keyword as its query. Every SQL file here asserts this.
- **A hard 30-day age ceiling** (`feed_actions.go:878`) — so `time_range: month` is
  the **only** window that is neither wasteful nor lossy. Widening it does nothing.
  And the age check is **guarded by `if publishedAt != nil`**, so an **undated item
  is never age-checked at all**.
- **Chrome is a stored artefact** (`bugs_open/117`) — publish it with a `nav_drift`
  work item → `nav-updater`, which queues the page re-renders itself. A page
  re-render will not rebuild it. This site is **clean on `bugs_open/118`**: all three
  slots have explicit `site_components` assignments, so the `ORDER BY name LIMIT 1`
  fallback never runs.
- **Repairs here are ARTEFACTS, not properties.** `content_data` + `rendered_html` +
  the published file all get overwritten by a full regeneration, because nothing
  upstream changed. True of the cross-links, the category cards and the anchors.
- **Reconcile any count before publishing it.** The category counts came to 62
  against 63 on disk on the first pass; the mismatch is the only reason the
  undercount was caught. This site has shipped invented figures twice before.

## 9. Files

`SQL_p7`→`p17` in this directory, each with its reasoning in the header. Most
recent: `p13` (beacon token), `p14` (query retune), `p15` (recency window),
`p16` (AI source + pool purge), `p17` (home page categories).
`gqls/sites` commits: cross-links, 404 pages ×20, category anchors, favicon.
