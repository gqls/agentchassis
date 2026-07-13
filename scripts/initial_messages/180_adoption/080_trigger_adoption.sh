#!/bin/bash
# trigger-adopt-site.sh — Adopt an existing site into the system
#
# Usage: ./trigger-adopt-site.sh <domain> [url]
#   domain: e.g. mortgagecalculator.co.uk
#   url:    e.g. https://mortgagecalculator.co.uk (defaults to https://<domain>)

DOMAIN="${1:?Usage: ./trigger-adopt-site.sh <domain> [url]}"
URL="${2:-https://$DOMAIN}"

DOMAIN='robot-hands.com'
URL='https://robot-hands.com'

DOMAIN='gamedesign.uk'
URL='https://gamedesign.uk'

DOMAIN='gamedesign.uk'
URL='https://gamedesign.uk'

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Site Adoption Trigger"
echo "========================================="
echo "  Domain:           $DOMAIN"
echo "  URL:              $URL"
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Time:             $TIMESTAMP"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

kubectl -n kafka run -i --rm kcat-adopt-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=adopt-site-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"site-adoption-agent"},"input_data":{"domain":"$DOMAIN","url":"$URL"}}
JSON

echo ""
echo "========================================="
echo "Adoption triggered for $DOMAIN"
echo "========================================="
echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100 | grep '$CORRELATION_ID'"
echo ""
echo "Watch crawl step:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 | grep '$CORRELATION_ID' | grep -E 'crawl|webscrape|firecrawl'"
echo ""
echo "Watch LLM analysis:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 | grep '$CORRELATION_ID' | grep -E 'analyze|adoption_analysis|execute_llm'"
echo ""
echo "Watch plan application:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 | grep '$CORRELATION_ID' | grep -E 'apply_adoption|specs_written|pages_created|items_created'"
echo ""
echo "Check orchestration state:"
echo "  SELECT status, current_step, error FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid;"
echo ""
echo "Check created specs:"
echo "  SELECT aspect, source, created_at FROM site_specs WHERE site_id = (SELECT id FROM sites WHERE domain = '$DOMAIN') AND is_current = true ORDER BY aspect;"
echo ""
echo "Check created pages:"
echo "  SELECT name, page_type, build_status FROM pages WHERE site_id = (SELECT id FROM sites WHERE domain = '$DOMAIN') ORDER BY name;"
echo ""
echo "Check work items:"
echo "  SELECT item_type, status, handler_agent, summary FROM site_work_items WHERE site_id = (SELECT id FROM sites WHERE domain = '$DOMAIN') AND pipeline = 'build' ORDER BY priority;"


{
  "headers": {
    "correlation_id": "CORR_ID_PLACEHOLDER",
    "orchestration_id": "ORCH_ID_PLACEHOLDER",
    "message_type": "request",
    "action": "orchestrate",
    "sender": {"agent_id": "cli-user", "agent_type": "cli", "pod_name": "cli"}
  },
  "config": {
    "workflow": {
      "start_step": "spawn",
      "steps": {
        "spawn": {
          "action": "spawn_agent",
          "config": {"agent_type": "site-adoption-agent"},
          "next_step": "call"
        },
        "call": {
          "action": "call_agent",
          "config": {
            "role": "default",
            "input_mapping": {
              "domain": "input_data.domain",
              "url": "input_data.url"
            }
          },
          "next_step": "complete"
        },
        "complete": {"action": "complete_workflow"}
      }
    }
  },
  "input_data": {
    "domain": "mortgagecalculator.co.uk",
    "url": "https://mortgagecalculator.co.uk"
  }
}

-- wipe database
BEGIN;

-- Cancel active work items
UPDATE site_work_items SET status = 'wont_fix'
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND status NOT IN ('complete', 'wont_fix', 'failed');

SELECT archive_completed_work_items(0, 10000);

-- Supersede all specs
UPDATE site_specs SET is_current = false, superseded_at = NOW()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND is_current = true;

-- Clear research results
DELETE FROM research_results
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk');

-- Clear non-cascading FK references then pages
DELETE FROM page_component_history
WHERE page_id IN (SELECT id FROM pages WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk'));

DELETE FROM redirects
WHERE source_page_id IN (SELECT id FROM pages WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk'));

DELETE FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk');

-- Reset site
UPDATE sites SET build_status = 'pending', style_collection_id = NULL
WHERE domain = 'gamedesign.uk';

COMMIT;

-- then verify wiped
SELECT
    (SELECT count(*) FROM pages WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')) as pages,
    (SELECT count(*) FROM site_specs WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk') AND is_current = true) as specs;


clients_db=# -- Check what formatted_crawl contained
SELECT
    substring(collected_data::text,
        position('formatted_crawl' in collected_data::text),
        300)
FROM orchestration_states
WHERE correlation_id = '013bca99-e8f4-408d-aabf-eb5c6638bcc8';
                                                                                                                                                  substring
--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 formatted_crawl": {"sources": [{"url": "https://robot-hands.com/", "index": 1, "title": "ROBOT-HANDS.COM | Mission Control"}, {"url": "https://robot-hands.com/products/index.html?id=onr-rg6", "index": 2, "title": "Product Analysis | ROBOT-HANDS.COM"}, {"url": "https://robot-hands.com/products/index.
(1 row)

SELECT data::text
FROM research_results
WHERE id = '551897ad-6f1a-4181-9f4b-033dbac14edf';

-- Check if pages were created
SELECT name, page_type, build_status FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
ORDER BY name;

-- Check work items
SELECT item_type, status, handler_agent, summary
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
AND pipeline = 'build' ORDER BY priority;

-- Check research_results for the stored crawl analysis
SELECT id, result_type, substring(summary, 1, 80) as summary,
       substring(data::text, 1, 200) as data_preview
FROM research_results
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
ORDER BY created_at DESC LIMIT 5;

-- Check orchestration for the LLM output
SELECT substring(collected_data::text from 'adoption_analysis.{0,500}')
FROM orchestration_states
WHERE correlation_id = '013bca99-e8f4-408d-aabf-eb5c6638bcc8';

---

-- Watch pages getting built
SELECT name, build_status FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
ORDER BY name;

-- Watch work items getting processed
SELECT item_type, status, summary
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
ORDER BY priority;

------------------------------------

clients_db=# -- 1. Check specs — did content_direction get written?
SELECT aspect, source,
       length(data::text) as data_len,
       data->>'formatted' IS NOT NULL as has_formatted,
       created_at
FROM site_specs
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
  AND is_current = true
ORDER BY aspect;

-- 2. Check pages created
SELECT name, page_type, build_status, created_at
FROM pages
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
ORDER BY name;

-- 3. Check work items created
SELECT item_type, status, handler_agent, priority, summary
FROM site_work_items
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
  AND pipeline = 'build'
ORDER BY priority;

-- 4. Check orchestration state
SELECT status, current_step, error, created_at, updated_at
FROM orchestration_states
WHERE agent_type = 'site-adoption-agent'
  AND created_at > '2026-04-02 13:00:00'
ORDER BY created_at DESC LIMIT 3;

-- Check the index page build — what's happening with the claimed item?
SELECT wi.item_type, wi.status, wi.handler_agent, wi.summary,
       wi.claimed_at, wi.updated_at
FROM site_work_items wi
WHERE wi.site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
  AND wi.status = 'claimed'
ORDER BY wi.updated_at DESC;

-- Check if any page-build-handler orchestrations are running
SELECT owner_agent_type, status, current_step, error, created_at, updated_at
FROM orchestration_states
WHERE owner_agent_type = 'page-build-handler'
  AND created_at > '2026-04-02 13:00:00'
ORDER BY created_at DESC LIMIT 5;


---------------------------------------------
tools
-- 1. Did it complete?
SELECT summary, status, error, completed_at
FROM site_work_items
WHERE item_key = 'add_tool_novel:tool-gas-unit-converter:5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND status != 'wont_fix';

-- 2. Component?
SELECT cc.id, cc.function, cc.display_name, cc.created_from,
       length(cc.html_template) as html_length
FROM content_components cc
WHERE cc.function = 'tool-gas-unit-converter'
  AND cc.is_active = true;

-- 3. Page component at position 2 with rendered_html?
SELECT pc.position, pc.slot_name, pc.build_status,
       length(pc.rendered_html) as rendered_html_length
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'tool-gas-unit-converter';

-- 4. Follow-up work items?
SELECT item_type, summary, handler_agent, status
FROM site_work_items
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND source = 'tool-generator'
ORDER BY created_at DESC;

-- 5. Companion guide?
SELECT p.name, p.url, p.page_type, p.build_status
FROM pages p
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name LIKE '%gas-unit%'
ORDER BY p.created_at DESC;

-- 6. Nav entry?
SELECT sng.group_key, sng.group_label, sni.label, sni.url
FROM site_nav_items sni
JOIN site_nav_groups sng ON sni.group_id = sng.id
WHERE sng.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND sng.group_key = 'tools';

------------------------
-- Full picture: adoption items + component items
SELECT item_type, status, spec->>'section_type' as section_type,
       spec->>'page_name' as page, created_at
FROM site_work_items
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
  AND created_at > now() - interval '30 minutes'
  AND item_type IN ('needs_new_component', 'needs_content_page', 'needs_design')
ORDER BY created_at;

-- Components landing in the library
SELECT function, section_type, display_name,
       length(html_template) as len, created_at
FROM content_components
WHERE created_at > now() - interval '30 minutes'
ORDER BY created_at;

-- Pages building
SELECT name, build_status, updated_at
FROM pages
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
ORDER BY updated_at DESC;



--- add work items for new tools
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
)
SELECT
    s.id,
    'admin', 'build', 'evaluate_tools', 'low',
    'Evaluate tool needs for ' || s.domain,
    '{"reason": "initial_tool_evaluation"}'::jsonb,
    130, 'tool-suggester', 'triaged', 'admin',
    'evaluate_tools:' || s.id::text
FROM sites s
WHERE s.status = 'deployed'
  AND NOT EXISTS (
      SELECT 1 FROM site_work_items wi
      WHERE wi.site_id = s.id
        AND wi.item_type = 'evaluate_tools'
        AND wi.status NOT IN ('complete', 'wont_fix', 'failed')
  )
ORDER BY s.domain;