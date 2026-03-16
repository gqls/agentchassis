#!/bin/bash
# Trigger rerender-pages agent (uses agent_definitions workflow) multiple pages

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN="leopardessconsulting.co.uk"

echo "========================================="
echo "Triggering rerender-pages agent"
echo "========================================="
echo "  Correlation ID:      $CORRELATION_ID"
echo "  Orchestration ID:    $ORCHESTRATION_ID"
echo "  Site ID:             $SITE_ID"
echo "  Domain:              $DOMAIN"
echo "  Time:                $TIMESTAMP"
echo "========================================="

kubectl -n kafka run -i --rm kcat-rerender-agent \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H message_type=request \
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"orchestrate","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"agent_type":"rerender-pages"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Message sent. Monitor with:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100"


-------------------------------------

# rerender single page
#!/bin/bash
# Trigger rerender-pages agent for a SINGLE PAGE
# Change PAGE_NAME to target a specific page

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN="leopardessconsulting.co.uk"

# Single page filter - use one of these:
PAGE_NAME="Home"  # Match by page name in DB
# PAGE_ID="uuid-of-page"  # Or match by page ID

echo "========================================="
echo "Triggering rerender-pages agent (SINGLE PAGE)"
echo "========================================="
echo "  Correlation ID:      $CORRELATION_ID"
echo "  Orchestration ID:    $ORCHESTRATION_ID"
echo "  Site ID:             $SITE_ID"
echo "  Domain:              $DOMAIN"
echo "  Page Name:           $PAGE_NAME"
echo "  Time:                $TIMESTAMP"
echo "========================================="

kubectl -n kafka run -i --rm kcat-rerender-single \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H message_type=request \
-H client_id=$CLIENT_ID \
-H action=process \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_rerender_agent","steps":{"spawn_rerender_agent":{"action":"spawn_agent","config":{"role":"rerenderer","agent_type":"rerender-pages"},"description":"Spawn rerender-pages agent","next_step":"call_rerender_agent","output_field":"rerender_agent_info"},"call_rerender_agent":{"action":"call_agent","config":{"agent_type":"rerender-pages","target_role":"rerenderer","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain","page_name":"input_data.page_name"},"timeout_seconds":300},"description":"Call rerender-pages agent for single page","next_step":"complete","output_field":"rerender_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["rerender_result"]}}}}},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}","page_name":"${PAGE_NAME}"}}
JSON

echo ""
echo "Message sent. Monitor with:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100"

---

-- one page at a timeout_seconds
-- Update rerender-pages workflow: replace spawn+loop with create_rerender_items
--
-- Old flow: spawn ONE page-rerender → loop { call for each page } → complete
--   Problem: 16 pages = 30min, reaper kills at 20min
--
-- New flow: get_pages → create work items → complete
--   Dispatch loop processes each page independently (own pod, own retry)

UPDATE agent_definitions
SET default_config = '{
  "workflow": {
    "start_step": "check_refresh_components",
    "processing_mode": "orchestrator",
    "timeout_seconds": 300,
    "steps": {
      "check_refresh_components": {
        "action": "conditional",
        "config": {
          "condition": "input_data.spec.refresh_site_components == true OR input_data.refresh_site_components == true",
          "then_step": "render_site_components",
          "else_step": "get_pages"
        },
        "description": "Check if we should re-render header/footer/head first"
      },
      "render_site_components": {
        "action": "render_site_components",
        "config": {
          "slots": ["header", "footer", "head"],
          "input_fields": ["site_id", "domain"],
          "force_rerender": true
        },
        "next_step": "get_pages",
        "description": "Re-render site-level components (header, footer, head)",
        "output_field": "site_components_result"
      },
      "get_pages": {
        "action": "get_pages_for_rerender",
        "config": {
          "input_fields": ["site_id", "domain"],
          "include_statuses": ["deployed", "active"]
        },
        "next_step": "check_pages_exist",
        "description": "Get page metadata for rerender",
        "output_field": "rerender_pages"
      },
      "check_pages_exist": {
        "action": "conditional",
        "config": {
          "condition": "rerender_pages.has_pages == true",
          "then_step": "create_rerender_items",
          "else_step": "complete"
        },
        "description": "Skip if no pages to process"
      },
      "create_rerender_items": {
        "action": "create_rerender_items",
        "config": {
          "pages_field": "rerender_pages.pages",
          "site_id": "rerender_pages.site_id",
          "domain": "rerender_pages.domain"
        },
        "next_step": "complete",
        "description": "Create one work item per page — dispatch loop handles each independently",
        "output_field": "items_result"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {
          "output_fields": ["rerender_pages", "items_result", "site_components_result"]
        },
        "description": "Rerender items created — dispatch loop will process each page"
      }
    }
  }
}'::jsonb
WHERE type = 'rerender-pages';

-- NOTE: Register the new action in registry.go:
--
-- "create_rerender_items": {
--     Handler:     CreateRerenderItemsAction,
--     Category:    "site",
--     Description: "Create per-page rerender work items for dispatch loop",
--     IsLocal:     true,
-- },

-- Also timeout_seconds dropped from 1800 to 300 — the workflow now just
-- creates work items and returns, no more long-running loop.

The new flow is:
```
rerender-pages (one short-lived orchestration, ~30 seconds):
  1. Render header/footer/head (if refresh requested)
  2. Get pages list from DB
  3. INSERT one work item per page (item_type: page_rerender, handler: page-rerender)
  4. Complete

dispatch loop (picks up items individually):
  page_rerender item for "index" → spawns page-rerender → assemble → git commit → done
  page_rerender item for "about" → spawns page-rerender → assemble → git commit → done
  page_rerender item for "blog" → spawns page-rerender → assemble → git commit → done
  ... each independently, own retry count, own pod



