# Session Handoff — March 21-26, 2026

## What to Upload to the New Chat

Upload this file along with the project knowledge files. The previous handoff (`ch_enrichment_session_handoff.md`) is now superseded by this one.

## System Context

- **Kubernetes namespace:** `ai-persona-system`, Kafka in `-n kafka`
- **Database:** `clients_db` (PostgreSQL via PgBouncer), schema `business_intel` for data collection
- **Agent chassis:** All agents share the `agent-chassis` Go binary, differentiated by `AGENT_TYPE` env var
- **Admin dashboard:** React SPA in `frontends/admin-dashboard/`, served by nginx, proxies API through auth-service to core-manager
- **Deployment:** Docker build → push → kustomize apply. GitHub Actions triggers Backblaze S3 for static sites.

## What's Working

### CH Enrichment Pipeline (Complete)

Collect → cascade match → LLM review → detail fetch → accounts fetch. All stages operational.

| Stage | Agent | Status | Notes |
|---|---|---|---|
| Bulk collect | `ch-collector` | Done (5,780 SIC 75000 companies) | Run manually, monthly |
| Local matching | `ch-matcher` | Running daily | Tiers 1-3, ~634 confirmed |
| LLM review | `ch-llm-reviewer` | Running daily | Haiku, ~$0.05/run |
| Detail fetch | `ch-detail-fetcher` | Complete backfill done (711) | Officers, PSC, succession risk |
| Accounts fetch | `ch-accounts-fetcher` | Running, ~250 done of 711 | iXBRL parsing for financials |

### Accounts Fetch — Technical Details

The accounts fetch action (`ch_fetch_accounts_action.go`) was built and debugged across several iterations:

1. **CH Document API flow:** Filing history → document metadata JSON → `/content` path with `Accept: application/xhtml+xml` → 302 redirect to S3 → iXBRL document
2. **iXBRL parsing:** Regex extracts `ix:nonFraction` tags. Attribute-order-independent regex (captures full tag, then extracts `name`, `contextRef`, `sign` attributes individually).
3. **Tag matching:** Exact local-name matching (strip namespace prefix, compare exactly). This prevents `GrossProfitLoss` matching a `ProfitLoss` mapping.
4. **Sign handling:** `sign="-"` attribute means the displayed value is positive but actual value is negative. Parser negates accordingly.
5. **Tag name variants:** Multiple iXBRL tag names map to each DB field (e.g. `ProfitLossOnOrdinaryActivitiesBeforeTax`, `ProfitLossForPeriod`, `ProfitLossForFinancialYear` all → `profit_loss_gbp`).

**Financial data coverage (from 250 companies):**
- 56% have net_worth (micro entities use abbreviated balance sheets)
- 87% have employee_count
- ~1% have turnover (small/micro exempt from disclosure)
- ~1% have profit_loss (only medium+ filers)

### HTTP Request Logging

New `http_request_log` table captures all outbound HTTP calls from Go actions. Fire-and-forget pattern matching `llm_call_logger.go`. Used by accounts fetch; other CH actions can adopt it.

```sql
SELECT * FROM http_request_stats;  -- 24h summary
```

### Pipeline Admin Dashboard

New "Pipelines" tab in the admin dashboard at `frontends/admin-dashboard/src/pages/PipelinesPage.tsx`.

**Backend** (`pipeline_admin_handlers.go` in `internal/core-manager/admin/`):
- `GET /admin/pipelines` — list tasks with computed state
- `PATCH /admin/pipelines/:name` — enable/disable, change interval
- `POST /admin/pipelines/:name/trigger` — force next run
- `GET /admin/pipelines/stats` — CH progress, verification counts, HTTP/LLM summaries

**Frontend:** Stats cards with progress bars, task table with enable/disable toggles and "Run Now" buttons. Auto-refreshes 15s.

**Routing:** Auth-service gateway proxies `/pipelines` to core-manager (added in `cmd/auth-service/main.go`).

### Kafka Topic Cleanup Fix

The `agent-job-cleanup` CronJob had wrong label selector: `strimzi.io/name=personae-kafka-cluster-combined-pool-prod` should be `strimzi.io/name=personae-kafka-cluster-kafka`. Fixed manually. 8000+ orphaned topics were deleted manually. Update the YAML in the repo to persist the fix.

## What's In Progress

### Accounts Fetch Throughput

467 companies remaining. Currently limited by:
- 15s delay between API calls (safe to reduce to 5s — CH allows 600/5min)
- 30 per batch (can increase to 50)
- `ch-enrichment` concurrency group shared with `ch-detail-fetch`

To speed up:
```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
    jsonb_set(default_config, '{workflow,steps,fetch_accounts,config,delay_ms}', '5000'),
    '{workflow,steps,fetch_accounts,config,batch_size}', '50'
)
WHERE type = 'ch-accounts-fetcher' AND version = 1;
```

And either disable `ch-detail-fetch` (backfill complete) or move accounts to its own concurrency group.

## What's Planned Next

### Veterinary Medicine Price Collection

Full design plan in `026_med_pricing_collection_plan.md`. Summary:

**Goal:** Scrape real prices from 4 UK online pet pharmacies (Pet Drugs Online, Animed Direct, VioVet, Hyperdrug) to power vetcomparison.uk's medicine comparison tool.

**Schema:** 4 new tables in `business_intel`: `med_products` (canonical catalog), `med_retailers`, `med_retailer_listings` (product URLs per retailer), `med_price_snapshots` (price history). Plus a materialized view for quick export.

**Pipeline:** URL discovery (Firecrawl crawl) → product matching (LLM) → price extraction (Firecrawl scrape) → static JSON export (git commit).

**Validation done:** Pet Drugs Online product pages have clean HTML with exact prices per size variant (e.g. Metacam 100ml: £17.48). Firecrawl can extract this.

**Starting point:** Schema migration → seed retailers → manual test extraction of 5 products from Pet Drugs Online → then automate.

### Other Threads (Lower Priority)

- **Upstream match rate** — tier 0 (company number from websites, ~750-1000 potential), tier 5 (corporate group matching, ~300-500)
- **RAG aggregation** — once 500+ companies have financial data, summarise into knowledge_base chunks
- **LoRA training** — llm_call_log accumulating training data; site-classifier and ch-llm-review are candidates for local 7B replacement

## Key Files Produced This Session

### Deployed
| File | Location | Purpose |
|---|---|---|
| `ch_fetch_accounts_action.go` | `platform/orchestration/actions/` | Accounts fetch with iXBRL parsing + HTTP logging |
| `http_request_logger.go` | `platform/orchestration/actions/` | Centralised HTTP request logger |
| `http_request_log_migration.sql` | Applied to clients_db | HTTP logging table + stats view |
| `ch_accounts_migration.sql` | Applied to clients_db | accounts_fetched columns + financial columns |
| `ch_accounts_fetch_task.sql` | Applied to clients_db | Scheduled task for accounts fetch |
| `pipeline_admin_handlers.go` | `internal/core-manager/admin/` | Pipeline admin API endpoints |
| `PipelinesPage.tsx` | `frontends/admin-dashboard/src/pages/` | Pipeline admin frontend |
| `pipeline_routes_patch` | `cmd/auth-service/main.go` + core-manager `server.go` | Route registration |

### Not Yet Deployed
| File | Purpose |
|---|---|
| `026_med_pricing_collection_plan.md` | Med price collection design plan |
| `session_summary_2026_03_21.md` | Detailed session notes |

## Scheduled Tasks State

```
ch-vet-collect:          disabled (monthly, run manually)
ch-local-match:          enabled, daily, concurrency: ch-matching
ch-llm-review:           enabled, daily, concurrency: ch-matching
ch-detail-fetch:         enabled, 20min, concurrency: ch-enrichment (backfill complete)
ch-fetch-accounts:       enabled, 20min, concurrency: ch-enrichment (467 remaining)
ch-enrichment:           disabled (legacy)
vet-batch-verify:        disabled (enable to re-verify for company numbers)
build-pipeline-trigger:  enabled
improvement-sweep:       disabled
stuck-task-reaper:       enabled
database-cleanup:        enabled
claimed-item-timeout:    enabled
stale-orchestration-reaper: enabled
feasibility-recheck:     enabled
```

## Schema Notes

**`agent_definitions` columns:** `display_name` (not `name`), `is_active` (not `enabled`), `image_repository` (not `container_image`), `capabilities` is jsonb (not text[]).

**`scheduled_tasks` columns:** `target_agent_type`, `target_topic`, `interval_seconds` (integer, not cron), `input_data` (jsonb), `pre_query` (must return 0 rows to skip — `SELECT EXISTS` always returns 1 row and doesn't work).

**Business-intel pod:** Single replica, all CH agents share it. Scheduler sends `config.agent_type` in message, `selectWorkflow` → `FindBestGroup` loads correct workflow. `ai_service` goes in step config, not agent_config, on shared pods.
