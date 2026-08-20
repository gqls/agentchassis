#!/usr/bin/env bash
# council-scope.sh — the council review gate's SCOPE, single-sourced (bugs_open/314).
#
# SOURCED, NEVER EXECUTED. It defines what "in scope for council review" means, in
# one place, for the three consumers that each used to carry their own copy:
#
#   fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh  ADMISSION gate
#       missing fragment -> fail LOUD (exit 1, deliberately NOT the refusal's exit 2:
#       a missing definition must neither admit everything nor refuse everything)
#   scripts/council-coverage-nudge.sh                        ADVISORY commit-msg hook
#       missing fragment -> exit 0 SILENTLY. This hook must never block a commit.
#   fixloop_eg_dartsonline/098_REPORT_unreviewed_commits_v1.sh  COVERAGE report
#       missing fragment -> fail LOUD. A report with no scope is not a report.
#
# WHY THIS FILE EXISTS. The scope lived in three hand-maintained copies
# (097:87, nudge:58, 098:76). That is the drift class this estate keeps filing bugs
# about — the 099 roster-mirror trap, the dedup-index/Go-list lockstep. There is now
# one implementation and three callers.
#
# WHAT IS IN SCOPE.
#   1. Platform code — platform/, internal/, pkg/ (owner ruling 2026-07-17).
#   2. An APPLIABLE DB migration — docs/agent_docs/sql_for_agents/NNN_name.sql
#      (bugs_open/314, widened 2026-08-19 on owner direction).
#
# Why (2) is not "docs". The 2026-07-17 ruling is about SUBJECT MATTER: prose does
# not spend council credits. It was implemented as a PATH test, and on this estate a
# migration IS the running system — it rewrites what a live agent does, is live the
# moment it is applied, and has no image tag to roll back. It reaches production
# faster than any Go change, and it was empirically the half the council found the
# sharpest objections in (bugs_open/314 §9; bugs_open/275's round).
#
# STILL OUT OF SCOPE, deliberately: all prose, site content, and the hand-run
# sidecars (NNN_name_ROLLBACK.sql / _VERIFY.sql / _HOLD.sql). A sidecar is not
# applied by scripts/migration/run-migrations.sh and is therefore not the change.
#
# THE MIGRATION VOCABULARY IS A VERBATIM COPY of the runner's own, because sourcing
# this file FROM the runner was considered and rejected: it would couple the review
# gate to the APPLY path, so a syntax error here would stop migrations being
# applied — a bigger blast radius than the bug being fixed. The copy is held honest
# by council_scope_drift_warn() below, which is called on every 097/098 run.

# Platform code (owner ruling 2026-07-17).
COUNCIL_SCOPE_CODE_RE='^(platform|internal|pkg)/'
# VERBATIM from scripts/migration/run-migrations.sh:283 (the appliable-name grep)
# and :65 (SIDECAR_RE). Change these only together with the runner.
COUNCIL_SCOPE_MIGRATION_RE='^docs/agent_docs/sql_for_agents/[0-9]{3}_[A-Za-z0-9_]+\.sql$'
COUNCIL_SCOPE_SIDECAR_RE='_[A-Z][A-Z0-9_]*\.sql$'

# in_council_scope — reads paths on stdin (one per line), prints the in-scope ones.
#
# TWO SEPARATE TESTS, ORed — not one clever regex. The migration arm is
# match-then-reject-sidecar, which is the runner's own idiom (:283-284, a
# `grep -E ... | grep -vE "$SIDECAR_RE"` pipe); a single negative-class regex for a
# trailing _TOKEN is unwritable in ERE. Keeping the arms separate also means a
# platform/ path containing an uppercase-suffixed .sql cannot be knocked out by the
# sidecar rule, which only applies to the migration arm.
#
# Every grep carries `|| true`: a no-match grep exits 1, and 097/098 run under
# `set -euo pipefail`. This function never exits and never returns non-zero.
in_council_scope() {
  local paths; paths=$(cat)
  [ -n "$paths" ] || return 0
  {
    printf '%s\n' "$paths" | grep -E  "$COUNCIL_SCOPE_CODE_RE" || true
    printf '%s\n' "$paths" | grep -E  "$COUNCIL_SCOPE_MIGRATION_RE" \
                           | grep -vE "$COUNCIL_SCOPE_SIDECAR_RE" || true
  } | sort -u
  return 0
}

# council_scope_drift_warn — $1 = repo root. WARNS on stderr, never blocks.
#
# The two patterns above are copies of the runner's. This asserts the originals are
# still byte-identical, by verbatim fixed-string match (grep -qF) on the runner's
# own lines. It fires on every 097/098 run — 11-43 times a day — so a divergence
# cannot sit silent. It deliberately cannot fail a consumer: a stale rule must not
# brick submissions, it must be noisy.
council_scope_drift_warn() {
  local runner="${1:-}/scripts/migration/run-migrations.sh"
  [ -f "$runner" ] || return 0
  if ! grep -qF "SIDECAR_RE='_[A-Z][A-Z0-9_]*\\.sql\$'" "$runner" \
     || ! grep -qF "grep -E '^[0-9]{3}_[A-Za-z0-9_]+\\.sql\$'" "$runner"; then
    echo "WARN: council scope's migration vocabulary no longer matches run-migrations.sh (:65, :283) — reconcile scripts/council-scope.sh (bugs_open/314)." >&2
  fi
  return 0
}
