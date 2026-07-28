#!/usr/bin/env python3
"""Render a live page in headless Chromium and measure what a visitor actually sees.

SUPERSEDED 2026-07-28 by internal/adapters/browserrunner/render_audit_action.go,
a faithful Go port carrying the same probe. KEPT FOR NOW, deliberately: the Go
version lives in the browser-runner-adapter and is inert until that image is
rebuilt and rolled. Deleting this before then would leave the fleet with no
render audit at all, and this is currently the only thing that catches
features_open/026 family 3 (a component hard-coding an ink over a themed fill).

RETIRE THIS FILE once `render_audit` is confirmed live on the adapter — a pod
that answers the action, and one real run. Until then prefer this script and
expect the two to agree; if they disagree, the Go port is wrong.

WHY THIS EXISTS
---------------
Every check the build pipeline runs today reads a SOURCE: a component template,
a palette row, a token vocabulary, a link href. None of them renders the page.
That gap is not incidental — three defect families are invisible to source-side
checking by construction, because each is a property of the COMPOSITION and not
of any input:

  1. a palette slot the layout supplies a literal for (a #ffffff card_bg on a
     dark site) — every input is individually valid;
  2. a token used in two roles (--color-primary as both a fill and a
     foreground) — correct in one place, invisible in the other;
  3. a component hard-coding an ink over a themed background — the template
     reads fine, the pairing is what fails.

On fundamentallyai.com on 2026-07-27 those three came to 101 WCAG-AA failures
across 5 pages, including every card title on the site, and nothing in the
platform had flagged any of it. This script found all 101 in about two minutes.

It also reports images that failed to load, which the DB-only image_url_404
check cannot see: that check compares an /assets/images/ reference against the
`assets` table and its own header says the HTTP half is deferred. Five broken
images on this site each had a perfectly good assets row.

WHAT IT MEASURES
----------------
For every visible text node: the computed colour, the effective background
(walking up through transparent ancestors and compositing alpha), and the WCAG
contrast ratio between them, against 4.5:1 for body text and 3.0:1 for large
text — the same thresholds the renderer now uses.

USAGE
    scripts/render_audit.py https://example.com/index.html [more urls...]
    scripts/render_audit.py --sitemap https://example.com      (crawls nav links)
    scripts/render_audit.py --json out.json https://example.com/index.html

Exit status is 1 if any page has a failure, so it works as a gate.

REQUIREMENTS
    A Chromium binary. Set CHROME to override; otherwise the first of
    $CHROME, the Playwright cache, /usr/bin/chromium, /usr/bin/google-chrome.
"""

import argparse
import glob
import html as htmllib
import json
import os
import re
import subprocess
import sys
import tempfile
import time
import urllib.parse
import urllib.error
import urllib.request

# Some origins (Cloudflare in front of several of our own sites) answer 403 to
# the default python-urllib agent. The probe is meant to see what a browser
# sees, so it asks as one.
UA = ("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) "
      "Chrome/120.0 Safari/537.36")


def fetch(url, timeout=30, attempts=3):
    """Retry before condemning: a burst of probes throttles our own origins, and
    a reset connection then reads exactly like a broken page."""
    last = None
    for i in range(attempts):
        try:
            req = urllib.request.Request(url, headers={
                "User-Agent": UA, "Accept": "text/html,application/xhtml+xml"})
            with urllib.request.urlopen(req, timeout=timeout) as r:
                return r.read().decode("utf-8", "replace")
        except Exception as e:                                # noqa: BLE001
            last = e
            time.sleep(1.5 * (i + 1))
    raise last

AUDIT_JS = r"""
(function () {
  function parseRGB(s){var m=s.match(/rgba?\(([^)]+)\)/);if(!m)return null;
    var p=m[1].split(',').map(function(x){return parseFloat(x.trim())});
    return {r:p[0],g:p[1],b:p[2],a:p.length>3?p[3]:1};}
  function lum(c){function f(v){v=v/255;return v<=0.03928?v/12.92:Math.pow((v+0.055)/1.055,2.4)}
    return 0.2126*f(c.r)+0.7152*f(c.g)+0.0722*f(c.b);}
  function ratio(a,b){var l1=lum(a),l2=lum(b);
    return (Math.max(l1,l2)+0.05)/(Math.min(l1,l2)+0.05);}
  function over(fg,bg){if(fg.a>=1)return fg;
    return {r:fg.r*fg.a+bg.r*(1-fg.a),g:fg.g*fg.a+bg.g*(1-fg.a),b:fg.b*fg.a+bg.b*(1-fg.a),a:1};}
  function effBG(el){
    var stack=[],node=el,anyImg=false;
    while(node&&node.nodeType===1){
      var cs=getComputedStyle(node),c=parseRGB(cs.backgroundColor);
      var hasImg=cs.backgroundImage&&cs.backgroundImage!=='none';
      if(c&&c.a>0)stack.push({c:c,img:hasImg});
      // A background IMAGE under the text: its colour is unknown, so assume a
      // mid grey. Reported as overImage so a reader can discount it rather
      // than trusting a number the page cannot actually justify.
      if(hasImg&&(!c||c.a<1))stack.push({c:{r:128,g:128,b:128,a:1},img:true});
      if(c&&c.a>=1)break;
      node=node.parentElement;
    }
    var base={r:255,g:255,b:255,a:1};
    for(var i=stack.length-1;i>=0;i--){if(stack[i].img)anyImg=true;base=over(stack[i].c,base);}
    return {bg:base,overImage:anyImg};
  }
  var out={url:location.href,contrast:[],images:[],overflow:null},seen={};
  var all=document.querySelectorAll('body *');
  for(var i=0;i<all.length;i++){
    var el=all[i],cs=getComputedStyle(el);
    if(cs.display==='none'||cs.visibility==='hidden'||parseFloat(cs.opacity)===0)continue;
    var r0=el.getBoundingClientRect();
    if(r0.width===0||r0.height===0)continue;
    var txt='';
    for(var n=0;n<el.childNodes.length;n++)
      if(el.childNodes[n].nodeType===3)txt+=el.childNodes[n].nodeValue;
    txt=txt.replace(/\s+/g,' ').trim();
    if(txt.length<2)continue;
    var fg=parseRGB(cs.color);if(!fg)continue;
    var eb=effBG(el),fgc=over(fg,eb.bg),r=ratio(fgc,eb.bg);
    var size=parseFloat(cs.fontSize),weight=parseInt(cs.fontWeight,10)||400;
    var large=size>=24||(size>=18.66&&weight>=700),need=large?3.0:4.5;
    if(r>=need)continue;
    var cls=(typeof el.className==='string'?el.className:'')||el.tagName;
    var key=cls+'|'+cs.color+'|'+txt.slice(0,40);
    if(seen[key])continue; seen[key]=1;
    out.contrast.push({cls:cls,tag:el.tagName,text:txt.slice(0,80),fg:cs.color,
      bg:'rgb('+Math.round(eb.bg.r)+','+Math.round(eb.bg.g)+','+Math.round(eb.bg.b)+')',
      ratio:Math.round(r*100)/100,need:need,overImage:eb.overImage,px:Math.round(size)});
  }
  for(var j=0;j<document.images.length;j++){
    var im=document.images[j];
    if(!im.complete||im.naturalWidth===0)
      out.images.push({src:im.getAttribute('src'),alt:(im.alt||'').slice(0,80)});
  }
  if(document.documentElement.scrollWidth>window.innerWidth+1)
    out.overflow={scrollWidth:document.documentElement.scrollWidth,viewport:window.innerWidth};
  var pre=document.createElement('pre');
  pre.id='AUDIT_RESULT';pre.textContent=JSON.stringify(out);
  document.body.appendChild(pre);
})();
"""


def find_chrome():
    cands = [os.environ.get("CHROME")]
    cands += sorted(glob.glob(os.path.expanduser(
        "~/.cache/ms-playwright/chromium-*/chrome-linux64/chrome")), reverse=True)
    cands += ["/usr/bin/chromium", "/usr/bin/chromium-browser", "/usr/bin/google-chrome"]
    for c in cands:
        if c and os.path.exists(c):
            return c
    sys.exit("no chromium found; set CHROME=/path/to/chrome")


def audit(url, chrome, width, height, workdir):
    """Fetch the page, inline the probe, render it, read the verdict back.

    The probe is injected into a LOCAL copy rather than run against the origin
    because headless Chrome's CLI has no evaluate-on-load hook. <base href> keeps
    every relative asset resolving against the real site, so what is measured is
    the live stylesheet and the live images — not a local approximation.
    """
    try:
        page = fetch(url)
    except Exception as e:                                    # noqa: BLE001
        return {"url": url, "error": "fetch failed: %s" % e}

    base = urllib.parse.urljoin(url, ".")
    injected = re.sub(r"<head([^>]*)>", r"<head\1><base href='%s'>" % base, page, count=1)
    if "<base" not in injected:
        injected = "<base href='%s'>" % base + injected
    injected = injected.replace("</body>", "<script>%s</script></body>" % AUDIT_JS)
    if "AUDIT_RESULT" not in injected and "</body>" not in page:
        injected += "<script>%s</script>" % AUDIT_JS

    path = os.path.join(workdir, re.sub(r"\W+", "_", url)[-80:] + ".html")
    with open(path, "w") as f:
        f.write(injected)

    proc = subprocess.run(
        [chrome, "--headless=new", "--disable-gpu", "--no-sandbox", "--hide-scrollbars",
         "--window-size=%d,%d" % (width, height), "--virtual-time-budget=10000",
         "--dump-dom", "file://" + path],
        capture_output=True, text=True, timeout=180)
    m = re.search(r'<pre id="AUDIT_RESULT">(.*?)</pre>', proc.stdout, re.S)
    if not m:
        return {"url": url, "error": "probe produced no result"}
    data = json.loads(htmllib.unescape(m.group(1)))
    data["url"] = url
    data["images_reported"] = len(data.get("images", []))
    data["images"] = verify_broken(data.get("images", []), url)
    return data


def verify_broken(images, page_url):
    """Re-check, serially, every image the browser reported as failed.

    A headless render fires every image request at once, and our own origins
    throttle a burst — so the browser's "this did not load" is evidence of a
    LOAD failure, not of a missing file. Measured 2026-07-27: 41 images
    reported broken across 7 pages, of which 35 served 200 on an unhurried
    re-check. Reporting those would have sent someone regenerating assets that
    were already there. Only a status that survives this pass is reported.
    """
    confirmed = []
    for im in images:
        src = im.get("src") or ""
        if not src:
            continue
        url = urllib.parse.urljoin(page_url, src)
        status = None
        for attempt in range(3):
            try:
                req = urllib.request.Request(url, headers={"User-Agent": UA})
                with urllib.request.urlopen(req, timeout=20) as r:
                    status = r.status
                break
            except urllib.error.HTTPError as e:
                status = e.code
                break
            except Exception:                                 # noqa: BLE001
                time.sleep(1.0 * (attempt + 1))
        if status is None or status >= 400:
            im["http_status"] = status if status is not None else "unreachable"
            confirmed.append(im)
        time.sleep(0.25)
    return confirmed


def discover(root):
    """Follow the nav links of the root page — the pages a visitor can reach."""
    try:
        page = fetch(root)
    except Exception as e:                                    # noqa: BLE001
        sys.exit("could not fetch %s: %s" % (root, e))
    urls, seen = [], set()
    for href in re.findall(r'href="([^"]+)"', page):
        if href.startswith(("http", "mailto:", "tel:", "#")) or not href.endswith(".html"):
            continue
        u = urllib.parse.urljoin(root, href)
        if u not in seen:
            seen.add(u)
            urls.append(u)
    return urls or [root]


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("urls", nargs="+")
    ap.add_argument("--sitemap", action="store_true",
                    help="treat the first url as a root and audit every .html it links to")
    ap.add_argument("--width", type=int, default=390, help="viewport width (default: mobile)")
    ap.add_argument("--height", type=int, default=844)
    ap.add_argument("--json", metavar="FILE", help="write the full findings as JSON")
    ap.add_argument("--quiet", action="store_true", help="totals only")
    args = ap.parse_args()

    chrome = find_chrome()
    urls = discover(args.urls[0]) if args.sitemap else args.urls

    results, total_c, total_i = [], 0, 0
    with tempfile.TemporaryDirectory() as workdir:
        for u in urls:
            res = audit(u, chrome, args.width, args.height, workdir)
            results.append(res)
            if res.get("error"):
                print("  %-58s ERROR %s" % (u[-58:], res["error"]))
                continue
            c, i = len(res["contrast"]), len(res["images"])
            total_c += c
            total_i += i
            flag = "ok" if not (c or i or res.get("overflow")) else "FAIL"
            reported = res.get("images_reported", i)
            note = "h-overflow" if res.get("overflow") else ""
            if reported > i:
                note += " (%d slow-loading image(s) re-checked OK)" % (reported - i)
            print("%-4s %-56s contrast=%-4d broken-img=%-3d %s" % (
                flag, u[-56:], c, i, note))
            if args.quiet:
                continue
            for f in sorted(res["contrast"], key=lambda x: x["ratio"])[:8]:
                note = "  (over an image — ratio approximate)" if f["overImage"] else ""
                print("       %5.2f:1 need %.1f  %-22s on %-18s .%s%s" % (
                    f["ratio"], f["need"], f["fg"], f["bg"], f["cls"][:34], note))
                print("               %r" % f["text"][:70])
            for im in res["images"]:
                print("       BROKEN IMAGE  %s  -> HTTP %s   (alt: %s)" % (
                    im["src"], im.get("http_status"), im["alt"][:50]))

    print("\n%d page(s): %d contrast failure(s), %d broken image(s)" % (
        len(results), total_c, total_i))
    if args.json:
        with open(args.json, "w") as f:
            json.dump(results, f, indent=2)
        print("full findings: %s" % args.json)
    return 1 if (total_c or total_i) else 0


if __name__ == "__main__":
    sys.exit(main())
