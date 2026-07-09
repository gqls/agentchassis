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
