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

--
-- ── VERIFY THE ROLL FIRST (council 4bd35ed8 r2, debug_historian: "post-roll" is
--    an assumption until the RUNNING BINARY says so; same-tag rebuilds have
--    shipped nothing before). Two probes, both against the pod, controls in the
--    same breath:
--      kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--        -> git merge-base --is-ancestor 35905c547 <stamp sha>   (must exit 0)
--      # fallback if the startup line has scrolled (capability probe + controls):
--      P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
--      kubectl -n ai-persona-system exec $P -- grep -ac 'section_subjects' /proc/1/exe   # >0 = shipped
--      kubectl -n ai-persona-system exec $P -- grep -ac 'section_facts' /proc/1/exe      # positive control, must be >0
--
-- ── AFTER HAND-APPLYING: _HOLD files never reach the migration ledger, so the
--    record is THIS FILE — append one line directly below this block and commit
--    it (pathspec):  -- APPLIED <date> by <session>; roll verified at <stamp sha>

SELECT snapshot_agent('page-build-handler', '639_page_build_handler_wires_section_subjects_HOLD.sql: pre-update');

BEGIN;

DO $$
DECLARE cfg jsonb; n int;
BEGIN
    -- Pre-flight (council 4bd35ed8 round 1, editquality): the duplicate-
    -- active-definition-row landmine would make the UPDATE's WHERE ambiguous;
    -- fail BEFORE writing, not after.
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'page-build-handler' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '639: expected exactly 1 active page-build-handler row BEFORE writing, found % — duplicate-active-row landmine', n;
    END IF;
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
