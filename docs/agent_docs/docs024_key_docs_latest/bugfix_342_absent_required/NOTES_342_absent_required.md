# NOTES — bugfix_342_absent_required (append-only, newest at the bottom)

## 2026-08-22 — state verification before touching anything

- who-owns: "OWNED or recently active" — but the activity is the `bugfix_260_render_fallback`
  lane's own filing + fix work, finished 2026-08-21 evening (their NOTES: APPROVED round 6,
  live-verified on v1.0.1322). The bug file says UNOWNED; RFC_041 §5 says "what it needs now
  is a name". This lane is the name. Tree checked: nothing dirty touches the 342 surfaces.
- **The bug file's warning banner is STALE.** It says the editor-route escalation "is NOT in
  v1.0.1322 and is inert until the next roll". The next roll happened: fleet is on v1.0.1323
  (both replicas, pod images read). Ancestry: `cd90e8b27` (wire editor routes), `65f1b0b95`
  (single emitter), `af4743464` (the seam) are ALL ancestors of the v1.0.1323 stamp
  `70e7b4f9c` (found via another lane's commit `49d90b280`, then verified at the artefact:
  `grep -aq "70e7b4f9c" /proc/1/exe` present on BOTH replicas, nonsense control absent).
- `record_absent_required_fields`: **0 agent_definitions rows name it.** Chrome escalation is
  dormant. 7 live steps use `render_site_components` (nav-link-fixer, nav-updater,
  pageflow-builder, rerender-chrome, rerender-pages, rerender-site, site-work-orchestrator).
- `required_fields_missing` items: 45 complete / 30 needs_human_review, ALL from the
  post-deploy check (latest 08-18, all naming deployed `ported-page` rows). Zero render-time
  items yet — expected: editor escalation live only since the 1323 roll this morning, chrome
  half unarmed, and the write is per-edit.

### Misstep (mine, caught in-flight): a VACUOUS census read as a finding

I ran the chrome census (site_components rows missing required source:llm fields), got zero
rows, and stated "no chrome row is missing a required field" — then the vacuity check showed
**candidate_pairs_tested = 0**: no site_components row references ANY component declaring a
required llm field, so the query could never have returned a hit. The zero was true but
meant something different: "the chrome store uses only 0-required components", not "all
required fields are supplied". The census was re-read on the corrected join and the
conclusion for arming purposes survives (0 rows fire either way), but the first phrasing
would have been a WRONG_CALLS row if it had reached a doc. Cheap check, now in RUNBOOK:
**count the candidate pairs your census tested in the same query; a zero over zero
candidates is not a finding.** Logged in WRONG_CALLS.md per the tally rule.

### Demand-control failure worth recording

Pod-log grep for render activity (`RenderTemplate: fields rendered empty|Re-rendered
component|render_site_components|RenderComponentAction`) returned 0 on both replicas while
the DB showed 38 page_components writes since pod start — the greps don't match what these
paths actually log at INFO. So "0 absent-required reports in the logs" could not be given a
log-side demand control; the DB-side census is the evidence that matters and is what the
plan uses.

### Populations re-measured (the bug file's figures had moved)

- Active components: 283 (was 253 at filing). No-schema: 100 — **95 are tools** (schema-less
  by design; `isSelfContainedSection` codifies it). Non-tool no-schema: **5**, one
  page_components usage each; 2 with `{{.field}}` placeholders (`report-request-form`,
  `audience-check-form`). §5's "expect the 75-of-253 to be the hard part" has dissolved into
  small data work.
- Chrome components WITH required llm fields exist in the library (footer-with-disclaimer 17,
  header-with-categories 16, header-with-search 5, header-with-cart-or-nav 4,
  header-minimal-tool 2) but **no site_components row references any of them** — the store
  uses site-header/site-footer/head (0 required). Hence arming the record = 0 items today.
- page_components writer-side census: 131 rows missing a required llm field top-level;
  breakdown: ported-page/body 77 deployed + 23 removed (the post-deploy check's existing,
  already-itemised population), hero/headline 14 deployed, scattered singles after that.
  NOTE this is the WRITER question (pre-gate); the seam's render-time answer is a strict
  subset after `contextToInterfaceMap` defaults.

### Code reads that shaped the plan

- Editor routes: emit item, then **persist the blank anyway** — refusal is the missing half
  (`section_editor_actions.go` applyContentEdit ~:1042, applyComponentSwap ~:1169).
- `ApplySectionEditAction` already has the ONE-persist-switch idiom (link repair :~445,
  envelope refusal :~455) — the refusal gate belongs there, not in the two branches.
- `section-editor.apply_edit` has **no error_step** (live row) → a refusal meets
  `bugs_open/344`'s completion-trample on the DRIVING item. Live page still protected; the
  item status is 344's bug, not this lane's. Interaction stated in the code comment.
- Six unwired sites re-verified: GateConvertedTemplate + tool_birth_instance_scope (raw
  candidate templates, no component row), legacy head render (template load only),
  RenderTemplateWithMap (contact-info block, callers have no component schema),
  offline audit ×2 (probes with fields REMOVED by design). No change owed.
- `assemble_from_library` renders a stitched TEMPLATE (content arrives later) — report-only
  is correct there; no escalation owed.
- RFC_022 budget: `apply_section_edit` counts 7 optional keys; ConfigKeys are not counted by
  the audit; the new key changes nothing. `render_site_components` has NO registered input
  spec at all (pre-existing; not widened by this lane).

## 2026-08-22 late morning — built, tested, submitted, committed

- Refusal half implemented: key + deciding arm in `mistyped_llm_fields_gate.go`
  (`refuse_absent_required_fields`, `refusePersistForAbsentRequired`), outcome field +
  branch copies + ONE gate at the persist switch in `section_editor_actions.go`, ConfigKeys
  declared. Tests: two table tests + a seam→outcome→gate chain test; **mutation-proven**
  (inverting the deciding arm fails TestEditorRefusalNeedsBothArmingAndAFinding AND
  TestSeamFindingSurvivesOntoTheEditOutcome; reverted, green). Honest limit written into the
  chain test's comment: the in-branch copy needs a DB — the post-roll canary covers it.
- Migrations: 550 (chrome record arm, appliable) + 551_HOLD (editor refusal arm, after the
  roll), both with rollback sidecars, DO/RAISE verifies, double-apply refusals; 550 also
  refuses if a render_site_components step has moved into a sub_workflow (the write only
  reaches top level).
- Council: corr `3626629a-f2bc-4089-9118-c1d6dd007807`, submitted 09:32Z, dispatched almost
  immediately (no queue wait this time). Two client-side schema rejects first: `create` is
  not a valid operation (use `add`), and `risks` must be a STRING not an array.
- Committed `0ee442cfb` with `Council-Submitted:` trailer, 13 files, scope report all mine.
  Clean `git archive HEAD` build verified after commit (platform/internal/cmd all compile).
- **My WRONG_CALLS entry (vacuous census) was swept into the 337 lane's commit `9e23fb852`
  as a same-file passenger** between my append and my commit — the exact CLAUDE.md case;
  stated in my commit message, nothing lost.
- MEMORY_workstreams: lane registered with the owed follow-ups (verdict read, 550 apply,
  post-roll 551 + canary, the 5 no-schema components decision).
- 550 deliberately NOT applied pre-verdict: migrations are council scope (314 — live the
  moment they apply), and the arm fires on 0 rows, so waiting costs nothing.

### The 9-of-15 arithmetic re-derived first-hand (not inherited from the bug file)

`grep -rn "RenderTemplate(" --include=*.go platform/ internal/ cmd/ | grep -v _test | grep -v
"func RenderTemplate"` → **14**, plus `RenderTemplateWithMap` = **15**. Schema-wiring sites,
counted with a grep that NAMES THE RECEIVER (the bug file's own landmine: a bare
`grep -c 'InputSchema = '` also matches `ci.InputSchema`, a different struct) → **9**:
assemble_from_library:302, v3_site_actions:2464, section_editor:1069 + :1206,
render_site_components:1051, component_library:1730/1805/2092, rerender_page_sections:655.
**9 + 6 = 15 closes.**

The six unwired, each read at its own call site rather than taken on trust:
- `GateConvertedTemplate(function, converted string, …)` — signature takes a raw string; no
  component row is in scope, so there is no schema to pass.
- `ScopeToolBirthTemplate(html, function string, …)` — same shape, raw candidate template.
- the legacy head render (`rerenderSinglePage` :538) — its loader
  `rerenderLoadHeadTemplate` selects `defaults->>'head'` then `html_template` only, never
  `input_schema`; and **`RerenderSitePagesAction` is in no `GlobalActionRegistry` entry**
  (grep: 0 hits), matching RFC_041 §4 — a dead path, so wiring it would be inert anyway.
- `RenderTemplateWithMap` — a different executor (contact-info block), callers hold no schema.
- the two `cmd/component-render-check` probes — they render with fields REMOVED on purpose, so
  a report there fires on every probe by design.
