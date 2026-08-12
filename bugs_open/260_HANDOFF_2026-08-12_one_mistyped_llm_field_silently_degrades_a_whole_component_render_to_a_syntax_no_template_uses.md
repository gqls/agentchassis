# 260 — One mistyped LLM field silently degrades a WHOLE component render to a regex path that no template on the estate uses, leaking Go control syntax into the page

**Filed 2026-08-12.** Status: **OPEN, root cause proven, no live damage, nothing fixed yet.**
Supersedes the open question in `brochure_component_library/HANDOFF_2026-08-12_fact_assignment_front_continue_here.md`
UPDATE-late ("the assembler lead named … Unowned"). That handoff's hypothesis is **refuted as
stated** and replaced by the mechanism below; its page-rebuild halt is **narrowed**, see §6.

## 1. Symptom

A page build is refused at `validate_content` with ~20 blockers, all `unrendered_template` /
`unrendered_template_block`. The assembled section HTML carries Go template control structures
verbatim — `{{if .eyebrow}}`, `{{range $s.branches}}`, `{{end}}` — **while the field values
inside those structures are correctly substituted**:

```
{{if .eyebrow}}<span class="mech-flow__eyebrow">The build flow</span>{{end}}
```

That combination — directives intact, fields resolved — is the whole fingerprint. It cannot be
produced by a writer emitting braces, and it is not a template-authoring error.

## 2. Root cause (proven, with a control)

`RenderTemplateReportingMissing` (`platform/orchestration/actions/component_library.go:965`)
executes the component template with Go's `text/template` via `executeGoTemplate`
(`platform/orchestration/actions/call_agent.go:1170`). **On ANY error it silently falls back to
a regex renderer** (`component_library.go:1010-1029`) built for **handlebars** syntax:

| fallback helper | recognises | line |
|---|---|---|
| `renderEachBlocks` | `{{#each x}}…{{/each}}` | `component_library.go:1517` |
| `renderIfBlocks` | `{{#if x}}…{{/if}}` | `component_library.go:1722` |
| `renderGoStyleSubstitutions` | `{{.field}}` — **substitutes these** | `component_library.go:1743` |
| `renderHandlebarsSubstitutions` | `{{field}}`; **skips anything containing a space** | `component_library.go:1761` |

Nothing in that chain handles Go's `{{if .x}}`, `{{range $s := .y}}` or `{{end}}`. So the
fallback resolves the scalar placeholders and leaves every control directive in the output —
exactly the fingerprint in §1.

**What triggers the error is a field TYPE violation, not a missing field.** Measured against
`missingkey=zero`, as the real code configures it:

| `steps` value | result |
|---|---|
| array of objects | renders |
| **key absent** | renders |
| **nil** | renders |
| **empty array `[]`** | renders |
| **a string** | `range can't iterate over …` → **fallback** |
| array of strings | `can't evaluate field title in type interface {}` → **fallback** |

So absence is safe and a wrong type is fatal — the opposite of the intuition that an empty
factless section is the dangerous case.

### The live instance

`mechanism-flow`'s schema correctly declares `steps[].branches` as an **array of objects**
(`{body, label}`); its template correctly does `{{range $s.branches}}`. The writer emitted
`branches` as a **prose string** on all four steps. Real template + real data, executed
locally:

```
A. real template + real data  → EXECUTE ERROR: template: component:116:20:
   executing "component" at <$s.branches>: range can't iterate over
   "Where a client already runs legacy APIs, this step adapts around them…"
B. CONTROL: branches coerced to the DECLARED array-of-objects, nothing else changed
   → OK, renders 8347 bytes, contains "{{": false
C. regex fallback over the same template → 16 leaked vars, 13 leaked blocks; first leak is
   byte-identical to the live run's stored page_content
```

B is what makes this a demonstration rather than an inference: **the single variable changed is
that field's type, and the failure vanishes.**

Evidence run: orchestration `07983216-929b-4494-8131-87c523058ea5` (fundamentallyai.com,
`production-backend-engineering`, 2026-08-12 13:06Z, `COMPLETED`/`complete_error`).
⚠ `orchestration_states` prunes ~24h — that row is gone by 2026-08-13. The durable copy is
`agent_error_log` (§4).

## 3. What this is NOT — two corrections to the record

- **REFUTED: "a field-substituting assembler that never executes `{{if}}`/`{{range}}`"**
  (the 08-12 handoff's lead). The assembler *does* execute Go templates, correctly. The leak is
  the **error path**, which is a different defect with a different fix: the bug is not that
  rendering is naive, it is that **failure is silent and degrades rather than stops**.
- **Seed 386 is exculpated, but not for the reason previously given.** The earlier argument was
  "386 adds one prose sentence with no braces, and the writer's LLM output is clean of `{{`".
  That test was void — the leak was never the model emitting braces, so clean model output is
  *consistent* with the defect, not exculpatory. The sound reason: the defect requires a nested
  field's **type** to be wrong, and 386 adds exactly one claim-restriction sentence
  ("Do NOT invent commitments, guarantees, warranties, or service promises…") that names no
  field, no shape, and not `branches`. `agent_definitions_bak_386` stays unused.
- **"New since 08-11 15:39" is better explained by exposure than by a change.** `mechanism-flow`
  is rare (5 stored sections estate-wide) and 08-11/08-12 were rebuild-heavy days. One
  `mechanism-flow` section stored a correctly-shaped `branches` array on 2026-07-28
  (oufe.com), so the writer can and does get it right; nothing establishes a regression point.
  **[INFERRED]** — stated as a competing explanation, not a finding.

## 4. Blast radius (measured 2026-08-12, each zero with a control)

- **Live damage: ZERO.** Of **1,452** stored `page_components`, **0** leak a control block and
  **1** leaks a `{{…}}` var — and that one is *not* this defect (webdesign.co.uk `ported-page`,
  legitimate prompt-library copy containing `{{TONE}}`/`{{COLOR}}`).
  The block regex was positive-controlled in the same query (`{{if .eyebrow}}`, `{{end}}`,
  `{{range $s := .steps}}` → true; plain HTML → false), so the zero is real and not a silent
  non-match. **Note this also refutes `bugs_open/203`'s claim that
  `idea.uk/tools/ab-test-calculator` stores a literal `{{.section_heading}}`** — no such row
  exists in `page_components` today.
- **The gate is why: `validate_content` refuses before persisting.** Every occurrence parked the
  page at `needs_human_review` with nothing written. The cost is wasted builds and a buried
  queue, not corruption.
- **Survivorship makes this invisible in stored data.** Of 5 stored `mechanism-flow` sections,
  4 omit `branches` entirely and 1 holds a proper array. A string `branches` is never stored
  *because* the gate catches it — so any census of stored `content_data` will report this defect
  as non-existent.
- **Recorded occurrences: 6 events, 4 domains, all identical** (`agent_error_log`,
  `error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'`): lendzy.co.uk 08-11 15:39 + 20:33,
  mortgagecalculator.co.uk 18:30 + 18:38, leopardessconsulting.co.uk 19:00,
  fundamentallyai.com 08-12 13:09. First of the 158 rows since 07-14 to carry this type.
- **Exposure: 33 components with a `{{range}}` have stored sections** (279 sections; `features`
  34, `faq` 28, `info-card-grid` 27, `differentiators` 23, `about-content` 20 …). Every one is
  exposed to the same degradation if a writer mistypes any ranged field.
- **NOTHING depends on the fallback.** Of **255** components, **0** use handlebars block syntax
  (`{{#`) and **0** use the fallback-only `{{nav_items_html}}`/`{{quick_links_html}}`
  placeholders. The fallback is a path that no template on the estate can be rendered by.
  This is the measurement `bugs_open/203` wanted when it called deleting the fallback
  "thinkable".

## 5. Fix candidates, ordered by what closes the door

1. **Make a template execution failure a hard error — delete the regex fallback.** §4 measures
   that no component needs it, and `RenderTemplateWithValidation` is already dead code (`203`),
   so `contextToMap`'s fabricated defaults die with it — closing `203`'s two remaining
   fabricate-a-URL members as a side effect. This makes the bad state unrepresentable: a
   mistyped field stops the build at the component with the real error
   (`range can't iterate over …`) instead of producing HTML that a downstream string-matcher has
   to catch. **Shared-mechanism change on rendering plumbing → council round + register entry.**
2. **Type-check `content_data` against `input_schema` before rendering.** The hook already
   exists: `missingRequiredLLMFields` (`platform/orchestration/actions/json_envelope.go:451`)
   checks **presence** only, and has exactly **one** production caller
   (`rerender_page_sections_action.go:333`) — so the page-**build** path has no schema gate at
   all. Extending it to validate declared types, and calling it on both paths, catches this
   class before any render. Reuse-first; complements (1) rather than competing.
3. **Coerce at the boundary** (a string where an array-of-objects is declared becomes a
   one-element array with the string as `body`). Cheapest, and it would have rendered this page
   correctly — but it silently rewrites writer output, so it hides the contract violation
   instead of surfacing it. Only as a companion to (2), never alone.
4. **Ask the writer to obey the schema.** Weakest — it makes correctness depend on an LLM
   getting a nested type right every time, with no check. Not a fix.

Also worth doing regardless: **`checkUnrenderedTemplates` caps each regex at 10 matches**
(`validate_page_content.go:793,804` — `FindAllString(html, 10)`), so "20 blockers" is the cap,
not a measurement (the real leak here was 29). Either raise the cap or report the true count;
as written, every instance of this defect reports an identical 9/1/9/1 signature regardless of
severity, which is what made four different domains look like one repeating event.

## 6. The page-rebuild halt: NARROWED, not lifted

The 08-12 handoff halted page rebuilds fleet-wide pending a cause. Cause is now established, so:

- **Rebuilding is SAFE** — no corruption is possible; `validate_content` refuses before
  persisting, proven by the 0/1,452 scan in §4 and by all 6 occurrences leaving nothing written.
- **Rebuilding a page whose plan includes a component that `{{range}}`s over a writer-supplied
  array may still be REFUSED** and will park the page at `needs_human_review`, wasting a build.
  33 components qualify (§4).
- **So: rebuild freely, but check the outcome**, and if it lands at `needs_human_review` with
  `unrendered_template` blockers, this is the cause — do not re-dispatch, and do not read it as
  a new defect. Until candidate 1 or 2 ships, a mistyped nested array will keep doing this.

## 7. How to verify a fix

- **Regression, offline:** the harness in §2 — real `mechanism-flow` template, real
  `content_data` with a string `branches`. Before: execute error → leaked HTML. After
  candidate 1: hard failure naming the field. After candidate 2: refusal at the schema gate,
  before any render. **Keep case B (coerced to the declared shape) as the control** — a fix that
  makes both cases fail is not a fix.
- **Live:** rebuild `fundamentallyai.com/production-backend-engineering` (site
  `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`, work item `7824d5ab-7531-400d-bee8-65ba6be870fe`)
  and read the page row + `page_components.updated_at`, not the work-item status.
- **Estate sweep** (should stay at 0 leaked blocks, with the positive control in the same query):
  `SELECT count(*) FILTER (WHERE rendered_html ~ '\{\{[\s]*(range|if|end|with|else)[\s\}]') FROM page_components;`
- **Detection is still absent:** `bugs_open/149` §B1 — the registered `unrendered_templates`
  discovery check is configured in no agent and has never run, so nothing sweeps pages that
  already serve this shape. That remains the detection half.

## 8. Diagnosis-loop provenance (owner ruling 2026-07-31 — declared substitution)

This file asserts a cross-cutting root cause in shared rendering plumbing, so it owes the `090`
loop or a stated substitution. Both apply:

- **A `090` was already filed on this symptom** by the previous session:
  `RUN_CORRELATION_ID=b885a92e-d308-4b9c-99ee-306ca2f6b373`. It completed in 5 iterations and
  **produced no locatable verdict** (no non-`bundle` artifact on the correlation, `final_result`
  NULL, no `doc_notes` row, no verdict language in the final bundle). Its bundle's in-scope list
  is what pointed at the assembler, and that lead is refuted in §3.
- **Substitution declared:** first-hand verification with an isolating control — the real
  template and the real writer output, one variable changed (that field's type), failure
  reproduced and then made to vanish (§2 A/B/C); every population claim in §4 carries a
  positive control in the same query. The loop's own read-out step also mis-attributed the
  symptom (it tested whether the blocker rows belong to `page-content-writer` runs and got 0
  rows — `validate_content` is a `page-build-handler` step), so re-running it on the same
  symptom text would likely land in the same place. **Where the loop's conclusion is supposed to
  be written remains an open defect in its own right** and is worth more than this bug.
