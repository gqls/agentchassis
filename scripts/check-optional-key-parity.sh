#!/usr/bin/env bash
# check-optional-key-parity.sh — ADVISORY. Runs RFC_022's parity guard, but only
# when the staged diff could actually have broken it.
#
# WHY THIS EXISTS. `cmd/config-key-audit/optional_budget_cron_parity_test.go`
# compares the budget cron's hand-generated literal against the Go registry. It
# works — and on 2026-08-18 it was found FAILING at HEAD, having been so for days:
# two actions (`retract_asset_files` 4 keys, `publish_site` 3) had entered the
# registry counted as ZERO, invisible to the accumulation check RFC_022 exists to
# provide. Nobody had run it, because it lives in a package no ordinary change
# touches. A guard that only runs when someone remembers is the failure mode this
# estate keeps paying for, so it now runs at the one moment the author is present.
#
# SCOPED, because every commit should not pay for a Go test: it fires only when the
# staged diff touches the cron's literal, the acknowledgements file, or a Go file
# that declares an ActionInputSpec.
#
# ADVISORY, never blocking — same rule as commit-scope-report and pattern-check.
# The pre-commit hook runs for every session on every commit; a stray non-zero exit
# here would stop the whole fleet committing.
#
# ⚠ A BUILD FAILURE IS NOT PARITY DRIFT. This tree is shared and often does not
# compile because of another session's work in progress. Reporting "the literal has
# drifted" when the truth is "I could not tell" would be exactly the kind of
# confident wrong signal the guard exists to prevent, so the two are separated and
# the undecidable case says so.

set -uo pipefail
ROOT="$(git rev-parse --show-toplevel 2>/dev/null)" || exit 0
cd "$ROOT" || exit 0

STAGED="$(git diff --cached --name-only 2>/dev/null)" || exit 0
[ -n "$STAGED" ] || exit 0

RELEVANT=0
while IFS= read -r f; do
  case "$f" in
    deployments/kustomize/services/optional-key-budget-check/base/check.py) RELEVANT=1 ;;
    */architecture_review/optional_key_budget_acks.json)                    RELEVANT=1 ;;
    cmd/config-key-audit/*)                                                 RELEVANT=1 ;;
    *.go)
      # Only Go files that declare an input spec can move the counts.
      if git show ":$f" 2>/dev/null | grep -qE 'ActionInputSpec\{|RegisterActionInputSpec\('; then
        RELEVANT=1
      fi
      ;;
  esac
done <<< "$STAGED"

[ "$RELEVANT" -eq 1 ] || exit 0
command -v go >/dev/null 2>&1 || exit 0

OUT="$(go test ./cmd/config-key-audit/ -run 'BudgetCron' -count=1 2>&1)"
RC=$?
[ "$RC" -eq 0 ] && exit 0

if printf '%s' "$OUT" | grep -qE 'build failed|cannot find|undefined:|redeclared|syntax error'; then
  printf '\n\033[2m── optional-key parity: NOT CHECKED (the tree does not build — not a parity claim) ──\033[0m\n'
  printf '\033[2m   run it yourself once the tree compiles: go test ./cmd/config-key-audit/\033[0m\n'
  exit 0
fi

printf '\n\033[1;33m── optional-key budget parity: DRIFTED (RFC_022) ──\033[0m\n'
printf '%s\n' "$OUT" | grep -E 'check\.py|FAIL|acks' | head -8 | sed 's/^/   /'
cat <<'MSG'
   The budget cron's literal no longer matches the Go registry. An action missing
   from it is counted as ZERO optional keys — permanently invisible to the
   accumulation check, while the cron keeps writing a clean row every morning.
   Regenerate (the command is in check.py's own comment), then RE-APPLY the overlay
   or the cluster keeps the old literal:
     kubectl apply -k deployments/kustomize/services/optional-key-budget-check/overlays/production/uk_001
   Advisory — this never blocks.
MSG
exit 0
