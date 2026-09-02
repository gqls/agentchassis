# NOTES — bugfix 444 (empty listing pages / brief-echo prose) — append-only, newest at bottom

## 2026-09-02 (evening) — session "bugs_open/444" opens the class-fix lane

**Ownership check before starting** (`who-owns.py 444` + ListAgents): portfolio_positioning
filed the bug and handed off at 20:25 same evening; its handoff §3 says fix candidate (1)
"has no owner yet". The Portfolio positioning session (alive, idle) confirmed by
cross-session message: bug file + handoff + NOTES (o)/(q) are the complete state, nothing
else in flight from them. designblog.co.uk session routes its own per-site instances
(feed source via feed lane; uk-design-studio kind via DIR-001 checklist) — the class
mechanism here must compose with that, not collide.

**Bug re-verified STILL VALID tonight (all direct measurements):**
- `content_sources` (active) per site: advertise.co.uk 0, designblog.co.uk 0,
  websitepromotion.co.uk 0, seotools.co.uk 0; idea.uk 5 (the working control).
- `directory_entities` active kinds unchanged: mortgage-lender 49, model 36, company 32,
  savings-provider 28, health-insurer 12, protocol 5 — no advertising-channel, no
  uk-design-studio. (**6 kinds as of 2026-09-02**.)
- advertise.co.uk/glossary.html served body: 0 terms, headings are still the meta-prose
  set ("What this glossary covers", "Where the terms come from", "Reading a definition").
- Absence claim re-verified per the filing thread's request: information_schema has NO
  glossary/term/showcase/inspiration table (only two `bak_cc_portfolio_showcase_*` backup
  tables from July, which are component backups, not producers).
- `capability_gap` is a LIVE item_type (e.g. `capability_gap:no_handler_for_audit_category:
  nav_restructure` open row) — the carrier fix candidate (1) names exists.

**New mechanism findings beyond the bug file:**
- NEWS-001's enablement path EXISTS end-to-end: the 6-hourly content-feed-trigger selects
  sites where classification `content_features.news_feed.recommended` is true, and
  seed_content_sources creates the rows from the spec. The four remakes have NO
  `content_features` key in their current classification spec at all (measured tonight).
  idea.uk's was HAND-AUTHORED 2026-08-25 because `matchVerticalNews` reads
  industry/site_type/category + domain substrings only and cannot reach these verticals.
  So the feed half of 444 is precisely: **the planner plans a news page independently of
  the classification driver that would feed it** — two mechanisms that never meet.
- Live page_type vocabulary on the affected pages: `news-index` (advertise /news/),
  `entity-directory` (channels-directory, uk-studios-directory), `content` (both
  glossaries — semantically listing pages typed as generic content), `section-index`
  (designblog /inspiration/, /the-design-feed/ — listing child pages that don't exist).
  So designblog's "feed" page is not even a news-index — a page-type-blind fix at the
  writer would miss it; the classes to catch are per ITEM SOURCE, not per page_type alone.
- A content-quality-audit verdict already flags "glossary index page … describe themselves"
  on another site (open `needs_content_planning` row) — the symptom is already DETECTED
  downstream, but only as deferred HITL prose after the page shipped. Detection is not the
  gap; refusal/degradation at plan time is.

## 2026-09-02 (late evening) — the render-layer mechanism, fully diagnosed

**A FIFTH broken instance, not in the bug file:** seotools.co.uk `/directory/index.html`
serves the same headline-only `directory-listing` (measured at the served body + the
`page_components` row tonight). Bare `directory-listing` is planned on FOUR sites
(**4 as of 2026-09-02**): advertise, designblog, seotools (all empty) and vetcomparison.uk
(FILLED — "24 Hour Vetcare" etc. at the served body).

**The discriminator is `query.business_directory`'s config lookup.** The `directory-listing`
component's schema ALREADY declares `entries: {source: query.business_directory,
required: true, min_items: 1, on_missing: skip_section}` (schema updated 2026-08-08, well
before the 09-02 builds) — the bugs_open/054 contract SHOULD have skipped these sections.
Why it didn't (verified by code read + live artefact + the resolver's own SQL re-run):

1. `resolveBusinessDirectory` (queryresolve/business_directory.go:104) returns an **ERROR**,
   not an empty list, when the site has no `directory-json-exporter` scheduled_tasks row —
   a DELIBERATE, council-reviewed choice (bugs_open/206 round 1: a missing config is a
   misconfiguration and "must surface as a loud failure, not a silent hollow section").
   Re-ran the lookup SQL tonight: only vetcomparison has the config row; advertise/
   designblog/seotools take the error path by construction.
2. `plan_sections`' query-error branch (plan_sections_action.go:~2710) DELIBERATELY does
   not route errors into on_missing (bugs_open/054: an errored resolve must never be
   masked as "no data") — it logs one Warn, leaves the field unresolved, and the section
   proceeds as **ready**.

**Two individually-correct guards compose in series into exactly the hollow section both
were built to prevent.** The observed artefact matches the error branch precisely: the
built section entry in the orchestration's `section_plan` (orchestration
`d0a858be-2b18-4639-83d1-3d5301c115f5`, page-build-handler, 13:48) has status "ready",
llm_fields=[headline] only, NO resolved_data, skipped_count 0. Build-time pod logs are
gone (chassis pods restarted 15:39/15:53, after the 13:49 build), so the Warn line itself
is unrecoverable — [INFERRED from artefact shape + code path + config-lookup re-run; all
three independently point at the same branch].

**news-listing is a THIRD shape:** its `items` field is `required: false, on_missing:
skip_field` BY DESIGN (schema notes: the JSON + client-side refresh is the freshness path
between rerenders — items legitimately arrive later without a rerender). So an empty news
archive stores `items: []` and renders empty legally; advertise's news-listing
content_data confirms (`"items": []`). No render-layer contract can close this class —
for a site with sources, empty-now is correct.

**Conclusion the fix design rests on:** the render layer either cannot close the class
(news: empty is legal by design) or should not close it alone (directory: the 054 and 206
guards each protect against masking failures; loosening either reopens a worse bug). Only
PLAN-time validation — "does this listing page's item source resolve for THIS site?" —
closes all three mechanisms. This independently confirms the bug file's fix candidate (1)
ordering, with a new secondary repair now visible: an errored resolve of a REQUIRED field
should DEFER the section (the loud HITL path plan_sections already has) rather than leave
it ready — errors stay distinguishable from no-data, and the section no longer ships
hollow.

## 2026-09-02 (night) — implementation, council round 1 REVISE, round 2 in

**Implemented** (all green: `go build` + 11 new tests): `listing_item_sources.go` (registry
+ gate + capability_gap emitter), `v3_site_actions.go` tail (opt-in call), `plan_sections`
error branch (carry→fallback→defer; the main hunk rode bugs_open/443's commit `dbb218a41`
as a DECLARED same-file passenger — coordinated by message, their commit records it),
`HasBusinessDirectoryConfig` export, migration `720` + `_ROLLBACK`, BLD-028, finding-code
declaration, doc_notes decision row (pipeline/listing-class-promise).

**Missteps / traps this session hit (checks attached):**
- **Migration number taken TWICE mid-session** — drafted as 718, then 718 AND 719 landed
  from other lanes within the hour. Check: `ls sql_for_agents | grep -E '^NNN'` at WRITE
  time and again at COMMIT time; the designblog session's ping was what saved the collision.
- **The seed is not the system, again**: the council submission quoted the SEED's "Plan the
  IDEAL site regardless" sentence (053) as the licence to retire — the LIVE row does not
  contain it; the live licence is rule 3's "may have empty sections arrays". Caught before
  the migration was written by reading the live `prompt_template`; the migration anchors on
  the LIVE text with exact-count guards. → WRONG_CALLS entry.
- **A whole-file JSON rewrite hiding in a one-entry edit**: python json.dump with default
  settings reformatted all 384 lines of finding_code_registry.json (indent + unicode
  escaping). Check: `git diff --numstat` after ANY scripted edit to a shared file — the
  added/deleted counts must match the size of your intended change (8/0, not 389/381).

**Council round 1: REVISE** (gating: bug_historian HIGH — replan could drop a BUILT listing
page). Answered in code, not argument: realised pages are NEVER dropped (preserve-guard
rule + receipt; test `TestEnforceListingItemSources_RealisedPageKeptWithReceipt`). The
round found one real behavioural widening (render_guardian: defer had jumped ahead of a
declared fallback) — fixed: carry→fallback→defer, and measured that ZERO of the 17
required/min_items query fields fleet-wide declare a fallback (**17 fields / 0 fallbacks as
of 2026-09-02**). Guardian's blast-radius HIGH answered by enumeration: `plan_sections`
runs in exactly 3 live workflows (page-build-handler, page-content-writer, page-rebuild;
page-rerender does NOT field-resolve); the deferral wave on rebuilds is bounded to **11
live pages as of 2026-09-02** (category-listing 2/1 site, featured_article 9/8) — every
one carrying an UNREGISTERED query base (category, category_posts, featured_post,
comparison_results, affiliate_products; `queryresolve.Resolve` errors on unknown names at
`queryresolve.go:413`), i.e. every deferral is a genuine hollow render of 444's own class
found by the repair. Rollback lever: per-component data-live opt-out (required:false).
validate_site_plan's callers: build-site-planner + site-planner (**2 as of 2026-09-02**);
only the first is armed by 720, by stated design. **Round 2 resubmitted** same trail
(RESUBMIT_CORR c0990eb3), run envelope `95b3603b`, ~20:5x. Commit goes out under
`Council-Submitted:` per the 2026-07-30 rule (never hold coherent code for a verdict).
