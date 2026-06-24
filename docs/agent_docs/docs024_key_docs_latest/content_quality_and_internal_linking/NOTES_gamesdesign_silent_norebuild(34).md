# Running notes — gamesdesign.co.uk silent no-op rebuild

## Orientation (what this whole effort is)

The platform builds and maintains multi-page websites with autonomous **agents** (rows
in `agent_definitions`) that run as K8s pods (namespace `ai-persona-system`), communicate
over Kafka, and persist run state in Postgres `clients_db`; each unit of work is a
`site_work_items` row claimed by a handler agent. Deploy is image-tag based (code →
GitHub → Backblaze → new chassis image → bump each agent's `image_tag`); workflow
(`default_config`) changes are DB-only. These notes track a connected run of fixes to the
page build/re-render pipeline, all the same shape — **work that reports success but
doesn't happen, or work dropped/duplicated by key collisions** — opened by
gamesdesign.co.uk's `index` reporting successful rebuilds while the live page never
changed. Part 1 (the dropped-result contract) shipped 2026-06-18; Part 2 (the no-LLM
re-render path) is deploying; Part B (item_key canonicalization) is next. Operational
steps (SQL / kubectl / git) are run by the human — Claude has no cluster or DB access.
Full operational runbook: `RUNBOOK_gamesdesign_index_rebuild.md`.

---

*Root cause CONFIRMED at the framework boundary. The coordinator's
`extractWorkflowResult` has, since commit `06a8c6e` (2026-01-14, "made extract
function respect output fields"), honoured only `output_fields` (plural) — and
that commit also added the 900k cap + `extractMinimalResult` stub. The
page-content-writer's `complete` step declares `output_field` (SINGULAR),
unchanged across source migration `023` and live (v1.0.981 Apr → v1.0.1063 Jun).
The singular key is ignored → the fallback dumps the writer's whole state →
exceeds the cap → the stub drops the compiled page. NOT a 06-13 regression and
NOT the Apr→Jun chassis bump (both images already contain the Jan-14 code) — a
long-standing config/coordinator mismatch. FIX DEPLOYED to production 2026-06-18
(verification in progress): `extractWorkflowResult` rewritten to use a centralised
`resolveResultSpec`
(new `platform/orchestration/result_spec.go` + body edit in `coordinator.go`) —
singular `output_field`/`result_from` → FLATTEN, plural → nest, `output`/
`result_mapping` → applied, deprecated names still read. Pure chassis change, no
agent-config edits. Supersedes `PREAMBLE_gamesdesign_diagnosis_handoff.md` (its
"generation shortfall" framing is falsified). A generalised entry for this
failure mode is in `016_debugging_guide_v2_50.md` §9.*

DB access for all queries:
`kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## Symptom

The gamesdesign.co.uk `index` rebuild reports success, but the live page stays
stale (deployed `page_components` last written 2026-06-06 16:59; nothing since).
Work item completes, no error surfaces — invisible without inspecting stored state.

---

## ROOT CAUSE (confirmed) — child result dropped by the size-limit fallback

The `page-content-writer` produces a full compiled page, but its result never
reaches the parent `page-build-handler`. Two compounding defects:

1. **Wrong output key (data, the trigger).** The writer's `complete` step is
   `{"action":"complete_workflow","config":{"output_field":"page_content"}}`.
   `output_field` (singular) only controls where a step stores its own output in
   `collected_data`. The coordinator's `extractWorkflowResult`
   (`coordinator.go` L3357–3367) reads `output_fields` (**plural**). The singular
   key is invisible, so it takes the fallback (L3380): serialise nearly all of
   `collected_data` minus `skipPatterns`. Those patterns skip `page_content_`
   (trailing underscore → `page_content_0`), `assembled_page`, etc. — but NOT
   `page_content`, `processed_sections`, or `section_output_0..N`. For a
   multi-section page that dump exceeds `MaxResultSizeBytes` (900000, L3432).
2. **Lossy size guard (code, the silent failure).**
   `extractWorkflowResultWithSizeLimit` (L3434) tries to shrink it, but the
   trimmer (L3454) only truncates top-level **string** fields > 10k (source
   comment: "Could also recurse into maps/arrays if needed" — it does not). The
   bulk is nested objects/arrays, so nothing shrinks and it falls to
   `extractMinimalResult` (L3475) — a stub `{orchestration_id, completed_steps,
   completed_at, status:"completed", message:"Workflow completed successfully.
   Full result exceeded size limit."}`. The compiled page is gone, and the
   `status:"completed"` is a false success.

The parent stores that stub under `page_content.response`. `save_page_sections`
reads `page_content.response.sections_metadata` → null → no-op ("no sections
found", `sections_saved:0`) or falls back to a short `validation_result.clean_html`
that trips the regression guard. Page stays stale; the false success rolls the
`page_rerender` work item up to `complete` (this is the visibility fault, same
root).

### Scope / impact

Not gamesdesign-specific. Any agent whose `complete` step uses `output_field`
(singular), or whose legitimate result exceeds 900k, hits this. A sibling
Kafka-size failure was seen on `image-build-handler` `complete_workflow`.

---

## The flow (call chain, with the break point)

`page-build-handler`: `ensure_site_record` → `load_page_record` →
`load_existing_content` → `load_spec_sections` → `plan_sections` →
`check_has_ready_sections` → `spawn_content_writer` → `call_content_writer`
(`call_agent`, `output_field: page_content`) → [child] `page-content-writer`:
… → `compile_page` (output `{page_html, page_name, section_count,
sections_metadata}`) → `complete` (`output_field: page_content` ← WRONG KEY) →
**[coordinator `extractWorkflowResult`: no `output_fields` → dump state → >900k →
trim fails → minimal stub]** → parent stores stub at `page_content.response` →
`check_content_produced` (passes — `skipped` not set) → `validate_content` →
`save_sections` (`sections_metadata_field: page_content.response.sections_metadata`,
`html_field: validation_result.clean_html`) → null → no-op/short-fallback →
`update_status` → deploy → `complete`.

---

## Two faults (one fully reframed)

1. **Child result dropped at `complete` (ROOT CAUSE).** Was filed as a generation
   shortfall; generation is sound. The compiled result is discarded by the
   coordinator's size-limit fallback because the writer never declared
   `output_fields`.
2. **Status rollup (visibility fault).** A no-op/blocked save still rolls the
   `page_rerender` work item to `complete` — because the dropped-result stub
   reports `status:"completed"` and the save returns success on 0 sections. Same
   root; the coordinator hardening below fixes the silent-success half.

---

## Ruled out (checked and falsified — do not re-derive)

- **"Generated sections never reach save."** The sections are produced in full;
  the *response carrying them* is replaced by a stub before save.
- **Persisted section status in `site_plan_sections`.** Not stored; runtime only.
- **Triage starving the writer.** FALSE — every recent index run is
  `ready_count=5, deferred=0, skipped=0`; all five sections ready.
- **`max_tokens: 2000` truncation.** FALSE — per-section `generated_content`
  ~100–400 tokens, far under the cap.
- **Recreate-mode-not-firing as the volume cause.** Recreate-mode IS broken
  (`has_existing=false` on every recreate run) but does not drive the shortfall —
  the 06-06 deploying build also had `has_existing=false`. Separate defect.
- **Query-resolved list data vanished.** FALSE — `resolved_data` equivalent in the
  deploying vs blocked build.
- **"Guard counts `content_data` / `section_output` is short."** FALSE — full
  rendered HTML; guard reads `rendered_html`.
- **`compile_page` builds short `sections_metadata`.** FALSE — `extractSectionFromMap`
  / `extractHTMLFromSectionMap` (v3_site_actions.go L1838/L1917) set
  `rendered_html` to the FULL HTML; writer-side `page_content.sections_metadata`
  is a 5-element array, per-section stripped ~23k total.
- **The save path is misconfigured.** FALSE — `page_content.response.sections_metadata`
  resolves to the full array on the 06-06 build; it is null only because the
  response is a stub.
- **The bundle's writer def was stale.** FALSE — live `version 2` (updated
  2026-06-13 12:16) matches: `complete = {output_field: page_content}`.

---

## Confirmed (with evidence)

- **Writer output full** (child `472eed7d`): `section_count=5`, `page_html`
  34,523; `page_content.sections_metadata` 5 entries, per-section
  rendered_html/stripped 2288/1984, 9047/5306, 7458/5672, 8155/5077, 7369/4999
  (~23k stripped). `page_content` ~81k.
- **Parent received a stub** (06-15, parent `cd73eea6`): `page_content.response`
  = `{completed_at, completed_steps(number), message, orchestration_id, status}`,
  `message` = "Workflow completed successfully. Full result exceeded size limit."
- **Parent received the real output on the deploying build** (06-06, `4e0b339a`):
  `page_content.response` = `{page_html, page_name, section_count,
  sections_metadata(array,5)}` — the compile output **flattened**.
- **Coordinator change that caused it (git, confirmed):** commit `06a8c6e`,
  2026-01-14, "checked sizes in success message, and made extract function
  respect output fields." Before it, `extractWorkflowResult` dumped all
  `collected_data` as-is; the commit rewrote it to honour `output_fields`
  (plural) else the `skipPatterns` fallback, switched `notifyParentOfSuccess` to
  `extractWorkflowResultWithSizeLimit`, and added `MaxResultSizeBytes`/
  `extractMinimalResult`. This code is in both `v1.0.981` (Apr) and `v1.0.1063`
  (Jun), so it predates 06-06 — the Apr→Jun chassis bump is NOT the cause.
- **Config unchanged (backups, confirmed):** writer `complete` =
  `{output_field: page_content}` in `agent_definitions_backup_20260422`
  (`v1.0.981`, Apr 21) and live (`v1.0.1063`, Jun 13). Same as source migration
  `023` (seed + every UPDATE block). The singular key never changed.
- **OPEN ANOMALY — why did 06-06 ever save?** The 06-06 parent's
  `page_content.response` = `{page_html, page_name, section_count,
  sections_metadata}` — flattened compile output, and crucially WITHOUT the
  `orchestration_id`/`completed_steps`/`completed_at` keys that post-Jan-14
  `extractWorkflowResult` always appends. The current code cannot produce that
  shape (fallback keeps `page_content` nested + appends metadata; the
  `output_fields` branch nests + appends metadata). So an unaccounted
  response/storage path let 06-06 through. Does not change the fix direction, but
  the exact returned shape the fix must produce should be pinned against a real
  successful response (inspect a 06-06-era successful child's full response and
  the `call_agent`/`ExtractStepData` storage) before finalising.

---

## Guardrails (the tempting wrong moves)

- **Do NOT raise `MaxResultSizeBytes`** — the limit guards the Kafka ceiling;
  pushing the whole working state across the bus is the thing to stop.
- **Do NOT loosen the content-regression guard** — it correctly protects the live
  page; the INSERT persists `s.HTML`.
- **Do NOT "fix" generation, compile, recreate-mode, or the save path config** —
  all exonerated. (Recreate-mode is a separate real defect.)
- **Do NOT assume `output_fields: ["page_content"]` alone is the fix** — it nests
  under a `page_content` key (array → `page_content.response.page_content.sections_metadata`)
  and appends metadata keys, neither matching page-build-handler's read path.
  Config alone cannot reproduce the flat path; see Step 1 options.
- **Do NOT bound the writer's result without the wiring check** — content-reviewer
  reads the writer's response as the fallback DUMP (`processed_sections`,
  `compile_page`), so flattening/nesting/bounding the writer can break it IF it's
  on the live path. Confirm live-vs-legacy (Step 0) first.

---

## Plan — under investigation, NOTHING applied

**Naming fragility (design goal).** Three `complete.config` keys mean "what to
return": `output_field` (singular), `output_fields` (plural — the ONLY one the
coordinator honours), `output` (mapping — also ignored → fallback). `output_field`
vs `output_fields` differ by one letter but, under any fix that gives them
different shapes, mean different things — a future-mistake magnet. Direction:
rename the plural to a non-confusable name (`multiple_output_fields` /
`result_fields`) AND remove the shape-fork, transitionally (coordinator reads both
old + new name, warn on old; migrate ~90 plural configs over time). The rename is
DECOUPLED from the bug fix and is a ~90-config migration ("both places" each), so
sequence it separately.

**Step 0 — live wiring (RESOLVED).** Writer's live parents: `page-build-handler`,
`page-rebuild`, `site-work-orchestrator` (pageflow-builder set aside as legacy).
All three read the writer's response FLAT: page-build-handler
(`page_content.response.{skipped,page_html,sections_metadata}`); page-rebuild and
site-work-orchestrator (`page_content.response.page_html` for `assemble_page`;
both `save_sections` use `assembled_page.html`, not sections_metadata). So
flatten satisfies all three with no consumer edits AND fixes all three (all
currently broken: page-build-handler nulls at sections_metadata; the other two
null at page_html). `content-reviewer` IS on the writer→reviewer wire in
page-rebuild + site-work-orchestrator (`page_content: "page_content"`), but its
`auto_eval` template reads dump-only keys (`processed_sections`, `compile_page`),
so it is blind on a bounded writer result — and blind NOW too (stub). Flatten is
neutral-to-better for it (HITL reads `page_content.response.page_html` → starts
working). Latent follow-up: repoint content-reviewer auto_eval to
`page_content.response.sections_metadata[].rendered_html`. Not on the gamesdesign
path (page-build-handler doesn't call the reviewer).

**Step 1 — DEPLOYED to production 2026-06-18.** `extractWorkflowResult` body replaced to
use `resolveResultSpec`; new file `platform/orchestration/result_spec.go` holds
the resolver + `fallbackDumpInto` + helpers. Singular → FLATTEN, plural → nest,
mapping (`output`/`result_mapping`) → applied, deprecated names read with a Warn,
multi-key conflict logged. NO agent-config edits required. Diff is contained to
the one function (verified). REUSE: field-name lists go through
`datahelpers.ToStringSlice` (import `platform/orchestration/datahelpers`, same as
coordinator.go); `toStringMap`/`setIfAbsent` are local (no datahelpers
equivalent). Pre-merge: grep the rest of package orchestration for an existing
`toStringMap`/`setIfAbsent` (none in coordinator.go/processor.go).

  - *A-nest (rejected)* — would leave site-work-orchestrator (`site_plan.response.
    needs_images`) AND model-trainer (`preparation_result.dataset_uri`) broken.
  - *B (config)* — superseded by the chassis fix.

**Mapping rollout gate — RESOLVED (shipping all three together):**
  - content-reviewer → consumers gate on `reviewed_content.approved` (mapped
    target). SAFE.
  - internal-link-resolver → read by the writer; mapping exposes
    `sections_ready`/`unresolved` (its only outputs); created 2026-06-11 with the
    mapping, ignored since, so honouring it is CORRECTIVE.
  - research-agent → writer reads `research_result.response.summary` (mapped →
    corrective) and `.sources` (not mapped → stays empty, but empty under the
    dump too → not a regression). FOLLOW-UP: add `"sources":
    "research_content.sources"` to research-agent's mapping for citations.
  - site-architect → read by `website-builder` only (experimental/legacy path);
    the one unverified consumer, contained.

**Claims VERIFIED against live workflows:** flatten breaks none of the four and is
corrective for more than two —
  - writer → page-build-handler/page-rebuild/site-work-orchestrator: FIXED.
  - site-planner → site-work-orchestrator: FIXED (flatten lifts `needs_images`,
    `image_prompts`, `style_collection`, `pages` to `site_plan.response.*`).
  - training-data-preparer → model-trainer: corrective.
  - thunder-reaper: `reaper_summary` never set → metadata-only; no consumer → safe.
  - content-reviewer (auto_eval): blind now and under flatten; HITL improves;
    follow-up repoint to sections_metadata. Not a regression.

**Deployed 2026-06-18 — verifying.** Chassis image rebuilt and agents rolled to
the new tag (was `v1.0.1063`). Verify on a gamesdesign index rebuild:
`page_content.response.sections_metadata` arrives populated; `page_components`
rendered_html timestamps advance past 2026-06-06; the content-regression guard
passes on real content; the `page_rerender` work item only completes on a real
save. Watch logs for `resolveResultSpec: resolved result contract mode=flatten
matched_key=output_field` on the writer, and the deprecation Warns (a census of
which agents still use old key names → drives the rename migration).

**Follow-ups:**
  1. Size-path hardening — DONE (implemented 2026-06-18, same coordinator.go,
     pending deploy with Step 1). `extractWorkflowResultWithSizeLimit` now returns
     `(map, error)`: it no longer truncates strings and no longer emits the
     `status:"completed"` stub (both removed, `extractMinimalResult` deleted). On
     oversize it returns an actionable error via new `oversizeResultError` (logs a
     per-field size breakdown at Error level + names the largest field and the
     remedy: declare a result contract). `notifyParentOfSuccess` now calls
     `notifyParentOfFailure` on that error → parent gets `error_unrecoverable` /
     `Success:false` / `CHILD_ORCHESTRATION_FAILED` (Recoverable:false) + a
     persistent `agent_error_log` entry (severity fatal). Chose fail-loud over
     persist+reference (no consumer can deref a reference today) and over recursing
     the trimmer (truncating content delivered as success is a corrupt result,
     worse than a clean failure). Behaviour change: any agent that was
     stub-"succeeding" on an oversize result now fails loudly until it declares a
     contract — that surfacing is the point. Flatten already bounds the writer
     (~81k) so it never reaches this path.
  2. content-reviewer auto_eval: repoint template to
     `page_content.response.sections_metadata[].rendered_html` (currently reads
     dump-only `processed_sections`/`compile_page`; blind on a bounded result).
  3. research-agent: add `"sources": "research_content.sources"` to its `output`
     mapping so the writer's citation loop has data.
  4. Rename migration (optional, transitional): move configs from `output_field`
     → `result_from`, `output_fields` → `multiple_output_fields`, `output` →
     `result_mapping`. Driven by the deprecation Warns. Old names keep working
     until migrated, so this can be gradual; both-places each.

---

## Verification when fixed

- **Fault 1:** an index rebuild lands the full `sections_metadata` at
  `page_content.response.sections_metadata`, the guard does not fire, and
  `page_components.rendered_html` updates (timestamps advance past 06-06).
- **Fault 1 (negative):** a deliberately oversized result no longer collapses to
  a stub — it fails loudly (parent receives `error_unrecoverable` + an
  `agent_error_log` entry naming the largest field and the remedy) and never
  reports success while dropping output.
- **Fault 2:** a no-op/blocked save drives the work item to failed/needs-review,
  not `complete`.

---

## Verification (2026-06-18/19) — two wrong turns, then verified on the content path

Fix deployed to `page-build-handler` (chassis `v1.0.1065`, confirmed off `v1.0.1063`).
Tried to verify on the index page; the fix has NOT yet been exercised because both
attempts ran the wrong path. Recording so we don't repeat them.

**Wrong turn 1 — trigger SQL didn't target index.** First runbook re-opened a
`page_rerender` item with `ORDER BY created_at DESC LIMIT 1` and no page filter.
This site has one `page_rerender` per page from a single 06-13 17:19 batch (guides,
tools, games…), all near-identical `created_at`, so the unfiltered pick did not
reliably hit index. Fix: filter by `item_key='page_rerender_index_<site_id>'` (or
`spec->>'page_name'='index'`). Also `pages.url` is a path (`/index.html`), not the
domain — resolve via `sites.domain` → `site_id` → `pages.name`, never filter pages
by domain. (`agent_error_log` column is `error_message`, not `error`; `page_components.content_hash`
is empty on this page so change-detection uses `updated_at` + `rendered_html` length.)

**Wrong turn 2 — `page_rerender` is the wrong agent entirely (the real lesson).**
After targeting index correctly and re-queuing (dispatch loop claims `triaged`, not
`approved`), it completed in ~41s and still didn't change the page. The work-item
`result` proved why: the `page_rerender` item is handled by the **`page-rerender`**
agent, whose plan is `render_page (rerender_single_page) → check_skipped → deploy_page
(git_commit) → update_status → complete`. **No `page-content-writer` step.** It
re-assembles the page from the `page_components` already in the DB and commits to git.
Re-rendering the frozen 06-06 components produced the same page and deployed it
successfully (`deploy_result.success=true`, committed `/index.html` to `gqls/sites`
2026-06-19 09:23:57). So: not the stub bug, not the false-complete — a faithful
re-assemble of stale components. Our flatten fix is on a DIFFERENT path and was never
touched (every `called_writer` = `f`; no `resolveResultSpec` line for this run).

**Corrected trigger (confirmed from live config, not inferred):**
- `page-build-handler` (chassis `v1.0.1065`) IS the content path: `spawn_content_writer`
  → `call_content_writer` (`output_field: page_content`) → `check_content_produced`
  → `validate_content` → `save_sections`. Its `save_sections` reads
  `sections_metadata_field = "page_content.response.sections_metadata"` — the exact
  flat path the flatten fix restores. This is the consumer we fixed.
- Item types routing to `page-build-handler`: **`needs_page`** and `needs_content_page`
  (+ `link_resolution_rebuild`). `page_rerender` → `page-rerender` (assemble+deploy only);
  `needs_rerender` → `rerender-pages`. So a CONTENT rebuild of index is a `needs_page`
  item with `source='manual-rebuild'`, NOT a `page_rerender` item.
- Next: raise/re-open a `needs_page` item for index (claimable status `triaged`), or
  set `pages.build_status='needs_rebuild'` and let the discovery check raise it, then
  re-run §5 Stage A against the `page-build-handler` orchestration.

**Third attempt — SUCCEEDED (2026-06-18/19).** Ran a `needs_page` rebuild of index
via `page-build-handler` (content path). New writer content reached the live site:
the deployed index now reads hero "Your Probability Maths Is Wrong" with rewritten
tool/guide/game intros, and the CTAs point at real pages (`/tools/index.html`,
`/guides/index.html`) instead of the prior empty `href=""`. End-to-end fix confirmed
on the path it targets. Stage C confirmed: all five `page_components.updated_at`
advanced to 2026-06-19 10:34:50 (past 06-06 16:59); `system-stats` kept the same
length (7369) but its timestamp moved, i.e. regenerated to a coincidentally-equal
length, not skipped. Persistence half verified.

**`item_type` vs `item_key` mismatch (record for the routing doc).** The `needs_page`
history shows three `needs_page` rows keyed `item_key='page_rerender:index'` plus one
`needs_page:index`, while `page_rerender` rows are keyed `page_rerender_index_<site_id>`.
So the `item_key` prefix does NOT reliably indicate `item_type`/agent — route and
diagnose by `item_type → handler_agent` only. Captured in `002 — System Architecture`,
"Work-item routing: content rebuild vs assemble-only".

**Incidental findings (separate from the fix, from the deploy payload):**
- `page-rerender` `deploy_page` commit message renders literally `"Rerender: "` —
  `{{.filename}}` not interpolated in that step's scope.
- Index hero still links `/contact.html` and `/services.html` (stale); several CTA
  `href=""` empty; `content_hash` empty on all five components. All content-layer,
  fixable only by a writer rebuild (the path above).

---

## item_key contract + image-rerender type — investigation & plan (2026-06-19)

Triggered by: `needs_page` rows keyed `page_rerender:index`. Investigated the key
minting across all creators (source confirmed). Two interlocking problems.

**Creators and their keys (confirmed from source):**

| creator | item_type | item_key | correct? |
|---|---|---|---|
| `reconcile_site_plan_action.go:184` | needs_page | `needs_page:<name>` | ✓ canonical |
| `reconcile_site_plan` (rerender) | needs_rerender | `reconcile_rerender:<plan_id>` | namespace ok, prefix≠type |
| `flag_page_image_rebuild_action.go:123` | needs_page | `page_rerender:<page>` | ✗ type wrong (see Part A) |
| `reconcile_section_data_action.go:212` | needs_page | `page_rerender:<page>` | ✗ type wrong; shares flag's key by design |
| `apply_adoption_plan_action.go:655` | needs_content_page | `needs_page:<name>` | ✗ prefix≠type |
| `apply_adoption_plan_action.go:655` | needs_tool_recreation | `needs_page:<name>` | ✗ prefix≠type AND collides with content_page |
| `apply_adoption_plan_action.go:678` | needs_rerender | `adoption_rerender_<site_id>` | namespace ok, prefix≠type |
| `create_rerender_items` (rerender-pages) | page_rerender | `page_rerender_<page>_<site_id>` | prefix ~matches type |

Two insertion mechanisms exist: `insertWorkItem(workItem{…})` and inline
`INSERT … VALUES`. Any canonical builder must serve both → plain `func(target) string`.

**Two confirmed bugs:**
- *Bug 1 — content rebuilds don't co-dedup.* reconcile → `needs_page:<page>`;
  flag + section-data → `page_rerender:<page>` (also needs_page type, same handler,
  same work). Different keys → `idx_swi_dedup` sees them as distinct; reconcile's
  `loadOpenPageItems` "already queued?" check matches on `needs_page:<page>` and
  won't see a `page_rerender:<page>` row → double-build of one page.
- *Bug 2 — collision drops work.* adoption keys BOTH `needs_tool_recreation` and
  `needs_content_page` as `needs_page:<name>`; a tool item and a needs_page/content
  item of the same name collide on the unique index → one silently dropped, to the
  wrong handler. Same class as the original silent-loss bug.

**Part A — image-rerender type mismatch (decide FIRST; it sets the keys).** flag +
section-data emit `needs_page` (→ page-build-handler → writer, full content regen)
when the real need is *re-resolve a field + re-render the affected section*. Their
own `page_rerender:` key reveals the intended op is a re-render, not a content
rebuild. The only field-re-resolving path today is the heavyweight writer path (the
lightweight rerender paths reassemble stored HTML without re-resolving fields), so
they were forced onto it — at the cost of LLM spend and exposure to the
content-regression guard for an asset swap.
  To do: (1) establish where a resolved asset URL enters the stored HTML (templated
  wrapper vs baked into writer output) — load-bearing unknown; (2) add a
  field-re-resolving re-render capability (re-resolve via `ensureAssets`/`queryresolve`,
  re-render affected sections reusing existing text, reassemble, deploy — no writer),
  reusing plan_sections' resolution logic; (3) repoint flag + section-data to that
  re-render type, at which point `page_rerender:<page>` correctly matches the type;
  (4) evaluate the cheaper alternative first — page-build-handler already threads a
  `mode`/`build_mode` to the writer; if a content-preserving mode exists/can be added,
  keep `needs_page` with `mode=image_only` and the guard won't fire (content
  unchanged). Open questions: where the URL enters HTML; whether the writer has a
  reuse mode; whether one section can re-render without the whole page. Needs
  plan_sections / rerender_single_page / page-content-writer source (not yet read).

**Part A — feasibility CONFIRMED + chosen direction (2026-06-19): re-resolve field +
re-render the affected section, no writer.** Reviewed plan_sections, rerender_single_page,
rerender_pages (deprecated), render_site_components, render_news_section, queryresolve,
component_selector, load_page_sections_from_spec, store_generated_component. Findings:
  - `render_news_section` does NOT re-render a section — it writes JSON data files
    (`data/latest-news.json`) that the deployed component fetches client-side ("no page
    rerender needed"). Sibling pattern, not the primitive.
  - The real precedent is `render_site_components`: re-resolves fields (nav/contact from
    `pages`), renders header/footer/head templates from a data map, stores `rendered_html`,
    no LLM. The page-section analogue is the missing piece.
  - Primitives all exist: generic render `RenderTemplateWithMap(template, data)` (lives in
    deprecated rerender_pages_action.go — lift to a live file); `queryresolve.Resolve` for
    `query.*` list fields; hero/asset resolution already in plan_sections; and crucially
    `page_components.content_data` stores the writer's structured output SEPARATELY from
    `rendered_html` (render_news_section reads it) — so a section can be re-rendered from
    content_data + re-resolved fields + the component template without the writer.
  - Gap: nothing re-renders a section body today except the writer. `rerender_single_page`
    loads sections as stored `rendered_html` and concatenates; the deprecated bulk path
    re-renders only the shell.
  Proposal: build a `rerender_page_sections` step (page-section analogue of
  render_site_components) — load component template + content_data for the affected
  section(s), re-resolve fields (queryresolve + asset resolution), render via
  RenderTemplateWithMap, write `page_components.rendered_html`; then existing page-rerender
  assembles + deploys. Slot it into the **page-rerender** agent as an optional pre-pass
  gated by spec (`reason=image_landed`/`section_data_resolved`); repoint flag +
  section-data to emit a `page_rerender`-TYPE item routed there. Fixes type + key together
  (`page_rerender:<page>` becomes the correct prefix) and avoids the writer + regression
  guard entirely. Supersedes the Part-A option-4 (`mode=image_only`) idea.
  UNKNOWN NOW RESOLVED (2026-06-19, from v3_site_actions.go + save_page_sections_action.go):
  `content_data` IS complete enough to re-render from. `RenderComponentAction`
  (v3_site_actions.go:1372) builds content_data = LLM copy (`content_from`) overlaid with
  `resolved_data` (`merge_with`) — deliberate, per its comment, so resolved items/urls/labels
  persist alongside the copy. `save_page_sections` (save_page_sections_action.go:352) writes
  content_data back next to rendered_html; its regression guard (line 246) only fires on
  >75% text loss, so an image-only re-render passes. Only the site-level render-context base
  (company/colours/contact) is absent from content_data, and `BuildRenderContextAction`
  (v3_site_actions.go:739) rebuilds it.
  MECHANISM (no LLM): `render_component` already renders from content_from + merge_with +
  context_from with no LLM (the writer's `render_from_template` branch calls it that way).
  Re-render a section = render_component with content_from = STORED content_data, merge_with =
  FRESH resolved_data. Re-running plan_sections yields the fresh resolved_data (re-resolves
  hero from site_plan_imagery + assets). save_page_sections is whole-page (DELETE+reINSERT all
  rows), so either re-render every section from stored content_data (cheap, reuses save as-is)
  or UPDATE the single affected row directly (lighter, bypasses history snapshot).
  NO FURTHER FILES NEEDED to confirm feasibility. Remaining is a design fork:
   - Option X (smallest surface): add a no-LLM `rerender` branch to the writer's
     process_sections_loop (render from stored content_data instead of generate_content);
     flag + section-data keep emitting needs_page WITH the mode → reuse the whole page-build
     pipeline, LLM step swapped out. item_type stays needs_page (mismatch becomes cosmetic —
     no LLM cost, no guard exposure).
   - Option Y (correct typing): build a dedicated `rerender_page_sections` step in the
     page-rerender agent; route flag + section-data to a `page_rerender`-TYPE item. More
     surface; type + routing come out right; lightweight work stays in the lightweight agent.
   Sub-decisions either way: (a) re-render only the affected section (carry the rest) vs
   re-render all from stored content_data; (b) NULL content_data on older pages — carry
   existing rendered_html untouched, or fall back to the writer for just that section.

**DECISION (2026-06-20): Option Y, NULL → content generator.**
  - CHOSEN: Option Y — dedicated `rerender_page_sections` action, slotted into the
    **page-rerender** agent as a pre-pass; repoint `flag_page_image_rebuild` and
    `reconcile_section_data` to emit a `page_rerender`-TYPE item routed there. Fixes the
    type and the key together (no writer, no guard exposure, correct routing).
  - CHOSEN sub-decision (b): NULL content_data → rewire into the content generator. A
    section we must re-render that has NULL content_data cannot be re-rendered from stored
    content, so escalate that PAGE to the full writer (emit needs_page → page-build-handler),
    which regenerates AND backfills content_data. Self-healing: legacy pages get a one-time
    full rebuild, after which the light path works. The light path only handles sections
    that HAVE content_data.
  - CHOSEN sub-decision (a): re-render ALL sections from stored content_data (not
    affected-section-only). Consequences: (i) reuse `compile_page_sections` +
    `save_page_sections` as-is (whole-page DELETE+reINSERT is fine — we supply every
    section), no targeted UPDATE needed; (ii) NULL handling becomes all-or-nothing PER PAGE —
    if ANY section on the page has NULL content_data, escalate the WHOLE page to the writer
    (needs_page → page-build-handler) and skip the light path, because re-render-all needs
    every section to have stored content; (iii) no affected-section derivation — `reason`
    (image_landed / section_data_resolved) is only a GATE for whether to run, since
    re-rendering everything re-resolves every dynamic field including the changed one;
    (iv) the regression guard stays satisfied — all copy is preserved (NULL pages escalate
    rather than render empty), so total visible text ≈ existing.

  BUILD PLAN (Option Y, re-render-all):
  1. New action `rerender_page_sections(site_id, page_name, reason)`:
     load all page_components rows for the page (component_id/function, content_data,
     position). PRE-CHECK: if any row has NULL/empty content_data → emit needs_page (full
     writer rebuild, which backfills content_data) and return without the light render.
     Otherwise, for every section: re-resolve its dynamic fields → FRESH resolved_data, call
     `render_component` (content_from = stored content_data, merge_with = fresh resolved_data,
     context_from = rebuilt render_context) → collect fresh rendered_html per section.
  2. Persist via the existing whole-page path: feed the re-rendered sections through
     `compile_page_sections` → `save_page_sections` (reused as-is; it snapshots history,
     re-inserts all rows, writes content_data back). The guard passes (copy preserved).
  3. Source FRESH resolved_data — SETTLED: route (ii). plan_sections' resolution is a
     separable struct `sourceResolver` (plan_sections_action.go): `newSourceResolver(siteID,
     db, logger, pageName)` + `resolve(ctx, source)`, where `ensureAssets` does the page-aware
     `site_plan_imagery`+`assets` hero join (the join that picks up the landed image) and
     query.* delegates to `queryresolve`. It's unexported but in the `actions` package, so the
     new action reuses it directly — no plan_sections re-run, no side effects, no plan↔stored
     matching. Per section: load the component input_schema (reuse `loadComponentSchemas` /
     `loadSingleComponentSchema`, also in-package), iterate fields with a `source`, call
     `resolver.resolve(source)` → fresh resolved_data; overlay onto stored content_data;
     render. (Route i — re-run plan_sections + match — rejected: side effects + matching.)
  3a. Rebuild render_context via `BuildRenderContextAction` (v3_site_actions.go:739) for the
     site-level base (company/colours/contact) the template needs.
  4. Slot `rerender_page_sections` into the page-rerender agent workflow as a pre-pass gated
     by spec.reason (image_landed / section_data_resolved); then the existing
     rerender_single_page assembles + deploys as today.
  5. Repoint creators: `flag_page_image_rebuild` and `reconcile_section_data` emit
     item_type `page_rerender`, handler page-rerender, key `page_rerender:<page>` (now the
     prefix matches the type). They co-dedup with each other (Bug 1 for these two closes).

  STATUS: investigation complete, all choices made (Y; re-render-all; NULL→writer; resolved_data
  route ii via reused sourceResolver); nothing shipped. Next action = write the
  `rerender_page_sections` action (load page_components → NULL pre-check → per-section reuse
  sourceResolver + render via RenderTemplate → emit sections_metadata) + thin page-rerender
  wiring (rerender_page_sections → save_page_sections → assemble → deploy) + the two creator
  repoints, as one change. Complexity in Go, workflow stays thin.

  BUILD TRACE FINDINGS (2026-06-20, from v3_site_actions.go + plan_sections + save):
  - GOOD: `planSection` (plan_sections_action.go:952) is reusable AND side-effect-free — the
    needs_new_component / deferred-item writes are in its caller (PlanSectionsAction lines
    653, 705), not in planSection. So calling planSection per section rebuilds resolved_data
    (query/asset/spec resolution, LLM fields skipped) with no side effects. Removes the
    field-loop duplication; route (ii) is clean.
  - GOOD: sections_metadata contract confirmed — each entry {rendered_html, component_id,
    component_name, component_function, content_data} (CompilePageSectionsAction +
    extractSectionFromMap, v3_site_actions.go:1847). `save_page_sections` reads exactly these
    via `sections_metadata_field`, so the new action emits that shape and feeds save directly —
    NO compile_page_sections step needed. save resolves page identity from page_name_field +
    site_id_field (derives page_id).
  - COMPLICATION (structural, flagged before building): the page-rerender path has NO
    render_context. RenderComponentAction renders against a RenderContext built from
    content_from + merge_with ON TOP OF context_from. Stored content_data carries the section's
    copy + resolved fields but NOT the ambient base (company/contact/year/inline colours) —
    RenderComponentAction only captures content_from+merge_with into content_data, not the
    context base. BuildRenderContextAction builds that base from reviewed_brief/style_collection/
    site_plan in collected_data, which the WRITER pipeline has loaded but page-rerender has not.
    → A section template referencing an ambient field would re-render blank unless page-rerender
    reproduces the writer's context setup.
    FORK inside Y: (Y-lean) new action rebuilds a minimal render_context from the DB directly
    (sites.content_data for company/contact, style collection for colours) — light, but any
    ambient field a template uses that we don't rebuild goes blank; vs (Y-full) page-rerender
    gains the writer's context setup (load specs/brief/style → build_render_context) ahead of
    the re-render — correct but heavier (page-rerender becomes "build pipeline minus LLM").
    NOTE: this is the one place Option X (no-LLM branch in the writer) would be simpler — it
    inherits the pipeline's already-built render_context. Not reversing Y; recording the cost.
    BLOCKER for lean-vs-full: how heavily the real section templates use ambient render_context
    fields. Settle via a couple of content_components.html_template rows (e.g. hero, tool-list)
    OR page-build-handler's call_content_writer input_mapping (what context the writer is handed).
    Then write the action against the real answer.

  RESOLUTION (2026-06-20, from hero + tool-list html_template + page-build-handler config):
  - COMPLICATION DISSOLVED → Y-LEAN CHOSEN. Neither template uses an ambient render_context
    field. hero: headline/subheadline/cta_*/hero_url all from content_data; colours are CSS
    custom properties with inline fallbacks (`var(--primary-color, #1a1a2e)`) resolved by the
    browser against the stylesheet, NOT interpolated. tool-list: eyebrow/heading/intro/cta from
    content_data, `.items` from the resolver, colours via `var(--color-*)`. No company_name,
    year, or contact referenced. page-build-handler confirms the writer's render_context is
    built only from site_record + reviewed_brief (both in site_record.content_data) +
    page_record — it does NOT pass style_collection. So Y-lean: rebuild a minimal render_context
    from site_record.content_data + the page row (≤2 queries) — reproduces what the writer had,
    and for these templates is barely used. Y-full (writer's full context setup) NOT needed.
  - NEW COMPLICATION (wiring): page_id surfacing on the work-item path. page-build-handler
    invokes page-rerender AS A SUB-AGENT via call_agent with input_mapping
    {domain, page_id: page_record.id, site_id} — so that path hands page_id directly. But the
    repointed flag/reconcile items dispatch page-rerender via WORK ITEMS (page_rerender type),
    whose spec carries page_name only. render_page (rerender_single_page) needs page_id (it
    derives domain + site_id from the pages⋈sites join in getPageInfo). So the work-item entry
    must make page_id available, which the sub-agent entry gets free.
    CHOICE: flag/reconcile resolve page_id at emit (site_id + page_name → pages.id) and set the
    work item's page_id COLUMN — idiomatic, exactly what check_sectionless_pages does — AND the
    new rerender_page_sections step resolves page_id from spec.page_name + site_id regardless
    (robust to column surfacing). save_sections wiring mirrors page-build-handler:
    page_name_field = input_data.spec.page_name, sections_metadata_field = <rerender output>.
    sections_metadata, site_id_field resolved in-step.
    VERIFY before wiring: how the build-dispatch-loop surfaces a directly-claimed item's page_id
    column / spec into render_page's input_fields (datahelpers.ExtractFields search semantics —
    top-level vs input_data vs input_data.spec). This is the last wiring unknown; the resolve→
    render→persist→assemble spine is otherwise fully settled and reuses existing code.
  STATUS: all design choices now made (Y; re-render-all; NULL→writer; route ii; Y-lean
  render_context; page_id via emit-time column + in-step resolve). One wiring fact to confirm
  (ExtractFields surfacing), then write `rerender_page_sections` + page-rerender wiring +
  flag/reconcile repoints (now also setting page_id column) as one change.

  BUILD COMPLETE (2026-06-20): page_id surfacing RESOLVED — ExtractFields' extractSingleField
  has Strategy 4 aggressive recursive search (unified_extractor.go:442), so render_page finds
  page_id from the rerender step's output; rerender_single_page derives domain+site_id from
  page_id via its pages⋈sites join. No page_id column needed on the items after all.
  - ACTION WRITTEN + VERIFIED against source: rerender_page_sections_action.go (in outputs).
    Reuses planSection (side-effect-free) + loadComponentSchemas for resolution; layers
    mergeIntoRenderContext (site base → stored content_data → fresh resolved_data, resolved
    wins); renders via RenderTemplate(comp.Raw["html_template"]); emits the exact
    sections_metadata shape (extractSectionFromMap, v3_site_actions.go:1847) that
    save_page_sections reads; escalates NULL-content_data pages to needs_page:<page>; carries
    deferred/skipped/template-missing sections untouched; outputs page_id/site_id/domain.
    Note: template check confirmed Y-lean is safe — hero + tool-list use only content_data +
    CSS-var colours, no ambient render_context fields.
  - WIRING (page-rerender default_config): start_step → check_rerender_mode (conditional on
    spec.reason in {image_landed, section_data_resolved}); rerender_sections
    (rerender_page_sections, config target_site_id="input_data.site_id",
    page_name="input_data.spec.page_name", reason="input_data.spec.reason"; output_field
    rerender_sections); check_escalated (escalated→complete, else→save_sections);
    save_sections (save_page_sections, site_id_field rerender_sections.site_id, page_name_field
    input_data.spec.page_name, sections_metadata_field rerender_sections.sections_metadata) →
    render_page. else_step→render_page keeps plain page_rerender items unchanged.
  - REPOINTS (2 lines each): flag_page_image_rebuild + reconcile_section_data change
    itemType needs_page→page_rerender and handlerAgent page-build-handler→page-rerender; key
    page_rerender:<page> and spec.reason already correct. They co-dedup (Bug 1 for these two
    closes). Update their now-stale header comments (cosmetic).
  - REGISTRY: register "rerender_page_sections" (Handler RerenderPageSectionsAction, Category
    site, IsLocal true).
  - GUIDELINE AUDIT (2026-06-20, against 001/003): PASS with one fix applied — input renamed
    site_id → target_site_id per 001 §Field name collisions (site_id is a nested-source key;
    same precedent as reconcile_site_plan); wiring maps target_site_id="input_data.site_id".
    Confirmed: no logger.Debug; SQL parameterised (escalation spec built with %q, passed as a
    value to insertWorkItem); reuses planSection/loadComponentSchemas/mergeIntoRenderContext/
    RenderTemplate (no recreation); single action, complexity in Go; no SQL subworkflow.
  - DOCS UPDATED (in outputs): 002 — page-rerender row now describes the re-render path + new
    "Work-item routing: content rebuild vs re-render" subsection (item_type→handler table,
    page_rerender semantics, flag/reconcile emit page_rerender, NULL→needs_page escalation).
    003 — data-flow block notes the re-render path; "Source of truth principle" documents the
    two re-render paths (full writer vs rerender_page_sections) and content_data = stored ⊕
    fresh-resolved. 001 needed no change (its field-collision guidance was the audit basis).
  - DEPLOY NOTE: re-render-all means a trigger on a page with any NULL-content_data section
    escalates the WHOLE page to a full writer rebuild (one-time, self-healing) — expected.
  STATUS 2026-06-21: Part A APPLIED + DEPLOYING. The three code edits are in (new action +
  registry registration + flag/reconcile repoints) and the page-rerender default_config UPDATE
  (start_step → check_rerender_mode → rerender_sections → check_escalated → save_sections →
  render_page) is applied; new chassis image building, image_tag bumped on page-rerender +
  image-build-handler. Verification follows runbook Part 2 — P2.1 FIRST (confirm a claimed
  page_rerender item surfaces site_id at input_data.site_id; if rerender_sections errors
  `missing required fields: [target_site_id]` the dispatch puts site_id elsewhere and the config
  path needs adjusting). Then P2.3 Test 1 = direct page_rerender insert proves the no-LLM
  re-render: item complete, copy (content_data.headline) unchanged, page_components.updated_at
  advances, NO page-content-writer orchestration, NO execute_llm_prompt, NO regression block.
  P2.4–P2.7 = real image-landed flow, section-data, NULL→needs_page escalation, plain
  page_rerender backward-compat.
  REMAINING after verify: Part B (key canonicalization) and the 002/003 doc flips (item_key now
  encodes type for the flag/reconcile paths).

  UPDATE 2026-06-21 (deploy in, verifying):
   - Part 1 re-confirmed healthy in prod: index rebuilt+deployed 2026-06-19 10:34–10:35, all 5
     page_components fresh, stub_rows=0, no CHILD_ORCHESTRATION_FAILED / delivery-cap errors.
     (Telemetry gap, non-blocking: page_components.deploy_commit blank and pages.last_built_at
     NULL though deployed_at is set — deploy step isn't writing those back; fold into the
     deploy-observability fix later.)
   - Part 2 workflow UPDATE confirmed: start_step=check_rerender_mode, rerender_sections step
     present.
   - P2.1 site_id pre-flight returned 0 ROWS — EXPECTED, not a missing-site_id signal: no
     page_rerender item has run since the deploy, so there's no page-rerender orchestration in
     the 15-min window to inspect. It gets answered by Test 1's own orchestration. (Per the
     don't-treat-0-rows-as-decisive rule: the query found nothing because nothing has run, not
     because input_data.site_id is null.)
   - Baseline captured: hero headline present ("Your Probability Maths Is Wrong"), hero_url
     EMPTY (hero uses the static /assets/images/hero.jpg fallback), list sections have no
     'headline' key — all expected.
   - PRE-TEST-1 CHECK still needed: confirm content_data is non-NULL for all 5 sections (a
     null on any section makes Test 1 ESCALATE to needs_page instead of re-rendering). Expected
     non-null since the writer rebuilt the page 06-19.
   - OBSERVATION (separate from the re-render path, confirm don't assume): the deployed
     index.html has 4 content sections (hero, tool-list, guide-list, game-list) but
     page_components lists 5 — system-stats (position 5) is absent from the live markup. Could
     be a stale paste or the assemble/deploy step dropping the trailing section. Test 1
     re-renders all 5 + re-assembles, so a clean Test 1 should bring system-stats back if it's
     genuinely missing; if it doesn't, the assemble (render_page/compile) step is where to look.
   - NEXT: run Test 1 (P2.3 direct page_rerender insert).

**Part B — key canonicalization (after Part A).** Canonical builders in
`work_items_common.go` (prefix == item_type); repoint adoption tool items →
`needs_tool_recreation:<name>` (fixes Bug 2), adoption content → `needs_content_page:<name>`;
flag/section-data keys follow Part A's type decision. Decisions needed: should
`needs_page` and `needs_content_page` (two types, same handler, same work) share a
dedup namespace or unify; should adoption-recreate and reconcile-build of the same
page dedup or stay distinct intents.

RUNBOOK: Part 3 (survey → decide → apply → verify) is written. The P3.2 survey is runnable
NOW, before any code — it confirms Bug 2 is live (needs_content_page + needs_tool_recreation
both keyed needs_page:<name>) and captures the prefix-mismatch baseline (split_part(item_key,
':',1) vs item_type). The two decisions above (P3.1) gate the build; the apply steps assume the
conservative default (keep types separate, prefix == type, distinct intents), which is
forward-compatible with either decision and is the minimal Bug-2 fix. Do AFTER Part A verifies
(don't stack unverified changes). Bug 1 is already closed by Part A (flag/reconcile now emit
page_rerender-type keyed page_rerender:<page>) — Part B only verifies it (runbook B4).

**Part B — files received + code prepared (2026-06-21).** `apply_adoption_plan_action.go`
and `work_items_common.go` now in hand. Findings:
  - `work_items_common.go` is the shared `site_work_items` helpers file (already holds
    `workItemTerminalStatuses` + `sqlInList`) — the right home for the key-builder. NOTE:
    `insertWorkItem` / the `workItem` struct are NOT in it (they live in another file in the
    `actions` package), and `workItem` has NO page_id field — which is why adoption's
    page-item INSERT is raw `tx.ExecContext` (it needs the page_id column), not
    `insertWorkItem`. The builder must therefore be a plain `func(...) string` usable from
    both call styles.
  - BUILDER PREPARED (in outputs/work_items_common.go): `func workItemKey(itemType, target
    string) string` → `itemType+":"+target`. Unexported, matching the file's convention
    (`sqlInList`). Confirm no existing `workItemKey` before adding (grep) per the reuse rule.
  - `apply_adoption_plan_action.go` work-item INSERT (lines 642–656): `item_key` is hardcoded
    `fmt.Sprintf("needs_page:%s", page.Name)` for BOTH `needs_tool_recreation` (tool branch,
    line 627) AND `needs_content_page` (content branch, line 634). The TOOL item sharing the
    content key is Bug 2 (different handler, same key → idx_swi_dedup silently drops one on a
    name overlap).
  - NEW EVIDENCE on the content key (lines 653–654 comment): the `needs_page:<name>` key was a
    DELIBERATE doc-029 Phase-0 choice, so the adoption content item co-dedups with
    planner-emitted `needs_page` builds of the same page. So decision 1 (share the
    needs_page / needs_content_page namespace?) is NOT a clean prefix==type call — the code
    already shares the namespace on purpose, and it is defensible (same handler, same work).
    This REVISES the earlier "content → needs_content_page:<name>" recommendation.
  - SETTLED (Bug 2 fix, decision-independent): tool item key →
    `workItemKey("needs_tool_recreation", page.Name)`. Stops the collision under any decision.
  - OPEN (content key — user's call):
      * Option B (now recommended): keep content in the needs_page namespace —
        `workItemKey("needs_page", page.Name)` — preserving the deliberate planner co-dedup.
        The 002/003 doc flip then reads "prefix == item_type EXCEPT needs_content_page shares
        the needs_page dedup namespace (same handler / same work)."
      * Option A: content → `workItemKey("needs_content_page", page.Name)` for a clean
        prefix==type everywhere — BUT then planner/reconcile must also emit needs_content_page
        (or the types unify) or adoption-recreate and planner-build of one page double-build.
        Bigger change (touches reconcile_site_plan + the planner).
  STATUS: builder written; tool repoint settled; content-key decision pending; nothing applied.
  After the content-key call: finalise the adoption INSERT diff (an `itemKey` var per branch,
  swap line 655), fold the resolution into runbook P3.1/P3.3 + the doc-flip, build, then apply
  AFTER Part A verifies (don't stack unverified changes).

**Part B — content key DECIDED: Option B (2026-06-21).** Content item stays in the
`needs_page` namespace (`workItemKey("needs_page", page.Name)`), preserving the doc-029
co-dedup with planner `needs_page` builds; only the tool item moves to its own namespace
(`workItemKey("needs_tool_recreation", page.Name)`), which is the actual Bug 2 fix. This also
settles decision 2 — adoption-recreate and planner/reconcile-build of one page DO co-dedup
(one open build serves both; page-build-handler rebuilds from current state regardless of
trigger). Known consequence: adoption sets `spec.mode="recreate"` while a planner build may
not, so whichever row wins `ON CONFLICT DO NOTHING` fixes the mode — accepted. The
prefix==item_type invariant now carries one documented exception (needs_content_page ⊂
needs_page namespace), which is how the 002/003 doc flip is worded. Adoption INSERT diff
finalised (var itemKey per branch + line-655 swap); runbook P3.1/P3.3/P3.5/P3.6 updated.
Remaining: apply builder + adoption diff, `go build ./...`, deploy (bump image_tag on the
adoption / site-adoption handler), run P3.4 — AFTER Part A verifies.

**Session 2026-06-22 — index Test 1 PASS; pathfinding game-drop investigation.**

INDEX Test 1 (page_rerender, reason=image_landed) — PASS. All 5 sections re-rendered:
updated_at advanced 2026-06-19 10:34:50 → 2026-06-22 10:24:26 on every section, which only
happens via rerender_sections → save_sections (assemble-only render_page does NOT rewrite
page_components). So the path executed correctly:
  - check_rerender_mode matched reason='image_landed' and routed to rerender_sections (not the
    else/assemble branch).
  - rerender_sections RAN → target_site_id resolved from input_data.site_id. **P2.1 CONFIRMED**
    (no missing-target_site_id error; the load-bearing wiring path is correct).
  - escalated=false (content_data non-null on all 5 — verified by the cd_preview check),
    save_sections persisted, render_page assembled+deployed, item complete, no error.
  - Copy preserved: hero headline identical ("Your Probability Maths Is Wrong").
  - Only the hero's rendered_html changed: html_len 2288 → 2364 (+76); the other four are
    byte-identical (8964/7476/8151/7369). Identical list sections = resolved data unchanged
    since 06-19 (re-rendered to the same bytes, or carried — both correct). The hero +76 is
    consistent with a resolved field refreshing, most likely the CTA now rendering (hero
    content_data has cta_text="Open Drop Rate Simulator"; the template emits the <a> only when
    cta_url also resolves). CONFIRM on the live hero whether the button now shows (ties to the
    earlier "index hero has empty/no buttons" note).
  - The result columns escalated/rerendered/carried were null because the work item stores the
    FINAL step's (render_page) output, not rerender_sections' — the section updated_at advance
    is the ground truth, not those columns. (Per-section rendered-vs-carried, if wanted, is in
    the orchestration's rerender_sections output / page-rerender logs.)
  Still optional/pending for Part 2: P2.4 (real image-landed flow), P2.5 (section_data_resolved),
  P2.6 (NULL content_data escalation), P2.7 (plain page_rerender backward-compat).

SYSTEM-STATS observation (from 06-21) — likely resolved. system-stats IS present in
page_components (position 5, re-rendered in Test 1). The earlier 4-section index HTML was a
stale capture; after Test 1 re-assembled+deployed, the live page should show all 5. CONFIRM the
live /index.html now includes the system-stats section.

PATHFINDING (game-pathfinding) — the game is MISSING from the page; a re-render is a dead end
here and the cause is a separate structural issue. Established facts:
  - Page 56af8679-1f7d-4da6-b148-f5727b16693d, /games/pathfinding/index.html, deployed
    2026-06-14 20:08:28; both sections last written 2026-06-14 20:07:52.
  - Only 2 sections: hero (content_components 23f95f00, hero template) + generic-text-block
    (content_components 8d81e665, the FALLBACK "Generic Text Block" for "any unmatched section").
    NO interactive game component. The "How A* Actually Works" copy is in the generic fallback,
    not a game. → page_rerender re-renders these two and will NOT add a game; the two prior
    page_rerenders (06-06, 06-13) did exactly that. Do NOT queue another.
  - The interactive build (needs_tool_recreation → tool-recreation-handler) completed 2026-06-05
    17:33 — BEFORE the 06-14 rewrite — and was mis-keyed `needs_page:game-pathfinding` (a live
    instance of Part B / Bug 2).
  - A link_resolution_rebuild for game-pathfinding (item 46c2da91, handler page-build-handler)
    completed 2026-06-14 20:08:37, coincident with the 20:07:52 section rewrite (sibling
    game-auto-battler completed 20:03:41). Both carry error "Claim timed out — handler pod likely
    died" yet status complete (itself suspect — a timed-out claim marked complete).
  HYPOTHESIS (NOT proven — confirm before any fix): the 06-14 link_resolution_rebuild routed
  through page-build-handler (the full content builder) regenerated the page as a standard
  content layout (hero + generic-text) and dropped the interactive game the 06-05 tool-recreation
  had attached. A links-only maintenance task went through the content builder and clobbered an
  interactive page; the generic-FALLBACK block is the tell that the rebuild couldn't match the
  game section. Distinct from the item_key mis-key, though both stem from "interactive page
  treated as a content page."
  NEXT DIAGNOSTICS (confirm, don't assume):
   1. page_component_history — did a game/interactive section exist before 06-14 and get replaced?
        SELECT position, slot_name, component_id, build_status, length(rendered_html) AS html_len,
               created_at, updated_at
        FROM page_component_history
        WHERE page_id = '56af8679-1f7d-4da6-b148-f5727b16693d' ORDER BY updated_at;
   2. orphan check — does a pathfinding game component still exist in content_components?
        SELECT id, name, function, section_type, render_mode, length(js_content) AS js_len,
               component_level, is_active, updated_at
        FROM content_components
        WHERE name ILIKE '%pathfind%' OR function ILIKE '%pathfind%' OR name ILIKE '%astar%'
           OR name ILIKE '%a-star%' OR description ILIKE '%pathfind%' ORDER BY updated_at DESC;
   3. what page-build-handler does on item_type=link_resolution_rebuild (full rebuild vs
      links-only) — its default_config + the 46c2da91 orchestration.
  REMEDIATION (TBD): re-running the tool-recreation regenerates+reattaches the game, but if a
  content/link rebuild dropped it, it'll be dropped again next time one fires for that page. The
  structural fix is to stop links-only (and content) rebuilds destroying interactive sections —
  either page-build-handler preserves existing interactive components on rebuild, or
  link_resolution_rebuild doesn't route through the full content builder. Decide after 1–3.
  Tracked as candidate "Part 4" in the runbook.

**Session 2026-06-22 (cont.) — pathfinding root cause CONFIRMED + blast radius.**
The earlier clobber hypothesis is confirmed, but the trigger was NOT the 06-05 tool-recreation
being overwritten by a later content rebuild — it was a same-day overwrite, and my mid-investigation
inference that ruled out item 46c2da91 on a timestamp was WRONG. Corrected chain:
  - page_component_history: at 2026-06-14 19:21:21 the hero held the FULL A* game (18,449 bytes:
    <canvas id="gridCanvas">, Wall/Mud/Eraser brushes, Run/Reset, tool-page markup). Real and working.
  - At 20:07:52 it was overwritten with hero + generic-text-block (source=save_page_sections_overwrite,
    source_item_id NULL).
  - WRITER = the pathfinding link_resolution_rebuild (46c2da91): page-build-handler orch 00615292
    (20:04→20:08:36) spawned page-content-writer 393df4bf (20:04→20:07:46) → gameless page_html →
    page-rerender 4b820f4b assembled+deployed. The work item's completed_at (20:08:37) is the orch's
    END; the destructive write happened mid-run in the child writer (20:07:46–52). I had excluded
    46c2da91 as "too late" — wrong; trace the orchestration, not the work-item timestamp.
  - ROOT CAUSE: link_resolution_rebuild (spec literally "preserve the existing copy; re-resolve
    internal links") is handled by page-build-handler, which runs the full page-content-writer.
    The writer regenerates from plan_sections; the plan has no knowledge of the tool; the interactive
    component is discarded. Same anti-pattern as Part 2 (a maintenance task going through the full
    content rebuild). NOT primarily the item_key mis-key, though that's present too.

BLAST RADIUS — one page. Sweep of all game-/tool- pages for <canvas>/game-container/tool-page markup:
ONLY game-pathfinding lost interactivity (has_interactive_markup=f, 2 sections). All others retain it.
game-auto-battler was in the SAME linking batch but survived — its deployed_at stayed 06-13, so its
link rebuild never re-deployed; pathfinding's ran the full writer-to-deploy chain. Survivors remain at
risk until the routing fix lands. game-jelly-invaders is separately needs_rebuild/undeployed (not a
casualty).

page-build-handler structure (confirmed): single linear flow ensure_site_record → … → plan_sections →
call/spawn_content_writer → save_sections → deploy_page, NO branch on item_type. Has load_existing_content
+ check_has_ready_sections/check_content_produced conditionals → a preserve-existing path MAY exist but
isn't taken for link_resolution_rebuild. NEXT DIAGNOSTIC: dump those step configs (action/condition/
then_step/else_step) to see if a links-only/preserve branch is bypassed or the writer always runs.

CONFIRMED STRUCTURAL FIXES (record now, regardless of routing decision):
  1. page_component_history overwrite rows carry NULL source_item_id → can't attribute a destructive
     write by id (had to trace by time+orchestration). save path should stamp the driving work item.
  2. Part-1 text-loss regression guard (>75% text drop) did NOT catch the 18KB-game→2KB-text swap
     because the loss is markup/JS, not prose. Need an interactivity-aware guard: an overwrite dropping
     a <canvas>/data-component game from a section should block or escalate, not save silently.
FIX DIRECTION (reuses Part 2): route link maintenance through a re-render-class path that re-resolves
link fields and PRESERVES sections, not the full writer. Open question (the step-config dump informs it):
are internal links resolved fields (repoint directly, like Part 2) or baked into LLM prose (needs a
dedicated links-only rewrite that still preserves interactive sections)?
REMEDIATION: re-run tool-recreation for game-pathfinding AFTER the routing fix (so it can't re-clobber).
No imminent linking batch → not urgent. Runbook Part 4 updated to "cause confirmed".

**Session 2026-06-22 (cont.) — page-build-handler step configs: NO item_type branch.** Dumped
the step configs. Every `next_step` is unconditional; the only conditionals are
`check_page_found`, `check_has_ready_sections` (`section_plan.ready_count > 0`),
`check_content_produced` (`page_content.response.skipped != true`) — none keyed on item type.
So `link_resolution_rebuild` runs the IDENTICAL full path as a fresh build
(`load_existing_content → load_spec_sections (load_page_sections_from_spec) → plan_sections →
spawn/call_content_writer → validate → save_sections → deploy`). Not a bypassed preserve-branch
— there is none → the fix is structural (routing), not a conditional fix. WHY THE TOOL DIES
PRECISELY: the rebuild plan is built from the page SPEC; the tool was attached as a section's
rendered_html but isn't in the spec, so plan_sections omits it, the writer produces only the
planned sections, and save_page_sections (DELETE+INSERT of the produced set) drops it.
`load_existing_content` runs but unconditionally proceeds to load_spec_sections — it does NOT
prevent regeneration (red herring as wired; worth knowing what it does, but not the lever).
FIX FORK (routing): link_resolution_rebuild should not run the writer. (a) if internal links
are resolved at render → route it to the Part 2 page_rerender path (preserves the tool +
re-resolves fields, reuses shipped machinery); (b) if links are baked into prose → a links-only
rewrite over stored content that preserves sections — and note there is ALREADY an
internal-link-resolver / resolve_internal_links_action.go (016 §v2_45), so the question may be
why the §5 batch routed these to page-build-handler instead of that resolver. NEXT DIAGNOSTIC:
determine link representation — inspect a content section's stored rendered_html/content_data
(literal `<a href>` vs resolved token) + read resolve_internal_links_action.go and the §5
batch creator (where handler_agent is set). The two guards (source_item_id stamping;
interactivity-aware save guard) stand regardless. Runbook + 016b Part 4 updated.

**Session 2026-06-22 (cont.) — P4.1 result: links are build-resolved CTAs; fix re-framed (layered).**
Ran the link-representation diagnostic + read resolve_internal_links_action.go + check_phantom_internal_links.go.
FINDINGS: (1) On the §5 pages, content_data has NO `<a>`; the only internal link is the hero CTA
in rendered_html — guide-economy-basics hero `cta_url = /contact.html` (a PHANTOM; real contact
is /contact/index.html). (2) resolve_internal_links is a BUILD-TIME augmenter: it writes
cta_url/secondary_cta_url/primary_cta_url into section.resolved_data (merged into stored
content_data, then rendered) for hero/call-to-action components, pointing at section-index hubs
by nav_order (excluding about/contact/legal); caller is page-content-writer; it is explicitly
"NOT a rendered-HTML patcher" and emits unresolved_cta HITL items. (3) check_phantom_internal_links
deliberately routes page_component link findings → page-build-handler (rebuild), site_component →
nav-link-fixer — precisely BECAUSE the resolver has no in-place surface. So the §5 routing to
page-build-handler was by design.
TWO EARLIER ASSUMPTIONS CORRECTED:
  - A bare page_rerender would NOT fix these links: the build-resolved cta_url is persisted in
    content_data and the re-render re-resolves only query/hero-image via planSection (not CTAs),
    so it reproduces the phantom. (Confirm: SELECT content_data->>'cta_url' FROM page_components
    WHERE page_id='8ed97fd2-…' AND slot_name='hero'; → expect /contact.html.) So my earlier
    "repoint link_resolution_rebuild → page_rerender" is DROPPED as insufficient.
  - The clobber is NOT specific to link_resolution_rebuild: the tool's interactive content is a
    page_component section that isn't in the page SPEC, so ANY full rebuild (needs_page /
    needs_content_page / content_rewrite / admin Regenerate Page / link rebuild) plans from spec,
    omits the tool, and the DELETE+INSERT drops it.
FIX (re-framed, layered, structural first): (1) interactivity-aware guard in save_page_sections —
block/escalate replacing a `<canvas>`/data-component section with non-interactive content (cheap,
immediate, all triggers); (2) preserve interactive sections through a rebuild — carry forward
existing non-spec interactive/tool sections (extend load_existing_content/save to merge, not
DELETE-all), or represent the tool in the spec as non-regenerated (the structural root; keeps CTA
re-resolution working via the writer, no routing change); (3) optional later — extend page_rerender
to re-run resolve_internal_links hub logic for CTA fields and route link_resolution_rebuild there
(no-LLM link fix). Order: 1 → 2 → 3.
OPEN SUB-QUESTIONS (non-blocking): (a) does save_page_sections honor page_components.locked_at? if
a rebuild carries forward locked sections, locking the tool section is a zero-code interim
mitigation — 013 says execution agents process explicit items regardless of lock, so probably not,
but confirm. (b) /contact.html origin — resolver excludes contact from CTA targets, so it didn't
set that; likely a hardcoded hero default or pre-resolver value (hero CTA may not be going through
the resolver on these pages). (c) stale duplicate z_context/check_phantom_internal_links.go routes
page-link fixes to internal-link-resolver (live platform copy routes to page-build-handler) —
delete it. Runbook Part 4 + 016b updated; canonical 002/005/020/026 hazard notes still say "route
link maintenance through a preserve-sections re-render path" — substance (preserve sections, guard,
source_item_id) holds, but the routing nuance is now "preserve interactive sections IN the rebuild";
left as-is, refine on the way in.

**Session 2026-06-22 (cont.) — cta_url confirm + refinement (page_rerender viability re-opened).**
Confirmed guide-economy-basics hero `cta_url = /contact.html` is stored in content_data. BUT the
resolver EXCLUDES contact, so /contact.html is NOT a resolver hub — it's the hero component's
SCHEMA field (cta_url source `pages.contact`), resolved when contact was at /contact.html and now
stale (contact moved to /contact/index.html). So the CTA is schema-sourced, not resolver-set. This
RE-OPENS what I over-closed last turn: a `page_rerender` MIGHT fix it if the Part-2 re-render
re-resolves schema-sourced fields (re-resolving a moved page ref is the point of that path). So
"repoint link_resolution_rebuild → page_rerender" is DEFERRED, not dropped — if page_rerender
re-resolves the CTA it's the clean no-LLM link fix. The §5 batch chose page-build-handler only
because it predates Part 2 (06-21), so not evidence. DECISIVE TEST (P4.2 in runbook): re-render
guide-economy-basics (a NON-tool page, safe) via a page_rerender (reason=image_landed, mirror the
P2.3 Test 1 insert) and see if cta_url flips to /contact/index.html. Or read
rerender_page_sections_action.go (does its field re-resolution run the standard source resolver
incl. pages.* sources, or only query+hero-image?). Layered fix (Layer 1 guard + Layer 2
preserve-sections) unchanged as the baseline regardless. I've now been wrong on page_rerender in
both directions — hence test, don't assert.

**Session 2026-06-22 (cont.) — P4.2 RESULT: page_rerender preserves sections, does NOT re-resolve CTA.**
Ran the controlled re-render on guide-economy-basics (item 398dedc2, complete, no error). Both sections
re-rendered (updated_at → 17:08:23), both intact → page_rerender CONFIRMED non-destructive. But cta_url
stayed /contact.html (no flip to /contact/index.html) → page_rerender does NOT re-resolve the
schema-sourced hero CTA. So routing link_resolution_rebuild → page_rerender is RULED OUT as a link fix
(preserves tool, reproduces phantom). My two-turns-ago read was right; last turn's hedge was unnecessary
— testing beat asserting. CLOBBER FIX is firmly Layer 1 (guard) + Layer 2 (preserve interactive sections
in the full rebuild), keeping link_resolution_rebuild on page-build-handler. Layer 3 (repoint) dead.
SEPARATE FINDING (confirm, don't conclude): guide-economy-basics had a §5 full rebuild (508c0ccb, 06-14)
and STILL shows /contact.html — so the phantom may survive a FULL rebuild too, meaning the §5 batch's
goal (fix phantom hero CTAs) is unmet by EITHER path. Distinct phantom-CTA-resolution bug: hero cta_url
source pages.contact appears to produce /contact.html (wrong), not the contact page's real
/contact/index.html. NEXT DIAGNOSTIC: page_component_history for this hero (what cta_url did the 06-14
rebuild store?) + read the pages.<area> field resolver (reads pages.url, or constructs /{area}.html?) +
whether resolve_internal_links runs on these heroes (should set a hub, not contact). Layers 1/2 do NOT
fix this — they make the rebuild safe; the phantom CTA needs its own fix. Runbook P4.2 + run-order +
016b updated.

**Session 2026-06-22 (cont.) — phantom hero CTAs CONFIRMED in deployed HTML; rebuild inference retracted.**
page_component_history for guide-economy-basics shows ONLY the P4.2 write (06-22), NO 06-14 row —
pathfinding had 06-14 rows, so history was capturing then. So I RETRACT "the phantom survives a full
rebuild": the §5 rebuild likely didn't re-save this hero; the phantom is probably an earlier build's
value no rebuild corrected. The deployed HTML pins the real bug: the hero has TWO phantom CTAs —
primary cta_url=/contact.html ("Read the Full Guide"), secondary secondary_cta_url=/services.html
("Browse Economy Guides"). Real contact is /contact/index.html; NO services page exists. The page's
own FOOTER links /contact/index.html correctly → the bug is the hero CTA resolution specifically. These
map to the hero schema sources (cta_url ← pages.contact, secondary_cta_url ← pages.services), which
appear to produce constructed /{area}.html (wrong for contact, fabricated for non-existent services).
resolve_internal_links (intent-aware hubs) is NOT overriding them. Structural point (on the site-goal):
the generic hero schema assumes a brochure site (contact/services CTAs); resolve_internal_links exists
to replace those with real content hubs, and it isn't winning. TWO-PART BUG: (A) pages.<area> source
resolver produces wrong/fabricated URLs; (B) resolve_internal_links not running/winning on these heroes
— (B) is the structural fix. NEXT DIAGNOSTICS (read-only): unresolved_cta items for this site? +
section-index hubs exist (page_type='section-index')? Then read the pages.<area> field resolver +
page-content-writer's resolve_internal_links wiring/precedence. Separate from the clobber (Part 4
Layer 1+2 unchanged); given the goal, phantom CTAs likely take priority (live, visible, on-goal).

**Session 2026-06-22 (cont.) — section-index hubs EXIST; phantom-CTA narrowed to wiring/precedence.**
Hubs: games-index (/games/index.html, nav 4), about (5), contact (6), tools (100), guides (100), all
active. Through resolve_internal_links' chooseCTATargets for a guide: exclude about/contact → games/
tools/guides → primary=games-index (/games/index.html), secondary=guides-index (/guides/index.html).
So if the resolver ran/won, the hero CTAs would be those REAL links — not the deployed /contact.html
+ /services.html (the schema brochure defaults). So the resolver is NOT failing for lack of hubs; its
output isn't reaching the hero. (unresolved_cta query result was GARBLED — a shell grep got pasted
into psql, -# continuation — not captured; but with hubs present the resolver wouldn't emit
unresolved_cta anyway, so re-running is low-value. Per the don't-trust-0-rows rule I'm not treating
the missing output as zero.) NARROWED to wiring/precedence (code question): is resolve_internal_links
called by page-content-writer, and does the generic schema field resolution (cta_url←pages.contact,
secondary←pages.services) overwrite it? NEXT: (a) SQL — does page-content-writer's default_config
reference resolve_internal_links (ILIKE) + dump top-level steps (section processing is a nested loop
sub_workflow per doc 005); (b) if referenced, drill the sub_workflow step order + read render_component
/ the field resolver for precedence (does resolved_data survive or does schema-source overwrite
cta_url?). Structural fix REUSES the existing resolver (it computes correct links); make it run+win,
not new code. Runbook phantom-CTA thread updated.

**Session 2026-06-22 (cont.) — resolver IS wired (spawned sub-agent); narrowed to output-threading/precedence.**
ILIKE '%resolve_internal_links%' on page-content-writer = FALSE, but that's a wrong probe (don't-trust-
one-query): the step list has prepare_link_context / spawn_link_resolver / resolve_links → page-content-
writer SPAWNS the internal-link-resolver as a sub-agent (action lives in that agent's config, not the
writer's). Flow: load_site_specs → select_sections → spawn_research_agent → build_render_context →
process_sections_loop (LLM) → prepare_link_context → spawn_link_resolver → resolve_links → compile_page →
complete. So the resolver runs after content gen, augments each CTA section's resolved_data with a real
hub, returns sections_ready. Yet deployed hero = schema defaults (/contact.html + /services.html), not
the hubs the resolver would compute (/games/index.html + /guides/index.html). So its output isn't
reaching the rendered hero. THREE candidates: (1) spawned agent's sections_ready not threaded back into
compile_page — output-mapping / workflow-var-name mismatch, or sub-agent responds to own topic not
parent's; (2) resolved_data overwritten at render by schema-source resolution (merge order); (3) spawn
target has no agent_definitions row → spawn no-ops. NEXT: (A) jsonb_pretty the configs of
prepare_link_context/spawn_link_resolver/resolve_links/compile_page/build_render_context (trace whether
sections_ready reaches compile_page); (B) ILIKE %link%/%resolver% agents — does the spawned agent exist +
call resolve_internal_links action? Then read render_component / field resolver merge precedence. Fix
REUSES the resolver — likely realign a var name (sections_ready → what compile_page reads) or fix merge
precedence; no new resolver code.

**Session 2026-06-22 (cont.) — render_site_components shows the phantom was already fixed for site components; fixes converge.**
render_site_components_action.go: the SAME /contact.html phantom was already fixed for SITE components —
it captures the real contact URL from nav items (GetNavItems → real pages.url) for the header CTA
"instead of the hardcoded phantom /contact.html" (lines 141–157), fixed /privacy.html//terms.html legal
phantoms likewise (159–171), gates to no-button when absent (226–233). That's why the deployed
footer/header use /contact/index.html. Phantom source = deprecated loadNavItems/loadFooterNavItems
COALESCE(url, '/' || name || '.html'); the hero cta_url (schema source pages.contact) produces that same
/{name}.html. CONVERGENCE: resolve_internal_links ALREADY does it right (real-page-only via
validPages.Contains, gate if no hub) and would set a real hub (/games/index.html) — it just isn't reaching
the rendered hero. So the page-side fix = the resolver-data-flow fix (get sections_ready to drive the
render); reuse, not new URL logic. STILL PENDING (decisive fork): the orchestration check —
internal-link-resolver final_result + responses_topic, and page-content-writer got_sections_ready at
collected_data#>'{resolved_links,sections_ready}'. false→output lost in transit (path/topic mismatch,
select_sections falls back to schema-phantom plan); true but still phantom→render_component overrides from
schema source (fix precedence so resolver resolved_data wins, mirroring render_site_components). NOTE: I
have render_site_components (site renderer) but NOT render_component (page renderer) — need the latter (or
the orchestration result) to settle override-vs-fallback.

**Session 2026-06-23 — phantom-CTA root CONFIRMED: two stacked config bugs (fallback, not override).**
Orchestration check results: (1) page-content-writer — got_sections_ready_at_expected_path = f on EVERY
run; resolved_links = {"response":{"input_data":{"site_id":…,"sections":[…]}}}. So select_sections reads
resolved_links.sections_ready but there's no sections_ready there AND the envelope nests under .response;
what's present is the resolver's INPUT echoed (response.input_data.sections), not its output. → falls back
to input_data.section_plan.sections_ready (un-augmented plan w/ schema /contact.html). So it's the FALLBACK
branch — render_component is EXONERATED (not an override). (2) internal-link-resolver — all COMPLETED but
final_result EMPTY; responses_topic is the parent's spawn topic and the parent received a response, so
respond-to-parent wiring is FINE. The resolver isn't surfacing its resolve_internal_links output.
TWO STACKED BUGS (known families, both config-level, no new Go): (a) resolver doesn't surface action output
— final_result empty ⇒ internal-link-resolver's complete_workflow isn't emitting {sections_ready, unresolved}
(Part-1 result-contract/output_field family); (b) select_sections reads wrong path — resolved_links.sections_ready
vs the nested resolved_links.response.… the call_agent envelope produces (variable/path sync). Fix BOTH:
align resolver complete output to emit sections_ready + point select_sections' first path at the real
location → hero CTA becomes the real hub (/games/index.html) the resolver already computes. NEXT READS:
(A) jsonb_pretty internal-link-resolver workflow (why final_result empty / does complete surface sections_ready);
(B) jsonb_pretty a recent page-content-writer resolved_links full (exact path). Then the two config edits.
Separate from Part 4 clobber (Layer 1+2 unchanged).

**Session 2026-06-23 — resolver workflow is config-correct; empty final_result → result-extraction layer (Part-1 chassis?).**
internal-link-resolver workflow: resolve_links runs resolve_internal_links (output_field link_resolution) →
complete maps config.output = {sections_ready: link_resolution.sections_ready, unresolved:
link_resolution.unresolved}. That SHOULD give final_result={sections_ready,…}, but it's empty every run. So
the break is between action-output and complete-surfaces-it. TWO candidates: (1) the resolver's chassis
PREDATES the Part-1 result-contract fix — Part 1 is what taught complete_workflow to honour the
output/result_mapping mapping form (pre-fix: only plural output_fields honoured; else dump/stub). If
internal-link-resolver's image_tag wasn't bumped to the Part-1 chassis, config.output is ignored → empty
final_result (the echoed input_data envelope the parent saw fits the pre-fix fallback). (2) the action isn't
populating link_resolution (input_data.sections not resolving). DECISIVE: query collected_data.link_resolution
+ final_result on a resolver run, and compare image_tag of internal-link-resolver vs page-content-writer
(v1.0.1070, 06-23). action_produced_sections=true & final_result empty → result-extraction → check image_tag
(fix=redeploy on Part-1 chassis, reuse not new code). link_resolution empty → action input issue. Bug 2
(select_sections path: resolved_links.sections_ready vs nested resolved_links.response.…) stands regardless.
Aside (psql): #> takes a text[] path ('{a,b}'); #>> returns text; -> / ->> are single-key. The earlier B
errored because 'resolve_internal_links' isn't an array literal AND the key is resolved_links (output_field),
not the action name.

**Session 2026-06-23 — phantom-CTA ROOT CAUSE CONFIRMED (path mismatch); chassis hypothesis retracted; fix ready.**
Both internal-link-resolver and page-content-writer are on the SAME image v1.0.1070 → NOT a Part-1-not-deployed
issue. RETRACT the chassis hypothesis. Full envelope: resolved_links = {response:{input_data, resolve_links,
link_resolution}, response_status,…}. The resolver's augmented sections ARE present at
resolved_links.response.link_resolution.sections_ready (and .response.resolve_links.sections_ready), but
select_sections reads resolved_links.sections_ready (top level) → null → fallback to
input_data.section_plan.sections_ready (un-augmented → schema phantom). So FALLBACK via PATH mismatch. Empty
final_result explained: the parent's reply is built from the resolver's collected_data (link_resolution present),
not final_result — data flows despite empty final_result, so "bug 1" is a cleanliness oddity not data-loss.
FIX (workflow-only, DB, immediate): repoint select_sections.config.fields.sections_ready[0] from
resolved_links.sections_ready → resolved_links.response.link_resolution.sections_ready (fallback kept). SQL:
snapshot_agent('page-content-writer',…) then jsonb_set on '{workflow,steps,select_sections,config,fields,sections_ready}'.
Takes effect on next FULL build (bare rerender won't re-run resolver); gamesdesign guide-economy-basics hero
(generic 'hero', resolver matches) then resolves cta_url to /games/index.html.
TWO FOLLOW-UPS (separate): (1) resolver ctaFieldNames matches only exact "hero"/"call-to-action" → variants
(hero-about, platform-comparison, gauntlet-cta) with renderer CTA fields are skipped (coverage gap); (2) resolver
returns its whole collected_data (echoed input incl. all component HTML — huge) + empty final_result; make it
return lean {sections_ready, unresolved}, then select_sections simplifies to resolved_links.response.sections_ready
(needs complete_workflow output-mapping vs output_field behaviour confirmed). Runbook phantom-CTA thread updated.

**Session 2026-06-24 — Part 4 (interactive clobber): fix spec PINNED; implementation gated on source (bundle created).**
Investigation is complete (cause confirmed; blast radius = game-pathfinding only; page-build-handler has NO
item_type branch — load_existing_content runs but unconditionally proceeds to the writer). Remaining = the
two-layer Go fix; the current save_page_sections + planning source is NOT in context, so the gating step is
getting it (clobber bundle invocation saved in HANDOFF, -out /tmp/bundle_clobber.md). Key facts confirmed from
the transcript: interactivity is detected from the page_components.**rendered_html** column via
ILIKE '%<canvas%' OR '%game-container%' OR '%tool-page%' (there is no stored has_interactive_markup column — it
was computed in the sweep).
LAYER 1 (pinned): EXTEND the existing content-regression guard in save_page_sections. It already holds old+new
rendered_html and computes a prose-text regression (the "content regression blocked: N chars vs M" errors); it
waved the game through because 18KB game → 2KB text GREW the prose while losing markup/JS. Add a parallel check
on the same rendered_html reusing the sweep signal (canvas/game-container/tool-page, page-level bool_or): if the
page is currently interactive and the incoming set is not → block + escalate (mark_needs_review), same as the
text guard. NOTE: data-component alone is NOT a signal (every section carries it). Reuse, not a new guard.
LAYER 2 (pinned): clobber happens because the rebuild plans from the spec (load_spec_sections =
load_page_sections_from_spec → plan_sections) and the tool is a page_component not in the spec. Preserve non-spec
interactive page_components by merging them into the planned set (in load_page_sections_from_spec or plan_sections)
so DELETE+INSERT keeps them. Keeps CTA re-resolution via the existing writer; no routing change (Layer 3 ruled out
by P4.2). Exact insertion point TBD with source.
SOURCE NEEDED: save_page_sections_action.go (L1), plan_sections_action.go + load_page_sections_from_spec (L2),
registry.go. Plus small items: stamp source_item_id into page_component_history on save; after the fix, re-create
the game-pathfinding A* tool (+ fix its needs_page mis-key per Part 3); delete stale
z_context/check_phantom_internal_links.go. HANDOFF §4 refined + Bundles section added.

**Session 2026-06-24 (cont.) — Part 4 fix WRITTEN (patched save_page_sections_action.go in /mnt/user-data/outputs/).**
Source arrived (save_page_sections_action.go, plan_sections_action.go, load_page_sections_from_spec_action.go).
KEY INSIGHT: the interactive tool exists ONLY as page_components.rendered_html — bespoke <canvas>/JS markup not
in the spec, not LLM-regeneratable. The planning path (load_page_sections_from_spec → plan_sections) deals in
section-NAME skeletons (lists of strings) + regeneratable content, so it CANNOT reconstruct the tool. =>
preservation of the actual markup must happen at the SAVE path; BOTH layers live in save_page_sections, not the
planner (this CORRECTS last turn's "Layer 2 in planning path"). Confirmed anchors: text-only regression guard
251-287; DELETE+INSERT 322-393; history snapshot 296-310 omitted source_item_id (the NULL we traced).
PATCH (brace/paren/bracket-balanced; gofmt unavailable in sandbox — user runs `go build ./...`; no new imports):
  - Layer 2 carry-forward, before the content-regression guard: SELECT existing deployed interactive sections
    (rendered_html ILIKE canvas/game-container/tool-page); per section — same slot present & interactive → keep;
    same slot but rebuild produced non-interactive content (hero-as-text case) → replace incoming HTML/content_data
    with preserved markup; slot dropped → re-append. Runs BEFORE Layer 1. INSERT loop uses i+1 for position
    (ignores SectionData.Position) so order = slice order; re-appended tools land last.
  - Layer 1 interactivity guard, after the content-regression guard: existing deployed interactive (bool_or) AND
    new set not interactive → return "interactivity regression blocked…". Safety net. Helper
    sectionHTMLIsInteractive(html) = strings.ToLower + Contains canvas/game-container/tool-page.
  - source_item_id: history INSERT now writes source_item_id=$2 from config-driven work_item_id_field
    (nil → SQL NULL; unchanged until the workflow sets it).
TO FINISH (separate from Go change): (1) workflow config — add "work_item_id_field":"<collected_data path to
driving work-item id>" to the save_sections step on every agent that saves sections; match what plan_sections gets
via its work_item_id input; confirm path from an orchestration collected_data dump (immediate, no redeploy).
(2) deploy Go change → go build → GitHub → Backblaze → image → snapshot_agent + bump image_tag on
page-build-handler, page-rerender, and any builder with a save_sections step. Layer 1 fires only on interactive
pages (guides/index unaffected). Interacts with Part 2: a rerender producing non-interactive sections for an
interactive page now blocks (loud) not clobbers; a NULL-content_data tool section escalating to needs_page now hits
Layer 2 and is preserved (loop broken). (3) after deploy, re-create game-pathfinding A* tool (+ fix needs_page
mis-key per Part 3); carried-forward tool ideally needs content_data so a future rerender doesn't escalate it.
(4) delete stale z_context/check_phantom_internal_links.go.

**BUNDLE-RUN ERRORS (both pasted runs failed — fixable; NOT bundle-content problems):**
  (1) Phantom-CTA bundle — bash "syntax error near unexpected token '('" ×3 then "-psql not found". CAUSE: the
      -doc paths contain unquoted parentheses (016_…_v2_56(1).md, 005_tool_pipeline(1).md,
      003_contracts_and_standards(7).md); bash treats ( ) as metacharacters, which also broke the backslash
      line-continuation so -psql ran as a command. FIX: single-quote EVERY -doc path containing parens. Also `ls
      docs/agent_docs/docs024_key_docs_latest/` to confirm real names — some have (1)/(7) parens; the clobber
      bundle assumed _7_/no-parens variants that may not exist.
  (2) Clobber bundle — got past schema/capabilities/runtime, then "load analysis: unexpected end of JSON input".
      CAUSE: /tmp/analysis_repo.json is empty/truncated. FIX: regenerate it (find the analyse command under
      contextkit/cmd — analyse.go/callgraph.go exist; run it to write /tmp/analysis_repo.json; confirm `wc -c`
      > 0) then re-run the bundle. Scope/docs/schema/runtime are otherwise correct.

**Session 2026-06-24 (cont.) — index rebuild VERIFIED on a real build; runbook dated + per-step checkboxes added.**
Runbook edits (user ask): added `**Created:** 2026-06-23 · **Last updated:** 2026-06-24` header + a checkbox
convention note; tagged 27 step headings and the 3 P4.3 sub-steps with `[ ]`.
Index verification dump (gamesdesign index page 6e988cc4): work item 0b8114c1 complete (build-dispatch-loop,
0/3, claimed 14:12:49 → completed 14:27:31); page-build-handler 959e9203 + page-content-writer c392ad29 both
COMPLETED. Writer collected_data page_content.response: sections_metadata = array, sm_count=5, has_html=t,
stub_message empty → Part 1 result contract working (no stub, no oversize). page_components: 5 deployed sections
(hero 2207, tool-list 8987, guide-list 7470, game-list 8136, system-stats 7369, updated 14:26:13).
pages.build_status=deployed, deployed_at 14:27:15, version 1. stub_rows=0; no CHILD_ORCHESTRATION_FAILED/delivery-cap
for the writer. So Part 1 + the Part 2 happy path are confirmed on a live build, and the long-pending "index has
5 sections" is DB-confirmed.
TWO THINGS TO CHECK (no conclusions yet):
  (1) The deployed index.html the user pasted shows only 4 sections in <main> (hero, tool-list, guide-list,
      game-list) — MISSING system-stats, which IS in page_components (position 5, 7369 bytes). Most likely the
      pasted HTML predates the 14:26 rebuild (stale copy); the alternative is the deploy step isn't emitting the
      5th section. Disambiguate: re-fetch live /index.html (or inspect the 14:27:15 deploy artifact) and check for
      data-component="system-stats". Also pages.last_built_at is NULL while deployed_at is set, and
      page_components.deploy_commit is empty — minor bookkeeping to glance at, not necessarily wrong.
  (2) NEW THREAD — build-dispatch-loop render_page → rerender_single_page failing: "invalid page_id: invalid UUID
      length: 31" (×2, 11:45 & 12:05) and "length: 18" (11:44), error_code CHILD_ORCHESTRATION_FAILED. A non-UUID
      (len 31/18, not 36) is being passed as page_id to rerender_single_page. Distinct from the index build (which
      used page-build-handler, not rerender_single_page) and predates this run. Next diagnostic: find what
      build-dispatch-loop's render_page step passes as page_id (a page name/slug or a truncated/concatenated id?)
      — i.e. the value source for rerender_single_page's page_id input. Track separately.

**Session 2026-06-24 (cont.) — index `system-stats` drop CONFIRMED not-stale; suspect = the assembler's visible-content filter.**
User re-fetched live /index.html (12 min after the 14:27 deploy): still 4 sections in <main> (hero, tool-list,
guide-list, game-list), NO system-stats; page_components had 5 (system-stats pos 5, 7369 bytes). So the deploy is
genuinely dropping it. Uploaded rerender_single_page_action.go is the assembler (assemblePage output byte-shape
matches the deployed HTML). MECHANISM SUSPECT: assemblePage → getPageSections → sectionHasVisibleContent strips
<style>/<script>/tags/entities/whitespace and DROPS any section with ≤10 visible chars (logs "getPageSections:
skipping section with no visible content"). A system-stats widget with JS-rendered or empty numbers has ~0 static
text → filtered. DO NOT CONCLUDE yet — the page may have been rebuilt again since 14:26, so re-check first:
  Q1: SELECT position, slot_name, length(rendered_html), updated_at FROM page_components WHERE page_id=
      '6e988cc4-4898-4021-aa5e-2ab0271f9b75' ORDER BY position;  (is system-stats still there?)
  Q2 (if present): compute approx_visible_len via nested regexp_replace mirroring sectionHasVisibleContent
      (style→script→tags→entities→whitespace); ≤10 means the filter dropped it.
  Logs: grep 'getPageSections: skipping…' / 'filtered empty sections' / 'RerenderSinglePageAction'.
FORK: (a) system-stats GONE from page_components → builder/plan drop, not the assembler (separate, Part-4 spec-path
family). (b) present + visible_len≤10 + genuinely empty numbers → upstream content bug (populate the stats); filter
correctly hid an empty shell. (c) present + real content but ≤10 because numbers are in <script>/data-attrs → filter
FALSE POSITIVE → make sectionHasVisibleContent component-aware (recognise stats/interactive/data-component as
visible) — the STRUCTURAL fix, and it should SHARE the Part-4 canvas/game-container/tool-page signal, not invent a
second one. THIS IS THE SAME TEXT-HEURISTIC BLIND SPOT as Part 1 (text regression guard) and Part 4 (interactivity).
CROSS-LINK: getPageSections is a SECOND silent-drop path for interactive content, independent of save_page_sections —
the Part-4 carry-forward keeps a game in page_components, but if its prose is thin this filter strips it at assembly.
game-auto-battler survives (its UI text presumably >10) — confirm. Do NOT patch until Q1/Q2 pick the branch.

**Invalid page_id (B) — mechanism, not yet root-caused to a value.** rerender_single_page is CORRECT: uuid.Parse →
"invalid page_id: invalid UUID length: 31"/"18" when handed a non-UUID. So an upstream render loop passes a
name/slug/path (or truncated/concatenated id) as page_id. loop_expansion_handler.go (SagaCoordinator) shows how a
loop injects per-iteration steps and resolves page_id from the loop item / propagated fields — so the loop is
iterating the wrong thing. NEXT DIAGNOSTIC: find the agent whose default_config contains BOTH 'render_page' and
'rerender_single_page' (SELECT type FROM agent_definitions WHERE default_config::text LIKE '%rerender_single_page%'
AND LIKE '%render_page%'), then inspect its render_page step config + the loop items_field — what does it iterate and
how is page_id mapped? If it iterates names/slugs not pages.id, that's the bug (resolve page_id to the real UUID).
build-dispatch-loop itself has NO render_page step (its sub_workflow is claim/spawn/call/mark) — the failing
render_page is in a spawned child handler.

**Session 2026-06-24 (cont.) — system-stats CONFIRMED true-empty (assembler filter is correct); page_id narrowed to page-rerender's parent loop.**
Q1: page_components still has 5 sections incl. system-stats (pos 5, 7369 bytes) → NOT a builder drop. Q2:
approx_visible_len = 0 for system-stats (vs hero 161, tool-list 1091, guide-list 1092, game-list 1018) → 7KB of
markup/CSS with ZERO visible text. So rerender_single_page.getPageSections.sectionHasVisibleContent correctly drops
it (0 ≤ 10). CORRECTION to prior framing: this is NOT the Part-1/4 false-positive pattern — the content is genuinely
empty, the filter is right. (The "would getPageSections also drop a low-text interactive game?" question stays open
— check game-auto-battler — but is separate from system-stats.) Deploy path confirmed = rerender_single_page (log:
response into page-build-handler carried rendered_page + "Rerender:" commit, files index.html + game-list.js; i.e.
page-build-handler used the rerender assembler to deploy). The user's grep returned nothing because it lacked `-E`
(literal `|`) and `logs*` won't match pod logs — NOT counter-evidence; Q2 is the proof. The skip WARN would be in the
page-rerender pod.
REAL BUG (upstream): system-stats rendered as an empty shell. NEXT DIAGNOSTIC: inspect page_components for
slot_name='system-stats' — content_data (empty → content-generation gap; populated → render/template gap) +
left(rendered_html,1400) (expect <style> + empty number/label slots). Filter is a fine safety net but the loss is
silent (WARN only); fix = populate system-stats (a wanted infographic), optionally raise a tracked signal for an
empty planned section instead of silently omitting it.
PAGE_ID (B) narrowed: agent with render_page+rerender_single_page = page-rerender. Its render_page reads page_id from
collected_data (input_fields page_id/site_id/domain); page-rerender + rerender_single_page are BOTH correct. page-
rerender is spawned per-page by a PARENT loop ("called by rerender-pages for each page in a loop") → the malformed
page_id (len 31/18) is set by the parent's loop→spawn input_mapping. NEXT: SELECT type FROM agent_definitions WHERE
default_config::text LIKE '%page-rerender%' AND type<>'page-rerender'; then inspect that parent's loop items_field +
page_id mapping (if it iterates names/urls/slugs and maps one to page_id instead of pages.id, that's the bug).
MINOR: page-rerender deploy_page commit_message "Rerender: {{.filename}}" renders as "Rerender: " (template var
unresolved) — cosmetic.
page-rerender def facts (v1.0.1074): start check_rerender_mode → (image_landed|section_data_resolved →
rerender_sections → check_escalated → save_sections) else render_page; render_page(rerender_single_page) →
check_skipped → deploy_page(git_commit) → update_status(page_id_field=rendered_page.page_id) → complete. save_sections
uses page_name_field=input_data.spec.page_name. So the patched save_page_sections (Part 4) runs in page-rerender too
→ page-rerender is one of the agents to bump on the Part-4 deploy (already listed).

**Doc gate:** the 002 routing caveat ("route by item_type, never by item_key") stays
until this ships; once the key encodes type it flips to "prefix equals item_type —
safe to filter." Do not pre-invalidate it.

---

## Key IDs / where things live

- Site `gamesdesign.co.uk`, page `index`. Deployed: 5 sections ~34k, written
  2026-06-06 16:59 (unchanged — confirms no rebuild has saved).
- Child writer runs (index): `472eed7d` (06-15, dropped; parent `cd73eea6` holds
  the stub) and `4e0b339a` (06-06, deployed; parent holds the real output — A/B).
- Stub signature: `page_content.response.message = "Workflow completed
  successfully. Full result exceeded size limit."`, `completed_steps` a number.
- Save step config (`page-build-handler`): `save_page_sections`,
  `sections_metadata_field = page_content.response.sections_metadata`,
  `html_field = validation_result.clean_html`, `output_field = sections_saved`.
- Writer `complete` (live, v2): `{output_field: page_content}` ← the bug.
  Correct sibling pattern: `page-build-handler` `complete` uses
  `output_fields: ["sections_saved","deploy_result"]`.
- Code read this session: `coordinator.go` `extractWorkflowResult` L3351,
  `MaxResultSizeBytes` L3432, `extractWorkflowResultWithSizeLimit` L3434,
  `extractMinimalResult` L3475; `v3_site_actions.go` `CompilePageSectionsAction`
  L1612, `extractSectionFromMap` L1838; `save_page_sections_action.go`
  instrumented earlier (Info logs at the save boundary; on a live rebuild now
  shows `metadata_array_len = -1` then "no sections found", corroborating the
  stub from the running system).
- Coordinator commit that caused it: `06a8c6e` (2026-01-14) on
  `platform/orchestration/coordinator.go` — "made extract function respect output
  fields"; introduced `extractWorkflowResultWithSizeLimit`/`extractMinimalResult`
  and switched `notifyParentOfSuccess` to the size-limited variant.
- Backups checked: writer `complete` = `{output_field: page_content}` at
  `agent_definitions_backup_20260422` (`v1.0.981`) and live (`v1.0.1063`); other
  listed backups had no `page-content-writer` row.
- Still open: the 06-06 successful-response shape (the anomaly above) — pin via a
  06-06-era successful child's `final_result`/parent `page_content.response` and
  the `call_agent`/`ExtractStepData` storage path, before finalising the fix shape.
- Empty in project mount: `002_intake_orchestrator.sql`, `003_site_classifier.sql`,
  `018_briefing_questionnaire.sql` (0 bytes).
- Docs updated this session: this file; `016_debugging_guide_v2_50.md` §9 (new
  entry "Child workflow result silently replaced by a stub — `output_field` vs
  `output_fields` …").
