#!/bin/bash
# ============================================================================
# reconcile_footer_nav.sh — propagate refreshed site chrome (header/footer) to
# every already-deployed page, and RECONCILE until each page actually serves it.
#
# WHY ASSEMBLE MODE, AND WHY IT IS THE LOAD-BEARING CHOICE.
# page-rerender regenerates section HTML from content_data ONLY when
# input_data.spec.reason is 'section_data_resolved'. This script deliberately
# omits it, so the else branch re-assembles the STORED section HTML with FRESH
# chrome — which is all a nav change needs.
#
# Sending the reason instead would run rerender_page_sections, whose pre-check
# escalates the WHOLE page to the automated content-writer when any section's
# content_data is missing a schema-required source:"llm" field. Measured on
# leopardessconsulting.co.uk, 2026-07-30: **5 of 34 active pages** would escalate,
# so a site-wide section rerender risks the writer rewriting five live pages.
# Chrome propagation must not carry that risk.
#
# WHAT THIS DOES NOT DO — and must not. It does not touch site_nav_items.
# `nav-updater` starts with populate_nav_tables, which does
# `DELETE FROM site_nav_items WHERE site_id = $1` and rebuilds from `pages` —
# and classifyPagesForNav SKIPS every page whose URL is under /tools/ (and
# /blog/, /guides/ …) unless its page_type is a section-index type. `tool` is
# not one. So running nav-updater on a site whose footer lists tool pages
# DELETES those links and puts none back. Verified by reading
# populate_nav_tables_action.go, 2026-07-30. Refresh chrome with
# `nav-link-fixer` (which has no populate step) and propagate with this.
#
# Publishing uses the container-COMMAND form with a PUBLISH_OK receipt: the
# `kubectl run -i --rm … kcat -P` stdin form loses ~4 of 5 messages at exit 0,
# and reconcile_headers.sh additionally sends both streams to /dev/null.
#
# Usage: ./reconcile_footer_nav.sh <site_id> <domain> <marker> [max_rounds]
#   marker: a string that must appear in each served page once chrome is current,
#           e.g. /tools/ai-vendor-trust-checklist.html
# ============================================================================
set -uo pipefail

S="${1:?site_id}"; DOMAIN="${2:?domain}"; MARKER="${3:?marker string}"
MAX_ROUNDS="${4:-3}"
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
PSQL=(kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -tAc)

# Owned pages are excluded: page-rerender's save_sections REFUSES a
# rebuild_policy='owned' page ("a generic section save would clobber it"), so
# firing at them only produces FAILED orchestrations. They are listed at the end
# so the operator can handle them deliberately.
mapfile -t PAGES < <("${PSQL[@]}" "SELECT name FROM pages WHERE site_id='$S' AND status='active' AND COALESCE(rebuild_policy,'generic') <> 'owned' ORDER BY nav_order, name")
mapfile -t OWNED < <("${PSQL[@]}" "SELECT name||' ('||url||')' FROM pages WHERE site_id='$S' AND status='active' AND rebuild_policy='owned' ORDER BY name")

echo "pages to reconcile: ${#PAGES[@]}   (owned, skipped: ${#OWNED[@]})"

fire() { # page_name -> publishes, echoes OK/DROPPED
  local PN="$1" CID OID B64 OUT PID
  CID=$(cat /proc/sys/kernel/random/uuid); OID=$(cat /proc/sys/kernel/random/uuid)
  # page_id is REQUIRED, not optional. The assemble branch goes straight to
  # rerender_single_page, which reads page_id/site_id/domain off the input and
  # errors "page_id not found in input" if it is absent — page_name is NOT
  # resolved for it. (Only the section-rerender branch resolves a name, via
  # spec.page_name.) reconcile_headers.sh sends page_name alone and therefore
  # fails this way on every page: 29 of 29 FAILED at render_page when this
  # script was first run, 2026-07-30.
  PID=$("${PSQL[@]}" "SELECT id FROM pages WHERE site_id='$S' AND name='$PN'")
  [ -n "$PID" ] || { echo "NO_PAGE_ID"; return; }
  B64=$(python3 -c '
import base64,json,sys
site,domain,page,pid = sys.argv[1:5]
# NO spec.reason -> assemble mode, no section regeneration, no escalation risk
msg={"action":"orchestrate","config":{"agent_type":"page-rerender"},
     "input_data":{"site_id":site,"domain":domain,"page_name":page,"page_id":pid}}
sys.stdout.write(base64.b64encode(json.dumps(msg,separators=(",",":")).encode()).decode())
' "$S" "$DOMAIN" "$PN" "$PID")
  OUT=$(kubectl -n kafka run "kcat-rfn-$(date +%s%N | tail -c 7)-$RANDOM" \
    --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
    --command -- sh -c "echo '$B64' | base64 -d | kcat -P -b $BROKER \
      -t system.agent.generic.requests \
      -H correlation_id=$CID -H orchestration_id=$OID \
      -H request_id=$(cat /proc/sys/kernel/random/uuid) \
      -H message_id=$(cat /proc/sys/kernel/random/uuid) \
      -H orchestration_name=rfn-$PN -H step_name=start -H client_id=demo_client \
      -H message_type=request -H action=orchestrate -H from_agent_type=user \
      -H from_agent_id=cli -H responses_topic=system.agent.generic.responses \
      && echo PUBLISH_OK" 2>/dev/null)
  case "$OUT" in *PUBLISH_OK*) echo "OK";; *) echo "DROPPED";; esac
}

serves_marker() { # page_name -> 0 if the DEPLOYED page already carries the marker
  local URL
  URL=$("${PSQL[@]}" "SELECT url FROM pages WHERE site_id='$S' AND name='$1'")
  curl -s "https://${DOMAIN}${URL}?cb=$(date +%s%N)" | grep -qF "$MARKER"
}

REMAINING=("${PAGES[@]}")
for (( round=1; round<=MAX_ROUNDS; round++ )); do
  # Only fire pages that do NOT already serve the marker — this is what makes it
  # a reconcile rather than a blind broadcast, and it is why a dropped publish
  # self-heals on the next round.
  TODO=(); for p in "${REMAINING[@]}"; do serves_marker "$p" || TODO+=("$p"); done
  echo ""
  echo "round $round: ${#TODO[@]} page(s) still missing the marker"
  [ ${#TODO[@]} -eq 0 ] && { echo "ALL PAGES CURRENT"; break; }

  dropped=0
  for p in "${TODO[@]}"; do
    r=$(fire "$p")
    printf '  %-42s %s\n' "$p" "$r"
    [ "$r" = "DROPPED" ] && dropped=$((dropped+1))
  done
  [ $dropped -gt 0 ] && echo "  !! $dropped publish(es) dropped — the next round re-fires them"

  REMAINING=("${TODO[@]}")
  echo "  waiting for renders + propagation ..."
  sleep 90
done

# Final census. Reports the COUNT measured, so "reconciled nothing" is
# distinguishable from "everything was already current".
still=0
for p in "${REMAINING[@]}"; do serves_marker "$p" || { still=$((still+1)); echo "  STILL MISSING: $p"; }
done
echo ""
echo "FINAL: $(( ${#PAGES[@]} - still )) of ${#PAGES[@]} reconcilable pages serve the marker"
if [ ${#OWNED[@]} -gt 0 ]; then
  echo ""
  echo "NOT ATTEMPTED — rebuild_policy='owned' (save_sections refuses these):"
  for o in "${OWNED[@]}"; do echo "  $o"; done
  echo "To refresh one deliberately: set rebuild_policy='generic', re-render, set it back"
  echo "to 'owned', then confirm the ownership refusal fires again."
fi
