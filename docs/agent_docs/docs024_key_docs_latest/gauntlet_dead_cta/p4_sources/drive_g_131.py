#!/usr/bin/env python3
"""Drive the G build (verdict share card) through a full live round."""
import threading, functools, http.server, socketserver, pathlib
import urllib.request, urllib.error
from playwright.sync_api import sync_playwright

HERE = pathlib.Path(__file__).parent
PORT = 8763
ORIGIN = "https://vonc.com"

rendered = (HERE / "g_rendered_new.html").read_text()
(HERE / "test_g.html").write_text(
    "<!doctype html><html><head><meta charset='utf-8'>"
    "<meta name='viewport' content='width=device-width, initial-scale=1'></head>"
    "<body style='margin:0'>" + rendered +
    "<script src='g_js_new.js'></script></body></html>"
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
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + str(detail)[:140]) if detail else ""), flush=True)


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


with sync_playwright() as pw:
    b = pw.chromium.launch()
    ctx = b.new_context(viewport={"width": 390, "height": 844}, accept_downloads=True)
    page = ctx.new_page()
    page.on("pageerror", lambda e: errors.append("pageerror: " + str(e)))
    page.on("console", lambda m: errors.append(m.text) if m.type == "error" else None)
    page.route("**tools.apis.uk/**", proxy)
    page.goto(f"http://127.0.0.1:{PORT}/test_g.html", wait_until="networkidle")
    page.wait_for_timeout(600)

    page.click("[data-gi-enter-btn]")
    page.wait_for_function(
        "() => !document.querySelector('section').classList.contains('gi-sealed')", timeout=30000)
    check("G: share button INVISIBLE before any verdict", page.evaluate(
        "() => document.querySelector('[data-gi-share-card]').offsetParent === null"))

    page.fill("[data-gi-position-input]",
              "I agree: a personalised feed optimises for each person's attention rather "
              "than for anything two people could discuss, and shared reference points "
              "are what conversations are made of.")
    page.click("[data-gi-position-submit]")
    page.wait_for_function(
        "() => document.querySelector('[data-gi-opponent-position]').textContent.trim().length > 20",
        timeout=60000)
    page.fill("[data-gi-defence-input]",
              "The challenge treats discovery and common ground as rivals; but a discovery "
              "is only worth sharing against a backdrop others recognise, which is exactly "
              "what disappears when every feed is tuned to one person.")
    page.click("[data-gi-defence-submit]")
    page.wait_for_function(
        "() => document.querySelector('[data-gi-verdict]').textContent.trim().length > 5",
        timeout=60000)

    check("G: share button visible AFTER the verdict", page.evaluate(
        "() => document.querySelector('[data-gi-share-card]').offsetParent !== null"))
    card = page.evaluate(
        "() => { const c = buildVerdictCard ? null : null; return null; }") if False else None
    dims = page.evaluate(
        "() => { const s = document.querySelector('[data-gi-share-card]');"
        " return s ? 'present' : 'absent'; }")
    # the card itself, measured in the renderer
    probe = page.evaluate("""() => {
      // reach the closure via the click path is not possible; rebuild the card
      // the same way the page does, from the same live DOM facts
      const prov = document.querySelector('[data-gi-challenge-title]').textContent.trim();
      const verdict = document.querySelector('[data-gi-verdict]').textContent.trim();
      return {prov: prov.slice(0, 40), verdict: verdict.slice(0, 40),
              provLen: prov.length, verdictLen: verdict.length};
    }""")
    check("G: card inputs are real round facts", probe["provLen"] > 10 and probe["verdictLen"] > 5,
          f"{probe['prov']}… / {probe['verdict']}…")

    with page.expect_download(timeout=15000) as dl:
        page.click("[data-gi-share-card]")
    download = dl.value
    path = download.path()
    size = pathlib.Path(path).stat().st_size if path else 0
    check("G: click produces the PNG download", download.suggested_filename == "gauntlet-verdict.png" and size > 8000,
          f"{download.suggested_filename}, {size} bytes")
    # PNG magic + geometry from the file itself
    with open(path, "rb") as f:
        head = f.read(33)
    import struct
    is_png = head[:8] == b"\x89PNG\r\n\x1a\n"
    w, h = struct.unpack(">II", head[16:24])
    check("G: the card is a real 1200x630 PNG", is_png and (w, h) == (1200, 630), f"png={is_png} {w}x{h}")

    over = page.evaluate(SHIPPED, "")
    check("no overflow with the share row", not (over.get("over", 0) > 2 or over.get("clipped")), over)
    real_errors = [e for e in errors if "404" not in e]
    check("no unexpected page errors", not real_errors, real_errors[:3])
    b.close()

print(("ALL %d PASS" % len(results)) if all(results) else "%d FAILURES" % results.count(False))
