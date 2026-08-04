# Where we are — bug 192, the section plan that got buried one level down

Plain prose, append-only, newest at the bottom.

---

## 2026-08-04, morning

**What was broken.** Since about 09:20 this morning, every page the platform tried to
build failed. Not some of them — all of them. Three lanes hit it independently within
half an hour: a vet-comparison guide, a tool page on another site, and webdesign.uk's
own landing page, which is the shopfront for the site-building product and could not
be built at all.

The error everyone saw said a key called `sections_ready` was missing. That was true,
and it was the wrong thing to look at.

**What actually happened.** Yesterday a different lane fixed a real bug (178): when we
asked the writer to make a small edit to an existing page — "add a link to X" — it was
regenerating the whole section from scratch and throwing most of the prose away. Their
fix was to hand the writer the page's current content first. Good fix, and the part
that does the handing works exactly as intended.

The problem is where it put the result. The build pipeline passes a "section plan"
between steps — a list of which sections to write. The new step was told to write its
answer back into that same slot, deliberately, so nothing downstream needed changing.
But instead of putting the plan back, it put **a box containing the plan** back — the
plan, plus two notes about what it had just done.

Everything downstream then looked in the slot, found a box where it expected a plan,
and found nothing. The data was never lost. It was one level deeper than everyone was
looking, and because nothing was missing and nothing errored, the step that caused it
reported success and moved on. The failure surfaced two steps later, in a completely
different file, naming a completely different thing.

That is the whole bug. It is worth saying plainly because the shape recurs: **the
wrong answer looked exactly like the right answer.** All the data was present. Every
status said complete. A query written to check it would have read the path we expected
and quietly returned nothing, which looks the same as "this was never here".

**One cause, not two.** The thing that made it confusing to the lane that filed it is
that it broke *both* of the fallback routes the writer has. The second route broke
directly. The first route broke because the link-resolver step reads from the same
slot, so it was handed nothing, did nothing, and returned an empty answer — which
read like a separate, second fault. It isn't. Fix the box and both routes come back.

**What I changed.** Four things, and only the first is really "the fix":

1. The new step now puts the plan back, not a box round it. Its notes go to the log.
2. The step that does the looking-up can now be told "this field is not optional". It
   used to shrug when it found nothing and let the run continue into a wall; now, when
   asked to, it stops there and says which field it wanted, where it looked, and what
   was actually available. This is off by default everywhere and switched on for the
   one step that needed it — I have not changed how anything else behaves.
3. The unhelpful error message now lists what it *did* find. Had it done that this
   morning, this would have been a five-minute bug.
4. A config change, applied immediately, that lets the running system cope with the
   box while the code fix waits for the next release. It retires itself: once the code
   ships, it is never consulted again.

**Is it fixed?** The outage is over and I have proof rather than an assumption. I took
one of the failed jobs, put it back in the queue, and watched it: it went all the way
through and finished. The three jobs immediately before it had all died. That check
could have come out the other way, which is why it is worth quoting.

**What is still owed.** Point 1 above is Go code, and Go code does nothing here until
a new image is built and rolled out — that is a whole-fleet release and not mine to
run. So today the box is still being made on every build; the config change is simply
stepping around it. When the next release goes out, the box goes away at source and a
short follow-up removes the stepping-around. Until then I would not describe this bug
as finished, and I have left the ticket open saying exactly that.

**Two things worth knowing beyond this bug.** The unit test for the broken step
asserted the box — so it *passed* on the code that took the fleet down, while the
comment three lines above it described the correct behaviour. I rewrote it, and then
deliberately broke the code again to check the new test would actually catch it. It
did. And the original bug report's timing was wrong in a way that mattered: it said
the failures started last night, which made yesterday's fix look innocent. Splitting
the failures by *which step* died shows two unrelated problems in one bucket, and this
one started this morning, twenty minutes after the config for it went live. The
overnight failures are a real and separate problem that nobody has looked at yet.

## 2026-08-04, late morning — closed

The new chassis went out and I could finish the job properly.

**It works, and I checked it the way that could have told me it didn't.** I looked inside
the two running copies of the program for text that only exists in the new version — plus
one line that was there before, as a control, to prove I was reading the right thing at
all. Both copies had everything. Then I made the system build a real page and watched the
data: the section plan came through in the right shape, not in the box. Both halves of the
build finished cleanly.

Then I removed the temporary workaround from this morning, and it turned out to matter
*how* I removed it. Another team's change had landed in the same list in the meantime.
Had I deleted "the third item" the workaround would have gone and so would theirs —
silently, with everything still looking plausible. I deleted it by name instead, which I'd
decided to do that morning for exactly this reason and half-expected never to need. It
needed it within the hour.

**The bug is closed.** Fixed properly, live, verified, and reviewed — the review panel
approved it on the second pass, after correctly telling me the first version had patched
the symptom's two locations while leaving the underlying mechanism open. Acting on that is
what turned up the second broken thing, in a completely different part of the system, that
nobody had noticed because it had never visibly failed.

**Three things now sit with other people, and all three are written down where they'll be
found**: the team whose fix caused this can finish the verification it blocked; the
webdesign.uk team is unblocked and can build its shopfront page whenever it likes (I left
that to them — it publishes a real page on a real site); and a human needs to decide
whether the platform should warn about this whole category automatically, which is now
written up in an existing open design question rather than a new one.

One thing I noticed and did **not** chase, because I'd only be guessing: there's a
component that gets started on every single page build and appears never to have done any
actual work in the four and a half months our records go back. It might be harmless. Nobody
has looked, and I didn't want to file a report on a symptom I hadn't understood.
