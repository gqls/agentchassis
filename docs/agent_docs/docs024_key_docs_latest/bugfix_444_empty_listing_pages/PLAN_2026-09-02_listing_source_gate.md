# PLAN — bugs_open/444 class fix: the listing-page item-source gate (2026-09-02)

Owner: session "bugs_open/444". Bug: `bugs_open/444_HANDOFF_2026-09-02_…`. Evidence and
diagnosis: `NOTES_bugfix_444.md` (same dir) + the fixing-thread findings block in the bug
file. This plan prefers the framework-level door-closer over per-site patches, per the
owner's standing direction and `order_fix_candidates_by_what_closes_the_door`.

## What the fix must guarantee

A plan that reaches persistence contains no listing page whose item source resolves to
zero for THIS site — such a page becomes a `capability_gap` work item naming the missing
producer, never a built page of meta-prose. And a section whose REQUIRED query source
ERRORS at build time is deferred loudly (HITL), never shipped hollow.

## Why plan time, not render time (the load-bearing finding)

The render layer cannot close the class:
- `news-listing.items` is `required: false` BY DESIGN — empty is legal between feed runs
  on a site with sources (client JSON refresh is the freshness path, bugs_open/027).
- `directory-listing.entries` HAS a full contract (required, min_items 1, skip_section)
  and it still shipped hollow: `resolveBusinessDirectory` errors on missing config
  (bugs_open/206, deliberate) and `plan_sections`' error branch deliberately bypasses
  `on_missing` (bugs_open/054, deliberate). Two correct guards in series produce the
  silent hollow section both were built to prevent. Loosening either reopens the bug it
  fixed.

## The changes (Phase A — one coherent task, one council round)

1. **`platform/orchestration/actions/listing_item_sources.go` (new).** One registry:
   for a planned page, is it LISTING-FAMILY (role via `datahelpers.isSectionIndexRole`
   vocabulary: section-index / blog-index / news-index / entity-directory; plus the six
   directory page_types from `directoryCheckProfiles`; plus any section naming a
   listing-family component), and does its item source resolve for this site?
   Resolvers, each returning (ok, producerNeeded, evidence):
   - `section-index` / `blog-index`: ≥1 child page in the SAME plan (parent_section /
     blog-post typed) or already-realised child pages in `pages`. Plan-internal, so the
     tools-arrive-later ordering (311 residue) cannot false-positive: a tools hub planned
     WITH its tool pages resolves even though tool-deployer builds them later.
   - `news-index` (or `news-listing` section): current classification spec has
     `content_features.news_feed.recommended = true` (the NEWS-001 driver — sources are
     then seeded by the existing trigger) OR ≥1 active `content_sources` row.
   - directory pages/sections: per-kind components → `content_features.<SpecKey>`
     recommended (DIR-001 opt-in); bare `directory-listing` → the SAME config lookup
     `resolveBusinessDirectory` uses (exported as a thin has-config helper) so the gate
     and the renderer can never disagree.
   - a listing-family page whose sections carry NO resolvable listing source → not
     producible.
2. **`ValidateSitePlanAction`** (`v3_site_actions.go`, after the section-name resolution
   block): behind a NEW optional config key `enforce_listing_sources` (default **false**
   = today's behaviour), each unproducible listing page is REMOVED from the plan,
   recorded through the existing dropped/findings machinery (`LogActionFindings`), and
   filed as a `capability_gap` work item — existing carrier, reconcile's key shape
   `capability_gap:<page_type>:<page_name>` for co-dedup, `handler_agent=''` (the
   078/291 livelock rule), `spec.gap_kind='producer_missing'`,
   `spec.builder_needed=<producer slug>` (the field `diagnose_triage` groups on),
   plus page/site context. No new item_type (the minting ratchet forbids it; the
   consumer exists).
3. **`plan_sections_action.go` secondary repair:** in the query-error branch, when the
   field is REQUIRED (or declares min_items ≥ 1): DEFER the section (the existing loud
   HITL path) instead of leaving it ready. Errors stay distinguishable from no-data
   (054's rule preserved — on_missing still not applied; defer is the third state), and
   206's "loud failure" finally lands somewhere durable instead of one Warn line in a
   restarting pod. Known extra deferral volume: bounded by resolver error rate; the only
   systematic producer today is the misconfigured-directory case this bug is about.
4. **Migration `docs/agent_docs/sql_for_agents/<next>_planner_listing_source_gate.sql`:**
   (a) set `enforce_listing_sources: true` on build-site-planner's `validate_plan` step;
   (b) narrow the prompt sentence "Plan the IDEAL site regardless — the build system
   handles which pages can be built now vs later" (053) — it currently licenses planning
   listing pages nothing can fill, and contradicts 433's own "a page for a kind the site
   has not opted into ships empty"; (c) add the glossary/showcase conditional rule
   (prompt-level: plan them only when a producer exists — stated as guidance, since no
   mechanical gate can see a glossary typed `content`; that half of 444 remains with the
   copy lane's title-promise design, split recorded in their NOTES).
5. **Tests:** registry resolvers (table-driven); validate-action gate (drop + gap row,
   kept when source resolves, OFF by default); plan_sections error→defer (mutation-proof:
   assert the section was previously ready under the same fixture). Run
   `optional_budget_cron_parity_test.go` + `scripts/audit-optional-key-budget.sh` after
   adding the config key (WFA-013).
6. **Concept register entry, SAME COMMIT as the seam** (2026-07-29 ruling condition 2):
   the gate, its config key, the producer_missing vocabulary, and the landmine (a gap
   row's builder_needed slugs are the enablement checklist's machine half).

## Explicitly out of scope (owned elsewhere / later)

- Per-site enablement of the five standing instances: designblog session (their two),
  feed lane (WebProNews → advertise news), portfolio_positioning (advertise/seotools
  directory decisions: kind vs business_directory vs drop the page).
- A glossary/showcase item producer (bug candidate 3) — new build, needs an owner.
- `ReconcileSitePlanAction` backstop arm (same registry, for pages arriving outside the
  planner) — Phase B if the council or field evidence asks for it; plan_sections'
  error→defer already covers the render path for those pages.
- Retro-cleanup of the five shipped pages — instance work, follows enablement.

## The enablement contract (what "444 enablement answered" means for brief-firing)

After this ships: fire the brief; the validator holds back unfillable listing pages and
files `capability_gap` receipts. Post-plan checklist query (goes in
`portfolio_positioning/RUNBOOK_remake_release.md` §6):

```sql
SELECT item_key, spec->>'builder_needed' AS needs, summary
FROM site_work_items
WHERE site_id = :site AND item_type='capability_gap'
  AND spec->>'gap_kind'='producer_missing'
  AND status NOT IN ('complete','cancelled','rejected');
```

To get the pages PLANNED rather than held: pre-enable before firing — feeds: author
`content_features.news_feed` in the classification spec (idea.uk 2026-08-25 is the worked
example) and/or seed `content_sources`; directories: DIR-001's seven-place kind checklist
or a `directory-json-exporter` config row (vetcomparison pattern).

## Council / review posture

- `enforce_listing_sources` is opt-in, unsafe-default-OFF, zero live consumers until the
  migration names it → NOT architecture-RFC scope per RFC_022 (all three conditions
  hold; the consumer enumeration is the one migration in this very plan).
- Changes 1–3 + the migration are council-gate scope (platform/ + sql_for_agents).
  One submission for the coherent task; `Council-Submitted:` trailer if committing
  before the verdict.
- Consumers told, not merely measured (2026-07-29 §3): portfolio_positioning (messaged,
  corresponding), designblog.co.uk, feed lane, copy_quality_two_stage (their title-promise
  scope depends on this landing), site design planner lane (planner prompt change).
