#!/bin/bash
# Read-only. Does a DISPATCHER still forward everything its handler declares?
#
# bugs_open/174: `diagnose-dispatch-loop` relays a work item spec to the handler
# named on the item, through TWO hand-maintained allow-lists on one path —
# `claim_item`'s SQL RETURNING projection, and `call_handler`'s `input_mapping`.
# They drifted from `diagnose-orchestrator`'s declared `input_contract` in
# lockstep, agreeing with each other and with nothing else, and `seed_scope` was
# dropped in silence. Three real diagnoses ran against a scope nobody chose, and
# nothing anywhere reported it: `input_mapping` is an allow-list, so an unlisted
# key is skipped at Info level, and the consuming action's scope fallback chain
# then supplied a different, plausible scope.
#
# This is the offline check — the same shape as audit-unregistered-actions.sh
# and for the same reason: the runtime says nothing about a key that was never
# there, so the only place to ask is the config itself, before a message finds
# out.
#
# WHAT THE THREE OUTPUT SECTIONS MEAN — they are not the same kind of thing:
#   findings                    a registered relay that CANNOT carry a key its
#                               handler declares. A defect. Exit 1.
#   unmatched_registry_entries  a relay we assert about that no longer matches
#                               the live config — the assertion silently stopped
#                               running. ALSO exit 1; this is the worse one.
#   uncovered_relays            a dispatcher-shaped relay nobody has registered.
#                               Advisory: it says the registry may be falling
#                               behind the fleet, not that anything is broken.
#
# Usage: scripts/audit-relay-gaps.sh [--json]
# Exit:  0 = clean · 1 = findings or unmatched entries · 2 = could not determine

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
JSON_OUT=0
[[ "${1:-}" == "--json" ]] && JSON_OUT=1

# stderr is deliberately NOT swallowed — see audit-config-keys.sh's identical note.
psql_q() {
    kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db -tAc "$1"
}

# Same population the runtime executes, PLUS input_contract — which is the
# authority this check compares against, and the one field the other audit modes
# do not need. WalkSteps (inside the Go binary) descends into loop sub-workflows
# itself; a hand-written descent here would be the bugs_open/144 mistake.
WORKFLOWS_JSON="$(psql_q "
SELECT jsonb_agg(jsonb_build_object(
         'type', type,
         'workflow', default_config->'workflow',
         'input_contract', input_contract))
FROM agent_definitions
WHERE deleted_at IS NULL
  AND COALESCE(is_snapshot,false) = false
  AND is_active
  AND default_config ? 'workflow';
")"
if [[ -z "$WORKFLOWS_JSON" || "$WORKFLOWS_JSON" == "null" ]]; then
    echo "ERROR: no live workflow definitions returned — query failed or the fleet is empty" >&2
    exit 2
fi

FINDINGS_JSON="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --relay-gaps) 2>/tmp/audit-relay-gaps.stderr)"
GO_EXIT=$?
if [[ $GO_EXIT -gt 1 ]]; then
    echo "ERROR: config-key-audit --relay-gaps failed:" >&2
    cat /tmp/audit-relay-gaps.stderr >&2
    exit 2
fi
cat /tmp/audit-relay-gaps.stderr >&2
rm -f /tmp/audit-relay-gaps.stderr

# REFUSE A RESULT THAT DID NOT COME FROM THIS MODE.
#
# config-key-audit dispatches on os.Args[1] and falls through to its DEFAULT
# report for an argument it does not recognise — which exits 0 and prints a
# perfectly valid JSON object about something else entirely. If --relay-gaps is
# ever unwired from main(), this script would otherwise report "clean" having
# checked nothing: the same silent-success shape as the bug it exists to catch.
# So assert the report's own keys before believing any of it.
if ! printf '%s' "$FINDINGS_JSON" | python3 -c '
import json, sys
try:
    r = json.load(sys.stdin)
except Exception:
    sys.exit(1)
sys.exit(0 if isinstance(r, dict) and {"findings","uncovered_relays","unmatched_registry_entries"} <= set(r) else 1)
'; then
    echo "ERROR: --relay-gaps did not return a relay-gap report." >&2
    echo "  config-key-audit falls through to its DEFAULT mode for an unrecognised" >&2
    echo "  argument, so this is almost certainly the mode not being wired into" >&2
    echo "  main()'s os.Args dispatch (cmd/config-key-audit/main.go). Refusing to" >&2
    echo "  report clean off a result that did not come from this check." >&2
    exit 2
fi

if [[ "$JSON_OUT" == "1" ]]; then
    printf '%s\n' "$FINDINGS_JSON"
else
    printf '%s' "$FINDINGS_JSON" | python3 -c '
import json, sys
r = json.load(sys.stdin)

print("=== RELAY GAPS (a dispatcher that cannot carry what its handler declares) ===")
if not r["findings"]:
    print("  none")
for f in r["findings"]:
    print("  {}.{} -> {}".format(f["caller"], f["step"], f["callee"]))
    if f["not_projected"]:
        print("     NOT PROJECTED by the claim query : {}".format(", ".join(f["not_projected"])))
        print("        (no input_mapping entry can carry these — fixing the mapping alone")
        print("         produces a source path that resolves to nothing, silently)")
    if f["not_forwarded"]:
        print("     NOT FORWARDED by input_mapping   : {}".format(", ".join(f["not_forwarded"])))
    if f["maps_to_nothing"]:
        print("     MAPPED BUT UNRESOLVABLE          : {}".format(", ".join(f["maps_to_nothing"])))

print()
print("=== UNMATCHED REGISTRY ENTRIES (an assertion that stopped running) ===")
if not r["unmatched_registry_entries"]:
    print("  none")
for u in r["unmatched_registry_entries"]:
    print("  {}".format(u))
if r["unmatched_registry_entries"]:
    print()
    print("  This is worse than a finding: a relay we claim to check is no longer")
    print("  being checked. Either the config moved (update declaredRelays in")
    print("  cmd/config-key-audit/relaygaps.go) or the agent is gone (drop the entry).")

print()
print("=== UNCOVERED DISPATCHER-SHAPED RELAYS (advisory) ===")
if not r["uncovered_relays"]:
    print("  none")
for u in r["uncovered_relays"]:
    print("  {}.{}  envelope={}".format(u["caller"], u["step"], u["envelope"]))
    print("     {}".format(u["reason"]))
if r["uncovered_relays"]:
    print()
    print("  These are NOT defects. They are relays nobody has asserted anything")
    print("  about. To register one you must first read its handler contract and")
    print("  confirm the callee — registering it unread would assert something")
    print("  nobody has checked, which is the state bugs_open/174 was already in.")
'
fi

exit $GO_EXIT
