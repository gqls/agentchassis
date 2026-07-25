#!/usr/bin/env bash
# Fire the improvement loop for ONE site — the same orchestrate envelope the
# (disabled) improvement-sweep scheduled_task would send, built by hand for a
# single site so a discovery cycle can be exercised on demand.
#
# Written for bugs_open/006 §B behavioural verification (2026-07-25): the
# contact_form_undeliverable check's auto-remediate branch (v1.0.1156) only
# fires inside a discovery cycle, and nothing runs discovery on a cadence
# (improvement-sweep is disabled), so proving the checker→handler chain needs
# a manual single-site run.
#
# Pattern: proven 049b/097 kcat publish — `kcat -P -c 1` + heredoc (-c 1 reads
# exactly one message, sidestepping the kubectl-run stdin race).
#
# Usage: ./091_TRIGGER_improvement_loop_single_site.sh <site_id> <domain>
# Watch: SELECT orchestration_id, current_step, status FROM orchestration_states
#        WHERE correlation_id='<printed corr>';
set -euo pipefail
SITE_ID="$1"; DOMAIN="$2"
CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "corr=$CORR site=$SITE_ID domain=$DOMAIN"
kubectl -n kafka run -i --rm "kcat-improve-$(date +%s)" \
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
{"action":"orchestrate","config":{"agent_type":"improvement-loop"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN"}}
JSON
echo "IMPROVEMENT_CORR=$CORR"
