-- 258_diagnose_loop_stamps_run_correlation.sql — bugs_open/124, part P3.
--
-- WHAT: diagnose-dispatch-loop's `claim_item` step now stamps the CLAIMING
--       RUN's own correlation onto the item it claims, as
--       spec.dispatch_correlation_id, inside the same atomic UPDATE that does
--       the claim.
--
-- WHY: diagnosis_artifacts, orchestration_states and the terminal doc_notes row
--      are all keyed on the ENVELOPE correlation
--      (params.ExecutionContext.CorrelationID — diagnose_assemble_bundle_action.go),
--      NOT on the correlation_id the loop passes down through input_mapping. So
--      for a loop-dispatched item, `spec.correlation_id` names NOTHING: it is a
--      uuid minted by whoever created the intake and never used by any run.
--      Both diagnosis-triage-created items are in exactly that state today, and
--      it went unnoticed because the field still LOOKS like a key.
--
--      The 090 trigger's stated design is that one key ties the intake item, the
--      diagnosis_artifacts bundles and the terminal doc_notes row together. This
--      is the row that makes that true for the automatic path as well as the
--      manual one. Without it, removing the 090 script's duplicate direct
--      publish (P1/P2 of the same fix) would leave the printed correlation
--      resolving to nothing for everybody.
--
-- ═══════════════════════════════════════════════════════════════════════════
-- ORDERING IS LOAD-BEARING. DO NOT APPLY THIS BEFORE THE IMAGE.
-- ═══════════════════════════════════════════════════════════════════════════
-- `params: ['$ctx.correlation_id']` needs the $ctx. execution-context parameter
-- namespace, added in platform/orchestration/actions/execution_context_params.go
-- and wired into QueryDatabaseAction. On a chassis that predates it, that path
-- falls through to ExtractNestedField, resolves to nil, and QueryDatabaseAction
-- returns "query param path '$ctx.correlation_id' resolved to nil" — which fails
-- claim_item and stops the diagnose lane dispatching AT ALL.
--
-- The guard at the bottom of this file cannot check a running binary from SQL.
-- So check it yourself, against the POD, before applying:
--
--   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
--     'strings /app/agent-chassis | grep -c "unknown execution-context field"'
--
-- Must print 1. That literal is a string this change INTRODUCES, so 0 is a real
-- negative rather than an unfalsifiable one.
--
-- Applied 2026-07-28 against chassis v1.0.1191
-- (digest sha256:2f96b795a5c4636d41bdc384318f3f2d264188b9bf4017fb2d74ff2746a760cc),
-- pod-verified before this file was run.
--
-- ROLLBACK: restore the pre-update snapshot this file takes —
--   SELECT * FROM agent_definitions
--    WHERE type='diagnose-dispatch-loop' AND COALESCE(is_snapshot,false)=true
--    ORDER BY created_at DESC LIMIT 1;
-- then copy its default_config back onto the live row. Reverting is also safe on
-- an OLD binary: the pre-update config binds no params at all.

BEGIN;

SELECT snapshot_agent('diagnose-dispatch-loop',
       '258_diagnose_loop_stamps_run_correlation: pre-update');

-- Replace claim_item's config wholesale rather than patching two keys: the query
-- and the params array are one contract ($1 must have exactly one binding), and
-- writing them separately leaves a window where the file is half-applied.
--
-- The RETURNING list is UNCHANGED and must stay so — call_handler's
-- input_mapping reads claimed.symptom / .owner / .repo / .ref / .target_site_id
-- / .runtime_site / .subject_type / .subject_key / .correlation_id /
-- .work_item_id off it, and spawn_handler reads claimed.handler_agent.
-- FOR UPDATE SKIP LOCKED is likewise unchanged: it is what makes concurrent
-- ticks safe, and it is the reason this is a claim and not a read.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,claim_item,config}',
      jsonb_build_object(
        'output_format', 'object',
        -- bugs_open/124: bind the RUN's own correlation. See the ordering note.
        'params', jsonb_build_array('$ctx.correlation_id'),
        'query', $q$UPDATE site_work_items SET status = 'diagnosing', claimed_by = 'diagnose-dispatch-loop', claimed_at = NOW(), spec = jsonb_set(COALESCE(spec, '{}'::jsonb), '{dispatch_correlation_id}', to_jsonb($1::text), true) WHERE id = (SELECT id FROM site_work_items WHERE pipeline = 'diagnose' AND item_type = 'needs_diagnosis' AND status = 'awaiting_diagnosis' AND attempt_count < max_attempts ORDER BY priority ASC, created_at ASC FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id::text AS work_item_id, handler_agent, spec->>'symptom' AS symptom, spec->>'owner' AS owner, spec->>'repo' AS repo, spec->>'ref' AS ref, spec->>'site_id' AS target_site_id, spec->>'runtime_site' AS runtime_site, spec->>'subject_type' AS subject_type, spec->>'subject_key' AS subject_key, spec->>'correlation_id' AS correlation_id$q$
      )
    ),
    updated_at = now()
WHERE type = 'diagnose-dispatch-loop'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

-- ── guard: assert the exact post-conditions, inside the transaction ──────────
DO $guard$
DECLARE
  cfg   jsonb;
  q     text;
  steps jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'claim_item'->'config',
         default_config->'workflow'->'steps'
    INTO cfg, steps
    FROM agent_definitions
   WHERE type = 'diagnose-dispatch-loop'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cfg IS NULL THEN
    RAISE EXCEPTION '258: no live diagnose-dispatch-loop row, or claim_item.config is missing';
  END IF;

  IF cfg->'params' <> jsonb_build_array('$ctx.correlation_id') THEN
    RAISE EXCEPTION '258: params is % — expected exactly ["$ctx.correlation_id"]', cfg->'params';
  END IF;

  q := cfg->>'query';

  -- the new behaviour
  IF position('dispatch_correlation_id' in q) = 0 THEN
    RAISE EXCEPTION '258: query does not stamp dispatch_correlation_id';
  END IF;
  IF position('to_jsonb($1::text)' in q) = 0 THEN
    RAISE EXCEPTION '258: query does not bind $1 — the params array would be unused';
  END IF;

  -- the behaviour that MUST NOT have changed. A stamp is not worth losing the
  -- claim's exclusivity or the envelope the handler is called with.
  IF position('FOR UPDATE SKIP LOCKED' in q) = 0 THEN
    RAISE EXCEPTION '258: query lost FOR UPDATE SKIP LOCKED — concurrent ticks would double-claim';
  END IF;
  IF position('status = ''awaiting_diagnosis''' in q) = 0 THEN
    RAISE EXCEPTION '258: query no longer selects on awaiting_diagnosis';
  END IF;
  IF position('AS work_item_id' in q) = 0
     OR position('AS symptom' in q) = 0
     OR position('AS target_site_id' in q) = 0
     OR position('AS runtime_site' in q) = 0
     OR position('AS subject_type' in q) = 0
     OR position('AS subject_key' in q) = 0
     OR position('AS correlation_id' in q) = 0
     OR position('handler_agent' in q) = 0 THEN
    RAISE EXCEPTION '258: the RETURNING list changed — call_handler/spawn_handler input_mapping would break';
  END IF;

  -- the rest of the workflow is untouched. mark_complete in particular: the
  -- filed bug claimed nothing closes these items, and the fix must not make
  -- that claim true by accident.
  IF NOT (steps ? 'mark_complete' AND steps ? 'mark_failed' AND steps ? 'reap_stuck'
          AND steps ? 'spawn_handler' AND steps ? 'call_handler' AND steps ? 'check_claimed') THEN
    RAISE EXCEPTION '258: a sibling step went missing — steps are now %', (SELECT jsonb_agg(k) FROM jsonb_object_keys(steps) k);
  END IF;

  RAISE NOTICE '258: claim_item now stamps spec.dispatch_correlation_id from $ctx.correlation_id';
END
$guard$;

COMMIT;
