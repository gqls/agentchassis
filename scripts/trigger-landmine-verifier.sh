#!/bin/bash
# trigger-landmine-verifier.sh — fire the landmine-verifier agent (RFC_005 §3.2)
# for ONE LANDMINES.md entry, via a kafka message to system.agent.generic.requests
# (the same generic spawn+call pattern as scripts/initial_messages/130_section_editor/
# 073_section_editor.sh).
#
# Usage:
#   ./scripts/trigger-landmine-verifier.sh '<doc_notes.source value>' [ref]
#   e.g.
#   ./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#some-slug' 087_towards_multiple_domains
#
# ref defaults to the current checked-out branch. It travels into the verify
# step's prompt for the verdict body only ("Checked against <ref>") — it does
# NOT pin what commit diagnose_code_lookup reads (that action reads the code
# index / live checkout, not a specific ref), so a stale ref value here would
# mislabel a verdict, not invalidate it.
#
# Verdict lands in doc_notes (categories: landmine-verification, subject_type:
# landmine, subject_key = the source value) — never edits LANDMINES.md itself.
set -euo pipefail

SOURCE="${1:?usage: $0 '<LANDMINES.md#slug>' [ref]}"
REF="${2:-$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

INPUT_DATA=$(python3 -c "
import json, sys
print(json.dumps({'source': sys.argv[1], 'ref': sys.argv[2]}))
" "$SOURCE" "$REF")

WORKFLOW=$(cat <<'ENDWF'
{
  "start_step": "spawn_verifier",
  "processing_mode": "orchestrator",
  "timeout_seconds": 300,
  "steps": {
    "spawn_verifier": {
      "action": "spawn_agent",
      "config": { "role": "verifier", "agent_type": "landmine-verifier" },
      "output_field": "verifier_agent",
      "next_step": "call_verifier",
      "description": "Spawn landmine-verifier agent"
    },
    "call_verifier": {
      "action": "call_agent",
      "config": {
        "agent_type": "landmine-verifier",
        "target_role": "verifier",
        "input_mapping": {
          "source": "input_data.source",
          "ref?": "input_data.ref"
        },
        "timeout_seconds": 240
      },
      "output_field": "verify_result",
      "next_step": "complete",
      "description": "Run one-shot verification"
    },
    "complete": {
      "action": "complete_workflow",
      "config": { "output_fields": ["verify_result"] },
      "description": "Verification complete"
    }
  }
}
ENDWF
)

WORKFLOW_COMPACT=$(echo "$WORKFLOW" | tr -d '\n' | sed 's/  */ /g')

MESSAGE_BODY=$(cat <<ENDMSG
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":${WORKFLOW_COMPACT}},"input_data":${INPUT_DATA}}
ENDMSG
)

echo "CORRELATION_ID=$CORRELATION_ID  SOURCE=$SOURCE"

kubectl -n kafka run -i --rm kcat-landmine-verify-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H client_id=$CLIENT_ID \
  -H action=process \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<ENDKAFKA
${MESSAGE_BODY}
ENDKAFKA
