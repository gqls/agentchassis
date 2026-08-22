import json
from playwright.sync_api import sync_playwright
PROBE = open('/tmp/deployed_probe.js').read()
URL = "https://vonc.com/tools/gauntlet/index.html?cb=witness1755"
with sync_playwright() as pw:
    b = pw.chromium.launch()
    # the adapter's own mobile profile: 390x844, DSF 3, iPhone UA
    ctx = b.new_context(viewport={"width":390,"height":844}, device_scale_factor=3,
        user_agent="Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1")
    p = ctx.new_page()
    p.goto(URL, wait_until="networkidle"); p.wait_for_timeout(4000)  # adapter settleDelay
    out = p.evaluate(PROBE)
    print("probe marker :", out.get("probe"))
    print("located      :", out.get("located"))
    print("scanned      :", out.get("scanned"))
    fails = out.get("failures", [])
    print("failures     :", len(fails), " firm:", sum(1 for f in fails if not f["overImage"]),
          " overImage:", sum(1 for f in fails if f["overImage"]))
    for f in fails[:8]:
        print(f"   {'OVERIMG' if f['overImage'] else 'FIRM   '} {f['ratio']:>5}:1 need {f['need']} {f['selector'][:52]}")
    # WHY: what does effBG see on the section?
    bg = p.evaluate("""() => {
      const s = document.querySelector('.gauntlet-interface-section');
      if (!s) return null;
      const cs = getComputedStyle(s);
      return {backgroundColor: cs.backgroundColor, backgroundImage: cs.backgroundImage.slice(0,90)};
    }""")
    print("section bg   :", json.dumps(bg))
    b.close()
