-- 376_llm_call_log_cache_token_columns.sql
--
-- Adds the two Anthropic prompt-caching counters to llm_call_log.
--
-- WHY THIS IS NOT OPTIONAL BOOKKEEPING. Once any agent uses the
-- CacheBreakpointMarker seam (platform/aiservice/anthropic.go), the API's
-- `input_tokens` stops meaning "prompt size" and starts meaning "the UNCACHED
-- REMAINDER". A cached council seat will report ~5k where the real prompt was
-- ~100k. Every existing cost query in this estate reads input_tokens and would
-- silently understate by ~95% — reporting a saving far larger than the real
-- one, in the direction nobody checks. The true prompt size becomes
--   input_tokens + cache_creation_input_tokens + cache_read_input_tokens.
--
-- ORDERING: this migration must be applied BEFORE the chassis image carrying
-- the cache seam is rolled. DB config is live immediately; Go is inert until
-- rebuilt and rolled — so data-before-code is the safe order here, and the
-- reverse would have the binary writing to columns that do not exist.
--
-- Both columns are nullable with no default deliberately: NULL means "this row
-- was written by a binary that predates cache support", which is a different
-- and more useful fact than 0 ("this call used no cache"). Backfilling them to
-- 0 would destroy exactly the distinction an operator needs when asking why a
-- number moved.

BEGIN;

ALTER TABLE llm_call_log
    ADD COLUMN IF NOT EXISTS cache_creation_input_tokens integer,
    ADD COLUMN IF NOT EXISTS cache_read_input_tokens     integer;

COMMENT ON COLUMN llm_call_log.cache_creation_input_tokens IS
    'Anthropic prompt-cache WRITE tokens for this call (billed ~1.25x at 5m TTL, ~2x at 1h). NULL = binary predates cache support; 0 = no cache write.';
COMMENT ON COLUMN llm_call_log.cache_read_input_tokens IS
    'Anthropic prompt-cache READ tokens for this call (billed ~0.1x). NULL = binary predates cache support; 0 = no cache hit. True prompt size = input_tokens + cache_creation + cache_read.';

-- Verify with DO/RAISE, not SELECT. A migration''s verification block made of
-- SELECTs CANNOT stop the COMMIT: ON_ERROR_STOP does not treat a non-empty
-- (or empty) result set as an error, so a "verification" that returns the
-- wrong answer still commits and still looks green. Only a raised exception
-- aborts the transaction.
DO $$
DECLARE
    n integer;
BEGIN
    SELECT count(*) INTO n
    FROM information_schema.columns
    WHERE table_name = 'llm_call_log'
      AND column_name IN ('cache_creation_input_tokens', 'cache_read_input_tokens');

    IF n <> 2 THEN
        RAISE EXCEPTION
            'MIGRATION 376 FAILED: expected 2 cache columns on llm_call_log, found %. Transaction aborted.', n;
    END IF;

    RAISE NOTICE 'migration 376 OK: % cache columns present on llm_call_log', n;
END $$;

COMMIT;
