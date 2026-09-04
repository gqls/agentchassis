#!/usr/bin/env bash
# Run a Dynappraisal window UNDER A LOCK, so two sessions (or one session's
# retried command) can never walk the same output file concurrently.
#
# Why this exists: on 2026-09-04 a shell retry launched a SECOND copy of
# dynadot-appraise-all.sh against the same output file. The walker's resume is
# `grep -q "^$domain," <out>` immediately before the API call, so two writers
# race in that window: both miss, both call, both append. The quota is 300/day
# shared across three lanes, so a double-spend is a third of a day's coverage
# and the only symptom is an early 429 that reads as someone else's collision.
#
# Usage: run_appraisal_window.sh <queue.csv> [more_queues.csv ...]
# Appends every queue, in order, to the cumulative valuations file, stopping
# at the first 429 (the walker breaks; a stop is a clean end of a day's run).
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/../../../.." && pwd)"
# Cumulative estate valuations, resume-by-skip. APPRAISAL_OUT overrides it for
# a PROBE — a run whose subjects are deliberately not estate domains (e.g. the
# .com twin of a .uk we own, to test whether the appraiser is TLD-aware). Those
# must never land in the estate file: build_working_table.py joins it on domain,
# so a stray twin is inert today and indistinguishable from a real appraisal the
# day somebody buys that name.
OUT="${APPRAISAL_OUT:-$HERE/inbound/dynadot_valuations_2026-09-02.csv}"
LOCK="$HERE/.appraisal_window.lock"
WALKER="$REPO/scripts/domains/dynadot-appraise-all.sh"

[[ $# -ge 1 ]] || { echo "usage: run_appraisal_window.sh <queue.csv> [...]" >&2; exit 2; }
[[ -x "$WALKER" ]] || { echo "no walker at $WALKER" >&2; exit 2; }

exec 9>"$LOCK"
if ! flock -n 9; then
  echo "REFUSED: another appraisal window holds $LOCK (pid $(cat "$LOCK" 2>/dev/null))." >&2
  echo "The quota is shared — check that run finished before starting another." >&2
  exit 3
fi
echo $$ >&9

[[ -f "$OUT" ]] || echo "domain,valuation,currency,source" > "$OUT"
before=$(( $(wc -l < "$OUT") - 1 ))
for q in "$@"; do
  echo "=== $(date -Is) queue: $q"
  "$WALKER" "$q" "$OUT"
done
after=$(( $(wc -l < "$OUT") - 1 ))
echo "=== $(date -Is) window done: $before -> $after appraisals (+$((after-before)))"
dups=$(cut -d, -f1 "$OUT" | sort | uniq -d)
[[ -z "$dups" ]] && echo "duplicate check: clean" || { echo "⚠ DUPLICATES:"; echo "$dups"; }
