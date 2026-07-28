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
