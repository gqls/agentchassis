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

---

**2026-07-29.**

The chassis was rebuilt and deployed, so the change is now actually doing something.
I checked it in the running pod rather than trusting the version number — three
distinctive phrases that only this change puts in the binary are present, a control
phrase that was always there is present, and a nonsense phrase I made up is absent.
That last one matters: it proves the search itself works, so "found nothing" would
have meant something. **The fleet-wide honesty check is live on all fifteen sites.**

While I was in there I ran an unrelated check that's owed after every single deploy —
a different bug's fix depends on the chassis being new enough, and if it isn't, a
whole diagnostic lane stops working silently with nothing in the logs. It's fine.
Nobody from that workstream saw this deploy happen, so somebody had to look.

I also wrote a proper test for the thing that actually mattered and wasn't covered.
All my earlier tests proved the *patterns* match the right sentences. None proved
that **the build gate bothers to check a site that has no register** — which is the
entire point of the bug. It now does, tested five ways: the overclaim fails the
build, and four kinds of legitimate sentence still pass.

Writing that test taught me something I had wrong. When the gate finds a blocker it
doesn't return a list of problems — it returns an *error*, and that error is how the
build gets stopped. My first version of the test treated the error as "the test
failed", so it reported a failure on precisely the outcome we want. Worse, if I'd
patched over it carelessly the test would have passed whether the gate worked or
not. What caught it was reading the failure message, which said "1 blockers" — the
gate telling me it had done its job.

**The review came back "revise", not "approved", and I want to be straight about
why.** The guardian reviewer — whose job is to object when a change touches shared
machinery — said the honest thing: this alters the contract of a check that runs on
every site's build, and no amount of "but the measurements are clean" changes that.
It stopped short of vetoing, partly because of how thoroughly it was measured, and
partly because at the time the change hadn't shipped yet. It has now, so I've told
it that plainly in the resubmission rather than letting it find out. It asked one
fair question — was your owner's decision actually reviewed by anyone, or are you
asking us to rubber-stamp it — and the answer is that you ruled on it out of band,
with the numbers in front of you, on a question that had been sitting filed and
unanswered for months. That's not a rubber stamp, but it isn't this council's review
either, and it deserved a straight answer.

One thing I'd flag as a problem with the review system rather than with us. The
council runs its own database queries to check the author's claims, and one of them
came back saying there were **zero** live pages in the estate — which would mean my
entire 908-page measurement was invented. It's wrong: it filtered on a site status
value that doesn't exist here, and the council's *own* next query printed the real
statuses that prove it. Re-run properly, the number is exactly 908. But a reviewer
reading "zero" would have been quite reasonable to conclude I'd made the whole thing
up, and nothing in the process checks the checker. I've put the corrected query in
the resubmission so this round doesn't have to work it out.

Where it stands: the bug is fixed and live and now properly tested. The review is
mid-second-round. If the guardian still thinks the scope was wrong after getting
straight answers, that's a judgement for you rather than something I should argue
into submission, and I'll bring it to you if so.

---

**2026-07-29, later the same day.** I went back and did the one thing this job left
owed, and it turned out to be more interesting than expected.

Recap of where it stood: we now have a set of ten "no site may say this" patterns, but
only nine were switched on. The tenth — the one that catches a site claiming its
information has been "independently verified" — was left out, because when I test-ran
it yesterday it flagged four sentences and all four were **honest** ones. Sentences
like "where manufacturer data has *not* been independently verified, we say so". The
pattern can see the phrase but not the word "not" in front of it, so it was failing
pages for being careful, which is the exact opposite of the point. I wrote a note in
the code saying it must not come back until something in the code could tell the
difference.

That something now exists. It looks backwards from the phrase for a "not", "never",
"cannot", "doesn't" and so on, and stops looking at the first comma — so a denial in
the same breath counts, but "we do not use AI, **and** every claim here is verified"
still gets caught, because the denial is about something else. All four honest
sentences now pass, and they pass for the right reason rather than by accident.

**Two things I got wrong, both caught by measuring rather than by thinking harder.**

First, I tried to be clever. I worried the pattern would also flag a legitimate
sentence like "our accounts are independently audited" — true of plenty of real
businesses and none of our business — so I narrowed it to only fire when it was talking
about claims, figures, prices and so on. Then I ran both versions over every page we
have, and my clever version found **nothing at all**. Zero. The plain version found
**two real problems**. So my narrowing hadn't made it safer, it had switched it off,
and a switched-off check looks exactly like a check with nothing to find. I threw my
version away and wrote down the trap instead.

Second, and more embarrassing: I "confirmed" that our page count still matched
yesterday's figure by adding up a column of numbers in my head and getting the answer I
expected. It's 919, not 908. It had grown by eleven overnight and I'd manufactured the
agreement rather than running the one-line query sitting right there.

**The finding that actually matters to you.** Yesterday I told you this scan came back
completely clean across the whole estate. That was true of the nine patterns switched
on — and it was **clean because the interesting pattern was switched off**. With it on,
two pages on robot-hands.com claim the gripper specification data is "independently
verified". Nothing independently verifies it; we scrape it from manufacturer
datasheets, and that site's own records have no entry saying otherwise. Worse, the same
site promises elsewhere that it labels unverified data — so it's contradicting itself on
its own pages.

I have not touched the copy. That site belongs to another line of work and rewriting
someone else's site voice isn't mine to do; I've written the case up with the exact
sentences and three suggested fixes, and left a note where that lane will see it. The
practical consequence is that those two page sections **won't rebuild** until the wording
changes. Nothing goes down — the live pages carry on being served — but anyone who tries
to regenerate them gets a clear refusal explaining why.

So: the honest read is that we've swapped a comfortable "zero problems" for a real
"two problems", and the two were always there. That's the check earning its keep. The
review board is looking at it now.
