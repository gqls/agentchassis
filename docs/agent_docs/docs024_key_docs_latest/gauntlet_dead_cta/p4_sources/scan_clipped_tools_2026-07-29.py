#!/usr/bin/env python3
"""
Scan live tool pages for the 131-B "clipped overflow" condition, using the
EXACT clause JS shipped in the browser-runner-adapter — extracted from the Go
source at runtime, never hand-copied, so a drift between this scan and the
deployed check is impossible by construction.

Purpose: find a page the DEPLOYED adapter would flag, so the new clause can be
witnessed catching something rather than only passing. A pass here means the
page is genuinely clean.

Mobile profile matched to run_checks_action.go: 390x844, iPhone UA, DSF 3,
is_mobile, has_touch, WaitUntil=load.
"""
import re
import sys
import json

GO_SRC = "/home/ant/projects/agentchassis/internal/adapters/browserrunner/run_checks_action.go"

MOBILE_UA = ("Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) "
             "AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1")


def extract_clause():
    """Pull the JS between the backticks of HorizontalOverflow's page.Evaluate."""
    src = open(GO_SRC, encoding="utf-8").read()
    anchor = "func (c *chromiumPage) HorizontalOverflow(container string)"
    i = src.index(anchor)
    start = src.index("`", i) + 1
    end = src.index("`", start)
    js = src[start:end]
    # Sanity: the clause must contain its own distinguishing markers, else we
    # extracted the wrong literal and every result below would be vacuous.
    for marker in ("cutCount", "overflowX", "getBoundingClientRect"):
        assert marker in js, f"extracted JS lacks {marker!r} — wrong literal"
    return js


def main(urls, container=".tool-container"):
    from playwright.sync_api import sync_playwright

    js = extract_clause()
    print(f"clause extracted: {len(js)} chars, "
          f"cutCount x{js.count('cutCount')}\n", flush=True)

    flagged, clean, errors = [], [], []
    with sync_playwright() as pw:
        browser = pw.chromium.launch()
        ctx = browser.new_context(
            viewport={"width": 390, "height": 844},
            user_agent=MOBILE_UA,
            device_scale_factor=3,
            is_mobile=True,
            has_touch=True,
        )
        page = ctx.new_page()
        for url in urls:
            try:
                page.goto(url, wait_until="load", timeout=30000)
                res = page.evaluate(js, container)
            except Exception as exc:                       # noqa: BLE001
                errors.append((url, str(exc)[:120]))
                print(f"ERR  {url}\n     {str(exc)[:120]}", flush=True)
                continue
            over = res.get("over", 0)
            # The clause returns ONLY {over: n} when nothing offends; any
            # attribution field present means it flagged something.
            attributed = any(k in res for k in ("element", "component", "slot",
                                                "width", "forced", "detail"))
            if attributed or over > 2:
                flagged.append((url, res))
                print(f"FLAG {url}\n     {json.dumps(res)[:400]}", flush=True)
            else:
                clean.append(url)
                print(f"ok   {url}  (over={over})", flush=True)
        browser.close()

    print(f"\n=== {len(clean)} clean · {len(flagged)} FLAGGED · {len(errors)} errors ===")
    for url, res in flagged:
        print(f"\nFLAGGED {url}\n{json.dumps(res, indent=2)}")
    return flagged


if __name__ == "__main__":
    urls = [l.strip() for l in open(sys.argv[1]) if l.strip()]
    main(urls)
