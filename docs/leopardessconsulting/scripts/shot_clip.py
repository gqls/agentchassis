import base64, subprocess, sys, time
sys.path.insert(0, "/home/ant/projects/agentchassis/docs/leopardessconsulting/scripts")
from cdp import WS, http_json, evaluate
URL, OUT, W, SEL = sys.argv[1], sys.argv[2], int(sys.argv[3]), sys.argv[4]
PORT = 9361
p = subprocess.Popen(["/snap/bin/chromium","--headless=new",f"--remote-debugging-port={PORT}",
  "--no-sandbox","--disable-gpu","--hide-scrollbars","--force-prefers-reduced-motion",
  "--user-data-dir=/tmp/cdp-clip","about:blank"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
try:
    for _ in range(60):
        try: http_json(PORT,"/json/version"); break
        except Exception: time.sleep(0.5)
    ws = WS([t for t in http_json(PORT,"/json/list") if t["type"]=="page"][0]["webSocketDebuggerUrl"])
    ws.call("Runtime.enable"); ws.call("Page.enable")
    ws.call("Emulation.setDeviceMetricsOverride",{"width":W,"height":900,"deviceScaleFactor":2,"mobile":W<600})
    ws.call("Page.navigate",{"url":URL+"?cb="+str(int(time.time()))})
    for _ in range(80):
        if evaluate(ws,"document.readyState")=="complete": break
        time.sleep(0.4)
    time.sleep(1.5)
    box = evaluate(ws, f"""(() => {{ const e=document.querySelector('{SEL}'); if(!e) return null;
        const r=e.getBoundingClientRect(); return {{x:r.x+scrollX,y:r.y+scrollY,w:r.width,h:r.height}}; }})()""")
    if not box: sys.exit(f"selector not found: {SEL}")
    r = ws.call("Page.captureScreenshot", {"format":"png","captureBeyondViewport":True,
        "clip":{"x":box["x"],"y":box["y"],"width":box["w"],"height":min(box["h"],8000),"scale":1}})
    open(OUT,"wb").write(base64.b64decode(r["data"]))
    print(f"{OUT}  {int(box['w'])}x{int(box['h'])}  ({SEL})")
finally: p.terminate()
