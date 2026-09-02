#!/usr/bin/env bash
# Walk Dynadot's Dynappraisal (RESTful v2) over a domain list, RESUMABLY.
#
# Usage: dynadot-appraise-all.sh <domains.csv> <valuations.csv>
#   <domains.csv>    header + domain in column 1 (the inbound inventory CSV).
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
[[ -f "$out" ]] || echo "domain,valuation,currency,source" > "$out"
today=$(date +%F)
fetched=0; skipped=0

while IFS= read -r d; do
  if grep -q "^$d," "$out"; then skipped=$((skipped+1)); continue; fi
  resp=$("$client" GET "/restful/v2/domains/$d/appraisal" 2>&1)
  if [[ $? -ne 0 ]]; then
    echo "stopped at $d after $fetched fetches this run ($skipped already present); response:"
    printf '%s\n' "$resp" | tail -3
    break
  fi
  price=$(printf '%s' "$resp" | sed -n 's/.*"appraisal_price":"\$\{0,1\}\([^"]*\)".*/\1/p' | tr -d ',')
  if [[ -z "$price" ]]; then
    echo "UNPARSED response for $d — stopping rather than writing junk:" >&2
    printf '%s\n' "$resp" >&2
    break
  fi
  echo "$d,$price,USD,dynadot_dynappraisal_$today" >> "$out"
  fetched=$((fetched+1))
  [[ $((fetched % 25)) -eq 0 ]] && echo "progress: $fetched fetched this run"
  sleep 1.2
done < <(tail -n +2 "$src" | cut -d, -f1)

total=$(( $(wc -l < "$out") - 1 ))
echo "done: $fetched fetched this run, $skipped already present, $total total rows in $out"
