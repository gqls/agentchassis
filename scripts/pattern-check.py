#!/usr/bin/env python3
"""pattern-check.py — ADVISORY. Runs the MECHANICALLY CHECKABLE subset of the
debugging guide's section 9 patterns against what a commit actually changes.

WHY THIS EXISTS (2026-07-19, and the origin is the point):
016b section 9 has 28 patterns. On 2026-07-19 a session wrote entry 26 — "a fix
applied to one branch of a two-branch router reads as done" — and then committed
exactly that mistake eight hours later: it fixed `withPriorCodeRequests` and left
its near-identical twin `withPriorRequests` carrying the same defect. The pattern
was not forgotten. It had been WRITTEN THAT MORNING, by the same session. It was
simply never connected to the moment of editing.

That is the gap this closes, and it is narrow on purpose. Knowing a pattern does
not fire it; something at the moment of the edit has to. A council of reviewers
did eventually catch that instance — correctly, at full reading load — but it
cost two rounds and real credits to answer a question `grep` answers in
milliseconds. The platform's own doctrine already says this for bugs ("this
platform's bugs mostly dissolve under grep + a schema read"); the same holds for
review. Spend the LLM council on judgement, not on what a string comparison can
settle.

WHAT IT DELIBERATELY DOES NOT DO:
- It does not check all 28 patterns. Most are judgement ("trust the rendered
  artefact, not the status") and are not expressible as a diff test. Section 25
  is literally a pattern about a check that CANNOT be written. Pretending
  otherwise would produce noise, and noise is fatal here — see below.
- It does not block. See ADVISORY below.
- It does not replace the council gate. It runs BEFORE it, so the council's
  reading budget is spent on things only a reader can find.

ADVISORY, AND WHY (do not "improve" this into a blocker without reading this):
`.githooks/pre-commit` warns that "a stray non-zero exit here stops the whole
fleet committing." Several sessions share this tree; a false positive that blocks
is a fleet-wide outage, and a check that blocks on a bad day gets disabled
permanently. The failure mode we are guarding against is worth a warning, not a
stoppage. 016b also records what happens to enforcement that annoys people: an
agreed invariant with no mechanical teeth decayed to a comment and was violated
by 84% of the CTA anchors it governed. A check nobody reads is worth nothing, so
precision matters more than coverage — every check here was measured against real
history before being included, and anything that fired on ordinary work was cut.

Set PATTERN_CHECK_STRICT=1 to make findings exit non-zero (opt-in, per-session).

USAGE:
    scripts/pattern-check.py            # staged changes (what the hook runs)
    scripts/pattern-check.py --ref HEAD~5   # audit a range, for measuring
"""
import os
import re
import subprocess
import sys

YELLOW, DIM, BOLD, RESET = "\033[1;33m", "\033[2m", "\033[1m", "\033[0m"


def sh(*args):
    return subprocess.run(args, capture_output=True, text=True).stdout


def changed_files(ref=None):
    if ref:
        out = sh("git", "diff", "--name-only", "--diff-filter=ACMR", ref, "HEAD")
    else:
        out = sh("git", "diff", "--cached", "--name-only", "--diff-filter=ACMR")
    return [f for f in out.splitlines() if f.strip()]


def file_content(path, ref=None):
    """Content as of the commit under test (HEAD for a ref audit, worktree for staged)."""
    if ref:
        return sh("git", "show", f"HEAD:{path}")
    try:
        with open(path, encoding="utf-8", errors="replace") as fh:
            return fh.read()
    except OSError:
        return ""


def changed_hunk_text(path, ref=None):
    """Only the ADDED/REMOVED lines for this file — what this commit actually touched."""
    if ref:
        d = sh("git", "diff", "-U0", ref, "HEAD", "--", path)
    else:
        d = sh("git", "diff", "--cached", "-U0", "--", path)
    return "\n".join(l for l in d.splitlines() if re.match(r"^[+-][^+-]", l))


# ── the CamelCase twin rule ─────────────────────────────────────────────────
# Two identifiers are TWINS when one is the other with exactly ONE CamelCase
# segment inserted. withPriorCodeRequests -> [with,Prior,Code,Requests] and
# withPriorRequests -> [with,Prior,Requests] differ by inserting "Code": twins.
# Anything differing by two or more segments is a different function that merely
# reads similarly, and pairing those is where the false positives live.
def segments(name):
    return re.findall(r"[A-Z]+(?![a-z])|[A-Z][a-z0-9_]*|^[a-z0-9_]+", name)


def is_twin(a, b):
    sa, sb = segments(a), segments(b)
    if len(sa) == len(sb):
        return False
    if abs(len(sa) - len(sb)) != 1:
        return False
    longer, shorter = (sa, sb) if len(sa) > len(sb) else (sb, sa)
    for i in range(len(longer)):
        if longer[:i] + longer[i + 1:] == shorter:
            return True
    return False


GO_FUNC = re.compile(r"^func(?:\s+\([^)]*\))?\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(", re.M)


def check_untouched_twin(files, ref, findings):
    """016b section 9 #26 — a fix applied to one of a near-identical pair."""
    for path in files:
        if not path.endswith(".go") or path.endswith("_test.go"):
            continue
        content = file_content(path, ref)
        if not content:
            continue
        all_funcs = GO_FUNC.findall(content)
        hunks = changed_hunk_text(path, ref)
        touched = {f for f in all_funcs if re.search(r"\b" + re.escape(f) + r"\b", hunks)}
        for t in sorted(touched):
            for other in all_funcs:
                if other in touched or not is_twin(t, other):
                    continue
                findings.append((
                    "untouched-twin", path,
                    f"changed {BOLD}{t}(){RESET} but not its twin {BOLD}{other}(){RESET}",
                    "016b section 9 #26 — same shape, same file, one edited. If the change is a "
                    "fix, the twin probably has the same defect; if it is a feature, the twin "
                    "probably needs it too. Deliberate? Say so in the commit message.",
                ))


def check_gofmt(files, ref, findings):
    """016b section 9 #16 — un-gofmt'd Go fails the build gate and yields no PR."""
    if ref:
        return  # only meaningful for a commit about to happen
    gofiles = [f for f in files if f.endswith(".go") and os.path.exists(f)]
    if not gofiles:
        return
    out = sh("gofmt", "-l", *gofiles).split()
    for f in out:
        findings.append((
            "gofmt", f, "not gofmt-clean",
            "016b section 9 #16 — the build gate rejects un-gofmt'd code, so this "
            "reaches CI as a failed gate and no PR. Run: gofmt -w " + f,
        ))


STDIN_EATERS = re.compile(r"\b(kubectl\s+[^|]*\bexec\b[^|]*\s-\w*i|ssh\s|ffmpeg\s)")


def check_stdin_eater(files, ref, findings):
    """016b section 9 #20 — a stdin-reading command truncates the while-read loop."""
    for path in files:
        if not (path.endswith(".sh") or path.endswith(".bash")):
            continue
        content = file_content(path, ref)
        lines = content.splitlines()
        depth_stack = []
        for i, line in enumerate(lines):
            if re.search(r"\bwhile\b.*\bread\b", line):
                depth_stack.append(i)
            if re.match(r"^\s*done\b", line) and depth_stack:
                start = depth_stack.pop()
                body = "\n".join(lines[start:i])
                m = STDIN_EATERS.search(body)
                if m and "</dev/null" not in body and "< /dev/null" not in body:
                    findings.append((
                        "stdin-eater", f"{path}:{start + 1}",
                        f"`while read` loop calls {BOLD}{m.group(1).strip()}{RESET} without redirecting its stdin",
                        "016b section 9 #20 — that command consumes the rest of the loop's input, "
                        "so the loop silently processes only the first few items (it hid 90% of "
                        "the council coverage report). Add < /dev/null to the inner command.",
                    ))


# Pairs that must change together, where the relationship is real but invisible
# to the compiler. Kept SHORT on purpose: each entry is a documented incident,
# not a guess. A speculative pair here would fire on ordinary work and teach
# people to ignore the whole script.
DECLARED_PAIRS = [
    (r"idx_swi_dedup", r"workItemTerminalStatuses",
     "016b/memory — the dedup index's status set and the Go terminal-status list are ONE "
     "contract. Drift gives fleet-wide 42P10 on every keyed insert (bit 2026-07-16)."),
]


def check_declared_pairs(files, ref, findings):
    blobs = {p: (changed_hunk_text(p, ref) or "") for p in files}
    joined = "\n".join(blobs.values())
    for a, b, why in DECLARED_PAIRS:
        has_a, has_b = re.search(a, joined), re.search(b, joined)
        if bool(has_a) != bool(has_b):
            present, missing = (a, b) if has_a else (b, a)
            findings.append((
                "unpaired-change", "(commit)",
                f"touches {BOLD}{present}{RESET} but not {BOLD}{missing}{RESET}", why,
            ))


def main():
    ref = None
    if "--ref" in sys.argv:
        ref = sys.argv[sys.argv.index("--ref") + 1]
    files = changed_files(ref)
    if not files:
        return 0

    findings = []
    for check in (check_untouched_twin, check_gofmt, check_stdin_eater, check_declared_pairs):
        try:
            check(files, ref, findings)
        except Exception as e:  # never let a check break a commit
            print(f"{DIM}pattern-check: {check.__name__} skipped ({e}){RESET}", file=sys.stderr)

    if not findings:
        return 0

    print(f"\n{YELLOW}── pattern check: {len(findings)} thing(s) worth a look ──{RESET}")
    for kind, where, what, why in findings:
        print(f"   {BOLD}{kind}{RESET}  {where}")
        print(f"     {what}")
        print(f"     {DIM}{why}{RESET}")
    print(f"   {DIM}Advisory — this never blocks. PATTERN_CHECK_STRICT=1 makes it exit non-zero.{RESET}")
    print(f"   {DIM}Full patterns: docs/agent_docs/docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md section 9{RESET}\n")

    return 1 if os.environ.get("PATTERN_CHECK_STRICT") == "1" else 0


if __name__ == "__main__":
    sys.exit(main())
