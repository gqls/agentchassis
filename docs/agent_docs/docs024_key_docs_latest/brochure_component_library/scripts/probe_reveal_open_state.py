#!/usr/bin/env python3
"""Measure the teaser panel's OPEN state in a real browser.

Why this is separate from render_audit.py: that tool renders a LOCAL copy of the
page, so a query string like ?open=review-council never reaches window.location
and the "open state audit" silently measures the CLOSED page. Same number, wrong
question. This probe forces every <details> open in the DOM and reads the
computed colours of the revealed body text.
"""
import glob, html as htmllib, json, os, re, subprocess, sys, tempfile, urllib.request

URL = sys.argv[1] if len(sys.argv) > 1 else "https://fundamentallyai.com/index.html"

PROBE = r"""
<script>
(function(){
  function lum(c){
    var m = c.match(/rgba?\(([^)]+)\)/); if(!m) return null;
    var p = m[1].split(',').map(function(x){return parseFloat(x)});
    var f = p.slice(0,3).map(function(v){ v/=255; return v<=0.03928 ? v/12.92 : Math.pow((v+0.055)/1.055,2.4); });
    return 0.2126*f[0]+0.7152*f[1]+0.0722*f[2];
  }
  function bg(el){
    while(el){ var c = getComputedStyle(el).backgroundColor;
      if(c && !/rgba\(0, 0, 0, 0\)|transparent/.test(c)) return c; el = el.parentElement; }
    return 'rgb(255, 255, 255)';
  }
  var out = {opened:0, measured:[], failures:[]};
  document.querySelectorAll('details.trp__card').forEach(function(d){ d.open = true; out.opened++; });
  document.querySelectorAll('.trp__body, .trp__body p, .trp__hook, .trp__continuation, .trp__control').forEach(function(el){
    var t = (el.textContent||'').trim(); if(t.length < 8) return;
    var cs = getComputedStyle(el);
    if(cs.display === 'none' || cs.visibility === 'hidden') return;
    var l1 = lum(cs.color), l2 = lum(bg(el));
    if(l1===null||l2===null) return;
    var ratio = (Math.max(l1,l2)+0.05)/(Math.min(l1,l2)+0.05);
    var rec = {cls: el.className, ratio: Math.round(ratio*100)/100, text: t.slice(0,45)};
    out.measured.push(rec);
    var big = parseFloat(cs.fontSize) >= 24 || (parseFloat(cs.fontSize) >= 18.66 && parseInt(cs.fontWeight,10) >= 700);
    if (ratio < (big ? 3.0 : 4.5)) out.failures.push(rec);
  });
  var pre = document.createElement('pre'); pre.id = 'PROBE_RESULT';
  pre.textContent = JSON.stringify(out); document.body.appendChild(pre);
})();
</script>
"""

cands = [os.environ.get("CHROME")] + sorted(glob.glob(os.path.expanduser(
    "~/.cache/ms-playwright/chromium-*/chrome-linux64/chrome")), reverse=True) + \
    ["/usr/bin/chromium", "/usr/bin/chromium-browser", "/usr/bin/google-chrome"]
chrome = next(c for c in cands if c and os.path.exists(c))

req = urllib.request.Request(URL, headers={"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36"})
raw = urllib.request.urlopen(req, timeout=60).read().decode("utf-8", "replace")
base = "/".join(URL.split("/")[:3]) + "/"
if "<base " not in raw:
    raw = raw.replace("<head>", '<head><base href="%s">' % base, 1)
raw = raw.replace("</body>", PROBE + "</body>", 1)

with tempfile.TemporaryDirectory() as wd:
    p = os.path.join(wd, "probe.html")
    open(p, "w").write(raw)
    r = subprocess.run([chrome, "--headless=new", "--disable-gpu", "--no-sandbox",
                        "--hide-scrollbars", "--window-size=1280,900",
                        "--virtual-time-budget=10000", "--dump-dom", "file://" + p],
                       capture_output=True, text=True, timeout=180)
    m = re.search(r'<pre id="PROBE_RESULT">(.*?)</pre>', r.stdout, re.S)
    if not m:
        sys.exit("probe produced no result")
    d = json.loads(htmllib.unescape(m.group(1)))

print("details forced open :", d["opened"])
print("elements measured   :", len(d["measured"]))
body = [x for x in d["measured"] if "trp__body" in x["cls"]]
print("revealed body texts :", len(body))
for x in body:
    print("   %5.2f:1  %s" % (x["ratio"], x["text"]))
print("FAILURES            :", len(d["failures"]))
for f in d["failures"]:
    print("   %5.2f:1  %s  | %s" % (f["ratio"], f["cls"], f["text"]))
sys.exit(1 if d["failures"] or not body else 0)
