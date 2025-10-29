Generic Agent
↓ spawns
Image-Generator Agent (dynamic)
↓ sends to
system.adapter.image-generator.requests (Kafka)
↓ consumed by
Image Generator Adapter (3 replicas)
↓ calls
Stability AI API → Generates image
↓ uploads to
Backblaze B2 Storage
↓ responds to
Parent Agent's Response Topic
↓ received by
Image-Generator Agent → Generic Agent

# Image Generation Agent Orchestration Flow Analysis

## Current Issue (What's Wrong)

```
[Parent Agent/Orchestrator]
    |
    ├─1─> spawn_image_creator action
    |     Creates: image-generator agent with role "image_creator"
    |     Topic: job.f484fffb-d6a591f5-image-generator-spawn_image_creator.requests
    |     ✅ WORKING - Agent initializes successfully
    |
    ├─2─> generate_hero_image step (call_agent action)
    |     Should send to: spawned image-generator agent
    |     ❌ PROBLEM: The agent's workflow starts but...
    |
    └─────> [Image Generator Agent (role: image_creator)]
            |
            ├─ Workflow step: "generate" 
            ├─ Action: "generate_image" (GenerateImageAction)
            |
            └─❌ BYPASSES itself and sends directly to:
                 system.adapter.image-generator.requests
                 
[Image Generator Agent] -- waiting idle on its topic --
                          Never receives proper work messages
                          
[Adapter] <- Receives message but response goes nowhere useful
```

## Root Cause

The `GenerateImageAction` function in `generate_image_actions.go` was designed as if it was being called from the PARENT agent, not from within the image-generator agent itself. It was:

1. Creating new dynamic topics (unnecessary)
2. Sending directly to adapter topic (bypassing agent orchestration)
3. Not using the agent's own response topic properly

## Correct Flow (After Fix)

```
[Parent Agent/Orchestrator]
    |
    ├─1─> spawn_image_creator action
    |     Creates: image-generator agent with role "image_creator"
    |     Topic: job.{correlation}-{orch}-image-generator-{step}.requests
    |     ✅ Agent initializes and responds
    |
    ├─2─> call_agent action (generate_hero_image step)
    |     Sends work to: spawned agent's requests topic
    |     With: prompt, width, height parameters
    |
    └────> [Image Generator Agent]
            |
            ├─ Receives message on its requests topic
            ├─ Executes workflow step: "generate"
            ├─ Calls GenerateImageAction (corrected)
            |
            ├─3─> Sends to adapter: system.adapter.image-generator.requests
            |     With reply_to_topic: agent's own responses topic
            |
            └──< [Adapter]
                 ├─ Generates image via external API
                 ├─ Uploads to S3
                 └─4─> Responds to: agent's responses topic
                       
            [Image Generator Agent]
            ├─ Receives adapter response
            ├─ Processes response (ProcessImageResponse)
            └─5─> Responds to parent with image URI
```

## Key Topics in Play

1. **Parent's Responses Topic**: `system.agent.generic.responses`
    - Where parent receives responses from spawned agents

2. **Image Generator Agent's Topics**:
    - Requests: `job.{id}-image-generator-{step}.requests`
    - Responses: `job.{id}-image-generator-{step}.responses`

3. **Adapter Topic**: `system.adapter.image-generator.requests`
    - Shared topic for all image generation adapters
    - Adapters form a consumer group for load balancing

## Message Flow Timeline

```
Time    Component                Action                                  Topic
----    ---------                ------                                  -----
T0      Parent                  → Spawn image_creator agent            → Agent requests topic
T1      Image Agent             ← Initialize                           ← Agent requests topic  
T2      Image Agent             → Init response                        → Parent responses topic
T3      Parent                  → Call agent with image request        → Agent requests topic
T4      Image Agent             ← Receive work request                 ← Agent requests topic
T5      Image Agent             → Send to adapter                      → Adapter topic
T6      Adapter                 ← Receive generation request           ← Adapter topic
T7      Adapter                 → Generate & upload image              → (External API + S3)
T8      Adapter                 → Send response                        → Agent responses topic
T9      Image Agent             ← Receive adapter response             ← Agent responses topic
T10     Image Agent             → Send final response to parent        → Parent responses topic
T11     Parent                  ← Receive image URI                    ← Parent responses topic
```

## Configuration Requirements

### Image Generator Agent
- Must have workflow with "generate" step
- Action must be "generate_image"
- Processing mode: "orchestrated" (not "adapter")

### Adapter
- Listens to: `system.adapter.image-generator.requests`
- Consumer group: `image-generator-adapter-group`
- Must respect `reply_to_topic` in request body

### Environment Variables Needed
- `RESPONSES_TOPIC`: Agent's responses topic
- `REQUESTS_TOPIC`: Agent's requests topic
- `PARENT_RESPONSES_TOPIC`: Where to send final responses
- `IMAGE_API_KEY`: For external image generation API
- `IMAGE_API_URL`: External API endpoint
- S3/MinIO credentials for storage

## Testing Steps

1. Deploy corrected `GenerateImageAction`
2. Ensure image-generator agent receives work on its topic
3. Verify adapter receives request with correct reply_to_topic
4. Confirm adapter response reaches agent's responses topic
5. Validate parent receives final image URI

## Common Pitfalls to Avoid

1. ❌ Don't bypass agent orchestration
2. ❌ Don't create unnecessary dynamic topics
3. ❌ Don't confuse parent actions with agent actions
4. ❌ Don't forget to set reply_to_topic for adapter
5. ❌ Don't send adapter responses to wrong topic

## Summary

The issue was a fundamental misunderstanding of where `GenerateImageAction` runs:
- **Wrong**: Thought it ran in parent, so sent directly to adapter
- **Right**: It runs in image-generator agent, should respect orchestration

The fix ensures proper agent orchestration flow where each component communicates through its designated topics, maintaining the chain of responsibility.