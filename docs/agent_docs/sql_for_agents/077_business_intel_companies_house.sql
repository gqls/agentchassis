-- =========================================================================
-- 2. AGENT DEFINITION
-- =========================================================================
-- The ch-enricher loads a batch of verified businesses that haven't been
-- enriched yet, and for each one: searches Companies House, picks the
-- best match, fetches profile + officers + PSC, stores the data.
--
-- Runs on the vet-intel pod (same topic).

renamed to business-intel

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active,
    capabilities, image_repository, image_tag, resources,
    agent_category, status, domain_tags,
    input_contract, output_contract, idle_timeout_seconds
) VALUES (
             'ch-enricher',
             'Companies House Enricher',
             'Enriches verified businesses with Companies House data: financials, officers, ownership, succession signals. Very gentle rate limiting (~2 businesses/min).',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "load_batch",
                     "steps": {
                         "load_batch": {
                             "action": "load_ch_enrichment_batch",
                             "description": "Load verified businesses not yet enriched",
                             "output_field": "batch",
                             "next_step": "check_batch",
                             "config": {
                                 "batch_size": 20,
                                 "vertical_slug": "veterinary"
                             }
                         },
                         "check_batch": {
                             "action": "conditional",
                             "description": "Check if there are businesses to enrich",
                             "config": {
                                 "condition": "batch.count != 0",
                                 "then_step": "process_batch",
                                 "else_step": "complete_empty"
                             }
                         },
                         "process_batch": {
                             "action": "loop",
                             "description": "Process each business: search CH, fetch details, store",
                             "output_field": "enrichment_results",
                             "next_step": "complete",
                             "config": {
                                 "items_field": "batch.items",
                                 "item_variable": "current_business",
                                 "max_iterations": 50,
                                 "continue_on_error": true,
                                 "sub_workflow": {
                                     "start_step": "search_ch",
                                     "steps": {
                                         "search_ch": {
                                             "action": "companies_house_search",
                                             "description": "Search Companies House by business name and postcode",
                                             "output_field": "ch_search",
                                             "next_step": "check_match",
                                             "config": {
                                                 "name_field": "current_business.name",
                                                 "postcode_field": "current_business.postcode",
                                                 "sic_filter": ["75000"],
                                                 "delay_ms": 15000
                                             }
                                         },
                                         "check_match": {
                                             "action": "conditional",
                                             "description": "Check if we found a match",
                                             "config": {
                                                 "condition": "ch_search.matched != 0",
                                                 "then_step": "fetch_details",
                                                 "else_step": "store_no_match"
                                             }
                                         },
                                         "fetch_details": {
                                             "action": "companies_house_fetch",
                                             "description": "Fetch company profile, officers, PSC",
                                             "output_field": "ch_details",
                                             "next_step": "store_enrichment",
                                             "config": {
                                                 "company_number_field": "ch_search.company_number",
                                                 "fetch_officers": true,
                                                 "fetch_psc": true,
                                                 "delay_ms": 15000
                                             }
                                         },
                                         "store_enrichment": {
                                             "action": "store_ch_enrichment",
                                             "description": "Store Companies House data and derive succession signals",
                                             "output_field": "store_result",
                                             "next_step": "done",
                                             "config": {
                                                 "input_fields": ["current_business", "ch_search", "ch_details"]
                                             }
                                         },
                                         "store_no_match": {
                                             "action": "store_ch_enrichment",
                                             "description": "Record that no CH match was found",
                                             "output_field": "store_result",
                                             "next_step": "done",
                                             "config": {
                                                 "input_fields": ["current_business", "ch_search"],
                                                 "no_match": true
                                             }
                                         },
                                         "done": {
                                             "action": "loop_complete",
                                             "description": "Move to next business"
                                         }
                                     }
                                 }
                             }
                         },
                         "complete_empty": {
                             "action": "complete_workflow",
                             "description": "No businesses to enrich",
                             "config": {
                                 "output_fields": ["batch"]
                             }
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Enrichment batch complete",
                             "config": {
                                 "output_fields": ["batch", "enrichment_results"]
                             }
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 3600
             }'::jsonb,
             true,
             '["companies-house", "enrichment", "financial-data"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.862',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             'specialist',
             'experimental',
             '["business-intelligence", "enrichment", "financial-data"]'::jsonb,
             '{"optional": ["batch_size", "vertical_slug"], "required": []}'::jsonb,
             '{"produces": {"batch": "object - businesses loaded", "enrichment_results": "object - CH data stored per business"}}'::jsonb,
             180
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       updated_at = NOW();

----

-- adds a notify_scheduler step and reroutes both paths through it:
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                jsonb_set(
                        jsonb_set(
                                default_config,
                            -- 1. Add notify_scheduler step
                                '{workflow,steps,notify_scheduler}',
                                '{
                                  "action": "query_database",
                                  "config": {
                                    "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''ch-enrichment''",
                                    "output_format": "object"
                                  },
                                  "next_step": "complete",
                                  "description": "Tell scheduler this execution finished",
                                  "output_field": "scheduler_notified"
                                }'::jsonb
                        ),
                    -- 2. Add notify_scheduler_empty step
                        '{workflow,steps,notify_scheduler_empty}',
                        '{
                          "action": "query_database",
                          "config": {
                            "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''ch-enrichment''",
                            "output_format": "object"
                          },
                          "next_step": "complete_empty",
                          "description": "Tell scheduler this execution finished (empty batch path)",
                          "output_field": "scheduler_notified"
                        }'::jsonb
                ),
            -- 3. Route process_batch → notify_scheduler instead of → complete
                '{workflow,steps,process_batch,next_step}',
                '"notify_scheduler"'
        ),
    -- 4. Route check_batch else → notify_scheduler_empty instead of → complete_empty
        '{workflow,steps,check_batch,config,else_step}',
        '"notify_scheduler_empty"'
                     )
WHERE type = 'ch-enricher';
```

This follows the same pattern as build-dispatch-loop and improvement-loop — separate notify steps for the success and idle paths, both updating `last_completed_at` before reaching `complete_workflow`.

The flow becomes:
```
load_batch → check_batch
├── has items → process_batch → notify_scheduler → complete
└── empty    → notify_scheduler_empty → complete_empty


---

-- rename to ch-enricher

UPDATE agent_definitions
SET type = 'business-intel',
    display_name = 'Business Intel Agent',
    description = 'Business intelligence enrichment agent. Currently handles Companies House enrichment for verified businesses. Will expand to cover other data sources and verticals.'
WHERE type = 'ch-enricher';


-----

-- boolean matching problem

UPDATE agent_definitions
SET default_config = REPLACE(
        default_config::text,
        '"condition": "ch_search.matched != 0"',
        '"condition": "ch_search.matched == true"'
                     )::jsonb
WHERE type = 'business-intel';

---

-- Sic codes
UPDATE agent_definitions
SET default_config = REPLACE(
        default_config::text,
        '"sic_filter": ["75000"]',
        '"sic_filter": ["75", "749", "869"]'
                     )::jsonb
WHERE type = 'business-intel';

----

-- data from companies house

-- Migration: Create ch_vet_companies table
-- This is the local mirror of all Companies House companies with SIC 75000 (veterinary activities).
-- Populated by the ch_bulk_collect action, used for local matching against our businesses table.

CREATE TABLE IF NOT EXISTS business_intel.ch_vet_companies (
                                                               company_number      VARCHAR(10) PRIMARY KEY,
    company_name        TEXT NOT NULL,
    company_status      TEXT,
    company_type        TEXT,
    date_of_creation    DATE,
    date_of_cessation   DATE,
    sic_codes           TEXT[],
    registered_address  JSONB,
    postcode            TEXT,            -- extracted from address for indexing
    postcode_prefix     TEXT,            -- outward code (e.g. "BT74") for matching
    locality            TEXT,            -- town/city from address

-- Matching state
    matched_business_id UUID REFERENCES business_intel.businesses(id),
    matched_at          TIMESTAMPTZ,
    match_confidence    NUMERIC(3,2),
    match_method        TEXT,            -- postcode_name, name_only, manual, etc.

-- Detail fetch state
    details_fetched     BOOLEAN NOT NULL DEFAULT FALSE,
    details_fetched_at  TIMESTAMPTZ,

    -- Discovery state (for companies not in our businesses table)
    is_discovered       BOOLEAN NOT NULL DEFAULT FALSE,  -- true = not matched, potential new lead
    discovery_status    TEXT DEFAULT 'pending',           -- pending, website_found, verified, ignored

-- Timestamps
    collected_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Indexes for matching
CREATE INDEX IF NOT EXISTS idx_ch_vet_postcode_prefix
    ON business_intel.ch_vet_companies (postcode_prefix)
    WHERE company_status = 'active';

CREATE INDEX IF NOT EXISTS idx_ch_vet_name_trgm
    ON business_intel.ch_vet_companies USING gin (lower(company_name) gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_ch_vet_unmatched
    ON business_intel.ch_vet_companies (matched_business_id)
    WHERE matched_business_id IS NULL AND company_status = 'active';

CREATE INDEX IF NOT EXISTS idx_ch_vet_status
    ON business_intel.ch_vet_companies (company_status);

CREATE INDEX IF NOT EXISTS idx_ch_vet_discovery
    ON business_intel.ch_vet_companies (discovery_status)
    WHERE is_discovered = TRUE;

-- Enable trigram extension if not already (needed for fuzzy name matching)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

COMMENT ON TABLE business_intel.ch_vet_companies IS
    'Local mirror of all CH companies with SIC 75000. Populated by ch_bulk_collect, used for local matching.';

-- Add trigram index for name-only matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_ch_vet_name_trgm
    ON business_intel.ch_vet_companies
    USING gin (lower(company_name) gin_trgm_ops);

-- Also add a cleaned name column for faster matching
ALTER TABLE business_intel.ch_vet_companies
    ADD COLUMN IF NOT EXISTS company_name_cleaned TEXT;

UPDATE business_intel.ch_vet_companies
SET company_name_cleaned = lower(
        REGEXP_REPLACE(
                REGEXP_REPLACE(company_name, '\s+(LIMITED|LTD|LLP|PLC)\.?$', '', 'gi'),
                '\s+(GROUP|SURGERY|CENTRE|CENTER|CLINIC|HOSPITAL|PRACTICE)$', '', 'gi'
        )
                           );

CREATE INDEX IF NOT EXISTS idx_ch_vet_name_cleaned
    ON business_intel.ch_vet_companies (company_name_cleaned);

---
-- redo
-- Drop the GIN index and create GiST for distance-operator queries
DROP INDEX IF EXISTS business_intel.idx_ch_vet_name_trgm;

CREATE INDEX IF NOT EXISTS idx_ch_vet_name_trgm_gist
    ON business_intel.ch_vet_companies
    USING gist (lower(company_name) gist_trgm_ops);

-- Also index the cleaned name column
CREATE INDEX IF NOT EXISTS idx_ch_vet_name_cleaned_gist
    ON business_intel.ch_vet_companies
    USING gist (company_name_cleaned gist_trgm_ops);