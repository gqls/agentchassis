#!/usr/bin/env bash
# Put the OWNER-APPROVED privacy copy onto noted.co.uk/privacy.html — verbatim.
#
# WHY THIS EXISTS. The page-content-writer wrote its own privacy prose (measured
# 2026-08-13: 0 of the approved copy's 22 sentences appeared). Root cause is
# precise: the writer's prompt template injects ONLY the writer_block STRING
# ({{.site_specs.specs.evidence_base.writer_block}}), and the 08-12 registration
# told the writer to use copy "under supplied_copy.privacy" — a JSON path that
# NEVER TRAVELS to the prompt. An instruction pointing at data outside the
# reader's context is wired to nothing.
#
# This script sets the rendered page's content_data via the section editor's
# content_edit (update content_data -> re-render template -> update
# rendered_html -> reassemble page -> git commit). Same proven path as 074.
#
# THE COPY IS EXTRACTED FROM THE DRAFT FILE at run time — the single source of
# truth, same as apply_privacy_copy.py — so this script never carries a second
# copy that can drift.
#
# ⚠ RE-RUN THIS AFTER ANY REGENERATION of the privacy page. A rerender MERGES
# content_data (these values survive); a REGENERATION replaces it and the writer
# writes its own prose again (memory: bugfix 238). The durable half of the fix
# is the copy embedded INLINE in writer_block (embed_privacy_copy_in_writer_block.py);
# this script is the immediate half and the re-arm path.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
DRAFT="${REPO_ROOT}/docs/agent_docs/docs024_key_docs_latest/noted_rebuild/COPY_2026-08-12_privacy_DRAFT_for_owner.md"
DOMAIN="noted.co.uk"
CLIENT_ID="demo_client"

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA)

WORKFLOW_COMPACT='{"start_step":"spawn_section_editor","processing_mode":"orchestrator","timeout_seconds":900,"steps":{"spawn_section_editor":{"action":"spawn_agent","config":{"role":"section_editor","agent_type":"section-editor"},"output_field":"section_editor_agent","next_step":"call_section_editor","description":"Spawn section-editor agent"},"call_section_editor":{"action":"call_agent","config":{"agent_type":"section-editor","target_role":"section_editor","input_mapping":{"domain":"input_data.domain","edit_type":"input_data.edit_type","page_name?":"input_data.page_name","slot_name?":"input_data.slot_name","field_updates?":"input_data.field_updates"},"timeout_seconds":600},"output_field":"edit_result","next_step":"complete","description":"Run section edit"},"complete":{"action":"complete_workflow","config":{"output_fields":["edit_result"]},"description":"Section edit complete"}}}'

# Build both field_updates payloads from the draft. Bold lead-ins become
# <strong>; the mailto link is presentation only — the words stay verbatim.
build_payloads() {
  python3 - "$DRAFT" <<'PY'
import json, re, sys
body = open(sys.argv[1], encoding="utf-8").read()
body = body.split("## THE DRAFT", 1)[1].split("## Verification", 1)[0]
body = body.split("### Your notes, and what happens to them", 1)[1].strip()

paras = [" ".join(p.split()) for p in re.split(r"\n\s*\n", body) if p.strip()]
intro, sections = paras[0], paras[1:]

def para_html(p):
    p = re.sub(r"\*\*(.+?)\*\*", r"<strong>\1</strong>", p)
    p = p.replace("noted@contactforsales.com",
                  '<a href="mailto:noted@contactforsales.com">noted@contactforsales.com</a>', 1)
    return "<p>" + p + "</p>"

hero = {"subheadline": intro}
text = {"heading": "What that means in practice",   # fragment of the copy's own sentence; template's <h2> is unconditional
        "content": "".join(para_html(p) for p in sections)}
print(json.dumps(hero, ensure_ascii=False))
print(json.dumps(text, ensure_ascii=False))
PY
}

fire_one() {
  local page="$1" slot="$2" updates="$3"
  local CORRELATION_ID ORCHESTRATION_ID REQUEST_ID MESSAGE_ID TIMESTAMP
  CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
  ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
  REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
  MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  echo "--- ${page} / ${slot}   corr=${CORRELATION_ID}"

  BODY=$(python3 - "$WORKFLOW_COMPACT" "$DOMAIN" "$page" "$slot" "$updates" \
         "$CORRELATION_ID" "$ORCHESTRATION_ID" "$REQUEST_ID" "$MESSAGE_ID" "$CLIENT_ID" "$TIMESTAMP" <<'PY'
import json, sys
wf, domain, page, slot, updates, corr, orch, req, mid, client, ts = sys.argv[1:12]
msg = {"headers": {"correlation_id": corr, "orchestration_id": orch, "request_id": req,
                   "message_id": mid, "message_type": "request", "client_id": client,
                   "action": "process",
                   "sender": {"agent_id": "cli-user", "agent_type": "cli", "pod_name": "cli"},
                   "timestamp": ts},
       "config": {"workflow": json.loads(wf)},
       "input_data": {"domain": domain, "page_name": page, "slot_name": slot,
                      "edit_type": "content_edit", "field_updates": json.loads(updates)}}
print(json.dumps(msg, separators=(",", ":"), ensure_ascii=False))
PY
)

  kubectl -n kafka run -i --rm --quiet "kcat-privcopy-$(date +%s)-$RANDOM" \
    --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -c 1 \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID -H orchestration_id=$ORCHESTRATION_ID \
    -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
    -H message_type=request -H client_id=$CLIENT_ID -H action=process \
    -H sender_agent_type=cli -H sender_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses -H timestamp=$TIMESTAMP \
    >/dev/null <<<"$BODY"

  local i st
  for i in $(seq 1 40); do
    st=$("${PSQL[@]}" -c "SELECT status||'/'||current_step FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
    if [ -n "$st" ]; then
      case "$st" in
        COMPLETED/*) echo "    -> $st"; return 0 ;;
        FAILED/*)    echo "    -> $st  (do NOT cancel; diagnose)"; return 1 ;;
      esac
    fi
    sleep 6
  done
  echo "    -> no terminal status after 4 min — check before re-firing"
  return 1
}

{ read -r HERO_JSON; read -r TEXT_JSON; } < <(build_payloads)
echo "hero payload:  ${#HERO_JSON} bytes"
echo "text payload:  ${#TEXT_JSON} bytes"

fire_one privacy hero "$HERO_JSON"
fire_one privacy generic-text-block "$TEXT_JSON"

cat <<'NOTES'

=== Verify at the artefact — the sentence diff, not the status ===
Re-run the verbatim check (expects 22/22):
  see NOTES 2026-08-13, or diff rendered_html against the draft's sentences.
Then after sitesync's next tick, on the box:
  curl -s -H "Host: noted.co.uk" http://127.0.0.1:8082/privacy.html | grep -c "spell it out"
NOTES
