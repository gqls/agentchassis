-- ============================================================================
-- 262_experience_council_deferral_seat_budget.sql
--
-- The deferral_honesty seat truncated at 6000 (migration 261 raised it) and then
-- truncated AGAIN at 8000. Two rounds, same seat, same failure — so the budget
-- was not the problem, my prompt was.
--
-- That seat is asked to do four things, each per-check, against an entry with
-- five checks plus fifteen binding-level deferrals, AND it receives the register
-- summary for context. Nothing in the prompt bounds how much it should write, so
-- it wrote until it ran out and never closed its JSON. Raising the ceiling again
-- would postpone the same failure rather than fix it: an unbounded "say
-- everything" ask finds a ceiling eventually, whatever it is set to.
--
-- So this does both, and the second part is the real fix:
--   * 8000 -> 12000, matching the seat's genuinely larger input;
--   * an explicit OUTPUT BUDGET in the prompt — a cap on objections and on the
--     length of each, and an instruction to spend the budget on the most
--     important findings rather than the first ones.
--
-- Worth noting what worked: nothing here was needed to make the failure VISIBLE.
-- The truncation guard refused to pass the fragment, `diagnose_council_decide`
-- counted the seat as `unreadable` rather than `abstained`, and the report's
-- metadata carried `unreadable: 1` where anyone reading the verdict would see
-- it. A lost opinion could not masquerade as a considered non-objection. That is
-- three independent mechanisms doing exactly their job.
-- ============================================================================

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(default_config,
        '{workflow,steps,review_deferral_honesty,config,ai_service,max_tokens}',
        '12000'::jsonb),
      '{workflow,steps,review_deferral_honesty,config,prompt_template}',
      to_jsonb(
        (default_config->'workflow'->'steps'->'review_deferral_honesty'->'config'->>'prompt_template')
        || E'\n\n## OUTPUT BUDGET — this is part of the task, not a formatting note\n\n'
        || E'You have a hard output limit and you have exceeded it twice, losing your entire opinion both times: a truncated reply is discarded, so an unfinished thorough answer is worth LESS than a finished selective one.\n\n'
        || E'So: at most SIX objections. One or two sentences each — name the check or clause, state what is wrong, stop. No preamble, no restating the entry back, no summarising your own findings at the end. Keep `notes` under 120 words.\n\n'
        || E'Spend the budget on the most important findings, not the first ones you noticed. If an entry has more problems than fit, say so in `notes` in one sentence and list the worst six — an explicit "there are more" is information; running out of room silently is not.'
      )),
    updated_at = now()
WHERE type = 'experience-approval-council'
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

DO $guard$
DECLARE
    tokens int; prompt text;
BEGIN
    SELECT (default_config->'workflow'->'steps'->'review_deferral_honesty'->'config'->'ai_service'->>'max_tokens')::int,
           default_config->'workflow'->'steps'->'review_deferral_honesty'->'config'->>'prompt_template'
      INTO tokens, prompt
      FROM agent_definitions
     WHERE type = 'experience-approval-council'
       AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    IF tokens <> 12000 THEN
        RAISE EXCEPTION '262: deferral seat max_tokens is %, expected 12000', tokens;
    END IF;
    IF prompt NOT LIKE '%OUTPUT BUDGET%' THEN
        RAISE EXCEPTION '262: the output budget was not appended — raising the ceiling alone just postpones the same truncation';
    END IF;
    -- The prompt must still contain its original remit; an append that replaced
    -- the body would be worse than the truncation it fixes.
    IF prompt NOT LIKE '%DEFERRAL HONESTY%' THEN
        RAISE EXCEPTION '262: the original remit is gone — the append clobbered the prompt';
    END IF;
END
$guard$;

COMMIT;

-- Verify
SELECT k AS seat,
       v->'config'->'ai_service'->>'max_tokens' AS max_tokens,
       length(v->'config'->>'prompt_template') AS prompt_chars,
       (v->'config'->>'prompt_template') LIKE '%OUTPUT BUDGET%' AS has_budget
FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') AS e(k, v)
WHERE a.type = 'experience-approval-council'
  AND COALESCE(a.is_snapshot,false) = false AND a.deleted_at IS NULL
  AND k LIKE 'review_%'
ORDER BY k;
