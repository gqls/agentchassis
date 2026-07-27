#!/usr/bin/env bash
# Publish ONE file to a site's repo through the git-adapter, without running a
# design pass.
#
# WHY THIS EXISTS
# ---------------
# `assets/css/styles.css` is written by exactly one path: webdesign-agent's
# deploy_css step (git_commit, file_path assets/css/styles.css). Reaching that
# step means running `analyze_design` first, and that step is an LLM that emits
# a fresh `color_scheme` every run which WINS over the palette row for the eight
# core slots (render_css_composition_helpers.go:corePaletteKeys). The pin that
# is supposed to hold it steady — design_intent.palette.reference_values — is
# handed to the model as "starting points, not exact targets ... you may adjust
# them", so it is advisory by construction. Measured on this site: the served
# stylesheet's core values differ from its own palette row in all five of
# background/surface/text/text_muted/border.
#
# So "regenerate the stylesheet" is not a safe way to apply a deterministic
# palette correction: it re-rolls the thing you are correcting. This publishes
# the file itself, and the palette row is corrected alongside so the next
# legitimate regeneration lands in the same place.
#
# Usage: deploy_stylesheet_direct.sh <domain> <local-css-file> [repo_path]
#   repo_path defaults to assets/css/styles.css
set -euo pipefail

DOMAIN="$1"; LOCAL="$2"; REPO_PATH="${3:-assets/css/styles.css}"
[ -s "$LOCAL" ] || { echo "refusing to publish an empty file: $LOCAL" >&2; exit 1; }

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "domain=$DOMAIN path=$REPO_PATH bytes=$(wc -c < "$LOCAL") corr=$CORR"

PAYLOAD=$(LOCAL="$LOCAL" DOMAIN="$DOMAIN" REPO_PATH="$REPO_PATH" \
          CORR="$CORR" ORCH="$ORCH" REQ="$REQ" TS="$TS" python3 - <<'PY'
import json, os
css = open(os.environ['LOCAL']).read()
h = {
  "correlation_id": os.environ['CORR'], "orchestration_id": os.environ['ORCH'],
  "orchestration_name": "operator-stylesheet-deploy", "parent_orchestration_id": "",
  "client_id": "demo_client", "step_name": "deploy_css", "step_id": "deploy_css",
  "request_id": os.environ['REQ'], "message_type": "request",
  "sender_agent_type": "cli", "sender_agent_id": "cli-user", "sender_pod_name": "cli",
  "sender_agent_version": "1", "sender_role": "operator",
  "responses_topic": "system.agent.generic.responses",
  "parent_responses_topic": "system.agent.generic.responses",
  "timestamp": os.environ['TS'], "action": "commit",
}
body = {
  "action": "commit",
  "data": {
    "repo_name": "sites", "domain": os.environ['DOMAIN'],
    "files": {os.environ['REPO_PATH']: css},
    "commit_message": "Correct palette slots and section defaults for contrast (operator)",
  },
  "reply_to_topic": "system.agent.generic.responses",
  "parent_responses_topic": "system.agent.generic.responses",
  "metadata": {"requesting_agent_id": os.environ['ORCH'], "requesting_agent_type": "cli",
               "requesting_step": "deploy_css", "client_id": "demo_client",
               "domain": os.environ['DOMAIN'], "file_count": 1},
  "request_context": {"correlation_id": os.environ['CORR'],
                      "orchestration_id": os.environ['ORCH'],
                      "request_id": os.environ['REQ']},
}
print(json.dumps({"headers": h, "body": body}))
PY
)

# kcat -P -c 1 reading a heredoc is the proven publish shape here: the payload
# travels on stdin, never in the container COMMAND (a large arg is where the
# silent-drop trap lives).
printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-css-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.adapter.git.requests \
  -H correlation_id=$CORR \
  -H orchestration_id=$ORCH \
  -H request_id=$REQ \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=commit \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TS

echo
echo "CORR=$CORR"
echo "Verify against the SERVED file, not the repo:"
echo "  curl -s https://$DOMAIN/$REPO_PATH | diff - $LOCAL && echo IDENTICAL"
