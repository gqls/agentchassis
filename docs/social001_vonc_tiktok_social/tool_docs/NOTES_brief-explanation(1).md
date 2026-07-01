# NOTES — brief-explanation

Append-only history for the `brief-explanation` component. Newest at the bottom.
`Categories:` uses the shared taxonomy (TOOL_DOCS_convention.md).

---

## 2026-06-29 — Identified as a Mode-B empty shell
**Observed:** `html_template` full of bare `<no value>`, `input_schema` `{}`, 0
slots, quality 50, id 58363894-9db9-4d2f-81ac-c47b54d97fc3. CSS selectors (`.be-*`)
intact. It is a "how it works" explainer — eyebrow, heading, 3 numbered steps, 3
stats, 2 CTAs, image + badge.
**Decision:** it is STATIC content, so the fix is REGENERATION with a real schema
(content writer fills at build), NOT a JS loader. (Distinction: Option-2 runtime JS
is only for daily-dynamic shells.)
`Categories:` empty-shell, mode-b-template

## 2026-07-01 — Regen attempt 1 (081): produced a STRAY generic hero
**Action:** triggered component-creator via spawn+call (081), mirroring the 080 test
script — all inputs mapped TOP-LEVEL (`section_type`, `description`, ... →
`input_data.section_type`, ...).
**Result:** component-creator ran but generated `function='general-hero'`,
`display_name='General Hero'`, `status='created'` (NEW row 0ef52c95), quality 100.
brief-explanation (58363894) untouched.
**Root cause:** component-creator's workflow reads `input_data.spec.*`, but 081 put
fields top-level, so the generate_template prompt got an EMPTY section_type +
description and defaulted to a generic hero. The emitted `function='general-hero'`
didn't match the existing `brief-explanation` row, so store INSERTed a stray.
**Cleanup:** deactivate the stray 0ef52c95 (unreferenced).
`Categories:` workflow-variable-path, input-shape

## 2026-07-01 — Regen attempt 2 (082): contract violation, no run
**Action:** 082 nested ALL inputs under `spec` and mapped `spec → input_data.spec`.
**Result:** `call_agent` failed: `contract violation ... missing required fields:
[section_type]. Provided fields: [domain site_id spec]`. component-creator never ran
(failure at the call_agent extraction/validation stage) — so NO stray this time.
**Root cause:** `call_agent` validates the target's `input_contract.required`
against the TOP-LEVEL extracted fields; `section_type` was nested under `spec`, so
the contract failed. Confirmed the split: the CONTRACT wants `section_type`
top-level, the WORKFLOW reads `input_data.spec.*`.
`Categories:` spawn-call-contract, input-shape

## 2026-07-01 — Regen attempt 3 (083): provide BOTH top-level + spec
**Fix:** 083 provides `section_type` TOP-LEVEL (satisfies the contract) AND the full
`spec` object (satisfies the workflow), with one-level mapping sources (proven to
resolve from the 082 log). Function name pinned to `brief-explanation` in the
description so the in-place UPDATE lands.
**Expected:** component-creator generates a real brief-explanation, emits
`function='brief-explanation'`, store UPDATES 58363894 in place (status
`regenerated`, active_rows=1, schema populated, no `<no value>`). Then a needs_page
rebuild of the index fills the new fields via page-content-writer.
`Categories:` workflow-variable-path, input-shape
(Durable lesson recorded in 016b §9: "Manually invoking an agent via spawn+call —
input_mapping must satisfy BOTH the input_contract (top-level) AND the workflow's
field paths (usually input_data.spec.*)".)

## 2026-07-01 12:46 — Regen 083 SUCCEEDED (in-place UPDATE)
**Result (corr ecbca5cb):** 58363894 updated in place — quality 50→100,
template_variable_count 0→20, schema_field_count 0→20, tmpl 8089→9760,
has_no_value t→f, active_rows=1, created_at unchanged, updated_at 12:46. Raised
needs_rerender `component_regen_rerender:58363894-...` (triaged).
The dual-placement (section_type top-level for the contract + full spec for the
workflow) with the function name pinned in the description landed the in-place UPDATE.
**Still to do:** the 20 schema fields are EMPTY. needs_rerender only ASSEMBLES; a
page content-rebuild (needs_content_page) is required to fill them via
page-content-writer — but a full index rebuild risks the interactive-page clobber
(hero + provocation-card). Render only brief-explanation onto the index, or confirm
the working sections survive, before triggering.
`Categories:` input-shape, detool-on-rebuild (hazard noted)

## 2026-07-01 — build-dispatch-loop lacks section_type (work-item path also blocked)
build-dispatch-loop's call_handler input_mapping (id 099b51e0) flattens only
work-item columns + optional (?) spec fields + the whole spec; NO section_type. So
component-creator via the loop would hit the SAME `missing required fields:
[section_type]` contract error. The generic loop can't satisfy a top-level-required
contract. Framework fix recorded in RUNNING_NOTES + runbook Appendix E + debugging
guide §9: the contract validator should accept required fields from top-level OR
input_data.spec.{field}. Do not patch the loop mapping per-handler; do not bless the
083 duplication as the rule.
`Categories:` spawn-call-contract, input-shape
