#!/usr/bin/env python3
"""compare_rebuilt.py — diff each REBUILT mortgagecalculator tool against the
golden captured from its hand-built ORIGINAL.

The golden (GOLDEN_2026-08-05_original_tools.json) is keyed by the original
file-form URLs; the recreations deploy to directory-form URLs. toolgolden's
--compare requires the same key on both sides, so this wrapper carries the
old→new mapping and reuses toolgolden's Runner / VECTORS / numeric_diff —
a second implementation of "did the numbers move" would be a second thing that
can be subtly wrong (verify_rewrite.py's argument, verbatim).

READ THE DIFF WITH THIS IN MIND: the rebuilt pages are LLM-written and the
recreate prompt carries the original markup but no id-preservation contract,
so RENAMED ids show up as one-sided diffs ("value -> None" / "None -> value").
A renamed id is a fence-authoring problem (the fence must name the NEW id); a
CHANGED NUMBER under an id present on both sides is an arithmetic divergence
and blocks adoption of that tool. The two failures look similar in the raw
diff and mean completely different things.

Usage:
  python3 compare_rebuilt.py                  # all 12
  python3 compare_rebuilt.py repayment simple # by original slug
"""
import json
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
LC_LANE = os.path.join(HERE, "..", "..", "loancalculator_couk")
sys.path.insert(0, os.path.abspath(LC_LANE))
from toolgolden import Runner, VECTORS, numeric_diff  # noqa: E402

GOLDEN = os.path.join(HERE, "GOLDEN_2026-08-05_original_tools.json")
BASE = "https://mortgagecalculator.co.uk"

# original file-form URL -> rebuilt directory-form URL (pages.url per pages table)
MAPPING = {
    "/repayment.html":       "/tools/repayment/index.html",
    "/affordability.html":   "/tools/affordability/index.html",
    "/simple.html":          "/tools/simple/index.html",
    "/overpayment.html":     "/tools/overpayment/index.html",
    "/stamp-duty.html":      "/tools/stamp-duty/index.html",
    "/bridging-loan.html":   "/tools/bridging-loan/index.html",
    "/equity-release.html":  "/tools/equity-release/index.html",
    "/fee-analyser.html":    "/tools/fee-analyser/index.html",
    "/rate-forecaster.html": "/tools/rate-forecaster/index.html",
    "/portfolio.html":       "/tools/portfolio/index.html",
    "/investor.html":        "/investor/index.html",
    "/fact-finder.html":     "/games/fact-finder/index.html",
}


def main():
    only = set(sys.argv[1:])
    golden = json.load(open(GOLDEN))["pages"]
    r = Runner()
    bad = 0
    try:
        for old, new in sorted(MAPPING.items()):
            slug = old.strip("/").replace(".html", "")
            if only and slug not in only:
                continue
            g = golden.get(BASE + old)
            if not g:
                print("NO GOLDEN  %s" % old)
                bad += 1
                continue
            try:
                n = r.capture(BASE + new)
            except Exception as e:  # a 404 rebuild is a finding, not a crash
                print("CAPTURE-ERROR  %-28s %s" % (slug, str(e)[:80]))
                bad += 1
                continue
            diffs = []
            vecs = [v for v, _ in VECTORS if v in g and v in n]
            for vec in vecs:
                for phase in ("after_input", "after_press"):
                    d = numeric_diff(g[vec][phase].get("ids", {}),
                                     n[vec][phase].get("ids", {}))
                    if d:
                        diffs.append("   %s / %s" % (vec, phase))
                        diffs.extend(d)
            if diffs:
                bad += 1
                print("\nDIVERGED  %s -> %s" % (old, new))
                print("\n".join(diffs[:40]))
            else:
                print("MATCHES   %-28s (%d vectors)" % (slug, len(vecs)))
    finally:
        r.close()
    if bad:
        print("\n%d tool(s) diverged / uncapturable" % bad)
        sys.exit(1)
    print("\nall compared tools reproduce the original answers exactly")


if __name__ == "__main__":
    main()
