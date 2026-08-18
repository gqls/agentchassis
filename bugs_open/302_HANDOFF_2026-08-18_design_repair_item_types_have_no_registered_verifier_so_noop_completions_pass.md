# 302 — design-repair item types have no registered verifier, so a no-op "repair" completes unverified

**Filed 2026-08-18 by the `finetuning_uk_service` (merged) lane.**

## Verification statement (owner ruling 2026-07-31 compliance)

A `090` ran first — twice. Run 1 (`f60d72d6`) FAILED with NULL step errors after
five bundles. Run 2 (`361605fe`) returned **UNVERIFIABLE** — "NOT confirmed
(stopped: scope-not-narrowing)", the broad hypothesis as submitted marked
REFUTED, and the trail explicitly handed to a human. **This filing substitutes
declared first-hand verification for the loop's confirmation, per the named
escape hatch:** the narrow claim below is established by direct code reading
with quotes, and the loop's own citation pointed at the deciding arm.

## The narrow, code-verified claim

`platform/orchestration/actions/complete_work_item_verification.go`
(`verifyBeforeComplete`):

```go
verifier, policy := checks.GetVerifier(itemType)
if verifier == nil {
    return nil, true, abstained   // <- completes; records only an abstention
}
```

The verifier registry (`discovery_checks/`, `RegisterVerifier` call sites) holds
**eleven item types, all discovery-check shapes**: `revenue_shape_cta`,
`missing_conversion_path`, `content_duplication`, `decision_regression`,
`page_canonical_collision`, `orphan_element_refs`, `empty_section`,
`truncated_component`, `dead_fragment_link`, `unbuilt_internal_link`,
`literal_markdown`. **No design-repair item type is registered** — nothing in
the family handled by `webdesign-agent` / `component-template-fixer` /
`color-variable-fixer`. So for those items, gate verification abstains and
completion passes with no check that the named defect changed at the artefact.
(There is a second abstain arm just above for unknown result shapes — same
consequence.)

## The evidence this explains (measured, not asserted)

finetuning.uk, 2026-08-12 (evidence rows deliberately left `complete`; full
tables in `finetuning_uk_repair/NOTES` §"ALL FOUR REPAIRS"): four repair items
completed in 6 minutes; the served page byte-identical on every defect before
vs after; zero writes to `page_components`/`content_components` in the window;
every `result` a four-key design-token blob with no `changes_made`. Same shape
on `needs_design_review` items of 2026-08-11. The 08-09 audit's
`hardcoded_section_colors` finding was re-filed by the 08-12 audit because
nothing had changed.

## What is NOT claimed (the 090 refused the broader claim; respect that)

WHY the handlers return analysis blobs instead of performing repairs is
**undiagnosed** — the loop could not narrow it (its runtime evidence covered an
unrelated target) and marked the broad claim REFUTED as stated. Do not treat
this file as saying "the handlers are broken"; it says **the gate that would
have caught them is absent for their item types**. The blob question is real
and separate; the loop's `NextScope` pointers (completion path for
`color-variable-fixer`, which was absent from indexed scope) are the thread to
pull.

## Fix candidates, ordered by what closes the door

1. **Class fix at the gate:** for item types whose *name or category marks them
   as repair-shaped*, make a missing verifier REFUSE completion rather than
   abstain (fail-closed for repairs, abstain-open only for informational
   types). Closes the whole class, including future unregistered types — the
   current design makes every NEW repair type silently unverified by default.
2. **Instance fix:** register artefact-level verifiers for the design-repair
   family (before/after fetch of the named defect, the same discipline the
   eleven existing verifiers use). Necessary anyway for repairs to be provable;
   does not close the door for the next type.
3. **Not a fix:** relying on audits re-filing the same finding — that is the
   current de facto detector and it costs a full audit cycle per miss.

## Relations

`bugs_open/201` §symptom 2 — same class, different item family, established the
"unregistered verifier + mark_complete checks nothing" pattern from code;
OWNED by `bugfix_201_…` lane (this file does not route work at their case).
`bugs_closed/213` — verifier/producer mismatch class. Fleet memory:
"a `complete` work item is not a repaired artefact". 016b §9 entry added this
date. 090 artefact trail: correlations `f60d72d6` (failed), `361605fe`
(UNVERIFIABLE, bundles + evidence trail on the item row).

## How to verify a fix

Re-run the finetuning.uk repair items (or any design-repair item) after the
change: a no-op completion must FAIL (candidate 1) or the verifier must measure
the artefact and refuse (candidate 2). The four retained `complete` rows are
the regression fixtures.
