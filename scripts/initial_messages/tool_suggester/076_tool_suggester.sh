kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-console-producer.sh \
--bootstrap-server localhost:9092 \
--topic system.agent.generic.requests \
--property "parse.headers=true" \
--property "headers.delimiter=|" <<'EOF'
correlation_id:tool-eval-gas-001|request_id:tool-eval-gas-001|message_type:request|action:orchestrate|from_agent_type:manual|responses_topic:system.scheduler.responses	{"action":"orchestrate","config":{"agent_type":"tool-suggester"},"input_data":{"site_id":"5fe15466-4e2e-4ff2-981e-98c1b7074002","domain":"gaswholesalers.co.uk"}}
EOF

--


SELECT id, domain, status FROM sites WHERE domain ILIKE '%gaswholesaler%';


That spawns a tool-suggester pod which runs the full workflow: reads site specs, loads pages, loads library, calls the LLM, and creates add_tool work items for each suggestion.
To watch it:

# Pod appears
kubectl -n ai-persona-system get pods | grep tool-suggester

# Logs
kubectl -n ai-persona-system logs -f -l agent-type=tool-suggester

# Check what it created
psql -c "SELECT item_type, summary, handler_agent, status, spec->>'function' as tool_fn
FROM site_work_items
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND source = 'tool-suggester'
ORDER BY created_at DESC;"


The add_tool items it creates will have handler_agent = 'tool-deployer', so they get picked up on the next dispatch sweep — or you can manually trigger tool-deployer the same way if you want to deploy one immediately.
If you'd rather go through the work item system (so it shows up in the normal pipeline logs), insert the item and wait for dispatch:

INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
) VALUES (
    '5fe15466-4e2e-4ff2-981e-98c1b7074002',
    'manual', 'build', 'evaluate_tools', 'low',
    'Evaluate tool needs for gaswholesalers.co.uk',
    '{}'::jsonb,
    100, 'tool-suggester', 'triaged', 'admin',
    'evaluate_tools_gaswholesalers'
);

