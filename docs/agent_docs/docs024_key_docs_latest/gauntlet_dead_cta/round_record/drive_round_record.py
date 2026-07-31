#!/usr/bin/env python3
"""Live browser checks on the Gauntlet round-record page.

    ~/.venvs/vonc_pw/bin/python drive_round_record.py <published-slug>

Proves the page in BOTH directions: that a real published round renders every
field, and that the three non-success states refuse honestly rather than
showing an empty skeleton. The negative cases are the point — a page that
renders a blank article for a bad address looks exactly like a working page in
a screenshot.

Every check appends to RESULTS and the script exits non-zero if any failed, so
a run that dies half way cannot be read as a pass. A missing dependency is a
FAILED check, never a skip: on 2026-07-31 a driver in this lane printed
"SKIP PIL unavailable" and then "ALL LIVE CHECKS PASSED" with 3 of 9 checks
never run.
"""
import json
import os
import sys
import urllib.request

BASE = "https://vonc.com/tools/gauntlet/round.html"
API = "https://tools.apis.uk/api/v1/tools/gauntlet/round/"
SHOT = os.path.expanduser("~/gauntlet_record_check.png")  # snap chromium cannot write /tmp

RESULTS = []


def check(name, ok, detail=""):
    RESULTS.append((name, ok, detail))
    print(("PASS  " if ok else "FAIL  ") + name + (("  — " + detail) if detail else ""))


def fetch_api(slug):
    # BOTH headers are load-bearing, and each omission returns 403 with a
    # different body — the two are indistinguishable by status code alone:
    #   no Origin      -> our CORS middleware, body {"error":"origin not allowed"}
    #   no User-Agent  -> Cloudflare browser integrity check, body "error code: 1010"
    # Read the body before concluding the API refused you; the second 403 never
    # reached the island at all.
    req = urllib.request.Request(API + slug, headers={
        "Origin": "https://vonc.com",
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/126 Safari/537.36",
    })
    with urllib.request.urlopen(req, timeout=20) as r:
        return json.load(r)


def main():
    if len(sys.argv) < 2:
        print("usage: drive_round_record.py <published-slug>")
        return 2
    slug = sys.argv[1]

    try:
        from playwright.sync_api import sync_playwright
    except ImportError as e:
        check("playwright importable", False, str(e) + " — use ~/.venvs/vonc_pw/bin/python")
        return 1

    truth = fetch_api(slug)
    prov = truth["provocation"]
    counter = truth["counter"]
    verdict = truth["verdict"]

    with sync_playwright() as pw:
        browser = pw.chromium.launch()
        page = browser.new_page(viewport={"width": 1280, "height": 900})
        errors = []
        page.on("console", lambda m: errors.append(m.text) if m.type == "error" else None)
        page.on("pageerror", lambda e: errors.append(str(e)))

        # ── 1. the happy path ────────────────────────────────────────────────
        page.goto(BASE + "?r=" + slug, wait_until="networkidle", timeout=45000)
        page.wait_for_timeout(1200)

        check("round article visible", page.is_visible("[data-gr-round]"))
        check("state box hidden", not page.is_visible("[data-gr-state]"))

        def txt(sel):
            return (page.text_content(sel) or "").strip()

        # Each rendered field must equal what the API actually returned. A check
        # that only asserts "non-empty" passes on the wrong round.
        fields = [
            ("headline", "[data-gr-headline]", prov["headline"].replace("<em>", "").replace("</em>", "")),
            ("provocation body", "[data-gr-provocation]", prov["body"]),
            ("position", "[data-gr-position]", truth["position"]),
            ("counter", "[data-gr-counter]", counter["counter_position"]),
            ("challenge", "[data-gr-challenge]", counter["challenge"]),
            ("defence", "[data-gr-defence]", truth["defence"]),
            ("reasons", "[data-gr-reasons]", verdict["reasons"]),
        ]
        for name, sel, want in fields:
            got = txt(sel)
            check("renders " + name, got == want.strip(),
                  "len got=%d want=%d" % (len(got), len(want.strip())))

        check("ruling line names the verdict",
              verdict["verdict"] in txt("[data-gr-verdict]"),
              repr(txt("[data-gr-verdict]")))
        check("dates line populated", len(txt("[data-gr-dates]")) > 8, repr(txt("[data-gr-dates]")))

        # The rail: nothing on this page may be a number the page did not
        # compute. provocation.stats is an array of exactly such numbers.
        body_text = page.text_content("body") or ""
        stats = prov.get("stats") or []
        leaked = [str(s) for s in stats if isinstance(s, (str, int, float)) and str(s) and str(s) in body_text]
        check("provocation.stats absent from the page", not leaked, "leaked: " + str(leaked))

        # The markup must carry no injected element from the visitor prose.
        html = page.content()
        check("no script tag inside the rendered round",
              "<script" not in (page.inner_html("[data-gr-round]") or ""))

        # CONTRAST. The first live render put the visitor's own argument at
        # rgb(139,133,176) on the purple — 2.06:1 — because .gr-text set no
        # colour and inherited one from the site chrome. Nothing in the static
        # checks could see that: the markup was correct and the text was all
        # there. So measure the COMPUTED colours in the browser and apply the
        # WCAG AA floor. This is the check that would have caught it.
        contrast = page.evaluate("""() => {
          const lin = c => { c /= 255; return c <= 0.03928 ? c/12.92 : Math.pow((c+0.055)/1.055, 2.4); };
          const lum = rgb => { const [r,g,b] = rgb; return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b); };
          const parse = s => s.match(/[\\d.]+/g).slice(0,4).map(Number);
          const over = (fg, bg) => {           // composite any alpha onto the bg
            const a = fg.length > 3 ? fg[3] : 1;
            return [0,1,2].map(i => fg[i]*a + bg[i]*(1-a));
          };
          const secBg = parse(getComputedStyle(document.querySelector('.gauntlet-record-section')).backgroundColor);
          const out = {};
          for (const [name, sel] of [
              ['prose', '[data-gr-position]'], ['defence', '[data-gr-defence]'],
              ['provocation', '[data-gr-provocation]'], ['ruling', '[data-gr-verdict]'],
              ['reasons', '[data-gr-reasons]'], ['label', '.gr-label'],
              ['headline', '[data-gr-headline]']]) {
            const el = document.querySelector(sel);
            if (!el) continue;
            let bg = secBg, n = el;
            while (n && n !== document.body) {          // nearest painted ancestor
              const c = parse(getComputedStyle(n).backgroundColor);
              if ((c.length < 4 || c[3] > 0) && !(c[0]===0&&c[1]===0&&c[2]===0&&c[3]===0)) { bg = over(c, secBg); break; }
              n = n.parentElement;
            }
            const fg = over(parse(getComputedStyle(el).color), bg);
            const l1 = lum(fg), l2 = lum(bg);
            out[name] = Math.round(100 * (Math.max(l1,l2)+0.05) / (Math.min(l1,l2)+0.05)) / 100;
          }
          return out;
        }""")
        for name, ratio in sorted(contrast.items()):
            check("contrast " + name + " >= 4.5:1", ratio >= 4.5, str(ratio) + ":1")

        check("no console errors (round)", not errors, "; ".join(errors[:3]))
        page.screenshot(path=SHOT, full_page=True)
        check("screenshot written", os.path.exists(SHOT) and os.path.getsize(SHOT) > 10000,
              SHOT)

        # ── 2. no slug ───────────────────────────────────────────────────────
        errors.clear()
        page.goto(BASE, wait_until="networkidle", timeout=45000)
        page.wait_for_timeout(600)
        check("no-slug: state box shown", page.is_visible("[data-gr-state]"))
        check("no-slug: article hidden", not page.is_visible("[data-gr-round]"))
        check("no-slug: says it needs an address",
              "one published round" in txt("[data-gr-state-head]").lower(),
              repr(txt("[data-gr-state-head]")))
        check("no-slug: no console errors", not errors, "; ".join(errors[:3]))

        # ── 3. malformed slug — must not even reach the API ──────────────────
        errors.clear()
        calls = []
        page.on("request", lambda r: calls.append(r.url) if "tools.apis.uk" in r.url else None)
        page.goto(BASE + "?r=NOT-A-SLUG!!", wait_until="networkidle", timeout=45000)
        page.wait_for_timeout(600)
        check("bad-slug: state box shown", page.is_visible("[data-gr-state]"))
        check("bad-slug: says the address is wrong",
              "not a round address" in txt("[data-gr-state-head]").lower(),
              repr(txt("[data-gr-state-head]")))
        check("bad-slug: no request was made to the API", not calls, str(calls))

        # ── 4. well-formed slug that was never published ─────────────────────
        errors.clear()
        page.goto(BASE + "?r=zzzzzzzzzz", wait_until="networkidle", timeout=45000)
        page.wait_for_timeout(1500)
        check("unknown-slug: state box shown", page.is_visible("[data-gr-state]"))
        check("unknown-slug: article hidden", not page.is_visible("[data-gr-round]"))
        check("unknown-slug: says no public record",
              "no public record" in txt("[data-gr-state-head]").lower(),
              repr(txt("[data-gr-state-head]")))

        browser.close()

    print()
    bad = [n for n, ok, _ in RESULTS if not ok]
    if bad:
        print("FAILED %d of %d: %s" % (len(bad), len(RESULTS), ", ".join(bad)))
        return 1
    print("ALL %d LIVE CHECKS PASSED" % len(RESULTS))
    return 0


if __name__ == "__main__":
    sys.exit(main())
