# Contact-info fabrication — complete, 2026-08-03

*A new file, not an edit of `SUMMARY_2026-08-02_…_closed.md`. That one was written when
the bug was closed and one judgement call was still open; this one is written because the
call was made and answered, and the lane now has nothing outstanding.*

## What we're trying to do

Stop the website builder from putting things on customers' pages that nobody ever told it.
That started as one specific scandal — a shared "contact details" block that invented a
phone number and opening hours — but the real goal is the general form of it: when the
platform has no information, it should say nothing, not fill the gap with something
plausible.

## Where we've come from

A shared component was serving `+1234567890` and "Monday – Friday, 9am – 6pm" on live
commercial sites, in the same styling as real details, on eight sites at once. The
organising discovery was that the component's own settings file already said "if there is
no phone number, leave the field out" — the template simply ignored its own published
rule, and no layer in between could tell.

Fixing that one component was the easy part. The uncomfortable question the review council
raised, twice and independently, was: nothing anywhere enforces those settings, so what
stops the next component doing the same thing? That became a written proposal with four
costed options, and the owner chose two of them: a daily check that reads the live
component library, and a guard that refuses a fabricating template at the moment it is
written. Both shipped. The proposal's most ambitious option — making the page-drawing code
enforce the settings for every component at once — was deliberately not taken, because
measurement showed nine fields in ten do not declare a setting at all, so it would have
been inert where it mattered and risky where it did not.

That left one item: 68 fields across 20 components which declare "leave me out if absent"
and then draw an empty space instead. Milder than inventing a phone number — an empty gap
asserts nothing false — but still wrong, and nobody had costed the repair.

## What we've done

The owner asked for them, and they are gated. Twenty components changed in one migration,
applied and verified, with a rollback recipe backed by stored before-images.

The work turned up three things worth more than the change itself.

**Our own description of the problem was misleading.** "These fields have no guard around
them" invites the repair "put a guard around each one" — and for 62 of the 68 that is
wrong. They are the second half of a pair: a spec table row is switched on by the row's
name, and the value beside it in the same row is the unguarded one. Guard the value where
it sits inside a table cell and the page renders exactly as before — an empty cell — while
the warning light goes off permanently. Guard the cell itself and a four-column comparison
row comes out with three cells. Both of those pass the automated check, because the check
can see that a guard exists but not what it wraps. Four different treatments were needed,
decided component by component.

**Two of the 68 were not harmless gaps at all.** One was an article image, which with
nothing to show emitted an image tag pointing at nothing — a broken-image icon and a
wasted request. Two others were button labels, so the page carried buttons with no words
on them: invisible and unclickable. Those belong to a defect class we already track
separately, and the write-up that called the whole group mild was too generous.

**We nearly reported the damage as twenty-five times worse than it is.** Counting stored
records, 75 fields were missing across live pages, 47 of them a hero subheadline. Counting
the actual pages, three carry the empty gap today. The other 46 hero sections have real
text in the saved page, written earlier, over a record that has since been emptied. Both
numbers are now written down, each labelled as what it is: one is risk, one is damage.

Proof was taken from the live system after the change, not from the file we wrote: all
twenty templates were fetched back out of the database and drawn twice each — once with
the information present, once with it missing — through the platform's real drawing code
rather than a copy of it. Twenty out of twenty. The half that does the work is the
positive one: a guard that is too aggressive sails through a test that only checks things
disappear.

Separately, an overnight notice from another thread claiming an earlier fix was not in the
running software was investigated and refuted — their search looked for a sentence that
only exists in a source-code comment, and comments do not survive compilation, so it would
have come back empty against any version. The fix is live on both machines.

## Where we are now

The daily check reports the component library completely clean: no template invents a
business fact, and every field that declares "leave me out" is now actually left out. The
number it reported yesterday was 68. Nothing in this thread is half-finished, nothing is
awaiting a verdict, and no page needs repairing by hand — the twenty templates correct
themselves the next time each page is properly rebuilt.

## Where we're going

One small decision is left, and it is genuinely small. The daily check reports these gaps
but does not treat them as a failure, because when it was written 68 already existed and a
warning nobody can clear is a warning everybody ignores. That reasoning has expired now
the count is zero. Making it a hard failure would stop the next one ever being introduced;
the cost is that a new component from another team could turn the check red until someone
adds one line to it. It was not done unasked, because it changes what an existing check
does to other people's work.

Beyond that, the ambitious option stays on the shelf with the two conditions that would
bring it back: if component authors start declaring these settings much more often, or if
a third invented fact ever slips past the daily check — which would mean the check is at
the wrong layer.
