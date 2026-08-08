#!/usr/bin/env bash
# Fire a BATCH of voice-H rewrites, wait for every orchestration to reach a
# terminal state, then grade them.
#
# Batch, not whole-site, on purpose: a bad batch should cost a batch. The grader
# is run by hand afterwards on the printed page list — this script only proves
# the orchestrations finished, which is NOT the same as proving they worked.
#
# Usage: voiceh_batch.sh <page-name> [<page-name> ...]
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
ORCHS=()
PAGES=()

for p in "$@"; do
  line=$(timeout 180 "$HERE/voiceh_rewrite.sh" "$p" 2>&1 | grep -E "^ITEM=" || true)
  if [ -z "$line" ]; then echo "DISPATCH FAILED: $p" >&2; continue; fi
  o=${line##*ORCH=}
  ORCHS+=("$o"); PAGES+=("$p")
  echo "fired $p -> $o"
done

if [ ${#ORCHS[@]} -eq 0 ]; then echo "nothing fired" >&2; exit 1; fi

IN=$(printf "'%s'," "${ORCHS[@]}"); IN=${IN%,}

# kcat -P exits 0 having sent nothing, so the orchestration rows are the ONLY
# proof of dispatch. A row that never appears is a dropped publish, not a slow one.
for i in $(seq 1 60); do
  read -r n_rows n_done <<<"$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -t -A -F' ' -c \
    "SELECT count(*), count(*) FILTER (WHERE status IN ('COMPLETED','FAILED','ERROR'))
     FROM orchestration_states WHERE orchestration_id IN ($IN);" 2>/dev/null)"
  echo "$(date -u +%H:%M:%S)  rows=$n_rows/${#ORCHS[@]}  terminal=$n_done"
  [ "${n_done:-0}" = "${#ORCHS[@]}" ] && break
  sleep 20
done

kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
  "SELECT collected_data->'input_data'->>'page_name' AS page, status, current_step
   FROM orchestration_states WHERE orchestration_id IN ($IN) ORDER BY 1;"

echo
echo "now grade:  python3 $HERE/voiceh_grade.py <baseline.json> ${PAGES[*]}"
