#!/bin/bash
# bugs_open/136 witness poller. Every 3s for up to 35 min (the generic lane has
# been measured at 29 min publish->run under load):
#   - capture any agent-alias-witness-136-* pod's full log (overwrite with the
#     latest capture; the last one before pod deletion is the keeper)
#   - check for the witness work-item row; on sight, print it and keep polling
#     the pod for a further 30s to catch the log, then exit.
set -u
NS=ai-persona-system
S="$(dirname "$0")"
OUT="$S/witness_136_result.log"
DEADLINE=$(( $(date +%s) + 2100 ))
ROW_SEEN=""

psqlq() {
  kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A 2>/dev/null
}

while [ "$(date +%s)" -lt "$DEADLINE" ]; do
  for P in $(kubectl -n "$NS" get pods -o name 2>/dev/null | grep 'alias-witness-136' | sed 's|pod/||'); do
    kubectl -n "$NS" logs "$P" --tail=-1 > "$S/witness_pod_${P}.log" 2>/dev/null
    echo "$(date -u +%FT%TZ) captured log of $P ($(wc -c < "$S/witness_pod_${P}.log") bytes)" >> "$OUT"
  done
  if [ -z "$ROW_SEEN" ]; then
    ROW=$(printf '%s' "SELECT id||' | '||pipeline||' | '||status||' | '||item_key||' | '||created_at FROM site_work_items WHERE item_type='alias_witness_136';" | psqlq)
    if [ -n "$ROW" ]; then
      ROW_SEEN=yes
      { echo "=== WITNESS ROW at $(date -u +%FT%TZ) ==="; echo "$ROW"; } >> "$OUT"
      DEADLINE=$(( $(date +%s) + 30 ))
    fi
  fi
  sleep 3
done
echo "$(date -u +%FT%TZ) poller done (row_seen=${ROW_SEEN:-no})" >> "$OUT"
