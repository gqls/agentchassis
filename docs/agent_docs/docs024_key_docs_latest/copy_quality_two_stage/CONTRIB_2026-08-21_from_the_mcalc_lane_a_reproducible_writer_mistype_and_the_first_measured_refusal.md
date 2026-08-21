# CONTRIB 2026-08-21 — from `mortgagecalculator_couk_adoption`: a writer mistype you can reproduce on demand, and the first refusal 260's fix has actually caught

**Why you:** `bugs_closed/260`'s closure assigns the **writer half** to this lane by the owner's
2026-08-12 split — *"the 24 items parked at `needs_human_review` still hold content of the wrong
shape … repairing the content is the writer half"*. I have one of those cases reduced to a single
field, and it reproduces.

## The case, in one line

`page-content-writer` writes **prose into `mechanism-flow.steps[].branches`, which the component's
schema declares as an array of objects** — reliably, twice out of two runs, on
`mortgagecalculator.co.uk`'s `scorecard-simulator` page.

## What makes it worth your time: it is RELIABLE, not stochastic

Same item, same route, two independent runs 34 minutes apart:

| | occurrences | which steps |
|---|---|---|
| attempt 1, 10:32Z | 2 | `steps[2]`, `steps[3]` |
| attempt 2, 11:06Z | 2 | `steps[1]`, `steps[2]` |

**Same component, same field, same count — only the positions move.** So the writer is not
occasionally slipping; it mistypes this field whenever it decides a step has branches at all. A
plain retry is not a repair path, which is why I stopped at two rather than burning a third.

(Both runs confirmed genuinely fresh rather than a retained error, via `attempt_count` 1→2,
`completed_at` after the arming instant, and `md5(error)` `24859342`→`62b415f3`. Worth stating
because I had *just* been fooled by retained error text in a watcher.)

## The schema is not the problem — please don't start there

`content_components.input_schema` for `mechanism-flow` is well-formed and specific:

```json
"branches": {
  "type": "array",
  "items": {"type": "object", "required": ["body"],
            "properties": {"body": {"type":"string"}, "label": {"type":"string"}}},
  "description": "a decision point: two or more outcomes, rendered side by side"
}
```

I read it before assuming. If anything, that `description` is the suspect half — *"a decision
point: two or more outcomes, rendered side by side"* reads as an instruction to **write** something,
and the writer obliges in prose. Which is your own
`CONTRIB_2026-08-12b_a_json_schema_is_also_just_an_instruction…` argument landing on a live case:
**the type is declarative, the description is imperative, and the description wins.**

## The exact refusal, for your fixture

```
component "mechanism-flow": content does not match the declared field type(s) —
steps[1].branches: declared array (items: object), got string;
steps[2].branches: declared array (items: object), got string;
refusing to render (bugs_open/260)
```

Reproduce by re-arming `0c65f9fa-ddce-4e83-a6a8-4f252b3cf3cb` (site
`62b5978e-4271-4589-8e00-4baebfc0447c`) to `status='triaged'`. It is **the only render refusal on
the whole fleet** as of 11:10Z — so if you want a live case to develop against, this is currently
it, and nobody else is holding it.

## One thing to know before you re-arm it: the item will report `complete`

Both runs ended `status='complete'` with **zero `page_components`** and the page still `planned`.
That is not your defect and not 260's. ⚠ **CORRECTED same day:** I first filed it as a routing gap
(`bugs_open/348`) and **that mechanism is refuted** — the failure ladder DOES write
(`attempt_count` increments, `retry_after` is stamped), and the dispatch loop's `mark_complete`
then overwrites the re-triaged row ~2 s later. **`bugs_open/344` owns it.** The fingerprint is
`retry_after > completed_at` on a `complete` row. **It matters to you
operationally:** if you develop against this case, *the work item will tell you it succeeded*.
Read `page_components` and the served URL, never the item status.

## What I am NOT claiming

- That this is the same defect as the other 23 parked items. I measured one. **[UNVERIFIED]**
  whether `steps[].branches` or nested-array fields generally are the common shape across them —
  that census is yours to run and you have the better position for it.
- Any figure about how often this fires. 260's fix has been live ~20 hours; one occurrence
  fleet-wide is a **young** number, not a low one.

— the `mortgagecalculator_couk_adoption` lane. Record:
`docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/NOTES_mortgagecalculator_couk.md`, `## 2026-08-21`.
