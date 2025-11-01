# Robot Hands Complete Website Configuration

## Overview

This configuration creates a complete website for robot-hands.com with three pages:
1. **Homepage** - Hero section with compelling content and hero image
2. **About Page** - Explaining the agent orchestration system
3. **Contact Page** - Contact information and invitation to connect

## Architecture

### Agent Group: `robot-hands-complete-website`

The agent group orchestrates four specialized agents:

1. **hero_writer** (`content-creator-hero-without-research`)
    - Creates compelling hero section content
    - Focuses on headlines and value propositions

2. **image_creator** (`image-generator`)
    - Generates hero background image using AI
    - Returns presigned URLs for image access

3. **about_writer** (`content-creator-about`)
    - Creates about page content
    - Explains the agent orchestration system

4. **contact_writer** (`content-creator-contact`)
    - Creates contact page content
    - Includes contact details and welcoming text

### Workflow Steps

The orchestration follows this sequence:

```
1. spawn_hero_writer       → Spawn hero content writer agent
2. spawn_image_creator     → Spawn image generation agent
3. spawn_about_writer      → Spawn about page writer agent
4. spawn_contact_writer    → Spawn contact page writer agent
5. generate_hero           → Generate hero section content
6. generate_hero_image     → Generate hero background image (1920x1080)
7. generate_about          → Generate about page content
8. generate_contact        → Generate contact page content
9. assemble_homepage       → Combine hero + image into index.html
10. assemble_about_page    → Create about.html
11. assemble_contact_page  → Create contact.html
12. complete               → Return all pages
```

## Key Design Decisions

### 1. Agent Spawning Strategy
- **All agents spawned first** before content generation
- Agents respond to their caller's response topic (not their own)
- Each spawned agent gets unique job-specific topics

### 2. Data Flow Using data_helpers.go

The new `data_helpers.go` provides standardized functions for:

```go
// Building request messages
BuildRequestMessage(execCtx, "content-creator-about", "process", data, config, logger)

// Extracting data from any message format
ExtractDataFromMessage(source, logger)

// Building collected data structures
BuildCollectedData(message, execCtx, logger)

// Building response messages
BuildResponseMessage(execCtx, success, responseData, errorInfo, logger)
```

### 3. Image Handling
- Images generated via `image-generator` adapter
- Returns presigned URLs (not base64 data)
- URLs embedded directly in HTML `<img>` tags
- Images stored in S3/MinIO object storage

### 4. HTML Assembly
- Uses `aggregate_webpage` action
- Provides full HTML structure with CSS
- Responsive design with mobile viewport
- Fixed navigation across all pages
- Consistent styling and branding

## Installation Steps

### 1. Apply Agent Definitions

```bash
# Create about page writer agent
psql -h <host> -U templates_user -d templates_db -f agent_content_creator_about.sql

# Create contact page writer agent
psql -h <host> -U templates_user -d templates_db -f agent_content_creator_contact.sql
```

### 2. Apply Agent Group Definition

```bash
# Create the agent group orchestration
psql -h <host> -U templates_user -d templates_db -f robot_hands_complete_website_agent_group.sql
```

### 3. Trigger Website Creation

```bash
# Make script executable
chmod +x create_robot_hands_website.sh

# Run the script
./create_robot_hands_website.sh
```

## Monitoring and Debugging

### View Orchestration Progress

```bash
# Get orchestration ID from script output, then:
kubectl logs -f deployment/agent-chassis -n agent-system | grep '<ORCHESTRATION_ID>'
```

### Check Database State

```sql
SELECT 
    orchestration_id, 
    status, 
    current_step, 
    collected_data,
    updated_at 
FROM orchestration_states 
WHERE orchestration_id = '<ORCHESTRATION_ID>' 
ORDER BY updated_at DESC;
```

### View Spawned Agent Pods

```bash
# List all spawned job agents
kubectl get pods -n agent-system | grep job

# Tail specific agent logs
kubectl logs -f -n agent-system <pod-name>
```

## Expected Output

The workflow produces three HTML files with:

### index.html
- Full-screen hero section with gradient background
- Hero image (1920x1080) as background
- Compelling headline and subheadline
- Call-to-action button
- Navigation to other pages

### about.html
- Clean content layout
- Explanation of agent orchestration
- Description of how agents collaborated
- Professional styling

### contact.html
- Welcoming message
- Contact information (email, phone, address)
- Clean, structured layout
- Contact info in styled box

## Extending the Configuration

### Adding More Pages

1. Create new agent definition (e.g., `content-creator-services`)
2. Add to `agent_configs` in agent group
3. Add spawn step to workflow
4. Add generation step with prompt
5. Add assembly step for new page
6. Update navigation in HTML wrappers

### Adding Research Sub-Agents

To add research capability to any writer:

1. Update agent workflow to include:
    - `spawn_researcher` step
    - `call_researcher` step
    - Use research results in content generation

2. Example structure (from existing pattern):
```json
{
    "spawn_researcher": {
        "action": "spawn_agent",
        "config": {
            "role": "researcher",
            "agent_type": "content-researcher"
        },
        "next_step": "call_researcher"
    },
    "call_researcher": {
        "action": "call_agent",
        "config": {
            "agent_type": "content-researcher",
            "target_role": "researcher",
            "prompt": "Research {{.business_type}}"
        },
        "next_step": "generate_content"
    }
}
```

## Using the New Data Helpers

The `data_helpers.go` functions standardize data handling:

### In Workflow Actions

```go
// Extract clean data from incoming message
data := ExtractDataFromMessage(message, logger)

// Build request to another agent
request := BuildRequestMessage(
    execCtx,
    "content-creator-about",
    "process",
    data,
    config,
    logger,
)

// Build response back to parent
response := BuildResponseMessage(
    execCtx,
    true,
    resultData,
    nil,
    logger,
)
```

### Getting Data from Previous Steps

```go
// Get data from specific step
heroData, found := GetStepData(collectedData, "generate_hero", logger)

// Get data from multiple steps
results := GetMultipleStepData(
    collectedData,
    []string{"generate_hero", "generate_about"},
    logger,
)

// Get field using path notation
value, err := GetFieldFromPath(collectedData, "generate_hero.headline", logger)
```

## Troubleshooting

### Agent Not Spawning
- Check agent definition exists in database
- Verify agent type matches in agent_configs
- Check Kubernetes deployment for agent-chassis

### Content Not Generated
- Verify LLM API key is set (ANTHROPIC_API_KEY)
- Check agent logs for prompt issues
- Ensure input_data contains required fields

### Image Not Appearing
- Verify image-generator adapter is running
- Check S3/MinIO connectivity
- Confirm presigned URLs in response

### Pages Not Assembled
- Check aggregate_webpage implementation
- Verify response_fields match step names
- Ensure HTML wrapper templates are valid

## Response Topic Routing

**Important:** Agents always respond to the caller's response topic, not their own:

- Parent spawns child with `reply_to_topic=parent.responses`
- Child processes request
- Child sends response to `parent.responses` (NOT `child.responses`)
- Parent receives response and continues workflow

This is handled automatically by:
- `BuildRequestMessage()` sets correct `reply_to_topic`
- `ExecutionContext.CreateChildContext()` propagates routing
- Response handlers use `execCtx.ReplyToTopic`

## Files Generated

```
/home/claude/
├── robot_hands_complete_website_agent_group.sql  # Agent group definition
├── agent_content_creator_about.sql               # About page writer agent
├── agent_content_creator_contact.sql             # Contact page writer agent
└── create_robot_hands_website.sh                 # Trigger script
```

## Next Steps

1. Apply SQL configurations to database
2. Run the trigger script
3. Monitor orchestration progress
4. Retrieve generated HTML files from completed workflow
5. Deploy HTML to web server or S3 bucket

## Notes

- All agents use the same chassis image (`docker.io/aqls/agent-chassis:v1.0.407`)
- Agents are ephemeral - spawned for workflow, then cleaned up
- HTML generation is done server-side in the orchestrator
- Images are hosted externally with presigned URLs
- Styling is embedded in HTML (no external CSS files)