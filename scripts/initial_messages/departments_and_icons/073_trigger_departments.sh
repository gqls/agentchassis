#!/bin/bash
# ============================================================================
# DEPARTMENTS DEPLOYMENT — Trigger Script
# ============================================================================
#
# Prerequisites (must be complete before running):
#   1. Go: ImagePurposes["icon"] = {240, 240, 85, "jpg"} in image_processing.go
#   2. Go: ActionInputSpec + deploy_path support in deploy_image_asset_action.go
#   3. Go: Chassis rebuilt and deployed
#   4. SQL: departments-grid component in content_components
#   5. SQL: asset-deployer in agent_definitions (016_asset_deployer_corrected.sql)
#   6. API: Topics created via POST .../asset-deployer/topics/recreate
#   7. 7 full-size department icon images uploaded to S3
#
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


# ============================================================================
# FILL THESE IN
# ============================================================================

# Site ID — get from:
#   kubectl -n ai-persona-system exec postgres-clients-0 -- \
#     psql -U clients_user -d clients_db -t -A -c \
#     "SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk' LIMIT 1;"
SITE_ID="PASTE_SITE_ID_HERE"

# S3 URIs — from the upload step
# Mapping: octopus→orchestration, clock→development, quillpen→design,
#          languagetablet→content, spiderandweb2→seo-marketing, tickertape→analytics
S3_ORCHESTRATION="dept-orchestration.jpg"
S3_DEVELOPMENT="dept-development.jpg"
S3_DESIGN="dept-design.jpg"
S3_CONTENT="dept-content.jpg"
S3_SEO="dept-seo-marketing.jpg"
S3_ANALYTICS="dept-analytics.jpg"
S3_STRATEGY="dept-strategy.jpg"


# ============================================================================
# 3a. DEPLOY IMAGES — 7 calls to asset-deployer
# ============================================================================

TOPIC="system.agent.asset-deployer.process"

deploy_icon() {
  local NAME="$1"
  local S3_URI="$2"
  local DEPLOY_PATH="$3"

  local CORR_ID=$(uuidgen)
  local REQ_ID=$(uuidgen)
  local MSG_ID=$(uuidgen)

  local HEADERS="correlation_id:${CORR_ID},request_id:${REQ_ID},message_id:${MSG_ID},client_id:${CLIENT_ID},message_type:request,action:orchestrate,from_agent_type:manual-trigger,responses_topic:system.agent.manual.responses"

  local BODY="{\"action\":\"orchestrate\",\"input_data\":{\"domain\":\"${DOMAIN}\",\"s3_uri\":\"${S3_URI}\",\"deploy_path\":\"${DEPLOY_PATH}\",\"purpose\":\"icon\"}}"

  echo "  ${NAME}: ${DEPLOY_PATH}"
  echo "    corr: ${CORR_ID}"
  kafka_send "${TOPIC}" "${HEADERS}" "${BODY}"
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
echo "  Logs:    kubectl -n ai-persona-system logs -l agent-type=asset-deployer --tail=50"
echo ""
echo "  Wait for all 7 to complete, then press Enter..."
read -r


# ============================================================================
# 3b. SECTION EDITOR — swap leadership-team → departments-grid
# ============================================================================

echo "=== 3b: Component swap: leadership-team → departments-grid ==="

SE_CORR_ID=$(uuidgen)
SE_REQ_ID=$(uuidgen)
SE_MSG_ID=$(uuidgen)

SE_TOPIC="system.agent.section-editor.process"
SE_HEADERS="correlation_id:${SE_CORR_ID},request_id:${SE_REQ_ID},message_id:${SE_MSG_ID},client_id:${CLIENT_ID},message_type:request,action:orchestrate,from_agent_type:manual-trigger,responses_topic:system.agent.manual.responses"

SE_BODY=$(cat <<'ENDJSON'
{"action":"orchestrate","input_data":{"domain":"leopardessconsulting.co.uk","page_name":"about","slot_name":"leadership-team","edit_type":"component_swap","new_component_function":"departments-grid","replacement_content_data":{"section_title":"Our AI Departments","section_intro":"Leopardess runs on specialist AI agent teams, each focused on a different discipline. They collaborate automatically to plan, build, and deliver your projects — coordinated by our orchestration layer so everything comes together seamlessly.","departments":[{"name":"Strategy","subtitle":"Planning & Direction","description":"Every project starts here. Our strategy agents analyse your domain, audience, and objectives to create a structured build plan — choosing the right site type, page structure, and content approach before any building begins.","icon":"/assets/images/departments/dept-strategy.jpg"},{"name":"Orchestration","subtitle":"Coordination & Workflow","description":"The nerve centre of every project. Our orchestration agents manage task sequencing, delegate work to specialist teams, and ensure all the moving parts come together on time and in the right order.","icon":"/assets/images/departments/dept-orchestration.jpg"},{"name":"Development","subtitle":"Engineering & Build","description":"Our development agents handle the technical build — assembling pages from component libraries, wiring up functionality, deploying to hosting infrastructure, and making sure everything works across devices.","icon":"/assets/images/departments/dept-development.jpg"},{"name":"Design","subtitle":"Visual & UX","description":"Design agents select colour palettes, typography, and layout patterns matched to your industry and brand. They create cohesive visual systems that look polished and professional from the first page to the last.","icon":"/assets/images/departments/dept-design.jpg"},{"name":"Content","subtitle":"Copywriting & Messaging","description":"Content agents research your industry and write targeted copy for every section — headlines, service descriptions, calls to action, and supporting text that speaks directly to your audience.","icon":"/assets/images/departments/dept-content.jpg"},{"name":"SEO & Marketing","subtitle":"Search & Visibility","description":"Our SEO agents weave search optimisation into every page — meta tags, heading structure, keyword placement, and semantic markup — so your site is discoverable from the moment it goes live.","icon":"/assets/images/departments/dept-seo-marketing.jpg"},{"name":"Analytics & Reporting","subtitle":"Data & Insights","description":"Analytics agents generate structured reports, market research, and business intelligence documents. They process data into clear, actionable outputs — charts, summaries, and recommendations you can use immediately.","icon":"/assets/images/departments/dept-analytics.jpg"}]}}}
ENDJSON
)

echo "  corr: ${SE_CORR_ID}"
kafka_send "${SE_TOPIC}" "${SE_HEADERS}" "${SE_BODY}"

echo ""
echo "  Monitor: kubectl -n ai-persona-system get jobs | grep section-editor"
echo "  Logs:    kubectl -n ai-persona-system logs -l agent-type=section-editor --tail=100"
echo ""
echo "=== Done ==="


