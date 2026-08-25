"""Does the stat-band count-up actually run?

A read-after-scroll races the tween: with CDP round-trip latency a short animation is
already finished. So install a MutationObserver BEFORE the band is scrolled into view and
count the mutations it records — a tween that ran leaves a trail whatever its duration.
"""
import json, subprocess, sys, time
sys.path.insert(0,"/home/ant/projects/agentchassis/docs/leopardessconsulting/scripts")
from cdp import WS, http_json, evaluate

def run(reduced):
    PORT = 9401 if not reduced else 9402
    args=["/snap/bin/chromium","--headless=new",f"--remote-debugging-port={PORT}","--no-sandbox",
          "--disable-gpu","--hide-scrollbars",f"--user-data-dir=/tmp/cdp-c2{PORT}","about:blank"]
    if reduced: args.insert(6,"--force-prefers-reduced-motion")
    p=subprocess.Popen(args,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
    try:
        for _ in range(60):
            try: http_json(PORT,"/json/version"); break
            except Exception: time.sleep(0.5)
        ws=WS([t for t in http_json(PORT,"/json/list") if t["type"]=="page"][0]["webSocketDebuggerUrl"])
        ws.call("Runtime.enable"); ws.call("Page.enable")
        ws.call("Emulation.setDeviceMetricsOverride",{"width":1280,"height":900,"deviceScaleFactor":1,"mobile":False})
        ws.call("Page.navigate",{"url":"https://leopardessconsulting.co.uk/index.html?cb="+str(int(time.time()))+str(PORT)})
        for _ in range(80):
            if evaluate(ws,"document.readyState")=="complete": break
            time.sleep(0.4)
        time.sleep(0.4)
        # arm the observer while the band is still off-screen
        armed = evaluate(ws,"""(() => {
          const el = document.querySelector('.stat-band__value');
          if (!el) return {armed:false};
          window.__trail = [el.textContent];
          new MutationObserver(() => window.__trail.push(el.textContent))
              .observe(el, {childList:true, characterData:true, subtree:true});
          const r = el.getBoundingClientRect();
          return {armed:true, offscreen: r.top > window.innerHeight,
                  init_ran_marker: el.hasAttribute('data-final') || el.dataset.final !== undefined,
                  snippet_fn: typeof window.__statBandInit};
        })()""")
        evaluate(ws,"document.querySelector('.stat-band').scrollIntoView({behavior:'auto',block:'center'})")
        time.sleep(3.0)
        trail = evaluate(ws,"window.__trail")
        return {"reduced_motion":reduced, **armed,
                "mutations": len(trail)-1, "trail_head": trail[:6], "final": trail[-1]}
    finally: p.terminate()

out=[run(False), run(True)]
print(json.dumps(out, indent=2))
print()
print("NORMAL   animates (mutations > 0):", out[0]["mutations"] > 0)
print("REDUCED  static   (mutations = 0):", out[1]["mutations"] == 0)
print("final value correct in both      :", out[0]["final"].strip()=="22" and out[1]["final"].strip()=="22")
