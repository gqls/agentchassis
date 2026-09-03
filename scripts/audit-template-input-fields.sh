#!/bin/bash
# Read-only. Which live prompt template names a variable its step's
# `input_fields` can never supply?
#
# A step renders its prompt_template against ExtractFields(CollectedData,
# input_fields) — a SUBSET of the run's data. A template variable whose root is
# outside that subset is absent when the template executes, and Go's text/template
# has no opinion: a {{range}} or {{if}} over an absent key renders NOTHING, an
# unguarded {{.x}} renders "<no value>". The step succeeds either way. The output
# is a plausible prompt with a section silently missing.
#
# bugs_open/453 records four catches of this class, every one found by a fixture
# somebody happened to write. This is the mechanical version, asked of every live
# template. It reports THREE kinds and fails on ONE:
#
#   unreachable_root   input_data is NOT among input_fields, so config alone
#                      decides it and the variable resolves on no row, ever.
#                      EXIT 1.
#   conditional_root   input_data IS among input_fields, so ExtractFields also
#                      promotes every key of the runtime input_data map to the
#                      root — undecidable from config. Advisory.
#   declared_unread    an input_fields entry no template variable reads. Costs a
#                      whole-tree extraction per run. Advisory.
#
# What it does NOT close, stated so nobody reads a clean run as more than it is:
# a root that IS present whose SUB-FIELD is missing from the row's data still
# renders "<no value>", and no check over config can see that — the config is
# correct. That arm belongs to RenderPromptTemplate's own "<no value>" scan.
#
# Usage: scripts/audit-template-input-fields.sh [--json]
# Exit:  0 = no unreachable roots · 1 = unreachable roots found · 2 = could not determine

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

# The export carries ONE key the other audit scripts' exports do not:
# agent_prompt_template. That is tier 2 of getPromptWithPriority — the template a
# step with no prompt_template of its own actually renders. Keep it in step with
# fleetExportQueryWithPrompts in cmd/config-key-audit/templateinputfields.go
# (the CronJob/DB route uses that one).
#
# If you drop the projection the binary REFUSES rather than running blind: a
# jsonb_build_object emits the key as null when the agent has none, so "no row
# carries it" can only mean the wrong query.
WORKFLOWS_JSON="$(psql_q "
SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow',
                                    'agent_prompt_template', default_config->'prompt_template'))
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

ERR=/tmp/audit-template-input-fields.stderr
FINDINGS_JSON="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --template-input-fields) 2>"$ERR")"
GO_EXIT=$?
# Read refusal from EMPTY stdout, never the exit code alone: `go run` folds the
# tool's exit 2 into its own 1 (the WFA-013 gotcha), so exit codes cannot
# separate "findings" from "refused to run".
if [[ -z "$FINDINGS_JSON" ]]; then
    echo "ERROR: config-key-audit --template-input-fields produced no output:" >&2
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
by = {"unreachable_root": [], "conditional_root": [], "declared_unread": []}
for f in out["findings"]:
    by[f["kind"]].append(f)

print("=== TEMPLATE VARIABLES THEIR STEP CANNOT SUPPLY ===")
print("  ({} live agents, {} steps walked, {} prompt templates checked, {} of them agent-tier)".format(
    out["agents_scanned"], out["steps_walked"], out["templates_checked"], out["templates_agent_tier"]))
if out["parse_failures"]:
    print()
    print("  !! {} TEMPLATE(S) FAILED TO PARSE and were NOT checked:".format(len(out["parse_failures"])))
    for p in out["parse_failures"]:
        print("       {} {}: {}".format(p["agent"], p["path"], p["error"]))

print()
print("-- UNREACHABLE (input_data not promoted: resolves on no row, ever) --")
if not by["unreachable_root"]:
    print("  none")
for f in by["unreachable_root"]:
    print("  {} {} [{} tier]".format(f["agent"], f["path"], f["template_tier"]))
    print("      template wants {} - input_fields: {}".format(
        ", ".join("{{."+r+"}}" for r in f["roots"]), f["input_fields"] or "(none declared)"))

print()
print("-- CONDITIONAL (input_data promoted: depends on a row, undecidable here) --")
if not by["conditional_root"]:
    print("  none")
for f in by["conditional_root"]:
    print("  {} {} wants {}".format(f["agent"], f["path"], ", ".join("{{."+r+"}}" for r in f["roots"])))

print()
print("-- DECLARED BUT UNREAD (extracted every run, read by no template variable) --")
if not by["declared_unread"]:
    print("  none")
for f in by["declared_unread"]:
    print("  {} {}: {}".format(f["agent"], f["path"], ", ".join(f["roots"])))

if by["unreachable_root"]:
    print()
    print("  Each unreachable root renders as NOTHING (guarded/ranged) or as the literal")
    print("  <no value> (unguarded), with no error and no verdict. Fix: add the root to")
    print("  that step input_fields - or delete the template block if it is dead.")
    print("  Background: bugs_open/453.")
'
fi

exit $GO_EXIT
