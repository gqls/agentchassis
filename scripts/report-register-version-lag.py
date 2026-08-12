#!/usr/bin/env python3
"""report-register-version-lag.py — which concept-register entries cite a chassis
version, and how far behind the fleet is it?

This is the "make version lag visible" item from
docs026_concept_register/HANDOFF_2026-08-10b_continue_here.md, and it is a REPORT,
not a checker. It never says an entry is wrong. It says **this entry's evidence has
expired** — and for one class it can say that mechanically and be right.

WHY IT KEYS ON THE FIELD, NOT THE PROSE (measured 2026-08-12, and the negative
result is the reason this script has the shape it does):

The register's design conclusion already forbids parsing the `status:` field's prose
— 38 entries matched a "not live yet" regex and reading all 38 said ~20. Version lag
was proposed as the clean mechanical alternative. It is cleanly EXTRACTABLE; what it
means is not. Classifying 315 citations by the words immediately before them left
**244 (77%) unclassified**, because a cited version can be either of two opposite
things:

    "deployed in chassis v1.0.1029"        → a permanent historical fact. Never expires.
    "both replicas of v1.0.1218 return X"  → a verification pinned to a version. Expires.

No pattern separates those. But the register has STRUCTURE that does not depend on
prose: its fixed bullet vocabulary. `status:` and `status-evidence:` are, by the
register's own convention, claims about the CURRENT state of the world; `what:`,
`why it exists:`, `sources:` are description and history. Keying on the field cuts
315 citations to 206 and needs no language understanding at all.

THE ONE SIGNAL THAT IS SHARP, and it is sharp because of a fact about the fleet:
a citation that quotes an AGENT ROW's container image tag as current-state evidence
is stale by construction. Measured 2026-08-12: **all 187 live `agent_definitions`
rows carry the live tag** (`v1.0.1290`), uniformly — the release rewrites them, so a
tag quoted from a row dates the OBSERVATION and never describes the row for long.
[INFERRED: the uniformity is measured; the rewriting mechanism is inferred from it.]
That is why those citations are listed separately and can be called expired outright.

CONTROLS PRINTED EVERY RUN, because a report that ranks by lag always produces a
ranking and that is not evidence of anything:
  - the newest citation in the register must show a lag near 0 (if everything looks
    stale, the live version was resolved wrongly — suspect that first);
  - the field-keying must visibly EXCLUDE citations (315 → 206), or the key is not
    doing any work;
  - the live version is stated WITH ITS SOURCE. A wrong live version makes every lag
    in the report wrong in the same direction, which is unfalsifiable from inside.

USAGE:
    scripts/report-register-version-lag.py             # summary + the sharp list
    scripts/report-register-version-lag.py --worklist  # + the oldest current-state citations
    scripts/report-register-version-lag.py --top 40    # how many worklist rows (default 25)
"""
import collections
import glob
import os
import re
import subprocess
import sys

REG = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                   "docs/agent_docs/docs026_concept_register/register")
MAKEFILE = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "makefile")

CITE = re.compile(r"v1\.0\.(\d{3,4})")
ENTRY = re.compile(r"^### ([A-Z]{2,4}-\d{3})\b")
FIELD = re.compile(r"^\s*-\s+\*\*([^:*]{2,40})")
CURRENT_STATE = ("status", "status-evidence")

# A container image tag quoted DIRECTLY. Tested by adjacency, not by proximity: strip
# the punctuation between the version and whatever precedes it, and require what is
# left to END with an image token. A window-based regex ("image within 12 chars")
# over-fires badly — it read "inert until an image roll AND … (2026-08-09" (BIZ-031)
# and "make deploy-… IMAGE_TAG=v1.0.1190" (DMR-002) as live image evidence, giving 2
# of 7 precision. Measured on this register: the adjacency form is 3 of 3.
GLUE = re.compile(r"[\s:`'\"(\[]+$")
IMAGE_TOKEN = re.compile(r"(?:^|[\s`'\"(])(?:image|image_tag|chassis image|docker\.io/aqls/[\w.-]+)$", re.I)


def quotes_image_tag(line, at):
    """Does the version at offset `at` sit immediately after an image reference?"""
    return bool(IMAGE_TOKEN.search(GLUE.sub("", line[:at])))


def live_version():
    """(number, human-readable source). Cluster first; the makefile is the fallback.

    Stated with its source because a wrong live version biases every lag in the
    report the same way, and nothing inside the report could reveal it.
    """
    try:
        p = subprocess.run(
            ["kubectl", "-n", "ai-persona-system", "get", "deploy", "-o",
             "jsonpath={range .items[*]}{.spec.template.spec.containers[0].image}{\"\\n\"}{end}"],
            capture_output=True, text=True, timeout=30)
        tags = {m.group(1) for m in CITE.finditer(p.stdout or "")}
        if tags:
            n = max(int(t) for t in tags)
            extra = f", {len(tags)} distinct tags live" if len(tags) > 1 else ", uniform"
            return n, f"the cluster (highest of the running deployments{extra})"
    except Exception:
        pass
    try:
        for line in open(MAKEFILE):
            if line.startswith("IMAGE_TAG"):
                m = CITE.search(line)
                if m:
                    return int(m.group(1)), "makefile IMAGE_TAG (cluster unreachable — may be AHEAD of live)"
    except OSError:
        pass
    return None, "UNRESOLVED"


def citations():
    """One record per version citation: entry id, field it sits in, lag input, line."""
    for fn in sorted(glob.glob(os.path.join(REG, "*.md"))):
        if os.path.basename(fn).startswith("000_"):        # the index, not entries
            continue
        entry, field = None, "(prose, no bullet)"
        for line in open(fn, errors="replace"):
            m = ENTRY.match(line)
            if m:
                entry, field = m.group(1), "(prose, no bullet)"
            f = FIELD.match(line)
            if f:
                k = f.group(1).strip().lower()
                field = ("status-evidence" if k.startswith("status-evidence")
                         else "status" if k.startswith("status") else k)
            for c in CITE.finditer(line):
                # Quote the CITATION's own surroundings, never the line's head. A line
                # can carry several versions, and printing its first 200 chars shows a
                # different citation from the one that matched — which read as three
                # false positives during this script's own review when the detector was
                # in fact 9 for 9. The evidence shown must be the evidence tested.
                ctx = " ".join(line[max(0, c.start() - 62):c.end() + 42].split())
                yield (entry, field, int(c.group(1)), os.path.basename(fn),
                       quotes_image_tag(line, c.start()), "…" + ctx + "…")


def main():
    top = 25
    if "--top" in sys.argv:
        top = int(sys.argv[sys.argv.index("--top") + 1])

    live, src = live_version()
    rows = list(citations())
    if not rows:
        print("no version citations found — check REG path")
        return 1
    if live is None:
        print("could not resolve the live version from the cluster or the makefile.")
        print("Every lag would be unfalsifiable, so nothing is reported. Fix that first.")
        return 1

    lag = lambda r: live - r[2]
    cur = [r for r in rows if r[1] in CURRENT_STATE]
    images = sorted([r for r in cur if r[4]], key=lag, reverse=True)

    print(f"live chassis version: v1.0.{live}   ← {src}")
    print(f"register: {len(rows)} version citations, {len({r[0] for r in rows if r[0]})} entries\n")

    print("by the FIELD the citation sits in (the register's own vocabulary, no prose parsed):")
    byf = collections.Counter(r[1] for r in rows)
    for f, n in byf.most_common(8):
        L = sorted(lag(r) for r in rows if r[1] == f)
        mark = " ← current-state claim" if f in CURRENT_STATE else ""
        print(f"   {n:4d}  {f:24s} median lag {L[len(L)//2]:4d}   50+ behind {sum(1 for x in L if x >= 50):3d}{mark}")

    print(f"\nCURRENT-STATE citations ({' + '.join(CURRENT_STATE)}): {len(cur)} "
          f"across {len({r[0] for r in cur if r[0]})} entries")
    print(f"   50+ versions behind: {sum(1 for r in cur if lag(r) >= 50)}")
    print("   ⚠ a large lag does NOT mean the entry is wrong. 77% of citations cannot be")
    print("     classified as fact-vs-verification by any pattern — read the line, then check.")

    print(f"\nA CONTAINER IMAGE TAG QUOTED DIRECTLY, as current-state evidence ({len(images)}):")
    print("   Every live agent_definitions row carries the live tag (187/187 on 2026-08-12), so a")
    print("   tag read off a LIVE ROW dates the observation and expires on the next release.")
    print("   ⚠ What this cannot tell you is WHICH artefact holds the tag, and the two differ:")
    print("     a live row or CronJob image  → expired, re-read it;  a repo seed file or a")
    print("     recorded command → permanent, leave it. The check is one line —")
    print("     `SELECT image_tag FROM agent_definitions WHERE type='<x>'` vs `grep` the seed.")
    for r in images:
        print(f"   [{lag(r):4d} behind] {r[0] or '?':10s} {r[1]:16s} {r[3]}")
        print(f"                 {r[5]}")

    if "--worklist" in sys.argv:
        rest = sorted([r for r in cur if not r[4]], key=lag, reverse=True)[:top]
        print(f"\nWORKLIST — oldest current-state citations, not image-tag ({len(rest)} of "
              f"{len([r for r in cur if not r[4]])} shown). Read the line; each is cheap to check:")
        for r in rest:
            print(f"   [{lag(r):4d} behind] {r[0] or '?':10s} {r[1]:16s} {r[3]}")
            print(f"                 {r[5]}")

    newest = min(rows, key=lag)
    print(f"\nCONTROLS")
    print(f"   newest citation in the register: v1.0.{newest[2]} (lag {lag(newest)}) in "
          f"{newest[0] or '?'} — must be near 0, or the live version is resolved wrongly")
    print(f"   field-keying excluded {len(rows) - len(cur)} of {len(rows)} citations "
          f"({100*(len(rows)-len(cur))//len(rows)}%) — if this is 0 the key is doing no work")
    print(f"   live version source: {src}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
