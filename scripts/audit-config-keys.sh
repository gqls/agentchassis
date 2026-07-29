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

# What live definitions actually carry.
#
# CHANGED 2026-07-29 (bugs_open/144). This was a single SQL walk over
# `default_config->'workflow'->'steps'` — the TOP LEVEL ONLY, which is exactly what
# the runtime validator saw. Both halves were blind in the same direction, so
# cross-checking one against the other could never reveal the gap, and 25 (action,
# key) pairs living only inside loop sub-workflows were invisible to the whole
# system. Rewriting the SQL to descend would have fixed the number and left two
# hand-written traversals to drift apart again.
#
# So the walk is no longer written here. The query fetches the definitions and the
# binary walks them with validation.WalkSteps — the SAME traversal the runtime
# validator enforces against. The audit cannot go blind again without the validator
# going blind in the same commit.
#
# The old query also guarded against a STRING config (jsonb_object_keys errors on
# one rather than skipping it). There are none today — 1,173 object configs, 63 null,
# 0 other, measured 2026-07-29 — and the Go path reports an undecodable definition
# instead of silently skipping it, which is the better failure.
WORKFLOWS_JSON="$(psql_q "
SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow'))
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

LIVE="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --live-pairs))"
if [[ -z "$LIVE" ]]; then
    echo "ERROR: no (action, key) pairs extracted from the live definitions" >&2
    exit 2
fi

DECLARED="$DECLARED" LIVE="$LIVE" JSON_OUT="$JSON_OUT" python3 - <<'PY'
import json, os, sys
from collections import defaultdict

_dump = json.loads(os.environ["DECLARED"])
declared = _dump["declared"]
conditional = _dump.get("conditional", {})
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
conditional_by_action = defaultdict(list)

# Where each pair was found. A pair carried ONLY by a step inside a loop
# sub-workflow was invisible to this report until 2026-07-29 (bugs_open/144), so the
# count is printed: if it ever returns to zero while sub-workflows still exist, the
# traversal has broken rather than the fleet having tidied itself.
#
# Pairs are deduplicated BEFORE classification. The same (action, key) can now arrive
# twice — once top-level, once nested — where the old query returned DISTINCT rows,
# and classifying both would list the key twice in the report.
seen_top = set()
seen_nested = set()

for line in os.environ["LIVE"].splitlines():
    line = line.strip()
    if not line or "\t" not in line:
        continue
    parts = line.split("\t")
    if len(parts) < 3:
        continue
    action, key, where = parts[0], parts[1], parts[2]
    (seen_nested if where == "nested" else seen_top).add((action, key))

for action, key in sorted(seen_top | seen_nested):
    if key in framework:
        continue
    if action in declared:
        if key not in declared[action]:
            unknown_by_action[action].append(key)
        elif key in conditional.get(action, {}):
            # Recognised, so NOT unknown — but honoured only under a condition,
            # which means this step may still describe behaviour it never
            # performs. Reported separately because "no unknown keys" was
            # previously printed over exactly this case.
            conditional_by_action[action].append(key)
    else:
        undeclared_by_action[action].append(key)

nested_only = seen_nested - seen_top

if os.environ["JSON_OUT"] == "1":
    json.dump({
        "unknown_keys": {k: sorted(v) for k, v in unknown_by_action.items()},
        "conditional_keys": {k: sorted(v) for k, v in conditional_by_action.items()},
        "undeclared_actions": {k: sorted(v) for k, v in undeclared_by_action.items()},
        "declared_action_count": len(declared),
        "undeclared_action_count": len(undeclared_by_action),
        "pairs_top_level": len(seen_top),
        "pairs_nested": len(seen_nested),
        "pairs_nested_only": len(nested_only),
    }, sys.stdout, indent=2)
    print()
else:
    print("=== COVERAGE OF THIS REPORT ===")
    print(f"  {len(seen_top)} (action, key) pairs at the top level, "
          f"{len(seen_nested)} inside loop sub-workflows, "
          f"{len(nested_only)} of which exist ONLY there.")
    print("  Until 2026-07-29 this report walked the top level only, so those")
    print("  nested-only pairs were invisible to it AND to the runtime validator —")
    print("  both blind in the same direction, agreeing with each other")
    print("  (bugs_open/144). Both now walk the definition with the same Go function.")
    print()
    print("=== UNKNOWN KEYS (action declared its contract; these are not in it) ===")
    if unknown_by_action:
        for action in sorted(unknown_by_action):
            print(f"  {action}: {', '.join(sorted(unknown_by_action[action]))}")
    else:
        print("  none")
        if conditional_by_action:
            print("  ^ read this WITH the next section — 'none' here does NOT mean")
            print("    'no step misdescribes itself'.")

    print()
    print("=== CONDITIONALLY HONOURED (declared, so not unknown — but may not apply) ===")
    if conditional_by_action:
        for action in sorted(conditional_by_action):
            for key in sorted(conditional_by_action[action]):
                print(f"  {action}.{key}: {conditional[action][key]}")
        print()
        print("  These steps carry a key that is recognised but takes effect only under")
        print("  the stated condition. Where it does not hold, the config describes")
        print("  behaviour that does not happen — the same defect as an unknown key,")
        print("  wearing a declaration. (bugs_closed/101; WRONG_CALLS.md 2026-07-28,")
        print("  where declaring these keys is what silenced the section above.)")
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
