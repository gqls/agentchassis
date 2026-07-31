#!/usr/bin/env python3
"""Splice the exchange-card renderer into the gauntlet JS.

Done in Python rather than by hand because the source stores every non-ASCII
character in a JS string as a literal \\uXXXX escape, and an editor channel that
decodes escapes cannot emit one. Here BULLET is built from a Python literal, so
what lands on disk is exactly backslash-u-0-0-B-7, matching the file's
convention (the deploy path base64s this column, so keeping it ASCII-clean
matters).
"""
import re
import sys

PATH = ("/home/ant/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/"
        "gauntlet_dead_cta/p4_sources/gauntlet_js_2026-07-31_exchange_card.js")

BULLET = "\\u00B7"   # -> the 6 characters · on disk

NEW = '''  // ── the share card: the EXCHANGE, not just the verdict ─────────────────
  //
  // Every string drawn on the card is a fact of THIS round: the provocation it
  // was started on, the challenge the engine actually put, the visitor's own
  // defence, the judge's actual ruling, the date and the page address. No
  // win-rate, no leaderboard, no count of anything — the fabrication classes
  // deleted from this site stay deleted. The button sits inside the verdict
  // step, hidden until /defend returns, so it cannot fire before a verdict.
  //
  // WHY IT CHANGED (owner ruling 2026-07-31, option "3 staged via 1"):
  // the card used to carry the provocation headline and the verdict word, and
  // the verdict word is 13 characters ("opponent wins") — so what travelled was
  // "a stranger scored badly on an argument you cannot read". Measured over the
  // 51 complete rounds stored at that date, a full round averages 3,109
  // characters, and one 1200x630 card holds ~700 legibly once a timeline has
  // downscaled it — so the whole debate provably cannot fit (it auto-fits at
  // 11px, ~4.6px in a feed). The exchange can: challenge + defence was 599
  // characters on the measured round and fits at 26px.
  //
  // Deliberately NOT on the card: the engine's counter-argument and the judge's
  // reasons — 2,285 characters between them on that round. They are the case
  // for the per-round permalink (step 2), not something to shrink onto a card.
  //
  // The card carries no per-round URL BY DESIGN: there is no per-round page
  // yet, and a link that 404s is worse than no link.

  // Measure-only wrapper: returns the lines, callers draw them. The previous
  // version wrapped and drew in one pass, which cannot be used to size a block
  // before committing to a type size.
  function wrapLines(x, text, maxWidth) {
    var words = String(text).split(/\\s+/);
    var line = "", out = [];
    for (var i = 0; i < words.length; i++) {
      var probe = line ? line + " " + words[i] : words[i];
      if (x.measureText(probe).width > maxWidth && line) {
        out.push(line);
        line = words[i];
      } else {
        line = probe;
      }
    }
    if (line) out.push(line);
    return out;
  }

  function buildVerdictCard() {
    var prov = state.provocation && state.provocation.headline
      ? String(state.provocation.headline).replace(/<[^>]*>/g, "").trim() : "";
    var verdict = el.verdict ? el.verdict.textContent.trim() : "";
    var challenge = el.opponentChallenge ? el.opponentChallenge.textContent.trim() : "";
    var defence = el.defenceInput ? el.defenceInput.value.trim() : "";

    // All four are facts of the round, so all four are required. A card with an
    // empty block would assert that one half of the exchange was blank. Both
    // the fresh path and the sessionStorage resume path populate every one of
    // these (restoreRound writes draftDefence back into the input and the
    // challenge back into its element), so this refuses only when a round
    // genuinely lacks a piece.
    if (!verdict || !challenge || !defence) return null;

    var W = 1200, H = 630, L = 70, MAXW = W - 140;
    var c = document.createElement("canvas");
    c.width = W; c.height = H;
    var x = c.getContext("2d");
    if (!x) return null;

    var BLOCKS = [
      ["VONC ASKED", challenge, "system-ui, sans-serif"],
      ["I ANSWERED", defence, "Georgia, serif"]
    ];

    // Fit the type to the round rather than truncating the round to the type: a
    // real challenge has run to 305 characters and a defence to 294, and both
    // vary per round, so a fixed size either clips or wastes the card. TOP and
    // FOOT reserve the header, the ruling line and the address. Measured
    // against the DRAWN layout, not a raw character budget — labels and chrome
    // take vertical space the prose then cannot have, and a 32px draft that was
    // "inside" a 737-character budget still overlapped its own ruling line.
    var TOP = 112, FOOT = 130, USABLE = H - TOP - FOOT;
    function heightAt(f) {
      var lh = Math.round(f * 1.3), total = 0;
      for (var i = 0; i < BLOCKS.length; i++) {
        x.font = "400 " + f + "px " + BLOCKS[i][2];
        total += 34 + wrapLines(x, BLOCKS[i][1], MAXW).length * lh + 26;
      }
      return total;
    }
    var size = 34;
    while (size > 12 && heightAt(size) > USABLE) size--;

    x.fillStyle = "#6d28d9";
    x.fillRect(0, 0, W, H);
    x.fillStyle = "#f59e0b";
    x.fillRect(0, 0, 14, H);

    x.fillStyle = "rgba(255,255,255,0.8)";
    x.font = "700 22px system-ui, sans-serif";
    x.fillText(("THE GAUNTLET __B__ " + prov.replace(/\\.$/, "")).toUpperCase(), L, 58);

    var lh = Math.round(size * 1.3), y = TOP;
    for (var b = 0; b < BLOCKS.length; b++) {
      x.fillStyle = "#fbbf24";
      x.font = "700 20px system-ui, sans-serif";
      x.fillText(BLOCKS[b][0], L, y);
      y += 34;
      x.fillStyle = "#ffffff";
      x.font = "400 " + size + "px " + BLOCKS[b][2];
      var lines = wrapLines(x, BLOCKS[b][1], MAXW);
      for (var j = 0; j < lines.length; j++) {
        x.fillText(lines[j], L, y);
        y += lh;
      }
      y += 26;
    }

    x.fillStyle = "#f59e0b";
    x.fillRect(L, H - 112, 120, 6);
    x.fillStyle = "#ffffff";
    x.font = "800 34px system-ui, sans-serif";
    x.fillText("The judge ruled: " + verdict + ".", L, H - 62);
    x.fillStyle = "rgba(255,255,255,0.7)";
    x.font = "600 22px system-ui, sans-serif";
    x.fillText("vonc.com/tools/gauntlet __B__ " + new Date().toLocaleDateString("en-GB"), L, H - 24);
    return c;
  }

'''.replace("__B__", BULLET)

src = open(PATH, encoding="utf-8").read()

START = "  // ── verdict share card"
END = "  function shareVerdictCard() {"
i, j = src.find(START), src.find(END)
if i < 0 or j < 0 or j <= i:
    sys.exit("anchors not found: start=%d end=%d" % (i, j))

out = src[:i] + NEW + src[j:]

# The old renderer must be gone and the new one present, or this is a no-op
# dressed as a success.
assert "wrapText(" not in out, "old wrapText survived"
assert "A JUDGED VERDICT" not in out, "old header string survived"
assert out.count("function wrapLines(") == 1, "wrapLines not spliced exactly once"
assert out.count("function buildVerdictCard()") == 1, "buildVerdictCard duplicated"
assert "VONC ASKED" in out and "I ANSWERED" in out, "new labels missing"
assert "The judge ruled: " in out, "ruling line missing"
assert out.count("\\u00B7") >= 2, "bullet escape did not survive as an escape"
assert "·" not in out.split("// ── the share card")[1][:6000], \
    "a RAW bullet character leaked into the new block"
assert out.count("function shareVerdictCard() {") == 1, "caller clobbered"

open(PATH, "w", encoding="utf-8").write(out)
print("spliced OK: %d -> %d chars" % (len(src), len(out)))
print("bullet escapes in file:", out.count("\\u00B7"))
