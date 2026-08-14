#!/bin/bash
# Read-only. Which live step-config entry can never take effect because the
# action's ActionInputSpec carries a Default for that field?
#
# bugs_open/231, and RE-SPECIFIED 2026-08-13 when the resolver changed. It used
# to be true that against a defaulted field a static config value was silently
# dead — spec.Defaults landed first and every strategy but Strategy 0 skipped a
# field that already held a value. pageflow-builder's `"purpose": "logo"` on
# deploy_image_asset (Defaults purpose="hero") shipped months of hero-shaped logos
# exactly that way. Candidate 2 (owner ruling 2026-08-11 #2) makes an explicit
# config VALUE beat a Default, so that shape is now honoured, and this report's
# two most populous classes became "live_override" instead of dead keys.
#
# Verdicts (assigned by the Go binary, printed here — see below):
#   dead        — the Default always wins: unextractable_field (defaulted field
#                 outside Required+Optional, dead to every config shape),
#                 type_mismatch (a scalar of the wrong kind, refused by Strategy
#                 6's guard), required_empty_string, composite_literal.
#   conditional — wins if and only if its path resolves at runtime:
#                 dotted_conditional and deprecated_bridge. Reported, never fatal
#                 — resolvability is a runtime fact this offline check cannot
#                 decide (231's second face, fixed for asset-deployer by 380).
#   live        — live_override: the resolver honours it. Listed for the census.
#
# THE VERDICT COMES FROM THE BINARY, deliberately. This script used to compute
# `dead = class != "dotted_conditional"` — a second copy of a rule that lives in
# defaultshadow.go, and the re-spec above would have falsified it silently,
# printing 99 working config entries as dead keys. Read f["verdict"]; never
# re-derive it from the class name.
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
# Group on the verdict the BINARY assigned. Anything with an unrecognised verdict
# is surfaced rather than dropped: a new class must be noisy here, not invisible.
known = ("dead", "conditional", "live")
dead = [f for f in findings if f.get("verdict") == "dead"]
cond = [f for f in findings if f.get("verdict") == "conditional"]
live = [f for f in findings if f.get("verdict") == "live"]
unknown = [f for f in findings if f.get("verdict") not in known]

def line(f, tag=None):
    return "  {agent} {path} action={action} {key}={cv!r} (default {dv!r}) [{cls}{tag}]".format(
        agent=f["agent"], path=f["path"], action=f["action"], key=f["key"],
        cv=f["config_value"], dv=f["default_value"], cls=f["class"],
        tag="; " + tag if tag else "")

print("=== DEAD: THE SPEC DEFAULT WINS, WHATEVER THE CONFIG SAYS ===")
if not dead:
    print("  none")
for f in dead:
    print(line(f, "matches default" if f["matches_default"]
               else "MISMATCH — live behaviour is the default, not this value"))
if cond:
    print()
    print("=== CONDITIONAL: beats the Default only if its path resolves; silent fallback otherwise ===")
    for f in cond:
        print(line(f))
if live:
    print()
    print("=== LIVE: the resolver honours these (Strategy 6) — census only, not defects ===")
    for f in live:
        print(line(f, "redundant — equals its default" if f["matches_default"] else "overrides the default"))
if unknown:
    print()
    print("=== UNRECOGNISED VERDICT — this script is older than the binary; update it ===")
    for f in unknown:
        print(line(f, "verdict={!r}".format(f.get("verdict"))))
'
fi

exit $GO_EXIT
