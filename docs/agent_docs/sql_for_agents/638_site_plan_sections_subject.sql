-- 638 — site_plan_sections.subject: what THIS section specifically covers,
-- decided at plan time.
--
-- apis_uk_bees_homepage lane, per-section subjects build (owner go-ahead
-- 2026-08-25). pages.sections is []string, so N same-named slots get one
-- identical brief and the writer produces N variations on one subject —
-- measured four times on apis.uk, where one content_rewrite rewrote all six
-- sections about the waggle dance. The structural fix follows RFC_016 §5.1
-- (RATIFIED 2026-08-08): the subject is the SECOND structured per-section
-- field, riding exactly the rails assigned_fact_ids (migration 327) built —
-- planner object entry -> validate_plan normalise -> this column ->
-- load_page_sections_from_spec section_subjects -> plan_sections ->
-- sectionPlanItem.subject -> the writer's current_section.
--
-- SEMANTICS:
--   NULL / '' -> no subject: the section gets the page-level brief only, i.e.
--                the behaviour every existing row already has.
--   'How honey is graded' -> the writer is told THIS section is about that,
--                and that sibling sections carry their own subjects.
--
-- ORDERING (stated, not assumed): this column MUST exist before the binary
-- that names it rolls — write_site_plan INSERTs it (hard failure without it)
-- and load_page_sections_from_spec SELECTs it (silent tier-degradation
-- without it: the Warn arm drops the build to the non-authoritative tiers).
-- It is nullable and additive, so the CURRENT binary ignores it completely.
-- Apply this file immediately; the Go rides the next roll.

BEGIN;

ALTER TABLE site_plan_sections
    ADD COLUMN IF NOT EXISTS subject text;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'site_plan_sections'
          AND column_name = 'subject'
          AND data_type = 'text'
    ) THEN
        RAISE EXCEPTION '638: site_plan_sections.subject is missing or not text — per-section subjects cannot be stored';
    END IF;
END $$;

COMMENT ON COLUMN site_plan_sections.subject IS
    'One line saying what THIS section specifically covers, assigned at plan time (apis_uk per-section subjects build, 2026-08-26; RFC_016 §5.1 second structured field). NULL/empty = no subject — the section gets the page-level brief only, the pre-existing behaviour. Distinct subjects on same-named sections are what stops N identical briefs producing N sections about one topic.';

COMMIT;
