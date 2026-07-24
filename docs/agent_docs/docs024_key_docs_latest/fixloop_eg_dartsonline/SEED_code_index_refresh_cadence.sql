-- SEED_code_index_refresh_cadence.sql
-- 2026-07-24 (chat "diagnosis fixloop 5") — bugs_open/059 fix #1: a reindex cadence.
--
-- code_symbols (read by the diagnosis code tier AND the prior_art council seat) had
-- NO refresh cadence and went 3 weeks stale, so every code-existence check silently
-- answered "absent" for anything built since 2026-07-02. This scheduled_task fires
-- index-orchestrator on a cadence so the index tracks the deployed code.
--
-- WHY A DERIVED ref, NOT A HARDCODED ONE. The deployed code lives on a feature
-- branch whose NAME churns (085_… → 086_… within a week). A hardcoded ref would BE
-- the "snapshot that drifts" class THIS bug is about — forget to bump it and the
-- index silently indexes a dead branch. So the ref is DERIVED at fire time by the
-- pre_query: the most-recent HUMAN/diagnosis-driven feature-branch ref (NNN_*) —
-- i.e. the branch the fleet is actually diagnosing on, which is exactly the ref the
-- 090 diagnosis loop reads. The scheduler takes the pre_query's first row and
-- shallow-merges {ref: …} into input_data (cmd/scheduler/main.go runPreQuery +
-- mergeJSON), and index-orchestrator's call_indexer maps input_data.ref → the
-- spawned code-indexer.
--
-- Two guards baked into the pre_query:
--   * SELF-EXCLUSION: owner_agent_type NOT IN ('index-orchestrator','code-indexer')
--     — the reindex runs several times a day carrying its own ref; without this
--     they would drown out the real signal and pin the derivation to a dead branch
--     when work moves on (re-creating a staleness bug in the fix for one).
--   * GATE: no rows (a quiet repo with no recent human code-ref activity) → the
--     tick is SKIPPED. Better to not reindex than to index a guessed ref.
-- Cheap: orchestration_states is kept small by the reaper (~1.5k rows), scan is trivial.
--
-- interval 86400s (24h) — matches the fleet's other freshness tasks and decisively
-- fixes the 3-week problem (worst case now <1 day behind). Tighten via the one
-- column if the index must stay within a single image roll (images roll several
-- times a day); tying to the actual roll would need a deploy hook (059 fix #3-adjacent).
-- The code-indexer re-embeds only CHANGED symbols, so a no-change run is a cheap
-- no-op upsert.
--
-- Idempotent (ON CONFLICT (name) DO UPDATE). Live on apply — no image roll.

INSERT INTO scheduled_tasks
  (name, description, interval_seconds, target_agent_type, target_topic,
   input_data, pre_query, concurrency_group, max_concurrent, timeout_seconds,
   enabled, fire_message)
VALUES (
  'code-index-refresh',
  'Reindex code_symbols on a cadence (bugs_open/059): fires index-orchestrator with a ref DERIVED from the most recent human/diagnosis-driven feature-branch orchestration, so the diagnosis code tier and the prior_art council seat read current code.',
  86400,
  'index-orchestrator',
  'system.agent.generic.requests',
  '{"owner": "gqls", "repo": "agentchassis", "language": "go"}'::jsonb,
  $q$SELECT collected_data->'input_data'->>'ref' AS ref
       FROM orchestration_states
      WHERE collected_data->'input_data'->>'ref' ~ '^[0-9]{3}_'
        AND COALESCE(owner_agent_type,'') NOT IN ('index-orchestrator','code-indexer')
        AND created_at > now() - interval '14 days'
      ORDER BY created_at DESC
      LIMIT 1$q$,
  'code-index',
  1,
  1800,
  true,
  true
)
ON CONFLICT (name) DO UPDATE SET
  description      = EXCLUDED.description,
  interval_seconds = EXCLUDED.interval_seconds,
  target_agent_type= EXCLUDED.target_agent_type,
  target_topic     = EXCLUDED.target_topic,
  input_data       = EXCLUDED.input_data,
  pre_query        = EXCLUDED.pre_query,
  concurrency_group= EXCLUDED.concurrency_group,
  max_concurrent   = EXCLUDED.max_concurrent,
  timeout_seconds  = EXCLUDED.timeout_seconds,
  enabled          = EXCLUDED.enabled,
  fire_message     = EXCLUDED.fire_message,
  updated_at       = now();

-- Verification (run after apply):
--   SELECT name, enabled, interval_seconds, target_agent_type, input_data, last_triggered_at
--   FROM scheduled_tasks WHERE name='code-index-refresh';
-- And confirm the pre_query resolves to the live branch:
--   <the pre_query above>  -- expect the current NNN_ branch, one row.
