#!/usr/bin/env python3
"""Drive the SEALED provocation-card and Arena renderers locally, before delivery.

WHY THIS EXISTS
The leak this fixes is invisible to every HTML-level check (the shell is served
empty and filled by JS), so the only honest pre-delivery check is to render the
real shell with the real new feed and read what a visitor would see. Delivering
first and checking after costs a live-site round trip per mistake.

WHAT IT ASSERTS — the positive AND the negative, because a blank card passes any
"today's provocation is absent" test while being a worse page than the leak:
  1. today's provocation text appears NOWHERE in the painted output
  2. the past-provocation sample DOES appear, headline and body
  3. the route into today's sealed round is present and points at the Gauntlet
  4. no element that should carry text is left empty
  5. the Arena's today block states the seal and names no provocation

USAGE
  ~/.venvs/vonc_pw/bin/python drive_seal_2026-07-31.py

It builds the feed by running build_provocations.py, so it tests what would ship.
"""
import json
import pathlib
import re
import subprocess
import sys

HERE = pathlib.Path(__file__).resolve().parent

# Today's real provocation, held HERE ONLY as the thing that must not appear. It is
# deliberately absent from build_provocations.py; this file is not published.
SEALED_TEXT = [
    "personalised internet",
    "dividing the room",
    "quiet removal of whatever",
    "have you seen",
]


def build_feed():
    out = subprocess.run([sys.executable, str(HERE / "build_provocations.py")],
                         capture_output=True, text=True)
    if out.returncode != 0:
        print("build_provocations.py refused to emit:\n" + out.stderr)
        sys.exit(1)
    return json.loads(out.stdout)


def shell_html():
    """The real provocation-card shell, as served. Pulled from the live page so the
    test cannot pass against a shell that differs from production."""
    import urllib.request
    req = urllib.request.Request("https://vonc.com/",
                                 headers={"User-Agent": "Mozilla/5.0"})
    with urllib.request.urlopen(req, timeout=30) as r:
        html = r.read().decode("utf-8", "replace")
    m = re.search(r'<section class="provocation-card-section".*?</section>', html, re.S)
    if not m:
        print("could not find the provocation-card shell in the served home page")
        sys.exit(1)
    return m.group(0)


def main():
    from playwright.sync_api import sync_playwright

    feed = build_feed()
    loader = (HERE / "pcard_loader_2026-07-31_seal.js").read_text()
    shell = shell_html()
    sample = feed["sample"]
    failures, checks = [], 0

    page_html = (
        "<!doctype html><html><head><meta charset='utf-8'></head><body>"
        + shell +
        "<script>window.__FEED__=" + json.dumps(feed) + ";"
        "window.fetch=function(){return Promise.resolve("
        "{ok:true,status:200,json:function(){return Promise.resolve(window.__FEED__);}});};"
        "</script><script>" + loader + "</script></body></html>"
    )

    with sync_playwright() as p:
        b = p.chromium.launch()
        pg = b.new_page()
        errors = []
        pg.on("pageerror", lambda e: errors.append(str(e)))
        pg.set_content(page_html, wait_until="load")
        pg.wait_for_timeout(600)
        text = pg.evaluate("() => document.body.innerText")
        card = {
            "eyebrow": pg.eval_on_selector(".pc-eyebrow", "e => e.textContent.trim()"),
            "headline": pg.eval_on_selector(".pc-headline", "e => e.textContent.trim()"),
            "body": pg.eval_on_selector(".pc-body", "e => e.textContent.trim()"),
            "primary_label": pg.eval_on_selector(".pc-btn-primary", "e => e.textContent.trim()"),
            "primary_href": pg.eval_on_selector(".pc-btn-primary", "e => e.getAttribute('href')"),
            "secondary_label": pg.eval_on_selector(".pc-btn-secondary", "e => e.textContent.trim()"),
            "secondary_href": pg.eval_on_selector(".pc-btn-secondary", "e => e.getAttribute('href')"),
            "stat0": pg.eval_on_selector(".pc-stat-value", "e => e.textContent.trim()"),
        }
        b.close()

    print("--- what a visitor would read on the home card ---")
    for k, v in card.items():
        print("  %-16s %s" % (k, (v or "")[:88]))
    print()

    # 1. the seal holds
    for probe in SEALED_TEXT:
        checks += 1
        if probe.lower() in text.lower():
            failures.append("LEAK: today's provocation text %r is painted" % probe)

    # 2. the sample is actually shown
    checks += 1
    if sample["headline"] not in card["headline"]:
        failures.append("sample headline missing: got %r" % card["headline"])
    checks += 1
    if sample["body"][:40] not in card["body"]:
        failures.append("sample body missing: got %r" % card["body"][:60])

    # 3. the route into today's sealed round
    checks += 1
    if card["primary_href"] != "/tools/gauntlet/index.html":
        failures.append("primary CTA does not point at the Gauntlet: %r"
                        % card["primary_href"])
    checks += 1
    if not card["primary_label"]:
        failures.append("primary CTA has no label")

    # 4. nothing that should carry text is blank — a blank card passes every
    #    "the provocation is absent" test and is a worse page than the leak.
    for key in ("eyebrow", "headline", "body", "secondary_label", "stat0"):
        checks += 1
        if not card[key]:
            failures.append("%s is EMPTY — the shell was not filled" % key)

    # 5. secondary CTA resolves somewhere real
    checks += 1
    if not (card["secondary_href"] or "").startswith("/"):
        failures.append("secondary CTA href is not a site path: %r"
                        % card["secondary_href"])

    checks += 1
    if errors:
        failures.append("page errors: %s" % errors[:3])

    print("--- Arena today block, from the same feed ---")
    t = feed["today"]
    print("  seal_headline    %s" % t.get("seal_headline"))
    print("  seal_body        %s" % (t.get("seal_body") or "")[:80])
    print("  lobby card[0]    %s | %s" % (feed["arena"]["cards"][0]["title"],
                                          feed["arena"]["cards"][0]["desc"]))
    print()
    checks += 1
    if "headline" in t or "body" in t:
        failures.append("feed today still carries headline/body")

    if failures:
        print("FAILED %d of %d checks:" % (len(failures), checks))
        for f in failures:
            print("   FAIL", f)
        sys.exit(1)
    print("ALL %d CHECKS PASSED — sample shown, seal intact, nothing blank." % checks)


if __name__ == "__main__":
    main()
