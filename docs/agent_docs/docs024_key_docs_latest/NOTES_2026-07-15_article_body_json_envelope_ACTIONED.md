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

---

## Session 2026-07-19 — re-verification of 005 against the LIVE system (no code changed)

Picked up `bugs_open/005_..._FIXED.md` to fix. **It is genuinely fixed.** The doc
claimed closure on 2026-07-16; this session re-derived that from the live system
rather than trusting the doc, because the closure predates a known config re-seed.
Five independent checks, all green:

1. **`max_tokens` still 8000, and it SURVIVED a re-seed.** This was the check most
   likely to fail — the re-seed clobber landmine (patch-style seeds overwrite
   `default_config`) and `agent_definitions.updated_at` = **2026-07-18 22:20:23**,
   i.e. the row *was* rewritten after 005 closed. The value held:
   ```sql
   SELECT type, updated_at,
          jsonb_path_query_first(default_config,'$.**.generate_content.config.ai_service.max_tokens')::text
   FROM agent_definitions WHERE type LIKE '%writer%';
   -- page-content-writer | 2026-07-18 22:20:23.9168+00 | 8000
   -- content-writer      | 2026-07-18 22:20:23.9168+00 | 8000
   ```
   Note the column is `type`, NOT `name` — `\d agent_definitions` first; a `name`
   query errors out.

2. **Repair code is in the RUNNING pod**, not merely in git:
   ```
   kubectl exec -n ai-persona-system agent-chassis-5c568b8c74-2f4qv -- \
     sh -c 'strings /app/agent-chassis | grep -c "repairJSONStringLiterals"'   -> 2
   ```

3. **19/19 article-body instances healthy** — zero LEAK, zero BLANK (the doc's own
   query, §Recovery status). Was 17/17 at closure; the 2 created since are clean,
   so the fix holds for NEW pages, not just the recovered ones.

4. **Zero page-content-writer truncation since the fix.** Fleet-wide sweep of
   `llm_call_log` for `output_tokens >= max_tokens` over 7 days: the writer's only
   truncated rows are `max_tokens=2000`, latest **2026-07-15 15:21** — i.e. all
   pre-fix. Nothing at 8000.

5. **All 005 tests pass at committed HEAD** (see the caveat below for why "at HEAD"
   is load-bearing): `TestParseLLMJSON_LiveEnvelopeDistribution`,
   `_RepairsEmbeddedQuotes`, `_RepairsEmbeddedQuotesInNonLastField`,
   `_TruncatedWithEmbeddedQuoteStillFails`, `TestMissingRequiredLLMFields` — all PASS.

**Conclusion: nothing to fix in 005.** No code or config changed this session.

### BAD — the `actions` package does not compile in the WORKING TREE (not ours)

`go test ./platform/orchestration/actions/` fails to BUILD:
```
complete_work_item_verification_test.go:123:22: assignment mismatch:
    2 variables but handlerReportedFailure returns 3 values
```
Another thread has changed `handlerReportedFailure`'s signature in
`complete_work_item_verification.go` (modified, uncommitted) without updating
`complete_work_item_verification_test.go`. **Left alone — it is their in-flight
work, and "confirm the state of the code beneath you" cuts both ways.**

Consequences worth knowing:
- It blocks `go test` on the whole package for *every* thread, including 005's tests.
- It does **NOT** poison builds: `make build-*` archives committed `HEAD`, and
  **HEAD compiles clean**. Verified by testing a `git archive HEAD` extract in the
  scratchpad rather than stashing — never `git stash` here, it would yank another
  session's WIP out from under them.
- Technique worth reusing: to test what a build would actually ship, run
  `git archive HEAD | tar -x -C <scratch>` and test there. Read-only w.r.t. the
  shared tree.

### Adjacent finding — 012's exposure is closed (verification for that thread)

The same truncation sweep flagged live-looking truncation elsewhere, all of it
already filed, so **nothing new was filed** (grepped `/bugs_open/` first):
- `tool-generator.generate_tool_html` 3/4 truncated, `tool-improver.improve_tool`
  3/4 — both `bugs_open/012`. Migration 168 **is applied**: both now read
  **32000** live. Every truncated row predates it (latest 2026-07-18). So 012's
  immediate exposure is genuinely closed, not just claimed.
- `generic.review_editquality` 1/20 at 8000 — that is `bugs_open/019`.

Reminder for whoever reads this: `output_tokens >= max_tokens` grouped by
`agent_type, step_name` over `llm_call_log` is a cheap fleet-wide truncation
sweep, and it distinguishes pre-fix from live rows via `max(created_at)` — which
is what stopped this session from re-filing 012 as a new bug.
