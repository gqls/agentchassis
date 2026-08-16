# SUMMARY 2026-08-16 — the token budget now lives with the thing that sends the request

## What we're trying to do

Make one configuration setting mean what it says. When this platform asks a
language model to write something, it must tell the model how long the answer may
be. That limit lives in configuration so an operator can raise it for work that
needs a longer answer. The goal of this piece of work was narrow and structural:
ensure that setting is honoured on **every** path to a model, not just the common
one.

## Where we've come from

For most of the system it already worked. There is a single function that reads
the configured limit and passes it along, and nearly all model calls go through
it. But nothing *required* you to go through it. Write code that talks to a model
directly and the configuration was never consulted — the request went out with a
hardcoded fallback of 2048, the smallest number we use anywhere.

That gap was found the hard way in August. Someone raised a step's limit to 8000,
the next run still failed reporting 2048, so they raised it again — both changes
against a setting nothing was reading. The error kept repeating the old number
back at them, which reads as "still too small" and actually means "your number
never left the building".

Three things made it near-invisible and they compound. The configuration looks
correct when you read it. The error echoes the stale number. And our own
truncation monitoring is written by the very function being bypassed, so the one
report that would say "these calls are being cut short" is blind to precisely the
calls most likely to be cut short.

A council review at the time flagged the underlying shape: the contract had one
enforced entry point and using it was optional, so any future author who reached
for the direct call would reproduce the bug by construction. Four reviewers
independently said so. The recommended remedy — move the decision into the client
itself — was recorded as owed and not done. It stayed owed for four days and was
written up as `bugs_open/257` on 12 August. Nobody had picked it up.

## What we've done

Moved the decision to where the request is actually built. The model clients were
already handed the whole configuration block when they start up — they were
already reading the *model name* out of it — they were simply discarding the
length. Now they keep it.

The resulting rule: an explicit per-call length still wins, which is what the
common path always supplies, so that path is untouched. Otherwise the client uses
what its configuration says. Only if nobody configured anything anywhere does it
fall back to 2048.

Three things are worth drawing out.

**We proved it changes nothing today rather than assuming so.** There are seven
places in the codebase where one of these clients is built. We went through them
one at a time. Every one either already specifies a length explicitly or has none
configured, so not a single request the fleet sends differs after this change.
What changes is that a decorative setting became a real one — and that the next
person to write a direct model call cannot reproduce the bug, because the way to
do it no longer exists.

**We fixed the twin the original report missed.** The Gemini client carried the
identical hardcoded 2048. The bug report never mentioned it, because whoever wrote
it had read only the Anthropic client. This codebase has a documented habit of
fixing one side of a shared judgement and leaving its sibling, so we did both, and
the local-model client as well.

**We answered a question the report had deliberately left open.** It flagged one
service as permanently stuck at 2048 but was careful to say it had not checked
whether that was hurting anything, and warned the usual monitoring could not
answer. We checked at the running service instead: three copies running, not one
request ever handled, and nothing anywhere in the codebase sends it work. It is
listening to a channel nobody speaks on. So the defect was real and the harm was
zero — and we deliberately left its limit alone rather than raising a number with
no evidence behind it, which is the exact mistake described above.

## Where we are now

The code is written, tested and committed, and it has gone to the review council.
Eleven deliberate breakages were applied to confirm the tests actually notice when
the fix is removed; nine were caught immediately, one exposed a genuine flaw in
one of our own tests, and one turned out to be a change with no effect at all.

That flaw is worth stating because it is the most useful thing we learned. We had
written a test to prove that a nonsense configured value would be safely ignored.
When we broke that safety check on purpose, the test still passed. The reason was
that a second safety net further downstream — added for an entirely unrelated
reason — was quietly catching the bad value. Both nets worked; the test was
claiming to prove the wrong one. The general rule, now written down: when two
safety checks sit on the same path, an end-to-end test can only tell you that at
least one of them held. If you want to claim a specific one is covered, the test
has to reach it with nothing downstream able to repair the answer.

Nothing is live yet. Go code only takes effect when a new image is built and
rolled out, so this waits on the next fleet release. The one service whose setting
this actually unlocks is a separate service with its own image, and will need its
own build.

## Where we're going

Three things, in order.

First, verification after the next roll — and by asking the running service what
it was built from, not by looking at git or a version tag. The behavioural proof
is that a configured limit above 2048 can now produce an answer longer than 2048;
the configuration value is not the measurement, the response is.

Second, there is one deliberate piece of unfinished business. Four reviewers
previously asked for two near-identical copies of the same "which number wins"
rule to be merged. We have not done that, and we have said why: our change removes
the damage the duplication could cause, but the two copies genuinely differ in
three specific ways, and merging them changes behaviour on the single busiest
model-calling path in the system. That deserves its own change and its own review.
The three differences are written down so whoever takes it does not have to find
them again.

Third, a smaller loose end recorded but not acted on: the idle service also pins a
two-generation-old model. It costs nothing while nothing calls it, but anyone
wiring up a caller should look at that first.
