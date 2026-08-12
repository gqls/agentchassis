#!/usr/bin/env python3
"""Arm the NARROW voice_gate on named sites — owner instruction 2026-08-12.

Enforces ONE banned phrase (the owner's own 2026-07-18 pattern). Density checks are
PARKED at values that cannot trip, because zero means "use leopardess's defaults" and
those encode one site's house style. Verified against the real Go parser
(datahelpers.ParseVoiceGate + ScanVoice): 9/9 cases, normal and long-form.

NOT parkable, by design of the check: `strawman` ("not X, but Y") and `flourish_ending`
(a block ending on "ultimately" / "in short" / "that's why"). Both are wanted here —
they are items 3 and 5 of the owner's 2026-08-12 critique.

Merges into an existing `voice` spec if there is one; creates the aspect otherwise.
Refuses to clobber a site that already has a gate.
"""
import json, subprocess, sys

SC = sys.argv[1]
SITES = sys.argv[2].split(",")
DRY = "--dry-run" in sys.argv

GATE = json.load(open(f"{SC}/voice_gate.json"))["voice_gate"]


def psql(sql, tuples=False):
    cmd = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
           "psql", "-U", "clients_user", "-d", "clients_db"]
    cmd += ["-tA", "-F", "\t"] if tuples else ["-v", "ON_ERROR_STOP=1"]
    cmd += ["-c", sql] if tuples else []
    r = subprocess.run(cmd, input=None if tuples else sql, text=True, capture_output=True)
    if r.returncode:
        raise SystemExit(f"psql failed: {r.stderr[-500:]}")
    return r.stdout


rows = psql("SELECT s.domain, coalesce(ss.data::text,'') FROM sites s "
            "LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='voice' "
            "AND ss.is_current WHERE s.domain IN (" +
            ",".join(f"'{d}'" for d in SITES) + ");", tuples=True)
existing = {}
for line in rows.splitlines():
    if "\t" in line:
        d, blob = line.split("\t", 1)
        existing[d] = blob

stmts, plan = [], []
for dom in SITES:
    if dom not in existing:
        raise SystemExit(f"site not found: {dom}")
    cur = json.loads(existing[dom]) if existing[dom].strip() else None
    if cur and cur.get("voice_gate"):
        print(f"SKIP {dom}: already has a voice_gate — not clobbering")
        continue
    if cur is None:
        data = {"voice_gate": GATE,
                "_note": ("Created 2026-08-12 to carry the fleet voice_gate. The site had "
                          "no `voice` aspect; nothing else in the platform reads this "
                          "aspect (checked: no Go reader beyond LoadVoiceGate, no live "
                          "agent config references specs.voice).")}
        plan.append(f"CREATE voice spec on {dom}")
    else:
        data = dict(cur)
        data["voice_gate"] = GATE
        plan.append(f"MERGE voice_gate into existing voice spec on {dom}")
    j = json.dumps(data, ensure_ascii=False)
    assert "$g$" not in j
    stmts.append(
        f"UPDATE site_specs SET is_current=false, superseded_at=now() "
        f"WHERE site_id=(SELECT id FROM sites WHERE domain='{dom}') "
        f"AND aspect='voice' AND is_current;")
    stmts.append(
        "INSERT INTO site_specs (site_id,aspect,data,source,source_agent,created_by,notes,"
        f"is_current) VALUES ((SELECT id FROM sites WHERE domain='{dom}'),'voice',"
        f"$g${j}$g$::jsonb,'operator:fleet_honest_20260812',"
        "'claude-ideauk-copy-20260812','claude-ideauk-copy-20260812',"
        "$n$Narrow voice_gate armed per owner instruction 2026-08-12: one banned phrase "
        "(honest), density checks parked deliberately (zero would inherit leopardess's "
        "house-style defaults). strawman and flourish_ending fire unconditionally by "
        "design of check_voice_tells and are wanted. Armed only on sites whose copy is "
        "already clean, so the gate starts from a clean baseline and any future item is "
        "a genuine regression.$n$,true);")

print("\n".join(plan))
sql = "BEGIN;\n" + "\n".join(stmts) + "\nCOMMIT;\n"
if DRY:
    print(f"\n-- {len(stmts)//2} sites, {len(sql)} bytes of SQL (dry run)")
    sys.exit(0)
psql(sql)
print(f"\napplied to {len(stmts)//2} sites")
