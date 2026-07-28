-- ============================================================================
-- 260_baseline_snapshots_error_step_handler_owners.sql
--   bugs_closed/086 — the cheap fix the per-handler audit named and did not do
--
-- WHY. The audit could not say where the 45 -> 44 step-level handler drift went,
-- because most of the agents that own those handlers have never been snapshotted:
-- there was nothing to diff against. Its own recommendation was
-- `snapshot_agent(<type>, 'baseline')` on them — one call each, and the next
-- drift becomes a two-table diff instead of an open question. This is that call,
-- made reproducible.
--
-- SIX of the sixteen agents that own live step-level `error_step` handlers had
-- NO row in agent_definitions_backup at all, as of 2026-07-28 19:2xZ:
--   css-patch-agent (4 handlers), improvement-loop (6), tool-auditor (2),
--   design-audit-agent (2), site-review-agent (1), internal-linker (2)
-- — 17 handlers with no baseline. Four more had one older than the 07-26 count.
--
-- WHAT THIS DOES. Takes one snapshot per handler-owning agent under a single
-- reason string, so a future drift question is one query. Purely additive: it
-- writes rows to agent_definitions_backup and changes no live definition.
--
-- The snapshot set is computed from the live table rather than hardcoded, so
-- rerunning it later picks up agents that have since gained handlers. Rerunning
-- is safe — it simply adds a newer baseline.
--
-- NOTE ON THE OVERLOAD (this has already cost one session a wrong finding):
-- two-arg snapshot_agent(type, reason) writes to `agent_definitions_backup`;
-- one-arg snapshot_agent(type) writes an is_snapshot row into `agent_definitions`
-- itself. Looking in the wrong table produced a confident "the safety net does
-- not exist" when it did.
-- ============================================================================

DO $$
DECLARE
  r       record;
  n       int := 0;
BEGIN
  FOR r IN
    SELECT DISTINCT d.type
    FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
    WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
      AND s.value->>'error_step' IS NOT NULL
      AND s.value->'config'->>'error_step' IS NULL
    ORDER BY 1
  LOOP
    PERFORM snapshot_agent(r.type,
      '260_baseline: error_step handler owner, baseline for drift diffs');
    n := n + 1;
  END LOOP;
  RAISE NOTICE 'baselined % handler-owning agents', n;
END $$;

-- post-check 1: one baseline row per handler-owning agent (want 16 as of 07-28)
SELECT count(*) AS baselines_written, min(snapshot_taken_at) AS taken_at
FROM agent_definitions_backup
WHERE snapshot_reason LIKE '260_baseline%';

-- post-check 2: nobody left without one (want 0 rows)
SELECT d.type AS still_unbaselined
FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND s.value->>'error_step' IS NOT NULL AND s.value->'config'->>'error_step' IS NULL
  AND NOT EXISTS (SELECT 1 FROM agent_definitions_backup b
                  WHERE b.type=d.type AND b.snapshot_reason LIKE '260_baseline%')
GROUP BY 1;

-- post-check 3: no live definition was altered — handler total unchanged (want 46)
SELECT count(*) AS live_step_level_handlers
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND value->>'error_step' IS NOT NULL AND value->'config'->>'error_step' IS NULL;

-- ---------------------------------------------------------------------------
-- THE DIFF THIS ENABLES — run it the next time the count moves.
-- It compares the step-level-ONLY handler set, which is the quantity the 086
-- census actually measures. Mind the two traps that made the first attempt at
-- this question produce six false positives:
--   * loop-expanded steps (`%_iter_%`) are synthesised at runtime and are never
--     in a definition;
--   * a step that gains a `config.error_step` twin LEAVES this set without
--     anything being removed — see the correction in bugs_closed/086.
-- ---------------------------------------------------------------------------
-- WITH base AS (
--   SELECT DISTINCT ON (type) type, default_config FROM agent_definitions_backup
--   WHERE snapshot_reason LIKE '260_baseline%' ORDER BY type, snapshot_taken_at DESC
-- ), base_h AS (
--   SELECT b.type, s.key AS step, s.value->>'error_step' AS target
--   FROM base b, jsonb_each(b.default_config->'workflow'->'steps') s
--   WHERE s.value->>'error_step' IS NOT NULL AND s.value->'config'->>'error_step' IS NULL
-- ), live_h AS (
--   SELECT d.type, s.key AS step, s.value->>'error_step' AS target
--   FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
--   WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
--     AND s.value->>'error_step' IS NOT NULL AND s.value->'config'->>'error_step' IS NULL
-- )
-- SELECT 'LOST' AS d, type, step, target FROM base_h
--  EXCEPT ALL SELECT 'LOST', type, step, target FROM live_h
-- UNION ALL
-- SELECT 'GAINED', type, step, target FROM live_h
--  EXCEPT ALL SELECT 'GAINED', type, step, target FROM base_h;
