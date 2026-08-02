#!/usr/bin/env python3
"""verify_shipped.py — the scheduled test of assemble_mirror.py, run at the wire.

WHAT THIS IS FOR. `assemble_mirror.py` is a Python reimplementation of the Go
assembler, and `verify_assembled.py` checks the decomposition against it. Those
two agree with each other by construction: if the mirror is wrong about how
assembly works, every one of the harness's 27/27 passes is wrong in the same
direction and nothing in the offline suite can notice.

So the mirror is treated as a hypothesis with ONE test, and this is it. A page is
decomposed, the REAL Go path renders and deploys it, and the served bytes are
diffed against the prediction the mirror wrote before any row was touched. Until
that diff is clean, the offline result means "consistent with my model of the
assembler", not "correct".

A CLEAN DIFF IS A STRONG RESULT AND A DIRTY ONE IS NOT A DISASTER. The likely
differences are in the parts hardest to mirror — JSON-LD key order and HTML
escaping, the exact newlines around `<main>`, and `repairOutboundPageLinks`,
which is deliberately NOT mirrored because it is a repair rather than a
transform. What matters is reading the diff and deciding which side is wrong,
before 26 more pages are written on the strength of it.

⚠ SET A USER-AGENT. Cloudflare fronts these zones and answers `Python-urllib`
with 403 whatever the method, so a checker without one reports a healthy page as
unreachable and its output is indistinguishable from a failed deploy.

Usage:  DECOMP_WORK=... python3 verify_shipped.py <page-name> [...]
        DECOMP_WORK=... python3 verify_shipped.py --all
"""
import difflib
import json
import os
import sys
import urllib.error
import urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
LANE = os.path.dirname(HERE)
DOMAIN = "loancalculator.co.uk"
UA = "Mozilla/5.0 (compatible; loancalculator-shipped-check/1.0)"


def fetch(url):
    req = urllib.request.Request(url, headers={"User-Agent": UA})
    try:
        with urllib.request.urlopen(req, timeout=30) as r:
            return r.status, r.read().decode("utf-8", "replace")
    except urllib.error.HTTPError as e:
        return e.code, ""
    except Exception as e:  # noqa: BLE001
        return 0, str(e)


def main():
    work = os.environ.get("DECOMP_WORK")
    if not work:
        sys.exit("set DECOMP_WORK")
    pred_dir = os.path.join(work, "predicted")
    pages = {}
    for line in open(os.path.join(work, "pages.txt"), encoding="utf-8"):
        n, u = line.rstrip("\n").split("|")
        pages[n] = u

    names = [a for a in sys.argv[1:] if not a.startswith("--")]
    if "--all" in sys.argv:
        names = sorted(n for n in pages
                       if os.path.exists(os.path.join(pred_dir, n + ".html")))
    if not names:
        sys.exit("name a page, or --all")

    stored = json.load(open(os.path.join(work, "manifest.json"), encoding="utf-8"))
    bad = 0
    for name in names:
        pred_path = os.path.join(pred_dir, name + ".html")
        if not os.path.exists(pred_path):
            print("NO PREDICTION  %s — run load_decomposition.py first" % name)
            bad += 1
            continue
        predicted = open(pred_path, encoding="utf-8").read()
        code, live = fetch("https://%s%s" % (DOMAIN, pages[name]))

        if code != 200:
            print("HTTP %-4s      %-32s %s" % (code, name, pages[name]))
            bad += 1
            continue

        if live == predicted:
            print("EXACT          %-32s %d bytes — the mirror predicted the "
                  "served bytes" % (name, len(live)))
            continue

        # Not exact. Say HOW different before saying whether it matters: a
        # two-line JSON-LD ordering difference and a missing calculator look the
        # same in a boolean.
        pl, ll = predicted.splitlines(), live.splitlines()
        diff = list(difflib.unified_diff(pl, ll, "predicted", "served", n=1, lineterm=""))
        added = sum(1 for d in diff if d.startswith("+") and not d.startswith("+++"))
        removed = sum(1 for d in diff if d.startswith("-") and not d.startswith("---"))
        print("DIFFERS        %-32s predicted %d b / served %d b, +%d -%d line(s)"
              % (name, len(predicted), len(live), added, removed))

        # The two questions that decide whether this is cosmetic.
        fn = stored[name]["tool_function"]
        if fn:
            marker = 'class="tool-%s-section"' % fn[len("tool-"):] if fn.startswith("tool-") else fn
            print("   tool section present in served page: %s"
                  % ("YES" if marker in live else "NO  <-- this is not cosmetic"))
        print("   footer present: %s   nav links: %d"
              % ("yes" if "site-footer" in live else "NO",
                 live.count('class="dropdown-content"')))
        for line in diff[:40]:
            print("   " + line[:160])
        if len(diff) > 40:
            print("   ... %d more diff line(s)" % (len(diff) - 40))
        bad += 1

    print()
    if bad:
        print("%d of %d page(s) did not match the mirror's prediction — READ THE "
              "DIFF and decide which side is wrong before writing any more rows."
              % (bad, len(names)))
        return 1
    print("all %d page(s) served EXACTLY what the mirror predicted — the mirror is "
          "validated for this page shape, and the offline 27/27 can be read as a "
          "result rather than a hypothesis." % len(names))
    return 0


if __name__ == "__main__":
    sys.exit(main())
