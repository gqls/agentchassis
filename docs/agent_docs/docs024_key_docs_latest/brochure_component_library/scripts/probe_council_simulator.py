#!/usr/bin/env python3
"""S6 gate for tool-review-council-simulator: does it OPERATE when driven?

Why this exists, and why it is not optional: every check the teaser-reveal-panel
ever passed (render harness, contrast probe, "verified live") exercised static
markup or forced DOM state directly. None of them ever called .click() or fired an
input event, so none of them could catch a JS init bug -- and one sat live for five
rounds. This probe drives the real controls in real Chromium and asserts the OUTPUT
CHANGED. A test that passes when the JS is deleted is not a test.

Modes:
  (default)         wrap the local template in a harness page and drive it
  --url <URL>       drive the SERVED page instead (post-deploy verification)
  --template <path> drive a specific template file (used to run MUTANTS: a probe
                    that cannot fail on a deliberately broken copy is not proving
                    anything about the working one)

Exit 0 only if every assertion passes; prints one line per assertion.
"""
import glob
import html as htmllib
import json
import os
import re
import subprocess
import sys
import tempfile
import urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
TEMPLATE = os.path.join(HERE, "..", "components", "tool-review-council-simulator", "template.html")

# The site's palette vars, so the local harness renders like the real page. Values
# are irrelevant to the behavioural assertions; their PRESENCE stops getComputedStyle
# returning empty strings and masking a real layout fault.
HARNESS_HEAD = """<!doctype html>
<html><head><meta charset="utf-8"><title>harness</title><style>
:root{
  --color-text:#14181f; --color-text-muted:#5a6473; --color-background:#ffffff;
  --color-surface:#f5f7fa; --color-border:#d9dee6; --color-primary:#1f5fbf;
  --color-primary-text:#ffffff; --color-secondary:#3b4a5f; --color-accent:#b4600a;
  --color-error:#b3261e;
}
body{margin:0}
</style></head><body>
"""

DRIVER = r"""
<script>
(function(){
  var out = {checks: [], fatal: null};
  function ck(name, pass, detail){ out.checks.push({name:name, pass:!!pass, detail:detail===undefined?null:detail}); }

  function emit(){
    var pre = document.createElement('pre');
    pre.id = 'PROBE_RESULT';
    pre.textContent = JSON.stringify(out);
    document.body.appendChild(pre);
  }

  // This driver is injected inline before </body>, so it executes DURING parsing --
  // earlier than the component's own DOMContentLoaded init. Driving now would measure
  // the pre-init page and report a bug that is not there (it did, on the first run).
  // Wait for load, which is strictly after the component's init either way.
  function deferUntilReady(run){
    if (document.readyState === 'complete') { run(); return; }
    window.addEventListener('load', run);
  }

  function q(id){ return document.getElementById(id); }
  function setRange(input, v){
    input.value = String(v);
    input.dispatchEvent(new Event('input', {bubbles:true}));
    input.dispatchEvent(new Event('change', {bubbles:true}));
  }
  function num(txt){ var m = String(txt).match(/-?[\d.]+/); return m ? parseFloat(m[0]) : null; }
  function countChecked(){
    return document.querySelectorAll('#rcs-seats input[type=checkbox]:checked').length;
  }

  deferUntilReady(function(){
  try {
    var root = document.querySelector('[data-component="tool-review-council-simulator"]');
    ck('component root present', !!root);
    if (!root) { throw new Error('no root'); }

    var thr = q('rcs-threshold'), rel = q('rcs-relevance'), rnd = q('rcs-rounds');
    var pass1 = q('rcs-pass1');
    ck('controls present', !!(thr && rel && rnd && pass1));

    // 1. Did init() run at all? If the script never executed, these stay '--'.
    var boot = pass1.textContent.trim();
    ck('init ran (headline is not the placeholder)', boot !== '--' && boot !== '', boot);

    var seatCount = document.querySelectorAll('#rcs-seats .rcs-seat').length;
    ck('roster built from data (26 seats)', seatCount === 26, seatCount);

    var bootChecked = countChecked();
    ck('default preset applied (8 typical seats checked)', bootChecked === 8, bootChecked);

    var rateNode = q('rcs-rate-guardian');
    var bootRate = rateNode ? rateNode.textContent.trim() : '';
    ck('seat rates rendered', /^\d+%$/.test(bootRate), bootRate);

    // The default must be the threshold we actually run (high-only), because 99 of our
    // 110 approvals carried a medium objection and passed. Asserted explicitly: the
    // slider's starting position is a factual claim about our own gate.
    ck('default threshold is the one we actually run', thr.value === '2', thr.value);
    ck('default readout says so', /actually run/i.test(q('rcs-threshold-out').title || ''),
       q('rcs-threshold-out').title);

    // 2. THE REAL TEST: move the threshold slider, assert the answer moves. Set each
    // position explicitly rather than relying on wherever the default happens to sit.
    setRange(thr, 2);
    var atHigh = num(pass1.textContent);
    ck('threshold readout updated (high)', /High severity/i.test(q('rcs-threshold-out').textContent),
       q('rcs-threshold-out').textContent);
    ck('seat rate matches the high-only figure', q('rcs-rate-guardian').textContent.trim() === '11%',
       q('rcs-rate-guardian').textContent.trim());

    setRange(thr, 1);
    var atMed = num(pass1.textContent);
    ck('threshold slider changes the result', atMed !== null && atMed < atHigh,
       atHigh + ' -> ' + atMed);
    ck('seat rate matches the medium-plus figure', q('rcs-rate-guardian').textContent.trim() === '67%',
       q('rcs-rate-guardian').textContent.trim());

    setRange(thr, 0);                      // any-objection: strictest
    var afterStrict = num(pass1.textContent);
    ck('strictest threshold is worse than loosest', afterStrict < atHigh,
       afterStrict + ' < ' + atHigh);

    setRange(thr, 2);                      // back to the measured default

    // 3. Relevance slider must move the answer too.
    var relBefore = num(pass1.textContent);
    setRange(rel, 10);
    var relAfter = num(pass1.textContent);
    ck('relevance slider changes the result', relAfter > relBefore, relBefore + ' -> ' + relAfter);
    setRange(rel, 70);

    // 4. Rounds slider must move pass-within-N but NOT first-round.
    var p1Before = num(pass1.textContent);
    var pnBefore = num(q('rcs-passn').textContent);
    setRange(rnd, 6);
    var pnAfter = num(q('rcs-passn').textContent);
    ck('rounds slider raises pass-within-N', pnAfter > pnBefore, pnBefore + ' -> ' + pnAfter);
    ck('rounds slider leaves first-round alone', num(pass1.textContent) === p1Before);
    ck('pass-within-N label tracks the slider', /6 rounds/.test(q('rcs-passn-label').textContent),
       q('rcs-passn-label').textContent);
    setRange(rnd, 2);

    // 5. Preset BUTTONS must actually click and change the roster.
    var allBtn = document.querySelector('.rcs-preset[data-preset="all"]');
    ck('preset buttons present', !!allBtn);
    allBtn.click();
    ck('preset "all" checks every seat', countChecked() === 26, countChecked());
    var allPass = num(pass1.textContent);

    document.querySelector('.rcs-preset[data-preset="minimal"]').click();
    ck('preset "minimal" narrows the panel', countChecked() === 2, countChecked());
    var minPass = num(pass1.textContent);
    ck('fewer seats pass more often', minPass > allPass, minPass + ' > ' + allPass);

    // 6. Empty state must be explicit, not a misleading 100%.
    document.querySelector('.rcs-preset[data-preset="none"]').click();
    ck('preset "clear" empties the panel', countChecked() === 0, countChecked());
    ck('zero seats shows n/a, not 100%', pass1.textContent.trim() === 'n/a', pass1.textContent.trim());
    ck('zero seats explains itself', /no seats selected/i.test(q('rcs-pass1-label').textContent));
    ck('zero seats shows blocker empty state',
       /Select at least one seat/i.test(q('rcs-blockers-list').textContent));

    document.querySelector('.rcs-preset[data-preset="typical"]').click();
    ck('preset "typical" restores 8 seats', countChecked() === 8, countChecked());

    // 7. An individual seat checkbox must recompute.
    var one = document.getElementById('rcs-cb-bug_historian');
    ck('individual seat checkbox present', !!one);
    var withBug = num(pass1.textContent);
    one.click();
    var withoutBug = num(pass1.textContent);
    ck('unchecking the harshest seat raises the pass rate', withoutBug > withBug,
       withBug + ' -> ' + withoutBug);
    ck('seat count fell to 7', countChecked() === 7, countChecked());
    one.click();
    ck('re-checking restores the original number', num(pass1.textContent) === withBug);

    // 8. Blocker chart must be populated and ordered.
    // '<1%' parses as NaN; it is a real sub-1% value and sorts below any printed
    // integer, so score it 0.5 rather than letting NaN silently pass the order check.
    var rawVals = Array.prototype.map.call(
      document.querySelectorAll('#rcs-blockers-list .rcs-bar-val'),
      function(n){ return n.textContent.trim(); });
    var vals = rawVals.map(function(t){
      return /^<1%$/.test(t) ? 0.5 : parseFloat(t);
    });
    ck('no blocker row reads a bare 0%', rawVals.indexOf('0%') === -1, JSON.stringify(rawVals));
    ck('every blocker value parsed', vals.every(function(v){ return !isNaN(v); }),
       JSON.stringify(rawVals));
    ck('blocker chart has rows', vals.length > 0, vals.length);
    var ordered = vals.every(function(v,i){ return i === 0 || vals[i-1] >= v; });
    ck('blocker chart is ranked descending', ordered, JSON.stringify(vals));
    var widths = Array.prototype.map.call(
      document.querySelectorAll('#rcs-blockers-list .rcs-bar-fill'),
      function(n){ return n.style.width; });
    ck('blocker bars have widths', widths.length > 0 && widths.every(function(w){ return /%$/.test(w); }),
       JSON.stringify(widths.slice(0,3)));

    // 9. Reality band: three measured markers plus the live "you" marker.
    ck('reality band has 3 measured markers',
       document.querySelectorAll('#rcs-reality-track .rcs-marker').length === 3);
    var you = document.querySelector('#rcs-reality-track .rcs-you');
    ck('reality band marks the user position', !!you, you ? you.style.left : null);
    ck('the 3 measured figures are listed as a legend, not printed on the track',
       document.querySelectorAll('#rcs-reality-legend li').length === 3,
       document.querySelectorAll('#rcs-reality-legend li').length);

    // Guards the defect the first screenshot showed: a label positioned with
    // translateX(-50%) at 2.6% hangs outside the track. Checked at BOTH extremes,
    // because a centre-only check passes on the broken version.
    function youLabelWithinTrack(){
      var track = q('rcs-reality-track');
      var lab = track.querySelector('.rcs-you-label');
      if (!lab) return null;
      var t = track.getBoundingClientRect(), l = lab.getBoundingClientRect();
      return {inside: l.left >= t.left - 1 && l.right <= t.right + 1,
              detail: Math.round(l.left) + '-' + Math.round(l.right) + ' in ' +
                      Math.round(t.left) + '-' + Math.round(t.right)};
    }
    var mid = youLabelWithinTrack();
    ck('user label sits inside the track at mid range', mid && mid.inside, mid && mid.detail);

    document.querySelector('.rcs-preset[data-preset="all"]').click();   // drives pass rate low
    var low = youLabelWithinTrack();
    ck('user label stays inside the track at the low end', low && low.inside, low && low.detail);

    document.querySelector('.rcs-preset[data-preset="none"]').click();
    document.getElementById('rcs-cb-mission').click();  // mission never objects -> ~100%
    var high = youLabelWithinTrack();
    ck('user label stays inside the track at the high end', high && high.inside, high && high.detail);
    document.querySelector('.rcs-preset[data-preset="typical"]').click();

    // 10. Thin-evidence honesty: the note must appear only when a thin seat is on.
    document.querySelector('.rcs-preset[data-preset="typical"]').click();
    ck('no thin note for the typical panel (all seats >= 20 runs)',
       q('rcs-thin-note').textContent.trim() === '', q('rcs-thin-note').textContent.trim());
    document.getElementById('rcs-cb-deferral_honesty').click();
    ck('thin note appears when a 4-run seat is added',
       /Thin evidence/i.test(q('rcs-thin-note').textContent));

    // 11. Nothing may overflow the viewport horizontally.
    ck('no horizontal overflow', document.documentElement.scrollWidth <= window.innerWidth + 1,
       document.documentElement.scrollWidth + ' vs ' + window.innerWidth);

  } catch (e) {
    out.fatal = String(e && e.stack || e);
  }
  emit();
  });
})();
</script>
"""


def find_chrome():
    cands = [os.environ.get("CHROME")] + sorted(
        glob.glob(os.path.expanduser("~/.cache/ms-playwright/chromium-*/chrome-linux64/chrome")),
        reverse=True,
    ) + ["/snap/bin/chromium", "/usr/bin/chromium", "/usr/bin/chromium-browser",
         "/usr/bin/google-chrome"]
    for c in cands:
        if c and os.path.exists(c):
            return c
    sys.exit("no chromium found")


def build_page(url=None, template=None):
    if url:
        req = urllib.request.Request(url, headers={
            "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 "
                          "(KHTML, like Gecko) Chrome/120.0 Safari/537.36"})
        raw = urllib.request.urlopen(req, timeout=60).read().decode("utf-8", "replace")
        base = "/".join(url.split("/")[:3]) + "/"
        if "<base " not in raw:
            raw = raw.replace("<head>", '<head><base href="%s">' % base, 1)
        return raw.replace("</body>", DRIVER + "</body>", 1)
    with open(template or TEMPLATE) as fh:
        tpl = fh.read()
    return HARNESS_HEAD + tpl + DRIVER + "</body></html>"


def main():
    url = None
    template = None
    if "--url" in sys.argv:
        url = sys.argv[sys.argv.index("--url") + 1]
    if "--template" in sys.argv:
        template = sys.argv[sys.argv.index("--template") + 1]

    chrome = find_chrome()
    page = build_page(url, template)

    with tempfile.TemporaryDirectory() as wd:
        p = os.path.join(wd, "probe.html")
        with open(p, "w") as fh:
            fh.write(page)
        r = subprocess.run(
            [chrome, "--headless=new", "--disable-gpu", "--no-sandbox", "--hide-scrollbars",
             "--window-size=1280,1000", "--virtual-time-budget=8000", "--dump-dom",
             "file://" + p],
            capture_output=True, text=True, timeout=180)

    m = re.search(r'<pre id="PROBE_RESULT">(.*?)</pre>', r.stdout, re.S)
    if not m:
        sys.stderr.write(r.stderr[-3000:] + "\n")
        sys.exit("probe produced no result (the driver script never ran)")

    d = json.loads(htmllib.unescape(m.group(1)))

    print("S6 probe: %s" % (url or template or "local template harness"))
    print("-" * 72)
    failed = 0
    for c in d["checks"]:
        flag = "PASS" if c["pass"] else "FAIL"
        if not c["pass"]:
            failed += 1
        detail = "" if c["detail"] is None else "   [%s]" % c["detail"]
        print("%s  %s%s" % (flag, c["name"], detail))

    if d.get("fatal"):
        print("\nFATAL in driver: %s" % d["fatal"])
        failed += 1

    print("-" * 72)
    print("%d checks, %d failed" % (len(d["checks"]), failed))
    sys.exit(1 if failed else 0)


if __name__ == "__main__":
    main()
