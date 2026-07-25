#!/bin/bash
# ============================================================================
# TRIGGER_submit_tier3.sh — FRESH domain submission WITH a roadmap brief
# ============================================================================
# Why this exists: 082_submit_domain_unified.sh only accepts --mission /
# --mission-file. It has NO roadmap flag — yet `domain-submitter` persists
# input_data.roadmap and input_data.roadmap_brief (068_domain_submitter_agent.sql
# steps persist_roadmap / persist_roadmap_brief), and build-site-planner treats
# the roadmap brief as THE AUTHORITY for the page list:
#
#   "ROADMAP OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the
#    current phase below … Do NOT invent additional pages."
#   (053_build_site_planner.sql:769)
#
# Without it the planner falls back to ensure_pages ["index","contact"] plus
# whatever it invents — which is how a deliberately small site becomes a
# twenty-page one.
#
# SHAPE THAT MATTERS: both briefs are OBJECTS with a `text` key, not strings.
# The prompt renders {{.site_specs.specs.roadmap_brief.text}} — a bare string
# renders empty and the roadmap is silently ignored.
#
# The structured `roadmap` field (per-page section_types) is deliberately NOT
# sent: naming section_types that are not real component names routes through
# the "unknown section types handled downstream" path, and we would rather the
# planner pick from the live component list.
#
# Usage:
#   ./TRIGGER_submit_tier3.sh <domain> <mission-file> <roadmap-file> [email]
#
# Brief files: everything after the FIRST '---' line is sent. Put commentary
# above it; it never reaches the model.
# ============================================================================

set -euo pipefail

DOMAIN="${1:?Usage: $0 <domain> <mission-file> <roadmap-file> [email]}"
MISSION_FILE="${2:?missing mission file}"
ROADMAP_FILE="${3:?missing roadmap file}"
EMAIL="${4:-}"

for f in "$MISSION_FILE" "$ROADMAP_FILE"; do
  [ -f "$f" ] || { echo "file not found: $f" >&2; exit 2; }
done

# Body = everything after the first '---' line; then JSON-escape (backslash,
# then double quote) and fold whitespace to single spaces so it is one JSON
# string value. Same treatment 082 gives --mission-file, no jq dependency.
extract() {
  awk 'f{print} /^---$/{f=1}' "$1" \
    | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g' \
    | tr '\n\t' '  ' | sed -e 's/  */ /g' -e 's/^ //' -e 's/ $//'
}

MISSION="$(extract "$MISSION_FILE")"
ROADMAP="$(extract "$ROADMAP_FILE")"

[ -n "$MISSION" ] || { echo "mission body empty — is there a '---' line?" >&2; exit 2; }
[ -n "$ROADMAP" ] || { echo "roadmap body empty — is there a '---' line?" >&2; exit 2; }

INPUT_DATA="{\"domain\":\"$DOMAIN\",\"fidelity\":\"medium\""
[ -n "$EMAIL" ] && INPUT_DATA="${INPUT_DATA},\"email\":\"$EMAIL\""
INPUT_DATA="${INPUT_DATA},\"mission_brief\":{\"text\":\"$MISSION\"}"
INPUT_DATA="${INPUT_DATA},\"roadmap_brief\":{\"text\":\"$ROADMAP\"}}"

# Fail before publishing rather than after: an unparseable envelope is a
# silently dropped message.
if command -v python3 >/dev/null; then
  printf '%s' "$INPUT_DATA" | python3 -c 'import json,sys; json.load(sys.stdin)' \
    || { echo "input_data is not valid JSON — aborting" >&2; exit 3; }
fi

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

echo "========================================="
echo "Tier-3 FRESH submit: ${DOMAIN}"
echo "  Agent:         domain-submitter"
[ -n "$EMAIL" ] && echo "  Email:         ${EMAIL}"
echo "  Mission chars: ${#MISSION}"
echo "  Roadmap chars: ${#ROADMAP}"
echo "  Correlation:   ${CORRELATION_ID}"
echo "  Orchestration: ${ORCHESTRATION_ID}"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-submit-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=submit-${DOMAIN}-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=demo_client \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"domain-submitter"},"input_data":$INPUT_DATA}
JSON

cat <<EOF

=== Monitor ===
psql: kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db

  SELECT status, current_step, error FROM orchestration_states
   WHERE correlation_id='${CORRELATION_ID}'::uuid;

  -- did BOTH briefs persist? (this is the whole point of this script)
  SELECT aspect, source_agent, is_current, length(data->>'text') AS chars
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='${DOMAIN}' AND aspect IN ('mission_brief','roadmap_brief');

  SELECT aspect, source_agent, is_current, created_at
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='${DOMAIN}' ORDER BY created_at;

  SELECT item_type, status, handler_agent, LEFT(summary,60)
    FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
   WHERE s.domain='${DOMAIN}' ORDER BY priority;

An absent orchestration row usually means QUEUED, not dropped — dispatch latency
of 16-30 minutes is normal under load. Check whether OTHER orchestrations started
in the meantime before concluding anything, and do not resubmit on that evidence.
EOF
