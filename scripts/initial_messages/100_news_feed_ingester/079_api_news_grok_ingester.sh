SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
SOURCE_ID="ebd5bb96-4cc5-479c-89a7-f8015ed0c2e3"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "GROK CORRELATION_ID=$CORRELATION_ID"

kubectl -n kafka run -i --rm kcat-grok-test-$$ \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_id=$MESSAGE_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H orchestration_name=feed-grok-test \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=cli \
    -H from_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses <<JSON
{"headers":{"correlation_id":"$CORRELATION_ID","orchestration_id":"$ORCHESTRATION_ID","message_type":"request","action":"orchestrate","client_id":"demo_client","message_id":"$MESSAGE_ID","request_id":"$REQUEST_ID","timestamp":"$TIMESTAMP","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"}},"config":{"workflow":{"start_step":"spawn_ingester","steps":{"spawn_ingester":{"action":"spawn_agent","config":{"agent_type":"feed-ingester","role":"grok-test"},"next_step":"call_ingester"},"call_ingester":{"action":"call_agent","config":{"target_role":"grok-test","input_mapping":{"site_id":"input_data.site_id","source_id":"input_data.source_id","source_type":"input_data.source_type","source_config":"input_data.source_config"}},"next_step":"complete"},"complete":{"action":"complete_workflow"}}},"processing_mode":"orchestrator","timeout_seconds":300},"input_data":{"site_id":"5fe15466-4e2e-4ff2-981e-98c1b7074002","source_id":"ebd5bb96-4cc5-479c-89a7-f8015ed0c2e3","source_type":"api_news","source_name":"Grok energy news","source_config":{"provider":"xai","model":"grok-3","prompt_template":"What are the most significant energy and gas wholesale market news stories from the last {{.hours}} hours? Include oil prices, gas prices, supply disruptions, and major market moves. For each story provide: title, a 2-3 sentence summary, source attribution, and approximate time. Return as a JSON array with fields: title, summary, source_url (if known), source_name, published_at (ISO format).","hours_lookback":24,"max_items":5}}}
JSON



---


# Check items after ~30s for search, ~60s for Grok
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
    SELECT source_title, status, created_at,
           (SELECT name FROM content_sources WHERE id = cfi.source_id) as source_name
    FROM content_feed_items cfi
    WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
    ORDER BY created_at DESC LIMIT 15;"

# Check orchestration results
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
    SELECT
      collected_data->'input_data'->>'source_type' as type,
      status,
      collected_data->'write_result'->>'written' as written,
      collected_data->'write_result'->>'skipped' as skipped,
      created_at
    FROM orchestration_states
    WHERE collected_data->'input_data'->>'site_id' = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
      AND collected_data->'write_result' IS NOT NULL
    ORDER BY created_at DESC LIMIT 10;"

# If either fails, check feed-ingester logs
kubectl -n ai-persona-system logs -l agent-type=feed-ingester --tail=100