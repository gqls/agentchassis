
first:
INSERT INTO build_queue (domain, priority) VALUES ('example.com', 10);

gaswholesalers.com
INSERT INTO build_queue (domain, priority) VALUES ('gaswholesalers.com', 10);

SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

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

echo "  -- 1. Current work item status for both sites "
echo "  SELECT s.domain, wi.item_type, wi.status, wi.handler_agent, "
echo "         LEFT(wi.error, 60) as error "
echo "  FROM site_work_items wi "
echo "  JOIN sites s ON s.id = wi.site_id "
echo "  WHERE s.domain IN ('finetuning.uk', 'gaswholesalers.com') "
echo "    AND wi.status NOT IN ('complete', 'wont_fix') "
echo "  ORDER BY s.domain, wi.priority; "
echo "   "
echo "  -- 2. Any blocked items? "
echo "  SELECT s.domain, wi.item_type, wi.handler_agent, wi.error "
echo "  FROM site_work_items wi "
echo "  JOIN sites s ON s.id = wi.site_id "
echo "  WHERE wi.status = 'blocked'; "
echo "   "
echo "  -- 3. Scheduled tasks running? "
echo "  SELECT name, enabled, last_triggered_at, "
echo "         EXTRACT(EPOCH FROM (NOW() - last_triggered_at))::int as seconds_ago "
echo "  FROM scheduled_tasks "
echo "  WHERE name IN ('claimed-item-timeout', 'feasibility-recheck', 'build-pipeline-trigger'); "
echo "   "
echo "  -- 4. Recent orchestrations (last 30 min) "
echo "  SELECT owner_agent_type, status, current_step, "
echo "         EXTRACT(EPOCH FROM (NOW() - last_activity))::int as idle_seconds "
echo "  FROM orchestration_states "
echo "  WHERE created_at > NOW() - INTERVAL '30 minutes' "
echo "  ORDER BY created_at DESC LIMIT 10; "
echo "   "
echo "  -- 5. Running pods "
echo "  -- kubectl -n ai-persona-system get pods | grep -v Completed "
echo "   "
echo "   "
echo "   SITE_ID='1368e337-dd1d-4799-bbb3-8221a1b79bcc'  "
echo "   DOMAIN='finetuning.uk'  "
echo "    "
echo "   SITE_ID='5fe15466-4e2e-4ff2-981e-98c1b7074002'  "
echo "   DOMAIN='gaswholesalers.com'  "
echo "   "
echo "  -- 6. Blog page progress (finetuning.uk) "
echo "  SELECT pc.slot_name, pc.build_status, LENGTH(pc.rendered_html) as html_len "
echo "  FROM page_components pc "
echo "  JOIN pages p ON pc.page_id = p.id "
echo "  WHERE p.site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc' "
echo "    AND p.name = 'blog' "
echo "  ORDER BY pc.position; "
echo "   "
echo "  -- 7. Audit findings created (if audit has run) "
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80) "
echo "  FROM site_work_items "
echo "  WHERE source = 'discovery' "
echo "    AND created_at > NOW() - INTERVAL '1 hour' "
echo "  ORDER BY created_at DESC LIMIT 20; "
echo " "
echo ""


kubectl -n ai-persona-system logs --tail=300 -l agent-type=asset-deployer -f | tee logs-asset-deployer.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=build-dispatch-loop -f | tee logs-build-dispatch-loop.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=domain-research-classifier -f | tee logs-domain-research-classifier.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=build-briefing-agent -f | tee logs-build-briefing-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=build-site-planner -f | tee logs-build-site-planner.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=page-content-writer -f | tee logs-page-content-writer.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=research-agent -f | tee logs-research-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=improvement-loop -f | tee logs-improvement-loop.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=quality-discovery-agent -f | tee logs-quality-discovery-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=design-discovery-agent -f | tee logs-design-discovery-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=design-audit-agent -f | tee logs-design-audit-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=completeness-discovery-agent -f | tee logs-completeness-discovery-agent.json
kubectl -n ai-persona-system logs --tail=500 -l app=git-adapter -f | tee logs-git-adapter.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=color-variable-fixer -f | tee logs-color-variable-fixer.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=visual-design-auditor -f | tee logs-visual-design-auditor.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=content-quality-auditor -f | tee logs-content-quality-auditor.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=content-gap-planner -f | tee logs-content-gap-planner.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=component-template-fixer -f | tee logs-component-template-fixer.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=page-build-handler -f | tee logs-page-build-handler.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=site-review-agent -f | tee logs-site-review-agent.json

kubectl -n ai-persona-system logs --tail=300 -l agent-type=webdesign-agent -f | tee logs-webdesign-agent.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=build-dispatch-loop -f | tee logs-build-dispatch-loop.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=section-editor -f | tee logs-section-editor.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=deployer-agent -f | tee logs-deployer-agent.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=page-rerender -f | tee logs-page-rerender.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=rerender-pages -f | tee logs-rerender-pages.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=image-generator -f | tee logs-image-generator.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=image-build-handler -f | tee logs-image-build-handler.json
kubectl -n ai-persona-system logs --tail=300 -l app=kafka-scheduler -f | tee logs-kafka-scheduler.json



SELECT wi.item_type, wi.status, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.domain = 'build' AND wi.status != 'complete' ORDER BY wi.created_at DESC;

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


UPDATE site_work_items SET status = 'triaged', claimed_at = NULL, claimed_by = NULL
WHERE status = 'claimed';

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


-- Quick status check - run periodically
SELECT wi.item_type, wi.status, s.domain, wi.completed_at
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status IN ('claimed', 'triaged') AND wi.domain = 'build'
ORDER BY s.domain, wi.status, wi.priority;

SELECT wi.item_type, wi.status, s.domain, wi.completed_at, wi.domain
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status IN ('claimed', 'triaged')
ORDER BY s.domain, wi.status, wi.priority;

---
check for everything pending:
-- 1. Finetuning email - should be finetuning@contactforsales.com
SELECT email FROM sites WHERE id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc';

-- 2. Gaswholesalers style collection - should be professional-dark
SELECT style_collection_id FROM sites WHERE id = '5fe15466-4e2e-4ff2-981e-98c1b7074002';

-- 3. Both sites hero components - should reference hero.jpg
SELECT s.domain, COUNT(*) as hero_with_image
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
JOIN sites s ON s.id = p.site_id
WHERE p.site_id IN ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', '5fe15466-4e2e-4ff2-981e-98c1b7074002')
  AND pc.rendered_html LIKE '%/assets/images/hero.jpg%'
GROUP BY s.domain;

-- 4. Site components linked?
SELECT s.domain, sc.slot_name, sc.component_id IS NOT NULL as linked, cc.name
FROM site_components sc
LEFT JOIN content_components cc ON sc.component_id = cc.id
JOIN sites s ON s.id = sc.site_id
WHERE sc.site_id IN ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', '5fe15466-4e2e-4ff2-981e-98c1b7074002')
ORDER BY s.domain, sc.slot_name;

-- 5. Blog work item status
SELECT item_type, status, error FROM site_work_items
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND item_type = 'needs_content_page'
ORDER BY created_at DESC LIMIT 1;

-- 6. Agent definitions created?
SELECT type, status FROM agent_definitions
WHERE type IN ('site-component-linker', 'component-template-fixer', 'page-build-handler')
  AND deleted_at IS NULL;

-- 7. Gaswholesalers email/phone
SELECT email, phone FROM sites WHERE id = '5fe15466-4e2e-4ff2-981e-98c1b7074002';

-- 8. Any stuck claimed items?
SELECT s.domain, wi.item_type, wi.status
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status = 'claimed'
  AND wi.claimed_at < NOW() - INTERVAL '10 minutes';

-- 9. Pending triaged items for both sites
SELECT s.domain, wi.item_type, wi.status, wi.handler_agent
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status = 'triaged'
  AND s.domain IN ('finetuning.uk', 'gaswholesalers.com')
ORDER BY s.domain, wi.priority;


Two stuck claimed items — needs_rerender for gaswholesalers and add_tool for finetuning. These are from a previous dispatch run that's either still running or timed out. Release them:
UPDATE site_work_items SET status = 'triaged', claimed_at = NULL, claimed_by = NULL
WHERE status = 'claimed' AND claimed_at < NOW() - INTERVAL '10 minutes';

--

-- How long have claimed items been stuck?
SELECT s.domain, wi.item_type, wi.handler_agent,
       EXTRACT(EPOCH FROM (NOW() - wi.claimed_at))::int as claimed_secs_ago
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status = 'claimed'
ORDER BY wi.claimed_at;

-- Is the dispatch loop pod still running?
-- kubectl -n ai-persona-system get pods | grep dispatch

-- Check recent orchestrations
SELECT owner_agent_type, status, current_step,
       EXTRACT(EPOCH FROM (NOW() - last_activity))::int as idle_secs
FROM orchestration_states
WHERE created_at > NOW() - INTERVAL '30 minutes'
ORDER BY created_at DESC LIMIT 10;

-----------
DEBUG
-----------

 cta_improvement          | failed  | finetuning.uk

Check the work item's error and result fields:
SELECT wi.id, wi.item_type, wi.handler_agent, wi.error, wi.result,
       wi.attempt_count, wi.claimed_at, wi.completed_at
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE s.domain = 'finetuning.uk'
  AND wi.item_type = 'cta_improvement'
  AND wi.status = 'failed';

Then check the orchestration that processed it:
  SELECT os.owner_agent_type, os.status, os.current_step,
         LEFT(os.error, 300) as error,
         os.created_at, os.last_activity
  FROM orchestration_states os
  WHERE os.owner_agent_type = 'component-template-fixer'
    AND os.status = 'FAILED'
  ORDER BY os.created_at DESC LIMIT 3;


  -------------------------------------------------------------------------

  what failed and why

-- Failed items with error details
SELECT s.domain, wi.item_type, wi.handler_agent,
       LEFT(wi.error, 200) as error,
       wi.claimed_at, wi.completed_at
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status = 'failed'
ORDER BY s.domain, wi.item_type;

-- Count by type and site for the full picture
SELECT s.domain, wi.item_type, wi.status, COUNT(*) as cnt
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.domain = 'build' AND wi.status != 'complete'
GROUP BY s.domain, wi.item_type, wi.status
ORDER BY s.domain, wi.status, cnt DESC;

-------------------------------------------------------------------------

The fix is to stop the audit cycle until the queue is drained:

-- Disable improvement-sweep temporarily
UPDATE scheduled_tasks SET enabled = false WHERE name = 'improvement-sweep';

-- Check how many items are being processed vs queued
SELECT s.domain,
    COUNT(*) FILTER (WHERE wi.status = 'complete') as done,
    COUNT(*) FILTER (WHERE wi.status IN ('triaged', 'claimed')) as pending,
    COUNT(*) FILTER (WHERE wi.status = 'failed') as failed
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.domain = 'build'
GROUP BY s.domain
ORDER BY pending DESC;