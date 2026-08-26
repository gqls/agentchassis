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
    next_media_id = 900
    chunk_next = 0
    chunk_uploads = {}
    chunk_aborts = []
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
        elif self.path.startswith("/api/notes/") and self.path.endswith("/media/uploads"):
            # Chunked begin (server_uploads.go's shape). MUST precede the
            # plain-"/media" branch below, which its path also matches. The
            # stub halves the file so every chunked case exercises >=2 parts;
            # the client's contract is to obey part_size, whatever it is.
            body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", 0) or 0)) or b"{}")
            StubAPI.chunk_next += 1
            uid = StubAPI.chunk_next
            part_size = max(1024, (int(body.get("size", 0)) + 1) // 2)
            StubAPI.chunk_uploads[uid] = {"size": int(body.get("size", 0)),
                                          "part_size": part_size, "parts": {}, "kind": body.get("kind")}
            total = (int(body.get("size", 0)) + part_size - 1) // part_size
            self._json(201, {"upload_id": uid, "part_size": part_size, "parts_total": total})
        elif self.path.startswith("/api/uploads/") and self.path.endswith("/finish"):
            uid = int(self.path.split("/")[3])
            u = StubAPI.chunk_uploads.get(uid)
            if u is None:
                self._json(404, {"error": "no such upload"})
                return
            total = (u["size"] + u["part_size"] - 1) // u["part_size"]
            for n in range(1, total + 1):
                if n not in u["parts"]:
                    self._json(409, {"error": "part %d has not arrived" % n})
                    return
            if sum(u["parts"].values()) != u["size"]:
                self._json(409, {"error": "the parts do not add up to the declared size"})
                return
            del StubAPI.chunk_uploads[uid]
            StubAPI.next_media_id += 1
            self._json(201, {"id": StubAPI.next_media_id, "byte_len": u["size"]})
        elif self.path.startswith("/api/notes/") and "/media" in self.path:
            body = self.rfile.read(int(self.headers.get("Content-Length", 0) or 0))
            time.sleep(MEDIA_DELAY_S)         # the observable window for "Uploading…"
            StubAPI.next_media_id += 1
            self._json(201, {"id": StubAPI.next_media_id, "byte_len": len(body)})
        else:
            self._json(404, {"error": "no such path"})

    def do_PUT(self):
        if self.path.startswith("/api/uploads/") and "/parts/" in self.path:
            uid = int(self.path.split("/")[3])
            n = int(self.path.split("/")[5])
            body = self.rfile.read(int(self.headers.get("Content-Length", 0) or 0))
            u = StubAPI.chunk_uploads.get(uid)
            if u is None:
                self._json(404, {"error": "no such upload"})
                return
            u["parts"][n] = len(body)
            self._json(200, {"part": n, "size": len(body)})
        else:
            self._json(404, {"error": "no such path"})

    def do_DELETE(self):
        if self.path.startswith("/api/uploads/"):
            uid = int(self.path.split("/")[3])
            StubAPI.chunk_uploads.pop(uid, None)
            StubAPI.chunk_aborts.append(uid)
            self._json(200, {"status": "discarded"})
        elif self.path.startswith("/api/media/"):
            self._json(200, {"status": "deleted"})
        elif self.path == "/api/account":
            self.rfile.read(int(self.headers.get("Content-Length", 0) or 0))
            self._json(200, {"status": "account deleted"})
        else:
            self._json(404, {"error": "no such path"})

    def do_PATCH(self):
        if self.path.startswith("/api/media/"):
            body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", 0) or 0)) or b"{}")
            self._json(200, {"status": "saved", "caption": body.get("caption", "")})
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

        def route_me(page, max_upload=25 * 1024 * 1024, small_upload_max=None):
            calls = {"n": 0}
            def handler(r):
                calls["n"] += 1
                if calls["n"] == 1:
                    r.continue_()
                else:
                    me = {"email": "probe@example.com", "media_bytes": 1234,
                          "media_quota": 50 * 1024 * 1024, "max_upload": max_upload}
                    if small_upload_max is not None:
                        me["small_upload_max"] = small_upload_max
                    r.fulfill(status=200, content_type="application/json", body=json.dumps(me))
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

        # ------------------------------------------------------------------
        # THE BOARD (pasteboard stage 2). Cases K–L. Drags are synthesized
        # PointerEvent sequences (like Case F's beforeunload, this tests OUR
        # handlers; real hit-testing belongs to the manual walkthrough).
        # ------------------------------------------------------------------

        NOTE_WITH_LAYOUT = {
            "id": 7, "title": "Earlier thought", "content": "kept",
            "media": [{"id": 901, "kind": "image", "mime": "image/png", "byte_len": len(PNG_1PX)}],
            "layout": {"v": 1, "items": [
                {"id": "text", "kind": "text", "x": 0.02, "y": 0.02, "w": 0.96, "h": 0.3, "z": 1},
                {"id": "m901", "kind": "media", "media_id": 901, "x": 0.02, "y": 0.36, "w": 0.47, "h": 0.36, "z": 2}]}}

        def drag_item(page, item_id, dx, dy, pointer_type="mouse"):
            return page.evaluate(
                """([itemId, dx, dy, ptype]) => {
                    const item = document.querySelector('[data-item-id="' + itemId + '"]');
                    const handle = item.querySelector('.noted-write__bhandle');
                    const r = handle.getBoundingClientRect();
                    const opts = {bubbles: true, cancelable: true, pointerId: 7,
                                  pointerType: ptype, clientX: r.x + 10, clientY: r.y + 10};
                    handle.dispatchEvent(new PointerEvent('pointerdown', opts));
                    document.dispatchEvent(new PointerEvent('pointermove',
                        Object.assign({}, opts, {clientX: opts.clientX + dx, clientY: opts.clientY + dy})));
                    document.dispatchEvent(new PointerEvent('pointerup', opts));
                    const after = document.querySelector('[data-item-id="' + itemId + '"]');
                    return after.style.left;
                }""", [item_id, dx, dy, pointer_type])

        # ---------- CASE K: the board — arrange, save, reopen arranged ----------
        print("\nCASE K — board: first arrangement marks dirty; drag moves; layout rides the save; reopen restores")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page)
        saved_bodies = []
        page.route("**/api/notes", lambda r: (
            saved_bodies.append(json.loads(r.request.post_data or "{}")),
            r.fulfill(status=200, content_type="application/json",
                      body=json.dumps(dict(json.loads(r.request.post_data or "{}"), id=7))))
            if r.request.method == "POST" else r.continue_())
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        page.click("#nw-list >> text=Earlier thought")
        page.wait_for_selector("#nw-media img", timeout=5000)
        check("toggle offered on a note with media", page.locator("#nw-view-toggle").is_visible())
        page.click("#nw-view-toggle")
        page.wait_for_selector("#nw-board:not([hidden])", timeout=3000)
        check("board shows text + media items", page.locator(".noted-write__bitem").count() == 2)
        check("first arrangement marks the note unsaved",
              page.locator("#nw-status").inner_text() == "Unsaved changes")

        before = page.locator('[data-item-id="m901"]').evaluate("n => n.style.left")
        after = drag_item(page, "m901", 120, 60)
        check("drag moved the item", after != before, f"{before} -> {after}")
        check("handle is a real touch target (>=44px)",
              page.locator('[data-item-id="m901"] .noted-write__bhandle').evaluate(
                  "n => n.getBoundingClientRect().height >= 44"))

        page.click("#nw-save")
        page.wait_for_function(
            "() => document.getElementById('nw-status').textContent.includes('Saved')", timeout=6000)
        body = saved_bodies[-1]
        lay = body.get("layout") or {}
        items = {i["id"]: i for i in lay.get("items", [])}
        check("layout rode the save", lay.get("v") == 1 and "m901" in items and "text" in items)
        ok_frac = all(0 <= items["m901"][k] <= 1.5 for k in ("x", "y", "w", "h"))
        check("coordinates are fractions of board width", ok_frac, str(items.get("m901")))
        ctx.close()

        # a note that ARRIVES with a layout opens straight onto the board
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page)
        page.route("**/api/notes", lambda r: r.fulfill(
            status=200, content_type="application/json",
            body=json.dumps({"notes": [NOTE_WITH_LAYOUT]}))
            if r.request.method == "GET" else r.continue_())
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        page.click("#nw-list >> text=Earlier thought")
        page.wait_for_selector("#nw-board:not([hidden])", timeout=3000)
        check("arranged note reopens in board mode", page.locator(".noted-write__bitem").count() == 2)
        check("reopening is not itself an edit",
              "Opened" in page.locator("#nw-status").inner_text())
        ctx.close()

        # ---------- CASE L: the board on a PHONE ----------
        print("\nCASE L — board on a phone: fits the viewport, touch-drag works")
        ctx = browser.new_context(viewport={"width": 390, "height": 844},
                                  has_touch=True, is_mobile=True)
        page = ctx.new_page()
        route_me(page)
        page.route("**/api/notes", lambda r: r.fulfill(
            status=200, content_type="application/json",
            body=json.dumps({"notes": [NOTE_WITH_LAYOUT]}))
            if r.request.method == "GET" else r.continue_())
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        page.click("#nw-list >> text=Earlier thought")
        page.wait_for_selector("#nw-board:not([hidden])", timeout=3000)
        check("no horizontal overflow at 390px",
              page.evaluate("() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1"))
        before = page.locator('[data-item-id="m901"]').evaluate("n => n.style.left")
        after = drag_item(page, "m901", 60, 40, "touch")
        check("touch-pointer drag moved the item", after != before, f"{before} -> {after}")
        check("drag marked it unsaved on mobile too",
              page.locator("#nw-status").inner_text() == "Unsaved changes")
        check("handles opt out of scroll (touch-action none)",
              page.locator('[data-item-id="m901"] .noted-write__bhandle').evaluate(
                  "n => getComputedStyle(n).touchAction === 'none'"))
        ctx.close()

        # ------------------------------------------------------------------
        # DELETION + EDITING (immediate deletion; stage 3 seed). Cases M–O.
        # ------------------------------------------------------------------

        # ---------- CASE M: account deletion is honest ----------
        print("\nCASE M — deletion: wrong password loud, failure loud, gone only on the 2xx")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page)
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        page.click("#nw-account")
        page.wait_for_selector("#nw-account-panel:not([hidden])", timeout=3000)

        page.click("#nw-del-confirm")   # no password typed
        page.wait_for_selector("#nw-del-error:not([hidden])", timeout=3000)
        check("empty password refused locally", "Type your password" in page.locator("#nw-del-error").inner_text())

        page.fill("#nw-del-password", "wrong-password-here")
        page.route("**/api/account", lambda r: r.fulfill(
            status=401, content_type="application/json",
            body=json.dumps({"error": "that password is not right — your account is untouched"}))
            if r.request.method == "DELETE" else r.continue_())
        page.once("dialog", lambda d: d.accept())
        page.click("#nw-del-confirm")
        page.wait_for_function(
            "() => document.getElementById('nw-del-error').textContent.includes('not right')", timeout=4000)
        check("engine's refusal shown verbatim, still signed in",
              page.locator("#nw-app").is_visible() and "Nothing has been deleted" in page.locator("#nw-del-error").inner_text())

        page.unroute("**/api/account")
        page.route("**/api/account", lambda r: r.abort() if r.request.method == "DELETE" else r.continue_())
        page.once("dialog", lambda d: d.accept())
        page.fill("#nw-del-password", "0123456789x")
        page.click("#nw-del-confirm")
        page.wait_for_function(
            "() => !document.getElementById('nw-del-error').hidden", timeout=4000)
        check("network failure: loud, still signed in", page.locator("#nw-app").is_visible())

        page.unroute("**/api/account")
        page.once("dialog", lambda d: d.accept())
        page.click("#nw-del-confirm")
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=4000)
        check("2xx: signed out with the goodbye", page.locator("#nw-goodbye").is_visible())
        ctx.close()

        # ---------- CASE N: captions kept only on the 2xx ----------
        print("\nCASE N — captions: saved via the server, shown; failure loud")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page)
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        page.click("#nw-list >> text=Earlier thought")
        page.wait_for_selector("#nw-media img", timeout=5000)
        page.once("dialog", lambda d: d.accept("the garden in May"))
        page.click("#nw-media >> text=Add caption")
        page.wait_for_function(
            "() => document.getElementById('nw-media').textContent.includes('the garden in May')", timeout=4000)
        check("caption shown after the 2xx", True)

        page.route("**/api/media/**", lambda r: r.abort() if r.request.method == "PATCH" else r.continue_())
        page.once("dialog", lambda d: d.accept("lost caption"))
        page.click("#nw-media >> text=Caption…")
        page.wait_for_function(
            "() => document.getElementById('nw-media-note').textContent.includes('NOT saved')", timeout=4000)
        check("failed caption is loud and the old caption stays",
              "the garden in May" in page.locator("#nw-media").inner_text())
        page.unroute("**/api/media/**")
        ctx.close()

        # ---------- CASE O: editing never destroys — copy first, original after ----------
        print("\nCASE O — edit: rotate+save uploads a COPY, original removed only on 2xx")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page)
        calls = []
        page.on("request", lambda r: calls.append((r.method, r.url)) if "/media" in r.url or "/api/notes/" in r.url else None)
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        page.click("#nw-list >> text=Earlier thought")
        page.wait_for_selector("#nw-media img", timeout=5000)
        page.click("#nw-media >> text=Edit")
        page.wait_for_selector("#nw-edit-panel:not([hidden])", timeout=4000)
        page.click("#nw-edit-rotate")
        page.click("#nw-edit-save")
        page.wait_for_function(
            "() => document.getElementById('nw-edit-panel').hidden", timeout=int(MEDIA_DELAY_S * 1000) + 6000)
        posts = [c for c in calls if c[0] == "POST" and "/media" in c[1]]
        deletes = [c for c in calls if c[0] == "DELETE" and "/api/media/901" in c[1]]
        check("edited copy uploaded as NEW media", len(posts) == 1)
        check("original removed only after the copy stored", len(deletes) == 1)
        check("strip shows the copy, not the original",
              "/api/media/901" not in (page.locator("#nw-media img").get_attribute("src") or ""))

        # failure branch: the upload dies -> the ORIGINAL is untouched
        page.route("**/api/notes/**", lambda r: r.abort() if r.request.method == "POST" else r.continue_())
        page.click("#nw-media >> text=Edit")
        page.wait_for_selector("#nw-edit-panel:not([hidden])", timeout=4000)
        page.click("#nw-edit-rotate")
        page.click("#nw-edit-save")
        page.wait_for_function(
            "() => document.getElementById('nw-edit-status').textContent.includes('NOT saved')", timeout=6000)
        check("failed edit: loud, original untouched",
              "original is untouched" in page.locator("#nw-edit-status").inner_text()
              and page.locator("#nw-media img").count() == 1)
        ctx.close()

        # ------------------------------------------------------------------
        # CHUNKED UPLOADS (2026-08-26, PLAN large_uploads). Cases P–S. The
        # split point (small_upload_max) is served tiny via route_me so a
        # 10 KB buffer exercises the multi-part path fast.
        # ------------------------------------------------------------------

        def chunked_reqs(reqs):
            return [(m, u) for (m, u) in reqs if "/media/uploads" in u or "/api/uploads/" in u]

        # ---------- CASE P: begin → parts → finish, STORED only at the end ----------
        print("\nCASE P — chunked upload: begin → both parts → finish, in that order; stored at the end")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page, max_upload=10 * 1024 * 1024, small_upload_max=4096)
        reqs = []
        page.on("request", lambda r: reqs.append((r.method, r.url)))
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        add_png(page, name="big.png", buffer=bytes(10000))   # > 4096 → chunked
        page.wait_for_selector("#nw-media img", timeout=15000)
        seq = chunked_reqs(reqs)
        check("exactly begin, part 1, part 2, finish — in order",
              [("POST" == m and "/media/uploads" in u) or True for (m, u) in seq] and
              len(seq) == 4 and
              seq[0][0] == "POST" and "/media/uploads" in seq[0][1] and
              seq[1][0] == "PUT" and "/parts/1" in seq[1][1] and
              seq[2][0] == "PUT" and "/parts/2" in seq[2][1] and
              seq[3][0] == "POST" and seq[3][1].endswith("/finish"),
              detail=str(seq))
        check("no single-shot POST for a file this size",
              not [1 for (m, u) in reqs if m == "POST" and "/media?kind=" in u])
        check("stored item renders after finish",
              "/api/media/" in page.locator("#nw-media img").get_attribute("src"))
        check("no pending item left", "Uploading" not in page.locator("#nw-media").inner_text())
        ctx.close()

        # ---------- CASE Q: one part fails once — retried alone, then success ----------
        print("\nCASE Q — a failed part retries by itself; the upload still completes")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page, max_upload=10 * 1024 * 1024, small_upload_max=4096)
        reqs = []
        page.on("request", lambda r: reqs.append((r.method, r.url)))
        flaky = {"failed": 0}
        def fail_part2_once(r):
            if r.request.method == "PUT" and "/parts/2" in r.request.url and flaky["failed"] == 0:
                flaky["failed"] += 1
                r.abort()
            else:
                r.continue_()
        page.route("**/api/uploads/**", fail_part2_once)
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        add_png(page, name="big.png", buffer=bytes(10000))
        page.wait_for_selector("#nw-media img", timeout=15000)
        part2 = [1 for (m, u) in reqs if m == "PUT" and "/parts/2" in u]
        check("part 2 was sent twice (one failure, one retry)", len(part2) == 2, detail=str(len(part2)))
        check("part 1 was NOT re-sent", len([1 for (m, u) in reqs if m == "PUT" and "/parts/1" in u]) == 1)
        check("upload completed despite the flaky part",
              "/api/media/" in page.locator("#nw-media img").get_attribute("src"))
        ctx.close()

        # ---------- CASE R: a hard failure is LOUD, held, and releases the reservation ----------
        print("\nCASE R — hard part failure: NOT stored said plainly, bytes held, reservation released")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page, max_upload=10 * 1024 * 1024, small_upload_max=4096)
        reqs = []
        page.on("request", lambda r: reqs.append((r.method, r.url)))
        page.route("**/api/uploads/**",
                   lambda r: r.abort() if r.request.method == "PUT" else r.continue_())
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        add_png(page, name="big.png", buffer=bytes(10000))
        page.wait_for_function(
            "() => document.getElementById('nw-media').textContent.includes('NOT stored')", timeout=15000)
        check("failure is loud and the item is held with Try again",
              page.locator("#nw-media >> text=Try again").count() == 1)
        check("three attempts were made for the failing part",
              len([1 for (m, u) in reqs if m == "PUT" and "/parts/1" in u]) == 3)
        check("the reservation was released (DELETE /api/uploads/{id})",
              any(m == "DELETE" and "/api/uploads/" in u for (m, u) in reqs))
        check("guard still ARMED — the bytes are held, not lost", guard_armed(page) is True)
        check("no <img> was ever claimed", page.locator("#nw-media img").count() == 0)
        ctx.close()

        # ---------- CASE S: a small file still takes the proven single request ----------
        print("\nCASE S — under the split point nothing changes: one POST, no chunk machinery")
        ctx = browser.new_context()
        page = ctx.new_page()
        route_me(page, max_upload=10 * 1024 * 1024, small_upload_max=4096)
        reqs = []
        page.on("request", lambda r: reqs.append((r.method, r.url)))
        page.goto(url)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=5000)
        sign_in(page)
        add_png(page)   # the 1px PNG, far under 4096
        page.wait_for_selector("#nw-media img", timeout=15000)
        check("single-shot POST used", bool([1 for (m, u) in reqs if m == "POST" and "/media?kind=" in u]))
        check("no chunk machinery touched", len(chunked_reqs(reqs)) == 0)
        ctx.close()

        browser.close()
    httpd.shutdown()

    print("\n" + ("ALL CASES PASSED" if not FAILS else f"{len(FAILS)} FAILED: {FAILS}"))
    return 1 if FAILS else 0


if __name__ == "__main__":
    sys.exit(main())
