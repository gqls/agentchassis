# SUMMARY — 2026-08-20. The walls came down, and the checker we owed turned out to be mostly built.

*Written to be read aloud. Current state only; the chronology is in `README_where_we_are.md` and the
evidence is in `NOTES_native_rebuild_of_ported_tools.md`, both in
`docs/agent_docs/docs024_key_docs_latest/webdesign_tool_rebuilds/`.*

## What we are trying to do

webdesign.co.uk has 63 small interactive tools. They were carried over from an older site as fixed
lumps of hand-written code sitting inside a framework page, so the framework cannot see inside them,
cannot check them, and cannot improve them. We are replacing all 63 with tools the framework itself
generates, at the same web addresses, so that from then on every improvement we make to the platform
reaches them automatically. The owner has ruled that all 63 go through the framework — there is no
"leave the good ones alone" option, because a tool the framework cannot see is a tool nobody is
maintaining.

## Where we have come from

We started with a pilot and a lot of caution, because replacing a live page is a thing you can get
visibly wrong. The recipe settled into six steps and is now boring: read what the old tool actually
does, write a brief describing what it *should* do, let the generator build it, check the result three
ways, retire the old copy, and confirm the real page on the internet has changed. Along the way we hit
three separate platform walls, filed them all, and other sessions fixed two of them. We also learned
the most useful thing about this project: the old tools are not merely old. Two thirds of the ones we
have opened were broken in ways no visitor could report, because they look like they are working.

## What we have done

Sixteen tools are rebuilt. Fourteen are confirmed live on the public site, checked by fetching the real
pages with caches defeated and proving that identifiers only the old versions contained are gone. Two
more were built and swapped over this morning and are waiting for their pages to be re-published.

We have also changed the framework itself rather than fixing tools one at a time. Six recurring faults
— buttons that claim success they never verified, pop-up alert boxes, controls wired to nothing,
unvalidated numbers, errors that destroy the user's work, and transformers with no before-and-after
readout — are now written into the generator's own contract. Every tool built since has come out with
those properties without any brief mentioning them; that has now happened six times running.

And we have proven another session's fix. The A/B test significance calculator had defeated three
rebuild attempts on three separate days, because the platform treated our site's copy of a shared tool
as a duplicate of the shared original. Another session built the fix yesterday; we rebuilt that tool
last night, first time, and confirmed at the database that the shared original was not touched in the
process. Their fix was half of a pair the owner wanted proven together, so we have written the result
into their notes.

## Where we are now

Forty-seven ported tools remain, and nothing is blocking any of them. Exactly one is set aside — the
meme generator — and only because it is one of the five rich hand-built applications the owner asked to
be done last, one at a time, with a person looking at the result.

The one piece of the owner's directive still owed was an automatic checker: something that finds these
faults in the tools we have not yet reached. We went to build it and found two things that make it much
smaller. First, we already own an automatic tool reviewer, and it had already reported one of the exact
faults we found by hand this week — four days earlier, in its own words. Second, and more important:
one of the six rules cannot be checked this way at all. The natural test for "does this tool validate
its numbers" looks for the standard guard against a bad number, and the tool we built last night does
not contain that guard because it uses a stricter check earlier instead. A checker written the obvious
way would have accused the best-validated tool on the site. So what is worth building has shrunk to two
rules that cannot be wrong, inside a check that already exists and already runs.

We also nearly recorded something false. Measuring whether these findings would even be acted on, we
found two hundred tool-improvement jobs in a dead state, which reads as "the repair path never works".
The rows themselves said otherwise: twenty of them had been filed under one shared name instead of one
per tool, so the platform's own "this failed twice already" rule condemned them before anything ran —
and a migration fixing exactly that was applied the same afternoon, after which the jobs do get done.
The graveyard is a scar, not the current state.

## Where we are going

Straight down the list, smallest first. The recipe is routine and the queue, not the work, sets the
pace. Two things will make it faster and both are already built by other sessions: a "replace the tool
in place" capability that removes the three manual database steps and the small race we run on every
rebuild, and which they have asked us to be the ones to test; and the checker above, once someone
writes its two rules.

The one thing we still have no answer for is the fault class we keep finding: a tool that asserts
something untrue about itself. Four of the sixteen have done it — prose presented as SQL, a comment
claiming a colour was generated when it was a fixed value, a control wired to nothing, and a
statistical verdict computed from a number that was not a number. The generator's new rules govern how
a tool behaves, not whether its account of itself is true, and a text search cannot tell the
difference. For now the defence is human: when writing each brief, read what the tool says about
itself and check the code does it. That has caught all four.
