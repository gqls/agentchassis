# SUMMARY 2026-08-17 — the editorial pass exists, and the surprise was its restraint

Fourth in the series (08-12 *why the copy reads wrong* → 08-14 *the house voice ships* →
08-15 *the plan is complete, four decisions pending* → this). Written because the last
thing on the list got built and behaved differently from how we expected it to behave.

## What we're trying to do

Make the copy on our sites read like a person wrote it. The machinery that builds a page
is graded on facts, coverage, structure, links and styling — never on whether the result
is worth reading, whether the most useful thing is at the top, or whether the page talks
about itself instead of to the reader. Something in the estate already noticed those
faults; nothing fixed them. Design faults had a fixer, copy faults had none. The whole
lane is about closing that asymmetry with a second, editorial pass over a page that has
already been written.

## Where we've come from

Three things had to land before an editorial pass was even worth building, and all three
are done. The house voice was rewritten and shipped fleet-wide as **one** database row that
every writer reads, instead of seven drifting copies. Two sites' briefs were corrected at
the source, where they were teaching the very register the owner had rejected. And an
evidence register was opened for the loan-and-mortgage site so its figures have something
to be checked against.

Along the way the lane's method hardened into a habit that has now paid repeatedly:
**assume the mechanism already exists and go and look**, because this estate is far larger
than any one thread's memory of it. The voice carrier had been sitting unused and
unregistered for seventeen days while we designed a replacement for it.

## What we've done

**Stage 2 is built, live, and has passed the test we set it in advance.**

It needed **no new code**. Everything it required already existed — a way to apply an edit
to one section safely, the voice carrier, a mechanism for putting work in front of a human
before it lands, and a page-scoped read we could copy from the auditor. The single missing
piece, letting one editor see a whole page at once rather than one section at a time, is a
database query in configuration. So it went live the moment it was applied: no build, no
release, no waiting.

Its safety is structural, not written down and hoped for. It **cannot** change a live page —
there is no step in it that can write to one, and the migration that installs it refuses to
apply if anyone adds one. Locked copy is excluded by the query that selects the work, not by
an instruction asking it politely. And it is denied the original brief that produced the
copy, because an editor handed that framing just writes the same page again.

**The test.** Since the 12th, six guide links have been missing from the loan-and-mortgage
homepage — kept there deliberately, unrepaired, on the owner's ruling that they were this
thing's proof. The editor read the page, reported that six required guides were missing,
named all six, and said everything else — order, naming, tone — was sound and needed no
change. Its proposed fix adds six lines and touches nothing else.

We also built the grader that decides such things, and built it **before** the run: does the
edit still carry every link, has it lost any styling or structure, has it invented a figure
that was never there, has it quietly shrunk the page. It only reports a pass after being
made to fail on purpose — six kinds of deliberate damage, all six caught.

## Where we are now

The proposal is sitting in the review queue, unapplied, waiting on one word from the owner.
The page is untouched.

The genuine surprise is the restraint. The failure we braced for — and the failure that
created this proof case in the first place — is an editor that improves what nobody asked it
to improve and loses things on the way. Three earlier rewrite rounds did exactly that; one
of them dropped five links while keeping the word count, which is why a mechanical grader
exists at all. This run changed only what was broken. That is one page and one run, and it
is not yet a claim about the mechanism — the next page is deliberately chosen to be harder.

One correction landed this session too, and it is a pleasant one. The webdesign.co.uk
homepage copy the owner objected to — *"a workbench, not a sales pitch"* — was recorded as
blocked and waiting. It had already fixed itself three hours before that note was written:
another thread's unrelated positioning fix regenerated the page, and because the new house
voice had gone fleet-wide the week before, the regenerated copy simply came out without the
phrasing. Nobody aimed at it. That is the first time we have watched the central voice
improve a page unattended, and it is better evidence than a targeted rewrite would have been.

## Where we're going

Apply the proof case, once the owner says so — that is the only step that exercises the
write path from end to end, and until it runs, the last mile is declared rather than
demonstrated.

Then a harder page: one with several components and a component declaring several fields,
where the type checks and the narrow-edit machinery are doing real work instead of standing
by, and where the fault is a matter of register rather than a missing list. That is the run
that tells us whether the restraint is a property of the design or a property of an
unusually legible defect.

Only after that does it get wired to fire on its own. Nothing dispatches it today, on
purpose. And routine operation at volume still waits on the human review queue having a
surface anyone can read — a different thread's work, and the one dependency this lane
cannot close for itself.
