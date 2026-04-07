#!/bin/bash
# Test script for tool-recreation-handler with gamedesign.uk
# Run each section separately and check output before proceeding
DANGEROUS CLEARANCE
SITE_ID="15a6cb16-5a86-4541-a8e4-d7106239b6a4"
DB_CMD="kubectl -n ai-persona-system exec -it deploy/agent-coordinator -- psql -U postgres -d clients_db"

echo "=== STEP 1: Insert agent definition ==="
echo "Run the tool_recreation_handler_agent_v2.sql in psql"
echo ""

echo "=== STEP 2: Verify agent definition ==="
echo "$DB_CMD -c \"SELECT type, display_name, status FROM agent_definitions WHERE type = 'tool-recreation-handler';\""
echo ""

echo "=== STEP 3: Check current specs ==="
echo "$DB_CMD -c \"SELECT aspect, length(data::text), created_at FROM site_specs WHERE site_id = '$SITE_ID' AND is_current = true ORDER BY aspect;\""
echo ""

echo "=== STEP 4: Clean old adoption data ==="
cat <<'CLEANUP'
kubectl -n ai-persona-system exec -it deploy/agent-coordinator -- psql -U postgres -d clients_db -c "
BEGIN;

-- Clear old work items from adoption (item_key has ON CONFLICT DO NOTHING so old ones block)
DELETE FROM site_work_items
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
  AND source = 'adoption';

-- Also clear improvement-loop items that built on the old adoption
DELETE FROM site_work_items
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
  AND status IN ('triaged', 'claimed');

-- Clear old adoption research results
DELETE FROM research_results
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
  AND result_type IN ('adoption_page', 'adoption_crawl');

-- Clear page_components (depends on pages)
DELETE FROM page_components
WHERE page_id IN (SELECT id FROM pages WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4');

-- Clear old pages
DELETE FROM pages
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4';

-- Clear old specs so adoption rewrites them fresh
DELETE FROM site_specs
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4';

-- Reset site status
UPDATE sites SET build_status = 'pending'
WHERE id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4';

COMMIT;
"
CLEANUP
echo ""

echo "=== STEP 5: Verify cleanup ==="
cat <<'VERIFY'
kubectl -n ai-persona-system exec -it deploy/agent-coordinator -- psql -U postgres -d clients_db -c "
SELECT 'pages' as what, count(*) FROM pages WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
UNION ALL
SELECT 'specs', count(*) FROM site_specs WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
UNION ALL
SELECT 'work_items', count(*) FROM site_work_items WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
UNION ALL
SELECT 'research', count(*) FROM research_results WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4' AND result_type IN ('adoption_page','adoption_crawl');
"
VERIFY
echo ""

echo "=== STEP 6: Trigger adoption ==="
cat <<'TRIGGER'
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "CORRELATION_ID=$CORRELATION_ID"
echo "Save this ^^"

kubectl -n kafka run -i --rm kcat-adopt-gamedesign \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.site-adoption-agent.process \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H client_id=cli-user \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.responses.site-adoption-agent \
  -H timestamp=$TIMESTAMP <<JSON
{
  "action": "orchestrate",
  "config": {"agent_type": "site-adoption-agent"},
  "input_data": {
    "domain": "gamedesign.uk",
    "url": "https://gamedesign.uk",
    "site_id": "15a6cb16-5a86-4541-a8e4-d7106239b6a4"
  }
}
JSON
TRIGGER
echo ""

echo "=== STEP 7: Watch adoption progress ==="
echo 'kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep -E "adoption|crawl|archetype|interactive|rawHtml|raw_html|with_raw_html|apply_adoption|tool-recreation"'
echo ""

echo "=== STEP 8: After adoption completes, check results ==="
cat <<'CHECK'
# Check work items — look for tool-recreation-handler
kubectl -n ai-persona-system exec -it deploy/agent-coordinator -- psql -U postgres -d clients_db -c "
SELECT item_type, handler_agent, status, summary
FROM site_work_items
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
  AND source = 'adoption'
ORDER BY priority;
"

# Check rawHtml captured
kubectl -n ai-persona-system exec -it deploy/agent-coordinator -- psql -U postgres -d clients_db -c "
SELECT
  data->>'page_name' as page,
  length(data->'existing_content'->>'raw_markdown') as md_len,
  length(data->'existing_content'->>'raw_html') as html_len,
  data->'interactive_features' IS NOT NULL as has_features
FROM research_results
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
  AND result_type = 'adoption_page';
"

# Check archetype written
kubectl -n ai-persona-system exec -it deploy/agent-coordinator -- psql -U postgres -d clients_db -c "
SELECT aspect, length(data::text) as size
FROM site_specs
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
  AND is_current = true;
"
CHECK
echo ""

echo "=== STEP 9: Watch tool recreation dispatch ==="
echo 'kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=20 | grep -E "tool-recreation|analyze_tool|recreate_tool|check_tool_completeness|opus|completeness"'