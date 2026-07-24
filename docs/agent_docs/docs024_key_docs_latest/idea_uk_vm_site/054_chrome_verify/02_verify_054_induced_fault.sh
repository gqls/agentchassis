#!/usr/bin/env bash
# 02_verify_054_induced_fault.sh — behaviourally verify bugs_open/054 (chrome
# dead-control drop + escalate) by INDUCING the failing branch on a SCRATCH site.
#
# WHY (verify-the-failing-branch rule): a pod-grep proves the code deployed, not
# that it detects/acts. 054's whole genesis (018/041) was a fix that shipped without
# effect. So we make a real dead chrome control and watch the live binary drop it.
#
# WHAT IT DOES (all on a throwaway .invalid site, cleaned up at the end):
#   1. create scratch site  scratch-054-verify.invalid
#   2. assign the UNGATED header-with-search_pre_037 to its header slot — its
#      template has bare  <a href="{{.nav_link_N_url}}">  and  src="{{.logo_url}}"
#      that a spec-less scratch site cannot resolve -> dead controls at render.
#   3. dispatch the render-site-chrome agent (01_render_site_chrome_agent.sql) —
#      renders chrome ONLY, no deploy (049b kcat -P -c 1 envelope).
#   4. ASSERT: stored rendered_html has NO href=""/src="" (Half 1 dropped them) AND
#      keeps real header markup; a chrome_dead_control row exists at
#      needs_human_review with NO handler_agent (Half 2).
#   5. clean up scratch site + its work items.
#
# PREREQ: 01_render_site_chrome_agent.sql applied (agent live). Chassis carrying
# the 054 code (v1.0.1150+). Run from anywhere; needs kubectl.
set -euo pipefail

PSQL() { kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -c "$1"; }
DOMAIN="scratch-054-verify.invalid"
HEADER_COMPONENT="d44490cb-89ef-4657-bcbe-793d5861f81c"  # header-with-search_pre_037 (ungated)
KEEP="${KEEP:-0}"   # KEEP=1 leaves the scratch site for inspection

cleanup() {
  [ "$KEEP" = "1" ] && { echo "KEEP=1 — leaving scratch site $SITE_ID ($DOMAIN)"; return; }
  [ -n "${SITE_ID:-}" ] || return
  echo "--- cleanup: removing scratch site $SITE_ID ---"
  PSQL "DELETE FROM site_work_items WHERE site_id='$SITE_ID';" >/dev/null || true
  PSQL "DELETE FROM site_components WHERE site_id='$SITE_ID';" >/dev/null || true
  PSQL "DELETE FROM sites WHERE id='$SITE_ID';" >/dev/null || true
}
trap cleanup EXIT

echo "=== 054 induced-fault verification ==="

# Pre-flight: agent live + code live in the pod
AGENT_LIVE=$(PSQL "SELECT count(*) FROM agent_definitions WHERE type='render-site-chrome' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;")
[ "$AGENT_LIVE" = "1" ] || { echo "FAIL: render-site-chrome agent not live — apply 01_render_site_chrome_agent.sql first"; exit 1; }
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
SYM=$(kubectl exec -n ai-persona-system "$POD" -- sh -c "strings /app/agent-chassis | grep -c DropDeadURLControls" 2>/dev/null || echo 0)
echo "pod=$POD  DropDeadURLControls symbols=$SYM  (0 => 054 not in this build; abort)"
[ "$SYM" -ge 1 ] || { echo "FAIL: 054 code not in the running binary"; exit 1; }

# 1. scratch site (fresh each run — delete any leftover first)
PSQL "DELETE FROM site_work_items WHERE site_id IN (SELECT id FROM sites WHERE domain='$DOMAIN');" >/dev/null || true
PSQL "DELETE FROM site_components WHERE site_id IN (SELECT id FROM sites WHERE domain='$DOMAIN');" >/dev/null || true
PSQL "DELETE FROM sites WHERE domain='$DOMAIN';" >/dev/null || true
# INSERT then SELECT (not RETURNING): psql -tA prints the "INSERT 0 1" command tag
# on its own line alongside a RETURNING value, which pollutes a captured id.
PSQL "INSERT INTO sites (domain, status) VALUES ('$DOMAIN','draft');" >/dev/null
SITE_ID=$(PSQL "SELECT id FROM sites WHERE domain='$DOMAIN';" | head -1 | tr -d '[:space:]')
echo "scratch site_id=$SITE_ID"
[ -n "$SITE_ID" ] || { echo "FAIL: could not create scratch site"; exit 1; }

# 2. assign the ungated header component to the header slot (rendered_html NULL so render will run)
PSQL "INSERT INTO site_components (site_id, slot_name, component_id, build_status)
      VALUES ('$SITE_ID','header','$HEADER_COMPONENT','pending');" >/dev/null
echo "assigned header-with-search_pre_037 (ungated) to header slot"

# 3. dispatch render-site-chrome (049b envelope: kcat -P -c 1 + heredoc)
CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid); MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "dispatch corr=$CORR"
kubectl -n kafka run -i --rm "kcat-054v-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR -H orchestration_id=$ORCH -H request_id=$REQ -H message_id=$MSG \
  -H message_type=request -H client_id=demo_client -H action=orchestrate \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses -H timestamp=$TS <<JSON
{"action":"orchestrate","config":{"agent_type":"render-site-chrome"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN"}}
JSON

# 4. poll for the render to store rendered_html (force_rerender writes build_status='rendered')
echo "--- polling for chrome render (up to ~5 min) ---"
HTML=""
for i in $(seq 1 60); do
  HTML=$(PSQL "SELECT COALESCE(rendered_html,'') FROM site_components WHERE site_id='$SITE_ID' AND slot_name='header';")
  [ -n "$HTML" ] && break
  sleep 5
done

echo ""
echo "=== ASSERTIONS ==="
FAIL=0

if [ -z "$HTML" ]; then
  echo "FAIL: header never rendered (rendered_html still empty after ~5 min — dispatch may be queued; retry or check logs for corr=$CORR)"
  FAIL=1
else
  RENDERED_LEN=${#HTML}
  echo "rendered_html length: $RENDERED_LEN"
  # Half 1: no dead URL attributes survived
  EMPTY_HREF=$(PSQL "SELECT count(*) FROM site_components WHERE site_id='$SITE_ID' AND slot_name='header' AND rendered_html ~ '(href|src)=\"\"';")
  if [ "$EMPTY_HREF" = "0" ]; then echo "PASS Half 1: no href=\"\"/src=\"\" in rendered chrome (dead controls dropped)"; else echo "FAIL Half 1: $EMPTY_HREF empty URL attrs survived"; FAIL=1; fi
  # render actually happened (real markup, not a wipe): the header CTA class survives
  HAS_MARKUP=$(PSQL "SELECT (rendered_html ILIKE '%header%')::int FROM site_components WHERE site_id='$SITE_ID' AND slot_name='header';")
  if [ "$HAS_MARKUP" = "1" ]; then echo "PASS: real header markup present (render succeeded, not a wipe)"; else echo "FAIL: header markup missing — render may have failed"; FAIL=1; fi
fi

# Half 2: chrome_dead_control work item filed at needs_human_review, no handler
ROW=$(PSQL "SELECT status || '|' || COALESCE(handler_agent,'<none>') || '|' || COALESCE(spec->>'slot_name','') || '|' || COALESCE(spec->>'dead_url_fields','')
            FROM site_work_items WHERE site_id='$SITE_ID' AND item_type='chrome_dead_control' ORDER BY created_at DESC LIMIT 1;")
if [ -n "$ROW" ]; then
  echo "chrome_dead_control row: $ROW"
  STATUS=$(echo "$ROW" | cut -d'|' -f1); HANDLER=$(echo "$ROW" | cut -d'|' -f2)
  if [ "$STATUS" = "needs_human_review" ]; then echo "PASS Half 2: filed at needs_human_review"; else echo "FAIL Half 2: status=$STATUS (expected needs_human_review)"; FAIL=1; fi
  if [ "$HANDLER" = "<none>" ] || [ -z "$HANDLER" ]; then echo "PASS Half 2: no handler_agent"; else echo "FAIL Half 2: handler_agent=$HANDLER (expected none)"; FAIL=1; fi
else
  echo "FAIL Half 2: no chrome_dead_control work item filed for the scratch site"
  FAIL=1
fi

echo ""
if [ "$FAIL" = "0" ]; then echo "=== RESULT: PASS — 054 verified end-to-end on the live binary ==="; else echo "=== RESULT: FAIL (see above; set KEEP=1 to inspect the scratch site) ==="; fi
exit $FAIL
