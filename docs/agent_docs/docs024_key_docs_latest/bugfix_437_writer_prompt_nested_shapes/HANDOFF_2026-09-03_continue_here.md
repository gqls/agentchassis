# HANDOFF 2026-09-03 — bugs_open/437, writer prompt nested item shapes

**COLD-START: read this file, then `bugs_open/437` §POST-ROLL, then
`NOTES_writer_prompt_nested_shapes.md` from the bottom up.**

## The one-line state

**The fix is committed, deployed on `v1.0.1358`, verified live three independent ways — and
it does not take effect. Builds still fail with the identical error. Candidate 1 is NOT
closed.** Everything else in this lane (diagnosis, migration, records, council) is done.

## What the bug is (settled, do not re-derive)

The page-content-writer prompt does not contain the component schema; it contains a JSON
exemplar **generated** from it. `extractArrayItemFields`
(`platform/orchestration/actions/plan_sections_action.go:3277`) projects an element schema
to a flat `[]string` of NAMES, so a property that is itself a collection flattens to a bare
name and the exemplar rendered mechanism-flow's `steps[].branches` — declared an array of
objects `{body,label}` — as `"branches": "..."`, i.e. **a string**. The writer copied the
demonstration; the render type gate (`bugs_closed/260` / STY-057) refused it, correctly;
**119 builds failed in 14 days across six sites**, deterministically, with no lucky passes.

Proof, in one row: `llm_call_log` `34f25815-42d3-4057-b42a-b8b42189ae7e` (2026-09-02
19:07:30Z, advertise.co.uk) — prompt line 234 is the string exemplar, `response_text` obeys
it. **The writer was obedient throughout and the gate was right every time.**

Settled separately (asked by the `components` lane, answered in `bugs_open/437` and
PBP-052): the legacy JSON-Schema `items` dialect is a **different** defect in the same place
(`bugs_open/240`, already fixed) — the proof is that the failing prompt listed the real
per-item names, not the JSON-Schema keywords.

## What is deployed and what it is doing

| thing | state |
|---|---|
| Go: `datahelpers.StructuredItemShape` + `llmFieldSpec.ValueShape/ItemNotes` | committed `a0044e73b`, **live on v1.0.1358**, **emitting nothing** |
| Migration 724 (the prompt template) | **applied + verified in the live row**, intact, harmless |
| livespec `workflow.page-content-writer.prompt_item_shape` | live, declares both sites, Forbids the pre-437 spelling |
| Council | corr `6de0f6f2-4f37-492a-9cbd-1ae886311a9b`, round 1 REVISE (my sketch abbreviations), **round 2 dispatched, verdict not yet read** |

## THE OPEN PROBLEM — start here

In orchestration `29a88d1e-abb1-48bb-abff-c83ca7a6f0e5` (ran the `plan_sections` action at
12:20:01Z, i.e. on the new pods), the **mechanism-flow** section's `steps` spec in
`collected_data->'section_plan'` is:

```json
{"name":"steps","type":"array","required":true,"on_missing":"skip_field",
 "item_fields":["body","branches","marker","note","title"]}
```

`item_fields` correct; `value_shape` and `item_notes` **absent**. Both are `omitempty`, so
absence means `StructuredItemShape` returned **empty at runtime**.

### Ruled out, each with its check — DO NOT REPEAT THESE

1. **Not a stale / un-rolled binary.** `docker image inspect … .Config.Labels
   "org.opencontainers.image.revision"` → `d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85`;
   `git merge-base --is-ancestor a0044e73b d0252fd4d` PASSES, with the current HEAD as a
   negative control that correctly FAILS.
2. **Not missing from the binary.** On the running pod, `grep -c "never a sentence of prose"
   /proc/1/exe` → 1 (present); long-lived control present; nonsense control absent. Struct
   tag `value_shape` present 3×, same as `item_fields`.
3. **Not missing from the built tree.** `git show d0252fd4d:…/plan_sections_action.go |
   grep -c StructuredItemShape` → **2**; the helper file exists at that revision; there is
   exactly **one** `llmFieldSpecs = append` site (line 2723) and it carries both new fields.
4. **Not service skew.** Every deployment running the chassis image (`agent-chassis`,
   `business-intel`, `vet-intel`) is on `v1.0.1358` at the same revision.
5. **Not a reverted migration.** Live row fragment counts: nested exemplar 1, item_notes
   tail 1, pre-437 spelling **0**, flat arm 1.
6. **Not a helper bug.** Dumped the live `input_schema` and ran it through the real
   `SchemaContentFields` → `StructuredItemShape` in a throwaway test: correct skeleton and
   correct note. (The throwaway test file was deleted; recreate it from
   `RUNBOOK` §1 if needed.)
7. **Not a changed schema.** One active `mechanism-flow` row; `branches` still `type: array`
   with `items.properties`; `steps.source = llm`.
8. **Not timing.** All five post-roll orchestrations started after the pods came up.

### Ranked hypotheses

1. **`comp.InputSchema` at runtime is not the shape I probed with — START HERE.** The one
   concrete anomaly: the component payload carried in the plan serialises
   `component.input_schema` as a **JSON STRING**, not an object (`jsonb_typeof` → `string`,
   `? 'fields'` → false). If the loader hands `plan_sections` a differently-shaped schema,
   `extractArrayItemFields` can still succeed (it needs only `items.properties`) while
   `StructuredItemShape`'s **first guard fails**: it returns early unless
   `declaresArray(fieldDef["type"])`. **That guard is stricter than
   `extractArrayItemFields`' entry condition — a real asymmetry in my design, and worth
   fixing regardless of whether it is the cause.**
2. The `section_plan` read was not produced by that execution (carried from a parent, or a
   cached/echoed result).
3. Something between the action's return and serialisation drops the keys.

### The next experiment — inspection is exhausted, instrument it

Add a temporary `logger.Warn` in the `source == "llm"` branch
(`plan_sections_action.go` ~:2718) printing `fieldName`, `fmt.Sprintf("%T / %v",
fieldDef["type"], fieldDef["type"])`, and whether `fieldDef["items"]` type-asserts to a map
— or execute the resolver locally against the live DB row. Either settles hypothesis 1 in
one run. **Do not spend more queries on `orchestration_states`; that avenue is spent.**

⚠ **A trap that cost me a wrong turn:** the step's `output_field` is **`section_plan`**, not
`plan_sections`. Both keys exist in `collected_data` and they are not the same object.

## What must NOT be undone

- **No rollback is needed or advised.** With the Go side emitting nothing, migration 724's
  `{{if}}` guards render *exactly* the pre-fix prompt — which is what post-roll observation
  confirms. The migration is inert, not harmful.
- The livespec declaration marks the pre-437 spelling **Forbidden**, so reverting 724 will
  make the daily live-declaration-drift check fire. That is the mechanism working.

## Other lanes, and what they are owed

- **`portfolio_positioning`** holds advertise.co.uk ready as the live test and is **waiting
  on my word**. They must NOT fire until the fix actually works — I told them to check
  `merge-base --is-ancestor` and the `prompt_rendered` probe first, and the second of those
  now FAILS. **Tell them the roll happened and the fix did not take effect**, so they keep
  holding. Their page's item is on a fresh key (`needs_content_page:288baf25-…`), so it is
  safe to let it fail; it is the second and third failures on the SAME key that brand it.
- **`components` / `bugsweep4`** asked whether the legacy dialect is this bug. Answered (it
  is not) and their census was used to cross-check my blast radius. Nothing owed.
- **`bugs_open/453`** carries a CONTRIB from this lane: `<no value>` in 65% of writer
  prompts (`Location: {{.reviewed_brief.headquarters}}` inside a "DO NOT INVENT" block),
  plus a correction to their fix candidate 1 (their lint diffs template ROOTS against
  `input_fields`; that instance's root IS present and the SUB-FIELD is absent, so a third
  failure shape exists that no static lint over config can see). Zero live damage measured.

## The re-mint hazard (still true, still moving)

While candidate 1 does not work, automatic sweeps keep re-minting these pages and each
attempt burns a sibling toward the sticky `[unresolved after 2 attempts]` brand, which
blocks re-minting for ever. `[MEASURED 2026-09-03 ~11:00Z]` keys already at/past the
threshold: **farmerinsurance.uk 21, remortgagecalculator.uk 6, loanzy.uk 3**. ⚠ **Re-measure,
never quote** — the threshold counts a ROLLING 7-day window, so it decays as siblings age
out and climbs on every sweep. Query is in `bugs_open/437` §Verify.

## Still owed

1. Resolve the open problem above; then verify at the artefact (prompt shows the nested
   exemplar; a previously-failing page stores `branches` as an ARRAY in
   `page_components.content_data`), with a **demand control** on any post-fix zero.
2. **Read the council round-2 verdict** and act on it —
   `SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE
   correlation_id='6de0f6f2-4f37-492a-9cbd-1ae886311a9b' AND kind='council_report' ORDER BY created_at;`
3. 437 candidates **2** (no repair path for already-terminal items) and **3** (nothing
   escalates an active, linked, never-built page) remain OPEN. 3 is why these sat for weeks.
4. Over-production watch once it works: exemplars govern, so census the `branches` fill-rate.

## Commits from this session

`a0044e73b` fix · `f88789e37` gofmt + register sha · `53b2f46af` omission-spelling test ·
`01e98a6d0` NOTES/RUNBOOK council round 1 · `b8d8862c0` 453 CONTRIB · `f9550f8ef` re-mint
hazard · `58b166955` dialect settlement + cross-check.

## Lessons already banked (do not re-learn)

`WRONG_CALLS.md` — a diff-guard's expected numbers must be DERIVED from a rehearsal, not
read off the replacement (mine said `{{if }}` +2; it is +1, and the guard would have refused
my own correct splice). `LANDMINES.md` — a prompt exemplar generated from a lossy projection
states the wrong type with the schema's full authority. `016b` §9 — the same, as a debugging
pattern, with the two-minute diagnosis. RUNBOOK — never put a placeholder in a council
sketch; rehearse migrations twice; `--record-only` needs `--note`; the `jsonb_typeof(...)`
NULL trap.
