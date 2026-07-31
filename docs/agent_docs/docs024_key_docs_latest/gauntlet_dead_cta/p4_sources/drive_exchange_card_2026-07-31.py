#!/usr/bin/env python3
"""Drive one real round on the live gauntlet page and capture the card it makes.

Closes the one gap verify_card_2026-07-31.py leaves. That harness runs the
shipped renderer against stub el/state objects, so it proves the drawing logic
but ASSUMES the challenge element and the defence textarea are still populated
when the share control fires. Only a real round on the real page settles it, and
it is exactly the assumption that fails silently: the card would come out with an
empty half, or refuse to build, with every offline check still green.

Costs three real LLM calls on the island and creates one real gauntlet_rounds
row. The AI calls are slow (measured 8-18s each), hence the long waits.

Run:  ~/.venvs/vonc_pw/bin/python3 drive_exchange_card_2026-07-31.py
      (playwright is absent from the other venvs; its browsers are cached)
"""
import pathlib
import sys

from playwright.sync_api import sync_playwright

URL = "https://vonc.com/tools/gauntlet/index.html"
OUT = pathlib.Path("live_card_2026-07-31.png")

POSITION = ("Personalisation is not the villain here. The same engine that narrows one "
            "feed hands another reader a subject they would never have searched for, "
            "and shared reference points were always thinner than nostalgia makes them.")
DEFENCE = ("The mechanism cuts both ways: if optimisation can erode a common reference "
           "point it can also manufacture one, which is how a single clip reaches ten "
           "million people who follow nothing else in common.")

results = []


def check(name, ok, detail=""):
    results.append((name, ok))
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + str(detail)[:150]) if detail else ""),
          flush=True)


with sync_playwright() as p:
    browser = p.chromium.launch()
    page = browser.new_page(viewport={"width": 1280, "height": 1000})
    page_errors = []
    page.on("pageerror", lambda e: page_errors.append(str(e)))

    page.goto(URL, wait_until="domcontentloaded", timeout=60000)

    asset = page.evaluate(
        "fetch('/tools/assets/gauntlet-interface.js?probe=1').then(r => r.text())")
    check("the live page is running the NEW card code",
          "VONC ASKED" in asset and "A JUDGED VERDICT" not in asset)

    # ── 1. a real round ──────────────────────────────────────────────────────
    page.click("[data-gi-enter-btn]")
    page.wait_for_selector("[data-gi-position-input]", state="visible", timeout=90000)
    check("a real /round started (page unsealed, position box visible)", True)

    # ── 2. a real /position ──────────────────────────────────────────────────
    page.fill("[data-gi-position-input]", POSITION)
    page.click("[data-gi-position-submit]")
    page.wait_for_function(
        "() => { const e = document.querySelector('[data-gi-opponent-challenge]');"
        "        return e && e.textContent.trim().length > 40; }",
        timeout=240000)
    challenge = page.eval_on_selector("[data-gi-opponent-challenge]",
                                      "e => e.textContent.trim()")
    check("the engine returned a real challenge", len(challenge) > 40,
          "%d chars" % len(challenge))

    # ── 3. a real /defend ────────────────────────────────────────────────────
    page.fill("[data-gi-defence-input]", DEFENCE)
    page.click("[data-gi-defence-submit]")
    page.wait_for_function(
        "() => { const e = document.querySelector('[data-gi-verdict]');"
        "        return e && e.textContent.trim().length > 3; }",
        timeout=240000)
    verdict = page.eval_on_selector("[data-gi-verdict]", "e => e.textContent.trim()")
    check("the judge returned a real verdict", len(verdict) > 3, verdict)

    # ── 4. THE POINT: are the card's inputs live at press time? ──────────────
    live_defence = page.eval_on_selector("[data-gi-defence-input]", "e => e.value.trim()")
    live_challenge = page.eval_on_selector("[data-gi-opponent-challenge]",
                                           "e => e.textContent.trim()")
    check("the defence textarea STILL holds the defence after /defend returned",
          live_defence == DEFENCE, "%d chars" % len(live_defence))
    check("the challenge element STILL holds the challenge",
          len(live_challenge) > 40, "%d chars" % len(live_challenge))

    # ── 5. press the real control, capture what it produces ──────────────────
    page.wait_for_selector("[data-gi-share-card]", state="visible", timeout=60000)
    with page.expect_download(timeout=90000) as dl:
        page.click("[data-gi-share-card]")
    dl.value.save_as(str(OUT))
    size = OUT.stat().st_size if OUT.exists() else 0
    check("pressing the button produced a PNG", size > 20000, "%d bytes" % size)

    check("no page errors during the round", not page_errors, page_errors[:2])
    browser.close()

# ── 6. is that PNG actually the exchange card? ──────────────────────────────
#
# These are the only checks that look at the card itself rather than at the page
# that made it, so a missing Pillow is a FAILED RUN, not a skip. The first run of
# this script (2026-07-31) printed "SKIP PIL unavailable" and still ended in ALL
# LIVE CHECKS PASSED — which is the quiet-test trap: the summary was reporting on
# a rule that had not run. Pillow is installed in ~/.venvs/vonc_pw for this.
try:
    from PIL import Image
except ImportError:
    check("Pillow available for the image assertions", False,
          "pip install Pillow into the interpreter running this script")
    Image = None

if Image is not None:
    im = Image.open(OUT)
    check("the card is exactly 1200x630", im.size == (1200, 630), str(im.size))

    # An empty-block failure leaves a large flat region. Count sampled rows that
    # carry any non-background pixel: the exchange card marks most of the frame,
    # the old verdict-only card left the bottom half bare.
    px = im.convert("RGB").load()
    bg = px[1150, 20]
    rows = [y for y in range(0, 630, 3)
            if any(px[x, y] != bg for x in range(70, 1130, 6))]
    check("text spans the frame, not one block", len(rows) > 40,
          "%d/210 sampled rows carry marks" % len(rows))
    check("the lower half is not blank (the old card's signature)",
          any(y > 420 for y in rows), "lowest marked row %d" % (max(rows) if rows else -1))

print()
bad = [n for n, ok in results if not ok]
if bad:
    print("FAILED (%d):" % len(bad))
    for n in bad:
        print("  -", n)
    sys.exit(1)
print("ALL LIVE CHECKS PASSED — card saved to %s" % OUT)
