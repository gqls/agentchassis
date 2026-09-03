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
- **D4 (2026-09-02): positioning agreed with `Portfolio positioning [b9957b]`, brief
  DRAFTED, awaiting owner review before dispatch.** The steer, verbatim in substance:
  - gamesdesign.co.uk is the **authority** seat (free self-serve tools + guides, for solo
    devs, students, small teams). gamedesign.uk takes the **professional practice** seat:
    how working studios actually run game design — process, workflow, balance sign-off,
    documentation practice, pipelines, hiring and roles, tooling-stack reviews, opinion.
    Audience: leads, producers, professional designers. This is the estate's standing
    cross-TLD twin rule (P5, executed 2026-08-01: `.co.uk` = authority, `.uk` = instrument).
  - **Avoid the free-calculator and guide-library content kinds entirely**, not re-angle
    them; link the sibling's where a tool is relevant. Cross-linking is the halo,
    duplication the collision.
  - **Commercial slot: prepare, never claim.** gamesdesign.co.uk's live strategy spec
    records a paid-tier path; the literal name "GameDesign.uk Pro" appears in NO current
    spec, so the brief does not use it. No copy asserting a paid product exists, none
    foreclosing one. Owner rule 2026-09-02: no negative-identity copy by default.
  - Collisions to avoid: `designblog.co.uk` (general design editorial — its brief fired
    2026-09-02; stay strictly on GAME design), `cartoon.co.uk` (PROTECTED, owner ruling
    2026-08-20), `gamerooms.co.uk` (no "game room" phrasing), `writesy.uk` (narrative
    design as a game-design discipline is fine; not a general writing resource).
  - Register rows GD1 (sibling, documenting its existing seat) and GD2 (gamedesign.uk,
    status "proposed 2026-09-02 — direction fixes at the mission brief") written by that
    lane in the `positioning_register` DB.
  - Operational: `ensure_site_record` scans `name`+`network_id` without `COALESCE`
    (broke a sibling-flow release 2026-09-02). The pre-seed MUST set both:
    `network_id = '00000000-0000-0000-0000-000000000002'`.

  The brief: `MISSION_2026-09-02_gamedesign_uk.txt` in this directory — 2,484 chars, plain
  prose in the owner's voice, modelled on `noted_rebuild/MISSION_2026-08-11_noted.txt`.
  Deliberately contains no `"` or `\` so 082's single-line JSON folding is lossless
  (checked by running its exact `sed | tr | sed` pipeline over the file).

## 5. Phasing (rewritten 2026-09-03 — the 09-02 version ended "close the lane"; the owner's review reopened it)

1–5. ~~diagnose · root cause · file 432 · positioning · brief v1~~ DONE 09-02.
6–10. ~~seed · retract old tree · dispatch · verify build #1~~ DONE 09-02 18:00Z — **and build #1 was
   WRONG FOR THE VERTICAL** (owner review 20:30Z; `bugs_open/446`; SITE_DEFECT_CATEGORIES §10).
11. ~~improvement loop~~ RAN — 27 record-mode verdicts, none dispatched (446 §4a); tool plant HELD
    (447; SEED d); growth_posture=hold (SEED e).
12. ~~re-seed imagery/evidence/design_intent v2 + brief v2; re-dispatch~~ DONE 20:10Z 09-02 — but 082 on
    a deployed site is a STRATEGY REFRESH (RUNBOOK §7a); chain enqueued by hand 08:31Z 09-03 (SEED 09-03).
13. **IN FLIGHT:** `needs_briefing` 95d834f8 unclaimed → briefing → site_plan → reconcile → build.
14. Verify as a READER (HANDOFF_2026-09-03 §6): plan shape (444), heroes (721, owed to components),
    imagery prompts (10.1), hub lists articles (10.2), copy, held growth items (WDS-020).
15. Owner decisions: email · author · newsletter · born-hold · 447 ownership.
16. Close when the owner reads the site and does not have to say it again. 432 stays open (scheduling).
