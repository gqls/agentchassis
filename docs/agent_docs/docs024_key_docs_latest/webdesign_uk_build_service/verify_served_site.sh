#!/bin/bash
# Visitor-grade verification of preview.webdesign.uk — pages, bans, head, and
# EVERY url() / href / src resolved from the serving root (style attrs included).
set -u
SCRATCH="/home/ant/.claude-scratch/claude-1000/-home-ant-projects-agentchassis/25c4e595-f646-44a8-a980-dae3e7289ecc/scratchpad"
mkdir -p "$SCRATCH/final"
FAIL=0
for p in index how-it-works what-you-get faq contact; do
  url="https://preview.webdesign.uk/$( [ $p = index ] && echo '' || echo $p.html )"
  code=$(curl -s -o "$SCRATCH/final/$p.html" -w '%{http_code}' --max-time 20 "$url")
  echo "PAGE $p: $code ($(wc -c < "$SCRATCH/final/$p.html") bytes)"
  [ "$code" = "200" ] || FAIL=1
done
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA \
  -c "SELECT jsonb_array_elements(data->'banned_claims')->>'pattern' FROM site_specs WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='evidence_base' AND is_current;" > "$SCRATCH/final/bans.txt"
python3 - <<'PY'
import re, html, glob, sys
scratch = "/home/ant/.claude-scratch/claude-1000/-home-ant-projects-agentchassis/25c4e595-f646-44a8-a980-dae3e7289ecc/scratchpad/final"
pats = [p for p in open(f"{scratch}/bans.txt").read().strip().split('\n') if p.strip()]
print(f"patterns loaded: {len(pats)}")
total = 0
urls = set()
for f in sorted(glob.glob(f"{scratch}/*.html")):
    raw = open(f).read()
    name = f.split('/')[-1]
    # every reference: href/src attributes AND url(...) in style attrs/blocks
    for m in re.finditer(r'(?:href|src)="([^"]+)"', raw): urls.add(m.group(1))
    for m in re.finditer(r'url\(\s*[\'"]?([^\'")]+)[\'"]?\s*\)', raw): urls.add(m.group(1))
    txt = re.sub(r'<(script|style)[^>]*>.*?</\1>', ' ', raw, flags=re.S|re.I)
    txt = html.unescape(re.sub(r'<[^>]+>', ' ', txt)); txt = ' '.join(txt.split())
    hits = 0
    for p in pats:
        found = re.findall(p, txt, re.I)
        if found:
            hits += len(found)
            print(f"BAN HIT {name} [{p[:38]}]: {found[:2]}")
    titles = [t for t in re.findall(r'<title>([^<]*)</title>', raw) if '—' in t]
    print(f"{name}: visible-ban-hits={hits} title-emdash={len(titles)}")
    total += hits + len(titles)
local = sorted(u for u in urls if u.startswith('/') and not u.startswith('//'))
open(f"{scratch}/local_urls.txt", 'w').write('\n'.join(local))
print(f"local references to check: {len(local)}")
sys.exit(1 if total else 0)
PY
SWEEP=$?
[ $SWEEP -ne 0 ] && FAIL=1
while IFS= read -r u; do
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 15 "https://preview.webdesign.uk$u")
  [ "$code" = "200" ] || { echo "REF FAIL $u: $code"; FAIL=1; }
done < "$SCRATCH/final/local_urls.txt"
echo "all local references checked"
echo "VERDICT: $( [ $FAIL = 0 ] && echo CLEAN || echo FAILURES-ABOVE )"
