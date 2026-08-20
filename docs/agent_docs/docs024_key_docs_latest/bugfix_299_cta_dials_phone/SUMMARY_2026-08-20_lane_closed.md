# SUMMARY 2026-08-20 — lane CLOSED: the button is fixed, and the framework learned three things

*(Second and final summary for this lane. The first, `SUMMARY_2026-08-19`, was written at the
milestone where protection went live and the repair was one decision away. This one is written
because the read-out has genuinely changed: the bug is closed at the served page and the lane
has nothing left of its own.)*

## What we were trying to do

A button on the webdesign.uk home page had words about one thing and a link to another: the copy
talked about the Brief Starter tool, and clicking it opened the phone dialler. The owner found it
himself — which was the part that mattered, because no queue had caught it and it had survived
several complete rewrites of the page.

So the job was never "fix the button". It was to answer why a system that already had a fix for
misdirected buttons produced this one, and to close that at the framework level.

## Where we came from

Four separate faults stacked on top of each other. The part of the pipeline whose whole job is to
work out where each button should point was **computing the right answer and having it thrown
away** — the step that hands finished sections to the page writer looked for that answer under a
name it was not stored under, and fell back silently to the previous build's values. The checker
for "the words promise one thing, the link does another" skipped phone and email links entirely,
so this exact button was invisible to it. The repair machinery, had it touched the button, would
have destroyed genuine "call us" buttons elsewhere. And the phone numbers themselves were written
in a form phones cannot dial.

The fix was built, calibrated against the whole fleet before shipping, reviewed by the council,
and committed with three switches deliberately left off — because throwing the wiring switch
before the protective code was live would have started clobbering good buttons everywhere.

## What we did

All three switches are thrown and the bug is closed at the served page: the button now reads
"Read the full terms in our FAQ before you pay." and it goes to the FAQ. Copy and destination
agree.

The framework half is the durable part. **Before the wiring fix, none of the computed link
answers survived to the page — 0 of 33. After it, all of them do.** Along the way we found a
sharper way to measure it: the link resolver writes a small note recording what each button points
at, and nothing else produces those notes, so a missing note *is* the discard — a test that works
even on the builds where the link happens to be right already.

Three of the five malformed phone links on the site repaired themselves as their pages rebuilt,
with no human involved. The two that could not be decided by a machine — an undialable number, and
whether a phone button was intentional at all — went to the owner and came back answered.

We also built the thing this bug forced us to notice. The estate's standard way of confirming new
code is live turns out to be frequently *impossible*: the version line scrolls out of the log
within hours, and the fallback check returns a confident wrong answer. Two separate threads had
already been burned by it. There is now a table that answers it in one query, ratified and scoped
by the owner to exactly that — recording, not enforcing.

## Where we are now

Closed. The bug is fixed and live; the framework defect underneath it (`bugs_open/312`) is fixed
and proven, though it stays open for two tripwires it has earned. The new detector is live and its
first run caught the very button that had to reach the owner's eye. The council approved both
rounds of work.

Three things were got wrong along the way and all three are written down: a measurement that
returned the exact opposite answer because it cast too wide a net; a council objection I was one
step from dismissing with a number that was the wrong *shape*; and an urgency figure I overstated
by seventeen times by extrapolating from a burst. The last two were caught by the tool this lane
built, answering questions it had not been asked.

The only live thread is housekeeping: the capability table clears itself from the next release
onwards, and until then a one-line command does it. That belongs to whoever runs the next release.

## Where we are going

Nowhere, as a lane — this one is done. What outlives it:

- **`bugs_open/312`** wants a loud fallback and a lockstep test. That seam has now failed
  silently in both directions twice, so those are earned rather than speculative.
- **Two more instances of this bug's class** sit in review on the same site, deliberately not
  hand-fixed: their phone links are genuine, so the fix is the copy — which makes them the first
  real test of the destination stamp this lane shipped.
- **`RFC_040`'s second half** — letting a configuration change refuse to apply itself against the
  live binary — waits for a second real demand, by the owner's ruling. Whoever wants it should be
  able to name two migrations that do.
