"""Ground the visual-design-auditor's five claims at the rendered artefact.

A subagent's report is another doc. Every claim below is re-asked of the LIVE
page with getComputedStyle, because four of the five are about COMPUTED values
and the auditor read CSS SOURCE — where `var(--x, #fallback)` literals are
present and never applied.
"""
import json, subprocess, sys, time
sys.path.insert(0, "/home/ant/projects/agentchassis/docs/leopardessconsulting/scripts")
from cdp import WS, http_json, evaluate

URL, PORT = "https://leopardessconsulting.co.uk/index.html", 9371
p = subprocess.Popen(["/snap/bin/chromium","--headless=new",f"--remote-debugging-port={PORT}",
  "--no-sandbox","--disable-gpu","--hide-scrollbars","--force-prefers-reduced-motion",
  "--user-data-dir=/tmp/cdp-verify","about:blank"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
try:
    for _ in range(60):
        try: http_json(PORT,"/json/version"); break
        except Exception: time.sleep(0.5)
    ws = WS([t for t in http_json(PORT,"/json/list") if t["type"]=="page"][0]["webSocketDebuggerUrl"])
    ws.call("Runtime.enable"); ws.call("Page.enable")

    def load(w):
        ws.call("Emulation.setDeviceMetricsOverride",
                {"width":w,"height":900,"deviceScaleFactor":1,"mobile":w<600})
        ws.call("Page.navigate",{"url":URL+"?cb="+str(int(time.time()))+str(w)})
        for _ in range(80):
            if evaluate(ws,"document.readyState")=="complete": break
            time.sleep(0.4)
        time.sleep(1.2)

    load(1280)
    out = evaluate(ws, """(() => {
      const cs = getComputedStyle(document.documentElement);
      const V = n => cs.getPropertyValue(n).trim();
      const body = getComputedStyle(document.body);
      const sb = document.querySelector('.stat-band') || document.querySelector('[class*="stat-band"]');
      const hero = document.querySelector('.hero') || document.querySelector('[class*="hero"]');
      const h1 = document.querySelector('h1');
      return {
        root_vars: {primary: V('--color-primary'), accent: V('--color-accent'),
                    secondary: V('--color-secondary'), header_bg: V('--color-header-bg'),
                    spacing_section: V('--spacing-section')},
        body_font: body.fontFamily,
        h1_font: h1 ? getComputedStyle(h1).fontFamily : null,
        statband_class: sb ? sb.className : null,
        statband_padding: sb ? getComputedStyle(sb).padding : null,
        statband_padTop: sb ? getComputedStyle(sb).paddingTop : null,
        hero_inline_style: hero ? (hero.getAttribute('style') || '') : null,
        h1_fontsize_1280: h1 ? getComputedStyle(h1).fontSize : null,
        // does anything animate the stat numbers?
        counter_attrs: [...document.querySelectorAll('[data-counter],[data-count],[data-count-to],[data-target]')].length,
        snippets_js: [...document.querySelectorAll('script[src]')].map(s=>s.getAttribute('src')).filter(s=>/snippet/.test(s)),
      };
    })()""")
    load(375)
    narrow = evaluate(ws, """(() => {
      const h1 = document.querySelector('h1');
      const hero = document.querySelector('.hero') || document.querySelector('[class*="hero"]');
      return {h1_fontsize_375: h1 ? getComputedStyle(h1).fontSize : null,
              doc_scrollW: document.documentElement.scrollWidth,
              doc_clientW: document.documentElement.clientWidth,
              hero_overflow: hero ? (hero.scrollWidth > hero.clientWidth) : null};
    })()""")
    out.update(narrow)
    print(json.dumps(out, indent=2))
finally:
    p.terminate()
