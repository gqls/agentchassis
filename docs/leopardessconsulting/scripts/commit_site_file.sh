#!/bin/bash
# ============================================================================
# commit_site_file.sh — commit one or more files into a site's directory in the
# deploy repo, through the platform's own git-adapter
# (topic system.adapter.git.requests, action "commit").
#
# WHY THIS EXISTS ALONGSIDE commit_brand_assets.sh (same directory):
#   commit_brand_assets.sh publishes with `kubectl run -i --rm … kcat -P < file`.
#   `kubectl run -i` attaches stdin ASYNCHRONOUSLY, so if the container reaches
#   kcat before stdin is wired it sees EOF and publishes NOTHING at exit 0 —
#   measured 2026-07-26 at four of five publishes lost. This script carries the
#   payload in the container COMMAND instead and prints PUBLISH_OK, the same fix
#   rerender_page_safe.sh applies to the orchestration topic. Nothing else about
#   the contract differs.
#
# THREE THINGS THAT ARE LOAD-BEARING, each a recorded landmine:
#   1. repo_name. "sites" is right ONLY for a domain whose `sites.github_repo`
#      is NULL. A domain that names `vm-sites` (idea.uk, relojistas.com) takes a
#      GREEN commit into the WRONG repo and the served file never changes.
#      This script READS the column and refuses rather than guessing.
#   2. Branch is never set. `gqls/sites` has no `main`; CommitToRepo falls back
#      to the repo default (`master`), which is the branch the B2 workflow
#      watches. Passing `sites.github_branch` ('main' on most rows) commits to a
#      branch that does not exist, or worse, to one nothing deploys.
#   3. Repo paths are repo-relative and UNPREFIXED — `sitemap.xml`, not
#      `/sitemap.xml`. CommitToRepo builds `{domain}/{path}`, so a leading slash
#      yields `domain.com//sitemap.xml` and an empty tree segment. Refused below.
#
# A green adapter log is NOT evidence: an unchanged file commits as an EMPTY
# commit and still reports success:true. Verify at the SERVED artefact.
#
# Usage: ./commit_site_file.sh <domain> <commit_message> <repo_path>=<local_file> [...]
# ============================================================================
set -euo pipefail

DOMAIN="${1:?domain}"; shift
MESSAGE="${1:?commit message}"; shift
[ "$#" -ge 1 ] || { echo "need at least one <repo_path>=<local_file>" >&2; exit 1; }

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c)
REPO=$("${PSQL[@]}" "SELECT COALESCE(NULLIF(github_repo,''),'sites') FROM sites WHERE domain='${DOMAIN}'" | tr -d '[:space:]')
[ -n "$REPO" ] || { echo "no sites row for ${DOMAIN} — refusing to guess a repo" >&2; exit 1; }
echo "domain:      $DOMAIN"
echo "repo:        $REPO   (from sites.github_repo, defaulted to 'sites' only when NULL/empty)"

CID=$(cat /proc/sys/kernel/random/uuid)
OID=$(cat /proc/sys/kernel/random/uuid)
RID=$(cat /proc/sys/kernel/random/uuid)
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

PAYLOAD_B64=$(python3 - "$DOMAIN" "$REPO" "$MESSAGE" "$CID" "$OID" "$RID" "$@" <<'PY'
import base64, json, os, sys
domain, repo, message, corr, orch, req = sys.argv[1:7]
files = {}
for pair in sys.argv[7:]:
    if "=" not in pair:
        sys.exit(f"expected <repo_path>=<local_file>, got: {pair}")
    repo_path, local = pair.split("=", 1)
    if repo_path.startswith("/") or ".." in repo_path.split("/"):
        sys.exit(f"repo path must be repo-relative and unprefixed: {repo_path}")
    if not os.path.exists(local):
        sys.exit(f"missing local file: {local}")
    files[repo_path] = {
        "content": base64.b64encode(open(local, "rb").read()).decode(),
        "encoding": "base64",
    }
    print(f"  {repo_path:<28s} <- {local} ({os.path.getsize(local)} bytes)", file=sys.stderr)
msg = {
    "headers": {
        "correlation_id": corr, "orchestration_id": orch, "request_id": req,
        "client_id": "demo_client", "step_name": "commit_site_file",
        "message_type": "request", "sender_agent_type": "user",
        "sender_agent_id": orch, "sender_pod_name": "cli",
        "responses_topic": "system.agent.generic.responses",
    },
    "body": {"action": "commit", "data": {
        "repo_name": repo, "domain": domain, "files": files,
        "commit_message": message,
    }},
}
line = json.dumps(msg, separators=(",", ":"))
assert "\n" not in line
sys.stdout.write(base64.b64encode(line.encode()).decode())
PY
)

echo "correlation: $CID"
echo "publishing to system.adapter.git.requests ..."

kubectl -n kafka run "kcat-git-$(date +%s)-$RANDOM" \
  --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "echo '$PAYLOAD_B64' | base64 -d | kcat -P \
    -b $BROKER \
    -t system.adapter.git.requests \
    -H correlation_id=$CID \
    -H orchestration_id=$OID \
    -H request_id=$RID \
    -H message_type=request && echo PUBLISH_OK"

cat <<EOF

No PUBLISH_OK above means NOTHING was published — re-run now.

Then, in order (GitHub Actions -> B2 is ~30-90s):
  kubectl -n ai-persona-system logs -l app=git-adapter --tail=200 | grep -i "$CID"
  git -C ~/projects/$REPO pull --ff-only && git -C ~/projects/$REPO show --stat HEAD
     ^ subject-only output = EMPTY commit = nothing deployed
  curl -s "https://$DOMAIN/<path>?cb=\$(date +%s)" | head
EOF
