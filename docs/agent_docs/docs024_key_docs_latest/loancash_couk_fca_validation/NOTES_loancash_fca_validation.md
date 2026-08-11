# NOTES — loancash.co.uk FCA validation

Append-only, newest at the bottom.

## 2026-08-11 — the premise was wrong: the caps are CURRENT, not dated

**This workstream was opened on a stated concern that turns out not to hold**, and
that is the first thing the next session needs to know, because the concern is
repeated in three places and will otherwise be inherited again.

`HANDOFF_2026-08-10c` §8 (and the two handoffs before it) say:

> *"two of its three tools hardcode dated FCA caps (0.8%/day, £15 default fee, 100%
> total cost) with nothing checking them against CONC 5A. Same shape as the SDLT bug
> (`bugs_open/225`), which was a tax rule 16 months out of date and under-quoted by
> £5,000."*

**Measured against the FCA's own Handbook, not the page and not memory**
(`handbook.fca.org.uk/handbook/CONC/5A/2.html`, fetched 2026-08-11):

| rule | Handbook | what the tool does | verdict |
|---|---|---|---|
| initial cost cap | **0.8% per day** of the amount of credit — CONC **5A.2.3R** | `maxInitial = amount * 0.008 * days` | ✅ correct |
| default charge | **£15** for breach-related charges combined — CONC **5A.2.14R** | `if (dflt > 15) breach` | ✅ correct |
| total cost cap | charges may not **exceed the amount of credit**, i.e. 100% — CONC **5A.2.2R** | `maxTotal = amount` | ✅ correct |

**Last amended 02/01/2015; no subsequent amendment indicated.** So the figures are
not stale, and the SDLT analogy does not transfer: SDLT thresholds move with
Budgets, the HCSTC price cap has not moved in eleven years (the FCA reviewed it in
2017 and left it unchanged).

### The arithmetic was also checked, not just the constants

**`tools/price-cap-checker.html`** — sound. The input labels are what make it
correct and they are easy to get wrong: *"Total interest and fees you paid"* (so
`charged` is **charges only**, not the total repaid, which is why `charged > amount`
is the right test for a 100% cap) and *"Of that, default fees"* (so the default fee
is a **subset** of `charged`, which is why the allowance is
`maxInitial + min(dflt, 15)` rather than a separate addition). Had either label meant
the other thing, the same code would be wrong by roughly a factor of two. **The
labels are load-bearing arithmetic here** — do not "tidy" them.

**`tools/true-cost-calculator.html`** — sound. Caps `rate` at 0.8, caps `charge` at
`amount`, and computes the equivalent APR as `(1 + rate/100)^365 - 1`, which is the
correct definition (annualised, compounded) and yields ~1,733% at the cap — the
familiar four-figure payday APR. Its credit-union comparator uses **3%/month**, the
GB statutory maximum (Northern Ireland is 1%), applied flat, and the surrounding
copy says *"and usually less, because interest is charged on the reducing balance"* —
which is the honest direction: a flat calculation **overstates** the credit-union
maximum, so the comparison never flatters the payday loan.

### What IS true, and is the actual gap

**Nothing checks any of this.** The three numbers are correct today and there is no
instrument that would notice if they stopped being. That is a real gap and it is the
one worth building for — but it is a *monitoring* gap, not a *correctness* defect,
and it does not carry the SDLT bug's urgency. Nobody is being under-quoted £5,000.

`[UNMEASURED]` The third tool, `tools/complaint-deadline-calculator.html`, has not
been checked. It encodes limitation periods (the six-year/three-year rule) and the
FOS six-month deadline, which are a **different legal source** — not CONC 5A — and
they are the kind of rule that does move. **That is now the highest-value unchecked
thing on this site**, and the concern that motivated this workstream should be
transferred to it rather than dropped.

### Correction to make in the upstream handoffs

`HANDOFF_2026-08-10c` §8 should be amended where it is read: the caps are verified
current as of 2026-08-11; the priority is the complaint-deadline tool and an
ongoing check, not a suspected stale-figure defect.
