-- SQL_2026-08-06_fleet_step_token_pressure_task.sql — bugs_open/183 fix candidate 4,
-- generalised. Seeds a CTE-only scheduled task that measures token-cap pressure for
-- EVERY non-review LLM step in the fleet and writes ONE doc_notes row when the
-- flagged set changes. Clone of FIX-058 (council-seat-token-pressure,
-- fixloop_eg_dartsonline/104_TASK_seat_token_pressure_v1.sql); the two tasks
-- partition the fleet exactly: FIX-058 owns step_name LIKE 'review_%', this owns
-- the rest. Same thresholds, so a step moving between the two families keeps its
-- meaning.
--
-- WHY (bugs_open/183): domain-research-classifier/classify_and_extract sat at
-- p95 = 92.5% of its 6000 cap for months — the fleet's only 6000 — then the tail
-- crossed and burned all 3 attempts per site, on multiple sites, in one afternoon.
-- The cap is config, so the failure looked site-specific and was diagnosed per
-- site first. Nothing anywhere watched headroom. This is the leading indicator
-- FIX-058 built for council seats, applied to the population 183 came from.
--
-- WHY A CTE-ONLY TASK: fire_message=false means the pre_query IS the work
-- (cmd/scheduler/main.go — no Kafka, no orchestration, no LLM, no credits, cannot
-- wake a chassis pod). Why a scheduled task and not a commit-time hook: caps are
-- live DB config, changed with no commit at all — RFC_006's decisive argument,
-- reaffirmed by RFC_012's second sitting ("online within the framework").
--
-- DEPARTURES FROM FIX-058, each deliberate:
--  * n >= 5, not n >= 20. 183's own step had 9 in-window calls the day it burned;
--    at n >= 20 this check cannot flag the case it exists for. Council seats run
--    hundreds of calls a fortnight; fleet steps are sparser.
--  * 90-day window, not 14. Same reason, measured: in the 14 days before 183's
--    step first truncated it had run 3 times (n < 5, silent); over its history
--    it showed p95 90% / peak 94% — a 90-day window flags P BEFORE the failure,
--    a 14-day clone only flags T after it. Sparse steps whose prompts grow
--    between runs are precisely the population this bug came from. The
--    current-cap discipline below keeps old rows from going stale-toxic: a cap
--    raise retires the whole superseded population at once.
--  * Current caps come from the OBSERVED stream, not an agent_definitions jsonb
--    path. The definitions read has two known failure shapes: a root ai_service
--    block silently shadows step-level max_tokens (MDL-039 / bugs_open/009), and
--    the cap is resolved through a fallback chain (LCO-002) a single path cannot
--    see. llm_call_log.max_tokens records what the code ACTUALLY resolved.
--    Current pair = the cap of the single most recent call per step_name.
--    NOT per (agent_type, step_name): over a 90-day window that resurrects
--    retired caps, because pre-2026-07-26 rows carry agent_type='generic' and
--    their stale "most recent generic call" re-admits the superseded pair —
--    observed live on classify_and_extract@6000 during this file's own dry run.
--    KNOWN LIMIT, stated: if two agents ever hold ONE step_name at DIFFERENT
--    caps concurrently, the older-idle holder's pair is not measured. Measured
--    2026-08-06: zero such step_names exist (query in the RUNBOOK); an advisory
--    miss there is cheaper than a standing false positive here.
--    Cost, also stated: a cap changed in config with no call since is invisible
--    until the next call; the note names the cap it measured.
--  * Key is (step_name, cap); agent_type is attribution DISPLAY only.
--    llm_call_log.agent_type was relabelled 2026-07-26 (generic -> resolved
--    type), so an agent_type key splits one population at the relabel line.
--
-- INHERITED FROM FIX-058 UNCHANGED:
--  * A truncated call has output_tokens = NULL, the cut stated only in
--    error_message ('output_tokens >= max_tokens' matched 4 of a true 94). A
--    truncated call scores frac = 1.0: it reached its cap.
--  * Thresholds: T any truncation; N near-miss peak >= 95%; P pressure p95 >= 85%.
--    FIX-058's open register question (should near-miss scale with cap size?)
--    remains open and applies here equally.
--  * Event, not heartbeat: subject_key carries an md5 of the flagged set; a
--    persisting condition announces once, an escalation re-announces, and the
--    dedup look-back is 30 days so a cleared-and-returned condition re-fires.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < SQL_2026-08-06_fleet_step_token_pressure_task.sql
-- Remove: DELETE FROM scheduled_tasks WHERE name='fleet-step-token-pressure';
-- Verify: RUNBOOK_step_token_pressure.md in this directory (includes the pinned
--         2026-08-02 known-case test this check MUST flag).

INSERT INTO scheduled_tasks
  (name, description, interval_seconds, target_agent_type, target_topic,
   input_data, fire_message, enabled, pre_query)
VALUES (
  'fleet-step-token-pressure',
  'bugs_open/183 fix candidate 4, generalised from FIX-058. CTE-only: flags any non-review LLM step whose output distribution approaches its token cap (or that has truncated), and writes a doc_notes row when the flagged set changes. A cap raise moves the cliff; this watches the distance to it.',
  21600,
  'generic',
  'system.agent.generic.requests',
  '{}'::jsonb,
  false,
  true,
$PQ$
WITH latest AS (
  -- current cap per step_name: what the single most recent call resolved.
  -- Per step_name, NOT per (agent_type, step_name) — the 2026-07-26 agent_type
  -- relabel means a per-agent latest resurrects retired caps via stale
  -- 'generic'-era rows (observed live in this file's dry run).
  SELECT DISTINCT ON (step_name) step_name, max_tokens AS cap
  FROM llm_call_log
  WHERE step_name NOT LIKE 'review_%' AND max_tokens > 0
    AND created_at > now() - interval '90 days'
  ORDER BY step_name, created_at DESC
), pairs AS (
  -- agent_types at the current cap, for display only (never a measurement key).
  SELECT l.step_name, l.cap,
         (SELECT string_agg(DISTINCT g.agent_type, ',')
            FROM llm_call_log g
           WHERE g.step_name = l.step_name AND g.max_tokens = l.cap
             AND g.created_at > now() - interval '90 days') AS agents
  FROM latest l
), calls AS (
  -- A truncated call has output_tokens NULL and states the cut in error_message,
  -- so 'output_tokens >= max_tokens' can never match one (FIX-058: 4 vs a true
  -- 94). A truncated call scores frac = 1.0: it reached its cap. 'response
  -- truncated:' is TruncatedError's own wording, provider-agnostic; the
  -- stop_reason form is kept for rows predating it.
  SELECT step_name, max_tokens AS cap,
         (error_message ILIKE '%response truncated:%'
          OR error_message ILIKE '%stop_reason=max_tokens%') AS was_truncated,
         COALESCE(output_tokens::numeric / max_tokens,
                  CASE WHEN error_message ILIKE '%response truncated:%'
                         OR error_message ILIKE '%stop_reason=max_tokens%' THEN 1.0 END) AS frac
  FROM llm_call_log
  WHERE created_at > now() - interval '90 days'
    AND step_name NOT LIKE 'review_%'
    AND max_tokens > 0
    AND (output_tokens IS NOT NULL
         OR error_message ILIKE '%response truncated:%'
         OR error_message ILIKE '%stop_reason=max_tokens%')
), agg AS (
  SELECT p.step_name, p.cap, p.agents,
         count(c.frac) AS n,
         round(100*(percentile_cont(0.95) WITHIN GROUP (ORDER BY c.frac))::numeric, 1) AS p95,
         round(100*max(c.frac), 1) AS peak,
         count(*) FILTER (WHERE c.was_truncated) AS trunc
  FROM pairs p
  JOIN calls c ON c.step_name = p.step_name AND c.cap = p.cap
  GROUP BY 1, 2, 3
), flagged AS (
  -- n >= 5, not FIX-058's 20: fleet steps are sparser than council seats, and at
  -- 20 this check cannot flag its own motivating case (11 calls in 21 days).
  SELECT *, CASE WHEN trunc > 0 THEN 'T' WHEN peak >= 95 THEN 'N' ELSE 'P' END AS kind
  FROM agg
  WHERE n >= 5 AND (trunc > 0 OR peak >= 95 OR p95 >= 85)
), fp AS (
  SELECT count(*) AS n_flagged,
         md5(string_agg(step_name || '@' || cap || ':' || kind, '|' ORDER BY step_name, cap)) AS digest,
         string_agg(kind || '  ' || step_name || '@' || cap ||
                    ' — n=' || n || ', p95 ' || p95 || '%, peak ' || peak ||
                    '%, truncated ' || trunc || ' — agents: ' || agents,
                    chr(10) ORDER BY trunc DESC, peak DESC) AS lines
  FROM flagged
), ins AS (
  INSERT INTO doc_notes
    (subject_type, subject_key, body, categories, source, source_agent, created_by)
  SELECT 'pipeline',
         'fleet-step-token-pressure:' || fp.digest,
         'FLEET STEP TOKEN PRESSURE — ' || fp.n_flagged || ' step/cap pair(s) flagged (non-review steps; council seats are council-seat-token-pressure''s).' || chr(10) || chr(10) ||
         fp.lines || chr(10) || chr(10) ||
         'T = has truncated at this cap.  N = near-miss, peak >= 95% of cap.  P = pressure, p95 >= 85% of cap.  frac scores a truncated call as 1.0; truncations are counted from error_message because a truncated call logs output_tokens NULL.' || chr(10) || chr(10) ||
         'WHY IT MATTERS (bugs_open/183): a step''s cap is config, so when the output tail finally crosses it, the failure presents as SITE-specific — several unrelated domains failing the same step in the same window, each burning its max_attempts. classify_and_extract sat at p95 92.5% of its cap for months before doing exactly that. A cap raise MOVES the cliff (measured twice on council seats); this note is the distance to it. Before raising a cap, check who owns the agent — every raise so far has been an owner call.' || chr(10) || chr(10) ||
         'A flagged step that emits a whole document may deserve a smaller unit of work rather than a bigger cap — see bugs_open/183 fix candidate 3.' || chr(10) || chr(10) ||
         'Runbook: docs/agent_docs/docs024_key_docs_latest/bugfix_183_step_token_pressure/RUNBOOK_step_token_pressure.md' || chr(10) ||
         'Thresholds: SELECT pre_query FROM scheduled_tasks WHERE name=''fleet-step-token-pressure'';',
         '["step-token-pressure","fleet","bugs_open/183"]'::jsonb,
         'scheduled_task:fleet-step-token-pressure',
         'scheduler',
         'fleet-step-token-pressure'
  FROM fp
  WHERE fp.n_flagged > 0
    AND NOT EXISTS (
      SELECT 1 FROM doc_notes d
      WHERE d.subject_type = 'pipeline'
        AND d.subject_key = 'fleet-step-token-pressure:' || fp.digest
        AND d.created_at > now() - interval '30 days'
    )
  RETURNING id
)
SELECT id::text AS note_id FROM ins;
$PQ$
);
