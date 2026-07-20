-- extract.sql — the reasoning-dataset extraction query.
--
-- READ-ONLY. Emits one JSON object per line, tagged with "_t", for
-- cmd/reasoningset to transform into training/eval JSONL.
--
-- Run it OUTSIDE the cluster and pipe it in (the claimscan idiom — psql
-- extracts, Go transforms). Do NOT run this as a pod: training_data_export.go:3-8
-- records that file-writing "landed on ephemeral chassis pods" and was retired.
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -At -f - < cmd/reasoningset/extract.sql \
--     | go run ./cmd/reasoningset --labels <labels.json> --out reasoning_v1.jsonl
--
-- Type warning: correlation_id is uuid on orchestration_states, varchar on
-- llm_call_log, text on diagnosis_artifacts. Every join below casts explicitly.
--
-- Emitted row kinds:
--   _t=step   one LLM reasoning call (the spine)
--   _t=trail  one run's coerced verdict trail (for the guard block)
--   _t=bundle one diagnosis bundle body (the real input_state for a verdict)

\pset tuples_only on
\pset format unaligned

-- ── 1. the spine: every reasoning-shaped LLM call ────────────────────────────
SELECT jsonb_build_object(
  '_t',            'step',
  'trajectory_id', l.correlation_id,
  'run_id',        l.orchestration_id,
  'step_name',     l.step_name,
  'created_at',    l.created_at,
  'input_state',   l.prompt_rendered,
  'reasoning_raw', l.response_text,
  'provenance', jsonb_build_object(
      'agent_type',     l.agent_type,
      'model',          l.model,
      'model_resolved', l.model_resolved,
      'max_tokens',     l.max_tokens,
      'input_tokens',   l.input_tokens,
      'output_tokens',  l.output_tokens,
      'latency_ms',     l.latency_ms,
      'success',        l.success,
      'error_message',  l.error_message,
      'created_at',     l.created_at
  )
)::text
FROM llm_call_log l
WHERE l.step_name ~ '^(verdict|review_|propose|repropose|reframe)$'
   OR l.step_name ~ '^review_'
ORDER BY l.correlation_id, l.created_at;

-- ── 2. the coerced verdict trail, one row per diagnosis run ──────────────────
-- Pairs against the raw verdict to yield the guard block. NOTE the trail
-- serialises with GO FIELD NAMES and INTEGER ENUMS (pkg/diagnose/loop.go:79-153
-- carries no json tags): Outcome 0=Unverifiable 1=Confirmed 2=Refuted,
-- Tier 0=static 1=state 2=runtime. collected_data->'verdict'->'result' is the
-- RAW snake_case wire form. Both shapes are handled in Go.
SELECT jsonb_build_object(
  '_t',            'trail',
  'trajectory_id', os.correlation_id::text,
  'run_id',        os.orchestration_id::text,
  'created_at',    os.created_at,
  'trail',         os.collected_data->'route'->'diagnose_state'->'trail',
  'verdict_raw',   os.collected_data->'verdict'->'result',
  'diagnosis',     os.collected_data->'diagnosis'
)::text
FROM orchestration_states os
WHERE os.collected_data ? 'verdict'
ORDER BY os.created_at;

-- ── 3. bundles: the real input_state a verdict step saw ──────────────────────
-- kind='bundle' is upserted per (correlation_id, iteration) by a partial unique
-- index, so there is exactly one current bundle per iteration.
SELECT jsonb_build_object(
  '_t',            'bundle',
  'trajectory_id', da.correlation_id,
  'run_id',        da.orchestration_id,
  'iteration',     da.iteration,
  'created_at',    da.created_at,
  'body',          da.body,
  'metadata',      da.metadata
)::text
FROM diagnosis_artifacts da
WHERE da.kind = 'bundle'
ORDER BY da.correlation_id, da.iteration;
