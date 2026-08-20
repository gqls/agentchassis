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
