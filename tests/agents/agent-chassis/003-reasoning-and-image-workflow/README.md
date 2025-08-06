kubectl -n ai-persona-system exec -it agent-chassis-test -- sh

kubectl exec -it postgres-clients-0 -n ai-persona-system -- sh
/ # psql -Uclient_user -d clients_db


SELECT config->'workflow' as workflow
FROM client_demo_client.agent_instances
WHERE id = '00000000-0000-0000-0000-000000000001';


Let's create a workflow that uses both the reasoning agent and image creation agent. This will demonstrate real agent-to-agent orchestration.
First, let's check what agents and topics we have:
# List agent types
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"SELECT DISTINCT type, description FROM agent_definitions WHERE is_active = true;"

# List Kafka topics
kubectl exec -it kafka-client-test -n kafka -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--list | grep system.agent

Now let's create a workflow that:

Sends a prompt to the reasoning agent to generate an image description
Takes that description and sends it to the image creation agent
Completes with both the reasoning and the image URL

kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"UPDATE client_demo_client.agent_instances
SET config = '{
\"workflow\": {
\"start_step\": \"reason_about_image\",
\"steps\": {
\"reason_about_image\": {
\"action\": \"generate_image_description\",
\"description\": \"Ask reasoning agent to create detailed image description\",
\"topic\": \"system.agent.reasoning.process\",
\"next_step\": \"create_image\"
},
\"create_image\": {
\"action\": \"generate_image\",
\"description\": \"Create image from description\",
\"topic\": \"system.agent.image-creator.process\",
\"dependencies\": [\"reason_about_image\"],
\"next_step\": \"complete\"
},
\"complete\": {
\"action\": \"complete_workflow\",
\"description\": \"Workflow completed with reasoning and image\"
}
}
}
}'
WHERE id = '00000000-0000-0000-0000-000000000001';"

Now let's create a test script that sends a creative prompt:

package main

import (
"context"
"encoding/json"
"fmt"
"log"
"time"

    "github.com/google/uuid"
    "github.com/segmentio/kafka-go"
)

func main() {
writer := kafka.NewWriter(kafka.WriterConfig{
Brokers: []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"},
Topic:   "system.agent.generic.process",
})
defer writer.Close()

    // Generate UUIDs
    correlationID := uuid.New().String()
    requestID := uuid.New().String()
    
    // Create a creative request
    payload := map[string]interface{}{
        "action": "create_artwork",
        "data": map[string]interface{}{
            "theme": "futuristic city",
            "style": "cyberpunk",
            "mood": "vibrant neon lights at night",
            "request": "Please describe and then create an image of a futuristic cyberpunk city",
        },
    }
    
    payloadBytes, _ := json.Marshal(payload)
    
    msg := kafka.Message{
        Headers: []kafka.Header{
            {Key: "correlation_id", Value: []byte(correlationID)},
            {Key: "request_id", Value: []byte(requestID)},
            {Key: "client_id", Value: []byte("demo_client")},
            {Key: "agent_instance_id", Value: []byte("00000000-0000-0000-0000-000000000001")},
            {Key: "fuel_budget", Value: []byte("1000")}, // More fuel for multiple agents
        },
        Key:   []byte("artwork-request"),
        Value: payloadBytes,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    err := writer.WriteMessages(ctx, msg)
    if err != nil {
        log.Fatal("failed to write message:", err)
    }
    
    fmt.Printf("Creative workflow initiated!\n")
    fmt.Printf("Correlation ID: %s\n", correlationID)
    fmt.Printf("\nMonitor with:\n")
    fmt.Printf("curl http://localhost:8080/monitor/workflow/%s | jq .\n", correlationID)
}

But wait - we need to make sure the reasoning and image-creator agents are configured to handle these requests. Let's check their configurations:

# Check reasoning agent instances
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"SELECT id, type, config->'workflow' as workflow
FROM client_demo_client.agent_instances
WHERE type IN ('reasoning', 'image-creator');"

# Create reasoning agent instance
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"INSERT INTO client_demo_client.agent_instances (id, type, config)
VALUES (
'00000000-0000-0000-0000-000000000002',
'reasoning',
'{\"workflow\": {\"start_step\": \"complete\", \"steps\": {\"complete\": {\"action\": \"complete_workflow\"}}}}'
);"

# Create image-creator agent instance
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"INSERT INTO client_demo_client.agent_instances (id, type, config)
VALUES (
'00000000-0000-0000-0000-000000000003',
'image-creator',
'{\"workflow\": {\"start_step\": \"complete\", \"steps\": {\"complete\": {\"action\": \"complete_workflow\"}}}}'
);"




clients_db=# SELECT schemaname, tablename FROM pg_tables WHERE tablename LIKE '%agent%' ORDER BY schemaname, tablename;
schemaname     |     tablename     
--------------------+-------------------
client_demo_client | agent_instances
public             | agent_definitions
(2 rows)
clients_db=# SELECT type, display_name, category, is_active FROM agent_definitions ORDER BY type;
type       |    display_name    |  category   | is_active
-----------------+--------------------+-------------+-----------
content-creator | Content Creator    | data-driven | t
copywriter      | Copywriter         | data-driven | t
generic         | Generic Agent      | code-driven | t
image-generator | Image Generator    | adapter     | t
reasoning       | Reasoning Agent    | code-driven | t
researcher      | Research Assistant | data-driven | t
web-search      | Web Search         | adapter     | t
(7 rows)
--
SELECT * FROM client_demo_client.agent_instances ;
id                  |             template_id              | owner_user_id |        name        |                                                                                                                                                                                                                                 config                                                                                                                                                                                                                                  | is_active |          created_at           |          updated_at           
--------------------------------------+--------------------------------------+---------------+--------------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+-----------+-------------------------------+-------------------------------
00000000-0000-0000-0000-000000000001 | 2a540b98-85d5-4762-a692-538bcf1be395 | test-user     | Test Generic Agent | {"workflow": {"steps": {"notify": {"action": "send_notification", "next_step": "complete", "description": "Send notification"}, "complete": {"action": "complete_workflow"}, "validate": {"action": "validate_input", "next_step": "transform", "description": "Validate the input"}, "transform": {"action": "transform_data", "config": {"transformation": "uppercase"}, "next_step": "notify", "description": "Transform to uppercase"}}, "start_step": "validate"}} | t         | 2025-08-03 08:08:04.765909+00 | 2025-08-03 08:08:04.765909+00
(1 row)
--
I'd like to leave that basic test in there and create another agent to test this latest more complicated call.

Perfect! Let's create a new agent instance specifically for testing multi-agent orchestration. We'll create an instance that orchestrates calls to the reasoning and image-generator agents:
bash

# Create a new orchestrator agent instance
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"INSERT INTO client_demo_client.agent_instances (
id,
template_id,
owner_user_id,
name,
config,
is_active
) VALUES (
'00000000-0000-0000-0000-000000000010',
'2a540b98-85d5-4762-a692-538bcf1be395',  -- generic template
'test-user',
'Creative Orchestrator',
'{
\"workflow\": {
\"start_step\": \"analyze_request\",
\"steps\": {
\"analyze_request\": {
\"action\": \"validate_input\",
\"description\": \"Analyze the creative request\",
\"next_step\": \"generate_description\"
},
\"generate_description\": {
\"action\": \"reason_about_image\",
\"description\": \"Use reasoning agent to create detailed image description\",
\"topic\": \"system.agent.reasoning.process\",
\"next_step\": \"create_image\"
},
\"create_image\": {
\"action\": \"generate_image\",
\"description\": \"Generate image from description\",
\"topic\": \"system.agent.image-generator.process\",
\"dependencies\": [\"generate_description\"],
\"next_step\": \"complete\"
},
\"complete\": {
\"action\": \"complete_workflow\",
\"description\": \"Complete with reasoning and image results\"
}
}
}
}',
true
);"

Now let's also create instances for the reasoning and image-generator agents so they can process the requests:
bash

# Create reasoning agent instance
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"INSERT INTO client_demo_client.agent_instances (
id,
template_id,
owner_user_id,
name,
config,
is_active
) VALUES (
'00000000-0000-0000-0000-000000000020',
(SELECT id FROM agent_definitions WHERE type = 'reasoning'),
'test-user',
'Reasoning Agent Instance',
'{
\"workflow\": {
\"start_step\": \"process\",
\"steps\": {
\"process\": {
\"action\": \"process_reasoning\",
\"description\": \"Process reasoning request\",
\"next_step\": \"complete\"
},
\"complete\": {
\"action\": \"complete_workflow\"
}
}
}
}',
true
);"

# Create image-generator agent instance
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"INSERT INTO client_demo_client.agent_instances (
id,
template_id,
owner_user_id,
name,
config,
is_active
) VALUES (
'00000000-0000-0000-0000-000000000030',
(SELECT id FROM agent_definitions WHERE type = 'image-generator'),
'test-user',
'Image Generator Instance',
'{
\"workflow\": {
\"start_step\": \"generate\",
\"steps\": {
\"generate\": {
\"action\": \"create_image\",
\"description\": \"Generate image from prompt\",
\"next_step\": \"complete\"
},
\"complete\": {
\"action\": \"complete_workflow\"
}
}
}
}',
true
);"

Let's also create a test that uses fan-out to call multiple agents in parallel:
bash

# Create a fan-out orchestrator
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"INSERT INTO client_demo_client.agent_instances (
id,
template_id,
owner_user_id,
name,
config,
is_active
) VALUES (
'00000000-0000-0000-0000-000000000040',
'2a540b98-85d5-4762-a692-538bcf1be395',
'test-user',
'Multi-Agent Orchestrator',
'{
\"workflow\": {
\"start_step\": \"prepare\",
\"steps\": {
\"prepare\": {
\"action\": \"validate_input\",
\"description\": \"Prepare the request\",
\"next_step\": \"fan_out_to_agents\"
},
\"fan_out_to_agents\": {
\"action\": \"fan_out\",
\"description\": \"Call multiple agents in parallel\",
\"subtasks\": [
{
\"step_name\": \"web_search\",
\"topic\": \"system.agent.web-search.process\"
},
{
\"step_name\": \"content_creation\",
\"topic\": \"system.agent.content-creator.process\"
},
{
\"step_name\": \"copywriting\",
\"topic\": \"system.agent.copywriter.process\"
}
],
\"next_step\": \"complete\"
},
\"complete\": {
\"action\": \"complete_workflow\",
\"description\": \"Complete with all agent results\"
}
}
}
}',
true
);"

Now let's create a test script for the creative orchestrator:
go

// test-creative-workflow.go
package main

import (
"context"
"encoding/json"
"fmt"
"log"
"time"

    "github.com/google/uuid"
    "github.com/segmentio/kafka-go"
)

func main() {
writer := kafka.NewWriter(kafka.WriterConfig{
Brokers: []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"},
Topic:   "system.agent.generic.process",
})
defer writer.Close()

    // Test the Creative Orchestrator
    correlationID := uuid.New().String()
    requestID := uuid.New().String()
    
    payload := map[string]interface{}{
        "action": "create_artwork",
        "data": map[string]interface{}{
            "prompt": "A futuristic cyberpunk city with neon lights reflecting on wet streets",
            "style": "digital art, highly detailed",
            "mood": "mysterious, atmospheric",
        },
    }
    
    payloadBytes, _ := json.Marshal(payload)
    
    msg := kafka.Message{
        Headers: []kafka.Header{
            {Key: "correlation_id", Value: []byte(correlationID)},
            {Key: "request_id", Value: []byte(requestID)},
            {Key: "client_id", Value: []byte("demo_client")},
            {Key: "agent_instance_id", Value: []byte("00000000-0000-0000-0000-000000000010")}, // Creative Orchestrator
            {Key: "fuel_budget", Value: []byte("1000")},
        },
        Key:   []byte("creative-request"),
        Value: payloadBytes,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    err := writer.WriteMessages(ctx, msg)
    if err != nil {
        log.Fatal("failed to write message:", err)
    }
    
    fmt.Printf("Creative workflow initiated!\n")
    fmt.Printf("Correlation ID: %s\n", correlationID)
    fmt.Printf("\nMonitor progress:\n")
    fmt.Printf("curl http://localhost:8080/monitor/workflow/%s | jq .\n", correlationID)
    
    // Test the Fan-out Orchestrator
    time.Sleep(2 * time.Second)
    
    correlationID2 := uuid.New().String()
    requestID2 := uuid.New().String()
    
    payload2 := map[string]interface{}{
        "action": "research_and_write",
        "data": map[string]interface{}{
            "topic": "artificial intelligence in healthcare",
            "requirements": "comprehensive research with marketing copy",
        },
    }
    
    payloadBytes2, _ := json.Marshal(payload2)
    
    msg2 := kafka.Message{
        Headers: []kafka.Header{
            {Key: "correlation_id", Value: []byte(correlationID2)},
            {Key: "request_id", Value: []byte(requestID2)},
            {Key: "client_id", Value: []byte("demo_client")}


clients_db=# \d
List of relations
Schema |        Name        | Type  |    Owner     
--------+--------------------+-------+--------------
public | agent_definitions  | table | clients_user
public | orchestrator_state | table | clients_user
(2 rows)

clients_db=# \d agent_definitions;
Table "public.agent_definitions"
Column     |           Type           | Collation | Nullable |      Default      
----------------+--------------------------+-----------+----------+-------------------
id             | uuid                     |           | not null | gen_random_uuid()
type           | character varying(100)   |           | not null |
display_name   | character varying(255)   |           | not null |
description    | text                     |           |          |
category       | character varying(50)    |           | not null |
default_config | jsonb                    |           | not null | '{}'::jsonb
is_active      | boolean                  |           |          | true
created_at     | timestamp with time zone |           | not null | now()
updated_at     | timestamp with time zone |           | not null | now()
deleted_at     | timestamp with time zone |           |          |
Indexes:
"agent_definitions_pkey" PRIMARY KEY, btree (id)
"agent_definitions_type_key" UNIQUE CONSTRAINT, btree (type)
"idx_agent_definitions_type_active" btree (type, is_active) WHERE deleted_at IS NULL
Check constraints:
"agent_definitions_category_check" CHECK (category::text = ANY (ARRAY['data-driven'::character varying, 'code-driven'::character varying, 'adapter'::character varying]::text[]))

clients_db=# \d orchestrator_state;
Table "public.orchestrator_state"
Column        |           Type           | Collation | Nullable |   Default   
----------------------+--------------------------+-----------+----------+-------------
correlation_id       | uuid                     |           | not null |
client_id            | character varying(100)   |           | not null |
status               | character varying(50)    |           | not null |
current_step         | character varying(255)   |           | not null |
awaited_steps        | jsonb                    |           |          | '[]'::jsonb
collected_data       | jsonb                    |           |          | '{}'::jsonb
initial_request_data | jsonb                    |           |          |
final_result         | jsonb                    |           |          |
error                | text                     |           |          |
created_at           | timestamp with time zone |           | not null | now()
updated_at           | timestamp with time zone |           | not null | now()
workflow_plan        | jsonb                    |           |          |
execution_metadata   | jsonb                    |           |          | '{}'::jsonb
execution_path       | jsonb                    |           |          | '[]'::jsonb
Indexes:
"orchestrator_state_pkey" PRIMARY KEY, btree (correlation_id)
"idx_orchestrator_state_client" btree (client_id)
"idx_orchestrator_state_client_status" btree (client_id, status)
"idx_orchestrator_state_status" btree (status)
"idx_orchestrator_state_updated_at" btree (updated_at)

clients_db=# SELECT schemaname, tablename FROM pg_tables WHERE tablename LIKE '%agent%' ORDER BY schemaname, tablename;
schemaname     |     tablename     
--------------------+-------------------
client_demo_client | agent_instances
public             | agent_definitions


===
now doing:
# Create a new orchestrator agent instance
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
INSERT INTO client_demo_client.agent_instances (
id,
template_id,
owner_user_id,
name,
config,
is_active
) VALUES (
'00000000-0000-0000-0000-000000000010',
'2a540b98-85d5-4762-a692-538bcf1be395',  -- generic template
'test-user',
'Creative Orchestrator',
'{
"workflow": {
"start_step": "analyze_request",
"steps": {
"analyze_request": {
"action": "validate_input",
"description": "Analyze the creative request",
"next_step": "generate_description"
},
"generate_description": {
"action": "reason_about_image",
"description": "Use reasoning agent to create detailed image description",
"topic": "system.agent.reasoning.process",
"next_step": "create_image"
},
"create_image": {
"action": "generate_image",
"description": "Generate image from description",
"topic": "system.agent.image-generator.process",
"dependencies": ["generate_description"],
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Complete with reasoning and image results"
}
}
}
}',
true
);

# Now let's also create instances for the reasoning and image-generator agents so they can process the requests:
# Create reasoning agent instance
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
"INSERT INTO client_demo_client.agent_instances (
id,
template_id,
owner_user_id,
name,
config,
is_active
) VALUES (
'00000000-0000-0000-0000-000000000020',
(SELECT id FROM agent_definitions WHERE type = 'reasoning'),
'test-user',
'Reasoning Agent Instance',
'{
\"workflow\": {
\"start_step\": \"process\",
\"steps\": {
\"process\": {
\"action\": \"process_reasoning\",
\"description\": \"Process reasoning request\",
\"next_step\": \"complete\"
},
\"complete\": {
\"action\": \"complete_workflow\"
}
}
}
}',
true
);"

# Create image-generator agent instance
kubectl exec -it postgres-clients-0 -n ai-persona-sy....
INSERT INTO client_demo_client.agent_instances (
id,
template_id,
owner_user_id,
name,
config,
is_active
) VALUES (
'00000000-0000-0000-0000-000000000030',
(SELECT id FROM agent_definitions WHERE type = 'image-generator'),
'test-user',
'Image Generator Instance',
'{
"workflow": {
"start_step": "generate",
"steps": {
"generate": {
"action": "create_image",
"description": "Generate image from prompt",
"next_step": "complete"
},
"complete": {
"action": "complete_workflow"
}
}
}
}',
true
);

--
# Let's also create a test that uses fan-out to call multiple agents in parallel:
# Create a fan-out orchestrator
kubectl exec -it postgres-clients-0 -n ai-perso/////
# Create a fan-out orchestrator
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c
INSERT INTO client_demo_client.agent_instances (
id,
template_id,
owner_user_id,
name,
config,
is_active
) VALUES (
'00000000-0000-0000-0000-000000000040',
'2a540b98-85d5-4762-a692-538bcf1be395',
'test-user',
'Multi-Agent Orchestrator',
'{
"workflow": {
"start_step": "prepare",
"steps": {
"prepare": {
"action": "validate_input",
"description": "Prepare the request",
"next_step": "fan_out_to_agents"
},
"fan_out_to_agents": {
"action": "fan_out",
"description": "Call multiple agents in parallel",
"subtasks": [
{
"step_name": "web_search",
"topic": "system.agent.web-search.process"
},
{
"step_name": "content_creation",
"topic": "system.agent.content-creator.process"
},
{
"step_name": "copywriting",
"topic": "system.agent.copywriter.process"
}
],
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Complete with all agent results"
}
}
}
}',
true
);

--
# Now let's create a test script for the creative orchestrator:
(in tests/agents/reasoning-and-image-workflow/test-creative-workflow.go)


