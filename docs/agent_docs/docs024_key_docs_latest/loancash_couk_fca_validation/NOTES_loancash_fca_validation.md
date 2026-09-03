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

---

## 2026-09-03 — the monitoring gap is CLOSED: migration 738 gives this site an evidence register

Owner-directed, on the back of `RFC_060` §1d: loancash was the **last** of the five register-less
finance sites. Method followed step for step from `lendzy_co_uk/RUNBOOK_lendzy_co_uk.md` §8.

**This discharges the ask this lane wrote on 2026-08-11 and could not then satisfy** — *"nothing is
checking … what actually earns its keep is something that reads the rulebook and shouts if it
disagrees with what is on our page."* That mechanism existed by 09-02 (the daily `evidence-freshness`
refresher re-fetches every citation URL and re-checks every quote); what was missing was a register
for it to read. There now is one.

**And the `[UNMEASURED]` item above is now measured and registered.** The complaint-deadline tool's
limitation periods were this lane's named highest-value risk *because they move*. All three are now
facts with citations: `DISP 2.8.2(1)` (six months from the final response), `DISP 2.8.2(2)(a)` (six
years from the event), `DISP 2.8.2(2)(b)` (three years from awareness), plus `DISP 1.6.2` (the eight
weeks). If the FCA changes any of them, the daily re-check now sees it.

### What went in

**19 facts** — 11 handbook rules, 8 statutory — covering everything the site actually asserts:
the price cap trio (CONC 5A.2.2/.3/.14), the refinancing cap (5A.2.10), the two-rollover limit
(CONC 6.7.23), both CPA rules (CONC 7.6.12/.14), affordability (CONC 5.2A.4), the complaint clock
(DISP 1.6.2, DISP 2.8.2 ×3), Breathing Space (SI 2020/1311 regs 26/32/24/16), the credit-union 3%
ceiling (SI 2013/2589 art 2) and the illegal-lending basis (FSMA 2000 ss.19/23). Plus **6
banned_claims** from the sibling lending set.

`[MEASURED 2026-09-03]` **338 regulatory-shaped sentences across all 30 served pages** (crawled at
the artefact, invented-URL 404 control confirming this is not a parked domain answering everything),
of which **only 3 cite a rule number**. That ratio is why this site was the riskiest of the five: the
other four calculate, this one *states the law*.

### Three wrong live claims — the base rate held for a fourth lane

Every lane that has read the source has found errors (lendzy 2, loanzy 1, loancalculator 2, now
loancash 3). All recorded as `corrects_site_citation`; **served copy untouched** — rewriting
published prose on an automated finding is authority the owner withheld.

1. **The £15 default cap is CUMULATIVE across the whole agreement.** CONC 5A.2.14R: charges must not
   exceed £15 *"whether in relation to one breach or cumulatively in relation to multiple breaches of
   the agreement"*. Two pages call it a per-missed-payment fee. **The site understates the protection
   it exists to explain** — a reader with two missed payments would accept a second £15 as lawful.
2. **The CPA limit is two REFUSED requests, and there is no £1 threshold in CONC 7.6 at all.**
   `guides/loan-sharks-and-illegal-lending.html` says *"cannot take more than one payment attempt of
   over £1"*. The site's own `guides/stopping-payments-the-cpa-rules.html` states it correctly
   (*"A lender can attempt to collect a payment using your CPA twice"*), so this is an internal
   contradiction on one page, not a house error.
3. **Affordability is CONC 5.2A, not "CONC 5A."** `guides/check-your-lender-is-authorised.html`
   attributes the affordability checks to CONC 5A — the *price cap* chapter, which contains no
   affordability rule. Cited correctly on three other pages.

### Two traps, both measured, both worth carrying

**The "shared finance banned-claims set" is TWO sets.** lendzy carries a bare `\bno credit checks?\b`;
loanzy and loancalculator carry a narrow variant requiring the product noun. `[MEASURED]` the **bare**
variant fires on this site's *correct* advice that an employer salary advance involves "no interest
and no credit check" (**1 hit**); the narrow one fires **0** across 30 pages. Adopting "the shared
set" without checking *which width* would have convicted this site of its own accurate guidance.
**A coverage count of "sites carrying the set" cannot see this difference.** The narrow variant was
adopted deliberately, and the reason is written into the pattern's `reason` field so the next reader
does not "tidy" it wider.

**A hand-transcribed quote fails silently and for ever.** The DISP 2.8.2(2)(b) quote was first
written with commas where the source has parentheses. It returned **false** on the production
matcher. Shipped, it would have read as `citation_lost` drift **every single day**, and that false
alarm is indistinguishable from a real one. `cmd/fcaquotecheck` caught it in the run that was
supposed to be a formality. **Paste the quote; never retype it.**

### Verification, and what was NOT done

- **19/19 quotes true through the production matcher** (`cmd/fcaquotecheck` →
  `datahelpers.VisibleTextFromHTML` / `QuoteFoundInText`), with a deliberately-absent control
  **false in every run**. Every URL confirmed by `<title>`, and the URL **stored** is the one that
  answered after the redirect.
- **6/6 banned patterns compile** under the production prefix `regexp.Compile("(?i)"+p)`
  (`claims.go:468`), **6/6 fire on a positive control**, **0/6 match anything served**.
- **Both migration guards mutation-tested against the live DB**: fact count 19→18 aborted with
  `738 VERIFY: expected 19 facts, found 19`; flipping the pre-insert guard predicate aborted before
  the INSERT. Then the whole file was run with `COMMIT` swapped for `ROLLBACK` — guard passed,
  `INSERT 0 1`, verify NOTICE fired, and the site was left with 0 rows and 14 current specs.
- Applied **by hand**, not by `run-migrations.sh --apply`, because `--apply` takes EVERY pending file
  and the dry run confirmed other lanes had files queued (708, 734). Registered afterwards with
  `--record-only` and a dated note.
- **`[NOT DONE, deliberately]` I did not dispatch `refresh_evidence_base` at the site to watch a pass
  run.** A pass rewrites the register as a new row and would expire migration 738's rollback guard
  while the council verdict is still pending. The two things such a dispatch would prove — that the
  quotes match and that the patterns compile — were already proven **at the same production
  functions** the refresher calls. The natural 09:10 UTC tick picks it up tomorrow.
- Council: `Council-Submitted: cf7470b7-d922-4e2d-aa84-b7aae489cadd`, verdict pending at time of
  writing. **Read it and act on it** — the migration is already applied, so a REVISE is discharged by
  a follow-up migration, not by an amend.
