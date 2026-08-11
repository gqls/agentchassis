-- 383 — tool-acceptance: wire the vision critique to a reader (bugs_open/243
-- candidate 3, owner decision 2026-08-11: "wire up the vision findings")
--
-- APPLIED BY HAND 2026-08-11 ~12:45Z (roll condition met: v1.0.1286 pod-grep
-- record_vision_finding=6 / neg ctrl 0 on both replicas; guard passed; verified
-- at the live row). The _HOLD suffix was dropped POST-APPLY so the ledger could
-- record it — the runner refuses to record uppercase-suffixed sidecars. NOTE:
-- the number 383 collided with another session's 383_rfc022_… file the same
-- day; the ledger is filename-keyed, so both stand.
--
-- ⚠ original hold condition (now met): DO NOT APPLY until a chassis image carrying the
-- `record_vision_finding` action is rolled (commit of 2026-08-11, this lane).
-- A workflow step naming an unregistered action fails at runtime ("image
-- first, then seeds"). Verify before applying:
--   kubectl exec -n ai-persona-system <agent-chassis-pod> -- sh -c \
--     'strings /app/agent-chassis | grep -c "record_vision_finding"'   # ≥1
-- Then apply by hand (psql -f) and record:
--   ./scripts/migration/run-migrations.sh --record-only \
--     383_tool_acceptance_vision_findings_visible_HOLD.sql --note '<pod-grep result>'
--
-- WHY. The look step (TL-035 e) writes its critique to a render-critique
-- doc_note no code reads. Measured 2026-08-11 — the vision half's first-ever
-- run reported a real defect (contrast 1.06:1) and the run was stamped PASSED
-- in the same second, with nothing raised. Three changes, all config:
--
--   1. look.prompt_template gains a trailing machine line
--      (FINDINGS: none | FINDINGS: reported) so prose stays for people and
--      one line is for the parser.
--   2. record_look.next_step -> file_vision_finding (was complete).
--   3. NEW step file_vision_finding: action record_vision_finding — files ONE
--      deduped vision_finding work item (human-review / needs_human_review)
--      when the critique reports or is unparseable; explicit "none" files
--      nothing. error_step = complete: a filing failure must not change the
--      acceptance outcome (TL-035: best-effort, never a verdict).
--
-- ROLLBACK: restore the agent row from the snapshot this file takes
-- (snapshot_agent note '383_tool_acceptance_vision_findings_visible: pre-update').

BEGIN;

SELECT snapshot_agent('tool-acceptance-agent',
  '383_tool_acceptance_vision_findings_visible: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,look,config,prompt_template}',
             to_jsonb(
               (default_config#>>'{workflow,steps,look,config,prompt_template}')
               || E'\n\nAfter your prose, end with exactly one line: "FINDINGS: none" if you found nothing worth a human''s attention, or "FINDINGS: reported" if you described any defect above. This last line is read by a machine; everything above it is read by a person.'
             )
           ),
           '{workflow,steps,record_look,next_step}',
           '"file_vision_finding"'
         ),
         '{workflow,steps,file_vision_finding}',
         '{
           "action": "record_vision_finding",
           "config": {
             "critique_field": "vision_look.result",
             "function_field": "input_data.spec.function",
             "site_id_field": "site_record.site_id",
             "images_field": "browser_run",
             "error_step": "complete"
           },
           "next_step": "complete",
           "description": "243 c3: file a reporting critique as ONE deduped vision_finding item (human-review). Explicit FINDINGS: none files nothing; unparsed files visibly. A failure here must not change the acceptance outcome.",
           "output_field": "vision_finding"
         }'::jsonb
       ),
       updated_at = now()
 WHERE type = 'tool-acceptance-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  cfg jsonb;
  n   int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'tool-acceptance-agent'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '383 guard: expected exactly 1 live tool-acceptance-agent row, found %', n;
  END IF;

  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type = 'tool-acceptance-agent'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cfg#>>'{workflow,steps,look,config,prompt_template}' NOT LIKE '%FINDINGS: none%' THEN
    RAISE EXCEPTION '383 guard: look prompt did not gain the FINDINGS line';
  END IF;
  IF cfg#>>'{workflow,steps,record_look,next_step}' <> 'file_vision_finding' THEN
    RAISE EXCEPTION '383 guard: record_look.next_step is %, not file_vision_finding',
      cfg#>>'{workflow,steps,record_look,next_step}';
  END IF;
  IF cfg#>>'{workflow,steps,file_vision_finding,action}' <> 'record_vision_finding' THEN
    RAISE EXCEPTION '383 guard: file_vision_finding step missing or wrong action';
  END IF;
  IF cfg#>>'{workflow,steps,file_vision_finding,config,error_step}' <> 'complete' THEN
    RAISE EXCEPTION '383 guard: file_vision_finding.error_step must be complete (a filing failure must not change the acceptance outcome)';
  END IF;
END $$;

COMMIT;
