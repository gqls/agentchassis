#!/usr/bin/env bash
# 086_TRIGGER_recreate_economy_simulator.sh — TASK-4 PROOF: recreate the broken
# economy-simulator on gamesdesign.co.uk via tool-recreation-handler, fixing two
# real bugs, and watch it write the FIRST machine-written ('pipeline','build')
# 'fix' NOTES row.
#
# Envelope mirrors 084/085 (kcat pod -> system.agent.generic.requests,
# action=orchestrate, config.agent_type=<target>).
#
# SAFETY: DRY-RUN BY DEFAULT. It builds + validates the message, prints it, and
# STOPS. To actually produce to Kafka you must pass SEND=1 (same-line prefix, or
# export — a bare `SEND=1; ./script` does NOT reach the process; see the
# env-prefix trap banked in the runbook).
#     SEND=1 ./086_TRIGGER_recreate_economy_simulator.sh
#
# REAL SIDE EFFECTS when SEND=1 (accepted: gamesdesign is the trial site):
#   - overwrites page_components for game-economy-simulator with the recreated
#     tool HTML; marks the page deployed; spawns page-rerender; deploys to git
#     via the normal pipeline (live domain changes);
#   - AND (Task 4) a doc_notes row, subject ('pipeline','build'),
#     categories ['fix'], source='tool-recreation-handler' — THE PROOF.
#
# The two fixes are carried in spec.interactive_features (see the JSON below):
#   (1) map the Player-Influx slider index 0..5 -> rate [0,1,5,15,40,100];
#   (2) give the Players chart series its own axis (was bound to yGold).
# recreate_tool (Opus, 64k) reads these via the analyze_tool functional spec.
#
# NOTE (established 2026-07-09): load_existing_content reads only the
# adoption_page crawl row, whose raw_html is the 5.8KB HTML SHELL (the origin
# site's logic lived in an external game.js the crawl did not capture). The
# 32KB inline-JS version now live on gamesdesign.co.uk was itself produced by a
# PRIOR recreation run (two tool_recreation_training rows for this page,
# 2026-06-05) — that run introduced the two bugs; the origin game.js has
# neither (no Players series at all; green line there is Total Gold on its own
# axis). So the agent rebuilds from the shell DOM/CSS + markdown + these fix
# requirements. If a faithful repair-what-is-live base is wanted instead,
# enrich research_results.data->existing_content->raw_html with OUR OWN
# deployed page (curl https://gamesdesign.co.uk/games/economy-simulator/
# index.html — no external scraping needed; snapshot the old value first),
# then run this.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_AGENT_TYPE='tool-recreation-handler'
DOMAIN="${DOMAIN:-gamesdesign.co.uk}"
SITE_ID="${SITE_ID:-e33263f4-74f8-494f-b191-546845dbbddf}"
PAGE_NAME="${PAGE_NAME:-game-economy-simulator}"
INPUT_JSON="${INPUT_JSON:-$HERE/086_input_data_recreate_economy_simulator.json}"

[ -f "$INPUT_JSON" ] || { echo "ERROR: input_data JSON not found: $INPUT_JSON" >&2; exit 1; }

# Validate AND COMPACT the input_data payload to a SINGLE LINE. kcat -P is
# line-delimited: a multi-line body is produced as one message PER LINE, so a
# pretty-printed JSON becomes ~37 invalid fragments all carrying our headers
# (this burned run 464102f4 on 2026-07-09 — the chassis married the headers to
# a neighbouring scheduler no-op body). Single-line is load-bearing.
INPUT_DATA="$(python3 -c 'import json,sys; print(json.dumps(json.load(open(sys.argv[1])), separators=(",", ":")))' "$INPUT_JSON")" \
  || { echo "ERROR: $INPUT_JSON is not valid JSON (or python3 missing)" >&2; exit 1; }
case "$INPUT_DATA" in *$'\n'*) echo "ERROR: compacted input_data still multi-line" >&2; exit 1;; esac

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'
ORCH_NAME="manual-recreate-economy-$(date +%Y%m%d-%H%M%S)"

# Full message body (input_data inlined).
MESSAGE_BODY="{\"action\":\"orchestrate\",\"config\":{\"agent_type\":\"$TARGET_AGENT_TYPE\"},\"input_data\":$INPUT_DATA}"
if command -v python3 >/dev/null 2>&1; then
  echo "$MESSAGE_BODY" | python3 -m json.tool >/dev/null || { echo "ERROR: assembled message body is not valid JSON" >&2; exit 1; }
fi

echo "========================================="
echo "Recreate economy-simulator via tool-recreation-handler (Task-4 PROOF)"
echo "========================================="
echo "  Target agent:  ${TARGET_AGENT_TYPE}"
echo "  Site:          ${DOMAIN}  (${SITE_ID})"
echo "  Page:          ${PAGE_NAME}"
echo "  spec.mode:     recreate"
echo "  Fixes:         slider index->rate map; Players own chart axis"
echo "  Correlation:   ${CORRELATION_ID}"
echo "  Orchestration: ${ORCHESTRATION_ID}"
echo "-----------------------------------------"
echo "  input_data:"
sed 's/^/    /' "$INPUT_JSON"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}  ORCHESTRATION_ID=${ORCHESTRATION_ID}"
echo ""

if [ "${SEND:-0}" != "1" ]; then
  echo ">>> DRY RUN (SEND != 1). Nothing produced to Kafka. Message validated only."
  echo ">>> To fire for real:  SEND=1 $0"
  exit 0
fi

echo ">>> SEND=1 — producing to Kafka (REAL side effects on the live site)..."
kubectl -n kafka run -i --rm "kcat-recreate-$(date +%s)" \
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
echo "========================================="
echo "recreation triggered."
echo "========================================="
echo ""
echo "1) Orchestration state (by correlation id, never by created_at):"
echo "  SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s,"
echo "         substring(COALESCE(error,''),1,200) AS err"
echo "  FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;"
echo ""
echo "1b) Which step failed and why (agent_error_log; orchestration_id is TEXT, no ::uuid):"
echo "  SELECT occurred_at, step_name, action, error_code, left(error_message,300) AS err"
echo "  FROM agent_error_log WHERE orchestration_id = '$ORCHESTRATION_ID' ORDER BY occurred_at;"
echo ""
echo "2) TASK-4 PROOF (once COMPLETED) — the first machine-written fix note:"
echo "  SELECT subject_type, subject_key, site_id, categories, source,"
echo "         left(body,140) AS head, created_at"
echo "  FROM doc_notes WHERE categories ? 'fix' ORDER BY created_at DESC LIMIT 5;"
echo ""
echo "3) The recreated page body (was empty; now the tool HTML):"
echo "  SELECT name, build_status, length(sections::text) AS sections_len"
echo "  FROM pages WHERE site_id='$SITE_ID'::uuid AND name='$PAGE_NAME';"
echo ""
echo "Timing: Sonnet analyze + Opus recreate (64k) + Sonnet note; workflow timeout 2400s."
