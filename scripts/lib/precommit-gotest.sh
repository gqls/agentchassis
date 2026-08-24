# scripts/lib/precommit-gotest.sh — SOURCED, never executed.
#
# The shared mechanics of an advisory, scoped, pre-commit Go-test guard. Two such
# guards now exist against cmd/config-key-audit — RFC_022's budget parity
# (check-optional-key-parity.sh) and bugs_open/358's finding-code registry
# (check-finding-code-registry.sh) — and before this file they were two
# independent hand-rolled implementations of the same shape: read the staged
# diff, decide relevance, run `go test`, tell a build failure apart from a real
# one, print advisory output, exit 0 whatever happens.
#
# WHY IT EXISTS. The council's reuse seat objected to the second one on exactly
# that ground (correlation be252395, round 2, medium): "two independently-built
# mechanisms solving overlapping problems in the same package, each maintained
# separately, with no unification once both exist". The seat is right, and this
# lane in particular has no standing to argue otherwise — retiring pairs of
# hand-maintained things that must stay in sync is most of what bugs_open/358 is.
#
# WHAT IS SHARED AND WHAT IS NOT. Shared: everything below. NOT shared, and
# deliberately left in each caller, is the RELEVANCE PREDICATE — one greps staged
# Go files for `ActionInputSpec`, the other reads reader paths out of a JSON
# registry — and the guidance text, which is the whole value of a guard firing.
# Parameterising those would produce a config file nobody can read, which is the
# other half of this estate's drift problem rather than a fix for it.
#
# THE ONE BEHAVIOUR THAT MUST NOT BE LOST, and the reason this is worth a shared
# file: A BUILD FAILURE IS NOT A FINDING. This tree is shared and often does not
# compile because of another session's work in progress (measured 2026-08-23: an
# uncommitted platform/livespec rename left this package unbuildable for hours).
# Reporting "your check failed" when the truth is "I could not tell" is the
# confident-wrong signal these guards exist to prevent — so the undecidable case
# says which it is, in both callers, from one implementation.

# precommit_staged_files — the staged diff, or empty. Never fails.
precommit_staged_files() {
  git diff --cached --name-only 2>/dev/null || true
}

# precommit_run_gotest <pkg> <run-filter|""> <subject> <failure-headline> <failure-body>
#
#   pkg              e.g. ./cmd/config-key-audit/
#   run-filter       a -run pattern, or "" for the whole package. A filter is a
#                    roster: a test whose name stops matching silently never runs.
#                    Prefer "".
#   subject          the NEUTRAL name of what is being checked ("optional-key
#                    parity"). Used for the could-not-tell line, and SEPARATE from
#                    the headline on purpose: the first cut of this helper reused
#                    the failure headline there and produced
#                    "…parity: DRIFTED (RFC_022): NOT CHECKED (the tree does not
#                    build)" — a single line asserting a finding and disclaiming
#                    one. Saying two contradictory things at once is worse than
#                    either, and this guard exists to keep an undecidable case
#                    legible as undecidable.
#   failure-headline the yellow heading for a REAL failure
#   failure-body     what the author should DO, already indented
#
# Prints nothing and returns 0 when the tests pass, when go is absent, or when
# the tree does not build (saying so, dimmed). ALWAYS returns 0: the pre-commit
# hook runs for every session on every commit, and a stray non-zero exit here
# stops the whole fleet committing.
precommit_run_gotest() {
  local pkg="$1" filter="$2" subject="$3" headline="$4" body="$5"
  command -v go >/dev/null 2>&1 || return 0

  local out rc
  if [ -n "$filter" ]; then
    out="$(go test "$pkg" -run "$filter" -count=1 2>&1)"
  else
    out="$(go test "$pkg" -count=1 2>&1)"
  fi
  rc=$?
  [ "$rc" -eq 0 ] && return 0

  # THE CLASSIFIER IS THE TOOLCHAIN'S OWN MARKER, not a bag of words. `go test`
  # prints "FAIL\t<pkg> [build failed]" on every compile/vet failure and
  # "[setup failed]" on a missing import or module — both unambiguous, both
  # absent from any ordinary test failure. The first cut matched
  # 'cannot find|undefined:|syntax error' instead, which a test's OWN failure
  # message can contain (a reader-file check saying "cannot find …" is the
  # obvious one), and a real failure matching it would have been swallowed as
  # NOT CHECKED — silently, which is the one thing this helper exists to
  # prevent. Measured 2026-08-24: no test in scope emits those tokens today;
  # tightened before one does.
  if printf '%s' "$out" | grep -qE '^FAIL[[:space:]].*\[(build|setup) failed\]'; then
    printf '\n\033[2m── %s: NOT CHECKED (the tree does not build — not a claim about your change) ──\033[0m\n' "$subject"
    printf '\033[2m   run it yourself once the tree compiles: go test %s\033[0m\n' "$pkg"
    return 0
  fi

  printf '\n\033[1;33m── %s ──\033[0m\n' "$headline"
  printf '%s\n' "$out" | grep -E '^(---|\s+\w+_test\.go|FAIL)' | head -8 | sed 's/^/   /'
  printf '%s\n' "$body"
  printf '   Advisory — this never blocks.\n'
  return 0
}
