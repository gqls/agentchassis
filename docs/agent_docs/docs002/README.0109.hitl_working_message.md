CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="eborg-content-approval-$(date +%Y%m%d-%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
STEP_NAME="client_eborg_content_request"
CLIENT_ID="demo_client"

echo "========================================="
echo "HITL Content Approval for Simple workflow without LLM"
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
-H responses_topic=system.generic.responses <<JSON
{"action":"orchestrate","config":{"workflow":{"start_step":"transform","steps":{"transform":{"action":"transform_data","config":{"transformation":"uppercase"},"next_step":"await_approval"},"await_approval":{"action":"await_approval","config":{"approval_fields":["transform"]},"next_step":"process_approval"},"process_approval":{"action":"process_approval_decision","next_step":"done"},"done":{"action":"complete_workflow"}}}},"input_data":{"message":"test that this is transformed to uppercase."}}
JSON


# This command simulates the HITL service sending an approval
# It targets the CHILD AGENT'S RESPONSES TOPIC
CHILD_RESPONSES_TOPIC=system.agent.generic.responses
CORRELATION_ID=6301033f-bf11-4d26-9b7e-87de77478677
CHILD_ORCHESTRATION_ID=47b25921-2a9b-4ffd-9724-be2f779d75ab
APPROVAL_TOKEN=a0a75bd8-262c-4f55-91eb-19398f73458e

kubectl -n kafka run -i --rm kcat-producer-eborg-approval \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t $CHILD_RESPONSES_TOPIC \
-H correlation_id=$CORRELATION_ID \
-H orchestration_id=$CHILD_ORCHESTRATION_ID \
-H in_response_to_request_id=$APPROVAL_TOKEN \
-H client_id=demo_client \
-H message_type=response \
-H action=approval_response \
-H from_agent_type=human \
-H status=complete \
-H from_agent_id="hitl_system" \
<<JSON
{"success":true,"status": "completed","body":{"approved":true,"comments":"Content doesnt need revision - in simple test","approved_by":"test-user@example.com"}}
JSON




