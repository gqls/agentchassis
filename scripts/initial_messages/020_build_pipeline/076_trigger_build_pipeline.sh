
first:
INSERT INTO build_queue (domain, priority) VALUES ('example.com', 10);

gaswholesalers.com
INSERT INTO build_queue (domain, priority) VALUES ('gaswholesalers.com', 10);

clients_db=# SELECT id, domain FROM sites;
                  id                  |           domain
--------------------------------------+----------------------------
 1368e337-dd1d-4799-bbb3-8221a1b79bcc | finetuning.uk
 859d7ad5-0f22-4ba1-8efd-cd59e8fb042f | gamesdesign.co.uk
 eac60db8-b032-432b-b36d-76f37632045d | system.internal
 00ff3af5-dad8-4770-9f70-3edc267a3c92 | robot-hands.com
 5fe15466-4e2e-4ff2-981e-98c1b7074002 | gaswholesalers.com
 4851f6fc-71cf-4160-a270-e03d6d3e0732 | leopardessconsulting.co.uk
 e1e22a7d-0552-405a-85b3-1b1e51384df5 | vonc.com
 2a8ebf9c-20a2-4c39-b191-840b012371da | ai-agent-orchestration.com
(8 rows)

SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

 SELECT id, name FROM sites;
                  id                  |            name
--------------------------------------+----------------------------
 4851f6fc-71cf-4160-a270-e03d6d3e0732 | leopardessconsulting.co.uk
 5fe15466-4e2e-4ff2-981e-98c1b7074002 | gaswholesalers.com
 2a8ebf9c-20a2-4c39-b191-840b012371da | ai-agent-orchestration.com
 1368e337-dd1d-4799-bbb3-8221a1b79bcc | FineTuning


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


UPDATE scheduled_tasks SET enabled = true WHERE name = 'build-pipeline-trigger';
kubectl -n ai-persona-system logs deploy/kafka-scheduler --tail=20

kubectl -n ai-persona-system port-forward svc/admin-dashboard 8080:8080

kubectl -n ai-persona-system logs --tail=300 -l app=agent-chassis -f | tee logs-agent-chassis.json
kubectl -n ai-persona-system logs --tail=300 -l app=core-manager -f | tee logs-core-manager.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=asset-deployer -f | tee logs-asset-deployer.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=build-dispatch-loop -f | tee logs-build-dispatch-loop.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=domain-research-classifier -f | tee logs-domain-research-classifier.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=domain-strategist -f | tee logs-domain-strategist.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=build-briefing-agent -f | tee logs-build-briefing-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=build-site-planner -f | tee logs-build-site-planner.json

kubectl -n ai-persona-system logs --tail=500 -l agent-type=research-agent -f | tee logs-research-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=improvement-loop -f | tee logs-improvement-loop.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=quality-discovery-agent -f | tee logs-quality-discovery-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=design-discovery-agent -f | tee logs-design-discovery-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=design-audit-agent -f | tee logs-design-audit-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=completeness-discovery-agent -f | tee logs-completeness-discovery-agent.json
kubectl -n ai-persona-system logs --tail=500 -l app=git-adapter -f | tee logs-git-adapter.json
kubectl -n ai-persona-system logs --tail=500 -l app=web-scrape-adapter -f | tee logs-web-scrape-adapter.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=color-variable-fixer -f | tee logs-color-variable-fixer.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=visual-design-auditor -f | tee logs-visual-design-auditor.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=content-quality-auditor -f | tee logs-content-quality-auditor.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=content-creator-agent -f | tee logs-content-creator-agent.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=content-gap-planner -f | tee logs-content-gap-planner.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=content-feed-orchestrator -f | tee logs-content-feed-orchestrator.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=component-template-fixer -f | tee logs-component-template-fixer.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=component-creator -f | tee logs-component-creator.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=component-quality-auditor -f | tee logs-component-quality-auditor.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=page-build-handler -f | tee logs-page-build-handler.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=page-content-writer -f | tee logs-page-content-writer.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=page-rerender -f | tee logs-page-rerender.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=rerender-pages -f | tee logs-rerender-pages.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=site-review-agent -f | tee logs-site-review-agent.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=training-data-exporter -f | tee logs-training-data-exporter.json

kubectl -n ai-persona-system logs --tail=300 -l agent-type=webdesign-agent -f | tee logs-webdesign-agent.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=build-dispatch-loop -f | tee logs-build-dispatch-loop.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=section-editor -f | tee logs-section-editor.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=deployer-agent -f | tee logs-deployer-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=site-adoption-agent -f | tee logs-site-adoption-agent.json

kubectl -n ai-persona-system logs --tail=500 -l app=image-generator-adapter -f | tee logs-image-generator-adapter.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=image-generator -f | tee logs-image-generator.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=image-build-handler -f | tee logs-image-build-handler.json
kubectl -n ai-persona-system logs --tail=300 -l app=kafka-scheduler -f | tee logs-kafka-scheduler.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=med-price-collector -f --max-log-requests 20 | tee logs-med-price-collector.json
kubectl -n ai-persona-system logs --tail=500 -l app=business-intel -f --max-log-requests 20 | tee logs-business-intel.json
kubectl -n ai-persona-system logs --tail=500 -l agent-type=training-data-preparer -f --max-log-requests 20 | tee logs-training-data-preparer.json
kubectl -n ai-persona-system logs --tail=500 -l app=thunder-adapter -f --max-log-requests 20 | tee logs-thunder-adapter.json

kubectl -n ai-persona-system logs -l 'app in (agent-chassis,dynamic-agent)'   -f --max-log-requests=20  | tee logs-all_chassis_logs_$(date +%H%M%S).log


SELECT wi.item_type, wi.status, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.pipeline = 'build' AND wi.status != 'complete' ORDER BY wi.created_at DESC;
SELECT wi.item_type, wi.status, s.domain, LEFT(wi.summary, 50) FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.pipeline = 'build' AND wi.status != 'complete' ORDER BY wi.created_at DESC;
SELECT wi.item_type, wi.status, s.domain, LEFT(wi.summary, 50), LEFT(wi.error, 150) FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.pipeline = 'build' AND wi.status != 'complete' ORDER BY wi.created_at DESC;

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


------------------------------------------------
see errors

SELECT id, status, attempt_count, handler_agent,
       created_at, claimed_at, completed_at, updated_at,
       result,
       LEFT(error, 300) AS error,
       spec
  FROM site_work_items
 WHERE id = 'dbdbe82a-ae2e-4609-91ac-4a4d5b8825f7';
----

-- monitoring query

--------------------

-- Dashboard: site summary + failure details
SELECT s.domain,
       COUNT(*) FILTER (WHERE wi.status = 'complete') as done,
       COUNT(*) FILTER (WHERE wi.status = 'claimed') as active,
       COUNT(*) FILTER (WHERE wi.status = 'triaged' AND wi.attempt_count < wi.max_attempts) as ready,
       COUNT(*) FILTER (WHERE wi.status = 'triaged' AND wi.attempt_count >= wi.max_attempts) as exhausted,
       COUNT(*) FILTER (WHERE wi.status = 'failed') as failed,
       COUNT(*) FILTER (WHERE wi.status = 'blocked') as blocked
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.domain = 'build'
GROUP BY s.domain
ORDER BY s.domain;

-- Failure + exhausted detail
SELECT s.domain, wi.status, wi.item_type, wi.handler_agent,
       wi.attempt_count || '/' || wi.max_attempts as attempts,
       LEFT(wi.spec->>'page_name', 20) as page,
       LEFT(COALESCE(wi.error, wi.spec->>'description'), 80) as detail
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.pipeline = 'build'
  AND (wi.status = 'failed'
       OR (wi.status = 'triaged' AND wi.attempt_count >= wi.max_attempts))
ORDER BY s.domain, wi.status, wi.item_type;


-- Categorise failures
-- Group the failures by error to see patterns
SELECT wi.item_type, COUNT(*) as cnt,
       LEFT(wi.error, 120) as error_pattern
FROM site_work_items wi
WHERE wi.status = 'failed' AND wi.domain = 'build'
GROUP BY wi.item_type, LEFT(wi.error, 120)
ORDER BY cnt DESC;


--------------------


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

-- what they contain
SELECT s.domain, wi.item_type, wi.handler_agent,
       wi.spec->>'page_name' as page,
       LEFT(wi.spec->>'description', 100) as description,
       wi.id
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status = 'failed'
ORDER BY s.domain;

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

------------------------------------------------------------------------

resetting attempt counts

-- Reset the 3 stale claimed items
UPDATE site_work_items
SET status = 'triaged', claimed_by = NULL, claimed_at = NULL,
    attempt_count = attempt_count + 1
WHERE status = 'claimed'
  AND domain = 'build'
RETURNING site_id, item_type, handler_agent;

-- Also reset the 19 exhausted-attempt items (bugs are now fixed)
UPDATE site_work_items
SET attempt_count = 0
WHERE status = 'triaged'
  AND domain = 'build'
  AND attempt_count >= max_attempts;

-- Verify claimed-item-timeout is running
SELECT name, enabled, fire_message, last_triggered_at
FROM scheduled_tasks
WHERE name = 'claimed-item-timeout';


============================================================================

# How's the pipeline doing

SELECT s.domain,
       COUNT(*) FILTER (WHERE wi.status = 'complete') as done,
       COUNT(*) FILTER (WHERE wi.status = 'claimed') as active,
       COUNT(*) FILTER (WHERE wi.status = 'triaged' AND wi.attempt_count < wi.max_attempts) as ready,
       COUNT(*) FILTER (WHERE wi.status = 'triaged' AND wi.attempt_count >= wi.max_attempts) as exhausted,
       COUNT(*) FILTER (WHERE wi.status = 'failed') as failed,
       COUNT(*) FILTER (WHERE wi.status = 'blocked') as blocked,
       COUNT(*) FILTER (WHERE wi.status = 'needs_human_review') as human_review
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.domain = 'build'
GROUP BY s.domain
ORDER BY s.domain;
           domain           | done | active | ready | exhausted | failed | blocked | human_review
----------------------------+------+--------+-------+-----------+--------+---------+--------------
 ai-agent-orchestration.com |   32 |      1 |     5 |         0 |      0 |       1 |            0
 finetuning.uk              |   71 |      1 |     5 |         0 |      0 |       1 |            0
 gaswholesalers.com         |   57 |      1 |    23 |         0 |      0 |       0 |            0
 leopardessconsulting.co.uk |   30 |      1 |    17 |         1 |      2 |       1 |            0
(4 rows)

clients_db=# -- Check the 2 failures and 1 exhausted
SELECT wi.item_type, wi.handler_agent, wi.status,
       wi.attempt_count || '/' || wi.max_attempts as attempts,
       wi.spec->>'page_name' as page,
       LEFT(COALESCE(wi.error, wi.spec->>'description'), 100) as detail
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE s.domain = 'leopardessconsulting.co.uk'
  AND wi.domain = 'build'
  AND (wi.status = 'failed'
       OR (wi.status = 'triaged' AND wi.attempt_count >= wi.max_attempts))
ORDER BY wi.status;
      item_type      |  handler_agent  | status  | attempts |    page    |     detail
---------------------+-----------------+---------+----------+------------+----------------
 needs_design_review | webdesign-agent | failed  | 3/3      | index.html | Handler failed
 needs_design_review | webdesign-agent | failed  | 3/3      | global     | Handler failed
 needs_design_review | webdesign-agent | triaged | 3/3      | site-wide  | Handler failed
(3 rows)

--

checking for orphans
SELECT name, enabled, fire_message, last_triggered_at
FROM scheduled_tasks
WHERE name = 'claimed-item-timeout';
         name         | enabled | fire_message |       last_triggered_at
----------------------+---------+--------------+-------------------------------
 claimed-item-timeout | t       | f            | 2026-03-09 18:11:30.731595+00
(1 row)

clients_db=# -- Has last_completed_at been updated recently?
SELECT name, last_triggered_at, last_completed_at,
       NOW() - last_triggered_at as since_trigger,
       NOW() - last_completed_at as since_complete
FROM scheduled_tasks
WHERE name IN ('build-pipeline-trigger', 'claimed-item-timeout');
          name          |       last_triggered_at       |       last_completed_at       |     since_trigger      | since_complete
------------------------+-------------------------------+-------------------------------+------------------------+-----------------
 build-pipeline-trigger | 2026-03-12 13:32:22.44988+00  | 2026-03-12 13:19:53.941417+00 | 00:04:21.581182        | 00:16:50.089645
 claimed-item-timeout   | 2026-03-09 18:11:30.731595+00 |                               | 2 days 19:25:13.299467 |
(2 rows)


=========================================================================

reset stale

=========================================================================

-- Check stale claims
SELECT s.domain, wi.item_type, wi.handler_agent,
       wi.claimed_at, NOW() - wi.claimed_at as stale_for
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status = 'claimed' AND wi.domain = 'build';

-- Reset them
UPDATE site_work_items
SET status = 'triaged', claimed_by = NULL, claimed_at = NULL
WHERE status = 'claimed' AND domain = 'build';

-- Kick the scheduler
UPDATE scheduled_tasks
SET last_completed_at = NOW()
WHERE name = 'build-pipeline-trigger';


-----------------------------------------------------------------------------

unblock routine
clients_db=# -- 1. Check what's blocking the maintenance group
SELECT name, concurrency_group, last_triggered_at, last_completed_at,
       NOW() - last_triggered_at as since_trigger,
       NOW() - COALESCE(last_completed_at, last_triggered_at) as since_complete
FROM scheduled_tasks
WHERE concurrency_group = 'maintenance';

-- 2. Reset stale claims immediately
UPDATE site_work_items
SET status = 'triaged', claimed_by = NULL, claimed_at = NULL
WHERE status = 'claimed' AND domain = 'build';

#-- 3. Fix the concurrency group — claimed-item-timeout should NOT share
#-- a group with database-cleanup. They're independent operations.
#UPDATE scheduled_tasks
#SET concurrency_group = 'claim-management'
#WHERE name = 'claimed-item-timeout';

-- 4. Unstick database-cleanup if it never completed
UPDATE scheduled_tasks
SET last_completed_at = NOW()
WHERE name = 'database-cleanup'
  AND last_completed_at IS NULL;

-- 5. Kick the pipeline
UPDATE scheduled_tasks
SET last_completed_at = NOW()
WHERE name = 'build-pipeline-trigger';

-- 6. Verify
SELECT name, concurrency_group, enabled,
       last_triggered_at, last_completed_at
FROM scheduled_tasks
WHERE name IN ('claimed-item-timeout', 'database-cleanup', 'build-pipeline-trigger');

 -- Check: is database-cleanup re-triggering before the timeout expires?
SELECT name, interval_seconds, timeout_seconds,
       last_triggered_at, last_completed_at
FROM scheduled_tasks
WHERE name = 'database-cleanup';


========================================================================================

deleting stale non duplicated items
----------------------------------------------------------------------------------------

BEGIN;

-- Step 1: Null out parent_item_id references to failed items we're about to delete
UPDATE site_work_items
SET parent_item_id = NULL
WHERE parent_item_id IN (
    SELECT id FROM site_work_items
    WHERE status = 'failed' AND pipeline = 'build'
);

-- Step 2: Delete failed items where a live (non-terminal) copy already exists
DELETE FROM site_work_items
WHERE status = 'failed' AND pipeline = 'build' AND item_key IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM site_work_items live
    WHERE live.site_id = site_work_items.site_id
      AND live.item_key = site_work_items.item_key
      AND live.id != site_work_items.id
      AND live.status NOT IN ('complete', 'verified', 'rejected', 'wont_fix', 'failed')
  );

-- Step 3: Among remaining failed items, keep only the newest per (site_id, item_key)
DELETE FROM site_work_items
WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (
            PARTITION BY site_id, item_key
            ORDER BY created_at DESC
        ) as rn
        FROM site_work_items
        WHERE status = 'failed' AND pipeline = 'build' AND item_key IS NOT NULL
    ) ranked
    WHERE rn > 1
);

-- Step 4: Now safe — each key has at most one failed row and no live duplicate
UPDATE site_work_items
SET status = 'triaged', attempt_count = 0, error = NULL,
    claimed_by = NULL, claimed_at = NULL
WHERE status = 'failed' AND pipeline = 'build';

COMMIT;


----------------------------------------------------------------------------------------


-- Check when content was last written vs rerendered
SELECT p.name, p.build_status, p.updated_at as page_updated,
       COUNT(pc.id) as component_count,
       SUM(CASE WHEN LENGTH(COALESCE(pc.rendered_html, '')) > 100 THEN 1 ELSE 0 END) as with_content,
       SUM(CASE WHEN LENGTH(COALESCE(pc.rendered_html, '')) <= 100 THEN 1 ELSE 0 END) as empty
FROM pages p
JOIN sites s ON p.site_id = s.id
LEFT JOIN page_components pc ON pc.page_id = p.id
WHERE s.domain = 'leopardessconsulting.co.uk'
GROUP BY p.name, p.build_status, p.updated_at
ORDER BY p.name;


----------------------------------------------------------------------------------------

unstick

-- Unstick the dispatch group
UPDATE scheduled_tasks
SET last_completed_at = NOW()
WHERE name = 'build-pipeline-trigger';

-- Unstick ch-enrichment
UPDATE scheduled_tasks
SET last_completed_at = NOW()
WHERE name = 'ch-enrichment';

-- Also unstick the reaper and vet tasks that never completed
UPDATE scheduled_tasks
SET last_completed_at = NOW()
WHERE name IN ('stale-orchestration-reaper', 'vet-batch-verify', 'vet-sweep-continue')
  AND last_completed_at IS NULL;


  ----------------------------------------------------------------------------------------

  investigate

SELECT wi.id, wi.spec->>'page_name' as page,
         wi.error, wi.attempt_count, wi.max_attempts,
         wi.claimed_at, wi.completed_at,
         LEFT(wi.summary, 100) as summary
  FROM site_work_items wi
  JOIN sites s ON s.id = wi.site_id
  WHERE s.domain = 'gaswholesalers.com'
    AND wi.item_type = 'content_rewrite'
    AND wi.status = 'failed'
  ORDER BY wi.completed_at DESC
  LIMIT 5;
                    id                  |        page         |                error                 | attempt_count | max_attempts | claimed_at | completed_at |                                               summary
  --------------------------------------+---------------------+--------------------------------------+---------------+--------------+------------+--------------+------------------------------------------------------------------------------------------------------
   b938e672-74d0-4545-9cb0-f04a0bd40eb0 | why-gas-wholesalers | Claim timed out (attempts exhausted) |             3 |            3 |            |              | Add content to why-gas-wholesalers: The why-gas-wholesalers page is the natural home for differentia
  (1 row)


-- Error log entries for gaswholesalers content_rewrite
SELECT agent_type, error_code, LEFT(error_message, 200) as error,
       step_name, action, occurred_at
FROM agent_error_log
WHERE domain = 'gaswholesalers.com'
  AND occurred_at > NOW() - INTERVAL '24 hours'
ORDER BY occurred_at DESC
LIMIT 10;


-- blockers
SELECT wi.id, wi.item_type, wi.spec->>'page_name' as page,
       wi.status, LEFT(wi.error, 200) as error
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE s.domain = 'leopardessconsulting.co.uk'
  AND wi.status IN ('failed', 'needs_human_review')
  AND wi.updated_at > NOW() - INTERVAL '6 hours'
ORDER BY wi.updated_at DESC;


-- status sweep
-- Current state of all non-complete work items
SELECT s.domain, wi.item_type, wi.status, COUNT(*)
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.domain = 'build'
  AND wi.status NOT IN ('complete', 'wont_fix', 'rejected', 'verified')
GROUP BY s.domain, wi.item_type, wi.status
ORDER BY s.domain, wi.status, COUNT(*) DESC;

-- check what they're actually doing
-- And check what dispatch loops are actually doing — they might all be blocked waiting on the claimed items:
SELECT o.orchestration_name, o.status, o.current_step, o.updated_at,
       AGE(NOW(), o.updated_at) as stale_for
FROM orchestrations o
WHERE o.owner_agent_type = 'build-dispatch-loop'
  AND o.status NOT IN ('COMPLETED', 'FAILED', 'CANCELLED')
ORDER BY o.updated_at DESC
LIMIT 10;

-- Check what those stuck dispatches are waiting on:
SELECT o.orchestration_id, o.current_step, o.updated_at,
       AGE(NOW(), o.updated_at) as stale_for,
       o.collected_data->>'work_item_id' as work_item_id,
       LEFT(o.collected_data->'current_item'->>'item_type', 30) as item_type,
       LEFT(o.collected_data->'current_item'->>'handler_agent', 30) as handler
FROM orchestration_states o
WHERE o.owner_agent_type = 'build-dispatch-loop'
  AND o.status NOT IN ('COMPLETED', 'FAILED', 'CANCELLED')
ORDER BY o.updated_at ASC
LIMIT 10;

-- How many items are currently claimed (being worked on)?
SELECT handler_agent, COUNT(*) as claimed
FROM site_work_items
WHERE status = 'claimed' AND domain = 'build'
GROUP BY handler_agent;

-- How many completed in the last hour?
SELECT handler_agent, COUNT(*) as completed
FROM site_work_items
WHERE status = 'complete' AND domain = 'build'
  AND completed_at > NOW() - INTERVAL '1 hour'
GROUP BY handler_agent
ORDER BY completed DESC;

reset
UPDATE site_work_items
     SET status = 'triaged', attempt_count = 0, error = NULL,
         claimed_by = NULL, claimed_at = NULL
     WHERE id = 'b938e672-74d0-4545-9cb0-f04a0bd40eb0';


UPDATE site_work_items
     SET status = 'triaged', attempt_count = 0, error = NULL,
         claimed_by = NULL, claimed_at = NULL
     WHERE id = '4851f6fc-71cf-4160-a270-e03d6d3e0732';

------------------------------------------------------------------------------------------
-- Search the generated HTML for known blocker patterns
-- Search the generated HTML for known blocker patterns
SELECT orchestration_id,
       collected_data->'site_record'->>'domain' as domain,
       collected_data->'input_data'->'spec'->>'page_name' as page_name,
       -- Check for each blocker category
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'coming soon' THEN 'FOUND: coming soon' END as coming_soon,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'placeholder' THEN 'FOUND: placeholder' END as placeholder,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'lorem ipsum' THEN 'FOUND: lorem ipsum' END as lorem,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* '\[insert' THEN 'FOUND: [insert' END as insert_bracket,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'to be added' THEN 'FOUND: to be added' END as to_be_added,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'sample text' THEN 'FOUND: sample text' END as sample_text,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* '\{\{[\s]*[\.\w]+' THEN 'FOUND: template var' END as template_var,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'finetuning' AND collected_data->'site_record'->>'domain' = 'leopardessconsulting.co.uk' THEN 'FOUND: cross-site finetuning' END as cross_site_ft,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'gaswholesalers' AND collected_data->'site_record'->>'domain' != 'gaswholesalers.com' THEN 'FOUND: cross-site gas' END as cross_site_gas,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'john doe|jane doe|acme corp' THEN 'FOUND: placeholder name' END as placeholder_name,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'test@|user@example|name@example' THEN 'FOUND: placeholder email' END as placeholder_email
FROM orchestration_states
WHERE owner_agent_type = 'page-build-handler'
  AND status = 'FAILED'
  AND current_step = 'validate_content'
  AND created_at > NOW() - INTERVAL '3 hours'
ORDER BY last_activity DESC
LIMIT 5;
SELECT orchestration_id,
       collected_data->'site_record'->>'domain' as domain,
       collected_data->'input_data'->'spec'->>'page_name' as page_name,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* '<no value>' THEN 'FOUND: <no value>' END as no_value,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'needs human review' THEN 'FOUND: needs human review' END as needs_review,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'needs_human_review' THEN 'FOUND: needs_human_review' END as needs_review_underscore,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'to be confirmed' THEN 'FOUND: to be confirmed' END as to_be_confirmed,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'to be updated' THEN 'FOUND: to be updated' END as to_be_updated,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'example text' THEN 'FOUND: example text' END as example_text,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'your company' THEN 'FOUND: your company' END as your_company,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'your name here' THEN 'FOUND: your name here' END as your_name,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'todo:' THEN 'FOUND: todo' END as todo,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'fixme:' THEN 'FOUND: fixme' END as fixme,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* '\[your ' THEN 'FOUND: [your' END as bracket_your,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* '\[company' THEN 'FOUND: [company' END as bracket_company,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* '\[add ' THEN 'FOUND: [add' END as bracket_add,
       CASE WHEN collected_data->'page_content'->'response'->>'page_html' ~* 'needs review' THEN 'FOUND: needs review' END as needs_review_plain
FROM orchestration_states
WHERE orchestration_id IN (
  'a4f80fdf-6b1f-4ca0-a4a1-9a63fa114929',
  '09b244a1-3a6c-42eb-9f41-324a75c28603',
  'd81e2869-cdd0-4476-833e-aab90e9fb160'
)
ORDER BY last_activity DESC;

----

-- Nuclear option for vonc.com rerender items only
DELETE FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'vonc.com')
  AND item_type IN ('needs_rerender', 'page_rerender');

-----------------------
 SELECT orchestration_id, status, current_step,
       created_at, updated_at,
       jsonb_object_keys(collected_data) AS data_key
  FROM orchestration_states
 WHERE orchestration_id = '7399a901-e05d-42da-a0ce-39cccb4d8669';

clients_db=# -- After aeacfb4c completes, see if any items got re-queued or hit conflicts
SELECT item_type, status, source, item_key, created_at,
       EXTRACT(EPOCH FROM (NOW() - created_at))::int AS age_s
FROM site_work_items
WHERE site_id = '3103b167-fc73-4a06-a0ab-cbf32294153e'
ORDER BY created_at DESC
LIMIT 30;

-- See the spec history (older becomes is_current=false when newer arrives)
SELECT aspect, is_current, source_agent, created_at
FROM site_specs
WHERE site_id = '3103b167-fc73-4a06-a0ab-cbf32294153e'
ORDER BY aspect, created_at DESC;

-- M) Every work item for this site, chronological — to see who touched what when
SELECT
    to_char(created_at,   'HH24:MI:SS') AS created_t,
    to_char(updated_at,   'HH24:MI:SS') AS updated_t,
    to_char(claimed_at,   'HH24:MI:SS') AS claimed_t,
    to_char(completed_at, 'HH24:MI:SS') AS completed_t,
    item_type,
    LEFT(item_key, 50) AS item_key,
    status,
    LEFT(claimed_by, 30) AS claimed_by,
    attempt_count,
    LEFT(COALESCE(error, ''), 120) AS error_head
  FROM site_work_items
 WHERE site_id = '68f1852b-adcf-4e54-89d7-f32107205649'
 ORDER BY created_at, updated_at;
