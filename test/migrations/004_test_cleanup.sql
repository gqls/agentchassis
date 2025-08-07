-- test/migrations/004_test_cleanup.sql
CREATE OR REPLACE FUNCTION cleanup_test_data()
RETURNS void AS $$
BEGIN
DELETE FROM orchestrator_state WHERE correlation_id LIKE 'test-%'
                                 AND created_at < NOW() - INTERVAL '1 day';

DELETE FROM client_demo_client.agent_instances
WHERE name LIKE 'test-%'
  AND created_at < NOW() - INTERVAL '1 day';
END;
$$ LANGUAGE plpgsql;