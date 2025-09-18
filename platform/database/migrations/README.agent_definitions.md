UPDATE agent_definitions
SET default_config = '{
"ai_service": {
"model": "claude-3-haiku-20240307",
"provider": "anthropic"
},
"processing_mode": "task",
"workflow": {
"start_step": "process",
"steps": {
"process": {
"action": "calculate",
"description": "Perform calculation",
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Complete calculation"
}
}
},
"prompt_template": "You are a calculator. Perform the operation ''{{.input_data.data.operation}}'' on the operands {{.input_data.data.operands}}. Respond ONLY with the final numerical result as a JSON object like {\"result\": 4}."
}'::jsonb
WHERE type = 'calculator';

UPDATE agent_definitions
SET default_config = '{
"ai_service": {
"model": "claude-3-5-sonnet-20241022",
"provider": "anthropic"
},
"processing_mode": "orchestrator",
"workflow": {
"start_step": "spawn_calculator",
"steps": {
"spawn_calculator": {
"action": "spawn_agent",
"config": {
"agent_type": "calculator"
},
"description": "Initialize calculator agent",
"next_step": "call_calculator"
},
"call_calculator": {
"action": "call_agent",
"config": {
"agent_type": "calculator"
},
"description": "Send calculation request",
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Complete workflow"
}
}
}
}'::jsonb
WHERE type = 'generic';

--

# two calculations
{
"workflow": {
"start_step": "spawn_adder",
"steps": {
"spawn_adder": {
"action": "spawn_agent",
"config": { "agent_type": "calculator" },
"description": "Spawn the first calculator for addition.",
"next_step": "spawn_multiplier"
},
"spawn_multiplier": {
"action": "spawn_agent",
"config": { "agent_type": "calculator" },
"description": "Spawn the second calculator for multiplication.",
"next_step": ["call_adder", "call_multiplier"]
},
"call_adder": {
"action": "call_agent",
"config": {
"agent_type": "calculator",
"spawn_step": "spawn_adder",
"input_data": "{{ .input_data.addition }}"
},
"description": "Call the first calculator to perform addition."
},
"call_multiplier": {
"action": "call_agent",
"config": {
"agent_type": "calculator",
"spawn_step": "spawn_multiplier",
"input_data": "{{ .input_data.multiplication }}"
},
"description": "Call the second calculator to perform multiplication."
},
"aggregate_results": {
"action": "aggregate_data",
"description": "Combine the results from both calculators.",
"dependencies": ["call_adder", "call_multiplier"],
"config": {
"sources": ["call_adder", "call_multiplier"]
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Complete the entire multi-calculation workflow."
}
}
},
"ai_service": {
"model": "claude-3-5-sonnet-20241022",
"provider": "anthropic"
},
"processing_mode": "orchestrator"
}

---
Option 1: Sequential Calculations (Simpler)
Modify the generic agent's workflow to call the calculator twice:

{
"workflow": {
"start_step": "spawn_calculator",
"steps": {
"spawn_calculator": {
"action": "spawn_agent",
"config": {"agent_type": "calculator"},
"description": "Initialize calculator agent",
"next_step": "first_calculation"
},
"first_calculation": {
"action": "call_agent",
"config": {
"agent_type": "calculator",
"target_action": "process",
"input_field": "first_calc_input"
},
"description": "First calculation",
"next_step": "second_calculation"
},
"second_calculation": {
"action": "call_agent",
"config": {
"agent_type": "calculator",
"target_action": "process",
"input_field": "second_calc_input"
},
"description": "Second calculation",
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Complete workflow"
}
}
}
}

--

Option 2: Parallel Calculations (More Complex)
Use a fan-out pattern to do both calculations simultaneously:

{
"workflow": {
"start_step": "spawn_calculators",
"steps": {
"spawn_calculators": {
"action": "spawn_group",
"config": {
"agents": [
{"type": "calculator", "name": "calc1"},
{"type": "calculator", "name": "calc2"}
]
},
"description": "Initialize calculator agents",
"next_step": "fan_out_calculations"
},
"fan_out_calculations": {
"action": "fan_out",
"sub_tasks": [
{
"step_name": "calc1",
"topic": "system.agent.calculator.requests",
"target_agent": "calc1",
"input_field": "first_calc_input"
},
{
"step_name": "calc2",
"topic": "system.agent.calculator.requests",
"target_agent": "calc2",
"input_field": "second_calc_input"
}
],
"next_step": "aggregate_results"
},
"aggregate_results": {
"action": "aggregate_data",
"description": "Combine both calculation results",
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Complete workflow"
}
}
}
}

Input Data Structure
Your initial request should include both calculations:

{
"action": "calculate",
"data": {
"first_calc_input": {
"operation": "add",
"operands": [2, 2]
},
"second_calc_input": {
"operation": "multiply",
"operands": [3, 4]
}
}
}

=============================
two calculations
=============================
-- Calculator stays the same
UPDATE agent_definitions
SET default_config = '{
"ai_service": {
"model": "claude-3-haiku-20240307",
"provider": "anthropic"
},
"processing_mode": "task",
"workflow": {
"start_step": "process",
"steps": {
"process": {
"action": "calculate",
"description": "Perform calculation",
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Complete calculation"
}
}
},
"prompt_template": "You are a calculator. Perform the operation ''{{.input_data.operation}}'' on the operands {{.input_data.operands}}. Respond ONLY with the final numerical result as a JSON object like {\"result\": 4}."
}'::jsonb
WHERE type = 'calculator';

-- Generic agent with sequential calculations
UPDATE agent_definitions
SET default_config = '{
"ai_service": {
"model": "claude-3-5-sonnet-20241022",
"provider": "anthropic"
},
"processing_mode": "orchestrator",
"workflow": {
"start_step": "spawn_calculator",
"steps": {
"spawn_calculator": {
"action": "spawn_agent",
"config": {
"agent_type": "calculator"
},
"description": "Initialize calculator agent",
"next_step": "first_calculation"
},
"first_calculation": {
"action": "call_agent",
"config": {
"agent_type": "calculator",
"input_field": "first_calc"
},
"description": "First calculation",
"next_step": "second_calculation"
},
"second_calculation": {
"action": "call_agent",
"config": {
"agent_type": "calculator",
"input_field": "second_calc"
},
"description": "Second calculation",
"next_step": "aggregate_results"
},
"aggregate_results": {
"action": "aggregate_data",
"config": {
"strategy": "group_responses"
},
"description": "Combine calculation results",
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Complete workflow"
}
}
}
}'::jsonb
WHERE type = 'generic';

2. Message to Send
   kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- bash << 'EOF'
   CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
   REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
   MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
   ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
   ORCHESTRATION_NAME="dual-calc-$(date +%H%M%S)"
   echo "Dual calculation test message:"
   echo "  Correlation: $CORRELATION_ID"
   echo "  Request: $REQUEST_ID"
   echo "  Orchestration: $ORCHESTRATION_ID"
   echo "  Name: $ORCHESTRATION_NAME"

# Create the message with two calculations
cat << JSON | \
/opt/kafka/bin/kafka-console-producer.sh \
--bootstrap-server localhost:9092 \
--topic system.agent.generic.requests \
--property parse.headers=true \
--property headers.delimiter=$'\t'
correlation_id:$CORRELATION_ID,orchestration_id:$ORCHESTRATION_ID,request_id:$REQUEST_ID,message_id:$MESSAGE_ID,client_id:demo_client,message_type:request,responses_topic:system.agent.generic.responses	{"action":"calculate","first_calc":{"operation":"add","operands":[2,2]},"second_calc":{"operation":"multiply","operands":[3,4]}}
JSON
EOF

# Wait and check for calculator job
sleep 10
kubectl -n ai-persona-system get jobs | grep calculator
kubectl -n ai-persona-system get pods | grep calculator