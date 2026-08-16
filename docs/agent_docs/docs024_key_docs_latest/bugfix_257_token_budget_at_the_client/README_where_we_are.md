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
