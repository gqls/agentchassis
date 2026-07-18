# The problems we faced this thread — a read-out

*A plain-language account, written to be read aloud. This is the "what went
wrong and what it taught us" story of the "diagnosis fixloop 3" thread
(2026-07-16 to 07-18). The triumphant version lives in
`SUMMARY_the_immune_system_2026-07-18.md`; this is the honest catalogue of the
friction, including my own mistakes.*

---

## The one-sentence shape of it

Most of our problems did not come from the tool being wrong. They came from the
world moving underneath it — other people's work landing while ours was in
flight — and from a handful of genuine bugs the tool uncovered, including bugs in
itself. A smaller share came from mistakes I made and had to walk back.

## Problem one: the ground kept moving

This was the dominant problem, and it appeared on the very first turn. We set out
to point the finished diagnosis tool at its first real bug — a case that had been
carefully written up as "landing an image silently blanks a page." Before we did
anything, we checked the live system against that write-up. The case had already
changed. A guard had shipped overnight, the broken pages were being repaired by
another thread, and the headline we were about to chase was half-solved. The
lesson was immediate: a filed bug report is a photograph, and the thing it
photographs keeps living.

That pattern then repeated, four times over three days, and it is worth being
honest that it cost real money each time. We fired a diagnosis at three pages
that looked broken; another session had queued a repair for those exact pages
five minutes before we dispatched, and it fixed them while our diagnosis was
mid-flight — so the tool correctly concluded the bug was no longer there, and the
run was spent proving something we could have known by looking. We aimed at a
last remaining broken page; it was repaired between our checking it and our
dispatching at it. We built a whole diagnosis around a symptom whose second half
had quietly become false because a fix had shipped that morning.

The tool behaved correctly every single time — it refused to confirm a bug that
was no longer real. The waste was upstream of the tool, in us dispatching at a
premise that had already shifted. This is now written into the platform's shared
rules: before you point the tool at anything, check not just the code but the
live system and the work queue, because someone else may already be fixing what
you are about to diagnose.

## Problem two: the tool found real bugs — and one of them by failing

Two genuine, platform-wide bugs surfaced. The first was the reason behind a whole
family of "pages mysteriously went blank" incidents: when the language model is
asked to write more than its token limit allows, the reply comes back cut off
mid-sentence, but the code that receives it never checks for that — so a truncated
answer is treated as a complete success and saved as if nothing were wrong. The
second was subtler and, tellingly, the tool found it by failing: a piece of
configuration meant to give one step a larger token budget had never taken
effect, because of a precedence rule in the code that the platform's own
documentation described backwards. We only discovered this because a diagnosis run
came out truncated and we chased why.

Both are now handed to fixing threads with full evidence. But the honest note is
that finding the second bug meant the tool's own diagnosis had been quietly
running under-budget for its entire existence — a problem we would not have seen
if we had not been looking closely.

## Problem three: moving to a better model had traps

Partway through we upgraded the models the tool uses to a newer, more capable
generation. This should have been a one-line change and was not. The new model
thinks before it answers by default, and that thinking is drawn from the same
budget as its reply — so on the first run it spent its entire budget thinking and
produced no answer at all, and the run failed outright. Fixing that meant raising
budgets everywhere the model is used, and a later run proved the point: two of the
reviewers wrote longer, more thorough assessments than the old budget would have
allowed, which means without the fix their verdicts would have been silently cut
off — the very bug we had just been diagnosing, reappearing inside the reviewers
of its own fix.

## Problem four: the tool's honesty kept creating friction — and kept being right

Several times the tool refused to give us the clean answer we wanted, and each
time the refusal turned out to be correct. It declined to confirm a diagnosis on
code evidence alone, insisting on seeing the problem actually happen — which was
right, and forced us to build a way for it to fetch that evidence. Its newest
reviewer, whose job is to remember the platform's past mistakes, blocked an
otherwise-approved fix by asking a question nobody could answer at the time — does
this same flaw exist in another place? It did, in a second file, and the reviewer
was right to hold. Every one of these was friction in the moment and correct in
hindsight. The tool is trustworthy precisely because it refuses, and living with
that meant repeatedly building the thing that would let it say yes honestly rather
than talking it into saying yes.

## Problem five: the plumbing has blind spots

Underneath the tool, the infrastructure showed its own gaps. A diagnosis run
launched overnight simply vanished — the child process that was meant to do the
work was never created, and the parent sat frozen for nearly fourteen hours. When
we looked, we found dozens of these frozen runs across the platform, some over
fifty days old, because the automatic cleaner only watches for one kind of stall
and is blind to this one. Separately, a tool meant to keep two copies of the
reviewer council identical turned out to check only whether the roster of
reviewers matched by name, never whether their settings matched — so when we
changed the model, it cheerfully reported "already in sync" while leaving half the
council on the old model. We fixed that one here.

## Problem six: my own mistakes, owned plainly

Not all the friction was the world's fault. I fired a trigger with the wrong
setting and launched a run against the wrong target. I wrote a database statement
with a bracket out of place and had to correct it. More importantly, I twice
overstated what I had found — I claimed the tool had been truncating its own work
for its whole life when it had come close but never actually had, and I named the
wrong component as the culprit for a truncation before checking. I caught and
corrected both, but they are the reason I now try to verify a claim against the
live system before saying it out loud rather than after. And on several commits I
had to actively guard against sweeping other threads' half-finished work into my
own, because the shared workspace makes that the default accident.

## What held the whole time

Through all of it, nothing was lost and nothing shipped that should not have. The
tool never merged its own work; every fix still waits for a human. Every wrong
premise was caught, usually by the tool itself. Every one of my overstatements was
corrected in the same session it was made. The recurring lesson, across every one
of these problems, is the same: on a platform where many hands are working at
once, the hardest part is not being right — it is staying right while the ground
moves, and checking the live world before trusting the map.
