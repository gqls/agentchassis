"""Run render_audit.py's OWN probe JS against the LIVE page over CDP.

Why not just run render_audit.py: its probe writes the fetched HTML to a temp
file and renders file://, and a snap-confined Chromium renders its own error
page instead — every page comes back "probe produced no result". This runs the
identical AUDIT_JS in a real page load of the real URL, so it also drops the
injected-copy fidelity gap (relative assets, CSP, the real <head>).
"""
import importlib.util, json, subprocess, sys, time
sys.path.insert(0, "/home/ant/projects/agentchassis/docs/leopardessconsulting/scripts")
from cdp import WS, http_json, evaluate

spec = importlib.util.spec_from_file_location(
    "ra", "/home/ant/projects/agentchassis/scripts/render_audit.py")
ra = importlib.util.module_from_spec(spec); spec.loader.exec_module(ra)
AUDIT_JS = ra.AUDIT_JS

URL, PORT = sys.argv[1], 9351
WIDTH = int(sys.argv[2]) if len(sys.argv) > 2 else 1280
proc = subprocess.Popen(
    ["/snap/bin/chromium", "--headless=new", f"--remote-debugging-port={PORT}",
     "--no-sandbox", "--disable-gpu", "--hide-scrollbars",
     "--force-prefers-reduced-motion", "--user-data-dir=/tmp/cdp-audit", "about:blank"],
    stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
try:
    for _ in range(60):
        try: http_json(PORT, "/json/version"); break
        except Exception: time.sleep(0.5)
    ws = WS([t for t in http_json(PORT, "/json/list") if t["type"] == "page"][0]["webSocketDebuggerUrl"])
    ws.call("Runtime.enable"); ws.call("Page.enable")
    ws.call("Emulation.setDeviceMetricsOverride",
            {"width": WIDTH, "height": 900, "deviceScaleFactor": 1, "mobile": WIDTH < 600})
    ws.call("Page.navigate", {"url": URL + ("&" if "?" in URL else "?") + "cb=" + str(int(time.time()))})
    for _ in range(80):
        if evaluate(ws, "document.readyState") == "complete": break
        time.sleep(0.4)
    time.sleep(1.5)
    evaluate(ws, AUDIT_JS)
    raw = evaluate(ws, "document.getElementById('AUDIT_RESULT') && document.getElementById('AUDIT_RESULT').textContent")
    if not raw:
        sys.exit("probe produced no result — AUDIT_JS did not write its block")
    d = json.loads(raw)
    print(f"url        {URL}  @{WIDTH}px")
    print(f"contrast   {len(d.get('contrast', []))} failure(s)")
    print(f"images     {len(d.get('images', []))} reported broken")
    print(f"overflow   {d.get('overflow')}")
    for f in sorted(d.get("contrast", []), key=lambda x: x["ratio"])[:12]:
        print(f"   {f['ratio']:5.2f}:1 need {f['need']:.1f}  {f['fg']:<20} on {f['bg']:<18} .{f['cls'][:38]}")
        print(f"           {f['text'][:80]!r}")
    for im in d.get("images", []):
        print(f"   BROKEN IMAGE {im['src']}  (alt: {im.get('alt','')[:50]})")
    open(sys.argv[3] if len(sys.argv) > 3 else "/dev/null", "w").write(json.dumps(d, indent=2))
finally:
    proc.terminate()
