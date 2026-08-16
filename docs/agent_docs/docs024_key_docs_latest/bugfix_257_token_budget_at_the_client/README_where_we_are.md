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
