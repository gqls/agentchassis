# Human-in-the-Loop (HITL) Testing Guide

## Overview
This demonstrates a simple HITL workflow where:
1. An agent generates content
2. The workflow pauses and requests human approval
3. A human reviews and approves/rejects the content
4. The workflow resumes and completes

## Prerequisites
1. Load the agent definitions into the database:
```bash
# Load the agent group definition
kubectl exec -it postgres-clients-0 -n personae -- psql -U clients_user -d clients_db -f /tmp/hitl_agent_group_definition.sql

# Load the agent definition
kubectl exec -it postgres-clients-0 -n personae -- psql -U clients_user -d clients_db -f /tmp/hitl_agent_definition.sql
```

## Testing the HITL Flow

### Step 1: Start the Approval Listener
In Terminal 1, start listening for approval requests:
```bash
chmod +x listen_for_approvals.sh
./listen_for_approvals.sh
```

### Step 2: Trigger the Workflow
In Terminal 2, start the HITL workflow:
```bash
chmod +x start_hitl_workflow.sh
./start_hitl_workflow.sh
```

Note the **Correlation ID** printed at the end.

### Step 3: Monitor the Approval Request
In Terminal 1, you should see an approval notification appear with:
- Headers containing the `correlation_id`
- A `request_id` (this is your approval token)
- The generated content that needs approval

Example notification structure:
```json
{
  "headers": {
    "correlation_id": "abc-123...",
    "request_id": "xyz-789...",  // This is your approval token
    "orchestration_id": "...",
    "reply_to_topic": "system.commands.workflow.resume"
  },
  "body": {
    "type": "content_approval",
    "title": "Content Approval Required",
    "description": "Please review and approve the generated content",
    "data": {
      "generated_content": "...",
      "business_name": "TechFlow Solutions",
      "business_type": "AI automation and workflow optimization"
    }
  }
}
```

### Step 4: Send Approval Response
In Terminal 3:
1. Edit the `send_approval.sh` script
2. Replace the placeholder values:
   - `CORRELATION_ID` - from the notification headers
   - `APPROVAL_TOKEN` - the `request_id` from the notification
3. Run the approval script:
```bash
chmod +x send_approval.sh
./send_approval.sh
```

### Step 5: Observe Workflow Completion
The workflow will resume and complete. You can monitor the logs:
```bash
# Check orchestration state
kubectl exec -it postgres-clients-0 -n personae -- psql -U clients_user -d clients_db -c \
  "SELECT correlation_id, current_step, status, execution_metadata \
   FROM orchestrator_state \
   WHERE correlation_id = 'YOUR_CORRELATION_ID';"

# View agent logs
kubectl logs -n personae -l agent-type=simple-content-writer --tail=50
```

## Workflow States

1. **RUNNING** - Workflow is executing steps
2. **AWAITING_RESPONSE** - Workflow paused, waiting for approval
3. **COMPLETED** - Workflow finished successfully
4. **FAILED** - Workflow encountered an error

## Customization

### Modifying the Approval Step
Edit the agent definition to change approval behavior:
- `timeout_seconds` - How long to wait for approval (default: 300s)
- `notification_data` - What information to include in the approval request
- `include_generated_content` - Whether to include the generated content

### Conditional Approval
The workflow can branch based on approval:
```json
"process_approval": {
  "action": "conditional_branch",
  "config": {
    "condition": "{{.await_human_approval.approved}}",
    "true_step": "finalize_content",
    "false_step": "regenerate_content"
  }
}
```

## Troubleshooting

1. **No approval notification received**
   - Check the listener is connected to the right topic
   - Verify the agent has the `await_approval` step
   - Check agent logs for errors

2. **Workflow doesn't resume after approval**
   - Ensure the `in_response_to_request_id` matches the approval token
   - Verify the correlation_id is correct
   - Check the resume topic is `system.commands.workflow.resume`

3. **Timeout errors**
   - Increase `timeout_seconds` in the approval step config
   - Send approval response faster

## API Integration (Future)

Once the API is ready, the manual approval process can be replaced with:
- REST endpoint to fetch pending approvals
- Web UI for reviewing content
- API call to approve/reject with the approval token

The underlying mechanism remains the same - just the interface changes from manual Kafka messages to API calls.