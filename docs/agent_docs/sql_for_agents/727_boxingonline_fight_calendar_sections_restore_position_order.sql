-- 727_boxingonline_fight_calendar_sections_restore_position_order.sql
--
-- bugs_open/427, Phase B step 2 residue. Corrects a defect in THIS LANE'S OWN
-- migration 719, found 2026-09-03 while preparing the ff91e666 council
-- resubmission — not by any detector.
--
-- WHAT 719 GOT WRONG. Its UPDATE rebuilt the array as
--
--     SELECT jsonb_agg(DISTINCT x) FROM jsonb_array_elements(sections) x
--     WHERE x <> '"generic-text-block"'::jsonb
--
-- `jsonb_agg` with DISTINCT and no ORDER BY does not preserve input order, so
-- the surviving entries came back in an arbitrary (here, sorted) order and
-- `event-list` was then appended. The array went from
--
--     ["hero-tool", "generic-text-block", "advertising"]      (before 719)
-- to  ["advertising", "hero-tool", "event-list"]              (after 719)
--
-- when the correct result — 719's own stated intent, "replaces the array
-- entry" — was ["hero-tool", "event-list", "advertising"].
--
-- WHY IT MATTERS: `pages.sections` IS ORDER-BEARING, and by index, not by
-- membership. save_page_sections_action.go:1979 states it outright ("stores the
-- planned section names in position order (1-indexed)"); adopt_fragment_section.go
-- replaces a section with `planned[Position-1]`; and section_editor_actions.go's
-- loadPageComponentBySlotRO carries a third match arm keyed on
-- `p.sections->(pc.position - 1)`. After 719, page_components position 1
-- (hero-tool) indexed to "advertising" and position 2 (event-list) indexed to
-- "hero-tool" — every one of those joins reads the wrong name.
--
-- NOT LIVE DAMAGE TODAY, stated rather than implied: the section-editor arm is
-- gated on `pc.slot_name IS NULL OR pc.slot_name = ''`, and both rows on this
-- page carry non-empty slot_names ('hero-tool', 'event-list'), so it cannot fire
-- here as things stand. This is a LATENT misalignment that becomes live the
-- moment a build leaves a slot_name empty or an adopt/fragment path runs — and
-- it is silent in both directions, because a wrong-but-present name matches
-- nothing rather than erroring.
--
-- SEPARATELY, PRE-EXISTING, NOT FIXED HERE: "advertising" is declared in this
-- array and has no page_components row on this page (content_components
-- 'ad_zone_inline', function='advertising', component_level='section'). That
-- predates 719 and predates this lane — it was in the array before the swap —
-- so it is left exactly as found and named here rather than quietly tidied.
-- If check_unresolved_sections is flagging this page, that entry is why, and
-- it is a separate decision from this one-row ordering fix.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/bugfix_427_event_render/
-- Council: submitted with 712 and 719 under the ff91e666 resubmission.
\set ON_ERROR_STOP on
BEGIN;

-- Pre-check 1: the page exists and still holds the exact post-719 array. If it
-- does not, something else has moved it and this migration must not guess.
DO $$
DECLARE cur jsonb;
BEGIN
  SELECT sections INTO cur FROM pages
   WHERE id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
     AND site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152'
     AND name = 'tool-fight-calendar';
  IF cur IS NULL THEN
    RAISE EXCEPTION '727 ABORT: tool-fight-calendar page not found on boxingonline.com';
  END IF;
  IF cur <> '["advertising", "hero-tool", "event-list"]'::jsonb THEN
    RAISE EXCEPTION '727 ABORT: sections is not the post-719 value this migration corrects — found %', cur;
  END IF;
END $$;

-- Pre-check 2: the target order must match the page's ACTUAL live composition,
-- so this restores an invariant rather than asserting a preference. Both rows
-- must still be live and in the positions the new array will claim for them.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE pc.page_id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
    AND COALESCE(pc.build_status,'pending') <> 'removed'
    AND ((pc.position = 1 AND cc.function = 'hero-tool')
      OR (pc.position = 2 AND cc.function = 'event-list'));
  IF n <> 2 THEN
    RAISE EXCEPTION '727 ABORT: expected hero-tool at position 1 and event-list at position 2, matched % row(s)', n;
  END IF;
END $$;

UPDATE pages
SET sections = '["hero-tool", "event-list", "advertising"]'::jsonb,
    updated_at = NOW()
WHERE id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
  AND site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152'
  AND name = 'tool-fight-calendar'
  AND sections = '["advertising", "hero-tool", "event-list"]'::jsonb;

-- VERIFY BEFORE COMMIT — a check that CAN fail the transaction, not a printed
-- SELECT (LANDMINES: a verify block of plain SELECTs cannot stop the COMMIT).
-- It asserts the INDEX ALIGNMENT itself, not just the array's value, because
-- the value is only interesting insofar as it indexes correctly.
DO $$
DECLARE bad int;
BEGIN
  SELECT count(*) INTO bad
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE pc.page_id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
    AND COALESCE(pc.build_status,'pending') <> 'removed'
    AND pc.position BETWEEN 1 AND jsonb_array_length(p.sections)
    AND trim(both '"' from (p.sections->(pc.position - 1))::text) <> cc.function;
  IF bad <> 0 THEN
    RAISE EXCEPTION '727 ABORT: % live page_component(s) still index to the wrong pages.sections entry', bad;
  END IF;
END $$;

COMMIT;

-- Rollback (by hand — one row, one column):
-- UPDATE pages SET sections = '["advertising", "hero-tool", "event-list"]'::jsonb
-- WHERE id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33';
