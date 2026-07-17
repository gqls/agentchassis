-- 160_compose_plan_interaction_shape.sql
--
-- The compose_plan prompt describes interaction checks in PROSE ("fill or
-- click steps plus an expect selector") but never shows the JSON shape. The
-- first Sonnet-5 birth (tool-loot-table-balancer, 2026-07-17, orchestration
-- f81bdcc9) followed the prose and improvised {"type":"click","selector":…,
-- "expect":"…"} — a shape the Tier-4 runner does not implement, so the check
-- is honestly SKIPPED ("click not implemented") and the tool's real behaviour
-- goes untested. Every earlier well-formed interaction check turns out to have
-- been hand-written in a migration (143/148), never composed — the composer
-- was never taught the shape.
--
-- Fix: spell out the exact interaction-check JSON in the prompt. Single-line
-- replace (prompt_template holds REAL newlines — 144/149's lesson: guards
-- must anchor on single-line substrings).
--
-- The mis-shaped check in the loot-table PLAN itself is superseded separately
-- once probed passing on the DEPLOYED page (the 148 rule: never author a
-- criterion you have not watched pass).
--
-- Applied via psql -f + ledger row (runner unblocked now, but this keeps the
-- one-file-at-a-time discipline used throughout this workstream).

BEGIN;

SELECT snapshot_agent('tool-generator', '160_compose_plan_interaction_shape: pre-update');

DO $$
DECLARE
  old_line text := 'Add ONE interaction check ONLY if you can copy real ids or classes from the HTML above (fill or click steps plus an expect selector). NEVER invent a selector — if unsure, omit the interaction check entirely.';
  new_line text := 'Add ONE interaction check ONLY if you can copy real ids or classes from the HTML above, using EXACTLY this shape: {"id":"<kebab-id>","type":"interaction","steps":[{"action":"click","selector":"#realId"}],"expect":{"selector":"#realResult"}} — allowed step actions are "fill" (with "value"), "click", and "select" (with "value"); "expect" is an OBJECT with "selector" (and optionally "text_matches"). No other check type exists for interactions — never emit "type":"click" or "type":"fill" as a check type. NEVER invent a selector — if unsure, omit the interaction check entirely.';
  cur text;
  n int;
BEGIN
  SELECT default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
    INTO cur FROM agent_definitions WHERE type='tool-generator' AND is_active;
  IF cur IS NULL THEN
    RAISE EXCEPTION '160: tool-generator compose_plan prompt_template not found';
  END IF;
  IF strpos(cur, old_line) = 0 THEN
    RAISE EXCEPTION '160: expected interaction sentence not found in compose_plan prompt — drifted, do not apply blind';
  END IF;

  UPDATE agent_definitions
  SET default_config = jsonb_set(default_config,
        '{workflow,steps,compose_plan,config,prompt_template}',
        to_jsonb(replace(cur, old_line, new_line)))
  WHERE type='tool-generator' AND is_active;

  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN
    RAISE EXCEPTION '160: expected 1 row updated, got %', n;
  END IF;
END $$;

-- Convention: workflow-altering migrations leave a pipeline note.
INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## compose_plan now specifies the interaction-check JSON shape
Observed: the first Sonnet-5 tool birth (tool-loot-table-balancer, 2026-07-17) emitted an interaction check as {"type":"click"} with a bare-string expect — not in the Tier-4 vocabulary, so the runner skips it and the tool''s behaviour goes untested. Selectors were real (the no-invention rule held); only the SHAPE was improvised, because the prompt described interactions in prose without an example.
Root cause: every previously well-formed interaction check was hand-written in a migration (143/148); compose_plan was never taught the shape.
Fix: prompt now contains the exact interaction JSON shape + allowed step actions, and forbids bare "click"/"fill" check types (migration 160). Snapshot taken pre-update.
Verified: post-apply prompt contains the shape example; next tool birth should emit type:"interaction".
Categories: fix',
  '["fix"]'::jsonb,
  'migration', '160_compose_plan_interaction_shape'
);

COMMIT;
