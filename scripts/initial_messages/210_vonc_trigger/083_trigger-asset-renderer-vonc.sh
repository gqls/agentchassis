#!/usr/bin/env bash
# trigger-asset-renderer-vonc.sh
#
# Manually trigger site-asset-renderer for vonc.com to rebuild
# /assets/js/snippets.js and commit it to the site repo.
#
# Why: a js_snippet (provocation-card-loader) was added to the DB AFTER the
# site was built. Nothing in the normal page build/rerender flow re-runs the
# snippet bundle (that is Gap 3 — site-asset-renderer is only invoked by
# webdesign-agent at initial design or by a full site rerender). This script
# invokes it directly so the new snippet is bundled and deployed.
#
# site-asset-renderer workflow: ensure_site_record -> render_js_snippets_for_site
# -> git_commit (writes assets/js/snippets.js). Deterministic, no LLM.
# Input contract: site_id required, domain optional.
#
# Spawned via the generic entry point (system.agent.generic.requests with
# config.agent_type), the same pattern as the classifier smoke test. Runs
# in-chassis — fine for a one-off; routine triggering should be wired into a
# discovery check (the real Gap 3 fix).
# ────────────────────────────────────────────────────────────────────────

set -euo pipefail

SITE_ID='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
DOMAIN='vonc.com'

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "========================================="
echo "Manual site-asset-renderer trigger"
echo "========================================="
echo "  Site:             $DOMAIN"
echo "  Site ID:          $SITE_ID"
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo "  TIMESTAMP=$TIMESTAMP"
echo ""

kubectl -n kafka run -i --rm "kcat-assetrender-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-assetrender-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"site-asset-renderer"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN"}}
JSON

echo ""
echo "========================================="
echo "site-asset-renderer triggered (generic entry point, runs in a chassis pod)."
echo "========================================="
echo ""
echo "Tail chassis logs for this correlation:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=500 | grep '$CORRELATION_ID'"
echo ""
echo "Watch the render + commit steps:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=500 | grep '$CORRELATION_ID' | grep -E 'render_js_snippets|RenderJSSnippets|snippet_count|snippet_names|deploy_js_snippets|git_commit'"
echo ""
echo "Confirm the snippet was included (expect provocation-card-loader in snippet_names):"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=500 | grep '$CORRELATION_ID' | grep -E 'snippet_names|snippet_count'"
echo ""
echo "Check orchestration state:"
echo "  psql -c \"SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s, substring(COALESCE(error,''), 1, 300) AS err FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;\""
echo ""
echo "========================================="
echo "After it completes, verify the deployed bundle:"
echo "========================================="
echo ""
echo "1. The committed file should contain the provocation-card-loader block."
echo "   Fetch it from the live site once the git->S3 deploy lands:"
echo "      curl -s https://$DOMAIN/assets/js/snippets.js | head -40"
echo "   Expect the header comment + the '--- provocation-card-loader ---' marker."
echo ""
echo "2. snippets.js only delivers the loader. The card still needs"
echo "   /data/provocations.json present (P2-2c) before it will fill."
echo ""
echo "NOTE: there may be a CDN/edge cache on /assets/js/snippets.js and"
echo "/data/provocations.json — if the live files look stale after the commit"
echo "+ S3 deploy, a cache purge may be needed. The loader fetches with"
echo "cache:'no-cache' which only covers the browser, not an edge cache."
