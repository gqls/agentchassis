#!/usr/bin/env bash
# 090_TRIGGER_needs_diagnosis_v1.sh — F0.1c: the ONE documented way a bug goes
# into the diagnosis→fix loop.
#
# It does two things, in order:
#   1. writes a durable `needs_diagnosis` work item (pipeline='diagnose') —
#      the intake RECORD, queryable long after the pods are reaped;
#   2. fires the diagnose-orchestrator envelope of 084, carrying the SAME
#      correlation_id, so the item, the bundles in diagnosis_artifacts, and the
#      terminal doc_notes row all join on one key.
#
# Extends drafts/084_TRIGGER_diagnose_v1.sh (the canonical envelope). It does
# not replace it: 084 remains the bare manual trigger for ad-hoc runs with no
# intake record. Q-B kept both, deliberately.
#
# ── THREE THINGS THE DESIGN GOT WRONG UNTIL THE SCHEMA WAS READ (2026-07-09) ──
#
# 1. site_id is NOT NULL. Q-B said "site-less code bugs ride null-site". They
#    cannot: the column forbids it, AND LoadWorkItemsAction parses site_id as a
#    required uuid and filters `WHERE wi.site_id = $1`, so a NULL-site item
#    would never be loaded even if the column allowed it.
#
# 2. The platform already solved this. `system.internal`
#    (eac60db8-b032-432b-b36d-76f37632045d, sites.status='system') is an
#    existing pseudo-site carrying platform-wide work — the maintenance-pipeline
#    component_quality_scan items live there today. We anchor to it rather than
#    invent a second mechanism. Reuse before recreate.
#
# 3. EVERY needs_diagnosis item anchors to system.internal — even when the bug
#    IS about a real site. The site under diagnosis travels in spec.site_id /
#    spec.runtime_site, never in the item's own site_id column. Why: the live
#    build-dispatch-loop's load_items step is configured with ONLY
#    {site_id, max_items} — it has NO item_pipeline filter — so an item parked
#    on a real site would be claimed by that site's next build dispatch and
#    handed to whatever handler_agent it names. Anchoring to system.internal
#    keeps the diagnose namespace physically off every per-site dispatch loop,
#    without editing a workflow the builder thread owns.
#
# 4. The item parks at status='awaiting_diagnosis', NOT 'detected'.
#    triage_detect_items promotes `WHERE site_id = $1 AND status='detected'`
#    with no pipeline filter, and REWRITES pipeline to 'build' — so 'detected'
#    would let a diagnose item be laundered into the build queue. No existing
#    sweep selects 'awaiting_diagnosis', so it is inert by construction.
#    See 0NN_diagnose_dispatch_loop.sql for the full inertness table.
#
# 5. max_attempts=1. claimed-item-timeout resets any claim older than 40 minutes
#    back to 'triaged' — which WOULD expose a slow diagnosis to the build
#    dispatcher. With max_attempts=1 that reset lands on 'failed' (terminal)
#    instead. It is also the right semantics: a 26-minute LLM loop should not
#    silently auto-retry.
#
# Automatic dispatch: `diagnose-dispatch-loop` + the `diagnose-pipeline-trigger`
# scheduled task (0NN_diagnose_dispatch_loop.sql) claim awaiting_diagnosis items
# on a 60s tick. The task ships DISABLED — enable it deliberately. Until then,
# and for any ad-hoc run, THIS SCRIPT is the dispatcher.
#
# Usage:
#   ./090_TRIGGER_needs_diagnosis_v1.sh "symptom text — keep free of \" and \\"
#
#   SLUG=guides-nav RUNTIME_SITE=dartsonline.com REF=main \
#     ./090_TRIGGER_needs_diagnosis_v1.sh "nav links to a guides section with no content"
#
# Env:
#   SLUG           item_key suffix (default: derived from the symptom)
#   REF            explicit git ref — NEVER HEAD (user decision 2026-07-02)
#   RUNTIME_SITE   domain of the site under diagnosis (runtime evidence tier)
#   SITE_ID        uuid of the site under diagnosis (goes in spec, not site_id)
#   SUBJECT_TYPE   tool | pipeline   } together, these open the tools chat's
#   SUBJECT_KEY    <function> | <pipeline> } persist_note subject gate (their 3b)
#   SEED_SCOPE     comma-separated path[:Symbol] entries for iteration 1
#                  (emitted into the envelope as a JSON ARRAY — a bare string
#                   parses to nil in ExtractStringListHelper and does nothing)
#   DISPATCH       1 (default) fires the Kafka envelope; 0 records the item only
set -euo pipefail

SYMPTOM="${1:?usage: $0 \"symptom text\"}"

# The pseudo-site every platform-wide work item is anchored to. Not a magic
# number: SELECT id FROM sites WHERE domain='system.internal'.
SYSTEM_SITE_ID='eac60db8-b032-432b-b36d-76f37632045d'

TARGET_AGENT_TYPE='diagnose-orchestrator'
OWNER="${OWNER:-gqls}"
REPO="${REPO:-agentchassis}"
REF="${REF:-main}"
RUNTIME_SITE="${RUNTIME_SITE:-}"
SITE_ID="${SITE_ID:-}"
SUBJECT_TYPE="${SUBJECT_TYPE:-}"
SUBJECT_KEY="${SUBJECT_KEY:-}"
SEED_SCOPE="${SEED_SCOPE:-}"
DISPATCH="${DISPATCH:-1}"
CLIENT_ID='demo_client'

PSQL=(kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1)

# item_key drives idx_swi_dedup (site_id, item_key) WHERE status not terminal —
# so re-running with the same SLUG while an intake is still open is a no-op
# rather than a duplicate. That is the intended idempotency.
if [ -z "${SLUG:-}" ]; then
  SLUG=$(printf '%s' "$SYMPTOM" | tr '[:upper:]' '[:lower:]' \
         | tr -cs 'a-z0-9' '-' | cut -c1-40 | sed 's/^-//; s/-$//')
fi
ITEM_KEY="needs_diagnosis:${SLUG}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

# ── the envelope, built once and shared by the item spec and the Kafka message ─
INPUT_DATA="{\"symptom\":\"$SYMPTOM\",\"owner\":\"$OWNER\",\"repo\":\"$REPO\",\"ref\":\"$REF\""
[ -n "$RUNTIME_SITE" ] && INPUT_DATA="${INPUT_DATA},\"runtime_site\":\"$RUNTIME_SITE\""
[ -n "$SITE_ID" ]      && INPUT_DATA="${INPUT_DATA},\"site_id\":\"$SITE_ID\""
if [ -n "$SEED_SCOPE" ]; then
  # comma-separated -> JSON array. ExtractStringListHelper takes []interface{}
  # or []string only; a bare "a,b" string yields nil and the seed is ignored.
  SCOPE_JSON=$(printf '%s' "$SEED_SCOPE" | awk -F, '{for(i=1;i<=NF;i++){printf "%s\"%s\"", (i>1?",":""), $i}}')
  INPUT_DATA="${INPUT_DATA},\"seed_scope\":[${SCOPE_JSON}]"
fi
if [ -n "$SUBJECT_TYPE" ] && [ -n "$SUBJECT_KEY" ]; then
  INPUT_DATA="${INPUT_DATA},\"subject_type\":\"$SUBJECT_TYPE\",\"subject_key\":\"$SUBJECT_KEY\""
fi
INPUT_DATA="${INPUT_DATA},\"correlation_id\":\"$CORRELATION_ID\"}"

echo "========================================="
echo "needs_diagnosis intake  (pipeline='diagnose')"
echo "========================================="
echo "  Symptom:       ${SYMPTOM}"
echo "  item_key:      ${ITEM_KEY}"
echo "  Anchor site:   system.internal  (target site travels in spec)"
echo "  Repo:          ${OWNER}/${REPO} @ ${REF}   (explicit — never HEAD)"
[ -n "$RUNTIME_SITE" ] && echo "  Under diagnosis: ${RUNTIME_SITE}"
[ -z "${RUNTIME_SITE}${SITE_ID}" ] && echo "  Anchor:        NONE (code-only) — needs the load_runtime error_step migration"
if [ -n "$SUBJECT_TYPE" ] && [ -n "$SUBJECT_KEY" ]; then
  echo "  Subject:       ${SUBJECT_TYPE}/${SUBJECT_KEY}   (persist_note WRITES a doc_notes row, post-3b)"
else
  echo "  Subject:       NONE — persist_note will SKIP (by design; do not guess)"
fi
echo "  Correlation:   ${CORRELATION_ID}"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo ""

# ── 1. the durable intake record ──────────────────────────────────────────────
# ON CONFLICT DO NOTHING pairs with idx_swi_dedup: a second intake for a still-open
# slug is silently ignored. rows_written tells you which happened.
echo "-- writing needs_diagnosis work item ..."
"${PSQL[@]}" <<SQL
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key, max_attempts
) VALUES (
    '${SYSTEM_SITE_ID}', '090_TRIGGER_needs_diagnosis', 'diagnose', 'needs_diagnosis',
    'medium', \$sum\$${SYMPTOM}\$sum\$,
    \$json\$${INPUT_DATA}\$json\$::jsonb,
    50, '${TARGET_AGENT_TYPE}', 'awaiting_diagnosis', '090_TRIGGER_needs_diagnosis', '${ITEM_KEY}',
    1   -- see header note 5: makes the 40-minute claim reset terminal, not 'triaged'
)
ON CONFLICT DO NOTHING;

SELECT count(*) AS open_intakes_for_this_slug
FROM site_work_items
WHERE site_id = '${SYSTEM_SITE_ID}' AND item_key = '${ITEM_KEY}'
  AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved');
SQL

if [ "$DISPATCH" != "1" ]; then
  echo ""
  echo "DISPATCH=0 — intake recorded, no Kafka message sent."
  exit 0
fi

# ── 2. dispatch, on the SAME correlation_id ───────────────────────────────────
echo ""
echo "-- dispatching to ${TARGET_AGENT_TYPE} ..."
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
  -H "orchestration_name=needs-diagnosis-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":$INPUT_DATA}
JSON

cat <<EOF

=========================================
intake recorded and dispatched.
=========================================

1) Follow the run:
   kubectl -n ai-persona-system logs -f -l agent-type=diagnose-agent --tail=500 | grep '$CORRELATION_ID'

2) Orchestration state (by correlation id, never by created_at):
   SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s,
          substring(COALESCE(error,''),1,200) AS err
   FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;

3) FETCH THE BUNDLES (F0.1 — this is the egress route; needs the chassis image
   carrying the diagnose_assemble_bundle write-through):
   SELECT iteration, length(body) AS bytes, metadata->>'symbol_count' AS symbols,
          metadata->>'truncated' AS truncated, created_at
   FROM diagnosis_artifacts
   WHERE correlation_id = '$CORRELATION_ID' AND kind = 'bundle'
   ORDER BY iteration;

   -- one bundle to a file:
   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db \\
     -t -A -c "SELECT body FROM diagnosis_artifacts WHERE correlation_id='$CORRELATION_ID' AND kind='bundle' AND iteration=1" \\
     > bundle_iter1.md

4) The intake record, and closing it by hand until a diagnose dispatch loop exists:
   SELECT status, created_at, completed_at FROM site_work_items WHERE item_key = '$ITEM_KEY';
   UPDATE site_work_items SET status='complete', completed_at=now()
   WHERE item_key = '$ITEM_KEY' AND status NOT IN ('complete','verified');

Timing: minutes, not seconds (repo tarball fetch + up to 5 verdict iterations).
EOF
