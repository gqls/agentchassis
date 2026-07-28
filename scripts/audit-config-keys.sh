#!/bin/bash
# Read-only. Which step-config keys in LIVE agent_definitions is no action known
# to read?
#
# bugs_open/101: unknown config keys are silently ignored, so a dead key looks
# exactly like a live one. The runtime validator warns about them — but only for
# actions that have opted in by declaring ActionInputSpec.ConfigKeys, and only
# for steps that actually run. This is the offline half: every live step, whether
# it has run recently or not.
#
# It reports TWO different things and does not conflate them, because the fix
# differs:
#
#   UNDECLARED ACTION — the action has not opted in at all. Nothing is known
#                       about its keys, so nothing here is evidence of a bug.
#                       The fix is to declare ConfigKeys on that action.
#   UNKNOWN KEY       — the action HAS declared its contract and this key is not
#                       in it. This is a real inert key: the config describes
#                       behaviour that does not happen.
#
# A count of zero in the second section is meaningful. A count of zero in the
# first would mean full adoption, which is not the case today and is the number
# this report exists to drive down.
#
# Usage: scripts/audit-config-keys.sh [--json]
# Exit:  0 = no unknown keys · 1 = unknown keys found · 2 = could not determine

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
JSON_OUT=0
[[ "${1:-}" == "--json" ]] && JSON_OUT=1

# stderr is deliberately NOT swallowed. Silencing it would turn a broken query
# into an empty section that reads exactly like "nothing to report" — the same
# silent no-op this whole report exists to expose.
psql_q() {
    kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db -tAc "$1"
}

# What the binary declares. Asking the binary rather than grepping the source is
# the point: the declarations are Go, registered by init(), and a grep would
# quietly disagree with the running code the first time someone builds a list
# dynamically.
DECLARED="$(cd "$REPO_ROOT" && go run ./cmd/config-key-audit 2>/dev/null)"
if [[ -z "$DECLARED" ]]; then
    echo "ERROR: could not read declared config keys (go run ./cmd/config-key-audit failed)" >&2
    exit 2
fi

# What live definitions actually carry. jsonb_typeof(...)='object' is required:
# some steps carry a STRING config (a reference, not a literal) and
# jsonb_object_keys errors on those rather than skipping them.
LIVE="$(psql_q "
SELECT DISTINCT v->>'action' || E'\t' || ck.key
FROM agent_definitions ad,
     jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v),
     jsonb_object_keys(v->'config') AS ck(key)
WHERE ad.deleted_at IS NULL
  AND COALESCE(ad.is_snapshot,false) = false
  AND ad.is_active
  AND v->'config' IS NOT NULL
  AND jsonb_typeof(v->'config') = 'object'
  AND v->>'action' IS NOT NULL
ORDER BY 1;
")"
if [[ -z "$LIVE" ]]; then
    echo "ERROR: no live step config returned — query failed or the fleet is empty" >&2
    exit 2
fi

DECLARED="$DECLARED" LIVE="$LIVE" JSON_OUT="$JSON_OUT" python3 - <<'PY'
import json, os, sys
from collections import defaultdict

declared = json.loads(os.environ["DECLARED"])
# Framework keys are read by the orchestrator on ANY step, whatever the action.
# Kept in step with datahelpers.frameworkStepConfigKeys — if these drift apart
# this report grows false positives, which is how a report gets ignored.
framework = {
    "action", "agent_type", "continue_on_error", "error_step", "input_fields",
    "loop_item_index", "loop_iteration", "loop_name", "loop_var_name",
    "output_mapping", "role", "target_action", "timeout_seconds",
    "total_iterations",
}

unknown_by_action = defaultdict(list)
undeclared_by_action = defaultdict(list)

for line in os.environ["LIVE"].splitlines():
    line = line.strip()
    if not line or "\t" not in line:
        continue
    action, key = line.split("\t", 1)
    if key in framework:
        continue
    if action in declared:
        if key not in declared[action]:
            unknown_by_action[action].append(key)
    else:
        undeclared_by_action[action].append(key)

if os.environ["JSON_OUT"] == "1":
    json.dump({
        "unknown_keys": {k: sorted(v) for k, v in unknown_by_action.items()},
        "undeclared_actions": {k: sorted(v) for k, v in undeclared_by_action.items()},
        "declared_action_count": len(declared),
        "undeclared_action_count": len(undeclared_by_action),
    }, sys.stdout, indent=2)
    print()
else:
    print("=== UNKNOWN KEYS (action declared its contract; these are not in it) ===")
    if unknown_by_action:
        for action in sorted(unknown_by_action):
            print(f"  {action}: {', '.join(sorted(unknown_by_action[action]))}")
    else:
        print("  none")

    print()
    print("=== UNDECLARED ACTIONS (not opted in — nothing is known about these keys) ===")
    print(f"  {len(undeclared_by_action)} actions, "
          f"{sum(len(v) for v in undeclared_by_action.values())} (action,key) pairs")
    print(f"  {len(declared)} actions have declared a contract.")
    print()
    print("  This section is NOT a list of bugs. It is the coverage gap: until an")
    print("  action declares ConfigKeys, a dead key on it is still invisible.")
    for action in sorted(undeclared_by_action)[:20]:
        print(f"    {action}: {', '.join(sorted(undeclared_by_action[action]))}")
    if len(undeclared_by_action) > 20:
        print(f"    … and {len(undeclared_by_action) - 20} more actions "
              f"(use --json for the full list — this cap is a display limit, not a filter)")

sys.exit(1 if unknown_by_action else 0)
PY
