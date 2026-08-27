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

---

## 2026-08-27, mid-afternoon — it's off the site

**The false sentence is gone from lendzy.co.uk.** Both pages read clean, and I checked it four ways
rather than once: the database holds no trace of it in either place it was stored; both pages fetched
from the live site contain zero occurrences; and — the one I trust most — our own honesty scanner, run
over the whole site, now reports nothing at all, having convicted three components on that same site
this morning. A clean result from a tool that was finding things a few hours ago means something. A
clean result from a tool you have never seen fire does not.

**The framework wrote better copy than the sentence it replaced**, which I did not expect and think is
worth recording. The guide now says that every figure and rule reference *is given* together with the
named rule and a link to where you can read it — and then adds, unprompted, "that does not make the
checker infallible… rather than take our word for it." We removed a claim asking readers to trust us
and got back one inviting them to check. That is the outcome the brief always implied, and it came out
of the machinery rather than out of me typing it.

**Two operational notes, because both cost me time and would cost the next person the same.** The
about page was repaired for three hours before I noticed, because its job was marked "failed" — the
step that failed runs *after* the step that saves, so the work had landed and the label was about the
wrong thing. And the rewrite of the second page was stuck for two hours on an account-level outage in
the AI service, not on anything to do with this bug; the owner adding credit cleared it, and it
finished on the next attempt.

**Why I am not calling this closed.** The bad sentence is gone and that half is genuinely done. The
part that stops it happening on the *next* site is code, and code on this system does nothing until
the whole fleet is next rebuilt. Until then, a planted instruction on another site would sail through
exactly as this one did. So the file stays in the open queue with three specific things written down
for whoever picks it up after the next release — including the right way to check that the running
service actually has the new rules, because the method I had originally written down was the one
documented as unreliable for this particular service, and a reviewer caught that before I used it.

**One thing I would like you to look at when you have a moment.** One of the review seats disagreed
with the decision you took this morning. The new rule about claiming we have checked something against
a regulator's handbook ships as a *warning* rather than a *refusal*. The argument for warning still
holds — a real compliance consultancy could say it truthfully, and at refusal-strength the layer would
have blocked the honest correction "nothing here has been checked against the handbook", which our
negation detection cannot see. But the seat whose entire job is this class of claim would have refused
it, on a finance site, where the false version ran for 24 days. It is on the record so you can revisit
it deliberately rather than have it quietly settle.

**And a new hole, found by that same seat, which is not this bug but is next to it.** The register of
verified facts is deliberately not scanned by the new check, because it stores the banned phrases
themselves as data. Which means a *poisoned register* — a fabricated source or a made-up fact written
into it — passes every layer we have, because every layer treats the register as the thing it checks
against rather than something to be checked. Nobody has found an instance. If one turns up, it wants
its own file.
