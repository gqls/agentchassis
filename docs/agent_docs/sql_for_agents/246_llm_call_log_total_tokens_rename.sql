-- ============================================================================
-- 246_llm_call_log_total_tokens_rename.sql
--
-- Phase 1 of 2: ADD llm_call_log.total_tokens alongside total_output_tokens.
-- Phase 2 (drop the old column) is a LATER migration, after the chassis rolls.
--
-- I REPRODUCED bugs_open/110's OWN DEFECT INSIDE THE FIX FOR IT. 110 exists because
-- one column (max_tokens) carried two different meanings depending on provider. In
-- closing it I added `total_output_tokens` and fed it Gemini's
-- `usageMetadata.totalTokenCount` — the total for the WHOLE CALL, prompt included, not
-- a total of output. The name asserts a meaning the value does not have: the same class
-- of error, one column over.
--
-- Caught by READING THE FIRST FOUR LIVE ROWS rather than trusting the name I chose. In
-- every one the column equals input + visible + thinking exactly:
--
--   prompt | visible | thinking | total_output_tokens | input+visible+thinking
--     5628 |     446 |     2619 |                8693 |  8693
--     4160 |      87 |     1638 |                5885 |  5885
--     5238 |      63 |     1849 |                7150 |  7150
--     4227 |     103 |     1582 |                5912 |  5912
--
-- Not coincidence: gemini.go:411 assigns `response.UsageMetadata.TotalTokenCount`, and
-- Gemini's totalTokenCount is promptTokenCount + candidatesTokenCount + thoughtsTokenCount.
--
-- WHY ADD-THEN-DROP AND NOT `RENAME COLUMN`. A rename is atomic in the DB and therefore
-- NOT atomic with the fleet. llm_call_log has ONE writer — the INSERT in
-- llm_call_logger.go — and it names its columns explicitly, so:
--   rename first  -> the RUNNING binary writes total_output_tokens, which no longer
--                    exists -> every INSERT fails -> llm_call_log stops recording for
--                    EVERY provider until the next image roll.
--   Go rolls first -> the new binary writes total_tokens, which does not exist yet ->
--                    same outage, mirrored.
-- There is no ordering of a rename that avoids the window; the window is the defect.
-- Both columns existing at once removes it: the old binary keeps writing the old
-- column, the new binary writes the new one, and neither can fail. Logging is
-- fire-and-forget (errors logged, never blocking), so the outage would have been silent
-- — which is exactly why it is worth two migrations instead of one.
--
-- WHY NOW AND NOT LATER. The column is hours old, holds four rows and has no reader
-- anywhere. The first person to write `sum(output_tokens + total_output_tokens)`, or to
-- compare it against wire_max_output_tokens, gets a wrong answer that looks right. A
-- misleading name is cheapest to fix before anything depends on it — 016b §9's "order
-- fix candidates by what closes the door": a rename makes the misreading
-- unrepresentable, a comment relies on someone reading it.
--
-- NOT RENAMED, deliberately: `output_tokens` keeps meaning VISIBLE text on every
-- provider, and `thinking_tokens` stays separate. Thinking bills as output but is not
-- answer; folding them would destroy the one distinction that makes the cost question
-- answerable, which is the entire point of 110 candidate 2.
--
-- Verify:
--   SELECT input_tokens, output_tokens, thinking_tokens, total_tokens,
--          total_tokens - (input_tokens + output_tokens + thinking_tokens) AS should_be_0
--   FROM llm_call_log WHERE total_tokens IS NOT NULL ORDER BY created_at DESC LIMIT 5;
--
-- PHASE 2, only once a pod-grep shows the new binary live (the INSERT must name
-- total_tokens, not total_output_tokens):
--   ALTER TABLE llm_call_log DROP COLUMN total_output_tokens;
-- ============================================================================

BEGIN;

ALTER TABLE llm_call_log
    ADD COLUMN IF NOT EXISTS total_tokens integer;

-- Carry the four existing rows across so no data is stranded on the old column.
UPDATE llm_call_log
   SET total_tokens = total_output_tokens
 WHERE total_output_tokens IS NOT NULL
   AND total_tokens IS NULL;

COMMENT ON COLUMN llm_call_log.total_tokens IS
    'The provider''s own total token count for the CALL — prompt + visible output + '
    'thinking. NOT a total of output: it includes input_tokens, which is why it '
    'supersedes total_output_tokens (migration 246). Kept alongside the parts rather '
    'than derived, because bugs_open/110 is precisely about not assuming the arithmetic '
    'between token counts holds across providers. NULL when the provider does not '
    'report it (anthropic, ollama).';

COMMENT ON COLUMN llm_call_log.total_output_tokens IS
    'SUPERSEDED by total_tokens (migration 246) — the name was wrong: it holds the '
    'whole-call total INCLUDING the prompt, not a total of output. Retained only so the '
    'pre-roll binary can keep writing it without failing every INSERT. DROP once a '
    'pod-grep shows the chassis INSERT naming total_tokens.';

COMMIT;
