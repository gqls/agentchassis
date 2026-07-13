#!/usr/bin/env bash
# TRIGGER_code_indexer_v2.sh — §7C reindex trigger (post §7B migration).
#
# PRE-FLIGHT (in order — item 1 is the hard blocker):
#   1. spawn gate DEPLOYED: spawn_actions.go isRepoCloningAgent now lists
#      "code-indexer"; chassis/core-manager image rebuilt + rolled out (note the
#      tag). Without it the SPAWNED indexer pod has no GITHUB_READ_TOKEN and
#      analyse_repo_local fails immediately (loud, pre-fetch) — recoverable,
#      but wastes the run.
#   2. branch PUSHED: tarball/<REF> fetches the REMOTE tip of REF — the same
#      push that fed the image build also sets the tree to be indexed.
#   3. REF explicit (user decision 2026-07-02: merge meaningful commits to main
#      manually; envelopes still carry the explicit branch).
set -euo pipefail
TARGET_AGENT_TYPE='index-orchestrator'   # spawn wrapper (2026-07-02): in-place code-indexer on the chassis pod has no token
OWNER='gqls'
REPO='agentchassis'
REF='083_imagery'
# dynamic alternative: REF="${REF:-$(git -C "$HOME/projects/agentchassis" rev-parse --abbrev-ref HEAD)}"
LANGUAGE='go'
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "========================================="
echo "Manual code-indexer trigger via index-orchestrator (§7C reindex)"
echo "========================================="
echo "  Target Agent Type: $TARGET_AGENT_TYPE"
echo "  Owner: $OWNER"
echo "  Repo: $REPO"
echo "  Ref: $REF   (explicit — never HEAD)"
echo "  Timestamp: $TIMESTAMP"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""
kubectl -n kafka run -i --rm "kcat-code-indexer-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-code-indexer-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":{"owner":"$OWNER","repo":"$REPO","ref":"$REF","language":"$LANGUAGE"}}
JSON
echo ""
echo "========================================="
echo "code-indexer triggered."
echo "========================================="
echo ""
echo "1) Spawned pod appears, then confirm the token injection landed:"
echo "  kubectl -n ai-persona-system get pods | grep -i code-indexer"
echo "  P=\$(kubectl -n ai-persona-system get pods -o name | grep agent-code-indexer | head -1)"
echo "  [ -n \"\$P\" ] && kubectl -n ai-persona-system describe \"\$P\" | grep -A3 GITHUB_READ_TOKEN || echo 'no agent-code-indexer pod yet'"
echo "  # NOTE: an empty grep + bare 'kubectl describe pod' describes EVERY pod — that is how the analyser-adapter's env got matched on 2026-07-02"
echo ""
echo "2) Follow the run (steps are request_analysis -> index_symbols -> complete):"
echo "  kubectl -n ai-persona-system logs -f -l agent_type=code-indexer --tail=500 | grep '$CORRELATION_ID'"
echo "  # step/action markers worth grepping alongside the correlation id:"
echo "  #   analyse_repo_local | index_code_symbols | request_analysis | index_symbols | tarball | GITHUB_READ_TOKEN"
echo ""
echo "3) Orchestration state (by correlation id, never by created_at):"
echo "  psql: SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s, substring(COALESCE(error,''),1,300) AS err FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;"
echo ""
echo "4) On COMPLETED, run the §7C verification block in RUNBOOK_code_retrieval_route.md:"
echo "   - the previously-zero-rows page/section/result_spec query must return rows"
echo "   - count(*) in the low thousands, count(DISTINCT path) ~572, ONE commit_sha (not 4c2c172)"
echo ""
echo "Timing note: first run at the new commit embeds EVERY symbol via the"
echo "ollama-adapter — expect minutes, not seconds (workflow timeout now 1800s)."
