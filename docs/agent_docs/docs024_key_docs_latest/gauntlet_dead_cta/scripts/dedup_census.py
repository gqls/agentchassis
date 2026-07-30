#!/usr/bin/env python3
"""Cross-page content duplication census — the bugs_open/151 method, run on any site.

Two halves, because 151's central finding is that EITHER ALONE MISSES IT:

  (a) FACT census — which approved facts does each section assert? Two sections
      can restate the identical facts while being only 18% textually similar
      (measured on fundamentallyai.com), so similarity alone reports "fine".
  (b) TEXT similarity — difflib.SequenceMatcher over section prose. Catches
      repeated copy that carries no countable fact at all, which the fact
      census cannot see.

Usage: dedup_census.py <sections.json> <evidence_base.json> [min_shared_facts]
"""
import json
import re
import sys
from difflib import SequenceMatcher
from itertools import combinations

SECTIONS = json.load(open(sys.argv[1]))
EVIDENCE = json.load(open(sys.argv[2]))
MIN_SHARED = int(sys.argv[3]) if len(sys.argv) > 3 else 2

FACTS = EVIDENCE.get("facts") or []

# ---------- text extraction ----------
SKIP_KEYS = re.compile(r"(url|href|src|image|icon|slug|id|class|target|colour|color)$", re.I)


def section_text(data):
    """Concatenate human-readable string values, skipping URLs/ids/asset paths."""
    out = []

    def walk(v, key=""):
        if isinstance(v, str):
            if SKIP_KEYS.search(key or ""):
                return
            if v.startswith(("/", "http", "#")) or len(v) < 3:
                return
            out.append(v)
        elif isinstance(v, dict):
            for k, vv in v.items():
                walk(vv, k)
        elif isinstance(v, list):
            for vv in v:
                walk(vv, key)

    walk(data)
    return " ".join(out)


def norm(t):
    t = re.sub(r"<[^>]+>", " ", t)
    return re.sub(r"\s+", " ", t).strip().lower()


for s in SECTIONS:
    s["text"] = norm(section_text(s.get("data") or {}))
    s["label"] = f"{s['page']}/{s.get('slot') or '?'}@{s.get('pos')}"

# ---------- (a) fact census ----------
# A section "asserts" a fact if its value appears as a standalone token AND at
# least one distinguishing content word from the claim appears in the section.
# The value alone is far too loose ("3" matches anything).
STOP = set("the a an of in on and or to for with is are as at by from that this "
           "it its into over under about your you we our".split())


def claim_words(f):
    words = re.findall(r"[a-z]{4,}", str(f.get("claim", "")).lower())
    return [w for w in words if w not in STOP]


def asserts(sec, f):
    val = str(f.get("value", "")).strip()
    if not val:
        return False
    if not re.search(rf"(?<![\d.]){re.escape(val)}(?![\d.])", sec["text"]):
        return False
    return any(w in sec["text"] for w in claim_words(f)) if claim_words(f) else True


for s in SECTIONS:
    s["facts"] = [f["id"] for f in FACTS if asserts(s, f)]

print(f"=== FACT CENSUS — {len(FACTS)} approved facts as the ruler, "
      f"{len(SECTIONS)} sections ===")
carriers = [s for s in SECTIONS if s["facts"]]
print(f"sections asserting >=1 fact: {len(carriers)}")
from collections import Counter
dist = Counter(len(s["facts"]) for s in SECTIONS)
for n in sorted(dist):
    print(f"  {n} fact(s): {dist[n]} sections")

flagged = [s for s in SECTIONS if len(s["facts"]) >= MIN_SHARED]
print(f"\nsections asserting >= {MIN_SHARED} facts (151's duplication shape):")
if not flagged:
    print("  none")
for s in sorted(flagged, key=lambda x: -len(x["facts"])):
    print(f"  {s['label']:44s} {s['facts']}")

# fact -> which sections repeat it
print("\nper-fact repetition (a fact asserted on many pages is the 151 symptom):")
for f in FACTS:
    holders = [s["label"] for s in SECTIONS if f["id"] in s["facts"]]
    pages = sorted({h.split("/")[0] for h in holders})
    flag = "  <-- REPEATED ACROSS PAGES" if len(pages) > 1 else ""
    print(f"  {f['id']:18s} value={str(f.get('value')):>4s}  "
          f"{len(holders)} section(s) on {len(pages)} page(s){flag}")
    if len(pages) > 1:
        for h in holders:
            print(f"      {h}")

# ---------- (b) text similarity ----------
print("\n=== TEXT SIMILARITY (difflib) — pairs over 0.35, cross-page first ===")
pairs = []
for a, b in combinations([s for s in SECTIONS if len(s["text"]) > 120], 2):
    r = SequenceMatcher(None, a["text"], b["text"]).ratio()
    if r >= 0.35:
        pairs.append((r, a, b))
pairs.sort(key=lambda t: -t[0])
if not pairs:
    print("  no pair over 0.35")
for r, a, b in pairs[:25]:
    tag = "CROSS-PAGE" if a["page"] != b["page"] else "same-page "
    print(f"  {r:.2f}  {tag}  {a['label']}  <->  {b['label']}")

# ---------- (c) identical blocks ----------
print("\n=== IDENTICAL TEXT BLOCKS (same normalised text in 2+ sections) ===")
by_text = {}
for s in SECTIONS:
    if len(s["text"]) > 80:
        by_text.setdefault(s["text"], []).append(s["label"])
dupes = {t: ls for t, ls in by_text.items() if len(ls) > 1}
if not dupes:
    print("  none")
for t, ls in dupes.items():
    print(f"  {len(ls)}x  {t[:90]!r}")
    for l in ls:
        print(f"       {l}")
