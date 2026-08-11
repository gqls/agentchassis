# PLAN — loancash.co.uk FCA validation

**Opened 2026-08-11.** Read `NOTES_loancash_fca_validation.md` first: the concern
this workstream was opened on **does not hold**, and the plan below is what it
became once measured.

## What we thought, and what is actually true

Three handoffs carried the claim that loancash's tools hardcode *dated* FCA caps,
by analogy with `bugs_open/225` (an SDLT rule 16 months stale, under-quoting by
£5,000). Verified against the FCA Handbook 2026-08-11: **all three caps are current
and the arithmetic implementing them is correct.** The price cap has not moved since
02/01/2015.

So the SDLT analogy is wrong in the way that matters. SDLT thresholds move with
Budgets — a stale figure there is a *when*, not an *if*. The HCSTC price cap has
been static for eleven years. **The risk is not that these numbers are wrong; it is
that nothing would tell us if they changed.**

## What this workstream is for, restated

1. **The unchecked tool, not the checked ones.**
   `tools/complaint-deadline-calculator.html` is unexamined and encodes limitation
   periods and the FOS six-month deadline — a *different* legal source from CONC 5A,
   and one that does move. This is the highest-value item and should be done first.
2. **A standing check, cheap enough to keep.** Something that fails loudly if a
   Handbook figure diverges from a hardcoded constant. Cost matters more than
   sophistication: a check nobody runs is worth nothing.
3. **Only then**, an arithmetic oracle in the LMC mould (`oracle.py`) if the tools
   turn out to need one. They may not — see below.

## Design notes, so the next session does not overbuild

**Do not clone LMC's `oracle.py` reflexively.** That oracle exists because LMC has
23 calculators doing genuinely complex, independently-derivable arithmetic
(amortisation, SDLT bands, affordability multiples) where an external right answer
exists and is worth computing separately. loancash has **three** tools and the two
checked ones are each a handful of multiplications against three constants. An
independent implementation of `amount * 0.008 * days` is not an oracle; it is the
same expression typed twice, and it would agree with a wrong constant just as
happily as with a right one.

**What actually has evidential value here is the CONSTANT-vs-SOURCE check**, not a
recomputation. The failure mode this site has is "the law changed and the page did
not", and only something that reads the law catches that.

Sketch, in the order that gets value soonest:

- **Phase 1 — read the tools' constants out of the HTML** and assert them against a
  small declared table of `(rule reference, expected value, source URL)`. Fails if a
  page's constant no longer matches what we recorded. This catches *edits to the
  site*, immediately and cheaply, and needs no network.
- **Phase 2 — verify the declared table against the Handbook**, on a slow cadence
  (a fetch of `CONC/5A/2.html`, asserting the three figures still appear against
  their rule numbers). This is the half that catches *the law changing*, and it is
  the half that must be allowed to fail noisily.
- **Phase 3 — the complaint-deadline tool**, once its legal source is identified and
  written down the same way.

**The trap to design against, from `bugs_open/225`'s lane:** a checker whose expected
value is copied from the page it checks agrees with itself for ever. Phase 1's table
must be populated from the **Handbook**, and its provenance (rule number + URL +
date read) recorded beside each figure — so that a future session can tell a verified
figure from a transcribed one. Mark anything transcribed `[UNVERIFIED]` until someone
has read the source.

**Induction requirement**, before either phase is believed: change a constant in a
scratch copy of the page and confirm Phase 1 fails; change an expected value in the
table and confirm Phase 2 fails. A green run from a checker that has never been seen
red is not evidence — that rule has been earned five times on this estate.

## Out of scope

- Rewriting or re-voicing loancash's copy. The framework owns copy (owner ruling
  2026-08-06).
- Decomposition of loancash's 18 pages. That is Track C of the decomposition
  workstream and is a separate brief with separate tooling needs.
