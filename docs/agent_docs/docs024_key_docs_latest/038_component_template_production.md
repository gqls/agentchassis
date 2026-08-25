# 038 — Component template production: how a writer's output becomes a stored, renderable component

**Reference doc, written 2026-08-25** from the `bugs_open/345` lane, which spent a week inside this
pipeline. It describes the path from "a page needs a section type nothing yet renders" to "a
component row that pages can bind content to", the two contracts the writer must satisfy, the
birth gate that enforces them, and the failure→feedback→retry→terminate loop around it.

Grounded in code at HEAD as of writing; every file:line is a pointer to verify, not a quote to
trust — this pipeline changes, so **read the function before relying on a line number here.** Where
a count appears it carries its date; re-measure before repeating it.

---

## 1. What a component IS, in plain terms

A **component** is a reusable section renderer: an HTML **template** with `{{.placeholder}}` slots,
plus an **input_schema** naming those slots and where each one's content comes from. A **page** does
not carry HTML; it carries **content_data** — values keyed by the schema's field names — and the
renderer fills the template's slots from it. One component serves every page that uses its section
type, so its field names are a shared contract: rename a field and every page keyed on the old name
renders that slot empty.

Two artefacts, then, and they must agree:

- **html_template** — the markup, with `{{.field_name}}` (and `{{placeholder "field_name"}}`) tokens.
- **input_schema** — the house dialect `{"fields": {"<name>": {"source", "required", "type", ...}}}`.
  `source` says where a field's value resolves from (a `site_specs.<aspect>` path, a `query.*`
  listing, or a literal/label default). The retired JSON-Schema dialect (top-level `properties`) is
  refused by a DB CHECK constraint and by the gate.

The component is written by the **`component-creator`** agent, dispatched for a
`needs_new_component` work item by **`build-dispatch-loop`**. The generation step is
`generate_template` (an LLM call, `execute_llm_prompt`); the persistence step is `store_component`
(action `store_generated_component`).

---

## 2. The two contracts the writer must satisfy

These are the two the birth gate was built to enforce (`bugs_closed/337` gave the writer the prompt
text for both, from the gate's own vocabulary):

### 2a. The field-name contract (schema ↔ template must be in sync)

Every `{{.field}}` in the template must have a schema entry, and — the half that matters for this
doc — every schema field should have a `{{.field}}` in the template. `scoreComponent`
(`compute_component_quality.go`) computes `SchemaTemplateSynced` and lists each offender in
`QualityIssues`, in two directions:

- **Unknown template variable** — `template var {{.x}} has no schema entry`. The template renders a
  slot the schema never declared, so the renderer has no content source for it. **This is a real
  defect: the slot renders empty on every page.**
- **Orphan schema field** — `schema field "x" has no template variable`. The schema declares a slot
  the template never renders. **This is harmless to the page** — the declared field is simply never
  used — but it wastes the content generation that fills it. Scored `warning`, not `error`.

Both directions are now reported **exhaustively and in sorted order** (they used to short-circuit on
the first offender, and one direction iterated a Go map non-deterministically — `bugs_open/345`
made them exhaustive because the rejection message that names them feeds a byte-identical repeat
detector; see §4).

### 2b. The source-vocabulary contract (a declared source must resolve)

A field's `source` must be something the platform can actually resolve. A `site_specs.<aspect>`
source names an aspect that must exist in the live `site_specs` vocabulary; a `query.*` source names
a real listing. `SourceVocabularyIssues` (`component_source_guard.go`) checks this against the live
aspect set. The classic failure: the writer invents a plausible-but-absent aspect
(`site_specs.locale.currency_symbol`, `site_specs.ctas.primary_url`), whose fields would resolve
nowhere on every site and render silently link-less (`bugs_open/309`). The value the writer wanted
often **exists nowhere in the data model** — the correct answer is then to hardcode it or drop the
field, not to invent a source.

---

## 3. The birth gate — `store_generated_component`'s pre-store checks

Before the INSERT/UPDATE, `store_generated_component` runs a battery of checks and collects
`blockingIssues`; a non-empty list refuses the store. In rough order (verify against the file — the
lines drift):

| check | what it catches | blocks? |
|---|---|---|
| structural (`<section>`/`<div>` present; `<style>`/tag balance) | CSS-only or truncated output | yes |
| empty / legacy-dialect `input_schema` | no content fields, or the retired `properties` dialect (DB CHECK) | yes |
| 0 placeholders but schema declares fields | content unreachable | yes |
| **schema/template mismatch** (`SchemaTemplateSynced` false) | unknown vars and/or orphan fields (§2a) | **see §3a** |
| substantive template (>500 chars) with 0 placeholders | static markup, no content path | yes |
| `<no value>` render artefacts / fabricated-fact fallback | a rendered-against-empty template, or a business fact substituted for a missing datum (`bugs_open/140`, RFC_009) | yes |
| **source vocabulary** (§2b) | an invented / unresolvable `source` | yes |
| **stranding** (regeneration removes/renames existing fields) | a regen that drops a field live pages' content_data is keyed on | yes |

A refusal is recorded to `agent_error_log` (`recordValidationRejection`) AND, since `bugs_open/345`,
to the item's typed **`retry_feedback`** channel (§4). The recorder classifies the reason into three
producer codes: `component_validation_rejected`, `component_validation_orphan_schema_field`,
`component_validation_unknown_template_var`.

### 3a. Orphan-only mismatch is DROPPED, not blocked (owner ruling, 2026-08-25)

The schema/template mismatch check has two directions (§2a), and they are graded differently at the
gate:

- **Any unknown template variable → still blocks.** The template references a field with no source;
  the slot would render empty. That is a real defect and must be fixed by the writer.
- **Orphan schema fields ONLY (no unknown vars) → the orphaned fields are DROPPED from the stored
  `input_schema` and the component is stored.** An orphan renders nothing whether kept or dropped,
  so dropping it is lossless for the page and stops a harmless mismatch from making the whole
  section un-renderable.

The rationale, and why it took an owner ruling: the recorder already scores orphan-only as
`warning`, yet the gate had been treating it as `blocking` — the same defect called harmless by one
half of the system and fatal by the other. [MEASURED 2026-08-25] **9 components had died with every
rejection orphan-only** (21 rejections) against **1** that completed — refused over fields the
framework would merely have ignored. Naming the offending field to the writer did **not** fix it
(`bugs_open/345`, refuted at n=2: two writers were shown the exact orphan and left it orphaned
again — they declare a CTA-label field and render the label as static text). So the repair had to be
at the gate, not in the prompt. This changes what the gate GUARANTEES, so it went through the
architecture path (2026-07-29 §1); the owner ruled "drop the orphan".

⚠ **The drop is safe on a regeneration too, and here is why it does not conflict with the stranding
guard:** the stranding guard refuses a regen that removes a field *currently rendering* on live
pages. An orphan is by definition **not rendering** (no template var), so dropping it cannot blank a
slot any page is showing. A field that is both an orphan in the new output and a live rendering field
of the incumbent is not an orphan of the incumbent — it would appear as a *rendered* field and the
sync check would pass. Keep this invariant in mind before widening the drop.

---

## 4. The failure → feedback → retry → terminate loop

A refused store does not end the item; it goes back to `triaged` and `component-creator`
regenerates. Two mechanisms, both from `bugs_open/345`, shape what happens next:

- **Typed retry feedback (`WII-026`, `site_work_items.retry_feedback`).** The refusal's message and
  code are written to a single-writer column; the loader surfaces them as
  `current_item.last_error` / `last_error_code`; `build-dispatch-loop`'s `call_handler` forwards
  them (a **strict allow-list** input_mapping — a new key is inert until it is added there); and the
  `component-creator` prompt renders a retry block **only for the three producer codes** (so an
  un-classified failure repeats with no feedback shown). The point: the retry is no longer
  regenerated from identical inputs — it is told what was wrong.
- **Repeat-failure early termination (`WII-029`).** When an opted-in item's failure is
  **byte-identical** to the one already recorded, the remaining attempts cannot help, so the budget
  ends early and the row is stamped `result.terminated_on_repeat = true`. Reached only after the
  transient/outage arm declines. Opted in per item type; `needs_new_component` is opted in.

**Reading a termination correctly:** it means "feedback was shown and did not help" ONLY when the
terminal error carries one of the three producer codes; on any other class it means "this class has
no feedback channel". Do not conflate the two in a census.

With §3a live, the orphan-only class no longer reaches this loop at all — it stores on the first
attempt. What remains in the loop is the genuinely unfixable-by-retry classes (invented sources,
stranding), where termination caps the spend.

---

## 5. Traps this pipeline sets (all with fuller entries in `LANDMINES.md` / `bugs_open|closed/`)

- **A rejection message with no field names is unactionable.** The bare "template variables and
  schema fields do not match" told the writer nothing; it now names every offender (§2a). A
  deficient message explains a failure but is not evidence that fixing it produces a success —
  `bugs_open/345` learned this the expensive way.
- **The retry-feedback key must be wired through `call_handler`'s allow-list**, or the loader and
  prompt are live and inert. This bug shipped a Go half and a prompt half that could not fire for a
  day (`bugs_open/345`, migrations 555/564).
- **A regeneration matches the incumbent by `function`, not by name similarity.** A different
  `function` silently creates a parallel duplicate instead of regenerating. Identity resolution
  (`resolveStorageIdentity`) precedes the gate.
- **A truncated generation persists a fragment and reports success.** `output_tokens == max_tokens`
  means the completion was cut, not finished; the structural checks catch the common shapes but
  check the artefact, not the status (`bugs_closed/337`, `bugs_open/012`).
- **`quality_issues` is now an exhaustive, sorted array**, not at-most-one-per-direction. A consumer
  that assumed one entry would be wrong — none did as of 2026-08-25, but it is a shared column.

---

## 6. Related

- `026_component_regeneration_flow` — the regeneration path in depth.
- `FOCUS_llm_reliability_for_component_generation` — why generations vary and how the loop copes.
- Concept register: **WII-026** (typed retry feedback), **WII-029** (repeat termination + marker),
  **WII-024** (the failure-write ladder the loop routes through).
- Bugs: `bugs_open/345` (the loop and the orphan-drop ruling), `bugs_closed/337` (the writer's
  contract prompt + truncation), `bugs_open/309` (source vocabulary), `bugs_open/140` (fabricated
  fallbacks), `bugs_open/388` (a two-resolver divergence that can refuse a compliant writer).
