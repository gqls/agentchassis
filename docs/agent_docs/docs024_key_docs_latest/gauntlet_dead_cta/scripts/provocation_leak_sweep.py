#!/usr/bin/env python3
# provocation_leak_sweep.py — does ANY page paint today's sealed provocation?
#
# WHY THIS EXISTS, AND WHY IT SWEEPS EVERY PAGE
# HANDOFF_2026-07-30_C measured three hand-picked pages and concluded the leak was
# the home page. It is not: /tools/arena/index.html leaks too, and that page's whole
# purpose is choosing what to argue. A three-page sample was the wrong denominator,
# so this takes its page list from pages.url and tests all of them.
#
# WHY IT RENDERS RATHER THAN GREPS — the text is in NEITHER content_data (pure site
# chrome) NOR rendered_html (an empty shell under data-runtime-fill="true"). It is
# written by snippets.js after load. A curl grep reports "absent" on every page
# INCLUDING the ones that show it, so an HTML-level check here is not a weak check,
# it is one that cannot see the defect at all.
#
# USAGE
#   ~/.venvs/vonc_pw/bin/python provocation_leak_sweep.py            # urls from the DB
#   ~/.venvs/vonc_pw/bin/python provocation_leak_sweep.py urls.txt   # one url path per line
# The lane's venv ~/.venvs/vonc_pw already has playwright + Pillow — do NOT build a
# new one (I did, having not read this lane's own NOTES first).
#
# LANDMINE: read pages.url, never construct one. /about/index.html 404s with a B2
# NoSuchKey body that reads as ~286 chars of page content; the real path is
# /about.html. That is why the status code is printed and a non-200 is UNSCORED
# rather than counted clean.
#
# LANDMINE: an <em> in the headline splits the text node. Match on a container's
# innerText, never XPath text() — that mistake reported "NOT IN DOM" on the very
# page that was painting it.

import json
import subprocess
import sys
import urllib.request

SITE = "https://vonc.com"
DATA = SITE + "/data/provocations.json"
SITE_ID = "9ec3b9ee-5b08-461b-b4f8-9e1e03579c74"


def load_today():
    # LANDMINE: vonc.com is behind Cloudflare and 403s urllib's default
    # User-Agent, while curl on the same URL returns 200. A bare urlopen here
    # fails with HTTP 403 that reads like the file is missing.
    req = urllib.request.Request(DATA, headers={"User-Agent": "Mozilla/5.0"})
    with urllib.request.urlopen(req, timeout=30) as r:
        d = json.load(r)
    t = d["today"]
    return {
        "headline": t["headline"].replace("<em>", "").replace("</em>", ""),
        "body": t["body"][:50],
        "lobby_title": d["arena"]["cards"][0]["title"],
        "lobby_desc": d["arena"]["cards"][0]["desc"][:45],
    }


def page_urls(argv):
    if len(argv) > 1:
        paths = [l.strip() for l in open(argv[1]) if l.strip()]
    else:
        sql = ("SELECT url FROM pages WHERE site_id='%s' AND status='active' "
               "ORDER BY url;" % SITE_ID)
        out = subprocess.run(
            ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
             "--", "psql", "-U", "clients_user", "-d", "clients_db", "-t", "-A",
             "-c", sql],
            capture_output=True, text=True, check=True).stdout
        paths = [l.strip() for l in out.splitlines() if l.strip()]
    # "/" is served separately from "/index.html" and both must be tested: "/" is
    # the normal arrival path.
    return [SITE + "/"] + [SITE + p for p in paths]


def main():
    from playwright.sync_api import sync_playwright
    today = load_today()
    urls = page_urls(sys.argv)
    leaks, unscored = [], []
    print(f"{'url':47s} {'code':5s} {'head':6s} {'body':6s} {'lobT':6s} {'lobD':6s} chars")
    with sync_playwright() as p:
        b = p.chromium.launch()
        for u in urls:
            pg = b.new_context().new_page()
            try:
                resp = pg.goto(u, wait_until="load", timeout=45000)
                code = resp.status if resp else 0
                pg.wait_for_timeout(3500)      # the client-side fetch must settle
                txt = pg.evaluate("() => document.body.innerText")
            except Exception as e:
                print(f"{u[:47]:47s} ERR   {str(e)[:40]}")
                unscored.append(f"{u} (error)")
                pg.close()
                continue
            hit = {k: (v in txt) for k, v in today.items()}
            if code != 200:
                unscored.append(f"{u} (HTTP {code})")
            elif any(hit.values()):
                leaks.append(u)
            mark = "  <<< LEAK" if code == 200 and any(hit.values()) else ""
            print(f"{u[:47]:47s} {code:<5} {str(hit['headline']):6s} "
                  f"{str(hit['body']):6s} {str(hit['lobby_title']):6s} "
                  f"{str(hit['lobby_desc']):6s} {len(txt)}{mark}")
        b.close()

    print(f"\n{len(leaks)} of {len(urls)} pages paint today's provocation:")
    for u in leaks:
        print("   LEAK    ", u)
    if unscored:
        # A capped or errored sweep reporting clean is a false green.
        print(f"\n{len(unscored)} page(s) NOT SCORED — this run is not a clean bill:")
        for u in unscored:
            print("   unscored", u)
    # Three distinct exits, because "it found leaks" and "it fell over" must never
    # share one. An earlier draft of this script exited 1 for both, so a crash in
    # load_today() was indistinguishable from a confirmed leak — the same false-signal
    # shape as the PIL skip that printed ALL LIVE CHECKS PASSED (see NOTES 2026-07-30).
    #   0 = every page scored, none leaks
    #   1 = at least one page paints today's provocation
    #   2 = the sweep is incomplete; it is neither a pass nor a fail
    if unscored:
        sys.exit(2)
    sys.exit(1 if leaks else 0)


if __name__ == "__main__":
    main()
