-- 01_render_site_chrome_agent.sql — a minimal "render site chrome ONLY" agent.
--
-- WHY THIS EXISTS (bugs_open/054 verification gap): to behaviourally verify the
-- chrome dead-control drop/escalate (054) you must run render_site_components
-- against a site with an ungated chrome component and observe the stored
-- rendered_html + the chrome_dead_control work item. But NO existing agent renders
-- chrome without also deploying: rerender-site and nav-updater both chain
-- render_site_components -> render_js_snippets -> deploy_js_snippets (git_commit)
-- -> page (re)assembly/deploy (~26-37 commits to gqls/sites + B2 sync + CF purge
-- per run). That is far too heavy — and unsafe — for a scratch verification.
--
-- This agent is the missing primitive: ensure_site_record -> render_site_components
-- -> complete. It renders header/footer chrome and stores site_components.rendered_html,
-- and does NOTHING else — no JS render, no git commit, no page deploy. It is
-- MANUAL-DISPATCH ONLY (never scheduled), so it is inert unless explicitly invoked.
--
-- Config-only (agent_definitions): LIVE immediately, no image roll. Idempotent on
-- (type, version). Remove with:
--   UPDATE agent_definitions SET is_active=false, deleted_at=now() WHERE type='render-site-chrome';

INSERT INTO agent_definitions (type, display_name, description, category, version, is_active, default_config)
VALUES (
  'render-site-chrome',
  'Render Site Chrome (only)',
  'Renders site header/footer chrome and stores site_components.rendered_html. No JS render, no git commit, no page deploy. Manual-dispatch verify/debug primitive (bugs_open/054). Dispatch with input_data {site_id, domain}.',
  'orchestrator',
  1,
  true,
  '{
    "workflow": {
      "start_step": "ensure_site_record",
      "steps": {
        "ensure_site_record": {
          "action": "ensure_site_record",
          "config": {},
          "next_step": "render_chrome",
          "description": "Find the site record by domain; produces site_record.site_id"
        },
        "render_chrome": {
          "action": "render_site_components",
          "config": { "slots": ["header", "footer"], "force_rerender": true },
          "next_step": "complete",
          "description": "Render site chrome ONLY (bugs_open/054 verify) — no render_js_snippets, no deploy_js_snippets, no page assembly"
        },
        "complete": {
          "action": "complete_workflow",
          "config": { "output_fields": ["site_record", "site_components_rendered"] },
          "description": "Done"
        }
      }
    }
  }'::jsonb
)
ON CONFLICT (type, version) DO UPDATE
  SET default_config = EXCLUDED.default_config,
      description    = EXCLUDED.description,
      is_active      = true,
      deleted_at     = NULL,
      updated_at     = now();
