{
"type": "approval_request",
"request_id": "approval-token-uuid",
"orchestration_id": "orchestration-uuid",
"correlation_id": "correlation-uuid",
"agent_type": "content-writer",
"step_name": "await_approval",
"reply_to_topic": "system.commands.workflow.resume",
"data": {
"content": "Generated content to approve..."
},
"ui_config": {
"title": "Content Review Required",
"editable_fields": ["content"]
}
}

Approval Response (sent by Human/UI)
Headers:
correlation_id: original-correlation-id
in_response_to_request_id: approval-token-uuid
message_type: response

Body:
{
"success": true,
"body": {
"approved": true,
"comments": "Approved with minor edits",
"approved_by": "user@example.com",
"modified_data": {
"content": "Edited content..."
}
}
}

-------

Manual Testing with kcat

Start notification listener:

kcat -C -b kafka:9092 -t system.notifications.ui -o json

Trigger workflow with approval step
Copy approval token from notification
Send approval response:

kcat -P -b kafka:9092 -t system.commands.workflow.resume \
-H in_response_to_request_id=$APPROVAL_TOKEN \
-H correlation_id=$CORRELATION_ID

Configuration Options
AwaitApprovalAction Config

approval_fields: Fields to include in approval request
notification_topic: Topic for notifications (default: system.notifications.ui)
approval_type: Type of approval (for routing/filtering)
timeout_seconds: Approval timeout
ui_config: UI hints for approval interface

ProcessApprovalDecisionAction Config

stop_on_reject: Whether to stop workflow on rejection
rejection_handler: Step to execute on rejection


Troubleshooting
Common Issues:

Workflow doesn't pause: Check await_response: true is returned
Can't resume workflow: Verify approval token matches exactly
Missing approval data: Check CollectedData extraction logic
Notification not received: Verify topic and producer configuration

Debug Logging:

Enable debug logs for AwaitApprovalAction
Check SagaCoordinator logs for AWAITING_RESPONSE status
Monitor notification topic with kcat