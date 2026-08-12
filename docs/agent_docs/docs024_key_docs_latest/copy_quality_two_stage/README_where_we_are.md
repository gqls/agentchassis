# Where we are — the copy quality problem, in plain prose

Append-only, newest at the bottom. The owner maintains this too.

---

**2026-08-12 — why this thread exists, and what I got wrong to start it.**

You read the new homepage copy and took it apart: the title was negative and read
like machine-written filler, the opening fired unwelcome consequences at the reader
one after another, and the second paragraph talked about the website — how many
calculators it has, that it needs no sign-up — when the reader had come to work
something out. You suggested splitting the job in two: one pass to get the facts
down, a second to reorder and rewrite them so a person can read them, possibly
using the offer analysis work to decide what deserves to come first.

The first thing to say is that the copy you rejected was **my instructions being
followed**. I had written the page a brief, and my brief contained that "23
calculators, no sign-up, runs in your browser" material as facts to convey, and it
described the subject in exactly the losing-out terms you objected to — a car loan
means a lender offers you less. The writing agent rendered what it was told. So this
was not the writing model having a bad day; it was the person writing the brief
embedding their own framing and not being able to see it, because from the inside a
framing looks like the facts.

That is the strongest argument for your two-stage idea, and it shapes the design in
one specific way: **the second stage must not be shown the brief.** If it reads the
brief it inherits the framing, which is the thing being fixed.

The second thing worth knowing is that half the machinery already exists and is
going to waste. There is already an agent that judges copy for tone, for gaps, and
for whether a page actually differentiates itself. It runs, it forms an opinion, and
**nothing anywhere consumes what it says** — I checked the whole history of the
system and it has never produced a single piece of follow-up work. Meanwhile the
design side of the house has exactly the missing piece: when a design audit finds a
problem, there is an agent whose whole job is to apply the fix. Copy has the
auditor and no fixer. That asymmetry is the gap, and it is a much smaller thing to
build than "teach it to write better".

There is also a sister case. The mortgage calculator site got the same complaint
from you on the same evening, and that thread has already written a very good voice
specification in response — including a rule about not stacking short clauses,
because it reads like a machine gun rather than a person, and one that ends "read it
aloud; if it sounds like a brochure describing itself, write what you would actually
say." Two sites, two threads, the same root cause found separately a day apart. I
have deliberately not copied their specification over the top of this site's own —
this site's was chosen by you a week ago and I am not going to overwrite a decision
you made without asking — but their rules are the raw material for the checkable
version.

For now: the immediate page has been re-briefed against each of your three
objections and is being rewritten as I write this. The lane has its own directory so
it stops squatting on the calculator site, and you were right that this is a good
place to test — it is quiet, and every round so far has taught us something.

**Three things I need you to decide when you have a moment** (they are in the plan
in more detail): whether the second pass should run over every page or only ones
flagged as poor; who wins when the editorial pass disagrees with the voice
specification you approved; and whether "reads well" can be graded mechanically at
all, or stays a human call. My instinct on the last one is that a handful of things
you have already rejected can be turned into real checks — no heading built on a
negation, no site-inventory in the opening, no run of clipped consequence sentences
— and that the rest stays judgement.

---

**2026-08-12, later — I went looking for what we'd already tried, and most of what I
told you this morning was wrong.**

You asked me to start this properly, so before designing anything I searched the whole
documentation tree and the code for earlier work on voice and tone. There is a lot, and
it changes the picture. Three things in particular.

**First, and this is the one I'd act on today: a decision you made three days ago was
never carried out.** On the 9th you decided that the "gentle explanatory" voice — the one
that thread called H — becomes the default for the whole fleet, and that one specific rule
should change form. The old rule tells the writer *"start with the fact"*. That is an
instruction about how to open, and the same thread had just proved by experiment that
telling a writer how to open makes a hundred sections open identically, which is a large
part of what "machine-written" means. Your decision replaced it with a prohibition and an
instruction to vary.

I checked all seven writing prompts this afternoon. **All seven still carry the old rule.
None carries yours.** And one of the seven is the exact agent that wrote the page you
rejected yesterday. So some part of what you were reading was a rule you had already
ordered removed. That is not a design problem, it is an undelivered decision, and it is
the cheapest thing on the list.

**Second, I was wrong that the quality auditor's work goes nowhere.** I told you this
morning that we have an agent which judges copy and that nothing anywhere consumes its
opinion — that I had checked the entire history and it had never produced a single piece
of follow-up work. That was a bad check. I searched for it under a name it has never been
filed under. It has in fact run thirty-four times, the most recent one today, and its
findings do turn into work, and there is already an agent whose job is to apply them.

What actually happens is stranger and much smaller: the auditor's findings get filed under
the *design* audit's name rather than its own, because of a default in the code that
overrides what the configuration asks for. So the copy auditor's work is real, it is being
done, and it is invisible to anybody looking for copy work. That is a small bug, not a
missing machine, and I have it written down as one.

**Third, the earlier thread already ran the experiment we were about to repeat.** Over
four days in early August it measured our copy three ways across six sites and nine hundred
sentences, found nothing, watched you find a fault in about a minute, then tried fixing it
by tuning the rules — and recorded, honestly, that almost all of it failed. The findings
are uncomfortable and worth having. Rules can only name a shape, and what goes wrong is an
instinct, which turns up next time in different grammar. Rules that are cheap to satisfy
crowd out the ones that need judgement. Rules sometimes *cause* the fault, and it then
arrives pre-approved so nothing questions it. And the strongest one: **a writer copies your
examples far more reliably than it follows your rules** — they deleted a rule, left its
worked examples in place, and the behaviour didn't budge.

There is also a finding that lands directly on your two-stage idea, and I want to flag it
rather than bury it. **The writer only ever sees one section of a page at a time, never the
others.** That is why a page can call the same thing by two different names in two
sections, which happened on a live test. Your second pass is supposed to ask "is the most
useful thing first" and "does this talk about the reader or about us" — those are questions
about a whole page. So the second stage has to *read* the whole page even though it should
only *write* one section at a time. My plan this morning had it reading one section, which
would have rebuilt the exact blind spot that caused the problem.

One more thing you should know before deciding anything. In a test on the 9th, two
different versions of the writer both tried to overwrite a paragraph you had personally
approved, and the only thing that stopped them was that the section was locked — not the
instructions, not the care taken writing them. So whatever we build, approved copy has to
be protected by being out of scope, not by being asked nicely.

**What I think we should do, in order.** Ship your voice decision into the seven prompts,
and change the worked examples at the same time, because on the evidence the examples are
what the writer actually obeys. Fix the small attribution bug so copy findings can be seen.
Turn on the claims checker for the calculator site — it is built, live and approved, and
that site simply never opted in, which is why nothing checked the new figure that appeared
in yesterday's rewrite except a query I wrote by hand. Only then build the second stage,
because until the first three are done we would be designing against a system none of us
has seen working properly.

**And two questions I can't answer for you.** Does the editorial pass get to change live
copy without a human reading it first? Our system's current answer is a deliberate no,
written into the machinery on purpose, so changing it is your call and not mine. And when
we ship the voice change, do we make seven separate edits or build the one shared place
they should all read from? Seven is quicker today; they have already drifted apart once
without anyone meaning it, and they will again.

---

**2026-08-12, evening — you made three decisions, and one of them has a catch I want you
to see now rather than in a fortnight.**

You chose to build the shared carrier for the voice change rather than making seven
separate edits. That is the more expensive route today and the right one: those seven
prompts have already drifted apart from a common ancestor without anyone deciding they
should, and doing it by hand a second time would guarantee a third. It goes through the
review council as one change, because it revises a default that governs every site we
will ever build.

You chose that an editorial pass does not get to change live copy without a human reading
it. That keeps the rule the system already has, and it makes the second stage a smaller
and cheaper thing to build than it looked this morning — it is no longer changing what the
machinery guarantees, just adding another thing that can raise a flag.

**Here is the catch, and it is a real one.** The queue those flags go into does not have a
working surface for a human to read them in. We have known this since July; it is written
up as one of the open bugs. You can see what it does in practice on the voice checker we
already run: thirty-four findings parked waiting for a human, exactly one ever closed, and
that one was closed by a machine re-checking the page rather than by anyone reading it. So
"queue it for a human" today means "park it where nobody looks."

That does not change your decision — it is the right call, and building an unsupervised
rewriter would have been the wrong one. It changes the **order**. The review surface stops
being a bug somebody should get to and becomes the thing that has to exist before the
second stage is worth building at all. I have written it into the plan as a blocking
dependency so the next person meets it in the plan rather than discovering it.

You also chose to look at the three lending sites that carry the same "authoritative"
setting we have just traced to the adversarial tone on the calculator site, and to leave
the other five alone. Agreed, and the reason is worth stating: the argument is that
"authoritative" goes wrong specifically where the reader is the weaker party — someone
asking to be lent money. That is not the situation on a gas-wholesale or an AI-consultancy
site, so changing those would be applying a theory well past the evidence for it.

**What happens next, in order.** The shared carrier first, since it is the only item that
reaches every page on every future site. Then two small things that are nearly free: a
one-field bug fix that makes copy findings visible as copy findings, and switching on the
claims checker for the calculator site — it is built, live and approved, and that site
simply never opted in, which is why nothing checked the new figure in yesterday's rewrite
except a query I wrote by hand. Then read the three lending sites. Then the review
surface. Then, and only then, the second stage.
