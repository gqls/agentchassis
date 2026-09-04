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

## 2026-08-18 — the harder page broke it, which is what a harder page is for

I ran the editor at a genuinely difficult page — ai-agent-orchestration.com's homepage,
eight sections, and a fault spread thinly through all of them rather than one obvious
hole. It failed. That is the useful outcome, and it corrects something I told you last
night.

Last night the editor changed six lines and left everything else alone, and I said the
restraint was striking. **It doesn't generalise.** Faced with a page where something was
slightly wrong everywhere, it tried to rewrite the whole thing at once, produced more than
the system would accept, and was cut off mid-answer. The platform handled that correctly —
it threw the half-answer away rather than saving it, which is exactly the failure we've
been bitten by before — but the lesson is that yesterday's good behaviour was a property of
an easy target, not of the design.

The fix isn't "let it write more". It's a limit: at most three edits per run, chosen by
what a reader actually loses. A bounded job can't overrun; a bigger allowance just moves
the wall further out.

With that in place it ran again and did something I'd call genuinely useful. It reported
that the same sales pitch is restated almost word for word in **five** different sections
of that page, and that one thing on the site is called four different names — "adoption
register", "adoption tracker", "Enterprise AI Agent Adoption Tracker", "adoption tracker
page". Neither of those is visible to whatever writes a single section, because each
section is written without sight of the others. That is the whole reason this stage reads
the entire page. Its three proposed edits **delete** things: a fake "feature" that was
really a sales pitch, an off-topic tangent, and one inconsistent name.

Then my own checker failed those edits — wrongly, twice, and both worth telling you about
because they're the same mistake in different clothes. It couldn't tell "deliberately cut
repeated material" from "gutted a section", so it rejected the editor for doing half its
job. And it was quietly skipping list-type content while reporting that it had checked
everything — technically true, thoroughly misleading. I fixed both, and made a point of
proving the checker can still catch real damage afterwards, because "it complained, so I
changed it" is how a check quietly becomes decorative.

**One thing needs you.** Those three edits are sitting unapplied, and they're a different
kind of decision from last night's: last night added six links, tonight *removes* live
copy. I'd rather you looked at that class of change before I apply it, which is precisely
what the review step is for.

One incidental finding worth knowing: the homepage component I fixed for you last night was
rebuilt by a routine process this morning, and the six links survived it intact. Good news
— but it also means anything we leave sitting in the queue can be pointing at a version of
a page that no longer exists by the time someone approves it.

## 2026-08-18 (evening) — applied, and I want to be straight about what it did and didn't do

The three edits are live on ai-agent-orchestration.com. The homepage no longer carries a
"feature" that was really a sales pitch, the departments section no longer detours into a
hundred words about industry-wide project failures, and the thing that was being called
four different names is now called one name.

What I checked afterwards: every number on the page is still there (sixteen of them), every
link is still there (thirty-seven), nothing new was invented, and the five sections nobody
asked it to touch are untouched. The page is about 250 words shorter and says its main point
once instead of five times.

**Now the part I'd rather report than have you find later.** This morning I measured the
specific fault I thought that page had — a mannerism where everything is defined by what it
isn't ("in days, not months", "rather than a survey result"). Nineteen instances. After the
edits: fifteen. Barely moved.

The editor didn't go after them. It read the page and decided the bigger problem was that
the same pitch appears in five different sections, and spent its whole allowance there. I
can't yet tell you whether it was right to make that call or whether it simply ran out of
room — three edits was the limit I imposed this morning after it tried to rewrite everything
at once. Either way, the honest summary is: it fixed repetition and naming, and left the
mannerism mostly intact.

I'm flagging that because the tidy version of this story — "measured a problem, ran the tool,
problem went down" — would be technically true and would misrepresent what happened. What the
tool actually found was something I hadn't measured and wouldn't have thought to look for:
that a reader going down that page meets the same argument five times.

One more thing worth knowing for the future: when I first checked the live page after the
edits, it still showed the old text, and I nearly wrote down that the naming fix had failed.
It hadn't — the site simply hadn't finished publishing. Checking a page seconds after a
deploy tests the publishing, not the edit.

---

**2026-08-19, evening.** I set out to build one small thing and found something bigger on the
way. Both are worth your time, and the second one is not what I was looking for.

**The small thing, first.** Yesterday I told you the fault was in the brief rather than in the
writer, and then had to correct my own numbers within the hour because I had counted the whole
brief document when the writer only ever reads part of it. The fix for that was a proper tool,
and it now exists. Before measuring anything it asks the writing agent itself which parts of a
site's brief it actually reads — rather than me assuming — and it turns out to be five fields.
It then measures only those. It also has a way of settling the question that matters: whether a
phrase in the brief genuinely ends up in the copy. You give it a phrase and it tells you how
many times that phrase went in and how many times it came back out.

That test is what makes the rest trustworthy, because it can say no. The tagline on the AI
orchestration site went into 1,369 prompts and came back in 409 pieces of copy — supplied by the
brief and reused, exactly as I said yesterday. But two other phrases I might have blamed went
into **zero** prompts and still appeared in 35 and 21 pieces of copy. Those are the model's own
words, not the brief's. Without that check I would have blamed the brief for both.

One thing did sharpen. That site's brief does not merely happen to contain the mannerism — it
**instructs** the writer to put that exact tagline in the homepage hero, the services hero, the
footer and every page description. It is the only instruction of its kind anywhere in the
estate. You objected to a hero; the brief ordered that hero.

**Now the bigger thing.** While reading the code to write the tool, I found that the brief the
writer reads is, on some sites, a fragment of the brief we wrote.

A brief is a document of about twenty parts — voice, things to avoid, example phrases, heading
style. The writer doesn't read that document. It reads a single readable summary of it, which is
rebuilt every time anyone edits the brief. The rebuild uses **only the part being edited**. So
if someone updates two sections, the summary becomes those two sections, and the other eighteen
silently stop reaching the writer. They stay in the document, so everything looks complete to
anyone who checks.

This actually happened, on the 18th of April, to four sites. A planning step wrote a small
update nine minutes after the full brief was created, and the writer's copy of the brief shrank
from about nine thousand characters to about three thousand. One of the four was repaired by
accident later. **The other three have been writing pages against a fragment for four months** —
including the AI orchestration site you complained about. Its list of things to avoid — don't
say "seamless", no urgency language, no AI hype vocabulary — has not reached a single page since
April.

**I want to be careful here, because there is a tidy story available and it is wrong.** This is
not the cause of the copy you objected to. The parts that went missing never mentioned the
mannerism, and one of the missing parts — the example phrases — is itself written in it, so
restoring it carelessly would make that particular fault worse, not better. Two separate
problems that happen to live in the same place. I have written that down in the bug file so the
next person can't quietly merge them.

**What I have not done:** I have not changed the code, because it is shared machinery that every
site's brief passes through, and that goes through review rather than round me. I have not
restored the three shrunken briefs, because restoring ten thousand characters of instruction
changes what every future page on those sites says, and someone should read that diff rather
than run a sweep. And I have not touched anyone's brief, because the trap here is nasty: a
careful, narrow correction to a brief is exactly the thing that destroys the rest of it. That
warning is now filed where a session will meet it before they act, not after.

**One admission.** My own tool lost data on its first full run — three sites vanished from a
twenty-five-site report and a fourth was truncated, because of how I was splitting the database
output. Nothing errored. I only caught it because one site's number had changed since a run ten
minutes earlier. That is the same lesson as the deploy-checking one further up this page: the
absence of an error message is not evidence that anything worked.

---

**2026-08-20.** I fixed the code defect I described yesterday, and I need a decision from you about
one consequence of it. The decision is at the bottom; the rest is what happened.

**The fix.** Yesterday's finding was that the brief our writer reads is rebuilt every time anyone
edits a site's brief, and rebuilt from **only the part being edited** — so eighteen sections quietly
stop reaching the writer while the stored document still looks complete. Nobody had picked it up
overnight, so I did it. Two lines of substance: build the brief from the merged document rather than
from the edit, and put its sections in a fixed order so that two identical briefs actually read
identically. It is committed. It does nothing until the next fleet release, which looks imminent —
someone has already set the next version number.

**The test was the interesting part, and it is worth a paragraph.** There were no tests at all on
either function. And the obvious test would have been useless: the code that formats the brief was
never wrong — hand it a document and it renders that document faithfully. The bug was in **which
document it was handed**. So the test drives the real write path end to end and checks what actually
gets stored. I then proved the test works by putting the old code back and watching it fail — and
that step caught me being sloppy twice, which I have written up: once my "old code" wasn't the old
code, and once my check printed nothing at all, which looks exactly the same as a check that doesn't
work. I nearly recorded both as proof.

**The review found things, and I was glad of it.** I put the change through the reviewer council. It
came back "revise" twice. The best objection was that the fix rests on an assumption I had asserted
rather than tested — that when the brief is rebuilt from the merged document, the *previous* version
of the brief (which is sitting in that same document) doesn't get swallowed into the new one. That
would have been worse than the bug: every edit nesting the whole previous brief inside the new one,
for ever. It doesn't happen, and it never did — the code has always skipped that field. But nothing
tested it, which was exactly the reviewer's point. It does now.

Another reviewer told me my instructions for checking whether the fix had actually shipped were
pointing at the wrong machines. They were right, and worse than they thought: seventy-five machines
run this program and the command I published looks at two of them. I have replaced it.

**Now the decision, and it is a real trade-off rather than a formality.**

Fixing the bug means the missing brief sections come back — which is the point. But on the
AI-orchestration site, one of the sections coming back is a list of example sentences for the writer
to imitate, and those examples are written in **exactly the mannerism you objected to last week**
("Agents fail in isolation — not in cascades", "Speed comes from engineering discipline, not from
skipping the hard parts"). So the moment anyone next edits that site's brief, the writer will be
handed those examples as models to follow.

One of the council's reviewers blocked the change over this, twice, and its argument is one you have
made yourself: telling other teams about a hazard in a document is not a control, because they may
not read it before they act. It wanted a per-site switch, defaulting to off.

I did not add the switch, and here is the honest cost of both choices:

- **Ship it as written.** The briefs repair themselves the next time anyone touches them. Risk: on
  that one site, bad examples reach the writer with no human in the loop.
- **Add the switch, defaulting to off.** Risk: the bug stays live on all twenty-five sites until
  somebody remembers to turn each one on — we would be keeping a fault as a safety feature, and this
  platform has been bitten before by switches nobody ever flips.

**My recommendation is to ship it, and fix that one list of examples first** — it is content on
another team's site, so it is their edit to make; I have told them the release is coming and offered
to make whatever change they specify. I have not written replacement examples myself, because your
ruling is that the framework writes the copy, not me.

I also went and measured the thing my own argument was resting on, rather than leaving it as a
guess: across every site whose examples do reach the writer, only three of fifty-two turn up
word-for-word in published copy, eighteen times in total — while the one *mandated* tagline turns up
four hundred and nine times. So examples do get copied, but nowhere near as much as an instruction
that says "use this sentence". That narrows the risk. It does not remove it, and I should say plainly
that my sample was the first sixty alphabetically rather than a proper random one.

---

**2026-08-21.** I ran the editorial pass a fourth time on the AI-orchestration homepage, to answer a
question we had left open: when I capped it at three changes per run, was that cap stopping it from
finding anything else, or just stopping it from biting off too much at once?

**It was the second, and that is the good answer.** The fourth run picked **three completely
different sections** from the third run — no overlap at all. So the cap bounds how much it does in one
go, not how much it can see. Run it again and it works down the list.

It also did something I want to point out, because it is the behaviour I was hoping for and could not
guarantee. It said, in its own words, that the real problem on that page is bigger than three edits
can fix: the "features" and "differentiators" sections are near-duplicates of each other with
identical headings, and the same story about pilots stalling in production appears **four separate
times**. Its conclusion was that this "should be revisited as a structural merge" — in other words, it
declined to pretend three edits had fixed a structural problem. That is the difference between a tool
that reports and a tool that just does something.

**All three changes passed the mechanical checks.** Nothing invented, no numbers changed, no links
dropped, no markup lost, the list field still a list with the same number of items. Two of the three
make the copy noticeably shorter — 37% and 31% — and in both cases the checker confirmed every fact
and link that disappeared is still present elsewhere on the page, so it is removing repetition rather
than removing content. Both are flagged for a human to read the prose, which is right: a machine can
prove nothing was lost, but not that what remains reads well.

**This is waiting for you, and it is a different kind of approval from last time.** The first time you
approved one of these it was six missing links being put back — a pure addition, nothing to lose. This
one **deletes live copy**. The largest change cuts the homepage call-to-action from 126 words to 79:

> *"We run over a thousand orchestrations a day across 14 live sites on Kubernetes, Kafka, and
> Postgres, and the Enterprise AI Agent Adoption Tracker records the same story across financial
> services, healthcare, and logistics: pilots clear the first hurdle easily, then stall once real
> production message volume hits a pipeline with no blast-radius containment…"*

Its reason: the original ran the same statistic past the reader twice.

So the question for you is not "is this safe" — the checks answer that — but **"do you want the
editor deleting live copy on your say-so, and is this the standard you want it held to?"** Nothing has
been applied and nothing will be until you say so.

One small thing worth knowing, because it nearly became a false alarm on my side: right after the run
finished I checked the log of AI calls and found nothing there, which looks exactly like our
instrumentation having broken. It hadn't — that log simply lands a few minutes late. I checked whether
the log was receiving anything at all before concluding anything, and it was.

---

**2026-08-23.** The three edits you approved are applied and live on the AI-orchestration homepage.

**What changed.** The call-to-action is the one you'd notice: it was 733 characters and is now 496 —
the version that ran the same statistic past the reader twice is gone. The news subheadline went from
340 to 255, and the differentiators list lost about 250 characters without losing any of its seven
items. I checked the actual served page afterwards, not just our database: the old phrasing is gone
from it and the new copy is there.

**It failed the first time, and both reasons are worth two minutes of your time**, because they are
the sort of thing that will bite whoever does this next.

The first was my own mistake wearing a convincing disguise. When I sent the instruction I gave it a
readable name for tracing — "cli-copyedit-apply". It turns out that field is not a label at all: the
system uses it as a **database schema name**, dropped straight into a query without quoting. The
hyphens made the query invalid, and what came back was a database syntax error deep in the code that
creates agents. **It looked exactly like a platform fault and it was entirely my doing.** The fleet
uses plain names like "demo_client"; I now do too, and the trap is written down where the next person
will meet it before they hit it.

The second is more interesting because it is a property of the system rather than a slip. That page is
automatically rebuilt every day, and rebuilding it **replaces** its components rather than updating
them — new internal identities, same content. So the addresses I had written down when you approved
the edits were dead by the next afternoon. This is the third time this has caught this lane. The fix
is to stop writing addresses down: look the component up by *which slot on which page* at the moment
of sending, which both the checker and the sending script now do.

Worth saying plainly: the content itself was **identical** across both rebuilds, so your approval
still applied to exactly what you approved. If the rebuild had changed the copy, the checker would
have failed the edits rather than shipping something you had not seen — that is what re-checking
immediately before sending is for.

**One admission.** Two days ago I wrote down a lesson for myself: when watching for a job to finish,
match it by its own identifier, never by "the most recent one". Today, on this task, I wrote a check
that took the most recent one — and it told me an edit had finished when it hadn't. I caught it only
because I also looked at whether the text had actually changed, and it hadn't. Writing the lesson down
did not stop me repeating it, which is why the check now lives inside the script rather than in a
document I have to remember to reread.

---

**2026-08-25.** Four things happened today, and one of them is yours to decide.

**The plumbing you approved is now fully on.** The change that sends tone findings to the copy editor
instead of rebuilding the page went live this morning, and by the afternoon every machine in the
fleet was running it. I proved it by asking the running program what it contains rather than trusting
the deployment paperwork. No finding has arrived through the new route yet — tone complaints really
are as rare as the safety argument said, roughly one a week — so the first automatic run is still
ahead of us, and I'll verify it end to end when it comes.

**The checker had a fault of its own, found by our first outside user, and it is fixed.** Another
team ran the editor on their page, and our checker wrongly complained on three of their four changes
that a required link was missing — it was demanding every edited field contain the link, when a
heading can never contain one. Worse, the one edit it complained about loudest was actually *adding*
a link their page had been missing all along. That kind of false alarm is dangerous beyond the
annoyance: a check that cries wolf teaches people to scroll past it, and sitting right next to the
false alarm was a real problem (their edit deletes an entire numbered list). The checker now asks the
right question — is the link still somewhere on the page after the edit — and I proved the fix both
ways: it stays quiet on their heading, and it still shouts if a link genuinely vanishes from the
whole page. I've told that team the false alarm is withdrawn and the real problem stands.

**The proposal waiting for your approval has gone stale, and I'd ask you not to approve two-thirds of
it.** The three edits on the AI-orchestration homepage were written against last Thursday's version of
the page, and that page rebuilds itself daily. Re-checked today: two of the three edits would now
write numbers onto the page that no longer match what the page says — the checker catches exactly
this, which is why re-checking before acting is the rule. The third edit (the call-to-action cut you
saw quoted last time) still checks out clean. So: **approve the call-to-action cut if you want it;
the other two need to be re-proposed against the current page** rather than patched.

**And the big one: your verdict on the finetuning pages reached this lane.** You said the pages fail
the "would a person actually say this" test and that this machinery needs to substantially improve —
that landed here in writing this morning, with the specimens. The uncomfortable, useful fact in it:
the other team's own checklist scored the worst section as clean hours before you rejected it. Our
pattern lists catch the tics we've already named; your ear catches a register — the methodical
scaffolding, the performed candour — that no list of banned phrases will ever pin down. So the honest
position is that "substantially improve" is not another rule added to a checklist; it needs a
different kind of acceptance test, closer to your sentence than to our patterns. **That's the
direction question I need from you before this lane starts building:** whether you want us to pursue
this — a rewrite capability judged by something nearer taste than rules — as the lane's next major
piece of work, ahead of everything else on the list.

---

**2026-08-25, later.** Your homegarden review reached this lane the same afternoon, and I did what it
asked, in the order it asked.

**The deep re-read is done.** Everything this project has learned about copy since the 12th is now
assembled in one document with every claim pointing at its evidence:
`REFRESH_2026-08-25_deep_context_the_accumulated_copy_discussion.md`. Reading it back, the story is
more coherent than it felt while it was happening. You have been making the same point since the
start — first "why is everything negative", then "this sounds like AI", now "the premise of the page
is wrong" — and each time we fixed the surface you named, the same instinct came back wearing new
grammar. The strongest thing the evidence says is about why: the writing machinery learns from what
our instructions *show* it far more than from what they *tell* it. The clearest single number: the
writer's own instructions demonstrate the "X rather than Y" habit sixteen times per call while
forbidding it. When we removed the demonstrations of one habit, that habit vanished; the habit whose
demonstrations we left in place got worse. Our checklists also have a measured ceiling — they scored
a section clean hours before you rejected it — so the standard for "fixed" has to be your sentence,
not our pattern counts. And your about-page point is a different kind of fault from all of the
above: not how the words sound, but that the page is about the wrong subject entirely. I have a
suspicion about where all that self-description comes from — the honesty rules we feed the writer
may be leaking into the pages as content — and it is written down as a suspicion with a test, not
as a finding.

**The prompt audit you asked for is scoped and counted.** "Every prompt in the database and code"
is, as of today: 173 prompts inside agent configurations, about 691,000 characters across 73 agent
types; 2,442 per-field writing instructions inside the component library; 31 live per-site briefs;
and prompt-building code in 26 Go files. The plan (`PLAN_2026-08-25_prompt_audit.md`) asks six
questions of each one, with your question — is it encouraging AI styles of writing — split the way
the evidence demands: judge what each prompt demonstrates first, weight by whether it actually
reaches a writer, and check whether its own prose is teaching the habits it bans. A cheap mechanical
pass orders the work; the judgment pass starts with the writer's own instruction block, which every
section of every page passes through.

**Per your instruction, no fixes have been proposed yet.** The refresh and the audit plan are the
deliverables today; fixing starts from them.

---

**2026-08-25, night.** The first pass of the prompt audit is done, and it found the thing you
pointed at.

**Where the about-page premise comes from.** Every time the writer produces a section, it reads a
rule we added on 26 July to stop a site claiming its content was infallible — a good rule with a
real cause. But the same rule goes on to tell the writer what to say *instead*: when a section is
about method, describe how the content is sourced, that we name sources and dates, and "say
plainly that we can still be wrong". Put that instruction in front of a writer building an "about"
page and you get fourteen headings about sourcing, dates, plainness and what the site won't do. The
words you objected to — "names the source behind any figure", "say plainly", "what this site will
not do" — are close to the instruction's own words. I'm calling this a strong lead rather than a
proven cause: the test is to build a method section with and without that clause and compare the
pages, and that comes next. The half of the rule that bans overclaiming stays; it's the "say this
instead" half that leaks.

**How much the machinery teaches the habits it bans.** Counted on what the writer actually reads —
the full assembled instruction of about 48,000 characters per section, three consecutive real calls
agreeing to within two — it demonstrates the "X, not Y / rather than" construction about
sixty-four times, uses "plainly" fourteen times and "honest" ten times. That comes from four layers
stacked: the writer's rule text, the house voice (which says "Say what a thing IS rather than what
it is not" — the rule against the habit, written in the habit), the site's own brief (every live
site's brief carries it, most more than ten times), and the per-field guidance in the component
library. Six and a half thousand of these calls a month. No single layer looks bad on its own;
together they are the loudest voice in the room.

**What I have not done:** proposed fixes. The reading order for the judgment pass is written down
— the writer's rule block first, then the house voice, then the site briefs, then the planners that
decide what an about page is for in the first place.

---

**2026-08-25, late.** I read the two instructions that reach every section of every page in full,
the way a person would, and wrote down what each is teaching.

**The writer's rule sheet** is right in everything it forbids and wrong in three things it says to
do instead. It tells the writer that a section about method should describe how sources are named
and dated (twice, in two places). It bans the word "honest" — your ruling in July, correctly — but
then tells the writer to show honesty by "naming the limit, the failure mode, or what the thing
cannot do" and to say "we cannot tell you X", which is where every "What this site will not do"
heading comes from. And where a page has a testimonials or case-study slot and nothing real to put
in it, it tells the writer to fill it with "statements in the company's own voice about their
values, approach, or commitment" — which is the "talking about ourselves" you objected to on the
very first day. The bans stay; the "instead" clauses are what I'd change, and the change is written
down as a shape, not applied, with the test that would prove it before anyone calls it fixed.

**The house voice** — the one you approved on the 14th — says the right things and says most of
them in the habit it forbids: "Say what a thing IS rather than what it is not" is the rule against
defining-by-negation, written as a defining-by-negation. Seventeen of those in six thousand
characters, the densest instruction in the fleet. The fix would change nothing about what it asks
for, only how its own sentences are built, and since it's your text I'm putting that to you rather
than doing it.

**Nothing has been changed yet.** Next reads, in order: the per-site briefs, then the planners that
decide what an about page is for.

---

**2026-08-25, the test you asked for.** We replayed the real instruction that built homegarden's
about section, twelve times in four variants, touching nothing live. Verbatim replay reproduces
the page you rejected — same score, same headings — so the test is sound. Removing the writer's
"say this instead" rules helped somewhat (10 down to 6 on our count) but "What this site will not
do" still appeared every time. The collapse came from somewhere else: the page had been PLANNED
with the title "About Home Garden — Editorial Approach and What We Will Not Do", and when we
replaced just that title with "About Home Garden", the score fell to 1, the "will not do" heading
vanished in every run, and the copy opened with "Home Garden exists to answer one question: is
this job urgent, or can it wait until a better month?" — unprompted. Your instinct was exactly
right: the premise was wrong before the writer ever started, and it was wrong in the PLAN. So the
first fix is to the planner, the second (smaller, still real) is the writer's "instead" rules, and
the wording I had drafted to replace them earned nothing in the test — plain deletion did as well.
The sentence-level habits ("X, not Y", "rather than") did not move in any variant — they are a
separate fault, fed by how many times our own instructions demonstrate those shapes, and they need
their own fix. The decisions you're being asked to make, and the places where your July rulings
and your August verdicts pull against each other, are in my reply in chat — in short: which halves
of the July "honesty" rules survive, whether the house voice may be rewritten in its own
recommended shape, whether empty testimonial slots should be omitted rather than filled with
philosophy, and whether "acknowledging what we don't know" survives anywhere as a trust device.
Nothing has been changed live.

---

**2026-08-25, your six answers, acted on.** Three of them are live tonight, with escape hatches.
The writer's "say this instead" rules are deleted — every ban stays, and the calculator warning
stays as you ruled. The house voice now practises what it preaches: same fifteen rules, same
meanings, rewritten so its own sentences state the positive first (it was demonstrating the
banned move seventeen times; now zero). And the planner will stop planning testimonial sections
on sites that have nothing real to put in them — we found its own worked example had been
showing it a testimonials slot, which is how the slot spread. Each change has a backup and a
one-file rollback, and the next new site's about page is the natural check.

Two answers came back from the research you ordered. The "best in class" mission is still there,
but in exactly one place — the instruction that classifies a new domain at birth ("The bar is not
'competent template' but 'stands comparison with the strongest sites in this vertical'"). No
site's specs carry it, the planner and writer never see it, and the note that would have wired
research to it says "separate TODO", never built. If you want it to mean something, it needs to
travel: into the strategy each site keeps, into the planner's brief, and into a research step
that actually goes looking for the latest findings and trusted reviews. That's a piece of work to
scope, not a line to add.

And the about-page premise: the planner never invented "What We Will Not Do" — it condensed it
from the mission the site was seeded with, which is written largely as prohibitions. Six of our
twenty-three about pages have titles like that. The fix I'd propose is one demonstrated good
example in the planner (your register: "We're hoping you can get a lot of useful tips from this
site…") plus a rule that mission constraints become internal rules, never page titles. Say the
word and it ships the same way tonight's three did.

---

**2026-08-26.** Your two answers from last night are done.

**The planner fix is live — and testing it first paid for itself.** The wording I had drafted
held in only one of two trial runs; the failing run titled the about page "Practical UK Guidance,
No Products to Sell" — the exact mistake the rule describes, made with the rule in front of it.
Adding one hard formatting line ("title the about page 'About' plus the site name and nothing
more: no subtitle, no dash, no qualifier") made it hold three out of three. That's a pattern now
confirmed twice at two layers: rules about register bend, rules about format hold. The tested
version is what shipped, with a one-file rollback.

**Your Go question, in one paragraph.** Putting the best-in-class text into a Go file wouldn't
break anything — it would just make the wrong thing better. Go text is compiled in: it can't
drift, but every wording change needs a build and a fleet roll, and this mission is text you'll
want to tune. The platform already solved this shape with the house voice: the plumbing lives in
Go (one injection point, so it can't scatter into drifting copies), the words live in one database
row (live within a minute, reviewed through migrations). The plan proposes the same: a second
carrier row for the build standard, injected into the planner and designers, plus a per-site
"benchmark" written at birth naming the strongest sites in that vertical, and then the research
wiring — inventory what the classifier already gathers, add a periodic refresh of latest findings
and trusted reviews, and route it so the planner and experience loop can ask "the best sites have
X; do we?". The finance-versus-grass-seed split rides the same plan.

**One more thing a reviewer's comment surfaced.** Four of our agent types quietly carry two active
configuration rows each; a config change keyed on the agent's name can hit both or the wrong one
without any error. None of last night's changes were affected — I checked each target — but it's
now written into the traps file with the check every future prompt migration must carry.

Still queued: executing the best-in-class plan, deleting the writer's now-dead testimonial-filler
rules, and retitling the six existing about pages that carry the old self-limiting premise — those
are per-site jobs, not prompt rules.

---

**2026-08-26, close of day.** The fresh build you deployed carries both of this lane's changes and
I checked at the running binary, not the tag: your truncation trial is live (the old instruction
is gone from the binary, the new one present), and the copy-editor routing safeguard is live with
it. The council approved the trial unanimously.

Six of the nine pages have now rebuilt. The two you asked about most — about and approach — read
much better in kind (the self-congratulating candour is gone) but still carry about ten
comparison constructions each, and the arithmetic says why: your earlier ruling forgives two mild
"rather than"s, the machinery applies that allowance per SECTION rather than per page, and two of
the shapes you named ("instead of", "not just") are ones the gate was never taught to see. So the
two questions now waiting on you, with numbers attached: should the trial repair every "rather
than" (that one change accounts for six of the ten), and I'll teach the gate "instead of" and
"not just" as soon as you confirm how strictly to treat them. The homepage and FAQ — the pages
you were actually reading when you called regression — are requested for rebuild and still queued.

---

**2026-08-26, the two decisions explained** (also given in chat, kept here so it survives the
session). The gate is the checker that runs after the writer and before a page is saved; it
recognises five named comparison shapes, repairs the sharp ones, and under the August ruling
lets a couple of mild "rather than"s stand. Decision one: that allowance turns out to be applied
per SECTION, not per page, so a seven-section page ships six "rather than"s without a repair ever
being attempted — the choice is repair them all under the trial (the truncation makes that cheap:
the sentence just ends earlier), keep an allowance but count it per page, or leave it be.
Decision two: "instead of" and "not just" are not among the five shapes the gate knows, so it
cannot see them at all — you named "instead of" in the trial, so this is confirmation plus a
strictness call, and if decision one is "repair them all", the strictness question answers
itself and both land as one change.

---

**2026-08-31, your reading of the Fable samples.** Both points taken, and they sharpen the whole
programme. The model that stopped making comparisons still writes like a model in a different
way: it compresses. "Judge the work on results" is four thoughts in seven words, and you're right
that this is how models are — so expansion becomes deliberate work we do every time, and it's now
a scored part of every model trial (Fable's verdict is downgraded to "register pass, density
fail"; Grok and Gemini get judged on both from the start). And "No vendor pays us" taught the
subtler lesson: a true fact, stated strongly, isn't yet a benefit — its relevance needed
unfolding, and it likely isn't the first thing on a visitor's mind anyway. The ordering question
— what a visitor's first doubt actually is, and the hidden questions behind it — has gone to the
offer and benefit thread as you asked. One honest note against ourselves: our read had scored
that sample "passes decisively"; your ear extended the criteria, which is the same lesson the
checklists taught — our instruments order the work, and you remain the acceptance test.

---

**2026-08-31, late evening (session note).** Three things from tonight, in plain terms. First:
you were right that Grok is set up to fetch news every day — and it has never once delivered an
article. Every call since it went live on the 30th has been refused by x.ai because the account
is out of credits, and the pipeline was built to shrug and report an empty day rather than an
error, so nobody could tell a broken feed from a quiet one. The news pages stayed full because
the ordinary RSS feeds kept working. It's written up as bug 418; topping up the xAI account is
yours to do and un-blocks both the news and the Grok writing trial. Second: your go-ahead on the
producer gate went to the offer team within the hour, and the review council pushed us — rightly
— to stop deferring the database safety check we'd learned from Friday's race; it is now written
into the next wash migration rather than promised for later. Third: the new build you deployed
carries both rule changes and the truncation trial — confirmed by reading the running programme
itself — so the next page rebuild is the real test of whether the competitive-comparison habit
is finally gone. The Gemini trial ran: it wrote shorter but slipped the same negative framing in
by the back door and invented two benefits we never claimed, so it does not replace what we
have. Decisions still with you: the question-hierarchy build, the hero-ranking axis, the xAI
top-up, and four smaller ones listed at the end of the new handoff.

**2026-08-31, afternoon.** Three things since the morning. First: the council approved the
migration-668 review on the third round — no serious objections left, and the two small ones
turned out to be about an earlier draft, not the file we actually committed. That review trail
is closed. Second: we finally have a clean before-and-after on the copy fixes shipped so far.
The approach page on finetuning.uk was rebuilt through the corrected templates on the 26th, and
counting the verbal tics you rejected: the old page had 27 of them, the rebuilt one has 9. The
"X, not Y" pattern — the worst offender, 15 on the old page — is completely gone. But "rather
than" actually went UP, from 4 to 6, and we can now prove it isn't coming from any instruction
we wrote: every document that used to teach it has been cleaned. The model just likes saying it.
That is precisely what the newest fix (live since midday today) is built to catch and repair.
Third: to measure that newest fix, we've sent the approach page round for one more rebuild —
this time through the full corrected machinery. If it works, the "rather than" sentences should
simply end earlier, with the comparison dropped. If the repairs mangle the meaning instead,
that's a finding too, and we'll report it either way rather than quietly keep it.

**2026-08-31, late afternoon.** The rebuild came back, and the newest fix works. The approach
page went from 27 of your rejected verbal tics (before any of this work) to 9 (after the
template clean-up) to 5 now — and reading rather than counting, the repaired sentences do
exactly what we hoped: they simply end where the comparison used to start, and the meaning
survives. "We work across several AI models rather than standardising on one, because…" became
"We work across several AI models, because…" — nothing lost. Twice the repair machinery judged
its own proposed fix would cut too much or stay too negative, refused it, and kept the original
— the right call both times, and it wrote down why. Of what remains, two tics sit in card
HEADINGS, and that turned out to be a genuine bug worth finding: the scanner has a list of
field types it treats as "never sentences", and card headings are stored in a field called
"name", which is on that list — so headings were invisible to it from birth. Filed with a
one-line fix identified (bug 420). The page still reads methodical in structure — that was
expected, nothing we shipped targets it, and your ear stays the judge of whether it now passes.

**2026-08-31, early evening.** Two more things done. The farmer insurance site's planning
documents no longer recommend building the seven tools you deleted — that was the layer that
would have quietly re-created them the next time the system generated content for that site.
The other lane that ran your deletion checked the work independently and signed it off; what
remains on farmer is the button labels and a handful of sentences that still mention the dead
tools, queued next. And since the approach-page test passed, the other eight finetuning pages
are now being rebuilt through the same corrected machinery — before-and-after counts will be
ready for each, and the pages themselves will be ready for your read.

**2026-08-31, end of day.** The nine-page rebuild finished, and the honest summary is: it works,
with two exceptions that are now named bugs rather than mysteries. Six pages we can compare
dropped from 30 of your rejected tics to 6, and of those six survivors, most are either quiz
answers where the contrast is the actual content, or cases where the repair machinery correctly
refused to cut a sentence that would have lost its meaning. The contact page took four goes —
not because the rebuilds were bad, but because the site's own planning document stored the
contact address twice with a one-letter difference, and the page builder reads the copy nobody
checks. That's fixed at the source, the wrong address is off the live site, and only one site in
the whole fleet had this disease. The two exceptions: the about and services pages can't be
rebuilt at all right now, because the repairs shorten their heroes so much that a separate
safety guard (which stops pages losing half their text) refuses to save — two protections
fighting, and the loser is the page. Bug filed with the fix shape. Separately: the council
approved the best-in-class propagation build on the second round, so that ships with the next
release, and the wording of the standard travels from one editable place exactly like the house
voice does.

**2026-09-02.** The offer-analysis thread audited the whole fleet and found something
uncomfortable but useful: the copy-repair machinery deliberately leaves alone anything a
site's own brief supplied — your rule, and the right one — but 32 of 34 sites have at least
one of the banned constructions sitting IN their briefs, so a dirty brief effectively smuggles
the register onto live pages with a licence. That raises a decision for you (it's number 11 on
the list): clean the briefs site by site the way we cleaned finetuning's, make the repair
machinery stop honouring phrases from dirty briefs (which changes what your
"a site's voice outranks the rules" principle means in practice), or leave it and work through
the nightly findings. I lean to the first — it fixes the source without touching your rule.
Meanwhile the cheap half was built and shipped the same day: the nightly brief checker now
watches the two WORDS you named first — "plainly" and "honest" — which, it turns out, no
automated check anywhere had ever looked for. Its first run promptly caught two sites quoting
them in their briefs. Also tidied: our daily health line was actually two different numbers
wearing one name; the reports now say both.

**2026-09-02, close of day.** All ten of your decisions from this afternoon are done or moving.
Done today: the farmer about-page fix is applied; the review batch is on the admin page waiting
for you (fifteen farmer items, one loanzy item — one farmer item carries a note where the
proposal deletes a dead button rather than renaming it, your call); the "AI slop" wording is
washed out of finetuning's planning document and the page rebuild is running; the banned-words
register is at version 2 (your "plain words" family, plus a dash-shaped loophole the offer
thread caught in live copy the same day); the negation checker is under council review; and the
best-in-class standard now actually reaches the planners and designer — the machinery rolled
out this afternoon and the next site-planning run is the proof. The loanzy edit you asked
evidence for: both claims checked out against the sources (National Debtline is genuinely free,
run by the Money Advice Trust; the credit-agreement right is real, section 77), and the wording
was tightened where the law reference was attached to the wrong claim — the same disease as the
MaPS error, one size smaller. Still yours when you have a minute: the batch review, the xAI
top-up at console.x.ai, and one word to the offer thread confirming the question-hierarchy
build — they'd rather hear it from you directly than through a relay, which I think is the
right instinct.

---

**2026-09-04 (early afternoon) — we finally have a Grok result, and it is the best answer yet to the
copy problem.**

You asked what happened when we compared Grok. Until today the honest answer was "nothing": the
trial was blocked since 31 August because the xAI account had no credit, so every call was refused.
You funded it yesterday afternoon — the news feed started delivering at 15:06 and has brought in 45
stories since, having delivered exactly zero before that. So this morning the comparison could
finally run.

It ran on the same piece of writing every other model has been tested on: the "Our Approach" page
section from finetuning.uk, the worst-scoring page of the canary, using the exact prompt the live
system sent on 26 August. Nothing about the live site was touched.

The result. The fault we are chasing is the habit of saying what something is NOT in order to say
what it is. On that section the shipped copy commits it eight times. Grok's newest model committed
it **zero times, twice running** — and not by being vague, which is how a model usually gets a zero.
It kept every fact it was given, wrote longer than our current model rather than shorter, and made
its points as statements: "a model aimed at a muddled process will still produce a muddle, and you
will have paid for it." That length matters, because it is the thing you rejected Fable for — Fable
also scored zero but compressed everything into riddles.

Grok's cheap model was the opposite: it reproduced our current model's tic word for word, invented a
swipe at ChatGPT that was nowhere in the brief, and returned a third of the words. It is not a
candidate.

I also checked the obvious cheaper alternative before recommending a change of model: whether our
current model simply needs to think harder. It does not. With thinking off it scores 9 and 6; with
thinking on at the normal setting, 9 and 7; the live page, 8. Turned up to maximum it spent its
entire budget thinking and returned no text at all, four times. So there is no setting that fixes
this — the difference really is the model.

**What I need from you is a read, not a decision on my numbers.** Our count is a machine measure and
your ear has overruled it twice already. The four Grok samples and the four control samples are
written out in plain prose in
`docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/AUDIT_prompts/TRIAL_OUTPUTS_2026-09-04_grok_arms_verbatim.md`.
Two things to weigh besides the writing: Grok's best model takes four to five minutes per section
against our thirty-eight seconds, and it costs about ten to fourteen cents a section. And one
honest caveat — in one sample it wrote that a model you own "can be explained to an insurer, an
auditor, or a client who issues a questionnaire". That is a reasonable thing to say and it was not
in the brief, which is the same class of fault, milder, that ruled Gemini out.

Two of my own mistakes today, both caught and written down: I claimed our writer does not think,
from a database column that turns out to be incapable of recording it; and I blamed a broken
command-line tool for a batch of errors that were actually the Anthropic account running out of
credit for half an hour this morning. Neither reached you or another team, and both are logged with
the check that would have caught them sooner.
