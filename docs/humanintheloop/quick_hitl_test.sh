#!/bin/bash

# Quick HITL Test - Single Command Version
# Copy and paste this entire command block into your terminal

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid) && \
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid) && \
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid) && \
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid) && \
ORCHESTRATION_NAME="hitl-test-$(date +%Y%m%d-%H%M%S)" && \
AGENT_ID=$(cat /proc/sys/kernel/random/uuid) && \
echo "===========================================" && \
echo "Starting HITL Test" && \
echo "Correlation ID: $CORRELATION_ID" && \
echo "===========================================" && \
kubectl -n kafka run -i --rm kcat-producer-hitl --image=edenhill/kcat:1.7.1 --restart=Never -- kcat -P -b personae-kafka-cluster-kafka-bootstrap:9092 -t system.agent.generic.requests -H correlation_id=$CORRELATION_ID -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID -H orchestration_id=$ORCHESTRATION_ID -H orchestration_name=$ORCHESTRATION_NAME -H step_name=client_hitl_request -H client_id=demo_client -H message_type=request -H action=orchestrate -H from_agent_type=user -H from_agent_id=$AGENT_ID -H responses_topic=system.generic.responses <<< '{"action":"orchestrate","config":{"group_type":"content-approval-hitl"},"input_data":{"business_name":"TechFlow Solutions","business_type":"AI automation and workflow optimization"}}' && \
echo "" && \
echo "===========================================" && \
echo "HITL workflow started!" && \
echo "Correlation ID: $CORRELATION_ID" && \
echo "" && \
echo "Now run in another terminal:" && \
echo "kubectl -n kafka run -i --rm kcat-consumer-hitl --image=edenhill/kcat:1.7.1 --restart=Never -- kcat -C -b personae-kafka-cluster-kafka-bootstrap:9092 -t system.notifications.ui -o end" && \
echo "==========================================="