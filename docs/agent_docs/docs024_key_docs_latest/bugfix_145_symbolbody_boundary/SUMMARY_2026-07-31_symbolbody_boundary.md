# SUMMARY — 2026-07-31 — bug 145: the symbol reader that would read anything

Written to be read aloud. First summary in this series; the workstream is complete.

## What we're trying to do

Our diagnosis loop shows an AI a bundle of our own source code and asks it what is broken.
The AI answers, and part of its answer is *"next, show me these bits of code"*. Whatever it
names, we fetch and paste into the next bundle, under a heading that says **"In-scope code"**,
wrapped in the markup that means *this is Go source*. The job was to make sure the fetcher
can only ever hand back what it claims to be handing back.

## Where we've come from

The flaw was found by **our own review council**, not by a human and not by a failure. A seat
was reviewing an unrelated change, saw that the author had closed their own instance of a
problem, and asked the useful question: *what happens to the next caller?* That became bug
145, filed two days ago and untouched since.

The fetcher had no opinion about what it was asked for. Given a filename rather than a
function name, it handed back the whole file, straight off the disk, without checking whether
that file was source code at all. A request for a document, or for a file shaped like a
secrets file, was honoured — and came back labelled as our own code.

## What we've done

**Fixed it, and the fix is small.** There is already a list of every file our analyser
parsed. It was already in memory, already consulted by the same function two lines further
down, and by construction it contains only real source — no documents, no test files, no
vendored copies. The fix consults that list *first* and refuses anything not on it. Four
lines moved, one error message improved, no new setting for anyone to remember.

The best part is that **we did not invent this rule — we put back one we already had.** The
tool this function was originally lifted from does exactly this check, in its calling code.
When the function moved into our platform, the caller was rewritten and the check was left
behind. So the honest description of the bug is *a safety step got lost in a move*, and of
the fix, *put it where a future move cannot lose it.*

**We also corrected the bug report itself, twice, and both corrections changed the work.**

- It said the dangerous path was unreachable in practice. It is not: the main source of these
  requests is **the AI's own reply**, checked by nothing but whitespace-trimming — and the
  bundle we send **literally invites it**, printing *"put the bare file path in next_scope to
  see it whole"*. That matters because the report's top-ranked fix was *delete this feature*,
  which would have broken something we advertise.
- It said the leak was limited to files in our repository. It was not. A file placed
  *outside* the checkout came back. I only found that because I made the test create a real
  file; my first attempt asked for files that did not exist, which passes whether the guard
  works or not.

**The council approved it first time**, with two advisory notes and nothing blocking. One
note was better than my own judgement: I had found a second flaw in the same function and
written it down as a footnote for a reviewer to decide about. The reviewer said a known-shape
flaw found while working on that very function should be a ticket. It is now a ticket.

## Where we are now

**Done and live.** The fix is committed, council-approved, and **verified running on both
production replicas** — not by trusting the version tag, but by grepping the actual binary
for a string the fix added, alongside a positive control (proving the grep works) and a
negative control (proving it is not matching everything). The ticket is closed and moved to
`bugs_closed/`.

Notably, **I did not do the deployment.** Another session had a build in flight, and rolling
mid-flight kills any review that happens to be running. Because our builds are taken from
committed history rather than from anyone's working copy, the owner's own build minutes later
carried the fix with no coordination needed. That is the shared-branch design working exactly
as intended: commit narrowly and promptly, and someone else's deployment ships you.

The work produced more than its own fix: **two new tickets** (one at the council's request,
one found while writing the warning note for this bug), **two new entries in the debugging
guide**, a fleet-wide warning note, and **three entries in our log of wrong calls** — all
three of which are the same mistake wearing different clothes: *I read the code that builds a
thing instead of running the thing.*

## Where we're going

Nothing further is owed on 145 itself. Two things are handed on, both filed, neither started:

- **The second flaw in the same function** — when the code bundle runs out of room it *stops*
  instead of *skipping*, so one oversized piece of code silently discards everything the AI
  asked for after it. The list is sorted, so what gets thrown away is alphabetical, not
  unimportant. First job is to measure how often it actually fires; I deliberately did not
  claim it has.
- **Our own warning-note checker is blind.** Two halves of it cannot agree: one asks its
  question in a format the other is incapable of answering. It has never once succeeded at
  this kind of check across every run it has ever made, and no care in how we write the notes
  can help. That belongs to the team that built it, and it matters more than it sounds:
  that checker is what decides whether the warnings every session reads at start-up are
  trustworthy.

The thing I would most want repeated is the cheap habit that saved this work twice: **after
claiming to have fixed something, re-run the thing.** It cost an hour and caught two wrong
root causes that would otherwise have gone into a handoff stated with confidence.
