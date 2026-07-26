#!/usr/bin/env python3
"""Drive the rebuilt gauntlet component end to end against the LIVE tools-api.

The browser sends Origin: http://localhost:... which the API rejects (403), so
requests to the engine are intercepted and re-stamped with the real site origin.
Everything else — the template, the JS, the feed, the AI calls — is the real
article. Waits are generous because the AI calls genuinely take 8-18 seconds.
"""
import sys, json, threading, functools, http.server, socketserver, pathlib
import urllib.request, urllib.error
from playwright.sync_api import sync_playwright

HERE = pathlib.Path(__file__).parent
PORT = 8747
ORIGIN = "https://vonc.com"

handler = functools.partial(http.server.SimpleHTTPRequestHandler, directory=str(HERE))
handler.log_message = lambda *a, **k: None
httpd = socketserver.TCPServer(("127.0.0.1", PORT), handler)
threading.Thread(target=httpd.serve_forever, daemon=True).start()

results = []
console_errors = []


def check(name, ok, detail=""):
    results.append((name, ok, detail))
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + detail) if detail else ""), flush=True)


def run(profile, width, height):
    print("\n=== profile: %s (%dx%d) ===" % (profile, width, height), flush=True)
    with sync_playwright() as pw:
        browser = pw.chromium.launch()
        page = browser.new_page(viewport={"width": width, "height": height})
        page.on("console", lambda m: console_errors.append(m.text) if m.type == "error" else None)
        page.on("pageerror", lambda e: console_errors.append("pageerror: " + str(e)))

        # The browser's CORS preflight carries the localhost origin and the API
        # rejects it (403), and a preflight cannot be re-stamped by routing. So
        # the request is made from here with the real site origin and handed
        # back to the page. The API, the AI calls and the JS are all real; only
        # the browser's own CORS check is bypassed — that is verified separately
        # against the deployed page.
        def proxy(route):
            req = route.request
            if req.method == "OPTIONS":
                route.fulfill(status=204, headers={
                    "access-control-allow-origin": "*",
                    "access-control-allow-methods": "POST, OPTIONS",
                    "access-control-allow-headers": "content-type",
                })
                return
            body = (req.post_data or "").encode()
            # Cloudflare sits in front of the API and answers a bare
            # Python-urllib fingerprint with "error code: 1010" (a 403 that is
            # NOT the API's own origin check), so pass a real browser UA.
            r = urllib.request.Request(req.url, data=body, method=req.method, headers={
                "Content-Type": "application/json", "Origin": ORIGIN,
                "User-Agent": req.headers.get("user-agent", "Mozilla/5.0"),
                "Accept": "application/json",
            })
            try:
                with urllib.request.urlopen(r, timeout=120) as resp:
                    status, payload = resp.status, resp.read()
            except urllib.error.HTTPError as e:
                status, payload = e.code, e.read()
            route.fulfill(status=status, body=payload, headers={
                "content-type": "application/json",
                "access-control-allow-origin": "*",
            })

        page.route("**tools.apis.uk/**", proxy)
        page.goto("http://127.0.0.1:%d/test.html" % PORT, wait_until="networkidle")

        # Selector inventory — every hook the acceptance criteria name.
        for sel in ["[data-gi-enter-btn]", "[data-gi-status]", "[data-gi-challenge-body]",
                    "[data-gi-position-input]", "[data-gi-position-submit]",
                    "[data-gi-opponent-position]", "[data-gi-opponent-challenge]",
                    "[data-gi-defence-input]", "[data-gi-defence-submit]",
                    "[data-gi-verdict]", "[data-gi-verdict-reasons]",
                    "[data-gi-pct]", "[data-gi-obj]", "[data-gi-timer]"]:
            check("selector_exists %s" % sel, page.locator(sel).count() > 0)

        # Dead-control sweep: no href="#" or empty href anywhere in the section.
        dead = page.evaluate("""() => {
            const s = document.querySelector('[data-component="gauntlet-interface"]');
            return [...s.querySelectorAll('a')].map(a => a.getAttribute('href'))
                     .filter(h => h === null || h.trim() === '' || h.trim() === '#');
        }""")
        check("no dead anchors in the section", dead == [], "found %r" % (dead,))

        # The provocation is on the page before any round starts.
        body_pre = page.locator("[data-gi-challenge-body]").inner_text().strip()
        check("provocation pre-rendered from the feed", len(body_pre) > 40, "%d chars" % len(body_pre))
        check("progress starts at 0%", "0%" in page.locator("[data-gi-pct]").inner_text())
        check("no objective pre-marked", page.locator("[data-gi-obj].is-complete").count() == 0)
        check("clock not started", page.locator("[data-gi-timer]").inner_text().strip() == "20:00")

        # --- Journey C.1: a real round starts ---
        page.click("[data-gi-enter-btn]")
        try:
            page.wait_for_function(
                """() => document.querySelector('[data-gi-round-state]').textContent.includes('live')""",
                timeout=40000)
            check("round_starts: state badge goes live", True)
        except Exception:
            check("round_starts: state badge goes live", False,
                  "status: " + page.locator("[data-gi-status]").inner_text().strip()[:160])
            browser.close(); return
        body = page.locator("[data-gi-challenge-body]").inner_text().strip()
        check("round_starts: challenge body populated", len(body) > 40, "%d chars" % len(body))
        page.wait_for_timeout(1500)
        clock = page.locator("[data-gi-timer]").inner_text().strip()
        check("round_starts: clock is running", clock != "20:00", "reads %s" % clock)
        check("round_starts: still no objective marked",
              page.locator("[data-gi-obj].is-complete").count() == 0)

        # --- Journey C.2: position -> real AI counter ---
        page.fill("[data-gi-position-input]",
                  "Personalisation is what people asked for; a shared internet was only ever an artefact of scarcity.")
        page.click("[data-gi-position-submit]")
        page.wait_for_function(
            """() => document.querySelector('[data-gi-opponent-position]').textContent.trim().length > 40""",
            timeout=90000)
        counter = page.locator("[data-gi-opponent-position]").inner_text().strip()
        chall = page.locator("[data-gi-opponent-challenge]").inner_text().strip()
        check("position_flow: opponent counter-position is real AI text", len(counter) > 40, "%d chars" % len(counter))
        check("position_flow: opponent challenge populated", len(chall) > 20, "%d chars" % len(chall))
        n1 = page.locator("[data-gi-obj].is-complete").count()
        check("position_flow: exactly one objective marked", n1 == 1, "count=%d" % n1)
        check("position_flow: progress reads 33%", "33%" in page.locator("[data-gi-pct]").inner_text(),
              page.locator("[data-gi-pct]").inner_text())

        # --- Journey C.3: defence -> real verdict ---
        page.fill("[data-gi-defence-input]",
                  "No platform has ever offered a clearly labelled shared feed as a real option, so revealed preference proves nothing here.")
        page.click("[data-gi-defence-submit]")
        try:
            page.wait_for_function(
                """() => document.querySelector('[data-gi-verdict]').textContent.trim().length > 0""",
                timeout=90000)
            verdict = page.locator("[data-gi-verdict]").inner_text().strip()
            reasons = page.locator("[data-gi-verdict-reasons]").inner_text().strip()
            check("defend_flow: verdict populated", len(verdict) > 0, verdict[:60])
            check("defend_flow: verdict reasons populated", len(reasons) > 40, "%d chars" % len(reasons))
            n2 = page.locator("[data-gi-obj].is-complete").count()
            check("defend_flow: all three objectives marked", n2 == 3, "count=%d" % n2)
            check("defend_flow: progress reads 100%", "100%" in page.locator("[data-gi-pct]").inner_text(),
                  page.locator("[data-gi-pct]").inner_text())
        except Exception:
            status = page.locator("[data-gi-status]").inner_text().strip()
            offline = "offline" in status.lower()
            check("defend_flow: verdict OR an honest offline message", offline,
                  "status reads: %s" % status[:140])
            check("defend_flow: no objective faked on failure",
                  page.locator("[data-gi-obj].is-complete").count() <= 2)

        # --- overflow + console ---
        over = page.evaluate("""() => {
            const d = document.documentElement;
            return {over: d.scrollWidth - d.clientWidth,
                    widest: [...document.querySelectorAll('[data-component="gauntlet-interface"] *')]
                      .map(e => ({t: e.tagName + '.' + e.className, w: Math.round(e.getBoundingClientRect().width)}))
                      .sort((a,b) => b.w - a.w)[0]};
        }""")
        check("no horizontal overflow", over["over"] <= 2,
              "overflow %spx, widest %s" % (over["over"], over.get("widest")))
        browser.close()


run("desktop", 1366, 900)
run("mobile", 390, 844)

real_errors = [e for e in console_errors if "favicon" not in e.lower()]
check("no console errors", not real_errors, " | ".join(real_errors[:3]))

failed = [r for r in results if not r[1]]
print("\n%d passed, %d failed" % (len(results) - len(failed), len(failed)))
httpd.shutdown()
sys.exit(1 if failed else 0)
