# News Feed Pipeline — Handoff (2026-03-31)

## What We're Building

A news/content feed pipeline for the multi-agent website builder. Sites get relevant industry news on their homepage via a JSON file fetched client-side, with a path toward full insights pages with research analysis.

Test site: gaswholesalers.com (`5fe15466-4e2e-4ff2-981e-98c1b7074002`)

## Current State — What's Working

### Ingestion (all 4 source types tested)
- RSS, scrape, news_search, api_news (Grok) — all writing to `content_feed_items`
- 34 items ingested for test site
- Dedup, error tracking with backoff, flexible config formats

### Triage (working)
- LLM scores items against full site spec (identity, classification, content_direction)
- 14 items scored relevant (62-85), 20 nav links rejected (10)
- Fix applied: `input_fields` was missing from score_relevance step — added `["input_data", "pending_items", "site_spec"]`

### JSON News File (working)
- `/data/latest-news.json` committed to git repo, deployed to S3
- 6 items with headlines, summaries, sources, relative dates
- Content-feed-orchestrator workflow: `dispatch_sources → render_news_json → commit_news → complete`

### Classification Enrichment (done)
- `content_features.news_feed.recommended: true` in classification spec
- Site plan updated with `latest-news` in homepage sections

### Planner Prompt (done)
- Rule 12 appended to build-site-planner's prompt template
- Future builds will include `latest-news` when classification recommends it

### Discovery Checks (deployed, 4 checks enabled)
- `missing_news_sources`, `missing_news_section`, `stale_news_section`, `all_sources_erroring`

### CSS Snippet (deployed)
- `Latest News Grid` in `css_snippets` table (uses `applies_to` JSONB, NOT `function`/`category`)

## What's NOT Working Yet

### Homepage doesn't show news section
The page-build-handler reads section list from `pages.sections`, which doesn't include `latest-news`. We traced this through the full build pipeline:

**Three places section lists exist:**
1. `site_specs.site_plan` → authoritative, has `latest-news` ✓
2. `pages.sections` → what the builder reads, missing `latest-news` ✗  
3. `sites.content_data` → legacy, not used by current builder

**The fix (in progress, not yet deployed):**

1. New Go action: `load_page_sections_from_spec_action.go` — reads from `site_specs.site_plan`, falls back to `pages.sections`, syncs the two
2. SQL amendment: `026h_page_build_handler_amend.sql` — adds `load_spec_sections` step to page-build-handler workflow

**Current state of the amendment:**
- Backup taken: `agent_def_page_build_handler_backup_20260331` ✓
- Go action written but NOT yet deployed (needs chassis rebuild)
- SQL amendment NOT yet run (depends on Go action being in chassis)
- Revert command needs UPDATE not DELETE due to FK on `agent_instances`

**Quick test path (no chassis rebuild):**
```sql
-- Add latest-news to pages.sections directly
UPDATE pages
SET sections = '["hero", "features", "services-grid", "differentiators-section", "social_proof", "latest-news", "call_to_action"]'::jsonb,
    updated_at = NOW()
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND name = 'index';
```
Then trigger a page build for index. The existing content writer handles `latest-news` (its `input_schema` has `source: "llm"` fields and `detectNeedsLLMContent` catches it).

## Schema Gotchas Discovered

- `css_snippets`: columns are `name, description, css_content, semantic_tags, applies_to`. NO `function` or `category` columns.
- `site_specs.created_by`: NOT NULL — all INSERTs must include it
- `agent_definitions`: FK to `agent_instances` — can't DELETE, must UPDATE for reverts
- `SavePageSectionsAction`: does `DELETE FROM page_components WHERE page_id` then re-inserts. Data-driven components get wiped on rebuild. The `load_spec_sections` fix is the proper solution (content writer generates the section from the template).
- Consumer group offsets go stale after chassis restart. Manual reset needed or deploy the `KAFKA_START_OFFSET=latest` env var fix (`consumer_offset_patch.go`).

## Files in /mnt/user-data/outputs/

### Go files (for chassis)
| File | Destination | Status |
|------|------------|--------|
| `feed_actions.go` | `actions/` | Needs deploy (has date validation + title-empty filter) |
| `feed_normalize_action.go` | `actions/` | Needs deploy |
| `feed_fetch_async_actions.go` | `actions/` | Needs deploy |
| `dispatch_feed_sources_action.go` | `actions/` | Needs deploy |
| `feed_triage_actions.go` | `actions/` | Needs deploy |
| `render_news_section_action.go` | `actions/` | Needs deploy (JSON output, date filtering, expiry) |
| `feed_news_recommendation_action.go` | `actions/` | Needs deploy |
| `load_page_sections_from_spec_action.go` | `actions/` | Needs deploy (NEW — for handler amendment) |
| `check_news_feed.go` | `actions/discovery_checks/` | Needs deploy |
| `feed_registry_entries.go` | merge into `registry.go` | 13 feed entries |
| `render_css_snippets_patch.go` | patch `render_css_from_spec_action.go` | Uses `applies_to &&` not `function` |
| `consumer_offset_patch.go` | patch `platform/kafka/consumer.go` | `KAFKA_START_OFFSET` env var |
| `save_page_sections_patch.go` | **SUPERSEDED** — don't use | Replaced by load_spec_sections approach |

### SQL files
| File | Status |
|------|--------|
| `026_content_sources_table.sql` | Run ✓ |
| `026b_agent_definitions_feed_pipeline.sql` | Run ✓ (has input_fields fix) |
| `026c_latest_news_component.sql` | Run ✓ |
| `026_stage2_deploy.sql` | Run ✓ (corrected version with applies_to, created_by) |
| `026h_page_build_handler_amend.sql` | NOT YET RUN — needs Go action deployed first |

### Docs
| File | What |
|------|------|
| `026_news_feed_pipeline.md` | Main pipeline doc (42 resolved decisions) |
| `026f_news_build_pipeline_integration.md` | How news integrates with build pipeline |
| `026g_news_expansion_architecture.md` | Three-tier expansion: JSON → insights → research |

## Next Steps (in order)

1. **Immediate test**: Update `pages.sections` for gaswholesalers index, trigger page build, verify news section appears
2. **Deploy chassis**: Merge all Go files, rebuild, deploy
3. **Run 026h SQL**: Amend page-build-handler to read from site_specs
4. **Test full cycle**: Trigger content-feed-orchestrator → verify JSON updates → verify page has news section after rebuild
5. **Consumer offset fix**: Apply `consumer_offset_patch.go`, add `KAFKA_START_OFFSET=latest` to chassis deployment env vars
6. **CSS snippets patch**: Apply to `render_css_from_spec_action.go` so news CSS loads automatically during webdesign

## Key Architectural Decisions

- **JSON for homepage news**, not page rerender (decision 43)
- **Content writer handles data-driven sections normally** — generates headline via LLM, renders the Go template (which has JS fetch code)
- **`site_specs.site_plan` is authoritative** for section lists, `pages.sections` synced from it (decision via 026h amendment)
- **Three-tier expansion**: homepage snippets → insights section → research analysis (decision 44)
- **Event timeline table** for research continuity (future, decision 45)
- **Post-classification enrichment** via `evaluate_news_feed`, not prompt editing (decision 40)

## Kafka Operational Notes

```bash
# Consumer group offset reset (needed after chassis restart sometimes)
kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=0
sleep 10
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --group generic-requests-group \
  --topic system.agent.generic.requests \
  --reset-offsets --to-latest --execute
kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=3
```

## Revert Commands

```sql
-- Revert page-build-handler amendment (UPDATE, not DELETE — FK constraint)
UPDATE agent_definitions
SET default_config = (SELECT default_config FROM agent_def_page_build_handler_backup_20260331 LIMIT 1),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Revert classification spec enrichment
-- (just set is_current=false on the enriched version, true on the original)

-- Revert planner rule 12
-- (restore from agent_definitions_backup_20260322 for build-site-planner)
```
