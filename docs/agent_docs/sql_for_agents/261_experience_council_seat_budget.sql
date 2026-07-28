-- ============================================================================
-- 261_experience_council_seat_budget.sql — raise the approval council's seats
-- from 6000 to 8000 output tokens, matching the code council.
--
-- WHY. On the first complete run (correlation 055df47a) the deferral_honesty
-- seat came back:
--   {"type":"text","result":"","__truncated":true,"__truncated_output_tokens":6000}
--
-- output_tokens == max_tokens means the completion was CUT, not finished — the
-- standing rule. The seat wrote until the budget ran out, never closed its JSON,
-- and the truncation guard correctly recorded an EMPTY result rather than
-- passing a fragment off as an opinion (bugs_closed/012's discipline working as
-- designed). The verdict that round was decided by a gating objection from
-- another seat, so it stands — but one seat of five was silently missing from
-- it, and `unreadable: 1` in the report metadata is the only reason that is
-- visible at all.
--
-- 6000 was my number, chosen without checking what the comparable council uses.
-- The code council's seats have run at 8000 for weeks. The two seats that also
-- receive the register summary have the largest prompts and the most to say.
--
-- NOT a fix for the root cause, which is that a reviewer with a lot to say will
-- always find some ceiling. The durable protections are already in place and are
-- what made this legible: the truncation guard refuses to pass a fragment, and
-- diagnose_council_decide counts `unreadable` separately from `abstained` so a
-- lost opinion cannot read as a considered non-objection.
-- ============================================================================

BEGIN;

UPDATE agent_definitions a
SET default_config = (
      SELECT jsonb_set(a.default_config, '{workflow,steps}',
               (SELECT jsonb_object_agg(
                         k,
                         CASE WHEN k LIKE 'review_%'
                              THEN jsonb_set(v, '{config,ai_service,max_tokens}', '8000'::jsonb)
                              ELSE v END)
                  FROM jsonb_each(a.default_config->'workflow'->'steps') AS e(k, v)))
    ),
    updated_at = now()
WHERE a.type = 'experience-approval-council'
  AND COALESCE(a.is_snapshot, false) = false
  AND a.deleted_at IS NULL;

DO $guard$
DECLARE
    low int; seats int;
BEGIN
    SELECT count(*) FILTER (WHERE (v->'config'->'ai_service'->>'max_tokens')::int < 8000),
           count(*)
      INTO low, seats
      FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') AS e(k, v)
     WHERE a.type = 'experience-approval-council'
       AND COALESCE(a.is_snapshot,false) = false AND a.deleted_at IS NULL
       AND k LIKE 'review_%';

    IF seats <> 5 THEN
        RAISE EXCEPTION '261: expected 5 review seats, found %', seats;
    END IF;
    IF low > 0 THEN
        RAISE EXCEPTION '261: % seat(s) still below 8000 output tokens', low;
    END IF;
END
$guard$;

COMMIT;

-- Verify
SELECT k AS seat, v->'config'->'ai_service'->>'max_tokens' AS max_tokens
FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') AS e(k, v)
WHERE a.type = 'experience-approval-council'
  AND COALESCE(a.is_snapshot,false) = false AND a.deleted_at IS NULL
  AND k LIKE 'review_%'
ORDER BY k;
