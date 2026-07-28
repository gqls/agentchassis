#!/usr/bin/env python3
"""131-C build: apply the sealed/reveal edits to template, rendered and JS.

Every edit asserts its anchor occurs exactly once BEFORE editing, and the
output is shape-checked (brace balance, selector presence, moved-block byte
preservation) — the workstream's own scripted-edit lesson.
"""
import sys, re

S = "/tmp/claude-1000/-home-ant-projects-agentchassis/86bed2a2-8f6f-49bd-a8f7-39f5e5f74070/scratchpad"

CSS_ADD = """
  /* Sealed entry (bugs_open/131 item C, owner ruling 2026-07-28: the button
     REVEALS the round). Before a round exists the provocation panel is hidden
     and the entry block is the only door; the sidebar (clock + rules) stays
     visible so the "How it works" anchor keeps a live target. .gi-sealed is
     removed ONLY by the JS success paths — /round returning 200, or resuming
     a genuinely live stored round — never by a click alone. */
  .gauntlet-interface-section .gi-entry {
    max-width: 560px;
    margin: 0 auto 2.5rem;
    text-align: center;
  }
  .gauntlet-interface-section .gi-entry .gi-cta-row {
    justify-content: center;
    margin-bottom: 0.75rem;
  }
  .gauntlet-interface-section .gi-entry .gi-btn-primary {
    font-size: 1.05em;
  }
  .gauntlet-interface-section .gi-entry .gi-cta-row > * {
    /* These three exist together or not at all: the row children compute
       content-box here, so max-width:100% alone still lets padding push the
       total to 358+40=398px on a 358px row — measured, twice (button, then
       link), each caught by the very clipped-cut check 131-B taught the
       fleet. border-box makes the cap include the padding; min-width:0
       defeats the flex min-content floor. */
    box-sizing: border-box;
    min-width: 0;
    max-width: 100%;
  }
  .gauntlet-interface-section .gi-entry-status {
    min-height: 1.3em;
    margin: 0;
  }
  .gauntlet-interface-section:not(.gi-sealed) .gi-entry {
    display: none;
  }
  .gauntlet-interface-section.gi-sealed .gi-challenge-panel {
    display: none;
  }
  .gauntlet-interface-section.gi-sealed .gi-arena {
    grid-template-columns: minmax(0, 1fr);
    max-width: 560px;
    margin: 0 auto;
  }
"""

ENTRY_COMMENT = """
    <!-- Sealed-entry block (bugs_open/131 item C, owner ruling 2026-07-28).
         The provocation is deliberately NOT visible before the button is
         pressed: the owner chose "the button reveals the round" over
         "position-as-entry", reversing the earlier pre-render-the-provocation
         behaviour. Pressing runs the real /round call; gi-sealed is removed
         only in that call's success handler (or when resuming a genuinely
         live stored round) — the reveal is itself an API-bound state change,
         per this page's standing rail. -->
"""


def edit_html(path_in, path_out):
    src = open(path_in).read()

    # 1. seal the section root
    root = '<section class="gauntlet-interface-section" data-component="gauntlet-interface">'
    assert src.count(root) == 1, f"root anchor x{src.count(root)}"
    out = src.replace(root, '<section class="gauntlet-interface-section gi-sealed" data-component="gauntlet-interface">')

    # 2. lift the cta-row block out of the challenge body (verbatim slice)
    start = out.find('<div class="gi-cta-row">')
    assert start != -1 and out.count('<div class="gi-cta-row">') == 1, "cta-row anchor"
    a_end = out.find('</a>', start)
    row_end = out.find('</div>', a_end)
    assert -1 not in (a_end, row_end), "cta-row bounds"
    row_end += len('</div>')
    row = out[start:row_end]
    assert 'data-gi-enter-btn' in row and 'gi-btn-secondary' in row, "cta-row content"
    # remove it plus the blank line indentation around it
    before, after = out[:start], out[row_end:]
    before = before.rstrip(' ')          # strip the row's leading indent
    out = before.rstrip('\n') + '\n' + after.lstrip('\n')

    # 3. insert the entry block after </header>
    assert out.count('</header>') == 1, "header anchor"
    # re-indent the moved row from its old depth (10 spaces) to entry depth (6)
    row_reind = '\n'.join(
        ('      ' + ln[10:] if ln.startswith(' ' * 10) else ln) for ln in row.split('\n')
    )
    entry = (ENTRY_COMMENT
             + '    <div class="gi-entry" data-gi-entry>\n'
             + '      ' + row_reind + '\n'
             + '      <p class="gi-status gi-entry-status" data-gi-entry-status role="status" aria-live="polite"></p>\n'
             + '    </div>\n')
    out = out.replace('</header>', '</header>\n' + entry, 1)

    # 4. CSS before </style>
    assert out.count('</style>') == 1, "style anchor"
    out = out.replace('</style>', CSS_ADD + '</style>', 1)

    # shape checks
    assert out.count('{') == src.count('{') + CSS_ADD.count('{'), "brace balance"
    assert out.count('data-gi-enter-btn') == 1, "one enter button"
    assert out.count('gi-entry-status') >= 2, "entry status present (css+markup)"
    ws = lambda s: re.sub(r'\s+', ' ', s)
    assert ws(row) in ws(out), "moved row content preserved"
    for sel in ['.gi-sealed .gi-challenge-panel', ':not(.gi-sealed) .gi-entry']:
        assert sel in out, f"missing {sel}"
    open(path_out, 'w').write(out)
    print(f"{path_out}: {len(src)} -> {len(out)} bytes, vars {len(re.findall(chr(123)*2, src))} -> {len(re.findall(chr(123)*2, out))}")


def edit_js(path_in, path_out):
    src = open(path_in).read()
    out = src

    def swap(old, new, n=1):
        nonlocal out
        assert out.count(old) == n, f"anchor x{out.count(old)}: {old[:60]!r}"
        out = out.replace(old, new)

    # element map gains the entry pair
    swap('    enter: section.querySelector("[data-gi-enter-btn]"),',
         '    enter: section.querySelector("[data-gi-enter-btn]"),\n'
         '    entry: section.querySelector("[data-gi-entry]"),\n'
         '    entryStatus: section.querySelector("[data-gi-entry-status]"),')

    # sealed helpers, after setStatus's closing brace
    swap('  // A 15-second wait with no feedback reads as a hang. Count it out loud.',
         '  // ── sealed entry (bugs_open/131 item C, owner ruling 2026-07-28) ────────\n'
         '  //\n'
         '  // The page opens SEALED: the provocation panel is hidden and the entry\n'
         '  // block is the only visible door. reveal() is called from exactly two\n'
         '  // places — /round returning 200, and resuming a stored round that is\n'
         '  // genuinely live or complete on the server. A click alone never reveals:\n'
         '  // the reveal is itself an API-bound state change, like every other state\n'
         '  // change in this file.\n'
         '  function sealed() {\n'
         '    return section.classList.contains("gi-sealed");\n'
         '  }\n'
         '  function reveal() {\n'
         '    section.classList.remove("gi-sealed");\n'
         '    if (el.entryStatus) el.entryStatus.textContent = "";\n'
         '  }\n'
         '  // While sealed, the main status line lives inside the hidden panel, so\n'
         '  // anything said there would be said to nobody. Speak at the entry instead.\n'
         '  function setEntryStatus(text, kind) {\n'
         '    if (!el.entryStatus) { setStatus(text, kind); return; }\n'
         '    stopWaitCounter();\n'
         '    el.entryStatus.textContent = text;\n'
         '    el.entryStatus.className = "gi-status gi-entry-status" + (kind ? " is-" + kind : "");\n'
         '  }\n\n'
         '  // A 15-second wait with no feedback reads as a hang. Count it out loud.')

    # wait counter learns a target (only startRound redirects it)
    swap('  function startWaitCounter(text) {\n'
         '    stopWaitCounter();\n'
         '    if (!el.status) return;\n'
         '    var started = Date.now();\n'
         '    el.status.className = "gi-status is-working";\n'
         '    var render = function () {\n'
         '      var secs = Math.round((Date.now() - started) / 1000);\n'
         '      el.status.textContent = text + " (" + secs + "s)";\n'
         '    };',
         '  function startWaitCounter(text, target) {\n'
         '    stopWaitCounter();\n'
         '    var box = target || el.status;\n'
         '    if (!box) return;\n'
         '    var started = Date.now();\n'
         '    box.className = box === el.entryStatus\n'
         '      ? "gi-status gi-entry-status is-working"\n'
         '      : "gi-status is-working";\n'
         '    var render = function () {\n'
         '      var secs = Math.round((Date.now() - started) / 1000);\n'
         '      box.textContent = text + " (" + secs + "s)";\n'
         '    };')

    # startRound: wait feedback at the entry while sealed; reveal on success;
    # errors to whichever status is visible
    swap('    busy(el.enter, true);\n'
         '    startWaitCounter("Drawing today\'s provocation and starting your clock…");',
         '    busy(el.enter, true);\n'
         '    startWaitCounter("Drawing today\'s provocation and starting your clock…",\n'
         '      sealed() ? el.entryStatus : el.status);')
    swap('        renderProvocation(data.provocation);\n'
         '        startClock();',
         '        renderProvocation(data.provocation);\n'
         '        reveal();\n'
         '        startClock();')
    swap('        busy(el.enter, false);\n'
         '        // No round: no clock, no objective, no pretending.\n'
         '        setStatus(explain(err, "starting the round"), "error");',
         '        busy(el.enter, false);\n'
         '        // No round: no clock, no objective, no reveal, no pretending.\n'
         '        if (sealed()) {\n'
         '          setEntryStatus(explain(err, "starting the round"), "error");\n'
         '        } else {\n'
         '          setStatus(explain(err, "starting the round"), "error");\n'
         '        }')

    # restoreRound: reveal on both live and closed stored rounds; expired-round
    # message must speak at the entry (the main status is hidden while sealed)
    swap('    if (state.verdictIn) {\n'
         '      setRoundState("Round closed");\n'
         '      renderClock();\n'
         '    } else {',
         '    reveal();\n'
         '    if (state.verdictIn) {\n'
         '      setRoundState("Round closed");\n'
         '      renderClock();\n'
         '    } else {')
    swap('      setStatus(\n'
         '        "Your previous round\'s clock ran out while the page was away. What you " +\n'
         '          "typed is still here — press Enter the Gauntlet to start a fresh round.",\n'
         '        "error"\n'
         '      );',
         '      setEntryStatus(\n'
         '        "Your previous round\'s clock ran out while the page was away. What you " +\n'
         '          "typed is still here — press Enter the Gauntlet to start a fresh round.",\n'
         '        "error"\n'
         '      );')

    # 404 wording: the Enter button is hidden once revealed, so point at the
    # control that is actually in view (filing a position auto-starts a round)
    swap('      case 404:\n'
         '        return "This round has expired. Press Enter the Gauntlet to start a fresh one.";',
         '      case 404:\n'
         '        return "This round has expired. File your position again — it starts a fresh " +\n'
         '          "round, and what you typed is kept.";')
    swap('      var why = state.roundId\n'
         '        ? "The opponent has not answered yet — file your position first."\n'
         '        : "There is no live round — press Enter the Gauntlet to start one. What you have typed is kept.";',
         '      var why = state.roundId\n'
         '        ? "The opponent has not answered yet — file your position first."\n'
         '        : "There is no live round — file your position again to start a fresh one. " +\n'
         '          "What you have typed is kept.";')

    # a 404 means the server no longer has the round: clear the id so the
    # position path's auto-start can genuinely start fresh
    swap('      .catch(function (err) {\n'
         '        busy(el.positionSubmit, false);\n'
         '        if (err && err.message === "busy") return;\n'
         '        setStatus(explain(err, "filing your position"), "error");\n'
         '      });',
         '      .catch(function (err) {\n'
         '        busy(el.positionSubmit, false);\n'
         '        if (err && err.message === "busy") return;\n'
         '        if (err && err.status === 404) { state.roundId = null; saveRound(); }\n'
         '        setStatus(explain(err, "filing your position"), "error");\n'
         '      });')
    swap('      .catch(function (err) {\n'
         '        busy(el.defenceSubmit, false);\n'
         '        setStatus(explain(err, "sending your defence"), "error");\n'
         '      });',
         '      .catch(function (err) {\n'
         '        busy(el.defenceSubmit, false);\n'
         '        if (err && err.status === 404) { state.roundId = null; saveRound(); }\n'
         '        setStatus(explain(err, "sending your defence"), "error");\n'
         '      });')

    # retire the pre-round provocation fetch: the sealed page shows no question
    # until a real round provides one
    fetch_block_start = out.find('  // Show today\'s provocation before a round starts')
    fetch_block_end = out.find('})();')
    assert fetch_block_start != -1 and fetch_block_end != -1, "fetch block bounds"
    out = (out[:fetch_block_start]
           + '  // No pre-round provocation fetch (bugs_open/131 item C, owner ruling\n'
           + '  // 2026-07-28). The page opens SEALED — the provocation is deliberately not\n'
           + '  // shown until the button press starts a real round, which returns its own\n'
           + '  // provocation. The previous behaviour (pre-rendering today\'s provocation\n'
           + '  // from /data/provocations.json "so a visitor can read what they\'d be\n'
           + '  // arguing before committing") was REVERSED by that ruling: the primary\n'
           + '  // button now reveals, so pre-showing the question would return it to a\n'
           + '  // button whose only visible effect is a clock starting. Do not "fix" this\n'
           + '  // by restoring the fetch. A resumed round still redraws its own\n'
           + '  // provocation inside restoreRound().\n'
           + '  restoreRound();\n'
           + out[fetch_block_end:])

    assert 'fetch("/data/provocations.json' not in out, "feed fetch call removed"
    assert 'Fetching today' not in out, "placeholder block removed"
    assert out.count('reveal()') >= 2, "reveal called from both paths"
    open(path_out, 'w').write(out)
    print(f"{path_out}: {len(src)} -> {len(out)} bytes")


edit_html(f"{S}/gauntlet_template_live.html", f"{S}/new_template.html")
edit_html(f"{S}/gauntlet_rendered_live.html", f"{S}/new_rendered.html")
edit_js(f"{S}/gauntlet_js_db.js", f"{S}/new_gauntlet.js")
print("ALL EDITS APPLIED")
