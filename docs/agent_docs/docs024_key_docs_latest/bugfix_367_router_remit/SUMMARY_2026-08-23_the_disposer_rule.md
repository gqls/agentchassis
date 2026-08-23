# SUMMARY — 2026-08-23: a router that could not find something was allowed to say it was gone

## What we're trying to do

Stop a queue from marking true findings as resolved. `required_fields_missing` work items
report that a page component is missing text its own schema declares as required. A router
agent sorts each one: convert it to a repair job, park it for a human, or close it. We want
every such finding to end up somewhere a person or a pipeline can act on — and never to end
up looking handled when it is not.

## Where we've come from

`bugs_closed/277` built that router in August and gave it a sensible way to find the
component a finding names: look among components that have been published. That mirrored the
only producer of these findings at the time, a post-deploy check that also only looks at
published components.

`bugs_closed/342` then added a second producer, at render time, whose stated justification
is reaching the population the post-deploy check structurally cannot — components that render
empty and never get published. The 342 lane spotted the collision on its last day, filed it
as `bugs_open/367`, corrected its own over-claim in two places, and closed itself.

## What we've done

Confirmed the defect on the live system first: the router's own SQL, run by hand, said the
component could not be located, while that component sat there unpublished with 9,220
characters of content and both named fields genuinely empty.

Then we changed the rule rather than the filter. The router may now close a finding only on
**positive evidence of absence** — the page is deleted, the component is explicitly retired,
or a human has locked it as accept-as-is. A lookup that simply finds nothing, or finds a
component that is real but unpublished, now **parks** for a human with the facts and the
three resolutions written on the row. That is the same rule the review-queue revalidator
already states in code, in almost the same words.

We deliberately did **not** widen the router into the repair path. Measurement showed why:
the repair step would have crashed on the new producer's spec shape, and the repair it leads
to deletes and rebuilds the whole page — a road on which 28 of 31 previous conversions have
already failed against hand-owned pages (`bugs_open/333`, owned elsewhere).

The change is one config-only migration, live on apply, no image and no roll. It was proven
inside a rolled-back transaction before being applied: three routing controls, a
whole-population re-run of all 65 findings ever filed showing exactly one changed route, and
an apply-then-rollback round trip returning the config byte-identical.

## Where we are now

Live and verified against the real row. The one affected finding is no longer closed as
imaginary; it will park with its state named. Every other route is untouched, and both
genuine close routes — retired component, deleted page — still close.

Three of our own missteps are logged in `WRONG_CALLS.md`: a JSON shape read wrongly that
produced a confident zero, a retention window inferred from a survivor artefact, and a count
quoted from a query that carried our own `LIMIT`.

## Where we're going

The findings are now honest, not repaired. Making them repairable needs `bugs_open/333`
(owned pages queue findings that can only be refused) and a producer that writes the convert
arm's read-set. Both are named; neither was taken here.

The larger finding is filed separately: of the three code paths that can stamp a work item
`complete`, the one this router uses never consults the verifier framework built to stop
false completions. That is why a silent close was reachable at all, and it is
architecture-scope — a decision for a human, not a change to smuggle into a bug patch.
