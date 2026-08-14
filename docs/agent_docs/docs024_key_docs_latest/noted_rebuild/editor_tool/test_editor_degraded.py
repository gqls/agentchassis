#!/usr/bin/env python3
"""
Degraded-state probe for the Noted editor (noted-write.html).

THE CONTRACT UNDER TEST (experience pattern authenticated-note-sync):
  "The editor may say 'Saved' only as a consequence of a successful server
   response" — plus its two corollaries: a failed save is LOUD and never
   touches the text, and leaving with unsaved text asks first.

WHY THIS SHAPE. The platform's check runner cannot induce a failing dependency
(HANDOFF §5 blocker 4), and the engine itself must never be broken to order.
Playwright can do both without touching any server: a local stub serves the
happy API (with a deliberately SLOW save, so the ordering claim is provable),
and per-case route interception induces the outage and the server error
in-browser.

Run:  /home/ant/.venvs/vonc_pw/bin/python test_editor_degraded.py
Exit 0 = all cases passed. Mutation verification is in the runner script the
NOTES entry describes — a suite that has never failed proves nothing.
"""
import http.server
import json
import socketserver
import sys
import threading
import time
from pathlib import Path

from playwright.sync_api import sync_playwright

HERE = Path(__file__).parent
TOOL = (HERE / "noted-write.html").read_text()
PAGE = "<!doctype html><html><head><meta charset='utf-8'><title>write</title></head><body>" + TOOL + "</body></html>"

SAVE_DELAY_S = 1.0   # slow enough that "Saving…" is observable, fast enough to test


class StubAPI(http.server.BaseHTTPRequestHandler):
    """The happy-path engine, faithful to server.go's shapes."""
    def _json(self, code, obj):
        body = json.dumps(obj).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        if self.path == "/api/me":
            self._json(401, {"error": "please sign in"})
        elif self.path == "/api/notes":
            self._json(200, {"notes": [{"id": 7, "title": "Earlier thought", "content": "kept"}]})
        elif self.path in ("/", "/write.html"):
            body = PAGE.encode()
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
        else:
            self._json(404, {"error": "no such path"})

    def do_POST(self):
        if self.path == "/api/login":
            self._json(200, {"email": "probe@example.com"})
        elif self.path == "/api/register":
            self._json(201, {"email": "probe@example.com"})
        elif self.path == "/api/logout":
            self._json(200, {"status": "signed out"})
        elif self.path == "/api/notes":
            time.sleep(SAVE_DELAY_S)          # the observable window for "Saving…"
            n = json.loads(self.rfile.read(int(self.headers.get("Content-Length", 0)) or 0) or b"{}")
            n.setdefault("id", 42)
            self._json(200, n)
        else:
            self._json(404, {"error": "no such path"})

    def log_message(self, *a):  # quiet
        pass


FAILS = []


def check(label, ok, detail=""):
    print(("  PASS  " if ok else "  FAIL  ") + label + (("  — " + detail) if detail else ""))
    if not ok:
        FAILS.append(label)


def sign_in(page):
    page.fill("#nw-email", "probe@example.com")
    page.fill("#nw-password", "0123456789x")
    page.click("#nw-signin")
    page.wait_for_selector("#nw-app:not([hidden])", timeout=5000)


def main():
    httpd = socketserver.ThreadingTCPServer(("127.0.0.1", 0), StubAPI)
    port = httpd.server_address[1]
    threading.Thread(target=httpd.serve_forever, daemon=True).start()
    url = f"http://127.0.0.1:{port}/write.html"

    with sync_playwright() as p:
        browser = p.chromium.launch()

        # ---------- CASE A: sign-in, and a failed sign-in is told plainly ----------
        print("\nCASE A — sign in, and a wrong password is told plainly")
        ctx = browser.new_context()
        page = ctx.new_page()
        page.route("**/api/login", lambda r: r.fulfill(
            status=401, content_type="application/json",
            body=json.dumps({"error": "that email address and password do not match"})))
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        page.fill("#nw-email", "probe@example.com")
        page.fill("#nw-password", "wrong-password")
        page.click("#nw-signin")
        page.wait_for_selector("#nw-auth-error:not([hidden])", timeout=5000)
        check("engine's own message shown", "do not match" in page.locator("#nw-auth-error").inner_text())
        page.unroute("**/api/login")
        sign_in(page)
        check("app visible after sign-in", page.locator("#nw-app").is_visible())
        check("existing note listed", "Earlier thought" in page.locator("#nw-list").inner_text())

        # ---------- CASE B: 'Saved' appears ONLY after the server answers ----------
        print("\nCASE B — the ordering claim: no 'Saved' before the response")
        page.fill("#nw-title", "Ordering")
        page.fill("#nw-content", "the text that must survive")
        check("typing marks it unsaved", page.locator("#nw-status").inner_text() == "Unsaved changes")
        page.click("#nw-save")
        page.wait_for_timeout(int(SAVE_DELAY_S * 300))     # well inside the delay window
        mid = page.locator("#nw-status").inner_text()
        check("mid-flight status is 'Saving…', not 'Saved'", mid == "Saving…", f"was {mid!r}")
        page.wait_for_function(
            "() => document.getElementById('nw-status').textContent.includes('Saved')",
            timeout=int(SAVE_DELAY_S * 1000) + 4000)
        check("'Saved ✓' after the 2xx", "Saved ✓" in page.locator("#nw-status").inner_text())

        # ---------- CASE C: the network dies mid-save ----------
        print("\nCASE C — outage: save fails LOUDLY, text survives")
        page.route("**/api/notes", lambda r: r.abort() if r.request.method == "POST" else r.continue_())
        page.fill("#nw-content", "the text that must survive an outage")
        page.click("#nw-save")
        page.wait_for_selector("#nw-failure:not([hidden])", timeout=5000)
        check("loud failure banner shown", page.locator("#nw-failure").is_visible())
        check("banner says nothing was lost", "nothing has been lost" in page.locator("#nw-failure").inner_text())
        check("status says NOT saved", "NOT saved" in page.locator("#nw-status").inner_text())
        check("text untouched in the editor",
              page.locator("#nw-content").input_value() == "the text that must survive an outage")

        # ---------- CASE D: the server answers 500 with its own words ----------
        print("\nCASE D — server error: the engine's message is surfaced")
        page.unroute("**/api/notes")
        page.route("**/api/notes", lambda r: r.fulfill(
            status=500, content_type="application/json",
            body=json.dumps({"error": "could not save that note"}))
            if r.request.method == "POST" else r.continue_())
        page.click("#nw-retry")
        page.wait_for_selector("#nw-failure:not([hidden])", timeout=5000)
        check("engine's error text shown", "could not save that note" in page.locator("#nw-failure").inner_text())
        check("text still untouched",
              page.locator("#nw-content").input_value() == "the text that must survive an outage")

        # ---------- CASE E: the outage ends; retry saves the SAME text ----------
        print("\nCASE E — retry after recovery")
        page.unroute("**/api/notes")
        page.click("#nw-retry")
        page.wait_for_function(
            "() => document.getElementById('nw-status').textContent.includes('Saved')",
            timeout=int(SAVE_DELAY_S * 1000) + 4000)
        check("'Saved ✓' after recovery", "Saved ✓" in page.locator("#nw-status").inner_text())
        check("failure banner gone", not page.locator("#nw-failure").is_visible())

        # ---------- CASE F: leaving with unsaved text asks first ----------
        print("\nCASE F — beforeunload guard (handler logic; the real prompt is manual)")
        # Playwright's headless shell SUPPRESSES the actual beforeunload prompt
        # (tried: real click + typed keys for user activation, then
        # close(run_before_unload=True) — no dialog event ever arrives). So this
        # case tests OUR guard — that it arms when dirty and disarms after a
        # save — by dispatching a cancelable beforeunload and reading
        # defaultPrevented. The browser's own prompt is exercised in the MANUAL
        # walkthrough, which is the half the platform genuinely cannot do.
        page.click("#nw-content")
        page.keyboard.type(" typed and not saved")
        armed = page.evaluate(
            "() => { const e = new Event('beforeunload', {cancelable: true});"
            " window.dispatchEvent(e); return e.defaultPrevented; }")
        check("guard ARMED while text is unsaved", armed is True)
        page.click("#nw-save")
        page.wait_for_function(
            "() => document.getElementById('nw-status').textContent.includes('Saved')",
            timeout=int(SAVE_DELAY_S * 1000) + 4000)
        disarmed = page.evaluate(
            "() => { const e = new Event('beforeunload', {cancelable: true});"
            " window.dispatchEvent(e); return e.defaultPrevented; }")
        check("guard DISARMED after a successful save", disarmed is False)
        ctx.close()
        ctx.close()

        browser.close()
    httpd.shutdown()

    print("\n" + ("ALL CASES PASSED" if not FAILS else f"{len(FAILS)} FAILED: {FAILS}"))
    return 1 if FAILS else 0


if __name__ == "__main__":
    sys.exit(main())
