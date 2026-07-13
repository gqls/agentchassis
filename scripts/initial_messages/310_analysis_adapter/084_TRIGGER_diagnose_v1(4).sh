#!/usr/bin/env bash
# 084_TRIGGER_diagnose_v1.sh — manual diagnosis trigger (Stage-3a gate test + general use).
# Envelope mirrors 082/083c: kcat pod -> system.agent.generic.requests, action=orchestrate.
#
# TARGET: diagnose-orchestrator (the spawn wrapper). It spawns a DEDICATED
# diagnose-agent pod and forwards the result — keeps the loop's substantive work
# off the shared chassis pods, and the SPAWNED pod gets GITHUB_READ_TOKEN via the
# spawn gate (083c PRE-FLIGHT item 1). Do NOT orchestrate diagnose-agent directly:
# an in-place run on a shared pod has no token and analyse_repo_local fails
# immediately (loud, pre-fetch) — recoverable, but wastes the run.
#
# REF is EXPLICIT — never HEAD (user decision 2026-07-02).
#
# ANCHOR NOTE (2026-07-06 incident): an ANCHORLESS run (no RUNTIME_SITE/SITE_ID)
# requires the load_runtime error routing migration
# (0NN_diagnose_load_runtime_error_step.sql) — without it, diagnose_load_runtime
# hard-fails ("need at least one of site_id / correlation_id / domain") and the
# workflow dies BEFORE persist_note. Either apply that migration, or pass
# RUNTIME_SITE=<domain> / SITE_ID=<uuid>.
#
# Stage-3a NOTE (why this run SKIPS persistence, by design):
#   diagnose-orchestrator's call_diagnoser input_mapping forwards only
#   symptom/owner/repo/ref/site_id/seed_scope/runtime_page/runtime_site/
#   correlation_id — NOT subject_type/subject_key (that threading is Stage 3b).
#   So persist_note runs, finds no explicit subject, logs
#   "persist_diagnosis_note: no explicit subject — skipping (do not guess)"
#   and returns persisted:false. Proving exactly that is this run's purpose.
#   (The commented subject fields below have NO effect until 3b is applied.)
#
# Usage:
#   ./084_TRIGGER_diagnose_v1.sh ["symptom text — keep free of \" and \\"]
#   REF=083_imagery RUNTIME_SITE=gamesdesign.co.uk ./084_TRIGGER_diagnose_v1.sh "pages stale after rebuild"
set -euo pipefail

SYMPTOM="${1:-smoke: subjectless diagnosis to verify persist_note skip gate (Stage 3a)}"
TARGET_AGENT_TYPE='diagnose-orchestrator'
OWNER="${OWNER:-gqls}"
REPO="${REPO:-agentchassis}"
REF="${REF:-main}"                 # explicit branch; override via env
RUNTIME_SITE="${RUNTIME_SITE:-}"   # optional: domain for the runtime evidence tier
SITE_ID="${SITE_ID:-}"             # optional: site uuid
SUBJECT_TYPE="${SUBJECT_TYPE:-}"   # 3b: tool | pipeline (persists a NOTES row)
SUBJECT_KEY="${SUBJECT_KEY:-}"     # 3b: <function> | build|content|design|maintenance
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# input_data — string concat, matching 082 (no jq dependency)
INPUT_DATA="{\"symptom\":\"$SYMPTOM\",\"owner\":\"$OWNER\",\"repo\":\"$REPO\",\"ref\":\"$REF\""
if [ -n "$RUNTIME_SITE" ]; then INPUT_DATA="${INPUT_DATA},\"runtime_site\":\"$RUNTIME_SITE\""; fi
if [ -n "$SITE_ID" ];      then INPUT_DATA="${INPUT_DATA},\"site_id\":\"$SITE_ID\""; fi
# 3b subject (takes effect only after 0NN_wire_diagnosis_subject_threading.sql):
if [ -n "$SUBJECT_TYPE" ] && [ -n "$SUBJECT_KEY" ]; then
  INPUT_DATA="${INPUT_DATA},\"subject_type\":\"$SUBJECT_TYPE\",\"subject_key\":\"$SUBJECT_KEY\""
fi
INPUT_DATA="${INPUT_DATA}}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'

echo "========================================="
echo "Manual diagnosis trigger via diagnose-orchestrator"
echo "========================================="
echo "  Symptom:       ${SYMPTOM}"
echo "  Repo:          ${OWNER}/${REPO} @ ${REF}   (explicit — never HEAD)"
[ -n "$RUNTIME_SITE" ] && echo "  Runtime site:  ${RUNTIME_SITE}"
[ -z "${RUNTIME_SITE}${SITE_ID}" ] && echo "  Anchor:        NONE (code-only — needs the load_runtime error_step migration, else fails at load_runtime)"
if [ -n "$SUBJECT_TYPE" ] && [ -n "$SUBJECT_KEY" ]; then
  echo "  Subject:       ${SUBJECT_TYPE}/${SUBJECT_KEY}   (persist_note WRITES a doc_notes row)"
else
  echo "  Subject:       NONE — persist_note will SKIP (for 3b: same-line prefix or export SUBJECT_TYPE/SUBJECT_KEY)"
fi
echo "  Correlation:   ${CORRELATION_ID}"
echo "  Orchestration: ${ORCHESTRATION_ID}"
echo "  Timestamp: $TIMESTAMP"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}  ORCHESTRATION_ID=${ORCHESTRATION_ID}"
echo ""

kubectl -n kafka run -i --rm "kcat-diagnose-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-diagnose-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":$INPUT_DATA}
JSON

echo ""
echo "========================================="
echo "diagnosis triggered."
echo "========================================="
echo ""
echo "1) Spawned diagnoser pod (~10-20s), then confirm token injection:"
echo "  kubectl -n ai-persona-system get pods -l agent-type=diagnose-agent --sort-by=.metadata.creationTimestamp | tail -3"
echo "  P=\$(kubectl -n ai-persona-system get pods -o name | grep agent-diagnose-agent | head -1)"
echo "  [ -n \"\$P\" ] && kubectl -n ai-persona-system describe \"\$P\" | grep -A3 GITHUB_READ_TOKEN || echo 'no diagnose-agent pod yet'"
echo "  # NOTE: an empty grep + bare 'kubectl describe pod' describes EVERY pod (083c, 2026-07-02)"
echo ""
echo "2) Follow the run (markers: analyse_repo_local -> lookup_code_symbols -> diagnose_load_runtime"
echo "   -> diagnose_assemble_bundle -> verdict -> route -> diagnose_emit -> persist_diagnosis_note):"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=diagnose-agent --tail=500 | grep '$CORRELATION_ID'"
echo ""
echo "3) Orchestration state (by correlation id, never by created_at):"
echo "  SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s,"
echo "         substring(COALESCE(error,''),1,200) AS err"
echo "  FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;"
echo ""
echo "4) STAGE-3a GATE VERIFICATION (only meaningful once status = COMPLETED — the 0-rows rule):"
echo "  a. skip log line present:"
echo "     kubectl -n ai-persona-system logs -l agent-type=diagnose-agent --tail=2000 | grep 'persist_diagnosis_note'"
echo "     # expect: 'no explicit subject — skipping (do not guess)'"
echo "  b. no diagnosis note written:"
echo "     SELECT count(*) AS diagnosis_notes_last_2h FROM doc_notes"
echo "     WHERE categories ? 'diagnosis' AND created_at > now() - interval '2 hours';   -- expect 0"
echo "     -- 0 is decisive HERE only because (a) the run COMPLETED and (b) the skip line exists."
echo ""
echo "Timing: minutes, not seconds (repo tarball fetch + up to 5 verdict iterations; timeout 1800s)."
