https://claude.ai/chat/f372c352-7833-47e0-b1e8-e43ad69e860a

#!/bin/bash
# QUICK START - Execute commands in this exact order

cat << 'EOF'
╔═══════════════════════════════════════════════════════════════════════════╗
║                                                                           ║
║  ROBOT HANDS WEBSITE BUILDER - QUICK START GUIDE                        ║
║  Step-by-step commands to get image generation working                   ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

PHASE 1: TEST IMAGE ADAPTER (DO THIS FIRST!)
═══════════════════════════════════════════════════════════════════════════

Step 1a: Run interactive debugger
─────────────────────────────────
./debug_image_adapter_step_by_step.sh

This will check:
✓ Adapter deployment exists
✓ Pods are running  
✓ Secrets configured
✓ Kafka topics exist
✓ Send test message
✓ Verify S3 upload

If any check fails, fix it before continuing!

─────────────────────────────────────────────────────────────────────────────

Step 1b: Run simple image test
─────────────────────────────────
./test_robot_hands_image.sh

In another terminal, watch logs:
kubectl logs -f -n agent-system -l app=image-generator-adapter

Expected logs:
→ Received request on system.agent.image-generator.requests
→ Calling external image API...
→ Uploading to S3...
→ Sending response...

Verify response received:
kubectl exec -it kafka-0 -n kafka -- kafka-console-consumer.sh \
--bootstrap-server localhost:9092 \
--topic system.agent.image-generator.responses \
--from-beginning --max-messages 10

Should see:
{"success":true,"body":{"data":{"image_uri":"s3://..."}}}

Check S3:
aws s3 ls s3://YOUR-BUCKET/images/ --recursive

─────────────────────────────────────────────────────────────────────────────

🛑 STOP! Do not proceed until:
✓ Adapter receives messages
✓ External API calls succeed
✓ Images upload to S3
✓ Response messages are sent

═══════════════════════════════════════════════════════════════════════════

PHASE 2: CREATE WORKFLOW IN DATABASE
═══════════════════════════════════════════════════════════════════════════

Step 2a: Port forward to database
─────────────────────────────────
kubectl port-forward svc/postgres-clients 5432:5432 -n ai-persona-system

Keep this running in a separate terminal!

─────────────────────────────────────────────────────────────────────────────

Step 2b: Run SQL Step 1 (create group)
─────────────────────────────────────
psql -h localhost -p 5432 -U clients_user -d clients_db \
-f create_image_workflow_group.sql

Look for output like:
id                   | name                          | group_type
---------------------|-------------------------------|---------------------------
abc12345-6789...     | Website Builder with Images   | website-builder-with-images

SAVE THIS UUID! You need it for the next step.

─────────────────────────────────────────────────────────────────────────────

Step 2c: Edit SQL file with your UUID
─────────────────────────────────────
1. Open create_image_workflow_group.sql
2. Find Step 2 section (around line 70)
3. Replace ALL instances of '<NEW_GROUP_ID>' with your UUID
4. Uncomment the INSERT statement
5. Save the file

─────────────────────────────────────────────────────────────────────────────

Step 2d: Run SQL Step 2 (add agent configs)
─────────────────────────────────────────
psql -h localhost -p 5432 -U clients_user -d clients_db \
-f create_image_workflow_group.sql

(It will skip Step 1 since it already exists)

─────────────────────────────────────────────────────────────────────────────

Step 2e: Verify workflow exists
─────────────────────────────────
psql -h localhost -p 5432 -U clients_user -d clients_db -c \
"SELECT id, name, group_type, version, is_active
FROM agent_groups
WHERE group_type = 'website-builder-with-images';"

Should show:
is_active | t (true)
version   | 1

─────────────────────────────────────────────────────────────────────────────

Step 2f: Check workflow steps
─────────────────────────────────
psql -h localhost -p 5432 -U clients_user -d clients_db -c \
"SELECT jsonb_object_keys(workflow_config->'steps') as step_name
FROM agent_groups
WHERE group_type = 'website-builder-with-images';"

Should show 4 steps:
- spawn_hero_writer
- generate_hero
- generate_hero_image
- complete

═══════════════════════════════════════════════════════════════════════════

PHASE 3: TEST FULL WORKFLOW
═══════════════════════════════════════════════════════════════════════════

Step 3a: Run workflow test
─────────────────────────────────
./test_workflow_with_images.sh

This sends a message to the generic orchestrator with:
- group_type: website-builder-with-images
- business: PrecisionBot Systems (robotics company)

Save the IDs shown in output:
- Correlation ID: <save-this>
- Orchestration ID: <save-this>

─────────────────────────────────────────────────────────────────────────────

Step 3b: Monitor orchestrator (Terminal 1)
─────────────────────────────────────
ORCH_ID="<your-orchestration-id>"
kubectl logs -f deployment/agent-chassis -n agent-system | grep "$ORCH_ID"

Expected flow:
→ Creating orchestration state
→ Starting workflow: website-builder-with-images
→ Executing step: spawn_hero_writer
→ Spawning agent: content-creator-hero
→ Executing step: generate_hero
→ Executing step: generate_hero_image
→ Workflow complete

─────────────────────────────────────────────────────────────────────────────

Step 3c: Monitor image adapter (Terminal 2)
─────────────────────────────────────
CORR_ID="<your-correlation-id>"
kubectl logs -f -n agent-system -l app=image-generator-adapter | grep "$CORR_ID"

Expected:
→ Received request
→ Calling API...
→ Uploading to S3...
→ Response sent

─────────────────────────────────────────────────────────────────────────────

Step 3d: Watch spawned agents (Terminal 3)
─────────────────────────────────────
watch kubectl get pods -n agent-system | grep hero

Should see:
content-creator-hero-<job-id>   Running

─────────────────────────────────────────────────────────────────────────────

Step 3e: Check final database state
─────────────────────────────────────
ORCH_ID="<your-orchestration-id>"

psql -h localhost -p 5432 -U clients_user -d clients_db -c \
"SELECT
orchestration_id,
status,
current_step,
collected_data->'generate_hero_image'->'image_uri' as image_uri
FROM orchestration_states
WHERE orchestration_id = '$ORCH_ID';"

Expected:
status        | complete
current_step  | complete
image_uri     | "s3://bucket/images/..."

─────────────────────────────────────────────────────────────────────────────

Step 3f: Check complete collected data
─────────────────────────────────────
psql -h localhost -p 5432 -U clients_user -d clients_db -c \
"SELECT jsonb_pretty(collected_data)
FROM orchestration_states
WHERE orchestration_id = '$ORCH_ID';"

Should show:
- generate_hero: { result: "text content..." }
- generate_hero_image: { image_uri: "s3://..." }
- spawn_hero_writer: { agent_id, topics, ... }

═══════════════════════════════════════════════════════════════════════════

SUCCESS CHECKLIST
═══════════════════════════════════════════════════════════════════════════

Phase 1: Image Adapter
☐ Adapter receives messages
☐ API calls succeed
☐ S3 uploads work
☐ Responses sent to Kafka

Phase 2: Database Setup
☐ Agent group created
☐ Group type: website-builder-with-images
☐ Agent configs added
☐ 4 workflow steps defined

Phase 3: End-to-End Flow
☐ Orchestrator receives client message
☐ Workflow starts
☐ Hero writer spawns
☐ Hero content generated
☐ Hero image generated
☐ Image URI stored in state
☐ Workflow completes
☐ Status = 'complete'

═══════════════════════════════════════════════════════════════════════════

TROUBLESHOOTING QUICK REFERENCE
═══════════════════════════════════════════════════════════════════════════

Image adapter not receiving messages:
→ Check: kubectl get pods -n agent-system -l app=image-generator-adapter
→ Fix: kubectl rollout restart deployment/image-generator-adapter -n agent-system

S3 upload fails:
→ Check: kubectl get secret s3-credentials -n agent-system
→ Test: kubectl exec -it <pod> -- aws s3 ls

API calls fail:
→ Check: kubectl get secret image-api-credentials -n agent-system
→ Verify key is valid

Workflow doesn't start:
→ Verify: group_type exactly matches 'website-builder-with-images'
→ Check: kubectl logs deployment/agent-chassis -n agent-system | grep ERROR

Workflow stuck:
→ Query: SELECT status, current_step, awaited_steps FROM orchestration_states
→ Check: processing_history for errors

Hero writer doesn't spawn:
→ Check: kubectl get events -n agent-system
→ Verify: Agent definition exists in database

═══════════════════════════════════════════════════════════════════════════

WHAT TO DO IF THINGS GO WRONG
═══════════════════════════════════════════════════════════════════════════

1. Start with Phase 1 - test adapter in isolation
2. Only move to Phase 2 when adapter works
3. Only move to Phase 3 when database is configured
4. Check logs at each step
5. Use correlation IDs to track messages
6. Query database to see orchestration state
7. Don't skip steps!

═══════════════════════════════════════════════════════════════════════════

KEY FILES:
• README_ROBOT_HANDS.md - Detailed explanation
• debug_image_adapter_step_by_step.sh - Interactive debugger
• test_robot_hands_image.sh - Simple adapter test
• test_workflow_with_images.sh - Full workflow test
• create_image_workflow_group.sql - Database setup

═══════════════════════════════════════════════════════════════════════════
EOF



---

# Robot Hands Website Builder - Step-by-Step Integration

## Overview

This package creates a new workflow that generates websites with AI-generated content and images. The theme is precision robotics and robot hands.

**Key Difference from Original**: This creates a completely NEW agent group (`website-builder-with-images`) separate from your existing `multi-section-website-builder` group.

## Files Included

1. **create_image_workflow_group.sql** - Creates new agent group with minimal workflow
2. **debug_image_adapter_step_by_step.sh** - Interactive debugging guide
3. **test_robot_hands_image.sh** - Simple standalone image adapter test
4. **test_workflow_with_images.sh** - Full workflow test with orchestrator

## Integration Strategy

We'll proceed in phases:

### Phase 1: Test Image Adapter in Isolation
Get the image adapter working and verified before integrating it into workflows.

### Phase 2: Create Minimal Workflow
Start with just hero text + hero image (simplest possible workflow).

### Phase 3: Verify End-to-End
Test the complete flow from client request to final result.

### Phase 4: Expand (Future)
Add more sections (features, testimonials, etc.) once the basic pattern works.

---

## Phase 1: Test Image Adapter in Isolation

### Prerequisites

Before starting, you need:

1. **Image adapter deployed**
    - Deployment: `image-generator-adapter`
    - Namespace: `agent-system`
    - At least 1 running pod

2. **Secrets configured**
    - `s3-credentials` (endpoint, access-key, secret-key)
    - `image-api-credentials` (api-key)

3. **Kafka cluster running**
    - Namespace: `kafka`
    - Accessible from agent-system namespace

### Step 1.1: Run Interactive Debugger

The debugging script will check everything step-by-step:

```bash
chmod +x debug_image_adapter_step_by_step.sh
./debug_image_adapter_step_by_step.sh
```

This script will:
1. ✓ Check if adapter deployment exists
2. ✓ Check if pods are running
3. ✓ Verify S3 credentials secret
4. ✓ Verify image API credentials secret
5. ✓ Check Kafka topics exist
6. ✓ Review adapter logs
7. ✓ Send test message
8. ✓ Monitor message processing
9. ✓ Verify S3 upload

**What to expect**: The script will pause at each step and ask you to verify before continuing. If any step fails, it will tell you exactly what to fix.

### Step 1.2: Run Simple Image Test

Once the debugger passes, run a standalone test:

```bash
chmod +x test_robot_hands_image.sh
./test_robot_hands_image.sh
```

This sends ONE message directly to the image adapter requesting a robot hands image.

**Watch adapter logs in another terminal:**
```bash
kubectl logs -f -n agent-system -l app=image-generator-adapter
```

**What to look for:**
```
Received request on system.agent.image-generator.requests
Correlation ID: <uuid>
Calling external image API...
API returned 200 OK
Uploading to S3...
S3 upload successful: s3://bucket/images/2025-10-27/...
Sending response to system.agent.image-generator.responses
```

**Verify the response:**
```bash
kubectl exec -it kafka-0 -n kafka -- kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic system.agent.image-generator.responses \
  --from-beginning \
  --property print.headers=true \
  --max-messages 10
```

Look for your correlation ID and a response like:
```json
{
  "success": true,
  "body": {
    "data": {
      "image_uri": "s3://bucket/images/2025-10-27/abc-123/image.png",
      "width": 512,
      "height": 512,
      "format": "png"
    }
  }
}
```

**Check S3:**
```bash
aws s3 ls s3://your-bucket/images/ --recursive --human-readable
```

You should see a newly uploaded image file.

### Phase 1 Success Criteria

✓ Adapter receives messages from Kafka  
✓ External API call succeeds  
✓ Image uploads to S3  
✓ Response message sent to Kafka  
✓ Response contains valid S3 URI

**DO NOT proceed to Phase 2 until all these work.**

---

## Phase 2: Create Minimal Workflow

Once the adapter works, create the workflow in the database.

### Step 2.1: Review the SQL

Open `create_image_workflow_group.sql` and review it. The workflow is intentionally minimal:

```
spawn_hero_writer → generate_hero → generate_hero_image → complete
```

This is the simplest possible workflow with image generation.

### Step 2.2: Run Step 1 of SQL

Connect to your database:

```bash
kubectl port-forward svc/postgres-clients 5432:5432 -n ai-persona-system
```

In another terminal, run just Step 1:

```bash
psql -h localhost -p 5432 -U clients_user -d clients_db \
  -c "$(sed -n '/^-- Step 1:/,/^-- Step 2:/p' create_image_workflow_group.sql | head -n -1)"
```

This creates the agent group and returns a UUID.

**Save this UUID!** You'll need it for Step 2.

Example output:
```
                  id                  |           name           |       group_type        | version
--------------------------------------+--------------------------+-------------------------+---------
 abc12345-6789-0123-4567-890abcdef123 | Website Builder with Images | website-builder-with-images |       1
```

### Step 2.3: Run Step 2 of SQL

Edit the SQL file and replace `<NEW_GROUP_ID>` with your UUID from Step 1.

Uncomment the Step 2 section and run it:

```sql
INSERT INTO agent_group_configs (agent_group_id, agent_type, functional_role, agent_config)
VALUES 
    ('YOUR-UUID-HERE', 'content-creator-hero', 'hero_writer', 
     jsonb_build_object(
        'processing_mode', 'task',
        'ai_service', jsonb_build_object(
            'provider', 'anthropic',
            'model', 'claude-3-5-sonnet-20241022'
        )
     )),
    
    ('YOUR-UUID-HERE', 'image-generator', 'image_creator', 
     jsonb_build_object(
        'max_concurrent_requests', 3,
        'request_timeout', 60,
        'image_settings', jsonb_build_object(
            'default_width', 1920,
            'default_height', 1080,
            'default_format', 'png'
        )
     ));
```

### Step 2.4: Verify Workflow

Run the verification queries from Step 3:

```bash
psql -h localhost -p 5432 -U clients_user -d clients_db -c \
  "SELECT id, name, group_type, version, is_active 
   FROM agent_groups 
   WHERE group_type = 'website-builder-with-images';"
```

Should show:
```
id      | name                          | group_type                   | version | is_active
--------+-------------------------------+------------------------------+---------+-----------
abc123  | Website Builder with Images   | website-builder-with-images  |       1 | t
```

Check the workflow steps:

```bash
psql -h localhost -p 5432 -U clients_user -d clients_db -c \
  "SELECT jsonb_pretty(workflow_config->'steps') 
   FROM agent_groups 
   WHERE group_type = 'website-builder-with-images';"
```

Should show:
```json
{
  "spawn_hero_writer": { ... },
  "generate_hero": { ... },
  "generate_hero_image": {
    "action": "generate_image",
    ...
  },
  "complete": { ... }
}
```

### Phase 2 Success Criteria

✓ Agent group created in database  
✓ Group type is `website-builder-with-images`  
✓ Workflow has 4 steps  
✓ Agent configs include `content-creator-hero` and `image-generator`

---

## Phase 3: Verify End-to-End

Now test the complete workflow from client request to final result.

### Step 3.1: Send Test Message

Run the full workflow test:

```bash
chmod +x test_workflow_with_images.sh
./test_workflow_with_images.sh
```

This sends a message to the generic orchestrator with:
- `group_type: "website-builder-with-images"`
- `business_name: "PrecisionBot Systems"`
- `business_type: "robotics automation company"`

### Step 3.2: Monitor Orchestrator

In terminal 1, watch the orchestrator:

```bash
ORCH_ID="<your-orchestration-id-from-test-output>"
kubectl logs -f deployment/agent-chassis -n agent-system | grep "$ORCH_ID"
```

**What to look for:**

```
Creating orchestration state for: <ORCH_ID>
Starting workflow: website-builder-with-images
Executing step: spawn_hero_writer
Spawning agent: content-creator-hero
Current step: generate_hero
Executing step: generate_hero
Calling agent: content-creator-hero (role: hero_writer)
Current step: generate_hero_image
Executing step: generate_hero_image
Calling action: generate_image
Current step: complete
Workflow complete
```

### Step 3.3: Monitor Image Adapter

In terminal 2, watch the image adapter:

```bash
CORR_ID="<your-correlation-id-from-test-output>"
kubectl logs -f -n agent-system -l app=image-generator-adapter | grep "$CORR_ID"
```

**What to look for:**

```
Received request: generate_image
Correlation ID: <CORR_ID>
Calling image API...
Uploading to S3: images/2025-10-27/<CORR_ID>/image.png
Sending response...
```

### Step 3.4: Check Spawned Agents

Watch for the hero writer agent to spawn:

```bash
kubectl get pods -n agent-system | grep hero
```

Should show something like:
```
content-creator-hero-<job-id>   1/1   Running   0   15s
```

### Step 3.5: Query Database State

Check the orchestration state:

```bash
psql -h localhost -p 5432 -U clients_user -d clients_db -c \
  "SELECT 
     orchestration_id,
     orchestration_name,
     status,
     current_step,
     jsonb_pretty(collected_data->'generate_hero'->'result') as hero_text,
     collected_data->'generate_hero_image'->'image_uri' as hero_image_uri
   FROM orchestration_states 
   WHERE orchestration_id = '<ORCH_ID>';"
```

**Expected result when complete:**

```
orchestration_id | orchestration_name      | status   | current_step | hero_text                    | hero_image_uri
-----------------+-------------------------+----------+--------------+------------------------------+------------------------
abc-123          | robot-website-162045    | complete | complete     | "Revolutionary Precision..." | "s3://bucket/images/..."
```

### Step 3.6: Verify Final Response

Check the response topic:

```bash
kubectl exec -it kafka-0 -n kafka -- kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic system.responses.generic \
  --property print.headers=true \
  --from-beginning | grep -A 30 "$CORR_ID"
```

The response should contain:
- `success: true`
- `generate_hero` with text content
- `generate_hero_image` with S3 URI

### Phase 3 Success Criteria

✓ Orchestrator receives message and starts workflow  
✓ Hero writer agent spawns successfully  
✓ Hero content is generated  
✓ Image generation request is sent  
✓ Image adapter generates and uploads image  
✓ Image URI is stored in orchestration state  
✓ Workflow completes with status='complete'  
✓ Final response contains both text and image

---

## Troubleshooting Guide

### Issue: Adapter doesn't receive messages

**Symptoms:**
- Test message sent but adapter logs show nothing
- No activity in adapter logs

**Debug:**
```bash
# 1. Check adapter is running
kubectl get pods -n agent-system -l app=image-generator-adapter

# 2. Check adapter is consuming from topic
kubectl logs -n agent-system -l app=image-generator-adapter | grep "Subscribed to"

# 3. Check topic exists
kubectl exec -it kafka-0 -n kafka -- kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --list | grep image-generator

# 4. Check consumer group
kubectl exec -it kafka-0 -n kafka -- kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --list | grep image
```

**Fixes:**
- Restart adapter deployment
- Check KAFKA_BROKERS environment variable
- Verify network policies allow Kafka access

### Issue: External API calls fail

**Symptoms:**
- Adapter receives message but logs show API errors
- HTTP 401, 403, or 500 errors

**Debug:**
```bash
# Check API secret exists
kubectl get secret image-api-credentials -n agent-system -o yaml

# Check API key is mounted correctly
kubectl exec -it <adapter-pod> -n agent-system -- env | grep API

# Test API manually
kubectl exec -it <adapter-pod> -n agent-system -- \
  curl -H "Authorization: Bearer $API_KEY" \
  https://api.stability.ai/v1/engines/list
```

**Fixes:**
- Verify API key is valid
- Check API rate limits
- Verify network policies allow external egress

### Issue: S3 upload fails

**Symptoms:**
- Image generation succeeds but S3 upload fails
- Logs show S3 permission errors

**Debug:**
```bash
# Check S3 secret
kubectl get secret s3-credentials -n agent-system -o yaml

# Check environment variables
kubectl exec -it <adapter-pod> -n agent-system -- env | grep S3

# Test S3 access
kubectl exec -it <adapter-pod> -n agent-system -- \
  aws s3 ls s3://your-bucket/
```

**Fixes:**
- Verify S3 credentials are correct
- Check bucket exists and is accessible
- Verify IAM permissions for PutObject

### Issue: Workflow gets stuck at image step

**Symptoms:**
- Orchestration state shows `current_step: generate_hero_image`
- Workflow doesn't progress to complete

**Debug:**
```bash
# Check awaited steps
psql ... -c "SELECT awaited_steps FROM orchestration_states WHERE orchestration_id = '...';"

# Check if response was sent
kubectl exec -it kafka-0 -n kafka -- kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic system.agent.generic.responses \
  --from-beginning | grep "<orchestration-id>"

# Check processing history
psql ... -c "SELECT jsonb_pretty(processing_history) FROM orchestration_states WHERE orchestration_id = '...';"
```

**Common causes:**
- Response topic mismatch
- await_response not set correctly in workflow
- Image adapter not sending response

### Issue: Hero writer doesn't spawn

**Symptoms:**
- Workflow starts but no hero writer pod appears
- Orchestration stuck at `spawn_hero_writer`

**Debug:**
```bash
# Check orchestrator logs for spawn errors
kubectl logs deployment/agent-chassis -n agent-system | grep -i spawn

# Check Kubernetes events
kubectl get events -n agent-system --sort-by='.lastTimestamp'

# Check agent definition exists
psql ... -c "SELECT * FROM agent_types WHERE type = 'content-creator-hero';"
```

**Fixes:**
- Verify agent definition in database
- Check Kubernetes RBAC for job creation
- Check resource quotas

---

## Next Steps

### Add More Sections

Once the basic workflow works, add features and other sections:

```sql
-- Add feature writer to workflow
UPDATE agent_groups
SET workflow_config = jsonb_set(
    workflow_config,
    '{steps,spawn_features_writer}',
    jsonb_build_object(
        'action', 'spawn_agent',
        'description', 'Spawn features writer',
        'next_step', 'generate_features',
        'config', jsonb_build_object(
            'agent_type', 'content-creator-features',
            'role', 'features_writer'
        )
    )
)
WHERE group_type = 'website-builder-with-images';

-- Link hero image to features spawn
UPDATE agent_groups
SET workflow_config = jsonb_set(
    workflow_config,
    '{steps,generate_hero_image,next_step}',
    '"spawn_features_writer"'
)
WHERE group_type = 'website-builder-with-images';
```

### Parallel Image Generation

Generate multiple images concurrently:

```sql
-- Change next_step to next_steps (array)
UPDATE agent_groups
SET workflow_config = jsonb_set(
    workflow_config,
    '{steps,generate_hero,next_steps}',
    '["generate_hero_image", "generate_features_image"]'::jsonb
)
WHERE group_type = 'website-builder-with-images';
```

### Dynamic Image Prompts

Use generated content in image prompts (requires template processing):

```sql
-- Reference content in image prompt
UPDATE agent_groups
SET workflow_config = jsonb_set(
    workflow_config,
    '{steps,generate_hero_image,config,input_mapping,prompt}',
    '"Based on this hero content: {{.generate_hero.result}} - Create a matching professional image"'
)
WHERE group_type = 'website-builder-with-images';
```

---

## Summary

This integration adds image generation in three phases:

**Phase 1**: Verify image adapter works in isolation
- Direct message to adapter
- Successful API call and S3 upload
- Response message sent

**Phase 2**: Create minimal workflow in database
- New agent group: `website-builder-with-images`
- Simple flow: spawn → generate text → generate image → complete

**Phase 3**: Test complete orchestration
- Client message triggers workflow
- Agents spawn dynamically
- Content and images generated
- Final result includes both

The workflow is minimal by design - get the pattern working first, then expand.


---

#!/bin/bash

# ============================================================================
# Step-by-Step Image Adapter Debugging Guide
# ============================================================================
# This script walks through testing the image adapter from scratch.
# Run each step and verify it works before moving to the next.
# ============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
step() {
echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}STEP $1: $2${NC}"
echo -e "${BLUE}=========================================${NC}"
}

success() {
echo -e "${GREEN}✓ $1${NC}"
}

error() {
echo -e "${RED}✗ $1${NC}"
}

warning() {
echo -e "${YELLOW}⚠ $1${NC}"
}

info() {
echo -e "  $1"
}

wait_for_user() {
echo ""
read -p "Press Enter to continue to next step..."
}

# ============================================================================
# STEP 1: Check Image Adapter Deployment
# ============================================================================
step "1" "Check Image Adapter Deployment"

info "Checking if image-generator-adapter deployment exists..."
if kubectl get deployment image-generator-adapter -n agent-system &> /dev/null; then
success "Deployment exists"

    # Check if pods are running
    RUNNING_PODS=$(kubectl get pods -n agent-system -l app=image-generator-adapter --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l)
    if [ "$RUNNING_PODS" -gt 0 ]; then
        success "Found $RUNNING_PODS running pod(s)"
        kubectl get pods -n agent-system -l app=image-generator-adapter
    else
        error "No running pods found"
        kubectl get pods -n agent-system -l app=image-generator-adapter
        echo ""
        warning "Image adapter is not running. Deploy it before continuing."
        exit 1
    fi
else
error "Image-generator-adapter deployment not found"
echo ""
warning "You need to deploy the image adapter first."
warning "Check your deployment configuration and apply it:"
info "  kubectl apply -f <image-adapter-deployment.yaml> -n agent-system"
exit 1
fi

wait_for_user

# ============================================================================
# STEP 2: Check Secrets Configuration
# ============================================================================
step "2" "Check Secrets Configuration"

info "Checking S3 credentials secret..."
if kubectl get secret s3-credentials -n agent-system &> /dev/null; then
success "S3 credentials secret exists"
kubectl get secret s3-credentials -n agent-system
else
error "S3 credentials secret not found"
echo ""
warning "Create the secret with:"
info "  kubectl create secret generic s3-credentials \\"
info "    --from-literal=endpoint=https://s3.amazonaws.com \\"
info "    --from-literal=access-key=YOUR_ACCESS_KEY \\"
info "    --from-literal=secret-key=YOUR_SECRET_KEY \\"
info "    -n agent-system"
exit 1
fi

echo ""
info "Checking image API credentials secret..."
if kubectl get secret image-api-credentials -n agent-system &> /dev/null; then
success "Image API credentials secret exists"
kubectl get secret image-api-credentials -n agent-system
else
error "Image API credentials secret not found"
echo ""
warning "Create the secret with:"
info "  kubectl create secret generic image-api-credentials \\"
info "    --from-literal=api-key=YOUR_STABILITY_API_KEY \\"
info "    -n agent-system"
exit 1
fi

wait_for_user

# ============================================================================
# STEP 3: Check Kafka Topics
# ============================================================================
step "3" "Check Kafka Topics"

info "Looking for image-generator topics..."
kubectl exec -it kafka-0 -n kafka -- kafka-topics.sh \
--bootstrap-server localhost:9092 \
--list 2>/dev/null | grep image || true

echo ""
info "Expected topics:"
info "  - system.agent.image-generator.requests"
info "  - system.agent.image-generator.responses"
echo ""

read -p "Do the expected topics exist? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
warning "Topics may not exist yet. They should be created automatically."
warning "If they don't exist, check your topic manager logs:"
info "  kubectl logs -n agent-system -l app=topic-manager"
fi

wait_for_user

# ============================================================================
# STEP 4: Check Adapter Logs (Current State)
# ============================================================================
step "4" "Check Adapter Logs (Current State)"

info "Showing last 20 lines of adapter logs..."
echo ""
kubectl logs -n agent-system -l app=image-generator-adapter --tail=20

echo ""
read -p "Do the logs look healthy? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
error "Adapter logs show issues. Review them before continuing:"
info "  kubectl logs -f -n agent-system -l app=image-generator-adapter"
exit 1
fi

wait_for_user

# ============================================================================
# STEP 5: Send Simple Test Message
# ============================================================================
step "5" "Send Simple Test Message"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)

info "Test message details:"
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Request ID:       $REQUEST_ID"
echo "  Message ID:       $MESSAGE_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo ""
info "Prompt: Robot hands assembling components"
echo ""

read -p "Send test message? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
warning "Skipping test message"
exit 0
fi

info "Sending message to system.agent.image-generator.requests..."
kubectl -n kafka run -i --rm kcat-producer-$$-$(date +%s) \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.image-generator.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H message_type=request \
-H action=generate_image \
-H from_agent_type=test \
-H from_agent_id=test-client-id \
-H responses_topic=system.agent.image-generator.responses <<EOF
{
"action": "generate_image",
"data": {
"prompt": "Professional photograph of precision robotic hands assembling electronic components, modern factory setting, photorealistic",
"width": 512,
"height": 512
}
}
EOF

success "Message sent!"
echo ""
info "Correlation ID for tracking: $CORRELATION_ID"

wait_for_user

# ============================================================================
# STEP 6: Monitor Adapter Logs
# ============================================================================
step "6" "Monitor Adapter Logs"

info "Watching adapter logs for the test message..."
info "Look for log entries containing correlation ID: $CORRELATION_ID"
info ""
info "Press Ctrl+C when you see the image generation complete"
info "Expected flow:"
info "  1. Received request"
info "  2. Calling external API"
info "  3. Uploading to S3"
info "  4. Sending response"
echo ""

sleep 2
kubectl logs -f -n agent-system -l app=image-generator-adapter --tail=50 | grep --line-buffered -E "$CORRELATION_ID|generate_image|s3|upload|response" || true

echo ""
read -p "Did you see the complete flow? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
error "Adapter did not process the message correctly"
echo ""
warning "Check full logs for errors:"
info "  kubectl logs -n agent-system -l app=image-generator-adapter | grep $CORRELATION_ID"
exit 1
fi

wait_for_user

# ============================================================================
# STEP 7: Check Response Topic
# ============================================================================
step "7" "Check Response Topic"

info "Checking for response message on system.agent.image-generator.responses..."
info "Looking for correlation_id: $CORRELATION_ID"
echo ""

timeout 10 kubectl exec -it kafka-0 -n kafka -- kafka-console-consumer.sh \
--bootstrap-server localhost:9092 \
--topic system.agent.image-generator.responses \
--from-beginning \
--property print.headers=true \
--property print.timestamp=true \
--max-messages 100 2>/dev/null | grep -A 5 "$CORRELATION_ID" || warning "No response found in first 100 messages"

echo ""
read -p "Did you see a response with your correlation ID? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
error "No response message found"
warning "The adapter may not be sending responses. Check:"
info "  1. Adapter logs for errors"
info "  2. Response topic exists"
info "  3. Kafka producer configuration in adapter"
exit 1
fi

wait_for_user

# ============================================================================
# STEP 8: Verify S3 Upload
# ============================================================================
step "8" "Verify S3 Upload"

info "Check your S3 bucket for the uploaded image."
echo ""
warning "This step requires AWS CLI configured with your credentials."
echo ""
read -p "Enter your S3 bucket name (or press Enter to skip): " BUCKET_NAME

if [ -n "$BUCKET_NAME" ]; then
info "Listing recent objects in s3://$BUCKET_NAME/images/..."
aws s3 ls "s3://$BUCKET_NAME/images/" --recursive --human-readable | tail -20 || warning "Could not list S3 objects"

    echo ""
    info "Look for an image uploaded around: $(date '+%Y-%m-%d %H:%M')"
    echo ""
    read -p "Did you find the uploaded image? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        error "Image not found in S3"
        warning "Check adapter logs for S3 upload errors"
        exit 1
    fi
else
warning "Skipped S3 verification"
fi

wait_for_user

# ============================================================================
# STEP 9: Summary
# ============================================================================
step "9" "Summary"

success "Image adapter test completed successfully!"
echo ""
info "Verified:"
info "  ✓ Adapter deployment is running"
info "  ✓ Secrets are configured"
info "  ✓ Kafka topics exist"
info "  ✓ Adapter receives messages"
info "  ✓ External API integration works"
info "  ✓ S3 upload functions"
info "  ✓ Response messages are sent"
echo ""
success "The image adapter is ready to use in workflows!"
echo ""
info "Next steps:"
info "  1. Run the SQL to create the website-builder-with-images group"
info "  2. Test the full workflow with: ./test_workflow_with_images.sh"
echo ""
info "Your test correlation ID: $CORRELATION_ID"
info "Save this if you need to debug further"

--

#!/bin/bash

# ============================================================================
# Simple Image Adapter Test - Robot Hands
# ============================================================================
# Sends a single message directly to the image-generator adapter
# Theme: Precision robotic hands
# ============================================================================

set -e

# Generate unique IDs
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="robot-hands-test-$(date +%H%M%S)"

echo "========================================="
echo "Image Adapter Test - Robot Hands"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Request ID:       $REQUEST_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Name:             $ORCHESTRATION_NAME"
echo "  Time:             $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="
echo ""
echo "Test Details:"
echo "  Topic:      system.agent.image-generator.requests"
echo "  Prompt:     Robot hands assembling components"
echo "  Size:       512x512 (small, fast for testing)"
echo "========================================="
echo ""

# Send message
kubectl -n kafka run -i --rm kcat-producer-$(date +%s) \
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
-H from_agent_id=test-client \
-H responses_topic=system.agent.image-generator.responses <<EOF
{
"action": "generate_image",
"data": {
"prompt": "Professional photograph of precision robotic hands assembling electronic components, modern factory setting, high-tech industrial photography, dramatic lighting, photorealistic, 8k quality",
"width": 512,
"height": 512,
"seed": 42
}
}
EOF

echo ""
echo "========================================="
echo "Message sent!"
echo "========================================="
echo ""
echo "IMPORTANT: Save this correlation ID for tracking:"
echo "  $CORRELATION_ID"
echo ""
echo "========================================="
echo "Monitor Commands:"
echo "========================================="
echo ""
echo "1. Watch adapter logs (real-time):"
echo "   kubectl logs -f -n agent-system -l app=image-generator-adapter"
echo ""
echo "2. Search logs for this specific request:"
echo "   kubectl logs -n agent-system -l app=image-generator-adapter | grep '$CORRELATION_ID'"
echo ""
echo "3. Watch response topic:"
cat <<'CONSUMER'
kubectl exec -it kafka-0 -n kafka -- kafka-console-consumer.sh \
--bootstrap-server localhost:9092 \
--topic system.agent.image-generator.responses \
--property print.headers=true \
--property print.timestamp=true \
--from-beginning | grep -A 10 'CORRELATION_ID'
CONSUMER
echo ""
echo "   (Replace CORRELATION_ID with: $CORRELATION_ID)"
echo ""
echo "4. Check S3 bucket (if AWS CLI configured):"
echo "   aws s3 ls s3://YOUR-BUCKET/images/ --recursive --human-readable | tail -20"
echo ""
echo "========================================="
echo "Expected Flow:"
echo "========================================="
echo ""
echo "In adapter logs, you should see:"
echo "  1. 'Received request' with correlation_id"
echo "  2. 'Calling external image API'"
echo "  3. 'Uploading to S3'"
echo "  4. 'Sending response'"
echo ""
echo "In response topic, you should see:"
echo "  - correlation_id: $CORRELATION_ID"
echo "  - success: true"
echo "  - image_uri: s3://bucket/images/YYYY-MM-DD/..."
echo ""
echo "========================================="
echo "Troubleshooting:"
echo "========================================="
echo ""
echo "If nothing happens:"
echo "  • Check adapter pod is running:"
echo "    kubectl get pods -n agent-system -l app=image-generator-adapter"
echo ""
echo "  • Check adapter can reach Kafka:"
echo "    kubectl logs -n agent-system -l app=image-generator-adapter | grep -i kafka"
echo ""
echo "If adapter receives but fails:"
echo "  • Check S3 credentials:"
echo "    kubectl get secret s3-credentials -n agent-system"
echo ""
echo "  • Check API credentials:"
echo "    kubectl get secret image-api-credentials -n agent-system"
echo ""
echo "  • Check full error in logs:"
echo "    kubectl logs -n agent-system -l app=image-generator-adapter | grep -i error"
echo ""
echo "========================================="

-- ============================================================================
-- Create NEW Agent Group: Website Builder with Images
-- ============================================================================
-- This creates a completely separate group from multi-section-website-builder
-- Theme: Robot hands / robotics company
-- ============================================================================

-- Step 1: Create the new agent group
INSERT INTO agent_groups (
id,
name,
group_type,
description,
version,
workflow_config,
is_active,
created_at,
updated_at
) VALUES (
gen_random_uuid(),
'Website Builder with Images',
'website-builder-with-images',
'Website builder that generates content and images for robotics companies',
1,
jsonb_build_object(
'start_step', 'spawn_hero_writer',
'steps', jsonb_build_object(
-- Spawn hero writer agent
'spawn_hero_writer', jsonb_build_object(
'action', 'spawn_agent',
'description', 'Spawn hero content writer',
'next_step', 'generate_hero',
'config', jsonb_build_object(
'agent_type', 'content-creator-hero',
'role', 'hero_writer'
)
),
-- Generate hero content
'generate_hero', jsonb_build_object(
'action', 'call_agent',
'description', 'Generate hero section content',
'next_step', 'generate_hero_image',
'config', jsonb_build_object(
'agent_type', 'content-creator-hero',
'target_role', 'hero_writer',
'prompt', 'Write a compelling hero section for {{.business_name}}, a {{.business_type}}. Include a powerful headline and engaging subheadline that highlights precision robotics and automation.'
)
),
-- Generate hero image
'generate_hero_image', jsonb_build_object(
'action', 'generate_image',
'description', 'Generate hero section image showing robot hands',
'next_step', 'complete',
'config', jsonb_build_object(
'input_mapping', jsonb_build_object(
'prompt', 'Professional photograph of precision robotic hands assembling electronic components, modern factory setting, high-tech industrial photography, dramatic lighting, photorealistic, 8k quality',
'width', 1920,
'height', 1080
),
'await_response', true
)
),
-- Complete workflow
'complete', jsonb_build_object(
'action', 'complete_workflow',
'description', 'Return complete website content with hero text and image'
)
)
),
true,
now(),
now()
) RETURNING id, name, group_type, version;

-- IMPORTANT: Copy the UUID returned above and use it in the next statements
-- Replace '<NEW_GROUP_ID>' with the actual UUID

-- ============================================================================
-- Step 2: Add agent configurations to the new group
-- ============================================================================
-- Replace '<NEW_GROUP_ID>' with the UUID from Step 1

/*
INSERT INTO agent_group_configs (agent_group_id, agent_type, functional_role, agent_config)
VALUES
-- Hero content writer
('<NEW_GROUP_ID>', 'content-creator-hero', 'hero_writer',
jsonb_build_object(
'processing_mode', 'task',
'ai_service', jsonb_build_object(
'provider', 'anthropic',
'model', 'claude-3-5-sonnet-20241022'
)
)),

    -- Image generator adapter
    ('<NEW_GROUP_ID>', 'image-generator', 'image_creator', 
     jsonb_build_object(
        'max_concurrent_requests', 3,
        'request_timeout', 60,
        'image_settings', jsonb_build_object(
            'default_width', 1920,
            'default_height', 1080,
            'default_format', 'png'
        )
     ));
*/

-- ============================================================================
-- Step 3: Verification Queries
-- ============================================================================

-- Show all website builder groups
SELECT
id,
name,
group_type,
version,
is_active,
created_at
FROM agent_groups
WHERE group_type LIKE '%website%'
ORDER BY created_at DESC;

-- Show the new group's workflow
SELECT
g.id,
g.name,
g.group_type,
g.version,
jsonb_pretty(g.workflow_config) as workflow_detail
FROM agent_groups g
WHERE g.group_type = 'website-builder-with-images';

-- Show agent configs for the new group (after Step 2 is completed)
/*
SELECT
g.name as group_name,
g.group_type,
c.agent_type,
c.functional_role,
jsonb_pretty(c.agent_config) as config
FROM agent_groups g
LEFT JOIN agent_group_configs c ON g.id = c.agent_group_id
WHERE g.group_type = 'website-builder-with-images'
ORDER BY c.agent_type;
*/

-- ============================================================================
-- USAGE NOTES
-- ============================================================================
-- After running Step 1, you'll get a UUID. Use that UUID to:
-- 1. Uncomment and run Step 2 (replace <NEW_GROUP_ID>)
-- 2. Test the workflow with the test message script
--
-- The workflow is intentionally minimal (just hero + image) to test
-- image generation first before expanding to more sections.
-- ============================================================================


---


#!/bin/bash

# ============================================================================
# Full Workflow Test - Robot Hands Website with Images
# ============================================================================
# Sends a message to the generic orchestrator to trigger the
# website-builder-with-images workflow
# ============================================================================

set -e

# Generate unique IDs
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="robot-website-$(date +%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
STEP_NAME="client_step_website_request"
CLIENT_ID="demo_client"

echo "========================================="
echo "Robot Hands Website Builder Test"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Request ID:       $REQUEST_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Name:             $ORCHESTRATION_NAME"
echo "  Time:             $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="
echo ""
echo "This will create:"
echo "  1. Hero content section (AI-generated text)"
echo "  2. Hero image (robot hands, 1920x1080)"
echo "  3. Complete workflow orchestration"
echo "========================================="
echo ""

# Verify the workflow exists in database
echo "Pre-flight check: Verifying workflow exists in database..."
echo ""
echo "Run this query to verify:"
cat <<SQL
psql -h postgres-clients.ai-persona-system.svc.cluster.local \\
-U clients_user -d clients_db -c \\
"SELECT id, name, group_type, version, is_active
FROM agent_groups
WHERE group_type = 'website-builder-with-images';"
SQL
echo ""
read -p "Have you verified the workflow exists? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
echo ""
echo "Please run the SQL script first:"
echo "  psql ... -f create_image_workflow_group.sql"
exit 1
fi

echo ""
echo "Sending message to generic orchestrator..."
echo ""

# Send the message
kubectl -n kafka run -i --rm kcat-producer-$(date +%s) \
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
{
"action": "orchestrate",
"config": {
"group_type": "website-builder-with-images"
},
"input_data": {
"business_type": "robotics automation company",
"business_name": "PrecisionBot Systems"
}
}
EOF

echo ""
echo "========================================="
echo "Message sent!"
echo "========================================="
echo ""
echo "TRACKING INFO:"
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo ""
echo "========================================="
echo "Expected Workflow Steps:"
echo "========================================="
echo ""
echo "1. spawn_hero_writer"
echo "   → Generic orchestrator spawns content-creator-hero agent"
echo ""
echo "2. generate_hero"
echo "   → Hero writer creates text content about PrecisionBot Systems"
echo ""
echo "3. generate_hero_image"
echo "   → Image adapter generates robot hands image"
echo "   → Uploads to S3"
echo "   → Returns s3:// URI"
echo ""
echo "4. complete"
echo "   → Returns final result with text + image URI"
echo ""
echo "========================================="
echo "Monitoring Commands:"
echo "========================================="
echo ""
echo "1. Watch orchestrator logs (shows workflow execution):"
echo "   kubectl logs -f deployment/agent-chassis -n agent-system | grep '$ORCHESTRATION_ID'"
echo ""
echo "2. Watch image adapter logs (shows image generation):"
echo "   kubectl logs -f -n agent-system -l app=image-generator-adapter | grep '$CORRELATION_ID'"
echo ""
echo "3. Check orchestration state in database:"
cat <<DBQUERY
psql -h postgres-clients.ai-persona-system.svc.cluster.local \\
-U clients_user -d clients_db -c \\
"SELECT
orchestration_id,
orchestration_name,
status,
current_step,
jsonb_pretty(collected_data->'generate_hero') as hero_text,
jsonb_pretty(collected_data->'generate_hero_image') as hero_image
FROM orchestration_states
WHERE orchestration_id = '$ORCHESTRATION_ID';"
DBQUERY
echo ""
echo "4. Watch spawned agents:"
echo "   kubectl get pods -n agent-system -l parent_orchestration_id=$ORCHESTRATION_ID"
echo ""
echo "5. Check response topic for final result:"
cat <<RESPONSE
kubectl exec -it kafka-0 -n kafka -- kafka-console-consumer.sh \\
--bootstrap-server localhost:9092 \\
--topic system.responses.generic \\
--property print.headers=true \\
--from-beginning | grep -A 20 '$CORRELATION_ID'
RESPONSE
echo ""
echo "========================================="
echo "Step-by-Step Verification:"
echo "========================================="
echo ""
echo "Step 1: Wait ~10 seconds for hero_writer to spawn"
echo "  Check: kubectl get pods -n agent-system | grep hero"
echo ""
echo "Step 2: Wait ~20 seconds for hero content generation"
echo "  Check database for generate_hero step completion"
echo ""
echo "Step 3: Wait ~30 seconds for image generation"
echo "  Check image adapter logs and S3 bucket"
echo ""
echo "Step 4: Check final orchestration status"
echo "  Should show status='complete' with both text and image"
echo ""
echo "========================================="
echo "Troubleshooting:"
echo "========================================="
echo ""
echo "If workflow doesn't start:"
echo "  • Verify group exists in database (see query above)"
echo "  • Check generic orchestrator logs for errors"
echo "  • Verify group_type exactly matches: 'website-builder-with-images'"
echo ""
echo "If hero_writer doesn't spawn:"
echo "  • Check orchestrator logs for spawn errors"
echo "  • Verify agent definition exists for content-creator-hero"
echo "  • Check Kubernetes job creation permissions"
echo ""
echo "If image generation fails:"
echo "  • Run the simple image test first: ./test_robot_hands_image.sh"
echo "  • Verify adapter is receiving the request"
echo "  • Check S3 and API credentials"
echo ""
echo "If workflow gets stuck:"
echo "  • Check current_step in database"
echo "  • Check awaited_steps to see what it's waiting for"
echo "  • Review processing_history for error details"
echo ""
echo "========================================="



==
corrected
==
-- ============================================================================
-- Create NEW Agent Group: Website Builder with Images (Robot Hands Theme)
-- ============================================================================
-- This creates a completely separate group from multi-section-website-builder
-- Schema: agent_group_definitions table
-- ============================================================================

-- Step 1: Create the new agent group with correct schema
INSERT INTO agent_group_definitions (
id,
name,
group_type,
agent_configs,
orchestration_workflow,
usage_count,
version,
created_at,
updated_at
) VALUES (
gen_random_uuid(),
'Website Builder with Images',
'website-builder-with-images',
-- agent_configs: Array of agent configurations
jsonb_build_array(
jsonb_build_object(
'role', 'hero_writer',
'agent_type', 'content-creator-hero'
),
jsonb_build_object(
'role', 'image_creator',
'agent_type', 'image-generator'
)
),
-- orchestration_workflow: Complete workflow definition
jsonb_build_object(
'start_step', 'spawn_hero_writer',
'steps', jsonb_build_object(
-- Step 1: Spawn hero writer agent
'spawn_hero_writer', jsonb_build_object(
'action', 'spawn_agent',
'description', 'Spawn agent for hero section',
'next_step', 'generate_hero',
'config', jsonb_build_object(
'role', 'hero_writer',
'agent_type', 'content-creator-hero'
)
),
-- Step 2: Generate hero content
'generate_hero', jsonb_build_object(
'action', 'call_agent',
'description', 'Generate hero section content',
'next_step', 'generate_hero_image',
'config', jsonb_build_object(
'prompt', 'Write a compelling hero section for {{.business_name}}, a {{.business_type}}. Include a powerful headline and engaging subheadline that highlights precision robotics and automation.',
'agent_type', 'content-creator-hero',
'target_role', 'hero_writer'
)
),
-- Step 3: Generate hero image (robot hands)
'generate_hero_image', jsonb_build_object(
'action', 'generate_image',
'description', 'Generate hero section image showing robot hands',
'next_step', 'complete',
'config', jsonb_build_object(
'input_mapping', jsonb_build_object(
'prompt', 'Professional photograph of precision robotic hands assembling electronic components, modern factory setting, high-tech industrial photography, dramatic lighting, photorealistic, 8k quality',
'width', 1920,
'height', 1080
),
'await_response', true
)
),
-- Step 4: Complete workflow
'complete', jsonb_build_object(
'action', 'complete_workflow',
'description', 'Return complete website content with hero text and image'
)
)
),
0,  -- usage_count
1,  -- version
now(),
now()
) RETURNING id, name, group_type, version;

-- ============================================================================
-- Verification Queries
-- ============================================================================

-- Show all website builder groups
SELECT
id,
name,
group_type,
version,
usage_count,
created_at
FROM agent_group_definitions
WHERE group_type LIKE '%website%'
ORDER BY created_at DESC;

-- Show the new group's complete definition
SELECT
id,
name,
group_type,
version,
jsonb_pretty(agent_configs) as agent_configs,
jsonb_pretty(orchestration_workflow) as workflow
FROM agent_group_definitions
WHERE group_type = 'website-builder-with-images';

-- Show just the workflow steps
SELECT
name,
group_type,
jsonb_object_keys(orchestration_workflow->'steps') as step_name
FROM agent_group_definitions
WHERE group_type = 'website-builder-with-images';

-- Show the workflow start step
SELECT
name,
orchestration_workflow->'start_step' as start_step
FROM agent_group_definitions
WHERE group_type = 'website-builder-with-images';

-- ============================================================================
-- USAGE NOTES
-- ============================================================================
-- After running this SQL, verify:
-- 1. New group exists with group_type = 'website-builder-with-images'
-- 2. agent_configs contains 2 entries (hero_writer, image_creator)
-- 3. orchestration_workflow has 4 steps
-- 4. start_step is 'spawn_hero_writer'
--
-- Then test with:
--   ./test_workflow_with_images.sh
--
-- The workflow is intentionally minimal (just hero + image) to test
-- image generation first before expanding to more sections.
-- ============================================================================

-- ============================================================================
-- Optional: Add more sections later
-- ============================================================================
-- Once the basic workflow works, you can expand it by updating the group:

/*
-- Example: Add features section with image
UPDATE agent_group_definitions
SET
-- Add features_writer to agent_configs
agent_configs = agent_configs || jsonb_build_array(
jsonb_build_object(
'role', 'features_writer',
'agent_type', 'content-creator-features'
)
),
-- Add features steps to workflow
orchestration_workflow = jsonb_set(
jsonb_set(
jsonb_set(
orchestration_workflow,
'{steps,spawn_features_writer}',
jsonb_build_object(
'action', 'spawn_agent',
'description', 'Spawn agent for features section',
'next_step', 'generate_features',
'config', jsonb_build_object(
'role', 'features_writer',
'agent_type', 'content-creator-features'
)
)
),
'{steps,generate_features}',
jsonb_build_object(
'action', 'call_agent',
'description', 'Generate features section',
'next_step', 'generate_features_image',
'config', jsonb_build_object(
'prompt', 'List 3-4 key features of {{.business_name}}.',
'agent_type', 'content-creator-features',
'target_role', 'features_writer'
)
)
),
'{steps,generate_features_image}',
jsonb_build_object(
'action', 'generate_image',
'description', 'Generate features section image',
'next_step', 'complete',
'config', jsonb_build_object(
'input_mapping', jsonb_build_object(
'prompt', 'Close-up of robotic hands working on precision assembly',
'width', 800,
'height', 600
),
'await_response', true
)
)
),
-- Update hero image to point to features spawn
orchestration_workflow = jsonb_set(
orchestration_workflow,
'{steps,generate_hero_image,next_step}',
'"spawn_features_writer"'
),
version = version + 1,
updated_at = now()
WHERE group_type = 'website-builder-with-images';
*/


---

improved:
UPDATE agent_group_definitions
SET orchestration_workflow = '{
"steps": {
"spawn_hero_writer": {
"action": "spawn_agent",
"config": {
"role": "hero_writer",
"agent_type": "content-creator-hero"
},
"description": "Spawn hero content writer",
"next_step": "spawn_image_creator"
},
"spawn_image_creator": {
"action": "spawn_agent",
"config": {
"role": "image_creator",
"agent_type": "image-generator"
},
"description": "Spawn image generator",
"next_step": "generate_hero"
},
"generate_hero": {
"action": "call_agent",
"config": {
"prompt": "Write a compelling hero section for {{.business_name}}, a {{.business_type}}. Focus on precision robotics and automation.",
"agent_type": "content-creator-hero",
"target_role": "hero_writer"
},
"next_step": "generate_hero_image",
"description": "Generate hero content"
},
"generate_hero_image": {
"action": "call_agent",
"config": {
"width": 1920,
"height": 1080,
"prompt": "Professional photograph of precision robotic hands assembling electronic components, modern factory setting, dramatic lighting, photorealistic, 8k",
"agent_type": "image-generator",
"target_role": "image_creator"
},
"next_step": "complete",
"description": "Generate hero image"
},
"complete": {
"action": "complete_workflow",
"description": "Complete workflow"
}
},
"start_step": "spawn_hero_writer"
}'::jsonb
WHERE group_type = 'robot-hands-website';


-- Fix the image-generator workflow
UPDATE agent_definitions
SET default_config = jsonb_set(
jsonb_set(
default_config - 'workflow',  -- Remove old workflow
'{workflow}',
'{
"start_step": "generate",
"steps": {
"generate": {
"action": "generate_image",
"next_step": "complete",
"description": "Generate image and upload to S3"
},
"complete": {
"action": "complete_workflow",
"description": "Return image URI"
}
}
}'::jsonb
),
'{processing_mode}',
'"adapter"'  -- Should be adapter, not orchestrator
)
WHERE type = 'image-generator';