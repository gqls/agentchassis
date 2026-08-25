-- 615_tool_cta_template_changed_fanout.sql
--
-- WHAT: file one `page_rerender` / `spec.reason='template_changed'` per active,
-- non-owned page carrying the `tool-cta` component, so 614's template change
-- reaches the deployed HTML.
--
-- RUN THIS AFTER 614. On its own it is harmless (a re-render that renders the
-- same bytes); before 614 it is pointless.
--
-- WHY IT IS NEEDED AT ALL. Nothing files `template_changed` re-renders for a
-- template edited by SQL. That fan-out lives in
-- `component-template-fixer.create_rerender`, keyed on the component_id the
-- FIXER just changed — and the fixer is an LLM repair agent, not something you
-- invoke to publish a hand-authored template. So a migration that edits
-- html_template ships NOTHING: every page keeps serving its stored
-- rendered_html, indefinitely, with a green status and no error anywhere. That
-- is bugs_open/283 §13's finding, and it cost a whole verified conversion.
--
-- `template_changed` is in page-rerender's check_rerender_mode reason list
-- (migration 460), so ONE re-render both re-resolves the query.* arrays AND
-- renders the new template. No LLM on this path.
--
-- ── THE SHAPE IS THE LIVE FIXER'S, WITH ONE FILTER IT IS MISSING ────────────
-- Copied from `agent_definitions` (the LIVE row, not migration 460's text —
-- the live query has since gained `p.rebuild_policy IS DISTINCT FROM 'owned'`,
-- which 460 did not have). But the live query STILL HAS NO PAGE-STATUS FILTER,
-- and that is a real defect, not a stylistic difference:
--
--   [MEASURED 2026-08-25] of 60 live tool-cta instances, **16 sit on ARCHIVED
--   pages**. The fixer's query would file re-renders for all 16.
--
-- That is exactly bugs_open/098: `build_status` records whether a page ever
-- shipped, `status` records whether the platform still WANTS it served, and
-- archiving sets the second while leaving the first. A selector keyed on
-- liveness alone re-renders and re-publishes a retired page — which also makes
-- a retraction self-undoing. This file carries `p.status = 'active'`, so it
-- targets **40 pages across 10 sites**, not 60.
--
-- The fixer's own missing filter is filed separately; do not copy that query
-- without this line.
--
-- SECOND DIVERGENCE, also deliberate: this file sets `page_id`. The fixer's
-- INSERT does not list that column, so [MEASURED 2026-08-25] only **4 of 272**
-- live `template_changed` items carry one, while every other reason class is at
-- ~100% (cta_links_stale 274/274, section_data_resolved 198/199). page_id is
-- what lets a reader — and this file's own verify block — join an item back to
-- its page and ask whether it landed somewhere it should not have. Without it
-- that check cannot be written at all.
--
-- ── ITEM KEY: the shared spelling, deliberately ─────────────────────────────
-- The fixer files these KEYLESS (334 of 338 live `template_changed` items have
-- item_key NULL as of 2026-08-25) and dedups with a NOT EXISTS. This file uses
-- the estate's shared key instead — `page_rerender_<page_name>_<site_id>_<reason>`
-- (discovery_checks.PageRerenderItemKey) — so dedup is enforced by
-- idx_swi_dedup at the DB rather than by a race-prone read-then-write. The
-- reason is IN the key by design (bugs_open/024 defect 6): a
-- `template_changed` item can never be dedup-suppressed by an assemble-only or
-- `section_data_resolved` item, nor suppress one. The NOT EXISTS is kept as
-- well, because it also covers the 334 keyless incumbents that the unique
-- index cannot see.
--
-- ── EXPECTED RESULT ────────────────────────────────────────────────────────
-- 40 rows inserted, one per page, across 10 sites, none on an archived or
-- `owned` page. The verify block below REFUSES THE COMMIT if the count is not
-- the count the same predicate selects, or if any inserted row landed on a page
-- that should have been excluded.
--
-- Escalation exposure: these ride the shared page_rerender path, which can
-- escalate a page to the content writer when a section lacks a required
-- source:"llm" field (STY-048). Baseline [MEASURED 2026-08-25]: 1 of 36
-- `section_data_resolved` runs escalated over 14 days. Re-read it after this
-- lands; the query is in 603's header.
--
-- ROLLBACK: 615_..._ROLLBACK.sql cancels any of these items still OPEN. Items
-- already claimed or complete are not undone — a completed re-render has
-- already published, and the way back from that is 614's rollback plus a fresh
-- fan-out, not a work-item deletion.

BEGIN;

-- Guard: 614 must have landed, or this fan-out re-renders the OLD template.
DO $$
DECLARE tmpl text;
BEGIN
    SELECT html_template INTO tmpl FROM content_components WHERE name='tool-cta' AND is_active;
    IF tmpl IS NULL THEN
        RAISE EXCEPTION '615: no active tool-cta component';
    END IF;
    IF tmpl !~* '\.image\y' THEN
        RAISE EXCEPTION '615: tool-cta does not render .image — apply 614 first, or this fan-out re-renders the unchanged template';
    END IF;
END $$;

CREATE TEMP TABLE _615_targets AS
SELECT DISTINCT p.id AS page_id, p.name AS page_name, p.site_id, s.domain
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id
  JOIN content_components cc ON cc.id = pc.component_id
 WHERE cc.name = 'tool-cta'
   AND pc.build_status <> 'removed'
   AND p.status = 'active'                                   -- bugs_open/098; the fixer's query lacks this
   AND COALESCE(p.rebuild_policy, 'generic') <> 'owned';     -- save_sections refuses owned pages on this branch

INSERT INTO site_work_items (site_id, page_id, source, pipeline, item_type, severity, summary,
                             priority, handler_agent, status, created_by, spec, item_key, batch_id)
SELECT t.site_id, t.page_id, 'side_effect', 'build', 'page_rerender', 'low',
       'Rerender ' || t.page_name || ' — tool-cta template now renders the listing image (bugs_open/384 decision 4, migration 614)',
       80, 'page-rerender', 'triaged', 'bugfix_384_toolcta_fanout',
       jsonb_build_object('reason', 'template_changed',
                          'page_id', t.page_id::text,
                          'page_name', t.page_name,
                          'domain', t.domain,
                          'cause', 'template_changed:tool-cta'),
       'page_rerender_' || t.page_name || '_' || t.site_id::text || '_template_changed',
       gen_random_uuid()
  FROM _615_targets t
 WHERE NOT EXISTS (
       SELECT 1 FROM site_work_items w
        WHERE w.site_id = t.site_id
          AND w.item_type = 'page_rerender'
          AND w.spec->>'page_id' = t.page_id::text
          AND w.spec->>'reason' = 'template_changed'
          AND w.status IN ('detected','triaged','claimed'))
ON CONFLICT DO NOTHING;

-- Verify, or refuse the COMMIT.
DO $$
DECLARE n_targets int; n_filed int; n_bad int;
BEGIN
    SELECT count(*) INTO n_targets FROM _615_targets;
    IF n_targets = 0 THEN
        RAISE EXCEPTION '615 verify: the target set is EMPTY — tool-cta is placed on no active, non-owned page; that contradicts the measurement this file was written against';
    END IF;

    SELECT count(*) INTO n_filed FROM site_work_items
     WHERE created_by = 'bugfix_384_toolcta_fanout';

    -- Every target must now be covered by an OPEN template_changed item —
    -- either one this file inserted, or a pre-existing one the NOT EXISTS
    -- deliberately deferred to. Anything less means rows were silently dropped.
    IF EXISTS (
        SELECT 1 FROM _615_targets t
         WHERE NOT EXISTS (
            SELECT 1 FROM site_work_items w
             WHERE w.site_id = t.site_id
               AND w.item_type = 'page_rerender'
               AND w.spec->>'page_id' = t.page_id::text
               AND w.spec->>'reason' = 'template_changed'
               AND w.status IN ('detected','triaged','claimed'))) THEN
        RAISE EXCEPTION '615 verify: a target page has no open template_changed item — the INSERT dropped rows (ON CONFLICT swallowed a key clash?)';
    END IF;

    -- Nothing may have landed on a page the filters exist to exclude.
    SELECT count(*) INTO n_bad
      FROM site_work_items w JOIN pages p ON p.id = w.page_id
     WHERE w.created_by = 'bugfix_384_toolcta_fanout'
       AND (p.status <> 'active' OR COALESCE(p.rebuild_policy,'generic') = 'owned');
    IF n_bad > 0 THEN
        RAISE EXCEPTION '615 verify: % filed item(s) sit on an archived or owned page — the bugs_open/098 filter did not hold', n_bad;
    END IF;

    RAISE NOTICE '615: % target page(s), % item(s) filed under bugfix_384_toolcta_fanout', n_targets, n_filed;
END $$;

COMMIT;
