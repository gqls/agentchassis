-- 217_site_work_items_handler_agent_not_null.sql
--
-- bugs_closed/078 — one work item with `handler_agent IS NULL` silently livelocks
-- the fleet build dispatcher. This is the durable close (fix candidate 3).
--
-- WHY
-- ---
-- `site_work_items.handler_agent` is nullable with no default, so a hand-written
-- INSERT that simply OMITS the column produces a row that stops builds on every
-- site:
--
--   * `find_dispatchable_site` (build-pipeline-trigger config SQL) counts the row
--     — it filters on status/attempt_count and says nothing about handler_agent.
--   * `LoadWorkItemsAction` (platform/orchestration/actions/load_work_item_actions.go:609)
--     scans handler_agent into a plain Go `string`, so SQL NULL fails the scan
--     ("converting NULL to string is unsupported") and the row is dropped by a
--     `continue` behind a Warn (:624).
--
-- The loop therefore returns item_count: 0, claims nothing, the row stays
-- 'triaged', and 120s later the trigger picks THE SAME SITE again. Forever.
-- Because selection is `ORDER BY wi.site_id ... LIMIT 1`, a NULL-handler row on a
-- low site_id starves every site sorting above it. Not degradation — a total
-- stop, while dispatch orchestrations COMPLETE briskly having done nothing.
--
-- MEASURED TWICE, BY TWO UNRELATED SESSIONS, IN 24 HOURS
-- ------------------------------------------------------
--   2026-07-25  leopardessconsulting.co.uk (lowest site_id), created_by
--               operator:bugfix_023 — no work item completed anywhere on the
--               fleet between 17:42 and at least 18:26.
--   2026-07-26  gaswholesalers.com (lowest site_id), created_by
--               operator:bugfix_049 — 42 minutes of zero completions fleet-wide
--               (last 17:00:32, found 17:42) while the trigger selected that one
--               site on every tick (17:32/17:35/17:37/17:40/17:42) and
--               build-dispatch-loop returned item_count 0 every time.
--               Setting the handler on that single row moved item_count 0 -> 1
--               at 17:45:30 and a completion landed. Nothing else changed.
--
-- Recurrence inside a day is the argument for closing the door in the schema
-- rather than asking operators to remember a column.
--
-- WHY DEFAULT '' AND NOT A REAL AGENT NAME
-- ----------------------------------------
-- "No handler" is a legitimate, deliberate state — flag-only / human-review items
-- carry it on purpose (check_image_url_404.go:109 "HandlerAgent intentionally
-- empty — flag-only"; lock_helpers.go:117 and render_site_components_action.go:693
-- both route to 'needs_human_review' with NO handler_agent). The empty string is
-- ALREADY the canonical spelling of that state: measured 2026-07-26, 169
-- needs_human_review rows carry '' against 121 carrying NULL. So this normalises
-- two spellings of one concept down to the one that is safe.
--
-- '' is safe end-to-end and NULL is not, and the whole outage turns on that
-- asymmetry: '' scans fine, so the item LOADS, gets CLAIMED, and then fails at
-- spawn_agent (agent_type_field: current_item.handler_agent) into its configured
-- error_step: mark_failed. Bounded, item-level, against its own attempt_count —
-- and critically the site is then mutex'd by its own claimed row, so every OTHER
-- site keeps building. NULL is dropped before any of that can happen, which is
-- why one row takes the fleet down. 018_site_work_items.sql:498 already records
-- the '' behaviour ("Empty handler_agent would crash the dispatch loop's
-- spawn_handler step") — a crash on one item was never the problem.
--
-- SAFETY (each checked live 2026-07-26; evidence in bugs_closed/078)
-- -----------------------------------------------------------------
-- * The 129 rows being rewritten are all non-dispatchable: 121
--   'needs_human_review' + 8 'complete'. ZERO are 'triaged'/'approved', so the
--   backfill cannot change what the dispatcher sees.
--   (The one triaged NULL row that caused today's outage was repaired by hand
--   before this file was written; it is gaswholesalers 709f0338, now
--   'page-rerender' — the value 704 of its 705 siblings carry.)
-- * NO WRITER EMITS AN EXPLICIT NULL, so SET NOT NULL cannot break an insert.
--   insertWorkItem (load_work_item_actions.go:1035) passes a Go `string`, which
--   is '' at worst. The three paths that actually created the NULL rows —
--   resolve_internal_links_action.go:257, plan_sections_action.go:1862,
--   reconcile_site_plan_action.go:255 — all OMIT the column from the INSERT
--   column list entirely, so they now pick up the DEFAULT and keep working
--   unchanged. That is the point: the fix lands without touching them.
-- * Nothing reads `handler_agent IS NULL` as a distinguished state. The two
--   places in sql_for_tables/018 that test it already write
--   `(handler_agent IS NULL OR handler_agent = '')`, i.e. they treat the two as
--   the same thing. feasibility-recheck promotes 'blocked' -> 'triaged' only
--   where EXISTS(agent_definitions WHERE type = wi.handler_agent), and '' matches
--   no agent type, so an unroutable row cannot be promoted into a retry loop.
-- * idx_swi_handler is btree(handler_agent, status) and indexes '' fine.
--
-- NOT DONE HERE, DELIBERATELY
-- ---------------------------
-- The case file's fix candidate 2 — adding `handler_agent IS NOT NULL` to
-- find_dispatchable_site — is NOT applied. Another live session owns that query
-- (213_dispatch_gate_matches_dispatcher.sql, bugs_open/029, rewrites it
-- wholesale), and after this file the clause is redundant anyway: the column can
-- no longer be NULL. Editing the same JSONB path underneath a concurrent thread
-- is the collision CLAUDE.md warns about.
--
-- ROLLBACK
-- --------
--   ALTER TABLE site_work_items
--     ALTER COLUMN handler_agent DROP NOT NULL,
--     ALTER COLUMN handler_agent DROP DEFAULT;
--   -- The backfill is deliberately NOT reversed: '' and NULL were already two
--   -- spellings of one state, and restoring NULLs would reinstate the defect.

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Backfill — collapse the two spellings of "no handler" onto the safe one.
-- ---------------------------------------------------------------------------
UPDATE site_work_items
   SET handler_agent = ''
 WHERE handler_agent IS NULL;

-- ---------------------------------------------------------------------------
-- 2. Shut the door. DEFAULT covers the omitted-column INSERT (the mistake that
--    caused both outages); NOT NULL covers the explicitly-NULL INSERT, which now
--    fails loudly at insert time against its own author instead of silently
--    against the whole fleet.
-- ---------------------------------------------------------------------------
ALTER TABLE site_work_items
  ALTER COLUMN handler_agent SET DEFAULT '',
  ALTER COLUMN handler_agent SET NOT NULL;

-- ---------------------------------------------------------------------------
-- Guards — assert the exact post-conditions inside the transaction, so any
-- failure rolls the whole file back.
--
-- The load-bearing guard is the PROBE: it re-runs the exact INSERT shape that
-- caused both outages (a work item whose column list omits handler_agent) and
-- asserts the row comes back '' rather than NULL. Asserting the catalogue flags
-- alone would only prove the DDL parsed; this proves the defect is extinct.
-- The probe is inserted and deleted inside this transaction, so it is never
-- visible to another session and never survives the file. It is created
-- 'detected' — non-dispatchable — so that even an impossible leak could not
-- reach the dispatcher.
-- ---------------------------------------------------------------------------
DO $guard$
DECLARE
    v_notnull  boolean;
    v_default  text;
    v_nulls    bigint;
    v_site     uuid;
    v_probe    uuid;
    v_handler  text;
    v_isnull   boolean;
BEGIN
    SELECT attnotnull INTO v_notnull
      FROM pg_attribute
     WHERE attrelid = 'site_work_items'::regclass
       AND attname  = 'handler_agent'
       AND NOT attisdropped;

    IF v_notnull IS NULL THEN
        RAISE EXCEPTION '217: site_work_items.handler_agent does not exist';
    END IF;
    IF NOT v_notnull THEN
        RAISE EXCEPTION '217: handler_agent is still nullable';
    END IF;

    SELECT pg_get_expr(adbin, adrelid) INTO v_default
      FROM pg_attrdef
     WHERE adrelid = 'site_work_items'::regclass
       AND adnum = (SELECT attnum FROM pg_attribute
                     WHERE attrelid = 'site_work_items'::regclass
                       AND attname = 'handler_agent');

    IF v_default IS NULL OR v_default NOT LIKE '''''%' THEN
        RAISE EXCEPTION '217: handler_agent default is not the empty string (got %)',
                        COALESCE(v_default, '<none>');
    END IF;

    SELECT count(*) INTO v_nulls
      FROM site_work_items WHERE handler_agent IS NULL;
    IF v_nulls <> 0 THEN
        RAISE EXCEPTION '217: % NULL handler_agent row(s) survived the backfill', v_nulls;
    END IF;

    -- The probe: the outage's own INSERT shape.
    SELECT id INTO v_site FROM sites ORDER BY created_at ASC LIMIT 1;
    IF v_site IS NULL THEN
        RAISE EXCEPTION '217 probe: no site row available to hang a probe work item on';
    END IF;

    INSERT INTO site_work_items
        (site_id, source, item_type, summary, created_by, status)
    VALUES
        (v_site, 'migration_217_probe', 'migration_probe',
         '217 probe: INSERT omitting handler_agent must default to empty, not NULL',
         'migration_217', 'detected')
    RETURNING id INTO v_probe;

    SELECT handler_agent, handler_agent IS NULL
      INTO v_handler, v_isnull
      FROM site_work_items WHERE id = v_probe;

    DELETE FROM site_work_items WHERE id = v_probe;

    IF v_isnull THEN
        RAISE EXCEPTION '217 probe: an INSERT omitting handler_agent still yields NULL — bugs_closed/078 is NOT closed';
    END IF;
    IF v_handler <> '' THEN
        RAISE EXCEPTION '217 probe: omitted handler_agent defaulted to %, expected the empty string', quote_literal(v_handler);
    END IF;

    RAISE NOTICE '217: handler_agent NOT NULL DEFAULT '''' — probe confirms an omitted column now yields '''' , not NULL';
END
$guard$;

COMMIT;
