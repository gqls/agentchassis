#!/bin/bash
# Re-assemble + deploy every deployable page of a site after a CHROME change.
# Generalised from idea_uk_vm_site/scripts/fire_reassemble_idea_uk.sh.
#
# Usage:  ./fire_reassemble_site.sh <domain> [max_parallel]
#
# ASSEMBLE MODE (no input_data.spec.reason) is required for a chrome-only change: it
# takes the ELSE branch of check_rerender_mode = stored section HTML + CURRENT chrome +
# deploy. Do NOT use section_data_resolved — it bails at page level for pages with no
# stored components and neither bail-out deploys, so chrome changes silently miss them
# (bugs_closed/031). page_id is REQUIRED; assemble mode fails "page_id not found" without it.
#
# ⚠ PUBLISH PATTERN IS LOAD-BEARING. `kubectl run -i … kcat -P` loses ~4 of 5 messages at
# exit 0 (stdin attaches asynchronously; kcat hits EOF first; --rm deletes the evidence).
# Payload goes in the container COMMAND, `--command` is required (image ENTRYPOINT is
# kcat), and every publish must print PUBLISH_OK.
set -uo pipefail

DOMAIN="${1:?usage: fire_reassemble_site.sh <domain> [max_parallel]}"
PAR="${2:-4}"
BROKER=personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092
TOPIC=system.agent.generic.requests
PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc)

SITE=$("${PSQL[@]}" "SELECT id FROM sites WHERE domain='$DOMAIN' AND status='deployed'")
[ -z "$SITE" ] && { echo "no deployed site for domain '$DOMAIN'"; exit 1; }

# Only pages that have sections. A page with none assembles to nothing and is skipped by
# the action anyway; including it just adds noise and failed-looking rows.
mapfile -t ROWS < <("${PSQL[@]}" "
  SELECT p.id || '|' || p.name FROM pages p
   WHERE p.site_id='$SITE' AND p.status='active'
     AND EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id=p.id)
   ORDER BY p.name")

OUT="/tmp/reassemble_${DOMAIN//[^a-z0-9]/_}_corrs.txt"
: > "$OUT"
echo "== $DOMAIN ($SITE): ${#ROWS[@]} page(s), parallel=$PAR"

fire() {
  local pid="${1%%|*}" pn="${1#*|}"
  local CID OID
  CID=$(cat /proc/sys/kernel/random/uuid); OID=$(cat /proc/sys/kernel/random/uuid)
  local PAYLOAD
  PAYLOAD=$(printf '{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"%s","domain":"%s","page_id":"%s","page_name":"%s"}}' \
    "$SITE" "$DOMAIN" "$pid" "$pn")
  local RES
  RES=$(kubectl -n kafka run "kcat-gtm-$(date +%s%N | tail -c 8)-$RANDOM" \
        --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
        --command -- sh -c "printf '%s' '$PAYLOAD' | kcat -P -b $BROKER -t $TOPIC \
          -H correlation_id=$CID -H orchestration_id=$OID \
          -H request_id=$(cat /proc/sys/kernel/random/uuid) \
          -H message_id=$(cat /proc/sys/kernel/random/uuid) \
          -H orchestration_name=gtm-$(date +%H%M%S)-$RANDOM -H step_name=start \
          -H client_id=demo_client -H message_type=request -H action=orchestrate \
          -H from_agent_type=user -H from_agent_id=cli \
          -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK" 2>&1)
  if grep -q PUBLISH_OK <<<"$RES"; then
    printf '%s|%s\n' "$CID" "$pn" >> "$OUT"; echo "  ok   $pn"
  else
    echo "  FAIL $pn :: $(tr '\n' ' ' <<<"$RES" | cut -c1-160)"
  fi
}

n=0
for row in "${ROWS[@]}"; do
  [ -z "$row" ] && continue
  fire "$row" &
  n=$((n+1)); (( n % PAR == 0 )) && wait
done
wait

echo "published $(wc -l < "$OUT")/${#ROWS[@]}  -> $OUT"
