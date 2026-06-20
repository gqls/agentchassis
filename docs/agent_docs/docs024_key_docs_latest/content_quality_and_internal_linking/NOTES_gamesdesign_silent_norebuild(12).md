# Running notes — gamesdesign.co.uk silent no-op rebuild

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

**Part B — key canonicalization (after Part A).** Canonical builders in
`work_items_common.go` (prefix == item_type); repoint adoption tool items →
`needs_tool_recreation:<name>` (fixes Bug 2), adoption content → `needs_content_page:<name>`;
flag/section-data keys follow Part A's type decision. Decisions needed: should
`needs_page` and `needs_content_page` (two types, same handler, same work) share a
dedup namespace or unify; should adoption-recreate and reconcile-build of the same
page dedup or stay distinct intents.

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
