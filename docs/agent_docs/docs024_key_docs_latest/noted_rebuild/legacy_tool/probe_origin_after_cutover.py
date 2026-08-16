"""Post-cutover origin probe (HANDOFF §6): notes written at https://noted.co.uk
must still be readable by the rescue tool AFTER the apex moved from B2 to the box.
Origin is scheme+host+port — unchanged here — so this must hold. Prove it."""
import sys
from pathlib import Path
from playwright.sync_api import sync_playwright

BASE = "https://noted.co.uk"
SEED = """
async () => {
  await new Promise((res, rej) => {
    const q = indexedDB.open('NotedDB', 4);
    q.onupgradeneeded = e => { const db = e.target.result;
      if(!db.objectStoreNames.contains('notes')) db.createObjectStore('notes',{keyPath:'id'});
      if(!db.objectStoreNames.contains('history')){const h=db.createObjectStore('history',{keyPath:'revId',autoIncrement:true});h.createIndex('noteId','noteId',{unique:false});}
      if(!db.objectStoreNames.contains('audio')) db.createObjectStore('audio',{keyPath:'noteId'});
      if(!db.objectStoreNames.contains('images')) db.createObjectStore('images',{keyPath:'noteId'});
    };
    q.onsuccess = () => { const db=q.result;
      const tx=db.transaction(['notes','audio'],'readwrite');
      tx.objectStore('notes').put({id:'cutover-1',title:'Written before cutover',content:'must survive the apex move'});
      tx.objectStore('audio').put({noteId:'cutover-1',items:[new Blob(['voice'],{type:'audio/webm'})]});
      tx.oncomplete=()=>{db.close();res();}; tx.onerror=()=>rej(tx.error);
    };
    q.onerror = () => rej(q.error);
  });
}
"""
fails = []
def check(l, ok, d=""):
    print(("  PASS  " if ok else "  FAIL  ")+l+(("  — "+d) if d else ""))
    if not ok: fails.append(l)

with sync_playwright() as p:
    b = p.chromium.launch(); ctx = b.new_context(); pg = ctx.new_page()
    # 1. the preserved legacy app, on the SAME origin
    pg.goto(BASE + "/legacy-app/", timeout=30000)
    check("legacy app served at the same origin",
          "Noted" in (pg.title() or ""), pg.title())
    # 2. write notes as the old app would
    pg.evaluate(SEED)
    check("seeded NotedDB at https://noted.co.uk", True)
    # 3. the rescue tool, same origin, different page
    pg.goto(BASE + "/tools/legacy-rescue/", timeout=30000)
    pg.wait_for_selector("#lr-found:not([hidden])", timeout=15000)
    counts = pg.locator("#lr-counts").inner_text()
    # NOTE: the legacy app creates its own note on load, so the count is >= the
    # one we seeded. Asserting "1 note" was MY error, not a product fault; the
    # meaningful assertion is that the SEEDED note is in the rescued payload.
    check("rescue tool found notes after cutover", "note" in counts, counts.replace("\n"," | "))
    check("and the recording", "1 voice recording" in counts)
    with pg.expect_download() as dl:
        pg.click("#lr-download")
    check("download offered without an account", dl.value.suggested_filename.startswith("noted-full-backup-"),
          dl.value.suggested_filename)
    import json as _j
    out = "/tmp/cutover_backup.json"; dl.value.save_as(out)
    data = _j.loads(Path(out).read_text())
    ids = [n.get("id") for n in data.get("notes", [])]
    check("the SEEDED note is in the rescued file", "cutover-1" in ids, str(ids))
    body = _j.dumps(data)
    check("its text survived verbatim", "must survive the apex move" in body)
    check("its recording came as a data URL", "data:audio/webm;base64," in body)
    check("engine-compatible format string", data.get("format") == "noted.co.uk/full-backup")
    ctx.close(); b.close()
print("\n" + ("ORIGIN PROBE PASSED — the migration premise survives cutover" if not fails else f"FAILED: {fails}"))
sys.exit(1 if fails else 0)
