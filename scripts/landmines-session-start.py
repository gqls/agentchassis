#!/usr/bin/env python3
"""SessionStart hook — surface landmines for the files this session is starting on.

D10(d), owner ruling 2026-07-29. The proposal called delivery-to-a-session "the
weak link" and recommended AGAINST relying on discipline alone; this is the
mechanism that stops it being alone.

    echo '{}' | ./scripts/landmines-session-start.py     # what a session would see

WHAT IT MATCHES ON. At SessionStart there is no session diff yet, so it uses the
WORKING TREE — the files currently in flight. On this shared checkout that is
usually the honest answer to "what is this session about to touch", because a
thread starts by picking up whatever is dirty.

WHY IT READS THE MARKDOWN AND NOT doc_notes. The markdown is the system of record
(D10(b)), it is on disk, and it costs no cluster round trip — this runs at every
session start and must never be the reason a session is slow, or fails when the
kubeconfig has expired (which happens every 3 days here).

IT MUST NEVER BREAK A SESSION. Any failure exits 0 with no output. A hook that
can stop you working would be a worse landmine than any it reports.
"""

import json
import os
import subprocess
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

MAX_ENTRIES = 6          # a wall of text at startup gets skimmed, then ignored
MAX_PATHS = 400          # this tree is routinely dirty with other sessions' work


def changed_paths():
    """Working-tree paths, from git. Empty list on any failure."""
    try:
        out = subprocess.run(
            ["git", "-C", os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
             "status", "--porcelain"],
            capture_output=True, text=True, timeout=10,
        )
        if out.returncode != 0:
            return []
        paths = []
        for line in out.stdout.splitlines()[:MAX_PATHS]:
            p = line[3:].strip()
            if " -> " in p:            # renames: take the destination
                p = p.split(" -> ", 1)[1]
            if p:
                paths.append(p.strip('"'))
        return paths
    except Exception:
        return []


def matches(entries, paths):
    """Landmines whose footprint intersects a changed path.

    Substring both ways: a footprint may be a directory ('cmd/') that a path sits
    under, or a full path that equals it. Bare table/command footprints
    ('agent_definitions', 'git commit') will not match a path and are deliberately
    NOT reported here — a session start is about files. Those are found by grep,
    which is what the standing discipline in CLAUDE.md is for.
    """
    hits = []
    for e in entries:
        for fp in e["footprints"]:
            if "/" not in fp and not fp.endswith(".go") and not fp.endswith(".py"):
                continue                      # not a path-shaped footprint
            for p in paths:
                if fp in p or p in fp:
                    hits.append((e, fp, p))
                    break
            else:
                continue
            break
    return hits


def main():
    try:
        entries = None
        import landmines_lib
        entries = landmines_lib.parse()
    except Exception:
        return 0                              # never break a session

    paths = changed_paths()
    if not entries or not paths:
        return 0

    hits = matches(entries, paths)
    if not hits:
        return 0

    shown = hits[:MAX_ENTRIES]
    lines = [
        "LANDMINES matching files already dirty in this working tree "
        f"({len(hits)} matched{', showing ' + str(MAX_ENTRIES) if len(hits) > MAX_ENTRIES else ''}). "
        "These are traps that mislead you BEFORE you have a symptom — read the "
        "check before touching the file, not after something looks wrong:",
        "",
    ]
    for e, fp, p in shown:
        lines.append(f"- {e['title']}")
        lines.append(f"    footprint `{fp}` matched `{p}`")
        if e["the_check"]:
            lines.append(f"    the check: {e['the_check']}")
    lines.append("")
    lines.append(
        "Full file: docs/agent_docs/docs024_key_docs_latest/LANDMINES.md — "
        "grep it by path/table/symbol before touching anything unfamiliar."
    )

    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "SessionStart",
            "additionalContext": "\n".join(lines),
        },
        "suppressOutput": True,
    }))
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:
        sys.exit(0)
