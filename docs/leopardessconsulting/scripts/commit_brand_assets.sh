#!/bin/bash
# ============================================================================
# commit_brand_assets.sh — commit the approved brand assets to a site's git repo
# through the platform's own git-adapter (topic system.adapter.git.requests).
#
# WHY THIS EXISTS (and not asset-deployer):
#   asset-deployer -> deploy_image_asset needs a storage client, built only when
#   IMAGE_BUCKET is set (platform/agentbase/agent.go: "Storage client not
#   configured (IMAGE_BUCKET not set)"). The agent-chassis deployment has AWS
#   creds but NOT IMAGE_BUCKET/S3_ENDPOINT, so deploy_image_asset ALWAYS fails
#   there with "storage client not available". This is a live platform bug —
#   83 of 102 active assets are decaying presigned URLs and most sites' logos
#   404. See docs/leopardessconsulting/AUDIT_verified_facts.md D8.
#
#   deploy_image_asset's only job after optimising is to send a git commit to the
#   git-adapter. We already have correctly sized, optimised files, so we send
#   that same message ourselves. Same adapter, same repo, same contract.
#
# The git-adapter prefixes every path with "{domain}/" and commits into the
# single "sites" repo (internal/adapters/git/github_client.go CommitToRepo).
# Commit -> GitHub Actions -> Backblaze B2 -> served via the Cloudflare Worker.
#
# Usage: ./commit_brand_assets.sh <domain> <brand_dir>
# ============================================================================
set -euo pipefail

DOMAIN="${1:?domain}"
BRAND_DIR="${2:?brand asset directory}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)

PAYLOAD_FILE=$(mktemp /tmp/git_commit_payload.XXXXXX.json)
trap 'rm -f "$PAYLOAD_FILE"' EXIT

# repo-relative path  <-  local file
#   logo:       header mark, transparent PNG
#   favicon:    browsers request /favicon.ico at the root by default
#   apple-touch: iOS home screen (opaque; iOS ignores alpha)
#   og-card:    social share preview
python3 - "$DOMAIN" "$BRAND_DIR" "$CORRELATION_ID" "$ORCHESTRATION_ID" "$REQUEST_ID" "$PAYLOAD_FILE" <<'PY'
import base64, json, os, sys
domain, brand, corr, orch, req, out = sys.argv[1:7]

MAPPING = {
    "assets/images/logo.png":    "logo.png",
    "favicon.ico":               "favicon.ico",
    "apple-touch-icon.png":      "apple-touch-icon.png",
    "assets/images/og-card.png": "og-card.png",
}

files = {}
for repo_path, local in MAPPING.items():
    p = os.path.join(brand, local)
    if not os.path.exists(p):
        sys.exit(f"missing brand asset: {p}")
    files[repo_path] = {
        "content": base64.b64encode(open(p, "rb").read()).decode(),
        "encoding": "base64",
    }
    print(f"  {repo_path:28s} {os.path.getsize(p)//1024:4d} KB", file=sys.stderr)

msg = {
    "headers": {
        "correlation_id": corr, "orchestration_id": orch, "request_id": req,
        "client_id": "demo_client", "step_name": "commit_brand_assets",
        "message_type": "request", "sender_agent_type": "user",
        "sender_agent_id": orch, "sender_pod_name": "cli",
        "responses_topic": "system.agent.generic.responses",
    },
    "body": {
        "action": "commit",
        "data": {
            "repo_name": "sites",
            "domain": domain,
            "files": files,
            "commit_message": f"brand: leopardess mark, favicon, apple-touch icon, OG card for {domain}",
        },
    },
}
# kcat splits stdin on newlines: the message MUST be exactly one line.
with open(out, "w") as f:
    f.write(json.dumps(msg, separators=(",", ":")))
    f.write("\n")
print(f"payload bytes: {os.path.getsize(out)}", file=sys.stderr)
PY

echo ""
echo "correlation: $CORRELATION_ID"
echo "committing to sites repo under ${DOMAIN}/ ..."

kubectl -n kafka run -i --rm "kcat-git-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.adapter.git.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_type=request \
  < "$PAYLOAD_FILE"

echo ""
echo "Verify by artifact (GitHub Actions -> B2 takes ~30-90s):"
echo "  kubectl -n ai-persona-system logs -l app=git-adapter --tail=40 | grep -i commit"
for p in assets/images/logo.png favicon.ico apple-touch-icon.png assets/images/og-card.png; do
  echo "  curl -s -o /dev/null -w '%{http_code} $p\\n' https://${DOMAIN}/$p"
done
