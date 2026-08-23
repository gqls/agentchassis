#!/usr/bin/env bash
# verify-head-builds.sh — does COMMITTED HEAD still compile?
#
# WHY THIS EXISTS AS A SCRIPT AND NOT AS A LINE YOU PASTE. This check was
# prescribed in NINE documents as a hand-typed command, and by 2026-08-23 the
# nine copies had drifted into six different directory names (/tmp/h, /tmp/x,
# /tmp/chk, /tmp/headcheck, /tmp/headtree, /tmp/cleantree), two different build
# targets, and — in the most-read copy, 016b's — a version that DOES NOT WORK:
# it extracts into a directory it never creates, so it fails first-use for
# anyone who follows it exactly and succeeds only for someone who happens to
# have run a different variant earlier. That is the hand-maintained-copies drift
# class this estate keeps filing bugs about; the answer here is the same as
# scripts/council-scope.sh — one implementation, N callers.
#
# WHAT IT CHECKS, and why the working tree cannot answer it. `go build ./...`,
# `go test ./...` and your editor all read the WORKING TREE, which on this
# machine is the union of every concurrent session's uncommitted work. A commit
# that references another session's untracked file compiles perfectly for you
# and breaks HEAD for everyone. `make build-<service>` builds from committed
# HEAD, so the gap is invisible exactly when the missing piece is someone
# else's. This extracts committed HEAD alone and builds THAT.
#
# RUN IT AFTER COMMITTING a file another session is also in — after, not before,
# and never against the working tree.
#
# WHERE IT WRITES, and why that is load-bearing. ~450MB per checkout. It writes
# to DISK, never to /tmp: /tmp on this box is a 16G tmpfs, i.e. RAM, and the old
# pasted recipes filled it to 100% every few days — 12GB of 15GB was this one
# check, and when it filled, the machine's swap went with it. It also points the
# Go LINKER at disk (GOTMPDIR/TMPDIR), because Go ignores CLAUDE_CODE_TMPDIR and
# otherwise puts its own scratch in /tmp regardless of where the tree is.
# Full account: docs024_key_docs_latest/tmpfs_exhaustion/.
#
# AND IT CLEANS UP AFTER ITSELF, which no pasted version did. That omission is
# why 28 abandoned checkouts were found holding 12GB.
#
# Usage:
#   scripts/verify-head-builds.sh                       # build ./...
#   scripts/verify-head-builds.sh ./platform/... ./cmd/...
#   KEEP_TREE=1 scripts/verify-head-builds.sh           # leave the tree for poking at
#   VERIFY_HEAD_DIR=/some/disk/path scripts/verify-head-builds.sh
#
# Exit: 0 HEAD builds · 1 HEAD does NOT build · 2 could not run the check

set -uo pipefail

OVERLAY=(); ARGS=(); RUN_TESTS=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        --with)  [[ $# -ge 2 ]] || { echo "verify-head-builds: --with needs a path" >&2; exit 2; }
                 OVERLAY+=("$2"); shift 2 ;;
        --test)  RUN_TESTS=1; shift ;;
        --)      shift; ARGS+=("$@"); break ;;
        -*)      echo "verify-head-builds: unknown flag $1 (want --with <path>, --test)" >&2; exit 2 ;;
        *)       ARGS+=("$1"); shift ;;
    esac
done

REPO="$(git rev-parse --show-toplevel 2>/dev/null)" || {
    echo "verify-head-builds: not inside a git repository." >&2; exit 2; }
cd "$REPO" || exit 2

BASE="${VERIFY_HEAD_DIR:-${CLAUDE_CODE_TMPDIR:-$HOME/.claude-scratch}/head-verify}"

# REFUSE TO WRITE A 450MB TREE INTO RAM. Not a nicety: this whole script exists
# because that was happening fleet-wide, and a future edit pointing BASE back at
# a tmpfs would silently recreate it. The check is on the FILESYSTEM TYPE of the
# target, not on the path spelling, so /tmp/anything and a remounted tmpfs are
# both caught.
mkdir -p "$BASE" 2>/dev/null || { echo "verify-head-builds: cannot create $BASE" >&2; exit 2; }
FSTYPE="$(findmnt -no FSTYPE --target "$BASE" 2>/dev/null || echo unknown)"
if [[ "$FSTYPE" == "tmpfs" || "$FSTYPE" == "ramfs" ]]; then
    echo "verify-head-builds: REFUSING — $BASE is on $FSTYPE, which is RAM, not disk." >&2
    echo "  A checkout is ~450MB and this is how /tmp reached 100% fleet-wide (and took" >&2
    echo "  the machine's swap with it). Set VERIFY_HEAD_DIR to a path on real storage." >&2
    exit 2
fi

TREE="$BASE/$$/tree"
# Go's scratch is a SIBLING of the tree, never the tree itself or a parent of it.
# Setting TMPDIR to the checkout makes Go consider the module to be inside the
# system temp root and it then IGNORES its go.mod entirely -- "pattern ./x/...:
# directory prefix does not contain main module". That failure reads exactly
# like a broken HEAD, which is the one wrong answer this script must never give.
GOSCRATCH="$BASE/$$/gotmp"
cleanup() { [[ -n "${KEEP_TREE:-}" ]] || rm -rf "$BASE/$$"; }
trap cleanup EXIT INT TERM

rm -rf "$BASE/$$" && mkdir -p "$TREE" "$GOSCRATCH" || { echo "verify-head-builds: cannot prepare $TREE" >&2; exit 2; }

SHA="$(git rev-parse --short HEAD)"
if ! git archive HEAD | tar -x -C "$TREE"; then
    echo "verify-head-builds: could not extract HEAD ($SHA) into $TREE" >&2
    exit 2
fi

for f in "${OVERLAY[@]}"; do
    if [[ ! -f "$REPO/$f" ]]; then
        echo "verify-head-builds: --with $f: no such file in the working tree" >&2; exit 2
    fi
    mkdir -p "$TREE/$(dirname "$f")"
    cp "$REPO/$f" "$TREE/$f" || { echo "verify-head-builds: could not overlay $f" >&2; exit 2; }
done
[[ ${#OVERLAY[@]} -gt 0 ]] && echo "verify-head-builds: overlaid ${#OVERLAY[@]} working-tree file(s) onto HEAD $SHA"

TARGETS=("${ARGS[@]}"); [[ ${#TARGETS[@]} -eq 0 ]] && TARGETS=("./...")

if [[ -n "$RUN_TESTS" ]]; then
    echo "verify-head-builds: testing COMMITTED HEAD $SHA (${TARGETS[*]}) in $TREE"
    if (cd "$TREE" && GOTMPDIR="$GOSCRATCH" TMPDIR="$GOSCRATCH" go test "${TARGETS[@]}" -count=1); then
        echo "verify-head-builds: OK — tests pass against HEAD $SHA."
        exit 0
    fi
    echo "verify-head-builds: FAILED — tests do NOT pass against HEAD $SHA." >&2
    echo "  Re-run with KEEP_TREE=1 to inspect $TREE." >&2
    exit 1
fi

echo "verify-head-builds: building COMMITTED HEAD $SHA (${TARGETS[*]}) in $TREE"
# TMPDIR as well as GOTMPDIR: older toolchains read only TMPDIR, and the linker
# is the part that needs gigabytes.
if (cd "$TREE" && GOTMPDIR="$GOSCRATCH" TMPDIR="$GOSCRATCH" go build "${TARGETS[@]}"); then
    echo "verify-head-builds: OK — HEAD $SHA builds."
    exit 0
fi

echo "verify-head-builds: FAILED — HEAD $SHA does NOT build, though your working tree may." >&2
echo "  The usual cause is a commit referencing a file another session has not committed yet." >&2
echo "  Re-run with KEEP_TREE=1 to inspect $TREE." >&2
exit 1
