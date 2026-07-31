#!/usr/bin/env python3
"""check_register.py — the overlap guard the platform does not have.

Measured 2026-07-31: there is NO cross-site duplicate-content or topical-overlap
machinery in this platform. check_content_duplication is single-site, and
cross_site_contamination detects company-name bleed, not topical overlap. So if two
register entries quietly claim the same ground, nothing downstream will ever notice —
the sites simply compete. This script is that missing check, run against the REGISTER
(where the claim is made), not the sites (where it is too late).

It parses the ```claims``` block — the machine-readable contract at the bottom of the
register — NOT the prose. The first version parsed the prose and mis-read shorthand like
"rate(s).co.uk/.uk"; per the fleet lesson (fixing-a-checker-to-agree-with-a-broken-site),
the fix was to make the ARTEFACT checkable, not to make the checker guess.

Checks:
  1. every claims-table domain is in PORTFOLIO_domains.txt, and vice versa — an
     unassigned domain means a lane will invent a direction; an unlisted one is a typo
  2. no domain is claimed by two entries
  3. every claims entry id has a matching '### <id> ' prose entry, and vice versa
  4. every prose entry carries a neighbours line (or explicit HOLD) — an entry with no
     neighbours has not said where its ground ends, which is how convergence starts
  5. no two entries share a (family × audience × mode) coordinate

Run:  python3 check_register.py          # exits 1 on any violation
"""
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
text = open(os.path.join(HERE, "REGISTER_positioning.md"), encoding="utf-8").read()
fails = []

# ── 1+2: the claims block against the portfolio list ─────────────────────────────
m = re.search(r"```claims\n(.*?)```", text, re.S)
if not m:
    sys.exit("no ```claims``` block — the register has lost its machine-readable half")
claims = {}          # domain -> entry id
entry_ids = []
for line in m.group(1).splitlines():
    if not line.strip():
        continue
    eid, _, doms = line.partition(":")
    eid = eid.strip()
    entry_ids.append(eid)
    for d in doms.split():
        if d in claims:
            fails.append(f"domain claimed twice: {d} by {claims[d]} and {eid}")
        claims[d] = eid

listed = {l.strip() for l in open(os.path.join(HERE, "PORTFOLIO_domains.txt"))
          if l.strip() and not l.startswith("#")}
for d in sorted(listed - set(claims)):
    fails.append(f"portfolio domain NOT claimed by any entry: {d}")
for d in sorted(set(claims) - listed):
    fails.append(f"claims table names a domain not in the portfolio list: {d}")

# ── 3: claims ids <-> prose entries ──────────────────────────────────────────────
prose = dict(re.findall(r"\n### ([A-Z]+\d+) — [^\n]*\n(.*?)(?=\n### |\n---|\Z)", text, re.S))
for eid in entry_ids:
    if eid not in prose:
        fails.append(f"claims id {eid} has no prose entry")
for eid in prose:
    if eid not in entry_ids:
        fails.append(f"prose entry {eid} missing from the claims table")

# ── 4+5: neighbours + coordinate uniqueness ──────────────────────────────────────
coords = {}
for eid, body in prose.items():
    if "neighbours:" not in body and "HOLD" not in body and "excluded from every collision" not in body:
        fails.append(f"entry {eid} has no neighbours line and is not HOLD — "
                     f"where does its ground end?")
    fam = eid.rstrip("0123456789")
    aud = re.search(r"\*\*audience:\*\*\s*([^·\n]+)", body)
    mode = re.search(r"\*\*mode(?:[^:*]*)?:\*\*\s*([^·\n]+)", body)
    if aud and mode:
        key = (fam, aud.group(1).strip().lower()[:40], mode.group(1).strip().lower()[:30])
        if key in coords:
            fails.append(f"coordinate collision: {eid} and {coords[key]} both claim {key}")
        coords[key] = eid

print(f"portfolio {len(listed)} · claimed {len(claims)} · entries {len(entry_ids)} "
      f"(prose {len(prose)}) · coordinates {len(coords)}")
if fails:
    print(f"\n{len(fails)} VIOLATION(S):")
    for f in fails:
        print("  FAIL ", f)
    sys.exit(1)
print("register invariants hold")
