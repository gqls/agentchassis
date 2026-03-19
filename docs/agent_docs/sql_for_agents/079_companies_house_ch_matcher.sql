-- ============================================================================
-- CH Local Matching — Agent Definition & Scheduled Task
-- ============================================================================
-- Matches verified businesses against the local ch_vet_companies table.
-- No API calls — pure SQL + Go scoring. Safe to re-run.
-- Runs on the business-intel pod, separate from collection and enrichment.
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources,
    topics, health_config,
    agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
             gen_random_uuid(),
             'ch-matcher',
             'CH Local Matcher',
             'Matches verified businesses against the local ch_vet_companies mirror table using postcode + name similarity. No API calls. Updates ch_vet_companies.matched_business_id for confirmed matches.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "match",
                     "steps": {
                         "match": {
                             "action": "ch_local_match",
                             "config": {
                                 "batch_size": 3000,
                                 "threshold": 0.40,
                                 "rematch": false
                             },
                             "next_step": "report",
                             "description": "Score all unmatched businesses against ch_vet_companies by postcode + name",
                             "output_field": "match_result"
                         },
                         "report": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT COUNT(*) as total_ch_companies, COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) as matched, COUNT(*) FILTER (WHERE matched_business_id IS NULL) as unmatched, ROUND(100.0 * COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) / NULLIF(COUNT(*), 0), 1) as match_pct FROM business_intel.ch_vet_companies WHERE company_status = ''active''",
                                 "output_format": "object"
                             },
                             "next_step": "notify_scheduler",
                             "description": "Report overall match stats",
                             "output_field": "stats"
                         },
                         "notify_scheduler": {
                             "action": "query_database",
                             "config": {
                                 "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''ch-local-match''",
                                 "output_format": "object"
                             },
                             "next_step": "complete",
                             "description": "Tell scheduler this execution finished"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["match_result", "stats"]
                             },
                             "description": "Matching complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600
             }'::jsonb,
             true,
             '["companies-house", "matching", "local"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.890',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "matching", "veterinary"]',
             '{"required": [], "optional": ["batch_size", "threshold", "rematch"]}'::jsonb,
             '{"produces": {"match_result": "object - total_processed, total_matched, total_no_match", "stats": "object - total_ch_companies, matched, unmatched, match_pct"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              image_tag = EXCLUDED.image_tag,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();

-- Scheduled task — disabled by default. Enable after first manual run.
-- Can run frequently since it's all local DB, no API cost.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled, timeout_seconds,
    pre_query
) VALUES (
             'ch-local-match',
             'Match verified businesses against local CH vet companies mirror. No API calls.',
             86400,     -- daily
             'ch-matcher',
             'system.agent.business-intel.requests',
             'ch-matching',
             1,
             false,
             600,       -- 10 min timeout
             'SELECT COUNT(*) as unmatched FROM business_intel.businesses b JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id LEFT JOIN business_intel.ch_vet_companies ch ON ch.matched_business_id = b.id WHERE bv.slug = ''veterinary'' AND b.verification_status = ''verified'' AND ch.company_number IS NULL HAVING COUNT(*) > 0'
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              pre_query = EXCLUDED.pre_query,
                              updated_at = NOW();

-- ============================================================================
-- Verification queries after matching:
-- ============================================================================
--
-- Overall stats:
--   SELECT COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) as matched,
--          COUNT(*) FILTER (WHERE matched_business_id IS NULL) as unmatched
--   FROM business_intel.ch_vet_companies WHERE company_status = 'active';
--
-- Sample matches:
--   SELECT b.name, ch.company_name, ch.match_confidence, ch.match_method,
--          ch.postcode, b.postcode
--   FROM business_intel.ch_vet_companies ch
--   JOIN business_intel.businesses b ON b.id = ch.matched_business_id
--   ORDER BY ch.match_confidence DESC
--   LIMIT 20;
--
-- Low-confidence matches to review:
--   SELECT b.name, ch.company_name, ch.match_confidence, ch.postcode, b.postcode
--   FROM business_intel.ch_vet_companies ch
--   JOIN business_intel.businesses b ON b.id = ch.matched_business_id
--   WHERE ch.match_confidence < 0.5
--   ORDER BY ch.match_confidence ASC;
--
-- Discovery candidates (CH companies not matched to any business):
--   SELECT company_name, postcode, locality, company_number
--   FROM business_intel.ch_vet_companies
--   WHERE matched_business_id IS NULL
--     AND company_status = 'active'
--   ORDER BY company_name
--   LIMIT 50;
-- ============================================================================

---

-- if names match then that's good

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,match,config,name_only_threshold}',
        '0.55'
                     )
WHERE type = 'ch-matcher';

---
--

-- ============================================================================
-- CH Local Matching — Agent Definition & Scheduled Task
-- ============================================================================
-- Matches verified businesses against the local ch_vet_companies table.
-- No API calls — pure SQL + Go scoring. Safe to re-run.
-- Runs on the business-intel pod, separate from collection and enrichment.
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources,
    topics, health_config,
    agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
             gen_random_uuid(),
             'ch-matcher',
             'CH Local Matcher',
             'Matches verified businesses against the local ch_vet_companies mirror table using postcode + name similarity. No API calls. Updates ch_vet_companies.matched_business_id for confirmed matches.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "match",
                     "steps": {
                         "match": {
                             "action": "ch_local_match",
                             "config": {
                                 "batch_size": 3000,
                                 "threshold": 0.40,
                                 "name_only_threshold": 0.55,
                                 "rematch": false
                             },
                             "next_step": "report",
                             "description": "Score all unmatched businesses against ch_vet_companies by postcode + name",
                             "output_field": "match_result"
                         },
                         "report": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT COUNT(*) as total_ch_companies, COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) as matched, COUNT(*) FILTER (WHERE matched_business_id IS NULL) as unmatched, ROUND(100.0 * COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) / NULLIF(COUNT(*), 0), 1) as match_pct FROM business_intel.ch_vet_companies WHERE company_status = ''active''",
                                 "output_format": "object"
                             },
                             "next_step": "notify_scheduler",
                             "description": "Report overall match stats",
                             "output_field": "stats"
                         },
                         "notify_scheduler": {
                             "action": "query_database",
                             "config": {
                                 "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''ch-local-match''",
                                 "output_format": "object"
                             },
                             "next_step": "complete",
                             "description": "Tell scheduler this execution finished"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["match_result", "stats"]
                             },
                             "description": "Matching complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600
             }'::jsonb,
             true,
             '["companies-house", "matching", "local"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.890',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "matching", "veterinary"]',
             '{"required": [], "optional": ["batch_size", "threshold", "rematch"]}'::jsonb,
             '{"produces": {"match_result": "object - total_processed, total_matched, total_no_match", "stats": "object - total_ch_companies, matched, unmatched, match_pct"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              image_tag = EXCLUDED.image_tag,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();

-- Scheduled task — disabled by default. Enable after first manual run.
-- Can run frequently since it's all local DB, no API cost.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled, timeout_seconds,
    pre_query
) VALUES (
             'ch-local-match',
             'Match verified businesses against local CH vet companies mirror. No API calls.',
             86400,     -- daily
             'ch-matcher',
             'system.agent.business-intel.requests',
             'ch-matching',
             1,
             false,
             600,       -- 10 min timeout
             'SELECT COUNT(*) as unmatched FROM business_intel.businesses b JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id LEFT JOIN business_intel.ch_vet_companies ch ON ch.matched_business_id = b.id WHERE bv.slug = ''veterinary'' AND b.verification_status = ''verified'' AND ch.company_number IS NULL HAVING COUNT(*) > 0'
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              pre_query = EXCLUDED.pre_query,
                              updated_at = NOW();

-- ============================================================================
-- Verification queries after matching:
-- ============================================================================
--
-- Overall stats:
--   SELECT COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) as matched,
--          COUNT(*) FILTER (WHERE matched_business_id IS NULL) as unmatched
--   FROM business_intel.ch_vet_companies WHERE company_status = 'active';
--
-- Sample matches:
--   SELECT b.name, ch.company_name, ch.match_confidence, ch.match_method,
--          ch.postcode, b.postcode
--   FROM business_intel.ch_vet_companies ch
--   JOIN business_intel.businesses b ON b.id = ch.matched_business_id
--   ORDER BY ch.match_confidence DESC
--   LIMIT 20;
--
-- Low-confidence matches to review:
--   SELECT b.name, ch.company_name, ch.match_confidence, ch.postcode, b.postcode
--   FROM business_intel.ch_vet_companies ch
--   JOIN business_intel.businesses b ON b.id = ch.matched_business_id
--   WHERE ch.match_confidence < 0.5
--   ORDER BY ch.match_confidence ASC;
--
-- Discovery candidates (CH companies not matched to any business):
--   SELECT company_name, postcode, locality, company_number
--   FROM business_intel.ch_vet_companies
--   WHERE matched_business_id IS NULL
--     AND company_status = 'active'
--   ORDER BY company_name
--   LIMIT 50;
-- ============================================================================

---
-- remove sql from workflow
-- Update collector workflow
UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "collect",
        "steps": {
            "collect": {
                "action": "ch_bulk_collect",
                "config": {
                    "sic_code": "75000",
                    "company_status": "active",
                    "page_size": 100,
                    "delay_ms": 2000,
                    "task_name": "ch-vet-collect"
                },
                "next_step": "complete",
                "description": "Paginate through CH advanced search, store all SIC 75000 companies, notify scheduler internally",
                "output_field": "collection_result"
            },
            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["collection_result"]
                },
                "description": "Collection complete"
            }
        }
    },
    "processing_mode": "orchestrator",
    "timeout_seconds": 1800
}'::jsonb,
updated_at = NOW()
WHERE type = 'ch-collector';

-- Update matcher workflow
UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "match",
        "steps": {
            "match": {
                "action": "ch_local_match",
                "config": {
                    "batch_size": 3000,
                    "threshold": 0.40,
                    "name_only_threshold": 0.55,
                    "rematch": false,
                    "task_name": "ch-local-match"
                },
                "next_step": "complete",
                "description": "Score all unmatched businesses against ch_vet_companies. Pass 1: postcode+name. Pass 2: trigram name-only. Notifies scheduler internally.",
                "output_field": "match_result"
            },
            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["match_result"]
                },
                "description": "Matching complete"
            }
        }
    },
    "processing_mode": "orchestrator",
    "timeout_seconds": 600
}'::jsonb,
updated_at = NOW()
WHERE type = 'ch-matcher';

-- same prob

-- ============================================================================
-- CH Local Matching — Agent Definition & Scheduled Task
-- ============================================================================
-- Matches verified businesses against the local ch_vet_companies table.
-- No API calls — pure SQL + Go scoring. Safe to re-run.
-- Runs on the business-intel pod, separate from collection and enrichment.
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources,
    topics, health_config,
    agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
             gen_random_uuid(),
             'ch-matcher',
             'CH Local Matcher',
             'Matches verified businesses against the local ch_vet_companies mirror table using postcode + name similarity. No API calls. Updates ch_vet_companies.matched_business_id for confirmed matches.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "match",
                     "steps": {
                         "match": {
                             "action": "ch_local_match",
                             "config": {
                                 "batch_size": 3000,
                                 "threshold": 0.40,
                                 "name_only_threshold": 0.55,
                                 "rematch": false,
                                 "task_name": "ch-local-match"
                             },
                             "next_step": "complete",
                             "description": "Score all unmatched businesses against ch_vet_companies. Pass 1: postcode+name. Pass 2: trigram name-only. Notifies scheduler internally.",
                             "output_field": "match_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["match_result"]
                             },
                             "description": "Matching complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600
             }'::jsonb,
             true,
             '["companies-house", "matching", "local"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.890',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "matching", "veterinary"]',
             '{"required": [], "optional": ["batch_size", "threshold", "rematch"]}'::jsonb,
             '{"produces": {"match_result": "object - total_processed, total_matched, total_no_match", "stats": "object - total_ch_companies, matched, unmatched, match_pct"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              image_tag = EXCLUDED.image_tag,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();

-- Scheduled task — disabled by default. Enable after first manual run.
-- Can run frequently since it's all local DB, no API cost.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled, timeout_seconds,
    pre_query
) VALUES (
             'ch-local-match',
             'Match verified businesses against local CH vet companies mirror. No API calls.',
             86400,     -- daily
             'ch-matcher',
             'system.agent.business-intel.requests',
             'ch-matching',
             1,
             false,
             600,       -- 10 min timeout
             'SELECT COUNT(*) as unmatched FROM business_intel.businesses b JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id LEFT JOIN business_intel.ch_vet_companies ch ON ch.matched_business_id = b.id WHERE bv.slug = ''veterinary'' AND b.verification_status = ''verified'' AND ch.company_number IS NULL HAVING COUNT(*) > 0'
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              pre_query = EXCLUDED.pre_query,
                              updated_at = NOW();

-- ============================================================================
-- Verification queries after matching:
-- ============================================================================
--
-- Overall stats:
--   SELECT COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) as matched,
--          COUNT(*) FILTER (WHERE matched_business_id IS NULL) as unmatched
--   FROM business_intel.ch_vet_companies WHERE company_status = 'active';
--
-- Sample matches:
--   SELECT b.name, ch.company_name, ch.match_confidence, ch.match_method,
--          ch.postcode, b.postcode
--   FROM business_intel.ch_vet_companies ch
--   JOIN business_intel.businesses b ON b.id = ch.matched_business_id
--   ORDER BY ch.match_confidence DESC
--   LIMIT 20;
--
-- Low-confidence matches to review:
--   SELECT b.name, ch.company_name, ch.match_confidence, ch.postcode, b.postcode
--   FROM business_intel.ch_vet_companies ch
--   JOIN business_intel.businesses b ON b.id = ch.matched_business_id
--   WHERE ch.match_confidence < 0.5
--   ORDER BY ch.match_confidence ASC;
--
-- Discovery candidates (CH companies not matched to any business):
--   SELECT company_name, postcode, locality, company_number
--   FROM business_intel.ch_vet_companies
--   WHERE matched_business_id IS NULL
--     AND company_status = 'active'
--   ORDER BY company_name
--   LIMIT 50;
-- ============================================================================

---
-- closer match to company name

-- ============================================================================
-- CH Local Matching — Agent Definition & Scheduled Task
-- ============================================================================
-- Matches verified businesses against the local ch_vet_companies table.
-- No API calls — pure SQL + Go scoring. Safe to re-run.
-- Runs on the business-intel pod, separate from collection and enrichment.
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources,
    topics, health_config,
    agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
             gen_random_uuid(),
             'ch-matcher',
             'CH Local Matcher',
             'Matches verified businesses against the local ch_vet_companies mirror table using postcode + name similarity. No API calls. Updates ch_vet_companies.matched_business_id for confirmed matches.',
             'data-driven',
             '{
                 "workflow": {
                     "start_step": "match",
                     "steps": {
                         "match": {
                             "action": "ch_local_match",
                             "config": {
                                 "batch_size": 3000,
                                 "threshold": 0.40,
                                 "name_only_threshold": 0.70,
                                 "rematch": false,
                                 "task_name": "ch-local-match"
                             },
                             "next_step": "complete",
                             "description": "Score all unmatched businesses against ch_vet_companies. Pass 1: postcode+name. Pass 2: trigram name-only. Notifies scheduler internally.",
                             "output_field": "match_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["match_result"]
                             },
                             "description": "Matching complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600
             }'::jsonb,
             true,
             '["companies-house", "matching", "local"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.890',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             'specialist',
             'experimental',
             '["companies-house", "matching", "veterinary"]',
             '{"required": [], "optional": ["batch_size", "threshold", "rematch"]}'::jsonb,
             '{"produces": {"match_result": "object - total_processed, total_matched, total_no_match", "stats": "object - total_ch_companies, matched, unmatched, match_pct"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              image_tag = EXCLUDED.image_tag,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();

-- Scheduled task — disabled by default. Enable after first manual run.
-- Can run frequently since it's all local DB, no API cost.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    concurrency_group, max_concurrent,
    enabled, timeout_seconds,
    pre_query
) VALUES (
             'ch-local-match',
             'Match verified businesses against local CH vet companies mirror. No API calls.',
             86400,     -- daily
             'ch-matcher',
             'system.agent.business-intel.requests',
             'ch-matching',
             1,
             false,
             600,       -- 10 min timeout
             'SELECT COUNT(*) as unmatched FROM business_intel.businesses b JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id LEFT JOIN business_intel.ch_vet_companies ch ON ch.matched_business_id = b.id WHERE bv.slug = ''veterinary'' AND b.verification_status = ''verified'' AND ch.company_number IS NULL HAVING COUNT(*) > 0'
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              pre_query = EXCLUDED.pre_query,
                              updated_at = NOW();

-- ============================================================================
-- Verification queries after matching:
-- ============================================================================
--
-- Overall stats:
--   SELECT COUNT(*) FILTER (WHERE matched_business_id IS NOT NULL) as matched,
--          COUNT(*) FILTER (WHERE matched_business_id IS NULL) as unmatched
--   FROM business_intel.ch_vet_companies WHERE company_status = 'active';
--
-- Sample matches:
--   SELECT b.name, ch.company_name, ch.match_confidence, ch.match_method,
--          ch.postcode, b.postcode
--   FROM business_intel.ch_vet_companies ch
--   JOIN business_intel.businesses b ON b.id = ch.matched_business_id
--   ORDER BY ch.match_confidence DESC
--   LIMIT 20;
--
-- Low-confidence matches to review:
--   SELECT b.name, ch.company_name, ch.match_confidence, ch.postcode, b.postcode
--   FROM business_intel.ch_vet_companies ch
--   JOIN business_intel.businesses b ON b.id = ch.matched_business_id
--   WHERE ch.match_confidence < 0.5
--   ORDER BY ch.match_confidence ASC;
--
-- Discovery candidates (CH companies not matched to any business):
--   SELECT company_name, postcode, locality, company_number
--   FROM business_intel.ch_vet_companies
--   WHERE matched_business_id IS NULL
--     AND company_status = 'active'
--   ORDER BY company_name
--   LIMIT 50;
-- ============================================================================