-- ============================================================================
-- Update completeness-discovery-agent with new integrity checks
--
-- Adds four new checks to the existing empty_sections check:
--   - cross_site_contamination: other site's company name in rendered HTML
--   - unrendered_templates: raw {{.field}} syntax in stored HTML
--   - missing_style_collection: site has no style_collection_id
--   - deactivated_site_components: site_components → inactive content_components
--
-- All are algorithmic — no LLM budget.
--
-- Run AFTER deploying integrity_checks.go (registers the checks via init()).
-- ============================================================================

UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "ensure_site_record",
        "processing_mode": "task",
        "timeout_seconds": 120,
        "steps": {
            "ensure_site_record": {
                "action": "ensure_site_record",
                "config": { "input_fields": ["site_id", "domain"] },
                "next_step": "run_structural_checks",
                "description": "Load site record from domain or site_id",
                "output_field": "site_record"
            },
            "run_structural_checks": {
                "action": "run_discovery_checks",
                "config": {
                    "site_id": "site_record.site_id",
                    "check_domain": "structural",
                    "checks": [
                        "missing_style_collection",
                        "deactivated_site_components",
                        "unrendered_templates",
                        "cross_site_contamination"
                    ]
                },
                "next_step": "run_content_checks",
                "description": "Run structural integrity checks — style collection, component linkage, template rendering, contamination",
                "output_field": "structural_result"
            },
            "run_content_checks": {
                "action": "run_discovery_checks",
                "config": {
                    "site_id": "site_record.site_id",
                    "check_domain": "content",
                    "checks": [
                        "empty_sections"
                    ]
                },
                "next_step": "complete",
                "description": "Run content completeness checks — empty deployed sections",
                "output_field": "content_result"
            },
            "complete": {
                "action": "complete_workflow",
                "config": { "output_fields": ["structural_result", "content_result"] },
                "description": "Completeness discovery complete"
            }
        }
    }
}'::jsonb,
description = 'Scans sites for structural integrity and content completeness issues: cross-site contamination, unrendered templates, missing style collections, deactivated components, empty sections. All checks algorithmic — no LLM budget needed.',
updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- Verify
SELECT type, description,
       default_config->'workflow'->'steps'->'run_structural_checks'->'config'->'checks' as structural_checks,
       default_config->'workflow'->'steps'->'run_content_checks'->'config'->'checks' as content_checks
FROM agent_definitions
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- several updates

-- ============================================================================
-- Wire integrity checks into the correct discovery agents
--
-- Per 009b improvement loop flow:
--   Step 3: design-discovery-agent  → structural linkage checks
--   Step 4: completeness-discovery-agent → content integrity checks
--
-- The improvement loop runs these in order:
--   quality-discovery (step 2) → design-discovery (step 3) → completeness-discovery (step 4)
-- so structural fixes (missing style, stale components) are found before
-- content checks (contamination, unrendered templates) run.
--
-- All checks produce status: 'detected'. The improvement loop's triage step (7)
-- promotes them to 'triaged' for dispatch.
--
-- Run AFTER deploying integrity_checks.go (registers checks via init()).
-- ============================================================================

-- ── 1. Update design-discovery-agent ──
-- Add: missing_style_collection, deactivated_site_components,
--       stale_site_components, shared_css_theme

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["undeployed_assets", "missing_css", "duplicate_palette", "hardcoded_section_colors", "forced_text_colors", "missing_tools", "missing_style_collection", "deactivated_site_components", "stale_site_components", "shared_css_theme"]'::jsonb
                     ),
    description = 'Scans sites for design-domain issues: undeployed assets, missing CSS, colour problems, missing style collections, deactivated components, stale header/footer renders, shared style collections. All algorithmic — no LLM budget.',
    updated_at = NOW()
WHERE type = 'design-discovery-agent' AND deleted_at IS NULL;

-- ── 2. Update completeness-discovery-agent ──
-- Add: cross_site_contamination, unrendered_templates
-- Keep: empty_sections (existing)

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["empty_sections", "cross_site_contamination", "unrendered_templates"]'::jsonb
                     ),
    description = 'Scans sites for content completeness and integrity: empty sections, cross-site company name contamination, unrendered Go template syntax. All algorithmic — no LLM budget.',
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- ── 3. Verify both agents ──

SELECT type,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as checks,
       substring(description, 1, 100) as desc
FROM agent_definitions
WHERE type IN ('design-discovery-agent', 'completeness-discovery-agent')
  AND deleted_at IS NULL
ORDER BY type;

-- Expected:
-- completeness-discovery-agent | ["empty_sections", "cross_site_contamination", "unrendered_templates"]
-- design-discovery-agent       | ["undeployed_assets", "missing_css", ..., "shared_css_theme"]



-- ============================================================================
-- Add missing_structure check to completeness-discovery-agent
--
-- This check runs alongside empty_sections. It flags deployed pages
-- where rendered_header or rendered_footer is NULL.
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["empty_sections", "missing_structure"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- Verify
SELECT type,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as checks
FROM agent_definitions
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

