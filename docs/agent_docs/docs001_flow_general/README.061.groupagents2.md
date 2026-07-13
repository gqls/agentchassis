Looking at your sophisticated agent group system, you're absolutely right - we should leverage it instead of constantly modifying the generic agent definition. Let's set up a proper agent group for testing.

## Step 1: Create a Simple Welcome Message Generator Group

First, let's create an agent group in the database that uses a single content-creator agent:

```sql
-- Insert into agent_groups table
INSERT INTO agent_groups (
    id, 
    name, 
    group_type, 
    version,
    agent_configs,
    orchestration_workflow,
    capabilities,
    is_active
) VALUES (
    'welcome-gen-001',
    'Welcome Message Generator',
    'welcome-message-generator',
    '1.0.0',
    '[
        {
            "role": "writer",
            "agent_type": "content-creator"
        }
    ]'::jsonb,
    '{
        "start_step": "create_welcome",
        "steps": {
            "create_welcome": {
                "action": "execute_llm_prompt",
                "config": {
                    "prompt_template": "Create a welcoming message for {{.business_name}}, a {{.business_type}}. Make it warm, inviting, and 2-3 sentences long."
                },
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow"
            }
        }
    }'::jsonb,
    '["content-generation", "welcome-messages"]'::jsonb,
    true
);
```

## Step 2: Update Your Message Processor to Handle Group-Based Workflows

In your `processor.go`, implement the workflow selection logic:

```go
func (p *MessageProcessor) selectWorkflow(ctx context.Context, message *types.RequestMessage) (map[string]interface{}, error) {
    // Priority 1: Check for inline workflow override
    if body, ok := message.Body.(map[string]interface{}); ok {
        if config, ok := body["config"].(map[string]interface{}); ok {
            if workflow, ok := config["workflow"].(map[string]interface{}); ok {
                p.logger.Info("Using inline workflow override")
                return workflow, nil
            }
        }
    }
    
    // Priority 2: Check for group-based workflow
    if message.Headers.Action == "spawn_group" || message.Headers.Action == "orchestrate" {
        if body, ok := message.Body.(map[string]interface{}); ok {
            // Check both places where group_type might be
            groupType := ""
            if config, ok := body["config"].(map[string]interface{}); ok {
                groupType, _ = config["group_type"].(string)
            }
            if groupType == "" {
                groupType, _ = body["group_type"].(string)
            }
            
            if groupType != "" {
                // Use GroupDiscovery to find the best group
                discovery := discovery.NewGroupDiscovery(p.db)
                group, err := discovery.FindBestGroup(ctx, groupType, nil)
                if err == nil && group != nil {
                    var workflow map[string]interface{}
                    if err := json.Unmarshal(group.Workflow, &workflow); err == nil {
                        p.logger.Info("Using group-based workflow",
                            zap.String("group_type", groupType),
                            zap.String("group_id", group.ID),
                            zap.String("version", group.Version))
                        return workflow, nil
                    }
                }
            }
        }
    }
    
    // Priority 3: Use agent's default workflow from database
    return p.loadAgentWorkflow(ctx, p.agentType)
}
```

## Step 3: Test Message Structure

Now you can send a simple message without any workflow:

```json
{
    "action": "orchestrate",
    "config": {
        "group_type": "welcome-message-generator"
    },
    "input_data": {
        "business_type": "pizza restaurant",
        "business_name": "Test Pizza"
    }
}
```

## Step 4: Create More Complex Groups

For website building, create a proper group:

```sql
INSERT INTO agent_groups (
    id,
    name,
    group_type,
    version,
    agent_configs,
    orchestration_workflow,
    capabilities,
    is_active
) VALUES (
    'website-builder-001',
    'Simple Website Builder',
    'simple-website-builder', 
    '1.0.0',
    '[
        {
            "role": "content_writer",
            "agent_type": "content-creator"
        },
        {
            "role": "html_coder",
            "agent_type": "html-developer"
        }
    ]'::jsonb,
    '{
        "start_step": "spawn_team",
        "steps": {
            "spawn_team": {
                "action": "spawn_agent",
                "config": {
                    "agent_type": "content-creator",
                    "role": "content_writer"
                },
                "next_step": "create_content"
            },
            "create_content": {
                "action": "call_agent",
                "config": {
                    "target_role": "content_writer"
                },
                "next_step": "generate_html"
            },
            "generate_html": {
                "action": "generate_html",
                "next_step": "store_website"
            },
            "store_website": {
                "action": "upload_to_s3",
                "config": {
                    "make_public": true
                },
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow"
            }
        }
    }'::jsonb,
    '["website-building", "content-generation", "html-generation"]'::jsonb,
    true
);
```

## Step 5: Test Commands

From within your Kubernetes cluster:

```bash
# Get into a pod with kcat/kafkacat
kubectl exec -it kafka-test-pod -n ai-persona-system -- /bin/sh

# Send message to generic agent to orchestrate
CID=$(cat /proc/sys/kernel/random/uuid)
echo '{
    "action": "orchestrate",
    "config": {
        "group_type": "welcome-message-generator"
    },
    "input_data": {
        "business_type": "pizza restaurant",
        "business_name": "Test Pizza"
    }
}' | kcat -P \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H "correlation_id=$CID" \
    -H "request_id=$(cat /proc/sys/kernel/random/uuid)" \
    -H "client_id=demo_client" \
    -H "fuel_budget=1000"
```

## Benefits of This Approach

1. **Versioned Workflows** - Each group has a version, allowing evolution
2. **Performance Tracking** - Groups track success rates and usage
3. **No Code Changes** - New workflows are just database entries
4. **Discoverable** - The system can find the best group for a task
5. **Auditable** - Complete history of workflow changes

This way, you maintain clean separation between:
- **Testing**: Use inline workflow overrides when needed
- **Production**: Use versioned agent groups from the database
- **Evolution**: The system can automatically improve groups based on performance

The generic agent remains generic - it just orchestrates based on the group type specified in the message!