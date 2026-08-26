-- 653_content_feed_due_lookahead_HOLD.sql
--
-- bugs_open/410 (the next_fetch_at phase lock — resolve by SLUG, the number
-- also names an unrelated scan-loss case): the site-level half of the
-- half-cadence due look-ahead.
--
-- WHY: both writers of content_sources.next_fetch_at stamp it at FETCH time
-- (NOW() + fetch_interval), so a source whose interval equals the 6 h trigger
-- cadence comes due seconds AFTER the next pass fires and is served every
-- OTHER pass — 10 of 12 news sites measured on a 12 h cadence under a 6 h
-- label on 2026-08-26. The fix: a source is due for a pass when it will fall
-- due within HALF the trigger cadence — serve on the nearest tick, not the
-- tick after. The look-ahead reads the live cadence from scheduled_tasks and
-- falls back to 3 hours (half of the 21600 s cadence this file was written
-- against; the guard below asserts that equality) if the task row is renamed.
--
-- ⚠ _HOLD — APPLY BY HAND, ONLY AFTER THE CHASSIS ROLL. This is the SITE
-- admission half; the SOURCE-level half is Go (feedSourceDuePredicate in
-- platform/orchestration/actions/feed_due_lookahead.go, same bug, committed
-- alongside this file) and is inert until an agent-chassis image rolls.
-- Applied config-first, this migration admits sites whose sources the
-- un-rolled dispatcher then refuses — no-op COMPLETED runs that consume cap
-- slots (LIMIT 10 / max_iterations 10, migration 556) and poison any
-- before/after cadence census. Verify the roll first:
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <the commit shipping feed_due_lookahead.go> <that stamp>
--
-- EXPECTED AFTERWARDS, so nobody files it as a regression: every 6 h-only
-- news site becomes due at EVERY pass, so ~12 sites contend for the cap of 10
-- every 6 hours and LCO-009 / --capped-schedule-ordering will report cap hits
-- on most passes. That is the demand the caps were sized against in migration
-- 556 finally arriving (fetch volume returns to the DESIGNED cadence, roughly
-- double today's phase-locked volume). The 554 due_at fair ordering bounds the
-- overflow: the two sites cut this pass carry the oldest due_at next pass.

SELECT snapshot_agent('content-feed-trigger',
       'migration 653: pre due-lookahead (bugs_open/410 phase lock)');

BEGIN;

-- Guard 1: the fallback literal below (interval '3 hours') is half the live
-- cadence. If the cadence has changed, this file is stale — recompute the
-- fallback (and re-read bugs_open/410 §6) rather than forcing this through.
DO $$
DECLARE secs integer;
BEGIN
    SELECT interval_seconds INTO secs FROM scheduled_tasks WHERE name = 'content-feed-refresh';
    IF secs IS NULL THEN
        RAISE EXCEPTION 'MIGRATION 653: scheduled_tasks has no content-feed-refresh row — the look-ahead''s cadence source is gone; do not apply blind.';
    END IF;
    IF secs IS DISTINCT FROM 21600 THEN
        RAISE EXCEPTION 'MIGRATION 653: content-feed-refresh cadence is % s, not the 21600 s this file''s 3-hour fallback was derived from — recompute before applying.', secs;
    END IF;
END $$;

-- Guard 2: the live query is exactly the one this file was written against
-- (migration 556's post-image, re-read live 2026-08-26). A mismatch means
-- another session changed it — merge, do not overwrite.
DO $$
DECLARE q text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query' INTO q
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    IF q IS DISTINCT FROM $pre$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10$pre$ THEN
        RAISE EXCEPTION 'MIGRATION 653: find_news_sites is not the 556 post-image this file expects — someone changed it since 2026-08-26; refusing to overwrite their change.';
    END IF;
END $$;

-- The only change: the due arm of the source EXISTS gains the look-ahead. The
-- COALESCE subquery is byte-identical to feedDueLookaheadSQL in
-- feed_due_lookahead.go (the cs. prefix on the column is the only difference
-- between the two layers' predicates) — keep them in step.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,find_news_sites,config,query}',
        to_jsonb($post$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW() + COALESCE((SELECT make_interval(secs => interval_seconds / 2.0) FROM scheduled_tasks WHERE name = 'content-feed-refresh'), interval '3 hours'))))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10$post$::text),
        false),
    updated_at = now()
WHERE type = 'content-feed-trigger' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- Post-verify: the look-ahead is in, and neither the 554 fairness ordering nor
-- the 556 capacity moved with it.
DO $$
DECLARE q text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query' INTO q
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF q NOT LIKE '%make_interval(secs => interval_seconds / 2.0)%' THEN
        RAISE EXCEPTION 'MIGRATION 653: the look-ahead did not land.';
    END IF;
    IF q NOT LIKE '%ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10' THEN
        RAISE EXCEPTION 'MIGRATION 653: the 554 ordering / 556 capacity tail changed — refusing.';
    END IF;
END $$;

COMMIT;
