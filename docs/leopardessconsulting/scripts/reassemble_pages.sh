#!/bin/bash
# Re-assemble pages (embed current header/footer slots + deploy) WITHOUT regenerating
# sections. page-rerender with no spec.reason takes the check_rerender_mode ELSE branch
# = assemble stored section HTML + fresh header/footer + deploy. Use after a header/
# footer/site_components change. One page per call → a stall can't block the rest.
set -euo pipefail
S="${1:?}"; DOMAIN="${2:?}"; shift 2
for PN in "$@"; do
  CID=$(cat /proc/sys/kernel/random/uuid); OID=$(cat /proc/sys/kernel/random/uuid)
  printf '{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"%s","domain":"%s","page_name":"%s"}}' "$S" "$DOMAIN" "$PN" \
  | kubectl -n kafka run -i --rm "kcat-as-${PN//[^a-z0-9]/}-$(date +%s%N | tail -c 6)" --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
    -H correlation_id=$CID -H orchestration_id=$OID -H request_id=$(cat /proc/sys/kernel/random/uuid) \
    -H message_id=$(cat /proc/sys/kernel/random/uuid) -H orchestration_name=as-${PN}-$(date +%H%M%S) -H step_name=start \
    -H client_id=demo_client -H message_type=request -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses >/dev/null 2>&1
  echo "$PN queued"
done
