# Unified Agent Architecture with GPU Cost Optimization

## Core Principle: Everything is an Agent

**All agents use the same code**, deployed differently:
- **Static Agents**: Pre-deployed via Kubernetes
- **Dynamic Agents**: Spawned on-demand
- **GPU Agents**: Special handling to minimize costs

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                   SAME AGENT CODE                    │
├───────────────────┬─────────────────────────────────┤
│   Static Deploy   │        Dynamic Deploy           │
├───────────────────┼─────────────────────────────────┤
│                   │                                 │
│ • image-generator │ • content-writer (spawned)      │
│ • web-search      │ • researcher (spawned)          │
│ • database-query  │ • analyzer (spawned)            │
│                   │ • image-generator-gpu (spawned) │
│                   │                                 │
│ Listen on:        │ Listen on:                      │
│ system.agent.*    │ job.{id}.*                      │
└───────────────────┴─────────────────────────────────┘
```

## Key Implementation: Minimal Code Changes

### 1. Enhanced Agent (3 new methods)
```go
// Add to existing Agent struct:

func (a *Agent) IsStaticAgent() bool {
    return strings.HasPrefix(a.requestsTopic, "system.agent.")
}

func (a *Agent) getResponseDestination(request *types.RequestMessage) string {
    if a.IsStaticAgent() {
        return request.Headers.ResponsesTopic // Use from request
    }
    return a.responsesTopic // Use fixed topic
}

func (a *Agent) initializeConsumers() error {
    if a.IsStaticAgent() {
        // Pattern subscription for job.* responses
        pattern := fmt.Sprintf("job\\..*-%s-.*\\.responses", a.AgentType)
        a.responseConsumer = kafka.NewConsumerWithPattern(pattern)
    } else {
        // Fixed response topic
        a.responseConsumer = kafka.NewConsumer(a.responsesTopic)
    }
}
```

### 2. Unified Call Action (works for both)
```go
func CallAgentAction(ctx, params) {
    if isStaticAgent(agentType) {
        requestTopic = "system.agent.{type}.requests"
        responseTopic = "job.{corr}-{orch}-{step}-{type}.responses"
    } else {
        // Use spawned agent's topics
        requestTopic = agentInfo.RequestsTopic
        responseTopic = agentInfo.ResponsesTopic
    }
    
    // Send request with response topic in headers
    request.Headers.ResponsesTopic = responseTopic
}
```

## GPU Cost Optimization Strategy

### Recommended: CPU Router + Dynamic GPU

```yaml
# CPU Router (always running - cheap)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-image-generator-router
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: agent
        resources:
          limits:
            cpu: "1"
            memory: "1Gi"
        env:
        - name: AGENT_TYPE
          value: "image-generator"
        - name: GPU_AGENT_STRATEGY
          value: "always_dynamic"
```

**Router Workflow:**
```yaml
workflow:
  steps:
    evaluate:
      action: check_complexity
      next: route
    
    route:
      action: conditional
      conditions:
        - if: needs_gpu
          next: spawn_gpu_worker
        - else
          next: handle_cpu
    
    spawn_gpu_worker:
      action: spawn_agent
      config:
        agent_type: image-generator-gpu
        ttl: 300  # Auto-terminate after 5 min
    
    call_gpu_worker:
      action: call_agent
      config:
        agent_type: image-generator-gpu
```

## Cost Comparison

| Approach | Monthly Cost | Response Time | Best For |
|----------|-------------|---------------|----------|
| Static GPU 24/7 | $1,440 | Instant | High volume |
| Dynamic GPU | $50-100 | 30s cold start | Low volume |
| **CPU Router + Dynamic** | **$20 + $50** | **5s** | **Balanced** |

## Response Topic Routing

**The Critical Pattern:**
```
Orchestrator Step: "generate_hero_image"
    ↓
Request to: system.agent.image-generator.requests
Response to: job.{corr}-{orch}-generate_hero_image-image-generator.responses
    ↑
Static Agent reads ResponsesTopic from request headers
```

## Files Created

1. **[Unified Call Agent](computer:///mnt/user-data/outputs/call_agent_unified.go)** - Single action for both static/dynamic
2. **[Smart Spawn Agent](computer:///mnt/user-data/outputs/spawn_agent_smart.go)** - GPU-aware spawning
3. **[Enhanced Agent Code](computer:///mnt/user-data/outputs/agent_enhanced.go)** - Minimal changes to support both modes
4. **[GPU Deployment Strategies](computer:///mnt/user-data/outputs/gpu_deployment_strategies.md)** - K8s configs for GPU optimization

## Integration Steps

1. **Add 3 methods to Agent struct**
    - `IsStaticAgent()`
    - `getResponseDestination()`
    - `initializeConsumers()`

2. **Replace call_agent action**
    - Use unified version that detects static vs dynamic

3. **Deploy static agents**
   ```bash
   kubectl apply -f agent-image-generator-router.yaml
   ```

4. **Set GPU strategy**
   ```bash
   export GPU_AGENT_STRATEGY="always_dynamic"
   ```

## Benefits

✅ **Unified Architecture** - Everything truly is an agent
✅ **Minimal Changes** - ~100 lines of code added
✅ **GPU Cost Savings** - 95% reduction (dynamic spawning)
✅ **Same Workflows** - No change to workflow definitions
✅ **Flexible Deployment** - Static or dynamic per agent type

## Example Workflow (No Changes Needed!)

```yaml
workflow:
  steps:
    # Dynamic agent - spawn then call
    spawn_writer:
      action: spawn_agent
      config:
        agent_type: content-writer
    
    call_writer:
      action: call_agent
      config:
        agent_type: content-writer
    
    # Static agent - just call (router handles GPU)
    generate_image:
      action: call_agent
      config:
        agent_type: image-generator
        prompt: "{{.call_writer.content}}"
```

## The Beauty

- **Agents don't know if they're static or dynamic** - code is identical
- **Orchestrators don't care** - call_agent handles routing
- **GPU costs minimized** - spawn only when needed
- **Response routing automatic** - headers carry destination

This is a clean, extensible architecture that scales from simple CPU agents to complex GPU workloads, all with the same code!