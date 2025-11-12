#!/bin/bash

# HITL Content Approval Workflow for EBORG
# This script triggers a content generation workflow with human approval for EBORG

# Generate IDs
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="eborg-content-approval-$(date +%Y%m%d-%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
STEP_NAME="client_eborg_content_request"
CLIENT_ID="demo_client"

# EBORG business details
BUSINESS_NAME="EBORG"
BUSINESS_TYPE="Evidence-Based Organisational Planning"
BUSINESS_DESCRIPTION="EBORG helps organisations evolve intelligently by mapping every role, responsibility, and objective, then pairing each with a framework of AI agents designed to augment human capability. These agents don't just automate routine work — they gather and analyse research, assess strategic options, and provide evidence-based reasoning to guide decision-making at every level. The result is a human-centered, continuously learning organisation where people lead with vision, and AI supports with insight. Eborg aims to empower every employee, from CEO downwards, with intelligent, evidence-based support — combining human insight with AI reasoning to create an organisation that thinks, learns, and performs better together. Not a futuristic AI revolution, but a necessary, structured evolution — a calm, practical, intelligent step forward in how modern organisations operate."

echo "========================================="
echo "HITL Content Approval for EBORG"
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
echo "Business: $BUSINESS_NAME"
echo "Type: $BUSINESS_TYPE"
echo ""
echo "This workflow will:"
echo "  1. Generate organisational description content"
echo "  2. Send approval request to system.notifications.ui"
echo "  3. Wait for human approval"
echo "  4. Return approved content with metadata"
echo ""
echo "Sending message..."
echo ""

# Send message to generic agent requests topic
kubectl -n kafka run -i --rm kcat-producer-eborg \
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
-H responses_topic=system.generic.responses <<EOF
{"action":"orchestrate","config":{"group_type":"content-approval-hitl"},"input_data":{"business_name":"$BUSINESS_NAME","business_type":"$BUSINESS_TYPE","business_description":"$BUSINESS_DESCRIPTION"}}
EOF

echo ""
echo "========================================="
echo "Message sent successfully!"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Monitor the approval queue in Terminal 1:"
echo "   kubectl -n kafka run -i --rm kcat-consumer-hitl \\"
echo "     --image=edenhill/kcat:1.7.1 --restart=Never -- \\"
echo "     kcat -C -b personae-kafka-cluster-kafka-bootstrap:9092 -t system.notifications.ui -o end"
echo ""
echo "2. When you see the approval request, note:"
echo "   - correlation_id: $CORRELATION_ID"
echo "   - request_id (the approval token from the notification)"
echo ""
echo "3. Send approval response in Terminal 2:"
echo "   ./send_approval_eborg.sh $CORRELATION_ID <APPROVAL_TOKEN>"
echo ""
echo "4. Monitor workflow status in Terminal 3:"
echo "   ./monitor_workflow.sh $CORRELATION_ID"
echo ""
echo "========================================="