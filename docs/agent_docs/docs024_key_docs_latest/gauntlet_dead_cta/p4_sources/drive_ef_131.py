#!/usr/bin/env python3
"""Drive the E/F build (provocation card + step emphasis) against the LIVE
engine, full round including a defence and verdict."""
import json, threading, functools, http.server, socketserver, pathlib
import urllib.request, urllib.error
from playwright.sync_api import sync_playwright

HERE = pathlib.Path(__file__).parent
PORT = 8761
ORIGIN = "https://vonc.com"

rendered = (HERE / "ef_rendered_new.html").read_text()
(HERE / "test_ef.html").write_text(
    "<!doctype html><html><head><meta charset='utf-8'>"
    "<meta name='viewport' content='width=device-width, initial-scale=1'></head>"
    "<body style='margin:0'>" + rendered +
    "<script src='ef_js_new.js'></script></body></html>"
)

socketserver.TCPServer.allow_reuse_address = True
handler = functools.partial(http.server.SimpleHTTPRequestHandler, directory=str(HERE))
handler.log_message = lambda *a, **k: None
httpd = socketserver.TCPServer(("127.0.0.1", PORT), handler)
threading.Thread(target=httpd.serve_forever, daemon=True).start()

SHIPPED = (HERE / "shipped_check.js").read_text()
results, errors = [], []


def check(name, ok, detail=""):
    results.append(ok)
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + str(detail)[:150]) if detail else ""), flush=True)


def proxy(route):
    req = route.request
    if req.method == "OPTIONS":
        route.fulfill(status=204, headers={
            "access-control-allow-origin": "*",
            "access-control-allow-methods": "POST, OPTIONS",
            "access-control-allow-headers": "content-type"})
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


EMPH = """(i) => {
  const s = document.querySelectorAll('.gi-steps .gi-step')[i];
  return ['is-current','is-done','is-future'].find(c => s.classList.contains(c)) || 'none';
}"""

with sync_playwright() as pw:
    b = pw.chromium.launch()
    page = b.new_context(viewport={"width": 390, "height": 844}).new_page()
    page.on("pageerror", lambda e: errors.append("pageerror: " + str(e)))
    page.on("console", lambda m: errors.append(m.text) if m.type == "error" else None)
    page.route("**tools.apis.uk/**", proxy)
    page.goto(f"http://127.0.0.1:{PORT}/test_ef.html", wait_until="networkidle")
    page.wait_for_timeout(600)

    check("sealed on load (C intact)", page.evaluate(
        "() => document.querySelector('section').classList.contains('gi-sealed')"))

    page.click("[data-gi-enter-btn]")
    page.wait_for_function(
        "() => !document.querySelector('section').classList.contains('gi-sealed')", timeout=30000)
    check("revealed on real /round (C intact)", True)

    # E: the question card
    check("E: provocation card present", page.evaluate(
        "() => !!document.querySelector('.gi-provocation-card')"))
    check("E: card holds eyebrow+title+text+intro", page.evaluate(
        "() => { const c = document.querySelector('.gi-provocation-card');"
        " return !!(c.querySelector('[data-gi-challenge-title]') && c.querySelector('[data-gi-challenge-body]')"
        " && c.querySelector('.gi-challenge-intro')); }"))
    edge = page.evaluate(
        "() => getComputedStyle(document.querySelector('.gi-provocation-card')).borderLeftColor")
    check("E: stage accent edge painted", edge == "rgb(245, 158, 11)", edge)
    check("E: intro styled as attached directive (border-top)", page.evaluate(
        "() => getComputedStyle(document.querySelector('.gi-provocation-card .gi-challenge-intro')).borderTopWidth") == "1px")

    # F: emphasis at round start — position current, defence future
    check("F: position is-current at start", page.evaluate(EMPH, 0) == "is-current", page.evaluate(EMPH, 0))
    check("F: defence is-future at start", page.evaluate(EMPH, 2) == "is-future", page.evaluate(EMPH, 2))
    op = page.evaluate(
        "() => getComputedStyle(document.querySelectorAll('.gi-steps .gi-step')[2]).opacity")
    check("F: future step visually muted", float(op) < 0.7, f"opacity={op}")
    check("F: future step's control still ENABLED (rank, not gate)", page.evaluate(
        "() => !document.querySelector('[data-gi-defence-submit]').disabled"))

    # file position -> defence becomes current
    page.fill("[data-gi-position-input]",
              "I agree: personalisation optimises for each person's attention, not for what "
              "two people can talk about, and shared reference points are the raw material "
              "of conversation.")
    page.click("[data-gi-position-submit]")
    page.wait_for_function(
        "() => document.querySelector('[data-gi-opponent-position]').textContent.trim().length > 20",
        timeout=60000)
    check("F: after position — step1 is-done", page.evaluate(EMPH, 0) == "is-done", page.evaluate(EMPH, 0))
    check("F: after position — defence is-current", page.evaluate(EMPH, 2) == "is-current", page.evaluate(EMPH, 2))
    cur_shadow = page.evaluate(
        "() => getComputedStyle(document.querySelectorAll('.gi-steps .gi-step')[2]).boxShadow")
    check("F: current step carries the inset marker", "inset" in cur_shadow, cur_shadow[:60])

    # reload mid-round: emphasis re-derived from restored state
    page.reload(wait_until="networkidle")
    page.wait_for_timeout(1000)
    check("F: reload re-derives — defence still is-current", page.evaluate(EMPH, 2) == "is-current",
          page.evaluate(EMPH, 2))

    # defence -> verdict current, everything else done
    page.fill("[data-gi-defence-input]",
              "The challenge assumes discovery and sharedness are rivals, but recommendation "
              "engines only feel valuable against a common backdrop; without one there is "
              "nothing for a discovery to be surprising relative to.")
    page.click("[data-gi-defence-submit]")
    page.wait_for_function(
        "() => document.querySelector('[data-gi-verdict]').textContent.trim().length > 5",
        timeout=60000)
    check("F: after verdict — verdict is-current", page.evaluate(EMPH, 3) == "is-current", page.evaluate(EMPH, 3))
    check("F: after verdict — defence is-done", page.evaluate(EMPH, 2) == "is-done", page.evaluate(EMPH, 2))

    over = page.evaluate(SHIPPED, "")
    check("no overflow (E card + F marks)", not (over.get("over", 0) > 2 or over.get("clipped")), over)
    real_errors = [e for e in errors if "404" not in e]
    check("no unexpected page errors", not real_errors, real_errors[:3])
    b.close()

print(("ALL %d PASS" % len(results)) if all(results) else "%d FAILURES" % results.count(False))
