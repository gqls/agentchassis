-- ============================================================================
-- 245_llm_call_log_thinking_telemetry.sql
--
-- bugs_open/110 candidate 2: persist what the Gemini client already computes,
-- so thinking cost is answerable from a query instead of only from a log line.
--
-- THE DEFECT. platform/aiservice/gemini.go writes four values into the options
-- map — the wire ceiling actually sent, the reserve added for thinking, the
-- thinking tokens spent, and the provider's total. Verified by grep, NONE has a
-- reader outside platform/aiservice/: they reach no column and no query. That is
-- 016b §9 "a field is only as live as its LAST reader". Thinking BILLS AS OUTPUT
-- and page-content-writer runs once per section across the whole estate, so the
-- one number that drives the cost decision is currently invisible. bugs_open/107,
-- its council submission (corr a1a5cf20) and three commit messages all claimed
-- this data was "visible to logging"; it was written, never persisted.
--
-- WHY THESE FOUR COLUMNS AND NOT THE FOUR THE BUG FILE NAMES
--   110 lists `visible_budget_tokens` as one of the four. That value is ALREADY
--   in llm_call_log.max_tokens — candidate 1 made __sent_max_tokens the sole feed
--   for it, precisely so the column means the caller's answer-budget for every
--   provider. Adding visible_budget_tokens would give ONE meaning TWO column
--   names, which is the defect 110 exists to close, reproduced a third time. The
--   genuinely unpersisted sent-side value is the WIRE ceiling, so that is what
--   goes in. Deviation from the filed candidate is deliberate and recorded here.
--
--   max_tokens              = what the CALLER asked for as answer  (all providers)
--   wire_max_output_tokens  = what we actually SENT the provider   (gemini only)
--   thinking_reserve_tokens = the difference, i.e. headroom for thinking
--   thinking_tokens         = thinking actually spent, billed as output
--   total_output_tokens     = the provider's own total
--
-- WHY NULLABLE, AND WHY THE GO SIDE MUST NOT USE nullIfZero()
--   NULL here means "this provider does not report it" — anthropic and ollama
--   produce none of these, and that is a real distinction from "reported, and it
--   was zero". llm_call_logger.go's existing nullIfZero() helper maps 0 -> NULL,
--   which would collapse a genuine "no thinking on this call" into "no thinking
--   data at all". That is the same empty-vs-absent confusion migration 243's
--   header warns about, one table over, so the four new params are passed as
--   interface{} and set to nil ONLY when the option key was absent. A Gemini row
--   may legitimately read thinking_tokens = 0.
--
-- WHY NO INDEX
--   These are aggregation targets (sum/avg over a provider+date range), not
--   selective predicates, and llm_call_log already carries five indexes on a
--   high-write table. Add one when a real query is shown to need it.
--
-- LIVE THE MOMENT IT APPLIES; the Go that populates it is INERT until the chassis
-- image rolls. Until then every new column is NULL on every row, which is honest:
-- the data genuinely was not captured. No backfill is possible — the values were
-- never persisted anywhere to backfill FROM.
--
-- Verify (bugs_open/110 §"How to verify" item 3) after the roll and one Gemini call:
--   SELECT provider, max_tokens, wire_max_output_tokens, thinking_reserve_tokens,
--          thinking_tokens, output_tokens, total_output_tokens
--   FROM llm_call_log WHERE provider = 'gemini' ORDER BY created_at DESC LIMIT 5;
--   -- expect for page-content-writer: max_tokens 8000, wire 16192, reserve 8192,
--   -- thinking in the 2,764-2,878 range measured for the writer's real prompt.
-- ============================================================================

BEGIN;

ALTER TABLE llm_call_log
    ADD COLUMN IF NOT EXISTS wire_max_output_tokens  integer,
    ADD COLUMN IF NOT EXISTS thinking_reserve_tokens integer,
    ADD COLUMN IF NOT EXISTS thinking_tokens         integer,
    ADD COLUMN IF NOT EXISTS total_output_tokens     integer;

COMMENT ON COLUMN llm_call_log.wire_max_output_tokens IS
    'Output ceiling actually sent to the provider. For thinking models this is the '
    'reserve-inflated total, which is NOT what the caller asked for as answer text — '
    'that is max_tokens. NULL when the provider does not report it (anthropic, ollama). '
    'bugs_open/110.';

COMMENT ON COLUMN llm_call_log.thinking_reserve_tokens IS
    'Headroom added on top of the caller''s visible budget so thinking cannot starve '
    'the answer (wire_max_output_tokens - max_tokens). bugs_open/107 exists because '
    'this was zero: Gemini spends thinking from the SAME ceiling as the answer, so an '
    'Anthropic-sized number passed straight through returned no text at all.';

COMMENT ON COLUMN llm_call_log.thinking_tokens IS
    'Thinking tokens spent, from the provider''s usage metadata. BILLED AS OUTPUT but '
    'not answer, so deliberately not folded into output_tokens. This is where '
    'essentially all of Gemini''s cost lives. 0 is a real value; NULL means the '
    'provider did not report it.';

COMMENT ON COLUMN llm_call_log.total_output_tokens IS
    'The provider''s own total token count for the call. Kept alongside rather than '
    'derived, because the arithmetic between visible, thinking and total is exactly '
    'what bugs_open/110 shows we cannot safely assume holds across providers.';

COMMIT;
