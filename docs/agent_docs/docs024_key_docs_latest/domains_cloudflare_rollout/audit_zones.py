#!/usr/bin/env python3
"""Audit every Cloudflare zone against the portfolio template:
apex A 199.59.243.228 proxied + routes {domain/*, *.domain/*} -> portfolio-sites-router.
Read-only. Emits a TSV + summary."""
import json, urllib.request, sys

TOKEN = open("/home/ant/.config/cloudflare/token").read().strip()
API = "https://api.cloudflare.com/client/v4"
TEMPLATE_IP = "199.59.243.228"
SCRIPT = "portfolio-sites-router"
SKIP = {"relojistas.com", "finetuning.uk", "webdesign.uk", "idea.uk"}

def get(path):
    req = urllib.request.Request(API + path, headers={
        "Authorization": "Bearer " + TOKEN,
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) domains-rollout/1.0"})
    with urllib.request.urlopen(req, timeout=30) as r:
        d = json.load(r)
    if not d["success"]:
        raise RuntimeError(f"{path}: {d['errors']}")
    return d

zones = []
page = 1
while True:
    d = get(f"/zones?per_page=50&page={page}")
    zones += d["result"]
    if page >= d["result_info"]["total_pages"]:
        break
    page += 1

rows = []
for z in zones:
    name, zid = z["name"], z["id"]
    recs = get(f"/zones/{zid}/dns_records?per_page=200")["result"]
    routes = get(f"/zones/{zid}/workers/routes")["result"]
    apex = [r for r in recs if r["name"] == name and r["type"] in ("A", "AAAA", "CNAME")]
    www = [r for r in recs if r["name"] == f"www.{name}"]
    wild = [r for r in recs if r["name"] == f"*.{name}"]
    other = [r for r in recs if r not in apex + www + wild]
    apex_tpl = any(r["type"] == "A" and r["content"] == TEMPLATE_IP and r.get("proxied") for r in apex)
    have = {r["pattern"] for r in routes if r.get("script") == SCRIPT}
    routes_tpl = {f"{name}/*", f"*.{name}/*"} <= have
    odd_routes = [r for r in routes if r.get("script") != SCRIPT]
    status = "SKIP" if name in SKIP else (
        "TEMPLATE" if apex_tpl and routes_tpl and not other and not odd_routes else "DEVIANT")
    rows.append({
        "zone": name, "id": zid, "zstatus": z["status"], "class": status,
        "apex": "; ".join(f"{r['type']} {r['content'][:40]} proxied={r.get('proxied')}" for r in apex) or "NONE",
        "www": "; ".join(f"{r['type']} {r['content'][:40]} proxied={r.get('proxied')}" for r in www) or "-",
        "wild": "; ".join(f"{r['type']} {r['content'][:40]}" for r in wild) or "-",
        "n_other_recs": len(other),
        "other_recs": "; ".join(f"{r['type']} {r['name']}" for r in other[:6]),
        "routes_tpl": routes_tpl,
        "odd_routes": "; ".join(f"{r['pattern']}->{r.get('script')}" for r in odd_routes),
    })

with open(sys.argv[1] if len(sys.argv) > 1 else "/dev/stdout", "w") as f:
    cols = list(rows[0].keys())
    f.write("\t".join(cols) + "\n")
    for r in rows:
        f.write("\t".join(str(r[c]) for c in cols) + "\n")

for cls in ("TEMPLATE", "DEVIANT", "SKIP"):
    sub = [r for r in rows if r["class"] == cls]
    print(f"\n== {cls} ({len(sub)}) ==")
    for r in sub:
        www_flag = "www:YES" if r["www"] != "-" else "www:no"
        extra = f" | apex={r['apex']} | other={r['n_other_recs']} {r['other_recs']} | routes_tpl={r['routes_tpl']} {r['odd_routes']}" if cls != "TEMPLATE" else f" [{www_flag}]"
        print(f"  {r['zone']:32} {r['zstatus']:8}{extra}")
print(f"\ntotal zones: {len(rows)}")
