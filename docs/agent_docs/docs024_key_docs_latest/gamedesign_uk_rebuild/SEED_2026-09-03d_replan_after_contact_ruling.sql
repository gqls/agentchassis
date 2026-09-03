\set ON_ERROR_STOP on
-- SEED_2026-09-03d — re-plan gamedesign.uk after the owner's midday contact ruling.
--
-- Owner, 2026-09-03 midday: "ok leave the contact page, the email can be
-- gamedesignuk@contactforsales.com. Replan please."
--
-- Preconditions already applied by SEED_2026-09-03c (mission_brief v4, evidence_base v4 with the
-- two contact/email bans removed, address updated in submission + briefing).
--
-- TWO steps, and step 1 is the one that is easy to miss:
--
-- 1. CANCEL work item ac76ec54 ("Build index page", status needs_human_review since 11:03:32Z,
--    blocked on the banned claim the owner has now reversed). It carries item_key
--    'needs_page:index', and `needs_human_review` is NOT in workItemTerminalStatuses
--    (platform/orchestration/actions/work_items_common.go:42 — complete/failed/verified/rejected/
--    wont_fix/unresolved/cancelled). idx_swi_dedup therefore still holds that key, so a re-plan
--    filing a fresh needs_page for index would be SILENTLY DEDUPED AWAY and the homepage would
--    never rebuild. Cancelling returns the slot ('cancelled' joined the closed set in migration
--    157 precisely so a cancelled row does not hold it). The row is preserved, not deleted.
--
-- 2. Enqueue needs_briefing in the domain-strategist's own shape (same as
--    SEED_2026-09-03_enqueue_briefing_chain.sql; the prior briefing_gamedesign.uk row is
--    'complete' and therefore terminal, so the dedup index allows this one). Chain:
--    build-briefing-agent -> needs_site_plan -> build-site-planner -> reconcile -> composition/
--    design/pages/imagery/rerender.
--
-- WHY A RE-PLAN AND NOT A RESCUE OF THE OLD BUILD: plan c920da7a was written 10:40:21Z. Migrations
-- 730 (10:59:59Z) and 731 (~11:03Z) added build-site-planner rule 20 ("NO LATER EDITORIAL PASS
-- RUNS…", instructing 3-6 launch posts on real subjects) off this lane's own evidence. The old
-- plan predates rule 20 by ~20 minutes and has ZERO article pages; only a new plan picks it up.
-- growth_posture='hold' stays on, so evaluate_tools/add_tool still file as records, not builds.

BEGIN;

-- 1. free the dedup slot
UPDATE site_work_items
   SET status = 'cancelled',
       error = 'Cancelled 2026-09-03 by the gamedesign.uk lane: blocked at validate_content on banned claim "contact page", which the owner reversed the same day (contact page stays, address gamedesignuk@contactforsales.com; SEED_2026-09-03c). Superseded by the re-plan enqueued in SEED_2026-09-03d. Cancelled rather than retried because plan c920da7a predates build-site-planner rule 20 (migrations 730/731) and carries zero article pages.'
 WHERE id = 'ac76ec54-8f04-455f-8018-03bc5834ac96'
   AND status = 'needs_human_review';

DO $g$
DECLARE held int;
BEGIN
  SELECT count(*) INTO held FROM site_work_items
   WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414'
     AND item_key = 'needs_page:index'
     AND status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled');
  IF held <> 0 THEN
    RAISE EXCEPTION '09-03d REFUSED: % non-terminal row(s) still hold item_key needs_page:index — a re-plan would be deduped away', held;
  END IF;
END $g$;

-- 2. enqueue the chain
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
VALUES ('8f17eb73-fc74-4718-8371-b3125bc4e414', 'domain-strategist', 'build', 'needs_briefing', 'high',
  'Re-brief + re-plan after the owner''s 2026-09-03 midday contact ruling (contact page stays, new address) and build-site-planner rule 20 (migrations 730/731, launch posts) — plan c920da7a predates both',
  '{"reason":"owner_ruling_contact_page_stays_and_replan","owner_ruling_at":"2026-09-03T11:30:00Z","supersedes_plan":"c920da7a-ac16-4ee1-88fe-9366dcdd9d72","preconditions":"SEED_2026-09-03c (mission_brief v4, evidence_base v4, address updated)","expects":"article pages under /articles/ via rule 20; contact page retained with gamedesignuk@contactforsales.com","lane":"gamedesign_uk_rebuild"}'::jsonb,
  10, 'build-briefing-agent', 'triaged', 'gamedesign_uk_rebuild lane 2026-09-03', 'briefing_gamedesign.uk')
RETURNING id, item_type, status, created_at::timestamp(0);

COMMIT;

SELECT id, item_type, status, item_key, created_at::timestamp(0)
  FROM site_work_items
 WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414'
   AND item_type IN ('needs_briefing','needs_page')
   AND created_at > '2026-09-03 10:00'
 ORDER BY created_at;
