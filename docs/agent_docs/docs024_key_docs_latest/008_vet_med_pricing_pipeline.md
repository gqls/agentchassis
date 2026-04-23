# 008 — Vet Med Pricing Pipeline

## Date: 2026-04-08

## Overview

This document describes the automated pipeline that collects veterinary medicine prices from UK online pharmacies and publishes them as JSON data to vet comparison websites. The pipeline runs on the agent chassis platform, uses Firecrawl for web scraping, CPU Mistral for fallback price extraction, and the git-adapter for deployment.

The same pipeline can serve multiple sites with different data slices — filtered by species, medicine category, or retailer subset.

## How It Works

```
Trigger (scheduled or manual)
    │
    ▼
Business-intel pod receives message
    │
    ▼
Spawns temporary K8s Job pod
    │
    ├─ URL Discovery: scrape category pages → find product URLs
    ├─ Price Collection: scrape product pages → extract prices + screenshots
    └─ JSON Export: query DB → build JSON → commit to git → live site
```

Each stage runs as its own agent in a spawned pod with clean logs. Stages can be triggered independently.

## Data Flow

### 1. URL Discovery

Two methods available per retailer:

**Category scraper** — scrapes specific category pages (e.g. `/prescriptions/`), follows pagination, extracts product links from markdown. Targeted: only discovers products listed on the configured category pages.

**Site map** — calls Firecrawl's `/map` endpoint for broad site-wide URL discovery. Returns up to 5,000 URLs in one call. Less targeted but catches products the category scraper misses.

Both methods filter discovered URLs through a deny-list (navigation pages, info pages, assets, non-product paths) and upsert into `med_retailer_listings`.

### 2. Price Collection

For each product listing, the scraper:

1. Calls Firecrawl v2 `/scrape` with screenshot capture enabled
2. Receives markdown + screenshot URL
3. Extracts the product section (truncates before "Description" / "Related Products" headings)
4. Runs regex patterns in order (retailer-specific formats → generic patterns)
5. If regex finds 0 variants but the page contains `£`, falls back to CPU Mistral LLM
6. Downloads the screenshot from Firecrawl's temporary GCS URL
7. Re-uploads screenshot to B2 persistent storage
8. Stores price snapshots, evidence (markdown + content hash + screenshot URL)
9. Refreshes the `med_price_current` materialized view

**Rate limiting:** 2-second delay between Firecrawl calls per batch. Batches of 10-50 listings per run.

**Evidence chain:** Every scraped page has its raw markdown, a content hash for change detection, and a screenshot proving the page state at scrape time. Screenshots persist in B2 indefinitely.

### 3. LLM Fallback

When regex extraction finds nothing but the page contains price indicators (`£`):

- Sends truncated markdown (~1500 chars) to CPU Mistral (`mistral-small3.1`) via the cluster Ollama service
- Prompt asks for a JSON array: `[{"size": "100ml", "price": 17.48, "tvp": 0, "in_stock": true}]`
- Temperature 0.1, 500 max tokens
- Response logged to `llm_call_log` with full prompt and response text
- This serves dual purpose: price extraction now, training data for future LoRA fine-tuning

LLM latency: 80-280s per call on CPU. Regex handles ~90% of pages, so LLM only triggers for edge cases.

### 4. JSON Export

Queries the latest prices and produces JSON files committed to the site's git repository:

| File | Purpose |
|---|---|
| `data/medicine-index.json` | Slim catalog for search/autocomplete (id, name, brand, option count) |
| `data/medicine-prices.json` | Full dataset: products → options (retailer, size, price, stock, TVP, URL) |
| `data/medicines_by_letter/{A-Z}.json` | Same data bucketed by first letter for pagination |
| `data/price-metadata.json` | Export timestamp, product/variant counts, per-retailer stats |

Git commit → GitHub Actions → B2/S3 → live site. No build step needed — the JSON files are consumed directly by the site's frontend.

## Multi-Site Configuration

The export action accepts configuration via `input_data`, enabling the same pipeline to serve different sites:

```json
{
  "domain": "vetcomparison.co.uk",
  "data_path": "data",
  "filters": {
    "species": ["dog", "cat"],
    "categories": [],
    "retailers": []
  },
  "outputs": {
    "index": true,
    "full": true,
    "by_letter": true,
    "metadata": true
  },
  "commit_message_prefix": "Update medicine prices"
}
```

Filters:
- `species` — only products for these species (uses PostgreSQL array overlap). Empty = all.
- `categories` — only these medicine categories (e.g. "nsaid", "antiparasitic"). Empty = all.
- `retailers` — only these retailer IDs. Empty = all active retailers.

Products without a `product_id` (not yet matched to `med_products`) pass all species/category filters, since their metadata is unknown. They'll be filtered properly once product matching is implemented.

## Retailer Profiles

### Pet Drugs Online (petdrugsonline.co.uk)
- **URL pattern:** Single-segment slugs (`/apoquel-16mg-tablets-for-dogs`)
- **Price format:** `Price: £X.XX` with `Regular Price: £Y.YY` for TVP
- **Discovery:** Category pages at `/dog-prescriptions`, `/cat-prescriptions`
- **Coverage:** 42 listings, 37 products, 406 variants. Fully scraped.
- **Notes:** Some short slugs (`/apoquel`) are category pages not products. Regex handles 91% of pages.

### Animed Direct (animed.co.uk)
- **URL pattern:** Single-segment slugs (`/cimalgex-chewable-tablets-for-dogs`)
- **Price format:** Multi-line: `SIZE\nwas £Price\nOut of stock`
- **Discovery:** Category page at `/prescriptions`
- **Coverage:** 202 listings, 157 products, 793 variants. Fully scraped.
- **Notes:** Many out-of-stock products with hidden prices (67.8% success rate is correct). Largest catalog.

### Hyperdrug (hyperdrug.co.uk)
- **URL pattern:** Single-segment slugs (`/metacam-1-5mg-ml-oral-suspension-for-dogs/`)
- **Price format:** Multi-line with 25-line lookahead: `- SIZE\n...\n£Price`
- **Discovery:** Category page at `/prescription-medicines/`
- **Coverage:** 53 listings, 38 products, 354 variants. Fully scraped.
- **Notes:** Reliable extraction (85.8%). Some products are equine/farm, not pet.

### VioVet (viovet.co.uk) — DISABLED
- **URL pattern:** `/Product-Name/c{digits}/` (product family pages)
- **Price format:** Multi-line: `- Species » Size\n£Price`
- **Discovery:** Category page at `/Prescription_Drugs/c1/` but JS-rendered — Firecrawl's static scrape returns no product links
- **Coverage:** 1 listing (manually seeded Metacam), 24 variants
- **Notes:** Disabled. Needs Firecrawl `waitFor` rendering or manual URL seeding. The `/map` endpoint returned the entire product catalogue (5000+ items including horse tack and dog food) — too broad without better filtering.

## Database Schema

All tables in the `business_intel` schema:

**`med_retailers`** — retailer configuration. `category_urls TEXT[]` for discovery, `is_active` to enable/disable.

**`med_retailer_listings`** — one row per product URL per retailer. Links to `med_products` via `product_id` (NULL until matched). Tracks `last_scraped_at` for scheduling.

**`med_price_snapshots`** — price history. One row per listing × size variant × collection time. Supports trend analysis.

**`med_price_current`** — materialized view. DISTINCT ON (listing_id, size_variant) ordered by collected_at DESC. Refreshed after each scrape run.

**`med_scrape_evidence`** — raw markdown, content hash, screenshot URL. Foreign-keyed to listings. Proves page state at scrape time.

**`med_products`** — canonical medicine catalog. Mostly unpopulated until product matching is implemented. The export works without it.

## Triggering

### Manual trigger (via Kafka)
```json
{"action":"orchestrate","config":{"agent_type":"med-price-scrape-orchestrator"},"input_data":{"batch_size":20,"retailer_id":"animed_direct"}}
```

### Direct execution (no spawn, runs on business-intel)
```json
{"action":"orchestrate","config":{"agent_type":"med-price-collector"},"input_data":{"batch_size":10}}
```

### JSON export
```json
{"action":"orchestrate","config":{"agent_type":"med-json-exporter"},"input_data":{"domain":"vetcomparison.co.uk"}}
```

### URL discovery
```json
{"action":"orchestrate","config":{"agent_type":"med-url-discoverer"},"input_data":{"retailer_id":"animed_direct"}}
```

All messages go to topic `system.agent.business-intel.requests`.

## Scheduled Automation

Three tasks configured (all disabled, enable after verification):

| Task | Agent | Interval | Purpose |
|---|---|---|---|
| `med-discover-urls` | `med-url-discover-orchestrator` | Weekly | Find new product URLs |
| `med-scrape-prices` | `med-price-scrape-orchestrator` | 2 days | Collect current prices |
| `med-export-json` | `med-json-exporter` | 2 days | Publish to site |

All share the `med-collection` concurrency group to avoid overlapping.

## Monitoring

```sql
-- Pipeline status
SELECT r.id, count(l.id) as listings,
       count(l.id) FILTER (WHERE l.last_scraped_at IS NOT NULL) as scraped,
       count(l.id) FILTER (WHERE l.last_scraped_at IS NULL) as pending
FROM business_intel.med_retailers r
LEFT JOIN business_intel.med_retailer_listings l ON l.retailer_id = r.id
WHERE r.is_active = true
GROUP BY r.id ORDER BY r.id;

-- Price coverage
SELECT retailer_id, count(DISTINCT listing_id) as products, count(*) as variants
FROM business_intel.med_price_snapshots
GROUP BY retailer_id ORDER BY retailer_id;

-- Extraction success rates
SELECT retailer_id, count(*) as scraped,
       count(*) FILTER (WHERE variants_found > 0) as had_prices,
       round(100.0 * count(*) FILTER (WHERE variants_found > 0) / count(*), 1) as pct
FROM business_intel.med_scrape_evidence
GROUP BY retailer_id ORDER BY retailer_id;

-- LLM fallback health
SELECT success, count(*), round(avg(latency_ms)) as avg_ms
FROM llm_call_log WHERE provider = 'ollama' AND step_name = 'scrape_prices'
GROUP BY success;
```

## Future Work

- **Product matching** — LLM links listings across retailers to canonical products. Enables cross-retailer comparison ("Metacam 100ml: £17.48 at PDO vs £18.25 at Animed").
- **VioVet** — JS rendering support via Firecrawl `waitFor` or manual URL seeding.
- **Additional retailers** — VetUK, UK Pet Drugs, The Pharm Pet Co. Same pipeline, new retailer rows.
- **LoRA fine-tuning** — `llm_call_log` accumulates markdown→JSON training pairs. Fine-tune Mistral for faster, more accurate extraction.
- **Affiliate feeds** — when approved, switch `collection_method` from 'scrape' to 'feed'. Schema supports both.
- **Price alerts** — snapshot history enables trend detection. "Apoquel up 12% this month."
- **Batch LLM processing** — overnight batch at cheaper rates for pages where regex fails.
