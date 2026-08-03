# SUMMARY — 2026-08-03 — from one wrong path to a queue that can change its mind

Third summary of this lane, and a new file rather than an edit: the last one ended with "live,
approved, closed, and three things left to decide". All three were decided, one of them was
built and shipped, and the lane found a second defect larger than the one it started on.

Five parts: what we're trying to do · where we've come from · what we've done · where we are
now · where we're going.

---

## What we're trying to do

Two things, and the second grew out of the first.

**Make the platform agree with itself about where a generated image lives.** The code that
commits a file into a site's repository and the code that points a page at it must produce the
same path. When they disagree you get a file nobody references, or a reference to a file that
isn't there, and each half looks correct on its own.

**And make the work queue able to notice when a job it wrote down stopped being true.** A work
item is a claim, recorded once. Nothing ever re-checked it.

## Where we've come from

The path defect had been met by three separate lanes and contained locally by each. One
declared a map of the two filenames that can't be derived and left a deliberate tripwire naming
its own remedy. Another nearly shipped a check that would have reported a broken favicon and
social card **on every site we run**, caught itself, and patched its own call site. A council
seat then objected that the residual had been deferred as "its own item" and never filed —
which is the only reason `bugs_open/168` existed to be picked up.

## What we've done

**Corrected the bug's own diagnosis before acting on it.** Its stated mechanism was too broad,
and one of its four suggested fixes would have *created* the drift it claimed to remove. Struck
through on the ticket with the reason, rather than quietly ignored.

**Fixed the class.** One derivation, resolved through by the writer *and* all six readers. They
can no longer drift, because there is nothing to drift from — where before they were held in
step by a comment asserting they matched.

**Took it through the council three times, and was wrong once.** Round one found a genuine
defect in the fix and three things done-but-not-shown. **Round two gated it at high severity and
was right**: the change made it possible to overwrite a site's live social card, and I had twice
denied that with measurements attached. The proof was eleven work items sitting in the queue,
false for three days, still dispatchable.

**Turned that into the second piece of work.** Both of my measurements had been sound and
neither could answer the question — one was about items created *from now on*, the other about
*readers* when the danger was in a *writer*. The population that answers it is the queue that
already exists. That became `RFC_010`, the owner ruled on it twice, and the first half is now
built and live: a check can say what it has positively observed to be fixed, and the runner
closes those items.

**Repaired the eleven**, with the reason recorded pointing at the still-real defect hiding
behind them, so retiring a false finding didn't retire a true one.

**Fixed the thing that let two lanes collide on a document number** — the register that was
supposed to prevent it had gone unmaintained for eight papers, including both of mine.

## Where we are now

Everything is committed, live, and verified on the current build. The original bug is closed.
The retraction mechanism is approved, deployed, and survives rolls.

**And it has retracted nothing at all**, because the only check wired to use it is enabled
nowhere. That is the expected first state of an opt-in mechanism and not a fault — but it is
the sentence a future reader most needs, because "zero" here means *nobody adopted it*, not
*it's broken*, and acting on the wrong reading would mean rewriting working code. It's recorded
that way in the register rather than left to be discovered.

Five wrong calls of mine are written up fleet-wide, each with the cheap check that would have
caught it. The most useful is the one that nearly shipped: **the council's most valuable finding
was the one I argued against twice, and I only caught it by running a query intended to prove
the reviewers wrong.**

Two smaller things worth carrying: both of my changes went live on *other sessions'* builds,
hours apart, before I'd finished reviewing them — that's shared HEAD working as designed, and it
means anything committed here must be safe the moment it lands. And the absence of an exposure
window on the round-two finding was **luck, not planning**; had the first commit rolled a few
hours earlier it would have been a live incident rather than a review comment.

## Where we're going

Three things, ranked, all written up in the handoff:

1. **Adoption.** The retraction mechanism is scaffolding until a check uses it. Find checks that
   already compute a positive observation and throw it away, and have them say so — one or two
   at a time, each independently reviewable, measuring a real retraction on a real site before
   anyone claims it works.
2. **The remaining bypass** (`bugs_open/179`). A caller can still publish a file to a path no
   reader can predict. Measured empty — and this lane has already been wrong once in exactly
   that way, so measured-empty is not closed.
3. **The deduplication half of the second ruling.** Blocked behind 87 duplicate rows and a
   change where getting the order backwards breaks every work-item insert in the estate. Its own
   project, with its own review, and someone watching the roll.
