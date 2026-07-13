#!/bin/bash
# test/e2e/scripts/cleanup_e2e.sh
#!/bin/bash
# test/e2e/scripts/cleanup_e2e.sh

set -e

echo "Cleaning up E2E test environment..."

# Delete test resources
kubectl delete namespace test-e2e --ignore-not-found

# Clean test data from database using the new pattern
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db <<EOF
-- Clean up test UUIDs (those starting with 0000)
DELETE FROM orchestrator_state WHERE correlation_id::text LIKE '0000%';
DELETE FROM client_demo_client.agent_instances WHERE id::text LIKE '0000%';
DELETE FROM agent_groups WHERE id::text LIKE '0000%';

-- Also clean legacy test- prefixed data
DELETE FROM orchestrator_state WHERE correlation_id LIKE 'test-e2e-%';
DELETE FROM client_demo_client.agent_instances WHERE name LIKE 'test-e2e-%';

-- Run cleanup function
SELECT cleanup_all_test_data();

VACUUM;
EOF

echo "E2E cleanup complete"