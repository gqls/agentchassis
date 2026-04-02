SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"

SITE_ID='5fe15466-4e2e-4ff2-981e-98c1b7074002'
DOMAIN='gaswholesalers.com'
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "ORCHESTRATOR CORRELATION_ID=$CORRELATION_ID"

kubectl -n kafka run -i --rm kcat-orch-test-$$ \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_id=$MESSAGE_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H orchestration_name=feed-orchestrator-test \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=cli \
    -H from_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses <<JSON
{"headers":{"correlation_id":"$CORRELATION_ID","orchestration_id":"$ORCHESTRATION_ID","message_type":"request","action":"orchestrate","client_id":"demo_client","message_id":"$MESSAGE_ID","request_id":"$REQUEST_ID","timestamp":"$TIMESTAMP","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"}},"config":{"workflow":{"start_step":"spawn_orchestrator","steps":{"spawn_orchestrator":{"action":"spawn_agent","config":{"agent_type":"content-feed-orchestrator","role":"orch-test"},"next_step":"call_orchestrator"},"call_orchestrator":{"action":"call_agent","config":{"target_role":"orch-test","input_mapping":{"site_id":"input_data.site_id"}},"next_step":"complete"},"complete":{"action":"complete_workflow"}}},"processing_mode":"orchestrator","timeout_seconds":600},"input_data":{"site_id":"$SITE_ID"}}
JSON

# see if it worked, should be in git under data/latest-news.json in gaswholesalers.com git repo
SELECT
    default_config->'workflow'->'steps'->'render_news_json' IS NOT NULL as has_render,
    default_config->'workflow'->'steps'->'commit_news'->'action' as commit_action
FROM agent_definitions WHERE type = 'content-feed-orchestrator' AND deleted_at IS NULL;
 has_render | commit_action
------------+---------------
 t          | "git_commit"
(1 row)


SELECT orchestration_id, owner_agent_type, current_step, status, created_at,
       collected_data->'seed_result'->>'has_sources' as has_sources,
       collected_data->'seed_result'->>'existing_count' as existing_count,
       collected_data->'dispatch_result'->>'source_count' as dispatched,
       collected_data->'news_render_result'->>'item_count' as rendered_items,
       error
FROM orchestration_states
WHERE owner_agent_type = 'content-feed-orchestrator'
ORDER BY created_at DESC
LIMIT 3;

-- Let's see the full results:
SELECT
    collected_data->'seed_result'->>'has_sources' as has_sources,
    collected_data->'seed_result'->>'existing_count' as existing_sources,
    collected_data->'dispatch_result'->>'source_count' as dispatched,
    collected_data->'triage_result' as triage,
    collected_data->'news_render_result'->>'item_count' as rendered_items,
    collected_data->'news_commit_result'->>'commit_sha' as commit_sha,
    current_step, status
FROM orchestration_states
WHERE orchestration_id = '143ffb61-32dd-4d02-bcca-7eee7c4b7676';

-- check if the news JSON was committed:
-- Check if feed items got re-scored
SELECT status, COUNT(*)
FROM content_feed_items
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
GROUP BY status
ORDER BY status;