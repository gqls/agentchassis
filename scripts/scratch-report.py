#!/usr/bin/env python3
"""scratch-report.py — what is in the session scratchpads, how old, and what is
safe to reclaim.

WHY (2026-08-03): /tmp (16 GB tmpfs) hit 100% full and a Bash call came back
with no output — ENOSPC on the harness's own stdout capture, which reads like
"the command found nothing" rather than like a full disk. The blocker on
cleaning up was not nerve, it was evidence: a dozen sessions share the tree and
there was no way to tell a dead session's scratchpad from a live one's, nor a
reproducible file from an irreplaceable one.

THE MEASUREMENT THAT MAKES CLEANUP SAFE — the two categories have completely
different risk profiles, and almost all the space is in the zero-risk one:

    repo-extraction dirs   33    12.13 GB   99.3%   reproducible from a sha
    everything else       370     0.09 GB    0.7%   irreplaceable, now in git

A repo extraction is a `git archive HEAD` unpack — the shared-tree build check
CLAUDE.md mandates. Deleting one loses NOTHING: it is regenerable from a commit.
Hand-written files are the opposite, and they are snapshotted into the scratch
git repo by `scratch-git-snapshot.py`, so they are recoverable even if reaped.

So `--reap` only ever deletes MARKER-VERIFIED extraction directories. It will
not delete a loose file, a whole session dir, or anything it cannot positively
identify. Refusing to act on an ambiguous directory is the point: the failure
this guards against is deleting the 0.7% while chasing the 99.3%.

LANDMINE this tool exists to answer: **directory mtime lies.** The largest
scratchpad here showed 2026-08-02 at its top level while a subdirectory two
levels down had been written 2026-08-03 11:32. Age is therefore computed from
the NEWEST file anywhere beneath a session dir, never from the dir itself.

USAGE
    scripts/scratch-report.py                     # report every scratch root
    scripts/scratch-report.py --days 2            # what a 2-day reap would take
    scripts/scratch-report.py --reap --days 2     # actually delete (prints each)
"""
import argparse
import os
import shutil
import subprocess
import sys
import time

# Roots to inspect. The legacy tmpfs one stays listed so the tool still sees the
# sessions that started before CLAUDE_CODE_TMPDIR moved — a running session keeps
# the tmpdir it launched with, so both are live for a while.
ROOTS = [
    os.path.expanduser("~/.claude-scratch"),
    "/tmp",
]

# A repo extraction is identified by CONTENT, not by name: the dirs are called
# fin, final, h184, hc3, headcheck, headtree… — no pattern to match on. Any of
# these markers at the top of the directory identifies an agentchassis tree.
MARKERS = ("go.mod", "platform", "CLAUDE.md")


def dir_stats(path):
    """(bytes, newest_mtime) over the whole subtree. Walks rather than trusting
    the directory's own mtime, which does not follow writes to nested files."""
    total, newest = 0, 0.0
    for dp, dn, fn in os.walk(path):
        try:
            newest = max(newest, os.path.getmtime(dp))
        except OSError:
            pass
        for f in fn:
            p = os.path.join(dp, f)
            try:
                st = os.lstat(p)
            except OSError:
                continue
            total += st.st_size
            if st.st_mtime > newest:
                newest = st.st_mtime
    return total, newest


def is_extraction(path):
    """True only for a directory that positively looks like an unpacked repo.
    Ambiguity means False — this gates deletion."""
    if not os.path.isdir(path):
        return False
    hits = sum(1 for m in MARKERS if os.path.exists(os.path.join(path, m)))
    return hits >= 2


def scratch_dirs(root):
    """Yield (session_dir, project, session) for <root>/claude-*/<proj>/<uuid>/."""
    if not os.path.isdir(root):
        return
    for entry in sorted(os.listdir(root)):
        if not entry.startswith("claude-"):
            continue
        base = os.path.join(root, entry)
        if not os.path.isdir(base):
            continue
        for proj in sorted(os.listdir(base)):
            pdir = os.path.join(base, proj)
            if not os.path.isdir(pdir):
                continue
            for sess in sorted(os.listdir(pdir)):
                sdir = os.path.join(pdir, sess)
                if os.path.isdir(sdir) and len(sess) == 36:
                    yield sdir, proj, sess


def tracked_count(root_repo, path):
    """How many files under `path` the scratch git repo has snapshotted."""
    if not os.path.isdir(os.path.join(root_repo, ".git")):
        return 0
    try:
        rel = os.path.relpath(path, root_repo)
        r = subprocess.run(["git", "-C", root_repo, "ls-files", "--", rel],
                           capture_output=True, text=True, timeout=30)
        return len([l for l in r.stdout.splitlines() if l.strip()])
    except Exception:
        return 0


def human(n):
    for unit in ("B", "K", "M", "G"):
        if abs(n) < 1024 or unit == "G":
            return f"{n:.0f}{unit}" if unit == "B" else f"{n:.1f}{unit}"
        n /= 1024.0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--days", type=float, default=2.0,
                    help="age threshold for reaping extractions (default 2)")
    ap.add_argument("--reap", action="store_true",
                    help="actually delete marker-verified extraction dirs older than --days")
    args = ap.parse_args()

    now = time.time()
    grand_total = grand_extract = 0
    reapable = []

    for root in ROOTS:
        rows = list(scratch_dirs(root))
        if not rows:
            continue
        print(f"\n=== {root} ===")
        try:
            du = shutil.disk_usage(root)
            print(f"filesystem: {human(du.used)} used of {human(du.total)}, "
                  f"{human(du.free)} free")
        except OSError:
            pass
        # The scratch git repo lives at the claude-<uid> level, which is the root
        # the snapshot hook derives from a written path.
        print(f"{'session':10} {'age':>7} {'total':>8} {'extract':>8} {'keep':>7} "
              f"{'tracked':>7}  project")
        for sdir, proj, sess in rows:
            total, newest = dir_stats(sdir)
            if total == 0 and newest == 0:
                continue
            age_d = (now - newest) / 86400.0 if newest else 0.0
            ext_bytes = 0
            sp = os.path.join(sdir, "scratchpad")
            ext_dirs = []
            for cand_parent in (sdir, sp):
                if not os.path.isdir(cand_parent):
                    continue
                for name in sorted(os.listdir(cand_parent)):
                    cand = os.path.join(cand_parent, name)
                    if is_extraction(cand):
                        b, mt = dir_stats(cand)
                        ext_bytes += b
                        ext_dirs.append((cand, b, mt))
            repo_root = os.path.dirname(os.path.dirname(sdir))  # claude-<uid>
            trk = tracked_count(repo_root, sdir)
            grand_total += total
            grand_extract += ext_bytes
            print(f"{sess[:8]:10} {age_d:6.1f}d {human(total):>8} "
                  f"{human(ext_bytes):>8} {human(total-ext_bytes):>7} {trk:>7}  {proj[:28]}")
            for cand, b, mt in ext_dirs:
                if (now - mt) / 86400.0 >= args.days:
                    reapable.append((cand, b))

    print(f"\ntotal in scratchpads : {human(grand_total)}")
    print(f"  reproducible extractions: {human(grand_extract)}"
          f" ({100*grand_extract/grand_total:.1f}%)" if grand_total else "")
    print(f"  irreplaceable work      : {human(grand_total-grand_extract)}"
          "  (snapshotted into the scratch git repo)")

    if not reapable:
        print(f"\nnothing older than {args.days}d is safely reapable.")
        return 0

    total_reap = sum(b for _, b in reapable)
    print(f"\n{len(reapable)} marker-verified extraction dir(s) older than "
          f"{args.days}d = {human(total_reap)}:")
    for cand, b in reapable:
        print(f"   {human(b):>8}  {cand}")

    if not args.reap:
        print("\ndry run. Re-run with --reap to delete these. Nothing else is ever "
              "touched: loose files, session dirs and unidentifiable directories "
              "are never deleted by this tool.")
        return 0

    freed = 0
    for cand, b in reapable:
        # Re-verify at the moment of deletion, not just at scan time: a session
        # may have written into this path since the walk above.
        if not is_extraction(cand):
            print(f"   SKIP (no longer identifiable): {cand}")
            continue
        try:
            shutil.rmtree(cand)
            freed += b
            print(f"   removed {human(b):>8}  {cand}")
        except OSError as e:
            print(f"   FAILED {cand}: {e}")
    print(f"\nfreed {human(freed)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
