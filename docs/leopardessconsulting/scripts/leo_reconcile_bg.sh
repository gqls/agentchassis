#!/bin/bash
S=4851f6fc-71cf-4160-a270-e03d6d3e0732
DOMAIN=leopardessconsulting.co.uk
LOG=/home/ant/leo_shots/reconcile.log
: > "$LOG"
mapfile -t ROWS < <(kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "SELECT id||'|'||name||'|'||url FROM pages WHERE site_id='$S' AND status='active' ORDER BY nav_order")
for round in 1 2 3 4 5 6 7 8; do
  todo=()
  for row in "${ROWS[@]}"; do
    PID="${row%%|*}"; rest="${row#*|}"; PN="${rest%%|*}"; URL="${rest#*|}"
    curl -s "https://${DOMAIN}${URL}?cb=$(date +%s%N)" | grep -q 'background: #C8A951' || todo+=("$PID|$PN")
  done
  echo "$(date +%H:%M:%S) round $round: $(( ${#ROWS[@]} - ${#todo[@]} ))/${#ROWS[@]} gold, ${#todo[@]} todo" >> "$LOG"
  [ ${#todo[@]} -eq 0 ] && { echo "ALL UNIFORM" >> "$LOG"; break; }
  for item in "${todo[@]}"; do
    PID="${item%%|*}"; PN="${item#*|}"
    CID=$(cat /proc/sys/kernel/random/uuid); OID=$(cat /proc/sys/kernel/random/uuid)
    printf '{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"%s","domain":"%s","page_id":"%s","page_name":"%s","spec":{"reason":"section_data_resolved","page_name":"%s"}}}' "$S" "$DOMAIN" "$PID" "$PN" "$PN" \
    | kubectl -n kafka run -i --rm "kr-$(echo $PN|tr -cd a-z0-9|cut -c1-8)-$(date +%s%N|tail -c 5)" --image=edenhill/kcat:1.7.1 --restart=Never -- \
      kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
      -H correlation_id=$CID -H orchestration_id=$OID -H request_id=$(cat /proc/sys/kernel/random/uuid) \
      -H message_id=$(cat /proc/sys/kernel/random/uuid) -H orchestration_name=rec-$PN -H step_name=start \
      -H client_id=demo_client -H message_type=request -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
      -H responses_topic=system.agent.generic.responses >/dev/null 2>&1
  done
  echo "$(date +%H:%M:%S)   fired ${#todo[@]}, draining ~$(( ${#todo[@]} * 50 ))s" >> "$LOG"
  sleep $(( 60 + ${#todo[@]} * 45 ))
done
echo "$(date +%H:%M:%S) DONE" >> "$LOG"
