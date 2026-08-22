#!/usr/bin/env bash
# witness_contrast_ratio.sh — prove the DEPLOYED contrast_ratio check fires.
#
# Drives the browser-runner adapter DIRECTLY (system.adapter.browser-runner.requests,
# action=run_checks) and reads the reply off a throwaway topic. Deliberately NOT a
# tool-acceptance run: acceptance invokes the judge, which files improve_tool items —
# and the witness target is another lane's live surface with a design pass queued for
# the very tokens involved. This path files NOTHING; it is a browser visiting a public
# page and reporting back.
#
# THE CONTROL IS THE POINT. Three checks against ONE page, same probe, same run:
#   legible-aa      contrast_ratio, default WCAG AA   -> must FAIL, with a culprit
#   legible-floor   contrast_ratio, min_ratio 1.05    -> must PASS (same page!)
#   reachable       page_status_ok                    -> must PASS
# A check that always fails would fail all three. A check that never fails would pass
# all three. Only a check that actually measures produces FAIL/PASS/PASS — and the
# middle one additionally proves the deployed binary DECODES the new min_ratio field
# (an unknown JSON key is dropped in silence: LANDMINES "unknown key in a criteria
# fence is dropped at unmarshal").
set -euo pipefail

URL="${1:-https://vonc.com/tools/gauntlet/index.html}"
PROFILE="${2:-mobile}"
BOOT=personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092
REQ_TOPIC=system.adapter.browser-runner.requests
STAMP=$(date +%s)
REPLY_TOPIC="witness.contrast.${STAMP}"
RUN_ID=$(cat /proc/sys/kernel/random/uuid)

CRITERIA=$(jq -c -n '{
  profiles: ["desktop","mobile"],
  container: ".gauntlet-interface-section",
  checks: [
    {id:"reachable",     type:"page_status_ok"},
    {id:"legible-aa",    type:"contrast_ratio"},
    {id:"legible-floor", type:"contrast_ratio", min_ratio: 1.05}
  ]}')

# ONE LINE: kcat -P splits stdin on newlines into separate messages.
PAYLOAD=$(jq -c -n \
  --arg corr "$(cat /proc/sys/kernel/random/uuid)" \
  --arg reply "$REPLY_TOPIC" --arg run "$RUN_ID" \
  --arg url "$URL" --arg prof "$PROFILE" --arg crit "$CRITERIA" \
  '{headers: {correlation_id: $corr, orchestration_id: $corr, client_id: "witness",
              message_id: $corr, request_id: $corr, message_type: "request",
              action: "run_checks", responses_topic: $reply},
    body: {action: "run_checks", reply_to_topic: $reply,
           data: {run_id: $run, urls: [$url], profiles: [$prof],
                  criteria_json: $crit, function: "witness-131-contrast"}}}')

echo "witness: $URL  profile=$PROFILE  run_id=$RUN_ID  reply=$REPLY_TOPIC"

printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-witness-p-${STAMP}" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b "$BOOT" -t "$REQ_TOPIC" >/dev/null

echo "published; consuming reply (up to ~150s — a real browser run takes ~20-60s)…"
kubectl -n kafka run -i --rm "kcat-witness-c-${STAMP}" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -C -b "$BOOT" -t "$REPLY_TOPIC" -o beginning -e -q -u 2>/dev/null \
  | head -c 200000
