#!/bin/bash
# ============================================================================
# reconcile_headers.sh — make every active page carry the current (forked) header.
#
# Fire-and-forget kcat at volume drops messages, so this RECONCILES: each round it
# fires page-rerender (assemble mode) only for pages whose DEPLOYED html still shows
# the old blue header (#0f3460), waits, and repeats until none remain or max rounds.
#
# render_single_page pulls header/footer from site_components (the fork), so plain
# re-assembly is enough — no section regen.
#
# Usage: ./reconcile_headers.sh <site_id> <domain>
# ============================================================================
set -uo pipefail
S="${1:?}"; DOMAIN="${2:?}"
PSQL=(kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -tAc)

mapfile -t URLS < <("${PSQL[@]}" "SELECT url FROM pages WHERE site_id='$S' AND status='active' ORDER BY nav_order")
echo "active pages: ${#URLS[@]}"

fire() { # page_name
  local PN="$1" CID OID
  CID=$(cat /proc/sys/kernel/random/uuid); OID=$(cat /proc/sys/kernel/random/uuid)
  printf '{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"%s","domain":"%s","page_name":"%s"}}' "$S" "$DOMAIN" "$PN" \
  | kubectl -n kafka run -i --rm "kcat-rc-${PN//[^a-z0-9]/}-$(date +%s%N | tail -c 5)" --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
    -H correlation_id=$CID -H orchestration_id=$OID -H request_id=$(cat /proc/sys/kernel/random/uuid) \
    -H message_id=$(cat /proc/sys/kernel/random/uuid) -H orchestration_name=rc-$PN -H step_name=start \
    -H client_id=demo_client -H message_type=request -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses >/dev/null 2>&1
}

for round in 1 2 3 4 5; do
  todo=()
  for u in "${URLS[@]}"; do
    body=$(curl -s "https://${DOMAIN}${u}?cb=$(date +%s%N)")
    # old header present if the blue button bg is there OR the logo is missing
    if echo "$body" | grep -q 'background: #0f3460' || ! echo "$body" | grep -q 'logo.png'; then
      pn=$(basename "$u" .html)
      todo+=("$pn")
    fi
  done
  echo "round $round: ${#todo[@]} pages still stale"
  [ ${#todo[@]} -eq 0 ] && { echo "ALL UNIFORM"; break; }
  for pn in "${todo[@]}"; do fire "$pn"; done
  echo "  fired ${#todo[@]}; waiting for deploys..."
  sleep 90
done
