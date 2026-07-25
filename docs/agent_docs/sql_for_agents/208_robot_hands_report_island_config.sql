-- 208_robot_hands_report_island_config.sql
-- Gripper dossier pilot — point robot-hands.com at its island intake.
-- Workstream: docs/agent_docs/docs024_key_docs_latest/robot_hands_gripper_dossier/
-- Design of record: DESIGN_2026-07-24_gripper_dossier_pilot.md §3/§5.
--
-- PRE-IMAGE SAFE (names no action, no agent). But applying it EARLY is not
-- free: pull_report_requests selects sites by `deploy_config ? 'report_island'`,
-- so once this row exists the pull task will try that endpoint every tick.
-- Apply it when the island service is actually answering, or leave the
-- scheduled task disabled (seed 210 ships both tasks disabled).
--
-- THE KEY IS NOT IN THIS FILE, AND MUST NOT BE.
-- pull_key is the shared secret the cluster sends as X-Internal-Key. A secret
-- committed to git is a secret you have to rotate, so this seed takes it as a
-- psql variable and REFUSES to run without one:
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v pull_key="<the island key>" \
--     < docs/agent_docs/sql_for_agents/208_robot_hands_report_island_config.sql
--
-- Generate the key on the island and paste it both places (island env +
-- here); it never needs to exist anywhere else.
--
-- MUST be applied via stdin or -f, NOT via `psql -c`: psql performs no
-- variable interpolation for -c, so :'pull_key' is a bare syntax error there
-- (verified against the live DB, 2026-07-25). Forgetting -v entirely is also
-- a syntax error, which with ON_ERROR_STOP aborts before the transaction
-- opens — the safe failure, and the reason the guard is shaped this way.
--
-- The key is bridged into plpgsql via set_config/current_setting because psql
-- does NOT substitute variables inside dollar-quoted strings (also verified:
-- `DO $$ ... :'pull_key' ... $$` is a syntax error, not a substitution). The
-- set_config output is sent to /dev/null so the secret is not echoed into the
-- terminal or the container logs.
--
-- MERGE, NEVER REPLACE. deploy_config is `{}` on this site today (checked
-- 2026-07-25) but another session may write a different key at any time, so
-- this uses `||` to add exactly one top-level key and preserves the rest.
-- Replacing the object wholesale is how a concurrent thread's config gets
-- silently dropped.

\set ON_ERROR_STOP on

-- Bridge the psql variable into something plpgsql can read (see header), with
-- the echo suppressed so the secret never reaches stdout or the pod logs.
SELECT set_config('gripper.pull_key', :'pull_key', false) \g /dev/null

BEGIN;

-- Refuse a placeholder or a weak key before touching the row.
DO $$
DECLARE
    k TEXT := current_setting('gripper.pull_key', true);
BEGIN
    IF k IS NULL OR k IN ('', 'SET_ME', 'changeme', 'REPLACE_ME') THEN
        RAISE EXCEPTION
            'pull_key not supplied — rerun with: psql -v pull_key="<the island key>" < 208_...sql';
    END IF;
    IF length(k) < 24 THEN
        RAISE EXCEPTION
            'pull_key is only % chars — this is a shared secret on a public endpoint, use >= 24 random chars', length(k);
    END IF;
END $$;

UPDATE sites
SET deploy_config = COALESCE(deploy_config, '{}'::jsonb) || jsonb_build_object(
        'report_island', jsonb_build_object(
            'base_url', 'https://tools.apis.uk/api/gripper/v1',
            'pull_key', current_setting('gripper.pull_key'),
            'note',     'Island intake for the gripper dossier. The cluster PULLS only; the island is never called by the cluster for anything else, and its payload deliberately carries no visitor email — PII stays on the island. Seed 208.'
        )
    ),
    updated_at = NOW()
WHERE domain = 'robot-hands.com';

-- Exactly one site, and the key readable back out.
DO $$
DECLARE
    n INTEGER;
    b TEXT;
BEGIN
    SELECT count(*) INTO n
    FROM sites
    WHERE domain = 'robot-hands.com' AND deploy_config ? 'report_island';

    IF n <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 configured site, found %', n;
    END IF;

    SELECT deploy_config->'report_island'->>'base_url' INTO b
    FROM sites WHERE domain = 'robot-hands.com';

    IF b IS NULL OR b = '' THEN
        RAISE EXCEPTION 'base_url did not persist';
    END IF;

    RAISE NOTICE 'report_island configured: %', b;
END $$;

COMMIT;

-- To undo (the pull then finds no sites and does nothing):
--   UPDATE sites SET deploy_config = deploy_config - 'report_island'
--   WHERE domain = 'robot-hands.com';
