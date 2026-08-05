#!/usr/bin/env python3
"""investor_golden.py — golden capture for mortgages/investor.html, which the
generic toolgolden.py CANNOT certify.

WHY THIS FILE EXISTS. toolgolden derives its vectors by scaling EVERY numeric
field's own default by the same factor (x1, x2, x0.5) — deliberately, so values
stay in-domain for any tool with no per-tool config. investor.html computes
only RATIOS of its inputs (gross yield = rent*12/price; LTV = loan/price), and
a ratio is invariant under uniform scaling, so all three vectors produce
identical outputs and toolgolden's inert-tool guard refuses to certify —
"reacts, but output is identical for every input value". Measured 2026-08-05
on the live page; the arithmetic was then read and is ratio-only, so the
refusal is the INSTRUMENT's vector scheme, not the page (the 6th adverse
verdict on this site, 5 of which have now been the instrument — see the 07-31
NOTES' harness-fault series).

THE FIX IS THE VECTOR SHAPE: move ONE field at a time. Scaling rent alone
doubles the yield; scaling loan alone doubles the LTV; no ratio survives.
Per-field staggered scalings of the page's own defaults keep the in-domain
property that motivated toolgolden's scheme.

Inherits toolgolden's hard-won session discipline verbatim: settle() waits for
a FULLY PARSED document (readyState alone certified a £0.00 golden once), and
each vector starts from a storage-cleared reload (a contaminated baseline is
perfectly self-consistent).

Usage (PYTHONPATH must include webdesign_tools_repair, as for toolgolden):
  python3 investor_golden.py --out     <golden.json> <url>
  python3 investor_golden.py --compare <golden.json> <url>
Exit 1 on divergence, or on a capture where outputs do not vary across the
staggered vectors — which after 2026-08-05 would mean the page genuinely
regressed to inert, the exact state this file exists to catch.
"""
import json
import sys
import tempfile
import time
import urllib.request

from toolprobe import CDP, start_chrome  # noqa: E402

FIELDS = ["btlPrice", "btlRent", "ltvPrice", "ltvLoan"]
OUTPUTS = ["yieldResult", "ltvResult", "ltvComment"]

# Staggered: exactly one field moves per non-baseline vector. A ratio of any
# two fields cannot be constant across all of these. ltvPrice uses x0.5 so the
# LTV crosses the script's >90% commentary threshold (225000/150000 = 150%),
# proving the comment branch is driven too, not only the arithmetic.
VECTORS = [
    ("defaults", {}),
    ("rent_x2", {"btlRent": 2.0}),
    ("btlprice_x2", {"btlPrice": 2.0}),
    ("loan_x2", {"ltvLoan": 2.0}),
    ("ltvprice_x0.5", {"ltvPrice": 0.5}),
]

DRIVE_JS = """
(function (scales) {
    if (!window.__ig_defaults) {
        window.__ig_defaults = {};
        %(fields)s.forEach(function (id) {
            var el = document.getElementById(id);
            window.__ig_defaults[id] = el ? el.value : null;
        });
    }
    var defaults = window.__ig_defaults;
    %(fields)s.forEach(function (id) {
        var el = document.getElementById(id);
        if (!el) return;
        var f = scales[id] || 1.0;
        el.value = String(parseFloat(defaults[id]) * f);
        el.dispatchEvent(new Event('input', {bubbles: true}));
        el.dispatchEvent(new Event('change', {bubbles: true}));
    });
    calcYield();
    calcLTV();
    var out = {};
    %(outputs)s.forEach(function (id) {
        var el = document.getElementById(id);
        out[id] = el ? el.innerText : null;
    });
    return JSON.stringify(out);
})(%(scales)s)
"""


def capture(url):
    import socket
    s = socket.socket()
    s.bind(("127.0.0.1", 0))
    port = s.getsockname()[1]
    s.close()
    chrome = start_chrome(port, tempfile.mkdtemp(prefix="investor-golden-"))
    try:
        req = urllib.request.Request(
            "http://127.0.0.1:%d/json/new?about:blank" % port, method="PUT")
        tab = json.loads(urllib.request.urlopen(req, timeout=10).read())
        cdp = CDP(tab["webSocketDebuggerUrl"])
        cdp.call("Runtime.enable")
        cdp.call("Page.enable")

        def ev(expr, timeout=20):
            r = cdp.call("Runtime.evaluate",
                         {"expression": expr, "returnByValue": True,
                          "awaitPromise": True}, timeout=timeout)
            return r.get("result", {}).get("result", {}).get("value")

        def settle():
            # toolgolden's rule, same reason: a drivable DOM whose inline
            # script has not parsed certifies garbage silently.
            for _ in range(80):
                if ev("document.readyState", timeout=10) == "complete":
                    break
                time.sleep(0.1)
            last = None
            for _ in range(40):
                cur = ev("document.querySelectorAll('script').length + ':' + "
                         "document.documentElement.outerHTML.length", timeout=10)
                if cur is not None and cur == last:
                    return cur
                last = cur
                time.sleep(0.15)
            return last

        out = []
        for name, scales in VECTORS:
            cdp.call("Page.navigate", {"url": url})
            settle()
            ev("try{localStorage.clear();sessionStorage.clear()}catch(e){};'cleared'",
               timeout=10)
            cdp.call("Page.navigate", {"url": url})
            shape = settle()
            raw = ev(DRIVE_JS % {"fields": json.dumps(FIELDS),
                                 "outputs": json.dumps(OUTPUTS),
                                 "scales": json.dumps(scales)})
            out.append({"name": name, "scales": scales, "dom_shape": shape,
                        "outputs": json.loads(raw) if raw else None})
        return out
    finally:
        chrome.terminate()


def main():
    if len(sys.argv) != 4 or sys.argv[1] not in ("--out", "--compare"):
        print("usage: investor_golden.py --out|--compare <golden.json> <url>")
        return 2
    mode, path, url = sys.argv[1], sys.argv[2], sys.argv[3]
    got = capture(url)

    if any(v["outputs"] is None for v in got):
        print("FAIL: a vector returned nothing — mid-parse capture or thrown "
              "JS. Not a golden. Vectors:")
        for v in got:
            print("  %-14s %s" % (v["name"], v["outputs"]))
        return 1

    distinct = {json.dumps(v["outputs"], sort_keys=True) for v in got}
    if len(distinct) < 3:
        print("FAIL: outputs do not vary across staggered vectors — the page "
              "is genuinely inert (this is the regression this file exists to "
              "catch, not an instrument artefact):")
        for v in got:
            print("  %-14s %s" % (v["name"], v["outputs"]))
        return 1

    if mode == "--out":
        with open(path, "w", encoding="utf-8") as fh:
            json.dump({"url": url, "vectors": got}, fh, indent=1)
        print("wrote %s (%d vectors, %d distinct output states)"
              % (path, len(got), len(distinct)))
        return 0

    want = json.load(open(path, encoding="utf-8"))["vectors"]
    bad = 0
    for w, g in zip(want, got):
        if w["outputs"] != g["outputs"]:
            bad += 1
            print("DIVERGES %-14s\n  want %s\n  got  %s"
                  % (w["name"], w["outputs"], g["outputs"]))
    print("%d/%d vectors match" % (len(want) - bad, len(want)))
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
