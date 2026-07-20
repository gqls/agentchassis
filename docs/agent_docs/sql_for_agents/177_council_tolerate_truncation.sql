-- 177: arm tolerate_truncation on the three fix/feature/gate councils' review seats
--
-- bugs_open/019. The Go side (commit a3b606798) makes a truncated reviewer
-- degrade instead of voiding the round — but tolerance is OPT-IN per step via
-- config, so without this flag the code change is a no-op and every review_*
-- step still hard-fails to complete_invalid / complete_refused on overrun.
--
-- Scope is DELIBERATELY narrow (owner decision, 2026-07-20): council review_*
-- steps only. NOT experience-planner, NOT compose/propose/design steps — their
-- consumers cannot handle a partial result; only the council decider knows how
-- to treat an unreadable seat (loud abstention, approve->revise downgrade).
--
-- PATCH-STYLE on purpose: jsonb_set adds the single key to each live step
-- config. Never re-seed whole rows here — the roster moves fast (5 changes in
-- 18h once) and a full-row seed clobbers concurrent changes. Iterates whatever
-- review_* steps exist at run time, so seat additions (e.g. v19's
-- review_constitution / review_mission) are covered without maintenance.
--
-- Safe to apply BEFORE the image roll: the old binary ignores the key. This is
-- the config half of a fix that only becomes live when both halves are in.
--
-- Verify:
--   SELECT type, k, v->'config'->>'tolerate_truncation'
--   FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps') e(k,v)
--   WHERE type IN ('council-gate','fix-proposer','feature-designer')
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
--     AND k LIKE 'review\_%' ORDER BY type, k;
--   -- expect: true on every row (13 + 15 + 5 seats as of 2026-07-20)

DO $$
DECLARE
    r RECORD;
    step_key TEXT;
BEGIN
    FOR r IN
        SELECT id, type, default_config
        FROM agent_definitions
        WHERE type IN ('council-gate', 'fix-proposer', 'feature-designer')
          AND COALESCE(is_snapshot, false) = false
          AND deleted_at IS NULL
    LOOP
        FOR step_key IN
            SELECT k FROM jsonb_object_keys(r.default_config->'workflow'->'steps') AS k
            WHERE k LIKE 'review\_%'
        LOOP
            UPDATE agent_definitions
            SET default_config = jsonb_set(
                    default_config,
                    ARRAY['workflow', 'steps', step_key, 'config', 'tolerate_truncation'],
                    'true'::jsonb,
                    true),
                updated_at = now()
            WHERE id = r.id;
            RAISE NOTICE 'set tolerate_truncation on %.%', r.type, step_key;
        END LOOP;
    END LOOP;
END $$;
