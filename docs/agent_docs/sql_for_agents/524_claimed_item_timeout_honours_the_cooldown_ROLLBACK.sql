-- ROLLBACK for 524 — removes the cooldown stamp from claimed-item-timeout's reset
-- CTE, restoring the pre-341 behaviour (a timed-out item is immediately re-claimable).
--
-- Anchored the same way as the forward file: it strips the exact block it added and
-- refuses if that block is not present exactly once, so it cannot damage a pre_query
-- another lane has since edited.
--
-- Consequence of rolling back while the Go contract is live: the two ladders diverge
-- again on the cooldown, which is bugs_open/341's original finding. That is the point
-- of a rollback — but it is silent, so re-open 341 if you run this.

BEGIN;

DO $do$
DECLARE
  v_pq    text;
  v_block text;
  v_hits  int;
BEGIN
  SELECT pre_query INTO v_pq FROM scheduled_tasks WHERE name = 'claimed-item-timeout';
  IF v_pq IS NULL THEN
    RAISE EXCEPTION 'rollback 524: claimed-item-timeout not found';
  END IF;
  IF v_pq NOT LIKE '%retry_after = CASE%' THEN
    RAISE NOTICE 'rollback 524: the stamp is already absent — nothing to do';
    RETURN;
  END IF;

  v_block := substring(v_pq from position('        retry_after = CASE' in v_pq)
                       for (position(E'        END\n    WHERE status = ''claimed''' in v_pq)
                            - position('        retry_after = CASE' in v_pq) + length(E'        END\n')));

  SELECT count(*) INTO v_hits FROM regexp_matches(v_pq, 'retry_after = CASE', 'g');
  IF v_hits <> 1 THEN
    RAISE EXCEPTION 'rollback 524: ABORTING — found % retry_after blocks, expected 1. Re-read the live column and re-derive; do NOT force', v_hits;
  END IF;

  UPDATE scheduled_tasks SET pre_query = replace(v_pq, v_block, ''), updated_at = now()
   WHERE name = 'claimed-item-timeout';
END
$do$;

DO $do$
BEGIN
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'claimed-item-timeout' AND pre_query LIKE '%retry_after%') <> 0 THEN
    RAISE EXCEPTION 'rollback 524: retry_after still present in the pre_query';
  END IF;
  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'claimed-item-timeout' AND pre_query LIKE '%completed_by_evidence%') <> 1 THEN
    RAISE EXCEPTION 'rollback 524: the strip removed more than the stamp';
  END IF;
END
$do$;

COMMIT;
