# SUMMARY 2026-08-06 — the two bugs that stopped a decomposed page being rebuilt, and how we proved they are gone

Written to be read aloud. Both bugs are fixed, live and verified end to end;
both files stay in `bugs_open/` by your direction of today.

## What we were trying to do

You asked, on the 5th, for loancalculator's copy to be rerun through the
framework in a particular voice rather than hand-edited. That turned out to be
impossible, and the reason was a bug rather than a missing feature. So the job
became: find out why the framework could not rebuild its own page, fix it
properly, and prove the fix on the real site.

## Where we had come from

Loancalculator's pages are "decomposed": each section is stored under a
positional name — `prose-0`, `prose-1`, `tool-2` — chosen deliberately so that
if a paragraph ever vanishes, the warning can say *which* paragraph. That naming
was the right decision and it is what the platform choked on.

Two separate defects sat on the path, and the second was hidden behind the first.

The build planner asked the component library "what is `prose-0`?" The library
answers by name and by function, and `prose-0` is neither — it is a slot on a
page, not a kind of component. So every section came back unresolvable, the
build refused with *"no sections ready to build"*, and — worse — it filed work
items asking the fleet to *build* components called `prose-0` and `prose-1`,
which already existed and were already attached to that very page. A full run
on that site would have filed 114 of those.

The same blindness had been found and fixed once before, four months of bug
numbers ago, on the *re-render* path. The fix never reached the *build* path,
and the commit that made it had actually edited the build path's file while
leaving its lookup alone. That is a pattern this estate keeps recording: one
call site of a shared judgement gets the rigorous fix, its sibling stays
heuristic.

## What we did

**First bug (204).** The planner now asks the page instead of the library: each
stored section already records exactly which component it is, and that fact does
not depend on how the slot was named. We reused the helper the earlier fix had
built rather than writing a third one. The council approved it, and it went live.

**Then the thing worth telling you about.** Making that work exposed a second
defect, which had been filed but never fixed — and our own fix armed it. When a
section resolved, the save step wrote it back under the *component's* name
rather than its own slot name. So `prose-0` silently became `ported-prose`. And
because the guard that protects human-locked sections matches on that same name,
the rename made the guard look for a section that no longer existed by that name
— so it failed to protect the locked row and inserted a duplicate beside it. On
the 3rd of August that put the loan calculator on the page twice.

**Second bug (189).** The slot name now travels with the section it names, as
its own piece of information, and the save trusts it exactly as given. Absence
of it means the old behaviour, unchanged, so nothing else on the platform had to
move. Approved, registered as a shared contract others must follow, and live.

## Where we are now

Both fixed, both live on the current build, and both *induced* rather than
argued — we made each one happen on the real site and watched the right thing
occur:

- The locked calculator section: four rows where the bug produced five, the
  positional names intact, and the locked row untouched down to its original
  timestamp.
- The guide page rebuilt through the framework in the requested voice: the prose
  genuinely changed, it opens the way the brief asked ("If you've ever looked at
  your monthly…"), the facts and links survived, and **zero** junk work items
  were filed.
- The live page checked directly: the new prose is there and the old prose is
  gone.

**One thing is worth your scepticism, because it nearly caught me.** The
official test for the locked-section bug was "the page should have four sections,
not five". I ran it, got four, and nearly wrote it down as proof. But a run that
did *nothing at all* also produces four. The test could not tell success from
inaction. What actually proved it was that the rewritten rows are *new records* —
the save really did replace them — plus the step's own report that it re-rendered
four sections and carried none. That near-miss is logged, because a count tells
you the damage was avoided and never that the work happened.

I also got two things wrong and had them caught. I told the review board that an
earlier decision had been "council-reviewed" when in fact only the submission
existed, never a verdict — one query would have shown me. And I claimed one part
of the system used the planner when two do. Both are logged where the next
session will trip over them.

## Where we are going

The framework rerun you asked for on the 5th is now unblocked, and the mechanism
is proven on that site's own pages — so that is the natural next step.

Two loose ends, both deliberately left rather than forgotten. A third piece of
the system still writes sections without the new slot-name field; it is safe
today only because of what it happens to do, not because anything enforces it.
And the identity lookup now exists in two places with the same logic written
twice — if a third ever needs it, it should become one shared piece first. Both
are recorded where whoever touches them next will see them.
