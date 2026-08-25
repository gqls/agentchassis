# SUMMARY 2026-08-25b — `bugs_open/381`: proven on a live build. 0 of 7 pages carried structure; now 19 of 20.

*(Second summary of the day. The first, `SUMMARY_2026-08-25_…`, said the planner arm was live and
completely untested. That is what changed, so this is a genuine inflection rather than a restatement.)*

## What we're trying to do

Stop pages promising things they cannot deliver. The complaint was concrete: a page headed *"What
your shed needs, month by month"* that contained no months, and a 300-word paragraph that should
have been a list.

## Where we've come from

The cause was that the part choosing a page's components could not see what any component was
capable of, and the part writing the words had been told, in effect, to write paragraphs. Both were
fixed in configuration over 24 August — eight changes, two independent review rounds, all approved
— and three missing components were built: a checklist, a period calendar, and a comparison table.

By this morning the writer half was proven and **the planner half had never once run**, because only
a new-site build triggers it. The owner offered a list of domains. We built `homegarden.uk`.

## What we've done

The build ran from 10:21Z to about 12:23Z and produced 21 pages. Every link in the chain was
measured at the layer where it applies, rather than wherever was convenient:

- **The planner was told.** Its rendered prompt was captured automatically the moment it existed —
  it carries the capability tokens and the new rule, and all three new components appear in its menu.
- **The planner chose.** It filed `period-calendar` on the landing page.
- **The writer filled it.** Twelve months, January to December, each with its own focus and
  practical detail — including honest hedging that nothing in the fix could have forced
  (*"none of this is urgent to the day"*).
- **And the writing changed everywhere else too.**

| measured at stored content | `homegarden.uk` | `garden-tools.uk` |
|---|---|---|
| pages with content | 20 of 21 | 7 |
| sections carrying a list | **20 of 45 (44%)** | **0 of 22** |
| sections carrying emphasis | **11 (24%)** | **0** |
| **pages carrying any structure** | **19 of 20** | **0 of 7** |

## Where we are now

**The bug as filed is fixed, and the fix is demonstrated end to end on a real site.**

Four things are honestly outstanding, none of them the original defect:

- **The checklist component was offered and never chosen.** A real negative, unexplained. Worth one
  investigation — though the mechanism it would test is already proven by the calendar.
- **The comparison table was never exercised.** The subject research came back with no comparison
  material at all, so the planner had no reason to reach for it. It needs a build whose subject is
  genuinely comparison-shaped.
- **The mobile card-wall complaint** from the original review has still never been filed as a bug.
- **The domain is parked.** DNS has never been pointed at the platform, so nothing is publicly
  visible. That is an operator action, and it blocks the other lane's promise-versus-delivery check,
  which reads served pages. It touches none of the results above, all of which are stored artefacts.

## Where we're going

Closing this bug is now a decision rather than a wait. What would make it worth *not* closing is the
checklist question — if a planner that can see capability still declines a component that fits, that
is a second-order version of the same defect and belongs to this lane rather than a new one.

The larger thing this build produced was not the result but the method. Seven claims were wrong
across two lanes in one morning, three of them ours, and **not one was caught by its author** — every
one was caught by the other lane asking a question, usually a question the asker thought was routine.
Five of the seven reasoned from a structural proxy one join away from the thing itself: a role
instead of a layout, a count instead of content, a status instead of an artefact, a clock instead of
a timestamp. That tally is filed fleet-wide, and it argues for something cheaper than more care:
before publishing a structural claim, say it to another lane first.
