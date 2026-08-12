#!/usr/bin/env python3
"""
Probe for the /legacy rescue tool (noted-legacy-rescue.html).

WHY A PLAYWRIGHT PROBE AND NOT THE PLATFORM'S CHECK RUNNER: the runner cannot seed
IndexedDB, so the one behaviour that matters here — what the page does when a real
NotedDB is present — is inexpressible as an experience check (HANDOFF §5). This is
the substitute, and it is the only test that exercises the code against data.

Run:  /home/ant/.venvs/vonc_pw/bin/python test_legacy_rescue.py
Exit 0 = all cases passed.

Every case asserts a NEGATIVE as well as a positive, because the failure mode that
matters is "reports success while having read nothing". The empty-database case in
particular asserts that we LEFT NO DATABASE BEHIND — opening a non-existent
IndexedDB creates one, and the tool aborts that creation.
"""
import http.server
import json
import socketserver
import sys
import threading
from pathlib import Path

from playwright.sync_api import sync_playwright

HERE = Path(__file__).parent
TOOL_HTML = (HERE / "noted-legacy-rescue.html").read_text()

PAGE = "<!doctype html><html><head><meta charset='utf-8'><title>legacy</title></head><body>" + TOOL_HTML + "</body></html>"

# --- the legacy schema, exactly as gqls/sites noted.co.uk/js/storage.js writes it
SEED_JS = """
async (spec) => {
  await new Promise((resolve, reject) => {
    const req = indexedDB.open('NotedDB', 4);
    req.onupgradeneeded = (e) => {
      const db = e.target.result;
      if (!db.objectStoreNames.contains('notes')) db.createObjectStore('notes', { keyPath: 'id' });
      if (!db.objectStoreNames.contains('history')) {
        const h = db.createObjectStore('history', { keyPath: 'revId', autoIncrement: true });
        h.createIndex('noteId', 'noteId', { unique: false });
      }
      if (!db.objectStoreNames.contains('audio')) db.createObjectStore('audio', { keyPath: 'noteId' });
      if (!db.objectStoreNames.contains('images')) db.createObjectStore('images', { keyPath: 'noteId' });
    };
    req.onsuccess = () => {
      const db = req.result;
      const tx = db.transaction(['notes','history','audio','images'], 'readwrite');
      spec.notes.forEach(n => tx.objectStore('notes').put(n));
      spec.history.forEach(h => tx.objectStore('history').put(h));
      // current shape: {noteId, items:[Blob]}
      spec.audioItems.forEach(a =>
        tx.objectStore('audio').put({ noteId: a.noteId, items: a.parts.map(p => new Blob([p], {type:'audio/webm'})) }));
      // LEGACY shape: {noteId, blob:Blob} — the one a naive reader silently drops
      spec.audioLegacy.forEach(a =>
        tx.objectStore('audio').put({ noteId: a.noteId, blob: new Blob([a.part], {type:'audio/webm'}) }));
      spec.imageItems.forEach(i =>
        tx.objectStore('images').put({ noteId: i.noteId, items: i.parts.map(p => new Blob([p], {type:'image/jpeg'})) }));
      tx.oncomplete = () => { db.close(); resolve(); };
      tx.onerror = () => reject(tx.error);
    };
    req.onerror = () => reject(req.error);
  });
}
"""

LIST_DBS = "async () => (indexedDB.databases ? (await indexedDB.databases()).map(d => d.name) : ['<unsupported>'])"


def serve(directory, port_holder):
    handler = lambda *a, **kw: http.server.SimpleHTTPRequestHandler(*a, directory=str(directory), **kw)
    httpd = socketserver.TCPServer(("127.0.0.1", 0), handler)
    port_holder.append(httpd.server_address[1])
    threading.Thread(target=httpd.serve_forever, daemon=True).start()
    return httpd


FAILS = []


def check(label, ok, detail=""):
    print(("  PASS  " if ok else "  FAIL  ") + label + (("  — " + detail) if detail else ""))
    if not ok:
        FAILS.append(label)


def main():
    tmp = HERE / ".probe"
    tmp.mkdir(exist_ok=True)
    (tmp / "legacy.html").write_text(PAGE)
    ports = []
    serve(tmp, ports)
    url = f"http://127.0.0.1:{ports[0]}/legacy.html"

    with sync_playwright() as p:
        browser = p.chromium.launch()

        # ---------- CASE 1: no database at all (a new visitor) ----------
        print("\nCASE 1 — no NotedDB in this browser")
        ctx = browser.new_context()
        page = ctx.new_page()
        page.goto(url)
        page.wait_for_timeout(700)
        check("shows the 'nothing here' panel", page.locator("#lr-empty").is_visible())
        check("does NOT show the found panel", not page.locator("#lr-found").is_visible())
        check("does NOT show an error", not page.locator("#lr-error").is_visible())
        dbs = page.evaluate(LIST_DBS)
        check("LEAVES NO DATABASE BEHIND", "NotedDB" not in dbs, f"databases now: {dbs}")
        ctx.close()

        # ---------- CASE 2: a real database, both media shapes ----------
        print("\nCASE 2 — notes, recordings (both record shapes), photos, history")
        ctx = browser.new_context()
        page = ctx.new_page()
        page.goto(url)
        spec = {
            "notes": [
                {"id": "abc-123", "title": "Shopping", "content": "milk", "updatedAt": "2026-01-02T03:04:05Z"},
                {"id": "def-456", "title": "Ideas", "content": "a thought"},
                {"id": "ghi-789", "title": "Plain", "content": "no media"},
            ],
            "history": [
                {"noteId": "abc-123", "title": "Shopping", "content": "mil"},
                {"noteId": "abc-123", "title": "Shopping", "content": "mi"},
            ],
            "audioItems": [{"noteId": "abc-123", "parts": ["voice one", "voice two"]}],
            "audioLegacy": [{"noteId": "def-456", "part": "old-shape recording"}],
            "imageItems": [{"noteId": "abc-123", "parts": ["photo bytes"]}],
        }
        page.evaluate(SEED_JS, spec)
        page.reload()
        page.wait_for_timeout(900)

        check("shows the found panel", page.locator("#lr-found").is_visible())
        counts = page.locator("#lr-counts").inner_text()
        check("counts 3 notes", "3 notes" in counts, counts.replace("\n", " | "))
        # 2 (items shape) + 1 (LEGACY blob shape) = 3. A reader that ignores the
        # legacy shape reports 2 here and looks perfectly healthy.
        check("counts 3 voice recordings (incl. the legacy-shape one)", "3 voice recordings" in counts)
        check("counts 1 photo", "1 photo" in counts)
        check("counts 2 earlier versions", "2 saved earlier versions" in counts)

        # the payload itself
        payload = page.evaluate("""async () => {
            const notes = await new Promise(r => { const q = indexedDB.open('NotedDB');
              q.onsuccess = () => { const db=q.result; const t=db.transaction('notes','readonly').objectStore('notes').getAll();
              t.onsuccess = () => { db.close(); r(t.result); }; }; });
            return notes.length;
        }""")
        check("database still readable after the tool ran", payload == 3, f"notes now: {payload}")

        with page.expect_download() as dl_info:
            page.click("#lr-download")
        dl = dl_info.value
        out = tmp / "backup.json"
        dl.save_as(str(out))
        data = json.loads(out.read_text())

        check("filename matches the old app's", dl.suggested_filename.startswith("noted-full-backup-"), dl.suggested_filename)
        check("format string is the engine's", data.get("format") == "noted.co.uk/full-backup", str(data.get("format")))
        check("version is 1", data.get("version") == 1)
        check("has exportedAt", bool(data.get("exportedAt")))
        check("all 3 notes present", len(data.get("notes", [])) == 3)
        check("note fields preserved whole", data["notes"][0].get("updatedAt") == "2026-01-02T03:04:05Z")
        check("audio keyed by note id, 2 clips on abc-123", len(data.get("audio", {}).get("abc-123", [])) == 2)
        check("LEGACY-shape recording rescued on def-456", len(data.get("audio", {}).get("def-456", [])) == 1,
              f"audio keys: {list(data.get('audio', {}).keys())}")
        check("images keyed by note id", len(data.get("images", {}).get("abc-123", [])) == 1)
        check("history keyed by note id", len(data.get("history", {}).get("abc-123", [])) == 2)
        check("media are base64 data URLs", str(data["audio"]["abc-123"][0]).startswith("data:audio/webm;base64,"),
              str(data["audio"]["abc-123"][0])[:40])
        check("note with no media absent from audio map", "ghi-789" not in data.get("audio", {}))

        # nothing destroyed
        dbs = page.evaluate(LIST_DBS)
        check("NotedDB still exists after export", "NotedDB" in dbs or dbs == ["<unsupported>"], str(dbs))
        ctx.close()

        # ---------- CASE 3: database exists but is not ours ----------
        print("\nCASE 3 — an unrelated IndexedDB database of the same name shape")
        ctx = browser.new_context()
        page = ctx.new_page()
        page.goto(url)
        page.evaluate("""async () => { await new Promise(r => {
            const q = indexedDB.open('NotedDB', 1);
            q.onupgradeneeded = e => { e.target.result.createObjectStore('somethingelse', {keyPath:'k'}); };
            q.onsuccess = () => { q.result.close(); r(); }; }); }""")
        page.reload()
        page.wait_for_timeout(700)
        check("treats a database without a notes store as empty", page.locator("#lr-empty").is_visible())
        check("does not error", not page.locator("#lr-error").is_visible())
        ctx.close()

        browser.close()

    print("\n" + ("ALL CASES PASSED" if not FAILS else f"{len(FAILS)} FAILED: {FAILS}"))
    return 1 if FAILS else 0


if __name__ == "__main__":
    sys.exit(main())
