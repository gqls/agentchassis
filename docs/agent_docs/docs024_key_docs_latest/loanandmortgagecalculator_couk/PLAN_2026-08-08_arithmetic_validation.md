# PLAN — arithmetic validation of the 23 calculators

**Started 2026-08-08**, from `HANDOFF_2026-08-08_arithmetic_validation.md`, which
carries the owner's *"build the arithmetic validation and 'is the tool doing the
right thing' checks."* That handoff is the brief; this file is the design, the
decisions and their reasons, and the corrections as they landed.

Results and evidence: `REPORT_2026-08-08_arithmetic_validation.md`.
Commands: `RUNBOOK` §13. Running log: `NOTES` (2026-08-08 entries).

---

## The problem, and why the obvious approach is the wrong one

The lane already had behavioural goldens. They prove consistency and cannot
prove correctness — a calculator wrong since birth is recorded faithfully and
certified for ever. The fix is an independent oracle. **The failure mode of an
"independent" oracle is that it is not independent**, and the route to that
failure is short and inviting: to drive a page you need its control ids, and the
fastest way to get them is to read `calculateLoan()` — one line above the
arithmetic.

## Decisions, and the reason for each

**D1 — the interface comes from LABELS; the arithmetic is not read until a check
has already failed.** `inventory.py` reports the visible `<label for=…>` bound to
each control, the button's text, and the caption above each result box. That is
the site's own claim about what each number means, read the way a user reads it.
Authoring from it makes contamination structurally impossible rather than a
matter of discipline. Reading a calculation body *after* a FAIL is diagnosis, not
authorship, and the ordering is what makes that distinction real.
*Consequence, and it is a good one:* if a label and the arithmetic disagree,
that disagreement is a finding rather than something the harness silently
absorbs.

**D2 — vectors are chosen per tool, on boundaries.** `toolgolden`'s x1/x2/x0.5
scaling of each field's own default is the right design for a harness that must
work on any tool with no configuration — and it is exactly why it can never land
on £500,000. An oracle knows its tool, so it can name the number. Band edges,
0% rates, one-month terms, a zero balloon, the £40,000 SDLT floor.

**D3 — four outcomes, not two.** PASS / FAIL / CONVENTION / N-A. A tool that
matches a different but defensible convention (billed vs exact payment) must not
be convicted; a check that could not be made must be LOUD rather than absent,
because a silently skipped check reads exactly like a passing one.

> **CORRECTED, same day, first run.** D3 as first built had ONE bucket for
> "matches something other than my expectation", and I populated it with *the
> superseded HMRC FTB rule*. So the very first stamp-duty run reported a
> calculator running an expired tax rule as a CONVENTION. **The machinery for
> naming a cause was the machinery for excusing one.** Split into `alt`
> (defensible → CONVENTION) and `defect_alt` (a NAMED WRONG ANSWER → still FAIL,
> plus a `DIAGNOSIS:` line). Keep the split in any future checker: a named cause
> and an accepted difference must not share a code path.

**D4 — tolerance is derived from the tool's own display precision, and printed.**
`£1,390` cannot be checked to the penny by anybody; `£202.29` can. One global
tolerance either convicts every whole-pound tool or excuses every penny-level
defect. Printing the resolution on each line makes the limit of the check part
of the check's output.

**D5 — controls are part of the deliverable, and their criterion is "no check
may PASS under mutation".** Not "some check must FAIL".

> **CORRECTED, same day.** The first criterion was "must produce a FAIL", and it
> marked `--mutate parse` inert. That was wrong: an `<input type=number>` rejects
> `£200,000`, the field holds `''`, and `set()` refuses to drive on — which IS
> the "silent 0" the control exists to forbid. A loud refusal satisfies the
> property as well as a FAIL does; only a PASS falsifies it.

**D6 — a stray PASS under a control is excluded BY NAME, never by loosening the
bar.** Two kinds occur and both are printed: `NON-TEST` (adjacent boundary
vectors that genuinely expect the same figure — £1,500,000 and £1,500,001 of
SDLT differ by 12p) and `MUTATION DID NOT BITE` (at £39,999 the borrowed
expectation equals what the *buggy* page prints). The alternative — relaxing a
threshold until the control goes green — is how a control becomes a formality,
and it was available at both points.

**D7 — class C gets INVARIANTS, labelled as not-arithmetic.** Monotonicity,
bounds, determinism, round-trip; plus `portfolio`'s aggregation, which genuinely
is arithmetic and is checked as such. No invented right answers. The report says
in words that a passing invariant is weaker evidence than a passing oracle,
because five green ticks otherwise read like fifteen.

**D8 — a source-free determinism probe, added after the first staleness detector
failed.** The first attempt compared each reading against a primed reading. It
MISSED the stale-answer mode, because the stale figure is whatever the last
*accepted* vector produced — including an intermediate state created halfway
through typing the new one (on `standard-calc`, £143.47, from a combination the
user never entered). Driving one final vector by two different routes has no
such blind spot: the two readings must agree whatever the tool computes. It
turned out to be the best statement of the whole defect family, and it needs no
oracle, no formula and no source.

## Corrections to the brief, recorded rather than silently applied

1. **Classification is 15 A / 3 B / 5 C, not 14 / 3 / 6.**
   `car-finance-calculator` C→A (measured: discounted-balloon PCP, one right
   answer); `portfolio` stays C with its aggregate checked as class A.
2. **`loans/overpayment-calculator` does NOT share `/assets/js/calculators.js`.**
   The brief lists it as sharing; it does not load that file at all. That is
   precisely why it is a zero-rate casualty while `mortgages/overpayment` is not,
   and the distinction is the whole of `bugs_open/224`.
3. **`settlement-calculator`'s rule is answerable and the brief's "name the rule
   (Rule of 78 vs actuarial)" resolves to neither exactly**: the page states its
   own rule in its breakdown text — 58 days' interest, the Consumer Credit (Early
   Settlement) Regulations 2004 deferment — and the arithmetic matches it. What
   it cannot do, having no term input, is the actuarial rebate a real settlement
   figure needs. Recorded as a model limitation, not a defect.

## Refuted, and kept

**My oracle was wrong about `mortgages/rate-forecaster`; the page was right.**
I asserted each window's payment on the FULL original principal over the FULL
original term and filed 4 FAILs. It amortises the balance REMAINING over the term
REMAINING — the correct model. The naive version is retained in the checks as a
`defect_alt`, so a future rewrite that regresses to it is diagnosed rather than
merely failed. A refuted claim costs one run; a wrong root cause in a handoff
costs every thread that believes it.

## Where it stops, deliberately

- **No `--emit-criteria`.** Emitting Tier-4 `computed_values` from a tool not yet
  proven correct pins its wrong answers into the platform's acceptance record and
  then defends them. That step is available once 224 and 225 are fixed, and is
  the obvious next piece of work.
- **No fixes applied.** Brief §9: several of these are consumer-credit and tax
  figures, and a changed answer is a changed claim. The findings go to the owner
  first.
- **No schedule.** The harness is run by hand. Making it a scheduled check is a
  separate decision, and probably belongs in the platform rather than in a lane
  directory — noted as `verify-later` on the concept-register entry (SQAM-003).
