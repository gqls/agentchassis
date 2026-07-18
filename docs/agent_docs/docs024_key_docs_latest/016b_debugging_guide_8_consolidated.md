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

**Fix.** Component restored from `component_versions`; migration **168** raised
`improve_tool` + `generate_tool_html` 8000 → 32000. That is the exposure, not the
bug. The real fixes, in order: (a) **refuse to persist a structurally-collapsed
artifact** — leave the row untouched, fail the item honestly
(`needs_human_review`), write a NOTE recording the refusal; never a silent
success; (b) **decode `stop_reason`** so the caller can reject a truncated
completion (`bugs_open/008`, platform-wide); (c) ceilings last.

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

**Transferable rules.**
1. Before blaming a fixer for not converging, read the signal it was given — a
   loop that repeats the same wrong fix is usually being told the wrong target.
2. "Which element is widest" and "which element is the cause" are different
   questions in any inherited-geometry system.
3. An escalation path is part of a fix loop, not an optional extra.

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

**Before flipping such a commit**, confirm the dedupe that redelivery now depends on
actually covers every inbound path — here `processed_messages`
(`platform/orchestration/state.go:170/207`) already existed, which makes the fix small, but
its coverage was never verified. Redelivery without dedupe trades lost work for duplicated
work. Full case and fix shape: `bugs_open/003` §3d/§4.4.

### An action that exists in code but in no registry fails as "requires a topic" — and the failure is stamped 'complete' (2026-07-18)

**Symptom.** A handler agent's work items complete on schedule, yet the defect they
describe never changes. The item's own `result` holds the confession:
`response.status='failed'`, `response.error='WORKFLOW_INVALID: … step X with action
Y requires a topic'`.

**Mechanism, leg 1 (the misleading error).** Workflow validation
(`platform/validation/workflow.go:80`) classifies any action it does not find in
`actioncheck.IsLocalAction` as *remote* and demands a Kafka `topic`. That list
(`actioncheck/local_actions.go`) is hand-maintained and marked DEPRECATED in its own
header, while `actions/registry.go:1866` has a registry-backed `IsLocalAction` the
validator does not use. So "requires a topic" almost never means a topic is missing —
it means **the action was never registered**. `fix_forced_text_colors` is the proven
case: handler + input spec written, present in NEITHER list (same never-registered
family as `checkpoint_for_review`, and the same two-hand-rosters drift class as the
council seats).

**Mechanism, leg 2 (the lie).** `CompleteWorkItemAction`
(`load_work_item_actions.go` ~735-800) verifies before completing only for item
types with a registered verifier, and never reads the response it stores: a payload
whose own `status` is `failed` is written next to `status='complete'` in one UPDATE.
The item_key dedup then suppresses re-detection — churn that reads as progress.

**Diagnose in one query.** `SELECT id, item_type FROM site_work_items WHERE
status='complete' AND result->'response'->>'status'='failed';` — any row is this
class. And when you see "requires a topic": grep BOTH registries for the action name
before touching topics.

**Cross-refs.** `bugs_open/017` (case + fix candidates: register once, reconcile the
validator to the registry-backed list, and make complete_work_item treat a failed
response as a failed attempt). Kin: "Trust the rendered artefact, not the status"
(§ durable invariants) — this is the mechanical version for work items. Category
tags: `never-registered-action`, `two-roster-drift`, `false-complete`.

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

**Cross-refs.** `bugs_open/018`. Category tags: `stdin-theft` (new),
`silent-truncation`.

## 10. Open bug queue (`/bugs_open/`) — index

The repo-root `/bugs_open/` directory is the live queue of diagnosed-or-filed bugs
awaiting a fixing thread (it was `docs024_key_docs_latest/aaa_fails_to_mend/`
until 2026-07-17; ~23 documents still reference the old path). §9 above holds the
durable PATTERNS; the files below hold the case detail, evidence and fix
candidates. Read the file before acting — several are already fixed.

| # | Bug | State |
|---|---|---|
| 001 | Re-planning a site silently discards its built pages' composition | FIX written |
| 002 | Errors surfaced but not fixed (multi-error handoff, route individually) | open |
| 003 | Spawned children lose their response; parents hang until reaped. **§3d (2026-07-18): second root cause — the consume loop commits Kafka offsets BEFORE processing (at-most-once), so any restart destroys in-flight work; §4.4 is the at-least-once + rollout fix that unlocks CD** | open |
| 004 | Landing an image can silently blank an article body | superseded by 005 |
| 005 | Article-body blanking — root cause LLM truncation (`max_tokens`) | FIXED |
| 006 | Three idea.uk infra errors (runner cgroup, dead contact endpoint, …) | open |
| 007 | Applied-but-unrecorded migrations block the runner | instance resolved; tooling open |
| 008 | `GenerateText` never decodes `stop_reason` (silent truncation) | fix COMMITTED `f32b208e5` (br 085, both providers); not yet deployed |
| 009 | Root `ai_service` SHADOWS the step block (dead per-step config) | diagnosed; fix + fleet sweep open |
| 010 | Fix loop non-convergent on layout-intrinsic overflow | candidate (a) SHIPPED v1.0.1135; (b) open |
| 011 | `kind:"hero"` routes to SDXL (cannot render text); the Gemini infographic lane works and was unused | open |
| 012 | tool-improver truncates a component and saves the wreckage | exposure fixed (168); guard open |
| 013 | fix-implementer commits un-`gofmt`'d LLM output; build gate rejects it, no PR | filed; fix candidate (format at commit-prep) |
| 014 | VM-site artefacts silently deploy to the default `sites` repo (two causes) | FIXED (v1.0.1126 + pin removal) |
| 015 | Mistyped `page_type` orphans a page from every gate that keys on it | worked around per-site; planner fix open |
| 016 | `ssh` ignores `$HOME` (uses passwd entry) — service-account git-over-ssh fails twice over | FIXED in the box scripts |
| 017 | `fix_forced_text_colors` never registered ("requires a topic" lie); failed saga stamped 'complete' | filed; both legs open |

**Related-bug rule.** Before filing a new bug, grep this index for the mechanism —
005/008/009/012 are all one truncation-and-config family found by four different
threads. Filing the fifth copy costs more than reading the first four.
