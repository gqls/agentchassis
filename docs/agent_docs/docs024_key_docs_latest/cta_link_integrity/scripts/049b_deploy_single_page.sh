#!/usr/bin/env bash
# Direct page-rerender orchestrate for ONE page — bypasses the stalled
# build-dispatch-loop queue (bug 029). Assemble-only (no reason stamped -> the
# render_page branch), so authored content_data/rendered_html is untouched.
# Uses the proven 049_TRIGGER pattern: `kcat -P -c 1` + heredoc (‑c 1 reads
# exactly one message, sidestepping the kubectl-run stdin race).
set -euo pipefail
PAGE_ID="$1"; SITE_ID="$2"; DOMAIN="$3"
CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "corr=$CORR page=$PAGE_ID domain=$DOMAIN"
kubectl -n kafka run -i --rm "kcat-legal-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR \
  -H orchestration_id=$ORCH \
  -H request_id=$REQ \
  -H message_id=$MSG \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=orchestrate \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TS <<JSON
{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"page_id":"$PAGE_ID","site_id":"$SITE_ID","domain":"$DOMAIN"}}
JSON
echo "CORR=$CORR"
