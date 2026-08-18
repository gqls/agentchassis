#!/bin/bash
# Give every portfolio zone a `www` that 301s to the apex (owner ruling 2026-08-18).
#
# ORDER MATTERS — `scripts/cloudflare/worker.js` must already carry the
# www->apex branch AND be deployed. This script only makes `www.<domain>` REACH
# the worker; the worker does the redirecting. Run it against an undeployed
# worker and you turn a clean DNS failure into a live 404, which is worse than
# no record at all (the object key is `<hostname><path>` and every object is
# stored under the bare domain).
#
# NOT every zone may be treated alike, and this is the whole reason the script
# exists rather than a for-loop (all measured 2026-08-18):
#   - 24 zones already carry a WILDCARD route `*.<domain>/*` to the worker, so
#     they need the DNS record ONLY — adding an explicit www route as well is
#     redundant.
#   - 12 zones have the apex route only, and need BOTH.
#   - Some zones are NOT served by this worker at all (no route to it). A
#     proxied A -> 192.0.2.1 there is a black hole: TEST-NET-1 never routes and
#     no worker intercepts, so the visitor gets a 522 instead of a redirect.
#     Those are SKIPPED, loudly. They need a different mechanism.
#   - Some zones ALREADY redirect www correctly, or deliberately send www
#     somewhere else (webdesign.uk -> webdesign.co.uk), or serve www off a
#     different host entirely. Touching those is a regression, not a fix.
#
# Classification is measured per run, never assumed — a zone's routes and its
# live www behaviour are both read before anything is written.
#
# Usage:
#   ./add_www_redirect.sh                        # dry run, all zones (DEFAULT)
#   ./add_www_redirect.sh --apply                # write
#   ./add_www_redirect.sh --apply <domain>...    # write, named zones only
set -uo pipefail

TOKEN_FILE=~/.config/cloudflare/portfoliotoken
WORKER=portfolio-sites-router
APEX_IP=192.0.2.1
APPLY=0; ZONES_WANTED=()
for a in "$@"; do
  case "$a" in
    --apply) APPLY=1 ;;
    -*) echo "unknown flag: $a" >&2; exit 2 ;;
    *) ZONES_WANTED+=("$a") ;;
  esac
done
[ -f "$TOKEN_FILE" ] || { echo "no token at $TOKEN_FILE" >&2; exit 1; }
T=$(tr -d '\n' < "$TOKEN_FILE")
api() { curl -s -H "Authorization: Bearer $T" -H "Content-Type: application/json" "$@"; }

curl -s "https://api.cloudflare.com/client/v4/zones?per_page=100" -H "Authorization: Bearer $T" \
 | python3 -c "
import sys,json
d=json.load(sys.stdin)
if not d.get('success'): sys.exit('zone list failed: %s' % d.get('errors'))
for z in sorted(d['result'],key=lambda x:x['name']): print(z['id'],z['name'])" > /tmp/_wwwzones.$$ || exit 1

printf '%-34s %-16s %s\n' DOMAIN ACTION WHY
printf '%-34s %-16s %s\n' ------ ------ ---
dns_added=0; route_added=0; skipped=0; failed=0; done_already=0

while read -r ZID D; do
  if [ ${#ZONES_WANTED[@]} -gt 0 ]; then
    m=0; for w in "${ZONES_WANTED[@]}"; do [ "$w" = "$D" ] && m=1; done; [ $m = 1 ] || continue
  fi

  # -- what routes exist, and do they cover www? --------------------------
  read -r cov_wild cov_www cov_any <<<"$(api "https://api.cloudflare.com/client/v4/zones/$ZID/workers/routes" | D="$D" W="$WORKER" python3 -c "
import sys,json,os
D=os.environ['D']; W=os.environ['W']
rs=json.load(sys.stdin).get('result',[])
mine=[r for r in rs if (r.get('script') or '')==W]
wild=any(r.get('pattern')=='*.%s/*'%D for r in mine)
www =any(r.get('pattern')=='www.%s/*'%D for r in mine)
print(int(wild), int(www), int(bool(mine)))")"

  # -- what does www do TODAY? (live, not inferred) -----------------------
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "https://www.$D/" 2>/dev/null)

  case "$code" in
    301|308) printf '%-34s %-16s %s\n' "$D" "skip" "www already redirects ($code)"; done_already=$((done_already+1)); continue ;;
    302)     printf '%-34s %-16s %s\n' "$D" "skip" "www already redirects ($code) — may be deliberate elsewhere, leave it"; done_already=$((done_already+1)); continue ;;
    200)     printf '%-34s %-16s %s\n' "$D" "skip" "www already SERVES a page — replacing it would be a regression"; skipped=$((skipped+1)); continue ;;
  esac
  if [ "$cov_any" = 0 ]; then
    printf '%-34s %-16s %s\n' "$D" "SKIP" "no route to $WORKER — a proxied A here is a black hole (522), needs another mechanism"
    skipped=$((skipped+1)); continue
  fi

  need_route=1; [ "$cov_wild" = 1 ] || [ "$cov_www" = 1 ] && need_route=0
  have_dns=$(api "https://api.cloudflare.com/client/v4/zones/$ZID/dns_records?name=www.$D" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print(len(d.get('result',[])) if d.get('success') else 'ERR')")
  case "$have_dns" in ''|*[!0-9]*) printf '%-34s %-16s %s\n' "$D" "UNREADABLE" "dns query failed — skipped"; failed=$((failed+1)); continue ;; esac
  need_dns=1; [ "$have_dns" = 0 ] || need_dns=0

  if [ "$need_dns" = 0 ] && [ "$need_route" = 0 ]; then
    printf '%-34s %-16s %s\n' "$D" "recheck" "record+route present but www answered $code — verify by hand"; skipped=$((skipped+1)); continue
  fi
  what="dns=$need_dns route=$need_route"
  if [ "$APPLY" = 0 ]; then
    printf '%-34s %-16s %s\n' "$D" "would-add" "$what (wildcard_route=$cov_wild, www_http_today=${code:-none})"; continue
  fi

  okflag=1
  if [ "$need_dns" = 1 ]; then
    r=$(api -X POST "https://api.cloudflare.com/client/v4/zones/$ZID/dns_records" \
      --data "{\"type\":\"A\",\"name\":\"www\",\"content\":\"$APEX_IP\",\"proxied\":true,\"comment\":\"www->apex 301 via $WORKER (owner ruling 2026-08-18)\"}" \
      | python3 -c "import sys,json;d=json.load(sys.stdin);print(1 if d.get('success') else 0)")
    [ "$r" = 1 ] && dns_added=$((dns_added+1)) || okflag=0
  fi
  if [ "$okflag" = 1 ] && [ "$need_route" = 1 ]; then
    r=$(api -X POST "https://api.cloudflare.com/client/v4/zones/$ZID/workers/routes" \
      --data "{\"pattern\":\"www.$D/*\",\"script\":\"$WORKER\"}" \
      | python3 -c "import sys,json;d=json.load(sys.stdin);print(1 if d.get('success') else 0)")
    [ "$r" = 1 ] && route_added=$((route_added+1)) || okflag=0
  fi
  [ "$okflag" = 1 ] && printf '%-34s %-16s %s\n' "$D" "added" "$what" \
                    || { printf '%-34s %-16s %s\n' "$D" "FAILED" "$what"; failed=$((failed+1)); }
done < /tmp/_wwwzones.$$
rm -f /tmp/_wwwzones.$$

echo
echo "dns_added=$dns_added route_added=$route_added already_ok=$done_already skipped=$skipped failed=$failed"
echo
echo "VERIFY — read the redirect, never the API response:"
echo "  for d in <domains>; do curl -s -o /dev/null -w \"\$d %{http_code} -> %{redirect_url}\\n\" https://www.\$d/; done"
echo "  expect: 301 -> https://<domain>/   (DNS can take a minute to propagate)"
