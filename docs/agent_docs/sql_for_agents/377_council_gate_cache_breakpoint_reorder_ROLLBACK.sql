-- 377_council_gate_cache_breakpoint_reorder_ROLLBACK.sql
--
-- Restores the council-gate agent config captured by 377 before it reordered
-- the 17 seat templates.
--
-- WHY RESTORE BYTES RATHER THAN REVERSE THE MOVE. The forward migration hoists
-- a shared block from a position that DIFFERS per seat. Reversing it would mean
-- rediscovering 17 original offsets and putting each block back exactly where
-- it came from — a second opportunity to be wrong, on live config that gates
-- every platform change in this estate. The backup table holds the exact prior
-- bytes, so restoring is a copy, not a derivation.
--
-- ⚠ SAME IN-FLIGHT RULE AS THE FORWARD MIGRATION: do not run this while a
-- council orchestration is mid-chain, or one verdict ends up assembled from two
-- different prompt generations with nothing downstream showing it.
--
-- ⚠ THIS DOES NOT UNDO THE GO SIDE, AND DOES NOT NEED TO. The client seam
-- (LCO-008) is inert without the marker: once these templates no longer carry
-- it, the request body is byte-identical to the pre-caching shape again, on
-- whatever chassis image happens to be running. Rolling back config alone is
-- therefore a complete rollback of the behaviour — that is the point of making
-- the seam opt-in rather than a mode flag.

BEGIN;

DO $$
DECLARE
    n integer;
BEGIN
    SELECT count(*) INTO n FROM information_schema.tables
     WHERE table_name = 'bak_council_gate_config_20260810';
    IF n <> 1 THEN
        RAISE EXCEPTION
          'ROLLBACK 377 REFUSED: backup table bak_council_gate_config_20260810 does not exist. Restoring from nothing would blank the council config; fix the backup first.';
    END IF;
END $$;

UPDATE agent_definitions d
SET default_config = b.default_config,
    version        = d.version + 1,   -- forward-only: a restore is a new version, never a rewind
    updated_at     = now()
FROM bak_council_gate_config_20260810 b
WHERE d.id = b.id;

-- Verify the marker is genuinely gone from every seat. A partial restore is the
-- dangerous outcome: some seats cached, some not, and the prefix set no longer
-- uniform — which reads as "caching is flaky" rather than "the rollback missed".
DO $$
DECLARE
    n_marker integer;
    n_seats  integer;
BEGIN
    SELECT count(*),
           count(*) FILTER (WHERE s.value->'config'->>'prompt_template' LIKE '%<!--CACHE_BREAKPOINT-->%')
      INTO n_seats, n_marker
    FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
    WHERE d.type='council-gate' AND d.is_active
      AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
      AND s.key LIKE 'review\_%';

    IF n_marker <> 0 THEN
        RAISE EXCEPTION 'ROLLBACK 377 INCOMPLETE: %/% seats still carry the cache marker', n_marker, n_seats;
    END IF;
    RAISE NOTICE 'rollback 377 OK: marker absent from all % seats; council prompts restored to pre-377 bytes', n_seats;
END $$;

COMMIT;
