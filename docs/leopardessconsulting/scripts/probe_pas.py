"""Real-gesture probe for tool-process-automation-scorer's `submit-shows-error`.

The fence asserts: click .pas-submit -> #pas-error appears. This drives the LIVE
page with real clicks and reads getComputedStyle, then runs the OPPOSITE branch
(answer all nine questions, click again) as the discrimination control: a probe
that reports "visible" on both branches is measuring nothing.
"""
import json, subprocess, sys, time, urllib.request
sys.path.insert(0, __file__.rsplit("/", 1)[0])
from cdp import WS, http_json, evaluate

PORT = 9333
URL = "https://leopardessconsulting.co.uk/tools/process-automation-scorer/index.html?cb=" + str(int(time.time()))
proc = subprocess.Popen(
    ["/snap/bin/chromium", "--headless=new", f"--remote-debugging-port={PORT}",
     "--no-sandbox", "--disable-gpu", "--force-prefers-reduced-motion",
     "--window-size=1280,900", "--user-data-dir=/tmp/cdp-pas", "about:blank"],
    stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

try:
    for _ in range(60):
        try:
            http_json(PORT, "/json/version"); break
        except Exception:
            time.sleep(0.5)
    else:
        sys.exit("chromium never opened the devtools port")

    # /json/new needs PUT on current Chrome; attach to the existing blank tab
    # and navigate instead — same target, one fewer thing to get wrong.
    tabs = [t for t in http_json(PORT, "/json/list") if t.get("type") == "page"]
    if not tabs:
        sys.exit("no page target to attach to")
    ws = WS(tabs[0]["webSocketDebuggerUrl"])
    ws.call("Runtime.enable")
    ws.call("Page.enable")
    ws.call("Page.navigate", {"url": URL})

    for _ in range(80):
        try:
            if evaluate(ws, "document.readyState") == "complete" and \
               evaluate(ws, "!!document.querySelector('.pas-submit')"):
                break
        except Exception:
            pass
        time.sleep(0.5)
    time.sleep(1.0)

    def state():
        return evaluate(ws, """(() => {
          const e = document.getElementById('pas-error');
          const r = document.getElementById('results');
          const btn = document.querySelector('.pas-submit');
          return {
            errorExists: !!e,
            errorDisplay: e ? getComputedStyle(e).display : null,
            errorVisible: !!(e && getComputedStyle(e).display !== 'none' && e.offsetParent !== null),
            resultsShown: !!(r && r.classList.contains('show')),
            score: (document.getElementById('score-number')||{}).textContent || '',
            submitExists: !!btn,
            answered: ['freq','judgement','inputs','exceptions','errorcost','docs','volume','systems','time']
                        .filter(n => document.querySelector('input[name="'+n+'"]:checked')).length
          };
        })()""")

    out = {"url": URL, "before": state()}

    # ---- BRANCH A: nothing answered, real click on the real button ----------
    evaluate(ws, "document.querySelector('.pas-submit').click()")
    time.sleep(0.6)
    out["afterClickEmpty"] = state()

    # ---- BRANCH B: answer all nine with real clicks, click submit again -----
    evaluate(ws, """(() => {
      ['freq','judgement','inputs','exceptions','errorcost','docs','volume','systems','time']
        .forEach(n => { const i = document.querySelector('input[name="'+n+'"]'); if (i) i.click(); });
    })()""")
    time.sleep(0.4)
    evaluate(ws, "document.querySelector('.pas-submit').click()")
    time.sleep(0.8)
    out["afterClickComplete"] = state()

    print(json.dumps(out, indent=2))

    a, b = out["afterClickEmpty"], out["afterClickComplete"]
    verdict = {
      "fence_submit_shows_error": a["errorVisible"] is True,
      "control_opposite_branch_hides_it": b["errorVisible"] is False,
      "control_complete_run_scores": b["resultsShown"] is True and b["score"] != "",
      "probe_discriminates": a["errorVisible"] != b["errorVisible"],
    }
    print(json.dumps(verdict, indent=2))
    sys.exit(0 if all(verdict.values()) else 1)
finally:
    proc.terminate()
