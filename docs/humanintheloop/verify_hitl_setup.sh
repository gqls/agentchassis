#!/bin/bash

# HITL System Verification Script
# Checks if all HITL components are properly configured

echo "========================================="
echo "HITL System Verification"
echo "========================================="
echo ""

# Check if agent definitions exist
echo "1. Checking agent definitions in database..."
kubectl exec -it postgres-clients-0 -n personae -- psql -U clients_user -d clients_db -t -c \
    "SELECT 
        CASE 
            WHEN EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'simple-content-writer') 
            THEN '✓ Simple content writer agent found' 
            ELSE '✗ Simple content writer agent NOT found - run hitl_agent_definition.sql' 
        END;" 2>/dev/null

kubectl exec -it postgres-clients-0 -n personae -- psql -U clients_user -d clients_db -t -c \
    "SELECT 
        CASE 
            WHEN EXISTS (SELECT 1 FROM agent_group WHERE type = 'content-approval-hitl') 
            THEN '✓ Content approval HITL group found' 
            ELSE '✗ Content approval HITL group NOT found - run hitl_agent_group_definition.sql' 
        END;" 2>/dev/null

echo ""
echo "2. Checking Kafka topics..."
# Check if required topics exist
TOPICS=$(kubectl -n kafka run -i --rm kcat-list --image=edenhill/kcat:1.7.1 --restart=Never -- kcat -L -b personae-kafka-cluster-kafka-bootstrap:9092 2>/dev/null | grep "topic" | grep -E "(system.notifications.ui|system.commands.workflow.resume)" | wc -l)

if [ "$TOPICS" -ge "2" ]; then
    echo "✓ Required HITL topics found:"
    echo "  - system.notifications.ui"
    echo "  - system.commands.workflow.resume"
else
    echo "✗ Some required topics may be missing"
fi

echo ""
echo "3. Checking agent chassis pods..."
CHASSIS_PODS=$(kubectl get pods -n personae -l app=agent-chassis --no-headers 2>/dev/null | wc -l)
if [ "$CHASSIS_PODS" -gt "0" ]; then
    echo "✓ Agent chassis pods running: $CHASSIS_PODS"
else
    echo "✗ No agent chassis pods found"
fi

echo ""
echo "========================================="
echo "Setup Instructions (if needed):"
echo "========================================="
echo ""
echo "1. Load agent definitions:"
echo "   kubectl cp hitl_agent_definition.sql postgres-clients-0:/tmp/ -n personae"
echo "   kubectl exec -it postgres-clients-0 -n personae -- psql -U clients_user -d clients_db -f /tmp/hitl_agent_definition.sql"
echo ""
echo "2. Load agent group:"
echo "   kubectl cp hitl_agent_group_definition.sql postgres-clients-0:/tmp/ -n personae"
echo "   kubectl exec -it postgres-clients-0 -n personae -- psql -U clients_user -d clients_db -f /tmp/hitl_agent_group_definition.sql"
echo ""
echo "========================================="