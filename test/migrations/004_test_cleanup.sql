-- test/migrations/004_test_cleanup.sql
-- test/migrations/004_test_cleanup.sql
CREATE OR REPLACE FUNCTION cleanup_test_data()
RETURNS void AS $$
BEGIN
    -- Clean up test UUIDs (those starting with 0000)
DELETE FROM orchestrator_state
WHERE correlation_id::text LIKE '0000%'
      AND created_at < NOW() - INTERVAL '1 day';

-- Also clean up any legacy test- prefixed data
DELETE FROM orchestrator_state
WHERE correlation_id::text LIKE 'test-%'
      AND created_at < NOW() - INTERVAL '1 day';

-- Clean up test agent instances
DELETE FROM client_demo_client.agent_instances
WHERE (id::text LIKE '0000%' OR name LIKE 'test-%')
  AND created_at < NOW() - INTERVAL '1 day';

-- Clean up test agent groups
DELETE FROM agent_groups
WHERE id::text LIKE '0000%'
      AND created_at < NOW() - INTERVAL '1 day';
END;
$$ LANGUAGE plpgsql;

-- Function to immediately clean all test data (for CI/CD)
CREATE OR REPLACE FUNCTION cleanup_all_test_data()
RETURNS void AS $$
BEGIN
    -- Clean up ALL test UUIDs immediately
DELETE FROM orchestrator_state
WHERE correlation_id::text LIKE '0000%';

-- Clean up test agent instances
DELETE FROM client_demo_client.agent_instances
WHERE id::text LIKE '0000%' OR name LIKE 'test-%';

-- Clean up test agent groups
DELETE FROM agent_groups
WHERE id::text LIKE '0000%';

RAISE NOTICE 'Cleaned up % test records', FOUND;
END;
$$ LANGUAGE plpgsql;