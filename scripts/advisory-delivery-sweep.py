#!/usr/bin/env python3
"""advisory-delivery-sweep.py — does the pre-commit advisory actually REACH the
session that commits? Measured from the harness's own transcripts.

This is OPP-007's verify-later, made runnable. The finding it comes from
(docs026_concept_register/FINDINGS_2026-08-11_advisory_delivery.md): the
pre-commit hook prints the commit-scope report and pattern-check's findings
FIRST, git prints its `[branch sha]` summary LAST, and sessions habitually pipe
`git commit … 2>&1 | tail -5`. On 2026-08-11 that cost 45% of multi-file commits
their advisory.

WHAT IT COUNTS, and why each filter is there rather than being a convenience:

- **multi-file commits only.** commit-scope-report.sh exits silently for a
  single-file commit by design (`[ "$n" -le 1 ] && exit 0`), so counting those
  would manufacture ~2,600 misses that were never due. This is the filter that
  decides the answer, so it is stated, not buried.
- **commits that resolve in THIS repo only.** Sessions commit into other repos
  through the same tool (the auto-memory git dir, scratch repos) where no hook is
  installed; 71 of 5,195 shas in the 2026-08-11 corpus were foreign.
- **since the report shipped** (eae296850, 2026-07-18 17:22). Earlier commits have
  no block to miss, and including them inflates the miss rate for free.

CONTROL, printed every run and the reason to believe the result: `tail -N` width
is compared between misses and commits that delivered DESPITE a pipe. If the pipe
is the cause, misses cluster at small N and deliveries at large N. If that
separation ever disappears, the mechanism claimed here is not the one operating
and the finding needs re-deriving, not re-quoting.

USAGE:
    scripts/advisory-delivery-sweep.py                  # full sweep
    scripts/advisory-delivery-sweep.py --since 2026-08-11   # after the fix landed
"""
import glob
import json
import os
import re
import subprocess
import sys
from collections import Counter, defaultdict

TRANSCRIPTS = os.path.expanduser("~/.claude/projects/-home-ant-projects-agentchassis")
REPO = os.environ.get("CLAUDE_PROJECT_DIR", "/home/ant/projects/agentchassis")
SHIPPED = "2026-07-18T17:00"          # eae296850 — the scope report's own commit

NFILES = re.compile(r"^ (\d+) files? changed", re.M)
TAIL = re.compile(r"\|\s*tail\s+-n?\s*(\d+)")


def records(since):
    """(sha, day, nfiles, delivered, tailN, transcript) per commit made via the tool."""
    for fn in sorted(glob.glob(os.path.join(TRANSCRIPTS, "*.jsonl"))):
        try:
            lines = open(fn, errors="replace").readlines()
        except OSError:
            continue
        for i, line in enumerate(lines):
            if '"gitOperation"' not in line:          # cheap prefilter: 1.4 GB of transcripts
                continue
            try:
                d = json.loads(line)
            except Exception:
                continue
            r = d.get("toolUseResult")
            if not isinstance(r, dict):
                continue
            commit = (r.get("gitOperation") or {}).get("commit")
            if not commit or commit.get("kind") != "committed":
                continue
            ts = d.get("timestamp", "")
            if ts < since:
                continue
            out = r.get("stdout") or ""
            m = NFILES.search(out)
            if not m:
                continue
            cmd = ""
            for j in range(i - 1, max(i - 6, -1), -1):      # the tool_use that produced it
                try:
                    dj = json.loads(lines[j])
                except Exception:
                    continue
                c = dj.get("message", {}).get("content")
                if isinstance(c, list):
                    tu = [b for b in c if b.get("type") == "tool_use"]
                    if tu:
                        cmd = " ".join((tu[-1].get("input") or {}).get("command", "").split())
                        break
            t = TAIL.search(cmd)
            yield (commit.get("sha") or "", ts[:10], int(m.group(1)),
                   "commit scope:" in out, int(t.group(1)) if t else None,
                   os.path.basename(fn), len(out.rstrip("\n").split("\n")))


def main():
    since = SHIPPED
    if "--since" in sys.argv:
        since = sys.argv[sys.argv.index("--since") + 1]
        if len(since) == 10:
            since += "T00:00"

    rows = [r for r in records(since) if r[2] > 1]
    if not rows:
        print("no multi-file commits in range")
        return 0

    shas = sorted({r[0] for r in rows})
    p = subprocess.run(["git", "cat-file", "--batch-check"],
                       input="\n".join(s + "^{commit}" for s in shas),
                       capture_output=True, text=True, cwd=REPO)
    here = {s for s, line in zip(shas, p.stdout.splitlines()) if " commit " in line}
    foreign = len(shas) - len(here)

    seen, uniq = set(), []
    for r in rows:                       # one row per sha; a sha can appear in two transcripts
        if r[0] in here and r[0] not in seen:
            seen.add(r[0])
            uniq.append(r)

    deliv = [r for r in uniq if r[3]]
    miss = [r for r in uniq if not r[3]]
    tailed = [r for r in miss if r[4] is not None]
    despite = [r for r in deliv if r[4] is not None]
    exact = [r for r in tailed if r[6] == r[4]]

    print(f"window: commits since {since}   (scope report shipped {SHIPPED})")
    print(f"multi-file commits in this repo, made through the tool : {len(uniq)}")
    print(f"   advisory block DELIVERED : {len(deliv):5d}  ({100*len(deliv)/len(uniq):.1f}%)")
    print(f"   NOT delivered            : {len(miss):5d}  ({100*len(miss)/len(uniq):.1f}%)")
    if miss:
        print(f"      cut by the session's own `| tail` : {len(tailed)} "
              f"({100*len(tailed)/len(miss):.1f}% of misses)")
        print(f"      of those, stdout is EXACTLY N lines (output existed and was cut): {len(exact)}")
        print(f"      unexplained residue (no tail pipe) : {len(miss)-len(tailed)}")
    print(f"   commits in other repos, excluded          : {foreign}")
    print(f"   distinct sessions that suppressed a block : {len({r[5] for r in miss})}")

    small = lambda rs: sum(1 for r in rs if r[4] <= 8)
    big = lambda rs: sum(1 for r in rs if r[4] > 8)
    print("\nCONTROL — tail -N width, misses vs delivered-despite-a-pipe")
    print(f"   N <= 8 :  misses {small(tailed):5d}   delivered {small(despite):5d}")
    print(f"   N >  8 :  misses {big(tailed):5d}   delivered {big(despite):5d}")
    print("   (the finding requires misses to cluster small and deliveries large;")
    print("    if this separation is gone, re-derive the cause — do not re-quote it)")

    by_day = defaultdict(lambda: [0, 0])
    for r in uniq:
        by_day[r[1]][0] += 1
        by_day[r[1]][1] += 1 if r[3] else 0
    print("\nby day (last 14):")
    for day in sorted(by_day)[-14:]:
        t, o = by_day[day]
        print(f"   {day}  due={t:4d}  delivered={o:4d}  ({100*o/t:3.0f}%)")
    print("\ntail -N among misses:", Counter(r[4] for r in tailed).most_common(6))
    return 0


if __name__ == "__main__":
    sys.exit(main())
