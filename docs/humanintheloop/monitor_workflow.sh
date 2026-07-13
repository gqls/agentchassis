#!/bin/bash

# HITL Workflow Monitor
# This script monitors the status of a HITL workflow

if [ -z "$1" ]; then
    echo "Usage: $0 <CORRELATION_ID>"
    echo "Example: $0 abc-123-def-456"
    exit 1
fi

CORRELATION_ID=$1

echo "========================================="
echo "HITL Workflow Monitor"
echo "Correlation ID: $CORRELATION_ID"
echo "========================================="
echo ""

while true; do
    clear
    echo "========================================="
    echo "HITL Workflow Status Monitor"
    echo "Correlation ID: $CORRELATION_ID"
    echo "Time: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "========================================="
    echo ""
    
    # Check orchestration state
    echo "Current Status:"
    kubectl exec -it postgres-clients-0 -n personae -- psql -U clients_user -d clients_db -t -c \
        "SELECT 
            'Status: ' || status || E'\\n' ||
            'Current Step: ' || current_step || E'\\n' ||
            'Last Activity: ' || updated_at || E'\\n' ||
            'Awaited Requests: ' || COALESCE(jsonb_pretty(awaited_requests), 'None') 
         FROM orchestrator_state 
         WHERE correlation_id = '$CORRELATION_ID';" 2>/dev/null
    
    echo ""
    echo "----------------------------------------"
    echo "Collected Data Summary:"
    kubectl exec -it postgres-clients-0 -n personae -- psql -U clients_user -d clients_db -t -c \
        "SELECT 
            jsonb_pretty(
                jsonb_build_object(
                    'generated_content', collected_data->'generate_draft'->'result',
                    'approval_status', collected_data->'await_human_approval'->'approved',
                    'approval_comments', collected_data->'await_human_approval'->'comments'
                )
            )
         FROM orchestrator_state 
         WHERE correlation_id = '$CORRELATION_ID';" 2>/dev/null
    
    echo ""
    echo "========================================="
    echo "Press Ctrl+C to stop monitoring"
    echo "Refreshing in 5 seconds..."
    
    sleep 5
done