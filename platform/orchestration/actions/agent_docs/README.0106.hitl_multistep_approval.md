#!/bin/bash
# Working HITL Test Workflows - Multiple Options

echo "================================"
echo "HITL TEST WORKFLOWS THAT WORK"
echo "================================"

# Option 1: Simple workflow without LLM (fastest to test)
echo ""
echo "OPTION 1: Simple Test Without LLM"
echo "---------------------------------"
cat << 'EOF'
CORRELATION_ID=$(uuidgen)
REQUEST_ID=$(uuidgen)

kubectl -n kafka run -i --rm kcat-producer-simple \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -P -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_type=request \
-H action=orchestrate \
<<JSON
{
"action": "orchestrate",
"config": {
"workflow": {
"start_step": "prepare_data",
"steps": {
"prepare_data": {
"action": "transform_data",
"config": {
"transformations": {
"greeting": "Hello {{name}}, welcome to the approval system!",
"timestamp": "2025-11-04T14:00:00Z",
"status": "pending_approval"
}
},
"next_step": "await_approval",
"description": "Prepare greeting data"
},
"await_approval": {
"action": "await_approval",
"topic": "system.agent.generic.requests",
"config": {
"approval_fields": ["prepare_data"],
"approval_type": "greeting_approval",
"ui_config": {
"title": "Greeting Approval Required",
"description": "Please approve this greeting message"
}
},
"next_step": "process_approval",
"description": "Wait for approval"
},
"process_approval": {
"action": "process_approval_decision",
"topic": "system.agent.generic.requests",
"next_step": "complete",
"description": "Process the decision"
},
"complete": {
"action": "complete_workflow",
"description": "Complete"
}
}
}
},
"input_data": {
"name": "Test User"
}
}
JSON
EOF

# Option 2: With LLM - includes AI service config
echo ""
echo "OPTION 2: With LLM and AI Service Config"
echo "----------------------------------------"
cat << 'EOF'
CORRELATION_ID=$(uuidgen)
REQUEST_ID=$(uuidgen)

kubectl -n kafka run -i --rm kcat-producer-llm \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -P -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_type=request \
-H action=orchestrate \
<<JSON
{
"action": "orchestrate",
"config": {
"ai_service": {
"provider": "anthropic",
"model": "claude-3-5-sonnet-20241022",
"api_key_env_var": "ANTHROPIC_API_KEY"
},
"workflow": {
"start_step": "generate",
"steps": {
"generate": {
"action": "execute_llm_prompt",
"config": {
"prompt_template": "Write a professional greeting message for {{name}} who is joining our team as a {{role}}. Keep it brief and welcoming.",
"input_fields": ["name", "role"]
},
"next_step": "await_approval",
"description": "Generate greeting content"
},
"await_approval": {
"action": "await_approval",
"topic": "system.agent.generic.requests",
"config": {
"approval_fields": ["generate"],
"approval_type": "content_review",
"ui_config": {
"title": "Content Approval Required",
"description": "Please approve the generated greeting"
}
},
"next_step": "process_approval",
"description": "Wait for approval"
},
"process_approval": {
"action": "process_approval_decision",
"topic": "system.agent.generic.requests",
"next_step": "complete",
"description": "Process the decision"
},
"complete": {
"action": "complete_workflow",
"description": "Complete"
}
}
}
},
"input_data": {
"name": "Jane Smith",
"role": "Senior Developer"
}
}
JSON
EOF

# Option 3: Multi-step with validation (no LLM needed)
echo ""
echo "OPTION 3: Multi-Step Validation and Approval"
echo "--------------------------------------------"
cat << 'EOF'
CORRELATION_ID=$(uuidgen)
REQUEST_ID=$(uuidgen)

kubectl -n kafka run -i --rm kcat-producer-multi \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -P -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_type=request \
-H action=orchestrate \
<<JSON
{
"action": "orchestrate",
"config": {
"workflow": {
"start_step": "validate",
"steps": {
"validate": {
"action": "validate_input",
"config": {
"required_fields": ["name", "department", "start_date"],
"optional_fields": ["manager", "location"]
},
"next_step": "prepare",
"description": "Validate input data"
},
"prepare": {
"action": "transform_data",
"config": {
"transformations": {
"employee_id": "EMP-{{department}}-001",
"welcome_message": "Welcome {{name}} to the {{department}} department!",
"onboarding_status": "pending_approval",
"created_at": "2025-11-04"
}
},
"next_step": "create_request",
"description": "Prepare employee data"
},
"create_request": {
"action": "create_approval_request",
"topic": "system.agent.generic.requests",
"config": {
"metadata": {
"request_type": "new_employee_onboarding",
"priority": "high"
}
},
"next_step": "await_approval",
"description": "Create approval request"
},
"await_approval": {
"action": "await_approval",
"topic": "system.agent.generic.requests",
"config": {
"approval_fields": ["prepare"],
"approval_type": "onboarding_approval",
"timeout_seconds": 3600,
"ui_config": {
"title": "New Employee Onboarding Approval",
"description": "Please review and approve the new employee details",
"editable_fields": ["welcome_message", "start_date"]
}
},
"next_step": "process_approval",
"description": "Wait for HR approval"
},
"process_approval": {
"action": "process_approval_decision",
"topic": "system.agent.generic.requests",
"config": {
"stop_on_reject": false
},
"next_step": "route_decision",
"description": "Process HR decision"
},
"route_decision": {
"action": "conditional_route",
"config": {
"condition_field": "process_approval.approved",
"routes": {
"true": "finalize",
"false": "handle_rejection"
}
},
"description": "Route based on approval"
},
"finalize": {
"action": "transform_data",
"config": {
"transformations": {
"status": "approved",
"employee_record": "{{prepare}}",
"approval_details": "{{process_approval}}",
"message": "Employee onboarding approved and ready to proceed"
}
},
"next_step": "complete",
"description": "Finalize approved onboarding"
},
"handle_rejection": {
"action": "transform_data",
"config": {
"transformations": {
"status": "rejected",
"reason": "{{process_approval.comments}}",
"message": "Onboarding request was not approved"
}
},
"next_step": "complete",
"description": "Handle rejection"
},
"complete": {
"action": "complete_workflow",
"description": "Return final result"
}
}
}
},
"input_data": {
"name": "John Doe",
"department": "Engineering",
"start_date": "2025-11-15",
"manager": "Jane Smith",
"location": "San Francisco"
}
}
JSON
EOF

# Option 4: Using an agent that has AI config
echo ""
echo "OPTION 4: Call Content Writer Agent (Has AI Config)"
echo "---------------------------------------------------"
cat << 'EOF'
# First, make sure you have a content writer agent deployed
# Then use this workflow:

CORRELATION_ID=$(uuidgen)
REQUEST_ID=$(uuidgen)

kubectl -n kafka run -i --rm kcat-producer-agent \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
kcat -P -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_type=request \
-H action=orchestrate \
<<JSON
{
"action": "orchestrate",
"config": {
"workflow": {
"start_step": "spawn_writer",
"steps": {
"spawn_writer": {
"action": "spawn_agent",
"config": {
"agent_type": "content-creator",
"role": "writer"
},
"next_step": "generate_content",
"description": "Spawn content writer"
},
"generate_content": {
"action": "call_agent",
"config": {
"target_role": "writer",
"agent_type": "content-creator",
"input_data": {
"prompt": "Write a welcome message for {{name}}",
"tone": "professional"
}
},
"next_step": "await_approval",
"description": "Generate content via agent"
},
"await_approval": {
"action": "await_approval",
"topic": "system.agent.generic.requests",
"config": {
"approval_fields": ["generate_content"],
"approval_type": "content_review"
},
"next_step": "process_approval",
"description": "Wait for approval"
},
"process_approval": {
"action": "process_approval_decision",
"topic": "system.agent.generic.requests",
"next_step": "complete",
"description": "Process decision"
},
"complete": {
"action": "complete_workflow",
"description": "Complete"
}
}
}
},
"input_data": {
"name": "New Team Member"
}
}
JSON
EOF

echo ""
echo "================================"
echo "MONITORING THE APPROVAL FLOW"
echo "================================"

echo ""
echo "1. Start notification listener:"
echo "kubectl -n kafka run -i --rm kcat-consumer-notify --image=edenhill/kcat:1.7.1 --restart=Never -- kcat -C -b personae-kafka-cluster-kafka-bootstrap:9092 -t system.notifications.ui -o json | jq '.'"

echo ""
echo "2. Start response listener:"
echo "kubectl -n kafka run -i --rm kcat-consumer-resp --image=edenhill/kcat:1.7.1 --restart=Never -- kcat -C -b personae-kafka-cluster-kafka-bootstrap:9092 -t system.agent.generic.responses -o json | jq '.'"

echo ""
echo "3. Check agent logs:"
echo "kubectl logs -n your-namespace deployment/agent-chassis --tail=100 | grep -E '(await_approval|process_approval|AwaitApprovalAction)'"

echo ""
echo "================================"
echo "COMMON ISSUES AND FIXES"
echo "================================"

echo ""
echo "Issue: 'ai_service configuration not found'"
echo "Fix: Use Option 1 (no LLM) or Option 2 (includes ai_service config)"

echo ""
echo "Issue: 'requires a topic'"
echo "Fix: Add 'topic: system.agent.generic.requests' to action steps"

echo ""
echo "Issue: Workflow doesn't pause"
echo "Fix: Check that await_approval returns 'await_response: true'"

echo ""
echo "Issue: Can't find approval token"
echo "Fix: Look in system.notifications.ui topic for 'request_id' field"