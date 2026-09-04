\set ON_ERROR_STOP on
-- SEED_2026-09-04 — re-plan gamedesign.uk now that bugs_open/463 is LIVE.
--
-- 463 fixed BOTH halves (9b540c2e6): Pass C no longer deletes new section children, and the
-- write path no longer relocates them to /blog/. Confirmed live 2026-09-04 10:53Z:
--   git merge-base --is-ancestor 9b540c2e6 <live agent-chassis stamp 239ab362> -> yes.
-- No site anywhere has been re-planned since the fix landed (checked: 0 site_plans rows with
-- created_at > 2026-09-03 17:52), so this is the first live test.
--
-- Prior plan 005fb393 (2026-09-03 14:15:25Z) is PRE-fix and still has zero article pages —
-- that is the state 463 explains, not a fresh symptom.
--
-- Verification owed on THIS run, per the 463 landmine entry (LANDMINES.md
-- "Comparing pages by their FIRST PATH SEGMENT..."):
--   1. proposed = survived at the orchestration step boundary (plan_site vs validate_plan page
--      counts) — a Pass-C-only fix passes this even if the write-path half regressed.
--   2. the new blog-post pages land at /articles/<slug>.html, NOT /blog/<slug>.html — the
--      discriminator a served-page check cannot make on its own.
-- Both must be read from the orchestration row and site_plan_pages once this completes; this
-- seed only enqueues the chain.
BEGIN;

INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
VALUES ('8f17eb73-fc74-4718-8371-b3125bc4e414', 'domain-strategist', 'build', 'needs_briefing', 'high',
  'Re-brief + re-plan now that bugs_open/463 is live (Pass C fix + write-path parent_section fix, 9b540c2e6) — plan 005fb393 predates it and has zero article pages',
  '{"reason":"bug_463_rolled_replan","fix_commit":"9b540c2e6","fix_confirmed_live_at":"2026-09-04T10:53:00Z","supersedes_plan":"005fb393-da95-4f87-a009-86ae11ce06c5","expects":"blog-post pages under /articles/, NOT /blog/","lane":"gamedesign_uk_rebuild"}'::jsonb,
  10, 'build-briefing-agent', 'triaged', 'gamedesign_uk_rebuild lane 2026-09-04', 'briefing_gamedesign.uk')
RETURNING id, item_type, status, created_at::timestamp(0);

COMMIT;
