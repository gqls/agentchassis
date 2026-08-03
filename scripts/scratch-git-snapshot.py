#!/usr/bin/env python3
"""scratch-git-snapshot.py — PostToolUse hook. Versions every file Claude
writes into a session scratchpad, so a scratchpad stops being unrecoverable.

WHY THIS EXISTS (2026-08-03):
/tmp is a 16 GB tmpfs and it hit 100% full, which surfaced as a Bash call
returning NO OUTPUT (ENOSPC on the harness's own stdout capture) rather than as
a disk error — a failure that reads like "the command found nothing".

The obvious fix is to delete old scratchpads. That is exactly what could not be
done safely: a dozen sessions share the tree, directory mtime lies (the largest
dir showed yesterday at the top level and today two levels down), and a
scratchpad had NO safety net — unlike the repo, where a wrong `rm` costs
nothing because git has a copy. Same argument, same shape, same remedy as
`memory-git-snapshot.py` (see CLAUDE.md on the 2026-07-20 memory clobber).

WHAT IT DOES AND DOES NOT CAPTURE — this is the load-bearing design decision,
and it comes from a measurement, not a preference. Of 12.2 GB of scratchpads:

    repo-extraction dirs   33    12.13 GB   99.3%
    everything else       370     0.09 GB    0.7%   (88 MB)

The 99.3% is `git archive HEAD` extractions — the shared-tree build check
CLAUDE.md mandates. Those are byte-for-byte reproducible from a sha and must
NEVER enter git: committing them would permanently store a duplicate repo per
check. The 0.7% is the irreplaceable part — hand-written scripts, SQL, notes,
submissions — and 88 MB is small enough that it never needs deleting at all.

The split falls out of the hook's scope for free: extractions are created by
`tar`/`git archive` in **Bash**, while this hook fires on **Write|Edit**. So it
sees exactly the hand-authored files and is structurally incapable of
committing an extraction. That is why there is no size heuristic doing the
separating — a heuristic can misfire, a tool boundary cannot.

DESIGN NOTES:
- `.gitignore` is `*` and every add is `-f`. So the repo contains ONLY what was
  deliberately snapshotted; `git status` stays clean and fast next to 12 GB of
  untracked bulk, instead of listing it. Untracked bulk is `scratch-report.py`'s
  job to show, not git's.
- Commits ONE file by pathspec, never `-A`. A dozen sessions write here
  concurrently; sweeping another session's half-finished file into this commit
  is the exact failure CLAUDE.md's commit-per-task rule exists to prevent.
- **Retries on index.lock.** The memory hook does not need this (memory writes
  are rare); scratch writes are frequent and concurrent, and a lock collision
  here would silently skip the snapshot — leaving a file the operator believes
  is protected and is not. A safety net with silent holes is worse than none,
  because it is trusted.
- Silent and non-blocking, always exit 0. A snapshot tool that can break the
  session it protects is worse than no snapshot tool.
- Skips files over MAX_BYTES so a generated artefact cannot bloat the history.
  This is a backstop, not the extraction filter — see above.
"""
import json
import os
import re
import subprocess
import sys
import time

# <tmpdir>/claude-<uid>/<project-slug>/<session-uuid>/scratchpad/...
# Anchored on the whole layout so it matches under ANY CLAUDE_CODE_TMPDIR — the
# new on-disk root and the legacy /tmp one, which sessions straddle during the
# move (a running session keeps the tmpdir it started with).
SCRATCH_RE = re.compile(
    r"^(?P<root>.*/claude-\d+)/(?P<proj>[^/]+)/(?P<sess>[0-9a-fA-F-]{36})/scratchpad/"
)

MAX_BYTES = 5 * 1024 * 1024
LOCK_TRIES = 5


def run(args, cwd=None):
    return subprocess.run(args, cwd=cwd, capture_output=True, text=True, timeout=20)


def git(root, *args):
    """Run git in the scratch repo, retrying while another session holds the lock."""
    for attempt in range(LOCK_TRIES):
        r = run(["git", "-C", root, *args])
        if r.returncode == 0:
            return r
        blob = (r.stderr or "") + (r.stdout or "")
        if "index.lock" not in blob and "Unable to create" not in blob:
            return r
        # Exponential-ish backoff, jittered by pid so concurrent sessions do not
        # retry in lockstep and collide again on the same tick.
        time.sleep(0.12 * (attempt + 1) + (os.getpid() % 7) / 100.0)
    return r


def ensure_repo(root):
    if os.path.isdir(os.path.join(root, ".git")):
        return True
    if run(["git", "init", "-q", root]).returncode != 0:
        return False
    run(["git", "-C", root, "config", "user.name", "claude-scratch"])
    run(["git", "-C", root, "config", "user.email", "noreply@anthropic.com"])
    # Ignore everything; the hook force-adds what it means to keep. Without this
    # the repo would try to track 12 GB of reproducible extractions.
    gi = os.path.join(root, ".gitignore")
    if not os.path.exists(gi):
        with open(gi, "w") as fh:
            fh.write(
                "# Everything is ignored by default.\n"
                "# scratch-git-snapshot.py force-adds (`git add -f`) the individual\n"
                "# files Claude writes via Write/Edit. Repo extractions, build output\n"
                "# and other reproducible bulk are created by Bash and stay untracked\n"
                "# on purpose — see scripts/scratch-report.py to see and reap them.\n"
                "*\n"
            )
    run(["git", "-C", root, "add", "-f", "--", ".gitignore"])
    run(["git", "-C", root, "commit", "-q", "-m",
         "baseline: scratch history begins (everything ignored; snapshots are force-added)"])
    return True


def main():
    try:
        payload = json.load(sys.stdin)
    except Exception:
        return 0

    ti = payload.get("tool_input") or {}
    tr = payload.get("tool_response") or {}
    path = tr.get("filePath") or ti.get("file_path") or ""
    if not path:
        return 0

    m = SCRATCH_RE.match(path)
    if not m or not os.path.isfile(path):
        return 0

    try:
        if os.path.getsize(path) > MAX_BYTES:
            return 0
    except OSError:
        return 0

    root = m.group("root")
    if not os.path.isdir(root) or not ensure_repo(root):
        return 0

    rel = os.path.relpath(path, root)
    if git(root, "add", "-f", "--", rel).returncode != 0:
        return 0
    # Nothing staged for this pathspec means the content did not change.
    if run(["git", "-C", root, "diff", "--cached", "--quiet", "--", rel]).returncode == 0:
        return 0

    sess = m.group("sess")[:8]
    tool = (payload.get("tool_name") or "write").lower()
    git(root, "commit", "-q", "-m", f"{tool}: {rel} (session {sess})", "--", rel)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:
        sys.exit(0)  # never break the session this is meant to protect
