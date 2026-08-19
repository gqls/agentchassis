# SUMMARY — 2026-08-19: the first sites are live, and we found the wall the fleet would have hit

*Written to be read aloud. Supersedes nothing — `SUMMARY_2026-08-17` remains the record of what
we believed two days ago, and one of its central claims has since been overtaken.*

## What we're trying to do

Turn a portfolio of about 140 owned domains into a fleet of genuinely different, framework-built
sites — each one positioned so it does not compete with its neighbours, each built by the
pipeline rather than by hand, and each one carrying the compliance guards that finance content
needs. The measure of success is not that a site exists; it is that a stranger could tell the
sites apart, and that nothing on them is asserted without a citation.

## Where we've come from

Two days ago we had proved the machinery worked. The pilot, `remortgagecalculator.uk`, had been
built end to end by the pipeline, and the directory mechanism — a shared, cross-site register of
facts about outside firms, where every fact carries a verbatim quotation — had been proved on a
real site for the first time. But the pilot was **built and not published**: deploy failures
meant nothing was actually live, and a queue of items was waiting for a person. "The machinery
works" and "the site is finished" were two different statements and only the first was true.

Since then: the owner set the Nominet nameservers, a second site was dispatched, and the owner
halted the builds pending two decisions about how sites get built at all.

## What we've done

**The pilot went live on its own domain.** Verified by reading the page, not the status code —
a parked domain answers 200 on every path, so the status code proves nothing.

**We found out why it had no calculator.** The owner noticed that a site whose entire
proposition is a calculator was serving without one. The cause turned out to be structural and
it is now `bugs_open/311`. Every component in the shared library has two names: a *type* name
used when the system looks for something to reuse, and a *function* name used when it saves one.
For the mortgage calculator these differ — it has a function name and no type name. So the
search found nothing ("build a new one"), and the save then found the existing one ("this is a
replacement") and correctly refused, because overwriting it would have blanked another site's
live pages. Three attempts, three refusals, and the page deployed one section short, looking
finished. We put it through the independent diagnosis loop before asserting any of it; it came
back confirmed on the first pass.

**It is not a one-off, and the scale is now known.** Another lane reproduced it the same day on
a greenfield site that lost **seven of seven** tool sections, and measured the result at the
served page: a live calculator page containing zero input boxes. Twenty-six components in the
library are in the same state, waiting for the next site.

**We widened the directory from 2 lenders to 25.** The owner asked whether the search should be
widened. Rather than guess, we asked the register which *sources* had ever produced facts, and
the answer redirected the work: twelve of our thirteen savings providers came from a single
gov.uk page, and all ten health insurers from two broker round-ups. Every productive source is a
page that lists many firms and says something about each. The mortgage search had never landed
on one. Worse, the extractor's own instructions called such pages weak sources — the opposite of
our own history. We corrected that, raised the pages read per search from four to ten, and added
four searches aimed at the right shape. Four runs, about fifteen minutes: 42 new facts, 25 named
lenders, including the specialist adverse-credit lenders our second site had nothing to list.

**`www` now works across the whole estate.** Typing `www.` in front of any of our addresses now
takes you to the site rather than a browser error, on 36 of the 36 domains where that makes
sense, each one verified by following the redirect. Two domains that had been quietly broken for
months were fixed as a by-product.

**An owner ruling was recorded**: we do not create sites that present themselves as accredited
finance brokers unless asked. The first `loanzy.uk` build was cleared on that basis.

## Where we are now

**Builds are halted, deliberately, and the queued work is preserved.** Two coupled decisions are
waiting: which build flow we use, and whether the classifier is given the positioning register
to read. Nothing should be unlocked before those are answered.

**Two sites are live; neither is finished.** The pilot serves, but without its calculator and
still listing 2 lenders rather than 25, because it is locked and nothing will refresh it. The
second site, `adversecreditmortgage.co.uk`, was dispatched and halted mid-build with 41 items
held.

**The most important thing we know today is a thing we did not know on Monday**: the fleet plan
has a wall in front of it. About 140 finance domains will want the same calculators, and as the
platform stands, whichever site creates a given calculator first owns the name and every later
site ships that tool hollow — with nothing in the page to show a reader that anything failed. A
neighbouring lane hit the same design fact from the other end and wrote out the remedy; we
established this morning that their fix, as scoped, would repair their half and leave ours live.
Neither half is built.

The honest read-out is therefore: **the machinery is proved, the first sites are live, and the
thing that would have quietly degraded the next fifty is now understood but not yet fixed.**

## Where we're going

Three things, in order.

**First, the two halted decisions**, because everything else waits on them and one of them got
larger: clearing the loanzy build established that an automatic seed alone is not enough — it
would supply the guards and still let a site invent a regulated identity.

**Second, the component collision.** It is a change to shared machinery every build passes
through, so it needs a review round and a chassis roll, and it is best done once for both
writers rather than twice. Until it lands, any site built on a shared calculator name will serve
a hollow tool, so it is properly a precondition for the fifty rather than an improvement to
them.

**Third, resume the builds** — starting by finishing the two sites we already have, which means
refreshing the pilot's lender page and releasing the 41 held items on the second site.

Nothing in the fleet plan changes. What changed is that we now know what to fix before running
it fifty times.
