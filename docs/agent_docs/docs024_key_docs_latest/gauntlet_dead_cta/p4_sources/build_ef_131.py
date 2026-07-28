#!/usr/bin/env python3
"""131 E+F build: provocation-as-question card + state-driven step emphasis.

Applied identically to template (vars intact) and rendered (substituted);
slice-based so each artefact keeps its own copy inside moved regions.
Anchors assert exactly-once before editing; shape checks after.
"""
import re

S = "/tmp/claude-1000/-home-ant-projects-agentchassis/86bed2a2-8f6f-49bd-a8f7-39f5e5f74070/scratchpad"

CSS_ADD = """
  /* E (bugs_open/131): the provocation must read as the question you are
     answering. One card, structurally unlike every editorial block: its own
     surface, a stage-accent edge, and the existing intro copy restyled as the
     directive attached to the question instead of a paragraph floating below
     it. Colour: the accent edge is decorative (text colours unchanged);
     --color-stage on this purple already measured 3.31:1 as the title accent. */
  .gauntlet-interface-section .gi-provocation-card {
    background: rgba(255, 255, 255, 0.06);
    border-left: 4px solid var(--color-stage, #f59e0b);
    border-radius: 0 var(--border-radius, 0.75rem) var(--border-radius, 0.75rem) 0;
    padding: 1.4rem 1.5rem 1.1rem;
    margin-bottom: 1.5rem;
  }
  .gauntlet-interface-section .gi-provocation-card .gi-challenge-intro {
    margin: 1.1rem 0 0;
    padding-top: 0.9rem;
    border-top: 1px solid rgba(255, 255, 255, 0.12);
    font-size: 0.92em;
  }
  /* F (bugs_open/131): visual ranking between "do this now", "done" and
     "later". The classes are set ONLY by the same API-response handlers that
     advance the round (applyStepEmphasis in the JS). Muted is a RANKING,
     never a gate: every control stays enabled, and the defence button already
     explains itself when pressed out of order. inset shadow, not border —
     zero layout shift against the step's existing 1px border. */
  .gauntlet-interface-section .gi-step {
    transition: opacity 0.35s ease;
  }
  .gauntlet-interface-section .gi-step.is-future {
    opacity: 0.55;
  }
  .gauntlet-interface-section .gi-step.is-done {
    opacity: 0.78;
  }
  .gauntlet-interface-section .gi-step.is-current {
    opacity: 1;
    box-shadow: inset 3px 0 0 var(--color-stage, #f59e0b);
  }
"""


def edit_html(path_in, path_out):
    src = open(path_in).read()
    for cls in ("is-current", "is-done", "is-future", "gi-provocation-card"):
        assert cls not in src, f"{cls} already present"

    # wrap eyebrow..intro in the card (slice keeps each artefact's own copy)
    start = src.find('<div class="gi-challenge-eyebrow"')
    assert start != -1 and src.count('<div class="gi-challenge-eyebrow"') == 1
    intro_at = src.find('<p class="gi-challenge-intro">', start)
    assert intro_at != -1 and src.count('<p class="gi-challenge-intro">') == 1
    intro_end = src.find('</p>', intro_at)
    assert intro_end != -1
    intro_end += len('</p>')
    block = src[start:intro_end]
    assert 'data-gi-challenge-title' in block and 'data-gi-challenge-body' in block

    inner = '\n'.join('  ' + ln if ln.strip() else ln for ln in block.split('\n'))
    wrapped = ('<div class="gi-provocation-card">\n'
               + '            ' + inner.lstrip() + '\n'
               + '          </div>')
    out = src[:start] + wrapped + src[intro_end:]

    assert out.count('</style>') == 1
    out = out.replace('</style>', CSS_ADD + '</style>', 1)

    ws = lambda s: re.sub(r'\s+', ' ', s)
    assert ws(block) in ws(out), "wrapped block content preserved"
    assert out.count('{') == src.count('{') + CSS_ADD.count('{'), "brace balance"
    assert out.count('gi-provocation-card') == 3, "card class: css x2 + markup x1"
    open(path_out, 'w').write(out)
    print(f"{path_out}: {len(src)} -> {len(out)}, vars {src.count(chr(123)+chr(123))} -> {out.count(chr(123)+chr(123))}")


def edit_js(path_in, path_out):
    src = open(path_in).read()
    out = src
    assert 'applyStepEmphasis' not in src

    def swap(old, new, n=1):
        nonlocal out
        assert out.count(old) == n, f"anchor x{out.count(old)}: {old[:50]!r}"
        out = out.replace(old, new)

    # definition, ahead of the clock section
    swap('  // ── clock ─',
         '  // F (bugs_open/131): the steps get a visual RANKING that follows the\n'
         '  // round state — current, done, future. Called only from the handlers\n'
         '  // that advance the round on real API responses (and from restore/init,\n'
         '  // which re-derive the same state). It ranks attention and gates\n'
         '  // nothing: every control stays enabled.\n'
         '  function applyStepEmphasis() {\n'
         '    var steps = section.querySelectorAll(".gi-steps .gi-step");\n'
         '    // 0 position · 1 opponent reply · 2 defence · 3 verdict\n'
         '    var stage = state.verdictIn ? 3 : state.positionFiled ? 2 : 0;\n'
         '    for (var i = 0; i < steps.length; i++) {\n'
         '      steps[i].classList.remove("is-current", "is-done", "is-future");\n'
         '      if (i < stage || (state.verdictIn && i !== 3)) {\n'
         '        steps[i].classList.add("is-done");\n'
         '      } else if (i === stage) {\n'
         '        steps[i].classList.add("is-current");\n'
         '      } else {\n'
         '        steps[i].classList.add("is-future");\n'
         '      }\n'
         '    }\n'
         '  }\n\n'
         '  // ── clock ─')

    # call sites: every real state advance, plus restore and init
    swap('        renderProvocation(data.provocation);\n        reveal();\n        startClock();',
         '        renderProvocation(data.provocation);\n        reveal();\n        applyStepEmphasis();\n        startClock();')
    swap('        state.positionFiled = true;\n        completeObjective(0);',
         '        state.positionFiled = true;\n        completeObjective(0);\n        applyStepEmphasis();')
    swap('        stopClock();\n        setRoundState("Round closed");',
         '        applyStepEmphasis();\n        stopClock();\n        setRoundState("Round closed");')
    swap('    reveal();\n    if (state.verdictIn) {',
         '    reveal();\n    applyStepEmphasis();\n    if (state.verdictIn) {')
    swap('  updateProgress();\n  setRoundState("No round yet");',
         '  updateProgress();\n  applyStepEmphasis();\n  setRoundState("No round yet");')

    assert out.count('applyStepEmphasis()') == 6, f"defn+5 calls, got {out.count('applyStepEmphasis()')}"
    open(path_out, 'w').write(out)
    print(f"{path_out}: {len(src)} -> {len(out)}")


edit_html(f"{S}/ef_template_base.html", f"{S}/ef_template_new.html")
edit_html(f"{S}/ef_rendered_base.html", f"{S}/ef_rendered_new.html")
edit_js(f"{S}/ef_js_base.js", f"{S}/ef_js_new.js")
print("ALL E/F EDITS APPLIED")
