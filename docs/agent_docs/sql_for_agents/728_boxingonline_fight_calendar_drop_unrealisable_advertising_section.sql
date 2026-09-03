-- 728_boxingonline_fight_calendar_drop_unrealisable_advertising_section.sql
--
-- bugs_open/427, closing the `advertising` entry the council's guardian seat
-- objected to on the ff91e666 round-2 REVISE (2026-09-03). Its objection:
-- "The 'advertising' array entry with no page_components row is left in place.
-- This is a known declared-but-unresolved shape that check_unresolved_sections'
-- next sweep can act on, potentially reopening the full-rebuild path this whole
-- correlation was built to avoid. Author names it but does not close it or
-- explain why it is safe to leave."
--
-- THE SEAT IS RIGHT, AND IT IS ARMED RIGHT NOW — not hypothetical. Migration
-- 727's own header called this "pre-existing, left exactly as found"; that was
-- true and it was not a justification. Checked against the detector's actual
-- predicate (discovery_checks/check_unresolved_sections.go:36-56):
--
--   page status='active'          -> tool-fight-calendar: active        MATCH
--   page build_status='deployed'  -> tool-fight-calendar: deployed      MATCH
--   a live component matches the declared name
--     ('ad_zone_inline', function='advertising', is_active, forked_from IS NULL)  MATCH
--   NO page_components row on this page joins to it                     MATCH
--
-- All four arms hold, so the next unresolved_sections sweep for this site sets
-- build_status='needs_rebuild' on this page and routes it into the full
-- page-build-handler pipeline — the precise path the component_swap route was
-- chosen to avoid, and whose carry-forward behaviour for the two live sections
-- is still unverified.
--
-- WHY REMOVE THE DECLARATION RATHER THAN REALISE IT. [MEASURED 2026-09-03]
-- ZERO page_components rows fleet-wide join to a component with
-- function='advertising', across every site and every non-removed build_status.
-- The component exists in the library and nothing has ever placed one on a page.
-- Three active pages declare it, all on boxingonline.com: 'index' (already
-- needs_rebuild), 'cruiserweight-boxings-best-kept-secret', and this one. So the
-- entry is PLAN RESIDUE — a name in a manifest that no pipeline realises — and
-- removing it makes the declared manifest agree with what the page is and will
-- be, which is the same invariant 719 and 727 were each restoring.
--
-- [NOT ESTABLISHED], and deliberately not asserted: whether a rebuild would
-- realise an advertising row at all. If it would not, these pages are on a
-- permanent re-arm treadmill (marked -> rebuilt -> still unresolved -> marked),
-- which would be a defect in its own right. I have not read the build path far
-- enough to claim that and it is not this migration's business.
--
-- SCOPE, STATED RATHER THAN SILENTLY NARROWED: this migration touches ONE page,
-- the one this lane owns and the one whose rebuild the correlation exists to
-- prevent. The other two boxingonline pages carrying the same residue are named
-- above and NOT touched here — 'index' is already needs_rebuild so the door has
-- already opened on it, and 'cruiserweight-...' is a content page this lane has
-- no business re-planning. Fleet population armed by the same predicate today:
-- 18 pages across 3 sites. That is a finding for bugs_open/427 §16, not a licence
-- for this migration to widen.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/bugfix_427_event_render/
-- Council: ff91e666 round 3.
\set ON_ERROR_STOP on
BEGIN;

-- Pre-check 1: exact expected pre-state. Refuses if anything else moved the row.
DO $$
DECLARE cur jsonb;
BEGIN
  SELECT sections INTO cur FROM pages
   WHERE id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
     AND site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152'
     AND name = 'tool-fight-calendar';
  IF cur IS NULL THEN
    RAISE EXCEPTION '728 ABORT: tool-fight-calendar page not found on boxingonline.com';
  END IF;
  IF cur <> '["hero-tool", "event-list", "advertising"]'::jsonb THEN
    RAISE EXCEPTION '728 ABORT: sections is not the post-727 value this migration edits — found %', cur;
  END IF;
END $$;

-- Pre-check 2: the entry really is unrealisable HERE. If an advertising row has
-- appeared on this page since the census, removing the declaration would be
-- wrong and this migration must refuse rather than proceed.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  WHERE pc.page_id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
    AND cc.function = 'advertising'
    AND COALESCE(pc.build_status,'pending') <> 'removed';
  IF n <> 0 THEN
    RAISE EXCEPTION '728 ABORT: % advertising row(s) now exist on this page — the declaration is realised, do not drop it', n;
  END IF;
END $$;

-- The write. A LITERAL, not an aggregate: the order is the thing 727 had to
-- restore, and jsonb_agg(DISTINCT x) without ORDER BY is exactly what lost it
-- (see LANDMINES, "a jsonb_agg(DISTINCT) rebuild of an ORDER-BEARING array").
UPDATE pages
SET sections = '["hero-tool", "event-list"]'::jsonb,
    updated_at = NOW()
WHERE id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
  AND site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152'
  AND name = 'tool-fight-calendar'
  AND sections = '["hero-tool", "event-list", "advertising"]'::jsonb;

-- VERIFY BEFORE COMMIT — RAISE, not a printed SELECT. Asserts BOTH properties
-- that matter: the detector can no longer arm this page, and the index
-- alignment 727 restored is still intact.
DO $$
DECLARE armed int; bad int;
BEGIN
  SELECT count(*) INTO armed
  FROM pages p, jsonb_array_elements_text(p.sections) AS sec(section_name)
  WHERE p.id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
    AND p.status = 'active' AND p.build_status = 'deployed'
    AND EXISTS (SELECT 1 FROM content_components cc WHERE cc.is_active AND cc.forked_from IS NULL
                AND (cc.section_type = sec.section_name OR cc.function = sec.section_name))
    AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc2 ON pc.component_id = cc2.id
                    WHERE pc.page_id = p.id
                      AND (cc2.section_type = sec.section_name OR cc2.function = sec.section_name));
  IF armed <> 0 THEN
    RAISE EXCEPTION '728 ABORT: page still arms check_unresolved_sections on % declared name(s)', armed;
  END IF;

  SELECT count(*) INTO bad
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE pc.page_id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
    AND COALESCE(pc.build_status,'pending') <> 'removed'
    AND pc.position BETWEEN 1 AND jsonb_array_length(p.sections)
    AND trim(both '"' from (p.sections->(pc.position - 1))::text) <> cc.function;
  IF bad <> 0 THEN
    RAISE EXCEPTION '728 ABORT: % live page_component(s) index to the wrong pages.sections entry', bad;
  END IF;
END $$;

COMMIT;

-- Rollback (by hand — one row, one column):
-- UPDATE pages SET sections = '["hero-tool", "event-list", "advertising"]'::jsonb
-- WHERE id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33';
