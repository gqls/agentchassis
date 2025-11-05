059_human_in_the_loop_adventures 5f91200

#!/bin/bash
# Testing HITL Approval Actions - Step by Step Guide

# ==============================================================================
# TESTING THE HITL APPROVAL WORKFLOW
# ==============================================================================

echo "Human-In-The-Loop Approval Testing Guide"
echo "========================================="

# Step 1: Start the notification listener
echo ""
echo "STEP 1: Start listening for approval requests"
echo "----------------------------------------------"
echo "In Terminal 1, run this command to watch for approval notifications:"
cat << 'EOF'
kubectl -n kafka run -i --rm kcat-consumer-notifications \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -C -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.notifications.ui -o json | jq '.'
EOF

# Step 2: Trigger a simple workflow with approval
echo ""
echo "STEP 2: Trigger a workflow that needs approval"
echo "----------------------------------------------"
echo "In Terminal 2, send a test message to trigger the workflow:"
cat << 'EOF'
CORRELATION_ID=$(uuidgen)
REQUEST_ID=$(uuidgen)

kubectl -n kafka run -i --rm kcat-producer-test \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -P -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_type=request \
-H action=orchestrate \
<<JSON
{
"action": "orchestrate",
"config": {
"workflow": {
"start_step": "generate",
"steps": {
"generate": {
"action": "execute_llm_prompt",
"config": {
"prompt_template": "Write a short greeting for {{name}}",
"input_fields": ["name"]
},
"next_step": "await_approval",
"description": "Generate content"
},
"await_approval": {
"action": "await_approval",
"config": {
"approval_fields": ["generate"],
"approval_type": "content_review",
"ui_config": {
"title": "Content Approval Required",
"description": "Please approve the generated greeting"
}
},
"next_step": "process_approval",
"description": "Wait for approval"
},
"process_approval": {
"action": "process_approval_decision",
"next_step": "complete",
"description": "Process the decision"
},
"complete": {
"action": "complete_workflow",
"description": "Complete"
}
}
}
},
"input_data": {
"name": "Test User"
}
}
JSON
EOF

# on one line
echo ""
echo "STEP 2: Trigger a workflow that needs approval"
echo "----------------------------------------------"
echo "In Terminal 2, send a test message to trigger the workflow:"
cat << 'EOF'
CORRELATION_ID=$(uuidgen)
REQUEST_ID=$(uuidgen)

kubectl -n kafka run -i --rm kcat-producer-test \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -P -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_type=request \
-H action=orchestrate \
{"action":"orchestrate","config":{"workflow":{"start_step":"generate","steps":{"generate":{"action":"execute_llm_prompt","config":{"prompt_template":"Write a short greeting for {{name}}","input_fields":["name"]},"next_step":"await_approval","description":"Generate content"},"await_approval":{"action":"await_approval","config":{"approval_fields":["generate"],"approval_type":"content_review","ui_config":{"title":"Content Approval Required","description":"Please approve the generated greeting"}},"next_step":"process_approval","description":"Wait for approval"},"process_approval":{"action":"process_approval_decision","next_step":"complete","description":"Process the decision"},"complete":{"action":"complete_workflow","description":"Complete"}}}},"input_data":{"name":"Test User"}}


# Step 3: Get the approval token
echo ""
echo "STEP 3: Extract the approval token"
echo "-----------------------------------"
echo "Look at Terminal 1 (notification listener) and find the message with:"
echo "  - 'type': 'approval_request'"
echo "  - 'request_id': '<APPROVAL_TOKEN>'"
echo "  - 'orchestration_id': '<ORCHESTRATION_ID>'"
echo ""
echo "Copy these values for the next step."

# Step 4: Send approval response
echo ""
echo "STEP 4: Send the approval response"
echo "-----------------------------------"
echo "In Terminal 3, approve the request using the tokens from Step 3:"
cat << 'EOF'
# Replace these with actual values from Step 3
CORRELATION_ID="<FROM_NOTIFICATION>"
APPROVAL_TOKEN="<FROM_NOTIFICATION>"

# To APPROVE:
kubectl -n kafka run -i --rm kcat-producer-approve \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -P -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.commands.workflow.resume \
-H correlation_id=$CORRELATION_ID \
-H in_response_to_request_id=$APPROVAL_TOKEN \
-H message_type=response \
-H status=complete \
<<JSON
{
"success": true,
"body": {
"approved": true,
"comments": "Looks good, approved!",
"approved_by": "test-user@example.com",
"modified_data": {}
}
}
JSON

# To REJECT:
kubectl -n kafka run -i --rm kcat-producer-reject \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -P -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.commands.workflow.resume \
-H correlation_id=$CORRELATION_ID \
-H in_response_to_request_id=$APPROVAL_TOKEN \
-H message_type=response \
-H status=complete \
<<JSON
{
"success": true,
"body": {
"approved": false,
"comments": "Content needs revision - too informal",
"approved_by": "test-user@example.com"
}
}
JSON
EOF

# Step 5: Monitor the workflow completion
echo ""
echo "STEP 5: Monitor workflow completion"
echo "------------------------------------"
echo "Start a consumer to watch for the final response:"
cat << 'EOF'
kubectl -n kafka run -i --rm kcat-consumer-responses \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -C -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.responses -o json | jq '.'
EOF

# Advanced Example: Multi-step approval with content modification
echo ""
echo "ADVANCED EXAMPLE: Approval with content modification"
echo "----------------------------------------------------"
cat << 'EOF'
# When approving, you can include modified data:
{
"success": true,
"body": {
"approved": true,
"comments": "Approved with minor edits",
"approved_by": "editor@example.com",
"modified_data": {
"content": "Hello Test User! Welcome to our platform. We're excited to have you here!",
"edited": true,
"edit_notes": "Made the greeting more welcoming"
}
}
}
EOF

# Testing with the content-writer-with-approval agent
echo ""
echo "TESTING WITH CONTENT WRITER AGENT"
echo "---------------------------------"
cat << 'EOF'
# Trigger the content writer with approval workflow:
CORRELATION_ID=$(uuidgen)
REQUEST_ID=$(uuidgen)

kubectl -n kafka run -i --rm kcat-producer-content \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -P -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.content-writer-with-approval.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_type=request \
-H action=process \
<<JSON
{
"action": "process",
"input_data": {
"topic": "AI in Healthcare",
"content_type": "blog post introduction",
"tone": "professional yet accessible",
"keywords": ["artificial intelligence", "medical diagnosis", "patient care"]
}
}
JSON
EOF

echo ""
echo "DEBUGGING TIPS"
echo "--------------"
echo "1. Check agent logs for approval token generation:"
echo "   kubectl logs -n <namespace> <agent-pod> | grep 'AwaitApprovalAction'"
echo ""
echo "2. Verify notification was sent:"
echo "   Look for 'Sent approval request' in agent logs"
echo ""
echo "3. Check orchestration status:"
echo "   Look for 'AWAITING_RESPONSE' status in coordinator logs"
echo ""
echo "4. Verify approval processing:"
echo "   Look for 'ProcessApprovalDecisionAction' in logs after sending approval"
echo ""
echo "5. Common issues:"
echo "   - Wrong correlation_id: Must match the original request"
echo "   - Wrong approval_token: Must match the request_id from notification"
echo "   - Topic mismatch: Ensure using system.commands.workflow.resume"
echo ""

echo "INTEGRATION NOTES"
echo "-----------------"
echo "For production use:"
echo "1. Build a UI that subscribes to system.notifications.ui"
echo "2. Display pending approvals with the data from notifications"
echo "3. Send properly formatted responses to system.commands.workflow.resume"
echo "4. Store approval history in database for audit trail"
echo "5. Implement timeout handling for approvals"
echo "6. Add role-based access control for different approval types"