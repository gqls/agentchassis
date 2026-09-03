#!/usr/bin/env python3
# Lives in finetuning_uk_service/. Needs: /snap/bin/chromium and a venv with websocket-client
# (python3 -m venv v && v/bin/pip install websocket-client). Loads the LIVE page so the Origin is
# https://finetuning.uk (a file:// copy gets Origin null and the route answers 403). Snap Chromium
# cannot read /tmp (LANDMINES) - this driver never writes there.
"""Drive the LIVE playground page in headless Chromium over CDP: load it (so the Origin is
https://finetuning.uk), type a message into the widget, click Send, wait for the streamed reply,
and print what the transcript shows. Exit 0 only if an assistant bubble with text appeared."""
import json, subprocess, sys, time, urllib.request, websocket
PORT = 9333
URL = sys.argv[1] if len(sys.argv) > 1 else "https://finetuning.uk/playground.html"
MSG = sys.argv[2] if len(sys.argv) > 2 else "In one sentence, what is fine-tuning?"
chrome = subprocess.Popen(["/snap/bin/chromium", "--headless=new", "--disable-gpu", "--no-first-run", "--remote-allow-origins=*",
                           f"--remote-debugging-port={PORT}", "--window-size=1280,2000", "about:blank"],
                          stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
try:
    for _ in range(40):
        try:
            targets = json.load(urllib.request.urlopen(f"http://127.0.0.1:{PORT}/json"))
            page = next(t for t in targets if t["type"] == "page"); break
        except Exception: time.sleep(0.5)
    else: print("FAIL: chromium did not expose a page target"); sys.exit(2)
    ws = websocket.create_connection(page["webSocketDebuggerUrl"], timeout=60); mid = 0
    def call(method, **params):
        global mid; mid += 1
        ws.send(json.dumps({"id": mid, "method": method, "params": params}))
        while True:
            r = json.loads(ws.recv())
            if r.get("id") == mid: return r.get("result", r)
    def js(expr):
        r = call("Runtime.evaluate", expression=expr, returnByValue=True, awaitPromise=True)
        return r.get("result", {}).get("value")
    call("Page.enable"); call("Runtime.enable")
    call("Page.navigate", url=URL); time.sleep(6)
    print("title:", js("document.title"))
    print("widget form:", js("!!document.querySelector('.playground-chat-tool form, [class*=playground] form')"))
    sel = js("""(function(){var f=document.querySelector('.playground-chat-tool form, [class*=playground] form'); if(!f) return null;
      var i=f.querySelector('textarea, input[type=text], input:not([type])'); var b=f.querySelector('button[type=submit], button'); return {input: !!i, button: !!b, btnText: b && b.textContent.trim()}})()""")
    print("controls:", sel)
    if not sel or not sel.get("input"): print("FAIL: no input in the widget"); sys.exit(1)
    js("""(function(){var f=document.querySelector('.playground-chat-tool form, [class*=playground] form'); var i=f.querySelector('textarea, input[type=text], input:not([type])');
      var setter = Object.getOwnPropertyDescriptor(i.tagName==='TEXTAREA'?HTMLTextAreaElement.prototype:HTMLInputElement.prototype,'value').set; setter.call(i, %s);
      i.dispatchEvent(new Event('input',{bubbles:true})); var b=f.querySelector('button[type=submit], button'); b.click(); return true})()""" % json.dumps(MSG))
    got = None
    for t in range(40):
        time.sleep(1)
        got = js("""(function(){var root=document.querySelector('.playground-chat-tool, [class*=playground]'); if(!root) return null;
           var bubbles=[].slice.call(root.querySelectorAll('[class*=msg], [class*=bubble], [class*=message], li, p')).map(function(e){return e.textContent.trim()}).filter(Boolean);
           var status=(root.querySelector('[role=status], [aria-live]')||{}).textContent||''; return {n:bubbles.length, last:bubbles.slice(-3), status:status.trim().slice(0,160)}})()""")
        if got and got.get("last") and len(got["last"]) >= 2 and len(got["last"][-1]) > 40 and MSG not in got["last"][-1]: break
    print("transcript tail:", json.dumps(got, ensure_ascii=False)[:900])
    ok = bool(got and got.get("last") and len(got["last"][-1]) > 40 and MSG not in got["last"][-1])
    print("RESULT:", "PASS (assistant reply rendered in the page)" if ok else "FAIL (no assistant reply rendered)")
    sys.exit(0 if ok else 1)
finally:
    chrome.terminate()
