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
MEDIA_DELAY_S = 0.8  # same idea for uploads: the "Uploading…" window must be provable

# A real 1x1 PNG, so a stored item renders as an actual <img> load, not a 404.
PNG_1PX = bytes.fromhex(
    "89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c489"
    "0000000d49444154789c626001000000ffff03000006000557bfabd40000000049454e44ae426082")


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
            self._json(200, {"notes": [{
                "id": 7, "title": "Earlier thought", "content": "kept",
                "media": [{"id": 901, "kind": "image", "mime": "image/png", "byte_len": len(PNG_1PX)}]}]})
        elif self.path.startswith("/api/media/"):
            self.send_response(200)
            self.send_header("Content-Type", "image/png")
            self.send_header("Content-Length", str(len(PNG_1PX)))
            self.end_headers()
            self.wfile.write(PNG_1PX)
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
        elif self.path.startswith("/api/notes/") and "/media" in self.path:
            body = self.rfile.read(int(self.headers.get("Content-Length", 0) or 0))
            time.sleep(MEDIA_DELAY_S)         # the observable window for "Uploading…"
            self._json(201, {"id": 901, "byte_len": len(body)})
        else:
            self._json(404, {"error": "no such path"})

    def do_DELETE(self):
        if self.path.startswith("/api/media/"):
            self._json(200, {"status": "deleted"})
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

        # ------------------------------------------------------------------
        # MEDIA (pasteboard stage 1). Same contract family, cases G–J.
        # /api/me is routed per-context: first call passes through (401, so the
        # auth form appears), later calls answer with the account's limits —
        # that is how the storage meter and the size pre-check get exercised.
        # ------------------------------------------------------------------

        def route_me(page, max_upload=25 * 1024 * 1024):
            calls = {"n": 0}
            def handler(r):
                calls["n"] += 1
                if calls["n"] == 1:
                    r.continue_()
                else:
                    r.fulfill(status=200, content_type="application/json",
                              body=json.dumps({"email": "probe@example.com", "media_bytes": 1234,
                                               "media_quota": 50 * 1024 * 1024, "max_upload": max_upload}))
            page.route("**/api/me", handler)

        def guard_armed(page):
            return page.evaluate(
                "() => { const e = new Event('beforeunload', {cancelable: true});"
                " window.dispatchEvent(e); return e.defaultPrevented; }")

        def add_png(page, name="photo.png", buffer=None):
            page.set_input_files("#nw-file", files=[{
                "name": name, "mimeType": "image/png", "buffer": buffer or PNG_1PX}])

        # ---------- CASE G: an upload is honest, and a NEW note saves itself first ----------
        print("\nCASE G — upload honesty: 'Uploading…' until the 2xx; a new note saves first")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page)
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        add_png(page)          # note is NEW: no text typed, no id yet
        page.wait_for_timeout(int(SAVE_DELAY_S * 300))
        check("item shows 'Uploading…' while in flight",
              "Uploading" in page.locator("#nw-media").inner_text())
        check("no <img> before the server's 2xx", page.locator("#nw-media img").count() == 0)
        check("guard ARMED by an in-flight upload alone (text is clean)", guard_armed(page) is True)
        page.wait_for_selector("#nw-media img", timeout=int((SAVE_DELAY_S + MEDIA_DELAY_S) * 1000) + 5000)
        check("stored item renders from /api/media/",
              "/api/media/" in page.locator("#nw-media img").get_attribute("src"))
        check("the note was saved first, through the save path",
              "Saved ✓" in page.locator("#nw-status").inner_text())
        check("no pending item left", "Uploading" not in page.locator("#nw-media").inner_text())
        check("guard DISARMED once stored", guard_armed(page) is False)
        page.wait_for_selector("#nw-storage:not([hidden])", timeout=3000)
        check("storage meter reads the account's numbers",
              "of 50 MB" in page.locator("#nw-storage").inner_text())
        ctx.close()

        # ---------- CASE H: a failed upload is LOUD, held, and retryable ----------
        print("\nCASE H — failed upload: loud, bytes held, Try again works; engine message verbatim")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page)
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        page.route("**/api/notes/**", lambda r: r.abort() if r.request.method == "POST" else r.continue_())
        page.fill("#nw-content", "media must not eat the text")
        add_png(page)
        page.wait_for_function(
            "() => document.getElementById('nw-media').textContent.includes('NOT stored')",
            timeout=int(SAVE_DELAY_S * 1000) + 6000)
        check("failure named plainly", "NOT stored" in page.locator("#nw-media").inner_text())
        check("no <img> claimed", page.locator("#nw-media img").count() == 0)
        check("text untouched by the media failure",
              page.locator("#nw-content").input_value() == "media must not eat the text")
        check("guard ARMED while a failed upload is held", guard_armed(page) is True)
        page.unroute("**/api/notes/**")
        page.route("**/api/notes/**", lambda r: r.fulfill(
            status=507, content_type="application/json",
            body=json.dumps({"error": "you have used all your storage for recordings and photos"}))
            if r.request.method == "POST" else r.continue_())
        page.click("#nw-media >> text=Try again")
        page.wait_for_function(
            "() => document.getElementById('nw-media').textContent.includes('all your storage')",
            timeout=5000)
        check("engine's own refusal shown verbatim",
              "you have used all your storage" in page.locator("#nw-media").inner_text())
        page.unroute("**/api/notes/**")
        page.click("#nw-media >> text=Try again")
        page.wait_for_selector("#nw-media img", timeout=int(MEDIA_DELAY_S * 1000) + 5000)
        check("retry stored the SAME bytes", page.locator("#nw-media img").count() == 1)
        check("pending entry cleared after the 2xx", "NOT stored" not in page.locator("#nw-media").inner_text())
        ctx.close()

        # ---------- CASE I: removing asks first, and leaves only on the server's 2xx ----------
        print("\nCASE I — remove: asks first; refusal keeps the item; 2xx removes it")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page)
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        page.click("#nw-list >> text=Earlier thought")
        page.wait_for_selector("#nw-media img", timeout=5000)
        deletes = []
        page.on("request", lambda r: deletes.append(r.url) if r.method == "DELETE" else None)

        page.once("dialog", lambda d: d.dismiss())
        page.click(".noted-write__mremove")
        page.wait_for_timeout(300)
        check("declining the confirm removes nothing", page.locator("#nw-media img").count() == 1)
        check("…and sends nothing", len(deletes) == 0)

        page.route("**/api/media/**", lambda r: r.abort() if r.request.method == "DELETE" else r.continue_())
        page.once("dialog", lambda d: d.accept())
        page.click(".noted-write__mremove")
        page.wait_for_function(
            "() => document.getElementById('nw-media-note').textContent.includes('NOT removed')",
            timeout=5000)
        check("failed delete is loud and the item STAYS", page.locator("#nw-media img").count() == 1)

        page.unroute("**/api/media/**")
        page.once("dialog", lambda d: d.accept())
        page.click(".noted-write__mremove")
        page.wait_for_function("() => document.querySelectorAll('#nw-media img').length === 0", timeout=5000)
        check("2xx removes the item", page.locator("#nw-media img").count() == 0)
        ctx.close()

        # ---------- CASE J: refusals happen BEFORE any bytes travel ----------
        print("\nCASE J — alien type and oversized file are refused with no request")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page, max_upload=10 * 1024)   # 10 KB cap, so oversize is testable
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        page.wait_for_selector("#nw-storage:not([hidden])", timeout=3000)  # limits have arrived
        uploads = []
        page.on("request", lambda r: uploads.append(r.url) if "/media" in r.url and r.method == "POST" else None)

        page.set_input_files("#nw-file", files=[{
            "name": "notes.pdf", "mimeType": "application/pdf", "buffer": b"%PDF-1.4 not media"}])
        page.wait_for_selector("#nw-media-note:not([hidden])", timeout=3000)
        check("alien type refused in plain words",
              "images, GIFs, video and audio only" in page.locator("#nw-media-note").inner_text())

        add_png(page, name="huge.png", buffer=b"x" * (20 * 1024))
        page.wait_for_function(
            "() => document.getElementById('nw-media-note').textContent.includes('the most a single file')",
            timeout=3000)
        check("oversized file refused before travelling",
              "the most a single file can be is 10 KB" in page.locator("#nw-media-note").inner_text())
        page.wait_for_timeout(300)
        check("no upload request was ever made", len(uploads) == 0)
        ctx.close()

        browser.close()
    httpd.shutdown()

    print("\n" + ("ALL CASES PASSED" if not FAILS else f"{len(FAILS)} FAILED: {FAILS}"))
    return 1 if FAILS else 0


if __name__ == "__main__":
    sys.exit(main())
