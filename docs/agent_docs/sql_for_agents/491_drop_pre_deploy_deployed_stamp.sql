-- 491_drop_pre_deploy_deployed_stamp.sql
--
-- bugs_open/315. Owner authorised 2026-08-19.
--
-- WHAT IS WRONG. `page-build-handler` and `tool-recreation-handler` both stamp
-- pages.build_status='deployed' (and therefore pages.deployed_at) BEFORE their
-- deploy is dispatched. Measured at the live config by joining every step on
-- next_step (jsonb_each key order is arbitrary and reverses this particular
-- sequence, so it must not be read by eye):
--
--     save_sections(save_page_sections) -> update_status(update_page_status,
--     status=deployed) -> spawn_rerender_agent -> deploy_page(call_agent)
--
-- So at the moment of the stamp NOTHING has been sent to the git-adapter. No
-- guard can help these two: there is no deploy evidence in scope to read. This
-- is the 2-of-5 half of 315 and it is the only half fixable in config alone.
--
-- WHAT THIS DOES. Deletes the premature step and points save_sections straight
-- at the spawn. Both agents already call `page-rerender`, whose own
-- update_status runs AFTER its git_commit and is keyed
-- page_id_field=rendered_page.page_id — a stronger identity than these two,
-- which resolve the page by NAME (site_id_field + page_name_field).
--
-- WHAT THIS DOES *NOT* DO, stated because the council's editquality seat was
-- right that burying it was the defect (round 1, corr 377167cd): this does NOT
-- make deployed_at evidence of publication. page-rerender, report-builder and
-- section-editor still stamp after a git_commit whose result they discard —
-- 'deploy_result' appears nowhere in v3_site_actions.go. Gating those three
-- needs the adapter to return the commit sha it computes and throws away, which
-- is architecture-scope and is now architecture_review/RFC_038. The claim here
-- is only that 2 of 5 stampers stop asserting a deploy that has not been asked
-- for, and that the surviving stamp is on a better identifier.
--
-- SAFETY. Removing the step breaks no declared output contract: neither agent's
-- `complete` step names status_updated in config.output_fields (page-build-handler
-- ["sections_saved","deploy_result"]; tool-recreation-handler ["tool_analysis",
-- "sections_saved","deploy_result","training_data_saved"]). Verified at the live
-- config 2026-08-19, closing the prior_art_librarian seat's round-1 objection.
--
-- FAILURE DIRECTION. If page-rerender's stamp does not fire for some first-build
-- shape, the page stays UN-stamped rather than stamped-early-but-wrong.
-- Un-stamped is the recoverable direction: the reconciler re-emits, bounded by
-- the existing park-after-3. A false 'deployed' is the bugs_closed/040 shape and
-- makes a page permanently fileless.
--
-- ROLLBACK: 491_drop_pre_deploy_deployed_stamp_ROLLBACK.sql — surgical (re-adds
-- the two steps), deliberately NOT a whole-config restore, because 488 is
-- pending against page-build-handler from another lane and a blob restore would
-- silently revert it.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-build-handler',      '491_drop_pre_deploy_deployed_stamp: pre-update');
SELECT snapshot_agent('tool-recreation-handler', '491_drop_pre_deploy_deployed_stamp: pre-update');

-- ALREADY-APPLIED probe arm. The runner executes pending files inside a doomed
-- transaction and reads a RAISE containing 'already' as LIKELY ALREADY APPLIED.
DO $$
DECLARE still int;
BEGIN
    SELECT count(*) INTO still
      FROM agent_definitions
     WHERE type IN ('page-build-handler','tool-recreation-handler')
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND default_config->'workflow'->'steps' ? 'update_status';
    IF still = 0 THEN
        RAISE EXCEPTION '491: already applied — neither agent carries an update_status step';
    END IF;
END $$;

-- COUNTED NEEDLE-GATE. Abort unless the shape is EXACTLY what was measured.
-- Concurrent sessions edit agent_definitions constantly; this is what makes a
-- stale premise abort instead of half-applying.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
      FROM agent_definitions
     WHERE type IN ('page-build-handler','tool-recreation-handler')
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND default_config->'workflow'->'steps'->'update_status'->'config'->>'status' = 'deployed'
       AND default_config->'workflow'->'steps'->'save_sections'->>'next_step'      = 'update_status'
       AND default_config->'workflow'->'steps'->'update_status'->>'next_step'
           IN ('spawn_rerender_agent','spawn_rerender');
    IF n <> 2 THEN
        RAISE EXCEPTION
          '491 needle-gate: expected exactly 2 rows in the measured shape, found % — re-derive against the live workflow before applying', n;
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config =
         jsonb_set(default_config, '{workflow,steps,save_sections,next_step}', '"spawn_rerender_agent"')
         #- '{workflow,steps,update_status}',
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config =
         jsonb_set(default_config, '{workflow,steps,save_sections,next_step}', '"spawn_rerender"')
         #- '{workflow,steps,update_status}',
       updated_at = NOW()
 WHERE type = 'tool-recreation-handler'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

-- POST-VERIFY. DO/RAISE, never a bare SELECT: ON_ERROR_STOP ignores a non-empty
-- result set, so a verify block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE bad int;
BEGIN
    SELECT count(*) INTO bad
      FROM agent_definitions
     WHERE type IN ('page-build-handler','tool-recreation-handler')
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND ( default_config->'workflow'->'steps' ? 'update_status'
             OR default_config->'workflow'->'steps'->'save_sections'->>'next_step'
                NOT IN ('spawn_rerender_agent','spawn_rerender') );
    IF bad <> 0 THEN
        RAISE EXCEPTION '491 post-verify: % row(s) still carry the pre-deploy stamp or a bad save_sections next_step', bad;
    END IF;

    -- The spawn step the rewired next_step points at must actually exist.
    SELECT count(*) INTO bad
      FROM agent_definitions ad
     WHERE ad.type IN ('page-build-handler','tool-recreation-handler')
       AND ad.is_active AND NOT COALESCE(ad.is_snapshot,false) AND ad.deleted_at IS NULL
       AND NOT (ad.default_config->'workflow'->'steps'
                ? (ad.default_config->'workflow'->'steps'->'save_sections'->>'next_step'));
    IF bad <> 0 THEN
        RAISE EXCEPTION '491 post-verify: % row(s) point save_sections at a step that does not exist', bad;
    END IF;
END $$;

COMMIT;
