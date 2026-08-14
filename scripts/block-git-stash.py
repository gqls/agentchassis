#!/usr/bin/env python3
"""PreToolUse hook: refuse mutating `git stash` on this shared working tree.

WHY (2026-08-12, measured): one bare `git stash` reverts EVERY dirty tracked
file to HEAD. That day it swept 38 files of ~10 sessions' uncommitted work and
put 18 production overlay manifests back ~50-100 releases — after which the
tree read clean and matched HEAD, so the next `kubectl apply -k` would have
been a silent fleet rollback. Git has no pre-stash hook, so the refusal lives
here, at the harness. CLAUDE.md § Git carries the rule; LANDMINES.md carries
the full trap and the recovery.

BLOCKED: every stash form that mutates the tree or the stash ref — bare
`git stash`, push, save, pop, apply, drop, clear, branch, create, store.
ALLOWED: `git stash list` and `git stash show` (read-only; the documented
recovery reads a stash another session left). `git show 'stash@{N}:<path>'`
is `git show`, not `git stash`, and is untouched.

Matching is per simple-command segment (split on |, ;, &, newlines), skipping
git's global options (-C <path>, -c k=v, --git-dir[=…], …), so
`cd x && git -C /repo stash` is caught while `git log -g refs/stash` and
`grep 'git stash' LANDMINES.md` are not.

Fail-open on malformed input — a crashed gate must surface, not block; the
decision paths themselves are deterministic. Test: --self-test (12 cases).
"""

import json
import re
import sys

READ_ONLY = {"list", "show"}
# git global options that consume the FOLLOWING token as their argument
ARG_OPTS = {"-C", "-c", "--git-dir", "--work-tree", "--namespace", "--exec-path"}


def offending_segment(command):
    """First simple-command segment that runs a mutating stash, else None."""
    for seg in re.split(r"(?:\|\||&&|[|;&\n])", command):
        toks = seg.strip().split()
        while toks and re.match(r"^[A-Za-z_][A-Za-z0-9_]*=", toks[0]):
            toks.pop(0)                       # leading env assignments
        if not toks:
            continue
        if toks[0] != "git" and not toks[0].endswith("/git"):
            continue
        i = 1
        while i < len(toks):                  # skip global options → subcommand
            if toks[i] in ARG_OPTS:
                i += 2
            elif toks[i].startswith("-"):
                i += 1
            else:
                break
        else:
            continue
        if i >= len(toks) or toks[i] != "stash":
            continue
        j = i + 1                             # stash's own first non-option arg
        while j < len(toks) and toks[j].startswith("-"):
            j += 1
        sub = toks[j] if j < len(toks) else ""  # bare stash = push
        if sub not in READ_ONLY:
            return seg.strip()
    return None


def main():
    try:
        payload = json.load(sys.stdin)
        command = payload.get("tool_input", {}).get("command", "")
    except Exception:
        return 0                              # fail-open: malformed input
    if not isinstance(command, str) or "stash" not in command:
        return 0
    seg = offending_segment(command)
    if seg is None:
        return 0
    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "deny",
            "permissionDecisionReason": (
                "git stash is FORBIDDEN on this shared tree (owner ruling "
                f"2026-08-14; blocked: `{seg}`). On 2026-08-12 one bare "
                "`git stash` swept 38 files of ~10 sessions' uncommitted work "
                "and reverted 18 production manifests ~50-100 releases while "
                "leaving the tree looking clean. Instead: commit your task "
                "now, narrowly, by pathspec (CLAUDE.md § Git). To READ an "
                "existing stash: `git stash list` / `git stash show` are "
                "allowed, and extract by path with "
                "`git show 'stash@{N}:<path>' > <path>` — never pop. Full "
                "trap: grep 'git stash' "
                "docs/agent_docs/docs024_key_docs_latest/LANDMINES.md"
            ),
        }
    }))
    return 0


def self_test():
    cases = [
        # (command, should_block)
        ("git stash", True),
        ("git stash push -m wip", True),
        ("git stash pop", True),
        ("git stash apply stash@{0}", True),
        ("git stash drop stash@{0}", True),
        ("cd /tmp && git stash", True),
        ("git -C /home/ant/projects/agentchassis stash", True),
        ("git add . ; git stash ; ls", True),
        ("git stash list", False),
        ("git stash show -p stash@{0}", False),
        ("git show 'stash@{0}:CLAUDE.md'", False),
        ("git log -g --pretty='%gd' refs/stash", False),
        ("grep -n 'git stash' LANDMINES.md", False),
        ("echo git stash is banned", False),
    ]
    failed = 0
    for cmd, want_block in cases:
        got_block = offending_segment(cmd) is not None
        ok = got_block == want_block
        failed += 0 if ok else 1
        print(f"  {'PASS' if ok else 'FAIL'}  block={got_block!s:5}  {cmd}")
    print(f"{len(cases) - failed}/{len(cases)} passed")
    return 1 if failed else 0


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "--self-test":
        sys.exit(self_test())
    sys.exit(main())
