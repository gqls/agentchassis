"""Screenshot a URL at one or more widths, full page, via the stdlib CDP driver."""
import base64, json, subprocess, sys, time
sys.path.insert(0, "/home/ant/projects/agentchassis/docs/leopardessconsulting/scripts")
from cdp import WS, http_json, evaluate

URL = sys.argv[1]
OUT = sys.argv[2]
WIDTHS = [int(w) for w in (sys.argv[3] if len(sys.argv) > 3 else "1280").split(",")]
PORT = 9345

proc = subprocess.Popen(
    ["/snap/bin/chromium", "--headless=new", f"--remote-debugging-port={PORT}",
     "--no-sandbox", "--disable-gpu", "--force-prefers-reduced-motion",
     "--hide-scrollbars", "--user-data-dir=/tmp/cdp-shot", "about:blank"],
    stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
try:
    for _ in range(60):
        try:
            http_json(PORT, "/json/version"); break
        except Exception:
            time.sleep(0.5)
    tabs = [t for t in http_json(PORT, "/json/list") if t.get("type") == "page"]
    ws = WS(tabs[0]["webSocketDebuggerUrl"])
    ws.call("Runtime.enable"); ws.call("Page.enable"); ws.call("Emulation.setDeviceMetricsOverride",
        {"width": WIDTHS[0], "height": 900, "deviceScaleFactor": 1, "mobile": False})
    for w in WIDTHS:
        ws.call("Emulation.setDeviceMetricsOverride",
                {"width": w, "height": 900, "deviceScaleFactor": 1, "mobile": w < 600})
        ws.call("Page.navigate", {"url": URL + ("&" if "?" in URL else "?") + "cb=" + str(int(time.time()))})
        for _ in range(80):
            if evaluate(ws, "document.readyState") == "complete":
                break
            time.sleep(0.4)
        time.sleep(1.2)
        h = evaluate(ws, "Math.min(document.body.scrollHeight, 20000)")
        r = ws.call("Page.captureScreenshot", {"format": "png", "captureBeyondViewport": True,
                     "clip": {"x": 0, "y": 0, "width": w, "height": h, "scale": 1}})
        path = f"{OUT}_{w}.png"
        open(path, "wb").write(base64.b64decode(r["data"]))
        print(f"{path}  {w}x{h}")
finally:
    proc.terminate()
