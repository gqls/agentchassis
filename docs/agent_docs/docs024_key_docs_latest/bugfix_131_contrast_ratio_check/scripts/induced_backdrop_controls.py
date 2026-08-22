# Induced controls for the bounded-backdrop refinement. Each case is built to
# come out a KNOWN way; a probe that always fails or never fails cannot pass all four.
from playwright.sync_api import sync_playwright
PROBE = open('/tmp/deployed_probe.js').read()

HTML = """<!doctype html><html><body style="margin:0">
<div id="A" style="background:#6d28d9;padding:20px">
  <p class="flat-bad" style="color:#7c3cff;font-size:16px">A flat opaque bad</p></div>
<div id="B" style="background:#6d28d9;background-image:linear-gradient(rgba(0,0,0,0.2),rgba(0,0,0,0));padding:20px">
  <p class="grad-bad" style="color:#7c3cff;font-size:16px">B gradient over opaque, bad</p></div>
<div id="C" style="background:#6d28d9;background-image:url('data:image/svg+xml;utf8,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%221%22 height=%221%22><rect width=%221%22 height=%221%22 fill=%22white%22/></svg>');padding:20px">
  <p class="url-bad" style="color:#7c3cff;font-size:16px">C url image, unknowable</p></div>
<div id="D" style="background:#6d28d9;background-image:linear-gradient(rgba(0,0,0,0.2),rgba(0,0,0,0));padding:20px">
  <p class="grad-good" style="color:#ffffff;font-size:16px">D gradient over opaque, good</p></div>
</body></html>"""

with sync_playwright() as pw:
    b = pw.chromium.launch(); p = b.new_context(viewport={"width":390,"height":844}).new_page()
    p.set_content(HTML); p.wait_for_timeout(300)
    out = p.evaluate(PROBE)
    got = {f["selector"].split(".")[-1]: f for f in out["failures"]}
    def show(name, expect):
        f = got.get(name)
        if not f: actual = "NOT FLAGGED"
        elif f["overImage"]: actual = f"approx (never fails) {f['ratio']}:1"
        elif f.get("gradientBounded"): actual = f"FIRM bounded {f['ratio']}:1"
        else: actual = f"FIRM flat {f['ratio']}:1"
        print(f"  {name:10} expect {expect:24} -> {actual:28} {'OK' if expect in actual else '** MISMATCH **'}")
    print(f"scanned={out['scanned']} failures={len(out['failures'])}")
    show("flat-bad",  "FIRM flat")
    show("grad-bad",  "FIRM bounded")
    show("url-bad",   "approx")
    show("grad-good", "NOT FLAGGED")
    b.close()
