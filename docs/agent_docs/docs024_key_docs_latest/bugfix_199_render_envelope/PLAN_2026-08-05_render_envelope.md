# PLAN — `bugs_open/199`, the render-seam transport-envelope guard

**Opened 2026-08-05.** Lane picked up from `bugs_open/199`, which was filed 2026-08-04 by the
`190` lane at the council's explicit request (the `bug_historian` seat objected that naming
this path in a rationale field and calling it "separable" was *"a silent decision"*).

## The problem, in one paragraph

When an LLM step's reply will not parse as JSON, `ExecuteLLMPromptAction` returns the transport
envelope `{"type":"text","result":"<raw reply>"}`. `bugs_open/190` closed the two **storage**
seams — nothing can persist that shape into `page_components.content_data`. Nothing sits on the
path into **render** context. `RenderComponentAction` resolves a component's content through
`extractContentWithFallbacks` and, when the step fell to the text path, the map it gets back
*is* the envelope. The render gate that could catch it (`missingRequiredLLMFields`) only fires
for a component whose `input_schema` declares a required `source:"llm"` field.

## Decisions, and their reasons

### D1 — The census comes first, because the bug file said so and it was right

The file filed its own claim as `[UNMEASURED]` and said plainly: *"The first task is the census,
and it decides whether this is a bug or a note."* It could have come out a note. It did not.

**Result: BUG.** 315 of 1212 live `page_components` rows (26%) sit on gate-blind components, and
the one envelope still live in `content_data` sits on one of them. Full figures in NOTES and in
the bug file.

**But the trigger rate is zero** (0 of 62 `generated_content` outputs envelope-shaped in the 25h
window, against 111 for `compose_note` in the same query). That is recorded everywhere the fix
is recorded, because it changes how the post-roll check must be read: a count of zero is the
*expected* reading and proves nothing on its own.

### D2 — Reuse `190`'s normaliser; do not write a second implementation

`normalizeRenderContentEnvelope` calls `normalizeContentDataEnvelope` — the same function.
One payload, one policy, three seams. Two implementations of one rule is the drift class this
codebase keeps being bitten by, and the storage half is already council-APPROVED (PBP-032).

### D3 — Decode when lossless, refuse when not — NOT "refuse everything"

Considered and rejected. This seam is **upstream** of the storage seam on the only live path
(`result["content_data"]` becomes the `sections_metadata` the save stores). Refusing everything
would mean a payload the council approved the storage seam to *decode* never reaches the branch
approved to decode it — the two seams would answer differently about one payload from one
constructor.

Decoding here also buys something no storage guard can: **the section renders its real content.**
Today an envelope renders through `missingkey=zero` as a blank, and the storage guard repairs
only the column — it cannot reach back and fix `rendered_html` that has already shipped.

### D4 — Reject the bug file's own candidate 1 as written

Candidate 1 was *"return no content found"*. For a schema-less component that is **not a
refusal**: nil content means the template renders with missing keys and `missingkey=zero`
produces the same blank section. It is the defect with a warning attached, and it discards the
recoverable class as well.

### D5 — Guard the CALLER, not the resolver

`extractContentWithFallbacks` leaks by **two** branches (see the correction below). One call at
its single caller covers both and cannot be reopened by someone "fixing" one of them. It also
keeps `190`'s stated deferral intact — that lane declined to change the shared resolver.

### D6 — Council gate, no RFC, no opt-in field

Per the 2026-07-29 ruling point 1, an RFC is needed when what the shared mechanism *guarantees*
changes. `RenderComponentAction`'s stated guarantee is **already refusal** ("Fail loud rather
than ship a silently-empty section", `v3_site_actions.go:1834`). This does not invert it; it
makes it true for the payload class the existing gate is structurally blind to.

Per RFC_010 (2026-08-02), an opt-in field is required when a seam's widest branch is licensed by
*"callers must all be X"* — caller identity. This refusal is licensed by the **payload**. That
argument is inherited verbatim from PBP-032, which carried it through council APPROVED.

Submitted: `dfb87f5e-6f01-42d4-8a01-6c59a4640c08`.

## Corrections to the originating brief

> **CORRECTION 1 — the bug file names the wrong branch.** It says the "last resort" branch of
> `extractContentWithFallbacks` returns the envelope. It is the **first fallback loop**
> (`:4494-4503`), which has no content check at all — just `len(m) > 0`. `hasContentFields`
> never gets a say. The last-resort branch *is* a second real leak (a superset envelope passes
> `hasContentFields` on its `content` key), which is why the fix sits at the caller. Corrected
> in the bug file, the `016b` §10 row, PBP-032 and `LANDMINES.md`.

> **CORRECTION 2 — do not copy the save seam's identity paths.** `writeContentDataEnvelopeLog`
> reads `site_record.site_id` / `current_page.name`; both resolve **0/110** inside the page
> writer's own runs. `renderEnvelopeIdentity` uses the measured chains instead.

## What is deliberately NOT in scope

- **Candidate 2** (make the render gate speak for schema-less components at all). This closes the
  *envelope* class only; other garbage still renders unchecked. Larger decision, stays open.
- **Guards on `merge_with` and the `context_field` merge.** Go/DB-constructed, the envelope has
  one constructor that writes neither, measured zero matches — and both are backstopped by the
  storage guard anyway.
- **Section-skip instead of page-fail on a refusal.** A one-key workflow-config change, live
  immediately, and the owner's call rather than this lane's.

## Phasing

1. ~~Census~~ **done** — decided bug, not note.
2. ~~Plan (fable) and implement (opus)~~ **done** — guard + 4 tests, all four named mutations run.
3. ~~Council submission, register entry, docs, commit~~ **done 2026-08-05.**
4. **OPEN: the roll.** Go changes are inert until an image is rebuilt and rolled. `bugs_open/199`
   stays open until pod-grep proves the guard live on both replicas.
