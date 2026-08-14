-- 406 — tool-suggester: requires-backend eligibility gate (VMB-010's
-- suggester half; PLAN_2026-08-11_chat_box_as_framework_capability §2)
--
-- WHY. On 2026-08-14 chat-input-box became component_level='tool' (PLAN step
-- 1), which made it visible to tool-suggester's load_library_tools query — a
-- query with no backend check. A requires-backend tool suggested for a static
-- B2-hosted site deploys a widget that POSTs to an API the site does not
-- have. The gate: a requires-backend-tagged tool is only offered for a site
-- whose sites.deploy_config->'capabilities' contains 'backend' (live today:
-- webdesign.uk, noted.co.uk, relojistas.com — 3 of the fleet). For every
-- other site the candidate list is exactly what it was before the
-- reclassification, so the pair (flip + gate) is additive-and-inert
-- (RFC_022 shape: opt-in tag, unsafe side defaults OFF, consumers
-- enumerated below).
--
-- Census at write time: exactly 2 components carry the tag —
--   intent-probe    (section-level; VMB-010's generic-planner half stays
--                    UNBUILT — this migration does not touch section planning)
--   chat-input-box  (tool-level; gated here)
-- Consumers of load_library_tools's output: the suggest_tools LLM step of
-- tool-suggester, nothing else (output_field library_tools, read only inside
-- this workflow).
--
-- The params binding "input_data.site_id" / $1 is the same idiom the
-- adjacent load_existing_tools step already uses in this same workflow, so
-- the binding path is proven wherever this agent can run at all.
--
-- Config-only: no image dependency, live on apply. Known pre-existing defect
-- deliberately NOT touched here: LIMIT 30 truncates a 68-master library, so
-- tool-suggester silently never sees 38 tools (chat-input-box sorts 16th, so
-- it is unaffected). Flagged in webdesign_uk_build_service NOTES 2026-08-14.
--
-- ROLLBACK: 406_tool_suggester_requires_backend_gate_ROLLBACK.sql, or
-- restore the snapshot this file takes
-- (snapshot_agent note '406_tool_suggester_requires_backend_gate: pre-update').

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '406_tool_suggester_requires_backend_gate: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,load_library_tools,config,query}',
           to_jsonb('SELECT id::text, function, display_name, category, description FROM content_components WHERE component_level = ''tool'' AND forked_from IS NULL AND is_active = true AND html_template != '''' AND (NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') OR EXISTS (SELECT 1 FROM sites s WHERE s.id = $1 AND COALESCE(s.deploy_config->''capabilities'', ''[]''::jsonb) ? ''backend'')) ORDER BY display_name LIMIT 30'::text)
         ),
         '{workflow,steps,load_library_tools,config,params}',
         '["input_data.site_id"]'::jsonb
       ),
       updated_at = now()
 WHERE type = 'tool-suggester'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  q text;
  p jsonb;
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'tool-suggester'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '406: expected exactly 1 live tool-suggester row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}',
         default_config#>'{workflow,steps,load_library_tools,config,params}'
    INTO q, p
    FROM agent_definitions
   WHERE type = 'tool-suggester'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('requires-backend' in q) = 0
     OR position('deploy_config' in q) = 0 THEN
    RAISE EXCEPTION '406: load_library_tools query does not carry the gate: %', q;
  END IF;
  IF p IS NULL OR jsonb_array_length(p) <> 1
     OR p->>0 <> 'input_data.site_id' THEN
    RAISE EXCEPTION '406: load_library_tools params wrong: %', p;
  END IF;
END $$;

COMMIT;
