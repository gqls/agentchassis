# Veterinary Medicine Price Collection — Implementation Status

## Date: 2026-04-08

## Objective

Collect real, current UK veterinary medicine prices from online pet pharmacies to power vetcomparison.co.uk and related sites. Prices are verifiable — sourced directly from retailer product pages with screenshot evidence stored in B2.

## Current Coverage (as of 2026-04-08)

| Retailer | Domain | Listings | Products w/ prices | Variants | Success Rate |
|---|---|---|---|---|---|
| Pet Drugs Online | petdrugsonline.co.uk | 42 | 37 | 406 | 91.4% |
| Animed Direct | animed.co.uk | 202 | 157 | 793 | 67.8% |
| Hyperdrug | hyperdrug.co.uk | 53 | 38 | 354 | 85.8% |
| VioVet | viovet.co.uk | 1 | 1 | 24 | — (disabled) |
| **Total** | | **297** | **233** | **1,577** | |

VioVet is disabled — their prescription pages are JS-rendered, so Firecrawl's static scrape doesn't extract product links. Needs Firecrawl's `waitFor` rendering or manual URL seeding. Parked for now.

Animed's lower success rate is expected — many products are out of stock with prices hidden in the HTML. The scraper correctly reports 0 variants for these.

## Schema (deployed in `business_intel`)

| Table | Purpose | Status |
|---|---|---|
| `med_products` | Canonical medicine catalog (TEXT PK) | Created, mostly unpopulated (stage 2) |
| `med_retailers` | 4 tracked pharmacies with `category_urls TEXT[]` | Created, 3 active + 1 disabled |
| `med_retailer_listings` | Product URLs per retailer, `UNIQUE(retailer_id, retailer_url)` | 297 rows |
| `med_price_snapshots` | Price history with size variants, TVP, stock status | 1,577 rows |
| `med_price_current` | Materialized view for quick export | Refreshed after each scrape run |
| `med_scrape_evidence` | Raw markdown + content hash + screenshot URL per page fetch | ~500 rows |

## Pipeline Stages — Implementation Status

### Stage 1: URL Discovery ✓ DEPLOYED

Two approaches available:

**Category page scraper** (`med_discover_urls` action, `med-url-discoverer` agent):
- Scrapes retailer `category_urls` with Firecrawl, extracts markdown links
- Pagination support (follows "Next" links and `page=N` parameters)
- Deny-list URL filtering (nav, info, asset pages)
- Link text filtering (rejects "Account", "Cookie Settings", etc.)
- Works well for PDO, Animed, Hyperdrug. Does not work for VioVet (JS-rendered).

**Site-wide map** (`med_map_urls` action, `med-url-mapper` agent):
- Uses Firecrawl's `/map` endpoint for broad site-wide discovery
- Single API call per retailer, up to 5,000 URLs
- Filtered through the same `isRetailerProductURL` deny-list
- Useful for new retailers or broad discovery. Less targeted than category scraping.

Both feed into `med_retailer_listings` with `match_method = 'url_discovery'`.

### Stage 2: Product Matching — NOT YET BUILT

Planned: LLM matches unmatched listings to canonical `med_products` entries. Creates new products for unknown medicines. Not blocking — the export works without it by grouping on `retailer_product_name`.

### Stage 3: Price Extraction ✓ DEPLOYED

**Agent:** `med-price-collector`
**Action:** `med_scrape_prices`

Extraction hierarchy (tried in order):
1. Variant pattern: `SIZE Price: £X.XX Regular Price: £Y.YY` (Pet Drugs Online)
2. Single price: `Price: £X.XX` with TVP lookup
3. Joined format: `£28.99£50.54 TVP` (NexGard on PDO)
4. Multi-line VioVet: `- Species » Size\n£Price`
5. Multi-line Hyperdrug: `- SIZE\n...\n£Price` (25-line lookahead)
6. Multi-line Animed: `SIZE\nwas £Price\nOut of stock`
7. Standalone: `£XX.XX` on own line
8. **LLM fallback**: CPU Mistral via Ollama when regex finds 0 variants but page contains `£`

LLM fallback stats: 31 successes at ~44s average, some timeouts at 240s (being increased to 600s). Regex handles ~90%+ of pages. LLM responses logged to `llm_call_log` for future LoRA fine-tuning.

**Evidence chain:** Each scrape stores raw markdown, content hash, and Firecrawl screenshot URL in `med_scrape_evidence`. Screenshots are downloaded from Firecrawl's temporary GCS URL and re-uploaded to B2 (`med-evidence/{retailer}/{date}/{uuid}.png`).

### Stage 4: JSON Export ✓ CODE WRITTEN, PENDING DEPLOY

**Agent:** `med-json-exporter`
**Action:** `med_export_json`

Configurable per-site via `input_data`:
- `domain` — target site (default: `vetcomparison.co.uk`)
- `repo_name` — git repo (default: `sites`)
- `data_path` — base path for JSON files (default: `data`)
- `filters.species` — only products for these species (empty = all)
- `filters.categories` — only these categories (empty = all)
- `filters.retailers` — only these retailers (empty = all active)
- `outputs` — which JSON files to generate (index, full, by_letter, metadata)

Produces:
- `data/medicine-index.json` — slim catalog for search/autocomplete
- `data/medicine-prices.json` — full dataset grouped by product
- `data/medicines_by_letter/{A-Z}.json` — letter-bucketed
- `data/price-metadata.json` — export metadata (timestamp, counts, retailer breakdown)

Commits via git-adapter → GitHub Actions → B2/S3 → live site.

## Agent Architecture

All med pricing agents run as dynamically spawned K8s Jobs via orchestrator wrappers on the business-intel pod:

| Orchestrator | Worker | What it does |
|---|---|---|
| `med-price-scrape-orchestrator` | `med-price-collector` | Spawn → call → scrape prices |
| `med-url-discover-orchestrator` | `med-url-discoverer` | Spawn → call → discover URLs |
| `med-url-map-orchestrator` | `med-url-mapper` | Spawn → call → /map discovery |
| `med-export-orchestrator` | `med-json-exporter` | Spawn → call → export JSON |

All listen on `system.agent.business-intel.requests`. The business-intel pod receives the trigger, spawns a temporary pod, sends it work via Kafka, waits for completion.

Direct execution (without spawn) also works — trigger with the worker agent type instead of the orchestrator.

## File Locations

| File | Path | Status |
|---|---|---|
| Price scrape action | `platform/orchestration/actions/vet_med_price_scrape_action.go` | Deployed |
| URL discovery action | `platform/orchestration/actions/vet_med_url_discovery_action.go` | Deployed |
| URL map action | `platform/orchestration/actions/vet_med_url_map_action.go` | Deployed |
| JSON export action | `platform/orchestration/actions/vet_med_export_action.go` | Pending deploy |
| Spawn actions (Firecrawl passthrough) | `platform/orchestration/actions/spawn_actions.go` | Deployed |

Registry entries in `platform/orchestration/actions/registry.go`:
```
"med_scrape_prices":  deployed
"med_discover_urls":  deployed
"med_map_urls":       deployed
"med_export_json":    pending deploy
```

## Scheduled Tasks (all disabled, pending verification)

| Task | Agent | Interval | Topic |
|---|---|---|---|
| `med-scrape-prices` | `med-price-scrape-orchestrator` | 2 days | `system.agent.business-intel.requests` |
| `med-discover-urls` | `med-url-discover-orchestrator` | Weekly | `system.agent.business-intel.requests` |
| `med-export-json` | `med-json-exporter` | 2 days | `system.agent.business-intel.requests` |

## Remaining Work

### Immediate
- Deploy JSON export action + registry entry
- Test export to vetcomparison.co.uk
- Enable scheduled tasks after verification

### Near-term
- Product matching (stage 2) — LLM matches listings across retailers to `med_products`
- Cross-retailer price comparison in the JSON output
- LLM timeout increase to 600s (code written, awaiting image rebuild)

### Later
- VioVet — Firecrawl `waitFor` rendering or manual URL seeding
- LoRA fine-tuning from `llm_call_log` training data
- Additional retailers (VetUK, UK Pet Drugs, The Pharm Pet Co)
- Affiliate feed integration when approved
- Price trend analysis and alerting
