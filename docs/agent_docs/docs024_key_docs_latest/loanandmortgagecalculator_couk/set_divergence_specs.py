#!/usr/bin/env python3
"""set_divergence_specs.py — record the two sites' divergent positioning as FRAMEWORK
CONFIG, in the only places the content pipeline actually reads.

THE BRIEF: "I'd like them both to evolve in different directions toward different target
markets to avoid duplicate penalties managed by the framework." The operative word is
*managed* — the divergence has to be configuration the pipeline reads, not a note in a
document that a future generation run will never see.

WHERE IT CAN AND CANNOT GO. Measured against the live database and every agent
definition, not assumed:

  * `identity.target_audience`  -> READ. Interpolated into page-content-writer's prompt
    as `Target Audience: {{.site_specs.specs.identity.target_audience}}`. `identity` is
    the most-read aspect in the system (13 agents). Present on 15/15 sites that have one.
  * `identity.key_differentiators` -> READ, same prompt block.
  * `content_direction.formatted` -> READ. **The writer reads exactly ONE field of
    content_direction, `.formatted`.** Every other key reaches the prompt only by being
    serialised INTO `formatted`. A hand-written content_direction that forgets to
    regenerate `formatted` is INVISIBLE to the writer: the edit looks applied and
    changes nothing.
  * `audience` -> **NOT READ BY ANYTHING.** It is the single most widely-populated
    aspect in the database (29 of 33 sites) and no agent, prompt or Go path consumes it.
    An earlier version of this workstream's plan named `audience` as one of three places
    to record the divergence; that third of the work would have produced a row that
    could not affect a single generated word. Same for `editorial`, `voice`,
    `content_standards`, `terminology_and_positioning` — all invented ad hoc by the
    gap-planner's unconstrained `aspect` write (apply_gap_plan_action.go:744).

So: new keys go INSIDE `content_direction` (where `formatted` picks them up), never as
new aspects. That is the pattern dartsonline.com used for `editorial` / `honesty_rails`.

`formatted` IS REGENERATED HERE, by a faithful reimplementation of
datahelpers.FormatContentDirection. The reimplementation is not trusted — it is GATED:
before writing anything, it regenerates the CURRENT spec and asserts the result matches
the stored `formatted` as a multiset of lines. (Not as a string: Go map iteration order
is random, so section and sub-key order in the stored value is arbitrary and carries no
meaning. Verified equal at 143 lines / 18,005 bytes.)

ASYMMETRY, DELIBERATE (decision D6). Only this site can hold specs.
mortgagecalculator.co.uk is a hand-built static tree with no `sites` row, and the owner
chose to leave it untouched. Its narrow-authority position is recorded in the workstream
docs as stated intent, not as configuration. Flagged rather than hidden.

AND THE LIMITATION WORTH STATING: there is **no cross-site duplicate-content or
topical-overlap machinery in this platform at all.** check_content_duplication is
single-site (WHERE p.site_id = $1), remove_duplicate_page_sections is single-page, and
cross_site_contamination detects another site's company_name bleeding into rendered HTML
— not topical overlap. So this spec is the entire mechanism, and nothing will warn us if
the two sites converge again.

Run:  python3 set_divergence_specs.py            # show what would change
      python3 set_divergence_specs.py --apply    # write it
"""
import json
import subprocess
import sys

APPLY = "--apply" in sys.argv
_dom_args = [a for a in sys.argv[1:] if not a.startswith("--")]
DOMAIN = _dom_args[0] if _dom_args else "loanandmortgagecalculator.co.uk"
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


# ── the positioning itself, per REGISTERED domain ────────────────────────────────
# One payload per register entry. Adding a domain here means its register row exists
# FIRST (the portfolio rule) — the payload restates the row, it never invents one.
PAYLOADS = {}

PAYLOADS["loancash.co.uk"] = {  # register entry L10 (P6, owner ruling 2026-08-01)
    "target_audience": (
        "The high-cost, small-sum, urgent UK borrower: someone looking at or already "
        "holding a payday-style loan, doorstep loan, rent-to-own agreement, logbook or "
        "guarantor loan — often about to borrow within days, or already in difficulty. "
        "This site serves them as a rights-holder, not a customer: it is not a lender, "
        "not a broker, and takes no applications. The reader who wants to CHOOSE between "
        "mainstream loan products is served by a different site; this one exists for "
        "'what are they allowed to do to me' and 'how do I make the rules bite'."
    ),
    "key_differentiators": [
        "A civilian champion of the FCA's consumer-credit rulebook: the price cap, "
        "affordability duties, CPA and rollover limits, complaint rights, "
        "authorised-lender checks and loan-shark reporting, each explained as a right "
        "with the exact next step.",
        "Rule-checking tools rather than product calculators: was this loan within the "
        "price cap, what is the legal maximum cost, when do the complaint deadlines "
        "fall.",
        "Explicitly independent of the FCA, stated on every page — championing the "
        "rules without ever appearing to be the regulator.",
        "Regulatory constants are quoted with the rule they come from; market rates are "
        "never quoted. No applications, no lead-gen, and free debt help signposted from "
        "every page.",
    ],
    "positioning": {
        "site_role": (
            "The borrower's guide to the FCA rulebook for high-cost credit. Its subject "
            "is the protections, not the products: what lenders must do, may not do, "
            "and what the borrower can claim when the rules are broken."
        ),
        "in_scope": (
            "The high-cost short-term credit price cap and its arithmetic; "
            "creditworthiness and affordability duties; continuous payment authority "
            "limits and cancellation; rollover limits; complaints to the lender and the "
            "Financial Ombudsman with their deadlines; forbearance rights and Breathing "
            "Space; authorised-lender checks and the consequences of unauthorised "
            "lending; loan-shark reporting; cheaper small-sum routes as protective "
            "signposting."
        ),
        "out_of_scope": (
            "Choosing between mainstream loan products (whichloan territory), debt "
            "consolidation depth (consolidateloans territory), adverse-credit mortgages "
            "(adversecreditmortgage territory), and anything that functions as a loan "
            "application or lead. This site never recommends a lender."
        ),
        "divergence_rule": (
            "If the visitor's question is 'which product should I get', link out to the "
            "choosing sites. If it is 'what are they allowed to do to me' or 'how do I "
            "get money back', it lives here. Never quote a market rate; quote regulatory "
            "constants only, each with the rule it comes from. The protective register "
            "is the identity: zero judgement, actionable answer first, free-help routes "
            "always visible."
        ),
    },
}

TARGET_AUDIENCE = (
    "UK borrowers whose unsecured debt and their mortgage interact: someone with a "
    "personal loan, a credit card balance or a car finance agreement who is applying "
    "for a mortgage, remortgaging, or deciding which debt to clear first. Not the "
    "single-subject researcher. A visitor who only wants a mortgage repayment figure, "
    "or only wants to compare two personal loans, is served better by a single-subject "
    "site; this site exists for the decisions that span both kinds of borrowing and "
    "that no single-subject site can answer."
)

KEY_DIFFERENTIATORS = [
    "Covers the crossing points between unsecured borrowing and a mortgage — how a car "
    "finance payment reduces what a lender will offer, whether to consolidate debt into "
    "a remortgage, whether the next £1,000 should go on the deposit or clear a loan.",
    "Holds both toolsets on one domain, so a question that spans them can be answered "
    "without sending the reader to a second site.",
    "Explains mechanisms rather than quoting current rates or tax bands, so the guidance "
    "does not silently go stale.",
    "Every calculator runs entirely in the browser: no sign-up, no credit check, and "
    "nothing transmitted anywhere.",
]

# New keys go INSIDE content_direction so `formatted` carries them to the writer.
POSITIONING = {
    "site_role": (
        "This is the whole-borrowing-picture site. Its subject is the interaction "
        "between unsecured borrowing and a mortgage, not either one in isolation."
    ),
    "in_scope": (
        "Questions that require both halves to answer: borrowing power against existing "
        "commitments, consolidation into secured debt, deposit versus debt repayment, "
        "credit file preparation before an application, total cost across a whole "
        "household budget, and what happens to both when rates move."
    ),
    "out_of_scope": (
        "Single-subject depth for its own sake. Do not write the definitive standalone "
        "guide to a mortgage product or to a loan product. If a subject can be fully "
        "answered without reference to the other kind of borrowing, it belongs on a "
        "single-subject site, and this site should cover only the part where it touches "
        "the other half."
    ),
    "divergence_rule": (
        "Sibling sites mortgagecalculator.co.uk (mortgages only) and loancalculator.co.uk "
        "(loans only) cover the single-subject material. This site must not converge on "
        "either. When a topic could be written either as single-subject explanation or as "
        "a crossing-point question, always choose the crossing-point framing — that "
        "framing is the site's reason to exist and the thing that keeps the three sites "
        "from competing with each other for the same search results."
    ),
}

# The original domain's payload, registered alongside the newer entries.
PAYLOADS["loanandmortgagecalculator.co.uk"] = {
    "target_audience": TARGET_AUDIENCE,
    "key_differentiators": KEY_DIFFERENTIATORS,
    "positioning": POSITIONING,
}

# lendzy.co.uk — the L10 BENCHMARK payload applied to the SHADOW domain (owner
# direction 2026-08-02; SUMMARY_2026-08-02_the_pipeline_should_build_everything.md).
# Deliberately NOT a register entry, which the rule above would otherwise require: the
# shadow build tests whether the pipeline can match hand-built loancash.co.uk given the
# SAME positioning input, so the payload is loancash's L10 by reference, not a copy that
# could drift. One addition: an acceptance MARKER the writer is instructed to include,
# so "the spec reaches the writer's prompt" (task #16, bug025 pattern) is proven in
# passing — grep page_components for the exact phrase, baseline-zero first, and never
# grep site_specs for it (this row itself carries the phrase).
# If lendzy.co.uk is ever built FOR REAL: write its own register row first, replace this.
_l10 = PAYLOADS["loancash.co.uk"]
PAYLOADS["lendzy.co.uk"] = {
    "target_audience": _l10["target_audience"],
    "key_differentiators": list(_l10["key_differentiators"]),
    "positioning": {
        **_l10["positioning"],
        "acceptance_marker": (
            "Somewhere in the site's written copy include the exact phrase: "
            "checked against the FCA handbook, rule by rule."
        ),
    },
}

if DOMAIN not in PAYLOADS:
    sys.exit(f"no positioning payload for {DOMAIN} — write its register row first, then "
             f"add the payload here restating it. Known: {', '.join(sorted(PAYLOADS))}")
_p = PAYLOADS[DOMAIN]
TARGET_AUDIENCE, KEY_DIFFERENTIATORS, POSITIONING = (
    _p["target_audience"], _p["key_differentiators"], _p["positioning"])

# ── load current state ───────────────────────────────────────────────────────────
site_id = psql(f"SELECT id FROM sites WHERE domain = '{DOMAIN}';")
if not site_id:
    sys.exit(f"no sites row for {DOMAIN} — adopt it first")

identity = json.loads(psql(
    f"SELECT data::text FROM site_specs WHERE site_id='{site_id}' "
    f"AND aspect='identity' AND is_current;"))
cd = json.loads(psql(
    f"SELECT data::text FROM site_specs WHERE site_id='{site_id}' "
    f"AND aspect='content_direction' AND is_current;"))

# ── GATE: prove the formatter reimplementation before trusting it ────────────────
regen, stored = format_cd(cd), cd.get("formatted", "")
if sorted(regen.splitlines()) != sorted(stored.splitlines()):
    sys.exit("ABORT: reimplemented FormatContentDirection does not reproduce the stored\n"
             "  `formatted` field. Regenerating it would silently change what the writer\n"
             "  reads. Re-port datahelpers/format_content_direction.go before continuing.")
print(f"  formatter gate PASSES — reproduces stored `formatted` exactly "
      f"({len(stored.splitlines())} lines, {len(stored)} bytes)")

# ── build the new specs ──────────────────────────────────────────────────────────
new_identity = dict(identity)
new_identity["target_audience"] = TARGET_AUDIENCE
new_identity["key_differentiators"] = KEY_DIFFERENTIATORS

new_cd = dict(cd)
new_cd["positioning"] = POSITIONING
new_cd["formatted"] = format_cd({k: v for k, v in new_cd.items() if k != "formatted"})

print(f"\n  identity.target_audience     {len(identity.get('target_audience',''))}"
      f" -> {len(TARGET_AUDIENCE)} chars")
print(f"  identity.key_differentiators {'absent' if 'key_differentiators' not in identity else 'present'}"
      f" -> {len(KEY_DIFFERENTIATORS)} entries")
print(f"  content_direction keys       {len(cd)} -> {len(new_cd)} (+positioning)")
print(f"  content_direction.formatted  {len(stored)} -> {len(new_cd['formatted'])} bytes")
print(f"    (+{len(new_cd['formatted']) - len(stored)} bytes, which is the positioning "
      f"block reaching the writer)")

if not APPLY:
    print("\n  dry run. Re-run with --apply to write.")
    sys.exit(0)


def lit(obj):
    return "'" + json.dumps(obj).replace("'", "''") + "'::jsonb"


# idx_site_specs_current is UNIQUE (site_id, aspect) WHERE is_current — so the supersede
# must precede the insert. Both in one transaction: after the UPDATE the old row is no
# longer in the partial index, so the INSERT does not collide.
sql = f"""BEGIN;
-- keep the pre-change values recoverable; the aspect history is the audit trail
UPDATE site_specs SET is_current = false, superseded_at = NOW()
 WHERE site_id = '{site_id}' AND aspect IN ('identity','content_direction') AND is_current;
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
VALUES ('{site_id}', 'identity', {lit(new_identity)}, 'manual', NULL,
        'Register-driven positioning ({DOMAIN}): target audience + differentiators per the portfolio register entry.', true, 'cqls'),
       ('{site_id}', 'content_direction', {lit(new_cd)}, 'manual', NULL,
        'Register-driven positioning ({DOMAIN}): positioning block inside content_direction so `formatted` carries it to page-content-writer; formatted regenerated by the gated port of datahelpers.FormatContentDirection.', true, 'cqls');
COMMIT;
"""
out = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1", "--echo-errors"],
                     input=sql, capture_output=True, text=True, timeout=180)
if out.returncode != 0:
    sys.exit(f"write FAILED (rolled back):\n{out.stderr.strip()}")

# Verify by reading back, not by trusting the exit code.
check = psql(f"""
SELECT aspect || ' current=' || is_current
       || ' has_positioning=' || (data ? 'positioning')
       || ' formatted_bytes=' || COALESCE(length(data->>'formatted')::text,'-')
       || ' audience_chars=' || COALESCE(length(data->>'target_audience')::text,'-')
FROM site_specs WHERE site_id='{site_id}'
  AND aspect IN ('identity','content_direction') AND is_current ORDER BY aspect;""")
print("\n  written and read back:")
for line in check.splitlines():
    print("   ", line)
print("\n  NOTE: this changes what the writer is TOLD, not what is already on the page.")
print("  The 13 guides are still rebuild_policy='owned' verbatim, so nothing regenerates")
print("  them yet. Proving the spec reaches a prompt needs the bug025 acceptance-test")
print("  pattern — a greppable marker driven through one page — and that is the next step.")
