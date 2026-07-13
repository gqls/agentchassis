# Robot Hands Website - Workflow Diagram

## Overall Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         User / Client System                              │
│                                                                           │
│  Sends initial orchestrate message to:                                   │
│  Topic: system.agent.generic.requests                                    │
│  Group: robot-hands-complete-website                                     │
└───────────────────────────────┬──────────────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                      Generic Agent (Orchestrator)                         │
│                                                                           │
│  - Loads agent group definition from database                            │
│  - Reads orchestration workflow                                          │
│  - Executes workflow steps sequentially                                  │
│  - Manages spawned agents                                                │
└───────────────────────────────┬──────────────────────────────────────────┘
                                │
                                ▼
                        WORKFLOW EXECUTION
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
   PHASE 1: SPAWN          PHASE 2: GENERATE      PHASE 3: ASSEMBLE
```

## Phase 1: Agent Spawning (Steps 1-4)

```
┌─────────────────────┐
│  Step 1:            │
│  spawn_hero_writer  │───┐
└─────────────────────┘   │
                          ├──> Creates K8s Job
┌─────────────────────┐   │    - Pod: content-creator-hero-xxx
│  Spawned Agent:     │<──┘    - Topics: job.xxx.requests/responses
│  hero_writer        │         - Initializes and waits for work
│  (hero content)     │
└─────────────────────┘

┌─────────────────────┐
│  Step 2:            │
│ spawn_image_creator │───┐
└─────────────────────┘   │
                          ├──> Creates K8s Job
┌─────────────────────┐   │    - Pod: image-generator-xxx
│  Spawned Agent:     │<──┘    - Topics: job.xxx.requests/responses
│  image_creator      │         - Initializes and waits for work
│  (image generation) │
└─────────────────────┘

┌─────────────────────┐
│  Step 3:            │
│ spawn_about_writer  │───┐
└─────────────────────┘   │
                          ├──> Creates K8s Job
┌─────────────────────┐   │    - Pod: content-creator-about-xxx
│  Spawned Agent:     │<──┘    - Topics: job.xxx.requests/responses
│  about_writer       │         - Initializes and waits for work
│  (about content)    │
└─────────────────────┘

┌─────────────────────┐
│  Step 4:            │
│spawn_contact_writer │───┐
└─────────────────────┘   │
                          ├──> Creates K8s Job
┌─────────────────────┐   │    - Pod: content-creator-contact-xxx
│  Spawned Agent:     │<──┘    - Topics: job.xxx.requests/responses
│  contact_writer     │         - Initializes and waits for work
│  (contact content)  │
└─────────────────────┘
```

## Phase 2: Content Generation (Steps 5-8)

```
┌──────────────────────┐
│  Step 5:             │      ┌──────────────────────────────┐
│  generate_hero       │─────>│  Orchestrator sends message  │
└──────────────────────┘      │  to hero_writer agent        │
                              │                              │
                              │  Includes:                   │
                              │  - Prompt template           │
                              │  - Business name/type        │
                              │  - reply_to_topic            │
                              └──────────┬───────────────────┘
                                         │
                         ┌───────────────▼─────────────────┐
                         │  hero_writer agent receives     │
                         │  Executes LLM prompt            │
                         │  Returns: Hero headline & text  │
                         └───────────────┬─────────────────┘
                                         │
                                         ▼
                              Orchestrator stores result
                              in collected_data["generate_hero"]

┌──────────────────────┐
│  Step 6:             │      ┌──────────────────────────────┐
│ generate_hero_image  │─────>│  Orchestrator sends message  │
└──────────────────────┘      │  to image_creator agent      │
                              │                              │
                              │  Includes:                   │
                              │  - Image prompt              │
                              │  - Dimensions (1920x1080)    │
                              │  - reply_to_topic            │
                              └──────────┬───────────────────┘
                                         │
                         ┌───────────────▼──────────────────────┐
                         │  image_creator calls image adapter   │
                         │  Adapter uses AI image generation    │
                         │  Returns: Presigned S3 URL           │
                         └───────────────┬──────────────────────┘
                                         │
                                         ▼
                              Orchestrator stores result
                              in collected_data["generate_hero_image"]

┌──────────────────────┐
│  Step 7:             │      ┌──────────────────────────────┐
│  generate_about      │─────>│  Orchestrator sends message  │
└──────────────────────┘      │  to about_writer agent       │
                              │                              │
                              │  Includes:                   │
                              │  - About page prompt         │
                              │  - Agent collaboration info  │
                              │  - reply_to_topic            │
                              └──────────┬───────────────────┘
                                         │
                         ┌───────────────▼─────────────────┐
                         │  about_writer agent receives    │
                         │  Executes LLM prompt            │
                         │  Returns: About page content    │
                         └───────────────┬─────────────────┘
                                         │
                                         ▼
                              Orchestrator stores result
                              in collected_data["generate_about"]

┌──────────────────────┐
│  Step 8:             │      ┌──────────────────────────────┐
│  generate_contact    │─────>│  Orchestrator sends message  │
└──────────────────────┘      │  to contact_writer agent     │
                              │                              │
                              │  Includes:                   │
                              │  - Contact page prompt       │
                              │  - Contact details           │
                              │  - reply_to_topic            │
                              └──────────┬───────────────────┘
                                         │
                         ┌───────────────▼─────────────────┐
                         │  contact_writer agent receives  │
                         │  Executes LLM prompt            │
                         │  Returns: Contact page content  │
                         └───────────────┬─────────────────┘
                                         │
                                         ▼
                              Orchestrator stores result
                              in collected_data["generate_contact"]
```

## Phase 3: Page Assembly (Steps 9-11)

```
┌────────────────────────┐
│  Step 9:               │
│  assemble_homepage     │
└───────────┬────────────┘
            │
            ▼
    ┌──────────────────────────────────────────┐
    │  Orchestrator executes aggregate_webpage │
    │                                          │
    │  Combines:                               │
    │  - HTML head (with CSS, nav)             │
    │  - Hero content from generate_hero       │
    │  - Hero image URL from generate_hero_image│
    │  - HTML footer                           │
    │                                          │
    │  Creates: index.html                     │
    └───────────┬──────────────────────────────┘
                │
                ▼
    Result stored in collected_data["assemble_homepage"]

┌────────────────────────┐
│  Step 10:              │
│  assemble_about_page   │
└───────────┬────────────┘
            │
            ▼
    ┌──────────────────────────────────────────┐
    │  Orchestrator executes aggregate_webpage │
    │                                          │
    │  Combines:                               │
    │  - HTML head (with CSS, nav)             │
    │  - About content from generate_about     │
    │  - HTML footer                           │
    │                                          │
    │  Creates: about.html                     │
    └───────────┬──────────────────────────────┘
                │
                ▼
    Result stored in collected_data["assemble_about_page"]

┌────────────────────────┐
│  Step 11:              │
│  assemble_contact_page │
└───────────┬────────────┘
            │
            ▼
    ┌──────────────────────────────────────────┐
    │  Orchestrator executes aggregate_webpage │
    │                                          │
    │  Combines:                               │
    │  - HTML head (with CSS, nav)             │
    │  - Contact content from generate_contact │
    │  - HTML footer                           │
    │                                          │
    │  Creates: contact.html                   │
    └───────────┬──────────────────────────────┘
                │
                ▼
    Result stored in collected_data["assemble_contact_page"]
```

## Phase 4: Completion (Step 12)

```
┌────────────────────────┐
│  Step 12:              │
│  complete              │
└───────────┬────────────┘
            │
            ▼
    ┌──────────────────────────────────────────┐
    │  Orchestrator executes complete_workflow │
    │                                          │
    │  Returns to client:                      │
    │  {                                       │
    │    "index.html": "<html>...</html>",     │
    │    "about.html": "<html>...</html>",     │
    │    "contact.html": "<html>...</html>"    │
    │  }                                       │
    │                                          │
    │  Status: complete                        │
    └───────────┬──────────────────────────────┘
                │
                ▼
    Response sent to system.responses.generic
```

## Data Flow Details

### Message Flow Pattern

```
Orchestrator                 Spawned Agent
    │                             │
    │  1. spawn_agent            │
    │───────────────────────────>│
    │                             │ (Creates K8s Job)
    │  2. initialization ack      │
    │<───────────────────────────│
    │                             │
    │  3. call_agent (with work) │
    │───────────────────────────>│
    │   (includes reply_to_topic) │
    │                             │ (Processes work)
    │                             │
    │  4. response with results  │
    │<───────────────────────────│
    │  (sent to reply_to_topic)  │
    │                             │
    │  (continues workflow)       │
    │                             │ (Waits for more work)
```

### CollectedData Structure

```
collected_data: {
  "__execution_context__": { ... },
  "__my_requests_topic__": "system.agent.generic.requests",
  "__my_responses_topic__": "system.responses.generic",
  "__parent_responses_topic__": "system.responses.generic",
  
  "input_data": {
    "business_name": "Robot Hands",
    "business_type": "precision robotics and automation",
    "domain": "robot-hands.com"
  },
  
  "spawn_hero_writer": {
    "agent_id": "xxx",
    "agent_type": "content-creator-hero-without-research",
    "topics": { ... }
  },
  
  "generate_hero": {
    "headline": "Transform Your Manufacturing with Precision Robotics",
    "subheadline": "...",
    "success": true
  },
  
  "generate_hero_image": {
    "image_url": "https://s3.../presigned-url",
    "width": 1920,
    "height": 1080,
    "success": true
  },
  
  "generate_about": {
    "content": "<div class='content'>...</div>",
    "success": true
  },
  
  "generate_contact": {
    "content": "<div class='content'>...</div>",
    "success": true
  },
  
  "assemble_homepage": {
    "html": "<!DOCTYPE html>...",
    "page_name": "index.html",
    "success": true
  },
  
  "assemble_about_page": {
    "html": "<!DOCTYPE html>...",
    "page_name": "about.html",
    "success": true
  },
  
  "assemble_contact_page": {
    "html": "<!DOCTYPE html>...",
    "page_name": "contact.html",
    "success": true
  }
}
```

## Kafka Topics

```
System Topics (persistent):
├── system.agent.generic.requests       (Generic agent listens here)
├── system.responses.generic            (Orchestrator sends final response here)
└── system.errors.generic               (Error messages)

Job-Specific Topics (ephemeral - created for each orchestration):
├── job.<orch-id>-hero_writer.requests
├── job.<orch-id>-hero_writer.responses
├── job.<orch-id>-image_creator.requests
├── job.<orch-id>-image_creator.responses
├── job.<orch-id>-about_writer.requests
├── job.<orch-id>-about_writer.responses
├── job.<orch-id>-contact_writer.requests
└── job.<orch-id>-contact_writer.responses
```

## Key Principles

1. **Sequential Execution** - Steps execute in order, waiting for each to complete
2. **Data Accumulation** - Each step adds results to collected_data
3. **Response Routing** - Agents always respond to caller's response topic
4. **State Persistence** - orchestration_states table tracks progress
5. **Error Recovery** - Failed steps can be retried from last good state
6. **Agent Isolation** - Each spawned agent has unique topics and job
7. **Standardized Data** - data_helpers.go functions ensure consistency