# Where we are — the token budget bug (bugs_open/257)

Plain prose, append-only, newest at the bottom.

## 2026-08-16 — what this is, and why it was worth doing

Every time this platform asks Claude (or Gemini, or a local model) to write
something, it has to say how long the answer may be. That number is called
`max_tokens`. If it is too small the model stops mid-sentence, and what you get
back is a truncated fragment that often still *looks* like a finished answer.

We keep that number in configuration, so an operator can raise it for a step that
needs a long answer. The bug was this: **for most calls the configuration worked,
and for some calls it was quietly ignored** — and there was no way to tell which
kind of call you were looking at.

The reason is a bit of plumbing. There is one function in the system that reads
the configured number and hands it to the model. Almost everything goes through
that function, so almost everything works. But nothing *forces* you to go through
it. If you write code that talks to the model directly, the configuration is
simply never read, and the model gets a hardcoded fallback of 2048 — the smallest
number we use anywhere.

Three things made it very hard to spot, and they compound:

1. **The configuration looks right.** You go and read it, and there is your 8000,
   sitting exactly where the documentation says it should be.
2. **The error repeats the old number back at you.** Someone hit this in August:
   they set the value to 8000, the next run still failed saying "2048", so they
   assumed 8000 was still too small and changed the configuration *again*. Both
   changes were against a setting nothing was reading.
3. **These calls are invisible to our own monitoring.** The table we use to spot
   truncation is written by the same function that gets bypassed. So the one
   report that would tell you "these calls are truncating" is blind to precisely
   the calls most likely to be truncating.

## 2026-08-16 — what we did about it

We moved the decision to where the request is actually built. The model clients
were already being handed the whole configuration block at startup — they were
already reading the *model name* out of it — they were just throwing the budget
away. Now they keep it.

So the rule is now: if the code makes a specific request for a length, that wins
(nothing changes for the vast majority of calls). Otherwise the client uses
whatever the configuration says. Only if nobody has configured anything at all
does it fall back to 2048.

**The important property is that this changes nothing today.** We checked all
seven places in the codebase where one of these clients gets built, one at a time,
rather than reasoning about the common case. Every single one either already
specifies a length explicitly, or has no length configured. So no request the
fleet sends is different after this change. What changes is that a setting which
was decorative is now real — and the next person who writes code that talks to a
model directly cannot accidentally reproduce the bug, because there is no longer a
way to.

We also found and fixed the same hardcoded 2048 in the **Gemini** client. The
original bug report did not mention it, because whoever wrote it had read only the
Anthropic client. Fixing one and leaving its twin is a mistake this codebase has
made several times before, so it was worth the extra half hour.

## 2026-08-16 — one open question in the report, now answered

The bug report flagged a service called `reasoning-agent` as permanently stuck at
2048 with no way to configure it, but was careful to say it had **not** checked
whether this was actually hurting anything, and warned that the usual monitoring
could not answer that.

We checked properly, at the running service rather than in the database. It has
three copies running, and **none of them has ever handled a request** — their
entire log is start-up messages. Searching the codebase, nothing anywhere sends it
work. It is a deployed service listening to a channel nobody speaks on.

So: the defect was real, the harm was zero. We have made its setting live, and
deliberately **left the value alone at 2048**. Raising a limit with no evidence
that anything is being cut off is exactly the mistake described above, and we are
not going to make it twice in the same file.

## 2026-08-16 — two things we got wrong along the way

Both are written up properly in the technical notes, but they are worth saying in
plain terms because they are the same kind of mistake.

**First:** adding the new setting to the clients meant that some of our own test
fixtures would be built without it, and would end up asking for a length of zero —
which the API rejects outright. We expected the test suite to catch that
immediately. It did not. Everything passed. That silence was the actual problem:
it meant our tests were not looking at this value at all. Rather than adjust the
fixtures until they agreed, we made it impossible for a zero to reach the API in
the first place.

**Second, and more embarrassing:** we wrote a test to prove that a nonsense value
(a length of zero in the configuration) would be safely ignored. Then we
deliberately broke that safety check to confirm the test would notice. **It did
not notice.** The test passed against broken code.

The reason is that we had, by then, added a second safety net further downstream,
and *that* one was catching the bad value. The test was proving the second net
worked, while claiming to prove the first one did. Both nets were fine; the test
was a lie about which. We rewrote it to check the first one directly, with nothing
downstream able to rescue the answer, and it now fails properly when broken.

The general lesson — and it is a good one — is that **when two safety checks sit
on the same path, an end-to-end test can only ever tell you that at least one of
them worked.** If you want to claim a specific one is tested, the test has to
reach it with nothing after it able to paper over the result.

## 2026-08-16 — where this leaves us

The code is written, tested, and has been through the review council. Nothing is
live until the next fleet build goes out, because Go code only takes effect when a
new image is rolled.

There is one deliberate piece of unfinished business, and it is recorded rather
than forgotten. Four reviewers previously asked for two near-identical copies of
this same "which number wins" rule to be merged into one. We have not done that
here. Our change removes the *damage* that duplication could cause — the cost of
getting it wrong used to be a silent truncation and is now just an inconsistency —
but the two copies genuinely are not the same, in three specific ways, and merging
them would change behaviour on the single busiest code path in the system. That is
its own job, with its own review, and the three differences are written down so
whoever picks it up does not have to find them again.

## 2026-08-16 (later) — a correction: the review never actually ran the first time

I told you earlier that the change had gone to the review council and we were waiting on a verdict.
**That was wrong, and I want to be plain about it.** The submission had already been refused before I
said it — I just had not looked at the right field yet.

The council review takes about half an hour to come back, so an empty result means "still queued"
almost every time, and our own guidance says not to keep re-checking on that evidence. What I did not
know is that a *refused* submission looks identical to a queued one: the job's status reads
**COMPLETED**, which everywhere else in this system means it worked. The word that tells you it failed
is in a different column entirely.

The cause was mundane. A submission is a list of proposed edits, and two of mine described a change
that genuinely spans two files, so I named both files in one entry — "this file plus that file". The
server requires exactly one file per entry and rejects anything else. It reads perfectly well to a
person, it passed every check our own submission tool runs, and it was refused.

**The part worth telling you is that the refusal left no explanation anywhere.** No error message, no
record in any of the three places such things are normally written. I could not read why it failed; I
had to work it out by reading the validating code and then re-running its exact rule against my own
submission, which pointed straight at the two offending entries.

Three things came out of it.

The submission is fixed and resubmitted — restructured so every entry names one file, without dropping
any content (two entries were about the same test file, so merging them freed the room).

Our submission tool now checks this before sending, so nobody loses half an hour to it again. That
tool already had four such checks, each added after somebody hit the same kind of silent rejection;
this is the fifth. I also proved the new check actually catches the thing it is for, on three
different bad shapes, plus a check that it does not simply reject everything.

And there is one thing I cannot undo. Three commits already record the dead submission's reference
number, and this project does not permit rewriting commits. So the record now carries a follow-up
commit naming the live one and saying plainly that it replaces the other. That is untidy, and it is
the honest version.

The underlying code change is unaffected by any of this — it was written, tested and committed before
the submission was ever sent.

## 2026-08-16 (end of session) — the review passed, and the round that failed was the useful one

The change has been through the review council and is **approved** — three advisory comments, none of
them serious, and none blocking.

It took two rounds, and the first one is the part worth telling you about. It came back **revise**, and
two of its three substantive objections were right.

The first: I had written a small helper to read a number out of configuration, and a reviewer asked
whether I had checked that the codebase already has helpers for exactly that. **I had not.** I went and
looked. There are two, and neither fits — both of them collapse "nobody chose a limit" and "somebody
chose zero" into the same answer, and that distinction is the whole reason my version exists, because
it is what lets one of the three model providers deliberately send nothing where the other two send a
default. So the helper stays, but the reviewer was right that I should have checked before writing it,
and the check is now written into the code so nobody has to ask again.

The second: a reviewer pointed out that these model clients have a second entry point for requests that
include images, that our own internal warnings list names it alongside the ordinary one, and that my
description never mentioned it. As it happens the image path already runs through the same code I
fixed, so it was working. But "it happens to share the code path" is a statement about code that
somebody can restructure next week, and this project's own rule is that an unverified claim is itself
the problem. So there are now three tests covering it, and I checked they genuinely fail if that
sharing is broken.

The second round then approved, and left three minor notes. Two of them I could settle with a single
search each, and I did — one asked whether I had counted a category of risk the same way I had counted
the others, and the honest answer was no, I had counted the one I thought of. Both came back clean.
The third was a fair criticism of presentation: one of my eight items was a comment-only change that I
had not labelled as clearly as the other two, so it could read as more than it was.

That is the real argument for these reviews, and it is not that they catch disasters. Both useful
findings were a single command away, neither had been run, and both came from a question I would not
have asked myself — because I had already done the version of that check that occurred to me.

**Where this leaves the work:** finished and approved, waiting only on the next fleet build to become
live. Nothing else is owed on it. Two things are deliberately left for you to decide, and both are
written down where the next person will find them: whether to merge the two near-duplicate copies of
the "which limit wins" rule, and whether to make direct model calls visible to our truncation
monitoring — that second one matters a little more now than it did before, because such calls used to
be invisible *and* stuck at a known small limit, and are now invisible and running at whatever their
configuration says.

## 2026-08-16 (evening) — it's live, and the live data immediately caught me out

A fresh build went out (v1.0.1305, around 22:08). The change is **live**, and I checked it the way
this project insists on: by asking the running programs what they were built from, rather than
trusting the version tag.

One service answered directly and its answer contains my fix. The main one had already scrolled its
start-up message out of reach — a known trap here — so I checked inside the running binary instead,
and in the same command checked for a made-up value that must *not* be there, so I could tell the
check was actually working rather than agreeing with everything. Both came out right. The small
configuration change shipped too, with the same kind of control.

I also confirmed the one service this actually unlocks came back healthy — three copies, no restarts,
no errors. That mattered more than it sounds: this is the first build in that service's life where
that setting is genuinely read, so a mistake in how I parse it would have shown up as a crash loop.

**Then the live data caught an error of mine, and it is worth telling you about.**

When I sized the risk of this change, I counted how many services configure a length limit and
reported **13**. Watching real traffic after the build, one service was sending a limit of 8192 where
my count said 4000. The reason is that these settings can be written at more than one level of the
configuration, and **I had only counted the top level**. Counted properly, it is **68**, not 13 — I
was out by a factor of five, and that figure had gone into the bug record, the review submission and
three commit messages.

The conclusion it was supporting is unaffected, and I want to be precise about why rather than wave
it away. The argument for this change being safe never depended on the count. It depends on a
structural fact: the code that builds the connection and the code that sets the limit read the *same*
merged settings, and that merge already includes the deeper levels. So they cannot disagree, whether
it is 13 services or 68.

But that is exactly what makes the wrong number worth flagging. It was decorating an argument that
did not need it, so **nothing broke when it was wrong** — it would simply have sat there until
somebody quoted it for a different purpose. The uncomfortable detail: this project keeps a file of
known traps, I read its entries on this exact setting at the start of the work, and it still did not
save me, because I was asking a different question and never noticed it was about the same thing.

**One thing I have deliberately not claimed.** The specific new behaviour — a caller that asks for no
limit now inheriting a large configured one — has nothing in production that exercises it. The only
caller of that shape is the idle service, its limit is deliberately left small, and nothing sends it
work. So that behaviour is proven by tests that inspect the actual outgoing request, which I verified
fail when the fix is removed, but it is not proven by production traffic. I have written that into the
bug record explicitly, including an instruction not to describe this as "verified end to end" without
doing that work, because a previous piece of work on this list *could* honestly say that and this one
cannot.

What I could check on real traffic, I did: no step's limit changed across the build boundary, with
enough calls afterwards to show the check wasn't simply looking at nothing, and no truncations either
side. That sample is small and early, and I've labelled it that way.

## 2026-09-03 — picked this back up after two and a half weeks, and the same fault has grown back

I went back to this to check whether it was still worth working on, and whether anyone else had it in
hand. Nobody did — the lane had been quiet since 17 August — so I took it. **I have written no code
today.** A new build went out while I was working; it is another piece of work and contains nothing of
mine. Everything below is findings.

**The part that was fixed is still fixed.** The change from August, where the model clients learned to
read the length limit out of their own configuration, is live and behaving. Nothing here retracts it.

**But the underlying fault has come back twice, in code written after that fix.** Two new pieces of code
call a model directly, and each one re-implements the length-limit rule by hand, ending with a number
typed straight into the source: 2000. One was added on 20 August, the other on 31 August.

There is an irony in this that I want to state plainly, because it changes how the remaining work should
be done. The August fix works by having the client supply the configured limit **when the caller asks
for nothing**. These two new callers always ask for something — the configured value if they can find
it, otherwise their typed-in 2000. So the fix can never help them. **Asking for nothing is now safer
than asking for a specific number**, which is not a rule anyone would guess, and is exactly why this
needs closing at the call sites rather than explained in a comment.

**The thing I want you to know about, because it nearly caught me.** Our own records show four of these
calls being cut short at exactly 2000 tokens in late August. The number in the code is 2000. I was one
sentence away from writing that the typed-in number had damaged live pages.

It had not. Another team had set that step's configured limit to 2000 on 21 August, deliberately and
with reasons, then measured that it was too small and raised it to 16000 on 23 August. So the cut-short
calls were the *configured* value being too small — a real problem, already found and already fixed
properly.

What makes this worth telling you rather than just correcting: **the configured number and the typed-in
number were the same number.** Every record we keep — the limit sent, the error text, the length of the
reply — reads identically whether the configuration was being obeyed or completely ignored. There was no
query I could have run that would have told me which. I only got it right by opening two old change
files and reading why each was written.

That matters beyond this one mistake, because the other of the two new pieces of code is in the same
position **right now**. Its configured limit is 2000 and its typed-in fallback is 2000, so we currently
have no way of knowing whether it is reading its configuration at all. It is not doing any harm — its
answers are about a third of the limit — but the instrument we would rely on to notice if that changed
is blind here by construction. I have written that up as a standing trap so the next person checking
"did the setting take effect?" does not lose the same hour.

**What I have not done, and why.** The fix itself is small and I know exactly what it is: delete both
hand-written copies and call the shared helper that already exists a few files away. I have not written
it, for one reason — the guard that would stop this happening a sixth time has to land alongside it, and
on this shared tree a guard that fails on code already in place would break the build for every other
session working today. That wants to be one careful change with a review, not a rushed one at the end of
a session. It is written up in full, with the three behaviour changes it causes and the measurements
showing each is harmless on today's fleet, ready for whoever picks it up next.

**Two decisions are still yours, unchanged from August:** whether to merge the near-duplicate copies of
the "which limit wins" rule, and whether making direct model calls visible to our truncation monitoring
belongs in this bug or a lane of its own. Today's finding pushes on the second one: four of the six
direct callers inside the main service now do report themselves, which is better than August, but the
reporting cannot distinguish the two cases described above, so more of it is not automatically more
truth.

---

**2026-09-03, later the same day — the fix is written, and it found three more problems on the way**

Picking up from the note above: the plan was to delete the two hand-written copies of the "how long may
the answer be" rule and call the shared helper instead, and to add a check that stops it happening a
sixth time. Both are done and committed. It has gone to the review council; the verdict usually takes
about half an hour and I will act on it when it lands.

The interesting part is what happened while writing the check.

The plan was based on a list of every place in that part of the system that talks to a language model.
That list has been compiled four times over the last three weeks and has been wrong every time. It was
wrong again. Two entries were marked "fine" and were not: one was handing the model an empty settings
object, so the step's own limit could never reach it, and the other would have passed a limit of zero
straight through if anyone ever typed one, which the provider rejects outright. And one place was
missing altogether — the news-fetching code talks to a different supplier over plain web requests
rather than through our usual client, so **every list we have ever made was structurally incapable of
seeing it.** It had a limit of 4,096 typed into it.

The reason the check found what four hand-compiled lists did not is that it asks a different question.
The lists asked "who calls our model client?" The check asks "is there a number typed in anywhere near a
word that means *how long may the answer be*?" — which does not care how the request leaves the
building. All five places now read the limit from configuration; the news fetcher keeps 4,096, but as a
default someone can override rather than a number nobody can reach.

**Does anything change for the sites we build today?** No, and I measured that rather than assuming it.
Every step affected already states its limit in configuration, and each will send exactly the number it
sends now. What changes is that the numbers are now genuinely under your control, and that the next
person who writes one into the code will be told, by a failing build, on the day they do it — not three
weeks later by someone reading the logs.

**I also proved the check can fail.** A check that has never been seen to go red is not evidence of
anything. I put each of the four mistakes back in, one at a time, in a throwaway copy of the code, and
confirmed each one is caught with a message that says what to do about it. Then I put it all back.

**One new thing I found and deliberately did not change.** Four steps of the site-adoption agent state a
limit — one of them asks for 32,000 — in a place nothing reads. All four are actually running at 16,000,
which is the general default for that agent. Nothing is being cut off today, so this is not urgent, but
somebody once asked for double and quietly got half. Moving those settings to where they would be read
is a live change that takes effect the instant it is applied and would raise what we spend on those
steps, so that is your call rather than mine, and I have written it up rather than acting on it.

**The two decisions from August are still open and unchanged** — whether to merge the two near-duplicate
copies of the "which limit wins" rule, and whether making direct model calls visible to our
truncation monitoring belongs in this bug or a lane of its own.

---

**2026-09-04 — it is live, and there is one decision I would like from you**

Your chassis build went out last night at 22:06. The change is aboard and running: 208 calls have gone
through the five affected places since the roll, every one of them sending exactly the limit its
configuration states, and nothing has been cut off.

**How I know it is aboard is worth one paragraph, because it is not the obvious way.** The service
normally announces which version of the code it was built from, in a line it prints when it starts —
but that line had scrolled out of reach ten hours later. My own attempt to read it out of the running
program failed: I asked it to check 474 possible answers at once, and the check was killed part-way
rather than answering, which would have looked exactly like "the change is not there" if I had taken it
at face value. What settled it was another team member's work: a different fix, committed twelve minutes
after mine, was proven to be in last night's build by a careful test with four controls. Since my change
is in the history *underneath* theirs, any build containing theirs contains mine. One command, and it is
solid.

**The uncomfortable part, said plainly.** This change was designed to alter nothing, and it succeeded —
which means no measurement taken after the fact can tell the new code from the old. Every step involved
already stated its limit properly, so both versions send the same numbers. Even the check the bug file
prescribes for exactly this moment — "watch a step whose limit is larger than the typed-in number send
the larger number" — cannot help, because that step was sending the larger number *before* the roll too.
This is the very same trap the round was about, turning up inside my own verification, one level out.

**So here is the decision.** There is one test that would settle it by behaviour rather than by
inference: put a limit in a place the *old* code never looked and the *new* code looks first, on one
step, and read what the next call sends. Old code would send 16,000; new code would send 15,999. One
setting, one token of difference, reversible in seconds.

I have not done it, because it writes to a live production agent's configuration and takes effect
immediately, and that is your call rather than mine. It is worth something beyond satisfying curiosity:
it would be the first time anything on the fleet has exercised that "a step can override the service
default" path at all, so it tests a capability we currently only believe in.

**The other two decisions are unchanged and still yours** — merging the last two copies of the "which
limit wins" rule, and whether making direct model calls visible to our truncation monitoring is part of
this bug or its own piece of work. And one new one from yesterday: four settings on the site-adoption
agent that ask for a limit nothing reads, one of them asking for double what it gets.
