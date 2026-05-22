

clients_db=# -- Pull every distinct action string referenced in any active agent's workflow JSON
SELECT DISTINCT regexp_matches(default_config::text, '"action"\s*:\s*"([a-z_]+)"', 'g') AS action_ref
FROM agent_definitions
WHERE status = 'active'
ORDER BY action_ref;
