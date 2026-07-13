# Phase 0.1 Imagery Changes — Assessment Against New Architecture

Reviewing my Phase 0.1 deliverables (`phase_0_1_generate_image_actions.diff`, `phase_0_1_image_generator_imagery_direction.sql`) against:

- `029_site_plan_and_reconciler_1_.md` — declarative plan + reconciler
- `030_phase1_plan_and_reconciler_2_.md` — Phase 1 commitments
- `007_adoption_pipeline_v4.patch` — adoption rewriting around the new split
- `001_development_guide.md` — engineering conventions
- `002_system_architecture.md`, `003_contracts_and_standards_v7.md` — operating norms

## Verdict

**The Phase 0.1 changes hold up.** Both files are compatible with the Phase 1 architecture and don't block any of its moves. Two caveats are worth recording, neither requiring a rewrite of what I produced.

The reasoning matters: doc 030 makes a clean split between *strategic* `site_specs` aspects (brand-level, written by classifier or adoption, never overwritten) and *plan-time* `site_plan_directives` rows (build-time guidance, written by plan-builder). My read targets the strategic spec — which Phase 1 actively protects from being overwritten — so it survives the architectural shift unchanged.

---

## What Phase 1 changes (relevant to my work)

The pieces that touch imagery flow:

| Area | Today | After Phase 1 | Effect on my Phase 0.1 |
|---|---|---|---|
| `site_specs.design_intent` | Written by three agents (classifier, adoption, build-site-planner). Last writer wins, so `imagery_direction` muddies. | Written by classifier or adoption only. Build-site-planner stops overwriting. The strategic spec persists across plan rebuilds. | **Improves my read.** `imagery_direction` becomes a stable, brand-level value sourced from one of two known agents. |
| Plan-time imagery direction | Implicitly part of `design_intent` after planner overwrites it. Lost on next planner run. | Lives in `site_plan_directives` (category=`design`, subject probably `imagery.direction` or similar). Site-wide row when scope=`site`; per-page row when scope=`page`. Survives plan rebuilds via lock transfer. | **My read does not see this layer.** When Phase 1 lands, image-generator could optionally also pull relevant directives via the brief renderer. Today this layer doesn't exist. |
| `build-site-planner` workflow | Calls `write_design_intent`, `write_content_direction`, `write_plan_spec`, then `write_build_items`. | Single LLM call, terminal step is `write_site_plan` (writes plan-domain tables). Removes the three site_specs writes and `write_build_items`. | **No effect.** I don't touch build-site-planner's workflow. |
| `pageflow-builder` (legacy) | Same shape as today; uses old `site-planner`. | Doc 030: "the older `site-planner` (used by `pageflow-builder`) is left alone." | **No effect.** My migration touches `pageflow-builder.generate_hero_image` and `pageflow-builder.call_logo_generation`. Both stay valid. |
| `image-build-handler` | Generates hero/logo from work item spec. | Unchanged in Phase 1. The reconciler emits `needs_page:<name>` items; page-build still wants hero images. | **No effect.** |
| `site-work-orchestrator` | Existing build orchestrator. | Coexists with the new flow during Phase 1 transition; eventually becomes the test path for the reconciler. Workflow unchanged in this phase. | **No effect on input_mapping additions.** |
| Adoption writes | Adoption writes `design_intent` containing `imagery_direction` (per existing classifier prompt and the adoption fingerprint flow). | Same — adoption is one of the two agents allowed to write `design_intent`. The patch to doc 007 is explicit: adoption's `design_intent` is not touched by the planner. | **Improves my read.** For adopted sites, `imagery_direction` reflects the actual fingerprinted source, not a planner overwrite. |

The architectural split is in my favour: strategic imagery direction stabilises, doesn't drift, and is exactly what the image-generator wants for hero/logo. The thing I miss — per-page plan-time imagery rules — is a future enhancement that Phase 1 *enables*, not something Phase 1 *requires*.

---

## Caveats worth recording

### Caveat 1 — `site_id` vs `target_site_id` naming

Doc 030 added a callout in the `target_site_id` section:

> `ExtractActionInputs` runs a nested-source loop late in its resolution chain... If a caller's `input_mapping` doesn't explicitly map `site_id` AND the cascade has populated `site_record.site_id` in CollectedData, an action with a `site_id` field can silently pick up that nested value... New code should use field names that don't collide with database columns or with the nested-source list... The convention here is: **leave existing code alone, but write new code with collision-free names**.

My migration adds `"site_id": "site_record.site_id"` to six input_mappings. Three observations:

1. **`GenerateImageAction` does not use `ExtractActionInputs`.** It uses the older `datahelpers.GetInputData(params.CollectedData, logger)` which returns `collectedData["input_data"]` directly with no nested-source walk. The collision risk the callout describes simply does not apply on this code path.

2. **Existing code already uses `site_id` everywhere in input_mappings.** The fix_items_loop in `site-work-orchestrator` passes `"site_id": "site_record.site_id"`. So do dozens of other call_agent steps. Renaming this single field would be inconsistent with surrounding code and would surprise anyone reading the input_mapping for image-generator.

3. **Doc 030's convention is "leave existing code alone."** Image-generator is existing code; the diff is additive.

Decision: keep `site_id`. This is consistent with the convention's own stated boundary. Worth a one-line comment on the eventual PR noting the choice.

### Caveat 2 — the diff's `<line>` placeholder

The diff file uses `<line>` as a placeholder for unified-diff line numbers because I built it from the bundled context dump rather than the live repo. It is not directly applicable via `git apply`. The intent is:

- The two helper functions (`getImageryDirectionForSite`, `composeImagePromptWithDirection`) get appended to `generate_image_actions.go` (or split into a sibling file in the same package).
- The 17-line block injecting `imagery_direction` goes immediately after the existing `getImagePromptWithPriority` call, before the existing `params.Logger.Info("Selected prompt for execution"...)` line.

For the actual PR, generating a real unified diff against the live file is a one-line `git diff` after applying the change manually. I can produce that once we have the file open if helpful.

---

## What Phase 0.1 deliberately doesn't capture (to be revisited after Phase 1)

When Phase 1 ships, plan-time imagery direction will live in `site_plan_directives`. The brief renderer (`datahelpers/page_brief.go` per the work order) is the consumer-facing read. There are two cases:

1. **Site-wide imagery direction (scope=`site`, category=`design`, subject likely `imagery`).** Today's logo and hero-home generation could legitimately want both the strategic `site_specs.design_intent.imagery_direction` *and* the plan-time directive. In practice the strategic spec is what was used to author the plan-time directive, so reading both for a site-wide image is largely redundant. **Phase 0.1's strategic-only read is fine.**

2. **Per-page imagery direction (scope=`page`, category=`design`).** This becomes meaningful for hero variants beyond `hero_home` — `hero_about`, `hero_services`, `hero_tools`, etc. PLAN section 11.5 ("Planner produces an `imagery_plan`") and section 13.2 ("Per-section/per-component imagery_plan entry") are the natural successors. **They become straightforward to implement once `site_plan_directives` exists**: the brief renderer assembles per-page directives, and image generation for non-`hero_home` purposes reads from the relevant page brief.

This is exactly what I deferred under "per-image granularity is nice-to-have" in the plan. Phase 1 changes the cost/benefit on that deferral favourably — the infrastructure for per-page directives lands as part of someone else's work, not as a separate plumbing exercise on the imagery side.

---

## Updates to PLAN_imagery_loop_closure.md

The plan document needs two amendments to reflect the Phase 1 architecture. Neither changes the sequencing.

### Amendment 1 — Phase 4.1 SQL needs a directive read once Phase 1 lands

Current text in PLAN section "Phase 4 — Visual auditor extension":

> Add subqueries to the existing big SELECT:
> - `(SELECT data->>'imagery_direction' FROM site_specs WHERE site_id = s.id AND aspect = 'design_intent' AND is_current = true) as imagery_direction`

After Phase 1, the auditor would also want to load relevant directives:

> - From `site_plan_directives` for the current plan, scope=`site`, category=`design` — ordered, joined with category=`content` for cross-cutting voice rules. Brief renderer already produces this shape; the auditor calls the renderer's site-scope helper rather than constructing the SQL.

Defer this until Phase 1 step 7 (brief renderer) is live. Adding it now would couple my work to schemas that don't exist yet.

### Amendment 2 — Phase 6 (imagery-quality-auditor) reads the brief renderer, not raw spec

Current text in PLAN section "Phase 6 — `imagery-quality-auditor` agent":

> Workflow:
>   ensure_site_record
>     → load_imagery_context        (SQL: assets, imagery_direction, identity)
>     → ...

After Phase 1, the load step changes shape:

>     → load_imagery_context        (SQL: assets, identity from site_specs;
>                                    brief-rendered imagery directives via
>                                    new query_database step or a new
>                                    `load_brief` action)

Same intent, different read path. Cosmetic when the brief renderer is in place; significant if the renderer is delayed.

---

## What about doc 030's Bug 4 ("Planner inventing hero images that ignore site_archetype.design.imagery")?

Doc 029 lists this under "What Phase 0 doesn't fix... Phase 1 territory." Worth noting because it overlaps with my work without being identical.

That bug is about the *planner* generating bad image prompts on adopted sites — `image_prompts.hero_home` reflecting an aspirational style that contradicts what was already adopted. My Phase 0.1 doesn't fix this — it composes the planner's prompt with the strategic imagery_direction, which on an adopted site will be the adoption's fingerprint-derived direction. So **the symptom partly improves**: even if the planner emits a misaligned `image_prompts.hero_home`, the prepended `Style direction: <adopted imagery_direction>` will pull the generation back toward the adopted look.

This is a side-benefit, not a fix. The full fix (planner reads `design_intent` as input but doesn't override it on the way back out) is Phase 1 step 4, not mine. The improvement Phase 0.1 gives is opportunistic.

---

## Compatibility with the development guide

Quick check against the explicit rules:

| Rule (from `001_development_guide.md`) | Phase 0.1 compliance |
|---|---|
| Don't use `logger.Debug` | Helpers use `logger.Warn` and `logger.Info`. |
| Schema check before SQL | Confirmed `site_specs.aspect`, `site_specs.data` shape via `some_schemas` and the existing classifier prompt. JSONB `->>'imagery_direction'` is the shape used by `page-content-writer` already. |
| Reuse existing patterns | `site_specs WHERE aspect = X AND is_current = true` mirrors `read_site_spec` and `domain-research-classifier`'s read. The prompt-prefix pattern (`Style direction: ...\n\nSubject: ...`) is the established convention in image generation prompting. |
| Workflows simple, complexity in Go | The migration is six surgical `jsonb_set` updates — pure data, no logic. The complexity (DB read, fallthrough on missing data, prompt composition) is in Go where it belongs. |
| Don't change variable names without flagging | Adds `siteID` local; doesn't rename anything. |
| Field name collisions warning | Discussed above — does not apply because the action uses `GetInputData` not `ExtractActionInputs`. |

No conflicts.

---

## Other guideline notes from this read

A few things in the new docs are worth carrying into how subsequent imagery phases ship:

- **`llm_tier` annotations (doc 029).** Image generation has a parallel concept (model selection — Stability vs FLUX vs Banana Pro vs Midjourney). Whatever annotation system lands for `llm_tier` is probably the right substrate for image-model selection too. PLAN's "provider router" phase should align with whatever `llm_tier` infrastructure ships, rather than building its own routing.

- **Lock transfer pattern (doc 030).** Lock transfer happens in `write_site_plan`. PLAN section 2 ("Schema groundwork: locking and multi-image readiness") proposed mirroring the `page_components` locking pattern on `assets`. Doc 030 uses the same pattern on `site_plan_directives`, which validates the approach. The locking semantics (composite key `(scope, scope_ref, category, subject, ordering)`) might be worth borrowing for asset locking too — though for assets the natural composite key is just `(site_id, asset_key)`.

- **Cycle budget (doc 029 reconciler).** The reconciler caps emissions per run via `cycle_budget`. PLAN's discovery checks for unfulfilled image prompts (Phase 1 of my plan) should respect the same budget once the reconciler exists, rather than emitting an unbounded number of `needs_image` items. Today no budget exists, so Phase 1.1-1.3 ship without it; once the reconciler is in place, the imagery checks should plug into the same budget mechanism.

- **`target_site_id` naming.** Adopted in PLAN for any *new* action I propose (e.g. `mirror_adoption_images_action.go` in PLAN Phase 3.1). For modifications to existing actions, keep `site_id` as established.

---

## Recommended next step

The Phase 0.1 deliverables stand. Before applying them I'd suggest one verification:

```sql
SELECT site_id,
       data->>'imagery_direction' AS imagery_direction,
       source,
       created_at
FROM site_specs
WHERE aspect = 'design_intent'
  AND is_current = true
LIMIT 5;
```

This confirms the field is populated on real sites today, and shows which `source` agent wrote it (classifier vs adoption vs build-site-planner). After Phase 1, the `source` column should only ever read `domain-research-classifier` or `site-adoption-agent`. Recording it pre-Phase-1 gives a useful diff when Phase 1 ships.

Once that verification clears, the Phase 0.1 SQL and Go patch can land. They're independent of Phase 1 timing — they help today, and behave better (more consistent `imagery_direction` source) once Phase 1 lands.

After that, the next imagery step is Phase 0.2 (`origin_model` population). Same shape: small additive change, independent of Phase 1 architecture.
