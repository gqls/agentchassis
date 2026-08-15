# What the arithmetic oracle is

**Written 2026-08-15 for the owner.** Every figure here was measured the day it was
written; the commands are in §7 so you can re-run any of them. Registered in the concept
register as **SQAM-003**.

---

## 1. What it is, in one paragraph

The oracle is a program that works out what each calculator on the site **ought** to say,
from the published maths — and then opens the real page in a real browser, types numbers
into it, presses the button, reads the answer off the screen, and compares. It is the
only thing we have that can tell us a calculator is **wrong**. Everything else we have can
only tell us a calculator has **changed**.

## 2. The distinction that makes it worth having

We have two kinds of check, and they answer different questions.

A **golden** records what each calculator produced on a particular day and re-runs it
later to see whether the answer moved. That catches breakage: if a page used to say
£1,390 and now says £0.00, the golden fires. It is cheap, it needs to know nothing about
any particular calculator, and it runs unattended across every page.

But a golden has a hole in it that no amount of care can close: **if a calculator has been
wrong since the day it was written, the golden faithfully records the wrong number, and
every later comparison certifies it.** The check passes forever. The site keeps
confidently telling people the wrong figure, and every instrument we own reports green.

The oracle exists to close that hole. It does not ask "is this the same as last time?" It
asks **"is this the right answer?"** — and it works out the right answer from a source that
has never seen our code.

This is not a theoretical concern. When the oracle was first run in August it immediately
found two things every existing check had passed for months:

- **Stamp duty was being calculated under an HMRC rule that expired on 31 March 2025** —
  under-quoting by a flat £5,000. Sixteen months out of date, on a page whose entire
  purpose is to tell someone what they will owe.
- **A 0% interest rate broke six of the seven calculators that accept one.** Not an
  academic input: 0% finance is a real product people are offered.

Neither is the kind of thing a "has it changed?" check can ever find, because neither had
changed.

## 3. Where the right answer comes from — the part that carries all the weight

The oracle's expectations are written from sources **outside this estate**: the standard
annuity formula for a repayment loan, month-by-month amortisation, the stamp duty bands as
published by HMRC on gov.uk with the date each took effect, and plain arithmetic identities
(a yield is annual rent over price; a loan-to-value is loan over value).

**Nothing is copied from the page's own code.** That rule is the whole value of the thing,
and it is easy to break by accident: an oracle transcribed from the calculator it is
checking is not an independent check at all — it is the same claim written twice, and it
will agree with the bug every time. The stamp duty defect above is the proof. A
transcribed oracle would have certified it, because the expired rule was in the code and
would have been copied along with everything else.

There is a second discipline underneath it: the person writing a tool's spec works from
what the page **shows** — its visible labels, its button text, its result captions — and
only opens the calculation code **after** a check has already failed, to find out why.

## 4. How it decides, and why there are four answers rather than two

A check can come out four ways, and the two extra ones are there for good reasons.

- **PASS** — matches the independent answer, within a tolerance derived from what the tool
  actually displays. A page that prints whole pounds cannot be checked to the penny by
  anybody, so the assertion is ±50p and the report says so; a page printing pence is held
  to ±1p. The tolerance is *derived*, never chosen, because one global tolerance would
  either convict every whole-pound tool or excuse every penny-level defect.
- **CONVENTION** — matches a *different but defensible* reading. "Total interest" genuinely
  has two: using the exact payment, or the rounded payment the borrower is really billed.
  They differ by pennies. Both appear on this site. That is reported loudly but not
  convicted — though two tools on one site answering the same question differently is
  itself worth knowing.
- **FAIL** — matches no stated convention. This is the finding the whole exercise exists
  to produce.
- **N/A** — the check could not be made. Printed in the report rather than quietly dropped,
  **because a silently skipped check reads exactly like a passing one.**

A related trap the design takes seriously: when a calculator names a *wrong* answer it has
given (say, an expired tax rule), that must not be filed in the same bucket as a
defensible convention. The very first run did exactly that and labelled the expired HMRC
rule a "convention" — the machinery for naming a cause was the same machinery that excused
one. They are separate now.

## 5. Choosing the numbers to type in

The vectors are chosen per tool, not generated. They sit deliberately on **boundaries**:
band edges (£125k, £250k, £500k, £625k, £925k, £1.5m), the £40,000 higher-rate floor, a 0%
rate, a one-month term, a zero balloon payment.

This is the other thing a general-purpose harness cannot do. A generic tool scales each
field's own default up and down to stay in a sensible range — which is precisely why it
can test 23 pages unattended, and precisely why it can never land on a band edge. A defect
that only exists between £500,000 and £625,000 is invisible to every input such a harness
can produce. The oracle knows which calculator it is looking at, so it can name the number.

## 6. Why the oracle is also checked — and what happened this week

A checker that has quietly stopped checking looks identical to a checker finding nothing
wrong. So the oracle is run in three deliberately-sabotaged modes, and in each one it
**must** come out red:

- corrupt the expected answers — every check should fail;
- corrupt the reading of numbers off the page — every check should fail;
- judge each set of inputs against the *neighbouring* case's expected answers, on a real
  working page — every comparison should fail.

If anything passes under sabotage, the comparator is not really looking, and the rule is
strict: **no check may pass.** Stray passes are excluded by printed name and reason, never
by loosening a threshold — loosening until it goes green is the trap these controls exist
to prevent.

**That machinery earned its place this week, by catching itself.** Closing the recent batch
of work, the sabotage run announced *"8 checks passed — the checker is inert"*. It was
wrong, and in an instructive way. Seven of those checks read **words** rather than numbers
("Option A is Cheaper", "2 Years 3 Months"), and corrupting numbers cannot disturb a check
reading words. The eighth was a coincidence: the sabotage asserts the number 100 on the
grounds that nothing computes it, and one test case is a £300,000 loan against a £300,000
property, where the true loan-to-value **is** 100%.

Then my fix for that was itself wrong on one axis, and a second session working this site
caught it: I excluded those word-checks from *both* sabotage modes when only one of them
required it — silently removing six real comparisons. That was corrected, and a third
fault the same session found and fixed (guards that recognised decimals but not whole
numbers) has left all three controls green for the first time.

The lesson worth keeping is not any of the three bugs. It is that **an alarm which goes off
wrongly is one people learn to ignore**, and this alarm is the last thing standing between
us and shipping a calculator that lies to people.

## 7. Where it stands today (measured 2026-08-15)

| | |
|---|---|
| calculators with an independent oracle | **18** |
| arithmetic checks | **170 PASS · 0 FAIL · 6 CONVENTION** |
| scoring tools with no external right answer | **23 invariant checks, 0 failing** |
| sabotage controls | **all three green** (expectation, parse, cross-case) |

The 6 CONVENTIONs are not defects — they are tools taking the billed-vs-exact rounding
choice, reported so the choice stays visible.

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
python3 $LANE/oracle.py                       # the full sweep
python3 $LANE/oracle.py --tools stamp-duty    # one tool
python3 $LANE/oracle.py --mutate expectation  # sabotage: must come out red
python3 $LANE/oracle.py --selftest-parse      # sabotage: must come out red
python3 $LANE/oracle.py --mutate crosstool    # sabotage: must come out red
python3 $LANE/invariants.py                   # the scoring tools
```

## 8. What it does not do, stated plainly

- **It covers 18 of the site's 21 calculator pages.** The rest are checklists and scoring
  tools with no external right answer to compare against; they get weaker invariant checks
  (does the score move the right way, stay in bounds, survive a round-trip) and the report
  says explicitly that this is weaker evidence.
- **It runs when someone runs it.** There is no schedule and no platform integration.
  Nothing else in the estate calls it. If nobody runs it before a change ships, it protects
  nothing.
- **Only this site has one.** `loancalculator.co.uk` and `mortgagecalculator.co.uk` carry
  overlapping tools and have no equivalent. Whether the specs belong in the platform rather
  than in one site's directory is an open question in the register.
- **One residual blind spot, known and written down:** the "corrupt the expectations" mode
  can only ever test the checks that compare numbers. It cannot test the seven that read
  words, so one of those failing silently would be caught by nothing.

## 9. The one-sentence version

It is the difference between *"this calculator still does what it did last year"* and
*"this calculator is telling people the truth"* — and on this site those two turned out to
be different answers for sixteen months.
