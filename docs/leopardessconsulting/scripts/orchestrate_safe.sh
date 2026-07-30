#!/bin/bash
# ============================================================================
# orchestrate_safe.sh — fire ANY agent_type with a publish receipt.
#
# Generic sibling of rerender_page_safe.sh / tool_acceptance_run.sh. Same reason
# for existing: the `echo "$PAYLOAD" | kubectl run -i --rm … kcat -P` form used
# throughout this repo's trigger scripts publishes NOTHING roughly four times in
# five (kubectl run -i attaches stdin asynchronously; the container reaches kcat,
# sees EOF, exits 0, and --rm deletes the evidence). The payload therefore
# travels in the container COMMAND, base64'd so no quoting can break it, and the
# publisher prints PUBLISH_OK.
#
# `--command` is load-bearing: the image's ENTRYPOINT is kcat, so without it the
# `sh -c …` arrives as arguments TO kcat and nothing is published.
#
# BEFORE RUNNING: no orchestration dispatch within ~300s of a chassis pod
# (re)start, or the spawn is silently dropped:
#   kubectl -n ai-persona-system get pods -l app=agent-chassis \
#     -o custom-columns='NAME:.metadata.name,START:.status.startTime'
#
# Usage: ./orchestrate_safe.sh <agent_type> <input_data_json>
#   e.g. ./orchestrate_safe.sh nav-link-fixer \
#          '{"site_id":"4851f6fc-...","domain":"leopardessconsulting.co.uk"}'
# ============================================================================
set -euo pipefail

AGENT="${1:?agent_type}"
INPUT_JSON="${2:?input_data as JSON object}"

CID=$(cat /proc/sys/kernel/random/uuid)
OID=$(cat /proc/sys/kernel/random/uuid)
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

PAYLOAD_B64=$(python3 - "$AGENT" "$INPUT_JSON" <<'PY'
import base64, json, sys
agent, input_json = sys.argv[1:3]
data = json.loads(input_json)          # fail loudly here, not in the cluster
msg = {"action": "orchestrate",
       "config": {"agent_type": agent},
       "input_data": data}
line = json.dumps(msg, separators=(",", ":"))
assert "\n" not in line
sys.stdout.write(base64.b64encode(line.encode()).decode())
PY
)

echo "agent:       $AGENT"
echo "correlation: $CID"
echo "publishing to system.agent.generic.requests ..."

kubectl -n kafka run "kcat-orch-$(date +%s)-$RANDOM" \
  --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "echo '$PAYLOAD_B64' | base64 -d | kcat -P \
    -b $BROKER \
    -t system.agent.generic.requests \
    -H correlation_id=$CID \
    -H orchestration_id=$OID \
    -H request_id=$(cat /proc/sys/kernel/random/uuid) \
    -H message_id=$(cat /proc/sys/kernel/random/uuid) \
    -H orchestration_name=$AGENT \
    -H step_name=start \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=user \
    -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"

echo ""
echo "No PUBLISH_OK above means NOTHING was published — re-run now."
echo "  SELECT current_step, status, error FROM orchestration_states"
echo "   WHERE correlation_id = '$CID';"
