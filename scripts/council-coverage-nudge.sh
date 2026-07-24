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
printf '\n\033[1;36m── council coverage · advisory ──\033[0m\n'
printf '  %s platform-code file(s) staged, no \033[1mCouncil-Reviewed\033[0m trailer.\n' "$n"
printf '  The gate now approves ~80%% of sound changes in ~30 min (submit: 097_TRIGGER_council_review_v1.sh).\n'
printf '  This commit will list as un-reviewed in the 098 report. Advisory — never blocks.\n'
exit 0
