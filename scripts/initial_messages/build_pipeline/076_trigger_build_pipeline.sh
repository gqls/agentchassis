
first:
INSERT INTO build_queue (domain, priority) VALUES ('example.com', 10);

gaswholesalers.com
INSERT INTO build_queue (domain, priority) VALUES ('gaswholesalers.com', 10);

#!/bin/bash
# =============================================================================
# Build Pipeline Trigger (manual heartbeat)
# =============================================================================
# Sends an orchestrate message to the build-pipeline-trigger agent.
# This is the manual equivalent of the CronJob heartbeat that would
# normally fire every 30 minutes.
#
# What it does:
#   1. seed_build_queue — processes build_queue entries → creates sites + work items
#   2. find_dispatchable_site — queries for sites with pending build work items
#   3. If found: spawns + calls build-dispatch-loop for that site
#   4. The dispatch loop processes items one at a time, chaining to itself
#
# Prerequisites:
#   - build-pipeline-trigger agent definition in agent_definitions table
#   - build-dispatch-loop agent definition in agent_definitions table
#   - Handler agents registered (domain-research-classifier, build-briefing-agent, etc.)
#   - Entries in build_queue table, e.g.:
#       INSERT INTO build_queue (domain, priority) VALUES ('example.com', 10);
#
# Usage:
#   ./054_trigger_build_pipeline.sh
# =============================================================================

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Build Pipeline Trigger"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Time:             $TIMESTAMP"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

kubectl -n kafka run -i --rm kcat-build-trigger-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=build-pipeline-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"build-pipeline-trigger"},"input_data":{}}
JSON

echo ""
echo "========================================="
echo "Build pipeline triggered"
echo "========================================="
echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100 | grep '$CORRELATION_ID'"
echo ""
echo "Check seed results:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 | grep '$CORRELATION_ID' | grep -E 'seed_queue|seed_build_queue|find_dispatchable'"
echo ""
echo "Check dispatch:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 | grep '$CORRELATION_ID' | grep -E 'spawn_dispatch|call_dispatch|dispatch_result'"
echo ""
echo "Check orchestration state:"
echo "  SELECT status, current_step, error FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid;"
echo ""
echo "Check build_queue:"
echo "  SELECT domain, status, priority, created_at FROM build_queue ORDER BY created_at DESC LIMIT 5;"
echo ""
echo "Check work items:"
echo "  SELECT wi.item_type, wi.status, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.domain = 'build' ORDER BY wi.created_at DESC LIMIT 10;"
echo ""