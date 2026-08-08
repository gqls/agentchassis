#!/usr/bin/env python3
"""fix_voice_exemplars.py — rewrite loancalculator's voice EXEMPLARS so they stop
teaching the tic. Second half of the 2026-08-08 rule-trim trial, whose first half
failed for exactly this reason.

WHY. `trim_voice_rules.py` removed the rule mandating a conditional opening. The tic
survived untouched, because THREE OF FOUR worked examples still demonstrate it:

    before_after_1_consolidation  AFTER: "If you're thinking about rolling several debts…"
    before_after_2_car_finance    AFTER: "If you're looking at car finance…"
    before_after_3_overpayment    AFTER: "If you have spare money each month…"

Our own voice document says why the examples won: "A writer model follows exemplars more
reliably than rules — the rules explain the register, the pairs teach it." So removing a
rule while leaving its worked example changes nothing. **The example IS the instruction.**

A SECOND instance of the same oversight, found writing this: `register_anchor` still ends
"Your numbers stay on your own screen" — the privacy line the owner cut from the homepage
on 2026-08-08, and whose RULE was demoted the same day. The exemplar kept teaching it.

THE DESIGN, AND THE TRAP INSIDE IT. The fix is NOT "no conditional openings". Four
examples that all avoid conditionals would teach the opposite tic just as reliably. The
lesson has to be VARIETY, and the only way an exemplar set can teach variety is to BE
varied. So the four openings below are deliberately four different shapes, and one of
them is still a conditional — used on the one case where the reader's situation genuinely
is the point:

    register_anchor   opens by normalising          "Most of us have…"
    consolidation     opens with the plain fact     "Rolling several debts into one…"
    car_finance       opens with the choice         "Car finance usually comes in…"
    overpayment       opens with the mechanism      "Paying extra off a loan saves…"
    debt_distress     opens with a conditional      "If you're reading this because…"

Also fixed here: the consolidation exemplar taught "the appeal is usually the monthly
payment". The owner flagged "appeal" on 2026-08-08 — in a debt context it reads as a legal
appeal before it reads as attraction — and the live page had it because the exemplar did.

THE GATE. Same formatter port and same refusal discipline as seed_voice_h.py and
trim_voice_rules.py: regenerate the CURRENT spec and require it to reproduce the STORED
`formatted` before writing, because the writer reads only that field.

Backup: site_specs_bak_20260808_ruletrim holds the pre-trim spec. This script takes its
own before writing.

Run:  python3 fix_voice_exemplars.py            # dry run
      python3 fix_voice_exemplars.py --apply    # write it
"""
import json
import subprocess
import sys

APPLY = "--apply" in sys.argv
DOMAIN = "loancalculator.co.uk"
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db"]


def psql(sql, tuples=True):
    cmd = PSQL + (["-t", "-A"] if tuples else [])
    r = subprocess.run(cmd + ["-c", sql], capture_output=True, text=True, timeout=180)
    if r.returncode != 0:
        sys.exit(f"psql failed:\n{r.stderr.strip()}")
    return r.stdout.strip()


def humanise(k):
    k = k.replace("_", " ")
    return k[:1].upper() + k[1:] if k else k


def fmt_value(label, val):
    if isinstance(val, str):
        return "" if val == "" else f"{label}: {val}"
    if isinstance(val, list):
        strs = [v for v in val if isinstance(v, str) and v != ""]
        if not strs:
            return ""
        return "\n".join([f"{label}:"] + [f"- {s}" for s in strs])
    if isinstance(val, dict):
        parts = [p for p in (fmt_value(humanise(k), v) for k, v in val.items()) if p]
        return f"{label}:\n" + "\n".join(parts) if parts else ""
    return ""


def format_cd(spec):
    return "\n\n".join(s for s in (fmt_value(humanise(k), v) for k, v in spec.items()
                                   if k != "formatted" and v is not None) if s)


# Five exemplars, five different opening shapes. The variety IS the lesson.
NEW_EXEMPLARS = {
    "register_anchor": (
        "Most of us have more than one kind of borrowing, and each kind affects the "
        "others. Take on a car loan and a mortgage lender will usually offer you less. "
        "Remortgage, and the cost of your other borrowing can shift too. The calculators "
        "on this site are free, and they are built to show you those connections."
    ),
    "opening_shape_1_plain_fact_consolidation": (
        "BEFORE: 'Consolidating can lower your monthly payments, but extending the term "
        "might cost you more overall. Enter your current debts below to compare the true "
        "total cost.'  AFTER: 'Rolling several debts into one loan usually brings the "
        "monthly payment down, and that is normally why people do it. What is easy to "
        "miss is the time. A consolidation loan often runs longer than the debts it "
        "replaces, so more months of interest can cost more in total even when each "
        "payment feels lighter. Put your current debts in below and this checker shows "
        "you both sides.'"
    ),
    "opening_shape_2_the_choice_car_finance": (
        "BEFORE: 'Understand the real cost of your car finance. Hire Purchase (HP) leads "
        "to ownership; Personal Contract Purchase (PCP) keeps payments low but carries a "
        "final Balloon.'  AFTER: 'Car finance usually comes in one of two shapes. With "
        "hire purchase you pay the car off month by month and it ends up yours. With a "
        "personal contract purchase the monthly payments are smaller, and a large final "
        "payment is left at the end. Dealers call that final payment the balloon. This "
        "calculator shows what each route really costs over the life of the deal.'"
    ),
    "opening_shape_3_the_mechanism_overpayment": (
        "BEFORE: 'Overpaying your loan reduces the total interest you pay. However, some "
        "lenders charge an Early Repayment Charge (ERC).'  AFTER: 'Paying extra off a "
        "loan saves you interest, because interest is charged on what you still owe. "
        "There is one thing worth checking first. Some lenders charge a fee for paying "
        "off early, and on a personal loan that is usually up to 58 days of interest. "
        "This calculator shows the saving and the fee side by side.'"
    ),
    "opening_shape_4_a_conditional_where_it_earns_it": (
        "A conditional opening is one option among several, not the house shape. It earns "
        "its place when the reader's own situation is genuinely the subject — as here, "
        "where the page is for someone already in difficulty:  'If you are reading this "
        "because the payments have stopped adding up, two things are worth saying before "
        "anything else.'  On a page that is simply explaining how something works, open "
        "with the thing instead."
    ),
}

raw = psql(f"""SELECT sp.data::text FROM site_specs sp JOIN sites s ON s.id=sp.site_id
WHERE s.domain='{DOMAIN}' AND sp.aspect='content_direction' AND sp.is_current;""")
if not raw:
    sys.exit(f"no current content_direction for {DOMAIN}")
cd = json.loads(raw)
stored = cd.get("formatted", "")

regen = format_cd({k: v for k, v in cd.items() if k != "formatted"})
if sorted(regen.split("\n")) != sorted(stored.split("\n")):
    sys.exit("  FORMATTER GATE FAILED — re-port datahelpers.FormatContentDirection.")
print(f"  formatter gate PASSES ({len(stored)} bytes, {len(stored.splitlines())} lines)")

old_ex = cd.get("voice_exemplars") or {}
if not old_ex:
    sys.exit("  REFUSING: no voice_exemplars present — the spec is not what this expects.")

# The three tic-teaching exemplars must actually be there, or the premise has moved.
expected_gone = ["before_after_1_consolidation", "before_after_2_car_finance",
                 "before_after_3_overpayment"]
missing = [k for k in expected_gone if k not in old_ex]
if missing:
    sys.exit(f"  REFUSING: expected to replace {missing} and they are absent. "
             "The spec changed since this was written — re-read it.")

new_cd = json.loads(raw)
new_cd["voice_exemplars"] = NEW_EXEMPLARS
new_cd["formatted"] = format_cd({k: v for k, v in new_cd.items() if k != "formatted"})

print(f"\n  {DOMAIN}")
print(f"    exemplars {len(old_ex)} -> {len(NEW_EXEMPLARS)}")
for k in old_ex:
    print(f"      REMOVED  {k}")
for k in NEW_EXEMPLARS:
    print(f"      ADDED    {k}")

fmt = new_cd["formatted"]
# Count conditional openings actually demonstrated in the new AFTER text.
tic = fmt.count("AFTER: 'If you")
print(f"\n  conditional openings demonstrated in AFTER examples: {tic} "
      f"(was 3 of 3 before/after pairs)")
for gone, why in [("Your numbers stay on your own screen", "the privacy line the owner cut"),
                  ("the appeal is usually", "the 'appeal' wording the owner flagged")]:
    if gone in fmt:
        sys.exit(f"  REFUSING: «{gone}» still present in `formatted` — {why}.")
print("  removal check PASSES — the privacy line and the 'appeal' wording are both gone "
      "from what the writer reads")
print(f"  formatted  {len(stored)} -> {len(fmt)} bytes ({len(fmt)-len(stored):+d})")

if not APPLY:
    print("\n  DRY RUN — rerun with --apply to write.")
    sys.exit(0)

psql("""CREATE TABLE IF NOT EXISTS site_specs_bak_20260808_exemplars AS
SELECT * FROM site_specs WHERE site_id=(SELECT id FROM sites WHERE domain='loancalculator.co.uk')
  AND aspect='content_direction' AND is_current;""", tuples=False)

payload = json.dumps(new_cd).replace("'", "''")
psql(f"""
BEGIN;
UPDATE site_specs SET is_current = false, superseded_at = NOW()
 WHERE site_id = (SELECT id FROM sites WHERE domain='{DOMAIN}')
   AND aspect='content_direction' AND is_current;
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
VALUES ((SELECT id FROM sites WHERE domain='{DOMAIN}'), 'content_direction',
        '{payload}'::jsonb, 'operator', 'fix_voice_exemplars.py',
        'Exemplar fix (owner 2026-08-08): the rule trim left three of four worked examples demonstrating the conditional opening, so the tic survived. Five exemplars now show five different opening shapes, one of them still a conditional so the opposite tic is not taught. Also removes the privacy line and the "appeal" wording the owner rejected. Backup: site_specs_bak_20260808_exemplars',
        true, 'cqls');
COMMIT;""", tuples=False)
print("\n  APPLIED.")
print(psql(f"""SELECT '  now: exemplars='||jsonb_object_keys_count||' formatted_bytes='||fb
FROM (SELECT (SELECT count(*) FROM jsonb_object_keys(data->'voice_exemplars')) AS jsonb_object_keys_count,
             length(data->>'formatted') AS fb
      FROM site_specs WHERE site_id=(SELECT id FROM sites WHERE domain='{DOMAIN}')
      AND aspect='content_direction' AND is_current) t;"""))
