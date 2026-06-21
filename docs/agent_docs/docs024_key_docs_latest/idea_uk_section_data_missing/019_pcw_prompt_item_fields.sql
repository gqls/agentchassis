-- 019_pcw_prompt_item_fields.sql
--
-- Renumber to the next free migration index in your sequence if 019 is taken.
--
-- Purpose
--   Make the page-content-writer LLM prompt declare the per-element field names
--   for array-typed component fields, in both the "What To Write" instruction
--   list and the "Output Format" JSON skeleton. Previously the prompt listed an
--   array field (e.g. `features`) with its type but never its element shape, so
--   the model guessed the item keys (title/body) which render empty against a
--   template that reads name/description. This reads the ItemFields now carried
--   on each llm_field_spec by plan_sections_action.go.
--
-- Pairing
--   Requires the Go change in plan_sections_action.go (populates item_fields on
--   llm_field_specs) and the render-time reconciler in v3_site_actions.go. This
--   migration is safe to run before or after that deploy: until item_fields is
--   populated, {{if .item_fields}} is simply always false and the prompt renders
--   exactly as it did before. No broken intermediate state.
--
-- Idempotency
--   Re-running is a no-op: the patched prompt contains the sentinel
--   "{{if .item_fields}}", and the block below skips when that is present. This
--   matters because the new fragment contains the old fragment as a prefix, so a
--   blind replace() on already-patched text would expand it twice.

BEGIN;

DO $migrate$
DECLARE
    path text[] := ARRAY[
        'workflow','steps','process_sections_loop','config','sub_workflow',
        'steps','generate_content','config','prompt_template'
    ];
    p       text;
    newp    text;
    old_wtw text := $owtw$- `{{.name}}` ({{.type}}{{if .required}}, required{{end}}){{if .description}}: {{.description}}{{end}}$owtw$;
    new_wtw text := $nwtw$- `{{.name}}` ({{.type}}{{if .required}}, required{{end}}){{if .description}}: {{.description}}{{end}}{{if .item_fields}} — each item is an object with exactly these fields: {{range $i, $f := .item_fields}}{{if $i}}, {{end}}`{{$f}}`{{end}}{{end}}$nwtw$;
    old_out text := $oout$  "{{$f.name}}": "..."$oout$;
    new_out text := $nout$  "{{$f.name}}": {{if $f.item_fields}}[{ {{range $j, $k := $f.item_fields}}{{if $j}}, {{end}}"{{$k}}": "..."{{end}} }]{{else}}"..."{{end}}$nout$;
BEGIN
    SELECT default_config #>> path
      INTO p
      FROM agent_definitions
     WHERE type = 'page-content-writer';

    IF p IS NULL THEN
        RAISE EXCEPTION 'page-content-writer prompt_template not found at expected path — aborting';
    END IF;

    IF position('{{if .item_fields}}' IN p) > 0 THEN
        RAISE NOTICE 'item_fields already present in prompt — migration already applied, skipping';
        RETURN;
    END IF;

    IF position(old_wtw IN p) = 0 OR position(old_out IN p) = 0 THEN
        RAISE EXCEPTION
            'expected prompt fragments not found (wtw_pos=%, out_pos=%) — prompt structure has changed; aborting without modifying',
            position(old_wtw IN p), position(old_out IN p);
    END IF;

    newp := replace(replace(p, old_wtw, new_wtw), old_out, new_out);

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config, path, to_jsonb(newp), false),
           updated_at = now()
     WHERE type = 'page-content-writer';

    RAISE NOTICE 'page-content-writer prompt patched: item_fields rendering added to What To Write and Output Format';
END
$migrate$;

COMMIT;

-- Verification (run separately; expects both > 0 after apply):
-- SELECT
--   position('{{if .item_fields}}'   IN (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}')) AS wtw_marker,
--   position('{{if $f.item_fields}}' IN (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}')) AS out_marker
-- FROM agent_definitions WHERE type = 'page-content-writer';
