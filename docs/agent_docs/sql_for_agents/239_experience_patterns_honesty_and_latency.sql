-- ============================================================================
-- 239_experience_patterns_honesty_and_latency.sql
--
-- Two columns the harvested entries need and the table did not have. Found the
-- only way such a thing is ever found honestly: by trying to load the nine real
-- harvested entries through the live write path, on 2026-07-27, instead of
-- through an example written to satisfy the validator.
--
--   honesty_clauses (jsonb array) — CC-006 count-up-stat-band carries three,
--     e.g. "the animation never invents digits — it counts only to a value that
--     was already rendered from authored data". These are not states and not
--     degraded states: they are clauses about what the component must never
--     ASSERT. The register's whole purpose is to stop a page claiming something
--     it has not got, so a field for that is not decoration.
--
--   latency_envelope (jsonb object) — MJ-002 timed-remote-challenge-loop
--     carries the measured 8–23 s per generated response, with the consequence
--     that acceptance checks for this pattern REQUIRE a wait or poll. That is
--     the fact which makes two checks in the council-APPROVED gauntlet
--     EXPERIENCE_PLAN unexecutable today (the runner asserts 300 ms after the
--     last step). Storing it on the entry is what lets a binding decide that a
--     check must be DEFERRED rather than discovering it against a live page.
--
-- WHY THIS IS A CORRECTION, NOT AN ADDITION
--   Migration 218's own header says every column below the taxonomy axes exists
--   "because a LIVE implementation needed it and the 2026-07-24 design could not
--   express it". These two are the same claim, made two days later against the
--   same evidence — which means 218 was written from the harvest NOTES rather
--   than from the harvested entry files themselves. The workstream's own thesis
--   is that harvesting bottom-up catches invented shapes; it caught this one,
--   late.
--
-- ORDERING: safe ahead of the image, on the same reasoning as 230 — additive
-- columns with defaults, no constraint tightened, no existing writer affected
-- (verified: only write_experience_pattern_action.go writes this table).
-- ============================================================================

BEGIN;

ALTER TABLE experience_patterns
  ADD COLUMN IF NOT EXISTS honesty_clauses jsonb NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE experience_patterns
  ADD COLUMN IF NOT EXISTS latency_envelope jsonb NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN experience_patterns.honesty_clauses IS
  'Clauses about what this experience must never ASSERT (distinct from states and degraded_states, which describe what it must never DO). Harvested from CC-006.';
COMMENT ON COLUMN experience_patterns.latency_envelope IS
  'Measured timing and its consequence for acceptance checks. When a pattern is slower than the runner asserts (stepDelay 300ms), its checks must be DEFERRED rather than failed. Harvested from MJ-002.';

DO $guard$
DECLARE
    missing text;
BEGIN
    SELECT string_agg(c, ', ') INTO missing
      FROM unnest(ARRAY['honesty_clauses','latency_envelope']) AS c
     WHERE NOT EXISTS (SELECT 1 FROM information_schema.columns
                        WHERE table_name = 'experience_patterns' AND column_name = c);
    IF missing IS NOT NULL THEN
        RAISE EXCEPTION '239: column(s) not added: %', missing;
    END IF;

    -- The approval constraint from 230 must survive an ALTER on this table.
    IF NOT EXISTS (SELECT 1 FROM pg_constraint
                    WHERE conname = 'experience_patterns_approved_needs_executable_check') THEN
        RAISE EXCEPTION '239: migration 230s approval constraint is gone after this ALTER';
    END IF;
END
$guard$;

COMMIT;

-- Verify
SELECT column_name, data_type, column_default FROM information_schema.columns
WHERE table_name = 'experience_patterns'
  AND column_name IN ('honesty_clauses','latency_envelope')
ORDER BY column_name;

-- Rollback recipe (hand-run):
--   ALTER TABLE experience_patterns DROP COLUMN IF EXISTS honesty_clauses;
--   ALTER TABLE experience_patterns DROP COLUMN IF EXISTS latency_envelope;
