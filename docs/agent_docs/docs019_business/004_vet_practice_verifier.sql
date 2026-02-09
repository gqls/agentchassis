-- =============================================
-- vet-practice-verifier
-- =============================================
-- Single-practice verification workflow:
--   load → search → scrape → extract (LLM) → store
-- Receives: { "business_id": "uuid" }
-- Can be called standalone or spawned by vet-batch-processor

INSERT INTO agent_definitions (
    type, display_name, description, category,
    is_active, status, agent_category,
    capabilities, domain_tags,
    default_config,
    input_contract, output_contract
) VALUES (
             'vet-practice-verifier',
             'Vet Practice Verifier',
             'Verifies and enriches a single veterinary practice record by searching the web, scraping the practice website, and extracting structured data via LLM.',
             'data-driven',
             TRUE, 'experimental', 'specialist',
             '["data-collection", "web-scraping", "veterinary", "verification"]'::jsonb,
             '["veterinary", "business-intelligence", "data-collection"]'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 300,
                 "ai_service": {
                     "provider": "anthropic",
                     "model": "claude-haiku-4-5",
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "workflow": {
                     "start_step": "load_business",
                     "steps": {
                         "load_business": {
                             "action": "load_business_record",
                             "description": "Load current business data from DB",
                             "config": {
                                 "input_fields": ["business_id"],
                                 "include_vet_details": true,
                                 "include_prices": true
                             },
                             "output_field": "business_record",
                             "next_step": "search_practice"
                         },

                         "search_practice": {
                             "action": "web_search",
                             "description": "Search for the practice website and online presence",
                             "config": {
                                 "query_template": "{{business_record.business.name}} {{business_record.business.postcode}} veterinary practice",
                                 "num_results": 5
                             },
                             "output_field": "search_results",
                             "next_step": "scrape_website"
                         },

                         "scrape_website": {
                             "action": "scrape_web",
                             "description": "Scrape the practice website for details and prices",
                             "config": {
                                 "url_field": "business_record.business.website_url",
                                 "fallback_url_field": "search_results.results.0.url",
                                 "extract_mode": "text",
                                 "max_pages": 3,
                                 "follow_links": ["fees", "prices", "about", "team", "contact", "services"]
                             },
                             "output_field": "scraped_data",
                             "next_step": "extract_and_reconcile"
                         },

                         "extract_and_reconcile": {
                             "action": "execute_llm_prompt",
                             "description": "LLM extracts structured data from scraped content and reconciles with existing record",
                             "config": {
                                 "input_fields": ["business_record", "scraped_data", "search_results"],
                                 "response_format": "json",
                                 "prompt_template": "You are a data extraction specialist for UK veterinary practices.\n\nCURRENT RECORD:\nName: {{business_record.business.name}}\nPostcode: {{business_record.business.postcode}}\nTown: {{business_record.business.town}}\nWebsite: {{business_record.business.website_url}}\nGroup: {{business_record.business.group_name}}\n\nSCRAPED WEBSITE CONTENT:\n{{scraped_data.content}}\n\nSEARCH RESULTS:\n{{search_results}}\n\nExtract and return a JSON object with these sections:\n\n1. business - updated/confirmed fields:\n   - name, address_line1, address_line2, town, county, postcode\n   - phone, email, website_url\n   - group_name, business_type\n\n2. vet_details - practice-specific:\n   - species_treated (array of strings)\n   - emergency_service (boolean)\n   - out_of_hours_provider (string or null)\n   - accepting_new_clients (boolean or null if unknown)\n   - accreditations (array)\n   - num_vets, num_nurses (integers or null)\n   - head_vet_name (string or null)\n   - has_own_lab, has_imaging, has_surgical_suite (booleans or null)\n   - parking_available, wheelchair_accessible (booleans or null)\n\n3. prices - array of objects, each with:\n   - service_category: one of consultation, vaccination, surgery, prescription, dental, diagnostic, other\n   - service_name: the specific service\n   - price_gbp: numeric price\n   - price_qualifier: fixed, from, or approximately\n\n4. confidence_score - 0.0 to 1.0, how confident you are in the data quality\n5. extraction_notes - brief notes on data quality, conflicts, missing data\n\nOnly include fields where you have actual data. Use null for unknown values. Do not invent or estimate prices."
                             },
                             "output_field": "verification_result",
                             "next_step": "store_results"
                         },

                         "store_results": {
                             "action": "store_business_verification",
                             "description": "Write verified data back to the database",
                             "config": {
                                 "input_fields": ["business_id", "verification_result"]
                             },
                             "output_field": "store_result",
                             "next_step": "complete"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Verification complete",
                             "config": {
                                 "output_fields": ["business_record", "verification_result", "store_result"]
                             }
                         }
                     }
                 }
             }'::jsonb,
             '{
                 "required": ["business_id"],
                 "optional": ["task_id"]
             }'::jsonb,
             '{
                 "produces": {
                     "business_record": "object - the loaded business data",
                     "verification_result": "object - extracted and reconciled data",
                     "store_result": "object - what was written to DB"
                 }
             }'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    capabilities = EXCLUDED.capabilities,
    domain_tags = EXCLUDED.domain_tags,
    status = EXCLUDED.status,
    updated_at = NOW();

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,search_practice}',
        '{
            "action": "web_search",
            "input_map": {
                "query": "{{.business_record.business.name}} {{.business_record.business.postcode}} veterinary practice"
            },
            "config": {
                "num_results": 5
            },
            "next_step": "scrape_website",
            "description": "Search for the practice website and online presence",
            "output_field": "search_results"
        }'::jsonb
                    )
WHERE type = 'vet-practice-verifier';

-- fix the above which is no longer needed
-- Fix vet-practice-verifier: search_practice step
--
-- Problem: The step used "input_map" at the step level, which is not recognized
-- by the chassis. The web_search action's extractSearchQuery function looks for
-- query in config (literal), query_from (path), query_template (Go template), etc.
--
-- Fix: Move the template into config as "query_template" so the web_search action
-- can resolve it against collected data.
--
-- NOTE: This requires the corresponding Go code change that adds query_template
-- support to extractSearchQuery in web_search_action.go

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,search_practice}',
        '{
            "action": "web_search",
            "config": {
                "num_results": 5,
                "query_template": "{{.business_record.business.name}} {{.business_record.business.postcode}} veterinary practice"
            },
            "next_step": "scrape_website",
            "description": "Search for the practice website and online presence",
            "output_field": "search_results"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-practice-verifier';

-- Verify the change
SELECT
    type,
    default_config->'workflow'->'steps'->'search_practice' as search_practice_step
FROM agent_definitions
WHERE type = 'vet-practice-verifier';


--

-- search for town rather than postcode

-- Fix vet-practice-verifier: search_practice step


UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,search_practice}',
        '{
            "action": "web_search",
            "config": {
                "num_results": 5,
                "query_template": "{{.business_record.business.name}} {{.business_record.business.town}} veterinary practice"
            },
            "next_step": "scrape_website",
            "description": "Search for the practice website and online presence",
            "output_field": "search_results"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-practice-verifier';

-- Verify
SELECT
    type,
    default_config->'workflow'->'steps'->'search_practice' as search_practice_step
FROM agent_definitions
WHERE type = 'vet-practice-verifier';