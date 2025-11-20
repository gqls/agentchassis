#!/bin/bash

# HITL Approval Response Script
# Update the CORRELATION_ID and APPROVAL_TOKEN with values from the notification

# REPLACE THESE VALUES with the ones from the approval notification
CORRELATION_ID="REPLACE_WITH_CORRELATION_ID"
APPROVAL_TOKEN="REPLACE_WITH_REQUEST_ID_FROM_NOTIFICATION"

# Optional: Modify the approval decision and comments
APPROVED="true"  # Set to false to reject
COMMENTS="Content approved via manual HITL review"

echo "========================================="
echo "Sending HITL Approval Response"
echo "========================================="
echo "  Correlation ID: $CORRELATION_ID"
echo "  Approval Token: $APPROVAL_TOKEN"
echo "  Approved:       $APPROVED"
echo "  Comments:       $COMMENTS"
echo "========================================="
echo ""

# Send the approval response
kubectl -n kafka run -i --rm kcat-producer-approval \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap:9092 \
  -t system.commands.workflow.resume \
  -H correlation_id=$CORRELATION_ID \
  -H in_response_to_request_id=$APPROVAL_TOKEN \
  -H message_type=response \
  -H status=complete \
  <<EOF
{
  "success": true,
  "body": {
    "approved": $APPROVED,
    "comments": "$COMMENTS",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "approved_by": "manual_hitl_user"
  }
}
EOF

echo ""
echo "========================================="
echo "Approval response sent!"
echo "The workflow should now resume and complete."
echo "========================================="