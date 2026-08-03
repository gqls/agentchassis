#!/usr/bin/env bash
# council-coverage-nudge.sh — ADVISORY. Runs from the commit-msg hook. If this
# commit touches platform CODE (platform/ internal/ pkg/) and the message carries
# NO `Council-Reviewed:` trailer, print a short nudge. It NEVER blocks — it always
# exits 0 — so a bug here can never stop the fleet committing.
#
# WHY (owner ruling 2026-07-24: "strengthen advisory, defer structural"). The
# council gate is deliberately advisory — it cannot intercept a direct commit to
# the shared branch. Until 2026-07-22 that was moot, because an APPROVED verdict
# was effectively unreachable (~5% approval; the decision rule ignored objection
# severity — bugs_open/057). That is fixed and live (v1.0.1149): approval ran ~80%
# over the two days after. So the norm is now worth reinforcing at the one moment
# it bites — an un-reviewed platform commit — WITHOUT enforcing it (PR-mode stays
# deferred). This is the "loud" half of loud-and-regular; the "regular" half is a
# scheduled 098 coverage report.
#
# It is a commit-msg hook, not pre-commit, ON PURPOSE: only commit-msg can read
# the message, so the trailer check makes this fire on the un-reviewed commits and
# stay silent on the reviewed ones — precise, not naggy.
set -u
trap 'exit 0' ERR

MSGFILE="${1:-}"
[ -n "$MSGFILE" ] && [ -f "$MSGFILE" ] || exit 0

# ── Trailer SHAPE validation (added 2026-08-03, bugs_open/185 lane) ──────────────
# Runs on EVERY commit with either trailer, before the platform-code gate below —
# a placeholder is worth catching whatever the diff touches. The correlation these
# trailers carry is a UUID minted by 097_TRIGGER_council_review_v1.sh
# (`/proc/sys/kernel/random/uuid`); 098 also accepts a PREFIX of one (LIKE
# '${id}%'), so a short hex id is legitimate — only hex-and-dash is valid, of at
# least 8 characters.
#
# WHY: found live on commit 171d52a1a — `Council-Submitted: pending-185-tranche-1`,
# a placeholder typed before the submission ran, meant to be "filled in later".
# There is no later: forward-only forbids an amend, so the trailer is a JOIN KEY
# that resolves to nothing, permanently, and 098 can never credit that commit.
# Advisory only, same as everything else in this hook — it cannot undo a bad
# trailer, only warn on the NEXT one before it is made permanent.
for trailer_name in Council-Submitted Council-Reviewed; do
  trailer_value=$(grep -iE "^[[:space:]]*${trailer_name}:" "$MSGFILE" \
                    | head -1 | sed -E "s/^[[:space:]]*${trailer_name}:[[:space:]]*//I")
  [ -n "$trailer_value" ] || continue
  if ! printf '%s' "$trailer_value" | grep -qE '^[0-9a-fA-F-]{8,}$'; then
    printf '\n\033[1;31m── %s trailer looks like a placeholder, not a correlation ──\033[0m\n' "$trailer_name"
    printf '  Value: %s\n' "$trailer_value"
    printf '  This trailer is a JOIN KEY for the 098 coverage report, not a note to\n'
    printf '  yourself — a value that is not a UUID (or a hex prefix of one, 8+ chars)\n'
    printf '  resolves to nothing, and forward-only forbids fixing it with an amend.\n'
    printf '  Submit FIRST (097_TRIGGER_council_review_v1.sh prints the correlation in\n'
    printf '  seconds), THEN commit with the id it printed. Advisory — never blocks.\n'
  fi
done

# Platform CODE only (owner scope: platform/ internal/ pkg/). Docs/site content
# never spend council credits and never need the trailer, so never nudge on them.
plat=$(git diff --cached --name-only --diff-filter=ACMR 2>/dev/null \
        | grep -E '^(platform|internal|pkg)/' || true)
[ -z "$plat" ] && exit 0

# Already reviewed? An APPROVED verdict earns the trailer; its presence means the
# author asserts review, so stay silent (098 buckets a false trailer separately —
# not this hook's job).
if grep -qiE '^[[:space:]]*Council-Reviewed:' "$MSGFILE"; then
  exit 0
fi

n=$(printf '%s\n' "$plat" | grep -c . || true)

# Submitted but not yet judged (the `Council-Submitted:` trailer, added
# 2026-07-30) is the CORRECT state for a thread following the 2026-07-20 rule to
# commit the moment the work is coherent — the verdict takes ~30 minutes and
# forward-only forbids the amend that `Council-Reviewed:` would need. Nudging
# that thread as "un-reviewed" punishes compliance and teaches everyone to
# ignore the nudge, which costs the advisory its only power. So acknowledge it
# instead: the trailer is live, 098 resolves the verdict at report time, and the
# one thing still owed is READING that verdict.
#
# Found 2026-07-31 by the bugfix_143 lane, which was nudged on a commit carrying
# `Council-Submitted:` for a submission that was mid-council at the time.
if grep -qiE '^[[:space:]]*Council-Submitted:' "$MSGFILE"; then
  printf '\n\033[1;36m── council coverage · advisory ──\033[0m\n'
  printf '  %s platform-code file(s) staged with a \033[1mCouncil-Submitted\033[0m trailer — verdict pending.\n' "$n"
  printf '  Correct for a pre-verdict commit: 098 credits this automatically once the\n'
  printf '  correlation is approved, with no amend. Still owed: READ the verdict, and act\n'
  printf '  on a REVISE/REJECTED (the code is already on the shared branch).\n'
  exit 0
fi

printf '\n\033[1;36m── council coverage · advisory ──\033[0m\n'
printf '  %s platform-code file(s) staged, no \033[1mCouncil-Reviewed\033[0m trailer.\n' "$n"
printf '  The gate now approves ~80%% of sound changes in ~30 min (submit: 097_TRIGGER_council_review_v1.sh).\n'
printf '  Committing before the verdict? Use \033[1mCouncil-Submitted: <corr>\033[0m — it asserts nothing,\n'
printf '  and 098 credits the commit when approval lands.\n'
printf '  This commit will list as un-reviewed in the 098 report. Advisory — never blocks.\n'
exit 0
