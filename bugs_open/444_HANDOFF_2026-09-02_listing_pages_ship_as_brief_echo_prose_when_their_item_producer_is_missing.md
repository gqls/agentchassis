# 444 — Remake listing pages ship EMPTY of their content type, filled with brief-echo prose

**Filed 2026-09-02 by portfolio_positioning**, from the owner's designblog.co.uk critique
("explaining the brief and not answering it") — verified same evening on advertise.co.uk, so it
is a CLASS across the day's four remakes, not a designblog instance. First-hand verification is
substituted for a 090 run per the 2026-07-31 owner ruling: every claim below is a direct
measurement of the served body, the live DB, or the concept register, made tonight; the one
absence claim is marked as such.

## Symptom (measured at served bodies, 2026-09-02 ~20:xx)

Listing-type pages serve 200 at full page weight but contain ZERO items of their own content
type, replaced by meta-prose about what the page WILL contain:

- advertise.co.uk `/channels-directory/index.html` — **0 entries** (plan said 15–20); one h2
  "Find UK advertising channels by type", then footer.
- advertise.co.uk `/glossary.html` — **0 terms**; h2/h3s are "What this glossary covers",
  "Where the terms come from", "Reading a definition".
- advertise.co.uk `/news/index.html` — **0 items**.
- designblog.co.uk: `/glossary.html` 0 terms · `/inspiration/` 0 showcases ·
  `/the-design-feed/` 0 items · `/uk-studios-directory/` 0 studios (that lane's own
  verification, `designblog_couk/CRITIQUE_2026-09-02_owner_site_review.md`).

The wrong result looks RIGHT: 200, ~60KB, plausible headings — every naive check passes.

## Root cause — three distinct mechanisms, one symptom (per listing type)

1. **Feed pages: the mechanism exists and is LIVE but per-site UNDRIVEN.** NEWS-001 fills feed
   pages from `content_sources` rows; idea.uk has 5 and a working feed. All four remake sites
   have **0 rows** (measured). Nothing in the build path creates a source row, and no sweep
   ever will — this never converges on its own.
2. **Directory pages: the mechanism exists and is fully wired but the KINDS do not.** DIR-001:
   one kind = one fleet-wide list; six kinds live (model/company/protocol/mortgage-lender/
   savings-provider/health-insurer — counts 51/36/33/28/12/5 tonight). "advertising-channel"
   and "uk-design-studio" are not kinds. Adding one is data-only across SEVEN places (two fail
   SILENTLY — see LANDMINES + directory-pipeline.md) plus a researcher run to populate. No
   sweep adds kinds.
3. **Glossary / inspiration-showcase pages: NO item producer found at all** [ABSENCE CLAIM —
   basis: no glossary/terms/showcase table in information_schema, no concept-register entry,
   grep of register + platform actions; a fixing thread should re-verify]. These were planned
   as `content` pages and the generic writer wrote prose.

**The copy half [INFERRED, writer prompt unread]:** the brief-echo prose is DOWNSTREAM of the
missing items — a section writer given a listing section with no items writes about the
intent. Pages with items (idea.uk news) do not show the pattern. The copy lane can confirm at
the prompt; fixing copy without items would produce nicer prose about an empty page.

> **CORRECTED 2026-09-02 (same night, by the fixing thread "bugs_open/444" — their code-read,
> credited):** the "writer proses over missing items" mechanism holds for the GLOSSARY only
> (generic-text-block, LLM prose). The DIRECTORY pages are NOT writer output: `resolveBusinessDirectory`
> deliberately ERRORS on a missing exporter config (206's loud-failure rule) and `plan_sections`'
> error branch deliberately bypasses `on_missing` (054's don't-mask-errors rule) — **two correct
> guards in series produced the silent hollow section both exist to prevent.** The NEWS page's
> items field is `required:false` BY DESIGN (client refresh is the freshness path), so its
> emptiness is legal at the render layer. My §"copy half [INFERRED]" was right to carry the
> marker: correct for glossary, wrong for the other two. Their conclusion independently confirms
> fix candidate (1): only PLAN-TIME validation closes all three mechanisms. Also: a FIFTH
> instance, seotools.co.uk /directory/index.html (their served-body measurement), and the
> discriminator — vetcomparison's identical bare directory-listing is FILLED because it alone
> has a `directory-json-exporter` config row. Fix plan:
> `bugfix_444_empty_listing_pages/PLAN_2026-09-02_listing_source_gate.md`.

## Related, same build-path family (measured tonight)

No remake got a TOOLS nav link or a `/tools/index.html` hub: advertise's nav is
index/guides/news/channels-directory/glossary with `nav_order` 4 conspicuously absent, while
seven tool pages serve. Tool pages arrive post-plan via tool-deployer; the nav rebuild ran
before they existed; the plan never planned a tools hub. (311 is CLOSED — the planner CAN see
library tools on sites that have them; the residue is ORDER: plan before tools exist.)

## Fixing-thread findings (2026-09-02 late evening, session "bugs_open/444")

The class fix now has a thread (this file's candidate (1) — the portfolio_positioning
handoff recorded it unowned). Working docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_444_empty_listing_pages/`. New findings,
all measured tonight unless marked:

- **A FIFTH instance: seotools.co.uk `/directory/index.html`** — same headline-only
  `directory-listing`, measured at the served body and the `page_components` row. Bare
  `directory-listing` is planned on **4 sites as of 2026-09-02**: the three empty ones plus
  vetcomparison.uk, which is FILLED — the discriminator is below.
- **Mechanism (2) refined — the emptiness is not a missing contract but two correct guards
  composing in series.** The `directory-listing` schema ALREADY declares
  `entries: {source: query.business_directory, required: true, min_items: 1, on_missing:
  skip_section}` (since 2026-08-08). It didn't fire because `resolveBusinessDirectory`
  returns an **error** (not an empty list) for a site with no `directory-json-exporter`
  config row (deliberate, bugs_open/206 council round: misconfiguration must be loud), and
  `plan_sections`' error branch deliberately does NOT route errors into `on_missing`
  (bugs_open/054: an errored resolve must not be masked as no-data) — one Warn log, field
  unresolved, section stays READY. Verified: the build orchestration's stored
  `section_plan` (`d0a858be-…`) shows directory-listing ready/headline-only/no
  resolved_data, and the resolver's own lookup SQL re-run tonight returns a config row
  only for vetcomparison. Build-time logs unrecoverable (pods restarted after the build).
- **Mechanism (1) refined — news-listing's emptiness is LEGAL by design.** Its `items`
  field is `required: false, on_missing: skip_field` because the client-side JSON refresh
  is the freshness path between rerenders; `items: []` in advertise's stored content_data
  confirms the path. So no render-layer contract can close the news case without breaking
  the legitimate empty-between-feed-runs state on sites WITH sources.
- **Why the four remakes have no sources: the classifier cannot reach these verticals.**
  `matchVerticalNews` reads industry/site_type/category + domain substrings; none of the
  four remakes' current classification specs carry `content_features` AT ALL (measured).
  idea.uk's working feed exists because its `content_features.news_feed` was HAND-AUTHORED
  2026-08-25. The planner plans a news page independently of the classification driver
  that would feed it — two mechanisms that never meet.
- **Consequence for fix design:** the render layer either cannot (news) or must not
  (directory — both guards are correct) close the class alone. Candidate (1), plan-time
  validation, is confirmed as the only door that closes all three mechanisms. Secondary
  repair identified: in `plan_sections`, an errored resolve of a REQUIRED field should
  DEFER the section (the existing loud HITL path) instead of leaving it ready — preserves
  the 054/206 distinction while ending the hollow-section outcome.
- The glossary/showcase absence claim RE-VERIFIED (information_schema + non-test Go grep:
  zero producers). The glossary pages are typed `content` with `generic-text-block` —
  invisible to any page_type/component-based gate; that half stays with the planner-prompt
  conditionality + the copy lane's title-promise design (split already recorded in their
  NOTES).

## Fix candidates, ordered by what closes the door

1. **Plan validation refuses (or explicitly degrades) a listing page whose item source resolves
   to zero** — kind missing, no content_sources row, no producer → a `capability_gap`/HITL item
   naming the missing producer, never a built page of meta-prose. Makes the bad state
   unrepresentable; the existing `capability_gap` carrier fits.
2. **Per-vertical enablement becomes part of the build path**: a brief that plans a directory
   page triggers the kind-creation checklist (7 places); a feed page triggers a
   content_sources seed (owner already flagged WebProNews for advertise — with the feed lane).
3. **A glossary/showcase item producer** — new build, smallest honest scope TBD by whoever
   takes it; until then briefs should keep those page types explicitly conditional.
4. (Weakest) copy-side: teach the writer to refuse listing sections with no items — treats the
   symptom at the last hop.

## How to verify a fix

The advertise pages above are the standing repro: a filled `/news/index.html` after a source
row + feed run; a filled `/channels-directory/` after the kind + researcher run; glossary needs
(3). Judge at the served body, item count > 0 — never at page status or byte size.

## Ownership / routing (as of filing)

designblog instances: the designblog.co.uk session. advertise instances: portfolio_positioning.
Class fixes: (1) build-pipeline/plan-validation owners; (2) kind additions per DIR-001's
runbook + feed lane for sources; nav/tools-hub: site-planner + the 149 nav-membership family.

## CONTRIB 2026-09-02 ~20:50Z — a FIFTH site and a FOURTH mechanism (gamedesign.uk lane)

`gamedesign.uk/articles/index.html` (fresh FRESH-path build, live 18:00Z) — 200, 8,396 B,
**zero articles**, body = the mission brief's constraints as prose, including a "What they avoid"
list (negative-identity copy, owner-banned 2026-09-02). Not feed, directory or glossary: the
content type is ARTICLES, and the plan created ONE article page with ZERO sections (parked by
`mark_no_ready_sections`, then owner-cancelled) — so the type has no producer because **no content
pages exist at all**. An editorial site with zero editorial. Candidate 1 (plan-time validation)
would catch it as "section-index whose section contains 0 planned content pages". Instance is owned
by the gamedesign.uk lane and filed in full as `bugs_open/446` §3.2 (which also carries the owner's
wider critique: no game imagery — a SPEC error the lane wrote itself — a hero over a 404, and the
class "nothing measures a site's energy against its vertical"). Not re-verified through the 090
loop; measured at the served bytes with a same-domain 404 control.
**Addendum ~21:05Z, for the 444 session's resolver:** the page's `page_type` is **`section-index`**, not `blog-index` — the planner typed an editorial site's articles hub as a generic section index. ~~A refusal keyed on `blog-index` alone will not hold this shape~~ **CORRECTED ~21:15Z (designblog relaying the 444 session): the resolver DOES hold it — section-index is a first-class arm (zero pages under the `/articles/` prefix in plan+realised → held, `builder_needed=section_children:<page>`), and the blog-index keying was our shared inference, not their code.** One nuance they asked for: if a hub carries `content-listing`, that arm resolves by SITE-WIDE `query.blog_posts`, unscoped, so a site with posts elsewhere legitimately serves a non-empty hub. Measured on gamedesign.uk's hub: components are `hero` + `call-to-action` (+ generic text) — **no `content-listing`**, so the section_children arm is the one that applies, and it would have held this page.
**The two facts the 444 session asked for (~22:20Z):** hub `page_type = section-index`;
`pages.sections = ["hero", "generic-text-block", "call-to-action"]`, components bound
`hero / generic-text-block / call-to-action` — **no `content-listing`**, so the section_children
arm applies. Plan rows (current plan, 17:33:13): `articles-index` role section-index slug `articles`;
the one content page `article` role blog-post slug `article` **with `parent_section` EMPTY and url
`/blog/article.html`** — i.e. not under the `/articles/` prefix at all. Zero children by prefix in
plan AND realised, so the gate (`6525b45ae`, inert until roll + migration 720) would hold it as
`builder_needed=section_children:articles-index`. Two mechanisms then, not one: no content pages
exist, AND the one planned was parented nowhere. **Caveat for the record:** this site is being
re-planned right now (corr `aab87c0c`, brief v2 asking for real articles) BEFORE the gate is live —
that plan is read by hand against the same predicate, not by the gate.

