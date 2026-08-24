# SUMMARY — 2026-08-24 — the unguarded completion writer (`bugs_open/375`)

First read-out for this lane. Written to be read aloud.

---

## What we're trying to do

Stop a particular kind of quiet lie. When one of our agents tries to fix something on a site,
a row in the database gets stamped "done". We have a mechanism whose only job is to make that
stamp honest: before a row is marked done, it re-runs the original problem's own test and
refuses the stamp if the problem is still there. It exists because an agent can report success
without having changed anything, and when that happens the defect stays on the site while every
dashboard says it was handled.

The bug is that **the mechanism only guards one of the doors.** There are three pieces of code
that can stamp a row done. One asks the re-check. One never asked. And a third — which we found
during this work — bypasses both by writing the row directly, and is held back by a different
mechanism entirely.

## Where we've come from

The bug was filed on 23 August by a different lane, which found it while working out why a
router had closed a real problem as "done" without fixing it. It filed the bug rather than
fixing it there, because changing what a shared completion path guarantees is the kind of change
that goes through review on its own merits.

That lane then closed, and left a handover note. The note included a measurement it had run on
our behalf: how many agents actually use the unguarded door. Nobody had picked the bug up.

## What we've done

**Re-ran the measurement rather than inheriting it.** Four of our two hundred agents use the
unguarded door, across six places, covering five kinds of problem — and none of those five kinds
has a re-check written for it. So nothing is being skipped today. We also controlled the
measurement, because a zero produced by a typo looks exactly like a real zero.

**Found the part that made this urgent.** Our internal register — the reference document
describing how each mechanism works — carries an explicit warning to whoever eventually writes
one of these re-checks: *"registering this will make one of the router's closing paths refuse
to complete; read the close paths first."* That warning is wrong, and wrong *because of this
bug*: that closing path uses the unguarded door, so registering the re-check would have caused
nothing at all. The next person was being told to brace for one wrong outcome and would have
walked into a different one, silently.

**Built the fix that fact points to.** The unguarded door can now ask the re-check — but only
where a step explicitly opts in, one closing path at a time. That is deliberate: switching it on
everywhere would have made the register's warning come true and broken a live route as a side
effect. Nothing changes anywhere until somebody arms it on purpose, which is also exactly where
the decision is visible to whoever reviews it.

**Added the half that stops a switch nobody flips being worthless.** When the unguarded door
completes a row whose problem type *does* have a re-check registered, it still completes it —
no change in behaviour, no risk — but writes on the row that it skipped a check that existed.
The trap now announces itself somewhere we can query, instead of being invisible.

**Proved the guard is real by breaking it.** Four separate ways: delete the wiring, force the
check to always pass, silence the record, and re-point the test seam. Each one made the right
tests fail. A test that still passes when you remove the thing it is testing proves nothing, and
that is the specific failure this codebase has been bitten by before.

**Corrected the two documents that were lying**, in place and visibly — the test file whose
header read as though registering a re-check protects a problem type, and the register entry
with the false warning.

## Where we are now

The code is committed and will go live on the next fleet roll. It is inert on every live path
even then, by design and by measurement: nothing arms it, and no problem type it can reach has a
re-check. It is submitted to the reviewer council; the verdict was still running at the time of
writing, and the commits carry the "submitted" marker that gets credited automatically once it
lands.

The mechanism is registered so another workstream can find it. A new landmine entry warns the
next person that registering a re-check protects a problem type only on the doors that ask.

**One thing we did not expect to find, and it is the most valuable thing here.** A guard we
accidentally tripped — while getting a test fixture wrong — turned out to be protecting the
*third* door, the one that bypasses both checks. That guard works by keeping a written-down list
plus a test that fails the build if the list and reality disagree. **That is the same problem as
this one, already solved once, three months of institutional memory ago, and nobody had connected
the two.** It is now written up as the concrete shape for the remaining work.

## Where we're going

Three things, in order of who they help.

1. **The remaining piece of protection**: copy that third door's solution to this one, so the
   build fails the moment somebody writes a re-check that this door would ignore — rather than
   us finding out after a row has already been wrongly completed. The two are complements: one
   catches it at authoring time, the one we shipped catches it at run time.
2. **The structural fix**: merge the two stamping paths into one, so the question "which door
   did this go through?" stops existing. Bigger, goes to architecture review, and this change is
   its first half — both paths now share one implementation of the check.
3. **Read the council verdict and act on it.** If it asks for revisions, the code is already on
   the shared branch, which is how review works here — after the fact, by design.

Not ours, noted and left alone: an unrelated population of 42 rows about broken image URLs that
have never been routed to the handler built for them. Different defect, someone else's to file.
