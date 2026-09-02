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
- ⚠ Do NOT fire it at a site whose tool pages are ALREADY shells (seotools tonight) until
  450 §7 is answered — what `tool-deployer` does to a page row that already carries generic
  components is UNVERIFIED (on advertise it created the rows itself).
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

## 5. Remake №5 ONLY — the chrome-pin experiment (theme kits, 2026-09-02)

After site row + style_collection exist, BEFORE the release rerender. Header only — one
variable. Pin `header-minimal-tool` by UUID (function names are NOT unique):

```sql
UPDATE style_collections
   SET header_component_id = 'b519d5d3-1af6-4c91-a1c3-ae81457369ee'::uuid
 WHERE id = (SELECT style_collection_id FROM sites WHERE domain = '<REMAKE_5_DOMAIN>');
```
Post-rerender, diff served header classes vs an unpinned sibling (cache-bust the curl).
**Read the result THREE ways** (the alternatives are `*_pre_037` legacy rows never rendered
live): changed-and-right → pins work, recipe safe for the 17 · unchanged → pin IGNORED,
chrome differentiation is a ChromeSlotFunction code change · broken/missing → component
stale, NOT a pin verdict — retry with another component before concluding. Send the served
diff to the `theme kits` session either way (feeds their 438 read).

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
LAYOUT intent naming a distinct layout (prefer the nine unused) · COLOUR referent = nearest
estate neighbour's actual served values · grep the plan for `contact-hero` ·
**444 pre-enablement, per the fixing thread's recipe (2026-09-02)** — fire direction needs NO
new field; instead, BEFORE firing a brief that wants listing pages:
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
