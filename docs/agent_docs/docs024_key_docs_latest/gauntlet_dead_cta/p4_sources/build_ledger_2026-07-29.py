#!/usr/bin/env python3
"""Opinion ledger build (owner direction 2026-07-29, PLAN § OWNER DIRECTION
2026-07-29: "a (dated) personal history of your opinions might be a goldmine").

A device-local diary of the visitor's own completed rounds: dated provocation,
the position they filed, the judge's verdict. localStorage (a HISTORY should
survive the tab; the round store stays sessionStorage because a round should
not), entries born in exactly ONE place — the /defend success handler — never
synthesised, backfilled, or written on restore. No accounts, no server copy.

Applies anchored edits to the 2026-07-28f (verdict card) triplet. Every anchor
is asserted unique before the swap, per the pattern that shipped C/D/E/F/G.
"""

S = "/tmp/claude-1000/-home-ant-projects-agentchassis/8a5e2611-422b-4596-9b52-4c3e3251ad63/scratchpad"

CSS_ADD = """
  /* Opinion ledger (owner direction 2026-07-29: "a (dated) personal history
     of your opinions"). A device-local diary of the visitor's own completed
     rounds, rendered ONLY from localStorage entries written by the /defend
     success handler — so it cannot show a round that did not happen — and
     hidden entirely while it has none. It sits OUTSIDE the arena grid so a
     returning visitor sees their record even while the page is sealed. */
  .gauntlet-interface-section .gi-ledger {
    max-width: 760px;
    margin: 3rem auto 0;
    background: var(--section-surface);
    border: 1px solid var(--section-border);
    border-radius: var(--border-radius, 6px);
    padding: 1.5rem;
    backdrop-filter: blur(10px);
    box-sizing: border-box;
  }
  .gauntlet-interface-section .gi-ledger.is-empty {
    display: none;
  }
  .gauntlet-interface-section .gi-ledger-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.75rem;
    flex-wrap: wrap;
    margin-bottom: 0.5rem;
  }
  .gauntlet-interface-section .gi-ledger-title {
    font-size: 1.05rem;
    font-weight: 800;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--color-accent, #fbbf24);
    margin: 0;
  }
  .gauntlet-interface-section .gi-ledger-count {
    font-size: 0.72rem;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--section-text-muted);
    font-variant-numeric: tabular-nums;
  }
  .gauntlet-interface-section .gi-ledger-hint {
    font-size: 0.8rem;
    color: var(--section-text-muted);
    line-height: 1.55;
    margin: 0 0 1.25rem;
  }
  .gauntlet-interface-section .gi-ledger-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .gauntlet-interface-section .gi-ledger-entry {
    border: 1px solid rgba(255,255,255,0.09);
    background: rgba(0,0,0,0.18);
    border-radius: 4px;
    padding: 1rem 1.1rem;
    box-sizing: border-box;
    max-width: 100%;
    min-width: 0;
  }
  .gauntlet-interface-section .gi-ledger-date {
    font-size: 0.68rem;
    font-weight: 800;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--color-stage, #f59e0b);
    margin-bottom: 0.4rem;
  }
  .gauntlet-interface-section .gi-ledger-provocation {
    margin: 0 0 0.6rem;
    font-size: 1.02rem;
    font-weight: 700;
    line-height: 1.45;
    color: var(--section-heading);
    overflow-wrap: anywhere;
  }
  .gauntlet-interface-section .gi-ledger-position,
  .gauntlet-interface-section .gi-ledger-verdict {
    margin: 0 0 0.45rem;
    font-size: 0.88rem;
    line-height: 1.6;
    color: var(--section-text);
    overflow-wrap: anywhere;
  }
  .gauntlet-interface-section .gi-ledger-verdict {
    margin-bottom: 0;
  }
  .gauntlet-interface-section .gi-ledger-tag {
    display: inline-block;
    font-size: 0.6rem;
    font-weight: 800;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--section-text-muted);
    margin-right: 0.5rem;
  }
  .gauntlet-interface-section .gi-ledger-foot {
    margin-top: 1.1rem;
    text-align: right;
  }
  .gauntlet-interface-section .gi-ledger-clear {
    font-size: 0.65rem;
    opacity: 0.8;
  }
"""

LEDGER_HTML = """
    <!-- Opinion ledger (owner direction 2026-07-29). Rendered ONLY from
         localStorage entries the /defend success handler wrote — the one
         place an entry can be born — so the list cannot contain a round that
         did not happen. Hidden while empty. Outside the arena grid so a
         returning visitor sees their record below the sealed door without
         starting a round. No accounts, no server copy (owner design
         constraint, PLAN section OWNER DIRECTION 2026-07-29). -->
    <section class="gi-ledger is-empty" data-gi-ledger aria-label="Your opinion ledger">
      <div class="gi-ledger-head">
        <h2 class="gi-ledger-title">Your opinion ledger</h2>
        <span class="gi-ledger-count" data-gi-ledger-count></span>
      </div>
      <p class="gi-ledger-hint">Every round you finish on this device is added here: the day's question, where you stood, and the judge's verdict &mdash; dated. The record lives in this browser only; nothing is sent anywhere, and you can erase it whenever you like.</p>
      <ol class="gi-ledger-list" data-gi-ledger-list></ol>
      <div class="gi-ledger-foot">
        <button type="button" class="gi-btn-secondary gi-ledger-clear" data-gi-ledger-clear>Erase this record</button>
      </div>
    </section>
"""

JS_LEDGER = """  // ── opinion ledger (owner direction 2026-07-29) ──────────────────────────
  //
  // "A (dated) personal history of your opinions might be a goldmine" — a
  // device-local diary of the visitor's own completed rounds. The round store
  // above is sessionStorage because a round should die with its tab; the
  // ledger is localStorage because a history's whole point is to survive it.
  // Entries are created in exactly one place — the /defend success handler —
  // so every line is a fact of a real judged round: the provocation the round
  // served, the position actually FILED (captured when /position succeeded,
  // not read back from the editable input), the judge's verdict, the date.
  // Nothing is synthesised, backfilled, or written on restore. No accounts,
  // no server copy: the record never leaves the browser, and the visitor can
  // erase it.
  var LEDGER_STORE = "vonc_gauntlet_ledger_v1";
  var LEDGER_MAX = 100;

  function readLedger() {
    try {
      var raw = localStorage.getItem(LEDGER_STORE);
      var list = raw ? JSON.parse(raw) : [];
      return Array.isArray(list) ? list : [];
    } catch (e) { return []; }
  }

  function recordLedgerEntry(verdictText) {
    if (!state.roundId || !verdictText) return;
    var entry = {
      roundId: state.roundId,
      date: new Date().toISOString(),
      provocation: state.provocation && state.provocation.headline
        ? String(state.provocation.headline).replace(/<[^>]*>/g, "") : "",
      position: state.positionText || "",
      verdict: verdictText
    };
    try {
      // One entry per round: a second verdict for the same round replaces
      // the first rather than manufacturing a second row in the diary.
      var list = readLedger().filter(function (e) {
        return e && e.roundId !== entry.roundId;
      });
      list.push(entry);
      if (list.length > LEDGER_MAX) list = list.slice(list.length - LEDGER_MAX);
      localStorage.setItem(LEDGER_STORE, JSON.stringify(list));
    } catch (e) { /* private mode: the round still happened; only the diary is unavailable */ }
    renderLedger();
  }

  function ledgerDate(iso) {
    var d = new Date(iso);
    if (isNaN(d.getTime())) return "";
    return d.toLocaleDateString("en-GB", { day: "numeric", month: "long", year: "numeric" });
  }

  function ledgerLine(tag, text, cls) {
    var p = document.createElement("p");
    p.className = cls;
    var t = document.createElement("span");
    t.className = "gi-ledger-tag";
    t.textContent = tag;
    p.appendChild(t);
    p.appendChild(document.createTextNode(text));
    return p;
  }

  function renderLedger() {
    if (!el.ledger || !el.ledgerList) return;
    var list = readLedger();
    el.ledgerList.textContent = "";
    if (!list.length) {
      el.ledger.classList.add("is-empty");
      return;
    }
    for (var i = list.length - 1; i >= 0; i--) {
      var e = list[i];
      if (!e || !e.verdict) continue;
      var li = document.createElement("li");
      li.className = "gi-ledger-entry";
      var date = document.createElement("div");
      date.className = "gi-ledger-date";
      date.textContent = ledgerDate(e.date);
      li.appendChild(date);
      if (e.provocation) {
        var q = document.createElement("blockquote");
        q.className = "gi-ledger-provocation";
        q.textContent = "\\u201C" + e.provocation + "\\u201D";
        li.appendChild(q);
      }
      if (e.position) li.appendChild(ledgerLine("You argued", e.position, "gi-ledger-position"));
      li.appendChild(ledgerLine("Verdict", e.verdict, "gi-ledger-verdict"));
      el.ledgerList.appendChild(li);
    }
    if (el.ledgerCount) {
      el.ledgerCount.textContent = list.length + (list.length === 1 ? " round on record" : " rounds on record");
    }
    el.ledger.classList.remove("is-empty");
  }

  // Erasing is two presses because there is no undo: the first arms and says
  // so on the button itself; the second, within 6 seconds, erases.
  var ledgerClearArm = null;

  function clearLedger() {
    if (!ledgerClearArm) {
      if (el.ledgerClear) el.ledgerClear.textContent = "Press again to erase for good";
      ledgerClearArm = setTimeout(function () {
        ledgerClearArm = null;
        if (el.ledgerClear) el.ledgerClear.textContent = "Erase this record";
      }, 6000);
      return;
    }
    clearTimeout(ledgerClearArm);
    ledgerClearArm = null;
    try { localStorage.removeItem(LEDGER_STORE); } catch (e) {}
    if (el.ledgerClear) el.ledgerClear.textContent = "Erase this record";
    renderLedger();
  }

"""


def edit_html(path_in, path_out):
    src = open(path_in).read()
    assert 'data-gi-ledger' not in src
    out = src

    def swap(old, new):
        nonlocal out
        assert out.count(old) == 1, f"anchor x{out.count(old)}: {old[:60]!r}"
        out = out.replace(old, new)

    swap('</style>', CSS_ADD + '</style>')
    # After the arena grid closes, before the container closes.
    swap('      </aside>\n    </div>\n  </div>\n</section>',
         '      </aside>\n    </div>\n' + LEDGER_HTML + '  </div>\n</section>')
    assert out.count('{') == src.count('{') + CSS_ADD.count('{')
    # section + -list + -count + -clear, each containing the bare prefix
    assert out.count('data-gi-ledger') == 4
    open(path_out, 'w').write(out)
    print(f"{path_out}: OK (+{len(out)-len(src)} chars)")


def edit_js(path_in, path_out):
    src = open(path_in).read()
    assert 'LEDGER_STORE' not in src
    out = src

    def swap(old, new):
        nonlocal out
        assert out.count(old) == 1, f"anchor x{out.count(old)}: {old[:60]!r}"
        out = out.replace(old, new)

    # 1. Element map: the four ledger nodes.
    swap('    shareCard: section.querySelector("[data-gi-share-card]")',
         '    shareCard: section.querySelector("[data-gi-share-card]"),\n'
         '    ledger: section.querySelector("[data-gi-ledger]"),\n'
         '    ledgerList: section.querySelector("[data-gi-ledger-list]"),\n'
         '    ledgerCount: section.querySelector("[data-gi-ledger-count]"),\n'
         '    ledgerClear: section.querySelector("[data-gi-ledger-clear]")')

    # 2. Persist the FILED position through the round store, so a reload
    #    between /position and /defend still records the true filed text.
    swap('        draftPosition: el.positionInput ? el.positionInput.value : "",',
         '        positionText: state.positionText || "",\n'
         '        draftPosition: el.positionInput ? el.positionInput.value : "",')
    swap('    state.positionFiled = !!d.positionFiled;\n'
         '    state.verdictIn = !!d.verdictIn;',
         '    state.positionFiled = !!d.positionFiled;\n'
         '    state.positionText = d.positionText || "";\n'
         '    state.verdictIn = !!d.verdictIn;')

    # 3. Capture the position as FILED, in the /position success handler.
    swap('        state.positionFiled = true;\n        completeObjective(0);',
         '        state.positionFiled = true;\n'
         '        state.positionText = text;\n'
         '        completeObjective(0);')

    # 4. The one birthplace of a ledger entry: /defend success.
    swap('        state.verdictIn = true;\n        saveRound();',
         '        state.verdictIn = true;\n'
         '        saveRound();\n'
         '        recordLedgerEntry(data.verdict || "");')

    # 5. The ledger functions, ahead of the counters section.
    swap('  // ── character counters ─', JS_LEDGER + '  // ── character counters ─')

    # 6. Wire the clear control with the other listeners.
    swap('  if (el.shareCard) el.shareCard.addEventListener("click", shareVerdictCard);',
         '  if (el.shareCard) el.shareCard.addEventListener("click", shareVerdictCard);\n'
         '  if (el.ledgerClear) el.ledgerClear.addEventListener("click", clearLedger);')

    # 7. First paint of the diary — after restore, which never writes to it.
    swap('  restoreRound();\n})();',
         '  restoreRound();\n  renderLedger();\n})();')

    open(path_out, 'w').write(out)
    print(f"{path_out}: OK (+{len(out)-len(src)} chars)")


edit_html(f"{S}/ledger_template_base.html", f"{S}/ledger_template_new.html")
edit_html(f"{S}/ledger_rendered_base.html", f"{S}/ledger_rendered_new.html")
edit_js(f"{S}/ledger_js_base.js", f"{S}/ledger_js_new.js")
print("ALL LEDGER EDITS APPLIED")
