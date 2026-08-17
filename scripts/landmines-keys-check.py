#!/usr/bin/env python3
"""Assert the doc_notes subject_keys MATCH the footprints in LANDMINES.md.

WHY THIS EXISTS SEPARATELY FROM `landmines-sync.py --check`: that flag's drift
test is `new or gone` — whether the entry's SOURCE key is present at all. An
entry whose FOOTPRINTS changed passes it while its `subject_key` rows are stale,
and `subject_key` is the whole point of the corpus: it is what a later session
greps and what the SessionStart hook matches against a path. On 2026-08-14 the
splitter fix re-keyed 185 of 482 entries, six of them at an UNCHANGED row count —
so neither `--check` nor a count comparison could see the repair. This asserts
identity, which is the only thing that could have.

    ./scripts/landmines-keys-check.py            # exit 0 in sync, 1 on mismatch
    ./scripts/landmines-keys-check.py --file X   # parse X instead (mutation tests)
    ./scripts/landmines-keys-check.py --self-test # prove it can FAIL

Read-only. Never writes a row and never edits LANDMINES.md (owner ruling D10:
the markdown is the system of record).
"""

import argparse
import os
import subprocess
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from landmines_lib import LANDMINES, parse  # noqa: E402

SOURCE_PREFIX = "LANDMINES.md#"
SEP = "\x1f"  # unit separator: cannot occur in a footprint
PSQL = os.environ.get("PSQL_CMD") or (
    "kubectl exec -i -n ai-persona-system postgres-clients-0 -- "
    "psql -U clients_user -d clients_db"
)


def db_keys():
    """{source: sorted subject_keys}. Aggregated in SQL — one round trip."""
    sql = (
        f"SELECT source, string_agg(subject_key, E'\\x1f' ORDER BY subject_key) "
        f"FROM doc_notes WHERE source LIKE '{SOURCE_PREFIX}%' GROUP BY 1;"
    )
    proc = subprocess.run(
        PSQL.split() + ["-v", "ON_ERROR_STOP=1", "-t", "-A"],
        input=sql.encode("utf-8"), capture_output=True,
    )
    if proc.returncode != 0:
        sys.exit("psql failed:\n" + proc.stderr.decode("utf-8", "replace"))
    out = proc.stdout.decode("utf-8", "replace")
    got = {}
    for line in out.splitlines():
        if "|" in line:
            src, keys = line.split("|", 1)
            got[src.strip()] = sorted(keys.strip().split(SEP))
    return got


def compare(path=None):
    want = {SOURCE_PREFIX + e["slug"]: sorted(e["footprints"]) for e in parse(path or LANDMINES)}
    have = db_keys()
    mismatched = {s: (have[s], want[s]) for s in want if s in have and have[s] != want[s]}
    missing = [s for s in want if s not in have]
    strays = [s for s in have if s not in want]
    midots = [k for keys in have.values() for k in keys if "·" in k]
    return want, have, mismatched, missing, strays, midots


def report(path=None):
    want, have, mismatched, missing, strays, midots = compare(path)
    print(f"LANDMINES.md: {len(want)} entr(ies)   doc_notes: {len(have)} entr(ies), "
          f"{sum(len(v) for v in have.values())} row(s)")
    for s, (got, exp) in list(mismatched.items())[:10]:
        print(f"  KEY MISMATCH  {s[:70]}")
        print(f"      db   : {got}")
        print(f"      file : {exp}")
    if len(mismatched) > 10:
        print(f"  … and {len(mismatched) - 10} more")
    for s in missing[:10]:
        print(f"  NOT IN DB     {s[:70]}   (run landmines-verify-dispatch.sh)")
    for s in strays[:10]:
        print(f"  STRAY ROW     {s[:70]}   (entry retitled or removed)")
    if midots:
        print(f"  ⚠ {len(midots)} subject_key(s) still contain '·' — the splitter "
              f"regressed; see landmines_lib.split_footprints and its self-test")
    bad = len(mismatched) + len(missing) + len(strays) + len(midots)
    print("\nin sync — every subject_key matches its footprint" if not bad
          else f"\nOUT OF SYNC: {bad} problem(s). Fix with ./scripts/landmines-verify-dispatch.sh")
    return 1 if bad else 0


def self_test():
    """Prove the check can FAIL: parse a MUTATED copy and require a mismatch."""
    import tempfile
    src = open(LANDMINES, encoding="utf-8").read()
    # Mutate the first footprint line BELOW the '# Entries' marker. parse() starts
    # there, so a mutation above it is a NO-OP — and this self-test caught exactly
    # that on its first run: the preamble documents the format and its example
    # footprint line is the first in the file. A mutation that cannot register
    # makes the whole self-test pass vacuously.
    from landmines_lib import ENTRIES_MARKER
    base = src.index(ENTRIES_MARKER)
    marker = "- **footprint:**"
    i = src.index(marker, base)
    end = src.index("\n", i)
    mutated = src[:end] + ", scripts/DELIBERATELY_NOT_A_REAL_FOOTPRINT.py" + src[end:]
    with tempfile.NamedTemporaryFile("w", suffix=".md", delete=False) as fh:
        fh.write(mutated)
        tmp = fh.name
    try:
        clean = compare()[2:]
        clean_bad = sum(len(x) for x in clean)
        mut = compare(tmp)[2:]
        mut_bad = sum(len(x) for x in mut)
        print(f"  unmutated corpus : {clean_bad} problem(s)  (expect 0)")
        print(f"  mutated corpus   : {mut_bad} problem(s)  (expect >0 — one added footprint)")
        ok = clean_bad == 0 and mut_bad > 0
        print("PASS — the check discriminates" if ok else
              "FAIL — a mutation did not register; this check cannot fail and proves nothing")
        return 0 if ok else 1
    finally:
        os.unlink(tmp)


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--file", help="parse this file instead of LANDMINES.md")
    ap.add_argument("--self-test", action="store_true", help="prove the check can fail")
    a = ap.parse_args()
    sys.exit(self_test() if a.self_test else report(a.file))
