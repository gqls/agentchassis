#!/bin/bash
# Read-only. Which live step-config entry can never take effect because the
# action's ActionInputSpec carries a Default for that field?
#
# bugs_open/231. ExtractActionInputs applies spec.Defaults first, and every
# later strategy except Strategy 0 skips a field that already has a value — so
# against a defaulted field a static config value is silently dead (the step
# runs with the Default, whatever the config says). pageflow-builder's
# `"purpose": "logo"` on deploy_image_asset (Defaults purpose="hero") shipped
# months of hero-shaped logos exactly this way. Migration 348 repaired the two
# definitions that carried it; this check is the durable half, so the next
# author to write a static for a defaulted field is told at audit time, not by
# a broken artefact.
#
# Finding classes (defined in cmd/config-key-audit/defaultshadow.go):
#   static_string / non_string_literal / composite_literal — dead literals
#   deprecated_bridge   — a *_field alias onto a defaulted field (equally dead)
#   unextractable_field — defaulted field outside Required+Optional: dead to
#                         EVERY config shape, dotted included
#   dotted_conditional  — a dotted path on a defaulted field: live only when it
#                         resolves; falls back to the Default SILENTLY when the
#                         dispatch shape doesn't carry it (231's second face,
#                         fixed for asset-deployer by migration 380). Reported,
#                         never fatal — resolvability is a runtime fact.
#
# CAVEAT: "dead" means dead on the ExtractActionInputs path. An action that
# reads step.Config directly in its own body can still honour the key (that is
# bugs_open/235's shape — honoured, and wrong). Read the flagged action before
# asserting live damage.
#
# Usage: scripts/audit-default-shadowed-keys.sh [--json]
# Exit:  0 = no dead mismatched entries · 1 = findings · 2 = could not determine
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

# Identical export to the other audit-*.sh wrappers, and identical for a reason:
# WalkSteps (in the Go binary) descends into loop sub-workflows itself, so this
# query must not try to. bugs_open/144 is what a second hand-written descent
# costs.
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

ERR=/tmp/audit-default-shadowed-keys.stderr
FINDINGS_JSON="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --default-shadowed-keys) 2>"$ERR")"
GO_EXIT=$?
if [[ $GO_EXIT -gt 1 ]]; then
    echo "ERROR: config-key-audit --default-shadowed-keys failed:" >&2
    cat "$ERR" >&2
    exit 2
fi
cat "$ERR" >&2
rm -f "$ERR"

if [[ "$JSON_OUT" == "1" ]]; then
    printf '%s\n' "$FINDINGS_JSON"
else
    printf '%s' "$FINDINGS_JSON" | python3 -c '
import json, sys
findings = json.load(sys.stdin)
dead = [f for f in findings if f["class"] != "dotted_conditional"]
cond = [f for f in findings if f["class"] == "dotted_conditional"]
print("=== CONFIG ENTRIES SHADOWED BY A SPEC DEFAULT ===")
if not dead:
    print("  none")
for f in dead:
    tag = "matches default" if f["matches_default"] else "MISMATCH — live behaviour is the default, not this value"
    print("  {agent} {path} action={action} {key}={cv!r} (default {dv!r}) [{cls}; {tag}]".format(
        agent=f["agent"], path=f["path"], action=f["action"], key=f["key"],
        cv=f["config_value"], dv=f["default_value"], cls=f["class"], tag=tag))
if cond:
    print()
    print("=== DOTTED PATHS ON DEFAULTED FIELDS (live only if they resolve; silent fallback otherwise) ===")
    for f in cond:
        print("  {agent} {path} action={action} {key}={cv!r} (default {dv!r})".format(
            agent=f["agent"], path=f["path"], action=f["action"], key=f["key"],
            cv=f["config_value"], dv=f["default_value"]))
'
fi

exit $GO_EXIT
