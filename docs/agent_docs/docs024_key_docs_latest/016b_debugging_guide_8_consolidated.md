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

*v5 (2026-07-04): MERGE — this copy folds in three entries that existed only in the vonc-thread working copy (plan_sections deferral drop; deploy verification via artifact not pod logs; marker REPLACE anchoring), plus the new §9 "Page build completes having built nothing — zero planned sections treated as success" (provocations-index; silent-noop-success + planning-gap). Guide had forked across chats; this is the cumulative version.*

*consolidated (2026-07-06): promoted the cumulative v5 copy (was `016b_debugging_guide(8).md`) to this canonical filename. Verified against ALL forked copies of 016b — the `(1)`–`(8)` download-duplicates, the `_5_`/`_6_`/`_7` version series, and `docs/016b_debugging_guide_6_(1).md` in the parent dir (17 files): a full heading-level AND content-line diff proved this copy already contains every one of the 9 distinct §9 entries and the newest top-matter. The only cross-copy divergences were superseded earlier drafts of the Part 4 / phantom-CTA investigation, already present here in corrected form (see the 2026-06-24 Update + Retraction under Open threads). No bug info was dropped. This is the base for future 016b edits.*

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

### Regenerated content section is deferred by plan_sections (required field, unresolved data source) and dropped from the page

**Symptom.** A content component is regenerated to quality 100 with a real schema, but after a page rebuild it never renders and its existing page_component instance disappears. The build result shows fewer `sections_saved` than the plan has sections. render_mode=`agent` sections vanish while `template` ones survive — this is correlation, NOT the cause (deriveRenderMode returns `agent` merely when any field source=`llm`).

**Diagnose.** `plan_sections` (`planSection`) classifies each planned section by resolving its schema fields. source=`llm` is always available; `query.*`, `renderer`, `static` resolve at render time or fall back; any OTHER source is a data path run through `resolver.resolve()`. For a REQUIRED field whose source returns nothing, the `on_missing` switch decides — `use_fallback`+fallback fills it, `skip_section` skips, `needs_human_review`/`block` defer, and the **`default:` case DEFERS** ("Unknown on_missing — default to defer for safety"). An empty `on_missing` defaults to `skip_field`, which is NOT a case in the required switch, so it hits `default:` → deferred. A deferred section is excluded from `sections_ready`; the content writer never renders it; `save_page_sections` then persists only the rendered set — dropping the deferred section's page_component instance (carry-forward fix pending).
Pinpoint the blocking field(s):
```sql
SELECT cc.function, f.key AS field, f.value->>'source' AS source,
       f.value->>'required' AS required, f.value->>'on_missing' AS on_missing,
       (f.value ? 'fallback') AS has_fallback
FROM content_components cc, LATERAL jsonb_each(cc.input_schema->'fields') f
WHERE cc.function = '<fn>' AND cc.is_active AND cc.forked_from IS NULL
  AND COALESCE(f.value->>'source','') <> 'llm'
  AND COALESCE(f.value->>'source','') NOT LIKE 'query.%'
ORDER BY field;
```
Culprit: `required=t`, `has_fallback=f`, `on_missing` empty/`skip_field`, data source that doesn't resolve for the site. (vonc index: brief-explanation.illustration_url ← site_assets.illustration; gauntlet-cta.cta_primary_url + system-stats.cta_url ← site_specs.cta.primary_url — the latter shared.)

**Root cause.** The schema requires a field whose site data isn't present, and its `on_missing` defers rather than degrades. `skip_field` on a REQUIRED field is contradictory — the default case defers.

**Fix (structural, in priority order).**
1. If the data SHOULD exist, POPULATE the site source so it resolves — set `site_specs.cta.primary_url`, register `site_assets.illustration`. A shared source is one fix for all dependants.
2. If the field is genuinely optional/decorative, fix the authoring: `required=false`+`on_missing=skip_field` (omit when absent) or `on_missing=use_fallback`+fallback. Never leave `required=true` with `on_missing=skip_field`.

**Also.** `save_page_sections` dropping deferred sections' instances is the interactive-page-clobber cousin; carry-forward belongs in `save_page_sections_action.go`.
Cross-refs: RUNNING_NOTES_vonc.md, NOTES_brief-explanation.md. Categories: `schema-data-mismatch`, `content-vs-runtime-mismatch`, `detool-on-rebuild`.


### Verifying a deploy: pod logs are ephemeral across rollouts — check the artifact, and grep zap by field

**Symptom.** After a change deploys and a rerender/build runs, you grep the pods for the
expected log line and find NOTHING, and start to doubt the fix ran.

**Diagnose / root causes (seen 2026-07-03, section-drop fix).**
1. **Wrong grep pattern.** Our logs are zap JSON: `{"level":"info","msg":"getPageSections:
   filtered empty sections","skipped":1,"kept":5}`. Grepping a `field=value` literal like
   `skipped=1 kept=5` never matches — `skipped`/`kept` are separate JSON keys. Grep the
   MESSAGE string (`filtered empty sections`) and read the JSON fields; never `field=value`.
2. **Saved log files don't span the window.** `grep ... logs*` against captured files misses a
   run that happened after those captures. Use `kubectl -n ai-persona-system logs -l
   app=agent-chassis --since=NNm` for a live window.
3. **The pod that ran it was rolled.** `kubectl logs` follow pod lifecycle — when a rollout
   replaces a pod (its ReplicaSet hash changes), the old pod's logs are gone. During active
   deploying, the pod that served your run may no longer exist, so its lines are unretrievable
   even though they were emitted. (A different log sink for action-level logs is also possible;
   check the logging config if in doubt.)
4. **logger.Debug.** If the line is a Debug call it won't surface at all (house rule). Confirm
   the level in code before assuming absence == didn't-run.

**Do not conclude "the fix didn't run" from missing logs.** Confirm against a GROUND-TRUTH
artifact instead: the deployed file (`curl -s https://<site>/<page> | grep -o '<marker>'`) or
the DB. In the section-drop case, the fix was proven by the asymmetric OUTCOME (one interactive
section kept, the other dropped) — behaviour only the patched code produces — not by any log.

**Lesson.** During active rollouts verify deploys via the artifact (curl/DB), treat pod logs as
best-effort, grep zap by message+field, and check the log level in code before trusting silence.


### The data-runtime-fill marker (or any attribute added by string REPLACE) must be anchored on the tag, not the bare attribute

**Symptom.** After adding `data-runtime-fill="true"` to a section via
`REPLACE(html, 'data-component="X"', 'data-component="X" data-runtime-fill="true"')`, the
section's own inline script stops working (console `SyntaxError`), though the section still
renders and (if it has an external loader) still fills.

**Root cause.** The bare string `data-component="X"` appears TWICE in a runtime-filled
section's HTML: once in the `<section ...>` opening tag (intended target) and once inside the
section's inline script, `document.querySelector('[data-component="X"]')`. A plain `REPLACE`
hits BOTH, turning the selector into `[data-component="X" data-runtime-fill="true"]` — two
attributes in one bracket, which is invalid CSS → `querySelector` throws → the IIFE dies at
its first line. (Loaders in `snippets.js` are unaffected — the REPLACE only touches
`page_components.rendered_html` / `content_components.html_template`.)

**Fix (targeted revert).** Revert only the in-selector copy (the one immediately followed by
`]`), keeping the tag copy (followed by a space + more attributes):
`REPLACE(html, 'data-component="X" data-runtime-fill="true"]', 'data-component="X"]')`.
Then rerender to redeploy. Verify: `still_broken=f`, `section_marker_kept=t`.

**Prevention.**
- Anchor the marker REPLACE on the OPENING TAG, not the bare attribute, e.g.
  `REPLACE(html, '<section class="X-section" data-component="X"', '<section class="X-section" data-component="X" data-runtime-fill="true"')`.
- Better: emit the marker at GENERATION (component-creator's runtime-feed tier writes
  `<section ... data-runtime-fill="true">` for dynamic sections) so no post-hoc string surgery
  is needed. The REPLACE approach is an interim hack — always anchor it.
Category: content-vs-runtime-mismatch, string-surgery.

### Page build "completes" having built nothing — zero planned sections treated as success (provocations-index)

**Symptom.** A page that the whole site links to (header nav, hero CTA, `site_specs.cta.primary_url`, arena card urls) returns 404 NoSuchKey from B2. The `pages` row shows `build_status='planned'` with `updated_at` still at its creation instant — yet `site_work_items` holds SEVEN items for the page (the original `needs_page` 14s after page creation, two manual `needs_page` rebuilds, four `page_rerender`s over two weeks), **all `complete`, none with an error**.

**Diagnose.**
- Plan check: `site_plan_sections` for the page's `page_name` in the current plan → **0 rows**. Verify the query first (the identical join returned another page's sections), then rule out a spelling mismatch: `SELECT DISTINCT sps.page_name FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id WHERE sp.site_id=... AND sp.is_current ORDER BY 1;` Also check the fallback: `SELECT sections FROM pages WHERE id=...` (default `'[]'`).
- Work-item history for the page (`page_id` or item_key match): all complete, no errors, and the page row untouched throughout (`updated_at` unchanged) ⇒ the handler exited before any step that writes the page, every time.

**Route of failure (the chain).**
1. **Planner emitted the page without sections.** The `pages` row, nav entry, and title exist; `site_plan_sections` has nothing for it; `pages.sections` is `'[]'`. Likely systemic for this page class: `section-index`/archive pages list dynamic items and the planner has no component vocabulary for them.
2. **page-build-handler treats zero planned sections as success.** `load_spec_sections` → nothing (plan and fallback both empty) → `plan_sections` triages nothing → `check_has_ready_sections` routes the zero case to a graceful completion: item `complete`, no error, `build_status` untouched. *(Confirm the exact branch in the workflow def before editing — the untouched page row across seven completes is the behavioural proof.)*
3. **page-rerender does the same.** No `page_components` for the page → `check_skipped` → complete without deploying a file.
4. Nothing ever writes the file → 404 at the destination of every primary CTA. **No auditor rule catches "active + linked-to + build_status planned + zero sections".** A success status masked a no-op for two weeks.

**Root cause (two layers).** (a) A planning gap — an internally inconsistent plan (page exists, no sections). (b) Build and rerender paths that treat "nothing to do" as success, so the planning defect is invisible: no failed item, no error, nothing for a human or auditor to see.

**Fix / prevention (structural, ordered).**
1. **Planner invariant at plan-store time:** every emitted page must have ≥1 plan section (or carry an explicit externally-generated marker). Reject or flag the plan otherwise — internal consistency by construction.
2. **Handler guard:** distinguish zero-**planned** (planning defect → fail the item with a stated reason, or raise `needs_plan_sections` routed to the planner/auditor) from zero-**ready** (deferrals → report which fields and why). Never complete silently on zero-planned.
3. **Rerender guard:** skipping deploy because an ACTIVE page has no components should warn/raise, not quietly complete.
4. **Auditor rules:** (a) page `status='active'` + linked (nav / `link_registry` / CTA specs) + `build_status IN ('planned','needs_rebuild')` beyond a threshold → work item; (b) post-deploy URL presence check (HEAD, or git-tree membership) for every active page → any 404 raises an item regardless of cause.
5. **Section-index vocabulary:** give the planner a component vocabulary for archive/list pages — a list component with `kind=dynamic` + `data_feed` in the section descriptor — so this page class gets planned sections at all. (Ties to PLAN_dynamic_sections_and_loaders.md and the complex-tool build loop.)

**Cross-refs.** RUNBOOK_phase2_provocation_js App I; NOTES_provocations-index.md; PLAN_dynamic_sections_and_loaders.md. Category tags: `silent-noop-success` (new), `planning-gap` (new).

### Every keyed work-item insert fails 42P10 after a dedup-index migration — ON CONFLICT arbiter inference broke (2026-07-17)

**Symptom.** Discovery runs (completeness-discovery-agent) FAIL at `run_checks`. Pod log shows a burst of `Failed to insert work item ... SQLSTATE 25P02 (current transaction is aborted)` — but 25P02 is the *cascade*, not the cause. The first failure in the log is the real one: `ERROR: there is no unique or exclusion constraint matching the ON CONFLICT specification (SQLSTATE 42P10)`. Because `RunDiscoveryChecksAction` runs ALL checks' inserts in one transaction, the first 42P10 poisons everything after it and the final commit fails — the whole discovery output for the site is lost, fleet-wide, on every run.

**Diagnose.**
- Find the FIRST error in the pod log, not the loudest. 25P02 lines name innocent items; the 42P10 line names the item whose insert was merely first after the break.
- 42P10 is a **plan-time** error: the `ON CONFLICT (cols) WHERE <clause>` could not be matched to any index. It fails on EVERY insert through that path regardless of data.
- Compare the live index predicate with the Go clause: `SELECT indexdef FROM pg_indexes WHERE indexname='idx_swi_dedup';` vs `workItemTerminalStatuses` (work_items_common.go), which is interpolated into `insertWorkItem`'s `ON CONFLICT ... WHERE status NOT IN (...)`.

**Root cause.** Migration `157_swi_dedup_excludes_cancelled.sql` (correctly) added `'cancelled'` to `idx_swi_dedup`'s excluded-status set, but the Go list stayed at six statuses. For partial-index arbiter inference the ON CONFLICT WHERE clause must IMPLY the index predicate; `status NOT IN (6)` does not imply `status NOT IN (7)` (a `cancelled` row satisfies the former, not the latter) → no arbiter → 42P10. The index and the interpolated list are ONE contract split across SQL and Go — either side changing alone breaks every keyed insert.

**Fix.** Add `cancelled` to `workItemTerminalStatuses` (commit `5e2711997`, shipped v1.0.1127) — comment on the constant now names the lockstep rule. Follow-up sweep (commit `21e74808e`): `create_tool_cross_link_items.go` had its own HARDCODED five-status ON CONFLICT clause (already missing `unresolved` — it had been silently 42P10-failing and Warn-logging for some time), plus three EXISTS/UPDATE *filters* using the stale list (store_generated_component, component_selector, core-manager site_admin_handlers) — those don't error but treated `cancelled`/`unresolved` rows as open blockers. All now derive from the shared constant (core-manager: literal kept in sync by comment).

**Transferable rules.**
1. When a migration touches a partial unique index used as an ON CONFLICT arbiter, grep for every `ON CONFLICT` naming those columns AND every interpolation of the status list — index predicate and clause must move in lockstep, in the same deploy window (DB change is instant; Go rides the next image — the gap is live breakage).
2. In a shared-transaction insert loop, the first error is the only real one; everything after is 25P02 noise. Also note the loop Warn-logs and continues, then dies at commit — so the work item table shows NOTHING, which reads as "site is clean". Zero findings from a FAILED run is not zero findings.
3. Never accept "zero new work items" as proof of a clean discovery pass without checking the orchestration status (`orchestration_states.status='COMPLETED'` for the discovery agent, not just the generic wrapper — the wrapper completes even when the child workflow fails).

Category tags: `split-contract-drift` (new), `first-error-vs-cascade`, `silent-noop-success`.

### LLM truncation persisted as a successful artifact — a fix agent destroyed the component it was repairing (2026-07-18)

**The family.** This is one mechanism with several victims: `bugs_open/005`
(article body blanked), `bugs_open/008` (`GenerateText` never decodes
`stop_reason` — 17 proven occurrences across 5 agent types, **diagnosed, fix NOT
yet shipped**), and `bugs_open/012` (below — the first case proven to DESTROY a
durable artifact). Treat any "the model returned something odd" report as a
member of this family until the token counts say otherwise.

**Symptom.** An agent that rewrites a whole artifact reports SUCCESS and writes a
confident note describing a sensible change — but the stored artifact is a
fragment. Seen 2026-07-18: `tool-improver` was dispatched to fix a mobile
overflow on `tool-loot-table-balancer`; it root-caused correctly, then reduced a
working **10,272-char** component to **1,253 chars of CSS only** — no `<script>`,
no `<div>`, no `<fieldset>`, ending mid-declaration (`font-weight: bold,`). The
work item read `complete`.

**Diagnose.**
1. **The signature is `output_tokens == max_tokens`.** Both are already recorded:
```sql
SELECT step_name, model, output_tokens, max_tokens, success
FROM llm_call_log WHERE work_item_id = '<item>' ORDER BY created_at;
-- out_tokens=8000 max=8000  ⇒ the completion was CUT, not finished.
```
2. **Check the artifact for structural collapse**, not just length:
```sql
SELECT length(html_template) AS len,
       (html_template LIKE '%<script%')   AS has_script,
       (html_template LIKE '%</script>%') AS script_closed,
       right(html_template, 60)           AS tail
FROM content_components WHERE function = '<fn>' AND is_active;
```
A truncated write ends mid-token; `script_closed=false` with `has_script=true` is
conclusive.
3. **Recovery source: `component_versions`** (one row per `update_component_html`
write). Find the last COMPLETE version — do not assume the newest is good; on
this incident the newest two were *both* truncated:
```sql
SELECT created_at, length(html_template) AS len,
       (html_template LIKE '%</script>%') AS complete, right(html_template,45) AS tail
FROM component_versions WHERE component_id = '<id>' ORDER BY created_at;
```

**Root cause (three layers, all required).**
1. **No truncation detection.** `stop_reason` is never decoded (`bugs_open/008`),
   so a cut-off completion is returned as an ordinary success string.
2. **No completeness guard on the write.** The save path accepted a CSS-only
   fragment as a valid `html_template` — no length-collapse check, no "did the
   markup/script survive" check.
3. **Whole-component writers were sized inconsistently.** Agents doing the SAME
   job had different ceilings: `recreate_tool` **64000** (correct), while
   `improve_tool` and `generate_tool_html` sat on **8000**. The tool's own BIRTH
   had used 6094/8000 — the generator was one slightly larger tool away from
   shipping a truncated component to a live site.

**Why the site did not go down — and why that is not reassurance.** The page had
been rendered from the previous complete component and the render had not
re-propagated. The durable source was wrecked while the artifact still served.
This is the `rendered artifact vs durable source` split (see the vonc footer
arc): here it hid the damage rather than causing it. One refresh would have
shipped a page with no markup and no JavaScript.

**Fix — all three layers now exist (2026-07-18); (a) and (b) await an image roll.**
(a) **A write-time completeness guard**, `component_write_guard.go`
(`componentRegressionIssues`), wired into `update_component_html`: on a
structurally-worse replacement it hard-errors — row untouched, step failed,
`error_code='component_write_regression_blocked'` in `agent_error_log`. Three
COMPARATIVE checks — detailed below the correction.

> **CORRECTED 2026-07-18** (by the thread that wrote the guard; the entry above
> was written by a second thread that found the code uncommitted in the shared
> tree and reasonably inferred the rest). Two claims were wrong:
> **(i)** the guard is wired into `update_component_html` ONLY, not into
> `store_generated_component` — the birth path keeps its own, separate,
> schema-shaped gate, and a proposed consolidation of the two was dropped on
> council objection (see the NOTE in `store_generated_component_action.go`).
> **(ii)** there was NO `error_step` routing to `needs_human_review`:
> `tool-improver`'s `update_component` step had `error_step = null`, so a refusal
> would have reached `failWorkflow` — orchestration FAILED, item left to the
> reaper, no note. Migration **169** supplies that route
> (`refuse_mangled_write` → `note_refusal` → `complete`) and is **not yet
> applied**. Until it is, the guard prevents the destruction but the outcome is
> still a bare orchestration failure.
> *What caught it:* writing the migration required dumping the live
> `tool-improver` workflow step graph, which showed the null `error_step`.

The three checks: size collapse <50% retained; unterminated
`<script>/<style>/<section>` where the current row was balanced; a mid-token tail
where the current row ended on a closed tag — each gated on *truncation cannot
grow an artifact*, and calibrated against all 29 live `component_versions`
transitions (1 block, 0 false positives). Verified against this incident's REAL
artifacts: the 1,253-char wreck blocked on all three; the 6,771-char intermediate
blocked on structure alone (it passed the size floor at 66% — the size check by
itself is NOT sufficient); the restore correctly allowed.
(b) **`stop_reason` decoded** in `GenerateText` so a capped completion hard-errors
(`f32b208e5`, `bugs_open/008`).
(c) Ceilings last — migration **168** raised `improve_tool`/`generate_tool_html`
8000 → 32000. That was the exposure, never the bug.

**Interaction to check before trusting any `max_tokens` change** (`bugs_open/009`):
a **root** `ai_service` block SHADOWS the step-level block — where a root block
exists, per-step `max_tokens` is dead config and the effective value may be the
hardcoded 2048. Verify which one is live before and after your migration:
```sql
SELECT type, (default_config #> '{ai_service}') AS root_block,
       default_config #>> '{workflow,steps,<step>,config,ai_service,max_tokens}' AS step_max
FROM agent_definitions WHERE type = '<agent>' AND is_active;
```
(The four tool-pipeline agents have NO root block, so 168's step-level change is
live — confirmed by the log showing `max=8000` exactly matching the step value.)

**`error_step` must be set INSIDE a step's `config`, not at the step's top level
(2026-07-20).** `pkg/models/contracts.go` declares `ErrorStep string
json:"error_step,omitempty"`, and `routeToErrorStepOrFail` checks it *first*
("preferred location"). Reasoning from that struct is wrong: the top-level key
does **not** survive into `orchestration_states.workflow_plan`, so at runtime the
coordinator finds neither field and calls `failWorkflow`. Proven on one run —
same stored plan, two steps:

```sql
SELECT workflow_plan #>  '{steps,update_component}',   -- no error_step key at all
       workflow_plan #>> '{steps,append_note,config,error_step}'  -- 'complete', survives
FROM orchestration_states WHERE orchestration_id = '<run>';
```

Symptom when you get it wrong: the step's own failure is correct and its side
effects are correct, but the workflow dies at that step instead of routing, so
every *downstream* consequence (work-item status, notes, compensation) silently
never happens. Easy to miss precisely because the interesting part worked.
This is the "config key read on a different path than it's set" invariant, and
the antidote is in the same file you are editing: **every other step in that
workflow already showed the working shape.** Copy the working example rather
than deriving one from the struct. (`bugs_closed/012`, migrations 169 → 170.)

**Transferable rules.**
1. `output_tokens == max_tokens` is truncation until proven otherwise — never
   read it as "the model finished".
2. A fix agent's own note is not evidence the artifact survived. Check the
   artifact's structure after any whole-file rewrite.
3. Never let a repair path shrink a durable artifact without a guard. "Docs never
   fail the work" has a sibling: **a fix must never destroy the work.**

**Cross-refs.** `bugs_open/012` (this incident, full evidence), `bugs_open/008`
(stop_reason), `bugs_open/005` (article body), `bugs_open/009` (root-block
shadowing), migration `168_component_writers_max_tokens.sql`,
RUNNING_NOTES_travelling_docs 2026-07-18. Category tags: `llm-truncation` (new),
`silent-noop-success`, `fix-destroys-artifact` (new), `durable-vs-rendered`.

### A defect signal that names the ANCESTOR makes the fix loop non-convergent (2026-07-17)

**Symptom.** The same check fails run after run. Each cycle the fixer reports
success, deploys, and the behavioural tier re-verifies RED with an identical
message. Seen on `tool-loot-table-balancer`: `mobile-fit@mobile` failed with
"widest offending element: **fieldset (419px)**"; `tool-improver` constrained the
fieldset (`width:100%; min-width:0`) — twice, on two separate cycles, the second
time while loading its own prior fix note — and the overflow never moved.

**Diagnose.** Ask what the signal actually names. The overflow was measured on the
widest element crossing the viewport edge, which is usually the ancestor that
*inherited* an overflowing descendant's width, not the element whose intrinsic
width forces it. The fixer can only act on what it is told, so it kept
constraining a container that was never the cause. The real culprit here was a
grid two levels down (`div.ltb-row-grid`) whose items keep the default
`min-width:auto` and refuse to shrink below content.

**Root cause.** Attribution granularity. Naming the widest offender is enough to
decide WHOSE problem it is (tool vs site chrome — that part was already right) but
not enough to say WHAT to change.

**Fix (shipped v1.0.1135).** Drill down: from the widest offender, descend through
the children that themselves cross the viewport edge; along that chain name the
outermost layout container (grid, or flex that cannot wrap) as the fix target —
else the deepest crossing leaf — and state WHY it will not shrink. New
`forced_by` / `forced_reason` ride the check result, the failure detail, and the
work-item spec (`overflow_forced_by`, `overflow_fix_hint`). The signal went from
`fieldset (419px)` to *"forced by div.ltb-row-grid [grid layout — a grid item is
not shrinking; set min-width:0 on the items or let the grid wrap]"*, and the next
fixer run root-caused to the grid instead of the fieldset.

**Second defect this exposed: nothing bounds a non-converging loop.** Each failed
verdict raises a FRESH work item, so a per-item attempt cap never engages; only the
7-day verdict cooldown gates re-tries. A defect the one-shot fixer cannot solve
re-fails weekly forever with no escalation. Candidate fix: track attempts per
(subject, criterion) and route to `needs_human_review` after N non-converging
cycles (`bugs_open/010` candidate b, still open).

> **UPDATE 2026-07-20 — candidate (b) BUILT (`b13238be6`, inert until the next
> chassis image), and this entry's framing needs one correction.**
>
> **The correction.** The evidence above — "two separate cycles, the second time
> while loading its own prior fix note" — reads as proof the fixer cannot aim.
> It is not. The live page had **never re-rendered since the tool was born**, so
> both re-verifications tested an unchanged page (`bugs_open/024`, a five-defect
> delivery chain). Transferable rule 1 below still holds and the drill-down fix
> was still right; but the same rule has a sharper form: **before blaming a
> fixer for not converging, confirm its output actually reached the thing being
> re-tested.** A fix loop has at least three places to fail — the signal, the
> fix, and the DELIVERY of the fix — and a RED re-verification looks identical
> in all three.
>
> **The counting design, which is the transferable part of (b).** A cross-cycle
> attempt count is easy to write and easy to get subtly wrong; three bounds
> earned their place, each against a real row in the live DB:
> - **exclude in-flight work.** Count only terminal attempts. An item still open
>   is the CURRENT cycle — counting it escalates a loop mid-flight. There was a
>   live open item on the benchmark tool the day this shipped, put there by
>   another session.
> - **reset on success.** Count only since the subject last PASSED, or a
>   regression months later inherits an old tally and escalates immediately.
> - **match the specific criterion.** A fixer that fixed X and left Y has not
>   failed at Y twice. Counting "attempts on this subject" conflates them.
>
> **And fail OPEN.** If the count query errors, do the un-guarded thing. A guard
> that fails closed converts a transient DB error into a silently dropped fix —
> strictly worse than one extra cycle.
>
> **The residual risk, which is general to escalation guards:** stopping a loop
> only helps if someone reads the escalation. A `needs_human_review` item with
> no handler and no dashboard affordance is a no-op that ALSO stops the retries.
> Ask where the escalation will be READ before deciding the loop should stop.

**Transferable rules.**
1. Before blaming a fixer for not converging, read the signal it was given — a
   loop that repeats the same wrong fix is usually being told the wrong target.
   **And confirm its output reached the artefact being re-tested** (2026-07-20).
2. "Which element is widest" and "which element is the cause" are different
   questions in any inherited-geometry system.
3. An escalation path is part of a fix loop, not an optional extra.
4. **A retry cap scoped to a record cannot bound a loop that creates a new record
   per attempt.** Wherever "attempts" are counted on the work item, ask what
   happens when the next cycle brings a new one. This is not specific to tool
   acceptance — it applies to any detect→fix→re-verify loop whose detector
   raises a fresh item each pass.
5. A cross-cycle count must exclude in-flight attempts, reset on success, and
   match the specific failing criterion (see the update block above).
6. Guards fail open. The un-guarded behaviour is the safe default, because the
   guard is the new and less-tested code.

**Cross-refs.** `bugs_open/010`; `run_checks_action.go` (`HorizontalOverflow`),
`tool_acceptance_actions.go` (judge threading). Category tags:
`attribution-granularity` (new), `non-convergent-fix-loop` (new).

### Applied-but-unrecorded migrations block the runner and wear someone else's name (2026-07-16)

**Symptom.** `run-migrations.sh` halts on a file that fails with a duplicate-key
error, and — because the runner stops at the first failure — every later
migration, from every workstream, is gated behind it. It reads as "someone
committed broken SQL".

**Diagnose.** The file is not broken; it has ALREADY been applied by hand and was
never recorded in `schema_migrations`, so the runner replays it. A duplicate-key
error (SQLSTATE 23505) on replay is the classic signature. Verify the file's
artifacts exist in the DB before concluding anything:
```sql
SELECT filename FROM schema_migrations ORDER BY filename DESC LIMIT 10;
-- then check the file's own objects (component/agent/rows) actually exist
```

**Root cause.** Two defects, both needed: (1) an out-of-band apply (`psql -f`)
records nothing — the runner only ledgers what IT applies, and there is no tooling
for registering a manual apply; (2) the migration was not idempotent — the shop
convention is that every migration carries its own guard `DO` block, and a guarded
file replays as a harmless no-op.

**Fix.** Verify each file's artifacts live, then backfill the ledger rows
(`applied_by='ledger-backfill'` plus a note citing the evidence). Longer term:
a `--record-only <file>` flag on the runner, and a failure message that names the
already-applied possibility when the error is a duplicate key.

**Rule.** **Whoever applies a migration records it** — in the same sitting. An
applied-but-unrecorded migration becomes a roadblock with the original author's
name on it, and the diagnosis cost is paid by whoever hits it next.

**Fixed 2026-07-20** — the runner now has `--record-only <file> --note "<why>"`
and a failure message that names the already-applied possibility. Shell only, so
live on commit.

**A companion trap in the same file: a file-selection pattern decides what is
"work", so anything that merely *looks* like the work item gets done to you.**
The runner's candidate regex `^[0-9]{3}_[A-Za-z0-9_]+\.sql$` matched
`180_..._ROLLBACK.sql` and `180_..._VERIFY.sql` — hand-run companions, never
meant to be auto-applied — so both were listed as pending migrations. The
rollback strips its migration's keys and deletes its ledger row; its guard is no
help, because the guard fires when the migration is *absent* and it was present.
Generalise: **when a glob or regex defines a work queue, ask what else in that
directory satisfies it** — the dangerous match is the one whose name is *more*
specific than the pattern, not less. And when you exclude something from a
queue, **print what you excluded**: a silent skip is indistinguishable from an
empty directory. Found by taking a baseline dry run before starting the fix,
which is the only reason it surfaced at all.

**Cross-refs.** `bugs_open/007`. Category tags: `migration-ledger` (new),
`shared-state-coordination`.

---

### A routing value resolved from workflow state is only as good as the workflows that populate it (2026-07-18)

`resolveGitRepoName` chose a site's deploy repo as: explicit step config → `site_record.github_repo`
**from CollectedData** → default `"sites"`. Only planner-tier workflows run `ensure_site_record`;
`page-rerender`, `build-dispatch-loop` and `content-feed-orchestrator` do not. So a correctly-marked
VM-hosted site had *every* artefact — pages, news JSON, nine images — committed to the default B2 repo,
reporting success each time, while its box kept serving the old page. The deploy target was in practice
a property of *which workflow happened to be committing*, not of the site.

Resolve per-entity routing from the **entity row** (`SELECT github_repo FROM sites WHERE domain=$1`),
not from accumulated workflow state; `ActionParams.DB` is already there. Two corollaries that cost a
second debugging round each:

- **Grep for explicit config that already sets the same key before trusting a new fallback.** Three
  agent definitions pinned `repo_name: "sites"` in their `git_commit` step config. Explicit config
  rightly outranks the fallback, so shipping the fix changed nothing. A pin whose value equals the old
  default reads as harmless documentation and silently defeats the new logic — it is invisible precisely
  because it is redundant.
- **A repo→host sync with `rsync --delete` will delete live files the repo doesn't know about.** Commit
  the target's current state *before* the first pipeline deploy (checksum-verified), or the first
  successful commit wipes what was hand-deployed earlier.

### A mistyped routing key produces silence in every gate at once, not one loud failure (2026-07-18)

A site's news-listing page was created as `page_type='section-index'` instead of `'news-index'`.
`page_type` is a routing key, not a label, and three independent mechanisms each select on it:
`render_news_section` emits the archive JSON only for `news-index`; `MissingNewsPageCheck` fires only
when no `news-index` exists; `page-build-handler` assigns no news component to a `section-index`, so
`sections` stayed `[]` and the build no-op'd (the zero-sections family above). Each mechanism behaved
correctly for its own key; collectively they left an empty page that never built, and the symptom
surfaced somewhere else entirely — a **404 on a live nav link**.

When several gates key on one value, a wrong value is silent everywhere rather than loud once. Diagnose
by asking *which key does each mechanism select on*, then compare that to the row — do not follow the
visible symptom, which will be downstream and in a different subsystem. Corollary for remediation:
prefer **adopting/re-typing** the existing row over the check's `approach=new_page`, which would mint a
duplicate (an English `/news.html` beside the Spanish `/noticias`) and leave the orphan in place.

### A generator that commits LLM whole-file output must format it before the gate — un-`gofmt`'d code fails the gate and yields no PR (2026-07-18)

**The family.** Another member of the truncation family above (`005/008/012`), one
layer up: where those are about a truncated body being *stored as success*, this is
about a body that is **correct but cosmetically unformatted** being *committed and
then rejected* — and about the fact that the same check cannot tell "misaligned" from
"truncated" apart from the parser.

**Symptom.** A fix-implementer run whose generated code is logically correct produces
**no PR**. The build gate log ends `gofmt FAILED for: <file>` and exits — the run
(LLM whole-file generation + git push + a k8s clone/build Job) is spent for nothing,
and a human must hand-finish. Seen 2026-07-18 on BUG A's first implementer run
(`70680566`): the model inserted a new `StopReason` field into a struct but did not
re-align the sibling `Usage struct` field, and left a trailing blank line — two edits
`gofmt -w` fixes instantly. LLMs misalign adjacent struct fields routinely, so this is
not a rare tail.

**The point.** The gate is RIGHT to fail loud — `gofmt -l` as a read-only verifier is
its charter ("no PRs for broken code"). The defect is **upstream**: the commit-prep
step commits the model's bytes verbatim. `diagnose_prepare_fix_commit_action.go`
assembles the `files` map (path → whole-file body) at L150 and begins payload assembly
at L166 without passing any `.go` body through `gofmt`. Format at write-time
(`go/format.Source` over each `.go` body between those lines), and the whitespace class
can never reach the gate; keep the gate's `gofmt -l` as belt-and-braces. Bonus: a body
that `format.Source` cannot *parse* is almost always a `max_tokens` truncation — fail
loud there with the path (cheaper and clearer than failing at `go build` in the Job),
which extends the same "fail loud on truncation" posture the envelope guard at L132-135
already holds. Full case + fix sketch + test: `bugs_open/013`.

**Heuristic.** When an automated writer feeds a downstream *verifier* (gofmt, a linter,
a schema check), the writer must satisfy the verifier's contract at write-time. A
verifier catching trivia is the writer's missing normaliser, not the verifier's bug —
and an "unparseable/unformattable" body is a truncation signal, not just a style miss.

### `$HOME` does not redirect `ssh` — it expands `~` from the passwd entry (2026-07-18)

**Symptom.** Running git-over-ssh as a service account fails in two ways that both point
at the credential: first `Host key verification failed`, then (once a host key is
accepted interactively) `Could not create directory '/var/www/.ssh' (Permission denied)`
followed by `git@github.com: Permission denied (publickey)`. The GitHub Deploy Key page
shows the key **"Never used"** the whole time — ssh never reached authentication.

**Mechanism.** OpenSSH resolves `~` via `getpwuid()`, **not** `$HOME`. So
`sudo -u www-data env HOME=/var/lib/sitesync git clone …` configures *git* (which does
read `$HOME`) while *ssh* still looks in `www-data`'s passwd home, `/var/www` — where the
key isn't, and which the account cannot even create. One cause, two symptoms wearing
different hats: a trust failure and an auth failure.

**Do this.** Never point ssh with `HOME`. Name the paths explicitly and hand git the same
command — and remember every *later* invocation needs it too (a systemd timer running
`git fetch` as the same account fails identically, so a hand-fixed clone still leaves a
dead sync):
```bash
export GIT_SSH_COMMAND="ssh -i /path/.ssh/id_ed25519 -o IdentitiesOnly=yes \
  -o UserKnownHostsFile=/path/.ssh/known_hosts -o StrictHostKeyChecking=yes"
```
`ssh -v -T git@host 2>&1 | grep -iE 'identity file|known hosts'` prints the paths ssh
*actually* opened — the one diagnostic that settles this in seconds.

**Two heuristics this cost.** (1) *"Not found" is about the path, not the contents.*
"Host key verification failed" means the key wasn't found **where ssh looked**; it says
nothing about the file you wrote. Inferring an empty file from an absence message — and
then "fixing" the file's contents — is the same error as the guide's "0 rows is not
decisive", applied to a filesystem path. Check *which* path first. (2) *A probe's exit
status can lie*: `ssh-keyscan` exits 0 having written nothing, so `set -e` never fires and
a suppressed stderr hides it; assert the artefact is non-empty rather than trusting the
exit code (sibling of `003`'s "don't swallow kcat stderr"). Full case: `bugs_open/016`.

### Acknowledging a message before doing the work makes every restart destroy it (2026-07-18)

**Symptom.** Orchestrations vanish across a deploy or pod restart — not delayed, *gone*.
Parents wedge at `EXECUTING_STEP` on a spawn step with no child row, forever (22 of them
when this was found, the oldest 1,224 hours). The folklore rule "no orchestration dispatch
within ~300s of a chassis (re)start — the spawn is silently dropped" is this bug wearing a
workaround.

**Mechanism.** `platform/kafka/consumer.go` `Consume()` fetches (L81) and **commits the
offset (L103) before the handler ever runs** — the handler executes back in the caller,
after `Consume()` returns. The comment directly above the commit says "After successful
processing"; no processing happens there. Intent at-least-once, behaviour **at-most-once**:
once committed, Kafka will never redeliver, so anything in flight when the process dies is
unrecoverable. Both chassis loops use it (`agent.go:468` requests, `:528` responses), while
`client.go`/`server.go` in the same package use the correct fetch→process→commit pattern —
the asymmetry is the tell that it's an oversight, not a design.

**The transferable rule.** *An acknowledgement is a promise that the work is done — never
send it before the work is done.* This generalises past Kafka to any at-least-once
transport, job claim, or `status='complete'` write. Two corollaries worth internalising:

- **Read the ordering, not the comment.** A comment asserting "after successful processing"
  sat directly above a commit that ran before processing, and survived because the words
  matched the intent. When a delivery guarantee matters, follow the control flow.
- **"It only happens on deploys" usually means state is lost on process death.** Deploys are
  merely the most *frequent* way a process dies; OOM kills, evictions and node failures hit
  the identical path. Fixing the deploy choreography (drain windows, quiet checks) treats
  the schedule, not the defect.

**The same defect usually appears more than once — go and look.** Here it did. One layer up,
the `processed_messages` dedupe that redelivery would depend on *also* records receipt rather
than completion: seen-check then **record** (`agent.go:801/811`, `processor.go:1296/1317`),
and only then the work (`agent.go:822`, `processor.go:1323+`). So fixing the offset commit
alone would have been **inert** — the redelivered copy is dropped as "Duplicate message
ignored" and the work is lost through a different door. An acknowledgement layer and a
dedupe layer fail the same way because they encode the same mistaken belief: *that receiving
a thing is the same as handling it.* Whenever you find this at one layer, audit the layer
above and below before costing the fix. (The first version of this entry claimed the dedupe
"makes the fix small"; verifying it showed the opposite — which is itself the lesson.)

**Recording completion has a shape.** "Move the write after the work" reopens a window where
two in-flight copies both pass the seen-check. The durable form is a **two-phase claim**:
record on receipt with a lease, mark complete on success, treat an expired-but-incomplete
lease as reprocessable — which is what `site_work_items` already does with `claimed_at` and
its 40-minute claimed-item timeout. Also check what happens when the dedupe key is absent:
here an empty `request_id` silently disables dedupe entirely
(`platform/orchestration/state.go:163–165, 188–190`), so those messages have never had
protection at all. Full case and fix shape: `bugs_open/003` §3d/§4.4.

### An action that exists in code but in no registry fails as "requires a topic" — and the failure is stamped 'complete' (2026-07-18)

**Symptom.** A handler agent's work items complete on schedule, yet the defect they
describe never changes. The item's own `result` holds the confession:
`response.status='failed'`, `response.error='WORKFLOW_INVALID: … step X with action
Y requires a topic'`.

**Mechanism, leg 1 (the misleading error).** Workflow validation
(`platform/validation/workflow.go:69,80`) classifies any action it does not find in
`actioncheck.IsLocalAction` as *remote* and demands a Kafka `topic`. So "requires a
topic" almost never means a topic is missing — it means **the action was never
registered**. `fix_forced_text_colors` is the proven case: handler + input spec
written, no `GlobalActionRegistry` entry (same never-registered family as
`checkpoint_for_review`, registered by another thread in `0540698a4`).

> **CORRECTION (2026-07-18, the thread that fixed it).** The first version of this
> entry — and `bugs_open/017` — blamed *two hand-maintained rosters drifting*:
> `registry.go` versus the DEPRECATED `actioncheck/local_actions.go` list. **That was
> wrong.** `actioncheck.IsLocalAction` (`actioncheck.go:20`) delegates to a checker
> that `registry.go` installs at `init`; the `LocalActions` map's own lookup was
> commented out (`local_actions.go:185-188`) and the map had **zero live references
> repo-wide**. There was only ever ONE live list. The dead map has now been deleted,
> along with the `batch_webscrape_action.go` header comment instructing authors to
> "add to TWO places" — the comment that seeded the belief and survived to misdirect
> this very diagnosis. **The transferable lesson is the sharper one:** when a
> deprecated-looking list sits next to a live one, confirm which one the caller
> actually reaches before theorising about drift — `grep` the symbol for live
> references. A plausible story about two rosters cost more than reading the ten-line
> delegation would have.

**Prevention (shipped).** `actions/registry_parity_test.go` fails the build if any
action registering an `ActionInputSpec` has no `GlobalActionRegistry` entry, with an
explicit `dormantActions` allowlist for the deliberately-unreachable ones — so the
next orphan is a red test, not a production silence.

**Mechanism, leg 2 (the lie).** `CompleteWorkItemAction`
(`load_work_item_actions.go` ~735-800) verifies before completing only for item
types with a registered verifier, and never reads the response it stores: a payload
whose own `status` is `failed` is written next to `status='complete'` in one UPDATE.
The item_key dedup then suppresses re-detection — churn that reads as progress.

The root confusion is **two different `status` fields one level apart**. The envelope's
`response_status` is set to `'complete'` by the coordinator (`coordinator.go:2398-99`)
whenever *any* reply arrives — it records **delivery**. The saga's own verdict lives at
`response.status`. Reading the outer one as success is the bug. Generalise it: *whenever
a stored envelope wraps a payload, name which layer you are trusting.*

**Fixed (2026-07-18), leg 2 is now guarded.** `handlerReportedFailure`
(`complete_work_item_verification.go`) blocks completion when `response.status` is
`failed`/`failure`/`error`, routing the item into the existing attempt machinery
(`attempt_count+1` → `triaged`, or `failed` at max attempts) instead of `complete`.
Deliberately keyed on an explicit failure verdict and **not** on the presence of
`response.error`, because handlers legitimately carry a non-fatal error string beside a
successful outcome. The predicate was chosen against live data, which is the reusable
move: `GROUP BY response.status` over completed items returned 2905 rows with no
`response.status` at all and 54 with `failed`, and nothing in between — a guard with a
measured empty gap between the two populations cannot mis-fire on the success path.

**Diagnose in one query.** `SELECT id, item_type FROM site_work_items WHERE
status='complete' AND result->'response'->>'status'='failed';` — any row is this
class. On 2026-07-18 it returned **54 rows across 6 sites and 4 item types**, back to
May; the filed case had spotted 2. After the guard ships this should only ever return
rows predating the deploy. And when you see "requires a topic": grep
`GlobalActionRegistry` for the action name before touching topics.

**A second cause reaches the same query.** One of the 54 was `render_js_snippets`
where the registry has `render_js_snippets_for_site` — a **seed naming an action that
does not exist**, not an unregistered action. Registration fixes one; only the leg-2
guard covers both. Fix the class, not the instance.

**Cross-refs.** `bugs_open/017` (case, now closed). Kin: "Trust the rendered artefact,
not the status" (§ durable invariants) — this is the mechanical version for work items.
Category tags: `never-registered-action`, `false-complete`, `envelope-vs-payload`.

### A command that reads stdin truncates the `while read` loop that calls it (2026-07-18)

**Symptom.** A shell loop that should process N items silently processes the first
few and stops. No error, no warning — the output simply looks like there were
fewer items. Seen in the council coverage report (`098_REPORT_unreviewed_commits_v1.sh`):
it announced **"In-scope commits found: 4"** when the identical `git log` query
returned **41**, hiding 90% of unreviewed platform commits from the report whose
entire purpose is to surface them.

**Diagnose.** Look for a stdin-reading command inside the loop body:
```
while IFS='|' read -r a b c; do
    kubectl exec -i POD -- psql ...      # <-- `-i` reads STDIN
done < <(git log ...)                    # <-- the loop's stdin IS that stream
```
`kubectl exec -i` / `ssh` / `psql` without `-c` / `ffmpeg` consume whatever remains
on stdin — here, the rest of the git log — and the loop then ends normally.
Tell-tale: the cut-off tracks the FIRST item that triggers the inner command (here,
the first commit carrying a trailer), so the truncation point moves as data changes.

**Confirm cheaply.** Run the loop with the inner command disabled (this script
already had a `NO_DB=1` path) and compare counts — 41 vs 4 settled it in one
command. Or count iterations inside the loop and diff against the source query.

**Fix.** Close the inner command's stdin (`< /dev/null`), drop the interactive flag
when it is not needed (the fix here: the SQL was passed with `-c`, so `-i` was
pointless), or feed the loop from a dedicated descriptor
(`while read -u 3 ...; done 3< <(cmd)`). A sibling call OUTSIDE the loop may
legitimately need `-i` (a heredoc-fed `psql`) — fix the one in the loop, not both.

**Why it earns a §9 entry.** This is "0 rows is not decisive" applied to shell: a
truncated loop reports a plausible smaller number rather than failing, so a
confident wrong conclusion follows ("coverage is good", "only 4 commits touched
platform"). Any report that counts things should have its total cross-checked
against its source query at least once.

**Cross-refs.** `bugs_closed/018` (coverage report — CLOSED 2026-07-21; the *idea.uk*
`018` is a different, still-open case — resolve by slug). Category tags:
`stdin-theft` (new), `silent-truncation`.

### A capability the platform already has, believed missing because one enum value routes elsewhere

**Symptom.** "The system can't do X." A generated artefact is visibly bad in a way that looks
like a hard limit of the underlying model or library, and the obvious inference is that a new
subsystem must be built to do X properly.

**What it actually was (2026-07-18, imagery).** A site's hero image came back as a
convincing-looking flowchart full of gibberish words, which reads as proof that image models
cannot render text. It was not a model limit: the provider is chosen by an enum
(`kind`), and `hero` fell through to the photographic model (SDXL, genuinely unable to render
text) while `infographic` had always routed to a model that renders text well
(`gemini-3-pro-image-preview`). The capable lane existed, was correctly wired, and had simply
never been called. The first attempt through it produced a publishable infographic.

**The transferable pattern.** Before concluding a capability is missing, **read the dispatch
table, not the output**. When behaviour is selected by an enum/kind/type string, one value
routing to a weaker backend is indistinguishable — from the artefact alone — from the whole
capability being absent. Ask: *which branch did my request actually take, and what else is in
that switch?* Two cheap greps (the switch, and the deployed model/env var) answer it.

**Corollaries.**
- A generalisation drawn from one backend ("this class of model can't do text") silently
  expires when a backend is upgraded. Re-test the premise before designing around it; the
  cost of being wrong is building a subsystem you already had.
- Output quality through a capable lane is dominated by **request specificity**. The same
  model produced rubbish from a thin prompt and production-quality work from one that named
  the layout, the exact copy, the permitted figures and the palette. If a lane "doesn't work",
  check what you asked it for before replacing it.
- Good-but-imperfect is the new failure mode, and it is harder to catch: the capable model
  still occasionally misspells a word inside an image, and nothing in the pipeline reads
  rendered text. Success signals do not cover artefact correctness — see "Trust the rendered
  artefact, not the status".

**Cross-refs.** `bugs_open/011` (routing fix, legibility guard, evidence-base-driven figures).
Category tags: `dispatch-table-not-output`, `expired-generalisation`, `unused-lane`,
`request-specificity`.

### A dispatch table's `default:` branch is a silent bug factory — the same one, once per new enum value (2026-07-18)

**Symptom.** The same class of defect ships two or three times, months apart, each found by a
human noticing a bad artefact rather than by any signal. Each instance looks like a one-off
oversight ("nobody added this case"), so each is fixed by adding one more case to the switch.

**What it actually was (2026-07-18, imagery provider routing).** Provider selection was a
hand-maintained `switch` on `kind` whose `default:` fell to the weaker provider. `content_hero`
was never added → shipped mis-routed. `hero` was never added → shipped a gibberish diagram as a
client homepage, and it was the **largest** kind on the fleet (84 of 155 planned images). The
switch's own comment already blamed itself in writing — *"a new kind that nobody adds to the
Banana branch falls to Stability… with no error anywhere"* — and the next fix was still going to
be "add `hero` to the list". The defect is not the missing case. **The defect is that a missing
case is indistinguishable from a deliberate one.**

**The transferable pattern.** A `default:` branch that silently produces *degraded but valid*
output is unfalsifiable at runtime: nothing can tell "routed here on purpose" from "nobody ever
routed this". Fix the mechanism, not the instance — **make the handled set enumerable data**
(a map, not switch cases), so the code can ask *"is this value actually routed?"* and say so
when it isn't. Adding a value stays a code change; **forgetting to stops being silent.**

**Getting the guard right.**
- **Warn on the unhandled case, not on the fallback itself.** Here, an *empty* kind is a
  documented legacy path that legitimately uses the fallback, so it must not warn — a warning
  that fires constantly is one nobody reads, which is how you lose the signal you just built.
- **Name the value and list the valid set in the log line.** Whoever reads it at 2am should not
  need to open the source to learn what was expected.
- **Still route.** An unknown value must be *visible*, not fatal — degrade and shout.
- **A log line is the floor, not the finish.** Detection that lives only in process logs still
  depends on someone tailing the right pod. The durable form is a row in `agent_error_log` or a
  `site_work_item`. **Trap:** if the detecting service has no DB handle (adapters here don't),
  do **not** relocate the detection to a service that does — separate services run separate
  images, so the second copy of the table drifts from the deployed one. Have the deciding
  service *report* the condition and let the one with the DB persist it.

**How to find these before they bite.** Grep for `default:` in provider/handler/renderer
selection and ask of each: *if a new enum value reached this, what would tell me?* If the answer
is "a human eventually looks at the output", it is this bug, already written and waiting.

**Cross-refs.** `bugs_open/011` §6 (the fix, and the persistence still owed);
"A capability the platform already has…" above (the same routing table seen from the diagnosis
side); "Every keyed work-item insert fails 42P10…" (the sibling failure where two hand-maintained
lists drift apart). Category tags: `silent-default`, `enumerable-not-switch`,
`degraded-not-failed`, `fix-the-mechanism-not-the-instance`.

### A pod-grep marker that the build does not retain reads exactly like a stale deploy — validate the marker before you believe its absence (2026-07-18)

> **This entry replaces an earlier version of itself.** The first version claimed the
> image-generator-adapter had shipped stale under a good tag, "proven three independent
> ways". That claim was **WRONG**, and the three proofs were the same invalid test run
> three times. The corrected finding below is the useful one — and it undermines a
> verification habit the whole fleet relies on, so it matters more than the original.

**Symptom.** You pod-grep a binary for a symbol from your change, per the CLAUDE.md rule.
It returns **0**. You grep a much older symbol as a control — also **0**. That reads as
overwhelming evidence the image is stale garbage, and it is very hard to argue with.

**What it actually was (imagery D14).** `strings /app/image-generator-adapter | grep -c
content_hero` returned 0, and so did the "old symbol" control `sprite_sheet` (weeks older).
I concluded the shipped adapter predated both changes, rebuilt and re-released it. The next
release did the same thing — which is what finally exposed the real cause:

| marker | shipped image (Dockerfile build) | plain `go build` on host |
|---|---|---|
| `content_hero` | **0** | 1 |
| `sprite_sheet` | **0** | 1 |
| `infographic` | 1 | 1 |
| `banana` | 72 | 72 |
| `"dispatching to provider"` (log) | 1 | 1 |
| `"reference images will be IGNORED"` (log) | 1 | 1 |

The binary was **current all along**. Proof: it contains `"reference images will be
IGNORED"`, a string added to that same file *later* than `content_hero` — a stale binary
cannot contain the newer string and miss the older one. The Dockerfile builds with
`CGO_ENABLED=0 go build -a -installsuffix cgo` on `golang:1.24-alpine`; under those flags
some switch-case string literals are not retained as greppable standalone strings, while
log-message literals always are. Which literals survive is **not predictable by reading the
source** — `infographic` survived from the very same `case` clause that dropped
`content_hero` and `sprite_sheet`.

**The transferable pattern.** A pod-grep is a **positive** test only. A hit proves the code
is there; **a miss proves nothing until you have shown that marker survives a known-good
build of the same pipeline.** Before concluding "not deployed", run the control:

```bash
# Extract the shipped binary and test the marker against a build you trust.
CID=$(docker create <registry>/<svc>:<tag>); docker cp $CID:/app/<svc> ./shipped; docker rm $CID
go build -o ./control ./cmd/<svc>
for m in "<your marker>" "<a log line you know is in it>"; do
  echo "$m: shipped=$(strings ./shipped | grep -c -- "$m") control=$(strings ./control | grep -c -- "$m")"
done
```
`shipped=0 control=0` means **your marker is useless**, not that the image is stale.

**Choose markers that survive.** Prefer, in order: a **log message string** you added; an
error message; a long distinctive identifier. Avoid: enum/`case` values, short tokens,
anything that might be merged or optimised into another literal. This is the practical
amendment to CLAUDE.md's "verify against the running pod" rule — the rule is right, but it
is only as good as the marker, and a badly-chosen marker fabricates evidence of a
non-existent outage.

**Also true (kept from the original, on its own merits).** `quick-agent-update` used to
build/push/deploy only the chassis while restarting other deployments, so a service sharing
the `kind` vocabulary could in principle drift across a version boundary. It now releases
the image-generator-adapter alongside the chassis (`c0ef457a1`). That change is sound
practice — but note it was **not** validated by a real incident, because the incident that
motivated it was this measurement error.

**Corollaries.**
- Three "independent" confirmations that share one method are one confirmation. All three
  of mine used the same absent marker; agreement between them measured nothing.
- Busybox `strings` (alpine pods) and GNU `strings` (host) differ; if you compare a pod grep
  against a host grep, you have changed two variables at once.
- The honest fallback when a marker is unavailable is a **runtime** check — here, the
  adapter's own log line proving `kind=content_hero → provider=banana` on a real request.
  Behaviour is evidence; string absence is not.

**Cross-refs.** CLAUDE.md § "Building & deploying images"; imagery `RUNNING_NOTES` Turns
49–51. Category tags: `marker-must-be-validated`, `absence-is-not-evidence`,
`positive-test-only`, `same-method-thrice`.

### A guard before an action switch breaks the one action that lacks the guarded field — at 100%, disguised as a timeout (2026-07-20)

**Symptom.** A workflow step calling an adapter fails with an EXPIRED awaited
request after retries, `__step_error` empty. Everything upstream looks perfect.
It reproduces every time, yet reads like a transient adapter timeout.

**Mechanism.** The webscrape adapter validated the top-level `url` before
switching on `action`. `batch_scrape` carries `data["urls"]` (array) and has no
top-level `url` by construction of its own sender — so every batch request was
rejected as "Empty URL in request" before reaching `handleBatchScrape`
(`bugs_closed/047`, CLOSED & LIVE v1.0.1145). The caller never sees the rejection: the adapter's error
response is not claimed as the awaited response (the 003/035-envelope family),
so each retry meets the same fate and the await expires silently. Blast radius
included the entire research lane (research-agent's scrape step).

**The transferable rules.**
1. Shared pre-switch validation must only assert fields that EVERY action
   requires; an action with a different payload shape branches out first.
2. A 100%-deterministic failure can masquerade as flakiness when the error
   channel is broken — before blaming timeouts, read the SERVING pod's log for
   the request id; one line there ("Empty URL") ended the mystery.
3. When adding an action to a shared handler, grep for guards ABOVE the switch
   you are extending — batch_scrape's author wired the case correctly and was
   defeated by a guard eight lines earlier.

### One truncated reviewer voids an entire council round — and six good reviews with it (2026-07-18)

**Symptom.** A council run reaches `COMPLETED @ complete_invalid` with no verdict,
no council report and no revise round, after every reviewer visibly executed. The
same terminal step is ALSO what a malformed submission produces, so the natural
reading ("my plan was rejected") is wrong.

**Mechanism.** `diagnose_council_decide` iterates the configured reviewer fields.
A field that is **absent** is treated as a principled abstention (the stage-3
relevance filter deliberately skips irrelevant seats). A field that is **present
but unparseable** hard-errors instead — and one seat cut at `max_tokens` produces
exactly that, so the action returns an error and the orchestration discards every
other seat's complete, well-formed review. Proven case: `review_guidelines` at
`output_tokens = max_tokens = 8000` binned six reviews of 2,738–5,548 tokens
including the guardian's and the bug-historian's. Four of seven seats used over
half the ceiling on that submission, so the margin is thin on exactly the
substantial changes most worth reviewing.

> **CORRECTED 2026-07-20 (bugfix-019 thread): the mechanism above is the MINORITY
> path.** Measured over 10 days of `complete_invalid` runs, 9 of 11 truncation
> voids never reached `diagnose_council_decide` at all — they died UPSTREAM at
> `execute_llm_prompt`, because the provider client returns `""` + a hard error on
> `stop_reason=max_tokens` (the partial is destroyed at the transport layer), the
> step fails, and its `error_step` routes the round to the terminal. Seats run
> SEQUENTIALLY with `review_editquality` first, so most voided rounds have ZERO
> completed reviews — the "six good reviews binned" case above is the rarer shape.
> Only 2 of 11 hit the `json.Valid` path this entry describes. The transferable
> lesson: when a failure can occur at two layers, count which layer before fixing
> either — the queued diagnosis item for this bug targeted the minority mechanism
> and would have sent the fixing thread to the wrong file.

**Tell the two `complete_invalid` causes apart** — read
`collected_data->'__step_error'` and look at `failed_step`:
`persist_submission` = your submission is malformed (operation must be one of
`modify|add|remove|config_change`; a file may appear in only ONE edit per stage);
any `review_*` step = that reviewer's CALL was truncated and the round died there
(the dominant case); `council_decide` = a truncated review reached the decider,
your plan was fine and fully reviewed.

**Wider rule this is the fourth instance of.** `output_tokens == max_tokens` means
the completion was CUT. The platform keeps detecting that *after* persisting or
acting on the fragment (005 article bodies, 008 stop_reason, 012 component
wreckage, now the council). When you add a step that consumes an LLM's structured
output, decide up front what happens to a truncated one — treating it as a loud
abstention that can never upgrade a verdict to "approve" beats both "trust it" and
"destroy the round".

**Three companion traps found while fixing it (2026-07-20).** (a) The councils log
under `agent_type='generic'`, so a diagnosis filtered by council names finds
nothing — the loop returned UNVERIFIABLE for exactly this reason; name `generic`
in any council-scoped symptom. (b) A NULL from a JSON path query is not evidence
of absence: the per-seat cap lives at `steps.<seat>.config.ai_service.max_tokens`,
and a query missing `->'config'` read all 13 seats as "unset". (c) Raising the cap
does not remove the class: an `experience-planner/compose` call truncated at a
32,000-token cap the same week.

**Fixed 2026-07-20** (`a3b606798` + migration 177, inert until image roll):
`TruncatedError` carries the partial out of aiservice; `tolerate_truncation`
step config lets a council seat degrade instead of aborting the chain; the
decider salvages a verdict or counts the seat `unreadable` and never approves
alongside one. **CLOSED 2026-07-20, verified live on v1.0.1140** — full case,
evidence and the corrected mechanism: `bugs_closed/019`
+ `docs024_key_docs_latest/bugfix_019_council_truncation/`.

### A strict field type turns "which one?" into a lost round — the wrong register is not garbage (2026-07-20)

**Symptom.** Identical to the truncation entry above — `COMPLETED @
complete_invalid`, `error` EMPTY, no `council_report`, every seat run and paid
for — but `__step_error` says `does not match the review schema: json: cannot
unmarshal string into Go struct field .objections.edit of type int`. The JSON is
**complete and valid**, so 019's salvage path cannot help: there is nothing to
repair. Same blast radius, different cause. (`bugs_open/036`.)

**Mechanism.** `councilReview.Objections[].Edit` was a plain `int` — a 1-based
index into the plan's edits, `0` = plan-wide. A model asked *which edit do you
object to?* answers in three natural registers and a plain `int` parses exactly
one of them. The other two returned an error from the per-seat loop, which
discarded every other seat's completed review at the **last step of the round**.

**The finding that reframes it — and the transferable one.** The three live
payloads were not malformed indices. They were `"plan-level (deploy
verification)"`, `"risks note on the 54 mis-stamped rows"`, `"risks/summary (item
5)"` — reviewers saying *this objection is about the plan, not any single edit*,
which the contract **already spells `0`**. The strict type was not rejecting
noise; it was **discarding a meaning it had a representation for**. So `json.Number`
— the obvious tolerant type, and the one the bug file first suggested — would have
parsed **none** of the three. When a strict field keeps failing, read what was
actually sent before choosing the tolerant type: the sender may be answering a
question your schema asked badly.

**Wider rule, and the third instance of it.** *One participant's malformed output
must cost one participant, never the round.* A council's whole value is that it is
many independent opinions; the same is true of any fan-out that aggregates. Every
per-seat failure mode should converge on one contract — recover the opinion if it
is there (mark it degraded), else count that seat **unreadable**, and fail only
when **zero** readable opinions remain. Enforce "cannot read its reviewers ⇒ must
not wave it through" at the *decision* (unreadable blocks an approval), not by
voiding the round: that way the guard can only make an outcome more conservative,
never less. Note the failure order — a per-item strictness bug fails at the
**aggregation** step, i.e. last and most expensively, so it is worth auditing
aggregators for `return err` inside a per-item loop *before* one bites.

**Go trap worth keeping.** `encoding/json` **continues past a TYPE error** (unlike
a syntax error) and keeps everything that did decode. So a mistyped field costs
that field, not the struct around it — a failed `Unmarshal` on valid JSON still
leaves usable data in the target. Field-by-field salvage retains far more than it
looks like it should, and a test asserting "the whole objection is lost" fails.

**Do not normalise a verdict.** An unrecognised verdict (`"approve-with-comments"`)
gets recorded unreadable, deliberately *not* mapped to the nearest legal value:
guessing what a seat meant is how a veto becomes an approval. Tolerance belongs on
the *pointer* fields, never on the decision field.

**Fixed and VERIFIED LIVE 2026-07-20** (v1.0.1140) by forced reproduction — a
scratch-council seat emitted `"edit": "edit 1 (comment-only change to …)"` and the
round reached `complete_revise` with a report written, where the same shape
previously gave `complete_invalid` and no report. The persisted report carries a
string and an int side by side, proving both the tolerant unmarshal and that the
reviewer's own token survives instead of being laundered to `0`. `/bugs_closed/036`.

**Verifying a "one bad participant" fix needs a FORCED bad participant.** A
passing round proves nothing here — the healthy path was never broken. The
discriminating test is to manufacture the exact malformed input and show the
aggregate survives, which is also what the sibling 019 case did (a scratch seat
capped at 200 tokens). Two costs worth knowing before you try it: a harness that
works by patching a reviewer's PROMPT is unreliable against a security-minded
seat — the guardian seat here read its own patched instruction as injected content
and refused to malform its output, correctly in spirit — so malform the payload
**below the model** (a stub step emitting fixed JSON) when the seat must
misbehave. And the council trigger's `TARGET_AGENT_TYPE` is hardcoded, so a
scratch run needs a COPY of the script, never an in-place edit of the shared one.

Category tags: `strict-type-loses-meaning`, `one-seat-not-the-round`,
`aggregator-fails-last`, `partial-decode-on-type-error`, `forced-bad-participant`.

### Two fields that must agree, with no relationship in the schema — the check you want cannot be written (2026-07-19)

**Symptom.** Live customer-facing buttons that have words on them and nowhere to go:
`href=""`, `href="#some-id"` where no such id exists, a link to a hostname that does not
resolve, and a label describing a completely different tool. Four buttons on one site, four
different mechanisms, and every one of them passed every gate.

**Mechanism.** A button is two schema fields — `cta_x_label` and `cta_x_url` — and **nothing
anywhere expresses that one implies the other.** The label is typically `source:static`,
which (`plan_sections_action.go:1210-1218`) writes its fallback and `continue`s, *bypassing
`required` and `on_missing` entirely* and re-applying every render. The URL may be absent
from the schema, unresolvable, empty, or LLM-authored. The template renders the anchor
ungated, so an empty value becomes `href=""` rather than no button.

**Why it hid.** Each failure mode fell in a different check's blind spot, and the blind spots
tile the space: `href=""` is a *warning* in `validate_page_content` (`:551-560`, and `:257`
blocks only on blockers/errors); `#frag` classifies as `LinkScopeAnchor` and is skipped by
phantom, misdirected and validate alike, with nothing anywhere resolving a fragment against
the page's ids; external URLs are skipped by every consumer and **no HTTP reachability check
exists in `platform/` at all**; and at content_data level `check_required_fields_missing.go:189-192`
skips any field whose `source != "llm"` — which is *every* CTA url field by design.

**Three transferable rules.**

1. **When two fields must agree and the schema encodes no relationship between them, the
   inconsistency is not merely undetected — it is unrepresentable.** No check can be written
   until the pairing exists as data. Reach for the pairing first; the validation is trivial
   afterwards. Corollary: a hardcoded map of "which components have CTAs"
   (`resolve_internal_links_action.go:91-98`, six entries, hand-patched by migrations 091,
   096, 097b and 098 — the same lesson four times) is that pairing written in the *least*
   durable place. Derive it from the schema.
2. **A `required:true` field whose source cannot answer is an instruction to fabricate.**
   `source:llm` + `required` + no fallback + nothing to look up = the model must return
   something, so it invents. Here it invented two different hostnames on two adjacent pages.
   **Audit for `required` fields whose source has no resolvable ground truth; that
   combination manufactures the fabrication you later hunt.**

   **And look at *how* it invents, because that is usually checkable.** Neither hostname was
   a wild guess. One was the obvious `.com` variant of the site's own name. The other,
   `leopardess.contactforsales.com`, was a **transform of a real contact email** in the
   site's own identity spec (`leopardess@contactforsales.com`, `@`→`.`). The parts were true
   and in-context; only the recombination was invented — the classic fabrication shape.
   That yields a *deterministic* check (string identity against data already held, no
   network call), where "is this hostname plausible?" yields none. **When you catch a
   fabrication, work out which real in-context tokens it recombined; that recombination rule
   is usually cheap to test for, and it generalises** — here to six sites sharing one contact
   domain. Corollary from the same case: an owner owning the fabricated domain does not mean
   the model knew about it. I made that inference, stated it as mechanism, and the owner
   corrected me.
3. **A finding filed at a status nothing consumes is indistinguishable from no finding.**
   One of these four was correctly detected, with the right component and page named, *two
   days before the owner clicked it* — and filed `needs_human_review`, which
   `TriageDetectedItemsAction` never promotes, no `handler_agent` claims, and
   `load_work_item_actions.go:804` excludes from re-open queries. Grepping `platform/` for
   `unresolved_cta` / `cta_names_unknown_destination` / `dead_control` returns emission sites
   only, **zero consumers**; 34 sat open on the one site. Detection gap and delivery gap are
   different bugs. **Before adding a check, grep for who consumes its output** — otherwise
   you convert a visible problem into a larger invisible one.

**A measurement worth repeating elsewhere.** The platform has a written invariant for this
(LNK-005: "an unresolvable destination renders nothing rather than a broken link"), and
**75 of 89 URL-bound CTA anchors in the component library violate it** (~84%, RUNBOOK R9).
An agreed invariant with no mechanical enforcement decays to a comment. When you find one
stated in a doc, measure its actual compliance before assuming it holds.

Full case, evidence and fix candidates: `bugs_open/023`. Plan, queries and fleet sizing:
`docs024_key_docs_latest/cta_link_integrity/`.

### A fix applied to one branch of a two-branch router reads as done — and the other branch keeps the bug (2026-07-19)

**Symptom.** A defect is filed, patched, closed in the notes, and stated as fixed in the
handoff — while the identical defect is still live a few steps away in the same workflow.
Nothing contradicts the "fixed" claim, because the path that was tested is genuinely fixed.

**The instance.** `bugs_open/016_…council_revise_prompts_drop_reviewer_output` finding 2
(resolve 016 by SLUG — `bugs_closed/016` is a different case, the ssh/`$HOME` one):
council revisers referenced seats one-by-one in
`input_fields`, so seats added later were invisible to them. The agreed fix was "read the
`council_report` artifact once" — roster-proof. On `fix-proposer` the load step was placed
**before** the routers (`council_decide → load_council_reviews → check_approved`), so every
downstream path inherited it. On `feature-designer` the same fix was wired onto the revise
branch only:

```
run_checks -> load_council_report -> repropose     <- fixed
check_reframe -> reframe                           <- still per-seat: 2 of 5 seats
```

`reframe` still named `review_editquality` + `review_guardian` and was blind to
bug_historian, guidelines and reuse_agent. The patch's own header asserted the agent was
"currently complete (5/5/5)" — true of the path it touched.

**Why it survives review.** A router has one entry and several exits. Reading the diff shows
a correct fix; reading the *routing* is what shows the exit it never reaches. Both patches
here were written by threads that understood the bug perfectly — the gap was placement, not
comprehension, and placement is exactly what a diff does not show.

**Diagnose.** Do not grep for the fix; grep for what the fix was supposed to remove, and
count what should have gone to zero:

```python
# per-seat refs that should no longer exist anywhere downstream of the council
for step in ('repropose', 'reframe'):
    refs = set(re.findall(r'review_[a-z_]+', json.dumps(steps[step])))
    print(step, refs or 'none')     # any non-empty set is this class
```

Then walk the graph from `start_step` following `next_step` / `then_step` / `else_step` and
confirm the shared step is reached on **every** path that needs it, not just the one you
exercised. That walk also catches orphaned and dangling steps for free.

**The general rule.** When a fix introduces a shared prerequisite step, place it **before the
branch**, not on a branch. One placement covers every current and future exit; per-branch
placement is a fix you must remember to repeat, which is the same non-idempotence the fix
was meant to remove — here it had already recurred once at the level above (per-seat prompt
refs), so the remedy reproduced the disease one layer down.

**MECHANICAL CHECK (added 2026-07-20).** This entry now has teeth:
`scripts/pattern-check.py`, run advisory from `.githooks/pre-commit`, flags a commit
that edits one Go function and not its near-identical twin (twins = names differing by
exactly one CamelCase segment; test doubles excluded). Measured before wiring in: **3
findings across 150 commits**, all three on one genuinely-paired function, no unrelated
hits. It also carries §9 #16 (gofmt), #20 (stdin-eating `while read`) and a short table
of declared co-change pairs.

The origin is the argument for it. This entry was written on the morning of 2026-07-19,
and the same session committed exactly this mistake eight hours later — fixing
`withPriorCodeRequests` and leaving `withPriorRequests`. The pattern was not forgotten;
it had just been written by the person who then broke it. **Knowing a pattern does not
fire it — something at the moment of the edit has to.** A reviewer council did catch it,
correctly, but spent two rounds and real credits on a question `grep` answers in
milliseconds. Most of §9 is judgement and cannot be checked this way; the few entries
that can, should be, so review effort goes where only a reader can help.

**Cross-refs.** `bugs_open/016_…council_revise_prompts_drop_reviewer_output` (by slug — see
above); `PATCH_feature_designer_018_reframe_reads_artifact.sql`;
`NOTES_running_fixloop(10).md` turn 41. Kin: "A dispatch table's `default:` branch is a
silent bug factory" (same family — the untravelled path is the one that rots).
Category tags: `partial-fix`, `router-branch-asymmetry`, `false-closure`.

### The pod-grep passes even when nothing shipped — grep a symbol your change CREATED, not one it merely uses (2026-07-20)

**Symptom.** You follow CLAUDE.md's deploy rule exactly — `strings /app/<binary> | grep -c
"<your symbol>"` against the running pod — and it returns a positive count. You sign off.
The check was worthless.

**Mechanism.** The natural symbol to grep is the *thing you were working on*: the action
name, the config key, the item type. But that string is usually already in the binary,
emitted by code that predates your change. Registering `fix_forced_text_colors` in
`GlobalActionRegistry` was the entire fix for bugs_open/017's leg 1 — and
`grep -c fix_forced_text_colors` returned **1 before the fix as well**, because the action
file had always called `RegisterActionInputSpec("fix_forced_text_colors", …)`. A pod
running the OLD image passes that check identically to one running the new one. The grep
confirms the subject exists, not that your change to it shipped.

**The rule.** Choose a symbol that **cannot exist unless your change shipped**:
- a new function name (`handlerReportedFailure`, `recordUnknownVerdict`);
- a new error code or constant (`UNKNOWN_HANDLER_VERDICT`);
- a **literal string you wrote in the changed line itself** — the registry entry's
  `Description`, the guard's error message. This is the best one for config-shaped
  changes that add no new identifier, and it is how 017 was finally verified:
  `strings … | grep -c "Strip forced child-text colours"` → 1, a phrase that exists
  nowhere but the registry entry the fix added.

**Always pair it with a control.** Grep an older, definitely-present symbol too. If the
control is also 0, your problem is `strings`/build flags, not the image — see the imagery
D14 entry above, where a 0-and-0 result triggered an unnecessary rebuild. Positive control
plus discriminating symbol is the pair; either alone misleads.

**Use it in reverse, too — to prove something is NOT live.** The same rule settles the
other question threads keep guessing at: *has someone else's fix shipped yet?* On
2026-07-20 an inbound handoff described `bugs_open/032` as an open defect with a fix
"drafted". The fix was in fact **written and committed** (`a467baa11`) — and pod-grep for
its discriminating string `"genuinely fixed or silently deleted"` returned **0**, with a
known-live symbol returning 1 as the control, and the commit timestamp postdating the pod
start. Conclusion in one command pair: *fixed but inert*, so it correctly stays in
`/bugs_open/` and its next action is an image roll, not code. Without the discriminating
symbol there is no way to distinguish "not shipped" from "shipped but I grepped the wrong
thing".

**Corollary — an inbound handoff is a claim about the PAST.** It describes the system when
it was written, and several threads commit here per hour. Before acting on one, re-verify
its load-bearing state claims against the code and the pod; before *forwarding* one into
your own handoff, do it again. The 032 handoff was ~28 hours old and already stale in its
headline item; copying it forward unchecked would have sent the next thread to reimplement
finished work.

**Kin.** "Trust the rendered artefact, not the status" (§ durable invariants); the
travelling-docs finding that verifying a fix by grepping a GENERIC CSS property always
passes — same error, different medium. Category tags: `deploy-verification`,
`false-positive-check`, `discriminating-symbol`, `stale-handoff`.

### A queued orchestration is indistinguishable from a dropped one — and "resubmit" is the expensive guess (2026-07-18)

**Symptom.** You dispatch a council run (or any orchestration). Nothing appears in
`orchestration_state_audit` for that `orchestration_id`. Two, then five minutes pass —
still nothing. A previous, apparently identical run had produced its first audit row
within ~10 seconds. Everything says the spawn was silently dropped, and CLAUDE.md
even documents a real drop mechanism (dispatch within ~300s of a chassis pod restart),
which makes the wrong conclusion feel confirmed.

**What it actually is.** The run is **queued**, not dropped. Under a deep backlog the
gap between dispatch and the first audit row stretched from ~10s to **~16 minutes** —
submitted 16:41, first audit row 16:57, then it ran normally to a verdict. Absence of
audit rows measures *queue depth*, not delivery.

**The cost of guessing wrong.** Reading it as a drop, I resubmitted three times,
including two "fixes" for hypotheses I had not tested — first that the ~27KB payload
exceeded what `kubectl run -i` stdin carries (the 097 script's own cap is 65536), then
that `RESUBMIT_CORR` was broken. Both were wrong; every attempt was queued and every
one eventually ran, so a single review cost four council runs' credits. **All four
hypotheses were consistent with the evidence I had, because "no rows yet" is consistent
with everything.**

**The check that settles it in one query** — ask when *other* orchestrations started,
not whether yours has:
```sql
SELECT orchestration_id, min(changed_at) AS started, max(changed_at) AS last
FROM orchestration_state_audit WHERE changed_at > now() - interval '45 minutes'
GROUP BY 1 ORDER BY started DESC LIMIT 10;
```
If recent rows exist but every `started` is minutes after its dispatch, the pipe is
moving and you are in a queue — wait. If nothing has started platform-wide, suspect
the dispatch path. Your own run appearing in that list *is* the answer: mine was
sitting there, 16 minutes late, while I was busy proving it had never arrived.

> **SUPERSEDED IN PART, 2026-07-20 — read `bugs_open/030` first.** This entry was written
> from one thread's observation (~16 min). `030` had already diagnosed the same thing the
> day before, with the ROOT CAUSE (`system.agent.generic.requests` has PartitionCount 1
> and a single consumer, so every dispatch queues in order behind every other), better
> measurements (**25–36 min** end-to-end), and a decisive one-command test this entry
> lacks:
> ```
> kubectl -n kafka exec <kafka-pod> -- /opt/kafka/bin/kafka-consumer-groups.sh \
>   --bootstrap-server localhost:9092 --describe --group generic-requests-group
> ```
> A non-zero LAG answers "queued or dropped" in seconds. Measured 181 on 2026-07-20.
> **Run that before any of the reasoning below.** The author of this entry then made the
> same mistake a second time the next day — at 31 minutes, inside `030`'s stated range —
> having never run the lag command (`WRONG_CALLS.md` 2026-07-20).

**Rules.** Before resubmitting ANY credit-spending dispatch: (0) run the lag command
above; (1) confirm the run is absent from the platform-wide start list, not merely absent
from your own polling window; (2) baseline the latency when the system is quiet so you know what
"late" means; (3) never ship a fix for a transport hypothesis you have not tested —
a resubmission is not a free retry, it is another full council. Kin: "Trust the
rendered artefact, not the status" (§ durable invariants) — the same error, applied
to a queue instead of an artifact. Category tags: `queue-latency`,
`false-drop-diagnosis`, `wasted-credits`.

### "A page was rewritten" does not say what rewrote it — attribute by the emitting row, not by the damage (2026-07-19)

**Symptom.** A page you had hand-corrected comes back rewritten, with fabrications
restored and a link to a page that does not exist. There is an open bug whose
description fits the damage exactly — *"re-planning a site silently discards its built
pages' composition"* — so the damage gets appended to that bug as fresh evidence, and
its severity is raised on the strength of it.

**What it actually is.** On this platform **at least four independent paths can rewrite
a live page**: `reconcile_site_plan` (a re-plan), `page-rerender`, `content_rewrite`
(from `discovery`, `content-gap-planner` **or** `tool-suggester`), and the tool
pipeline. They produce near-identical damage — regenerated components, new copy, new
links — and the damage carries no signature of its source. Two recurrences on
leopardessconsulting.co.uk were filed under the re-plan bug (`/bugs_open/001`). Neither
was: `reconcile_site_plan` has **never** emitted a work item on that site. One was a
`page-rerender` run, the other a `tool-suggester` `content_rewrite`
(`/bugs_open/029`).

**The check that settles it in one query** — ask what has *ever* emitted work for the
site, before believing any attribution:
```sql
SELECT source, item_type, count(*), min(created_at), max(created_at)
FROM site_work_items WHERE site_id = '<site>'
GROUP BY 1,2 ORDER BY 4 DESC;
```
If the source you are blaming is not in that list, it did not do it — full stop. Then
pin the specific event by time: find the `page_components.updated_at` cluster for the
damaged page, and read `orchestration_states.initial_request_data` in a window around
it. That request payload names the agent and carries the spec that drove the write, so
it identifies the culprit outright — `{"page":"services","source":"tool-suggester",
"suggestion":"Weave a natural reference to 'Monitoring Coverage Gap Finder'…"}` is not
something a re-plan could ever produce.

**Why this one is expensive.** A misattribution does not fail loudly — it *strengthens*
the wrong bug. The evidence reads as corroboration ("not dormant, not confined to one
site, hit twice in 24 hours"), which raises the wrong fix's priority, and the fix then
ships and appears not to work because the real emitter is untouched. Worse, the true
mechanism is left unfiled: the leopardess damage was autonomous and recurring, while the
bug it was filed under needs a human to deliberately fire a re-plan. **The urgency
belonged to the mechanism, and it was transferred to the wrong one.**

**Rules.** (1) Symptom similarity is not attribution — a shared *effect* means nothing
when several paths share it. (2) Before appending evidence to an existing bug, confirm
the accused path actually ran **on that site**, by row, not by plausibility. (3) When
several subsystems can produce one symptom, enumerate them *first* and eliminate, rather
than matching the symptom to the bug you already know about. Kin: "A fix applied to one
branch of a two-branch router reads as done" (§9) — both are the failure to enumerate
the paths before believing one. Category tags: `misattribution`, `shared-symptom`,
`multi-writer`, `evidence-by-source-row`.

### A one-source check of a two-source authority under-reports, and looks clean doing it

*Added 2026-07-19 from bugs_open/002 error C.*

`load_page_sections_from_spec` resolves a page's section list from **table if
present, else aspect**, then syncs the winner down over `pages.sections`. A drift
sweep that compares only `site_plan_sections` therefore misses every
aspect-authoritative site — and some sites have a current `site_plans` row with
**zero** `site_plan_sections` rows, so source 1 misses entirely. My table-only
sweep returned "1 drifted page out of 91" and read as near-clean; the corrected
two-source sweep found the real set, including two pages whose next rebuild would
have deleted live editorial copy on a client site.

Two compounding traps in the same query:

- **JSON `null` where you assumed an array.** `jsonb_array_elements` on a scalar
  raises `cannot extract elements from a scalar`; and if you paper over that, the
  NULL it produces gets silently removed by a later `WHERE … IS NOT NULL`. The
  page then disappears from the report rather than erroring. Guard
  `jsonb_typeof(x)='array'` on **both** sides.
- **The `null` case is genuinely safe, so you cannot just treat it as drift.**
  Source 2's type assertion fails and it falls through to source 3. Of 16 pages
  carrying the section, only 2 were actually at risk. A check that flags all 16
  gets ignored as noise; one that flags 0 misses the 2.

**Rules.** (1) Before writing a fleet sweep, read the resolver and enumerate
*every* source it consults, in order — a sweep is only as wide as the authority
model it encodes. (2) A near-zero result from a sweep you just wrote is a reason
to test the sweep, not to relax. (3) Prove the sweep can *see* a known-bad row
before trusting a clean run. Kin: "A fix applied to one branch of a two-branch
router reads as done" (§9). Category tags: `partial-coverage`, `authority-order`,
`jsonb-null`, `false-clean`.

### An unbusted GET is not the rendered artefact — it is a cache's opinion of it

*Added 2026-07-20 from bugs_open/002 D (see `WRONG_CALLS.md` for the full row).*

The standing rule here is **"trust the rendered artefact, not the status"** — it
has caught real damage and it stays. This is its missing corollary: **only if you
have proven the artefact is the current one.**

A thread fetched a live page, found a broken empty section, and reconstructed a
whole mechanism to explain how it survived (removed from the DB, orphaned in the
deployed artifact) — one step from forcing a production re-deploy of a live
client page. The section was already gone. The fetch had hit a **stale edge
cache**: origin updated at 12:46, `cache-control: max-age=3600` still serving the
pre-update copy. Stale 52,624 bytes; real page 44,880.

**The tell was already on screen and skimmed past:** `last-modified` was two
hours *before* the fetch and two days *after* `pages.deployed_at`. A page whose
file is newer than its recorded deploy, serving content older than its database,
is a caching question, not a mechanism question.

**The diagnostic inversion that matters.** Four independent DB checks said the
section did not exist; the page said it did. The thread kept inventing mechanisms
to explain the page. **When the artefact contradicts the database, suspect the
fetch before you invent a mechanism** — the database is one system, the fetch
crosses three (origin, CDN, whatever your client caches). The elaborate
explanation is the smell: if reconciling a page with its own DB needs a novel
mechanism nobody has seen before, test the cheap boring cause first.

```bash
curl -s -H 'Cache-Control: no-cache' "<url>?cb=$(date +%s)" -o /tmp/a -w '%{size_download}\n'
curl -sI "<url>" | grep -iE 'last-modified|cf-cache-status|age'
# compare size against an unbusted fetch; compare last-modified against pages.deployed_at
```

Related trap in the same family: verifying a fix by grepping a **generic** CSS
property always passes (travelling-docs). Both are "I checked the page" where the
check could not have failed. Category tags: `stale-cache`, `artefact-not-current`,
`verification-that-cannot-fail`.

### "The fix needs X to be built" is a claim about the codebase — grep before repeating it

*Added 2026-07-20 from bugs_open/002 D; the error was made independently by two
threads, which is why it earns a pattern.*

A 2026-07-15 handoff said an empty section on a content-rich page needed "a
TARGETED single-section repair (fill only the empty fields, no whole-page
re-save)" to be designed. The signing-off thread (2026-07-20) repeated it as
"no targeted single-section repair path exists — an architectural gap". The
owner asked for a double-check. **The mechanism existed the whole time**:
`apply_section_edit` / the `section-editor` agent
(`section_editor_actions.go`, added **2026-02-19**) edits one
`page_components` row's `content_data`, re-renders that component, reassembles
via `assemblePage` — and never passes through `save_page_sections`, so the
content-regression guard is not in its path. Seeded, active, in the deployed
binary, trigger script present, **3 COMPLETED production runs**. Five months
old at the time it was declared missing. What was actually missing was
narrower: nothing *generates* the content (the editor applies caller-authored
fields), and nothing *routes* `empty_section` items to it.

**Why two threads missed it:** a "fix candidates" section reads as design
guidance, so nobody treats it as an assertion needing evidence. But "X does
not exist" is exactly as checkable as "X is broken", and cheaper: one grep of
`registry.go` (`grep -nE '"[a-z_]*(section|component|repair)[a-z_]*":'`) put
`apply_section_edit` on screen. The check for "does machinery exist" is:
registry → agent_definitions row (active? workflow?) → pod binary (`strings`)
→ `orchestration_states` (ever run?). Four queries, two minutes.

**Rules.** (1) Before repeating a handoff's "needs X built", grep for X — the
platform convention "reuse existing machinery before building new" applies to
*claims about* machinery too. (2) When signing off someone else's entry,
their fix-candidate list inherits none of the evidence standard of their
symptom report — re-ground it. (3) Scope the surviving claim precisely: the
guard-blocks-whole-page-regen half was true, but "can't be repaired at all"
had silently widened it. Kin: the CLAUDE.md 2026-07-19 correction ("the
failure mode is not missing information — it is not looking"). Category tags:
`asserted-absence`, `fix-candidate-unverified`, `reuse-before-build`.

**The detection gap — why NOTHING caught it (added 2026-07-20, same day).**
The thread-side rules above fix the human error. The sharper finding is that
every automated layer that could have caught it is blind to this class, by
design:

- **Discovery checks detect broken artifacts, not unused capabilities.** Every
  check walks deployed pages/components asking "is this wrong?". None asks
  "does a registered mechanism exist that nothing routes to?". Dormant
  machinery produces no artifact to inspect, so the sweep can never see it.
- **The work-item loop's failure path never consults the capability
  inventory.** The `empty_section` item for gripper-cycle-time-estimator
  burned two attempts into the regression guard and went `unresolved`
  (2026-07-17) while the section-editor — the registered mechanism for
  exactly that repair shape — had sat idle since **2026-07-10**. The retry
  logic re-runs the SAME handler; nothing at exhaustion asks "is there
  another registered path for this item's shape?".
- **The council reviews the plan against its rationale — and takes the
  rationale's factual claims on trust.** A submission saying "no targeted
  repair path exists, so build one" would be judged on edit quality, blast
  radius, guardian concerns… by seats that all inherit the false premise. No
  seat's responsibility is *verifying asserted absences*. (Same shape as
  bugs_open/031, where a seat *quoted* a false register entry as contract:
  the council's inputs are only as true as what someone wrote down.)
- **Handoffs give fix-candidate lists no evidence standard**, so an absence
  asserted once propagates by citation — this one survived a handoff, a
  re-grounding pass, and a sign-off, and fell only to a direct challenge.

The class, named for filing and for the council (see the seat proposal in the
§9 preamble of the council-gate runbook or below): **asserted-absence /
dormant-machinery** — one bug, two faces. *Asserted-absence*: a durable claim
that capability X does not exist, made without an existence check, driving a
plan to build X or a gap to be declared unfixable. *Dormant-machinery*: the
complementary platform state that makes the false claim easy — X exists, is
deployed, may even be production-proven, but nothing routes work to it and no
inventory surfaces it, so knowledge of X decays into folklore at the pace of
session turnover. The existence check that collapses both: registry grep →
`agent_definitions` (active? workflow?) → pod binary (`strings`) →
`orchestration_states` (ever ran?). Four lookups, two minutes, and it is
mechanical enough for a council seat or a discovery check to run verbatim.
Category tags: `dormant-machinery`, `capability-inventory`,
`absence-unverified-by-council`.

### A section name is not a component name — `section_type` is an invisible alias

*Added 2026-07-19 from bugs_open/002 error C.*

Section→component resolution is exact-match, but against **three** columns in
order: `name`, then `function`, then `section_type`
(`v3_site_actions.go:3383-3465`, `component_selector.go:164-190`). `section_type`
is written from the *requested* section name, while the LLM names the row
whatever it likes (`store_generated_component_action.go:636-645`), and nothing
reconciles the two. So a plan asking for `hero-contact` binds a component
actually named `contact-hero` — permanently, invisibly, and correctly.

The consequence for debugging: **a name mismatch between a plan and
`page_components` is not by itself a defect.** It may be an alias that resolves
fine. Conversely a name that *matches* something real is not safe either — when
the plan named `contact-info`, Pass 1 exact-matched the generic library component
of that name and the bespoke `contact-block` was never a candidate. Determine
which pass will win before concluding either way. Category tags:
`alias-by-side-channel`, `resolution-order`, `name-is-not-identity`.

### A false positive is a LOCATION, not a dismissal

*Added 2026-07-19 from bugs_open/026, 027 (relojistas).*

`check_empty_sections` flagged `news-listing` on a page that demonstrably renders
twenty items. That finding is genuinely wrong — the section is a runtime-fill
template, the data arrives client-side, the check only ever sees server HTML.
The correct verdict is "false positive". **Dismissing it there would have cost
two real fleet-wide defects**, both sitting inside the very component the checker
pointed at:

- the shared template hardcodes English `"Loading latest news..."` — visible on a
  Spanish site, and permanently visible to any client that does not run JS (026);
- its `headline` is declared `required` in `input_schema` and rendered **empty**,
  so something that should have refused the save didn't (026);
- and the reason the check "wrongly" sees nothing is itself the defect: news
  exists **only** client-side, so every news page on the platform serves zero
  news to a non-JS consumer (027).

The generalisable move: when a check fires and you believe it is wrong, **read
the artefact it named before closing it**. The checker's *reasoning* was wrong;
its *aim* was perfect. A check pointed at a defective surface will often
misdiagnose that surface — the wrongness of the verdict says nothing about the
health of the thing it is looking at.

Corollary, and the cheaper habit: a false positive that recurs is a standing
invitation to stop reading a component that nobody else is reading either.
Category tags: `false-positive-is-a-pointer`, `runtime-fill`,
`server-vs-rendered`, `dismissal-cost`.

### A field the backend has no parameter for gets accepted and discarded — and the comment saying so is why it survives (2026-07-20)

*Added 2026-07-20 from `bugs_closed/028` (Banana discards every negative prompt). CLOSED same day: fixed, live in v1.0.1140, and proven end-to-end — the fold reached the model at 517→905 chars. The pattern below is what generalises, not the instance.*

`provider.Request.NegativePrompt` crossed three layers — style guide → action →
Kafka → adapter → provider — and died in the Banana provider at `logger.Debug`,
because Gemini has no negative-prompt parameter. Every `avoid` list in every
site's `imagery_style_guide` was therefore inert fleet-wide, along with every
`kindDefaults.NegativePrompt`. Nothing errored. The field was still set, still
validated, still shipped; it just reached nothing.

**The general shape: a field is only as live as its LAST reader.** A struct
field that crosses a boundary into a backend with no equivalent concept is a
void unless someone translated it. Grep the field to its terminus before
believing it does anything — `avoid` had exactly three references in the whole
tree, and the third was a log line.

Four things made this survive, each transferable:

- **The drop was documented as intent.** The file header said *"callers
  shouldn't rely on it being honoured"*. That reads as a considered decision, so
  nobody questions it. A comment explaining why data loss is fine is a defect
  smell, not a mitigation.
- **The interface contract licensed it.** `provider.Request` told implementers
  that providers without support "log and ignore". Banana was *obeying the
  contract*. Fixing only the call site would leave the next provider author the
  same licence — when a defect was permitted by a contract, fix the contract.
- **`Debug` is `/dev/null`.** A discard nobody can see is a discard nobody
  finds. If a layer drops a caller's constraint, that is Warn at minimum.
- **The routing change that caused it was correct.** Every kind moved to Banana
  for good reasons (bugs_open/011). The defect was its unnoticed *consequence*.
  When you migrate a path to a new backend, enumerate what the old backend
  honoured that the new one does not — the migration diff will not show you a
  field that simply stops being read.

**The expensive half is an attribution trap, and it is the part worth
remembering.** A style-guide `avoid` edit was made, the next generation came out
better, and "ground colour is fixed via `avoid`, not `medium`" was written up in
three documents as a hard-won fact. The edited field was never read. What
happened was a re-roll that came out darker, and the coincident edit took the
credit. **With a nondeterministic backend, one before/after observation cannot
distinguish a working fix from a lucky sample** — and a config edit costs
nothing, so it is exactly the kind of change people "verify" at n=1. Before
recording *how the system works* from an output change, confirm the thing you
changed is read at all.

Related signature, useful on sight: of nine images, four violated the avoid list
and five complied. **That ratio is what an IGNORED constraint looks like** —
compliance by luck. A partly-working constraint and a wholly-ignored one are
indistinguishable at small n, so count violations across a set rather than
eyeballing one, and treat "mostly fine" as unproven rather than as evidence.

Category tags: `field-only-as-live-as-its-last-reader`, `silent-discard`,
`documented-as-intent`, `contract-licensed-the-bug`, `debug-is-devnull`,
`coincident-edit-takes-the-credit`, `compliance-by-luck`.

### "No error anywhere" usually means no error *surface*, not no error

*Added 2026-07-20 from bugs_open/002 F → bugs_open/034.*

A thread reported a Kafka envelope "accepted, never executed, no error anywhere"
and left it unexplained for two days. There *was* an error — it just had nowhere
durable to land. `agent.go:828-845` classifies a failed message by **substring**
(`"is required"`, `"validation"`, `"invalid"`) and on a match returns early,
skipping `handleProcessingError`. That skips the error response to the waiting
parent, the retry, **and any DB write**. The residue is one `zap.Warn` on a pod
that rotates ~3.6k lines/10min, plus a Prometheus counter labelled
`(agent_type, reason)` with no correlation_id. Everyone here investigates via the
database, so the message reads as though it never existed.

**The generalisable point:** before concluding "it vanished silently", establish
whether the failure mode you are hunting has a durable surface *at all*. Absence
of a row is evidence only if something was supposed to write one. Ask: which
table would record this, and does the code path actually reach the write? A
`return` above the recording call makes every downstream investigation a search
through an empty room.

**Substring error classification is the underlying smell.** `strings.Contains(err,
"invalid")` is unanchored: it catches `invalid character 'w' after object
key:value pair` (a truncated-LLM parse failure), `invalid memory address`,
`invalid connection`, `strconv … invalid syntax`, `x509: … invalid`. A branch
meant for malformed envelopes silently eats driver errors, nil derefs and TLS
failures. Prefer a typed sentinel (`errors.Is`). Kin: bugs_open/017, the same
class already fixed at one narrow site (`c80fffc83`) — when you find one
`Warn`-only branch, grep for its siblings rather than fixing the one in front of
you. Category tags: `no-durable-surface`, `substring-classification`,
`silent-drop`, `absence-is-not-evidence`.

### A hard cap that silently discards its input's tail rewrites meaning — and the tail is whatever was composed LAST (2026-07-20)

*Added 2026-07-20 from `bugs_open/027` §4b (imagery style-guide palette truncated
away before generation); the council's bug_historian seat asked for the
transferable pattern (round 6, correlation 0a07f5ed).*

`composeDirection` joins medium → mood → **palette last**;
`composeImagePromptWithDirection` truncates the result at
`maxImageryDirectionInPrompt = 200` — unconditional, unreported. The most
brand-identifying field silently never reached the model on 5 of the fleet's 8
palette-carrying directions; the model invented an accent per image; the output
looked deliberate. The testbed that PASSED its owner gate was over the cap too —
its cut just happened to land after the accent hex. A false "hard-won fact"
(`avoid` fixes ground drift) grew in the gap and reached three documents.

**Two generalisations:**
- **A silent cap converts "too long" into "means something else".** If a budget
  must drop content it must (a) drop by stated priority — compose so the
  load-bearing field cannot be the tail — and (b) say so (WARN naming the loss).
  "Fits, or silently shrinks" is how a passed gate and a false lesson coexisted.
- **A cap is a claim about its CONSUMER; re-check it when the consumer changes.**
  The constant's own comment stated its precondition — "until provider routing
  lands... the only generation backend (SDXL)". Routing landed (v1.0.1139, every
  declared kind → Banana, ~1000+ chars tolerated) and nothing flagged the expiry.
  When you swap the thing a limit protects, grep limit constants for their
  justifying comments.

Sibling shape, found the same weekend: `016b` §9 "a field is only as live as its
LAST reader" (`bugs_closed/028`). One discards the tail of what it keeps, the
other discards a whole field — both silently, both leaving output that reads as
intended. Fleet enumeration + fix candidates: `bugs_open/027` §4b.

### A cap on a READ path reports a backlog as "nothing to do" (2026-07-20)

**Symptom.** 303 work items sat at `needs_human_review`, oldest four months, none
ever actioned. Filed as *"the human-review queue has no working surface"*
(`bugs_open/033`) — the reasonable inference from four months of total silence.

**Mechanism.** The surface exists and is complete (approve / retry / resolve /
save-and-rebuild / editable review forms, `frontends/admin-dashboard/src/App.tsx:397,1160-1189`).
`HandleListWorkItems` hardcodes `limit := 50` `ORDER BY created_at DESC`
(`site_admin_handlers.go:483,519`); the frontend passes **no** `status` param and
filters client-side, under a comment asserting the opposite — `// Load all
non-complete items, filter client-side for accurate counts` (`App.tsx:440`).
Against live data the cross-site view could show **0 of 208** review items, and
the "Needs Review (N)" badge — computed over the same 50-row window — read **0**.

**The generalisation, and why it is not the truncation shape above.** A silent cap
on a *write* path corrupts a stored artifact: the damage is visible in the row,
and someone eventually reads it and frowns. A silent cap on a *read* path
corrupts the **absence of evidence**, and absence is what people reason from when
deciding whether a problem exists. It does not produce a wrong page; it produces a
confident, well-founded, entirely false belief that there is no work — held by
every human and every thread that looks. Four months of "nobody cares about this
queue" was four months of the dashboard truthfully reporting what it could see.

**Two checks that generalise:**
- **When a UI filters client-side, the cap and the filter are one bug.** `LIMIT n`
  is safe only if the client can also ask the server to narrow. Here the backend
  already supported `?status=` and `?item_type=` (`:497,:513`) — the frontend
  simply never used them, so the cap it did not know about silently pre-empted the
  filter it did use. Grep the caller before trusting a filter's counts.
- **"Nobody uses X" needs a search wide enough to be falsifiable.** 033's claim
  that the three admin routes had never run came from a **backend-only** grep; the
  callers were in `.tsx`. Before asserting a cross-cutting negative, name the
  languages and directories the grep covered — if that list is narrower than the
  claim, the claim is about your search, not the system. (`WRONG_CALLS.md`,
  2026-07-20; third entry of this shape that week.)

**Companion trap in the same file.** The designed HITL path (`checkpoint_for_review`
→ `spec.checkpoint` → **Approve & Continue** → `HandleApproveWorkItem`) is fully
implemented at every stage and has never run: **0 of 5,622 items in the platform's
history have ever carried `spec.checkpoint`**, and the action is wired to 0 live
agents. A button whose render condition has never been true is indistinguishable,
from the outside, from a button nobody clicks — and the two have opposite fixes.

### A recorded user decision with no enforcement point is decorative — an LLM re-guesses it every run (2026-07-20)

**Symptom.** robot-hands runs `tool-portal-dark` (`layouts.scheme='dark'`, a
user decision taken twice); a routine webdesign run shipped
`--color-background: #F4F5F7` — light — onto it. No error, no warning
(`bugs_closed/022`).

**Mechanism.** The palette merge's core-slot rule ("spec wins") is right for
site identity and wrong for *scheme*: `analyze_design` emits a fresh
`color_scheme` on every run, and **nothing anywhere compared the proposed
background against the layout's declared scheme**. The decision was recorded
in the DB but had no enforcement point, so the per-run LLM guess silently
overrode it — every run was a fresh roll of the dice.

**The transferable test.** When config records a user decision (a scheme, a
lock, a pin) and any pipeline stage merges generated output over related
fields, grep for the code that COMPARES them. If no comparison exists, the
decision is decorative regardless of how prominently it is stored. Prompt-side
mitigation (design_intent pins) only lowers the trigger rate — the merge seam
is where authority must live.

**Fix shape that survived five council rounds** (corr `0328ddc7`): guard at
the merge boundary's single verified call site; restore paired slots
TOGETHER (background+text — a half-swap breaks contrast; round 1 caught my
draft doing exactly that); **hard-fail what you cannot repair** — a Warn
nobody is paged on is the original silent failure re-shaped (round 2); and a
resubmission is judged standalone, so carry all standing evidence every
round (round 4, now in `RUNBOOK_council_gate.md` traps).

**Cross-refs.** `bugs_closed/022` (full trail + live verification),
`enforceLayoutScheme` in `render_css_from_spec_action.go` (`9c3b0c3e7`,
live v1.0.1140). Category tags: `authority-boundary`, `merge-seam`,
`llm-overrides-user-decision`, `fail-loud-not-silent`.

### A renderer fix is inert until something re-renders — and nothing schedules a re-render (2026-07-20)

**Symptom.** A defect you can see on a live site is *already fixed* in the code, and has been for
weeks. Grepping the repo shows the fix, the pod carries it, and the site still shows the bug.

**Mechanism.** The artefact is written once and cached in a table (`site_components.rendered_html`,
`page_components.rendered_html`, `pages.rendered_*`). The renderer runs only when something
explicitly triggers it. If nothing sweeps that surface, a corrected renderer simply never runs for
sites that don't happen to get rebuilt — so "committed and live" says nothing about what is served.
`bugs_open/049`: the chrome renderer's hardcoded `{/privacy.html, /terms.html}` legal slice was
fixed 2026-06-10 (`0681e1542`); three sites' footers date from 28 Apr / 21 May and **still served the
phantom links 41 days later**, on every page.

**The diagnostic, and it is cheap.** Compare the **artefact's timestamp** against the **commit date
of the fix**. Two columns and a `git log -S`:
```sql
SELECT s.domain, sc.slot_name, sc.updated_at FROM site_components sc JOIN sites s ON s.id=sc.site_id ORDER BY 3;
```
```bash
git log -S'<a literal from the fixed code>' --format='%h %ad %s' --date=short -- <file>
```
Any artefact older than the fix is a suspect, and the sites *newer* than the fix are your control —
if they are correct, the fix works and the only defect is that it has never run. That two-directional
check turns "is the fix good?" into a settled question in one query.

**Generalises to:** any cached render (chrome, page sections, CSS), any config-derived artefact, and
any "we fixed that weeks ago" claim about a live surface. Sibling of the *"Go changes are inert until
an image is rebuilt"* rule — one level further out: **the image rolled and the artefact still didn't.**

**Cross-refs.** §9 *"Light site renders dark chrome — … stale `site_components` chrome"* (same
underlying property — nothing refreshes `site_components` — different consequence: that entry is
about chrome pointing at *deactivated components*, this one about chrome predating a *code fix*).
`bugs_open/049`, `bugs_open/041`, `bugs_open/018`. Category tags: `stale-artefact`,
`fix-inert-until-rerender`, `artefact-timestamp-vs-commit-date`, `two-directional-control`.

### A `regexp_matches(…,'g')` census with a lookback prefix silently drops every other match (2026-07-20)

**Symptom.** A SQL census over template/HTML text returns a plausible number that is wrong by
roughly half, in exactly the places the thing you are counting clusters.

**Mechanism.** `regexp_matches(col, pattern, 'g')` returns **non-overlapping** matches. If the
pattern carries a context prefix — `'(.{0,60})<a[^>]*href="…"'`, the standard trick for "is there an
`{{if}}` before this anchor?" — that prefix is *part of the match* and consumes the preceding text.
In a run of adjacent matches (nav lists, footer link columns, repeated CTA blocks) each match eats
its predecessor, so about every second one is never counted. RUNBOOK R9 reported **70 ungated
anchors / 37 components**; a real parse found **171 / 41** — a 2.4× undercount, concentrated in the
components with the most anchors.

**Rule.** A lookback prefix makes a counting regex unsound. Use it to establish *direction*, never to
build a worklist. For a worklist, parse: tokenise the template, keep a block stack, and decide
membership structurally (`cta_link_integrity/scripts/parse_gates.py` is a 60-line worked example).
**And when a query carries its own "re-derive this properly before acting" warning, that warning is
the finding** — this one had been attached to the figure since the day it was written, through three
status updates that repeated the number.

**And the sequel, which is the more useful half.** The *corrected* figure (171) was then written
into four documents — and 15 of the 171 were not the thing being counted at all: they sit inside a
`{{range}}`, so the field belongs to the ranged item (`{{range .items}}<a href="{{.url}}">`), an
item link rather than a label/url pair. A corrected number inherits the authority of the correction
and stops being questioned. It surfaced only because a migration post-condition was written as a
blanket *"no ungated `{{.x_url}}` anchor remains"* and a range-scoped `.url` would have tripped it,
rolling back a correct change. **Two transferable rules:** (1) a blanket structural assertion in a
migration gate must be scoped to the exact needles you edited, or unrelated members of the same
syntactic class will veto you; (2) when a caveat has to be *remembered*, it will be forgotten —
put it in the tool's output. `parse_gates.py` now prints the range/CTA split on every run.

**Cross-refs.** `bugs_open/023` (correction + resizing), RUNBOOK R9/R9b, `WRONG_CALLS.md`,
migration `181`. Category tags: `measurement-artefact`, `regex-census`, `heuristic-vs-parse`,
`warning-was-the-finding`, `corrected-figures-stop-being-questioned`.

### An empty result overloaded to mean "nothing here" AND "ask elsewhere" fires the fallback on the truthful-empty case (2026-07-21)

**Symptom.** robot-hands.com's footer `.footer-legal` slot renders **14 links, none legal** — the
whole footer navigation rendered a second time in the legal position, three entries duplicated,
on every page of at least six live sites (`bugs_open/053`).

**Mechanism.** `GetNavItems` (`nav_tables.go`) queried the nav tables, and on **zero rows** fell
through to a pages-table fallback built for sites that predate those tables. But a zero-row result
is *two different facts*: "this site has no nav tables" (fall back) and "this nav-table site
genuinely has no items in this group" (the correct, expected answer — most sites have no legal
pages). The fallback cannot tell them apart, so it ran on both; for the `legal` group its footer
branch matches every `in_footer` page, and the `in_footer` disjunct dominates the legal-name
disjunct — nothing constrains the result *to* legal pages. A truthful empty answer was read as a
missing table.

**The transferable test.** When a lookup returns empty and the code *does something else* on empty
— a fallback source, a default, a broader query — ask whether empty is **overloaded**. If "no rows
for X" and "X is not configured here" produce the identical empty value, the branch that fires on
empty fires on the legitimate case too. The fix is to **disambiguate empty with a second, cheap
probe** that distinguishes the two states, not to tune the fallback query. Here: gate the fallback
on `SELECT EXISTS(SELECT 1 FROM site_nav_items WHERE site_id=$1)` — fall back only when the site
has **no** nav rows *in any group*; otherwise respect the empty answer. That fixes every group type
at once, where narrowing only the legal query would have papered over `legal` alone.

**Watch the error path in the probe.** The original empty could also come from the table *not
existing* (older deployments). The disambiguating probe must treat its own `does not exist` error
as "fall back", or the backward-compat branch it was protecting breaks. Test both the real-empty and
the table-absent cases (`nav_tables_fallback_test.go`, four sqlmock cases).

**Sibling trap — the fix is inert until a re-render.** The output lives in `site_components.
rendered_html`; the corrected `GetNavItems` changes nothing served until chrome re-renders. Same
property as the entry above — do not report it fixed on the strength of the code change.

**Cross-refs.** `bugs_open/053` (case + candidates 1/2/3; candidate 3 deferred to `bugs_open/052`
because the current `deployedOnly` flag emits the wrong `build_status='deployed'` predicate),
`bugs_open/049` (found while measuring its re-render). Commit `85d39f9b9`. Category tags:
`overloaded-empty`, `fallback-fires-on-legitimate-case`, `disambiguate-with-second-probe`,
`fix-inert-until-rerender`.

### A generic section name resolves to a product-specific component (the library has only one candidate)

**Symptom.** A page plan asks for a *generic* section (e.g. `hero-tool`) and gets a panel
full of another product's frozen vocabulary — on an LLM-cost page, "Start Ranking Free" /
"Try the Bayesian Ranker". Looks like the planner proposed the wrong thing. It did not:
`pages.sections` contains the literal string `"hero-tool"`; nothing about the other product
was ever planned (`bugs_open/045`, FIXED & LIVE 2026-07-21, DB config, no image roll).

**Diagnose.** A section name resolves through `component_selector.go` → `queryCandidates`,
which matches on **`section_type`** — NOT `function` or `name` (023 R2 is about `slot_name`↔
`function`; the *selector* keys on `section_type`):
```sql
SELECT function, name, section_type FROM content_components
WHERE section_type = '<the section name>' AND component_level='section'
  AND is_active=true AND forked_from IS NULL ORDER BY <score> DESC;   -- caller takes row 0
```
If that returns exactly ONE row and it is product-specific, every generic request inherits
it. Two amplifiers make it stick: (1) `SelectComponentByType` has **no minimum-score
threshold** — a sole candidate always wins, however low it scores; (2) the offending row's
labels are `source:static` with product fallbacks, which re-apply on every render and cannot
be overridden by `content_data` (`plan_sections_action.go` static branch) — so page-level
edits can't paper over it.

**Root cause.** A **library gap**, not a planner or selector bug: there is no neutral
component carrying that `section_type`, so selection correctly resolves the only match.

**Fix.** Add a generic component with that `section_type` (labels `source:llm` or genuinely
generic; CTA anchors gated `{{if .x_url}}` with `*_url` fields `source:renderer` — LNK-005 by
construction; optional gated stats with anti-fabrication guidance), THEN re-point the
specific row's `section_type` to its own function so it leaves the generic pool — kept
`is_active=true`, **never deleted** (it is the sole active row for its function; 023 R10). Do
both atomically so the pool is never empty or ambiguous. Precedent: migration 179
(`tool-guide-intro`) for the CTA shape; migration 183 for the whole fix. Sibling failure:
`bugs_open/039` — a section name resolving to **no** component renders a hollow stub (same
selector, opposite end).

**Verification landmine — a rerender does NOT prove a selection fix.**
`RerenderSinglePageAction` "assembles a page from stored/pre-rendered components" — it
re-renders existing `page_components`, it does **not** re-run `plan_sections` or re-select.
So re-opening a `page_rerender` work item on a page that has no placement for the section
will never create one, and proves nothing about selection. Only the full **site-build** path
(`get_pages_to_build_actions.go` → `plan_sections` → `SelectComponentByType`, per-site over
`planned`/`needs_rebuild`) re-selects. To confirm a component-selection fix at the artefact
level you need a real build, not a rerender — or mirror `queryCandidates` in SQL (it is a
verbatim copy of the Go query) and confirm the sole candidate, which is deterministic.

## 10. Open bug queue (`/bugs_open/`) — index

The repo-root `/bugs_open/` directory is the live queue of diagnosed-or-filed bugs
awaiting a fixing thread (it was `docs024_key_docs_latest/aaa_fails_to_mend/`
until 2026-07-17; ~23 documents still reference the old path). §9 above holds the
durable PATTERNS; the files below hold the case detail, evidence and fix
candidates. Read the file before acting — several are already fixed.

**Two directories, one index (split 2026-07-19).** Rows marked **`→ bugs_closed/`**
have moved to `/bugs_closed/`; everything else is still in `/bugs_open/`. The bar
for moving is **fixed AND live in prod** — a fix that is committed but inert until
the next image roll stays open. (This sentence used to name `008`, `012` and
`017`-unregistered as the examples; all three have since shipped and closed —
a named example of a transient state expires within days, so check the rows
below rather than the prose.) **Numbering is one sequence across both dirs and is
never reassigned**, so a stale `bugs_open/NNN` or `aaa_fails_to_mend/NNN` pointer
resolves by number in the other directory. **`016` and `017` are each used by two
different cases** — resolve a bare number by slug, never by the number alone.
See `/bugs_closed/README.md`.

| # | Bug | State |
|---|---|---|
| 001 | Re-planning a site silently discards its built pages' composition | **FIXED & PROVEN LIVE 2026-07-20** (v1.0.1138/1139): preservation set widened to adoption-locked OR `build_status='deployed'` + non-empty-gated composition snap-back + truncation must-keep; migration 173 live; 7 discriminating tests. Two re-plans on dartsonline: run 2 snapped `index` (LLM dropped `category-listing`, added `content-listing`) and `shipping-returns` (LLM added `faq`) back to realised — the SAME `index` that lost sections in run 1 as `needs_rebuild`, protected in run 2 as `deployed`, so the guard keys on status generically. **CLOSED 2026-07-20 → `/bugs_closed/`** — every residual now has an owner: `037` (needs_rebuild boundary), `038` (content regenerated), `039`, `040`, `035`, plus `051` (the "90-day adoption lock" premise was false — `adoption_locked` is a per-SITE first-plan flag, so Pass C2 can only fire on a site's first plan) and `050` (the Pass B emptiness residual — **and 001's prescription for it was UNSAFE**: for a deployed page `sections=[]` means "rendered by another subsystem", true of all 18 such pages fleet-wide, so "take the LLM's sections when realised is empty" would inject layouts onto tool/blog-index pages). Pages still invented, unfixed by design. Never council-reviewed — both rounds voided by `019`. **Its "FRESH EVIDENCE" (leopardess) section is MISATTRIBUTED — that damage is `029`/page-rerender, corrected in-file** |
| 002 | Errors surfaced but not fixed (multi-error handoff, route individually) | **SIGNED OFF 2026-07-20 — `→ bugs_closed/`.** A routing doc that routed. B was already fixed (`005`); C fixed via SQL 175+176 (two sites, fleet drift 0); D closed — the section had been gone since 07-10, 8 stale items closed; A→`003`, F→`034`; E is an owner decision. **Its two most confident entries were its two wrongest** (A's retracted root cause; D's "needs a targeted repair built" when the mechanism had shipped 5 months earlier) — see §9 `asserted-absence` and two `WRONG_CALLS.md` rows |
| 003 | Spawned children lose their response; parents hang until reaped. **§3d (2026-07-18): second root cause — the consume loop commits Kafka offsets BEFORE processing (at-most-once), so any restart destroys in-flight work; §4.4 is the at-least-once + rollout fix that unlocks CD** | open |
| 004 | Landing an image can silently blank an article body | superseded by 005 — **`→ bugs_closed/`** |
| 005 | Article-body blanking — root cause LLM truncation (`max_tokens`) | FIXED; re-verified live 2026-07-19 (19/19 healthy, `max_tokens` 8000 survived a re-seed, repair fn in the running pod, zero writer truncation since 07-15) — **`→ bugs_closed/`** |
| 006 | Three idea.uk infra errors (runner cgroup, dead contact endpoint, …) | open |
| 007 | Applied-but-unrecorded migrations block the runner | instance resolved; tooling open |
| 008 | `GenerateText` never decodes `stop_reason` — a truncation returned as success, a refusal as "no text content in response (had 1 blocks)" | **CLOSED 2026-07-20 → `/bugs_closed/`** — all 5 items live in the 18:58 BST image, pod-verified (`model declined to answer` → 1). Items 1–4 in `f32b208e5` (both providers, plus `TruncatedError` carrying the partial — the transport half of `019`); item 5 + the provider-parity CI guard in `45e90acbb` |
| 009 | Root `ai_service` SHADOWS the step block (dead per-step config) | diagnosed; fix + fleet sweep open |
| 010 | Fix loop non-convergent on layout-intrinsic overflow | candidate (a) SHIPPED v1.0.1135; (b) open |
| 011 | `kind:"hero"` routes to SDXL (cannot render text); the Gemini infographic lane works and was unused | open |
| 012 | tool-improver truncates a component and saves the wreckage | **CLOSED 2026-07-20 → `/bugs_closed/`** — guard live in v1.0.1139, migrations 168/169/**170**; whole chain driven against prod (component untouched · refusal logged · item `needs_human_review` · note written). 170 was found BY that test: 169 put `error_step` top-level, where the workflow plan drops it |
| 013 | fix-implementer commits un-`gofmt`'d LLM output; build gate rejects it, no PR | **CLOSED 2026-07-20 → `/bugs_closed/`** — `formatGeneratedGo` runs `go/format` at commit-prep (`fc38c6058`), pod-verified. Unparseable bodies still fail LOUD there rather than falling back to raw bytes — usually a `max_tokens` truncation, so the message names it |
| 032 | The completion verifier reads a DELETED component as a successful fix — absence is equally a rebuild silently dropping it | **CLOSED 2026-07-20 → `/bugs_closed/`** — returns an error, not a verdict, so the gate's fail-OPEN policy turns a false success into a visible unknown (`a467baa11`), pod-verified. **Closed on its safe floor:** treating absence as *deletion* when the page still expects the component is the better verdict and stays open for the `empty_sections_loop_integrity` thread. Coverage half remains in `021` |
| 014 | VM-site artefacts silently deploy to the default `sites` repo (two causes) | FIXED (v1.0.1126 + pin removal) — **`→ bugs_closed/`** |
| 015 | Mistyped `page_type` orphans a page from every gate that keys on it | worked around per-site; planner fix open |
| 016 | `ssh` ignores `$HOME` (uses passwd entry) — service-account git-over-ssh fails twice over | FIXED in the box scripts — **`→ bugs_closed/`** (note: a *different* case also numbered 016 — council revise — is ALSO now in `bugs_closed/`, closed 2026-07-21; resolve by slug) |
| 017 | Static cutover orphans a backend tool's entry forms — funnel unreachable, no auditor models it | open; needs site fix + new check |
| 018 | idea.uk chrome renders with every link `href=""` (31/33) — site unnavigable; check fleet | open, unstarted |
| 017 | `fix_forced_text_colors` never registered ("requires a topic" lie); failed saga stamped 'complete' | **CLOSED 2026-07-20 → `/bugs_closed/`** — live in v1.0.1139, pod-verified with discriminating strings; both legs + parity test + dead-map deletion; 54 rows corrected; sweep 0 |
| 019 | One truncated reviewer (`output_tokens==max_tokens`) voids a whole council round, discarding every other seat's review | **CLOSED 2026-07-20 → `/bugs_closed/`** — verified by live reproduction on v1.0.1140 (scratch council, seat capped at 200): one `TOLERATED`-prefixed forensic row, 9 further readable seats, `complete_revise` with `unreadable: 1` in the report — vs the old zero-review `complete_invalid`. Corrected mechanism: 9 of 11 voids died UPSTREAM at `execute_llm_prompt`, not at the decider. See `036` for the sibling cause on the same seam (fix in the same image, verification pending) |
| 020 | Tool-recreation invents a dataset when the original tool was data-backed; shipped fake practices live, all items `complete` | filed; fix candidates in 020 |
| 021 | The 012 completeness guard covers ONE write path; `page_components.rendered_html` and `pages.rendered_*` have the same unguarded overwrite shape | filed (council bug_historian objection); needs scope decision |
| 022 | Nothing compares a spec's proposed `background` against `layouts.scheme` — a per-run LLM guess silently overrides the user's recorded scheme decision (robot-hands: light `#F4F5F7` onto scheme=dark, shipped) | **CLOSED 2026-07-20 → `/bugs_closed/`** — `enforceLayoutScheme` guards the merge seam (restore background+text together; hard-fail what the theme can't repair), `9c3b0c3e7`, live v1.0.1140 pod-verified (symbol + refusal string); pass-through path proven by a real pin-removed run (CSS stayed dark); reject path unit-proven, live-fire [UNEXERCISED] (two dispatches vanished — `003`-class, noted there). See §9 `authority-boundary` |
| 023 | A button's label and its destination are unrelated schema fields — nothing checks that a control with text has somewhere to go. Filed from four owner-reported buttons; 51 dead controls / 7 of 11 sites; 84% of library CTA anchors ungated | **SYMPTOM FIXED & VERIFIED LIVE 2026-07-20**, bug stays OPEN on its structural scope. Done: all placements of both bad components removed fleet-wide (3 sites), migration 179 fixed `tool-guide-intro` (gated anchors, renderer-owned urls, dead `#guide-start` gone) — census 51→39, **fragment class extinct (4→0)**, held through the v1.0.1140 roll. Remaining scope = classes A/B/C/E: **70 ungated anchors / 37 components**, **22 `source:llm` url fields**, no build-time pairing check. **Rescoped 2026-07-20: class F → `045`, class G → `033`, class H → council trail `2525f980` (observe stage live in v1.0.1140).** Verify-criteria rewritten — it no longer waits on another bug's work |
| 024 | A tool-improver fix is written durably to `content_components.html_template` and **never rendered to the page**. Six defects were diagnosed on the **generic** rerender path, but the real blocker was that that path is **deliberately forbidden for tool pages** (`save_page_sections` ownership guard, migration 164 — every tool page is `rebuild_policy='owned'`). Sanctioned delivery is the `section-editor` (`apply_section_edit`), which delivered the benchmark fix LIVE. Explains `bugs_open/010`'s "non-convergence" — the page never changed | **CLOSED 2026-07-21** (`bugs_closed/024`, fixed & live). Remaining work → `features_open/009` (Option A rewire; generic-path residuals `cdd858402` + candidate 4) |
| 031 | A wrong entry in the concept register claimed scoped page-rerender "SKIPS pages whose content hash is unchanged". No such code exists and `git log -S` shows it never did — but a council seat quoted it as "the pipeline's own contract" and blocked a correct plan at HIGH severity. Replication was wider than filed: 6 register-file occurrences **plus the live seat prompts** in both `fix-proposer` and `council-gate` rows | **FIXED & LIVE 2026-07-20** — register + sources corrected, live rows patched (`PATCH_render_guardian_031` + 099 sync, verified), citation convention added to docs026 README — **`→ bugs_closed/`** |
| 029 | `tool-suggester` emits `content_rewrite` items at SUGGESTION time telling the writer to "weave a natural reference" to a tool that has no `pages` row — the writer invents the URL, producing an owner-visible 404. Autonomous, and it regenerates human-reviewed copy while doing it. **This is the true cause of the leopardess damage wrongly filed under `001`** | filed 2026-07-19; primary DB evidence, no fix started; check `023` first — likely the same family |
| 034 *(collision RESOLVED 2026-07-20: the second `034`, slug `replan_rebuilds_every_deployed_page…`, was renumbered to `038` by its own author while both were minutes old — this number is now unambiguous)* | A failed message is classified by **substring** (`"is required"`/`"validation"`/`"invalid"`) and, on a match, returns before `handleProcessingError` — skipping the error response to the waiting parent, the retry, **and any DB write**. Residue is one `zap.Warn` on a pod that rotates in minutes + a counter with no correlation_id. The match is unanchored, so it also eats driver errors, nil derefs and truncated-LLM parse failures. Explains `002` F.2's "accepted, never executed, no error anywhere" (`client_id is required` returns before `getOrCreateState` → zero rows) | filed 2026-07-20; mechanism proven from code, application to F.2's two correlations is hypothesis (evidence rotated away — which is the bug). Fix 1 = the `017`/`c80fffc83` template applied here. Do NOT conflate with `003` |
| 035 | `site_work_items.updated_at` is not maintained (4,156 of 4,643 complete rows have `updated_at == created_at`; no trigger, one unrelated Go writer) — so a finished job reads as one that never ran. `completed_at` IS reliable | filed 2026-07-20; one-trigger fix candidate |
| 036 | A reviewer emitting `"edit": "<free text>"` where the struct wants `int` voids the whole council round at `council_decide` — after every seat has run and been paid for. **Not `019`**: the JSON is complete and VALID, so the truncation salvage cannot help. 3 of 5 `council_decide` voids in 14 days, all naming the same seat | **CLOSED 2026-07-20 → `/bugs_closed/`** — fixed (`58f5a6bb6` + `ab158c32a`) and VERIFIED LIVE on v1.0.1140 by **forced reproduction**: a scratch-council seat emitted `"edit": "edit 1 (comment-only change to…)"` and the round reached `complete_revise` with a `council_report` written and `unreadable 0`, where the same shape previously gave `complete_invalid` + no report. The report round-trips a string and an int side by side. All 3 live payloads had been plan-level *prose*, so `json.Number` (the case file's own candidate) would have parsed NONE of them. Residual: candidate (2)'s `salvageMistypedReview` is unit-proven only — the guardian seat refused to self-corrupt, correctly. Shared by **5 pipelines**, not one |
| 037 | A page flagged `needs_rebuild` is outside `001`'s guard, so a re-plan takes the LLM's composition — dartsonline `index` lost `differentiators` + `content-listing`. May be fix step 4's intended escape hatch; filed so the boundary is decided, not inherited. 34 pages currently `needs_rebuild` | filed 2026-07-20; decision needed |
| 038 | A re-plan rebuilds EVERY deployed page and regenerates its content — `decideEmit` needs `built_from_plan_version == planID`, and a re-plan changes `planID` for the whole site, so `skip_built` never fires after the first plan (`pages_skipped_built: 0` measured). `001` secures structure; this is copy | filed 2026-07-20; measured live, no fix started |
| 039 | `pages.sections` stores the component **function**, `page_components` reference the component **name** (`hero-about` ⟷ `about-hero`) — a naive comparison reads correct pages as regressed. AND: 11 section entries resolve to no component at all, rendering a hollow 208-byte `<section>` on deployed pages (7 live stubs) while the build reports success. Detected by `check_empty_sections`, but every item is `unresolved` — same delivery gap as 023/033 | filed 2026-07-20; convention + real defect |

| 040 *(slug `failed_page_build…`; a second `040` exists, slug `kafka_dial_timeouts…` — cite this one as **040-partial-build**)* | A **partial** page build leaves the page `build_status='deployed'`, partially composed, AND stamped with `built_from_plan_version` — so `decideEmit` returns `skip_built` and the reconciler never revisits it. dartsonline `index`: 5 of 6 sections. **CORRECTED 2026-07-20: a build reporting `complete` produced the same shortfall, so the failure was never the cause — and the page is now stamped with the CURRENT plan, so `skip_built` fires and it is permanently a five-sixths page.** Why the section is dropped is UNKNOWN (template validity + deferral both ruled out). Fleet: **25 deployed pages short of their plan across 6 sites, 39 sections missing, 4 with zero components** — one verified live-blank (`<main>` empty, chrome only) | filed 2026-07-20; found by rebuilding via the framework's own route |
| 044 *(slug: `no_capability_inventory_dormant_agents_undetectable` — **a second, unrelated `044` exists**, slug `plan_sections_defers_empty_schema_components…`; resolve by slug per `/bugs_closed/README.md`)* | Nothing detects a capability that **exists but nothing routes work to**. All 49 discovery checks are site-scoped; none inspects the platform's own inventory, and "which agents never run" is not a per-site question. **57 of 122 measurable active agents have never run** (step-fingerprint method; `owner_agent_type` is USELESS here — 95,797 orchestrations are `generic`, which is how a first pass got a wrong 110). Mostly *retired* agents still `is_active`, but ~8 are current-generation repair capabilities that never fired — incl. `feature-implementer`, which its own workstream independently records as never-fired. Producer-side mirror of `033` | filed 2026-07-20 out of `bugs_closed/002` D; measured + method-validated, no fix started. Two halves: the detector, and an owner call on `is_active` hygiene (5 types have >1 active row) |

| 045 | The active library contains **exactly one** component able to serve a generic `hero-tool` section, and it is hard-wired to a Bayesian ranker — 14 `source:static` fallbacks (`Start Ranking Free`, `Calculate Rankings`, `Try the Bayesian Ranker`) that `content_data` cannot override. So every tool page asking for a tool hero gets another product's vocabulary. **Not a planner defect** — the plan asked for `hero-tool` and that was correct; the library is missing the component. Sibling of `039` (same selector: that one resolves to *nothing*, this one to the *wrong* thing) | filed 2026-07-20, split out of `023` class F. **Armed on 2 live pages** (`finetuning.uk/ai-agent-roi-estimator`, `ai-agent-orchestration.com/agent-complexity-estimator`) — both `needs_rebuild` with `hero-tool` still in `pages.sections`, clean today only because they have not been rebuilt. Fix = build a generic gated tool hero; **do NOT delete the `_pre_037` row**, it is the sole active row for its function |

| 041 | Section lookup (`loadSectionComponents`, `v3_site_actions.go:3353`) resolves by name then by function and **never normalises**, while a sibling path (`:3730`) does — so a `snake_case` section (`call_to_action`) matches nothing, `plan_sections` Path 3 raises a `needs_new_component` for a component that **already exists** as `call-to-action`, and the page deploys without the section. 10 such items, 4 sites, since 2026-05-18, all `failed`. Explains the BULK of 040's fleet sweep (`call_to_action` ×14, `hero` ×6, vs only 4 legitimate `skip_section`) | filed 2026-07-20; cause proven from code + live rows, no fix started |
| 047 | Every `batch_scrape` rejected as "Empty URL" before its own handler (guard preceded the action switch) — research-agent's scrape lane + V5 acquisition dead, disguised as await-timeouts | **CLOSED & LIVE v1.0.1145 (2026-07-21)** — fix `8d9d9051a`, moved to `/bugs_closed/`. Verified behaviourally (pure reordering, no pod-grep symbol): a `verify047` batch_scrape smoke reached `handleBatchScrape` and succeeded, no "Empty URL" |
| 049 | **312 live broken links across 7 sites** (68 unique 404 targets, 117 of 180 live pages), measured against the shipped artefacts. 204 of them are `/privacy.html` + `/terms.html` in the footer of **every page** of three sites. Cause proven both directions: the chrome renderer's hardcoded legal-link slice was **fixed 2026-06-10 (`0681e1542`)**, but chrome re-renders only on explicit trigger and nothing sweeps it — those three sites' `site_components` date from **28 Apr / 21 May**, pre-fix; every post-fix render in the fleet is correct. Two further mechanisms: a page row that exists but was never built (`active` + `needs_rebuild`) which `check_phantom_internal_links` passes **by design** ("a planned-but-unbuilt page has a row and is not flagged"), and 32 extension-less targets on a `.html` fleet. **023's gating sweep does not cover any of it** — `{{if .x_url}}` tests non-emptiness and `/privacy.html` is non-empty | filed 2026-07-20 from the `023` sweep; cause proven from code + timestamps + a two-directional control, no fix started. Candidate 1 (re-render chrome on 3 sites) removes 204 and needs **no image** — but is outward-facing, owner go required |

| 050 | For a **deployed** page `sections=[]` means "rendered by another subsystem", not "awaiting composition" — all 18 such pages fleet-wide are `tool` (14), `blog-index` (2) or tool-ish `content` (2), and 15 have rendered `page_components` regardless. So `001`'s prescription for its own residual ("take the LLM's sections when the realised ones are empty") would let a re-plan attach a generic layout to all 18: **content injection onto built pages, the class `001` exists to stop.** The genuine defect survives only where the preserved page is NOT deployed — the first-plan case per `051`, which is exactly where every observed benefit occurred (dartsonline `guides-index`/`brands-index`/`shop-index`, all `planned`). Corrected fix gates on deployed-ness, not emptiness, in Pass B **and** Pass B2 — the latter closes a pre-existing exposure `001` did not introduce | filed 2026-07-20, split out of `001` so that case could close. Measured, not shipped: a fleet-wide rule about what a re-plan may write to a built page resting on an interpretation of an empty column — routed to the council gate, which `001` itself never got |
| 051 *(**a second, unrelated `051` may appear** — this one is the planner/adoption-lock case; the other `049`-numbered case that day is stale chrome + 404s. Resolve by slug)* | The per-page **90-day adoption lock does not exist.** `053` §054 designs `adoption_locked` as "no current plan OR a live timed per-page preserve-directive"; the live query carries only the first branch, and the second would match nothing anyway — 462 `site_plan_directives` rows, `locked_by` NULL on every one, **zero** rows at `scope='page'`+`category='preserve'`, and no code writes one. So `adoption_locked` is a **per-SITE** flag meaning "this site has no current plan": true on a site's FIRST plan, false on every re-plan. Consequence: **Pass C2 can only ever fire on a first plan**, and the comment justifying its scope appealed to a 90-day window that is not there. Also: the documented "adoption faithfulness for 90 days" is undelivered | filed 2026-07-20 while closing `001`. Live exposure ≈ nil (3 `planned` pages, all dartsonline artefacts of `001`'s own verification) — **filed for the wrong premise, not the damage.** Comments corrected in `1a13e265d`; the `needs_rebuild` slice belongs to `037` |
| 053 | `GetNavItems` overloaded a zero-row nav-table result: "no nav tables here" and "this nav-table site has no items in this group" both fell through to the pages fallback, which for `legal` matches every `in_footer` page. So a full-nav site with no legal pages (robot-hands + ≥5 more) filled its footer's legal slot with **all 14 footer links, none legal**; on gaswholesalers one is a live 404. See §9 *"An empty result overloaded to mean two things"* | **FIX COMMITTED (candidate 1, `85d39f9b9`, 2026-07-21)** — fallback now gated on `siteHasAnyNavItems` (fall back only when the site has **no** nav rows at all), fixing every group type at once; sqlmock regression test. **Stays OPEN**: inert until an image roll **and** a chrome re-render. Candidate 3 (deployedOnly for chrome nav) deferred to `052` — its correct predicate is `deployed_at IS NULL`, not today's `build_status='deployed'` |

> **Index gap (noted 2026-07-19, partly closed 2026-07-20):** `025`–`033` exist in
> `/bugs_open/` but are not all indexed here (`034`–`041` are; `042`–`047` exist and are not), and `027` is already used by **two** different cases. Filed by
> concurrent threads; list them with `ls bugs_open/` rather than trusting this table
> to be complete.

### A confident claim in our own knowledge base is not evidence (2026-07-19)

**Symptom.** A review, a plan or a debugging session is stopped by a stated
"contract" of the system — quoted with authority, specific enough to sound
load-bearing, and fatal to the approach if true. In the originating case a
council seat blocked a correct fix at HIGH severity by quoting the concept
register: scoped page-rerender *"SKIPS pages whose content hash is unchanged"*,
which would have made the fix a guaranteed no-op.

**Root cause.** The claim had **never been true**. `grep` found no `content_hash`
anywhere in the rerender path and `git log -S "content_hash"` over those files
returned **no commits at all**. What almost certainly happened: an earlier thread
observed a real symptom (pages not updating after a change), inferred a mechanism
for it, and wrote the inference into the register in the register's confident
declarative voice — with no file:line and no reproduction. From outside, the
three legitimate `carried` paths and a pre-check escalation are indistinguishable
from "skipped because nothing changed".

**Why it bites harder than an ordinary stale doc.** The register is **agent-facing
input**. A wrong entry is not just misleading prose — it gets quoted as evidence
in machine reviews and treated as decisive. And it was replicated across five
files, so correcting one fixes nothing.

**Durable rules.**
1. **Verify a cited contract against the code before revising your plan around
   it.** Three greps and one live probe disproved this one. Reorganising a
   correct design around an unverified quote is the expensive failure.
2. **`git log -S "<symbol>" -- <files>` distinguishes *stale* from *never true*.**
   That distinction changes the fix: stale means update the entry; never-true
   means find out what the author actually observed, because that symptom was
   real and is probably still unexplained.
3. **An assertion about code behaviour needs a file:line.** Without one it is an
   observation, and should be voiced as one ("we observed X"), not as a contract
   ("the pipeline skips Y"). Applies to the register, to handoffs, and to §9
   entries like this one.
4. **When a doc and the code disagree, the code wins — and the doc is now a bug.**
   File it (`bugs_open/031`), don't just fix your own copy; the next reader gets
   the same wrong answer.
5. **Chase the objection anyway.** Disproving this one meant reading the area
   properly, which surfaced a genuine adjacent defect nobody had named
   (`plan_sections_action.go` carries, never re-renders, an empty-schema component
   whose function name matches article/content/body/text/blog). A wrong objection
   pointed at the right neighbourhood.

### "Verified live" by grepping a generic property — the fix that was never rendered (2026-07-19)

**Symptom.** A fix is applied, deployed, reported successful, and re-verification
still fails with a byte-identical signal. Repeated cycles produce "materially
identical" fixes and the defect never moves. It reads as a model that cannot
aim, or a loop that will not converge.

**Root cause (the class).** The artefact being *tested* was never updated. A
durable-layer write (`content_components.html_template`) and the thing actually
served (`page_components.rendered_html` → the live page) are **different
storage**, joined only by a re-render step that can silently no-op. In
`bugs_open/024` three mechanisms stacked: a `build_status='pending'` flag nothing
reads; a rerender gate keyed on `spec.reason` that falls through to
assemble-the-stored-HTML when the field is absent; and a safety guard that
escalates (and discards the render) for any section with empty `content_data` —
which is every self-contained tool. All three report success.

**The verification mistake that hid it for two days.** Every "the fix is live"
claim was made by grepping a **generic CSS property** (`max-width`, `min-width`,
`minmax(`) on the page. Those appear many times in unrelated site-chrome rules,
so the check finds a hit and can never fail. Three separate claims — including
one in a turn log and two made by the diagnosing thread itself, one caught before
publication — were wrong for exactly this reason.

**Durable rules.**
1. **Verify a specific fix by matching its specific rule, not a property it uses.**
   `grep -A5 '\.the-actual-selector {'`, never `grep -c 'max-width'`.
2. **Compare the layers explicitly** — template vs stored render vs live bytes —
   and treat agreement as something to prove, not assume. If the stored render's
   length still equals the **v1 born length** in `component_versions`, it has
   never been re-rendered, whatever the page's `build_status` says.
3. **A green work item is not evidence that content changed.** Extend the
   standing `complete` ≠ *it happened* invariant one layer: here the work item,
   the orchestration AND the page status were all legitimately green over an
   unchanged page.
4. **Before diagnosing a fixer as unable to converge, prove the input to the
   re-test actually changed between attempts.** Two identical RED results are
   much more likely to mean an identical page than an identically-wrong fix.
5. **A flag is not a mechanism.** `build_status='pending'` looked like wiring and
   was read by nothing. Grep for a consumer before believing a status field.

### A recreated tool with no data source invents its own records — and says so in a comment (2026-07-18)

**Symptom.** An adopted site's data-backed widget (search/filter/list over a real
dataset) is rebuilt by the cascade and comes back *working beautifully* — search,
sort, region filter, pagination all responsive — over records that do not exist.
Nothing fails: `needs_tool_recreation` and every `page_rerender` stamp `complete`.
On vetcomparison.uk this shipped fabricated UK veterinary practices to live
visitors, on a site whose whole remediation history is about never publishing an
unsourced figure.

**Mechanism.** Two defects compound. (1) The tool-recreation path has **no
data-dependency contract**: adoption's `extract_interactive_fingerprint` does not
carry the original tool's `fetch()` target through to `tool-recreation-handler`,
whose prompt asks for self-contained HTML/CSS/JS. A tool whose behaviour *is* its
data therefore cannot be recreated faithfully — the model must emit a dead empty
widget or invent records so the interactions demonstrably work. It invents, and
documents the decision: *"For this recreation we generate a large, realistic,
deterministic dataset."* (2) The prompt's prohibition is **scoped to arithmetic,
not data**: rule 9 reads "No fake data or dummy outputs — calculations must be
mathematically correct", sitting among rules about function completeness, so it
reads as a statement about calculators and does not bind record invention.

**Generalise it.** Any generative step that must produce list-shaped output while
its real source is unreachable has this failure available to it, and the output
is *more* convincing than a broken one — plausible names, plausible postcodes,
deterministic so it looks stable across loads. Prohibitions phrased about
"correctness" do not cover invention; the prohibition has to name **records**.

**Tells to grep in generated JS** (cheap, catches the variant family): seeded PRNG
(`Mulberry32`, `imul(`, `let seed`), fragment arrays crossed to build labels
(`PREFIXES`/`SUFFIXES`/`TOWNS`), `buildData(`/`generate*(`, literal record arrays
over ~20 entries, and comments containing "realistic"/"representative"/"for this
recreation".

**Verify on the artefact, never the item.** `curl <page> | grep -iE
'Mulberry32|makePostcode|buildData|SUFFIXES'` must be empty. Also read the page's
*visible text*: the same rebuild claimed "pricing information, ownership data" the
site does not publish and called a real 2,109-row directory "a representative
sample for demonstration purposes" — copy-level fabrication that no structural
check would catch. Full case: `/bugs_open/020`. Sibling mechanism: `001`
(re-plan clobber resurrecting audited-out fabrication).

**Related-bug rule.** Before filing a new bug, grep this index for the mechanism —
005/008/009/012 are all one truncation-and-config family found by four different
threads. Filing the fifth copy costs more than reading the first four.

### A verifier that treats a MISSING target as success cannot tell repair from deletion (2026-07-19)

**Shape.** A check re-runs its own predicate to confirm a defect is gone. The
target row has vanished. The natural branch — *"gone, so nothing is wrong with
it"* — is wrong whenever deletion is itself a failure mode of the system being
checked.

**Live instance.** `VerifyEmptySectionResolved`
(`check_empty_sections.go:205`) returned
`VerifyResult{Resolved: true, Detail: "component no longer exists"}`. But a
missing `page_components` row is the signature of a rebuild silently dropping a
component — the platform's most repeated content-loss failure (`021`, `012`). So
a content-loss incident was being recorded as a *verified fix*, by the mechanism
built to stop self-reported completions being trusted. Full case: `/bugs_open/032`.

**Why it recurs.** The branch reads as defensive — it avoids an error on a row
that isn't there — so it survives review as robustness. It was also the
*reference implementation*: `verifiers.go:17-19` tells new verifiers to copy it,
and a plan to copy it to two more item types was in council review when a
`bug_historian` seat caught it. A blind spot in a reference implementation
propagates by being followed correctly.

**The rule.** *Absent is UNKNOWN, not resolved.* Where the framework has a
fail-open policy (this one completes the item and records the error), returning
an error is strictly better than a verdict: item flow is unchanged, and a silent
false positive becomes a visible unknown. Only claim `Resolved: false` if you can
positively establish the target *should* still exist (a plan row, a slot
reference) — otherwise a legitimate removal burns an attempt and, at
`max_attempts`, strands a fine item in `failed`.

**Generalises to.** Any confirm-the-fix path whose target can be deleted by
something other than the fix: cache entries, rendered artefacts, generated files,
child rows behind a DELETE+INSERT rebuild. Ask "could absence here mean the thing
I am guarding against *already happened*?" — if yes, absence is evidence, not
exemption.

### "Verify against the current code" silently means "against every session's uncommitted work" (2026-07-20)

**Symptom.** A triage pass audited all 33 files in `/bugs_open/` against the tree,
explicitly to avoid trusting each file's stale self-description. It reported
`bugs_open/008` item 5 (undecoded `stop_reason=refusal`) as already fixed and
shipped in commit `f32b208e5`, citing a line number and a test file. Both
citations were real and readable. The verdict was wrong: `f32b208e5` is the
*truncation* fix and contains neither.

**Mechanism.** This repo has many concurrent sessions on one working tree. The
audit read the tree — correctly, since that is where current code lives — found
another session's uncommitted refusal fix and its untracked test file, and then
attributed them to the most plausible nearby commit. Every individual step is
defensible; the composition invents provenance. Nothing in the tree distinguishes
"committed and shipped" from "someone is mid-edit", so a file-level read cannot
tell a fixed bug from a bug being fixed *right now* by someone else.

**Why it is worse than a stale bug file.** The failure it produced was
*confidence in the wrong direction*. A stale file over-reports work as
outstanding, and the cost is a re-check. This under-reported work as done, and
recommended moving the case to `/bugs_closed/` — which would have closed a live
defect on the strength of an edit that might never be committed at all. It also
credited one session's work to another session's commit, which is exactly the
attribution damage `bugs_open/001`'s misattributed evidence section caused.

**The rule.** For any claim about what the codebase *is* — as opposed to what you
are about to change — read the committed ref, not the tree:

```
git show HEAD:<path>                       # what the code actually says
git merge-base --is-ancestor <sha> HEAD    # is that fix even in this branch
git status --short <path>                  # is what I just read someone's WIP
```

For any claim about what *production* does, neither is enough — grep the running
pod's binary (§ "Verifying a deploy"). The three answer different questions and
this pattern is what happens when one is used for another's.

**Generalises to.** Every automated or delegated audit in a shared tree: coverage
reports, "is this still broken" sweeps, dependency checks, subagent triage. If
the task says "check current state", say which ref, and have it report the ref it
read. A verdict without a ref is not a verification — a distinction this guide
already draws for "verified" claims generally, now with a second way to get it
wrong.

**Corollary for delegated work.** A subagent given "verify against the tree" will
not know the tree is shared. Put the constraint in the prompt, not in your own
head, and treat any provenance claim it returns (commit shas, "already shipped")
as unverified until you check it yourself. Three of that pass's four
already-fixed verdicts held up; the fourth was the one touching the file the
delegating session had open.

### A config value that equals its code default proves nothing about whether config is wired up (2026-07-20)

**Symptom.** `content-feed-orchestrator`'s `render_news_json` step config carried
`max_age_hours: 72`, and the news card behaved like a 72-hour window. Everything agreed.
Raising it to `720` changed nothing: the run still rendered `item_count: 0` while the same
query, run by hand with 720, returned rows.

**Mechanism.** `ExtractActionInputs`
(`platform/orchestration/datahelpers/action_inputs.go`) reads step config in four places and
every one asserts `config[field].(string)` — and even then treats that string as a *reference*
to resolve against `collectedData`, never as a literal. A JSON number fails the assertion, so
nothing lands in `result.Values` and the action falls through to its call-site fallback:
`inputs.GetInt("max_age_hours", 72)`. The configured value had **never** been read. Neither had
`max_items: 6` in the same block. Full case: `/bugs_open/042`.

**Why it survived so long.** *The seeded value equalled the code default.* 72 in config, 72 in
Go. Config and behaviour agreed, so the config looked live and load-bearing while being purely
decorative. Nothing could look wrong, because nothing did look wrong.

**The rule.** **To test whether a config value is wired up, set it to something whose effect
would be visibly different, then observe the behaviour — never re-read the config.** Agreement
between a setting and observed behaviour is not evidence the setting caused the behaviour; it is
equally consistent with the setting being ignored and a default coinciding. This is the config
analogue of "trust the rendered artefact, not the status".

**Corollary for tests.** A regression test that configures a value equal to the fallback passes
whether or not the plumbing works. The tests added with the fix deliberately configure values
that *differ* from their fallbacks (720 against a 72 fallback, 30 against 6) — otherwise they
would prove nothing.

**Generalises to.** Any layered-default system where a declared value and a hardcoded default
can coincide: step config → action fallback, env var → compiled default, site spec → platform
default, kustomize overlay → base. The dangerous case is not disagreement, which surfaces; it is
*agreement*, which hides. Where a config key exists specifically to be tuned, prove once that
tuning it does anything at all.

### First-found-wins config lookup makes the MORE SPECIFIC block dead — and the trap is invisible to the author who wrote it (2026-07-20)

**Symptom.** `diagnose-agent`'s verdict step declared `max_tokens: 8000` inside its
step-level `ai_service` block. Every verdict call since 2026-07-10 logged
`max_tokens=2048`, and fix plans were truncated mid-JSON. Setting `max_tokens: 32000`
on the agent's **root** block fixed it immediately — the opposite of where the config
had been written.

**Mechanism.** `ExecuteLLMPromptAction` resolved the `ai_service` block by trying
three locations **in order, stopping at the first hit**: root, then the current
workflow step's `config.ai_service`, then `params.StepConfig.Config`. The two step
lookups sat behind `if aiServiceConfig == nil`. So for any agent that also declared a
root block, the step lookup **never ran**, and the step's ENTIRE block — model,
provider, max_tokens — was unreachable. Not the key; the whole block. Where neither
block declared a cap, `platform/aiservice/anthropic.go`'s hardcoded
`"max_tokens": 2048` applied and cut the completion. Full case: `/bugs_open/009`.

**Why it survived so long, and this is the transferable part.** The precedence is
**backwards from every convention** — specific normally beats general — so the
config author writes the override in the natural place, sees plausible output, and
never suspects it. Nothing errors. Nothing logs "your block was ignored". The value
is *visible in the config*, which is exactly what makes it look wired up; the only
tell is the runtime value, and only if you go and read it (`llm_call_log`, or
byte-counts clustering at ~7.5–8.1KB ≈ 2048 tokens).

**And it inverted the documentation.** One agent (`page-content-writer`) has no root
block, so ITS step-level fix worked — and that single success was generalised into a
fleet-wide runbook rule stating the exact opposite of the truth ("max_tokens lives
INSIDE a step's `ai_service` block; root is dead config"). That rule then sat in two
handoffs for three months. **A rule derived from one agent that happened to lack the
shadowing block is indistinguishable from a correct rule until you test it on an
agent that has one.**

**The fix shape.** Overlay, not selection: copy the general block, then apply each
more-specific block **key by key**, so a step overriding only `max_tokens` still
inherits the root provider and model. Step-wins-*wholesale* is the tempting
simplification and it is wrong — four `site-adoption-agent` steps declare
provider/model and omit `max_tokens`, and wholesale would have dropped them from
16000 to the hardcoded 2048. The blast radius of a precedence change is decided by
which keys each block **omits**, so audit omissions, not just conflicts.

**Generalises to:** any resolver written as a chain of `if x == nil` fallbacks over
config scopes — env/file/flag, tenant/site/global, root/step. The chain reads as
"precedence" but implements "first non-empty wins at whole-object granularity",
which silently discards the specific scope's other keys. Two questions that find it
fast: *does a lower-priority source existing make a higher-priority one unreachable?*
and *if I set a value here and it did nothing, would anything at all tell me?* If the
answer to the second is no, verify the RUNTIME value before trusting the config.

**Related:** §9 *"A config value that equals its code default proves nothing about
whether config is wired up"* (same family — config that reads as live but was never
parsed); `/bugs_open/008` (a capped call should fail loudly, closing this class end
to end); `/bugs_open/012` (`output_tokens == max_tokens` means CUT, not finished).

### A step name is not an agent — several workflows share one, and the run that "proves" your fix may belong to a different agent (2026-07-20)

**Symptom.** You verify a fix by querying a log keyed on `step_name`, get a clean
post-fix row with correct timestamps, and conclude the fix is proven. It is proven
— for a different agent. Your agent never ran. Nothing in the numbers shows this,
because every column you selected is genuinely correct.

**The instance.** `bugs_open/016_…council_revise_prompts_drop_reviewer_output`
(resolve 016 by SLUG — `bugs_closed/016` is the ssh/`$HOME` case). Finding 1: council
revisers rendered every reviewer as `<no value>`. The fix landed on `fix-proposer` at
13:15:11Z. A verification query joined `llm_call_log.step_name='repropose'` to
`orchestration_states.created_at` — correctly avoiding the step-time-vs-run-time trap
the same document warns about — and found run `a8b66dee`, started 15:27:33Z, rendering
**no** `<no value>`. That was written up as "PROVEN IN THE WILD".

`a8b66dee` is the **feature-designer**. Three agents (`fix-proposer`,
`feature-designer`, `experience-planner`) each have a step called `repropose`.
`fix-proposer`'s last repropose belongs to run `48cf0339`, started 13:11:13Z — four
minutes *before* its own fix. Every fix-proposer repropose on record is pre-fix, and
the fix is still unexercised two days later. Two threads read the write-up and neither
caught it; a third nearly "corrected" the bug file's accurate closing line to match
the inaccurate proof.

**Why the usual defences don't fire.** `owner_agent_type` is `generic` fleet-wide on
these rows, and `llm_call_log.agent_type='fix-proposer'` returns *nothing ever* — so
the obvious discriminators read as "never ran" rather than as "wrong question", and
you fall back to the step name. The document even warned that `persist_plan` is not
fix-proposer-only. That warning was scoped to one step name, and the scoping stood:
**a document warning you about a trap does not inoculate you against the same trap in
its neighbouring paragraph.**

**The check.** Fingerprint the run by the steps it actually carries, which is
unambiguous and costs one query:

```sql
SELECT string_agg(k, ', ' ORDER BY k)
FROM orchestration_states o, LATERAL jsonb_object_keys(o.workflow_plan->'steps') k
WHERE o.orchestration_id::text LIKE '<orch-prefix>%';
```

`load_diagnosis`/`propose`/`code_lookup`/`select_panel` + 13 seats → `fix-proposer`;
`load_spec`/`design`/`check_spec_approved` + 5 seats → `feature-designer`;
`review_journeys`/`review_feasibility`/`compose` → `experience-planner`. Better still,
key the probe on a step only ONE agent has (`load_council_reviews` is fix-proposer's).

Cheapest tell of all, and what actually caught it: **read the rendered prompt, not just
the flags you selected on it.** `a8b66dee`'s opened *"REVISE the staged build plan …
stages are commits … capabilities listed missing"* — designer vocabulary. `fix-proposer`
revises an *edit plan* against a *diagnosis*. One `left(prompt_rendered, 700)` settles
what no amount of timestamp arithmetic can.

**The general form.** `step_name`, `agent_type`, `orchestration_name` and `source_agent`
are all *labels applied by the writer*, and the chassis flattens several of them to
`generic`. The `workflow_plan` is *structure* and cannot be flattened. When a label is
load-bearing for a claim, verify it against structure.

**Related:** §9 *"A pod-grep marker that the build does not retain reads exactly like a
stale deploy"* and *"The pod-grep passes even when nothing shipped"* (same family — a
check that passes on something other than what you meant); §9 *"A fix applied to one
branch of a two-branch router reads as done"* (the sibling entry from the same bug);
`WRONG_CALLS.md` 2026-07-20 for the near-miss this came from.

### A scheduled task's `input_data` authored as a full message envelope double-wraps — the payload lands one level too deep and the action aborts as if unconfigured (2026-07-21)

**Symptom.** A config-driven action fails on a REQUIRED field it was clearly given —
here `directory_export_json` aborted "requires an explicit domain" while
`scheduled_tasks.input_data` plainly carried `domain: vetcomparison.uk`. Reproduces
every run; looks like the config was ignored (cf. `bugs_closed/042`, a genuine
config-ignored bug — same symptom, different cause).

**Mechanism.** The scheduler's `fireTrigger` (`cmd/scheduler/main.go`) is generic: it
ALWAYS builds the Kafka body as `{action:orchestrate, config:{agent_type},
input_data:<the whole scheduled_tasks.input_data column>}`. So the column must hold the
**payload only**. This task's seed had wrapped the payload in its own
`{action,config,input_data:{...}}` envelope, so the body became double-enveloped and the
real fields sat at `body.input_data.input_data.*`. `BuildCollectedData` unwraps exactly
one `input_data`, so `collectedData["input_data"]` = `{action,config,input_data:{domain}}`;
the action merges that one level and finds no `domain`. The action's hard refusal was
**correct fail-closed behaviour** — the data was wrong, not the code (`bugs_open/054`,
FIXED & LIVE 2026-07-21, DB + seed, no image roll).

**The transferable rules.**
1. `scheduled_tasks.input_data` is the PAYLOAD ONLY — the scheduler supplies
   `action`/`config`/the `input_data` wrapper. Never put those keys *inside* it. Correct
   models already in the table: `ch-enrichment`, `vet-batch-verify`, `vet-sweep-continue`.
2. When a required field is "present but missing", print the field's ACTUAL path in the
   live `collected_data`, not the DB column — the gap is almost always a nesting level,
   not an absent value. One `jsonb_pretty(collected_data->'input_data')` on the failed
   run ended this in one query.
3. Do NOT "fix" it by teaching the transport to flatten a nested `input_data`: that
   reintroduces the `bugs_closed/042` regression class (a legitimate payload field named
   `input_data`/`action`/`config` would be silently eaten). Fix the data; document the
   contract.
4. Symptom-family ≠ cause-family. 042 grouped this bug as "literal string not reaching
   an action" from the symptom, without reading the failing action — which does not even
   use `ExtractActionInputs`. Read the action before assigning a case to a family.

**Related:** `bugs_closed/042` (the mis-grouped sibling; corrected there);
`bugs_open/054`; §9 *"Manually invoking an agent via spawn+call — input_mapping must
satisfy BOTH the input_contract AND the workflow's field paths"* (same nesting-contract
family); `WRONG_CALLS.md` 2026-07-20 (the diagnosis-filing near-miss on this same case).

### A reader that understands one schema dialect FAILS OPEN on every other dialect — "couldn't read the contract" becomes "there is no contract" (2026-07-21)

**Symptom.** A field declared `required` in a component's `input_schema` renders empty,
saves, deploys, and serves — with nothing anywhere logging that a required field was
missing. The enforcement gate that exists precisely to stop this (`missingRequiredLLMFields`,
"refusing to render an empty section") ran and found nothing wrong.

**The instance.** `bugs_open/026`: the shared `news-listing` component's `<h1>{{.headline}}</h1>`
served empty. `headline` was `required:true, source:"llm"`. Root cause was **not** a
validator treating `""` as present — it was schema-**shape** blindness. The component's
`input_schema` was the old JSON-Schema dialect
(`{"type":"object","required":["headline"],"properties":{…}}`). Two independent readers parse
**only** the v2 `input_schema.fields.<name>` dialect:

- **Generation** — `planSection` (`plan_sections_action.go:1182`):
  `comp.InputSchema["fields"].(map)` misses → falls to `"no field schema — all fields from
  LLM"` with an empty `llmFieldSpecs`. The writer is never told the field exists → never
  generates it. (A name-keyword backstop at :1206 defers only `article/content/body/text/blog`;
  `news-listing` matches none.)
- **Enforcement** — `missingRequiredLLMFields` (`json_envelope.go:192`): reads only
  `inputSchema["fields"]` → `nil` → zero enforcement.

Both readers returned "nothing here" on a schema that in fact declared a required field.
Downstream treated *"I can't read this contract"* identically to *"there is no contract."*
That is failing **open**. A required field was therefore neither **requested** (generation)
nor **checked** (enforcement) — the two independent safety nets shared the exact same blind
spot, so having two did not help.

**The general shape.** Any parser of the form `x, ok := m["known_key"].(T); if !ok { return
nil }` fails open when the input is a *different valid dialect* rather than genuinely empty.
The nil/empty return is indistinguishable to the caller from "legitimately nothing to do."
Empty-schema components (tools, decorative sections) rely on exactly that nil return, so you
cannot simply make the miss an error — you must distinguish *"no contract"* (empty `{}`) from
*"a contract I don't speak"* (non-empty, carries `properties`/`required[]` but no `fields`).

**What to do.**
- When a schema/format flips dialects (here: a seed migrating `properties`→`fields`), **grep
  every consumer of the OLD dialect before assuming the flip is inert.** The flip fixed the
  one migrated component; every *other* consumer still silently mis-reads any un-migrated row.
- Ground "is the old dialect extinct?" against the live DB, don't assume the migration was
  total: `count(*) FILTER (WHERE input_schema ? 'properties' AND NOT (input_schema ? 'fields'))`.
- Prefer normalising the old dialect into the view the readers already consume (so old-shape
  rows become *understood*, not merely *rejected*) over bolting a second parser onto each site.

**Related:** `bugs_open/026` (the case); the "two-branch router" entry above (same "two nets,
one blind spot" family, different mechanism — there it was placement, here it is dialect).

### A completeness gate must count ROWS, not names — the plan's section names do not equal the live component names (2026-07-21)

**Symptom.** `bugs_open/040-partial-build`: a page build that wrote 5 of 6 planned sections was
stamped `deployed` + `built_from_plan_version = current`, so `decideEmit` returns `skip_built`
and the reconciler never revisits the missing section — a permanent five-sixths page. The fix
(candidate 1): at the deploy mark (`UpdatePageStatusAction`), refuse to stamp when the build is
short of its plan.

**The trap that nearly shipped a fleet-wide false positive.** The obvious way to ask "is this page
complete?" is per-section name matching: for each name in `pages.sections`, does a `page_components`
row exist whose `slot_name` (or `content_components.function`) equals it? That is the
`bugs_open/039` Part-1 pattern and it reads as the *more precise* check. **It is wrong as a gate.**
`pages.sections` names are authored by the planner and do **not** reliably equal the live
component `slot_name`/`function`. Measured live 2026-07-21: name matching flagged **74** deployed
pages; row-count matching flagged **28**. The ~46 extra were healthy pages, e.g.
`gaswholesalers.com/services` — planned `["services-hero",…,"call_to_action"]` vs live slots
`["hero-services",…,"call-to-action"]`: **word order swapped** and **underscore vs hyphen**, on a
page serving four full components. A per-name gate refuses it and drives it into a `needs_rebuild`
loop.

**The heuristic.** For a *gate* (something that refuses/loops on a hit), a false positive is far
more expensive than a false negative — it damages healthy state. Prefer the coarser, robust signal
(`count(page_components) < count(sections − suppressed_sections)`) over the precise-but-brittle one.
The count method's own false negative — a page with a duplicate row *and* a missing section counts
equal and passes — is acceptable because the fleet sweep and the `incomplete_page_group` /
`empty_sections` discovery checks are the other layers. **Validate any completeness/eligibility
predicate fleet-wide against live data before wiring it to an action**; a predicate that looks
sharper on one grounding page (dartsonline, where names happened to match exactly) can be
systematically wrong across the fleet.

**Also:** exclude `suppressed_sections` from the planned count (a deliberately-dropped section is
not a shortfall), and remember `sections=[]` pages are legitimately rendered by another subsystem
(`bugs_open/050`) — `planned=0`, never a shortfall.

**Related:** `bugs_open/040-partial-build` (the case + fix); `bugs_open/039` (the per-name matching
pattern — correct for *detecting* a hollow section, wrong as a *deploy gate*); `bugs_open/050`
(`sections=[]` means "rendered elsewhere", not "empty").

### A scheduler that claims a slot before deciding it has work starves its group — and a no-op that never stamps pins itself at the queue head forever (2026-07-21)

**Symptom.** `bugs_open/048`: four jobs in the scheduler's `maintenance` concurrency group
(`max_concurrent=1`) had not run for **79 days** — nothing errored, `enabled` stayed true, the
admin view showed them healthy.

**Mechanism (two parts, both required).** `cmd/scheduler/main.go` (1) incremented the in-memory
per-group `inFlight[group]` counter *before* running the task's pre-query, then `continue`d on the
no-rows / pre-query-error / merge-error paths **without releasing it** — so a job that found
nothing to do consumed its group's only slot for the whole tick; and (2) `loadDueTasks` orders
`last_triggered_at ASC NULLS FIRST`, and the no-rows path never stamped `last_triggered_at`, so the
job stayed permanently at the **head** of its group and re-won the slot on the very next tick. It
could not lose its place by doing nothing, and doing nothing is what kept it there. "Nothing to do"
is the *normal steady state* for a reaper, so this fires forever once the backlog it watches is
empty.

**The verification trap (a false GREEN).** The starvation is **intermittent under the old binary**:
whenever the watched backlog is momentarily non-empty, the head job returns rows, runs its real
work, stamps via the *success* path, and rotates to the back — briefly un-starving the whole group —
until the backlog empties and it re-pins. So a post-roll snapshot showing all the group's
`last_triggered_at` freshly advanced is **not** proof of the fix; it can be the old binary in its
transient un-stuck phase. The discriminating test is to reproduce the *failing* condition: with the
watched set empty (here `blocked=0`), does the head job **stamp**? Old binary → no (silent
`continue`); fixed binary → yes. Confirm the empty condition AND the stamp AND a second advance one
interval later. (Same shape as the "verify the failing branch" invariant.)

**The fix (candidates 1+2 together).** Claim the slot only once the task commits to firing/doing
work (move the increment past every early exit) — cleaner than decrementing on each `continue`,
because it cannot leak on a path added later. And stamp `last_triggered_at`+`last_completed_at` on
a *successful no-op* too (a shared helper), so the job rotates like any completed run and
`last_triggered_at` means "we looked" — which every operator reading that column assumes. Leave
genuine error paths un-stamped so they retry and stay visible. Live on `kafka-scheduler:v1.0.1146`.

**Transferable heuristic.** Any capacity counter incremented *before* the "do we actually have
work?" decision leaks on every no-op path; take the slot at the point of commitment, not the point
of consideration. And any queue that sorts by a "last done" timestamp starves its tail if a no-op
doesn't advance that timestamp — a run that found nothing has still *run*.

**Related:** `bugs_open/048` (the case + fix); `bugs_open/029` (same starvation family, hung
in-flight spawns rather than a leaked in-memory slot); `bugs_open/044` (silent dormancy is
undetectable — the missing liveness alert that would have caught this in May).
