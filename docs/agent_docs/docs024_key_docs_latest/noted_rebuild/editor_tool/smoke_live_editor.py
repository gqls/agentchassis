#!/usr/bin/env python3
"""
LIVE smoke of the Noted editor at https://app.noted.co.uk/tools/write/.

What it proves that the local suite cannot: the real page, served by the real
nginx, calling the real engine over the open internet — register, save, reopen
from a SECOND independent browser session (the product's whole promise), plus
the induced-outage degraded path against the live service.

The outage is induced IN THE BROWSER (Playwright route.abort on POST /api/notes)
— the server is never touched, broken, or even aware. This is the half the
platform runner cannot do (it cannot induce a failing dependency); the local
suite proves the contract exhaustively against a stub, this proves the wiring.

Accounts: one throwaway per run (noted-smoke-<epoch>@example.invalid). The engine
has no account-deletion endpoint, so the account remains; the address is
unroutable by design (.invalid) and holds one test note.

Run:  /home/ant/.venvs/vonc_pw/bin/python smoke_live_editor.py [base_url]
Default base_url: https://app.noted.co.uk    (pass the apex at cutover — §6:
origin is scheme+host+port, so this MUST be re-run on the new origin.)
Exit 0 = all checks passed.
"""
import sys
import time

from playwright.sync_api import sync_playwright

BASE = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "https://app.noted.co.uk"
URL = BASE + "/tools/write/"
EMAIL = f"noted-smoke-{int(time.time())}@example.invalid"
PASSWORD = f"smoke-{int(time.time())}-0123456789"
NOTE_TITLE = "Live smoke"
NOTE_TEXT = "written by the live smoke probe; survived an induced outage"

FAILS = []


def check(label, ok, detail=""):
    print(("  PASS  " if ok else "  FAIL  ") + label + (("  — " + detail) if detail else ""))
    if not ok:
        FAILS.append(label)


def main():
    print(f"target: {URL}\naccount: {EMAIL}\n")
    with sync_playwright() as p:
        browser = p.chromium.launch()

        # ---------- session 1: register, save, outage, recover ----------
        ctx = browser.new_context()
        page = ctx.new_page()
        page.goto(URL, timeout=30000)
        page.wait_for_selector("#nw-auth:not([hidden])", timeout=15000)
        check("editor served and shows sign-in", True)

        page.fill("#nw-email", EMAIL)
        page.fill("#nw-password", PASSWORD)
        page.click("#nw-register")
        page.wait_for_selector("#nw-app:not([hidden])", timeout=15000)
        check("registered against the live engine", page.locator("#nw-app").is_visible())

        page.fill("#nw-title", NOTE_TITLE)
        page.fill("#nw-content", NOTE_TEXT)
        check("typing marks unsaved", page.locator("#nw-status").inner_text() == "Unsaved changes")

        # induced outage FIRST — the degraded path against the real page
        page.route("**/api/notes", lambda r: r.abort() if r.request.method == "POST" else r.continue_())
        page.click("#nw-save")
        page.wait_for_selector("#nw-failure:not([hidden])", timeout=10000)
        check("outage: loud banner", page.locator("#nw-failure").is_visible())
        check("outage: status NOT saved", "NOT saved" in page.locator("#nw-status").inner_text())
        check("outage: text untouched", page.locator("#nw-content").input_value() == NOTE_TEXT)
        check("outage: never said Saved", "Saved ✓" not in page.locator("#nw-status").inner_text())

        # outage ends; the SAME text goes through for real
        page.unroute("**/api/notes")
        page.click("#nw-retry")
        page.wait_for_function(
            "() => document.getElementById('nw-status').textContent.includes('Saved')",
            timeout=20000)
        check("recovery: Saved ✓ from the real engine", True)
        check("recovery: banner gone", not page.locator("#nw-failure").is_visible())
        ctx.close()

        # ---------- session 2: a different browser finds the note ----------
        ctx2 = browser.new_context()
        page2 = ctx2.new_page()
        page2.goto(URL, timeout=30000)
        page2.wait_for_selector("#nw-auth:not([hidden])", timeout=15000)
        page2.fill("#nw-email", EMAIL)
        page2.fill("#nw-password", PASSWORD)
        page2.click("#nw-signin")
        page2.wait_for_selector("#nw-app:not([hidden])", timeout=15000)
        page2.wait_for_selector("#nw-list li", timeout=10000)
        check("second session: note listed", NOTE_TITLE in page2.locator("#nw-list").inner_text())
        page2.click("#nw-list li a")
        check("second session: same text came back",
              page2.locator("#nw-content").input_value() == NOTE_TEXT)
        ctx2.close()
        browser.close()

    print("\n" + ("LIVE SMOKE PASSED" if not FAILS else f"{len(FAILS)} FAILED: {FAILS}"))
    return 1 if FAILS else 0


if __name__ == "__main__":
    sys.exit(main())
