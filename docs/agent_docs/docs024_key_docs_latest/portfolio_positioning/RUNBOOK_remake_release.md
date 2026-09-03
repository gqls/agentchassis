# RUNBOOK — releasing a remake build (proven on №1–4, 2026-09-02)

Every command below was got wrong at least once on 2026-09-02 before being got right; the
gotcha rides with each step. Sequence: brief approved by the owner → this runbook → build runs
itself → serve-verify at the body.

## 0. The site row must be COMPLETE before the brief is even fired

```sql
INSERT INTO sites (domain, name, network_id, status, email, locked_at)
VALUES ('<domain>', '<domain>', '00000000-0000-0000-0000-000000000002', 'test',
        '<name>@contactforsales.com', now());
```
- ⚠ `name` + `network_id` NULL → the release stalls WEEKS later at `ensure_site_record`
  ("Scan error on column \"name\"") — upsertSite scans both bare and never backfills.
- ⚠ email NULL → `needs_section_data` human-review stall at the contact page. Estate
  convention `<shortname>@contactforsales.com` (11/12 recent sites).
- LOCKED until release; the brief-writer builds nothing on a locked site by design.

## 1. Release (the owner's word arrives)

Per the review item's own `spec->>'how_to_release'`. One transaction for the work items,
a SEPARATE one for the unlock — unlock is the release act and goes LAST:

```sql
-- TX1: close the review with the approval recorded; create the research item.
UPDATE site_work_items SET status='complete', approved_by='owner', completed_at=now(),
       result = result || '{"released": true}'::jsonb  -- + quote the owner instruction
 WHERE id='<review item id>' AND status='needs_human_review';
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, spec, priority,
        handler_agent, status, created_by, item_key, pipeline, triaged_at)
VALUES ('<site_id>', 'brief-release', 'needs_domain_research', 'high',
        'Research and classify domain', '{"domain":"<domain>","reason":"brief_released"}',
        5, 'domain-research-classifier', 'triaged', 'portfolio_positioning',
        'research_<domain>', 'build', now());
-- TX2, the deliberate act:
UPDATE sites SET status='active', locked_at=NULL WHERE id='<site_id>'
  AND status='test' AND locked_at IS NOT NULL;
```
- ⚠ Guard every UPDATE on the exact prior state; assert row counts in a DO block —
  a verify block of bare SELECTs cannot stop a COMMIT.
- ⚠ site_specs supersedes: retire-then-insert as SEPARATE statements in one transaction.
  A chained CTE trips `idx_site_specs_current` (measured; the CTEs share a snapshot).
- ⚠ `needs_section_data` and its kin are HANDLERLESS notifications —
  `swi_no_handlerless_promotable` refuses to re-queue them. Supply the datum, complete the
  item, let the page's own build item read the spec.
- Dispatch: `build-pipeline-trigger` ticks every 30s; claim within ~1 min. The selector needs
  `locked_at IS NULL`, item `status IN ('triaged','approved')`, no claimed sibling item.

## 2. What the build does on its own (don't intervene)

Classifier (reads the owner's mission_brief, writes 4 aspects, brief untouched) → vertical
research → strategy → briefing → site plan (retry survives transient failures; attempt/3) →
discovery sweeps interleave (design-discovery files evaluate_tools ON ITS ROTATION — a site
can DEPLOY before its tools arrive; the owned_page_review holds are correct meanwhile) →
pages/imagery/rerenders → `deployed` by pipeline action.
- ⚠ `owned_page_review` "not_built" for a tool whose component IS deployed = the validator
  read an earlier surface; verify at page_components bytes, close with evidence.
- ⚠ `unresolved_cta` piles are the build-order effect (pages before hubs); the rerender wave
  re-resolves them — batch-verify at the SERVED body after convergence, then close.
- ⚠ Content validation blockers: the issue detail is in agent_error_log's SECOND row
  (severity `warning`, "see context.issues"), not the CONTENT_VALIDATION_FAILED row.
- ⚠ On any P5-PAIR site (cross-TLD twins): `tool-suggester` reads identity+classification
  ONLY — it cannot see a seat split and will propose the SIBLING'S tools by name
  (happened on gamedesign.uk 2026-09-02, `bugs_open/447`, held reversibly). Until 447 lands,
  eye every evaluate_tools/add_tool wave on a paired site against the pair rule; the
  brief-fidelity auditor RECORDS the violation but dispatches nothing.
- ⚠ 444's repair: council corr `c0990eb3` **APPROVED, round 3, 20:53:22Z 2026-09-02** (read by
  this lane at `orchestration_states`: `complete_approved` COMPLETED); **migration
  `720_planner_listing_source_gate.sql` APPLIED the same night — the fixing session's committed
  close-out `2d7a98446` (bug file + their RUNBOOK_bugfix_444 liveness recipe) says applied AND
  verified on the live row** [their verification, not this lane's — our token expired at
  21:08Z before `enforce_listing_sources` could be re-read here]. So the PROMPT half is live for every brief fired from
  now: the planner is told a listing page needs a live item source and a glossary/showcase
  needs a named producer. ~~The Go gate (`6525b45ae`) is inert until a chassis roll carries it~~
  **The Go gate is PROVEN LIVE (444 session, `560a24c07`, ~21:2xZ 2026-09-02): the running
  chassis build ∈ [`6525b45ae`, `c610898d1`)** — gate symbol present via the NUL-split probe
  with both controls coherent; flag re-verified `true`. So for every brief fired from now the
  MECHANICAL gate holds listing pages with no item source and files `capability_gap` receipts
  (§6 query meaningful on the first plan run). **The r2 refinements ride the NEXT roll**
  (derived vocabularies, optional-error durable record, shared-writer receipts — and with them
  the 326 anti-churn deferral of receipts): until that roll, receipts insert directly, so the
  "missing receipt ≤~3h = anti-churn" line below does NOT yet apply. ⚠ **Proving a roll: do NOT
  use CLAUDE.md's `build provenance` log grep or a plain `grep -aq` on `/proc/1/exe` — both are
  LANDMINED (no such log line on backend services; BusyBox grep false-absences with controls
  passing).** The working instrument is the NUL-split probe with two controls in the same
  pipeline: `bugfix_444_empty_listing_pages/RUNBOOK_bugfix_444.md` "Prove the gate is LIVE".
  Once it is live: a DEFERRED listing section means the
  item source ERRORED — check the exporter config / kind opt-in, NOT the component. New
  deferred-section HITL rows where there was silence are genuine findings, not noise; nothing
  new can FAIL a build that previously completed. (Pre-repair, the same condition built HOLLOW
  and completed normally — 444's original symptom.) **And a MISSING gate receipt whose
  item_key recently hit a terminal row may be the shared anti-churn policy deferring it ~3h
  (bugs_open/326) — not gate failure; the findings row still lands.** A replan never drops a
  BUILT listing page (their round-1 preserve rule + test).

- ⚠ **A remake's TOOL URLs will serve 200 as PROSE SHELLS before its tools exist
  (`bugs_open/450`, seotools 7/7 on 2026-09-02).** The planner names a tool page it cannot yet
  fill (`hero-tool,generic-text-block`), the `owned_page_review` hold has no consumer, and the
  phantom-link repair (`unbuilt_internal_link`, one per LINK) builds the target through the
  generic builder — 2–6 rebuilds per page in 40 min. Tools arrive only when design-discovery's
  rotation selects the site (≈ one site per 3 h; advertise was reached by luck at plan time,
  which is why №1 never showed this). So: judge tools at the BODY (§4 form probe), keep the
  seven-style holds open, expect `dead_internal_link_live` to clear on its own (the shell 200s).
  **The shells OUTLIVE the tool wave** (seotools 2026-09-03: 8 real tools landed under other
  names; the 7 shells re-deployed at 00:0xZ and still serve prose). A planned tool page with NO
  plan sections takes the other fork: the repair parks at `mark_no_ready_sections` (HITL ×N
  links) and a `needs_content_page` "0 component rows — build it" follows (websitepromotion's
  `tool-channel-prioritiser`). Neither fork builds the planned tool.

## 2b. Mitigation for №5 onward — fire a one-shot design discovery when the PLAN completes (UNEXERCISED on a remake; №5's third canary duty)

Until 450's platform fix lands, make advertise's lucky ordering deliberate: tools before links.
The worked shape is `scheduled_tasks` row `oneshot-design-discovery-mcalc-20260814` (owner
decision 2026-08-14, one-shot not fleet re-enable; disable after first firing):

```sql
INSERT INTO scheduled_tasks (name, description, interval_seconds, target_agent_type, target_topic,
        input_data, max_concurrent, enabled, timeout_seconds, fire_message)
VALUES ('oneshot-design-discovery-<short>-<yyyymmdd>',
        'One-shot design discovery for <domain> at plan completion (bugs_open/450 mitigation, portfolio_positioning). Disable after first firing.',
        86400, 'design-discovery-agent', 'system.agent.scheduled.requests',
        jsonb_build_object('domain','<domain>','site_id','<site_id>'), 1, true, 900, true);
-- after it fires (last_triggered_at set): UPDATE scheduled_tasks SET enabled=false WHERE name='...';
```
- WHEN: the moment `site_plans.is_current` exists for the site and BEFORE any hub page deploys
  (pages build within ~1 h of release — the window is short; watch the plan item, not the clock).
- ~~⚠ Do NOT fire it at a site whose tool pages are ALREADY shells until 450 §7 is answered~~
  **450 §7 ANSWERED 2026-09-03: `tool-deployer` never touches a planned tool page — it creates
  its OWN page rows under the suggester's names, which never coincide with the planner's (0/7
  on seotools). So firing it is safe for existing shells (they are simply left standing) and
  the point of firing it EARLY is that the plan then NAMES the suggester's tools (advertise) —
  a plan written before the sweep names tools nobody will ever build.** After the sweep: the
  planner's own tool stubs are dead weight to retire (instance work, every remake until 450
  lands).
- ⚠ **The suggester is CLUSTER-blind as well as seat-blind (447, instance 2026-09-03):** it
  copied advertise's CPM/CPC comparator onto BOTH cluster siblings, and five more tools are
  duplicated pairwise across advertise/seotools/websitepromotion. The brief's flagship-deference
  clause lives in `site_specs`, which it does not read. Eye every `add_tool` wave against the
  cluster, not only against a P5 pair; the auditor holds nothing.
- ⚠ On a P5-pair site the wave it triggers is the seat-blind one (447) — eye it.
- Read the result at the body (§4), not at `evaluate_tools`/`add_tool` statuses.

## 3. Serve cutover (Cloudflare + Nominet) — ZONE FIRST

Full detail: `RUNBOOK_dns_pointing_a_domain_at_the_serving_worker.md` + its 2026-09-02
addendum. Zone → (confirm assigned NS pair) → Nominet NS → proxied apex A → worker route.
⚠ NS-before-zone = lame delegation, domain DARK including the old site (happened to all four).
⚠ The 404-token's empty zone list is NOT zone absence. Judge at the served body.
⚠ **A sitemap-refresh selection during the dark window publishes NOTHING and looks finished**:
  642's probe drops every URL (`url_count` 0, `probe_dropped` = the page count), the run reads
  COMPLETED, the stamp advances, and `/sitemap.xml` 404s until the change-and-quiet arm
  re-selects the site. Measured 2026-09-02: advertise 16:37:46Z 0/22, seotools 17:38:46Z 0/14
  (both dark then); websitepromotion 19:11Z and designblog 18:41Z were selected after the zones
  came up and serve theirs. Self-heals IF pages change after the stamp; the age arm alone is 3
  days — so a dark-window site still 404ing a day later needs "did any page change after the
  stamp?" answered before 642 is suspected.

## 4. Serve-verification (the build isn't done until this passes)

```bash
curl -s https://<domain>/ | grep -o "<title>[^<]*</title>"   # the NEW title, not Drupal's
# negative-identity guard: expect 0
curl -s https://<domain>/ | grep -ic "we don't sell\|we do not sell\|does not sell"
# LISTING PAGES: count the ITEMS, never the bytes (bugs_open/444) — 0 items on a
# directory/glossary/feed page is a finding even at 200/60KB of good prose.
# TOOL PAGES: a tool is a FORM, never a size. Probe every /tools/<slug>/ and CONTROL against
# a known-real tool page on a sibling site in the same run:
for t in <slug…>; do b=$(curl -s "https://<domain>/tools/$t/?cb=$RANDOM"); printf "%-30s %7dB forms=%d inputs=%d selects=%d\n" "$t" "${#b}" "$(grep -o '<form' <<<"$b"|wc -l)" "$(grep -o '<input' <<<"$b"|wc -l)" "$(grep -o '<select' <<<"$b"|wc -l)"; done
```
- ⚠ **The same `owned_page_review` text carries BOTH cases, and only the artefact tells them
  apart.** advertise (NOTES (e), 13:09Z): components already deployed, validator stale → close
  with evidence. seotools (21:1xZ, all 7): `/tools/<slug>/` served 200 at ~56 KB with the tool's
  H1 and **0 `<form>`, 0 `<input>`, 0 `<select>`, 1 `<button>` (the mobile-menu toggle)**;
  control = advertise's 3 real tools, same probe: 1 form, 0–11 inputs, up to 2 selects. So the
  seotools holds were TRUE — prose shells at the tool URLs. A cache-busted curl is the check,
  not the item's wording and not the page's size.

## 5. Remake №5 — the chrome-pin experiment, RE-SCOPED 2026-09-03 (do NOT run the old one-UPDATE form)

**A pin SELECTS a component; nothing POPULATES it.** Correction from the theme-kits lane
2026-09-03, then measured here and it is worse than their framing:

| | |
|---|---|
| Header components available to pin | `header-with-cart-or-nav` 11 vars · `header-with-search` 12 · `header-minimal-tool` 16 · `header-with-categories` 16 · `header-docs` 19 |
| Sites supplying ANY header `content_data` | **2 of 39. The other 37 supply ZERO keys** (the two supply 4 and 5) |

**INDEPENDENTLY REPLICATED** — the theme-kits lane re-derived every figure with separate queries
and got the same numbers (their `bead338cb`, which also retracts the claim at source; their §2
recommendation table is struck rather than deleted, because each field is right and only the SET
was incomplete). Two lanes, two query paths, one answer.

[MEASURED 2026-09-03, `content_components.html_template` `{{.var}}` distinct count ×
`site_components.content_data` key count where `slot_name='header'`.]

**So the estate's chrome sameness is a DATA problem, not a selection problem.** The default
`site-header` renders everywhere because it is the only header that needs (almost) no data. Pinning
any alternative on any of those 37 sites today renders 11–19 empty variables — an empty
`action="{{.search_action_url}}"` posts the form to the current page, aria-labels vanish, nav
disappears. **That is worse than the sameness.** This also revises the theme-kits CONTRIB's original
claim that selection was the bottleneck, and it re-sizes the differentiation programme: chrome
variety exists and has never been supplied, so "pin a nicer header" is not a per-site UPDATE.

**PRE-FLIGHT, mandatory, before any pin (theirs, adopted verbatim):**
```sql
-- what the target site supplies
SELECT (SELECT string_agg(k, ', ' ORDER BY k) FROM jsonb_object_keys(sic.content_data) k)
  FROM sites s JOIN site_components sic ON sic.site_id=s.id AND sic.slot_name='header'
 WHERE s.domain = '<TARGET>';
-- what the candidate requires
SELECT DISTINCT unnest(regexp_matches(html_template, '\{\{\.([a-zA-Z_]+)', 'g'))
  FROM content_components WHERE function='<CANDIDATE>' AND is_active;
```

**The experiment, in its only safe form** — at №5's release, AFTER composition has created
`site_components` and BEFORE the release rerender:
1. Run the pre-flight. Expect zero supplied keys.
2. **Supply the candidate's vocabulary first** (one guarded UPDATE of that site's header
   `content_data`), choosing a component whose vocabulary the site can honestly fill — for
   copyonline `header-minimal-tool`'s tool-shaped keys are fillable (it ships four tools);
   cart/docs are wrong for it. ⚠ **`header-with-search` carries a trap for LATER sites (theme
   kits, 2026-09-03): its form action is `{{.search_action_url}}` and the template references no
   handler and no `/search` — so supplying that variable COMMITS you to an endpoint that really
   exists. A populated variable pointing at a 404 is worse than an empty one, because it looks
   like it works.** ⚠ And do not fill a vocabulary the site cannot honestly answer:
   `tool_status_label` on a site with no tools gets invented text, which is a worse failure than
   an empty header.
3. Pin by UUID (function names are NOT unique):
   `UPDATE style_collections SET header_component_id = '<uuid>'::uuid WHERE id = (SELECT style_collection_id FROM sites WHERE domain='<TARGET>');`
4. Rerender, then diff the SERVED header against an unpinned sibling (cache-bust the curl).

**Three-way read, third branch corrected (theirs):**
- changed-and-right → pinning works AND the data was sufficient; the recipe is safe for the 17,
  but each still needs its own pre-flight and its own data.
- unchanged → the pin was IGNORED; chrome differentiation is a `ChromeSlotFunction` code change.
- **renders but empty/broken → the pin WAS honoured and the component's `content_data` is
  unsupplied.** A component-data problem, NOT a verdict on pinning — and the EXPECTED outcome for
  any site with an empty header `content_data`, which is 37 of 39. Check supplied keys against the
  template's variables before concluding anything about the mechanism.

Send the served diff to the `theme kits` session either way (feeds their 438 read).

## 5b. Remake №5 is ALSO the imagery-prompt canary (migration "718", planner prompt, live 19:59:56Z)

The designblog lane shipped content-carrying imagery expectations into build-site-planner's
plan_site prompt (owner-directed; corr 2dae4f20 Council-Submitted). On №5's PLAN, check
`imagery.sections`: expect ≥1 illustration/infographic under index, more where sections warrant.
Two named residual risks to eye: (a) entries must land on sections whose component can DISPLAY
them (instruction-only limit — an illustration on a plain prose block is the
generated-and-undisplayable class); (b) volume scales with section-scope entries (owner
accepted). Blog-post/guide/empty-section pages are EXEMPT — an imagery-free article page is not
the change failing. ⚠ TWO lanes have referenced "migration 718" tonight (this planner prompt +
a file in the 444 submission) — resolve by FILENAME, never bare number:
`718_planner_imagery_content_expected_prompt_and_exemplar.sql` is the live one described here.

## 6. Fire-direction template (v2, NOTES (q))

Per-domain vertical · best-in-vertical · no-omission · fullness · no negative-identity ·
LAYOUT intent naming a distinct layout ~~(prefer the nine unused)~~ **— ⚠ DO NOT "prefer the nine
unused": [MEASURED 2026-09-03] NINE of 18 layouts are reachable by ZERO sites' tags** (soft-editorial,
affiliate-hub, comparison-aggregator, docs-sidebar, industry-hub, media-grid, portfolio-kinetic,
tool-first-landing, utility-tool), because their `industry_tags` are a designer's INDUSTRY dialect
(`wellness, bakery, artisan`, `law`, `boxing`) that the classifier never emits. Only
`tool-portal-dark`/`-light` have vocabularies written in the classifier's own language, which is why
15 of 33 sites sit on two layouts. A brief cannot reach an unreachable layout by naming it; the
lever is the TAG vocabulary, which is 445's ground (NOTES (cc)) — ⚠ **an INFLUENCE, NOT a lever
(NOTES (z), 2026-09-03)**: the resolver scores `classification.category` + `industry_tags` only,
and takes just the light/dark SCHEME from `design_intent.style_direction`; it *"never consults
design_intent"* for the layout itself. A named layout reaches it only if the CLASSIFIER copies the
name into `industry_tags` — ~~which is how three remakes landed on magazine-grid~~ **REFUTED by the
445 lane 2026-09-03: that tag is INERT in the scorer (not in the layout's own tags, not in the
description-bonus strings, not the category bonus). The 12-site census stands as prompt hygiene, not
as a cause.** What actually happens: seven sites match `magazine-grid` on ONE tag
(`editorial-publication`, score 3.05, **7–10% coverage of their declared identity**), and four sites
are recorded `tags 0.00` with `layout_source: library_match`. So name the layout AND **verify at
composition**, capturing FOUR fields — the SCORE is what separates a real fit from a near-miss
(`tool-portal-light` at 8.31 on three tags ≈ 23% coverage vs `magazine-grid` at 3.05 on one ≈ 7%
are very different outcomes, even though both are "not the layout we asked for"):
```sql
SELECT data->>'layout_name'                 AS layout,
       data->'lineage'->>'layout_source'    AS source,
       data->'lineage'->'layout_candidates' AS candidates,
       data->>'reasoning'                   AS reasoning   -- carries the score
  FROM site_specs WHERE site_id = :id AND aspect='resolved_composition' AND is_current;
```
(also `sites.style_collection_id` → `style_collections.css_theme_id` → `css_themes.layout_id`).
⚠ **`layout_match_score` does NOT exist** — migration `103_site_design_planner.sql` specified it as
a normalised 0–1 tag-overlap score with a 0.5 threshold in April 2026 and it was never built
(0 of 33 rows carry the key), which is why coverage must be read out of the `reasoning` prose.
**Send all four to the `445` lane.** The only
hard lever is a theme kit · COLOUR referent = nearest
estate neighbour's actual served values · grep the plan for `contact-hero` ·
**444 pre-enablement, per the fixing thread's recipe (2026-09-02)** — fire direction needs NO
new field; instead, BEFORE firing a brief that wants listing pages:
- ⚠ **THREE OWNER RULINGS relayed by the designblog lane, 2026-09-03** — they change this list:
  (1) **`page_archetypes` APPLY** — the theme-kits gate is lifted, so the CONTRIB's "do NOT size on
  page_archetypes" is superseded; sequence the next briefs behind the theme-kits roll if it is
  close. (2) **GLOSSARY / INSPIRATION: hold them in briefs AND build a producer** — 444's candidate
  3 is owner-sanctioned and being built, and until it lives a direction must stop promising a
  glossary or inspiration SURFACE. A prose `guide` that writes its definitions inline is the
  compliant form (copyonline's brief does this; flagged to that lane). (3) **Feed-shaped pages stay
  SECTION-INDEX and fill from child pages** — no replan; say so in the direction.
- FEEDS: author `content_features.news_feed` in the classification spec (idea.uk 2026-08-25 is
  the worked example — the 6-hourly trigger then seeds `content_sources` itself), or seed
  sources directly.
- DIRECTORIES: DIR-001's seven-place kind checklist, or a `directory-json-exporter` config row
  (the vetcomparison pattern — that row is the filled-vs-hollow discriminator).
- GLOSSARY/SHOWCASE: no producer exists (444 candidate 3) — keep these page types out or
  explicitly conditional until one does.
Post-plan receipt query (opt-in on validate_site_plan, files capability_gap per dropped page —
**returns rows ONLY after a chassis roll carrying `6525b45ae`**; before that the prompt half
alone makes the planner decline the page, with no receipt. ⚠ So an EMPTY result is ambiguous
between gate-not-live and nothing-held until the stamp is read. The half you CAN read without
the stamp: `needs_section_data` rows whose summary carries "required query source errored:" —
that Reason string exists only in 444's defer branch (`dbb218a41`), so its presence proves the
DEFER half rolled; it does not prove the gate did. 444 session, 2026-09-02):
```sql
SELECT item_key, spec->>'builder_needed', summary FROM site_work_items
 WHERE site_id=:site AND item_type='capability_gap'
   AND spec->>'gap_kind'='producer_missing'
   AND status NOT IN ('complete','cancelled','rejected');
```
