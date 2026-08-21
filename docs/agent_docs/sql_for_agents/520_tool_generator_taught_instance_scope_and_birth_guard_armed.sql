-- 520 — tool-generator learns the instance-scope rules; the birth guard is ARMED
--       (bugs_open/283 FLOW half; WRONG_CALLS 2026-08-21 producer-census entry;
--       owner ruling 2026-08-21: log it, teach the tool, turn the guard on,
--       prefer the Go guard auto-converting the LLM's ids).
--
-- TWO halves, deliberately in ONE file because neither harms the other in any
-- interleaving with the binary roll:
--   (1) prompt: two new rules (single IIFE + id discipline). Live immediately;
--       harmless against every binary (better-shaped generations, nothing reads
--       differently).
--   (2) save_tool step gains enforce_instance_scope=true. INERT until a binary
--       carrying ScopeToolBirthTemplate + the ConfigKeys declaration rolls
--       (commit of 2026-08-21). Pre-roll binaries neither read nor refuse the
--       key: CreateToolComponentInputSpec declared no config contract before
--       this change, so the strict validator (bugs_closed/336) does not fire.
--       NO _HOLD needed — verified against platform/validation/workflow.go
--       semantics: only a spec that declares its contract complete refuses
--       unknown keys.
--
-- The id-prefixing itself stays in Go (the proven deterministic transform),
-- NOT in the prompt: teaching the LLM to emit {{.InstanceID}} literally would
-- also have to survive the prompt_template's own Go-template rendering (an
-- escaping trap), and a prompt is a request where the guard is a control.
BEGIN;

-- Double-apply guard: the new rule text is the anchor.
DO $$
DECLARE p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}'
    INTO p FROM agent_definitions
   WHERE type='tool-generator' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF p IS NULL THEN
    RAISE EXCEPTION '520: no active tool-generator row with a generate_tool_html prompt';
  END IF;
  IF p LIKE '%exactly one IIFE%' THEN
    RAISE EXCEPTION '520: already applied (IIFE rule present) — nothing to do';
  END IF;
  -- Anchor check: the surgical replace below is verbatim-anchored on the end
  -- of rule 20 + the Structure heading; abort BEFORE the snapshot if the
  -- anchor has drifted (the 099/381 lesson: never hand-patch around a moved
  -- anchor, re-derive the edit).
  IF p NOT LIKE '%cannot tell that from a tool that is broken.' || E'\n\n' || '## Structure%' THEN
    RAISE EXCEPTION '520: rule-20/Structure anchor not found — the prompt has drifted; re-derive this edit against the live text';
  END IF;
END $$;

SELECT snapshot_agent('tool-generator', '520 pre-image: instance-scope rules + birth guard arming');

-- (1) The two new rules, spliced between rule 20 and ## Structure.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,generate_tool_html,config,prompt_template}',
      to_jsonb(replace(
        default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}',
        E'cannot tell that from a tool that is broken.\n\n## Structure',
        E'cannot tell that from a tool that is broken.\n'
        || E'21. Wrap everything in the <script> block below the tool-doc header in exactly one IIFE: (function () { ''use strict''; ... })();. Declare nothing at the top level and never assign window.onload; if you need setup when the page loads, register a DOMContentLoaded listener inside the IIFE.\n'
        || E'22. Give every element the script uses a unique, descriptive, kebab-case id, and look it up with document.getElementById using the same literal text the markup declares. If you must build an id at runtime, start it with a static prefix ending in a hyphen (for example ''row-'' + n). The platform namespaces every id per instance when the tool is saved; a script that hides element names from that pass (top-level declarations, dynamic ids with no static prefix) is refused and regenerated.\n'
        || E'\n## Structure')),
      false),
    updated_at = now()
WHERE type='tool-generator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- (2) Arm the birth guard on the save step.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,save_tool,config,enforce_instance_scope}',
      'true'::jsonb, true),
    updated_at = now()
WHERE type='tool-generator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config#>>'{workflow,steps,save_tool,action}' = 'create_tool_component';

-- Verify: both halves landed, exactly once, on the one live row.
DO $$
DECLARE p text; armed text; n int;
BEGIN
  SELECT default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}',
         default_config#>>'{workflow,steps,save_tool,config,enforce_instance_scope}'
    INTO p, armed FROM agent_definitions
   WHERE type='tool-generator' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  n := (length(p) - length(replace(p, 'exactly one IIFE', ''))) / length('exactly one IIFE');
  IF n <> 1 THEN
    RAISE EXCEPTION '520 verify: IIFE rule appears % times, expected exactly 1', n;
  END IF;
  IF p NOT LIKE '%22. Give every element%' THEN
    RAISE EXCEPTION '520 verify: id-discipline rule missing';
  END IF;
  IF armed IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '520 verify: enforce_instance_scope not armed on save_tool (got %)', COALESCE(armed,'NULL');
  END IF;
  IF p LIKE '%{{.InstanceID}}%' THEN
    RAISE EXCEPTION '520 verify: the prompt must NOT carry the literal placeholder — that is the Go guard''s job, and the prompt template would try to render it';
  END IF;
END $$;

COMMIT;
