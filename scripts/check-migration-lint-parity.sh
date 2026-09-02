#!/usr/bin/env bash
# check-migration-lint-parity.sh — ADVISORY. Runs bugs_closed/314's migration-lint
# parity guard, but only when the staged diff could actually have broken it.
#
# WHY THIS EXISTS. `scripts/pattern-check.py`'s bugs_open/007 Class C idempotency lint
# decides which changed files are migrations using a COPY of the runner's appliable-name
# pattern plus a DELIBERATELY DIFFERENT exclusion (_HOLD is linted; the runner will not
# apply it). Two hand-kept rules that must stay related is the drift class this estate
# keeps paying for — and this pair did not merely drift, it was WRONG AT BIRTH: the
# runner gained [A-Za-z] on 2026-07-20 (a51333fd7), the lint was written lowercase-only
# on 2026-07-25 (9d95e1c31), and for six weeks it silently skipped every migration with
# a capital in its name. Nobody noticed because nothing ever compared the two.
# `cmd/config-key-audit/migration_lint_predicate_parity_test.go` now does, and this runs
# it at the one moment the author is present.
#
# SCOPED, because every commit should not pay for a Go test: only when the staged diff
# touches one of the two sources of the rule, or the test itself.
#
# ⚠ DELIBERATELY NOT SCOPED TO THE MIGRATIONS DIRECTORY. Firing on every
# docs/agent_docs/sql_for_agents/ commit (~20/day) was considered and rejected: it would
# compile a package importing platform/orchestration/... on a shared tree that often does
# not build because of another session's WIP, so most of those runs would print
# "could not tell" — training every author to ignore the line. Rare and decidable beats
# frequent and undecidable.
#
# ⚠ THE RESIDUAL, STATED. A staged-diff guard fires when drift is INTRODUCED; it cannot
# see drift that is already at HEAD. That is a real gap and this estate has measured it
# — RFC_022's parity test was found FAILING at HEAD for days (see
# check-optional-key-parity.sh's header). `scripts/verify-head-builds.sh --test` and any
# `go test ./cmd/config-key-audit/` close it; nothing here does.
#
# A FILTER IS USED, against scripts/lib/precommit-gotest.sh's stated preference for "".
# The helper is right that a filter is a roster which can silently stop matching — but
# this package carries four unrelated parity suites, and running them all would report
# ANOTHER guard's real failure under this one's headline. A confidently mis-attributed
# finding is worse than a narrow one, and the roster risk is pinned instead by the test
# file's own naming (every test there starts TestMigrationLint).
#
# ADVISORY, never blocking — same rule as commit-scope-report and pattern-check. The
# pre-commit hook runs for every session on every commit; a stray non-zero exit here
# would stop the whole fleet committing.

set -uo pipefail
ROOT="$(git rev-parse --show-toplevel 2>/dev/null)" || exit 0
cd "$ROOT" || exit 0

# Staged-diff / go-test / build-failure-vs-real-failure mechanics are shared with
# check-optional-key-parity.sh and check-finding-code-registry.sh (council reuse
# objection, corr be252395 round 2). What stays HERE is the relevance predicate and the
# guidance — the two things that are actually about bugs_closed/314.
# shellcheck source=scripts/lib/precommit-gotest.sh
. "$ROOT/scripts/lib/precommit-gotest.sh" 2>/dev/null || exit 0

STAGED="$(precommit_staged_files)"
[ -n "$STAGED" ] || exit 0

RELEVANT=0
while IFS= read -r f; do
  case "$f" in
    scripts/migration/run-migrations.sh)                                    RELEVANT=1 ;;
    scripts/pattern-check.py)                                               RELEVANT=1 ;;
    cmd/config-key-audit/migration_lint_predicate_parity_test.go)           RELEVANT=1 ;;
  esac
done <<< "$STAGED"

[ "$RELEVANT" -eq 1 ] || exit 0

precommit_run_gotest ./cmd/config-key-audit/ 'MigrationLint' \
  'migration-lint parity' \
  'migration idempotency lint: predicate DRIFTED from the runner (bugs_closed/314)' \
  '   pattern-check.py'"'"'s MIGRATION_NAME_RE must be run-migrations.sh'"'"'s appliable-name
   pattern VERBATIM, every excluded suffix must be one the runner would refuse to
   apply, and _HOLD.sql must stay LINTED.
   ⚠ DO NOT "reconcile" the exclusion to the runner'"'"'s SIDECAR_RE. That drops _HOLD,
   which is hand-applied, CANNOT be ledger-recorded while it carries the suffix
   (run-migrations.sh:245-250 refuses --record-only on a sidecar), and is then renamed
   into the appliable set — so the runner replays it. Read the Q1/Q2/Q3 block above the
   constants first: three questions are asked of a migration filename here and they are
   deliberately not the same question.'
exit 0
