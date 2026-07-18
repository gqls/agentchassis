-- 168_component_writers_max_tokens.sql — stop truncating whole components.
--
-- INCIDENT 2026-07-18 (bugs_open/012): tool-improver rewrote
-- tool-loot-table-balancer, its improve_tool step returned output_tokens=8000
-- against max_tokens=8000 (hit the ceiling exactly), and the TRUNCATED result
-- was saved straight over content_components.html_template — reducing a working
-- 10,272-char tool to a 1,253-char fragment of CSS ending mid-declaration: no
-- <script>, no <div>, no <fieldset>. The durable source was destroyed; the live
-- page survived only because the render had not yet propagated it.
--
-- Systemic shape: the agents that emit a WHOLE component were sized
-- inconsistently. tool-recreation-handler.recreate_tool — which does exactly
-- this job — was correctly given 64000. improve_tool and generate_tool_html do
-- the same job on 8000. (This tool's own BIRTH used 6094/8000: the generator was
-- already one slightly larger tool away from shipping a truncated component.)
--
-- This migration removes the immediate exposure. It is NOT the real fix: the
-- structural fix is a completeness guard that REFUSES to save a component whose
-- markup/script structure vanished (bugs_open/012 candidate a) — a bigger
-- ceiling only makes truncation rarer, never impossible. Same lesson as the
-- article-body truncation arc (max_tokens 2000->8000, 2026-07-15): raising the
-- limit treats the symptom; refusing to persist a mangled artifact treats the bug.
--
-- Applied out of band (psql -f + ledger row same sitting, per bugs_open/007).

BEGIN;

SELECT snapshot_agent('tool-improver',  '168_component_writers_max_tokens: pre-update');
SELECT snapshot_agent('tool-generator', '168_component_writers_max_tokens: pre-update');

DO $$
DECLARE
  n int := 0;
BEGIN
  -- improve_tool: rewrites the whole component (the step that truncated).
  UPDATE agent_definitions
  SET default_config = jsonb_set(default_config,
        '{workflow,steps,improve_tool,config,ai_service,max_tokens}', '32000'::jsonb)
  WHERE type='tool-improver' AND is_active
    AND default_config #>> '{workflow,steps,improve_tool,config,ai_service,max_tokens}' = '8000';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN
    RAISE EXCEPTION '168: expected tool-improver.improve_tool at 8000, updated %', n;
  END IF;

  -- generate_tool_html: emits the whole component at birth, same exposure.
  UPDATE agent_definitions
  SET default_config = jsonb_set(default_config,
        '{workflow,steps,generate_tool_html,config,ai_service,max_tokens}', '32000'::jsonb)
  WHERE type='tool-generator' AND is_active
    AND default_config #>> '{workflow,steps,generate_tool_html,config,ai_service,max_tokens}' = '8000';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN
    RAISE EXCEPTION '168: expected tool-generator.generate_tool_html at 8000, updated %', n;
  END IF;
END $$;

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## Whole-component writers raised off the 8000-token ceiling
Observed: tool-improver.improve_tool returned output_tokens=8000 against max_tokens=8000 and its truncated output was saved over content_components.html_template, cutting tool-loot-table-balancer from 10,272 chars to a 1,253-char CSS fragment (no script/div/fieldset, ending mid-declaration). tool-generator.generate_tool_html carried the same 8000 ceiling; that tool''s birth had already used 6094/8000.
Root cause: agents that emit a WHOLE component were sized inconsistently — recreate_tool (same job) had 64000 while improve_tool/generate_tool_html had 8000.
Fix: both raised to 32000 (migration 168). Snapshots taken. NOT the real fix — the structural fix is a completeness guard that refuses to persist a component whose markup/script structure vanished (bugs_open/012).
Verified: component restored from component_versions (last complete 10,272-char version, matching the live page); truncated state kept in tmp_loot_truncated_20260718.
Categories: fix',
  '["fix"]'::jsonb,
  'migration', '168_component_writers_max_tokens'
);

COMMIT;
