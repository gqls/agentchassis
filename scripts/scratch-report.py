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
import re
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

# System directories in /tmp. They hold 0 bytes, so excluding them costs nothing
# and deleting them breaks running services (.X11-unix, the 11 systemd-private-*,
# snap-private-tmp). Idle time is the right gate for SCRATCH; it is not a filter
# for WHAT a directory is.
PROTECTED_RE = re.compile(
    r"^(\.X11-unix|\.ICE-unix|\.XIM-unix|\.font-unix|\.Test-unix|systemd-private-.*"
    r"|snap-private-tmp|snap\..*|pulse-.*|ssh-.*|dbus-.*|tmux-.*|\.?claude.*|\.X[0-9]+-lock)$")

# Go's LINKER work dirs. Dead the moment the build ended, and invisible to the
# marker test because they contain no repo files at all. Go ignores
# CLAUDE_CODE_TMPDIR, so before GOTMPDIR was set these landed in /tmp — i.e. RAM
# — and were 3.1 GB of the 15.3 GB that filled it on 2026-08-23.
GOBUILD_RE = re.compile(r"^go-build[0-9]+$")


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


def loose_reapables(root, now, days):
    """Yield (path, bytes) for reapable scratch sitting at the TOP LEVEL of a
    root, i.e. NOT inside a <root>/claude-*/<proj>/<uuid>/ session directory.

    WHY THIS EXISTS (2026-08-24). `ROOTS` has always listed /tmp, and the OPP-005
    register entry says "both tools read BOTH roots ... a check that inspects
    only one will be confidently wrong". That was FALSE for this tool: every
    candidate came from scratch_dirs(), which requires the claude-*/<proj>/<uuid>
    layout, and /tmp has never had a claude-* directory in it — measured
    2026-08-24, `ls -d /tmp/claude-*` finds nothing and the report prints no
    "=== /tmp ===" section at all. So the /tmp arm was inert while looking
    covered, for three weeks, which is the worst of both.

    Two shapes only, both regenerable by construction: a marker-verified repo
    extraction, and a Go linker work dir. Anything else at the top of a root is
    left alone however old it is -- same rule as the session side.
    """
    if not os.path.isdir(root):
        return

    def candidate(path, name):
        if PROTECTED_RE.match(name):
            return False
        if not os.path.isdir(path) or os.path.islink(path):
            return False
        if os.path.exists(os.path.join(path, ".git")):
            return False      # a working tree is not disposable
        return bool(is_extraction(path) or GOBUILD_RE.match(name))

    for name in sorted(os.listdir(root)):
        path = os.path.join(root, name)
        if PROTECTED_RE.match(name) or not os.path.isdir(path) or os.path.islink(path):
            continue
        if name.startswith("claude-"):
            continue          # session layout: scratch_dirs() owns that arm
        if candidate(path, name):
            b, mt = dir_stats(path)
            if mt and (now - mt) / 86400.0 >= days:
                yield path, b
            continue
        # One level down, and no further. A holding directory such as
        # ~/.claude-scratch/gotmp or .../adhoc is not itself scratch, but the
        # linker dirs and extracts sit directly inside it. Descending further
        # would start walking the insides of real work.
        try:
            children = sorted(os.listdir(path))
        except OSError:
            continue
        for cname in children:
            cpath = os.path.join(path, cname)
            if candidate(cpath, cname):
                b, mt = dir_stats(cpath)
                if mt and (now - mt) / 86400.0 >= days:
                    yield cpath, b


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


def self_test():
    """PROVE the deletion guards refuse, by planting the hazard.

    A guard that has never refused anything is indistinguishable from a guard
    that CANNOT refuse anything, so each destructive case below is paired with a
    control that must be CAUGHT -- otherwise "refused" could just mean the
    candidate was never built in the first place.
    """
    import tempfile
    fails = []

    def ok(msg):
        print(f"  PASS  {msg}")

    def bad(msg):
        print(f"  FAIL  {msg}")
        fails.append(msg)

    print("scratch-report --self-test")
    now = time.time()
    old = now - 10 * 86400

    with tempfile.TemporaryDirectory(prefix="scratch-selftest-") as td:
        def mk(name, markers=(), git=False):
            d = os.path.join(td, name)
            os.makedirs(d, exist_ok=True)
            for mkr in markers:
                open(os.path.join(d, mkr), "w").close()
            if git:
                os.makedirs(os.path.join(d, ".git"), exist_ok=True)
            for dp, _, fn in os.walk(d):
                os.utime(dp, (old, old))
                for f in fn:
                    os.utime(os.path.join(dp, f), (old, old))
            os.utime(d, (old, old))
            return d

        # CONTROL: a genuine extraction IS found. Without this every refusal
        # below is vacuous.
        mk("realextract", markers=("go.mod", "CLAUDE.md"))
        found = {os.path.basename(p) for p, _ in loose_reapables(td, now, 2.0)}
        if "realextract" in found:
            ok("control: a marker-verified extraction reaches the reap list")
        else:
            bad("control: a real extraction was NOT found -- every result below is vacuous")

        # GUARD: a working tree is not disposable, however old.
        mk("worktree", markers=("go.mod", "CLAUDE.md"), git=True)
        found = {os.path.basename(p) for p, _ in loose_reapables(td, now, 2.0)}
        if "worktree" in found:
            bad("guard: a directory holding .git was offered for deletion")
        else:
            ok("guard: refuses a directory holding a .git")

        # GUARD: one marker is not enough -- ambiguity must mean keep.
        mk("onemarker", markers=("go.mod",))
        found = {os.path.basename(p) for p, _ in loose_reapables(td, now, 2.0)}
        if "onemarker" in found:
            bad("guard: a single marker was treated as a positive identification")
        else:
            ok("guard: requires >=2 markers, so an ambiguous dir is kept")

        # GUARD: real work with no markers is never touched.
        mk("realwork", markers=("NOTES.md", "analysis.tsv"))
        found = {os.path.basename(p) for p, _ in loose_reapables(td, now, 2.0)}
        if "realwork" in found:
            bad("guard: unidentifiable real work was offered for deletion")
        else:
            ok("guard: leaves unidentifiable directories alone however old")

        # GUARD: the age gate. The same real extraction, freshly touched.
        d = os.path.join(td, "realextract")
        os.utime(d, (now, now))
        for dp, _, fn in os.walk(d):
            os.utime(dp, (now, now))
            for f in fn:
                os.utime(os.path.join(dp, f), (now, now))
        found = {os.path.basename(p) for p, _ in loose_reapables(td, now, 2.0)}
        if "realextract" in found:
            bad("guard: a freshly-touched extraction passed the age gate")
        else:
            ok("guard: age gate holds back a fresh extraction (same dir as the control)")

    # GUARD: the protected-name list, with a control that must NOT match --
    # an exclusion matching everything would "pass" while deleting nothing.
    for n in (".X11-unix", ".ICE-unix", "systemd-private-abc", "snap-private-tmp",
              "claude-1000", ".claude-scratch"):
        if not PROTECTED_RE.match(n):
            bad(f"guard: protected name {n!r} does NOT match the exclusion")
    for n in ("headcheck", "headtree", "go-build123", "ht6", "final2"):
        if PROTECTED_RE.match(n):
            bad(f"guard: exclusion wrongly matches disposable name {n!r}")
    if not fails:
        ok("guard: exclusion matches all 6 system names and none of 5 scratch names")

    print()
    if fails:
        print(f"scratch-report --self-test: {len(fails)} FAILED.")
        return 1
    print("scratch-report --self-test: all guards fire.")
    return 0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--days", type=float, default=2.0,
                    help="age threshold for reaping extractions (default 2)")
    ap.add_argument("--reap", action="store_true",
                    help="actually delete marker-verified extraction dirs older than --days")
    ap.add_argument("--self-test", action="store_true", dest="self_test",
                    help="plant each hazard and assert the guard refuses it")
    ap.add_argument("--summary", action="store_true",
                    help="totals and a one-line verdict only, timestamped -- for a cron log. "
                         "Suppresses the ~750-row per-session table and the per-directory list, "
                         "which are what make an appended daily log unreadable.")
    args = ap.parse_args()

    if args.self_test:
        return self_test()

    now = time.time()
    if args.summary:
        stamp = time.strftime("%Y-%m-%d %H:%M:%S %Z", time.localtime(now))
        try:
            du = shutil.disk_usage("/")
            free = f"{human(du.free)} free on / ({100*du.used//du.total}% used)"
        except OSError:
            free = "disk usage unavailable"
        print(f"\n===== {stamp} · {free} · gate {args.days}d · "
              f"{'REAPING' if args.reap else 'dry run'} =====")
    grand_total = grand_extract = 0
    reapable = []

    for root in ROOTS:
        rows = list(scratch_dirs(root))
        loose = list(loose_reapables(root, now, args.days))
        # EVERY root gets a header, even an empty one. Until 2026-08-24 this was
        # `if not rows: continue`, and the /tmp arm was structurally incapable of
        # producing rows -- so the ONLY evidence of a broken root was an ABSENT
        # header, which is indistinguishable from a root that is simply clean.
        # A missing row is the hardest refutation to notice; so never emit one.
        print(f"\n=== {root} ===")
        if not os.path.isdir(root):
            print("  root does not exist on this machine.")
            continue
        if not rows and not loose:
            print(f"  nothing reapable, and no session directories here "
                  f"(gate {args.days}d).")
            continue
        try:
            du = shutil.disk_usage(root)
            print(f"filesystem: {human(du.used)} used of {human(du.total)}, "
                  f"{human(du.free)} free")
        except OSError:
            pass
        # The scratch git repo lives at the claude-<uid> level, which is the root
        # the snapshot hook derives from a written path.
        if not args.summary:
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
            if not args.summary:
                print(f"{sess[:8]:10} {age_d:6.1f}d {human(total):>8} "
                      f"{human(ext_bytes):>8} {human(total-ext_bytes):>7} {trk:>7}  {proj[:28]}")
            for cand, b, mt in ext_dirs:
                if (now - mt) / 86400.0 >= args.days:
                    reapable.append((cand, b))

        if loose:
            loose_bytes = sum(b for _, b in loose)
            grand_total += loose_bytes
            grand_extract += loose_bytes
            print(f"  + {len(loose)} loose extraction/linker dir(s) outside any "
                  f"session dir = {human(loose_bytes)}")
            reapable.extend(loose)

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
    if not args.summary:
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
