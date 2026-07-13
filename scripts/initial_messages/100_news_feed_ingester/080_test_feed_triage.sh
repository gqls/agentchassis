SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "TRIAGE CORRELATION_ID=$CORRELATION_ID"

kubectl -n kafka run -i --rm kcat-triage-test-$$ \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_id=$MESSAGE_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H orchestration_name=feed-triage-test \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=cli \
    -H from_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses <<JSON
{"headers":{"correlation_id":"$CORRELATION_ID","orchestration_id":"$ORCHESTRATION_ID","message_type":"request","action":"orchestrate","client_id":"demo_client","message_id":"$MESSAGE_ID","request_id":"$REQUEST_ID","timestamp":"$TIMESTAMP","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"}},"config":{"workflow":{"start_step":"spawn_triage","steps":{"spawn_triage":{"action":"spawn_agent","config":{"agent_type":"feed-triage","role":"triage-test"},"next_step":"call_triage"},"call_triage":{"action":"call_agent","config":{"target_role":"triage-test","input_mapping":{"site_id":"input_data.site_id"}},"next_step":"complete"},"complete":{"action":"complete_workflow"}}},"processing_mode":"orchestrator","timeout_seconds":300},"input_data":{"site_id":"$SITE_ID"}}
JSON


check after 60 timeout_seconds-- Should show items moved from ingested to relevant/rejected
SELECT status, COUNT(*) FROM content_feed_items
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
GROUP BY status;

-- See the scores
SELECT source_title, relevance_score, status, topics
FROM content_feed_items
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
AND relevance_score IS NOT NULL
ORDER BY relevance_score DESC;

-- Check if the triage orchestration ran at all
SELECT status, current_step, error,
       collected_data->'pending_items'->>'count' as item_count,
       collected_data->'triage_result' as triage_result
FROM orchestration_states
WHERE collected_data->'input_data'->>'site_id' = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND (correlation_id = '<paste-triage-correlation-id>'
       OR collected_data->'agent_config'->'workflow'->'steps'->'load_items' IS NOT NULL)
ORDER BY created_at DESC
LIMIT 5;



-- What did score_relevance (the LLM step) produce?
SELECT
    collected_data->'score_relevance' as score_relevance_output,
    substring(collected_data->'pending_items'->>'items', 1, 500) as items_preview,
    collected_data->'pending_items'->>'count' as item_count
FROM orchestration_states
WHERE correlation_id = '1dd83396-88c9-42bf-aa63-513be9740061'
  AND status = 'COMPLETED'
  AND collected_data->'pending_items'->>'count' = '34';

  -- See what the LLM step actually produced
  SELECT
      collected_data->'scores' as scores_field,
      collected_data->'triage_scores' as triage_scores,
      substring(collected_data->'scores'->>'result', 1, 500) as llm_result_preview,
      jsonb_object_keys(collected_data) as top_keys
  FROM orchestration_states
  WHERE correlation_id = '1dd83396-88c9-42bf-aa63-513be9740061'
    AND status = 'COMPLETED'
    AND collected_data->'pending_items'->>'count' = '34'
  LIMIT 1;