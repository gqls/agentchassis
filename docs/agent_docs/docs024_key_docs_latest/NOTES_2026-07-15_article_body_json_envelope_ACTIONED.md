# NOTES — Actioning the article-body JSON-envelope fix (2026-07-15)

Companion to `HANDOFF_2026-07-14_article_body_json_envelope.md`. Records what was
verified, what the handoff got wrong, what was changed, and the open execution
step. Kept as a separate file (not edited into the handoff) because that thread
has concurrent editors.

## Corrected diagnosis — it is TRUNCATION, not unescaped newlines

The handoff's stated root cause ("All 14 broken envelopes are MALFORMED JSON —
literal newlines inside the string values") is wrong for 13 of 14. All 14 raw
envelopes were captured from `content_data->>'result'` and classified:

- **12 TRUNCATED** — the writer's `generate_content` step ran at
  `max_tokens: 2000`. `llm_call_log` confirms `page-content-writer /
  generate_content` calls hitting exactly 2000 output tokens across loop
  iterations 0–7 (comparable long-form generators run 8000–16000). The article
  JSON is cut off mid-sentence, so it is genuinely INCOMPLETE and cannot be
  "parsed" or losslessly salvaged — the tail was never generated.
- **1 embedded-quote** — unescaped `"` from HTML `href="…"` attributes.
- **1 newline-only** — the case the handoff describes.

Consequence for recovery: "the words are all still there" (handoff §5) is false.
Salvage yields a partial article. Decision taken: **regenerate** the broken pages
through the writer with a raised token ceiling, rather than salvage.

## Blast-radius note (recorded at user request, 2026-07-15)

> Both `needs_page` and `content_rewrite` route to `page-build-handler`, which
> regenerates the WHOLE page (every section), not just the article body — there
> is no surgical single-section regeneration path. That is a materially bigger
> blast radius than "fresh articles" implied, and it lands on live client sites
> in an area with concurrent activity. Re-verify the code still builds against
> the latest working tree, and confirm the broken set is stable, before
> executing the regeneration.

Implication: regenerating the 13 rewrites all copy on 13 live client pages. This
is why execution is gated behind a canary rather than a bulk fire.

## Changes made this session

Root cause (DB, live, reversible — def backed up to `bak_agentdef_pcw_20260715`):
- `page-content-writer` → `generate_content` → `ai_service.max_tokens`: **2000 → 8000**.
  Path: `default_config.workflow.steps.process_sections_loop.config.sub_workflow.steps.generate_content.config.ai_service.max_tokens`.

Code (branch `085_debug_and_feature_loops`, NOT yet deployed — an image would
bundle other sessions' uncommitted work; ships on the next normal release):
- `platform/orchestration/actions/json_envelope.go` (new) — `ParseLLMJSON`
  (repairs escaping-only malformations; leaves truncated docs unparseable on
  purpose) and `missingRequiredLLMFields` (schema-required-field check). Unit
  tested in `json_envelope_test.go` against all 14 captured live envelopes.
- `ai_actions.go` — `ExecuteLLMPromptAction` now repairs via `ParseLLMJSON`
  instead of silently falling back to a raw-text envelope on minor malformation.
- `v3_site_actions.go` — `RenderComponentAction` refuses to render a section when
  a schema-required `source:"llm"` field is absent (no more silent empty div).
- `rerender_page_sections_action.go` — the light rerender ESCALATES to the writer
  (instead of blanking) when stored `content_data` lacks a schema-required field.
- `rerender_single_page_action.go` — `getPageSections` now names the dropped slot
  when it filters an empty section (stops the silent drop that hid the blanking).

Discovery check: the handoff's requested check ALREADY EXISTS
(`discovery_checks/check_required_fields_missing.go`) and was confirmed to flag
all the broken article-body rows — "extend rather than duplicate" → nothing to add.

## Backups

- `bak_pc_articlebody_20260715` — the broken `page_components` rows.
- `bak_agentdef_pcw_20260715` — the writer definition before the max_tokens raise.

## Open execution step

Regenerate the currently-broken **13** article bodies (9 BLANKED + 4 JSON-LEAK)
via `page-build-handler`, canary-first (one robot-hands blog page), then the rest.
Concurrent activity is present in the area (robot-hands/vonc), but no work items
are queued for the 13, so no collision — target the current broken set, verify
each before/after.
