# SUMMARY — 2026-08-11: Track A complete. The prose pages belong to the framework now.

## What we're trying to do

Take a site we built by hand and hand it over to the framework, page by page, without
anybody noticing on the way. Every page on loanandmortgagecalculator.co.uk was
adopted as a single frozen lump of HTML — one row in the database holding an entire
document. The framework could store that lump and serve it, but it could not *edit*
it, because it could not see inside it. Decomposition is the work of splitting each
page into real components the framework can address: the words in one piece, and on
the calculator pages, the working calculator in another, protected piece.

The prize is control. Once a page is components, we can rewrite its copy without
touching its calculator, rebuild it from a plan, and stop hand-editing files in a
repository. The risk is that this site does consumer-finance sums that people act on,
so a conversion that quietly damages a page is worse than no conversion at all.

## Where we've come from

The site went live and was adopted in late July as forty-one frozen pages. Since then
the work has been about making it safe to touch: proving the calculators are
arithmetically right rather than merely unchanged (an independent oracle, which found
a stamp-duty rule sixteen months out of date and under-quoting by £5,000), and
building the tooling that splits a page apart and can put it back.

Two pages were converted on the fifth and sixth of August as canaries and left to
prove themselves. The plan then split the remaining work into three tracks by risk:
the seventeen prose pages first, the twenty-two calculator pages second, and a second
site last. Immediately before this session, a check caught that six live calculators
had been mistakenly unlocked for about seven hours — found by cross-checking against
a hand-written list rather than trusting the automated classifier that had caused the
problem in the first place.

## What we've done

All seventeen prose pages are converted and live. Every one serves bytes identical to
what we predicted offline before touching anything — not approximately, identically,
on all seventeen. The site has no frozen pages left outside the calculators.

We did it in an order chosen to make failure cheap: the smallest, least-visited page
first; then one page of each of the two layout shapes, each proved live before the
rest of its shape followed; the homepage last and alone.

Three things were found along the way that mattered more than the conversion itself.

**The undo button did not work.** Every page's previous state is copied to a backup
before it is replaced, so any page can be restored. That backup could not run at all —
the main table had gained a column since the backup was created, and the copy failed
outright. It failed *before* changing anything, which is the only reason this was an
inconvenience. The worse half was underneath: the backup copied in any row it had not
already seen, so once a page was converted, the next run copied its **new** content
in beside its old. Restoring such a page would have put both versions on it at once —
producing exactly the corruption the process exists to prevent, delivered by the
safety net. It had already happened to one page on the fifth. Converting seventeen
pages one at a time would have silently ruined the undo for about sixteen of them.
Both faults are fixed, the damaged backup is repaired, and the restore was then
proved by converting a page, putting it back, and confirming it matched the original
exactly. The same faults are still live in the sibling site's copy of the tool, which
is written up as a bug.

**A check that passed for no reason.** The single most important safety property is
that none of the seventeen pages is a calculator. The check confirmed it perfectly —
and was worthless, because the two lists it compared write page names differently and
could never have overlapped whatever the truth was. It was caught by one extra line
that was also supposed to be empty and came back full.

**The homepage now names the wrong address.** Our checker flagged that the converted
homepage tells search engines its canonical address is `/index.html` rather than the
plain domain. Checking ten other sites first showed nine already do the same: it is
long-standing platform behaviour that this work surfaced rather than caused. It is
filed with its cause and fix.

## Where we are now

Seventeen of seventeen live and byte-exact. The arithmetic is untouched and proved so
with its controls run in the same sitting — 170 checks pass, none fail, and the
deliberately-broken control run does fail as it must. The component byte gate passes.
Reader-visible text was compared, page by page, between what was served before and
what is served now: identical on all seventeen, so nobody reading these pages sees a
changed word.

Two accepted losses are now real on nineteen pages rather than two: the per-page
social-sharing tags, and the page's language tag narrowing from British English to
plain English. Both were written down and accepted on the fifth of August when they
applied to two pages. They are worth restating at this scale rather than letting a
decision made when it was small carry all the way to fifty-nine pages.

Honest about the size of the prize: each of these pages became **one** editable block
of text, not paragraph-level sections. The framework can rewrite that block; it
cannot yet act on "the third paragraph". And these pages still share a template with
twenty-eight others across two sites — better than the hundred and fifty-four they
shared before, but not isolation.

## Where we're going

Track B: the twenty-two calculator pages, and a genuinely different risk class,
because those pages compute things people act on. One question must be answered
before it goes wide, and Track A could not answer it: what happens to a locked
calculator component when its page is rebuilt. None of these seventeen has a locked
component, so nothing here exercised that path. It is reasoned about but unmeasured,
and it should be measured on a single calculator page, watching for the calculator
landing at the bottom of the page and for the one non-silent signal that would say so.

Before that, two smaller debts: the sibling site's copy of the backup fault, and a
decision about forty dormant rebuild requests sitting against this site which are
harmless while nothing promotes them, but which now point at pages that would rebuild
rather than refuse.
