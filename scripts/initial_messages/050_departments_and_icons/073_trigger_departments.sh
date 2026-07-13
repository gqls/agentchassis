#!/bin/bash
# ============================================================================
# DEPARTMENTS DEPLOYMENT — Trigger Script
# ============================================================================
# Sends to generic agent which spawns+calls specialist agents.
# Same pattern as 073_section_editor.sh.
# ============================================================================
set -euo pipefail

KAFKA_POD="personae-kafka-cluster-kafka-0"
KAFKA_NS="kafka"
BOOTSTRAP="localhost:9092"
CLIENT_ID="demo_client"
DOMAIN="leopardessconsulting.co.uk"

kafka_send() {
  local TOPIC="$1"
  local HEADERS="$2"
  local BODY="$3"

  echo "${HEADERS}"$'\t'"${BODY}" | \
    kubectl -n ${KAFKA_NS} exec -i ${KAFKA_POD} -- \
    /opt/kafka/bin/kafka-console-producer.sh \
    --bootstrap-server ${BOOTSTRAP} \
    --topic "${TOPIC}" \
    --property parse.headers=true \
    --property headers.delimiter=$'\t'
}

# S3 URIs — update with actual filenames/URIs after upload
S3_STRATEGY="s3://personae-prod-uk001-images/images/demo_client/leopardess/departments/dept-strategy.jpg"
S3_ORCHESTRATION="s3://personae-prod-uk001-images/images/demo_client/leopardess/departments/dept-orchestration.jpg"
S3_DEVELOPMENT="s3://personae-prod-uk001-images/images/demo_client/leopardess/departments/dept-development.jpg"
S3_DESIGN="s3://personae-prod-uk001-images/images/demo_client/leopardess/departments/dept-design.jpg"
S3_CONTENT="s3://personae-prod-uk001-images/images/demo_client/leopardess/departments/dept-content.jpg"
S3_SEO="s3://personae-prod-uk001-images/images/demo_client/leopardess/departments/dept-seo-marketing.jpg"
S3_ANALYTICS="s3://personae-prod-uk001-images/images/demo_client/leopardess/departments/dept-analytics.jpg"

# ============================================================================
# 3a. DEPLOY IMAGES — 7 calls to asset-deployer via generic
# ============================================================================

ASSET_WORKFLOW=$(cat <<'ENDWF'
{
  "start_step": "spawn_deployer",
  "processing_mode": "orchestrator",
  "timeout_seconds": 120,
  "steps": {
    "spawn_deployer": {
      "action": "spawn_agent",
      "config": {
        "role": "asset_deployer",
        "agent_type": "asset-deployer"
      },
      "output_field": "deployer_agent",
      "next_step": "call_deployer",
      "description": "Spawn asset-deployer agent"
    },
    "call_deployer": {
      "action": "call_agent",
      "config": {
        "agent_type": "asset-deployer",
        "target_role": "asset_deployer",
        "input_mapping": {
          "domain": "input_data.domain",
          "s3_uri": "input_data.s3_uri",
          "deploy_path": "input_data.deploy_path",
          "purpose": "input_data.purpose"
        },
        "timeout_seconds": 90
      },
      "output_field": "deploy_result",
      "next_step": "complete",
      "description": "Deploy image asset"
    },
    "complete": {
      "action": "complete_workflow",
      "config": { "output_fields": ["deploy_result"] },
      "description": "Asset deploy complete"
    }
  }
}
ENDWF
)
ASSET_WORKFLOW_COMPACT=$(echo "$ASSET_WORKFLOW" | tr -d '\n' | sed 's/  */ /g')

deploy_icon() {
  local NAME="$1"
  local S3_URI="$2"
  local DEPLOY_PATH="$3"

  local CORR_ID=$(uuidgen)
  local ORCH_ID=$(uuidgen)
  local REQ_ID=$(uuidgen)
  local MSG_ID=$(uuidgen)
  local TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  local INPUT_DATA="{\"domain\":\"${DOMAIN}\",\"s3_uri\":\"${S3_URI}\",\"deploy_path\":\"${DEPLOY_PATH}\",\"purpose\":\"icon\"}"

  local MESSAGE_BODY=$(cat <<ENDMSG
{"headers":{"correlation_id":"${CORR_ID}","orchestration_id":"${ORCH_ID}","request_id":"${REQ_ID}","message_id":"${MSG_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":${ASSET_WORKFLOW_COMPACT}},"input_data":${INPUT_DATA}}
ENDMSG
)

  echo "  ${NAME}: ${DEPLOY_PATH} (${CORR_ID})"

  kubectl -n kafka run -i --rm kcat-asset-${NAME}-$(date +%s) \
    --image=edenhill/kcat:1.7.1 \
    --restart=Never -- \
    kcat -P \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORR_ID \
    -H orchestration_id=$ORCH_ID \
    -H request_id=$REQ_ID \
    -H message_id=$MSG_ID \
    -H message_type=request \
    -H client_id=$CLIENT_ID \
    -H action=process \
    -H sender_agent_type=cli \
    -H sender_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses \
    -H timestamp=$TIMESTAMP <<ENDKAFKA
${MESSAGE_BODY}
ENDKAFKA

  sleep 2
}

echo "=== 3a: Deploying 7 department icons ==="
deploy_icon "strategy"      "${S3_STRATEGY}"      "assets/images/departments/dept-strategy.jpg"
deploy_icon "orchestration" "${S3_ORCHESTRATION}" "assets/images/departments/dept-orchestration.jpg"
deploy_icon "development"   "${S3_DEVELOPMENT}"   "assets/images/departments/dept-development.jpg"
deploy_icon "design"        "${S3_DESIGN}"         "assets/images/departments/dept-design.jpg"
deploy_icon "content"       "${S3_CONTENT}"        "assets/images/departments/dept-content.jpg"
deploy_icon "seo-marketing" "${S3_SEO}"            "assets/images/departments/dept-seo-marketing.jpg"
deploy_icon "analytics"     "${S3_ANALYTICS}"      "assets/images/departments/dept-analytics.jpg"

echo ""
echo "  Monitor: kubectl -n ai-persona-system get jobs | grep asset-deployer"
echo "  Wait for all 7 to complete, then press Enter..."
read -r

# ============================================================================
# 3b. SECTION EDITOR — swap leadership-team -> departments-grid
# ============================================================================
# Uses replacement_content_data (not content_data) to avoid collision with
# site_record.content_data — see checklist "Avoid Field Names That Collide".
# Caller sends content_data in input_data; the input_mapping renames it to
# replacement_content_data for the child (matching 073_section_editor.sh).
# ============================================================================
echo "=== 3b: Component swap: leadership-team -> departments-grid ==="

SE_CORR_ID=$(uuidgen)
SE_ORCH_ID=$(uuidgen)
SE_REQ_ID=$(uuidgen)
SE_MSG_ID=$(uuidgen)
SE_TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# --- Departments data (readable, compacted before send) ---
DEPARTMENTS_JSON=$(cat <<'EOF'
[
  {
    "name": "Strategy",
    "subtitle": "Planning & Direction",
    "description": "Every project starts here. Our strategy agents analyse your domain, audience, and objectives to create a structured build plan.",
    "icon": "/assets/images/departments/dept-strategy.jpg"
  },
  {
    "name": "Orchestration",
    "subtitle": "Coordination & Workflow",
    "description": "The nerve centre of every project. Our orchestration agents manage task sequencing and delegate work to specialist teams.",
    "icon": "/assets/images/departments/dept-orchestration.jpg"
  },
  {
    "name": "Development",
    "subtitle": "Engineering & Build",
    "description": "Our development agents handle the technical build — assembling pages, wiring up functionality, and deploying to infrastructure.",
    "icon": "/assets/images/departments/dept-development.jpg"
  },
  {
    "name": "Design",
    "subtitle": "Visual & UX",
    "description": "Design agents select colour palettes, typography, and layout patterns matched to your industry and brand.",
    "icon": "/assets/images/departments/dept-design.jpg"
  },
  {
    "name": "Content",
    "subtitle": "Copywriting & Messaging",
    "description": "Content agents research your industry and write targeted copy for every section of your site.",
    "icon": "/assets/images/departments/dept-content.jpg"
  },
  {
    "name": "SEO & Marketing",
    "subtitle": "Search & Visibility",
    "description": "Our SEO agents weave search optimisation into every page — meta tags, heading structure, and semantic markup.",
    "icon": "/assets/images/departments/dept-seo-marketing.jpg"
  },
  {
    "name": "Analytics & Reporting",
    "subtitle": "Data & Insights",
    "description": "Analytics agents generate structured reports, market research, and business intelligence documents.",
    "icon": "/assets/images/departments/dept-analytics.jpg"
  }
]
EOF
)
DEPARTMENTS_COMPACT=$(echo "$DEPARTMENTS_JSON" | tr -d '\n' | sed 's/  */ /g')

# --- input_data for section editor ---
# Caller sends content_data; the input_mapping renames it to
# replacement_content_data for the child (same as 073_section_editor.sh)
SE_INPUT=$(cat <<EOF
{
  "domain": "${DOMAIN}",
  "page_name": "about",
  "slot_name": "departments-grid",
  "edit_type": "component_swap",
  "new_component_function": "departments-grid",
  "content_data": {
    "section_title": "Our AI Departments",
    "section_intro": "Leopardess runs on specialist AI agent teams, each focused on a different discipline. They collaborate automatically to plan, build, and deliver your projects.",
    "departments": ${DEPARTMENTS_COMPACT}
  }
}
EOF
)
SE_INPUT_COMPACT=$(echo "$SE_INPUT" | tr -d '\n' | sed 's/  */ /g')

# --- Spawn+call workflow (same as 073_section_editor.sh) ---
SE_WORKFLOW=$(cat <<'ENDWF'
{
  "start_step": "spawn_section_editor",
  "processing_mode": "orchestrator",
  "timeout_seconds": 900,
  "steps": {
    "spawn_section_editor": {
      "action": "spawn_agent",
      "config": {
        "role": "section_editor",
        "agent_type": "section-editor"
      },
      "output_field": "section_editor_agent",
      "next_step": "call_section_editor",
      "description": "Spawn section-editor agent"
    },
    "call_section_editor": {
      "action": "call_agent",
      "config": {
        "agent_type": "section-editor",
        "target_role": "section_editor",
        "input_mapping": {
          "domain": "input_data.domain",
          "edit_type": "input_data.edit_type",
          "page_name?": "input_data.page_name",
          "slot_name?": "input_data.slot_name",
          "field_updates?": "input_data.field_updates",
          "replacement_content_data?": "input_data.content_data",
          "new_component_function?": "input_data.new_component_function",
          "page_component_id?": "input_data.page_component_id"
        },
        "timeout_seconds": 600
      },
      "output_field": "edit_result",
      "next_step": "complete",
      "description": "Run section edit"
    },
    "complete": {
      "action": "complete_workflow",
      "config": { "output_fields": ["edit_result"] },
      "description": "Section edit complete"
    }
  }
}
ENDWF
)
SE_WORKFLOW_COMPACT=$(echo "$SE_WORKFLOW" | tr -d '\n' | sed 's/  */ /g')

# --- Build full message body ---
SE_MESSAGE_BODY=$(cat <<ENDMSG
{"headers":{"correlation_id":"${SE_CORR_ID}","orchestration_id":"${SE_ORCH_ID}","request_id":"${SE_REQ_ID}","message_id":"${SE_MSG_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${SE_TIMESTAMP}"},"config":{"workflow":${SE_WORKFLOW_COMPACT}},"input_data":${SE_INPUT_COMPACT}}
ENDMSG
)

echo "  corr: ${SE_CORR_ID}"

kubectl -n kafka run -i --rm kcat-dept-swap-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$SE_CORR_ID \
  -H orchestration_id=$SE_ORCH_ID \
  -H request_id=$SE_REQ_ID \
  -H message_id=$SE_MSG_ID \
  -H message_type=request \
  -H client_id=$CLIENT_ID \
  -H action=process \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$SE_TIMESTAMP <<ENDKAFKA
${SE_MESSAGE_BODY}
ENDKAFKA

echo ""
echo "  Monitor: kubectl -n ai-persona-system get jobs | grep section-editor"
echo "  CORRELATION_ID=$SE_CORR_ID"
echo "=== Done ==="