#!/bin/bash
# bugs_open/234 strict-witness poller, v2.
#
# v1 LESSON (2026-08-10): v1 captured pod logs to files and grepped the FILES in
# a later statement — the pod's deletion raced the two reads, so the poller
# recorded "REJECTION SEEN" while the evidence lines themselves were lost. A
# marker without its evidence is not proof on this estate. v2 greps the SAME
# capture it just took, in the same iteration, and appends matched lines to the
# result log immediately; every capture is also preserved append-only
# (per-iteration filenames, never overwritten, never truncated by a failed
# re-read).
set -u
NS=ai-persona-system
S="$(cd "$(dirname "$0")" && pwd)"
OUT="$S/witness_234_result.log"
DEADLINE=$(( $(date +%s) + 2100 ))
REJECT_SEEN=""
N=0

psqlq() {
  kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A 2>/dev/null
}

while [ "$(date +%s)" -lt "$DEADLINE" ]; do
  N=$((N+1))
  for P in $(kubectl -n "$NS" get pods -o name 2>/dev/null | grep 'strict-witness-234' | sed 's|pod/||'); do
    CAP="$S/witness_pod_${P}_iter${N}.log"
    if kubectl -n "$NS" logs "$P" --tail=-1 > "$CAP" 2>/dev/null && [ -s "$CAP" ]; then
      HITS=$(grep -m4 "Invalid workflow configuration\|zzz_strict_witness_234\|contract as complete" "$CAP" 2>/dev/null || true)
      if [ -n "$HITS" ] && [ -z "$REJECT_SEEN" ]; then
        REJECT_SEEN=yes
        { echo "=== REJECTION SEEN at $(date -u +%FT%TZ) in $P (iter $N) ==="
          printf '%s\n' "$HITS"
        } >> "$OUT"
        cp "$CAP" "$S/witness_234_evidence.log"
        DEADLINE=$(( $(date +%s) + 30 ))
      fi
    else
      rm -f "$CAP"   # failed or empty capture — keep only real ones
    fi
  done
  ROW=$(printf '%s' "SELECT id||' | '||status||' | '||created_at FROM site_work_items WHERE item_type='strict_witness_234';" | psqlq)
  if [ -n "$ROW" ]; then
    { echo "=== ROW FILED (strict did NOT refuse) at $(date -u +%FT%TZ) ==="; echo "$ROW"; } >> "$OUT"
  fi
  sleep 3
done
echo "$(date -u +%FT%TZ) poller v2 done (reject_seen=${REJECT_SEEN:-no})" >> "$OUT"
echo "cleanup when finished: UPDATE agent_definitions SET is_active=false, deleted_at=now() WHERE type='strict-witness-234';" >> "$OUT"
