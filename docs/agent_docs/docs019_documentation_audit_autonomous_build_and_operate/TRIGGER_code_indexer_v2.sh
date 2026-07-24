#!/usr/bin/env bash
# TRIGGER_code_indexer_v2.sh — §7C manual reindex trigger, via index-orchestrator.
#
# > CORRECTED 2026-07-24 (bugs_open/059 fix #2): this script used to dispatch
# > agent_type=code-indexer DIRECTLY. That gets adopted in-place by a chassis pod
# > which has no GITHUB_READ_TOKEN, so analyse_repo_local fails immediately —
# > verified 2026-07-22: direct dispatch FAILED, index-orchestrator dispatch
# > COMPLETED. The corrected form existed since 2026-07-02 but only in a stray
# > "TRIGGER_code_indexer_v2(1).sh" duplicate, so the canonical name kept wasting
# > a run. Folded together here; the (1) copy is deleted. Also corrected: REF was
# > hardcoded to a dead branch (083_imagery) — the drift class 059 is about.
#
# WHEN TO USE. A scheduled_tasks row `code-index-refresh` (bugs_open/059 fix #1)
# already fires index-orchestrator every 24h with a DERIVED ref. Run this by hand
# only when you need the index current NOW — e.g. just before firing a diagnosis
# whose evidence lives in code committed today.
#
# PRE-FLIGHT:
#   1. PUSH the branch first: the indexer fetches tarball/<REF> — the REMOTE tip.
#      An unpushed branch reindexes stale code silently (2026-07-24: the index sat
#      exactly at origin's tip while local HEAD was 60 commits ahead — the cadence
#      cannot see what you have not pushed).
#   2. REF defaults to this repo's current branch; override with REF=<branch>.
#   3. Target is index-orchestrator (spawn_indexer -> call_indexer -> complete).
#      It spawns the code-indexer as its own pod; isRepoCloningAgent
#      (spawn_actions.go) injects GITHUB_READ_TOKEN into the spawned pod.
set -euo pipefail
TARGET_AGENT_TYPE='index-orchestrator'   # NOT code-indexer — direct dispatch has no token (bugs_open/059)
OWNER='gqls'
REPO='agentchassis'
REF="${REF:-$(git -C "$HOME/projects/agentchassis" rev-parse --abbrev-ref HEAD)}"
LANGUAGE='go'
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "========================================="
echo "Manual code reindex via index-orchestrator (§7C)"
echo "========================================="
echo "  Target Agent Type: $TARGET_AGENT_TYPE"
echo "  Owner: $OWNER"
echo "  Repo: $REPO"
echo "  Ref: $REF   (indexes the REMOTE tip — push first)"
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
  -H "orchestration_name=manual-index-orchestrator-$(date +%Y%m%d-%H%M%S)" \
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
echo "index-orchestrator triggered."
echo "========================================="
echo ""
echo "1) The orchestrator spawns a code-indexer pod; confirm the token landed:"
echo "  P=\$(kubectl -n ai-persona-system get pods -o name | grep agent-code-indexer | head -1)"
echo "  [ -n \"\$P\" ] && kubectl -n ai-persona-system describe \"\$P\" | grep -A3 GITHUB_READ_TOKEN || echo 'no agent-code-indexer pod yet'"
echo "  # NOTE: an empty grep + bare 'kubectl describe pod' describes EVERY pod — that is how the analyser-adapter's env got matched on 2026-07-02"
echo ""
echo "2) Orchestration state (by correlation id, never by created_at; steps are"
echo "   spawn_indexer -> call_indexer -> complete, call_indexer timeout 1800s):"
echo "  psql: SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s, substring(COALESCE(error,''),1,300) AS err FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;"
echo ""
echo "3) On COMPLETED, verify the index actually moved (bugs_open/059 check):"
echo "  psql: SELECT count(*), (SELECT DISTINCT commit_sha FROM code_symbols LIMIT 1) AS sha, max(updated_at) FROM code_symbols;"
echo "  # sha must equal the short sha of the REMOTE tip you pushed; a same-sha"
echo "  # result means the remote had nothing new (did you push?)."
echo ""
echo "Timing note: only CHANGED symbols are re-embedded via the ollama-adapter —"
echo "an incremental run is minutes; a full first run at a new commit is longer."
