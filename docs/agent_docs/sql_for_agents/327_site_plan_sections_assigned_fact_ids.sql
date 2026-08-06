-- 327 — site_plan_sections.assigned_fact_ids: which verified facts THIS section
-- is responsible for stating, decided at plan time.
--
-- bugs_open/151 candidate 1. Today every per-section content-writer call receives
-- the identical whole-site writer_block (site_specs evidence_base), so sibling
-- sections independently restate the same facts — measured on fundamentallyai.com
-- as 18 sections across 5 pages each restating 3+ of the same 9 facts, and now
-- counted fleet-wide by the content_duplication census (9 fact-overlap pairs on
-- the motivating site, 2026-08-06). The structural fix is scoping: the planner
-- assigns facts to sections, the writer sees only its section's assignment, and
-- a section that is never told a fact exists cannot restate it.
--
-- SEMANTICS (the writer path relies on this distinction):
--   NULL  -> unscoped: the section gets the whole-site writer_block, i.e. the
--            behaviour every existing row already has. Every pre-existing row is
--            NULL, so nothing changes until a plan is written WITH assignments.
--   '[]'  -> deliberately factless: the writer is told to state no verified
--            business numbers or named-entity claims in this section.
--   '["F1-live-sites", ...]' -> the section's writer block is composed from ONLY
--            these facts (matched against site_specs evidence_base facts[].id at
--            BUILD time — the assignment pins WHICH facts, never their values;
--            a re-verified fact's current number is substituted at compose time).
--            An ID matching no current fact is inert and logged, never fatal.
--
-- ORDERING (stated, not assumed): this column MUST exist before the binary that
-- writes it. It is nullable and additive, so the current binary ignores it
-- completely, and the new binary tolerates NULL everywhere. There is no window
-- in which either half breaks the other.

BEGIN;

ALTER TABLE site_plan_sections
    ADD COLUMN IF NOT EXISTS assigned_fact_ids jsonb;

-- Guard: the column exists and is the type the build path expects. Additive and
-- idempotent, so a re-run is a no-op rather than an error — but a WRONG type
-- would be silent at DDL time and wrong at build time, so assert it.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'site_plan_sections'
          AND column_name = 'assigned_fact_ids'
          AND data_type = 'jsonb'
    ) THEN
        RAISE EXCEPTION '327: site_plan_sections.assigned_fact_ids is missing or not jsonb — plan-time fact scoping cannot be stored';
    END IF;
END $$;

COMMENT ON COLUMN site_plan_sections.assigned_fact_ids IS
    'Verified-fact IDs (site_specs evidence_base facts[].id) this section states, assigned at plan time (bugs_open/151 candidate 1). NULL = unscoped, section gets the whole-site writer_block (pre-existing behaviour). [] = deliberately factless. IDs are matched against the CURRENT evidence_base at build time: values stay live, unknown IDs are inert.';

COMMIT;
