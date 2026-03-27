SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
SOURCE_ID="<paste-actual-id>"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "CORRELATION_ID=$CORRELATION_ID"

kubectl -n kafka run -i --rm kcat-scrape-test2-$$ \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_id=$MESSAGE_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H orchestration_name=feed-scrape-test2 \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=cli \
    -H from_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses <<JSON
{"headers":{"correlation_id":"$CORRELATION_ID","orchestration_id":"$ORCHESTRATION_ID","message_type":"request","action":"orchestrate","client_id":"demo_client","message_id":"$MESSAGE_ID","request_id":"$REQUEST_ID","timestamp":"$TIMESTAMP","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"}},"config":{"workflow":{"start_step":"spawn_ingester","steps":{"spawn_ingester":{"action":"spawn_agent","config":{"agent_type":"feed-ingester","role":"scrape-test"},"next_step":"call_ingester","description":"Spawn feed-ingester"},"call_ingester":{"action":"call_agent","config":{"target_role":"scrape-test","input_mapping":{"site_id":"input_data.site_id","source_id":"input_data.source_id","source_type":"input_data.source_type","source_config":"input_data.source_config"}},"next_step":"complete","description":"Call feed-ingester"},"complete":{"action":"complete_workflow"}}},"processing_mode":"orchestrator","timeout_seconds":300},"input_data":{"site_id":"$SITE_ID","source_id":"$SOURCE_ID","source_type":"scrape","source_name":"OilPrice Latest News Scrape","source_config":{"url":"https://oilprice.com/Latest-Energy-News/World-News","scrape_config":{"only_main_content":true},"max_items":5}}}
JSON


# get scrape source id
# Get the scrape source ID
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
    SELECT id FROM content_sources
    WHERE name = 'OilPrice Latest News Scrape';"


    monitor

# Watch the feed-ingester job
kubectl -n ai-persona-system get jobs | grep feed-ingester

# Tail the ingester logs — look for firecrawl_scrape and normalize
kubectl -n ai-persona-system logs -l agent-type=feed-ingester --tail=200 -f

# Check orchestration state after ~30s (scrape takes longer than RSS)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
    SELECT status, current_step,
           collected_data->'write_result'->>'written' as written,
           collected_data->'write_result'->>'skipped' as skipped
    FROM orchestration_states
    WHERE correlation_id = '$CORRELATION_ID'
      AND collected_data->'write_result' IS NOT NULL;"