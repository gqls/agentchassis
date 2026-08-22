-- 565_component_creator_prompt_names_the_source_vocabulary_ROLLBACK.sql
--
-- Undoes 565 by exact-replace removal of the inserted block, leaving the TIER C
-- anchor exactly as it was. Anchored on the same verbatim strings, so it cannot
-- silently remove a later, different edit.
--
-- Note the asymmetry with the forward migration and it is deliberate: this
-- refuses if the block is ABSENT (nothing to undo — do not pretend to succeed),
-- where the forward migration refuses if it is PRESENT.

SELECT snapshot_agent('component-creator',
  'migration 565 ROLLBACK: pre-removal (bugs_open/337 source vocabulary block)');

BEGIN;

DO $$
DECLARE
  tpl  text;
  hits int;
BEGIN
  SELECT default_config ->> 'prompt_template' INTO tpl
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF tpl IS NULL THEN
    RAISE EXCEPTION 'MIGRATION 565 ROLLBACK: no live component-creator prompt_template';
  END IF;

  hits := (length(tpl) - length(replace(tpl, '{{if .existing_component.aspect_paths}}', '')))
          / length('{{if .existing_component.aspect_paths}}');
  IF hits <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 565 ROLLBACK: guard opener present % times, expected exactly 1 — nothing to remove, or the prompt has drifted', hits;
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{prompt_template}',
      to_jsonb(replace(
        default_config ->> 'prompt_template',
        E'\n' ||
'{{if .existing_component.aspect_paths}}' || E'\n' ||
'     VALID site_specs SOURCES — the first path segment after "site_specs." MUST' || E'\n' ||
'     be one of these EXACTLY. Anything else is refused at store time and the' || E'\n' ||
'     whole component is thrown away, so do not invent one.' || E'\n' ||
'     "(N sites)" is how many sites carry that path. A shared component renders' || E'\n' ||
'     on all of them, so anything short of every site MUST be "required": false' || E'\n' ||
'     with "on_missing": "skip_field", and the template must gate its markup on' || E'\n' ||
'     the field.' || E'\n' ||
'{{.existing_component.aspect_paths}}' || E'\n' ||
'     Only the ASPECT (the first segment) is validated. A listed aspect with a' || E'\n' ||
'     key that is not listed above will pass validation and then resolve to' || E'\n' ||
'     nothing, rendering the section blank with no error — which is worse than' || E'\n' ||
'     being refused. So use a path exactly as listed, or do not use site_specs.' || E'\n' ||
'     IF THE VALUE YOU NEED IS NOT IN THE LIST, DO NOT INVENT A PATH. Use' || E'\n' ||
'     source "static" with a sensible fallback, or source "llm". Some values a' || E'\n' ||
'     section wants (a currency symbol, a CTA destination) exist nowhere in' || E'\n' ||
'     site_specs on any site, and static-with-a-fallback is the correct answer.' || E'\n' ||
'{{end}}',
        ''
      )),
      false
    ),
    version    = version + 1,
    updated_at = now()
WHERE type = 'component-creator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  tpl    text;
  anchor text := '   TIER C — SITE DATA (source: "site_specs.{path}" or "site_assets.{type}")';
BEGIN
  SELECT default_config ->> 'prompt_template' INTO tpl
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF position('{{if .existing_component.aspect_paths}}' in tpl) > 0 THEN
    RAISE EXCEPTION 'MIGRATION 565 ROLLBACK: the block is still present after removal';
  END IF;
  IF (length(tpl) - length(replace(tpl, anchor, ''))) / length(anchor) <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 565 ROLLBACK: the TIER C anchor did not survive the removal';
  END IF;
END $$;

COMMIT;
