-- 019_pcw_prompt_item_fields_down.sql
--
-- Reverse of 019_pcw_prompt_item_fields.sql. Restores the page-content-writer
-- prompt to its pre-item_fields form by reversing the two replacements.
-- Idempotent: skips when the "{{if .item_fields}}" sentinel is absent (i.e. the
-- prompt is already at the old form). Run only if you need to roll the prompt
-- back; pair it with redeploying the previous chassis image if you also revert
-- the Go change.

BEGIN;

DO $down$
DECLARE
    path text[] := ARRAY[
        'workflow','steps','process_sections_loop','config','sub_workflow',
        'steps','generate_content','config','prompt_template'
    ];
    p    text;
    newp text;
    old_wtw text := $owtw$- `{{.name}}` ({{.type}}{{if .required}}, required{{end}}){{if .description}}: {{.description}}{{end}}$owtw$;
    new_wtw text := $nwtw$- `{{.name}}` ({{.type}}{{if .required}}, required{{end}}){{if .description}}: {{.description}}{{end}}{{if .item_fields}} — each item is an object with exactly these fields: {{range $i, $f := .item_fields}}{{if $i}}, {{end}}`{{$f}}`{{end}}{{end}}$nwtw$;
    old_out text := $oout$  "{{$f.name}}": "..."$oout$;
    new_out text := $nout$  "{{$f.name}}": {{if $f.item_fields}}[{ {{range $j, $k := $f.item_fields}}{{if $j}}, {{end}}"{{$k}}": "..."{{end}} }]{{else}}"..."{{end}}$nout$;
BEGIN
    SELECT default_config #>> path INTO p
      FROM agent_definitions WHERE type = 'page-content-writer';

    IF p IS NULL THEN
        RAISE EXCEPTION 'page-content-writer prompt_template not found at expected path — aborting';
    END IF;

    IF position('{{if .item_fields}}' IN p) = 0 THEN
        RAISE NOTICE 'item_fields sentinel absent — prompt already at old form, nothing to revert';
        RETURN;
    END IF;

    newp := replace(replace(p, new_out, old_out), new_wtw, old_wtw);

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config, path, to_jsonb(newp), false),
           updated_at = now()
     WHERE type = 'page-content-writer';

    RAISE NOTICE 'page-content-writer prompt reverted: item_fields rendering removed';
END
$down$;

COMMIT;
