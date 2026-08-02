-- 292 — arm capture_renders: photograph an acceptance run that PASSES
--
-- WHAT: sets tool-acceptance-agent.request_run.config.capture_renders = true,
-- the last step of TL-035. With it, request_browser_run tells the
-- browser-runner adapter to keep a full-page screenshot of each (url, profile)
-- that passed every check, and the chassis files them on the run's doc_note
-- under a "Rendered:" line. Without it the payload carries capture_renders:false,
-- which is the adapter default and today's exact behaviour.
--
-- WHY IT EXISTS. Every defect that has reached the owner on this lane —
-- padding collapsed to 0px, baseline drift, overlapping labels — happened on a
-- page where every check PASSED, so nothing was ever photographed and there was
-- nothing to look at afterwards. Screenshots are failure evidence and stay that
-- way; Renders is a separate list that is a LOOK, never a verdict.
--
-- THE ORDER IS LOAD-BEARING, AND THIS FILE IS THE SECOND HALF OF IT. DB config
-- is live immediately; the Go half is not. Run this only against a chassis whose
-- binary carries the caller half, or you get a step config that reads
-- switched-on while the running binary drops the field — a state that looks
-- exactly like success. Verified before this file was applied, both replicas:
--
--   $ kubectl exec -n ai-persona-system <agent-chassis-pod> -- sh -c \
--       'grep -acF -- "capture_renders" /app/agent-chassis'          -> 1
--     positive control  "request_browser_run"                        -> 8
--     negative control  ".response.data.screenshots" (a literal the  -> 0
--       caller-half refactor DELETED; 0 here is what proves the
--       binary is post-change rather than a stale cached layer)
--
-- Caller half: 9cc63c775 + 72463f51e, council 2c895dd1 APPROVED r1.
-- Adapter half: council ab21beac APPROVED r1, live since v1.0.1225.
-- Idempotent. DB-only; snapshot-prefixed (agent_definitions).

BEGIN;

SELECT snapshot_agent('tool-acceptance-agent',
    '292_acceptance_runs_photograph_a_page_that_passes.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,request_run,config,capture_renders}',
      'true'::jsonb,
      true)
WHERE type = 'tool-acceptance-agent' AND deleted_at IS NULL;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 292: acceptance runs now photograph a page that PASSES
Observed: every defect that reached the owner (0px padding, baseline drift, overlapping labels) was on a page where every criterion passed, so no screenshot was ever taken and there was no artefact to inspect after the fact.
Root cause: not-applicable (a missing capability, not a defect). The browser-runner adapter only captured on failure; TL-035 added an opt-in and nothing could set it, because request_browser_run built its payload from a fixed key map with no path to the flag.
Fix: tool-acceptance-agent.request_run.config.capture_renders = true, applied only after the chassis binary carrying the caller half was pod-verified on both replicas with a positive AND a negative control.
Verified: a passing run must leave a "Rendered:" line on its own acceptance-run doc_note. A render is a LOOK, never a verdict — Renders carries no failing_checks by construction and nothing may branch on its presence.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-acceptance-agent' AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,request_run,config,capture_renders}' = 'true'::jsonb;
    IF n <> 1 THEN RAISE EXCEPTION '292: capture_renders not set (found %)', n; END IF;
END $$;

-- The sibling keys 147 and 145 set must survive: jsonb_set is key-level, but a
-- verify block that only checks its OWN key cannot tell a surgical write from
-- one that flattened the step. Assert a neighbour too.
DO $$
DECLARE p jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,request_run,config,profiles}' INTO p
    FROM agent_definitions
    WHERE type = 'tool-acceptance-agent' AND deleted_at IS NULL;
    IF p IS DISTINCT FROM '["desktop","mobile"]'::jsonb THEN
        RAISE EXCEPTION '292: 147''s profiles key did not survive (found %)', p;
    END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #> '{workflow,steps,request_run,config}'
--   FROM agent_definitions WHERE type='tool-acceptance-agent' AND deleted_at IS NULL;
-- Verify at the ARTEFACT, which is the only check that proves the flag reached
-- the adapter and object storage accepted the upload:
--   SELECT created_at, left(body, 400) FROM doc_notes
--    WHERE subject_type='tool' AND categories ? 'acceptance-run'
--    ORDER BY created_at DESC LIMIT 3;
-- Rollback: restore the snapshot, or set capture_renders back to false (the
-- key may also simply be removed — absent and false are the same to the caller).
