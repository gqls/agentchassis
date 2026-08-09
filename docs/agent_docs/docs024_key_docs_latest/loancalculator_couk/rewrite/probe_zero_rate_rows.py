#!/usr/bin/env python3
"""Drive loancalculator.co.uk's REWRITTEN tool rows at 0%, before they ship.

The rows are written but no page has been reassembled yet, so the live site
still serves the old bytes. This drives exactly what WILL ship, from the DB,
so a mistake costs an edit rather than a live wrong number on a credit site.

Expectations are derived here from the definitions, not read off the tool.
"""
import json, socket, subprocess, sys, time
from playwright.sync_api import sync_playwright

SITE = "0162cde4-633e-45e9-8ca6-87a6b2fe1d26"
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db"]
OUT = "/home/ant/.claude-scratch/claude-1000/-home-ant-projects-agentchassis/4bbacd62-7bab-4eef-bcb2-e578638667f9/scratchpad/lc_rows"

ROWS = {   # url -> slot
    "/index.html": "tool-3",
    "/tools/compare-loans.html": "tool-2",
    "/tools/interest-rate-stress-test.html": "tool-2",
    "/tools/overpayment-calculator.html": "tool-3",
    "/tools/settlement-calculator.html": "tool-2",
    "/tools/consolidation.html": "tool-2",
}

import os
os.makedirs(OUT, exist_ok=True)
for url, slot in ROWS.items():
    sql = ("SELECT pc.rendered_html FROM page_components pc JOIN pages p ON pc.page_id=p.id "
           "WHERE p.site_id='%s' AND p.url='%s' AND pc.slot_name='%s';" % (SITE, url, slot))
    r = subprocess.run(PSQL + ["-tA", "-c", sql], capture_output=True, text=True)
    if r.returncode != 0 or not r.stdout.strip():
        sys.exit("could not fetch %s %s: %s" % (url, slot, r.stderr[:200]))
    name = url.strip("/").replace("/", "_").replace(".html", "") + ".html"
    open(os.path.join(OUT, name), "w", encoding="utf-8").write(
        '<!DOCTYPE html><html><head><meta charset="UTF-8"></head><body>\n'
        + r.stdout.rstrip("\n") + "\n</body></html>")

fails = []
def ck(n, got, want):
    ok = got == want
    print(("PASS " if ok else "FAIL ") + f"{n}: got {got!r} want {want!r}")
    if not ok: fails.append(n)

s = socket.socket(); s.bind(("127.0.0.1", 0)); port = s.getsockname()[1]; s.close()
srv = subprocess.Popen([sys.executable, "-m", "http.server", str(port), "--bind",
                        "127.0.0.1", "--directory", OUT],
                       stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
try:
    for _ in range(50):
        try: socket.create_connection(("127.0.0.1", port), timeout=0.2).close(); break
        except OSError: time.sleep(0.1)
    base = f"http://127.0.0.1:{port}/"
    with sync_playwright() as pw:
        b = pw.chromium.launch(); pg = b.new_page(); errs = []
        pg.on("pageerror", lambda e: errs.append(str(e)))

        def load(n):
            pg.goto(base + n); pg.wait_for_load_state("load"); pg.wait_for_timeout(200)
        def txt(sel): return pg.locator(sel).first.inner_text()
        def nan_in_outputs():
            vals = pg.eval_on_selector_all(
                "[id]", "els=>els.map(e=>e.textContent||'')")
            return any("NaN" in v for v in vals)

        # 1. index widget — 10000 at 0% over 5y -> 10000/60
        load("index.html")
        pg.fill("#amount", "10000"); pg.fill("#interest", "0"); pg.fill("#years", "5")
        pg.wait_for_timeout(150)
        ck("index 0% monthly", txt("#monthly-display"), "£166.67")
        ck("index 0% interest", txt("#total-interest"), "£0.00")
        ck("index 0% total", txt("#total-cost"), "£10,000.00")

        # 2. compare-loans — A 5000 at 0% over 3y -> 5000/36, no interest, A wins
        load("tools_compare-loans.html")
        for sel, v in (("#amt-a","5000"),("#apr-a","0"),("#term-a","3"),
                       ("#amt-b","5000"),("#apr-b","10"),("#term-b","3")):
            pg.fill(sel, v)
        pg.wait_for_timeout(150)
        ck("compare 0%A monthly", txt("#res-m-a"), "£138.89")
        ck("compare 0%A interest", txt("#res-i-a"), "£0.00")
        ck("compare verdict not empty", txt("#verdict") != "", True)
        # Search the DISPLAYED text, never pg.content(): the fix comments in the
        # source explain the NaN defect, so a content grep matches its own
        # explanation and reports the bug it just fixed.
        ck("compare no NaN shown", nan_in_outputs(), False)

        # 3. stress test — 10000 at 0% over 3y; STRESS delta applies to the new rate
        load("tools_interest-rate-stress-test.html")
        pg.fill("#stress-bal","10000"); pg.fill("#stress-apr","0"); pg.fill("#stress-term","3")
        pg.wait_for_timeout(150)
        ck("stress 0% current", txt("#curr-pay"), "£277.78")
        ck("stress no NaN shown", nan_in_outputs(), False)

        # 4. overpayment — 0% means nothing to save; 15000/60 = 250, +50 -> 50 months
        load("tools_overpayment-calculator.html")
        pg.fill("#bal","15000"); pg.fill("#rate","0"); pg.fill("#term","5"); pg.fill("#over","50")
        pg.wait_for_timeout(150)
        ck("overpay 0% saved", txt("#save-display"), "£0.00")
        ck("overpay 0% months saved", txt("#time-display"), "10")

        # 5. settlement — 0% APR: no 58-day interest, settle the balance
        load("tools_settlement-calculator.html")
        pg.fill("#settle-bal","5000"); pg.fill("#settle-apr","0")
        pg.wait_for_timeout(150)
        ck("settle 0% total", txt("#settle-result"), "£5,000.00")

        # 6. consolidation — 5000 of debt, new loan 0% over 5y -> 83.33/mo, no interest
        load("tools_consolidation.html")
        pg.fill(".d-bal", "5000"); pg.fill(".d-rate", "10"); pg.fill(".d-months", "24")
        pg.fill("#new-rate", "0"); pg.fill("#new-term", "5")
        pg.wait_for_timeout(200)
        ck("consol 0% new monthly", txt("#new-monthly"), "£83.33")
        ck("consol 0% new interest", txt("#new-int"), "£0.00")
        ck("consol verdict non-empty", txt("#verdict") != "", True)
        # blank rate must NOT be priced as interest-free
        pg.fill("#new-rate", ""); pg.wait_for_timeout(200)
        ck("consol blank rate withholds verdict", txt("#verdict"), "")
        ck("consol no NaN shown", nan_in_outputs(), False)

        ck("no JS errors anywhere", errs, [])
        b.close()
finally:
    srv.terminate()

print("\n%d FAIL(s)" % len(fails))
sys.exit(1 if fails else 0)
