SITE_ID="<PASTE_SITE_ID>"
SOURCE_ID="<PASTE_SOURCE_ID>"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "CORRELATION_ID=$CORRELATION_ID"
echo "ORCHESTRATION_ID=$ORCHESTRATION_ID"

kubectl -n kafka run -i --rm kcat-feed-test-$$ \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_id=$MESSAGE_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H orchestration_name=feed-rss-test \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=cli \
    -H from_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses <<JSON
{
  "headers": {
    "correlation_id": "$CORRELATION_ID",
    "orchestration_id": "$ORCHESTRATION_ID",
    "message_type": "request",
    "action": "orchestrate",
    "client_id": "demo_client",
    "message_id": "$MESSAGE_ID",
    "request_id": "$REQUEST_ID",
    "timestamp": "$TIMESTAMP",
    "sender": {"agent_id": "cli-user", "agent_type": "cli", "pod_name": "cli"}
  },
  "config": {
    "workflow": {
      "start_step": "spawn_ingester",
      "steps": {
        "spawn_ingester": {
          "action": "spawn_agent",
          "config": {"agent_type": "feed-ingester", "role": "rss-test"},
          "next_step": "call_ingester",
          "description": "Spawn feed-ingester"
        },
        "call_ingester": {
          "action": "call_agent",
          "config": {
            "target_role": "rss-test",
            "input_mapping": {
              "site_id": "input_data.site_id",
              "source_id": "input_data.source_id",
              "source_type": "input_data.source_type",
              "source_config": "input_data.source_config"
            }
          },
          "next_step": "complete",
          "description": "Call feed-ingester with source config"
        },
        "complete": {"action": "complete_workflow"}
      }
    },
    "processing_mode": "orchestrator",
    "timeout_seconds": 300
  },
  "input_data": {
    "site_id": "$SITE_ID",
    "source_id": "$SOURCE_ID",
    "source_type": "rss",
    "source_name": "Test BBC Boxing RSS",
    "source_config": {
      "feed_url": "https://feeds.bbci.co.uk/sport/boxing/rss.xml",
      "max_items": 5
    }
  }
}
JSON





--- monitor
# Watch the generic agent pick it up
kubectl -n ai-persona-system logs deployment/agent-chassis --tail=100 -f | grep "$CORRELATION_ID"

# After ~10 seconds, check if a feed-ingester job was spawned
kubectl -n ai-persona-system get jobs | grep feed-ingester

# Check the feed-ingester pod logs
kubectl -n ai-persona-system logs -l agent-type=feed-ingester --tail=100


---- check results
# Did items land in the database?
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
    SELECT source_title, source_url, status, created_at
    FROM content_feed_items
    WHERE site_id = '$SITE_ID'
    ORDER BY created_at DESC LIMIT 10;"

# Was the source timestamp updated?
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
    SELECT name, last_fetched_at, next_fetch_at, error_count, last_error
    FROM content_sources
    WHERE id = '$SOURCE_ID';"

# Check orchestration state if something went wrong
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
    SELECT orchestration_id, status, current_step, error
    FROM orchestration_states
    WHERE correlation_id = '$CORRELATION_ID'
    ORDER BY created_at;"



What you're looking for at each stage:

Step 6 produces the Kafka message → generic agent picks it up
Generic agent runs spawn_agent → a feed-ingester K8s job appears
Generic agent runs call_agent → sends process message to the ingester
Ingester runs conditional_route → routes to fetch_rss step
fetch_rss does HTTP GET to BBC RSS → parses XML → returns items
write_feed_items inserts into content_feed_items → rows appear
update_source_timestamps updates content_sources → last_fetched_at set, next_fetch_at moved forward

If it stalls, the most likely failure points are: the new actions not being in the compiled binary (step 3 check), the agent definition not being found (step 2 check), or the conditional_route not finding source_type in input_data (check orchestration_states collected_data).

