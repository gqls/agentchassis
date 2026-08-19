-- bugs_open/029 — rebuilds the whole wedge population from awaited_requests in ONE pass.
-- orchestration_states has pruned these rows (~26h); awaited_requests retains 7 days.
-- Returns 20 rows for 2026-08-17, every one with next_call_registered = false.
WITH err AS (
  SELECT orchestration_id,
         (regexp_match(step_name,'iter_([0-9]+)_call_handler'))[1]::int AS n
    FROM awaited_requests
   WHERE step_name ~ '^process_item_iter_[0-9]+_call_handler$'
     AND retry_version >= 3 AND status = 'error'
), sp AS (
  SELECT orchestration_id,
         (regexp_match(step_name,'iter_([0-9]+)_spawn_handler'))[1]::int AS n,
         count(*) AS spawn_rows, min(sent_at) AS first_spawn, max(sent_at) AS last_spawn
    FROM awaited_requests
   WHERE step_name ~ '^process_item_iter_[0-9]+_spawn_handler$'
   GROUP BY 1,2
), nxt AS (
  SELECT DISTINCT orchestration_id,
         (regexp_match(step_name,'iter_([0-9]+)_call_handler'))[1]::int AS n
    FROM awaited_requests
   WHERE step_name ~ '^process_item_iter_[0-9]+_call_handler$'
)
SELECT e.orchestration_id, e.n AS err_iter, s.spawn_rows,
       s.first_spawn, s.last_spawn,
       (x.orchestration_id IS NOT NULL) AS next_call_registered
  FROM err e
  JOIN sp  s ON s.orchestration_id = e.orchestration_id AND s.n = e.n + 1
  LEFT JOIN nxt x ON x.orchestration_id = e.orchestration_id AND x.n = e.n + 1
 ORDER BY s.first_spawn;
