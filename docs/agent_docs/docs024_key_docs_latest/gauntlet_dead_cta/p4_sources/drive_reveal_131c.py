#!/usr/bin/env python3
"""Drive the sealed/reveal gauntlet build end to end against the LIVE engine.

Scenarios:
  A. sealed on load: no provocation visible, entry CTA in the first viewport,
     rules target visible, no page errors, nothing fetched from the feed
  B. rail: /round fails (injected 503) -> page STAYS sealed, error at entry
  C. press Enter -> real /round -> reveal: panel visible, provocation from the
     round, entry gone, clock running
  D. file a position -> real AI counter + challenge (8-20s)
  E. reload mid-round -> revealed immediately, round + drafts restored
  F. overflow: the shipped no_horizontal_overflow JS reads clean at 390px
"""
import json, threading, functools, http.server, socketserver, pathlib
import urllib.request, urllib.error
from playwright.sync_api import sync_playwright

HERE = pathlib.Path(__file__).parent
PORT = 8753
ORIGIN = "https://vonc.com"

# test.html = the new rendered component + the new JS, nothing else
rendered = (HERE / "new_rendered.html").read_text()
(HERE / "test.html").write_text(
    "<!doctype html><html><head><meta charset='utf-8'>"
    "<meta name='viewport' content='width=device-width, initial-scale=1'></head>"
    "<body style='margin:0'>" + rendered +
    "<script src='new_gauntlet.js'></script></body></html>"
)

handler = functools.partial(http.server.SimpleHTTPRequestHandler, directory=str(HERE))
handler.log_message = lambda *a, **k: None
socketserver.TCPServer.allow_reuse_address = True
httpd = socketserver.TCPServer(("127.0.0.1", PORT), handler)
threading.Thread(target=httpd.serve_forever, daemon=True).start()

SHIPPED = (HERE / "shipped_check.js").read_text()
results, errors = [], []


def check(name, ok, detail=""):
    results.append(ok)
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + str(detail)[:160]) if detail else ""), flush=True)


def make_proxy(inject_503_once):
    state = {"inject": inject_503_once}

    def proxy(route):
        req = route.request
        if req.method == "OPTIONS":
            route.fulfill(status=204, headers={
                "access-control-allow-origin": "*",
                "access-control-allow-methods": "POST, OPTIONS",
                "access-control-allow-headers": "content-type"})
            return
        if state["inject"] and req.url.endswith("/round"):
            state["inject"] = False
            route.fulfill(status=503, body=json.dumps({"error": "gauntlet opponent unavailable"}),
                          headers={"content-type": "application/json",
                                   "access-control-allow-origin": "*"})
            return
        body = (req.post_data or "").encode()
        r = urllib.request.Request(req.url, data=body, method=req.method, headers={
            "Content-Type": "application/json", "Origin": ORIGIN,
            "User-Agent": req.headers.get("user-agent", "Mozilla/5.0"),
            "Accept": "application/json"})
        try:
            with urllib.request.urlopen(r, timeout=120) as resp:
                status, payload = resp.status, resp.read()
        except urllib.error.HTTPError as e:
            status, payload = e.code, e.read()
        route.fulfill(status=status, body=payload, headers={
            "content-type": "application/json", "access-control-allow-origin": "*"})
    return proxy


with sync_playwright() as pw:
    browser = pw.chromium.launch()
    ctx = browser.new_context(viewport={"width": 390, "height": 844})
    page = ctx.new_page()
    page.on("pageerror", lambda e: errors.append("pageerror: " + str(e)))
    page.on("console", lambda m: errors.append(m.text) if m.type == "error" else None)
    page.route("**tools.apis.uk/**", make_proxy(inject_503_once=True))
    page.goto(f"http://127.0.0.1:{PORT}/test.html", wait_until="networkidle")
    page.wait_for_timeout(800)

    # A. sealed on load
    check("A1 section sealed", page.evaluate("() => document.querySelector('section').classList.contains('gi-sealed')"))
    check("A2 provocation panel hidden", page.evaluate(
        "() => { const p = document.querySelector('.gi-challenge-panel'); return !p || p.offsetParent === null; }"))
    check("A3 entry visible", page.evaluate(
        "() => { const e = document.querySelector('[data-gi-entry]'); return !!e && e.offsetParent !== null; }"))
    btn_top = page.evaluate("() => document.querySelector('[data-gi-enter-btn]').getBoundingClientRect().top")
    check("A4 enter button in first viewport", 0 < btn_top < 844, f"top={btn_top:.0f}px (was ~1913px)")
    check("A5 rules card visible (anchor target lives)", page.evaluate(
        "() => { const r = document.getElementById('gi-rules'); return !!r && r.offsetParent !== null; }"))
    check("A6 provocation title empty (no feed prefetch)", page.evaluate(
        "() => document.querySelector('[data-gi-challenge-title]').textContent.trim() === ''"))
    load_errs = [e for e in errors if "404" not in e]
    check("A7 no page errors on load (component asset 404 is local-only)", not load_errs, load_errs[:2])

    # B. rail: injected 503 -> stays sealed, error spoken at the entry
    page.click("[data-gi-enter-btn]")
    page.wait_for_timeout(1200)
    check("B1 STAYS sealed on /round failure", page.evaluate(
        "() => document.querySelector('section').classList.contains('gi-sealed')"))
    entry_status = page.evaluate("() => document.querySelector('[data-gi-entry-status]').textContent")
    check("B2 error visible at the entry", "offline" in entry_status or "unavailable" in entry_status, entry_status)

    # C. press again -> real round -> reveal
    page.click("[data-gi-enter-btn]")
    page.wait_for_function(
        "() => !document.querySelector('section').classList.contains('gi-sealed')", timeout=30000)
    check("C1 revealed after real /round", True)
    check("C2 entry hidden", page.evaluate(
        "() => document.querySelector('[data-gi-entry]').offsetParent === null"))
    check("C3 panel visible", page.evaluate(
        "() => document.querySelector('.gi-challenge-panel').offsetParent !== null"))
    prov = page.evaluate("() => document.querySelector('[data-gi-challenge-title]').textContent.trim()")
    check("C4 provocation from the round", len(prov) > 10, prov[:70])
    timer = page.evaluate("() => document.querySelector('[data-gi-timer]').textContent")
    check("C5 clock running", timer.startswith("19:") or timer.startswith("20:"), timer)

    # D. file a position -> real AI reply
    page.fill("[data-gi-position-input]",
              "I agree. A feed tuned to one person optimises for what keeps them scrolling, "
              "not for what two people can talk about, and shared reference points are what "
              "conversations are made of.")
    page.click("[data-gi-position-submit]")
    page.wait_for_function(
        "() => document.querySelector('[data-gi-opponent-position]').textContent.trim().length > 20",
        timeout=60000)
    counter = page.evaluate("() => document.querySelector('[data-gi-opponent-position]').textContent.trim()")
    check("D1 real counter-position", len(counter) > 20, counter[:70])

    # E. reload mid-round -> revealed + restored (same page => same sessionStorage)
    page.reload(wait_until="networkidle")
    page.wait_for_timeout(1000)
    check("E1 NOT sealed after mid-round reload", page.evaluate(
        "() => !document.querySelector('section').classList.contains('gi-sealed')"))
    check("E2 provocation restored", page.evaluate(
        "() => document.querySelector('[data-gi-challenge-title]').textContent.trim().length > 10"))
    check("E3 opponent reply restored", page.evaluate(
        "() => document.querySelector('[data-gi-opponent-position]').textContent.trim().length > 20"))

    # F. overflow, both states
    over_revealed = page.evaluate(SHIPPED, "")
    check("F1 no overflow when revealed", not (over_revealed.get("over", 0) > 2 or over_revealed.get("clipped")), over_revealed)
    ctx2 = browser.new_context(viewport={"width": 390, "height": 844})
    fresh = ctx2.new_page()
    fresh.route("**tools.apis.uk/**", make_proxy(inject_503_once=False))
    fresh.goto(f"http://127.0.0.1:{PORT}/test.html", wait_until="networkidle")
    fresh.wait_for_timeout(800)
    check("F2 fresh visitor (new context) is sealed", fresh.evaluate(
        "() => document.querySelector('section').classList.contains('gi-sealed')"))
    over_sealed = fresh.evaluate(SHIPPED, "")
    check("F3 no overflow when sealed", not (over_sealed.get("over", 0) > 2 or over_sealed.get("clipped")), over_sealed)
    ctx2.close()

    late_errors = [e for e in errors if "503" not in e and "Failed to load resource" not in e]
    check("G1 no unexpected page errors end-to-end", not late_errors, late_errors[:3])
    browser.close()

print(("\nALL %d PASS" % len(results)) if all(results) else "\n%d FAILURES" % results.count(False))
