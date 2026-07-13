#!/bin/bash
# ============================================================================
# BUILD: finetuning.uk
# Full intake pipeline: initial message + 3 HITL responses
# ============================================================================

KAFKA_POD="personae-kafka-cluster-kafka-0"
KAFKA_NS="kafka"
BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
CLIENT_ID="demo_client"

# ============================================================================
# STEP 1: Initial intake message
# ============================================================================

  CORRELATION_ID=$(uuidgen)
  ORCHESTRATION_ID=$(uuidgen)
  MESSAGE_ID=$(uuidgen)
  REQUEST_ID=$(uuidgen)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  ORCHESTRATION_NAME="intake-finetuning-$(date +%Y%m%d-%H%M%S)"

  echo "========================================="
  echo "STEP 1: Initial Intake - finetuning.uk"
  echo "========================================="
  echo "  CORRELATION_ID=$CORRELATION_ID"
  echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
  echo "========================================="

  kubectl -n kafka run -i --rm kcat-ft-init-$(date +%s) \
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
{"action":"orchestrate","config":{"agent_type":"intake-orchestrator"},"input_data":{"domain":"finetuning.uk","objective":"Build a professional, warm, and credible corporate website for FineTuning — a UK-based AI consultancy that makes practical AI accessible to small and medium businesses. This is the parent company that operates Leopardess Consulting (automated website/tool builder) and the AI Agent Orchestration framework.\n\nFineTuning bridges the gap between cutting-edge AI and everyday business needs. We don't sell hype — we build things that work: custom fine-tuned models, intelligent automation systems, multi-agent orchestration platforms, and practical AI tools that solve real problems.\n\nKey themes:\n• AI for the rest of us — demystifying AI for SMEs who feel left behind by the AI revolution\n• Practical over theoretical — we build working systems, not slide decks\n• Full stack AI capability — from model fine-tuning to agent orchestration to end-user tools\n• UK-based, human-centred, honest about what AI can and can't do\n\nCore site sections:\n• Home — warm hero, clear positioning, overview of capabilities, trust signals\n• About — our story, philosophy (start small, build up), the team\n• Services — Custom AI Model Fine-Tuning, Agentic Systems Development, Automated Digital Solutions (via Leopardess), AI Strategy & Consulting\n• Our Products — showcase Leopardess Consulting and AI Agent Orchestration as products/platforms we've built and operate\n• Case Studies / Portfolio\n• Insights / Blog\n• Contact\n\nTone: Confident but not arrogant. Technical credibility without jargon overload. Friendly, approachable, British understatement. The kind of company you'd trust with your first AI project.\n\nImagery: Clean, modern, subtly techy. Think warm gradients, abstract neural network patterns, professional photography style. Not cold/corporate — more like a trusted advisor who happens to be technically brilliant.","model":"AIDA","hitl_mode":"interactive","repo_name":"sites","target_audience":"SME owners, managing directors, operations managers, and department heads at UK businesses (10-500 employees) who know they should be doing something with AI but don't know where to start. Also: technical founders who need AI infrastructure but don't want to build it from scratch."}}
JSON

  echo ""
  echo "SAVE THESE:"
  echo "CORRELATION_ID=$CORRELATION_ID"
  echo "ORCHESTRATION_ID=$ORCHESTRATION_ID"
  echo ""
  echo "NEXT: Wait ~10s, then find HITL request_id:"
  echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 | grep '$CORRELATION_ID' | grep hitl_confirm_type | grep 'Action requires waiting'"


# ============================================================================
# STEP 2: Confirm type (brochure / pageflow-builder)
# ============================================================================

CORRELATION_ID=08b6fd43-7ada-430f-ac8f-eabd5f12adca
ORCHESTRATION_ID=ba4e25cd-58d3-41a7-8be7-e3ac0216de24
HITL_REQUEST_ID=74a00d21-ad45-4876-a1d5-73f640e5b88a

  MESSAGE_ID=$(uuidgen)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  echo "========================================="
  echo "STEP 2: Confirm Type — brochure"
  echo "========================================="

  kubectl -n kafka run -i --rm kcat-ft-confirm-$(date +%s) \
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





# ============================================================================
# STEP 3: Review brief — the actual content brief
# ============================================================================

CORRELATION_ID=08b6fd43-7ada-430f-ac8f-eabd5f12adca
ORCHESTRATION_ID=ba4e25cd-58d3-41a7-8be7-e3ac0216de24
HITL_REQUEST_ID=892b215b-fd68-4a50-a772-01f3db779ee4

  MESSAGE_ID=$(uuidgen)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  echo "========================================="
  echo "STEP 3: Review Brief — finetuning.uk"
  echo "========================================="

  kubectl -n kafka run -i --rm kcat-ft-brief-$(date +%s) \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","message_id":"${MESSAGE_ID}","message_type":"response","client_id":"demo_client","in_response_to_request_id":"${HITL_REQUEST_ID}","in_response_to_step_name":"hitl_review_brief","in_response_to_action":"request_human_input","status":"complete","is_complete":true,"is_error":false,"sender":{"agent_id":"cli-user","agent_type":"human","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"body":{"success":true,"company_name":"FineTuning","tagline":"AI for the Rest of Us","about_us":"FineTuning is a UK-based AI consultancy that builds practical, working AI systems for small and medium businesses. We were founded on a simple observation: the AI revolution is moving fast, and most businesses are being left behind — not because the technology isn't ready, but because nobody is meeting them where they are.\n\nWe start small and build up. A fine-tuned model that actually understands your industry terminology. An automation that saves your team three hours a day. A set of AI agents that handle your repetitive workflows while your people focus on the work that matters.\n\nWe built Leopardess Consulting to demonstrate what's possible — AI agent teams that create complete websites, tools, and reports in minutes. We built the AI Agent Orchestration framework because we needed proper infrastructure for coordinating dozens of specialist AI agents on complex tasks. These aren't theoretical products; they're systems we use every day.\n\nBased in London, we work with businesses across the UK. We're honest about what AI can do well and where it still struggles. We'd rather build you something modest that works brilliantly than promise the moon and deliver a PowerPoint.","services":[{"name":"Custom AI Model Fine-Tuning","description":"We take foundation models and make them yours. Fine-tuning on your data, your terminology, your domain — so the AI actually understands your business. We handle data assessment, model selection, training, validation, and deployment. The result is a model that performs dramatically better on your specific tasks than any off-the-shelf solution."},{"name":"Agentic Systems Development","description":"Intelligent automation that goes beyond simple scripts. We build AI agent systems that can make decisions, handle exceptions, learn from outcomes, and coordinate complex multi-step workflows. From customer service automation to document processing pipelines to inventory management — agents that work reliably without constant supervision."},{"name":"Automated Digital Solutions","description":"Through our Leopardess Consulting platform, we offer AI-powered creation of websites, custom tools, business reports, calculators, planners, and more — all delivered in minutes rather than weeks. Our multi-agent teams handle planning, design, content, images, and deployment end-to-end."},{"name":"AI Strategy & Consulting","description":"Not sure where to start with AI? We help businesses identify the practical opportunities — the tasks where AI will genuinely save time and money right now. We try not to use jargon. We are realistic about AI timelines. We give clear advice on what to build first and how to build it properly."}],"target_audience":"UK SMEs, managing directors, operations managers, and technical founders who want practical AI solutions","tone":"warm, confident, British, understated, honest","key_differentiators":"We actually build and operate AI systems ourselves — Leopardess and AI Agent Orchestration are our own products\nWe start small and prove value before scaling up\nHonest about AI capabilities and limitations\nFull-stack: from model fine-tuning to agent orchestration to end-user tools\nUK-based with a focus on SMEs who are underserved by the AI industry\nPractical results over theoretical potential","leadership_team":[{"name":"Peter Grenfell","title":"Founder & Principal Consultant","bio":"Peter is a hands-on technologist who has spent the last two years building production AI agent systems from the ground up. He founded FineTuning to bring practical AI capabilities to the businesses that need them most — the SMEs that drive the UK economy but are often overlooked by the AI industry. Previously, Peter built distributed systems and led technology teams across financial services and consulting."}],"case_studies":[{"title":"Leopardess Consulting — AI-Powered Website Builder","summary":"We built an AI agent platform that creates complete, professional websites from just a domain name. Over 30 specialist agents coordinate to handle planning, design, content writing, image generation, and deployment. Sites go from brief to live on Cloudflare in under 30 minutes. Now available as a service to businesses that need quality digital presence without the traditional agency costs and timelines."},{"title":"AI Agent Orchestration Framework","summary":"We developed a production-grade framework for building hierarchical AI agent systems. Kafka-based messaging, saga coordination, recursive agent spawning, and full observability. Originally built to power our own products, now available as infrastructure for teams building their own multi-agent systems."},{"title":"Document Processing Pipeline for a Professional Services Firm","summary":"A mid-size consultancy needed to extract, classify, and summarise information from hundreds of client documents monthly. We built an agent pipeline that processes incoming documents, extracts key data, cross-references against existing records, and produces structured summaries — reducing a three-day manual process to under an hour."}],"contact_email":"finetuning@contactforsales.com","contact_phone":"+44 (0) 7934 524 911","headquarters":"London, UK","has_blog":true,"has_careers":true}}
JSON

  echo ""
  echo "NEXT: Wait for builder, then find escalation request_id:"
  echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=500 | grep '$CORRELATION_ID' | grep escalate_to_human"






# ============================================================================
# STEP 4: Approve auto-eval escalation
# ============================================================================

CORRELATION_ID=08b6fd43-7ada-430f-ac8f-eabd5f12adca
ORCHESTRATION_ID=f6904935-0957-4a95-9df8-3aa561e6aad1
HITL_REQUEST_ID=9ee499e2-c8ac-4acc-9658-af8efa2bebbf
RESPONSES_TOPIC=job.08b6fd43-d4523661-content-reviewer-spawn_reviewer.responses

  MESSAGE_ID=$(uuidgen)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  echo "========================================="
  echo "STEP 4: Approve Escalation"
  echo "========================================="

  kubectl -n kafka run -i --rm kcat-ft-approve-$(date +%s) \
    --image=edenhill/kcat:1.7.1 \
    --restart=Never -- \
    kcat -P \
    -b $BOOTSTRAP \
    -t $RESPONSES_TOPIC \
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


