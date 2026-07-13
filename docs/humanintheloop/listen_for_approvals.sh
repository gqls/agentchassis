#!/bin/bash

# HITL Approval Notification Listener
# This script listens for approval requests on the system.notifications.ui topic

echo "========================================="
echo "HITL Approval Notification Listener"
echo "========================================="
echo "Listening for approval requests..."
echo "Press Ctrl+C to stop"
echo ""
echo "When you see an approval request, you'll need:"
echo "  - correlation_id (from headers)"
echo "  - request_id (the approval token)"
echo ""
echo "========================================="
echo ""

kubectl -n kafka run -i --rm kcat-consumer-hitl \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -C \
  -b personae-kafka-cluster-kafka-bootstrap:9092 \
  -t system.notifications.ui \
  -f 'Headers: %h\nKey: %k\nPayload: %s\n------------------\n' \
  -o end