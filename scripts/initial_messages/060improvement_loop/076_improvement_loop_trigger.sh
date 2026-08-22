#!/bin/bash
# ============================================================================
# IMPROVEMENT LOOP — Manual Trigger
# ============================================================================
# Runs post-build quality checks and dispatches fixes for a site.
#
# What it does:
#   1. Spawns quality-discovery-agent     (broken nav, placeholder contact, generic theme)
#   2. Spawns design-discovery-agent      (undeployed assets, missing CSS, duplicate palette)
#   3. Spawns completeness-discovery-agent (empty sections)
#   4. Triages findings (detected → triaged, domain → build)
#   5. If issues found: inserts needs_rerender, dispatches fixes via build-dispatch-loop
#
# Usage:
#   ./076_improvement_loop_trigger.sh <site_id> <domain>           # specific site
#   SITE_ID=xxx DOMAIN=yyy ./076_improvement_loop_trigger.sh       # env vars
#   ./076_improvement_loop_trigger.sh                              # refuses: no default target
#
# THREE TRAPS FIXED 2026-08-22 (dartsonline_traffic lane). All three were silent,
# and the third is why the first two survived: the script never returned, so
# nobody read what it had actually done.
#
#   1. IT IGNORED ITS OWN ARGUMENTS. It parsed $1/$2, then re-assigned
#      SITE_ID/DOMAIN to robot-hands.com unconditionally two lines later. So
#      `./076... <my-site-id> mysite.com` fired the improvement loop at
#      robot-hands.com -- ANOTHER LANE'S SITE -- and printed "Improvement loop
#      triggered" as though it had worked. There is now NO default target: with
#      no arguments it refuses, because a silent wrong-site dispatch is worse
#      than an error. It also cross-checks the site_id/domain pair against the
#      database and refuses on a mismatch.
#
#   2. IT PUBLISHED ON STDIN. `kubectl run -i --rm ... kcat -P -c 1 <<JSON`
#      loses ~4 of 5 messages at exit 0: stdin attaches asynchronously, a
#      container reaching kcat first sees EOF, produces nothing, exits clean,
#      and --rm deletes the evidence. Now the container-COMMAND form with a
#      PUBLISH_OK receipt -- no PUBLISH_OK in the output means nothing was sent.
#
#   3. IT NEVER EXITED. Six `kubectl logs -f` calls at the end followed logs for
#      ever and wrote logs-*.json into $PWD. They are printed as guidance now.


# ============================================================================
set -euo pipefail

AGENT_TYPE="improvement-loop"
SITE_ID="${1:-${SITE_ID:-}}"
DOMAIN="${2:-${DOMAIN:-}}"

# No default target, deliberately (trap 1 in the header): this script carried one
# and silently overrode whatever the caller asked for.
if [ -z "$SITE_ID" ] || [ -z "$DOMAIN" ]; then
  echo "REFUSING: pass the target explicitly." >&2
  echo "  usage: $0 <site_id> <domain>   (or SITE_ID=... DOMAIN=... $0)" >&2
  exit 2
fi

# Assert the pair agrees with the database before dispatching. A site_id sitting
# next to someone else's domain is exactly the shape trap 1 produced.
ACTUAL_DOMAIN=$(kubectl exec -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -tAc "SELECT domain FROM sites WHERE id='${SITE_ID}'" 2>/dev/null | tr -d '\r')
if [ "$ACTUAL_DOMAIN" != "$DOMAIN" ]; then
  echo "REFUSING: site_id ${SITE_ID} is '${ACTUAL_DOMAIN:-<not found>}', not '${DOMAIN}'." >&2
  exit 3
fi

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Improvement Loop"
echo "========================================="
echo "  Domain:         $DOMAIN"
echo "  Site ID:        $SITE_ID"
echo "  Correlation ID: $CORRELATION_ID"
echo "========================================="

# The body MUST be built with an UNQUOTED heredoc so ${SITE_ID} and the ids expand.
# A single-quoted printf (my first attempt) ships the literal text ${SITE_ID} to the
# loop, which then targets nothing — and the publish still reports PUBLISH_OK.
PAYLOAD=$(cat <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_improvement_loop","steps":{"spawn_improvement_loop":{"action":"spawn_agent","config":{"role":"improver","agent_type":"improvement-loop"},"description":"Spawn improvement loop agent","next_step":"call_improvement_loop","output_field":"improver_agent"},"call_improvement_loop":{"action":"call_agent","config":{"agent_type":"improvement-loop","target_role":"improver","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain"},"timeout_seconds":1800},"description":"Run improvement loop — discover issues, triage, dispatch fixes, rerender","next_step":"complete","output_field":"improvement_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["improvement_result"]},"description":"Improvement loop complete"}}}},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON
)
PAYLOAD_B64=$(printf '%s' "$PAYLOAD" | base64 -w0)

kubectl -n kafka run "kcat-improve-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet --command -- sh -c \
  "echo '$PAYLOAD_B64' | base64 -d | kcat -P \
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
   -H timestamp=$TIMESTAMP && echo PUBLISH_OK"

echo ""
echo "========================================="
echo "Improvement loop triggered"
echo "========================================="
echo ""
echo "Monitor discovery phase:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=quality-discovery-agent --tail=20"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=design-discovery-agent --tail=20"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=completeness-discovery-agent --tail=20"
echo ""
echo "Monitor fix dispatch:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=build-dispatch-loop --tail=30"
echo ""
echo "Check findings:"
echo "  psql -c \"SELECT item_type, status, severity, summary FROM site_work_items WHERE site_id='${SITE_ID}' AND source='discovery' ORDER BY created_at DESC LIMIT 20;\""
echo ""
echo "Overall progress:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep '${CORRELATION_ID}'"
echo ""
echo "CORRELATION_ID=$CORRELATION_ID"





echo ""
echo "Follow the agents (each of these BLOCKS -- run them in your own shell, not here):"
for a in improvement-loop quality-discovery-agent design-discovery-agent \
         completeness-discovery-agent build-dispatch-loop; do
  echo "  kubectl -n ai-persona-system logs --tail=300 -l agent-type=$a -f"
done
