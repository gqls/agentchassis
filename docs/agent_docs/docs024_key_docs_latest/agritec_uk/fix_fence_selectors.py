#!/usr/bin/env python3
"""fix_fence_selectors.py — point the acceptance fence at the ids the tool ACTUALLY has.

THE DEFECT. The tool's element ids are INSTANCE-SCOPED — `ScopeToolBirthTemplate`
rewrites them at birth so two copies of a tool on one page cannot collide, giving
`c-tool-sfi26-revenue-stacker-total-area-input`. The auto-generated acceptance
fence names the UNSCOPED form, `#total-area-input`. So the Tier-4 interaction
check addresses markup that does not exist and can never pass — while the tool
itself is fine.

This is the estate's own identity-agreement landmine one layer down: the fence,
the markup and the scoping rule have to name the same thing, and nothing checks
that they do. A fence that cannot pass looks exactly like a tool that fails.

RULE: selectors are READ FROM THE SERVED PAGE, never constructed by prefixing.
The scoping rule is the platform's, not ours, and reconstructing it in Python is
how the next rename becomes a silent failure.

Usage:  python3 fix_fence_selectors.py [--apply]      (dry run by default)
"""
import json, re, subprocess, sys

KEY = "tool-sfi26-revenue-stacker"
URL = "https://agritec.uk/tools/sfi26-revenue-stacker/"
PAGE = "/tmp/live_tool.html"   # fetched with curl: Cloudflare 403s urllib's default agent
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-Atc"]

def q(sql):
    return subprocess.run(PSQL + [sql], capture_output=True, text=True, timeout=90).stdout.strip()

subprocess.run(["curl","-sS","-L","--max-time","30",URL,"-o",PAGE], check=True, timeout=60)
html = open(PAGE, encoding="utf-8", errors="replace").read()
ids = set(re.findall(r'id="([^"]+)"', html))
print(f"ids present on the served page: {len(ids)}")

body = q(f"SELECT body FROM doc_plans WHERE subject_key='{KEY}' AND is_current;")
m = re.search(r"```criteria\s*(\{.*?\})\s*```", body, re.S)
crit = json.loads(m.group(1))

# every #selector the fence names, checked against the page
named = sorted(set(re.findall(r'"#([\w-]+)"', json.dumps(crit))))
missing = [s for s in named if s not in ids]
print(f"fence names {len(named)} id selectors; {len(missing)} are not on the page: {missing}")
if not missing:
    print("nothing to fix")
    sys.exit(0)

# resolve each missing one by finding the served id that ENDS with it — the
# scoping rule is a prefix, so this reads the answer off the page rather than
# rebuilding the rule. Refuse if it is not exactly one match.
fixed = dict(crit)
raw = json.dumps(crit)
for s in missing:
    cands = [i for i in ids if i.endswith("-" + s) or i == s]
    if len(cands) != 1:
        sys.exit(f"REFUSING: '{s}' resolves to {len(cands)} ids {cands} — ambiguous, not guessing")
    print(f"  #{s}  ->  #{cands[0]}")
    raw = raw.replace(f'"#{s}"', f'"#{cands[0]}"')
crit = json.loads(raw)

# post-condition: every selector the fence names now exists on the page
still = [s for s in set(re.findall(r'"#([\w-]+)"', json.dumps(crit))) if s not in ids]
assert not still, f"still missing after fix: {still}"
assert len(crit.get("facts", [])) == 24, "facts declaration must survive untouched"
assert len(crit.get("checks", [])) == 5, "all five checks must survive"
print(f"post-check: all fence selectors exist; facts={len(crit['facts'])} checks={len(crit['checks'])}")

new_body = body[:m.start()] + "```criteria\n" + json.dumps(crit, indent=2) + "\n```" + body[m.end():]
if "--apply" not in sys.argv:
    print("\nDRY RUN — pass --apply to write. Nothing changed.")
    sys.exit(0)

tag = "$fence$"
assert tag not in new_body
sql = f"""\\set ON_ERROR_STOP on
BEGIN;
UPDATE doc_plans SET is_current=false, superseded_at=now() WHERE subject_key='{KEY}' AND is_current;
INSERT INTO doc_plans (subject_type, subject_key, body, source, is_current, pinned, created_by)
VALUES ('tool','{KEY}',{tag}{new_body}{tag},'manual',true,false,'agritec-workstream-2026-08-25');
COMMIT;
"""
r = subprocess.run(["kubectl","-n","ai-persona-system","exec","-i","postgres-clients-0","--",
                    "psql","-U","clients_user","-d","clients_db"], input=sql,
                   capture_output=True, text=True, timeout=120)
print(r.stdout.strip() or r.stderr.strip())
