-- 547_arm_the_three_unarmed_deploy_stampers.sql
--
-- bugs_open/315 / PLAN_2026-08-19 decision **D7**. Arms the three remaining
-- `update_page_status` steps that stamp a page `deployed` without recording what
-- was sent, so `pages.content_hash` can never go stale behind a deploy.
--
-- NOT a _HOLD file: the config key `deploy_result_field` is ALREADY declared by
-- the running binary (migration 494 armed three other agents on 2026-08-20 and
-- 232 fingerprints have been written since), so there is no image-before-config
-- ordering constraint here. It is safe to run against the current fleet.
--
-- ── WHY THIS EXISTS ────────────────────────────────────────────────────────
--
-- `UpdatePageStatusAction` writes `pages.content_hash` ONLY when the
-- deploy-evidence guard ran. An UNARMED step that deploys NEW BYTES therefore
-- leaves the PREVIOUS deploy's fingerprint in place, and the divergence check
-- (DGH-015, `page_content_divergence`) then reports a perfectly healthy page as
-- diverged — permanently, because nothing rewrites that row.
--
-- Migration 526 (which enables that check) REFUSES to apply while any unarmed
-- stamper exists. This migration is what makes 526 appliable.
--
-- ── THE MEASUREMENT THAT MADE THIS NECESSARY, AND HOW IT WAS MISSED ────────
--
-- The 315 lane originally recorded "exactly three live steps stamp deployed and
-- all three are armed". That was FALSE. The census walked
-- `default_config.<workflow>.steps.*` — ONE LEVEL — and these three live one
-- level deeper, at `workflow.steps.<loop>.config.sub_workflow.steps.
-- update_page_status`. The council gate's `guardian` seat (corr be85a6d3)
-- predicted exactly that blindness from the SHAPE of the claim, without seeing
-- the query. `[MEASURED 2026-08-21, recursive]` six stampers, three unarmed:
--
--	armed    page-rerender          workflow.steps.update_status            deploy_result
--	armed    report-builder         workflow.steps.update_status            deploy_result
--	armed    section-editor         workflow.steps.update_page_status       git_result
--	UNARMED  page-rebuild           …build_pages_loop.config.sub_workflow…  <- this file
--	UNARMED  pageflow-builder       …build_pages_loop.config.sub_workflow…  <- this file
--	UNARMED  site-work-orchestrator …build_items_loop.config.sub_workflow…  <- this file
--
-- Use the RECURSIVE walk to re-derive it — the one-level version is what got it
-- wrong (LANDMINES, "Censusing the steps that stamp a page deployed"):
--   SELECT ad.type, step->'config'->>'deploy_result_field'
--     FROM agent_definitions ad
--     CROSS JOIN LATERAL jsonb_path_query(ad.default_config,
--           '$.**{0 to 25} ? (@.action == "update_page_status" && @.config.status == "deployed")') AS step
--    WHERE ad.is_active AND NOT COALESCE(ad.is_snapshot,false) AND ad.deleted_at IS NULL;
--
-- ── BLAST RADIUS: MEASURED, AND IT IS ZERO TODAY ───────────────────────────
--
-- Arming changes behaviour in exactly one case — the stamp is REFUSED when the
-- deploy step reported a skip, instead of proceeding. So the exposure is "how
-- often do these three run at all":
--
--   `[MEASURED 2026-08-21]` runs in ALL HISTORY: page-rebuild 0,
--   pageflow-builder 0, site-work-orchestrator 0. Zero scheduled_tasks target
--   them; zero site_work_items are routed at them.
--
--   REACHABILITY, which is the part a run-count cannot answer: exactly ONE live
--   dispatch reference exists fleet-wide — `maintenance-triage` carries
--   `agent_type = page-rebuild`. (A text search for the names also hits
--   `council-gate`, `fix-proposer` and `domain-research-classifier`, but those
--   are PROSE inside reviewer prompts, not wiring — checked by matching only
--   VALUES at dispatch keys rather than substrings anywhere in the config.)
--   `pageflow-builder` and `site-work-orchestrator` have no dispatch reference
--   at all.
--
-- So this is behaviourally inert today and protective the moment
-- `maintenance-triage` routes anything at `page-rebuild`. It is deliberately
-- ARM rather than DELETE: these are live definitions this lane does not own,
-- and arming is the reversible half of that choice.
--
-- ── WHY ARM RATHER THAN TAKE D6 (NULL the hash on an unarmed stamp) ────────
--
-- Arming RAISES fingerprint coverage — three more paths start recording what
-- they sent. D6's stamp-side NULLing LOWERS it: a page rebuilt by one of these
-- would lose its fingerprint until an armed path redeployed it. D6 remains
-- worth doing as the backstop for the NEXT unarmed stamper — the one nobody
-- notices being added — but it is not the fix for a stamper we can simply arm.
--
-- ── AFTER APPLYING, THE FIRST QUERY IS "WHAT DID I BREAK?" ─────────────────
--
-- This is bugs_open/336's lesson, learned by this lane by taking every
-- page-publish in the estate down for 33 minutes while confirming its config
-- was right:
--   SELECT count(*) FROM orchestration_states WHERE error ILIKE '%deploy_result_field%';  -- must stay 0
--   SELECT count(*) FROM agent_error_log WHERE error_code='DEPLOY_EVIDENCE_UNREADABLE'
--     AND created_at > now() - interval '1 hour';
--   SELECT status, count(*) FROM site_work_items WHERE item_type='page_rerender' GROUP BY 1;
-- Rollback: 547_arm_the_three_unarmed_deploy_stampers_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('page-rebuild',           '547_arm_the_three_unarmed_deploy_stampers: pre-update');
SELECT snapshot_agent('pageflow-builder',       '547_arm_the_three_unarmed_deploy_stampers: pre-update');
SELECT snapshot_agent('site-work-orchestrator', '547_arm_the_three_unarmed_deploy_stampers: pre-update');

-- Already-applied gate (the runner reads a RAISE containing 'already').
DO $$
DECLARE done int;
BEGIN
    SELECT count(*) INTO done
      FROM agent_definitions ad
      CROSS JOIN LATERAL jsonb_path_query(ad.default_config,
            '$.**{0 to 25} ? (@.action == "update_page_status" && @.config.status == "deployed")') AS step
     WHERE ad.is_active AND NOT COALESCE(ad.is_snapshot,false) AND ad.deleted_at IS NULL
       AND ad.type IN ('page-rebuild','pageflow-builder','site-work-orchestrator')
       AND (step->'config'->>'deploy_result_field') IS NOT NULL;
    IF done = 3 THEN
        RAISE EXCEPTION '547: already applied — all three steps already name a deploy_result_field';
    END IF;
END $$;

-- COUNTED NEEDLE-GATE. Each of the three must still be the shape this migration
-- believes it is: the stamp sits in a loop's sub_workflow ALONGSIDE a git_commit
-- step named `deploy_page` whose output_field is `page_deployed` — which is the
-- field we are about to name. If the shape has moved, abort rather than wire a
-- guard to a field nothing writes.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions ad
     WHERE ad.is_active AND NOT COALESCE(ad.is_snapshot,false) AND ad.deleted_at IS NULL
       AND ad.type IN ('page-rebuild','pageflow-builder','site-work-orchestrator')
       AND EXISTS (
         SELECT 1
           FROM jsonb_path_query(ad.default_config, '$.**{0 to 25} ? (exists(@.sub_workflow.steps.deploy_page))') sw
          WHERE sw->'sub_workflow'->'steps'->'deploy_page'->>'action' = 'git_commit'
            AND sw->'sub_workflow'->'steps'->'deploy_page'->>'output_field' = 'page_deployed'
            AND sw->'sub_workflow'->'steps'->'update_page_status'->>'action' = 'update_page_status'
            AND sw->'sub_workflow'->'steps'->'update_page_status'->'config'->>'status' = 'deployed'
       );
    IF n <> 3 THEN
        RAISE EXCEPTION
          '547 needle-gate: expected 3 agents whose loop sub_workflow pairs a git_commit step deploy_page (output_field page_deployed) with a deployed-stamping update_page_status, found % — re-derive against the live workflow', n;
    END IF;
END $$;

-- The stamps live inside a loop step's sub_workflow, and the loop step is named
-- differently on site-work-orchestrator, so each path is written explicitly
-- rather than by a clever shared expression. Three plain statements are easier
-- to review than one that is right for reasons a reader has to reconstruct.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,build_pages_loop,config,sub_workflow,steps,update_page_status,config,deploy_result_field}',
         '"page_deployed"'),
       updated_at = NOW()
 WHERE type IN ('page-rebuild','pageflow-builder')
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,build_items_loop,config,sub_workflow,steps,update_page_status,config,deploy_result_field}',
         '"page_deployed"'),
       updated_at = NOW()
 WHERE type = 'site-work-orchestrator'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

-- VERIFY INSIDE THE TRANSACTION, AS A RAISE. A verify block of SELECTs cannot
-- stop the COMMIT — ON_ERROR_STOP does not fire on a non-empty result — so only
-- an exception actually rolls this back.
DO $$
DECLARE armed int; unarmed_fleetwide int;
BEGIN
    SELECT count(*) INTO armed
      FROM agent_definitions ad
      CROSS JOIN LATERAL jsonb_path_query(ad.default_config,
            '$.**{0 to 25} ? (@.action == "update_page_status" && @.config.status == "deployed")') AS step
     WHERE ad.is_active AND NOT COALESCE(ad.is_snapshot,false) AND ad.deleted_at IS NULL
       AND ad.type IN ('page-rebuild','pageflow-builder','site-work-orchestrator')
       AND step->'config'->>'deploy_result_field' = 'page_deployed';
    IF armed <> 3 THEN
        RAISE EXCEPTION '547 verify: expected all 3 target steps armed with page_deployed, found %', armed;
    END IF;

    -- THE WHOLE POINT: after this, the fleet must have NO unarmed deployed-stamper
    -- left, because that is exactly what migration 526 gates on. Asserting the
    -- fleet-wide zero (not just our three) catches a fourth appearing meanwhile.
    SELECT count(*) INTO unarmed_fleetwide
      FROM agent_definitions ad
      CROSS JOIN LATERAL jsonb_path_query(ad.default_config,
            '$.**{0 to 25} ? (@.action == "update_page_status" && @.config.status == "deployed")') AS step
     WHERE ad.is_active AND NOT COALESCE(ad.is_snapshot,false) AND ad.deleted_at IS NULL
       AND (step->'config'->>'deploy_result_field') IS NULL;
    IF unarmed_fleetwide <> 0 THEN
        RAISE EXCEPTION
          '547 verify: % unarmed deployed-stamper(s) REMAIN fleet-wide after arming our three — a fourth has appeared; 526 will still refuse. Re-run the recursive census.',
          unarmed_fleetwide;
    END IF;
END $$;

COMMIT;
