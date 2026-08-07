-- 326 — directory-build-handler: the builder bugs_open/206 found missing
--
-- `load_work_item_actions.go`'s own dispatch map named this agent type
-- (`directory-build-handler`) and item type (`needs_directory`) years before
-- anyone built it — commented out, listed under "known page types whose
-- builders don't exist yet". This is that agent.
--
-- WHAT IT DOES. Two steps, no content-writing logic of its own:
--   1. ensure_layout   — calls the new `ensure_page_section_layout` Go action
--      (this migration's sibling Go change). Fills the target page's plan
--      with a default layout ONLY if it currently has none from any source
--      (refuses otherwise — see the action's own header for the guard that
--      makes this safe under "never re-plan this site", bugs_closed/001).
--   2. call_page_build_handler — delegates the actual build/deploy to the
--      EXISTING generic `page-build-handler`, passing through the same
--      site_id/domain/page_name. No new build logic; this agent's whole job
--      is making sure page-build-handler has something to build.
--
-- image_tag: set to the build this migration ships alongside — v1.0.1260 is
-- the FIRST tag containing ensure_page_section_layout_action.go and the
-- query.business_directory resolver. If that tag is not yet live when this
-- runs, this row is inert config (harmless) until it is — but do NOT
-- dispatch this agent type before confirming the image is live: pod-grep
-- BOTH replicas for `ensure_page_section_layout` per this repo's own
-- standing practice, never trust the tag alone.
--
-- ORDER: DB config, live the moment it commits. This row alone does
-- nothing — nothing targets `directory-build-handler` until
-- `load_work_item_actions.go`'s map entry (already in this same change set)
-- ships in that image, AND a work item's `handler_agent` is actually set to
-- this type (done separately, as a live re-triage of the two named
-- bugs_open/206 rows, once the image is verified live — not bundled into
-- this migration, since it is a one-off operational action on two specific
-- rows, not a repeatable schema/config change).
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 326_….sql

BEGIN;

INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, is_active, status,
    default_config, topics, env_vars, idle_timeout_seconds
) VALUES (
    'directory-build-handler',
    'Directory Page Build Handler',
    'Ensures an entity-directory page''s plan carries a real, data-backed section layout (directory-listing, sourced from the site''s own business_intel directory), then delegates the actual build/deploy to page-build-handler. bugs_open/206.',
    'site',
    'docker.io/aqls/agent-chassis', 'v1.0.1260',
    true, 'active',
    '{
        "workflow": {
            "start_step": "ensure_layout",
            "steps": {
                "ensure_layout": {
                    "action": "ensure_page_section_layout",
                    "config": {
                        "site_id": "input_data.site_id",
                        "page_name": "input_data.page_name"
                    },
                    "next_step": "call_page_build_handler",
                    "output_field": "layout_result",
                    "description": "Fill the page''s plan with a default layout if it has none"
                },
                "call_page_build_handler": {
                    "action": "call_agent",
                    "config": {
                        "agent_type": "page-build-handler",
                        "target_role": "page_builder",
                        "input_mapping": {
                            "input_data.site_id": "input_data.site_id",
                            "input_data.domain": "input_data.domain",
                            "input_data.page_name": "input_data.page_name"
                        },
                        "timeout_seconds": 900
                    },
                    "next_step": "complete",
                    "output_field": "build_result",
                    "description": "Delegate the actual build/deploy to the existing generic builder"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": { "output_fields": ["layout_result", "build_result"] },
                    "description": "Directory page build complete"
                }
            }
        },
        "processing_mode": "task"
    }'::jsonb,
    '["system.agent.generic.requests", "system.agent.generic.responses"]'::jsonb,
    '[]'::jsonb, 0
) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config INTO cfg FROM agent_definitions
    WHERE type = 'directory-build-handler' AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '326: directory-build-handler not found after insert';
    END IF;
    IF cfg #>> '{workflow,steps,ensure_layout,action}' != 'ensure_page_section_layout' THEN
        RAISE EXCEPTION '326: ensure_layout step does not call ensure_page_section_layout';
    END IF;
    IF cfg #>> '{workflow,steps,call_page_build_handler,config,agent_type}' != 'page-build-handler' THEN
        RAISE EXCEPTION '326: call_page_build_handler does not delegate to page-build-handler';
    END IF;

    RAISE NOTICE '326 OK: directory-build-handler seeded, ensure_layout -> page-build-handler';
END $$;

COMMIT;
