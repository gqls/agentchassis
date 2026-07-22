-- 189_wire_tool_fabrication_gate.sql — wire the /bugs_open/020 fabrication gate
-- into tool-recreation-handler's workflow. DB-only; effective immediately.
--
-- IMAGE-FIRST SATISFIED: this names the `check_tool_fabrication` action, which is
-- confirmed live in the running pod BEFORE applying (2026-07-22, agent-chassis
-- v1.0.1146, pod agent-chassis-687cdf6db5-fq2fd):
--   strings /app/agent-chassis | grep -c check_tool_fabrication   => 4
--   strings /app/agent-chassis | grep -c corroborated_corpus      => 1  (Tier B live)
-- NOTE: `uninspectable` => 0 in this image — the fail-SAFE hardening (commit
-- 37d3bb119) is NOT in v1.0.1146; this pod runs the fail-OPEN detector. That is an
-- edge-case only (empty/missing output, which is not fabrication — real fabricated
-- content is non-empty and still gets inspected), so wiring now still closes the
-- core bug-020 hole; the fail-safe improves the SAME wired action on the next roll.
--
-- What it wires (after check_completeness, before save_training_data):
--   check_completeness -> check_fabrication -> route_fabrication
--       fabricated == true  -> request_fabrication_review -> complete   (NO deploy)
--       else                -> save_training_data (the original next step)
-- On a fabrication the tool is NEVER saved/deployed; checkpoint_for_review raises a
-- needs_human_review item carrying the full collected_data (the generated HTML lives
-- nowhere else on this branch) + the fabrication signals.
--
-- This is the needle-gated content of
-- docs024_key_docs_latest/bug020_tool_recreation_data_integrity/WIRING_check_tool_fabrication_APPLY_AFTER_IMAGE.sql
-- promoted to a numbered migration now that the image is live. Standing rule:
-- snapshot_agent opens the transaction.

BEGIN;

SELECT snapshot_agent('tool-recreation-handler', '189_wire_tool_fabrication_gate.sql: pre-update');

-- Needle-gate: confirm blast radius BEFORE mutating. Exactly one live row, and NOT
-- already wired (idempotency guard — a re-run must not double-apply).
DO $$
DECLARE n int; already int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
      WHERE type='tool-recreation-handler' AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION 'needle-gate: expected exactly 1 live tool-recreation-handler row, found % — aborting', n;
    END IF;
    SELECT count(*) INTO already FROM agent_definitions
      WHERE type='tool-recreation-handler' AND deleted_at IS NULL
        AND default_config #>> '{workflow,steps,check_completeness,next_step}' = 'check_fabrication';
    IF already > 0 THEN
        RAISE EXCEPTION 'needle-gate: gate already wired — nothing to do, aborting';
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

-- (2) The router: fabricated -> review, else -> the original next step.
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
--     review_fields_from omitted on purpose so the whole collected_data is saved —
--     the generated HTML exists ONLY here on this branch (save_sections is skipped).
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

-- (4) Repoint check_completeness onto the gate. RETURNING confirms the row + edge.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,check_completeness,next_step}',
      '"check_fabrication"'::jsonb, false)
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL
RETURNING id, type, version,
          default_config #>> '{workflow,steps,check_completeness,next_step}' AS check_completeness_next_step;

-- Verify the resulting step graph.
DO $$
DECLARE cfg jsonb; n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
      WHERE type='tool-recreation-handler' AND deleted_at IS NULL;
    IF n <> 1 THEN RAISE EXCEPTION 'expected exactly one live row, found %', n; END IF;

    SELECT default_config INTO cfg FROM agent_definitions
      WHERE type='tool-recreation-handler' AND deleted_at IS NULL;

    IF cfg #>> '{workflow,steps,check_completeness,next_step}' <> 'check_fabrication' THEN
        RAISE EXCEPTION 'check_completeness.next_step not repointed'; END IF;
    IF cfg #>> '{workflow,steps,check_fabrication,action}' <> 'check_tool_fabrication' THEN
        RAISE EXCEPTION 'check_fabrication step missing'; END IF;
    IF cfg #>> '{workflow,steps,route_fabrication,config,then_step}' <> 'request_fabrication_review' THEN
        RAISE EXCEPTION 'route_fabrication router missing'; END IF;
    IF cfg #>> '{workflow,steps,route_fabrication,config,else_step}' <> 'save_training_data' THEN
        RAISE EXCEPTION 'route_fabrication else-branch wrong'; END IF;
    IF cfg #>> '{workflow,steps,request_fabrication_review,action}' <> 'checkpoint_for_review' THEN
        RAISE EXCEPTION 'request_fabrication_review step missing'; END IF;
    IF cfg #>> '{workflow,steps,request_fabrication_review,next_step}' <> 'complete' THEN
        RAISE EXCEPTION 'review branch must go to complete (no deploy)'; END IF;
    RAISE NOTICE 'bug 020 fabrication gate WIRED: check_completeness -> check_fabrication -> route_fabrication (fabricated -> review -> complete, else -> save_training_data)';
END $$;

COMMIT;
