# Where we are — bug 168, the asset path helper

Plain prose, append-only, newest at the bottom.

---

## 2026-08-02, morning

I picked up bug 168 off the open pile. It had been filed on 31 July by the lane that fixed
128, and nobody had touched it since. I checked that twice — once with the ownership script,
and once by reading the live transcripts of the 27 other Claude sessions currently working
this repo, because the script only knows about work that has already been committed and half
these sessions are mid-fix. None of them was near this one.

The bug is about a small function with a big job. When the platform generates an image — a
hero, a logo, an icon, a social card — it commits the file into the site's git repo and then
some page has to point at it. The function `DeployedWebPath` is what everything asks for that
path. Six different places call it.

The complaint in the bug file is that this function gets one case wrong: the social card. It
answers `og_card.png`, with an underscore, when the actual file on disk is `og-card.png`,
with a hyphen. That is true. But when I went to fix it I found the filed explanation of *why*
was wrong, and wrong in a way that mattered — one of the four suggested fixes would have made
things worse rather than better. That is the first thing worth saying.

**What was actually going on.** There are two different bits of code that publish these image
files, and they name files differently. The main one derives the name from the image's
purpose. The other one — the bit that makes the favicon and the social card — just writes
`favicon.png` and `og-card.png` as fixed names, because those are what browsers and Facebook
expect. `DeployedWebPath` only ever knew about the first one. And here is the thing: from
what a caller passes in, there is no way to tell which of the two published any given image.
So every single place that used this function had to separately remember "…except for the
favicon and the social card". Some did. One nearly didn't, and if it had shipped it would
have reported a broken social card and a broken favicon on every site we run.

The suggested fix I nearly took would have swapped underscores for hyphens everywhere. That
sounds harmless. It isn't: the main publisher *also* uses underscores in that situation, so
the two agree today, and the "fix" would have pulled them apart. I only caught it by going
and reading the publisher instead of trusting the comment on the function, which claims to
mirror it. A comment claiming two things match is not the same as two things matching.

**What I did instead.** I made it one function. Where before there were two copies of the
naming rule — one in the publisher, one in the lookup — kept in step by a comment saying they
matched, now both call the same code. They can't drift, because there's nothing to drift
from. And the special case for the favicon and the social card lives inside that one
function, so none of the six callers has to remember it any more. The one that had been
carefully remembering it can now stop.

**On being wrong, deliberately checked.** Before I asserted any of this I put it through the
diagnosis loop, which is the thing that reads the real code and the real database and tells
you whether your theory holds. It came back **REFUTED**. That is worth explaining rather than
hiding, because it was useful in both directions. It was right about the thing that matters
for how alarmed anyone should be: the code that writes the social card tag into the page
doesn't use this function at all, it writes the correct name directly — so nothing is broken
on any live site today, and this is a trap waiting to be sprung rather than a fire. That
matches what the bug file itself said, and it's why I've described the change as removing a
hazard rather than fixing an outage.

It was also wrong about something, and I've written that down rather than just accepting the
verdict. It claimed the function had only one caller. It has six. Its conclusion rested on
having missed five of them, including the one place where the problem genuinely bites. So I
took the useful half and recorded the rest as an error of the tool's, with the evidence.

**Where it stands.** The code is written, builds clean, and all the tests pass. I don't
trust "the tests pass" on its own, so I broke the code three separate ways on purpose to
check that each new test actually notices — it did, all three times, and then I put it back.
It's gone to the review council, which takes about half an hour to come back. It's committed,
because on this repo holding code back isn't actually available: everyone shares one branch,
and the next person who builds ships whatever is there.

The one thing I want to be plain about: this is **not live yet**. The change is Go code, and
Go code does nothing until someone builds a new image and rolls it out. Until that happens
the old behaviour is still what's running. So the bug ticket stays open — the house rule is
that a bug is only closed when the fix is actually running in production, not when it's
written — and I'd rather leave it honestly open than tidy it away into the closed pile.

## 2026-08-02, midday — what the review council said

The council came back **REVISE**, which means "not yet, fix these first". Twelve reviewers
looked at it; eight approved, four objected. That is a good outcome to get, and I want to
explain why rather than treat it as a setback.

**One of them found a real bug in my fix.** The reviewer that reads for code quality noticed
that my new function took the stored path for the favicon and the social card, threw away
everything except the filename, and then rebuilt the path by sticking it under the images
folder. That works for the two we have, because both live there. It would have been silently
wrong for the first one that didn't — and it gave the obvious example: lots of sites put
their favicon at the very top level, `/favicon.ico`, not in an images folder. My own tests
couldn't have caught it, because they only had the two existing cases to test with. I've
fixed it so the stored path is taken whole, and added a test that uses a top-level favicon
specifically. Then I broke the code again on purpose to check the new test actually notices.
It does.

**The other three objections were not about the code — they were about what I showed them.**
This is the part worth remembering. One reviewer said I hadn't checked whether the running
system actually had my change in it. I had — I just hadn't put it in the submission. Another
said there's a known-traps entry about this exact function that I hadn't accounted for. I had
account for it; I'd *rewritten* that entry as part of this work. And the reviewer whose job is
architecture said that although I'd correctly declared this was a big-picture change and
brought it to the right place, declaring it isn't the same as filing the document that records
it — so I've now written that document.

So three of the four objections were "you did the right thing and didn't show me", and those
were graded *more* serious than the one that was an actual defect. That is the right way
round, I think, and it's a lesson I've written into the notes: the council can only review
what you hand it. Evidence you have but don't cite is evidence you don't have.

**Two things I found by answering their questions**, which is the real value of being asked.
First, one reviewer asked whether a future caller could route the social card through the
deploying code and stamp on the real file. Before my change that would have written a file
nobody looks at — harmless litter. After my change it writes to the *real* path, so it would
overwrite the actual social card. Nothing does this today and I measured that, but the nature
of the failure has changed, and that deserves its own ticket rather than a paragraph in mine.
I've filed it as bug 179, along with a second escape hatch in the same code that lets a caller
put a file anywhere it likes with nothing able to predict where.

Second, and I nearly missed it: one of my changes has left a helper function with **no callers
at all**. It's harmless, but a function nobody calls looks exactly like a job someone finished,
and the next person to read it will assume it's load-bearing. I've said so out loud in the
resubmission rather than let someone find it later.

**Still not live.** I checked the running system directly, three different ways to be sure the
check itself was working, and confirmed the change is not in it — the current build predates my
commit. So the ticket stays open. It will go live the next time anyone builds and rolls out the
system, which happens several times a day here, and the verification commands are written down
ready for whoever does it.

## 2026-08-02, early afternoon — the council caught a real one, and I was wrong

The second review round came back **REVISE** again, and this time it was not about
presentation. Two reviewers, independently, said the same thing: my change made it possible
for one part of the system to overwrite a site's real social card and favicon, and the
protection against that was sitting in a *ticket* rather than in the code. I had told them
twice, with measurements attached, that this could not actually happen.

**They were right and I was wrong**, and the way I was wrong is worth writing down because it
looked exactly like being right.

I had measured two things. First, that the check which creates this kind of work item no
longer creates them for the favicon and social card — true, fixed back in July. Second, that
every part of the system that *reads* an image path only ever deals with heroes, icons and
illustrations — also true, and I'd been careful to confirm the query wasn't returning nothing
by accident. Both measurements were sound. **Neither of them could answer the question.** The
first is about work items created from now on. The second is about readers, and the risk was
in a writer.

The thing that answers it is the queue of work already sitting there. One query: **eleven
items** are queued right now asking to deploy the favicon or the social card, and two of them
are in a state that will get picked up and run. They've been there since mid-July, from
*before* the check was fixed. Under the old code they were harmless — they'd have written a
file nobody looks at. Under my change they would have replaced the real thing.

Fixing a check stops new bad items being created. It does nothing about the ones already in
the queue, and nothing here goes back and tidies those up. I'd leaned on "that was fixed in
July" without noticing that the fix was only ever true *of the code that was running in July*
— and I'd just changed the code.

What caught it, in the end, was going to prove the reviewers wrong. I thought they were being
cautious about something I'd already settled, and ran the query to show it. It came back with
eleven rows.

**So I've built the guard rather than filing it.** The deploying code now refuses to touch the
favicon or the social card, and says why — and it refuses *before* it downloads anything or
commits anything, which matters, because a check that runs after the file has already been
written isn't a check. I tested that by deliberately moving the guard to later in the function
to confirm the test notices. It does.

One piece of genuine luck I want to be honest about: none of this is live yet, so the risky
change and its guard will go out together in the same build. If my earlier work had already
been rolled out, this would have been a live incident rather than a review comment.

There was a second, smaller error the same hour. Checking whether a particular override
setting is ever used, I got nine hits — which appeared to contradict something I'd already
told the reviewers twice. All nine turned out to be **my own review submissions**: the system
stores the text of what you submit, and I'd written about that setting at length. The real
answer is zero. The more thoroughly I'd argued the point, the more evidence I appeared to
generate against myself.

Round three is with the reviewers now. Both mistakes are written up in the shared log of wrong
calls, with the cheap check that would have caught each.
