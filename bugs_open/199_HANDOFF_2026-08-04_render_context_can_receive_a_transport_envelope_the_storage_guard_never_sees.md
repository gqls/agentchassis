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

**What stops it today, and why that is not the same as it being closed:** the render gate
(`missingRequiredLLMFields`) refuses when the component declares required `source:"llm"`
fields. A component with an **empty or optional-only input schema** declares nothing, so
there is nothing for that gate to find missing. That is the hole — and it is the *same* hole
shape as `missingkey=zero`, where a template rendered a blank rather than failing.

## Status of the claim — read this before acting

`[UNMEASURED]` — **whether any live component has an empty or optional-only input schema, and
whether this path has ever actually fired, has not been measured.** `190`'s lane named the
path from a code read and deliberately did not ship a change to it, because it alters render
behaviour for schema-less components and belongs on its own merits rather than inside a bug
patch (the 2026-07-28 platform-seams ruling). Do not restate the mechanism above as an
observed failure; it is a read of control flow with an unmeasured population.

**The first task is the census, and it decides whether this is a bug or a note.** Roughly:

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

## Related

- **`bugs_open/190`** — the storage half, fixed at both write seams, concept register PBP-032,
  council `09bc4b3d-6721-4479-85b8-b5b56bf9b5d7` APPROVED. Its guard's predicate
  (`isLLMTransportEnvelope`) is directly reusable here.
- `bugs_closed/088` — the same producer path, from the parse side.
- `json_envelope.go` — the header is the diagnosis of the whole class, including the nine
  live article bodies blanked by `missingkey=zero` when an envelope reached a template.
