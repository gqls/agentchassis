# 009 — Agent Message Structure

How to send messages to agents from the CLI (or any external system). Covers the message format, Kafka headers, the spawn+call pattern, and response handling.

## Overview

All agents receive work via Kafka messages. A message has three layers:

1. **Kafka headers** — routing metadata (correlation_id, action, topic names)
2. **Message body headers** — duplicate of Kafka headers in JSON (for systems that can't read Kafka headers)
3. **Message body payload** — `config` (workflow definition) + `input_data` (the actual work)

## Kafka Headers

Every message needs these headers. They travel as Kafka record headers and are also mirrored inside the JSON body.

| Header | Purpose | Example |
|---|---|---|
| `correlation_id` | Links all messages in one job together | UUID |
| `orchestration_id` | Identifies the orchestration state row | UUID |
| `request_id` | Unique ID for this specific message | UUID |
| `message_id` | Deduplication key | UUID |
| `message_type` | `request` or `response` | `request` |
| `client_id` | Tenant identifier | `demo_client` |
| `action` | What the receiving agent should do | `process` or `orchestrate` |
| `sender_agent_type` | Who sent this | `cli` |
| `sender_agent_id` | Sender instance ID | `cli-user` |
| `responses_topic` | Where the agent should send its reply | `system.agent.generic.responses` |
| `timestamp` | ISO 8601 timestamp | `2025-02-19T10:00:00Z` |

The `responses_topic` header is important — agents always reply to the **caller's** responses topic, not their own. This is how parent-child orchestration works.

## Message Body Structure

```json
{
  "headers": {
    "correlation_id": "...",
    "orchestration_id": "...",
    "request_id": "...",
    "message_id": "...",
    "message_type": "request",
    "client_id": "demo_client",
    "action": "process",
    "sender": {
      "agent_id": "cli-user",
      "agent_type": "cli",
      "pod_name": "cli"
    },
    "timestamp": "..."
  },
  "config": {
    "workflow": { ... }
  },
  "input_data": { ... }
}
```

The three top-level keys:

- **`headers`** — JSON mirror of the Kafka headers, plus the `sender` object
- **`config`** — contains the `workflow` definition (what steps to run)
- **`input_data`** — the domain-specific payload (domain name, edit instructions, etc.)

## The Spawn+Call Pattern

Most agent triggers from the CLI use the **generic agent** as a thin launcher. The generic agent doesn't do domain work itself — it spawns the specialist agent and calls it. This keeps the launch message simple and lets the specialist run self-contained with its own workflow.

```
CLI message → system.agent.generic.requests
               │
               └─ generic agent runs inline workflow:
                   1. spawn_agent (creates specialist instance)
                   2. call_agent  (forwards input_data to specialist)
                   3. complete    (returns specialist's result)
                        │
                        └─ specialist runs its own full workflow
                            on its own topic, with its own workers
```

The inline workflow goes in `config.workflow`:

```json
{
  "workflow": {
    "start_step": "spawn_specialist",
    "processing_mode": "orchestrator",
    "timeout_seconds": 900,
    "steps": {
      "spawn_specialist": {
        "action": "spawn_agent",
        "config": {
          "role": "my_role",
          "agent_type": "my-specialist"
        },
        "output_field": "specialist_agent",
        "next_step": "call_specialist"
      },
      "call_specialist": {
        "action": "call_agent",
        "config": {
          "agent_type": "my-specialist",
          "target_role": "my_role",
          "input_mapping": {
            "domain": "input_data.domain",
            "some_field": "input_data.some_field"
          },
          "timeout_seconds": 600
        },
        "output_field": "result",
        "next_step": "complete"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {
          "output_fields": ["result"]
        }
      }
    }
  }
}
```

The `input_mapping` in `call_agent` controls which fields from the original `input_data` get forwarded to the specialist. Fields that aren't present in `input_data` resolve to nil and are ignored.

## Generating IDs

From bash, UUIDs come from the kernel:

```bash
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
```

The correlation_id ties everything together — use it to grep logs:

```bash
kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep "$CORRELATION_ID"
```

## Sending via kcat

Messages go to Kafka using `kcat` (run as a temporary pod in the kafka namespace):

```bash
kubectl -n kafka run -i --rm kcat-my-trigger-$(date +%s) \
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
{"headers":{...},"config":{"workflow":{...}},"input_data":{...}}
JSON
```

The JSON body must be on a single line (no newlines) when piped to kcat. Use `tr -d '\n'` or write it inline.

## HITL Responses

When an agent workflow hits a human-in-the-loop step, it creates an `awaited_requests` row and pauses. The HITL response goes to the agent's **responses** topic (not requests), with these additional headers:

| Header | Value |
|---|---|
| `message_type` | `response` |
| `in_response_to_request_id` | The request_id from the awaited_requests row |
| `in_response_to_step_name` | The step that's waiting (e.g. `hitl_confirm_type`) |
| `status` | `complete` |
| `sender_agent_type` | `human` |

The body contains `{"body": {"success": true, "human_response": true, ...}}` with whatever fields the waiting step expects.

## Topic Naming Convention

```
system.agent.{agent-type}.requests    — where agents receive work
system.agent.{agent-type}.responses   — where agents send results
```

The generic agent uses `system.agent.generic.requests` / `system.agent.generic.responses`.

Specialist agents get their own topics (e.g. `system.agent.section-editor.requests`). When spawned, the orchestrator automatically routes messages to the right topic based on agent_type.

## Database State

Each orchestration creates a row in `orchestration_states` with:
- `workflow_plan` — the workflow definition
- `collected_data` — accumulates step outputs as the workflow progresses
- `current_step` — which step is executing
- `status` — RUNNING, AWAITING_RESPONSES, COMPLETED, etc.

The `collected_data` is how steps communicate. Step A stores output at `output_field`, Step B reads it via `input_fields` or `ExtractFields`.