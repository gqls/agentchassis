# SUMMARY — 2026-08-16: the first imported tool has been rebuilt by the framework, and it is live

*Written for the owner to read aloud. First summary in this lane — the milestone is that the
replacement recipe now works end to end, on a real tool, at the real address.*

## What we are trying to do

webdesign.co.uk has 63 tools that were **imported** rather than built. They were dropped onto the
site as finished lumps of HTML, all sharing one wrapper, and because of that the framework has never
been able to see them properly — it cannot check them, improve them, or rebuild them the way it does
everything else it made itself. The owner asked us to replace them with tools the framework builds
and owns, at the same web addresses, so nothing a visitor has bookmarked breaks.

## Where we have come from

For most of this lane's life the replacement was not actually possible, and we did not know it. The
part of the system that saves a newly built tool could only ever **create** a page — it had no way to
say "put this on the page that already exists". Since the whole point is to rebuild a tool at its own
address, every single replacement would have hit that wall. Our first attempt did, and we initially
blamed a known flaky handshake between agents; that was wrong, and reading the error log rather than
the orchestration record is what corrected it.

We also learned the hard way what a bad rebuild looks like. The A/B test calculator was rebuilt
earlier and the result was a 13 KB shell: every heading and button label had been left as a blank for
something else to fill in, nothing ever filled them, and the live page served 47 raw template
markers to visitors. It passed the automated checks of the day, because those checks counted styling
and found plenty. We put the original back and it is queued for a proper rebuild.

## What we have done

The missing "attach to the existing page" path was diagnosed, built as an opt-in switch that is off
by default, reviewed and approved by the council, and shipped. It went live in the fleet build this
morning, and the switch was turned on for the tool generator alone this afternoon.

Then we ran the pilot: the aspect-ratio calculator. The generator built it and attached it to the
existing page in 54 seconds. We graded what it had built **before** switching anything off — the
lesson from the A/B test shell — and it is genuinely good: it reduces a width and height to a proper
simplified ratio, it works backwards from a target ratio to a missing dimension, the 16:9 / 4:3 /
1:1 / 21:9 shortcuts are there, and all the wording is baked into the tool rather than left blank.
Only then did we switch the imported version off, and we proved its content was untouched by the
switch, so it can be put back in one step.

## Where we are now

**The rebuilt tool is live at https://webdesign.co.uk/tools/aspect-ratio/index.html.** The page was
re-published at 16:47 and we checked the actual served bytes: the imported wrapper is gone, there are
no raw template markers, the tool's own controls and headings are all present, and there is real
readable text on the page rather than an empty shell. The old version sits alongside it in the
database, switched off and intact.

That makes the recipe proven rather than theoretical, and the order of its steps is the valuable
part: build, then judge the result in the database, then switch the old one off **before** the
automatic re-publish runs, then check the live page. Doing it in that order means visitors never see
a page with two tools on it, and it costs one re-publish instead of two.

Three corrections came out of it that matter more than they sound. The instruction to record an
"archive row" as each tool's undo handle cannot be followed — no such row is created when you switch
a tool off — so the undo handle is the old tool's own record, which is better. The address we had
been checking live pages at was a 404, and a missing page passes every "is it clean?" check
perfectly. And the generator's "does this already exist?" test ignores whether the existing copy has
been switched off, so the A/B test tool would have silently blocked its own replacement and the run
would have reported success having built nothing.

## Where we are going

The owner sees this page first — that was the agreed gate, and it is the only step of the recipe a
machine cannot do. After that, the remaining 62 imported tools go through the same route one at a
time, simplest first, with the A/B test calculator second in line. The owner has since ruled that the
rich hand-built applications — the mind-map studio, meme studio, logic architect, the mini-CMS, the
pasteboard — are in scope too, accepting that those are reimplementations rather than copies; the
recommendation is to take them last and one at a time, once the simple ones have proved the process.

One risk sits outside this lane and is worth naming: the dispatcher that runs every job on the
estate is affected by a newly filed bug (`bugs_open/289`) that makes its stored state double on every
lap. A sister agent has already died outright from it. Nothing has stalled here yet, but a batch of
62 jobs runs through that dispatcher.
