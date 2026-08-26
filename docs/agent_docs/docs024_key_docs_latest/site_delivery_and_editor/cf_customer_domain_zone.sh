#!/usr/bin/env bash
# cf_customer_domain_zone.sh — Cloudflare zone + records + routes for ONE customer
# domain, wrapping the recipe WORKED AND VERIFIED 2026-08-25 (homegarden.uk) in
# domains_cloudflare_rollout/RUNBOOK. DRY-RUN DEFAULT; --apply executes.
#
# Part of the domain service (BRIEF_2026-08-26, P4), SEVERABLE BY DESIGN (owner
# ruling 2026-08-26): this script touches ONLY the new domain's own zone. The
# site keeps serving at <slug>.ugg2.com regardless; abandoning the programme
# means never running this again, and a created zone is one API DELETE away.
#
# Uses ~/.config/cloudflare/portfoliotoken (NOT 'token' — the 08-25 recipe's own
# instruction). Token caveats (LANDMINES): account-wide across all 36 zones, so
# this script NAMES the zone id it just created and touches nothing else; DNS
# comments capped at 100 chars; re-read `success` in each response, never the
# HTTP status.
#
# Usage: ./cf_customer_domain_zone.sh <domain> [--apply]
# Prints: the zone id and THE ASSIGNED NAMESERVER PAIR — capture it: that pair
# is what nominet-epp-domain-register.py --ns (or the ns-change client) must set.
set -euo pipefail

DOMAIN="${1:?usage: cf_customer_domain_zone.sh <domain> [--apply]}"
APPLY="${2:-}"
ACC=13044f178ae0b156961065f55c8fada8
REF_ZONE=82d90228c20877e2b3fc8470c2bc73d1   # garden-tools.uk, the known-good template zone
T=$(tr -d '[:space:]' < ~/.config/cloudflare/portfoliotoken)   # never echo it
API=https://api.cloudflare.com/client/v4

cf() { curl -s -H "Authorization: Bearer $T" -H "Content-Type: application/json" "$@"; }
ok() { python3 -c "import json,sys; d=json.load(sys.stdin); print(json.dumps(d['result'])) if d.get('success') else (sys.stderr.write(json.dumps(d.get('errors'))+'\n'), sys.exit(1))"; }

case "$DOMAIN" in
  *.uk) ;;
  *) echo "refusing '$DOMAIN': the offer is .uk / .co.uk only (owner ruling 2026-08-21)" >&2; exit 2 ;;
esac

if [ "$APPLY" != "--apply" ]; then
  cat <<EOF
DRY-RUN for $DOMAIN — would execute, in order:
  1. POST $API/zones                    {"name":"$DOMAIN","account":{"id":"$ACC"},"type":"full"}
  2. POST .../dns_records  x2           proxied A 192.0.2.1 for $DOMAIN and www.$DOMAIN
  3. POST .../workers/routes x2         $DOMAIN/* and www.$DOMAIN/* -> portfolio-sites-router
  4. GET  re-read records+routes, diff against reference zone $REF_ZONE (2 A, 2 routes)
Re-run with --apply. Reversal: DELETE $API/zones/<id>.
EOF
  exit 0
fi

echo "creating zone $DOMAIN ..." >&2
Z=$(cf -X POST --data "{\"name\":\"$DOMAIN\",\"account\":{\"id\":\"$ACC\"},\"type\":\"full\"}" "$API/zones" | ok)
ZID=$(printf '%s' "$Z" | python3 -c "import json,sys; print(json.load(sys.stdin)['id'])")
NS=$(printf '%s' "$Z" | python3 -c "import json,sys; print(' '.join(json.load(sys.stdin)['name_servers']))")
echo "zone_id=$ZID"
echo "ASSIGNED_NS=$NS   <- set THESE at Nominet (capture per-zone; the pair is convention, not contract)"

for n in "$DOMAIN" "www.$DOMAIN"; do
  cf -X POST --data "{\"type\":\"A\",\"name\":\"$n\",\"content\":\"192.0.2.1\",\"proxied\":true,\"ttl\":1}" \
    "$API/zones/$ZID/dns_records" | ok >/dev/null
done
for p in "$DOMAIN/*" "www.$DOMAIN/*"; do
  cf -X POST --data "{\"pattern\":\"$p\",\"script\":\"portfolio-sites-router\"}" \
    "$API/zones/$ZID/workers/routes" | ok >/dev/null
done

# Verify by RE-READING, never by trusting the POST receipts (the recipe's own rule).
REC=$(cf "$API/zones/$ZID/dns_records" | ok | python3 -c "import json,sys; rs=json.load(sys.stdin); print(sum(1 for r in rs if r['type']=='A' and r['proxied']))")
RTS=$(cf "$API/zones/$ZID/workers/routes" | ok | python3 -c "import json,sys; rs=json.load(sys.stdin); print(sum(1 for r in rs if r.get('script')=='portfolio-sites-router'))")
echo "re-read: proxied A records=$REC (expect 2), router routes=$RTS (expect 2)"
[ "$REC" = 2 ] && [ "$RTS" = 2 ] && echo "ZONE READY (status stays 'pending' until the NS move lands — that is not a failure)" || { echo "MISMATCH — inspect the zone before using it" >&2; exit 1; }
