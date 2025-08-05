# check the monitoring:
# Your monitoring system is now fully operational! You can track:

- Active workflows in progress
- Completed workflows with full execution history
- Stuck workflows that haven't progressed
- Overall metrics and success rates
- Detailed step-by-step execution paths

# Port-forward to the health server
kubectl port-forward -n ai-persona-system deployment/agent-chassis 8080:8080

# In another terminal, check active workflows
curl http://localhost:8080/monitor/workflows?client_id=demo_client | jq .

# Check if there are any errors
kubectl exec -it kafka-client-test -n kafka -- kafka-console-consumer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.errors.generic \
--max-messages 5

# Check the orchestrator state directly
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"SELECT correlation_id, status, current_step,
execution_metadata->>'completed_steps' as completed,
execution_metadata->>'total_steps' as total
FROM orchestrator_state
WHERE correlation_id LIKE 'test-%'
ORDER BY created_at DESC
LIMIT 5;"

# Also, make sure the workflow is configured properly:
# Check what workflow is configured for the generic agent
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"SELECT config->'workflow' as workflow
FROM client_demo_client.agent_instances
WHERE id = '00000000-0000-0000-0000-000000000001';"