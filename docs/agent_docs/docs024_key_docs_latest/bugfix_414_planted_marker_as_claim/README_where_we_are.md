# Where we are — the planted "checked against the FCA handbook" claim (bug 414)

*Plain prose, append-only, newest at the bottom. This is the owner's document too — add below, never
rewrite.*

---

## 2026-08-27, morning

**What the problem actually was.** Back on 2 August someone ran an experiment on lendzy.co.uk. To
check that the site-building machinery really does what its brief tells it, they hid an instruction
in the site's brief: "somewhere in the copy, include the exact phrase *checked against the FCA
handbook, rule by rule*". The machinery obeyed — which is what the experiment was testing — and then
nobody took the instruction out. So a site about consumer credit rights has been telling readers, in
its own voice, that its content was checked against the regulator's handbook rule by rule. Nobody did
that. On a site whose entire selling point is independence and accuracy, it is the worst sentence on
the page.

It got worse on its own. Our own quality-audit machinery came along, read the sentence off the live
page, decided it must be the site's main selling point, and filed a job asking a writer to add a "how
we verify our guides" section to back it up. So the system was about to start manufacturing evidence
for something that never happened. That is the part that made this more than a typo.

**The thing yesterday's fix missed, and it is the interesting bit.** The lane that found this last
night stripped the instruction out of the brief and recorded the source as fixed. It wasn't. Ten days
after the original plant, one of our own agents — the one that writes a site's strategy — had *read*
the planted instruction and written it out again, in its own words, in a different part of the site's
records. That copy was still live this morning. So the lesson is bigger than one site: **deleting a
planted instruction from the place you found it does not retract it, because our agents copy
instructions to each other.** I have written that up as a trap for other sessions to read before they
"fix a source", and turned it into a query that takes seconds.

**What I did, in order.**

1. Removed the surviving copy from the strategy record, under a check that refused to run unless the
   exact sentence was where I expected it. Nothing was overwritten; the old version is kept.
2. Rejected the audit job that wanted to substantiate the claim, with the reason written on the job
   itself so nobody re-opens it. It was one click away from regenerating the page around the false
   claim.
3. Asked the framework to rewrite the three bits of copy — not me. The instruction to the writer says
   what to remove, why it is false, and what is *true* according to the site's own brief (we name the
   exact rule beside every figure and link to it, so a reader can check for themselves). That is a
   real claim the site can stand behind.
4. Closed the hole so this cannot happen quietly again — three small changes, described below.

**The three framework changes, in plain terms.**

The first two teach the existing honesty checks two sentences they were missing by a hair. We already
refuse copy that claims *everything* on a site has been checked — that is unverifiable by anyone,
including us. But the rule only looked 30 characters ahead, and lendzy's sentence put 38 characters
between "every figure" and "is checked", so it slipped through; and the rule didn't recognise the word
"Everything" at all. Both fixed. Worth knowing: measured across every page we serve, that rule had
been firing **zero** times — it was asleep, and these two sentences are the first things it catches.

The third change is the one that matters. Every honesty check we have looks at what the site *says*.
Nothing looked at what the brief *tells it to say*. So a brief could lawfully order a page to state
something the page checks would refuse — which is exactly what happened, for 24 days. There is one
small daily job that already reads briefs and judges them, so it now asks this second question too,
across everything any of our agents reads (not just the writer — the copy that survived yesterday's
fix was in a part the writer never sees). Where it finds one, it files it for a *person* to decide.
Deliberately no automatic handler: an automatic brief-rewriter is precisely how the audit fleet
canonised the planted sentence in the first place.

**One thing I want to be honest about: I nearly made it much noisier.** My first design ran all our
honesty rules over every brief in the fleet. I measured it before building it, and it would have
produced about 21 findings a day, essentially all wrong — because 15 of them are *our own*
instructions telling writers "never invent a person, company or statistic", and because each site's
list of banned phrases stores the banned phrase itself, so the check would have convicted every
site's own immune system, every day, for ever. Narrowed to the one family that fits, it finds nothing
today and finds exactly the two planted rows when pointed at the history. That is the version that
shipped.

**Where it stands right now.** The false claim is no longer in any brief anywhere in the fleet. The
audit job is rejected. The code is committed and has gone to the review council. The copy rewrite has
been dispatched and is retrying — the platform is having a bad morning with a particular internal
timeout, several other lanes are hitting it too, so this is queue weather rather than anything wrong
with the request. **The sentence is still on the two pages until that rewrite lands**, and I will not
call this done on a job status: I will read it off the live pages.
