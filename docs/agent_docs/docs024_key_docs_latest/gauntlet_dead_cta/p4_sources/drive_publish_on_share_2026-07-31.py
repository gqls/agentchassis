#!/usr/bin/env python3
"""End-to-end: play a real round, press the button, follow the address it prints.

    ~/.venvs/vonc_pw/bin/python drive_publish_on_share_2026-07-31.py

This is the only check that closes the loop the owner actually asked for — the
card links through to a record of the full debate. Everything else in this lane
proves one half: the island tests prove the endpoints, drive_round_record.py
proves the page renders a round, and drive_exchange_card proves the card draws.
None of them prove the BUTTON connects the two.

Costs three real LLM calls on the island and publishes one real round.

HOW THE CARD IS INSPECTED. Rather than OCR the PNG or guess at pixel widths, the
script replaces CanvasRenderingContext2D.prototype.fillText before the press and
records every string the card draws. That is exact: "the permalink is on the
card" becomes a substring test over the text the renderer actually emitted, and
it cannot pass because a strip of pixels happened to be wide enough.
"""
import json
import pathlib
import sys
import urllib.request

URL = "https://vonc.com/tools/gauntlet/index.html"
RECORD = "https://vonc.com/tools/gauntlet/round.html"
API = "https://tools.apis.uk/api/v1/tools/gauntlet/round/"
OUT = pathlib.Path.home() / "gauntlet_publish_card.png"

POSITION = ("A public record changes what the argument is for. If the exchange is "
            "going to be readable by anyone, the person filing a position is writing "
            "for an audience rather than thinking against an opponent.")
DEFENCE = ("Writing for an audience is not the same as writing dishonestly. The "
           "record is what makes a claim checkable at all, and an argument nobody "
           "can retrieve is not preserved by being private.")

results = []


def check(name, ok, detail=""):
    results.append((name, ok))
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + str(detail)[:160]) if detail else ""),
          flush=True)


def api_round(slug):
    req = urllib.request.Request(API + slug, headers={
        "Origin": "https://vonc.com",
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/126 Safari/537.36",
    })
    with urllib.request.urlopen(req, timeout=20) as r:
        return json.load(r)


try:
    from playwright.sync_api import sync_playwright
except ImportError as e:
    check("playwright importable", False, str(e))
    sys.exit(1)

published = {}

with sync_playwright() as p:
    browser = p.launch_args = p.chromium.launch()
    page = browser.new_page(viewport={"width": 1280, "height": 1000})
    page_errors = []
    page.on("pageerror", lambda e: page_errors.append(str(e)))

    def on_response(res):
        if res.request.method == "POST" and res.url.endswith("/publish"):
            published["status"] = res.status
            try:
                published["body"] = res.json()
            except Exception:
                published["body"] = None
    page.on("response", on_response)

    page.goto(URL, wait_until="domcontentloaded", timeout=60000)

    asset = page.evaluate(
        "fetch('/tools/assets/gauntlet-interface.js?probe=1').then(r => r.text())")
    check("the live page is running the publish-on-share code",
          'post("publish"' in asset and "data-gi-share-note" in asset)
    check("the old one-line share wiring is gone",
          "if (el.shareCard) el.shareCard.addEventListener" not in asset)

    # ── the consent surface, BEFORE anything is pressed ──────────────────────
    label = page.eval_on_selector("[data-gi-share-card]", "e => e.textContent.trim()")
    check("the button says it publishes", "publish" in label.lower(), repr(label))

    note = page.eval_on_selector("[data-gi-share-note]", "e => e.textContent.trim()")
    check("a consent note is present", len(note) > 80, "%d chars" % len(note))
    check("the note says the round becomes public",
          "public page" in note and "vonc.com" in note, repr(note[:90]))
    check("the note names what is published",
          all(w in note for w in ("provocation", "position", "challenge", "defence")))
    check("the note says nobody is named", "No name" in note)
    check("the note comes BEFORE the button in reading order",
          page.evaluate("""() => {
            const n = document.querySelector('[data-gi-share-note]');
            const b = document.querySelector('[data-gi-share-card]');
            return !!(n && b) &&
              (n.compareDocumentPosition(b) & Node.DOCUMENT_POSITION_FOLLOWING) !== 0;
          }"""))

    # ── a real round ─────────────────────────────────────────────────────────
    page.click("[data-gi-enter-btn]")
    page.wait_for_selector("[data-gi-position-input]", state="visible", timeout=90000)
    page.fill("[data-gi-position-input]", POSITION)
    page.click("[data-gi-position-submit]")
    page.wait_for_function(
        "() => { const e = document.querySelector('[data-gi-opponent-challenge]');"
        "        return e && e.textContent.trim().length > 40; }", timeout=240000)
    page.fill("[data-gi-defence-input]", DEFENCE)
    page.click("[data-gi-defence-submit]")
    page.wait_for_function(
        "() => { const e = document.querySelector('[data-gi-verdict]');"
        "        return e && e.textContent.trim().length > 3; }", timeout=240000)
    verdict = page.eval_on_selector("[data-gi-verdict]", "e => e.textContent.trim()")
    check("a real round completed with a verdict", len(verdict) > 3, verdict)

    # ── record every string the card draws ───────────────────────────────────
    page.evaluate("""() => {
      window.__drawn = [];
      const f = CanvasRenderingContext2D.prototype.fillText;
      CanvasRenderingContext2D.prototype.fillText = function (t) {
        window.__drawn.push(String(t));
        return f.apply(this, arguments);
      };
    }""")

    page.wait_for_selector("[data-gi-share-card]", state="visible", timeout=60000)
    with page.expect_download(timeout=120000) as dl:
        page.click("[data-gi-share-card]")
    dl.value.save_as(str(OUT))

    check("the publish endpoint was called", bool(published), str(published.get("status")))
    check("publish returned 200", published.get("status") == 200)
    body = published.get("body") or {}
    slug = body.get("slug", "")
    check("publish returned a slug", len(slug) == 10, repr(slug))
    check("publish returned the agreed path shape",
          body.get("path", "").startswith("/tools/gauntlet/round.html?r="), body.get("path"))

    drawn = page.evaluate("() => window.__drawn || []")
    check("the card drew something", len(drawn) > 3, "%d fillText calls" % len(drawn))
    footer = [t for t in drawn if "vonc.com" in t]
    check("the card carries an address line", bool(footer), str(footer))
    check("that address is the PERMALINK, not the tool's front door",
          any(slug in t for t in footer), str(footer))
    # Exactly ONE address line. Two means the card was drawn twice — which it
    # was, until roundIsComplete() replaced "build a whole card and throw it
    # away" as the is-there-a-round test. This assertion is what noticed.
    check("the card is drawn ONCE per press", len(footer) == 1, str(footer))

    # NOT "[data-gi-status], .gi-status": there are two .gi-status paragraphs and
    # querySelector returns the FIRST in document order, which is the sealed-entry
    # one. It is empty by the time a round is over, so the comma selector read a
    # blank string and reported the status line as broken when it was fine.
    status = page.eval_on_selector("[data-gi-status]", "e => e.textContent.trim()")
    check("the visitor is told it was published", "ublished" in status, repr(status[:110]))

    size = OUT.stat().st_size if OUT.exists() else 0
    check("a PNG was saved", size > 20000, "%d bytes" % size)
    check("no page errors during the round", not page_errors, page_errors[:2])

    # ── pressing twice must not publish twice ────────────────────────────────
    published.clear()
    with page.expect_download(timeout=120000) as dl2:
        page.click("[data-gi-share-card]")
    dl2.value.save_as(str(OUT.with_name("gauntlet_publish_card_2.png")))
    second = (published.get("body") or {}).get("slug", "")
    check("a second press returns the SAME slug (idempotent)", second == slug,
          "first %s second %s" % (slug, second))

    browser.close()

# ── the loop closes: does the record page serve THIS round? ─────────────────
if slug:
    truth = api_round(slug)
    check("the published round is the one just argued",
          truth.get("position", "").strip() == POSITION.strip()
          and truth.get("defence", "").strip() == DEFENCE.strip())

    with sync_playwright() as p:
        b = p.chromium.launch()
        pg = b.new_page(viewport={"width": 1280, "height": 1000})
        errs = []
        pg.on("pageerror", lambda e: errs.append(str(e)))
        pg.goto(RECORD + "?r=" + slug, wait_until="networkidle", timeout=60000)
        pg.wait_for_timeout(1200)
        check("the card's address renders the round", pg.is_visible("[data-gr-round]"))
        shown_pos = (pg.text_content("[data-gr-position]") or "").strip()
        shown_def = (pg.text_content("[data-gr-defence]") or "").strip()
        check("the record page shows the position that was filed", shown_pos == POSITION.strip())
        check("the record page shows the defence that was written", shown_def == DEFENCE.strip())
        check("no console errors on the record page", not errs, errs[:2])
        pg.screenshot(path=str(pathlib.Path.home() / "gauntlet_record_endtoend.png"),
                      full_page=True)
        b.close()

print()
bad = [n for n, ok in results if not ok]
if bad:
    print("FAILED %d of %d:" % (len(bad), len(results)))
    for n in bad:
        print("  -", n)
    sys.exit(1)
print("ALL %d LIVE CHECKS PASSED" % len(results))
print("published slug: %s" % slug)
print("record page   : %s?r=%s" % (RECORD, slug))
