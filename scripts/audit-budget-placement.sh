#!/bin/bash
# Read-only. Where does each live step's output-token budget actually come from?
#
# bugs_open/257 round 3, owner decision 4 (2026-09-04): "the limits are set in each
# individual's config, sometimes it has been set in the wrong place and sometimes
# the agent reads the wrong place, please fix it properly."
#
# The code half gave the estate ONE precedence ladder over the four places a budget
# can be written. This is the report that notices the next misplacement. It calls
# production's own ladder (actions.ResolveStepBudget) — it carries no copy of the
# rule, so it cannot answer a different question from the one the pods answer.
#
# THE THREE KINDS, and the fix differs for each:
#
#   UNCONFIGURED   The step declares an ai_service block and NO budget at any
#                  level, so it runs at the provider floor of 2048 — the smallest
#                  number in the estate. bugs_open/205 counted 64 truncations
#                  before anything said so. FIX: declare one.
#   AMBIGUOUS      ONE level declares the budget in BOTH spellings with different
#                  numbers. The canonical ladder takes the ai_service one; the
#                  direct-caller ladder takes the bare one at the step level — so
#                  which of your two numbers is sent depends on which action the
#                  step runs. FIX: delete one.
#   NON_CANONICAL  Declared outside an ai_service block. HONOURED — the ladder
#                  reads it — and reported only so the fleet converges on one
#                  spelling. ADVISORY: it never fails this script.
#
# A root declaration beaten by a step declaration is NOT a finding: that is the
# documented overlay design (root is the fleet default, the step overrides it) and
# feed-triage does it on purpose. Nor is one number written at two levels: for a
# top-level step the runtime StepConfig and the definition block are one declaration
# arriving by two routes. The first cut of this report flagged both and produced 18
# findings, every one healthy.
#
# Usage: scripts/audit-budget-placement.sh [--json]
# Exit:  0 = clean (or advisory-only) · 1 = ambiguous/unconfigured found
#        2 = could not determine
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

# The export carries `agent_config` on top of the projection the other audit-*.sh
# wrappers use, because the two lowest rungs of the ladder (root ai_service.max_tokens
# and the bare root key) live there. The binary REFUSES an export without it rather
# than grading every step against a truncated rule.
#
# It must NOT descend into loop sub-workflows: WalkSteps in the Go binary does that
# itself, and a second hand-written descent is bugs_open/144.
WORKFLOWS_JSON="$(psql_q "
SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow',
                                    'agent_config', default_config - 'workflow'))
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

ERR=$(mktemp -t audit-budget-placement.XXXXXX.stderr)
REPORT_JSON="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --budget-placement) 2>"$ERR")"
GO_EXIT=$?
if [[ $GO_EXIT -gt 1 ]]; then
    echo "ERROR: config-key-audit --budget-placement failed:" >&2
    cat "$ERR" >&2
    rm -f "$ERR"
    exit 2
fi
cat "$ERR" >&2
rm -f "$ERR"

if [[ "$JSON_OUT" == "1" ]]; then
    printf '%s\n' "$REPORT_JSON"
else
    printf '%s' "$REPORT_JSON" | python3 -c '
import json, sys
r = json.load(sys.stdin)
f = r["findings"]
# Group on the KIND the binary assigned. An unrecognised kind is surfaced rather
# than dropped: a new class must be noisy here, not invisible.
known = ("unconfigured", "ambiguous", "non_canonical")
def of(k): return [x for x in f if x.get("kind") == k]

print("scanned: {} model steps across {} live agents, {} max_tokens declarations".format(
      r["steps_scanned"], r["agents_scanned"], r["declarations"]))
if r.get("agents_undecoded"):
    print("  ⚠ {} agent row(s) failed to decode and were NOT scanned".format(r["agents_undecoded"]))
print()

def show(title, rows, note=None):
    print("=== {} ===".format(title))
    if not rows:
        print("  none")
    for x in rows:
        decls = ", ".join("{}={}{}".format(d["level"], d["value"], "*" if d["effective"] else "")
                          for d in (x.get("declarations") or []))
        print("  {} {}".format(x["agent"], x["path"]))
        print("    sends {} from {}   [{}]".format(x["effective"], x["from"] or "(nothing)", decls or "no declaration"))
        print("    {}".format(x["detail"]))
    if note and rows:
        print("  -> {}".format(note))
    print()

show("UNCONFIGURED: running at the 2048 provider floor, because nobody sized the step", of("unconfigured"),
     "declare ai_service.max_tokens on the step")
show("AMBIGUOUS: one level, both spellings, different numbers - the two readers disagree", of("ambiguous"),
     "delete one of them; the * marks the one the canonical ladder takes")
show("NON-CANONICAL (advisory, honoured): declared outside an ai_service block", of("non_canonical"),
     "move it inside the ai_service block it sits beside, so the fleet has one spelling")

unknown = [x for x in f if x.get("kind") not in known]
if unknown:
    print("=== UNRECOGNISED KIND — this script is older than the binary; update it ===")
    for x in unknown:
        print("  {} {} kind={!r}".format(x["agent"], x["path"], x.get("kind")))
'
fi

exit $GO_EXIT
