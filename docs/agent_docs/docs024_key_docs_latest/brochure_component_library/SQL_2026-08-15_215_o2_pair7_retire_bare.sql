-- SQL_2026-08-15_215_o2_pair7_retire_bare.sql
-- bugs_open/215 O2 — PAIR 7, robot-hands.com `gripper-cycle-time-estimator`.
-- Runbook steps 3, 4 and 5 in ONE transaction. Step 6 (retraction) is a dispatch,
-- not SQL, and follows this file. Step 7 does not exist (there is no redirect
-- mechanism — RUNBOOK ⚠ CORRECTION 2026-08-14). Modelled on pair 5's
-- SQL_2026-08-14_215_o2_pair5_payload_calculator.sql, which ran clean end to end.
--
-- OWNER RULING 2026-08-13: keep `tool-gripper-cycle-time-estimator`, MERGE the bare
-- page's prose into it, THEN retire bare. Unchanged by the 08-14 reversal (pairs 3+4
-- only). Scope of "the prose" settled by the 2026-08-15 ruling: explainer + FAQ.
--
-- ⚠ THE MERGE HALF IS ALREADY DONE AND LIVE — DO NOT RUN THIS FILE BEFORE IT.
-- SQL_2026-08-15_215_o2_pair7_cycle_time_merge.sql moved `generic-text-block` and
-- `faq` onto the survivor, deployed assemble-only (corr 537f5b76…), verified at the
-- artefact: survivor 32,165 b/2,129 w -> 44,478 b/3,694 w. **Retiring the bare page
-- before that merge would have destroyed 1,587 words.** The pre-flight below asserts
-- the merge landed, so this file cannot run in the wrong order.
--
-- WHY THIS PAIR IS SAFE TO EXECUTE NOW [MEASURED 2026-08-15, read-only]:
--   * the loser has ZERO editorial inbound referrers (0 page bodies, 0 site chrome)
--     and ZERO active nav rows — from retract_page_graph.go's own three census
--     queries, run in the same breath over pair 6 (`matchmatrix`), which returned
--     2 bodies + 2 chrome + 1 nav. **That non-zero is the positive control**: the
--     query CAN match on this site, so this pair's zero is not an inert query, and
--     it independently reproduces §15's "4 editorial + 1 nav" for pair 6.
--     (link_registry is EMPTY fleet-wide and can never be the instrument —
--     LANDMINES 2026-08-14.)
--   * served artefacts: loser 46,158 b / survivor 44,478 b / 404 control 2,886 b.
--   * the survivor is `rebuild_policy='owned'`, so PBP-036 already guards it.
--
-- ORDER MATTERS, AND STEP 3 IS MANDATORY HERE: the current plan (7a40a0f9…,
-- is_current) names BOTH sides. Archiving an in-plan page re-arms the refile chain
-- site_plans -> site_plan_pages -> pages and the row returns.
--
-- "OPEN" IS `workItemClosedStatuses`, NOT `workItemTerminalStatuses` (§15,
-- work_items_common.go:40-70). `unresolved` and `failed` are OPEN by owner ruling
-- RFC_010. **Two of this page's four open items are `unresolved`** — the terminal
-- list, which is the one whose name sounds right, would have skipped half of them.
--
-- EXACT REVERT for the mutations that are not a status flip (captured from the live
-- rows immediately before the DELETE):
--
--   INSERT INTO site_plan_pages (id,plan_id,name,role,slug,url,parent_section,in_header,in_footer,nav_order,created_at,title,meta_description,nav_label) VALUES ('8bf82bfa-26c8-4eb1-90ff-fc1a059c89ad'::uuid,'7a40a0f9-a1cd-4259-8654-cc0922e942aa'::uuid,'gripper-cycle-time-estimator','content','gripper-cycle-time-estimator','/gripper-cycle-time-estimator.html',NULL,false,false,100,'2026-07-08 15:44:27.411951+00'::timestamptz,'Gripper Cycle Time Estimator — Overview | Robot-Hands.com',NULL,NULL);
--   INSERT INTO site_plan_sections (id,plan_id,page_name,ordering,component_name,component_version_id,palette_id,layout_id,typography_set_id,created_at,assigned_fact_ids) VALUES ('234e7b62-a7d8-4ccc-ad29-5a8e067ce558'::uuid,'7a40a0f9-a1cd-4259-8654-cc0922e942aa'::uuid,'gripper-cycle-time-estimator',0,'hero',NULL,NULL,NULL,NULL,'2026-07-08 15:44:27.411951+00'::timestamptz,NULL);
--   INSERT INTO site_plan_sections (id,plan_id,page_name,ordering,component_name,component_version_id,palette_id,layout_id,typography_set_id,created_at,assigned_fact_ids) VALUES ('891434cb-11a4-4fc0-8cb7-e7a07131f664'::uuid,'7a40a0f9-a1cd-4259-8654-cc0922e942aa'::uuid,'gripper-cycle-time-estimator',1,'generic-text-block',NULL,NULL,NULL,NULL,'2026-07-08 15:44:27.411951+00'::timestamptz,NULL);
--   INSERT INTO site_plan_sections (id,plan_id,page_name,ordering,component_name,component_version_id,palette_id,layout_id,typography_set_id,created_at,assigned_fact_ids) VALUES ('757ddd52-cd58-48fd-a209-8e2d20f55b3e'::uuid,'7a40a0f9-a1cd-4259-8654-cc0922e942aa'::uuid,'gripper-cycle-time-estimator',2,'call-to-action',NULL,NULL,NULL,NULL,'2026-07-08 15:44:27.411951+00'::timestamptz,NULL);
--   UPDATE pages SET status='active' WHERE id='abae9dc9-8f3b-4e3f-97f7-b31439b29e1b';
--   -- work items, exact prior statuses (printed again by this file before the UPDATE):
--   UPDATE site_work_items SET status='unresolved',        handled_by=NULL, resolution_path=NULL WHERE id='7123351c-6640-44ec-9205-96cffaa45831';
--   UPDATE site_work_items SET status='needs_human_review',handled_by=NULL, resolution_path=NULL WHERE id='bf8599fe-e983-470d-a4be-62d7d674123e';
--   UPDATE site_work_items SET status='unresolved',        handled_by=NULL, resolution_path=NULL WHERE id='aef5244c-b058-4615-9688-d2bd1ddb8be5';
--   UPDATE site_work_items SET status='needs_human_review',handled_by=NULL, resolution_path=NULL WHERE id='7315e4d5-6519-42c9-ae86-606ddc3969e9';
--
-- NOTE the page_components rows are NOT deleted: the bare page's own copy survives in
-- the DB after archiving, so the retirement stays recoverable. Only the deployed FILE
-- goes, and that is step 6's dispatch, not this file.
--
-- ⚠ THE VERIFY BLOCK IS `DO`/`RAISE`, NOT A LIST OF SELECTs. `ON_ERROR_STOP` ignores a
-- non-empty result set, so a block of SELECTs cannot stop the COMMIT (LANDMINES /
-- RFC_006). Every assertion below aborts the transaction. Induced before trusting it.

BEGIN;

-- prior statuses, for the record and for the revert above
SELECT id, item_type, status, spec->>'page_name' AS page
FROM site_work_items
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND spec->>'page_name'='gripper-cycle-time-estimator'
  AND status NOT IN ('complete','cancelled','rejected')
ORDER BY item_type, id;

-- ---------------------------------------------------------------- pre-flight
DO $pre$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE id='abae9dc9-8f3b-4e3f-97f7-b31439b29e1b'
     AND site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND status='active';
  IF n <> 1 THEN RAISE EXCEPTION 'PRE-FLIGHT: loser is not the active robot-hands page it was measured as (got %)', n; END IF;

  SELECT count(*) INTO n FROM pages
   WHERE id='acc27598-28c6-4950-9ec5-61b1a9f5061d' AND status='active';
  IF n <> 1 THEN RAISE EXCEPTION 'PRE-FLIGHT: survivor is not active — do not retire the loser (got %)', n; END IF;

  -- THE MERGE MUST HAVE LANDED. Retiring first would destroy 1,587 words.
  SELECT count(*) INTO n FROM page_components
   WHERE page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d'
     AND slot_name IN ('generic-text-block','faq');
  IF n <> 2 THEN
    RAISE EXCEPTION 'PRE-FLIGHT: survivor carries % of the 2 merged prose sections — RUN THE MERGE FILE FIRST, retiring now would destroy the prose', n;
  END IF;

  SELECT count(*) INTO n FROM site_plans
   WHERE id='7a40a0f9-a1cd-4259-8654-cc0922e942aa'
     AND site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'PRE-FLIGHT: plan 7a40a0f9 is no longer the current robot-hands plan — re-measure, another lane may have replanned'; END IF;

  SELECT count(*) INTO n FROM site_plan_pages
   WHERE plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND name='tool-gripper-cycle-time-estimator';
  IF n <> 1 THEN RAISE EXCEPTION 'PRE-FLIGHT: survivor is not in the current plan (got %) — stop', n; END IF;

  SELECT count(*) INTO n FROM site_plan_pages
   WHERE plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND name='gripper-cycle-time-estimator';
  IF n <> 1 THEN RAISE EXCEPTION 'PRE-FLIGHT: expected exactly 1 loser plan row, got % — re-measure', n; END IF;
END
$pre$;

-- ------------------------------------------------- step 3: out of the CURRENT plan
DELETE FROM site_plan_sections
 WHERE plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa'
   AND page_name='gripper-cycle-time-estimator';

DELETE FROM site_plan_pages
 WHERE plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa'
   AND name='gripper-cycle-time-estimator';

-- ------------------------------------------------- step 4: cancel open work items
-- Site-scoped. An UNSCOPED filter on spec->>'page_name' hits other sites' items —
-- §11 measured 29 items across four sites for pair 1's name.
UPDATE site_work_items
   SET status='cancelled',
       handled_by='brochure_215_o2_thread',
       resolution_path='cancelled by bugs_open/215 O2 pair 7: page retired as the duplicate of tool-gripper-cycle-time-estimator (owner ruling 2026-08-13, merge-then-retire; prose merged 2026-08-15)'
 WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
   AND spec->>'page_name'='gripper-cycle-time-estimator'
   AND status NOT IN ('complete','cancelled','rejected');

-- ------------------------------------------------- step 5: archive the loser
UPDATE pages SET status='archived', updated_at=now()
 WHERE id='abae9dc9-8f3b-4e3f-97f7-b31439b29e1b';

-- ---------------------------------------------------------------- post-flight
DO $post$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_plan_pages
   WHERE plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND name='gripper-cycle-time-estimator';
  IF n <> 0 THEN RAISE EXCEPTION 'POST: loser still in the current plan (%) — the refile chain would re-create it', n; END IF;

  SELECT count(*) INTO n FROM site_plan_sections
   WHERE plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND page_name='gripper-cycle-time-estimator';
  IF n <> 0 THEN RAISE EXCEPTION 'POST: loser still has % plan section rows', n; END IF;

  -- The survivor must be untouched by all of the above.
  SELECT count(*) INTO n FROM site_plan_pages
   WHERE plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND name='tool-gripper-cycle-time-estimator';
  IF n <> 1 THEN RAISE EXCEPTION 'POST: survivor is no longer in the plan (%) — ABORT', n; END IF;

  SELECT count(*) INTO n FROM pages
   WHERE id='acc27598-28c6-4950-9ec5-61b1a9f5061d' AND status='active';
  IF n <> 1 THEN RAISE EXCEPTION 'POST: survivor is no longer active — ABORT'; END IF;

  SELECT count(*) INTO n FROM page_components
   WHERE page_id='acc27598-28c6-4950-9ec5-61b1a9f5061d';
  IF n <> 3 THEN RAISE EXCEPTION 'POST: survivor has % components, expected 3 (tool + explainer + faq)', n; END IF;

  SELECT count(*) INTO n FROM pages
   WHERE id='abae9dc9-8f3b-4e3f-97f7-b31439b29e1b' AND status='archived';
  IF n <> 1 THEN RAISE EXCEPTION 'POST: loser is not archived'; END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
     AND spec->>'page_name'='gripper-cycle-time-estimator'
     AND status NOT IN ('complete','cancelled','rejected');
  IF n <> 0 THEN RAISE EXCEPTION 'POST: % work items still open on the loser', n; END IF;

  -- The loser's own content must survive in the DB, so the retirement is recoverable.
  SELECT count(*) INTO n FROM page_components
   WHERE page_id='abae9dc9-8f3b-4e3f-97f7-b31439b29e1b';
  IF n <> 5 THEN RAISE EXCEPTION 'POST: loser has % components, expected its 5 to survive archiving', n; END IF;

  RAISE NOTICE 'OK: loser out of plan, work items cancelled, page archived; survivor active, in plan, 3 components; loser content preserved';
END
$post$;

COMMIT;

-- VERIFY
SELECT p.name, p.status,
       (SELECT count(*) FROM site_plan_pages spp
         WHERE spp.plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND spp.name=p.name) AS in_current_plan,
       (SELECT count(*) FROM page_components pc WHERE pc.page_id=p.id) AS comps
FROM pages p
WHERE p.id IN ('abae9dc9-8f3b-4e3f-97f7-b31439b29e1b','acc27598-28c6-4950-9ec5-61b1a9f5061d')
ORDER BY p.name;
