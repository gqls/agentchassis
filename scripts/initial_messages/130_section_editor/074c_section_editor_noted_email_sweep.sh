#!/usr/bin/env bash
# Sweep noted.co.uk's page components for the OLD contact address and re-render
# each through the section editor (owner 2026-08-17: email is
# noted@contactforsales.com). content_edit both updates content_data AND
# re-renders + re-assembles + commits, so this is the whole path per slot.
#
# The slot list is DISCOVERED at run time (any component whose content_data
# carries the old address), so a page added since this was written is swept too.
# Field_updates carry only the top-level fields that actually contained the
# address, with the replacement applied recursively (FAQ arrays and the like).
# The privacy page is EXCLUDED here: its copy is owner-approved verbatim and is
# re-rendered from the draft by 074b — run that separately.
set -euo pipefail

OLD="hello@noted.co.uk"; NEW="noted@contactforsales.com"
DOMAIN="noted.co.uk"; CLIENT_ID="demo_client"
PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA)

WORKFLOW='{"start_step":"spawn_section_editor","processing_mode":"orchestrator","timeout_seconds":900,"steps":{"spawn_section_editor":{"action":"spawn_agent","config":{"role":"section_editor","agent_type":"section-editor"},"output_field":"section_editor_agent","next_step":"call_section_editor","description":"Spawn section-editor agent"},"call_section_editor":{"action":"call_agent","config":{"agent_type":"section-editor","target_role":"section_editor","input_mapping":{"domain":"input_data.domain","edit_type":"input_data.edit_type","page_name?":"input_data.page_name","slot_name?":"input_data.slot_name","field_updates?":"input_data.field_updates"},"timeout_seconds":600},"output_field":"edit_result","next_step":"complete","description":"Run section edit"},"complete":{"action":"complete_workflow","config":{"output_fields":["edit_result"]},"description":"Section edit complete"}}}'

# one JSON line per target: {"page":..,"slot":..,"updates":{..}}
TARGETS=$("${PSQL[@]}" -c "
SELECT json_build_object('page', p.name, 'slot', pc.slot_name, 'cd', pc.content_data)::text
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='noted.co.uk' AND p.name <> 'privacy'
  AND pc.content_data::text LIKE '%${OLD}%';" | python3 -c "
import json, sys
OLD, NEW = '${OLD}', '${NEW}'
def rep(v):
    if isinstance(v, str): return v.replace(OLD, NEW)
    if isinstance(v, list): return [rep(x) for x in v]
    if isinstance(v, dict): return {k: rep(x) for k, x in v.items()}
    return v
for line in sys.stdin:
    line = line.strip()
    if not line: continue
    row = json.loads(line)
    updates = {k: rep(v) for k, v in row['cd'].items() if OLD in json.dumps(v)}
    if updates:
        print(json.dumps({'page': row['page'], 'slot': row['slot'], 'updates': updates}, ensure_ascii=False))
")

fire_one() {
  local page="$1" slot="$2" updates="$3"
  local C=$(cat /proc/sys/kernel/random/uuid) O=$(cat /proc/sys/kernel/random/uuid) R=$(cat /proc/sys/kernel/random/uuid) M=$(cat /proc/sys/kernel/random/uuid) T=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  echo "--- ${page} / ${slot}   corr=${C}"
  BODY=$(python3 - "$WORKFLOW" "$DOMAIN" "$page" "$slot" "$updates" "$C" "$O" "$R" "$M" "$CLIENT_ID" "$T" <<'PY'
import json, sys
wf, domain, page, slot, updates, corr, orch, req, mid, client, ts = sys.argv[1:12]
msg = {"headers": {"correlation_id": corr, "orchestration_id": orch, "request_id": req,
                   "message_id": mid, "message_type": "request", "client_id": client, "action": "process",
                   "sender": {"agent_id": "cli-user", "agent_type": "cli", "pod_name": "cli"}, "timestamp": ts},
       "config": {"workflow": json.loads(wf)},
       "input_data": {"domain": domain, "page_name": page, "slot_name": slot,
                      "edit_type": "content_edit", "field_updates": json.loads(updates)}}
print(json.dumps(msg, separators=(",", ":"), ensure_ascii=False))
PY
)
  kubectl -n kafka run -i --rm --quiet "kcat-em-$(date +%s)-$RANDOM" \
    --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -c 1 -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$C -H orchestration_id=$O -H request_id=$R -H message_id=$M \
    -H message_type=request -H client_id=$CLIENT_ID -H action=process \
    -H sender_agent_type=cli -H sender_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses -H timestamp=$T >/dev/null <<<"$BODY"
  for i in $(seq 1 40); do
    st=$("${PSQL[@]}" -c "SELECT status||'/'||current_step FROM orchestration_states WHERE correlation_id='${C}'::uuid LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
    case "$st" in COMPLETED/*) echo "    -> $st"; return 0;; FAILED/*) echo "    -> $st"; return 1;; esac
    sleep 6
  done
  echo "    -> no terminal status after 4 min"; return 1
}

rc=0; n=0
# fd 3, NOT stdin: fire_one's status poll runs `kubectl exec -i`, which inherits
# the loop's stdin and DRAINS the remaining targets as its own input — measured
# 2026-08-17: 6 discovered, 1 swept, no error anywhere.
while IFS= read -r -u 3 t; do
  [ -z "$t" ] && continue
  n=$((n+1))
  page=$(python3 -c "import json,sys;print(json.loads(sys.argv[1])['page'])" "$t")
  slot=$(python3 -c "import json,sys;print(json.loads(sys.argv[1])['slot'])" "$t")
  upd=$(python3 -c "import json,sys;print(json.dumps(json.loads(sys.argv[1])['updates'],ensure_ascii=False))" "$t")
  fire_one "$page" "$slot" "$upd" || rc=1
done 3<<< "$TARGETS"
echo "swept $n slot(s)"
exit $rc
