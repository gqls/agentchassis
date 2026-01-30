/*not sql just the trigger - using an inline workflow, but we now have an agent workflow type='rerender-agent':
#!/bin/bash
# Re-render all deployed pages with current components
# This applies component updates (head, header, footer, CSS links) without regenerating content

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN="leopardessconsulting.co.uk"

echo "========================================="
echo "Re-rendering site pages with updated components"
echo "========================================="
echo "  Correlation ID:      $CORRELATION_ID"
echo "  Orchestration ID:    $ORCHESTRATION_ID"
echo "  Site ID:             $SITE_ID"
echo "  Domain:              $DOMAIN"
echo "  Time:                $TIMESTAMP"
echo "========================================="

kubectl -n kafka run -i --rm kcat-rerender-pages \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"rerender_pages","steps":{"rerender_pages":{"action":"rerender_site_pages","config":{"site_id_field":"input_data.site_id","domain_field":"input_data.domain","include_statuses":["deployed","active"]},"description":"Re-render all pages with current components","next_step":"deploy_pages","output_field":"rerender_result"},"deploy_pages":{"action":"loop","config":{"items_field":"rerender_result.pages","item_variable":"current_page","max_iterations":50,"sub_workflow":{"start_step":"commit_page","steps":{"commit_page":{"action":"git_commit","config":{"domain_field":"input_data.domain","content_field":"current_page.html","page_field":"current_page","commit_message":"Re-render page: {{.filename}}"},"description":"Commit re-rendered page via git adapter","next_step":"done","output_field":"commit_result"},"done":{"action":"loop_complete"}}}},"description":"Deploy each re-rendered page","next_step":"complete","output_field":"deploy_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["rerender_result","deploy_result"]}}}}},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Message sent. Monitor with:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100"*/

---

-- sql for workflow:


-- Rerender Pages Agent
-- Re-assembles all deployed pages with current components
-- Use for applying component updates without regenerating content

INSERT INTO agent_definitions (
    type,
    version,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    input_contract,
    output_contract
) VALUES (
    'rerender-pages',
    1,
    'Rerender Pages Agent',
    'Re-assembles all deployed pages with current components (head, header, footer). Use when component templates are updated, CSS links change, or navigation structure changes.',
    'builder',
    '{
        "processing_mode": "orchestrator",
        "timeout_seconds": 300,
        "workflow": {
            "start_step": "rerender_pages",
            "steps": {
                "rerender_pages": {
                    "action": "rerender_site_pages",
                    "config": {
                        "site_id_field": "input_data.site_id",
                        "domain_field": "input_data.domain",
                        "include_statuses": ["deployed", "active"]
                    },
                    "description": "Re-render all pages with current components",
                    "next_step": "deploy_pages",
                    "output_field": "rerender_result"
                },
                "deploy_pages": {
                    "action": "loop",
                    "config": {
                        "items_field": "rerender_result.pages",
                        "item_variable": "current_page",
                        "max_iterations": 50,
                        "sub_workflow": {
                            "start_step": "commit_page",
                            "steps": {
                                "commit_page": {
                                    "action": "git_commit",
                                    "config": {
                                        "domain_field": "input_data.domain",
                                        "content_field": "current_page.html",
                                        "page_field": "current_page",
                                        "commit_message": "Re-render page: {{.slug}}"
                                    },
                                    "description": "Commit re-rendered page via git adapter",
                                    "next_step": "done",
                                    "output_field": "commit_result"
                                },
                                "done": {
                                    "action": "loop_complete"
                                }
                            }
                        }
                    },
                    "description": "Deploy each re-rendered page",
                    "next_step": "complete",
                    "output_field": "deploy_result"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {
                        "output_fields": ["rerender_result", "deploy_result"]
                    }
                }
            }
        }
    }',
    true,
    '["rerender", "maintenance", "components", "css"]',
    '{"required": [], "optional": ["site_id", "domain"], "description": "Provide either site_id or domain to identify the site"}',
    '{"produces": {"rerender_result": "Pages rendered with updated components", "deploy_result": "Git commit results for each page"}}'
)
ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                                                                                                                                                                                                                                                                                                                                                                                      description = EXCLUDED.description,
                                                                                                                                                                                                                                                                                                                                                                                                      default_config = EXCLUDED.default_config,
                                                                                                                                                                                                                                                                                                                                                                                                      capabilities = EXCLUDED.capabilities,
                                                                                                                                                                                                                                                                                                                                                                                                      input_contract = EXCLUDED.input_contract,
                                                                                                                                                                                                                                                                                                                                                                                                      output_contract = EXCLUDED.output_contract,
                                                                                                                                                                                                                                                                                                                                                                                                      updated_at = NOW();
