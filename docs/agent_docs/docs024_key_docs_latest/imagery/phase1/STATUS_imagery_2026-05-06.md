# Imagery work — STATUS as of 2026-05-06

Single overlay document tracking what's deployed, verified, and pending in
the imagery loop closure work. References the PLAN, FOCUS, and ASSESSMENT
documents already in this folder; doesn't replace them.

---

## At-a-glance

| Phase | Title | Code | Schema | DB-state | E2E verified |
|---|---|:-:|:-:|:-:|:-:|
| 0.1 | Read `imagery_direction` into image prompt | ✅ | n/a | ✅ | ⏳ |
| 0.2 | Populate `origin_prompt` and `origin_model` | ✅ | n/a | ✅ | ⏳ |
| 1.1 | Discovery check `unfulfilled_image_prompt` | ✅ | n/a | ✅ | ✅ |
| 1.2 | Discovery check `placeholder_image_in_use` | ✅ | n/a | ✅ | partial |
| 1.3 | Discovery check `image_url_404` | ✅ | n/a | ✅ | partial |
| 1.4 | Register checks in `design-discovery-agent` | n/a | ✅ | ✅ | ✅ |
| 1.5 | Smoke test handler dispatch | n/a | n/a | n/a | ⏳ |
| 2 | Asset locking + multi-image readiness | — | — | — | — |
| 3 | Adoption image mirror | — | — | — | — |
| 4 | Visual auditor — text-only imagery awareness | — | — | — | — |
| 5 | Vision-capable LLM path | — | — | — | — |
| 6 | `imagery-quality-auditor` agent | — | — | — | — |

Legend: ✅ done · ⏳ pending · — not started · "partial" = check fires correctly but no candidate site has been observed exercising the symptom path yet.

---

## What's deployed and verified

### Phase 0.1 — `imagery_direction` reaches image_generator

**Code change** in `platform/orchestration/actions/generate_image_actions.go`:
- New helpers: `getImageryDirectionForSite`, `composeImagePromptWithDirection`, `endsWithSentenceBoundary`, constant `maxImageryDirectionInPrompt = 200`.
- 17-line block injected after `getImagePromptWithPriority` reads `inputData["site_id"]`, queries `site_specs.data->>'imagery_direction'` for `aspect='design_intent' AND is_current=true`, and prepends with sentence-boundary truncation.
- Format: `"<truncated direction>. <subject>"` — no `Style direction:`/`Subject:` labels (saves SDXL token budget).

**Schema change**: none.

**Migration** (`phase_0_combined_migration.sql`, section 1): added `"site_id": "site_record.site_id"` to image-generator input_mappings in 6 step locations across `image-build-handler`, `site-work-orchestrator`, `pageflow-builder`.

**Verified**: SQL confirmed migration applied (n=14 input_mappings updated). Helper unit-tested against the 5 sites from the verification corpus.

**Not yet verified end-to-end**: that during a real image generation, `assets.origin_prompt` actually begins with the imagery_direction prefix. Needs a site to run `image-build-handler` end-to-end. This is Option 1 below.

### Phase 0.2 — `origin_prompt` and `origin_model` populated

**Code change** in `platform/orchestration/actions/v3_site_actions.go`:
- 16-line extraction block in `StoreAssetAction` reading `origin_prompt_field` (path) and `origin_model` (literal) or `origin_model_field` (path).
- Both `INSERT` statements expanded to write `origin_prompt` and `origin_model`. `ON CONFLICT` clause uses `COALESCE(EXCLUDED.X, assets.X)` so re-stores with empty values preserve prior provenance.

**Schema change**: none — both columns existed already, just unused.

**Bug fix included**: `origin_prompt_field` was already passed by workflows but the action's INSERT silently dropped it. Every row in `assets` had `origin_prompt = NULL`. After this lands, new generations populate it.

**Migration** (`phase_0_combined_migration.sql`, sections 2 and 3): added `origin_model: "sdxl"` literal to 6 store_asset configs; normalised `origin_prompt_field` in `site-work-orchestrator` and `pageflow-builder` from `site_plan.image_prompts.X` to `<X>_result.prompt` so the recorded prompt reflects what the model actually saw post-Phase-0.1 composition.

**Verified**: migration applied. Action signature compile-checked.

**Not yet verified end-to-end**: same as 0.1 — needs a real generation to confirm the columns populate correctly. Bundled with Option 1.

### Phase 1 — Discovery checks live

**Files added** under `platform/orchestration/actions/discovery_checks/`:
- `imagery_helpers.go` — shared `loadImagePromptsForSite` and `hasActiveAssetForPurpose`.
- `check_unfulfilled_image_prompt.go` — three categories (`logo`, `hero_home`, `hero_<page>`), v2 with `classifyPromptKey`.
- `check_placeholder_image_in_use.go` — looks for fallback paths in deployed HTML.
- `check_image_url_404.go` — regex-extracts image references, compares against active asset purposes.

**Migration** (`phase_1_register_imagery_checks.sql`): appended the three names to `design-discovery-agent.run_checks.config.checks` using `||` array-append pattern (idempotent).

**Verified end-to-end on `00ff3af5-...` (robot-hands.com)**:
- Trigger fired discovery agent at 17:45:00 UTC.
- Check classified the planner's 7 image_prompts:
  - 2 routable: `unfulfilled_image_prompt:logo`, `unfulfilled_image_prompt:hero` (severity high, handler image-build-handler).
  - 5 flag-only: `unfulfilled_image_prompt:hero_about`, `:hero_tools`, `:hero_matchmatrix`, `:hero_how_it_works`, `:hero_selection_guide` (severity medium, no handler).
- All 7 work items present in `site_work_items` with `status='detected'`, distinct `item_key` per row, batch insert with identical timestamp.
- Idempotency confirmed: original 2 routable rows from an earlier 07:28 trigger kept their original `created_at`; only the 5 new variants inserted (via `ON CONFLICT … DO NOTHING` skipping the existing keys).

---

## What's deployed but not yet exercised end-to-end

The dispatch-side flow (PLAN section 1.5):
- `triage_detected_items` promoting 7 detected items → triaged, pipeline=build.
- `build-dispatch-loop` claiming the 2 routable items.
- `image-build-handler` workflow running through call_image_generator → store_asset → deploy_image_asset.
- Phase 0.1's `imagery_direction` prefix landing in `assets.origin_prompt`.
- Phase 0.2's `origin_model='sdxl'` populating.

These are next, and are what Option 1 below verifies.

---

## Live data state (as observed)

### Site under test
- **site_id**: `00ff3af5-dad8-4770-9f70-3edc267a3c92`
- **domain**: `robot-hands.com`
- **planner-supplied prompts** (7): `logo`, `hero_home`, `hero_about`, `hero_tools`, `hero_matchmatrix`, `hero_how_it_works`, `hero_selection_guide`.
- **active assets**: none.
- **imagery_direction**: rich (~720 chars, robotic gripper industrial photography theme).

### `site_work_items` for that site (imagery-related, post-1745 trigger)
| item_type | status | severity | handler_agent | item_key | created_at |
|---|---|---|---|---|---|
| needs_logo | detected | high | image-build-handler | unfulfilled_image_prompt:logo | 2026-05-06 07:28 |
| needs_hero_image | detected | high | image-build-handler | unfulfilled_image_prompt:hero | 2026-05-06 07:28 |
| unfulfilled_hero_variant | detected | medium | (empty) | unfulfilled_image_prompt:hero_about | 2026-05-06 17:45 |
| unfulfilled_hero_variant | detected | medium | (empty) | unfulfilled_image_prompt:hero_tools | 2026-05-06 17:45 |
| unfulfilled_hero_variant | detected | medium | (empty) | unfulfilled_image_prompt:hero_matchmatrix | 2026-05-06 17:45 |
| unfulfilled_hero_variant | detected | medium | (empty) | unfulfilled_image_prompt:hero_how_it_works | 2026-05-06 17:45 |
| unfulfilled_hero_variant | detected | medium | (empty) | unfulfilled_image_prompt:hero_selection_guide | 2026-05-06 17:45 |

---

## Findings worth recording (not blockers)

1. **`created_by` populated as `generic` rather than `design-discovery-agent`** when triggered via `system.agent.generic.requests` directly. The action reads `params.ExecutionContext.Sender.AgentType`, which in a manual-trigger path carries the pod's identity instead of the workflow's intended agent type. Same `dctx.AgentType` plumbing as every other check — not specific to ours. Worth knowing for future log/SQL filters: `created_by` is unreliable when triggers come via the generic topic. When `improvement-loop` calls discovery normally, the value populates correctly (see other checks' historical rows).

2. **Earlier 07:28 trigger ran on the pre-rewrite binary**. Original v1 of `check_unfulfilled_image_prompt.go` had a static 2-key mapping; the variant code was only deployed at the ~15:48 redeploy. Explains why the first trigger emitted 2 items, not 7. The second trigger at 17:45 ran on the v2 binary and emitted the 5 variants.

3. **`findings:5, items_inserted:0` log line** observed in an earlier `RunDiscoveryChecksAction: Complete` line. Came from the existing 11-check set (not ours), and predates the Phase 1 migration. The gap between findings and inserts suggests existing checks may be hitting suppression or conflict-on-key. Worth investigating in its own right but unrelated to the imagery work.

4. **Image-build-handler's deploy step is hardcoded** to `assets/images/hero.jpg` and `assets/images/logo.png` with the `assets.UNIQUE(site_id, purpose)` constraint. This is why the 5 hero variants are flag-only today — Phase 2 unblocks them.

5. **`/mnt/user-data/outputs/PLAN_imagery_loop_closure.md` Phase 4 amendment**: When Phase 1 of doc 030 (brief renderer / `site_plan_directives`) lands, `getImageryDirectionForSite` should be extended to also pull plan-time imagery directives. Tracked in `ASSESSMENT_phase_0_1_vs_phase_1_architecture.md`. Not urgent — the strategic-only read is correct for today's data.

---

## What's next — Option 1: verify the dispatch-side end-to-end

Goal: drive the 2 routable work items through to deployed assets and confirm Phase 0.1 + Phase 0.2 land in real `assets` rows.

### Steps

1. **Trigger improvement-loop for the site.**
   ```
   AGENT_TYPE="improvement-loop"
   SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"
   DOMAIN="robot-hands.com"
   ```
   Same kcat block as before, with `AGENT_TYPE=improvement-loop`. Improvement-loop's first action is `triage_detected_items` which promotes all 7 work items from `detected` → `triaged` and rewrites pipeline `design` → `build`. It then calls `build-dispatch-loop` which claims items by handler_agent.

2. **Watch for asset generation.**
   ```bash
   kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=0 | \
     grep -E 'image-build-handler|image-generator|store_asset|deploy_image|imagery_direction|Selected prompt for execution'
   ```
   Workflow will spawn `image-build-handler` twice (once for `needs_logo`, once for `needs_hero_image`). Each runs ensure_site_record → call_image_generator → store_asset → deploy_image_asset → complete. Stability adapter call adds 30-60 seconds per image.

3. **After completion, query `assets`:**
   ```sql
   SELECT purpose, origin_type, origin_model,
          LEFT(origin_prompt, 280) AS origin_prompt_preview,
          url, created_at
   FROM assets
   WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
     AND status = 'active'
   ORDER BY created_at DESC;
   ```

   Expected for each row:
   - `purpose` ∈ {`logo`, `hero`}
   - `origin_type` = `generated`
   - `origin_model` = `sdxl` ← confirms Phase 0.2 wiring
   - `origin_prompt` begins with the imagery_direction prefix ← confirms Phase 0.1 composition

4. **Confirm work-item state:**
   ```sql
   SELECT item_type, status, attempt_count, claimed_at, claimed_by, updated_at
   FROM site_work_items
   WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
     AND item_type IN ('needs_logo','needs_hero_image','unfulfilled_hero_variant')
   ORDER BY created_at DESC;
   ```
   Expected: 2 routable items in `complete`, 5 variants still `triaged` with empty handler_agent (build-dispatch-loop skipped them as designed).

### Failure modes to watch for

| Symptom | Likely cause | Where to look |
|---|---|---|
| Items stay in `detected` after improvement-loop runs | Triage didn't promote them; possibly improvement-loop didn't reach triage step | Chassis logs for `triage_detected_items` |
| Items reach `triaged` but never `claimed` | build-dispatch-loop didn't pick them up; possibly wrong pipeline value or filter mismatch | `LoadWorkItemsAction` logs |
| Items `claimed` but never `complete` | Image-build-handler workflow stuck somewhere | Filter chassis logs by `agent_type=image-build-handler` and the orchestration_id |
| Asset row created but `origin_prompt` empty | Phase 0.2's `origin_prompt_field` config didn't hit the read path; check workflow definition | `agent_definitions` row for relevant parent |
| Asset row created but `origin_prompt` lacks imagery_direction prefix | Phase 0.1's read returned empty; check `inputData["site_id"]` reached the action | `Selected prompt for execution` log line should show `+imagery_direction` in source field |
| Asset row created but `origin_model` empty | Migration's `origin_model` literal didn't apply, or action didn't read it | `agent_definitions` row for relevant store_asset step |

If everything goes through cleanly, Phase 0.1, 0.2, and 1 are fully verified end-to-end. We can then move to Phase 2.
