# 016b — Debugging Guide (Vol. 2)

Continues `016_debugging_guide` (closed at v2_56, ~2500 lines). 016 holds the full
back-catalogue of failure patterns and their fixes; consult it for anything not covered
here. New debugging entries go in THIS file from 2026-06-22 onward.

*v1 (2026-06-22): opened as the continuation of 016. Carries forward the orientation, the
durable invariants/heuristics, and the open threads (Parts 2/3/4 of the page
build/re-render work). No new failure patterns yet — the most recent one (interactive page
rebuilt as content) is the last §9 entry in 016 v2_56.*

*v2 (2026-06-24): added a "Wrong turns" log (the false leads taken during the page-pipeline
triage) and two new **triaged** threads — `system-stats` content↔template key mismatch, and the
`rerender_single_page` invalid `page_id`. All active threads marked **triaged** (root-caused or
narrowed to a named remedy, not yet applied/deployed); the Part 4 fix is now written (un-deployed).*

*v3 (2026-06-29): added §9 pattern "Data-driven component shells render empty — inline `<script>` never extracted to js_content" (provocation-card / lobby-grid; stored rows aren't `separateInlineJS`'s output, and provocation-card is additionally truncated mid-script). Fix = regenerate through the current store path + harden validation/extractor. Added the negative-result heuristic.*

*v4 (2026-07-02): three additions from the scheme-to-components arc — §9 "Light site renders dark chrome" (two assembly paths with different chrome sources; stale `site_components`; the provenance greps + legacy-variable tell; rerender re-fossilises, only a full page-build rebuild re-renders templates); §9 "SQL verification pitfalls" (Postgres regex quantifier ≤255; `substring` returns the first capture group; the needle-gate template-surgery pattern); §9 "`sites.status`" (writer + vocabulary; 'active' is legacy; nothing filters on it — never scope blast-radius by status='active'; plus the set_updated_at reuse-gate). Fix work itself lives in RUNBOOK/SPEC_scheme_to_components.*

*v5 (2026-07-06): four additions from the diagnosis-loop persist_note wiring arc —
§9 "error_step: config-level placement, existing-target requirement, derive-from-
next_step" (incl. the loop-substep prefix corollary and dormant step-LEVEL
instances observed in tool-recreation-handler / tool-auditor); §9 "anchorless
diagnosis dies at load_runtime" (an optional evidence tier made mandatory in
practice; error-route to the gather step's own next_step; live-validated ×5 per
run); §9 "pod label key is agent-type (hyphen) + multi-pod log residue"; §9 "two
failure envelopes — a COMPLETED parent does not mean the child succeeded".
Feature/rollout state lives in RUNBOOK_travelling_docs.md; travelling-docs design
in PLAN_travelling_docs.md.*

*v8 (2026-07-09): error containment does not protect against a hang (missing deadline on rag_index's embedding call).*

*v7 (2026-07-09): prompt-template vs config-path resolvers (TEMPLATE_FIELD_ERROR), plus two schema traps (agent_error_log.orchestration_id is text; provenance stamps the generic chassis).*

*v6 (2026-07-08): two entries from the tool-generation incident — §9
"agent_error_log is the first read" (it outlives the pod and names the failing
step; polling current_step attributes the failure to the wrong step) and §9
"code ahead of DB: schema drift, SQLSTATE 42703, latent until the first
caller".*

---

## Orientation (read this first)

The platform builds and maintains multi-page websites with autonomous **agents** — rows
in `agent_definitions` that run as Kubernetes pods (namespace `ai-persona-system`),
communicate over Kafka, and persist run state in Postgres `clients_db`. Each unit of work
(build a page, swap a hero image, re-resolve links) is a row in `site_work_items`, claimed
off `build-dispatch-loop` and routed to a handler agent; the handler usually spawns child
agents (e.g. `page-content-writer`, `page-rerender`) whose state lives in
`orchestration_states` (linked by `parent_orchestration_id`). Deployment is image-tag
based: code ships GitHub → Actions → Backblaze into a new chassis image, then each agent's
`image_tag` is bumped to adopt it; workflow (`default_config`) changes are DB-only and take
effect immediately. The operator runs all SQL / `kubectl` / `git`; the agents do the building.

DB access: `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`.

---

## Durable invariants and heuristics (the lessons that keep paying off)

These are distilled from 016. When an investigation stalls, it is almost always because one
of these was skipped.

**Trust the rendered artefact, not the status.** A work item `complete` — or a green git
commit — is NOT proof the work happened. The silent-completion family (016 §9) recurs:
a result-contract drop replaced output with a stub that still said `status: completed`;
a content-regression guard error was masked by `complete_error` as success; a pod dying
mid-flight left `complete` with a non-empty `error`. Verify against `page_components`
(timestamps + `rendered_html`) and the live page, not the work item.

**`completed_at` is the orchestration END, not the write instant.** The actual page write
happens mid-run inside a child orchestration. Never exclude a candidate work item on a
near-miss timestamp. To find who wrote a page/row, trace `orchestration_states` filtered by
the target id in `collected_data` over the window — the writer is the child whose
`owner_agent_type` + mid-run timing line up. (016: "recurring debugging trap" parts 1–3.)

**A config key read on a different path than it's set is a silent no-op, not an error.**
`output_field` vs `output_fields`; `max_tokens`/`ai_service` shadowing; dispatch
`input_mapping` path mismatches. When a handoff "completes" but the payload isn't there,
compare the producer's stored output to the consumer's received field by exact path before
suspecting generation, compile, or a guard.

**Know who writes `page_components`.** Only `save_page_sections` writes it (DELETE+INSERT),
recording a row in `page_component_history` with `source` set but `source_item_id`
currently NULL on the overwrite path (a traceability gap — see Part 4). `page-content-writer`
regenerates a page from `plan_sections` and does NOT preserve interactive sections absent
from the plan. Interactive game/tool pages store the whole tool as a section's
`rendered_html` (often the hero slot, ~18 KB of `tool-page` markup), not as a separate
library component — so a writer run on such a page silently drops the game.

**0 rows is not decisive.** Check the query/path before concluding absence — a parent-side
JSON path queried against a child row, a name that doesn't match, a wrong timestamp window,
or a column in another schema all return 0 rows. (016 Schema reminders; assumption checklist.)

**A negative inference from an artefact's shape needs the mechanism checked in ALL cases.**
"The raw inline `<script>` is still in the template, so `separateInlineJS` never ran" only
holds if the extractor would have transformed *every* such script. It wouldn't: its regex
(`<script\s*>`) matches only ATTRIBUTE-LESS tags, deliberately leaving `<script type=…>`,
`<script defer>`, `<script src=…>` raw. So a raw script can coexist with the function having
run. Verify the tag shape (and whether a later write could have reverted an
already-extracted row) before inferring "code path X didn't run" from absence. Sibling of
"0 rows is not decisive": the same trap, applied to code behaviour rather than a SQL count.

**Reuse before rebuild; check the schema before SQL.** Look for an existing function/struct
(sibling functions in the same file are the canonical pattern) before writing new code, and
read the table's columns (`information_schema.columns` / `\d`) before composing a query —
tables reported "not found" by `\d` may live in another schema.

**Key schema gotchas (full list in 016):** `site_work_items.error` (not `error_message`);
`site_work_items.pipeline` (not `domain`); `orchestration_states` PK is `orchestration_id`
(no `id` column); back up an agent with `snapshot_agent('<type>','<reason>')` /
`revert_agent('<type>')`, not a hand-rolled `agent_definitions_backup`.

---

## Open threads (carried forward 2026-06-22)

> **Status — TRIAGED (updated 2026-06-24).** Every active page-pipeline thread below is *triaged*:
> root-caused or narrowed to a named remedy, but **not yet applied or deployed**. The one exception is
> Part 4, whose two-layer fix is now *written* (still un-deployed). Two threads found this arc —
> `system-stats` dropped and the `rerender_single_page` invalid `page_id` — are added at the end of this
> list. The false leads taken along the way are logged under "Wrong turns logged this arc" below so they
> aren't re-walked.

The page build/re-render work is a connected series of fixes to the same failure shape —
work that reports success but doesn't happen, or work dropped/duplicated/clobbered.
Operational detail (apply/deploy/verify, with SQL) is in
`RUNBOOK_gamesdesign_index_rebuild.md`; the running investigation log is
`NOTES_gamesdesign_silent_norebuild.md`.

**Part 1 — result-contract drop (DONE).** The chassis coordinator discarded a child's
workflow result (singular `output_field`, or oversize) and substituted a stub that reported
success → no-op save → work item `complete`. Shipped 2026-06-18 (`result_spec.go` +
`coordinator.go`); verified — gamesdesign `index` rebuilt+deployed 06-19 and re-rendered
cleanly 06-22. This also resolved the long-open "index returns thin content" question (it
was the stub/fallback, not thin generation). Full write-up: 016 §9 "Child workflow result
silently replaced by a stub."

**Part 2 — no-LLM re-render path (VERIFIED for image_landed; remainder pending).** Added
`rerender_page_sections` + a gated pre-pass on `page-rerender`, and repointed
`flag_page_image_rebuild` / `reconcile_section_data` to emit a `page_rerender`-type item
(was `needs_page`). Index Test 1 (06-22) passed: all 5 sections re-rendered, copy preserved,
no LLM, P2.1 wiring confirmed. **Remaining:** P2.4 (real image-landed flow), P2.5
(section_data_resolved), P2.6 (NULL content_data escalation → needs_page), P2.7 (plain
page_rerender backward-compat). Also confirm the live `/index.html` now shows all 5 sections
(system-stats). Runbook Part 2.

**Part 3 / Bug 2 — item_key canonicalization (CODE PREPARED; NOT APPLIED).** `item_key`
prefixes drifted from `item_type` across creators. Confirmed live: the adoption creator keys
BOTH `needs_content_page` and `needs_tool_recreation` as `needs_page:<name>`, so a tool and a
content page of the same name collide on `idx_swi_dedup` and one is dropped. Prepared:
`workItemKey(itemType, target)` builder in `work_items_common.go`; adoption tool item →
`workItemKey("needs_tool_recreation", page.Name)`; content item stays
`workItemKey("needs_page", page.Name)` (Option B — preserves the deliberate doc-029 co-dedup
with planner `needs_page` builds; decision recorded in the runbook P3.1). **Apply after Part 2
verifies** (don't stack unverified changes), then run the P3.4 tests. Runbook Part 3.

**Part 4 — interactive page rebuilt as content (CAUSE CONFIRMED; FIX PENDING).** A
`link_resolution_rebuild` (spec: "preserve the copy; re-resolve links") is handled by
`page-build-handler`, which runs the full `page-content-writer` → regenerates from
`plan_sections` (no knowledge of the tool) → the interactive game is discarded, falling back
to `generic-text-block`. Blast radius: one page (`game-pathfinding`). **Step-config result
(2026-06-22):** `page-build-handler` has NO `item_type` branch — `link_resolution_rebuild`
runs the full writer path, and the tool dies because it isn't in the page spec the plan is
built from. **P4.1 result (2026-06-22):** internal links here are build-resolved CTA fields,
not prose — `content_data` has no `<a>`; the only link is the hero CTA in `rendered_html`
(a phantom `/contact.html`). `resolve_internal_links` is a build-time augmenter (writes
`cta_url` into `resolved_data`, consumed by page-content-writer), explicitly "not a
rendered-HTML patcher," and `check_phantom_internal_links` routes page-link fixes to
page-build-handler by design. So (a) **P4.2 (2026-06-22) confirmed** `page_rerender` preserves
all sections (non-destructive) but does NOT re-resolve the schema-sourced hero CTA
(`/contact.html` survived) — so routing `link_resolution_rebuild` → `page_rerender` is **ruled
out** as a link fix; (b) the clobber
hits ANY full rebuild, not just link rebuilds. Plus a separate, possibly bigger thread CONFIRMED
2026-06-22 in deployed HTML: the `guide-economy-basics` hero has TWO phantom CTAs
(`cta_url=/contact.html`, `secondary_cta_url=/services.html`) from the hero schema sources
`pages.contact`/`pages.services`, while the page's footer links the real `/contact/index.html` — so
the hero CTA resolution is producing constructed/fabricated `/{area}.html`, and `resolve_internal_links`
(intent-aware hubs) isn't overriding it. (Retraction: `page_component_history` shows no 06-14 write, so
the claim that a full rebuild produces the phantom is withdrawn — the §5 rebuild likely didn't re-save
this hero.) A phantom-CTA-resolution bug distinct from
the clobber; Layers 1–2 don't fix it. **Fix (layered, structural first):** (1)
interactivity-aware guard in `save_page_sections` (block/escalate dropping a `<canvas>`/
data-component section) — cheap, immediate, all triggers; (2) preserve interactive sections
through a rebuild (carry forward non-spec tool sections; or represent the tool in the spec as
non-regenerated) — the structural root, keeps CTA re-resolution working via the writer, no
routing change; (3) optional later — extend `page_rerender` to re-run the CTA hub logic and
route `link_resolution_rebuild` there (no-LLM). Plus `source_item_id` stamping, and re-create
`game-pathfinding` after. Open: does the rebuild honor `page_components.locked_at` (interim
mitigation if so)? Full write-up: 016 §9 (last entry); runbook Part 4.

**Update (2026-06-24) — status triaged → fix WRITTEN (un-deployed).** Both layers are now in a patched
`save_page_sections_action.go`: Layer 1 (interactivity guard — block a non-interactive set replacing a
deployed interactive one) and Layer 2 (carry-forward of existing interactive sections), plus
`source_item_id` stamping on the history write. **Correction to the layered plan above:** Layer 2 lives in
`save_page_sections`, NOT in `load_existing_content` / `plan_sections` — the tool's bespoke `<canvas>`/JS
exists ONLY as `page_components.rendered_html`, and the planning path (section-*name* skeletons) cannot
reconstruct it (see Wrong turns #1). Three `save_page_sections` callers to bump on deploy: `page-build-handler`,
`page-rerender`, `tool-recreation-handler`. The phantom-CTA hero is a *separate* confirmed bug — `select_sections`
reads `resolved_links.sections_ready` (null) instead of `resolved_links.response.link_resolution.sections_ready`,
so it falls back to the un-augmented plan; workflow-only fix staged. Runbook P4.3.

**Part 5 — `system-stats` section silently dropped (TRIAGED, 2026-06-24).** The live index deploys 4 of 5
sections; `system-stats` is in `page_components` (~7.4 KB) but renders with **zero** visible text, so
`rerender_single_page`'s `getPageSections` visible-content filter (≤10 chars) correctly drops it. Root cause:
the component template (`content_components` `fdd92ad4`) binds `{{.eyebrow_label}}` / `{{.section_headline}}` /
`{{.stat1_value}}` / … but the stored `content_data` has `eyebrow` / `heading` / `stat_1_number` / … —
**zero shared keys**, so every placeholder renders empty. `component-creator` rewrote the component at 15:06
(after the 14:26 build); `usage_count`=22, so the whole fleet loses the band the same way. Remedy (un-applied):
full content rebuild of the index (re-open WI `0b8114c1`, NOT a bare `page_rerender` — that reuses the mis-keyed
`content_data`) → re-check whether the writer now emits the schema's keys. If yes → rebuild the 21 other pages
and make a component schema change trigger dependent rebuilds; if no → fix the content-generator ↔
component-`input_schema` binding. The filter is correct and stays. Runbook Part 5.

**Bug — `rerender_single_page` invalid `page_id` (TRIAGED, 2026-06-24).** `build-dispatch-loop` logged
`step render_page failed: … invalid page_id: invalid UUID length: 31`/`18` — a non-UUID (a name/URL) reaching
`rerender_single_page`, which correctly rejects it via `uuid.Parse`. The per-page looper that maps the bad value
is **`rerender-pages`** — NOT `rerender-site`, which has **zero orchestrations ever** (see Wrong turns #5). Next:
read `rerender-pages`' loop `items_field` + the `page_id` it maps when spawning `page-rerender`, and dump a recent
run. Lower priority than `system-stats`.

---

## Wrong turns logged this arc (2026-06-22 → 2026-06-24)

False leads taken during the page-pipeline triage, kept so the next pass doesn't re-walk them. Each ties
back to a durable heuristic above.

1. **"Preserve the interactive tool in the planner."** Layer 2 of the Part-4 fix was first scoped for
   `load_existing_content` / `plan_sections`. Wrong: the tool's `<canvas>`/JS exists ONLY as
   `page_components.rendered_html`; the planning path traffics in section-*name* skeletons +
   LLM-regeneratable content and cannot reconstruct it. Correction: both layers live in `save_page_sections`,
   the only place that holds the markup. *Lesson: before deciding **where** to fix, confirm where the artefact
   actually lives — the planner can't carry what only exists as rendered output.*

2. **"The missing `system-stats` is just a stale/cached page."** Wrong: a re-fetch 12 min after deploy still
   lacked it, so it's dropped at deploy, not cached. *Lesson: don't explain a missing section away as
   staleness — re-fetch the live artefact (trust the artefact, not the assumption).*

3. **"`system-stats` is the same text-heuristic false-positive as Parts 1/4."** Wrong: `approx_visible_len`
   was literally 0 — the section is genuinely text-empty, so the assembler filter is *correct*, not
   over-triggering. *Lesson: measure before analogising — a prior pattern's shape doesn't apply until the
   numbers say so.*

4. **"`system-stats` is empty because the content wasn't generated (content gap)"** → then *"a render gap
   (template didn't interpolate)."* Both were way-points, not the cause. `content_data` was fully populated;
   the precise cause is a **key-name mismatch** — `content_data` keys (`eyebrow` / `heading` / `stat_1_number`
   …) share zero keys with the template (`eyebrow_label` / `section_headline` / `stat1_value` …). *Lesson: a
   populated-but-unrendered section is a content↔template **key-contract** check, not a generation failure —
   diff the two key sets directly.*

5. **"The bad `page_id` comes from `rerender-site`."** It was the only `page-rerender` caller mapping
   `current_page.page_id`, so it looked guilty — but `GROUP BY owner_agent_type` showed `rerender-site` has
   **zero orchestrations ever**; the real looper is `rerender-pages`. The run-dump's 0 rows weren't pruning —
   the query named an agent that never runs. (Also: the `LIKE '%page-rerender%'` config grep matched
   `rerender-site` but not `rerender-pages`, which spawns `page-rerender` via role/field, not a literal.)
   *Lesson: "0 rows isn't decisive" includes "the entity may not exist or run" — verify it runs (GROUP BY)
   before attributing; and config-text greps miss dynamic spawns.*

6. **"The skip log isn't firing — grep found nothing."** The grep returned nothing because it lacked `-E`
   (the `|` alternation was literal) and `logs*` doesn't match the pod-log path. Not evidence of absence — the
   `visible_len = 0` query had already proven the drop. *Lesson: a no-match grep is as suspect as a 0-row
   query until the pattern, anchor, and target path are each verified.*

---

## 9. Specific Failure Patterns (new)

*(New patterns from 2026-06-22 onward go here, in the 016 §9 style: `### <descriptive
title>`, then Symptom / Diagnose / Root cause / Fix, with the SQL and cross-references. The
immediately-prior pattern — the interactive-page clobber — is the final §9 entry in 016
v2_56; start the next one below.)*

### Data-driven component shells render empty — inline `<script>` never extracted to `js_content`

**Symptom.** A component with built-in interactivity (and, for data-driven ones, a
client-side data loader) renders as an empty/blank shell on the live site, and its
interactivity does nothing. The *page* builds and deploys fine. Seen on vonc.com:
`provocation-card`, `lobby-grid` (and the static `brief-explanation` — see note).

**Diagnose.** Compare the component's stored shape against a known-good one:
```sql
SELECT function,
       LENGTH(COALESCE(js_content,'')) AS js_len,
       (html_template LIKE '%<script src=%') AS has_src_ref,
       (html_template LIKE '%<script>%')     AS has_raw_inline,
       POSITION('</script>' IN html_template) AS has_script_close
FROM content_components
WHERE function = '<fn>' AND is_active = true AND forked_from IS NULL;
```
- **Healthy (current store path):** `js_len>0`, `has_src_ref=t`, `has_raw_inline=f`.
  Confirmed for `gauntlet-interface` (3909), `latest-news` (2174),
  `tool-archetype-taster-quiz` (8313).
- **Broken:** `js_len=0`, `has_src_ref=f`, `has_raw_inline=t` — the raw inline
  `<script>` is still in the template and was never moved to `js_content`.
- `collectJSAssets` (rerender_single_page_action.go) only deploys
  `/tools/assets/{function}.js` when `js_content` is non-empty — so empty
  `js_content` ⇒ no asset ⇒ no interactivity/loader on the page.

**Root cause.** `separateInlineJS` (store_generated_component_action.go) extracts a
bare inline `<script>` into `js_content` and replaces it with a
`<script src="/tools/assets/{function}.js">` reference. The broken rows are NOT
that function's output (raw script present, no src ref, empty `js_content`) — i.e.
whatever last wrote them did not run extraction. Most likely they predate
`separateInlineJS`'s addition to the store path (or a seed/bulk path bypassed it)
and were never regenerated. `provocation-card` is additionally **truncated
mid-`<script>`** in the DB (`has_script_close=0`, stray trailing backslash) — a
generation-time truncation that slipped through because the store-path validation
checks unclosed `<style>` but **not** unclosed `<script>`.

The code itself (`separateInlineJS`, `collectJSAssets`) is correct for
normally-created components — do **not** "fix" it. Two precision points:
- `separateInlineJS` matches only an attribute-less `<script>` (regex
  `<script\s*>`); `<script type=…>`/`<script defer>` are left raw by design. So
  "raw script present" does not by itself prove extraction never ran — confirm the
  tag is bare first (these components' tags are bare, so the inference holds here).
- The only residual ambiguity (never-ran vs ran-then-overwritten) does not change
  the fix.

**Fix.** Regenerate the affected components through the current store path so
`separateInlineJS` runs and the schema/template are rebuilt. `provocation-card`
**must** be regenerated (truncated source — nothing to re-extract); `lobby-grid`
could in principle be re-stored but is also a Mode-B empty shell, so regenerate for
consistency. For data-driven components, regenerate with a **complete** inline
`<script>` containing BOTH the interactivity AND the data fetch-fill loader, so it
all extracts to `js_content` and ships as `/tools/assets/{function}.js` on the next
page rerender (the durable "Path 1" home; retire any interim `js_snippets` loader
afterwards).

**Hardening (prevent recurrence).** (a) Add a `<script>` open/close balance check to
the store-path validation alongside the existing `<style>` check, so a template
truncated mid-script is rejected at store time. (b) Make `separateInlineJS` log a
warning when it sees a `<script>` opener with no matching `</script>` (silent skip
hides truncation).

**Note — same symptom, different fix (`brief-explanation`).** It is also a Mode-B
empty shell, but its content is STATIC and it has no inline script. Its fix is
regeneration with a real input_schema so the content writer fills it at build time
— NOT a JS loader. Same cause class (empty-shell), different content type.

**Cross-refs.** RUNBOOK_phase2_provocation_js.md (EXTRACTION BUG §),
NOTES_provocation-card.md, RUNNING_NOTES_vonc.md (2026-06-29). Category tags:
`js-not-extracted`, `mode-b-template`, `schema-template-drift`.

### Manually invoking an agent via spawn+call — input_mapping must satisfy BOTH the input_contract (top-level) AND the workflow's field paths (usually `input_data.spec.*`)

**Symptom.** A manual `spawn_agent` + `call_agent` trigger for an agent that is
normally driven by work items (e.g. `component-creator`) fails in one of two ways
depending on how you shape the inputs:
- **(a) Contract violation:** `contract violation for agent 'X': missing required
  fields: [section_type]. Provided fields: [domain site_id spec]. Hint: Check
  input_mapping in the step config` — the step fails and the agent never runs.
- **(b) Empty-context generation:** the call succeeds but the agent generates from
  BLANK inputs — e.g. component-creator emits a generic `general-hero` (status
  `created`, a NEW row) instead of the requested `section_type`, and because the
  emitted `function` differs, `store_generated_component` INSERTs a stray instead
  of regenerating the intended component in place.

**Diagnose.** Check where the fields landed relative to two different readers:
- `call_agent` extracts fields via `input_mapping`, then validates them against the
  target agent's `input_contract.required` — this check looks at the **top-level**
  extracted keys. Nesting the required field under `spec` → case (a).
- The agent's **workflow steps** reference `input_data.spec.*` (because the agent is
  normally driven by WORK ITEMS whose `spec_data` becomes `input_data.spec`). Putting
  fields top-level satisfies the contract but leaves `input_data.spec.*` empty → the
  prompt/steps see nothing → case (b).
Confirm by reading the agent's `input_contract` (top-level required fields) and its
step configs / prompt_template (which paths they read, e.g.
`{{.input_data.spec.section_type}}`, `"section_type": "input_data.spec.section_type"`).

**Root cause.** The `input_contract` and the workflow read the same logical field in
DIFFERENT places — contract at top-level, workflow under `spec`. Neither
pure-top-level nor pure-nested input satisfies both. (This is a latent design smell
in the agent: the contract and the workflow should agree. Worth reconciling in the
agent definition — either validate the contract within `spec`, or have the workflow
read top-level.)

**Fix.** Provide the required field(s) BOTH top-level (for the contract) AND inside a
`spec` object (for the workflow), keeping `input_mapping` sources one-level:
```
input_mapping: {
  site_id: "input_data.site_id",
  domain:  "input_data.domain",
  section_type: "input_data.section_type",   // top-level → satisfies contract
  spec:         "input_data.spec"             // full object → satisfies workflow
}
input_data: {
  site_id, domain,
  section_type: "brief-explanation",          // top-level copy
  spec: { section_type: "brief-explanation", description, design_direction, ... }
}
```
Preferred alternative: drive the agent via its DESIGNED work-item trigger
(e.g. `needs_new_component` / `needs_component_regeneration`), where the dispatch loop
delivers `spec_data → input_data.spec` consistently and this mismatch doesn't arise.

**Also (regeneration keying).** `store_generated_component` looks up the existing row
by the LLM's EMITTED `function`, not by any spec field. For an in-place UPDATE the LLM
must emit the same `function`; pass `section_type` correctly (so it mirrors) and pin
the function name in the description. A mismatched name INSERTs a duplicate — verify
`active_rows = 1` and `status = regenerated` afterward, and deactivate any stray.

**Cross-refs.** RUNNING_NOTES_vonc.md (2026-06-30/07-01), NOTES_brief-explanation.md,
triggers 081 (top-level → stray hero), 082 (nested → contract violation), 083 (both →
correct). Category tags: `workflow-variable-path`, `input-shape`, `spawn-call-contract`.

### Light site renders dark chrome — deployed pages are RERENDER output carrying stale `site_components` chrome and old-template section renders

**Symptom.** A site resolved to a light scheme (layout `tool-portal-light`, light palette) deploys with a dark gradient header, dark footer, dark bands. The variable layer is provably correct (`:root` chrome vars flip with scheme; `buildSectionDefaults` emits nothing for light palettes) yet the deployed HTML is dark.

**Mechanism.** There are TWO page-assembly paths with DIFFERENT chrome sources. Build path (`page-build-handler` → `CompilePageSectionsAction` → `InjectHeader/Footer` → `RenderHeader/Footer`) renders chrome FRESH: `style_collections.header_component_id` (a dead column — `install_site_composition` inserts NULL, nothing writes it) → `GetComponentByFunction("site-header")` → dark `RenderFallbackHeader`. Rerender path (`rerender-pages`) reassembles stored `page_components.rendered_html` and injects **stored `site_components.rendered_html`** — which can point at long-deactivated components. Nothing refreshes `site_components` when a component is deactivated.

**Provenance greps that settle it.** In the deployed page: `site-header-section` = current generated header (build path, healthy); `site-header--gradient` = stale `site_components` render (rerender path); bare `site-header` + stacked-nav hamburger + no logo = `RenderFallbackHeader`. Git commit messages `Rerender: <page>` identify the rerender path. **Legacy-variable tell:** a section consuming `var(--accent-color, …)` (old naming) instead of `var(--color-accent)` proves the stored render predates the current template — sections fossilise too, so `needs_rerender` re-fossilises them; only a full `page-build-handler` rebuild re-renders templates.

**Also.** `RenderHead` looks up function `head` (the only such component is inactive → build path always uses `RenderFallbackHead`); `site-head` (component_level=section) is unreachable as chrome. `InjectHeader` SKIPS injection when incoming HTML already contains a `site-header` class — remember when verifying rebuilds. `component_library.go` uses `logger.Debug` in this area (invisible in cluster logs).

**Cross-refs.** `RUNBOOK_scheme_to_components.md` (Checks 1–4 + results), `SPEC_scheme_to_components.md` (the paired-variable fix + workstreams). Category tags: `two-assembly-paths`, `stale-site-components`, `rerender-vs-rebuild`, `provenance-grep`.

### SQL verification pitfalls met during template surgery (Postgres)

- **Regex quantifier bounds cap at 255.** `substring(col from '--x:.{0,420}')` fails with "invalid repetition count(s)". Use `substr(col, position('--x:' in col), N)` — no bound.
- **`substring(string from pattern)` returns the FIRST capture group when the pattern contains parentheses.** `substring(t from 'background(-color)?:…')` returns `''`/`-color`, not the match — the query looks broken when it isn't. Use character classes or drop the parens for display extraction.
- **Per-template classification regexes miss gradient-embedded hexes.** `background(-color)?:\s*#…` flags direct hex paints only; `background: linear-gradient(135deg, #1a1a2e …)` escapes it because the hex is not immediately after the colon. The five `hero-*` variants were mis-classed "paint nothing" this way — in fact they hardcode a legacy-palette dark gradient. Classification queries need a companion `linear-gradient\([^)]*#` test.
- **Template-surgery pattern that held up** (layouts.css_template, content_components.html_template): (1) needle-gate read — every needle as a `LIKE` boolean PLUS occurrence counts `(length(t)-length(replace(t,needle,'')))/length(needle)` so partial coverage is visible BEFORE mutating; (2) shell backup of the full column via `psql -Atc … > file.bak`; (3) guarded, idempotent UPDATE — nested exact-string `replace()` (byte-for-byte needles, no regex) or `regexp_replace` anchored with a `\1` backreference when whitespace is uncertain, plus a `NOT LIKE` pre-state guard; (4) `RETURNING` boolean checks; (5) value-agnostic rollback file + the .bak. Gate rule (refined after a 2026-07-02 false alarm): a count that differs from expectation means EITHER the stored text drifted OR the expectation itself was mis-derived — verify the expectation mechanically against the dump (`grep -o 'needle' dump | wc -l`) before concluding drift. The per-needle booleans are the COVERAGE test; the counts are a completeness cross-check whose expected values must be counted, not recalled (the hero gate misfired on 7-vs-6 and 2-vs-3 because two expectations were remembered wrongly — the twelve needles were complete and every post-condition passed). A third LIKE pitfall joins the family: needles containing literal `%` (gradient stops `0%/50%/100%`) cannot be gate-checked with `LIKE` — use `position(needle in col) > 0`.

**Cross-refs.** w1_*/w2_*/w3_* files in the scheme-to-components outputs. Category tags: `sql-pitfall`, `template-surgery`, `needle-gate`.

### `sites.status`: vocabulary, the legacy 'active' value, and a wrong blast-radius filter

`UpdateSiteStatusAction` (v3_site_actions.go:323) is the writer; validated vocabulary **draft / building / review / published / deployed / archived / error** (with `deployed_at:"now"` it also stamps `last_deployed_at`). **'active' is NOT in the vocabulary** — a row carrying it is legacy/hand-written. No on-disk code FILTERS sites on status; it is an informational lifecycle label, and build dispatch keys on `site_work_items` — a `deployed` site is still rebuildable. Heuristic: never scope a blast-radius or "live sites" query with `sites.status='active'` (an assumption borrowed from an old handoff query did exactly this and silently dropped the site under investigation from the count). Enumerate `GROUP BY status` first.

Related reuse-gate lesson: before adding an `updated_at` trigger, check `pg_proc`/`pg_trigger` — a shared `set_updated_at()` already exists (site_specs, site_plans, content_feed_items, training_runs use it). `CREATE FUNCTION` (not `OR REPLACE`) as the loud collision gate; add `\set ON_ERROR_STOP on` to multi-statement files only when later statements DEPEND on earlier ones.

**Cross-refs.** RUNBOOK §"sites.status RESOLVED", §W2b RESULTS. Category tags: `status-vocabulary`, `blast-radius`, `reuse-gate`.

### `error_step`: config-level placement, existing-target requirement, and the derive-from-next_step pattern

**Symptom (two stages).** (1) Error routing added to a workflow step does nothing —
the step's failure still fails the whole orchestration. (2) After moving it to the
right place, the routing FIRES (`error_routed` in ProcessingHistory) and the
workflow then fails anyway: `step 'assemble' not found`.

**Mechanism.** The coordinator reads `step.Config["error_step"]` — step-LEVEL
`error_step` is parsed into the Step struct but never consulted
(`routeToErrorStepOrFail` → 001 §16), so placement outside `config` is silently
inert. Once placed correctly, `routeToErrorStep` → `continueExecution` looks the
target up in the live step map; an unknown name fails the whole workflow — i.e. a
mistyped `error_step` CONVERTS a recoverable step failure into a fatal one. The
wrong name here came from inferring the step name from its ACTION name
(`diagnose_assemble_bundle` ≠ step `assemble_bundle` ≠ guessed `assemble`).

**Fix pattern (derive, don't name).** When converging failure onto the success
path, set `error_step` to the step's own `next_step` READ FROM THE SAME ROW:
`jsonb_set(default_config, '{workflow,steps,<step>,config,error_step}',
to_jsonb(default_config #>> '{workflow,steps,<step>,next_step}'), true)` — with a
guard that `error_step = next_step` AND the target exists in the step map.
Convergence by construction; nothing is guessed. Two companions: (a) `jsonb_set`
does NOT create missing parents — when a step might lack a `config` object,
COALESCE-merge it (`COALESCE(cfg,'{}') || '{"error_step":…}'`); (b) verify targets
against the LIVE step map (a run's state dump prints `state.WorkflowPlan` in
full — the authoritative source when the definition paste has scrolled away).

**Loop corollary (001 Appendix C / loop_expansion_handler.go).** Inside a loop's
substeps, `error_step`/`then_step`/`fallback_step` values are ITERATION-PREFIXED
at expansion — they must name SUBSTEPS of the same loop, never a top-level step.
`continue_on_error: true` on the loop is the iteration-scoped alternative (record
the error, advance to the next iteration — loop_error_handler.go).

**Dormant instances observed.** `tool-recreation-handler` and `tool-auditor`
definitions carry step-LEVEL `error_step` on several steps, and
`tool-generator` on ALL THREE of its routed steps (`save_tool`,
`generate_tool_html`, `load_brand_context` — observed 2026-07-07) — the
documented silently-ignored form. Never copy that shape; correct adjacent instances when
touching those workflows, as its own noted change.

**Cross-refs.** 001 §16 + Appendix C; RUNBOOK_travelling_docs §3a (runs 1–3);
drafts 0NN_diagnose_load_runtime_error_step.sql /
0NN_fix_load_runtime_error_step_target.sql. Category tags:
`error-step-placement`, `routing-target`, `jsonb-parents`, `loop-substep-prefix`.

### Anchorless (code-only) diagnosis dies at `load_runtime` — an optional evidence tier made mandatory in practice

**Symptom.** A subjectless/anchorless diagnosis run fails the WHOLE child
workflow: `diagnose_load_runtime: need at least one of site_id / correlation_id /
domain in collected_data` — never reaching verdict/emit/persist_note.

**Mechanism.** The bundle treats runtime evidence as OPTIONAL
(`diagnose_assemble_bundle` omits the section when the field is empty), but the
gather step hard-errors with no anchor and had no error routing — so the tier's
absence was fatal and a legitimate code-only diagnosis mode could not run.

**Fix.** Config-level `error_step` on `load_runtime` pointing at its own
`next_step` (`assemble_bundle`, derived — see the entry above). Live-validated
×5 in one run: `route.config.gather_step = "load_runtime"`, so EVERY loop-back
iteration re-enters the gather step; an anchorless run now degrades per-iteration
to a code+schema bundle and proceeds. **Reading rule:** five
`error_routed … routed to assemble_bundle` lines in one anchorless run are
NORMAL, not a fault. **Cost note:** such a run still executes the full loop
(≤5 LLM verdicts) before emit — minutes, not seconds. Pending softening (next
chassis build): the action returns `{runtime_evidence:"", skipped:true, reason}`
for the no-anchor case, keeping the hard error for genuine DB failures.

**Verification discipline (0-rows rule, gate edition).** When a downstream step's
non-firing is the SUCCESS condition (e.g. persist_note's skip gate), a 0-row
count is decisive ONLY alongside (a) a COMPLETED child and (b) the step's own
explicit skip log line — a run that died upstream also produces 0 rows. The skip
LOG LINE has a capture window (the idle reaper, 3600s); past it, a
post-completion STATE DUMP is the accepted substitute — ProcessingHistory
showing the step EXECUTED then the terminal step, plus the terminal status,
plus the 0-count (run-3 was closed this way; the subject gate preceding any DB
access settles skip-vs-error structurally).

**Cross-refs.** RUNBOOK_travelling_docs §3a; live step map: analyse_repo →
lookup_symbols → load_runtime → assemble_bundle → verdict → route → emit →
persist_note → complete. Category tags: `diagnosis-loop`,
`optional-tier-made-mandatory`, `anchorless-run`, `zero-rows-gate`.

### Pod label key is `agent-type` (hyphen) — and multi-pod log selectors mix runs

**Symptom.** `kubectl logs -l agent_type=<agent>` returns nothing; with the
correct key, a PREVIOUS run's failure dump appears inside the current run's
capture window and nearly gets read as current.

**Mechanism.** The pod LABEL key is `agent-type` (hyphen); the log JSON FIELDS
say `agent_type` (underscore) — easy to conflate, and the underscore selector
silently matches zero pods. Separately, `-l agent-type=<agent>` spans ALL live
pods of that type (spawned diagnosers persist until the idle reaper, 3600s), so
tails contain residue from earlier runs.

**Rules.** Hyphen in `-l` selectors. Attribute every log line by
orchestration_id / pod_name / timestamp before treating it as current — "the
plan table is ground truth; the rest is weather", log edition. Older trigger
scripts (082/083c) carry the underscore form in their monitor echoes — fix when
touched. Category tags: `label-key`, `log-attribution`, `multi-pod-residue`.

### Two failure envelopes — a COMPLETED parent does not mean the child succeeded

**Symptom set.** (1) A child workflow fails mid-run; the ORCHESTRATOR's
`orchestration_states` row shows COMPLETED — with the child's error text sitting
in the parent row's `error` column. (2) A child fails at workflow START; the
parent instead receives an explicit error envelope.

**Mechanism.** Two reporting paths: a step-level workflow failure is returned by
`sendWorkflowResponse` with HEADER `status: "complete"` and the failure in the
BODY (`body: {status: "failed", error: …}`) — the parent forwards it and
completes; a failed-to-start workflow goes via `notifyParentOfFailure` with
header `status: "error_unrecoverable"`, `is_error: true`, body
`{success:false, error:{code: "CHILD_ORCHESTRATION_FAILED", …}}`.

**Rules.** Consumers of a child result must check the BODY (`body.status` /
`success`), never the header status alone. A COMPLETED parent row with a
non-empty `error` column = a forwarded child failure, not success. Expect BOTH
shapes when monitoring; which one appears tells you WHERE the child died
(mid-workflow vs at start). Category tags: `failure-envelope`,
`parent-completed-child-failed`, `body-status`.

### `agent_error_log` is the FIRST read — not pod logs, not `current_step`

**Symptom.** A run ends at a terminal error step. The pod is gone (idle reaper),
`orchestration_states.error` is empty, `collected_data` is too large to read,
and the last polled `current_step` names a step that actually SUCCEEDED.

**Mechanism.** Step failures are recorded in `agent_error_log` with
`orchestration_id`, `step_name`, `action`, `error_message`, `error_code`,
`severity`, `pod_name`, `context` — it survives pod cleanup and is queryable by
site or orchestration. Meanwhile `current_step` is a *sampled* value: a poll
every 60-120s can miss a step entirely, so the last observed step is evidence of
where the run WAS, never of where it FAILED. A terminal step like
`complete_error` also carries a generic `success_message` (e.g. "Tool generation
failed") that may name the wrong phase.

**Rule.** Read `agent_error_log` FIRST, filtered by `orchestration_id` (not by
site + time, which mixes runs):
```sql
SELECT occurred_at, agent_type, step_name, action, error_code, severity,
       left(error_message, 400) AS err, context
FROM agent_error_log
WHERE orchestration_id = '<uuid>'::uuid
ORDER BY occurred_at;
```
Only then reach for pod logs (which may not exist) or `collected_data` (which
may be enormous — the terminal step often emits the full LLM output). Corollary:
a downstream step producing zero rows can be the CORRECT outcome of an upstream
failure — check which step actually failed before calling anything a regression.

Category tags: `agent-error-log`, `step-attribution`, `poll-sampling`,
`terminal-step-message`.

### Code ahead of DB — schema drift (SQLSTATE 42703), latent until the first caller

**Symptom.** An action fails with `column "X" of relation "Y" does not exist
(SQLSTATE 42703)`. The code is correct; the column is genuinely absent.

**Mechanism.** A binary that references new columns was deployed before its
migration was applied. Nothing fails until something CALLS that code path — so
the drift can sit dormant for months if the caller is rare (here:
`create_tool_component`, no tool created since 2026-05-16, while
`component-creator` kept working because it inserts a different column set).

**Detection.** The failing INSERT's own comment often names the missing
migration. Confirm with a latency probe before assuming a fresh regression:
```sql
SELECT function, created_from, created_at FROM content_components
WHERE created_from IN ('generated','tool') ORDER BY created_at DESC LIMIT 10;
```
A long gap since the last successful call = latent drift, not a new break.

**Fix pattern.** Apply the missing migration; if the named file was never
written, write it and commit it so the code's reference resolves. MIRROR the
types from the table the code says it mirrors, dynamically, rather than guessing:
`format_type(a.atttypid, a.atttypmod)` from `pg_attribute` +
`EXECUTE format('ALTER TABLE t ADD COLUMN IF NOT EXISTS c %s', t_type)`.
Additive, nullable, idempotent columns are safe for old and new binaries alike.
No `snapshot_agent()` — that is for `agent_definitions`, not data tables.

**Standing check** before deploying a binary that touches new columns: grep the
diff for new column names, and confirm each exists in production.

Category tags: `schema-drift`, `42703`, `migration-not-applied`, `latent-bug`,
`mirror-types`.

### Prompt templates vs action config paths — `TEMPLATE_FIELD_ERROR`

**Symptom.** A step using `execute_llm_prompt` fails before the API call:
`failed to render prompt template: ... at <.X.result>: can't evaluate field
result in type interface {}` (`error_code = TEMPLATE_FIELD_ERROR`). The same
workflow's OTHER steps reference `X.result` in their config and work.

**Mechanism — two different resolvers, and the template one depends on
`output_format`:**
- `output_format: "text"` → the prompt template receives the **bare string**.
  `{{.generated_html}}` is correct; `{{.generated_html.result}}` errors, because
  a string has no fields. (Map values like `{{.site_record.domain}}` traverse
  normally in the same template, so map access is not the issue.)
- `output_format: "json"` → the template receives a **map**; the live, working
  form is `{{.tool_analysis.result | toJSON}}` (`toJSON` is a registered
  template func — reuse it rather than reaching into keys you have not verified).
- **Action CONFIG field paths** (`html_content: "generated_html.result"`,
  `html_field: "improved_html.result"`) use a different resolver and keep the
  `.result` suffix. Proven in the same run: `save_tool` hard-fails on empty
  `html_content` and did not.

**Rules.** Never reach a nested key inside an LLM output from a prompt template
unless the step is `output_format: json` AND the key is in that step's own
output schema; prefer `{{.X.result | toJSON}}` (json) or `{{.X}}` (text). Dump
whole objects with `| toJSON` rather than guessing keys. This is a
render-time error: it fires before any tokens are spent, and if the step has
`config.error_step`, it routes there silently — the workflow "succeeds" while
the step's product is missing.

**Reading rule.** A run that ends at its normal terminal step with a DOWNSTREAM
artefact missing = a contained step failure. Look for the containment target in
`config.error_step`, then read `agent_error_log` for the real step.

Category tags: `template-field-error`, `output-format`, `tojson`,
`config-vs-template-resolver`, `contained-failure`.

### Two schema traps met while debugging the above

- **`agent_error_log.orchestration_id` is `text`, not `uuid`.** Casting the
  literal (`= '...'::uuid`) fails with `operator does not exist: text = uuid`.
  Compare as text: `WHERE orchestration_id = '<uuid-string>'`.
- **Provenance stamps the CHASSIS agent type, not the logical agent.**
  `content_components.source_agent_type` came back `generic` for a
  `tool-generator` run (pods are `agent-chassis-*`; `agent_error_log.agent_type`
  says `generic` too). So `ExecutionContext.Sender.AgentType` is no better than
  `Headers["agent_type"]` for "which agent did this" — prefer the
  **config-declared** source (e.g. `plan_source`, `note_source`), which the doc
  actions already carry.

Category tags: `schema-trap`, `text-vs-uuid`, `provenance`, `generic-chassis`.

### Error containment does not protect against a HANG

**Symptom.** A step sits in `EXECUTING_STEP` for far longer than the workflow's
`timeout_seconds`; `agent_error_log` is silent (nothing failed); the pod is
alive; `config.error_step` never fires.

**Mechanism.** `config.error_step` routes *errors*. A network call that connects
but never answers produces no error, so nothing routes. Observed:
`timeout_seconds: 480` did not terminate a step still executing at 2641s —
i.e. the workflow timeout does not govern in-process action execution
(hypothesis: it governs awaited responses / child orchestrations). Concretely,
`rag_index` chunks content and calls `GenerateEmbedding` per chunk against the
Ollama adapter; embedding *failures* are handled ("store without embeddings"),
but a stall has no deadline to trip.

**Rules.**
- Any action making a network call needs its own `context.WithTimeout` (or an
  `http.Client{Timeout}`), so a stall degrades into the error path the action
  already handles.
- "Non-fatal on error" is not "non-fatal on hang" — audit every optional/derived
  step (indexing, notifications, telemetry) for deadlines, not just error
  routing.
- Triage: a step whose `since_s` exceeds `timeout_seconds` with an empty
  `agent_error_log` is a stall, not a failure. Look for an outbound call in that
  action.
- Stopgap for a derived step: re-point the previous step's `next_step` past it
  (keep the step defined and annotate why), then fix the deadline in code.

Category tags: `hang-vs-error`, `missing-deadline`, `timeout-scope`,
`derived-step-bypass`.
