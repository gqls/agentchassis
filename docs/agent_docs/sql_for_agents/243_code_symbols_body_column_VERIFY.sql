-- ============================================================================
-- 243_code_symbols_body_column_VERIFY.sql
--
-- Hand-run companion to 243. UPPERCASE-suffixed, so run-migrations.sh's
-- SIDECAR_RE recognises it as NOT a migration and never auto-applies it
-- (run-migrations.sh:65). Re-runnable; it only SELECTs.
--
-- RUN IT TWICE: once BEFORE the reindex that populates bodies, once AFTER.
-- Checks 1, 2 and 4 must return 0 both times. Check 3 is the only one whose
-- answer is allowed to change, and changing is the point.
--
-- WHY THIS EXISTS AS CODE RATHER THAN AS A REQUEST TO A REVIEWER. The plan's
-- most dangerous failure mode is silent: if the body write disturbs content or
-- content_hash, every one of the 4,535 rows re-embeds on the next refresh and
-- the vectors shift, with no error anywhere. Round 2 of the council guarded that
-- by asking a reviewer to check. A reviewer who forgets looks exactly like a
-- reviewer who checked and found nothing — which is the same defect, one floor
-- up, that this whole migration is fixing.
--
-- THE HASH EXPRESSION IS THE CORRECTED ONE. The approved plan wrote
-- sha256(content::bytea), which ERRORS with "invalid input syntax for type
-- bytea": that cast PARSES the text as a bytea literal, it does not encode it.
-- convert_to(content,'UTF8') is the encode. editquality objected (medium) that
-- the formula was asserted rather than run; it was right, and the defect was a
-- cast. Verified live 2026-07-27: 4,535 of 4,535 rows match, 0 mismatches.
--
-- Run:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db \
--     -f - < docs/agent_docs/sql_for_agents/243_code_symbols_body_column_VERIFY.sql
-- ============================================================================

\echo '=== 1. content_hash drift — MUST be 0, before and after ==='
-- If this is ever non-zero, something OTHER than the indexer has rewritten
-- content or content_hash and the "no re-embed" claim is void.
SELECT count(*) AS content_hash_drift
FROM code_symbols
WHERE content_hash IS DISTINCT FROM encode(sha256(convert_to(content, 'UTF8')), 'hex');

\echo '=== 2. body folded into the hash input — MUST be 0 ==='
-- Catches the specific mistake of appending body to content before hashing.
SELECT count(*) AS hash_contaminated_by_body
FROM code_symbols
WHERE body IS NOT NULL
  AND content_hash = encode(sha256(convert_to(content || body, 'UTF8')), 'hex');

\echo '=== 3. body coverage — 0 before the roll, ~total after the reindex ==='
-- BEFORE: bodies_populated = 0 and that is correct, not a failure. The column is
-- live the moment 243 applies; the Go that fills it is inert until the image
-- rolls. Every content check degrades to today's declaration-only behaviour in
-- between — it does not break.
SELECT count(*)                                   AS total_rows,
       count(body)                                AS bodies_populated,
       count(*) FILTER (WHERE body = '')          AS empty_string_bodies,  -- MUST be 0: NULL means not-indexed
       round(100.0 * count(body) / NULLIF(count(*), 0), 1) AS pct_with_body
FROM code_symbols;

\echo '=== 4. the fix is real: a body-only string must now be findable ==='
-- "stop_reason" is the example the action's own contract comment has always
-- given and which has never worked (bugs_open/108 defect B). AFTER the reindex
-- this must be > 0. The negative control must be 0 in both runs — without it, a
-- broken predicate that matches everything would read as success.
SELECT count(*) AS stop_reason_hits
FROM code_symbols
WHERE body ILIKE '%stop_reason%';

SELECT count(*) AS negative_control_must_be_zero
FROM code_symbols
WHERE body ILIKE '%zzz_string_that_appears_nowhere_zzz%';

\echo '=== 5. the OR predicate the lookup will run — check the plan, not the row count ==='
-- guardian (low): an OR across two trigram-indexed columns may use NEITHER index
-- and fall back to a seq scan. At 4,535 rows that is still fast, so this is
-- information, not a gate — but if it seq-scans, the alternative is one index on
-- (COALESCE(body,'') || ' ' || content) instead of the OR.
EXPLAIN
SELECT path, symbol
FROM code_symbols
WHERE (COALESCE(body, '') ILIKE '%stop_reason%' OR content ILIKE '%stop_reason%');
