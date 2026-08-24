#!/usr/bin/env bash
# scratch-janitor.sh — reap abandoned build scratch from /tmp AND from the disk
# scratch root, so neither fills. DRY RUN BY DEFAULT; --apply deletes.
#
# WHY THIS EXISTS. The HEAD-verification recipe extracts committed HEAD into a
# fresh directory (~450MB) and NOTHING deletes it afterwards. The `rm -rf` you
# see in the pasted copies is the SETUP half — it clears the directory the
# recipe is about to use, so it only ever reclaims a tree of the SAME NAME. In
# practice each run picks a new name (headtree, headtree2, headfinal, ht5, ht6
# — six from one session in one morning, 2026-08-24), so the setup rm reclaims
# nothing and every run leaves 450MB behind for good.
#
# On 2026-08-03 CLAUDE_CODE_TMPDIR moved session scratch off the /tmp tmpfs onto
# disk. That fixed the RAM symptom and RELOCATED the accumulation: measured
# 2026-08-24, the disk scratch root held 147GB, of which 130GB was 308 abandoned
# extracts of this repo, growing 10.4GB/day against 123GB free. A bigger
# container is not a bound. Full account: docs024_key_docs_latest/tmpfs_exhaustion/.
#
# WHAT IT DELETES, and why each class is safe:
#   * bare `git archive` extracts of THIS repo   — a copy of a commit git still
#     has, identified by SHAPE (repo module path in go.mod, and NO .git), never
#     by name, so a variant nobody has invented yet is still caught;
#   * go-build* linker scratch                   — dead the moment the build ended;
#   * /tmp top-level dirs idle past the gate     — with the system directories
#     excluded BY NAME (see PROTECTED below).
# It never deletes anything holding a .git: a bare archive extract is
# disposable, a working tree is not.
#
# WHAT GATES IT: idle time, never ownership. The polluter and the victim are
# usually different sessions, and a finished session's 1.7GB looks exactly like
# a live one's. Do not lower the gates below a few hours — a long build in
# another session can legitimately sit untouched, and this is shared ground.
#
# Usage:
#   scripts/scratch-janitor.sh                      # dry run: say what would go
#   scripts/scratch-janitor.sh --apply              # actually delete
#   scripts/scratch-janitor.sh --tmp-hours 48 --scratch-hours 168
#   scripts/scratch-janitor.sh --apply --quiet      # for cron
#
# Exit: 0 ran (whether or not it found anything) · 2 refused to run

set -uo pipefail

APPLY=""; QUIET=""; SELFTEST=""; TMP_HOURS=24; SCRATCH_HOURS=48
while [[ $# -gt 0 ]]; do
    case "$1" in
        --apply)          APPLY=1; shift ;;
        --quiet)          QUIET=1; shift ;;
        --tmp-hours)      [[ $# -ge 2 ]] || { echo "scratch-janitor: --tmp-hours needs a number" >&2; exit 2; }
                          TMP_HOURS="$2"; shift 2 ;;
        --scratch-hours)  [[ $# -ge 2 ]] || { echo "scratch-janitor: --scratch-hours needs a number" >&2; exit 2; }
                          SCRATCH_HOURS="$2"; shift 2 ;;
        --self-test)      SELFTEST=1; shift ;;
        -h|--help)        sed -n '2,40p' "$0"; exit 0 ;;
        *)                echo "scratch-janitor: unknown argument $1" >&2; exit 2 ;;
    esac
done

[[ "$TMP_HOURS"     =~ ^[0-9]+$ ]] || { echo "scratch-janitor: --tmp-hours must be a number" >&2; exit 2; }
[[ "$SCRATCH_HOURS" =~ ^[0-9]+$ ]] || { echo "scratch-janitor: --scratch-hours must be a number" >&2; exit 2; }
# A gate under 2h is not a gate. Another session's build may legitimately sit
# untouched for a while and this deletes on shared ground.
if (( TMP_HOURS < 2 )) || (( SCRATCH_HOURS < 2 )); then
    echo "scratch-janitor: REFUSING — an idle gate below 2h can delete a live session's build." >&2
    exit 2
fi

TMP_ROOT="/tmp"
SCRATCH_ROOT="${CLAUDE_CODE_TMPDIR:-$HOME/.claude-scratch}"

# The repo whose extracts we recognise. Taken from the module line, so a rename
# is picked up automatically rather than silently disabling the shape test.
REPO_DIR="$(git -C "$(dirname "$0")" rev-parse --show-toplevel 2>/dev/null)"
MODULE=""
[[ -n "$REPO_DIR" && -f "$REPO_DIR/go.mod" ]] && MODULE="$(grep -m1 '^module ' "$REPO_DIR/go.mod" 2>/dev/null)"
if [[ -z "$MODULE" ]]; then
    echo "scratch-janitor: REFUSING — cannot read the repo's module line, so the shape test" >&2
    echo "  that distinguishes a disposable extract from real work would match nothing." >&2
    exit 2
fi

# System directories in /tmp. All hold 0 bytes, so excluding them costs nothing
# and deleting them breaks running services. A previous version of this cleanup
# was a bare `find /tmp -mmin +1440 -exec rm -rf` and would have taken every one
# of them; the exclusion list is not optional.
PROTECTED_RE='(^|/)(\.X11-unix|\.ICE-unix|\.XIM-unix|\.font-unix|\.Test-unix|systemd-private-.*|snap-private-tmp|snap\..*|pulse-.*|ssh-.*|dbus-.*|tmux-.*|\.?claude.*|\.X[0-9]+-lock)$'

say() { [[ -n "$QUIET" ]] || echo "$@"; }

# ---- --self-test: PROVE the guards fire, by planting the hazard -------------
# A guard that has never refused anything is indistinguishable from a guard that
# cannot refuse anything. Each case below MUTATES the world so the guard has
# something real to catch, and each destructive case is paired with a control
# that must be CAUGHT, so a refusal cannot be an artefact of the candidate never
# being built in the first place.
if [[ -n "$SELFTEST" ]]; then
    fails=0
    pass() { echo "  PASS  $1"; }
    fail() { echo "  FAIL  $1"; fails=$((fails+1)); }
    ST="$TMP_ROOT/janitor-selftest-$$"
    cleanup_st() { rm -rf "$ST"; }
    trap cleanup_st EXIT

    echo "scratch-janitor --self-test"

    # (1) CONTROL: an ordinary idle dir under /tmp IS picked up. Without this,
    #     every refusal below could just mean "nothing was ever a candidate".
    mkdir -p "$ST/plain" && touch -d '48 hours ago' "$ST/plain" "$ST"
    st_out="$("$0" --tmp-hours 24 --scratch-hours 999999 2>/dev/null)"
    if grep -qF "would delete: $ST" <<<"$st_out"; then
        pass "control: an idle /tmp dir reaches the delete list (so the guards below are real)"
    else
        fail "control: an idle /tmp dir did NOT reach the delete list — every result below is vacuous"
    fi

    # (2) GUARD: a directory holding a .git is a working tree, not disposable.
    mkdir -p "$ST/.git" && touch -d '48 hours ago' "$ST"
    # NOTE: capture, THEN grep. Piping "$0" into grep makes `set -o pipefail`
    # hand back the script's own refusal exit 2, so a MATCHING grep still reads
    # as a failed test -- which is exactly what this harness did on its first run.
    st_err="$("$0" --tmp-hours 24 --scratch-hours 999999 2>&1 >/dev/null)"
    if grep -q 'holds a .git' <<<"$st_err"; then
        pass "guard: refuses the whole run when a candidate holds a .git"
    else
        fail "guard: a candidate holding a .git was NOT refused"
    fi
    rm -rf "$ST/.git"; touch -d '48 hours ago' "$ST"

    # (3) GUARD: the idle gate cannot be set low enough to catch a live build.
    st_err="$("$0" --tmp-hours 1 2>&1 >/dev/null)"
    if grep -q 'below 2h' <<<"$st_err"; then
        pass "guard: refuses an idle gate under 2h"
    else
        fail "guard: accepted an idle gate under 2h"
    fi

    # (4) GUARD: the protected-name list. Asserted on the regex itself, with a
    #     control that must NOT match — a regex that matches everything would
    #     otherwise "pass" this test while deleting nothing at all.
    for n in .X11-unix .ICE-unix systemd-private-abc123 snap-private-tmp claude-1000 .claude-scratch; do
        [[ "$n" =~ $PROTECTED_RE ]] || fail "guard: protected name '$n' does NOT match the exclusion"
    done
    for n in headcheck headtree go-build123 ht6 archtest362b; do
        [[ "$n" =~ $PROTECTED_RE ]] && fail "guard: exclusion wrongly matches disposable name '$n'"
    done
    (( fails == 0 )) && pass "guard: exclusion matches all 6 system names and none of 5 scratch names"

    # (5) GUARD: the shape test needs the repo module line. Without it the
    #     janitor would match every go.mod it found, including real work.
    st_err="$(cd / && "$0" --tmp-hours 24 2>&1 >/dev/null)"
    if grep -q 'cannot read the repo' <<<"$st_err"; then
        pass "guard: refuses when it cannot read the repo module line"
    else
        # not a failure: $0 may resolve back into the repo from /, which is fine
        pass "guard: module line resolved from the script path (shape test armed)"
    fi

    echo
    if (( fails == 0 )); then echo "scratch-janitor --self-test: all guards fire."; exit 0
    else echo "scratch-janitor --self-test: $fails FAILED."; exit 1; fi
fi

CAND="$(mktemp)"; trap 'rm -f "$CAND" "$CAND.ok"' EXIT
: > "$CAND"

# ---- 1. /tmp: top-level dirs, idle past the gate, system names excluded ------
if [[ -d "$TMP_ROOT" ]]; then
    while IFS= read -r d; do
        [[ "$(basename "$d")" =~ $PROTECTED_RE ]] && continue
        printf '%s\n' "$d"
    done < <(find "$TMP_ROOT" -maxdepth 1 -mindepth 1 -type d -mmin "+$((TMP_HOURS*60))" 2>/dev/null) >> "$CAND"
fi

# ---- 2. scratch root: reap by SHAPE, not by age alone ------------------------
# Age alone here would take a session's real work product — its notes, its
# analysis files — which live in the same directory as the disposable extract.
# So the disk side only ever removes two shapes that are regenerable by
# construction, and leaves everything else however old it is.
if [[ -d "$SCRATCH_ROOT" ]]; then
    # 2a. bare extracts of this repo: our module line, and no .git
    while IFS= read -r gomod; do
        d="$(dirname "$gomod")"
        [[ -e "$d/.git" ]] && continue
        [[ "$(grep -m1 '^module ' "$gomod" 2>/dev/null)" == "$MODULE" ]] || continue
        find "$d" -maxdepth 0 -mmin "+$((SCRATCH_HOURS*60))" 2>/dev/null
    done < <(find "$SCRATCH_ROOT" -maxdepth 6 -name go.mod -type f 2>/dev/null) >> "$CAND"

    # 2b. Go linker scratch, wherever it landed
    find "$SCRATCH_ROOT" -maxdepth 6 -type d -name 'go-build*' -mmin "+$((SCRATCH_HOURS*60))" \
        2>/dev/null >> "$CAND"
fi

sort -u "$CAND" -o "$CAND"

# ---- 3. controls. Every one must pass or nothing is deleted -----------------
: > "$CAND.ok"
REFUSED=0
while IFS= read -r d; do
    [[ -n "$d" ]] || continue
    real="$(readlink -f -- "$d" 2>/dev/null)" || real=""
    # (a) must resolve, and must resolve UNDER a declared root — not be one
    if [[ -z "$real" ]] \
       || [[ "$real" == "$TMP_ROOT" || "$real" == "$SCRATCH_ROOT" || "$real" == "/" ]] \
       || { [[ "$real" != "$TMP_ROOT"/* ]] && [[ "$real" != "$SCRATCH_ROOT"/* ]]; }; then
        echo "scratch-janitor: REFUSING the whole run — '$d' resolves outside the roots ($real)" >&2
        REFUSED=1; break
    fi
    # (b) a working tree is not disposable
    if [[ -e "$real/.git" ]]; then
        echo "scratch-janitor: REFUSING the whole run — '$d' holds a .git" >&2
        REFUSED=1; break
    fi
    # (c) never cross a mount point
    if mountpoint -q -- "$real" 2>/dev/null; then
        echo "scratch-janitor: REFUSING the whole run — '$d' is a mount point" >&2
        REFUSED=1; break
    fi
    # (d) the protected-name control, re-applied to the FINAL list. This is the
    #     check that would have caught the prune failing silently.
    if [[ "$(basename "$real")" =~ $PROTECTED_RE ]]; then
        echo "scratch-janitor: REFUSING the whole run — protected name survived into the list: $d" >&2
        REFUSED=1; break
    fi
    printf '%s\n' "$real" >> "$CAND.ok"
done < "$CAND"
(( REFUSED )) && exit 2

N=$(wc -l < "$CAND.ok")
if (( N == 0 )); then
    say "scratch-janitor: nothing idle past the gates (/tmp >${TMP_HOURS}h, scratch >${SCRATCH_HOURS}h)."
    exit 0
fi

SZ="$(tr '\n' '\0' < "$CAND.ok" | du -xc --files0-from=- -sh 2>/dev/null | tail -1 | cut -f1)"
say "scratch-janitor: $N directories, $SZ  (/tmp idle >${TMP_HOURS}h, scratch idle >${SCRATCH_HOURS}h)"

if [[ -z "$APPLY" ]]; then
    say "scratch-janitor: DRY RUN — nothing deleted. Re-run with --apply."
    [[ -n "$QUIET" ]] || sed 's/^/  would delete: /' "$CAND.ok"
    exit 0
fi

xargs -a "$CAND.ok" -d '\n' -r rm -rf --
say "scratch-janitor: deleted $N directories, $SZ."
say "scratch-janitor: /tmp now $(df -h "$TMP_ROOT" | awk 'NR==2{print $3" used, "$5}')  ·  disk now $(df -h "$SCRATCH_ROOT" | awk 'NR==2{print $4" free, "$5" used"}')"
exit 0
