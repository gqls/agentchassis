# HANDOFF — page build/re-render pipeline + phantom-CTA work

**Purpose.** Continuation point for a fresh chat (prior chat ran out of compaction room). Assume no
memory of the conversation. This file is the "start here" orientation; the detailed trail lives in the
other docs listed at the end. Everything below is in `/mnt/user-data/outputs/`.

---

## 0. How work happens here (operating model)

- **The platform** builds/maintains multi-page websites with a fleet of agents. Each agent is a row in
  Postgres `agent_definitions` (DB `clients_db`), runs as a K8s pod from the chassis image
  `docker.io/aqls/agent-chassis`, communicates over Kafka, and persists run state in
  `orchestration_states`.
- **Work** = rows in `site_work_items`, claimed by the `build-dispatch-loop` (claimable status
  `triaged`/`approved`), routed to handler agents that spawn child agents.
- **I (the assistant) have NO cluster or DB access** — my network is locked to package registries. The
  human runs all `kubectl`/`psql`/`git` and pastes results back. So: I propose exact SQL/commands; they run
  them; I read the output.
- **DB access (human runs):**
  `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`
  Give **bare SQL** (no `\x`, no psql meta unless asked).
- **Two kinds of change, very different cost:**
  - **Workflow change** = editing an agent's `default_config` JSON in the DB. **Immediate, no redeploy.**
    This is how most fixes here land.
  - **Go code change** = code → GitHub → Backblaze → new chassis image → **bump `image_tag`** on the
    affected agent rows. Slow. Only when logic must change in Go.
- **Back up before editing an agent:** `SELECT snapshot_agent('<type>','<reason>');` and revert with
  `SELECT revert_agent('<type>');` (these are the human's helpers; exact signature is theirs).

### Schema gotchas (verified this project)
- `orchestration_states`: PK is `orchestration_id` (there is **no** `id` column). Useful columns:
  `parent_orchestration_id`, `owner_agent_type`, `collected_data` (jsonb), `final_result` (jsonb),
  `site_id`, `responses_topic`, `status`, `created_at`.
- `site_work_items`: error column is `error` (not `error_message`); category column is `pipeline` (not
  `domain`). Dedup index `idx_swi_dedup` = UNIQUE(`site_id`,`item_key`) over **non-terminal** rows. Terminal
  statuses: `complete`, `failed`, `verified`, `rejected`, `wont_fix`, `unresolved`.
- `agent_definitions`: the identity column is **`type`** (not `name`).
- A work-item spec that needs page routing **must include `page_name`**.
- **psql JSON operators:** `#>` takes a **text[] path** (`'{a,b}'`) and returns jsonb; `#>>` returns text;
  `->` / `->>` take a single key. (A bare string after `#>` errors with "malformed array literal".)

### Known gamesdesign.co.uk IDs
- site `e33263f4-74f8-494f-b191-546845dbbddf`
- index page `6e988cc4-4898-4021-aa5e-2ab0271f9b75`
- game-pathfinding `56af8679-1f7d-4da6-b148-f5727b16693d` (`/games/pathfinding/index.html`)
- guide-economy-basics `8ed97fd2-2e33-4bdb-a24e-0f3badaca382`
- guide-rng-design `e40b26cc-049b-4c60-becf-803f8bf1430b`
- contact-index `66d42e30-…` (url `/contact/index.html`)

---

## 1. Standing rules (the human's preferences — follow these)

- **Prefer structural fixes over quick patches.** Reuse existing functions/structs/actions rather than
  adding new ones.
- **Every agent is an orchestrator.** Thin workflows, complexity in Go. **No SQL sub-workflows** — spawn a
  sub-agent instead.
- **Agents respond to the caller's (parent's) responses topic.**
- **Check the schema before writing SQL.** Don't treat a `0 rows` / `false` result as decisive until you've
  ruled out that the query/path itself is wrong (this bit us twice — an ILIKE probe and a path typo).
- **Don't rename workflow variables** except deliberately, and when you do, **say so explicitly**. Keep
  workflow variable names in sync with what the actions actually return.
- Go style: **no `logger.Debug`** (use `Info`). Avoid the words "perfect/critical/excellent"; no
  congratulatory filler. Keep replies moderately brief.

---

## 2. Status of every thread

| Thread | State | Next action |
|---|---|---|
| **Part 1** — result-contract drop | **DONE** — shipped 2026-06-18, verified | none |
| **Part 2** — no-LLM re-render path | **DEPLOYED** 2026-06-21; `image_landed` verified; index 5 sections in `page_components` | live `/index.html` re-checked 2026-06-24 → deploys only 4 (see `system-stats` row); finish P2.4–P2.7 |
| **`system-stats` component `fdd92ad4` (cross-site)** | **concluded for gamesdesign (closed-by-removal); shared-component fix OUT OF SCOPE** — only 1 component version (the 15:06 `component-creator:regen`, `change_source=manual` → likely the other chat); **no revert target**; live component coherent (template-scheme, 22/22); the **5** broken instances are all other sites; durable platform bug = regen renames fields + re-renders dependents from un-migrated `content_data` (writer `stat_1_number` vs component `stat1_value`) | leave the shared component (co-managed, no revert target); **flag** the platform bug to its owners; gamesdesign needs nothing here — pivot to Part 4 deploy / phantom-CTA / `rerender-pages` |
| **Part 3** — `item_key` canonicalization | **CODE PREPARED**, not applied | apply *after* Part 2 verifies |
| **Part 4** — interactive page rebuilt as plain content | **CAUSE CONFIRMED; fix WRITTEN** (patched file in outputs) | `go build ./...`, deploy + bump image_tag, set `work_item_id_field`, then re-create the game |
| **Phantom hero CTA** | **ROOT CAUSE CONFIRMED** 2026-06-23; fix SQL ready | **apply the path-fix SQL (below) — this is the top item** |
| **(new) `rerender_single_page` invalid page_id** | re-pointed — `rerender-site` has **0 orchestrations ever**; the real per-page looper is **`rerender-pages`** (16 runs, last 15:19); some morning rows may still be retained | pull `rerender-pages` `default_config` (loop `items_field` + `page_id` mapping when spawning `page-rerender`) + dump a recent run. Lower priority than `system-stats` |

---

## 3. Phantom hero CTA — CONFIRMED, fix ready (DO THIS FIRST)

**Symptom.** Deployed `guide-economy-basics` hero links to phantom URLs: `cta_url=/contact.html`
("Read the Full Guide") and `secondary_cta_url=/services.html` ("Browse Economy Guides"). Real contact is
`/contact/index.html`; there is no services page. The page's footer correctly uses `/contact/index.html`,
so the bug is specifically hero CTA resolution. These are the generic hero schema's brochure defaults
(`cta_url ← pages.contact`, `secondary_cta_url ← pages.services`).

**Root cause (confirmed via orchestration_states):** a **path mismatch**, not a chassis-version or
render-override issue (both ruled out — see below). `page-content-writer` spawns `internal-link-resolver`,
which DOES run and DOES return augmented sections. Its reply lands in the parent at
`resolved_links.response.link_resolution.sections_ready`. But the writer's `select_sections` step reads
`resolved_links.sections_ready` (top level) → null → falls back to `input_data.section_plan.sections_ready`
(the un-augmented plan, which carries the schema-default `/contact.html`). So the resolver's correct hub
links are computed and then discarded.

Things ruled out: both agents are on the same image `v1.0.1070` (not a Part-1-not-deployed issue);
`render_component` is NOT overriding (it's the fallback branch); section-index hubs exist (the resolver
isn't starved of targets); the resolver's empty `final_result` is a red herring (the parent's reply is built
from the resolver's collected_data, so data flows regardless).

**THE FIX (workflow-only, immediate, no redeploy). Variable-path change, noted explicitly:**
`select_sections.config.fields.sections_ready[0]`: `resolved_links.sections_ready` →
`resolved_links.response.link_resolution.sections_ready` (the fallback to
`input_data.section_plan.sections_ready` stays).

```sql
SELECT snapshot_agent('page-content-writer', 'pre select_sections resolved_links path fix');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,select_sections,config,fields,sections_ready}',
      '["resolved_links.response.link_resolution.sections_ready","input_data.section_plan.sections_ready"]'::jsonb
    )
WHERE type = 'page-content-writer';

-- verify
SELECT default_config #> '{workflow,steps,select_sections,config,fields,sections_ready}'
FROM agent_definitions WHERE type = 'page-content-writer';
```

**Then clear the live phantom.** This takes effect only on the **next FULL build** of a page (a bare
`page_rerender` won't re-run the resolver). Trigger a full rebuild / content rewrite for
`guide-economy-basics`; the hero is the generic `hero` component (function `"hero"`, which the resolver
matches), so it will resolve `cta_url` to a real hub (`/games/index.html`) and `secondary_cta_url` to
`/guides/index.html`. Verify by re-pulling the deployed HTML or the page's `page_components`.

**Two follow-ups (separate from the fix above):**
1. **Resolver function matching is too narrow.** `resolve_internal_links`' `ctaFieldNames` matches only the
   exact functions `"hero"` and `"call-to-action"`. Component variants that carry `source:"renderer"` CTA
   fields — e.g. `hero-about`, `platform-comparison`, `gauntlet-cta` — are skipped, so their CTAs never
   resolve even after the path fix. Broaden the matching. (Confirmed from a real envelope dump of another
   site's about page.)
2. **Resolver returns its whole collected_data** (the echoed input includes every component's full HTML —
   the reply is enormous) and its `final_result` is empty. Make `internal-link-resolver` return a lean
   `{sections_ready, unresolved}`; then `select_sections` can simplify to
   `resolved_links.response.sections_ready`. Needs the `complete_workflow` result-spec behaviour (`output`
   mapping vs `output_field`) confirmed first — its `complete` step currently uses `config.output` mapping
   and that isn't flattening into `final_result`.

---

## 4. Part 4 — interactive page rebuilt as plain content (cause confirmed; build the fix)

**Symptom.** `game-pathfinding` should host a playable A* game; the live page is hero + generic text only.
**Confirmed:** the hero once held the full 18,449-byte A* game (`<canvas id="gridCanvas">`, brushes,
Run/Reset). A full rebuild (`link_resolution_rebuild`) ran the page from the page **spec**; the interactive
tool is a `page_component` **not in the spec**, so it was omitted, and `save_page_sections` (DELETE+INSERT)
dropped it. Any full rebuild path (`link_resolution_rebuild`/`needs_page`/`needs_content_page`/
`content_rewrite`/admin Regenerate) has this hazard. Blast radius today is one page (`game-pathfinding`).

**Fix — WRITTEN 2026-06-24.** Patched `save_page_sections_action.go` is in `/mnt/user-data/outputs/`
(brace/paren-balanced; needs `go build ./...` in-repo, then deploy). Interactivity is detected from
`page_components.rendered_html` via `ILIKE '%<canvas%' OR '%game-container%' OR '%tool-page%'` (there is
**no** stored `has_interactive_markup` column — computed in the sweep). **Key insight from the source:** the
interactive tool exists ONLY as `rendered_html` in `page_components`; the planning path
(`load_page_sections_from_spec` → `plan_sections`) deals in section-*name* skeletons + LLM-regeneratable
content and CANNOT reconstruct a bespoke `<canvas>`/JS tool. So preservation must happen at the **save**
path (carry the existing markup forward) — **both layers live in `save_page_sections`**, not the planner.
- **Layer 2 — carry-forward** (inserted before the content-regression guard). SELECT existing deployed
  interactive sections; for each: same slot present & interactive → keep; same slot but rebuild produced
  non-interactive content (the hero-regenerated-as-text case) → replace incoming HTML/`content_data` with
  the preserved interactive markup; slot dropped → re-append the tool. Runs BEFORE Layer 1, so a successful
  carry-forward stops Layer 1 firing. (INSERT loop uses `i+1` for position, so order = slice order;
  re-appended tools land last.)
- **Layer 1 — interactivity guard** (inserted after the content-regression guard). If the existing deployed
  page is interactive (`bool_or` on the same signal) and the new set is not → return
  `"interactivity regression blocked…"`. Safety net for when Layer 2 can't preserve (e.g. preload query
  fails). Helper `sectionHTMLIsInteractive(html)` added (`strings.ToLower` + `Contains` on the three markers).
- **source_item_id** — the history snapshot INSERT now writes `source_item_id = $2` from a config-driven
  `work_item_id_field` (nil → SQL NULL, unchanged until the workflow sets it). No new imports.
- **Layer 3 (re-route through `page_rerender`) RULED OUT** — P4.2: preserves sections but doesn't re-resolve CTAs.

**To finish (separate from the Go change):**
1. **Workflow config** for `source_item_id` to populate: add `"work_item_id_field": "<collected_data path to
   the driving work-item id>"` to the `save_sections` step config on every agent that saves sections — match
   the value `plan_sections` already receives via its `work_item_id` input. Confirm the exact path from an
   orchestration `collected_data` dump. (Immediate, no redeploy.)
2. **Deploy the Go change:** `go build ./...` → GitHub → Backblaze → new chassis image → `snapshot_agent`
   then bump `image_tag` on agents that run `save_page_sections` — confirmed **three**: `page-build-handler`,
   `page-rerender`, and `tool-recreation-handler`. Layer 1 fires only on interactive pages → guides/index unaffected.
3. **After deploy:** re-create the `game-pathfinding` A* tool (and fix its `needs_page:game-pathfinding`
   mis-key per Part 3). The carried-forward tool ideally needs `content_data` so a future `page_rerender`
   doesn't escalate it to `needs_page` (tool-recreation's concern — follow-up).
4. Delete the stale duplicate `z_context/check_phantom_internal_links.go` (the live `platform/...` copy is correct).

---

## 5. Part 2 — no-LLM re-render path (deployed; finish verification)

`flag_page_image_rebuild` / `reconcile_section_data` now emit `page_rerender` (not full-LLM `needs_page`).
New `rerender_page_sections_action.go` re-renders ALL sections from stored `content_data` + re-resolved
fields (NULL `content_data` escalates to `needs_page`); the page-rerender `default_config` got a gated
pre-pass. **Verified:** P2.3 Test 1 (index, `image_landed`) passed — all 5 sections re-rendered, copy
preserved, hero CTA resolving. **Pending:** P2.4 (real image flow), P2.5 (`section_data_resolved`), P2.6
(NULL escalation), P2.7 (plain backward-compat); and confirm the live index shows all 5 sections. Test plan
detail is in the RUNBOOK.

---

## 6. Part 3 — item_key canonicalization (code prepared; apply after Part 2)

`apply_adoption_plan_action.go` hardcodes `needs_page:%s` for BOTH the tool and content branches → key
collision. **Decision (Option B):** tool branch → `workItemKey("needs_tool_recreation", page.Name)`; content
branch stays `workItemKey("needs_page", page.Name)` (preserves the deliberate doc-029 co-dedup). The
`workItemKey(itemType, target string) string` builder is prepared in `work_items_common.go`. **To do:** add
the builder + the adoption-loop diff, `go build ./...`, bump `image_tag` on the adoption / site-adoption
handler, run P3.4 — **after** Part 2 is verified. (This is a Go change, so it needs a redeploy.)

---

## 7. Part 1 — result-contract drop (DONE, for reference)

Chassis coordinator `extractWorkflowResult` previously honoured only plural `output_fields`; singular
`output_field` and the `output`/`result_mapping` mapping form were ignored → fallback dump → oversize → a
stub `{status:completed … exceeded size limit}` that parents stored as a silent no-op. Fixed in
`result_spec.go` (`resolveResultSpec`: singular `output_field` → flatten, plural `output_fields` → nest,
`output`/`result_mapping` → mapping) + `coordinator.go` (oversize now fails loudly; stub removed). Shipped
2026-06-18, documented in 016 §9. **Note:** the phantom-CTA follow-up #2 above suggests the `output` mapping
form may still not flatten into `final_result` on `v1.0.1070` — worth re-checking against this fix if you
pursue that cleanup.

---

## 8. Prioritized next actions

1. **Apply the phantom-CTA `select_sections` path fix** (§3 SQL) and trigger a full rebuild of
   `guide-economy-basics`; verify the hero CTAs become real hubs. Lowest-risk, confirmed, user-visible.
2. **Part 4:** implement Layer 1 guard + Layer 2 preserve-sections; re-create the `game-pathfinding` game.
3. **Part 2:** run P2.4–P2.7; confirm the live index shows 5 sections.
4. **Part 3:** apply the `workItemKey` builder + adoption diff, build, bump image, run P3.4 (after #3).
5. **Follow-ups:** widen `resolve_internal_links` `ctaFieldNames`; make the resolver return a lean result;
   delete stale `z_context/check_phantom_internal_links.go`; stamp `source_item_id` in
   `page_component_history`.

Threads 1–5 are independent of each other except where noted (Part 3 gates on Part 2).

---

## 9. The maintained docs (where the detail lives)

- **`RUNBOOK_gamesdesign_index_rebuild.md`** — operational apply/deploy/verify runbook. Has Parts 1–4 with
  apply SQL + verify steps, the P4.2 test, and the full phantom-CTA thread (now ending in the confirmed
  root cause + fix SQL + the two follow-ups). **Most useful day-to-day.**
- **`NOTES_gamesdesign_silent_norebuild.md`** — dated running investigation log; keeps the
  hypothesis→correction history (including the retractions). Authoritative trail.
- **`016_debugging_guide_v2_56.md`** — canonical debugging guide, CLOSED at v2_56 (final §9 = the Part-4
  writeup + "recurring debugging trap part 3": `completed_at` is the orchestration END, not the write
  instant — trace by `page_id` + time).
- **`016b_debugging_guide.md`** — Vol.2 (continues 016); durable heuristics + open threads.
- Edited canonical copies (kept their uploaded filenames; renumber/bump on the way in):
  `000_documentation_index.md`, `002_system_architecture_2_.md`, `005_tool_pipeline.md`,
  `020_tool_lifecycle_1_.md`, `026_component_regeneration_flow_1_.md`.
- **`work_items_common.go`** — prepared Part 3 `workItemKey` builder.

---

## 10. Useful diagnostic recipes (reusable)

```sql
-- What a page's components currently hold (and when last written)
SELECT slot_name, updated_at, left(content_data::text, 200)
FROM page_components WHERE page_id = '<page_id>' ORDER BY slot_name;

-- History of writes to a page's components (source + when)
SELECT created_at, source, source_item_id, content_data->>'slot_name' AS slot
FROM page_component_history WHERE page_id = '<page_id>' ORDER BY created_at;

-- Did a spawned agent surface its output, and on what topic?
SELECT orchestration_id, parent_orchestration_id, status, created_at, responses_topic,
       left(final_result::text, 500)
FROM orchestration_states
WHERE owner_agent_type = '<agent-type>' ORDER BY created_at DESC LIMIT 5;

-- Dump an agent's workflow steps (top level)
SELECT jsonb_object_keys(default_config #> '{workflow,steps}')
FROM agent_definitions WHERE type = '<agent-type>';

-- Full config of specific steps
SELECT key AS step, jsonb_pretty(value)
FROM agent_definitions ad, jsonb_each(ad.default_config #> '{workflow,steps}')
WHERE ad.type = '<agent-type>' AND key IN ('stepA','stepB') ORDER BY step;
```

**Reminder when reading results:** a `0 rows` / `false` / empty field is not proof of absence until the
query path itself is verified. Several wrong turns this project came from trusting a single probe.

---

## 11. Bundles (to gather code+docs for a fresh chat)

The human runs these manually (the `cmd/bundle` tool, contextkit) and pastes the resulting `.md` into the
next chat. Verify the `docs/...` prefixes and any uncertain action filenames against `registry.go` before
running. Scopes are the entry points; the tool expands via the call graph from the named symbol.

**Phantom-CTA bundle** (`-out /tmp/bundle_phantom_cta.md`): scopes
`resolve_internal_links_action.go:ResolveInternalLinksAction`, `extract_fields_action.go:ExtractFieldsAction`,
`call_agent_action.go:CallAgentAction`, `complete_workflow_action.go:CompleteWorkflowAction`; includes
`compile_page_sections_action.go`, `render_component_action.go`, `registry.go`; docs `016_debugging_guide_v2_56`,
`005_tool_pipeline_1_`, `003_contracts_and_standards_7_`; schema
`orchestration_states,agent_definitions,pages,page_components,page_component_history,site_work_items`; runtime
`gamesdesign.co.uk` / `guide-economy-basics`.

**Interactive-clobber bundle** (`-out /tmp/bundle_clobber.md`): scopes
`save_page_sections_action.go:SavePageSectionsAction`, `plan_sections_action.go`,
`load_page_sections_from_spec_action.go`; includes `registry.go`; docs `016_debugging_guide_v2_56`,
`026_component_regeneration_flow_1_`, `003_contracts_and_standards_7_`, `020_tool_lifecycle_2_`; schema
`page_components,pages,page_component_history,site_work_items`; runtime `gamesdesign.co.uk` / `game-pathfinding`.

Both use the common flags: `-analysis /tmp/analysis_repo.json -root ~/projects/agentchassis -constitution
thin_slice_constitution.md -step debug -psql '<kubectl psql>' -capabilities -df-filter snapshot`. The full
`go run` invocations were given in chat; reproduce them from the scope/doc/schema/runtime lists above.

**Known `cmd/bundle` run errors (to fix before regenerating either bundle):** (1) quote `-doc` paths
containing parentheses — unquoted `()` is a bash syntax error and stops `go run` executing; (2) regenerate
`/tmp/analysis_repo.json` if empty (`load analysis: unexpected end of JSON input`). Full write-up for the
bundling-tool maintainers is in `contextkit_bundle_issues.md` (worst-first, with repro).

**Not covered by any bundle:** the session output docs (`HANDOFF_page_pipeline.md`,
`RUNBOOK_gamesdesign_index_rebuild.md`, `NOTES_gamesdesign_silent_norebuild.md`, the `016b` additions) live in
`/mnt/user-data/outputs/`, not the repo — carry them into the next chat directly.
