#!/bin/bash
# ============================================================================
# tool_acceptance_run.sh — fire tool-acceptance-agent at ONE tool: S6 of the
# staged build ladder (real clicks, real Chromium, desktop + mobile).
#
# The agent's own steps do the work: ensure_site_record -> load_docs ->
# request_run -> judge -> complete. load_docs reads the ```criteria fence from
# doc_plans using subject_key = input_data.spec.function, and request_run
# resolves the deployed URL from `pages`.
#
# THREE THINGS MUST LINE UP OR THE RUN IS A NO-OP, and two of them fail QUIETLY:
#
#  1. doc_plans.subject_key MUST equal <function>. If no PLAN is found the
#     fence is empty and request_browser_run SKIPS with reason=needs_criteria.
#     That is honest by design ("no fake pass") but it is not a failure either,
#     so a mistyped key looks like a clean run that asserted nothing.
#
#  2. pages.name MUST equal <function> or 'tool-'||<function>. The lookup is
#     `name IN ($2, 'tool-' || $2)`. A page named <function> MINUS the 'tool-'
#     prefix — the convention several sites use — matches NEITHER, and the step
#     hard-errors with "no deployed page URL". Measured 2026-07-30: 6 of 22
#     hosted tools fleet-wide are unresolvable this way.
#
#  3. Every check TYPE in the fence must be one the RUNNING browser-runner
#     binary implements. An unknown type is SKIPPED, not failed, and an
#     all-skipped result set reads as PASS plus a 7-day cooldown. Grep the pod
#     with a LONG marker before trusting a green run — short string literals are
#     compiled to immediate comparisons and never reach rodata, so a short
#     marker returns 0 on a binary that fully supports the type.
#
# Usage: ./tool_acceptance_run.sh <site_id> <domain> <function>
# ============================================================================
set -euo pipefail

SITE="${1:?site_id}"
DOMAIN="${2:?domain}"
FUNCTION="${3:?function (must match doc_plans.subject_key AND pages.name)}"

CID=$(cat /proc/sys/kernel/random/uuid)
OID=$(cat /proc/sys/kernel/random/uuid)
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

PAYLOAD_B64=$(python3 - "$SITE" "$DOMAIN" "$FUNCTION" <<'PY'
import base64, json, sys
site, domain, function = sys.argv[1:4]
msg = {
    "action": "orchestrate",
    "config": {"agent_type": "tool-acceptance-agent"},
    "input_data": {
        "site_id": site,
        "domain": domain,
        "spec": {"function": function},
    },
}
line = json.dumps(msg, separators=(",", ":"))
assert "\n" not in line
sys.stdout.write(base64.b64encode(line.encode()).decode())
PY
)

echo "function:    $FUNCTION"
echo "correlation: $CID"
echo "publishing to system.agent.generic.requests ..."

kubectl -n kafka run "kcat-ta-$(date +%s)-$RANDOM" \
  --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "echo '$PAYLOAD_B64' | base64 -d | kcat -P \
    -b $BROKER \
    -t system.agent.generic.requests \
    -H correlation_id=$CID \
    -H orchestration_id=$OID \
    -H request_id=$(cat /proc/sys/kernel/random/uuid) \
    -H message_id=$(cat /proc/sys/kernel/random/uuid) \
    -H orchestration_name=ta-$FUNCTION \
    -H step_name=start \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=user \
    -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"

echo ""
echo "No PUBLISH_OK above means NOTHING was published — re-run now."
echo "Follow it:"
echo "  SELECT current_step, status FROM orchestration_states WHERE correlation_id='$CID';"
echo "Read the per-check results — and CHECK FOR SKIPS, which are not passes:"
echo "  SELECT collected_data->'browser_run' FROM orchestration_states WHERE correlation_id='$CID';"
