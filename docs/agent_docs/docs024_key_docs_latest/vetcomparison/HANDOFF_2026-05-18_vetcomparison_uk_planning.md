# HANDOFF — vetcomparison.uk planning & schema unification

**Date:** 2026-05-18
**Status:** Five artefacts written, two operations blocked on Go work, schema unification decided. Nothing applied to production yet.
**Continuation goal:** apply the artefacts in order, ship the paired Go changes, take vetcomparison.uk live as v1 (online-only medicine search + vet directory + news).

---

## TL;DR

We're recreating `vetcomparison.uk` on the existing agent-chassis platform. The site was broken because the prototype data was wrong and the `med-json-exporter` was targeting `vetcomparison.co.uk` (which the user does not own). Replacing it is acceptable.

**V1 scope** is deliberately small: online-only medicine search (top 3-5 prices per medicine), vet directory with filter/sort, news feed (RSS + keyword search + Grok), and two adopted guide pages. **No local-pharmacy comparison panel** until per-vet medicine prices accumulate.

Five artefacts written. None applied. The original migration 001 was scrapped after discovering `business_intel.products` + `product_prices` already exist and cover the use case — replaced by migration 006 which unifies the previously-split `business_prices` (services) and `product_prices` (products) under a single polymorphic schema rooted on a new `kind` column on `products`. This is a structural fix the user explicitly asked for.

Two Go changes are blocking. Three DB queries are still owed. The runbook needs a small re-order. Everything else is in place.

---

## Files produced (all in `/mnt/user-data/outputs/` if available, or attach as needed)

| File | Status | Purpose |
|---|---|---|
| `vetcomparison_uk_plan.md` | **Stale — references the scrapped 001 and the old news-decision-in-Go plan.** Either rewrite or read in conjunction with this handoff. | Comprehensive plan doc from the first half of the session. |
| `001_business_med_prices_migration.sql` | **SCRAPPED. Do not apply.** | Would have created a third sibling price table; unnecessary now. |
| `002_vet_practice_verifier_prompt_patch.sql` | Apply when ready. Paired Go change required (see below). | Extends the vet-practice-verifier LLM prompt to extract `medicine_prices[]` alongside service prices. Sections renumber: medicine_prices=5, confidence_score=6, extraction_notes=7. |
| `003_vet_json_exporter_agents.sql` | Apply when ready. Paired Go change required (see below). | Two new agent_definitions: `vet-json-exporter` (specialist, single `vet_export_json` action) and `vet-export-orchestrator` (wrapper, spawn → call → complete per doc 001). |
| `004_vetcomparison_uk_runbook.md` | **Stale on three points** — see "Runbook corrections" section below. | Step-by-step operator runbook with kcat triggers + SQL. |
| `005_classifier_content_features_prompt_patch.sql` | Apply ANY TIME. Self-contained, no paired Go change required. | Moves the news_feed/tools/guides decision out of the Go `verticalNewsMap` and into the `domain-research-classifier` LLM prompt. Operation A patches the classifier prompt; Operation B removes the now-redundant `enrich_news_feed` step from `improvement-loop`. |
| `006_unify_prices_schema.sql` | Apply AFTER Go-B ships (or before, then re-run; it's idempotent). | Adds `kind` column to `business_intel.products`. Migrates `business_prices` rows into `(products kind=service) + product_prices`. Deprecates `business_prices` via comment (does not drop). |

If picking up cold, **read this document and `vetcomparison_uk_plan.md` together** — the plan doc has the architecture overview and reference paths, this document has the current-state corrections.

---

## What changed this session vs the original plan doc

Three structural shifts that the plan doc does not reflect:

### 1. The original `business_med_prices` table was scrapped

The user ran `\dt business_intel.*` and confirmed `business_med_prices` doesn't exist, but pointed out that `business_intel.products` and `business_intel.product_prices` already do. Schema check confirmed they cover the case:

```
products(id, vertical_id, slug, name, dosage, category,
         typical_vet_price_gbp, requires_prescription, species,
         active_ingredient, manufacturer, is_active, ...)

product_prices(id, product_id FK, business_id FK,
               price_gbp, price_qualifier, includes_vat,
               in_stock, product_url, source,
               observed_at, is_current)
```

The partial index `idx_bi_prodprices_current ON (product_id, business_id) WHERE is_current = true` is exactly what we'd have built. The `typical_vet_price_gbp` column on `products` was clearly designed for benchmarking against online prices. **Reuse-before-create applies; migration 001 deleted.**

### 2. The two price tables get unified instead of leaving the split

`business_prices` (service_category + service_name + business_id) and `product_prices` (product_id + business_id) were two separate tables with different identifier shapes. User flagged this and asked to fix it structurally.

Migration 006 adds `kind` to `products` (default `medicine`, also allows `service`, `supplement`, `other`), inserts a `products` row per distinct service offering with `kind='service'`, migrates `business_prices` rows into `product_prices`, deprecates `business_prices` via table comment. Idempotent.

Slug strategy:
- **Services** → `service-{lowercased,hyphenated category-name}` (e.g. `service-consultation-15-minute-consultation`). Same category+name from any business → same slug → same products row.
- **Medicines (handled in paired Go)** → `{name}-{dosage}-{size_variant}` (e.g. `apoquel-3-6mg-20-tablets`). Some fragmentation expected initially; a future product-matcher action can consolidate.

### 3. The news/tools/guides decision moves from Go to the LLM

Original plan had a Go change (`Go-C`) to rebalance `verticalNewsMap` so most verticals default to news-recommended. User overrode that: move the decision entirely into the LLM via the classifier prompt.

Migration 005 does this. The `domain-research-classifier` prompt (`e6ca8cca-398b-49f9-a435-337b1cc26c38`) now produces a `content_features` block nested inside the `classification` JSON output. The block contains `news_feed` (full schema as previously consumed by `EvaluateNewsFeedAction`), `tools` (with examples), and `guides` (with topics). Infographics dropped — no production mechanism yet.

The new prompt guidance says: **recommend by default for all three**, except for single-location service businesses where brand voice would suffer. Aggregators, comparison sites, hubs, directories all get everything recommended. This matches the user's policy: "tools, guides, news, infographics on most sites because they are good for long-term user usefulness and therefore traffic."

Operation B of migration 005 surgically removes the now-redundant `enrich_news_feed` step from `improvement-loop`. The Go `EvaluateNewsFeedAction` and `verticalNewsMap` remain in code as orphaned but functional — slated for deletion in a follow-up PR after the new flow proves stable.

---

## Runbook corrections (004 is stale on these)

When updating the runbook, three changes are needed:

**Correction A: apply 005 BEFORE adoption.** Current runbook applies 001 → 002 → 003 → then adopts. New order should be: apply 005 → apply 002, 003, 006 → ship Go-A + Go-B → adopt. Reasoning: with 005 in place, the classifier writes `content_features.news_feed.recommended=true` automatically during adoption. The manual spec-patch (current step 3) becomes a fallback rather than the primary path.

**Correction B: scrap migration 001 from step 1, add migration 006.** The runbook step 1 should read:
```
psql $CLIENTS_DB -f 005_classifier_content_features_prompt_patch.sql
psql $CLIENTS_DB -f 002_vet_practice_verifier_prompt_patch.sql
psql $CLIENTS_DB -f 003_vet_json_exporter_agents.sql
# 006 applied AFTER Go-B ships (or before; it's idempotent)
psql $CLIENTS_DB -f 006_unify_prices_schema.sql
```

**Correction C: remove "step 0c" (verticalNewsMap rebalance) and "step 10" (verticalNewsMap follow-up).** Those described the Go map rebalance which is now replaced by 005's LLM-driven approach. The follow-up Go work shrinks to "delete `EvaluateNewsFeedAction` and `verticalNewsMap`" — pure dead-code cleanup.

The RSS seeding step (current step 4) stays manual per user's choice (Option A). The content_features bootstrap waits for the improvement-sweep cron per user's choice (Option E). Both decisions stand.

---

## Go code still required before the runbook is fully executable

Three pieces of paired Go work. **None are written in this session.** They are listed in priority order for the next person picking this up.

### Go-A — `vet_export_json` action

A new action handler modelled on `MedExportJSONAction` (`platform/orchestration/actions/med_export_json_action.go`). Referenced by the `vet-json-exporter` workflow from migration 003.

Behaviour:
- Read `params.StepConfig.Config`, merge `input_data` overrides on top.
- Query `business_intel.businesses` joined with `vet_practice_details`. Apply filters: `min_confidence`, `require_website`, `require_postcode`.
- For each business, optionally include service prices grouped by category. **Read from `product_prices` joined to `products` filtered on `kind='service'` — NOT from the deprecated `business_prices`.** This is the critical change vs the original plan because of the schema unification.
- Build the JSON shape consumed by the prototype's `vet-search.js` (one full-index file at minimum, optional metadata file).
- Send to git-adapter the same way `sendMedExportToGit` does.
- Update `scheduled_tasks.last_completed_at` for `task_name`.
- Register in `registry.go` as `"vet_export_json"` with `Category: "business_intel"`, `IsLocal: true`.

**BLOCKER:** the user owes query 4 output — distinct `service_name` strings per `service_category`. Cannot safely write the consult/Rx mapping switch without this. See "Owed DB queries" below.

### Go-B — `store_business_verification` extension AND `insertPrice` rewrite

Two related changes in the same Go area:

1. **Rewrite `insertPrice`** (currently writes to `business_prices`) to upsert into `products` + `product_prices` with `kind='service'`. Logic:
   - Compute slug from `service_category` + `service_name` using the same regexp as migration 006.
   - `INSERT INTO products (slug, name, category, kind='service', ...) ON CONFLICT (slug) DO UPDATE SET updated_at=NOW() RETURNING id`.
   - Mark prior `is_current=true` rows for `(business_id, product_id)` as `is_current=false`.
   - INSERT the new `product_prices` row.

2. **Add a paired medicine helper** that reads `verification_result.medicine_prices[]` from the LLM output, computes slugs from `(product_name, dosage, size_variant)`, upserts into `products` with `kind='medicine'`, and INSERTs into `product_prices`. Same upsert dance.

3. **Skip writes to the deprecated `business_prices` table.** Once Go-B is deployed, no new rows should land there.

After Go-B deploys, apply migration 006 to backfill historical `business_prices` rows. The migration is idempotent so re-running is safe.

### Go-C — delete dead code (follow-up only, after 005 proves stable)

After 005 has been in production for a couple of weeks and the LLM-produced `content_features` is reliably driving the news feed:

- Delete the action registration for `"evaluate_news_feed"` in `registry.go`.
- Delete the file `feed_news_recommendation_action.go` (or just the action handler and `verticalNewsMap`).
- The `MissingNewsSourcesCheck`, `MissingNewsPageCheck`, and `SeedContentSourcesAction` continue to work unchanged — they read `content_features` from the classification spec data, which is now written by the LLM instead of by `EvaluateNewsFeedAction`. Shape is identical.

---

## Owed DB queries (unblock Go-A and confirm row counts)

User did not have DB access for some of the planning session and ran some checks (the `\dt` and `\d` outputs that drove the schema unification). These are still owed:

```sql
-- 4. CRITICAL FOR GO-A: distinct service_name strings per service_category
--    Drives the consult/Rx/vaccination mapping switch in vet_export_json.
SELECT service_category, service_name, COUNT(*) AS rows
FROM business_intel.business_prices
WHERE is_current = true
GROUP BY service_category, service_name
ORDER BY service_category, rows DESC;

-- 7a. Full \d output for the verifier's write targets
\d business_intel.businesses
\d business_intel.vet_practice_details

-- Row counts to size the migration's scope
SELECT COUNT(*) FROM business_intel.business_prices;
SELECT COUNT(*) FROM business_intel.business_prices WHERE is_current = true;
SELECT COUNT(*) FROM business_intel.products;
SELECT COUNT(*) FROM business_intel.product_prices;

-- Publishable vet counts (drives JSON export expectations)
SELECT
    COUNT(*) FILTER (WHERE verification_status = 'verified') AS verified,
    COUNT(*) FILTER (WHERE verification_status = 'verified'
                       AND confidence_score >= 0.5) AS verified_conf_05,
    COUNT(*) FILTER (WHERE verification_status = 'verified'
                       AND confidence_score >= 0.5
                       AND website_url IS NOT NULL) AS with_website,
    COUNT(*) FILTER (WHERE verification_status = 'verified'
                       AND confidence_score >= 0.5
                       AND website_url IS NOT NULL
                       AND postcode IS NOT NULL) AS publishable
FROM business_intel.businesses
WHERE business_type = 'veterinary_practice';

-- Site row state (will exist only after adoption)
SELECT id, domain, github_repo, name, created_at
FROM sites
WHERE domain IN ('vetcomparison.uk', 'vetcomparison.co.uk');
```

---

## Recommended order of operations from this point

This is the operational sequence for taking vetcomparison.uk live as v1, assuming a fresh start with all five artefacts in hand.

**Phase 1 — apply structural migrations and prompt patches (no Go required)**

1. Apply `005_classifier_content_features_prompt_patch.sql`. No paired code. Self-contained. Verify via the SELECT at end of file. After this, all newly-classified sites get LLM-produced `content_features`.
2. Apply `002_vet_practice_verifier_prompt_patch.sql`. Paired Go-B is required for the new medicine_prices output to land in the DB — but the prompt change is safe to apply first. The LLM will start producing the field, but it'll be discarded by current Go until Go-B ships. Acceptable.
3. Apply `003_vet_json_exporter_agents.sql`. Creates the two new agent_definitions. They cannot do anything until Go-A ships, but the definitions are safe to land.

**Phase 2 — ship Go-B and apply schema unification**

4. Implement Go-B per the spec in "Go code still required" above. Test against staging. Confirm `insertPrice` writes to `products + product_prices`, not `business_prices`. Confirm medicine helper writes equivalent rows.
5. Deploy chassis with Go-B.
6. Apply `006_unify_prices_schema.sql`. Backfills historical `business_prices` rows. Verify row counts via the SELECT at the end. If they look wrong (e.g. fewer service products than expected because of slug collisions), investigate before proceeding.

**Phase 3 — ship Go-A**

7. Get query 4 output from the user. Use the actual distinct `service_name` values to build the consult/Rx/vaccination mapping switch in `vet_export_json`.
8. Implement Go-A. Test against staging. Confirm JSON output matches the shape `vet-search.js` consumes.
9. Deploy chassis with Go-A.

**Phase 4 — adopt vetcomparison.uk**

10. Trigger `site-adoption-orchestrator` per runbook step 2 (existing kcat command pattern). With 005 in place, the classifier will produce `content_features.news_feed.recommended=true` directly. Wait 3-6 minutes for adoption + classification to complete.
11. Verify the spec was written correctly:
    ```sql
    SELECT data #> '{content_features,news_feed}' AS news_feed_config
    FROM site_specs
    WHERE site_id = '<new vetcomparison.uk id>'
      AND aspect = 'classification' AND is_current = true;
    ```
    Expect `recommended:true` and 4-12 keywords appropriate for the aggregator context.
12. If the LLM wrote wrong values (it has the adoption context so it shouldn't, but if), use the runbook step 3 patch as a one-off fallback for this site.

**Phase 5 — manual RSS seeding and exporter retargeting**

13. INSERT the three RSS rows for `(Vet Times, BVA Vet Record, RCVS)` per runbook step 4. Verify RCVS feed URL before committing (the pattern in the runbook is a guess).
14. UPDATE the `med-export-json` scheduled task to override the hardcoded `.co.uk` default — runbook step 5.
15. Create the `vetcomparison-uk-exports` concurrency group — runbook step 6.

**Phase 6 — first exports and build**

16. Smoke-test the vet exporter via kcat trigger per runbook step 7.
17. Enable the `vet-export-json` scheduled task once smoke test passes.
18. Trigger the first build cycle via the dispatch loop (or wait for it to pick up the `needs_*` work items created by adoption).

**Phase 7 — tool deployment**

19. Build the online-medicine-search tool component (slim variant of prototype's `calc.js` — no local-vet panel). Add to library, fork onto vetcomparison.uk via `deploy_tool_to_site`.
20. Build the vet-directory tool component (lift from prototype's `vet-search.js`). Same library + fork pattern.

**Phase 8 — wait for traffic + data**

The improvement-sweep cron will trigger content-feed-orchestrator on its schedule (default 6h), which runs `seed_content_sources` to auto-create the `news_search` (one per keyword) and `api_news` (Grok) source rows. Within ~6 hours of step 13 the news feed should have ingested first items.

**Phase 9 (later) — V1.5 local-pharmacy comparison**

Once `product_prices` with `kind='medicine'` has accumulated enough data (heuristic: 30+ practices in at least one postcode area with 5+ distinct medicines each), extend the online-medicine-search tool to add a "near you" panel. The data path and exporter already support it — only the tool UI extends.

---

## Critical things to be careful of

A few things easy to get wrong if picking up cold.

**The verifier prompt change in 002 only matters if Go-B ships.** The LLM will produce `medicine_prices[]` after 002 lands, but the current `store_business_verification` doesn't know to read that field. Until Go-B is deployed, the data is generated and discarded. This is acceptable as a phased rollout but the operator needs to know.

**Migration 006 idempotency depends on Go-B being deployed.** If Go-B is not yet deployed and the verifier keeps writing to `business_prices`, re-running 006 will migrate the new rows. Safe but wasteful. The cleanest order is Go-B first, then 006 once.

**`vet-search.js` and `calc.js` from the prototype both fetch JSON paths that don't yet exist.** The prototype `calc.js` loads `/data/medicine-index.json` (exists — produced by med-json-exporter) AND `/data/vets_by_postcode/AB.json` (does NOT exist, was dummy data). The v1 scope removed the local-vet panel so the second fetch goes away. The `vet-search.js` consumes `/data/vet-full-index.json` (produced by the new `vet-json-exporter` once Go-A ships). Tool components when forked onto the site need their JS revised to match v1 scope.

**The `med-json-exporter` default domain.** Hardcoded to `vetcomparison.co.uk` in `parseMedExportConfig`. Override via the scheduled task `input_data.domain` per runbook step 5. The Go default should also flip in a follow-up cleanup PR, low priority.

**`content_features` is nested inside `classification` in the LLM output JSON.** Don't expect it as a fifth top-level section. The classification spec data shape is unchanged for consumers — `content_features` sits alongside `site_type`, `category`, etc. at the top of the spec data.

**The Adoption fidelity guidance in the classifier prompt.** The classifier already has detailed guidance about following adopted specs as ground truth. For vetcomparison.uk, the adoption agent will capture the prototype's content (medicine comparison, vet directory, CMA guides) and the classifier will use those as authoritative signals. This is why we expect `content_features.news_feed.recommended=true` to come out automatically — the LLM sees the adopted archetype as an aggregator and applies the recommend-by-default policy.

---

## User preferences and working style (for whoever picks this up)

Captured from the conversation history. Following these will save back-and-forth:

- **Prose-y, light formatting.** Avoid heavy bullet lists. Tables sparingly.
- **No congratulating.** Don't say "perfect", "excellent", "great choice". Pragmatic only.
- **Schema-first.** Always check the schema before writing SQL. The session caught the products/product_prices reuse opportunity only because of this.
- **Reuse before create.** Look for existing structures before building new ones.
- **Every agent is an orchestrator.** Wrapper pattern (spawn → call → complete) for anything reached via the generic entry point.
- **Complex Go in actions, simple workflows.** Workflow JSON should be readable.
- **Variable names stay in sync.** If a workflow step says `input_data.foo`, the action reads `foo`. Don't rename without flagging.
- **Step-sized chunks.** Don't try to do everything in one go.
- **`logger.Info`, not `logger.Debug`.** Debug doesn't appear in K8s logs.
- **Kubernetes namespace.** `-n ai-persona-system` for chassis pods. `-n kafka` for kafka cluster. Cluster names like `personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092`.
- **One question at a time when asking.** Mutually exclusive options.
- **Don't say "final fix" or "last fix".** Don't promise it's done.

---

## Reference paths

### Artefact files
- `/mnt/user-data/outputs/vetcomparison_uk_plan.md` (stale on three points, see Runbook corrections)
- `/mnt/user-data/outputs/001_business_med_prices_migration.sql` (SCRAPPED)
- `/mnt/user-data/outputs/002_vet_practice_verifier_prompt_patch.sql`
- `/mnt/user-data/outputs/003_vet_json_exporter_agents.sql`
- `/mnt/user-data/outputs/004_vetcomparison_uk_runbook.md` (stale, see corrections)
- `/mnt/user-data/outputs/005_classifier_content_features_prompt_patch.sql`
- `/mnt/user-data/outputs/006_unify_prices_schema.sql`

### Prototype uploads (read-only reference)
- `/mnt/user-data/uploads/index.html`, `style.css`, `shared.js` — site chrome
- `/mnt/user-data/uploads/calc.js` — medicine calculator (v1 ships slim variant without local panel)
- `/mnt/user-data/uploads/vet-search.js` — directory search (lift to tool component)
- `/mnt/user-data/uploads/A.json`, `AB.json`, `medicine-index.json` — sample data shapes
- `/mnt/user-data/uploads/cma-compliance.html`, `cma-market-investigation.html`, `independent-strategy.html` — guide pages adopted during build

### Platform docs (project knowledge)
- `001_development_guide` — wrapper-orchestrator pattern, agent definition conventions, the `?` marker for optional input_mapping fields, fine-tuning roadmap
- `002_system_architecture` — agent inventory, layer model, dispatch loop
- `003_contracts_and_standards` — agent definition required fields, handler input_data paths
- `006_news_feed_pipeline_v2` — content-feed-orchestrator architecture, source seeding logic, `EvaluateNewsFeedAction` (which 005 supersedes)
- `007_adoption_pipeline_v4` — adoption flow, kcat trigger pattern
- `008_vet_med_pricing_pipeline` — med pricing schema (med_products, med_price_snapshots, etc.) and exporters

### Existing agent IDs referenced
- `d35a5612-c961-4b03-8f1d-e9e45801f874` — `vet-practice-verifier` (target of 002)
- `ea5f6fac-eb7d-4087-831b-82240bf022b7` — `med-json-exporter` (shape model for vet-json-exporter)
- `9953a150-4009-4dab-bb20-36ddd72cc3e5` — `med-export-orchestrator` (shape model, with corrected input_mapping)
- `e6ca8cca-398b-49f9-a435-337b1cc26c38` — `domain-research-classifier` (target of 005)
- `5dbe100a-47d8-4e72-8b21-867025d2f119` — `improvement-loop` (Operation B of 005 removes `enrich_news_feed`)
- `c9fa5603-874c-472e-a45a-d19739404494` — `content-feed-orchestrator` (downstream consumer of `content_features.news_feed`)
- `c0756913-04b1-489d-86b4-9ec249dc804d` — `tool-suggester` (could consume `content_features.tools` in future)

### Database tables touched or referenced
- `business_intel.businesses` — vet practice records
- `business_intel.vet_practice_details` — practice-specific details
- `business_intel.business_prices` — service prices, **DEPRECATED after 006**
- `business_intel.products` — canonical catalog of vet offerings, gets `kind` column from 006
- `business_intel.product_prices` — canonical price observations, becomes the unified prices table
- `business_intel.med_products`, `med_retailer_listings`, `med_price_snapshots`, `med_price_current` — online-retailer medicine pipeline, unchanged
- `business_intel.companies_house_data` — CH enrichment, unchanged
- `agent_definitions` — patched by 002, 003, 005
- `site_specs` — produced by classifier; consumers of `content_features` read here
- `content_sources` — news source rows; auto-seeded by `seed_content_sources`, RSS rows added manually
- `scheduled_tasks` — `med-export-json` retargeted via input_data override
- `sites` — created by adoption
- `site_work_items` — dispatch loop queue, drives the build cycle

---

## What a fresh session should do first

1. Read this document end-to-end.
2. Read `vetcomparison_uk_plan.md` for the mission/scope/architecture overview (skip the parts marked stale).
3. Confirm with the user: are we proceeding with the apply-then-Go-then-adopt sequence above? Any decisions to revisit?
4. Ask for the four owed DB queries if they haven't been provided.
5. Pick the next concrete step based on what's been completed:
   - If nothing applied yet → start with 005 (safest, self-contained, immediate effect on future classifications).
   - If 005 applied → write Go-B (the unified `insertPrice` + medicine helper).
   - If Go-B deployed → apply 006 and verify row counts.
   - If 006 applied → write Go-A using query 4 output.
   - If Go-A deployed → trigger adoption.

Don't try to do multiple Go pieces in one batch. Step-sized chunks. Each piece testable in isolation.

---

## Session changelog

For accuracy when reviewing:

- **First half** — produced the comprehensive plan doc, runbook, three SQL artefacts. Identified that adoption + verifier extension + new exporter were the bulk of the work.
- **Decision: move news decision into LLM** — produced 005, surgically extending the classifier prompt with `content_features` block and removing the redundant `enrich_news_feed` step from improvement-loop. The Go `verticalNewsMap` becomes orphan code.
- **DB check** — user confirmed `business_med_prices` doesn't exist, but spotted `products` + `product_prices` already in place. Schema check confirmed reuse viable.
- **Schema unification decision** — rather than create a new sibling table, unify under existing schema. Migration 006 written: adds `kind` to `products`, migrates `business_prices` → `(products kind=service + product_prices)`, deprecates `business_prices`. Original migration 001 scrapped.
- **Bootstrap automation decisions** — RSS seeding stays manual (Option A); content_features bootstrap waits for improvement-sweep cron (Option E). Both decisions deliberate; revisit later if scaling needs change.

This handoff written 2026-05-18.
