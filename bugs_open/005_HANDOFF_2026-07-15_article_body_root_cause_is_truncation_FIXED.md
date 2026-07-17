# HANDOFF — article-body blanking: ROOT CAUSE is LLM TRUNCATION (max_tokens), fix applied + recovery in flight

**Filed 2026-07-15.** Companion/correction to `004_HANDOFF_image_landing_blanks_article_body.md`
and the parent `../HANDOFF_2026-07-14_article_body_json_envelope.md`. Read those for
the mechanism; read THIS for the corrected root cause and what was actually done.

## TL;DR

The blanked / JSON-leaking article bodies were NOT primarily an "unescaped
newlines" malformation (parent handoff §3). The dominant cause is **output-token
TRUNCATION**: `page-content-writer`'s `generate_content` step ran at
`max_tokens: 2000`, and long article bodies were cut off mid-object, producing
JSON that cannot be parsed OR losslessly salvaged (the tail was never generated).

Evidence: all 14 original envelopes were captured and classified — **12 truncated,
1 embedded-quote (`href="…"`), 1 newline-only.** `llm_call_log` shows the writer's
`generate_content` hitting exactly 2000 output tokens across loop iterations.

## Fix applied (this session)

1. **ROOT CAUSE — `max_tokens` 2000 → 8000** on the writer's `generate_content`
   step (`agent_definitions.default_config`, path
   `…process_sections_loop.config.sub_workflow.steps.generate_content.config.ai_service.max_tokens`).
   DB config = **live immediately, no pod restart** — CONFIRMED: a rebuild logged
   `max_tokens=8000` and the article-body iteration generated **3952 tokens**
   (was capped at 2000), i.e. a complete article. Writer def backed up to
   `bak_agentdef_pcw_20260715`.

2. **Code (committed to HEAD, ships next chassis build — NOT yet in prod):**
   - `json_envelope.go` — `ParseLLMJSON` repairs escaping-only malformation;
     leaves truncated docs unparseable ON PURPOSE (so the loud path fires, not a
     half-article salvage). `missingRequiredLLMFields` is the schema-required
     check. **Its tests PASS** — the `TestParseLLMJSON_RepairsLiveEnvelopes`
     failure noted in `002 §B` / `004 §3` was a superseded in-progress version;
     the current test is `TestParseLLMJSON_LiveEnvelopeDistribution` and asserts
     the real 1-repairable / 13-unparseable split. Full `actions` package test
     suite is green.
   - `ai_actions.go` `ExecuteLLMPromptAction` — repairs via `ParseLLMJSON` instead
     of silently storing a raw-text envelope.
   - `v3_site_actions.go` `RenderComponentAction` — refuses to render when a
     schema-required `source:"llm"` field is absent (no more silent empty div).
   - `rerender_page_sections_action.go` — light rerender ESCALATES to the writer
     when stored content_data lacks a required field, instead of blanking.
   - `rerender_single_page_action.go` `getPageSections` — names the dropped slot
     when it filters an empty section (stops the silent drop).

## GOTCHA found while recovering — pages with an EMPTY section plan

Recovery = regenerate via a `needs_page` work item → `page-build-handler`, which
rebuilds the whole page from its section plan (writer now uncapped). **This works
only for pages that HAVE a section plan.** `page-build-handler` resolves sections
via `load_page_sections_from_spec`: `site_plan_sections` → `site_specs.site_plan`
→ `pages.sections`. A page missing from all three plans to **0 sections**, the
handler finds nothing to build, and the item bails **safely to
`needs_human_review`** (no damage, no writer call).

Of the 13 broken pages, **12 had a plan (2–3 sections); 1 (robot-hands
`/blog/tool-gripper-payload-calculator-guide.html`) had 0.** For that one, I
reconstructed `pages.sections` from its deployed `page_components` slot names
(`["hero","article-body","call-to-action"]`) before enqueuing. If you hit a
`needs_page` that bails instantly to `needs_human_review`, check
`jsonb_array_length(pages.sections)` FIRST.

## Recovery status (as of filing)

- Canary #2 (robot-hands `/guides/tool-gripper-cycle-time-estimator-guide.html`):
  **regenerated complete + clean, verified healthy in DB.** Proof the fix works.
- Remaining **12** `needs_page` items enqueued (`created_by='json-leak-fix'`),
  processing on the build-dispatch loop. Watch:
  ```sql
  SELECT CASE WHEN btrim(pc.rendered_html) LIKE '{%' THEN 'LEAK'
              WHEN pc.content_data ? 'content' THEN 'healthy'
              WHEN length(pc.rendered_html)=1326 THEN 'BLANK' ELSE 'other' END AS state,
         count(*)
  FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
  WHERE cc.name='article-body' GROUP BY 1;
  ```
- **Full-page rebuild rewrites ALL sections** (hero + CTA too), not just the
  article body — bigger blast radius than "fresh article". Spot-check output,
  especially leopardess (audited-content workstream), for any fabricated claims;
  the writer prompt forbids invented metrics/testimonials/case-studies, so this is
  a verification, not an expected failure.

## RESOLUTION (2026-07-16) — all 17 article-body instances healthy, verified on served pages

All 13 recovered. 12 regenerated cleanly via `needs_page` rebuilds at the raised
token ceiling (some needed one retry — transient LLM failures under the load of 12
concurrent rebuilds). Served pages spot-checked live (article renders,
no `{ "content"` leak): robot-hands cycle-time, leopardess hierarchical,
finetuning data-risk. Leopardess regenerated content checked for fabricated
metrics — none (writer prompt's anti-fabrication rules held).

## FOLLOW-UP #1 (important) — the EMBEDDED-QUOTE gap causes recurrence

One page (finetuning `tool-ai-data-risk-checker-guide`) failed all 3 automated
retries. Root cause found: its article-body includes a contact CTA —
`<a href="mailto:…">` / `<a href="tel:…">`. The model emits those attribute quotes
**unescaped inside the JSON string**, so the response is malformed even though it
is COMPLETE (2144 tokens, not truncated). `ParseLLMJSON` escapes control chars but
does NOT repair embedded quotes (by design), so:
- the DEPLOYED code stores a text envelope → empty render;
- even AFTER the committed fix deploys, `ExecuteLLMPromptAction` still can't parse
  it → text envelope → `RenderComponentAction` fails loud → the build fails /
  needs_human_review. **It does not silently blank (good), but it also never
  succeeds.**

**Any article the writer decides to end with a contact link will hit this.** That
page was fixed manually this session: extract the complete content
(last-quote-before-brace), render through the real article-body template
(`text/template`, `missingkey=zero`), write `content_data.content` +
`rendered_html`, deploy via an **assemble-only** `page_rerender` (safe path).
Row backed up to `bak_pc_holdout_20260716`.

### CLOSED 2026-07-16 (after chassis redeploy) — `repairJSONStringLiterals`

`json_envelope.go`'s repair function was upgraded (superseding the earlier
"escape control chars only" version and the removed single-field
`SalvageStringField`): it now distinguishes a quote that ENDS a JSON string from
a quote that is just literal content (an HTML attribute) by checking what follows
it. A quote followed (after whitespace) by a JSON structural character — `,` `}`
`]` `:` — is a real terminator; anything else (a letter, `>`, another `"`) means
it's content, so it gets escaped and the string continues.

This is a GENERAL fix, not scoped to one field name or to the field being last in
the object (unlike the removed `SalvageStringField`, which only worked when
`content` was the sole/last key) — proven by
`TestParseLLMJSON_RepairsEmbeddedQuotesInNonLastField`. It preserves the
fail-loud property for genuine truncation: a truncated string never reaches a
quote followed by real JSON structure, so it stays unterminated and
`json.Unmarshal` still rejects it
(`TestParseLLMJSON_TruncatedWithEmbeddedQuoteStillFails`).

Verified against all 14 live fixtures: exactly 2 are genuinely JSON-complete
(raw bytes end `"}`) — the newline-only one (repaired since 2026-07-15) and
`850e356d` (finetuning.uk llm-cost-calculator-guide's original envelope — a
DIFFERENT page than the one fixed manually, but the same defect class: it also
signs off with the same finetuning.uk contact link) — both now repair correctly;
the other 12 are genuinely truncated (confirmed by inspecting raw byte tails,
independent of which parse error encoding/json happened to report) and
correctly remain unparseable. See `TestParseLLMJSON_LiveEnvelopeDistribution`
(repaired=2, unparseable=12) and `TestParseLLMJSON_RepairsEmbeddedQuotes`.

No further action needed — this closes the follow-up. Ships on the next chassis
build containing this branch.

## Backups
`bak_pc_articlebody_20260715` (broken rows), `bak_agentdef_pcw_20260715` (writer def),
`bak_pc_holdout_20260716` (the manually-repaired row).
