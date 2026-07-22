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

### 2026-07-21 later — first council submission ruled INVALID (my error, not a verdict)

First run (`corr a85c1220`) completed at step `complete_invalid` with `__step_error`:
*"plan failed validation: edit 1: operation \"create\" not in the allowlist; edit 4:
operation \"create\" not in the allowlist."* The reviewers never ran — this is a
persist-time schema reject, before any credits are spent. `diagnose_persist_fix_plan_action.go:80`
allows exactly `modify | add | remove | config_change`. **A NEW file is `add`, not `create`**
(I assumed a git-verb vocabulary; it is not). Fixed both new-file edits to `add` and resubmitted:
new `SUBMISSION_CORR=cbbc7c83-d073-419a-bfc5-6ab26e687d9c`, orch `f31330cb-5f6b-4f0d-8f4c-6080229b9702`.
The code commit `fd87c8ebf` is unaffected — this was only the council submission's plan JSON.
(Logged to RUNBOOK + WRONG_CALLS.)

### 2026-07-21 later — council round 1 ran: REVISE from bug_historian (2/3 seats approved); addressed and resubmitted

Round-1 verdict on `corr cbbc7c83`: **REVISE**, `abstained:6`. `editquality` and `reuse_agent`
APPROVED — reuse_agent called it "the reuse discipline working correctly" and editquality
confirmed both named readers get a real covering edit. `bug_historian` objected, two points,
both fair:

1. **Call-site completeness** (medium): *"is two call sites really all of them?"* It flagged
   `content_components.schema_field_count`/`schema_template_synced` as implying more readers.
   **It was right.** I audited every `input_schema["fields"]` reader repo-wide — **9 genuine**
   (the many `config["fields"]` hits are workflow-step config, not input_schema). Classified by
   legacy-consequence:
   - *correctness (required field ships empty — the 026 class):* plan_sections (generation) +
     missingRequiredLLMFields (render gate) — already rewired.
   - *the direct safety-net companion:* `check_required_fields_missing` (post-deploy "did a
     required field ship empty?" audit) — **now rewired**; it would have failed open on the same
     dialect the gate now handles.
   - *different consequence (wrong metric/CTA/array, NOT missing content):* compute_component_quality,
     store_generated_component, load_existing_component, check_image_source_unsatisfiable,
     ctafields, expectedItemFieldsFromComponentSchema — named and **left direct**, covered by the
     tripwire below.
   To rewire the audit (a different package), I **relocated the reader to `datahelpers`**
   (`SchemaContentFields`) — the shared home both `actions` and `discovery_checks` import
   (no cycle: datahelpers imports neither). reuse-disciplined and it completes the extraction.
2. **No fail-loud signal** (missing): a silent-correct-path means a re-seeded/restored legacy
   dialect is absorbed invisibly forever. **Added** `fromLegacy` return + `WarnLegacyDialect`;
   plan_sections (every build) and the audit (post-deploy) fire it. A Warn is what bug_historian
   listed as acceptable ("a log line, diagnosis_artifact, or site_work_item").

Round-2 code committed `f27c5ad1d` (7 files; datahelpers + actions tests green; the
`discovery_checks` package RED is **pre-existing** — verifier_coverage for
`contact_form_undeliverable`/`backend_entry_orphaned`, other sessions' checks, not
`required_fields_missing`; confirmed my change is not the cause). Resubmitted on the SAME corr
`cbbc7c83` (RESUBMIT_CORR, trail accumulates), orch `f1457624`.

### 2026-07-22 — council round 2: REVISE again (bug_historian medium), addressed in round 3

Round-2 verdict: **REVISE**. editquality now APPROVED (with a low note); bug_historian held on
two "verify-don't-assert" points, both fair:
1. *Gate tripwire gap* (low, both seats): the render gate is reached on a re-render/redeploy
   WITHOUT a plan_sections pass (`rerender_page_sections_action.go:210` runs on stored
   content_data), so a reintroduced legacy dialect could hit the gate silently. → **Fixed r3**:
   new `datahelpers.WarnIfLegacyDialect` fired at BOTH gate call sites (v3 render + rerender).
   Tripwire now comprehensive across generation + render + rerender + audit — no legacy
   component is silent on any path to a served page. Kept `missingRequiredLLMFields` 2-arg (the
   tripwire lives at the call sites) to avoid a signature ripple into 11 test calls incl. the
   article-body workstream's `json_envelope_test.go`.
2. *Triage asserted not demonstrated* (medium): I'd *claimed* the 6 left-direct readers were a
   "lesser class" without reading them. bug_historian named two — and was **right**. I read them:
   - `check_image_source_unsatisfiable` (`if !ok continue`) skips a legacy component entirely →
     a required image never checked → silently-absent image. **Rewired + tripwire.**
   - `DeriveCTAURLFields`/`UncoveredCTAURLFields` return nil on legacy → underived → broken/empty
     CTA once precedence flips off observe-only. **Rewired.**
   The 3 truly-left readers (compute_component_quality field-count, store_generated sync flags,
   load_existing field-name print) are demonstrably metric/metadata/creation-aid — verified by
   reading them; none emits served content. The principled line: *does the fail-open silently
   break/hide SERVED content?* (yes → rewire; no → leave, and the tripwire covers visibility).

Round-3 code committed `cbacd450c` (6 files; datahelpers+actions green; discovery_checks RED
still pre-existing/unrelated). Resubmitted r3 on corr `cbbc7c83`, orch `670e611a`.
LESSON reinforced: bug_historian's whole method is "you asserted, demonstrate it" — the same
verify-don't-assert discipline this bug is about. Reading the two named call sites turned an
assertion into two real rewires. Don't classify a reader's consequence from its name; read it.

**MISSTEP — `git stash` in a shared tree (nearly pulled another session's WIP).** To verify the
discovery_checks RED was pre-existing I ran `git stash push -- <two paths>` where one path was
an *untracked* file → the push errored (untracked paths don't match a pathspec stash) and did
nothing useful. Then a bare `git stash pop` tried to pop `stash@{0}` — which in this many-session
tree was **another branch's stash** (`066_hitl_questionnaire`), not mine — and would have applied
its `coordinator.go`/awaited_requests WIP into my tree. It failed harmlessly only because
`coordinator.go` had conflicting local changes. **Never `git stash pop` here without checking
whose `stash@{0}` it is** (`git stash list` first). My work was intact throughout; no code lost.
Logged to WRONG_CALLS. Better verification for "is this RED mine?": `git show HEAD:<file> | go
test` against an overlay, or just reason about whether the change *can* affect the failing
assertion (mine can't — the failing item types are files I never touched).
