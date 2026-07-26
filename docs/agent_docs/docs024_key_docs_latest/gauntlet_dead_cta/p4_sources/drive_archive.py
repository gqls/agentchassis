#!/usr/bin/env python3
"""Drive the rewritten provocations-archive-loader against the REAL live page.

The live /provocations/index.html is fetched and served locally with the new
loader substituted for the shipped snippets bundle and the new feed in place —
so the markup under test is the actual deployed component template, not a
reconstruction.
"""
import sys, re, threading, functools, http.server, socketserver, pathlib, urllib.request
from playwright.sync_api import sync_playwright

HERE = pathlib.Path(__file__).parent
PORT = 8749

results = []
console_errors = []


def check(name, ok, detail=""):
    results.append((name, ok, detail))
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + detail) if detail else ""), flush=True)


# --- assemble the local mirror of the live page -----------------------------
req = urllib.request.Request(
    "https://vonc.com/provocations/index.html?cb=harness",
    headers={"User-Agent": "Mozilla/5.0"})
html = urllib.request.urlopen(req, timeout=60).read().decode("utf-8", "replace")

# Only the archive loader is under test; serve it in place of the whole bundle.
(HERE / "assets" / "js").mkdir(parents=True, exist_ok=True)
(HERE / "assets" / "js" / "snippets.js").write_text((HERE / "archive_loader_new.js").read_text())

# Strip cross-origin chrome that would just add noise (fonts, analytics).
html = re.sub(r'<link[^>]+href="https?://(?!vonc\.com)[^"]+"[^>]*>', "", html)
(HERE / "provocations_index.html").write_text(html)

handler = functools.partial(http.server.SimpleHTTPRequestHandler, directory=str(HERE))
handler.log_message = lambda *a, **k: None
httpd = socketserver.TCPServer(("127.0.0.1", PORT), handler)
threading.Thread(target=httpd.serve_forever, daemon=True).start()

BASE = "http://127.0.0.1:%d/provocations_index.html" % PORT


def run(profile, width, height):
    print("\n=== profile: %s (%dx%d) ===" % (profile, width, height), flush=True)
    with sync_playwright() as pw:
        browser = pw.chromium.launch()
        page = browser.new_page(viewport={"width": width, "height": height})
        page.on("console", lambda m: console_errors.append(m.text) if m.type == "error" else None)
        page.on("pageerror", lambda e: console_errors.append("pageerror: " + str(e)))
        page.goto(BASE, wait_until="networkidle")

        check("archive_list_exists", page.locator(".provocations-archive__list").count() > 0)
        check("archive_detail_region_exists", page.locator(".provocations-archive__detail").count() > 0)

        linked = page.locator(".provocations-archive__item--linked")
        static = page.locator(".provocations-archive__item--static")
        check("archive_linked_item_exists", linked.count() == 7, "count=%d" % linked.count())
        check("archive_static_item_exists", static.count() == 1, "count=%d" % static.count())

        # Journey B.3 — a static entry is genuinely not offered as openable.
        st = page.evaluate("""() => {
            const e = document.querySelector('.provocations-archive__item--static');
            return e ? {href: e.getAttribute('href'), tabindex: e.getAttribute('tabindex')} : null;
        }""")
        check("static entry has no href and no tabindex",
              st is not None and st["href"] is None and st["tabindex"] is None, repr(st))

        # No dead controls anywhere in the section.
        dead = page.evaluate("""() => {
            const s = document.querySelector('[data-component="provocations-archive-list"]');
            return [...s.querySelectorAll('a')].map(a => a.getAttribute('href'))
                     .filter(h => h !== null && (h.trim() === '' || h.trim() === '#'));
        }""")
        check("no dead anchors in the archive section", dead == [], "found %r" % (dead,))

        # Detail region is present but empty before any click.
        pre = page.locator(".provocations-archive__detail").inner_text().strip()
        check("detail region empty before opening", pre == "", repr(pre[:60]))

        # Journey B.1 — click a linked row, detail populates, URL updates, no reload.
        page.evaluate("window.__stillHere = true")
        first = linked.first
        slug = first.get_attribute("data-slug")
        first.click()
        page.wait_for_selector(".provocations-archive__detail:not([hidden])", timeout=5000)
        txt = page.locator(".provocations-archive__detail").inner_text().strip()
        check("archive_open_detail_populates", len(txt) > 200, "%d chars" % len(txt))
        check("detail carries the entry title",
              first.locator(".provocations-archive__item-title").inner_text().strip()
              in txt, "")
        check("URL updated to ?entry=<slug>", ("entry=" + slug) in page.url, page.url)
        check("no full page reload (pushState)", page.evaluate("window.__stillHere === true"))

        # Journey B.2 — direct load of the deep link pre-populates without a click.
        page2 = browser.new_page(viewport={"width": width, "height": height})
        page2.goto(BASE + "?entry=" + slug, wait_until="networkidle")
        page2.wait_for_selector(".provocations-archive__detail:not([hidden])", timeout=8000)
        deep = page2.locator(".provocations-archive__detail").inner_text().strip()
        check("deep link pre-populates on direct load", len(deep) > 200, "%d chars" % len(deep))
        page2.close()

        # Close returns to the clean URL.
        page.click(".provocations-archive__detail-close")
        check("close hides the detail",
              page.locator(".provocations-archive__detail[hidden]").count() == 1)
        check("close clears the query string", "entry=" not in page.url, page.url)

        over = page.evaluate("""() => document.documentElement.scrollWidth - document.documentElement.clientWidth""")
        check("no horizontal overflow", over <= 2, "overflow %spx" % over)
        browser.close()


run("desktop", 1366, 900)
run("mobile", 390, 844)

real_errors = [e for e in console_errors if "favicon" not in e.lower() and "net::ERR" not in e and "404" not in e]
check("no console errors", not real_errors, " | ".join(real_errors[:3]))

failed = [r for r in results if not r[1]]
print("\n%d passed, %d failed" % (len(results) - len(failed), len(failed)))
httpd.shutdown()
sys.exit(1 if failed else 0)
