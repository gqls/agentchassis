#!/usr/bin/env bash
# Dynadot legacy API3 wrapper (JSON). Reads API_KEY from ~/.config/dynadot/credentials
# (one line: API_KEY=...) and never prints it — the key travels only inside curl's
# argv/query string, never to stdout, so transcripts stay clean.
#
# Usage: dynadot.sh <command> [param=value ...]
#   dynadot.sh list_domain
#   dynadot.sh domain_info domain=example.com
#   dynadot.sh get_ns domain=example.com
#   dynadot.sh set_ns domain=a.com,b.com ns0=alexis.ns.cloudflare.com ns1=leah.ns.cloudflare.com
#
# Gotchas (RUNBOOK_domains_cloudflare_rollout.md, "Dynadot"):
#   - set_ns targets must already exist in the account: add_ns them once, first.
#   - Rate tier Regular: 1 thread, 60 req/min — never run calls in parallel.
#   - A zone's Cloudflare NS pair comes from the zone-create response — the
#     account uses TWO pairs (29 alexis/leah, 11 betty/ivan as of 2026-08-25),
#     so never assume the pair when repointing.
#
# Exit codes: 0 = ResponseCode 0 (success); 1 = API returned an error (body is
# still printed); 2 = local setup problem (missing credentials / bad usage).
set -euo pipefail

CRED="${DYNADOT_CREDENTIALS:-$HOME/.config/dynadot/credentials}"
if [[ ! -r "$CRED" ]]; then
  echo "dynadot.sh: no credentials at $CRED (expected a line: API_KEY=...)" >&2
  exit 2
fi
API_KEY=$(sed -n 's/^API_KEY=//p' "$CRED" | tr -d '[:space:]')
if [[ -z "$API_KEY" ]]; then
  echo "dynadot.sh: API_KEY= line missing or empty in $CRED" >&2
  exit 2
fi

if [[ $# -lt 1 ]]; then
  echo "usage: dynadot.sh <command> [param=value ...]" >&2
  exit 2
fi
cmd="$1"; shift

params=(--data-urlencode "key=$API_KEY" --data-urlencode "command=$cmd")
for kv in "$@"; do
  params+=(--data-urlencode "$kv")
done

resp=$(curl -sS -G "https://api.dynadot.com/api3.json" "${params[@]}")
printf '%s\n' "$resp"

# Every API3 response carries a ResponseCode; "0" is the only success. Assert on
# it here so a caller cannot mistake an error body for a result (a receipt nobody
# asserts on is a log line).
if ! printf '%s' "$resp" | grep -qE '"ResponseCode" *: *"?0"?'; then
  echo "dynadot.sh: API error (ResponseCode != 0) — see body above" >&2
  exit 1
fi
