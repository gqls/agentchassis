#!/bin/bash
# trigger-adopt-site.sh — Adopt an existing site into the system
#
# Usage: ./trigger-adopt-site.sh <domain> [url]
#   domain: e.g. mortgagecalculator.co.uk
#   url:    e.g. https://mortgagecalculator.co.uk (defaults to https://<domain>)

DOMAIN="${1:?Usage: ./trigger-adopt-site.sh <domain> [url]}"
URL="${2:-https://$DOMAIN}"

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
  -H responses_topic=system.generic.responses <<JSON
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
