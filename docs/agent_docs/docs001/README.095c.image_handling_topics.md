https://claude.ai/chat/63dee3cb-e441-453a-a2e4-b85340e5e6e0
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

--

# Image Generation with Dynamic Topics and S3 Integration

## Overview

This implementation provides a robust image generation system that follows the same orchestration patterns as your existing agent spawning architecture. It ensures proper message isolation through dynamic topic creation and stores generated images in S3.

## Key Features

1. **Dynamic Topic Creation**: Each image generation request creates unique topics to prevent message collision
2. **S3 Integration**: All generated images are stored in S3 with organized paths
3. **Load Balancing**: Multiple image adapter containers share load through consumer groups
4. **Parent Response Pattern**: Image generators always respond to the parent's response topic
5. **Circuit Breaker**: Resilient handling of external API failures

## Architecture

### Topic Flow

```
Parent Agent (e.g., website-builder)
    |
    | 1. Creates dynamic topics:
    |    - job.{correlationId}-{orchId}-image-{stepName}.requests
    |    - job.{correlationId}-{orchId}-image-{stepName}.responses
    |
    | 2. Sends request to: system.adapter.image-generator.requests
    v
Image Generator Adapter (one of many)
    |
    | 3. Processes request:
    |    - Calls external image API
    |    - Uploads to S3
    |
    | 4. Responds to parent's response topic
    v
Parent Agent receives response and continues
```

### Files Created

1. **image_actions.go**: Action handlers for image generation
2. **dynamic_image_adapter.go**: Updated adapter with dynamic topic support
3. **workflow_with_images_example.yaml**: Example workflow configuration
4. **README.md**: This documentation

## Implementation Steps

### 1. Add Image Actions to Your Agent Chassis

Copy `image_actions.go` to your actions directory:
```bash
cp /home/claude/image_actions.go internal/actions/image_actions.go
```

Update your action registry to include the new actions:
```go
// In your action registry
registry.Register("generate_image", actions.GenerateImageAction)
registry.Register("call_image_generator", actions.CallImageGeneratorAction)
registry.Register("process_image_response", actions.ProcessImageResponse)
```

### 2. Deploy the Dynamic Image Adapter

Replace your existing image adapter with the dynamic version:
```bash
cp /home/claude/dynamic_image_adapter.go internal/adapters/imagegenerator/dynamic_adapter.go
```

Update the adapter's main.go:
```go
adapter, err := imagegenerator.NewDynamicImageAdapter(ctx, cfg, appLogger)
```

### 3. Configure Environment Variables

Set the required environment variables for the image adapters:

```bash
# S3 Configuration
export S3_ENDPOINT="https://s3.amazonaws.com"
export S3_REGION="us-east-1"
export S3_ACCESS_KEY="your-access-key"
export S3_SECRET_KEY="your-secret-key"
export IMAGE_BUCKET="your-image-bucket"
export S3_USE_SSL="true"

# Image API Configuration
export IMAGE_API_URL="https://api.stability.ai/v1/generation/stable-diffusion-v1-6/text-to-image"
export IMAGE_API_KEY="your-api-key"

# Kafka Configuration
export KAFKA_BROKERS="kafka1:9092,kafka2:9092"
```

### 4. Deploy Multiple Adapter Instances

Deploy 3-5 image adapter containers for load balancing:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: image-generator-adapter
spec:
  replicas: 3
  selector:
    matchLabels:
      app: image-generator-adapter
  template:
    metadata:
      labels:
        app: image-generator-adapter
    spec:
      containers:
      - name: adapter
        image: your-registry/image-generator-adapter:latest
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        # ... other env vars ...
```

## Usage in Workflows

### Simple Image Generation

```yaml
generate_hero_image:
  action: "generate_image"
  input_mapping:
    prompt: "A beautiful landscape for a travel website"
    width: 1920
    height: 1080
    style: "photorealistic"
  await_response: true
  next_steps:
    - "process_hero_image"
```

### Parallel Image Generation

```yaml
generate_product_images:
  action: "parallel_image_generation"
  parallel_config:
    max_concurrent: 3
  for_each: "products"
  template:
    action: "generate_image"
    input_mapping:
      prompt: "Product photo of {{ item.name }}: {{ item.description }}"
      width: 800
      height: 800
  await_response: true
```

### Using Generated Images

```yaml
process_hero_image:
  action: "process_image_response"
  dependencies:
    - "generate_hero_image"
  next_steps:
    - "build_webpage"

build_webpage:
  action: "execute_llm_prompt"
  input_mapping:
    hero_image_url: "process_hero_image.image_uri"
    template: |
      <div class="hero">
        <img src="{{ hero_image_url }}" alt="Hero Image">
        <h1>{{ hero_headline }}</h1>
      </div>
```

## How It Works

### 1. Dynamic Topic Creation

When an agent executes a `generate_image` action:

```go
// Create stable identity
stableIdentity := fmt.Sprintf("%s-%s-image-generator-%s",
    correlationID[:8],
    orchestrationID[:8], 
    stepName,
)

// Create unique topics
requestsTopic := fmt.Sprintf("job.%s.requests", stableIdentity)
responsesTopic := fmt.Sprintf("job.%s.responses", stableIdentity)
```

### 2. Request Routing

The request includes metadata for proper routing:

```go
request := &ImageRequest{
    Headers: execContext.ToRequestHeaders(),
    Body: {
        Action: "generate",
        Data: {
            Prompt: prompt,
            Width: width,
            Height: height,
        },
        ReplyToTopic: parentResponsesTopic,  // Critical: where to send response
        Metadata: {
            StableIdentity: stableIdentity,
            RequestingAgent: agentType,
        },
    },
}
```

### 3. Load Balancing

Multiple adapter instances share the load:
- All adapters join the same consumer group: `image-generator-adapter-group`
- Kafka automatically distributes messages across instances
- Each message is processed by only one adapter

### 4. S3 Storage

Images are stored with organized paths:
```
s3://bucket/images/{client_id}/{date}/{image_id}.png
```

Example:
```
s3://mybucket/images/demo_client/20251027/a3b2c1d4-e5f6-7890-abcd-ef1234567890.png
```

### 5. Response Routing

The adapter sends responses to the parent's topic:

```go
// Adapter determines response topic
responseTopic := request.Body.ReplyToTopic
if responseTopic == "" {
    responseTopic = request.Headers.ParentResponsesTopic
}

// Send response to parent
adapter.sendResponse(responseTopic, response, logger)
```

## Monitoring and Debugging

### Check Adapter Health

```bash
curl http://image-adapter-pod:9090/health
```

Response:
```json
{
  "status": "healthy",
  "adapter": "image-generator",
  "circuit_breaker": "closed",
  "circuit_counts": {
    "requests": 1000,
    "total_successes": 980,
    "total_failures": 20
  }
}
```

### View Logs

```bash
# View adapter logs
kubectl logs -l app=image-generator-adapter -f

# Filter by correlation ID
kubectl logs -l app=image-generator-adapter | grep "correlation_id:3d3bdbff"
```

### Check Topics

```bash
# List dynamic image topics
kafka-topics --bootstrap-server kafka:9092 --list | grep "job.*image"

# View messages in a topic
kafka-console-consumer --bootstrap-server kafka:9092 \
  --topic job.3d3bdbff-1de14728-image-generate_hero.responses \
  --from-beginning
```

### S3 Verification

```bash
# List generated images
aws s3 ls s3://your-bucket/images/demo_client/ --recursive

# Download an image
aws s3 cp s3://your-bucket/images/demo_client/20251027/image.png ./
```

## Troubleshooting

### Issue: Images not generating

1. Check adapter is running:
```bash
kubectl get pods -l app=image-generator-adapter
```

2. Verify environment variables:
```bash
kubectl exec -it image-adapter-pod -- env | grep IMAGE
```

3. Check Kafka connectivity:
```bash
kafka-topics --bootstrap-server kafka:9092 --list
```

### Issue: Wrong response topic

Ensure the parent agent sets `PARENT_RESPONSES_TOPIC`:
```go
os.Setenv("PARENT_RESPONSES_TOPIC", "job.parent.responses")
```

### Issue: S3 upload failures

1. Verify S3 credentials:
```bash
aws s3 ls --endpoint-url=$S3_ENDPOINT
```

2. Check bucket permissions:
```bash
aws s3api get-bucket-acl --bucket $IMAGE_BUCKET
```

### Issue: Circuit breaker open

The adapter automatically handles this with exponential backoff. To manually reset:
```bash
# Restart the adapter pod
kubectl rollout restart deployment/image-generator-adapter
```

## Performance Considerations

### Scaling Guidelines

- **Adapter Instances**: 1 per 10 requests/second
- **Kafka Partitions**: 2-3 per adapter instance
- **S3 Connections**: Use connection pooling
- **Circuit Breaker**: Adjust thresholds based on API limits

### Optimization Tips

1. **Batch Processing**: Group similar prompts
2. **Caching**: Cache frequently used images
3. **Compression**: Compress images before S3 upload
4. **CDN**: Use CloudFront for image delivery

## Security

1. **API Keys**: Store in Kubernetes secrets
2. **S3 Access**: Use IAM roles with minimal permissions
3. **Network**: Use VPC endpoints for S3
4. **Encryption**: Enable S3 encryption at rest

## Future Enhancements

1. **Image Variations**: Support multiple variations per prompt
2. **Style Transfer**: Apply styles to existing images
3. **Image Editing**: Support inpainting and outpainting
4. **Webhook Support**: Notify on completion via webhooks
5. **Cost Tracking**: Track API usage per client
6. **Image Optimization**: Auto-resize and compress
7. **Prompt Templates**: Reusable prompt patterns
8. **A/B Testing**: Test different prompts/styles

## Summary

This implementation provides a production-ready image generation system that:
- Integrates seamlessly with your existing agent orchestration
- Prevents message collision through dynamic topics
- Scales horizontally with multiple adapter instances
- Stores images reliably in S3
- Handles failures gracefully with circuit breakers
- Follows your established patterns for agent communication

The key insight is treating image generation like any other agent interaction, using the same topic patterns and response routing that make your multi-level agent spawning work correctly.