-- 694 ROLLBACK — restore content-quality-auditor's default_config as it stood
-- before 694 widened its sight.
--
-- 694 changed exactly two strings inside one agent's default_config:
--   workflow.steps.load_page_content.config.query   (the four-name allow-list,
--     the unordered 1000-char sample, no style stripping)
--   workflow.steps.run_content_llm_audit.config.prompt  (TOP 5 cap, the
--     category enum without 'audience', no promise-keeping dimensions)
-- It created no table, no column, no agent, no key and no row outside
-- agent_definitions_bak_694, so a restore of default_config is the whole undo.
-- Nothing is DELETEd here: the correct inverse of "replaced a string" is
-- "put the old string back", and the old strings are in the backup table 694
-- took inside its own transaction.
--
-- SAFE TO RUN ONLY while agent_definitions_bak_694 exists. It aborts rather
-- than guessing if the backup is missing, holds anything other than exactly one
-- row, or if the live row is already back to the pre-694 config (in which case
-- there is nothing to undo and a silent no-op would read as a successful
-- rollback).

BEGIN;

DO $$
DECLARE
    nbak int; nlive int; live_md5 text; bak_md5 text;
BEGIN
    IF to_regclass('agent_definitions_bak_694') IS NULL THEN
        RAISE EXCEPTION '694 ROLLBACK: agent_definitions_bak_694 does not exist — 694 was never applied, or the backup was dropped. Restore from the snapshot_agent row instead.';
    END IF;

    SELECT count(*) INTO nbak FROM agent_definitions_bak_694;
    IF nbak <> 1 THEN
        RAISE EXCEPTION '694 ROLLBACK: expected exactly 1 backed-up row, found %', nbak;
    END IF;

    SELECT count(*) INTO nlive FROM agent_definitions
    WHERE type = 'content-quality-auditor' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF nlive <> 1 THEN
        RAISE EXCEPTION '694 ROLLBACK: expected exactly 1 active content-quality-auditor row, found %', nlive;
    END IF;

    SELECT md5(default_config::text) INTO live_md5 FROM agent_definitions
    WHERE type = 'content-quality-auditor' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    SELECT md5(default_config::text) INTO bak_md5 FROM agent_definitions_bak_694;

    IF live_md5 = bak_md5 THEN
        RAISE EXCEPTION '694 ROLLBACK: the live config already equals the backup — 694 is not applied, so there is nothing to roll back (refusing rather than reporting a no-op as success)';
    END IF;
END $$;

UPDATE agent_definitions a
SET default_config = b.default_config,
    updated_at = NOW()
FROM agent_definitions_bak_694 b
WHERE a.id = b.id
  AND a.type = 'content-quality-auditor' AND a.is_active
  AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL;

DO $$
DECLARE q text; p text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'load_page_content'->'config'->>'query',
           default_config->'workflow'->'steps'->'run_content_llm_audit'->'config'->>'prompt'
    INTO q, p
    FROM agent_definitions
    WHERE type = 'content-quality-auditor' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- Assert the PRE-694 shape is back, positively — not merely that 694's
    -- markers are gone, which an empty config would also satisfy.
    IF position($a$p.name IN ('index', 'about', 'services', 'contact')$a$ in q) = 0 THEN
        RAISE EXCEPTION '694 ROLLBACK: verify failed — the original four-name allow-list is not back';
    END IF;
    IF position($a$LEFT(string_agg(pc.rendered_html, ' '), 1000)$a$ in q) = 0 THEN
        RAISE EXCEPTION '694 ROLLBACK: verify failed — the original 1000-char sample is not back';
    END IF;
    IF position('TOP 5 most impactful' in p) = 0 THEN
        RAISE EXCEPTION '694 ROLLBACK: verify failed — the original TOP 5 cap is not back';
    END IF;
    -- and 694's additions must be gone
    IF position('6. PROMISE:' in p) > 0 OR position('differentiation|content|audience' in p) > 0
       OR position('ORDER BY pc.position' in q) > 0 THEN
        RAISE EXCEPTION '694 ROLLBACK: verify failed — 694 additions survive in the restored config';
    END IF;
END $$;

COMMIT;

-- Leaves agent_definitions_bak_694 in place deliberately: a rollback that
-- destroys its own evidence cannot be checked afterwards, and re-applying 694
-- would recreate it anyway (CREATE TABLE IF NOT EXISTS).
