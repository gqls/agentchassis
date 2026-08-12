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

TWO CHANNELS, and reading only one of them is how this script was wrong for a day
(corrected 2026-08-12):

1. **the command's own stdout** — the pre-commit hook's print, which a `| tail -N`
   cuts. This is the channel the 45% finding is about.
2. **out of band** — OPP-007 (`scripts/commit-advisory-postuse.py`), a PostToolUse
   hook delivering `hookSpecificOutput.additionalContext`, which the harness records
   as a SEPARATE transcript record: `type: "attachment"`, `attachment.type:
   "hook_success"`, the text in `attachment.stdout`. It is not in `toolUseResult`
   at all.

Until 2026-08-12 this script read only channel 1, while FINDINGS_2026-08-11 and this
docstring both named `--since <the day after the fix>` as OPP-007's verify-later. So
the instrument was structurally blind to the one path it existed to verify: it scored
every out-of-band delivery as a MISS, and its first post-fix run read 38% — *worse*
than the 55% baseline — on a day when the hook was in fact delivering. A reading like
that invites exactly the wrong conclusion (that the fix failed, or that enforcement is
needed), which is the conclusion OPP-007's own evidence had just weakened.

CONTROLS, printed every run and the reason to believe the result:

- **`tail -N` width, channel 1 only.** If the pipe is the cause, misses cluster at
  small N and deliveries-despite-a-pipe at large N. If that separation ever
  disappears, the mechanism claimed here is not the one operating and the finding
  needs re-deriving, not re-quoting.
- **channel 2 cannot predate its own hook.** Zero advisory attachments may exist
  before 2026-08-11T18:11Z, when OPP-007 first ran. If the parse were matching
  something else (this script, the FINDINGS doc, a Read of either), the count would
  light up across the whole corpus back to July. ⚠ Transcript timestamps are **UTC**
  and the estate writes **BST**: the hook shipped at 19:18 BST = 18:18 UTC, and it was
  live in the working tree ~7 minutes before its own commit. A naive compare against
  "19:00" scores 24 legitimate deliveries as impossible ones.

USAGE:
    scripts/advisory-delivery-sweep.py                  # full sweep
    scripts/advisory-delivery-sweep.py --since 2026-08-12   # after OPP-007 landed
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
# OPP-007's first live run, measured from the transcripts (18:11:38Z). UTC, not BST,
# and EARLIER than its own commit (19:18 BST = 18:18Z) because it was live in the
# working tree before it was committed. Used only as the channel-2 control bound.
OPP007_LIVE = "2026-08-11T18:11"

NFILES = re.compile(r"^ (\d+) files? changed", re.M)
TAIL = re.compile(r"\|\s*tail\s+-n?\s*(\d+)")
# channel 2: the sha out of OPP-007's own note. Its wording is the discriminator
# rather than the hook's name, so renaming the hook cannot silently empty this.
OOB = re.compile(r"commit you just made \(([0-9a-f]{7,40})\)")


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


def out_of_band():
    """Channel 2. {sha_prefix_len: {sha}} for every OPP-007 delivery, plus the
    pre-ship control count and a per-day tally.

    Keyed by prefix length because the hook quotes git's ABBREVIATED sha (9 chars
    today) while the population carries the full 40 — a straight set membership test
    silently matches nothing, which would look exactly like "the hook never fired".
    """
    bylen, byday, preship, total = defaultdict(set), Counter(), 0, 0
    for fn in sorted(glob.glob(os.path.join(TRANSCRIPTS, "*.jsonl"))):
        try:
            lines = open(fn, errors="replace").readlines()
        except OSError:
            continue
        for line in lines:
            if "commit you just made" not in line:      # cheap prefilter
                continue
            try:
                d = json.loads(line)
            except Exception:
                continue
            a = d.get("attachment")
            if d.get("type") != "attachment" or not isinstance(a, dict):
                continue
            if a.get("type") != "hook_success":
                continue
            m = OOB.search(a.get("stdout") or "")
            if not m:
                continue
            ts = d.get("timestamp") or ""
            total += 1
            if ts < OPP007_LIVE:
                preship += 1
            bylen[len(m.group(1))].add(m.group(1))
            byday[ts[:10]] += 1
    return bylen, total, preship, byday


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

    oob_bylen, oob_total, oob_preship, oob_byday = out_of_band()
    oob_hit = lambda sha: any(sha[:n] in s for n, s in oob_bylen.items())

    own = [r for r in uniq if r[3]]                       # channel 1: their own stdout
    oob = [r for r in uniq if not r[3] and oob_hit(r[0])]  # channel 2: OPP-007
    both = [r for r in own if oob_hit(r[0])]              # scope in stdout, pattern-check out of band
    miss = [r for r in uniq if not r[3] and not oob_hit(r[0])]
    reached = len(own) + len(oob)
    tailed = [r for r in miss if r[4] is not None]
    despite = [r for r in own if r[4] is not None]
    exact = [r for r in tailed if r[6] == r[4]]

    print(f"window: commits since {since}   (scope report shipped {SHIPPED})")
    print(f"multi-file commits in this repo, made through the tool : {len(uniq)}")
    print(f"   REACHED the session, either channel : {reached:5d}  ({100*reached/len(uniq):.1f}%)")
    print(f"      in the command's own output          : {len(own):5d}")
    print(f"      out of band, by OPP-007              : {len(oob):5d}"
          f"   (+{len(both)} where stdout had the scope block and the hook added pattern-check)")
    print(f"   NOT delivered by either channel      : {len(miss):5d}  ({100*len(miss)/len(uniq):.1f}%)")
    if miss:
        print(f"      cut by the session's own `| tail` : {len(tailed)} "
              f"({100*len(tailed)/len(miss):.1f}% of misses)")
        print(f"      of those, stdout is EXACTLY N lines (output existed and was cut): {len(exact)}")
        print(f"      unexplained residue (no tail pipe) : {len(miss)-len(tailed)}")
    print(f"   commits in other repos, excluded          : {foreign}")
    print(f"   distinct sessions still not reached       : {len({r[5] for r in miss})}")

    print("\nCONTROL — channel 2 cannot predate its own hook")
    print(f"   OPP-007 deliveries recorded, all time : {oob_total}"
          f"   (before it went live at {OPP007_LIVE}Z: {oob_preship} — must be 0)")
    print("   ⚠ transcripts are UTC, the estate writes BST; the hook was live in the")
    print("     working tree ~7 min before its own commit. Compare in UTC or this fails.")
    if oob_preship:
        print("   ✗ CONTROL FAILED — the parse is matching something other than a delivery.")

    small = lambda rs: sum(1 for r in rs if r[4] <= 8)
    big = lambda rs: sum(1 for r in rs if r[4] > 8)
    print("\nCONTROL — tail -N width, CHANNEL 1 ONLY: misses vs delivered-despite-a-pipe")
    print(f"   N <= 8 :  misses {small(tailed):5d}   delivered {small(despite):5d}")
    print(f"   N >  8 :  misses {big(tailed):5d}   delivered {big(despite):5d}")
    print("   (the finding requires misses to cluster small and deliveries large;")
    print("    if this separation is gone, re-derive the cause — do not re-quote it)")

    by_day = defaultdict(lambda: [0, 0, 0])
    for r in uniq:
        by_day[r[1]][0] += 1
        by_day[r[1]][1] += 1 if r[3] else 0
        by_day[r[1]][2] += 1 if (not r[3] and oob_hit(r[0])) else 0
    print("\nby day (last 14) — 'own' is the command's output, 'oob' is OPP-007:")
    for day in sorted(by_day)[-14:]:
        t, o, b = by_day[day]
        print(f"   {day}  due={t:4d}  own={o:4d}  oob={b:4d}  reached={100*(o+b)/t:3.0f}%")
    print("\ntail -N among misses:", Counter(r[4] for r in tailed).most_common(6))
    return 0


if __name__ == "__main__":
    sys.exit(main())
