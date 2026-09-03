# Where we are — the bug sweep lane

Plain prose, append-only, newest at the bottom. This is the running log for the owner: what was
found, what broke and why, what was decided, and what needs a choice. No field names or file
paths unless they genuinely help.

> Started late, on 2026-09-03. The lane has been running since 2026-08-26 and kept its technical
> log (`NOTES_bugsweep.md`) and its handoffs, but never this one. The earlier history is in those
> files and in the summaries; this picks up from today rather than pretending to reconstruct it.

---

## 2026-09-03, late morning — the meta description bug, and a message that had been lying for a fortnight

Some background, because the thing that was wrong is small and the reason it stayed wrong is not.

Every page on our sites carries a one-sentence description — the line a search engine prints
under the page title in its results. Most of them were written by an automated job that runs every
hour: it finds pages with no description, asks a model to write one, and saves it.

Before it saves, it checks the sentence against the site's own writing rules — the phrases that
site has banned, and the claims we are not allowed to make without evidence. If a sentence breaks
one of those rules, the job refuses it and writes nothing. That is correct: we would rather have
no description than a bad one.

The problem is what a refusal looks like from the outside. It looks like nothing at all. No error,
no alert, no record anyone reads. The job finishes, reports success, and the page stays blank —
and tries again an hour later, and refuses again.

There is exactly one place that tells a person how to read one of these runs: a summary line the
job prints when it finishes. It said that a page can be skipped for one of four reasons, and
listed them. There are seven. The three it left out are precisely the three that matter — the ones
where a rule refused the sentence and a human needs to look at it. Anyone reading that summary
would reasonably conclude that the writing rules never refuse anything.

**Why it went wrong is more interesting than the fact.** The four it listed and the three it
omitted are written in two different styles in the code. If you search for them the obvious way,
you find exactly the four already listed — and the search tells you the list is complete. It takes
two searches, and nobody had reason to think a second was needed. The list was correct the day it
was written, in August; the three new reasons were added a few days later by a separate piece of
work, and nobody thought to go back and update a sentence in a different file.

**What I did.** Rewrote that summary. It now names all seven, and — this is the part I think
matters more — it separates them into "no action needed" and "a person has to read this", and then
says plainly that the list is a copy, where the real list lives, and that finding it takes two
searches. Listing seven instead of four would have been correct today and wrong again the next
time someone adds an eighth. This is the third time in one piece of work that we have found a list
that was right when written and quietly went wrong by something being added to the world around
it, and we have no rule that catches that.

It is live. It changed a configuration value, not code, so there was nothing to rebuild or deploy.

**What I did NOT do, and it needs your decision.** This does not make a refusal *loud*. It stops
the one human-facing surface from actively misleading people, but a refused page still goes by in
silence. Making it loud means putting each refusal in front of a person, and the obvious place to
put it is a review queue we already have.

I checked that queue before proposing it. It holds sixty-six items waiting for a human, and five
have ever been dealt with, the most recent in late August. Sending refusals there would look like
a fix and would be a way of moving the silence somewhere quieter. So I have left that decision
with you rather than doing it — the question is not technical, it is "who reads this, and what
makes a new item different from the sixty-six already there".

**Two other things worth knowing.**

The first is a correction to our own record. When this bug was written up yesterday, I said the
system keeps no readable trace of a refusal — that the records expire too fast. That was wrong,
and I had drawn it from a query that could not have told me otherwise: I searched for refusals,
found none, and concluded the records were gone. They were not gone. There were no refusals. The
records survive for about a day and every one of them was there, intact, saying the job had
written a description successfully. The write-up had also turned that mistake into an instruction
telling the next person not to look in the one place the evidence actually is. Corrected, and
written into our shared list of traps, because the two explanations look identical in every query
that searches for the failure itself.

The second is something new. The instruction we give the model says, in so many words, that if a
page's content does not support a good description it should skip that page, and that returning
fewer than it was asked for is a correct answer. Nothing anywhere compares how many pages were
offered with how many came back. So there are *two* ways a page can be silently skipped, not one —
the rules refusing a sentence, and the model quietly declining to write one — and they leave the
same trace, which is none. The fix everyone was discussing would only have caught the first of
them. I have written that down as a fifth option: compare the two numbers. It is two integers that
are already sitting there.

**Is anything broken right now?** No. I checked, with a control to make sure the check could
actually see a problem if there were one. Thirty-seven pages have no description, and every one of
them is a near-empty page the job deliberately will not touch — which you have already ruled
should stay that way. Not one page is both worth describing and undescribed. So both of these
silent paths are real, and neither is costing us anything today. That is an argument about how
urgent this is, not about whether it is worth fixing.

---

## 2026-09-03, early afternoon — you said make them loud, so I did, and one query changed how

You read the earlier note and said: **"yes, make them loud."** That is built. Here is what it
actually does and the one thing I had to check first.

**The check first, because it changed the design.** I had told you the obvious place to send these
refusals — the review queue — holds sixty-six items and has cleared five. What I had not asked was
*why*. I had been quietly assuming it was about people being busy. It isn't.

Every work item on this platform either names an agent that will act on it, or names nobody and
waits for a human. I counted both, across the live table and the archive. Items that name an
agent: fifty-six thousand of them, and **83% get finished**. Items that name nobody: six and a
half thousand, and **17% get finished**. And the queue I nearly used holds sixty-nine items of
which **every single one names nobody**.

So the graveyard is not a fact about your team's attention. It is a fact about the shape of the
row. Sending refusals there would have looked exactly like a fix and been one more row in the 17%.

**What it does now.** When the writing rules refuse a description, the system files a job for a
small new agent whose only task is to try again — this time being *told what it got wrong last
time*. That sounds obvious and it is the whole point: the hourly job has been re-offering the same
page every hour with the same instructions and no idea why the last sentence was rejected. The
rewrite goes back through exactly the same rules, so nothing sneaks past them. If it is refused a
second time, only then does it go to a person — and it goes with the original sentence, the rule
that rejected it, and the failed rewrite attached, which is a far more useful thing to hand
someone than "this page has no description".

**What I deliberately left alone.** One of the seven refusal reasons means "the rules themselves
could not be loaded" — a plumbing fault, not a judgement about the writing. Sending that to a
rewriting agent would be asking the wrong thing to fix it: it would write a new sentence, the
rules still wouldn't load, and it would spin. That one stays quiet for now and I have written it
down as an unfinished edge rather than folding it in to look complete.

**One thing I got wrong, and it is worth telling you because it is the kind that hides.** I wrote
a test to prove that the *harmless* refusals stay silent — that we don't start filing jobs for
things nobody needs to look at. The test passed. I then deliberately broke the code to check the
test would notice, and **it still passed**. The test had been checking nothing at all: I had asked
it "did anything unexpected fail?" when nothing had been set up that *could* fail. My own comment
above it said it couldn't possibly pass by accident. It could. Rewritten so it now asserts a fact
that can come out either way, and it correctly fails when I break the code. Only running the
break-it check found this — reading the test never would have, and I had read it twice.

**And one boundary I nearly walked past.** While filing the paperwork I found a note from a
fortnight ago recording that you had explicitly *withheld* permission for automated jobs to
rewrite descriptions that already exist — you allowed it once, for a one-off cleanup, and no
standing mechanism was ever given that power. What I have just built is exactly the shape that
note warns about. It does **not** have that power: it can fill an empty description and nothing
else, and I verified that against the live configuration rather than trusting my own reading of
it. I have flagged it in two places, because turning that power on later would be a single
innocuous-looking line in a config file, and it should be your decision, not a config edit.

**Is any of this doing anything yet?** No, and honestly it can't be. There are currently no pages
that both deserve a description and lack one, so there is nothing to refuse. The code half also
only starts working after the next chassis deployment. I have written down the check to run before
anyone concludes it's broken — because "no records" and "not working" look identical, which is the
same trap I fell into yesterday reading a different empty result.

---

## 2026-09-03, mid-afternoon — it's live, on the second attempt

The rewrite-on-refusal work is now running in production. It went out with the chassis you
deployed at half past one.

**One thing worth knowing, because I got it wrong in front of you earlier.** When the first new
chassis went out at ten past twelve, I told you it would contain my change. It didn't — that build
had been cut from a slightly older snapshot of the code. My change was sitting in the shared
codebase, a deployment happened after it, and it still wasn't in the deployed program. Those are
two different questions and I had treated them as one.

I only knew because I asked the running program directly rather than reasoning about it: I searched
the live binary for two words that only exist in my change, alongside a word that was already there
(to prove the search was looking at the right thing) and a nonsense word (to prove it could say
no). First build: absent. Second build: present, on both machines. That's the check to trust, and
it's written into the handoff.

**What's actually working now.** If the writing rules reject a description, the system no longer
shrugs. It files a job for a small agent whose only task is to write it again, this time *told what
it got wrong*. That goes back through the same rules. Only if the second attempt is also rejected
does it reach a person — and then with the original sentence, the rule that rejected it, and the
failed retry all attached.

**The reviewers found one more thing, and it was a fair hit.** I had checked whether four other
parts of the system share the same silent-failure problem, and my check was a keyword search rather
than actually reading them. I said so in my write-up — but a reviewer pointed out that an admission
buried in a submission is invisible six months later, and asked me to file it as a proper numbered
bug so somebody can finish the job. That's now `bugs_open/464`. It says plainly that those four
files are *unread*, not *cleared*.

**And one habit I keep repeating.** Three review rounds in a row have told me the same thing:
I state that I checked something instead of showing what the check returned. Every one of those
claims turned out to be true, which is exactly why it keeps happening — and exactly why it doesn't
help. A reviewer can't tell a checked claim from an unchecked one. I had written that lesson down
myself yesterday morning and then not applied it twice. It's logged.

**Is it doing anything yet?** No, and it can't be. There are still no pages that both deserve a
description and lack one, so nothing has been refused and no job has been filed. The first one that
appears is the real proof. Until then it's live and untested, and I'd rather say that than let it
read as finished.
