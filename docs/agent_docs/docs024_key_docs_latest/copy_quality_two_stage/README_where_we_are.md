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

---

**2026-08-14 — the house voice you approved is live everywhere, and here is what that
actually means.**

You read both candidate texts, asked to see real copy written under the new one, and
approved v2. It is now the voice every writer in the fleet carries — the seven writing
agents and the blog writer — and it lives in exactly one place, so the next time you
correct it, the correction is one edit and reaches everything at once. That was the point
of the whole exercise: the last correction you made took three days to reach nothing,
because it had nowhere to land.

The way it shipped is worth one paragraph. First the seven writers had their private
copies of the old voice deleted and replaced with a reference to the shared one — a change
that altered nothing about what they wrote, which we proved before going further. Then the
shared text was swapped for v2 in a single edit. Then, rather than trusting the
configuration, we ran a real writer at a real page with everything locked and read the
prompt it actually received: the new voice present, once; the old rule gone. Nothing on
any live page changed during any of this, and every prior text is backed up and restorable
with one command.

Two things to know for what comes next. The sample run that convinced you also showed the
writer inventing two links it was told not to add — the fourth time we have watched set
discipline fail by instruction — so the mechanical link gate stays the only real answer
there, exactly as the plan already says. And the new voice is the *default*: any site's
own voice specification still wins where they disagree, so the mortgage calculator site
you refined by hand keeps behaving exactly as you left it.

---

**2026-08-15 — the four decisions on your desk, in plain terms.**

One: the loan calculator site's stored statement of why it exists ends by promising to
reveal "what lenders have no incentive to volunteer". That is the same move we corrected
on the mortgage site, one size smaller, and it sits upstream of every future page. I'd
restate it as what the site explains, calmly, rather than what lenders withhold. One field.

Two: the same site's settings hold up "reveal the true cost of credit" as a model
sentence — a family you rejected a week ago. Since writers copy examples far more
faithfully than they follow rules, a rejected sentence listed as a model is an
instruction to write it again. I'd replace it with a lever-shaped equivalent rather than
just delete it, because the example slot is doing real teaching work.

Three: the claims register I drafted for the calculator site includes its flagship
£5,000–£7,000 figure, but the only source for that number is the site itself, which makes
the register vouch for its own author. You can send someone to find an external source,
register it honestly as the site's own estimate, or open the register instead with the
stamp-duty facts already verified against GOV.UK this week — which would also catch the
live wrong stamp-duty figure that site is currently serving. I'd start with the
stamp-duty set and do the sourcing work later. Nothing gets applied until the other
thread's page rework finishes either way.

Four: you told me the review-surface bug is being worked in a separate thread. The
editorial pass was queued behind it because its output needs a human to read it. The
question is whether we build the editorial pass now, in parallel, so the two land
together — its first job is the six missing homepage links you ruled should wait for it,
which you would review personally regardless. I'd build in parallel; if the other thread
stalls, we pause exactly where we would have been anyway.

---

**2026-08-15, evening — you answered all four questions, and the two quick ones are
already done.**

You ruled: restate the loan calculator site's value proposition, replace its model
sentence, open the claims register with the stamp-duty facts (noting the wrong-figure
bug is being fixed in a separate thread), and build the editorial pass now rather than
waiting.

The two one-field fixes are applied and checked. The site's stored reason-for-existing
no longer ends by promising to reveal what lenders keep quiet — it now says the site
explains the rights and mechanics that decide what you pay. And the model sentence
writers copy is no longer "reveal the true cost of credit" but "show what a loan costs
in total, and which parts of that cost you can change" — same information, no
accusation, and it hands the reader the lever.

One thing we found on the way in: that site's settings had been regenerated this very
morning by the rebuild running over there, and the two bad clauses came through the
regeneration word-for-word. So the corrections were applied to today's settings, not
stale ones — and I've left a note in the rebuild team's folder warning that if they
regenerate again, our corrections would be silently overwritten and would need
re-applying. The old settings are kept, so undoing or re-applying either change is one
step.

Next: the claims register (checking first what the stamp-duty thread is doing so we
don't trip over each other), then the editorial pass build.

---

**2026-08-15, late evening — the claims register is on for the second calculator site,
and the flood we feared measures zero.**

The register is live: the thirteen stamp-duty facts, each one carrying its GOV.UK
citation, copied exactly from the mortgage site's register — which the machinery had
itself re-verified against GOV.UK this very morning. Your two unsourced borrowing-power
figures are left out, as you decided; finding them a proper source is still on the list.

Two things came out of doing it carefully. First, the stamp-duty bug you mentioned is
in better shape than our notes said: the wrong figure was actually fixed and confirmed
gone from the live pages six days ago, so the register isn't catching a live error —
it's insurance against the number ever creeping back, which one of that bug's own
warnings says is possible.

Second, we caught ourselves in an error worth owning. Last night's prep work claimed the
number-checking part of the scanner "cannot be tested without going live", and sized the
risk around that. It can be — it always could — and the test we thought proved otherwise
had simply been given a sentence the scanner isn't designed to flag. Tested properly
tonight: across every checkable piece of the site, the new register flags nothing, and a
deliberately-planted fake figure IS flagged. So the worry that switching this on would
bury the review queue in false alarms is settled by measurement: it won't. The mistaken
claim is corrected where it was written, with the lesson logged.

Next: the editorial pass build, in parallel with the review-surface thread, as you ruled.

## 2026-08-17 — the editorial pass exists now, and it passed the test we set it in June's terms

The second stage is built. That was the last thing on the list, and it works.

To recap what it is, because the name doesn't say much: the writer that builds a page is
judged on whether the facts are right, the sections are all there, and the links and
styling are correct. Nobody has ever judged it on whether the result reads like a person
wrote it, whether the most useful thing is at the top, or whether the page talks about
itself instead of talking to the reader. We had something that noticed those problems and
nothing that fixed them — design problems had a fixer, copy problems didn't. This is the
missing fixer.

**It turned out to need no new code at all.** Before building anything I went looking for
what already existed, and nearly all of it did: a way to read a page, a way to apply an
edit to one section safely, the house voice we shipped last week, and a way to put work in
front of a human before it lands. The only genuinely missing piece — letting the editor see
a whole page at once instead of one section at a time — could be done with a database query
in configuration. So it went live the moment it was applied. No build, no wait for a
release.

**The test.** You ruled on the 12th that the six guide links missing from the loan-and-
mortgage homepage should stay there, unrepaired, as this thing's proof. We kept them. Today
the editor read that page and, without being told what was wrong, reported that six
required guides were missing from the list, named all six, and then said everything else on
the page — order, naming, tone — was sound and didn't need touching. Its proposed fix is
six added lines and nothing else. Not one sentence of the existing copy was rewritten.

That restraint is the part I'd flag. The failure we kept expecting was an editor that
"improves" a page nobody asked it to improve and quietly loses things on the way — which is
precisely how those six links went missing in the first place. This one changed only what
was broken.

**Nothing has been applied to the site.** By your decision in June, this thing is not
allowed to change live copy without a person reading it first, and it is built so it
physically cannot: there is no step in it that can write to a page. Its proposal is sitting
in the review queue waiting for you. **That's the one thing I need from you** — a yes, and
I'll apply it and confirm the links are live.

I also built the checker that grades any proposal before it is applied: does it still carry
every link, has it lost any styling, has it invented a figure that wasn't there, has it
quietly shrunk the page. It reports a pass only after failing on purpose first — I feed it
five kinds of deliberate damage and it has to catch all five, or I don't trust its green
light. It does.

**One correction to last night's handover.** It said the webdesign.co.uk copy fix was
blocked, waiting on the duplicated-section bug. It wasn't — both halves had already been
fixed three hours before that was written. Worth understanding *why* the copy got better,
because it's a genuinely good sign: nobody rewrote it. Another thread fixed something
unrelated about the homepage's positioning, which regenerated the page, and because the new
house voice went fleet-wide last week, the regenerated copy simply came out without the
"a workbench, not a sales pitch" phrasing you objected to. The improvement arrived on its
own, in work aimed at something else entirely. That is what shipping the voice centrally
was supposed to buy, and it's the first time we've seen it happen unattended.

## 2026-08-17 (evening) — applied, and the six links are live

You said apply it, so I did. The six guides that have had no homepage link since the 12th
now have one, on the live site.

What ran: the editor's proposal went through the checker one more time against the page as
it stood tonight — not against the copy it was written from, which is the mistake that
would have quietly undone somebody else's work if the page had changed in between — and
then through the normal section-editing route that writes one component and redeploys it.
It took about forty seconds.

What I checked afterwards, because "the job says it finished" is not the same as "the page
is right": the six links are on the served page; the headline is untouched; the twelve
calculator cards, the two calculator grids and every styling attribute are exactly as they
were; nothing new was invented; and the page is forty words longer, which is the six lines
that were added. The formal test you set this thing on the 12th — a script that fails if
any required link is missing — now passes, and I've kept its verdict alongside the failing
one from the 12th so the pair is on the record.

One thing worth telling you because it nearly bit me. The checker has a shortcut for
grading a proposal straight out of the review queue, and I had written the handover telling
the next person to use it. It didn't work — I had guessed the shape of the thing it reads,
and a real proposal is shaped differently. It would have failed for anyone who tried it,
including me tomorrow. I only found it because I used it rather than trusting it. Fixed,
and the fix is committed separately so the reason is visible.

Where that leaves us: the editorial pass is built, it has done one real job on a live page
correctly, and the route from "proposal" to "live" is proven rather than assumed. What is
still not proven is the automatic version — today approval means a person runs two
commands, because the review queue still has no screen anyone can use. That is the other
thread's work and I have not touched it.

Next, unless you'd rather something else: run it on a harder page — one with several
sections and a fault that's a matter of tone rather than six obviously missing links — to
find out whether tonight's restraint is the design working or just an easy target.
