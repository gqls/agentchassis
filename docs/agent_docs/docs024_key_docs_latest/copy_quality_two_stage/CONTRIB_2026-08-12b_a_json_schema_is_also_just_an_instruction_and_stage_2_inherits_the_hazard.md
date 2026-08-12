# CONTRIB 2026-08-12b — a JSON Schema is also just an instruction, and stage 2 inherits the hazard on already-live pages

**From:** the `brochure_component_library` fact-assignment front, filing `bugs_open/260`.
**Why it is here:** the owner's direction that **language work lives in ONE thread** — this
lane. `260`'s root cause is a rendering defect and stays with me; its **writer-output half is
handed to you** and I am not working it. Nothing below needs a reply to be useful; §2 is the
part that could save you a round.

---

## 1. What happened, in one paragraph

A page build was refused because a component's HTML came out carrying Go template directives
verbatim (`{{if …}}`, `{{range …}}`, `{{end}}`) with the words correctly filled in around them.
Cause: the renderer executes Go `text/template` and, **on any error, silently falls back to a
regex renderer written for a different dialect**, which substitutes `{{.field}}` but leaves every
control directive in the page. What made it error: `mechanism-flow`'s `steps[].branches` is
declared in the component's `input_schema` as an **array of objects** (`{body, label}`), and
`page-content-writer` emitted it as a **prose string**. Asking `range` to iterate a string is an
error, so the whole component fell to the fallback. Proven with an isolating control: coerce that
one field to its declared shape, change nothing else, and it renders clean.

Full evidence, commands and fix candidates: `bugs_open/260`. Occurrences: 6 builds, 4 domains,
2026-08-11 15:39 → 2026-08-12 13:09. **No live page ever carried it** — 0 of 1,452 stored
components — because `validate_content` refuses before persisting.

## 2. This is your set-preservation ruling again, at the strongest instruction that exists — so the obvious remedy is already refuted

Your `NOTES` (round 7) states it as well as it can be stated:

> *"Set preservation is not achievable by instruction at all — prose or data — and must be a
> mechanical check plus a repair step."*

You reached that over rounds 4, 6 and 7 with three different instructions and three different
losses, and you generalised past prose to "data". **My case is the same finding at one further
rung, and it matters because it closes the escape hatch a reader of your NOTES will reach for
next.** The writer here was not given a prose instruction, nor an enumerated list, nor an example.
It was given a **formal JSON Schema** declaring `type: array`, `items: {type: object, required:
[body]}` — the most machine-readable form of "this is a list of objects" the platform has — and it
emitted a sentence. Nothing checked, and the copy read perfectly well; the schema violation was
invisible in the prose.

So: **"we tried prose, now let us hand the writer the schema" is not a round worth spending.**
It has been run, on a live component, and it failed. The conclusion is exactly yours, unweakened:
the check must be at the boundary, not in the prompt.

**Two ways my instance is stronger than a lost-set instance, both of which argue for building the
gate sooner:**

1. **The cost is not degradation, it is destruction.** Losing 5 of 13 links leaves a worse page.
   One mistyped field takes out the **entire component** — every other field in it, correctly
   written, is discarded along with the structure.
2. **The failure surfaces as something else entirely.** It arrives as template gibberish at a
   downstream string-matching gate, so the first three theories are all about the renderer. It
   cost this front a day and a diagnosis run that produced no verdict before anyone looked at the
   writer's output shape. A copy defect that presents as a rendering defect will not be routed to
   you.

## 3. ⚠ The part that is about YOUR design: stage 2 inherits this hazard and raises the stakes

Verified in code today, not inferred:

- `section-editor` — your stage-2 executor (`PLAN` §"stage 2 … executes via section-editor") —
  takes `replacement_content_data`, which its own source comments call **"agent-supplied"**
  (`platform/orchestration/actions/section_editor_actions.go:405`), writes it to
  `page_components.content_data`, and then **re-renders the component template from it**
  (`:219-229` describes the contract; the calls are `RenderTemplate` at **`:805`** and
  **`:895`**).
- `RenderTemplate` is the function carrying the silent fallback
  (`platform/orchestration/actions/component_library.go:952` → `:965`).

**So stage 2 is on the same path, and it runs on pages that are already live.** Stage 1's version
of this defect is caught by `validate_content` before anything persists — that is the only reason
`260` has no live damage. I have **not** established that the `section-editor` path has an
equivalent gate in front of its write, and you should not assume it does; that is the one check I
would run before building stage 2's writer.

**Why a readability pass is unusually likely to trip it.** The failure mode is a nested array
flattened into prose. That is not a random error — it is *what a readability rewrite is for*.
"Three bullet outcomes read as a list; make them flow" produces exactly one sentence where an
array of objects belongs. Your stage 2 is being asked, in its own remit, to do the thing that
breaks this.

**And D2 does not cover it.** Queueing stage 2's output for human review protects against bad
copy. This failure happens at **render**, which is downstream of the content decision and
upstream of anything a human would read — a reviewer would be shown either a refused page or a
broken one, and neither presents as a copy question.

**Concretely, for the gate you are already building:** your `gate_page_links.py` asserts a
page's own declared `required_links` appear as hrefs. The type-shape analogue is the same shape of
check against a different declaration — assert `content_data` conforms to the component's own
`input_schema`, which like your required-set "cannot drift from the brief" because it is the
component's own declaration.

## 4. The hook already exists, is the wrong shape, and is on only one of the two paths

`missingRequiredLLMFields` (`platform/orchestration/actions/json_envelope.go:451`) checks writer
output against `input_schema` — but for **presence only**, never type. It has exactly **one**
production caller (`rerender_page_sections_action.go:333`); every other reference is a test. So
the page-**build** path has no schema gate at all.

That makes "extend it to validate declared types, and call it on every path that accepts
agent-supplied `content_data`" a reuse-first change rather than new machinery — and the third
caller it wants is `ApplySectionEditAction`, i.e. **your stage 2**. Listed as candidate 2 in
`bugs_open/260` §5. **I am not building it**; if this lane owns the writer-output contract, it is
better placed here than in a rendering bug, and it wants a council round either way.

## 5. Two things about `page-content-writer` you should know, since it is your agent

- **Seed 386 landed 2026-08-11 12:36Z** and is live: STRICT RULE 5 now also bans invented
  **commitments** — *"Do NOT invent commitments, guarantees, warranties, or service promises
  (response times, refunds, availability) not stated in this prompt."* It is a constraint on any
  stage-1 prompt work you do. `Council-Reviewed: d1e8c36e`.
- **386 was a suspect for `260` and is exculpated — but the reasoning first offered was void, and
  it is the kind of reasoning worth not repeating.** The original argument was "386 adds one prose
  sentence containing no braces, and the writer's LLM output is clean of `{{`". That tests nothing:
  the leak was never the model emitting braces, so clean model output is *consistent* with the
  defect rather than exculpatory. The sound reason is mechanism — the defect requires a nested
  field's **type** to be wrong, and 386 names no field, no shape and not `branches`. Its rollback
  (`agent_definitions_bak_386`) is verified and **unused**; do not spend it on a guess.

## 6. Two measurement traps, because this lane builds gates and reads their counts

Both are now in `LANDMINES.md`; both bit me in this diagnosis, and your NOTES already records the
same class ("a green run from the broken version would have been exactly as convincing as its
false red").

- **`validate_content`'s blocker count is a cap, not a measurement.** `checkUnrenderedTemplates`
  calls `FindAllString(html, 10)` on each of two regexes, so a refusal saturates at "20 blockers"
  regardless of severity — the real leak here was 29. **I built a component fingerprint out of
  those counts and it confidently excluded the component that was actually at fault.** Six events
  across four domains showing an identical signature meant only "both regexes saturated".
- **My estate-wide "no live damage" scan nearly shipped with a control that controlled nothing.**
  I reached for `bugs_open/203`'s cited literal (`{{.section_heading}}` on
  `idea.uk/tools/ab-test-calculator`) as my known-present case — **that row does not exist today**,
  so my zero was unguarded. Fixed by controlling the regex itself in the same query. Worth knowing
  if you cite 203: its census figures are from 08-05 and at least that one has expired.

## 7. What I am NOT handing over

The **fallback itself** stays with `bugs_open/260`: deleting it is a change to shared rendering
plumbing, wants its own council round, and is not a language question. The measurement that was
missing for that argument now exists — **0 of 255 components use the handlebars dialect the
fallback understands, and 0 use its `{{nav_items_html}}`/`{{quick_links_html}}` placeholders** — so
it is a path nothing on the estate can be rendered by. If it is deleted, a mistyped field becomes a
hard error naming the field, and everything in §2 and §3 becomes much cheaper to detect. That is
the ordering I would prefer, and it is not mine to schedule.
