#!/usr/bin/env python3
"""Drive the rebuilt Arena component in a real browser, against the real feed,
BEFORE anything is delivered to production.

Two things this catches that a "does the element exist" check cannot:
  - a fabrication surviving in a code path that only runs on render
  - the offline path leaving "Fetching..." on screen forever

Serves the component from a local HTTP server (not set_content) so the relative
fetch of /data/provocations.json resolves the way it does in production.

Usage:  drive_arena.py [--feed <path-or-url>]
Exit 0 iff every check passes on BOTH viewports.
"""
import http.server
import json
import pathlib
import re
import socketserver
import sys
import threading
import urllib.request

HERE = pathlib.Path(__file__).parent
PORT = 8759
FEED_URL = "https://vonc.com/data/provocations.json"
LIVE_URL = "https://vonc.com/tools/arena/index.html"

# Mutable so run() can target either the local harness or the DEPLOYED page.
# --live drives the real page through the real CDN, which is the only way the
# published artefact gets tested rather than the source that was meant to become it.
TARGET = ["http://127.0.0.1:%d/index.html" % PORT]

DESKTOP = {"name": "desktop", "viewport": {"width": 1280, "height": 900}}
MOBILE = {"name": "mobile", "viewport": {"width": 390, "height": 844}}

# Strings that must NOT appear anywhere in the served page or its behaviour.
FABRICATIONS = [
    "FLOOR_TAKES", "REMIX_CHAINS", "REACTION_EMOJI", "REACTIONS",
    "@synthetix", "@inkblot_vera", "@loud_ratio", "@3am_take_factory",
    "@contrarian_ms", "@fracture_line", "@vinyl_argument",
    "Delusional", "Suspicious", "Cursed", "reaction-chip", "floor-take",
    "remix-card", "remix-handle", "Mutation 1", "Credit:",
    "localStorage", "seed:", "Drop It", "Re-file", "your-take",
]

# A string that IS expected — without this, a zero above is indistinguishable
# from a broken grep. (Landmine logged in NOTES 2026-07-26.)
NEGATIVE_CONTROL = "lobby-card"


def build_page(template: str) -> str:
    return (
        "<!doctype html>\n<html lang=\"en\"><head><meta charset=\"utf-8\">\n"
        "<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n"
        "<title>Arena local harness</title>\n"
        "<style>body{margin:0;background:#0b0b0d;font-family:system-ui,sans-serif;}</style>\n"
        "</head><body>\n" + template + "\n</body></html>\n"
    )


def load_feed(src: str) -> dict:
    if src.startswith("http"):
        # Cloudflare 403s a bare Python-urllib fingerprint with plain-text
        # "error code: 1010" — that is NOT the origin check. Send a browser UA.
        # (Landmine recorded in HANDOFF_2026-07-26b §5.)
        req = urllib.request.Request(src, headers={
            "User-Agent": ("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 "
                           "(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
            "Accept": "application/json,*/*",
        })
        with urllib.request.urlopen(req, timeout=30) as r:
            return json.loads(r.read().decode())
    return json.loads(pathlib.Path(src).read_text())


def serve(root: pathlib.Path):
    class Handler(http.server.SimpleHTTPRequestHandler):
        def __init__(self, *a, **kw):
            super().__init__(*a, directory=str(root), **kw)

        def log_message(self, *a):
            pass

    # allow_reuse_address is read during server_bind(), so it must be set on the
    # CLASS before construction — setting it on the instance afterwards is too
    # late and leaves the next run with EADDRINUSE for the TIME_WAIT window.
    class Server(socketserver.TCPServer):
        allow_reuse_address = True

    httpd = Server(("127.0.0.1", PORT), Handler)
    t = threading.Thread(target=httpd.serve_forever, daemon=True)
    t.start()
    return httpd


def strip_tags(s: str) -> str:
    return re.sub(r"<[^>]+>", "", s or "").strip()


def run(page, feed, profile, results):
    def check(name, ok, detail=""):
        results.append((profile["name"], name, bool(ok), detail))

    errors = []
    page.on("console", lambda m: errors.append(m.text) if m.type == "error" else None)
    page.on("pageerror", lambda e: errors.append(str(e)))

    page.goto(TARGET[0], wait_until="networkidle")
    page.wait_for_timeout(600)

    html = page.content()

    # ---- fabrication removal -------------------------------------------------
    for token in FABRICATIONS:
        check("absent: %s" % token, token not in html,
              "" if token not in html else "STILL PRESENT")

    check("negative control present (%s)" % NEGATIVE_CONTROL,
          NEGATIVE_CONTROL in html,
          "grep is live" if NEGATIVE_CONTROL in html else "HARNESS BROKEN - all absences vacuous")

    # Scoped to .tool-container, NOT body: on the deployed page the site chrome
    # legitimately contains an email address (vonc@contactforsales.com) and a
    # mobile-nav <button>. Asserting over the whole page would fail a correct
    # component for things it does not own. The claim under test is "this
    # component invents no users and changes no state", so scope it to this
    # component. Verified 2026-07-27: 0 of each inside the container.
    comp_text = page.inner_text(".tool-container")
    handles = re.findall(r"@[a-z0-9_]{3,}", comp_text, re.I)
    check("no @handles in the component", not handles, ", ".join(handles[:5]))

    # ---- the provocation is the REAL one ------------------------------------
    today = feed["today"]
    prov = page.inner_text("[data-arena-provocation]").strip()
    check("provocation matches feed today.headline",
          prov == strip_tags(today["headline"]),
          "page=%r feed=%r" % (prov[:60], strip_tags(today["headline"])[:60]))

    body = page.inner_text("[data-arena-provocation-body]").strip()
    check("body matches feed today.body", body == today["body"].strip(),
          "page=%r" % body[:60])

    # text_content(), NOT inner_text(): .provocation-day and .btn-primary carry
    # `text-transform: uppercase`, and inner_text() returns the RENDERED text, so
    # it reports "TODAY'S PROVOCATION" for a DOM value of "Today's Provocation".
    # Asserting on rendered text here would fail a correct page.
    eyebrow = page.text_content("[data-arena-eyebrow]").strip()
    check("eyebrow matches feed", eyebrow == today["eyebrow"], "page=%r" % eyebrow)

    # the date shown is the FEED's date, not "now"
    date_txt = page.inner_text("[data-arena-date]").strip()
    gen = feed["generated_at"][:10]  # YYYY-MM-DD
    y, m, d = gen.split("-")
    months = ["January", "February", "March", "April", "May", "June", "July",
              "August", "September", "October", "November", "December"]
    expect_date = "%d %s %s" % (int(d), months[int(m) - 1], y)
    check("date is the feed's date, not today's", date_txt == expect_date,
          "page=%r expected=%r" % (date_txt, expect_date))

    # ---- the CTA is a real route --------------------------------------------
    cta = page.query_selector("[data-arena-gauntlet-cta]")
    check("gauntlet CTA exists", cta is not None)
    if cta:
        href = cta.get_attribute("href")
        label = cta.text_content().strip()  # see the text-transform note above
        check("CTA href from feed", href == today["primary_cta"]["url"],
              "href=%r" % href)
        check("CTA label from feed", label == today["primary_cta"]["label"],
              "label=%r" % label)

    check("no textarea (take box removed)",
          page.query_selector(".tool-container textarea") is None)
    # Scoped for the same reason as the @handle check above — the site's
    # mobile-menu toggle is a real control and not this component's.
    comp_buttons = page.query_selector_all(".tool-container button")
    check("no buttons in the component (only links remain)",
          len(comp_buttons) == 0, "%d found" % len(comp_buttons))

    # ---- the lobby ----------------------------------------------------------
    cards = page.query_selector_all(".lobby-card")
    expected_cards = len([c for c in feed["arena"]["cards"] if c.get("url") and c.get("title")])
    check("lobby renders every usable feed card",
          len(cards) == expected_cards,
          "rendered=%d expected=%d" % (len(cards), expected_cards))

    dead = []
    for c in cards:
        h = c.get_attribute("href")
        if not h or h == "#":
            dead.append(h)
    check("no dead card anchors", not dead, str(dead))

    if cards:
        check("first card routes to the Gauntlet",
              cards[0].get_attribute("href") == feed["arena"]["cards"][0]["url"],
              cards[0].get_attribute("href"))

    arch = page.query_selector("[data-arena-archive-cta]")
    if arch:
        check("archive CTA href from feed",
              arch.get_attribute("href") == feed["arena"]["cta"]["url"],
              arch.get_attribute("href"))

    # every anchor on the page has a real destination
    bad = [a.get_attribute("href") for a in page.query_selector_all("a")
           if not a.get_attribute("href") or a.get_attribute("href") == "#"]
    check("no dead anchors anywhere", not bad, str(bad))

    # ---- layout -------------------------------------------------------------
    overflow = page.evaluate(
        "() => document.documentElement.scrollWidth - document.documentElement.clientWidth")
    check("no horizontal overflow", overflow <= 0, "overflow=%spx" % overflow)

    check("no console errors", not errors, " | ".join(errors[:3]))


def run_offline(page, results):
    """The feed 503s: the page must say so, not sit on 'Fetching...'."""
    def check(name, ok, detail=""):
        results.append(("offline", name, bool(ok), detail))

    page.route("**/data/provocations.json",
               lambda route: route.fulfill(status=503, body="upstream down"))
    page.goto(TARGET[0], wait_until="networkidle")
    page.wait_for_timeout(800)

    prov = page.inner_text("[data-arena-provocation]").strip()
    check("does not sit on 'Fetching'", "Fetching" not in prov, "shows %r" % prov[:60])
    check("says it could not load", "could not be loaded" in prov, "shows %r" % prov[:60])

    cta = page.query_selector("[data-arena-gauntlet-cta]")
    check("gauntlet link still works offline",
          cta is not None and cta.get_attribute("href") == "/tools/gauntlet/index.html",
          cta.get_attribute("href") if cta else "MISSING")

    empty = page.query_selector(".empty-state")
    check("lobby shows an honest empty state", empty is not None,
          empty.inner_text().strip() if empty else "MISSING")


def main():
    feed_src = FEED_URL
    if "--feed" in sys.argv:
        feed_src = sys.argv[sys.argv.index("--feed") + 1]
    live = "--live" in sys.argv

    feed = load_feed(feed_src)

    httpd = None
    if live:
        # cache-bust: the CDN will happily serve the pre-delivery page otherwise,
        # and a stale 200 looks exactly like a successful deploy.
        TARGET[0] = LIVE_URL + "?cb=%d" % len(json.dumps(feed))
        print("driving DEPLOYED page: %s" % TARGET[0])
    else:
        template = (HERE / "arena_new.html").read_text()
        root = HERE / "_harness"
        (root / "data").mkdir(parents=True, exist_ok=True)
        (root / "index.html").write_text(build_page(template))
        (root / "data" / "provocations.json").write_text(json.dumps(feed))
        httpd = serve(root)
    results = []
    try:
        from playwright.sync_api import sync_playwright
        with sync_playwright() as p:
            browser = p.chromium.launch()
            for profile in (DESKTOP, MOBILE):
                ctx = browser.new_context(viewport=profile["viewport"])
                pg = ctx.new_page()
                run(pg, feed, profile, results)
                ctx.close()
            ctx = browser.new_context(viewport=DESKTOP["viewport"])
            pg = ctx.new_page()
            run_offline(pg, results)
            ctx.close()
            browser.close()
    finally:
        if httpd is not None:
            httpd.shutdown()
            httpd.server_close()  # shutdown() stops serving; it does NOT free the port

    passed = [r for r in results if r[2]]
    failed = [r for r in results if not r[2]]
    for prof, name, ok, detail in results:
        if not ok:
            print("FAIL  [%-8s] %-46s %s" % (prof, name, detail))
    print("\n%d passed, %d failed (of %d)" % (len(passed), len(failed), len(results)))
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
