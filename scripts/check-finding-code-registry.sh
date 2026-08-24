#!/usr/bin/env bash
# check-finding-code-registry.sh — ADVISORY. Runs bugs_open/358's finding-code
# registry tests, but only when the staged diff could actually have broken them.
#
# WHY THIS EXISTS, and it is not the usual "a guard nobody runs" story — it is
# one rung worse. `cmd/config-key-audit --finding-codes` has two arms that grade
# the registry against Go SOURCE: a `consumed` entry's `reader` must be a real
# file that names the code, and its `reader_sink` must be a table that file
# mentions. Those two arms are what stop `consumed` being satisfiable by typing.
#
# From phase 2 (the finding-code-registry-check CronJob) they no longer run in
# the cluster: that image ships no repo, so it runs `--no-source` and says so in
# every doc_notes row. That is the right split — both halves of those arms change
# only by commit, so a daily re-run in the cluster could not come out differently
# than it did at build time — but it is only honest if the arms actually run AT
# commit time. Before this script they ran nowhere automatically:
# `check-optional-key-parity.sh` compiles this same package but with
# `-run 'BudgetCron'`, which excludes TestShippedRegistryIsSelfConsistent, the
# test that grades the file the estate actually ships.
#
# SCOPED, because every commit should not pay for a Go test. It fires when the
# staged diff touches the registry, the mode's own source, or — computed from the
# registry at run time, never a hand-kept list — any file a `consumed` entry
# currently names as its reader. That last clause is the one that matters: a
# session deleting the query out of `page_build_failure_guard.go` has not
# touched the registry and would otherwise hear nothing.
#
# NO `-run` FILTER, deliberately. A filter is a roster, and a roster drifts: a
# new test whose name did not match it would silently never run, which is the
# exact class this lane keeps retiring. The package is the unit.
#
# ADVISORY, never blocking — same rule as commit-scope-report, pattern-check and
# check-optional-key-parity. The pre-commit hook runs for every session on every
# commit; a stray non-zero exit here would stop the whole fleet committing.
#
# ⚠ A BUILD FAILURE IS NOT A REGISTRY FAULT. This tree is shared and often does
# not compile because of another session's work in progress (measured 2026-08-23:
# an uncommitted `platform/livespec` rename left this very package unbuildable).
# Reporting "the registry is broken" when the truth is "I could not tell" would be
# exactly the confident-wrong signal this check exists to prevent, so the two are
# separated and the undecidable case says which it is.

set -uo pipefail
ROOT="$(git rev-parse --show-toplevel 2>/dev/null)" || exit 0
cd "$ROOT" || exit 0

# The staged-diff / go-test / build-failure-vs-real-failure mechanics are SHARED
# with scripts/check-optional-key-parity.sh rather than re-implemented here — the
# council's reuse seat objected to the first cut on exactly that ground (corr
# be252395, round 2, medium: "two independently-built mechanisms solving
# overlapping problems in the same package, each maintained separately"), and this
# lane has no standing to argue: retiring pairs of hand-maintained things that
# must stay in sync is most of what bugs_open/358 is.
# shellcheck source=scripts/lib/precommit-gotest.sh
. "$ROOT/scripts/lib/precommit-gotest.sh" 2>/dev/null || exit 0

REGISTRY="docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json"

STAGED="$(precommit_staged_files)"
[ -n "$STAGED" ] || exit 0

# The reader files, read out of the registry ITSELF so this can never become a
# fourth hand-maintained roster. Read from the WORKING TREE copy: a staged
# registry edit that adds a reader should put that file in scope immediately.
READERS=""
if [ -f "$REGISTRY" ] && command -v python3 >/dev/null 2>&1; then
  READERS="$(python3 - "$REGISTRY" <<'PYREADERS' 2>/dev/null || true
import json, sys
try:
    d = json.load(open(sys.argv[1]))
except Exception:
    sys.exit(0)
for k, v in d.items():
    if k.startswith("_") or not isinstance(v, dict):
        continue
    if v.get("disposition") == "consumed":
        r = (v.get("reader") or "").strip()
        if r:
            print(r.rsplit(":", 1)[0] if r.rsplit(":", 1)[-1].isdigit() else r)
PYREADERS
)"
fi

# TWO packages, TWO relevance sets. The checker's tests (cmd/config-key-audit)
# grade the registry against reader files; the actions package's scan test
# grades the package's OWN ErrorCode: writes against the registry. The first cut
# of this script ran only the first — so a session adding `ErrorCode: "NEW"` to
# an actions file walked through the hook unremarked, which is the exact commit
# shape that produced LINK_CONTEXT_UNAVAILABLE, and the scan test's own header
# claimed commit-time coverage it was not getting (review finding, Fable
# 2026-08-24). A claimed early warning that nothing runs is this lane's own
# founding defect, one level up.
RELEVANT_CHECKER=0
RELEVANT_SCAN=0
while IFS= read -r f; do
  [ -n "$f" ] || continue
  case "$f" in
    "$REGISTRY")                          RELEVANT_CHECKER=1; RELEVANT_SCAN=1 ;;
    cmd/config-key-audit/findingcodes*)   RELEVANT_CHECKER=1 ;;
    platform/orchestration/actions/finding_code_roster_test.go) RELEVANT_SCAN=1 ;;
    platform/orchestration/actions/findingcodes_scan_test.go)   RELEVANT_SCAN=1 ;;
    platform/orchestration/actions/*.go)
      # Only a staged actions file that carries an ErrorCode: field can add a
      # code the scan must see. Read the STAGED content, as the parity script
      # does for ActionInputSpec, so a working-tree edit that is not being
      # committed cannot trigger or suppress this.
      if git show ":$f" 2>/dev/null | grep -q 'ErrorCode:'; then
        RELEVANT_SCAN=1
      fi
      ;;
  esac
  # A reader file named by a live `consumed` entry. This clause is the one
  # that catches a session deleting the query out from under a declaration
  # without ever touching the registry.
  while IFS= read -r r; do
    [ -n "$r" ] && [ "$f" = "$r" ] && RELEVANT_CHECKER=1
  done <<< "$READERS"
done <<< "$STAGED"

[ "$RELEVANT_CHECKER" -eq 1 ] || [ "$RELEVANT_SCAN" -eq 1 ] || exit 0

# NO -run FILTER, deliberately. A filter is a roster, and a roster drifts: a new
# test whose name did not match it would silently never run, which is the exact
# class this lane keeps retiring. The package is the unit.
if [ "$RELEVANT_CHECKER" -eq 1 ]; then
  precommit_run_gotest ./cmd/config-key-audit/ '' \
    'finding-code registry' \
    'finding-code registry: a declaration does not hold up (bugs_open/358)' \
    '   These are the arms that CANNOT run in the daily CronJob — that image ships no
   source tree and runs --no-source. Commit time is where they live, so a failure
   here is not advisory noise: it is the only place this gets caught.
   Most likely one of:
     * a `consumed` entry whose `reader` file:line no longer names its code
       (a reader reference pointing at the wrong file reads as a CLOSED LOOP);
     * a `reader_sink` the reader file never mentions;
     * a code declared with a disposition whose required field is missing.
   Re-check by hand:  ./scripts/audit-finding-codes.sh'
fi

if [ "$RELEVANT_SCAN" -eq 1 ]; then
  precommit_run_gotest ./platform/orchestration/actions/ '' \
    'finding-code scan' \
    'finding-code scan: this package writes a code the registry does not declare (bugs_open/358)' \
    '   A staged file in platform/orchestration/actions writes an ErrorCode: that is
   neither declared in finding_code_registry.json nor in its _scan_baseline — so
   it is NEW. Declare it in the SAME commit (consumed / instrumented /
   human-evidence / operational, or `unruled` if the decision is genuinely open).
   Catching it here is the whole point: LINK_CONTEXT_UNAVAILABLE reached the live
   table on 2026-08-24 because nothing ran this at commit time.'
fi
exit 0
