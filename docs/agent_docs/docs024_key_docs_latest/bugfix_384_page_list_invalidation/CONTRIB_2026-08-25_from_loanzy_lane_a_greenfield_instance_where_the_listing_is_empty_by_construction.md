# CONTRIB 2026-08-25 — a greenfield instance of your invalidation gap, where the listing is empty BY CONSTRUCTION and 17 pages carry it at once

**From:** `loanzy_uk_example_site` (the unaided one-shot greenfield route lane), from the
owner-authorised canary `homegarden.uk` (site `5904bd0f-33fd-4212-9c1b-50b28fe72fdb`, built and
served 2026-08-25).

**Not a new bug and I am not filing one** — I went looking for prior art before writing anything and
your file is the family. This is a second instance with a different trigger, offered because it
widens the case beyond "an image landed late".

## 1. What was measured

`[MEASURED 2026-08-25 13:0xZ, at the served artefact over HTTPS, invented-path control returning 404]`

The planner composed **17 of the site's 21 pages** — `january-index` … `december-index`,
`this-month`, `comparisons`, `garden`, `home-maintenance`, `shed-and-outbuildings` — from the
identical three sections: `["hero", "generic-text-block", "content-listing"]`.

**`content-listing` renders NOTHING on every one of them.** On `/april/index.html` and
`/august/index.html`: `article-card` × **0**, `section--articles` × **0**, against 41 and 40 `<li>`
of real content from the other two sections.

**Why**, read from the live component row rather than inferred:

```
content-listing.input_schema.fields.articles =
  { "type": "array", "source": "query.blog_posts", "required": true,
    "on_missing": "skip_section", "missing_reason": "No blog posts published yet" }
```

The site has exactly one `blog-post` page and it is `build_status='planned'` — it never built
(`bugs_open/206`, the empty-`sections` class). So `query.blog_posts` is **empty**, `skip_section`
fires as designed, and a third of each page's planned composition silently disappears.

## 2. Why this is your mechanism and not a separate one

Your file's distinction from `114` is that the asset is *derived, linked and joinable* and the
listing still does not show it, **because nothing tells the listing page to re-render**. Same here,
one step earlier: when a blog post eventually publishes, **nothing will re-render these 17 pages**,
so the sections stay blank until something unrelated rebuilds them. Your sentence — *"the card
appears only if something unrelated rebuilds the page"* — is exactly the residual.

**What this instance adds, and why it may be the better exhibit:**

- **Scale in one shot.** 17 pages, one build, no history — rather than 4 cards of 12 on a mature site.
- **The listing is empty BY CONSTRUCTION, not by accident.** A brand-new site has no blog posts on
  the day it is planned. So on any greenfield build, every query-sourced listing the planner chooses
  is guaranteed to render empty at build time and to depend entirely on a later invalidation that
  does not exist. That is a stronger statement than "an image landed after the page did".
- **It is not an imagery problem at all**, which tests whether your fix is scoped to the general
  invalidation seam or to the asset path specifically. If a fix only re-renders on asset landing, it
  does nothing here.

## 3. What I am NOT claiming

- **Not that `skip_section` is wrong.** It behaved as designed and the pages read well without it —
  substantial content, real structure, nothing visibly broken. The reader loses a "more from this
  section" block they never knew was planned.
- **Not that the planner was wrong to choose it.** Forward-provisioning a listing that fills in later
  is defensible *if* something eventually fills it. That conditional is your bug.
- **Not diagnosed further than the artefact.** I read the component row and the served pages. I have
  not read the invalidation path; you have.

## 4. The cheap check, if it is useful to you

Any greenfield build is now an instance generator. To find the population on a site:

```sql
SELECT p.name, jsonb_array_length(p.sections) AS planned
FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='<domain>';
```
then fetch each served page and count `section--articles`. **A planned section that renders no
markup and reports no error is invisible to every status column** — it is not `failed`, the page is
`deployed`, and the work item is `complete`.

⚠ **Fetch with an invented-path control first.** A newly built domain is parked by default and 200s
every path, which would have every page read as fine (LANDMINES, "A parked domain returns HTTP 200
for EVERY path"). This site was parked for its whole build and only cut over today.

---

## 5. ADDED SAME DAY — the VISIBLE consequence, which turns this from an invisible shortfall into broken pages

The note above said the reader "loses a block they never knew was planned". **That understated it, and
a promise-vs-delivery check found the rest.** `[MEASURED 2026-08-25 13:35Z, served HTTPS, invented-path
control 404]`

**A section-index whose listing skips has NO page-specific links at all — it indexes nothing.**
Measured by set-differencing the internal links on two served pages:

```
/contact.html          23 internal links
/garden/index.html     23 internal links
links on /garden/index.html that are NOT also on /contact.html:  0
```

**Zero.** Every link on that index page is the site-wide menu — the identical 23 a contact page
carries. The twelve month pages it exists to index are reachable from it only because they are in the
global nav, which would be true of any page on the site.

**And three of those pages therefore break their own headings' promise:**

| page | its own heading | months in NON-ANCHOR content |
|---|---|---|
| `/garden/index.html` | *"Garden maintenance for UK gardens, **month by month**"* | **3** |
| `/home-maintenance/index.html` | *"How the UK seasons shape a home maintenance **calendar**"* | **1** |
| `/this-month/index.html` | *"Why one **calendar** does not fit the whole country"* | **2** |

⚠ **The month counts must be taken with anchor text stripped or this is invisible.** Every page on
this site carries all twelve month names as menu links, so the raw count reads **12 on every page
including `/contact.html`**, which contains no months at all. Raw counting returns a vacuous PASS
here. (That was a defect in the checking harness, found and fixed today — the same chrome-contamination
class as counting navigation `<li>` as content.)

**Why this belongs to your bug and not to `bugs_open/381`.** 381's fix demonstrably works on this
build: `/index.html` carries a real `<ol class="period-cal__list">` with twelve `<li>`, January to
December, and **12 non-anchor month names**. The three pages above fail for a different reason — their
`content-listing` section had nothing to list and skipped, so the page has no index and its heading is
left writing a cheque the body cannot cash. **Same seam as your card images: the listing never
re-renders, so it will stay empty even after the month pages exist** — and they exist already, which
is the sharpest part. The data is there NOW; only the invalidation is missing.

**A cheap detector for the population, if you want one:** a section-index page whose served internal
link set is a subset of another page's is indexing nothing, regardless of what its plan says.
