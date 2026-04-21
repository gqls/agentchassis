-- ============================================================================
-- 002_create_site_adoption_orchestrator.sql
-- ============================================================================
-- Creates a minimal wrapper orchestrator that spawns site-adoption-agent
-- as a separate K8s Job and passes input_data through.
--
-- Pattern copied verbatim from med-export-orchestrator / med-url-map-
-- orchestrator / med-price-scrape-orchestrator — "spawns a temporary pod
-- to do X" in its minimal form.
--
-- Workflow:
--   start → spawn_adopter → call_adopter → complete
--
--   spawn_adopter   : spawn_agent, creates K8s Job running site-adoption-agent
--   call_adopter    : call_agent, sends the work payload to the spawned pod
--                     and waits for completion
--   complete        : complete_workflow, returns the adoption result to caller
--
-- After this lands, the caller (trigger-adopt-site.sh) invokes
-- site-adoption-orchestrator instead of site-adoption-agent directly. The
-- orchestrator lives briefly in an agent-chassis pod; the actual adoption
-- runs in its own dedicated Job pod with a clean correlation-per-lifetime
-- log stream.
--
-- site-adoption-agent itself is UNCHANGED — workflow, Go code, all of
-- Phase 1 (target_url/destination_domain/source_url_field) stands.
-- ============================================================================

BEGIN;

-- Guard: don't double-create.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'site-adoption-orchestrator'
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'site-adoption-orchestrator already exists — aborting';
END IF;
END $$;

-- Guard: the child must exist and be live.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'site-adoption-agent'
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'site-adoption-agent not found — cannot create wrapper';
END IF;
END $$;

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    env_vars,
    version,
    delegation_preferences,
    agent_category,
    status,
    domain_tags,
    input_contract,
    output_contract,
    idle_timeout_seconds
) VALUES (
             'site-adoption-orchestrator',
             'Site Adoption Orchestrator',
             'Spawns a temporary pod to run site-adoption-agent. Thin wrapper that gives each adoption its own Job pod for clean per-lifetime logs. Passes input_data through unchanged (target_url, destination_domain, etc).',
             'orchestrator',
             $json$
                 {
      "workflow": {
        "steps": {
          "complete": {
            "action": "complete_workflow",
             "description": "Adoption orchestration complete",
             "config": {
              "output_fields": ["adoption_result"]
            }
          },
             "spawn_adopter": {
            "action": "spawn_agent",
             "config": {
              "role": "site_adopter",
             "agent_type": "site-adoption-agent"
            },
             "next_step": "call_adopter",
             "description": "Spawn a temporary pod for site adoption",
             "output_field": "adopter_spawn"
          },
             "call_adopter": {
            "action": "call_agent",
             "config": {
              "agent_type": "site-adoption-agent",
             "target_role": "site_adopter",
             "input_mapping": {
                "input_data": "input_data"
              },
             "timeout_seconds": 1800
            },
             "next_step": "complete",
             "description": "Send adoption work to spawned pod and wait",
             "output_field": "adoption_result"
          }
        },
             "start_step": "spawn_adopter"
      },
             "processing_mode": "orchestrator",
             "timeout_seconds": 2100
    }
    $json$::jsonb,
             true,                                                                       -- is_active
             'docker.io/aqls/agent-chassis',
             'v1.0.977',                                                                 -- matches current agent-chassis deployment
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator',
             'experimental',
             '["adoption", "orchestration", "wrapper"]'::jsonb,
             $json${
               "required": [],
               "optional": ["target_url", "destination_domain", "url", "domain"],
               "description": "Accepts the same input_data as site-adoption-agent and passes it through. Prefer target_url + destination_domain for the separated flow; url + domain work for legacy single-domain adoption."
             }$json$::jsonb,
             $json${
      "produces": {
        "adoption_result": "The full result from site-adoption-agent (specs written, pages created, work items, etc)"
      }
    }$json$::jsonb,
             180                                                                         -- idle_timeout_seconds — matches other orchestrators
         );

-- Verification
DO $$
DECLARE
v_id            UUID;
    v_start         TEXT;
    v_spawn_type    TEXT;
    v_call_type     TEXT;
BEGIN
SELECT
    id,
    default_config #>> '{workflow,start_step}',
        default_config #>> '{workflow,steps,spawn_adopter,config,agent_type}',
        default_config #>> '{workflow,steps,call_adopter,config,agent_type}'
INTO v_id, v_start, v_spawn_type, v_call_type
FROM agent_definitions
WHERE type = 'site-adoption-orchestrator'
  AND deleted_at IS NULL;

IF v_id IS NULL THEN
        RAISE EXCEPTION 'insert appeared to succeed but no row found';
END IF;
    IF v_start    IS DISTINCT FROM 'spawn_adopter'       THEN RAISE EXCEPTION 'start_step wrong: %', v_start;    END IF;
    IF v_spawn_type IS DISTINCT FROM 'site-adoption-agent' THEN RAISE EXCEPTION 'spawn agent_type wrong: %', v_spawn_type; END IF;
    IF v_call_type  IS DISTINCT FROM 'site-adoption-agent' THEN RAISE EXCEPTION 'call agent_type wrong: %',  v_call_type;  END IF;

    RAISE NOTICE 'site-adoption-orchestrator created: id=%', v_id;
    RAISE NOTICE '  start_step             = %', v_start;
    RAISE NOTICE '  spawn_adopter.agent_type = %', v_spawn_type;
    RAISE NOTICE '  call_adopter.agent_type  = %', v_call_type;
END $$;

COMMIT;

---
-- fix input mapping

-- ============================================================================
-- 003_fix_site_adoption_orchestrator_input_mapping.sql
-- ============================================================================
-- Bug: the wrapper's call_adopter step used `input_mapping: {"input_data":
-- "input_data"}` which wrapped the inner input_data under a new input_data
-- namespace, producing the shape `input_data.input_data.target_url` in the
-- spawned adopter's CollectedData. The adopter's workflow reads fields at
-- top-level paths (`input_data.target_url`, `input_data.destination_domain`)
-- and couldn't find them, so ensure_site_record silently wrote junk and
-- crawl_site crashed with "URL not found".
--
-- Diagnosis: the contract in 003_contracts_and_standards_v7.md ("Agent
-- Definition SQL Conventions") makes clear that dispatch-loop-style
-- `input_data.spec.*` is a deliberate convention for work items — fields
-- sit under a named nested path, not a generic double-wrap. The wrapper
-- orchestrator has no spec column; it's passing ordinary named fields.
-- The right pattern is per-field mapping.
--
-- Fix: replace the namespace-wrap mapping with an explicit per-field
-- mapping that places each field directly at the top of the child's
-- input_data. All four known keys (target_url, destination_domain, url,
-- domain) are mapped so both the separated and legacy flows work.
-- ============================================================================

BEGIN;

-- Sanity
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'site-adoption-orchestrator' AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'site-adoption-orchestrator not found — aborting';
END IF;
END $$;

-- Replace the input_mapping value with explicit per-field mapping.
UPDATE agent_definitions
SET
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,call_adopter,config,input_mapping}',
            $json${
              "target_url":         "input_data.target_url",
              "destination_domain": "input_data.destination_domain",
              "url":                "input_data.url",
              "domain":             "input_data.domain"
            }$json$::jsonb,
            false   -- path must exist (set by 002_create_site_adoption_orchestrator.sql)
                     ),
    updated_at = NOW()
WHERE type = 'site-adoption-orchestrator'
  AND deleted_at IS NULL;

-- Verify the new mapping is in place.
DO $$
DECLARE
v_target_url      TEXT;
    v_destination     TEXT;
    v_url             TEXT;
    v_domain          TEXT;
BEGIN
SELECT
    default_config #>> '{workflow,steps,call_adopter,config,input_mapping,target_url}',
        default_config #>> '{workflow,steps,call_adopter,config,input_mapping,destination_domain}',
        default_config #>> '{workflow,steps,call_adopter,config,input_mapping,url}',
        default_config #>> '{workflow,steps,call_adopter,config,input_mapping,domain}'
INTO v_target_url, v_destination, v_url, v_domain
FROM agent_definitions
WHERE type = 'site-adoption-orchestrator' AND deleted_at IS NULL;

IF v_target_url  IS DISTINCT FROM 'input_data.target_url'         THEN RAISE EXCEPTION 'target_url mapping wrong: %', v_target_url; END IF;
    IF v_destination IS DISTINCT FROM 'input_data.destination_domain' THEN RAISE EXCEPTION 'destination_domain mapping wrong: %', v_destination; END IF;
    IF v_url         IS DISTINCT FROM 'input_data.url'                THEN RAISE EXCEPTION 'url mapping wrong: %', v_url; END IF;
    IF v_domain      IS DISTINCT FROM 'input_data.domain'             THEN RAISE EXCEPTION 'domain mapping wrong: %', v_domain; END IF;

    RAISE NOTICE 'site-adoption-orchestrator call_adopter.input_mapping fixed:';
    RAISE NOTICE '  target_url         -> %', v_target_url;
    RAISE NOTICE '  destination_domain -> %', v_destination;
    RAISE NOTICE '  url                -> %', v_url;
    RAISE NOTICE '  domain             -> %', v_domain;
END $$;

COMMIT;

