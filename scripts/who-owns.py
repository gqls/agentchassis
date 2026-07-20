#!/usr/bin/env python3
"""who-owns.py — ADVISORY. Answers "does this bug already have a thread working
on it?" before you route work at it.

WHY THIS EXISTS (2026-07-20, and the origin is the point):
A session finished a site fix, wrote a fresh handoff, and promoted
`bugs_open/023` to "the highest-value fix here, and it is a code fix" — with
implementation direction, as though nobody was on it. The `cta_link_integrity`
workstream owned it, was **six council rounds in**, and its observe-only stage
had already shipped live in the image that rolled that same evening.

Nothing caught it. The session had followed CLAUDE.md's "grep before you file"
— which covers filing a NEW bug — and then routed work at an EXISTING one,
which that rule does not cover and which has exactly the same failure mode. It
surfaced only because the owner happened to mention the fresh build, which led
to the commit, which led to the workstream. Pure luck.

The cost of the miss is asymmetric and that is why this is worth a script: the
next session would have started a competing fix against a live staged rollout,
on a shared branch, in the area a council trail was mid-flight. Duplicated work
is the cheap outcome; the expensive one is two threads editing the same
functions from different premises.

WHAT IT DOES:
Resolves a bug number or slug to its file(s), finds every workstream directory
that references it, and shows recent commit activity for both. Then prints one
verdict line. All of it is grep and `git log` — no LLM, no cluster calls, runs
in well under a second.

WHAT IT DELIBERATELY DOES NOT DO:
- It does not decide for you. "OWNED" is a prompt to go and read, not a refusal.
- It does not check the live DB for work items. Dispatching at a *site* is a
  different coverage question and CLAUDE.md already covers it (the 090 trigger
  performs that check itself and refuses on a hit).
- It has no opinion about bugs with no owner. Most bugs have no owner; that is
  the normal case and is not a finding.

USAGE
    scripts/who-owns.py 023
    scripts/who-owns.py cta_label_url          # slug fragment
    scripts/who-owns.py bugs_open/023_HANDOFF_2026-07-19_cta_label_url_pairing_unchecked.md
    scripts/who-owns.py 023 --days 30          # widen the activity window
"""
import argparse
import os
import re
import subprocess
import sys
from collections import defaultdict

ROOT = subprocess.run(["git", "rev-parse", "--show-toplevel"],
                      capture_output=True, text=True).stdout.strip() or "."
BUG_DIRS = ["bugs_open", "bugs_closed"]
DOCS = os.path.join("docs", "agent_docs", "docs024_key_docs_latest")


def sh(*args):
    return subprocess.run(args, cwd=ROOT, capture_output=True, text=True).stdout


def find_bug_files(token):
    """Resolve a token to bug files. Numbering is ONE sequence shared across
    bugs_open and bugs_closed and is never reassigned, so a number can resolve
    into either directory — and several numbers are used by two unrelated
    cases (016 and 017 documented; more have appeared since). Return all
    matches and let the caller warn."""
    hits = []
    is_num = re.fullmatch(r"\d{1,3}", token)
    for d in BUG_DIRS:
        p = os.path.join(ROOT, d)
        if not os.path.isdir(p):
            continue
        for f in sorted(os.listdir(p)):
            if not f.endswith(".md"):
                continue
            if is_num:
                if re.match(rf"0*{int(token):03d}_", f):
                    hits.append(os.path.join(d, f))
            elif token.lower() in f.lower():
                hits.append(os.path.join(d, f))
    if not hits and os.path.exists(os.path.join(ROOT, token)):
        hits.append(token)
    return hits


def referencing_workstreams(token, bug_files):
    """Which workstream dirs mention this bug. Search on the number, the slug
    stem, and the filename — a doc may cite any of them."""
    needles = {token}
    for bf in bug_files:
        stem = os.path.basename(bf)[:-3]
        needles.add(stem)
        m = re.match(r"(\d{3})_", stem)
        if m:
            needles.add(f"bugs_open/{m.group(1)}")
            needles.add(f"bugs_closed/{m.group(1)}")
            # the bare "NNN" form used in prose, e.g. "bug 023" / "/bugs_open/023"
            needles.add(m.group(1))
        tail = re.sub(r"^\d{3}_HANDOFF_\d{4}-\d{2}-\d{2}_", "", stem)
        if len(tail) > 8:
            needles.add(tail)

    docs_path = os.path.join(ROOT, DOCS)
    if not os.path.isdir(docs_path):
        return {}

    # Count MENTIONS, not just which files matched. A workstream that cites a
    # bug once in a "related" list is not its owner; the one whose notes are
    # full of it usually is. Conflating the two is what makes a tool cry wolf.
    per_dir = defaultdict(lambda: {"files": set(), "mentions": 0})
    for needle in needles:
        # -F fixed string, -c count per file. A bare 3-digit needle is noisy,
        # so require it to sit next to a bug-ish word.
        if re.fullmatch(r"\d{3}", needle):
            pat = rf"bugs?[_/ ]?(open|closed)?[_/ ]?{needle}\b"
            out = sh("grep", "-rcoiE", pat, "--include=*.md", docs_path)
        else:
            out = sh("grep", "-rcoiF", needle, "--include=*.md", docs_path)
        for line in out.splitlines():
            if ":" not in line:
                continue
            path, _, count = line.rpartition(":")
            if not count.isdigit() or count == "0":
                continue
            rel = os.path.relpath(path, ROOT)
            parts = rel.split(os.sep)
            if len(parts) > 4:
                d = os.sep.join(parts[:4])
                per_dir[d]["files"].add(os.path.basename(rel))
                per_dir[d]["mentions"] += int(count)
    return per_dir


def commits_about(token, bug_files, days):
    """Commits whose SUBJECT is about this bug. The strongest ownership signal
    available without asking anyone: a thread that merely cites a bug does not
    write its number into commit subjects; the thread working it does."""
    pats = [rf"\b0*{int(token):03d}\b"] if re.fullmatch(r"\d{1,3}", token) else [re.escape(token)]
    for bf in bug_files:
        stem = os.path.basename(bf)[:-3]
        tail = re.sub(r"^\d{3}_HANDOFF_\d{4}-\d{2}-\d{2}_", "", stem)
        if len(tail) > 8:
            pats.append(re.escape(tail.replace("_", ".")))
    seen, out = set(), []
    for p in pats:
        # NB `git log --grep` matches the WHOLE commit message, not the subject.
        # Left unfiltered it pulls in every commit that merely cites the bug in
        # its body — which is precisely the mention-vs-ownership conflation this
        # script exists to avoid, so post-filter on the subject line itself.
        raw = sh("git", "log", f"--since={days} days ago", "-E", f"--grep={p}",
                 "-i", "--format=%h|%ad|%an|%s", "--date=short")
        rx = re.compile(p, re.I)
        for line in raw.splitlines():
            if not line.strip():
                continue
            h, _, _, subj = (line.split("|", 3) + ["", "", ""])[:4]
            if h in seen or not rx.search(subj):
                continue
            seen.add(h)
            out.append(line)
    return out


def last_activity(path, days):
    out = sh("git", "log", f"--since={days} days ago", "--format=%h|%ad|%an|%s",
             "--date=short", "--", path)
    return [l for l in out.splitlines() if l.strip()]


def main():
    ap = argparse.ArgumentParser(add_help=True)
    ap.add_argument("token", help="bug number (023), slug fragment, or path")
    ap.add_argument("--days", type=int, default=14,
                    help="activity window in days (default 14)")
    a = ap.parse_args()

    bug_files = find_bug_files(a.token)
    if not bug_files:
        print(f"No bug file matches '{a.token}'.")
        print("Numbering is one sequence across bugs_open/ and bugs_closed/ and is")
        print("never reassigned — try the slug if a number does not resolve.")
        return 0

    print(f"\n=== bug file(s) matching '{a.token}' ===")
    for f in bug_files:
        print(f"  {f}")
    if len(bug_files) > 1:
        print("\n  ** AMBIGUOUS NUMBER — this is a documented trap. Two unrelated cases")
        print("     share it. Refer to the one you mean BY SLUG, never by number. **")

    active = False

    print(f"\n=== commits touching the bug file(s), last {a.days}d ===")
    any_commit = False
    for f in bug_files:
        for line in last_activity(f, a.days):
            h, d, who, subj = line.split("|", 3)
            print(f"  {d}  {h}  {who:<12} {subj[:78]}")
            any_commit = True
            active = True
    if not any_commit:
        print("  (none)")

    subject_commits = commits_about(a.token, bug_files, a.days)
    print(f"\n=== commits whose SUBJECT is about this bug, last {a.days}d ===")
    if subject_commits:
        for line in subject_commits[:10]:
            h, dt, who, subj = line.split("|", 3)
            print(f"  {dt}  {h}  {who:<12} {subj[:78]}")
        if len(subject_commits) > 10:
            print(f"  ... and {len(subject_commits)-10} more")
        active = True
    else:
        print("  (none)")

    per_dir = referencing_workstreams(a.token, bug_files)
    # Separate the workstream that is WORKING it from those merely CITING it.
    # Heuristic, deliberately crude and stated as such: many mentions, or a dir
    # whose own name overlaps the bug slug. One passing "see also" is not
    # ownership, and treating it as such is how a checker earns being ignored.
    owners, citers = [], []
    for d, info in per_dir.items():
        stem = os.path.basename(bug_files[0])[:-3].lower()
        dirname = os.path.basename(d).lower()
        name_overlap = any(w in stem for w in dirname.split("_") if len(w) > 4)
        (owners if (info["mentions"] >= 5 or name_overlap) else citers).append((d, info))

    owners.sort(key=lambda t: -t[1]["mentions"])
    citers.sort(key=lambda t: -t[1]["mentions"])

    print(f"\n=== likely OWNING workstream(s) ===")
    if not owners:
        print("  (none identified)")
    for d, info in owners:
        commits = last_activity(d, a.days)
        flag = f"ACTIVE, {len(commits)} commits/{a.days}d" if commits else f"quiet {a.days}d"
        print(f"\n  {d}   [{flag}]  {info['mentions']} mentions")
        docs = sorted(info["files"])
        print(f"    in: {', '.join(docs[:5])}"
              + (f"  (+{len(docs)-5} more)" if len(docs) > 5 else ""))
        for line in commits[:3]:
            h, dt, who, subj = line.split("|", 3)
            print(f"    {dt}  {h}  {who:<12} {subj[:68]}")
        if commits:
            active = True

    if citers:
        print(f"\n=== also cites it (probably just a cross-reference) ===")
        for d, info in citers:
            print(f"  {d}  ({info['mentions']} mention{'s' if info['mentions'] > 1 else ''})")

    print("\n" + "=" * 72)
    if owners or subject_commits:
        print("VERDICT: OWNED or recently active.")
        print("  Read the owning workstream's docs BEFORE routing work at this bug.")
        print("  Contribute findings INTO the bug file itself (the shared account),")
        print("  not a parallel one. Do not start a competing fix.")
    elif active:
        print(f"VERDICT: some activity in {a.days}d but no clear owner — read the")
        print("  citing dirs above, then proceed with care.")
    else:
        print(f"VERDICT: no activity in {a.days}d and no owning workstream found.")
        print("  Likely unowned. Widen with --days if the bug is older.")
    print("=" * 72 + "\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
