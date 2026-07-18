# 016 — Council revise/reframe prompts silently drop every reviewer's output

*Found 2026-07-18 by the experience-loop thread while building its own council.
Affects `fix-proposer` and `feature-designer` (both LIVE). Not fixed here —
those agents belong to the fixloop / feature-builder / council-gate threads,
and the council-gate notes warn to diff the seed against the LIVE row before
re-applying (the roster moves fast).*

## The defect

`fix-proposer`'s `repropose` and `reframe` prompts inject the council's reviews
with `{{.review_editquality.result}}`, `{{.review_guardian.result}}`,
`{{.review_bug_historian.result}}` (and in `feature-designer`, also
`review_guidelines`, `review_reuse_agent`, `review_tooling_provenance`).

**Those render `<no value>`.** The revise loop therefore re-proposes WITHOUT
seeing the objections it is supposed to be addressing — it only sees the
previous plan and the instruction "address every objection".

## Why

Prompt template data is built by `ExtractFields` (`datahelpers`), which calls
**`UnwrapDeep`** on every extracted field. `unwrapRecursive` Pattern 2 strips
any map carrying a `result` key and returns the value. So for an
`execute_llm_prompt` step:

| Where | Shape | Correct reference |
|---|---|---|
| RAW `collected_data` (config dot-paths: `review_fields`, `plan_body_field`, `check_fields`) | `{type, result}` — documented in 001 §"Common Data Shapes" | keep `.result` |
| Prompt TEMPLATE data (after ExtractFields → UnwrapDeep) | the unwrapped value itself | **`{{.step}}`, NOT `{{.step.result}}`** |

That asymmetry is the trap: the same dotted string is right in config and wrong
in a template.

The two failure modes differ, and the quiet one is the dangerous one:
- **text-output step** → `{{.step.result}}` is a nested access on a string →
  hard execution error (001 §Step 6: "`{{.missing.nested}}` → execution error").
- **json-output step** → the unwrapped value is a map with no `result` key →
  Go renders **`<no value>`**, silently. This is the fix-proposer case.

## Evidence

- Regression test locking the whole contract:
  `platform/orchestration/datahelpers/template_result_wrapper_test.go`
  (`TestUnwrapDeep_TemplateVsConfigPaths`). It asserts config paths still
  resolve with `.result`, that template data is unwrapped, that
  `{{.proposal.result}}` errors, and it logs the confirmed silent render:
  `CONFIRMED silent degradation: {{.review_journeys.result}} rendered "[<no value>]"`.
- Live-DB sweep of active agents whose `execute_llm_prompt` prompt_template
  contains `.result}}`:

| Agent | Steps | Referenced |
|---|---|---|
| `fix-proposer` | `repropose`, `reframe` | review_editquality, review_guardian, review_bug_historian, review_guidelines, review_reuse_agent, review_tooling_provenance |
| `feature-designer` | `repropose`, `reframe` | same family |
| `content-creator-hero` | `generate_hero_content` | `call_researcher` — **different shape**, verify separately: `call_agent` returns `{response, response_status}` and UnwrapDeep has a distinct Pattern 4 for it; do NOT assume this one is broken |

- The experience-planner hit the loud half of this on its first live run
  (`{{.proposal.result}}`, a text step) and was fixed by dropping `.result`
  in templates only — commit `6c5dc9e13`, `167_experience_planner_and_council.sql`.

## Suggested fix (for the owning threads)

In `repropose`/`reframe` prompt templates only, `{{.review_X.result}}` →
`{{.review_X}}`. **Do not touch** `review_fields`, `check_fields`,
`plan_field`/`plan_body_field` or any other config dot-path — those read raw
collected_data and are correct as-is.

Worth a quick check of whether any council has been converging *because* its
reviser never saw the objections.
