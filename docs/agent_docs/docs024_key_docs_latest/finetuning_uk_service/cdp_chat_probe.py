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
    print("widget form:", js("!!document.querySelector('form.playground-form, form[id*=playground]')"))
    sel = js("""(function(){var f=document.querySelector('form.playground-form, form[id*=playground]'); if(!f) return null;
      var i=f.querySelector('textarea, input[type=text], input:not([type])'); var b=f.querySelector('button[type=submit], button'); return {input: !!i, button: !!b, btnText: b && b.textContent.trim()}})()""")
    print("controls:", sel)
    if not sel or not sel.get("input"): print("FAIL: no input in the widget"); sys.exit(1)
    js("""(function(){var f=document.querySelector('form.playground-form, form[id*=playground]'); var i=f.querySelector('textarea, input[type=text], input:not([type])');
      var setter = Object.getOwnPropertyDescriptor(i.tagName==='TEXTAREA'?HTMLTextAreaElement.prototype:HTMLInputElement.prototype,'value').set; setter.call(i, %s);
      i.dispatchEvent(new Event('input',{bubbles:true})); var b=f.querySelector('button[type=submit], button'); b.click(); return true})()""" % json.dumps(MSG))
    got = None
    for t in range(40):
        time.sleep(1)
        got = js("""(function(){var tr=document.querySelector('[id$=-transcript], .playground-transcript'); if(!tr) return null;
           var bubbles=[].slice.call(tr.children).map(function(e){return e.textContent.trim()}).filter(Boolean);
           var err=(document.querySelector('[id$=-input-error]')||{}).textContent||''; var btn=document.querySelector('form[id*=playground] button[type=submit]');
           return {n:bubbles.length, last:bubbles.slice(-3), error:err.trim().slice(0,200), sendDisabled: btn?btn.disabled:null, text: tr.innerText.slice(-500)}})()""")
        if got and got.get("last") and len(got["last"]) >= 2 and len(got["last"][-1]) > 40 and MSG not in got["last"][-1]: break
    print("location after send:", js("location.href"))
    print("form still present:", js("!!document.querySelector('form.playground-form, form[id*=playground]')"))
    print("transcript tail:", json.dumps(got, ensure_ascii=False)[:900])
    print("transcript text tail:", (got or {}).get("text"))
    ok = bool(got and got.get("last") and len(got["last"][-1]) > 40 and MSG not in got["last"][-1])
    print("RESULT:", "PASS (assistant reply rendered in the page)" if ok else "FAIL (no assistant reply rendered)")
    sys.exit(0 if ok else 1)
finally:
    chrome.terminate()
