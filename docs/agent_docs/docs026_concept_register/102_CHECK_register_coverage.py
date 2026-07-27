#!/usr/bin/env python3
"""
102_CHECK_register_coverage.py — does the concept register still cover the estate?

WHY THIS EXISTS (bugs_open/106)
  The register froze on 2026-07-13. Three subsystems have since been found missing
  — fixloop (07-16), model-directory (07-17), claims-verification (07-27) — and
  ALL THREE were found by coincidence, because somebody happened to be working
  beside the hole. Three coincidental detections of one failure mode is a missing
  detector, not bad luck.

  This is that detector. It answers one question: what exists in the estate that
  the register has never heard of?

SHAPE — sensor + ratchet, copied deliberately from
`platform/orchestration/actions/discovery_checks/verifier_coverage_test.go`:
  * the SENSOR enumerates observable subsystems from the filesystem;
  * the RATCHET (`102_coverage_ratchet.txt`) lists what is KNOWN uncovered and
    accepted for now;
  * only the difference is news. Everything already on the ratchet is backlog,
    not an alarm.

  That is what stops a coverage report becoming wallpaper: on a healthy day it
  prints nothing, so the day it prints something you read it.

ADVISORY BY DESIGN. It exits 0 even when it finds drift, unless --strict.
Breadth of uncovered work is a backlog, not an error, and a check that blocks on
a backlog gets disabled within a week.

READ-ONLY. Filesystem and git only — no cluster, no DB, no network. Safe to run
anywhere, any time, by anyone.

USAGE
  ./102_CHECK_register_coverage.py                  # report drift vs the ratchet
  ./102_CHECK_register_coverage.py --all            # every uncovered subsystem
  ./102_CHECK_register_coverage.py --update-ratchet # accept current state as baseline
  ./102_CHECK_register_coverage.py --strict         # exit 1 on new drift (for CI)

WHAT IT DELIBERATELY DOES NOT DO
  It does not judge whether an existing entry is ACCURATE — that is stage-2
  verification's job and it is far more expensive. It only asks whether a
  subsystem is REPRESENTED AT ALL. Those are different questions and conflating
  them is how a coverage check turns into an audit nobody runs.
"""

import argparse
import re
import subprocess
import sys
from datetime import date
from pathlib import Path

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
REGISTER = HERE / "register"
RATCHET = HERE / "102_coverage_ratchet.txt"
WORKSTREAMS = REPO / "docs/agent_docs/docs024_key_docs_latest"
FREEZE = "2026-07-13"

# Directories that are not subsystems: archives, scratch, per-site content.
SKIP_DIRS = {"multi_session_coordination"}


def register_text() -> str:
    """Every register file concatenated once — the corpus we test membership against."""
    parts = []
    for f in sorted(REGISTER.glob("*.md")):
        parts.append(f.read_text(errors="replace").lower())
    return "\n".join(parts)


def workstream_dirs():
    """Sensor: every workstream directory, with the date it first appeared in git."""
    out = []
    for d in sorted(p for p in WORKSTREAMS.iterdir() if p.is_dir()):
        if d.name in SKIP_DIRS:
            continue
        try:
            first = subprocess.run(
                ["git", "log", "--reverse", "--format=%ad", "--date=short", "--", str(d)],
                cwd=REPO, capture_output=True, text=True, timeout=30,
            ).stdout.split("\n")[0].strip()
        except Exception:
            first = ""
        out.append((d.name, first))
    return out


def is_covered(name: str, corpus: str) -> bool:
    """
    A subsystem counts as covered if its name — or a loosened form of it — appears
    anywhere in the register. Loose on purpose: a false 'covered' is cheap (one
    already-known subsystem stays quiet), a false 'uncovered' is expensive (it
    trains the reader to ignore the report).
    """
    cands = {name, name.replace("_", "-"), name.replace("_", " ")}
    # bugfix_003_spawn_loss -> "spawn loss"; drop leading bug/bugfix/NNN tokens
    stripped = re.sub(r"^(bugfix|bug|feature)[_-]?\d*[_-]?", "", name)
    stripped = re.sub(r"^\d+[_-]", "", stripped)
    if stripped:
        cands |= {stripped, stripped.replace("_", "-"), stripped.replace("_", " ")}
    return any(c and c.lower() in corpus for c in cands)


def read_ratchet() -> set:
    if not RATCHET.exists():
        return set()
    return {
        ln.strip()
        for ln in RATCHET.read_text().splitlines()
        if ln.strip() and not ln.startswith("#")
    }


def stamps():
    """covers-through dates, so the report can say how old the map is."""
    out = []
    for f in sorted(REGISTER.glob("*.md")):
        if f.name == "000_concept_index.md":
            continue
        m = re.search(r"covers-through:\s*(\d{4}-\d{2}-\d{2})", f.read_text(errors="replace"))
        out.append((f.name, m.group(1) if m else None))
    return out


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--all", action="store_true", help="list every uncovered subsystem, not just new drift")
    ap.add_argument("--update-ratchet", action="store_true", help="accept the current state as the baseline")
    ap.add_argument("--strict", action="store_true", help="exit 1 when there is NEW drift")
    args = ap.parse_args()

    if not REGISTER.is_dir():
        print(f"register not found at {REGISTER}", file=sys.stderr)
        return 2

    corpus = register_text()
    ws = workstream_dirs()
    uncovered = [(n, d) for n, d in ws if not is_covered(n, corpus)]
    ratchet = read_ratchet()
    new = [(n, d) for n, d in uncovered if n not in ratchet]

    st = stamps()
    unstamped = [f for f, d in st if d is None]
    oldest = min((d for _, d in st if d), default=None)

    print("── concept register coverage ─────────────────────────────────────────")
    print(f"  register files      : {len(st)}")
    print(f"  workstreams on disk : {len(ws)}")
    print(f"  post-freeze ({FREEZE}) : {sum(1 for _, d in ws if d and d > FREEZE)}")
    print(f"  uncovered           : {len(uncovered)}   (ratchet accepts {len(ratchet)})")
    if oldest:
        print(f"  oldest covers-through: {oldest}")
    if unstamped:
        print(f"  ⚠ unstamped files   : {len(unstamped)} — {', '.join(unstamped[:5])}")

    if args.update_ratchet:
        RATCHET.write_text(
            "# Concept-register coverage ratchet — subsystems known to be uncovered.\n"
            "# Generated by 102_CHECK_register_coverage.py --update-ratchet.\n"
            "# A name here is ACCEPTED BACKLOG, not a resolved problem. Remove a line\n"
            "# when you add its register entry; the report then stays quiet about it.\n"
            f"# Baseline taken {date.today().isoformat()}.\n\n"
            + "\n".join(n for n, _ in sorted(uncovered)) + "\n"
        )
        print(f"\n  ratchet written: {len(uncovered)} accepted → {RATCHET.name}")
        return 0

    show = uncovered if args.all else new
    label = "uncovered" if args.all else "NEW since the ratchet"
    if show:
        print(f"\n  {len(show)} {label}:")
        for n, d in sorted(show, key=lambda x: (x[1] or "", x[0])):
            flag = "  ← post-freeze" if d and d > FREEZE else ""
            print(f"    {d or '????-??-??'}  {n}{flag}")
        if not args.all and uncovered:
            print(f"\n  ({len(uncovered) - len(new)} more are on the ratchet — run --all to see them.)")
    else:
        print(f"\n  no {label}.")

    print("\n  Absence from the register is not absence from the platform — that is the")
    print("  whole point of this check. See bugs_open/106.")

    return 1 if (args.strict and new) else 0


if __name__ == "__main__":
    sys.exit(main())
