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
-- a5cabea0 has two triaged content_rewrite dependents (9e9ec430, 18bc832c).
-- A dependency is released only by complete/verified (bugs_closed/176), so
-- they are cleared explicitly — the crosslink work stands on its own merits:
-- the tool page IS deployed (widget live), which is all a crosslink needs.
-- Same reasoning as 286's treatment of 93f2a3b7.
--
-- Verify after applying (expect: first query 0 rows; second query the two
-- ids with depends_on NULL, status unchanged 'triaged'):
--   SELECT id FROM site_work_items
--   WHERE item_key ~ '^tool_content:' AND status='needs_human_review';
--   SELECT left(id::text,8), status, depends_on FROM site_work_items
--   WHERE id IN ('9e9ec430-ff92-4264-83cc-6072840faad8',
--                '18bc832c-c937-4608-9a05-718772d44c88');
-- (ids verified against the live DB 2026-08-03 11:45; both are item_type
--  'content_rewrite' with item_key 'tool_crosslink:tool-cma-obligation-checker:…')

BEGIN;

-- The two dependents first, so the ids are printed before anything changes.
SELECT left(id::text,8) AS dependent, status, depends_on
FROM site_work_items
WHERE depends_on && ARRAY(
  SELECT id FROM site_work_items
  WHERE item_key ~ '^tool_content:' AND status = 'needs_human_review'
);

UPDATE site_work_items
SET depends_on = NULL,
    updated_at = NOW()
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
  IF blocked <> 0 THEN
    RAISE EXCEPTION 'sweep incomplete: % live items still depend on a tool_content row', blocked;
  END IF;
END $$;

COMMIT;
