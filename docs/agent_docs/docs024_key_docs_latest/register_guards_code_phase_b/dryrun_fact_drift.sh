#!/usr/bin/env bash
# One-off DRY RUN of refresh_evidence_base, to see what the fact-drift fan-out
# (CLM-022) would file without writing anything.
#
# Why an inline workflow rather than the evidence-freshness agent: `dry_run` is
# read from the STEP's config, not from input_data, so the live agent row cannot
# be asked for a dry run. The chassis honours an inline workflow at
# body.config.workflow (Priority 1).
#
# ── PUBLISH FORM — CORRECTED 2026-08-17, and the first version was WRONG ─────
# This script originally used `kubectl run -i --rm … kcat -P <<JSON`, i.e. the
# payload on STDIN. That is the stdin-race form LANDMINES records: `kubectl run -i`
# attaches stdin asynchronously, so if the container reaches kcat first it sees
# EOF, publishes NOTHING and exits 0. Measured on the leopardess lane 2026-07-26:
# four of five publishes lost, silently. The failure is invisible — an empty
# orchestration_states five minutes later reads exactly like ordinary fleet
# latency, which CLAUDE.md separately tells you not to retry on.
# Caught by the mortgagecalculator lane using this script (their CONTRIB_REPLY,
# 2026-08-17). The payload now goes in the container COMMAND with a PUBLISH_OK
# receipt: no receipt means nothing was published, so re-run immediately.
#
# Other gotchas, each of which has cost someone a cycle:
#   - the JSON must be ONE line: kcat -P sends one message PER LINE, and each
#     fragment arrives as invalid JSON wearing the full header set.
#   - no dispatch within ~300s of a chassis pod restart — the spawn is dropped.
#   - verify by the orchestration row, never by an exit code.
#
# ── READING THE RESULT — the counter next door LIES ──────────────────────────
# fact_drift is PER-SITE and NESTED:  refresh_result->'results'->N->'fact_drift'
# There is no top-level fact_drift key, and the top-level `total_drifted` counts
# CITATION drift, not fact drift — it reads 0 while each site carries entries.
# The obvious query says "it did not fire" and the neighbouring counter appears
# to confirm it. (mortgagecalculator lane, 2026-08-17; their WRONG_CALLS row.)
#
# Usage: ./dryrun_fact_drift.sh [site_id]   (omit to sweep every site holding a
#        current evidence_base spec)
set -uo pipefail
SITE_ID="${1:-}"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

if [ -n "$SITE_ID" ]; then INPUT="{\"site_id\":\"${SITE_ID}\"}"; else INPUT="{}"; fi
WF='{"start_step":"refresh_evidence","processing_mode":"orchestrator","timeout_seconds":600,"steps":{"refresh_evidence":{"action":"refresh_evidence_base","config":{"dry_run":true},"output_field":"refresh_result","next_step":"complete","description":"Dry-run the evidence sweep incl. the fact-drift fan-out"},"complete":{"action":"complete_workflow","config":{"output_fields":["refresh_result"]},"description":"done"}}}'
BODY="{\"action\":\"orchestrate\",\"config\":{\"workflow\":${WF}},\"input_data\":${INPUT}}"

# base64 so no quoting in BODY can break out of the sh -c string.
B64=$(printf '%s' "$BODY" | base64 -w0)

echo "=== refresh_evidence_base DRY RUN  corr=${CORRELATION_ID}  site=${SITE_ID:-ALL} ==="
kubectl -n kafka run "kcat-factdrift-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "echo '${B64}' | base64 -d | kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
  -H correlation_id=${CORRELATION_ID} \
  -H orchestration_id=$(cat /proc/sys/kernel/random/uuid) \
  -H request_id=$(cat /proc/sys/kernel/random/uuid) \
  -H message_id=$(cat /proc/sys/kernel/random/uuid) \
  -H message_type=request -H client_id=demo_client -H action=orchestrate \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=${TIMESTAMP} && echo PUBLISH_OK"

cat <<EOT

corr=${CORRELATION_ID}
NO 'PUBLISH_OK' ABOVE => nothing was published. Re-run now; do not wait for latency.

Read it back — note the NESTED path, and ignore total_drifted:
  SELECT status,
         jsonb_pretty(jsonb_path_query_array(collected_data,
           '\$.refresh_result.results[*].fact_drift[*]'))
  FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}';
EOT
