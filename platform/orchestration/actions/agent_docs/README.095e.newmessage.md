[057_image_handling 798e621] amends image creation, untested

# Example workflow configuration with image generation
# This shows how to integrate image generation into agent workflows

workflow:
name: "website-builder-with-images"
version: "1.0"
description: "Multi-section website builder with image generation"

start_step: "spawn_hero_writer"

steps:
# First spawn the hero content writer
spawn_hero_writer:
action: "spawn_agent"
agent_type: "content-creator-hero"
config:
role: "hero_writer"
functional_role: "hero_section_creator"
input_mapping:
business_name: "input_data.business_name"
business_type: "input_data.business_type"
next_steps:
- "generate_hero_image"
- "spawn_features_writer"
await_response: true

    # Generate an image for the hero section
    generate_hero_image:
      action: "generate_image"
      description: "Generate hero image for the website"
      input_mapping:
        # Use the hero content to create an image prompt
        prompt: |
          Create a professional hero image for {{ spawn_hero_writer.hero_headline }}.
          Business: {{ input_data.business_name }}, Type: {{ input_data.business_type }}.
          Style: modern, professional, high-quality photography
        style: "photorealistic"
        width: 1920
        height: 1080
      await_response: true
      next_steps:
        - "process_hero_image"

    # Process the generated hero image
    process_hero_image:
      action: "process_image_response"
      description: "Process and store the hero image URI"
      dependencies:
        - "generate_hero_image"
      next_steps:
        - "aggregate_hero_content"

    # Spawn features section writer
    spawn_features_writer:
      action: "spawn_agent"
      agent_type: "content-creator-features"
      config:
        role: "features_writer"
      input_mapping:
        business_name: "input_data.business_name"
        business_type: "input_data.business_type"
        hero_context: "spawn_hero_writer.hero_content"
      next_steps:
        - "generate_feature_images"
      await_response: true

    # Generate images for each feature
    generate_feature_images:
      action: "parallel_image_generation"
      description: "Generate images for each feature"
      parallel_config:
        max_concurrent: 3
      for_each: "spawn_features_writer.features"
      template:
        action: "generate_image"
        input_mapping:
          prompt: |
            Create an icon or illustration for: {{ item.title }}
            Description: {{ item.description }}
            Style: flat design, modern, colorful
          width: 512
          height: 512
      await_response: true
      next_steps:
        - "aggregate_features"

    # Testimonials section with avatar generation
    spawn_testimonials_writer:
      action: "spawn_agent"
      agent_type: "content-creator-testimonials"
      config:
        role: "testimonials_writer"
      parallel_with:
        - "spawn_features_writer"
      input_mapping:
        business_name: "input_data.business_name"
        business_type: "input_data.business_type"
      next_steps:
        - "generate_avatar_images"
      await_response: true

    # Generate avatar images for testimonials
    generate_avatar_images:
      action: "batch_image_generation"
      description: "Generate avatar images for testimonials"
      batch_config:
        batch_size: 5
        timeout_per_batch: 30
      items:
        - prompt: "Professional headshot, friendly smile, business attire, male, 40s"
        - prompt: "Professional headshot, confident, business casual, female, 30s"
        - prompt: "Professional headshot, approachable, smart casual, male, 50s"
      image_config:
        width: 256
        height: 256
        style: "portrait"
      await_response: true
      next_steps:
        - "aggregate_testimonials"

    # Aggregate all content sections
    aggregate_hero_content:
      action: "merge_data"
      description: "Combine hero content with image"
      dependencies:
        - "spawn_hero_writer"
        - "process_hero_image"
      merge_fields:
        hero_section:
          content: "spawn_hero_writer.hero_content"
          image_uri: "process_hero_image.image_uri"
      next_steps:
        - "final_aggregation"

    aggregate_features:
      action: "merge_data"
      description: "Combine features with images"
      dependencies:
        - "spawn_features_writer"
        - "generate_feature_images"
      merge_fields:
        features_section:
          features: "spawn_features_writer.features"
          images: "generate_feature_images.image_uris"
      next_steps:
        - "final_aggregation"

    aggregate_testimonials:
      action: "merge_data"
      description: "Combine testimonials with avatars"
      dependencies:
        - "spawn_testimonials_writer"
        - "generate_avatar_images"
      merge_fields:
        testimonials_section:
          testimonials: "spawn_testimonials_writer.testimonials"
          avatars: "generate_avatar_images.image_uris"
      next_steps:
        - "final_aggregation"

    # Final website aggregation
    final_aggregation:
      action: "aggregate_website"
      description: "Assemble complete website with all sections and images"
      dependencies:
        - "aggregate_hero_content"
        - "aggregate_features"
        - "aggregate_testimonials"
      output_format: "html"
      include_images: true
      s3_upload: true
      next_steps:
        - "publish_website"

    # Publish the website
    publish_website:
      action: "call_agent"
      agent_type: "site-publisher"
      input_mapping:
        website_data: "final_aggregation.website_html"
        s3_assets: "final_aggregation.asset_uris"
        domain: "input_data.domain"
      await_response: true

# Configuration for image generation behavior
image_generation:
# Multiple image generator containers handle requests
adapter_count: 3

# Load balancing across adapters
consumer_group: "image-generator-adapter-group"

# S3 configuration for image storage
storage:
provider: "s3"
bucket: "${IMAGE_BUCKET}"
path_template: "images/{client_id}/{date}/{image_id}.{format}"

# Circuit breaker settings
resilience:
max_failures: 5
reset_timeout_seconds: 30
request_timeout_seconds: 30

# Topic patterns for dynamic image generation
topics:
# Main topic that all adapters listen to
main_requests: "system.adapter.image-generator.requests"

    # Dynamic topic pattern for specific requests
    # Uses correlation_id + orchestration_id + step_name
    dynamic_pattern: "job.{correlation_short}-{orchestration_short}-image-{step_name}"

# Example of how topics are created dynamically:
# 
# When "generate_hero_image" step executes:
# 1. Creates stable_identity: "3d3bdbff-1de14728-image-generate_hero_image"
# 2. Creates topics:
#    - job.3d3bdbff-1de14728-image-generate_hero_image.requests
#    - job.3d3bdbff-1de14728-image-generate_hero_image.responses
# 3. Image adapter receives request and replies to parent's response topic
# 4. Parent agent receives response and continues workflow
#
# This ensures:
# - Each image request has unique topics
# - Multiple containers don't pick up wrong messages
# - Responses go to the right parent agent
# - Load balancing across adapter instances


---

old message
# 1. FIRST - Stop all consumers by scaling down deployments
kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=0
# Wait for pods to terminate
kubectl -n ai-persona-system wait --for=delete pod -l app=agent-chassis --timeout=60s

# 3. Clear database tables
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "TRUNCATE TABLE processed_messages, orchestration_states, pending_requests CASCADE;"

# 4. Delete ALL spawned agent jobs (not just calculator)
kubectl -n ai-persona-system delete jobs -l spawned-by=orchestrator

# 5. List and clean up job topics (they follow pattern: job.<correlation>.<orchestration>.<step>)
# List all job topics
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--list | grep "^job\."

# Delete all job topics (be careful!)
kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- bash -c '
for topic in $(/opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list | grep "^job\."); do
echo "Deleting topic: $topic"
/opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic "$topic"
done
'

# delete initial topics
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--delete \
--topic system.agent.generic.requests

kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--delete \
--topic system.agent.generic.responses

kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--delete \
--topic system.responses.generic

# reset all offsets
echo 'resetting all offsets'
kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-consumer-groups.sh \
--bootstrap-server localhost:9092 \
--reset-offsets \
--to-earliest \
--all-groups   \
--all-topics    \
--execute

kubectl -n ai-persona-system delete jobs -l agent-type=calculator
kubectl -n ai-persona-system delete pods -l agent-type=calculator

# 6. Scale back up
kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=1


# =======

sleep 5
kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=0

kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-consumer-groups.sh \
--bootstrap-server localhost:9092 \
--reset-offsets \
--to-earliest \
--all-groups   \
--all-topics    \
--execute

kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=1




====
=========






=== run === multi section website builder =====

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="multi-site-$(date +%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
STEP_NAME="client_step_website_request"
CLIENT_ID="demo_client"

echo "========================================="
echo "Multi-Section Website Builder Test"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Request ID:       $REQUEST_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Name:             $ORCHESTRATION_NAME"
echo "  Time:             $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="

kubectl -n kafka run -i --rm kcat-producer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H orchestration_name=$ORCHESTRATION_NAME \
-H step_name=$STEP_NAME \
-H client_id=$CLIENT_ID \
-H message_type=request \
-H action=orchestrate \
-H from_agent_type=user \
-H from_agent_id=$AGENT_ID \
-H responses_topic=system.responses.generic <<EOF
{"action":"orchestrate","config":{"group_type":"multi-section-website-builder"},"input_data":{"business_type":"artisanal bakery","business_name":"Golden Crust Bakery"}}
EOF

echo ""
echo "Message sent. Waiting 20 seconds for processing..."
sleep 20

echo ""
echo "========================================="
echo "Checking Spawned Jobs"
echo "========================================="
kubectl -n ai-persona-system get jobs | grep content-creator | tail -10

echo ""
echo "========================================="
echo "Checking Spawned Pods"
echo "========================================="
kubectl -n ai-persona-system get pods | grep content-creator | tail -10

echo ""
echo "========================================="
echo "Checking Orchestration State in Database"
echo "========================================="
kubectl -n ai-persona-system exec -it postgres-clients-0 -- \
psql -U clients_user -d clients_db -c \
"SELECT
orchestration_id,
status,
current_step,
owner_agent_type,
to_char(created_at, 'HH24:MI:SS') as created,
to_char(updated_at, 'HH24:MI:SS') as updated
FROM orchestration_states
WHERE correlation_id = '$CORRELATION_ID'
ORDER BY created_at DESC
LIMIT 10;"

echo ""
echo "========================================="
echo "Checking for Any Errors in Orchestration"
echo "========================================="
kubectl -n ai-persona-system exec -it postgres-clients-0 -- \
psql -U clients_user -d clients_db -c \
"SELECT
orchestration_id,
status,
error
FROM orchestration_states
WHERE correlation_id = '$CORRELATION_ID'
AND (status = 'FAILED' OR error IS NOT NULL);"

echo ""
echo "========================================="
echo "Monitoring Response Topic"
echo "========================================="
echo "Listening for final response on system.responses.generic..."
echo "Will show correlation_id header and message body"
echo "Press Ctrl+C when you see the final response"
echo ""

kubectl -n kafka run -i --rm kcat-consumer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -C \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.responses.generic \
-o end \
-e \
-f 'CORRELATION: %h{correlation_id}\nBODY: %s\n\n========================================\n\n'





make quick-agent-update ENVIRONMENT=production REGION=uk001
  
  
---
=====
=====
---
test for below

#!/bin/bash

# ============================================================================
# Simple Image Generation Test
# ============================================================================
# This script tests JUST the image generation adapter in isolation,
# without running the full website workflow.
#
# Use this to verify:
# 1. Image adapter is receiving messages
# 2. External API integration works
# 3. S3 upload is functioning
# 4. Response messages are being sent back
# ============================================================================

# Generate unique IDs
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="simple-image-test-$(date +%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)

echo "========================================="
echo "Simple Image Generation Test"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Request ID:       $REQUEST_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Time:             $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="

# Send direct message to image-generator adapter
# Note: This assumes the adapter is listening on system.agent.image-generator.requests
kubectl -n kafka run -i --rm kcat-producer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.image-generator.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H orchestration_name=$ORCHESTRATION_NAME \
-H message_type=request \
-H action=generate_image \
-H from_agent_type=test \
-H from_agent_id=$AGENT_ID \
-H responses_topic=system.agent.image-generator.responses <<EOF
{
"action": "generate_image",
"data": {
"prompt": "A golden retriever puppy playing in a sunny park, professional photography, high quality",
"width": 512,
"height": 512,
"seed": 42
}
}
EOF

echo ""
echo "========================================="
echo "Test message sent!"
echo "========================================="
echo ""
echo "Expected flow:"
echo "  1. Image adapter receives request on system.agent.image-generator.requests"
echo "  2. Adapter calls external image API"
echo "  3. Adapter uploads result to S3"
echo "  4. Adapter sends response to system.agent.image-generator.responses"
echo ""
echo "To monitor:"
echo ""
echo "  1. Image adapter logs:"
echo "     kubectl logs -f deployment/image-generator-adapter -n agent-system"
echo ""
echo "  2. Response topic:"
echo "     kubectl exec -it kafka-0 -n kafka -- kafka-console-consumer.sh \\"
echo "       --bootstrap-server localhost:9092 \\"
echo "       --topic system.agent.image-generator.responses \\"
echo "       --property print.headers=true \\"
echo "       --property print.timestamp=true"
echo ""
echo "  3. S3 bucket:"
echo "     aws s3 ls s3://your-agent-images-bucket/images/ --recursive"
echo ""
echo "Expected response format:"
cat <<RESPONSE
{
"success": true,
"body": {
"data": {
"image_uri": "s3://bucket-name/images/YYYY-MM-DD/correlation-id/image.png",
"width": 512,
"height": 512,
"format": "png"
}
},
"error": null
}
RESPONSE
echo ""
echo "========================================="
echo ""
echo "If this test fails, check:"
echo "  - Image adapter deployment is running"
echo "  - S3 credentials are correct"
echo "  - External API key is valid"
echo "  - Kafka topics exist and are accessible"
echo "  - Network policies allow adapter to reach external API"
echo "========================================="




