#!/usr/bin/env python3
"""install_fence_facts.py — add the `facts` declaration to this lane's tool PLAN.

Why a script and not a SQL migration: the 288 lane's CONTRIB says "install it
through the lane's own fence installer; never hand-edit the doc_plans row", and
the reason is that the fence is JSON embedded in markdown — an edit that is
valid SQL can still produce a fence nothing can parse, and a fence nothing can
parse looks exactly like a tool that declares nothing.

FOUR RULES, taken from mortgagecalculator_couk_adoption/acceptance/install_fences.py
and re-checked against this site rather than inherited.

1. THE SUBJECT KEY IS READ FROM THE LADDER'S OWN DERIVATION, never built from a
   page name:
       CASE WHEN cc.component_level='tool' THEN cc.function
            ELSE regexp_replace(p.name,'^tool-','') END
   Ours has a tool-level component, so the key is `tool-sfi26-revenue-stacker`.
   Had it been a section component it would be `sfi26-revenue-stacker`, and a
   PLAN under the wrong key is a row nothing ever reads — silently, permanently.

2. `facts` GOES AT THE TOP LEVEL OF THE CRITERIA OBJECT, as a sibling of
   `profiles`/`checks`. Confirmed by the 288 lane against the parser:
   `json.Unmarshal(criteria, &struct{ Facts json.RawMessage })`. A per-check
   `facts` is refused by the validator's P7 inert-field rule — another silent
   nothing.

3. DECLARE WHAT THE TOOL ENCODES, not what is separately fenced with
   artifact_check. `facts` means "tell me when these move", so an incomplete list
   lets the rest drift silently, which is bugs_open/288's own class by omission.
   24 rates encoded, 24 declared. The four artifact_check entries are a different
   mechanism answering a different question and are deliberately a subset.

4. VERIFY BOTH DIRECTIONS BEFORE WRITING. Every declared id must exist in the
   register (or the sweep files fact_declaration_broken), and every code the tool
   encodes must be declared (or it drifts unwatched). A one-way check passes
   while missing half the failure surface.

Usage:  python3 install_fence_facts.py [--apply]      (dry run by default)
"""
import json, re, subprocess, sys

DOMAIN = "agritec.uk"
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-Atc"]

def q(sql):
    return subprocess.run(PSQL + [sql], capture_output=True, text=True, timeout=90).stdout.strip()

# Rule 1 — the key, from the ladder's own expression
key = q(f"""
SELECT CASE WHEN cc.component_level='tool' THEN cc.function
            ELSE regexp_replace(p.name,'^tool-','') END
FROM pages p JOIN sites s ON s.id=p.site_id
JOIN page_components pc ON pc.page_id=p.id AND pc.build_status<>'removed'
JOIN content_components cc ON cc.id=pc.component_id
WHERE s.domain='{DOMAIN}' AND p.page_type='tool';""")
if not key:
    sys.exit("no eligible tool page found — nothing to install")
print(f"subject_key (from the ladder's derivation): {key}")

# Rule 3/4 — what the tool encodes, and does each id exist?
tmpl = q(f"SELECT html_template FROM content_components WHERE function='{key}' LIMIT 1;")
encoded = sorted(set(re.findall(r"code:'(\w+)'", tmpl)))
ids = [f"ATT-sfi26-{c}" for c in encoded]
present = q(f"""
SELECT count(*) FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
 LATERAL jsonb_array_elements(ss.data->'facts') f
WHERE s.domain='{DOMAIN}' AND ss.aspect='evidence_base' AND ss.is_current
  AND f->>'id' = ANY(ARRAY[{','.join(chr(39)+i+chr(39) for i in ids)}]);""")
print(f"encoded rates: {len(encoded)} | declared: {len(ids)} | ids found in register: {present}")
if int(present) != len(ids):
    sys.exit("REFUSING: a declared id is not in the register — that files fact_declaration_broken")

# Rule 2 — put it at the top level of the criteria object
body = q(f"SELECT body FROM doc_plans WHERE subject_key='{key}' AND is_current;")
m = re.search(r"```criteria\s*(\{.*?\})\s*```", body, re.S)
if not m:
    sys.exit("no criteria fence found in the PLAN")
crit = json.loads(m.group(1))
if "facts" in crit:
    print("fence already declares facts:", crit["facts"])
    sys.exit(0)
crit["facts"] = ids
new_fence = "```criteria\n" + json.dumps(crit, indent=2) + "\n```"
new_body = body[:m.start()] + new_fence + body[m.end():]

# it must still parse after the edit — the whole reason this is a script
assert json.loads(re.search(r"```criteria\s*(\{.*?\})\s*```", new_body, re.S).group(1))["facts"] == ids
print(f"fence keys after edit: {list(crit.keys())}")

if "--apply" not in sys.argv:
    print("\nDRY RUN — pass --apply to write. Nothing changed.")
    sys.exit(0)

# Write via a dollar-quoted heredoc on stdin. psql variable interpolation with -c
# did not survive a multi-statement body containing backticks and quotes (it
# failed with "syntax error at or near :" and wrote nothing, which is the right
# way for it to fail); a dollar-quoted literal has no escaping surface at all.
tag = "$fence$"
if tag in new_body:
    sys.exit("REFUSING: the body contains the dollar-quote tag")
sql = f"""\\set ON_ERROR_STOP on
BEGIN;
UPDATE doc_plans SET is_current=false, superseded_at=now()
 WHERE subject_key='{key}' AND is_current;
INSERT INTO doc_plans (subject_type, subject_key, body, source, is_current, pinned, created_by)
VALUES ('tool', '{key}', {tag}{new_body}{tag}, 'manual', true, false, 'agritec-workstream-2026-08-25');
COMMIT;
"""
r = subprocess.run(["kubectl","-n","ai-persona-system","exec","-i","postgres-clients-0","--",
                    "psql","-U","clients_user","-d","clients_db"],
                   input=sql, capture_output=True, text=True, timeout=120)
print(r.stdout.strip() or r.stderr.strip())
