-- 728_..._ROLLBACK.sql — restore the four-reason result_message verbatim.
--
-- This puts back a message that is WRONG BY OMISSION (bugs_open/442 §4): it
-- names four of the action's seven refusal reasons and omits the three copy-gate
-- ones. Roll back only if the longer message breaks a reader of
-- `result_message` — nothing in the workflow branches on it, so that would be
-- something outside this agent.

BEGIN;

SELECT snapshot_agent('meta-description-backfiller',
                      '728_ROLLBACK: pre-restore');

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
      FROM agent_definitions
     WHERE type = 'meta-description-backfiller'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config->'workflow'->'steps'->'complete'->'config'->>'result_message'
           LIKE '%voice_gate_unreadable%';
    IF n <> 1 THEN
        RAISE EXCEPTION
            'ABORT: no live meta-description-backfiller carries 728''s message (found %) — '
            'there is nothing here to roll back.', n;
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,complete,config,result_message}',
           to_jsonb($msg$Meta description backfill finished. Read each save_result: "updated" true is a write, false carries a named reason (empty_candidate / candidate_looks_internal / candidate_too_long / already_has_description).$msg$::text),
           false),
       updated_at = now()
 WHERE type = 'meta-description-backfiller'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE msg text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'complete'->'config'->>'result_message'
      INTO msg
      FROM agent_definitions
     WHERE type = 'meta-description-backfiller'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF position('voice_gate_unreadable' in msg) > 0 THEN
        RAISE EXCEPTION 'ABORT: 728''s message is still in place after the restore';
    END IF;
    IF position('already_has_description' in msg) = 0 THEN
        RAISE EXCEPTION 'ABORT: the restored message is not the four-reason original';
    END IF;
    RAISE NOTICE '728 ROLLBACK: the four-reason message is back, omissions and all.';
END $$;

COMMIT;
