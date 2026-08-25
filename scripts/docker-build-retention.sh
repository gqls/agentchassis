#!/usr/bin/env bash
# docker-build-retention.sh — keep the newest N RELEASE BUILDS locally, drop older ones.
# DRY RUN BY DEFAULT; --apply deletes.
#
# WHY A COUNT AND NOT A SIZE (owner ruling, 2026-08-25). Docker's own build-cache
# GC understands only two axes -- a byte ceiling (`keepStorage`) and an age
# (`unused-for`) -- because the CACHE is a shared pool of layer/step records with
# no per-build unit in it to count. "Old builds", as a countable thing, live on
# the IMAGE side: one release tag (v1.0.NNNN) per `make release`, ~25 images under
# it, one per service. That is what this bounds. Cache age GC is a separate knob
# and this script prints the config for it rather than editing daemon.json, which
# needs root.
#
# THE SCALE, MEASURED 2026-08-25: 97 distinct release tags, v1.0.1229..v1.0.1339,
# ~1,020 images, 104 GB. Cadence is BURSTY -- 2 to 19 releases in a day (19 on
# 08-19) -- so a tag count does NOT translate to a fixed number of days, and that
# is exactly why the --min-age-hours floor below is not optional.
#
# WHAT IT WILL NOT DO, and each of these is load-bearing:
#  * it never touches a tag that is not release-shaped (^v1\.0\.[0-9]+$) -- base
#    images like 20-alpine and 3-alpine are pull-cached, small, and needed;
#  * it never touches the newest N, so a build finishing right now is always safe.
#    That was the hazard that ruled out `docker image prune -a`: push-*/deploy-*
#    are git-blind and ship whatever is tagged LOCALLY, so a locally-built,
#    not-yet-pushed release image is a session's work;
#  * it never touches anything younger than --min-age-hours, whatever N says --
#    the burst days above are why a pure count is not safe on its own;
#  * it never touches an image a container references.
#
# Usage:
#   scripts/docker-build-retention.sh                     # dry run at the default N
#   scripts/docker-build-retention.sh --keep 30
#   scripts/docker-build-retention.sh --keep 30 --apply
#   scripts/docker-build-retention.sh --plan             # what several N values would cost
#   scripts/docker-build-retention.sh --self-test
#
# Exit: 0 ran · 1 nothing to do · 2 refused

set -uo pipefail

KEEP=25; MIN_AGE_H=48; APPLY=""; PLAN=""; SELFTEST=""
RELEASE_RE='^v1\.0\.[0-9]+$'

while [[ $# -gt 0 ]]; do
    case "$1" in
        --keep)           [[ $# -ge 2 ]] || { echo "need a number after --keep" >&2; exit 2; }; KEEP="$2"; shift 2 ;;
        --min-age-hours)  [[ $# -ge 2 ]] || { echo "need a number after --min-age-hours" >&2; exit 2; }; MIN_AGE_H="$2"; shift 2 ;;
        --apply)          APPLY=1; shift ;;
        --plan)           PLAN=1; shift ;;
        --self-test)      SELFTEST=1; shift ;;
        -h|--help)        sed -n '2,40p' "$0"; exit 0 ;;
        *)                echo "unknown argument $1" >&2; exit 2 ;;
    esac
done

[[ "$KEEP" =~ ^[0-9]+$ ]] || { echo "--keep must be a number" >&2; exit 2; }
[[ "$MIN_AGE_H" =~ ^[0-9]+$ ]] || { echo "--min-age-hours must be a number" >&2; exit 2; }

# A retention of a handful of releases is not a policy, it is an outage waiting
# for a bad release. The fleet ships several times a day; keep days, not minutes.
if (( KEEP < 5 )); then
    echo "docker-build-retention: REFUSING --keep $KEEP — under 5 release tags is less than a" >&2
    echo "  day's shipping here (2-19 releases/day, measured 2026-08-25) and leaves no rollback." >&2
    exit 2
fi
if (( MIN_AGE_H < 12 )); then
    echo "docker-build-retention: REFUSING --min-age-hours $MIN_AGE_H — the floor exists because the" >&2
    echo "  release cadence is bursty; 19 tags landed on one day, so N alone can reach into today." >&2
    exit 2
fi

command -v docker >/dev/null || { echo "docker not on PATH" >&2; exit 2; }
docker info >/dev/null 2>&1 || { echo "docker daemon unreachable" >&2; exit 2; }

# ---- inventory -------------------------------------------------------------
# %-separated because a tag never contains one and CreatedAt does contain spaces.
LIST="$(docker images --format '{{.Repository}}%{{.Tag}}%{{.ID}}%{{.CreatedAt}}' 2>/dev/null)"
[[ -n "$LIST" ]] || { echo "docker-build-retention: no images."; exit 1; }

# Newest creation timestamp per release tag, as epoch seconds.
declare -A TAG_NEWEST TAG_COUNT
# ⚠ Docker's CreatedAt is "2026-08-25 19:39:29 +0100 BST" — a numeric offset AND a
# zone NAME. `date -d` rejects that outright, so the first cut of this loop had
# every row fail its date parse and skip on `|| continue`. The script then printed
# "nothing to remove", which on a RETENTION tool reads as "you are within policy"
# — a silent no-op wearing the shape of a clean bill of health. Trim to the offset.
parse_epoch() {
    local c="${1% *}"                       # drop the trailing zone NAME
    date -d "$c" +%s 2>/dev/null || date -d "${c% *}" +%s 2>/dev/null
}
parsed=0; unparsed=0
while IFS='%' read -r repo tag id created; do
    [[ "$tag" =~ $RELEASE_RE ]] || continue
    if ! ep="$(parse_epoch "$created")" || [[ -z "$ep" ]]; then
        unparsed=$((unparsed+1)); continue
    fi
    parsed=$((parsed+1))
    TAG_COUNT["$tag"]=$(( ${TAG_COUNT["$tag"]:-0} + 1 ))
    (( ep > ${TAG_NEWEST["$tag"]:-0} )) && TAG_NEWEST["$tag"]=$ep
done <<< "$LIST"

if (( unparsed > 0 && parsed == 0 )); then
    echo "docker-build-retention: REFUSING — could not parse the creation date of ANY of the" >&2
    echo "  $unparsed release image(s). A retention tool that cannot read dates must not report" >&2
    echo "  'nothing to remove'; that is indistinguishable from being within policy." >&2
    exit 2
fi
(( unparsed > 0 )) && echo "docker-build-retention: WARNING — $unparsed image(s) had an unparseable date and were skipped." >&2
if (( ${#TAG_NEWEST[@]} == 0 )); then echo "docker-build-retention: no release-shaped tags."; exit 1; fi

# Version-sorted, newest first.
mapfile -t TAGS < <(printf '%s\n' "${!TAG_NEWEST[@]}" | sort -rV)
NOW="$(date +%s)"

# Images a container references must never be removed.
PROTECTED_IDS="$(docker ps -a --format '{{.Image}}' 2>/dev/null | sort -u)"

if [[ -n "$PLAN" ]]; then
    echo "docker-build-retention: ${#TAGS[@]} release tags, ${TAGS[0]} down to ${TAGS[-1]}"
    printf '\n  %-6s %-8s %s\n' "keep" "would go" "oldest kept"
    for n in 10 20 25 30 40 60; do
        (( n >= ${#TAGS[@]} )) && continue
        printf '  %-6s %-8s %s\n' "$n" "$(( ${#TAGS[@]} - n )) tags" "${TAGS[$((n-1))]}"
    done
    echo
    echo "  cadence is bursty (2-19/day), so translate with care, and the --min-age-hours"
    echo "  floor (default ${MIN_AGE_H}h) applies on top of whichever N you pick."
    exit 0
fi

# ---- select ----------------------------------------------------------------
DOOMED=(); SKIPPED_AGE=0; SKIPPED_USED=0
for i in "${!TAGS[@]}"; do
    (( i < KEEP )) && continue                     # newest N always kept
    t="${TAGS[$i]}"
    age_h=$(( (NOW - ${TAG_NEWEST["$t"]}) / 3600 ))
    if (( age_h < MIN_AGE_H )); then SKIPPED_AGE=$((SKIPPED_AGE+1)); continue; fi
    if grep -qF ":$t" <<< "$PROTECTED_IDS"; then SKIPPED_USED=$((SKIPPED_USED+1)); continue; fi
    DOOMED+=("$t")
done

echo "docker-build-retention: ${#TAGS[@]} release tags (${TAGS[0]} .. ${TAGS[-1]})"
echo "  keeping newest $KEEP, and anything under ${MIN_AGE_H}h old regardless"
(( SKIPPED_AGE  )) && echo "  $SKIPPED_AGE tag(s) past N but held by the age floor"
(( SKIPPED_USED )) && echo "  $SKIPPED_USED tag(s) held because a container references them"

if (( ${#DOOMED[@]} == 0 )); then echo "  nothing to remove."; exit 1; fi

tot=0; for t in "${DOOMED[@]}"; do tot=$(( tot + ${TAG_COUNT["$t"]} )); done
echo "  ${#DOOMED[@]} tag(s) / $tot image(s) to remove: ${DOOMED[0]} .. ${DOOMED[-1]}"

if [[ -z "$APPLY" ]]; then
    echo "  DRY RUN — nothing removed. Re-run with --apply."
    printf '    %s (%s images)\n' "${DOOMED[0]}" "${TAG_COUNT[${DOOMED[0]}]}"
    (( ${#DOOMED[@]} > 2 )) && echo "    … ${#DOOMED[@]} tags …"
    printf '    %s (%s images)\n' "${DOOMED[-1]}" "${TAG_COUNT[${DOOMED[-1]}]}"
    exit 0
fi

before="$(df -B1 --output=avail / | tail -1 | tr -d ' ')"
removed=0
for t in "${DOOMED[@]}"; do
    while IFS='%' read -r repo tag id created; do
        [[ "$tag" == "$t" ]] || continue
        docker rmi "$repo:$tag" >/dev/null 2>&1 && removed=$((removed+1))
    done <<< "$LIST"
done
after="$(df -B1 --output=avail / | tail -1 | tr -d ' ')"
awk -v r="$removed" -v t="${#DOOMED[@]}" -v b="$before" -v a="$after" \
  'BEGIN {printf "  removed %d image(s) across %d tag(s); / freed %.1f GB\n", r, t, (a-b)/2^30}'
exit 0
