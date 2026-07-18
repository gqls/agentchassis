#!/usr/bin/env bash
# commit-scope-report.sh — ADVISORY. Prints what this commit will actually
# contain, grouped by area, so a file belonging to another session is visible
# BEFORE it lands. It never blocks: it always exits 0.
#
# WHY ADVISORY AND NOT A THRESHOLD (measured 2026-07-18, do not "improve" this
# into a blocker without re-measuring):
#   Breadth does not predict damage here. The commit that swept another thread's
#   makefile change (69d6f3ecc) was 16 files across 3 areas — inside the normal
#   range (last 300 commits: median 2 files, p90 14, p95 29; 90% touch <=2
#   areas). Every other known bundling commit was 5-9 files / 2-3 areas. So any
#   threshold low enough to catch the real cases fires constantly on ordinary
#   work, and one high enough to be quiet catches nothing. The harm is "this
#   commit carries one file that isn't mine", which no breadth rule can see.
#   Git has no notion of which session staged what in a shared working tree, so
#   the honest intervention is visibility, not judgement.
#
# It cannot catch a SAME-FILE passenger (two sessions editing one file; whoever
# commits takes both edits). Only separate working trees fix that.
# Context: docs/agent_docs/docs024_key_docs_latest/multi_session_coordination/
set -u

# Never let this script break a commit, whatever happens below.
trap 'exit 0' ERR

files=$(git diff --cached --name-only --diff-filter=ACMRD 2>/dev/null | grep . || true)
[ -z "$files" ] && exit 0

n=$(printf '%s\n' "$files" | wc -l | tr -d ' ')
# A single-file commit has nothing to scan for passengers.
[ "$n" -le 1 ] && exit 0

# Group by top-level path element (a bare file like `makefile` is its own area).
areas=$(printf '%s\n' "$files" | awk -F/ '{print (NF>1 ? $1"/" : $1)}' | sort -u)
na=$(printf '%s\n' "$areas" | grep -c . || true)

printf '\n\033[1;33m── commit scope: %s files across %s area(s) ──\033[0m\n' "$n" "$na"
while IFS= read -r a; do
  [ -z "$a" ] && continue
  if [ "${a%/}" = "$a" ]; then
    match=$(printf '%s\n' "$files" | grep -x -- "$a" || true)     # bare file
  else
    match=$(printf '%s\n' "$files" | grep -- "^$a" || true)
  fi
  c=$(printf '%s\n' "$match" | grep -c . || true)
  printf '   \033[1m%-22s\033[0m %3s  ' "$a" "$c"
  printf '%s' "$(printf '%s\n' "$match" | head -3 | sed "s|^$a||" | paste -sd, - | sed 's/,/, /g')"
  [ "$c" -gt 3 ] && printf ' … +%s more' "$((c - 3))"
  printf '\n'
done <<< "$areas"

printf '   \033[2m↳ any file above that is not part of YOUR task belongs to another session —\033[0m\n'
printf '   \033[2m  commit with an explicit pathspec to leave it out, or say "sweep:" in the message.\033[0m\n\n'

exit 0
