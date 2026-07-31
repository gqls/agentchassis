#!/usr/bin/env python3
"""Look at a live page, or at one element of it, without guessing.

    look.py <url>
    look.py <url> --selector '.trp'
    look.py <url> --selector '.trp' --width 1400 --pad 40
    look.py <url> --profile mobile
    look.py <url> --selector '.trp' --out ~/carousel.png

Why this exists. Nothing in the framework renders a page and puts it in front of
anyone unless a declared check has already FAILED (browserrunner captures
screenshots as failure evidence; `capture_renders` — TL-035 — makes a passing
render possible but needs a caller with a criteria fence, which a brochure page
does not have). So looking at a page is still a manual act, and on 2026-07-30/31
three defects reached the owner that no assertion covered. This script is the
manual act, done properly, so the next person spends their time looking rather
than fighting the tooling.

Three traps it encodes, each paid for the hard way:

1. **snap chromium cannot read or write outside $HOME.** A temp file in /tmp gives
   `ERR_FILE_NOT_FOUND` and a screenshot written to /tmp silently does not appear.
   Everything here stays under $HOME.

2. **Do NOT try to render the whole document, and do not trust a document height.**
   Page height is not a fixed number: sections sized in `vh` grow with the viewport,
   so "measure the document, then render that tall" DIVERGES — measured on one page
   as 1000 -> 2854 -> 4152 -> 6141 -> 6453, never settling. Anything that screenshots
   at a height taken from an earlier pass also silently crops the bottom off the
   image. This scrolls the element to the top of a NORMAL viewport instead, which is
   bounded, convergent, and also what a visitor actually sees.

3. **A `file://` copy does not run its cross-origin scripts, but an
   `http://127.0.0.1` copy DOES.** That is the whole trick here. To measure an
   element you must inject a probe, which means serving your own copy of the page;
   served over file:// the site's real scripts never execute and the layout is
   wrong. Served over loopback HTTP they do, because a cross-origin <script src>
   is allowed — it is file:// that is special-cased, not cross-origin.
   So the probe and the screenshot both run against ONE loopback render, which is
   what makes the crop land where the measurement said it would. Measuring one
   render and cropping another is how the first version of this script produced a
   confidently wrong crop.

Prints the element's measured box, so a run that found nothing is distinguishable
from a run that found something.
"""
import argparse
import json
import os
import re
import subprocess
import sys
import tempfile

PROFILES = {
    # width, height, user agent ('' = default)
    "desktop": (1400, 1000, ""),
    "mobile": (390, 844,
               "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 "
               "(KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"),
}

MEASURE = """
(function(){
  var sel = %s;
  function emit(o){
    var p=document.createElement('pre'); p.id='LOOK'; p.textContent=JSON.stringify(o);
    document.body.appendChild(p);
  }
  function run(){
    var out={doc_height: Math.round(document.documentElement.scrollHeight),
             view_h: window.innerHeight, view_w: window.innerWidth};
    if (sel) {
      var el=document.querySelector(sel);
      if(!el){ out.found=false; emit(out); return; }
      var b=el.getBoundingClientRect();
      // DOCUMENT-absolute. Not viewport-relative and not scroll-adjusted: headless
      // --screenshot photographs the page from the top and IGNORES scroll position,
      // so scrolling the element into view moves the measurement and not the image.
      out.found=true;
      out.box={top:Math.round(b.top+window.scrollY), left:Math.round(b.left+window.scrollX),
               width:Math.round(b.width), height:Math.round(b.height)};
    }
    emit(out);
  }
  if(document.readyState==='complete') run(); else window.addEventListener('load', run);
})();
"""


def find_chrome():
    for c in [os.environ.get("CHROME"), "/snap/bin/chromium", "/usr/bin/chromium",
              "/usr/bin/chromium-browser", "/usr/bin/google-chrome"]:
        if c and os.path.exists(c):
            return c
    sys.exit("no chromium found (set CHROME=/path/to/chrome)")


def run_chrome(chrome, url, width, height, ua, extra):
    """One chromium run at an exact viewport. Returns (stdout, stderr)."""
    cmd = [chrome, "--headless=new", "--disable-gpu", "--no-sandbox", "--hide-scrollbars",
           "--window-size=%d,%d" % (width, height), "--virtual-time-budget=10000"]
    if ua:
        cmd.append("--user-agent=" + ua)
    cmd += extra + [url]
    r = subprocess.run(cmd, capture_output=True, text=True, timeout=300,
                       cwd=os.path.expanduser("~"))  # snap chromium: stay under $HOME
    return r.stdout, r.stderr


class _Serve:
    """Serve one spliced copy of the page over loopback, so the site's real
    cross-origin scripts execute (they do not over file://) AND our probe is
    present. One render is then both the measurement and the screenshot."""

    def __init__(self, html):
        import http.server, threading
        body = html.encode("utf-8")

        class H(http.server.BaseHTTPRequestHandler):
            def do_GET(inner):
                inner.send_response(200)
                inner.send_header("Content-Type", "text/html; charset=utf-8")
                inner.send_header("Content-Length", str(len(body)))
                inner.end_headers()
                inner.wfile.write(body)

            def log_message(inner, *a):
                pass

        self.srv = http.server.HTTPServer(("127.0.0.1", 0), H)
        self.port = self.srv.server_address[1]
        self.t = threading.Thread(target=self.srv.serve_forever, daemon=True)
        self.t.start()

    @property
    def url(self):
        return "http://127.0.0.1:%d/look.html" % self.port

    def close(self):
        self.srv.shutdown()
        self.srv.server_close()


def fetch_spliced(url, ua, selector):
    """The live page with our measuring probe appended, absolute-based.

    Retries: the origin self-throttles under a burst and answers with a connection
    reset (or an EMPTY 200, which is worse because it looks like a page). Both are
    transient. A short backoff is the difference between "the tool is broken" and
    "you asked six times in a minute".
    """
    import time
    import urllib.error
    import urllib.request
    req = urllib.request.Request(url, headers={"User-Agent": ua or "Mozilla/5.0"})
    raw = ""
    for attempt in range(4):
        try:
            raw = urllib.request.urlopen(req, timeout=60).read().decode("utf-8", "replace")
            if len(raw) > 500:
                break
            reason = "empty/short body (%d bytes) — origin throttling" % len(raw)
        except (urllib.error.URLError, OSError) as e:
            reason = str(e)
        if attempt == 3:
            sys.exit("could not fetch %s: %s" % (url, reason))
        wait = 3 * (attempt + 1)
        print("fetch attempt %d failed (%s) — retrying in %ds" % (attempt + 1, reason, wait),
              file=sys.stderr)
        time.sleep(wait)
    base = "/".join(url.split("/")[:3]) + "/"
    if "<base " not in raw:
        raw = raw.replace("<head>", '<head><base href="%s">' % base, 1)
    probe = MEASURE % (json.dumps(selector) if selector else "null")
    return raw.replace("</body>", "<script>" + probe + "</script></body>", 1)


def measure_and_shoot(chrome, served_url, width, height, ua, shot_path):
    """One chromium run: dump the DOM (carrying the probe result) — then a second
    run at the SAME viewport for the pixels. Same URL, same size, so the numbers
    from the first describe the image from the second."""
    out, err = run_chrome(chrome, served_url, width, height, ua, ["--dump-dom"])
    m = re.search(r'<pre id="LOOK">(.*?)</pre>', out, re.S)
    if not m:
        sys.exit("measurement produced no result\n" + err[-800:])
    import html as htmllib
    info = json.loads(htmllib.unescape(m.group(1)))
    if shot_path:
        _, err2 = run_chrome(chrome, served_url, width, height, ua, ["--screenshot=" + shot_path])
        if not os.path.exists(shot_path):
            sys.exit("screenshot not written\n" + err2[-800:])
    return info


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("url")
    ap.add_argument("--selector", default="", help="CSS selector to crop to")
    ap.add_argument("--profile", default="desktop", choices=sorted(PROFILES))
    ap.add_argument("--width", type=int, default=0, help="override the profile width")
    ap.add_argument("--pad", type=int, default=24, help="padding around the cropped element")
    ap.add_argument("--out", default="", help="output png (default ~/look.png)")
    args = ap.parse_args()

    chrome = find_chrome()
    w, h, ua = PROFILES[args.profile]
    if args.width:
        w = args.width

    out = os.path.expanduser(args.out) if args.out else os.path.expanduser("~/look.png")
    if not out.startswith(os.path.expanduser("~")):
        sys.exit("output must be under $HOME — snap chromium cannot write elsewhere (trap 1)")

    srv = _Serve(fetch_spliced(args.url, ua, args.selector))
    try:
        # THE RULE: measure and screenshot at ONE height. Page height is not a fixed
        # number — sections sized in vh grow with the viewport, so chasing the
        # document height diverges (measured 1000 -> 2854 -> 4152 -> 6141 -> 6453 on
        # one page). And scrolling does not help: headless --screenshot photographs
        # from the top and ignores scroll. So pick a generous fixed height, measure
        # AT it, shoot AT it, and crop by that measurement. Correct by construction,
        # because both numbers come from the same geometry. Grow only if the element
        # falls outside the frame.
        probe_h = max(h, 4000)
        for _ in range(3):
            info = measure_and_shoot(chrome, srv.url, w, probe_h, ua, None)
            if not (args.selector and info.get("found")):
                break
            bottom = info["box"]["top"] + info["box"]["height"] + args.pad
            if bottom <= probe_h:
                break
            probe_h = min(int(bottom * 1.15), 16000)
        info = measure_and_shoot(chrome, srv.url, w, probe_h, ua, out)
        full_h = probe_h
    finally:
        srv.close()

    if args.selector and not info.get("found"):
        sys.exit("selector %r matched nothing on %s — nothing to look at" % (args.selector, args.url))

    print("page   : %s  [%s, %dx%d, doc %dpx tall]" % (args.url, args.profile, w, full_h,
                                                       info["doc_height"]))
    print("full   : %s" % out)
    if not args.selector:
        return
    b = info["box"]
    print("element: %s  at (%d,%d) %dx%d" % (args.selector, b["left"], b["top"],
                                             b["width"], b["height"]))
    try:
        from PIL import Image
    except ImportError:
        print("       (install Pillow to crop)")
        return
    im = Image.open(out)
    crop = im.crop((max(0, b["left"] - args.pad), max(0, b["top"] - args.pad),
                    min(im.size[0], b["left"] + b["width"] + args.pad),
                    min(im.size[1], b["top"] + b["height"] + args.pad)))
    crop_path = out.replace(".png", "-crop.png")
    crop.save(crop_path)
    print("cropped: %s  %dx%d" % (crop_path, crop.size[0], crop.size[1]))


if __name__ == "__main__":
    main()
