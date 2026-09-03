#!/usr/bin/env python3
"""verify_criteria.py — recompute every value about to be PINNED into a tool
PLAN, from a definition that is not this page's script.

WHY THIS EXISTS AT ALL. `toolgolden.py --emit-criteria` records what the tool
currently prints. That is a *consistency* record, and on its own it is the F3
failure PLAN_2026-08-09 §0 names: "a golden captured from an already-wrong tool
pins the wrong answer and then defends it" — the browser runner's own source
says so at run_checks_action.go:775-781. This estate has already shipped that
mistake twice (bugs_open/224, bugs_open/225: an expired FTB cap certified green
for sixteen months). So nothing goes into `doc_plans` here until it has been
re-derived from somewhere else.

THE TWO SOURCES OF "SOMEWHERE ELSE", and they are NOT equally strong — the
distinction is reported per assertion and must not be flattened:

  DEFINITION  the published formula. The annuity identity, compound interest,
              amortisation run month by month. `oracles.py` (the neighbouring
              loanandmortgagecalculator lane) is reused rather than re-written:
              it was authored from the definitions, never from a page, and
              re-writing it here would only give us a second thing to keep
              right. Reuse over rebuild, per the platform conventions.

  REGISTER    for stamp-duty ONLY, and this is the point of this whole lane:
              the bands are built from THIS SITE'S 13 registered SDLT facts —
              each a scalar carrying its own verbatim GOV.UK quote, re-verified
              daily by the evidence-freshness sweep. Not from oracles.py's
              hard-coded band table, which would be a second hand-typed copy of
              the same law. If the register and the tool disagree, one of them
              is wrong and a human is told which vector shows it.

  CONVENTION  a rule read off the tool because it is the TOOL'S design choice,
              not a published fact — rate-forecaster's 24/36-month phase split,
              fee-analyser's definition of "total cost". These assertions are
              WEAKER evidence and are labelled so in the output. They still
              catch a rewrite that changes the arithmetic; they cannot catch a
              tool whose convention was wrong to begin with. The neighbouring
              lane learned this the expensive way and wrote it down: its first
              oracle asserted the naive reprice model, reported 4 FAILs against
              a CORRECT page, and had to be corrected (oracles.py:108-138).

A tool not modelled here is reported as NOT VERIFIED. That is the honest state,
and it is why `portfolio` and `equity-release`'s max-cash line are named in the
output rather than quietly omitted.

INDUCE THE RED BEFORE TRUSTING THE GREEN. `--mutate <fact-id>=<value>` perturbs
one registered fact in memory and re-runs. It exists because 80 agreements prove
nothing about whether the register is actually load-bearing here: a script that
read the wrong column, or fell back to oracles.py's hard-coded bands, would
print the same 80. The mutation is the only run that can distinguish them, and
PLAN_2026-08-09 §7 requires it before any of this reaches `doc_plans`.

Usage:
    python3 verify_criteria.py                 # every emitted criteria file
    python3 verify_criteria.py stamp-duty      # one tool
    python3 verify_criteria.py --mutate sdlt-ftb-relief-cap=625000   # must FAIL
"""
import glob
import json
import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
LANE = os.path.abspath(os.path.join(HERE, "..", "..", "loanandmortgagecalculator_couk"))
sys.path.insert(0, LANE)
import oracles  # noqa: E402

CRIT = os.path.join(HERE, "criteria")
DOMAIN = "mortgagecalculator.co.uk"
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db"]

# Pence, except where a tool displays whole pounds (named per tool rather than
# loosened globally, which would blind the pence-accurate ones).
TOL = 0.02
ROUNDS_TO_POUND = {"stamp-duty", "overpayment", "equity-release"}


# Unicode dashes a page may render instead of ASCII '-'. Written as code points
# rather than literals on purpose: this channel rewrites some non-ASCII on
# emission, so a retyped '−' is a character nobody can verify by reading.
_DASHES = {ord(chr(cp)): "-" for cp in (0x2212, 0x2013, 0x2014, 0x2010, 0x2011)}


def num(s):
    """Parse a displayed money/number string.

    TWO traps here, and the second one convicted this file on its first run.

    (1) rate-forecaster renders a fall with U+2212 MINUS SIGN, not ASCII '-'.

    (2) The sign is OUTSIDE the currency symbol: '<minus>£961.61'. So a
        `re.search(r"-?\\d...")` — which requires the sign to be adjacent to the
        first digit — matches at the '9' and returns +961.61. That is how the
        first run of this script reported the tool as wrong by £1,923.22 when
        the tool was right: a checker bug wearing the costume of the defect it
        was written to find. STRIP the noise, do not scan past it.
    """
    s = s.translate(_DASHES)
    s = re.sub(r"[^0-9.\-]", "", s)
    if not re.search(r"\d", s):
        raise ValueError("no number in %r" % s)
    return float(s)


def driven(check):
    """{selector: value} for every driven input — `select` counts, not just
    `fill`. Collecting only fills silently dropped stamp-duty's #buyerType in
    the neighbouring lane and graded an FTB vector as a standard buyer, then
    reported the resulting 5,000 gap as a defect in a correct tool."""
    return {s["selector"]: s.get("value") for s in check["steps"]
            if s["action"] in ("fill", "select")}


# ── the register: SDLT bands built from the site's own registered facts ─────

def load_register_bands():
    """Build the SDLT band table from this site's 13 registered facts.

    Every number here arrives as a scalar `value` off a fact that carries its
    own verbatim GOV.UK quote and is re-fetched daily. Nothing is typed in this
    file. A1 (handoff 10b §1c) re-shaped the register precisely so this
    function could exist: while `sdlt-standard-bands` was one fact of `value:
    12` with the bands in prose, no oracle could read a band table at all.
    """
    sql = ("SELECT f->>'id', f->>'value' FROM site_specs ss "
           "JOIN sites s ON s.id=ss.site_id, jsonb_array_elements(ss.data->'facts') f "
           "WHERE s.domain='%s' AND ss.aspect='evidence_base' AND ss.is_current;" % DOMAIN)
    r = subprocess.run(PSQL + ["-tA", "-F", "\t", "-v", "ON_ERROR_STOP=1", "-c", sql],
                       capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit((r.stderr or r.stdout).strip()[:600])
    facts = {}
    for line in r.stdout.strip().splitlines():
        # A fact whose `value` is JSON null makes `->>` return SQL NULL, and
        # psql -tA prints that row with NO trailing separator at all — so this
        # MUST tolerate a one-field line. `split("\t")` unpacked into two names
        # and raised "not enough values to unpack" on exactly such a row
        # (CIT-8a9ee77ad646fd10, 2026-09-03).
        parts = line.split("\t", 1)
        k, v = parts[0], (parts[1] if len(parts) > 1 else "")
        # Only SCALAR facts belong in the band table. The register also carries
        # value-less kinds — as of 2026-09-03 this site holds 5 `CIT-*` citation
        # facts with an empty `value`, and float("") raised ValueError here,
        # which killed install_fences.py outright (it calls this at import time).
        # Skipping a non-numeric is SAFE and does not weaken the guarantee
        # below: any of the 13 `need` keys that failed to parse is simply absent
        # from `facts`, lands in `missing`, and still stops the script. A silent
        # fall-through to oracles.py's hard-coded bands remains impossible.
        try:
            facts[k] = float(v)
        except ValueError:
            continue
    need = ["sdlt-standard-nil-band-upper", "sdlt-standard-band-250k-upper",
            "sdlt-standard-band-925k-upper", "sdlt-standard-band-1500k-upper",
            "sdlt-standard-rate-125k-250k", "sdlt-standard-rate-250k-925k",
            "sdlt-standard-rate-925k-1500k", "sdlt-standard-top-rate",
            "sdlt-ftb-nil-band-upper", "sdlt-ftb-rate-300k-500k",
            "sdlt-ftb-relief-cap", "sdlt-additional-surcharge",
            "sdlt-additional-surcharge-floor"]
    missing = [k for k in need if k not in facts]
    if missing:
        # SKIPPED IS NOT PASSED (PLAN §5.5): a register that cannot supply the
        # bands must stop this script, never fall through to oracles.py's
        # hard-coded copy — that would report the register as verified while
        # reading a different source entirely.
        sys.exit("register incomplete, refusing to verify stamp-duty: missing %s"
                 % ", ".join(missing))
    return facts


def sdlt_from_register(price, buyer, f):
    """SDLT for England & NI, computed from the registered scalars alone."""
    std = [(f["sdlt-standard-nil-band-upper"], 0.0),
           (f["sdlt-standard-band-250k-upper"], f["sdlt-standard-rate-125k-250k"]),
           (f["sdlt-standard-band-925k-upper"], f["sdlt-standard-rate-250k-925k"]),
           (f["sdlt-standard-band-1500k-upper"], f["sdlt-standard-rate-925k-1500k"]),
           (float("inf"), f["sdlt-standard-top-rate"])]

    def banded(bands, surcharge=0.0):
        due, lower = 0.0, 0.0
        for upper, rate in bands:
            if price > lower:
                due += (min(price, upper) - lower) * (rate + surcharge) / 100.0
            lower = upper
        return due

    if buyer == "ftb":
        # "If the purchase price is more than £500,000, you cannot claim the
        # relief and you must pay the standard rates on the total purchase
        # price" — sdlt-ftb-relief-cap's own claim.
        if price > f["sdlt-ftb-relief-cap"]:
            return banded(std)
        return banded([(f["sdlt-ftb-nil-band-upper"], 0.0),
                       (f["sdlt-ftb-relief-cap"], f["sdlt-ftb-rate-300k-500k"])])
    if buyer == "additional":
        # The surcharge applies only at or above the floor; below it the
        # standard rates stand. This is the £40,000 floor the 08-10 rebuild
        # DELETED because the register did not yet carry it (handoff 10b §1b) —
        # which is exactly why it is asserted here now.
        if price >= f["sdlt-additional-surcharge-floor"]:
            return banded(std, f["sdlt-additional-surcharge"])
        return banded(std)
    return banded(std)


# ── per-tool models ─────────────────────────────────────────────────────────
# Each returns {selector: (expected_number, strength)}.

DEF, REG, CONV = "DEFINITION", "REGISTER", "CONVENTION"


def m_simple(v, f):
    P, apr, yrs = num(v["#amt"]), num(v["#rate"]), num(v["#years"])
    n = yrs * 12
    m = oracles.monthly_payment(P, apr, n)
    return {"#monthlyResult": (m, DEF)}


def m_repayment(v, f):
    P, apr, yrs = num(v["#loanAmount"]), num(v["#interestRate"]), num(v["#termYears"])
    n = yrs * 12
    m = oracles.monthly_payment(P, apr, n)
    return {"#displayMonthly": (m, DEF),
            "#displayTotalInterest": (m * n - P, DEF),
            "#displayTotalRepayable": (m * n, DEF)}


def m_stamp_duty(v, f):
    price = num(v["#price"])
    buyer = v.get("#buyerType", "next")
    buyer = {"ftb": "ftb", "additional": "additional"}.get(buyer, "standard")
    return {"#sdltResult": (sdlt_from_register(price, buyer, f), REG)}


def m_rate_forecaster(v, f):
    # CONVENTION: the 24-month then 36-month phase split is the tool's own
    # design (read from its script), not a published rule. The annuity and the
    # balance recursion inside each phase ARE definitional.
    P, yrs = num(v["#fcAmount"]), num(v["#fcTerm"])
    r1, r2, r3 = num(v["#rate1"]), num(v["#rate2"]), num(v["#rate3"])
    n = int(yrs * 12)
    pays = oracles.reprice_schedule(P, n, [(r1, 24), (r2, 36), (r3, n - 60)])
    return {"#pay1": (pays[0], DEF), "#pay2": (pays[1], CONV),
            "#pay3": (pays[2], CONV), "#diff2": (pays[1] - pays[0], CONV)}


def m_equity_release(v, f):
    # #erMaxCash IS modelled since 2026-08-11: the owner pinned the ORIGINAL's
    # age->LTV step table (work item 97f4d0ab, spec contract) and the rebuild
    # now carries it. CONVENTION, not law — the original states it as an
    # industry-averages approximation. (The comment that stood here said
    # "rebuild 124k vs original 120k" — figures SWAPPED, and that swap
    # propagated into the 08-11 morning decision text; WRONG_CALLS 2026-08-11.
    # The original's table gives 124k at 65 on 400k; 120k was the pre-08-11
    # rebuild's linear formula.)
    age, val = num(v["#erAge"]), num(v["#erValue"])
    P, apr = num(v["#erLoan"]), num(v["#erRate"])
    if age < 55:
        # Minimum-age refusal is part of the pinned contract. The tool's
        # terse markers, not the prose sentence (which dies on a copy edit).
        dash = "—"
        return {"#erMaxCash": ("N/A", CONV), "#debt10": (dash, CONV),
                "#debt20": (dash, CONV), "#debt30": (dash, CONV)}
    for floor, ltv in ((85, 0.52), (80, 0.47), (75, 0.42), (70, 0.36),
                       (65, 0.31), (60, 0.25), (55, 0.20)):
        if age >= floor:
            break
    return {"#erMaxCash": (val * ltv, CONV),
            "#debt10": (oracles.compound(P, apr, 10), DEF),
            "#debt20": (oracles.compound(P, apr, 20), DEF),
            "#debt30": (oracles.compound(P, apr, 30), DEF)}


def m_bridging_loan(v, f):
    # DEFINITION: retained-interest gross-up. The gross advance carries both the
    # arrangement fee and the whole term's interest, so
    #   gross = net / (1 - fee% - rate%*months), all as fractions.
    net, rate, months, fee = (num(v["#brLoan"]), num(v["#brRate"]),
                              num(v["#brTerm"]), num(v["#brFee"]))
    gross = net / (1.0 - fee / 100.0 - (rate / 100.0) * months)
    return {"#resGross": (gross, DEF),
            "#resFee": (gross * fee / 100.0, DEF),
            "#resInterest": (gross * (rate / 100.0) * months, DEF),
            "#dispNet": (net, DEF)}


def m_fee_analyser(v, f):
    # The annuity is definitional; what counts as "total cost" is the TOOL's
    # convention and the whole reason this page now shows both figures
    # (handoff 08-10 §1b): tcTotal = interest over the deal period + fees;
    # tcOutlay = every payment made over the deal period + fees.
    P, apr = num(v["#tcAmount"]), num(v["#tcRate"])
    yrs, deal = num(v["#tcTerm"]), num(v["#tcDeal"])
    fee, other = num(v["#tcFee"]), num(v["#tcOther"])
    n, dn = yrs * 12, int(round(deal * 12))
    m = oracles.monthly_payment(P, apr, n)
    bal = oracles.balance_after(P, apr, n, dn)
    interest_paid = m * dn - (P - bal)
    return {"#tcMonthly": (m, DEF),
            "#tcTotal": (interest_paid + fee + other, CONV),
            "#tcOutlay": (m * dn + fee + other, CONV)}


def m_overpayment(v, f):
    P, apr, yrs = (num(v["#opBalance"]), num(v["#opRate"]), num(v["#opYears"]))
    extra = num(v["#opAmount"])
    saved_i, _ = oracles.overpayment_saving(P, apr, int(yrs * 12), extra)
    return {"#saveInterest": (saved_i, DEF)}

# #saveTime IS NOT ASSERTED, and the reason is worth more than the assertion.
#
# It was modelled first ("3 years 6 months" from oracles.overpayment_saving's
# months-saved), and disagreed with the page by EXACTLY ONE MONTH on three of
# four vectors — always one month more, never less. That pattern is not an
# arithmetic fault: both schedules are the same textbook amortisation, and they
# part company only on WHEN A BALANCE COUNTS AS CLEARED. The page stops at
# `remaining > 0.005` (half a penny); oracles.amortise stops at 1e-9. A residual
# between those two thresholds ends the schedule a month earlier on one side
# than the other, and which vectors land in that window is arbitrary.
#
# Nothing published settles a sub-penny residual, so asserting either number
# would be pinning MY convention as the tool's law — the precise mistake the
# neighbouring lane recorded in WRONG_CALLS on 2026-08-09 (six "mismatches"
# that were its own rounding convention) and that PLAN_2026-08-09 §5.4 forbids:
# state the formula/convention distinction, never edit the checker into
# agreement. The arithmetic is already defended by #saveInterest, which agrees
# to the penny. To assert the duration later, first settle the residual
# threshold as a stated convention of the tool — then it can be pinned honestly.


MODELS = {"simple": m_simple, "repayment": m_repayment, "stamp-duty": m_stamp_duty,
          "rate-forecaster": m_rate_forecaster, "equity-release": m_equity_release,
          "bridging-loan": m_bridging_loan, "fee-analyser": m_fee_analyser,
          "overpayment": m_overpayment}


def main():
    args = sys.argv[1:]
    mutations = {}
    while "--mutate" in args:
        i = args.index("--mutate")
        fid, _, val = args[i + 1].partition("=")
        mutations[fid] = float(val)
        del args[i:i + 2]
    only = set(args)
    facts = load_register_bands()
    for fid, val in mutations.items():
        if fid not in facts:
            sys.exit("--mutate: %r is not a registered fact id" % fid)
        print("MUTATED %s: %g -> %g  (this run MUST report a mismatch)"
              % (fid, facts[fid], val))
        facts[fid] = val
    print("register: %d SDLT facts loaded live from site_specs\n" % len(facts))
    checked = agreed = mismatched = 0
    by_strength = {DEF: 0, REG: 0, CONV: 0}
    unmodelled, unasserted = [], []

    for fn in sorted(glob.glob(os.path.join(CRIT, "*.criteria.json"))):
        slug = os.path.basename(fn)[: -len(".criteria.json")]
        if only and slug not in only:
            continue
        doc = json.load(open(fn))
        model = MODELS.get(slug)
        if not model:
            unmodelled.append(slug)
            continue
        for check in doc["checks"]:
            v = driven(check)
            try:
                want = model(v, facts)
            except Exception as e:
                print("MODEL-ERROR %s/%s: %s" % (slug, check["id"], e))
                continue
            tol = 0.50 if slug in ROUNDS_TO_POUND else TOL
            for sel, pinned in sorted(check["expect_values"].items()):
                if sel not in want:
                    unasserted.append("%s/%s" % (slug, sel))
                    continue
                exp, strength = want[sel]
                checked += 1
                if isinstance(exp, str):
                    # A units assertion — compared as text on purpose (see
                    # m_overpayment); whitespace is the only latitude, which is
                    # also all `computed_values` itself allows.
                    ok = " ".join(pinned.split()) == " ".join(exp.split())
                    detail = "expected %r" % exp
                else:
                    ok = abs(num(pinned) - exp) <= tol
                    detail = "%s %.4f  delta %+.4f" % (strength, exp, num(pinned) - exp)
                if ok:
                    agreed += 1
                    by_strength[strength] += 1
                else:
                    mismatched += 1
                    print("MISMATCH  %-16s %-16s %-22s pinned %-14s %s"
                          % (slug, check["id"], sel, pinned, detail))

    print("\n%d pinned value(s) recomputed: %d agree, %d MISMATCH" % (checked, agreed, mismatched))
    print("   of the agreements: %d DEFINITION, %d REGISTER, %d CONVENTION"
          % (by_strength[DEF], by_strength[REG], by_strength[CONV]))
    if unmodelled:
        print("NOT VERIFIED (no independent model): " + ", ".join(sorted(unmodelled)))
    if unasserted:
        seen = sorted(set(unasserted))
        print("pinned but NOT re-derived (%d): %s" % (len(seen), ", ".join(seen)))
    sys.exit(1 if mismatched else 0)


if __name__ == "__main__":
    main()
