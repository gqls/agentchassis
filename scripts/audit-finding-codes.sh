#!/bin/bash
# Read-only. Which agent_error_log finding code is firing with no declared
# disposition?
#
# bugs_open/358. Detectors, guards and audits across this estate record what
# they notice as a row in agent_error_log under a named error_code. Most of
# those codes have no automated reader, and migration 466's `database-cleanup`
# pre_query (live, hourly) deletes an unresolved row at 30 days — so the
# system's memory of its own detections is a sliding window that, for most
# codes, nothing looks at before it slides. The fix is not "build a reader for
# each": some are legitimately human evidence, some are deliberate time-boxed
# instrumentation, and operational plumbing is correctly consumed by the generic
# diagnostic readers. The fix is that a code cannot enter the estate with NO
# DECLARED DISPOSITION and nobody notice. This is what notices.
#
# WHY IT READS THE TABLE AND NOT THE SOURCE. agenterrors.go:3 calls itself "The
# ONE writer against agent_error_log" and it is no longer true — four of five
# INSERT paths bypass it, and internal/agents/contentcreator CANNOT use it (it
# holds a *pgxpool.Pool against Write's *sql.DB). So a source scan or a guard at
# the seam would read clean while most writers walked past. `SELECT DISTINCT
# error_code` is blind to none of them: it sees every writer regardless of
# language, seam, or whether the code is a literal, a constant, a positional
# argument or a value from config.
#
# Usage: scripts/audit-finding-codes.sh [--json]
# Exit:  0 = every observed code is declared (the `unruled` BACKLOG is reported,
#            not a finding — a check that fails from day one over a pre-existing
#            backlog is a check that gets ignored)
#        1 = an observed code is undeclared, or a declaration does not hold up
#        2 = could not determine
#
# NOTE on exit codes (LANDMINES.md — `go run` collapses the child's exit
# status): the refusal is discriminated by EMPTY OUTPUT where the report
# belongs, never by branching on exit code 2. That branch would be dead code
# under `go run`, exactly as audit-optional-key-budget.sh records.

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REGISTRY="$REPO_ROOT/docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json"

if [[ ! -f "$REGISTRY" ]]; then
    echo "finding-code registry missing at $REGISTRY — refusing to run: with no" >&2
    echo "declarations every observed code reads as undeclared, and the report" >&2
    echo "would be noise rather than a finding." >&2
    exit 2
fi

# The authority. DISTINCT over the whole retained window: this asks which codes
# EXIST, never how many rows each has, so no occurred_at bound belongs here —
# the retention job already provides the only window there is.
CODES=$(kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -At -c \
    "SELECT DISTINCT error_code FROM agent_error_log
      WHERE error_code IS NOT NULL AND error_code <> ''" 2>/dev/null) || CODES=""

# VACUITY GUARD. An empty read produces a clean report indistinguishable from a
# healthy estate — and this check's own subject is checks that cannot fail, so
# passing here would be the bug reproducing itself. agent_error_log has carried
# tens of thousands of rows continuously since 2026-07-23; zero distinct codes
# means the query did not run.
if [[ -z "$CODES" ]]; then
    echo "could not read agent_error_log (kubectl/psql failed, or the table came" >&2
    echo "back empty) — NOT reporting a clean run. An empty result here is an" >&2
    echo "instrument failure, not a healthy estate." >&2
    exit 2
fi

printf '%s\n' "$CODES" \
  | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --finding-codes \
        --registry "$REGISTRY" --root "$REPO_ROOT" "$@")
