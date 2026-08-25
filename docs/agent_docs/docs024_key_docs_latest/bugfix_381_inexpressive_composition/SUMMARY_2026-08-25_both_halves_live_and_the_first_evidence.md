# SUMMARY 2026-08-25 — `bugs_open/381`: both halves live, the writer arm proven, the planner arm untested

## What we're trying to do

Stop pages promising things they cannot deliver. The complaint that started it was concrete: a page
headed *"What your shed needs, month by month"* that contained no months, and a 300-word paragraph
that should have been a list. The underlying problem was that the system had no way of knowing the
difference — the part that chooses a page's components could not see what any component was capable
of, and the part that writes the words had been told, in effect, to write paragraphs and nothing else.

## Where we've come from

The bug was filed on 24 August by the lane that had just finished building `garden-tools.uk`, and
left unowned. Picking it up, the first job was to check it was real and not confined to one site. It
was fleet-wide: **44% of 741 recently-built pages contained no list, no table and no bold text
anywhere in their content**, and **94% of all section placements used a component that physically
cannot produce a list.**

The first explanation was wrong, and the data corrected it. It looked as though the writer was
*forbidden* from producing lists by a rule about field types. But another component, declared exactly
the same way under exactly the same rule, was producing lists **76%** of the time against **7%** for
the one on the owner's page. The difference was not the rule. It was that one component carried a
one-line note telling the writer what to do, and the other carried **no note at all**.

## What we've done

Two halves, eight configuration changes, two independent review rounds, everything approved.

**The planner now sees capability.** A function works out from each component's own template what it
can produce — a list, a table, a repeating set, or only paragraphs — and every one of the three
planners now shows that in its menu, with a rule saying a page promising a calendar or a checklist
needs a section that can render one. This is *derived*, never typed in by hand: the column someone
once created for this purpose had drifted to being wrong on twelve components and read by nothing.

**The writer is now told to use structure**, in the four text blocks that carry most of the prose on
the estate — modelled on the note that already worked on the component doing 76%.

**And the three missing components were built**, because the first half was explicitly not enough:
the library's forty-four structural components turned out to be all special-purpose — directories,
calculators, quizzes, one pricing table — so a planner that knew what everything could express still
had nothing general to choose. There is now a **checklist**, a **period calendar** (generic over
months, quarters or seasons) and a **comparison table** that stacks into readable blocks on a phone.
A fourth was dropped after reading the nearest existing component and finding it already did the job.

## Where we are now

**The writer half is live and proven.** Measured at the writer's own output, where the change
applies: of the writes that received the new instruction, **72% now produce a list against a 10%
baseline, and 100% produce a subheading.** Where that content reaches a page, the structure reaches
it too.

**The planner half is live and completely untested.** The component-choosing step has not run once
since the change — new-site builds are the only thing that trigger it, and none has happened. So the
three new components have **never been placed on a page**, and the central claim of the fix is built
but unproven.

Four claims made during the work turned out to be wrong, and none was caught by the person making
them. The most instructive: a warning was filed fleet-wide about a problem the platform already
solves — the measurement behind it was correct and the explanation was not, which one command would
have shown. It was withdrawn the same day, with the correction left where the next person will find
it rather than deleted.

## Where we're going

**One thing is left, and it needs a decision rather than more work: a single new-site build.** That
is the only event that runs the component-choosing step, and therefore the only way to find out
whether a planner that can see capability actually composes a page that keeps its promise. The
subject matters more than the domain — something genuinely structured, a buying guide or a how-to,
will exercise all three components; a two-page brochure will exercise none.

Two things stay open regardless. The comparison table can still be filled with an invented price,
because roughly two-thirds of our sites have nothing to check figures against — that is another
thread's work and the component says so in its own header rather than implying it is solved. And the
owner's third complaint from the original review — too many card sections stacking up on a phone — is
a decision about page composition that has never been filed as a bug at all.
