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

# psql helper, defined HERE (before section 2) rather than mid-file.
# ⚠ WHY: I first wrote section 2's landmine count using a helper named `q` that is defined
# in a DIFFERENT script in this directory, and `psql_do` below was defined AFTER the point
# of use. The result printed "ok: corpus intact at ? rows" — a GREEN line wrapped around a
# measurement that had failed with `q: command not found` on stderr. Ninth instance in two
# days of the one class this lane exists to defeat, and this one I introduced while fixing
# a different staleness in the same file. `</dev/null` on the call matters too: it is used
# inside a `while read` loop, and `kubectl exec -i` will otherwise eat the loop's stdin.
psql_do() {
  kubectl -n "$NS" exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -t -A -v ON_ERROR_STOP=1 -c "$1" 2>&1 </dev/null
}

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
    doc_notes_subject_type_check:*landmine*)
      LM=$(psql_do "SELECT count(*) FROM doc_notes WHERE categories ? 'landmine';" | tr -dc '0-9')
      if [ -n "$LM" ]; then
        good "$name still allows 'landmine' — corpus intact at $LM rows (LIVE count: read 57 on 07-30, 190 on 07-31 — never hardcode it)"
      else
        bad "$name allows 'landmine' BUT the corpus count could not be measured — do not read this as intact"
      fi;;
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
say "=== 4. THE DEFINITIVE CHECK — RUN, not printed ================================="
# WHY THIS SECTION EXECUTES INSTEAD OF PRINTING INSTRUCTIONS. Council round 2 approved
# this change with three advisory objections, and all three said the same thing
# independently: build-dating is 'necessary, not sufficient' by my own admission, and the
# sufficient probe was left as prose for a human to remember. bug_historian named the
# shape exactly — "a gate whose only untested branch is the one that refuses was already
# this exact workstream's own bug once (the grep -c / || echo 0 false-pass). Leaving the
# sufficient probe optional reproduces the same shape one layer up." That is correct, so
# the probe runs here, and its result decides this script's exit code.
#
# IT IS A RED/GREEN PAIR, WHICH IS THE POINT. The script reads the live constraint and
# orients itself, so the SAME probe is meaningful before and after the migration:
#   constraint narrow  -> the INSERT MUST be refused  (the red half, free, provable today)
#   constraint widened -> the INSERT MUST succeed and read back (the green half)
# A probe that can only ever pass is the thing this lane exists to refuse.
PROBE_KEY='__verify273_probe__'
WIDENED=$(psql_do "SELECT pg_get_constraintdef(oid) LIKE '%component%'
                     FROM pg_constraint
                    WHERE conrelid='public.doc_plans'::regclass
                      AND conname='doc_plans_subject_type_check';" | tr -d '[:space:]')

# Clean any leftover from an interrupted run before probing, so a stale row cannot make
# the insert fail for the wrong reason and read as a correct refusal.
psql_do "DELETE FROM doc_plans WHERE subject_key='$PROBE_KEY';" >/dev/null 2>&1

INS=$(psql_do "INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
               VALUES ('component','$PROBE_KEY','# probe','verify-273','staged_component_build');")

if [ "$WIDENED" = "t" ]; then
  # GREEN half: post-migration. The write must land and be readable.
  if printf '%s' "$INS" | grep -qi 'ERROR'; then
    bad "post-migration, but the probe INSERT was refused: $INS"
  else
    BODY=$(psql_do "SELECT body FROM doc_plans WHERE subject_key='$PROBE_KEY' AND subject_type='component';" | tr -d '[:space:]')
    if [ "$BODY" = "#probe" ]; then
      good "probe wrote AND read back a subject_type='component' PLAN — the vocabulary works"
    else
      bad "probe wrote but read back '$BODY' instead of the body — something else is wrong"
    fi
  fi
  psql_do "DELETE FROM doc_plans WHERE subject_key='$PROBE_KEY';" >/dev/null 2>&1
  say "note: the probe row is deleted. The REMAINING half of the definitive check cannot be"
  say "      done from psql — read it back through load_doc_context and confirm"
  say "      docSubjectGateReason no longer returns 'unsupported subject_type'. Until that"
  say "      has been watched, the Go gate is verified only by build date."
else
  # RED half: pre-migration. The write MUST be refused, and by the CHECK specifically.
  if printf '%s' "$INS" | grep -qi 'doc_plans_subject_type_check'; then
    good "probe correctly REFUSED by doc_plans_subject_type_check — the red half of the pair passes, so this probe can distinguish states"
  elif printf '%s' "$INS" | grep -qi 'ERROR'; then
    bad "probe was refused, but NOT by the subject_type CHECK — read the error before trusting anything above: $INS"
  else
    bad "probe INSERT SUCCEEDED while the constraint still reads as narrow — the constraint is not what section 2 reported. STOP."
    psql_do "DELETE FROM doc_plans WHERE subject_key='$PROBE_KEY';" >/dev/null 2>&1
  fi
fi
say ""
if [ "$FAIL" -ne 0 ]; then
  say "RESULT: DO NOT APPLY migration 273. Fix the above first."
  exit 1
fi
if [ "$WIDENED" = "t" ]; then
  say "RESULT: 273 is ALREADY APPLIED and the DB half is PROVEN — the probe in section 4"
  say "wrote and read back a subject_type='component' PLAN."
  say "STILL UNVERIFIED, and it is the only half left: the GO gate. Read a component PLAN"
  say "back through load_doc_context and confirm docSubjectGateReason no longer returns"
  say "'unsupported subject_type'. Until then the Go half rests on build date alone, which"
  say "section 1 says is necessary and not sufficient."
else
  say "RESULT: pre-flight clear — safe to apply migration 273, then re-run this script:"
  say "section 2 should report 'component' on BOTH constraints with 'landmine' still on"
  say "doc_notes, and section 4's probe should flip from refusing to writing."
fi
exit 0
