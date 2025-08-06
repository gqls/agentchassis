#!/bin/bash
# test/e2e/scripts/cleanup_e2e.sh

set -e

echo "Cleaning up E2E test environment..."

# Delete test resources
kubectl delete namespace test-e2e --ignore-not-found

# Clean test data from database
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db <<EOF
DELETE FROM orchestrator_state WHERE correlation_id LIKE 'test-e2e-%';
DELETE FROM client_demo_client.agent_instances WHERE name LIKE 'test-e2e-%';
VACUUM;
EOF

echo "E2E cleanup complete"