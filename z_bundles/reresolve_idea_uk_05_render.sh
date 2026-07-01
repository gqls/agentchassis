#!/usr/bin/env bash
# =====================================================================
# idea.uk composition RE-RESOLVE — STEP 5 of 5: RENDER + DEPLOY styles.css
#
# WHY THIS STEP EXISTS: orchestrating site-design-planner standalone (step 3)
# INSTALLS the composition (its workflow ends install_site_composition ->
# complete) but does NOT emit the needs_design handoff the full build cascade
# emits, so nothing re-rendered styles.css. Confirmed: after step 3 the only
# site_work_items were the 8 original page_rerender rows (2026-06-21); the
# re-resolve created zero new work items. This step renders + deploys the CSS.
#
# WHAT IT DOES: orchestrates webdesign-agent (processing_mode task, directly
# orchestratable; input_schema optional [site_id, domain, site_context]) with
# {site_id, domain}. Its workflow:
#   load_site_context (include_style_collection -> picks up the NEW
#     tool-portal-light composition) -> read_site_specs -> analyze_design
#     (LLM, claude-sonnet-4-6; reads design_intent parchment as STARTING POINTS,
#     may adjust) -> update_site (stores design_spec) -> generate_css
#     (render_css_from_spec, deterministic Go template) -> deploy_css
#     (git_commit assets/css/styles.css -> fires GH Actions -> B2) ->
#     spawn/call asset_renderer (snippets.js) -> check_should_fork -> complete.
#
# IMPORTANT:
#   * input_data does NOT set should_fork_theme, so check_should_fork -> complete
#     (no fork; the installed composition is left intact).
#   * analyze_design is an LLM step -> the deployed palette may drift slightly
#     from the exact parchment in the palettes row. Compare after (verify below).
#   * it renders styles.css + snippets.js, NOT page HTML. tool-portal-light
#     shares tool-portal-dark's class contract, so the existing pages should
#     re-skin from the new CSS. If they look wrong, a page_rerender may be
#     needed -- assess visually first.
#
# PREREQ: step 3 done (composition installed on tool-portal-light) -- confirmed
# by step 4 verify (layout tool-portal-light, palette parchment, no gap item).
# Envelope identical to step 3 / 082 / 080c.
# =====================================================================
set -euo pipefail

SITE_ID="1244516d-014d-421c-88c6-090bb1e9552a"   # idea.uk
DOMAIN="idea.uk"
AGENT="webdesign-agent"
INPUT_DATA="{\"site_id\":\"${SITE_ID}\",\"domain\":\"${DOMAIN}\"}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"

echo "Rendering + deploying styles.css for site ${SITE_ID} (${DOMAIN}) via ${AGENT}"
echo "  correlation_id=${CORRELATION_ID}"

kubectl -n kafka run -i --rm kcat-render-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=render-idea-uk-$(date +%Y%m%d-%H%M%S) \
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
echo "=== Monitor (kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db) ==="
echo "  SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid;"
echo "  -- expect: check_site_context -> load_site_context -> read_site_specs -> analyze_design ->"
echo "  --   update_site -> generate_css -> deploy_css -> spawn/call asset_renderer -> complete"
echo ""
echo "=== Then verify the DEPLOYED CSS (outside SQL) ==="
echo "  1. git: a new commit 'Update stylesheet via webdesign-agent' to assets/css/styles.css on idea.uk's repo,"
echo "     and the GH Actions -> B2 deploy succeeded."
echo "  2. styles.css reads LIGHT/PARCHMENT: --color-background ~ #EFE7D6, --color-accent ~ #A8391A, ink #1A1816,"
echo "     and the tool-portal-light structure (flat 1px rule, serif headings). Compare the palette to the"
echo "     installed parchment (palettes row) -- flag if analyze_design's LLM drifted materially."
echo "  3. load a couple of pages: they should re-skin to light/parchment from the new CSS (page HTML unchanged)."
