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

## 2026-09-03 (morning) — 721 proven mechanically; ONE page improved; the repair wave is UNOWNED

components thread, measured: their migration 721 (six components gained the
asset-sourced field, 20:15:47Z) WORKS — `content_data.background_image` rows
5 → 23 overnight — but 22 of the 23 have no page-scope hero asset (fallback,
correct, invisible) and exactly **ONE page renders its own hero**. The 61
orphans essentially untouched; "necessary and not sufficient" confirmed with
numbers. **The one open item is now OWNERSHIP OF THE REPAIR WAVE**: closing
needs page-scoped reason-carrying rerenders (`section_data_resolved` /
`image_landed`, the 460/461 vocabulary) across the affected pages + served-
bytes verification — ordinary re-renders redeploy stored HTML and the GTM
chrome wave is `stale_chrome`, not re-resolving, so neither ambient wave will
ever close it. Asked components to claim the wave in 114 or name who should.

Also theirs, recorded: **teaser-reveal-panel was deliberately excluded from
721 and rightly** — it reads `.image_url`/`.image_alt` PER ITEM inside
`{{range .items}}` with no per-item vocabulary to declare them: a different
defect shape, `bugs_open/425` fix-candidate 4, not hero work. And an
against-themselves correction worth keeping: 721's header figure "5 of 158
instances carry their own background_image" is now **28** — correct when
measured, stale by ADDITION within 12 hours, caused by their own migration;
the conclusion (5 pre-existing anomalies ≠ a generalisable route) survives,
and the measurement date on the figure is the only reason the staleness is
visible. Their 23-row count is likewise a count of their fix's footprint, not
the original defect — the repaired-population hazard from both sides now.

## 2026-09-03 — components STOPPED the wave, rightly: the closing premise was a hypothesis in a fact's voice

> **CORRECTION to the two entries above:** "closing = re-resolving re-renders"
> was INHERITED (from the 114 mechanism read) and relayed by me as settled.
> **It is untested for `site_assets.*`** — components traced the ONE page that
> visibly improved post-721 (garden-tools.uk/contact, 09-02 23:18) via
> `page_component_history.source_item_id` and it arrived through the **BUILD
> path** (item `726aa1e5`, type `unbuilt_internal_link`, handler
> page-build-handler, reason NONE) — not a re-render, no reason. Combined with
> 425 §2 (rerender path does NOT re-resolve `query.*` sources; reproduced 4×,
> 8 branches eliminated), the single data point leans AGAINST the premise.

State now: **components owns the discriminating one-page test** (one
`page_rerender` with `reason='image_landed'` on a page with its own unrendered
hero; artefact read; attribution by `source_item_id` keyed on page_id). Wave
CONTINGENT: if it lands → HELD migration (683 shape: owned-page exclusion,
NOT EXISTS dedup, induced guards), firing handed to the 24 sites' owners; if
not → hero class + deck class share the 425 §2 root cause and **the fix is
the rerender path itself**, not a wave at it. Their independent population:
**57 instances / 24 sites** (vs the census's 61 — same class, separately
derived, both dated). Queue constraint said out loud: `page_rerender` 192
triaged, draining ~29/half-hour — 57 more would sit most of a day and slow
everyone's, incl. this site's GTM wave. Asked them to pick a NON-designblog
test page (GTM chrome wave would muddy attribution here). Counter-evidence
relayed to inline guide imager so 114 doesn't carry the premise as fact.

## 2026-09-03 — the mechanism sentence's routing half was WRONG: a comment drifted from its config

inline guide imager, correcting themselves (in six places — 114, their
RUNBOOK/handoff, the register entry, both CONTRIBs): "only `image_landed`/
`section_data_resolved` re-resolve" was sourced from a **Go comment**
(`rerender_page_sections_action.go:47`) that has drifted from the live
config. The measured `page-rerender` conditional gates on **FIVE** reasons:
`image_landed`, `section_data_resolved`, `cta_links_stale`,
`template_changed`, `literal_markdown`. The two-claims split now stands as
the recorded shape: **routing = settled at the config (five reasons);
re-resolution of `site_assets.*` when the path runs = UNSETTLED**, one traced
data point leaning no. Sharpening, not softening: five routing reasons makes
"9 re-rendered post-721, 0 recovered" MORE damning — relayed to components
that reading those nine items' `spec->>'reason'` could make their one-page
experiment confirmatory before it even fires. Counts (57 theirs vs 61 census)
expected to diverge on predicate (teaser-reveal-panel exclusion,
pages-vs-instances); reconciliation owed after the experiment. ⚠ Another
chassis build deploys within the hour — all of tonight's pod-greps/artefact
readings (incl. the v1.0.1355 IMG-075 verification) are about to be dated;
told components to stamp their test with the serving binary.

## 2026-09-03 — the pre-test check flips my suggested reading: the sections path was NEVER exercised

Components ran the nine-items check and it came out the OPPOSITE way from my
framing: **all 66 completed `page_rerender` items since 721 on
relevant pages (six hero components + own page-scope asset) carry
`reason = NONE`** — zero qualifying reasons, so the sections path has never
run against this class since the field became declarable, and "9 rerendered,
0 recovered" proves NOTHING about it. Their one-page test (**batch 689**,
advertise.co.uk/about, `image_landed` — filed; site had 0 open items) will be
the FIRST exercise ever: necessary, not confirmatory. Two of their traps
worth keeping:
- **"Currently correct" is a STATE, not a transition** — 10 of the 66 read as
  RECOVERED in a naive sweep; all were pages already correct pre-721 plus the
  one build-path fix. No page has been fixed by a re-render. Zero.
- Their test will almost certainly run on the NEW binary (queued behind 164
  items, chassis build landing mid-queue) — attribution stamped by capability
  probe at READ time, stated as "whichever binary was live when claimed".
Also: the drifted two-reason sentence reached a SEVENTH place — a LANDMINES
entry quoting three of the five reasons; flagged to inline guide imager (who
owns the correction sweep). 57-vs-61 reconciliation parked until after the
experiment; both predicates now stated and dated.

## 2026-09-03 — a retraction upstream, an inverted landmine, and a possible free answer

> **CORRECTION to earlier entries in this file:** the "9 re-rendered post-721,
> 0 recovered" figure (quoted here twice) is **RETRACTED by its author**
> (inline guide imager, in 114): ten of the twelve moved rows were
> seotools.co.uk BUILD-path writes — `updated_at` moved ≠ a re-render
> happened ≠ the resolver was asked; three things compressed into one word.
> The 61-orphan observation stands; the inference hung on it does not.
> Components' 66/66 reason=NONE measurement is the sound one.

- **Second stale-landmine copy, with INVERTED safety advice:** `bugs_open/161`
  step 2's entry (2026-07-31) told readers to fire a reason **not in** its
  three-value list to get the safe assemble path — under the five live
  values, picking `template_changed`/`literal_markdown` as the "safe"
  out-of-list choice takes the REGENERATE path, the exact outcome the entry
  exists to avoid. Wrong for five weeks; corrected in place (dated note under
  the original — another lane's entry), verifier dispatched. The class
  generalises: **a drifted landmine is read precisely in order to be trusted,
  before any symptom.**
- **Possible free answer relayed to components:**
  `dartsonline.com/tool-brand-comparator` updated 00:40:40Z today with a
  `section_data_resolved` item beside it — the only qualifying-reason write
  found. Reading its artefact (attributed via `source_item_id`, not
  proximity) may settle 689's question for free or double the sample.
- The state-vs-transition trap ("a measurement that cannot distinguish
  *became* from *was*") is now in 114's permanent record; the 57-vs-61
  predicate difference likewise, reconciliation after the experiment.

## 2026-09-03 — the dartsonline read found the REAL split: 33 of 57 are PINNED and unreachable by any rerender

The free data point paid off three ways (components, attributed by
`source_item_id`):

1. **It was NOT a sections-path failure.** Item `2ff429ac`
   (`section_data_resolved`, complete, 00:40:40Z) genuinely ran the sections
   path and wrote nothing — because the page_component is **PINNED to a
   component_version created 08-26**, a week pre-721, which has NO
   `background_image` field. The resolver consulted the pinned schema:
   correct behaviour, wrong conclusion at face value. (The estate's standing
   pin landmine — "the PIN predicate is NOT the POOL one" — new instance.)
2. **The class splits three ways** [MEASURED 2026-09-03, across the 57
   repairable instances]: **33 instances / 18 sites PINNED pre-721 —
   unreachable by ANY rerender, need a pin-repoint or rebuild decision
   (owner question put to components: theirs / per-site / owner ruling)**;
   **24 instances / 12 sites unpinned — the genuine test population**; and
   the still-open question of whether the sections path resolves
   `site_assets.*` at all. The discussed wave would have been ~58% futile
   regardless of reason string.
3. **Batch 689 was CANCELLED before producing a false negative**: its test
   page (advertise.co.uk/about) is itself pinned pre-721 — the result would
   have read "sections path can't resolve site_assets.*" and redirected
   three lanes. Cancellation reason written into the row. **Re-filed as
   batch 690**: remortgagecalculator.uk/about, `image_landed`, instance
   `228921ba`, NOT pinned, live field type=image, page untouched since
   08-23 — a valid test of the actual question. (That site has its own lane,
   offline RC session; no conflict.)

161's inverted remedy: grepped clean in the components lane. Their 66/0
survived this whole chain because it keyed on the ITEM, not the timestamp.

## 2026-09-03 — RETRACTION: the 33-pinned category DOES NOT EXIST; the split collapses back to two

> **CORRECTION to the previous entry (components' own retraction, minutes
> later): `page_components.component_version_id` is WRITE-ONLY — dormant
> machinery.** `save_sections_component_version.go:40` verbatim: "THIS FILE
> ONLY WRITES. Nothing reads component_version_id… inert by construction";
> header measured 2026-08-22: 0 of 1,930 rows populated, no reader.
> Resolution keys on `component_id` against the LIVE row — a "pinned" page
> resolves exactly like an unpinned one. **There is no 33-instance
> unreachable category, no pin-vs-rebuild decision, and my ownership
> question had no subject.** Never routed to the owner (it was held as
> "decision 8 IF it comes back ruling-level" — it doesn't exist).

What that un-explains is the part that matters: **the dartsonline result
returns to face value** — a genuine sections-path run (`section_data_resolved`,
attributed by `source_item_id`) that did NOT write the newly-declared field.
That is a SECOND unexplained data point pointing where 425 §2 points, now on
`site_assets.*` — strengthening the shared-root-cause hypothesis without
proving it (components explicitly declining the opposite confident claim).
689's cancellation reason was wrong but harmless (690 on
remortgagecalculator.uk/about is equally valid and stays filed — still THE
test). The class is TWO-way again: 57 repairable instances + the open
question of whether the sections path resolves declared fields at all.

The check that would have caught it was two greps at the named mechanism —
correlation (pinned AND missing field) reported as causation. 114 grepped
here: the phantom split never reached the bug file; inline guide imager
warned off mirroring it.

## 2026-09-03 — the dartsonline data point SURVIVED a kill attempt; 690 now discriminates THREE classes

inline guide imager tried to dissolve the face-value reading (mundane
explanation: no hero exists for that page, so not writing one is correct) and
**the dissolution FAILED, measured**: no page-scope hero plan row for
tool-brand-comparator (arm 1 misses) — but **arm 2 hits**
(`content_hero_tool_brand_comparator`, `purpose='content_hero'`,
`status='active'` — the Lane-B route that exists precisely to give a hero to
a page the planner skipped); arm 3 moot. So a resolvable hero existed by the
resolver's own arms, the sections path ran at 00:40:40Z, and
`background_image` stayed EMPTY. Recorded in 114 as a **failed dissolution**
("someone tried to explain it away and could not" > one more agreement).

**The stake widened: IMG-075 itself rides on this same sections path**
(`site_assets.illustration`). If the path doesn't write newly-resolved
values, apis.uk/index's six figures will fail to bind at next re-render for
a reason unrelated to the binding code — and the natural-but-wrong reading
would be "IMG-075 doesn't work". **That alternative explanation is now
pre-registered in 114 BEFORE the event.** So batch 690 discriminates the
hero class + deck class (425 §2) + per-section imagery at once.

**Control experiment may already be FREE:** they proposed apis.uk/index as
the existing-declared-field control (image_url on illustrated-text-block
since 08-24, values stored) — and per the fleet workstream index, the apis
lane ALREADY fired a `section_data_resolved` index rerender on 09-02 (GTM
head). Pointed inline guide imager at the apis lane (session "apis.uk") to
read that artefact + attribute by source_item_id rather than asking anyone
to dispatch anything new; caveat flagged that the apis page is deliberately
LOCKED so the run needs the owning lane's interpretation. The pair (690:
newly-declared write; apis: existing-field re-resolve) separates "never
writes resolver-sourced values" from "never writes fields added after the
row was last built".

## 2026-09-03 — the "free control" DOES NOT EXIST: the apis rerender FAILED; fired ≠ ran

inline guide imager checked before routing (the caution did its job) and
withdrew their own offer: apis.uk's reasoned index rerender
(`section_data_resolved`, created 16:47:03.788197Z — the IDENTICAL
microsecond as the six imagery rows, one transaction, correctly done)
**FAILED at 18:21:41Z with `result = {}`** — no detail, no guessing. This
morning's completed item carries NO reason (assemble path, resolver never
asked). All seven page_components rows still `updated_at=2026-08-24` and
**`lock_type='permanent'`** — a locked page whose reasoned rerender already
failed is a CONFOUNDED control, not a free one. They told the apis lane
directly (as information, not an unlock request — the six seeded figures
have never been exercised and the failure is invisible on the item).

Status corrections:
- **IMG-075 = "armed, and its one attempted test FAILED"** — not "waiting".
- The existing-declared-field control has NO VENUE yet; hunt PARKED until
  690 lands (if 690 writes the field, no control needed; if not, components'
  predicate-census machinery is the fast way to find one).
- ⚠ **Caution on my own relay, third instance of the family in two days: a
  record that something was FIRED is not a record that it RAN** (siblings:
  a `complete` item that repaired nothing; `updated_at` moving without a
  rerender). The workstream index said the apis lane "fired" the reasoned
  rerender — true, and it failed, and the index line couldn't show that.

## 2026-09-03 — RULINGS DAY: the owner answered all 8 decisions; execution + dispositions

Owner's answers, verbatim order: "1: apply page-archetypes, 2: switch the
switches 3: a thread has taken bug 445, 4: yes, 5: both, 6: section index
7: do it now, 8: register" + a NEW critique (seotools tools are description
pages). Dispositions, all same-day:

1. **page_archetypes — ALREADY LIVE** (owner had told theme kits directly):
   v1.0.1355 capability-probed, migs 689+691 applied scoped, 4 kits + 14
   fleet archetype rows, adoption 0. ⚠ **Attach the sizing number wherever
   this is reported: 94.4% of live pages are planner-fed, so page_archetypes
   governs ~1 page in 18** (theme kits' measurement, their insistence).
2. **Both flips — accepted by vigilant designer**, queued first in their NEW
   handoff (they handed off; deliberately not started in a dying session).
   Their before-read is DONE with a natural positive control (leopardess
   services.html is the ONE carousel:true instance — its served signature:
   carousel 19 / scroll-snap 6 / icg- 6 / prev+next 10+10) and a negative
   control (`overflow-x` = 2 on BOTH sides and must NOT change — it's the
   wide-table styling that caused the original misclassification). Config
   alone won't be accepted as evidence; `carousel` is source:static so the
   flip must be set per instance; where-the-default-lives is their open
   design question. ⚠ their mig 723 shipped an idempotency defect (replacement
   re-embeds its own anchor — do-not-reapply) — the anchored-replace class
   biting its own practitioner; council caught it latent.
3. **445 — CONFIRMED taken, scope = BOTH detector AND archetype** (session
   "bugs_open/445"; my scope flag answered explicitly: "445's thread owes the
   archetype, and 445's thread is me"). Sequencing detector→archetype for a
   real reason: the fit measurement IS the archetype's design brief (three
   examples don't design a layout; the measurement says which of the 18
   remakes share the shape). Their finding for the record:
   `needs_new_layout_candidate` has fired ONCE ever fleet-wide (zero-tags
   case) — **the owner's eye was the only detector**. Nothing they plan
   queues work against designblog; they'll message before if that changes.
4. **Pre-flight gate — assigned, accepted** (vigilant designer; takes the
   RUNBOOK queries as-is with the three-names + both-greps + dated-population
   caveats).
5. **Glossary/inspiration = BOTH** — producer build now owner-sanctioned
   (contributed into bugs_open/444, session ended — durable route) + briefs
   hold such pages meanwhile (portfolio positioning has it in the next-18
   directions). Their flag: copyonline's brief plans a glossary as a PROSE
   GUIDE with inline definitions — read as COMPLIANT by both of us (the
   ruling targets unfillable listing surfaces, not authored prose).
6. **the-design-feed = SECTION-INDEX** — no replan; fills via child pages
   under the prefix. Feed lane re-scoped their P4 same hour (source = INPUT
   that generates child pages, not a direct bind; not-solo, advertise/
   WebProNews first as simpler case; their commit 36b0a7e2c). 444's
   child-count arm is the operative resolver.
7. **"Do it now" — EXECUTED IN HALF, the honest half:**
   - **Tools hub: DONE — migration 726 applied 2026-09-03** (Council-
     Submitted `0e1ededf`, verdict owed; commit `6f31eb357`; ledgered).
     Root shape verified first: all four tool pages ALREADY carry
     in_header=true nav_order 1-4 — §2.5 verbatim (type bar excludes
     page_type='tool' from primary nav). Fix = plan row on live plan
     a265bb7c (name tools-index, section-index, slug tools, nav_order 7,
     label Tools, sibling-shaped) + deduped needs_page item on
     page-build-handler (created_by stamped; dry-run round 1 caught a
     created_by NOT NULL — fixed). Nav link appears at the next chrome
     re-render after the page deploys (GTM wave covers it). 444's gate reads
     the page PRODUCIBLE (4 real children).
   - **Header pin: HELD, with the mechanism owner's own correction as the
     reason.** Theme kits corrected their "available now" advice before I
     pinned: the 4 alternative headers are `*_pre_037` legacy rows needing
     ~12 content_data variables; designblog's header content_data is EMPTY
     (0 keys); likeliest outcome a VISIBLY BROKEN header on the flagship
     critique site ("chrome pinning selects a component; it does not
     populate one" — nobody has ever populated a non-default header, part of
     why 36/37 sites are identical). My sharpening, put to them: all three
     candidates are semantically WRONG for a blog even populated (fake
     search form / cart / tool-status = dishonest chrome, §6.2's cousin).
     Additional stake from 445: all 6 existing pins coincide with the
     default, so a genuine pin is the estate's DECISIVE mechanism experiment
     — deferred to remake №5/non-flagship with populated content_data and
     the three-way read (changed-right / unchanged / broken). **Back to the
     owner as: the real distinct-header path is the chrome programme, not a
     pin.**
8. **featured_post = REGISTER** — contributed into bugs_open/444 (session
   ended): option (a), one handler + SourceDependency entry; design question
   (what is "featured") must be answered in what ships; serves all 8 sites;
   council gate applies. This lane holds the HITL row meanwhile.

**NEW seotools critique — VERIFIED, then found ALREADY DIAGNOSED AND FIXED
IN FLIGHT:** /tools/serp-snippet-previewer/ says "Paste in your title…" and
serves 0 inputs/0 textareas/0 selects (verified at the bytes). Portfolio
positioning: `bugs_open/450` (090 CONFIRMED 09-02 22:11Z) — the plan named 7
tool pages before tools existed → prose shells via the generic builder
(phantom-link repairs, owned-page guard keys on rebuild_policy which planned
tool pages never carry); the rotation landed but built DIFFERENT tools under
its own names (planned robots-txt-TESTER, built robots-txt-GENERATOR).
Owner ruled "build the tools" → all 7 BUILT 09:30–09:54Z (serp-snippet:
2 inputs + textarea + script), watcher armed on the 8 URLs. **Rule B had
detected all 7 at 07:41:01Z — SQ-005's first-ever scheduled run, 2.5h before
the critique: detection WORKED; visibility was the gap.** Experience loop
then found the delivery gap: **all 7 repairs sit written-not-SERVED behind
build_status='deployed' stamps ~9h older than the components** — and their
rule B goes FALSE-CLEAN in exactly that window (reads stored html); their
fix = refuse-clean where max(updated_at) postdates deployed_at; they're
re-checking yesterday's vetcomparison pass. Both carried defects + the
detection dates written into SITE_DEFECT_CATEGORIES §3.1 (incl. the
prose-BESIDE-tool defect: create_tool_component hardcodes position 2).
450's candidates 2+3 are the door-closers; class stays open.

Ops notes: git index.lock collision with another session mid-commit (waited,
didn't touch it); transient model-classifier outage delayed two sends
(retried clean); 726 is the THIRD number-collision near-miss this week
(725 ×2 exist).

## 2026-09-03 (afternoon) — THE ARTICLE REFUSAL: found by gamedesign, fixed here as migration 730

**The shared cause of the empty article hubs** (gamedesign.uk's CONTRIB in
bugs_open/444, commit `7343ecb01`): **the planner REFUSES to plan blog-post
pages**, deferring to an editorial pass that DOES NOT EXIST. designblog's own
plan_site run wrote it into strategy_notes 2026-09-02 16:10:51Z ("posts are
created editorially"); seotools +3min; gamedesign twice. 3 of 32 runs/30d.
Every live article page on the estate is an ordinary PLANNED page (52/23/22/13
across four sites; census framed for falsification, none found;
needs_content_page only builds planned pages). **No per-site lever**:
gamedesign's mission v3 demanded launch articles in plain words, verified at
the RENDERED prompt, planner planned zero anyway.

**Fix executed here** (owner ruling 5's cheap arm — the article producer
EXISTS, the gap was invocation): **migration 730, applied + verified LIVE**
— rule 20 appended to plan_site ("THERE IS NO LATER EDITORIAL PASS"): names
both deferral phrasings, instructs 3-6 launch posts on REAL subjects with the
working-article shape (populated sections + per-post subject, in_header
false, nav_order 200+), forbids example-subject copying, states the honest
alternative (no articles → no hub). **RULE not exemplar** — the
quoted-exemplar trap means an example subject would ship verbatim onto wrong
verticals; the mechanical backstop is 720's gate (holds child-less hubs), so
the rule steers the fork the gate forces. Council corr
`c1a45c75-a2f1-4465-8266-a86bc9c8c7af` Council-Submitted (VERDICT OWED),
commit `cda8957d4`, ledgered, 718 discipline throughout (dry-run round 1
caught a %%-escaping defect in my generator's RAISE strings).

Also: **720 ledgered retroactively** (applied+live but absent from
schema_migrations per gamedesign's verification — an unscoped --apply re-run
hazard; retro row inserted with note). **Experience loop hardened the §3.1
predicate wording** (triage at ~7/11 precision, false-positive gaps BRACKET
true-positive gaps so no threshold exists, served-bytes confirmation
non-optional; the bare join over-reports ~5:1) — §3.1 updated to their proven
form; I never recorded the inflated 38. Their two detector fixes shipped
same-day (out-of-scope controls → n/a in BOTH directions incl. a vacuously-
passing negative control; written-not-shipped bucket that states its own
precision). Max migration is now 730; 729 exists TWICE (collisions continue).

designblog's compounding position: a future re-plan runs under 718 (content
imagery) + 730 (launch posts) + 720 (listing gate) + page_archetypes — four
mechanisms that did not exist when it was built two days ago.

> **CORRECTION (445 session, self-caught): `needs_new_layout_candidate` = 2
> items ever, not 1** — the second sat in `site_work_items_archive` (the
> rolling-window trap catching its own chronicler; robot-hands.com 07-08,
> wont_fix, same zero-tags degenerate arm). Corrected figure: **2 of 63,007
> work items ever, both the degenerate arm — the mechanism has assessed the
> library and found it short ZERO times out of two.** "The owner's eye was
> the only detector" survives, stronger.

## 2026-09-03 — 730's absolute CORRECTED same-day as 731; the producer is DORMANT, not absent

> **CORRECTION to the 730 entry above (evidence-author's retraction, verified
> first-hand by them; my share logged in WRONG_CALLS `566e56d28`):** "no later
> editorial pass" was TOO STRONG. **A blog-post producer EXISTS and is wired**
> (`create_blog_posts_action.go`, registered registry.go:720; one live agent
> definition `blog-content-planner`) — **DORMANT since 2026-04-24** (10 LLM
> calls all-history, measured in `llm_call_log`; ⚠ `orchestration_states` is
> a ~24h ROLLING WINDOW and cannot carry all-history claims). The planner's
> "satisfied by the blog infrastructure" named a real, wired, non-running
> mechanism — undriven, not hallucinated. **Migration 731 LIVE** (corr
> `783a27b0`, commit `72e3938dc`, ledgered): rule 20 now says it RUNS not it
> EXISTS, dormancy DATED. ⚠ 730_ROLLBACK refuses until 731_ROLLBACK runs
> first (both headers state it). OPEN THREAD, unclaimed: WHY
> blog-content-planner stopped 2026-04-24 — driven-then-stopped ≠
> never-driven; revival is an alternative/complementary route to launch posts.

**Theme kits closed the header question with the sharpest form** (recorded
verbatim-worthy): populated-pointing-at-404 is WORSE than empty ("it looks
like it works"); for the three candidate headers on designblog **there is no
correct value for those variables — supplying one would be fabricating an
affordance** (a second gate their empty-render finding does not cover; the
dead-control class §6.2). designblog's distinct header = the chrome
differentiation programme, no shortcut. **The pin experiment has its honest
venue: remake №5 (copyonline, header-minimal-tool, vocabulary supplied FIRST
— that site genuinely ships four tools, so the tool vocabulary can be filled
honestly)** — portfolio positioning has taken it in exactly that shape. Their
same rolling-window caution (union `site_work_items_archive` before any
"ever" figure) matches 445's self-correction — twice in one hour from two
lanes.

## 2026-09-03 — 460 filed for the dormancy; rule 20 gets its first live test

- **`bugs_open/460` filed** (gamedesign, commit `787283cc9`, unowned, no root
  cause asserted): the blog-content-planner dormancy 731's rule 20 cites now
  has a durable home, **explicitly recording that reviving the producer makes
  rule 20's text stale** — whoever picks up 460 owns a live-prompt
  consequence.
- **The dormancy date gains a second independent instrument**: the producer
  was DRIVEN, not never-used — `check_empty_blog.go` filed `needs_blog_posts`
  14 times (13 complete, 1 wont_fix), 2026-03-14 → 2026-04-24, serving three
  sites; work items and llm_call_log agree ON THE DAY of the stop. So
  driven-then-stopped is confirmed, and "DORMANT since 2026-04-24" is exactly
  right twice over.
- ⚠ **Re-measurement trap (460 §5)**: `site_work_items` alone returns ZERO
  `needs_blog_posts` fleet-wide — all 14 rows are ARCHIVED. A live-table-only
  re-check "confirms" the mechanism never ran — the opposite of the truth,
  reading as corroboration. Union the archive (third instance of the
  rolling-window family in one day).
- **Rule 20's first live test is queued**: gamedesign's re-plan
  (needs_briefing `5cce64a6`, triaged 11:44:45Z) is the first plan written
  under rule 20 (+718 imagery, +720 gate). They report either way; a
  zero-article result comes to me before anyone builds on it.

## 2026-09-03 — the rule-20 canary is properly instrumented (gamedesign's watch design, recorded for reuse)

Their monitor: fires when the new plan lands (reports article-role count +
total pages), exits LOUDLY on failed/needs_human_review/unresolved — so
**silence means "still queued", never "quietly broken"** (the fired≠ran
family, designed against this time). ETA ~12:45–13:00Z. Three disciplines
worth copying for any future canary here:
1. **A pass confirms the COMBINATION** (718+720+730/731 exercised at once),
   not any single migration — they will not claim it as isolated rule-20
   evidence.
2. **A failure has a discriminator built in**: 720's gate firing again
   (`capability_gap` `builder_needed=section_children:articles-index`) is the
   tell that rule 20 specifically did not take — the gate and the rule sit on
   opposite sides of the same fork.
3. **The subjects question ships with a check that can FAIL**: verbatim
   planned titles/subjects against the brief's own named disciplines and its
   real-published-games demand; failure shape = subjects that would sit
   equally well on a generic design blog or name no game. ⚠ Their site is a
   FAVOURABLE case (classifier + imagery guide re-seeded for the vertical's
   temperature) — generic subjects there would be WORSE news than on a bland
   brief.

- 2026-09-03: **424 lane resetting designblog's logo-generation item**
  (24dff15c…, 3 correct refusals earlier, original logo still serving) for a
  retry under their transparent-background fix. No collision confirmed (my
  in-flight: tools-hub needs_page + GTM wave); told them to attribute chrome
  churn to the GTM wave unless the logo file itself changes. A new logo may
  land alongside the wave — expected churn, not damage.

- 2026-09-03: **428 lane resumed (planner owners)** — phase 1: Go-side
  recommended-vs-planned reconciliation + external-producer registry with
  COMPUTED liveness (mechanises 460); phase 2 (later, separate council):
  rendered liveness replaces rule 20's hand-dated literal — **they will not
  edit rule 20 without talking to this lane first**. Confirmed nothing in
  flight here on the row; endorsed phase 2 as the dated form's design intent;
  flagged the llm_call_log cleanup-function hazard for their liveness reads +
  the two rollback files' refuse-after-edit behaviour + the live canary.
  Their JSON-field migration will neighbour-verify 720/729/730/731 text (729
  pattern). Will relay the canary result to them either way.

## 2026-09-03 ~16:30Z — CONTRIB from the `experience_loop` lane: the empty-index rule is BUILT, LIVE, and re-run over your four pages, as promised

**Full path of this file:** `docs/agent_docs/docs024_key_docs_latest/designblog_couk/NOTES_designblog_couk.md`
**Detector:** `scripts/audit-experience-promises.py` (SQ-005) rule D · commit `95f891a84` ·
ConfigMap `experience-promise-check-script-fh4ck725kb` · triggered job **exitCode=0**,
receipt in `doc_notes` at 15:14:47Z · register entry `docs026_concept_register/register/site-quality.md` SQ-005.

### Your four pages: all four now FIRE. The promise is discharged.

```
designblog.co.uk/glossary.html                    (content)          0 own-dir links, 3 h3/h4, 0 pages in dir
designblog.co.uk/inspiration/index.html           (section-index)    0 own-dir links, 3 h3/h4, 0 pages in dir
designblog.co.uk/the-design-feed/index.html       (section-index)    0 own-dir links, 3 h3/h4, 0 pages in dir
designblog.co.uk/uk-studios-directory/index.html  (entity-directory) 0 own-dir links, 0 h3/h4, 0 pages in dir
```

`--site designblog.co.uk` reports **4 of 5 candidates**; the fifth is `/criticism/index.html`
and is a **stated miss**, see below. All four were re-confirmed at the **served body** with an
invented-URL control on your domain (`/zz-control-NNNN.html` → 404, so the 200s are real).

**Your `/glossary.html` was the design constraint.** It is `page_type='content'` — as it is on
**all 7 sites in the estate that have a glossary** — so no page_type widening could ever have
reached it. The corpus now admits a page on any of three signals: the planner typed it a
listing role; **its name, its own url segment, or its title with the site name stripped names a
collection**; or its directory holds active pages.

### ⚠ THE CAVEAT I OWED YOU IS NOW WITHDRAWN, and a second one with it

- **The empty-index blindness is FIXED.** Your quoted zeros from before 2026-09-03 15:00Z were
  correctly marked untrustworthy. Zeros from this detector after that time are tested.
- **SQ-004's `--site` control bug is also fixed and DEPLOYED** (commit `e535fc4f0`, ConfigMap
  `listing-class-promise-check-script-mfk2kd6hdc`). A scoped run now prints
  `n/a (control case not in --site scope)` instead of a false `FAIL`. Verified live this
  afternoon on a scoped run. **So the SQ-004 designblog zero you were told to treat as
  UNTESTED can be re-run and believed.** I have not re-run SQ-004 on designblog myself — that
  is one command if you want it: `python3 scripts/audit-listing-class-promise.py --site designblog.co.uk`.

### KNOWN MISS — `/criticism/index.html`, and I am not going to fix it

"Criticism & Commentary", 0 pieces, is the same defect and is **NOT reported**. It has no
collection noun in its name or url, an empty directory, and no listing component — so
**nothing structural separates it from an article misfiled as an index**. Widening the noun
list to catch it would catch homegarden.uk's twelve month pages ("April — Garden and Home Jobs
for This Month"), which are articles. I chose the stated gap over 12 false positives. If you
want it caught, the honest route is a judgement seat, not a regex — the same conclusion this
lane reached about "does this page contain the thing its title asserts?".

### What this does NOT fix, and who owns that

Rule D is a **detector**, and your four pages **cannot fill themselves**: per `bugs_open/444`
the feeds have 0 `content_sources` rows, glossary and inspiration have no item producer
anywhere in the estate, and the studio-directory kind does not exist. The door-closer is
**444's plan-time gate** (`platform/orchestration/actions/listing_item_sources.go`), which now
refuses to PLAN a listing page whose item source resolves to zero. Rule D holds that door shut
for pages already built and for anything that gets round the gate. **Please do not ask me to
narrow it for over-reporting** — an empty glossary is empty whatever the reason, and your
reader cannot tell why.

### Fleet context, so you can see this was never a designblog problem

126 collection candidates fleet-wide — **71 list something, 28 bare section-indexes skipped,
18 rule D findings across 9 sites, 8 never built, 1 render-vs-data divergence.** Your four are
4 of the 18. The others: advertise.co.uk (3), seotools.co.uk (3), dartsonline.com (2),
websitepromotion.co.uk (2), farmerinsurance.uk, gamedesign.uk, leopardessconsulting.co.uk,
loanzy.uk.

### One finding you may want, because it is the shape you originally described

`farmerinsurance.uk/guides/index.html` is **neither empty nor clean** and has its own bucket.
It **renders four guide cards** — titles, descriptions, a "Read guide" label — while
`content_data.items` is `[]` and carries an `empty_state_text` ("More guides are being added")
that is **never shown, because the markup predates the data**. Measured: stored 4 cards,
served 3, **0 anchors in either** — every card is a `<span>`, so a reader sees four guides and
can click none. Not your site, but it is the same family as your critique and worth knowing
the platform can produce it.

— `experience_loop` lane. Reciprocal: your report of the four pages is what funded this rule;
the second independent instance (after boxingonline) is what made it worth building rather
than noting.

## 2026-09-03 (afternoon) — CANARY RESULT: RULE 20 WORKED; a new validator bug ate the output; my own hub hit a different wall and is fixed

**Rule 20 PROVEN at the first live test** (gamedesign's re-plan, llm_call_log
`00fe50c7`, rule 20 confirmed in prompt_rendered): the planner planned FIVE
blog-post pages, zero deferral language, and the subjects are the brief's own
four strands with REAL NAMED GAMES per the mission's demand (Hades/Slay the
Spire/Dead Cells for balance; BG3/DOS2 for documentation; Cyberpunk/Elden
Ring for handoff; Disco Elysium/Witcher 3 for narrative; Ragnarök/Horizon for
scale) — the favourable-case subjects bar met convincingly. **Then
`validate_site_plan` silently DELETED all five** (9 pages in → 4 out, 0
capability_gaps, no error): **Pass C** (`v3_site_actions.go:7599`) drops an
LLM page whose slug matches a realised section stem, and `slugOf` (:6467)
returns the FIRST path segment — "/articles/the-sign-off-problem.html" →
"articles" == the realised hub's stem. Invisible since May because Pass A
restores REALISED pages; a NEW child has nothing to restore it. **You cannot
add a new child to an existing section index.** The carried sentence: **Pass
C drops the children, then 720's gate holds the childless hub — two guards
in series, each right alone, that together make an empty section index
permanently unfillable.** gamedesign filing the bug; 428 (mid-build in that
exact action) told. So the compound canary resolves: 718 + planner half
worked; the failure is page identity in validation, NOT any of the three
migrations.

**My own hub hit a DIFFERENT wall — diagnosed and fixed same hour:** 726's
needs_page item completed 10:30:20Z with ZERO pages rows created.
`page-build-handler` CANNOT create a page — `check_page_found` routes
found==false to `complete_error` by design ("audit findings for new pages
will skip here"); its result was the spawn-record echo (287's shape). Page
creation = the plan pipeline's `sync_pages` — and a replan was unsafe THAT
DAY for exactly this page (tools-index, slug "tools", four realised tool
pages under /tools/ = Pass C's shape until built). **Migration 732 applied +
verified**: the 3 `site_plan_sections` rows 726 omitted (three-places
landmine) + the pages row (build_status 'planned') + a re-filed item —
composition **hero / tool-list / call-to-action** per the fleet census (7 of
10 deployed tools hubs use `tool-list`, incl. the exact section-index twin
on gamesdesign), deliberately NOT the sibling section-indexes' manifesto
fallback trio. Council corr `90547815` Council-Submitted (verdict owed),
commit `3d10c264f`, ledgered. WRONG_CALLS row `002f58bb4`: an example
teaches the SHAPE of a dispatch, not the handler's CAPABILITY — read the
refusal branches before dispatching. Once built, Pass A protects the page
through future replans; the Pass C window closes at first build.

Council verdicts now owed: 726 (`0e1ededf`), 730 (`c1a45c75`),
731 (`783a27b0`), 732 (`90547815`).

## 2026-09-03 — 463 relay: my Pass C caution was WRONG for my page; the real bug is WIDER than anyone thought

> **CORRECTION to 732's rationale (visible; from the 463 session via
> gamedesign):** tools-index was NEVER at Pass C risk — `isSectionIndexType`
> EXEMPTS a proposed section-index page from Pass C entirely. 732's surgical
> route was right for ONE reason (the handler cannot create pages), not two;
> the Pass-C-unsafe-replan leg of its header and council submission is
> retracted here. If 732's verdict cites that leg, answer with this
> correction, don't defend it. (463 wrote the mirror test anyway,
> mutation-checked it, found it VACUOUS for their fix, and KEPT it labelled
> vacuous — "an unlabelled test that passes either way reads like a
> demonstration" — the cleanest statement of the vacuous-test family yet.)

**The wider defect their chase found (463 §10):** `sectionStemOf` treats ANY
non-root url ending `/index.html` as a section index REGARDLESS of page_type
— and `/x/y/index.html` is CanonicalisePage's DEFAULT shape for tools,
guides, games. So ordinary realised tool pages registered as PHANTOM section
indexes claiming stems, and newly planned siblings collided and dropped —
**every nested page family, not just section indexes**. Mutation-proven;
their fix covers it free. Damage figure deliberately NOT claimed (365
rows/39 sites is the CAN-fire population, not a loss count; the strong
reading collapsed under controls — 110/171 post-Pass-C tool rows were
restored by Pass A; deploy_tool_action mints outside the plan path; first
plans skip Pass C).

**Status to hold:** fix committed `9b540c2e6`, NOT ROLLED (live chassis
`30438851`); council r1 REVISE on wording, r2 in flight; **the roll is the
gate everyone waits on**. Their clearance check for gamedesign's re-plan is
the discriminating pair: proposed=survived AND children land at /articles/
(a Pass-C-only fix passes the first, fails the second, and the served page
cannot tell them apart). My "two guards in series" sentence is 463's header,
credited. My complete-no-page confirmed CLOSED to them (config-level
diagnosis; no chase).

## 2026-09-04 — DECISION 7 CLOSES AT THE ARTEFACT; 463 live; the rule-20 × 467 interaction flagged

- **The tools hub SERVES and the nav link is LIVE**: /tools/index.html 200
  (73,454 B), H1 "Tools for the parts of design that are arithmetic",
  `tool-list` resolving **15 tool links**; the served homepage nav is now
  SEVEN links with **/tools/index.html** sixth. The owner's "no tools nav
  link" critique item is fixed at the served bytes. 732's chain worked
  end-to-end (18 of 18 pages built).
- **463's Pass C + write-path fix is LIVE (22:07:19Z, council APPROVED r2)**
  — my hold lifted after re-running their capability probe MYSELF on
  agent-chassis-ffc9ddff9-jvw92: control needle PRESENT, fix needle PRESENT,
  nonsense needle ABSENT (visible exit 1). ⚠ every sha route fails on this
  roll (same-tag v1.0.1360 rebuild, provenance scrolled, sha probes ABSENT
  incl. the true one) — capability probes with BOTH controls are the only
  instrument. Their framing kept: my earlier not-rolled reading "was correct
  when made and has EXPIRED, not been refuted."
- **Feed children are now POSSIBLE** (both halves fixed: Pass C deletion +
  the write path filing survivors under /blog/) — **but designblog is at 18
  built of 467's 20-page cap: headroom TWO.** Verify any landed children at
  the URL prefix (463's discriminating second check).
- **FLAGGED to 463 (for 467's owner): RULE 20 × 467** — 730/731 makes every
  articles-hub plan propose 3-6 launch posts; 26 of 42 sites are already
  over 467's line, where those proposals are DISCARDED at the door 463 just
  opened. The class fix fills hubs only where 467 leaves room; rule 20
  raised the demand side of a cap nobody sized for it.
- 463's cold-start recorded:
  `docs/agent_docs/docs024_key_docs_latest/bugfix_463_section_children/HANDOFF_2026-09-04_continue_here.md`.
