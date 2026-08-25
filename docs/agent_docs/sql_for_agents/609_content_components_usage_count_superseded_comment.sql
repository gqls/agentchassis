-- 609 — content_components.usage_count is SUPERSEDED: replace a column comment that
--       actively misdescribes the column with one that says what is true
--
-- bugs_closed/378. The column's comment currently reads:
--
--     "Times this component has been assigned to a page. Incremented by selector.
--      Higher = more battle-tested."
--
-- Every clause of that is false as of 2026-08-25:
--   * "Times assigned to a page" — it counted RESOLUTION ATTEMPTS on ONE of the three
--     resolution paths in plan_sections_action.go. 96 of 151 active section components
--     read 0 while bound to live pages, hiding 1,865 bindings; component_level='tool'
--     was 0 of 115. It also counted things that never became a binding at all: the two
--     largest values in the column belong to components with ZERO page bindings.
--   * "Incremented by selector" — the incrementing helper was DELETED in commit
--     5074367f7, live on chassis 48f55f21 since 2026-08-24 18:55Z. Nothing increments it.
--   * "Higher = more battle-tested" — nothing scores on it. The selector's usage term was
--     removed rather than repaired, because an accurate "prefer what is already used"
--     term is a preferential-attachment loop (bugs_open/107).
--
-- WHY A COMMENT AND NOT THE DROP. The DROP is written and held as
-- 610_..._HOLD.sql. It cannot run until the LAST writer is live: the birth INSERT in
-- store_generated_component_action.go named this column explicitly until the same commit
-- as this migration, and dropping the column against the currently-running binary would
-- break every component creation. This file is safe against BOTH binaries — a COMMENT
-- touches no DML.
--
-- This is the cheap half and it is the half that matters for a reader: `\d+
-- content_components` is where the next person meets this column, and until now it told
-- them a maintained figure was on offer.
--
-- IDEMPOTENT: COMMENT ON is a replace, not an append. Safe to re-run.

\set ON_ERROR_STOP on
BEGIN;

-- Guard: refuse if the column is gone (610 already ran) — commenting a dropped column
-- errors anyway, but this says WHY rather than leaving a bare "column does not exist".
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'content_components' AND column_name = 'usage_count'
  ) THEN
    RAISE EXCEPTION
      'content_components.usage_count no longer exists — 610_..._HOLD.sql has already run, so this comment migration is obsolete. Nothing to do.';
  END IF;
END $$;

COMMENT ON COLUMN content_components.usage_count IS
  'SUPERSEDED AND DEAD as of 2026-08-25 (bugs_closed/378). Nothing writes it and nothing '
  'reads it. Its values are a frozen snapshot of resolution attempts on ONE of three '
  'component-resolution paths, and they both MISS real bindings and COUNT non-usages — '
  'the two largest values belong to components with zero page bindings. DO NOT quote it '
  'as a usage figure and DO NOT revive it. "How proven is this component" is derived at '
  'read time from page_components: see ComponentUsageSitesSQL in '
  'platform/orchestration/actions/component_selector.go. Scheduled for DROP by migration 610.';

-- VERIFY: a DO/RAISE, not a SELECT. A verify block made of SELECTs cannot stop the
-- COMMIT — ON_ERROR_STOP ignores a non-empty result set (LANDMINES).
DO $$
DECLARE
  c text;
BEGIN
  SELECT col_description('content_components'::regclass, a.attnum) INTO c
    FROM pg_attribute a
   WHERE a.attrelid = 'content_components'::regclass AND a.attname = 'usage_count';

  IF c IS NULL OR c NOT LIKE 'SUPERSEDED AND DEAD%' THEN
    RAISE EXCEPTION 'VERIFY FAILED: comment did not take (got: %)', COALESCE(left(c, 60), '<null>');
  END IF;
  IF c LIKE '%battle-tested%' THEN
    RAISE EXCEPTION 'VERIFY FAILED: the OLD comment is still in place';
  END IF;
  RAISE NOTICE 'VERIFY: PASS — content_components.usage_count is documented as superseded';
END $$;

COMMIT;
