# SUMMARY 2026-08-16b — the token budget fix is live, and a census of mine was wrong by 5×

## What we're trying to do

Make one configuration setting mean what it says. When the platform asks a language model to write
something, it must state how long the answer may be. That limit lives in configuration so an operator
can raise it where longer output is needed. The goal was structural: make the setting arrive on
**every** path to a model, not just the common one.

## Where we've come from

The common path already worked. One function read the configured limit and passed it along, and nearly
all calls went through it — but nothing *required* that. Code that talked to a model directly never
consulted the configuration and silently used a hardcoded floor of 2048, the smallest number in the
estate.

It was near-invisible: the configuration reads correctly, the failure echoes the *old* number back at
you after you change it, and our own truncation monitoring is written by the very function being
bypassed. Someone raised a limit twice in one day against a setting nothing read. A council review
identified the shape — one enforced entry point, optional to use — and the recommended remedy was
recorded as owed and left undone for four days.

## What we've done

Moved the decision into the model clients themselves, which were already handed the whole
configuration block at start-up and were already reading the model name out of it. They now keep the
limit too. Explicit per-call limits still win, so the common path is untouched; otherwise the client
uses its configuration; only if nothing is configured anywhere does the old floor apply.

Fixed the same hardcoded floor in the Gemini client, which the original report never mentioned because
its author had read only one of the two. Covered the image-generation path with tests after a reviewer
pointed out it shares the same entry point and I had not said so.

The review took two rounds. The first came back **revise**, and both of its substantive findings were
right: a reuse check I had genuinely skipped, and the image path I had not thought about. The second
approved, with three minor notes, two of which I settled with one search each.

## Where we are now

**Live at v1.0.1305**, and verified by asking the running binaries what built them rather than
trusting the tag — including a deliberately fabricated value in the same check, so a check that agreed
with everything would have been visible. The service this actually unlocks came back healthy, which
mattered because it was the first build ever to genuinely parse that setting.

On real traffic since the build: no step's limit changed, with enough calls afterwards to prove the
check wasn't blind, and no truncations. Small and early, and labelled that way.

**And the live traffic caught an error of mine.** I had reported that 13 services configure a limit.
Real traffic showed one sending a value my count did not have, because these settings can be written
at more than one level and I had counted only the top. The true figure is **68**. The conclusion it
supported is unaffected — that argument rests on the two code paths reading one merged set of
settings, which already includes the deeper levels — but the number was wrong in three commit
messages, the bug record and the review submission, and is corrected in all of them. The
uncomfortable part is that our own trap file warns about this exact setting at this exact level, I had
read it at the start of the work, and it did not help.

## Where we're going

**Not claimed, deliberately:** the distinguishing new behaviour has nothing in production exercising
it. The only caller of that shape is idle, its limit is small by design, and nothing sends it work. It
is proven by tests that inspect the outgoing request and fail when the fix is removed — not by
production. The bug record says so, and says not to upgrade that wording without doing the work.

Two decisions are yours, both written where the next person will find them:

1. **Whether to merge the two near-duplicate copies of the "which limit wins" rule.** Constrained now:
   the obvious direction is an import cycle, so it needs either a new shared leaf package or a
   dependency the lowest-level package currently does not have.
2. **Whether to make direct model calls visible to our truncation monitoring.** This matters more
   after this change, not less: such calls used to be invisible *and* pinned to a known small limit,
   and are now invisible and running at whatever their configuration says. Every instrument we have
   built for truncation reads the table these calls never write to.

Re-running the traffic comparison after a day of normal load is the one cheap thing left.
