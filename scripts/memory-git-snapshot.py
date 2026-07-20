#!/usr/bin/env python3
"""memory_git_snapshot.py — PostToolUse hook. Commits every write to the
auto-memory directory into a git repo living in that directory.

WHY THIS EXISTS (2026-07-20):
A session overwrote another session's memory file with `cat >` where a merge was
needed. The auto-memory directory is not under version control, so the original
was unrecoverable — it had to be reconstructed from a surviving index line and
the workstream's repo docs, and the owning thread then had to restore the rest
by hand.

The instinct was to guard the redirect: block `cat >` onto files you have not
read. That is the wrong shape. The Write tool ALREADY refuses to overwrite an
unread file — the hole is that `cat >` in Bash does not, and a guard would have
to enumerate every way a shell can clobber a file (`>`, `tee`, `cp`, `mv`,
`sed -i`, `python -c`...). Meanwhile the same mistake made in the repo would have
cost nothing, because git had a copy.

So the real difference was never the syntax — it was that one location was
versioned and the other was not. This levels that up rather than policing
writes: memory becomes recoverable, and the whole class dissolves. Multiple
concurrent sessions share this directory (see CLAUDE.md on multi-session
coordination), which is exactly the condition that makes an unversioned shared
store dangerous.

DESIGN NOTES:
- Matches ANY path under `~/.claude/projects/<project>/memory/`, not just this
  project's, so it protects every auto-memory store on the machine.
- The git repo lives INSIDE the memory directory. That keeps it self-contained
  and means it cannot interfere with the project repo.
- Silent and non-blocking. It always exits 0, even on failure: a snapshot tool
  that can break the session it protects is worse than none.
- Commits only the one file that was written, by pathspec. Another session may
  be mid-write in the same directory; sweeping its half-finished file into this
  commit would repeat, in miniature, the exact problem CLAUDE.md's commit-per-
  task rule exists to prevent.
"""
import json
import os
import re
import subprocess
import sys

MEMORY_RE = re.compile(r"/\.claude/projects/[^/]+/memory(/|$)")


def run(args, cwd):
    return subprocess.run(args, cwd=cwd, capture_output=True, text=True, timeout=15)


def main():
    try:
        payload = json.load(sys.stdin)
    except Exception:
        return 0

    ti = payload.get("tool_input") or {}
    tr = payload.get("tool_response") or {}
    path = tr.get("filePath") or ti.get("file_path") or ""
    if not path or not MEMORY_RE.search(path) or not os.path.isfile(path):
        return 0

    mem_dir = path[: path.index("/memory/") + len("/memory")] if "/memory/" in path else os.path.dirname(path)
    if not os.path.isdir(mem_dir):
        return 0

    if not os.path.isdir(os.path.join(mem_dir, ".git")):
        if run(["git", "init", "-q"], mem_dir).returncode != 0:
            return 0
        run(["git", "config", "user.name", "claude-auto-memory"], mem_dir)
        run(["git", "config", "user.email", "noreply@anthropic.com"], mem_dir)
        # Snapshot whatever is already there, so the first real edit has a
        # baseline to diff against instead of appearing to create the world.
        run(["git", "add", "-A"], mem_dir)
        run(["git", "commit", "-q", "-m", "baseline: memory as found when snapshotting began"], mem_dir)

    rel = os.path.relpath(path, mem_dir)
    run(["git", "add", "--", rel], mem_dir)
    # --  nothing staged for this pathspec means the content did not change.
    if run(["git", "diff", "--cached", "--quiet", "--", rel], mem_dir).returncode == 0:
        return 0

    session = (payload.get("session_id") or "unknown")[:8]
    tool = payload.get("tool_name") or "write"
    run(["git", "commit", "-q", "-m", f"{tool.lower()}: {rel} (session {session})", "--", rel], mem_dir)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:
        sys.exit(0)  # never break the session this is meant to protect
