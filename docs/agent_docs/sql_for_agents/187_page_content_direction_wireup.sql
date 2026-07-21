-- 187_page_content_direction_wireup.sql
--
-- bug 025: wire up pages.content_direction (per-page writer steering that was
-- documented but dead). Owner chose "wire it up as documented" 2026-07-21.
--
-- The Go half (load_page_record_action.go + get_pages_to_build_actions.go now
-- SELECT content_direction and put it on the page map -> current_page) ships in a
-- chassis image roll. This migration is the config half: it renders
-- .current_page.content_direction in the page-content-writer prompt, guarded so it
-- only appears when a page actually has a value.
--
-- Placement: immediately after the existing site-level "Content Direction (from
-- site spec)" block, giving the model a site -> page hierarchy of steering
-- (per-section steering via .current_section.component.content_brief already sits
-- lower in the same prompt).
--
-- Structure rendered matches the column's documented shape:
--   { "instruction": "...", "format": "...", "examples": [...], "avoid": [...] }
--
-- Idempotent: the WHERE guard skips the row if the block is already present, so a
-- re-run cannot double-insert.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(
            replace(
                default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template',
                E'{{.site_specs.specs.content_direction.formatted}}\n{{end}}{{end}}',
                E'{{.site_specs.specs.content_direction.formatted}}\n{{end}}{{end}}\n'
                || E'{{if .current_page.content_direction}}## Page-Specific Content Direction (for THIS page - follow closely)\n'
                || E'{{if .current_page.content_direction.instruction}}Instruction: {{.current_page.content_direction.instruction}}\n{{end}}'
                || E'{{if .current_page.content_direction.format}}Format: {{.current_page.content_direction.format}}\n{{end}}'
                || E'{{if .current_page.content_direction.examples}}Examples:\n{{range .current_page.content_direction.examples}}- {{.}}\n{{end}}{{end}}'
                || E'{{if .current_page.content_direction.avoid}}Avoid (do NOT do these):\n{{range .current_page.content_direction.avoid}}- {{.}}\n{{end}}{{end}}'
                || E'{{end}}'
            )
        )
    )
WHERE type = 'page-content-writer'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template'
      NOT LIKE '%current_page.content_direction%';

-- Correct the column comment: it now describes implemented behaviour (bug 025).
COMMENT ON COLUMN pages.content_direction IS
    'Optional per-page content direction for rebuilds. Wired to the page-content-writer prompt as .current_page.content_direction (bug 025, 2026-07-21). Structure: { "instruction": "...", "format": "...", "examples": [...], "avoid": [...] }. All keys optional. Requires chassis binary that SELECTs this column into current_page (load_page_record_action.go / get_pages_to_build_actions.go).';
