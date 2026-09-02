#!/usr/bin/env bash
# Create + wire a portfolio Cloudflare zone for domains already (or about to be)
# delegated to this account's nameserver pair. Encodes the recipe PROVEN 2026-08-25
# (homegarden.uk, domains_cloudflare_rollout RUNBOOK): zone -> 2 proxied A records
# at the TEST-NET-1 placeholder -> 2 portfolio-sites-router routes -> re-read and
# assert against the garden-tools.uk reference shape -> activation check.
#
# Written 2026-09-02 by the nominet_domain_management lane after the owner's
# Nominet NS batch left four domains delegated to alexis/leah with NO zone in the
# account (the LANDMINES "dangling delegation" trap): the registry answers, the
# edge REFUSES, the domain goes dark as caches expire.
#
# Credentials: ~/.config/cloudflare/portfoliotoken (one line). Read with
# tr -d '[:space:]' (trailing-newline gotcha), exported to curl only, never echoed.
# A session's permission classifier may refuse the WRITE calls — run --check
# from a session, and the mutating form from the owner's own prompt (`! ...`).
#
# Usage:
#   scripts/domains/cf-zone-bootstrap.sh --check <domain> [...]   # read-only report
#   scripts/domains/cf-zone-bootstrap.sh <domain> [...]           # create + wire + activate
set -euo pipefail

API=https://api.cloudflare.com/client/v4
ACC=13044f178ae0b156961065f55c8fada8
TOKEN_FILE="${CF_TOKEN_FILE:-$HOME/.config/cloudflare/portfoliotoken}"
CHECK_ONLY=0
[ "${1:-}" = "--check" ] && { CHECK_ONLY=1; shift; }
[ $# -ge 1 ] || { echo "usage: $0 [--check] <domain> [...]" >&2; exit 2; }
[ -r "$TOKEN_FILE" ] || { echo "FATAL: token file $TOKEN_FILE unreadable" >&2; exit 2; }
T=$(tr -d '[:space:]' < "$TOKEN_FILE")

cf() { # method path [json-body]
  local m=$1 p=$2 b=${3:-}
  if [ -n "$b" ]; then
    curl -sS -X "$m" -H "Authorization: Bearer $T" -H "Content-Type: application/json" --data "$b" "$API$p"
  else
    curl -sS -X "$m" -H "Authorization: Bearer $T" "$API$p"
  fi
}

# parse <json> <python-expr over d> — small, loud on API failure
parse() { python3 -c "
import sys, json
d = json.loads(sys.argv[1])
if not d.get('success', True):
    print('API-ERROR', d.get('errors'), file=sys.stderr); sys.exit(1)
print($2)" "$1"; }

FAIL=0
for D in "$@"; do
  echo "== $D"
  R=$(cf GET "/zones?name=$D")
  ZID=$(parse "$R" "d['result'][0]['id'] if d['result'] else ''")
  if [ "$CHECK_ONLY" = 1 ]; then
    if [ -z "$ZID" ]; then echo "NO-ZONE"; FAIL=1; continue; fi
    parse "$R" "'zone %s status=%s ns=%s' % (d['result'][0]['id'], d['result'][0]['status'], ','.join(d['result'][0]['name_servers']))"
    RECS=$(cf GET "/zones/$ZID/dns_records")
    parse "$RECS" "'\n'.join('record %s %s -> %s %s' % (r['type'], r['name'], r['content'], 'proxied' if r['proxied'] else 'DNS-ONLY') for r in d['result']) or 'NO-RECORDS'"
    RTS=$(cf GET "/zones/$ZID/workers/routes")
    parse "$RTS" "'\n'.join('route %s -> %s' % (r['pattern'], r.get('script')) for r in d['result']) or 'NO-ROUTES'"
    continue
  fi

  if [ -z "$ZID" ]; then
    R=$(cf POST /zones '{"name":"'"$D"'","account":{"id":"'"$ACC"'"},"type":"full"}')
    ZID=$(parse "$R" "d['result']['id']")
    parse "$R" "'created %s status=%s assigned-ns=%s' % (d['result']['id'], d['result']['status'], ','.join(d['result']['name_servers']))"
  else
    echo "zone exists: $ZID (continuing — records/routes are idempotently topped up)"
  fi

  # records: add whichever of apex/www is missing (top-up, never duplicate)
  HAVE=$(cf GET "/zones/$ZID/dns_records"); \
  for n in "$D" "www.$D"; do
    if ! parse "$HAVE" "'\n'.join(r['name'] for r in d['result'])" | grep -qx "$n"; then
      R=$(cf POST "/zones/$ZID/dns_records" '{"type":"A","name":"'"$n"'","content":"192.0.2.1","proxied":true,"ttl":1}')
      parse "$R" "'record %s -> %s proxied=%s' % (d['result']['name'], d['result']['content'], d['result']['proxied'])"
    fi
  done

  HAVER=$(cf GET "/zones/$ZID/workers/routes")
  for p in "$D/*" "www.$D/*"; do
    if ! parse "$HAVER" "'\n'.join(r['pattern'] for r in d['result'])" | grep -qxF "$p"; then
      R=$(cf POST "/zones/$ZID/workers/routes" '{"pattern":"'"$p"'","script":"portfolio-sites-router"}')
      parse "$R" "'route %s' % d['result']['pattern']"
    fi
  done

  # verify by RE-READING, never by POST receipts (RUNBOOK rule)
  RECS=$(cf GET "/zones/$ZID/dns_records")
  NREC=$(parse "$RECS" "sum(1 for r in d['result'] if r['type']=='A' and r['content']=='192.0.2.1' and r['proxied'])")
  NRT=$(parse "$(cf GET "/zones/$ZID/workers/routes")" "sum(1 for r in d['result'] if r.get('script')=='portfolio-sites-router')")
  echo "verify: $NREC/2 proxied placeholder A records, $NRT/2 router routes"
  [ "$NREC" = 2 ] && [ "$NRT" = 2 ] || { echo "VERIFY-FAILED: $D does not match the garden-tools.uk reference shape" >&2; FAIL=1; continue; }

  # activation: the PUT's 200 is acceptance, not activation (LANDMINES) — poll after
  cf PUT "/zones/$ZID/activation_check" >/dev/null || true
  for i in 1 2 3 4 5 6; do
    sleep 5
    ST=$(parse "$(cf GET "/zones/$ZID")" "d['result']['status']")
    echo "status: $ST (poll $i)"
    [ "$ST" = active ] && break
  done
  if [ "$ST" != active ]; then
    # a pending zone has two very different causes: propagation (clears alone)
    # and a PAIR MISMATCH (never clears) — measured 2026-09-02: this account
    # assigns betty/ivan to new zones while the estate's older zones sit on
    # alexis/leah, so a delegation copied from an old zone never activates.
    ASSIGNED=$(parse "$(cf GET "/zones/$ZID")" "','.join(sorted(d['result']['name_servers']))")
    REG=$(dig +norec NS "$D" @dns1.nic.uk +time=3 +tries=1 2>/dev/null \
          | awk '/\tNS\t/{print tolower($5)}' | sed 's/\.$//' | sort | paste -sd, -)
    echo "assigned-ns: $ASSIGNED"
    echo "registry-ns: ${REG:-unreadable (non-.uk or query failed — compare by hand)}"
    if [ -n "$REG" ] && [ "$REG" != "$ASSIGNED" ]; then
      echo "PAIR MISMATCH: the registry delegates to a DIFFERENT pair than Cloudflare assigned — this zone can NEVER activate until the registrar-side NS move to the assigned pair" >&2
      FAIL=1
    else
      echo "NOTE: still '$ST' — pairs match (or unverified); Cloudflare's own re-check usually clears it within the hour; re-run --check later"
    fi
  fi
done
exit $FAIL
