#!/bin/bash
# ============================================================================
# commit_tool_asset.sh — commit ONE static asset to a site's git repo through
# the platform's own git-adapter (topic system.adapter.git.requests).
#
# Generalised from commit_brand_assets.sh, with ONE deliberate difference that
# is the whole reason this file exists:
#
#   commit_brand_assets.sh pipes the payload into `kubectl run -i --rm … kcat -P`
#   on STDIN. That pattern silently produces NOTHING roughly four times in five:
#   `kubectl run -i` attaches stdin asynchronously, so if the container reaches
#   kcat before stdin is wired it sees EOF, publishes nothing, exits 0, and
#   `--rm` deletes the evidence. Measured 2026-07-26 (four of five publishes
#   vanished) and it produces the same signature as queue latency and the ~300s
#   post-restart drop window: no rows anywhere.
#
#   So: the payload goes in the container COMMAND, and the publisher confirms
#   itself with PUBLISH_OK. No PUBLISH_OK in the output means nothing was
#   published — re-fire immediately instead of diagnosing a consumer.
#
#   `--command` is load-bearing: the image's ENTRYPOINT is kcat, so without it
#   the `sh -c …` arrives as arguments TO kcat and nothing is published.
#
# The payload is base64-encoded before it reaches the shell, so no quoting in
# the JSON (or in the file being committed) can break the command line. Base64
# is alphanumeric plus + / = and contains no quote characters at all.
#
# The git-adapter prefixes every path with "{domain}/" and commits into the
# single "sites" repo. Commit -> GitHub Actions -> B2 -> Cloudflare Worker.
#
# Usage: ./commit_tool_asset.sh <domain> <repo-relative-path> <local-file> <commit-message>
# ============================================================================
set -euo pipefail

DOMAIN="${1:?domain}"
REPO_PATH="${2:?repo-relative path, e.g. tools/assets/tool-x.js}"
LOCAL_FILE="${3:?local file to commit}"
COMMIT_MSG="${4:?commit message}"

[[ -f "$LOCAL_FILE" ]] || { echo "missing local file: $LOCAL_FILE" >&2; exit 1; }

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

PAYLOAD_B64=$(python3 - "$DOMAIN" "$REPO_PATH" "$LOCAL_FILE" "$COMMIT_MSG" \
                       "$CORRELATION_ID" "$ORCHESTRATION_ID" "$REQUEST_ID" <<'PY'
import base64, json, os, sys
domain, repo_path, local, msg, corr, orch, req = sys.argv[1:8]

raw = open(local, "rb").read()
payload = {
    "headers": {
        "correlation_id": corr, "orchestration_id": orch, "request_id": req,
        "client_id": "demo_client", "step_name": "commit_tool_asset",
        "message_type": "request", "sender_agent_type": "user",
        "sender_agent_id": orch, "sender_pod_name": "cli",
        "responses_topic": "system.agent.generic.responses",
    },
    "body": {
        "action": "commit",
        "data": {
            "repo_name": "sites",
            "domain": domain,
            "files": {
                repo_path: {
                    "content": base64.b64encode(raw).decode(),
                    "encoding": "base64",
                }
            },
            "commit_message": msg,
        },
    },
}
line = json.dumps(payload, separators=(",", ":"))
assert "\n" not in line, "payload must be exactly one line"
print(f"  {repo_path}  <-  {local}  ({len(raw)} bytes)", file=sys.stderr)
print(f"  payload {len(line)} bytes", file=sys.stderr)
sys.stdout.write(base64.b64encode(line.encode()).decode())
PY
)

echo ""
echo "correlation: $CORRELATION_ID"
echo "publishing to system.adapter.git.requests ..."

kubectl -n kafka run "kcat-tool-asset-$(date +%s)-$RANDOM" \
  --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "echo '$PAYLOAD_B64' | base64 -d | kcat -P \
    -b $BROKER \
    -t system.adapter.git.requests \
    -H correlation_id=$CORRELATION_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_type=request && echo PUBLISH_OK"

echo ""
echo "If PUBLISH_OK did not appear above, NOTHING was published — re-run now."
echo "Then verify at the ARTEFACT (GitHub Actions -> B2 is ~30-90s; a first-try"
echo "404 or a stale body is normal, retry once with a cache-buster):"
echo "  curl -s -o /dev/null -w '%{http_code}\\n' https://${DOMAIN}/${REPO_PATH}?cb=\$(date +%s)"
