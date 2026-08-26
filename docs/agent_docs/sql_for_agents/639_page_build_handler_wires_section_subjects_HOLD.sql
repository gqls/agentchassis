-- 639 — page-build-handler wires section_subjects into plan_sections.
--
-- ⚠ _HOLD: apply BY HAND, AFTER the image carrying sectionPlanItem.Subject has
-- rolled (image first, config second — same ordering note as seed 328, whose
-- shape this mirrors exactly). Before the roll the key is dead config; the
-- binary warns on the unknown key and continues.
--
-- Adds ONE key to page-build-handler's plan_sections step config:
--
--   "section_subjects": "spec_sections.section_subjects"
--
-- load_page_sections_from_spec emits section_subjects (aligned with sections)
-- ONLY when its authoritative tier (site_plan_sections) served the list;
-- plan_sections consumes it ONLY when this config key names it — opt-in at
-- the step config, unsafe default OFF, per the owner ruling of 2026-08-02.
-- The writer's own fallback plan_sections step (bugs_open/087) is left
-- unwired, exactly as it is for section_facts: a fallback re-plan has no
-- authoritative alignment to trust.

SELECT snapshot_agent('page-build-handler', '639_page_build_handler_wires_section_subjects_HOLD.sql: pre-update');

BEGIN;

DO $$
DECLARE cfg jsonb;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_sections'->'config' INTO cfg
    FROM agent_definitions
    WHERE type = 'page-build-handler' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg IS NULL THEN
        RAISE EXCEPTION '639: page-build-handler plan_sections config not found';
    END IF;
    IF cfg ? 'section_subjects' THEN
        RAISE EXCEPTION '639: already applied — section_subjects key present';
    END IF;
    IF cfg->>'section_facts' IS DISTINCT FROM 'spec_sections.section_facts' THEN
        RAISE EXCEPTION '639: sibling section_facts wiring (seed 328) absent or changed — re-derive this seed from the live row';
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_sections,config,section_subjects}',
        '"spec_sections.section_subjects"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type = 'page-build-handler' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
      AND default_config->'workflow'->'steps'->'plan_sections'->'config'->>'section_subjects'
          = 'spec_sections.section_subjects';
    IF n <> 1 THEN
        RAISE EXCEPTION '639: expected exactly 1 page-build-handler row carrying the section_subjects wiring, found %', n;
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run):
--   UPDATE agent_definitions
--   SET default_config = default_config #- '{workflow,steps,plan_sections,config,section_subjects}'
--   WHERE type='page-build-handler' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
