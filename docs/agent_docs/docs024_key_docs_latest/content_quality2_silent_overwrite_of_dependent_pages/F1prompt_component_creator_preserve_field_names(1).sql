-- FILE: migrations/NNN_component_creator_preserve_field_names.sql
--
-- F1-prompt (Option A): make component-creator preserve a shared component's
-- field names on REGENERATION, so it stops silently stranding dependents'
-- content_data. This is the generation-time complement to the
-- StoreGeneratedComponentAction field-contract guard (the store-time backstop).
--
-- Three coupled edits to component-creator's default_config, applied atomically
-- under one snapshot:
--   1. insert a `load_existing_component` step (looks up the canonical existing
--      component by section_type and outputs `existing_component.field_names`);
--   2. run it before generate_template (rewire read_site_spec.next_step and the
--      two error_steps to it) and expose its output to the prompt
--      (append "existing_component" to generate_template.config.input_fields);
--   3. insert a dormant prompt rule that, WHEN existing field names are present,
--      instructs the LLM to reuse them verbatim (anti-drift anchor).
--
-- PREREQUISITE: deploy the `load_existing_component` Go action FIRST
-- (load_existing_component_action.go + registry.go). If this migration lands
-- before the action exists, the workflow references an unknown action and
-- generation fails. The action is advisory (never errors, always returns a
-- well-formed map), so once deployed a missing/!existing component just leaves
-- the prompt rule dormant.
--
-- PATH NOTE (072 nested-prompt trap): prompt_template is at the TOP LEVEL of
-- default_config (default_config->>'prompt_template'), NOT at
-- workflow.steps.generate_template.config.prompt_template.
--
-- Convention: snapshot-first, idempotent, drift-checked, live-row only.

BEGIN;

DO $mig$
DECLARE
    v_cfg  jsonb;
    p      text;
    v_if   jsonb;
    anchor text := '== COMPONENT CONTRACT';           -- unique, ASCII (avoids the em-dash)
    marker text := '== REGENERATION FIELD-NAME RULE ==';
    n      int;
    v_step jsonb := jsonb_build_object(
        'action',       'load_existing_component',
        'config',       jsonb_build_object('section_type', 'input_data.spec.section_type'),
        'next_step',    'generate_template',
        'description',  'Load existing component field names (regeneration field-name preservation)',
        'output_field', 'existing_component'
    );
    block  text := $blk$
{{if .existing_component.field_names}}
== REGENERATION FIELD-NAME RULE ==
This component ALREADY EXISTS and pages across the site have stored content keyed to its current field names. You are REGENERATING it, not creating a new one. In input_schema.fields and as {{placeholder "..."}} tokens you MUST reuse these exact field names, spelled identically (same case, same underscores):
{{.existing_component.field_names}}
You MAY add new fields, and you MAY restyle or restructure the template and CSS freely. You MUST NOT rename or drop any of the existing field names listed above. Renaming or removing one strands the stored content for every page that uses this shared component and renders those sections blank. If a field seems unneeded, keep its name present and reachable rather than removing it.
== END REGENERATION FIELD-NAME RULE ==
{{end}}

$blk$;
BEGIN
    SELECT default_config, default_config->>'prompt_template'
      INTO v_cfg, p
    FROM agent_definitions
    WHERE type = 'component-creator' AND deleted_at IS NULL AND is_active = true
      AND (is_snapshot IS NULL OR is_snapshot = false);

    IF v_cfg IS NULL THEN
        RAISE EXCEPTION 'component-creator live row not found';
    END IF;
    IF p IS NULL THEN
        RAISE EXCEPTION 'component-creator: top-level prompt_template not found (check the path / 072 trap)';
    END IF;

    -- idempotent: the step being present means this migration already ran
    IF v_cfg #> '{workflow,steps,load_existing_component}' IS NOT NULL
       AND position(marker IN p) > 0 THEN
        RAISE NOTICE 'component-creator: field-name preservation already applied — no change';
        RETURN;
    END IF;

    -- drift guard: the prompt anchor must appear exactly once
    n := (length(p) - length(replace(p, anchor, ''))) / length(anchor);
    IF n <> 1 THEN
        RAISE EXCEPTION 'component-creator: anchor "%" found % times (expected 1) — prompt drifted, aborting', anchor, n;
    END IF;

    -- append "existing_component" to generate_template input_fields (idempotent)
    v_if := v_cfg #> '{workflow,steps,generate_template,config,input_fields}';
    IF v_if IS NULL THEN
        RAISE EXCEPTION 'component-creator: generate_template.config.input_fields not found';
    END IF;
    IF NOT (v_if @> '["existing_component"]'::jsonb) THEN
        v_if := v_if || '["existing_component"]'::jsonb;
    END IF;

    PERFORM snapshot_agent('component-creator',
        'F1-prompt: regeneration field-name preservation (load step + prompt rule)');

    UPDATE agent_definitions
    SET default_config =
        jsonb_set(
          jsonb_set(
            jsonb_set(
              jsonb_set(
                jsonb_set(
                  jsonb_set(default_config,
                    '{workflow,steps,load_existing_component}', v_step, true),
                  '{workflow,steps,read_site_spec,next_step}', '"load_existing_component"'::jsonb, true),
                '{workflow,steps,read_site_spec,error_step}', '"load_existing_component"'::jsonb, true),
              '{workflow,steps,ensure_site_record,error_step}', '"load_existing_component"'::jsonb, true),
            '{workflow,steps,generate_template,config,input_fields}', v_if, true),
          '{prompt_template}', to_jsonb(replace(p, anchor, block || anchor)), true)
    WHERE type = 'component-creator' AND deleted_at IS NULL AND is_active = true
      AND (is_snapshot IS NULL OR is_snapshot = false);

    RAISE NOTICE 'component-creator: load_existing_component wired + regeneration field-name rule inserted';
END
$mig$;

-- verify
SELECT default_config #> '{workflow,steps,load_existing_component}' IS NOT NULL              AS step_present,
       default_config #> '{workflow,steps,read_site_spec,next_step}'                          AS read_site_spec_next,
       default_config #> '{workflow,steps,generate_template,config,input_fields}'             AS gen_input_fields,
       position('== REGENERATION FIELD-NAME RULE ==' IN (default_config->>'prompt_template')) > 0 AS rule_present,
       position('{{.existing_component.field_names}}' IN (default_config->>'prompt_template')) > 0 AS injects_names
FROM agent_definitions
WHERE type = 'component-creator' AND deleted_at IS NULL AND is_active = true;

COMMIT;
