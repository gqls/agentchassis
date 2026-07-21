# NOTES — bugs_open/026 schema-dialect fail-open (append-only, newest at bottom)

## 2026-07-21 — session 1: diagnosis, fix, council submit

**Entry state.** Assigned "bugfix 026". Grep showed no active owner: `who-owns.py 026`
returned OWNED pointing at news_feed_pooling, but that was a **false positive** — the tool
matched "2026" date substrings, not "bug 026". Precise grep (`bugs_open/026|bug 026`) found
only 027's cross-reference. The bug-backlog-clearing thread (HANDOFF_2026-07-21) has NOT
touched 026. So: unowned, clear to work.

**Grounding moved the whole case.** Filed 2026-07-19; by 2026-07-21 the relojistas/027 thread
had landed seed 179, which:
- server-renders news items from `query.news_archive` (Defect A dead — placeholder no longer
  shown when items exist) and added a `loading_text` LLM field for the empty case;
- converted news-listing's input_schema from the legacy JSON-Schema dialect to v2 `fields`
  (incidentally fixing news-listing's own instance of Defect B).
Verified from `curl` (JS off): relojistas/gaswholesalers/robot-hands serve 0 placeholders +
150+ server `<article>`s, relojistas in Spanish. All 3 served `<h1>`s non-empty.

**The near-miss I avoided.** My first instinct was to read `validate_page_content.go` (named
in the bug as the enforcer). It does NOT enforce input_schema at all — it scans rendered HTML
for placeholder/template/contamination patterns; its `<no value>` check would only catch a
*fully-absent* template key, never a present-but-empty string. If I'd stopped there I'd have
"fixed" the wrong file. The real enforcer is `missingRequiredLLMFields` (json_envelope.go),
found by grepping `on_missing`/`required` across the actions package.

**Root cause (both layers).** [VERIFIED by reading each function, not grep alone]
- Generation `plan_sections_action.go:1182`: `comp.InputSchema["fields"].(map)` miss →
  "all fields from LLM", empty llmFieldSpecs; name-keyword backstop (:1206) defers only
  article/content/body/text/blog — news-listing matches none.
- Enforcement `json_envelope.go:193`: `inputSchema["fields"].(map)` miss → nil.
Both read only `fields`; the legacy `properties` shape → both fall open.

**Confirmation I did not expect:** seed 179's own header comment states the exact mechanism
(*"INVISIBLE to plan_sections — only 'fields' is parsed ... required headline could ship
empty"*) and explicitly assigns *"the structural half of 026 ... stays open and is 026's."*
Independent arrival at the same cause + an explicit hand-off. Good sign the diagnosis is right.

**Verified the render path actually hits the gate** (the "read the function you skipped"
check). Worried news pages might render via `render_news_section_action.go` and bypass the
gate. Read it: its own comment (:322-339) says items now render through *"the normal template
path (plan_sections → queryresolve → content_data → RenderComponentAction)"* — the gated path.
So the gate does apply. render_news_section only writes the JSON files + queues the rerender;
it defaults headline to "Latest News" for the JSON only, not the HTML `<h1>`.

**Gate is live in the pod** (v1.0.1144): "refusing to render an empty section" x2 +
"escalating page to writer instead of blanking" x1 grep non-zero in /app/agent-chassis;
positive control present.

**Fleet-wide dialect check:** 173 components — 124 v2 / **0 legacy** / 42 empty / 7 bare
example-value. Legacy shape extinct → 026's specific recurrence gone; residual = the
fail-open itself.

**Residuals are NOT this defect.** Live `content_data`: idea.uk news-listing `headline=''`
(but page 404s; `page_type='section-index'`) and ai-agent-orchestration `/news.html`
(`page_type='content'`, old May-1 page). Neither is `page_type='news-index'`, so
`render_news_section_action.go:215,282` never touches them → stale content persists. That is
the bugs_open/015 mistyped-page_type class. Routed there in the 026 file.

**Fix built** (commit `fd87c8ebf`): shared `schemaContentFields()` reader
(`component_schema_fields.go`) + both call sites swapped + regression test. Package green,
existing v2 test unchanged. Design decision recorded in PLAN: normalise (understood) over
reject (blocked); bare example-value + empty schemas stay ok=false so the 7 legacy core
sections are untouched.

**MISSTEP — council submission schema.** First 097 submission failed: I put `plan` as a
top-level array and `grounded_in` at top level. The trigger wants `plan` as an **object**
with `plan.summary` + `plan.edits[]` + `plan.grounded_in[]`. Fixed and resubmitted.
`SUBMISSION_CORR=a85c1220-7174-41fe-8892-64009eadcf47`. (Logged to RUNBOOK so the next
submitter doesn't repeat it.)

**Sequencing decision:** committed the code BEFORE the verdict (the tree moved HEAD twice
during this session; holding uncommitted risks a sweep). No `Council-Reviewed` trailer yet —
it is earned by APPROVED only; putting it on early would be a false claim. Deploy is gated on
the verdict, not the commit.

**Open:** council verdict pending (~30 min queue). Then build/deploy/verify behaviourally,
then close 026.
