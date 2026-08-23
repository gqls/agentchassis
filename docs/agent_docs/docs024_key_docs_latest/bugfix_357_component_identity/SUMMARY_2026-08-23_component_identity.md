# SUMMARY — 2026-08-23, component identity (`bugs_open/357`)

## What we are trying to do

Stop the platform lying about what it has stored. A page section is kept as a row
that says which component made it, alongside the actual markup it serves. On
twenty-two live pages those two things disagree completely: the row says "I am the
shared hero banner" and the markup is an entire interactive calculator. Nothing
errors, because nothing compares them. Everything downstream then reasons about a
banner that is not there — a checker files "this page is missing its headline"
about a page serving its own headline, and no repair can safely touch the row,
because the only repair we have would regenerate the banner over the tool.

The owner ruled in favour of the durable answer rather than a patch: **record
which component actually produced a section's bytes, at the moment it is
produced**, and then fix the mislabelled pages once that record exists.

## Where we have come from

The previous session built that record and got it approved. It shipped on
2026-08-22 at 15:10, and the review council approved it at 18:02 — nine minutes
after the lane stopped writing, so nobody had read the verdict.

Two earlier attempts at the fix itself had been stopped by the council, both
rightly. The first was unsafe: correcting a row's label makes the next rebuild
fail to recognise it, re-add the tool at the end, and keep the newly generated
banner too, so a "tidy-up" would have changed what four live sites serve. The
second was safe and useless — it recorded the problem and stopped it happening to
nobody. That the safe plan and the effective plan could not be the same plan was
itself the finding, and it is what sent the question to an architecture ruling.

## What we have done today

**First, we asked whether yesterday's approved, rolled work had actually done
anything.** It had not. In the day since it went live, 820 new sections were
written and **not one** carried the record. The proof that this is a mechanism and
not a hunch: of 546 live payloads, 546 carried a control field and 0 carried ours,
so the copying step definitely runs and definitely leaves ours out. Between the
part that produces the record and the part that stores it, a step rebuilds the
parcel from a hand-written list of contents, and ours was not on the list.

**This had happened before, and the previous fix is why it happened again.** The
same step lost a different field a while back. It was fixed by adding that one
field and writing a test that checks that one field — a test which passed happily
throughout, because it only knows about its own field. Three fields were being
dropped in total and nobody knew. So we did not add our field to the list. We made
the list a contract: everything is either carried or explicitly dropped with a
written reason, and the build now fails if someone adds a field to neither, or if
the storing end reads something the carrying end was never told to carry.

**Connecting it up naively would have made things worse, not better.** There is a
protective mechanism that puts an interactive tool back when a rebuild would have
replaced it with plain text. It swapped the content but left the record of where
that content came from untouched — harmless while the record was always empty, and
the moment we fixed the plumbing it would have confidently recorded a falsehood.
A missing fact becomes a false one exactly when you repair the pipe, and a false
one is harder to live with, because other machinery trusts it.

**Then we stopped the mislabelling at its source.** The insight that makes a safe
plan and an effective plan the same plan is that a section carries two different
identities: the slot it occupies on its page, and the component that made it. All
of this bug's damage is on the second; all of the danger is on the first. So the
new code corrects what the row says it IS and never touches what the row is
CALLED. A fragment nothing can identify is now attached to a component that
provably reproduces it — proved by rendering it and comparing byte for byte — or
left honestly unlabelled. Never labelled by position. It ships switched **off**,
so nothing changes until we deliberately turn it on.

**And we wrote the repair for the twenty-two existing pages, and left it
unapplied**, as the ruling requires: it may only run once the record is real and
readable on a live page.

## Where we are now

Three rounds of work committed today. The first is approved by the council; the
second is in review as this is written; the third is written and deliberately
held. **Nothing has changed on any live site, and nothing will until the next
release, at which point the fix must be checked at the artefact rather than
assumed** — the whole lesson of today is that approved and rolled does not mean
working.

The bug itself is still live and still producing new bad rows at roughly a dozen
a day; twelve of the twenty-two were created today.

One useful thing we learned about the twenty-two: six of them are pages a human
has claimed, which is why they have been stable since June while the others are
rewritten constantly. They are still mislabelled, and whether to repair someone's
claimed page is a decision for the owner, so the migration now names them out loud
instead of quietly skipping them.

**Three mistakes of our own, all written down with the checks that catch them.**
The one that matters: a test was written, watched to pass, and then the code was
deliberately broken to confirm the test would notice — and it did not. Two
different failures produced the same visible result, so the test could not tell
them apart. It was green, sensible-looking and worthless, and it was about the
very mechanism this lane exists to stop from answering when it does not know.

## Where we are going

1. **The next release, and then proof at the artefact.** The named failing result
   is written down in advance: if new sections still carry no record, a producer is
   still disconnected; and within a day the re-mislabelled rows must still show no
   record, because if one appears with the *wrong* record, the protection above has
   failed and we stop there.
2. **Turn the fix on**, once the record is proven — a separate, reversible
   decision, which is why it ships off.
3. **Then repair the twenty-two**, re-counting on the day rather than trusting the
   number, and taking the owner's decision on the six claimed pages.
4. Still open, and named so it is not mistaken for covered: five other pieces of
   code write these rows and none of them is watched yet.
