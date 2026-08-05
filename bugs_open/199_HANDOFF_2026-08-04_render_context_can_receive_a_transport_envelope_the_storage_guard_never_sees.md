# 199 — `extractContentWithFallbacks` can hand a transport envelope to RENDER context, where the storage guard cannot see it

**Filed 2026-08-04 by the `bugs_open/190` lane, at the council's explicit request.** The
`bug_historian` seat objected (medium) that naming this path in a rationale field and calling
it "separable" is *"a silent decision"*, and that the correct disposition is a tracked bug:

> "the shared defect class (envelope→empty/raw-JSON render) gets a rigorous fix at two seams
> while a third, named, live path stays generic and exploitable — the same shape as case 7
> (Go template `missingkey=zero` fixed at one call site, root behaviour left open).
> **Disclosing the gap is good practice, but 'separable' does not make it safe.**"

It is right, and this file is the disposition. `190`'s guard is not weakened or widened to
cover this; the two concerns have different blast radii and want separate decisions.

## The gap

`190` closed the **storage** seam: nothing can now persist
`{"type":"text","result":"<string>"}` into `page_components.content_data` — it is decoded when
that is lossless and the save is refused when it is not
(`platform/orchestration/actions/content_data_envelope_guard.go`, concept register PBP-032).

That guard sits at the two write seams. It does **not** sit on the path from an LLM step's
output into **render context**. `extractContentWithFallbacks`
(`platform/orchestration/actions/v3_site_actions.go`) resolves a component's content from a
step's output, and when the LLM fell to the text path the value it finds is the envelope map
itself — `result` is a *string*, so the "unwrap `.result`" branch does not fire, and the
"last resort" branch can return the whole envelope. `RenderComponentAction` then renders
against it.

> **CORRECTED 2026-08-05 — the branch named above is the wrong one, and the right answer
> changes where the fix goes.** It is the **first fallback loop** (`:4494-4503`) that returns
> the envelope, not the last-resort branch. For the live config
> (`content_from: "generated_content.result"`), `pathsToTry` is
> `[generated_content.result, generated_content, generated_content.response,
> generated_content.content]`: path[0] resolves to the envelope's `result` **string**, the
> map assertion fails and the loop continues; path[1] resolves to the envelope **map**,
> passes the bare `len(m) > 0` test, and is returned. `hasContentFields` never gets a say —
> it guards only the last-resort branch.
>
> The last-resort branch is nonetheless a **second real leak**, not a dead one: a superset
> envelope such as finetuning's `{content,result,type}` passes `hasContentFields` precisely
> because `content` is in its list (`:4531-4536`). **This is why the fix sits at the single
> CALLER rather than inside `extractContentWithFallbacks`** — one call covers both branches
> identically, and a future reader who "fixes" the first loop cannot silently reopen the
> second. Caught by reading the resolver against the live `page-content-writer` config; the
> misstep is logged in `WRONG_CALLS.md`.

**What stops it today, and why that is not the same as it being closed:** the render gate
(`missingRequiredLLMFields`) refuses when the component declares required `source:"llm"`
fields. A component with an **empty or optional-only input schema** declares nothing, so
there is nothing for that gate to find missing. That is the hole — and it is the *same* hole
shape as `missingkey=zero`, where a template rendered a blank rather than failing.

## Status of the claim — read this before acting

~~`[UNMEASURED]`~~ **`[MEASURED]` 2026-08-05, live — and the census says BUG, not note.**
The original text is kept below the results because its instruction (take the denominator in
the same query) is what made the answer readable.

### The census, run 2026-08-05 against `clients_db`

**The population the render gate cannot speak for is 315 of 1212 live `page_components` rows
(26%).** Numerator and denominator in one query, components classified exactly as
`missingRequiredLLMFields` + `SchemaContentFields` classify them:

| class | components | active | live `page_components` rows |
|---|---|---|---|
| no / empty `input_schema` → gate skipped (`len(comp.InputSchema) > 0` false) | 56 | 35 | **44** |
| unrecognised dialect → `SchemaContentFields` returns `!ok` | 8 | 1 | **1** |
| v2 dialect, no `required` `source:"llm"` field → gate finds nothing missing | 64 | 48 | **270** |
| **gate CAN speak** (v2 + legacy with a required llm field) | 102 | 98 | 855 |

**The mechanism has fired against exactly this population.** The one envelope still live in
`page_components.content_data` — gaswholesalers.com, `how-pricing-works`, slot `pricing`,
keys `{result,type}`, 1363-byte `result` — sits on component `pricing`, which is v2-dialect
with **no required llm field**. It is gate-blind. `190`'s lane established that its payload
recovers only via `prose_around`, so it is a REFUSE case at the render seam too, and both
seams agree about it.

**Zero false positives at this seam.** Of 114 live `render_context` maps in
`orchestration_states`, **not one carries a `type` key at all** — so `render_from_template`,
whose `content_from` *is* `render_context`, cannot trip the predicate. Every envelope-shaped
object anywhere in live `collected_data` has keys exactly `{result,type}` and is an LLM step
output.

### ⚠ The trigger rate is ZERO, and that is stated rather than buried

Across the ~25h `orchestration_states` retention window, **62 runs carried a
`generated_content` step output and none was envelope-shaped.** The measurement is
disconfirmable — the same query returns **111** hits for `compose_note` and 10 for
`generate_css` in the same window, so it could have come out otherwise. Bound per status, not
whole-table: `COMPLETED` reaches back to 2026-08-04 09:28, `FAILED` to 10:35.

So the honest reading is **"the door is open and currently unused"**, not "the door is shut".
Two consequences a later reader must not confuse:

- an `agent_error_log` count of zero `action='render_component'` rows is the **expected**
  post-roll reading and proves nothing on its own;
- it also cannot be read until the guard is proven live by pod-grep — **an unrolled image
  and a quiet guard produce the identical count**.

### The original instruction, kept because it is what made the answer legible

`190`'s lane named the path from a code read and deliberately did not ship a change to it,
because it alters render behaviour for schema-less components and belongs on its own merits
rather than inside a bug patch (the 2026-07-28 platform-seams ruling).

```sql
-- components whose input_schema declares NO required source:'llm' field, joined to
-- live page_components — the population the render gate cannot speak for.
-- Take the DENOMINATOR in the same query (all active components), because a zero
-- here is indistinguishable from a broken predicate otherwise.
```

## Why it was not fixed inside 190

Stated plainly rather than left as an inference:

1. **Different blast radius.** `190`'s guard narrows what a seam will *store* and cannot
   affect a page that is not being saved. A change to `extractContentWithFallbacks` changes
   what schema-less components *render*, on every build, immediately.
2. **The platform-seams ruling.** A change to shared render-resolution behaviour is not the
   same decision as a storage guard, and bundling it would have put an architecture-scope
   change inside a bug patch — exactly what `bugs_closed/124` drew a REJECTED verdict for.
3. **It was declared, not omitted.** `190`'s council submission listed it as risk 6, "not
   shipped, deliberately", and the `constitution` seat credited that as satisfying the
   stated-deferral requirement. The `bug_historian` seat's point is that a declaration in a
   rationale field is not a tracked item — hence this file.

## Fix candidates, ordered by what closes the door

1. **Refuse the envelope shape at the resolution point.** `extractContentWithFallbacks`
   returns "no content found" instead of a map matching `isLLMTransportEnvelope`, so the
   existing not-found path handles it. The predicate already exists and is tested; this is a
   call, not new logic. Closes it for every component regardless of schema. **Needs the
   census first**, because it changes what a schema-less component renders.
2. **Make the render gate speak for schema-less components** — treat "declares no required
   fields" as a case needing an explicit opt-in rather than a silent pass. Structurally
   stronger and a much larger decision; probably an RFC.
3. **Do nothing and rely on 190's storage guard.** Defensible *only* if the census shows no
   live component can reach the branch. Not a fix — a measurement that would close the file.

## How to verify a fix

Do **not** grade it on a build passing. Take a component with an empty input schema, drive a
step whose LLM output falls to the text path, and assert the render refuses rather than
producing an empty section — then confirm a normally-schema'd component is byte-identical
through the change.

## What was shipped, 2026-08-05 — fix candidate 1, upgraded

**Status: FIXED IN CODE, INERT UNTIL THE NEXT CHASSIS ROLL, so this file stays OPEN.**
Go changes do nothing until an image is rebuilt and rolled; the bar for `bugs_closed/` is
fixed **and** live.

Candidate 1 was taken and **strengthened**, because as written it does not actually refuse.
For a schema-less component, "return no content found" means the template renders with
missing keys and `missingkey=zero` produces the **same blank section** — the defect with a
warning attached — and it throws away the payloads the storage seam would have decoded, so
the two seams would then disagree about one payload from one constructor.

What shipped instead: `normalizeRenderContentEnvelope`
(`platform/orchestration/actions/render_content_envelope_guard.go`) calls **`190`'s own
`normalizeContentDataEnvelope`, unchanged** — the same function, not a second implementation
— at the resolution point in `RenderComponentAction`, before `reconcileGeneratedItemKeys`
and `mergeIntoRenderContext`:

- not an envelope → returned unchanged, byte-identical (the dominant path);
- envelope, losslessly decodable (`clean` / `repaired`) → **the section renders its real
  content**, where today it renders a blank whose `content_data` the storage guard can only
  repair after the fact. The storage guard cannot reach back and fix `rendered_html`;
- envelope, not losslessly decodable → the render is **refused** with an error naming the
  component, the `content_field` and this bug.

Registered as the **third seam of PBP-032**, same commit. Council `dfb87f5e-6f01-42d4-8a01-6c59a4640c08`.

**Blast radius of a refusal, traced not assumed:** `shouldContinueLoopOnError` defaults false
(`loop_error_handler.go:70-90`) and the live `page-content-writer` sets `continue_on_error`
on neither `process_sections_loop` nor any substep — so a refused render fails the **whole
page**, not one section. That is already the disposition for the 855 rows the existing gate
covers, and with `190` live a REFUSE-tier payload fails the run anyway at the save, later and
after more LLM spend. If section-skip is ever wanted it is a one-key workflow-config change,
live immediately; deliberately not pre-empted here.

**Deliberately NOT shipped**, each with its backstop rather than a hope: guards on
`merge_with` (`current_section.resolved_data`) and on the `context_field` merge. Both are
Go/DB-constructed, the envelope has exactly one constructor which writes neither, and
`merge_with` keys land in `sectionContentData` and therefore still reach the storage guard.

**Candidate 2 remains OPEN and is the reason this file is not simply closed on the roll.**
This closes the *envelope* class only. A schema-less component whose step output is some
**other** kind of garbage still renders unchecked. Making the render gate speak for
schema-less components at all is a much larger decision and stays on its own merits.

### How to verify after the roll

1. Pod-grep **both** replicas for a string the change ADDED and one it did not:
   `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "normalizeRenderContentEnvelope"'`
   — and check the pod's start time in the same breath, because a guard that looks inert is
   indistinguishable from one running on the old image.
2. `SELECT action, context->>'outcome', count(*) FROM agent_error_log
    WHERE error_code='CONTENT_DATA_ENVELOPE' GROUP BY 1,2;` — **expect zero
   `render_component` rows** (trigger rate is zero); a `refused` row means the door was open
   and something walked through it, a `decoded` row means a section that would have rendered
   blank rendered real content.
3. The two assertions this file demanded are unit tests, run and mutation-checked:
   `TestRenderRefusesEnvelopeForSchemalessComponent` and
   `TestRenderNonEnvelopeContentByteIdentical` in `render_content_envelope_guard_test.go`.

## Related

- **`bugs_open/190`** — the storage half, fixed at both write seams, concept register PBP-032,
  council `09bc4b3d-6721-4479-85b8-b5b56bf9b5d7` APPROVED. Its guard's predicate
  (`isLLMTransportEnvelope`) is directly reusable here.
- `bugs_closed/088` — the same producer path, from the parse side.
- `json_envelope.go` — the header is the diagnosis of the whole class, including the nine
  live article bodies blanked by `missingkey=zero` when an envelope reached a template.
