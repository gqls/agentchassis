#!/bin/bash
# run this script:
# kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- bash < scripts/operate/build-a-website.sh
# monitor it
# Watch the website-builder logs
# kubectl -n ai-persona-system logs $(kubectl -n ai-persona-system get pods | grep agent-website-builder | awk '{print $1}') -f --tail=50
# kubectl -n ai-persona-system logs $(kubectl -n ai-persona-system get pods | grep agent-domain-analyst | awk '{print $1}') -f --tail=20
# kubectl -n ai-persona-system logs $(kubectl -n ai-persona-system get pods | grep agent-site-architect | awk '{print $1}') -f --tail=20
# kubectl -n ai-persona-system logs $(kubectl -n ai-persona-system get pods | grep agent-content-creator | awk '{print $1}') -f --tail=20
# kubectl -n ai-persona-system logs $(kubectl -n ai-persona-system get pods | grep agent-html-developer | awk '{print $1}') -f --tail=20

#!/bin/bash

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "$(date +%s)-$(shuf -i 1000-9999 -n 1)")
REQUEST_ID="req-$(date +%s)"

echo "Sending to generic orchestrator with call_agent (will spawn if needed):"
printf "correlation_id:$CORRELATION_ID,request_id:$REQUEST_ID,client_id:demo_client,agent_instance_id:generic-001,fuel_budget:5000\t{\"action\":\"call_agent\",\"config\":{\"agent_type\":\"website-builder\"},\"data\":{\"message\":\"Build a professional website for TechStart Solutions...\"}}\n" | /opt/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic system.agent.generic.process \
  --property parse.headers=true \
  --property headers.delimiter=$'\t'




# Step 1: Spawn the website-builder
#CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "$(date +%s)-$(shuf -i 1000-9999 -n 1)")
#REQUEST_ID="req-$(date +%s)"
#
#echo "Step 1: Spawning website-builder agent"
#printf "correlation_id:$CORRELATION_ID,request_id:$REQUEST_ID,client_id:demo_client,agent_instance_id:generic-001,fuel_budget:1000\t{\"action\":\"spawn_agent\",\"config\":{\"agent_type\":\"website-builder\"}}\n" | /opt/kafka/bin/kafka-console-producer.sh \
#  --bootstrap-server localhost:9092 \
#  --topic system.agent.generic.process \
#  --property parse.headers=true \
#  --property headers.delimiter=$'\t'
#
#echo "Waiting for agent to spawn..."
#sleep 10
#
## Step 2: Send the actual website build request
#CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "$(date +%s)-$(shuf -i 1000-9999 -n 1)")
#REQUEST_ID="req-$(date +%s)"
#
#echo "Step 2: Sending website build request"
#printf "correlation_id:$CORRELATION_ID,request_id:$REQUEST_ID,client_id:demo_client,agent_instance_id:website-builder-001,fuel_budget:5000\t{\"action\":\"process\",\"data\":{\"message\":\"Build a professional website for TechStart Solutions...\"}}\n" | /opt/kafka/bin/kafka-console-producer.sh \
#  --bootstrap-server localhost:9092 \
#  --topic system.agent.website-builder.process \
#  --property parse.headers=true \
#  --property headers.delimiter=$'\t'






# -------

# Generate IDs
#CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "$(date +%s)-$(shuf -i 1000-9999 -n 1)")
#REQUEST_ID="req-$(date +%s)"
#WORKFLOW_ID="website-$(date +%s)"
#
#echo "Sending website build request:"
#echo "  correlation_id: $CORRELATION_ID"
#echo "  request_id: $REQUEST_ID"
#echo "  workflow_id: $WORKFLOW_ID"
#
## Create the message with the correct structure
## The processor expects: { "action": "...", "data": {...} }
#printf "correlation_id:$CORRELATION_ID,request_id:$REQUEST_ID,client_id:demo_client,agent_instance_id:generic-001,fuel_budget:5000\t{\"action\":\"process\",\"workflow_id\":\"${WORKFLOW_ID}\",\"data\":{\"task\":\"build_website\",\"website_config\":{\"business_name\":\"TechStart Solutions\",\"business_type\":\"Technology Consulting\",\"industry\":\"Software Development\",\"target_audience\":\"Startups and SMBs\",\"brand\":{\"primary_color\":\"#2563eb\",\"secondary_color\":\"#10b981\",\"font_family\":\"Inter\",\"tone\":\"professional\",\"style\":\"modern\"},\"pages\":[{\"type\":\"home\",\"title\":\"TechStart Solutions - Technology Consulting\",\"sections\":[\"hero\",\"services\",\"about\",\"testimonials\",\"cta\"]},{\"type\":\"about\",\"title\":\"About Us\",\"content\":\"Leading technology consulting firm helping startups scale\"},{\"type\":\"services\",\"title\":\"Our Services\",\"services\":[\"Cloud Architecture\",\"DevOps Implementation\",\"Digital Transformation\",\"Technical Advisory\"]},{\"type\":\"contact\",\"title\":\"Contact Us\",\"include_form\":true}],\"features\":[\"responsive_design\",\"seo_optimized\",\"contact_forms\",\"analytics_ready\",\"fast_loading\"],\"content\":{\"tagline\":\"Accelerate Your Digital Journey\",\"hero_text\":\"We help startups and growing businesses build scalable technology solutions\",\"cta_text\":\"Start Your Transformation Today\",\"value_proposition\":\"10+ years of experience helping 200+ startups succeed\"}}}}\n" | /opt/kafka/bin/kafka-console-producer.sh \
#  --bootstrap-server localhost:9092 \
#  --topic system.agent.generic.process \
#  --property parse.headers=true \
#  --property headers.delimiter=$'\t'
#
#echo "Message sent successfully!"

# Send to the website-builder agent
#CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "$(date +%s)-$(shuf -i 1000-9999 -n 1)")
#REQUEST_ID="req-$(date +%s)"
#
#echo "Sending website build request to website-builder agent:"
#echo "  correlation_id: $CORRELATION_ID"
#echo "  request_id: $REQUEST_ID"
#echo "  agent_instance_id: website-builder-001"
#
## Include a 'message' field that describes what we want
#printf "correlation_id:$CORRELATION_ID,request_id:$REQUEST_ID,client_id:demo_client,agent_instance_id:website-builder-001,fuel_budget:5000\t{\"action\":\"process\",\"data\":{\"message\":\"Build a professional website for TechStart Solutions, a technology consulting company that helps startups and SMBs with cloud architecture, DevOps, and digital transformation.\",\"website_config\":{\"business_name\":\"TechStart Solutions\",\"business_type\":\"Technology Consulting\",\"industry\":\"Software Development\",\"target_audience\":\"Startups and SMBs\",\"brand\":{\"primary_color\":\"#2563eb\",\"secondary_color\":\"#10b981\",\"font_family\":\"Inter\",\"tone\":\"professional\",\"style\":\"modern\"},\"pages\":[{\"type\":\"home\",\"title\":\"TechStart Solutions - Technology Consulting\",\"sections\":[\"hero\",\"services\",\"about\",\"testimonials\",\"cta\"]},{\"type\":\"about\",\"title\":\"About Us\",\"content\":\"Leading technology consulting firm helping startups scale\"},{\"type\":\"services\",\"title\":\"Our Services\",\"services\":[\"Cloud Architecture\",\"DevOps Implementation\",\"Digital Transformation\",\"Technical Advisory\"]},{\"type\":\"contact\",\"title\":\"Contact Us\",\"include_form\":true}],\"features\":[\"responsive_design\",\"seo_optimized\",\"contact_forms\",\"analytics_ready\",\"fast_loading\"],\"content\":{\"tagline\":\"Accelerate Your Digital Journey\",\"hero_text\":\"We help startups and growing businesses build scalable technology solutions\",\"cta_text\":\"Start Your Transformation Today\",\"value_proposition\":\"10+ years of experience helping 200+ startups succeed\"}}}}\n" | /opt/kafka/bin/kafka-console-producer.sh \
#  --bootstrap-server localhost:9092 \
#  --topic system.agent.website-builder.process \
#  --property parse.headers=true \
#  --property headers.delimiter=$'\t'
#
#echo "Message sent!"