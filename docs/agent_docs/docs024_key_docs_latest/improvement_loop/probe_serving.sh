#!/usr/bin/env bash
# Which estate domains are actually SERVING our site?  (improvement_loop lane, 2026-09-02)
#
# Classification is from the SERVED BYTES, never from sites.status — both domains this
# script was written to find are recorded `deployed`.
#
# ⚠ THE CONTROL IS THE POINT. Each domain gets TWO fetches: its root, and an invented
# path on the same domain. A parked domain answers 200 to EVERY path, so a bare root
# probe reports it healthy. The control is what separates "our page" from "this host
# answers 200 to anything".
#
# Usage:  ./probe_serving.sh            # reads domains from the live DB
#         ./probe_serving.sh domains.txt
set -uo pipefail

PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
LIST="${1:-}"
if [ -z "$LIST" ]; then
  LIST=$(mktemp)
  $PSQL -A -t -c "SELECT domain FROM sites WHERE status IN ('active','deployed') ORDER BY domain;" > "$LIST"
fi

printf '%-34s %-6s %-9s %-6s %-9s %s\n' DOMAIN ROOT BYTES CTRL CTRLBYTES VERDICT
while read -r d; do
  [ -z "$d" ] && continue
  root=$(curl -sL -o /tmp/pr_root.html -w '%{http_code}' -m 25 "https://$d/" 2>/dev/null) || root=000
  rb=$(wc -c < /tmp/pr_root.html 2>/dev/null || echo 0)
  ctrl=$(curl -sL -o /tmp/pr_ctrl.html -w '%{http_code}' -m 25 "https://$d/zz-improvement-loop-control-zz.html" 2>/dev/null) || ctrl=000
  cb=$(wc -c < /tmp/pr_ctrl.html 2>/dev/null || echo 0)
  lander=$(grep -c -i '/lander' /tmp/pr_root.html 2>/dev/null) || lander=0

  if   [ "$root" = "000" ];                            then v="NO RESPONSE (dns/tls) — POINT"
  elif [ "$lander" -gt 0 ] && [ "$rb" -lt 2000 ];      then v="PARKED -> /lander — POINT"
  elif [ "$rb" -lt 2000 ];                             then v="STUB ${rb}b — INVESTIGATE"
  elif [ "$ctrl" = "200" ] && [ "$cb" -eq "$rb" ];     then v="CATCH-ALL (control identical) — check"
  elif [ "$root" != "200" ];                           then v="ROOT $root — INVESTIGATE"
  else                                                      v="SERVING"
  fi
  printf '%-34s %-6s %-9s %-6s %-9s %s\n' "$d" "$root" "$rb" "$ctrl" "$cb" "$v"
done < "$LIST"

echo
echo "For anything not SERVING, the next question is DELEGATION, not the A record:"
echo "  dig +short NS <domain>   — a serving estate domain returns the Cloudflare pair;"
echo "                             a parked one returns its marketplace's nameservers."
