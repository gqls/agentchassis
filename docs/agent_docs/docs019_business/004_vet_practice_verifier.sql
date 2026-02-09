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

--

-- fixing llm path into config.ai-service

-- Fix vet-practice-verifier: move ai_service into extract_and_reconcile step config
--
-- Problem: ai_service is at root level of default_config, but execute_llm_prompt
-- looks for it inside the step's config block (like the briefing agent does).
--
-- Fix: Add ai_service inside extract_and_reconcile.config and remove from root.

-- Step 1: Add ai_service into the step config
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,extract_and_reconcile,config,ai_service}',
        '{
            "model": "claude-haiku-4-5",
            "provider": "anthropic",
            "api_key_env_var": "ANTHROPIC_API_KEY"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-practice-verifier';

-- Step 2: Remove ai_service from root level
UPDATE agent_definitions
SET default_config = default_config - 'ai_service',
    updated_at = NOW()
WHERE type = 'vet-practice-verifier';

-- Verify: ai_service should be inside the step config, not at root
SELECT
    type,
    default_config->'ai_service' as root_ai_service,
    default_config->'workflow'->'steps'->'extract_and_reconcile'->'config'->'ai_service' as step_ai_service
FROM agent_definitions
WHERE type = 'vet-practice-verifier';


--

-- more GO template formatting fixes

-- Fix vet-practice-verifier: prompt_template needs leading dots for Go template syntax
--
-- Problem: Template uses {{business_record.business.name}} but Go's text/template
-- interprets that as a function call. Needs {{.business_record.business.name}} to
-- traverse the data map. Same issue as query_template needing the dot.
--
-- The briefing agent's working prompt uses {{.input_data.domain}} - with the dot.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,extract_and_reconcile,config,prompt_template}',
        to_jsonb(
                'You are a data extraction specialist for UK veterinary practices.

        CURRENT RECORD:
        Name: {{.business_record.business.name}}
        Postcode: {{.business_record.business.postcode}}
        Town: {{.business_record.business.town}}
        Website: {{.business_record.business.website_url}}
        Group: {{.business_record.business.group_name}}

        SCRAPED WEBSITE CONTENT:
        {{.scraped_data.content}}

        SEARCH RESULTS:
        {{.search_results}}

        Extract and return a JSON object with these sections:

        1. business - updated/confirmed fields:
           - name, address_line1, address_line2, town, county, postcode
           - phone, email, website_url
           - group_name, business_type

        2. vet_details - practice-specific:
           - species_treated (array of strings)
           - emergency_service (boolean)
           - out_of_hours_provider (string or null)
           - accepting_new_clients (boolean or null if unknown)
           - accreditations (array)
           - num_vets, num_nurses (integers or null)
           - head_vet_name (string or null)
           - has_own_lab, has_imaging, has_surgical_suite (booleans or null)
           - parking_available, wheelchair_accessible (booleans or null)

        3. prices - array of objects, each with:
           - service_category: one of consultation, vaccination, surgery, prescription, dental, diagnostic, other
           - service_name: the specific service
           - price_gbp: numeric price
           - price_qualifier: fixed, from, or approximately

        4. confidence_score - 0.0 to 1.0, how confident you are in the data quality
        5. extraction_notes - brief notes on data quality, conflicts, missing data

        Only include fields where you have actual data. Use null for unknown values. Do not invent or estimate prices.'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'vet-practice-verifier';

-- Verify the dots are in place
SELECT
    substring(
            default_config->'workflow'->'steps'->'extract_and_reconcile'->'config'->>'prompt_template',
        1, 200
    ) as prompt_preview
FROM agent_definitions
WHERE type = 'vet-practice-verifier';


-- content path fix

-- Fix vet-practice-verifier: scraped_data path in prompt template
--
-- Problem: Prompt uses {{.scraped_data.content}} but the actual structure is:
--   scraped_data.response.data.markdown_content
--
-- The scrape adapter wraps its response in {response: {data: {...}}}
-- and the content fields are markdown_content, html_content, raw_html etc.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,extract_and_reconcile,config,prompt_template}',
        to_jsonb(
                'You are a data extraction specialist for UK veterinary practices.

        CURRENT RECORD:
        Name: {{.business_record.business.name}}
        Postcode: {{.business_record.business.postcode}}
        Town: {{.business_record.business.town}}
        Website: {{.business_record.business.website_url}}
        Group: {{.business_record.business.group_name}}

        SCRAPED WEBSITE CONTENT:
        {{.scraped_data.response.data.markdown_content}}

        SEARCH RESULTS:
        {{.search_results}}

        Extract and return a JSON object with these sections:

        1. business - updated/confirmed fields:
           - name, address_line1, address_line2, town, county, postcode
           - phone, email, website_url
           - group_name, business_type

        2. vet_details - practice-specific:
           - species_treated (array of strings)
           - emergency_service (boolean)
           - out_of_hours_provider (string or null)
           - accepting_new_clients (boolean or null if unknown)
           - accreditations (array)
           - num_vets, num_nurses (integers or null)
           - head_vet_name (string or null)
           - has_own_lab, has_imaging, has_surgical_suite (booleans or null)
           - parking_available, wheelchair_accessible (booleans or null)

        3. prices - array of objects, each with:
           - service_category: one of consultation, vaccination, surgery, prescription, dental, diagnostic, other
           - service_name: the specific service
           - price_gbp: numeric price
           - price_qualifier: fixed, from, or approximately

        4. confidence_score - 0.0 to 1.0, how confident you are in the data quality
        5. extraction_notes - brief notes on data quality, conflicts, missing data

        Only include fields where you have actual data. Use null for unknown values. Do not invent or estimate prices.'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'vet-practice-verifier';

-- Verify
SELECT
    substring(
            default_config->'workflow'->'steps'->'extract_and_reconcile'->'config'->>'prompt_template',
        1, 300
    ) as prompt_preview
FROM agent_definitions
WHERE type = 'vet-practice-verifier';
