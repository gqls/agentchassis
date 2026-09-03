#!/usr/bin/env bash
# Walk Dynadot's Dynappraisal (RESTful v2) over a domain list, RESUMABLY.
#
# Usage: dynadot-appraise-all.sh <domains.csv> <valuations.csv>
#   <domains.csv>    CSV with a header row containing a "domain" column — found
#                    BY NAME, not position (the inventory CSV has it first, the
#                    priority list has it second; positional parsing is the
#                    OPP-013 paste trap). An OPTIONAL "proxy_domain" column,
#                    also found by name, switches PROXY MODE per row where it
#                    is non-empty: the API call appraises proxy_domain (e.g.
#                    the .com equivalent of a .co.uk name Dynappraisal does not
#                    cover — see RUNBOOK/LANDMINES), but the output row is
#                    still keyed on the real domain, resume dedup still keys on
#                    the real domain, and `source` names the proxy used so a
#                    proxied value is never mistaken for a direct appraisal.
#   <valuations.csv> created with header if absent; domains already present are
#                    SKIPPED, so re-running after the daily quota 429 resumes
#                    exactly where it stopped. Appraisal quota is PER DAY by
#                    account tier — measured 2026-09-02: 300/day on this account
#                    (stopped at success #301 with "Daily quota ... try again
#                    tomorrow"). A 429 stop is the expected end of a day's run,
#                    not an error.
set -uo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: dynadot-appraise-all.sh <domains.csv> <valuations.csv>" >&2
  exit 2
fi
src="$1"; out="$2"
client="$(dirname "$0")/dynadot-restful.sh"
col=$(head -1 "$src" | tr -d '\r' | tr ',' '\n' | grep -nx 'domain' | cut -d: -f1)
if [[ -z "$col" ]]; then
  echo "dynadot-appraise-all.sh: no 'domain' column in $src header: $(head -1 "$src")" >&2
  exit 2
fi
proxycol=$(head -1 "$src" | tr -d '\r' | tr ',' '\n' | grep -nx 'proxy_domain' | cut -d: -f1 || true)
[[ -f "$out" ]] || echo "domain,valuation,currency,source" > "$out"
today=$(date +%F)
fetched=0; skipped=0

while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  d=$(cut -d, -f"$col" <<<"$line" | tr -d '\r')
  [[ -z "$d" ]] && continue
  target="$d"
  proxy=""
  if [[ -n "$proxycol" ]]; then
    proxy=$(cut -d, -f"$proxycol" <<<"$line" | tr -d '\r')
    [[ -n "$proxy" ]] && target="$proxy"
  fi
  prefix="dynadot_dynappraisal"
  [[ -n "$proxy" ]] && prefix="dynadot_dynappraisal_proxy_via_${proxy}"

  if grep -q "^$d," "$out"; then skipped=$((skipped+1)); continue; fi
  resp=$("$client" GET "/restful/v2/domains/$target/appraisal" 2>&1)
  if [[ $? -ne 0 ]]; then
    echo "stopped at $d (target $target) after $fetched fetches this run ($skipped already present); response:"
    printf '%s\n' "$resp" | tail -3
    break
  fi
  raw=$(printf '%s' "$resp" | sed -n 's/.*"appraisal_price":"\([^"]*\)".*/\1/p')
  if [[ -z "$raw" ]]; then
    echo "UNPARSED response for $d (target $target) — stopping rather than writing junk:" >&2
    printf '%s\n' "$resp" >&2
    break
  fi
  price=$(printf '%s' "$raw" | tr -d '$,')
  if [[ ! "$price" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
    # Dynappraisal answers HTTP 200 with appraisal_price "$--" for TLDs it does
    # not cover (measured 2026-09-03: .co.uk/.org.uk/.me.uk, vs working .com/
    # .net/.uk) — a real, distinct outcome, not a parse failure. Write an
    # explicit marker (empty valuation+currency) so the domain still counts as
    # "present" for the resume skip, instead of writing the literal "--" as a
    # price or retrying it forever.
    echo "$d,,,${prefix}_no_appraisal_$today" >> "$out"
    echo "no appraisal for $d (target $target, raw '$raw') — marked, not retried"
    fetched=$((fetched+1))
    sleep 1.2
    continue
  fi
  echo "$d,$price,USD,${prefix}_$today" >> "$out"
  fetched=$((fetched+1))
  [[ $((fetched % 25)) -eq 0 ]] && echo "progress: $fetched fetched this run"
  sleep 1.2
done < <(tail -n +2 "$src")

total=$(( $(wc -l < "$out") - 1 ))
echo "done: $fetched fetched this run, $skipped already present, $total total rows in $out"
