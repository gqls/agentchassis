#!/usr/bin/env bash
# =====================================================================
# idea.uk composition RE-RESOLVE — STEP 3 of 4: RE-TRIGGER (kcat)
#
# RUNBOOK_idea_uk_chassis_site_and_vm_deploy(21).md
#
# Re-enters the pipeline at COMPOSITION by orchestrating site-design-planner
# with the existing site. Its first step, ensure_site_record, loads the site
# by DOMAIN, so input_data MUST carry domain (site_id alone fails with
# "domain not found in input_data"). The normal cascade passes domain via the
# needs_composition work item; this matches that. Envelope identical to 082/080c.
#
# PREREQS (in order):
#   1. matcher code deployed: merged fork_theme_composition.go +
#      resolve_composition_layout_action.go built into the chassis image AND
#      site-design-planner rolled to it — else it re-runs the old scheme-blind
#      matcher and re-picks tool-portal-dark.
#   2. step 2 committed (composition detached/cleared) — else
#      install_site_composition refuses to overwrite.
#
# When the planner finishes install_site_composition it emits its needs_design
# handoff, which the build-dispatch-loop routes to webdesign-agent to re-render
# styles.css. If design does not advance, re-run with AGENT="webdesign-agent".
# =====================================================================
set -euo pipefail

SITE_ID="1244516d-014d-421c-88c6-090bb1e9552a"   # idea.uk
DOMAIN="idea.uk"
AGENT="site-design-planner"
INPUT_DATA="{\"site_id\":\"${SITE_ID}\",\"domain\":\"${DOMAIN}\"}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"

echo "Re-resolving composition for site ${SITE_ID} (${DOMAIN}) via ${AGENT}"
echo "  correlation_id=${CORRELATION_ID}"

kubectl -n kafka run -i --rm kcat-reresolve-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=reresolve-idea-uk-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"$AGENT"},"input_data":$INPUT_DATA}
JSON

echo ""
echo "=== Monitor (kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db) ==="
echo "  SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid;"
echo "  -- expect it to walk ensure_site_record -> validate_composition_inputs -> check_ready ->"
echo "  --   resolve_composition_layout -> resolve_composition_typography -> resolve_composition_palette ->"
echo "  --   install_site_composition -> complete, then run reresolve_idea_uk_04_verify.sql"
