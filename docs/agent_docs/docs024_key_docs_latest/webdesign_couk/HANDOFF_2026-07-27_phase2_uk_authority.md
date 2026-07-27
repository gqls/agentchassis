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

## THE ONE THING TO DECIDE BEFORE ORDERING BY POPULARITY

**We have no popularity data. None.** This is the load-bearing premise in the
brief and it must not be quietly guessed at.

- The site went live on **2026-07-26**. It is one day old with no inbound links.
- It is served from **B2 behind Cloudflare** — there is **no nginx access log**.
  (The `traffic_probe` numbers used elsewhere in this repo come from
  `/var/log/nginx/access.log` on a *VM*; webdesign.co.uk has no VM.)
- No analytics beacon is installed. No `analytics`/`pageview` table exists in the
  platform DB.
- The source sites are also static on B2, so they have no logs to borrow either.

**Recommended sequence — instrument first, order later:**

1. **Add Cloudflare Web Analytics now** (free, cookieless, one `<script>` in the
   head chrome fork — `SQL_p5_chrome_forks.sql`, component
   `webdesign-couk-head`). It gives per-path pageviews. This is a ~10-minute job
   and every week it is delayed is a week of data not collected. **Do this first,
   whatever else is decided.**
2. **Meanwhile, order by a declared editorial proxy, not by an invented
   metric.** Write the ordering into `catalogue_additions.json` as an explicit
   `rank` field with a stated basis (e.g. "editorial, by breadth of use") and a
   **named reversal trigger**: *revisit once 30 days of analytics exist.*
   Anything else risks a fabricated ordering presented as a popularity ordering,
   which is the same class of error as the invented tool count this project
   already made twice.
3. **Then reorder from real data.** See the payoff below.

**The payoff, and it is large:** the tools and learn indexes are **generated from
the catalogue**, not hand-authored. Reordering the whole site is a *data edit
plus a re-run* — add `rank` to the entries, `webdesignport harvest && transform`,
import, re-render two pages. No code change, no page rewriting. The architecture
already supports exactly what the owner wants.

**Ask the owner** whether he has any existing analytics (Plausible, GA, Search
Console) on the two source domains — Search Console in particular would give real
query and click data for the *same content* under different URLs, which is the
best available proxy and costs nothing to check.

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
| 1 | W1 instrumentation. Ask owner re Search Console. Decide the two-audience question (W3). Agree the directory inclusion bar (W4). Browser-QA the 16 tier-1 tools *or* widen Tier 4 to do it automatically. |
| 2 | W2 copy rewrite + editorial ordering (declared as editorial, with the reversal trigger). Visual/usability pass on the indexes. |
| 3 | W5 news section via the existing pipeline, through the owner gate. W4 directory pilot with a small curated set. |
| 4 | Re-rank from the first real analytics if enough has accumulated. UK-focus content. Review. |

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

1. **Any existing analytics on the source domains?** Search Console especially —
   it would give real query/click data for this exact content.
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
