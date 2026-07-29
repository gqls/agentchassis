#!/usr/bin/env python3
"""toolaudit.py — audit every tool page against what it CLAIMS, in a real browser.

WHY THIS REPLACES toolprobe.py. The old probe asked one question: "I drove a
control, did anything change?" It called that OK. The owner then opened two
tools scored OK and both were broken — fluid-typography showed no fluidity at
any width a desktop visitor actually uses, and micro-cms rendered an empty box
labelled "No Project Loaded". Neither is visible to a liveness test, and one of
them is invisible to source analysis too. "Responds" was never the claim; the
tool's own promise is.

Every check below exists because something real got past its absence:

  console        errors and unhandled rejections after load. The commonest
                 failure on this site, and it is silent to the visitor.
  subresources   every script/css/img the page requests, with its status. A
                 tool whose JS 404s looks identical to one with no JS.
  external-refs  THE GAP THAT MISSED micro-cms. The static orphan check reads
                 the page's own HTML; 13 of these tools keep their logic in
                 relative <script src> files, so every id those files address
                 was invisible. This fetches them and resolves the references
                 against the live DOM.
  dom-refs       every id the page's own inline script addresses, resolved in
                 the loaded DOM rather than in the source. Runtime beats regex:
                 an id built as `id="${f.name}"` exists in one and not the other.
  emptiness      a main region with no text and no controls. micro-cms passed
                 every other check while serving a blank white box.
  controls       what a visitor can actually touch, excluding no-ops (undo with
                 nothing to undo, an already-selected type button).
  drive          type a VALID value into the first real control and diff the
                 whole output region — markup, form values, and canvas pixels.
  press          then press the action button, because most of these tools are
                 paste-then-press and never react to typing alone.
  responsive     THE GAP THAT MISSED fluid-typography. Measure a computed style
                 at five viewport widths. A clamp() can be present, correct, and
                 pinned at its maximum across every width a desktop uses — the
                 CSS is right and the tool demonstrates nothing.

A tool passes only when it has no console errors, no failed subresources, every
reference resolves, its main region is not empty, and driving it changes
something. Anything else is reported with the reason, not averaged away.

    toolaudit.py <url> [<url> ...]      audit specific pages
    toolaudit.py --all                  every tool page on webdesign.co.uk
    toolaudit.py --json out.json --all  machine-readable, for diffing runs
"""
import json
import os
import re
import subprocess
import sys
import tempfile
import time
import urllib.parse
import urllib.request

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from toolprobe import CDP, start_chrome  # noqa: E402

DOMAIN = "https://webdesign.co.uk"
REF_ID = re.compile(r"""getElementById\(\s*['"]([A-Za-z0-9_\-:.]+)['"]\s*\)""")
REF_QS = re.compile(r"""querySelector(?:All)?\(\s*['"](#[A-Za-z0-9_\-]+)['"]\s*\)""")
# Controls that are SUPPOSED to do nothing on a fresh page cannot test liveness.
# 'default' joins this list because cubic-bezier's first button is a preset
# named Default — pressing it on a fresh page restores the state the page is
# already in, and the harness scored the tool DEAD for correctly doing nothing.
# Same family as undo-with-nothing-to-undo: a control that is SUPPOSED to be a
# no-op here cannot test liveness.
NOOP = re.compile(r"undo|redo|reset|clear|copy|download|share|print|remove|delete|back|default", re.I)
ACTION = re.compile(r"generate|run|convert|build|calculat|format|minif|compil|analys|analyz"
                    r"|check|creat|make|render|split|extract|optimi|inject|apply|go\b", re.I)


def tool_urls():
    out = subprocess.run(["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
                          "--", "psql", "-U", "clients_user", "-d", "clients_db", "-tAc",
                          "SELECT p.url FROM pages p JOIN sites s ON s.id=p.site_id "
                          "WHERE s.domain='webdesign.co.uk' AND p.page_type='tool' "
                          "AND p.status='active' ORDER BY p.url;"],
                         capture_output=True, text=True)
    # Read the URL, never construct it — a verdict from a URL you built is a
    # verdict about a page you invented (WRONG_CALLS.md, 2026-07-29).
    return [DOMAIN + u.strip() for u in out.stdout.splitlines() if u.strip()]


class Auditor:
    # The debug port is per-instance and randomised by default: two audits
    # running at once on a fixed port fight over the same Chrome and both fail
    # with a broken pipe, which prints in the same column as a site verdict.
    # Pass an explicit port only when you need a predictable one.
    def __init__(self, port=None):
        port = port or (9400 + os.getpid() % 500)
        self._init(port)

    def _init(self, port):
        self.port = port
        self.chrome = None
        self.profile = tempfile.mkdtemp(prefix="toolaudit-")
        self.done = 0

    def _restart(self):
        if self.chrome:
            self.chrome.terminate()
            time.sleep(0.5)
        self.chrome = start_chrome(self.port, tempfile.mkdtemp(prefix="toolaudit-"))
        self.done = 0

    def tab(self):
        # The DevTools endpoint stops answering around the 14th tab in one
        # instance, and a harness failure prints in the same column as a real
        # verdict. Restart well before that.
        if self.chrome is None or self.done >= 6:
            self._restart()
        self.done += 1
        req = urllib.request.Request(
            "http://127.0.0.1:%d/json/new?about:blank" % self.port, method="PUT")
        for attempt in range(3):
            try:
                return json.loads(urllib.request.urlopen(req, timeout=20).read())
            except Exception:
                if attempt == 2:
                    raise
                self._restart()

    def audit(self, url):
        r = {"url": url, "console": [], "bad_subresources": [], "unresolved": [], "self_removed": [],
             "controls": 0, "changed": False, "pressed": None, "empty_regions": [],
             "responsive": None, "verdict": "?", "why": []}
        cdp = CDP(self.tab()["webSocketDebuggerUrl"])
        cdp.call("Runtime.enable"); cdp.call("Log.enable")
        cdp.call("Page.enable"); cdp.call("Network.enable")
        cdp.call("Page.navigate", {"url": url})

        requests_seen, failed = {}, []
        end = time.time() + 9
        while time.time() < end:
            try:
                m = cdp._recv_frame(timeout=1.0)
            except Exception:
                continue
            meth, p = m.get("method"), m.get("params", {})
            if meth == "Network.requestWillBeSent":
                requests_seen[p.get("requestId")] = p.get("request", {}).get("url", "")
            elif meth == "Network.responseReceived":
                st = p.get("response", {}).get("status", 0)
                u = p.get("response", {}).get("url", "")
                if st >= 400 and not u.endswith("favicon.ico"):
                    failed.append("%s %s" % (st, u))
            elif meth == "Network.loadingFailed":
                u = requests_seen.get(p.get("requestId"), "?")
                if not u.endswith("favicon.ico"):
                    failed.append("FAILED %s" % u)
            elif meth in ("Runtime.exceptionThrown",):
                d = p.get("exceptionDetails", {})
                txt = (d.get("exception", {}) or {}).get("description") or d.get("text") or ""
                r["console"].append(txt.split("\n")[0][:160])
            elif meth == "Log.entryAdded":
                e = p.get("entry", {})
                if e.get("level") == "error" and "favicon" not in (e.get("url") or ""):
                    r["console"].append((e.get("text") or "")[:160])
        r["bad_subresources"] = sorted(set(failed))

        # --- references, including EXTERNAL scripts (the micro-cms gap) -------
        # Fetched THROUGH THE PAGE, not with urllib. The first version used
        # urllib and every page came back "UNREADABLE" — Cloudflare declines a
        # bare Python request. That printed in the same column as a real
        # verdict and would have condemned 63 working pages. In-page fetch is
        # also the more honest instrument: it sees exactly what the tool sees.
        bodies = [self._eval(cdp, "document.documentElement.outerHTML") or ""]
        fetched = self._eval(cdp, """(async () => {
            const out = [];
            for (const s of [...document.querySelectorAll('script[src]')]) {
              if (!s.src.startsWith(location.origin)) continue;
              try { const r = await fetch(s.src); out.push({u: s.src, ok: r.ok, body: r.ok ? await r.text() : ''}); }
              catch (e) { out.push({u: s.src, ok: false, body: ''}); }
            }
            return JSON.stringify(out);
        })()""") or "[]"
        for f in json.loads(fetched):
            if f["ok"]:
                bodies.append(f["body"])
            else:
                r["bad_subresources"].append("SCRIPT UNREACHABLE %s" % f["u"])
        refs = set()
        for b in bodies:
            refs |= set(REF_ID.findall(b))
            refs |= {x[1:] for x in REF_QS.findall(b)}
        if refs:
            got = self._eval(cdp, "JSON.stringify(%s.filter(i=>!document.getElementById(i)))"
                             % json.dumps(sorted(refs)))
            unresolved = json.loads(got or "[]")
            # An id that IS in the served HTML but is gone from the DOM was
            # removed by the page itself, and that is ordinary: regex-tester
            # replaces its own editor div via outerHTML at startup and works
            # perfectly. Reporting it BROKEN was a false positive that would
            # have sent a fixer at a correct tool. Only an id present in
            # NEITHER the source nor the DOM is the ported-markup defect.
            served = self._served_html(url)
            # An id the script CREATES counts as present even when it lands in
            # a document this one cannot see: micro-cms's editor.js does
            # style.id = "micro-cms-styles" inside its iframe, so the top
            # document will never resolve it and the tool is perfectly correct.
            created = "\n".join(bodies)
            def known(i):
                return (('id="%s"' % i) in served or ("id='%s'" % i) in served
                        or ('.id = "%s"' % i) in created or (".id = '%s'" % i) in created
                        or ('.id="%s"' % i) in created or (".id='%s'" % i) in created
                        or ("'%s'" % i) in created.replace("getElementById('%s')" % i, "")
                           and ("setAttribute" in created))
            r["unresolved"] = [i for i in unresolved if not known(i)]
            r["self_removed"] = [i for i in unresolved if i not in r["unresolved"]]

        # --- emptiness: a visible region with nothing in it -------------------
        # --- controls, drive, press ------------------------------------------
        drive = json.loads(self._eval(cdp, DRIVE_JS) or "{}")
        r.update({k: drive.get(k, r[k]) for k in ("controls", "changed", "pressed")})

        # --- emptiness, measured AFTER the interaction ------------------------
        # Order is the whole point. Measured BEFORE driving, this flagged seven
        # tools whose output boxes are SUPPOSED to start empty — you have not
        # given them anything yet. An output region still blank after the tool
        # has been driven and its action button pressed is a real finding; the
        # same region blank on arrival is just an empty form.
        #
        # Also exempt: anything painted rather than written. svg-patterns'
        # #preview carries a live background-image and no text at all, which is
        # exactly what that tool is for.
        r["empty_regions"] = json.loads(self._eval(cdp, """JSON.stringify(
            [...document.querySelectorAll('main [id], main [class]')]
              .filter(e => !/^(INPUT|TEXTAREA|SELECT|BUTTON|CANVAS|SVG|IMG|LABEL|IFRAME)$/.test(e.tagName))
              .filter(e => !e.isContentEditable && !e.getAttribute('placeholder'))
              // Painted, not written: exclude anything carrying a background
              // image OR a background colour its parent does not have. Without
              // the colour half this flagged vibe-equalizer's .card-img, a
              // 200px image placeholder that is decorative by design.
              .filter(e => { const s = getComputedStyle(e);
                             const p = e.parentElement ? getComputedStyle(e.parentElement) : null;
                             return s.backgroundImage === 'none' && s.visibility !== 'hidden'
                                    && (!p || s.backgroundColor === p.backgroundColor); })
              // childElementCount === 0 is load-bearing. Without it this flagged
              // clip-path, community-growth and shadow-stacker, whose regions hold
              // a styled child carrying no text — a clip-path polygon, twelve chart
              // bars, a live box-shadow. "Empty" has to mean nothing is there, not
              // "nothing is WRITTEN there", or every visual tool on the site is a
              // false positive.
              .filter(e => e.childElementCount === 0)
              .filter(e => { const b = e.getBoundingClientRect();
                             return b.width > 250 && b.height > 120
                                    && !e.innerText.trim()
                                    && !e.querySelector('input,select,textarea,button,canvas,svg,img,iframe,[contenteditable]'); })
              .map(e => (e.id ? '#'+e.id : '.'+String(e.className).split(' ')[0])).slice(0,4))""") or "[]")

        # --- responsive: the check that would have caught fluid-typography ---
        # Only for pages that actually claim viewport-relative sizing. A tool
        # emitting clamp(...vw...) is promising the visitor that something
        # scales; the only instrument that can test that is CHANGING THE
        # VIEWPORT. Reading the CSS cannot: a clamp can be present, correct, and
        # pinned at its maximum across every width a desktop uses, which is
        # exactly the state fluid-typography shipped in.
        r["responsive"] = self._responsive(cdp, bodies)

        cdp.call("Runtime.evaluate", {"expression": "1", "returnByValue": True}, timeout=10)
        return self._verdict(r, cdp)

    def _responsive(self, cdp, bodies):
        page = "\n".join(bodies)
        if "vw" not in page or "clamp(" not in page:
            return None  # makes no scaling claim; nothing to hold it to
        target = self._eval(cdp, """(() => {
            const el = [...document.querySelectorAll('main *')]
              .find(e => (e.getAttribute('style') || '').includes('vw'));
            if (!el) return null;
            if (!el.id) el.id = '__audit_scale_target';
            return el.id;
        })()""")
        if not target:
            return None
        # TWO BANDS, and the second is the one that matters. The full range
        # (360/900/1600) only proves a clamp is wired up at all — it detects
        # ANY correct clamp, including the one the owner complained about,
        # because 360 sits below min-width so the size always differs there.
        # It would NOT have caught the reported defect. The desktop band is
        # what the visitor actually inhabits: a preview constant across
        # 1280-1920 demonstrates nothing to anyone on a laptop, even though a
        # clamp pinned above its max-width is perfectly correct CSS. So the
        # desktop band is reported as DATA, not as a failure — the judgement
        # ("this tool exists to show scaling and shows none") belongs in the
        # tool's own criteria fence, not in a generic rule.
        sizes = []
        for w in (360, 900, 1600, 1280, 1920):
            cdp.call("Emulation.setDeviceMetricsOverride",
                     {"width": w, "height": 900, "deviceScaleFactor": 1, "mobile": False})
            time.sleep(0.4)
            sizes.append(self._eval(cdp, "getComputedStyle(document.getElementById(%s)).fontSize"
                                    % json.dumps(target)))
        cdp.call("Emulation.clearDeviceMetricsOverride")
        full, desktop = sizes[:3], sizes[3:]
        return {"target": target,
                "full_range": dict(zip((360, 900, 1600), full)),
                "desktop_band": dict(zip((1280, 1920), desktop)),
                "scales": len(set(full)) > 1,
                "scales_on_desktop": len(set(desktop)) > 1}

    def _verdict(self, r, cdp):
        if r["console"]:
            r["verdict"] = "BROKEN"; r["why"].append("throws: " + r["console"][0][:70])
        elif r["bad_subresources"]:
            r["verdict"] = "BROKEN"; r["why"].append("subresource: " + r["bad_subresources"][0][:70])
        elif r["unresolved"]:
            r["verdict"] = "BROKEN"; r["why"].append("script addresses absent: " + ", ".join(r["unresolved"][:4]))
        elif r["empty_regions"]:
            r["verdict"] = "EMPTY"; r["why"].append("blank region: " + ", ".join(r["empty_regions"]))
        elif r["controls"] == 0:
            r["verdict"] = "NO-CONTROL"; r["why"].append("nothing a visitor can touch")
        elif not r["changed"]:
            r["verdict"] = "DEAD"; r["why"].append("driven and pressed, nothing changed")
        elif r.get("responsive") and not r["responsive"]["scales"]:
            # Not BROKEN — the tool works, it just never shows what it promises.
            # fluid-typography's own class, and it needs its own word: calling
            # it RESPONDS is what let it sit unnoticed.
            r["verdict"] = "NO-SCALING"
            r["why"].append("claims viewport scaling; %s identical at 360/900/1600px"
                            % r["responsive"]["target"])
        else:
            r["verdict"] = "RESPONDS"
        return r

    _served_cache = {}

    def _served_html(self, url):
        """The bytes the server sent, before any script ran. Cached per URL."""
        if url not in self._served_cache:
            try:
                out = subprocess.run(["curl", "-s", url], capture_output=True, text=True, timeout=40)
                self._served_cache[url] = out.stdout
            except Exception:
                self._served_cache[url] = ""
        return self._served_cache[url]

    def _eval(self, cdp, expr):
        try:
            res = cdp.call("Runtime.evaluate",
                           {"expression": expr, "returnByValue": True, "awaitPromise": True},
                           timeout=25)
            return res.get("result", {}).get("result", {}).get("value")
        except Exception:
            return None


DRIVE_JS = r"""
(() => {
  const NOOP = /undo|redo|reset|clear|copy|download|share|print|remove|delete|back|default/i;
  const ACTION = /generate|run|convert|build|calculat|format|minif|compil|analys|analyz|check|creat|make|render|split|extract|optimi|inject|apply/i;
  const label = e => ((e.id||'') + ' ' + (e.className||'') + ' ' + (e.innerText||'') + ' ' +
                      (e.getAttribute('aria-label')||'') + ' ' + (e.placeholder||''));
  const all = [...document.querySelectorAll('main input, main select, main textarea, main button, main [contenteditable]')];
  const usable = all.filter(e => !e.disabled && !NOOP.test(label(e)));
  const snap = () => {
    const outs = [...document.querySelectorAll('main pre, main code, main canvas, main output, main [id*=output], main [id*=result], main [id*=preview], main .preview, main #canvas')];
    return outs.map(o => o.tagName === 'CANVAS' ? (()=>{try{return o.toDataURL().length}catch(e){return 'x'}})()
                                                : (o.innerHTML||'') + '|' + (o.value||'')).join('~~') +
           '##' + all.map(e => e.value ?? '').join('|');
  };
  // Preference order, tried IN TURN — a comma selector returns document order,
  // which is how three working tools were once scored dead.
  const pick = sel => usable.find(e => e.matches(sel));
  const target = pick('input[type=range]') || pick('input[type=number]') ||
                 pick('input[type=text]') || pick('textarea') || pick('select') ||
                 pick('input[type=color]') || pick('[contenteditable]');
  const before = snap();
  let drove = null;
  if (target) {
    drove = target.id || target.tagName;
    const t = (target.type || '').toLowerCase();
    // A VALID value, inferred from the field — typing 'probe123' into a colour
    // input scores correct input-validation as breakage.
    if (t === 'range' || t === 'number') {
      const n = parseFloat(target.value || '0');
      const mn = parseFloat(target.min || '0'), mx = parseFloat(target.max || (n + 10));
      target.value = String(n === mx ? mn : Math.min(mx, n + Math.max(1, (mx - mn) / 4)));
    } else if (t === 'color' || /^#[0-9a-f]{3,8}$/i.test(target.value || '')) {
      target.value = (target.value || '').toLowerCase() === '#3366cc' ? '#cc6633' : '#3366cc';
    } else if (target.tagName === 'SELECT') {
      if (target.options.length > 1) target.selectedIndex = (target.selectedIndex + 1) % target.options.length;
    } else if (/^\s*[\{\[]/.test(target.value || '')) {
      target.value = '{"probe": 1}';
    } else if (/</.test(target.value || '')) {
      target.value = '<p>probe</p>';
    } else if (target.isContentEditable) {
      target.innerText = 'Probe text';
    } else {
      target.value = 'probe';
    }
    ['input', 'change', 'keyup'].forEach(ev => target.dispatchEvent(new Event(ev, {bubbles: true})));
  }
  let changed = snap() !== before, pressed = null;
  if (!changed) {
    // TRY EACH BUTTON IN TURN, not just the first. Pressing one button and
    // concluding DEAD cannot tell an already-selected option from a dead
    // control: clip-path's first button is Triangle and the shape shown on
    // arrival IS a triangle, so the correct no-op scored the tool dead. Same
    // family as cubic-bezier's Default preset, except a shape name cannot be
    // recognised by a regex — only by trying the next one.
    const buttons = usable.filter(e => e.tagName === 'BUTTON')
                          .sort((a, b) => (ACTION.test(label(b)) ? 1 : 0) - (ACTION.test(label(a)) ? 1 : 0));
    for (const btn of buttons.slice(0, 6)) {
      btn.click();
      pressed = btn.innerText.trim().slice(0, 30);
      if (snap() !== before) { changed = true; break; }
    }
  }
  return new Promise(res => setTimeout(() => res(JSON.stringify({
    controls: usable.length, changed: changed || snap() !== before,
    pressed, drove
  })), 1400));
})()
"""


def main():
    args = [a for a in sys.argv[1:]]
    out_json = None
    if "--json" in args:
        i = args.index("--json"); out_json = args[i + 1]; del args[i:i + 2]
    urls = tool_urls() if "--all" in args else args
    a = Auditor()
    results = []
    try:
        for u in urls:
            try:
                r = a.audit(u)
            except Exception as e:
                # A harness failure must never print in the same column as a
                # site verdict. This is the fault that once mislabelled two
                # working tools BROKEN.
                r = {"url": u, "verdict": "HARNESS-ERROR", "why": [str(e)[:90]]}
            results.append(r)
            print("%-11s %-46s %s" % (r["verdict"], u.replace(DOMAIN + "/tools/", "").replace("/index.html", ""),
                                      "; ".join(r.get("why", []))[:95]))
            sys.stdout.flush()
    finally:
        if a.chrome:
            a.chrome.terminate()
    if out_json:
        json.dump(results, open(out_json, "w"), indent=1)
    counts = {}
    for r in results:
        counts[r["verdict"]] = counts.get(r["verdict"], 0) + 1
    print("\n" + "  ".join("%s=%d" % kv for kv in sorted(counts.items())))


main()
