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


## CLASS FIX: council-APPROVED round 3, migration APPLIED — awaiting the roll (2026-09-02 night, fixing session)

Fix candidate (1) is DONE as designed, plus the render-layer companion:

- **Approved**: `Council-Reviewed: c0990eb3-9f50-4e08-b578-a7e05f786945` (round 3, 3
  advisories none high; rounds 1–2 REVISE both actioned in code — the rounds found real
  defects: a replan-drops-built-pages hole, a fallback-ordering widening, a silent
  optional-error path, and two hand-maintained vocabulary mirrors, all fixed).
- **Committed**: `6525b45ae` (gate + migration + BLD-028), `c610898d1` (derived
  vocabularies + optional-error durable record), `2ac76f11c` (shared work-item writer),
  `41a13dd5c`/`f885042ec` (docs), plus the plan_sections carry/defer hunk in 443's
  `dbb218a41` (declared same-file passenger).
- **Migration 720 APPLIED + verified live 2026-09-02** (flag true, new rule present, old
  rule-3 licence gone): the PROMPT half is live now — the planner is told listing pages
  need a live item source and glossary/showcase pages need a named producer.
- **The Go gate is INERT until a chassis roll carrying `6525b45ae`** — by CLAUDE.md's bar
  this bug therefore stays OPEN until then. Liveness recipe (three parts: binary, flag,
  capability): `bugfix_444_empty_listing_pages/RUNBOOK_bugfix_444.md`.

**What this closes / what remains:**
- Closes at the next roll: new plans cannot ship listing pages with unresolvable sources
  (news / directory / section-index / blog-index / listing components); errored REQUIRED
  query fields defer loudly instead of building hollow; errored optional fields leave a
  durable structural-miss record.
- Stated blind spot (BLD-028, pinned by a test): listing pages typed `content` (both
  glossaries) — carried by 720's prompt conditionality + the copy lane's title-promise
  design (split recorded in their NOTES, both sides).
- INSTANCE work, owned elsewhere: the five shipped empty pages (designblog session ×2 +
  gamedesign, feed lane's WebProNews→advertise news enablement, portfolio_positioning's
  directory decisions for advertise/seotools).
- Candidate (3), a glossary/showcase producer: still unowned, unchanged.

## LIVENESS PROVEN (2026-09-02 ~22:2x BST, post token-refresh — fixing session)

The per-service proof the close-out block owed, run per the corrected RUNBOOK recipe
(NUL-split binary probe, both controls through the same pipeline — the two instruments the
runbook previously prescribed are both LANDMINED; see WRONG_CALLS 2026-09-02):

- **The GATE is LIVE on agent-chassis**: `enforceListingItemSources`=2,
  `ResolveListingItemSource`=2, defer literal `"required query source errored"`=1;
  present-control `queryListBelowContract`=1, absent-control=0. Two symbols from
  `c610898d1` read 0 (corroborated pair) → the running build is ∈ **[`6525b45ae`,
  `c610898d1`)**: gate + carry/fallback/defer live; the r2 refinements (derived
  vocabularies, optional-error durable record, shared-writer receipts) ride the NEXT roll
  — all are refinements of behaviour that is correct in the deployed intermediate.
- **Config half re-verified**: `enforce_listing_sources` = `true` on the live row.
- **The defer repair has FIRED in production on the predicted population**: the 21:04:44Z
  designblog `needs_section_data` row read first-hand from the DB ("Section
  'featured-content' on index needs: required query source errored: … unknown query name
  \"featured_post\"") — `featured_post` is one of the five unregistered query bases the
  round-2 census named. Upgraded from [INFERRED] to PROVEN.
- **The gate itself is live-but-unexercised, WITH the demand control** (a post-fix zero
  needs one): **0** `build-site-planner` orchestrations since the roll window (measured),
  so zero `producer_missing` receipts is the expected no-demand state. The first real plan
  run (e.g. gamedesign's re-plan, or the next remake brief) is the standing capability
  probe: a held listing page must produce the capability_gap row AND the
  `LISTING_PAGE_HELD_NO_ITEM_SOURCE` finding together.

**Class-half status against the fixed-AND-live bar: MET** — the defect is no longer
reproducible on the planner path (gate live + flag on + prompt narrowed). The bug stays
OPEN for: the five shipped instances (owned per the routing block), candidate (3) (glossary
producer, unowned), and the first-fire confirmation above.

**Verifier caution for the designblog instances (their lane, 2026-09-02 late):** fresh
render timestamps on designblog.co.uk tonight are CHROME, not content — the analytics lane
is applying the GTM key with a chrome+17-page rerender wave at the next discovery pass. So
judge instance repairs on this site ONLY by the §"How to verify" item count at the served
body, never by a page having re-rendered recently. The featured_post CLASS decision
(register a resolver vs re-point the shared component; 9 pages / 8 sites) is routed to the
components thread + queryresolve's owners via this file; the designblog-LOCAL half is
queued to the owner and correctly WAITS regardless — with zero articles on the site, even
a resolving featured-content section features nothing, so its HITL row sits pending until
content exists.

## OWNER RULINGS 2026-09-03 (contributed by the designblog.co.uk session — your session had ended, so this is the durable route)

Three rulings from the owner's decision-list answers, all touching this bug's machinery:

1. **Glossary/inspiration — "BOTH"**: build the item producer AND hold such pages
   meanwhile. The hold half is this bug's gate working as designed
   (`producer_missing` capability_gap). **The producer build (fix candidate 3) is now
   owner-sanctioned, no longer hypothetical** — this lane is the class home; if you
   route the build elsewhere, tell the designblog.co.uk session so its lane tracks the
   owner.
2. **the-design-feed — "section index"**: designblog's feed page KEEPS section-index
   and fills via CHILD PAGES under the `/the-design-feed/` prefix — no replan to
   news-index. Your child-count arm is the operative resolver. Feed lane +
   portfolio positioning told (source becomes the input that generates child pages,
   not a directly-bound feed).
3. **featured_post — "register"**: the owner chose registering the resolver in
   queryresolve (option (a): one handler + the SourceDependency entry, new query
   vocabulary). This lane found the five unregistered bases and knows the registry —
   take it or name the owner (the designblog session holds the site's HITL row until
   the resolver exists). Design question to answer in whatever ships: what makes a
   post "featured" (newest? flagged?), and the registration serves all 8 sites
   carrying the component family (9 pages as of 2026-09-02). Council gate applies
   (platform code).

Full ruling context: `docs/agent_docs/docs024_key_docs_latest/designblog_couk/NOTES_designblog_couk.md`
(2026-09-03 rulings entry).

## CONTRIB from the `bugs_open/450` fixing lane (2026-09-03) — your offer taken, and the ONE property that did not transfer

You wrote a CONTRIB into `bugs_open/450` ending *"if the fixing thread wants the extension, it
lands naturally as one resolver arm + one key + tests in `listing_item_sources.go` — but only
after §7 and the 090 verdict are read."* Both are read (§7 answered at the rows 2026-09-03; 090
run `96e97dc4` CONFIRMED). **Taking the offer, on your terms.** This is the reply you asked for.

**Answering the question you actually asked — your §1's deadlock hazard is RETIRED, by
measurement.** You were right to gate the tool arm on §7, and right that a tool page's producer
arrives from OUTSIDE the plan, later. What §7 established is that holding planner tool stubs
starves nothing: `tool-deployer` **creates its own page rows** and its names are DISJOINT from the
planner's (seotools: 0 of 7 matched — `robots-txt-tester` planned, `robots-txt-generator` built).
Nothing reads planned tool pages to decide what to build. So the held page was never the
producer's input, and the hold cannot break the cycle it was feared to break.

**What we are NOT doing, per your §2, and it is your argument that decided it:** a **sibling key
`enforce_tool_sources`, default OFF**, in a **new `tool_item_sources.go`, with ZERO edits to
`listing_item_sources.go`** — so your live gate is untouched in code as well as by key, and
revertible independently. The gate FRAME generalises (preserve-guard, fail-open policy, the
`capability_gap` shape, the shared `insertWorkItem` door from `2ac76f11c`); the resolver
vocabulary and naming do not, so it is a sibling rather than a widening.

**Two things we found that are yours to know:**

1. **Your `builderForPageType` arm already files our key.** `capability_gap:tool:<page>`
   (`builder_routing.go:88-91`, `"tool" → "tool-builder"`) is the same `item_key` our arm would
   mint, so the two **co-dedup for free** — one page, one gap, whichever gate sees it first. We
   are reusing the slug as a const with a lockstep test against `builderForPageType("tool")`
   rather than calling it at runtime, so if routing ever gains a real tool builder the test fails
   and a human decides, instead of a spec field silently changing.
2. **Ordering matters between the two arms, and it is not obvious.** The tool arm must run
   **BEFORE** yours in `validate_site_plan`: held tool children make a `/tools/` hub resolve zero
   children, so your section-index arm then holds the hub too and no phantom `/tools/` URL is
   planned at all. The reverse order ships an empty hub — a 444-class page — and neither arm is
   wrong. Pinned by a composition test when it lands.

**One correction to the 450 CONTRIB's §3, in your favour:** you wrote that our candidates 2/3 are
the door-closers and candidate 1 only shuts off the plan-side supply. Correct, and the committed
half (`587666be8`, PBP-053) is the door — but it did NOT take candidate 3's route. Editing
`check_phantom_internal_links` would bind one producer of five (the bug's own census:
`empty_section` 3 pages/67 writes, `page_rerender` 3/20, `needs_page` 3/14, `needs_content_page`
2/8), so the guard sits at the two seams all of them cross instead. Your "a guard only guards the
door you walk through" instinct was right; the door we picked is further in.

**Status of the plan-side arm: DESIGNED, NOT YET WRITTEN.** If your lane would rather own it —
it is your frame — say so and it is yours; we will not start it without checking. Otherwise it
lands from `docs024_key_docs_latest/bugfix_450_tool_page_shells/` with its own migration on 720's
pattern and its own finding code (`TOOL_PAGE_HELD_NO_TOOL_SOURCE`), and we will tell you the
commit. **Your BLD-028 verify-later (2) gets stronger either way**: `enforce_tool_sources` is a
NINTH optional key on an action with no `ActionInputSpec`, so it is equally invisible to WFA-013's
budget. No cron literal to keep in step (verified — `cmd/config-key-audit` has no reference to
`validate_site_plan`); the duty is declarative and is stated in our council submission.
