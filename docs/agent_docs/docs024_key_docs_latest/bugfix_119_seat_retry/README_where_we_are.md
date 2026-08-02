# Where we are — the council seat that loses a whole round

*(Plain-prose log, append-only, newest at the bottom.)*

---

**2026-08-01, morning.**

I picked up bug 119. The short version of it: when we send a change to the review
council, thirteen or so reviewers each write their opinion as a small structured
document. If **one** of them writes that document badly enough that our code cannot read
it, we throw away the *entire* round — every other reviewer's completed work included —
and the submitter has to start again. It costs their credits and about half an hour, and
the verdict that comes back says nothing at all about the change they proposed.

Before doing anything I checked whether this was still real, because the bug was filed
five days ago. It is real and it is worse than filed. Across every council round we have
ever run — 424 of them — **23 were decided by "we couldn't read one reviewer" rather than
by anyone's judgement**, and 15 of those were in the last week alone. It has happened to
seven different reviewers, so it isn't one badly-written reviewer prompt; it's something
about the whole arrangement. The worst single case was on 31 July, where one submission
burned **three rounds in ten minutes** — same reviewer failing the same way each time,
because nothing anywhere tries again.

**Then the measurement changed my mind about what to build.** The bug says the cause is a
reviewer writing a document that's complete but has a bracket in the wrong place. I went
looking for live examples so I could fix exactly that — and there are none. Not "a few":
zero. Of the 39 cases I could find, the evidence for 36 had already been deleted by
routine cleanup, and all 3 survivors were a different failure: the reviewer ran out of
room mid-thought and produced nothing at all.

That mattered, because I was about ten minutes from building something that would have
fired zero times and reporting it as a fix for fifteen lost rounds a week. What is
actually and permanently true is the *other* half of the bug: **nothing ever asks the
reviewer to try again.** So that's what I built against.

**While reading the code I found something I wasn't looking for, and it's the better
half of the fix.** Each step in our system can declare what it needs back — "I need
JSON", say. It turns out the code reads a setting called `output_type`, and almost
nothing in the fleet writes that. What the fleet actually writes is `output_format` — a
near-identical name — and **nothing anywhere reads it**. The numbers: 6 steps use the
name the code understands; 100 use the name it ignores. Ninety of those hundred are
asking for JSON, across 32 different agents, and that includes every single council
reviewer.

The practical effect is small to describe and large in aggregate: those 90 steps were
never given the instruction sheet for producing JSON. And the very first line of that
instruction sheet is *"Ensure valid JSON syntax (proper quotes, commas, brackets)"* —
which is, word for word, the thing the bug says the reviewers keep getting wrong. I want
to be careful here, because I can't prove that instruction would have prevented the
failure; nobody can prove that. What I can say plainly is that we were asking these
agents for something and never telling them how, and now we do.

**So the fix is two things.** Make the setting the fleet actually writes mean something
(prevention). And when a step that asked for JSON gets back something unreadable, ask it
**once** more — not with the identical question, which we know just reproduces the same
failure, but with a short corrective note that depends on what went wrong: "you ran out
of room, same judgement but shorter" or "that wasn't valid JSON, same answer as one valid
document". If the second attempt fails too, everything behaves exactly as it does today.

I checked what that costs before building it rather than guessing: across 785 of these
steps in the recent record, this would have triggered **twice**. It only ever runs on a
path that was already producing something unusable.

One deliberate non-decision worth flagging: my first instinct for the "ran out of room"
case was to give the reviewer a bigger budget on the retry. I didn't, because there's a
note in our own code from an earlier round of this exact problem saying, in effect, don't
— whatever the limit is, the reviewer with the most to say will reach it, so raising it
buys a little time and no cure. Asking for the same judgement more briefly is the honest
fix.

It's committed and it's had tests written against it that I verified by deliberately
breaking the fix to check the tests actually notice. It's now in front of the review
council itself, which is a pleasing sort of circularity. **It is not live yet** — Go
changes only take effect when a new image is built and rolled — so the bug stays open
until I can prove it's running in the actual pod rather than just in git.

---

**2026-08-01, later. The council said no, and it was right.**

The review came back REVISE, and the objection was a good one: I had claimed the fix
reaches 90 steps across 32 agents, and four of the reviewers pointed out — correctly —
that I'd never actually checked the setting sits somewhere the code can see it. They cited
a known trap here: in these agent definitions, a step's prompt and its token limit live at
*different* depths, so counting a setting in one place tells you nothing about whether the
code reads it there.

I hadn't checked. I'd measured the count and assumed the location. So I went and measured
the location, and two things came out of it.

The good news: the assumption held. Every one of the declarations sits exactly where the
code looks. The fix does reach what I said it reaches.

The correction: **my count was one short**, because my query only looked at top-level
steps and missed steps that live *inside a loop*. There's one of those — in the page
content writer, the step that actually writes the words on our pages. So it's 91 steps
across 33 agents, not 90 and 32. Small in itself, but the shape of the error is the
interesting part: a query that walks the obvious level and silently skips the nested one
would miss a lot more on a different agent.

A second reviewer made a sharper point that I'd rather have thought of myself. My retry
makes the failure *rarer*; it doesn't change what happens when it still fails. The step
goes on quietly returning prose to something that asked for structured data, and reporting
success. Reducing how often a thing silently goes wrong is not the same as fixing that it
can. So I've added a small flag on that outcome — when a step asked for JSON and, even
after the retry, didn't get any, that's now recorded on the result. I deliberately stopped
short of making it a hard failure: that would turn ninety-odd steps that currently limp
along into steps that break outright, over content they didn't write, and we've been
bitten by exactly that before. The aim was to end the silence, not to start an outage.

**And the review handed me a gift.** That very round had one of its own reviewers fail in
exactly the way this bug describes — and because I went and grabbed the evidence
immediately, before our cleanup deleted it, I now have a fourth example. All four are the
same thing: the reviewer ran out of room and produced nothing, not the bracket-in-the-
wrong-place the bug was originally filed about.

It also showed me something I'd have missed otherwise. Someone on another thread has been
raising the limit for reviewers as they fail — four of them are now on double the budget.
The other thirteen are still on the original. And this round's failure landed on one of
the thirteen. That's the whole argument for my approach in one picture: raising the ceiling
moves the problem to whoever hasn't had theirs raised yet. Asking the reviewer to be
briefer is the thing that doesn't just relocate.

Resubmitted. The code is committed either way — on this shared setup you can't hold work
back pending a review — and it's still not live until someone builds and rolls a new image.

---

**2026-08-02, morning. It's live, and the bug is closed.**

Someone rolled a new chassis this morning — version 1228 — and it carries the fix. I checked
that properly rather than taking the version number's word for it: I went into both running
copies of the service and looked for text that only exists because of this change. It's there
in both, five different markers. One of them is worth mentioning, because it's the one that
proves the *timing*: the phrase "Keep every citation" only entered the code in my very last
commit, the one that answered the reviewers. Finding it in the running program means the
image isn't just newer than my first commit — it's newer than all of them.

**I caught myself measuring the wrong thing twice this morning, and the second one nearly went
into the record.**

The first was small and embarrassing. My own instructions for checking this had told the next
person to look for a line the change *deleted*, on the theory that if the old line is gone,
the new code must be there. But the line I'd picked was a code comment, and comments don't
survive into the finished program at all. That check would have come back "not found" no
matter what — including if nothing had shipped — and it would have looked like a pass. I've
rewritten it.

The second one matters more. I wanted to show, with numbers, that the problem I fixed was
real: that steps asking for JSON were never actually being told how to produce it. We keep a
log of the prompts we send, so I counted how many of those prompts contained the missing
instruction. The answer came back 31 out of 9,063 — a small trickle, which seemed about right.

It was completely wrong, and the giveaway was the *shape* rather than the number. All 31 were
from a single day, and they came in **exactly two per reviewer**. Real traffic never lands
that evenly; two-per-reviewer is two review rounds. They were mine. When you send a change to
the review council, everything you wrote explaining it gets pasted into each reviewer's
prompt — and I had quoted the missing instruction word for word while explaining that it was
missing. So my own bug report was showing up in the log as evidence the bug wasn't there.

I re-did it looking for a heading that no explanation would ever quote in passing, and the
real answer is starker than the one I nearly published: **9,061 of 9,063** — over four months
and twenty-five different agents — got an instruction sheet that says nothing about JSON at
all. And the control makes it plainer still: the setting the old code *could* read has not
been used by a single call in the entire four months. So this wasn't a trickle with some gaps.
It was never happening.

There's a pattern here I'd flag to anyone working this way. This is the second time in two
days on this one bug that **writing something down changed what I then measured**: first when
the landmine note I wrote about this change came back as six reviewers objecting to it, and
now when my own explanation of the bug polluted the log I used to confirm the bug. Writing it
down is still right — I'm not arguing for less documentation. But you can't then measure with
a yardstick made of your own words.

**One honest gap.** I can't yet show the fix *working* in production, because there's almost
nothing running. Four calls in the half hour after the roll, and none of them the kind this
change touches. No review council has convened since my own. So the retry has fired zero
times — out of roughly zero opportunities, which tells you nothing, and I've written it down
as telling you nothing rather than dressing it up. The queries to re-check are in the closed
file for whoever is next past when the fleet is busy.

I also want to be straight about the closing decision. My own handoff note had said the real
test was seeing the "wasted review rounds" number come down. That number depends on other
people submitting reviews, not on this fix — I could sit on this for days and not learn
anything. The project's actual bar is "fixed and live", and that's met and proven. So I've
closed it on that, and left the other measurement written up as owed rather than pretending
it's done or quietly deleting the promise.
