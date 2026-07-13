# HITL (Human-in-the-Loop) Response Flow

## How the HITL Request Works

When the content-reviewer reaches `escalate_to_human` step:

1. **Notification Sent**: A message is published to `system.notifications.ui` topic
2. **Workflow Paused**: The agent sets `await_response: true` and waits
3. **Awaited Request Created**: An entry is added to track the pending request

## Notification Message Structure

The notification sent to `system.notifications.ui` looks like:

```json
{
  "type": "input_request",
  "request_type": "review",
  "request_id": "<uuid>",
  "orchestration_id": "<content-reviewer-orchestration-id>",
  "correlation_id": "d6bc7920-6fad-47d4-a4ac-2777d193432a",
  "agent_type": "content-reviewer",
  "agent_id": "<agent-id>",
  "step_name": "escalate_to_human",
  "reply_to_topic": "job.d6bc7920-xxxx-content-reviewer-xxx.responses",
  "data": { /* page content and issues */ },
  "message": "Auto-review found issues with index - human review required",
  "timeout_seconds": 3600,
  "editable": true,
  "ui_config": {
    "title": "Content Review - Issues Found",
    "description": "Auto-review flagged issues. Please review and fix.",
    "show_issues": true,
    "issues_field": "eval_result.issues"
  }
}
```

## What Needs to Consume system.notifications.ui

You need a UI service that:
1. Subscribes to `system.notifications.ui` topic
2. Displays notifications to humans (web UI, Slack, email, etc.)
3. Collects the human's response
4. Publishes the response to the `reply_to_topic`

## How to Respond to Continue the Flow

The human (or a service on behalf of the human) needs to send a message to the `reply_to_topic` specified in the notification.

### Response Message Structure

Send to: The `reply_to_topic` from the notification (e.g., `job.d6bc7920-xxxx-content-reviewer-xxx.responses`)

```json
{
  "headers": {
    "correlation_id": "d6bc7920-6fad-47d4-a4ac-2777d193432a",
    "request_id": "<new-uuid>",
    "message_type": "response",
    "in_response_to_request_id": "<request_id from notification>",
    "in_response_to_step_name": "escalate_to_human",
    "orchestration_id": "<from notification>",
    "status": "complete",
    "is_complete": "true"
  },
  "body": {
    "approved": true,
    "status": "approved",
    "responded_by": "user@example.com",
    "edits": {
      // Optional: any edits made to the content
    },
    "comments": "Looks good, approved with minor edits"
  }
}
```

### Kafka Headers (as separate header fields)

```
correlation_id: d6bc7920-6fad-47d4-a4ac-2777d193432a
request_id: <new-uuid>
message_type: response
in_response_to_request_id: <request_id from notification>
in_response_to_step_name: escalate_to_human
status: complete
is_complete: true
```

## Manual Testing with kafkacat/kcat

To manually respond and unblock a workflow:

```bash
# First, find the notification that was sent
kubectl exec -it kafka-pod -- kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic system.notifications.ui \
  --from-beginning

# Extract the reply_to_topic and request_id from the notification

# Then send a response
echo '{"headers":{"correlation_id":"d6bc7920-6fad-47d4-a4ac-2777d193432a","request_id":"'$(uuidgen)'","message_type":"response","in_response_to_request_id":"REQUEST_ID_FROM_NOTIFICATION","status":"complete","is_complete":"true"},"body":{"approved":true,"status":"approved","responded_by":"manual-test"}}' | \
kubectl exec -i kafka-pod -- kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic REPLY_TO_TOPIC_FROM_NOTIFICATION
```

## Why There Were No Awaited Requests

Looking at the logs:
```
15:28:38 - content-reviewer reaches "escalate_to_human"
15:33:20 - pageflow-builder times out (5 min timeout on call_agent)
15:33:33 - content-reviewer: "Cleaned up expired awaited requests" count=1
```

The awaited request DID exist but:
1. The **parent** (pageflow-builder) timed out first (5 min)
2. Parent sent a retry with `body: null`
3. Content-reviewer's awaited request was cleaned up as "expired"

The fix is to increase the parent's timeout so it doesn't time out before the child's HITL can complete.

## Current Problem: No UI Service

It appears there is no service consuming `system.notifications.ui` to present HITL requests to humans. You need to either:

1. **Build a HITL UI service** that:
    - Consumes `system.notifications.ui`
    - Displays requests in a web interface
    - Allows humans to respond
    - Publishes responses to the appropriate topic

2. **Or use auto-approval** by modifying the content-reviewer workflow to:
    - Set higher approval threshold in auto_eval
    - Skip HITL for most cases
    - Only escalate for truly problematic content

## Checking for Pending HITL Requests

```sql
-- Check for pending input requests in the database
SELECT * FROM input_requests 
WHERE status = 'pending' 
ORDER BY created_at DESC;

-- Or check awaited_requests table
SELECT * FROM awaited_requests
WHERE status = 'waiting'
ORDER BY timeout_at DESC;
```