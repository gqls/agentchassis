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
# STILL OUT OF SCOPE, deliberately: all prose and site content, plus the SQL files
# that are not the change — `_ROLLBACK` (the undo), `_VERIFY` (the check) and
# `_SUPERSEDED` (retired). Everything else matching NNN_name.sql is in scope.
#
# ⚠ `_HOLD.sql` IS IN SCOPE, and getting this wrong was a real defect in the first
# cut of this file (caught by the council's editquality seat, corr 85fac99c,
# 2026-08-20). The first version excluded every uppercase-suffixed sidecar by reusing
# the runner's own SIDECAR_RE — which answers "will `--apply` run this?" — as a proxy
# for "is this the change?". Those are DIFFERENT QUESTIONS, and conflating a proxy
# with the intent is precisely the defect bugs_open/314 exists to fix: it is the same
# mistake one level down. A `_HOLD.sql` is a migration deliberately held back from the
# runner FOR ORDERING (LANDMINES: "write the two commands to apply it by hand into the
# file's own header, because the person who runs them will not be you"). It is live
# agent config, hand-applied, and merely not ledger-recorded — so excluding it exempted
# exactly the risk this fix exists to catch. Measured in the directory when this was
# corrected: 157 `_ROLLBACK`, 12 `_HOLD`, 7 `_VERIFY`, 2 `_SUPERSEDED`; the four HOLD
# files sampled all carried real UPDATE/jsonb_set writes.
#
# THE MIGRATION-NAME PATTERN IS A VERBATIM COPY of the runner's own, because sourcing
# this file FROM the runner was considered and rejected: it would couple the review
# gate to the APPLY path, so a syntax error here would stop migrations being
# applied — a bigger blast radius than the bug being fixed. The copy is held honest
# by council_scope_drift_warn() below, which is called on every 097/098 run.

# Platform code (owner ruling 2026-07-17), plus the check-detector binary
# (owner ruling 2026-08-23).
#
# WHY cmd/config-key-audit AND NOT ALL OF cmd/. A "check service" — one of the ~20
# nightly CronJobs that watch for a class of defect — keeps its DETECTOR LOGIC in
# cmd/, its image recipe in build/docker/ and its schedule in deployments/. None of
# those paths was in scope, so the entire class was invisible to this gate: measured
# 2026-08-23, both commits shipping component-source-vocabulary-check contained ZERO
# in-scope files, and a second lane hit the identical wall the same day
# (bugs_open/309, bugs_open/363).
#
# The owner was offered all of cmd/ and chose this one directory, on the measurement
# rather than the principle. As of 2026-08-23 cmd/ is 18,402 non-test lines — about
# 6% of the reviewable-shaped Go here — and it is CONCENTRATED, not spread:
# cmd/config-key-audit alone is 6,417 of them (35%), and it carries the rules for
# roughly ten of the twenty nightly checks. The rest of cmd/ is overwhelmingly thin
# main() wiring that would spend credits to review a flag parser. So the honest
# statement of the gap was never "a sixth of the codebase is unreviewed" — it was
# "the detector logic for the daily check fleet has accumulated in one binary the
# gate cannot see", and this widening closes exactly that and nothing else.
#
# Same shape as the 2026-08-19 migration widening (bugs_open/314): a TARGETED
# addition justified by where the risk actually sits, not a general loosening.
#
# ⚠ IF THE DETECTOR LOGIC MOVES, MOVE THIS. The moment a check's rules live somewhere
# else under cmd/ — a new cmd/<something>-check with real logic rather than a thin
# main — this pattern stops covering the class it was widened for, and it will do so
# SILENTLY, because an out-of-scope path is refused with no finding. cmd/ line counts
# per directory are two commands: see bugs_open/309's notes for the one-liner.
# ⚠ WIDENING THIS REGEX IS NOT ENOUGH — 098 CARRIES A SECOND, PATHSPEC COPY.
# scripts/... this file is the single source for the *decision*, and 097 and the commit-msg
# nudge do read only this. But 098_REPORT_unreviewed_commits_v1.sh must enumerate commits
# BEFORE it can judge them, and `git log` takes pathspecs, not regexes — so it carries
# SCOPE_PATHS, a hand-kept array, as a PRE-filter. A path added here and not there is
# INVISIBLE to the coverage report: not UNREVIEWED, absent, which reads as nothing to
# report. MEASURED 2026-08-23, the day cmd/config-key-audit was added below: 22 commits in
# the previous fortnight were in scope by this regex and in no bucket of that report, across
# four lanes. Change both, in one commit.
#
# ── OWNER RULING 2026-09-02: cmd/brief-negation-check enters scope. ──────────────
# Raised by the 414 lane on 08-31, ruled GO by the owner 2026-09-02 (copy lane's
# ledger). Same argument that admitted cmd/config-key-audit: the binary FILES WORK
# ITEMS against live sites (brief_supplies_negation + spec claims), so its rules act
# on production, and its word arm (v1.0.1352) is now the estate's only reader of the
# banned-words register in briefs. Widened HERE and in 098's SCOPE_PATHS in the SAME
# commit, per the CLAUDE.md warning that a path in one and not the other is invisible
# to the coverage report.
# ── OWNER RULING 2026-08-24: scripts/pattern-check.py enters scope. ──────────────
# The 2026-08-23 widening admitted cmd/config-key-audit on the reasoning that the
# DETECTOR LOGIC FOR THE CHECK FLEET had accumulated where the gate could not see it.
# That argument applies with more force to pattern-check.py, and it was missed at the
# time: `[MEASURED 2026-08-24]` it is **2,058 lines carrying 22 checks**, versus **2,220
# lines for EVERY other audit-*/check-* script under scripts/ combined**. So one file is
# about half the estate's non-Go detector surface, and it is the half that runs most
# often: the nightly cron fleet fires ~20 checks once a day, while this file runs on
# EVERY COMMIT IN EVERY SESSION via .githooks/pre-commit. Its own header states the
# stake — "a false positive that blocks is a fleet-wide outage".
#
# Same shape as both prior widenings (bugs_open/314's migrations, 08-23's cmd/): a
# TARGETED addition justified by where the risk sits, not a general loosening of scripts/.
#
# ⚠ DELIBERATELY NOT INCLUDED, and worth revisiting rather than assuming:
# scripts/audit-advisory-findings.py (429 lines) IMPORTS this file's CHECKS tuple, so the
# two are coupled — but it REPORTS on findings rather than deciding them, and a reporting
# bug cannot block a commit. The other ~10 audit-*/check-* scripts are 90-343 lines each
# and individually thin. If detector logic starts accumulating in any of them, this
# pattern stops covering the class it was widened for — silently, because an out-of-scope
# path is refused with no finding. Re-measure with:
#   wc -l scripts/audit-*.sh scripts/audit-*.py scripts/check-*.sh scripts/check-*.py
#
# ⚠ ANCHORED TO THE ONE FILE ($), not to scripts/. Adding a second detector file means
# editing this regex AND 098's SCOPE_PATHS — see the warning directly above, which is
# what caught this change before it shipped half-done.
COUNCIL_SCOPE_CODE_RE='^(platform|internal|pkg)/|^cmd/config-key-audit/|^cmd/brief-negation-check/|^scripts/pattern-check\.py$'
# VERBATIM from scripts/migration/run-migrations.sh:283 (the appliable-name grep).
# Change this only together with the runner.
COUNCIL_SCOPE_MIGRATION_RE='^docs/agent_docs/sql_for_agents/[0-9]{3}_[A-Za-z0-9_]+\.sql$'
# The NOT-THE-CHANGE suffixes. Deliberately an ENUMERATION, not the runner's
# catch-all SIDECAR_RE: a suffix must be shown to be "not the change" to be excluded,
# so anything new (including a suffix nobody here anticipated) defaults to IN scope —
# a wasted credit, never an unreviewed config change. The trailing [A-Z0-9_]* catches
# stacked suffixes such as 470_..._ROLLBACK_SUPERSEDED.sql.
COUNCIL_SCOPE_NOT_THE_CHANGE_RE='_(ROLLBACK|VERIFY|SUPERSEDED)[A-Z0-9_]*\.sql$'

# in_council_scope — reads paths on stdin (one per line), prints the in-scope ones.
#
# TWO SEPARATE TESTS, ORed — not one clever regex. The migration arm is
# match-then-reject-not-the-change, following the runner's own idiom (:283-284, a
# `grep -E ... | grep -vE ...` pipe); a single negative-class regex for a trailing
# _TOKEN is unwritable in ERE. Keeping the arms separate also means a platform/ path
# containing an uppercase-suffixed .sql cannot be knocked out by the exclusion, which
# only applies to the migration arm.
#
# Every grep carries `|| true`: a no-match grep exits 1, and 097/098 run under
# `set -euo pipefail`. This function never exits and never returns non-zero.
in_council_scope() {
  local paths; paths=$(cat)
  [ -n "$paths" ] || return 0
  {
    printf '%s\n' "$paths" | grep -E  "$COUNCIL_SCOPE_CODE_RE" || true
    printf '%s\n' "$paths" | grep -E  "$COUNCIL_SCOPE_MIGRATION_RE" \
                           | grep -vE "$COUNCIL_SCOPE_NOT_THE_CHANGE_RE" || true
  } | sort -u
  return 0
}

# council_scope_drift_warn — $1 = repo root. WARNS on stderr, never blocks.
#
# TWO checks, and the second is the one that earns its keep.
#
# (1) The migration-NAME pattern above is a copy of the runner's :283. Assert the
#     original is still byte-identical, by verbatim fixed-string match (grep -qF).
#
# (2) Census the ACTUAL uppercase suffixes in the migrations directory and warn about
#     any this file has not classified. This is what a drift check should watch here.
#     The first cut of this file checked (1) only — it asserted "my copy matches the
#     runner's SIDECAR_RE" and passed happily while the RULE was wrong, because the
#     runner's question ("will --apply run this?") is not this file's question ("is
#     this the change?"). A check that can only confirm a copy cannot notice a
#     misclassification. This one can: a new `_STAGED`/`_PENDING`/whatever suffix
#     appears, and the next 097/098 run says so. Until someone classifies it, such a
#     file is IN scope by default — a wasted credit, never an unreviewed change.
#
# Fires on every 097/098 run — 11-43 times a day — so a divergence cannot sit silent,
# and deliberately cannot fail a consumer: a stale rule must be noisy, not blocking.
council_scope_drift_warn() {
  local root="${1:-}"
  local runner="$root/scripts/migration/run-migrations.sh"
  if [ -f "$runner" ] \
     && ! grep -qF "grep -E '^[0-9]{3}_[A-Za-z0-9_]+\\.sql\$'" "$runner"; then
    echo "WARN: council scope's migration-name pattern no longer matches run-migrations.sh:283 — reconcile scripts/council-scope.sh (bugs_open/314)." >&2
  fi

  local dir="$root/docs/agent_docs/sql_for_agents"
  [ -d "$dir" ] || return 0
  # ⚠ THE `|| true` IS LOAD-BEARING, and omitting it broke the gate outright during
  # this very fix: with every suffix classified the final `grep -v` matches nothing,
  # exits 1, the command substitution inherits that status, and a caller running
  # `set -euo pipefail` (097 and 098 both do) DIES — with no output, before the scope
  # check, so every submission fails. It fires only when the census is CLEAN, i.e. the
  # healthy case. Caught by tracing an unexplained `exit 1`, not by the control matrix.
  local unclassified
  unclassified=$( { ls -1 "$dir" 2>/dev/null \
    | grep -E '^[0-9]{3}_[A-Za-z0-9_]+\.sql$' \
    | grep -E '_[A-Z][A-Z0-9_]*\.sql$' \
    | grep -vE "$COUNCIL_SCOPE_NOT_THE_CHANGE_RE" \
    | grep -vE '_HOLD[A-Z0-9_]*\.sql$' \
    | sed -E 's/.*(_[A-Z][A-Z0-9_]*)\.sql$/\1/' | sort -u | tr '\n' ' '; } || true )
  if [ -n "${unclassified// /}" ]; then
    echo "WARN: unclassified migration suffix(es) in sql_for_agents: ${unclassified}— scripts/council-scope.sh treats these as IN scope (the safe default). If one is not the change, add it to COUNCIL_SCOPE_NOT_THE_CHANGE_RE (bugs_open/314)." >&2
  fi
  return 0
}
