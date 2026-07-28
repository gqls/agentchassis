#!/usr/bin/env python3
"""131 G build (owner pick: option 2): a shareable card generated from the
REAL verdict text of the visitor's own round. Client-side canvas only — no
backend, no numbers, nothing that is not a fact of that round."""
import re

S = "/tmp/claude-1000/-home-ant-projects-agentchassis/86bed2a2-8f6f-49bd-a8f7-39f5e5f74070/scratchpad"

CSS_ADD = """
  /* G (bugs_open/131, owner pick 2026-07-28: "a shareable card generated from
     the real verdict text"). The control lives INSIDE the verdict step, which
     is display:none until /defend returns — it cannot exist before the fact. */
  .gauntlet-interface-section .gi-verdict-share {
    margin-top: 1.1rem;
  }
"""

BTN = """
              <div class="gi-verdict-share">
                <button type="button" class="gi-btn-secondary" data-gi-share-card>Save this verdict as a card</button>
              </div>"""


def edit_html(path_in, path_out):
    src = open(path_in).read()
    assert 'data-gi-share-card' not in src
    anchor = '<div class="gi-verdict-reasons" data-gi-verdict-reasons></div>'
    assert src.count(anchor) == 1, f"anchor x{src.count(anchor)}"
    out = src.replace(anchor, anchor + BTN)
    assert out.count('</style>') == 1
    out = out.replace('</style>', CSS_ADD + '</style>', 1)
    assert out.count('{') == src.count('{') + CSS_ADD.count('{')
    open(path_out, 'w').write(out)
    print(f"{path_out}: OK (+{len(out)-len(src)} chars)")


def edit_js(path_in, path_out):
    src = open(path_in).read()
    out = src
    assert 'buildVerdictCard' not in src

    def swap(old, new, n=1):
        nonlocal out
        assert out.count(old) == n, f"anchor x{out.count(old)}: {old[:50]!r}"
        out = out.replace(old, new)

    swap('    verdictReasons: section.querySelector("[data-gi-verdict-reasons]")',
         '    verdictReasons: section.querySelector("[data-gi-verdict-reasons]"),\n'
         '    shareCard: section.querySelector("[data-gi-share-card]")')

    # card builder + share handler, ahead of the counters section
    swap('  // ── character counters ─',
         "  // ── verdict share card (G — owner pick 2026-07-28: option 2) ──────────\n"
         "  //\n"
         "  // Every string drawn on the card is a fact of THIS round: the provocation\n"
         "  // the round was started on, the judge's actual verdict line, the date and\n"
         "  // the page address. No win-rate, no leaderboard, no count of anything —\n"
         "  // the fabrication classes deleted from this site stay deleted. The button\n"
         "  // sits inside the verdict step, which is hidden until /defend returns, so\n"
         "  // the control cannot fire before the verdict exists.\n"
         "  function wrapText(x, text, left, top, maxWidth, lineHeight, maxLines) {\n"
         "    var words = String(text).split(/\\s+/);\n"
         "    var line = \"\", lines = 0, y = top;\n"
         "    for (var i = 0; i < words.length; i++) {\n"
         "      var probe = line ? line + \" \" + words[i] : words[i];\n"
         "      if (x.measureText(probe).width > maxWidth && line) {\n"
         "        lines++;\n"
         "        if (lines >= maxLines) {\n"
         "          x.fillText(line.replace(/.{3}$/, \"\") + \"\\u2026\", left, y);\n"
         "          return y + lineHeight;\n"
         "        }\n"
         "        x.fillText(line, left, y);\n"
         "        y += lineHeight;\n"
         "        line = words[i];\n"
         "      } else {\n"
         "        line = probe;\n"
         "      }\n"
         "    }\n"
         "    if (line) { x.fillText(line, left, y); y += lineHeight; }\n"
         "    return y;\n"
         "  }\n\n"
         "  function buildVerdictCard() {\n"
         "    var prov = state.provocation && state.provocation.headline\n"
         "      ? String(state.provocation.headline).replace(/<[^>]*>/g, \"\") : \"\";\n"
         "    var verdict = el.verdict ? el.verdict.textContent.trim() : \"\";\n"
         "    if (!verdict) return null;\n"
         "    var W = 1200, H = 630;\n"
         "    var c = document.createElement(\"canvas\");\n"
         "    c.width = W; c.height = H;\n"
         "    var x = c.getContext(\"2d\");\n"
         "    if (!x) return null;\n"
         "    x.fillStyle = \"#6d28d9\";\n"
         "    x.fillRect(0, 0, W, H);\n"
         "    x.fillStyle = \"#f59e0b\";\n"
         "    x.fillRect(0, 0, 14, H);\n"
         "    x.fillStyle = \"rgba(255,255,255,0.85)\";\n"
         "    x.font = \"700 28px system-ui, sans-serif\";\n"
         "    x.fillText(\"THE GAUNTLET \\u2014 A JUDGED VERDICT\", 70, 90);\n"
         "    x.fillStyle = \"#ffffff\";\n"
         "    x.font = \"800 52px Georgia, serif\";\n"
         "    var afterProv = wrapText(x, \"\\u201C\" + prov + \"\\u201D\", 70, 175, W - 140, 62, 3);\n"
         "    x.fillStyle = \"#f59e0b\";\n"
         "    x.fillRect(70, afterProv + 4, 120, 6);\n"
         "    x.fillStyle = \"#ffffff\";\n"
         "    x.font = \"400 32px system-ui, sans-serif\";\n"
         "    wrapText(x, verdict, 70, afterProv + 62, W - 140, 42, 3);\n"
         "    x.fillStyle = \"rgba(255,255,255,0.7)\";\n"
         "    x.font = \"600 24px system-ui, sans-serif\";\n"
         "    x.fillText(\"vonc.com/tools/gauntlet \\u00B7 \" + new Date().toLocaleDateString(\"en-GB\"), 70, H - 48);\n"
         "    return c;\n"
         "  }\n\n"
         "  function shareVerdictCard() {\n"
         "    var c = buildVerdictCard();\n"
         "    if (!c || !c.toBlob) return;\n"
         "    c.toBlob(function (blob) {\n"
         "      if (!blob) return;\n"
         "      var file = null;\n"
         "      try { file = new File([blob], \"gauntlet-verdict.png\", { type: \"image/png\" }); } catch (e) {}\n"
         "      if (file && navigator.canShare && navigator.canShare({ files: [file] })) {\n"
         "        navigator.share({ files: [file], title: \"The Gauntlet \\u2014 my verdict\" })\n"
         "          .catch(function () { /* visitor dismissed the sheet; nothing owed */ });\n"
         "        return;\n"
         "      }\n"
         "      var a = document.createElement(\"a\");\n"
         "      a.href = URL.createObjectURL(blob);\n"
         "      a.download = \"gauntlet-verdict.png\";\n"
         "      document.body.appendChild(a);\n"
         "      a.click();\n"
         "      a.remove();\n"
         "      setTimeout(function () { URL.revokeObjectURL(a.href); }, 4000);\n"
         "    }, \"image/png\");\n"
         "  }\n\n"
         '  // ── character counters ─')

    swap('  if (el.defenceSubmit) el.defenceSubmit.addEventListener("click", submitDefence);',
         '  if (el.defenceSubmit) el.defenceSubmit.addEventListener("click", submitDefence);\n'
         '  if (el.shareCard) el.shareCard.addEventListener("click", shareVerdictCard);')

    open(path_out, 'w').write(out)
    print(f"{path_out}: OK")


edit_html(f"{S}/g_template_base.html", f"{S}/g_template_new.html")
edit_html(f"{S}/g_rendered_base.html", f"{S}/g_rendered_new.html")
edit_js(f"{S}/g_js_base.js", f"{S}/g_js_new.js")
print("ALL G EDITS APPLIED")
