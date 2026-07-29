#!/usr/bin/env python3
"""Verify the opinion ledger on the SERVED page (never the row).

Static: served HTML carries the ledger markers, zero template residue, and the
served JS is byte-identical to what was written to js_content. Behavioural: a
real Chromium visit to https://vonc.com/tools/gauntlet — fresh visitor sees the
sealed door and NO ledger; then one real round end-to-end (real CORS, no
proxy); after the verdict the diary holds exactly that round; a new tab shows
sealed door + diary; the shipped overflow probe stays clean throughout.
"""
import json, pathlib, sys, urllib.request
from playwright.sync_api import sync_playwright

SCRATCH = pathlib.Path("/tmp/claude-1000/-home-ant-projects-agentchassis/8a5e2611-422b-4596-9b52-4c3e3251ad63/scratchpad")
UA = ("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) "
      "Chrome/126.0.0.0 Safari/537.36")
# The host serves NO directory index for this path — /tools/gauntlet/ is 404;
# only the explicit index.html answers 200 (checked 2026-07-29, all variants).
PAGE = "https://vonc.com/tools/gauntlet/index.html"
LEDGER_KEY = "vonc_gauntlet_ledger_v1"

results, errors = [], []


def check(name, ok, detail=""):
    results.append(ok)
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + str(detail)[:140]) if detail else ""), flush=True)


def fetch(url):
    r = urllib.request.Request(url, headers={"User-Agent": UA})
    with urllib.request.urlopen(r, timeout=60) as resp:
        return resp.read()


import time
html = fetch(PAGE + "?cb=%d" % time.time()).decode()
js = fetch("https://vonc.com/tools/assets/gauntlet-interface.js?cb=%d" % time.time())

check("served page carries the ledger block", 'data-gi-ledger' in html and 'Your opinion ledger' in html)
check("served page has ZERO template residue", '{{.' not in html)
check("negative control: sealed door still present", 'gi-sealed' in html)
local_js = (SCRATCH / "ledger_js_new.js").read_bytes()
check("served JS byte-identical to delivered js_content", js == local_js,
      f"served {len(js)}B vs local {len(local_js)}B")

with sync_playwright() as pw:
    b = pw.chromium.launch()
    ctx = b.new_context(viewport={"width": 390, "height": 844}, user_agent=UA)
    page = ctx.new_page()
    page.on("pageerror", lambda e: errors.append("pageerror: " + str(e)))
    page.on("console", lambda m: errors.append(m.text) if m.type == "error" else None)
    page.goto(PAGE, wait_until="networkidle")
    page.wait_for_timeout(600)

    SHIPPED = (SCRATCH / "shipped_check.js").read_text()
    sealed = page.evaluate(
        "() => document.querySelector('[data-component=\"gauntlet-interface\"]').classList.contains('gi-sealed')")
    check("live fresh visit: page opens sealed", sealed is True)
    check("live fresh visit: ledger hidden", page.evaluate(
        "() => document.querySelector('[data-gi-ledger]').offsetParent === null"))
    over = page.evaluate(SHIPPED, "")
    check("live fresh visit: no overflow", not (over.get("over", 0) > 2 or over.get("clipped")), over)

    # One real round on the production page — real origin, real CORS, no proxy.
    page.click("[data-gi-enter-btn]")
    page.wait_for_function(
        "() => !document.querySelector('[data-component=\"gauntlet-interface\"]').classList.contains('gi-sealed')",
        timeout=30000)
    page.fill("[data-gi-position-input]",
              "I agree: a record of what you argued, dated, is worth more than a "
              "memory of it, because memory quietly edits itself and a ledger cannot.")
    page.click("[data-gi-position-submit]")
    page.wait_for_function(
        "() => document.querySelector('[data-gi-opponent-position]').textContent.trim().length > 20",
        timeout=120000)
    page.fill("[data-gi-defence-input]",
              "The challenge asks who would reread it; but the value is not in the "
              "rereading, it is in knowing at the time of writing that you cannot "
              "later pretend to have said something else.")
    page.click("[data-gi-defence-submit]")
    page.wait_for_function(
        "() => document.querySelector('[data-gi-verdict]').textContent.trim().length > 5",
        timeout=120000)

    raw = page.evaluate("() => localStorage.getItem('%s')" % LEDGER_KEY)
    store = json.loads(raw) if raw else None
    dom_verdict = page.evaluate("() => document.querySelector('[data-gi-verdict]').textContent.trim()")
    check("live round: one ledger entry after the verdict",
          isinstance(store, list) and len(store) == 1)
    check("live round: entry.verdict is the judge's real line",
          bool(store) and store[0].get("verdict", "") == dom_verdict, dom_verdict[:60])
    check("live round: ledger visible with the entry", page.evaluate(
        "() => document.querySelector('[data-gi-ledger]').offsetParent !== null"))
    over = page.evaluate(SHIPPED, "")
    check("live round: no overflow with populated ledger",
          not (over.get("over", 0) > 2 or over.get("clipped")), over)

    # Returning visitor: new tab, same context.
    page2 = ctx.new_page()
    page2.goto(PAGE, wait_until="networkidle")
    page2.wait_for_timeout(600)
    check("live returning visitor: sealed door + diary visible", page2.evaluate(
        "() => document.querySelector('[data-component=\"gauntlet-interface\"]').classList.contains('gi-sealed')"
        " && document.querySelector('[data-gi-ledger]').offsetParent !== null"))

    real_errors = [e for e in errors if "404" not in e]
    check("no unexpected page errors", not real_errors, real_errors[:3])
    b.close()

print(("ALL %d PASS" % len(results)) if all(results) else "%d FAILURES" % results.count(False))
sys.exit(0 if all(results) else 1)
