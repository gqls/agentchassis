-- 209_report_pipeline_agents.sql
-- Gripper dossier pilot — the three agent definitions.
-- Workstream: docs/agent_docs/docs024_key_docs_latest/robot_hands_gripper_dossier/
-- Design of record: DESIGN_2026-07-24_gripper_dossier_pilot.md §3 (agents).
--
-- ############################################################################
-- ## STRICTLY POST-IMAGE. Apply only AFTER a chassis image carrying the new  ##
-- ## actions is rolled and verified against a running pod.                   ##
-- ############################################################################
-- These workflows name actions that do not exist in any earlier image:
--   pull_report_requests · score_grippers · verify_report_prose ·
--   create_report_page · emit_report_status_files · fail_workflow
-- A workflow naming an unregistered action fails at RUNTIME with
-- WORKFLOW_INVALID — and that failure is precisely the shape complete_work_item
-- was hardened against (a saga that never ran being stamped 'complete'). Verify
-- the roll first, with a string the new code CREATED plus a positive control:
--   kubectl exec -n ai-persona-system <pod> -- sh -c \
--     'strings /app/agent-chassis | grep -c "report-dossier"'      # expect >0
--     'strings /app/agent-chassis | grep -c "pull_report_requests"' # expect >0
--     'strings /app/agent-chassis | grep -c "complete_workflow"'    # control
--
-- Seed 207 (the report-dossier component) must also be applied — compose_page
-- resolves it by function and fails loudly if it is missing.
--
-- IDEMPOTENT: ON CONFLICT (type, version) DO UPDATE.

\set ON_ERROR_STOP on

BEGIN;

-- ============================================================================
-- 1. report-request-collector — pull pending requests from the island.
--    Single step; the action loops every configured site itself (the
--    intent-collector shape: complexity in Go, the workflow stays one step).
-- ============================================================================
INSERT INTO agent_definitions (
    type, version, display_name, description, category, agent_category,
    image_repository, image_tag, command, resources,
    default_config, capabilities, topics
) VALUES (
    'report-request-collector', 1,
    'Report Request Collector',
    'Pulls pending gripper-dossier requests from each site''s report island '
    || '(deploy_config.report_island) and creates one report_request work item '
    || 'per request. One-way: the cluster is never called by the public, and '
    || 'the island payload deliberately carries no visitor email.',
    'business_intel', NULL,
    'docker.io/aqls/agent-chassis',
    'v1.0.1159',
    ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
    '{"requests": {"cpu": "100m", "memory": "256Mi"},
      "limits":   {"cpu": "500m", "memory": "512Mi"}}'::jsonb,
    '{
      "processing_mode": "orchestrator",
      "workflow": {
        "start_step": "pull",
        "processing_mode": "orchestrator",
        "timeout_seconds": 600,
        "steps": {
          "pull": {
            "action": "pull_report_requests",
            "description": "GET {base_url}/requests?since=<checkpoint> for every report-island site; insert deduped work items.",
            "config": {},
            "output_field": "pull_result",
            "next_step": "notify_scheduler"
          },
          "notify_scheduler": {
            "action": "query_database",
            "description": "Tell the scheduler this execution finished so it can fire again.",
            "config": {
              "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''report-request-pull''",
              "output_format": "object"
            },
            "output_field": "scheduler_notified",
            "next_step": "complete"
          },
          "complete": {
            "action": "complete_workflow",
            "description": "Pull tick complete.",
            "config": {"output_fields": ["pull_result"]}
          }
        }
      }
    }'::jsonb,
    '["report-island", "pull", "work-items"]'::jsonb,
    '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb
)
ON CONFLICT (type, version) DO UPDATE SET
    display_name   = EXCLUDED.display_name,
    description    = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    capabilities   = EXCLUDED.capabilities,
    topics         = EXCLUDED.topics,
    is_active      = true,
    updated_at     = NOW();

-- ============================================================================
-- 2. report-dispatch-loop — claim ONE queued request and run its handler.
--    Cloned from the live diagnose-dispatch-loop (its own reaper, its own
--    custom statuses, FOR UPDATE SKIP LOCKED, notify_scheduler). Deliberately
--    NOT claim_work_item: that action claims only triaged/approved, the
--    statuses that expose an item to build-dispatch-loop.
--
--    NOTE on the failure contract: mark_failed fires from call_handler's
--    error_step, which is reached when the handler workflow ENDS IN FAILURE.
--    The report-builder's own failure path publishes the island sidecar and
--    then calls fail_workflow precisely so that this branch is taken — a
--    handler that tidied up and then reported success would be stamped
--    'complete' beside the evidence it failed.
-- ============================================================================
INSERT INTO agent_definitions (
    type, version, display_name, description, category, agent_category,
    image_repository, image_tag, command, resources,
    default_config, capabilities, topics
) VALUES (
    'report-dispatch-loop', 1,
    'Report Dispatch Loop',
    'Claims one awaiting_report work item per tick and runs its handler '
    || '(report-builder). Owns its own reaper because ''reporting'' is inert '
    || 'to the generic claimed-item timeout.',
    'orchestrator', 'coordinator',
    'docker.io/aqls/agent-chassis',
    'v1.0.1159',
    ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
    '{"requests": {"cpu": "100m", "memory": "256Mi"},
      "limits":   {"cpu": "500m", "memory": "512Mi"}}'::jsonb,
    '{
      "workflow": {
        "start_step": "reap_stuck",
        "processing_mode": "orchestrator",
        "timeout_seconds": 2400,
        "steps": {
          "reap_stuck": {
            "action": "query_database",
            "description": "Fail any report build whose pod died. We own this because ''reporting'' is inert to the generic claimed-item timeout.",
            "config": {
              "query": "UPDATE site_work_items SET status = ''failed'', attempt_count = attempt_count + 1, error = ''report build exceeded 30m — handler pod likely died'', claimed_at = NULL WHERE pipeline = ''reports'' AND status = ''reporting'' AND claimed_at < NOW() - INTERVAL ''30 minutes'' RETURNING id::text",
              "output_format": "object"
            },
            "output_field": "reaped",
            "next_step": "claim_item"
          },
          "claim_item": {
            "action": "query_database",
            "description": "Atomically take ONE awaiting_report item into reporting.",
            "config": {
              "query": "UPDATE site_work_items SET status = ''reporting'', claimed_by = ''report-dispatch-loop'', claimed_at = NOW() WHERE id = (SELECT id FROM site_work_items WHERE pipeline = ''reports'' AND item_type = ''report_request'' AND status = ''awaiting_report'' AND attempt_count < max_attempts ORDER BY priority ASC, created_at ASC FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id::text AS work_item_id, handler_agent, site_id::text AS target_site_id, spec->>''request_id'' AS request_id, (SELECT domain FROM sites WHERE sites.id = site_work_items.site_id) AS domain",
              "output_format": "object"
            },
            "output_field": "claimed",
            "next_step": "check_claimed"
          },
          "check_claimed": {
            "action": "conditional",
            "description": "Nothing queued (count = 0) is the normal case: tell the scheduler and finish.",
            "config": {
              "condition": "claimed.count > 0",
              "then_step": "spawn_handler",
              "else_step": "notify_scheduler"
            }
          },
          "spawn_handler": {
            "action": "spawn_agent",
            "description": "Spawn the handler named on the item (report-builder).",
            "config": {
              "role": "handler",
              "agent_type_field": "claimed.handler_agent",
              "error_step": "mark_failed"
            },
            "output_field": "handler_spawned",
            "next_step": "call_handler"
          },
          "call_handler": {
            "action": "call_agent",
            "description": "Build the dossier. 20 minutes: the LLM prose step dominates.",
            "config": {
              "target_role": "handler",
              "input_mapping": {
                "work_item_id": "claimed.work_item_id",
                "site_id": "claimed.target_site_id",
                "request_id?": "claimed.request_id",
                "domain?": "claimed.domain"
              },
              "timeout_seconds": 1200,
              "error_step": "mark_failed"
            },
            "output_field": "handler_result",
            "next_step": "mark_complete"
          },
          "mark_complete": {
            "action": "complete_work_item",
            "description": "Mark the request complete, carrying the build result. The action''s own guard re-checks the handler verdict before stamping complete.",
            "config": {
              "work_item_id": "claimed.work_item_id",
              "result": "handler_result"
            },
            "output_field": "item_completed",
            "next_step": "notify_scheduler"
          },
          "mark_failed": {
            "action": "fail_work_item",
            "description": "Handler died, timed out, or ended in a deliberate failure. max_attempts=1 means this is terminal.",
            "config": {
              "work_item_id": "claimed.work_item_id",
              "error_message": "report handler failed or timed out"
            },
            "output_field": "item_failed",
            "next_step": "notify_scheduler"
          },
          "notify_scheduler": {
            "action": "query_database",
            "description": "Tell the scheduler this execution finished so it can fire again.",
            "config": {
              "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''report-dispatch''",
              "output_format": "object"
            },
            "output_field": "scheduler_notified",
            "next_step": "complete"
          },
          "complete": {
            "action": "complete_workflow",
            "description": "Dispatch tick complete.",
            "config": {"output_fields": ["claimed", "handler_result", "reaped"]}
          }
        }
      }
    }'::jsonb,
    '["dispatch", "orchestration", "work-items", "reports"]'::jsonb,
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb
)
ON CONFLICT (type, version) DO UPDATE SET
    display_name   = EXCLUDED.display_name,
    description    = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    capabilities   = EXCLUDED.capabilities,
    topics         = EXCLUDED.topics,
    is_active      = true,
    updated_at     = NOW();

-- ============================================================================
-- 3. report-builder — the handler. Deterministic scoring, LLM prose bounded
--    by the fact block, a hard verification gate, then compose → validate →
--    render → deploy → sidecar.
--
--    ORDERING IS LOAD BEARING at two points:
--      * verify_prose runs BEFORE compose_page. The gate is worth having only
--        if nothing unverified can reach a page.
--      * publish_ready commits the sidecar AFTER deploy_page has committed the
--        page. The island polls ONLY the sidecar, so a sidecar that landed
--        first would tell a visitor the report is ready before the artefact
--        exists (the gauntlet misstep-4 lesson: status is not the artefact).
--
--    validate_page_content runs with check_claims:false BY DESIGN: its
--    evidence-base number check compares against the SITE register, and every
--    figure on a report is computed per request, so check 8 would fail every
--    honest report. The per-request equivalent is verify_report_prose, which
--    binds prose to this run''s fact_block instead — including, since council
--    round 2, vendor names that carry no digits.
-- ============================================================================
INSERT INTO agent_definitions (
    type, version, display_name, description, category, agent_category,
    image_repository, image_tag, command, resources,
    default_config, capabilities, topics
) VALUES (
    'report-builder', 1,
    'Gripper Dossier Report Builder',
    'Builds one Gripper Selection & Integration Dossier: deterministic '
    || 'MatchMatrix v2 scoring over the site''s verified gripper index, prose '
    || 'bounded by the resulting fact block and gated deterministically, then '
    || 'a composed, validated, deployed report page plus the island status '
    || 'sidecar. An honest no-match report is a SUCCESS.',
    'reports', 'coordinator',
    'docker.io/aqls/agent-chassis',
    'v1.0.1159',
    ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
    '{"requests": {"cpu": "200m", "memory": "512Mi"},
      "limits":   {"cpu": "1000m", "memory": "1Gi"}}'::jsonb,
    '{
      "workflow": {
        "start_step": "load_request",
        "processing_mode": "orchestrator",
        "timeout_seconds": 1200,
        "steps": {
          "load_request": {
            "action": "query_database",
            "description": "Load the claimed work item and its spec as flat columns (precedent: fix-implementer.load_plan).",
            "config": {
              "query": "SELECT swi.id::text AS work_item_id, swi.site_id::text AS site_id, s.domain AS domain, swi.spec->>''request_id'' AS request_id, swi.spec->>''mass_kg'' AS mass_kg, swi.spec->>''travel_mm'' AS travel_mm, swi.spec->>''accel_ms2'' AS accel_ms2, swi.spec->>''surface_material'' AS surface_material, swi.spec->>''surfaces_n'' AS surfaces_n, swi.spec->>''safety_factor'' AS safety_factor, swi.spec->>''cycle_rate'' AS cycle_rate, swi.spec->>''ip_min'' AS ip_min, swi.spec->>''mounting'' AS mounting, swi.spec->>''part_geometry'' AS part_geometry, swi.spec->>''application'' AS application, swi.spec->>''submitted_at'' AS submitted_at FROM site_work_items swi JOIN sites s ON s.id = swi.site_id WHERE swi.id = $1::uuid",
              "params": ["input_data.work_item_id"],
              "output_format": "object"
            },
            "output_field": "request",
            "next_step": "score"
          },
          "score": {
            "action": "score_grippers",
            "description": "Deterministic MatchMatrix v2 physics. Emits the fact_block that bounds every number and name in the prose. A malformed spec is a hard error here, never a guessed default.",
            "config": {
              "site_id": "request.site_id",
              "mass_kg": "request.mass_kg",
              "travel_mm": "request.travel_mm",
              "surface_material": "request.surface_material",
              "surfaces_n": "request.surfaces_n",
              "accel_ms2": "request.accel_ms2",
              "safety_factor": "request.safety_factor",
              "cycle_rate": "request.cycle_rate",
              "ip_min": "request.ip_min",
              "error_step": "handle_failure"
            },
            "output_field": "scoring",
            "next_step": "write_prose"
          },
          "write_prose": {
            "action": "execute_llm_prompt",
            "description": "Write the four prose sections AROUND the computed facts. The fact block is the only permitted source of numbers and names.",
            "config": {
              "ai_service": {
                "provider": "anthropic",
                "model": "claude-sonnet-5",
                "max_tokens": 16000,
                "api_key_env_var": "ANTHROPIC_API_KEY"
              },
              "temperature": 0.2,
              "output_format": "json",
              "input_fields": ["scoring", "request"],
              "error_step": "handle_failure",
              "prompt_template": "You are a robotics applications engineer writing one section of a gripper selection dossier for a named customer application. You are writing for an engineer who will be held responsible for the choice.\n\n## THE ONLY FACTS YOU MAY USE\n\nEverything below is computed or quoted from manufacturer-published specifications. It is the complete set of numbers and product names available to you.\n\n{{.scoring.fact_block}}\n\n## THE CUSTOMER''S APPLICATION (context you may echo)\n\nMounting: {{.request.mounting}}\nPart geometry: {{.request.part_geometry}}\nApplication notes: {{.request.application}}\n\n## HARD RULES — a violation fails this run, it is not softened\n\nA deterministic gate checks your output before it can reach a page. It rejects the whole report on any of these, so there is no partial credit for a good paragraph containing one invented figure.\n\n1. NUMBERS. Every number you write must appear in the fact block above (or in the application context). Do not compute new figures, do not round to a ''nicer'' number, do not convert units, do not estimate, do not average. If you want to make a point that needs a number you do not have, make the point without the number or leave it out.\n\n2. NAMES. Name only the products and manufacturers that appear in the fact block. Do NOT mention any other vendor — not as an alternative, not as a comparison, not as ''you might also consider''. Naming a vendor that was not assessed fails the run. Do not invent sibling model numbers: if the fact block lists one model of a range, that is the only model of that range you may name.\n\n3. UNPUBLISHED FIGURES. Where the fact block marks a figure as NOT PUBLISHED by the manufacturer, you may say that it is not published — and you must never fill the gap with an estimate, a typical value, or an inference from a similar product. ''The manufacturer does not publish an IP rating for this unit'' is a useful sentence. Guessing one is the failure this whole system exists to prevent.\n\n4. NO MATCH. If the fact block opens with the sentence stating that no gripper in the index meets the requirement, you MUST reproduce that sentence VERBATIM in your summary, and you must not soften it anywhere in the report. Do not write that something ''nearly meets'' or ''could meet with adjustment'' or ''is suitable for your application''. Explain the shortfall using the computed figures and say what would have to change. An honest no-match is a correct, valuable answer and is treated as a successful report — a softened one is a failure.\n\n5. NO PURCHASE RECOMMENDATION. You are not recommending a purchase. Present the assessment; let the engineer decide. Never write ''we recommend buying/ordering/purchasing''.\n\n6. NO HTML BEYOND SIMPLE TEXT MARKUP. Use only <p>, <ul>, <li>, <strong>, <em>. No <script>, no <style>, no <a>, no headings — the page supplies its own structure.\n\n## WHAT TO WRITE\n\nReturn ONLY a JSON object with exactly these four keys, each a string of HTML:\n\n- summary_html: 2-3 short paragraphs. What was asked, what the physics requires, what the assessment found. Lead with the answer.\n- candidates_html: the shortlist discussed candidate by candidate, using the computed headroom and the published figures. Say plainly where a candidate fails and why. Where a figure is unpublished, say so.\n- integration_html: practical integration notes for the leading candidate(s) — mounting, controls, what to check on first articles. Ground every claim in the fact block or the application context.\n- vendor_questions_html: a <ul> of the specific questions this engineer should put to the vendor before ordering, especially the ones created by unpublished figures.\n\nWrite plainly. Short sentences. No marketing language, no filler, no hedging that hides the answer. British English."
            },
            "output_field": "report_prose",
            "next_step": "verify_prose"
          },
          "verify_prose": {
            "action": "verify_report_prose",
            "description": "The deterministic gate. Numbers, SKU-shaped names, unassessed vendors, the verbatim no-match sentence, empty sections, and a tolerated truncation are all hard failures here — BEFORE anything reaches a page.",
            "config": {
              "prose_field": "report_prose.result",
              "scoring_field": "scoring",
              "context_field": "request",
              "error_step": "handle_failure"
            },
            "output_field": "prose_verified",
            "next_step": "compose_page"
          },
          "compose_page": {
            "action": "create_report_page",
            "description": "Compose and persist the page: request echo, printed formulas, SVG headroom chart, the verified prose, scored candidate cards, provenance footer. Nav-invisible, rebuild_policy=owned.",
            "config": {
              "site_id": "request.site_id",
              "request_id": "request.request_id",
              "scoring_field": "scoring",
              "prose_field": "report_prose.result",
              "error_step": "handle_failure"
            },
            "output_field": "composed",
            "next_step": "validate_page"
          },
          "validate_page": {
            "action": "validate_page_content",
            "description": "Structural/contamination checks. check_claims MUST stay false — see the header: every figure here is per-request and would fail the site-register check by construction.",
            "config": {
              "html_field": "composed.rendered_html",
              "domain": "composed.domain",
              "check_claims": false,
              "check_emails": true,
              "check_internal_links": true,
              "error_step": "handle_failure"
            },
            "output_field": "page_validated",
            "next_step": "render_page"
          },
          "render_page": {
            "action": "rerender_single_page",
            "description": "Assemble the page from stored components (concatenates rendered_html; does not re-render templates).",
            "config": {
              "input_fields": ["page_id", "site_id", "domain"],
              "page_id": "composed.page_id",
              "site_id": "composed.site_id",
              "domain": "composed.domain",
              "error_step": "handle_failure"
            },
            "output_field": "rendered_page",
            "next_step": "check_skipped"
          },
          "check_skipped": {
            "action": "conditional",
            "description": "A report page with no components assembled is a FAILURE, not a skip — the visitor is waiting on an artefact.",
            "config": {
              "condition": "rendered_page.skipped == true",
              "then_step": "handle_failure",
              "else_step": "deploy_page"
            }
          },
          "deploy_page": {
            "action": "git_commit",
            "description": "Commit the report page. This is the artefact the island will poll for.",
            "config": {
              "files_field": "rendered_page.files",
              "domain_field": "rendered_page.domain",
              "commit_message": "Gripper dossier report {{.filename}}",
              "error_step": "handle_failure"
            },
            "output_field": "deploy_result",
            "next_step": "update_status"
          },
          "update_status": {
            "action": "update_page_status",
            "description": "Mark the report page deployed.",
            "config": {
              "status": "deployed",
              "commit_from": "deploy_result.commit_sha",
              "page_id_field": "composed.page_id"
            },
            "next_step": "emit_ready"
          },
          "emit_ready": {
            "action": "emit_report_status_files",
            "description": "Build the ready sidecar. Requires page_url, so it cannot claim ready without one.",
            "config": {
              "status": "ready",
              "request_id": "request.request_id",
              "page_url": "composed.page_url",
              "error_step": "handle_failure"
            },
            "output_field": "ready_sidecar",
            "next_step": "publish_ready"
          },
          "publish_ready": {
            "action": "git_commit",
            "description": "Commit the sidecar AFTER the page. The island polls only this file, so ready can never precede the artefact.",
            "config": {
              "files_field": "ready_sidecar.files",
              "domain_field": "composed.domain",
              "commit_message": "Gripper dossier ready {{.filename}}",
              "error_step": "handle_failure"
            },
            "output_field": "sidecar_deployed",
            "next_step": "complete"
          },
          "complete": {
            "action": "complete_workflow",
            "description": "Report built and published. An honest no-match report reaches HERE — it is a success.",
            "config": {"output_fields": ["composed", "deploy_result", "sidecar_deployed"]}
          },

          "handle_failure": {
            "action": "emit_report_status_files",
            "description": "FAILURE PATH. Build the failed sidecar so the island can tell the visitor promptly rather than waiting out its 24h expiry.",
            "config": {
              "status": "failed",
              "request_id": "request.request_id",
              "error_step": "fail_out"
            },
            "output_field": "failed_sidecar",
            "next_step": "publish_failed"
          },
          "publish_failed": {
            "action": "git_commit",
            "description": "Publish the failed sidecar. Best effort: if this cannot commit we still end in failure.",
            "config": {
              "files_field": "failed_sidecar.files",
              "domain_field": "request.domain",
              "commit_message": "Gripper dossier failed {{.filename}}",
              "error_step": "fail_out"
            },
            "output_field": "failed_sidecar_deployed",
            "next_step": "fail_out"
          },
          "fail_out": {
            "action": "fail_workflow",
            "description": "End in a deliberate FAILURE verdict, AFTER the sidecar attempt. This is what makes the dispatch loop take call_handler''s error_step and mark the item failed — a run that tidied up and then reported success would be stamped complete beside the evidence it failed.",
            "config": {
              "reason": "gripper dossier build failed — see the step error and agent_error_log"
            }
          }
        }
      }
    }'::jsonb,
    '["reports", "orchestration", "grippers", "deterministic-scoring"]'::jsonb,
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb
)
ON CONFLICT (type, version) DO UPDATE SET
    display_name   = EXCLUDED.display_name,
    description    = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    capabilities   = EXCLUDED.capabilities,
    topics         = EXCLUDED.topics,
    is_active      = true,
    updated_at     = NOW();

-- ============================================================================
-- Assertions: the three agents exist, and every action they name is one this
-- image is expected to carry. The second check is a spelling check, not a
-- registration check — only the pod can confirm registration (see header).
-- ============================================================================
DO $$
DECLARE
    n            INTEGER;
    named_action TEXT;
    known        TEXT[] := ARRAY[
        'pull_report_requests', 'score_grippers', 'verify_report_prose',
        'create_report_page', 'emit_report_status_files', 'fail_workflow',
        'query_database', 'execute_llm_prompt', 'validate_page_content',
        'rerender_single_page', 'git_commit', 'update_page_status',
        'conditional', 'spawn_agent', 'call_agent', 'complete_work_item',
        'fail_work_item', 'complete_workflow'
    ];
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type IN ('report-request-collector', 'report-dispatch-loop', 'report-builder')
      AND version = 1 AND is_active = true
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF n <> 3 THEN
        RAISE EXCEPTION 'expected 3 active report agents, found %', n;
    END IF;

    FOR named_action IN
        SELECT DISTINCT v->>'action'
        FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') AS e(k, v)
        WHERE d.type IN ('report-request-collector', 'report-dispatch-loop', 'report-builder')
          AND d.version = 1
    LOOP
        IF NOT (named_action = ANY(known)) THEN
            RAISE EXCEPTION
                'workflow names action % which is not in this seed''s expected set — a typo here fails at RUNTIME as WORKFLOW_INVALID', named_action;
        END IF;
    END LOOP;

    RAISE NOTICE 'report pipeline agents seeded; all named actions spelled as expected';
END $$;

COMMIT;

-- Post-apply sanity (the seed cannot check these — only the pod can):
--   1. Every action above is registered in the RUNNING image (pod-grep, header).
--   2. Seed 207 applied: SELECT count(*) FROM content_components
--      WHERE function='report-dossier' AND is_active;  -- expect 1
--   3. Then seed 210 for the scheduled tasks (both ship disabled).
