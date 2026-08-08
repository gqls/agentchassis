#!/usr/bin/env python3
"""trim_voice_rules.py — remove the PRESCRIPTIONS from loancalculator's voice spec,
keep the prohibitions. Owner question, 2026-08-08: "it's all too mechanical, either we
extend the rules further or remove some rules?"

THE ARGUMENT THIS TESTS. A prohibition applied to 100 sections still permits 100
different openings. A PRESCRIPTION applied to 100 sections produces 100 openings of the
same shape — it is a template wearing a rule's clothes, and a template everywhere is
what "mechanical" means. Every fault the owner named came from a prescription:

  * the "If you're…" tic on nearly every page  -> rules 7 AND 3, which independently
    mandate the same opening shape
  * "your numbers stay on your own screen" on a homepage about money -> rule 20, verbatim

And the decisive evidence that ADDING rules cannot fix this: rule 23 already exempts
compliance/legal lines from the voice, and `/legal.html` still opens "If you're using
the calculators…". A rule we already have did not fire. So a new rule to fix the legal
page would be a second copy of a rule that is already being ignored.

FOUR EDITS, ALL REMOVALS OR DEMOTIONS. Nothing is added.

  rule 7   demote to its PROHIBITION half. Keep "never open cold with a bare assertion,
           never with a negative twist" — the thing it was written FOR. Drop the mandate
           to open with a conditional clause, which is the tic.
  rule 3   DELETE. It restates rule 7's mandate in different words, which is why the
           opening shape is so relentless: two rules pushing one template.
  rule 20  demote to its PROHIBITION half ("never a negation pile"). Drop "state facts
           positively, including privacy… 'your numbers stay on your own screen'" — a
           rule about HOW to phrase something silently authorised WHETHER to include it.
  rule 2   relax the 2-4 sentence paragraph cap to guidance. Rule 17 ("one idea, one
           sentence") already governs clarity; together the two enforce a uniform
           texture that nothing in the brief ever asked for.

THE GATE. The writer reads exactly ONE field, `content_direction.formatted`, produced by
datahelpers.FormatContentDirection. This ports that function and, before writing,
regenerates the CURRENT spec and asserts it reproduces the STORED value as a multiset of
lines. If the port has drifted from the Go, it refuses rather than silently changing what
the writer sees. Same gate as seed_voice_h.py, which is where the port comes from.

Backup taken before any run: site_specs_bak_20260808_ruletrim.

Run:  python3 trim_voice_rules.py            # dry run, shows the diff
      python3 trim_voice_rules.py --apply    # write it
      python3 trim_voice_rules.py --revert   # restore from the backup table
"""
import json
import subprocess
import sys

APPLY = "--apply" in sys.argv
REVERT = "--revert" in sys.argv
DOMAIN = "loancalculator.co.uk"
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db"]


def psql(sql, tuples=True):
    cmd = PSQL + (["-t", "-A"] if tuples else [])
    r = subprocess.run(cmd + ["-c", sql], capture_output=True, text=True, timeout=180)
    if r.returncode != 0:
        sys.exit(f"psql failed:\n{r.stderr.strip()}")
    return r.stdout.strip()


# ── faithful port of datahelpers.FormatContentDirection (from seed_voice_h.py) ────
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


if REVERT:
    psql("""
BEGIN;
UPDATE site_specs SET is_current=false, superseded_at=NOW()
 WHERE site_id=(SELECT id FROM sites WHERE domain='loancalculator.co.uk')
   AND aspect='content_direction' AND is_current;
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT site_id, aspect, data, source, source_agent,
       'REVERTED to the pre-trim spec from site_specs_bak_20260808_ruletrim', true, 'cqls'
FROM site_specs_bak_20260808_ruletrim;
COMMIT;""", tuples=False)
    print("  REVERTED from site_specs_bak_20260808_ruletrim.")
    sys.exit(0)

# key = distinctive substring of the LIVE rule; value = replacement, or None to delete
EDITS = {
    "Open sections where the reader is standing": (
        "Never open a section cold with a bare assertion, and never with a negative twist "
        "('X isn't about Y…'). Beyond that, vary how sections open: the right opening "
        "depends on what the section has to say, and a page whose sections all begin the "
        "same way reads as a template rather than as writing."
    ),
    "Lead every guide section with the practical bottom line": None,
    "State facts positively, including privacy and cost": (
        "Never write a negation pile ('no sign-up, no credit check, nothing sent anywhere'). "
        "Where a fact about cost or privacy genuinely belongs, state it positively — but "
        "whether it belongs at all is a judgement about what the reader came for, not "
        "something this rule settles."
    ),
    "Paragraphs are 2-4 sentences maximum": (
        "Keep paragraphs to one idea. Vary their length: some ideas need a sentence and "
        "some need five, and a page of identically-sized paragraphs reads as mechanical "
        "however good each one is."
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
    sys.exit("  FORMATTER GATE FAILED — the port no longer reproduces the stored "
             "`formatted`. Re-port datahelpers.FormatContentDirection before using this.")
print(f"  formatter gate PASSES ({len(stored)} bytes, {len(stored.splitlines())} lines)")

new_cd = json.loads(raw)
rules = list(new_cd.get("writing_rules") or [])
out, hit, changes = [], set(), []
for r in rules:
    k = next((k for k in EDITS if k in r), None)
    if k:
        hit.add(k)
        if EDITS[k] is None:
            changes.append(f"DELETED   «{k[:52]}…»")
        else:
            out.append(EDITS[k])
            changes.append(f"DEMOTED   «{k[:52]}…»  (prescription -> prohibition)")
    else:
        out.append(r)
missing = [k for k in EDITS if k not in hit]
if missing:
    sys.exit("  REFUSING: expected to edit rules matching, and found none:\n    " +
             "\n    ".join(missing) +
             "\n  The spec changed since this script was written — re-read it.")
new_cd["writing_rules"] = out
new_cd["formatted"] = format_cd({k: v for k, v in new_cd.items() if k != "formatted"})

print(f"\n  {DOMAIN}")
for c in changes:
    print(f"    {c}")
print(f"\n  writing_rules  {len(rules)} -> {len(out)}")
print(f"  formatted      {len(stored)} -> {len(new_cd['formatted'])} bytes "
      f"({len(new_cd['formatted']) - len(stored):+d})")

# The tic-generating mandate must be gone from what the writer actually sees.
for gone in ("begin with a conditional or situational clause",
             "your numbers stay on your own screen"):
    if gone in new_cd["formatted"]:
        sys.exit(f"  REFUSING: «{gone}» still present in `formatted`.")
print("  removal check PASSES — neither the conditional-opening mandate nor the "
      "privacy phrasing survives in `formatted`")

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
        '{payload}'::jsonb, 'operator', 'trim_voice_rules.py',
        'Rule trim trial (owner 2026-08-08): 2 prescriptions demoted to prohibitions, 1 deleted as a duplicate mandate, 1 relaxed. Tests whether the mechanical feel comes from prescriptive rules applied uniformly. Revert: trim_voice_rules.py --revert',
        true, 'cqls');
COMMIT;""", tuples=False)
print("\n  APPLIED.")
print(psql(f"""SELECT '  now: rules='||jsonb_array_length(data->'writing_rules')
||' formatted_bytes='||length(data->>'formatted')
||' tic_mandate_present='||(data->>'formatted' LIKE '%conditional or situational clause%')::text
FROM site_specs WHERE site_id=(SELECT id FROM sites WHERE domain='{DOMAIN}')
AND aspect='content_direction' AND is_current;"""))
