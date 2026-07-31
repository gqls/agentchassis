#!/usr/bin/env python3
"""bugs_open/161 step 2 — repair the 10 components asserting our tools run Monte Carlo.

Every replacement is ASSERTED to have matched. A replace() that quietly matches
nothing is precisely how this fix would look done without being done.
Run with --apply to write; default is a dry run.
"""
import json, subprocess, sys, re

APPLY = '--apply' in sys.argv

def psql(sql, tuples_only=True):
    cmd = ['kubectl','-n','ai-persona-system','exec','-i','postgres-clients-0','--',
           'psql','-U','clients_user','-d','clients_db','-v','ON_ERROR_STOP=1']
    if tuples_only: cmd += ['-At']
    r = subprocess.run(cmd, input=sql, capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError(f"psql failed: {r.stderr[:2000]}")
    return r.stdout

# (component_id, field, old, new)
# `field` is the content_data key; rendered_html gets the same swap.
REPAIRS = [
 # 1. game-auto-battler / generic-text-block
 ("ee2e15fe-016a-45ab-b908-2127519dad1a", "content",
  "Every query calculates 10,000 Monte Carlo trials to expose the absolute limits of player frustration.",
  "Every query models up to 10,000 attempts to expose the absolute limits of player frustration."),

 # 2. game-economy-simulator / generic-text-block  (two sentences)
 ("1a7c13d8-64fb-4fe1-a788-f29f468b8410", "content",
  "We run 10,000 Monte Carlo trials per query in the drop-rate tools to map the exact distribution of player outcomes.",
  "We model up to 10,000 attempts per query in the drop-rate tools to map the exact distribution of player outcomes."),
 ("1a7c13d8-64fb-4fe1-a788-f29f468b8410", "content",
  "Run Monte Carlo simulations to locate the worst-case scenario in your player base.",
  "Model the full distribution to locate the worst-case scenario in your player base."),

 # 3. guide-fairness-in-rng / generic-text-block
 ("b30cd04c-da23-489a-a3d7-f435508870f9", "content",
  "The tool runs 10,000 Monte Carlo trials per query to show the distribution of rewards across a player population.",
  "The tool models up to 10,000 attempts per query to show the distribution of rewards across a player population."),

 # 4. guide-fairness-in-rng / hero
 ("fc0a2929-ae9b-490a-a6c8-502655227a97", "subheadline",
  "Run 10,000 Monte Carlo trials per query directly in your browser to verify the odds yourself.",
  "Model up to 10,000 attempts per query directly in your browser to verify the odds yourself."),

 # 5. guide-rng-design / generic-text-block — the worst one: it spells out the falsehood
 ("bbcf1a0b-6e74-436e-9a06-d6ed0e4f2993", "content",
  "Our drop-rate tuner runs 10,000 Monte Carlo trials (repeated random simulations) per query to map this variance.",
  "Our drop-rate tuner models up to 10,000 attempts per query using exact probability rather than sampling, mapping this variance directly."),

 # 6. guide-rng-design / hero
 ("f4098832-2a15-482e-be91-a121c94f004c", "subheadline",
  "Model your drop chance, pity timers, and target hours using 10,000 Monte Carlo trials per query.",
  "Model your drop chance, pity timers, and target hours across up to 10,000 attempts per query."),

 # 7. guide-skinner-box / generic-text-block
 ("2c505b7a-4d89-46fd-a633-f2337319fe33", "content",
  "Our drop-rate tuner runs exactly 10,000 Monte Carlo trials per query to simulate live sessions.",
  "Our drop-rate tuner models up to 10,000 attempts per query to map live sessions exactly."),

 # 8. guide-skinner-box / hero
 ("c4ee1634-f242-401b-a7d7-71c5e8d05f34", "subheadline",
  "The drop-rate tuner runs 10,000 Monte Carlo trials per query across 4 configurable inputs.",
  "The drop-rate tuner models up to 10,000 attempts per query across 4 configurable inputs."),

 # 9. tool-spawn-rate-balancer-guide / article-body — heading + framing + claim + CTA
 ("f5cc4012-fc3c-4572-a068-5d7fdd079bc6", "content",
  "The Solution: Pity Timers and Monte Carlo",
  "The Solution: Pity Timers and Exact Probability"),
 ("f5cc4012-fc3c-4572-a068-5d7fdd079bc6", "content",
  "We solve this by simulating the player lifetime using Monte Carlo methods. A Monte Carlo simulation runs thousands of random attempts to map the exact distribution of outcomes.",
  "We solve this by computing the distribution directly: a closed-form binomial and geometric model of the player lifetime, rather than approximating it by sampling. (A Monte Carlo simulation would run thousands of random attempts to estimate the same distribution; computing it exactly is faster and carries no sampling error.)"),
 ("f5cc4012-fc3c-4572-a068-5d7fdd079bc6", "content",
  "It runs 10,000 Monte Carlo trials per query directly in your browser.",
  "It models up to 10,000 attempts per query directly in your browser."),
 ("f5cc4012-fc3c-4572-a068-5d7fdd079bc6", "content",
  "Run Monte Carlo simulations to map exactly how many players will miss your target hours.",
  "Map the full distribution to see exactly how many players will miss your target hours."),

 # 10. tool-spawn-rate-balancer-guide / call-to-action
 ("346897f3-94c0-4f09-a5c5-76c9aced39ac", "subheadline",
  "The drop-rate tuner runs 10,000 Monte Carlo trials per query.",
  "The drop-rate tuner models up to 10,000 attempts per query."),
]

# tool-loot-table-balancer-guide (c80dbae1) is DELIBERATELY ABSENT: its only
# Monte Carlo prose is general technique teaching and advice to the reader, which
# is true and stays. Verified it does not fire the banned patterns.

def fetch(cid):
    # Fetched separately: -At joins columns with a pipe and the copy contains pipes.
    cd = psql(f"SELECT content_data::text FROM page_components WHERE id='{cid}';")
    rh = psql(f"SELECT COALESCE(rendered_html,'') FROM page_components WHERE id='{cid}';")
    return cd, rh

failures = []
plan = []
for cid, field, old, new in REPAIRS:
    cd_raw, rh_raw = fetch(cid)
    try:
        cd = json.loads(cd_raw)
    except Exception as e:
        failures.append(f"{cid}: content_data did not parse: {e}")
        continue
    val = cd.get(field)
    in_field = (val or "").count(old)
    in_html  = rh_raw.count(old)
    status = "OK" if in_field >= 1 else "*** FIELD MISS ***"
    if in_field < 1:
        failures.append(f"{cid} [{field}]: old string not found in content_data")
    if in_html < 1:
        failures.append(f"{cid} [{field}]: old string not found in rendered_html (served bytes would stay false)")
    plan.append((cid, field, old, new, in_field, in_html, status))

print(f"{'component':38} {'field':14} {'cd':>3} {'html':>4}  status")
for cid, field, old, new, a, b, st in plan:
    print(f"{cid:38} {field:14} {a:3} {b:4}  {st}")
    print(f"    - {old[:110]}")
    print(f"    + {new[:110]}")

if failures:
    print("\n*** PRE-FLIGHT FAILURES — nothing applied ***")
    for f in failures: print("  -", f)
    sys.exit(1)

print(f"\npre-flight clean: {len(plan)} replacements, every one matched in BOTH content_data and rendered_html")

if not APPLY:
    print("dry run — pass --apply to write")
    sys.exit(0)

# Apply, one statement per replacement, asserting the row changed.
for cid, field, old, new, _, _, _ in plan:
    o = old.replace("'", "''")
    n = new.replace("'", "''")
    sql = f"""
BEGIN;
UPDATE page_components
SET content_data = jsonb_set(content_data, '{{{field}}}',
        to_jsonb(replace(content_data->>'{field}', '{o}', '{n}'))),
    rendered_html = replace(rendered_html, '{o}', '{n}'),
    updated_at = now()
WHERE id = '{cid}'
  AND (content_data->>'{field}' LIKE '%{o}%' OR rendered_html LIKE '%{o}%');
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM page_components
    WHERE id = '{cid}'
      AND ((content_data->>'{field}') LIKE '%{o}%' OR rendered_html LIKE '%{o}%');
    IF n <> 0 THEN
        RAISE EXCEPTION 'bug161: the old string survived on {cid} [{field}]';
    END IF;
END $$;
COMMIT;
"""
    psql(sql, tuples_only=False)
    print(f"applied: {cid} [{field}]")

print("\nall replacements applied and asserted")
