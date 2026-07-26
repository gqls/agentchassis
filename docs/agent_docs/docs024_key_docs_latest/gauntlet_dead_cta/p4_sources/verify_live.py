#!/usr/bin/env python3
"""Verify the DEPLOYED pages, in a real browser, with no harness in the way.

Loaded straight from https://vonc.com, so the browser's own CORS check against
tools.apis.uk is exercised for real — the one thing the local harness had to
bypass. Waits are generous because the AI calls take 8-18 seconds.
"""
import sys
from playwright.sync_api import sync_playwright

GAUNTLET = "https://vonc.com/tools/gauntlet/index.html"
ARCHIVE = "https://vonc.com/provocations/index.html"

results, console_errors = [], []


def check(name, ok, detail=""):
    results.append((name, ok, detail))
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + detail) if detail else ""), flush=True)


def run(profile, width, height):
    print("\n=== LIVE %s (%dx%d) ===" % (profile, width, height), flush=True)
    with sync_playwright() as pw:
        b = pw.chromium.launch()
        page = b.new_page(viewport={"width": width, "height": height})
        page.on("console", lambda m: console_errors.append(m.text) if m.type == "error" else None)
        page.on("pageerror", lambda e: console_errors.append("pageerror: " + str(e)))

        r = page.goto(GAUNTLET + "?cb=verify" + profile, wait_until="networkidle")
        check("gauntlet_page_ok", r.status == 200, "HTTP %d" % r.status)
        for sel in ["[data-gi-enter-btn]", "[data-gi-status]", "[data-gi-challenge-body]",
                    "[data-gi-position-input]", "[data-gi-position-submit]",
                    "[data-gi-opponent-position]", "[data-gi-opponent-challenge]",
                    "[data-gi-defence-input]", "[data-gi-defence-submit]",
                    "[data-gi-verdict]", "[data-gi-verdict-reasons]", "[data-gi-pct]"]:
            check("exists " + sel, page.locator(sel).count() > 0)
        check("no fabricated stats on the page",
              not any(s in page.content() for s in ["1,284", "94,210", "12,847", "62% Disagree", "Win Rate"]))
        check("no dead anchors in the tool", page.evaluate("""() => {
            const s = document.querySelector('[data-component="gauntlet-interface"]');
            return [...s.querySelectorAll('a')].map(a => a.getAttribute('href'))
                     .filter(h => h === null || !h.trim() || h.trim() === '#').length === 0;}"""))

        # Journey C, for real, through the browser's own CORS check.
        page.click("[data-gi-enter-btn]")
        page.wait_for_function(
            "() => document.querySelector('[data-gi-round-state]').textContent.includes('live')",
            timeout=45000)
        check("C.1 round starts against the live API", True)
        check("C.1 challenge body populated",
              len(page.locator("[data-gi-challenge-body]").inner_text().strip()) > 40)
        page.wait_for_timeout(1500)
        check("C.1 clock running",
              page.locator("[data-gi-timer]").inner_text().strip() != "20:00",
              page.locator("[data-gi-timer]").inner_text().strip())

        # /position and /defend fail intermittently upstream (~15% measured,
        # bugs_open pending) — retry once so a flaky engine is not reported as a
        # broken page, and say so plainly if both attempts fail.
        got = False
        for attempt in (1, 2):
            page.fill("[data-gi-position-input]",
                      "Personalisation is what people asked for; the shared internet was an artefact of scarcity.")
            page.click("[data-gi-position-submit]")
            try:
                page.wait_for_function(
                    "() => document.querySelector('[data-gi-opponent-position]').textContent.trim().length > 40",
                    timeout=60000)
                got = True
                break
            except Exception:
                st = page.locator("[data-gi-status]").inner_text().strip()
                print("      (attempt %d did not return: %s)" % (attempt, st[:100]), flush=True)
        check("C.2 real AI counter-position", got and
              len(page.locator("[data-gi-opponent-position]").inner_text().strip()) > 40,
              "" if got else "no counter-position after 2 attempts")
        if not got:
            b.close(); return
        check("C.2 opponent challenge populated",
              len(page.locator("[data-gi-opponent-challenge]").inner_text().strip()) > 20)
        check("C.2 exactly one objective marked",
              page.locator("[data-gi-obj].is-complete").count() == 1)

        page.fill("[data-gi-defence-input]",
                  "No platform has offered a labelled shared feed as a real option, so revealed preference proves nothing.")
        page.click("[data-gi-defence-submit]")
        try:
            page.wait_for_function(
                "() => document.querySelector('[data-gi-verdict]').textContent.trim().length > 0",
                timeout=90000)
            check("C.3 verdict returned", True,
                  page.locator("[data-gi-verdict]").inner_text().strip()[:50])
            check("C.3 verdict reasons populated",
                  len(page.locator("[data-gi-verdict-reasons]").inner_text().strip()) > 40)
            check("C.3 all three objectives marked",
                  page.locator("[data-gi-obj].is-complete").count() == 3)
            check("C.3 progress reads 100%", "100%" in page.locator("[data-gi-pct]").inner_text())
        except Exception:
            st = page.locator("[data-gi-status]").inner_text().strip()
            check("C.3 verdict OR honest offline message", "offline" in st.lower(), st[:120])
            check("C.3 nothing faked on failure",
                  page.locator("[data-gi-obj].is-complete").count() <= 2)

        check("gauntlet no horizontal overflow", page.evaluate(
            "() => document.documentElement.scrollWidth - document.documentElement.clientWidth") <= 2)

        # --- archive ---
        r2 = page.goto(ARCHIVE + "?cb=verify" + profile, wait_until="networkidle")
        check("archive_page_ok", r2.status == 200, "HTTP %d" % r2.status)
        check("archive_list_exists", page.locator(".provocations-archive__list").count() > 0)
        nl = page.locator(".provocations-archive__item--linked").count()
        ns = page.locator(".provocations-archive__item--static").count()
        check("archive linked/static split", nl == 7 and ns == 1, "linked=%d static=%d" % (nl, ns))
        if nl:
            check("archive_detail_region_exists", page.locator(".provocations-archive__detail").count() > 0)
            check("detail empty before opening",
                  page.locator(".provocations-archive__detail").inner_text().strip() == "")
            slug = page.locator(".provocations-archive__item--linked").first.get_attribute("data-slug")
            page.locator(".provocations-archive__item--linked").first.click()
            page.wait_for_selector(".provocations-archive__detail:not([hidden])", timeout=8000)
            check("B.1 detail populates on click",
                  len(page.locator(".provocations-archive__detail").inner_text().strip()) > 200)
            check("B.1 URL deep-links", "entry=" + slug in page.url, page.url)
            page.goto(ARCHIVE + "?entry=" + slug, wait_until="networkidle")
            page.wait_for_selector(".provocations-archive__detail:not([hidden])", timeout=10000)
            check("B.2 deep link works on direct load",
                  len(page.locator(".provocations-archive__detail").inner_text().strip()) > 200)
            check("B.3 static entry not interactive", page.evaluate("""() => {
                const e = document.querySelector('.provocations-archive__item--static');
                return !!e && e.getAttribute('href') === null && e.getAttribute('tabindex') === null;}"""))
        check("archive no horizontal overflow", page.evaluate(
            "() => document.documentElement.scrollWidth - document.documentElement.clientWidth") <= 2)
        b.close()


run("desktop", 1366, 900)
run("mobile", 390, 844)
real = [e for e in console_errors if "favicon" not in e.lower()]
check("no console errors", not real, " | ".join(real[:3]))
failed = [r for r in results if not r[1]]
print("\n%d passed, %d failed" % (len(results) - len(failed), len(failed)))
sys.exit(1 if failed else 0)
