original calculator workflow to just use the calculate action:

{
"workflow": {
"start_step": "process",
"steps": {
"process": {
"action": "calculate",
"next_step": "complete",
"description": "Perform calculation via LLM"
},
"complete": {
"action": "complete_workflow",
"description": "Complete calculation"
}
}
},
"ai_service": {
"model": "claude-3-haiku-20240307",
"provider": "anthropic"
},
"processing_mode": "task",
"prompt_template": "You are a calculator. Perform the operation '{{.input_data.operation}}' on the operands {{.input_data.operands}}. Respond ONLY with the final numerical result as a JSON object like {\"result\": 4}."
}


new one using llm

UPDATE agent_definitions
SET default_config = $$
{
"workflow": {
"start_step": "process",
"steps": {
"process": {
"action": "execute_llm_prompt",
"next_step": "complete",
"description": "Perform calculation via LLM"
},
"complete": {
"action": "complete_workflow",
"description": "Complete calculation"
}
}
},
"ai_service": {
"model": "claude-3-haiku-20240307",
"provider": "anthropic"
},
"processing_mode": "task",
"prompt_template": "You are a calculator. Perform the operation '{{.input_data.operation}}' on the operands {{.input_data.operands}}. Respond ONLY with the final numerical result as a JSON object like {\"result\": 4}."
}
$$::jsonb
WHERE id = 'afbefbd4-2934-4131-9082-abac236eaa49';

changing input data path

UPDATE agent_definitions
SET default_config = $$
{
"workflow": {
"start_step": "process",
"steps": {
"process": {
"action": "execute_llm_prompt",
"next_step": "complete",
"description": "Perform calculation via LLM"
},
"complete": {
"action": "complete_workflow",
"description": "Complete calculation"
}
}
},
"ai_service": {
"model": "claude-3-haiku-20240307",
"provider": "anthropic"
},
"processing_mode": "task",
"prompt_template": "You are a calculator. Perform the operation '{{.input_data.data.operation}}' on the operands {{.input_data.data.operands}}. Respond ONLY with the final numerical result as a JSON object like {\"result\": 4}."
}
$$::jsonb
WHERE id = 'afbefbd4-2934-4131-9082-abac236eaa49';