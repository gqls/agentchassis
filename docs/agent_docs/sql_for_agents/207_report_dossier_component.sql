-- 207_report_dossier_component.sql
-- Gripper dossier pilot — the shared section component every report page mounts.
-- Workstream: docs/agent_docs/docs024_key_docs_latest/robot_hands_gripper_dossier/
-- Design of record: DESIGN_2026-07-24_gripper_dossier_pilot.md §3/§4.
--
-- PRE-IMAGE SAFE. This seed names no action and no agent; it inserts one
-- content_components row. It can be applied before or after the chassis image
-- roll. (Seeds 209/210, which name the new actions, are strictly POST-image.)
--
-- WHY THE TEMPLATE IS A ONE-LINE PASSTHROUGH.
-- The report's HTML is built in Go by create_report_page (renderReportSection)
-- and stored on page_components.rendered_html. rerender_single_page
-- CONCATENATES that stored HTML — it does not render templates — so this
-- html_template is a fallback shell, not the renderer. Precedent and shape
-- copied from the live 'ported-page' component, which is the same
-- body-passthrough case:
--     <section class="ported-page" data-component="ported-page">{{.body}}</section>
--
-- DO NOT grow this template into the real report markup. Two reasons, both
-- load bearing: (1) the markup would then exist twice and drift, and the Go
-- side is the one that actually renders; (2) Go's template engine renders a
-- missing field as empty with NO error (missingkey=zero, the platform's most
-- recent unpatched root cause), which on a page whose entire claim is "every
-- number traces or the run fails" is the worst available failure mode. The Go
-- renderer is deliberately template-free string building for exactly this
-- reason — see the header of create_report_page_action.go.
--
-- The report's CSS is inlined by the Go renderer, NOT here: rerender collects
-- no component stylesheets, and robot-hands.com's site stylesheet defines no
-- report-* class (checked live 2026-07-25), so a stylesheet left to the site
-- would ship the deliverable unstyled (bugs_open/027 class).
--
-- IDEMPOTENT: keyed on content_components_name_key (name is UNIQUE; function
-- is unique only for component_level='tool', which this is not).

BEGIN;

INSERT INTO content_components (
    name,
    display_name,
    description,
    function,
    component_level,
    render_mode,
    category,
    section_type,
    html_template,
    input_schema,
    semantic_tags,
    suitable_page_types,
    is_active,
    is_dark_section,
    created_from
) VALUES (
    'Gripper Selection & Integration Dossier',
    'Gripper Selection & Integration Dossier',
    'Per-request engineering dossier for robot-hands.com: a deterministic '
    || 'scored shortlist over the verified gripper index, with printed '
    || 'formulas, an SVG headroom chart, verified prose and a provenance '
    || 'footer. Mounted on /reports/<uuid>.html pages created by the '
    || 'create_report_page action; the section HTML is rendered in Go and '
    || 'stored on page_components.rendered_html, so this template is a '
    || 'passthrough shell (see seed 207 header before editing it).',
    'report-dossier',
    'section',
    'template',
    'report',
    'report-dossier',
    '<section class="section report-dossier" data-component="report-dossier">{{.body}}</section>',
    '{
       "type": "object",
       "properties": {
         "body": {
           "type": "string",
           "description": "Pre-rendered dossier HTML from create_report_page. Never authored by an LLM and never assembled from a template — see the seed header."
         }
       },
       "required": ["body"]
     }'::jsonb,
    '["report", "engineering", "gripper", "per-request", "deterministic"]'::jsonb,
    '["report"]'::jsonb,
    true,
    false,
    -- chk_created_from_valid allows only manual|generated|adopted|tool|forked
    -- (found by dry-running this seed, not by reading it off — the constraint
    -- is not in the column default and rejects anything else outright).
    'manual'
)
ON CONFLICT (name) DO UPDATE SET
    display_name        = EXCLUDED.display_name,
    description         = EXCLUDED.description,
    function            = EXCLUDED.function,
    component_level     = EXCLUDED.component_level,
    render_mode         = EXCLUDED.render_mode,
    category            = EXCLUDED.category,
    section_type        = EXCLUDED.section_type,
    html_template       = EXCLUDED.html_template,
    input_schema        = EXCLUDED.input_schema,
    semantic_tags       = EXCLUDED.semantic_tags,
    suitable_page_types = EXCLUDED.suitable_page_types,
    is_active           = true,
    updated_at          = NOW();

-- Assert the row create_report_page will actually resolve. Its lookup is
--   SELECT id FROM content_components
--   WHERE function = 'report-dossier' AND is_active = true
--   ORDER BY created_at DESC LIMIT 1
-- so a second active row under the same function would shadow this one.
DO $$
DECLARE
    n INTEGER;
BEGIN
    SELECT count(*) INTO n
    FROM content_components
    WHERE function = 'report-dossier' AND is_active = true;

    IF n <> 1 THEN
        RAISE EXCEPTION
            'expected exactly 1 active report-dossier component, found % — create_report_page takes the newest and would shadow the other', n;
    END IF;
END $$;

COMMIT;
