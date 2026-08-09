#!/usr/bin/env bash
# verify_228_deployed_page.sh <url> <must-contain-regex> <must-not-contain-regex> <control-regex> [deployed-at-domain] [deployed-at-page-name]
#
# Verifies a single deployed artefact (page HTML or JS asset), observing
# both LANDMINES.md traps a naive curl-after-dispatch hits:
#
# 1. "Verifying a page straight after firing its rerender shows a 404 or
#    stale copy, and both look like you broke the site" -- if a
#    domain+page-name are given, checks pages.deployed_at FIRST and waits
#    out the deploy window rather than trusting an immediate fetch.
#    Cache-busts every request. Polls on a REMOVED string, an ADDED string,
#    AND an untouched CONTROL string together -- "new string absent" alone
#    is indistinguishable from "page broken" without the control.
# 2. "`curl | grep` twice against the same URL during a deploy reports a
#    regression that never happened" -- fetches ONCE per iteration to a
#    file, then greps that SAME file for every property. Never re-curls to
#    check a second thing.
#
# Examples (see RUNBOOK for the exact values used for bugs_open/228):
#   ./verify_228_deployed_page.sh \
#     'https://robot-hands.com/contact.html' \
#     'action="mailto:robot-hands@contactforsales\.com' \
#     'x-never-matches-x' \
#     'cb-form|contact-block-section' \
#     robot-hands.com contact
#   ./verify_228_deployed_page.sh \
#     'https://robot-hands.com/tools/assets/contact-block.js' \
#     'Opening your email client' \
#     'setTimeout|has been sent' \
#     'cb-contact-form|validateEmail'

set -euo pipefail

URL="$1"; MUST_CONTAIN="$2"; MUST_NOT_CONTAIN="$3"; CONTROL="$4"
DOMAIN="${5:-}"; PAGE_NAME="${6:-}"
TMPFILE=$(mktemp)
trap 'rm -f "$TMPFILE"' EXIT

PSQL() {
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -tA "$@"
}

if [ -n "$DOMAIN" ] && [ -n "$PAGE_NAME" ]; then
  echo "== deploy-window check: pages.deployed_at for $DOMAIN/$PAGE_NAME =="
  DEPLOYED_AT=$(PSQL -c "
    SELECT p.deployed_at FROM pages p JOIN sites s ON s.id = p.site_id
    WHERE s.domain = '$DOMAIN' AND p.name = '$PAGE_NAME';
  ")
  echo "  deployed_at = $DEPLOYED_AT"
  AGE_S=$(( $(date -u +%s) - $(date -u -d "$DEPLOYED_AT" +%s 2>/dev/null || echo 99999) ))
  if [ "$AGE_S" -lt 120 ] && [ "$AGE_S" -ge 0 ]; then
    echo "  ${AGE_S}s old -- inside the deploy window (landmine: a fetch here proves nothing)."
    echo "  Waiting 90s before the first read..."
    sleep 90
  fi
fi

echo "== polling $URL (cache-busted, ONE fetch per iteration, re-grepped for every property) =="
ATTEMPTS=0
MAX_ATTEMPTS=10
SEP="?"
case "$URL" in *\?*) SEP="&" ;; esac
while [ "$ATTEMPTS" -lt "$MAX_ATTEMPTS" ]; do
  ATTEMPTS=$((ATTEMPTS + 1))
  curl -s "${URL}${SEP}cb=$(date +%s%N)" -o "$TMPFILE"

  MUST_CONTAIN_OK=0
  grep -qE "$MUST_CONTAIN" "$TMPFILE" && MUST_CONTAIN_OK=1
  MUST_NOT_CONTAIN_OK=1
  grep -qE "$MUST_NOT_CONTAIN" "$TMPFILE" && MUST_NOT_CONTAIN_OK=0
  CONTROL_OK=0
  grep -qE "$CONTROL" "$TMPFILE" && CONTROL_OK=1

  echo "  attempt $ATTEMPTS: must_contain=$MUST_CONTAIN_OK must_not_contain_clean=$MUST_NOT_CONTAIN_OK control=$CONTROL_OK (bytes=$(wc -c < "$TMPFILE"))"

  if [ "$MUST_CONTAIN_OK" = "1" ] && [ "$MUST_NOT_CONTAIN_OK" = "1" ] && [ "$CONTROL_OK" = "1" ]; then
    echo "PASS: $URL"
    exit 0
  fi
  if [ "$CONTROL_OK" = "0" ]; then
    echo "  control absent too -- could be a genuine mid-write 404/stale-empty response, retrying..."
  fi
  sleep 20
done

echo "FAIL after $MAX_ATTEMPTS attempts (~$((MAX_ATTEMPTS * 20))s of polling) for $URL" >&2
cp "$TMPFILE" "${TMPFILE}.last" 2>/dev/null || true
echo "Last fetch saved at ${TMPFILE}.last" >&2
exit 1
