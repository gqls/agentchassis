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
-- ── BLAST RADIUS — CORRECTED 2026-08-21 AFTER A COUNCIL REVISE ────────────
--
-- > **⚠ THE FIRST VERSION OF THIS SECTION WAS WRONG, and two seats caught it
-- > (`guardian` and `prior_art_librarian`, round 1 of corr 9e8d73b8).** It said:
-- > *"runs in ALL HISTORY: page-rebuild 0, pageflow-builder 0,
-- > site-work-orchestrator 0"*, and argued from that the change was
-- > behaviourally inert. The figure came from `orchestration_states`, which
-- > RETAINS TERMINAL ROWS FOR ROUGHLY TWO DAYS — measured 2026-08-21, only 24 of
-- > its 3,154 rows are older than 48h. So "0 runs in all history" was really
-- > "0 SURVIVING rows", which is not the same claim at all. Exactly the class of
-- > error the same author had just been corrected on in the previous round.
--
-- The durable source is `agent_run_stats` (per-agent lifetime counters).
-- `[MEASURED 2026-08-21]`, with the control first, because a silent table proves
-- nothing: it tracks **134 agent types, 85,389 runs, since 2026-07-26**.
--
--	agent_type              run_count   first_ran_at   last_ran_at
--	page-rebuild                    7   2026-07-26     2026-08-08
--	pageflow-builder                3   2026-08-09     2026-08-09
--	site-work-orchestrator          1   2026-08-09     2026-08-09
--	maintenance-triage              4   2026-08-05     2026-08-08   (the dispatcher)
--
-- So they are **RARE, NOT DEAD**: 11 runs between them in ~26 days, last activity
-- 12 days ago. (`agent_run_stats` itself only begins 2026-07-26, so this is
-- "since 26 July", not "for ever" — stating that rather than repeating the same
-- over-claim one table along.) `maintenance-triage`'s last run matches
-- `page-rebuild`'s to the same minute (2026-08-08 15:21), which corroborates the
-- single dispatch reference found below.
--
-- WHY THIS STILL DOES NOT MEAN A HAZARD IS ALREADY LOOSE, and this is the
-- reassuring part, checkable in one query: the three last ran **2026-08-09
-- 13:50**, and the first `content_hash` value was ever written **2026-08-20
-- 17:36** — eleven days LATER. They therefore cannot have stranded a stale
-- fingerprint, because the column had no values to strand when they last ran.
-- That is why the fleet-wide sweep found 228 of 228 pages matching: structural,
-- not lucky. Arming them is protective for their NEXT run.
--
-- REACHABILITY (a run-count cannot answer "could it fire tomorrow"): exactly ONE
-- live dispatch reference exists fleet-wide — `maintenance-triage` carries
-- `agent_type = page-rebuild`. A substring search for the names over
-- `default_config::text` also hits `council-gate`, `fix-proposer` and
-- `domain-research-classifier`, but there they are PROSE inside reviewer prompts,
-- not wiring; matching only VALUES at dispatch keys removes them. `guardian` also
-- objected that reachability was proven only for one of the three — true of the
-- dispatch scan, and now moot: `agent_run_stats` shows all three have RUN, which
-- is stronger evidence of reachability than any config scan.
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

-- ⚠⚠ GATE: EXACTLY ONE ACTIVE DEFINITION ROW PER TARGET TYPE.
-- Raised by `editquality` and `debug_historian` (both, independently). Four agent
-- types on this estate carry TWO active definition rows and only the higher
-- version loads, so an `UPDATE ... WHERE type = <x>` can land on a dormant row
-- while the live one stays unarmed — and this migration's own verify, which
-- counts armed steps, would PASS against the wrong row. `[MEASURED 2026-08-21]`
-- our three carry one row each, and 4 other types fleet-wide do not, so the trap
-- is real and simply does not bite here today. Asserted rather than assumed.
DO $$
DECLARE rows_per_type int;
BEGIN
    SELECT count(*) INTO rows_per_type FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type IN ('page-rebuild','pageflow-builder','site-work-orchestrator');
    IF rows_per_type <> 3 THEN
        RAISE EXCEPTION
          '547 single-row gate: expected exactly 3 active definition rows across the 3 target types, found % — a type has gained a second active row and an UPDATE by type could arm the dormant one while the live one stays unarmed. Scope the UPDATE to the loaded (max-version) row first.',
          rows_per_type;
    END IF;
END $$;

-- ⚠⚠ GATE: THE LOOP MUST NOT CARRY `substeps`, OR WE WOULD ARM A DEAD KEY.
-- This is the gating objection (`editquality`, HIGH) and it is the sharpest one
-- in the round. `LoopAction` reads `config["substeps"]` FIRST and falls back to
-- `config["sub_workflow"]["steps"]` only when substeps is absent or empty
-- (platform/orchestration/actions/loop_actions.go:91-104). So if a loop carried
-- BOTH, the executing copy would be `substeps` — and `jsonb_set` into
-- `sub_workflow.steps` would silently arm a DEAD key while the real step stayed
-- unarmed, with the recursive verify below cheerfully finding the armed dead copy
-- and passing. That is this bug's own census-blindness reproduced one level
-- deeper, inside the migration written to fix it.
--
-- `[MEASURED 2026-08-21]` none of the three carries `substeps`; each uses
-- `sub_workflow.steps` exclusively (9, 10 and 8 steps). The gate makes the
-- migration fail loudly rather than depend on that silently.
DO $$
DECLARE with_substeps int;
BEGIN
    SELECT count(*) INTO with_substeps
      FROM agent_definitions ad
      CROSS JOIN LATERAL jsonb_path_query(ad.default_config,
            '$.**{0 to 25} ? (exists(@.sub_workflow.steps.update_page_status))') AS loopcfg
     WHERE ad.is_active AND NOT COALESCE(ad.is_snapshot,false) AND ad.deleted_at IS NULL
       AND ad.type IN ('page-rebuild','pageflow-builder','site-work-orchestrator')
       AND jsonb_typeof(loopcfg->'substeps') = 'object'
       AND (SELECT count(*) FROM jsonb_object_keys(loopcfg->'substeps')) > 0;
    IF with_substeps <> 0 THEN
        RAISE EXCEPTION
          '547 substeps gate: % target loop(s) carry a non-empty config.substeps, which TAKES PRECEDENCE over sub_workflow.steps at runtime (loop_actions.go:91). Arming sub_workflow.steps would create a dead key and leave the executing step unarmed. Re-target the migration at substeps.',
          with_substeps;
    END IF;
END $$;

-- The stamps live inside a loop step's sub_workflow, and the loop step is named
-- differently on site-work-orchestrator, so each path is written explicitly
-- rather than by a clever shared expression. Three plain statements are easier
-- to review than one that is right for reasons a reader has to reconstruct.
--
-- On the concurrent-edit race `debug_historian` raised: each UPDATE takes a row
-- lock and `jsonb_set` operates on the row's CURRENT value at UPDATE time, not on
-- the value the gates above read — so a concurrent session's other keys survive,
-- and only a concurrent write to THIS one key could be lost. The gates, the
-- UPDATEs and the verify all run inside one transaction, so the window is the
-- transaction, not the session. `snapshot_agent()` backs up but does not lock;
-- that is stated rather than implied.
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
