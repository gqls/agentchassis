# SUMMARY 2026-08-10 — the whole estate has answered the offer question, and the answers hold

*Read-out for the owner. Current state only; chronology in NOTES and README_where_we_are.
Previous: `SUMMARY_2026-08-09_offer_track_running_ahead.md`.*

---

## What we're trying to do

Unchanged: a vigilant designer and an offer-and-benefit analyser, each a standing faculty,
each built to the rule that every detector ships with the thing that acts on it. The offer
side keeps asking one question per site — does this site answer its target market's need,
in a way that pays us — judged against the site's **own recorded premise**, because we have
no visitor data and an analyser that pretended otherwise would be confidently wrong.

## Where we've come from

Yesterday's read-out ended with the two offer checks written as tested code, council
verdict pending, and deliberately inert: the names could not be switched on until an image
carrying the code had rolled. That happened, they were switched on observe-only, and three
sites were hand-picked to prove the checks could both speak and stay quiet. This read-out
is what happened when the whole estate went past them unattended.

## What we've done

**Every active site has now been examined, and the population is small and specific.**
Twenty-one sites, four findings, and each one verified by hand against live rows rather
than taken on the check's word:

- **Three sites have no usable premise at all** — loancash.co.uk,
  loanandmortgagecalculator.co.uk (no strategy record whatsoever) and gaswholesalers.com
  (the pre-2026-05 shape, no revenue model). Nothing else we build can judge these sites
  until that is fixed, which is precisely why this check exists.
- **mortgagecalculator.co.uk, recorded as lead-generation, has no conversion path.** Its
  contact page was planned and never shipped; the thirty pages that did ship are guides and
  calculators, and the only forms on the site are calculator inputs. A site whose entire
  recorded model is enquiries cannot receive one.

**The silences were tested, not assumed — and this is the part that makes the findings
trustworthy.** A detector that never fires and a detector that is broken look identical, so
each quiet site was checked by hand. The three tool sites: a word-bounded search for all
twelve service-selling phrases across every shipped component returned exactly **one** hit
fleet-wide, and it is prose in an article, not a button — the check correctly ignored it,
because it reads anchor and button text only. That design decision took an argument to
settle and earned its keep on the first unattended run; a whole-HTML matcher would have
opened this programme with a false positive. oufe.com's silence was likewise verified
truthful: real contact page, real form, linked from site chrome.

**One silence carries no information, and it should be said out loud.** vetcomparison.uk is
a sponsored-listings site, and doc 028 states no rule for that model, so the check's default
arm returns nothing by design. Silence-because-no-rule is indistinguishable from
silence-because-fine at the point of reading, and that is a gap in what we can *say* about
that site rather than a defect in the check.

**We also found five sites that had been marked examined without being examined**, re-ran
all of them (one turned out to be a missing-premise finding that had been sitting invisible
behind its own tick), and contributed the underlying evidence to the lane that owns the
scheduling mechanism. The trade-off that causes it is their deliberate, documented choice —
we initially wrote it up as a flaw and corrected that in place. What stands is narrower and
real: their daily health check compares fleet totals, so a partial loss of any size hides
inside the arithmetic, and a lane's own re-runs inflate the number that clears it.

## Where we are now

- **B1, B2, B3 all live; B3 observe-only across the full estate.** Findings are born
  `detected`; nothing acts on them.
- **Four true findings, zero false positives, zero check failures** across every run since
  enablement.
- **The council trail is parked one round short of approval**, blocked on the three
  migration-ledger commands only the owner can run (the session permission classifier
  refuses the INSERT). Four reviewer seats independently called an applied-but-unrecorded
  migration a loaded gun for whoever runs the migration tool next, and they are right.
- **The three missing-premise findings are actionable now** — they are the exact item the
  strategist consumes, and B2's gate means running one against a live site no longer
  re-plans it.

## Where we're going

The findings are arguments, not orders, and the next decision is the owner's: whether the
three `needs_strategy` items get promoted and drained (each one writes a site's first
recorded premise), and whether `missing_conversion_path` on mortgagecalculator.co.uk gets
a fix route or stays a roadmap row. After that, B4 — the analyser proper, which reads
everything B1–B3 built and only ever judges what a mechanical check cannot.

Two smaller things stay open: the sponsored-listings model has no stated rule, so one site
is structurally unjudgeable; and the greenfield negative control has still never been
exercised, because no greenfield build has run since B2's gate shipped.

The open questions from `features_open/030` remain the owner's: which council (if any) the
offer judgement belongs to, which of the two missing correspondence routes matters first,
and whether premise *quality* ever comes into scope.
