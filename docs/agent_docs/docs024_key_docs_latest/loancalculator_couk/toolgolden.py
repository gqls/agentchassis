#!/usr/bin/env python3
"""toolgolden.py — capture WHAT A TOOL COMPUTES, so a rewrite can be proved equivalent.

WHY THIS EXISTS. Nothing in the platform verifies that a tool computes the right
numbers. Tier 2 `tool_acceptance` validates that selectors EXIST and says so in
its own header ("static checks CONFIRM, never refute"); `toolaudit.py` proves a
tool RESPONDS — that something changed when driven. A rewritten loan calculator
returning subtly wrong monthly payments passes both, and passes a "12 RESPONDS"
baseline too. That is the gap this closes: it records the ANSWERS.

WHAT IT DOES. Drives each tool in a real browser with deterministic inputs and
records a full behavioural fingerprint: the text of every id-bearing element,
plus every control's value, after each input vector. Then `--compare` re-runs the
identical vectors against a new implementation and diffs. No knowledge of any
individual calculator is needed or used.

WHY VECTORS DERIVED FROM THE PAGE'S OWN DEFAULTS, rather than fixed numbers.
These pages ship sensible defaults (`value="10000"`, `value="7.9"`, `value="5"`).
A fixed vector applied positionally would put an interest rate of 12345 into
field 2 and a term of 0.5 years into field 3: still deterministic, but it drives
the arithmetic into NaN/Infinity where every implementation agrees and no real
difference can show. Scaling each numeric field's OWN default (x1, x2, x0.5)
keeps every value inside its intended domain automatically, for any tool, with
no per-tool configuration.

WHY THE FINGERPRINT IS "EVERY id-BEARING ELEMENT" and not "the output field".
Guessing which element holds the answer is what made toolaudit blind twice
(fault 11/12: a wizard that reveals a step, a verdict box that unhides). Reading
everything with an id cannot guess wrong, and these scripts address their regions
by id — it is how they are written.

TWO PHASES PER VECTOR, because most of these tools are paste-then-press:
  A  set every control, dispatch input+change, read the fingerprint;
  B  click the first button carrying an onclick, read it again.
Both are recorded. `display` is captured alongside text because a tool that
responds by revealing a hidden region changes nothing else.

Usage:
  python3 toolgolden.py --out golden.json  <url> [<url> ...]
  python3 toolgolden.py --compare golden.json <url> [<url> ...]

Exit 1 on any numeric/textual divergence in --compare mode.
"""
import json
import os
import re
import sys
import tempfile
import time
import urllib.request

HARNESS = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                       "..", "webdesign_tools_repair")
sys.path.insert(0, os.path.abspath(HARNESS))
from toolprobe import CDP, start_chrome  # noqa: E402

# Deterministic scalings of each numeric field's OWN default value.
VECTORS = [("defaults", 1.0), ("double", 2.0), ("half", 0.5)]

# Fingerprint: every id-bearing element's visible text + display, and every
# control's value. Returned sorted so a diff is stable.
SNAP_JS = r"""
(() => {
  const norm = s => (s || '').replace(/\s+/g, ' ').trim().slice(0, 200);
  const out = {ids: {}, controls: {}};
  document.querySelectorAll('[id]').forEach(e => {
    const cs = getComputedStyle(e);
    out.ids[e.id] = norm(e.textContent) + '|' + cs.display +
                    (e.tagName === 'CANVAS' ? '|canvas' : '');
  });
  document.querySelectorAll('input,select,textarea').forEach((e, i) => {
    const k = e.id || e.name || (e.tagName.toLowerCase() + '#' + i);
    out.controls[k] = e.type === 'checkbox' || e.type === 'radio'
                      ? String(e.checked) : String(e.value);
  });
  return JSON.stringify(out);
})()
"""

# Some tools build their inputs on demand: consolidation and application-tracker
# start with no debt/application rows at all, so there is nothing to drive and no
# arithmetic to observe until a row exists. Pressing "add" first is what makes
# them measurable. Bounded at 2 presses so a runaway adder cannot inflate state.
SETUP_JS = r"""
(() => {
  const add = Array.from(document.querySelectorAll('button,input[type=button]'))
    .filter(e => !e.disabled && /add|new|\+|another|create/i.test(
      (e.textContent || e.value || '') + ' ' + (e.id || '')));
  add.slice(0, 1).forEach(b => { b.click(); b.click(); });
  return JSON.stringify(add.slice(0, 1).map(b => b.id || (b.textContent || '').trim().slice(0, 24)));
})()
"""

# Drive EVERY kind of control, not just numeric ones. The first version handled
# number/range only, so credit-health-check (buttons), damage-checker
# (checkboxes) and application-tracker (text) were never driven at all -- the
# harness reported "controls driven: NONE" and the inert-output gate refused the
# whole run, which is exactly what should happen but says nothing about those
# tools. Non-numeric controls are set to a FIXED deterministic value rather than
# a scaled one, because scaling a text field or a checkbox is meaningless; that
# is why the between-vector test only applies to tools with numeric inputs.
DRIVE_JS = r"""
(scale) => {
  const fired = [];
  const fire = e => ['input', 'change'].forEach(
      ev => e.dispatchEvent(new Event(ev, {bubbles: true})));
  document.querySelectorAll('input,select,textarea').forEach(e => {
    if (e.disabled) return;
    const t = (e.type || e.tagName).toLowerCase();
    if (t === 'number' || t === 'range') {
      const base = parseFloat(e.getAttribute('value'));
      if (!isFinite(base)) {
        // No default to scale: use a fixed value so the field is still driven.
        e.value = String(Math.round(1000 * scale)); fire(e);
        fired.push((e.id || e.name || '?') + '=nodefault'); return;
      }
      let v = base * scale;
      // Respect the field's own declared domain; a value outside it is what the
      // browser would clamp anyway, and clamping differs between engines.
      const mn = parseFloat(e.min), mx = parseFloat(e.max);
      if (isFinite(mn)) v = Math.max(mn, v);
      if (isFinite(mx)) v = Math.min(mx, v);
      const st = parseFloat(e.step);
      v = (isFinite(st) && st >= 1) ? Math.round(v) : Math.round(v * 100) / 100;
      e.value = String(v); fire(e);
      fired.push(e.id || e.name || '?');
    } else if (t === 'checkbox' || t === 'radio') {
      if (!e.checked) { e.click(); fired.push((e.id || e.name || '?') + '=ticked'); }
    } else if (t === 'select' || t === 'select-one') {
      if (e.options && e.options.length > 1) {
        e.selectedIndex = 1; fire(e);
        fired.push((e.id || e.name || '?') + '=opt1');
      }
    } else if (t === 'text' || t === 'textarea' || t === 'tel' || t === 'search') {
      if (!e.value) { e.value = 'Probe'; fire(e);
        fired.push((e.id || e.name || '?') + '=text'); }
    }
  });
  return JSON.stringify(fired);
}
"""

PRESS_JS = r"""
(() => {
  // First button that carries an onclick handler and is not navigation.
  const b = Array.from(document.querySelectorAll('button,input[type=button],input[type=submit]'))
    .find(e => !e.disabled && (e.getAttribute('onclick') || e.onclick));
  if (!b) return 'none';
  b.click();
  return b.id || (b.textContent || '').trim().slice(0, 40) || 'button';
})()
"""


class Runner:
    def __init__(self):
        import socket
        s = socket.socket(); s.bind(("127.0.0.1", 0))
        self.port = s.getsockname()[1]; s.close()
        self.chrome = start_chrome(self.port, tempfile.mkdtemp(prefix="toolgolden-"))

    def _tab(self):
        req = urllib.request.Request(
            "http://127.0.0.1:%d/json/new?about:blank" % self.port, method="PUT")
        return json.loads(urllib.request.urlopen(req, timeout=10).read())

    def capture(self, url):
        cdp = CDP(self._tab()["webSocketDebuggerUrl"])
        cdp.call("Runtime.enable"); cdp.call("Page.enable")

        def ev(expr, timeout=20):
            r = cdp.call("Runtime.evaluate",
                         {"expression": expr, "returnByValue": True,
                          "awaitPromise": True}, timeout=timeout)
            return r.get("result", {}).get("result", {}).get("value")

        def settle():
            """Wait for a FULLY PARSED document, not merely readyState.

            A bare `readyState == 'complete'` poll is not enough and the failure
            is silent and catastrophic for this tool's purpose. On
            standard-calc it returned a DOM in which `.container` existed (so
            the inputs were all present and drivable) while the inline
            <script> at line 77 had not yet been parsed — so `calculateLoan`
            was undefined, every driven vector produced £0.00, and the capture
            SUCCEEDED. A golden file recorded from that state says every answer
            is £0.00, and would then certify a completely broken rewrite as
            byte-perfect. Verified against the live page: parsed properly it
            reads £202.29 on defaults and £404.57 at double the principal.

            So: wait for complete, then require the script count AND the
            serialised DOM length to stop moving across consecutive reads.
            """
            for _ in range(80):
                if ev("document.readyState", timeout=10) == "complete":
                    break
                time.sleep(0.1)
            last = None
            for _ in range(40):
                cur = ev("document.querySelectorAll('script').length + ':' + "
                         "document.documentElement.outerHTML.length", timeout=10)
                if cur is not None and cur == last:
                    return cur
                last = cur
                time.sleep(0.15)
            return last

        # A modal dialog BLOCKS the renderer, so the next Runtime.evaluate
        # simply times out with no indication why. application-tracker's remove
        # button calls confirm(), and the whole 12-page run died on it. Stubbed
        # rather than handled via Page.handleJavaScriptDialog because this CDP
        # client is request/response and would have to race the event. Stated
        # because it IS a behaviour change: a tool gated on confirm() proceeds
        # as though the user accepted — deterministically, for every run and
        # every implementation, which is what a differential test needs.
        DIALOG_STUB = ("window.confirm=()=>true;window.alert=()=>undefined;"
                       "window.prompt=()=>'';'stubbed'")

        page = {}
        for name, scale in VECTORS:
            cdp.call("Page.navigate", {"url": url})
            shape = settle()
            ev(DIALOG_STUB, timeout=10)
            added = ev(SETUP_JS, timeout=20)
            # `before` is captured AFTER setup and BEFORE driving, so "does this
            # tool react at all" can be tested independently of "does its output
            # depend on the input values". Without it the only available test was
            # a between-vector comparison, which is a no-op by construction for
            # any tool that has no numeric fields to scale.
            z = ev(SNAP_JS)
            drove = ev("(%s)(%r)" % (DRIVE_JS.strip(), scale))
            a = ev(SNAP_JS)
            pressed = ev(PRESS_JS)
            b = ev(SNAP_JS)
            page[name] = {
                "scale": scale,
                # dom_shape is recorded so a mid-parse capture is VISIBLE in the
                # golden file and in every later diff, rather than inferred.
                "dom_shape": shape,
                "added": json.loads(added) if added else [],
                "drove": json.loads(drove) if drove else [],
                "pressed": pressed,
                "before": json.loads(z) if z else {},
                "after_input": json.loads(a) if a else {},
                "after_press": json.loads(b) if b else {},
            }
        return page

    def close(self):
        if self.chrome:
            self.chrome.terminate()


def _ids(page, vec, phase):
    return page[vec][phase].get("ids", {})


def reacted(page):
    """GATE A — does the tool respond to being driven AT ALL?

    Compares the fingerprint before driving against after driving and after
    pressing, within each vector. This is the honest version of the question
    `toolaudit.py`'s RESPONDS is meant to answer and cannot: its fingerprint
    includes the driven control's own value, so a page with one number input and
    no script whatsoever scores RESPONDS (proven by construction, 2026-07-31).
    Here the control values are excluded — only id-bearing elements count — so
    setting a field cannot satisfy the test by itself.
    """
    out = set()
    for vec, _ in VECTORS:
        z = _ids(page, vec, "before")
        for phase in ("after_input", "after_press"):
            cur = _ids(page, vec, phase)
            for k in set(z) | set(cur):
                if z.get(k) != cur.get(k):
                    out.add(k)
    return out


def moved_between_vectors(page):
    """GATE B — does the output actually DEPEND on the input values?

    A golden file whose outputs are identical across every vector cannot detect a
    rewrite that ignores its inputs. Same failure the decomposition prover hit
    (four passing proofs, none of which could fail on a no-op), so it is written
    in from the start rather than after being caught.

    Only meaningful for tools with numeric fields to scale: for a button- or
    checkbox-driven tool the vectors are identical BY CONSTRUCTION, so the caller
    applies this gate only where `drove` contains a scaled numeric field.
    """
    moved = set()
    for phase in ("after_input", "after_press"):
        ia, ib = _ids(page, "defaults", phase), _ids(page, "double", phase)
        for k in set(ia) | set(ib):
            if ia.get(k) != ib.get(k):
                moved.add(k)
    return moved


def scaled_numeric(page):
    """Fields driven by SCALING a numeric default — the ones gate B applies to."""
    return [d for d in page["defaults"]["drove"]
            if not any(d.endswith(s) for s in ("=ticked", "=opt1", "=text", "=nodefault"))]


def numeric_diff(a, b):
    """Report every id whose fingerprint moved, marking numeric changes."""
    out = []
    for k in sorted(set(a) | set(b)):
        va, vb = a.get(k), b.get(k)
        if va == vb:
            continue
        na = re.findall(r"-?\d[\d,]*\.?\d*", va or "")
        nb = re.findall(r"-?\d[\d,]*\.?\d*", vb or "")
        kind = "NUMBER" if na != nb else "text/display"
        out.append("      %-28s %-9s %r -> %r" % (k, kind, va, vb))
    return out


def main():
    args = sys.argv[1:]
    out_path = cmp_path = None
    if "--out" in args:
        i = args.index("--out"); out_path = args[i + 1]; del args[i:i + 2]
    if "--compare" in args:
        i = args.index("--compare"); cmp_path = args[i + 1]; del args[i:i + 2]
    urls = args
    if not urls or (not out_path and not cmp_path):
        print(__doc__); sys.exit(2)

    r = Runner()
    got, inert, broke = {}, [], []
    try:
        for u in urls:
            try:
                got[u] = r.capture(u)
            except Exception as e:
                # One page must not destroy an 11-page run, but a partial golden
                # is worse than none: the missing pages would silently never be
                # compared again. Recorded and refused below.
                broke.append((u, str(e)[:110]))
                print("CAPTURE-ERROR  %-46s %s" % (u.replace("https://", "")[:46],
                                                   str(e)[:60]))
                sys.stdout.flush()
                continue
            n = sum(len(v["after_press"].get("ids", {})) for v in got[u].values())
            react = reacted(got[u])
            moved = moved_between_vectors(got[u])
            numeric = scaled_numeric(got[u])
            drove = got[u]["defaults"]["drove"]
            flag = ""
            if not react:
                inert.append((u, drove, "does not react to being driven at all"))
                flag = "  <-- INERT"
            elif numeric and not moved:
                inert.append((u, drove, "reacts, but output is identical for every "
                                        "input value — arithmetic ignores its inputs"))
                flag = "  <-- INPUT-INDEPENDENT"
            print("captured  %-46s %2d fields  react=%-3d vary=%-3d %s"
                  % (u.replace("https://", "")[:46], n // len(VECTORS),
                     len(react), len(moved), flag))
            sys.stdout.flush()
    finally:
        r.close()

    if broke:
        print("\nREFUSING TO WRITE GOLDEN — %d page(s) failed to capture:" % len(broke))
        for u, why in broke:
            print("   %-52s %s" % (u, why))
        print("\nA partial golden file is worse than none: the missing pages would\n"
              "never be compared again, and their absence reads as 'nothing to check'.")
        sys.exit(1)

    if inert:
        print("\nREFUSING TO WRITE GOLDEN — %d tool(s) cannot be certified:" % len(inert))
        for u, drove, why in inert:
            print("   %s\n      %s\n      controls driven: %s"
                  % (u, why, drove or "NONE"))
        print("\nEither the harness drove nothing (controls driven: NONE), or the page\n"
              "was captured before its script parsed, or the tool is genuinely inert.\n"
              "A golden file recorded from this state certifies nothing and would mark\n"
              "a completely broken rewrite as correct. Fix the cause, do not record it.")
        sys.exit(1)

    if out_path:
        json.dump({"vectors": [v[0] for v in VECTORS], "pages": got},
                  open(out_path, "w"), indent=1, sort_keys=True)
        print("\nwrote %s" % out_path)
        return

    golden = json.load(open(cmp_path))["pages"]
    bad = 0
    for u in urls:
        g, n = golden.get(u), got.get(u)
        if not g:
            print("NO GOLDEN  %s" % u); bad += 1; continue
        diffs = []
        for vec, _ in VECTORS:
            for phase in ("after_input", "after_press"):
                for field in ("ids", "controls"):
                    d = numeric_diff(g[vec][phase].get(field, {}),
                                     n[vec][phase].get(field, {}))
                    if d:
                        diffs.append("   %s / %s / %s" % (vec, phase, field))
                        diffs.extend(d)
        if diffs:
            bad += 1
            print("\nDIVERGED  %s" % u)
            print("\n".join(diffs[:40]))
        else:
            print("MATCHES   %s" % u)
    if bad:
        print("\n%d of %d tools diverged from golden" % (bad, len(urls)))
        sys.exit(1)
    print("\nall %d tools reproduce their golden values exactly" % len(urls))


main()
