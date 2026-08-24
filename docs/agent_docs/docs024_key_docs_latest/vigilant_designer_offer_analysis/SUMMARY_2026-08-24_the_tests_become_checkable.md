# SUMMARY — 2026-08-24: the analyser's own tests become checkable, and three pages we had called finished are not

## What we are trying to do

Give every site a written, ranked answer to "what should this page lead with, and what should it
never lead with" — derived from the site's own recorded strategy rather than from anyone's taste —
and then turn the gaps between that answer and the live site into work the platform does by itself.

## Where we have come from

A fortnight ago this was an argument about copy. It became a premise record per site, then an
analyser that reads it, then findings that flow into the ordinary work queue. On 15 August the owner
enrolled it in the automatic improvement sweep; by the 17th it was choosing sites on its own and its
findings had already turned up a platform fault. The last read-out ended by naming the next build:
*the analyser writes a test alongside each finding saying what "fixed" would look like, and nothing
reads it.*

## What we have done

**Two things, and the second is the one worth your attention.**

**First, we closed the honesty problem we found in our own machinery.** The analyser had written, at
the top of one site's ranked record, a confident sentence containing a number that was simply not in
the source it named — and the field that exists to prove where a claim came from vouched for it. Both
halves are now fixed: the instruction that stops the model writing such a sentence, and a mechanical
check that removes one if it does. It is live, and every one of the five enrolled sites now carries
the check's own record of what it removed — which is nothing, on all five. That last part is the
honest catch: the safety net has never actually caught anything, because the instruction prevents the
sentence in the first place. It is proven by tests, not in the wild, and nobody should point at a
clean run and call that proof.

**Second — and this is the finding — we went to see whether the analyser's tests were worth making
checkable, and discovered we have been closing jobs that did not do what they said.**

Every finding comes with a test in plain English: *the description must state the no-account promise
before any mention of how many tools there are.* Nothing has ever read one back. So we read all
thirty-seven we have written and went and looked at the pages.

Three are sitting on jobs marked **done** — page rebuilt, deployed, closed — where the criterion the
job itself stated is still not met, and can be shown false in one line. The clearest is our own
webdesign.co.uk front page: the description we are serving right now opens *"Sixty-three browser
tools…"*, with the count first and the promise after. It has been rebuilt twice since that job closed
and still reads that way.

Nobody was careless. There is nothing in the platform that reads the test back after the work is
done, so "the handler finished" and "the page now passes" are the same word.

**So a finding can now carry a small checkable condition next to the sentence** — *this phrase must
be absent*, *at least two of these three must appear*, *this must come before that* — over a page's
title and description only. Two rules keep it honest. It can only ever say **no**: passing means "not
caught", never "fixed", because two thirds of these tests bolt a checkable clause onto a matter of
judgement and a tick against the easy half would be worse than no tick. And it must **already fail
today**, or it is discarded and the reason written down — which throws out the useless case, a rule
about a word that appears nowhere, that would pass for ever and look like verification.

It went through the review council and was approved. Two reviewers independently found a hole I had
not: if the page lookup ever stopped matching, every condition would be refused and the whole thing
would go quiet wearing exactly the face of its acceptable outcome. That failure is now loud.

**One thing nearly went the other way, and it is the reason to trust the rest.** My first version also
checked the header menu, and it "found" a fourth broken job — our database says thirteen menu items
where the test allows seven. Then I loaded the page: the menu shows seven. The database column is not
the menu. The test had been passing all along and I had a check ready to declare it broken with an
air of arithmetic. Menus are out of the feature entirely, and the trap is written down.

## Where we are now

**Five sites of twenty-eight carry the ranked record.** Last time that read five of twenty-three: our
coverage has not moved and the estate has grown by five, so the gap widened while nobody did anything
wrong. Opening the sweep again is a cost decision and stays yours — roughly a site every fifteen
minutes, and every visit is the expensive kind.

The honesty check is live and proven across all five. The new checkable-condition work is built,
tested and reviewed, but **not switched on**: it needs a rebuilt service first, and two deployments
today were both built from a point before it. Turning it on before that would break the analyser
rather than degrade it, so it is deliberately held with the instructions attached.

The three jobs that closed without meeting their own criterion are still live, and still wrong on the
page.

## Where we are going

**One: switch it on, and find out how often the analyser actually reaches for it.** Zero is a possible
and acceptable answer on the first run — we invited it in prose rather than putting it in the required
output, so silence is the safe direction. What matters is the count, either way.

**Two: the piece that would actually stop a job closing while its own test fails.** Today this makes
the test checkable; it does not check it at the moment of closing. That change touches handlers other
lanes own, so it deserves its own review rather than riding in on this one.

**Three: claim-checking the premise records themselves** — approved on 14 August, still not designed.
Today's work checks whether a page matches what we said about it; that one asks whether what we said
was true in the first place.

**Four: coverage.** Everything above improves five sites out of twenty-eight. A tie-break another lane
has asked us for is worth little at five and a real mechanism at twenty-eight, and the same is true of
all of this.
