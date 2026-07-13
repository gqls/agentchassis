# Triggering Agent Workflows via Kafka Messages

## Core Concept: Agents Must Be Spawned Before Called

Agents don't exist as running processes until spawned. The orchestration system:

1. **Spawn** - Creates an agent instance (pod) with a specific role
2. **Call** - Sends work to that spawned agent instance
3. **Complete** - Collects results and ends the workflow

You cannot call an agent that hasn't been spawned in the current workflow.

## Message Structure

Messages to `system.agent.generic.requests` have two main parts:

```json
{
  "headers": { ... },
  "config": {
    "workflow": {
      "start_step": "first_step_name",
      "steps": { ... }
    }
  },
  "input_data": { ... }
}
```

### Required Headers (also passed as Kafka headers)

| Header | Purpose |
|--------|---------|
| `correlation_id` | Groups related messages (UUID) |
| `orchestration_id` | Identifies this workflow run (UUID) |
| `request_id` | Unique message identifier (UUID) |
| `message_id` | Kafka message ID (UUID) |
| `message_type` | Always `request` |
| `client_id` | Tenant identifier |
| `action` | Always `process` for workflows |
| `sender_agent_type` | Who's sending (e.g., `cli`) |
| `timestamp` | ISO 8601 format |

## Workflow Pattern: Spawn → Call → Complete

### Minimal Two-Step Pattern

```json
{
  "config": {
    "workflow": {
      "start_step": "spawn_agent",
      "steps": {
        "spawn_agent": {
          "action": "spawn_agent",
          "config": {
            "role": "worker",
            "agent_type": "my-agent-type"
          },
          "output_field": "spawned_agent",
          "next_step": "call_agent"
        },
        "call_agent": {
          "action": "call_agent",
          "config": {
            "agent_type": "my-agent-type",
            "target_role": "worker",
            "input_mapping": {
              "field1": "input_data.field1",
              "field2": "input_data.field2"
            },
            "timeout_seconds": 300
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
  },
  "input_data": {
    "field1": "value1",
    "field2": "value2"
  }
}
```

### Key Rules

| Rule | Details |
|------|---------|
| `role` in spawn must match `target_role` in call | Links the spawned instance to the call |
| `agent_type` must exist in `agent_definitions` table | The agent's workflow is loaded from DB |
| `input_mapping` paths reference `input_data.*` | Maps workflow input to agent input |
| `output_field` stores step results | Available to subsequent steps |

## Common Agent Types

| Agent Type | Purpose | Typical Timeout |
|------------|---------|-----------------|
| `pageflow-builder` | Build pages with content generation | 1800s |
| `rerender-pages` | Re-render pages from stored sections | 900s |
| `site-classifier` | Classify site type | 300s |
| `briefing-questionnaire` | Generate site brief | 600s |

## Example: Trigger pageflow-builder

```bash
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

kubectl -n kafka run -i --rm kcat-trigger \
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
-H timestamp=$TIMESTAMP <<JSON
{
  "headers": {
    "correlation_id": "${CORRELATION_ID}",
    "orchestration_id": "${ORCHESTRATION_ID}",
    "request_id": "${REQUEST_ID}",
    "message_id": "${MESSAGE_ID}",
    "message_type": "request",
    "client_id": "${CLIENT_ID}",
    "action": "process",
    "sender": {"agent_id": "cli-user", "agent_type": "cli", "pod_name": "cli"},
    "timestamp": "${TIMESTAMP}"
  },
  "config": {
    "workflow": {
      "start_step": "spawn_builder",
      "steps": {
        "spawn_builder": {
          "action": "spawn_agent",
          "config": {
            "role": "builder",
            "agent_type": "pageflow-builder"
          },
          "output_field": "builder_agent",
          "next_step": "call_builder"
        },
        "call_builder": {
          "action": "call_agent",
          "config": {
            "agent_type": "pageflow-builder",
            "target_role": "builder",
            "input_mapping": {
              "site_id": "input_data.site_id",
              "domain": "input_data.domain",
              "repo_name": "input_data.repo_name"
            },
            "timeout_seconds": 1800
          },
          "output_field": "build_result",
          "next_step": "complete"
        },
        "complete": {
          "action": "complete_workflow",
          "config": {
            "output_fields": ["build_result"]
          }
        }
      }
    }
  },
  "input_data": {
    "site_id": "uuid-here",
    "domain": "example.com",
    "repo_name": "sites"
  }
}
JSON
```

## Chaining Multiple Agents

For workflows that need multiple agents:

```json
{
  "start_step": "spawn_first",
  "steps": {
    "spawn_first": {
      "action": "spawn_agent",
      "config": { "role": "classifier", "agent_type": "site-classifier" },
      "next_step": "call_first",
      "output_field": "classifier_agent"
    },
    "call_first": {
      "action": "call_agent",
      "config": {
        "agent_type": "site-classifier",
        "target_role": "classifier",
        "input_mapping": { "domain": "input_data.domain" }
      },
      "next_step": "spawn_second",
      "output_field": "classification"
    },
    "spawn_second": {
      "action": "spawn_agent",
      "config": { "role": "builder", "agent_type": "pageflow-builder" },
      "next_step": "call_second",
      "output_field": "builder_agent"
    },
    "call_second": {
      "action": "call_agent",
      "config": {
        "agent_type": "pageflow-builder",
        "target_role": "builder",
        "input_mapping": {
          "site_id": "classification.response.site_id",
          "domain": "input_data.domain"
        }
      },
      "next_step": "complete",
      "output_field": "build_result"
    },
    "complete": {
      "action": "complete_workflow",
      "config": { "output_fields": ["classification", "build_result"] }
    }
  }
}
```

## Troubleshooting

| Problem | Cause | Fix |
|---------|-------|-----|
| "agent not found" | Agent not spawned | Add spawn_agent step before call_agent |
| "target_role not found" | Role mismatch | Ensure spawn `role` matches call `target_role` |
| "agent_type not found" | Missing definition | Check `agent_definitions` table |
| Timeout | Agent took too long | Increase `timeout_seconds` |
| "input_mapping failed" | Wrong path | Check available paths in error logs |

## Watching Workflow Progress

```bash
# Watch agent pods
kubectl -n ai-persona-system get pods -w

# Watch logs for specific orchestration
kubectl -n ai-persona-system logs -f -l app=agent-chassis | grep $ORCHESTRATION_ID

# Check Kafka topics
kubectl -n kafka run -i --rm kcat-consumer --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -C -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.responses -o end
```