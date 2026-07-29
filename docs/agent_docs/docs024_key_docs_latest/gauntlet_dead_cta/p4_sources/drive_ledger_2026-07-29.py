#!/usr/bin/env python3
"""Drive the opinion-ledger build through two full live rounds.

Proves the rail, not just the render: no entry exists before /defend returns;
exactly one exists after; a reload's restore path writes NOTHING; a new tab
(the true returning visitor: fresh sessionStorage, shared localStorage) sees
the sealed door with the diary below it; a second round appends and the list
is newest-first; erasing takes two presses; the shipped overflow probe stays
clean with entries on a mobile viewport.
"""
import threading, functools, http.server, socketserver, pathlib, json
import urllib.request, urllib.error
from playwright.sync_api import sync_playwright

HERE = pathlib.Path(__file__).parent
SCRATCH = pathlib.Path("/tmp/claude-1000/-home-ant-projects-agentchassis/8a5e2611-422b-4596-9b52-4c3e3251ad63/scratchpad")
PORT = 8771
ORIGIN = "https://vonc.com"
LEDGER_KEY = "vonc_gauntlet_ledger_v1"

rendered = (SCRATCH / "ledger_rendered_new.html").read_text()
(SCRATCH / "test_ledger.html").write_text(
    "<!doctype html><html><head><meta charset='utf-8'>"
    "<meta name='viewport' content='width=device-width, initial-scale=1'></head>"
    "<body style='margin:0'>" + rendered +
    "<script src='ledger_js_new.js'></script></body></html>"
)

socketserver.TCPServer.allow_reuse_address = True
handler = functools.partial(http.server.SimpleHTTPRequestHandler, directory=str(SCRATCH))
handler.log_message = lambda *a, **k: None
httpd = socketserver.TCPServer(("127.0.0.1", PORT), handler)
threading.Thread(target=httpd.serve_forever, daemon=True).start()

SHIPPED = (SCRATCH / "shipped_check.js").read_text()
results, errors = [], []


def check(name, ok, detail=""):
    results.append(ok)
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + str(detail)[:160]) if detail else ""), flush=True)


def proxy(route):
    req = route.request
    if req.method == "OPTIONS":
        route.fulfill(status=204, headers={
            "access-control-allow-origin": "*",
            "access-control-allow-methods": "POST, OPTIONS",
            "access-control-allow-headers": "content-type"})
        return
    body = (req.post_data or "").encode()
    r = urllib.request.Request(req.url, data=body, method=req.method, headers={
        "Content-Type": "application/json", "Origin": ORIGIN,
        "User-Agent": req.headers.get("user-agent", "Mozilla/5.0"),
        "Accept": "application/json"})
    try:
        with urllib.request.urlopen(r, timeout=120) as resp:
            status, payload = resp.status, resp.read()
    except urllib.error.HTTPError as e:
        status, payload = e.code, e.read()
    route.fulfill(status=status, body=payload, headers={
        "content-type": "application/json", "access-control-allow-origin": "*"})


def ledger_store(page):
    raw = page.evaluate("() => localStorage.getItem('%s')" % LEDGER_KEY)
    return json.loads(raw) if raw else None


def ledger_visible(page):
    return page.evaluate("() => {"
                         " const l = document.querySelector('[data-gi-ledger]');"
                         " return l ? l.offsetParent !== null : null; }")


def play_round(page, position, defence):
    page.click("[data-gi-enter-btn]")
    page.wait_for_function(
        "() => !document.querySelector('section').classList.contains('gi-sealed')", timeout=30000)
    page.fill("[data-gi-position-input]", position)
    page.click("[data-gi-position-submit]")
    page.wait_for_function(
        "() => document.querySelector('[data-gi-opponent-position]').textContent.trim().length > 20",
        timeout=90000)
    mid_store = ledger_store(page)
    page.fill("[data-gi-defence-input]", defence)
    page.click("[data-gi-defence-submit]")
    page.wait_for_function(
        "() => document.querySelector('[data-gi-verdict]').textContent.trim().length > 5",
        timeout=90000)
    return mid_store


POSITION_1 = ("I agree: an opinion you cannot date is an opinion you cannot be "
              "held to, and a record of where you stood, when, is the only "
              "honest measure of whether you ever change your mind.")
DEFENCE_1 = ("The challenge assumes memory is enough; but memory rewrites "
             "itself to flatter the present, which is exactly why a dated, "
             "unedited record of what you argued is worth keeping.")
POSITION_2 = ("I disagree: novelty is doing the work here — most daily rituals "
              "survive on habit, not value, and arguing with a judge is only "
              "worth doing while the verdicts still surprise you.")
DEFENCE_2 = ("A surprise that stops surprising you has taught you something; "
             "the ritual earns its keep precisely by making its own novelty "
             "wear off, which habit alone never does.")

with sync_playwright() as pw:
    b = pw.chromium.launch()
    ctx = b.new_context(viewport={"width": 390, "height": 844})
    page = ctx.new_page()
    page.on("pageerror", lambda e: errors.append("pageerror: " + str(e)))
    page.on("console", lambda m: errors.append(m.text) if m.type == "error" else None)
    page.route("**tools.apis.uk/**", proxy)
    page.goto(f"http://127.0.0.1:{PORT}/test_ledger.html", wait_until="networkidle")
    page.wait_for_timeout(500)

    # ── fresh visitor ──
    check("fresh: ledger block hidden", ledger_visible(page) is False)
    check("fresh: no ledger key in localStorage", ledger_store(page) is None)
    over = page.evaluate(SHIPPED, "")
    check("fresh: no overflow, sealed with hidden ledger",
          not (over.get("over", 0) > 2 or over.get("clipped")), over)

    # ── round 1 ──
    mid = play_round(page, POSITION_1, DEFENCE_1)
    check("rail: NO entry exists after /position, before /defend", mid is None)

    store = ledger_store(page)
    check("after verdict: exactly one entry", isinstance(store, list) and len(store) == 1,
          f"{len(store) if store else store}")
    e0 = store[0] if store else {}
    dom_prov = page.evaluate("() => document.querySelector('[data-gi-challenge-title]').textContent.trim()")
    dom_verdict = page.evaluate("() => document.querySelector('[data-gi-verdict]').textContent.trim()")
    check("entry.provocation is the round's real question", e0.get("provocation", "") == dom_prov,
          e0.get("provocation", "")[:60])
    check("entry.position is the text actually FILED", e0.get("position") == POSITION_1)
    check("entry.verdict is the judge's real line", e0.get("verdict", "") == dom_verdict,
          e0.get("verdict", "")[:60])
    check("entry has a roundId and a parseable date",
          bool(e0.get("roundId")) and "T" in str(e0.get("date", "")), e0.get("date", ""))
    check("after verdict: ledger visible", ledger_visible(page) is True)
    shown_date = page.evaluate("() => document.querySelector('.gi-ledger-date').textContent")
    check("rendered date is en-GB long form", "2026" in shown_date and any(c.isalpha() for c in shown_date), shown_date)
    count_txt = page.evaluate("() => document.querySelector('[data-gi-ledger-count]').textContent")
    check("count line says one round", count_txt == "1 round on record", count_txt)

    # ── reload: restore must write nothing ──
    page.reload(wait_until="networkidle")
    page.wait_for_timeout(700)
    store = ledger_store(page)
    check("rail: reload+restore writes NOTHING (still one entry)",
          isinstance(store, list) and len(store) == 1, f"{len(store) if store else store}")
    check("reload: ledger visible with the resumed closed round", ledger_visible(page) is True)

    # ── returning visitor: new tab = fresh sessionStorage, shared localStorage ──
    page2 = ctx.new_page()
    page2.on("pageerror", lambda e: errors.append("pageerror2: " + str(e)))
    page2.on("console", lambda m: errors.append(m.text) if m.type == "error" else None)
    page2.route("**tools.apis.uk/**", proxy)
    page2.goto(f"http://127.0.0.1:{PORT}/test_ledger.html", wait_until="networkidle")
    page2.wait_for_timeout(500)
    sealed2 = page2.evaluate("() => document.querySelector('section').classList.contains('gi-sealed')")
    check("returning visitor: page opens SEALED", sealed2 is True)
    check("returning visitor: diary visible below the sealed door", ledger_visible(page2) is True)
    over = page2.evaluate(SHIPPED, "")
    check("returning visitor: no overflow, sealed + populated ledger",
          not (over.get("over", 0) > 2 or over.get("clipped")), over)

    # ── round 2 on the new tab ──
    play_round(page2, POSITION_2, DEFENCE_2)
    store = ledger_store(page2)
    check("after round 2: two entries", isinstance(store, list) and len(store) == 2,
          f"{len(store) if store else store}")
    first_shown_pos = page2.evaluate(
        "() => document.querySelector('.gi-ledger-entry .gi-ledger-position').textContent")
    check("list is newest-first (round 2 at the top)", POSITION_2[:40] in first_shown_pos,
          first_shown_pos[:60])
    count_txt = page2.evaluate("() => document.querySelector('[data-gi-ledger-count]').textContent")
    check("count line says two rounds", count_txt == "2 rounds on record", count_txt)
    over = page2.evaluate(SHIPPED, "")
    check("no overflow after verdict + two-entry ledger",
          not (over.get("over", 0) > 2 or over.get("clipped")), over)

    # ── two-press erase ──
    page2.click("[data-gi-ledger-clear]")
    page2.wait_for_timeout(200)
    armed_txt = page2.evaluate("() => document.querySelector('[data-gi-ledger-clear]').textContent")
    store = ledger_store(page2)
    check("erase press 1: arms only, store intact",
          armed_txt == "Press again to erase for good" and isinstance(store, list) and len(store) == 2,
          armed_txt)
    page2.click("[data-gi-ledger-clear]")
    page2.wait_for_timeout(200)
    store = ledger_store(page2)
    check("erase press 2: store gone, block hidden",
          store is None and ledger_visible(page2) is False)

    # ── fresh context = genuinely new visitor ──
    ctx3 = b.new_context(viewport={"width": 390, "height": 844})
    page3 = ctx3.new_page()
    page3.route("**tools.apis.uk/**", proxy)
    page3.goto(f"http://127.0.0.1:{PORT}/test_ledger.html", wait_until="networkidle")
    check("new browser context: no diary, block hidden",
          ledger_store(page3) is None and ledger_visible(page3) is False)

    real_errors = [e for e in errors if "404" not in e]
    check("no unexpected page errors", not real_errors, real_errors[:3])
    b.close()

print(("ALL %d PASS" % len(results)) if all(results) else "%d FAILURES" % results.count(False))
