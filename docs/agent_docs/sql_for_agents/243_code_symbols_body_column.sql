-- ============================================================================
-- 243_code_symbols_body_column.sql
--
-- D11 layer 1 of the architecture-seat workstream: give the code index the one
-- thing every seat assumed it already had — SYMBOL BODIES.
--
-- Council gate APPROVED this plan on round 3, correlation
-- 18fe4035-4fa6-4079-ab44-8541d6e58944 (2026-07-27 19:49:36, "approved with 4
-- advisory objection(s) — none high-severity"). The submission is committed at
-- docs/agent_docs/docs024_key_docs_latest/architecture_review/SUBMISSION_2026-07-27_code_tier_lookup.json
-- and is deliberately NOT edited, so the artifact keeps matching the verdict.
--
-- THE DEFECT (bugs_open/108 defect B). diagnose_code_lookup's `content` kind
-- documents itself as matching "symbol source bodies", but code_symbols.content
-- is composed by composeSymbolContent() from kind + symbol + signature + doc +
-- path — DECLARATION text only. Bodies are read on demand by
-- internal/analysis.ReadSymbolBody, which the indexer never calls. So every
-- reviewer content check for a route, a registry key, a table name, a config key
-- or any string literal inside a function returns ZERO ROWS — and a zero read as
-- absence is how a council concludes "no such code exists" about code that does.
--
-- WHY A SEPARATE COLUMN AND NOT AN EXTENSION OF `content`
--   content does THREE jobs at once: the trigram search text, the text that is
--   EMBEDDED, and the input to content_hash — which is the re-embed trigger
--   (loadExistingHashes skips embedding when the hash is unchanged). Appending
--   bodies to it would silently re-embed all 4,535 rows and skew every vector
--   in the index, as an invisible side effect of a search fix.
--
-- WHY NOT `CONCURRENTLY`
--   scripts/migration/run-migrations.sh:129-139 dry-runs each file inside a
--   transaction it has poisoned so it cannot commit. CREATE INDEX CONCURRENTLY
--   cannot run inside a transaction block, so the safe-looking option is the one
--   that breaks the safety check. At 4,535 rows a plain build is milliseconds.
--
-- WHY `body` STAYS NULLABLE
--   NULL means "not indexed yet" (this migration is live the moment it applies;
--   the Go that populates it is inert until the image rolls). '' would make that
--   indistinguishable from a genuinely empty body — the exact empty-vs-absent
--   confusion bugs_open/108 is about, reintroduced one layer down.
--
-- NUMBERING. The approved plan said 241. Between submission and approval another
-- session committed 241_page_writer_uses_voice_placeholder.sql, and a third has
-- 242_backfill_103_tool_meta_descriptions.sql staged. guardian's low objection —
-- that "241 is the next free number" was asserted, not checkable from SQL — was
-- vindicated twice inside two hours. This is a shared sequence; re-read it
-- immediately before writing a file:
--   ls docs/agent_docs/sql_for_agents/*.sql | xargs -n1 basename \
--     | grep -oE '^[0-9]+' | sort -n | tail -1
--
-- VERIFY: 243_code_symbols_body_column_VERIFY.sql — run BEFORE and AFTER the
-- reindex. The invariant it asserts (content and content_hash byte-identical
-- across this work) is the one that matters.
-- ============================================================================

BEGIN;

-- Fail fast rather than queue behind a running reindex: the single writer is the
-- index_code_symbols action, which runs on a 24h cadence, not continuously.
SET LOCAL lock_timeout = '5s';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name   = 'code_symbols'
          AND column_name  = 'body'
    ) THEN
        ALTER TABLE code_symbols ADD COLUMN body text;
        RAISE NOTICE '243: added code_symbols.body';
    ELSE
        RAISE NOTICE '243: code_symbols.body already present — nothing to add';
    END IF;
END $$;

COMMENT ON COLUMN code_symbols.body IS
    'Source text of the symbol, lines [line_start,line_end] inclusive, sliced by '
    'internal/analysis.SliceLines at index time from the local checkout '
    'analyse_repo_local leaves behind. NULL = not indexed (never ""). '
    'DELIBERATELY NOT part of content/content_hash: content is the embedded text '
    'and the hash is the re-embed trigger, so folding bodies in would re-embed '
    'every row. Searched by diagnose_code_lookup kind=content.';

CREATE INDEX IF NOT EXISTS idx_code_symbols_body_trgm
    ON code_symbols USING gin (body gin_trgm_ops);

COMMIT;
