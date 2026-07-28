# Where we are — the retries that were never really retries

Append-only, newest at the bottom. Plain prose.

---

## 2026-07-28 evening — what this is

I went looking through `bugs_open/` for something nobody else was working on, and
picked **129**. On the face of it, it is a narrow complaint: when we start a
helper agent in its own pod and send it a job, sometimes the helper just sits
there. It comes up healthy, it receives the message, it writes "completed
successfully" in its log, and it does nothing at all. Six minutes later the
agent that asked for the work gives up.

The bug file blamed the helper, and suggested three ways to make the helper
behave better. That turned out to be the wrong end of it entirely. **The helper
was doing exactly what it was told.** It was being handed somebody else's name
badge.

## What was actually going wrong

When one agent asks another to do something, it waits for an answer, and if no
answer comes it tries again. That "try again" is the part that was broken.

You would expect a retry to be *the same request, sent a second time*. It wasn't.
Instead, the waiting agent built a **brand new message from scratch**, out of
what it knew about itself. So the second message went out saying, in effect, "I
am the agent that is waiting" — the waiter's own identity, not the worker's. It
also went out **with the job description removed** (replaced by a note saying
"this is a retry") and with the wrong instruction on the envelope.

So the helper received a message that named the waiting agent. It looked up that
name, found a job that was already sitting there waiting for an answer, and quite
reasonably concluded there was nothing for it to do. It said so in its log as a
success. Meanwhile the thing it was actually supposed to do had never been
described to it at all.

## How widespread — this is the part that surprised me

I expected this to be an occasional thing. I asked the database instead of
guessing, and the answer was much starker.

**In the last fourteen days, every single retry the whole system sent — 430 of
them — went out this broken way. Not most. All of them.** And 294 of those, about
seven in ten, went on to burn through all three attempts and fail.

There is a second number that says the same thing from another angle. If retries
worked, you would expect most problems to clear up on the first retry, fewer on
the second, fewer still on the third — a tail that thins out. What we actually
have is 93 requests that needed one retry, 45 that needed two, and **294 stuck at
three**. It piles up at the end instead of thinning out. That is the shape of a
retry that never rescues anything.

I want to be careful about one thing, because it would be easy to overclaim: a
hundred-odd of those requests did eventually succeed after a retry was sent, and
I **cannot tell** from the records whether the retry rescued them or whether the
original answer simply turned up late. So the honest sentence is "seven in ten
retried requests exhausted their attempts", and not "retries never work".

There was also a clue sitting in the code the whole time. There is a special case
in there for one family of tasks, with a comment explaining that those ones need
the full job description, so they get handled differently. Somebody noticed the
missing payload years ago, carved out an exception for their own case, and moved
on. Everything else kept getting the broken version — and that exception has not
actually been used once in the last fortnight.

## What I changed

One rule: **a retry is the original request sent again.** We now keep a copy of
the message we sent, and when we need to retry, we send that copy — same
identity, same instruction, same job description. The only things allowed to
differ are the attempt number and a timestamp.

Two supporting changes, both about failing loudly instead of quietly:

- If for any reason we do *not* have a copy of what we sent, we now **refuse to
  retry** rather than making something up. That turns six minutes of silence into
  an immediate, named error somebody can find.
- On the receiving side, if a worker is ever handed a job that names the waiting
  agent again, it now says so as an error instead of reporting success. That one
  is a belt-and-braces check — the cause is fixed upstream — but it means this
  particular failure can never again be *invisible*.

I checked, rather than assumed, that only two kinds of request in the entire live
system can reach this path, so wiring those two covers everything.

> **CORRECTED 2026-07-28, later — "covers everything" is not true; it covers 422
> of the 428 retries we actually sent in a fortnight.** See the entry at the
> bottom of this file for the six that are left and why they are a different
> problem rather than a missed one.

## Where it stands tonight

The code is written, tested and committed. The database change it needs is
applied. The new image is built and pushed but **deliberately not yet rolled
out**, because releasing a new image kills any review that is running at the
time, and this change is in front of the review council right now. Once the
verdict comes back it goes out, and then I want to see it work on a real failure
— not a happy path, which would only prove the image shipped.

## A note on how I got there, since it is the useful bit

I got this wrong twice before I got it right, both times the same way: I matched
on what something *looked like* and stopped looking.

First I started fixing the helper, because the bug report pointed there. Then,
having worked out it was the retry's fault, I found a piece of retry code with
exactly the defect I expected and wrote it up as the answer — except that code
has not run in months. The real one was in a different function, and it was worse:
as well as the wrong identity it was also dropping the job description, which I
would never have found if I had stopped at the first plausible hit.

What caught it both times was going back to the actual evidence — the exact error
text in the bug report — rather than to my idea of what the evidence meant.

## 2026-07-28, later — the review said no, and it said no to the right thing

The review council came back and **rejected** it. Worth being precise about what
they rejected, because it is not the fix.

Ten seats looked at it. **Six approved. Not one of them disagreed with the
diagnosis** — the guardian, the seat that actually blocked it, went out of its way
to say the analysis was right and that the homework had been done before asking.

What it objected to was **where this change was being made, not what it does**. My
change adds a new shared piece of plumbing and a new database column that several
other bits of code now read through. That is the sort of thing that is supposed to
go through an architecture review on its own, rather than arriving inside a bug
fix — and I had *said so myself in the submission*, which the guardian pointed out
does not help: saying a thing belongs somewhere else is not the same as taking it
there. There is a standing ruling from earlier today that says exactly this, from a
near-identical case, so it is a fair cop.

The awkward part is that the seats **disagree with each other about the remedy**.
The guardian's suggested safe alternative is to ship only the small child-side
guard — but two other seats approved the plan specifically *because* it treats
that guard as a safety net rather than as the fix. And on the merits the guardian's
alternative does not actually fix the bug: it turns a silent failure into a loud
one, which is genuinely worth having, but the work still does not get done and the
caller still times out. That combination — a blocking objection about process, and
the seats contradicting each other on what to do instead — is the case the ruling
says a person has to settle, not me.

**So I have stopped short of rolling it out.** Everything is ready: the code is
committed, the database change is applied (it is a new optional column, so it does
nothing at all until the new software is running), and the new build is made,
uploaded and checked. Turning it on is one command. Not turning it on costs
nothing and changes nothing.

There are three ways forward and I would pick the third:

1. **Roll it out now and send the design for review in parallel.** The bug is fixed
   everywhere immediately. But it does the exact thing the ruling was written to
   stop, two days after the ruling.
2. **Ship only the small guard the guardian suggested.** Cheap and safe, and it
   does not fix the bug — it makes the failure visible instead of silent.
3. **Leave it built but off, and put the design in front of a person.** Nothing
   can regress, and it goes live the moment someone says yes. The cost is that the
   defect keeps wasting roughly 430 failed retries a fortnight until then.

**I also got something wrong, and the review is what caught it.** I had claimed the
fix covered every case. It covers 422 of the 428 retries we actually sent in the
last fortnight. The remaining six come from two web-fetching steps that have a
*different* version of the same problem — they put the caller's own name on the
very first message, not just on the retry — so they need their own diagnosis rather
than being bundled in here. In practice those six now fail immediately rather than
retrying, and four of the six were failing anyway, so the real cost is about two
requests a fortnight losing a retry that might have worked. I would rather say that
plainly than round it up to "complete".

I got there by asking the database a question that was *next door* to the one I
needed answered: I asked which kinds of task exist in the configuration, when what
mattered was which kinds actually wait for a reply. Both are real queries, both
return real numbers, and only one of them was about my claim. Three separate
reviewers flagged the claim as unproven — none of them could run my query, which
is precisely why they noticed it was an assertion.

**So: the bug stays open.** Our own rule is that a bug closes when it is fixed
*and live*, and this is fixed and built but not live. It is one decision away.

## 2026-07-28, ~22:20 — it went live on its own, which turns out to be the interesting bit

I had decided not to switch it on, and to put the choice in front of you. **It got
switched on anyway, about an hour ago, and not by me.**

Another session rebuilt and rolled the system out for its own change at 21:48 our
time. Our build tool always builds from the last saved state of the shared code —
on purpose, so that nobody's half-finished work can accidentally ride along in a
release. My work was saved. So it went out with theirs. Nobody did anything wrong;
the session that pressed the button had no way of knowing what else it was carrying.

I checked this rather than assuming it: I looked inside the actual running programs
on both machines for a phrase my change **deletes** and three phrases it **adds**.
The deleted one is gone and all three new ones are there, on both. It is live.

**This is worth your attention beyond this one bug.** There is a standing rule from
earlier today that says a change to shared plumbing should not ship ahead of its
review. The rule quietly assumes the person who writes the change decides when it
ships. **On this codebase they don't.** Saved work goes out in whoever's release
comes next, and we release many times a day. So "wait for the review before shipping"
isn't a thing anyone can actually do — the only way to genuinely hold something back
is to ship it *switched off*, behind a toggle, and the rule doesn't currently ask for
that. That's a change to the rule, and it's your call, not mine.

**How it is behaving now that it is live.** The half I can check is working properly:
every request the system has sent since the rollout has correctly kept a copy of
itself, with the right recipient named on it — the exact thing that was wrong before.
The check for the specific broken state returns zero, as it must. And the copies are
about a kilobyte each, which answers a reviewer's worry that they might be huge.

**The half I cannot yet claim** is the retry itself, because nothing has needed
retrying since the rollout — the system has been quiet. I deliberately triggered a
job that used to fail two times in three, and it sailed through and did real work,
which is good news but is *not* proof: it succeeded first time, so it never needed
the retry path at all. I would rather say that plainly than round it up.

**So the bug stays open**, with exactly one thing owed: catch one real retry
happening and confirm it goes out addressed to the right recipient. I have written
down the two commands that check it, and what does and does not count as proof, in
the handoff.

---

## 2026-07-28, about half past ten at night — the last piece arrived, and the bug is closed

The thing I said was owed has happened, and it happened on its own about twenty
minutes after I wrote that note.

A real job — a price scraper — sent a request off and waited half an hour for an
answer that never came. That is exactly the situation this bug was about. The
system noticed the silence and sent the request again. Under the old code, that
second attempt was the moment everything went wrong: it would rebuild the request
from scratch and put the *wrong name* on the envelope, so the worker receiving it
looked up the sender's own job, saw it was already sitting there waiting, decided
there was nothing to do, and quietly said nothing. The job then waited until it ran
out of patience. That happened to every single retried request in the fortnight I
measured — four hundred and thirty of them — and two thirds of those jobs died of it.

This time the second attempt went out as an exact copy of the original, with the
right name on the envelope. The worker picked it up, did the work, and answered. The
job finished, and so did the parent job that was waiting on it. About a minute and
three quarters, start to finish.

That is the whole thing proven now. The half I could already show you — that the
system keeps a faithful copy of every request it sends — was working. The half I
could not show you, that the copy actually gets used properly when something needs
retrying, is now witnessed on real traffic rather than in a test I wrote myself.
That distinction mattered to me: a test I write can only prove the new code does
what I think it does, whereas this was the system's own work, unprompted.

**Two things I got wrong on the way, both caught quickly.**

The first is worth passing on because it is a trap anyone would fall into. My own
handoff note told me to check the logs of a service called the chassis. I did, and
found nothing, which looked like "no retry has happened yet". It had happened — in
a *different* service's logs. All these services run the same program; which one
does the retrying depends on which one was waiting. So the check I had written down
for whoever came next would have told them the wrong thing. I have fixed it in the
notes.

The second was simply reading too fast: I saw a request still marked "waiting" and
assumed it had failed, when in fact it was still in mid-air and finished
successfully seconds later. I had not looked at the clock.

**One small correction to something in the notes.** I had written that if anything
other than two specific web-scraping steps turned up without a saved copy of its
request, that was a real problem to chase. Four such things turned up. They are not
a problem: they are a different kind of request that gets re-run from the beginning
rather than re-sent, so it has no need of a saved copy. I checked this by reading
the code rather than assuming, and then re-ran the count in a way that tells the two
kinds apart. Genuine problems: none.

**What is still open, and it is all yours, not mine.** The review panel objected to
*how* this change reached production rather than to the change itself, and that
question is untouched by tonight's result — the fix works, the objection was about
process. There is also the awkward finding I raised last time: on this setup,
committing work *is* shipping it, so "hold it back pending a decision" was never
actually available to me. Both of those need a call from you. And the two
web-scraping steps I mentioned have a separate, smaller fault of their own that
deserves looking at properly rather than being bundled in here — bundling is
precisely what got objected to.
