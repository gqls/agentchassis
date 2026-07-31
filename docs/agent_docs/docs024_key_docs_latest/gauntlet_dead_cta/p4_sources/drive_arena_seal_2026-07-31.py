import json, pathlib, subprocess, sys
from playwright.sync_api import sync_playwright
HERE = pathlib.Path(__file__).resolve().parent
SEALED = ["personalised internet","dividing the room","quiet removal of whatever","have you seen"]
feed = json.loads(subprocess.run([sys.executable, str(HERE/"build_provocations.py")],
                                 capture_output=True, text=True, check=True).stdout)
tpl = (HERE/"arena_template_2026-07-31_seal.html").read_text()
# The stub MUST precede the template: the template's inline script fetches on parse,
# so a stub installed after it never runs and the page shows its error state (that
# cost one confusing red run).
page = ("<!doctype html><html><head><meta charset='utf-8'><script>window.__F__=" +
        json.dumps(feed) + ";window.fetch=function(){return Promise.resolve({ok:true,"
        "status:200,json:function(){return Promise.resolve(window.__F__);}});};"
        "</script></head><body>" + tpl + "</body></html>")
fails=[]
with sync_playwright() as p:
    b=p.chromium.launch(); pg=b.new_page(); errs=[]
    pg.on("pageerror", lambda e: errs.append(str(e)))
    pg.set_content(page, wait_until="load"); pg.wait_for_timeout(900)
    txt = pg.evaluate("() => document.body.innerText")
    prov = pg.eval_on_selector("[data-arena-provocation]","e=>e.textContent.trim()")
    n = pg.eval_on_selector_all("[data-arena-provocation-body]","es=>es.length")
    pbody = pg.eval_on_selector("[data-arena-provocation-body]","e=>e.textContent.trim()") if n else ""
    cards = pg.eval_on_selector_all(".lobby-card",
        "es=>es.map(e=>({t:(e.querySelector('.lobby-card-tag')||{}).textContent,"
        "ti:(e.querySelector('.lobby-card-title')||{}).textContent,h:e.getAttribute('href')}))")
    cta = pg.eval_on_selector("[data-arena-gauntlet-cta]","e=>e.getAttribute('href')+' | '+e.textContent.trim()")
    b.close()
print("provocation block :", prov)
print("seal body         :", pbody[:95])
print("gauntlet cta      :", cta)
print("cards             :", len(cards))
for c in cards: print("   [%-14s] %-46s -> %s" % ((c['t'] or '').strip(), (c['ti'] or '')[:46], c['h']))
for s in SEALED:
    if s.lower() in txt.lower(): fails.append("LEAK: %r painted" % s)
if not prov: fails.append("provocation block empty")
elif "seal" not in prov.lower(): fails.append("block does not state the seal: %r" % prov)
if not pbody: fails.append("seal body empty")
today=[c for c in cards if (c['t'] or '').strip().lower()=='today']
if len(today)!=1: fails.append("expected exactly 1 Today card, got %d" % len(today))
elif today[0]['h']!="/tools/gauntlet/index.html": fails.append("Today card href wrong: %s" % today[0]['h'])
if len(cards)!=6: fails.append("expected 6 lobby cards, got %d" % len(cards))
if errs: fails.append("page errors: %s" % errs[:2])
print()
if fails:
    print("FAILED:"); [print("   FAIL",f) for f in fails]; sys.exit(1)
print("ARENA OK — seal stated, today's question absent, 6 cards, route intact.")
