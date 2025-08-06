https://claude.ai/chat/2d57dc36-1686-4021-8382-919342d4fa6e

Agent Chassis Architecture Report
Executive Summary
The Agent Chassis system implements a distributed, multi-agent orchestration platform where every agent is both a worker and an orchestrator. This architecture eliminates single points of failure while enabling complex multi-agent workflows with comprehensive monitoring and state management.
Core Architecture Components
1. Agent Chassis Structure
   Each deployed agent contains the following components:

Agent Chassis Instance
├── AgentServer         # Handles incoming work requests
├── AgentClient         # Handles responses from other agents
├── Orchestrator        # Manages workflow execution
├── MessageProcessor    # Routes messages to appropriate handlers
├── WorkflowValidator   # Validates workflow configurations
├── HealthServer        # Provides health checks and monitoring
└── MonitoringEndpoints # REST API for workflow visibility

2. Database Architecture
   The system uses two PostgreSQL databases with distinct responsibilities:
   Templates Database (Configuration)

agent_definitions
├── id: UUID
├── type: VARCHAR(100)           # e.g., "reasoning", "image-generator"
├── display_name: VARCHAR(255)   
├── category: VARCHAR(50)        # "code-driven" | "data-driven" | "adapter"
├── default_config: JSONB        # Default workflow and settings
├── is_active: BOOLEAN
└── timestamps

Clients Database (Runtime)

-- Client-specific schema (e.g., client_demo_client)
agent_instances
├── id: UUID                     # Instance identifier
├── template_id: UUID            # References agent_definitions
├── owner_user_id: VARCHAR
├── config: JSONB               # Can override template config
└── timestamps

-- Shared orchestration state
orchestrator_state
├── correlation_id: UUID         # Unique workflow instance ID
├── client_id: VARCHAR(100)      # Tenant identifier
├── status: VARCHAR(50)          # RUNNING, AWAITING_RESPONSES, COMPLETED, FAILED
├── current_step: VARCHAR(255)
├── workflow_plan: JSONB         # Complete workflow definition
├── execution_path: JSONB        # Step-by-step execution history
├── execution_metadata: JSONB    # Metrics and analytics
├── collected_data: JSONB        # Data gathered during execution
├── awaited_steps: JSONB         # Request IDs being waited for
└── timestamps

3. Kafka Topic Structure

Main Processing Topics:
├── system.agent.{agent-type}.process     # Incoming work
├── system.responses.{agent-type}         # Responses from other agents
├── system.errors.{agent-type}            # Error messages
└── system.adapter.{adapter-name}         # Adapter-specific topics

System Topics:
├── system.notifications.ui               # UI notifications
├── system.commands.workflow.resume       # Resume paused workflows
└── orchestrator.state-changes            # State change events

Workflow Execution Flow
1. Workflow Initiation

sequenceDiagram
    participant Client
    participant Kafka
    participant AgentPod1
    participant Database
    participant OtherAgents

    Client->>Kafka: Send request with headers
    Note over Kafka: correlation_id: unique<br/>agent_instance_id: config pointer<br/>client_id: tenant
    
    Kafka->>AgentPod1: Deliver to any available pod
    AgentPod1->>Database: Load agent instance config
    AgentPod1->>AgentPod1: Create workflow from config
    AgentPod1->>Database: Store orchestrator_state
    AgentPod1->>AgentPod1: Execute workflow steps

2. Multi-Agent Orchestration

graph TD
    A[AgentPod1: Receives Request] -->|Creates Workflow| B[Orchestrator State in DB]
    B --> C{Workflow Step Type?}

    C -->|Local Action| D[Execute Locally]
    D --> E[Update State in DB]
    
    C -->|Remote Action| F[Send to Agent Topic]
    F --> G[Update State: AWAITING_RESPONSES]
    
    C -->|Fan-out| H[Send to Multiple Agents]
    H --> I[Track Multiple Response IDs]
    
    J[AgentPod2: Receives Response] --> K[Load Workflow from DB]
    K --> L[Update Collected Data]
    L --> M{All Responses Received?}
    M -->|Yes| N[Continue Workflow]
    M -->|No| O[Keep Waiting]

3. Response Handling Flow

1. Other Agent processes request
2. Sends response to system.responses.{requesting-agent-type}
3. Response includes:
    - correlation_id (workflow identifier)
    - causation_id (original request_id)
    - Result data

4. ANY pod of requesting agent type:
    - Receives response via AgentClient
    - Loads workflow state from DB
    - Matches causation_id to awaited_steps
    - Updates workflow state
    - Continues execution if all responses received

Ownership Model
Workflow Ownership

Primary Owner: Client ID (tenant)
Identifier: Correlation ID
Executor: Any available agent pod
State Location: Database (not in memory)

Agent Instance Ownership

Owner: User within a client
Purpose: Configuration template
Usage: Defines HOW to process, not WHO processes

Key Principles:

Workflows are ephemeral - exist only during execution
Agents are stateless - any pod can handle any message
State is centralized - all state in database
Configuration is hierarchical - instance → template → defaults

Distributed Orchestration Model
Traditional vs Distributed
Traditional Centralized Orchestrator:

┌─────────────┐
│ Orchestrator│ ← Single point of failure
└──────┬──────┘
┌─────────┼─────────┐
↓         ↓         ↓
Agent A   Agent B   Agent C

Agent Chassis Distributed Model:
AgentPod A       AgentPod B       AgentPod C
├─ Server        ├─ Server        ├─ Server
├─ Client        ├─ Client        ├─ Client
└─ Orchestrator  └─ Orchestrator  └─ Orchestrator
↓                ↓                ↓
└────────────────┴────────────────┘
↓
Shared State (DB)

Benefits:

No single orchestrator service to manage
Horizontal scaling of orchestration capability
Fault tolerance through redundancy
Simplified deployment model

Local vs Remote Actions
Local Actions
Executed within the orchestrator itself:

actionRegistry = map[string]ActionHandler{
"validate_input":    ValidateInputAction,
"transform_data":    TransformDataAction,
"send_notification": SendNotificationAction,
}

Characteristics:

Synchronous execution
No network calls
Immediate state updates
Lower latency

Remote Actions
Executed by other agents via Kafka:

step:
action: "generate_image"
topic: "system.agent.image-generator.process"
next_step: "continue_after_image"

Characteristics:

Asynchronous execution
Network communication
State tracking via awaited_steps
Higher latency but better scalability

Monitoring and Observability
REST API Endpoints
Each agent exposes monitoring endpoints:

GET /monitor/workflows?client_id={client}
→ List active workflows

GET /monitor/workflow/{correlation_id}
→ Detailed workflow state and execution path

GET /monitor/stuck?hours={n}
→ Find workflows not progressing

GET /monitor/metrics?client_id={client}
→ Aggregate workflow metrics

Execution Tracking
Every workflow maintains:

{
"execution_path": [
    {
        "step": "validate",
        "action": "validate_input",
        "start_time": "2025-08-04T21:19:21.822Z",
        "end_time": "2025-08-04T21:19:23.171Z",
        "result": "success"
    }
    ],
        "execution_metadata": {
        "total_steps": 4,
        "completed_steps": 2,
        "failed_steps": 0,
        "checkpoints": {...}
    }
}

Scalability and Resilience
Horizontal Scaling

Agent pods can be scaled independently
Each pod is self-sufficient with embedded orchestrator
Kafka provides load balancing across pods

Fault Tolerance

Workflow state persisted in database
Any pod can continue any workflow
No in-memory state dependencies
Automatic failover through Kafka consumer groups

Multi-Tenancy

Client isolation through schemas
Separate workflow instances per client
Independent configuration per tenant
Shared infrastructure with logical separation

Example Workflow Execution
Scenario: Multi-Agent Creative Workflow

Request Arrives

Topic: system.agent.generic.process
Headers: {
    correlation_id: "ed240794-64b5-4133-8f5b-f19bb9ab6f00",
    agent_instance_id: "00000000-0000-0000-0000-000000000050",
    client_id: "demo_client"
}

Pod A Creates Workflow

Loads config from agent_instances
Creates orchestrator_state record
Executes fan_out to reasoning and image agents


Reasoning Agent Processes

Receives on system.agent.reasoning.process
Performs analysis
Sends response to system.responses.generic


Pod B Handles Response

AgentClient receives response
Loads workflow state
Updates collected_data
Checks if all responses received


Workflow Completes

All responses collected
Final step executed
Status updated to COMPLETED



Key Architectural Decisions

Embedded Orchestration: Every agent contains full orchestration capability
Stateless Processing: Pods don't maintain workflow state in memory
Database-Centric State: All state persisted for durability and visibility
Kafka for Distribution: Provides reliable message delivery and load balancing
Client/Server Separation: Clear responsibilities for request vs response handling
Comprehensive Monitoring: Built-in REST API for complete visibility

Conclusion
The Agent Chassis architecture provides a robust, scalable platform for multi-agent orchestration. 
By embedding orchestration capability within each agent and maintaining state in a central database, 
the system achieves high availability, fault tolerance, and operational simplicity while enabling 
complex multi-agent workflows with full observability.

--

It's been a pleasure working through this sophisticated architecture with you. You've built something really impressive here - the distributed orchestration model with embedded orchestrators is elegant, and the comprehensive monitoring system gives you excellent visibility into the workflows.
The fact that you avoided the traditional centralized orchestrator pattern in favor of this distributed approach shows great architectural thinking. It's much more resilient and scalable.