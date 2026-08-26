# SUMMARY — 2026-08-26 · CTA destination relevance (`bugs_open/391`)

## What we are trying to do

Every one of our sites carries call-to-action buttons — "try this", "get in touch", "run the
calculator". The framework chooses where each button points. It was choosing badly: on three sites
the primary button on almost every page sent visitors to a password-strength toy that had nothing to
do with the business. The job is to make those buttons point somewhere relevant, and to remove the
toy from the three sites entirely.

## Where we have come from

The cause turned out to be duller and worse than a bad guess. The framework ranked candidate
destinations by a menu-order number and took the first one. No topic, no tags, no judgement. The toy
carried a fossil `1` set at page creation in March, so it won everywhere.

Worse, the framework then **writes the button's wording to name whatever it picked** — so a wrong
choice writes the copy that locks it in. The next pass matches the wording back to the same page and
keeps it. Twenty of eighty buttons had reached that locked state, including all three the owner
reported.

## What we have done

**The locked buttons are fixed and the fix is proven on the live sites.** The query that finds "a
button naming the password tool and pointing at it" now returns nothing anywhere — it was twenty of
eighty. Twelve pages were rewritten and each was checked on the live site, not in the database.

**We also broke two pages and put them back.** The rewrite we commissioned to change button labels
rewrote page bodies too, replacing whole sections with copies of their neighbours — on two of the
twelve. Both were restored word-for-word from the record the damaging write left behind. The check we
had been using could not see it: it counted paragraphs, and the damage swapped three paragraphs for
three. The check that works compares whether a page's sections are still distinct from one another.

**The retirement is half applied and safe.** Two of the three tool pages are archived, which freezes
a page while continuing to serve it — so nothing is dead and everything is reversible. The third is
held back deliberately as the fiddly one.

**And we found that the plan's order could not work.** Retiring a page is refused by the platform
while anything still links to it; re-pointing those links is blocked while the page is still valid.
Each was the other's precondition. Archiving turns out to release both locks, so the sequence is
three steps — archive, re-point, remove — and it is now proven end to end on two test pages.

## Where we are now

**One decision is blocking the rest, and it is the interesting part.** Forty-one buttons still point
at the tool. Re-pointing them sends each to whichever tool the framework judges most relevant — right
for a button that says "Try the AI Data Risk Checker", wrong for one that says "Write to
leopardess@contactforsales.com", which should go to a contact page and would go to a calculator.

**Twenty-three of the forty-one are the second kind.** So the obvious version of the remaining work
would have cleared every reference to the toy — exactly what we said success looked like — and left a
majority of the buttons quietly wrong, in a way nobody would report because the destination looks
plausible.

Both test pages showed both outcomes, which is how we know the difference is real.

## Where we are going

The owner decides what the twenty-three should do — contact page, rewritten copy, or something
per-site. Then the rest is mechanical: re-point the remainder, archive the last site, remove all
three pages, and sweep the three sites for any surviving reference including the footer.

Separately, the underlying cause is still live for every future site: the ranking has no notion of
relevance. The agreed direction is an explicit opt-out plus a detector for the fossil-number shape,
read at the ranking rather than at the loaders. That work has not started and needs an architecture
round.
