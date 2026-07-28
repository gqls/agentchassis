#!/usr/bin/env python3
"""Verify E/F on the DEPLOYED page: card + emphasis, real CORS, one real round."""
import pathlib
from playwright.sync_api import sync_playwright

HERE = pathlib.Path(__file__).parent
SHIPPED = (HERE / "shipped_check.js").read_text()
URL = "https://vonc.com/tools/gauntlet/index.html?cb=verifyef"
results, errors = [], []


def check(name, ok, detail=""):
    results.append(ok)
    print(("PASS  " if ok else "FAIL  ") + name + ((" — " + str(detail)[:140]) if detail else ""), flush=True)


EMPH = """(i) => {
  const s = document.querySelectorAll('.gi-steps .gi-step')[i];
  return ['is-current','is-done','is-future'].find(c => s.classList.contains(c)) || 'none';
}"""

with sync_playwright() as pw:
    b = pw.chromium.launch()
    page = b.new_context(viewport={"width": 390, "height": 844}).new_page()
    page.on("pageerror", lambda e: errors.append("pageerror: " + str(e)))
    page.on("console", lambda m: errors.append(m.text) if m.type == "error" else None)
    page.goto(URL, wait_until="networkidle")
    page.wait_for_timeout(1500)

    check("sealed on load", page.evaluate(
        "() => document.querySelector('.gauntlet-interface-section').classList.contains('gi-sealed')"))
    check("0 raw placeholders", page.evaluate(
        "() => (document.documentElement.outerHTML.match(/{{\\./g) || []).length") == 0)
    over = page.evaluate(SHIPPED, "")
    check("no overflow sealed", not (over.get("over", 0) > 2 or over.get("clipped")), over)

    page.click("[data-gi-enter-btn]")
    page.wait_for_function(
        "() => !document.querySelector('.gauntlet-interface-section').classList.contains('gi-sealed')",
        timeout=30000)
    check("revealed on real /round", True)
    check("E: provocation card live", page.evaluate(
        "() => !!document.querySelector('.gi-provocation-card')"))
    check("E: accent edge painted", page.evaluate(
        "() => getComputedStyle(document.querySelector('.gi-provocation-card')).borderLeftColor") == "rgb(245, 158, 11)")
    check("E: provocation text populated inside the card", page.evaluate(
        "() => document.querySelector('.gi-provocation-card [data-gi-challenge-title]').textContent.trim().length > 10"))
    check("F: position is-current", page.evaluate(EMPH, 0) == "is-current", page.evaluate(EMPH, 0))
    check("F: defence is-future + muted", page.evaluate(EMPH, 2) == "is-future" and float(page.evaluate(
        "() => getComputedStyle(document.querySelectorAll('.gi-steps .gi-step')[2]).opacity")) < 0.7)
    over2 = page.evaluate(SHIPPED, "")
    check("no overflow revealed", not (over2.get("over", 0) > 2 or over2.get("clipped")), over2)

    real_errors = [e for e in errors if "og-card" not in e and "favicon" not in e]
    check("no unexpected page errors", not real_errors, real_errors[:3])
    b.close()

print(("ALL %d PASS" % len(results)) if all(results) else "%d FAILURES" % results.count(False))
