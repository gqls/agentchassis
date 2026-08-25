-- 612_component_creator_store_honours_advised_identity_ROLLBACK.sql
--
-- The exact inverse of 612. Removes the `advised_identity?` wire and puts the
-- two function-pin sentences back inside the field_names block.
--
-- WHAT ROLLING BACK ACTUALLY BUYS YOU, so the decision is made with open eyes:
--
--   * Removing the WIRE returns store_generated_component to resolving the row
--     it overwrites from the function name the LLM wrote into its own output.
--     The Go half is unharmed — an absent pin is a legal, expected state and
--     the legacy path runs verbatim — so this is a clean disarm, not a break.
--     It also re-arms bugs_open/388: on the 27 divergent section_types the
--     advisory and the store can name different rows again.
--
--   * Removing the RE-GATE puts the pin back behind
--     {{if .existing_component.field_names}}, so a resolved row with an empty
--     input_schema loses its pin again. That half is an improvement even on the
--     OLD binary, so consider rolling back only half (a) — the two UPDATEs
--     below are independent and either may be run alone.
--
-- ⚠ THIS IS NOT A ROLLBACK OF THE GO. Commit 30d223291 stays live; it simply
-- finds no pin. There is nothing to undo on that side and nothing to rebuild.
--
-- ⚠ IF YOU ARE ROLLING BACK BECAUSE OF A SUSPECTED MISWRITE, PREFER THE
-- SNAPSHOT. 612 takes one via snapshot_agent before it touches anything, and
-- restoring that is exact where these replaces are merely inverse.

BEGIN;

-- ── Pre-state gates ─────────────────────────────────────────────────────────
DO $$
DECLARE
  live_rows int;
  tpl       text;
  step_cfg  jsonb;
BEGIN
  SELECT count(*) INTO live_rows
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF live_rows <> 1 THEN
    RAISE EXCEPTION 'ROLLBACK 612: expected exactly 1 live component-creator row, found %', live_rows;
  END IF;

  SELECT default_config ->> 'prompt_template',
         default_config -> 'workflow' -> 'steps' -> 'store_component' -> 'config'
    INTO tpl, step_cfg
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF NOT (step_cfg ? 'advised_identity?')
     AND position('{{if .existing_component.function}}' in tpl) = 0 THEN
    RAISE EXCEPTION 'ROLLBACK 612: neither half of 612 is present — nothing to roll back';
  END IF;
END $$;

-- ── Inverse of (a): remove the wire ─────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,store_component,config}',
      (default_config -> 'workflow' -> 'steps' -> 'store_component' -> 'config') - 'advised_identity?',
      false
    ),
    version    = version + 1,
    updated_at = now()
WHERE type = 'component-creator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND (default_config -> 'workflow' -> 'steps' -> 'store_component' -> 'config') ? 'advised_identity?';

-- ── Inverse of (b): put the pin back where it was ───────────────────────────
-- Removes the standalone block, then restores the two sentences inside the
-- field-name block, immediately after the field_names placeholder — which is
-- exactly where 612 took them from.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{prompt_template}',
      to_jsonb(
        replace(
          replace(
            default_config ->> 'prompt_template',
'{{if .existing_component.function}}' || E'\n' ||
'== REGENERATION IDENTITY ==' || E'\n' ||
'This component ALREADY EXISTS in the library. Set the top-level "function" in your output JSON to exactly:' || E'\n' ||
'{{.existing_component.function}}' || E'\n' ||
'Do NOT choose a different function name. Use it for the data-component attribute and the CSS scope class too, so the template, the schema and the stored row all agree.' || E'\n' ||
'{{end}}' || E'\n',
            ''
          ),
          '{{.existing_component.field_names}}' || E'\n',
          '{{.existing_component.field_names}}' || E'\n' ||
          'Also set the top-level "function" in your output JSON to exactly: {{.existing_component.function}}' || E'\n' ||
          'Do NOT choose a different function name — the component library matches regenerations by function, and a different name silently creates a parallel duplicate component instead of regenerating this one.' || E'\n'
        )
      ),
      false
    ),
    version    = version + 1,
    updated_at = now()
WHERE type = 'component-creator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND position('{{if .existing_component.function}}' in (default_config ->> 'prompt_template')) > 0;

-- ── Post-state verification ─────────────────────────────────────────────────
DO $$
DECLARE
  tpl      text;
  step_cfg jsonb;
  hits     int;
  pin      text := 'Also set the top-level "function" in your output JSON to exactly: {{.existing_component.function}}';
BEGIN
  SELECT default_config ->> 'prompt_template',
         default_config -> 'workflow' -> 'steps' -> 'store_component' -> 'config'
    INTO tpl, step_cfg
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF step_cfg ? 'advised_identity?' THEN
    RAISE EXCEPTION 'ROLLBACK 612: advised_identity? is still wired';
  END IF;

  IF position('{{if .existing_component.function}}' in tpl) > 0 THEN
    RAISE EXCEPTION 'ROLLBACK 612: the standalone function block is still present';
  END IF;

  -- The pin must be back, exactly once — not zero (lost) and not twice
  -- (restored on top of a copy the removal missed).
  hits := (length(tpl) - length(replace(tpl, pin, ''))) / length(pin);
  IF hits <> 1 THEN
    RAISE EXCEPTION 'ROLLBACK 612: the pin sentence is present % times, expected exactly 1', hits;
  END IF;

  hits := (length(tpl) - length(replace(tpl, '{{if .existing_component.field_names}}', '')))
          / length('{{if .existing_component.field_names}}');
  IF hits <> 1 THEN
    RAISE EXCEPTION 'ROLLBACK 612: the field_names guard is present % times, expected 1', hits;
  END IF;
END $$;

COMMIT;
