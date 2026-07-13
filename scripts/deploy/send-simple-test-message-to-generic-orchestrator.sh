#!/bin/bash
# run this script: kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- bash < scripts/deploy/send-simple-test-message-to-generic-orchestrator.sh

# Generate UUIDs for correlation_id
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "$(date +%s)-$(shuf -i 1000-9999 -n 1)-4000-8000-$(shuf -i 100000000000-999999999999 -n 1)")
REQUEST_ID="req-$(date +%s)"
WORKFLOW_ID="wf-$(date +%s)"

echo "Sending message with:"
echo "  correlation_id: $CORRELATION_ID"
echo "  request_id: $REQUEST_ID"
echo "  workflow_id: $WORKFLOW_ID"

# The format is: header1:value1,header2:value2<TAB>messageBody
# All on ONE line - headers, then tab, then JSON body
printf "correlation_id:$CORRELATION_ID,request_id:$REQUEST_ID,client_id:demo_client,agent_instance_id:generic-001,fuel_budget:1000\t{\"action\":\"process\",\"workflow_id\":\"${WORKFLOW_ID}\",\"client_id\":\"demo_client\",\"payload\":{\"task\":\"test\",\"data\":\"Test message with UUID\"}}\n" | /opt/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic system.agent.generic.process \
  --property parse.headers=true \
  --property headers.delimiter=$'\t'