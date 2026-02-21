#!/bin/bash
# ============================================================================
# BUILD: ai-agent-orchestration.com
# Full intake pipeline: initial message + 3 HITL responses
# ============================================================================

KAFKA_POD="personae-kafka-cluster-kafka-0"
KAFKA_NS="kafka"
BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
CLIENT_ID="demo_client"

# ============================================================================
# STEP 1: Initial intake message
# ============================================================================
step1_initial() {
  CORRELATION_ID=$(uuidgen)
  ORCHESTRATION_ID=$(uuidgen)
  MESSAGE_ID=$(uuidgen)
  REQUEST_ID=$(uuidgen)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  ORCHESTRATION_NAME="intake-aiao-$(date +%Y%m%d-%H%M%S)"

  echo "========================================="
  echo "STEP 1: Initial Intake - ai-agent-orchestration.com"
  echo "========================================="
  echo "  CORRELATION_ID=$CORRELATION_ID"
  echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
  echo "========================================="

  kubectl -n kafka run -i --rm kcat-aiao-init-$(date +%s) \
    --image=edenhill/kcat:1.7.1 \
    --restart=Never -- \
    kcat -P \
    -b $BOOTSTRAP \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_id=$MESSAGE_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H orchestration_name=$ORCHESTRATION_NAME \
    -H step_name=start \
    -H client_id=$CLIENT_ID \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=user \
    -H from_agent_id=cli \
    -H responses_topic=system.generic.responses <<'JSON'
{"action":"orchestrate","config":{"agent_type":"intake-orchestrator"},"input_data":{"domain":"ai-agent-orchestration.com","objective":"Build a technical but accessible website for ai-agent-orchestration.com — an open-source-flavoured platform and framework for building hierarchical, recursive AI agent systems. This is NOT a consulting site; it is a developer/technical product site for people who want to build and deploy their own multi-agent orchestration systems.\n\nThe site should convey technical depth and credibility while remaining approachable. Think: a well-designed open-source project site crossed with a SaaS product page.\n\nKey themes:\n• Hierarchical agent architectures — agents that spawn, coordinate, and manage sub-agents\n• Recursive task decomposition — complex problems broken into specialist sub-tasks automatically\n• Production-grade reliability — fault tolerance, monitoring, distributed execution, Kafka-based messaging\n• Developer-first — Go SDK, clean APIs, workflow DSL, easy local development\n• Real orchestration — not just prompt chaining, but proper workflow state machines with saga coordination\n\nCore site sections:\n• Home — hero with clear value prop, architecture overview visual, key capabilities\n• How It Works — visual walkthrough of agent spawning, message flow, workflow execution\n• Features — detailed feature cards: hierarchical orchestration, recursive decomposition, saga coordination, loop actions, LLM integration, human-in-the-loop, observability\n• Use Cases — real examples: automated website builders, document processing pipelines, research agents, data enrichment workflows\n• Architecture — technical deep-dive: Kafka messaging, stateless agents, workflow DSL, collected_data pattern\n• Documentation (link placeholder)\n• GitHub (link placeholder)\n• Get Started / Contact\n\nDo NOT invent fake statistics or testimonials. The product speaks for itself through its technical capabilities. Use concrete, specific descriptions of what the system actually does rather than marketing fluff.\n\nImagery: Think circuit-board patterns, network graphs, hierarchical tree diagrams, dark mode aesthetics with accent colours. Technical but beautiful.","model":"AIDA","hitl_mode":"interactive","repo_name":"sites","target_audience":"Software engineers, ML engineers, platform engineers, technical CTOs, and AI researchers who want to build production-grade multi-agent systems. People who are frustrated with brittle prompt chains and want proper orchestration infrastructure."}}
JSON

  echo ""
  echo "SAVE THESE:"
  echo "  CORRELATION_ID=$CORRELATION_ID"
  echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
  echo ""
  echo "NEXT: Wait ~10s, then find HITL request_id:"
  echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 | grep '$CORRELATION_ID' | grep hitl_confirm_type | grep 'Action requires waiting'"
}

# ============================================================================
# STEP 2: Confirm type (brochure / pageflow-builder)
# ============================================================================
step2_confirm_type() {
  if [ -z "$CORRELATION_ID" ] || [ -z "$ORCHESTRATION_ID" ] || [ -z "$HITL_REQUEST_ID" ]; then
    echo "Set CORRELATION_ID, ORCHESTRATION_ID, and HITL_REQUEST_ID first"
    return 1
  fi

  MESSAGE_ID=$(uuidgen)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  echo "========================================="
  echo "STEP 2: Confirm Type — brochure"
  echo "========================================="

  kubectl -n kafka run -i --rm kcat-aiao-confirm-$(date +%s) \
    --image=edenhill/kcat:1.7.1 \
    --restart=Never -- \
    kcat -P \
    -b $BOOTSTRAP \
    -t system.agent.generic.responses \
    -H correlation_id=$CORRELATION_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H message_id=$MESSAGE_ID \
    -H message_type=response \
    -H client_id=$CLIENT_ID \
    -H in_response_to_request_id=$HITL_REQUEST_ID \
    -H in_response_to_step_name=hitl_confirm_type \
    -H status=complete \
    -H sender_agent_type=human \
    -H sender_agent_id=cli-user \
    -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","message_id":"${MESSAGE_ID}","message_type":"response","client_id":"demo_client","in_response_to_request_id":"${HITL_REQUEST_ID}","in_response_to_step_name":"hitl_confirm_type","in_response_to_action":"request_human_input","status":"complete","is_complete":true,"is_error":false,"sender":{"agent_id":"cli-user","agent_type":"human","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"body":{"success":true,"human_response":true,"site_type":"brochure","recommended_builder":"pageflow-builder","status":"confirmed","message":"Site type confirmed by user"}}
JSON

  echo ""
  echo "NEXT: Find hitl_review_brief request_id:"
  echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 | grep '$CORRELATION_ID' | grep hitl_review_brief | grep 'Action requires waiting'"
}

# ============================================================================
# STEP 3: Review brief — the actual content brief
# ============================================================================
step3_review_brief() {
  if [ -z "$CORRELATION_ID" ] || [ -z "$ORCHESTRATION_ID" ] || [ -z "$HITL_REQUEST_ID" ]; then
    echo "Set CORRELATION_ID, ORCHESTRATION_ID, and HITL_REQUEST_ID first"
    return 1
  fi

  MESSAGE_ID=$(uuidgen)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  echo "========================================="
  echo "STEP 3: Review Brief — ai-agent-orchestration.com"
  echo "========================================="

  kubectl -n kafka run -i --rm kcat-aiao-brief-$(date +%s) \
    --image=edenhill/kcat:1.7.1 \
    --restart=Never -- \
    kcat -P \
    -b $BOOTSTRAP \
    -t system.agent.generic.responses \
    -H correlation_id=$CORRELATION_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H message_id=$MESSAGE_ID \
    -H message_type=response \
    -H client_id=$CLIENT_ID \
    -H in_response_to_request_id=$HITL_REQUEST_ID \
    -H in_response_to_step_name=hitl_review_brief \
    -H status=complete \
    -H sender_agent_type=human \
    -H sender_agent_id=cli-user \
    -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","message_id":"${MESSAGE_ID}","message_type":"response","client_id":"demo_client","in_response_to_request_id":"${HITL_REQUEST_ID}","in_response_to_step_name":"hitl_review_brief","in_response_to_action":"request_human_input","status":"complete","is_complete":true,"is_error":false,"sender":{"agent_id":"cli-user","agent_type":"human","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"body":{"success":true,"company_name":"AI Agent Orchestration","tagline":"Production-Grade Multi-Agent Systems. Built Right.","about_us":"AI Agent Orchestration is a framework for building hierarchical, recursive AI agent systems that actually work in production. Born from the real-world challenge of coordinating dozens of specialist AI agents to build complete websites, process documents, and run complex research pipelines — we built the orchestration layer we wished existed.\n\nUnlike prompt-chaining tools or simple agent wrappers, this is proper infrastructure: Kafka-based messaging, stateful saga coordination, automatic agent spawning, loop constructs, human-in-the-loop gates, and full observability. Agents don't just call each other — they spawn sub-agents, delegate tasks, collect results, and handle failures gracefully.\n\nThe framework runs on Kubernetes, communicates via Kafka, stores state in PostgreSQL, and treats every agent as both a worker and a potential orchestrator. It's the orchestration backbone behind systems that build multi-page websites from a single domain name, generate comprehensive research reports, and run data enrichment pipelines across thousands of records.","services":[{"name":"Hierarchical Agent Orchestration","description":"Build agent hierarchies where any agent can spawn and coordinate sub-agents. Parent agents delegate complex tasks, monitor progress, and collect results — creating natural decomposition of complex problems into manageable specialist work."},{"name":"Workflow Engine & Saga Coordination","description":"Define agent workflows as declarative step graphs with branching, looping, and error recovery. The saga coordinator manages distributed state, handles timeouts, supports retry with backoff, and ensures workflows complete reliably even when individual steps fail."},{"name":"Kafka-Native Messaging","description":"Every agent communicates through Kafka topics with structured message headers. Request-response patterns, fan-out broadcasts, and dedicated job topics for parent-child isolation. No HTTP spaghetti — clean, auditable, replayable message flows."},{"name":"Human-in-the-Loop Integration","description":"Built-in HITL gates for approval workflows, content review, and quality checkpoints. Agents pause execution, notify humans, and resume seamlessly when input arrives. Supports configurable timeouts and auto-escalation."},{"name":"LLM Integration & Tool Use","description":"Native integration with Claude, GPT, and other LLMs as action steps within workflows. Structured prompt templates with variable injection from collected data. LLM calls are just another action — same retry logic, same observability, same error handling."},{"name":"Observability & Debugging","description":"Correlation IDs trace every message across the entire agent tree. Structured JSON logging, orchestration state inspection, trace files per correlation, and step-by-step execution history. When something goes wrong, you can see exactly where and why."}],"target_audience":"Software engineers, ML engineers, platform engineers, and technical leaders building production AI agent systems","tone":"technical, direct, confident — show don't tell","key_differentiators":"Real orchestration infrastructure, not another prompt chain wrapper\nHierarchical spawning — agents create and manage sub-agents dynamically\nKafka-based messaging with full auditability and replay\nSaga coordination with distributed state management\nBattle-tested on real workloads: website builders, document processors, research pipelines\nEvery agent is an orchestrator — uniform architecture from top to bottom","leadership_team":[],"case_studies":[{"title":"Automated Multi-Page Website Builder","summary":"A complete website built from just a domain name. The intake orchestrator classifies the project, briefs a builder agent, which spawns specialist agents for site planning, design, content writing, image generation, HTML assembly, and deployment. Over 30 agents coordinate across a single build, producing a live Cloudflare-hosted site in under 30 minutes."},{"title":"Veterinary Practice Discovery Pipeline","summary":"A rolling pipeline that sweeps geographic areas, identifies potential veterinary practices from web data, scores candidates, and dispatches verification agents to confirm details. Processes thousands of leads with automatic retry, deduplication, and progressive enrichment across multiple runs."},{"title":"Section-Level Content Editing","summary":"A targeted editing system where a section-editor agent receives edit instructions, loads the right page component, applies content or component swaps, re-renders the affected section using stored templates, and commits the updated page to git — triggering automatic deployment. Surgical precision without rebuilding entire sites."}],"contact_email":"hello@ai-agent-orchestration.com","has_blog":true,"has_careers":false}}
JSON

  echo ""
  echo "NEXT: Wait for builder, then find escalation request_id:"
  echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=500 | grep '$CORRELATION_ID' | grep escalate_to_human"
}

# ============================================================================
# STEP 4: Approve auto-eval escalation
# ============================================================================
step4_approve_escalation() {
  if [ -z "$CORRELATION_ID" ] || [ -z "$ORCHESTRATION_ID" ] || [ -z "$HITL_REQUEST_ID" ]; then
    echo "Set CORRELATION_ID, ORCHESTRATION_ID, and HITL_REQUEST_ID first"
    return 1
  fi

  MESSAGE_ID=$(uuidgen)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  echo "========================================="
  echo "STEP 4: Approve Escalation"
  echo "========================================="

  kubectl -n kafka run -i --rm kcat-aiao-approve-$(date +%s) \
    --image=edenhill/kcat:1.7.1 \
    --restart=Never -- \
    kcat -P \
    -b $BOOTSTRAP \
    -t system.agent.generic.responses \
    -H correlation_id=$CORRELATION_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H message_id=$MESSAGE_ID \
    -H message_type=response \
    -H client_id=$CLIENT_ID \
    -H in_response_to_request_id=$HITL_REQUEST_ID \
    -H in_response_to_step_name=escalate_to_human \
    -H status=complete \
    -H sender_agent_type=human \
    -H sender_agent_id=cli-user \
    -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","message_id":"${MESSAGE_ID}","message_type":"response","client_id":"demo_client","in_response_to_request_id":"${HITL_REQUEST_ID}","in_response_to_step_name":"escalate_to_human","in_response_to_action":"request_human_input","status":"complete","is_complete":true,"is_error":false,"sender":{"agent_id":"cli-user","agent_type":"human","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"body":{"success":true,"approved":true,"status":"approved","responded_by":"cli-user@example.com","responded_at":"${TIMESTAMP}","review_mode":"escalated","edits":{},"comments":"Content reviewed and approved via CLI. Auto-eval issues have been addressed."}}
JSON

  echo "Escalation approved. Build should continue."
}

# ============================================================================
# Run the requested step
# ============================================================================
case "${1:-}" in
  1|init)     step1_initial ;;
  2|confirm)  step2_confirm_type ;;
  3|brief)    step3_review_brief ;;
  4|approve)  step4_approve_escalation ;;
  *)
    echo "Usage: $0 {1|2|3|4}"
    echo "  1/init    - Send initial intake message"
    echo "  2/confirm - Confirm site type (set CORRELATION_ID, ORCHESTRATION_ID, HITL_REQUEST_ID)"
    echo "  3/brief   - Send review brief"
    echo "  4/approve - Approve auto-eval escalation"
    ;;
esac