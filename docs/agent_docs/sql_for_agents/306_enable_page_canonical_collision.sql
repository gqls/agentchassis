-- 306 — enable the page_canonical_collision check on completeness-discovery-agent
--
-- WHAT: appends 'page_canonical_collision' to completeness-discovery-agent's
-- run_checks list. This is the enable switch for bugs_open/080's fix candidate
-- 3 (bugfix_080_canonical_collisions lane): two `pages` rows on one site whose
-- canonical form collides — the same logical page existing twice because two
-- creation surfaces derived its identity differently. The check mutates
-- nothing: a group with >=2 ACTIVE claimants files ONE needs_human_review item
-- (no handler — a decision, the bugs_closed/081 idiom) whose spec carries the
-- decided section-index-family convention; groups with fewer active claimants
-- are findings only; a group a human has ruled wont_fix/rejected is not
-- re-filed; open items whose group has lost its second claimant are retracted
-- via the RFC_010 seam.
--
-- WHY completeness-discovery-agent: its list is the rows-of-a-page family
-- (empty_sections, sectionless_pages, unlinked_page_components,
-- content_duplication). This check reads `pages` row structure.
--
-- ORDER IS LOAD-BEARING: image first, then this seed. Verify on EVERY replica
-- before applying (grep -acF on the binary; 'strings' is absent from the image):
--
--   page_canonical_collision          -> >=1  the check itself
--   collisionCanonName                -> >=1  its helper
--   "/tools/%s.html", function        -> 0    negative control: the silent
--                                             fallback the same commit REMOVED
--                                             from create_tool_component
--   request_browser_run               -> >=1  positive control (pre-existing)
--   zqxjkwv_nonexistent               -> 0    discriminator
--
-- EXPECTED FIRST SWEEP (measured 2026-08-03, goes stale by design): exactly 2
-- items fleet-wide, both robot-hands.com (/news and /gripper-catalog — two live
-- claimants each), and 4 finding-only groups (dartsonline x3, robot-hands
-- learning-center). A different count is worth reading, not alarming.
--
-- Idempotent (the append is fenced on absence). DB-only; snapshot-prefixed.

BEGIN;

SELECT snapshot_agent('completeness-discovery-agent',
    '306_enable_page_canonical_collision.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config #> '{workflow,steps,run_checks,config,checks}')
        || '["page_canonical_collision"]'::jsonb,
      false)
WHERE type = 'completeness-discovery-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps,run_checks,config,checks}')
        ? 'page_canonical_collision';

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 306: page_canonical_collision is enabled on completeness-discovery-agent
Observed: bugs_open/080 — two page-creation surfaces derived the same logical page's identity differently; pages is unique on (site_id, name) with NO unique index on url, so the divergent identity INSERTED a second row. robot-hands.com served two live duplicate pairs (/news.html beside /news/index.html with an identical title; /gripper-catalog.html beside /gripper-catalog/index.html) for weeks, the stray carrying a self-referential rel=canonical.
Root cause: the creation surfaces, fixed separately (gap-planner v1.0.1177; create_blog_posts, deploy_tool, create_tool_component in this lane's commit). This check is the detection half: nothing watched the stored state for collisions that already exist or arrive by a path not yet closed.
Fix: 'page_canonical_collision' appended to completeness-discovery-agent's run_checks list, only after the image carrying check_page_canonical_collision.go was pod-verified on every replica with positive and negative controls. Two grouping signals union-merged (canonical-name via datahelpers.CanonicalisePage, URL path-key) — measured: each signal alone misses one of the two live pairs.
Verified: expected first sweep files exactly 2 items, both robot-hands.com, both needs_human_review with no handler. NOTE the queue reality: promotion of detected items is manual while bugs_open/083 stands, and needs_human_review has no working surface (bugs_open/033) — visible-and-stuck beats silent-and-looping, the same call bugs_closed/081 made.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE lst jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,run_checks,config,checks}' INTO lst
    FROM agent_definitions
    WHERE type = 'completeness-discovery-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF lst IS NULL OR NOT lst ? 'page_canonical_collision' THEN
        RAISE EXCEPTION '306: page_canonical_collision not in checks list (%)', lst;
    END IF;
    IF (SELECT count(*) FROM jsonb_array_elements_text(lst) e
        WHERE e = 'page_canonical_collision') <> 1 THEN
        RAISE EXCEPTION '306: page_canonical_collision appears more than once — a re-run appended a duplicate (%)', lst;
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
        RAISE EXCEPTION '306: neighbour check sectionless_pages did not survive (%)', lst;
    END IF;
    IF steps -> 'ensure_site_record' IS NULL OR steps -> 'complete' IS NULL THEN
        RAISE EXCEPTION '306: a sibling step vanished — the write flattened the workflow';
    END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #> '{workflow,steps,run_checks,config,checks}'
--   FROM agent_definitions WHERE type='completeness-discovery-agent'
--     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- Verify at the artefact after the next sweep (expect exactly the 2
-- robot-hands items while the 2026-08-03 census holds):
--   SELECT s.domain, swi.item_key, swi.status, swi.summary
--     FROM site_work_items swi JOIN sites s ON s.id = swi.site_id
--    WHERE swi.item_type = 'page_canonical_collision'
--    ORDER BY swi.created_at DESC;
-- Rollback: restore the snapshot, or remove the entry:
--   UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--     '{workflow,steps,run_checks,config,checks}',
--     (SELECT jsonb_agg(e) FROM jsonb_array_elements(default_config
--        #> '{workflow,steps,run_checks,config,checks}') e
--       WHERE e <> '"page_canonical_collision"'::jsonb))
--   WHERE type='completeness-discovery-agent' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
