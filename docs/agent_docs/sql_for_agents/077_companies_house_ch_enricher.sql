-- =========================================================================
-- 2. AGENT DEFINITION
-- =========================================================================
-- The ch-enricher loads a batch of verified businesses that haven't been
-- enriched yet, and for each one: searches Companies House, picks the
-- best match, fetches profile + officers + PSC, stores the data.
--
-- Runs on the vet-intel pod (same topic).

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
