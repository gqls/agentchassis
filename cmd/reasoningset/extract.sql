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

-- ── 4. the council lane: one row per (report x seat) ─────────────────────────
-- 5.6x the diagnosis lane and, until now, unextracted. Each row is one
-- reviewer's argument about one plan, with the round's decision alongside — so
-- dissent (a seat voting against its round's outcome) is computable without any
-- new labelling. That is the signal: where seats split, the reasoning is
-- load-bearing.
--
-- CAUTION on the semantics of 'approve'. Some seats are relevance-gated and
-- approve because the change is OUTSIDE their footprint, writing a scope
-- determination rather than a merit judgement ("doesn't touch classifier logic,
-- revenue-shape decisions..."). Six seats have never objected. Sampled, they are
-- substantive, NOT rubber stamps — but their approve does not mean what
-- edit-quality's approve means. seat_never_objects is emitted so a consumer can
-- separate the two rather than pooling them.
WITH seat_stats AS (
  SELECT r->>'reviewer' AS seat,
         count(*) AS votes,
         count(*) FILTER (WHERE r->>'verdict' <> 'approve') AS non_approvals
  FROM diagnosis_artifacts da, jsonb_array_elements(da.body::jsonb->'reviews') r
  WHERE da.kind = 'council_report'
  GROUP BY 1
)
SELECT jsonb_build_object(
  '_t',               'council',
  'trajectory_id',    da.correlation_id,
  'run_id',           da.orchestration_id,
  'created_at',       da.created_at,
  'round_decision',   da.metadata->>'decision',
  'reviewers',        (da.metadata->>'reviewers'),
  'abstained',        (da.metadata->>'abstained'),
  'seat',             r->>'reviewer',
  'seat_verdict',     r->>'verdict',
  'objections',       COALESCE(r->'objections', '[]'::jsonb),
  'missing',          COALESCE(r->'missing', '[]'::jsonb),
  'notes',            COALESCE(r->>'notes', ''),
  'seat_votes',       ss.votes,
  'seat_never_objects', (ss.non_approvals = 0)
)::text
FROM diagnosis_artifacts da,
     jsonb_array_elements(da.body::jsonb->'reviews') r
     JOIN seat_stats ss ON ss.seat = r->>'reviewer'
WHERE da.kind = 'council_report'
ORDER BY da.created_at, r->>'reviewer';

-- ── 5. the feed-triage lane ─────────────────────────────────────────────────
-- The best-shaped non-loop reasoning source on the platform, and the only one
-- where the judgement AND its outcome sit on the same row.
--
-- One call scores MANY items, so 439 calls carry thousands of judgements. The
-- response is a JSON array and is exploded in Go, NOT here: 23 successful calls
-- carry malformed/truncated JSON that the output_tokens>=max_tokens rule does
-- NOT catch (0 flagged), so a SQL-side ::jsonb cast aborts the whole query on
-- the first bad row. Go skips and counts them instead.
--
-- The per-item `reason` (why this score) is ONLY in response_text — the table
-- persists credibility_reason but not reason. Both are needed for the full
-- judgement, which is why this lane takes two queries.
SELECT jsonb_build_object(
  '_t',            'feedjudge',
  'run_id',        l.orchestration_id,
  'trajectory_id', l.correlation_id,
  'created_at',    l.created_at,
  'reasoning_raw', l.response_text,
  'provenance', jsonb_build_object(
      'agent_type', l.agent_type, 'model', l.model, 'model_resolved', l.model_resolved,
      'max_tokens', l.max_tokens, 'input_tokens', l.input_tokens,
      'output_tokens', l.output_tokens, 'success', l.success,
      'error_message', l.error_message, 'created_at', l.created_at)
)::text
FROM llm_call_log l
WHERE l.step_name = 'score_relevance'
ORDER BY l.created_at;

-- The outcome side. status is NOT a restatement of the score: `rejected` spans
-- 0-90 and 494 items scoring >=60 were rejected anyway, with duplicate_of and
-- expiry explaining NONE of them. So something downstream decides, and what it
-- is remains an open question — do not label these as score-derived.
-- source_content is deliberately excluded (large, and the title+summary is what
-- the triage prompt actually sees).
SELECT jsonb_build_object(
  '_t',                 'feeditem',
  'item_id',            c.id,
  'site_id',            c.site_id,
  'status',             c.status,
  'relevance_score',    c.relevance_score,
  'credibility',        c.credibility,
  'credibility_reason', c.credibility_reason,
  'source_attribution', c.source_attribution,
  'topics',             c.topics,
  'source_title',       c.source_title,
  'source_summary',     left(COALESCE(c.source_summary, ''), 600),
  'created_at',         c.created_at
)::text
FROM content_feed_items c
WHERE c.relevance_score IS NOT NULL
ORDER BY c.created_at;
