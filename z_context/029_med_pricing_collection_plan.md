For the new chat, start with: "Continue from the handoff doc. I want to start building the med pricing collection pipeline — schema first, then test extraction from Pet Drugs Online."



# Veterinary Medicine Price Collection — Design Plan

## Objective

Collect real, current UK veterinary medicine prices from online pet pharmacies to power the vetcomparison.uk medicine cost comparison tool. Prices must be verifiable — sourced directly from retailer product pages, not estimated.

## Target Retailers

| Retailer | Domain | Group | URL Pattern |
|---|---|---|---|
| Pet Drugs Online | petdrugsonline.co.uk | IVC Evidensia | `/prescriptions/m/metacam-oral-suspension-for-dogs` |
| Animed Direct | animeddirect.co.uk | CVS Group | `/metacam-15mgml-oral-suspension-for-dogs-180ml` |
| VioVet | viovet.co.uk | Covetrus | `/Metacam-Oral-Suspension/c5/` |
| Hyperdrug | hyperdrug.co.uk | Independent | `/metacam-1-5mg-ml-oral-suspension-for-dogs/` |

Potential future additions: VetUK (vetuk.co.uk), UK Pet Drugs (ukpetdrugs.co.uk), The Pharm Pet Co (thepharmpetco.co.uk).

## Observed Price Structure

From Pet Drugs Online's Metacam page:
```
Metacam Oral Suspension for Dogs 1.5mg/ml
  - 10ml bottle:  £3.89  (TVP £14.09)
  - 32ml bottle:  £6.29  (TVP £24.34)
  - 100ml bottle: £17.48 (TVP £67.45)
  - 180ml bottle: £23.99 (TVP £90.54)
```

Key observations:
- Products have **size variants** (same medicine, different pack sizes)
- Some retailers show "TVP" (Typical Vet Price) — useful for comparison but we shouldn't rely on it
- Prices are precise (£17.48 not £18.00) — confirms real data
- POM-V (prescription-only) medicines require prescription — affects user flow
- Some medicines are OTC (over-the-counter) — no prescription needed

## Database Schema

New tables in `business_intel` schema.

### `med_products` — Canonical medicine catalog

```sql
CREATE TABLE business_intel.med_products (
    id              TEXT PRIMARY KEY,          -- e.g. 'm_metacam_dog_100'
    name            TEXT NOT NULL,             -- 'Metacam Oral Suspension (Dog)'
    generic_name    TEXT,                      -- 'Meloxicam'
    brand           TEXT,                      -- 'Metacam'
    manufacturer    TEXT,                      -- 'Boehringer Ingelheim'
    species         TEXT[],                    -- {'dog','cat'}
    category        TEXT,                      -- 'nsaid', 'antiparasitic', 'cardiac', etc.
    form            TEXT,                      -- 'oral_suspension', 'tablet', 'spot_on', 'injection'
    strength        TEXT,                      -- '1.5mg/ml', '5.4mg'
    pack_size       TEXT,                      -- '100ml', '100 tabs', '3 pipettes'
    prescription_required BOOLEAN DEFAULT true,
    is_active       BOOLEAN DEFAULT true,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_med_products_name ON business_intel.med_products USING gin (to_tsvector('english', name));
CREATE INDEX idx_med_products_category ON business_intel.med_products (category);
CREATE INDEX idx_med_products_species ON business_intel.med_products USING gin (species);
```

### `med_retailers` — Tracked pharmacies

```sql
CREATE TABLE business_intel.med_retailers (
    id              TEXT PRIMARY KEY,          -- 'pet_drugs_online'
    name            TEXT NOT NULL,             -- 'Pet Drugs Online'
    domain          TEXT NOT NULL,             -- 'petdrugsonline.co.uk'
    group_name      TEXT,                      -- 'IVC Evidensia'
    base_url        TEXT NOT NULL,             -- 'https://www.petdrugsonline.co.uk'
    category_urls   TEXT[],                    -- URLs to crawl for product discovery
    delivery_cost   NUMERIC,                   -- standard delivery cost
    free_delivery_threshold NUMERIC,           -- free delivery above this amount
    is_active       BOOLEAN DEFAULT true,
    scrape_config   JSONB DEFAULT '{}',        -- retailer-specific extraction hints
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
```

### `med_retailer_listings` — Product URLs per retailer

```sql
CREATE TABLE business_intel.med_retailer_listings (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    retailer_id     TEXT NOT NULL REFERENCES business_intel.med_retailers(id),
    product_id      TEXT REFERENCES business_intel.med_products(id),  -- NULL until matched
    retailer_url    TEXT NOT NULL,
    retailer_product_name TEXT,                -- their name for this product
    retailer_sku    TEXT,                      -- their internal ID if available
    match_confidence NUMERIC,                  -- how confident is the product mapping
    match_method    TEXT,                      -- 'manual', 'llm', 'exact_name'
    is_active       BOOLEAN DEFAULT true,
    last_scraped_at TIMESTAMPTZ,
    last_price      NUMERIC,                   -- most recent price (denormalized for quick access)
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (retailer_id, retailer_url)
);

CREATE INDEX idx_med_listings_product ON business_intel.med_retailer_listings (product_id);
CREATE INDEX idx_med_listings_retailer ON business_intel.med_retailer_listings (retailer_id);
CREATE INDEX idx_med_listings_unmatched ON business_intel.med_retailer_listings (retailer_id) 
    WHERE product_id IS NULL AND is_active = true;
```

### `med_price_snapshots` — Price history

```sql
CREATE TABLE business_intel.med_price_snapshots (
    id              BIGSERIAL PRIMARY KEY,
    listing_id      UUID NOT NULL REFERENCES business_intel.med_retailer_listings(id),
    product_id      TEXT REFERENCES business_intel.med_products(id),
    retailer_id     TEXT NOT NULL REFERENCES business_intel.med_retailers(id),
    size_variant    TEXT,                      -- '100ml', '180ml', '100 tabs'
    price           NUMERIC NOT NULL,
    currency        TEXT DEFAULT 'GBP',
    in_stock        BOOLEAN,
    typical_vet_price NUMERIC,                -- TVP if shown on page
    collected_at    TIMESTAMPTZ DEFAULT NOW(),
    collection_method TEXT DEFAULT 'scrape',   -- 'scrape', 'feed', 'manual'
    raw_data        JSONB DEFAULT '{}',        -- full extracted data for debugging
    UNIQUE (listing_id, size_variant, collected_at)
);

CREATE INDEX idx_med_prices_product ON business_intel.med_price_snapshots (product_id, collected_at DESC);
CREATE INDEX idx_med_prices_retailer ON business_intel.med_price_snapshots (retailer_id, collected_at DESC);
CREATE INDEX idx_med_prices_latest ON business_intel.med_price_snapshots (listing_id, collected_at DESC);

-- Partition by month if volume grows
```

### `med_price_current` — Materialized view for quick export

```sql
CREATE MATERIALIZED VIEW business_intel.med_price_current AS
SELECT DISTINCT ON (ps.listing_id, ps.size_variant)
    ps.listing_id,
    ps.product_id,
    ps.retailer_id,
    r.name AS retailer_name,
    r.group_name,
    r.domain,
    l.retailer_url,
    p.name AS product_name,
    p.brand,
    p.category,
    p.species,
    p.pack_size,
    ps.size_variant,
    ps.price,
    ps.in_stock,
    ps.typical_vet_price,
    ps.collected_at
FROM business_intel.med_price_snapshots ps
JOIN business_intel.med_retailer_listings l ON l.id = ps.listing_id
JOIN business_intel.med_retailers r ON r.id = ps.retailer_id
LEFT JOIN business_intel.med_products p ON p.id = ps.product_id
WHERE ps.collected_at > NOW() - INTERVAL '14 days'
ORDER BY ps.listing_id, ps.size_variant, ps.collected_at DESC;

CREATE UNIQUE INDEX idx_med_price_current ON business_intel.med_price_current (listing_id, size_variant);
```

## Pipeline Stages

### Stage 1: URL Discovery (Crawl)

**Agent:** `med-url-discoverer`
**Action:** `med_discover_urls`
**Method:** Firecrawl crawl mode on each retailer's category pages.
**Frequency:** Weekly or on-demand.
**Output:** New rows in `med_retailer_listings`.

For each retailer, crawl their prescription/medicine category pages and extract all product URLs. Firecrawl's crawl mode follows internal links and returns page URLs with metadata.

The category URLs vary by retailer:
- Pet Drugs Online: `/prescriptions/` (alphabetical A-Z)
- VioVet: `/Dogs/Pharmacy-Dog/c6/` + `/Cats/Pharmacy-Cat/c158/`
- Hyperdrug: `/pet-prescriptions/`
- Animed: `/prescriptions/`

This stage produces unmatched listings — URLs with the retailer's product name but no `product_id` yet.

### Stage 2: Product Matching (LLM-assisted)

**Agent:** `med-product-matcher`
**Action:** `med_match_products`
**Method:** LLM batch review of unmatched listings.
**Frequency:** After each discovery run.

For each unmatched listing, the LLM receives the retailer's product name and page title and matches it to an existing `med_products` entry. If no match exists, it creates a new canonical product entry.

Example prompt context:
```
Retailer product: "Metacam 1.5mg/ml Oral Suspension for Dogs 100ml"
Existing products: [list of med_products entries]
→ Match to: m_metacam_dog_100 (confidence: 0.98)
```

The LLM also normalises product attributes: species, form, strength, pack size.

New medicines discovered during crawling get added to `med_products` automatically, expanding the catalog beyond the initial 130.

### Stage 3: Price Extraction (Scrape)

**Agent:** `med-price-collector`
**Action:** `med_scrape_prices`
**Method:** Firecrawl scrape with structured extraction per product page.
**Frequency:** Daily or every 2-3 days.
**Output:** New rows in `med_price_snapshots`, updated `last_price` on listings.

For each active listing, scrape the product page and extract:
- Product name (for verification)
- Size variants with prices
- In-stock status
- TVP (typical vet price) if shown

Firecrawl's extract mode with a JSON schema works here:
```json
{
  "product_name": "string",
  "variants": [{
    "size": "string",
    "price": "number",
    "in_stock": "boolean",
    "vet_price": "number or null"
  }]
}
```

Rate limiting: 1 request per second per domain. ~500 listings × 4 retailers = 2000 pages. At 1/sec per domain running in parallel, ~8 minutes for a full sweep.

### Stage 4: Static JSON Export

**Agent:** `med-export` (or a CTE-only scheduled task)
**Action:** `med_export_json`
**Method:** Query `med_price_current`, format as JSON, commit to git.
**Frequency:** After each price collection run.
**Output:** Updated `medicine-index.json` and letter-bucketed files.

Export format matches the current site structure:

`medicine-index.json` — slim catalog for autocomplete:
```json
[{"id": "m_metacam_dog_100", "name": "Metacam Oral Suspension (Dog) 100ml"}]
```

`medicines_by_letter/M.json` — full price data per letter:
```json
[{
  "id": "m_metacam_dog_100",
  "name": "Metacam Oral Suspension (Dog)",
  "dosage": "1.5mg/ml (100ml)",
  "options": [
    {"retailer": "Pet Drugs Online", "group": "IVC Evidensia", "price": 17.48, "url": "https://..."},
    {"retailer": "Animed Direct", "group": "CVS Group", "price": 18.25, "url": "https://..."}
  ]
}]
```

Git commit via the existing `git_commit` action to the vetcomparison.uk repo.

## Scheduling

| Task | Agent | Interval | Concurrency |
|---|---|---|---|
| `med-discover-urls` | `med-url-discoverer` | Weekly (604800s) | med-collection |
| `med-match-products` | `med-product-matcher` | Daily (86400s) | med-collection |
| `med-scrape-prices` | `med-price-collector` | 2 days (172800s) | med-collection |
| `med-export-json` | CTE-only or agent | After price collection | med-collection |

All share a `med-collection` concurrency group to avoid overlapping.

## Implementation Order

1. **Schema migration** — create the 4 tables + materialized view
2. **Seed retailers** — insert the 4 retailer rows with category URLs
3. **Manual test** — scrape 5 products from Pet Drugs Online, validate extraction
4. **URL discovery action** — crawl category pages, populate listings
5. **Product matching action** — LLM match unmatched listings to products
6. **Price extraction action** — scrape product pages, store snapshots
7. **Export action** — generate JSON, commit to git
8. **Pipeline admin** — add to scheduled_tasks, visible in admin dashboard

## Key Decisions

**Why not affiliate feeds?** Need a working site with traffic first. Scraping gives us real data now. When affiliate programmes accept us, we switch the `collection_method` from 'scrape' to 'feed' with no schema changes.

**Why size variants?** The same medicine comes in 10ml, 32ml, 100ml, 180ml bottles at different prices. Users need to compare the exact size their vet prescribed. The current static data ignores this — the new schema captures it.

**Why snapshot history?** Price trends over time are useful content ("Apoquel has increased 12% in the last 6 months"). Also catches extraction errors — if a price jumps 10x overnight, it's probably wrong.

**Why materialized view?** The export and site both need "latest price per product per retailer per size." Computing this from snapshots on every request is expensive. Refreshing a materialized view after each collection run is fast.

**Expanding the catalog:** The current `medicine-index.json` has ~130 medicines. Pet Drugs Online alone has 500+. By crawling their full catalog, we'll discover medicines we didn't know about. The matching stage handles adding these to `med_products` automatically.

## Risks

- **Scraping TOS** — all 4 retailers are commercial sites. Rate-limit aggressively (1 req/sec). Consider adding `User-Agent` identifying our service. Switch to affiliate feeds when available.
- **Price extraction accuracy** — HTML structure varies between retailers. Each needs a tailored extraction schema or CSS selector. Test on 10-20 products per retailer before running at scale.
- **Product matching errors** — "Metacam 1.5mg/ml 100ml" vs "Metacam Oral Suspension for Dogs 100ml Bottle" are the same product. LLM matching handles this, but needs human review for the first batch.
- **Stock/availability** — some products go out of stock. Track `in_stock` and don't show out-of-stock options in comparisons.
