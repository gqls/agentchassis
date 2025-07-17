# Reasoning Agent Internal API Documentation

## Overview

The Reasoning Agent is a code-driven agent that performs logical analysis, decision making, and complex reasoning tasks. It processes requests via Kafka messages and returns structured reasoning results.

## Kafka Topics

### Consumed Topics

#### agents.reasoning.process
Main processing topic for reasoning requests.

**Message Format:**
```json
{
  "action": "analyze|decide|evaluate|reason",
  "data": {
    "context": "Background information for reasoning",
    "question": "What needs to be analyzed or decided",
    "constraints": [
      "List of constraints to consider"
    ],
    "options": [
      {
        "id": "option1",
        "description": "First option to consider"
      }
    ],
    "criteria": {
      "factors": ["cost", "time", "quality"],
      "weights": {
        "cost": 0.3,
        "time": 0.2,
        "quality": 0.5
      }
    }
  }
}
```

**Required Headers:**
- `correlation_id`: Unique request identifier
- `request_id`: Request tracking ID
- `client_id`: Client identifier
- `agent_instance_id`: Specific agent instance
- `fuel_budget`: Available fuel for this operation

### Produced Topics

#### agents.reasoning.results
Results of reasoning operations.

**Message Format:**
```json
{
  "success": true,
  "data": {
    "reasoning_type": "analysis|decision|evaluation",
    "conclusion": "Main conclusion or recommendation",
    "reasoning_steps": [
      {
        "step": 1,
        "description": "Identified key factors",
        "findings": ["Factor 1", "Factor 2"]
      },
      {
        "step": 2,
        "description": "Evaluated options",
        "analysis": {
          "option1": {
            "score": 0.85,
            "pros": ["Pro 1", "Pro 2"],
            "cons": ["Con 1"]
          }
        }
      }
    ],
    "confidence": 0.85,
    "assumptions": [
      "List of assumptions made"
    ],
    "recommendations": [
      {
        "priority": "high",
        "action": "Recommended action",
        "rationale": "Why this is recommended"
      }
    ]
  }
}
```

#### system.agent.metrics
Performance metrics and usage data.

**Message Format:**
```json
{
  "agent_type": "reasoning",
  "agent_instance_id": "instance-uuid",
  "metrics": {
    "processing_time_ms": 1500,
    "fuel_consumed": 25,
    "reasoning_depth": 3,
    "options_evaluated": 5
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Agent Configuration

The reasoning agent accepts the following configuration parameters:

```json
{
  "model": "claude-3-opus",
  "temperature": 0.2,
  "max_reasoning_steps": 10,
  "reasoning_style": "analytical|creative|balanced",
  "output_format": "structured|narrative",
  "enable_assumptions": true,
  "confidence_threshold": 0.7,
  "parallel_evaluation": true,
  "memory_enabled": true,
  "memory_context_limit": 5
}
```

## Supported Actions

### analyze
Performs deep analysis of a situation or problem.

**Input:**
```json
{
  "action": "analyze",
  "data": {
    "subject": "Market entry strategy for Product X",
    "context": "Current market conditions and company capabilities",
    "aspects": ["market_size", "competition", "regulatory", "timing"]
  }
}
```

### decide
Makes decisions based on given criteria and options.

**Input:**
```json
{
  "action": "decide",
  "data": {
    "decision": "Which cloud provider to choose",
    "options": ["AWS", "GCP", "Azure"],
    "criteria": {
      "factors": ["cost", "features", "support"],
      "constraints": ["Must support Kubernetes", "Budget < $10k/month"]
    }
  }
}
```

### evaluate
Evaluates proposals, plans, or strategies.

**Input:**
```json
{
  "action": "evaluate",
  "data": {
    "proposal": "Details of the proposal",
    "evaluation_criteria": ["feasibility", "roi", "risk"],
    "benchmark": "Optional benchmark for comparison"
  }
}
```

### reason
General reasoning about complex scenarios.

**Input:**
```json
{
  "action": "reason",
  "data": {
    "scenario": "Description of the scenario",
    "question": "What needs to be determined",
    "approach": "deductive|inductive|abductive"
  }
}
```

## Memory Integration

When memory is enabled, the agent:
- Retrieves relevant past reasoning for similar problems
- Stores successful reasoning patterns
- Learns from previous decisions and their outcomes

**Memory Entry Format:**
```json
{
  "type": "reasoning_pattern",
  "content": {
    "problem_type": "decision|analysis|evaluation",
    "pattern": "Description of successful reasoning approach",
    "effectiveness": 0.9,
    "context_tags": ["finance", "strategy", "risk"]
  }
}
```

## Error Handling

The agent returns errors in the following format:

```json
{
  "success": false,
  "error": {
    "code": "REASONING_001",
    "message": "Unable to complete reasoning",
    "details": {
      "reason": "Insufficient context provided",
      "missing_elements": ["criteria", "constraints"]
    }
  }
}
```

### Error Codes
- `REASONING_001`: Invalid input format
- `REASONING_002`: Insufficient context
- `REASONING_003`: Reasoning timeout
- `REASONING_004`: Conflicting constraints
- `REASONING_005`: Low confidence result

## Performance Characteristics

- Average processing time: 1-3 seconds
- Fuel consumption: 10-50 units depending on complexity
- Memory usage: ~100MB baseline + variable based on context
- Concurrent request handling: Yes
- Max context size: 100KB

## Integration Notes

### For Orchestrators
- Provide comprehensive context for better reasoning
- Include all relevant constraints upfront
- Specify desired output format
- Handle low-confidence results appropriately

### For Other Agents
- Can chain reasoning results with other agents
- Use structured output for easier parsing
- Consider confidence scores in workflows

## Environment Variables

```bash
# Agent Configuration
AGENT_TYPE=reasoning
KAFKA_CONSUMER_GROUP=reasoning-agent-group
KAFKA_TOPICS=agents.reasoning.process

# AI Service
AI_PROVIDER=anthropic
ANTHROPIC_API_KEY=<api-key>

# Performance
MAX_CONCURRENT_REQUESTS=10
REQUEST_TIMEOUT=30s
MEMORY_LIMIT=512MB

# Monitoring
METRICS_ENABLED=true
METRICS_INTERVAL=60s
```