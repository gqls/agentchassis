#!/usr/bin/env python3
"""Verify the exchange-card renderer by running the SHIPPED code, not a copy.

Extracts wrapLines + buildVerdictCard verbatim out of the JS that will be
deployed, wires them to stub el/state objects holding a REAL round's values, and
renders the result to a PNG. Also parses the whole file in a browser so a syntax
error cannot reach the deploy.

Run from a $HOME directory: chromium here is a snap and cannot write to /tmp.
"""
import json
import pathlib
import re
import subprocess
import sys

JS = pathlib.Path("/home/ant/projects/agentchassis/docs/agent_docs/"
                  "docs024_key_docs_latest/gauntlet_dead_cta/p4_sources/"
                  "gauntlet_js_2026-07-31_exchange_card.js")
ROUND = pathlib.Path("round_real.json")
CHROME = ["chromium", "--headless", "--disable-gpu", "--no-sandbox", "--hide-scrollbars"]

src = JS.read_text(encoding="utf-8")
rnd = json.loads(ROUND.read_text(encoding="utf-8"))
fails = []


def extract(name):
    """Pull one top-level function out of the source by brace balance."""
    i = src.find("  function %s(" % name)
    if i < 0:
        sys.exit("cannot find function %s" % name)
    depth, j = 0, src.index("{", i)
    for k in range(j, len(src)):
        if src[k] == "{":
            depth += 1
        elif src[k] == "}":
            depth -= 1
            if depth == 0:
                return src[i:k + 1]
    sys.exit("unbalanced braces in %s" % name)


funcs = extract("wrapLines") + "\n\n" + extract("buildVerdictCard")

# ── 1. does the WHOLE file parse? ───────────────────────────────────────────
pathlib.Path("_parse.html").write_text(
    '<meta charset="utf-8"><script>window.onerror=function(m){'
    'document.title="SYNTAX_FAIL "+m};</script>'
    '<script src="gauntlet_under_test.js"></script>'
    '<script>if(!document.title)document.title="PARSE_OK"</script>',
    encoding="utf-8")
pathlib.Path("gauntlet_under_test.js").write_text(src, encoding="utf-8")
dom = subprocess.run(CHROME + ["--dump-dom", "_parse.html"],
                     capture_output=True, text=True, timeout=180).stdout
m = re.search(r"<title>([^<]*)</title>", dom)
title = m.group(1) if m else "(no title)"
if "PARSE_OK" in title:
    print("[PASS] the whole file parses in a browser")
else:
    # The real file is an IIFE that needs the page DOM; a missing-section abort
    # is expected and is NOT a syntax error. Only a genuine SyntaxError fails.
    if "SyntaxError" in title or "Unexpected" in title:
        fails.append("syntax error: %s" % title)
        print("[FAIL] %s" % title)
    else:
        print("[PASS] parses (runtime abort without the page DOM is expected: %s)"
              % title[:70])

# ── 2. run the REAL renderer against a REAL round ───────────────────────────
harness = """<meta charset="utf-8">
<style>html,body{margin:0;padding:0;overflow:hidden}canvas{display:block}</style>
<canvas id="probe" width="10" height="10"></canvas>
<script>
var R = %s;
// Stubs shaped exactly like the page's own el/state, holding real round values.
var state = { provocation: { headline: R.headline } };
var el = {
  verdict:           { textContent: R.verdict },
  opponentChallenge: { textContent: R.challenge },
  defenceInput:      { value: R.defence }
};
var __created = null;
var realCreate = document.createElement.bind(document);
document.createElement = function (t) {
  var n = realCreate(t);
  if (t === "canvas") __created = n;
  return n;
};
%s
var card = buildVerdictCard();
if (!card) { document.title = "NULL_CARD"; }
else {
  document.body.innerHTML = "";
  document.body.appendChild(card);
  document.title = "CARD " + card.width + "x" + card.height;
}
</script>
""" % (json.dumps(rnd, ensure_ascii=False), funcs)
pathlib.Path("_card.html").write_text(harness, encoding="utf-8")

dom = subprocess.run(CHROME + ["--dump-dom", "_card.html"],
                     capture_output=True, text=True, timeout=180).stdout
m = re.search(r"<title>([^<]*)</title>", dom)
title = m.group(1) if m else "(none)"
if title == "CARD 1200x630":
    print("[PASS] renderer returned a 1200x630 canvas from a real round")
else:
    fails.append("renderer did not produce a 1200x630 canvas: %r" % title)
    print("[FAIL] %s" % title)

subprocess.run(CHROME + ["--force-device-scale-factor=1", "--window-size=1200,630",
                         "--screenshot=card_under_test.png", "_card.html"],
               capture_output=True, timeout=180)
png = pathlib.Path("card_under_test.png")
if png.exists() and png.stat().st_size > 20000:
    print("[PASS] card_under_test.png written (%d bytes)" % png.stat().st_size)
else:
    fails.append("no usable PNG produced")
    print("[FAIL] PNG missing or too small")

# ── 3. the rail: refuse a round that is missing any piece ───────────────────
for missing in ("verdict", "challenge", "defence"):
    partial = dict(rnd)
    partial[missing] = ""
    h = harness.replace(json.dumps(rnd, ensure_ascii=False),
                        json.dumps(partial, ensure_ascii=False), 1)
    pathlib.Path("_rail.html").write_text(h, encoding="utf-8")
    d = subprocess.run(CHROME + ["--dump-dom", "_rail.html"],
                       capture_output=True, text=True, timeout=180).stdout
    t = re.search(r"<title>([^<]*)</title>", d)
    t = t.group(1) if t else "(none)"
    if t == "NULL_CARD":
        print("[PASS] refuses a round with no %s" % missing)
    else:
        fails.append("drew a card with an empty %s (title=%r)" % (missing, t))
        print("[FAIL] drew a card with an empty %s -> %s" % (missing, t))

# ── 4. positive control: the check can distinguish pass from fail ───────────
broken = funcs.replace("if (!verdict || !challenge || !defence) return null;",
                       "if (false) return null;")
h = harness.replace(funcs, broken, 1)
p = dict(rnd); p["defence"] = ""
h = h.replace(json.dumps(rnd, ensure_ascii=False),
              json.dumps(p, ensure_ascii=False), 1)
pathlib.Path("_control.html").write_text(h, encoding="utf-8")
d = subprocess.run(CHROME + ["--dump-dom", "_control.html"],
                   capture_output=True, text=True, timeout=180).stdout
t = re.search(r"<title>([^<]*)</title>", d)
t = t.group(1) if t else "(none)"
if t == "CARD 1200x630":
    print("[PASS] control: with the guard removed the empty round DOES draw "
          "— so check 3 was testing the guard, not a broken harness")
else:
    fails.append("control failed: guard-removed build gave %r, so check 3 proves nothing" % t)
    print("[FAIL] control gave %s" % t)

print()
if fails:
    print("FAILED (%d):" % len(fails))
    for f in fails:
        print("  -", f)
    sys.exit(1)
print("ALL CHECKS PASSED")
