-- 612_component_creator_store_honours_advised_identity.sql
--
-- bugs_open/388 — the CONFIG half. Wires the advised storage identity into
-- store_generated_component, and re-gates the prompt's function pin so it
-- survives a component with an empty schema.
--
-- WHAT IS BROKEN. Which content_components row a regeneration overwrites is
-- decided by the function name the LLM writes into its own output
-- (store_generated_component_action.go parseGeneratedTemplate, falling back to
-- NormaliseToKebab(section_type)). The field contract the writer is TOLD to
-- preserve comes from a different resolver in load_existing_component_action.go,
-- keyed on section_type. Since e1951c24b (2026-08-22) the two have been bridged
-- by a SENTENCE IN THIS PROMPT asking the model to echo the advised name back —
-- which nothing validates and nothing records.
--
-- TWO EDITS, AND THEY ARE INDEPENDENT.
--
-- (a) THE WIRE. `advised_identity?` on the store_component step points at the
--     `existing_component` object the advisory already returns and this prompt
--     already consumes. The Go half (commit 30d223291) reads `component_id`
--     from it and resolves the write target BY PRIMARY KEY.
--
--     THE `?` IS A SUFFIX AND IT IS DELIBERATE. datahelpers.MarkedConfigKey is
--     strings.HasSuffix(key, "?"), and the live config surface agrees
--     (related_pages?, component_id?, page_type?, replace_existing?). An
--     OPTIONAL-EXPLICIT wire resolves its declared path or LEAVES THE FIELD
--     ABSENT — it never falls through to the aggressive whole-tree search. That
--     matters more here than for most keys: a field that decides WHICH ROW IS
--     OVERWRITTEN must resolve from its declared path or not at all. The ack is
--     recorded in architecture_review/optional_explicit_wire_acks.json.
--
-- (b) THE RE-GATE. The pin currently renders only inside
--     {{if .existing_component.field_names}}. So a row that RESOLVED but whose
--     input_schema carries no fields gets an identity and no pin — a guard
--     conditional on the very thing it protects. [MEASURED 2026-08-25] 5 of 154
--     active non-forked section rows are schema-less; 4 of those 5 happen to
--     carry function == section_type, so it is benign by luck, not design. The
--     two pin sentences move into their own {{if .existing_component.function}}
--     block, which is non-empty on every path where an identity resolved.
--
-- WHY THE PIN STAYS AT ALL, now that the Go decides the row. It is still the
-- only defence on the UN-PINNED paths (a genuine creation, an unknown requester
-- with no work item, an error_step that routed around the advisory, either
-- fail-open), and it keeps the model's own output internally coherent — the
-- template's data-component attribute and its CSS scope derive from the same
-- name. Removing it would trade a cheap redundancy for a silent divergence.
--
-- NO ORDERING CONSTRAINT IS CLAIMED, BECAUSE NONE EXISTS (owner ruling
-- 2026-07-29 §2). Applied before the roll: the old binary's InputSpec does not
-- declare `advised_identity`, so the key is inert, and the re-gated pin renders
-- from `function`, which the old advisory ALREADY returns on every found-a-row
-- path — so half (b) is an improvement immediately. Rolled before applied: the
-- new binary simply finds no pin and runs the legacy resolution verbatim.
-- Either order is safe and neither is preferred.
--
-- Rollback: 612_component_creator_store_honours_advised_identity_ROLLBACK.sql

SELECT snapshot_agent('component-creator',
  'migration 612: pre-update (bugs_open/388 — advised storage identity wire + function-pin re-gate)');

BEGIN;

-- ── Pre-state gates ─────────────────────────────────────────────────────────
-- DO/RAISE, not a verify block of SELECTs: ON_ERROR_STOP ignores a non-empty
-- result set, so only an exception can stop the COMMIT.
DO $$
DECLARE
  live_rows int;
  tpl       text;
  pin       text := 'Also set the top-level "function" in your output JSON to exactly: {{.existing_component.function}}';
  fields_if text := '{{if .existing_component.field_names}}';
  hits      int;
  step_cfg  jsonb;
BEGIN
  SELECT count(*) INTO live_rows
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF live_rows <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 612: expected exactly 1 live component-creator row, found %', live_rows;
  END IF;

  SELECT default_config ->> 'prompt_template',
         default_config -> 'workflow' -> 'steps' -> 'store_component' -> 'config'
    INTO tpl, step_cfg
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  -- The prompt lives at the TOP level of default_config, not under the step's
  -- config (generate_template.config carries only ai_service and input_fields).
  -- Ten minutes were lost to that on 2026-08-25; the gate is here so nobody
  -- repeats it silently.
  IF tpl IS NULL OR length(tpl) = 0 THEN
    RAISE EXCEPTION 'MIGRATION 612: default_config->>''prompt_template'' is absent or empty';
  END IF;

  IF step_cfg IS NULL THEN
    RAISE EXCEPTION 'MIGRATION 612: workflow.steps.store_component.config is absent — the workflow shape has moved';
  END IF;

  -- Anchors read from the LIVE row, never from an on-disk seed.
  hits := (length(tpl) - length(replace(tpl, pin, ''))) / length(pin);
  IF hits <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 612: the function-pin sentence was found % times, expected exactly 1 — the prompt has drifted; re-read it before applying', hits;
  END IF;

  hits := (length(tpl) - length(replace(tpl, fields_if, ''))) / length(fields_if);
  IF hits <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 612: the field_names guard opener was found % times, expected exactly 1', hits;
  END IF;

  -- Double-apply refusal, on each half independently.
  IF position('{{if .existing_component.function}}' in tpl) > 0 THEN
    RAISE EXCEPTION 'MIGRATION 612: the function-pin block is already present — refusing to insert a second copy';
  END IF;
  IF step_cfg ? 'advised_identity?' THEN
    RAISE EXCEPTION 'MIGRATION 612: store_component.config already carries advised_identity? — refusing to re-wire';
  END IF;
END $$;

-- ── (a) The wire ────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,store_component,config,advised_identity?}',
      '"existing_component"'::jsonb,
      true
    ),
    version    = version + 1,
    updated_at = now()
WHERE type = 'component-creator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── (b) The re-gate: lift the pin out of the field_names block ──────────────
-- Two replaces, in this order. The first REMOVES the two pin sentences from the
-- field-name block (they are consecutive lines, so the trailing newline of the
-- second goes with them). The second INSERTS the standalone block immediately
-- above the field_names guard, preserving that guard.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{prompt_template}',
      to_jsonb(
        replace(
          replace(
            default_config ->> 'prompt_template',
            'Also set the top-level "function" in your output JSON to exactly: {{.existing_component.function}}' || E'\n' ||
            'Do NOT choose a different function name — the component library matches regenerations by function, and a different name silently creates a parallel duplicate component instead of regenerating this one.' || E'\n',
            ''
          ),
          '{{if .existing_component.field_names}}',
'{{if .existing_component.function}}' || E'\n' ||
'== REGENERATION IDENTITY ==' || E'\n' ||
'This component ALREADY EXISTS in the library. Set the top-level "function" in your output JSON to exactly:' || E'\n' ||
'{{.existing_component.function}}' || E'\n' ||
'Do NOT choose a different function name. Use it for the data-component attribute and the CSS scope class too, so the template, the schema and the stored row all agree.' || E'\n' ||
'{{end}}' || E'\n' ||
'{{if .existing_component.field_names}}'
        )
      ),
      false
    ),
    version    = version + 1,
    updated_at = now()
WHERE type = 'component-creator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── Post-state verification ─────────────────────────────────────────────────
DO $$
DECLARE
  tpl       text;
  step_cfg  jsonb;
  fn_if     text := '{{if .existing_component.function}}';
  fields_if text := '{{if .existing_component.field_names}}';
  hits      int;
BEGIN
  SELECT default_config ->> 'prompt_template',
         default_config -> 'workflow' -> 'steps' -> 'store_component' -> 'config'
    INTO tpl, step_cfg
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF (step_cfg ->> 'advised_identity?') IS DISTINCT FROM 'existing_component' THEN
    RAISE EXCEPTION 'MIGRATION 612: advised_identity? did not land as "existing_component" (got %)', step_cfg ->> 'advised_identity?';
  END IF;

  hits := (length(tpl) - length(replace(tpl, fn_if, ''))) / length(fn_if);
  IF hits <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 612: the new function guard is present % times, expected 1', hits;
  END IF;

  -- The field_names block must SURVIVE: the insert goes above it, never over it.
  hits := (length(tpl) - length(replace(tpl, fields_if, ''))) / length(fields_if);
  IF hits <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 612: the field_names guard is present % times, expected exactly 1 — the insert overwrote or duplicated it', hits;
  END IF;

  -- The OLD pin sentence must be GONE from the field-name block. If both
  -- copies survived, the writer is told twice and the re-gate achieved nothing.
  IF position('Also set the top-level "function" in your output JSON to exactly:' in tpl) > 0 THEN
    RAISE EXCEPTION 'MIGRATION 612: the original pin sentence is still present — the removal replace did not match';
  END IF;

  -- The new block must close, or every line after it is swallowed.
  IF position('{{end}}' in substring(tpl from position(fn_if in tpl))) = 0 THEN
    RAISE EXCEPTION 'MIGRATION 612: the inserted {{if}} block has no {{end}}';
  END IF;

  -- The placeholder must still be reachable exactly once.
  hits := (length(tpl) - length(replace(tpl, '{{.existing_component.function}}', ''))) / length('{{.existing_component.function}}');
  IF hits <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 612: the function placeholder is present % times, expected 1', hits;
  END IF;
END $$;

COMMIT;
