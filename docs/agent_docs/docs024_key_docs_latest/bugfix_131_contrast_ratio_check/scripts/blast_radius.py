# How much does the bounded-backdrop refinement change, fleet-wide?
# Runs the REFINED probe (extracted from source) on live pages and reports, per page:
#   firm      = would FAIL the check now
#   bounded   = of those, how many are the NEW branch (previously discarded)
#   unbounded = still not judged
import sys, json
from playwright.sync_api import sync_playwright
PROBE = open('/tmp/deployed_probe.js').read()
# TARGETS COME FROM THE FLEET, NEVER FROM RECALL. A hand-typed list produced a
# would-be bug report against someone else's domain on 2026-08-22 (WRONG_CALLS,
# fourth row): I wrote dartsonline.co.uk; the fleet's site is dartsonline.com.
# Pipe the list in:
#   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user \
#     -d clients_db -tAc "SELECT DISTINCT s.domain FROM sites s JOIN pages p \
#     ON p.site_id=s.id WHERE p.deployed_at IS NOT NULL ORDER BY 1;" \
#     | sed 's|^|https://|' | python3 blast_radius.py
PAGES = [l.strip() for l in sys.stdin if l.strip()] if not sys.stdin.isatty() else [
 "https://vonc.com/tools/gauntlet/index.html",   # fallback: the witness page only
]
with sync_playwright() as pw:
    b = pw.chromium.launch()
    ctx = b.new_context(viewport={"width":390,"height":844}, device_scale_factor=3)
    print(f"{'page':42} {'scan':>5} {'firm':>5} {'newly':>6} {'unbnd':>6}  worst-new")
    tot_firm=tot_new=0
    for url in PAGES:
        p = ctx.new_page()
        try:
            p.goto(url, wait_until="networkidle", timeout=45000); p.wait_for_timeout(3000)
            o = p.evaluate(PROBE)
            if o.get("probe") != "contrast_ratio/v1" or o.get("scanned",0)==0:
                print(f"{url[:42]:42} {'--':>5}  probe did not measure (fail-closed)"); p.close(); continue
            f = o["failures"]
            firm=[x for x in f if not x["overImage"]]
            new=[x for x in firm if x.get("gradientBounded")]
            unb=[x for x in f if x["overImage"]]
            worst = min(new, key=lambda x:x["ratio"]) if new else None
            wtxt = f'{worst["ratio"]}:1 {worst["selector"][:30]}' if worst else "-"
            print(f"{url[:42]:42} {o['scanned']:>5} {len(firm):>5} {len(new):>6} {len(unb):>6}  {wtxt}")
            tot_firm+=len(firm); tot_new+=len(new)
        except Exception as e:
            print(f"{url[:42]:42} ERROR {str(e)[:40]}")
        p.close()
    print(f"\nTOTAL firm={tot_firm}  of which NEWLY judged by the refinement={tot_new}")
    b.close()
