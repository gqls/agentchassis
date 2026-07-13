SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

#SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
#DOMAIN="finetuning.uk"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"


echo "=== Rerender: $CORRELATION_ID ==="

kubectl -n kafka run -i --rm kcat-rerender-$(date +%s) \
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
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_rerender","steps":{"spawn_rerender":{"action":"spawn_agent","config":{"role":"rerenderer","agent_type":"rerender-pages"},"output_field":"rerender_agent","next_step":"call_rerender"},"call_rerender":{"action":"call_agent","config":{"agent_type":"rerender-pages","target_role":"rerenderer","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain","refresh_site_components":"input_data.refresh_site_components"},"timeout_seconds":1800},"output_field":"rerender_result","next_step":"complete"},"complete":{"action":"complete_workflow","config":{"output_fields":["rerender_result"]}}}}},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}","refresh_site_components":true}}
JSON

echo "Monitor: kubectl -n ai-persona-system logs -f -l agent-type=rerender-pages --tail=50"



SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"
--
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"



-- 1. Fix gaswholesalers.com email to match contact page
UPDATE sites SET email = 'info@gaswholesalers.com', phone = '+44 (0) 7934 524 911'
WHERE id = '5fe15466-4e2e-4ff2-981e-98c1b7074002';

-- 2. Fix gaswholesalers hero images (same pattern as finetuning)
UPDATE page_components pc
SET rendered_html = REPLACE(
    rendered_html,
    'background: linear-gradient(135deg, var(--primary-color, #1a1a2e) 0%, var(--secondary-color, #16213e) 50%, var(--accent-color, #0f3460) 100%);',
    'background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url(''/assets/images/hero.jpg''); background-size: cover; background-position: center;'
),
updated_at = NOW()
FROM pages p
WHERE pc.page_id = p.id
  AND p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND pc.rendered_html LIKE '%data-component="hero%"'
  AND pc.rendered_html LIKE '%linear-gradient(135deg, var(--primary-color%';

