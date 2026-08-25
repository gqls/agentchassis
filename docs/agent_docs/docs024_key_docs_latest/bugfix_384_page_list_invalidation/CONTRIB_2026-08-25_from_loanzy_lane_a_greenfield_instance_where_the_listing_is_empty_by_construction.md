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

> ## ⚠⚠ §5's HEADLINE MEASUREMENT IS RETRACTED BY ITS AUTHOR, ~40 MINUTES LATER. THE CONCLUSION STANDS ON OTHER EVIDENCE; THE LINK SET-DIFFERENCE DOES NOT.
>
> **What is withdrawn:** the "0 links on `/garden/index.html` that are not also on `/contact.html`"
> figure, and the inference *"an index page with zero page-specific links is indexing nothing"*.
>
> **Why:** `[MEASURED 2026-08-25 13:5xZ]` the shared 23-link menu contains **20 of the site's 21 page
> URLs** — every page except the one that 404s — plus 3 assets. **The nav links the entire site.** So
> every internal page link anywhere is already in the menu, and the set-difference is **pinned at zero
> for every page whether a listing rendered or not**. Had `/garden/index.html` rendered its
> `content-listing` with twelve month links, those twelve are all in the menu and the difference would
> still read 0. **The measure returns the same answer under both hypotheses**, which makes it a
> tautology rather than evidence.
>
> ⚠ **The tell was in the data and read as a strength:** *every* page scored 0, including a month page
> that is not an index at all, and including the negative control. **A measure that returns one value
> for an entire population including its control is not detecting a universal defect — it is failing
> to discriminate.**
>
> **What still supports §5's conclusion, all of it disconfirmable:**
> 1. `article-card` = **0** and `section--articles` = **0** on `/april/index.html` and
>    `/august/index.html`. A rendered `content-listing` emits both — this can come out otherwise, and
>    does on sites whose listings have data.
> 2. The live component row: `articles` = `{"source":"query.blog_posts","on_missing":"skip_section"}`.
> 3. The source is empty: the site's only `blog-post` page is `build_status='planned'` and 404s.
> 4. The three heading-vs-content failures, measured on **non-anchor** month counts, which do
>    discriminate (`/contact.html` 0, `/index.html` 12, `/garden/index.html` 3).
>
> **Where the set-difference WOULD be valid:** a site whose nav does *not* enumerate every page. Then
> a page adding nothing beyond the menu genuinely indexes nothing. **That precondition is one query
> and I did not run it before offering the measure.** Check it before reusing this:
> `SELECT count(*) FROM pages WHERE site_id=… ` against the distinct internal hrefs on any one page.

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


---

## 6. THE POSITIVE CONTROL — added 2026-08-25 after the `bugs_open/381` lane pointed out that my zero had none

A zero with no demonstration that the measure *can* be non-zero is an argument, not evidence. §5's
`article-card = 0` had no such demonstration. It does now, and the measure discriminates:

| page | raw `article-card` | after `<style>`/`<script>` strip | ELEMENTS bearing the class | `section--articles` |
|---|---|---|---|---|
| `homegarden.uk/april/index.html` | 46 | **0** | **0** | **0** |
| `homegarden.uk/august/index.html` | 46 | **0** | **0** | **0** |
| `dartsonline.com/` *(positive control — a listing WITH data)* | 108 | 108 | **12** | **1** |

**A listing with data scores 12 cards / 1 section; a skipped one scores 0 / 0.** That is the
disconfirming result §5 was missing.

> ⚠ **THE UNIT MATTERS AND MY FIRST VERSION OF THIS TABLE OVERSTATED IT BY 9×** (corrected
> 2026-08-25 by the `bugs_open/381` lane, verified here). I originally recorded **108**, having
> counted elements whose `class` attribute *contains the substring* `article-card`. Measured:
> the page carries **12 actual cards**, each emitting **nine BEM class names** —
> `article-card`, `__image`, `__category`, `__content`, `__title`, `__excerpt`, `__meta`,
> `__date`, `__read-time`, twelve of each — so a substring match counts the card container **plus
> its eight descendants**: 12 × 9 = 108.
> **Count the class as a whole TOKEN in the class list, not as a substring.** The inflation factor
> is the component's internal complexity, so it differs per component and looks like a real
> quantity: `108 vs 0` invites a reader to think the gap is nine times what it is, and nothing in
> the number itself says otherwise.

> ⚠ **AND THE CONTROL CARRIES A TRAP WORTH MORE THAN ITSELF: THE CONTAMINANT IS SITE-DEPENDENT.**
> On `homegarden.uk` the class name appears **46 times inside an inlined `<style>` block**, so a raw
> `grep -c 'article-card'` reports the cards as PRESENT on a page that has none. On `dartsonline.com`
> the stylesheet is external, so raw and stripped are **identical (108)** and the same raw command is
> perfectly correct.
>
> **So a markup-counting command verified on one site can be silently wrong on the next**, and
> "I checked this grep and it was fine" is worthless as assurance. **Strip `<style>` and `<script>`
> unconditionally**, and count elements bearing the class **as a whole token in the class list** — not string
> occurrences (inflated by every BEM descendant) and not substring matches on the attribute (same
> inflation). The token form is immune to CSS, to prose mentions, and to BEM.

**Provenance of this section:** the positive control and the CSS observation are the `bugs_open/381`
lane's; verified here independently before recording. Their framing of why it matters is the best
statement of it either lane produced: *"reproducing a number is not testing an instrument — the
question is never 'do I get the same figure' but 'what figure would I get if the thing were false'."*
