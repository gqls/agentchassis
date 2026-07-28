import pathlib, sys
from playwright.sync_api import sync_playwright
JS = pathlib.Path(sys.argv[1]).read_text()
URL = "https://vonc.com/tools/gauntlet/index.html"
errs=[]
with sync_playwright() as p:
    b=p.chromium.launch(); ctx=b.new_context(viewport={"width":1280,"height":900}); pg=ctx.new_page()
    pg.on("pageerror", lambda e: errs.append(str(e)))
    pg.on("console", lambda m: errs.append(m.text) if m.type=="error" else None)
    pg.route("**/tools/assets/gauntlet-interface.js",
             lambda r: r.fulfill(status=200, content_type="application/javascript", body=JS))

    pg.goto(URL, wait_until="domcontentloaded", timeout=90000)
    early = pg.evaluate("() => (document.querySelector('[data-gi-challenge-title]')||{}).textContent||''")
    print(f"  1. at load, provocation area : {early.strip()[:52]!r}")
    pg.wait_for_timeout(3500)
    late = pg.evaluate("() => (document.querySelector('[data-gi-challenge-title]')||{}).textContent||''")
    print(f"     3.5s later                : {late.strip()[:52]!r}")

    # the silent-refusal case: press Send Defence with no round
    pg.fill("[data-gi-defence-input]", "a defence typed before any round exists")
    pg.click("[data-gi-defence-submit]"); pg.wait_for_timeout(600)
    note = pg.evaluate("() => (document.querySelector('.gi-defence-note')||{}).textContent||''")
    print(f"\n  2. Send Defence with no round : note={note.strip()[:66]!r}")

    # real round -> challenge
    pg.fill("[data-gi-position-input]", "Personalisation removes common ground without consent.")
    pg.click("[data-gi-position-submit]")
    pg.wait_for_function("() => {const c=document.querySelector('[data-gi-opponent-challenge]');return c&&c.textContent.trim().length>0;}", timeout=90000)
    before = pg.evaluate("""() => ({
      ch:(document.querySelector('[data-gi-opponent-challenge]').textContent||'').trim(),
      draft:(document.querySelector('[data-gi-defence-input]').value||''),
      clock:(document.querySelector('[data-gi-timer]')||{}).textContent||''})""")
    print(f"\n  3. challenge received         : {len(before['ch'])} chars, clock={before['clock'].strip()}")

    pg.fill("[data-gi-defence-input]", "Visibility is not agreement, and that gap is the whole point.")
    pg.wait_for_timeout(400)

    # THE BUG: reload
    pg.reload(wait_until="domcontentloaded", timeout=90000); pg.wait_for_timeout(2500)
    after = pg.evaluate("""() => ({
      ch:(document.querySelector('[data-gi-opponent-challenge]').textContent||'').trim(),
      hidden:document.querySelector('[data-gi-opponent-block]').classList.contains('is-empty'),
      draft:(document.querySelector('[data-gi-defence-input]').value||''),
      clock:(document.querySelector('[data-gi-timer]')||{}).textContent||''})""")
    print(f"\n  4. AFTER RELOAD")
    print(f"     challenge survived        : {after['ch']==before['ch']}  ({len(after['ch'])} chars)")
    print(f"     block visible             : {not after['hidden']}")
    print(f"     typed defence survived    : {after['draft']!r}")
    print(f"     clock still running       : {after['clock'].strip()}")

    # and can the defence actually be sent after the reload?
    pg.click("[data-gi-defence-submit]")
    ok = pg.wait_for_function("() => {const v=document.querySelector('[data-gi-verdict]');return v&&v.textContent.trim().length>0;}", timeout=90000)
    verdict = pg.evaluate("() => document.querySelector('[data-gi-verdict]').textContent.trim()")
    print(f"\n  5. defence sent after reload  : verdict={verdict[:44]!r}")
    ctx.close(); b.close()
print(f"\n  console/page errors: {errs if errs else 'none'}")
