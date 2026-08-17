"""Which of leopardess's 14 banned_phrases patterns actually fire, and where.

Runs the SITE'S OWN patterns (read live from site_specs) over the served text of
every active+deployed page. Tags are stripped first so hrefs and class names
cannot score — the gate reads slot TEXT, not markup, and counting a URL as copy
would inflate exactly the pattern this is trying to size.
"""
import html, json, re, subprocess, sys, time
from collections import defaultdict

SITE = "4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN = "leopardessconsulting.co.uk"
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-t", "-A", "-c"]

def q(sql):
    r = subprocess.run(PSQL + [sql], capture_output=True, text=True)
    if r.returncode: sys.exit(r.stderr[:400])
    return r.stdout.strip()

pats = json.loads(q(f"SELECT data->'voice_gate'->'banned_phrases' FROM site_specs "
                    f"WHERE site_id='{SITE}' AND aspect='voice' AND is_current"))
paths = [l for l in q(f"SELECT url FROM pages WHERE site_id='{SITE}' AND status='active' "
                      f"AND noindex=false AND deployed_at IS NOT NULL ORDER BY url").splitlines() if l]
print(f"{len(pats)} patterns x {len(paths)} pages")

cb = str(int(time.time()))
tally = defaultdict(lambda: {"hits": 0, "pages": set(), "examples": []})
for p in paths:
    body = subprocess.run(["curl", "-s", f"https://{DOMAIN}{p}?cb={cb}"],
                          capture_output=True, text=True).stdout
    body = re.sub(r"(?is)<(script|style)[^>]*>.*?</\1>", " ", body)
    main = re.search(r"(?is)<main.*?</main>", body)
    text = html.unescape(re.sub(r"<[^>]+>", " ", main.group(0) if main else body))
    text = re.sub(r"\s+", " ", text)
    for spec in pats:
        rx = re.compile(spec["pattern"], re.I)
        ms = list(rx.finditer(text))
        if ms:
            t = tally[spec["pattern"]]
            t["hits"] += len(ms); t["pages"].add(p); t["reason"] = spec["reason"]
            if len(t["examples"]) < 3:
                m = ms[0]
                t["examples"].append(f"{p}: …{text[max(0,m.start()-60):m.end()+60].strip()}…")

rows = sorted(tally.items(), key=lambda kv: -kv[1]["hits"])
total = sum(v["hits"] for v in tally.values())
print(f"\nTOTAL served banned-phrase hits: {total}\n")
for pat, v in rows:
    print(f"{v['hits']:>4} hits / {len(v['pages']):>2} pages   {pat}")
    print(f"                            reason: {v['reason']}")
    for e in v["examples"]:
        print(f"        {e[:190]}")
    print()
unfired = [s["pattern"] for s in pats if s["pattern"] not in tally]
print(f"patterns that fire on NOTHING ({len(unfired)}): " + ", ".join(unfired))
