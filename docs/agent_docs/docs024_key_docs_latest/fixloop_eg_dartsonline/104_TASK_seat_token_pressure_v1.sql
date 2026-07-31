-- 104_TASK_seat_token_pressure_v1.sql — the AUTOMATIC half of bugs_open/138
-- candidate 2. Seeds a CTE-only scheduled task that measures council seat token
-- pressure every 6 hours and writes ONE doc_notes row when the flagged set changes.
--
-- WHY A CTE-ONLY TASK AND NOT AN AGENT. `fire_message = false` means the pre_query
-- IS the work (cmd/scheduler/main.go:266) — no Kafka message, no orchestration, no
-- LLM call, no credits. The whole check is one query, so paying an agent to run it
-- would be paying for a wrapper. It also means it cannot wake a chassis pod or
-- collide with the 300s post-restart dispatch window.
--
-- WHY IT IS AN EVENT, NOT A HEARTBEAT. `subject_key` carries an md5 of the flagged
-- set (seat@cap plus which threshold it crossed), and the insert is skipped when a
-- note with that exact key already exists. So a condition that persists — and
-- `review_guardian` will sit at a 99.2% peak for as long as the 14-day window holds
-- it — is announced ONCE, while a NEW seat crossing, or an existing one escalating
-- from near-miss to truncated, changes the digest and announces itself. Without
-- this the task would write an identical note every 6 hours and be ignored inside a
-- day, which is the failure mode of every alert nobody has tuned.
--
-- The dedup look-back is bounded at 30 days so a condition that clears and later
-- returns is announced again rather than being permanently silenced by its own
-- historical note.
--
-- THRESHOLDS LIVE HERE AND NOWHERE ELSE. 104_REPORT_seat_token_pressure_v1.sh
-- prints numbers and does not re-encode the rule; it points readers at this
-- pre_query. Two copies of a threshold is the drift class 099/102 exist to fight.
-- Their derivation from the live distribution is in the report script's header.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < 104_TASK_seat_token_pressure_v1.sql
-- Remove: DELETE FROM scheduled_tasks WHERE name='council-seat-token-pressure';

INSERT INTO scheduled_tasks
  (name, description, interval_seconds, target_agent_type, target_topic,
   input_data, fire_message, enabled, pre_query)
VALUES (
  'council-seat-token-pressure',
  'bugs_open/138 candidate 2. CTE-only: flags council review seats close to their token cap and writes a doc_notes row when the flagged set changes. A truncated review gates its round regardless of severity, so a seat being cut off looks like a noisy seat.',
  21600,
  'generic',
  'system.agent.generic.requests',
  '{}'::jsonb,
  false,
  true,
$PQ$
WITH live AS (
  SELECT a.type AS council, s.key AS seat,
         (s.value->'config'->'ai_service'->>'max_tokens')::int AS cap
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.key LIKE 'review_%'
    AND (s.value->'config'->'ai_service'->>'max_tokens') IS NOT NULL
), pairs AS (
  SELECT seat, cap, string_agg(DISTINCT council, ',') AS councils
  FROM live GROUP BY 1, 2
), calls AS (
  SELECT step_name AS seat, max_tokens AS cap, agent_type,
         output_tokens::numeric / max_tokens AS frac
  FROM llm_call_log
  WHERE created_at > now() - interval '14 days'
    AND step_name LIKE 'review_%'
    AND max_tokens > 0 AND output_tokens IS NOT NULL
), agg AS (
  SELECT p.seat, p.cap, p.councils,
         count(c.frac) AS n,
         round(100*(percentile_cont(0.95) WITHIN GROUP (ORDER BY c.frac))::numeric, 1) AS p95,
         round(100*max(c.frac), 1) AS peak,
         count(*) FILTER (WHERE c.frac >= 1) AS trunc,
         count(c.frac) FILTER (WHERE c.agent_type = ANY(string_to_array(p.councils, ','))) AS n_holder
  FROM pairs p
  LEFT JOIN calls c ON c.seat = p.seat AND c.cap = p.cap
  GROUP BY 1, 2, 3
), flagged AS (
  SELECT *, CASE WHEN trunc > 0 THEN 'T' WHEN peak >= 95 THEN 'N' ELSE 'P' END AS kind
  FROM agg
  WHERE n >= 20 AND (trunc > 0 OR peak >= 95 OR p95 >= 85)
), fp AS (
  SELECT count(*) AS n_flagged,
         md5(string_agg(seat || '@' || cap || ':' || kind, '|' ORDER BY seat, cap)) AS digest,
         string_agg(kind || '  ' || seat || '@' || cap ||
                    ' — n=' || n || ' (holder ' || n_holder || '), p95 ' || p95 ||
                    '%, peak ' || peak || '%, truncated ' || trunc ||
                    ' — cap held by ' || councils,
                    chr(10) ORDER BY trunc DESC, peak DESC) AS lines
  FROM flagged
), ins AS (
  INSERT INTO doc_notes
    (subject_type, subject_key, body, categories, source, source_agent, created_by)
  SELECT 'pipeline',
         'council-seat-token-pressure:' || fp.digest,
         'COUNCIL SEAT TOKEN PRESSURE — ' || fp.n_flagged || ' seat/cap pair(s) flagged.' || chr(10) || chr(10) ||
         fp.lines || chr(10) || chr(10) ||
         'T = has truncated at this cap.  N = near-miss, peak >= 95% of cap.  P = pressure, p95 >= 85% of cap.' || chr(10) || chr(10) ||
         'WHY IT MATTERS (bugs_open/138): a review cut off at max_tokens is recovered, marked degraded, and a degraded "object" GATES its round to revise regardless of the severities that survived. So an over-long ADVISORY review silently becomes a BLOCKING one — and a high object rate is also the documented kill-switch for retiring a noisy seat. A seat can be pulled for being noisy when it was being cut off. Before judging a seat by its object rate, check whether it was truncating.' || chr(10) || chr(10) ||
         'READ n_holder BEFORE ACTING. It counts only the calls LABELLED with a council that still holds that cap. llm_call_log logged every review call as agent_type=generic before 2026-07-26 14:54, and fix-proposer has never appeared at all, so n_holder is a LOWER BOUND for the fix-lane councils and exact for the others. A flag whose n_holder is small is inferred from a sibling council at the same cap — a reason to look, not a finding.' || chr(10) || chr(10) ||
         'Two ways out, and they are not equivalent: raising the cap MOVES the cliff (review_architecture was raised to 16000 and a longer prompt reintroduced truncation against the new cap within hours); a length budget in the prompt removes the pressure. Measured together on that seat they took peak output to 28% of cap.' || chr(10) || chr(10) ||
         'Full report: docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/104_REPORT_seat_token_pressure_v1.sh' || chr(10) ||
         'Thresholds: SELECT pre_query FROM scheduled_tasks WHERE name=''council-seat-token-pressure'';',
         '["seat-token-pressure","council-gate","bugs_open/138"]'::jsonb,
         'scheduled_task:council-seat-token-pressure',
         'scheduler',
         'council-seat-token-pressure'
  FROM fp
  WHERE fp.n_flagged > 0
    AND NOT EXISTS (
      SELECT 1 FROM doc_notes d
      WHERE d.subject_type = 'pipeline'
        AND d.subject_key = 'council-seat-token-pressure:' || fp.digest
        AND d.created_at > now() - interval '30 days'
    )
  RETURNING id
)
SELECT id::text AS note_id FROM ins;
$PQ$
);
