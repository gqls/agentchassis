-- ROLLBACK for 506 — restores both dispatch reads to their pre-307 form
-- (no retry_after clause). Apply this BEFORE 505's rollback: dropping the
-- column while a pre_query still names it fails that task at every tick, and a
-- build-pipeline-trigger that errors dispatches nothing while reporting nothing.
--
-- Consequence of rolling this back while the chassis ladder is live: the ladder
-- keeps stamping retry_after and nothing honours it, so the backoff becomes
-- decorative and three attempts can again land inside one outage. That is the
-- pre-307 behaviour, which is the point of a rollback — but it is silent, so
-- disarm the ladder too with DISABLE_WORK_ITEM_RETRY_BACKOFF=1 if the intent is
-- to stop the backoff rather than only the reads.

BEGIN;

-- ── PRE-STATE GATE (added 2026-08-20 after the council's `debug_historian` seat
-- objected, corr 4cdec68b: the forward migration overwrote an embedded query
-- string with jsonb_set and no pre-state check, in a table with no updated_at
-- trigger and documented drift history — migrations 052/213/285).
--
-- The objection was right as discipline. It did not bite on the forward file:
-- verified 2026-08-20 that mine was the only write to that row (updated_at
-- 14:14:08Z) and no other lane's migration touched that agent in the window. But
-- a ROLLBACK is the run MOST likely to happen under time pressure, months later,
-- by someone who did not write it — so it gates rather than trusts.
--
-- If either RAISE fires, another lane has edited these since 506 applied.
-- Re-derive this rollback against the CURRENT text; do NOT force. Overwriting is
-- how one lane silently reverts another.
DO $do$
DECLARE
  v_sel_md5 text;
  v_pq_md5  text;
BEGIN
  SELECT md5(default_config #>> '{workflow,steps,find_dispatchable_site,config,query}')
    INTO v_sel_md5 FROM agent_definitions
   WHERE type = 'build-pipeline-trigger' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  SELECT md5(pre_query) INTO v_pq_md5 FROM scheduled_tasks WHERE name = 'build-pipeline-trigger';

  IF v_sel_md5 IS DISTINCT FROM 'e8bf1939535bdecfff627860ef4e2d50' THEN
    RAISE EXCEPTION 'rollback 506 ABORTING: find_dispatchable_site is not what 506 produced (expected md5 e8bf1939535bdecfff627860ef4e2d50, found %). Another lane has edited it since 2026-08-20 14:14Z — re-derive against the live text, do NOT force', v_sel_md5;
  END IF;
  IF v_pq_md5 IS DISTINCT FROM '200246f7ede3e33b14be2fc064efa7da' THEN
    RAISE EXCEPTION 'rollback 506 ABORTING: build-pipeline-trigger.pre_query is not what 506 produced (expected md5 200246f7ede3e33b14be2fc064efa7da, found %). Same instruction', v_pq_md5;
  END IF;
END
$do$;

UPDATE scheduled_tasks
   SET pre_query = $PQ$SELECT COUNT(*)::text as pending_sites
FROM sites s
WHERE s.locked_at IS NULL
  AND EXISTS (
    SELECT 1 FROM site_work_items wi
    WHERE wi.site_id = s.id
      AND wi.status = 'triaged'
      AND wi.pipeline = 'build'
      AND wi.attempt_count < wi.max_attempts
)
HAVING COUNT(*) > 0$PQ$,
       updated_at = now()
 WHERE name = 'build-pipeline-trigger';

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,find_dispatchable_site,config,query}',
         to_jsonb($Q$SELECT wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE s.locked_at IS NULL AND wi.status IN ('triaged', 'approved') AND wi.attempt_count < wi.max_attempts AND (COALESCE(wi.approval_mode, 'auto') = 'auto' OR wi.status = 'approved') AND (wi.depends_on IS NULL OR NOT EXISTS (SELECT 1 FROM unnest(wi.depends_on) dep_id WHERE dep_id NOT IN (SELECT id FROM site_work_items WHERE site_id = wi.site_id AND status IN ('complete', 'verified')))) AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = 'claimed') ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1$Q$::text)
       ),
       updated_at = now()
 WHERE type = 'build-pipeline-trigger'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $do$
BEGIN
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'build-pipeline-trigger' AND pre_query LIKE '%retry_after%') <> 0 THEN
    RAISE EXCEPTION 'rollback 506: pre_query still names retry_after';
  END IF;
  IF (SELECT count(*) FROM agent_definitions
       WHERE type = 'build-pipeline-trigger' AND is_active
         AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
         AND (default_config #>> '{workflow,steps,find_dispatchable_site,config,query}') LIKE '%retry_after%') <> 0 THEN
    RAISE EXCEPTION 'rollback 506: find_dispatchable_site still names retry_after';
  END IF;
END
$do$;

COMMIT;
