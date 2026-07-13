#!/usr/bin/env bash
# =====================================================================
# idea.uk composition RE-RESOLVE — STEP 3 of 4: RE-TRIGGER (kcat)
#
# Re-enters the pipeline at COMPOSITION by orchestrating site-design-planner
# directly with the existing site_id — NOT domain-submitter (we do not want to
# redo research/strategy/briefing). Envelope is identical to 082/080c (same
# topic, headers, responses_topic, kcat pod naming).
#
# When site-design-planner finishes install_site_composition it emits its
# needs_design handoff, which the build-dispatch-loop should route to
# webdesign-agent to re-render styles.css — i.e. the cascade continues on its
# own, exactly as in a fresh build. If for any reason design does not advance,
# re-trigger it the same way with AGENT="webdesign-agent" (same input_data).
#
# Prereqs: step 2 committed (composition detached/cleared); the layouts
# migration applied; the matcher code deployed (site-design-planner rolled).
# =====================================================================
set -euo pipefail

SITE_ID="1244516d-014d-421c-88c6-090bb1e9552a"   # idea.uk
AGENT="site-design-planner"
INPUT_DATA="{\"site_id\":\"${SITE_ID}\"}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"

echo "Re-resolving composition for site ${SITE_ID} via ${AGENT}"
echo "  correlation_id=${CORRELATION_ID}"

kubectl -n kafka run -i --rm kcat-reresolve-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=reresolve-idea-uk-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"$AGENT"},"input_data":$INPUT_DATA}
JSON

echo ""
echo "=== Monitor (psql shortcut / kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db) ==="
echo "  SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid;"
echo "  -- then run reresolve_idea_uk_04_verify.sql"
