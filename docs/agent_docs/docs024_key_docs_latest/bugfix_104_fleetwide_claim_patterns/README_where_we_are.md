# Where we are — fleet-wide claim patterns (`bugs_open/104`)

Plain prose, append-only, newest at the bottom.

---

**2026-07-28, session "bugsearch 6".**

I picked up bug 104. The short version of it is this: we have a scanner that
stops a site publishing a claim we know to be false, it works, and it is pointed
at almost nothing. The patterns it checks are stored per site, so a lesson learned
on one site cannot reach any other, and every new site starts with no protection
at all. Seven of our fifteen live sites have some patterns. Eight have none —
including vetcomparison, the site that published invented prices for three
thousand real vet practices, and idea.uk, the only site taking real money.

That much was already written down, twice: once as bug 104 and once as oufe's
decision O11. Both say the recommended fix is small — copy the trick we already
use for the "AI-sounding phrases" checker, which keeps one shared list everybody
gets plus whatever each site adds itself. Both also say this is the owner's call
rather than a thread's, because it changes what every site's build is allowed to
block.

So the useful thing I could do was not to decide it, but to find out what it would
actually cost — because our own rules say measure the blast radius yourself
rather than hand the question to a reviewer. There is already a tool for this
(`claimscan`) which runs the exact same checking engine as the live build gate, so
I ran the ten proposed patterns against every page of all fifteen sites — about
nine hundred pieces of page content.

**It would have broken three sites, and most of what it blocked would have been
us telling the truth.**

Seven pages would be flagged. Four of the seven are honest sentences that happen
to contain the banned words in a *negative* form. robot-hands says "where
manufacturer data has **not** been independently verified, that is stated
explicitly". vonc says a comparison is "Spark's own assessment, **not**
independently verified". These are exactly the careful, hedged sentences we want
sites to write — and the pattern would have refused to build the page for writing
them. One single pattern out of the ten causes six of the seven hits and all four
of the false ones.

I want to be clear that this is not a criticism of the work that produced those
patterns. They were tested carefully before going live on oufe — ten bad sentences
blocked, thirteen good ones allowed through. The gap is subtler than carelessness:
the thirteen good sentences did not include one that *contradicts* a banned
phrase. Nobody thinks to test "the opposite of the thing I am banning", and on one
site it never came up. Across fifteen sites it comes up immediately.

There is a deeper reason this was waiting to happen. The claims scanner has no
false-positive protection at all, and that was a deliberate choice with a good
reason written next to it — every pattern was read and approved by a human *for
that specific site*, so a match is a known falsehood, and blocking the build is
right. The moment the patterns become shared across every site, that reason stops
being true, because nobody has read them against the other fourteen sites' copy.
The safety net was the human review, and going fleet-wide quietly removes it
while keeping the consequence, which is a failed build.

The good news: nothing is broken right now. I checked every site against its own
current patterns and not one page fails. This is a trap in the proposal, not a
live fault. And the underlying complaint in bug 104 is still completely valid —
eight sites genuinely are unprotected, and the two most exposed are among them.

So what I need from you is the O11 decision, but with better information than the
bug file had. The recommendation has changed: the fix is still the right idea, but
it is no longer "small, precedented, one roll". Either we drop the one bad pattern
and ship a smaller shared set that measurably fires on nothing, or we do the fix
properly and teach the scanner to ignore negated sentences — which is real code,
because the regex language Go uses cannot express "not preceded by 'not'", and
nothing else in the estate does this yet. I have costed both and will put them to
you next.

One process note worth recording: I got three things wrong while measuring this,
and all three looked like clean results. A loop that silently stopped after one
site; a word-count that searched for a string the tool never prints, so every site
read as zero findings; and an error I had hidden from myself, which turned a
momentary database hiccup into "this site has no patterns" — about the one site
that mattered most. The thing that caught the second one was deliberately feeding
the checker sentences it *must* reject and confirming it did. That test is the
only reason this session found anything at all.

---

**2026-07-28, later the same evening.**

You made both calls and I built it. Nine patterns are now shared across the whole
estate rather than living on whichever sites somebody remembered to arm, and they
apply even to a site that has no fact register at all — so vetcomparison and
idea.uk are covered, and so is the sixteenth site on the day someone creates it.
That was the whole point of the bug.

I checked it the way I'd want it checked: the finished code, not the proposal,
against every page of all fifteen sites. Nothing on the estate would fail a build.
The checker still catches all six of the fabrication shapes I tested it with, on a
site with no register whatsoever, and the four honest sentences that the original
patterns would have blocked now pass — and they're committed as tests, so if
someone puts that pattern back, the test suite stops them rather than a live site
build stopping a colleague.

One thing changed my mind while building. The bug file recommends copying how the
"AI-sounding phrases" checker does this, which is to merge the shared list into
each site's register as it's read. **That would have quietly written the shared
patterns into all fifteen sites' stored data**, because two other bits of the
system write that register back to the database after reading it. The warning was
sitting in the same file, two hundred lines above, about a different field where
someone nearly did the same thing. So the shared list is kept separate and joined
only at the moment of checking. Same result, nothing gets written anywhere.

Now the part I'd rather report than have you find. **I broke the build for about
four minutes.** The two files I edited were also being edited, right then, by
another session working on a different bug. When I committed my files by name, I
took their half-finished work with it — specifically, I committed the code that
*uses* a new thing while the code that *defines* it was still unsaved on their
side. So the shared repository briefly didn't compile, and anyone starting a
build in that window would have seen a failure with my name on it. They committed
their half four minutes later and it's whole again; I checked it properly this
time, by exporting the repository as it actually stands rather than testing my own
copy. It's written up in the wrong-calls log, because my commit message claimed
the tests were green — which was true of my copy and false of what I'd committed.

The honest lesson is small and annoying: our rule for keeping sessions out of each
other's way works on every file except one that two people are editing at the same
moment, and nothing can detect that case. What I should have done is check that
the shared repository still compiled straight after committing. That's one
command, and it's now written down in the runbook.

Two smaller corrections, both mine, both caught before they could mislead anyone:
I claimed the reduced pattern set "fires on nothing" before running it (it fires
once, on a real overclaim), and two of my test examples were sentences I'd retyped
from a truncated display rather than copied from the site — so they read as
quotations of copy that was never published. Both fixed, both recorded.

It is not finished. Code like this does nothing until the next time someone builds
and deploys a chassis image, so the bug stays open until then, with the exact
checks to run written into it. And the strongest of the original ten patterns is
deliberately left out — it's the one that caused the false alarms — so someone will
need to teach the checker to tell "this is verified" from "this is not verified"
before it can come back. That's a separate piece of work and nobody owns it yet.
