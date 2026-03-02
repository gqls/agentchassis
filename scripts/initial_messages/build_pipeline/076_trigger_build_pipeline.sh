
first:
INSERT INTO build_queue (domain, priority) VALUES ('example.com', 10);

gaswholesalers.com
INSERT INTO build_queue (domain, priority) VALUES ('gaswholesalers.com', 10);

#!/bin/bash
# =============================================================================
# Build Pipeline Trigger (manual heartbeat)
# =============================================================================
# Sends an orchestrate message to the build-pipeline-trigger agent.
# This is the manual equivalent of the CronJob heartbeat that would
# normally fire every 30 minutes.
#
# What it does:
#   1. seed_build_queue — processes build_queue entries → creates sites + work items
#   2. find_dispatchable_site — queries for sites with pending build work items
#   3. If found: spawns + calls build-dispatch-loop for that site
#   4. The dispatch loop processes items one at a time, chaining to itself
#
# Prerequisites:
#   - build-pipeline-trigger agent definition in agent_definitions table
#   - build-dispatch-loop agent definition in agent_definitions table
#   - Handler agents registered (domain-research-classifier, build-briefing-agent, etc.)
#   - Entries in build_queue table, e.g.:
#       INSERT INTO build_queue (domain, priority) VALUES ('example.com', 10);
#
# Usage:
#   ./054_trigger_build_pipeline.sh
# =============================================================================

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Build Pipeline Trigger"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Time:             $TIMESTAMP"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

kubectl -n kafka run -i --rm kcat-build-trigger-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=build-pipeline-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"build-pipeline-trigger"},"input_data":{}}
JSON

echo ""
echo "========================================="
echo "Build pipeline triggered"
echo "========================================="
echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100 | grep '$CORRELATION_ID'"
echo ""
echo "Check seed results:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 | grep '$CORRELATION_ID' | grep -E 'seed_queue|seed_build_queue|find_dispatchable'"
echo ""
echo "Check dispatch:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 | grep '$CORRELATION_ID' | grep -E 'spawn_dispatch|call_dispatch|dispatch_result'"
echo ""
echo "Check orchestration state:"
echo "  SELECT status, current_step, error FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid;"
echo ""
echo "Check build_queue:"
echo "  SELECT domain, status, priority, created_at FROM build_queue ORDER BY created_at DESC LIMIT 5;"
echo ""
echo "Check work items:"
echo "  SELECT wi.item_type, wi.status, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.domain = 'build' ORDER BY wi.created_at DESC LIMIT 30;"
echo ""



kubectl -n ai-persona-system logs --tail=500 -l agent-type=build-dispatch-loop -f | tee logs-build-dispatch-loop.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=domain-research-classifier -f | tee logs-domain-research-classifier.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=build-briefing-agent -f | tee logs-build-briefing-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=build-site-planner -f | tee logs-build-site-planner.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=image-generator -f | tee logs-image-generator.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=page-content-writer -f | tee logs-page-content-writer.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=improvement-loop -f | tee logs-improvement-loop.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=quality-discovery-agent -f | tee logs-quality-discovery-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=design-discovery-agent -f | tee logs-design-discovery-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=webdesign-agent -f | tee logs-webdesign-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=completeness-discovery-agent -f | tee logs-completeness-discovery-agent.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=build-dispatch-loop -f | tee logs-build-dispatch-loop.json




reset
-- Reset the claimed (failed) content page back to triaged
UPDATE site_work_items
SET status = 'triaged',
    claimed_by = NULL,
    claimed_at = NULL,
    completed_at = NULL,
    result = '{}'::jsonb,
    error = NULL,
    attempt_count = 0
WHERE status = 'claimed'
  AND site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND domain = 'build';

--- another reset for comparison (same?)
-- Reset needs_design to triaged so webdesign-agent runs again with the toJSON fix
UPDATE site_work_items
SET status = 'triaged',
    completed_at = NULL,
    result = NULL,
    error = NULL,
    claimed_by = NULL,
    claimed_at = NULL
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND item_type = 'needs_design'
  AND status = 'complete';




  then
    Step 1: Webdesign (CSS generation)
    Step 2: Rerender (assemble all pages)


    clients_db=# -- 1. Check page components have rendered HTML
    SELECT pc.id, p.name as page_name, pc.position,
           LENGTH(pc.rendered_html) as html_len,
           LEFT(pc.rendered_html, 120) as preview,
           pc.build_status
    FROM page_components pc
    JOIN pages p ON pc.page_id = p.id
    WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
    ORDER BY p.name, pc.position
    LIMIT 40;

    -- 2. Check pages themselves
    SELECT name, url, title, build_status,
           LENGTH(rendered_header) as header_len,
           LENGTH(rendered_footer) as footer_len,
           LEFT(rendered_head, 150) as head_preview
    FROM pages
    WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
    ORDER BY nav_order;

    -- 3. Check css_themes table
    SELECT id, site_id, LEFT(css_content, 200) as preview, LENGTH(css_content) as len
    FROM css_themes
    WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002';

    -- 4. All site_specs aspects (not just CSS)
    SELECT aspect, LENGTH(data::text) as len, created_at
    FROM site_specs
    WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
    AND is_current = true
    ORDER BY created_at;
     id | page_name | position | html_len | preview | build_status
    ----+-----------+----------+----------+---------+--------------
    (0 rows)

                name             |                url                |                               title                                | build_status | header_len | footer_len | head_preview
    -----------------------------+-----------------------------------+--------------------------------------------------------------------+--------------+------------+------------+--------------
     wholesale-fuel-distribution | /wholesale-fuel-distribution.html | Wholesale Fuel Distribution | Gas Wholesalers                      | deployed     |            |            |
     fleet-fuel-services         | /fleet-fuel-services.html         | Fleet Fuel Services | Gas Wholesalers                              | deployed     |            |            |
     natural-gas-distribution    | /natural-gas-distribution.html    | Natural Gas Distribution | Gas Wholesalers                         | deployed     |            |            |
     rack-pricing-programs       | /rack-pricing-programs.html       | Rack Pricing Programs | Gas Wholesalers                            | deployed     |            |            |
     index                       | /index.html                       | Gas Wholesalers | Wholesale Fuel Distribution & Natural Gas Supply | deployed     |            |            |
     about                       | /about.html                       | About Us | Gas Wholesalers                                         | deployed     |            |            |
     services                    | /services.html                    | Our Services | Gas Wholesalers                                     | deployed     |            |            |
     contact                     | /contact.html                     | Contact Us | Gas Wholesalers                                       | deployed     |            |            |
    (8 rows)

    ERROR:  column "site_id" does not exist
    LINE 1: SELECT id, site_id, LEFT(css_content, 200) as preview, LENGT...
                       ^
         aspect     | len  |          created_at
    ----------------+------+-------------------------------
     identity       | 1634 | 2026-02-25 10:08:34.477358+00
     classification |  778 | 2026-02-25 10:08:34.583426+00
     briefing       | 1460 | 2026-02-25 10:09:45.326584+00
     site_plan      | 3744 | 2026-02-25 11:02:52.248378+00
    (4 rows)


SELECT aspect, LEFT(data::text, 300) as preview
FROM site_specs
WHERE site_id='5fe15466-4e2e-4ff2-981e-98c1b7074002';


-- you need to reset the status so the dispatch loop picks it up again. Something like:

UPDATE site_work_items
SET status = 'triaged', claimed_by = NULL, claimed_at = NULL
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gaswholesalers.com')
  AND item_type IN ('needs_design', 'generic_theme')
  AND status IN ('complete', 'claimed', 'failed');


---

just one site:
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Build Pipeline Trigger for gaswholesalers.com"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Time:             $TIMESTAMP"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

# Direct dispatch for gaswholesalers, bypassing site selection
kubectl -n kafka run -i --rm kcat-build-trigger-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=build-pipeline-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"build-dispatch-loop"},"input_data":{"site_id":"5fe15466-4e2e-4ff2-981e-98c1b7074002","domain":"gaswholesalers.com"}}
JSON