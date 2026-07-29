#!/usr/bin/env python3
"""Evaluate a JS expression on a live page. evalpage.py <url> <expr>"""
import json
import os
import sys
import tempfile
import time
import urllib.request

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from toolprobe import CDP, start_chrome  # noqa: E402

url, expr = sys.argv[1], sys.argv[2]
port = 9355
chrome = start_chrome(port, tempfile.mkdtemp(prefix="evalpage-"))
try:
    req = urllib.request.Request(f"http://127.0.0.1:{port}/json/new?about:blank", method="PUT")
    tab = json.loads(urllib.request.urlopen(req, timeout=15).read())
    cdp = CDP(tab["webSocketDebuggerUrl"])
    cdp.call("Runtime.enable")
    cdp.call("Log.enable")
    cdp.call("Page.enable")
    cdp.call("Page.navigate", {"url": url})
    errs = []
    end = time.time() + 8
    while time.time() < end:
        try:
            m = cdp._recv_frame(timeout=1.0)
        except Exception:
            continue
        if m.get("method") == "Runtime.exceptionThrown":
            d = m["params"]["exceptionDetails"]
            errs.append((d.get("exception", {}).get("description") or d.get("text", "")).split("\n")[0])
        elif m.get("method") == "Log.entryAdded":
            e = m["params"]["entry"]
            if e.get("level") == "error":
                errs.append("LOG: " + (e.get("text") or "")[:200])
        elif m.get("method") == "Page.loadEventFired":
            end = min(end, time.time() + 2)
    r = cdp.call("Runtime.evaluate", {"expression": expr, "returnByValue": True, "awaitPromise": True}, timeout=25)
    res = r.get("result", {}).get("result", {})
    print("VALUE:", json.dumps(res.get("value"), indent=1)[:3000] if "value" in res else res)
    if r.get("result", {}).get("exceptionDetails"):
        print("EXCEPTION:", r["result"]["exceptionDetails"].get("text"))
    for e in errs:
        print("ERROR:", e[:250])
finally:
    chrome.terminate()
