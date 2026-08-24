# SUMMARY 2026-08-24 — `bugfix_342_absent_required`: the lane is closed

## What we were trying to do

Stop pages losing content silently. A page section is built by filling a template with text: a
headline, a body, a call to action. If the writer never produces one of those, Go's template
engine does not complain — it renders that spot as an empty string. Page assembly then discards
any section that looks visually empty. So the content does not arrive broken and get noticed; it
does not arrive at all, and nothing anywhere says so. That is the mechanism behind an earlier
fleet-wide incident where article bodies went blank across the estate.

When this lane picked the bug up, the check that catches it ran at **two of fifteen** places where
the platform renders a component.

## Where we came from

A previous lane had already done the loudest part and had it approved: the render seam itself now
reports which schema-required fields came out empty, and the two routes that write straight onto a
live page also file a note on the work queue. What that lane's own bug file said it stayed open
for was blunt — *"no refusal was added anywhere"*. Everything built so far noticed and reported.
Nothing declined.

That file also carried three things that turned out not to be true, which is most of why this lane
took four days rather than one.

## What we did

**Built the refusal.** The two page-editing routes now decline to save a section whose required
fields rendered empty, leaving the live page exactly as it was, with the work item still filed
first so refusing never costs the record. It sits at the single point where every edit is saved,
so future edit types inherit it. Switched off by default and turned on per agent by a migration
held back until the code had actually deployed — this estate's rule for granting new authority.
The site-chrome path got the same protection through the same decision code, deliberately left
switched off because nothing can trigger it yet, with the signal that says otherwise written down.

**Corrected three claims that were wrong.** The bug file said one half of the earlier work was
still waiting to ship; it had shipped. It warned that ~75 components with no content contract
would be "the hard part"; 95 of the 100 such components are self-contained tools that need no
contract, and the remaining five turned out to need nothing either — three have no template
placeholders at all and two handle their own absences. And it claimed the notes we file would be
picked up automatically by an existing handler: they were not, because we never supplied the two
fields that handler looks up. Two notes had been filed since the feature went live and both had
been rejected and parked — a 100% failure rate, and the first was real customer traffic, not our
test.

**Proved it rather than asserted it.** Every mechanism was checked on the running system: a real
edit refused with the live page left untouched to the byte, a clean edit still saving, and — for
the queue note — the handler's own lookup query taken out of the live configuration and run by
hand against the real note, which is what showed the fix working and then showed the next problem.

## Where we are now

**Closed.** `bugs_open/342` moved to `bugs_closed/342`. Re-verified this morning on **v1.0.1332**
(the fleet has rolled twice since the fix landed): all three mechanisms present on both replicas
with a negative control, both switches still armed. Three council reviews, all approved.

The bug's own defect is fixed on every path where a fix was licensed: reported at nine of fifteen
render sites — the other six cannot carry the check by construction, each verified individually —
escalated with a note the handler can now read, and refused where refusing is allowed.

## Where we're going — three things on file, none of them this lane's

- **`bugs_open/367`**, filed by this lane on its last day: the queue handler only looks at pages
  already published, so it closes as "cannot be found" exactly the unpublished sections our note
  exists to report. Confirmed by the live handler, not just predicted. The uncomfortable part is
  that the new failure is *quieter* than the one it replaced — a true finding now sits in the
  queue marked complete, with no error, and any count of "did we action these?" scores it a
  success.
- **`bugs_open/344`**: when an edit is refused, the job record that drove it may still report
  success. Genuinely untested rather than cleared — our test was hand-run and had no job record.
- **The chrome refusal**, built and deliberately off until something can trigger it.

Two lessons left where the next person will hit them, both in `LANDMINES.md`: **reusing a type is
not reusing its contract** — a work item can be born unroutable while every signal on the
producing side says success — and its inverse, that **the leftovers have a convention too**, which
the review council caught us ignoring the day after we wrote the first one down.
