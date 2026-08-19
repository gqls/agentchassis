#!/bin/bash
# Read-only. Which create_work_item step, filed PER ITEM inside a loop, still
# builds its item_key PER SITE — so every iteration after the first collides on
# idx_swi_dedup and is silently dropped?
#
# bugs_open/321 is the instance: tool-suggester filed one add_tool item per LLM
# suggestion inside a loop, every item on a site shared the key
# 'add_tool[_novel]_<domain>', and 40 suggestions became 11 work items (~72%
# silently lost) before migration 493 added the per-item suffix. The class is
# older than the instance — LANDMINES.md's "a keyed work item whose KEY is
# coarser than its FINDING" entry (2026-08-02) states the rule as a question a
# session must remember to ask; this check is the mechanical version that asks
# it of every live loop, daily. Same shape and same argument as
# audit-single-owner-actions.sh: no run of the defective agent can ever reveal
# the defect (the loop reports success either way), so the check has to be
# offline and fleet-wide.
#
# The once-per-site fleet is deliberately NOT reported: a top-level
# create_work_item's site-wide key is the intended dedupe, and a loop over
# SITES writes a distinct site_id per iteration. Only a loop-nested step whose
# site_id is constant across iterations is a finding. A suffix that is
# DECLARED but unhonoured (empty string, wrong type — silently ignored by the
# action) is convicted too, flagged suffix_declared_but_unhonoured.
#
# Usage: scripts/audit-loop-sitewide-item-keys.sh [--json]
# Exit:  0 = no findings · 1 = findings found · 2 = could not determine

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

# Identical export to audit-single-owner-actions.sh, and identical for a reason:
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

ERR=/tmp/audit-loop-sitewide-item-keys.stderr
FINDINGS_JSON="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --loop-sitewide-item-keys) 2>"$ERR")"
GO_EXIT=$?
# Read refusal from EMPTY stdout, never the exit code alone: `go run` folds the
# tool's exit 2 into its own 1 (the WFA-013 gotcha), so exit codes cannot
# separate "findings" from "refused to run".
if [[ -z "$FINDINGS_JSON" ]]; then
    echo "ERROR: config-key-audit --loop-sitewide-item-keys produced no output:" >&2
    cat "$ERR" >&2
    rm -f "$ERR"
    exit 2
fi
cat "$ERR" >&2
rm -f "$ERR"

if [[ "$JSON_OUT" == "1" ]]; then
    printf '%s\n' "$FINDINGS_JSON"
else
    printf '%s' "$FINDINGS_JSON" | python3 -c '
import json, sys
out = json.load(sys.stdin)
fs = out["findings"]
print("=== LOOP-NESTED create_work_item STEPS STILL KEYED PER SITE ===")
print("  ({} live agents scanned, {} undecoded)".format(out["agents_scanned"], out["agents_undecoded"]))
if not fs:
    print("  none")
else:
    for f in fs:
        extra = " (suffix DECLARED but unhonoured)" if f["suffix_declared_but_unhonoured"] else ""
        print("  {} {}".format(f["agent"], f["path"]))
        print("      prefix {!r}, loop over {!r}{}".format(f["item_key_prefix"], f["loop_variable"], extra))
    print()
    print("  Every iteration of these loops after the first collides on idx_swi_dedup")
    print("  and is silently dropped — the loop reports success either way. Fix: set")
    print("  item_key_suffix_field to a per-item path (usually <loop_variable>.<id>);")
    print("  unresolved is a deliberate hard error, so evidence the path resolves")
    print("  first. Worked example: migration 493. Background: bugs_open/321.")
'
fi

exit $GO_EXIT
