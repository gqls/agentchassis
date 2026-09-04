-- 773_..._ROLLBACK.sql — strip exactly the block 773 appended.
--
-- 773 APPENDED to the row's own current value rather than restating 728's text,
-- so the reverse is a suffix strip, not a restore-from-literal. That is the
-- safer direction: it cannot resurrect a stale copy of 728's message if another
-- migration has legitimately edited the earlier part in the meantime.
--
-- The guard is that the message must END with the appended block. If it does
-- not, something else has edited this message since 773 applied and a blind
-- strip would silently truncate that instead — so this aborts and asks for eyes.

BEGIN;

SELECT snapshot_agent('meta-description-backfiller',
                      '773_ROLLBACK: pre-revert');

DO $$
DECLARE
    msg      text;
    appended text := $msg$

⚠ WHERE THIS RUN'S RESULTS ACTUALLY ARE — AND THE KEY THAT WILL MISLEAD YOU.

Each page's outcome is written to collected_data as save_result_0, save_result_1, ... — one per page, in write order. Read those.

There is ALSO a bare save_result. It holds ONLY THE LAST PAGE. So on a run covering more than one page, a refusal on an earlier page sits in save_result_0 while the bare key reads "updated": true, and a reader who takes the obvious key sees a clean run. [MEASURED 2026-09-04, bugs_open/442 §11d] the bare key equalled the last page on every run checked; the same convention holds on other loops (page-content-writer's copy_gate and generated_content) and, confusingly, NOT on every field in them — so do not infer it from one example.

Two cheap controls: the number of save_result_<N> keys should equal the number of descriptions the writer returned, and a short series means pages are MISSING, not that they passed.

Nothing is lost when this misleads you: a copy-gate refusal also files a meta_description_refused work item at meta-description-repair, and THAT is the durable record. This message is the convenient surface, not the authoritative one.$msg$;
    n int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'complete'->'config'->>'result_message'
      INTO msg
      FROM agent_definitions
     WHERE type = 'meta-description-backfiller'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF msg IS NULL THEN
        RAISE EXCEPTION 'ABORT: no live meta-description-backfiller result_message to revert';
    END IF;
    IF right(msg, length(appended)) <> appended THEN
        RAISE EXCEPTION 'ABORT: the live message does not END with the block 773 appended. '
                        'Something has edited it since. Stripping blindly would truncate '
                        'that edit — revert by hand after reading the row.';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(
               default_config,
               '{workflow,steps,complete,config,result_message}',
               to_jsonb(left(msg, length(msg) - length(appended))),
               false),
           updated_at = now()
     WHERE type = 'meta-description-backfiller'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- VERIFY the revert: the amendment is gone and 728's content is still there.
    SELECT count(*) INTO n
      FROM agent_definitions
     WHERE type = 'meta-description-backfiller'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config->'workflow'->'steps'->'complete'->'config'->>'result_message'
           NOT LIKE '%save_result_0%'
       AND default_config->'workflow'->'steps'->'complete'->'config'->>'result_message'
           LIKE '%metaDescriptionFailsCopyGates%';
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: after the revert the row is not the 728 message (matched %)', n;
    END IF;

    RAISE NOTICE '773 ROLLBACK: appended block removed; 728 message intact.';
END $$;

COMMIT;
