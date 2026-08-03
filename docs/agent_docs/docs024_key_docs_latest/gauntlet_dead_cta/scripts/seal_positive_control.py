#!/usr/bin/env python3
# Positive control for the leak sweep's negative result (2026-08-03).
#
# The sweep reported 0 of 20 pages painting today's provocation. That is only
# evidence if the same instrument, on the same page, can still SEE provocation
# text when provocation text is there. Otherwise a broken renderer, a failed
# fetch, or a probe that no longer matches anything would all read as "sealed".
#
# So: on "/" assert BOTH directions with one innerText read —
#   sample.headline / sample.body   MUST be painted  (the designed replacement)
#   today.headline  / today.body    MUST NOT be      (the seal)
# A run where the sample is also absent is UNSCORED, not a pass: it means the
# card did not render at all and the seal's "absence" proves nothing.

import json
import re
import sys
import urllib.request

SITE = "https://vonc.com"


def strip_tags(s):
    return re.sub(r"<[^>]+>", "", s or "").strip().rstrip(".")


req = urllib.request.Request(SITE + "/data/provocations.json",
                             headers={"User-Agent": "Mozilla/5.0"})
with urllib.request.urlopen(req, timeout=30) as r:
    feed = json.load(r)

today = feed["today"]
sample = feed.get("sample") or {}
seal = feed.get("seal") or {}

must_be_absent = {
    "today.headline": strip_tags(today.get("headline")),
    "today.body_open": (today.get("body") or "").strip()[:45],
}
must_be_painted = {
    "sample.headline": strip_tags(sample.get("headline")),
    "sample.body_open": (sample.get("body") or "").strip()[:45],
}

for name, val in list(must_be_absent.items()) + list(must_be_painted.items()):
    if not val:
        print(f"UNSCORED: {name} is empty in the feed — no claim possible")
        sys.exit(2)

print("today.slug =", today.get("slug"), "| date =", today.get("date"))
print("sample.slug=", sample.get("slug"), "| date =", sample.get("date"))
print("seal copy  =", json.dumps(seal)[:120])
print()

from playwright.sync_api import sync_playwright

failures, unscored = [], []
with sync_playwright() as p:
    b = p.chromium.launch()
    for url in (SITE + "/", SITE + "/tools/arena/index.html"):
        pg = b.new_context().new_page()
        resp = pg.goto(url, wait_until="load", timeout=45000)
        pg.wait_for_timeout(3500)
        txt = pg.evaluate("() => document.body.innerText")
        print(f"{url}  HTTP {resp.status if resp else 0}  {len(txt)} chars")

        painted = {k: (v in txt) for k, v in must_be_painted.items()}
        absent = {k: (v not in txt) for k, v in must_be_absent.items()}

        for k, ok in painted.items():
            print(f"   {'PAINTED ' if ok else 'MISSING '} {k}")
        for k, ok in absent.items():
            print(f"   {'SEALED  ' if ok else 'LEAK    '} {k}")

        # The control: on "/" the sample MUST be visible, else the negative is vacuous.
        if url.endswith("/") and not any(painted.values()):
            unscored.append(f"{url}: sample not painted — seal result is VACUOUS")
        for k, ok in absent.items():
            if not ok:
                failures.append(f"{url}: {k} is painted")
        pg.close()
        print()
    b.close()

if unscored:
    print("UNSCORED — the instrument could not have seen a leak here:")
    for u in unscored:
        print("  ", u)
    sys.exit(2)
if failures:
    print("LEAK CONFIRMED:")
    for f in failures:
        print("  ", f)
    sys.exit(1)
print("PASS — sample painted (instrument proven live), today absent (seal holds)")
sys.exit(0)
