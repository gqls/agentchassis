#!/usr/bin/env bash
# bundle_diagnosis_loop.sh
#
# A read-only context bundle for the task "build the diagnosis loop" — the
# task-scoped equivalent of your gamesdesign example, but the SUBJECT is the
# loop's own code rather than a chassis bug. It assembles, into one markdown
# file, the loop's decisive symbols + the four diagnose actions + the governing
# docs + the constitution + (optionally) the diagnose-relevant schema/runtime,
# so the next chat (or a sub-agent) has the focused context to continue the
# gated work without re-reading the whole tree.
#
# This drives cmd/bundle directly (a single bundle). It is NOT cmd/diagnose (the
# loop runner). The flag names are the ones BundleGatherer.buildArgs actually
# passes to cmd/bundle (gatherer.go), plus -include / -df-filter from your
# example. cmd/bundle is read-only by construction (\d / capped SELECT / log
# read + the pure assembler): nothing here triggers a build, spawn, or write.
#
# ─────────────────────────────────────────────────────────────────────────────
# CONFIRM BEFORE RUNNING (flagged — I could not verify these from the mounted
# files; only the contextkit engine .go files were available):
#
#  1. ROOT + ANALYSIS default to the CHASSIS (matching your example + the RUNBOOK
#     convention). pkg/diagnose and the four diagnose actions are DRAFTS
#     (chassis-drafts/). If they are not yet committed to ~/projects/agentchassis
#     AND re-analysed into chassis_clean.json, cmd/bundle will SKIP those -scope
#     entries (it can only slice symbols the analysis knows). Either commit +
#     re-analyse first, or use the contextkit ALT block at the bottom, which
#     resolves every engine symbol today.
#
#  2. The four ACTION FILENAMES are derived from the action NAMES in the handoff
#     (diagnose_load_runtime / _assemble_bundle / _route / _emit) + the chassis
#     "<snake>_action.go" convention. Confirm them against the actual
#     chassis-drafts filenames (the migration's registry_entries name them).
#
#  3. -doc / -constitution paths resolve from the contextkit working dir ($CK),
#     where cmd/bundle runs. Use absolute paths or paths relative to $CK.
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

# ── where cmd/bundle lives (everything runs from here) ──────────────────────
: "${CK:=$HOME/projects/agentchassis/docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit}"

# ── the repo being bundled + its analysis (call graph + symbol line ranges) ──
: "${ROOT:=$HOME/projects/agentchassis}"
: "${ANALYSIS:=/tmp/chassis_clean.json}"

# ── standing authored context ───────────────────────────────────────────────
: "${CONSTITUTION:=$CK/thin_slice_constitution.md}"
: "${VERDICT_PROMPT:=$ROOT/docs/PROMPT_diagnosis_verdict.md}"   # the cite-or-abstain contract
: "${RUNBOOK:=$ROOT/docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/RUNBOOK_design_diagnosis_loop.md}"

# ── read-only DB gather — set PSQL="" to skip the DB section entirely ────────
: "${PSQL:=kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db}"

# ── output ──────────────────────────────────────────────────────────────────
: "${OUT:=/tmp/bundle_diagnosis_loop.md}"

cd "$CK"

# Optional: (re)generate the chassis analysis so the loop symbols resolve.
# Slow on a big tree — uncomment once the drafts are committed.
#   go run ./cmd/analyser "$ROOT" -exclude docs/ -exclude _archive/ > "$ANALYSIS"

# The one-sentence framing the bundle is built around (cmd/bundle -task).
TASK="Build the read-only chassis diagnosis loop: hypothesise -> gather scoped read-only evidence -> cite-or-abstain verdict -> re-scope by FOLLOWING evidence (the call graph for code; vetted queries for data), NOT by re-searching the symptom. Human-gated, never a fix, never triggers a run. Falsification (abandoning a wrong hypothesis) is the move it must do well."

ARGS=(
  -analysis     "$ANALYSIS"
  -root         "$ROOT"
  -constitution "$CONSTITUTION"
  -step         debug
  -task         "$TASK"

  # ── the loop's decisive symbols (pkg/diagnose), mapped to its motion ──
  -scope pkg/diagnose/advance.go:Advance          # chassis-facing per-iteration step the action calls
  -scope pkg/diagnose/step.go:DecideStep          # the one pure decision core (route + next scope)
  -scope pkg/diagnose/loop.go:nextScope           # re-scope by FOLLOWING the call graph
  -scope pkg/diagnose/loop.go:guardAfter          # guards: falsification / no-progress / seen-before / cap
  -scope pkg/diagnose/callgraph.go:Neighbourhood  # the follow primitive (drops ubiquitous names)
  -scope pkg/diagnose/verdict_wire.go:toVerdict   # wire->Verdict; UNVERIFIABLE fail-safe; read-only data-request drop
  -scope pkg/diagnose/sqlguard.go:IsReadOnlySQL   # Guard 2: the read-only lint

  # ── the four diagnose actions (the chassis agent surface) ──
  -scope platform/orchestration/actions/diagnose_load_runtime_action.go
  -scope platform/orchestration/actions/diagnose_assemble_bundle_action.go
  -scope platform/orchestration/actions/diagnose_route_action.go
  -scope platform/orchestration/actions/diagnose_emit_action.go

  # ── supporting context (full text, not a focus symbol) ──
  -include platform/orchestration/actions/registry.go

  # ── governing docs (pasted verbatim into "Reference documents") ──
  -doc "$VERDICT_PROMPT"
  -doc "$RUNBOOK"
  #  -doc "$ROOT/docs/agent_docs/.../016_debugging_guide_v2_56.md"   # the §9 catalogue (large — pastes whole)
  #  -doc "$ROOT/docs/agent_docs/.../001_development_guide.md"        # agent-creation guide (the loop IS an agent)
  #  -doc "$ROOT/docs/agent_docs/.../003_contracts_and_standards.md"  # the output_field / result-spec contract

  # ── read-only DB gather: \d of the tables the loop reads / may query ──
  -psql "$PSQL"
  -schema-tables agent_error_log,site_work_items,page_components,pages
  -capabilities
  #  -runtime-site gamesdesign.co.uk -runtime-page index   # attach the eval-fixture's runtime rows (the data side)
  #  -df-filter snapshot                                    # carried from your example; keep/drop as needed

  -out "$OUT"
)

# Drop the whole DB block if PSQL was cleared.
if [[ -z "${PSQL}" ]]; then
  FILTERED=(); skip=0
  for a in "${ARGS[@]}"; do
    case "$a" in
      -psql)          skip=1; continue ;;
      -schema-tables) skip=1; continue ;;
      -capabilities)  continue ;;
    esac
    if [[ $skip -eq 1 ]]; then skip=0; continue; fi
    FILTERED+=("$a")
  done
  ARGS=("${FILTERED[@]}")
fi

go run ./cmd/bundle "${ARGS[@]}"
echo "bundle written: $OUT"

# ─────────────────────────────────────────────────────────────────────────────
# ALT — bundle the STANDALONE engine in contextkit (resolves every symbol TODAY,
# before the chassis integration is committed). Set:
#
#   ROOT="$CK"; ANALYSIS=/tmp/contextkit.json
#   # go run ./cmd/analyser "$CK" -exclude docs/ > "$ANALYSIS"
#
# then re-point the engine scopes at internal/diagnose/* and DROP the four action
# -scope lines + -include (those exist only in the chassis):
#   -scope internal/diagnose/advance.go:Advance
#   -scope internal/diagnose/step.go:DecideStep
#   -scope internal/diagnose/loop.go:nextScope
#   -scope internal/diagnose/loop.go:guardAfter
#   -scope internal/diagnose/callgraph.go:Neighbourhood
#   -scope internal/diagnose/verdict_wire.go:toVerdict
#   -scope internal/diagnose/sqlguard.go:IsReadOnlySQL
#   -scope internal/diagnose/gatherer.go:BundleGatherer   # the harness->cmd/bundle adapter
# ─────────────────────────────────────────────────────────────────────────────
