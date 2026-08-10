#!/usr/bin/env python3
"""Verify the exchange card AFTER RFC_020 §5.4 added the verdict-scope line.

Forked from verify_card_2026-07-31.py. Checks 1-4 are that harness unchanged
(parse, render, the missing-field rail, and the rail's own positive control) and
are kept because the §5.4 edit moved the footer geometry those checks sit on.

Checks 5-8 are new, and 7 is the one that matters. Adding a line to a canvas is
not like adding one to a page: canvas clips nothing, reports nothing, and throws
nothing. A block of prose drawn over the ruling line produces a card that is
still 1200x630, still non-null, still a valid PNG, and unreadable. Every check
in the 07-31 harness passes on that card. So the new geometry is asserted from
the DRAWN PIXELS, against the LONGEST round the pool has actually produced, and
check 8 removes the fix to prove check 7 can fail.

Run from a $HOME directory: chromium here is a snap and cannot write to /tmp.
"""
import json
import pathlib
import re
import subprocess
import sys

BASE = pathlib.Path("/home/ant/projects/agentchassis/docs/agent_docs/"
                    "docs024_key_docs_latest/gauntlet_dead_cta/p4_sources")
JS = BASE / "gauntlet_js_2026-08-10_verdict_scope.js"
ROUND = BASE / "round_real.json"
CHROME = ["chromium", "--headless", "--disable-gpu", "--no-sandbox", "--hide-scrollbars"]

src = JS.read_text(encoding="utf-8")
rnd = json.loads(ROUND.read_text(encoding="utf-8"))
fails = []


def check(name, ok, detail=""):
    print(("[PASS] " if ok else "[FAIL] ") + name + (("  " + detail) if detail else ""))
    if not ok:
        fails.append(name + " :: " + detail)


def extract(name, source=None):
    """Pull one top-level function out of the source by brace balance."""
    s = source if source is not None else src
    i = s.find("  function %s(" % name)
    if i < 0:
        sys.exit("cannot find function %s" % name)
    depth, j = 0, s.index("{", i)
    for k in range(j, len(s)):
        if s[k] == "{":
            depth += 1
        elif s[k] == "}":
            depth -= 1
            if depth == 0:
                return s[i:k + 1]
    sys.exit("unbalanced braces in %s" % name)


funcs = extract("wrapLines") + "\n\n" + extract("buildVerdictCard")

# TEST-ONLY instrumentation. `size` is a local, and the fitted type size is the
# quantity the FOOT change actually moves, so the harness has to be able to read
# it. This adds one assignment and changes no behaviour; every check that cares
# about pixels runs against the UNinstrumented `funcs`.
FIT_LOOP = "while (size > 12 && heightAt(size) > USABLE) size--;"
probed = funcs.replace(FIT_LOOP, FIT_LOOP + " window.__fit = size;")
assert probed != funcs, "could not instrument the fit loop"

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
    check("the whole file parses in a browser", True)
else:
    # The real file is an IIFE that needs the page DOM; a missing-section abort
    # is expected and is NOT a syntax error. Only a genuine SyntaxError fails.
    check("parses (runtime abort without the page DOM is expected)",
          "SyntaxError" not in title and "Unexpected" not in title, title[:70])

# ── the shared harness ──────────────────────────────────────────────────────
# Draws the card, then reports geometry measured off the pixels. Results come
# back in a <pre>, not the title, because there are more of them than a title
# can carry.
HARNESS = """<meta charset="utf-8">
<style>html,body{margin:0;padding:0;overflow:hidden}canvas{display:block}
#out{display:none}</style>
<!-- display:none, so the screenshot is the CARD and nothing else. --dump-dom
     reads textContent regardless of display, so the results still come back.
     The first version left this visible and it pushed the card down the page:
     the pixel checks were unaffected (they read the canvas, not the screenshot)
     but the saved PNG had a JSON blob across the top and the address line
     cropped off the bottom — and the PNG is the artefact a human looks at. -->
<pre id="out"></pre>
<script>
var R = %s;
var state = { provocation: { headline: R.headline } };
var el = {
  verdict:           { textContent: R.verdict },
  opponentChallenge: { textContent: R.challenge },
  defenceInput:      { value: R.defence }
};
%s
var card = buildVerdictCard();
var res = {};
if (!card) { res.card = null; }
else {
  res.card = card.width + "x" + card.height;
  var cx = card.getContext("2d");
  var px = cx.getImageData(0, 0, card.width, card.height).data;
  // The card background is a literal in the renderer; read it off the pixels
  // anyway, at a corner the drawing never reaches, so this cannot silently
  // agree with a background that changed.
  var o = (5 * card.width + 400) * 4;
  var BG = [px[o], px[o+1], px[o+2]];
  res.bg = BG;
  // Ink = anything that is not the background, skipping the amber spine in the
  // first 14 columns. Tolerance absorbs antialiasing, not glyphs.
  function rowInk(y) {
    var n = 0;
    for (var x = 20; x < card.width; x++) {
      var i = (y * card.width + x) * 4;
      if (Math.abs(px[i]-BG[0]) > 12 || Math.abs(px[i+1]-BG[1]) > 12 || Math.abs(px[i+2]-BG[2]) > 12) n++;
    }
    return n;
  }
  res.rows = {};
  for (var y = 440; y < card.height; y++) res.rows[y] = rowInk(y);
  // The scope line's own drawn colour, sampled at its densest row.
  var best = 0, bestY = 0;
  for (var y2 = 548; y2 <= 572; y2++) { var v = rowInk(y2); if (v > best) { best = v; bestY = y2; } }
  res.scopeRow = bestY;
  // Take the pixel FURTHEST from the background on that row, not the first one
  // over a threshold. The first is always a leading antialiased edge — sampling
  // it measured [218,184,157] for a glyph actually drawn #fde68a, and reported
  // 3.83:1 for a colour that is 5.71:1. An edge pixel is a blend of ink and
  // background by definition, so it can only ever understate the contrast.
  var seen = null, far = -1;
  for (var x3 = 20; x3 < card.width; x3++) {
    var i3 = (bestY * card.width + x3) * 4;
    var dd = Math.abs(px[i3]-BG[0]) + Math.abs(px[i3+1]-BG[1]) + Math.abs(px[i3+2]-BG[2]);
    if (dd > far) { far = dd; seen = [px[i3], px[i3+1], px[i3+2]]; }
  }
  res.scopeInk = seen;
  res.fit = (typeof window.__fit === "number") ? window.__fit : null;
  document.body.appendChild(card);
}
document.getElementById("out").textContent = JSON.stringify(res);
</script>
"""


def render(round_obj, source_funcs=funcs, tag="_card"):
    html = HARNESS % (json.dumps(round_obj, ensure_ascii=False), source_funcs)
    pathlib.Path(tag + ".html").write_text(html, encoding="utf-8")
    dom = subprocess.run(CHROME + ["--dump-dom", tag + ".html"],
                         capture_output=True, text=True, timeout=180).stdout
    m = re.search(r'<pre id="out">(.*?)</pre>', dom, re.S)
    if not m:
        return None
    return json.loads(m.group(1))


# ── 2. run the REAL renderer against a REAL round ───────────────────────────
real = render(rnd)
check("renderer returned a 1200x630 canvas from a real round",
      real and real.get("card") == "1200x630", repr(real and real.get("card")))

subprocess.run(CHROME + ["--force-device-scale-factor=1", "--window-size=1200,630",
                         "--screenshot=card_under_test_2026-08-10.png", "_card.html"],
               capture_output=True, timeout=180)
png = pathlib.Path("card_under_test_2026-08-10.png")
check("card_under_test_2026-08-10.png written",
      png.exists() and png.stat().st_size > 20000,
      "%d bytes" % (png.stat().st_size if png.exists() else 0))

# ── 3. the rail: refuse a round that is missing any piece ───────────────────
for missing in ("verdict", "challenge", "defence"):
    partial = dict(rnd)
    partial[missing] = ""
    r = render(partial, tag="_rail")
    check("refuses a round with no %s" % missing,
          r is not None and r.get("card") is None, repr(r and r.get("card")))

# ── 4. positive control for check 3 ─────────────────────────────────────────
broken = funcs.replace("if (!verdict || !challenge || !defence) return null;",
                       "if (false) return null;")
assert broken != funcs, "check 4 could not remove the guard — the control is inert"
p = dict(rnd)
p["defence"] = ""
r = render(p, source_funcs=broken, tag="_control")
check("control: with the guard removed the empty round DOES draw, so check 3 "
      "was testing the guard and not a broken harness",
      r is not None and r.get("card") == "1200x630", repr(r and r.get("card")))

# ── 5. the drawn string is the ESCAPE form, not a literal character ─────────
# Anchored on the assignment, not a grep of the file: a loose search matches the
# comment above it first, and the comment is not what gets drawn.
scope_line = [l for l in src.splitlines() if l.strip().startswith("var SCOPE =")]
check("exactly one SCOPE assignment", len(scope_line) == 1, str(len(scope_line)))
if scope_line:
    line = scope_line[0]
    check("SCOPE uses the \\u2014 escape", chr(92) + "u2014" in line)
    check("SCOPE contains no literal non-ASCII",
          all(ord(c) < 128 for c in line),
          repr([c for c in line if ord(c) >= 128]))

# ── 6. contrast of the scope line against the card's own background ─────────
def lum(rgb):
    def lin(c):
        c /= 255.0
        return c / 12.92 if c <= 0.03928 else ((c + 0.055) / 1.055) ** 2.4
    r, g, b = rgb
    return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b)


if real and real.get("scopeInk") and real.get("bg"):
    l1, l2 = lum(real["scopeInk"]), lum(real["bg"])
    ratio = (max(l1, l2) + 0.05) / (min(l1, l2) + 0.05)
    check("scope line contrast >= 4.5:1 measured off the drawn pixels",
          ratio >= 4.5, "%.2f:1  ink=%s bg=%s" % (ratio, real["scopeInk"], real["bg"]))
else:
    check("scope line contrast measurable", False, "no ink found on the scope row")

# ── 7. THE LAYOUT CHECK — nothing collides, on the longest real round ───────
# Geometry, all measured back from H=630:
#   prose must stop by H-FOOT      = 458
#   amber rule                     = 476..481
#   "The judge ruled:"  baseline   = 526
#   scope               baseline   = 566
#   address             baseline   = 606
# The gutters between them must be clean background. The first one is the whole
# point of the FOOT change: it is where a full-length round would land.
GUTTERS = [("prose -> rule", 462, 473),
           ("rule -> ruling", 484, 490),
           ("ruling -> scope", 538, 544),
           ("scope -> address", 576, 582)]

# The longest round the pool has produced, not the average one: the fit loop
# only bites at the top of the range, so an average round cannot exercise it.
long_round = dict(rnd)
long_round["challenge"] = rnd["challenge"] + " " + rnd["challenge"]
long_round["defence"] = rnd["defence"] + " " + rnd["defence"]
longr = render(long_round, tag="_long")
check("long round still renders", longr and longr.get("card") == "1200x630",
      repr(longr and longr.get("card")))

if longr and longr.get("rows"):
    rows = {int(k): v for k, v in longr["rows"].items()}
    for name, y0, y1 in GUTTERS:
        ink = {y: rows.get(y, 0) for y in range(y0, y1 + 1) if rows.get(y, 0) > 0}
        check("gutter clean (%s), long round" % name, not ink, "ink at rows " + str(ink))
    # and the scope line is actually drawn on the long round too
    drew = any(rows.get(y, 0) > 0 for y in range(548, 573))
    check("scope line present on the long round", drew)

# ── 8. positive control for check 7 ────────────────────────────────────────
#
# CORRECTED 2026-08-10. The first version of this control put FOOT back to 130
# and expected the long round to collide. It did not, and the reason matters:
# FOOT does NOT prevent collision on an ordinary round, because the fit loop
# above absorbs any reserve by shrinking the type until the prose fits. What
# FOOT actually buys is TYPE SIZE (check 9), and it only becomes load-bearing
# past the loop's own floor of 12px, where there is nothing left to shrink.
#
# So the control has to be a round that overflows AT 12px. The app's own
# MAX_CHARS is 2000 a field, so that round is not synthetic — it is the largest
# thing a visitor is allowed to submit.
MAXCH = 2000
huge = dict(rnd)
huge["challenge"] = ("word " * (MAXCH // 5))[:MAXCH]
huge["defence"] = ("lorem " * (MAXCH // 6))[:MAXCH]
h = render(huge, tag="_huge")
if h and h.get("rows"):
    rows = {int(k): v for k, v in h["rows"].items()}
    spill = {y: rows.get(y, 0) for y in range(462, 474) if rows.get(y, 0)}
    check("control: a max-length round (2000 chars a field) DOES overflow the "
          "reserve, so check 7's gutters can fail and are not vacuous",
          bool(spill), "ink 462-473: " + str(spill))
else:
    check("control rendered", False, repr(h))

# Is that overflow something §5.4 introduced, or was it already there? The
# honest answer needs the OLD geometry measured on the SAME input.
pre = funcs.replace("var TOP = 112, FOOT = 172, USABLE = H - TOP - FOOT;",
                    "var TOP = 112, FOOT = 130, USABLE = H - TOP - FOOT;")
assert pre != funcs, "could not restore the old FOOT"
h_old = render(huge, source_funcs=pre, tag="_huge_pre")
if h_old and h_old.get("rows"):
    rows_old = {int(k): v for k, v in h_old["rows"].items()}
    spill_old = {y: rows_old.get(y, 0) for y in range(462, 474) if rows_old.get(y, 0)}
    check("PRE-EXISTING: the max-length round overflowed before §5.4 too, so "
          "this is the card's own 12px floor and not a regression this change "
          "introduced",
          bool(spill_old), "ink 462-473 at FOOT=130: " + str(spill_old))

# ── 9. what the reserve actually cost: the fitted type size ────────────────
fit_new = render(rnd, source_funcs=probed, tag="_fit_new")
fit_old = render(rnd, source_funcs=pre.replace(FIT_LOOP, FIT_LOOP + " window.__fit = size;"),
                 tag="_fit_old")
a = fit_new and fit_new.get("fit")
b = fit_old and fit_old.get("fit")
check("fitted type size reported for both geometries",
      isinstance(a, int) and isinstance(b, int), "after=%s before=%s" % (a, b))
if isinstance(a, int) and isinstance(b, int):
    print("       real round (challenge %d, defence %d chars): "
          "%dpx before §5.4 -> %dpx after" % (len(rnd["challenge"]), len(rnd["defence"]), b, a))
    # 20px is the smallest the lane has called legible in a downscaled timeline
    # card; below that the prose is decoration. This is the number to watch if
    # anything else is ever added to the footer.
    check("prose still fits at >= 20px on a real round after the reserve grew",
          a >= 20, "%dpx" % a)

print()
if fails:
    print("FAILED (%d):" % len(fails))
    for f in fails:
        print("  -", f)
    sys.exit(1)
print("ALL CHECKS PASSED")
