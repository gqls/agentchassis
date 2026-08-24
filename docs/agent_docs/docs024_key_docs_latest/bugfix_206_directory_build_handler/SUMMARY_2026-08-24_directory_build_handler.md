# SUMMARY 2026-08-24 — directory-build-handler (`bugs_open/206`): the fix was right, and it was only ever on one of the doors

*New file, not an edit of `SUMMARY_2026-08-08`. That one recorded what we believed at the
milestone; this one records what turned out to be true sixteen days later. The difference
between them is the useful part.*

## What we're trying to do

Make the platform able to build the kinds of page it had never been able to build: a directory
page that lists real businesses, an index page that lists a site's real guides. Not by hand —
through the ordinary pipeline, so any site can have one and nobody has to remember a trick.

## Where we've come from

In August we found that these page types were, in the platform's own words, "builders that don't
exist yet" — named in a comment and never written. We wrote the missing pieces: a resolver that
pulls a site's real business data from the same source its public export uses, and an action
that fills in a page's layout only when it has none, so it can never overwrite a choice someone
else made. A new handler chains them and hands off to the ordinary page builder.

That passed the council review gate on the third round, and two more faults appeared only when
real builds ran — both fixed the same day. On 8 August both pages went live through the normal
queue with no manual dispatch, which was the whole point: the directory page listed sixty-one
real practices, the guides page listed exactly the three guides that exist and invented nothing.

The summary written that day said: *"Nothing further is owed on this specific bug."*

## What we've done

Picked the lane back up today and checked that claim, because a fix that was live three weeks
ago is not evidence about today.

**The 8 August work holds.** Both pages still serve their real content — and, more convincingly,
a fleet-wide re-render swept them yesterday and the real listings survived it, so the pipeline
genuinely reproduces the result rather than a good page having been frozen in place. A separate
team pushed a third kind of page through the same handler this morning, first attempt.

**But the claim that nothing further was owed was wrong, and in an instructive way.** The
platform decides "which builder builds this page" in more than one place. In August we taught
one of those places about the new handler. A second one — the routine that reconciles a site's
plan against what has actually been built — never asks the question at all: it had the answer
typed directly into it as text. So it has gone on sending these pages to a builder that cannot
build them, where they fail, park, and wait for a human who never comes.

Five pages sit in that state right now, across three sites we have never worked on. The one that
settles it is a directory page on garden-tools.uk: it has been waiting fifteen days for a builder
that existed and worked the whole time. Nobody ever told it the builder's name.

So the fix today is not another builder — it is making that decision exist **once**, with both
routines asking it, and putting tests on it. There were none before, in either copy.

## Where we are now

The change is committed, reviewed-pending (the council sent round one back and was right to — we
had cited a database column that does not exist), and it is Go code, so it does nothing until the
fleet's images are next rebuilt. The bug file stays open, deliberately: the fault is still
happening on the fleet today, and it stops when the fix ships, not when it is written.

Three things are worth the owner's attention more than the code is.

**One decision is not ours to make.** Another team examined this same set of stuck pages three
weeks ago and deliberately chose to leave that routine alone, reasoning that stuck items are real
findings and should stay visible. Half of today's change touches exactly that. Our view is that
it keeps them just as visible and describes them more honestly — nothing is hidden — but they
decided it on purpose and with the council's backing, so it has gone to the council as an open
question, the notice is in their file, and we have offered to take that half back out.

**We are not fixing the biggest part of the problem, and should not be read as if we were.**
Once measured properly, this class is 87 stuck items across 16 sites. Sixty-nine of them are tool
pages, and most of those are being created by a *different* bug we ourselves filed in August
(`bugs_open/220`) — which means that queue is being refilled faster than anyone drains it. Our
change reaches about eleven items. The rest is a separate piece of work with a separate cause.

**Our own checking failed three times today, and each time somebody else's habit caught it.**
We told another team something about our own lane's history that our own notes had already
corrected. Our first survey of how widespread this is returned *zero* — a real number, produced
by a question that could not have found the answer, and it took a passing remark from a
neighbouring session to expose it. And we cited that non-existent column to the council, which
found it independently. All three are written up in the fleet's log of wrong calls, with the
cheap check that would have caught each one. The pattern in all three is the same: the part we
had verified that morning was right, and the part restated from memory was wrong, and they
arrived in the same confident sentence.

## Where we're going

When the next fleet build lands: confirm the code is actually in the running services, watch a
reconcile run file a typed page at the right builder, and re-trigger the five stuck pages. Then
the bug moves to the closed set — and not before, because until it ships the fault is real.

Two things stay deliberately undone. There is still a third copy of that routing decision, in a
file nobody can safely commit to right now because someone else's unfinished work is sitting in
it; the swap is recorded as owed, in the code and in the register, so whoever finds that file
clean can finish it in minutes. And we are still not building individual entity pages — profiles
for a single practice, a single brand — because doing that automatically means inventing facts
about real businesses, which is the one thing this platform must not do. That waits for real
source data and a person deciding what those pages are.
