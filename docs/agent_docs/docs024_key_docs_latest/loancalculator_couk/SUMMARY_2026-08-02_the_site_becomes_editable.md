# SUMMARY — 2026-08-02 · loancalculator.co.uk becomes editable

*Written to be read aloud. Previous in the series:
`SUMMARY_2026-07-30_adopting_a_hand_built_site.md` and
`SUMMARY_2026-07-31_calculators_that_prove_their_own_arithmetic.md`.*

## What we're trying to do

Take a hand-built website of twenty-seven pages and twelve working calculators,
bring it under the platform's management, and end up with a site that can be
edited, restyled and improved like every other site we run — without any of the
calculators quietly starting to produce different numbers along the way.

The owner's phrasing is the specification: the site must "evolve and improve like
the other sites will, just as long as it starts similarly enough with working
tools."

## Where we've come from

We adopted the site in a deliberately frozen state. Each page was stored as one
complete document and served back byte-for-byte, which made the adoption perfectly
safe and completely useless for the actual goal: nothing could be edited, nothing
restyled, no part of a page changed without replacing the whole thing. That was
always a holding position, and we said so at the time.

Then we rewrote all eleven distinct calculators as proper components and proved
each one produces identical numbers to the original across three sets of inputs.
That work stood on its own but had not been connected to anything — the components
sat in the database with no page pointing at them.

The open question, asked three times over two weeks, was how far to go: split the
pages fully, or take the cheaper route the neighbouring site took and simply
freeze the calculators inside otherwise-editable pages. The owner answered on
2026-08-02: **full decomposition.**

## What we've done

Every page is now stored as its constituent parts rather than as one block:
**sixty-three pieces across twenty-seven pages — fifty-one editable blocks of
prose and twelve calculators.** No frozen documents remain. Each piece can be
rewritten, restyled or replaced on its own, which is the whole point.

Three things had to be built to get there, and one of them was a surprise.

**The splitting rule needed fixing.** We had proved it safe over all twenty-seven
pages a day earlier. It still classified the homepage calculator's entire results
box — the big monthly figure and both totals — as ordinary editable text. The
reason is that the homepage, alone among the pages, keeps its arithmetic in a
shared script file rather than inside the page, and the rule only ever read
scripts written inside the page. Its own safety proof passed, and could not have
failed, because that proof asks a narrower question than the risk. It now reads
every script a page loads and refuses to run at all if it cannot read one.

**The site's furniture was broken and nobody could have known.** The header,
footer and page-head had been created by another process the previous morning, and
a rebuild of all twenty-seven pages had run against them reporting twenty-seven
successes while changing nothing at all — because frozen pages skip the assembly
step entirely. That furniture pointed at a stylesheet that does not exist (one
letter different), had a navigation bar with no links in it, and linked two
missing images. The first page we decomposed would have gone out unstyled and
unnavigable. We replaced all three, and the installer now refuses to write them
unless every file they reference actually loads.

**We had to test an assembler we could not run.** To check the decomposition
before shipping it, we needed to know what the platform's page-assembler would
produce — but that code needs a live page to run against, and creating one is the
very thing under test. So we wrote our own copy of it. That is an uncomfortable
thing to do: if the copy is wrong in the same way the test is wrong, everything
passes and nothing is verified. We handled it by making the copy write down, in
advance, the exact bytes it predicted the real system would produce, and then
comparing once the real system had run.

## Where we are now

Twenty-seven pages decomposed; the rebuilt pages are rolling out and **every one
that has landed is byte-for-byte identical to the prediction** — same checksum,
down to the punctuation inside a machine-readable block in the page head. So the
earlier result ("all twenty-seven check out, all twelve calculators produce
identical numbers") can now be read as a fact about the site rather than a fact
about our model of it. That distinction is the difference between testing
something and agreeing with yourself.

On the live pages, driving the calculators produces forty-one differences from the
original record and **not one of them is a changed value**. They are all the same
controls being named rather than counted: the originals had no identifiers on
those inputs, so they could only be referred to by position. That is why three of
the tools have never had any numeric acceptance coverage — you cannot assert
anything about "the fourth input". They can be named now.

Two visible changes, both deliberate. Every page gains a footer, because the legal
page had no links to it from anywhere on the site. And the homepage gains the
"late repayment can cause you serious money problems" warning — we tried to strip
it for fidelity and our own tooling refused, because the warning was marked
required when the component was built, on the grounds that it belongs beside a
credit promotion. It was right and we were wrong. The two dated market claims were
*not* carried across, because duplicating a figure that goes stale just doubles
what has to be corrected later.

Everything is reversible one page at a time, and every original is preserved in
the database.

## Where we're going

Finish the rollout and verify the remainder, then capture a fresh acceptance
baseline — which for the first time can drive every control on every calculator,
including the three that have never had numeric coverage. The old baseline is
kept, not replaced: it is the only record of what the hand-built site computed,
and every equivalence claim we have made is stated against it.

After that the queue of real defects the rewrite surfaced and deliberately did not
fix: money shown to three decimal places on one tool, another that computes
nothing at zero per cent interest, a consolidation checker that counts a debt
towards a balance but not towards interest, and a verdict distinguished only by
colour. Each needs its own baseline refresh, which is now cheap.

Two things still need the owner. The **GitHub token cannot see the repository
holding the site's source**, which needs someone with admin. And the site is still
marked as ours-to-hold rather than ours-to-rebuild — flipping that is what lets
the improvement loop touch it, and it is a separate, deliberate decision now that
the pages are actually in pieces.

The neighbouring sites, loanandmortgagecalculator.co.uk and loancash.co.uk, are
where this site was a week ago and have **no furniture at all**. If either is
decomposed without building it first, the platform's fallback links a stylesheet
neither of them serves and every page goes out unstyled — silently. Both lanes
have been told, with the measurements.
