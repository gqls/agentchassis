import json, sys
from playwright.sync_api import sync_playwright
P = json.load(open(sys.argv[1]))
head = P["today"]["headline"].replace("<em>","").replace("</em>","")
body = P["today"]["body"]
pages = {"home":"https://vonc.com/","gauntlet":"https://vonc.com/tools/gauntlet/index.html",
         "provocations_index":"https://vonc.com/provocations/index.html"}
with sync_playwright() as p:
    b = p.chromium.launch()
    for n,u in pages.items():
        pg = b.new_context().new_page()
        pg.goto(u, wait_until="load", timeout=45000); pg.wait_for_timeout(3500)
        txt = pg.evaluate("() => document.body.innerText")
        print(f"{n:20s} headline_painted={head in txt!s:5s}  body_painted={body[:60] in txt!s:5s}  chars={len(txt)}")
    b.close()
