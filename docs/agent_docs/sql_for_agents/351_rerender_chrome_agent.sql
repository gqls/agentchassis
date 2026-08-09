-- 351_rerender_chrome_agent.sql
--
-- bugs_open/226 follow-through (owner instruction 2026-08-09: "please dispatch
-- the rebuilds"): fingerprint the fleet's chrome WITHOUT rebuilding pages.
--
-- Why a new agent type instead of reusing one of the six that already carry
-- `render_site_components` (measured 2026-08-09, live agent_definitions):
--
--   rerender-pages           forces chrome, then files one page_rerender work
--                            item PER PAGE (542 deployed/active pages fleet-wide
--                            at measurement) into a queue whose drain half is
--                            bugs_open/083 — two orders of magnitude more churn
--                            than the goal;
--   rerender-site            same page fan-out via spawned rerenderers/deployers;
--   nav-link-fixer           smallest (6 steps) but MUTATES nav templates
--                            (fix_nav_link_templates find/replace) before it
--                            renders, and only renders header+footer — the head
--                            slot would stay unstamped;
--   nav-updater              rewrites nav content_data from the page graph;
--   pageflow-builder /       34/38-step build orchestrators; chrome is
--   site-work-orchestrator   incidental to what they do.
--
-- None is "render the three chrome slots, stamp, stop". This agent is exactly
-- that: TWO steps, both existing registered actions, zero Go changes, zero LLM
-- calls (render_site_components is deterministic template fill). Each render
-- lands in the store UPDATE that stamps `rendered_html_digest = md5(html)`
-- (v1.0.1270+, live fleet v1.0.1274 binary-verified 2026-08-09), and the
-- mig-344 trigger archives any outgoing bytes that differ, so a latent
-- hand-patch on an unstamped slot is archived (divergence='unstamped'), never
-- silently destroyed. Human-locked slots (lock gate, bugs_open/069) are
-- refused as on any forced render — the two authored-chrome sites keep their
-- locks and simply stay unstamped until an owner lifts them.
--
-- Register entry STY-055 (styling-render-pipeline.md) ships in the same
-- commit, per the platform-seam registration rule (CLAUDE.md 2026-07-28/29).
--
-- Dispatch envelope (same route the 226 e2e drill proved):
--   topic system.agent.generic.requests, action=orchestrate,
--   config={"agent_type":"rerender-chrome"},
--   input_data={"site_id":"<uuid>","domain":"<domain>"}
--
-- ROLLBACK RECIPE (also in 351_rerender_chrome_agent_ROLLBACK.sql):
--   UPDATE agent_definitions SET is_active=false, deleted_at=now()
--    WHERE type='rerender-chrome' AND deleted_at IS NULL;
-- (soft-delete; the row is config only and holds no data worth keeping live).

BEGIN;

INSERT INTO agent_definitions (type, display_name, description, category, status, is_active, default_config)
SELECT
  'rerender-chrome',
  'Rerender Chrome (stamp-only)',
  'Force re-renders the three site chrome slots (header, footer, head) for one site so the store stamps rendered_html_digest (bugs_open/226). Deliberately does NOT touch pages, JS snippets or deploys — the narrow tool the six page-scale agents are not. Dispatch with input_data {site_id, domain}.',
  'builder',
  'experimental',
  true,
  jsonb_build_object('workflow', jsonb_build_object(
    'start_step', 'render_site_components',
    'steps', jsonb_build_object(
      'render_site_components', jsonb_build_object(
        'action', 'render_site_components',
        'config', jsonb_build_object(
          'slots', jsonb_build_array('header', 'footer', 'head'),
          'input_fields', jsonb_build_array('site_id', 'domain'),
          'force_rerender', true
        ),
        'next_step', 'complete',
        'description', 'Force re-render header, footer and head; the guarded store UPDATE stamps rendered_html_digest in the same statement',
        'output_field', 'site_components_result'
      ),
      'complete', jsonb_build_object(
        'action', 'complete_workflow',
        'config', jsonb_build_object('output_fields', jsonb_build_array('site_components_result')),
        'description', 'Chrome re-rendered and stamped. No page assembly, no deploys - served pages are untouched until their own next rerender.'
      )
    )
  ))
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions
  WHERE type = 'rerender-chrome' AND deleted_at IS NULL
);

DO $$
DECLARE
  cfg jsonb;
  n_steps integer;
BEGIN
  SELECT default_config INTO cfg
  FROM agent_definitions
  WHERE type = 'rerender-chrome' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cfg IS NULL THEN
    RAISE EXCEPTION 'verify: rerender-chrome row absent or inactive after insert';
  END IF;

  SELECT count(*) INTO n_steps
  FROM jsonb_object_keys(cfg->'workflow'->'steps');
  IF n_steps <> 2 THEN
    RAISE EXCEPTION 'verify: expected exactly 2 workflow steps, found %', n_steps;
  END IF;

  IF (cfg->'workflow'->'steps'->'render_site_components'->'config'->>'force_rerender') IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION 'verify: force_rerender is not true — a non-forced render skips every slot that already has HTML and stamps nothing';
  END IF;

  IF jsonb_array_length(cfg->'workflow'->'steps'->'render_site_components'->'config'->'slots') <> 3 THEN
    RAISE EXCEPTION 'verify: expected all 3 chrome slots (a partial list leaves slots permanently unstamped — the nav-link-fixer gap)';
  END IF;
END $$;

COMMIT;
