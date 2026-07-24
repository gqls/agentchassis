#!/usr/bin/env bash
# regenerate_orphan_component.sh <component_id> <label>
#
# Regenerate a truncated component that has NO page placement (bugs_open/046's
# three orphans). tool-improver cannot be used: its load_tool step JOINs
# page_components/pages, so an unplaced component resolves to not_found. This
# drive runs the SAME pipeline inline (query component -> LLM whole-rewrite ->
# update_component_html), minus the page join and minus delivery (there is no
# page to deliver to). update_component_html keeps the component write guard +
# create_version in the loop, exactly as tool-improver's update_component step.
# LLM settings mirror the live improve_tool step (claude-sonnet-5 @ 32000).
set -euo pipefail
COMPONENT_ID="${1:?component_id}"; LABEL="${2:-orphan}"

ISSUE="This component's html_template is a TRUNCATED LLM generation: a <script> block is opened but never closed - the original completion was cut mid-stream, so the JavaScript never runs and any markup after the cut would be swallowed as script text. Rewrite it as a COMPLETE self-contained fragment: keep the design, controls and behaviour already visible in the current HTML (the markup and CSS are largely intact; the JavaScript is cut partway through), complete the JavaScript so every control declared in the markup actually works end-to-end, and terminate every tag - every <script>, <style>, <section>, <div> and <fieldset> you open must be closed. The fragment must end on a closing tag. Do not invent external data sources or embed fabricated datasets; the component is fully self-contained."

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "CORRELATION_ID=$CORRELATION_ID"

kubectl -n kafka run -i --rm "kcat-orphan-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "request_id=$REQUEST_ID" -H "message_id=$MESSAGE_ID" \
  -H "message_type=request" -H "client_id=demo_client" -H "action=process" \
  -H "sender_agent_type=cli" -H "sender_agent_id=cli-user" \
  -H "responses_topic=system.agent.generic.responses" -H "timestamp=$TIMESTAMP" <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"load_component","processing_mode":"orchestrator","timeout_seconds":700,"steps":{"load_component":{"action":"query_database","config":{"query":"SELECT cc.id::text as component_id, cc.function, cc.display_name, cc.html_template, cc.description, cc.component_level FROM content_components cc WHERE cc.id = \$1::uuid AND cc.is_active = true LIMIT 1","params":["input_data.component_id"],"output_format":"object"},"output_field":"tool_data","next_step":"improve_tool","description":"Load the unplaced component (no page join - it has no placement)"},"improve_tool":{"action":"execute_llm_prompt","config":{"ai_service":{"model":"claude-sonnet-5","provider":"anthropic","max_tokens":32000,"api_key_env_var":"ANTHROPIC_API_KEY"},"input_fields":["input_data","tool_data"],"output_format":"text","prompt_template":"You are repairing an interactive component on a website.\n\n## Issue to Fix\n{{.input_data.issue}}\n\n## Component Info\nName: {{.tool_data.display_name}}\nFunction: {{.tool_data.function}}\nLevel: {{.tool_data.component_level}}\n\n## Current HTML\n{{.tool_data.html_template}}\n\n{{if .tool_data.description}}## Component Description\n{{.tool_data.description}}{{end}}\n\n## Rules\n1. Fix the specific issue described above\n2. Keep the core functionality intact\n3. Use CSS custom properties (var(--color-primary) etc) for colours - never hardcode hex values\n4. Ensure it works on mobile (min-width considerations, touch targets)\n5. Keep all interactive JavaScript working\n6. The output replaces html_template - it must be a complete, self-contained HTML fragment\n7. Include inline <style> for layout-only CSS (colours come from the site stylesheet)\n8. Include inline <script> for any JavaScript\n9. Do not add external dependencies\n10. If a tool-doc header comment block (/* === tool-doc === ... === /tool-doc === */) is present at the top of the <script>, preserve and update it. Never use */ inside it.\n\nOutput ONLY the repaired HTML fragment. No markdown fences. No explanation. Start directly with the HTML."},"output_field":"improved_html","next_step":"update_component","description":"LLM completes the truncated component"},"update_component":{"action":"update_component_html","config":{"component_id":"tool_data.component_id","html_field":"improved_html.result","create_version":true},"output_field":"update_result","next_step":"complete","description":"Guarded write back to content_components (write guard + version snapshot)"},"complete":{"action":"complete_workflow","config":{"output_fields":["update_result"]},"description":"done"}}}},"input_data":{"component_id":"${COMPONENT_ID}","issue":"${ISSUE}","label":"${LABEL}"}}
JSON
echo ""
echo "Watch: SELECT status,current_step FROM orchestration_states WHERE correlation_id='$CORRELATION_ID'::uuid ORDER BY updated_at DESC;"
