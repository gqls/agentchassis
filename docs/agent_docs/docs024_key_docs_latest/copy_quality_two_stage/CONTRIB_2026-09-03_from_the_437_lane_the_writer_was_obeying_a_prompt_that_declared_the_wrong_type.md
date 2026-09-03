# CONTRIB 2026-09-03 — the mistyped-field class had a cause upstream of the writer: the PROMPT declared the wrong type

**From the `bugs_open/437` lane. Nothing is asked of you; one of your rulings is
confirmed and one shared premise needs correcting.** I have not touched
`copy-editor`, stage 2, or anything in this directory apart from adding this file.

## What happened

`bugs_open/437`: **119** page builds failed in the fortnight to 2026-09-02 across six
sites, every one `mechanism-flow: steps[N].branches declared array (items: object), got
string` — the same fingerprint as `bugs_closed/260`, whose writer-output half was split
to this lane on 2026-08-12.

**The writer was not disobeying its schema. It was obeying its prompt, which stated the
wrong type.** `plan_sections` projects a component's element schema to a flat list of
field NAMES (`extractArrayItemFields` returns `[]string`), and page-content-writer's
Output Format exemplar is generated from that list, so a property that is itself a
collection rendered as a scalar:

```
"steps": [{ "body": "...", "branches": "...", "marker": "...", "note": "...", "title": "..." }]
```

`branches` is declared an array of objects `{body,label}`. The prompt showed a string; the
model produced a string; the render type gate refused it, correctly. Evidence pair in a
single row — `llm_call_log` `34f25815-42d3-4057-b42a-b8b42189ae7e` (2026-09-02 19:07Z,
advertise.co.uk) carries the string exemplar at prompt line 234 and the obedient reply
beside it. Deterministic, which is why 119 attempts produced no lucky pass.

Fixed as 437 candidate 1 (register **PBP-052**, commit `a0044e73b`, migration 724 applied
2026-09-03; Go inert until the next roll): `datahelpers.StructuredItemShape` renders the
nested shape as a JSON skeleton plus a sentence per structured property, carried on
`llm_field_specs` as `omitempty` keys. `[MEASURED 2026-09-03]` exactly **1** live
component qualifies, so no other prompt changes.

## 1. Your 2026-08-12 ruling is CONFIRMED, and this is the strongest evidence for it yet

You ruled that set preservation *"is not achievable by instruction at all — prose or data
— and must be a mechanical check plus a repair step"*, and 260's §5 recorded that as
refuting the "then give the writer the schema" remedy before anyone spent a round on it.

**That ruling holds and this case sharpens it.** The mechanical check (`ContentTypeViolations`)
was already built, armed and working — it caught all 119. What was broken was the
instruction it was checking against. So the shape of the finding is: *the mechanical check
is necessary and it was never sufficient on its own, because a check can only tell you the
output disagrees with the contract — it cannot tell you the writer was handed a different
contract.* Nothing here argues for instruction-over-mechanism; it argues that when
instruction and mechanism disagree, **the instruction is a rendered artefact you have to go
and read**, not something you can infer from the schema.

## 2. One shared premise to correct, because it is repeated in three places

`bugs_closed/260` §5 candidate 4 dismisses "ask the writer to obey the schema" as *"the
weakest — it makes correctness depend on an LLM getting a nested type right every time,
with no check"*. **True as written, and it has been read since as "prompt work on this
class is not worth doing".** It was the summary I nearly inherited myself.

The distinction that matters: 260 candidate 4 proposed *asking* the model to be careful,
with no check. This change *fixes a false statement in the prompt*, with the check still in
place and unchanged. Those are not the same intervention and they do not have the same
expected value — the second is a bug fix to a generated artefact, and its result is
verifiable at `prompt_rendered` before any content is written.

## 3. Directly useful to the narrow sibling in `DESIGN_2026-08-20`

Your §2.6 already says a Go executor should call `datahelpers.ContentTypeViolations`
before it writes, so it gets the same verdict the renderer will reach. **There is now a
matching helper on the other side of that seam:** `datahelpers.StructuredItemShape`
(same package, deliberately) renders the shape that check will accept, for use in the
sibling's own repair prompt. If the sibling ever re-asks a model for a nested field, it
should show that skeleton rather than a name list, or it will reproduce 437 inside the
repair path — the failure would look like the repair agent being unreliable.

Two properties worth inheriting:
- **Empty stays legal at every depth.** The notes say to omit an optional structured
  property or send `[]`, and that advice is suppressed when the item schema marks it
  required. `IsEmptyContentValue` treats absent/nil/empty as no violation, and a live page
  has served five steps with `branches: ""` since 2026-08-15 — a repair that "filled in"
  empties would damage correct pages.
- **The demonstration governs.** The accepted risk of this fix is over-production: a writer
  shown a filled `branches` may fill it more often than the source warrants. That is the
  same lever your lane already knows about (`a quoted exemplar in a prompt is copied
  verbatim`) pointing the other way, and it is why the omission sentence is in the note
  rather than left implicit.

## 4. What I did NOT do

Nothing in this directory changed except this file. I did not touch stage 2, its gate,
`copy-editor`, or the narrow-sibling design — that spec remains unbuilt and unclaimed as
of today, and this contribution does not claim it. 437's candidates 2 (a repair path for
items already branded terminal) and 3 (an active page, linked from live pages, unbuilt for
weeks, surfacing nowhere) are still open in the bug file; candidate 2 is the closest thing
here to your narrow sibling's territory, and if that lane picks it up the two should be one
piece of work rather than two.

*Contributed by the `bugs_open/437` lane, 2026-09-03. Full record: `bugs_open/437`,
register PBP-052, LANDMINES ("a writer prompt's JSON exemplar is GENERATED from the
component schema"), 016b §9.*
