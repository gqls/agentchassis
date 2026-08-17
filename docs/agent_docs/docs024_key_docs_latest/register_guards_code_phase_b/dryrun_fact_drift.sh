#!/bin/bash
# One-off DRY RUN of refresh_evidence_base, to see what the fact-drift fan-out
# (CLM-022) would file without writing anything.
#
# Why an inline workflow rather than the evidence-freshness agent: `dry_run` is
# read from the STEP's config, not from input_data, so the live agent row cannot
# be asked for a dry run. The chassis honours an inline workflow at
# body.config.workflow (Priority 1).
#
# GOTCHAS, each of which has cost someone a cycle:
#   - kcat -P splits on newlines: the JSON MUST be one line, or it silently
#     publishes fragments and exits 0.
#   - No dispatch within ~300s of a chassis pod restart — the spawn is dropped.
#   - Verify by the orchestration row, never by kcat's exit code.
#
# Usage: ./dryrun_fact_drift.sh [site_id]     (omit site_id to sweep every
#        site holding a current evidence_base spec)
set -euo pipefail
SITE_ID="${1:-}"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

if [ -n "$SITE_ID" ]; then INPUT="{\"site_id\":\"${SITE_ID}\"}"; else INPUT="{}"; fi

WF='{"start_step":"refresh_evidence","processing_mode":"orchestrator","timeout_seconds":600,"steps":{"refresh_evidence":{"action":"refresh_evidence_base","config":{"dry_run":true},"output_field":"refresh_result","next_step":"complete","description":"Dry-run the evidence sweep incl. the fact-drift fan-out"},"complete":{"action":"complete_workflow","config":{"output_fields":["refresh_result"]},"description":"done"}}}'

kubectl -n kafka run -i --rm kcat-factdrift-$(date +%s) \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
  -H message_type=request -H client_id=demo_client -H action=process \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":${WF}},"input_data":${INPUT}}
JSON

echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo "Read it back (a missing row is LATENCY, not a dropped dispatch — do not retry on that evidence):"
echo "  SELECT status, current_step, jsonb_pretty(collected_data->'refresh_result') FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}';"
