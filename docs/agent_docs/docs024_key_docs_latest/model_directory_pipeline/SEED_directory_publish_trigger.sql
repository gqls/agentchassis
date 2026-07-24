-- SEED — model-directory publish trigger (Phase D final leg)
--
-- The piece that was PLANNED but not seeded with the rest of Phase D
-- (noticed 2026-07-24, the day the registry got its first verified rows):
-- without it, nothing commits data/model-directory.json into opted-in
-- sites' repos or queues the scoped rerenders that keep the baked HTML
-- fresh. Mirrors the content-feed-trigger → content-feed-orchestrator
-- pattern exactly (trigger finds due sites, loops spawn+call of a per-site
-- worker agent).
--
-- APPLY ONLY AFTER an image carrying `render_model_directory` is deployed
-- and pod-verified (live since agent-chassis v1.0.1149; re-verify anyway):
--   kubectl -n ai-persona-system exec <agent-chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c render_model_directory'
--
-- SELF-GATING: the find-sites query requires BOTH the site_specs opt-in
-- flag AND a deployed page carrying a model-directory component. Until the
-- auto-created page exists (discovery checks → content-gap-planner →
-- page-build-handler), every cycle completes idle — harmless by design.
--
-- Config-only: live immediately, no image roll. All three inserts are
-- guarded (WHERE NOT EXISTS) — idempotent on replay.

-- ── 1. Per-site publisher: render the registry JSON, commit, done ─────────
INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status, is_active,
    image_repository, image_tag, input_contract, output_contract, default_config
)
SELECT
    'model-directory-publisher',
    'Model Directory Publisher',
    'Per-site publish leg of the model directory pipeline: renders the global directory_entities/directory_claims registry to data/model-directory.json (and the full listing file when the site has a listing page), commits via git-adapter, and queues scoped page rerenders so the server-rendered HTML tracks the registry.',
    'orchestrator', 'coordinator', 'active', true,
    'docker.io/aqls/agent-chassis',
    (SELECT image_tag FROM agent_definitions WHERE type = 'directory-researcher'),
    '{"required": ["site_id", "domain"]}'::jsonb,
    '{"produces": {"directory_commit_result": "model-directory JSON files committed to the site repo"}}'::jsonb,
    $cfg${
  "workflow": {
    "start_step": "render_model_directory_json",
    "processing_mode": "orchestrator",
    "timeout_seconds": 600,
    "steps": {
      "render_model_directory_json": {
        "action": "render_model_directory",
        "config": {"site_id": "input_data.site_id"},
        "next_step": "commit_model_directory",
        "output_field": "directory_render_result",
        "description": "Build data/model-directory.json (+ full listing when a listing page exists) from the global registry"
      },
      "commit_model_directory": {
        "action": "git_commit",
        "config": {
          "files_field": "directory_render_result.files",
          "domain_field": "directory_render_result.domain",
          "commit_message": "Update model directory"
        },
        "next_step": "complete",
        "output_field": "directory_commit_result",
        "description": "Commit the JSON files into the site repo via git-adapter"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {"output_fields": ["directory_render_result", "directory_commit_result"]}
      }
    }
  }
}$cfg$::jsonb
WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'model-directory-publisher' AND deleted_at IS NULL);

-- ── 2. Trigger: find opted-in sites with a live component, loop publisher ─
INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status, is_active,
    image_repository, image_tag, input_contract, output_contract, default_config
)
SELECT
    'model-directory-trigger',
    'Model Directory Trigger',
    'Scheduled fan-out for the model directory publish leg: finds sites that opted in (site_specs classification content_features.model_directory) AND have a deployed page carrying a model-directory component, then spawn+calls model-directory-publisher per site. Mirrors content-feed-trigger.',
    'orchestrator', 'coordinator', 'active', true,
    'docker.io/aqls/agent-chassis',
    (SELECT image_tag FROM agent_definitions WHERE type = 'directory-researcher'),
    '{}'::jsonb,
    '{"produces": {"publish_results": "one publisher run per due site"}}'::jsonb,
    $cfg${
  "workflow": {
    "start_step": "find_directory_sites",
    "processing_mode": "orchestrator",
    "timeout_seconds": 900,
    "steps": {
      "find_directory_sites": {
        "action": "query_database",
        "config": {
          "query": "SELECT DISTINCT s.id::text AS site_id, s.domain FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'model_directory'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p JOIN page_components pc ON pc.page_id = p.id JOIN content_components cc ON cc.id = pc.component_id WHERE p.site_id = s.id AND p.build_status = 'deployed' AND cc.function IN ('model-directory', 'model-directory-listing')) AND EXISTS (SELECT 1 FROM directory_claims WHERE is_current AND status = 'found') ORDER BY s.domain LIMIT 5",
          "output_format": "object"
        },
        "next_step": "check_has_sites",
        "output_field": "directory_sites",
        "description": "Opted-in sites with a deployed model-directory component, only while the registry has publishable claims"
      },
      "check_has_sites": {
        "action": "evaluate_condition",
        "config": {
          "condition_field": "directory_sites.count",
          "conditions": {"0": "notify_scheduler_idle"},
          "default": "process_sites"
        },
        "description": "Skip if no sites are due"
      },
      "process_sites": {
        "action": "loop",
        "config": {
          "items_field": "directory_sites.rows",
          "item_variable": "current_site",
          "max_iterations": 5,
          "continue_on_error": true,
          "sub_workflow": {
            "start_step": "spawn_publisher",
            "steps": {
              "spawn_publisher": {
                "action": "spawn_agent",
                "config": {"role": "directory_publisher", "agent_type": "model-directory-publisher"},
                "next_step": "call_publisher",
                "output_field": "publisher_spawned",
                "description": "Spawn model-directory-publisher for this site"
              },
              "call_publisher": {
                "action": "call_agent",
                "config": {
                  "target_role": "directory_publisher",
                  "input_mapping": {"site_id": "current_site.site_id", "domain": "current_site.domain"},
                  "timeout_seconds": 600
                },
                "next_step": "done",
                "error_step": "done",
                "output_field": "publish_result",
                "description": "Publish the model directory for this site"
              },
              "done": {"action": "loop_complete", "description": "Next site"}
            }
          }
        },
        "next_step": "notify_scheduler",
        "output_field": "publish_results",
        "description": "Publish per due site"
      },
      "notify_scheduler": {
        "action": "query_database",
        "config": {"query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = 'model-directory-publish'", "output_format": "object"},
        "next_step": "complete",
        "output_field": "scheduler_notified",
        "description": "Stamp completion"
      },
      "notify_scheduler_idle": {
        "action": "query_database",
        "config": {"query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = 'model-directory-publish'", "output_format": "object"},
        "next_step": "complete_idle",
        "output_field": "scheduler_notified",
        "description": "Stamp completion (idle)"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {"output_fields": ["directory_sites", "publish_results"]}
      },
      "complete_idle": {
        "action": "complete_workflow",
        "config": {"output_fields": ["directory_sites"], "success_message": "No sites due for model-directory publish"}
      }
    }
  }
}$cfg$::jsonb
WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'model-directory-trigger' AND deleted_at IS NULL);

-- ── 3. Scheduled task (same cadence as news publish, 6h) ──────────────────
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'model-directory-publish',
    'Model directory pipeline: publish data/model-directory.json + scoped rerenders to every opted-in site with a deployed model-directory component. Self-gating (idles until pages exist and the registry has found claims).',
    21600,
    'model-directory-trigger',
    -- Generic topic, NOT a per-type topic: custom chassis agent types are
    -- dispatched via system.agent.generic.requests with config.agent_type
    -- carrying the real type (verified against content-feed-refresh; the
    -- routing trap this workstream caught pre-ship on 2026-07-22).
    'system.agent.generic.requests',
    '{}'::jsonb,
    'model-directory-publish',
    1,
    true,
    900
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'model-directory-publish');

-- ── Post-apply verification ────────────────────────────────────────────────
-- 1. All three rows exist:
--    SELECT type, is_active FROM agent_definitions
--    WHERE type IN ('model-directory-publisher','model-directory-trigger');
--    SELECT name, enabled FROM scheduled_tasks WHERE name='model-directory-publish';
-- 2. While no page carries the component, each cycle completes idle:
--    the trigger orchestration ends at complete_idle with directory_sites.count=0.
-- 3. Once the auto-created page deploys: data/model-directory.json appears in
--    the site repo; page_rerender items queued (item_key page_rerender:<page>).
--
-- ── Rollback ────────────────────────────────────────────────────────────────
--    UPDATE scheduled_tasks SET enabled=false WHERE name='model-directory-publish';
--    UPDATE agent_definitions SET status='disabled', is_active=false
--    WHERE type IN ('model-directory-publisher','model-directory-trigger');
