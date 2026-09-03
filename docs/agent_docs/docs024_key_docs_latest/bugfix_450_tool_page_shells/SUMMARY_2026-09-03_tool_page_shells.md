# SUMMARY 2026-09-03 — the tool pages that were not tools (bugs_open/450)

*First summary in this lane. Written to be read aloud.*

## What we're trying to do

Stop the platform publishing pages that promise an interactive tool and deliver an article about
one. Not repair the ones that exist — that is a separate job, already in hand elsewhere — but make
the class of defect stop happening, in a way that works for every site the framework builds rather
than for the site where we noticed it.

## Where we've come from

A new site, seotools.co.uk, went live with seven pages at addresses like `/tools/robots-txt-tester/`.
All seven answered, all seven looked finished, all seven carried the right headline, and not one
contained a tool. Every automatic check passed: the pages were marked built, the components were
marked deployed, the links resolved, the URLs returned success. The only thing that knew anything
was wrong was a note in a queue that nobody reads.

The sequence turned out to be ordinary machinery doing exactly what it was told. The planner writes
the site plan and names the tool pages; at that moment the tools do not exist, because they are
invented later by a separate process that visits one site every few hours — and when it arrives it
invents its own names, which on seotools matched none of the seven the planner had used. Other
pages link to the promised addresses. A checker notices a link pointing at an unbuilt page and asks
for it to be built. The builder reads the plan, and the plan says the page is a headline and a text
block, so that is what it writes and deploys. A guard exists to stop precisely this, but it asks
whether the page is "owned" — a question that does not apply — and waves it through.

The blunt version: **a tool page is judged by whether it has a form on it, and nothing was asking
that question.** Sixty-seven such pages across sixteen sites.

## What we've done

Two independent fixes, both written, both reviewed and approved by the platform's own review
council, both committed.

The first is a rule the builders now consult: *a page typed "tool" with no tool on it is not
available for generic building*, enforced at all six places a build can start. The important
property is what it is **not** — it is not a flag someone sets and someone else must remember to
clear. It is a question asked fresh each time, so the moment a real tool arrives the page is
released automatically. That mattered more than it sounds: the obvious version of this fix marks
the page, and when we looked, the field everyone would use for that had never once been changed by
any code in this system, in either direction. A mark nobody clears is a page stuck for ever.

The second stops the planner inventing these pages at all, so the supply dries up rather than being
caught downstream. It reuses the design another session shipped a day earlier for a near-identical
problem, deliberately with its own on/off switch so that if ours misbehaves theirs is untouched.

Along the way we corrected several things that were wrong, including three of our own. Two were
claims we had written down and repeated — one about how a database column behaves, one a count that
turned out to be a floor rather than a total. The third was a comment claiming a test existed when
it did not, on the riskiest part of the change. Each is logged in the fleet's ledger of wrong calls
under our own name, because the tally of what keeps catching us is worth more than any individual
entry.

## Where we are now

**Nothing is live.** Both halves are committed and both are inert: this system only picks up code
changes when a new server image is built and rolled out, and the one running was built before any
of this existed. The accompanying database change is written and rehearsed but deliberately not
applied, because the text it installs tells the planner that a validation is running which will not
be running until the code ships.

The sixty-seven existing pages are untouched by us. Seven of them — the seotools ones — have had
real tools built by the lane doing the instance repairs and are queued to publish.

The single riskiest thing we know about our own work is written down rather than buried: the
plan-side gate's whole justification is a measurement that *nothing reads planned tool pages*. That
is a negative finding, and negative findings go stale by addition. If a future process starts
reading them, our gate would quietly starve it. Turning that assumption into a standing check,
rather than a sentence in a comment, is the single most valuable next piece of work.

## Where we're going

Three things, in order. Wait for the next server roll and then prove the fix at the artefact — by
watching a real piece of queued work be refused with a receipt, not by observing that no new bad
pages appeared, since an absence proves nothing without something having tried. Then apply the
database change, once the code behind it is genuinely running. Then build the standing check for
the assumption above.

Two things we are deliberately not doing: repairing the existing pages, which belongs to the lane
already doing it, and gating the re-render path, which is a different and older problem with its
own bug file.
