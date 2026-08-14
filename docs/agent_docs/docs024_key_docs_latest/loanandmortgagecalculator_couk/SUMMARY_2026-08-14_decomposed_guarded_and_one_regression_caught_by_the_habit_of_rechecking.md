# SUMMARY — 2026-08-14: the site is decomposed, the platform now guards what decomposition exposed, and one real regression was caught by the habit of re-checking

## What we're trying to do

Hand a site we built by hand — loanandmortgagecalculator.co.uk, fifty-nine pages
across two sites, twenty-three of them live consumer-finance calculators — over to
the framework, page by page, without anyone visiting the site noticing the
handover. Every page was adopted as a single frozen lump of HTML the framework
could serve but not edit. Decomposition splits each page into parts the framework
can address: the words in one piece, the calculator in another. The prize is
control — rewrite the copy without touching the calculator, rebuild a page from
components, stop hand-editing files. The risk is the same as it always was: these
pages do sums people act on, and a conversion that quietly damages one is worse
than no conversion at all.

## Where we've come from

Track A — the seventeen prose pages — was finished on the eleventh: every page
byte-for-byte identical to its offline prediction, and the undo mechanism repaired
and proven before it was needed (it had been silently broken in two ways since the
fifth). Along the way we filed platform bugs the work surfaced rather than caused:
every rebuilt homepage in the fleet declares the wrong canonical address, and
every assembled page loses its social-sharing tags and declares the wrong language.

Then came the finding that changed the plan. Four hours after Track A converted
the homepage, the framework rewrote its text — legitimately, that is the point of
decomposition — and in doing so stripped every card, grid and button from the
page while keeping the words. The words survived; the shopfront became a flat list
of headings. The existing safety guard measures text volume and is blind to
markup: the rewrite kept 84% of the words and 2% of the layout. So before
converting twenty-two calculator pages into that exposure, the owner directed:
build the missing guard first.

## What we've done

**The guard exists, is live, and was reviewed hard.** A "component floor" now
refuses any save that keeps a section's words but strips more than half of its
layout, calibrated against the real incident (the bad rewrite kept 2%, a good
rewrite later kept 72% — the two cases are thirty-five times apart, so the
threshold is not delicate). The reviewer council sent it back twice, and both
times it was right to. The first objection: the guard covered only one of the
nine places that write page content — including missing the section editor, the
very tool decomposition exists to enable. Checking that objection revealed the
older text guard had the same hole since the second of August. The second
objection: our own submission contradicted its own evidence. Both fixed; approved
on the third round, ten seats to one. The lasting fix is not the extra wiring but
a test: any new code that rewrites page content must either enforce the floors or
declare, in writing, why it is exempt — because when we unwired the guard as an
experiment, nothing noticed, and a guard nothing proves is reached is the same
defect one level up.

**Track B largely happened, under a better architecture than planned.** Another
session converted eighteen of the twenty-two calculator pages, moving the
calculation engines into one shared, corrected file instead of each page carrying
its own copy. The owner has ruled the calculators stay decomposed and editable —
no locks that would block editing.

**And that conversion caused exactly one real casualty, which routine re-checking
caught.** The standard loan calculator's template kept its old embedded arithmetic
— from before a bug fix — while the corrected engine sat unloaded beside it. For
about twenty hours the live page gave wrong answers at 0% interest and different
answers for the same inputs depending on the route taken to them. A routine re-run
of the arithmetic oracle caught it; the mechanism was measured, not guessed; and
the repair followed the new architecture rather than fighting it — the page now
loads the shared engine and always writes its display, so the stale-answer bug is
structurally impossible there, not merely guarded against. The whole estate now
passes 176 arithmetic checks with none failing — cleaner than before the
regression, because the shared engine matches the oracle exactly.

## Where we are now

Eighteen of twenty-three calculator pages and all eighteen prose pages are
decomposed. Every protection is live on the current build and verified in the
running binary, not inferred from a version tag. The arithmetic estate is fully
clean, with the checker's own controls run in the same sitting. The offline
prediction mechanism — the lane's safety model — survives each new build and is
re-proven after every roll.

What is deliberately unfinished: the last five calculator pages are **held**,
because the new architecture owes the owner a written ruling before they convert —
namely, what protects the calculator *machinery* now that the old row-locks are
gone by design. The floors guard both write paths, but nothing yet guards the
templates, and the standard-calc incident is the proof that a template can carry
stale arithmetic. Four other loan pages still carry their own embedded arithmetic
— passing today, but each is a second copy of the maths, which is precisely how
the regression happened.

## Where we're going

First, the ruling from the session that owns the re-architecture: what guards the
templates, the rule that templates carry wiring while engines carry arithmetic,
and superseding the now-false "tool row born locked" language in the briefs. Then
the last five conversions, under that ruling. In parallel or after: move the four
inline-arithmetic pages onto the shared engine as one small batch; fix the
fleet-wide canonical bug and then the sharing-tags bug, in that order, since the
second depends on the first; and prove the sibling lane's restored undo mechanism
with one live round trip. Further out, the second site (loancash) repeats this
whole story with its own tooling — and its complaint-deadline calculator is the
one tool on the estate whose legal inputs genuinely change over time and which
nothing yet checks.
