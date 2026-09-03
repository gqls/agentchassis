#!/usr/bin/env bash
# Direct page-rerender orchestrate for ONE page — bypasses the stalled
# build-dispatch-loop queue (bug 029). Assemble-only (no reason stamped -> the
# render_page branch), so authored content_data/rendered_html is untouched.
# Uses the proven 049_TRIGGER pattern: `kcat -P -c 1` + heredoc (‑c 1 reads
# exactly one message, sidestepping the kubectl-run stdin race).
#
# OPTIONAL 4th arg REASON (added 2026-07-25, bugs_open/023): with no reason the
# workflow takes the assemble-only render_page branch, which stitches the STORED
# rendered_html and therefore cannot pick up a component template change. Pass
# `section_data_resolved` (or `image_landed`) to take the rerender_sections
# pre-pass instead: every section is re-rendered from its stored content_data +
# freshly resolved fields through the CURRENT html_template, with NO LLM call
# (rerender_page_sections_action.go:1-17). `cta_links_stale` additionally
# recomputes CTA destinations. GOTCHA: if ANY section has NULL content_data the
# whole page escalates to the content writer and the copy IS regenerated —
# check first:
#   SELECT slot_name, content_data IS NULL FROM page_components WHERE page_id='...';
set -euo pipefail
PAGE_ID="$1"; SITE_ID="$2"; DOMAIN="$3"; REASON="${4:-}"
CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# The conditional reads input_data.spec.reason — NOT a top-level reason. A
# top-level one is silently ignored and you get the assemble-only branch (cost
# me a full queue round on 2026-07-25; verified against the live step config:
#   default_config->'workflow'->'steps'->'check_rerender_mode'->'config'->>'condition'
#   = "input_data.spec.reason == 'image_landed' OR ... 'section_data_resolved' ...")
# A REASONED dispatch must ALSO carry spec.page_name, or save_page_sections
# SKIPS with {"success":true,"sections_saved":0,"reason":"no page name"} — its
# config reads page_name_field=input_data.spec.page_name — and the pipeline
# then assembles + deploys the STALE stored rows with every step green
# (proven live 2026-09-03, corr 7487f8fa: rerendered=4, saved=0, deploy
# success, rows untouched; re-fired with page_name → saved=4 and the change
# actually shipped). Derived here from page_id so the caller cannot omit it.
REASON_JSON=""
if [ -n "$REASON" ]; then
  PAGE_NAME=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -t -A \
    -c "SELECT name FROM pages WHERE id='$PAGE_ID'")
  [ -n "$PAGE_NAME" ] || { echo "049b: no pages row for $PAGE_ID — refusing a reasoned dispatch whose save would silently skip" >&2; exit 2; }
  REASON_JSON=",\"spec\":{\"reason\":\"$REASON\",\"page_name\":\"$PAGE_NAME\"}"
fi
echo "corr=$CORR page=$PAGE_ID domain=$DOMAIN reason=${REASON:-<assemble-only>}"
kubectl -n kafka run -i --rm "kcat-legal-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR \
  -H orchestration_id=$ORCH \
  -H request_id=$REQ \
  -H message_id=$MSG \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=orchestrate \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TS <<JSON
{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"page_id":"$PAGE_ID","site_id":"$SITE_ID","domain":"$DOMAIN"$REASON_JSON}}
JSON
echo "CORR=$CORR"
