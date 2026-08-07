-- SQL_2026-08-06c_close_the_clock_blind_spot.sql
-- Closes two gaps in `fleet-step-token-pressure` (LCO-007), found by ENUMERATING the
-- error families in llm_call_log instead of writing a predicate from what I expected
-- to find. That enumeration is the check that should have preceded v1:
--
--   SELECT left(error_message,52), count(*) FROM llm_call_log
--    WHERE created_at > now()-interval '90 days' AND error_message IS NOT NULL
--    GROUP BY 1 ORDER BY 2 DESC;
--
-- GAP 1 — a truncation the vocabulary misses. `RETRY (bugs_open/119) TRUNCATED and
-- tolerated` is a genuine cap-reaching truncation (output_tokens = max_tokens = 120 on
-- the observed row) and contains NEITHER `response truncated:` NOR `stop_reason=max_tokens`,
-- so v1 scored it as an ordinary call. The sibling wrappers are fine and were already
-- matched: `TOLERATED (step continued on the partial): response truncated: …` and
-- `REFUSED (bugs_open/076 …): response truncated: …` both carry the original text.
--
-- GAP 2 — the clock is a SECOND limit and v1 could not see it at all. The chassis does
-- not stream (`aiservice/anthropic.go:72`, gemini.go:185 — 600s; ollama.go:55 — 120s),
-- so a long generation can die on wall-clock instead of on its cap. Such a call logs
-- `output_tokens = NULL` and an error matching none of the truncation strings, so v1's
-- WHERE excluded it from the population entirely — a step degrading from truncating to
-- timing out would have looked like a step that IMPROVED.
--
-- WHY 'C' IS A SEPARATE KIND AND NOT FOLDED INTO 'T' — this is the load-bearing design
-- decision. A truncation says "the cap was reached": the remedy is a bigger cap or a
-- smaller unit of work. A clock kill says "the cap could NOT be reached": raising the
-- cap is then actively WRONG advice, because the new number is unreachable too. Scoring
-- a clock kill as frac = 1.0 would have merged the two and produced exactly that wrong
-- recommendation. 'C' therefore ranks ABOVE 'T' when a step has both.
--
-- WHY THE CLOCK ARM DOES NOT REQUIRE A KNOWN CAP. The only real clock-exhaustion case
-- in the history — `scrape_prices`, 246 calls, April 2026, peak latency 600,001 ms,
-- i.e. the literal HTTP timeout — recorded **no max_tokens at all**. Keying the clock
-- arm on (step, cap) would have made the detector blind to the one case that proves it
-- works. Clock exhaustion is not a property of the cap, so it is keyed on step alone.
--
-- WHY `context canceled` NEEDS A LATENCY FLOOR. 34 such rows in the last 90 days are
-- ordinary caller-side cancellations — median latency 23s, peak 110s — nothing to do
-- with the clock. Only cancellations at >= 480,000 ms (80% of the 600s timeout) are
-- treated as clock exhaustion. Without that floor this arm would fire on every pod
-- restart. [MEASURED 2026-08-06: 74 cancellations all-history, of which 16 are >= 480s.]
--
-- HONEST SCOPE — the clock has NEVER fired on Anthropic. All 231 `context deadline
-- exceeded` / `Client.Timeout` rows are ollama, April 2026, and in the last 90 days
-- there are zero of either and the peak latency is 495,177 ms (82.5% of the limit).
-- So GAP 2 is a LATENT hole being armed before it is needed, not a live incident.
-- The 2026-08-06b summary overstated this by citing the recent `context canceled` rows
-- as evidence of clock pressure; they are not, and that claim is corrected there.
--
-- NOT DONE HERE, DELIBERATELY: FIX-058 (`council-seat-token-pressure`) has GAP 1
-- identically — its vocabulary is `stop_reason=max_tokens` only, and the one observed
-- `TRUNCATED and tolerated` row is on `review_adoption_guardian`, i.e. squarely in ITS
-- population, not this one. It is another lane's instrument and unilaterally editing a
-- live shared check is the config-clobber pattern this estate keeps getting bitten by.
-- Flagged for its owner rather than fixed.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < SQL_2026-08-06c_close_the_clock_blind_spot.sql
-- Revert: re-apply SQL_2026-08-06_fleet_step_token_pressure_task.sql after a DELETE.

BEGIN;

-- Guard: the task must exist and still be the v1 we are amending.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='fleet-step-token-pressure') THEN
    RAISE EXCEPTION 'ABORT: fleet-step-token-pressure does not exist — apply the v1 seed first.';
  END IF;
  IF EXISTS (SELECT 1 FROM scheduled_tasks
              WHERE name='fleet-step-token-pressure' AND pre_query ILIKE '%died on the CLOCK%') THEN
    RAISE EXCEPTION 'ABORT: already amended (clock arm present) — another session may have applied this.';
  END IF;
END $$;

UPDATE scheduled_tasks SET pre_query = $PQ$
WITH latest AS (
  SELECT DISTINCT ON (step_name) step_name, max_tokens AS cap
  FROM llm_call_log
  WHERE step_name NOT LIKE 'review_%' AND max_tokens > 0
    AND created_at > now() - interval '90 days'
  ORDER BY step_name, created_at DESC
), pairs AS (
  SELECT l.step_name, l.cap,
         (SELECT string_agg(DISTINCT g.agent_type, ',') FROM llm_call_log g
           WHERE g.step_name = l.step_name AND g.max_tokens = l.cap
             AND g.created_at > now() - interval '90 days') AS agents
  FROM latest l
), calls AS (
  -- Truncation vocabulary, all three live wrapper forms. 'TRUNCATED and tolerated' is
  -- the bugs_open/119 retry form and carries neither of the other two strings.
  SELECT step_name, max_tokens AS cap,
         (error_message ILIKE '%response truncated:%'
          OR error_message ILIKE '%stop_reason=max_tokens%'
          OR error_message ILIKE '%TRUNCATED and tolerated%') AS was_truncated,
         COALESCE(output_tokens::numeric / max_tokens,
                  CASE WHEN error_message ILIKE '%response truncated:%'
                         OR error_message ILIKE '%stop_reason=max_tokens%'
                         OR error_message ILIKE '%TRUNCATED and tolerated%' THEN 1.0 END) AS frac
  FROM llm_call_log
  WHERE created_at > now() - interval '90 days'
    AND step_name NOT LIKE 'review_%' AND max_tokens > 0
    AND (output_tokens IS NOT NULL
         OR error_message ILIKE '%response truncated:%'
         OR error_message ILIKE '%stop_reason=max_tokens%'
         OR error_message ILIKE '%TRUNCATED and tolerated%')
), agg AS (
  SELECT p.step_name, p.cap, p.agents, count(c.frac) AS n,
         round(100*(percentile_cont(0.95) WITHIN GROUP (ORDER BY c.frac))::numeric, 1) AS p95,
         round(100*max(c.frac), 1) AS peak,
         count(*) FILTER (WHERE c.was_truncated) AS trunc
  FROM pairs p
  JOIN calls c ON c.step_name = p.step_name AND c.cap = p.cap
  GROUP BY 1, 2, 3
), clock AS (
  -- The SECOND limit. Keyed on step alone and with NO max_tokens requirement: the one
  -- real case in the history (scrape_prices, Apr 2026, peak 600,001ms) recorded no cap.
  -- The 480s floor on 'context canceled' separates clock exhaustion from ordinary
  -- caller cancellation (median 23s).
  SELECT step_name, count(*) AS clock_n, max(latency_ms) AS peak_ms,
         string_agg(DISTINCT agent_type, ',') AS agents
  FROM llm_call_log
  WHERE created_at > now() - interval '90 days'
    AND step_name NOT LIKE 'review_%'
    AND (error_message ILIKE '%context deadline exceeded%'
         OR error_message ILIKE '%Client.Timeout%'
         OR (error_message ILIKE '%context canceled%' AND latency_ms >= 480000))
  GROUP BY 1
), flagged AS (
  SELECT step_name || '@' || cap AS subject,
         CASE WHEN trunc > 0 THEN 'T' WHEN peak >= 95 THEN 'N' ELSE 'P' END AS kind,
         'n=' || n || ', p95 ' || p95 || '%, peak ' || peak || '%, truncated ' || trunc ||
         ' — agents: ' || agents AS detail,
         trunc AS sev
    FROM agg
   WHERE n >= 5 AND (trunc > 0 OR peak >= 95 OR p95 >= 85)
  UNION ALL
  -- No n floor: a clock kill is rare and severe enough to report on its own, and a step
  -- failing 100% on the clock has NO successful calls to build an n from.
  SELECT step_name || '@clock', 'C',
         clock_n || ' call(s) died on the CLOCK (not on the cap), peak ' || peak_ms ||
         'ms — agents: ' || agents,
         1000000 + clock_n
    FROM clock
), fp AS (
  SELECT count(*) AS n_flagged,
         md5(string_agg(subject || ':' || kind, '|' ORDER BY subject, kind)) AS digest,
         string_agg(kind || '  ' || subject || ' — ' || detail, chr(10) ORDER BY sev DESC, subject) AS lines
  FROM flagged
), ins AS (
  INSERT INTO doc_notes
    (subject_type, subject_key, body, categories, source, source_agent, created_by)
  SELECT 'pipeline',
         'fleet-step-token-pressure:' || fp.digest,
         'FLEET STEP TOKEN PRESSURE — ' || fp.n_flagged || ' finding(s) (non-review steps; council seats are council-seat-token-pressure''s).' || chr(10) || chr(10) ||
         fp.lines || chr(10) || chr(10) ||
         'C = died on the CLOCK.  T = has truncated at this cap.  N = near-miss, peak >= 95% of cap.  P = pressure, p95 >= 85% of cap.' || chr(10) || chr(10) ||
         'C AND T WANT OPPOSITE FIXES — read the kind before acting. T means the cap was REACHED: raise it, or make the unit of work smaller. C means the cap could NOT be reached, because the request died on wall-clock first: the chassis does not stream and every provider call is one blocking HTTP request (anthropic/gemini 600s, ollama 120s). Raising a cap in response to a C is actively wrong — the bigger number is unreachable too. The levers for C are streaming, a smaller unit of work, or a faster model.' || chr(10) || chr(10) ||
         'THE CLOCK CONVERTS TO A TOKEN CEILING. Output runs ~98 tok/s on Sonnet 5 and 47-82 on Sonnet 4.6, so 600s is roughly 58,000 tokens on Sonnet 5 and 28,000-42,000 on Sonnet 4.6. A max_tokens above that cannot be reached whatever the config says.' || chr(10) || chr(10) ||
         'BEFORE ACTING ON A T, CHECK THE SHAPE, NOT THE COUNT: SELECT md5(prompt_rendered), count(*), count(*) FILTER (WHERE success) FROM llm_call_log WHERE step_name=''X'' AND created_at > now()-interval ''7 days'' GROUP BY 1 ORDER BY 2 DESC. Many distinct prompts near the cap is genuine cap drift. One or two prompts repeating is a STUCK ITEM on a retry loop, where a bigger cap treats the symptom (bugs_open/205).' || chr(10) || chr(10) ||
         'Runbook: docs/agent_docs/docs024_key_docs_latest/bugfix_183_step_token_pressure/RUNBOOK_step_token_pressure.md' || chr(10) ||
         'Thresholds and error vocabulary: SELECT pre_query FROM scheduled_tasks WHERE name=''fleet-step-token-pressure'';',
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
$PQ$,
description = 'bugs_open/183 fix candidate 4, generalised from FIX-058. CTE-only: flags any non-review LLM step whose output distribution approaches its token cap (T/N/P), or that died on wall-clock rather than on its cap (C). C and T want opposite fixes. Writes a doc_notes row when the flagged set changes.',
updated_at = now()
WHERE name = 'fleet-step-token-pressure';

-- Guard: the amendment must be present afterwards. A verify block of bare SELECTs
-- cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty result) — so RAISE.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM scheduled_tasks
                  WHERE name='fleet-step-token-pressure'
                    AND pre_query ILIKE '%died on the CLOCK%'
                    AND pre_query ILIKE '%TRUNCATED and tolerated%') THEN
    RAISE EXCEPTION 'ABORT: amendment not present after UPDATE.';
  END IF;
END $$;

COMMIT;
