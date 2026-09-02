# NOTES — designblog.co.uk lane (append-only, newest at the bottom)

## 2026-09-02 — lane opened on the owner's critique; every point verified

Owner critiqued the day-old site (verbatim text + full verification:
`CRITIQUE_2026-09-02_owner_site_review.md`, this directory). Before relaying
anything I fetched the six named pages from the public domain (all HTTP 200)
and checked each claim against the served bytes. **All 8 points confirmed** —
notably:

- Nav: 6 links, no Tools, while `/tools/smart-contrast/index.html` serves 200
  (92,206 bytes) — the tool is reachable only from body copy.
- Four listing pages carry **zero content items** as of 2026-09-02: glossary
  0 terms, inspiration 0 showcases, feed 0 entries, studios directory 0 studios
  (its only `<p>`s after the intro are the footer). Each instead carries prose
  describing its intended content — meta-`<h3>`s like "What gets included",
  "How the entries are written".
- Both AI-sounding sentences the owner quoted are verbatim in the served
  smart-contrast page.
- Exactly **1 `<img>` per page** on all 6 pages fetched.

Method + gotchas in `RUNBOOK_designblog_couk.md`. The four-empty-listing-pages
pattern matches the experience loop's detector class (listing-class live 08-31,
experience-promise live 09-02) — asked them whether those ran on the remakes.

[INFERRED] The "design is exactly the same" point is a mechanism property
(one composition library / chrome pattern across the fleet) — not re-measured
here; routed to the design threads as the owner directed.

### Routing log (messages sent 2026-09-02, this session → live sessions)

| To | Sent (delivered, msg_id prefix) | ACK |
|---|---|---|
| Portfolio positioning | 2026-09-02 `0e5b4c6e` | **ACK + measurements 2026-09-02** — filed `bugs_open/444` (digest below); loop closed same evening |
| components | 2026-09-02 `fb6573be` | **ACK + measurements 2026-09-02** (digest below) |
| experience loop | 2026-09-02 `3a74099e` | **ACK + measurements 2026-09-02** — taking the empty-index rule build (digest below) |
| theme kits | 2026-09-02 `07c378b9` | **ACK + measurements 2026-09-02** (digest below) |
| site design planner | 2026-09-02 `859bc81d` | **ACK 2026-09-02**, then full answer — filed `bugs_open/445` (digest below) |
| offer analyser benefit analyser visual designer [4628f9] | 2026-09-02 `c14808ca` (first send bounced on a name shared with an offline Remote Control session — resent with the ref) | **ACK + measurements 2026-09-02** (digest below; + contact-hero reconciliation, LANDMINE `0b3151337`) |
| copy quality two stage | 2026-09-02 `bd922b64` | **ACK + measurements 2026-09-02** — four-class decomposition (digest below) |
| editorial design uplift | 2026-09-02 `c7d84234` (supplementary — routed there by theme kits: imagery/graphic treatment is their remit, not theming's) | **ACK + measurements + a CORRECTION to my claim 2026-09-02** (digest below) |
| feed lane | 2026-09-02 `81cd60ce` (supplementary — 0 content_sources on the-design-feed; design-vertical source ask) | — |
| inline guide imager | 2026-09-02 `2ac36700` (supplementary — at editorial design uplift's request: the mig-644 vocabulary precedent, extend to illustration/infographic rows on the build path) | — |

**All 7 owner-named threads: receipt CONFIRMED with substantive answers,
2026-09-02.** (The 7 = portfolio positioning + the five design threads + copy
quality. Feed lane + inline guide imager are supplementary routes the answers
pointed at; their ACKs pending, chase if silent by 09-03.)

## 2026-09-02 (later) — three measured answers land; the picture converges on SELECTION

All figures below are the answering lane's, dated 2026-09-02, quoted with
attribution; each lane offered its queries for re-running. My independent
artefact check (this session): **all 6 designblog pages' single `<img>` is the
header logo** (`/assets/images/logo.png`, inside `<header>`) — content carries
zero images, corroborating the components census below.

**components thread:** sameness is a **selection problem, not a library
problem**. 156 shared section components active; **40 have zero live instances,
61 are single-site** — while ~10 components carry 75–87% of all slots on many
sites (advertise 81%, websitepromotion 78% of slots from the fleet top-10).
Fleet top: hero 37 sites/634 instances, call-to-action 36/528, article-body
32/311, Generic Text Block 30/249. Images: designblog stored markup **0 images
in 50 slots** (controls: garden-tools 4/43, boxingonline 5/48). Caveat they
flagged on the theme-kits fix: `defaultSectionsForPage` is only the FALLBACK
when the planner returns no sections — **someone should measure planner-vs-
fallback share of live pages before sizing that programme** (asked theme kits
to run it). Their lane can add: variant templates per component function,
richer per-item slot vocabulary (bugs_open/425 is that class).

**theme kits thread:** chrome is the sharpest instance — **36 of 37 sites
render `site-header` + `site-footer`**; 10 chrome-eligible functions exist
(header-minimal-tool, header-with-search, header-with-categories, …) **all
unused**, because `ChromeSlotFunction()` hardcodes slot→function and only
**6 of 40 style_collections pin `header_component_id`**. Layout: 9 of 18 in
use, 73% of sites on three. The hardcoded planner fallback emits
`hero, generic-text-block, call-to-action` identically everywhere; their
`page_archetypes` table replaces it — **committed but UNAPPLIED, no binary
carries the apply action; going live is an owner decision, pending**. ⚠ Kits
CANNOT fix colour: `render_css_from_spec` makes the 8 core slots spec-wins and
the owner ruled 09-02 the machine stays free to override any theme — **the
lever on served colour is the brief + design overlay**. Available NOW without
their mechanism: per-site `UPDATE` pinning an unused header on
`style_collections`; they also offer per-remake chrome+layout+structure
recommendations as data.

**vigilant designer thread:** `visual-design-auditor` is live but **pointed at
a different question** — coherence/correctness (5 dimensions: hardcoded hex,
spacing, type hierarchy, dark-section contracts, responsive), 0 mentions of
imagery/impact/distinctiveness in its config. Structurally cannot see this
critique: it never sees the rendered page (can't count `<img>`), audits one
site in isolation (cross-site sameness invisible by construction), and has no
representation of the brief's ambition. Their cross-site measurement: hero /
call-to-action / Generic Text Block / info-card-grid on **4 of 4** sites
checked (designblog, advertise, websitepromotion, webdesign.co.uk). **Two
already-ruled, already-built things are simply not switched on** (both theirs):
`Illustrated Text Block` loses to `Generic Text Block` on all but ONE site in
the estate (post-IMG-074), and `info-card-grid`'s carousel — owner-ruled
default-on — is on for **1 instance of 49**. Offers: a cheap pre-flight gate
(component-set overlap + per-page img count, plain SQL) they'd take on **if
the owner says so**; warns the larger half is the **imagery SUPPLY gap** —
placing machinery works, there are almost no pictures to place.

Cross-thread convergence, three independent measurements: **the library and
mechanisms for difference exist and are not selected** — hardcoded chrome
slot→function, hardcoded planner fallback, unused headers/layouts/components,
ruled-but-unflipped defaults. Nobody needs to author new components to end the
sameness.

Open measurement (asked of theme kits): what fraction of live pages got their
sections from the planner vs the fallback — decides whether the structure lever
is `page_archetypes` or the planner's prompt.

## 2026-09-02 (later still) — portfolio positioning, experience loop and site design planner answer; the class gets a bug number

**Portfolio positioning:** re-measured on advertise (channels-directory 0
entries, glossary same meta-prose, news 0 items) — **the empty-listing class is
now `bugs_open/444`** (committed `4d807e2fc`), the durable home; grep it before
filing anything overlapping. Per-type mechanism: FEED pages — news mechanism
live+proven elsewhere (idea.uk) but all four remakes have **0 `content_sources`
rows** and nothing in the build path creates one (undriven, not broken; route =
feed lane, design-vertical source for designblog). DIRECTORY — DIR-001 wired,
but "uk-design-studio" isn't a kind; adding one is data-only across SEVEN
places (two fail silently; checklist in the register's directory-pipeline.md +
LANDMINES) plus a researcher run. GLOSSARY/INSPIRATION — **no item producer
exists anywhere** (absence claim, marked for re-verification in 444); fix
candidate (3). Copy [INFERRED, their marker]: brief-echo copy downstream of
missing items — the door-closer is plan validation refusing/degrading a listing
page whose item source resolves to zero (444 candidate (1)). NAV: advertise
also lacks a tools link AND any /tools/index.html hub; ownership split between
site-planner (plan a tools hub) and the bugfix-149 nav-membership family; cheap
per-site fix = tools section-index page with `in_header` + nav rebuild. They
carry the fleet critique into the next 18 briefs' directions.

**Experience loop:** answer to "did the detectors run?" — **NO, and schedule is
not the real reason; SCOPE is.** SQ-004 (listing-class) last fired 07:25:11Z,
builds landed 11:47Z+; SQ-005 (experience-promise) has NEVER fired on schedule
(CronJob created 10:30:27Z, after its own slot; first run tomorrow 07:40Z). Run
BY HAND against designblog today: **Rule C scanned 0 of 0 index pages** — two
query filters (`articles`/`items` key must exist and be a non-empty array) drop
an empty index before the rule runs; on our four pages `arr_type` is NULL on
all 11 component rows. **"An index that lists NOTHING is invisible to it" — by
construction.** Second independent instance of their boxingonline CONTRIB gap ⇒
**they are building the empty-index rule and will re-run over our four pages.**
Also: /glossary.html is `page_type='content'`, outside Rule C's selector — the
fix needs a content-class trigger, not just a page_type widening. ⚠ Their
SQ-004 `--site designblog` run printed a FAILED positive control (control's own
case filtered out of the corpus — their defect, theirs to fix): **treat
SQ-004's designblog zero as UNTESTED, not clean.** Remit: they own checkable
promises; the "delight" gap needs a judgement seat (`brief-fidelity-auditor`,
filing_mode='record'), not a detector. FYI: **migration 694 applied today** —
`content-quality-auditor` now samples every page_type (was 4 hardcoded names,
7.7% of pages) and its new dimensions include "does an index page actually list
its own items, or write ABOUT itself instead", reads the brief,
filing_mode='record' (owner ruling today).

**Site design planner** (full answer; evidence in **`bugs_open/445`**): all
three remakes resolved to the **identical layout `magazine-grid`** — matcher
read line-by-line and working correctly; the cause is a **library gap**: of 18
layouts exactly ONE is professional-register editorial; there is **no "content
hub with embedded tools" archetype**, which is what all three briefs
structurally are, so three different briefs independently found the same
defensible answer. Fix = build that archetype (bounded design task, compounds
across the 18 queued remakes). Secondary: 9/18 layouts never chosen for any
live site; 3 cover 73% (independently matches theme kits' numbers). Ruled out:
classifier prompt-hygiene issue (real but non-scoring), matcher-ignores-signal
(shortlists genuinely differed). Nav: all three sites have **NULL header/footer
component links in style_collection → hardcoded default** — consistent with
theme kits' ChromeSlotFunction finding; components' territory.

**Convergent picture across five measured answers:** the sameness has three
independent, named mechanisms — (1) chrome hardcoded + unpinned
(ChromeSlotFunction, 6/40 pins), (2) layout library gap (one professional
editorial archetype, no content-hub-with-tools), (3) planner fallback trio +
top-10 component concentration — and the emptiness has one: **no item producer
for glossary/inspiration, no content_sources for feeds, missing directory
kind** (444). Detection missed it because both live detectors are blind to
empty indexes by construction (fix in flight, experience loop).

Cross-check flagged to theme kits + vigilant designer: contact-hero — theme
kits say ZERO component rows fleet-wide (18 live contact pages use
hero-contact); vigilant designer's cross-site table lists contact-hero on 3 of
4 sites. One is reading function names, the other stored rows; asked them to
reconcile before either number is quoted onward.

> **RESOLVED same evening (vigilant designer, measured):** both numbers were
> right — ONE component with `name='contact-hero'`, `function='hero-contact'`,
> 25 instances/25 sites. **76% of the 442 active components have
> `name <> function`** (6 are word-reversals), and `page_components.slot_name`
> is a THIRD spelling that agrees with neither reliably. Filed as a LANDMINE
> (`0b3151337`) with the resolving query: resolve through `component_id` to the
> row, then read both columns. A zero from the wrong column reads as "does not
> exist" and licenses rebuilding a thing that's on half the estate.

## 2026-09-02 (closing the round) — copy quality answers; all 7 threads ACKed

**Copy quality two stage** (measured at this build's own markers): **YES, the
build went through the gated path** — writer ran post-roll, gate scanned every
section, and repaired plenty on this site. The served examples decompose into
four KNOWN classes:
1. Feed-index "a starting point, not the final word": the gate SAW and
   TARGETED it; the repair model returned **`no_answer_for_target`** (child
   `c464393e` iter2) so the original shipped by design (guard never splices an
   unanswered target). Repair-model reliability miss, now a counted class;
   candidate fix = one re-ask for unanswered targets.
2. "says so plainly" (×3): banned WORDS currently have detection only (nightly
   checker v2, live since 09-01), no page-side repairer — repair arm queued as
   a design question.
3. "before your users have to" + a 450-char essay in a BUTTON field: the
   CQ-032 AI-tell + structural-mistype classes, already ruled and queued.
4. The four zero-content listing pages = instances 3–6 of the boxingonline
   "empty-room" class (page shaped like a promise, filled from the model when
   input is missing); their title-promise/empty-index gate design covers it;
   upstream items question stays with 444.
Bonus finding theirs: designblog's newborn brief carries **16 `x_not_y` in
content_direction + 8 in briefing** — proof the briefing agent mints the
register at BIRTH (turns the ruled fleet brief wash from stock into a
birth-producer gate question). advertise + websitepromotion: same path, expect
same four classes; nightly checker covers their briefs from tonight.

**Portfolio positioning closed their loop:** they've messaged theme kits
directly for the next-18-briefs directions (+ asked about retro-differentiating
the four live remakes — theirs + owner's call); no further action owed me.

**Round complete: 7 of 7 threads ACKed with substance** (+ editorial design
uplift and feed lane routed supplementarily, ACKs pending). Owner decisions
gathered for the report: page_archetypes apply/roll; pre-flight gate
assignment; cheap per-site fixes now vs class fixes only; who designs the
content-hub-with-tools layout archetype; priority on the two unblocked-and-
unstarted flips (Illustrated Text Block selection, carousel default).

## 2026-09-02 (late) — editorial design uplift answers WITH A CORRECTION to my own claim; imagery is a VOCABULARY gap

**CORRECTION (mine, caught by them, re-verified here):** "content imagery is
absent" was FALSE — heroes render as CSS `background-image` URLs and **all 6
pages serve a real hero** (`hero-home.jpg`/`hero.jpg`/`hero-glossary.jpg`
behind a darkening gradient; only 3 distinct files across 6 pages — the feed
and contrast tool reuse the homepage hero). `<img>` censuses miss every hero on
the estate by construction. Corrected visibly in the CRITIQUE addendum;
WRONG_CALLS row appended (`5d76472d8`): two agreeing measurements sharing an
encoding are one measurement.

**Editorial design uplift's measurements (2026-09-02):** designblog's planned
imagery = 10 rows, 10 active serving assets (5 page heroes, 4 section icons,
1 logo), **ZERO illustration rows, ZERO infographic rows**. Supply is COMPLETE
for what was requested; placement WORKS for what exists; the site is bare
because **the planner asked for chrome and never asked for a picture that
carries content**. Fleet-wide, everything planners have EVER requested: hero
399 rows/34 sites, icon 211/29, logo 50/33, illustration 25/5, **infographic
1/1** — ONE infographic ever planned on the estate. Not a designblog defect;
the estate's imagery VOCABULARY, and the remakes are just the newest sites to
show it. Second structural finding (recorded in `bugs_open/114`, three lanes
converged): across all 462 blog/guide pages fleet-wide the max PROSE sections
per article is **1** — one prose slab plus chrome everywhere, so there is no
article structure to place illustrations BETWEEN even if they existed.
⚠ Their hard-won caution (their own component-level fix this morning:
council-approved, applied, **rolled back 69 minutes later** — right for 9
pages, wrong for 292 that already had a hero reading the same key): **a remedy
is fitted to a POPULATION, not to a defect** — no component-level fix for
this. The actionable question (filed by them 08-31, unanswered): **what would
ever WRITE an illustration/infographic row?** Planner/vocabulary, upstream of
both lanes. Encouraging counter-example: migration 644 taught planner menus
the word for an illustrated section on 08-26 and 6 of the 9 pages carrying one
were composed AFTER that date — **the menu works when it has the word.**
⚠ They also decline to endorse the two "unflipped levers" until someone
measures what sites that already have them flipped actually render — present
those to the owner as unblocked/unstarted AND unverified-effect.

Routed onward at their request: **inline guide imager** (owns in-body imagery,
has the mig-644 counter-example) — message sent.

## 2026-09-02 (copy loop closed) — division of labour settled on the empty-room class

Copy quality two stage confirmed my representation of the four classes is
accurate as reported, and recorded a real design consequence (their commit
`1b5209a1a`, with attribution): their title-promise gate design now
**deliberately narrows to the NON-listing half** of the empty-room class (a
titled promise with no data behind it), because 444's candidate (1) — plan
validation refusing/degrading a listing page whose item source resolves to
zero — closes the listing half upstream of any writer fix. So the two
mechanisms are complements, not overlaps: **444(1) guards listing pages at
plan time; their gate guards titled promises everywhere else.** They carry the
birth-producer finding to the owner themselves, and will come back to this
lane for served-vs-stored test cases when the repair re-ask fix is sized —
standing offer holds.

## 2026-09-02 (imagery answer lands) — NOT a vocabulary gap: the planner is OBEYING its prompt

**Inline guide imager answered the open question, and it REFINES the
"vocabulary gap" framing recorded above** (and relayed to the owner — corrected
in README same evening): the live `build-site-planner` prompt **already carries
the full vocabulary** (`kind` enum documents `illustration` and `infographic`;
rule 15 repeats it). What suppresses them is three things in the same prompt,
verbatim quotes in their write-up:
1. An explicit default-to-zero instruction: *"Use sparingly in v1 — most plans
   will have zero section-scope entries."*
2. The stated MINIMUM is chrome only (rule 13: logo + heroes); no floor for
   illustration/infographic.
3. **The worked example's `sections` block contains ONLY icons; no infographic
   appears anywhere in the example** — and this estate's recorded trap is that
   a quoted exemplar ships verbatim ([[a-quoted-exemplar-in-a-prompt-is-copied-verbatim]];
   same mechanism as the copy lane's "demonstrations govern, instructions
   don't").
So hero 399 / icon 211 / logo 50 / illustration 25 / infographic 1 is **the
planner obeying** — nothing is failing, which is why vocabulary work cannot
move it. **644 does not transfer**: it fixed component SELECTION
(`component_expresses` had no image token), independent of the `imagery` block
that decides whether a picture is REQUESTED (evidence: 9 pages with an
illustration-capable section and no imagery row; vonc.com/about with a planned
illustration and no section that can show it — both in `bugs_open/114`).

⚠ **Two cautions attached to any prompt edit** (blast radius = every new
build, incl. the 18 remakes; cost = real generated images per section):
- Keep rule 16 ("each entry produces exactly ONE image") in the SAME edit, or
  section-scope volume produces multi-panel collage failures.
- **Article pages have nowhere to put images yet** (462 article pages, max 1
  prose section; 8 of 9 illustration-capable sections fleet-wide are on
  LANDING pages, 0 on blog/guide) — a prompt change lands pictures on landing
  pages only; in-article imagery ALSO needs composition (114). **The two asks
  must be separated when put to the owner.**

~~They are not editing the prompt (planner owners' seam + a cost decision);
neither is this lane.~~ **SUPERSEDED same evening — the owner RULED: "please go
ahead with the planner prompt and exemplar changes", and this lane executed it
(§ below).** Full write-up:
`docs/agent_docs/docs024_key_docs_latest/inline_guide_imagery/NOTES_inline_guide_imagery.md`
§15 (mechanisms: IMG-074 selection-side; IMG-075 per-section binding, live as
of today's 15:39 roll). Relayed to editorial design uplift (their 08-31
question, now answered).

## 2026-09-02 — this session's defect classes pushed to the fleet acceptance checklist (owner ask)

Owner: "add these site errors to SITE_DEFECT_CATEGORIES.md". The file already
existed (created earlier today from the boxingonline review — same purpose,
append-only, any thread may add). Added from this session:
**1.5** (AI-tell copy that went THROUGH the gate), **2.6** (live surface with
no nav path — the tools-nav class), **4.6** (chrome-only imagery plan; planner
obeying), **4.7** (one-slab articles), **§9** (cross-site sameness: 9.1
component overlap + three-names caveat, 9.2 identical chrome, 9.3 identical
layout/445, 9.4 hero reuse). Extended in place: **2.2** (designblog's four
empty pages + the 444 no-producer upstream check + the rule-C blindness note),
**4.2** (visible CORRECTION: the `<img>` check is half-blind to CSS-background
heroes — run both greps). New-site reviews should run
`docs/agent_docs/docs024_key_docs_latest/SITE_DEFECT_CATEGORIES.md` top to
bottom; this lane keeps its instances here and points there for the classes.

(Will be updated in place with send + ACK status; "the owner asked for receipt
to be checked" is the reason this table exists.)

Interpretation note: the owner named "the designer thread" — matched to the
live session **site design planner** (opened 09-02, owns the
`site-design-planner` composition mechanism, lane
`docs024_key_docs_latest/site_design_planner/`). "The vigilant designer thread"
= session "offer analyser benefit analyser visual designer" (lane
`docs024_key_docs_latest/vigilant_designer_offer_analysis/`). If the owner meant
a different thread by "designer", say so and it will be re-routed.

## 2026-09-02 (night) — MIGRATION 718 EXECUTED: the planner prompt now expects content imagery

Owner ruling (this session): "please go ahead with the planner prompt and
exemplar changes." Executed end-to-end tonight:

- **Read the live text first** (agent_definitions `f263eaa1`, `plan_site
  .config.prompt_template`, 28,117 chars; dumped whole per the 2KB-preview
  trap). All of inline guide imager's §15 quotes verified verbatim. Found a
  FOURTH suppressor they hadn't named: the Return-JSON skeleton demonstrates
  `"sections": {}` — an empty exemplar.
- **Five anchored replaces** (LANDMINES 17024: this row is co-edited by other
  lanes — anchored `replace()` with exact-count guards, never wholesale
  `jsonb_set`): (A) default-to-zero bullet → content-imagery-EXPECTED with two
  limits (display-capable sections only — the §4.1 undisplayable trap; only
  where there is something to depict); (B) rule 13 + index-page
  illustration-or-infographic floor + rule-3 exemption (114: article pages
  have no structure — in-article imagery deliberately NOT forced); (C) worked
  example + `index:1` illustration + `tools:1` infographic (text-free,
  headings-in-HTML); (D) example note extended; (E) skeleton demonstrates one
  entry. **Rule 16 untouched and verify-asserted.**
- **Proven before touching the DB**: local dry-run of all five splices +
  balance/length checks + ROLLBACK roundtrip byte-exact; then the full
  migration run against the live row with COMMIT→ROLLBACK (all guards passed,
  verify NOTICE fired).
- **Files**: `docs/agent_docs/sql_for_agents/718_planner_imagery_content_
  expected_prompt_and_exemplar.sql` + `_ROLLBACK` (commit `40dbeaea4`).
  ⚠ 715 and 716 each have TWO files (the collision landmine continues) — 718
  chosen from max 717.
- **Council**: corr `2dae4f20-3baf-46f9-96c5-fb35609ed7bd`, Council-Submitted
  trailer on the commit. **VERDICT OWED — read it and act on REVISE/REJECTED.**
- **Applied 19:59:56Z** (scoped `psql -f`, never the unscoped runner);
  `snapshot_agent` pre-update; recorded in `schema_migrations` with sha256.
  **Independently re-verified at the live row after apply**: 'Use sparingly'
  ×0, new bullet ×1, `infographic_selection_steps` ×1, rule 16 ×1, template
  28,117→30,164 (exact expected delta). Config = live immediately; no roll.
- **Canary**: portfolio positioning asked to check the FIRST post-718 plan's
  `imagery.sections` (they fire the next remake brief). Residual risks named
  in the council submission: instruction-only display-capability limit (§4.1
  base-rate), volume/cost scaling (owner accepted).
- Notified: bugs_open/427 (build-site-planner owners — heads-up with anchors),
  inline guide imager (their §15 = evidence base), portfolio positioning
  (canary), SITE_DEFECT_CATEGORIES §4.6 gained a REMEDY-LANDED note (the check
  inverts for post-718 plans: zero entries is now a defect signal).

### Same evening — gamedesign.uk correspondence (their site, same critique classes)

Owner reviewed gamedesign.uk (live 18:00Z) — "same problems as designblog".
Answered their four questions (landed-vs-planned + routing; hero seam; what a
clean improvement-loop pass proves — little, both detectors blind to the
empty-listing class; fleet-wide rulings). **Their "third state" (heroes
requested-produced-unbound) was SELF-CORRECTED within the hour: the heroes ARE
bound, as CSS backgrounds — the `<img>`-census trap's SECOND victim tonight**
(stood inline guide imager down). What survives, theirs (`bugs_open/446`):
/articles/index.html has NO imagery item at all (planner run requested no
site-scope hero — anomalous vs 8+ sites), its fallback hero.jpg 404s on their
domain, and `check_image_url_404.go` reads `<img src>` only — a CSS `url()`
404 is invisible to it. They CONTRIB'd 444 (fifth site, fourth mechanism: no
article pages exist at all) and are re-seeding their imagery guide + brief for
the vertical's temperature (their own root cause, stated plainly), re-planning
post-718.

### 444 fixing-session wiring facts for THIS site (recorded from their message)

- **/uk-studios-directory/ as built uses the BARE `directory-listing`
  component** (`query.business_directory`) which resolves via a
  `directory-json-exporter` config row (vetcomparison pattern) — NOT via a
  DIR-001 kind. **Decide which producer before populating** or the filled data
  and the component won't meet.
- **/the-design-feed/ was planned as `page_type='section-index'`** (not
  news-index) **with 0 child pages** — enabling a news source alone will NOT
  fill it; it needs a replan as news-index/news-listing OR child pages.
  → owner decision added to the gathered list.
- Their class mechanism (plan validation drops zero-source listing pages,
  files `capability_gap` with `spec.gap_kind='producer_missing'`) is in
  council, corr `c0990eb3`; lane dir `bugfix_444_empty_listing_pages/`.

## 2026-09-02 (night, closing) — council run started; two follow-ons

- **718 council run STARTED** (orchestration row for corr `2dae4f20`:
  `gate_tooling_provenance | EXECUTING_STEP` at ~20:2xZ). Verdict still owed.
- **444 session: my 718 apply caught a NUMBER COLLISION just in time** — their
  gate-enable migration was drafted as 718, now 720 (719 also taken tonight).
  The collision landmine keeps earning its line. Also from them: their
  resolver classifies "fed" by listing-component NAME only, so 718's imagery
  entries can never make an empty listing read as fed; and the feed fork
  resolves cleanly under either arm (news-index via content_sources;
  section-index via child pages).
- **inline guide imager, on gamedesign (relayed to that session): two of their
  three generated heroes are ORPHANED** — about + contact render the
  homepage's hero; cause is the component library (`hero-about`/`hero-contact`
  declare NO image field, so the correctly-resolved per-page asset has nowhere
  to land). The generalisable sentence: **"not 'the page has no image' but
  'the page has the WRONG image, so every check that asks whether it has one
  passes'"** — and their query caution: a component whose NAME contains the
  asset's name makes an unanchored `LIKE '%hero-about%'` census return the
  OPPOSITE of the truth (matched the CSS class); anchor on filename+extension
  with a known-referenced control. This is also 718's flagged residual one
  step later: entries must land where a component can display them —
  gamedesign is the live worked example.

## 2026-09-02 (final) — orphaned-heroes confirmed fleet-wide; a resolver gap caught pre-verdict

gamedesign re-measured inline guide imager's finding correctly (filename+ext,
control=3) — CONFIRMED, and the library shows it is a FLEET class: `hero`
declares `background_image` sourced `site_assets.hero` (38 sites);
`hero-about` (28 sites) and `hero-contact` (25 sites) declare NO image field
while their templates read `{{or .hero_url .background_image}}`. Fix (add the
field, same shape as `hero`) sent by them to the components thread directly.
Thanks relayed to inline guide imager.

Second: gamedesign's articles hub is `page_type='section-index'`, NOT
blog-index — the fresh build path types an editorial hub generically, so
444's blog-posts resolver arm (keyed on blog-index) would not hold it. Pinged
the 444 session pre-verdict (their gate migration is in council NOW — cheaper
to widen before round 2); offered /the-design-feed/ as the second instance of
the section-index shape for a test pair.

> **RESOLVED (444 session, same night): the section-index "gap" was NOT a
> gap** — section-index is a first-class arm of their resolver (zero children
> under the prefix → held, `builder_needed=section_children:<page>`), and
> /the-design-feed/'s exact shape is their test suite's held-case fixture
> (`TestResolveListing_SectionIndex_PlanChildren`, green). The blog-index-only
> keying was an inference gamedesign and I both made from their earlier
> message; corrected to gamedesign so their CONTRIB doesn't assert it. Kept
> nuance: `content-listing` resolves by SITE-WIDE blog-post count
> (`query.blog_posts` is unscoped), which is render-accurate.

## 2026-09-02 (analytics notice) — GTM key being applied; rerender wave expected

analytics_gtm session (397 §9 notice): designblog was born after the 08-26
backfill, has no analytics key, serves no tag. They are applying the standing
fix (owner-instructed standard for new builds): `site_config.analytics
.gtm_container_id` → one `stale_chrome` → chrome + all 17 pages re-render with
GTM-PQ3WCTBD at the next discovery pass. **Expected, not damage.** For this
lane: the next serve-side verification should expect the GTM head everywhere;
fresh render timestamps from that wave are CHROME, not content; the four empty
listing pages will re-render exactly as empty (fill is upstream, 444).

## 2026-09-02 (orphaned heroes: the census supersedes the hunch)

inline guide imager derived the class from the PREDICATE (template reads an
image key + schema names no `site_assets` source) instead of the reported
case: **SEVEN components, not three** — hero-about 28/43, hero-contact 25/25,
**hero-tool 23/76 (biggest, previously unnamed)**, hero-services 6/6,
hero-case-studies 4/5, teaser-reveal-panel 2/5, hero-use-cases 2/2. Damage
counted at the artefact: **157 live instances; 65 with their own
planned+active page hero; 61 orphaned** (render something else while passing
every has-an-image check). ⚠ **The blanket version is FALSE**: 4 of 65 DO
receive their own asset — leopardess's tool page renders its own hero with no
field, so another writer supplies it there; **identify that route before
editing seven components** (if it generalises it is the cheaper fix). All in
`bugs_open/114` second CONTRIB. Relayed to gamedesign (whose 3-component fix
proposal is superseded) and to components (asked: don't build the 3-component
version; read 114 first). The transferable lesson, third time tonight: a
population derived from the reported case is a GUESS; derive it from the
predicate, then SAMPLE the consequence — one of five samples disproved the
blanket claim.

## 2026-09-02 (handoff cut) — fresh chassis roll deployed; token expired; handoff written

Owner: fresh chassis build deployed. Two consequences for this lane: the 718
council run (last seen EXECUTING at ~20:25Z) may have been ROLL-KILLED —
checking it is the next session's FIRST action — and the kubeconfig token
expired at exactly this moment (3-day expiry, owner refreshes), so nothing
cluster-side could be verified at handoff time. 718 itself is DB config and
survives a roll. COUNCIL_SUBMISSION_718.json preserved in this directory for
a possible RESUBMIT_CORR resubmission. **COLD-START =
designblog_couk/HANDOFF_2026-09-02_continue_here.md** (supersedes the
CRITIQUE-doc-first reading order for a fresh session).

> **Feed lane ACKed just after the handoff was cut** — the last open route.
> designblog's zero-source /the-design-feed/ queued as THEIR priority 4
> (design-vertical source, explicitly not WebProNews), in
> `docs/agent_docs/docs024_key_docs_latest/news_feed_ingestion/HANDOFF_2026-09-02b_continue_here.md`
> §4. Routing table above now fully ACKed, every row. (They too are handing
> off on the fleet-wide kubectl expiry.)

## 2026-09-02 (post-refresh) — 718 verdict: APPROVED round 1; run was NOT roll-killed

Token refreshed; both first actions green. The council run COMPLETED
`complete_approved` at **20:06:19Z** — seven minutes after apply, BEFORE the
roll, so no kill. 718 re-confirmed in effect post-roll (new bullet present,
suppressor gone). Verdict: **APPROVED, round 1, 2 advisory objections, none
high-severity** (8 approve / 2 object / 7 abstained). Full report in
`diagnosis_artifacts kind='council_report'` corr `2dae4f20` (column is `body`,
not artifact_data).

The objections, and their state:
- **bug_historian (medium):** the display-capability limit is instruction-only
  — wants a MECHANICAL plan-application-time or discovery check that refuses/
  flags an imagery entry whose target section cannot express it (their named
  instrument: validate against `component_expresses`). Cites 039/044/114 as
  the recurring silent-drop class. **This is the real follow-up** — routed to
  inline guide imager (owns `component_expresses`/IMG-074/075 and the 114
  CONTRIB where the 61-orphan census already lives). The architecture seat's
  framing agrees: a monitoring item; if undisplayable imagery recurs
  post-718, the next move is a mechanical section-compatibility guard, not a
  redesign.
- **guardian (medium):** "re-verify anchors immediately before apply" —
  SATISFIED MECHANICALLY: the migration re-reads the row and re-asserts all
  five anchor counts at execution time inside the same transaction; the apply
  passed those guards at 19:59:56Z. **(low):** the two-active-rows
  agent_definitions landmine — the migration's precondition (`exactly 1
  active row` else RAISE) plus `ROW_COUNT=1` on the UPDATE proves
  build-site-planner is not one of the afflicted types. **(low):** fleet cost
  increase — flagged for the record; owner ruled.
- **prior_art_librarian (missing):** IMG-075 liveness is a deployed-binary
  claim their seat can't settle — a pod-grep would; noted, not owed here (the
  claim is inline guide imager's, made with capability probes on the 15:39
  roll).

Trailer state: commit `40dbeaea4` carries `Council-Submitted:` — 098 credits
it automatically now the correlation is approved; forward-only forbids an
amend, and none is needed.

## 2026-09-02 (late night) — the component fix landed FAST and repaired NOTHING; guard feasible; IMG-075 re-proven

inline guide imager, three findings (all relayed/recorded):

1. **URGENT, relayed to components:** six of the seven orphan-class components
   gained an asset-sourced image field at **20:15:47Z** — and the 61 orphaned
   pages are UNCHANGED; **9 re-rendered since the fix, 0 recovered**. The
   field is necessary-not-sufficient: only `image_landed` /
   `section_data_resolved` re-renders RE-RESOLVE; all other reasons redeploy
   stored HTML. Closing the class = re-resolving re-renders across the 61 +
   verification at the SERVED bytes (filename-anchored with control), never
   the component row (which now reads correct everywhere). Also flagged:
   confirm which seventh component was left out and why.
2. **The bug_historian guard IS feasible:** `component_expresses`' own
   predicate (`source LIKE 'site_assets.%' AND type IN ('url','image',
   'image_url')`) is exactly the discriminator. ⚠ **Validation caveat:** any
   test of the guard from now on measures the REPAIRED population — the 20:15
   fix changed what `component_expresses` returns for the class 75 min after
   the census; a clean result ≠ guard useless (they nearly published that
   refutation and caught the staleness themselves — a measurement went stale
   inside two hours). Sequencing agreed all round: canary base rate first;
   **IMG-077** (a section-scope arm on `check_unrendered_page_imagery`,
   offered by the bugfix_114 lane) is the cheaper instrument to consider
   before any new mechanism. One plan, in `bugs_open/114`.
3. **prior_art's IMG-075 liveness note DISCHARGED:** re-probed post-roll on
   `v1.0.1355`, both replicas, errors unsuppressed, negative control absent
   with visible exit code. ⚠ their first probe returned "absent" for all six
   INCLUDING the must-be-present control — the expired token behind
   `2>/dev/null` turned a failed exec into the word "absent" (the never-
   extract-keys/probe-from-pod family's oldest trap, still biting).

## 2026-09-02 (444 session, two proven facts about OUR index — commit 560a24c07)

1. **The `needs_section_data` row filed 21:04:44Z on designblog's index is
   their error-defer repair FIRING, not a fresh defect** ("Section
   'featured-content' … required query source errored: unknown query name
   \"featured_post\"" — binary-probed, the chassis carries the fix). Do NOT
   file it as new damage.
2. **Genuine finding for this lane:** the index plans a `featured-content`
   section whose component declares source **`query.featured_post` — a query
   base that DOES NOT EXIST in queryresolve's registry** (one of five
   unregistered bases they found; before tonight's fix it built silently
   hollow on every run). Options, per their message: (a) register a
   `featured_post` resolver (NEW query vocabulary — one handler +
   SourceDependency entry; fleet mechanism, not this lane's to mint
   unilaterally), (b) re-point the component's declared source to an existing
   base like `query.blog_posts` (**shared component — 9 pages / 8 sites as of
   2026-09-02**, components-thread territory), (c) drop the section
   (site-local). The HITL row is the designed decision surface.
   **This lane's read:** for designblog the section is downstream of 444's
   fill anyway (zero articles exist to feature), so the site-local half can
   wait for content; the CLASS half (register-vs-repoint for all 8 sites)
   is a components + queryresolve-owner call. **Added to the owner-decisions
   list** (handoff §5, item 7).
