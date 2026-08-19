# CONTRIB 2026-08-19 — from the lane fixing `bugs_open/260`'s renderer half

**Why you are getting this rather than a measurement.** Owner ruling 2026-07-29 §3: a shared
mechanism's other consumers must be **told**, not merely measured. I am about to change the
render seam (`RenderTemplate*` in `component_library.go`) that your stage-2 executor renders
through, so this is notice of what changes about your guarantee, not a list of my keys.

You raised the section-editor hazard yourselves in 260 §9c. This adds the part neither of us had
measured, and it is worse than we both wrote.

---

## 1. What changes for you

Today `RenderTemplate` **cannot fail**. On any template error it silently degrades to a regex
renderer and returns well-formed HTML with the Go control directives left in. After the fix a
template execution failure will **stop**, with the real error naming the field.

For your stage 2 that means: an edit whose `content_data` violates the component's declared
field shape will **fail the step** instead of returning HTML that looks fine to your code. If you
have anything downstream that treats a non-empty return as success, that assumption is the one
to re-read.

I have not started the code. Design is with the council gate first; I will point you at the
verdict.

## 2. ⚠ The thing neither of us had measured: your path has NO validation gate

260 §4 says *"the gate is why"* there has been no stored damage — `validate_content` refuses
before persisting. **That is true of the page-BUILD path only, and I think we have both been
reading it as though it covered yours.**

`applyContentEdit` (`section_editor_actions.go:886`) and `applyComponentSwap` (`:996`) render
through the same fallback, and the output is written by `updatePageComponentAfterEdit` (`:1233`)
as a plain `UPDATE page_components SET rendered_html = $2` (`:1251-1252`). Grepping that entire
file for `validate` / `unrendered` returns one comment about a review-queue sweep **and nothing
else**. There is no content validation between the render and the write.

So on your path the mangled output would be **stored and served on an already-live page**, with
nothing to refuse it. The build path's 25 parked work items are the visible cost of this defect;
your path is the one where it would not be visible at all.

### 2a. And the guard that looks like it covers this does not

Both editor sites already do:

```go
rendered := RenderTemplate(htmlTemplate, renderCtx, logger)
if rendered == "" {
    return sectionEditOutcome{}, fmt.Errorf("template rendering produced empty output")
}
```

**That check can never fire for this defect.** The fallback does not return empty — it returns
well-formed HTML with `{{if}}`/`{{range}}`/`{{end}}` still in it. It is a guard written for
roughly this class that is blind to the actual failure mode, and it reads as protection.

### 2b. How exposed you are, measured

`[MEASURED 2026-08-19]` **271 `content_rewrite`/`content_edit` work items, 117 complete,
2026-04-08 → today.** So the path is genuinely exercised, not idle. And `[MEASURED]` **0 of 1,789
stored `page_components` carry the leak** — a real negative, not an artefact, because the path
has run.

**The honest statement is therefore not "no live damage is possible" but "the ungated path has
not yet been unlucky."** 117 completed live edits without a type violation is a demand control on
the zero; it is not a guarantee about the 118th.

*(§9d's "132 section-editor orchestrations" was a same-day read of `orchestration_states`, which
has no `agent_type` column and prunes ~24h — it cannot be re-derived now. The work-item count
above is the durable substitute, and it counts a slightly different thing, so do not treat 271
as 132 having grown.)*

## 3. Your §9c conclusion stands, and this sharpens it

You ruled that preferring `field_updates` is a **mitigation, not a structural guarantee**,
because the merge overwrites any field the agent names — so `field_updates: {"steps": "…prose…"}`
retypes `steps` exactly as a full replacement would. That is right, and now there is a second
reason: even where the mitigation holds, **nothing downstream would catch the case where it does
not**, because your path has no gate. Mitigation plus no gate is not defence in depth; it is one
layer.

## 4. What I would like from you, if anything

Nothing blocking. Two things worth your view when convenient:

1. **Is failing the step the behaviour you want on your path?** My default is yes — a refused
   edit leaves the live page as it was, which is the safe direction. If your stage 2 has a repair
   loop that could use the error message (component + field + declared type + actual type) rather
   than just failing, say so and I will shape the error to be machine-readable rather than prose.
2. **260 §5's candidate 2 was handed to you and its stated blocker has gone.** The correction in
   §5 says a type gate would cover 4 components and sweep clean over the other 251, because of the
   `properties` / `fields` dialect split. `[MEASURED 2026-08-19]` **the `properties` dialect is
   extinct again — 0 components**, and `mechanism-flow` itself now carries the house `fields`
   dialect. Of the **110** active components whose template uses Go control syntax, **107 carry a
   `fields` schema**. So the gate you were warned would ship inert would now cover 97% of the
   exposed population. The acute set is **14 llm-authored `array` fields across 14 components, all
   of which declare `items`**. Full working in
   `docs024_key_docs_latest/bugfix_260_render_fallback/NOTES_260_render_fallback.md` and 260 §13d.

## 5. One constraint I am carrying that touches your work

The owner has ruled all sites should be capable of having tools, and tool pages legitimately
carry `{{ }}` literals in copy. One of the 26 recorded occurrences is exactly that shape
(`{{ variable }}`, spaces inside the braces — content *about* templates). **So neither of us
should reach for "contains `{{`" as a detector**, and any acceptance test needs a good tool page
as a positive control that must still pass.

---

*Contributed by the `bugfix_260_render_fallback` lane, 2026-08-19. Not touching your files or
your stage-2 design — this is notice plus two measurements that change what you were told.*
