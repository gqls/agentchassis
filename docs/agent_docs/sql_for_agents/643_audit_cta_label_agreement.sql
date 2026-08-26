-- 643 — arm the write-time CTA label/destination audit (bugs_open/399).
--
-- ✅ HOLD DISCHARGED 2026-08-26 — RELEASED (_HOLD suffix off), APPLIED out-of-band.
-- The hold required the binary carrying this key to have rolled, verified at the
-- POD and not from a commit sha. Done on chassis v1.0.1345, pod
-- agent-chassis-5864bf97c5-68t5h, WITH BOTH CONTROLS IN THE SAME EXEC:
--     PRESENT audit_cta_label_agreement      <- the key this file sets
--     PRESENT CTA_LABEL_MISMATCH             <- the literal the code emits
--     ABSENT  zzNOTAREALSYMBOL9f2a           <- control
--     ABSENT  deadbeefdeadbeefdeadbeef       <- control
-- Both controls absent is what makes the two PRESENTs mean anything: a grep that
-- matches everything returns the same answer as a grep that works.
-- ⚠ The startup 'build provenance' line had ALREADY scrolled out of a 200-line
-- tail ~10 minutes after the pod started — on a busy chassis it is not a
-- fallback you can rely on. The binary probe has no shelf life; use it.
--
-- ORIGINAL HOLD TEXT, kept because the reason is the lesson:
-- ⚠⚠ HELD. DO NOT APPLY UNTIL THE BINARY THAT READS THIS KEY HAS ROLLED, AND
-- DO NOT DISCHARGE THE HOLD FROM A COMMIT SHA — PROBE THE RUNNING POD.
-- Raised by the council gate's debug_historian seat (corr e9bda035, severity
-- HIGH) and it is right: this file originally shipped un-held with only a
-- sentence of prose asking for the right order. A banner cannot hold a
-- migration — the runner takes every pending file in the directory — so the
-- ordering constraint has to be in the FILENAME (CLAUDE.md, migration runner
-- practice). That is the whole reason for the _HOLD suffix here.
--
-- WHY THE ORDER MATTERS, concretely: an armed config key naming behaviour the
-- binary does not have reads as applied and does nothing (the bugs_open/380
-- trap). The demand control that tells the two states apart is the same one
-- migration 476 used for its sibling key — ask the POD, not git:
--
--   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
--            -o jsonpath='{.items[0].metadata.name}')
--   kubectl -n ai-persona-system exec "$POD" -- \
--     grep -aq "audit_cta_label_agreement" /proc/1/exe && echo KEY-PRESENT
--   kubectl -n ai-persona-system exec "$POD" -- \
--     grep -aq "CTA_LABEL_MISMATCH" /proc/1/exe && echo LITERAL-PRESENT
--
-- BOTH must print, and run a CONTROL in the same breath — a string that must be
-- ABSENT (e.g. a sha you know is not deployed) — because a grep that matches
-- everything and a grep that matches nothing look identical behind `&&`.
-- Discharge the hold by renaming (drop the _HOLD suffix) once both print.
--
-- ⚠ STAGED, NOT FLEET-WIDE IN ONE SHOT (guardian seat, same round). This file
-- now arms only the TWO PRIMARY WRITERS — page-build-handler (build) and
-- page-rerender (repair). The remaining four steps are armed by the sibling
-- 645, deliberately a separate apply.
--
--   ⚠⚠ AND THAT MEANS THE RATE IS BIASED UNTIL 645 IS APPLIED. The whole
--   argument for arming every writer is that an instrument armed on half its
--   writers reports a RATE that reads fleet-wide and is silently wrong. Staging
--   does not dissolve that — it schedules it. So: DO NOT READ THE RATE between
--   643 and 645. The RUNBOOK says the same thing beside the query.
--
-- ⚠ TAKES A SNAPSHOT FIRST (debug_historian, same round). jsonb_set surgery on
-- a live agent_definitions row with only an in-transaction verify block leaves
-- no pre-image to restore from. snapshot_agent() is this estate's convention
-- for exactly this class and it was missing.

BEGIN;

-- Pre-image, per the convention this file was missing (167, 434 are the worked
-- examples). Cheap, and the only thing that makes the UPDATEs below reversible
-- by something other than hand-written inverse SQL.
SELECT snapshot_agent('page-build-handler', 'pre-update: 643 arm audit_cta_label_agreement (bugs_open/399)');
SELECT snapshot_agent('page-rerender',      'pre-update: 643 arm audit_cta_label_agreement (bugs_open/399)');

-- Fail loudly if the population has changed since the census above. A migration
-- that silently arms 5 of 7 is how a biased rate ships looking complete.
DO $$
DECLARE
    n integer;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions a,
         LATERAL jsonb_path_query(a.default_config,
                 'strict $.**.action ? (@ == "save_page_sections")') x
    WHERE a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL;

    IF n <> 6 THEN
        RAISE EXCEPTION
            'save_page_sections step census is %, expected 6 (censused 2026-08-26). '
            'A step was added or removed: re-run the census in this file''s header and '
            'extend the UPDATEs below before arming, or the recorded mismatch RATE is '
            'silently biased by the writers this migration did not reach.', n;
    END IF;
END $$;

-- Top-level save_sections: page-build-handler, page-rerender, tool-recreation-handler.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,save_sections,config,audit_cta_label_agreement}', 'true'::jsonb, true),
    updated_at = NOW()
WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND type IN ('page-build-handler', 'page-rerender')
  AND default_config->'workflow'->'steps'->'save_sections'->>'action' = 'save_page_sections';

-- VERIFY as a DO/RAISE, never a bare SELECT: ON_ERROR_STOP does not abort a
-- COMMIT on a non-empty result set, so a SELECT-shaped verify block cannot stop
-- a bad migration (the RFC_006 lesson).
DO $$
DECLARE
    armed integer;
BEGIN
    -- ⚠ COUNTS THE KEY ON A STEP WHOSE ACTION IS save_page_sections, never the
    -- key alone. editquality's objection in the same round: jsonb_set with
    -- create_missing writes a DEAD key at a path nothing reads if a step name
    -- ever differs, and a verify that counts occurrences of the key would pass
    -- on that dead key and report success. Asking for the sibling `action`
    -- makes the assertion fail instead.
    SELECT count(*) INTO armed
    FROM agent_definitions a,
         LATERAL jsonb_path_query(a.default_config,
                 'strict $.**.steps.save_sections ? (@.action == "save_page_sections" && @.config.audit_cta_label_agreement == true)') x
    WHERE a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
      AND a.type IN ('page-build-handler', 'page-rerender');

    IF armed <> 2 THEN
        RAISE EXCEPTION 'armed % of the 2 primary save_page_sections steps — aborting (the key may have landed on a path nothing reads)', armed;
    END IF;
    RAISE NOTICE 'audit_cta_label_agreement armed on the 2 primary writers; the other four are 645 and THE RATE IS BIASED UNTIL IT APPLIES';
END $$;

COMMIT;

-- POST-APPLY READING, and the demand control that makes a zero informative.
-- The pass is INERT until an image carrying cta_label_audit.go rolls: an older
-- binary reads an unknown config key and ignores it. So a zero here means
-- "binary not rolled yet", NOT "no mismatches" — do not record a pre-roll zero
-- as evidence about this migration.
--
--   SELECT count(*) FILTER (WHERE (context->>'contradicts')::int > 0) AS pages_with_contradictions,
--          sum((context->>'contradicts')::int)                        AS contradictions,
--          sum((context->>'ambiguous')::int)                          AS ambiguous,
--          count(DISTINCT agent_type)                                 AS producing_agents
--   FROM agent_error_log
--   WHERE error_code = 'CTA_LABEL_MISMATCH' AND occurred_at > now() - interval '24 hours';
--
-- producing_agents MUST reach at least 2 (page-build-handler AND page-rerender)
-- once both paths have run. One producer means the coverage claim above is
-- failing silently, which is the failure this file's census exists to prevent.
