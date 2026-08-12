# 260 — One mistyped LLM field silently degrades a WHOLE component render to a regex path that no template on the estate uses, leaking Go control syntax into the page

**Filed 2026-08-12.** Status: **OPEN, root cause proven, no live damage, nothing fixed yet.**
Supersedes the open question in `brochure_component_library/HANDOFF_2026-08-12_fact_assignment_front_continue_here.md`
UPDATE-late ("the assembler lead named … Unowned"). That handoff's hypothesis is **refuted as
stated** and replaced by the mechanism below; its page-rebuild halt is **narrowed**, see §6.

**SPLIT OF OWNERSHIP (owner direction 2026-08-12 — language work lives in ONE thread).** This bug
has two halves and they belong to different lanes:
- **The renderer half — the silent fallback — stays here.** Shared rendering plumbing; candidate 1
  below; wants its own council round.
- **The writer-output half — that an LLM violated its component's declared field shape and nothing
  checked — is handed to `copy_quality_two_stage`:**
  `docs024_key_docs_latest/copy_quality_two_stage/CONTRIB_2026-08-12b_a_json_schema_is_also_just_an_instruction_and_stage_2_inherits_the_hazard.md`.
  That lane already ruled that set preservation "is not achievable by instruction at all — prose or
  data — and must be a mechanical check plus a repair step"; this case is the same finding at a
  formal JSON Schema, so it refutes the "then give the writer the schema" remedy before anyone
  spends a round on it. **Candidate 2 (type-check `content_data` against `input_schema`) is
  therefore theirs, not mine** — and its third caller is their stage-2 executor, which supplies
  agent-written `content_data` and re-renders through this same fallback
  (`section_editor_actions.go:805,895`) on pages that are already live.

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
  component **`tool-blueprint-compiler`**, legitimate prompt-library copy containing
  `{{TONE}}`/`{{COLOR}}`. **If anyone tightens the detector to a bare `{{`, this row is its first
  false positive** — component name supplied by the `copy_quality_two_stage` lane, 2026-08-12).
  > **DEMAND CONTROL, added 2026-08-12 (supplied by the `copy_quality_two_stage` lane — the zero
  > was uncontrolled on the axis that matters).** On its own, "0 stored components carry the
  > damage" is equally consistent with **"the exercising path never ran"**. It did run:
  > `section-editor` has **132 orchestrations, all COMPLETED, most recent today**, against
  > **0 of 1,454** stored components carrying `{{if|range|end|with}}`. So this is a real negative
  > — "the guards hold", not "nothing has tried". This is the distinction that makes the zero
  > evidence rather than an artefact.
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
   > **⚠ CORRECTED 2026-08-12 (the `copy_quality_two_stage` lane's CONTRIB_…12c caught this, and
   > it would have shipped inert).** As first written, "against `input_schema`" implied the JSON
   > Schema shape `mechanism-flow` uses. **The library is overwhelmingly NOT that shape.** Of 255
   > components: **4** use `properties`, **164** use the house v2 dialect
   > `{"fields": {<name>: {"type": …}}}`, **87** declare neither. A gate written against
   > `input_schema->'properties'` would cover **4 components and report a clean sweep over the
   > other 251** — armed-but-inert, looking exactly like success. **I generalised from the one
   > component in front of me, and it is the unrepresentative one.**
   > **Their prescription needs one refinement, though: do NOT write the gate against the house
   > dialect either.** A `fields`-only gate is blind to the 4 `properties` components — including
   > `mechanism-flow`, the only component with a proven live failure. Call
   > **`datahelpers.SchemaContentFields`** (`platform/orchestration/datahelpers/component_schema_fields.go:58`),
   > which already normalises BOTH dialects onto the v2 field shape and preserves `items` /
   > `min_items`, so a nested array-of-objects check is expressible. It is also already the
   > helper the estate uses for exactly this question.
   > **And match `type IN ('array','list')`** — `list` is a fifth declared type (5 fields across
   > 2 components, 0 of them `source: llm`); an `'array'`-only filter never sees them.
   > **Coverage honesty:** the **87** schema-less components are gated by nothing whatever this
   > candidate does, so a green result from it is NOT fleet coverage and must not be reported as
   > one.
   > **AND THE LIMIT OF MY OWN PRESCRIPTION, added 2026-08-12 (the copy lane's second addendum
   > named it; verified here).** "Call `SchemaContentFields`" is the right route for the *dialect*,
   > but a gate built on it can only ever enforce what that helper carries forward. Its copy list
   > is exactly six keys — `source`, `on_missing`, `fallback`, `missing_reason`, `items`,
   > `min_items` (`component_schema_fields.go:87`) — plus `type`, `llm_guidance`/`description`,
   > and a `minItems`→`min_items` remap at `:92-96`. **Everything else in a legacy property is
   > dropped silently, and the asymmetry is unsignalled: `minItems` is remapped, `maxItems` does
   > not appear anywhere in the file.** So `items` and `min_items` are enforceable (which is all
   > the array-of-objects check in this bug needs), but any other JSON Schema constraint —
   > `maxItems`, `enum`, `pattern` — is invisible to a gate written this way, for legacy
   > components only. Do not let "SchemaContentFields is the dialect-safe route" harden into
   > "SchemaContentFields sees the schema": it sees six keys of it.
   > **This result also has an expiry date**, and it is the lane's point rather than mine: the
   > "nothing is dropped" measurement in `bugs_open/265` holds for what **four** components
   > declare **today**, and arrivals are still landing (newest 2026-08-10). Re-run the key
   > enumeration before trusting it again.
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

---

## 9. ADDENDUM 2026-08-12 — the exchange with `copy_quality_two_stage`, and one adjacent defect it exposed

Their reply is `copy_quality_two_stage/CONTRIB_2026-08-12c`. Three of their numbers are folded in
above (candidate 2's dialect correction, §4's demand control, the benign `{{` row's component).
Two further things came out of verifying them.

### 9a. Sizing the acute set — better than my "33 ranging components"

My §4 exposure figure counted components whose *template* ranges. The sharper population is
components that **declare** an array field the **writer** is told to author, because that is the
set where a type violation is reachable at all. Measured over all 255, house dialect:

| declared type | fields | `source: llm` | components |
|---|---|---|---|
| `array` | 63 | **13** | 49 |
| `list` | 5 | 0 | 2 |

**The 13 `source: llm` array fields are the acute set** — declared as arrays, authored by the
writer, checked by nothing. `mechanism-flow` is one of them and is **not** a special case. (The
lane measured 49/12 over 191 *active* components; the delta is population, not method — and
`list` accounts for a fifth type neither of us filtered for initially.)

### 9b. ⚠ ADJACENT DEFECT — an "extinct" schema dialect has been reintroduced four times, and its tripwire fires into logs nobody reads

`SchemaContentFields`'s own doc comment
(`platform/orchestration/datahelpers/component_schema_fields.go:53-56`) states:

> *"the legacy dialect is extinct fleet-wide (0 of 173 as at 2026-07-21), so a true here means a
> regression reintroduced it"*

**There are 4 today, and every one was created AFTER that census:** `report-dossier` (created
07-27), `mechanism-flow` (07-28), `evidence-timeseries` (07-28), `loans-consolidation` (created
**2026-08-10**). So the reintroduction is not a one-off — it is ongoing, most recently two days
before this filing.

The platform anticipated exactly this: `WarnIfLegacyDialect` is wired into the render/rerender
paths for it. **It fires at `Warn`, and I measured that the entire `RenderTemplate` log family is
absent from a 4,661-line 24h window on `agent-chassis`** (§7 of this file) — so the tripwire has
been firing into a channel with a shelf life of hours. **A detector whose only output is a Warn
on a busy service is not a detector.** (The comment cites `bugs_open/026`; that file is CLOSED and
is about `news-listing` hardcoding English — a different case, so the citation is stale and should
not be followed as prior art.)

This is not the same bug as the fallback, and I have not filed it separately — bug numbers were
colliding at a rate of one per few minutes on 2026-08-12 (my own 260 was claimed by another
session 40s before I filed). Whoever picks it up: the component-creator path is the likely
producer, and the check is one query —
`SELECT function, created_at FROM content_components WHERE input_schema ? 'properties';`

### 9c. Their stage-2 design constraint is TOO STRONG — `field_updates` does not put the hazard structurally out of reach

The lane adopted, as a hard constraint, *"prefer `field_updates`, which puts the hazard
structurally out of reach, over a gate"*. Read against the code, that is true only of fields the
editor does **not** name:

- Merge mode is `for k, v := range updates { existingContentData[k] = v }`
  (`section_editor_actions.go:746-748`). **Any field the agent names is overwritten with whatever
  it supplies** — so `field_updates: {"steps": "…prose…"}` retypes `steps` exactly as a full
  replacement would.
- The code comment they are reading (`:403-405`) says a merge *"carries type and result forward
  **untouched**"* — `untouched` is the operative word, and it means untouched fields, not the
  edited one.
- **For this hazard that distinction collapses**, because the field a readability pass edits *is*
  the array field. Preferring `field_updates` narrows the blast radius from "every field in the
  component" to "the fields being edited" — worth having, but it is a mitigation, **not a
  structural guarantee, and not a substitute for the gate.**

### 9d. Their open question answered: the merge/replace mode is NOT recoverable from the DB

They asked for a way to discriminate full-replace from field-update runs, having found that
matching `collected_data` for the two key names returns 132/132 (the action's config echo carries
both names regardless of use). The answer is that it cannot be done as the code stands:

- The action's returned map (`section_editor_actions.go:504-514`) records `edit_type`
  (`content_edit` / `component_swap`) and **not** the merge-vs-replace mode.
- The mode exists only as two `logger.Info` lines (`:750`, `:758`, `:766`) — and those rotate.
- Reading the *input value* rather than the key name is the only live route, and it has a trap the
  code documents at `:729-732`: `ExtractActionInputs` does a **nested** lookup, and
  `content_data` is aliased to `replacement_content_data` (`:60`), so it can false-positive on
  `site_record.content_data` (the site plan). That is why `field_updates` is checked FIRST, and
  why any measurement must apply the same precedence: **merge iff a non-empty `field_updates`
  value resolves; only then consider replacement.**
- **One-line fix if the split is needed repeatedly** (it is, for their Phase 4): add the mode and
  the two field counts to the returned map, which lands in `collected_data` and makes the question
  answerable retrospectively for ever after.
