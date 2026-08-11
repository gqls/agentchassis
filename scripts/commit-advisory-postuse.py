#!/usr/bin/env python3
"""commit-advisory-postuse.py — re-emit the pre-commit advisory AFTER the commit,
out of band, because roughly half the fleet never sees it.

WHY THIS EXISTS (measured 2026-08-11, and the origin is the point):

`.githooks/pre-commit` prints two advisory blocks — the commit-scope report (which
files this commit actually contains, the estate's main defence against sweeping
another session's work) and pattern-check's findings (016b section 9, including the
concept-register authoring gate OPP-006). Both print FIRST; git prints its own
`[branch sha] message` summary LAST. Sessions overwhelmingly run

    git commit <paths> -m "..." 2>&1 | tail -5

which keeps git's summary and throws the advisory away. It is not a bad habit
inconsistently applied — it is nearly universal, and the arithmetic is what makes
it systematic: the part you want is at the tail, the part that warns you is at the
head, and `tail -N` is a head-cutter.

MEASURED over every commit made through the tool since the scope report shipped
(2026-07-18) — 2,665 multi-file commits in this repo, from the session transcripts:

    advisory block DELIVERED to the session : 1,467  (55%)
    NOT delivered                           : 1,198  (45%)
       of those, cut by the session's own `| tail` : 1,136  (95%)

with the control that makes it a finding rather than a correlation: among the
misses `tail -N` clusters at N=3-8 (median 5); among commits that DID deliver
despite a pipe, N clusters at 12-30. The cause is the pipe width, and the
measurement could have come out otherwise.

That is also the answer to a question the concept-register lane had left open in
HANDOFF_2026-08-10b: OPP-006 fired on commit 5c7b115c5 three hours after shipping
and the entry landed with no index row anyway. The handoff said the honest first
question is whether that session never saw the output or saw it and judged the row
could wait, "and those have opposite fixes (delivery vs. enforcement)". The
transcript settles it: that command ended `2>&1 | tail -8`, and the recorded
stdout is exactly 8 lines long. The session never saw it. This is the delivery fix;
no case for teeth was made out.

HOW IT DELIVERS, which is the part that is easy to get wrong:
`hookSpecificOutput.additionalContext` on STDOUT with exit 0. A PostToolUse hook
that writes to stderr and exits 0 reaches neither the model nor the transcript —
scripts/memory-index.py was wired that way and was mute for six days while sessions
hand-rolled the number it was computing. Siblings that work: landmines-session-start.py,
.claude/hooks/psql_readonly_gate.py.

WHAT IT DOES NOT DO:
- It does not block, and it cannot: PostToolUse runs after the tool. The commit
  already exists. This buys visibility one moment later, not enforcement.
- It does not duplicate. If the session's own output already carries a block, that
  block is dropped from what we re-emit.
- It does not guess. The commit sha is read from git's own summary line, and must
  resolve in THIS repo, so a commit made in another repo (the auto-memory git dir,
  a scratch repo) is silently ignored rather than reported against the wrong tree.
"""
import json
import os
import re
import subprocess
import sys

ANSI = re.compile(r"\x1b\[[0-9;]*m")
# git's own summary: `[branch abc1234] subject`, or `[detached HEAD abc1234] subject`.
# It is the LAST thing git prints, which is precisely why it survives `| tail -N`
# when the advisory does not.
SUMMARY = re.compile(r"^\[[^\]\n]*?\b([0-9a-f]{7,40})\]", re.M)
MAX_CHARS = 4000


def run(args, cwd):
    try:
        p = subprocess.run(args, capture_output=True, text=True, cwd=cwd, timeout=15)
        return ANSI.sub("", p.stdout).strip("\n")
    except Exception:
        return ""


def main():
    try:
        payload = json.load(sys.stdin)
    except Exception:
        return 0

    if payload.get("tool_name") != "Bash":
        return 0
    command = (payload.get("tool_input") or {}).get("command") or ""
    if "git commit" not in command:          # cheap guard: no subprocess for ordinary Bash calls
        return 0

    resp = payload.get("tool_response")
    if isinstance(resp, dict):
        seen = (resp.get("stdout") or "") + (resp.get("stderr") or "")
    else:
        seen = str(resp or "")

    m = SUMMARY.search(seen)
    if not m:                                 # commit failed, was quiet, or output went to /dev/null
        return 0
    sha = m.group(1)

    root = os.environ.get("CLAUDE_PROJECT_DIR") or os.getcwd()
    # Must be a commit in THIS repo. A commit in the auto-memory git dir or a
    # scratch repo resolves nowhere here, and reporting the wrong tree's scope
    # would be worse than staying quiet.
    try:
        ok = subprocess.run(["git", "cat-file", "-e", sha + "^{commit}"],
                            cwd=root, capture_output=True, timeout=10).returncode == 0
    except Exception:
        return 0
    if not ok:
        return 0

    blocks = []
    scope = run([os.path.join(root, "scripts", "commit-scope-report.sh"), "--commit", sha], root)
    if scope and "commit scope:" not in seen:
        blocks.append(scope)
    patt = run([os.path.join(root, "scripts", "pattern-check.py"), "--commit", sha], root)
    if patt and "pattern check:" not in seen:
        blocks.append(patt)

    if not blocks:
        return 0

    body = "\n".join(blocks).strip()
    if len(body) > MAX_CHARS:
        body = body[:MAX_CHARS] + "\n   … truncated; run the command above yourself for the rest."

    note = (f"Advisory for the commit you just made ({sha}). Your command's output did not "
            f"carry this — the pre-commit hook prints it FIRST and git prints its summary "
            f"LAST, so a `| tail -N` keeps the summary and cuts this. Nothing is blocked and "
            f"nothing needs undoing; forward-only still holds. Act on it with a follow-up "
            f"commit if it names something of yours.\n\n" + body + "\n\n"
            f"   ↳ re-run at will: scripts/commit-scope-report.sh --commit {sha} ; "
            f"scripts/pattern-check.py --commit {sha}")

    json.dump({"hookSpecificOutput": {"hookEventName": "PostToolUse",
                                      "additionalContext": note}}, sys.stdout)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:
        sys.exit(0)          # a hook that breaks the Bash tool for every session is unthinkable
