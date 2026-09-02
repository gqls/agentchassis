-- 717 — pages gains section_subjects + section_facts: same-row aligned siblings
-- of pages.sections, so a page the plan tables do not serve can still scope its
-- repeated component types (bugs_open/443).
--
-- WHY. load_page_sections_from_spec publishes per-section subjects and facts
-- only from the authoritative tier (site_plan_sections), because index-aligning
-- a scoping list against a DIFFERENT tier's section list is a guess. On the 6
-- real sites with no current site_plans row every page resolves at tier 2/3/4,
-- so scoping is structurally unreachable and repeated component types write the
-- same section ([MEASURED 2026-09-02] 11 pages, all serving real repetition,
-- >=8 with verbatim-identical h2 pairs, all resolving at tier 3). These columns
-- give tier 3 a store whose alignment is a FACT: written beside the very list
-- they describe, same row, read in the same statement.
--
-- SEMANTICS (the loader enforces these; the columns are dumb storage):
--   * NULL = no scoping = pre-existing behaviour, byte-identical.
--   * A jsonb array aligned BY INDEX with pages.sections: entries are a one-line
--     subject string (or null) / an array of fact ids (or null).
--   * Aligned or absent, never guessed: the loader applies them only when the
--     length (and, for a collected_data-served list, content) matches the list
--     it is serving. A misaligned array is IGNORED with a WARN — kept for the
--     operator to re-align, never applied, never destroyed.
--   * Deliberately NO check constraint and NO trigger: pages.sections has 19
--     candidate writer files (as of 2026-09-02) that know nothing of these
--     columns; a constraint would error them all, a nulling trigger would
--     destroy operator data. The read guard + the build-side detector
--     (REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT) make the degrade visible.
--
-- ORDER: column-before-binary, same as 638 and for the same reason — the loader
-- SELECTs these columns, so the column must exist before the image rolls.
-- Config: NONE needed (the plan_sections wiring for both arrays is already
-- live — seed 639, applied 2026-09-02).
--
-- Council: Council-Submitted b7c59309-1f70-448f-9d20-1c47ebf64196.

BEGIN;

ALTER TABLE pages ADD COLUMN IF NOT EXISTS section_subjects jsonb;
ALTER TABLE pages ADD COLUMN IF NOT EXISTS section_facts jsonb;

COMMENT ON COLUMN pages.section_subjects IS
  'bugs_open/443: per-slot one-line subjects, a jsonb array aligned by index with sections (null entries = no subject). Applied by load_page_sections_from_spec ONLY when aligned with the list it serves; misaligned = ignored with a WARN, never guessed, never auto-deleted. NULL = no subjects (pre-existing behaviour). Any writer that replaces sections without re-aligning this column silently disarms it — see LANDMINES.';
COMMENT ON COLUMN pages.section_facts IS
  'bugs_open/443: per-slot fact-id scoping, a jsonb array aligned by index with sections (null entries = unscoped). Same aligned-or-absent contract as section_subjects; mirrors site_plan_sections.assigned_fact_ids one tier down.';

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM information_schema.columns
    WHERE table_name = 'pages'
      AND column_name IN ('section_subjects', 'section_facts')
      AND data_type = 'jsonb' AND is_nullable = 'YES';
    IF n <> 2 THEN
        RAISE EXCEPTION '717: expected 2 nullable jsonb columns on pages, found %', n;
    END IF;
END $$;

COMMIT;
