-- bugs_open/215 O2 — PAIR 5, robot-hands.com `gripper-payload-calculator`
-- Runbook steps 3, 4 and 5 in ONE transaction. Step 6 (retraction) is a
-- dispatch, not SQL, and follows this file. Step 7 does not exist (there is no
-- redirect mechanism — RUNBOOK ⚠ CORRECTION 2026-08-14).
--
-- OWNER RULING 2026-08-13 (DECISION_INPUT_2026-08-12_seven_twin_pairs.md):
--   keep `tool-gripper-payload-calculator`, retire bare `gripper-payload-calculator`.
--   Unchanged by the 2026-08-14 reversal, which touched pairs 3+4 only.
--
-- WHY THIS PAIR IS SAFE TO EXECUTE NOW [MEASURED 2026-08-14 ~16:40Z, read-only]:
--   * the loser has ZERO editorial inbound referrers and ZERO active nav rows —
--     reproduced from retract_page_graph.go's own three census queries, run in
--     the same breath over pair 6 (`matchmatrix`), which returned 4 editorial +
--     1 nav. That non-zero is the positive control: the query CAN match on this
--     site, so this pair's zero is not an inert query. (link_registry is EMPTY
--     fleet-wide and can never be the instrument — LANDMINES 2026-08-14.)
--   * served artefacts: loser 23,015 b / survivor 34,157 b / 404 control 2,886 b.
--   * the survivor is `rebuild_policy='owned'`, so PBP-036 already guards it.
--
-- ORDER MATTERS, AND STEP 3 IS MANDATORY HERE: the current plan
-- (7a40a0f9-…, is_current) names BOTH sides. Archiving an in-plan page re-arms
-- the refile chain site_plans -> site_plan_pages -> pages and the row returns.
--
-- EXACT REVERT for the two mutations this file makes that are not a status flip
-- (captured from the live rows immediately before the DELETE):
--
--   INSERT INTO site_plan_pages (id,plan_id,name,role,slug,url,parent_section,in_header,in_footer,nav_order,created_at,title,meta_description,nav_label) VALUES ('927d67ce-2473-41b1-82a6-d05787b7d846'::uuid,'7a40a0f9-a1cd-4259-8654-cc0922e942aa'::uuid,'gripper-payload-calculator','content','gripper-payload-calculator','/gripper-payload-calculator.html',NULL,false,false,100,'2026-07-08 15:44:27.411951+00'::timestamptz,'Gripper Payload Calculator — Overview | Robot-Hands.com',NULL,NULL);
--   INSERT INTO site_plan_sections (id,plan_id,page_name,ordering,component_name,component_version_id,palette_id,layout_id,typography_set_id,created_at,assigned_fact_ids) VALUES ('b9911372-444d-4dce-b46d-262237f38a4a'::uuid,'7a40a0f9-a1cd-4259-8654-cc0922e942aa'::uuid,'gripper-payload-calculator',0,'hero',NULL,NULL,NULL,NULL,'2026-07-08 15:44:27.411951+00'::timestamptz,NULL);
--   INSERT INTO site_plan_sections (id,plan_id,page_name,ordering,component_name,component_version_id,palette_id,layout_id,typography_set_id,created_at,assigned_fact_ids) VALUES ('1f298b56-b42b-4b5d-9292-3de2cd74ae24'::uuid,'7a40a0f9-a1cd-4259-8654-cc0922e942aa'::uuid,'gripper-payload-calculator',1,'generic-text-block',NULL,NULL,NULL,NULL,'2026-07-08 15:44:27.411951+00'::timestamptz,NULL);
--   INSERT INTO site_plan_sections (id,plan_id,page_name,ordering,component_name,component_version_id,palette_id,layout_id,typography_set_id,created_at,assigned_fact_ids) VALUES ('e1b944e5-26a0-4be9-b0f1-49fbbe340202'::uuid,'7a40a0f9-a1cd-4259-8654-cc0922e942aa'::uuid,'gripper-payload-calculator',2,'call-to-action',NULL,NULL,NULL,NULL,'2026-07-08 15:44:27.411951+00'::timestamptz,NULL);
--   UPDATE pages SET status='active' WHERE id='48d52965-4e63-4ee6-9215-1bb578ea06b6';
--   -- (work items: their prior statuses are printed by this script before the UPDATE)
--
-- ⚠ THE VERIFY BLOCK IS `DO`/`RAISE`, NOT A LIST OF SELECTs. `ON_ERROR_STOP`
-- ignores a non-empty result set, so a block of SELECTs cannot stop the COMMIT
-- (LANDMINES / RFC_006). Every assertion below aborts the transaction.

\set ON_ERROR_STOP on
\set site   '00ff3af5-dad8-4770-9f70-3edc267a3c92'
\set plan   '7a40a0f9-a1cd-4259-8654-cc0922e942aa'
\set loser  '48d52965-4e63-4ee6-9215-1bb578ea06b6'
\set winner '40b87756-3176-4296-861c-526c5496f5a5'

BEGIN;

-- ---------------------------------------------------------------- pre-flight
-- Abort before any write if the world is not the one measured above.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE id = '48d52965-4e63-4ee6-9215-1bb578ea06b6' AND status = 'active'
     AND site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name = 'gripper-payload-calculator';
  IF n <> 1 THEN RAISE EXCEPTION 'PRE-FLIGHT: loser is not the active robot-hands page it was measured as (got %)', n; END IF;

  SELECT count(*) INTO n FROM pages
   WHERE id = '40b87756-3176-4296-861c-526c5496f5a5' AND status = 'active'
     AND name = 'tool-gripper-payload-calculator';
  IF n <> 1 THEN RAISE EXCEPTION 'PRE-FLIGHT: survivor is not active — do not retire the loser (got %)', n; END IF;

  SELECT count(*) INTO n FROM site_plans
   WHERE id = '7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND is_current
     AND site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92';
  IF n <> 1 THEN RAISE EXCEPTION 'PRE-FLIGHT: plan 7a40a0f9 is no longer the current robot-hands plan — re-measure, another lane may have replanned'; END IF;

  -- The survivor must be IN the plan we are about to edit, or step 3 removes
  -- the only entry and the site loses the page at the next refile.
  SELECT count(*) INTO n FROM site_plan_pages
   WHERE plan_id = '7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND name = 'tool-gripper-payload-calculator';
  IF n <> 1 THEN RAISE EXCEPTION 'PRE-FLIGHT: survivor is not in the current plan (got %) — stop', n; END IF;
END $$;

-- What the work-item step is about to touch, printed before it changes.
-- OPEN is workItemClosedStatuses, NOT workItemTerminalStatuses: `unresolved`
-- and `failed` are OPEN (work_items_common.go:57-70, owner ruling RFC_010
-- 2026-08-02 "Decision 2: `unresolved` is OPEN").
SELECT id, item_type, status, handler_agent, created_at
  FROM site_work_items
 WHERE site_id = :'site'::uuid
   AND (page_id = :'loser'::uuid
        OR spec->>'page_name' IN ('gripper-payload-calculator','/gripper-payload-calculator.html'))
   AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
 ORDER BY created_at;

-- ------------------------------------------------- step 3: out of the PLAN
DELETE FROM site_plan_sections
 WHERE plan_id = :'plan'::uuid AND page_name = 'gripper-payload-calculator';

DELETE FROM site_plan_pages
 WHERE plan_id = :'plan'::uuid AND name = 'gripper-payload-calculator';

-- ------------------------------------------ step 4: cancel open work items
-- Site-scoped. §12's lesson: the same page_name exists on other sites, and an
-- unscoped predicate cancelled 29 items across four of them in rehearsal.
UPDATE site_work_items
   SET status = 'cancelled',
       handled_by = 'brochure_215_o2_thread',
       resolution_path = 'bugs_open/215 O2 pair 5: retired as duplicate of tool-gripper-payload-calculator (owner ruling 2026-08-13). Cancelled so the queue stops targeting a retired page.',
       updated_at = now()
 WHERE site_id = :'site'::uuid
   AND (page_id = :'loser'::uuid
        OR spec->>'page_name' IN ('gripper-payload-calculator','/gripper-payload-calculator.html'))
   AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled');

-- ------------------------------------------------- step 5: archive the row
-- NEVER `DELETE`: three FKs onto `pages` are NO ACTION
-- (link_registry.target_page_id, redirects.source_page_id,
-- page_component_history.page_id) so a delete can fail rather than cascade.
-- `archived` IS this platform's delete for a page, and it is hand-run by
-- design — no Go writer of status='archived' exists
-- (retract_page_deployment_action.go's own header).
UPDATE pages
   SET status = 'archived', updated_at = now()
 WHERE id = :'loser'::uuid AND status = 'active';

-- --------------------------------------------------------------- verify
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages WHERE id='48d52965-4e63-4ee6-9215-1bb578ea06b6' AND status='archived';
  IF n <> 1 THEN RAISE EXCEPTION 'VERIFY: loser not archived'; END IF;

  -- The whole point of the transaction: the survivor must be untouched.
  SELECT count(*) INTO n FROM pages WHERE id='40b87756-3176-4296-861c-526c5496f5a5' AND status='active';
  IF n <> 1 THEN RAISE EXCEPTION 'VERIFY: survivor is no longer active — ABORT'; END IF;

  SELECT count(*) INTO n FROM site_plan_pages
   WHERE plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND name='gripper-payload-calculator';
  IF n <> 0 THEN RAISE EXCEPTION 'VERIFY: loser still in plan (%)', n; END IF;

  SELECT count(*) INTO n FROM site_plan_sections
   WHERE plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND page_name='gripper-payload-calculator';
  IF n <> 0 THEN RAISE EXCEPTION 'VERIFY: loser sections still in plan (%)', n; END IF;

  SELECT count(*) INTO n FROM site_plan_pages
   WHERE plan_id='7a40a0f9-a1cd-4259-8654-cc0922e942aa' AND name='tool-gripper-payload-calculator';
  IF n <> 1 THEN RAISE EXCEPTION 'VERIFY: survivor plan row lost (%) — ABORT', n; END IF;

  -- Nothing belonging to the OTHER two robot-hands pairs may have moved.
  SELECT count(*) INTO n FROM site_work_items
   WHERE handled_by='brochure_215_o2_thread' AND updated_at > now() - interval '1 minute'
     AND NOT (page_id = '48d52965-4e63-4ee6-9215-1bb578ea06b6'
              OR spec->>'page_name' IN ('gripper-payload-calculator','/gripper-payload-calculator.html'));
  IF n <> 0 THEN RAISE EXCEPTION 'VERIFY: % work item(s) outside pair 5 were cancelled — ABORT', n; END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
     AND (page_id='48d52965-4e63-4ee6-9215-1bb578ea06b6'
          OR spec->>'page_name' IN ('gripper-payload-calculator','/gripper-payload-calculator.html'))
     AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled');
  IF n <> 0 THEN RAISE EXCEPTION 'VERIFY: % open work item(s) still target the loser', n; END IF;
END $$;

COMMIT;

-- Post-state, for the record.
SELECT p.name, p.status, p.build_status, p.deployed_at
  FROM pages p WHERE p.id IN (:'loser'::uuid, :'winner'::uuid) ORDER BY p.name;
SELECT id, item_type, status, handled_by FROM site_work_items
 WHERE site_id = :'site'::uuid
   AND (page_id = :'loser'::uuid
        OR spec->>'page_name' IN ('gripper-payload-calculator','/gripper-payload-calculator.html'))
 ORDER BY created_at;
