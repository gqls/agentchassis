#!/usr/bin/env python3
"""Post-deploy verification for a B2 batch. Identity over counts, everywhere it can be:
the tool render must appear VERBATIM; every inline script BODY from the pinned source
must appear VERBATIM; calculators.js compared by exact count; classes by per-class
count against the whole source page; ids by set difference."""
import json, os, re, subprocess, sys, urllib.request
S=os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0,S)
from b2_load import rendered_tool
# PIN IS IMPORTED, NEVER RESTATED (fixed 2026-08-15). This file used to hardcode
# PIN="0a0e89326" — the pin decompose_lmc had ALREADY abandoned as poisoned: it
# captured the 08-12 restore-deploys whose backup predates the bugs_open/224 fix,
# so it holds reverted arithmetic for some pages. The verifier's whole job is to
# say "live matches the source", and it was asking a source that is wrong.
# Measured 2026-08-15: harmless for the mixed-card five (byte-identical at both
# pins) and REAL for loans/standard-calc, which differs between them — so this was
# a live trap on the very page the 224 fix was about. Importing it means the pin
# can only ever be changed in one place.
from decompose_lmc import PINNED_REF as PIN
# seeds default to the lane's b2_seeds/, overridable for a batch built elsewhere
SEEDS=os.environ.get("B2_SEEDS", os.path.join(S,"b2_seeds"))
def get(u):
    r=urllib.request.Request(u, headers={"User-Agent":"Mozilla/5.0"})
    return urllib.request.urlopen(r, timeout=30).read().decode("utf-8","replace")
def src(name):
    rel=name.replace("-","/",1)+".html"
    return subprocess.run(["git","show",f"{PIN}:loanandmortgagecalculator.co.uk/{rel}"],
                          capture_output=True,text=True,cwd="/home/ant/projects/sites").stdout
def counts(s):
    c={}
    for m in re.findall(r'class="([^"]+)"',s):
        for x in m.split(): c[x]=c.get(x,0)+1
    return c
fails=0
for name in sys.argv[1:]:
    seed=json.load(open(f"{SEEDS}/{name}.json"))
    live=get("https://loanandmortgagecalculator.co.uk/"+name.replace("-","/",1)+".html")
    source=src(name)
    probs=[]
    if not live.startswith("<!DOCTYPE") or len(live)<3000: probs.append("not a page")
    if rendered_tool(seed) not in live: probs.append("tool render not verbatim")
    for m in re.finditer(r'<script(?!\s+src)[^>]*>(.*?)</script>', source, re.S):
        body=m.group(1).strip()
        if body and body not in live: probs.append("inline script body missing: "+body[:40].replace("\n"," "))
    if source.count("calculators.js") != live.count("calculators.js"):
        probs.append("calculators.js %d->%d" % (source.count("calculators.js"), live.count("calculators.js")))
    b,a=counts(source),counts(live)
    drops={k:(v,a.get(k,0)) for k,v in b.items() if a.get(k,0)<v and k not in ("container","container-tight")}
    if drops: probs.append("class drops %s" % drops)
    ids=sorted(set(re.findall(r'id="([^"]+)"',source))-set(re.findall(r'id="([^"]+)"',live)))
    if ids: probs.append("ids lost %s" % ids)
    fails += 1 if probs else 0
    print(("ok   " if not probs else "FAIL ")+f"{name:38} "+("; ".join(probs) if probs else "verbatim render + all script bodies + classes + ids"))
print(f"\n{len(sys.argv)-1-fails} of {len(sys.argv)-1} verified, {fails} failing")
sys.exit(1 if fails else 0)
