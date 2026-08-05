#!/usr/bin/env python3
"""seed_voice_h.py — seed the "gentle explanatory" (H register) voice into
loancalculator.co.uk's content_direction, so the FRAMEWORK's page-content-writer
produces it. Owner chose H on 2026-08-05 (portfolio_positioning/
VOICE_gentle_explanatory_v1.md, trial H round 2).

WHY A SCRIPT AND NOT A HAND-WRITTEN UPDATE. The writer reads exactly ONE field of
content_direction — `{{.site_specs.specs.content_direction.formatted}}`, live in
page-content-writer's prompt_template. Every other key reaches the prompt only by
being serialised INTO `formatted` by datahelpers.FormatContentDirection. **A
hand-written content_direction that forgets to regenerate `formatted` is INVISIBLE
to the writer: the edit looks applied and changes nothing.** The formatter port
below is lifted from the proven set_divergence_specs.py in the
loanandmortgagecalculator lane, including its gate.

THE GATE. The port is not trusted. Before writing anything it regenerates the
CURRENT spec and asserts the result matches the STORED `formatted` as a multiset
of lines — not as a string, because Go map iteration order is random so section
order in the stored value is arbitrary and carries no meaning. If the port has
drifted from the Go, this refuses to write.

THE MERGE IS NOT ADDITIVE, AND THAT IS THE POINT. loancalculator's existing spec
carries rules that CONTRADICT the H register — most sharply "Avoid contractions in
declarative or authoritative statements", against H's "contractions wherever they
would be spoken". Appending H's rules while leaving those in place would hand the
writer two opposing instructions and let the model pick. Each conflict below is
REPLACED, and the replacement is listed in CHANGES so the diff is reviewable.

Run:  python3 seed_voice_h.py            # dry run, shows the diff
      python3 seed_voice_h.py --apply    # write it
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


# ── faithful port of datahelpers.FormatContentDirection ──────────────────────────
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


# ── the H register, as this site's own voice ─────────────────────────────────────
# voice.* keys REPLACED (the rest of the voice dict is preserved untouched).
VOICE_REPLACE = {
    "register": (
        "Gentle explanatory: a knowledgeable friend who has worked in lending and has "
        "nothing to sell, taking the time to explain non-obvious ideas at the pace of "
        "someone hearing them for the first time. Plain truths the industry prefers you "
        "not to hear, delivered without lecturing, performing expertise, or assuming "
        "context the reader has not been given."
    ),
    "formality": (
        "Conversational-plain: full sentences, no slang, no exclamation marks, and "
        "contractions wherever they would be spoken (it's, you're, they'll, don't). The "
        "test for every paragraph is that it could be read aloud to a friend, unchanged, "
        "without either of you wincing. Short declarative sentences keep it from feeling "
        "stiff; closer to a knowledgeable friend talking than to a broadsheet explainer."
    ),
    "emotional_tone": (
        "Calm, warm, and mildly adversarial toward the lending industry on the reader's "
        "behalf — not angry, but consistently alert to how lenders frame things versus "
        "how they actually work. The effect on the reader is quiet empowerment, and the "
        "reader should feel ordinary, not behind."
    ),
}

# writing_rules entries REPLACED. Key = a distinctive substring of the LIVE rule to
# retire; value = its replacement (None = delete outright).
RULES_REPLACE = {
    "Avoid contractions in declarative": (
        "Use contractions wherever they would be spoken: it's, you're, they'll, don't. "
        "Formal 'cannot' survives only inside a genuinely earned contrast. (REPLACES the "
        "previous no-contractions-in-declaratives rule, which contradicted this register.)"
    ),
    "Rhetorical questions are permitted": (
        "Open sections where the reader is standing, not where the fact is: begin with a "
        "conditional or situational clause ('If you have a car loan…', 'If you're thinking "
        "about…') before the first assertion. Never open cold with the assertion, and never "
        "with a negative twist ('X isn't about Y…'). Headings are statements or noun "
        "phrases, not questions. (REPLACES the rhetorical-question opener rule.)"
    ),
    "Lead every guide section with the reader's most likely question": (
        "Lead every guide section with the practical bottom line, reached through the "
        "reader's own situation rather than announced. Never bury the answer deep in the "
        "explanation, and never advertise its importance before giving it."
    ),
}

# writing_rules ADDED (H rules with no conflicting incumbent).
RULES_ADD = [
    "Explain before you name: when a technical term must appear, describe the thing in "
    "plain words first, then attach its name — '…a large final payment is left at the end. "
    "Dealers call that final payment the balloon.' Bold the term at the point it is named.",
    "One idea, one sentence: no em-dash chains and no semicolon joins. Where a consequence "
    "needs attaching, give it its own short sentence.",
    "Hedge for accuracy, never for performance: 'usually', 'can', 'often', 'roughly' are "
    "honest in lending and welcome; never advertise importance ('crucially', 'the most "
    "important thing to understand is') — put the important fact early instead.",
    "Normalise where it is true ('Most of us have more than one kind of borrowing') — the "
    "reader should feel ordinary, not behind.",
    "State facts positively, including privacy and cost: 'free', 'your numbers stay on your "
    "own screen' — never a negation pile ('no sign-up, no credit check, nothing sent "
    "anywhere').",
    "Keep a negation only for a genuine wrong turn the reader would really take, and walk in "
    "before springing it: 'A lender cares less about what you owe than about what you pay "
    "out each month.'",
    "Numbers arrive with their meaning attached: 'every £100 a month going to loans takes "
    "roughly £5,000 to £7,000 off what a lender will offer you' — never a bare figure the "
    "reader has to interpret alone.",
    "Compliance and legal lines are exempt from this voice: they follow the site's "
    "compliance rules and the chrome carrier.",
]

# Worked examples. A writer model follows exemplars more reliably than rules — the rules
# explain the register, the pairs teach it. These are loancalculator's OWN copy, not
# another site's, per VOICE_gentle_explanatory_v1 step 2.
EXEMPLARS = {
    "register_anchor": (
        "Most of us have more than one kind of borrowing. Each kind affects the others. If "
        "you take on a car loan, a mortgage lender will usually offer you less. If you "
        "remortgage, the cost of your other borrowing can shift too. The calculators on this "
        "site are free, and they're built to show you those connections. Your numbers stay "
        "on your own screen."
    ),
    "before_after_1_consolidation": (
        "BEFORE: 'Consolidating can lower your monthly payments, but extending the term "
        "might cost you more overall. Enter your current debts below to compare the true "
        "total cost.'  AFTER: 'If you're thinking about rolling several debts into one loan, "
        "the appeal is usually the monthly payment, which often comes down. What's easy to "
        "miss is the time. A new loan usually runs longer, and more months of interest can "
        "cost more in total even when each month feels lighter. Put your current debts in "
        "below and this checker shows you both sides.'"
    ),
    "before_after_2_car_finance": (
        "BEFORE: 'Understand the real cost of your car finance. Hire Purchase (HP) leads to "
        "ownership; Personal Contract Purchase (PCP) keeps payments low but carries a final "
        "Balloon.'  AFTER: 'If you're looking at car finance, you'll usually be offered one "
        "of two things. With hire purchase, you pay the car off month by month and it ends "
        "up yours. With a personal contract purchase, the monthly payments are smaller, and "
        "a large final payment is left at the end. Dealers call that final payment the "
        "balloon. This calculator shows you what each route really costs over the life of "
        "the deal.'"
    ),
    "before_after_3_overpayment": (
        "BEFORE: 'Overpaying your loan reduces the total interest you pay. However, some "
        "lenders charge an Early Repayment Charge (ERC).'  AFTER: 'If you have spare money "
        "each month, putting it against a loan usually saves you interest, because interest "
        "is charged on what you still owe. There's one thing to check first. Some lenders "
        "charge a fee for paying off early, and on a personal loan that's usually up to 58 "
        "days of interest. This calculator shows the saving and the fee side by side.'"
    ),
}

# ── read current ─────────────────────────────────────────────────────────────────
raw = psql(f"""SELECT sp.data::text FROM site_specs sp JOIN sites s ON s.id=sp.site_id
WHERE s.domain='{DOMAIN}' AND sp.aspect='content_direction' AND sp.is_current;""")
if not raw:
    sys.exit(f"no current content_direction for {DOMAIN}")
cd = json.loads(raw)
stored = cd.get("formatted", "")

# ── GATE: the port must reproduce the stored `formatted` before we trust it ───────
regen = format_cd({k: v for k, v in cd.items() if k != "formatted"})
if sorted(regen.split("\n")) != sorted(stored.split("\n")):
    sys.exit(
        "  FORMATTER GATE FAILED — the Python port no longer reproduces the stored\n"
        "  `formatted` field. Regenerating it would silently change what the writer\n"
        f"  sees. Port produced {len(regen)} bytes / {len(regen.splitlines())} lines;\n"
        f"  stored is {len(stored)} bytes / {len(stored.splitlines())} lines.\n"
        "  Re-port datahelpers.FormatContentDirection before using this script."
    )
print(f"  formatter gate PASSES — reproduces stored `formatted` exactly "
      f"({len(stored)} bytes, {len(stored.splitlines())} lines)")

# ── build the new spec ───────────────────────────────────────────────────────────
new_cd = json.loads(raw)
changes = []

voice = dict(new_cd.get("voice") or {})
for k, v in VOICE_REPLACE.items():
    if voice.get(k) != v:
        changes.append(f"voice.{k}  REPLACED")
    voice[k] = v
new_cd["voice"] = voice

rules = list(new_cd.get("writing_rules") or [])
out_rules, replaced = [], set()
for r in rules:
    hit = next((k for k in RULES_REPLACE if k in r), None)
    if hit:
        replaced.add(hit)
        if RULES_REPLACE[hit] is not None:
            out_rules.append(RULES_REPLACE[hit])
            changes.append(f"writing_rules  REPLACED  «{hit[:48]}…»")
        else:
            changes.append(f"writing_rules  DELETED   «{hit[:48]}…»")
    else:
        out_rules.append(r)
for k in RULES_REPLACE:
    if k not in replaced:
        sys.exit(f"  REFUSING: expected to replace a live rule matching «{k[:60]}» "
                 f"and found none. The spec has changed since this script was written — "
                 f"re-read it before seeding, or the conflict it exists to resolve is "
                 f"still live under different wording.")
for r in RULES_ADD:
    if not any(r[:40] in x for x in out_rules):
        out_rules.append(r)
        changes.append(f"writing_rules  ADDED     «{r[:48]}…»")
new_cd["writing_rules"] = out_rules

if new_cd.get("voice_exemplars") != EXEMPLARS:
    new_cd["voice_exemplars"] = EXEMPLARS
    changes.append("voice_exemplars  SET (4 worked examples, this site's own copy)")

new_cd["formatted"] = format_cd({k: v for k, v in new_cd.items() if k != "formatted"})

print(f"\n  {DOMAIN}")
for c in changes:
    print(f"    {c}")
print(f"\n  writing_rules  {len(rules)} -> {len(out_rules)}")
print(f"  content_direction.formatted  {len(stored)} -> {len(new_cd['formatted'])} bytes "
      f"({len(new_cd['formatted']) - len(stored):+d})")

# Contradiction check: the retired rule must not survive anywhere in what the writer sees.
if "Avoid contractions in declarative" in new_cd["formatted"]:
    sys.exit("  REFUSING: the retired no-contractions rule is still present in `formatted`.")
print("  contradiction check PASSES — the retired no-contractions rule is gone from "
      "`formatted`")

if not APPLY:
    print("\n  DRY RUN — rerun with --apply to write.")
    sys.exit(0)

payload = json.dumps(new_cd).replace("'", "''")
psql(f"""
BEGIN;
UPDATE site_specs SET is_current = false, superseded_at = NOW()
 WHERE site_id = (SELECT id FROM sites WHERE domain='{DOMAIN}')
   AND aspect='content_direction' AND is_current;
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
VALUES ((SELECT id FROM sites WHERE domain='{DOMAIN}'), 'content_direction',
        '{payload}'::jsonb, 'operator', 'seed_voice_h.py',
        'Gentle-explanatory (H) voice, owner choice 2026-08-05. Conflicting incumbent rules REPLACED not appended (chiefly no-contractions-in-declaratives). formatted regenerated by the gated port of datahelpers.FormatContentDirection.',
        true, 'cqls');
COMMIT;""", tuples=False)
print("\n  APPLIED.")
print(psql(f"""SELECT '  now: aspect='||aspect||' bytes='||length(data::text)
||' formatted_bytes='||COALESCE(length(data->>'formatted')::text,'-')
||' has_H='||(data->>'formatted' LIKE '%where the reader is standing%')::text
FROM site_specs sp WHERE site_id=(SELECT id FROM sites WHERE domain='{DOMAIN}')
AND aspect='content_direction' AND is_current;"""))
