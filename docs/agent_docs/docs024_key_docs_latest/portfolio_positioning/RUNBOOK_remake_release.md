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
- ⚠ Once 444's repair rolls (council corr `c0990eb3`, submitted 2026-09-02 late): a DEFERRED
  listing section means the item source ERRORED — check the exporter config / kind opt-in,
  NOT the component. New deferred-section HITL rows where there was silence are genuine
  findings, not noise; nothing new can FAIL a build that previously completed. (Pre-repair,
  the same condition built HOLLOW and completed normally — 444's original symptom.)

## 3. Serve cutover (Cloudflare + Nominet) — ZONE FIRST

Full detail: `RUNBOOK_dns_pointing_a_domain_at_the_serving_worker.md` + its 2026-09-02
addendum. Zone → (confirm assigned NS pair) → Nominet NS → proxied apex A → worker route.
⚠ NS-before-zone = lame delegation, domain DARK including the old site (happened to all four).
⚠ The 404-token's empty zone list is NOT zone absence. Judge at the served body.

## 4. Serve-verification (the build isn't done until this passes)

```bash
curl -s https://<domain>/ | grep -o "<title>[^<]*</title>"   # the NEW title, not Drupal's
# negative-identity guard: expect 0
curl -s https://<domain>/ | grep -ic "we don't sell\|we do not sell\|does not sell"
# LISTING PAGES: count the ITEMS, never the bytes (bugs_open/444) — 0 items on a
# directory/glossary/feed page is a finding even at 200/60KB of good prose.
```

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
Post-plan receipt query (once their validator lands — opt-in on validate_site_plan, files
capability_gap per dropped page):
```sql
SELECT item_key, spec->>'builder_needed', summary FROM site_work_items
 WHERE site_id=:site AND item_type='capability_gap'
   AND spec->>'gap_kind'='producer_missing'
   AND status NOT IN ('complete','cancelled','rejected');
```
