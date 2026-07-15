-- 147_acceptance_agent_mobile_profile.sql — request desktop AND mobile in the
-- acceptance run. DB-only; snapshot-prefixed (agent_definitions).
--
-- WHAT: sets tool-acceptance-agent.request_run.config.profiles =
-- ["desktop","mobile"] so request_browser_run tells the browser-runner
-- adapter to run both profiles. Each no-profile criterion then runs on both;
-- profile-pinned criteria (e.g. mobile-only no_horizontal_overflow) run only
-- where declared.
--
-- SAFE WITH THE CURRENT (P0) ADAPTER: it runs desktop and reports mobile-only
-- checks as skipped regardless — so this changes nothing until the P1/P2
-- adapter image (mobile profile + no_horizontal_overflow + interaction
-- checks) deploys, at which point mobile and interaction criteria become
-- LIVE instead of skipped. Idempotent.

BEGIN;

SELECT snapshot_agent('tool-acceptance-agent', '147_acceptance_agent_mobile_profile.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,request_run,config,profiles}',
      '["desktop","mobile"]'::jsonb,
      true)
WHERE type = 'tool-acceptance-agent' AND deleted_at IS NULL;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 147: acceptance runs go desktop + mobile
Observed: request_browser_run defaulted to desktop only; mobile and interaction criteria were reported skipped.
Root cause: not-applicable (capability gated on the P1/P2 adapter).
Fix: tool-acceptance-agent.request_run.config.profiles = ["desktop","mobile"]. Safe now (P0 adapter still desktop-only + skips mobile); mobile no_horizontal_overflow and interaction checks activate when the P1/P2 browser-runner-adapter image deploys.
Verified: P1/P2 runner unit-tested + proven live (xp-curve select→#tableWrap tr on desktop AND mobile; mobile no-overflow).
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-acceptance-agent' AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,request_run,config,profiles}' = '["desktop","mobile"]'::jsonb;
    IF n <> 1 THEN RAISE EXCEPTION '147: profiles not set (found %)', n; END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #> '{workflow,steps,request_run,config,profiles}'
--   FROM agent_definitions WHERE type='tool-acceptance-agent' AND deleted_at IS NULL;
-- Rollback: restore the snapshot, or set profiles back to ["desktop"].
