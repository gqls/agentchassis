-- 297: bugfix 177 — sweep the unsatisfiable tool_content items and release
--      the two dependents blocked on one of them.
--
-- Context: bugs_open/177. tool-generator raised needs_content_page items
-- (item_key 'tool_content:%') for tool pages created with pages.sections=[].
-- page-build-handler resolves sections from site_plan_sections →
-- site_specs.site_plan → pages.sections → sibling synthesis; a freshly
-- generated tool page is in none, so every item no-ops into
-- needs_human_review. 8 of 8 open rows carry the identical error; the class
-- has a 0% completion rate over its entire history (2026-07-14 → 2026-08-02).
--
-- The emit-side fix (the guard in tool_content_item.go) stops NEW rows; this
-- sweep retires the EXISTING ones. Follows the 286 triage precedent:
-- wont_fix, NOT complete — no work was performed, and complete would both
-- assert a success that did not happen and release dependents on a lie.
--
-- DEVIATION FROM THE APPROVED PLAN (2026-08-03 ~12:10, NARROWING — recorded
-- in the lane PLAN and in bugs_open/178): the two triaged content_rewrite
-- dependents (9e9ec430, 18bc832c) are deliberately NOT released. The plan
-- (council 982507b0, edit 4) cleared their depends_on on 286's reasoning
-- ("the crosslink stands on its own merits"). Between approval and apply, the
-- diagnosis run on this exact class COMPLETED and found that dispatching one
-- (93f2a3b7 — the item 286 released on that same reasoning) REGENERATED whole
-- slots and DROPPED PARAGRAPHS on the target page (bugs_open/178's mechanism;
-- doc_notes 2026-08-03). Releasing two more would repeat a known-destructive
-- outcome. Their dep on a wont_fix row keeps them non-dispatchable (the
-- loader skips unresolved deps — bugs_closed/176) — an ugly but VISIBLE
-- interlock. The 154 lane (owner of 178) releases them with its fix.
--
-- Verify after applying (expect: first query 0 rows; second query the two
-- ids still 'triaged' WITH depends_on INTACT — the 178 interlock):
--   SELECT id FROM site_work_items
--   WHERE item_key ~ '^tool_content:' AND status='needs_human_review';
--   SELECT left(id::text,8), status, depends_on FROM site_work_items
--   WHERE id IN ('9e9ec430-ff92-4264-83cc-6072840faad8',
--                '18bc832c-c937-4608-9a05-718772d44c88');
-- (ids verified against the live DB 2026-08-03 11:45; both are item_type
--  'content_rewrite' with item_key 'tool_crosslink:tool-cma-obligation-checker:…')

BEGIN;

-- The two dependents, printed for the record — and deliberately NOT updated
-- (see the deviation note above).
SELECT left(id::text,8) AS dependent, status, depends_on
FROM site_work_items
WHERE depends_on && ARRAY(
  SELECT id FROM site_work_items
  WHERE item_key ~ '^tool_content:' AND status = 'needs_human_review'
);

UPDATE site_work_items
SET status = 'wont_fix',
    error = 'bugfix 177 sweep (2026-08-03): item was unsatisfiable at birth — '
         || 'tool page created with no declared sections, so page-build-handler '
         || 'had nothing to build. Emit-side guard now prevents the class. '
         || 'Original error: ' || COALESCE(error, '(none)'),
    updated_at = NOW()
WHERE item_key ~ '^tool_content:'
  AND status = 'needs_human_review';

-- Induced check (RFC_006 lesson: a verify block of SELECTs cannot stop a
-- COMMIT — use DO/RAISE).
DO $$
DECLARE remaining int; blocked int;
BEGIN
  SELECT count(*) INTO remaining FROM site_work_items
  WHERE item_key ~ '^tool_content:' AND status = 'needs_human_review';
  IF remaining <> 0 THEN
    RAISE EXCEPTION 'sweep incomplete: % tool_content rows still needs_human_review', remaining;
  END IF;
  SELECT count(*) INTO blocked FROM site_work_items w
  WHERE w.status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled')  -- = workItemTerminalStatuses (work_items_common.go:40)
    AND w.depends_on && ARRAY(
      SELECT id FROM site_work_items WHERE item_key ~ '^tool_content:'
    );
  -- EXACTLY the two 178-interlocked dependents must remain; anything else is
  -- a row this sweep did not know about.
  IF blocked <> 2 THEN
    RAISE EXCEPTION 'expected exactly 2 deliberately-blocked dependents (the 178 interlock), found %', blocked;
  END IF;
END $$;

COMMIT;
