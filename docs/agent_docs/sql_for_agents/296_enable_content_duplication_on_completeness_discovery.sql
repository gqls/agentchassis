-- 296 — enable the content_duplication check on completeness-discovery-agent
--
-- WHAT: appends 'content_duplication' to completeness-discovery-agent's
-- run_checks list. This is the enable switch bugs_open/151 candidate 3 was
-- built for (gauntlet_dead_cta lane, council da3f2d9b APPROVED r3; plan guard
-- 6c5d1491 APPROVED). The check finds same-page + same-slot + byte-identical
-- page_components rows and dispatches the deterministic repair
-- (remove_duplicate_page_sections via the deduplicate-sections handler); the
-- near-duplicate residue files one flag-only capability_gap per site with
-- do_not_auto_rewrite: true and no handler.
--
-- WHY completeness-discovery-agent: its list is the rows-of-a-page family
-- (empty_sections, sectionless_pages, unlinked_page_components,
-- truncated_component). quality-discovery-agent reads prose; this check reads
-- row structure.
--
-- ORDER IS LOAD-BEARING. The enable decision belonged to the
-- brochure_component_library lane (CONTRIB_2026-07-31_151_candidate_3_is_built)
-- and the owner said do it (2026-08-03). Verified before this file was applied,
-- both replicas of v1.0.1238, one exec each ('strings' is absent from the
-- image; grep -acF on the binary):
--
--   plan_specified_repetition   -> 1   guard half, 5c4dc317f
--   lock_skipped                -> 1   lock predicate fix, 594b422e7
--   content_duplication         -> 7   the check itself
--   remove_duplicate_page_sections -> 6  the repair action
--   request_browser_run         -> 5   positive control (pre-existing)
--   zqxjkwv_nonexistent         -> 0   discriminator
--
-- Census re-run 2026-08-03 with the SHIPPED functions
-- (gauntlet_dead_cta/scripts/dedup_census_shipped.go): sections=1189,
-- SHIPPED RULE groups=0 rows_deleted=0. Upstream plan repetition re-measured:
-- exactly 1 group within a current plan (webdesign.co.uk/index/info-card-grid
-- x2), which the guard refuses to touch by design.
--
-- Idempotent (the append is fenced on absence). DB-only; snapshot-prefixed.

BEGIN;

SELECT snapshot_agent('completeness-discovery-agent',
    '296_enable_content_duplication_on_completeness_discovery.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config #> '{workflow,steps,run_checks,config,checks}')
        || '["content_duplication"]'::jsonb,
      false)
WHERE type = 'completeness-discovery-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps,run_checks,config,checks}')
        ? 'content_duplication';

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 296: content_duplication is enabled on completeness-discovery-agent
Observed: bugs_open/151's candidate 3 (check + deterministic repair + handler) was built, council-approved twice (checker da3f2d9b r3, plan guard 6c5d1491) and deliberately left inert; the enable decision sat with the brochure_component_library lane.
Root cause: not-applicable (a switch, not a defect). A discovery check runs only when an agent's workflow names it, and none did.
Fix: 'content_duplication' appended to completeness-discovery-agent's run_checks list, only after (1) the guard chain was pod-verified on both v1.0.1238 replicas with positive and negative controls, (2) the census was re-run with the shipped Go functions over all 1,189 live sections (0 groups, 0 deletions), (3) the single plan-specified repetition fleet-wide was re-measured and confirmed guarded.
Verified: the repair refuses plan-specified repetition (plan_specified_repetition skip reason) and locked rows (lock_skipped); the near-duplicate residue is flag-only with do_not_auto_rewrite. First sweeps are expected to file 0 dedup items; a non-zero item is worth reading, not alarming — the census figure goes stale by design.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE lst jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,run_checks,config,checks}' INTO lst
    FROM agent_definitions
    WHERE type = 'completeness-discovery-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF lst IS NULL OR NOT lst ? 'content_duplication' THEN
        RAISE EXCEPTION '296: content_duplication not in checks list (%)', lst;
    END IF;
    IF (SELECT count(*) FROM jsonb_array_elements_text(lst) e
        WHERE e = 'content_duplication') <> 1 THEN
        RAISE EXCEPTION '296: content_duplication appears more than once — a re-run appended a duplicate (%)', lst;
    END IF;
END $$;

-- A verify block that only checks its own key cannot tell a surgical append
-- from a write that flattened the step or the workflow. Assert neighbours at
-- both levels: a pre-existing check in the same array, and a sibling step.
DO $$
DECLARE lst jsonb; steps jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,run_checks,config,checks}',
           default_config #> '{workflow,steps}'
      INTO lst, steps
    FROM agent_definitions
    WHERE type = 'completeness-discovery-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF NOT lst ? 'sectionless_pages' THEN
        RAISE EXCEPTION '296: neighbour check sectionless_pages did not survive (%)', lst;
    END IF;
    IF steps -> 'ensure_site_record' IS NULL OR steps -> 'complete' IS NULL THEN
        RAISE EXCEPTION '296: a sibling step vanished — the write flattened the workflow';
    END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #> '{workflow,steps,run_checks,config,checks}'
--   FROM agent_definitions WHERE type='completeness-discovery-agent'
--     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- Verify at the artefact after the next sweep (expect flag-only rows, 0 deletes
-- while the census holds):
--   SELECT site_id, item_type, status, summary FROM site_work_items
--    WHERE item_type IN ('content_duplication','capability_gap')
--      AND created_at > now() - interval '2 days' ORDER BY created_at DESC;
-- Rollback: restore the snapshot, or remove the entry:
--   UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--     '{workflow,steps,run_checks,config,checks}',
--     (SELECT jsonb_agg(e) FROM jsonb_array_elements(default_config
--        #> '{workflow,steps,run_checks,config,checks}') e
--       WHERE e <> '"content_duplication"'::jsonb))
--   WHERE type='completeness-discovery-agent' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
