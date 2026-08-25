import json, subprocess, sys, time
sys.path.insert(0, "/home/ant/projects/agentchassis/docs/leopardessconsulting/scripts")
from cdp import WS, http_json, evaluate
PORT=9381
p=subprocess.Popen(["/snap/bin/chromium","--headless=new",f"--remote-debugging-port={PORT}","--no-sandbox",
 "--disable-gpu","--hide-scrollbars","--force-prefers-reduced-motion","--user-data-dir=/tmp/cdp-car","about:blank"],
 stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
try:
    for _ in range(60):
        try: http_json(PORT,"/json/version"); break
        except Exception: time.sleep(0.5)
    ws=WS([t for t in http_json(PORT,"/json/list") if t["type"]=="page"][0]["webSocketDebuggerUrl"])
    ws.call("Runtime.enable"); ws.call("Page.enable")
    ws.call("Emulation.setDeviceMetricsOverride",{"width":1280,"height":900,"deviceScaleFactor":1,"mobile":False})
    ws.call("Page.navigate",{"url":"https://leopardessconsulting.co.uk/services.html?cb="+str(int(time.time()))})
    for _ in range(80):
        if evaluate(ws,"document.readyState")=="complete": break
        time.sleep(0.4)
    time.sleep(1.5)
    print(json.dumps(evaluate(ws,"""(() => {
      const sec=document.querySelector('.info-card-grid-section');
      const track=document.querySelector('[data-hcc-track]');
      const arrow=document.querySelector('[data-hcc-next]');
      const cs=sec?getComputedStyle(sec):null;
      return {
        section_found: !!sec,
        icg_track_gap_custom_prop: cs?cs.getPropertyValue('--icg-track-gap').trim():null,
        icg_arrow_size_custom_prop: cs?cs.getPropertyValue('--icg-arrow-size').trim():null,
        track_computed_gap: track?getComputedStyle(track).gap:null,
        track_column_gap: track?getComputedStyle(track).columnGap:null,
        arrow_w: arrow?getComputedStyle(arrow).inlineSize:null,
        arrow_h: arrow?getComputedStyle(arrow).blockSize:null,
        arrow_rect: arrow?(r=>({w:Math.round(r.width),h:Math.round(r.height)}))(arrow.getBoundingClientRect()):null,
        slides: document.querySelectorAll('[data-hcc-slide]').length
      };})()"""), indent=2))
finally: p.terminate()
