#!/bin/bash
# Re-assemble + deploy idea.uk pages after a CHROME change (head/header/footer).
#
# WHY ASSEMBLE MODE (no spec.reason)
#   page-rerender with no `input_data.spec.reason` takes check_rerender_mode's ELSE
#   branch = assemble stored section HTML + the CURRENT chrome slots + deploy. That is
#   exactly a chrome-only change. Do NOT use section_data_resolved here: scoped mode
#   bails at page level (skipped / escalated, rerender_page_sections_action.go:157,:186)
#   and neither bail-out writes or deploys, so chrome changes silently miss those pages.
#   (Provenance: docs/leopardessconsulting/scripts/reassemble_pages.sh, corrected
#   2026-07-20 against bugs_closed/031.)
#
# ⚠ THE PUBLISH PATTERN IS LOAD-BEARING — DO NOT "SIMPLIFY" IT BACK TO STDIN.
#   `kubectl run -i --rm … kcat -P` loses ~4 of 5 messages at exit 0: kubectl attaches
#   stdin asynchronously, kcat reaches EOF first, produces nothing, exits clean, and
#   --rm deletes the evidence. Measured 2026-07-26. So the payload goes in the container
#   COMMAND and every publish must print PUBLISH_OK. `--command` is required because the
#   image ENTRYPOINT is kcat itself.
#
# Usage:  ./fire_reassemble_idea_uk.sh [page_name ...]     (default: all deployable)
set -uo pipefail

SITE=1244516d-014d-421c-88c6-090bb1e9552a
DOMAIN=idea.uk
BROKER=personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092
TOPIC=system.agent.generic.requests
PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc)

# Default set: active pages that have sections AND have actually shipped. A page with
# no sections assembles to nothing and is skipped by the action anyway; including it
# only adds noise. (idea.uk's `tool-audience-check` is such a stub — a #fragment of
# /tools.html with 0 sections and deployed_at NULL.)
if [ "$#" -gt 0 ]; then
  PAGES=("$@")
else
  mapfile -t PAGES < <("${PSQL[@]}" "
    SELECT p.name FROM pages p
     WHERE p.site_id='$SITE' AND p.status='active'
       AND EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id=p.id)
     ORDER BY p.name")
fi

echo "Re-assembling ${#PAGES[@]} page(s) for $DOMAIN"
printf '%s\n' "CORR|PAGE" > /tmp/idea_uk_reassemble_corrs.txt

ok=0; fail=0
for PN in "${PAGES[@]}"; do
  [ -z "$PN" ] && continue
  PID=$("${PSQL[@]}" "SELECT id FROM pages WHERE site_id='$SITE' AND name='$PN'")
  if [ -z "$PID" ]; then echo "  $PN: NO page_id — skipped"; fail=$((fail+1)); continue; fi

  CID=$(cat /proc/sys/kernel/random/uuid)
  OID=$(cat /proc/sys/kernel/random/uuid)
  RID=$(cat /proc/sys/kernel/random/uuid)
  MID=$(cat /proc/sys/kernel/random/uuid)

  # assemble mode: page_id is REQUIRED (render_page fails "page_id not found" without it)
  PAYLOAD=$(printf '{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"%s","domain":"%s","page_id":"%s","page_name":"%s"}}' \
    "$SITE" "$DOMAIN" "$PID" "$PN")

  OUT=$(kubectl -n kafka run "kcat-gtm-$(date +%s%N | tail -c 7)-$RANDOM" \
        --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
        --command -- sh -c "printf '%s' '$PAYLOAD' | kcat -P -b $BROKER -t $TOPIC \
          -H correlation_id=$CID -H orchestration_id=$OID -H request_id=$RID \
          -H message_id=$MID -H orchestration_name=gtm-${PN}-$(date +%H%M%S) \
          -H step_name=start -H client_id=demo_client -H message_type=request \
          -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
          -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK" 2>&1)

  if grep -q PUBLISH_OK <<<"$OUT"; then
    echo "  $PN  queued  corr=$CID"
    printf '%s|%s\n' "$CID" "$PN" >> /tmp/idea_uk_reassemble_corrs.txt
    ok=$((ok+1))
  else
    echo "  $PN  PUBLISH FAILED: $(tr '\n' ' ' <<<"$OUT" | cut -c1-200)"
    fail=$((fail+1))
  fi
done

echo
echo "published=$ok failed=$fail    correlations: /tmp/idea_uk_reassemble_corrs.txt"
