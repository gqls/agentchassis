#!/usr/bin/env bash
# poll_open_critiques.sh — the dispatcher thread's tick query (DISPATCHER_README, this dir).
# Prints open, UNROUTED owner_critique items (and any item whose spec names the
# thread-dispatcher as consumer), oldest first, with the critique text and the exact
# routing-stamp UPDATE to run after routing each one.
# Read-only by itself; the stamp is printed, never executed.
set -euo pipefail

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -X)

"${PSQL[@]}" <<'SQL'
SELECT w.id, s.domain, w.created_at::timestamp(0) AS filed,
       w.summary,
       left(w.spec->>'critique', 2000) AS critique,
       w.spec->>'origin_item_id' AS origin_item
FROM site_work_items w
JOIN sites s ON s.id = w.site_id
WHERE (w.item_type = 'owner_critique' OR w.spec->>'consumer' = 'thread-dispatcher')
  AND w.status NOT IN ('complete','cancelled','rejected')
  AND w.result->>'routed_at' IS NULL
ORDER BY w.created_at;
SQL

cat <<'HOWTO'
-- After routing an item to live thread(s) or queueing it for the owner, stamp it
-- (first stamp wins; a second dispatcher must skip stamped rows):
--   UPDATE site_work_items
--   SET result = COALESCE(result,'{}'::jsonb) || jsonb_build_object(
--         'routed_at', now()::text,
--         'routed_to', '<thread names or OWNER-TO-OPEN:<lane>>',
--         'routed_by', '<your session name>')
--   WHERE id = '<item id>' AND result->>'routed_at' IS NULL;
-- The WHERE guard on routed_at IS the double-route protection — check UPDATE 1.
HOWTO
