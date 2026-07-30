#!/usr/bin/env bash
# VERIFY_273_before_apply.sh — the OPERATIONAL gate on migration 273.
#
# WHY THIS FILE EXISTS. The council gate's debug_historian seat objected at HIGH
# severity to the first submission of migration 273: the plan named the split-contract
# risk (DB CHECK widened before the Go gate is live) and then defused it with an
# ARGUMENT — "no producer exists yet, so an early apply is inert". The house discipline
# for exactly this situation is VERIFY AGAINST THE POD, never an inertness argument and
# never git/image-tag trust, because a same-tag rebuild ships a stale binary. The seat
# was right: this is the second attempt at the split that produced bugs_open/064, and
# ordering enforced only by a comment in a SQL header is the shape that produced it.
#
# So: run this. It exits non-zero unless the RUNNING chassis carries the Go half.
# It is the gate; the migration header is only documentation.
#
# NOT a numbered file, and NOT in sql_for_agents/ — deliberately. The migration runner
# takes every pending .sql in that directory, so a rollback or verify script living
# there could be applied as if it were a forward migration. Keep both here.
#
# Usage:  ./VERIFY_273_before_apply.sh            (check only, safe, read-only)
set -uo pipefail

NS=ai-persona-system
FAIL=0

say()  { printf '%s\n' "$*"; }
bad()  { printf 'FAIL: %s\n' "$*"; FAIL=1; }
good() { printf 'ok:   %s\n' "$*"; }

say "=== 1. Is the Go half in the RUNNING chassis? ==================================="
# WHY THIS IS A BUILD-DATE CHECK AND NOT A STRING GREP. All three obvious markers were
# tried against the live binary on 2026-07-30 and every one of them FAILED as a marker:
#
#   'component carries a PLAN…'   -> 0   it is a Go COMMENT; comments never reach a binary
#   'is a split contract; …'      -> 0   likewise a comment — useless as a control
#   bare 'component'              -> 761 matches ('content_components', 'component_level', …)
#
# The change adds "component" to a []string literal. Slice literals ARE in rodata —
# 'experience-pattern' greps 1, which is how we know — but the token this change adds is
# short and extremely common, and the literals are not laid out contiguously
# ('actionexperience-pattern' and 'experience-patterncomponent' both grep 0). So there is
# NO unique new string to grep for, which is precisely the documented case for dating the
# build instead. The control grep is retained to prove the binary is readable at all, so a
# tooling failure cannot masquerade as a clean answer.
CTL_MARKER='experience-pattern'
GO_HALF_COMMIT='c659e312b'

PODS=$(kubectl get pods -n "$NS" -o name 2>/dev/null | grep 'agent-chassis' | cut -d/ -f2)
if [ -z "$PODS" ]; then
  bad "no agent-chassis pods found in $NS — cannot verify; refusing to bless the migration"
else
  # count <pod> <marker> — echoes a bare integer, ALWAYS.
  #
  # ⚠ THIS FUNCTION EXISTS BECAUSE THE FIRST VERSION OF THIS SCRIPT REPORTED A FALSE
  # PASS, on its very first run, against pods that demonstrably lacked the marker.
  # `grep -c` EXITS 1 WHEN THE COUNT IS ZERO. The original line was
  #     NEW=$(kubectl exec ... "grep -ac '$M' /app/agent-chassis" || echo 0)
  # so on zero matches grep printed "0" AND the `|| echo 0` fallback printed another,
  # giving the two-line string "0\n0". That is not equal to "0", so `[ "$CTL" = "0" ]`
  # was false, the guard fell through to the else branch, and the script announced
  # "Go half present (new=0 0, control=0 0)" — a green verdict built out of two zeroes.
  # Exactly the failure class this lane exists to stop: a gate whose only untested
  # branch is the one that refuses. Normalise to one integer, and fail CLOSED.
  count() {
    kubectl exec -n "$NS" "$1" -- sh -c "grep -ac '$2' /app/agent-chassis 2>/dev/null; true" 2>/dev/null \
      | tr -dc '0-9\n' | head -1 | sed 's/^$/0/'
  }

  # The commit that added "component" to validDocSubjectTypes. A binary older than this
  # cannot contain it. Committer date, epoch seconds — not the author date.
  COMMIT_EPOCH=$(git -C "$(git rev-parse --show-toplevel)" log -1 --format=%ct "$GO_HALF_COMMIT" 2>/dev/null)
  if [ -z "${COMMIT_EPOCH:-}" ]; then
    bad "cannot resolve commit $GO_HALF_COMMIT — refusing to bless the migration on an unknown baseline"
    COMMIT_EPOCH=99999999999
  else
    say "note: Go half committed at epoch $COMMIT_EPOCH ($(date -u -d "@$COMMIT_EPOCH" '+%Y-%m-%d %H:%M:%S UTC'))"
  fi

  for POD in $PODS; do
    # No `strings` in these images — grep the binary directly with -a.
    CTL=$(count "$POD" "$CTL_MARKER"); CTL=${CTL:-0}
    BIN_EPOCH=$(kubectl exec -n "$NS" "$POD" -- sh -c 'stat -c %Y /app/agent-chassis 2>/dev/null; true' 2>/dev/null | tr -dc '0-9\n' | head -1)
    BIN_EPOCH=${BIN_EPOCH:-0}

    if ! [ "$CTL" -gt 0 ] 2>/dev/null; then
      bad "$POD: CONTROL marker '$CTL_MARKER' absent (=$CTL) — the binary is not readable by this grep (wrong path? stripped?). Every other reading from this pod is therefore worthless."
    elif ! [ "$BIN_EPOCH" -gt 0 ] 2>/dev/null; then
      bad "$POD: could not read the binary's mtime — cannot date the build, so cannot clear it."
    elif [ "$BIN_EPOCH" -lt "$COMMIT_EPOCH" ]; then
      bad "$POD: binary built $(date -u -d "@$BIN_EPOCH" '+%Y-%m-%d %H:%M:%S UTC'), BEFORE the Go half was committed — this pod cannot carry 'component'. DO NOT APPLY 273."
    else
      good "$POD: binary built $(date -u -d "@$BIN_EPOCH" '+%Y-%m-%d %H:%M:%S UTC'), after the Go half (control=$CTL)"
      say  "      ⚠ build-date is NECESSARY, not SUFFICIENT: a same-tag rebuild from a stale"
      say  "        checkout would also date late. The definitive check is behavioural —"
      say  "        see section 4 — and it is the only one that proves the vocabulary works."
    fi
  done
fi

say ""
say "=== 2. What do the constraints say RIGHT NOW? =================================="
kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -F'|' -c "
SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint
 WHERE conname IN ('doc_plans_subject_type_check','doc_notes_subject_type_check')
 ORDER BY conname;" 2>/dev/null | while IFS='|' read -r name def; do
  [ -z "$name" ] && continue
  case "$name:$def" in
    doc_notes_subject_type_check:*landmine*) good "$name still allows 'landmine' (the 57-row corpus is safe)";;
    doc_notes_subject_type_check:*)          bad  "$name does NOT allow 'landmine' — 273 will refuse, and it is right to; apply 270 first";;
  esac
  case "$def" in
    *component*) say "note: $name ALREADY allows 'component' (273 already applied here)";;
  esac
done

say ""
say "=== 3. Is anything already writing component docs? ============================="
CNT=$(kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c \
  "SELECT count(*) FROM doc_plans WHERE subject_type='component';" 2>/dev/null || echo "?")
say "doc_plans rows with subject_type='component': ${CNT:-?}"
say "(0 is expected pre-apply. The inertness claim in 273's header depends on this being 0;"
say " it is printed rather than asserted so the claim is checkable rather than believed.)"

say ""
say "=== 4. THE DEFINITIVE CHECK — behavioural, and it can only run AFTER 273 ======="
say "Sections 1-3 are pre-flight. None of them proves the vocabulary WORKS; they only"
say "prove it is not obviously impossible. The gate's own doctrine is to verify through"
say "the caller's real path, so after applying 273 against a post-commit image, prove it"
say "end to end by writing and reading one component doc:"
say ""
say "  -- write (should SUCCEED; before the image+migration it fails the CHECK or the Go gate)"
say "  INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)"
say "  VALUES ('component','teaser-reveal-panel','# PLAN probe','verify-273','staged_component_build');"
say ""
say "  -- read back through the ACTION, not the table: a load_doc_context step with"
say "  -- subject_type='component' must return the body rather than"
say "  -- 'unsupported subject_type \"component\"' from docSubjectGateReason."
say ""
say "  -- then clean up"
say "  DELETE FROM doc_plans WHERE source='verify-273';"
say ""
say "Until that has been watched to pass, 'component' is a capability nobody has exercised."
say ""
if [ "$FAIL" -ne 0 ]; then
  say "RESULT: DO NOT APPLY migration 273. Fix the above first."
  exit 1
fi
say "RESULT: pre-flight clear — safe to apply migration 273, then do section 4."
say "After applying, re-run this script: section 2 should then report 'component' on BOTH"
say "constraints, and 'landmine' still present on doc_notes."
exit 0
