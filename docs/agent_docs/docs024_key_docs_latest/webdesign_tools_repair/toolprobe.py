#!/usr/bin/env python3
"""Real-browser probe for webdesign.co.uk tool pages.

Loads a page in headless chromium via the DevTools protocol and reports what a
VISITOR gets: JS console errors, whether the tool's controls exist, and whether
driving the first control changes the page. Static source reading cannot answer
any of those — this is the witness.

Usage: toolprobe.py <url> [<url> ...]
"""
import json
import os
import subprocess
import sys
import tempfile
import time
import urllib.request
import http.client

CHROMIUM = "/snap/bin/chromium"


def start_chrome(port, profile):
    p = subprocess.Popen(
        [CHROMIUM, "--headless=new", f"--remote-debugging-port={port}",
         f"--user-data-dir={profile}", "--no-sandbox", "--disable-gpu",
         "--disable-dev-shm-usage", "--window-size=1280,900", "about:blank"],
        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(60):
        try:
            urllib.request.urlopen(f"http://127.0.0.1:{port}/json/version", timeout=1).read()
            return p
        except Exception:
            time.sleep(0.5)
    raise RuntimeError("chromium did not start")


class CDP:
    """Minimal CDP client over a raw websocket (no third-party deps)."""

    def __init__(self, ws_url):
        import base64
        import socket
        u = ws_url.split("://", 1)[1]
        host_port, path = u.split("/", 1)
        host, port = host_port.split(":")
        self.sock = socket.create_connection((host, int(port)), timeout=30)
        key = base64.b64encode(os.urandom(16)).decode()
        req = (f"GET /{path} HTTP/1.1\r\nHost: {host_port}\r\nUpgrade: websocket\r\n"
               f"Connection: Upgrade\r\nSec-WebSocket-Key: {key}\r\n"
               f"Sec-WebSocket-Version: 13\r\n\r\n")
        self.sock.sendall(req.encode())
        buf = b""
        while b"\r\n\r\n" not in buf:
            buf += self.sock.recv(4096)
        self.buf = buf.split(b"\r\n\r\n", 1)[1]
        self.msg_id = 0

    def _send(self, payload):
        data = json.dumps(payload).encode()
        header = bytearray([0x81])
        n = len(data)
        mask = os.urandom(4)
        if n < 126:
            header.append(0x80 | n)
        elif n < 65536:
            header.append(0x80 | 126)
            header += n.to_bytes(2, "big")
        else:
            header.append(0x80 | 127)
            header += n.to_bytes(8, "big")
        header += mask
        masked = bytes(b ^ mask[i % 4] for i, b in enumerate(data))
        self.sock.sendall(bytes(header) + masked)

    def _recv_frame(self, timeout=None):
        if timeout is not None:
            self.sock.settimeout(timeout)

        def need(n):
            while len(self.buf) < n:
                chunk = self.sock.recv(65536)
                if not chunk:
                    raise RuntimeError("socket closed")
                self.buf += chunk
        need(2)
        b1, b2 = self.buf[0], self.buf[1]
        ln = b2 & 0x7F
        off = 2
        if ln == 126:
            need(4)
            ln = int.from_bytes(self.buf[2:4], "big")
            off = 4
        elif ln == 127:
            need(10)
            ln = int.from_bytes(self.buf[2:10], "big")
            off = 10
        need(off + ln)
        payload = self.buf[off:off + ln]
        self.buf = self.buf[off + ln:]
        return json.loads(payload.decode("utf-8", "replace"))

    def call(self, method, params=None, timeout=30):
        self.msg_id += 1
        mid = self.msg_id
        self._send({"id": mid, "method": method, "params": params or {}})
        end = time.time() + timeout
        # Set our OWN deadline on every read: the probe loop leaves a short
        # socket timeout behind, and inheriting it turned a quiet page into a
        # false "probe failed: timed out" — a harness fault wearing a site
        # fault's clothes (NOTES 2026-07-29).
        while time.time() < end:
            try:
                msg = self._recv_frame(timeout=max(0.5, min(5.0, end - time.time())))
            except Exception:
                continue
            if msg.get("id") == mid:
                return msg
        raise RuntimeError(f"timeout waiting for {method}")


PROBE_JS = r"""
(() => {
  const out = {errors: [], controls: 0, buttons: 0, canvases: 0, outputs: 0,
               changed: null, note: ''};
  out.controls = document.querySelectorAll('main input, main select, main textarea').length;
  out.buttons  = document.querySelectorAll('main button').length;
  out.canvases = document.querySelectorAll('main canvas').length;
  out.outputs  = document.querySelectorAll('main pre, main output, main [id*="out"], main [id*="result"], main [id*="status"]').length;
  // Compare the WHOLE of main, not its length: a tool that rewrites "3.5" as
  // "2.8" leaves the length identical, and reading that as "nothing happened"
  // marked working tools DEAD (NOTES 2026-07-29). Includes live form values,
  // which innerHTML does not reflect, and canvas pixels, which it also doesn't.
  // The DRIVEN element is excluded — otherwise the probe would be measuring
  // its own keystroke instead of the tool's response.
  // Try each kind IN PREFERENCE ORDER. A comma-separated querySelector returns
  // the first match in DOCUMENT order, not selector order — so on any page
  // whose buttons precede its inputs the probe was clicking a button (often an
  // already-selected radio, or "Undo" with nothing to undo) and recording DEAD
  // for a working tool. svg-patterns and several others (NOTES 2026-07-29).
  // Buttons last, and skip ones that look inert to click.
  const PREF = ['main input[type=range]', 'main input[type=number]',
                'main input[type=text]', 'main input[type=color]',
                'main textarea', 'main select',
                'main button:not(.active):not([disabled])', 'main button'];
  let el = null;
  for (const sel of PREF) {
    const cand = Array.from(document.querySelectorAll(sel))
      .filter(e => !/undo|redo|reset|clear|copy|download|share|print/i.test(
        (e.id || '') + ' ' + (e.className || '') + ' ' + (e.textContent || '')));
    if (cand.length) { el = cand[0]; break; }
  }
  if (!el) { out.note = 'no control to drive'; return out; }
  el.setAttribute('data-probe-driven', '1');
  const snap = () => document.querySelector('main').innerHTML + '|' +
    Array.from(document.querySelectorAll('main input, main textarea, main select'))
      .filter(e => !e.hasAttribute('data-probe-driven'))
      .map(e => e.value).join('~') + '|' +
    Array.from(document.querySelectorAll('main canvas'))
      .map(c => { try { return c.toDataURL().length; } catch (e) { return 'x'; } }).join('~');
  const before = snap();
  out.beforeSnap = before;
  out.note = 'drove ' + el.tagName.toLowerCase() + (el.type ? '[' + el.type + ']' : '') + (el.id ? '#' + el.id : '');
  try {
    if (el.tagName === 'BUTTON') { el.click(); }
    else if (el.tagName === 'SELECT') {
      if (el.options.length > 1) { el.selectedIndex = (el.selectedIndex + 1) % el.options.length; }
      el.dispatchEvent(new Event('change', {bubbles: true}));
      el.dispatchEvent(new Event('input', {bubbles: true}));
    } else if (el.type === 'range' || el.type === 'number') {
      const cur = parseFloat(el.value || '0');
      const min = parseFloat(el.min || '0'), max = parseFloat(el.max || String(cur + 10));
      let v = cur + Math.max(1, (max - min) / 4);
      if (v > max) v = min;
      el.value = String(v);
      el.dispatchEvent(new Event('input', {bubbles: true}));
      el.dispatchEvent(new Event('change', {bubbles: true}));
    } else {
      // Drive with input the tool would ACCEPT. Appending 'probe123' to a hex
      // colour field made smart-contrast look dead when it was correctly
      // rejecting nonsense — a tool that validates its input was being scored
      // as broken (NOTES 2026-07-29). Infer a plausible value from what the
      // field already holds, then fall back to generic text.
      const cur = (el.value || '').trim();
      const ph = (el.placeholder || '').trim();
      let next;
      if (/^#?[0-9a-f]{6}$/i.test(cur)) {
        next = (cur.startsWith('#') ? '#' : '') + (cur.replace('#','').toLowerCase() === '3b82f6' ? 'c2410c' : '3b82f6');
      } else if (/^-?\d+(\.\d+)?$/.test(cur)) {
        next = String(parseFloat(cur) + 7);
      } else if (/^\s*[{[]/.test(cur) || /json/i.test(el.id + ' ' + ph)) {
        next = '{"probe": 1, "list": [1,2,3]}';
      } else if (/</.test(cur) || /html|markup/i.test(el.id + ' ' + ph)) {
        next = '<div class="probe"><p>hello</p></div>';
      } else if (/svg/i.test(el.id + ' ' + ph)) {
        next = '<svg xmlns="http://www.w3.org/2000/svg" width="8" height="8"><circle cx="4" cy="4" r="3"/></svg>';
      } else if (cur) {
        next = cur + ' probe';           // extend, don't corrupt
      } else {
        next = 'probe text';
      }
      out.note += ' value=' + JSON.stringify(next).slice(0, 40);
      el.value = next;
      el.dispatchEvent(new Event('input', {bubbles: true}));
      el.dispatchEvent(new Event('change', {bubbles: true}));
      el.dispatchEvent(new KeyboardEvent('keyup', {bubbles: true, key: 'a'}));
    }
  } catch (e) { out.errors.push('drive threw: ' + e.message); }
  out.changed = snap() !== before;
  // Phase 2: many tools are paste-then-press, not live-update. If typing alone
  // changed nothing, press the action button and measure again — sri-generator
  // produced a provably correct SRI hash while being scored DEAD, purely
  // because nobody pressed Generate (NOTES 2026-07-29).
  if (!out.changed) {
    const ACTION = /generate|run|convert|build|calculate|format|minify|compile|analy|check|create|render|process|apply|split|extract|optimi|encode|decode|clean|parse|go\b|submit/i;
    const btn = Array.from(document.querySelectorAll('main button, main input[type=submit]'))
      .filter(b => !b.disabled)
      .filter(b => ACTION.test((b.textContent || '') + ' ' + (b.id || '') + ' ' + (b.className || '')))
      .filter(b => !/undo|redo|reset|clear|copy|download|share|print/i.test(
        (b.textContent || '') + ' ' + (b.id || '')))[0];
    if (btn) {
      out.note += ' | pressed ' + JSON.stringify((btn.textContent || btn.id || '').trim().slice(0, 24));
      try { btn.click(); } catch (e) { out.errors.push('click threw: ' + e.message); }
      out.pressed = true;
    }
  }
  return out;
})()
"""


def probe(cdp, url):
    cdp.call("Runtime.enable")
    cdp.call("Log.enable")
    cdp.call("Page.enable")
    cdp.call("Page.navigate", {"url": url})
    errors = []
    end = time.time() + 8
    loaded = False
    while time.time() < end:
        try:
            msg = cdp._recv_frame(timeout=max(0.3, min(2.0, end - time.time())))
        except Exception:
            if loaded:
                break
            continue
        m = msg.get("method")
        if m == "Runtime.exceptionThrown":
            d = msg["params"]["exceptionDetails"]
            txt = d.get("exception", {}).get("description") or d.get("text", "")
            errors.append(txt.split("\n")[0][:160])
        elif m == "Log.entryAdded":
            e = msg["params"]["entry"]
            if e.get("level") == "error":
                errors.append((e.get("text") or "")[:160])
        elif m == "Page.loadEventFired":
            loaded = True
            end = min(end, time.time() + 2.5)
    # Don't trust the event drain to tell us the page is ready: on some pages
    # loadEventFired never reached the drain and the probe reported a timeout
    # as if the SITE had hung (chromium --dump-dom rendered all of them fine).
    # Poll the document instead, which cannot lie about its own readyState.
    for _ in range(20):
        try:
            rs = cdp.call("Runtime.evaluate",
                          {"expression": "document.readyState", "returnByValue": True},
                          timeout=8)
            if rs.get("result", {}).get("result", {}).get("value") == "complete":
                break
        except Exception:
            pass
        time.sleep(0.5)
    res = cdp.call("Runtime.evaluate",
                   {"expression": PROBE_JS, "returnByValue": True, "awaitPromise": False})
    val = res.get("result", {}).get("result", {}).get("value") or {}
    # If phase 2 pressed an action button, give async work (crypto.subtle,
    # FileReader, image decode) time to land, then re-measure from scratch.
    if val.get("pressed"):
        time.sleep(1.2)
        res2 = cdp.call("Runtime.evaluate", {
            "expression": """(() => {
                const el = document.querySelector('main [data-probe-driven]');
                const snap = document.querySelector('main').innerHTML + '|' +
                  Array.from(document.querySelectorAll('main input, main textarea, main select'))
                    .filter(e => !e.hasAttribute('data-probe-driven'))
                    .map(e => e.value).join('~');
                return snap;
            })()""", "returnByValue": True}, timeout=15)
        after = res2.get("result", {}).get("result", {}).get("value")
        if after is not None and val.get("beforeSnap") is not None:
            val["changed"] = after != val["beforeSnap"]
    # collect post-interaction errors
    end2 = time.time() + 1.5
    while time.time() < end2:
        try:
            msg = cdp._recv_frame()
        except Exception:
            break
        if msg.get("method") == "Runtime.exceptionThrown":
            d = msg["params"]["exceptionDetails"]
            txt = d.get("exception", {}).get("description") or d.get("text", "")
            errors.append("after-interaction: " + txt.split("\n")[0][:140])
    val["errors"] = errors + val.get("errors", [])
    val["loaded"] = loaded
    return val


def main():
    urls = sys.argv[1:]
    if not urls:
        print("usage: toolprobe.py <url>...")
        return 2
    port = 9333
    profile = tempfile.mkdtemp(prefix="toolprobe-")
    chrome = start_chrome(port, profile)
    # A single chromium instance degrades after a handful of tabs (measured:
    # the DevTools HTTP endpoint stops answering around the 14th). Restart on a
    # fixed cadence AND on first failure rather than losing the run.
    batch = int(os.environ.get("PROBE_BATCH", "6"))
    try:
        results = []
        for i, url in enumerate(urls):
            if i and i % batch == 0:
                chrome.terminate()
                try:
                    chrome.wait(timeout=10)
                except Exception:
                    chrome.kill()
                profile = tempfile.mkdtemp(prefix="toolprobe-")
                chrome = start_chrome(port, profile)
            tabs = None
            for attempt in range(2):
                try:
                    req = urllib.request.Request(
                        f"http://127.0.0.1:{port}/json/new?about:blank", method="PUT")
                    tabs = json.loads(urllib.request.urlopen(req, timeout=15).read())
                    break
                except Exception:
                    chrome.kill()
                    profile = tempfile.mkdtemp(prefix="toolprobe-")
                    chrome = start_chrome(port, profile)
            if tabs is None:
                r = {"errors": ["probe failed: no tab"], "changed": None}
            else:
                try:
                    cdp = CDP(tabs["webSocketDebuggerUrl"])
                    r = probe(cdp, url)
                except Exception as e:
                    r = {"errors": [f"probe failed: {e}"], "changed": None}
            r["url"] = url
            results.append(r)
            if tabs:
                try:
                    http.client.HTTPConnection("127.0.0.1", port, timeout=5).request(
                        "GET", f"/json/close/{tabs['id']}")
                except Exception:
                    pass
            slug = url.rstrip("/").split("/")[-2] if url.endswith("index.html") else url
            verdict = ("BROKEN" if r.get("errors") else
                       ("DEAD" if r.get("changed") is False else
                        ("OK" if r.get("changed") else "NO-CONTROL")))
            print(f"{verdict:10} {slug:26} ctl={r.get('controls',0):2} btn={r.get('buttons',0):2} "
                  f"canvas={r.get('canvases',0)} changed={r.get('changed')} :: {r.get('note','')}")
            for e in r.get("errors", [])[:3]:
                print(f"           ! {e}")
        with open(os.environ.get("PROBE_JSON", "/dev/null"), "w") as f:
            json.dump(results, f, indent=1)
    finally:
        chrome.terminate()
    return 0


if __name__ == "__main__":
    sys.exit(main())
