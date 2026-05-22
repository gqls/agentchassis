# PLAN — Imagery Audit and Fix Loop

Sequenced work plan for closing the gap between what the planner / spec / other agents ask for in imagery, and what is actually delivered. Builds on `FOCUS_imagery_assessment.md`.

---

## Decisions taken

| Question | Decision |
|---|---|
| Imagery audit — extend existing or new agent? | New agent: **`imagery-quality-auditor`**, sibling of `visual-design-auditor` under `design-audit-agent`. |
| Max regeneration attempts per finding | **2** (matches the existing `max_fix_attempts: 2` on structured findings). |
| Asset locking | **Mirror `page_components` exactly**: add `locked_at timestamptz` and `locked_by text` to `assets`, same query exclusion patterns. Hard-vs-soft via `locked_by` (per 013 / 031). Timed expiry deferred to a future project — approved policy keeps human-set locks permanent; only `'deploy'` and auditor approvals get timed expiry. |
| Per-section / per-component image granularity | **Deferred**. Site-level imagery audit is sufficient for now. Don't introduce a richer `imagery_plan` schema until the basic loop is working. |

These decisions shape the plan: small additive changes, reusing existing handlers, no wholesale planner refactor.

---

## Progress (updated 2026-05-08)

Phases 0 and 1 are complete and end-to-end verified. The current row of an
`assets` insert against `00ff3af5-...` confirms the imagery_direction
prefix lands in `origin_prompt`, `origin_model='sdxl'` populates, and the
URL extraction works. See `STATUS_imagery_2026-05-08.md` for the
verification trail.

| Phase | Status |
|---|---|
| 0.1 | ✅ verified |
| 0.2 | ✅ verified |
| 1.1–1.5 | ✅ verified (Phase 1.5 needed an unplanned hotfix; details below) |
| 2 onwards | not started |

**Unplanned Phase 1.5 hotfix.** During verification we found that
`image-build-handler.call_logo_gen` and `call_hero_gen` were missing the
`output_mapping` block that `pageflow-builder` and `site-work-orchestrator`
already had. Without it, the URL the generator returned lived three levels
deeper than `store_asset` looked for it, so every dispatch-path image asset
silently `{stored: false}`'d. Fix copies the canonical 4-field
output_mapping to the two missing steps. SQL only, no code. Files:
`phase_1_5_pre_migration_backup.sql`, `phase_1_5_image_build_handler_output_mapping.sql`.

**Findings from verification (not blockers, but tracked):**

- Build-dispatch-loop isn't claiming triaged image items when page items
  are also in queue. Imagery items have been sitting in `triaged` for
  hours while page work flows. Worth a separate investigation; doesn't
  block Phase 2.
- The 5 hero variant flag-only items are being claim-attempted by
  dispatch despite empty `handler_agent`. They reach `failed` after 3
  attempts. Phase 2 makes them routable, sidestepping the issue.

---

## Sequencing principles

- Each phase is shippable on its own. No phase depends on a later phase.
- Schema changes are isolated to Phase 2; everything before that runs against today's schema.
- LLM-side changes (Phase 4 onwards) are gated on the algorithmic checks in Phase 1 working — the auditor should not have to re-discover what algorithms already caught.
- Each step within a phase is sized to a single PR / single deploy. No step touches more than 2-3 files.
- Reuse is called out per step. New code is the exception, not the default.

---

## Phase 0 — Wire what's already written but unread (no schema change)

**Goal:** Stop ignoring data we already have. Smallest possible change, demonstrates the path is live.

### 0.1 Read `imagery_direction` into the image prompt

- **File:** `platform/orchestration/actions/generate_image_actions.go` (or wherever `getImagePromptWithPriority` lives — confirmed by grep before editing).
- **Change:** After resolving the prompt via three-tier priority, look up `site_specs.aspect = 'design_intent'` for the current site, extract `data.imagery_direction`. If non-empty, prepend as a directive section: `"Style direction: <direction>\n\nSubject: <prompt>"`.
- **Reuse:** `datahelpers.ExtractNestedFieldString`, the existing site-spec read pattern from webdesign-agent.
- **Acceptance:** Log shows `imagery_direction` text in the prompt sent to the adapter. Site with `design_intent.imagery_direction = "warm hand-drawn illustration"` and `image_prompts.hero_home = "law firm office"` produces a hero whose origin_prompt contains both fragments.
- **Risk:** None significant. Worst case the directive is empty/missing and we fall through to today's behaviour.

### 0.2 Populate `origin_model` in `store_asset`

- **File:** the `store_asset` action (Go).
- **Change:** Read the active model from `agent_config["model"]` or the agent_definition's `default_config.model` for the calling agent. Write to `assets.origin_model` on insert.
- **Reuse:** Existing `loadAgentDefinitionForImageAction` or equivalent — it already returns the agent's `default_config`.
- **Acceptance:** Newly built site has `assets.origin_model = 'sdxl'` (or whatever Stability returns) on the hero/logo rows.
- **Why now:** It's a one-line write that costs nothing and immediately makes provenance visible. When we add provider routing later, the column is already populated.

### 0.3 (Verification only) Confirm the data is reaching where it's supposed to

Quick SQL check after a build:
```sql
SELECT a.purpose, a.origin_model, LEFT(a.origin_prompt, 200) as prompt_preview
FROM assets a WHERE a.site_id = '<site>';
```

`prompt_preview` should include the imagery_direction tokens. If it doesn't, 0.1 isn't fully wired.

---

## Phase 1 — Algorithmic discovery checks for spec-to-delivery gap

**Goal:** Catch the simple, structural mismatches without any LLM cost.

All three checks live in `platform/orchestration/actions/discovery_checks/`, follow the existing `DiscoveryCheck` interface, register via `init()` into the registry. Run from `design-discovery-agent`.

### 1.1 `check_unfulfilled_image_prompt.go`

- **Detects:** `site_specs` aspect `site_plan` has `data.image_prompts.hero_home` (or `data.image_prompts.logo`) but no `assets` row exists for that purpose.
- **Work item:** `item_type = 'needs_hero_image'` or `'needs_logo'`, **routes to existing `image-build-handler`**. No new handler.
- **Spec construction:** Copy `image_prompts` from the site_plan spec into the work item's `spec` field so the handler workflow finds it at `input_data.spec.image_prompts.hero_home`.
- **SQL (rough shape, validate against site_specs schema):**
  ```sql
  SELECT data->'image_prompts' AS prompts
  FROM site_specs
  WHERE site_id = $1 AND aspect = 'site_plan' AND is_current = true;
  -- then in Go: for each key in prompts, check if assets row exists
  SELECT COUNT(*) FROM assets
  WHERE site_id = $1 AND purpose = $2 AND status = 'active';
  ```
- **Acceptance:** Site with `image_prompts.logo` set but `assets.purpose='logo'` empty → work item created, `image-build-handler` picks it up via dispatch loop, generates and deploys logo.

### 1.2 `check_placeholder_image_in_use.go`

- **Detects:** Any `page_components.rendered_html` for the site contains the fallback path `/assets/images/hero.jpg` (or `/assets/images/logo.png`) **and** no `assets` row exists with the matching purpose. The asset never got generated; the component schema fell through to its hardcoded fallback.
- **Work item:** Same `needs_hero_image` / `needs_logo` types, same handler.
- **Acceptance:** Site whose hero section HTML contains `/assets/images/hero.jpg` but `assets.purpose='hero'` is empty → work item created, fixed by image-build-handler.
- **Why this matters:** Today this case fails silently. The site looks "fine" because the fallback resolves, but the brief asked for a hero image and one was never made.

### 1.3 `check_image_url_404.go` (lightweight version)

- **Detects:** `page_components.rendered_html` references `/assets/images/X.{jpg,png,webp}` where `X` doesn't match any `assets.purpose` for this site.
- **Work item:** Best-effort — if we can derive intent from context, route to `image-build-handler`; otherwise route to a flag-only finding.
- **Note:** This is the DB-only version. The "is the file actually committed in git?" version would need git-adapter integration — defer that.
- **Acceptance:** Site with manual `<img src="/assets/images/team-photo.jpg">` injected into a component but no `assets` row for `team-photo` → work item flagged.

### 1.4 Wire all three into `design-discovery-agent`

- **File:** `bk_agent_definitions_backup.sql` row for `design-discovery-agent`, `default_config.workflow.steps.run_checks.config.checks` array.
- **Change:** Append `"unfulfilled_image_prompt"`, `"placeholder_image_in_use"`, `"image_url_404"` to the existing checks list.
- **Reuse:** The single-line config update is the only change to the agent itself. The discovery check Go files do all the work.
- **Acceptance:** Trigger `design-discovery-agent` manually for a test site (`./trigger-audit.sh design-discovery-agent <site_id>`); look at `site_work_items` for the new item types.

### 1.5 `image-build-handler` smoke test for new item types

- **No code change** — the handler already handles `needs_logo` and `needs_hero_image`.
- **Verification:** Check that work items created by the new discovery checks flow through `claim → spawn → call → store_asset → deploy_image_asset → complete`, same as items created by the planner.
- **Risk:** The work item spec format from the discovery checks must match what `image-build-handler` reads at `input_data.spec.image_prompts.hero_home`. Confirm with a real test before declaring 1.1 done.

---

## Phase 2 — Schema groundwork: locking and multi-image readiness

**Goal:** Mirror the component locking pattern on assets. Open the door for multi-image purposes without forcing existing callers to change.

### 2.1 Add `locked_at` and `locked_by` to `assets`

- **Migration SQL:**
  ```sql
  ALTER TABLE assets
    ADD COLUMN locked_at timestamptz,
    ADD COLUMN locked_by text;
  CREATE INDEX idx_assets_locked ON assets(locked_at)
    WHERE locked_at IS NOT NULL;
  ```
- **Schema check:** `\d page_components` — confirm exact column types and pattern. Mirror precisely (Pattern A per 031_locks.md).
- **Lock semantics today:** Detection via `locked_at IS NULL` (unlocked) vs `IS NOT NULL` (locked); classification via `locked_by` value (hard = `admin`/`admin-removed`/`checkpoint`; soft = `deploy`/`manual`/auditor names/other). Reference helper: `platform/orchestration/actions/check_component_lock.go`'s `CheckComponentLock` function. **No time-based expiry exists today.**
- **Timed expiry (deferred):** The `lock_type` + `lock_expires_at` design from 004 v4 / 007 v4 is documented intent but not implemented. Approved policy is for a future, focused project that adds these columns uniformly across all four Pattern A tables in one migration. Approved policy keeps `'admin'` permanent (human edits never silently expire); only `'deploy'` and auditor approvals get timed expiry. See `LOCKS_should_locks_expire.md` for the rationale and implementation sketch. Imagery's lock needs are met by hard-vs-soft (encoded in `locked_by`) for now.
- **`locked_by` vocabulary** for assets (free-text identifier per 031_locks.md, no CHECK):
  - `'manual'` — human upload, hard
  - `'admin'` — human edit, hard
  - `'visual-design-auditor'` — Phase 4 auditor approved, soft
  - `'imagery-quality-auditor'` — Phase 6 auditor (future), soft
  - `'audit-pending'` — transient, agent-cleared on audit completion, soft

### 2.2 Update audit and discovery data queries to honour locks

Anywhere a query reads `assets` for a "what should I regenerate / fix" decision, add the lock filter:

```sql
WHERE locked_at IS NULL
```

When timed locks land (per 031_locks.md), this expands to
`WHERE locked_at IS NULL OR lock_expires_at < NOW()`. Until then,
`locked_at IS NULL` is sufficient.

- **Files to update:**
  - `discovery_checks/check_undeployed_assets.go`
  - `discovery_checks/check_unfulfilled_image_prompt.go` (from Phase 1 — add filter on first write)
  - `discovery_checks/check_placeholder_image_in_use.go`
  - The `load_design_context` query in `visual-design-auditor` (Phase 4)
  - Any future imagery-quality-auditor query

### 2.3 Add `asset_key` column (preparation for multi-image)

- **Migration SQL:**
  ```sql
  ALTER TABLE assets ADD COLUMN asset_key text;
  UPDATE assets SET asset_key = purpose WHERE asset_key IS NULL AND purpose IS NOT NULL;
  CREATE UNIQUE INDEX idx_assets_site_asset_key_unique
    ON assets(site_id, asset_key)
    WHERE asset_key IS NOT NULL AND status = 'active';
  ```
- **Do not drop the existing `idx_assets_site_purpose_unique` yet.** It still applies for `purpose IS NOT NULL`. Keep both for now — they overlap harmlessly while `asset_key = purpose` for everything.
- **Why:** Adoption mirroring (Phase 3) will write multiple `assets` rows for the same logical purpose-area (e.g. five mirrored images all with `purpose='adopted_image'`, distinguished by `asset_key='adopted:<filename>'`). Without the asset_key path, those inserts would fail the existing unique constraint.

### 2.4 Update `store_asset` to write `asset_key`

- **File:** the `store_asset` action.
- **Change:** Accept an optional `asset_key` config field. If absent, default to `purpose`. Write it on insert/upsert.
- **Backwards compatibility:** Existing callers (pageflow-builder, image-build-handler) keep working unchanged because their workflow doesn't pass `asset_key` and the default falls back to `purpose`.
- **Acceptance:** Existing builds produce `assets.asset_key = assets.purpose`. New callers can pass `asset_key='adopted:hero-bg.jpg'` and it lands in the column.

### 2.5 Drop the old purpose-only unique constraint

- **Only after** Phase 3 (adoption mirror) is using `asset_key` and Phase 2.3 has been live for some weeks with `asset_key` populated.
- **Migration SQL:**
  ```sql
  DROP INDEX idx_assets_site_purpose_unique;
  ```
- **Risk:** If any callsite relies on the old constraint for upsert behaviour (`ON CONFLICT (site_id, purpose) DO UPDATE`), those need updating to `ON CONFLICT (site_id, asset_key)` first. Grep before dropping.

---

## Phase 3 — Adoption image mirror

**Goal:** Stop discarding crawled imagery. Every adopted site retains its source images either as direct deploys (where licensing allows) or as references for our own generations.

### 3.1 New action `mirror_adoption_images`

- **Location:** `platform/orchestration/actions/mirror_adoption_images_action.go`.
- **Inputs:** `site_id`, `research_results_id` (or read latest `research_results` row with `result_type = 'adoption_crawl'` for the site).
- **Behaviour:**
  1. Read crawl result, extract image URLs from `images[]` array and link-extension matches.
  2. For each image: download (existing `webscrape` or direct HTTP through circuit breaker), upload to S3 via existing `storage.S3Client.Upload`.
  3. Insert one `assets` row per image:
     - `origin_type = 'adopted'`
     - `origin_url = <source URL>`
     - `purpose = 'adopted_image'`
     - `asset_key = 'adopted:<filename or hash>'`
     - `attribution`, `license` from any signal we can derive (often blank initially).
- **Reuse:** S3 client, HTTP client with circuit breaker (`platform/resilience`), the data extraction pattern from the existing scrape-result handlers.
- **Limits:** Cap per site (e.g. 50 images) and per-image size (e.g. 5MB). Log skipped items, don't fail the action for individual download failures.

### 3.2 Wire into `apply_adoption_plan`

- **File:** `bk_agent_definitions_backup.sql` row for `site-adoption-agent`, workflow steps.
- **Change:** Add a step `mirror_adoption_images` after `apply_adoption_plan`, before `complete`.
- **Reuse:** No new agent. The existing adoption agent is the natural place.
- **Acceptance:** Adopt a site that has 12 source images → 12 `assets` rows with `origin_type='adopted'` after adoption completes.

### 3.3 New discovery check `check_crawled_images_discarded.go` (backfill)

- **Detects:** Site has `research_results` rows with `result_type = 'adoption_crawl'` containing image URLs, but no `assets` rows with `origin_type = 'adopted'` for that site.
- **Work item:** `item_type = 'crawled_images_discarded'`, routes to a new `adoption-image-mirror` agent (3.4).
- **Why a check rather than just a backfill script:** Sites adopted before Phase 3 lands need this. After all old sites are processed, the check becomes a no-op for new sites (because 3.2 prevents the gap).

### 3.4 New agent `adoption-image-mirror`

- **Category:** `specialist`. Orchestrator processing mode.
- **Workflow:**
  ```
  ensure_site_record → mirror_adoption_images → complete
  ```
- **The agent is essentially a one-step wrapper** around the action from 3.1. We make it an agent (not just an action call from the dispatch loop) because that's the project pattern: every dispatched work item type maps to an agent.
- **Definition:** New row in `agent_definitions`. Reuse the resource limits, image_repository, image_tag from sibling specialists.

### 3.5 Reference imagery available to image-generator

- **No code change yet** — once `assets` has `origin_type='adopted'` rows, downstream work (img2img generation, style references, future per-section illustrations) can read them.
- **Future hook (deferred):** When the image-generator gains `reference_image_uri` support, prefer adopted images for the matching purpose / section.

---

## Phase 4 — Visual auditor extension (text-only, cheapest LLM-side win)

**Goal:** Make the existing visual-design-auditor aware that imagery exists, without a vision model.

### 4.1 Extend `load_design_context` SQL

- **File:** `visual-design-auditor` agent definition, `default_config.workflow.steps.load_design_context.config.query`.
- **Change:** Add subqueries to the existing big SELECT:
  - `(SELECT json_agg(json_build_object('purpose', purpose, 'asset_key', asset_key, 'url', url, 'origin_prompt', origin_prompt)) FROM assets WHERE site_id = s.id AND status = 'active' AND locked_at IS NULL AND origin_type IN ('generated', 'adopted')) as assets`
  - `(SELECT data->>'imagery_direction' FROM site_specs WHERE site_id = s.id AND aspect = 'design_intent' AND is_current = true) as imagery_direction`
- **Schema check:** Confirm `site_specs.aspect = 'design_intent'.data.imagery_direction` is the actual JSON path. Some site_specs rows have nested structure — verify with a sample row.
- **Reuse:** Same query template style as the existing `component_samples` and `index_samples` subqueries.
- **Risk:** Query becomes large. If it breaks the action's prompt template, split into two query steps (`load_design_context` + `load_imagery_context`) and merge in the prompt.

### 4.2 Update auditor LLM prompt

- **File:** Same agent definition, `default_config.workflow.steps.run_visual_llm_audit.config.prompt`.
- **Change:** Add `IMAGERY` as the 6th check category in the existing `Check for:` list. Add a paragraph providing the new context:
  ```
  Imagery context:
  - Imagery direction (designer's intent): {{.design_context.imagery_direction}}
  - Generated assets: {{.design_context.assets}}
  
  6. IMAGERY: missing required images, hero/logo absent, imagery that contradicts the imagery_direction
  ```
- **Acceptance:** Site with no hero asset gets a finding like `{"category": "imagery", "description": "no hero image generated", ...}`.
- **Risk of false positives:** The auditor might double-flag what the algorithmic checks (Phase 1) already caught. Mitigate by adding to the prompt: `"Algorithmic check results (already handled — do not re-report these): missing hero: {{.algorithmic_results.missing_hero}}, missing logo: {{.algorithmic_results.missing_logo}}"`. Same pattern as the existing algorithmic_results passthrough.

### 4.3 Tune for false positives over a few real sites

- Run 4.1 + 4.2 against 5-10 existing sites without enabling automatic fixes.
- Read findings. Any imagery findings that are obviously wrong (e.g. flagging a correctly-generated hero as "doesn't match direction") get triaged against the prompt.
- **Acceptance:** ≥80% of imagery findings on a random sample are accurate (i.e. fixing them would actually improve the site).

---

## Phase 5 — Vision-capable LLM path

**Goal:** Add the foundational capability for an auditor to actually look at images. Required by Phase 6.

### 5.1 Extend `aiservice.AIService` interface

- **File:** `platform/aiservice/interface.go`.
- **Change:** Add a method:
  ```go
  GenerateTextWithImages(ctx context.Context, prompt string, imageURLs []string, options map[string]interface{}) (string, error)
  ```
  Or, equivalently, accept `image_urls` in the existing options map — interface preference depends on local conventions; check what `GenerateText` does today before deciding.
- **Backwards compatibility:** Existing `GenerateText` callers unchanged.

### 5.2 Implement Anthropic vision

- **File:** `platform/aiservice/anthropic.go`.
- **Change:** Build messages with image content blocks per Anthropic's vision API:
  ```json
  {
    "role": "user",
    "content": [
      {"type": "image", "source": {"type": "url", "url": "<image_url>"}},
      {"type": "text", "text": "<prompt>"}
    ]
  }
  ```
- **Note:** Image URLs must be publicly accessible. Our presigned B2 URLs work (valid for 7 days), but expired ones won't. Auditor should check `assets.url` is a fresh presigned URL or fall back to regenerating one before the call.
- **Reuse:** Existing HTTP client, error handling, retry logic.

### 5.3 Workflow action wrapper

- **Decision:** Extend `execute_llm_prompt` action to accept an `image_urls_field` config that resolves to a list of URLs at runtime, OR add a new `execute_llm_prompt_with_vision` action.
- **Lean toward extending** the existing action — fewer registry entries, same workflow shape. Add the new field only if vision is requested.
- **Acceptance:** A test workflow calling `execute_llm_prompt` with `image_urls_field: "design_context.assets[].url"` produces an LLM response that reflects the image content (not just text).

### 5.4 Cost monitoring

- Anthropic vision is more expensive per call than text-only. Add `vision_call: true` to `llm_call_log` metadata so cost analysis can separate them.
- **Reuse:** existing `llm_call_log` writer.

---

## Phase 6 — `imagery-quality-auditor` agent

**Goal:** A separate agent dedicated to imagery audit, vision-capable, sibling of `visual-design-auditor`.

### 6.1 Agent definition

- **Category:** `analyst`. Processing mode: `orchestrator` (every agent is an orchestrator).
- **Resource limits:** Match `visual-design-auditor`.
- **Workflow:**
  ```
  ensure_site_record
    → load_imagery_context        (SQL: assets, imagery_direction, identity)
    → check_has_imagery           (conditional — skip if no assets to audit)
    → run_imagery_audit_vision    (execute_llm_prompt with image_urls)
    → write_audit_findings        (existing action, source = 'imagery-quality-audit')
    → complete
  ```
- **Reuse:** `ensure_site_record`, `query_database`, `write_audit_findings`, the conditional pattern — all existing actions. Only the audit prompt is new.

### 6.2 Audit prompt

The prompt should follow the existing TOP 5 / acceptance_test contract used by `visual-design-auditor`. Categories specific to imagery:

- `direction_mismatch` — image doesn't reflect the imagery_direction
- `brand_mismatch` — colour palette of generated image clashes with site palette
- `inconsistency` — multiple hero variants don't share a visual style
- `quality` — visible artefacts, garbled text in generated image, AI giveaways
- `inappropriate` — generated content unsuitable for the site's audience or tone

Each finding outputs `max_fix_attempts: 2` (Decision 2 above) and a handler routing recommendation:
- For direction/brand/inconsistency mismatches: `image-build-handler` regenerates with an updated prompt
- For quality issues: `image-build-handler` regenerates with a different seed
- For inappropriate content: `image-build-handler` regenerates with stronger negative prompts; if 2 attempts fail, mark as `needs_human_review`

### 6.3 Lock honouring

- The `load_imagery_context` query filters `WHERE locked_at IS NULL AND origin_type IN ('generated', 'adopted')`. Human uploads (origin_type='uploaded') are excluded — same pattern as content-quality-auditor's locked-component exclusion. (Filter expands to `locked_at IS NULL OR lock_expires_at < NOW()` when timed locks land per 031_locks.md.)

### 6.4 Hook into `design-audit-agent`

- **File:** `design-audit-agent` agent definition.
- **Change:** Add `spawn_imagery_auditor` and `call_imagery_auditor` steps, parallel to the existing visual and content auditor calls. Sequence: visual → content → imagery → triage.
- **Reuse:** Identical pattern to the existing two auditor calls. Copy-modify the step definitions.

### 6.5 Pass count and rate limiting

- Imagery audit findings count toward the existing 3-pass cap on the improvement loop (`sites.settings.audit_pass_count`). No new counter — same mechanism.
- **Per-finding `max_fix_attempts: 2`** caps regeneration thrash.
- **No new infrastructure needed** — both guardrails use existing fields.

### 6.6 Initial deployment as gated rollout

- Deploy with `is_active = false` in the `design-audit-agent` workflow first; verify the agent runs cleanly on a test site (`./trigger-audit.sh imagery-quality-auditor <site_id>`).
- Then flip the call site in design-audit-agent live.
- **Reuse:** Same gradual-rollout pattern from the batch processing migration.

---

## What we're explicitly NOT doing in this plan

The user's deferral on per-section granularity simplifies several things:

- **Planner output stays as-is.** No richer `imagery_plan` schema. `image_prompts.{logo, hero_home}` continues to be the contract.
- **No per-component asset_key enforcement.** `asset_key = purpose` for everything in the build pipeline. Adoption-image-mirror uses `asset_key = adopted:<filename>` because it has multiple per site, but no other code path uses asset_key as a discriminator.
- **No icon resolver.** The `icon` string field in features content_data continues to be unrendered. Defer until per-section imagery is on the table.
- **No infographic generator agent.** Same reason.
- **No image-generator provider router.** Stability remains the only provider. Multi-provider is a separate plan when we want to bring Banana Pro / FLUX / Midjourney in.
- **No img2img / reference-image generation path.** Adopted images are persisted but only as historical record for now. They become useful when reference-image generation is added.

These are valid follow-ups; they're just not in this plan because they don't belong to the spec-to-delivery loop closure.

---

## Phase summary table

| Phase | Goal | New files | Modified files | Schema changes | Agents touched |
|---|---|---|---|---|---|
| 0 | Wire imagery_direction; populate origin_model | none | `generate_image_actions.go`, `store_asset_action.go` | none | none |
| 1 | Algorithmic discovery checks | `check_unfulfilled_image_prompt.go`, `check_placeholder_image_in_use.go`, `check_image_url_404.go` | `agent_definitions` (design-discovery-agent config) | none | design-discovery-agent (config only) |
| 2 | Asset locking + multi-image readiness | none | `assets` table | `locked_at`, `locked_by`, `asset_key` columns; new index; eventually drop old constraint | none |
| 3 | Adoption image mirror | `mirror_adoption_images_action.go`, `check_crawled_images_discarded.go` | `agent_definitions` (site-adoption-agent workflow) | none | site-adoption-agent, new `adoption-image-mirror` |
| 4 | Visual auditor sees imagery (text-only) | none | `agent_definitions` (visual-design-auditor) | none | visual-design-auditor |
| 5 | Vision-capable LLM path | maybe `execute_llm_prompt_with_vision_action.go` (TBD) | `aiservice/interface.go`, `aiservice/anthropic.go`, possibly `execute_llm_prompt_action.go` | none | none directly |
| 6 | imagery-quality-auditor agent | none (workflow lives in agent_definitions) | `agent_definitions` (design-audit-agent) | none | new `imagery-quality-auditor`, design-audit-agent |

---

## Risks and mitigations

| Risk | Phase | Mitigation |
|---|---|---|
| Discovery check spec format mismatch with image-build-handler | 1.5 | Run a real test before declaring 1.1 done. Confirm work item spec lands at `input_data.spec.image_prompts.hero_home`. |
| `assets` UNIQUE constraint blocks adoption mirror | 2.3 → 3.1 | 2.3 must land before 3.1 deploys. The plan sequences this correctly. |
| Visual auditor false positives on imagery findings | 4.3 | Manual sample run on 5-10 sites before enabling fixes. |
| Vision API URL expiry | 5.2 | Refresh presigned URLs immediately before the call. Add to the auditor's `load_imagery_context` step. |
| Audit pass thrash on imagery findings | 6.5 | `max_fix_attempts: 2` per finding; existing 3-pass site cap. Both already in place; no new logic needed. |
| Cost spike from vision calls | 5.4, 6 | Tag `vision_call: true` in llm_call_log. Monitor costs in week 1 of Phase 6 rollout. |

---

## Open items / future phases (not part of this plan)

- Per-section/per-component imagery (richer planner output) — the deferred granularity question. Worth revisiting once Phases 0-6 are stable and we have enough data to know which sections genuinely need their own imagery.
- Multi-provider image generation (Banana Pro, Midjourney, FLUX). Separate plan.
- Icon library and infographic generator — separate plans, may share scaffolding with the imagery_plan work above.
- Per-vertical LoRA fine-tunes — pre-existing plan in `018_canine_biology.md`, unblocked by Phase 5's vision work but not dependent on it.
- Vision-based hero variant deduplication ("the same image is on three pages") — natural extension of Phase 6, low priority until per-section imagery exists.
