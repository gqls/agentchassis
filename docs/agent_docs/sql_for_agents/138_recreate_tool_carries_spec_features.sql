-- 0NN_recreate_tool_carries_spec_features.sql — recreate_tool must SEE the
-- spec's behaviour requirements. DRAFT 2026-07-09. DB-only; effective
-- immediately. Renumber 0NN.
--
-- FINDING (run e1018366, 2026-07-09): a recreation was triggered with two
-- explicit fixes in spec.interactive_features (map the Player-Influx slider
-- index to a real rate array [0,1,5,15,40,100]; give the Players chart series
-- its own axis). analyze_tool's prompt DOES render spec.interactive_features,
-- and its functional spec carried both fixes verbatim (INFLUX_MAP in
-- data_model.internal_state; "raw value must be mapped to rate array before
-- use" in the input constraints). recreate_tool's prompt, however, includes
-- input_data in input_fields but NEVER RENDERS input_data.spec — so the fixes
-- reached Opus only as details buried in the analysis JSON, while the original
-- HTML shell it is told to study says "New players joining per tick" beside
-- the slider. Opus followed the visible HTML semantics and faithfully
-- recreated both bugs (deployed 19:36:03Z; parseInt(slPop.value) used as the
-- rate; Players bound to the gold axis y1).
--
-- FIX: insert a "Mandatory Behaviour Requirements" section into
-- recreate_tool.prompt_template, rendered from input_data.spec
-- .interactive_features, placed BETWEEN the functional spec and the design
-- context, and marked as overriding the original source. Anchored on the
-- unique "## Design Context" heading via replace().
--
-- Standing rule: snapshot_agent opens the transaction.

BEGIN;

SELECT snapshot_agent('tool-recreation-handler', '0NN_recreate_tool_carries_spec_features.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,recreate_tool,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}',
        '## Design Context',
        '## Mandatory Behaviour Requirements (from the recreation spec)' || chr(10) ||
        'These are explicit requirements for THIS recreation. They OVERRIDE anything implied by the original source code or the functional specification. Implement every one of them exactly as described.' || chr(10) ||
        '{{if .input_data.spec.interactive_features}}{{range .input_data.spec.interactive_features}}- {{.name}} ({{.type}}): {{.description}}' || chr(10) ||
        '{{end}}{{else}}None supplied - recreate faithfully from the specification above.{{end}}' || chr(10) || chr(10) ||
        '## Design Context'))
    )
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;

DO $$
DECLARE n int; tmpl text;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;
    IF n <> 1 THEN RAISE EXCEPTION 'expected exactly one live row, found %', n; END IF;

    SELECT default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}' INTO tmpl
    FROM agent_definitions
    WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;

    -- new section present exactly once, before Design Context, and renders the spec features
    IF (length(tmpl) - length(replace(tmpl, 'Mandatory Behaviour Requirements', ''))) / length('Mandatory Behaviour Requirements') <> 1 THEN
        RAISE EXCEPTION 'Mandatory Behaviour Requirements heading not present exactly once';
    END IF;
    IF strpos(tmpl, '{{range .input_data.spec.interactive_features}}') = 0 THEN
        RAISE EXCEPTION 'interactive_features range block missing';
    END IF;
    IF strpos(tmpl, 'Mandatory Behaviour Requirements') > strpos(tmpl, '## Design Context') THEN
        RAISE EXCEPTION 'section landed after Design Context, expected before';
    END IF;
    -- untouched invariants
    IF strpos(tmpl, '{{.tool_analysis.result | toJSON}}') = 0 THEN
        RAISE EXCEPTION 'functional-spec block disturbed';
    END IF;
END $$;

COMMIT;

-- Verify after apply:
--   SELECT strpos(default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}',
--                 'Mandatory Behaviour Requirements') > 0 AS has_requirements_section
--   FROM agent_definitions WHERE type='tool-recreation-handler' AND deleted_at IS NULL;
--
-- Rollback: restore from the snapshot taken at the top, or replace() the
-- inserted section (from 'Mandatory Behaviour Requirements' through the blank
-- line before '## Design Context') back out of the template.
