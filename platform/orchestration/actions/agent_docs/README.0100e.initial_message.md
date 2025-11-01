#!/bin/bash
# 
# Robot Hands Complete Website Creation Script
# This sends an initial message to create a complete website with home, about, and contact pages
#

# Generate IDs
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="robot-hands-complete-$(date +%Y%m%d-%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
STEP_NAME="client_step_website_request"
CLIENT_ID="demo_client"

echo "========================================="
echo "Robot Hands Complete Website Build"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Request ID:       $REQUEST_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Name:             $ORCHESTRATION_NAME"
echo "  Client:           $CLIENT_ID"
echo "  Message ID:       $MESSAGE_ID"
echo "  Agent ID:         $AGENT_ID"
echo "  Time:             $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="
echo ""
echo "This will create:"
echo "  - Homepage with hero section and image"
echo "  - About page explaining the agent system"
echo "  - Contact page with contact information"
echo ""
echo "Sending message..."
echo ""

# Send message to generic agent requests topic
kubectl -n kafka run -i --rm kcat-producer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H orchestration_name=$ORCHESTRATION_NAME \
-H step_name=$STEP_NAME \
-H client_id=$CLIENT_ID \
-H message_type=request \
-H action=orchestrate \
-H from_agent_type=user \
-H from_agent_id=$AGENT_ID \
-H responses_topic=system.responses.generic <<EOF
{"action":"orchestrate","config":{"group_type":"robot-hands-complete-website"},"input_data":{"business_name":"Robot Hands","business_type":"precision robotics and automation","domain":"robot-hands.com"}}
EOF

echo ""
echo "========================================="
echo "Message sent successfully!"
echo "========================================="
echo ""
echo "Monitor the orchestration:"
echo "  Generic Agent: kubectl logs -f deployment/agent-chassis -n agent-system | grep '$ORCHESTRATION_ID'"
echo ""
echo "Check database status:"
echo "  psql -h <host> -U <user> -d templates_db -c \"SELECT orchestration_id, status, current_step, updated_at FROM orchestration_states WHERE orchestration_id = '$ORCHESTRATION_ID' ORDER BY updated_at DESC;\""
echo ""
echo "View spawned agents:"
echo "  kubectl get pods -n agent-system | grep job"
echo ""
echo "Tail spawned agent logs:"
echo "  kubectl logs -f -n agent-system <pod-name>"
echo ""