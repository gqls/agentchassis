# FOCUS — LLM reliability strategy for component generation

## Date: 2026-05-11

## Status: proposal in progress

---

## Context

Component-creator generates HTML+CSS+schema for reusable components via
LLM (`claude-sonnet-4-6`). Output passes through pre-store validation
which enforces a contract between template variables and schema fields.

During the directory-builder / Tier D work (May 2026) we observed that
LLM output is structurally good (HTML, CSS, semantic markup, iteration
shape, field name choices) but **unreliable at exact list reconciliation**
between the schema's declared fields and the template's variable
references. Failures observed:

- Schema declares `card_link_label`, template uses literal "Open tool"
- Schema declares `nav_label` in array sub-schema, template doesn't render it
- Schema and template both populated but slightly different name sets
- `{{$.X}}` vs `{{.X}}` confusion

These aren't creative failures. They're bookkeeping failures —
the kind LLMs are known to be weak at and deterministic code is
strong at.

---

## Strategy

Three tracks, in priority order. The earlier items are smaller and
should be done first.

### Track 1 — Make rejection observable

Done in this iteration. Pre-store validator writes structured rows to
`agent_error_log` on rejection, with the orphan/unknown field names
parsed into typed arrays in the context JSONB. Severity reflects
whether the failure is bookkeeping (`warning`) or structural (`error`).

This unblocks pattern analysis. Without it, every rejection is a
needle in a haystack of chassis logs. With it, "show me the 20 most
common orphan field names across the last 7 days" is one SQL query.

### Track 2 — Move bookkeeping out of the LLM

Where the LLM keeps failing the contract, see if the contract can be
narrowed instead. Three candidates:

**(2a) Root section wrapper injection.** The `<section class="X-section"
data-component="X">...</section>` wrapper is always the same shape per
component. Inject at store time. The LLM produces inner HTML only.
Removes a class of failures (`data-component` attribute missing,
wrong class name, wrong root tag).

**(2b) Tier D sub-schemas declared centrally.** The resolver
(`queryresolve`) is the authority on what fields a
`query.pages_where_type:X` returns. The LLM shouldn't redeclare these
per component. Move the declaration to the resolver; have the validator
look there. The LLM stops writing the sub-schema entirely.

**(2c) Schema fields derived from template (ambitious).** If the
parser extracts the canonical field list from the template, the
schema's `fields.X` keys could be derived rather than LLM-declared.
The LLM still provides per-field metadata (`source`, `llm_guidance`,
`fallback`) but the KEY list is parser-extracted. Eliminates
direction-2 mismatches by construction. Requires changing the LLM's
return format though.

Order: (2a) first, (2b) next, (2c) only if needed.

### Track 3 — Prompt and model adjustments

Only AFTER tracks 1 and 2 have been worked. Premature here because
the patterns aren't yet visible. Once `agent_error_log` shows the top
failure modes, the prompt can be tightened against them. If the prompt
plateau is reached, consider a larger model.

---

## What we are NOT doing

**Auto-correction at the validator.** Considered and rejected. The
validator silently dropping LLM-declared fields would mask intent and
prevent the pattern analysis that drives prompt fixes. The validator
fails loudly; the LLM is expected to meet the contract; when it
doesn't, the failure goes into `agent_error_log` and we act on
patterns.

**Single ad-hoc fixes per component.** Every time we hit a mismatch,
the temptation is to retry the LLM or write the component by hand
(as we did with `tool-list` in migration 041). That's appropriate
for unblocking but doesn't reduce the failure rate. Treat hand-written
components as known-good references AND as data points about what
shape the LLM should converge on. Don't accumulate hand-written
components without also addressing the underlying prompt or
decomposition.

---

## Success criteria

A freshly-launched site's adoption + plan + build cycle produces a
deployable site without manual SQL intervention to fix component
output.

Measurable:
- Zero `agent_error_log` entries with `severity=error` from
  `store_generated_component` during a full site build
- All `severity=warning` entries are recoverable by retry within the
  configured attempt cap (currently 3)

---

## Open questions

- **Retry budget.** Currently 3 attempts before a work item lands in
  `failed`. Worth bumping to 5 once tracks 1-2 are in place? Each
  retry costs an LLM call, but bookkeeping-only failures are often
  resolved by a re-roll. Decide after a week of `agent_error_log`
  data.
- **Sub-schema enforcement.** Currently the validator accepts
  sub-schema fields declared but unused. After Track 2b lands, this
  changes — sub-schema fields come from the resolver, not the LLM,
  and "unused" is fine because they're available to any template that
  ranges over the same query. Worth keeping the lenient direction-2
  check.
- **Component versions.** Each LLM regen attempt currently increments
  `component_versions`. Rejected attempts DON'T currently increment
  (the rejection happens before the INSERT). Worth verifying after
  the track 1 changes deploy.

---

## References

- `026_component_regeneration_flow.md` — full validation gate
  description and migration history
- `019_tool_library.md` — Tier D / query.* source resolution
- `FOCUS_directory_builder_and_list_components.md` — the work that
  surfaced these failure modes
- Migration 041 — hand-written `tool-list` as canonical Tier D
  reference
