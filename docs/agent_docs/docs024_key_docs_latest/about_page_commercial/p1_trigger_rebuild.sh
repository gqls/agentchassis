#!/usr/bin/env bash
# p1_trigger_rebuild.sh — Phase 1 pilot dispatch: rebuild finetuning.uk's about
# page (the ONLY page flagged needs_rebuild) so the page-rebuild agent
# re-selects sections — picking up about-commercial-block — and re-resolves
# site_specs.commercial.* into resolved_data.
#
# PRE-REQS (both applied 2026-07-24): 202_about_commercial_block_component.sql,
# p1_pilot_finetuning.sql. Envelope mirrors travelling_docs/086 (kcat pod ->
# system.agent.generic.requests, action=orchestrate, config.agent_type=...).
#
# SAFETY: DRY-RUN BY DEFAULT — validates + prints, does NOT produce.
#     SEND=1 ./p1_trigger_rebuild.sh        # fire for real
# (env must be same-line prefix or exported; `SEND=1; ./script` does NOT reach
# the process — the env-prefix trap.)
#
# REAL SIDE EFFECTS when SEND=1: rebuilds the LIVE about page on finetuning.uk —
# re-runs the content writer on its other sections (hero-about, content-block-
# about, leadership-team, departments-grid, differentiators, call-to-action),
# renders, git-commits and deploys. Accepted for the pilot (PLAN D14).
#
# Do NOT dispatch within ~300s of a chassis pod (re)start (spawn silently
# dropped). Queue latency can be ~30 min — a missing orchestration row is
# latency, not a drop; find the run by correlation_id, never by created_at.
set -euo pipefail

TARGET_AGENT_TYPE='page-rebuild'
DOMAIN="${DOMAIN:-finetuning.uk}"
SITE_ID="${SITE_ID:-1368e337-dd1d-4799-bbb3-8221a1b79bcc}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'
ORCH_NAME="pilot-about-commercial-$(date +%Y%m%d-%H%M%S)"

# input_data single-line (kcat -P is line-delimited; multi-line body = one
# message per line — the fragment trap that burned run 464102f4).
INPUT_DATA="{\"domain\":\"$DOMAIN\",\"site_id\":\"$SITE_ID\"}"
MESSAGE_BODY="{\"action\":\"orchestrate\",\"config\":{\"agent_type\":\"$TARGET_AGENT_TYPE\"},\"input_data\":$INPUT_DATA}"
echo "$MESSAGE_BODY" | python3 -m json.tool >/dev/null \
  || { echo "ERROR: assembled message body is not valid JSON" >&2; exit 1; }

echo "========================================="
echo "Phase-1 pilot: page-rebuild ${DOMAIN} about page (about-commercial-block)"
echo "  Correlation:   ${CORRELATION_ID}"
echo "  Orchestration: ${ORCHESTRATION_ID}"
echo "  Body:          ${MESSAGE_BODY}"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo ""

if [ "${SEND:-0}" != "1" ]; then
  echo ">>> DRY RUN (SEND != 1). Nothing produced. To fire: SEND=1 $0"
  exit 0
fi

echo ">>> SEND=1 — producing to Kafka (REAL rebuild+deploy of the live about page)..."
kubectl -n kafka run -i --rm "kcat-acb-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=$ORCH_NAME" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
$MESSAGE_BODY
JSON

echo ""
echo "Track (by correlation, NEVER created_at):"
echo "  SELECT status, current_step, EXTRACT(EPOCH FROM (NOW()-last_activity))::int AS since_s"
echo "  FROM orchestration_states WHERE correlation_id='$CORRELATION_ID'::uuid;"
echo ""
echo "Verify against the LIVE PAGE (bugs_closed/024 lesson):"
echo "  curl -s https://$DOMAIN/about.html | grep -c 'data-component=\"about-commercial-block\"'   # placement"
echo "  curl -s https://$DOMAIN/about.html | grep -c 'Built by'                                     # template-created phrase"
echo "  curl -s https://$DOMAIN/about.html | grep -c 'available to acquire'                         # MUST be 0 (gated off)"
