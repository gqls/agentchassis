#!/usr/bin/env bash
# Fire design-discovery-agent at one site — payload passed as the container COMMAND,
# not on stdin.
#
# WHY NOT THE HEREDOC FORM: `kubectl run -i --rm ... -- kcat -P -c 1 <<JSON` races.
# If the container starts before stdin is attached, kcat sees EOF, produces NOTHING,
# and exits 0 — the pod is then deleted and the wrapper prints a correlation id as if
# it had worked. Measured 2026-07-26: 1 of 5 publishes landed, silently.
set -u
SITE_ID="${1:?site_id}"; DOMAIN="${2:?domain}"
AGENT_TYPE="design-discovery-agent"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BODY="{\"action\":\"orchestrate\",\"config\":{\"agent_type\":\"${AGENT_TYPE}\"},\"input_data\":{\"site_id\":\"${SITE_ID}\",\"domain\":\"${DOMAIN}\"}}"

echo "=== ${AGENT_TYPE} -> ${DOMAIN} corr=${CORRELATION_ID} ==="
kubectl -n kafka run "kcat-disc-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '${BODY}' | kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
  -H correlation_id=${CORRELATION_ID} \
  -H orchestration_id=$(cat /proc/sys/kernel/random/uuid) \
  -H request_id=$(cat /proc/sys/kernel/random/uuid) \
  -H message_id=$(cat /proc/sys/kernel/random/uuid) \
  -H message_type=request -H client_id=demo_client -H action=orchestrate \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=${TIMESTAMP} && echo PUBLISH_OK"
echo "corr=${CORRELATION_ID}"
