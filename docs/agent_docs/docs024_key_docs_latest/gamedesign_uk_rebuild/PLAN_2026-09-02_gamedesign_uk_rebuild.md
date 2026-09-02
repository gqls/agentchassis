# PLAN — gamedesign.uk rebuild

**Opened 2026-09-02** by the `gamedesign.uk` session. Owner asked: "look up previous
threads for gamedesign.uk and fix the site, it is in a bad way."

## 1. What the site is, in plain terms

`gamedesign.uk` is a domain we own that is **still serving pages to the public while the
platform has no record it exists**. There is no row for it in `sites`. Its `pages` rows
were deleted at some point. The files it was last built from are still sitting in the
`portfolio-sites` bucket answering requests, and they have been since **2026-08-02**.

Most of those pages are empty shells: a header, a footer, and literally nothing in
between.

## 2. Measured state [MEASURED 2026-09-02, first-hand]

Controls first, because the finding is a 200 and a 200 proves nothing on a parked domain:
an invented URL (`/this-path-does-not-exist-9z8x7.html`) returns **404**, so this domain
is not a catch-all and its 200s are real pages.

| path | code | body content outside header/footer |
|---|---|---|
| `/` and `/index.html` | 200 | **0 chars** — `<main>\n\n</main>` |
| `/about.html` | 200 | **0 chars** |
| `/getting-started.html` | 200 | **0 chars** |
| `/services.html` | 200 | **0 chars** |
| `/tools.html` | 200 | **0 chars** — this is the flagship "Utility Engine" page |
| `/contact.html` | 200 | 276 chars |
| `/games.html` | 200 | 374 chars |
| `/guides.html` | 200 | 1093 chars |
| `/privacy.html` | **404** | linked from every page footer |
| `/terms.html` | **404** | linked from every page footer |
| `/sitemap.xml` | **404** | |
| `/robots.txt` | 200 | present |

**Six of nine pages are empty.** Verified not to be client-side injection: the only
`<script>` on the page is a 320-char mobile-menu toggle, and a fetch with a Chrome
user-agent returns the same empty `<main>`.

Other live defects: the footer ships `href="mailto:"` with no address on every page; the
footer nav repeats items ("Tools About Get Started Guides Games Get Started Games
Services Our Services Tools Guides Games Get Started Games Guides Contact").

### The CSS vocabulary split

The page carries its own inline `<style>` block that references **7 custom properties
that `/assets/css/styles.css` never defines, none with a fallback** — so every one of
those declarations is dropped by the parser:

`--border-radius`, `--shadow`, `--color-heading`, `--color-hero-title`,
`--color-hero-subtitle`, `--color-secondary-text`, `--color-secondary-hover`

`styles.css` supplies a different vocabulary for the same jobs (`--radius`,
`--shadow-sm/md/lg`). Two design systems glued together. Cosmetic while `<main>` is
empty; it would bite the moment content returns, so it is recorded here rather than
fixed in place.

The palette also carries the `bugs_open/113` shape — a dark theme
(`--color-background: #121212`, `--color-text: #e0e0e0`) holding light-theme literals
(`--color-card-bg: #ffffff`, `--color-header-bg: #ffffff`).

> **CORRECTION, same session, before it reached anyone:** I first read that white-card
> literal as live damage — light-grey body text on a white card is 1.32:1, unreadable.
> It is **not** live: no page on the site uses `class="card"`, so the rule never
> instantiates. The check that caught it was extracting the classes actually present in
> the markup instead of reasoning from the stylesheet. **A CSS rule is not damage until
> the markup instantiates it.**

## 3. Why nothing caught it

`scripts/audit-archived-still-serving.sh` (from `bugs_closed/359`) exists precisely to
find retired-but-still-serving pages. It **cannot see this site**: it enumerates
`pages.status='archived'` with a non-null `deployed_at`, and gamedesign.uk has no `pages`
rows at all. 359's damage class is a page retired in the DB; this is a **whole site whose
DB rows were deleted while its artefacts kept serving** — the same gap one level up, and
outside 359's detector by construction. See NOTES for whether this earns its own bug.

## 4. Decisions

- **D1 (owner, 2026-09-02): rebuild gamedesign.uk through the framework.** Chosen over
  redirecting to gamesdesign.co.uk, over making it primary and retiring the sibling, and
  over taking it down. The duplicate-content cost of a second game-design domain was
  stated in the option and accepted.
- **D2 (owner, 2026-09-02): it must go in a DIFFERENT direction to gamesdesign.co.uk**,
  and the positioning is to be agreed with the `Portfolio positioning` thread. This
  resolves D1's duplication problem: the two domains are not to hold the same product.
- **D3 (this session): FRESH build, not ADOPT.** `082_submit_domain_unified.sh` with no
  `--from`. Adopting from gamedesign.uk itself would ingest the empty shells; adopting
  from gamesdesign.co.uk would reproduce the sibling, which D2 forbids. FRESH creates the
  site row via `domain-submitter` and enters the cascade at `needs_domain_research`.
- **D4 (pending): the mission brief.** Blocked on the positioning answer — the brief is
  what fixes the site's whole direction at the classifier, and it is expensive to undo
  once pages are written. Asked `Portfolio positioning [b9957b]` 2026-09-02.

## 5. Phasing

1. ~~Diagnose the live state with controls~~ **DONE 2026-09-02.**
2. Agree differentiated positioning with `Portfolio positioning`. **IN FLIGHT.**
3. Seed the site row + `evidence_base` + `imagery_style_guide` before submitting, per the
   `oufe` worked example (`docs024_key_docs_latest/oufe/SEED_2026-07-25_oufe_site_and_specs.sql`)
   — the email matters (`bugs_open/063`: the hallucinated-email check fails OPEN with no
   contact email), and `evidence_base` must exist before the first page is written or the
   claims layer silently no-ops.
4. Dispatch the FRESH build with the agreed mission brief.
5. Verify at the artefact, not at the status: `<main>` non-empty on every page, the two
   legal pages resolving, a sitemap, and no empty `mailto:`.
6. Decide what happens to the orphaned artefacts currently serving — they must not
   survive the rebuild as stale siblings.
