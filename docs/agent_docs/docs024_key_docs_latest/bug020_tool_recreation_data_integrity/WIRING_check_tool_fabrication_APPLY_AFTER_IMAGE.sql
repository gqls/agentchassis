-- WIRING_check_tool_fabrication_APPLY_AFTER_IMAGE.sql — the workflow half of the
-- /bugs_open/020 mechanical gate. STAGED, NOT APPLIED, and DELIBERATELY NOT a
-- numbered migration in sql_for_agents/.
--
-- ┌──────────────────────────────────────────────────────────────────────────┐
-- │  IMAGE FIRST. Do NOT apply this until a chassis image carrying the        │
-- │  `check_tool_fabrication` action is live in the pod. This file wires a    │
-- │  workflow step that NAMES that action; on a pod without it the step       │
-- │  errors at runtime (CLAUDE.md: "a seed naming an unregistered action      │
-- │  fails at runtime"). It is kept out of sql_for_agents/ precisely so the   │
-- │  migration runner cannot pick it up on a `--apply` sweep before then.     │
-- └──────────────────────────────────────────────────────────────────────────┘
--
-- Verify the action is in the running pod FIRST (discriminating grep, per CLAUDE.md):
--   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
--         -o jsonpath='{.items[0].metadata.name}')
--   kubectl -n ai-persona-system exec $POD -- sh -c \
--     'strings /app/agent-chassis | grep -c check_tool_fabrication'   # must be >= 1
--
-- THEN: renumber this file into docs/agent_docs/sql_for_agents/1NN_wire_tool_fabrication_gate.sql,
-- apply it out of band (psql -f), and record it with run-migrations.sh --record-only.
--
-- What it wires (inserted after check_completeness, before save_training_data):
--   check_completeness → check_fabrication → route_fabrication
--       fabricated == true  → request_fabrication_review → complete   (NO deploy)
--       else                → save_training_data (the original next step)
--
-- On a fabrication the tool is NEVER saved/deployed; instead checkpoint_for_review
-- raises a needs_human_review item carrying the full collected_data (the generated
-- HTML — which lives nowhere else on this branch — plus the fabrication signals).
--
-- Live step graph this was written against (2026-07-21):
--   check_completeness.next_step = save_training_data
--   save_training_data → validate_tool → save_sections → update_status →
--   spawn_rerender → deploy_page → compose_note → append_note → complete
--
-- Standing rule: snapshot_agent opens the transaction.

BEGIN;

SELECT snapshot_agent('tool-recreation-handler', 'WIRING_check_tool_fabrication: pre-update');

-- Needle-gate (per bugs_open/016b needle discipline): confirm the blast radius
-- BEFORE any mutation. This is jsonb surgery on a LIVE production workflow row, so
-- assert exactly one live row will be touched and that it is NOT already wired
-- (idempotency guard — a re-run must not double-apply). Each UPDATE below also
-- prints "UPDATE 1" (its own row count), and the tail DO block re-verifies the
-- resulting step graph.
DO $$
DECLARE n int; already int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
      WHERE type='tool-recreation-handler' AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION 'needle-gate: expected exactly 1 live tool-recreation-handler row, found % — aborting before mutation', n;
    END IF;
    SELECT count(*) INTO already FROM agent_definitions
      WHERE type='tool-recreation-handler' AND deleted_at IS NULL
        AND default_config #>> '{workflow,steps,check_completeness,next_step}' = 'check_fabrication';
    IF already > 0 THEN
        RAISE EXCEPTION 'needle-gate: gate already wired (check_completeness.next_step=check_fabrication) — nothing to do, aborting';
    END IF;
    RAISE NOTICE 'needle-gate ok: exactly 1 live row, not yet wired — proceeding';
END $$;

-- (1) The detector step.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,check_fabrication}',
      $j${
        "action": "check_tool_fabrication",
        "config": {
          "html_field": "completeness_check.clean_html",
          "original_html_field": "existing_content.existing_content.raw_html",
          "analysis_field": "tool_analysis.result"
        },
        "next_step": "route_fabrication",
        "error_step": "save_training_data",
        "description": "Bug 020 gate: detect an invented/synthetic dataset before deploy",
        "output_field": "fabrication_check"
      }$j$::jsonb, true)
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;

-- (2) The router: fabricated → review, else → the original next step.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,route_fabrication}',
      $j${
        "action": "conditional",
        "config": {
          "condition": "fabrication_check.fabricated == true",
          "then_step": "request_fabrication_review",
          "else_step": "save_training_data"
        },
        "description": "Bug 020: route invented-data recreations to human review, else continue"
      }$j$::jsonb, true)
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;

-- (3) The review checkpoint: raise a needs_human_review item, do NOT deploy.
--     review_fields_from is intentionally omitted so the whole collected_data is
--     saved — the generated HTML exists ONLY here on this branch (save_sections
--     is skipped), so the reviewer must be given it.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,request_fabrication_review}',
      $j${
        "action": "checkpoint_for_review",
        "config": {
          "item_type": "needs_human_review",
          "severity": "high",
          "site_id_from": "site_record.site_id",
          "page_id_from": "page_record.id",
          "summary_from": "Tool recreation for {{.site_record.domain}} appears to INVENT data — held for review (bug 020)"
        },
        "next_step": "complete",
        "description": "Bug 020: raise a needs_human_review item with the fabrication evidence; do NOT deploy"
      }$j$::jsonb, true)
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;

-- (4) Repoint check_completeness onto the gate. RETURNING confirms exactly which
--     row changed and the resulting edge (needle discipline).
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,check_completeness,next_step}',
      '"check_fabrication"'::jsonb, false)
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL
RETURNING id, type, version,
          default_config #>> '{workflow,steps,check_completeness,next_step}' AS check_completeness_next_step;

-- Verify.
DO $$
DECLARE cfg jsonb; n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
      WHERE type='tool-recreation-handler' AND deleted_at IS NULL;
    IF n <> 1 THEN RAISE EXCEPTION 'expected exactly one live row, found %', n; END IF;

    SELECT default_config INTO cfg FROM agent_definitions
      WHERE type='tool-recreation-handler' AND deleted_at IS NULL;

    IF cfg #>> '{workflow,steps,check_completeness,next_step}' <> 'check_fabrication' THEN
        RAISE EXCEPTION 'check_completeness.next_step not repointed to check_fabrication';
    END IF;
    IF cfg #>> '{workflow,steps,check_fabrication,action}' <> 'check_tool_fabrication' THEN
        RAISE EXCEPTION 'check_fabrication step missing';
    END IF;
    IF cfg #>> '{workflow,steps,route_fabrication,config,then_step}' <> 'request_fabrication_review' THEN
        RAISE EXCEPTION 'route_fabrication router missing';
    END IF;
    IF cfg #>> '{workflow,steps,request_fabrication_review,action}' <> 'checkpoint_for_review' THEN
        RAISE EXCEPTION 'request_fabrication_review step missing';
    END IF;
    RAISE NOTICE 'bug 020 fabrication gate wired into tool-recreation-handler';
END $$;

COMMIT;
