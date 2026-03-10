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