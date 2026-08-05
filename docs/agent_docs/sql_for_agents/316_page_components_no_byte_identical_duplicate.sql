-- 316_page_components_no_byte_identical_duplicate.sql
--
-- bugs_closed/156 § "What remains", item 2 — the DB-level invariant.
-- OWNER-DIRECTED, 2026-08-05, answering the council's open MEDIUM objection on
-- correlation 1a3f4f27 (`bug_historian` seat):
--
--   "the guard lands only in SavePageSectionsAction. page_components still has no
--    DB-level invariant … Six other Go call sites insert into page_components;
--    today they write single rows so this specific bug can't recur through them,
--    BUT THAT IS A FACT ABOUT CURRENT CALLERS, NOT AN ENFORCED MECHANISM."
--
-- The Go guard (register PBP-033, live v1.0.1252) collapses byte-identical
-- sections at ONE of seven writers. This makes the invariant structural: no page
-- may hold two non-removed rows that are byte-identical in every column the
-- INSERT binds.
--
-- ─────────────────────────────────────────────────────────────────────────────
-- THE SHAPE, AND WHY EACH PIECE IS THERE. Two intuitive versions are WRONG and
-- the database says so itself — all three tested against production and rolled
-- back, 2026-08-05.
--
--   A. (page_id, slot_name, md5(content_data)) NULLS NOT DISTINCT
--      ** REFUSED BY POSTGRES AGAINST LIVE DATA: **
--        ERROR: could not create unique index
--        DETAIL: Key (…)=(1bcc0213-…, generic-text-block, null) is duplicated.
--      That is finetuning.uk/our-position-on-ai, and it is a LEGITIMATE pair:
--      both rows have NULL content_data but DIFFERENT markup (1,812 vs 1,790
--      bytes, distinct md5s). The database refuses the constraint, naming the
--      rows it would have destroyed. Same correction the Go guard needed.
--
--   B. + md5(rendered_html), but WITHOUT component_id
--      Builds today (0 violations) and is a LATENT TRAP: it is STRICTER than the
--      Go guard, which keeps rows differing only by component_id. A constraint
--      stricter than the guard is the worst combination — the guard says "save
--      both", the database then rejects one, and the disagreement surfaces as a
--      dropped section nobody asked for. Measured: 0 such groups today, so this
--      is about what CAN happen, not what has.
--
--   C. THIS ONE. + component_id, so the index is EXACTLY as permissive as
--      collapseDuplicateSections. The two can never disagree by construction.
--
-- WHY `NULLS NOT DISTINCT` (PostgreSQL 15+; this cluster is 15.18): without it,
-- NULLs are distinct, so the index is BLIND to all 132 rows with NULL
-- content_data — a silent hole exactly where the trap lives (see variant A).
--
-- WHY PARTIAL ON build_status <> 'removed': a soft-removed row is not on the
-- page, so it must not block re-adding the same content later. 1 such row today.
-- The property preserved is the one that matters: no two NON-REMOVED rows on a
-- page are byte-identical.
--
-- ─────────────────────────────────────────────────────────────────────────────
-- BLAST RADIUS — all seven writers read, not assumed. NONE can violate it today.
--
--   save_page_sections        DELETE-then-INSERT, and PBP-033 collapses
--                             duplicates before the INSERT — cannot reach it.
--   deploy_tool               ON CONFLICT DO NOTHING → becomes a true no-op.
--   create_tool_component     inserts onto a page it CREATED in the same action
--                             (INSERT INTO pages :295) — page_id is fresh.
--   create_report_page        checks for an existing component first (existingPC)
--                             and updates instead of inserting.
--   adopt_verbatim            SELECT id … WHERE page_id AND slot_name → UPDATE if
--                             found, INSERT only into an empty slot.
--   rebuild_blog_listing      same check-then-update-else-insert pattern.
--   cmd/webdesignport/import  offline import CLI, not a runtime path.
--
-- CORRECTION TO AN EARLIER CLAIM, recorded because it changes the risk profile
-- this change was approved on: I previously said a violation would make "the save
-- fail, the build fail, and the page not deploy". THAT IS WRONG FOR
-- save_page_sections. Its INSERT error handler (save_page_sections_action.go:837)
-- logs a Warn and `continue`s — the row is skipped and the save proceeds with a
-- lower savedCount. So even the worst path degrades to a SKIPPED SECTION, not a
-- failed build. Of the rest, only create_tool_component returns an error, and it
-- cannot collide (fresh page). The true risk is therefore materially smaller than
-- stated when this was agreed.
--
-- ROLLBACK — one command, no data change, takes no long lock:
--   DROP INDEX CONCURRENTLY IF EXISTS uq_page_components_no_byte_identical_duplicate;
--
-- MONITORING — a violation is otherwise a Warn line that expires. This finds
-- sections silently skipped by a constraint hit:
--   SELECT * FROM agent_error_log
--    WHERE error_message ILIKE '%uq_page_components_no_byte_identical%'
--       OR context::text ILIKE '%uq_page_components_no_byte_identical%'
--    ORDER BY occurred_at DESC;
--   -- and in the pod logs: grep 'Failed to insert section'
--
-- COST: 184 kB index on a 1,212-row / 6,760 kB table. Build is sub-second; the
-- ACCESS EXCLUSIVE lock is taken with lock_timeout so it fails fast rather than
-- queueing behind a live save.
--
-- APPLY:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 316_….sql

SET statement_timeout = '60s';
SET lock_timeout = '5s';

BEGIN;

CREATE UNIQUE INDEX uq_page_components_no_byte_identical_duplicate
    ON page_components
       (page_id, slot_name, md5(content_data::text), md5(rendered_html), component_id)
    NULLS NOT DISTINCT
    WHERE build_status <> 'removed';

COMMENT ON INDEX uq_page_components_no_byte_identical_duplicate IS
    'bugs_closed/156 / PBP-033. No page may hold two non-removed rows identical in every column save_page_sections binds (position excluded). EXACTLY as permissive as collapseDuplicateSections in save_sections_dedup.go — component_id is in the key so the two can never disagree; md5(rendered_html) is in it because NULL content_data is legitimate and common (132 rows). Changing either half without the other reintroduces the disagreement.';

-- VERIFY as DO/RAISE, never a block of SELECTs: ON_ERROR_STOP does not fire on a
-- non-empty result set, so a SELECT-based verify block cannot stop the COMMIT.
DO $$
DECLARE n int; ok boolean;
BEGIN
    SELECT count(*) INTO n FROM pg_indexes
     WHERE tablename='page_components'
       AND indexname='uq_page_components_no_byte_identical_duplicate';
    IF n <> 1 THEN RAISE EXCEPTION '316: the index does not exist after CREATE'; END IF;

    SELECT indisunique INTO ok FROM pg_index i
      JOIN pg_class c ON c.oid=i.indexrelid
     WHERE c.relname='uq_page_components_no_byte_identical_duplicate';
    IF NOT ok THEN RAISE EXCEPTION '316: the index exists but is NOT unique — it would enforce nothing'; END IF;

    -- The whole point is that it REJECTS a byte-identical row. Prove it here,
    -- inside the transaction, rather than asserting it: a unique index that
    -- silently fails to fire is indistinguishable from one that works until the
    -- day it matters.
    BEGIN
        INSERT INTO page_components
            (page_id, position, rendered_html, slot_name, component_id,
             content_data, content_brief, build_status)
        SELECT page_id, 9999, rendered_html, slot_name, component_id,
               content_data, content_brief, build_status
          FROM page_components
         WHERE build_status <> 'removed'
         ORDER BY id LIMIT 1;
        RAISE EXCEPTION '316: a byte-identical row was ACCEPTED — the index is not enforcing';
    EXCEPTION WHEN unique_violation THEN
        RAISE NOTICE '316: confirmed — a byte-identical row is rejected by the index';
    END;

    -- And the other direction: the 11 legitimate same-slot/different-content
    -- groups must still exist untouched. If this number moved, the index
    -- destroyed something it was written not to touch.
    SELECT count(*) INTO n FROM (
        SELECT page_id, slot_name FROM page_components
         GROUP BY 1,2 HAVING count(*) > 1) g;
    IF n < 11 THEN
        RAISE EXCEPTION '316: only % duplicate-slot groups remain, expected >= 11 legitimate ones', n;
    END IF;
    RAISE NOTICE '316: % legitimate repeated-slot groups still present', n;
END $$;

COMMIT;
