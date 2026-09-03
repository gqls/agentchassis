#!/usr/bin/env bash
# OWNER RULING 2026-09-03: ALL Dynadot domains auto-renew, ALWAYS.
#
# Sweep: list every domain whose RenewOption != "auto-renew"; with --apply,
# flip each (set_renew_option renew_option=auto) and VERIFY the flip by
# re-reading domain_info — the receipt is not the proof. Dry-run by default.
#
# Why a sweep exists at all: the account default renew option was set to auto
# on 2026-09-03, but that write is RECEIPT-ONLY — account_info does not expose
# the default, so whether it covers marketplace acquisitions (both 09-02
# arrivals came in as "manual renewal" via Atom) is unverifiable until the next
# arrival. This sweep is the enforcement that does not depend on it. Run it
# whenever list_domain is being pulled anyway.
set -euo pipefail

apply=0
[[ "${1:-}" == "--apply" ]] && apply=1
sh="$(dirname "$0")/dynadot.sh"

off=$("$sh" list_domain 2>/dev/null | python3 -c '
import json, sys
for m in json.load(sys.stdin)["ListDomainInfoResponse"]["MainDomains"]:
    if m.get("RenewOption") != "auto-renew":
        print(m["Name"])')

if [[ -z "$off" ]]; then
  echo "all domains auto-renew — nothing to do"
  exit 0
fi
echo "NOT auto-renew:"
echo "$off"
if [[ $apply -eq 0 ]]; then
  echo "(dry run — pass --apply to flip)"
  exit 0
fi

fail=0
while read -r d; do
  "$sh" set_renew_option domain="$d" renew_option=auto >/dev/null
  sleep 1.2
  got=$("$sh" domain_info domain="$d" 2>/dev/null | python3 -c 'import json,sys; print(json.load(sys.stdin)["DomainInfoResponse"]["DomainInfo"]["RenewOption"])')
  if [[ "$got" == "auto-renew" ]]; then
    echo "flipped+verified: $d"
  else
    echo "FLIP NOT VERIFIED: $d reads '$got'" >&2
    fail=1
  fi
  sleep 1.2
done <<< "$off"
exit $fail
