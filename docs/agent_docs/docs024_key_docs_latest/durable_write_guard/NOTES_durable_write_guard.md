# NOTES — durable-write completeness guard (bugs_open/021 INSTANCE 1)

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

Scope of THIS workstream: **INSTANCE 1 only** — extend (or decide not to extend)
the `componentRegressionIssues` guard beyond the one path it covers
(`content_components.html_template` via `update_component_html`). INSTANCE 2 (the
verifier-coverage framework) is OWNED by the `work_item_completion_integrity`
thread — do not touch it.

---

## 2026-07-21 — session start, evidence gathering

**Where the guard lives today.** `platform/orchestration/actions/component_write_guard.go`
(`componentRegressionIssues`, pure function, three comparative checks: size
collapse <30%, unterminated balanced tags, mid-token tail). Wired into ONE
caller: `update_component_html_action.go:146`. Escape hatch
`allow_structural_regression` is step-config-only. Rejections land in
`agent_error_log` (`error_code='component_write_regression_blocked'`).

**The handoff's named targets, evaluated against the live system:**

### Target A — `pages.rendered_header` / `rendered_footer` / `rendered_head`
**FINDING: dormant columns, NOT an exposure.** No Go code anywhere in
`platform/`, `pkg/`, `internal/`, `cmd/` writes these columns (exhaustive grep of
the three literals across `*.go/*.sql/*.sh/*.py`). The only Go touch is a READER:
`discovery_checks/check_missing_structure.go:94-100` flags them `IS NULL`. DDL
(`sql_for_tables/003_pages.sql:203-205`) added the columns; nothing populates
them. Corroborated by `content_quality_and_internal_linking/README_find_phantom_links.sql:13`
("pages.rendered_header was empty").
Page chrome actually lives in `site_components.rendered_html`
(keyed `site_id`+`slot_name`), assembled deterministically at render time
(`rerender_single_page_action.go:329-356`), never stored back into `pages.*`.
→ A guard on `pages.rendered_*` would protect a write that never happens.
**[VERIFIED by grep + column DDL; the "never written" claim still to be
double-checked against live data — see the count query in RUNBOOK before final
sign-off.]**

### Target B — `page_components.rendered_html`
**FINDING: a DERIVED render of the already-guarded template, recoverable — with
one birth-path nuance.**
- The platform's own design principle is stated in the guard's caller,
  `update_component_html_action.go:248-253`: *"Do NOT set rendered_html here …
  rendered_html on page_components is the per-page render after variable
  substitution and cannot be recreated by copying the new template verbatim. The
  rerender pipeline regenerates rendered_html correctly."* So `rendered_html` is
  the **projection**; `content_components.html_template` is the **source**.
- The render mechanism: `section_editor_actions.go` loads `html_template`
  (line 603), renders with `content_data` (624-629), UPDATEs
  `page_components.rendered_html` (930-961). For a tool `content_data={}` so the
  render is ~identity over the template (tool-doc header retained,
  `platform/content/tool_doc_header.go:11-12`).
- **The tool BIRTH path writes the whole LLM artifact into BOTH columns.**
  `create_tool_component_action.go`: LLM `html_content` → INSERT
  `content_components.html_template` (186-199) AND INSERT
  `page_components.rendered_html` (241-246). `deploy_tool_action.go` forks the
  template then writes it to `rendered_html` (332-344). So `rendered_html` CAN
  receive a whole-artifact write directly — but the same bytes are in
  `html_template` in the same operation.
- Recovery: `page_component_history` (9,933 rows, fresh to 2026-07-21) snapshots
  **`content_data` only**, NOT `rendered_html`. But `html_template` has
  `component_versions` (the 012 recovery table). So a truncated `rendered_html`
  is recovered by **re-rendering from the guarded, versioned `html_template`**;
  it is unrecoverable only if `html_template` itself is truncated — which is
  exactly what the existing guard already prevents.
- The one confounder: `save_page_sections_action.go:276-285` claims interactive
  tools "exist ONLY as rendered_html … not LLM-regeneratable". Agent trace says
  this is scoped to the rebuild-from-spec path, which carries the existing
  `rendered_html` forward without re-reading `html_template` — authoritative
  *within that path*, but the durable bytes still sit in `html_template`.
  **[This claim needs its own verification before it can be dismissed — is there
  a live tool whose html_template does NOT reproduce its rendered_html? See
  RUNBOOK query. If one exists, rendered_html IS a durable source for it.]**

**Provisional read:** the strongest handoff target (`pages.rendered_*`) is inert,
and the other (`page_components.rendered_html`) is mostly a recoverable
projection. The real residual risk, if any, is (i) the tool BIRTH path
(`create_tool_component` / `deploy_tool`) writing an unguarded whole LLM artifact,
and (ii) whether `site_components.rendered_html` (the chrome) is itself an
LLM-whole-artifact write that shares the 012 shape. Both need the writer
classification (agent 1) before I commit to a scope recommendation.

**Awaiting:** agent 1 = full classification of every `rendered_html` writer
(class A = whole LLM artifact / class B = deterministic render / class C = other).

## 2026-07-21 — Target A + B VERIFIED against live data

**Target A — CONFIRMED inert.** `SELECT count(*) FILTER (rendered_header/footer/head
non-empty) ... FROM pages` → **`0 | 0 | 0 | 301`**. All three columns empty across
every one of 301 pages. Grep (no writer) + live data (all NULL/empty) agree: a
guard on `pages.rendered_*` protects a write that does not occur. **Dismiss this
target.**

**Target B — CONFIRMED recoverable.** For every tool component, `rendered_html`
length vs its durable `html_template` length:
```
tool-ai-readiness-quiz            24466 / 24466  = 100%
tool-grip-force-friction-calc     23874 / 24278  =  98%
tool-arena-interface              23353 / 23353  = 100%
tool-matchmatrix                  23281 / 23281  = 100%
tool-llm-cost-calculator          21724 / 17828  = 122%   ← render EXPANDED template
game-master-explanation           11570 / 10213  = 113%
...  (25 rows, range 98%–122%, most exactly 100%)
```
The render is ~identity over the template (variance = the fixed tool-doc header +
content_data substitution). **The template is ALWAYS present** — so the
`save_page_sections:276-285` "exists ONLY as rendered_html, not LLM-regeneratable"
claim is false as a statement about the durable store: the LLM-regeneratable bytes
are sitting in `content_components.html_template`, which is guarded + versioned.
A truncated `rendered_html` is recovered by re-render. **Dismiss this target too**
— with one caveat carried forward: the few rows where rendered_html > template
(122%) mean the render is deterministic-LARGER, not that rendered_html holds
un-regeneratable content; still recoverable, still not a durable source.

**So both handoff-named surfaces are non-exposures.** The handoff's *generic
mechanism* ("any whole-artifact overwrite from LLM output with no comparison")
remains valid in the abstract — the open question is whether ANY genuine instance
of it survives that is NOT already `update_component_html`. The remaining
candidates are the BIRTH paths (`create_tool_component`, `deploy_tool`,
`store_generated_component`) — but those INSERT with **no prior row to compare
against**, so the *comparative* guard cannot apply by construction; they need an
absolute structural check, which is a different (and partly-existing) mechanism.
Checking birth-path gate coverage next + awaiting agent 1's full writer census
(esp. whether `site_components.rendered_html` chrome is a whole-LLM-artifact write).

## 2026-07-21 — site_components.rendered_html (the chrome) is NOT the 012 shape

Read `fix_component_template_action.go` (the main site_components.rendered_html
writer, 4 write sites: lines 276/452/528/598). Its header (lines 8-28) is
explicit: the three fix types that touch `site_components.rendered_html` are
`inject_nav_flex_css` (adds `display:flex` CSS), `responsive_fix` (adds media
queries), `remove_element` (regex-removes an element). All **deterministic string
edits**, not whole-LLM-artifact overwrites — line 452 is literally an APPEND
(`rendered_html = rendered_html || $3`). And lines 25-28: *"rewriting rendered_html
is acceptable for site_components because they are re-rendered from templates."*
So site_components.rendered_html is (a) written by deterministic edits and (b)
itself a re-renderable projection. **Not a 012-shape exposure.**

## Emerging picture (to confirm against agent 1)

The 012 shape is specifically: *a whole LLM completion OVERWRITES an existing
durable row with no comparison*. Across the surfaces examined so far it exists at
**exactly one overwrite path — `update_component_html` → `content_components.html_template`
— which is already guarded.** Everything else is either:
- a **deterministic render/edit** (section_editor render, fix_component_template
  edits, site_components chrome) → re-renderable, no LLM truncation to persist;
- a **dormant column** (pages.rendered_*, 0/301);
- a **BIRTH-path INSERT** (create_tool_component / deploy_tool /
  store_generated_component) → no prior row, so the *comparative* guard is
  inapplicable by construction, and a truncated birth makes a NEW bad component
  (recoverable/re-runnable) rather than destroying a good one. That residual maps
  to fix-candidate (b) — stop_reason truncation detection — already BUILT
  (f32b208e5) for GenerateText callers.

If agent 1 confirms no OTHER overwrite path takes a whole LLM completion into a
durable row, the correct outcome is NOT a second comparative guard but a
**documented evaluation** that the named surfaces are non-exposures, plus (maybe)
a small absolute mid-token/unterminated-tag check on the birth INSERTs if they
lack one. That is a scope call for the owner — the handoff explicitly asks a human
to decide scope before code, and this finding contradicts the handoff's premise
("the same shape is unguarded elsewhere"), so it must be surfaced, not silently
actioned.

## 2026-07-21 — agent census returned + reconciled with LIVE evidence (bug 046)

**Agent census of every `rendered_html` writer (cited file:line, both tables):**
NONE matches the 012 overwrite shape ("whole LLM completion overwriting a durable
row with no comparison"). Grouped:
- Deterministic RENDERS (`section_editor` render output, `rebuild_blog_listing`,
  `render_site_components`) — a comparative guard would fire on legitimate render
  output. Must NOT touch.
- Deterministic in-place TRANSFORMS (`fix_harcoded_colours`,
  `fix_forced_text_colours`, `fix_component_template`, `link_site_components`
  writes NULL) — read-modify-write existing HTML, write only on change. Must NOT
  touch.
- Template COPY / first-insert (`deploy_tool` copies a stored template).
- The ONE class-A source reaching a durable column: **`create_tool_component`**
  (writer #3) — a whole LLM tool artifact, but a **first INSERT into a newly
  created page**, never an overwrite. Gated only by `HasToolDocHeader`.
- `save_page_sections` is the only DELETE-all+INSERT writer; its inputs are
  render/assembly output and it already snapshots to `page_component_history` +
  refuses `rebuild_policy=owned`. A guard here would be a **section-count
  regression** check, a different shape from `componentRegressionIssues`.

**The reconciliation — bug 046 is the live proof of where the real hole is.**
046 (travelling-docs, filed 2026-07-20; live HTTP evidence): **8 tools + 1 section
serve unterminated JavaScript on 6 live customer domains NOW**, born truncated
(7 of 8 have 0 `component_versions` → never had an intact version → they were
*generated* broken, not overwritten). This is the 012 shape at BIRTH, not at
overwrite.

**Why they were born broken — the precise gap, code-verified:**
- `create_tool_component_action.go:105` gates the birth write on
  `HasToolDocHeader` ONLY. That function (`tool_doc_header.go:37`) checks the
  header SENTINELS are present (opener before closer) — the header sits at the
  TOP of the `<script>`. **A tool cut mid-`<script>` at the TAIL keeps its header
  and passes.** Lines 97-100 of the action say it outright: *"the generator
  workflow has no check_tool_completeness step."*
- The whole `htmlContent` is then INSERTed into `content_components.html_template`
  (186-199) and `page_components.rendered_html` (241-246).
- `toolTemplateValid` (`plan_sections_action.go:1046`) IS the correct absolute
  tail-cut check (reuses the guard's `balancedPairs` + `endsCleanly`), live in
  v1.0.1140 — calibrated against all 27 tools, the 8 truncated ones fail, 0 FP.
  **But it is only called at the schema-LOAD path** (`componentTemplateValid`,
  which decides re-render), **NOT at the birth WRITE.** So a truncated tool can
  still be BORN; it just won't later be re-rendered.

**Section-birth path (`store_generated_component`):** gates on
`scoreComponent(...).TemplateClosed` (line 264/270). Per
`component_write_guard.go:84-86`, TemplateClosed requires balanced `<section>`
tags ONLY — it does not check `<script>/<style>` balance, so a section whose
`<script>` is cut but whose `<section>` closes passes. 046's `section` casualty
(archetype-taster-quiz) is consistent with this. Secondary weakness.

**Calibration boundary (046 plan, lines 13-17) — load-bearing for any birth gate:**
the 5-pair TAG-IMBALANCE predicate catches exactly the 9 casualties fleet-wide,
**0 over-fire**; the `ends-mid-token` heuristic adds **36 false positives
fleet-wide**. So: `toolTemplateValid` (tag-balance + endsCleanly) is safe on the
TOOL population it was calibrated on (create_tool_component only ever handles
tools). A BROADER birth gate (sections) must use tag-imbalance ALONE.

**Division of labour is already documented and non-competing.** 046's own PLAN
(line 71) puts the "Render-surface write guard → bugs_open/021 (owned)" and keeps
only the SWEEP (restore grip-force, surface the other 7, new `truncated_component`
discovery check). So 046 = clean up existing casualties; **021 (this workstream)
= PREVENTION: gate the birth write so no new truncated tool is born.** Exactly the
stop-new / clean-old split that 012 established.

### CONCLUSION (the scope call, to put to the owner)
1. **DISMISS** the handoff's literal targets: `pages.rendered_*` (dormant, 0/301)
   and a comparative guard on `page_components.rendered_html` (inert vs
   deterministic renders, WRONG vs deterministic CSS fixes). Evidence: agent
   census + live data + the platform's own code comments.
2. **THE REAL FIX = gate the tool BIRTH write.** Wire the already-live
   `componentTemplateValid`/`toolTemplateValid` into `create_tool_component`
   (hard gate, after `HasToolDocHeader`, before INSERT). Same predicate at write
   AND load seams → a tool that would be dropped at load can never be born.
   Reuses live+calibrated machinery; closes the hole that born 046's 8 tools.
3. **SECONDARY (owner's call):** strengthen `store_generated_component`'s section
   gate with the 5-pair tag-imbalance check (0 over-fire fleet-wide) so a
   `<script>`-cut section can't be born either.
This is a scope call because it (a) contradicts the handoff premise and (b)
changes WHERE the guard goes (birth write, absolute) vs what the handoff asked
(rendered_html, comparative). Surface to owner before coding.

## 2026-07-21 — owner chose Phase 1 + section gate; IMPLEMENTED + committed `ba702c8c6`

Owner scope decision: **Phase 1 (tool birth gate) + Phase 2 (section birth gate).**
Shipped (all in package `actions`, inert until an image roll):
1. `component_write_guard.go` — new pure helper `hasUnbalancedStructuralTags`
   (absolute 5-pair tag-imbalance; NO endsCleanly, per the 046 +36-FP calibration).
2. `create_tool_component_action.go` — after the `HasToolDocHeader` gate (which
   can't see a tail cut), a hard `componentTemplateValid(htmlContent, "tool")`
   gate → `recordComponentWriteRejection` (`tool_birth_truncation_blocked`) →
   error routes to `needs_human_review` via the workflow's existing `error_step`.
   Placed with the other raw-input validation (before any DB work).
3. `store_generated_component_action.go` — added to the existing `blockingIssues`
   structural checks: `len>=100 && hasUnbalancedStructuralTags` → rejects a section
   whose `<script>/<style>` is cut but whose `<section>` closes upstream (the
   TemplateClosed blind spot; 4 of 046's 8 casualties).
4. `component_write_guard_test.go` — `TestHasUnbalancedStructuralTags`, incl. the
   046 discriminator "cut `<script>` after a closed `</section>`". All existing
   `toolTemplateValid`/`componentTemplateValid` tests unchanged and green.
Build OK; `go test ./platform/orchestration/actions/` = ok.

### MISSTEP navigated — a same-file passenger in plan_sections_action.go
Originally I ALSO refactored `toolTemplateValid` (its inline 5-pair loop → a call
to the new `hasUnbalancedStructuralTags`) — a cosmetic DRY change. When I went to
commit, `git diff --stat` showed `plan_sections_action.go | 53 +++++`, far more
than my ~4-line edit. The diff revealed **another session's uncommitted WIP in the
same file**: `aliasNormalisedSectionKeys` (bugs_open/041) + an `isSelfContainedSection`
exemption in `planSection` (bugs_open/044). A pathspec commit of that file would
have swept their work under my message (the same-file passenger CLAUDE.md warns
about — no hook can prevent it), and `git checkout` would have destroyed their WIP.
**Resolution:** reverted ONLY my hunk (Edit the loop back), leaving the file to the
041/044 session, and did NOT commit it. Cost: `toolTemplateValid` keeps its own
inline loop instead of sharing the helper — purely cosmetic, functionally
identical, and both still key off the same `balancedPairs`. Lesson reaffirmed: the
shared tree is live mutable state; check the per-file diff size against your own
edit before committing, and never fold a cosmetic refactor into a file you don't
otherwise need to touch.

### Observation (not mine to fix) — a pre-existing broken test at HEAD
`discovery_checks/verifier_coverage_test.go::TestEveryCheckProducedItemTypeIsClassified`
FAILS at committed HEAD (reproduced with my changes stashed): two committed checks
— `contact_form_undeliverable` (3913a0adf) and `backend_entry_orphaned`
(7b03f296a) — were added without a verifier or a classification entry. That is the
INSTANCE 2 coverage guard (owned by `work_item_completion_integrity`) doing its
job on OTHER threads' omissions. Not this workstream's; the `actions` package
(all my changes) passes clean. Flagged here for the record only.

### Still to do
- Build + roll a chassis image (owner-sanctioned — outward-facing), then VERIFY by
  fault-injection per the PLAN (drive a tail-cut tool through create_tool_component
  → not created, item `needs_human_review`, refusal in `agent_error_log`), plus a
  discriminating pod-grep of the CREATED literal `tool_birth_truncation_blocked`.
- Optional: advisory council gate on the platform change (credits + ~30 min).
- Until it ships, INSTANCE 1 stays OPEN (bar = fixed AND live).

## 2026-07-21 (later) — the fix is ALREADY LIVE in v1.0.1146 (rode a sweep build)

I did not have to build or deploy. Between committing my code (`ba702c8c6`) and
committing my docs, the owner ran a sweep build **`fe2ba5e52` "v1.0.1146 - sweep"**.
That commit is a descendant of my code commit (`git merge-base --is-ancestor
ba702c8c6 fe2ba5e52` → true), the makefile IMAGE_TAG is now `v1.0.1146`, and the
running pod is on `v1.0.1146`. So my birth gates rode that build — the exact
"committed code rides ANYONE's next sweep build" pattern the memory notes and
CLAUDE.md warn about. (My three doc files were ALSO swept into `fe2ba5e52` — the
`git add -A` sweep took my half-finished docs too; nothing lost, forward-only
holds, only the README remained for my own commit `213e3eb4d`.)

**Pod-verified (discriminating grep, pod on v1.0.1146):**
```
tool_birth_truncation_blocked              -> 1   (my Phase 1 error_code, CREATED)
"generated HTML is structurally incomplete"-> 2   (my Phase 1 msg + returned err)
"leaves a structural tag"                  -> 1   (my Phase 2 section msg)
component_write_regression_blocked         -> 1   (positive control = 012 guard, live)
```
So DEPLOYMENT is proven. **CORRECTNESS is NOT** — per [[verify-the-failing-branch]],
a pod-grep proves the string shipped, not that the branch fires. The guard's whole
job is to DETECT a truncated birth, so it must be verified by INDUCING one, not by
a green happy path. Status: **fix LIVE in v1.0.1146, failing branch
LIVE-UNEXERCISED.** (Same honest register another thread used for bugs_open/010 on
this very image — `9a525d46a`.)

**To fully close** (fault-injection, has real cost/latency — put to owner):
dispatch `create_tool_component` (via the tool-generator workflow, or a scratch
orchestration) with an `html_content` that has a valid tool-doc header but an
unbalanced `<script>` (tail-cut). Expect: component NOT created; work item
`needs_human_review`; `agent_error_log.error_code='tool_birth_truncation_blocked'`.
Then a healthy generation must still be created. Mind the CLAUDE.md timing traps
(~300s after a chassis restart; ~30 min dispatch queue latency). Clean up scratch
fixtures after.

**OWNER DECISION 2026-07-21:** leave it **live-but-unexercised** for now — no
scratch dispatch this session. 021 INSTANCE 1 stays OPEN with that status. The
fault-injection above is the one remaining step whenever it's picked up (or a real
truncated generation trips it in the wild and lands the `agent_error_log` row).

## 2026-07-23 — FAULT-INJECTION DONE, failing branch PROVEN live on v1.0.1149

Owner said go. First re-verified my code survived the new roll (image v1.0.1149,
pod `…-cm786`): all 3 created literals + the 012 positive control present. No
regression-out.

**Harness (controlled, LLM bypassed).** The real dispatch can't inject a
"complete-looking but structurally broken" HTML — the `stop_reason` guard
(bugs_open/008) would intercept a provider-truncated generation upstream, testing
the wrong layer, and an unresolved-string config value is NOT taken as a literal
(`action_inputs_literal_test.go`). So I fed the HTML through the payload: a scratch
one-step agent_definition `scratch-toolgate-021` whose sole step is
`create_tool_component`, reading `site_id/html_content/function/display_name` from
`input_data.*`. Spawned via the kcat `orchestrate` envelope (mirrors 091). Runs on
the generic chassis (create_tool_component is IsLocal); processed in ~2.4s — no
queue latency. My gate sits at line ~110, BEFORE the site-domain load (line 133),
which enabled a zero-cleanup design: nothing is ever created on either path.

**NEGATIVE — tail-cut tool (valid tool-doc header, `<script>`+`<section>`
unterminated, ends `'Epic`), `orch 937888f4`:**
- state: `save_tool → complete_error → COMPLETED` (the guard's hard error routed
  through the workflow's error_step).
- `agent_error_log`: **`tool_birth_truncation_blocked`** (severity error, action
  `create_tool_component`, step `save_tool`) — *"tool birth refused for site …:
  generated HTML is structurally incomplete — a paired tag
  (<script>/<style>/<section>/<div>/<fieldset>) is unterminated or it ends
  mid-…"*; context `{html_length:368, ends_cleanly:false}`. Plus the workflow-level
  `UNKNOWN` row recording the returned error (*"refusing to persist truncated
  tool …"*).
- `content_components` with the scratch function: **0**. Nothing persisted. ✅

**POSITIVE control — healthy tool (fully balanced, ends `</section>`) + a
NONEXISTENT site_id, `orch 0042b1e7`:**
- state: `save_tool → complete_error → COMPLETED`, but the error is **"failed to
  load site domain: sql: no rows in result set"** — a DIFFERENT error, at the
  site-load step that runs only AFTER my gate. So the healthy tool **passed the
  gate**.
- `agent_error_log`: NO `tool_birth_truncation_blocked` (only the `UNKNOWN`
  site-load row). The gate did not false-block a healthy generation. ✅
- component created: **0**.

**So both branches are proven on the live binary:** the guard fires on a tail cut
(refuse + log + no persist) and passes a healthy tool. This satisfies
[[verify-the-failing-branch]] — the fault was INDUCED, not inferred from a green
happy path.

**Cleanup:** leak check 0 (no components/pages/work-items/scratch-site rows), then
deleted the scratch agent_definition, both orchestration_states (+12 audit rows),
and the 3 scratch `agent_error_log` rows (so the immune-system sweep doesn't treat
a scratch test as a real failure). Verified 0 remaining. Evidence preserved here
because the DB rows are gone.

**STATUS: INSTANCE 1 is FIXED + LIVE (v1.0.1149) + BEHAVIOURALLY VERIFIED.** The
021 file stays in `/bugs_open/` ONLY because INSTANCE 2 (not ours) is still open;
INSTANCE 1 itself is done.

## 2026-07-25 — clearing the INSTANCE 2 residue, and closing 021

Owner pointed this session at 021 with "check other threads are not already
looking". Coverage check first, all three clear: `who-owns.py 021` names
`durable_write_guard` (this lane) and `work_item_completion_integrity` (dormant
since 07-20, and 07-24's work was written INTO their docs, not forked);
`site_work_items` has no open `needs_diagnosis` or in-flight item touching the
verifier lane; `git status` shows none of the four INSTANCE 2 files dirty.

### The council verdict we were waiting for never existed

Corr `56c7e177` had no `diagnosis_artifacts` rows at all — which reads exactly
like "still queued", and is why it sat unchecked overnight. It was not queued.
The orchestration ran at 21:14:19 UTC on 07-24 and was **dead 6 seconds later**:

```
current_step=complete_invalid  status=COMPLETED  completed_steps=0 of 42
collected_data->'__step_error':
  "step persist_submission failed: failed to execute action
   diagnose_persist_fix_plan: plan failed validation:
   edit 3: operation \"create\" not in the allowlist"
```

The allowlist is `modify | add | remove | config_change` (`allowedFixOperations`,
`diagnose_persist_fix_plan_action.go:80`); a new file is `add`. **No reviewer ever
saw the plan.** The RUNBOOK_council_gate now documents this trap — two other
threads hit the identical pair (`risks` as an array, `operation: "create"`) on
07-25 and wrote it up hours apart; ours on 07-24 was the earliest instance and
nobody had it written down yet. Resubmitted 07-25 with that one field corrected
and the new live evidence appended, `RESUBMIT_CORR=56c7e177…` so the trail
accumulates.

**Transferable:** a council run with no artifacts is ambiguous between QUEUED and
INVALID, and the two want opposite responses (wait vs fix-and-refire). Always
poll `orchestration_states.current_step` alongside, and treat `complete_invalid`
as terminal.

### INSTANCE 2's verifier is LIVE — and BEHAVIOURALLY VERIFIED

Pod `agent-chassis-774877f4c6-zjh4t`, image **v1.0.1159**: the literal this
change CREATED (`"no unlocked component carries a colour within the fixer's
remit"`) present ×1, positive control `tool_birth_truncation_blocked` ×1. So
`34adb171c` is no longer inert — it rode a roll between 07-24 and now.

Deployment is not correctness ([[verify-the-failing-branch]]), so both branches
were exercised. The method that made the probe a **discriminator** rather than a
green happy path: dump the verifier's whole live population (32 components,
8 sites) and run a **verbatim copy** of `ReplaceHardcodedColors` over it in a
scratch stdlib-only `go run`, so the expected verdict is known before firing.
Result — the remit is much narrower than the detector, and on most sites it is
EMPTY:

| site | detector pop | inside fixer's remit |
|---|---|---|
| robot-hands.com | 3 | **3** |
| gamesdesign.co.uk | 4 | 1 |
| leopardessconsulting.co.uk | 4 | 1 |
| finetuning.uk | 8 | **0** |
| gaswholesalers.com | 6 | **0** |
| ai-agent-orchestration.com | 4 | **0** |
| webdesign.co.uk | 2 | **0** |
| dartsonline.com | 1 | **0** |

That table IS the argument for the 07-24 design decision: a verifier re-running
the detector's predicate would refuse every completion on 5 of 8 sites for ever.

**NEGATIVE branch (must refuse) — robot-hands.com, 3/3 in remit.** Scratch
one-step `complete_work_item` orchestration (`7d4257ea`, 17:12:02 UTC):
`claimed → triaged`, `attempt_count 0 → 1`, and

```
error:  completion blocked: post-fix verification found the defect still present:
        3 component(s) still carry colours the fixer's own transform would
        replace (first: tool-matchmatrix/tool-matchmatrix)
result: {"_verification": {"status": "defect_persists", "item_type":
        "hardcoded_section_colors", "detail": "3 component(s) …"}}
```
The count and the named component match the locally-computed prediction exactly —
the probe was graded, not merely observed. A second identical fire (`c26462b8`)
reproduced it (`attempt_count → 2`).

**POSITIVE branch in PRODUCTION, not a fixture.** Found while censusing
`_verification` records: `site_work_items 51054090-1b63-431d-aa55-0c6a873ff47a`
(vonc.com) completed **2026-07-25 10:18:52 by `build-dispatch-loop`** carrying
`_verification.status=verified`. The gate is being exercised by live dispatch
traffic, unprompted. **Caveat, stated because it matters:** vonc.com has ZERO
detector matches, so that pass is trivially true and does NOT discriminate
remit-from-predicate.

**POSITIVE branch, DISCRIMINATING — finetuning.uk, 8 detector matches / 0 in
remit** (`5ea236ae`, 17:21:39 UTC). Completion **allowed**: `status=complete`,
`attempt_count` still **0**, `_verification.status=verified`, detail *"no unlocked
component carries a colour within the fixer's remit (out-of-remit hexes — light,
3-digit, inline style attributes — may legitimately remain)"*. This is the whole
design decision proven on the live binary: eight components that the DETECTOR
matches, and the gate correctly does not hold the item hostage to them. A verifier
built the obvious way — re-run the detector — would have refused this completion
and, at `max_attempts`, stranded it in `failed`.

**Both branches on one image, graded against a prediction, one refusing and one
passing on the SAME predicate.** That is the bar from [[verify-the-failing-branch]]
met properly, not a green happy path.

**Cleanup:** 2 work items, 1 scratch `agent_definitions` row, 3
`orchestration_states` + 21 audit rows deleted; leak check returns 0 on all five
counts, including `agent_error_log` (which was 0 throughout — this gate refuses
via the completion path, not via an error row, unlike INSTANCE 1's).

**Containment note worth keeping:** the fixture had to sit on a REAL production
site, and a refusal releases the claim (`triaged`, `claimed_by` NULL) — so the
fixture became dispatchable to a live handler on robot-hands.com. Giving it
`handler_agent='scratch-021-nonexistent-agent'` is what saved it: the dispatch
loop picked it up **within 5 seconds** and parked it `blocked` ("Handler agent not
registered: …") instead of running a real colour fix on a real site. Recorded in
the RUNBOOK; do not run this probe with a plausible handler name.

### The measurement that did NOT hold up

The 07-24 note cited "21 complete / 7 unresolved / 5 failed" for this item type.
Live today the whole type is **13 real rows** (4 complete, 8 unresolved, 1
detected). I did not reconcile the difference and I am not going to assert one —
`site_work_items` rows are known to be pruned, but I have not proved that is what
happened here, so treat the 07-24 figure as unverifiable rather than wrong.
`[UNVERIFIED]`. What the fresh count does support is the same conclusion by a
different route: 8 items sit permanently `unresolved`, and on 5 of the 8 sites in
the detector's population the handler's remit is EMPTY, so no handler run can
ever clear them. That is the finding, and it is now filed as `bugs_open/077`
rather than left inside a closing bug file.

**Lesson, again:** a figure carried forward from another day's note is a claim,
not a measurement. Re-run the count before you repeat it — even when you are the
one who wrote it.

### Council: APPROVED round 1 — and the one medium objection was our own fault

`56c7e177`, 17:32:21 UTC (10 minutes after the resubmission cleared
`persist_submission`). **approved**, 12 reviewers, 4 abstained, `unreadable: 0`,
*"approved with 2 advisory objection(s) — none high-severity"*.

**No trailer, deliberately:** the code commit `34adb171c` is a day older than its
verdict, and forward-only forbids an amend — so this pair is a permanent `098`
false negative, recorded in the bug file so the coverage gap reads as explained
rather than unreviewed.

**`bug_historian`'s medium objection is REFUTED, and we manufactured it.** It
argued the verifier filters `pc.locked_at IS NULL` while the detector — *"quoted
in grounded_in"* — does not, so a locked in-remit component would be silently
excluded and the item falsely verified. The two queries are **byte-identical**,
both carrying `AND pc.locked_at IS NULL` (`:100` and `:214`). The reviewer
reasoned correctly from evidence *we* gave it: the `grounded_in` entry rendered
the detector SQL abbreviated, dropping the `locked_at` line.

**Transferable, and it cost a medium objection: an abbreviated quote in
`grounded_in` is not a shorter quote, it is a DIFFERENT claim.** Reviewers cannot
open the file; an ellipsis is an implicit assertion that nothing load-bearing was
elided. Paste the whole predicate. Note the shape — this is the same failure as
the two wrong calls above (`-c 1`, the 07-24 counts): reasoning confidently from
a partial view of the evidence. Three times in one session, twice by me and once
by a reviewer I misled.

The other `bug_historian` objection (*"fixes one item type, ~68 still complete on
self-report"*) is the standing position and is answered by the design: coverage
is 3 of 77 deliberately, the coverage map is the build-enforced backlog, and the
held `page_rerender` verifier is the proof that writing verifiers faster than you
can scope them to their handler's remit makes things worse.
