-- ROLLBACK for 539 — remove the `commit_sha?` declaration from
-- build-dispatch-loop's nested `mark_complete` step, returning `commit_sha` to
-- the whole-tree search.
--
-- WHEN YOU WOULD RUN THIS: if `result.commit_sha` stops being recorded on items
-- whose handler DID deploy something. The observable is an ABSENT field on
-- completed items, never an error — `commit_sha` is Optional, so nothing fails
-- loudly. Compare against the pre-539 rate rather than against zero:
--   SELECT count(*) FILTER (WHERE result ? 'commit_sha')::float / count(*)
--     FROM site_work_items WHERE status='complete' AND jsonb_typeof(result)='object'
--      AND completed_at >= '<apply time>';
-- and remember ~49% of completions legitimately carry no sha, so the ratio to
-- watch is roughly 0.51, not 1.0.
--
-- NOTE it restores the SEARCH, not the conflict rows; the rows follow because the
-- search is what writes them. It also restores the pick-a-winner behaviour this
-- file exists to end, so prefer correcting the PATH over rolling back.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config}'
      INTO cfg
      FROM agent_definitions
     WHERE type = 'build-dispatch-loop' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '539 ROLLBACK: no live mark_complete step at the expected nested path';
    END IF;
    IF NOT (cfg ? 'commit_sha?') THEN
        RAISE EXCEPTION '539 ROLLBACK: no commit_sha? key — 539 is not applied, or already rolled back';
    END IF;

    UPDATE agent_definitions
       SET default_config = default_config
             #- '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,commit_sha?}',
           updated_at = NOW()
     WHERE type = 'build-dispatch-loop' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
END $$;

DO $$
BEGIN
    IF (SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config}'
          FROM agent_definitions
         WHERE type = 'build-dispatch-loop' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL) ? 'commit_sha?' THEN
        RAISE EXCEPTION '539 ROLLBACK VERIFY FAILED: commit_sha? still present';
    END IF;
    RAISE NOTICE '539 ROLLBACK OK: commit_sha? removed';
END $$;

COMMIT;
