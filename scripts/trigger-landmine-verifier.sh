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

# ---------------------------------------------------------------------------
# PUBLISH — via the shared, receipt-asserting publisher (bugs_open/327).
#
# WHAT THIS REPLACED, AND WHY IT MATTERED PARTICULARLY HERE. This was
# `kubectl -n kafka run -i --rm … kcat -P … <<ENDKAFKA`: the payload on stdin, which
# `kubectl run -i` attaches ASYNCHRONOUSLY. Lose that race and kcat sees EOF, publishes
# NOTHING and exits 0, and `--rm` deletes the evidence.
#
# The reason this file was the first migration after the filed case: its caller,
# scripts/landmines-verify-dispatch.sh:45-62, increments FAILED **only when this script
# returns non-zero**, and then prints "Dispatched N, 0 failed to publish." On the silent
# arm the old form returned 0 — so the landmine system reported success about its own
# dispatch using the one signal that is always absent when a publish is lost, and a
# verdict that never arrived was indistinguishable from the async wait that script's own
# closing message tells you to expect.
#
# That caller now gets a truthful count for free: kafka_publish_checked returns 10 (not
# published) or 11 (indeterminate), both non-zero, so FAILED counts what it claims to.
# ---------------------------------------------------------------------------
REPO_ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$REPO_ROOT" ] || [ ! -f "$REPO_ROOT/scripts/kafka-publish-lib.sh" ]; then
  echo "ERROR: scripts/kafka-publish-lib.sh not found (repo root: ${REPO_ROOT:-<not in a git repo>})." >&2
  echo "       Refusing to publish unverified — an unverified dispatch is bugs_open/327." >&2
  exit 1
fi
. "$REPO_ROOT/scripts/kafka-publish-lib.sh"

# The body is already single-line (WORKFLOW_COMPACT strips newlines); the library refuses
# it outright if that ever stops being true, rather than letting `kcat -P` publish it as
# one message per line.
PUBLISH_RC=0
kafka_publish_checked \
  --topic system.agent.generic.requests \
  --correlation "$CORRELATION_ID" \
  --payload "$MESSAGE_BODY" \
  --header "orchestration_id=$ORCHESTRATION_ID" \
  --header "request_id=$REQUEST_ID" \
  --header "message_id=$MESSAGE_ID" \
  --header "message_type=request" \
  --header "client_id=$CLIENT_ID" \
  --header "action=process" \
  --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-user" \
  --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$TIMESTAMP" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "VERIFICATION NOT DISPATCHED for $SOURCE — no verdict will ever arrive for this run." >&2
  exit "$PUBLISH_RC"
fi
