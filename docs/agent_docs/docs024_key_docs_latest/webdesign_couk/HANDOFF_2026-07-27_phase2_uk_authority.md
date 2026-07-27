# HANDOFF — webdesign.co.uk Phase 2: the UK authority site

**Written 2026-07-27.** Cold-start for the next chat. Phase 1 (the merge) is done
and live — see `HANDOFF_RESUME_webdesign_couk.md` for how the site is built and
`SUMMARY_2026-07-27_corrections_and_handover.md` for where it stands. This
document is about where it goes next.

---

## What the owner asked for (2026-07-27, his framing)

> Leave the existing domains as they are. Rewrite the copy in webdesign.co.uk to
> make it clearer and better, and in order of popularity of the tools and guides.
> Focus on the **UK** and on helping solve **today's** problems for web designers
> and for people who *want* web design. Many helpful tools for images, colours,
> CSS — basically visual stuff. Pointers to the **best third-party tools in the
> UK**. A renewed focus on **AI in web design**. Be the **pre-eminent source** for
> the modern-day web design environment. Migrate gradually over the next month or
> so by **adding rather than removing**, and by improving what we have —
> usability, clarity, design, visually, up-to-dateness. Also a **news section**.

Source domains stay live and untouched. No redirects, no canonicals. **Decided.**

---

## First: the duplication question, answered with evidence

The owner asked whether `website-design.com` and `websitedesign.com` duplicate
each other. **Essentially no — they barely touch.** My earlier phrasing ("the
same content now sits on three domains") was loose and is corrected here.

**Written content: zero topical overlap.** The two sets are disjoint:

| | |
|---|---|
| **website-design.com** (23 articles) | engineering and maths deep dives — *The Physics of UI*, *The End of Hex Codes* (OKLCH), *Ambient Occlusion in CSS*, *Fractional Layouts*, *The Physics of Latency*, *The Bayesian Truth*, *Why You Can't Scrape Google*, *The Invisible Focus*, *The 44px Rule* |
| **websitedesign.com** (10 guides) | AI website builders, exclusively — *Understanding v0 by Vercel*, *Understanding Lovable*, *Understanding Bolt.new*, *Preventing "AI Slop"*, *The 70% Wall*, *The Realities of Browser Storage* |

Not one shared subject. Different positioning too: A is *"The Operating System
for Creatives & Engineers"*; B is *"Tools for thoughtful AI web design."*

**Files: no duplication.** `md5sum` across both trees returns **no byte-identical
files** — no shared CSS, JS, images or symlinks. All 63 tool slugs are unique.

**Five conceptual near-neighbours**, checked at code level and genuinely
different (recorded as PLAN D7): `seo-schema` (Article/Product/FAQ JSON-LD) vs
`seo-injector` (LocalBusiness, as an AI prompt); `text-sanitizer` (cleans after)
vs `insight-injector` (constrains before); `css-variables` (token file) vs
`vibe-equalizer` (mood sliders); `shadow-stacker` (manual layers) vs
`smooth-shadow` (parametric); five prompt tools with five distinct jobs.

**The only true duplicate found was *within* site B** — `hosting-economics.html`
was byte-identical to `local-ai.html`, wrong title and all. Already dropped.

**So the accurate statement is:** the two source sites do not duplicate each
other; `webdesign.co.uk` duplicates *both*, because it is their union. Since the
sources stay live, that is the duplication to be aware of for search — and the
owner has decided to accept it for now.

**Useful consequence for Phase 2:** the merge produced a genuinely
complementary library — deep visual/engineering craft from A, current AI-builder
practice from B. The "renewed focus on AI" is not a new direction; it is site
B's half, currently under-weighted by the index ordering.

---

## STATE AS OF 2026-07-27 — what is already done

The owner answered the open questions and two workstreams were actioned. **Start
from here, not from a blank page.**

### Analytics: wired, gated, awaiting ONE dashboard step (W1)

`SQL_p7_cloudflare_analytics.sql` — **applied live.** The Cloudflare Web
Analytics beacon is in the head chrome fork, **gated on a token**: with no token
the tag does not render at all, so nothing broken has shipped.

The token can only be minted in the Cloudflare dashboard — `CF_API_TOKEN` in this
repo is a GitHub Actions secret and is not reachable from the workstation. **The
owner has two routes and needs only one:**

- **Route A (recommended, zero further work):** Cloudflare dashboard → Web
  Analytics → add `webdesign.co.uk` → **Automatic Setup**. The zone is already
  proxied, so the edge injects the beacon itself. No deploy, nothing in this repo
  changes, and the gated tag below simply stays closed and harmless.
- **Route B (version-controlled):** paste the token from the dashboard's Manual
  Setup snippet into the commented `UPDATE` at the foot of `SQL_p7`, then
  re-render chrome.

**Until one of these happens, no data is being collected.** That is the single
highest-value five minutes available on this project.

### Popularity ordering: deferred, and for a better reason than "no data" (W2)

The owner confirmed **no Google Search Console** on either source domain, so
there is no proxy dataset to borrow. But he also made the decisive point:

> *"don't change the order until we have stats, but we will change the tools and
> guides now anyhow to make them all better so historical stats will be out of
> date."*

That is right and it settles the sequencing. **Do not add a `rank` field yet.**
Ordering measured against content that is about to be rewritten would be stale on
arrival. The order is:

1. instrument (above);
2. rewrite and improve the tools and guides;
3. *then* let stats accumulate against the improved content;
4. *then* order by popularity.

Reordering remains cheap whenever that moment comes — the indexes are generated
from the catalogue, so it is a data edit plus a re-run, not a rewrite.

### News: enabled, mid-flight (W5)

`SQL_p8_news_section.sql` — **applied live.** Scoped to this site only; it is
explicitly *not* `features_open/005` (a fleet programme to onboard ~37 pool
domains, parked for its own reasons). The verify block asserts no pool site was
touched.

What was created:

- **5 `news_search` sources** — UK web design; AI web design tools; CSS and
  browsers; web accessibility UK; web design trends. `news_search` was chosen
  over `rss` because a query states topic focus and degrades gracefully, whereas
  a curated feed list rots silently. **These queries are an editorial choice,
  listed in full in the SQL so they can be argued with, and untuned — nobody has
  seen what they return yet. Expect to revise after the first fetch.**
- **`/news/index.html`** — `page_type='news-index'`, sections `["news-listing"]`,
  `build_status='planned'`, and deliberately **`rebuild_policy='generic'`** unlike
  this site's other 97 owned pages, because the listing is machine-maintained and
  must stay free to re-render.
- **A `News` nav item** in the primary group.

**THE SEQUENCE FROM HERE MATTERS — do these in order:**

1. **Wait for the feed.** `content-feed-refresh` runs every 6h; sources were
   primed with `next_fetch_at = now()`, next task tick was due 13:49 UTC on
   2026-07-27. Check: `SELECT count(*) FROM content_items ci JOIN sites s ON
   s.id = ci.site_id WHERE s.domain='webdesign.co.uk';`
2. **Then build the page.** Only once items exist — a `news-listing` with nothing
   in it renders near-empty, and the assembler drops any section with ≤10
   characters of visible text.
3. **Then re-render chrome**, which publishes the News nav link.

**Do not re-render chrome before step 2 completes.** The nav row already exists
in the database but chrome has not been re-rendered, so no News link is live yet
(verified: 0 occurrences on the live home page). Publishing it early would put a
link to a 404 in the header of all 98 pages — `bugs_open/049`'s exact shape.

---

## The Phase 2 shape

The brief is five workstreams. Ordered by dependency, not by size.

### W1 — Instrumentation (do first, blocks W2)
Cloudflare Web Analytics beacon into the head fork. Optionally a
`/stats`-style internal endpoint later. **Nothing about "popularity" is real
until this has run for weeks.**

### W2 — Copy rewrite and reordering
- **Ordering**: `rank` in the catalogue → regenerate indexes. Cheap, repeatable.
- **Copy**: every tool's card label, subtitle and description currently comes
  from the *source sites' own index cards*, harvested verbatim. That copy is
  serviceable but was written for a different positioning and is uneven in
  voice. Rewriting it is a `catalogue_additions.json` edit, not a page edit.
- **Voice**: British English throughout (already the standing rail), plain and
  calm, no hype — the mission brief's register. There is a tested house style at
  `travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT.md` worth
  reading before rewriting outward-facing prose.
- **Hard rail, carried from Phase 1**: no invented statistics, ever. Counts are
  substituted from the catalogue via `{{TOOL_COUNT}}`; never type one. This
  project produced that error twice.

### W3 — UK focus and "today's problems"
Two distinct audiences in the brief that should not be blurred:
**web designers** (practitioners wanting a tool) and **people who want web
design** (buyers). They need different entry points and probably different
sections. Worth an explicit decision before writing copy.

UK angle candidates, all needing owner input on which matter:
accessibility duty under the Equality Act / WCAG 2.2, UK GDPR and cookie
consent, `.uk` domain and hosting choices, UK pricing expectations, Core Web
Vitals as a commercial rather than technical concern.

**Rail:** UK-specific *legal or regulatory* claims are exactly the class this
repo has been burned by (see the vetcomparison workstream's legal record). Any
such claim needs a citation to primary source, or it does not ship.

### W4 — Third-party UK tool directory
New content type: curated pointers to the best third-party tools, UK-relevant.
- **Decide the inclusion bar first** — this is a credibility surface, and an
  un-criteria'd directory becomes link-farm-shaped very quickly.
- **Check the platform first**: `model-directory` machinery already exists
  (`section_type='model-directory'`, a directory-researcher agent, freshness and
  export tasks). Read `model_directory_pipeline/` before building anything.
- **Commercial rail**: the site currently sells nothing and collects nothing,
  and says so on the about page. If any pointer becomes affiliate, that promise
  changes and the about copy must change with it. **Owner decision.**

### W5 — News section
**Do not build this.** It exists:
- Components live and active: `news-listing`, `latest-news`, `blog-listing`.
- A `content-feed-refresh` scheduled task runs every 6 hours.
- Docs: `006_news_feed_pipeline_v2.md`, and the `news_feed_pooling/` workstream.
- **Gate:** that workstream is *parked behind an explicit owner gate* — "do NOT
  onboard/arm without owner go", sequence defined in `features_open/005`. The
  owner's request here is that go-ahead, but route it through that workstream's
  own sequence rather than improvising a parallel one.
- Known trap from that lane: `bugs_open/015`/`081` concern news-index pages and
  mistyped pages with no repair path. Read them before arming a feed.

---

## Suggested sequencing over the month

| week | focus |
|---|---|
| 1 | **Owner: one Cloudflare dashboard step** (see above). Finish the news sequence. Decide the two-audience question (W3). Agree the directory inclusion bar (W4). Browser-QA the 16 tier-1 tools *or* widen Tier 4 to do it automatically. |
| 2 | **W2 copy rewrite — the main event.** Improve every tool and guide: clarity, usability, up-to-dateness. NO reordering yet (see above). Visual/usability pass on the indexes. |
| 3 | Tune the news queries against what they actually returned. W4 directory pilot with a small curated set. |
| 4 | UK-focus content. Review. **Ordering waits until stats have accumulated against the REWRITTEN content** — not before. |

**"Adding rather than removing" is already the architecture.** Every page is an
owned page; adding tools and articles is a catalogue entry plus a transform run.
Nothing needs to be taken down to add to it.

---

## What exists that you should use rather than rebuild

This project's most expensive mistake was filing a bug asserting a capability
did not exist when it had been running in production for weeks. Before building
**anything** in Phase 2, grep for it.

| you might want | it already exists as |
|---|---|
| browser verification that a tool works | Tier 4 `internal/adapters/browserrunner/` — real headless Chromium, live v1.0.1167. Gated to `component_level='tool'`; widening that predicate is `bugs_open/084` candidate 3 |
| a news feed | `news-listing` / `latest-news` components + `content-feed-refresh` + `news_feed_pooling/` |
| a curated directory | `model-directory` machinery + `model_directory_pipeline/` |
| reordering the site | `rank` in `catalogue_additions.json` → `harvest && transform` |
| adding a tool or article | one catalogue entry + the transform; indexes and search regenerate themselves |
| dead links / dead buttons | `dead_controls`, `phantom_internal_links`, `misdirected_cta` discovery checks, all live |

---

## Open questions for the owner

1. ~~Any existing analytics on the source domains?~~ **Answered 2026-07-27: no
   Search Console on either.** Hence instrument-then-rewrite-then-measure.
2. **Two audiences or one?** Practitioners and buyers want different things; the
   site currently addresses only the first.
3. **Directory inclusion bar**, and whether any third-party pointer may ever be
   affiliate — the site currently promises it sells nothing.
4. **Does "pre-eminent source" imply original journalism** in the news section,
   or curation of others' reporting? Very different cost and risk.

---

## Carried-forward landmines

- A ~98-page re-render takes **3.5 hours** and shows nothing for the first ~20
  minutes. Queue it and walk away. Do not diagnose it as broken (I did).
- **A palette pin governs colour values, not component selection.** Review the
  planner's *section* list, not just its page list.
- **No re-render path re-renders a section whose component changed** — render the
  template directly.
- **Counts are derived, never typed** (`{{TOOL_COUNT}}`).
- **Commit with an explicit pathspec**, never a bare directory — this is a shared
  tree and I swept another session's work by doing exactly that.
- The port refuses to build if any page loses JavaScript (`checkScriptParity`) —
  if it fails, that gate is right and the transform is wrong.
